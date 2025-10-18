---
name: Deputy Integration
description: Access Deputy API for workforce management, timesheets, rosters, and leave management. Use when working with Deputy, employee timesheets, roster scheduling, or time tracking data.
allowed-tools: Read, Write, Bash, Grep, Glob
---

# Deputy Integration

This skill provides comprehensive access to Deputy API for workforce management operations, with a focus on timesheet retrieval and validation for payroll processing.

## Prerequisites

1. Deputy account with API access
2. API access token
3. Deputy installation domain (e.g., yourcompany.deputy.com)

## Authentication

Deputy uses token-based authentication (simpler than OAuth).

### Initial Setup

1. Get your API token from Deputy:
   - Log into Deputy
   - Go to Settings > API Integration
   - Generate or copy your API token

2. Create credentials file from template:
```bash
cp config/deputy_credentials.json.template config/deputy_credentials.json
```

3. Add your API token and domain to `config/deputy_credentials.json`

4. Test authentication:
```bash
python .claude/skills/deputy/scripts/test_auth.py
```

## Core Operations

### 1. Get Current User Info

```bash
python .claude/skills/deputy/scripts/api_wrapper.py --endpoint me
```

### 2. Timesheets

Get timesheets for a date range:
```bash
python .claude/skills/deputy/scripts/get_timesheets.py \
  --start-date 2025-10-01 \
  --end-date 2025-10-18 \
  --output timesheets.json
```

Get timesheets for a specific employee:
```bash
python .claude/skills/deputy/scripts/get_timesheets.py \
  --employee-id 123 \
  --start-date 2025-10-01 \
  --end-date 2025-10-18
```

### 3. Employees

Get all employees:
```bash
python .claude/skills/deputy/scripts/api_wrapper.py \
  --endpoint resource/Employee
```

Get employee by ID:
```bash
python .claude/skills/deputy/scripts/api_wrapper.py \
  --endpoint resource/Employee/123
```

### 4. Rosters

Get roster for a date range:
```bash
python .claude/skills/deputy/scripts/api_wrapper.py \
  --endpoint resource/Roster/QUERY \
  --data '{"search":{"s1":{"field":"Date","type":"ge","data":"2025-10-01"}}}'
```

### 5. Leave Applications

Get leave applications:
```bash
python .claude/skills/deputy/scripts/api_wrapper.py \
  --endpoint resource/Leave
```

## Timesheet Validation

Deputy timesheets can be validated for payroll issues:

```bash
python .claude/skills/deputy/scripts/validate_timesheets.py \
  --input timesheets.json \
  --check-overtime \
  --check-breaks \
  --check-late-starts \
  --output validation_report.json
```

This checks for:
- Overtime hours (>38 hours/week in Australia)
- Meal break violations (>5 hours without break)
- Late clock-ins
- Early clock-outs
- Missing timesheet data
- Unusual patterns

## API Reference

For detailed API documentation, see [api-reference.md](api-reference.md).

### Base API Structure

All Deputy API calls use this pattern:
```bash
curl -X GET \
  "https://yourcompany.deputy.com/api/v1/{endpoint}" \
  -H "Accept: application/json" \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json"
```

### Common Endpoints

- `me` - Current user information
- `resource/Employee` - Employee management
- `resource/Timesheet` - Timesheet data
- `resource/Roster` - Roster/schedule data
- `resource/Leave` - Leave applications
- `resource/Location` - Location/site information
- `resource/OperationalUnit` - Department/area information

## Workflow Examples

### Export Timesheets for Payroll

1. Get timesheets from Deputy:
```bash
python .claude/skills/deputy/scripts/get_timesheets.py \
  --start-date 2025-10-01 \
  --end-date 2025-10-14 \
  --output deputy_timesheets.json
```

2. Validate timesheets:
```bash
python .claude/skills/deputy/scripts/validate_timesheets.py \
  --input deputy_timesheets.json \
  --check-all \
  --output validation_report.json
```

3. Review validation report for issues

4. Export to Xero format:
```bash
python .claude/skills/deputy/scripts/export_to_xero.py \
  --input deputy_timesheets.json \
  --output xero_import.json
```

5. Upload to Xero:
```bash
python .claude/skills/xero/scripts/import_timesheets.py \
  --input xero_import.json
```

### Daily Roster Check

```bash
python .claude/skills/deputy/scripts/get_roster.py \
  --date today \
  --notify-missing
```

### Leave Balance Report

```bash
python .claude/skills/deputy/scripts/get_leave_balances.py \
  --output leave_report.csv
```

## Common Queries

### Get Active Employees
```bash
python .claude/skills/deputy/scripts/api_wrapper.py \
  --endpoint resource/Employee/QUERY \
  --data '{"search":{"s1":{"field":"Active","type":"eq","data":"1"}}}'
```

### Get Today's Clock-Ins
```bash
python .claude/skills/deputy/scripts/get_timesheets.py \
  --start-date $(date +%Y-%m-%d) \
  --end-date $(date +%Y-%m-%d)
```

### Get Pending Leave Requests
```bash
python .claude/skills/deputy/scripts/api_wrapper.py \
  --endpoint resource/Leave/QUERY \
  --data '{"search":{"s1":{"field":"Status","type":"eq","data":"1"}}}'
```

## Error Handling

The scripts include comprehensive error handling:

1. **Authentication Errors**: Validates API token before requests
2. **Rate Limiting**: Respects Deputy API rate limits
3. **Network Errors**: Retries with exponential backoff
4. **Data Validation**: Validates timesheet data integrity

## Security Best Practices

1. Never commit API tokens to git
2. Use environment variables in production
3. Rotate API tokens regularly
4. Enable IP restrictions in Deputy if available
5. Log all API access for audit trail

## Tips for Claude Code Users

When I receive a request related to Deputy:

1. **Always verify authentication first**: Check if API token is valid
2. **Validate timesheet data**: Check for anomalies before export
3. **Check date ranges**: Ensure correct pay period
4. **Flag issues**: Alert user to overtime, missing breaks, etc.
5. **Summarize results**: Provide clear summaries of timesheet data
6. **Request confirmation**: For bulk operations, always confirm first

## Deputy-Specific Features

### Auto Clock-In/Out Detection

Deputy tracks actual clock-in/out times vs. rostered times:
- `StartTime` - Rostered start time
- `Start` - Actual clock-in time
- `EndTime` - Rostered end time
- `End` - Actual clock-out time

### Break Tracking

Deputy tracks breaks:
- `TotalBreakTime` - Total break time in minutes
- `Break1Start`, `Break1End` - Individual break times

### Award Interpretation

Deputy can calculate award-based pay rates:
- `OrdinaryHours` - Normal hours
- `OvertimeHours` - Overtime hours
- `DoubleTimeHours` - Double-time hours
- Various penalty rates and allowances

## Troubleshooting

### Authentication Issues
```bash
# Test API token
python .claude/skills/deputy/scripts/test_auth.py

# Regenerate token in Deputy and update config
```

### Missing Timesheets
```bash
# Check for incomplete timesheets
python .claude/skills/deputy/scripts/check_incomplete.py \
  --start-date 2025-10-01 \
  --end-date 2025-10-14
```

### API Errors
Check the error log:
```bash
tail -f logs/deputy_api.log
```

## Integration Notes

### Deputy → Xero Payroll
- Deputy timesheets map to Xero timesheet lines
- Deputy employees must be matched to Xero employees
- Award rates in Deputy should match pay items in Xero
- Leave types must be mapped between systems

### Deputy → Quickbooks
- Deputy timesheets can be exported to Quickbooks Time Tracking
- Employee mapping required
- Department/location mapping for job costing

## Next Steps

After setting up Deputy integration:
1. Test API access
2. Map employees to Xero/Quickbooks
3. Set up validation rules for your awards/agreements
4. Test timesheet export workflow
5. Configure automated exports
6. Set up anomaly alerts

For integration with Xero and Quickbooks, see the respective skill documentation.
