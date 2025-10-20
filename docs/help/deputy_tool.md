# Deputy Connection Tool

The Deputy connection tool helps confirm that stored credentials for a
selected client are present before syncing workforce data.

## Quick start
1. Run one of the first-run assistants if `config/credentials.sqlite` has not
   been created yet.
2. Launch the application.
3. Provide the path to the credential database when prompted, or accept the
   default (`../config/credentials.sqlite`).
4. Enter the client name and click **Test Connection**.
5. Review the status message and success dialog.

## Expected results
A success dialog indicates that the credential database contains Deputy entries
for the selected client. A warning appears if the client is missing or lacks
Deputy access.
