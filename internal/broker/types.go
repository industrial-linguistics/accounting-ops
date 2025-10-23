package broker

import "time"

// TokenEnvelope is the serialised response handed to CLI clients.
type TokenEnvelope struct {
	Provider     string         `json:"provider"`
	Profile      string         `json:"profile,omitempty"`
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time      `json:"-"`
	ExpiresUnix  int64          `json:"expires_at"`
	Scope        string         `json:"scope,omitempty"`
	RealmID      string         `json:"realmId,omitempty"`
	Endpoint     string         `json:"endpoint,omitempty"`
	TokenType    string         `json:"token_type,omitempty"`
	IDToken      string         `json:"id_token,omitempty"`
	Tenants      []XeroTenant   `json:"tenants,omitempty"`
	Raw          map[string]any `json:"raw,omitempty"`
}

// XeroTenant captures metadata returned by /connections.
type XeroTenant struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenantId"`
	TenantType string    `json:"tenantType"`
	CreatedAt  time.Time `json:"createdDateUtc"`
	UpdatedAt  time.Time `json:"updatedDateUtc"`
	TenantName string    `json:"tenantName"`
}

// MarshalJSON customises expiry serialisation.
func (t TokenEnvelope) MarshalJSON() ([]byte, error) {
	type Alias TokenEnvelope
	a := Alias(t)
	if t.ExpiresUnix == 0 && !t.ExpiresAt.IsZero() {
		a.ExpiresUnix = t.ExpiresAt.Unix()
	}
	a.ExpiresAt = time.Time{}
	return jsonMarshal(a)
}

// UnmarshalJSON recovers expiry timestamps.
func (t *TokenEnvelope) UnmarshalJSON(data []byte) error {
	type Alias TokenEnvelope
	var a Alias
	if err := jsonUnmarshal(data, &a); err != nil {
		return err
	}
	*t = TokenEnvelope(a)
	if a.ExpiresUnix != 0 {
		t.ExpiresAt = time.Unix(a.ExpiresUnix, 0).UTC()
	}
	return nil
}
