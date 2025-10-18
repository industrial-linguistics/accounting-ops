# Accounting Operations - Claude Code Skills

This repository contains Claude Code skills for automating accounting operations across Deputy, Xero, and Quickbooks.

## Project Goals

### Phase 1: API Integration Skills
Create Claude Code skills that enable API access to:
1. **Deputy** - Time tracking and workforce management
2. **Xero** - Accounting and payroll platform
3. **Quickbooks** - Accounting and financial management

All integrations will use curl commands for maximum portability and transparency.

### Phase 2: Automated Accounting Operations

#### 1. Payroll Processing
**Goal**: Automate the complete payroll workflow
- Query Deputy for timesheets within a date range
- Identify overtime hours and calculate overtime pay
- Detect meal break violations or issues
- Flag other payment-relevant anomalies (late clock-ins, missing breaks, etc.)
- Enter validated timesheet data into Xero
- Run payroll in Xero with approval checkpoints

#### 2. Invoice and PO Management
**Goal**: Process accounting correspondence automatically
- Monitor and parse emails for:
  - Invoices to be recorded
  - Purchase Orders (POs) to be raised
  - Accounting queries requiring responses
- Extract relevant data from email attachments (PDFs, spreadsheets)
- Create invoice entries in Xero/Quickbooks
- Generate and send PO documents
- Draft responses to common accounting queries

#### 3. Payment Processing and Cash Management
**Goal**: Identify and prepare scheduled payments
- Pull data from Xero to identify:
  - Levies due this week
  - Interest payments scheduled
  - Utility bills to be paid
  - Other recurring or scheduled payments
- Generate payment spreadsheets for review
- Create ABA (Australian Banking Association) files for batch payments
- Categorize and schedule payments by priority and due date

#### 4. Bank Reconciliation
**Goal**: Automate bank reconciliation in both platforms
- **Xero reconciliation**:
  - Import bank statements
  - Match transactions to invoices and bills
  - Flag unmatched or suspicious transactions
  - Suggest categorization for new transactions

- **Quickbooks reconciliation**:
  - Import bank feeds
  - Auto-match transactions
  - Identify discrepancies
  - Generate reconciliation reports

## Technical Architecture

### Skills Structure
```
.claude/skills/
├── deputy/
│   ├── SKILL.md
│   ├── api-reference.md
│   └── scripts/
├── xero/
│   ├── SKILL.md
│   ├── api-reference.md
│   └── scripts/
└── quickbooks/
    ├── SKILL.md
    ├── api-reference.md
    └── scripts/
```

### Integration Approach
- **Authentication**: OAuth 2.0 for all platforms, credentials stored securely
- **API Calls**: curl commands with proper error handling
- **Data Storage**: JSON files for caching and intermediate results
- **Logging**: Comprehensive audit trail for all operations
- **Error Handling**: Graceful degradation with user notifications

## Implementation Priority

### Phase 1A: Quickbooks Integration (PRIORITY)
1. Set up OAuth authentication
2. Create basic CRUD operations (Create, Read, Update, Delete)
3. Implement bank transaction retrieval
4. Build reconciliation workflows

### Phase 1B: Xero Integration
1. Set up OAuth authentication
2. Implement payroll APIs
3. Create invoice and bill management
4. Build payment processing

### Phase 1C: Deputy Integration
1. Set up API authentication
2. Implement timesheet retrieval
3. Create validation rules for overtime/breaks
4. Build export functionality

### Phase 2: Workflow Automation
1. Connect Deputy → Xero for payroll
2. Email parsing and invoice processing
3. Payment file generation (ABA format)
4. Automated reconciliation

## Security Considerations
- API credentials never committed to repository
- Use environment variables or secure credential storage
- Audit logging for all financial operations
- Dry-run mode for testing without actual API calls
- User approval required for financial transactions

## Development Approach
- Test against sandbox environments first
- Build incrementally with frequent testing
- Document all API endpoints and parameters
- Create example workflows for common tasks
- Version control for all configuration changes

## Success Criteria
- [ ] Can authenticate with all three platforms
- [ ] Can retrieve and parse data from each API
- [ ] Payroll process reduces manual work by 80%+
- [ ] Invoice processing is fully automated for standard cases
- [ ] Bank reconciliation accuracy exceeds 95%
- [ ] All operations have complete audit trails
- [ ] Zero financial errors in production
