#!/bin/sh
# Deploy OAuth broker to development environment (auth-dev.industrial-linguistics.com)
# This script is specifically for deploying to the development/sandbox broker instance.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Development environment configuration
DOMAIN="auth-dev.industrial-linguistics.com"
BROKER_BIN_PATH="/var/www/vhosts/${DOMAIN}/cgi-bin/broker"
BROKER_DATA_DIR="/var/www/vhosts/${DOMAIN}/data"
BROKER_CONF="/var/www/vhosts/${DOMAIN}/conf/broker.env"

echo "${BLUE}========================================${NC}"
echo "${BLUE}  Development Broker Deployment${NC}"
echo "${BLUE}========================================${NC}"
echo ""
echo "${GREEN}==> Target: ${DOMAIN}${NC}"
echo ""

# Navigate to broker directory
SCRIPT_DIR="$(dirname "$0")"
cd "${SCRIPT_DIR}/../cmd/broker" || exit 1

echo "${GREEN}==> Building OAuth broker for OpenBSD (development)...${NC}"
echo "Building static binary for OpenBSD/amd64..."

# Build for OpenBSD amd64 with CGO (required for mattn/go-sqlite3)
CGO_ENABLED=1 GOOS=openbsd GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o broker

if [ ! -f broker ]; then
    echo "${RED}ERROR: Build failed - broker binary not created${NC}"
    exit 1
fi

echo "${GREEN}✓ Build successful${NC}"

# Check file details
file broker 2>/dev/null || true
ls -lh broker

echo ""
echo "${GREEN}==> Deploying development broker binary...${NC}"

# Copy to CGI directory
echo "Copying to ${BROKER_BIN_PATH}..."
cp broker "${BROKER_BIN_PATH}"

# Set permissions
echo "Setting permissions (755)..."
chmod 755 "${BROKER_BIN_PATH}"

# Verify
if [ -x "${BROKER_BIN_PATH}" ]; then
    echo "${GREEN}✓ Development broker deployed successfully${NC}"
    ls -lh "${BROKER_BIN_PATH}"
else
    echo "${RED}ERROR: Deployment failed - binary not executable${NC}"
    exit 1
fi

echo ""
echo "${GREEN}==> Verifying development environment...${NC}"

# Check data directory exists
if [ -d "${BROKER_DATA_DIR}" ]; then
    echo "${GREEN}✓ Data directory exists: ${BROKER_DATA_DIR}${NC}"
else
    echo "${YELLOW}⚠ Data directory missing: ${BROKER_DATA_DIR}${NC}"
    echo "  Creating data directory..."
    mkdir -p "${BROKER_DATA_DIR}" 2>/dev/null || echo "${YELLOW}  (Need root to create)${NC}"
fi

# Check if config file exists
if [ -f "${BROKER_CONF}" ]; then
    echo "${GREEN}✓ Configuration file exists${NC}"

    # Check permissions (OpenBSD stat syntax)
    if command -v stat >/dev/null 2>&1; then
        PERMS=$(stat -f "%Op" "${BROKER_CONF}" 2>/dev/null | tail -c 4)
        OWNER=$(stat -f "%Su:%Sg" "${BROKER_CONF}" 2>/dev/null)
        echo "  Permissions: ${PERMS} Owner: ${OWNER}"

        if [ "$PERMS" != "0640" ] && [ "$PERMS" != "640" ]; then
            echo "${YELLOW}  ⚠ WARNING: Permissions should be 640 (rw-r-----)${NC}"
            echo "  Current: ${PERMS}"
            echo "  Fix with: doas chmod 640 ${BROKER_CONF}"
        fi

        if [ "$OWNER" != "root:www" ]; then
            echo "${YELLOW}  ⚠ WARNING: Owner should be root:www${NC}"
            echo "  Current: ${OWNER}"
            echo "  Fix with: doas chown root:www ${BROKER_CONF}"
        fi
    fi

    # Check if www user can read it
    if doas -u www cat "${BROKER_CONF}" >/dev/null 2>&1; then
        echo "${GREEN}  ✓ www user can read configuration${NC}"
    else
        echo "${RED}  ✗ www user CANNOT read configuration${NC}"
        echo "  This will cause the broker to fail!"
        echo "  Fix with: doas chmod 640 ${BROKER_CONF} && doas chown root:www ${BROKER_CONF}"
    fi
else
    echo "${YELLOW}⚠ Configuration file missing: ${BROKER_CONF}${NC}"
    echo "  Create this file with OAuth sandbox/development credentials"
    echo "  See: docs/BROKER_ENV_TEMPLATE.md"
    echo "  Set permissions: doas chmod 640 ${BROKER_CONF}"
    echo "  Set ownership: doas chown root:www ${BROKER_CONF}"
fi

echo ""
echo "${GREEN}==> Development broker deployment complete!${NC}"
echo ""
echo "${BLUE}Environment Details:${NC}"
echo "  Environment: ${YELLOW}development${NC}"
echo "  Domain: https://${DOMAIN}"
echo "  Binary: ${BROKER_BIN_PATH}"
echo "  Config: ${BROKER_CONF}"
echo ""
echo "${BLUE}Next Steps:${NC}"
echo "  1. Ensure ${BROKER_CONF} has correct permissions:"
echo "     ${YELLOW}doas chmod 640 ${BROKER_CONF}${NC}"
echo "     ${YELLOW}doas chown root:www ${BROKER_CONF}${NC}"
echo ""
echo "  2. Test health endpoint:"
echo "     ${YELLOW}curl https://${DOMAIN}/cgi-bin/broker/healthz${NC}"
echo ""
echo "  3. Test OAuth flow with CLI:"
echo "     ${YELLOW}export ACCOUNTING_OPS_BROKER=https://${DOMAIN}/cgi-bin/broker${NC}"
echo "     ${YELLOW}acct connect qbo --profile test-sandbox${NC}"
echo ""
echo "  4. Check server logs if issues:"
echo "     ${YELLOW}doas tail -f /var/www/logs/error.log${NC}"
echo ""
