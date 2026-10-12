#!/usr/bin/env bash
# Run on a Proxmox host. Template must already be downloaded.
set -euo pipefail
CTID=${1:?Usage: create-lxc.sh CTID TEMPLATE [STORAGE] [BRIDGE]}
TEMPLATE=${2:?Provide a Debian 13 or Ubuntu 24.04+ template volume}
STORAGE=${3:-local-lvm}
BRIDGE=${4:-vmbr0}
[[ $CTID =~ ^[0-9]+$ ]] || exit 1
[[ $EUID == 0 ]] || { echo 'Run as root on the Proxmox host.' >&2; exit 1; }
command -v pct >/dev/null
if pct status "$CTID" >/dev/null 2>&1; then echo "CT $CTID already exists; refusing to overwrite." >&2; exit 1; fi
pct create "$CTID" "$TEMPLATE" --hostname rickroll --unprivileged 1 \
  --cores 1 --memory 1024 --swap 512 --rootfs "$STORAGE:4" \
  --net0 "name=eth0,bridge=$BRIDGE,ip=dhcp" --onboot 1
pct start "$CTID"
# Allow DHCP/DNS to settle before the first package install.
pct exec "$CTID" -- bash -c 'for i in {1..30}; do getent hosts deb.debian.org >/dev/null && exit 0; sleep 1; done; exit 1'
pct exec "$CTID" -- bash -c 'set -eu; apt-get update; apt-get install -y git ca-certificates; git clone https://github.com/hpware/rickroll-tui.git /opt/rickroll-src; bash /opt/rickroll-src/scripts/install.sh'
pct exec "$CTID" -- hostname -I
echo 'Ready: ssh -tt -p 2222 anyone@LXC_IP (use the address above).'
