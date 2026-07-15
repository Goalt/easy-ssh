# easy-ssh

<p align="center">
  <img src="https://img.shields.io/github/actions/workflow/status/Goalt/easy-ssh/ci.yml?branch=main&style=flat-square" alt="Build Status" />
  <img src="https://img.shields.io/github/v/release/Goalt/easy-ssh?style=flat-square" alt="Latest Release" />
  <img src="https://img.shields.io/github/license/Goalt/easy-ssh?style=flat-square" alt="License" />
</p>

`easy-ssh` is a modern, beautiful terminal user interface (TUI) command-line tool written in Go that simplifies creating instant, secure Cloudflare Tunnels for local TCP services—with zero configuration.

Specifically designed for exposing local SSH services (`port 22`), `easy-ssh` checks if your port is active, automatically installs the required `cloudflared` daemon in a isolated user cache if it isn't present, sets up the tunnel, and renders a stunning connection dashboard with ready-to-copy SSH proxy commands.

---

## ✨ Features

- **Modern TUI Experience:** Powered by Bubble Tea and Lip Gloss for smooth animations, status indicators, and gorgeous layouts.
- **Surgical Port Checking:** Automatically verifies if a service is actively listening on the specified port. If not, it displays a warning and continues.
- **Zero-Dependency Bootstrapping:** Automatically detects and downloads the correct official `cloudflared` binary for your OS/Arch directly to `~/.easy-ssh/bin/` if not found in your `PATH`.
- **Pre-packaged Install Script:** Single-command installer script matching modern standards.
- **Multi-platform Docker Images:** Ready-to-use Docker images supporting both `amd64` (x86) and `arm64` (aarch64) platforms, hosted on GitHub Container Registry (GHCR).

---

## 🚀 Getting Started

### Quick Install & Run (curl | bash)

You can easily download and install the latest `easy-ssh` binary to `/usr/local/bin` (or `~/.local/bin` if running without root permissions) by executing:

```bash
curl -sSfL https://raw.githubusercontent.com/Goalt/easy-ssh/main/install.sh | bash
```

Once installed, simply run:

```bash
easy-ssh --port 22
```

### Go Install (Go 1.24+)

You can also install the latest stable version of `easy-ssh` directly using Go:

```bash
go install github.com/Goalt/easy-ssh@latest
```

*Note: Make sure your `GOBIN` directory (usually `~/go/bin`) is included in your system's `PATH`.*

### Manual Download & Install

You can also manually download the pre-compiled binary matching your platform directly from the [GitHub Releases](https://github.com/Goalt/easy-ssh/releases) page. For example, to download and install for Linux AMD64:

```bash
# Download the latest binary for Linux AMD64
curl -sSfL -o easy-ssh https://github.com/Goalt/easy-ssh/releases/latest/download/easy-ssh-linux-amd64

# Make the binary executable
chmod +x easy-ssh

# Move the binary to a directory in your PATH (e.g., /usr/local/bin)
sudo mv easy-ssh /usr/local/bin/
```

### Docker

Run `easy-ssh` directly as a lightweight container utilizing your host's network:

```bash
docker run -it --net=host ghcr.io/goalt/easy-ssh:latest --port 22
```

### Run/Build From Source

Ensure you have [Go 1.24+](https://go.dev/) installed, then:

```bash
# Clone the repository
git clone https://github.com/Goalt/easy-ssh.git
cd easy-ssh

# Run directly
go run main.go --port 22

# Or build the binary
go build -o easy-ssh main.go
./easy-ssh --port 22
```

---

## ⚙️ CLI Options

`easy-ssh` supports the following optional flags:

```bash
easy-ssh --help
```

- `-p, --port int`: The target TCP port number you want to tunnel (default: `22`).
- `-c, --cloudflared-path string`: Path to a custom pre-installed `cloudflared` executable (skips auto-detection/download).

---

## 🔒 Connecting to your Tunnel

When the tunnel is successfully established, `easy-ssh` will display a success card with your temporary `trycloudflare.com` address and the exact SSH client connection command:

```bash
ssh -o "ProxyCommand=cloudflared access tcp --hostname %h --port %p" user@<your-tunnel-subdomain>.trycloudflare.com
```

*Note: The connecting client machine must have `cloudflared` installed locally to support the SSH proxy connection.*

---

## 🛠️ Development

This project comes pre-configured with a VSCode Devcontainer to provide a complete, instant development environment.

### Run Tests

To run the unit tests:

```bash
go test -v ./...
```
