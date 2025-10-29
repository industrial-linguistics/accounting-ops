# QuickBooks Interface Deployment Checklist

This checklist ensures all components are properly deployed and configured for the QuickBooks Online integration.

## Server Access & Environment

- [ ] **SSH Access Verified**
  ```bash
  ssh aops@merah.cassia.ifost.org.au
  ```
  - Confirm access credentials are available
  - Verify sudo/doas privileges if required

- [ ] **Server Platform Confirmed**
  - OS: OpenBSD (expected)
  - Web server: httpd with slowcgi
  - Verify server version: `uname -a`

## OAuth Broker Deployment

### CGI Binary

- [ ] **Broker Binary Installed**
  ```bash
  # Check broker binary exists and is executable
  ls -la /var/www/vhosts/auth.industrial-linguistics.com/cgi-bin/broker
  file /var/www/vhosts/auth.industrial-linguistics.com/cgi-bin/broker
  ```
  - Binary should be statically linked (CGO_ENABLED=0)
  - Verify ownership and permissions (typically www:www or _www:_www)
  - Check binary is not stripped if debugging needed

- [ ] **Deployment Script Available**
  - Script location: `scripts/build_deploy_openbsd.sh`
  - Review script for current deployment procedures
  - Verify script has correct SSH host and paths

### Configuration Files

- [ ] **Broker Configuration Present**
  ```bash
  # Check broker.env file
  ls -la /var/www/vhosts/auth.industrial-linguistics.com/conf/broker.env
  cat /var/www/vhosts/auth.industrial-linguistics.com/conf/broker.env
  ```
  - QuickBooks OAuth client ID configured
  - QuickBooks OAuth client secret configured (never commit to repo)
  - Database path configured for session storage
  - Verify environment variable names match broker code

- [ ] **QuickBooks OAuth Credentials Valid**
  - Client ID from Intuit Developer Portal
  - Client Secret from Intuit Developer Portal
  - Scopes: `com.intuit.quickbooks.accounting` (minimum)
  - Redirect URI: `https://auth.industrial-linguistics.com/v1/callback/qbo`
  - Verify redirect URI is registered in Intuit Developer Portal

### Chroot Environment

- [ ] **Chroot DNS Configuration**
  ```bash
  # Verify DNS resolution works in chroot
  cat /var/www/etc/resolv.conf
  ```
  - Should contain valid nameservers
  - Test DNS from within chroot if possible

- [ ] **Chroot CA Certificates**
  ```bash
  # Verify SSL/TLS CA bundle
  ls -la /var/www/etc/ssl/cert.pem
  ```
  - Required for HTTPS requests to Intuit APIs
  - Should be up-to-date CA bundle

- [ ] **Session Database**
  ```bash
  # Check SQLite database for sessions
  ls -la /var/www/vhosts/auth.industrial-linguistics.com/data/
  ```
  - Directory must be writable by www user
  - Database will be created on first use
  - Verify permissions allow broker to write

## Web Server Configuration

- [ ] **httpd Configuration**
  ```bash
  # Review httpd.conf for broker vhost
  cat /etc/httpd.conf | grep -A 20 "auth.industrial-linguistics.com"
  ```
  - Virtual host configured for auth.industrial-linguistics.com
  - CGI enabled with slowcgi
  - FastCGI socket path correct

- [ ] **slowcgi Running**
  ```bash
  # Check slowcgi service status
  rcctl check slowcgi
  ps aux | grep slowcgi
  ```
  - Service should be enabled and running
  - Verify socket location matches httpd config

- [ ] **httpd Service Status**
  ```bash
  rcctl check httpd
  ```
  - Service should be enabled and running
  - Check logs for errors: `tail -f /var/www/logs/error_log`

## TLS/SSL Configuration (Cloudflare)

- [ ] **Cloudflare TLS Certificate Valid**
  ```bash
  # Check certificate expiration (Cloudflare-issued)
  echo | openssl s_client -connect auth.industrial-linguistics.com:443 2>/dev/null | openssl x509 -noout -dates
  ```
  - Certificate not expired
  - Issued for correct domain
  - Chain valid and trusted
  - Note: Certificate is managed by Cloudflare, not origin server

- [ ] **Cloudflare SSL/TLS Mode**
  - Verify SSL/TLS encryption mode in Cloudflare dashboard
  - Recommended: "Full" or "Full (strict)" mode
  - "Flexible" mode (Cloudflare↔origin uses HTTP) acceptable if origin is secured
  - Ensure mode matches origin server configuration (HTTP vs HTTPS)

- [ ] **Origin Server Configuration**
  ```bash
  # Check if origin serves HTTP or HTTPS
  # This is internal to the infrastructure
  ```
  - Origin may use HTTP (Cloudflare terminates TLS)
  - Or origin may use HTTPS (Full/Full strict mode)
  - No acme-client or Let's Encrypt needed on origin
  - httpd config should match Cloudflare SSL mode

- [ ] **HTTPS Redirect**
  - Cloudflare automatically redirects HTTP to HTTPS
  - Test: `curl -I http://auth.industrial-linguistics.com`
  - Should see 301/302 redirect to HTTPS

## Broker Endpoints Testing

- [ ] **Health Check (if implemented)**
  ```bash
  curl https://auth.industrial-linguistics.com/health
  ```

- [ ] **OAuth Start Endpoint**
  ```bash
  # Test auth initiation (should return auth URL)
  curl -X POST https://auth.industrial-linguistics.com/v1/auth/start \
    -H "Content-Type: application/json" \
    -d '{"provider":"qbo","profile":"test"}'
  ```
  - Returns session ID
  - Returns authorization URL
  - Returns poll URL

- [ ] **OAuth Callback Handling**
  - Callback endpoint: `/v1/callback/qbo`
  - Accepts authorization code from Intuit
  - Extracts `realmId` from query parameters
  - Stores session data correctly

- [ ] **Poll Endpoint**
  ```bash
  # Test polling (with valid session ID)
  curl https://auth.industrial-linguistics.com/v1/auth/poll/{session_id}
  ```
  - Returns pending status while waiting
  - Returns tokens + realmId once OAuth completes

- [ ] **Token Refresh Endpoint**
  ```bash
  # Test token refresh
  curl -X POST https://auth.industrial-linguistics.com/v1/token/refresh \
    -H "Content-Type: application/json" \
    -d '{"provider":"qbo","refresh_token":"<token>"}'
  ```
  - Accepts refresh token
  - Returns new access token + refresh token
  - New tokens rotated (old refresh token invalidated)

## Client Tools Deployment

### Go CLI (`acct`)

- [ ] **CLI Binary Built**
  ```bash
  # Build acct CLI
  cd cmd/acct
  go build -trimpath -ldflags="-s -w" -o acct
  ```
  - Binary available for macOS/Linux/Windows
  - Verify version matches VERSION file

- [ ] **CLI Installation**
  - Binary installed to `~/.local/bin/acct` or system path
  - Executable permissions set
  - Test: `acct --help`

- [ ] **CLI OAuth Flow**
  ```bash
  # Test QuickBooks OAuth via CLI
  acct connect qbo --profile "Test Company"
  ```
  - Opens browser with Intuit auth page
  - Redirects to broker callback
  - Polls broker for completion
  - Stores tokens in OS keychain
  - Stores realmId with credentials

- [ ] **CLI Token Refresh**
  ```bash
  acct refresh --profile "Test Company"
  ```
  - Fetches refresh token from keychain
  - Calls broker refresh endpoint
  - Updates keychain with new tokens

### C++ Qt Tools

- [ ] **Qt Tools Built**
  ```bash
  cmake -S . -B build -DCMAKE_BUILD_TYPE=Release
  cmake --build build
  ```
  - All binaries compile without errors
  - QuickBooks tool: `build/tools/quickbooks_tool/quickbooks_tool`

- [ ] **Skills Library**
  - `quickbooks.skill.json` present in `skills/data/`
  - Skill data installed to `share/accounting-ops/skills/`
  - Loaded correctly by SkillRepository

- [ ] **First Run Tools**
  - `first_run_gui_tool` - GUI wizard
  - `first_run_cli_tool` - CLI wizard
  - Both integrate with `acct` CLI for QuickBooks setup
  - Credentials saved to SQLite: `config/credentials.sqlite`

- [ ] **QuickBooks Tool Testing**
  ```bash
  ./build/tools/quickbooks_tool/quickbooks_tool
  ```
  - Loads credentials from SQLite
  - Displays connection test widget
  - Can verify API connectivity
  - Shows realmId and company info

## QuickBooks API Configuration

- [ ] **Intuit Developer Account**
  - App created in Intuit Developer Portal
  - App type: Desktop or Web App (not Mobile)
  - OAuth 2.0 configured

- [ ] **Redirect URI Registered**
  - Production URI: `https://auth.industrial-linguistics.com/v1/callback/qbo`
  - URI must be HTTPS (no localhost/IP allowed in production)
  - URI exactly matches broker callback endpoint

- [ ] **OAuth Scopes**
  - `com.intuit.quickbooks.accounting` - minimum required
  - Additional scopes only if needed (OpenID, Payments, etc.)

- [ ] **API Endpoints**
  - Sandbox: `https://sandbox-quickbooks.api.intuit.com`
  - Production: `https://quickbooks.api.intuit.com`
  - Broker configured for correct environment

- [ ] **RealmId Handling**
  - Extracted from OAuth callback query parameters
  - Stored with credentials for API requests
  - Used in API base paths: `https://quickbooks.api.intuit.com/v3/company/{realmId}/...`

## Credential Storage

- [ ] **OS Keychain (Go CLI)**
  - macOS: Keychain Access stores tokens
  - Linux: Secret Service API
  - Windows: Credential Manager
  - Test storage: `acct list`

- [ ] **SQLite Database (Qt Tools)**
  ```bash
  # Verify credentials database
  ls -la config/credentials.sqlite
  sqlite3 config/credentials.sqlite ".schema"
  ```
  - Database created on first use
  - Schema includes realmId field for QuickBooks
  - Credentials readable by Qt tools

## Security Verification

- [ ] **Client Secrets**
  - NEVER in Qt tools or CLI binaries
  - ONLY in `conf/broker.env` on server
  - Environment file not in git
  - File permissions restrict access (600 or 640)

- [ ] **Token Storage**
  - No long-lived refresh tokens in broker database
  - Broker session data expires after OAuth completion
  - OS keychain encrypted by platform

- [ ] **HTTPS Enforcement**
  - All broker traffic over HTTPS
  - No HTTP endpoints accept credentials
  - TLS 1.2+ enforced

- [ ] **Minimal Scopes**
  - Only request `com.intuit.quickbooks.accounting`
  - Don't request OpenID unless needed

## Testing & Validation

- [ ] **End-to-End OAuth Flow**
  1. Run `acct connect qbo --profile "Test"`
  2. Browser opens to Intuit
  3. Authorize app
  4. Redirect to broker
  5. CLI polls and receives tokens
  6. Tokens stored in keychain
  7. Verify: `acct whoami --profile "Test"`

- [ ] **Token Refresh Flow**
  1. Wait for access token to expire (~1 hour)
  2. Run `acct refresh --profile "Test"`
  3. New tokens retrieved
  4. Old refresh token invalidated
  5. Verify: `acct whoami --profile "Test"`

- [ ] **API Connectivity**
  ```bash
  # Test API call with acct CLI
  acct whoami --profile "Test"
  ```
  - Returns company info from QuickBooks
  - Verifies access token works
  - Confirms realmId correct

- [ ] **Qt Tool Integration**
  1. Launch `quickbooks_tool`
  2. Load credentials from SQLite
  3. Select profile
  4. Run connection test
  5. Verify API connectivity
  6. Check company info displayed

## Monitoring & Logs

- [ ] **Broker Logs**
  ```bash
  # Check httpd error logs
  tail -f /var/www/logs/error_log
  # Check access logs
  tail -f /var/www/logs/access_log
  ```
  - Monitor for OAuth errors
  - Watch for API failures
  - Check for certificate issues

- [ ] **Session Cleanup**
  - Verify old sessions are purged
  - Database doesn't grow indefinitely
  - Consider cron job for cleanup

## Documentation

- [ ] **Deployment Documentation**
  - `docs/auth-broker-architecture.md` - comprehensive broker design
  - `docs/BUILDING.md` - build instructions
  - This checklist - deployment verification

- [ ] **User Documentation**
  - Man pages installed: `man quickbooks_tool`, `man acct`
  - Help files available in Qt tools
  - README includes QuickBooks setup

## Backup & Recovery

- [ ] **Configuration Backup**
  - Backup `conf/broker.env` securely
  - Document OAuth credentials location
  - Save Intuit Developer Portal screenshots

- [ ] **Credential Migration**
  - Plan for moving credentials between systems
  - Test credential export/import
  - Verify refresh tokens survive migration

## Version Control

- [ ] **Version Tracking**
  - Current version in `/VERSION` file
  - Broker binary version logged on startup
  - CLI `--version` flag shows correct version

- [ ] **Git Status**
  ```bash
  git status
  git log --oneline -5
  ```
  - No uncommitted broker code
  - Latest changes pushed to remote
  - Tag releases appropriately

---

## Quick Reference Commands

### Server Access
```bash
ssh aops@merah.cassia.ifost.org.au
```

### Check Broker Status
```bash
ls -la /var/www/vhosts/auth.industrial-linguistics.com/cgi-bin/broker
rcctl check httpd
rcctl check slowcgi
tail -f /var/www/logs/error_log
```

### Deploy Broker
```bash
# From local dev machine
./scripts/build_deploy_openbsd.sh
```

### Test OAuth Flow
```bash
# From client machine
acct connect qbo --profile "My Company"
acct whoami --profile "My Company"
```

### Verify TLS (Cloudflare)
```bash
# Verify HTTPS is working (via Cloudflare)
curl -I https://auth.industrial-linguistics.com
# Check certificate details (Cloudflare-issued)
openssl s_client -connect auth.industrial-linguistics.com:443 -servername auth.industrial-linguistics.com
```
