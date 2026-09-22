# hostcli

A modern, cross-platform `/etc/hosts` manager written in Go, featuring an interactive [Bubble Tea](https://github.com/charmbracelet/bubbletea) Terminal User Interface (TUI), scriptable CLI, safe atomic application, one-click rollback, and cross-platform migration for macOS, Linux, and Windows.

---

## Features

- **Split-Pane Interactive TUI**: Beautiful keyboard-driven interface built with Bubble Tea and Lipgloss.
  - **Left Sidebar**: Group hierarchies with active counts and Quick Actions cheatsheet (`[a] Add`, `[e] Edit`, `[d] Delete`, `[Space] Toggle`, `[A] Apply`, `[R] Rollback`).
  - **Right Main Area**: Host entries table with real-time status badges (`● ON` / `○ OFF`), IP, hostnames, aliases, and comments.
- **Zero-Data-Loss Architecture**: Changes are saved instantly to `~/.hostcli/hosts.yaml` in user space without requiring root or administrator privileges. Edits are never lost if an apply or sudo prompt is cancelled.
- **Deterministic Boundary Markers**: All system hosts edits are strictly bounded between `# BEGIN HOSTCLI` and `# END HOSTCLI`. System entries and custom rules are preserved untouched.
- **First-Class Rollback**:
  - `hostcli rollback`: Instantly strips the managed block from system hosts.
  - `hostcli rollback --original`: Restores the pristine pre-migration backup (`hosts.original.bak`).
- **Cross-Platform Migration (`hostcli import`)**: Automatically detects and migrates existing hosts on macOS (`/etc/hosts`), Linux (`/etc/hosts`), and Windows (`System32\drivers\etc\hosts`).
- **Sudo / UAC Elevation Handoff**: `tea.ExecProcess` temporarily yields terminal control to `sudo` for clean password entry without crashing the TUI or corrupting terminal states.
- **Automatic DNS Cache Flush**: Flushes OS-level DNS caches automatically after apply or rollback (`dscacheutil` on macOS, `resolvectl` on Linux, `ipconfig /flushdns` on Windows).

---

## Installation

### Homebrew (macOS & Linux)

```bash
brew tap orekasep/tap
brew install hostcli
```

### Windows (Chocolatey)

```powershell
choco install hostcli
```

### Build from Source (Go 1.23+)

```bash
git clone https://github.com/orekasep/go-hosts-cli.git
cd go-hosts-cli
go build -o hostcli ./cmd/hostcli
sudo mv hostcli /usr/local/bin/
```

Or install directly with `go install`:

```bash
go install github.com/orekasep/go-hosts-cli/cmd/hostcli@latest
```

---

## Quick Start

### 1. Interactive TUI Mode

Simply run `hostcli` with no arguments:

```bash
hostcli
```

#### Keybindings

| Key | Context | Action |
|---|---|---|
| `Tab` / `h` / `l` | Normal | Switch focus between Groups sidebar and Host Table |
| `j` / `k` / `↓` / `↑` | Normal | Navigate items / rows |
| `Space` | Host Table | Toggle selected host enabled / disabled |
| `a` | Normal | Open Add Host modal |
| `e` / `Enter` | Host Table | Open Edit Host modal |
| `d` | Host Table | Open Delete confirmation modal |
| `/` | Normal | Fuzzy search across hostnames, aliases, and IPs |
| `A` | Normal | Apply enabled entries to system hosts |
| `R` | Normal | Open Rollback options modal |
| `m` | Normal | Import existing system hosts file |
| `?` | Normal | Open keybindings help dialog |
| `q` / `Ctrl+C` | Normal | Quit `hostcli` |

---

### 2. Scriptable CLI Mode

```bash
# Import existing system hosts into ~/.hostcli/hosts.yaml
hostcli import

# Add a new entry to a group with aliases and comment
hostcli add 192.168.1.50 api.dev.local --group work --alias api,gateway --comment "Dev API Gateway"

# List configured entries
hostcli list

# Preview changes without modifying system hosts
hostcli apply --dry-run

# Apply enabled entries to system hosts (prompts for sudo if unprivileged)
hostcli apply

# Toggle entries without deleting them
hostcli disable api.dev.local
hostcli enable api.dev.local

# Remove an entry
hostcli rm api.dev.local

# Inspect differences against system hosts
hostcli diff

# Rollback managed block from system hosts
hostcli rollback

# Restore pristine pre-migration backup
hostcli rollback --original
```

---

## Configuration (`~/.hostcli/hosts.yaml`)

```yaml
version: 1
groups:
  - name: system
    entries:
      - id: 01HQX100000000000000000001
        ip: 127.0.0.1
        hostname: localhost
        aliases: []
        enabled: true
        comment: "Default loopback"
  - name: work/dev
    entries:
      - id: 01HQX200000000000000000001
        ip: 192.168.1.50
        hostname: api.dev.local
        aliases: [api, gateway]
        enabled: true
        comment: "Dev API Gateway"
```

---

## License

MIT License. Copyright (c) 2026 orekasep.
