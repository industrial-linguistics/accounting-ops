#!/usr/bin/env python3
"""
Xero API Wrapper

This script provides a wrapper around the Xero API using curl.
It handles authentication, token refresh, and common API operations.

Usage:
    # Get invoices
    python api_wrapper.py --endpoint invoices

    # Get invoices with filter
    python api_wrapper.py --endpoint invoices --where "Status == \"AUTHORISED\""

    # Get payroll data
    python api_wrapper.py --endpoint payrollau/employees --api-type payroll
"""

import argparse
import json
import sys
import subprocess
import logging
from pathlib import Path
from datetime import datetime
from token_manager import XeroTokenManager

# Set up logging
log_dir = Path(__file__).parent.parent.parent.parent / 'logs'
log_dir.mkdir(exist_ok=True)
log_file = log_dir / 'xero_api.log'

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(log_file),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class XeroAPI:
    """Wrapper for Xero API calls."""

    def __init__(self):
        """Initialize API wrapper."""
        self.token_manager = XeroTokenManager()
        self.tenant_id = None
        self.base_url = None
        self.access_token = None

    def initialize(self, api_type='accounting'):
        """
        Initialize authentication and configuration.

        Args:
            api_type: 'accounting' or 'payroll'
        """
        try:
            self.tenant_id = self.token_manager.get_tenant_id()
            self.access_token = self.token_manager.get_valid_token()

            if api_type == 'accounting':
                self.base_url = self.token_manager.get_base_url()
            else:
                # Assume AU for payroll, can be overridden
                self.base_url = self.token_manager.get_payroll_base_url('AU')

            logger.info(f"Initialized with Tenant ID: {self.tenant_id}")
            logger.info(f"Using base URL: {self.base_url}")
        except Exception as e:
            logger.error(f"Initialization failed: {e}")
            raise

    def make_request(self, method, endpoint, data=None, params=None):
        """
        Make an API request using curl.

        Args:
            method: HTTP method (GET, POST, PUT, DELETE)
            endpoint: API endpoint
            data: Request body (dict or JSON string)
            params: Query parameters (dict)

        Returns:
            Response dict
        """
        # Build URL
        url = f"{self.base_url}/{endpoint}"

        if params:
            param_str = '&'.join([f"{k}={v}" for k, v in params.items()])
            url = f"{url}?{param_str}"

        # Build curl command
        curl_cmd = [
            'curl', '-s', '-X', method,
            url,
            '-H', 'Accept: application/json',
            '-H', f'Authorization: Bearer {self.access_token}',
            '-H', f'xero-tenant-id: {self.tenant_id}',
            '-H', 'Content-Type: application/json'
        ]

        # Add data if present
        if data:
            if isinstance(data, dict):
                data = json.dumps(data)
            curl_cmd.extend(['-d', data])

        # Log the request
        logger.info(f"API Request: {method} {endpoint}")
        if params:
            logger.info(f"  Parameters: {params}")

        try:
            # Execute curl
            result = subprocess.run(curl_cmd, capture_output=True, text=True, check=True)

            # Parse response
            response = json.loads(result.stdout)

            # Log response
            logger.info(f"  Response: Success")

            return response

        except subprocess.CalledProcessError as e:
            logger.error(f"Curl failed: {e.stderr}")
            raise Exception(f"API request failed: {e.stderr}")
        except json.JSONDecodeError as e:
            logger.error(f"Failed to parse response: {result.stdout}")
            raise Exception(f"Invalid JSON response: {e}")

    def get_organisations(self):
        """Get organisation information."""
        return self.make_request('GET', 'Organisations')

    def get_invoices(self, where=None, order=None):
        """
        Get invoices.

        Args:
            where: Filter expression (e.g., "Status == \"AUTHORISED\"")
            order: Order by expression

        Returns:
            Invoices response
        """
        params = {}
        if where:
            params['where'] = where
        if order:
            params['order'] = order

        return self.make_request('GET', 'Invoices', params=params)

    def get_bank_transactions(self, where=None):
        """Get bank transactions."""
        params = {}
        if where:
            params['where'] = where

        return self.make_request('GET', 'BankTransactions', params=params)

    def get_contacts(self, where=None):
        """Get contacts."""
        params = {}
        if where:
            params['where'] = where

        return self.make_request('GET', 'Contacts', params=params)

def main():
    """Main CLI interface."""
    parser = argparse.ArgumentParser(description='Xero API Wrapper')

    parser.add_argument('--endpoint', required=True,
                        help='API endpoint (e.g., invoices, organisations)')
    parser.add_argument('--method', default='GET',
                        choices=['GET', 'POST', 'PUT', 'DELETE'],
                        help='HTTP method')
    parser.add_argument('--where', help='Filter expression')
    parser.add_argument('--order', help='Order by expression')
    parser.add_argument('--data', help='Path to JSON file with request data')
    parser.add_argument('--output', help='Output file for response')
    parser.add_argument('--api-type', default='accounting',
                        choices=['accounting', 'payroll'],
                        help='API type')
    parser.add_argument('--tenant-id', help='Override tenant ID')

    args = parser.parse_args()

    # Initialize API
    api = XeroAPI()

    # Override tenant ID if provided
    if args.tenant_id:
        api.tenant_id = args.tenant_id
        api.base_url = api.token_manager.get_base_url()
        api.access_token = api.token_manager.get_valid_token()
    else:
        api.initialize(args.api_type)

    try:
        # Load data if provided
        data = None
        if args.data:
            with open(args.data, 'r') as f:
                data = json.load(f)

        # Build params
        params = {}
        if args.where:
            params['where'] = args.where
        if args.order:
            params['order'] = args.order

        # Make request
        response = api.make_request(args.method, args.endpoint, data=data, params=params)

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
