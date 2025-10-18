#!/bin/bash
# Get Quickbooks Company Information

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

python3 "$SCRIPT_DIR/api_wrapper.py" --endpoint companyinfo "$@"
