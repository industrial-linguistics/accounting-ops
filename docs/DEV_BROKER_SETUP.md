# Development Broker Setup Guide

This guide covers setting up the development OAuth broker at `auth-dev.industrial-linguistics.com` for testing with sandbox/development credentials.

## Purpose

The development broker allows you to:
- Test OAuth flows with sandbox credentials (QuickBooks sandbox, Xero demo companies)
- Develop and test broker changes before deploying to production
- Isolate development and production credentials for security

## Architecture

We deploy **two separate broker instances**:

| Environment | Domain | Credentials | Use Case |
|-------------|--------|-------------|----------|
| **Production** | `auth.industrial-linguistics.com` | Production OAuth apps | Live customer integrations |
| **Development** | `auth-dev.industrial-linguistics.com` | Sandbox OAuth apps | Testing, development, staging |

Both brokers run the same code but load different `broker.env` configuration files.

## Prerequisites

1. **DNS Configuration**: Ensure `auth-dev.industrial-linguistics.com` points to your server
2. **TLS/SSL**: Cloudflare handles TLS termination (same as production)
3. **Server Access**: SSH access to the OpenBSD server
4. **OAuth Credentials**: Development/sandbox credentials from each provider:
   - QuickBooks: Sandbox keys from developer.intuit.com
   - Xero: Development keys from developer.xero.com (can connect to demo companies)
   - Deputy: Development keys from Deputy API portal

## Step 1: DNS Configuration

Add DNS record for development broker:

```
Type: A or CNAME
Name: auth-dev.industrial-linguistics.com
Target: [your OpenBSD server IP or hostname]
TTL: 300 (or your preference)
```

**Note**: If using Cloudflare, ensure:
- Proxy status: Proxied (orange cloud)
- SSL/TLS mode: Full (not Full Strict, unless you have origin certificates)

Verify DNS propagation:
```bash
dig auth-dev.industrial-linguistics.com
# Should resolve to your server
```

## Step 2: httpd Configuration

On the OpenBSD server, add a new virtual host for the development broker.

Edit `/etc/httpd.conf`:

```nginx
# Development OAuth Broker (auth-dev.industrial-linguistics.com)
server "auth-dev.industrial-linguistics.com" {
    listen on * port 80
    root "/htdocs/auth-dev.industrial-linguistics.com"

    location "/cgi-bin/*" {
        fastcgi socket "/run/slowcgi.sock"
        root "/vhosts/auth-dev.industrial-linguistics.com"
    }

    location * {
        block return 404
    }
}
```

**Directory Structure:**
```
/var/www/
├── vhosts/
│   ├── auth.industrial-linguistics.com/          # Production
│   │   ├── cgi-bin/
│   │   │   └── broker                            # Production broker binary
│   │   └── conf/
│   │       └── broker.env                        # Production credentials
│   └── auth-dev.industrial-linguistics.com/      # Development
│       ├── cgi-bin/
│       │   └── broker                            # Development broker binary
│       └── conf/
│           └── broker.env                        # Development/sandbox credentials
└── htdocs/
    └── auth-dev.industrial-linguistics.com/
        └── index.html                            # Optional placeholder page
```

Create directories:
```bash
doas mkdir -p /var/www/vhosts/auth-dev.industrial-linguistics.com/cgi-bin
doas mkdir -p /var/www/vhosts/auth-dev.industrial-linguistics.com/conf
doas mkdir -p /var/www/htdocs/auth-dev.industrial-linguistics.com
doas chown -R root:daemon /var/www/vhosts/auth-dev.industrial-linguistics.com
```

Reload httpd:
```bash
doas rcctl reload httpd
```

## Step 3: Obtain Development OAuth Credentials

### QuickBooks Online (Sandbox)

1. Go to https://developer.intuit.com/
2. Sign in with your Intuit account
3. Go to "My Apps" → Select your app (or create a new one)
4. Go to "Keys & credentials" tab
5. Under **Development** section (not Production!):
   - Copy **Client ID**
   - Copy **Client Secret**
6. Under **Redirect URIs**, add:
   ```
   https://auth-dev.industrial-linguistics.com/v1/callback/qbo
   ```
7. Save changes

**Environment setting:**
```bash
QBO_ENVIRONMENT=sandbox
```

### Xero (Production with Demo Companies)

Xero doesn't have a separate sandbox mode. You'll use production OAuth endpoints but connect to Xero demo companies for testing.

1. Go to https://developer.xero.com/
2. Sign in with your Xero account
3. Go to "My Apps" → Create new app (or select existing)
4. Note your **Client ID**
5. Under **Redirect URIs**, add:
   ```
   https://auth-dev.industrial-linguistics.com/v1/callback/xero
   ```
6. Generate OAuth 2.0 credentials if needed

**Environment setting:**
```bash
XERO_ENVIRONMENT=production  # Use production endpoints, connect to demo companies
```

**Creating a Demo Company:**
- Log into Xero
- From organization selector, choose "Add Organization"
- Select "Try Xero for free" or create a demo company
- Use this demo company for testing

### Deputy (Development)

1. Contact Deputy support or use Deputy API documentation to obtain development credentials
2. Register redirect URI:
   ```
   https://auth-dev.industrial-linguistics.com/v1/callback/deputy
   ```

**Environment setting:**
```bash
DEPUTY_ENVIRONMENT=production  # Deputy may not have separate sandbox
```

## Step 4: Create Development broker.env

Create `/var/www/vhosts/auth-dev.industrial-linguistics.com/conf/broker.env`:

```bash
# Development OAuth Broker Configuration
# For testing with sandbox/development credentials

# QuickBooks Online - Sandbox
QBO_CLIENT_ID=your_sandbox_client_id_here
QBO_CLIENT_SECRET=your_sandbox_client_secret_here
QBO_REDIRECT=https://auth-dev.industrial-linguistics.com/v1/callback/qbo
QBO_SCOPES=com.intuit.quickbooks.accounting
QBO_ENVIRONMENT=sandbox

# Xero - Production (connects to demo companies)
XERO_CLIENT_ID=your_development_xero_client_id
XERO_REDIRECT=https://auth-dev.industrial-linguistics.com/v1/callback/xero
XERO_SCOPES=offline_access accounting.transactions accounting.contacts
XERO_ENVIRONMENT=production

# Deputy - Development
DEPUTY_CLIENT_ID=your_development_deputy_client_id
DEPUTY_CLIENT_SECRET=your_development_deputy_client_secret
DEPUTY_REDIRECT=https://auth-dev.industrial-linguistics.com/v1/callback/deputy
DEPUTY_SCOPES=longlife_refresh_token
DEPUTY_ENVIRONMENT=production

# Security - Development Key (DO NOT use in production!)
BROKER_MASTER_KEY=dev_key_not_for_production_use_12345

# Optional: Lower rate limits for testing
RATE_LIMIT_AUTH_START=5
RATE_LIMIT_POLL=60
RATE_LIMIT_REFRESH=30
```

Set correct permissions (readable by www user):
```bash
doas chmod 640 /var/www/vhosts/auth-dev.industrial-linguistics.com/conf/broker.env
doas chown root:www /var/www/vhosts/auth-dev.industrial-linguistics.com/conf/broker.env
```

**Note:** The broker runs as user `www` (httpd/slowcgi), so it must be able to read the file. Permissions `640` (rw-r-----) allow root to edit and www to read, while preventing world access.

## Step 5: Deploy Development Broker Binary

### Option A: Manual Deployment

From your development machine:

```bash
cd ~/accounting-ops
cd cmd/broker

# Build for OpenBSD
CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o broker

# Deploy to development broker location
scp broker aops@merah.cassia.ifost.org.au:/tmp/broker-dev
ssh aops@merah.cassia.ifost.org.au "doas mv /tmp/broker-dev /var/www/vhosts/auth-dev.industrial-linguistics.com/cgi-bin/broker"
ssh aops@merah.cassia.ifost.org.au "doas chmod 755 /var/www/vhosts/auth-dev.industrial-linguistics.com/cgi-bin/broker"
```

### Option B: Automated Deployment (Recommended)

Use the provided development deployment script:

```bash
# On the server (via SSH):
ssh aops@merah.cassia.ifost.org.au
cd ~/accounting-ops
git pull origin main

# Run the development deployment script
./scripts/deploy_dev_broker.sh
```

The script (`scripts/deploy_dev_broker.sh`) will:
- Build the broker binary for OpenBSD
- Deploy to `/var/www/vhosts/auth-dev.industrial-linguistics.com/cgi-bin/broker`
- Set correct permissions (755)
- Verify the development environment configuration
- Check broker.env exists with correct permissions (640, root:www)
- Provide detailed next steps if any issues are found

## Step 6: GitHub Actions (Optional)

Add a separate workflow for development deployments in `.github/workflows/deploy-broker-dev.yml`:

```yaml
name: Deploy Development Broker

on:
  push:
    branches:
      - develop  # Auto-deploy on develop branch
    paths:
      - 'cmd/broker/**'
      - 'internal/broker/**'
  workflow_dispatch:  # Manual trigger

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Setup SSH
        run: |
          mkdir -p ~/.ssh
          echo "${{ secrets.DEPLOY_SSH_KEY }}" > ~/.ssh/id_ed25519
          chmod 600 ~/.ssh/id_ed25519
          echo "${{ secrets.SSH_KNOWN_HOSTS }}" > ~/.ssh/known_hosts

      - name: Deploy development broker to server
        run: |
          ssh aops@merah.cassia.ifost.org.au << 'ENDSSH'
            set -e
            cd ~/accounting-ops
            git pull origin develop
            ./scripts/build_deploy_broker_dev.sh
          ENDSSH

      - name: Verify deployment
        run: |
          sleep 2
          curl -f https://auth-dev.industrial-linguistics.com/cgi-bin/broker/healthz || exit 1
          echo "Development broker is healthy!"
```

## Step 7: Verify Development Broker

Test the development broker health endpoint:

```bash
curl https://auth-dev.industrial-linguistics.com/cgi-bin/broker/healthz
# Expected: {"status":"ok","version":"..."}
```

Test OAuth flow with development credentials:

```bash
# Set CLI to use development broker
export ACCOUNTING_OPS_BROKER=https://auth-dev.industrial-linguistics.com/cgi-bin/broker

# Test QuickBooks sandbox connection
acct connect qbo --profile test-sandbox

# Browser should open with QuickBooks sandbox login
# Select a sandbox company
# CLI should receive tokens successfully
```

## Step 8: CLI Development Workflow

### Environment-Based Configuration

Add to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
# Accounting Ops - Development
alias acct-dev='ACCOUNTING_OPS_BROKER=https://auth-dev.industrial-linguistics.com/cgi-bin/broker acct'
```

Usage:
```bash
# Production (default)
acct connect qbo --profile my-real-company

# Development (sandbox)
acct-dev connect qbo --profile test-company
```

### Separate Profiles for Dev/Prod

Create distinct profile names to avoid confusion:

```bash
# Production profiles
acct connect qbo --profile "ACME Corp"
acct connect xero --profile "ACME Corp"

# Development profiles (different names)
acct-dev connect qbo --profile "Test Company (Sandbox)"
acct-dev connect xero --profile "Demo Company (Xero)"
```

Profiles are stored separately in your OS keychain, so dev and prod credentials won't conflict.

## Troubleshooting

### Issue: 404 Not Found

**Symptom:** `curl https://auth-dev.industrial-linguistics.com/cgi-bin/broker/healthz` returns 404

**Checks:**
1. Verify httpd configuration includes the development virtual host
2. Check broker binary exists: `ls -lh /var/www/vhosts/auth-dev.industrial-linguistics.com/cgi-bin/broker`
3. Check slowcgi is running: `rcctl check slowcgi`
4. Check httpd logs: `tail -f /var/www/logs/access.log /var/www/logs/error.log` (no doas needed, logs are readable by aops user)

### Issue: Missing broker.env

**Symptom:** Broker returns errors about missing configuration

**Fix:**
1. Check file exists: `ls -lh /var/www/vhosts/auth-dev.industrial-linguistics.com/conf/broker.env`
2. Check permissions: Should be `rw-------` (600) owned by `root:daemon`
3. Verify broker loads config from correct path (check broker code or logs)

### Issue: OAuth Redirect URI Mismatch

**Symptom:** OAuth provider returns "redirect_uri_mismatch" error

**Fix:**
1. Log into provider's developer portal
2. Verify redirect URI exactly matches: `https://auth-dev.industrial-linguistics.com/v1/callback/{provider}`
3. Ensure no trailing slashes or protocol mismatches (http vs https)
4. For QuickBooks: Check both sandbox and production apps have correct URIs

### Issue: QuickBooks Returns Production Data

**Symptom:** Connected to real company instead of sandbox

**Check:**
1. Verify `broker.env` has `QBO_ENVIRONMENT=sandbox`
2. Verify using sandbox credentials (Client ID from "Development" section in Intuit developer portal)
3. Check broker logs to confirm it's using sandbox API base URL

### Issue: Xero Shows Real Companies

**Expected behavior:** Xero doesn't have a sandbox; you'll see real companies but should select a demo company.

**Solution:** Create a Xero demo company and use that for testing. Xero allows multiple organizations per account.

## Security Notes

- **Never commit broker.env** to version control (already in `.gitignore`)
- **Use different master keys** for dev and production (`BROKER_MASTER_KEY`)
- **Rotate credentials** if development broker is exposed or compromised
- **Monitor access logs** for unexpected traffic
- **Limit rate limits** in development to prevent abuse

## Maintenance

### Updating Development Credentials

1. Edit `/var/www/vhosts/auth-dev.industrial-linguistics.com/conf/broker.env`
2. Update relevant `*_CLIENT_ID` or `*_CLIENT_SECRET` values
3. No restart needed - broker reads config on startup for each request (CGI model)

### Syncing Production and Development Code

When deploying broker changes:

1. Test in development first:
   ```bash
   git checkout develop
   # Make changes, commit
   git push origin develop
   # Triggers auto-deploy to auth-dev.industrial-linguistics.com
   ```

2. Test thoroughly with sandbox credentials

3. Merge to main and deploy to production:
   ```bash
   git checkout main
   git merge develop
   git push origin main
   # Triggers auto-deploy to auth.industrial-linguistics.com
   ```

## Summary

After completing this setup, you'll have:

✅ Two independent broker instances (production and development)
✅ Separate OAuth credentials for each environment
✅ CLI configured to easily switch between brokers
✅ GitHub Actions for automated deployments
✅ Isolated testing environment for broker changes

**Next Steps:**
- Test QuickBooks sandbox OAuth flow
- Create Xero demo company and test connection
- Document any provider-specific quirks or limitations
- Set up monitoring/alerting for both brokers
