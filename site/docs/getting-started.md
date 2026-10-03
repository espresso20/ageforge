# Installation

AgeForge is a single static binary with no runtime dependencies. Download the one for your system from the latest GitHub release, or build it from source.

> **Version note:** the downloads below are the 3.6 release. This wiki describes version 4.0, which is not released yet, so 3.6 still has some older systems (for example the `speed` command and the old city map); build from source to play 4.0 now.

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

Requires **Go 1.26+**.

```bash
git clone https://github.com/espresso20/ageforge.git
cd ageforge
go build -o ageforge .
./ageforge
```

---

## Terminal requirements

AgeForge draws a full-screen text interface. It works best with:

- a terminal **at least 120 columns wide and 40 rows tall**. A smaller one works, but the dashboard's mini map hides to leave the Buildings list room
- 24-bit (truecolor) ANSI color, which modern terminals support
- a **monospace font**, such as JetBrains Mono, Cascadia Code or Fira Code

---

## Updating

The main menu has **Check for updates** (`u`). It asks GitHub for the latest release and, if it is newer than yours, offers to download and install it. The download is checked against the release's published checksums. On macOS and Linux it replaces the `ageforge` binary in place, and the new version runs the next time you start the game. On Windows it saves the new `.exe` beside the old one for you to swap in.

Release builds also check in the background when the main menu opens. If a newer release exists, **Update available (u)** appears beside the version number.

A build from source is a development build: it skips the background check, and Check for updates says it isn't available. Update it with `git pull` and `go build -o ageforge .` again.

Updating replaces only the binary: your `data/` folder, with your accounts and saves, stays as it is.

---

## Save files

The game keeps its data (accounts, saves and backups) in a `data/` folder next to the `ageforge` binary, on every platform. Keep the binary and its `data/` folder together when you move them. See [Saving & Loading](saving-and-loading.md) for the layout.

The game autosaves every 60 seconds. `Esc` closes the open panel; with no panel open, it saves, stops the game and returns to the main menu.
