    # Fallback to redirect header on GitHub releases URL
    tag=$(curl -sI "https://github.com/${REPO}/releases/latest" 2>/dev/null | grep -i "^location:" | sed -E 's/.*tag\/(.*)/\1/' | tr -d '\r\n' || true)
    if [ -n "$tag" ]; then
        echo "$tag"
        return
    fi

    # Static fallback
    echo "v0.4.0"
}
