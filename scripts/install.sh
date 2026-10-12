#!/usr/bin/env bash
# Run as root inside a Debian 13 / Ubuntu 24.04+ LXC.
set -euo pipefail
[[ $EUID == 0 ]] || { echo 'Run as root.' >&2; exit 1; }
port=${1:-2222}
[[ $port =~ ^[0-9]{1,5}$ ]] && ((10#$port >= 1024 && 10#$port <= 65535)) || { echo 'Port must be 1024-65535.' >&2; exit 1; }
port=$((10#$port))
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."
command -v systemctl >/dev/null
if ! command -v go >/dev/null; then
  apt-get update
  apt-get install -y golang-go ca-certificates
fi
build=$(mktemp -d)
trap 'rm -rf "$build"' EXIT
CGO_ENABLED=0 go build -trimpath -o "$build/rickroll-tui" .
install -d /opt/rickroll-tui
install -m 755 "$build/rickroll-tui" /opt/rickroll-tui/rickroll-tui.new
mv -f /opt/rickroll-tui/rickroll-tui.new /opt/rickroll-tui/rickroll-tui
cat > /etc/systemd/system/rickroll-tui.service <<UNIT
[Unit]
Description=Anonymous ASCII Rickroll SSH
After=network.target

[Service]
ExecStart=/opt/rickroll-tui/rickroll-tui
DynamicUser=yes
StateDirectory=rickroll-tui
Environment=RICKROLL_HOST_KEY=/var/lib/rickroll-tui/host_key.pem
Environment=RICKROLL_BIND_ADDRESS=0.0.0.0
Environment=RICKROLL_PORT=$port
Restart=on-failure
RestartSec=2
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes

[Install]
WantedBy=multi-user.target
UNIT
systemctl daemon-reload
systemctl enable rickroll-tui
systemctl restart rickroll-tui
systemctl is-active --quiet rickroll-tui
printf 'Installed. Connect: ssh -tt -p %s anyone@LXC_IP\n' "$port"
