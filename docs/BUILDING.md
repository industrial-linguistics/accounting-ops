# Building and Packaging Accounting Ops Tools

## Prerequisites
* CMake 3.20 or newer
* A C++20 compiler toolchain
* Qt 6 with Widgets (or Qt 5.15+ when `-DUSE_QT6=OFF`)

Platform-specific notes:
* **Linux**: install the `qtbase5-dev` or `qt6-base-dev` package depending on
the Qt version you wish to use.
* **macOS**: install Qt using Homebrew (`brew install qt`) or the official
installer. Ensure `CMAKE_PREFIX_PATH` points to the Qt installation.
* **Windows**: install Qt via the Qt Online Installer and run builds from the
"Qt Command Prompt" so that environment variables are configured. Visual
Studio 2022 or MinGW builds are supported.
* **Haiku**: install the `qt6` package (`pkgman install qt6`) and set
`CMAKE_PREFIX_PATH=/boot/system/lib/cmake`.

## Configure and build
```bash
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build
```

Each tool is produced as an individual executable:
* `deputy_tool`
* `xero_tool`
* `quickbooks_tool`
* `client_manager_tool`
* `skill_editor_tool`

The `skills_core` and `tooling_common` libraries are linked into the relevant
tools but can also be consumed by other Qt applications.

## Packaging
CPack is configured to emit platform-appropriate artifacts:
* Linux: `accounting-ops-<version>.tar.gz`
* macOS: drag-and-drop disk image (`.dmg`)
* Windows: ZIP archive for sideloading or publishing to the Microsoft Store
after signing
* Haiku: HPKG package via the Haiku package generator

Run the following command after a successful build:
```bash
cmake --build build --target package
```

The generated package contains binaries in `bin/`, man pages under
`share/man/man1`, and help files under `share/accounting-ops/help` ready for
redistribution.

## Installing locally
```bash
cmake --install build --prefix "$HOME/.local"
```

Executables will be placed under `$HOME/.local/bin`. Update your `PATH`
variable accordingly.
