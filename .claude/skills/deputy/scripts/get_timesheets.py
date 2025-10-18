#!/usr/bin/env python3
"""
Get Timesheets from Deputy

This script retrieves timesheets from Deputy for a specified date range.

Usage:
    python get_timesheets.py --start-date 2025-10-01 --end-date 2025-10-18
"""

import argparse
import json
import sys
from pathlib import Path
from datetime import datetime
from api_wrapper import DeputyAPI
import logging

logger = logging.getLogger(__name__)

def build_timesheet_query(start_date, end_date, employee_id=None):
    """
    Build a Deputy query for timesheets.

    Args:
        start_date: Start date (YYYY-MM-DD)
        end_date: End date (YYYY-MM-DD)
        employee_id: Optional employee ID to filter

    Returns:
        Query dict
    """
    # Convert dates to Unix timestamps
    start_timestamp = int(datetime.strptime(start_date, '%Y-%m-%d').timestamp())
    end_timestamp = int(datetime.strptime(end_date, '%Y-%m-%d').timestamp())

    query = {
        "search": {
            "s1": {
                "field": "Date",
                "type": "ge",  # greater than or equal
                "data": start_timestamp
            },
            "s2": {
                "field": "Date",
                "type": "le",  # less than or equal
                "data": end_timestamp
            }
        }
    }

    if employee_id:
        query["search"]["s3"] = {
            "field": "Employee",
            "type": "eq",
            "data": employee_id
        }

    return query

def process_timesheets(timesheets):
    """
    Process and enrich timesheet data.

    Args:
        timesheets: List of Deputy timesheets

    Returns:
        Processed timesheets with calculated fields
    """
    processed = []

    for ts in timesheets:
        # Calculate total hours
        start_time = ts.get('StartTime', ts.get('Start'))
        end_time = ts.get('EndTime', ts.get('End'))

        if start_time and end_time:
            duration_seconds = end_time - start_time
            total_hours = duration_seconds / 3600

            # Subtract break time
            break_time_seconds = ts.get('TotalBreak', 0)
            net_hours = total_hours - (break_time_seconds / 3600)
        else:
            total_hours = 0
            net_hours = 0

        processed.append({
            'id': ts.get('Id'),
            'employee_id': ts.get('Employee'),
            'date': datetime.fromtimestamp(ts.get('Date', 0)).strftime('%Y-%m-%d'),
            'start_time': datetime.fromtimestamp(start_time).isoformat() if start_time else None,
            'end_time': datetime.fromtimestamp(end_time).isoformat() if end_time else None,
            'total_hours': round(total_hours, 2),
            'break_hours': round(ts.get('TotalBreak', 0) / 3600, 2),
            'net_hours': round(net_hours, 2),
            'approved': ts.get('Approved', False),
            'raw': ts
        })

    return processed

def generate_timesheet_summary(timesheets, start_date, end_date):
    """Generate summary statistics."""

    total_hours = sum(ts['net_hours'] for ts in timesheets)
    total_shifts = len(timesheets)
    approved = sum(1 for ts in timesheets if ts['approved'])
    pending = total_shifts - approved

    # Group by employee
    by_employee = {}
    for ts in timesheets:
        emp_id = ts['employee_id']
        if emp_id not in by_employee:
            by_employee[emp_id] = {
                'shifts': 0,
                'total_hours': 0
            }
        by_employee[emp_id]['shifts'] += 1
        by_employee[emp_id]['total_hours'] += ts['net_hours']

    return {
        'metadata': {
            'date_range': {
                'start': start_date,
                'end': end_date
            },
            'generated_at': datetime.now().isoformat()
        },
        'summary': {
            'total_shifts': total_shifts,
            'total_hours': round(total_hours, 2),
            'approved_shifts': approved,
            'pending_shifts': pending,
            'employees': len(by_employee)
        },
        'by_employee': by_employee,
        'timesheets': timesheets
    }

def print_summary(report):
    """Print human-readable summary."""

    print("\n" + "=" * 70)
    print("DEPUTY TIMESHEET SUMMARY")
    print("=" * 70)

    print(f"\nPeriod: {report['metadata']['date_range']['start']} to {report['metadata']['date_range']['end']}")
    print(f"Generated: {report['metadata']['generated_at']}")

    print("\n" + "-" * 70)
    print("SUMMARY")
    print("-" * 70)
    print(f"  Total Shifts: {report['summary']['total_shifts']}")
    print(f"  Total Hours: {report['summary']['total_hours']:.2f}")
    print(f"  Approved Shifts: {report['summary']['approved_shifts']}")
    print(f"  Pending Approval: {report['summary']['pending_shifts']}")
    print(f"  Employees: {report['summary']['employees']}")

    if report['by_employee']:
        print("\n" + "-" * 70)
        print("BY EMPLOYEE")
        print("-" * 70)
        for emp_id, data in report['by_employee'].items():
            print(f"  Employee {emp_id}: {data['shifts']} shifts, {data['total_hours']:.2f} hours")

    print("\n" + "=" * 70)

def main():
    """Main workflow."""

    parser = argparse.ArgumentParser(description='Get Deputy Timesheets')

    parser.add_argument('--start-date', required=True,
                        help='Start date (YYYY-MM-DD)')
    parser.add_argument('--end-date', required=True,
                        help='End date (YYYY-MM-DD)')
    parser.add_argument('--employee-id', type=int,
                        help='Filter by employee ID')
    parser.add_argument('--output', default='deputy_timesheets.json',
                        help='Output file')

    args = parser.parse_args()

    print("=" * 70)
    print("DEPUTY TIMESHEET RETRIEVAL")
    print("=" * 70)

    # Initialize API
    api = DeputyAPI()
    api.load_credentials()

    try:
        # Build query
        query = build_timesheet_query(args.start_date, args.end_date, args.employee_id)

        # Get timesheets
        print(f"\nFetching timesheets from {args.start_date} to {args.end_date}...")
        response = api.get_timesheets(query)

        # Process results
        raw_timesheets = response if isinstance(response, list) else response.get('data', [])
        print(f"Retrieved {len(raw_timesheets)} timesheets")

        # Process timesheets
        print("\nProcessing timesheet data...")
        processed_timesheets = process_timesheets(raw_timesheets)

        # Generate summary
        report = generate_timesheet_summary(processed_timesheets, args.start_date, args.end_date)

        # Save report
        with open(args.output, 'w') as f:
            json.dump(report, f, indent=2)
        print(f"Report saved to: {args.output}")

        # Print summary
        print_summary(report)

    except Exception as e:
        logger.error(f"Failed to get timesheets: {e}")
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    main()
