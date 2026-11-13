<div align="center">
  <h1>🦢 thewittyswan-terminal</h1>
  <p><strong>A real, interactive developer terminal in your browser.</strong></p>
  <p>Explore the blog, inspect network security, and try the tools behind the site.</p>
  <p>
    <a href="https://thewittyswan.space/terminal"><strong>Open the live terminal</strong></a>
    &nbsp;·&nbsp;
    <a href="https://thewittyswan.space">Visit the main site</a>
  </p>
  <p>
    <a href="https://github.com/DhawalShankar/thewittyswan-terminal/actions/workflows/ci.yml">
      <img src="https://github.com/DhawalShankar/thewittyswan-terminal/actions/workflows/ci.yml/badge.svg?branch=main" alt="CI status">
    </a>
    <img src="https://img.shields.io/badge/Go-1.26.5-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26.5">
  </p>
  <sub>Author: <strong>Er. Dhawal Shukla</strong></sub>
</div>

---

## The idea

The main site is for everyone; this is for the curious developer. Type commands,
get real answers, and explore the blog and portfolio through a terminal-style
interface. One lightweight Go process serves the UI and API—no frontend build
pipeline or component framework required.

## At a glance

| Layer | Technology |
| --- | --- |
| Backend and HTTP server | Go |
| Terminal interface | xterm.js |
| Hosting | GCP e2-micro VM |
| Reverse proxy and HTTPS | Caddy |
| Process supervision | systemd |

The terminal reads public JSON feeds from the separate `thewittyswan` Jekyll
blog and portfolio site. It does not share their codebase or database. SQLite
persistence and the Swan assistant are planned, not current runtime features.

## Recent improvements (13 Nov 2026)

- Added GitHub Actions CI for Go formatting, dependency consistency, vetting,
  linting, race-enabled tests, and builds. The lint step is currently
  non-blocking. Successful CI on `main` triggers deployment over SSH to the
  existing VM; the deploy workflow runs the VM's existing `deploy.sh`, so no
  second VM or container registry is involved.
- `resume pdf`, `resume open`, and `resume download` open the resume PDF in a
  new tab when the browser allows it. HTTP and HTTPS links printed in the
  terminal are clickable.
- `cat` accepts multiword blog titles, quoted or unquoted, and has regression
  coverage.
- Hardened HTTP handling with server timeouts, an 8 KiB `/run` request limit,
  loopback-only trust for forwarded client IPs, and a `/healthz` endpoint.
  Feed requests have timeouts and report failures in the logs; `curl` reads at
  most 2,000 response bytes.
- Improved terminal clipboard shortcuts, multiline paste in the built-in
  editor, responsive layout, zoom accessibility, scroll controls, and
  keyboard-accessible window resizing.

## Explore

### Blog and portfolio

| Command | What it does |
| --- | --- |
| `help`, `whoami` | Show available commands and terminal info |
| `ls blog/`, `ls blog/netsec`, `ls blog/literature` | Browse posts from the live blog feed |
| `cat <slug or title>` | Read a post by slug or title; multiword titles may be quoted |
| `grep <term>` | Search post titles and content |
| `refresh` | Refresh the blog and portfolio feeds |
| `resume pdf` | Open the resume PDF (new tab when allowed by the browser) |
| `about`, `resume`, `projects`, `skills`, `experience` | Explore portfolio details |
| `education`, `achievements`, `timeline`, `github`, `leetcode`, `contact` | More portfolio commands |

Virtual filesystem commands are handled client-side by
[`static/js/terminal.js`](static/js/terminal.js). HTTP and HTTPS URLs in terminal
output open in a new tab with Ctrl+click; on touch screens, tap the link.
For example, `cat "Use AI only where it belongs"` reads the matching post.
Use Ctrl/Cmd+V to paste. Select terminal text and press Ctrl/Cmd+C to copy;
Ctrl/Cmd+C with no selection cancels the current input or request. In the
built-in `edit` command, pasted newlines are preserved.

### Network and security

| Command | What it does |
| --- | --- |
| `headers <domain>` | Inspect security headers and cookie flags |
| `tls <domain>` | Inspect certificate, protocol, cipher, and expiry |
| `dig <host>` | Look up DNS records |
| `ping <host>` | Check TCP reachability |
| `curl <url>` | Fetch and preview an HTTP response |
| `ports <host>` | Check common ports |
| `pkgcheck <ecosystem> <package>` | Look up dependency advisories from OSV.dev |

## Run locally

Use the Go version specified in [`go.mod`](go.mod), then run:

```powershell
go run .
```

Open [http://localhost:8080/terminal](http://localhost:8080/terminal).

## How it works

The app fetches `https://thewittyswan.space/posts.json` and
`https://thewittyswan.space/portfolio.json` at startup, then refreshes both
feeds every 10 minutes. Results are cached in memory; if a feed is temporarily
unavailable, the terminal keeps serving its last known cache. The `refresh`
command updates both feeds on demand. Feed requests have an eight-second
timeout and log fetch, status, or decode errors without discarding the last
valid cache.

## Security

The `/run` endpoint treats commands as untrusted input:

- Per-IP token-bucket rate limiting with `golang.org/x/time/rate`
- Security headers including CSP, HSTS, frame blocking, and MIME-sniffing
  protection
- Hostname validation and DNS checks that block loopback, private, link-local,
  multicast, unspecified, and cloud metadata addresses
- Three-second DNS and eight-second outbound HTTP timeouts
- HTTP server read/write/idle timeouts and an 8 KiB limit on `/run` request
  bodies
- Forwarded client IPs are trusted only when the connection comes from the
  loopback reverse proxy
- The `curl` command previews at most 2,000 response bytes
- TLS inspection and network tools routed through the same host guard
- No arbitrary shell execution; commands use an allowlisted dispatcher

See `main.go`, `commands.go`, and `portfolio_commands.go` for the relevant
implementation.

## CI/CD and production

Production runs on a GCP e2-micro VM. Caddy routes
`thewittyswan.space/terminal` to the Go service; systemd keeps the binary alive
across crashes and reboots. GitHub Actions runs CI on pushes to `main` and pull
requests. After successful CI on `main`, the deploy workflow connects to the
same VM over SSH and runs its existing `deploy.sh`; it can also be started
manually. Afterward, it checks the local `/healthz` endpoint, verifies that the
systemd service is active, and confirms the VM's `authorized_keys` file is
nonempty.

See [`deploy/README.md`](deploy/README.md) for deployment details.

## License

Personal project, all rights reserved.
