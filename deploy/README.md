# deploy/README.md
# GCP VM one-time setup

The repository does not contain an existing systemd unit, service name,
application directory, deploy user, binary name, or health endpoint. Before
using this guide, replace every `TODO_FILL_*` value in
`.github/workflows/deploy.yml` and `deploy/TODO_SERVICE_NAME.service`.

The application listens on port 8080. It has no `/health`, `/healthz`, or
other dedicated health endpoint, so leave `HEALTH_URL` as
`TODO_FILL_HEALTH_URL` to use the systemd active-state check.

## 1. Create the deploy user and application directory

Run on the VM as a user with sudo access. Set these shell variables to the
same values used in the workflow and service unit:

```sh
export DEPLOY_USER="TODO_FILL_DEPLOY_USER"
export APP_DIR="TODO_FILL_APP_DIR"
export SERVICE_NAME="TODO_FILL_SERVICE_NAME"
export BINARY_NAME="TODO_FILL_BINARY_NAME"

sudo useradd --system --create-home --shell /bin/bash "$DEPLOY_USER"
sudo mkdir -p "$APP_DIR"
sudo chown "$DEPLOY_USER:$DEPLOY_USER" "$APP_DIR"
sudo usermod -aG systemd-journal "$DEPLOY_USER"
```

## 2. Install and enable systemd

After replacing the placeholders in the unit file, copy it from a checked-out
repository on the VM:

```sh
sudo cp "deploy/$SERVICE_NAME.service" "/etc/systemd/system/$SERVICE_NAME.service"
sudo systemctl daemon-reload
sudo systemctl enable "$SERVICE_NAME"
```

## 3. Create the environment file

Create the file on the VM only. Do not commit it:

```sh
sudo -u "$DEPLOY_USER" touch "$APP_DIR/.env"
sudo chmod 600 "$APP_DIR/.env"
```

Add the service's runtime environment values using your preferred editor:

```sh
sudo -u "$DEPLOY_USER" vi "$APP_DIR/.env"
```

The checked-in application currently has no required environment variables.

## 4. Allow only the required systemctl commands

Create the sudoers drop-in without a heredoc:

```sh
printf '%s\n' \
  "$DEPLOY_USER ALL=(root) NOPASSWD: /usr/bin/systemctl restart $SERVICE_NAME, /usr/bin/systemctl is-active $SERVICE_NAME" \
  | sudo tee "/etc/sudoers.d/$SERVICE_NAME-deploy" >/dev/null
sudo chmod 440 "/etc/sudoers.d/$SERVICE_NAME-deploy"
sudo visudo -cf "/etc/sudoers.d/$SERVICE_NAME-deploy"
```

## 5. Add the deploy SSH key

Generate the key on Windows PowerShell, not on the VM:

```powershell
ssh-keygen -t ed25519 -C "thewittyswan-terminal-deploy" -f "$HOME\.ssh\thewittyswan-terminal-deploy"
Get-Content "$HOME\.ssh\thewittyswan-terminal-deploy.pub"
```

On the VM, add the printed public key to the deploy user's authorized keys:

```sh
sudo -u "$DEPLOY_USER" mkdir -p "/home/$DEPLOY_USER/.ssh"
sudo chmod 700 "/home/$DEPLOY_USER/.ssh"
sudo -u "$DEPLOY_USER" touch "/home/$DEPLOY_USER/.ssh/authorized_keys"
sudo chmod 600 "/home/$DEPLOY_USER/.ssh/authorized_keys"
sudo chown -R "$DEPLOY_USER:$DEPLOY_USER" "/home/$DEPLOY_USER/.ssh"
sudo -u "$DEPLOY_USER" vi "/home/$DEPLOY_USER/.ssh/authorized_keys"
```

Paste the public key as one line, save it, and verify:

```sh
sudo ssh-keygen -lf "/home/$DEPLOY_USER/.ssh/authorized_keys"
```

Store the private key in GitHub Actions with `gh` from Windows PowerShell:

```powershell
gh secret set SSH_HOST --repo DhawalShankar/thewittyswan-terminal
gh secret set SSH_USER --repo DhawalShankar/thewittyswan-terminal --body "TODO_FILL_DEPLOY_USER"
Get-Content "$HOME\.ssh\thewittyswan-terminal-deploy" | gh secret set SSH_KEY --repo DhawalShankar/thewittyswan-terminal
gh secret set SSH_PORT --repo DhawalShankar/thewittyswan-terminal --body "22"
```

Set `SSH_HOST` interactively when prompted. It must not be written to this
repository.

## 6. Allow SSH through the GCP firewall

Run from a machine authenticated to the intended GCP project. Replace the
placeholder values before running:

```powershell
gcloud compute firewall-rules create TODO_FILL_SSH_FIREWALL_RULE `
  --allow tcp:22 `
  --target-tags TODO_FILL_VM_NETWORK_TAG `
  --source-ranges TODO_FILL_TRUSTED_CIDR
```

Do not use `0.0.0.0/0` unless that exposure is an intentional, documented
decision.

## 7. First deployment

After the placeholders are filled and the secrets are set, push the branch and
open the pull request. Once the change is merged to `main`, CI success triggers
the deployment workflow. To test it without waiting for a merge, use the
workflow dispatch action in GitHub or run:

```powershell
gh workflow run deploy.yml --repo DhawalShankar/thewittyswan-terminal
```

The workflow builds a static Linux amd64 binary, uploads it as
`BINARY_NAME.new`, atomically swaps it into place, restarts systemd, and
validates the service.

## 8. Manual rollback

Run on the VM:

```sh
cd "$APP_DIR"
sudo systemctl stop "$SERVICE_NAME"
mv -f "$BINARY_NAME" "$BINARY_NAME.failed"
mv -f "$BINARY_NAME.prev" "$BINARY_NAME"
chmod +x "$BINARY_NAME"
sudo systemctl start "$SERVICE_NAME"
sudo systemctl is-active "$SERVICE_NAME"
```

If the service is not active, inspect the logs:

```sh
journalctl -u "$SERVICE_NAME" -n 30 --no-pager
```

The deploy workflow uploads only the binary. The application also reads the
checked-in `static/` directory at runtime, so provision that directory
separately on the VM before the first deployment; it is not copied by CI/CD.
