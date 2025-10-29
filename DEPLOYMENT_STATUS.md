# QuickBooks Deployment Status Report

**Generated:** 2025-10-30
**Server:** aops@merah.cassia.ifost.org.au
**Domain:** auth.industrial-linguistics.com

---

## Summary

The server infrastructure is properly configured, but the OAuth broker application is **NOT DEPLOYED**. The broker binary and configuration files need to be built and deployed before QuickBooks integration can function.

---

## ✅ What's Working

### Infrastructure
- **SSH Access:** ✅ Working
- **Server Platform:** ✅ OpenBSD 7.7 amd64
- **Web Server:** ✅ httpd is running and configured
- **FastCGI:** ✅ slowcgi is running and configured
- **Cloudflare:** ✅ TLS termination working, traffic routing to origin
- **Directory Structure:** ✅ /var/www/vhosts/auth.industrial-linguistics.com exists with proper subdirectories
  - cgi-bin/ (owned by aops)
  - conf/ (owned by root)
  - data/ (owned by aops, writable by www)
  - htdocs/
  - logs/

### Chroot Environment
- **DNS Configuration:** ✅ /var/www/etc/resolv.conf exists (updated Jun 4, 2024)
- **CA Certificates:** ✅ /var/www/etc/ssl/cert.pem exists (339KB, updated Jun 4, 2024)

### httpd Configuration
- **Virtual Host:** ✅ auth.industrial-linguistics.com configured
- **FastCGI Support:** ✅ /cgi-bin/* path configured with fastcgi
- **Root Path:** ✅ Points to /vhosts/auth.industrial-linguistics.com

---

## ❌ What's Missing

### Critical - Deployment Blockers

1. **OAuth Broker Binary NOT DEPLOYED**
   - Location: `/var/www/vhosts/auth.industrial-linguistics.com/cgi-bin/broker`
   - Status: **MISSING** (directory is empty)
   - Impact: OAuth flows cannot work without the broker

2. **Broker Configuration NOT DEPLOYED**
   - Location: `/var/www/vhosts/auth.industrial-linguistics.com/conf/broker.env`
   - Status: **MISSING** (directory is empty)
   - Required contents:
     - QuickBooks OAuth client ID
     - QuickBooks OAuth client secret
     - Deputy OAuth credentials
     - Xero OAuth credentials
     - Database path for session storage

3. **Deployment Script NOT CREATED**
   - Expected location: `scripts/build_deploy_openbsd.sh`
   - Status: **MISSING** (scripts directory doesn't exist)
   - Referenced in: CLAUDE.md, deployment docs

### Testing Status

- **Endpoint Tests:** All endpoints return 404 (expected without broker binary)
  - POST /v1/auth/start → 404
  - GET /v1/auth/poll/{session} → 404
  - POST /v1/token/refresh → 404

- **Cloudflare Status:** Working correctly
  - HTTPS: Returns OpenBSD httpd 404 (correct behavior - shows origin is reachable)
  - HTTP: No automatic redirect to HTTPS configured

---

## 📋 Action Items - What You Need To Do

### 1. Build the OAuth Broker Binary

```bash
# From the project root
cd cmd/broker
CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o broker
```

**Expected output:** `cmd/broker/broker` (statically-linked OpenBSD binary)

### 2. Create Broker Configuration File

Create `broker.env` with the following format (example):

```bash
# QuickBooks Online
QBO_CLIENT_ID=your_intuit_client_id
QBO_CLIENT_SECRET=your_intuit_client_secret
QBO_REDIRECT_URI=https://auth.industrial-linguistics.com/v1/callback/qbo
QBO_ENVIRONMENT=production  # or "sandbox"

# Deputy
DEPUTY_CLIENT_ID=your_deputy_client_id
DEPUTY_CLIENT_SECRET=your_deputy_client_secret
DEPUTY_REDIRECT_URI=https://auth.industrial-linguistics.com/v1/callback/deputy

# Xero
XERO_CLIENT_ID=your_xero_client_id
XERO_REDIRECT_URI=https://auth.industrial-linguistics.com/v1/callback/xero

# Session database
DB_PATH=/vhosts/auth.industrial-linguistics.com/data/sessions.db
```

**Note:** You need to obtain these credentials from:
- Intuit Developer Portal (QuickBooks)
- Deputy Developer Portal
- Xero Developer Portal

### 3. Deploy to Server

```bash
# Copy broker binary
scp cmd/broker/broker aops@merah.cassia.ifost.org.au:/var/www/vhosts/auth.industrial-linguistics.com/cgi-bin/

# Set permissions (execute on server)
ssh aops@merah.cassia.ifost.org.au "chmod 755 /var/www/vhosts/auth.industrial-linguistics.com/cgi-bin/broker"

# Copy configuration (as root or with doas)
scp broker.env aops@merah.cassia.ifost.org.au:/tmp/
ssh aops@merah.cassia.ifost.org.au "doas mv /tmp/broker.env /var/www/vhosts/auth.industrial-linguistics.com/conf/ && doas chmod 600 /var/www/vhosts/auth.industrial-linguistics.com/conf/broker.env"
```

### 4. Verify OAuth Redirect URIs Are Registered

For each OAuth provider, ensure the redirect URI is registered:

**QuickBooks (Intuit Developer Portal):**
- URI: `https://auth.industrial-linguistics.com/v1/callback/qbo`
- Must be HTTPS (no localhost/IP in production)

**Deputy:**
- URI: `https://auth.industrial-linguistics.com/v1/callback/deputy`

**Xero:**
- URI: `https://auth.industrial-linguistics.com/v1/callback/xero`

### 5. Create Deployment Script (Optional but Recommended)

Create `scripts/build_deploy_openbsd.sh`:

```bash
#!/bin/sh
set -e

echo "Building OAuth broker for OpenBSD..."
cd cmd/broker
CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o broker

echo "Deploying to server..."
scp broker aops@merah.cassia.ifost.org.au:/var/www/vhosts/auth.industrial-linguistics.com/cgi-bin/

echo "Setting permissions..."
ssh aops@merah.cassia.ifost.org.au "chmod 755 /var/www/vhosts/auth.industrial-linguistics.com/cgi-bin/broker"

echo "Deployment complete!"
```

Make it executable: `chmod +x scripts/build_deploy_openbsd.sh`

### 6. Test Deployment

After deploying, test the broker endpoints:

```bash
# Test OAuth start endpoint
curl -X POST https://auth.industrial-linguistics.com/v1/auth/start \
  -H "Content-Type: application/json" \
  -d '{"provider":"qbo","profile":"test"}'

# Expected: JSON response with session ID, auth URL, and poll URL
# Not expected: 404 error
```

### 7. Configure Cloudflare (Optional - HTTPS Redirect)

If you want HTTP→HTTPS redirect:
- Go to Cloudflare dashboard → SSL/TLS → Edge Certificates
- Enable "Always Use HTTPS"

Or configure in httpd.conf on the server (redirect rule).

### 8. Test End-to-End OAuth Flow

```bash
# From your local machine
cd cmd/acct
go build -o acct

./acct connect qbo --profile "Test Company"
```

Expected behavior:
1. Browser opens to Intuit authorization page
2. After authorization, redirects to broker
3. CLI polls broker and receives tokens
4. Tokens stored in OS keychain

---

## 🔍 Current Broker Code Status

Based on the documentation, the broker should support:
- QuickBooks Online OAuth
- Deputy OAuth
- Xero OAuth

**Need to verify:**
- Does `cmd/broker/main.go` exist?
- Does it implement the required endpoints?
- Does it have proper configuration loading from `broker.env`?

**Next step:** Check if the broker code is implemented or needs to be written.

---

## 📊 Deployment Checklist Progress

| Category | Item | Status |
|----------|------|--------|
| Infrastructure | SSH access | ✅ Done |
| Infrastructure | Server platform (OpenBSD) | ✅ Done |
| Infrastructure | httpd running | ✅ Done |
| Infrastructure | slowcgi running | ✅ Done |
| Infrastructure | Directory structure | ✅ Done |
| Infrastructure | Chroot DNS | ✅ Done |
| Infrastructure | Chroot CA certs | ✅ Done |
| Deployment | Broker binary | ❌ **NOT DEPLOYED** |
| Deployment | Broker config (broker.env) | ❌ **NOT DEPLOYED** |
| Deployment | Deployment script | ❌ **NOT CREATED** |
| Configuration | QuickBooks OAuth credentials | ⚠️ Unknown |
| Configuration | Deputy OAuth credentials | ⚠️ Unknown |
| Configuration | Xero OAuth credentials | ⚠️ Unknown |
| Configuration | Redirect URIs registered | ⚠️ Unknown |
| Testing | Broker endpoints responding | ❌ Not working (404) |
| Testing | End-to-end OAuth flow | ❌ Cannot test yet |

---

## 🚦 Priority Order

1. **HIGH:** Check if `cmd/broker/main.go` exists and is implemented
2. **HIGH:** Obtain OAuth credentials from provider portals
3. **HIGH:** Build broker binary for OpenBSD
4. **HIGH:** Create broker.env configuration
5. **HIGH:** Deploy broker binary and config to server
6. **MEDIUM:** Register OAuth redirect URIs with providers
7. **MEDIUM:** Test broker endpoints
8. **MEDIUM:** Test end-to-end OAuth flow with CLI
9. **LOW:** Create deployment script for future updates
10. **LOW:** Configure HTTPS redirect in Cloudflare

---

## 📝 Notes

- The server infrastructure is solid - no changes needed there
- Cloudflare TLS termination is working correctly
- The main blocker is deploying the broker application code
- Once deployed, you'll need to test with real OAuth credentials
- Consider using sandbox/development credentials first for testing
