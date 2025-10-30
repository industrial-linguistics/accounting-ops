package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"github.com/pkg/browser"

	"auth.industrial-linguistics.com/accounting-ops/internal/broker"
)

// App wraps the CLI runtime state.
type App struct {
	BrokerBaseURL string
	HTTPClient    *http.Client
	Keyring       keyring.Keyring
	Stdout        io.Writer
	Stderr        io.Writer
	Stdin         io.Reader
}

// NewApp creates a new CLI app with default configuration.
func NewApp() (*App, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		cfgDir = filepath.Join(os.TempDir(), "accounting-ops")
	}
	kr, err := keyring.Open(keyring.Config{
		ServiceName:             "accounting-ops",
		FileDir:                 filepath.Join(cfgDir, "accounting-ops"),
		KeychainName:            "accounting-ops",
		WinCredPrefix:           "accounting-ops",
		LibSecretCollectionName: "accounting-ops",
		KWalletAppID:            "accounting-ops",
		KWalletFolder:           "accounting-ops",
	})
	if err != nil {
		return nil, err
	}
	// Default to production broker, override with ACCOUNTING_OPS_BROKER environment variable
	brokerURL := "https://auth.industrial-linguistics.com/cgi-bin/broker"
	if envURL := os.Getenv("ACCOUNTING_OPS_BROKER"); envURL != "" {
		brokerURL = strings.TrimRight(envURL, "/")
	}
	return &App{
		BrokerBaseURL: brokerURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Keyring: kr,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Stdin:   os.Stdin,
	}, nil
}

// Run executes the CLI with the provided arguments.
func (a *App) Run(args []string) int {
	if len(args) == 0 {
		a.printUsage()
		return 1
	}
	switch args[0] {
	case "connect":
		return a.runConnect(args[1:])
	case "list":
		return a.runList(args[1:])
	case "whoami":
		return a.runWhoAmI(args[1:])
	case "refresh":
		return a.runRefresh(args[1:])
	case "revoke":
		return a.runRevoke(args[1:])
	case "help", "-h", "--help":
		a.printUsage()
		return 0
	default:
		fmt.Fprintf(a.Stderr, "unknown command %q\n", args[0])
		a.printUsage()
		return 1
	}
}

func (a *App) printUsage() {
	fmt.Fprintf(a.Stdout, `Accounting Ops CLI

Commands:
  connect <provider> --profile NAME [--broker URL]
  list
  whoami --profile NAME --provider PROVIDER
  refresh --profile NAME --provider PROVIDER [--broker URL]
  revoke --profile NAME --provider PROVIDER

Environment Variables:
  ACCOUNTING_OPS_BROKER  Override default broker URL
                         Production (default): https://auth.industrial-linguistics.com/cgi-bin/broker
                         Development: https://auth-dev.industrial-linguistics.com/cgi-bin/broker
`)
}

func (a *App) runConnect(args []string) int {
	fs := flag.NewFlagSet("connect", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	profile := fs.String("profile", "", "profile name")
	brokerURL := fs.String("broker", "", "override broker base URL")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(a.Stderr, "provider argument required")
		return 1
	}
	provider := strings.ToLower(fs.Arg(0))
	if *profile == "" {
		fmt.Fprintln(a.Stderr, "--profile is required")
		return 1
	}
	baseURL := a.BrokerBaseURL
	if *brokerURL != "" {
		baseURL = strings.TrimRight(*brokerURL, "/")
	}

	startResp, err := a.startAuth(baseURL, provider, *profile)
	if err != nil {
		fmt.Fprintf(a.Stderr, "start auth failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(a.Stdout, "Opening browser for %s authorisation...\n", provider)
	if err := browser.OpenURL(startResp.AuthURL); err != nil {
		fmt.Fprintf(a.Stderr, "unable to open browser automatically: %v\n", err)
		fmt.Fprintf(a.Stdout, "Please open this URL manually:\n%s\n", startResp.AuthURL)
	}

	pollURL := startResp.PollURL
	if !strings.HasPrefix(pollURL, "http") {
		base, err := url.Parse(baseURL)
		if err != nil {
			fmt.Fprintf(a.Stderr, "invalid broker URL: %v\n", err)
			return 1
		}
		rel, err := url.Parse(pollURL)
		if err != nil {
			fmt.Fprintf(a.Stderr, "invalid poll URL from broker: %v\n", err)
			return 1
		}
		pollURL = base.ResolveReference(rel).String()
	}

	fmt.Fprintln(a.Stdout, "Waiting for authorisation...")
	envelope, err := a.pollForTokens(pollURL)
	if err != nil {
		fmt.Fprintf(a.Stderr, "authorisation failed: %v\n", err)
		return 1
	}
	envelope.Provider = provider

	prof := envelopeToProfile(envelope, *profile)

	if provider == "xero" {
		if err := a.promptForXeroTenant(&prof, envelope); err != nil {
			fmt.Fprintf(a.Stderr, "tenant selection failed: %v\n", err)
			return 1
		}
	}

	if err := a.saveProfile(prof); err != nil {
		fmt.Fprintf(a.Stderr, "unable to save credentials: %v\n", err)
		return 1
	}

	a.printProfileSummary(prof)
	return 0
}

func (a *App) runList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	if err := fs.Parse(args); err != nil {
		return 1
	}
	keys, err := a.Keyring.Keys()
	if err != nil {
		fmt.Fprintf(a.Stderr, "unable to enumerate profiles: %v\n", err)
		return 1
	}
	if len(keys) == 0 {
		fmt.Fprintln(a.Stdout, "No stored profiles.")
		return 0
	}
	fmt.Fprintf(a.Stdout, "Stored profiles (%d):\n", len(keys))
	for _, key := range keys {
		item, err := a.Keyring.Get(key)
		if err != nil {
			fmt.Fprintf(a.Stderr, "  %s: error reading: %v\n", key, err)
			continue
		}
		var prof ProfileData
		if err := json.Unmarshal(item.Data, &prof); err != nil {
			fmt.Fprintf(a.Stderr, "  %s: corrupt entry: %v\n", key, err)
			continue
		}
		fmt.Fprintf(a.Stdout, "  %s (%s) – expires %s\n", prof.Name, prof.Provider, prof.ExpiresAt.Format(time.RFC3339))
	}
	return 0
}

func (a *App) runWhoAmI(args []string) int {
	fs := flag.NewFlagSet("whoami", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	profile := fs.String("profile", "", "profile name")
	provider := fs.String("provider", "", "provider name")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	prof, err := a.loadProfile(*profile, *provider)
	if err != nil {
		fmt.Fprintf(a.Stderr, "unable to load profile: %v\n", err)
		return 1
	}
	fmt.Fprintf(a.Stdout, "Profile %s (%s)\n", prof.Name, prof.Provider)
	fmt.Fprintf(a.Stdout, "  Access token expires: %s\n", prof.ExpiresAt.Format(time.RFC3339))
	if prof.Provider == "xero" {
		fmt.Fprintf(a.Stdout, "  Tenant ID: %s\n", prof.TenantID)
		fmt.Fprintf(a.Stdout, "  Tenant Name: %s\n", prof.TenantName)
	}
	if prof.Provider == "deputy" {
		fmt.Fprintf(a.Stdout, "  Endpoint: %s\n", prof.Endpoint)
	}
	if prof.Provider == "qbo" {
		fmt.Fprintf(a.Stdout, "  Realm ID: %s\n", prof.RealmID)
	}
	return 0
}

func (a *App) runRefresh(args []string) int {
	fs := flag.NewFlagSet("refresh", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	profile := fs.String("profile", "", "profile name")
	provider := fs.String("provider", "", "provider name")
	brokerURL := fs.String("broker", "", "override broker base URL")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	prof, err := a.loadProfile(*profile, *provider)
	if err != nil {
		fmt.Fprintf(a.Stderr, "unable to load profile: %v\n", err)
		return 1
	}

	var envelope broker.TokenEnvelope
	switch prof.Provider {
	case "xero":
		envelope, err = a.refreshXero(*prof)
	case "deputy", "qbo":
		baseURL := a.BrokerBaseURL
		if *brokerURL != "" {
			baseURL = strings.TrimRight(*brokerURL, "/")
		}
		envelope, err = a.refreshViaBroker(baseURL, *prof)
	default:
		err = fmt.Errorf("unsupported provider %s", prof.Provider)
	}
	if err != nil {
		fmt.Fprintf(a.Stderr, "refresh failed: %v\n", err)
		return 1
	}

	updated := envelopeToProfile(envelope, prof.Name)
	if prof.Provider == "xero" {
		updated.TenantID = prof.TenantID
		updated.TenantName = prof.TenantName
		updated.TenantType = prof.TenantType
	}
	if prof.Provider == "deputy" && updated.Endpoint == "" {
		updated.Endpoint = prof.Endpoint
	}
	if prof.Provider == "qbo" && updated.RealmID == "" {
		updated.RealmID = prof.RealmID
	}

	if err := a.saveProfile(updated); err != nil {
		fmt.Fprintf(a.Stderr, "unable to save refreshed credentials: %v\n", err)
		return 1
	}
	fmt.Fprintln(a.Stdout, "Token refreshed.")
	return 0
}

func (a *App) runRevoke(args []string) int {
	fs := flag.NewFlagSet("revoke", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	profile := fs.String("profile", "", "profile name")
	provider := fs.String("provider", "", "provider name")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if *provider == "" {
		fmt.Fprintln(a.Stderr, "--provider is required")
		return 1
	}
	key := makeProfileKey(*provider, *profile)
	if err := a.Keyring.Remove(key); err != nil {
		if !errors.Is(err, keyring.ErrKeyNotFound) {
			fmt.Fprintf(a.Stderr, "unable to remove profile: %v\n", err)
			return 1
		}
	}
	fmt.Fprintf(a.Stdout, "Removed stored credentials for %s (%s).\n", *profile, *provider)
	return 0
}

func (a *App) startAuth(baseURL, provider, profile string) (*startResponse, error) {
	body := map[string]string{
		"provider": provider,
		"profile":  profile,
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, baseURL+"/v1/auth/start", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("broker error: %s", strings.TrimSpace(string(payload)))
	}
	var out startResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (a *App) pollForTokens(pollURL string) (broker.TokenEnvelope, error) {
	for {
		req, err := http.NewRequest(http.MethodGet, pollURL, nil)
		if err != nil {
			return broker.TokenEnvelope{}, err
		}
		resp, err := a.HTTPClient.Do(req)
		if err != nil {
			return broker.TokenEnvelope{}, err
		}
		if resp.StatusCode >= 400 {
			payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			resp.Body.Close()
			return broker.TokenEnvelope{}, fmt.Errorf("broker error: %s", strings.TrimSpace(string(payload)))
		}
		var raw map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
			resp.Body.Close()
			return broker.TokenEnvelope{}, err
		}
		resp.Body.Close()
		if status, ok := raw["status"].(string); ok && status == "pending" {
			time.Sleep(2 * time.Second)
			continue
		}
		data, err := json.Marshal(raw)
		if err != nil {
			return broker.TokenEnvelope{}, err
		}
		var env broker.TokenEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			return broker.TokenEnvelope{}, err
		}
		return env, nil
	}
}

func (a *App) refreshViaBroker(baseURL string, prof ProfileData) (broker.TokenEnvelope, error) {
	body := map[string]string{
		"provider":      prof.Provider,
		"refresh_token": prof.RefreshToken,
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, baseURL+"/v1/token/refresh", bytes.NewReader(data))
	if err != nil {
		return broker.TokenEnvelope{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return broker.TokenEnvelope{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return broker.TokenEnvelope{}, fmt.Errorf("broker error: %s", strings.TrimSpace(string(payload)))
	}
	var env broker.TokenEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return broker.TokenEnvelope{}, err
	}
	return env, nil
}

func (a *App) refreshXero(prof ProfileData) (broker.TokenEnvelope, error) {
	clientID := os.Getenv("XERO_CLIENT_ID")
	if clientID == "" {
		return broker.TokenEnvelope{}, errors.New("XERO_CLIENT_ID must be set in the environment for refresh")
	}
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", prof.RefreshToken)
	data.Set("client_id", clientID)

	endpoint := "https://identity.xero.com/connect/token"
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return broker.TokenEnvelope{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if secret := os.Getenv("XERO_CLIENT_SECRET"); secret != "" {
		req.SetBasicAuth(clientID, secret)
	}
	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return broker.TokenEnvelope{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return broker.TokenEnvelope{}, fmt.Errorf("xero token error: %s", strings.TrimSpace(string(payload)))
	}
	var env broker.TokenEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return broker.TokenEnvelope{}, err
	}
	env.Provider = "xero"
	return env, nil
}

func (a *App) promptForXeroTenant(prof *ProfileData, env broker.TokenEnvelope) error {
	if len(env.Tenants) == 0 {
		return errors.New("no tenants returned; connect to an organisation before continuing")
	}
	fmt.Fprintln(a.Stdout, "Select a Xero tenant:")
	for i, t := range env.Tenants {
		fmt.Fprintf(a.Stdout, "  [%d] %s (%s)\n", i+1, t.TenantName, t.TenantID)
	}
	reader := bufio.NewReader(a.Stdin)
	for {
		fmt.Fprint(a.Stdout, "Enter number: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		idx, err := parseIndex(line, len(env.Tenants))
		if err != nil {
			fmt.Fprintf(a.Stderr, "%v\n", err)
			continue
		}
		tenant := env.Tenants[idx]
		prof.TenantID = tenant.TenantID
		prof.TenantName = tenant.TenantName
		prof.TenantType = tenant.TenantType
		return nil
	}
}

func parseIndex(input string, max int) (int, error) {
	i, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid number")
	}
	if i < 1 || i > max {
		return 0, fmt.Errorf("selection must be between 1 and %d", max)
	}
	return i - 1, nil
}

func (a *App) saveProfile(prof ProfileData) error {
	prof.Provider = strings.ToLower(prof.Provider)
	prof.Name = strings.TrimSpace(prof.Name)
	prof.ExpiresAt = prof.ExpiresAt.UTC()
	data, err := json.Marshal(prof)
	if err != nil {
		return err
	}
	item := keyring.Item{Key: makeProfileKey(prof.Provider, prof.Name), Data: data, Label: prof.Provider + " profile"}
	return a.Keyring.Set(item)
}

func (a *App) loadProfile(name, provider string) (*ProfileData, error) {
	if name == "" {
		return nil, errors.New("--profile is required")
	}
	provider = strings.ToLower(provider)
	if provider == "" {
		// attempt to auto-detect by scanning entries
		keys, err := a.Keyring.Keys()
		if err != nil {
			return nil, err
		}
		var matches []ProfileData
		for _, key := range keys {
			item, err := a.Keyring.Get(key)
			if err != nil {
				continue
			}
			var prof ProfileData
			if err := json.Unmarshal(item.Data, &prof); err != nil {
				continue
			}
			if strings.EqualFold(prof.Name, name) {
				matches = append(matches, prof)
			}
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("profile %s not found", name)
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("multiple providers for profile %s; specify --provider", name)
		}
		provider = matches[0].Provider
	}
	item, err := a.Keyring.Get(makeProfileKey(provider, name))
	if err != nil {
		return nil, err
	}
	var prof ProfileData
	if err := json.Unmarshal(item.Data, &prof); err != nil {
		return nil, err
	}
	return &prof, nil
}

func (a *App) printProfileSummary(prof ProfileData) {
	fmt.Fprintf(a.Stdout, "Connected %s (%s).\n", prof.Name, prof.Provider)
	switch prof.Provider {
	case "xero":
		fmt.Fprintf(a.Stdout, "  Tenant: %s (%s)\n", prof.TenantName, prof.TenantID)
	case "deputy":
		fmt.Fprintf(a.Stdout, "  Endpoint: %s\n", prof.Endpoint)
	case "qbo":
		fmt.Fprintf(a.Stdout, "  Realm ID: %s\n", prof.RealmID)
	}
}

type startResponse struct {
	AuthURL string `json:"auth_url"`
	PollURL string `json:"poll_url"`
	Session string `json:"session"`
}

// ProfileData represents stored profile credentials.
type ProfileData struct {
	Name         string         `json:"name"`
	Provider     string         `json:"provider"`
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresAt    time.Time      `json:"expires_at"`
	Scope        string         `json:"scope,omitempty"`
	RealmID      string         `json:"realmId,omitempty"`
	Endpoint     string         `json:"endpoint,omitempty"`
	TenantID     string         `json:"xero_tenant_id,omitempty"`
	TenantName   string         `json:"xero_tenant_name,omitempty"`
	TenantType   string         `json:"xero_tenant_type,omitempty"`
	TokenType    string         `json:"token_type,omitempty"`
	Extras       map[string]any `json:"extras,omitempty"`
}

func makeProfileKey(provider, name string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "-")
	return fmt.Sprintf("%s:%s", provider, name)
}

func envelopeToProfile(env broker.TokenEnvelope, profileName string) ProfileData {
	expires := env.ExpiresAt
	if expires.IsZero() && env.ExpiresUnix != 0 {
		expires = time.Unix(env.ExpiresUnix, 0)
	}
	p := ProfileData{
		Name:         profileName,
		Provider:     env.Provider,
		AccessToken:  env.AccessToken,
		RefreshToken: env.RefreshToken,
		ExpiresAt:    expires,
		Scope:        env.Scope,
		RealmID:      env.RealmID,
		Endpoint:     env.Endpoint,
		TokenType:    env.TokenType,
	}
	if env.Raw != nil {
		p.Extras = env.Raw
	}
	return p
}
