#!/usr/bin/env python3
"""
Bank Reconciliation Script

This script performs automated bank reconciliation by matching Quickbooks
transactions with bank statement transactions.

Usage:
    python reconcile_bank.py --account-id 35 --statement-file bank_statement.csv
"""

import argparse
import json
import csv
import sys
from pathlib import Path
from datetime import datetime
from get_bank_transactions import get_bank_transactions
from api_wrapper import QuickbooksAPI
import logging

logger = logging.getLogger(__name__)

def parse_bank_statement(statement_file):
    """
    Parse bank statement CSV file.

    Expected CSV format:
    Date, Description, Debit, Credit, Balance

    Args:
        statement_file: Path to CSV file

    Returns:
        List of transaction dicts
    """
    transactions = []

    with open(statement_file, 'r') as f:
        reader = csv.DictReader(f)

        for row in reader:
            # Parse date (handle different formats)
            date_str = row.get('Date', row.get('date', ''))
            try:
                # Try different date formats
                for fmt in ['%Y-%m-%d', '%d/%m/%Y', '%m/%d/%Y', '%d-%m-%Y']:
                    try:
                        date = datetime.strptime(date_str, fmt).strftime('%Y-%m-%d')
                        break
                    except ValueError:
                        continue
            except:
                logger.warning(f"Could not parse date: {date_str}")
                continue

            # Parse amounts
            debit = row.get('Debit', row.get('debit', '0'))
            credit = row.get('Credit', row.get('credit', '0'))

            # Clean amount strings
            debit = debit.replace('$', '').replace(',', '').strip()
            credit = credit.replace('$', '').replace(',', '').strip()

            debit = float(debit) if debit and debit != '' else 0
            credit = float(credit) if credit and credit != '' else 0

            # Calculate net amount (credit positive, debit negative)
            amount = credit - debit

            transactions.append({
                'date': date,
                'description': row.get('Description', row.get('description', '')),
                'amount': amount,
                'balance': row.get('Balance', row.get('balance', '')),
                'raw': row
            })

    return transactions

def match_transactions(qb_transactions, bank_transactions, tolerance=0.01):
    """
    Match Quickbooks transactions with bank transactions.

    Args:
        qb_transactions: List of Quickbooks transactions
        bank_transactions: List of bank transactions
        tolerance: Amount tolerance for matching (default $0.01)

    Returns:
        Dict with matched, unmatched_qb, unmatched_bank lists
    """
    matched = []
    unmatched_qb = list(qb_transactions)
    unmatched_bank = list(bank_transactions)

    # Try to match each QB transaction
    for qb_txn in qb_transactions:
        best_match = None
        best_score = 0

        for bank_txn in bank_transactions:
            # Skip if already matched
            if bank_txn in [m['bank'] for m in matched]:
                continue

            score = 0

            # Date match (exact)
            if qb_txn['date'] == bank_txn['date']:
                score += 40

            # Amount match (within tolerance)
            amount_diff = abs(qb_txn['amount'] - bank_txn['amount'])
            if amount_diff <= tolerance:
                score += 50

            # Description similarity (basic)
            qb_desc = qb_txn.get('memo', '').lower()
            bank_desc = bank_txn.get('description', '').lower()

            if qb_desc and bank_desc:
                # Check for any common words
                qb_words = set(qb_desc.split())
                bank_words = set(bank_desc.split())
                common = qb_words.intersection(bank_words)
                if common:
                    score += 10 * len(common)

            # Consider it a match if score is high enough
            if score >= 90 and score > best_score:
                best_match = bank_txn
                best_score = score

        if best_match:
            matched.append({
                'quickbooks': qb_txn,
                'bank': best_match,
                'confidence': best_score
            })
            unmatched_qb.remove(qb_txn)
            unmatched_bank.remove(best_match)

    return {
        'matched': matched,
        'unmatched_quickbooks': unmatched_qb,
        'unmatched_bank': unmatched_bank
    }

def generate_reconciliation_report(results, account_id, start_date, end_date):
    """
    Generate a reconciliation report.

    Args:
        results: Matching results dict
        account_id: Account ID
        start_date: Start date
        end_date: End date

    Returns:
        Report dict
    """
    report = {
        'metadata': {
            'account_id': account_id,
            'date_range': {
                'start': start_date,
                'end': end_date
            },
            'generated_at': datetime.now().isoformat(),
        },
        'summary': {
            'total_matched': len(results['matched']),
            'unmatched_quickbooks': len(results['unmatched_quickbooks']),
            'unmatched_bank': len(results['unmatched_bank']),
            'match_rate': len(results['matched']) / (len(results['matched']) + len(results['unmatched_quickbooks'])) * 100 if results['matched'] or results['unmatched_quickbooks'] else 0,
        },
        'matched_transactions': results['matched'],
        'unmatched_quickbooks_transactions': results['unmatched_quickbooks'],
        'unmatched_bank_transactions': results['unmatched_bank']
    }

    return report

def print_reconciliation_summary(report):
    """Print a human-readable reconciliation summary."""

    print("\n" + "=" * 70)
    print("BANK RECONCILIATION REPORT")
    print("=" * 70)

    print(f"\nAccount ID: {report['metadata']['account_id']}")
    print(f"Period: {report['metadata']['date_range']['start']} to {report['metadata']['date_range']['end']}")
    print(f"Generated: {report['metadata']['generated_at']}")

    print("\n" + "-" * 70)
    print("SUMMARY")
    print("-" * 70)
    print(f"  Matched Transactions: {report['summary']['total_matched']}")
    print(f"  Unmatched in Quickbooks: {report['summary']['unmatched_quickbooks']}")
    print(f"  Unmatched in Bank Statement: {report['summary']['unmatched_bank']}")
    print(f"  Match Rate: {report['summary']['match_rate']:.1f}%")

    if report['unmatched_quickbooks_transactions']:
        print("\n" + "-" * 70)
        print("UNMATCHED QUICKBOOKS TRANSACTIONS")
        print("-" * 70)
        for txn in report['unmatched_quickbooks_transactions']:
            print(f"  {txn['date']} | ${txn['amount']:>10.2f} | {txn['payee']} | {txn['memo']}")

    if report['unmatched_bank_transactions']:
        print("\n" + "-" * 70)
        print("UNMATCHED BANK TRANSACTIONS")
        print("-" * 70)
        for txn in report['unmatched_bank_transactions']:
            print(f"  {txn['date']} | ${txn['amount']:>10.2f} | {txn['description']}")

    print("\n" + "=" * 70)

def main():
    """Main reconciliation workflow."""

    parser = argparse.ArgumentParser(description='Quickbooks Bank Reconciliation')

    parser.add_argument('--account-id', required=True,
                        help='Bank account ID')
    parser.add_argument('--statement-file', required=True,
                        help='Bank statement CSV file')
    parser.add_argument('--start-date',
                        help='Start date (YYYY-MM-DD) - optional if in statement')
    parser.add_argument('--end-date',
                        help='End date (YYYY-MM-DD) - optional if in statement')
    parser.add_argument('--output', default='reconciliation_report.json',
                        help='Output report file')
    parser.add_argument('--tolerance', type=float, default=0.01,
                        help='Amount matching tolerance (default 0.01)')

    args = parser.parse_args()

    print("=" * 70)
    print("QUICKBOOKS BANK RECONCILIATION")
    print("=" * 70)

    # Parse bank statement
    print(f"\n1. Parsing bank statement: {args.statement_file}")
    try:
        bank_transactions = parse_bank_statement(args.statement_file)
        print(f"   Found {len(bank_transactions)} bank transactions")
    except Exception as e:
        logger.error(f"Failed to parse bank statement: {e}")
        print(f"Error: Failed to parse bank statement: {e}")
        sys.exit(1)

    # Determine date range
    if args.start_date and args.end_date:
        start_date = args.start_date
        end_date = args.end_date
    else:
        # Use date range from bank transactions
        dates = [t['date'] for t in bank_transactions]
        start_date = min(dates)
        end_date = max(dates)
        print(f"   Date range from statement: {start_date} to {end_date}")

    # Get Quickbooks transactions
    print(f"\n2. Fetching Quickbooks transactions...")
    api = QuickbooksAPI()
    api.initialize()

    try:
        qb_transactions = get_bank_transactions(api, args.account_id, start_date, end_date)
        print(f"   Found {len(qb_transactions)} Quickbooks transactions")
    except Exception as e:
        logger.error(f"Failed to get Quickbooks transactions: {e}")
        print(f"Error: Failed to get Quickbooks transactions: {e}")
        sys.exit(1)

    # Match transactions
    print(f"\n3. Matching transactions (tolerance: ${args.tolerance})...")
    results = match_transactions(qb_transactions, bank_transactions, args.tolerance)

    # Generate report
    print(f"\n4. Generating reconciliation report...")
    report = generate_reconciliation_report(results, args.account_id, start_date, end_date)

    # Save report
    with open(args.output, 'w') as f:
        json.dump(report, f, indent=2)
    print(f"   Report saved to: {args.output}")

    # Print summary
    print_reconciliation_summary(report)

    # Exit code based on match rate
    if report['summary']['match_rate'] < 95:
        print("\nWARNING: Match rate below 95%. Manual review recommended.")
        sys.exit(1)
    else:
        print("\nSUCCESS: Reconciliation complete with good match rate.")
        sys.exit(0)

if __name__ == '__main__':
    main()
