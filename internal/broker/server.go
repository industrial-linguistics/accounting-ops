package broker

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Server implements the CGI HTTP handlers for the broker endpoints.
type Server struct {
	Config     Config
	Store      *Store
	HTTPClient *http.Client
	Logger     *log.Logger

	successTemplate *template.Template
	failureTemplate *template.Template
}

// NewServer constructs a broker Server.
func NewServer(cfg Config, store *Store, logger *log.Logger) *Server {
	return &Server{
		Config: cfg,
		Store:  store,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Logger:          logger,
		successTemplate: template.Must(template.New("success").Parse(successHTML)),
		failureTemplate: template.Must(template.New("failure").Parse(failureHTML)),
	}
}

// ServeHTTP routes incoming requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/v1/auth/start"):
		s.handleAuthStart(w, r)
	case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/callback/"):
		s.handleCallback(w, r)
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/v1/auth/poll/"):
		http.NotFound(w, r)
	case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/v1/auth/poll/"):
		s.handlePoll(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/v1/token/refresh"):
		s.handleRefresh(w, r)
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/healthz"):
		s.handleHealthz(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) basePathForRequest(r *http.Request, suffix string) string {
	p := r.URL.Path
	if idx := strings.Index(p, suffix); idx != -1 {
		return strings.TrimSuffix(p[:idx], "/")
	}
	return path.Dir(p)
}

func (s *Server) handleAuthStart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Provider string `json:"provider"`
		Profile  string `json:"profile"`
		PubKey   string `json:"pubkey"`
	}
	if err := decodeJSONBody(r.Body, &req); err != nil {
		respondJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if provider == "" {
		respondJSONError(w, http.StatusBadRequest, "provider is required")
		return
	}
	if req.Profile == "" {
		respondJSONError(w, http.StatusBadRequest, "profile is required")
		return
	}

	sessionID, err := randomID(24)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "failed to allocate session")
		return
	}
	state, err := randomID(32)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "failed to allocate state")
		return
	}

	var authURL string
	var codeVerifier sql.NullString
	switch provider {
	case "xero":
		authURL, codeVerifier, err = s.startXeroAuth(state)
	case "deputy":
		authURL, err = s.startDeputyAuth(state)
	case "qbo":
		authURL, err = s.startQBOAuth(state)
	default:
		respondJSONError(w, http.StatusBadRequest, "unsupported provider")
		return
	}
	if err != nil {
		s.Logger.Printf("start auth error provider=%s error=%v", provider, err)
		respondJSONError(w, http.StatusInternalServerError, "unable to start authorisation flow")
		return
	}

	expires := time.Now().Add(s.Config.SessionTTL)
	sess := Session{
		ID:           sessionID,
		Provider:     provider,
		State:        state,
		CodeVerifier: codeVerifier,
		CreatedAt:    time.Now(),
		ExpiresAt:    expires,
	}
	if err := s.Store.InsertSession(r.Context(), sess); err != nil {
		s.Logger.Printf("insert session error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, "unable to persist session")
		return
	}

	base := s.basePathForRequest(r, "/v1/auth/start")
	pollURL := fmt.Sprintf("%s/v1/auth/poll/%s", base, sessionID)
	resp := map[string]any{
		"auth_url": authURL,
		"poll_url": pollURL,
		"session":  sessionID,
		"state":    state,
	}
	respondJSON(w, http.StatusOK, resp)
}

func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	provider := providerFromCallbackPath(r.URL.Path)
	if provider == "" {
		http.NotFound(w, r)
		return
	}
	q := r.URL.Query()
	if errStr := q.Get("error"); errStr != "" {
		s.renderFailure(w, fmt.Sprintf("%s: %s", errStr, q.Get("error_description")))
		return
	}
	state := q.Get("state")
	if state == "" {
		s.renderFailure(w, "missing state parameter")
		return
	}
	sess, err := s.Store.LookupByState(r.Context(), provider, state)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.renderFailure(w, "unknown or expired session")
			return
		}
		s.Logger.Printf("lookup session failed: %v", err)
		s.renderFailure(w, "internal error")
		return
	}
	if time.Now().After(sess.ExpiresAt) {
		s.renderFailure(w, "session expired")
		return
	}

	var envelope TokenEnvelope
	switch provider {
	case "xero":
		envelope, err = s.exchangeXero(r.Context(), sess, q.Get("code"))
	case "deputy":
		envelope, err = s.exchangeDeputy(r.Context(), q.Get("code"))
	case "qbo":
		envelope, err = s.exchangeQBO(r.Context(), q.Get("code"), q.Get("realmId"))
	default:
		err = fmt.Errorf("unknown provider")
	}
	if err != nil {
		s.Logger.Printf("exchange tokens failed provider=%s error=%v", provider, err)
		s.renderFailure(w, "token exchange failed")
		return
	}

	envelope.Provider = provider
	envelope.ExpiresUnix = envelope.ExpiresAt.Unix()

	payload, err := jsonMarshal(envelope)
	if err != nil {
		s.Logger.Printf("marshal envelope error: %v", err)
		s.renderFailure(w, "internal serialisation error")
		return
	}

	var realmID *string
	if envelope.RealmID != "" {
		realmID = &envelope.RealmID
	}
	if err := s.Store.MarkReady(r.Context(), sess.ID, payload, realmID); err != nil {
		s.Logger.Printf("mark ready failed: %v", err)
		s.renderFailure(w, "internal persistence error")
		return
	}

	if err := s.successTemplate.Execute(w, envelope); err != nil {
		s.Logger.Printf("render success error: %v", err)
	}
}

func (s *Server) handlePoll(w http.ResponseWriter, r *http.Request) {
	sessionID := lastPathComponent(r.URL.Path)
	if sessionID == "" {
		http.NotFound(w, r)
		return
	}
	sess, err := s.Store.LoadForPoll(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondJSONError(w, http.StatusNotFound, "session not found")
			return
		}
		s.Logger.Printf("load session error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if time.Now().After(sess.ExpiresAt) {
		_ = s.Store.Delete(r.Context(), sessionID)
		respondJSONError(w, http.StatusGone, "session expired")
		return
	}
	if !sess.ReadyAt.Valid || len(sess.Result) == 0 {
		respondJSON(w, http.StatusOK, map[string]any{"status": "pending"})
		return
	}

	var envelope TokenEnvelope
	if err := json.Unmarshal(sess.Result, &envelope); err != nil {
		s.Logger.Printf("unmarshal session result error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if err := s.Store.Delete(r.Context(), sessionID); err != nil {
		s.Logger.Printf("delete session error: %v", err)
	}
	respondJSON(w, http.StatusOK, envelope)
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Provider     string `json:"provider"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := decodeJSONBody(r.Body, &req); err != nil {
		respondJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	provider := strings.ToLower(req.Provider)
	if provider == "" || req.RefreshToken == "" {
		respondJSONError(w, http.StatusBadRequest, "provider and refresh_token are required")
		return
	}

	var (
		envelope TokenEnvelope
		err      error
	)
	switch provider {
	case "deputy":
		envelope, err = s.refreshDeputy(r.Context(), req.RefreshToken)
	case "qbo":
		envelope, err = s.refreshQBO(r.Context(), req.RefreshToken)
	case "xero":
		envelope, err = s.refreshXero(r.Context(), req.RefreshToken)
	default:
		respondJSONError(w, http.StatusBadRequest, "unsupported provider")
		return
	}
	if err != nil {
		s.Logger.Printf("refresh failed provider=%s error=%v", provider, err)
		respondJSONError(w, http.StatusBadGateway, "token refresh failed")
		return
	}
	envelope.Provider = provider
	respondJSON(w, http.StatusOK, envelope)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) startXeroAuth(state string) (string, sql.NullString, error) {
	verifier, err := randomID(64)
	if err != nil {
		return "", sql.NullString{}, err
	}
	hashed := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hashed[:])

	v := url.Values{}
	v.Set("response_type", "code")
	v.Set("client_id", s.Config.XeroClientID)
	v.Set("redirect_uri", s.Config.XeroRedirectURL)
	v.Set("scope", strings.Join(s.Config.XeroScopes, " "))
	v.Set("state", state)
	v.Set("code_challenge", challenge)
	v.Set("code_challenge_method", "S256")
	authURL := "https://login.xero.com/identity/connect/authorize?" + v.Encode()
	return authURL, sql.NullString{String: verifier, Valid: true}, nil
}

func (s *Server) startDeputyAuth(state string) (string, error) {
	v := url.Values{}
	v.Set("response_type", "code")
	v.Set("client_id", s.Config.DeputyClientID)
	v.Set("redirect_uri", s.Config.DeputyRedirectURL)
	v.Set("scope", strings.Join(s.Config.DeputyScopes, " "))
	v.Set("state", state)
	authURL := "https://once.deputy.com/my/oauth/login?" + v.Encode()
	return authURL, nil
}

func (s *Server) startQBOAuth(state string) (string, error) {
	v := url.Values{}
	v.Set("client_id", s.Config.QBOClientID)
	v.Set("redirect_uri", s.Config.QBORedirectURL)
	v.Set("response_type", "code")
	v.Set("scope", strings.Join(s.Config.QBOScopes, " "))
	v.Set("state", state)
	authURL := "https://appcenter.intuit.com/connect/oauth2?" + v.Encode()
	return authURL, nil
}

func (s *Server) exchangeXero(ctx context.Context, sess *Session, code string) (TokenEnvelope, error) {
	if code == "" {
		return TokenEnvelope{}, fmt.Errorf("missing code")
	}
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", s.Config.XeroRedirectURL)
	data.Set("client_id", s.Config.XeroClientID)
	if sess.CodeVerifier.Valid {
		data.Set("code_verifier", sess.CodeVerifier.String)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://identity.xero.com/connect/token", strings.NewReader(data.Encode()))
	if err != nil {
		return TokenEnvelope{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if s.Config.XeroClientSecret != "" {
		req.SetBasicAuth(s.Config.XeroClientID, s.Config.XeroClientSecret)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return TokenEnvelope{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return TokenEnvelope{}, fmt.Errorf("xero token error: %s", body)
	}
	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		Scope        string `json:"scope"`
		TokenType    string `json:"token_type"`
		IDToken      string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return TokenEnvelope{}, err
	}

	tenants, err := s.fetchXeroConnections(ctx, payload.AccessToken)
	if err != nil {
		s.Logger.Printf("fetch connections failed: %v", err)
	}

	return TokenEnvelope{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second),
		Scope:        payload.Scope,
		TokenType:    payload.TokenType,
		IDToken:      payload.IDToken,
		Tenants:      tenants,
	}, nil
}

func (s *Server) exchangeDeputy(ctx context.Context, code string) (TokenEnvelope, error) {
	if code == "" {
		return TokenEnvelope{}, fmt.Errorf("missing code")
	}
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", s.Config.DeputyClientID)
	data.Set("client_secret", s.Config.DeputyClientSecret)
	data.Set("redirect_uri", s.Config.DeputyRedirectURL)
	data.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://once.deputy.com/my/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return TokenEnvelope{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return TokenEnvelope{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return TokenEnvelope{}, fmt.Errorf("deputy token error: %s", body)
	}
	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		Scope        string `json:"scope"`
		Endpoint     string `json:"endpoint"`
		TokenType    string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return TokenEnvelope{}, err
	}
	return TokenEnvelope{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second),
		Scope:        payload.Scope,
		Endpoint:     payload.Endpoint,
		TokenType:    payload.TokenType,
	}, nil
}

func (s *Server) exchangeQBO(ctx context.Context, code, realmID string) (TokenEnvelope, error) {
	if code == "" {
		return TokenEnvelope{}, fmt.Errorf("missing code")
	}
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", s.Config.QBORedirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer", strings.NewReader(data.Encode()))
	if err != nil {
		return TokenEnvelope{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.Config.QBOClientID, s.Config.QBOClientSecret)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return TokenEnvelope{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return TokenEnvelope{}, fmt.Errorf("qbo token error: %s", body)
	}
	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		XRefresh     int64  `json:"x_refresh_token_expires_in"`
		Scope        string `json:"scope"`
		TokenType    string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return TokenEnvelope{}, err
	}
	env := TokenEnvelope{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second),
		Scope:        payload.Scope,
		TokenType:    payload.TokenType,
		RealmID:      realmID,
	}
	if payload.XRefresh > 0 {
		if env.Raw == nil {
			env.Raw = make(map[string]any)
		}
		env.Raw["refresh_token_expires_in"] = payload.XRefresh
	}
	return env, nil
}

func (s *Server) refreshDeputy(ctx context.Context, refreshToken string) (TokenEnvelope, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", s.Config.DeputyClientID)
	data.Set("client_secret", s.Config.DeputyClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://once.deputy.com/my/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return TokenEnvelope{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return TokenEnvelope{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return TokenEnvelope{}, fmt.Errorf("deputy refresh error: %s", body)
	}
	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		Scope        string `json:"scope"`
		Endpoint     string `json:"endpoint"`
		TokenType    string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return TokenEnvelope{}, err
	}
	return TokenEnvelope{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second),
		Scope:        payload.Scope,
		Endpoint:     payload.Endpoint,
		TokenType:    payload.TokenType,
	}, nil
}

func (s *Server) refreshQBO(ctx context.Context, refreshToken string) (TokenEnvelope, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer", strings.NewReader(data.Encode()))
	if err != nil {
		return TokenEnvelope{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.Config.QBOClientID, s.Config.QBOClientSecret)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return TokenEnvelope{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return TokenEnvelope{}, fmt.Errorf("qbo refresh error: %s", body)
	}
	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		XRefresh     int64  `json:"x_refresh_token_expires_in"`
		Scope        string `json:"scope"`
		TokenType    string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return TokenEnvelope{}, err
	}
	env := TokenEnvelope{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second),
		Scope:        payload.Scope,
		TokenType:    payload.TokenType,
	}
	if payload.XRefresh > 0 {
		if env.Raw == nil {
			env.Raw = make(map[string]any)
		}
		env.Raw["refresh_token_expires_in"] = payload.XRefresh
	}
	return env, nil
}

func (s *Server) refreshXero(ctx context.Context, refreshToken string) (TokenEnvelope, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", s.Config.XeroClientID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://identity.xero.com/connect/token", strings.NewReader(data.Encode()))
	if err != nil {
		return TokenEnvelope{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if s.Config.XeroClientSecret != "" {
		req.SetBasicAuth(s.Config.XeroClientID, s.Config.XeroClientSecret)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return TokenEnvelope{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return TokenEnvelope{}, fmt.Errorf("xero refresh error: %s", body)
	}
	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		Scope        string `json:"scope"`
		TokenType    string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return TokenEnvelope{}, err
	}
	tenants, err := s.fetchXeroConnections(ctx, payload.AccessToken)
	if err != nil {
		s.Logger.Printf("fetch connections failed: %v", err)
	}
	return TokenEnvelope{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second),
		Scope:        payload.Scope,
		TokenType:    payload.TokenType,
		Tenants:      tenants,
	}, nil
}

func (s *Server) fetchXeroConnections(ctx context.Context, accessToken string) ([]XeroTenant, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.xero.com/connections", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("xero connections error: %s", body)
	}
	var tenants []XeroTenant
	if err := json.NewDecoder(resp.Body).Decode(&tenants); err != nil {
		return nil, err
	}
	return tenants, nil
}

func decodeJSONBody(body io.ReadCloser, dst any) error {
	defer body.Close()
	decoder := json.NewDecoder(io.LimitReader(body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	return nil
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	_ = enc.Encode(payload)
}

func respondJSONError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}

func randomID(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func providerFromCallbackPath(p string) string {
	idx := strings.Index(p, "/callback/")
	if idx == -1 {
		return ""
	}
	return strings.Trim(strings.TrimPrefix(p[idx+len("/callback/"):], "/"), "/")
}

func lastPathComponent(p string) string {
	parts := strings.Split(strings.Trim(p, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func (s *Server) renderFailure(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	if err := s.failureTemplate.Execute(w, map[string]string{"Message": msg}); err != nil {
		s.Logger.Printf("render failure template error: %v", err)
	}
}

const successHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Authorisation complete</title>
    <style>
      body { font-family: sans-serif; margin: 2rem; }
      .card { max-width: 520px; padding: 1.5rem; border: 1px solid #ccd; border-radius: 8px; }
      h1 { font-size: 1.6rem; }
    </style>
  </head>
  <body>
    <div class="card">
      <h1>Authorisation complete</h1>
      <p>You can return to the Accounting Ops application to finish setup.</p>
    </div>
  </body>
</html>`

const failureHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Authorisation failed</title>
    <style>
      body { font-family: sans-serif; margin: 2rem; }
      .card { max-width: 520px; padding: 1.5rem; border: 1px solid #fcc; border-radius: 8px; background: #fff5f5; }
      h1 { font-size: 1.6rem; color: #a00; }
      p { color: #333; }
      code { background: #f7f7f7; padding: 0.2rem 0.4rem; border-radius: 4px; }
    </style>
  </head>
  <body>
    <div class="card">
      <h1>Authorisation failed</h1>
      <p>{{ .Message }}</p>
    </div>
  </body>
</html>`
