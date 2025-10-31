# Authentication Broker Architecture

## Design Overview
First-run experiences for both the CLI and optional GUI open a browser to a hosted broker at `https://auth.industrial-linguistics.com`. The broker completes OAuth, returns tokens to the CLI via a short-lived, one-time channel, and centrally handles refresh flows that require a client secret (Deputy, QuickBooks Online). Xero Auth Code + PKCE refresh is performed locally by the client.

## Why the Hosted Broker Is Mandatory
- **QuickBooks Online (QBO)** requires HTTPS redirect URIs in production; localhost or direct IP addresses are disallowed. The OAuth callback also returns the `realmId`, so the hosted broker must capture and return it. Access tokens last ~1 hour, refresh tokens up to 100 days with rotation; the newest refresh token must always be stored.
- **Deputy** OAuth 2.0 uses `https://once.deputy.com`. Access tokens last ~24 hours. Refresh flows require the client secret and rotate the refresh token while also returning the customer endpoint (subdomain); the broker must perform these refreshes and pass back the newest values.
- **Xero** supports Auth Code with PKCE for native clients. Access tokens last 30 minutes. Refresh tokens expire if unused for 60 days and rotate on refresh. Every API call must include the `xero-tenant-id` header, discovered via the `/connections` API. Xero also limits uncertified apps to 25 tenant connections total and a maximum of two uncertified apps per organisation, so App Store certification is required for broad distribution.

## Major Components
1. **Broker CGI (Go)** running on Merah under OpenBSD `httpd` + `slowcgi`, chrooted to `/var/www`.
2. **CLI (`acct`, Go)** shipping as cross-platform binaries for customer use.
3. **First-run GUI (optional, Go)** implemented with Fyne/Wails and calling the same broker endpoints as the CLI.

## Domains, DNS, and TLS
- Create `auth.industrial-linguistics.com` with A/AAAA records pointing to Merah.
- Use `acme-client(1)` for TLS issuance. Leave TCP 80 open for the ACME HTTP-01 challenge.
- Provide chrooted DNS/CA data for outbound HTTPS from CGI binaries:
  ```sh
  install -d /var/www/etc /var/www/etc/ssl
  cp -p /etc/resolv.conf /var/www/etc/resolv.conf
  cp -p /etc/ssl/cert.pem /var/www/etc/ssl/cert.pem
  ```

## OpenBSD Layout and Server Configuration
```
/var/www/
  vhosts/auth.industrial-linguistics.com/
    htdocs/                          # static success/failure pages
    v1/broker                        # Go CGI binary
    data/broker.sqlite               # SQLite DB
    logs/broker.log
    conf/broker.env                  # client IDs/secrets, master key
    tmp/                             # scratch
```
- `v1/broker`: `0755`, owner `aops:daemon`.
- `data/` directory and `broker.sqlite`: owner `www:www`, modes `0700/0600`.
- `conf/broker.env`: owner `root:www`, mode `0640`.
- Ensure `/var/www/dev/{null,zero,random,urandom}` exist (default install).

### httpd + slowcgi
Enable and start the services:
```sh
rcctl enable httpd slowcgi && rcctl start httpd slowcgi
```
Example `httpd.conf`:
```nginx
types { include "/usr/share/misc/mime.types" }
server "auth.industrial-linguistics.com" {
  listen on * tls port 443
  root "/vhosts/auth.industrial-linguistics.com/htdocs"
  tls { certificate "/etc/ssl/auth.industrial-linguistics.com.pem"
        key "/etc/ssl/private/auth.industrial-linguistics.com.key" }
  location "/cgi-bin/*" {
    fastcgi
    fastcgi socket "/run/slowcgi.sock"
    root "/vhosts/auth.industrial-linguistics.com"
  }
  location "/.well-known/acme-challenge/*" {
    root "/acme"
    request strip 2
  }
}
server "auth.industrial-linguistics.com" {
  listen on * port 80
  root "/vhosts/auth.industrial-linguistics.com/htdocs"
  location "/.well-known/acme-challenge/*" { root "/acme"; request strip 2 }
  block return 301 "https://auth.industrial-linguistics.com$REQUEST_URI"
}
```
`httpd` runs chrooted to `/var/www` and serves CGI via the `slowcgi` socket (`/run/slowcgi.sock`).

## Broker CGI Behaviour
- Primary goal: start OAuth, receive callbacks, exchange codes for tokens, and hand tokens to the CLI via a one-time poll. Long-term tokens are never stored server-side.
- For Deputy and QBO refresh, expose `/v1/token/refresh` that takes the client’s refresh token and returns rotated tokens using stored client secrets. Xero PKCE refresh can remain local.
- Persist only short-lived session state in SQLite.

### Endpoints (JSON)
- `POST /v1/broker/v1/auth/start`
  - Body: `{ "provider":"xero|deputy|qbo", "profile":"string", "pubkey":"base64(optional)" }`
  - Response: `{ "auth_url":"…", "poll_url":"/v1/broker/v1/auth/poll/{session}", "session":"id" }`
  - Server creates state, PKCE verifier (if applicable), and records a session row.
- `GET /v1/callback/{provider}`
  - Validates state. For QBO, capture `realmId`. Exchanges code for tokens, persists tokens inside the session, marks `ready_at`, and renders a success page.
- `GET /v1/broker/v1/auth/poll/{session}`
  - Performs long or short polling. Returns tokens once ready, then deletes or tombstones them.
- `POST /v1/broker/v1/token/refresh`
  - Body: `{ "provider":"deputy|qbo|xero", "refresh_token":"…" }`
  - Uses provider secrets when required and returns rotated tokens. Xero PKCE refresh does not need a secret.
- `GET /v1/broker/healthz` → `200 OK`.

### Provider-Specific Notes
- **Xero**: Use S256 PKCE. After token exchange, call `/connections` to list tenants so the CLI can select and store the `xero-tenant-id` for API calls. Access tokens last 30 minutes; refresh tokens expire after 60 days of inactivity and must be rotated.
- **Deputy**: Start URL `https://once.deputy.com/my/oauth/login?...&scope=longlife_refresh_token`. Exchange at `/my/oauth/access_token`. Response returns `{ access_token, expires_in, scope, endpoint, refresh_token }`. Refresh requires the client secret and rotates the refresh token.
- **QuickBooks Online**: Start URL `https://appcenter.intuit.com/connect/oauth2?...` with scope `com.intuit.quickbooks.accounting` (add OpenID scopes only when identity data is required). Production redirect URIs must be HTTPS, no localhost/IP. Callback includes `realmId`. Access tokens ~1 hour, refresh tokens 100 days rolling and rotate; persist the newest value. Token endpoint per Intuit discovery docs.

### Transport Security
- Enforce TLS everywhere.
- Optionally encrypt poll payloads using the CLI’s ephemeral public key with `nacl/box`; otherwise rely on TLS plus single-use sessions.

### SQLite Schema
```sql
CREATE TABLE auth_session (
  id TEXT PRIMARY KEY,
  provider TEXT NOT NULL,
  state TEXT NOT NULL,
  code_verifier TEXT,
  realm_id TEXT,
  created_at INTEGER NOT NULL,
  expires_at INTEGER NOT NULL,
  ready_at INTEGER,
  result_cipher BLOB
);
CREATE INDEX idx_auth_session_exp ON auth_session(expires_at);
```
- Store only session state and short-lived results. Do **not** store client secrets in SQLite; load them from `conf/broker.env`.

### Broker Configuration (`conf/broker.env`)
```
XERO_CLIENT_ID=...
# Optional secret for web-app registration
XERO_CLIENT_SECRET=...
XERO_REDIRECT=https://auth.industrial-linguistics.com/v1/callback/xero

DEPUTY_CLIENT_ID=...
DEPUTY_CLIENT_SECRET=...
DEPUTY_REDIRECT=https://auth.industrial-linguistics.com/v1/callback/deputy

QBO_CLIENT_ID=...
QBO_CLIENT_SECRET=...
QBO_REDIRECT=https://auth.industrial-linguistics.com/v1/callback/qbo

BROKER_MASTER_KEY=base64-32B-aesgcm
```

At runtime the broker accepts environment overrides:

* `BROKER_ENV_PATH` — custom path to the configuration file (defaults to `conf/broker.env`).
* `BROKER_DB_PATH` — custom SQLite path (defaults to `data/broker.sqlite`).
* When running the CGI binary in standalone HTTP mode, the flags `-env`, `-db`, and `-addr` provide equivalent overrides for local testing.

### Implementation Notes
- Use `net/http/cgi` with a small router parsing `PATH_INFO`.
- Configure HTTP clients with sane timeouts and trust `/etc/ssl/cert.pem` inside the chroot.
- Enable SQLite WAL mode, `busy_timeout=5000`, and `PRAGMA journal_mode=WAL`.
- Emit structured logs, redact tokens, and log session IDs only.

## CLI (`acct`) Behaviour
- `acct connect xero|deputy|qbo --profile NAME`
  - Calls `/v1/auth/start`, opens the browser, polls for completion, and displays connected org info.
  - Xero: list tenants via `/connections`, prompt for selection, persist `xero-tenant-id`.
  - Deputy: persist returned endpoint (customer subdomain).
  - QBO: persist `realmId`.
- `acct list` — list profiles.
- `acct whoami --profile NAME` — quick API probe.
- `acct refresh --profile NAME`
  - Xero: refresh locally via PKCE.
  - Deputy/QBO: call broker `/v1/token/refresh`.
- `acct revoke --profile NAME` — forget local credentials and instruct users to revoke vendor-side if required.

Environment requirements for refresh flows:

* Export `XERO_CLIENT_ID` (and optionally `XERO_CLIENT_SECRET`) before running `acct refresh --provider xero` so the CLI can perform the PKCE refresh locally.
* Deputy and QBO refreshes continue to proxy through the broker and therefore use the secrets stored in `broker.env`.

### Token Storage
Use the OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service). Store per-profile payloads:
- **Xero**: `{ access_token, refresh_token, expires_at, xero_tenant_id, scopes }`
- **Deputy**: `{ access_token, refresh_token, expires_at, endpoint }`
- **QBO**: `{ access_token, refresh_token, expires_at, realmId, scopes }`

### User Experience Example
```
$ acct connect qbo --profile acme
Opening browser...
Waiting for authorisation... done.
Connected: ACME Pty Ltd (realmId=1234567890)
```

## First-run GUI (Optional)
A simple three-step wizard: pick system → open browser → confirm connection. Uses the same `/v1/auth/start` and `/v1/auth/poll` endpoints. After connection, display discovered tenant/company/endpoints and a “Done” action.

## Vendor Setup Checklists
### Xero
1. Create a developer account and application. Choose Auth Code with PKCE for native flows or Web App for server-side secret usage. Configure redirect URI `https://auth.industrial-linguistics.com/v1/callback/xero`.
2. Request only required scopes plus `offline_access` for refresh.
3. Every API call requires the `xero-tenant-id` header; discover tenants via `/connections`.
4. Respect uncertified caps (25 tenant connections total; two uncertified apps per organisation). Plan for Xero App Store certification for broad rollout.

### Deputy
1. Create a trial account and Once profile, then register an OAuth client at `once.deputy.com`.
2. Redirect URI: `https://auth.industrial-linguistics.com/v1/callback/deputy`.
3. Always request scope `longlife_refresh_token`. Refresh exchanges require the client secret and rotate the refresh token. Token responses include the endpoint domain that must be stored.

### QuickBooks Online (Intuit)
1. Create an Intuit Developer app. Register only HTTPS redirects: `https://auth.industrial-linguistics.com/v1/callback/qbo` (no localhost or IPs in production).
2. Scope: `com.intuit.quickbooks.accounting` (add OpenID scopes only when identity is needed).
3. Expect `realmId` in the callback. Store it per profile for API base paths. Access tokens last ~1 hour; refresh tokens rotate and are valid up to 100 days.

## Security Model
- No client secrets in the CLI; secrets reside in `broker.env` only.
- Always use state + PKCE where supported. Xero PKCE is explicitly supported.
- Rotate refresh tokens on every refresh (Deputy, QBO, Xero).
- Request minimal scopes.
- Enforce HTTPS across all transport.
- Optional NaCl encryption of poll payloads using CLI ephemeral keys.

## Deployment Pipeline
```
/cmd/broker/         # Go CGI broker
/cmd/acct/           # CLI
/web/                # static success/failure pages
/scripts/build_deploy_openbsd.sh
```

### Remote Build Script (`scripts/build_deploy_openbsd.sh`)
```sh
set -euo pipefail
cd /var/www/vhosts/auth.industrial-linguistics.com/accounting-ops
git pull --ff-only
export CGO_ENABLED=0
( cd cmd/broker && go build -trimpath -ldflags="-s -w" -o ../../v1/broker )
install -d -o root -g daemon -m 0755 /var/www/vhosts/auth.industrial-linguistics.com/v1
install -o aops -g daemon -m 0755 v1/broker /var/www/vhosts/auth.industrial-linguistics.com/v1/broker
install -d -o www -g www -m 0700 /var/www/vhosts/auth.industrial-linguistics.com/data
touch /var/www/vhosts/auth.industrial-linguistics.com/data/broker.sqlite
chown www:www /var/www/vhosts/auth.industrial-linguistics.com/data/broker.sqlite
# Ensure chroot DNS/CA
install -d /var/www/etc /var/www/etc/ssl
cp -p /etc/resolv.conf /var/www/etc/resolv.conf
cp -p /etc/ssl/cert.pem /var/www/etc/ssl/cert.pem
echo "deployed $(date)"
```

### GitHub Actions Workflow (`.github/workflows/deploy.yml`)
```yaml
name: deploy-merah
on: { push: { branches: [ "main" ] }, workflow_dispatch: {} }
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: webfactory/ssh-agent@v0.9.0
        with: { ssh-private-key: ${{ secrets.DEPLOYMENT_SSH_KEY }} }
      - run: |
          mkdir -p ~/.ssh
          ssh-keyscan -H merah.cassia.ifost.org.au >> ~/.ssh/known_hosts
      - run: |
          ssh aops@merah.cassia.ifost.org.au \
            'cd /var/www/vhosts/auth.industrial-linguistics.com/accounting-ops && ./scripts/build_deploy_openbsd.sh'
```

## HTTP Examples
- **Start (CLI → Broker)**
  ```http
  POST /v1/broker/v1/auth/start
  { "provider":"qbo", "profile":"acme" }
  → { "auth_url":"https://appcenter.intuit.com/connect/oauth2?...",
      "poll_url":"/v1/broker/v1/auth/poll/3f9c...", "session":"3f9c..." }
  ```
- **Poll (CLI → Broker)**
  ```http
  GET /v1/broker/v1/auth/poll/3f9c...
  → { "access_token":"...", "refresh_token":"...", "expires_at":1699999999,
       "provider":"qbo", "realmId":"1234567890", "scopes":"..." }
  ```
- **Refresh (CLI → Broker)**
  ```http
  POST /v1/broker/v1/token/refresh
  { "provider":"qbo", "refresh_token":"..." }
  → { "access_token":"...", "refresh_token":"...", "expires_at":... }
  ```

## Error Handling Surfaced to Users
- QBO: "Redirect URI must be HTTPS; localhost/IP rejected."
- Deputy: "Refresh token rotated; store the new refresh token."
- Xero: "Add xero-tenant-id header; select a tenant via /connections."
- Xero limits: "Uncertified cap reached (25 total or 2 per org)."

## Operations Runbook
- **ACME renewals**: schedule `acme-client` and send `SIGHUP` to `httpd`.
- **Logs**: rotate with `newsyslog`.
- **Backups**: `sqlite3 broker.sqlite ".backup '/backup/broker-$(date).db'"`.
- **Chroot outages**: missing `/var/www/etc/resolv.conf` or CA bundle causes DNS/TLS failures; copy both to restore service.

## Vendor-Specific Callouts (Must Follow)
- Xero PKCE is supported for native apps; access 30 min; refresh expires if unused for 60 days; rotate on refresh; every API call needs `xero-tenant-id`.
- Deputy OAuth via `once.deputy.com`; always request `longlife_refresh_token`; endpoint domain returned; refresh requires client secret and rotates tokens.
- QBO requires HTTPS redirect, returns `realmId`; access ~1 hour; refresh tokens valid up to 100 days and rotate.

## Sign-Up Checklist Summary
- **Xero**: Developer account → OAuth 2.0 app (PKCE or Web) → add redirect → request scopes (incl. `offline_access`) → note uncertified limits and plan certification.
- **Deputy**: Trial + Once profile → create OAuth client → add redirect → request `longlife_refresh_token` scope.
- **QBO**: Intuit Developer account → create app → add HTTPS redirect only → request `com.intuit.quickbooks.accounting` scope.

## Sample OpenBSD Configuration
`/etc/acme-client.conf`:
```conf
authority letsencrypt {
  api url "https://acme-v02.api.letsencrypt.org/directory"
  account key "/etc/acme/letsencrypt-privkey.pem"
}
domain auth.industrial-linguistics.com {
  domain key "/etc/ssl/private/auth.industrial-linguistics.com.key"
  domain full chain certificate "/etc/ssl/auth.industrial-linguistics.com.pem"
  sign with letsencrypt
}
```
Enable with:
```sh
rcctl enable acme_client && acme-client -v auth.industrial-linguistics.com && rcctl restart httpd
```

## Test Plan Essentials
- Happy paths: connect Xero, Deputy, and QBO; ensure tenant/realm/endpoint data persists locally.
- Expiry handling: simulate access token expiry; verify refresh flows (Xero local, Deputy/QBO via broker) rotate and persist new refresh tokens.
- Limits: surface the Xero "uncertified app limit reached" error cleanly.
- Chroot validation: temporarily remove chroot DNS/CA files to confirm failure paths, then restore using the commands above.

## Non-Negotiables
- No client secrets in the CLI.
- Broker stores only session data, not customer refresh tokens.
- HTTPS everywhere.
- Minimal scopes.
- Always rotate refresh tokens and store the newest.
