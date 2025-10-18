# Testing Guide

This guide explains how to test each skill and workflow to ensure everything is working correctly before using in production.

## Prerequisites

- Access to sandbox/test environments for all systems:
  - Quickbooks Sandbox Account
  - Xero Demo Company
  - Deputy Test Installation
- Test credentials configured
- Sample data available

## Testing Checklist

### 1. Quickbooks Integration

#### Authentication Testing

```bash
# Test OAuth setup
python .claude/skills/quickbooks/scripts/oauth_setup.py
```

**Expected**: Authorization URL, successful token exchange, Realm ID saved

```bash
# Check token status
python .claude/skills/quickbooks/scripts/check_token.py
```

**Expected**: Token valid with remaining time displayed

#### API Testing

```bash
# Get company info
python .claude/skills/quickbooks/scripts/api_wrapper.py --endpoint companyinfo
```

**Expected**: JSON with company details

```bash
# Query accounts
python .claude/skills/quickbooks/scripts/api_wrapper.py \
  --endpoint query \
  --query "SELECT * FROM Account MAXRESULTS 5"
```

**Expected**: JSON with 5 accounts

#### Bank Transaction Testing

```bash
# Get bank transactions
python .claude/skills/quickbooks/scripts/get_bank_transactions.py \
  --account-id YOUR_ACCOUNT_ID \
  --start-date 2025-10-01 \
  --end-date 2025-10-18 \
  --output test_transactions.json
```

**Expected**: JSON file with transactions

#### Reconciliation Testing

```bash
# Test reconciliation with sample data
python .claude/skills/quickbooks/scripts/reconcile_bank.py \
  --account-id YOUR_ACCOUNT_ID \
  --statement-file .claude/skills/quickbooks/templates/bank_statement_template.csv \
  --output test_reconciliation.json
```

**Expected**: Reconciliation report with match statistics

### 2. Xero Integration

#### Authentication Testing

```bash
# Test OAuth setup
python .claude/skills/xero/scripts/oauth_setup.py
```

**Expected**: Authorization URL, token exchange, tenant connections list

```bash
# Verify token
python .claude/skills/xero/scripts/check_token.py
```

**Expected**: Valid token status

#### API Testing

```bash
# Get organizations
python .claude/skills/xero/scripts/api_wrapper.py --endpoint Organisations
```

**Expected**: JSON with organization details

```bash
# Get invoices
python .claude/skills/xero/scripts/api_wrapper.py \
  --endpoint Invoices \
  --where "Type == \"ACCREC\"" \
  --output test_invoices.json
```

**Expected**: JSON with invoices

#### Payroll Testing

```bash
# Get employees (Australia)
python .claude/skills/xero/scripts/api_wrapper.py \
  --endpoint payrollau/employees \
  --api-type payroll
```

**Expected**: JSON with employee list

```bash
# Get timesheets
python .claude/skills/xero/scripts/get_timesheets.py \
  --region AU \
  --start-date 2025-10-01 \
  --end-date 2025-10-18 \
  --output test_timesheets.json
```

**Expected**: Timesheet report with validation

### 3. Deputy Integration

#### Authentication Testing

```bash
# Test API credentials
python .claude/skills/deputy/scripts/test_auth.py
```

**Expected**: Success message with user details

#### API Testing

```bash
# Get current user
python .claude/skills/deputy/scripts/api_wrapper.py --endpoint me
```

**Expected**: JSON with user information

```bash
# Get employees
python .claude/skills/deputy/scripts/api_wrapper.py \
  --endpoint resource/Employee \
  --output test_employees.json
```

**Expected**: JSON with employee list

#### Timesheet Testing

```bash
# Get timesheets
python .claude/skills/deputy/scripts/get_timesheets.py \
  --start-date 2025-10-01 \
  --end-date 2025-10-18 \
  --output test_deputy_timesheets.json
```

**Expected**: Timesheet summary report

## Integration Testing

### Deputy → Xero Workflow

#### Test Data Preparation

1. Create test timesheets in Deputy sandbox
2. Ensure test employees exist in both systems
3. Create employee mapping file

#### Execute Test Workflow

```bash
# 1. Get Deputy timesheets
python .claude/skills/deputy/scripts/get_timesheets.py \
  --start-date 2025-10-01 \
  --end-date 2025-10-14 \
  --output test_deputy_ts.json

# 2. Validate
python .claude/skills/deputy/scripts/validate_timesheets.py \
  --input test_deputy_ts.json \
  --check-all

# 3. Export to Xero format
python .claude/skills/deputy/scripts/export_to_xero.py \
  --input test_deputy_ts.json \
  --mapping test_employee_mapping.json \
  --output test_xero_import.json

# 4. Import to Xero (dry-run)
python .claude/skills/xero/scripts/import_timesheets.py \
  --input test_xero_import.json \
  --region AU \
  --dry-run
```

**Expected**:
- No errors at any stage
- Validation report shows any expected issues
- Dry-run shows what would be imported
- No actual data modified in Xero

### Bank Reconciliation Workflow

#### Test with Sample Data

```bash
# Quickbooks
python .claude/skills/quickbooks/scripts/reconcile_bank.py \
  --account-id TEST_ACCOUNT \
  --statement-file .claude/skills/quickbooks/templates/bank_statement_template.csv \
  --tolerance 0.01 \
  --output test_qb_reconciliation.json

# Xero
python .claude/skills/xero/scripts/reconcile_bank.py \
  --account-code TEST_CODE \
  --statement-file test_statement.csv \
  --output test_xero_reconciliation.json
```

**Expected**:
- Reconciliation completes
- Match statistics reasonable
- Unmatched transactions identified
- Reports generated

## Error Testing

### Test Invalid Authentication

```bash
# Temporarily corrupt token file
cp config/xero_tokens.json config/xero_tokens.json.backup
echo '{"invalid": "token"}' > config/xero_tokens.json

# Attempt API call
python .claude/skills/xero/scripts/api_wrapper.py --endpoint Organisations

# Should get error about invalid token

# Restore
mv config/xero_tokens.json.backup config/xero_tokens.json
```

**Expected**: Clear error message about authentication failure

### Test Token Refresh

```bash
# Manually expire token (edit tokens.json, set refreshed_at to yesterday)
# Then make API call
python .claude/skills/xero/scripts/api_wrapper.py --endpoint Organisations
```

**Expected**: Token automatically refreshed, call succeeds

### Test Invalid Queries

```bash
# Test with invalid query
python .claude/skills/quickbooks/scripts/api_wrapper.py \
  --endpoint query \
  --query "SELECT * FROM NonExistentEntity"
```

**Expected**: Clear error message from API

### Test Network Issues

```bash
# Temporarily disable network or use invalid domain
# Edit credentials to use invalid domain
python .claude/skills/deputy/scripts/test_auth.py
```

**Expected**: Clear error message about connection failure

## Performance Testing

### Bulk Operations

```bash
# Test with large date range
time python .claude/skills/xero/scripts/api_wrapper.py \
  --endpoint Invoices \
  --where "Date >= DateTime(2020,01,01)"
```

**Monitor**:
- Response time
- Memory usage
- Rate limiting behavior

### Concurrent Requests

```bash
# Run multiple operations in parallel
python .claude/skills/quickbooks/scripts/api_wrapper.py --endpoint companyinfo &
python .claude/skills/xero/scripts/api_wrapper.py --endpoint Organisations &
python .claude/skills/deputy/scripts/test_auth.py &
wait
```

**Expected**: All complete successfully without interference

## Data Validation Testing

### Timesheet Data Integrity

```bash
# Create test timesheet with known values
# Retrieve and verify calculations

python .claude/skills/deputy/scripts/get_timesheets.py \
  --start-date 2025-10-01 \
  --end-date 2025-10-01 \
  --employee-id TEST_EMP_ID \
  --output test_single_timesheet.json

# Manually verify:
# - Hours calculated correctly
# - Breaks subtracted properly
# - Overtime identified correctly
```

### Reconciliation Accuracy

```bash
# Create controlled test scenario:
# - Known bank statement
# - Known QB/Xero transactions
# - Expected matches

python .claude/skills/quickbooks/scripts/reconcile_bank.py \
  --account-id TEST_ACCOUNT \
  --statement-file controlled_test_statement.csv \
  --output test_reconciliation.json

# Verify:
# - All expected matches found
# - No false positives
# - Correct unmatched identification
```

## Regression Testing

After any code changes, run this test suite:

```bash
#!/bin/bash
# run_regression_tests.sh

echo "Running Regression Test Suite"
echo "=============================="

# Test Quickbooks
echo "Testing Quickbooks..."
python .claude/skills/quickbooks/scripts/check_token.py || exit 1
python .claude/skills/quickbooks/scripts/api_wrapper.py --endpoint companyinfo || exit 1

# Test Xero
echo "Testing Xero..."
python .claude/skills/xero/scripts/check_token.py || exit 1
python .claude/skills/xero/scripts/api_wrapper.py --endpoint Organisations || exit 1

# Test Deputy
echo "Testing Deputy..."
python .claude/skills/deputy/scripts/test_auth.py || exit 1

echo "=============================="
echo "All regression tests passed!"
```

## Security Testing

### Credential Leak Check

```bash
# Verify credentials not in git
git log --all -p | grep -i "api_key\|client_secret\|access_token"

# Should return nothing
```

### Permission Testing

```bash
# Verify restricted operations require confirmation
# Attempt to delete/modify data without confirmation

python .claude/skills/xero/scripts/delete_invoice.py --invoice-id TEST_ID

# Should prompt for confirmation
```

## Monitoring and Logging

### Log Verification

```bash
# Check logs are being written
ls -lh logs/

# Verify log content
tail -100 logs/xero_api.log
tail -100 logs/quickbooks_api.log
tail -100 logs/deputy_api.log

# Check for errors
grep ERROR logs/*.log
```

### Audit Trail

```bash
# Verify all operations are logged
python .claude/skills/xero/scripts/api_wrapper.py --endpoint Invoices
grep "API Request" logs/xero_api.log | tail -1

# Should show the request just made
```

## Test Environment Setup

### Quickbooks Sandbox

1. Create sandbox company at https://developer.intuit.com
2. Add test data (accounts, invoices, etc.)
3. Configure OAuth app with http://localhost:8000/callback
4. Save sandbox credentials

### Xero Demo Company

1. Create demo company at https://developer.xero.com
2. Import sample data
3. Configure OAuth app
4. Save demo company credentials

### Deputy Test Install

1. Request test installation from Deputy
2. Create test employees and rosters
3. Generate API token
4. Save test credentials

## Continuous Integration

Add these tests to your CI/CD pipeline:

```yaml
# .github/workflows/test.yml
name: Test Skills

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Set up Python
        uses: actions/setup-python@v2
        with:
          python-version: 3.9
      - name: Run auth tests
        env:
          QB_CLIENT_ID: ${{ secrets.QB_CLIENT_ID }}
          QB_CLIENT_SECRET: ${{ secrets.QB_CLIENT_SECRET }}
        run: |
          python .claude/skills/quickbooks/scripts/check_token.py
      # Add more tests...
```

## Sign-Off Checklist

Before deploying to production:

- [ ] All authentication tests pass
- [ ] All API tests pass
- [ ] Integration workflows tested end-to-end
- [ ] Error handling tested
- [ ] Logs reviewed for unexpected errors
- [ ] Security checks completed
- [ ] Performance acceptable
- [ ] Documentation reviewed and updated
- [ ] Backup and recovery tested
- [ ] User acceptance testing completed

## Troubleshooting Test Failures

### Authentication Failures
1. Check credential files exist
2. Verify tokens not expired
3. Try refreshing tokens
4. Re-run oauth_setup if needed

### API Failures
1. Check network connectivity
2. Verify API endpoints haven't changed
3. Check rate limiting
4. Review API error messages

### Data Mismatch
1. Verify test data exists
2. Check date ranges
3. Verify account IDs/codes
4. Review query syntax

### Integration Failures
1. Check employee mappings
2. Verify data formats
3. Test each step individually
4. Review logs for specific errors

---

Remember: **Never test destructive operations in production!** Always use sandbox/demo environments for testing.
