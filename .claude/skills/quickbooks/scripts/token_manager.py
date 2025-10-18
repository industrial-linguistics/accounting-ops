#!/usr/bin/env python3
"""
Token Manager Utility

This module provides utilities for managing Quickbooks OAuth tokens,
including automatic refresh when needed.

Usage:
    from token_manager import get_valid_token

    token = get_valid_token()
    # Use token for API calls
"""

import json
import sys
import subprocess
import base64
from pathlib import Path
from datetime import datetime, timedelta

class TokenManager:
    """Manages Quickbooks OAuth tokens."""

    def __init__(self, config_dir=None):
        """Initialize token manager."""
        if config_dir is None:
            config_dir = Path(__file__).parent.parent.parent.parent / 'config'
        else:
            config_dir = Path(config_dir)

        self.config_dir = config_dir
        self.credentials_path = config_dir / 'quickbooks_credentials.json'
        self.tokens_path = config_dir / 'quickbooks_tokens.json'
        self.realm_id_path = config_dir / 'quickbooks_realm_id.txt'

    def load_credentials(self):
        """Load credentials from file."""
        if not self.credentials_path.exists():
            raise FileNotFoundError(f"Credentials file not found: {self.credentials_path}")

        with open(self.credentials_path, 'r') as f:
            return json.load(f)

    def load_tokens(self):
        """Load tokens from file."""
        if not self.tokens_path.exists():
            raise FileNotFoundError(f"Tokens file not found: {self.tokens_path}")

        with open(self.tokens_path, 'r') as f:
            return json.load(f)

    def save_tokens(self, tokens):
        """Save tokens to file."""
        tokens['refreshed_at'] = datetime.now().isoformat()

        with open(self.tokens_path, 'w') as f:
            json.dump(tokens, f, indent=2)

    def get_realm_id(self):
        """Get the Realm ID (Company ID)."""
        if not self.realm_id_path.exists():
            raise FileNotFoundError(
                f"Realm ID file not found: {self.realm_id_path}\n"
                "Please run oauth_setup.py to set up authentication."
            )

        with open(self.realm_id_path, 'r') as f:
            return f.read().strip()

    def is_token_valid(self, tokens):
        """Check if access token is still valid."""
        if 'refreshed_at' not in tokens or 'expires_in' not in tokens:
            return False

        refreshed_at = datetime.fromisoformat(tokens['refreshed_at'])
        expires_in = int(tokens['expires_in'])
        expiry_time = refreshed_at + timedelta(seconds=expires_in)

        # Consider token valid if it has more than 5 minutes remaining
        buffer = timedelta(minutes=5)
        return datetime.now() < (expiry_time - buffer)

    def refresh_token(self, refresh_token):
        """Refresh the access token."""
        credentials = self.load_credentials()

        # Create Basic Auth header
        auth_string = f"{credentials['client_id']}:{credentials['client_secret']}"
        auth_bytes = auth_string.encode('ascii')
        auth_b64 = base64.b64encode(auth_bytes).decode('ascii')

        # Build curl command
        curl_cmd = [
            'curl', '-s', '-X', 'POST',
            credentials['token_url'],
            '-H', 'Accept: application/json',
            '-H', 'Content-Type: application/x-www-form-urlencoded',
            '-H', f'Authorization: Basic {auth_b64}',
            '-d', f'grant_type=refresh_token',
            '-d', f'refresh_token={refresh_token}'
        ]

        try:
            result = subprocess.run(curl_cmd, capture_output=True, text=True, check=True)
            response = json.loads(result.stdout)

            if 'access_token' in response:
                return response
            else:
                raise Exception(f"Token refresh failed: {response}")

        except subprocess.CalledProcessError as e:
            raise Exception(f"Token refresh curl failed: {e.stderr}")
        except json.JSONDecodeError as e:
            raise Exception(f"Failed to parse token response: {e}")

    def get_valid_token(self):
        """Get a valid access token, refreshing if necessary."""
        try:
            tokens = self.load_tokens()
        except FileNotFoundError:
            raise Exception(
                "No tokens found. Please run oauth_setup.py to authenticate."
            )

        # Check if token is valid
        if self.is_token_valid(tokens):
            return tokens['access_token']

        # Token needs refresh
        refresh_token = tokens.get('refresh_token')
        if not refresh_token:
            raise Exception(
                "No refresh token found. Please run oauth_setup.py to re-authenticate."
            )

        # Refresh the token
        new_tokens = self.refresh_token(refresh_token)
        tokens.update(new_tokens)
        self.save_tokens(tokens)

        return tokens['access_token']

    def get_base_url(self):
        """Get the base API URL based on environment."""
        credentials = self.load_credentials()
        environment = credentials.get('environment', 'sandbox')
        return credentials['base_url'][environment]

# Convenience function for simple usage
def get_valid_token(config_dir=None):
    """Get a valid access token, refreshing if necessary."""
    manager = TokenManager(config_dir)
    return manager.get_valid_token()

def get_realm_id(config_dir=None):
    """Get the Realm ID."""
    manager = TokenManager(config_dir)
    return manager.get_realm_id()

def get_base_url(config_dir=None):
    """Get the base API URL."""
    manager = TokenManager(config_dir)
    return manager.get_base_url()
