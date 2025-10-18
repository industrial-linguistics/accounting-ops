#!/usr/bin/env python3
"""
Get Bank Transactions from Quickbooks

This script retrieves bank transactions for a specified account and date range.
This is used for bank reconciliation workflows.

Usage:
    python get_bank_transactions.py --account-id 123 --start-date 2025-10-01 --end-date 2025-10-18
"""

import argparse
import json
import sys
from pathlib import Path
from datetime import datetime
from api_wrapper import QuickbooksAPI
import logging

logger = logging.getLogger(__name__)

def get_bank_transactions(api, account_id, start_date, end_date):
    """
    Get bank transactions for an account within a date range.

    Args:
        api: QuickbooksAPI instance
        account_id: Account ID
        start_date: Start date (YYYY-MM-DD)
        end_date: End date (YYYY-MM-DD)

    Returns:
        List of transactions
    """
    # Query for purchases (debits/withdrawals)
    purchase_query = f"""
        SELECT * FROM Purchase
        WHERE TxnDate >= '{start_date}'
        AND TxnDate <= '{end_date}'
        AND AccountRef = '{account_id}'
    """

    # Query for deposits (credits)
    deposit_query = f"""
        SELECT * FROM Deposit
        WHERE TxnDate >= '{start_date}'
        AND TxnDate <= '{end_date}'
        AND DepositToAccountRef = '{account_id}'
    """

    # Query for payments
    payment_query = f"""
        SELECT * FROM Payment
        WHERE TxnDate >= '{start_date}'
        AND TxnDate <= '{end_date}'
    """

    transactions = []

    # Get purchases
    logger.info("Fetching purchases...")
    purchase_response = api.query(purchase_query)
    if 'QueryResponse' in purchase_response and 'Purchase' in purchase_response['QueryResponse']:
        purchases = purchase_response['QueryResponse']['Purchase']
        for p in purchases:
            transactions.append({
                'type': 'Purchase',
                'id': p.get('Id'),
                'date': p.get('TxnDate'),
                'amount': -float(p.get('TotalAmt', 0)),  # Negative for withdrawal
                'payee': p.get('EntityRef', {}).get('name', 'Unknown'),
                'memo': p.get('PrivateNote', ''),
                'reconciled': p.get('CreditCardPayment', {}).get('CreditCardAccountRef', {}).get('name', ''),
                'raw': p
            })

    # Get deposits
    logger.info("Fetching deposits...")
    deposit_response = api.query(deposit_query)
    if 'QueryResponse' in deposit_response and 'Deposit' in deposit_response['QueryResponse']:
        deposits = deposit_response['QueryResponse']['Deposit']
        for d in deposits:
            transactions.append({
                'type': 'Deposit',
                'id': d.get('Id'),
                'date': d.get('TxnDate'),
                'amount': float(d.get('TotalAmt', 0)),  # Positive for deposit
                'payee': 'Deposit',
                'memo': d.get('PrivateNote', ''),
                'raw': d
            })

    # Get payments
    logger.info("Fetching payments...")
    payment_response = api.query(payment_query)
    if 'QueryResponse' in payment_response and 'Payment' in payment_response['QueryResponse']:
        payments = payment_response['QueryResponse']['Payment']
        for p in payments:
            transactions.append({
                'type': 'Payment',
                'id': p.get('Id'),
                'date': p.get('TxnDate'),
                'amount': float(p.get('TotalAmt', 0)),
                'payee': p.get('CustomerRef', {}).get('name', 'Unknown'),
                'memo': p.get('PrivateNote', ''),
                'raw': p
            })

    # Sort by date
    transactions.sort(key=lambda x: x['date'])

    return transactions

def main():
    """Main CLI."""
    parser = argparse.ArgumentParser(description='Get Quickbooks Bank Transactions')

    parser.add_argument('--account-id', required=True,
                        help='Bank account ID')
    parser.add_argument('--start-date', required=True,
                        help='Start date (YYYY-MM-DD)')
    parser.add_argument('--end-date', required=True,
                        help='End date (YYYY-MM-DD)')
    parser.add_argument('--output', help='Output JSON file')

    args = parser.parse_args()

    # Validate dates
    try:
        start = datetime.strptime(args.start_date, '%Y-%m-%d')
        end = datetime.strptime(args.end_date, '%Y-%m-%d')
    except ValueError as e:
        print(f"Error: Invalid date format. Use YYYY-MM-DD: {e}")
        sys.exit(1)

    # Initialize API
    api = QuickbooksAPI()
    api.initialize()

    try:
        # Get transactions
        transactions = get_bank_transactions(
            api, args.account_id, args.start_date, args.end_date
        )

        # Create summary
        summary = {
            'account_id': args.account_id,
            'date_range': {
                'start': args.start_date,
                'end': args.end_date
            },
            'transaction_count': len(transactions),
            'total_debits': sum(t['amount'] for t in transactions if t['amount'] < 0),
            'total_credits': sum(t['amount'] for t in transactions if t['amount'] > 0),
            'transactions': transactions
        }

        # Output
        if args.output:
            with open(args.output, 'w') as f:
                json.dump(summary, f, indent=2)
            print(f"Retrieved {len(transactions)} transactions")
            print(f"Saved to {args.output}")
        else:
            print(json.dumps(summary, indent=2))

        # Print summary
        print(f"\nSummary:")
        print(f"  Transactions: {len(transactions)}")
        print(f"  Total Debits: ${summary['total_debits']:,.2f}")
        print(f"  Total Credits: ${summary['total_credits']:,.2f}")
        print(f"  Net: ${(summary['total_credits'] + summary['total_debits']):,.2f}")

    except Exception as e:
        logger.error(f"Failed to get transactions: {e}")
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    main()
