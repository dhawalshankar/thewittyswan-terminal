# deploy/README.md
# GCP VM one-time setup

The existing VM deployment confirms:

- SSH deploy account: `dhawalmannu`
- Application directory: `/home/dhawalmannu/thewittyswan-terminal`
- Binary and systemd service: `thewittyswan-terminal`
- Listen port: 8080
- There is no `/health` or `/healthz` endpoint, so deployment checks the
  systemd active state.

The checked-in systemd unit runs the application as a separate non-root
`thewittyswan-terminal` system user. The deploy workflow ships only the binary.
The app also serves the `static/` directory, so the existing directory must
remain in place on the VM and be readable by the service user.

## 1. Create the service user and prepare the application directory

Run on the VM as `dhawalmannu` (or another administrator with sudo access):

```sh
export DEPLOY_USER="dhawalmannu"
export APP_DIR="/home/dhawalmannu/thewittyswan-terminal"
export SERVICE_NAME="thewittyswan-terminal"
export SERVICE_USER="thewittyswan-terminal"

sudo groupadd --system "$SERVICE_USER"
sudo useradd --system --no-create-home --gid "$SERVICE_USER" --shell /usr/sbin/nologin "$SERVICE_USER"
sudo usermod -aG systemd-journal "$DEPLOY_USER"
sudo chgrp -R "$SERVICE_USER" "$APP_DIR"
sudo chmod -R g+rX "$APP_DIR"
sudo chmod 2755 "$APP_DIR"
```

The existing `dhawalmannu` user owns the directory and uploads releases;
setgid ensures uploaded files inherit the service group. The service user
needs read/execute access to the binary and `static/` files, but no write
access.

## 2. Install and enable systemd

From the checked-out repository on the VM:

```sh
sudo cp deploy/thewittyswan-terminal.service /etc/systemd/system/thewittyswan-terminal.service
sudo systemctl daemon-reload
sudo systemctl enable thewittyswan-terminal
```

## 3. Create the environment file

The current application has no required environment variables. Create the
file on the VM only; do not commit it:

```sh
sudo -u "$DEPLOY_USER" touch "$APP_DIR/.env"
sudo chmod 600 "$APP_DIR/.env"
sudo chown "$DEPLOY_USER:$SERVICE_USER" "$APP_DIR/.env"
```

Edit it on the VM if runtime environment values are added in the future:

```sh
sudo -u "$DEPLOY_USER" vi "$APP_DIR/.env"
```

## 4. Allow only the required systemctl commands

Create a sudoers drop-in without a heredoc:

```sh
printf '%s\n' \
  "$DEPLOY_USER ALL=(root) NOPASSWD: /usr/bin/systemctl restart $SERVICE_NAME, /usr/bin/systemctl is-active $SERVICE_NAME" \
  | sudo tee "/etc/sudoers.d/$SERVICE_NAME-deploy" >/dev/null
sudo chmod 440 "/etc/sudoers.d/$SERVICE_NAME-deploy"
sudo visudo -cf "/etc/sudoers.d/$SERVICE_NAME-deploy"
```

## 5. Add the deploy SSH key

Generate a dedicated key on Windows PowerShell:

```powershell
ssh-keygen -t ed25519 -C "thewittyswan-terminal-deploy" -f "$HOME\.ssh\thewittyswan-terminal-deploy"
Get-Content "$HOME\.ssh\thewittyswan-terminal-deploy.pub"
```

Add the displayed public key as one line to the existing `dhawalmannu`
account's authorized keys on the VM:

```sh
mkdir -p "$HOME/.ssh"
chmod 700 "$HOME/.ssh"
touch "$HOME/.ssh/authorized_keys"
chmod 600 "$HOME/.ssh/authorized_keys"
vi "$HOME/.ssh/authorized_keys"
ssh-keygen -lf "$HOME/.ssh/authorized_keys"
```

Store the secrets in GitHub Actions. Run from Windows PowerShell:

```powershell
gh secret set SSH_HOST --repo DhawalShankar/thewittyswan-terminal
gh secret set SSH_USER --repo DhawalShankar/thewittyswan-terminal --body "dhawalmannu"
Get-Content "$HOME\.ssh\thewittyswan-terminal-deploy" | gh secret set SSH_KEY --repo DhawalShankar/thewittyswan-terminal
```

`SSH_HOST` is prompted interactively; the VM address must not be written to
this repository. `SSH_PORT` is optional and defaults to 22.

## 6. Allow SSH through the GCP firewall

From a machine authenticated to the intended GCP project, create a rule
allowing TCP port 22 to the VM. Replace the three placeholders with the rule
name, the VM's network tag, and a trusted source CIDR:

```powershell
gcloud compute firewall-rules create TODO_FILL_SSH_FIREWALL_RULE `
  --allow tcp:22 `
  --target-tags TODO_FILL_VM_NETWORK_TAG `
  --source-ranges TODO_FILL_TRUSTED_CIDR
```

Do not use `0.0.0.0/0` unless that exposure is an intentional, documented
decision.

## 7. First deployment

After pushing the branch and opening the pull request, merge it to `main` for
the normal CI-triggered deployment. To manually test the deployment workflow
from the GitHub Actions UI, or from Windows PowerShell:

```powershell
gh workflow run deploy.yml --repo DhawalShankar/thewittyswan-terminal
```

The workflow builds a static Linux amd64 binary, uploads it as
`thewittyswan-terminal.new`, atomically swaps it into place, restarts
`thewittyswan-terminal`, and checks that systemd reports it active.

## 8. Manual rollback

Run on the VM:

```sh
cd /home/dhawalmannu/thewittyswan-terminal
mv -f thewittyswan-terminal thewittyswan-terminal.failed
mv -f thewittyswan-terminal.prev thewittyswan-terminal
chmod +x thewittyswan-terminal
sudo systemctl restart thewittyswan-terminal
sudo systemctl is-active thewittyswan-terminal
```

If the service is not active, inspect the logs:

```sh
journalctl -u thewittyswan-terminal -n 30 --no-pager
```
