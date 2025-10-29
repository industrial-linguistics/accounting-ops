# GitHub Actions Deployment Setup

This document explains how to configure GitHub Actions for automated broker deployment to the OpenBSD server.

## Overview

The deployment workflow automatically:
1. SSHs into `aops@merah.cassia.ifost.org.au`
2. Navigates to `~/accounting-ops`
3. Pulls the latest code from `main` branch
4. Runs `scripts/build_deploy_broker.sh` to build and deploy the broker
5. Verifies the broker is responding

## Workflow Triggers

The workflow runs when:
- Code is pushed to `main` branch AND changes affect:
  - `cmd/broker/**`
  - `internal/broker/**`
  - `scripts/build_deploy_broker.sh`
  - `.github/workflows/deploy-broker.yml`
- Manually triggered via GitHub Actions UI (workflow_dispatch)

## Required GitHub Secrets

You need to configure two secrets in your GitHub repository:

### 1. `DEPLOY_SSH_KEY`

This is the **private SSH key** that can authenticate as `aops@merah.cassia.ifost.org.au`.

**To generate and configure:**

```bash
# On your local machine, generate a new SSH key pair (if you don't have one)
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/github_actions_deploy
# Press Enter for no passphrase (required for automated deployment)

# Copy the public key to the server
ssh-copy-id -i ~/.ssh/github_actions_deploy.pub aops@merah.cassia.ifost.org.au

# Test the key works
ssh -i ~/.ssh/github_actions_deploy aops@merah.cassia.ifost.org.au "echo 'SSH key works'"

# Copy the PRIVATE key contents
cat ~/.ssh/github_actions_deploy
```

**Add to GitHub:**
1. Go to your repository on GitHub
2. Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Name: `DEPLOY_SSH_KEY`
5. Value: Paste the **entire private key** (including `-----BEGIN OPENSSH PRIVATE KEY-----` and `-----END OPENSSH PRIVATE KEY-----`)
6. Click "Add secret"

### 2. `SSH_KNOWN_HOSTS`

This contains the server's host key fingerprint to prevent man-in-the-middle attacks.

**To get the known_hosts entry:**

```bash
# Get the host key fingerprint
ssh-keyscan merah.cassia.ifost.org.au

# Or if you've already connected to the server:
grep merah.cassia.ifost.org.au ~/.ssh/known_hosts
```

**Expected output format:**
```
merah.cassia.ifost.org.au ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleKeyFingerprint...
merah.cassia.ifost.org.au ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDExampleRSAKey...
```

**Add to GitHub:**
1. Go to your repository on GitHub
2. Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Name: `SSH_KNOWN_HOSTS`
5. Value: Paste the output from `ssh-keyscan` (all lines)
6. Click "Add secret"

## Server Setup Requirements

On the OpenBSD server (`aops@merah.cassia.ifost.org.au`), ensure:

### 1. Repository is cloned

```bash
# As user aops
cd ~
git clone https://github.com/industrial-linguistics/accounting-ops.git
cd accounting-ops
```

### 2. Go is installed

```bash
# Check Go version
go version

# If not installed, install Go 1.21+ on OpenBSD:
doas pkg_add go
```

### 3. Repository path is correct

The workflow expects the repository at `~/accounting-ops`. If it's elsewhere, update the workflow file.

### 4. User aops has write access to deployment directories

```bash
# Verify permissions
ls -la /var/www/vhosts/auth.industrial-linguistics.com/cgi-bin/
# Should be owned by aops or writable by aops

ls -la /var/www/vhosts/auth.industrial-linguistics.com/data/
# Should be owned by aops or writable by aops
```

If permissions are incorrect:
```bash
doas chown aops:daemon /var/www/vhosts/auth.industrial-linguistics.com/cgi-bin/
doas chown aops:www /var/www/vhosts/auth.industrial-linguistics.com/data/
```

## Testing the Workflow

### Manual Trigger

1. Go to GitHub repository → Actions tab
2. Click "Deploy OAuth Broker" workflow
3. Click "Run workflow"
4. Select branch: `main`
5. Optionally enter a reason
6. Click "Run workflow"

### Automatic Trigger

Push changes to broker code:

```bash
# Make a change to broker code
vim cmd/broker/main.go

git add cmd/broker/main.go
git commit -m "Update broker configuration"
git push origin main

# Workflow will automatically trigger
```

## Monitoring Deployments

### View workflow status

1. Go to GitHub repository → Actions tab
2. Click on the running workflow
3. View logs for each step

### Check server logs

```bash
# SSH to server
ssh aops@merah.cassia.ifost.org.au

# View httpd error logs
tail -f /var/www/logs/error_log

# View broker binary
ls -lh /var/www/vhosts/auth.industrial-linguistics.com/cgi-bin/broker

# Test broker manually
curl https://auth.industrial-linguistics.com/v1/auth/start \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"provider":"qbo","profile":"test"}'
```

## Troubleshooting

### SSH authentication failed

**Symptom:** Workflow fails with "Permission denied (publickey)"

**Solution:**
1. Verify `DEPLOY_SSH_KEY` secret contains the complete private key
2. Verify the corresponding public key is in `~aops/.ssh/authorized_keys` on server
3. Test SSH connection manually:
   ```bash
   ssh -i ~/.ssh/github_actions_deploy aops@merah.cassia.ifost.org.au
   ```

### Host key verification failed

**Symptom:** Workflow fails with "Host key verification failed"

**Solution:**
1. Update `SSH_KNOWN_HOSTS` secret with current host keys
2. Get fresh host keys:
   ```bash
   ssh-keyscan merah.cassia.ifost.org.au
   ```

### Repository not found

**Symptom:** "cd: ~/accounting-ops: No such file or directory"

**Solution:**
1. SSH to server and clone repository:
   ```bash
   ssh aops@merah.cassia.ifost.org.au
   cd ~
   git clone https://github.com/industrial-linguistics/accounting-ops.git
   ```

### Go build fails

**Symptom:** "go: command not found" or build errors

**Solution:**
1. Install Go on server:
   ```bash
   doas pkg_add go
   ```
2. Verify Go version is 1.21 or later:
   ```bash
   go version
   ```

### Permission denied writing to cgi-bin

**Symptom:** "cp: permission denied" when copying broker binary

**Solution:**
1. Fix directory ownership:
   ```bash
   doas chown aops:daemon /var/www/vhosts/auth.industrial-linguistics.com/cgi-bin/
   ```

### Broker returns 404

**Symptom:** Verification step shows "HTTP 404" after deployment

**Possible causes:**
1. **broker.env missing** - Expected if OAuth credentials not yet configured
2. **httpd not configured** - Check `/etc/httpd.conf` has FastCGI enabled
3. **slowcgi not running** - Check `rcctl check slowcgi`

**Not necessarily an error:** If you haven't created `broker.env` yet, 404 is expected. The workflow allows this and continues.

## Security Considerations

- **Never commit the private SSH key** to the repository
- **Use a dedicated deploy key** rather than personal SSH keys
- **Rotate keys periodically** (update both server and GitHub secret)
- **Limit key access** - Deploy key should only access the deployment server
- **Monitor workflow runs** - Review Actions logs for suspicious activity

## Manual Deployment (Fallback)

If GitHub Actions is unavailable, deploy manually:

```bash
# SSH to server
ssh aops@merah.cassia.ifost.org.au

# Navigate to repository
cd ~/accounting-ops

# Pull latest changes
git pull origin main

# Run build script
./scripts/build_deploy_broker.sh
```

## Workflow File Location

`.github/workflows/deploy-broker.yml`

To modify the workflow, edit this file and commit changes. The workflow will update automatically on next push.
