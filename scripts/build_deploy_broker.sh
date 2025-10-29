#!/bin/sh
# Build and deploy OAuth broker to OpenBSD/httpd
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Deployment paths
BROKER_BIN_PATH="/var/www/vhosts/auth.industrial-linguistics.com/cgi-bin/broker"
BROKER_DATA_DIR="/var/www/vhosts/auth.industrial-linguistics.com/data"

echo "${GREEN}==> Building OAuth broker for OpenBSD...${NC}"

# Navigate to broker directory
cd "$(dirname "$0")/../cmd/broker" || exit 1

# Build for OpenBSD amd64 with static linking
echo "Building static binary for OpenBSD/amd64..."
CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o broker

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
BROKER_CONF="/var/www/vhosts/auth.industrial-linguistics.com/conf/broker.env"
if [ -f "${BROKER_CONF}" ]; then
    echo "${GREEN}✓ Configuration file exists${NC}"
else
    echo "${YELLOW}⚠ Configuration file missing: ${BROKER_CONF}${NC}"
    echo "  Create this file with OAuth credentials before testing"
fi

echo ""
echo "${GREEN}==> Deployment complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Ensure ${BROKER_CONF} exists with OAuth credentials"
echo "  2. Test broker: curl https://auth.industrial-linguistics.com/v1/auth/start"
echo "  3. Check logs: tail -f /var/www/logs/error_log"
