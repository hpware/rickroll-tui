#!/usr/bin/env bash
# Run inside a Debian 13 or Ubuntu 24.04+ LXC as root.
set -euo pipefail
[[ $EUID == 0 ]] || { echo 'Run as root (or pipe to sudo bash).' >&2; exit 1; }
port=${1:-2222}
[[ $port =~ ^[0-9]{1,5}$ ]] && ((10#$port >= 1024 && 10#$port <= 65535)) || { echo 'Port must be 1024-65535.' >&2; exit 1; }
command -v systemctl >/dev/null || { echo 'Requires systemd.' >&2; exit 1; }
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
apt-get update
apt-get install -y ca-certificates curl tar gzip golang-go
curl -fsSL https://github.com/hpware/rickroll-tui/archive/refs/heads/main.tar.gz -o "$tmp/source.tar.gz"
tar -xzf "$tmp/source.tar.gz" --strip-components=1 -C "$tmp"
bash "$tmp/scripts/install.sh" "$port"
