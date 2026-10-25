# Installation

AgeForge is a single binary with no runtime dependencies. Download the one for your system from the latest GitHub release, or build it from source.

> Version note: the downloads below are version 3.6, the latest release. This wiki describes version 4.0, which has no release yet. Version 3.6 has the `speed` command and the old city map. To play version 4.0, build from source.

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

If you downloaded the binary in a browser and macOS blocks it, right-click it in Finder and choose Open, or run `xattr -d com.apple.quarantine ./ageforge`.

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

Download `ageforge-windows-amd64.exe` from the [GitHub Releases](https://github.com/espresso20/ageforge/releases/latest) page and run it in Windows Terminal. In PowerShell, one line does both:

```powershell
iwr https://github.com/espresso20/ageforge/releases/latest/download/ageforge-windows-amd64.exe -OutFile ageforge.exe; .\ageforge.exe
```

---

## Build from source

You need Go 1.26.10 or newer.

```bash
git clone https://github.com/espresso20/ageforge.git
cd ageforge
go build -o ageforge .
./ageforge
```

`./ageforge --version` prints the version. A build from source reports `dev`.

---

## Terminal requirements

AgeForge draws a full-screen text interface.

- A terminal of 120 columns by 40 rows or larger draws the full layout. A smaller terminal gets a more compact one, and the dashboard fits any terminal from 80 columns by 24 rows up.
- The themes use full RGB color, so a terminal with 24-bit (truecolor) color shows them as designed.
- Use a monospace font. The map draws with `unicode` glyphs by default. The `nerd` glyph set needs a Nerd Font: the `icons` command checks your font and can install JetBrains Mono Nerd Font for you. `map glyphs ascii` draws the map in plain characters.

---

## Updating

The [main menu](how-to-play.md#the-main-menu) has Check for updates (`u`). It asks GitHub for the latest release and, if it is newer than yours, offers to download and install it. The download is checked against the checksums published with the release. On macOS and Linux it replaces the `ageforge` binary in place, and the new version runs the next time you start the game. On Windows it saves the new `.exe` beside the old one for you to swap in.

Release builds also check in the background when the main menu opens. If a newer release exists, `update available (u)` appears beside the version number.

A build from source is a development build. It skips the background check, and Check for updates says the check is not available. Update it with `git pull` and `go build -o ageforge .` again.

Updating replaces only the binary. Your `data/` folder, with your accounts and saves, stays as it is.

---

## Save files

The game keeps its data (accounts, saves and backups) in a `data/` folder next to the `ageforge` binary, on every platform. Keep the binary and its `data/` folder together when you move them. See [Saving & Loading](saving-and-loading.md) for the layout.

The game autosaves every 60 seconds. `Esc` closes the open panel. With no panel open, it saves, stops the game and returns to the main menu, where Continue opens that game again.
