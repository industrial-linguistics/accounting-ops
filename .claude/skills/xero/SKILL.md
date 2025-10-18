---
name: Xero Integration
description: Access Xero API for accounting, payroll, invoices, and financial operations. Use when working with Xero, payroll processing, timesheets, invoices, or Australian/NZ/UK accounting operations.
allowed-tools: Read, Write, Bash, Grep, Glob
---

# Xero Integration

This skill provides comprehensive access to Xero API for accounting and payroll operations, with specialized support for Australian, New Zealand, and UK payroll.

## Prerequisites

1. Xero account (sandbox or production)
2. OAuth 2.0 credentials (Client ID and Client Secret)
3. Access tokens and Tenant ID (obtained through OAuth flow)

## Authentication

Xero uses OAuth 2.0 for authentication. Access tokens expire after 30 minutes and must be refreshed.

### Initial Setup

1. Create credentials file from template:
```bash
cp config/xero_credentials.json.template config/xero_credentials.json
```

2. Add your Client ID and Client Secret to `config/xero_credentials.json`

3. Run the OAuth authorization flow:
```bash
python .claude/skills/xero/scripts/oauth_setup.py
```

This will:
- Generate the authorization URL
- Open it in your browser
- Exchange the authorization code for tokens
- Retrieve and save your Tenant ID(s)
- Save tokens to `config/xero_tokens.json`

### Token Management

Access tokens expire after 30 minutes. The scripts automatically handle token refresh when needed.

To manually refresh tokens:
```bash
python .claude/skills/xero/scripts/refresh_token.py
```

### Tenant ID

Xero requires a Tenant ID (Organization ID) for most API calls. After authentication, you'll have one or more Tenant IDs. The default will be saved, or you can specify one explicitly.

## Core Operations

### 1. Organization Information

Get organization details:
```bash
python .claude/skills/xero/scripts/api_wrapper.py --endpoint organisations
```

### 2. Invoices

Get all invoices:
```bash
python .claude/skills/xero/scripts/api_wrapper.py \
  --endpoint invoices \
  --where "Date >= DateTime(2025,10,01)"
```

Create an invoice:
```bash
python .claude/skills/xero/scripts/create_invoice.py \
  --contact "ABC Company" \
  --amount 100.00 \
  --description "Services rendered"
```

### 3. Bank Transactions

Get bank transactions:
```bash
python .claude/skills/xero/scripts/api_wrapper.py \
  --endpoint banktransactions \
  --where "Date >= DateTime(2025,10,01) AND BankAccount.Code == \"090\""
```

### 4. Payroll - Timesheets

Get timesheets (Australia):
```bash
python .claude/skills/xero/scripts/get_timesheets.py \
  --region AU \
  --start-date 2025-10-01 \
  --end-date 2025-10-18
```

The script will:
1. Retrieve all timesheets in date range
2. Calculate total hours worked
3. Identify overtime hours
4. Flag any meal break issues
5. Generate a summary report

### 5. Payroll - Pay Runs

Get pay runs:
```bash
python .claude/skills/xero/scripts/api_wrapper.py \
  --endpoint payrollau/payruns \
  --api-type payroll
```

Create pay run:
```bash
python .claude/skills/xero/scripts/create_payrun.py \
  --payroll-calendar-id CALENDAR_ID \
  --period-start 2025-10-01 \
  --period-end 2025-10-14
```

### 6. Bank Reconciliation

Run bank reconciliation:
```bash
python .claude/skills/xero/scripts/reconcile_bank.py \
  --account-code 090 \
  --statement-file bank_statement.csv
```

## Payroll Operations

Xero has region-specific payroll APIs:

### Australia (Payroll AU)
- Base URL: `/payroll.xro/1.0`
- Endpoints: employees, timesheets, payitems, payruns, leaveapplications

### New Zealand (Payroll NZ)
- Base URL: `/payroll.xro/2.0`
- Endpoints: employees, timesheets, employeeleaves, payruns

### UK (Payroll UK)
- Base URL: `/payroll.xro/2.0`
- Endpoints: employees, timesheets, payrunsuk

### Common Payroll Queries

Get employees:
```bash
python .claude/skills/xero/scripts/api_wrapper.py \
  --endpoint payrollau/employees \
  --api-type payroll
```

Get timesheet details:
```bash
python .claude/skills/xero/scripts/get_timesheets.py \
  --region AU \
  --employee-id EMPLOYEE_ID \
  --start-date 2025-10-01 \
  --end-date 2025-10-18 \
  --check-overtime \
  --check-breaks
```

## API Reference

For detailed API documentation, see [api-reference.md](api-reference.md).

### Base API Structure

**Accounting API:**
```bash
curl -X GET \
  "https://api.xero.com/api.xro/2.0/{endpoint}" \
  -H "Accept: application/json" \
  -H "Authorization: Bearer {access_token}" \
  -H "xero-tenant-id: {tenant_id}"
```

**Payroll API (AU):**
```bash
curl -X GET \
  "https://api.xero.com/payroll.xro/1.0/{endpoint}" \
  -H "Accept: application/json" \
  -H "Authorization: Bearer {access_token}" \
  -H "xero-tenant-id: {tenant_id}"
```

### Common Endpoints

**Accounting:**
- `organisations` - Organization information
- `invoices` - Invoice operations
- `contacts` - Contact management
- `banktransactions` - Bank transactions
- `accounts` - Chart of accounts
- `payments` - Payment tracking

**Payroll (AU/NZ/UK):**
- `payroll{region}/employees` - Employee management
- `payroll{region}/timesheets` - Timesheet operations
- `payroll{region}/payruns` - Payroll runs
- `payroll{region}/payitems` - Pay items (earnings, deductions)

## Workflow Examples

### Complete Payroll Workflow (Deputy → Xero)

1. Get timesheets from Deputy:
```bash
python .claude/skills/deputy/scripts/get_timesheets.py \
  --start-date 2025-10-01 \
  --end-date 2025-10-14 \
  --output deputy_timesheets.json
```

2. Validate and transform for Xero:
```bash
python .claude/skills/xero/scripts/import_timesheets.py \
  --input deputy_timesheets.json \
  --check-overtime \
  --check-breaks \
  --output xero_timesheets.json
```

3. Upload to Xero:
```bash
python .claude/skills/xero/scripts/create_timesheets.py \
  --input xero_timesheets.json
```

4. Review and create pay run:
```bash
python .claude/skills/xero/scripts/create_payrun.py \
  --calendar-id CALENDAR_ID \
  --period-start 2025-10-01 \
  --period-end 2025-10-14 \
  --draft
```

5. Review draft pay run, then finalize:
```bash
python .claude/skills/xero/scripts/finalize_payrun.py \
  --payrun-id PAYRUN_ID
```

### Invoice Management Workflow

1. Parse invoice from email:
```bash
python .claude/skills/xero/scripts/parse_invoice_email.py \
  --email-file invoice_email.eml \
  --output invoice_data.json
```

2. Create invoice in Xero:
```bash
python .claude/skills/xero/scripts/create_invoice.py \
  --input invoice_data.json
```

3. Send for approval:
```bash
python .claude/skills/xero/scripts/submit_for_approval.py \
  --invoice-id INVOICE_ID
```

## Error Handling

The scripts include comprehensive error handling:

1. **Token Expiration**: Automatically refreshes tokens (every 30 minutes)
2. **Rate Limiting**: Respects API rate limits (60 requests per minute)
3. **Validation Errors**: Validates data before sending
4. **Network Errors**: Retries with exponential backoff

## Security Best Practices

1. Never commit credentials to git
2. Use environment variables in production
3. Enable audit logging for all operations
4. Review all payroll operations before finalizing
5. Use sandbox environment for testing

## Tips for Claude Code Users

When I receive a request related to Xero:

1. **Always verify authentication first**: Check if tokens exist and are valid
2. **Get Tenant ID**: Ensure correct organization is being used
3. **Use sandbox for testing**: Never test with production data unless explicitly requested
4. **Validate payroll data**: Always check overtime, breaks, and totals
5. **Provide summaries**: After operations, provide clear summaries
6. **Flag anomalies**: Alert user to unusual timesheets or amounts
7. **Request confirmation**: For payroll/financial operations, always confirm first

## Region-Specific Considerations

### Australia
- Award interpretation for overtime
- Superannuation calculations
- PAYG withholding
- Meal break requirements

### New Zealand
- Holiday pay calculations
- KiwiSaver contributions
- PAYE withholding

### UK
- National Insurance contributions
- PAYE tax calculations
- Holiday entitlements

## Troubleshooting

### Token Issues
```bash
# Check token status
python .claude/skills/xero/scripts/check_token.py

# Refresh token
python .claude/skills/xero/scripts/refresh_token.py

# Re-authenticate if refresh token expired
python .claude/skills/xero/scripts/oauth_setup.py
```

### API Errors
Check the error log:
```bash
tail -f logs/xero_api.log
```

### Rate Limiting
Xero limits to 60 requests per minute. Scripts automatically handle this with backoff.

## Next Steps

After setting up Xero integration:
1. Test with sandbox account
2. Run sample queries to verify access
3. Test payroll workflows end-to-end
4. Configure automated token refresh
5. Set up audit logging
6. Create custom workflows for your use case

For integration with Deputy and Quickbooks, see the respective skill documentation.
