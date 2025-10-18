#!/usr/bin/env python3
"""
Test Deputy API Authentication

This script tests your Deputy API credentials.

Usage:
    python test_auth.py
"""

import sys
from pathlib import Path
from api_wrapper import DeputyAPI

def main():
    """Test authentication."""

    print("=" * 70)
    print("DEPUTY API AUTHENTICATION TEST")
    print("=" * 70)

    try:
        # Initialize API
        print("\n1. Loading credentials...")
        api = DeputyAPI()
        api.load_credentials()
        print(f"   Domain: {api.domain}")
        print(f"   API Key: {api.api_key[:10]}...")

        # Test API call
        print("\n2. Testing API access...")
        response = api.get_me()

        print("\n" + "=" * 70)
        print("SUCCESS!")
        print("=" * 70)

        if 'Id' in response:
            print(f"\nUser ID: {response.get('Id')}")
            print(f"Name: {response.get('DisplayName', 'N/A')}")
            print(f"Email: {response.get('Email', 'N/A')}")

        print("\nYour Deputy API credentials are working correctly!")

    except FileNotFoundError as e:
        print(f"\nError: {e}")
        print("\nPlease create config/deputy_credentials.json with your API credentials.")
        sys.exit(1)
    except Exception as e:
        print(f"\nError: {e}")
        print("\nAuthentication failed. Please check your API credentials.")
        sys.exit(1)

if __name__ == '__main__':
    main()
