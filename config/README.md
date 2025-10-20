# Configuration Directory

This directory holds local-only configuration and credential data for the
Accounting Ops toolkit. Everything stored here is ignored by git so that real
customer secrets never leave the operator's workstation.

## Security Notice

**IMPORTANT**: Do not commit any credential database produced by the tools.
The `.gitignore` at the repository root prevents accidental check-ins, but
always double check before sharing archives.

## Credential Database

All utilities share a single SQLite database named `credentials.sqlite`. The
file is created automatically the first time you run either of the
first-run assistants:

* `first_run_gui_tool` – a guided Qt wizard that collects credentials for
  Deputy, QuickBooks, and Xero.
* `first_run_cli_tool` – a terminal-based workflow suited to headless
  environments or remote shells.

The database stores each client's display name plus their service credentials.
You can re-run either assistant at any time to add new clients or update
existing records.

## Manual editing

Directly editing the SQLite file is not recommended. Use the provided tools to
modify credentials:

1. Launch `client_manager_tool` to review the currently stored clients and
   verify which services are configured.
2. Use the service-specific diagnostic tools (`deputy_tool`, `xero_tool`,
   `quickbooks_tool`) to validate connectivity after making changes.

If you must inspect the database manually (for backup or migration purposes),
a standard SQLite browser will open the file, but avoid editing fields while
any Accounting Ops tool is running.
