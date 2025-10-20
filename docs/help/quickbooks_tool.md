# QuickBooks Connection Tool

Validate that QuickBooks Online credentials are configured for each client.

## Usage
1. Ensure the shared credential database (`config/credentials.sqlite`) exists by
   running one of the first-run assistants if necessary.
2. Start the tool and open the credential database when prompted.
3. Provide the client name that maps to QuickBooks.
4. Click **Test Connection**. A success dialog confirms simulated endpoint
   checks.

## Notes
The tool surfaces missing or mislabelled credentials to help non-technical
users correct setup issues before automation runs.
