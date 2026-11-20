# Installation

AgeForge is a single binary with no runtime dependencies. It opens in its own window: F11 or Alt+Enter toggles fullscreen (Cmd+Ctrl+F on a Mac), and Ctrl+= and Ctrl+- (Cmd on a Mac) change the text size, Ctrl+0 resets it. The Settings page picks a pixel or a clean font. Building from source on Linux needs the X11 and OpenGL development packages (`libgl1-mesa-dev xorg-dev`). Download the one for your system from the latest GitHub release, or build it from source.

> Version note: the downloads below arrive with version 4.0, which has no release yet. Until then, build from source.

---

## macOS

[Download AgeForge for macOS](https://github.com/espresso20/ageforge/releases/latest/download/AgeForge-macos.zip). It is one app for Apple Silicon and Intel Macs.

1. Unzip it. Safari does this for you.
2. Drag AgeForge to Applications.
3. Open it.

The app is not notarized by Apple, so the first time macOS stops it:

- **macOS 15 or later:** open it once and click Done on the warning. Then go to System Settings → Privacy & Security, scroll to Security, click Open Anyway next to AgeForge and confirm.
- **macOS 14 or earlier:** Control-click (or right-click) AgeForge in Applications, choose Open, then Open again.

After that it opens normally. To clear the block in Terminal instead:

```bash
xattr -dr com.apple.quarantine /Applications/AgeForge.app
```

---

## Windows

[Download AgeForge for Windows](https://github.com/espresso20/ageforge/releases/latest/download/AgeForge-windows.zip).

1. Right-click the zip and choose Extract All.
2. Open the `AgeForge` folder and run `AgeForge.exe`.

The game is not code-signed, so Windows may show "Windows protected your PC". Choose More info, then Run anyway. It asks once. Do not run it from inside the zip.

---

## Linux

[Download AgeForge for Linux](https://github.com/espresso20/ageforge/releases/latest/download/AgeForge-linux-amd64.tar.gz) (x86_64).

```bash
tar xzf AgeForge-linux-amd64.tar.gz
./AgeForge/ageforge
```

It needs working OpenGL drivers.

---

Each release also has `SHA256SUMS.txt`, so you can check a download: `shasum -a 256 -c SHA256SUMS.txt` on a Mac, or `sha256sum -c SHA256SUMS.txt` on Linux.

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

## Window size

AgeForge draws a text interface in its own window, on a grid of character cells.

- A window of 120 columns by 40 rows or larger draws the full layout. A smaller one gets a more compact layout, and the dashboard fits any window from 80 columns by 24 rows up. Changing the text size changes how many columns and rows fit.
- The map draws with `unicode` glyphs by default. `map glyphs ascii` draws the map in plain characters.

---

## Updating

Download the new version and replace the old app or folder. Your accounts and saves are not in the download. They live in your user folder and stay put:

- macOS: `~/Library/Application Support/AgeForge`
- Windows: `%APPDATA%\AgeForge`
- Linux: `~/.config/AgeForge`

A build from source keeps them in `data/` beside the binary. Update it with `git pull` and `go build -o ageforge .` again.

---

## Save files

Release builds keep the game's data (accounts, saves and backups) in your user folder, as listed under Updating. A build from source keeps it in a `data/` folder next to the `ageforge` binary, so keep the two together when you move them. See [Saving & Loading](saving-and-loading.md) for the layout.

The game autosaves every 60 seconds. `Esc` closes the open panel. With no panel open, it saves, stops the game and returns to the main menu, where Continue opens that game again.
