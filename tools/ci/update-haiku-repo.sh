#!/usr/bin/env bash
# Update a Haiku repository with the built .hpkg package(s).
set -euo pipefail

REPO_DIR="${1:-repo}"
PACKAGE_GLOB="${PACKAGE_GLOB:-*.hpkg}"
REPO_NAME="${REPO_NAME:-AccountingOps}"
REPO_VENDOR="${REPO_VENDOR:-\"IFOST Pty Ltd trading as Industrial Linguistics\"}"
REPO_SUMMARY="${REPO_SUMMARY:-Accounting Ops toolkit repository}"
REPO_PRIORITY="${REPO_PRIORITY:-1}"
REPO_BASEURL="${REPO_BASEURL:-http://packages.industrial-linguistics.com/accounting-ops/haiku/repo}"
REPO_IDENTIFIER="${REPO_IDENTIFIER:-tag:industrial-linguistics.com,2025:accounting-ops}"
REPO_ARCH="${REPO_ARCH:-x86_64}"

if ! command -v package_repo >/dev/null 2>&1; then
    if [ -d "$HOME/haiku-hosttools" ]; then
        export PATH="$HOME/haiku-hosttools:$PATH"
        export LD_LIBRARY_PATH="$HOME/haiku-hosttools:${LD_LIBRARY_PATH:-}"
    fi
fi

if ! command -v package_repo >/dev/null 2>&1; then
    echo "package_repo command not found. Run setup-haiku-cross-env.sh first." >&2
    exit 1
fi

mkdir -p "$REPO_DIR/packages"
shopt -s nullglob
packages=( $PACKAGE_GLOB )
shopt -u nullglob

if [ ${#packages[@]} -eq 0 ]; then
    echo "No packages matching '$PACKAGE_GLOB' were found." >&2
    exit 1
fi

for pkg in "${packages[@]}"; do
    cp "$pkg" "$REPO_DIR/packages/"
done

cat > "$REPO_DIR/repo.info" <<EOF_INFO
name $REPO_NAME
vendor $REPO_VENDOR
summary "$REPO_SUMMARY"
priority $REPO_PRIORITY
baseurl $REPO_BASEURL
identifier $REPO_IDENTIFIER
architecture $REPO_ARCH
EOF_INFO

package_repo create "$REPO_DIR/repo.info" "$REPO_DIR"/packages/*.hpkg
