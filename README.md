# thewittyswan-terminal

The dev-facing half of thewittyswan. A real, interactive terminal in the
browser — networking and security tools, and a browsable feed of the
blog itself. Lives at `thewittyswan.space/terminal`.

Built by **Dhawal**.

## What this repository demonstrates

This project is a compact Go systems application that combines HTTP serving,
background synchronization, a virtual terminal UX, outbound network tooling,
rate limiting, security headers, SSRF defenses, and Linux service deployment.
It is intentionally framework-light: one Go process serves the UI and API.

## Why this exists

The main site (`thewittyswan`) is for anyone. This one is for the "I'm a
dev" crowd — a place to actually poke at something instead of just
reading about it. Type commands, get real answers, maybe learn something
about how the site itself is built along the way.

## Why Go + xterm.js, not a framework

Same philosophy as the blog repo: simple, boring, easy to reason about.
Go compiles to a single binary — no runtime dependencies, no build
pipeline to fight with, easy to deploy to a plain VM. xterm.js gives a
genuine terminal look (cursor, monospace grid, keyboard capture) without
needing React or any component framework wrapped around it. One Go
process serves both the frontend page and the API — nothing else
running, nothing else to manage.

## Stack

- **Go** — backend, command parsing, HTTP server
- **xterm.js** — terminal emulator, loaded via CDN, no build step
- **SQLite** — planned for persistence; not currently wired in
- **Groq API** — planned for the swan assistant; not currently wired in
- **GCP e2-micro VM** — hosting (Always Free tier, always-on, no cold starts)
- **Caddy** — reverse proxy + automatic HTTPS on the VM
- **systemd** — keeps the Go binary running, auto-restarts on crash/reboot

This is one of two repos. The other, `thewittyswan`, is the Jekyll blog
and portfolio, hosted separately on Vercel. The two are independent —
this terminal reads the blog's public `posts.json` feed rather than
sharing a codebase or database with it.

## Local setup

```powershell
go run main.go blog.go commands.go
```

Visit `http://localhost:8080`.

## Commands (current)

- `help` — list available commands
- `whoami` — basic info about the terminal
- `ls blog/`, `ls blog/netsec`, `ls blog/literature` — list posts from the live blog feed
- `cat <slug or title>` — print a post's content
- `grep <term>` — search post titles and content
- `refresh` — manually re-fetch `posts.json` (blog auto-syncs every 10 minutes otherwise)

The terminal also exposes portfolio commands such as `about`, `resume`,
`projects`, `skills`, `experience`, `education`, `achievements`, `timeline`,
`github`, `leetcode`, and `contact`. Virtual filesystem commands are handled
client-side by `static/js/terminal.js`.

## Network and security commands

- `headers <domain>` — security headers and cookie-flag report
- `tls <domain>` — certificate, protocol, cipher, and expiry inspection
- `dig <host>` — DNS record diagnostics
- `ping <host>` — TCP reachability check
- `curl <url>` — controlled HTTP fetch and response preview
- `ports <host>` — common-port reachability checks
- `pkgcheck <ecosystem> <package>` — OSV.dev dependency vulnerability lookup

The Swan assistant and SQLite persistence remain planned; they are not current
runtime features.

## How the blog connection works

`blog.go` fetches `https://thewittyswan.space/posts.json` on startup and
every 10 minutes after that, caching it in memory. The terminal never
touches the blog's GitHub repo or database directly — just this one
public JSON feed. If the blog is temporarily unreachable, the terminal
keeps serving its last known cache rather than breaking.

`portfolio.go` follows the same integration pattern for
`https://thewittyswan.space/portfolio.json`. Both caches are protected by
read/write mutexes and refreshed by background goroutines; the `refresh`
command updates both feeds on demand.

## Security model

The `/run` endpoint is treated as an untrusted-input boundary:

- Per-IP token-bucket rate limiting (`golang.org/x/time/rate`)
- Security headers including CSP, HSTS, frame blocking, and MIME sniffing
  protection
- Hostname validation before any outbound network request
- DNS resolution checks against loopback, private, link-local, multicast,
  unspecified, and cloud metadata addresses
- Three-second DNS resolution and eight-second HTTP timeouts
- TLS inspection and network utilities routed through the same host guard
- No arbitrary shell execution; commands are parsed and dispatched through
  an allowlisted switch in `commands.go`

The relevant implementation is in `main.go`, `commands.go`, and
`portfolio_commands.go`.

## Production deployment

The production deployment uses a GCP e2-micro VM, a compiled Go binary,
systemd for process supervision, and Caddy for reverse proxying and automatic
HTTPS. Caddy routes `thewittyswan.space/terminal` to the Go service, while
the root `thewittyswan.space` application is the Jekyll site deployed on
Vercel.

The VM deployment was completed using an SSH/bash deployment workflow in the
free-tier GCP environment. `go.mod` and `go.sum` define the minimal Go
dependency surface; systemd keeps the process alive across crashes and
reboots.

## License

Personal project, all rights reserved.