#!/bin/sh
# Build and deploy OAuth broker to OpenBSD/httpd
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Determine environment based on parameter or git branch
ENVIRONMENT="${1:-}"
if [ -z "$ENVIRONMENT" ]; then
    # Auto-detect from git branch
    BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")
    if [ "$BRANCH" = "main" ]; then
        ENVIRONMENT="production"
    else
        ENVIRONMENT="development"
    fi
fi

# Set deployment paths based on environment
if [ "$ENVIRONMENT" = "production" ]; then
    DOMAIN="auth.industrial-linguistics.com"
    echo "${GREEN}==> Deploying to PRODUCTION (${DOMAIN})${NC}"
else
    DOMAIN="auth-dev.industrial-linguistics.com"
    echo "${GREEN}==> Deploying to DEVELOPMENT (${DOMAIN})${NC}"
fi

BROKER_BIN_PATH="/var/www/vhosts/${DOMAIN}/v1"
BROKER_DATA_DIR="/var/www/vhosts/${DOMAIN}/data"
BROKER_CONF="/var/www/vhosts/${DOMAIN}/conf/broker.env"

echo "${GREEN}==> Building OAuth broker for OpenBSD...${NC}"

# Navigate to broker directory
cd "$(dirname "$0")/../cmd/broker" || exit 1

# Build for OpenBSD amd64 with CGO (required for mattn/go-sqlite3)
echo "Building binary for OpenBSD/amd64..."
CGO_ENABLED=1 GOOS=openbsd GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o broker

if [ ! -f broker ]; then
    echo "${RED}ERROR: Build failed - broker binary not created${NC}"
    exit 1
fi

echo "${GREEN}✓ Build successful${NC}"

# Check file details
file broker
ls -lh broker

echo ""
echo "${GREEN}==> Deploying broker binary...${NC}"

# Copy to CGI directory
echo "Copying to ${BROKER_BIN_PATH}..."
cp broker "${BROKER_BIN_PATH}"

# Set permissions
echo "Setting permissions (755)..."
chmod 755 "${BROKER_BIN_PATH}"

# Verify
if [ -x "${BROKER_BIN_PATH}" ]; then
    echo "${GREEN}✓ Broker deployed successfully${NC}"
    ls -lh "${BROKER_BIN_PATH}"
else
    echo "${RED}ERROR: Deployment failed - binary not executable${NC}"
    exit 1
fi

echo ""
echo "${GREEN}==> Verifying environment...${NC}"

# Check data directory exists and is writable
if [ -d "${BROKER_DATA_DIR}" ]; then
    echo "${GREEN}✓ Data directory exists: ${BROKER_DATA_DIR}${NC}"
else
    echo "${YELLOW}⚠ Data directory missing: ${BROKER_DATA_DIR}${NC}"
fi

# Check if config file exists
if [ -f "${BROKER_CONF}" ]; then
    echo "${GREEN}✓ Configuration file exists${NC}"
    # Check permissions
    PERMS=$(stat -f "%Op" "${BROKER_CONF}" 2>/dev/null | tail -c 4)
    OWNER=$(stat -f "%Su:%Sg" "${BROKER_CONF}" 2>/dev/null)
    echo "  Permissions: ${PERMS} Owner: ${OWNER}"
    if [ "$PERMS" != "0640" ] && [ "$PERMS" != "640" ]; then
        echo "${YELLOW}  ⚠ WARNING: Permissions should be 640 (rw-r-----)${NC}"
        echo "  Fix with: doas chmod 640 ${BROKER_CONF}"
    fi
    if [ "$OWNER" != "aops:www" ]; then
        echo "${YELLOW}  ⚠ WARNING: Owner should be aops:www${NC}"
        echo "  Fix with: doas chown aops:www ${BROKER_CONF}"
    fi
else
    echo "${YELLOW}⚠ Configuration file missing: ${BROKER_CONF}${NC}"
    echo "  Create this file with OAuth credentials before testing"
    echo "  Set permissions: doas chmod 640 ${BROKER_CONF}"
    echo "  Set ownership: doas chown aops:www ${BROKER_CONF}"
fi

echo ""
echo "${GREEN}==> Deployment complete!${NC}"
echo ""
echo "Environment: ${ENVIRONMENT}"
echo "Domain: https://${DOMAIN}"
echo ""
echo "Next steps:"
echo "  1. Ensure ${BROKER_CONF} exists with OAuth credentials (640, aops:www)"
echo "  2. Test broker: curl https://${DOMAIN}/v1/healthz"
echo "  3. Check logs: tail -f /var/www/logs/error_log"
