#!/usr/bin/env bash
# Copy source and build natively inside the target LXC; Docker is not needed.
set -euo pipefail
TARGET=${1:?Usage: ./deploy.sh root@lxc-host [port]}
PORT=${2:-2222}
[[ $TARGET != -* && $TARGET != *[[:space:]]* ]] || exit 1
[[ $PORT =~ ^[0-9]{1,5}$ ]] && ((10#$PORT >= 1024 && 10#$PORT <= 65535)) || { echo 'Port must be 1024-65535.' >&2; exit 1; }
cd -- "$(dirname -- "${BASH_SOURCE[0]}")"
# Stream only source files. Never copy host keys or repository credentials.
archive=$(mktemp)
trap 'rm -f "$archive"' EXIT
tar -czf "$archive" main.go go.mod go.sum scripts/install.sh
ssh "$TARGET" "set -eu; test \"\$(id -u)\" = 0; d=\$(mktemp -d); trap 'rm -rf \"\$d\"' EXIT; tar -xzf - -C \"\$d\"; bash \"\$d/scripts/install.sh\" '$PORT'" < "$archive"
