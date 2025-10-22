# Agent Guidelines for Accounting Ops Toolkit

This document provides guidelines for AI coding assistants and agents working on the Accounting Ops Toolkit codebase.

## Version Management

**Location**: `/VERSION`

The project uses a single source of truth for version numbers:

- **VERSION file**: Contains the current version in `MAJOR.MINOR.PATCH` format (e.g., `1.0.0`)
- This file is read by CMake during build configuration
- The version is automatically embedded into:
  - All CLI and GUI application binaries
  - Generated `.pkg` and other installation packages
  - macOS `.app` bundle Info.plist files

**How to update the version:**
1. Edit the `VERSION` file in the project root
2. Rebuild the project - CMake will automatically pick up the new version

**DO NOT** hardcode version numbers in:
- CMakeLists.txt
- Source code files
- Package configuration files

The version is available in C++ code via:
```cpp
#include "version.h"
ACCOUNTING_OPS_VERSION_STRING  // e.g., "1.0.0"
ACCOUNTING_OPS_VERSION_MAJOR   // e.g., 1
ACCOUNTING_OPS_VERSION_MINOR   // e.g., 0
ACCOUNTING_OPS_VERSION_PATCH   // e.g., 0
```

## Project Structure

### Tools
All tools are located in `tools/` and consist of:
- **GUI Tools**: Built as macOS `.app` bundles (deputy, xero, quickbooks, client_manager, skill_editor, first_run_gui)
- **CLI Tools**: Built as command-line executables (first_run_cli)

### Shared Libraries
- **skills_core**: Credential storage and skill repository
- **tooling_common**: Reusable Qt widgets

## Build System

### macOS Packaging
- Use `cmake --build build --target package-macos` to create the `.pkg` installer
- GUI apps install to `/Applications`
- CLI tools install to `/usr/local/bin`
- Resources install to `/usr/local/share/accounting-ops`

### Cross-platform Notes
- The project supports Linux, macOS, Windows, and Haiku
- Use platform-specific CMake conditionals when needed
- Test packaging on each target platform

## Code Guidelines

### Version Display
All tools should display version information via:
```cpp
QCoreApplication::setApplicationVersion(ACCOUNTING_OPS_VERSION_STRING);
```

This ensures `--version` flags display the correct version.

## Documentation

When making significant changes, update:
- `README.md` - User-facing documentation
- `docs/BUILDING.md` - Build instructions
- `AGENTS.md` - This file (agent guidelines)
- `CLAUDE.md` - Claude-specific instructions
