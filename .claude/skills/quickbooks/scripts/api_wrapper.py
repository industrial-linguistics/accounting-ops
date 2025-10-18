#!/usr/bin/env python3
"""
Quickbooks API Wrapper

This script provides a wrapper around the Quickbooks API using curl.
It handles authentication, token refresh, and common API operations.

Usage:
    # Query data
    python api_wrapper.py --endpoint query --query "SELECT * FROM Account"

    # Get company info
    python api_wrapper.py --endpoint companyinfo

    # Get specific entity
    python api_wrapper.py --endpoint invoice --id 123

    # Create or update (provide JSON file)
    python api_wrapper.py --endpoint invoice --method POST --data invoice.json
"""

import argparse
import json
import sys
import subprocess
import logging
from pathlib import Path
from datetime import datetime
from token_manager import TokenManager

# Set up logging
log_dir = Path(__file__).parent.parent.parent.parent / 'logs'
log_dir.mkdir(exist_ok=True)
log_file = log_dir / 'quickbooks_api.log'

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(log_file),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class QuickbooksAPI:
    """Wrapper for Quickbooks API calls."""

    def __init__(self):
        """Initialize API wrapper."""
        self.token_manager = TokenManager()
        self.realm_id = None
        self.base_url = None
        self.access_token = None

    def initialize(self):
        """Initialize authentication and configuration."""
        try:
            self.realm_id = self.token_manager.get_realm_id()
            self.base_url = self.token_manager.get_base_url()
            self.access_token = self.token_manager.get_valid_token()
            logger.info(f"Initialized with Realm ID: {self.realm_id}")
            logger.info(f"Using base URL: {self.base_url}")
        except Exception as e:
            logger.error(f"Initialization failed: {e}")
            raise

    def make_request(self, method, endpoint, data=None, params=None):
        """
        Make an API request using curl.

        Args:
            method: HTTP method (GET, POST, PUT, DELETE)
            endpoint: API endpoint (e.g., 'invoice', 'query')
            data: Request body (dict or JSON string)
            params: Query parameters (dict)

        Returns:
            Response dict
        """
        # Build URL
        url = f"{self.base_url}/v3/company/{self.realm_id}/{endpoint}"

        if params:
            param_str = '&'.join([f"{k}={v}" for k, v in params.items()])
            url = f"{url}?{param_str}"

        # Build curl command
        curl_cmd = [
            'curl', '-s', '-X', method,
            url,
            '-H', 'Accept: application/json',
            '-H', f'Authorization: Bearer {self.access_token}',
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
            if 'QueryResponse' in response:
                count = len(response['QueryResponse'].get(list(response['QueryResponse'].keys())[0], []))
                logger.info(f"  Response: Query returned {count} results")
            elif 'fault' in response:
                logger.error(f"  API Error: {response['fault']}")
            else:
                logger.info(f"  Response: Success")

            return response

        except subprocess.CalledProcessError as e:
            logger.error(f"Curl failed: {e.stderr}")
            raise Exception(f"API request failed: {e.stderr}")
        except json.JSONDecodeError as e:
            logger.error(f"Failed to parse response: {result.stdout}")
            raise Exception(f"Invalid JSON response: {e}")

    def query(self, sql):
        """
        Execute a SQL-like query.

        Args:
            sql: Query string (e.g., "SELECT * FROM Account")

        Returns:
            Query results
        """
        logger.info(f"Executing query: {sql}")
        params = {'query': sql}
        return self.make_request('GET', 'query', params=params)

    def get_company_info(self):
        """Get company information."""
        return self.make_request('GET', f'companyinfo/{self.realm_id}')

    def get_entity(self, entity_type, entity_id):
        """
        Get a specific entity by ID.

        Args:
            entity_type: Type of entity (e.g., 'invoice', 'customer')
            entity_id: Entity ID

        Returns:
            Entity data
        """
        return self.make_request('GET', f'{entity_type}/{entity_id}')

    def create_entity(self, entity_type, data):
        """
        Create a new entity.

        Args:
            entity_type: Type of entity (e.g., 'invoice', 'customer')
            data: Entity data (dict)

        Returns:
            Created entity data
        """
        return self.make_request('POST', entity_type, data=data)

    def update_entity(self, entity_type, data):
        """
        Update an existing entity.

        Args:
            entity_type: Type of entity
            data: Entity data including Id and SyncToken

        Returns:
            Updated entity data
        """
        return self.make_request('POST', entity_type, data=data)

    def delete_entity(self, entity_type, entity_id, sync_token):
        """
        Delete an entity (soft delete).

        Args:
            entity_type: Type of entity
            entity_id: Entity ID
            sync_token: Current sync token

        Returns:
            Deleted entity data
        """
        params = {'operation': 'delete'}
        data = {'Id': entity_id, 'SyncToken': sync_token}
        return self.make_request('POST', entity_type, data=data, params=params)

def main():
    """Main CLI interface."""
    parser = argparse.ArgumentParser(description='Quickbooks API Wrapper')

    parser.add_argument('--endpoint', required=True,
                        help='API endpoint (e.g., query, companyinfo, invoice)')
    parser.add_argument('--method', default='GET',
                        choices=['GET', 'POST', 'PUT', 'DELETE'],
                        help='HTTP method')
    parser.add_argument('--query', help='SQL query (for query endpoint)')
    parser.add_argument('--id', help='Entity ID (for specific entity queries)')
    parser.add_argument('--data', help='Path to JSON file with request data')
    parser.add_argument('--output', help='Output file for response')
    parser.add_argument('--realm-id', help='Override realm ID')

    args = parser.parse_args()

    # Initialize API
    api = QuickbooksAPI()

    # Override realm ID if provided
    if args.realm_id:
        api.realm_id = args.realm_id
        api.base_url = api.token_manager.get_base_url()
        api.access_token = api.token_manager.get_valid_token()
    else:
        api.initialize()

    try:
        # Handle different endpoint types
        if args.endpoint == 'query':
            if not args.query:
                print("Error: --query required for query endpoint")
                sys.exit(1)
            response = api.query(args.query)

        elif args.endpoint == 'companyinfo':
            response = api.get_company_info()

        elif args.id:
            response = api.get_entity(args.endpoint, args.id)

        elif args.data:
            # Load data from file
            with open(args.data, 'r') as f:
                data = json.load(f)
            response = api.make_request(args.method, args.endpoint, data=data)

        else:
            # Generic endpoint call
            response = api.make_request(args.method, args.endpoint)

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
