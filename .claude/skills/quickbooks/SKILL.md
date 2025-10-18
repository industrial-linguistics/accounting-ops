---
name: Quickbooks Online Integration
description: Access Quickbooks Online API to manage accounting data, bank reconciliation, invoices, bills, and financial reports. Use when working with Quickbooks, accounting operations, bank reconciliation, invoices, or financial data management.
allowed-tools: Read, Write, Bash, Grep, Glob
---

# Quickbooks Online Integration

This skill provides comprehensive access to Quickbooks Online API for accounting operations, with a focus on bank reconciliation, invoice management, and financial reporting.

## Prerequisites

1. Quickbooks Online account (sandbox or production)
2. OAuth 2.0 credentials (Client ID and Client Secret)
3. Access tokens (obtained through OAuth flow)

## Authentication

Quickbooks uses OAuth 2.0 for authentication. Access tokens expire after 1 hour and must be refreshed.

### Initial Setup

1. Create credentials file from template:
```bash
cp config/quickbooks_credentials.json.template config/quickbooks_credentials.json
```

2. Add your Client ID and Client Secret to `config/quickbooks_credentials.json`

3. Run the OAuth authorization flow:
```bash
python .claude/skills/quickbooks/scripts/oauth_setup.py
```

This will:
- Generate the authorization URL
- Open it in your browser
- Exchange the authorization code for tokens
- Save tokens to `config/quickbooks_tokens.json`

### Token Management

Access tokens expire after 1 hour. The scripts automatically handle token refresh when needed.

To manually refresh tokens:
```bash
python .claude/skills/quickbooks/scripts/refresh_token.py
```

## Core Operations

### 1. Company Information

Get basic company information:
```bash
bash .claude/skills/quickbooks/scripts/get_company_info.sh
```

Or use the API wrapper:
```bash
python .claude/skills/quickbooks/scripts/api_wrapper.py --endpoint companyinfo --realm-id YOUR_REALM_ID
```

### 2. Chart of Accounts

Query accounts:
```bash
python .claude/skills/quickbooks/scripts/api_wrapper.py \
  --endpoint query \
  --query "SELECT * FROM Account" \
  --realm-id YOUR_REALM_ID
```

### 3. Invoices

Get all invoices:
```bash
python .claude/skills/quickbooks/scripts/api_wrapper.py \
  --endpoint query \
  --query "SELECT * FROM Invoice WHERE TxnDate >= '2025-01-01'" \
  --realm-id YOUR_REALM_ID
```

Create an invoice:
```bash
python .claude/skills/quickbooks/scripts/create_invoice.py \
  --customer-id 1 \
  --amount 100.00 \
  --description "Services rendered"
```

### 4. Bank Transactions

Get bank transactions for reconciliation:
```bash
python .claude/skills/quickbooks/scripts/get_bank_transactions.py \
  --account-id YOUR_ACCOUNT_ID \
  --start-date 2025-10-01 \
  --end-date 2025-10-18
```

### 5. Bank Reconciliation

Run bank reconciliation workflow:
```bash
python .claude/skills/quickbooks/scripts/reconcile_bank.py \
  --account-id YOUR_ACCOUNT_ID \
  --statement-file bank_statement.csv
```

This script will:
1. Fetch transactions from Quickbooks
2. Load transactions from bank statement
3. Match transactions automatically
4. Generate reconciliation report
5. Flag unmatched items for manual review

## Common Queries

### Find Unreconciled Transactions
```bash
python .claude/skills/quickbooks/scripts/api_wrapper.py \
  --endpoint query \
  --query "SELECT * FROM Purchase WHERE TxnDate >= '2025-10-01' AND Reconciled = 'NotReconciled'"
```

### Get Vendor Bills
```bash
python .claude/skills/quickbooks/scripts/api_wrapper.py \
  --endpoint query \
  --query "SELECT * FROM Bill WHERE DueDate <= '2025-10-25'"
```

### List Customers
```bash
python .claude/skills/quickbooks/scripts/api_wrapper.py \
  --endpoint query \
  --query "SELECT * FROM Customer"
```

## API Reference

For detailed API documentation, see [api-reference.md](api-reference.md).

### Base API Structure

All API calls follow this pattern:
```bash
curl -X GET \
  "https://sandbox-quickbooks.api.intuit.com/v3/company/{realmId}/{endpoint}" \
  -H "Accept: application/json" \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json"
```

### Common Endpoints

- `companyinfo/{realmId}` - Company information
- `query?query={SQL_QUERY}` - Query data using SQL-like syntax
- `invoice` - Invoice operations
- `bill` - Bill operations
- `account` - Chart of accounts
- `customer` - Customer management
- `vendor` - Vendor management
- `payment` - Payment tracking

## Error Handling

The scripts include comprehensive error handling:

1. **Token Expiration**: Automatically refreshes tokens
2. **Rate Limiting**: Respects API rate limits (500 requests per minute)
3. **Invalid Queries**: Validates queries before sending
4. **Network Errors**: Retries with exponential backoff

## Security Best Practices

1. Never commit credentials to git
2. Use environment variables in production
3. Enable audit logging for all operations
4. Review all transactions before finalizing
5. Use sandbox environment for testing

## Workflow Examples

### Complete Bank Reconciliation Workflow

1. Fetch transactions from Quickbooks:
```bash
python .claude/skills/quickbooks/scripts/get_bank_transactions.py \
  --account-id 123 \
  --start-date 2025-10-01 \
  --end-date 2025-10-18 \
  --output qb_transactions.json
```

2. Process bank statement:
```bash
python .claude/skills/quickbooks/scripts/parse_bank_statement.py \
  --input bank_statement.csv \
  --output bank_transactions.json
```

3. Match and reconcile:
```bash
python .claude/skills/quickbooks/scripts/match_transactions.py \
  --qb-file qb_transactions.json \
  --bank-file bank_transactions.json \
  --output reconciliation_report.json
```

4. Review unmatched:
```bash
python .claude/skills/quickbooks/scripts/review_unmatched.py \
  --input reconciliation_report.json
```

5. Create reconciliation in Quickbooks:
```bash
python .claude/skills/quickbooks/scripts/finalize_reconciliation.py \
  --input reconciliation_report.json \
  --account-id 123
```

## Tips for Claude Code Users

When I receive a request related to Quickbooks:

1. **Always verify authentication first**: Check if tokens exist and are valid
2. **Use sandbox for testing**: Never test with production data unless explicitly requested
3. **Validate inputs**: Always validate amounts, dates, and IDs before API calls
4. **Provide summaries**: After operations, provide clear summaries of what was done
5. **Flag anomalies**: Alert user to any unusual transactions or data
6. **Request confirmation**: For destructive operations, always confirm with user first

## Troubleshooting

### Token Issues
If you get authentication errors:
```bash
# Check token expiration
python .claude/skills/quickbooks/scripts/check_token.py

# Refresh token
python .claude/skills/quickbooks/scripts/refresh_token.py
```

### API Errors
Check the error log:
```bash
tail -f logs/quickbooks_api.log
```

### Rate Limiting
If you hit rate limits, the scripts will automatically retry with exponential backoff.

## Next Steps

After setting up Quickbooks integration:
1. Test with sandbox account
2. Run sample queries to verify access
3. Set up automated token refresh
4. Configure audit logging
5. Create custom workflows for your use case

For integration with other systems (Xero, Deputy), see the respective skill documentation.
