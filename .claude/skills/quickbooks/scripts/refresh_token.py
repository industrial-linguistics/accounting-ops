#!/usr/bin/env python3
"""
Quickbooks Token Refresh Script

This script refreshes the Quickbooks OAuth access token using the refresh token.
Access tokens expire after 1 hour, so this should be run periodically.

Usage:
    python refresh_token.py
"""

import json
import os
import sys
import subprocess
import base64
from pathlib import Path
from datetime import datetime, timedelta

def load_credentials():
    """Load Quickbooks credentials."""
    config_path = Path(__file__).parent.parent.parent.parent / 'config' / 'quickbooks_credentials.json'

    if not config_path.exists():
        print(f"Error: Credentials file not found at {config_path}")
        sys.exit(1)

    with open(config_path, 'r') as f:
        return json.load(f)

def load_tokens():
    """Load current tokens."""
    tokens_path = Path(__file__).parent.parent.parent.parent / 'config' / 'quickbooks_tokens.json'

    if not tokens_path.exists():
        print(f"Error: Tokens file not found at {tokens_path}")
        print("Please run oauth_setup.py first to get initial tokens.")
        sys.exit(1)

    with open(tokens_path, 'r') as f:
        return json.load(f)

def save_tokens(tokens):
    """Save tokens to file."""
    tokens_path = Path(__file__).parent.parent.parent.parent / 'config' / 'quickbooks_tokens.json'

    # Add timestamp
    tokens['refreshed_at'] = datetime.now().isoformat()

    with open(tokens_path, 'w') as f:
        json.dump(tokens, f, indent=2)

    print(f"Tokens saved to {tokens_path}")

def refresh_access_token(credentials, refresh_token):
    """Refresh the access token using curl."""

    # Create Basic Auth header
    auth_string = f"{credentials['client_id']}:{credentials['client_secret']}"
    auth_bytes = auth_string.encode('ascii')
    auth_b64 = base64.b64encode(auth_bytes).decode('ascii')

    # Build curl command
    curl_cmd = [
        'curl', '-X', 'POST',
        credentials['token_url'],
        '-H', 'Accept: application/json',
        '-H', 'Content-Type: application/x-www-form-urlencoded',
        '-H', f'Authorization: Basic {auth_b64}',
        '-d', f'grant_type=refresh_token',
        '-d', f'refresh_token={refresh_token}'
    ]

    print("Refreshing access token...")

    try:
        result = subprocess.run(curl_cmd, capture_output=True, text=True, check=True)
        response = json.loads(result.stdout)

        if 'access_token' in response:
            print("Success! New access token obtained.")
            print(f"Access token expires in: {response.get('expires_in', 'unknown')} seconds")
            print(f"Refresh token expires in: {response.get('x_refresh_token_expires_in', 'unknown')} seconds")
            return response
        else:
            print(f"Error: Unexpected response: {response}")
            return None

    except subprocess.CalledProcessError as e:
        print(f"Error executing curl: {e}")
        print(f"stdout: {e.stdout}")
        print(f"stderr: {e.stderr}")
        return None
    except json.JSONDecodeError as e:
        print(f"Error parsing response: {e}")
        return None

def check_token_expiry(tokens):
    """Check if token needs refresh."""

    if 'refreshed_at' in tokens and 'expires_in' in tokens:
        refreshed_at = datetime.fromisoformat(tokens['refreshed_at'])
        expires_in = int(tokens['expires_in'])
        expiry_time = refreshed_at + timedelta(seconds=expires_in)
        time_remaining = expiry_time - datetime.now()

        print(f"\nToken status:")
        print(f"  Last refreshed: {refreshed_at}")
        print(f"  Expires at: {expiry_time}")
        print(f"  Time remaining: {time_remaining}")

        if time_remaining.total_seconds() < 300:  # Less than 5 minutes
            print("  Status: NEEDS REFRESH (expires in less than 5 minutes)")
            return True
        else:
            print("  Status: OK")
            return False
    else:
        print("  Status: UNKNOWN (no refresh timestamp)")
        return True

def main():
    """Main refresh flow."""

    print("=" * 70)
    print("Quickbooks Token Refresh")
    print("=" * 70)

    # Load credentials and tokens
    credentials = load_credentials()
    tokens = load_tokens()

    # Check if refresh is needed
    needs_refresh = check_token_expiry(tokens)

    if not needs_refresh:
        response = input("\nToken is still valid. Refresh anyway? (y/n): ")
        if response.lower() != 'y':
            print("Refresh cancelled.")
            sys.exit(0)

    # Get refresh token
    refresh_token = tokens.get('refresh_token')
    if not refresh_token:
        print("Error: No refresh token found in tokens file.")
        print("Please run oauth_setup.py again to get new tokens.")
        sys.exit(1)

    # Refresh the token
    new_tokens = refresh_access_token(credentials, refresh_token)

    if new_tokens:
        # Merge with existing tokens (preserve any extra fields)
        tokens.update(new_tokens)
        save_tokens(tokens)

        print("\n" + "=" * 70)
        print("SUCCESS! Token refreshed.")
        print("=" * 70)
    else:
        print("\nToken refresh failed. Please check the errors above.")
        print("If refresh token has expired, run oauth_setup.py again.")
        sys.exit(1)

if __name__ == '__main__':
    main()
