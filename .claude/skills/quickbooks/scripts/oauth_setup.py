#!/usr/bin/env python3
"""
Quickbooks OAuth 2.0 Setup Script

This script helps you set up OAuth 2.0 authentication with Quickbooks Online.
It generates the authorization URL and handles the token exchange.

Usage:
    python oauth_setup.py

After running, you'll get an authorization URL to visit in your browser.
After authorizing, you'll be redirected to your callback URL with a code.
Paste that code back into this script to complete the setup.
"""

import json
import os
import sys
import urllib.parse
import secrets
import webbrowser
from pathlib import Path

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent))

def load_credentials():
    """Load Quickbooks credentials from config file."""
    config_path = Path(__file__).parent.parent.parent.parent / 'config' / 'quickbooks_credentials.json'

    if not config_path.exists():
        print(f"Error: Credentials file not found at {config_path}")
        print("Please copy quickbooks_credentials.json.template and fill in your credentials.")
        sys.exit(1)

    with open(config_path, 'r') as f:
        return json.load(f)

def generate_auth_url(credentials):
    """Generate the OAuth authorization URL."""

    # Generate state for CSRF protection
    state = secrets.token_urlsafe(32)

    # Save state for verification
    state_path = Path(__file__).parent.parent.parent.parent / 'config' / 'oauth_state.txt'
    with open(state_path, 'w') as f:
        f.write(state)

    # Build authorization URL
    params = {
        'client_id': credentials['client_id'],
        'redirect_uri': credentials['redirect_uri'],
        'response_type': 'code',
        'scope': 'com.intuit.quickbooks.accounting',
        'state': state
    }

    auth_url = f"{credentials['auth_url']}?{urllib.parse.urlencode(params)}"

    return auth_url, state

def exchange_code_for_tokens(credentials, code):
    """Exchange authorization code for access and refresh tokens using curl."""

    import subprocess
    import base64

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
        '-d', f'grant_type=authorization_code',
        '-d', f'code={code}',
        '-d', f'redirect_uri={credentials["redirect_uri"]}'
    ]

    print("\nExecuting token exchange...")
    print(f"curl command: {' '.join(curl_cmd[:3])} ... [credentials hidden]")

    try:
        result = subprocess.run(curl_cmd, capture_output=True, text=True, check=True)
        response = json.loads(result.stdout)

        if 'access_token' in response:
            # Save tokens
            tokens_path = Path(__file__).parent.parent.parent.parent / 'config' / 'quickbooks_tokens.json'
            with open(tokens_path, 'w') as f:
                json.dump(response, f, indent=2)

            print(f"\nSuccess! Tokens saved to {tokens_path}")
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

def main():
    """Main setup flow."""

    print("=" * 70)
    print("Quickbooks OAuth 2.0 Setup")
    print("=" * 70)

    # Load credentials
    print("\n1. Loading credentials...")
    credentials = load_credentials()
    print(f"   Client ID: {credentials['client_id'][:10]}...")
    print(f"   Environment: {credentials['environment']}")

    # Generate authorization URL
    print("\n2. Generating authorization URL...")
    auth_url, state = generate_auth_url(credentials)

    print("\n" + "=" * 70)
    print("AUTHORIZATION URL:")
    print("=" * 70)
    print(auth_url)
    print("=" * 70)

    # Ask if we should open browser
    response = input("\nOpen this URL in your browser? (y/n): ")
    if response.lower() == 'y':
        webbrowser.open(auth_url)
        print("Browser opened. Please authorize the application.")
    else:
        print("\nPlease manually open the URL above in your browser.")

    print("\nAfter authorizing, you'll be redirected to your callback URL.")
    print("The URL will contain a 'code' parameter.")
    print("\nExample: http://localhost:8000/callback?code=AB12345...")

    # Get the full callback URL or just the code
    print("\n" + "=" * 70)
    callback_input = input("\nPaste the full callback URL or just the code: ").strip()

    # Extract code from URL if full URL was provided
    if 'code=' in callback_input:
        parsed = urllib.parse.urlparse(callback_input)
        params = urllib.parse.parse_qs(parsed.query)
        code = params.get('code', [None])[0]
        returned_state = params.get('state', [None])[0]
        realm_id = params.get('realmId', [None])[0]

        # Verify state
        if returned_state != state:
            print("Error: State mismatch! Possible CSRF attack.")
            sys.exit(1)

        if realm_id:
            print(f"\nRealm ID (Company ID): {realm_id}")
            print("Save this Realm ID - you'll need it for API calls!")

            # Save realm ID
            realm_path = Path(__file__).parent.parent.parent.parent / 'config' / 'quickbooks_realm_id.txt'
            with open(realm_path, 'w') as f:
                f.write(realm_id)
            print(f"Realm ID saved to {realm_path}")
    else:
        code = callback_input

    if not code:
        print("Error: No authorization code found.")
        sys.exit(1)

    print(f"\nAuthorization code: {code[:20]}...")

    # Exchange code for tokens
    print("\n3. Exchanging code for tokens...")
    tokens = exchange_code_for_tokens(credentials, code)

    if tokens:
        print("\n" + "=" * 70)
        print("SUCCESS! OAuth setup complete.")
        print("=" * 70)
        print("\nYou can now use the Quickbooks API.")
        print("Your access token will expire in 1 hour.")
        print("Use refresh_token.py to get a new access token.")
    else:
        print("\nSetup failed. Please check the errors above.")
        sys.exit(1)

if __name__ == '__main__':
    main()
