# Installation

AgeForge is a single static binary with no runtime dependencies.

---

## macOS

```bash
# Apple Silicon (M1/M2/M3)
curl -L https://github.com/espresso20/ageforge/releases/latest/download/ageforge-macos-arm64 -o ageforge

# Intel Mac
curl -L https://github.com/espresso20/ageforge/releases/latest/download/ageforge-macos-amd64 -o ageforge

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

> ⚠ Windows Terminal is strongly recommended. The classic `cmd.exe` does not support the ANSI escape codes AgeForge uses for colors and cursor positioning.

---

## Build from source

Requires **Go 1.24+**.

```bash
git clone https://github.com/espresso20/ageforge.git
cd ageforge
go build -o ageforge .
./ageforge
```

---

## Terminal requirements

AgeForge draws a full-screen text interface. It works best with:

- a terminal **at least 130 columns** wide (the game caps its content at 130)
- 24-bit (truecolor) ANSI color, which modern terminals support
- a **monospace font**, such as JetBrains Mono, Cascadia Code or Fira Code

---

## Save files

The game keeps its data (accounts, saves and backups) in a `data/` folder next to the `ageforge` binary, on every platform. Keep the binary and its `data/` folder together when you move them. See [Saving & Loading](saving-and-loading.md) for the layout.

The game autosaves every 60 seconds. `Esc` closes the open panel; with no panel open, it saves, stops the game and returns to the main menu.
