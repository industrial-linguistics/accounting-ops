# Broker Environment Configuration Template

This file documents all available environment variables for `broker.env`.

## QuickBooks Online (QBO) Configuration

```bash
# QuickBooks OAuth Credentials
# Get these from: https://developer.intuit.com/ → My Apps → Keys & credentials
QBO_CLIENT_ID=your_client_id_here
QBO_CLIENT_SECRET=your_client_secret_here

# Redirect URI (must match what's registered in Intuit Developer Portal)
QBO_REDIRECT=https://auth.industrial-linguistics.com/v1/callback/qbo

# OAuth Scopes (space-separated)
QBO_SCOPES=com.intuit.quickbooks.accounting

# Environment Mode: "sandbox" or "production" (default: production)
# - sandbox: Uses sandbox-quickbooks.api.intuit.com for API calls
# - production: Uses quickbooks.api.intuit.com for API calls
# Note: OAuth endpoints (appcenter.intuit.com, oauth.platform.intuit.com) are the same for both
QBO_ENVIRONMENT=sandbox

# Optional: Override OAuth authorization URL
# QBO_AUTH_URL=https://appcenter.intuit.com/connect/oauth2

# Optional: Override OAuth token exchange URL
# QBO_TOKEN_URL=https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer

# Optional: Override API base URL
# QBO_API_BASE_URL=https://sandbox-quickbooks.api.intuit.com
```

## Xero Configuration

```bash
# Xero OAuth Credentials
# Get these from: https://developer.xero.com/ → My Apps
XERO_CLIENT_ID=your_client_id_here

# Optional: Client secret (not required for PKCE flow, but supported)
# XERO_CLIENT_SECRET=your_client_secret_here

# Redirect URI (must match what's registered in Xero Developer Portal)
XERO_REDIRECT=https://auth.industrial-linguistics.com/v1/callback/xero

# OAuth Scopes (space-separated)
XERO_SCOPES=offline_access accounting.transactions accounting.contacts

# Environment Mode: "production" (default: production)
# Note: Xero doesn't have a separate sandbox mode in the same way QB does
# You can connect to demo companies in production
XERO_ENVIRONMENT=production

# Optional: Override OAuth authorization URL
# XERO_AUTH_URL=https://login.xero.com/identity/connect/authorize

# Optional: Override OAuth token exchange URL
# XERO_TOKEN_URL=https://identity.xero.com/connect/token

# Optional: Override API base URL
# XERO_API_BASE_URL=https://api.xero.com
```

## Deputy Configuration

```bash
# Deputy OAuth Credentials
# Get these from: https://www.deputy.com/api-doc/API/Getting_Started
DEPUTY_CLIENT_ID=your_client_id_here
DEPUTY_CLIENT_SECRET=your_client_secret_here

# Redirect URI (must match what's registered in Deputy)
DEPUTY_REDIRECT=https://auth.industrial-linguistics.com/v1/callback/deputy

# OAuth Scopes (space-separated)
# Note: Always include longlife_refresh_token for refresh token support
DEPUTY_SCOPES=longlife_refresh_token

# Environment Mode: "production" (default: production)
# Note: Deputy may have sandbox/test environments - check their docs
DEPUTY_ENVIRONMENT=production

# Optional: Override OAuth authorization URL
# DEPUTY_AUTH_URL=https://once.deputy.com/my/oauth/login

# Optional: Override OAuth token exchange URL
# DEPUTY_TOKEN_URL=https://once.deputy.com/my/oauth/access_token
```

## Security Configuration

```bash
# Master Key for Session Encryption (recommended for production)
# Generate a random 32+ character string
# Example: openssl rand -base64 32
BROKER_MASTER_KEY=your_random_32_byte_key_here
```

## Session Management

```bash
# Session TTL in seconds (default: 600 = 10 minutes)
# How long OAuth sessions stay valid before expiring
SESSION_TTL_SECONDS=600

# Poll timeout in seconds (default: 5)
# How long to wait before returning "pending" on poll requests
POLL_TIMEOUT_SECONDS=5
```

## Rate Limiting

```bash
# Rate limit for /v1/auth/start endpoint
RATE_LIMIT_AUTH_START=10
RATE_LIMIT_AUTH_START_WINDOW_SECONDS=60

# Rate limit for /v1/auth/poll endpoint
RATE_LIMIT_POLL=120
RATE_LIMIT_POLL_WINDOW_SECONDS=60

# Rate limit for /v1/token/refresh endpoint
RATE_LIMIT_REFRESH=60
RATE_LIMIT_REFRESH_WINDOW_SECONDS=60
```

---

## Example Configurations

### Development (Sandbox) Configuration

```bash
# broker.env - Development/Sandbox Configuration

# QuickBooks - Sandbox
QBO_CLIENT_ID=AByy3Gi73gqarhqmV7iSYccpLxCGo1ry1yJwHHBiM2OnbGrweP
QBO_CLIENT_SECRET=mG1Vo0kXIlDq5wfggH5SjCwy3FKPUnH0LkeGG3bh
QBO_REDIRECT=https://auth.industrial-linguistics.com/v1/callback/qbo
QBO_SCOPES=com.intuit.quickbooks.accounting
QBO_ENVIRONMENT=sandbox

# Xero - Production (connects to demo companies)
XERO_CLIENT_ID=your_xero_dev_client_id
XERO_REDIRECT=https://auth.industrial-linguistics.com/v1/callback/xero
XERO_SCOPES=offline_access accounting.transactions accounting.contacts
XERO_ENVIRONMENT=production

# Deputy - Production
DEPUTY_CLIENT_ID=your_deputy_dev_client_id
DEPUTY_CLIENT_SECRET=your_deputy_dev_client_secret
DEPUTY_REDIRECT=https://auth.industrial-linguistics.com/v1/callback/deputy
DEPUTY_SCOPES=longlife_refresh_token
DEPUTY_ENVIRONMENT=production

# Security
BROKER_MASTER_KEY=dev_key_not_for_production_use_12345
```

### Production Configuration

```bash
# broker.env - Production Configuration

# QuickBooks - Production
QBO_CLIENT_ID=your_production_qbo_client_id
QBO_CLIENT_SECRET=your_production_qbo_client_secret
QBO_REDIRECT=https://auth.industrial-linguistics.com/v1/callback/qbo
QBO_SCOPES=com.intuit.quickbooks.accounting
QBO_ENVIRONMENT=production

# Xero - Production
XERO_CLIENT_ID=your_production_xero_client_id
XERO_REDIRECT=https://auth.industrial-linguistics.com/v1/callback/xero
XERO_SCOPES=offline_access accounting.transactions accounting.contacts
XERO_ENVIRONMENT=production

# Deputy - Production
DEPUTY_CLIENT_ID=your_production_deputy_client_id
DEPUTY_CLIENT_SECRET=your_production_deputy_client_secret
DEPUTY_REDIRECT=https://auth.industrial-linguistics.com/v1/callback/deputy
DEPUTY_SCOPES=longlife_refresh_token
DEPUTY_ENVIRONMENT=production

# Security - Use a strong random key!
BROKER_MASTER_KEY=$(openssl rand -base64 32)
```

---

## QuickBooks Environment Details

### Sandbox vs Production

| Aspect | Sandbox | Production |
|--------|---------|----------|
| **Client ID/Secret** | From "Development Keys" in developer portal | From "Production Keys" in developer portal |
| **OAuth Auth URL** | https://appcenter.intuit.com/connect/oauth2 (same) | https://appcenter.intuit.com/connect/oauth2 (same) |
| **OAuth Token URL** | https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer (same) | https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer (same) |
| **API Base URL** | https://sandbox-quickbooks.api.intuit.com | https://quickbooks.api.intuit.com |
| **Companies** | Test sandbox companies only | Real customer companies |
| **Data** | Fake/test data | Real financial data |

**Key Point:** OAuth endpoints are the same for sandbox and production. The difference is:
1. Which credentials you use (sandbox keys vs production keys)
2. Which API base URL you hit for actual API calls

The `QBO_ENVIRONMENT` variable controls which API base URL to use when making QuickBooks API calls (not implemented in current broker, but planned for future when broker makes API calls).

---

## URL Override Examples

If a provider changes their OAuth endpoints or you need to use a different environment:

```bash
# Override QuickBooks to use a test environment
QBO_AUTH_URL=https://appcenter-e2e.intuit.com/connect/oauth2
QBO_TOKEN_URL=https://oauth-e2e.platform.intuit.com/oauth2/v1/tokens/bearer

# Override Xero to use a demo environment (if one exists)
XERO_AUTH_URL=https://login-demo.xero.com/identity/connect/authorize
XERO_TOKEN_URL=https://identity-demo.xero.com/connect/token

# Override Deputy to use a staging environment
DEPUTY_AUTH_URL=https://staging.deputy.com/my/oauth/login
DEPUTY_TOKEN_URL=https://staging.deputy.com/my/oauth/access_token
```

---

## Deployment Checklist

- [ ] Obtain OAuth credentials from provider developer portals
- [ ] Set correct environment mode (sandbox for testing, production for live)
- [ ] Verify redirect URIs match what's registered in provider portals
- [ ] Generate a strong BROKER_MASTER_KEY for production
- [ ] Set appropriate rate limits based on expected usage
- [ ] Test with sandbox/development credentials first
- [ ] Switch to production credentials when ready to go live
- [ ] Keep broker.env file permissions restrictive (640, owned by root:www)
- [ ] Never commit broker.env to version control
