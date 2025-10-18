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
| `deputy_tool` | Verify Deputy credentials for a selected client. |
| `xero_tool` | Confirm Xero OAuth configuration before payroll or invoicing. |
| `quickbooks_tool` | Check QuickBooks credential readiness per client. |
| `client_manager_tool` | List and inspect multi-client credential sets. |
| `skill_editor_tool` | Browse and edit skill definition files. |

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
Credentials are managed centrally in `config/credentials.json`. The data model
supports multiple clients, each of which can hold service-specific credentials
for Deputy, Xero, QuickBooks, or other integrations. Use the Client Manager GUI
to review the current inventory.

## Documentation
* Man pages: `docs/man/man1/*.1`
* GUI help files: `docs/help/*.md`

These resources are installed automatically and included in packaged builds to
assist end users.
