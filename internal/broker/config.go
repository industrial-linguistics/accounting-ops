package broker

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config contains runtime configuration for the broker service.
type Config struct {
	XeroClientID     string
	XeroClientSecret string
	XeroRedirectURL  string
	XeroScopes       []string

	DeputyClientID     string
	DeputyClientSecret string
	DeputyRedirectURL  string
	DeputyScopes       []string

	QBOClientID     string
	QBOClientSecret string
	QBORedirectURL  string
	QBOScopes       []string

	MasterKey []byte

	SessionTTL  time.Duration
	PollTimeout time.Duration

	RateLimitAuthStart       int
	RateLimitAuthStartWindow time.Duration
	RateLimitPoll            int
	RateLimitPollWindow      time.Duration
	RateLimitRefresh         int
	RateLimitRefreshWindow   time.Duration
}

// DefaultConfig returns a Config populated with safe defaults.
func DefaultConfig() Config {
	return Config{
		SessionTTL:               time.Minute * 10,
		PollTimeout:              time.Second * 5,
		RateLimitAuthStart:       10,
		RateLimitAuthStartWindow: time.Minute,
		RateLimitPoll:            120,
		RateLimitPollWindow:      time.Minute,
		RateLimitRefresh:         60,
		RateLimitRefreshWindow:   time.Minute,
	}
}

// LoadConfigFromEnvFile parses a key=value file such as conf/broker.env.
func LoadConfigFromEnvFile(path string) (Config, error) {
	cfg := DefaultConfig()
	file, err := os.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("open env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexRune(line, '=')
		if idx == -1 {
			return cfg, fmt.Errorf("invalid line %d in %s", lineNo, filepath.Base(path))
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") && len(val) >= 2 {
			val = strings.Trim(val, "\"")
		}
		switch key {
		case "XERO_CLIENT_ID":
			cfg.XeroClientID = val
		case "XERO_CLIENT_SECRET":
			cfg.XeroClientSecret = val
		case "XERO_REDIRECT":
			cfg.XeroRedirectURL = val
		case "XERO_SCOPES":
			cfg.XeroScopes = parseScopes(val)
		case "DEPUTY_CLIENT_ID":
			cfg.DeputyClientID = val
		case "DEPUTY_CLIENT_SECRET":
			cfg.DeputyClientSecret = val
		case "DEPUTY_REDIRECT":
			cfg.DeputyRedirectURL = val
		case "DEPUTY_SCOPES":
			cfg.DeputyScopes = parseScopes(val)
		case "QBO_CLIENT_ID":
			cfg.QBOClientID = val
		case "QBO_CLIENT_SECRET":
			cfg.QBOClientSecret = val
		case "QBO_REDIRECT":
			cfg.QBORedirectURL = val
		case "QBO_SCOPES":
			cfg.QBOScopes = parseScopes(val)
		case "BROKER_MASTER_KEY":
			if val != "" {
				cfg.MasterKey = []byte(val)
			}
		case "SESSION_TTL_SECONDS":
			if val != "" {
				d, err := parseSeconds(val)
				if err != nil {
					return cfg, fmt.Errorf("SESSION_TTL_SECONDS: %w", err)
				}
				cfg.SessionTTL = d
			}
		case "POLL_TIMEOUT_SECONDS":
			if val != "" {
				d, err := parseSeconds(val)
				if err != nil {
					return cfg, fmt.Errorf("POLL_TIMEOUT_SECONDS: %w", err)
				}
				cfg.PollTimeout = d
			}
		case "RATE_LIMIT_AUTH_START":
			if val != "" {
				n, err := strconv.Atoi(val)
				if err != nil {
					return cfg, fmt.Errorf("RATE_LIMIT_AUTH_START: %w", err)
				}
				cfg.RateLimitAuthStart = n
			}
		case "RATE_LIMIT_AUTH_START_WINDOW_SECONDS":
			if val != "" {
				d, err := parseSeconds(val)
				if err != nil {
					return cfg, fmt.Errorf("RATE_LIMIT_AUTH_START_WINDOW_SECONDS: %w", err)
				}
				cfg.RateLimitAuthStartWindow = d
			}
		case "RATE_LIMIT_POLL":
			if val != "" {
				n, err := strconv.Atoi(val)
				if err != nil {
					return cfg, fmt.Errorf("RATE_LIMIT_POLL: %w", err)
				}
				cfg.RateLimitPoll = n
			}
		case "RATE_LIMIT_POLL_WINDOW_SECONDS":
			if val != "" {
				d, err := parseSeconds(val)
				if err != nil {
					return cfg, fmt.Errorf("RATE_LIMIT_POLL_WINDOW_SECONDS: %w", err)
				}
				cfg.RateLimitPollWindow = d
			}
		case "RATE_LIMIT_REFRESH":
			if val != "" {
				n, err := strconv.Atoi(val)
				if err != nil {
					return cfg, fmt.Errorf("RATE_LIMIT_REFRESH: %w", err)
				}
				cfg.RateLimitRefresh = n
			}
		case "RATE_LIMIT_REFRESH_WINDOW_SECONDS":
			if val != "" {
				d, err := parseSeconds(val)
				if err != nil {
					return cfg, fmt.Errorf("RATE_LIMIT_REFRESH_WINDOW_SECONDS: %w", err)
				}
				cfg.RateLimitRefreshWindow = d
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return cfg, fmt.Errorf("scan env file: %w", err)
	}

	applyProviderDefaults(&cfg)

	return cfg, nil
}

func applyProviderDefaults(cfg *Config) {
	if len(cfg.XeroScopes) == 0 {
		cfg.XeroScopes = []string{"offline_access", "accounting.transactions", "accounting.contacts"}
	}
	if len(cfg.DeputyScopes) == 0 {
		cfg.DeputyScopes = []string{"longlife_refresh_token"}
	}
	if len(cfg.QBOScopes) == 0 {
		cfg.QBOScopes = []string{"com.intuit.quickbooks.accounting"}
	}
}

func parseScopes(val string) []string {
	if val == "" {
		return nil
	}
	parts := strings.FieldsFunc(val, func(r rune) bool {
		return r == ' ' || r == ',' || r == '\t'
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseSeconds(val string) (time.Duration, error) {
	if val == "" {
		return 0, errors.New("empty value")
	}
	dur, err := time.ParseDuration(val + "s")
	if err != nil {
		return 0, err
	}
	return dur, nil
}

// Validate ensures the config has required values for production use.
func (c Config) Validate() error {
	var missing []string
	if c.XeroClientID == "" {
		missing = append(missing, "XERO_CLIENT_ID")
	}
	if c.XeroRedirectURL == "" {
		missing = append(missing, "XERO_REDIRECT")
	}
	if c.DeputyClientID == "" {
		missing = append(missing, "DEPUTY_CLIENT_ID")
	}
	if c.DeputyClientSecret == "" {
		missing = append(missing, "DEPUTY_CLIENT_SECRET")
	}
	if c.DeputyRedirectURL == "" {
		missing = append(missing, "DEPUTY_REDIRECT")
	}
	if c.QBOClientID == "" {
		missing = append(missing, "QBO_CLIENT_ID")
	}
	if c.QBOClientSecret == "" {
		missing = append(missing, "QBO_CLIENT_SECRET")
	}
	if c.QBORedirectURL == "" {
		missing = append(missing, "QBO_REDIRECT")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing configuration keys: %s", strings.Join(missing, ", "))
	}
	return nil
}
