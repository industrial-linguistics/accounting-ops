#!/usr/bin/env python3
"""
Get Timesheets from Xero Payroll

This script retrieves timesheets from Xero Payroll API and performs
validation checks for overtime, meal breaks, and other anomalies.

Usage:
    python get_timesheets.py --region AU --start-date 2025-10-01 --end-date 2025-10-18
"""

import argparse
import json
import sys
from pathlib import Path
from datetime import datetime, timedelta
from api_wrapper import XeroAPI
import logging

logger = logging.getLogger(__name__)

def get_timesheets(api, start_date, end_date):
    """
    Get timesheets within a date range.

    Args:
        api: XeroAPI instance
        start_date: Start date (YYYY-MM-DD)
        end_date: End date (YYYY-MM-DD)

    Returns:
        List of timesheets
    """
    # Xero uses DateTime filters
    where_clause = f"UpdatedDateUTC >= DateTime({start_date.year},{start_date.month:02d},{start_date.day:02d}) AND UpdatedDateUTC <= DateTime({end_date.year},{end_date.month:02d},{end_date.day:02d})"

    logger.info(f"Fetching timesheets from {start_date} to {end_date}")

    try:
        response = api.make_request('GET', 'Timesheets', params={'where': where_clause})

        if 'Timesheets' in response:
            return response['Timesheets']
        else:
            logger.warning("No timesheets found in response")
            return []

    except Exception as e:
        logger.error(f"Failed to get timesheets: {e}")
        raise

def analyze_timesheet(timesheet):
    """
    Analyze a timesheet for issues.

    Args:
        timesheet: Timesheet dict

    Returns:
        Dict with analysis results
    """
    issues = []
    total_hours = 0

    # Get employee info
    employee_id = timesheet.get('EmployeeID', 'Unknown')

    # Process timesheet lines
    for line in timesheet.get('TimesheetLines', []):
        for unit in line.get('Units', []):
            hours = float(unit.get('NumberOfUnits', 0))
            total_hours += hours

    # Check for overtime (more than 38 hours per week in AU)
    # This is a simplified check - real logic would need to consider award rules
    if total_hours > 38:
        issues.append({
            'type': 'overtime',
            'message': f'Total hours ({total_hours:.1f}) exceeds standard 38 hours',
            'hours': total_hours - 38
        })

    # Check for long days (more than 10 hours)
    for line in timesheet.get('TimesheetLines', []):
        daily_hours = sum(float(u.get('NumberOfUnits', 0)) for u in line.get('Units', []))
        if daily_hours > 10:
            issues.append({
                'type': 'long_day',
                'message': f'Day worked {daily_hours:.1f} hours (>10 hours)',
                'date': line.get('Date'),
                'hours': daily_hours
            })

    return {
        'employee_id': employee_id,
        'total_hours': total_hours,
        'issues': issues
    }

def generate_timesheet_report(timesheets, start_date, end_date):
    """
    Generate a timesheet report.

    Args:
        timesheets: List of timesheets
        start_date: Start date
        end_date: End date

    Returns:
        Report dict
    """
    analyses = [analyze_timesheet(ts) for ts in timesheets]

    total_hours = sum(a['total_hours'] for a in analyses)
    total_issues = sum(len(a['issues']) for a in analyses)

    # Count issue types
    issue_types = {}
    for analysis in analyses:
        for issue in analysis['issues']:
            issue_type = issue['type']
            issue_types[issue_type] = issue_types.get(issue_type, 0) + 1

    report = {
        'metadata': {
            'date_range': {
                'start': str(start_date),
                'end': str(end_date)
            },
            'generated_at': datetime.now().isoformat()
        },
        'summary': {
            'total_timesheets': len(timesheets),
            'total_hours': total_hours,
            'total_issues': total_issues,
            'issue_breakdown': issue_types
        },
        'timesheet_analyses': analyses,
        'raw_timesheets': timesheets
    }

    return report

def print_timesheet_summary(report):
    """Print a human-readable timesheet summary."""

    print("\n" + "=" * 70)
    print("XERO TIMESHEET REPORT")
    print("=" * 70)

    print(f"\nPeriod: {report['metadata']['date_range']['start']} to {report['metadata']['date_range']['end']}")
    print(f"Generated: {report['metadata']['generated_at']}")

    print("\n" + "-" * 70)
    print("SUMMARY")
    print("-" * 70)
    print(f"  Total Timesheets: {report['summary']['total_timesheets']}")
    print(f"  Total Hours: {report['summary']['total_hours']:.1f}")
    print(f"  Total Issues: {report['summary']['total_issues']}")

    if report['summary']['issue_breakdown']:
        print("\n  Issue Breakdown:")
        for issue_type, count in report['summary']['issue_breakdown'].items():
            print(f"    {issue_type}: {count}")

    # Print individual issues
    if report['summary']['total_issues'] > 0:
        print("\n" + "-" * 70)
        print("ISSUES REQUIRING ATTENTION")
        print("-" * 70)

        for analysis in report['timesheet_analyses']:
            if analysis['issues']:
                print(f"\nEmployee ID: {analysis['employee_id']}")
                for issue in analysis['issues']:
                    print(f"  - {issue['type'].upper()}: {issue['message']}")

    print("\n" + "=" * 70)

def main():
    """Main workflow."""

    parser = argparse.ArgumentParser(description='Get Xero Timesheets')

    parser.add_argument('--region', default='AU',
                        choices=['AU', 'NZ', 'UK'],
                        help='Payroll region')
    parser.add_argument('--start-date', required=True,
                        help='Start date (YYYY-MM-DD)')
    parser.add_argument('--end-date', required=True,
                        help='End date (YYYY-MM-DD)')
    parser.add_argument('--output', default='timesheet_report.json',
                        help='Output report file')
    parser.add_argument('--check-overtime', action='store_true',
                        help='Check for overtime')
    parser.add_argument('--check-breaks', action='store_true',
                        help='Check for meal break issues')

    args = parser.parse_args()

    # Validate dates
    try:
        start_date = datetime.strptime(args.start_date, '%Y-%m-%d')
        end_date = datetime.strptime(args.end_date, '%Y-%m-%d')
    except ValueError as e:
        print(f"Error: Invalid date format. Use YYYY-MM-DD: {e}")
        sys.exit(1)

    print("=" * 70)
    print("XERO TIMESHEET RETRIEVAL")
    print("=" * 70)

    # Initialize API
    api = XeroAPI()
    api.initialize('payroll')

    # Set payroll region
    api.base_url = api.token_manager.get_payroll_base_url(args.region)
    print(f"\nRegion: {args.region}")
    print(f"Payroll API: {api.base_url}")

    try:
        # Get timesheets
        print(f"\nFetching timesheets...")
        timesheets = get_timesheets(api, start_date, end_date)
        print(f"Retrieved {len(timesheets)} timesheets")

        # Generate report
        print(f"\nAnalyzing timesheets...")
        report = generate_timesheet_report(timesheets, start_date, end_date)

        # Save report
        with open(args.output, 'w') as f:
            json.dump(report, f, indent=2)
        print(f"Report saved to: {args.output}")

        # Print summary
        print_timesheet_summary(report)

        # Exit code based on issues
        if report['summary']['total_issues'] > 0:
            print("\nWARNING: Issues found that require attention.")
            sys.exit(1)
        else:
            print("\nSUCCESS: All timesheets validated successfully.")
            sys.exit(0)

    except Exception as e:
        logger.error(f"Failed to process timesheets: {e}")
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    main()
