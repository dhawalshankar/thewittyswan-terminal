# deploy/README.md
# GCP VM deploy setup

The VM already has a working deployment script at
`/home/dhawalmannu/thewittyswan-terminal/deploy.sh`. The GitHub Actions deploy
job connects as the existing `dhawalmannu` account and runs that script. It
does not create another Linux user, replace the systemd unit, upload a binary,
or modify the application code.

The existing script runs `git pull`, builds the Go binary on the VM, restarts
`thewittyswan-terminal`, and prints its status. The VM repository must remain
on `main`, and Go must remain installed on the VM.

## 1. SSH key: which file goes where

The SSH key pair has two files:

- The **private key** (`thewittyswan-terminal-deploy`, without `.pub`) stays
  on your Windows PC and is stored in the GitHub Actions secret `SSH_KEY`.
- The **public key** (`thewittyswan-terminal-deploy.pub`) is appended as one
  full line to `/home/dhawalmannu/.ssh/authorized_keys` on the VM.

Do not delete existing lines from `authorized_keys`. Never commit the private
key, VM address, or secret values to the repository.

Generate the key in Windows PowerShell:

```powershell
ssh-keygen -t ed25519 -C "thewittyswan-terminal-deploy" -f "$HOME\.ssh\thewittyswan-terminal-deploy"
Get-Content "$HOME\.ssh\thewittyswan-terminal-deploy.pub"
```

In the VM SSH session as `dhawalmannu`, append the full public-key line:

```sh
mkdir -p ~/.ssh
chmod 700 ~/.ssh
touch ~/.ssh/authorized_keys
vi ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

In `vi`, add the public key on its own line and save. Keep the existing keys.

## 2. Add GitHub Actions secrets

In GitHub, open **DhawalShankar/thewittyswan-terminal → Settings → Secrets and
variables → Actions → New repository secret** and add:

| Secret | Value |
| --- | --- |
| `SSH_HOST` | VM external IP address or DNS name |
| `SSH_USER` | `dhawalmannu` |
| `SSH_KEY` | Contents of the private key file, without `.pub` |
| `SSH_PORT` | Optional; omit for the default `22` |

These are repository secrets—not values to put in workflow files. SSH uses
the key pair, not the VM account password.

## 3. Check VM prerequisites

The existing script must be executable, the checkout must be on `main`, and
the account must be able to run the systemd commands non-interactively:

```sh
cd /home/dhawalmannu/thewittyswan-terminal
git branch --show-current
test -x ./deploy.sh && echo "deploy.sh executable"
sudo -n systemctl status --no-pager
```

If the sudo check reports that a password is required, create a restricted
sudoers drop-in for the commands used by `deploy.sh` and the workflow:

```sh
printf '%s\n' \
  'dhawalmannu ALL=(root) NOPASSWD: /usr/bin/systemctl restart thewittyswan-terminal, /usr/bin/systemctl status --no-pager, /usr/bin/systemctl is-active --quiet thewittyswan-terminal' \
  | sudo tee /etc/sudoers.d/thewittyswan-terminal-deploy >/dev/null
sudo chmod 440 /etc/sudoers.d/thewittyswan-terminal-deploy
sudo visudo -cf /etc/sudoers.d/thewittyswan-terminal-deploy
```

## 4. How deploy runs

1. A pull request to `main` runs CI.
2. After a successful CI run on `main`, GitHub Actions connects to the VM
   using `SSH_HOST`, `SSH_USER`, `SSH_KEY`, and optionally `SSH_PORT`.
3. Actions runs `/home/dhawalmannu/thewittyswan-terminal/deploy.sh`. A script
   failure or inactive systemd service makes the workflow fail.

To run **Deploy** manually, use **GitHub → Actions → Deploy → Run workflow**.
The workflow does not use a new service account or a new systemd unit.

The VM's existing firewall must allow SSH from the GitHub Actions runner.
Avoid opening SSH to all sources just to enable deployment.
