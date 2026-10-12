# Rickroll TUI

An anonymous SSH server with a looping, hand-drawn ASCII dancer and rotating
chorus subtitles. No client software beyond SSH. No audio is transmitted.

```sh
ssh -tt -p 2222 anyone@LXC_IP
```

Any username works, without a password or key. `q` and Ctrl-C are ignored;
press Enter, then `~.` to disconnect with OpenSSH, or close the connection.
Space pauses/resumes. This is a cartoon animation, not a conversion of the music video.

## Create a Proxmox LXC

Download a **Debian 13 or Ubuntu 24.04+** template in Proxmox first. On the
Proxmox host as root, run (replace CT ID, template, storage and bridge):

```sh
git clone https://github.com/hpware/rickroll-tui.git
cd rickroll-tui
./scripts/create-lxc.sh 120 local:vztmpl/YOUR_TEMPLATE.tar.zst local-lvm vmbr0
```

Creates an unprivileged container with DHCP, 1 CPU, 1 GiB RAM and a 4 GiB disk,
installs the service and enables startup on boot. It refuses existing CT IDs.
Internet access and DHCP are required. If installation fails, the new CT remains
for diagnosis; use the existing-container instructions to finish installation.
Docker and LXC nesting are not required. Allow inbound TCP 2222 in your firewall.

## Deploy to an existing LXC

Use Debian 13 or Ubuntu 24.04+ with systemd and root SSH access:

```sh
git clone https://github.com/hpware/rickroll-tui.git
cd rickroll-tui
./deploy.sh root@LXC_IP
# Optional custom port:
./deploy.sh root@LXC_IP 2200
```

Or run inside the LXC as root:

```sh
git clone https://github.com/hpware/rickroll-tui.git
cd rickroll-tui
./scripts/install.sh
```

The installer installs Go when absent, builds natively for the LXC architecture,
and runs the service as a systemd dynamic user. Re-run to update. SSH host keys
persist in `/var/lib/rickroll-tui`. Check logs with
`journalctl -u rickroll-tui`; stop with `systemctl disable --now rickroll-tui`.

## Local development

Requires Go 1.21+:

```sh
go run .
```

Defaults: `RICKROLL_BIND_ADDRESS=0.0.0.0`, `RICKROLL_PORT=2222`,
`RICKROLL_HOST_KEY=host_key.pem`. Never commit the host key.

## Docker

```sh
docker build -t rickroll-tui .
docker run -d --name rickroll-tui --restart unless-stopped -p 2222:2222 \
  -v rickroll-data:/data -e RICKROLL_HOST_KEY=/data/host_key.pem rickroll-tui
```
