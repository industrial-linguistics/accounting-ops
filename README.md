# Accounting Operations Toolkit

Cross-platform C++/Qt applications and shared skill libraries for connecting
to Deputy, Xero, and QuickBooks. The toolkit is designed to be distributed as
individual desktop utilities suitable for non-technical accounting staff while
retaining a common credential and skill management backend.

## Components

### Shared Libraries
* **skills_core** – credential storage with multi-client support and skill file
  repository helpers.
* **tooling_common** – reusable Qt widgets for selecting clients and testing
  service connectivity.

### GUI Tools
Each tool builds as its own executable and can be packaged independently for
Linux, Haiku, macOS, and Windows via CPack.

| Tool | Purpose |
| --- | --- |
| `first_run_gui_tool` | Guided wizard that collects credentials and writes the shared database. |
| `first_run_cli_tool` | Terminal-based first-run flow for headless environments. |
| `deputy_tool` | Verify Deputy credentials for a selected client. |
| `xero_tool` | Confirm Xero OAuth configuration before payroll or invoicing. |
| `quickbooks_tool` | Check QuickBooks credential readiness per client. |
| `client_manager_tool` | List and inspect multi-client credential sets. |
| `skill_editor_tool` | Browse and edit skill definition files. |

### OAuth Broker and Go CLI
The repository also provides a Go-based authentication broker (`cmd/broker`) and
customer-facing CLI (`cmd/acct`). The broker mediates OAuth flows for Deputy,
Xero, and QuickBooks Online using a one-time polling channel, while the CLI
orchestrates account connections, token refresh, and credential storage via the
OS keychain. Refer to [docs/auth-broker-architecture.md](docs/auth-broker-architecture.md)
for detailed deployment and runtime guidance.

### Skills
Skill definitions are stored as JSON under `skills/data` and can be loaded in
the Qt skill editor or bundled into deployments.

## Building
Refer to [docs/BUILDING.md](docs/BUILDING.md) for full prerequisites and
commands. A typical release build is:

```bash
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build
```

## Packaging
`cmake --build build --target package` generates platform-specific archives or
installers that include executables, shared documentation, man pages, and help
files suitable for redistribution and app store submission (after code signing
on macOS/Windows).

## Credentials
Credentials are managed centrally in a SQLite database stored at
`config/credentials.sqlite`. The file is created automatically the first time
you run either `first_run_gui_tool` or `first_run_cli_tool`. Each entry stores a
client display name alongside per-service credentials for Deputy, Xero,
QuickBooks, or future integrations.

Use the Client Manager GUI to review the current inventory and rerun the
first-run tools whenever you need to add or update credentials.

## Documentation
* Man pages: `docs/man/man1/*.1`
* GUI help files: `docs/help/*.md`
* Authentication broker design: `docs/auth-broker-architecture.md`

These resources are installed automatically and included in packaged builds to
assist end users.
