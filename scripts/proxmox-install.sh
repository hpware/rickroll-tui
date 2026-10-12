#!/usr/bin/env bash
# Run as root on the Proxmox VE host, not inside an existing container.
set -euo pipefail
[[ $EUID == 0 ]] || { echo 'Run as root on your Proxmox host.' >&2; exit 1; }
for cmd in pct pvesh pvesm pveam curl; do
  command -v "$cmd" >/dev/null || { echo "Missing $cmd. Run this on a Proxmox VE host." >&2; exit 1; }
done

# Optional environment overrides: CTID, STORAGE, TEMPLATE_STORAGE, BRIDGE.
CTID=${CTID:-$(pvesh get /cluster/nextid)}
[[ $CTID =~ ^[0-9]+$ ]] || { echo 'Invalid CTID.' >&2; exit 1; }
if pct status "$CTID" >/dev/null 2>&1; then
  echo "CT $CTID already exists; refusing to overwrite." >&2
  exit 1
fi
choose_storage() {
  pvesm status --content "$1" --enabled 1 | awk 'NR > 1 && $3 == "active" {print $1; exit}'
}
STORAGE=${STORAGE:-$(choose_storage rootdir)}
TEMPLATE_STORAGE=${TEMPLATE_STORAGE:-$(choose_storage vztmpl)}
[[ -n $STORAGE && -n $TEMPLATE_STORAGE ]] || {
  echo 'Need active storage supporting rootdir and vztmpl.' >&2; exit 1;
}
BRIDGE=${BRIDGE:-vmbr0}
[[ -d /sys/class/net/$BRIDGE/bridge ]] || { echo "Bridge $BRIDGE not found. Set BRIDGE to your LAN bridge." >&2; exit 1; }

pveam update
TEMPLATE=$(pveam available --section system | awk '$2 ~ /^debian-13-standard_.*_amd64\.tar\.(zst|gz|xz)$/ {print $2}' | sort -V | tail -n 1)
[[ -n $TEMPLATE ]] || { echo 'No Debian 13 template available.' >&2; exit 1; }
echo "Creating CT $CTID on $STORAGE, bridge $BRIDGE (DHCP), using $TEMPLATE."
pveam download "$TEMPLATE_STORAGE" "$TEMPLATE"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
curl -fsSL https://raw.githubusercontent.com/jsserve-org/rickroll-tui/main/scripts/create-lxc.sh -o "$tmp/create-lxc.sh"
bash "$tmp/create-lxc.sh" "$CTID" "$TEMPLATE_STORAGE:vztmpl/$TEMPLATE" "$STORAGE" "$BRIDGE"
