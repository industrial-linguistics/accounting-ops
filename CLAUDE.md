# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Accounting Operations Toolkit is a cross-platform application suite for connecting to Deputy, Xero, and QuickBooks. The codebase is hybrid: C++/Qt desktop tools for end users, and Go services for OAuth broker/CLI.

**Two parallel technology stacks:**
1. **C++/Qt** - Desktop GUI tools and shared credential/skill libraries
2. **Go** - OAuth broker service (CGI) and CLI (`acct`) for authentication flows

## Build Commands

### C++/Qt Tools (CMake)

```bash
# Configure and build (Release)
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build

# Debug build
cmake -S . -B build -DCMAKE_BUILD_TYPE=Debug
cmake --build build

# Use Qt5 instead of Qt6
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release -DUSE_QT6=OFF
cmake --build build

# Install locally
cmake --install build --prefix "$HOME/.local"

# Create platform packages (DMG, PKG, TGZ, ZIP, HPKG)
cmake --build build --target package

# macOS-specific package with distribution file
cmake --build build --target package-macos
```

### Go Services

```bash
# Build the OAuth broker CGI
cd cmd/broker
go build -trimpath -ldflags="-s -w" -o broker

# Build the CLI
cd cmd/acct
go build -trimpath -ldflags="-s -w" -o acct

# Run tests
go test ./...
```

## Version Management

- The project version is stored in the `/VERSION` file at the project root
- This is the single source of truth for all version numbers
- Never hardcode version numbers in source code or CMake files
- The version automatically propagates to all binaries and packages via `configure_file()` in CMakeLists.txt

## Architecture

### Dual-Stack Design

The repository contains two complementary systems that work together:

**C++/Qt Desktop Tools** provide end-user GUI/CLI interfaces:
- `first_run_gui_tool` / `first_run_cli_tool` - Initial credential setup wizards
- `deputy_tool`, `xero_tool`, `quickbooks_tool` - Per-service verification tools
- `client_manager_tool` - Multi-client credential management
- `skill_editor_tool` - Edit skill definition files

**Go Authentication System** handles OAuth flows:
- `cmd/broker` - CGI binary running on OpenBSD/httpd for OAuth callbacks (requires HTTPS redirect URIs for QBO)
- `cmd/acct` - Customer-facing CLI that orchestrates OAuth, token refresh, and OS keychain storage

### Shared Libraries (C++)

**skills_core** (`skills/`)
- `CredentialStore`: Multi-client credential storage in SQLite (`config/credentials.sqlite`)
- `SkillRepository`: Loads skill definition JSON files from `skills/data/`
- Core data structures: `ClientProfile`, `ServiceCredential`

**tooling_common** (`tools/common/`)
- `CredentialSelector`: Qt widget for client selection
- `ConnectionTestWidget`: Service connectivity testing UI
- Reusable across all GUI tools

### OAuth Broker Architecture

The Go broker is **mandatory** for OAuth flows because:
- **QuickBooks Online** requires HTTPS redirect URIs (no localhost/IP in production) and returns `realmId` in callbacks
- **Deputy** OAuth requires client secret for token refresh and returns customer endpoint (subdomain)
- **Xero** supports Auth Code + PKCE; access tokens last 30 minutes; requires `xero-tenant-id` header from `/connections` API

**Broker Deployment:**
- Runs on OpenBSD under `httpd` + `slowcgi` in chroot `/var/www`
- Domain: `auth.industrial-linguistics.com`
- TLS termination handled by Cloudflare (origin server may use HTTP or HTTPS)
- Session data stored in SQLite; **never stores long-lived refresh tokens**
- Client secrets live in `conf/broker.env` only

**Key Endpoints:**
- `POST /v1/auth/start` - Initiate OAuth, returns auth URL and poll URL
- `GET /v1/auth/poll/{session}` - Client polls for completion, returns tokens once
- `POST /v1/token/refresh` - Refresh Deputy/QBO tokens (Xero refresh is local via PKCE)

**CLI (`acct`) Commands:**
```bash
acct connect xero|deputy|qbo --profile NAME   # Open browser, complete OAuth
acct list                                     # List profiles
acct whoami --profile NAME                    # Test API connectivity
acct refresh --profile NAME                   # Refresh tokens
acct revoke --profile NAME                    # Remove credentials
```

Token storage uses OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service).

See [docs/auth-broker-architecture.md](docs/auth-broker-architecture.md) for comprehensive OAuth broker design, deployment, and vendor-specific requirements.

### Project Structure

```
/
├── VERSION                    # Single source of truth for version
├── CMakeLists.txt             # Top-level build configuration
├── go.mod                     # Go module definition
├── skills/                    # Shared C++ credential/skill library
│   ├── include/skills/        # Public headers
│   ├── src/                   # Implementation
│   └── data/                  # Skill JSON definitions
├── tools/                     # Individual Qt applications
│   ├── common/                # tooling_common library (shared widgets)
│   ├── first_run_gui/         # Setup wizard (GUI)
│   ├── first_run_cli/         # Setup wizard (terminal)
│   ├── deputy_tool/           # Deputy credential testing
│   ├── xero_tool/             # Xero credential testing
│   ├── quickbooks_tool/       # QuickBooks credential testing
│   ├── client_manager_tool/   # Multi-client credential browser
│   └── skill_editor_tool/     # Skill file editor
├── cmd/                       # Go services
│   ├── broker/                # OAuth broker CGI
│   └── acct/                  # Customer CLI
└── docs/
    ├── BUILDING.md            # Build prerequisites and instructions
    ├── auth-broker-architecture.md  # OAuth broker design doc
    ├── man/man1/              # Man pages for each tool
    └── help/                  # GUI help markdown files
```

### Credential Flow

1. User runs `first_run_gui_tool` or `first_run_cli_tool` (C++ Qt app)
2. Tool captures client display name and launches Go CLI or browser flow
3. OAuth handled by `acct` CLI → broker → vendor → callback
4. Go CLI stores tokens in OS keychain
5. C++ tools read from SQLite database at `config/credentials.sqlite`
6. Each tool loads credentials via `CredentialStore` and displays verification UI

**Important:** C++ tools and Go CLI maintain separate credential stores. Qt tools use SQLite; Go CLI uses OS keychain. The first-run tools coordinate both.

### Skill Definitions

Skill JSON files in `skills/data/` define service metadata:
- `deputy.skill.json`
- `xero.skill.json`
- `quickbooks.skill.json`

Loaded by `SkillRepository`, editable via `skill_editor_tool`, and bundled into packages under `share/accounting-ops/skills`.

## Testing

C++ unit tests would use Qt Test framework (not currently implemented).
Go tests: `go test ./...` in `cmd/broker/` and `cmd/acct/`.

## Platform-Specific Notes

**macOS:**
- Install Qt via Homebrew: `brew install qt`
- Set `CMAKE_PREFIX_PATH` to Qt installation
- Removes deprecated AGL framework from WrapOpenGL (deprecated in 10.14)
- Package target creates `.pkg` installer via `productbuild`

**Windows:**
- Install Qt via Qt Online Installer
- Use Qt Command Prompt for builds
- Supports Visual Studio 2022 or MinGW
- Package target creates ZIP archive

**Linux:**
- Install `qt6-base-dev` or `qtbase5-dev`
- Package target creates `.tar.gz`

**Haiku:**
- Install `qt6` package: `pkgman install qt6`
- Set `CMAKE_PREFIX_PATH=/boot/system/lib/cmake`
- Package target creates HPKG via Haiku package generator

## Claude Code Skills

This repository includes three managed Claude Code skills for API integration:
- **deputy** - Deputy API for workforce management, timesheets, rosters
- **quickbooks** - QuickBooks Online API for accounting, invoices, bills
- **xero** - Xero API for accounting, payroll, invoices

Invoke with `/deputy`, `/quickbooks`, or `/xero` when working with those APIs.

## Security Considerations

- **No client secrets in CLI or Qt tools** - secrets only in `cmd/broker/conf/broker.env` on the server
- Always use state + PKCE where supported (Xero)
- Rotate refresh tokens on every refresh (Deputy, QBO, Xero)
- Request minimal OAuth scopes
- Enforce HTTPS for all broker transport
- Credentials in SQLite are stored locally; production deployments should consider encryption

## Vendor-Specific OAuth Requirements

**Xero:**
- Auth Code + PKCE for native apps
- Access tokens: 30 min; refresh tokens: 60 day inactivity expiry
- Every API call requires `xero-tenant-id` header (from `/connections`)
- Uncertified app limits: 25 tenant connections total, 2 uncertified apps per org
- App Store certification required for broad distribution

**Deputy:**
- OAuth via `once.deputy.com`
- Always request `longlife_refresh_token` scope
- Access tokens: ~24 hours
- Refresh requires client secret, rotates refresh token, returns endpoint domain
- Store endpoint (subdomain) per profile

**QuickBooks Online:**
- Production redirect URIs must be HTTPS (no localhost/IP)
- Callback returns `realmId` - required for API base paths
- Access tokens: ~1 hour; refresh tokens: up to 100 days rolling
- Refresh rotates tokens; always store newest value
- Scope: `com.intuit.quickbooks.accounting` (add OpenID scopes only when needed)

## Deployment

**C++ Tools:** Use CPack artifacts from `cmake --build build --target package`. Executables, man pages, help files, and skill data are bundled for redistribution.

**Go Broker:** Deployed to OpenBSD server via `scripts/build_deploy_openbsd.sh`:
1. Pull latest code
2. Build `cmd/broker` with `CGO_ENABLED=0`
3. Install to `/var/www/vhosts/auth.industrial-linguistics.com/cgi-bin/broker`
4. Ensure chroot DNS/CA files (`/var/www/etc/resolv.conf`, `/var/www/etc/ssl/cert.pem`)

**GitHub Actions:** `.github/workflows/deploy-*.yml` handles platform-specific packaging; broker deployment is SSH-triggered via `scripts/build_deploy_openbsd.sh`.

## Common Development Tasks

### Adding a New Qt Tool

1. Create directory under `tools/new_tool/`
2. Add `CMakeLists.txt` with executable target
3. Link `skills_core` and/or `tooling_common`
4. Add `add_subdirectory(new_tool)` to `tools/CMakeLists.txt`
5. Install target and create man page in `docs/man/man1/`

### Modifying OAuth Flows

Edit `cmd/broker/main.go` for broker endpoints or `cmd/acct/main.go` for CLI commands. Refresh token handling must preserve rotation behavior (store newest token). Refer to [docs/auth-broker-architecture.md](docs/auth-broker-architecture.md).

### Adding a New Service

1. Create skill JSON in `skills/data/new_service.skill.json`
2. Add provider to broker (`cmd/broker/`) with OAuth URLs and token endpoints
3. Add `acct connect new_service` command to CLI
4. Create Qt verification tool under `tools/new_service/`
5. Update `CredentialStore` schema if new fields required

## Git Workflow

- After doing any significant amount of work, git add, commit and push the work
