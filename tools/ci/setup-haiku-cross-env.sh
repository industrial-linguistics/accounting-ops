#!/usr/bin/env bash
# Setup environment for Haiku cross compilation.
set -euo pipefail

ARCH=${HAIKU_ARCH:-x86_64}
CROSS_DIR=${CROSS_DIR:-"$HOME/cross-tools-${ARCH}"}
CROSS_BIN=${CROSS_BIN:-"$CROSS_DIR/bin"}
SYSROOT=${SYSROOT:-"$CROSS_DIR/sysroot"}
HOSTTOOLS_DIR=${HOSTTOOLS_DIR:-"$HOME/haiku-hosttools"}

fetch_tools() {
    echo "Fetching Haiku cross compiler..."
    sudo apt-get update
    sudo apt-get install -y jq unzip qmake6 qt6-base-dev qt6-base-dev-tools
    if [ ! -d toolchain ]; then
        git clone --depth=1 https://github.com/haiku/haiku-toolchains-ubuntu.git toolchain
    fi
    pushd toolchain >/dev/null
    hosttools_url=$(./fetch.sh --hosttools)
    curl -sLJO "$hosttools_url"
    buildtools_url=$(./fetch.sh --buildtools --arch="$ARCH")
    curl -sLJO "$buildtools_url"
    unzip -qo ${ARCH}-linux-hosttools-*.zip -d "$HOSTTOOLS_DIR"
    unzip -qo ${ARCH}-linux-buildtools-*.zip -d "$HOME"
    popd >/dev/null
}

install_qt_packages() {
    QT_PKGS="qt6_base qt6_svg qt6_multimedia qt6_translations"
    BASE="https://eu.hpkg.haiku-os.org/haikuports/master/${ARCH}/current"

    if ! curl -sfI "$BASE/repo" >/dev/null; then
        echo "Unable to access HaikuPorts repository at $BASE" >&2
        exit 1
    fi

    mkdir -p "$SYSROOT/boot/system"
    curl -sSL "$BASE/repo" -o repo.hpkg
    if [ ! -s repo.hpkg ]; then
        echo "Failed to download repository index from $BASE" >&2
        exit 1
    fi
    package_repo list -f repo.hpkg | sed 's/^[[:space:]]*//' > repo.txt

    for p in $QT_PKGS; do
        FILE=$(grep -E "^${p}-.*-${ARCH}\\.hpkg$" repo.txt | sort -V | tail -1)
        if [ -z "$FILE" ]; then
            echo "Unable to determine package filename for $p" >&2
            exit 1
        fi
        if [ ! -f "$FILE" ]; then
            curl -sSL -o "$FILE" "$BASE/packages/$FILE"
        fi
        package extract -C "$SYSROOT/boot/system" "$FILE"
    done
}

main() {
    if [ ! -d "$CROSS_BIN" ] || [ ! -d "$HOSTTOOLS_DIR" ]; then
        fetch_tools
    fi
    export PATH="$CROSS_BIN:$HOSTTOOLS_DIR:$PATH"
    export LD_LIBRARY_PATH="$HOSTTOOLS_DIR:${LD_LIBRARY_PATH:-}"

    if [ ! -d "$SYSROOT/boot/system" ]; then
        install_qt_packages
    fi

    if [ -n "${GITHUB_ENV:-}" ]; then
        cat <<ENV_VARS >> "$GITHUB_ENV"
CROSS_DIR=$CROSS_DIR
CROSS_BIN=$CROSS_BIN
SYSROOT=$SYSROOT
HOSTTOOLS_DIR=$HOSTTOOLS_DIR
LD_LIBRARY_PATH=$LD_LIBRARY_PATH
ENV_VARS
    fi

    if [ -n "${GITHUB_PATH:-}" ]; then
        echo "$CROSS_BIN" >> "$GITHUB_PATH"
        echo "$HOSTTOOLS_DIR" >> "$GITHUB_PATH"
    fi
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    main
fi
