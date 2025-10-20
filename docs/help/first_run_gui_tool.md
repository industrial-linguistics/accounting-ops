# First-Run GUI Tool

The first-run GUI walks operators through the credential collection process for
a new client.

## Steps
1. Launch the application. It automatically targets
   `config/credentials.sqlite` next to the executables.
2. Provide a display name for the client.
3. For each service (QuickBooks, Xero, Deputy), choose whether to configure it
   and enter the requested fields.
4. Use the **Test** button on each page to confirm the entries look correct.
5. Finish the wizard to save the credentials.

You can rerun the wizard at any time to add another client or update an existing
record.
