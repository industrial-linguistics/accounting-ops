# Accounting Operations Workflows

This document provides step-by-step workflows for common accounting tasks using the Claude Code skills.

## Table of Contents

1. [Payroll Processing Workflow](#payroll-processing-workflow)
2. [Invoice Processing Workflow](#invoice-processing-workflow)
3. [Bank Reconciliation Workflow](#bank-reconciliation-workflow)
4. [Payment Processing Workflow](#payment-processing-workflow)
5. [End-of-Week Procedures](#end-of-week-procedures)
6. [End-of-Month Procedures](#end-of-month-procedures)

---

## Payroll Processing Workflow

**Goal**: Process payroll from Deputy timesheets through Xero

**Frequency**: Fortnightly (or as per pay schedule)

### Prerequisites
- Deputy timesheets submitted and approved
- Xero payroll calendar configured
- Employee mapping between Deputy and Xero

### Steps

#### 1. Retrieve Deputy Timesheets

```bash
python .claude/skills/deputy/scripts/get_timesheets.py \
  --start-date 2025-10-01 \
  --end-date 2025-10-14 \
  --output deputy_timesheets.json
```

**Expected Output**: JSON file with all timesheets for the period

#### 2. Validate Timesheets

```bash
python .claude/skills/deputy/scripts/validate_timesheets.py \
  --input deputy_timesheets.json \
  --check-overtime \
  --check-breaks \
  --check-late-starts \
  --output validation_report.json
```

**Review**:
- Check validation_report.json for any issues
- Flag overtime hours for review
- Verify meal break compliance
- Identify any late clock-ins or early departures

#### 3. Export to Xero Format

```bash
python .claude/skills/deputy/scripts/export_to_xero.py \
  --input deputy_timesheets.json \
  --mapping employee_mapping.json \
  --output xero_import.json
```

**Note**: employee_mapping.json maps Deputy employee IDs to Xero employee IDs

#### 4. Upload Timesheets to Xero

```bash
python .claude/skills/xero/scripts/import_timesheets.py \
  --input xero_import.json \
  --region AU \
  --dry-run
```

**Review the dry-run results**, then execute:

```bash
python .claude/skills/xero/scripts/import_timesheets.py \
  --input xero_import.json \
  --region AU
```

#### 5. Create Draft Pay Run

```bash
python .claude/skills/xero/scripts/create_payrun.py \
  --calendar-id YOUR_CALENDAR_ID \
  --period-start 2025-10-01 \
  --period-end 2025-10-14 \
  --draft
```

**Expected Output**: Draft pay run ID

#### 6. Review Pay Run

```bash
python .claude/skills/xero/scripts/get_payrun.py \
  --payrun-id PAYRUN_ID \
  --output payrun_details.json
```

**Review**:
- Total gross pay
- Total deductions (tax, super, etc.)
- Net pay amounts
- Any unusual amounts or rates

#### 7. Finalize Pay Run

After approval:

```bash
python .claude/skills/xero/scripts/finalize_payrun.py \
  --payrun-id PAYRUN_ID \
  --confirm
```

#### 8. Generate Pay Slips

```bash
python .claude/skills/xero/scripts/generate_payslips.py \
  --payrun-id PAYRUN_ID \
  --output-dir payslips/2025-10-14/
```

**Deliverables**:
- Pay run completed in Xero
- Pay slips generated
- Payroll summary report
- Audit trail in logs/

---

## Invoice Processing Workflow

**Goal**: Process incoming invoices from email to accounting system

**Frequency**: Daily or as invoices arrive

### Steps

#### 1. Parse Invoice from Email

```bash
python .claude/skills/xero/scripts/parse_invoice_email.py \
  --email-file invoice_email.eml \
  --extract-pdf \
  --output invoice_data.json
```

This extracts:
- Vendor name
- Invoice number
- Date
- Amount
- Line items (if detectable)

#### 2. Validate Invoice Data

```bash
python .claude/skills/xero/scripts/validate_invoice.py \
  --input invoice_data.json
```

**Manual Review Required**:
- Verify vendor details
- Check amounts match PDF
- Assign account codes
- Confirm GST treatment

#### 3. Create Bill in Xero

```bash
python .claude/skills/xero/scripts/create_bill.py \
  --input invoice_data.json \
  --dry-run
```

Review, then:

```bash
python .claude/skills/xero/scripts/create_bill.py \
  --input invoice_data.json
```

#### 4. Attach PDF to Bill

```bash
python .claude/skills/xero/scripts/attach_file.py \
  --bill-id BILL_ID \
  --file invoice.pdf
```

#### 5. Submit for Approval (if required)

```bash
python .claude/skills/xero/scripts/submit_for_approval.py \
  --bill-id BILL_ID
```

**Alternative: Record in Quickbooks**

```bash
# Create bill in Quickbooks
python .claude/skills/quickbooks/scripts/create_bill.py \
  --vendor "Office Supplies Inc" \
  --amount 250.00 \
  --due-date 2025-10-30 \
  --account-id 45 \
  --memo "Office supplies - October"
```

---

## Bank Reconciliation Workflow

**Goal**: Reconcile bank statements with accounting records

**Frequency**: Weekly or monthly

### For Xero

#### 1. Export Bank Statement

Export your bank statement as CSV with these columns:
- Date
- Description
- Debit
- Credit
- Balance

#### 2. Run Reconciliation

```bash
python .claude/skills/xero/scripts/reconcile_bank.py \
  --account-code 090 \
  --statement-file bank_statement.csv \
  --output reconciliation_report.json
```

#### 3. Review Unmatched Transactions

```bash
python .claude/skills/xero/scripts/review_unmatched.py \
  --input reconciliation_report.json \
  --interactive
```

This will prompt you to:
- Create missing transactions
- Match similar transactions
- Categorize new transactions

#### 4. Finalize Reconciliation

```bash
python .claude/skills/xero/scripts/finalize_reconciliation.py \
  --input reconciliation_report.json \
  --statement-balance 12500.00
```

### For Quickbooks

#### 1. Get Quickbooks Transactions

```bash
python .claude/skills/quickbooks/scripts/get_bank_transactions.py \
  --account-id 35 \
  --start-date 2025-10-01 \
  --end-date 2025-10-31 \
  --output qb_transactions.json
```

#### 2. Run Reconciliation

```bash
python .claude/skills/quickbooks/scripts/reconcile_bank.py \
  --account-id 35 \
  --statement-file bank_statement.csv \
  --output reconciliation_report.json
```

#### 3. Review and Finalize

Similar to Xero process above.

---

## Payment Processing Workflow

**Goal**: Identify and schedule payments due this week

**Frequency**: Weekly (typically Monday)

### Steps

#### 1. Get Bills Due This Week

```bash
python .claude/skills/xero/scripts/api_wrapper.py \
  --endpoint invoices \
  --where "Type == \"ACCPAY\" AND Status == \"AUTHORISED\" AND DueDate >= DateTime(2025,10,18) AND DueDate <= DateTime(2025,10,25)" \
  --output bills_due.json
```

#### 2. Generate Payment Summary

```bash
python .claude/skills/xero/scripts/generate_payment_summary.py \
  --input bills_due.json \
  --output payment_summary.xlsx
```

This creates a spreadsheet with:
- Vendor name
- Invoice number
- Amount
- Due date
- Priority (based on due date and penalties)

#### 3. Review and Approve Payments

**Manual Review**:
- Verify all bills are legitimate
- Check payment terms
- Identify any early payment discounts
- Prioritize by due date and importance

#### 4. Create ABA File (Australia)

```bash
python .claude/skills/xero/scripts/create_aba_file.py \
  --input approved_payments.json \
  --bank-account YOUR_ACCOUNT_NUMBER \
  --output payments_2025-10-18.aba
```

#### 5. Upload ABA to Bank

Upload the ABA file to your bank's batch payment system.

#### 6. Record Payments in Xero

After processing:

```bash
python .claude/skills/xero/scripts/record_payments.py \
  --input payments_2025-10-18.json \
  --payment-date 2025-10-19
```

---

## End-of-Week Procedures

**Friday afternoon checklist**

### 1. Review Timesheet Completion

```bash
python .claude/skills/deputy/scripts/check_incomplete.py \
  --start-date $(date -d 'last monday' +%Y-%m-%d) \
  --end-date $(date +%Y-%m-%d)
```

**Action**: Follow up on any incomplete timesheets

### 2. Review Pending Invoices

```bash
python .claude/skills/xero/scripts/api_wrapper.py \
  --endpoint invoices \
  --where "Status == \"DRAFT\"" \
  --output draft_invoices.json
```

**Action**: Finalize or delete draft invoices

### 3. Check Cash Position

```bash
python .claude/skills/xero/scripts/get_cash_position.py \
  --output cash_position.json
```

**Review**:
- Current bank balances
- Payments scheduled for next week
- Expected receipts

### 4. Backup Data

```bash
python .claude/skills/common/backup_data.py \
  --systems deputy,xero,quickbooks \
  --output-dir backups/$(date +%Y-%m-%d)/
```

---

## End-of-Month Procedures

**Month-end close process**

### 1. Bank Reconciliation

Complete bank reconciliation for all accounts (see workflow above)

### 2. Review Accounts Payable Aging

```bash
python .claude/skills/xero/scripts/get_ap_aging.py \
  --as-of $(date +%Y-%m-%d) \
  --output ap_aging.xlsx
```

**Review**:
- Overdue bills
- Upcoming due dates
- Payment scheduling

### 3. Review Accounts Receivable Aging

```bash
python .claude/skills/xero/scripts/get_ar_aging.py \
  --as-of $(date +%Y-%m-%d) \
  --output ar_aging.xlsx
```

**Action**: Follow up on overdue invoices

### 4. Generate Financial Reports

```bash
# Profit & Loss
python .claude/skills/xero/scripts/api_wrapper.py \
  --endpoint "reports/ProfitAndLoss?fromDate=2025-10-01&toDate=2025-10-31" \
  --output pl_october.json

# Balance Sheet
python .claude/skills/xero/scripts/api_wrapper.py \
  --endpoint "reports/BalanceSheet?date=2025-10-31" \
  --output bs_october.json
```

### 5. Accrue for Unpaid Invoices/Bills

**Manual Entry Required**: Create accrual journal entries for:
- Received but not yet invoiced
- Services delivered but not yet billed

### 6. Review and Post Adjustments

**Manual Review**: Review any adjustments needed for:
- Depreciation
- Prepayments
- Accruals
- Provisions

### 7. Close Month

```bash
python .claude/skills/xero/scripts/close_month.py \
  --period 2025-10 \
  --lock-date 2025-10-31
```

### 8. Generate Month-End Report Package

```bash
python .claude/skills/common/generate_month_end_pack.py \
  --period 2025-10 \
  --output month_end_october_2025.pdf
```

**Includes**:
- Profit & Loss
- Balance Sheet
- Cash Flow Statement
- Aged Receivables
- Aged Payables
- Key metrics and KPIs

---

## Integration Workflows

### Deputy → Xero Payroll (Full Pipeline)

```bash
# Single command to run entire payroll
python .claude/skills/common/run_payroll_pipeline.py \
  --start-date 2025-10-01 \
  --end-date 2025-10-14 \
  --calendar-id YOUR_CALENDAR_ID \
  --approve-overtime \
  --dry-run
```

This orchestrates:
1. Deputy timesheet retrieval
2. Validation checks
3. Format conversion
4. Xero upload
5. Pay run creation

### Email → Xero Invoice (Automated)

```bash
# Set up automated processing
python .claude/skills/common/setup_email_monitor.py \
  --mailbox invoices@yourcompany.com \
  --target xero \
  --auto-categorize \
  --approval-threshold 500.00
```

Automatically:
- Monitors email inbox
- Extracts invoice data
- Creates draft bills in Xero
- Routes >$500 for approval
- Auto-approves <$500 if vendor is trusted

---

## Troubleshooting

### Authentication Issues

**Quickbooks**:
```bash
python .claude/skills/quickbooks/scripts/check_token.py
python .claude/skills/quickbooks/scripts/refresh_token.py
```

**Xero**:
```bash
python .claude/skills/xero/scripts/check_token.py
python .claude/skills/xero/scripts/refresh_token.py
```

**Deputy**:
```bash
python .claude/skills/deputy/scripts/test_auth.py
```

### Data Sync Issues

```bash
python .claude/skills/common/verify_sync.py \
  --source deputy \
  --target xero \
  --check-employees
```

### Logs

All operations are logged:
```bash
tail -f logs/quickbooks_api.log
tail -f logs/xero_api.log
tail -f logs/deputy_api.log
```

---

## Best Practices

1. **Always use --dry-run first** for operations that modify data
2. **Review validation reports** before finalizing
3. **Keep backups** before bulk operations
4. **Use sandbox environments** for testing
5. **Document manual adjustments** in commit messages
6. **Set up automated backups** of config and data
7. **Rotate API credentials** regularly
8. **Monitor logs** for errors or anomalies
9. **Test workflows** end-to-end in sandbox before production
10. **Maintain audit trail** of all automated operations

---

## Support and Documentation

- Quickbooks API: [api-reference.md](.claude/skills/quickbooks/api-reference.md)
- Xero Skill: [SKILL.md](.claude/skills/xero/SKILL.md)
- Deputy Skill: [SKILL.md](.claude/skills/deputy/SKILL.md)
- Repository: [README.md](../README.md)

For issues or questions, check the logs and error messages first. Most authentication issues can be resolved by refreshing tokens.
