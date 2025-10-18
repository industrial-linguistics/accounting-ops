#!/usr/bin/env python3
"""
Deputy API Wrapper

This script provides a wrapper around the Deputy API using curl.

Usage:
    # Get current user
    python api_wrapper.py --endpoint me

    # Get employees
    python api_wrapper.py --endpoint resource/Employee

    # Query with search
    python api_wrapper.py --endpoint resource/Timesheet/QUERY --data query.json
"""

import argparse
import json
import sys
import subprocess
import logging
from pathlib import Path
from datetime import datetime

# Set up logging
log_dir = Path(__file__).parent.parent.parent.parent / 'logs'
log_dir.mkdir(exist_ok=True)
log_file = log_dir / 'deputy_api.log'

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(log_file),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class DeputyAPI:
    """Wrapper for Deputy API calls."""

    def __init__(self):
        """Initialize API wrapper."""
        self.api_key = None
        self.domain = None
        self.base_url = None

    def load_credentials(self):
        """Load Deputy credentials."""
        config_path = Path(__file__).parent.parent.parent.parent / 'config' / 'deputy_credentials.json'

        if not config_path.exists():
            raise FileNotFoundError(f"Credentials file not found: {config_path}")

        with open(config_path, 'r') as f:
            creds = json.load(f)

        self.api_key = creds.get('api_key')
        self.domain = creds.get('domain')
        self.base_url = creds.get('base_url', f'https://{self.domain}/api/v1')

        if not self.api_key or not self.domain:
            raise ValueError("Missing api_key or domain in credentials file")

        logger.info(f"Initialized with domain: {self.domain}")

    def make_request(self, method, endpoint, data=None):
        """
        Make an API request using curl.

        Args:
            method: HTTP method (GET, POST, PUT, DELETE)
            endpoint: API endpoint
            data: Request body (dict or JSON string)

        Returns:
            Response dict
        """
        # Build URL
        url = f"{self.base_url}/{endpoint}"

        # Build curl command
        curl_cmd = [
            'curl', '-s', '-X', method,
            url,
            '-H', 'Accept: application/json',
            '-H', f'Authorization: Bearer {self.api_key}',
            '-H', 'Content-Type: application/json'
        ]

        # Add data if present
        if data:
            if isinstance(data, dict):
                data = json.dumps(data)
            curl_cmd.extend(['-d', data])

        # Log the request
        logger.info(f"API Request: {method} {endpoint}")

        try:
            # Execute curl
            result = subprocess.run(curl_cmd, capture_output=True, text=True, check=True)

            # Parse response
            response = json.loads(result.stdout)

            logger.info(f"  Response: Success")

            return response

        except subprocess.CalledProcessError as e:
            logger.error(f"Curl failed: {e.stderr}")
            raise Exception(f"API request failed: {e.stderr}")
        except json.JSONDecodeError as e:
            logger.error(f"Failed to parse response: {result.stdout}")
            raise Exception(f"Invalid JSON response: {e}")

    def get_me(self):
        """Get current user information."""
        return self.make_request('GET', 'me')

    def get_employees(self):
        """Get all employees."""
        return self.make_request('GET', 'resource/Employee')

    def get_timesheets(self, query):
        """
        Query timesheets.

        Args:
            query: Search query dict

        Returns:
            Timesheets response
        """
        return self.make_request('POST', 'resource/Timesheet/QUERY', data=query)

def main():
    """Main CLI interface."""
    parser = argparse.ArgumentParser(description='Deputy API Wrapper')

    parser.add_argument('--endpoint', required=True,
                        help='API endpoint (e.g., me, resource/Employee)')
    parser.add_argument('--method', default='GET',
                        choices=['GET', 'POST', 'PUT', 'DELETE'],
                        help='HTTP method')
    parser.add_argument('--data', help='Path to JSON file with request data or JSON string')
    parser.add_argument('--output', help='Output file for response')

    args = parser.parse_args()

    # Initialize API
    api = DeputyAPI()
    api.load_credentials()

    try:
        # Load data if provided
        data = None
        if args.data:
            # Try to parse as JSON first
            try:
                data = json.loads(args.data)
            except json.JSONDecodeError:
                # If that fails, treat as file path
                with open(args.data, 'r') as f:
                    data = json.load(f)

        # Make request
        response = api.make_request(args.method, args.endpoint, data=data)

        # Output response
        if args.output:
            with open(args.output, 'w') as f:
                json.dump(response, f, indent=2)
            print(f"Response saved to {args.output}")
        else:
            print(json.dumps(response, indent=2))

    except Exception as e:
        logger.error(f"Operation failed: {e}")
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    main()
