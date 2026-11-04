#!/usr/bin/env bash
set -euo pipefail

REPO="Senasphy/nm-go"
BINARY="nm-go"

info() {
    printf "[INFO] %s\n" "$*"
}

error() {
    printf "[ERROR] %s\n" "$*" >&2
}

# 1. Verify operating system
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
if [ "$OS" != "linux" ]; then
    error "nm-go is designed specifically for Linux (iwd service layer)."
    exit 1
fi

# 2. Detect CPU architecture
ARCH_RAW="$(uname -m)"
case "$ARCH_RAW" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    armv7l|armv6l|arm)
        ARCH="arm"
        ;;
    *)
        error "Unsupported system architecture: $ARCH_RAW"
        exit 1
        ;;
esac

# 3. Determine target install directory
if [ -n "${INSTALL_DIR:-}" ]; then
    DEST_DIR="$INSTALL_DIR"
elif [ "$(id -u)" -eq 0 ] || [ -w "/usr/local/bin" ]; then
    DEST_DIR="/usr/local/bin"
else
    DEST_DIR="${HOME}/.local/bin"
fi

mkdir -p "$DEST_DIR"

# 4. Determine release version
resolve_version() {
    if [ -n "${VERSION:-}" ]; then
        echo "$VERSION"
        return
    fi

    # Attempt resolution via GitHub CLI if installed
    if command -v gh >/dev/null 2>&1; then
        local gh_tag
        gh_tag=$(env -u GITHUB_TOKEN gh release view -R "$REPO" --json tagName -q .tagName 2>/dev/null || true)
        if [ -n "$gh_tag" ]; then
            echo "$gh_tag"
            return
        fi
    fi

    # Attempt resolution via GitHub Releases API
    local auth_header=()
    if [ -n "${GITHUB_TOKEN:-}" ]; then
        auth_header=(-H "Authorization: Bearer $GITHUB_TOKEN")
    fi

    local tag
    tag=$(curl -sSL "${auth_header[@]}" "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)
    if [ -n "$tag" ]; then
        echo "$tag"
        return
    fi

    # Fallback to redirect header on GitHub releases URL
    tag=$(curl -sI "https://github.com/${REPO}/releases/latest" 2>/dev/null | grep -i "^location:" | sed -E 's/.*tag\/(.*)/\1/' | tr -d '\r\n' || true)
    if [ -n "$tag" ]; then
        echo "$tag"
        return
    fi

    # Static fallback
    echo "v0.2.0"
}

TAG="$(resolve_version)"
# Ensure tag has leading 'v'
if [[ "$TAG" != v* ]]; then
    TAG="v$TAG"
fi
VER_NUM="${TAG#v}"

ARCHIVE="${BINARY}_${VER_NUM}_linux_${ARCH}.tar.gz"
CHECKSUMS="checksums.txt"

info "Selected nm-go release ${TAG} (${ARCH})"

# 5. Download release assets
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

download_assets() {
    # Method 1: Try gh CLI if available
    if command -v gh >/dev/null 2>&1; then
        if env -u GITHUB_TOKEN gh release download "$TAG" -R "$REPO" -p "$ARCHIVE" -p "$CHECKSUMS" -D "$TMP_DIR" 2>/dev/null; then
            return 0
        fi
    fi

    # Method 2: Standard curl download
    local auth_header=()
    if [ -n "${GITHUB_TOKEN:-}" ]; then
        auth_header=(-H "Authorization: Bearer $GITHUB_TOKEN")
    fi

    local archive_url="https://github.com/${REPO}/releases/download/${TAG}/${ARCHIVE}"
    local checksums_url="https://github.com/${REPO}/releases/download/${TAG}/${CHECKSUMS}"

    info "Downloading ${archive_url}..."
    curl -fsSL "${auth_header[@]}" "$archive_url" -o "${TMP_DIR}/${ARCHIVE}"
    curl -fsSL "${auth_header[@]}" "$checksums_url" -o "${TMP_DIR}/${CHECKSUMS}"
}

download_assets

# 6. Verify SHA256 checksum
info "Verifying SHA256 checksum..."
(
    cd "$TMP_DIR"
    if command -v sha256sum >/dev/null 2>&1; then
        grep "${ARCHIVE}" "${CHECKSUMS}" | sha256sum -c -
    elif command -v shasum >/dev/null 2>&1; then
        grep "${ARCHIVE}" "${CHECKSUMS}" | shasum -a 256 -c -
    else
        info "Notice: Neither sha256sum nor shasum found; skipping checksum verification."
    fi
)

# 7. Unpack and install
info "Extracting ${BINARY}..."
tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "$TMP_DIR" "${BINARY}"

info "Installing to ${DEST_DIR}/${BINARY}..."
if [ -w "$DEST_DIR" ]; then
    install -m 755 "${TMP_DIR}/${BINARY}" "${DEST_DIR}/${BINARY}"
elif command -v sudo >/dev/null 2>&1; then
    sudo install -m 755 "${TMP_DIR}/${BINARY}" "${DEST_DIR}/${BINARY}"
else
    error "Cannot write to ${DEST_DIR}. Please run with sudo or set INSTALL_DIR."
    exit 1
fi

info "nm-go installed successfully at ${DEST_DIR}/${BINARY}"

# 8. Check PATH
case ":$PATH:" in
    *":${DEST_DIR}:"*)
        ;;
    *)
        printf "\n[NOTE] %s is not currently in your PATH.\n" "$DEST_DIR"
        printf "Add it to your shell configuration (e.g. ~/.bashrc or ~/.zshrc):\n"
        printf "  export PATH=\"%s:\$PATH\"\n\n" "$DEST_DIR"
        ;;
esac
