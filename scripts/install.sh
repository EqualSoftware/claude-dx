#!/bin/sh
# Install claude-dx from GitHub Releases on macOS and Linux.
# Usage: install.sh [version]     Example: install.sh v0.1.0
# Environment:
#   CLAUDE_DX_INSTALL_DIR  Install directory. Default: ~/.local/bin
#   CLAUDE_DX_BASE_URL     Release download URL, for tests. Default: GitHub Releases

set -eu

REPO="EqualSoftware/claude-dx"
INSTALL_DIR="${CLAUDE_DX_INSTALL_DIR:-$HOME/.local/bin}"

fail() {
    echo "claude-dx install: $*" >&2
    exit 1
}

detect_os() {
    case "$(uname -s)" in
        Darwin) echo darwin ;;
        Linux) echo linux ;;
        *) fail "this OS is not supported: $(uname -s). On Windows, use install.ps1." ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64 | amd64) echo amd64 ;;
        arm64 | aarch64) echo arm64 ;;
        *) fail "this architecture is not supported: $(uname -m)" ;;
    esac
}

download() {
    if command -v curl > /dev/null 2>&1; then
        curl -fsSL "$1" -o "$2"
    elif command -v wget > /dev/null 2>&1; then
        wget -q "$1" -O "$2"
    else
        fail "curl or wget is necessary"
    fi
}

sha256() {
    if command -v sha256sum > /dev/null 2>&1; then
        sha256sum "$1" | awk '{print $1}'
    elif command -v shasum > /dev/null 2>&1; then
        shasum -a 256 "$1" | awk '{print $1}'
    else
        fail "sha256sum or shasum is necessary"
    fi
}

latest_version() {
    download "https://api.github.com/repos/$REPO/releases/latest" "$1/latest.json"
    sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' "$1/latest.json" | head -n 1
}

main() {
    os="$(detect_os)"
    arch="$(detect_arch)"
    tmp="$(mktemp -d)"
    trap 'rm -rf "$tmp"' EXIT

    version="${1:-}"
    if [ -z "$version" ]; then
        version="$(latest_version "$tmp")"
        [ -n "$version" ] || fail "cannot find the latest release"
    fi

    archive="claude-dx_${version#v}_${os}_${arch}.tar.gz"
    base="${CLAUDE_DX_BASE_URL:-https://github.com/$REPO/releases/download/$version}"

    download "$base/$archive" "$tmp/$archive" || fail "cannot download $archive"
    download "$base/checksums.txt" "$tmp/checksums.txt" || fail "cannot download checksums.txt"

    expected="$(awk -v f="$archive" '$2 == f {print $1}' "$tmp/checksums.txt")"
    [ -n "$expected" ] || fail "checksums.txt has no entry for $archive"
    actual="$(sha256 "$tmp/$archive")"
    [ "$expected" = "$actual" ] || fail "checksum mismatch for $archive. Nothing is installed."

    tar -xzf "$tmp/$archive" -C "$tmp" claude-dx
    mkdir -p "$INSTALL_DIR"
    cp "$tmp/claude-dx" "$INSTALL_DIR/.claude-dx.new"
    chmod 0755 "$INSTALL_DIR/.claude-dx.new"
    mv -f "$INSTALL_DIR/.claude-dx.new" "$INSTALL_DIR/claude-dx"

    echo "claude-dx $version installed to $INSTALL_DIR/claude-dx"
    case ":$PATH:" in
        *":$INSTALL_DIR:"*) ;;
        *) echo "Add $INSTALL_DIR to your PATH." ;;
    esac
}

main "$@"
