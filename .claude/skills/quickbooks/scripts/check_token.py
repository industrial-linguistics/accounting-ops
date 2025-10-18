#!/usr/bin/env python3
"""
Check Quickbooks Token Status

This script checks the status of your Quickbooks OAuth tokens and
indicates if they need to be refreshed.

Usage:
    python check_token.py
"""

import json
import sys
from pathlib import Path
from datetime import datetime, timedelta

def load_tokens():
    """Load current tokens."""
    tokens_path = Path(__file__).parent.parent.parent.parent / 'config' / 'quickbooks_tokens.json'

    if not tokens_path.exists():
        print(f"Error: Tokens file not found at {tokens_path}")
        print("Please run oauth_setup.py first to get initial tokens.")
        return None

    with open(tokens_path, 'r') as f:
        return json.load(f)

def main():
    """Check token status."""

    print("=" * 70)
    print("Quickbooks Token Status")
    print("=" * 70)

    tokens = load_tokens()
    if not tokens:
        sys.exit(1)

    # Check if we have required fields
    has_access_token = 'access_token' in tokens
    has_refresh_token = 'refresh_token' in tokens

    print(f"\nAccess Token: {'Present' if has_access_token else 'MISSING'}")
    print(f"Refresh Token: {'Present' if has_refresh_token else 'MISSING'}")

    if not has_access_token or not has_refresh_token:
        print("\nError: Missing tokens. Please run oauth_setup.py")
        sys.exit(1)

    # Check expiry
    if 'refreshed_at' in tokens and 'expires_in' in tokens:
        refreshed_at = datetime.fromisoformat(tokens['refreshed_at'])
        expires_in = int(tokens['expires_in'])
        expiry_time = refreshed_at + timedelta(seconds=expires_in)
        time_remaining = expiry_time - datetime.now()

        print(f"\nLast Refreshed: {refreshed_at.strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"Expires At: {expiry_time.strftime('%Y-%m-%d %H:%M:%S')}")

        if time_remaining.total_seconds() > 0:
            print(f"Time Remaining: {int(time_remaining.total_seconds())} seconds ({int(time_remaining.total_seconds()/60)} minutes)")

            if time_remaining.total_seconds() < 300:
                print("\nStatus: WARNING - Token expires soon (less than 5 minutes)")
                print("Action: Run refresh_token.py to get a new access token")
            else:
                print("\nStatus: OK - Token is valid")
        else:
            print(f"Time Remaining: EXPIRED {int(abs(time_remaining.total_seconds()))} seconds ago")
            print("\nStatus: EXPIRED")
            print("Action: Run refresh_token.py to get a new access token")

    else:
        print("\nWarning: No timestamp information available.")
        print("Cannot determine token expiry status.")
        print("\nRecommendation: Run refresh_token.py to be safe")

    # Check refresh token expiry
    if 'x_refresh_token_expires_in' in tokens and 'refreshed_at' in tokens:
        refresh_expires_in = int(tokens['x_refresh_token_expires_in'])
        refresh_expiry = refreshed_at + timedelta(seconds=refresh_expires_in)
        refresh_time_remaining = refresh_expiry - datetime.now()

        print(f"\nRefresh Token Expires At: {refresh_expiry.strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"Refresh Token Time Remaining: {int(refresh_time_remaining.total_seconds()/86400)} days")

        if refresh_time_remaining.total_seconds() < 86400 * 7:  # Less than 7 days
            print("\nWarning: Refresh token expires in less than 7 days!")
            print("Consider running oauth_setup.py again to get a new refresh token.")

    print("\n" + "=" * 70)

if __name__ == '__main__':
    main()
