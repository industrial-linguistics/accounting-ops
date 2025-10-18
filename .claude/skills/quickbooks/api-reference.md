# Quickbooks Online API Reference

Comprehensive reference for Quickbooks Online API endpoints and operations.

## Base Information

**Base URLs:**
- Sandbox: `https://sandbox-quickbooks.api.intuit.com`
- Production: `https://quickbooks.api.intuit.com`

**API Version:** v3

**Authentication:** OAuth 2.0 Bearer token in Authorization header

**Content-Type:** application/json

**API Rate Limits:**
- 500 requests per minute per company (Realm ID)
- 100 concurrent requests

## Common Request Pattern

```bash
curl -X GET \
  "https://sandbox-quickbooks.api.intuit.com/v3/company/{realmId}/{endpoint}" \
  -H "Accept: application/json" \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json"
```

## Core Endpoints

### Company Information

**Get Company Info**
```bash
GET /v3/company/{realmId}/companyinfo/{realmId}
```

Example:
```bash
python scripts/api_wrapper.py --endpoint companyinfo
```

Response includes:
- Company name and contact info
- Address
- Tax information
- Email addresses
- Phone numbers

### Query Endpoint

Execute SQL-like queries to retrieve data.

**Query Syntax:**
```sql
SELECT * FROM {EntityType} WHERE {conditions} ORDER BY {field} {ASC|DESC}
```

**Supported Operators:**
- `=`, `!=`, `<`, `>`, `<=`, `>=`
- `IN`, `LIKE`

**Example Queries:**

Get all active customers:
```bash
python scripts/api_wrapper.py \
  --endpoint query \
  --query "SELECT * FROM Customer WHERE Active = true"
```

Get invoices in date range:
```bash
python scripts/api_wrapper.py \
  --endpoint query \
  --query "SELECT * FROM Invoice WHERE TxnDate >= '2025-10-01' AND TxnDate <= '2025-10-18'"
```

Get unreconciled purchases:
```bash
python scripts/api_wrapper.py \
  --endpoint query \
  --query "SELECT * FROM Purchase WHERE Reconciled = 'NotReconciled'"
```

### Accounts (Chart of Accounts)

**Query All Accounts:**
```sql
SELECT * FROM Account
```

**Get Active Accounts:**
```sql
SELECT * FROM Account WHERE Active = true
```

**Get Bank Accounts:**
```sql
SELECT * FROM Account WHERE AccountType = 'Bank'
```

**Account Types:**
- `Bank`
- `Accounts Receivable`
- `Other Current Asset`
- `Fixed Asset`
- `Other Asset`
- `Accounts Payable`
- `Credit Card`
- `Other Current Liability`
- `Long Term Liability`
- `Equity`
- `Income`
- `Cost of Goods Sold`
- `Expense`
- `Other Income`
- `Other Expense`

### Customers

**Query Customers:**
```sql
SELECT * FROM Customer
```

**Create Customer:**
```bash
python scripts/api_wrapper.py \
  --endpoint customer \
  --method POST \
  --data customer_data.json
```

Sample `customer_data.json`:
```json
{
  "DisplayName": "ABC Company",
  "PrimaryEmailAddr": {
    "Address": "contact@abccompany.com"
  },
  "PrimaryPhone": {
    "FreeFormNumber": "555-1234"
  },
  "CompanyName": "ABC Company Pty Ltd"
}
```

### Vendors

**Query Vendors:**
```sql
SELECT * FROM Vendor
```

**Create Vendor:**
```json
{
  "DisplayName": "Office Supplies Inc",
  "PrimaryEmailAddr": {
    "Address": "billing@officesupplies.com"
  }
}
```

### Invoices

**Query Invoices:**
```sql
SELECT * FROM Invoice WHERE TxnDate >= '2025-10-01'
```

**Get Unpaid Invoices:**
```sql
SELECT * FROM Invoice WHERE Balance > '0'
```

**Invoice Structure:**
```json
{
  "Line": [
    {
      "Amount": 100.0,
      "DetailType": "SalesItemLineDetail",
      "SalesItemLineDetail": {
        "ItemRef": {
          "value": "1",
          "name": "Services"
        }
      }
    }
  ],
  "CustomerRef": {
    "value": "1"
  }
}
```

### Bills

**Query Bills:**
```sql
SELECT * FROM Bill WHERE DueDate <= '2025-10-25'
```

**Get Unpaid Bills:**
```sql
SELECT * FROM Bill WHERE Balance > '0'
```

### Purchases

**Query Purchases:**
```sql
SELECT * FROM Purchase WHERE TxnDate >= '2025-10-01'
```

**Purchase by Account:**
```sql
SELECT * FROM Purchase WHERE AccountRef = '35'
```

### Deposits

**Query Deposits:**
```sql
SELECT * FROM Deposit WHERE TxnDate >= '2025-10-01'
```

**Deposit to Specific Account:**
```sql
SELECT * FROM Deposit WHERE DepositToAccountRef = '35'
```

### Payments

**Query Payments:**
```sql
SELECT * FROM Payment WHERE TxnDate >= '2025-10-01'
```

**Customer Payments:**
```sql
SELECT * FROM Payment WHERE CustomerRef = '1'
```

### Bank Transfers

**Query Transfers:**
```sql
SELECT * FROM Transfer WHERE TxnDate >= '2025-10-01'
```

## Bank Reconciliation

### Get Account Transactions

Use the specialized script:
```bash
python scripts/get_bank_transactions.py \
  --account-id 35 \
  --start-date 2025-10-01 \
  --end-date 2025-10-18 \
  --output transactions.json
```

### Transaction Status

Transactions have a reconciliation status:
- `NotReconciled` - Not yet reconciled
- `Reconciled` - Reconciled
- `Pending` - Pending reconciliation

### Reports

**Available Reports:**
- Balance Sheet
- Profit and Loss
- Trial Balance
- General Ledger
- Cash Flow

**Get Report:**
```bash
GET /v3/company/{realmId}/reports/{ReportName}?start_date={YYYY-MM-DD}&end_date={YYYY-MM-DD}
```

Example:
```bash
python scripts/api_wrapper.py \
  --endpoint "reports/BalanceSheet?date_macro=This%20Month"
```

## Entity Operations

### Create Entity

```bash
POST /v3/company/{realmId}/{entity}
```

Requires JSON body with entity data.

### Read Entity

```bash
GET /v3/company/{realmId}/{entity}/{id}
```

### Update Entity

```bash
POST /v3/company/{realmId}/{entity}
```

Must include `Id` and `SyncToken` in the request body.

### Delete Entity (Soft Delete)

```bash
POST /v3/company/{realmId}/{entity}?operation=delete
```

Requires `Id` and `SyncToken`.

## Error Handling

### Common Error Codes

- `401` - Authentication failure (invalid or expired token)
- `400` - Bad request (invalid data or query)
- `500` - Server error
- `429` - Rate limit exceeded

### Error Response Format

```json
{
  "fault": {
    "error": [
      {
        "message": "Error message here",
        "detail": "Detailed error information",
        "code": "error_code",
        "element": "field_name"
      }
    ],
    "type": "ValidationFault"
  },
  "time": "2025-10-18T12:34:56.789-07:00"
}
```

## Best Practices

1. **Always use pagination** for large queries (max 1000 results per query)
2. **Cache company info** - it rarely changes
3. **Respect rate limits** - implement exponential backoff
4. **Validate SyncToken** before updates to avoid conflicts
5. **Use webhooks** for real-time updates when possible
6. **Test in sandbox** before production
7. **Log all API calls** for audit trail
8. **Handle token expiration** gracefully (refresh automatically)

## Query Pagination

For large result sets, use `STARTPOSITION` and `MAXRESULTS`:

```sql
SELECT * FROM Invoice STARTPOSITION 1 MAXRESULTS 100
SELECT * FROM Invoice STARTPOSITION 101 MAXRESULTS 100
```

## Date Formats

All dates use `YYYY-MM-DD` format:
- `2025-10-18`
- `2025-01-01`

## Amount Formats

Amounts are decimal numbers:
- `100.00`
- `1234.56`

## Useful Query Examples

### Find Overdue Invoices
```sql
SELECT * FROM Invoice WHERE DueDate < '2025-10-18' AND Balance > '0'
```

### Get This Month's Revenue
```sql
SELECT * FROM Invoice WHERE TxnDate >= '2025-10-01' AND TxnDate <= '2025-10-31'
```

### Unreconciled Bank Transactions
```sql
SELECT * FROM Purchase WHERE AccountRef = '35' AND Reconciled = 'NotReconciled'
```

### Top Customers by Revenue
```sql
SELECT DisplayName, Balance FROM Customer WHERE Balance > '0' ORDER BY Balance DESC
```

### Expenses by Vendor
```sql
SELECT * FROM Bill WHERE VendorRef = '12'
```

## Tips for Claude Code

When making Quickbooks API calls:

1. **Always check authentication first** - Use `token_manager.py` to ensure valid token
2. **Use the query endpoint** for most read operations - it's more flexible
3. **Validate data** before POST operations
4. **Check SyncToken** before updates
5. **Handle errors gracefully** - parse fault messages
6. **Provide context** - explain what each query does
7. **Summarize results** - don't just dump JSON

## Further Reading

- [Official Quickbooks API Documentation](https://developer.intuit.com/app/developer/qbo/docs/api/accounting/most-commonly-used/account)
- [OAuth 2.0 Guide](https://developer.intuit.com/app/developer/qbo/docs/develop/authentication-and-authorization)
- [Error Handling](https://developer.intuit.com/app/developer/qbo/docs/develop/error-codes)
