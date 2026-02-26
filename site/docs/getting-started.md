# Installation

AgeForge is a single static binary with no runtime dependencies.

---

## macOS

```bash
# Apple Silicon (M1/M2/M3)
curl -L https://github.com/espresso20/ageforge/releases/latest/download/ageforge-darwin-arm64 -o ageforge

# Intel Mac
curl -L https://github.com/espresso20/ageforge/releases/latest/download/ageforge-darwin-amd64 -o ageforge

chmod +x ageforge
./ageforge
```

> **macOS Gatekeeper:** If macOS blocks the binary, right-click it in Finder and choose **Open**, or run `xattr -d com.apple.quarantine ./ageforge`.

---

## Linux

```bash
# x86_64
curl -L https://github.com/espresso20/ageforge/releases/latest/download/ageforge-linux-amd64 -o ageforge

# ARM64
curl -L https://github.com/espresso20/ageforge/releases/latest/download/ageforge-linux-arm64 -o ageforge

chmod +x ageforge
./ageforge
```

---

## Windows

Download `ageforge-windows-amd64.exe` from the [GitHub Releases](https://github.com/espresso20/ageforge/releases/latest) page and run it in **Windows Terminal**.

> ⚠ Windows Terminal is strongly recommended. The classic `cmd.exe` does not support the ANSI escape codes AgeForge uses for colours and cursor positioning.

---

## Build from Source

Requires **Go 1.22+**.

```bash
git clone https://github.com/espresso20/ageforge.git
cd ageforge
go build -o ageforge .
./ageforge
```

---

## Terminal requirements

AgeForge draws a full-screen TUI. For the best experience:

- Terminal width of **at least 130 columns** (the game caps its content at 130)
- ANSI 24-bit colour support (all modern terminals)
- A **monospace font** — JetBrains Mono, Cascadia Code, or Fira Code are great choices

---

## Save files

Saves are stored at:

| Platform | Path |
|---|---|
| macOS / Linux | `~/.local/share/ageforge/` |
| Windows | `%APPDATA%\ageforge\` |

Press `Esc` at any time to save, or the game auto-saves every 60 seconds.
