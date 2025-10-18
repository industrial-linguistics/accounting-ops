# Deputy Connection Tool

The Deputy connection tool helps confirm that stored credentials for a
selected client are present before syncing workforce data.

## Quick start
1. Launch the application.
2. Provide the path to `credentials.json` when prompted, or accept the
   default (`../config/credentials.json`).
3. Enter the client name and click **Test Connection**.
4. Review the status message and success dialog.

## Expected results
A success dialog indicates that the credential structure contains entries for
Deputy. A warning appears if the client is missing or lacks Deputy access.
