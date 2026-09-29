// Package nerdfont installs JetBrains Mono Nerd Font for the current user,
// for the map's Nerd Font glyph tier. It downloads one pinned release zip
// from the official ryanoasis/nerd-fonts GitHub releases, checks its
// SHA-256, and copies the Mono variant's TTFs into the per-user font
// folder. It needs no admin rights, runs nothing but fc-cache on Linux,
// and never touches a terminal's configuration.
package nerdfont

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// The pinned release. Bump Version, URL and SHA256 together; the checksum
// is of the zip GitHub serves for that tag (it matches the digest GitHub
// lists for the asset).
const (
	Version = "v3.5.1"
	URL     = "https://github.com/ryanoasis/nerd-fonts/releases/download/" + Version + "/JetBrainsMono.zip"
	SHA256  = "fab782a66f7d3019da64f6572db9fc5d3a4bcb19f9fa13e2d8a62e3693d6396e"
	// FamilyName is the font family a terminal lists after the install.
	FamilyName = "JetBrainsMono Nerd Font Mono"
	// monoPrefix picks the Mono variant (single-width icons, what a
	// terminal grid wants) and leaves out the NL (no ligatures) and the
	// proportional-icon variants.
	monoPrefix = "JetBrainsMonoNerdFontMono-"
	// maxZip bounds the download (the zip is about 134 MB).
	maxZip = 400 << 20
)

// ErrChecksum is returned when the download is not the pinned file.
var ErrChecksum = errors.New("the download's checksum does not match the pinned release")

// Result is what an install did.
type Result struct {
	Dir   string   // the font folder
	Files []string // the TTFs written, full paths
	// Registered: Windows per-user registry entries were written.
	Registered bool
	// ManualStep is a step the player still has to take ("" when none).
	ManualStep string
	// FontCache: fc-cache ran (Linux).
	FontCache bool
}

// Installer installs the font. The zero value is not usable; New fills in
// the real environment, and tests replace any field.
type Installer struct {
	Client *http.Client
	URL    string
	SHA256 string
	GOOS   string
	// Home is the user's home folder; LocalAppData is %LOCALAPPDATA% on
	// Windows (empty elsewhere); DataHome is $XDG_DATA_HOME on Linux.
	Home, LocalAppData, DataHome string
	// TempDir holds the download while it is checked ("" = os.TempDir).
	TempDir string
	// FontCache refreshes the Linux font cache; nil skips it. New sets it
	// to run fc-cache -f when fc-cache is on the PATH.
	FontCache func(ctx context.Context) error
	// Register writes the Windows per-user font registry entries; nil
	// means the build has no registry support and the player gets a
	// manual step instead.
	Register func(files []string) error
	// Progress reports progress lines ("Downloading... 40%"); may be nil.
	Progress func(string)
}

// New returns an installer for the real environment.
func New() *Installer {
	home, _ := os.UserHomeDir()
	in := &Installer{
		Client:       &http.Client{Timeout: 10 * time.Minute},
		URL:          URL,
		SHA256:       SHA256,
		GOOS:         runtime.GOOS,
		Home:         home,
		LocalAppData: os.Getenv("LOCALAPPDATA"),
		DataHome:     os.Getenv("XDG_DATA_HOME"),
		Register:     registerFonts, // nil off Windows (registry_other.go)
	}
	if p, err := exec.LookPath("fc-cache"); err == nil {
		in.FontCache = func(ctx context.Context) error {
			return exec.CommandContext(ctx, p, "-f").Run()
		}
	}
	return in
}

// FontDir is the per-user font folder for the installer's OS.
func (in *Installer) FontDir() (string, error) {
	switch in.GOOS {
	case "darwin":
		if in.Home == "" {
			return "", errors.New("no home folder")
		}
		return filepath.Join(in.Home, "Library", "Fonts"), nil
	case "windows":
		base := in.LocalAppData
		if base == "" {
			if in.Home == "" {
				return "", errors.New("no LOCALAPPDATA folder")
			}
			base = filepath.Join(in.Home, "AppData", "Local")
		}
		return filepath.Join(base, "Microsoft", "Windows", "Fonts"), nil
	default: // linux and the BSDs
		if in.DataHome != "" {
			return filepath.Join(in.DataHome, "fonts"), nil
		}
		if in.Home == "" {
			return "", errors.New("no home folder")
		}
		return filepath.Join(in.Home, ".local", "share", "fonts"), nil
	}
}

func (in *Installer) say(format string, args ...any) {
	if in.Progress != nil {
		in.Progress(fmt.Sprintf(format, args...))
	}
}

// Install downloads, checks and installs the font.
func (in *Installer) Install(ctx context.Context) (Result, error) {
	dir, err := in.FontDir()
	if err != nil {
		return Result{}, err
	}
	zipPath, err := in.download(ctx)
	if zipPath != "" {
		defer os.Remove(zipPath)
	}
	if err != nil {
		return Result{}, err
	}
	res := Result{Dir: dir}
	res.Files, err = extractMono(zipPath, dir)
	if err != nil {
		return res, err
	}
	in.say("Installed %d font files in %s.", len(res.Files), dir)
	switch in.GOOS {
	case "windows":
		if in.Register == nil {
			res.ManualStep = "Open " + dir + ", select the JetBrainsMonoNerdFontMono files, right-click and choose Install."
		} else if err := in.Register(res.Files); err != nil {
			res.ManualStep = "Windows did not take the font registration (" + err.Error() + "). Open " + dir +
				", select the JetBrainsMonoNerdFontMono files, right-click and choose Install."
		} else {
			res.Registered = true
		}
	case "darwin":
	default:
		if in.FontCache != nil {
			if err := in.FontCache(ctx); err == nil {
				res.FontCache = true
			}
		}
	}
	return res, nil
}

// download fetches the zip into a temp file, hashing as it goes, and
// returns the file only when the hash matches.
func (in *Installer) download(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, in.URL, nil)
	if err != nil {
		return "", err
	}
	client := in.Client
	if client == nil {
		client = http.DefaultClient
	}
	in.say("Downloading JetBrains Mono Nerd Font %s from GitHub...", Version)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("the download could not start (are you online?): %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("the download failed: GitHub answered %s", resp.Status)
	}
	f, err := os.CreateTemp(in.TempDir, "ageforge-nerdfont-*.zip")
	if err != nil {
		return "", err
	}
	name := f.Name()
	h := sha256.New()
	pw := &progressWriter{total: resp.ContentLength, say: in.say}
	n, err := io.Copy(io.MultiWriter(f, h, pw), io.LimitReader(resp.Body, maxZip+1))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return name, fmt.Errorf("the download was cut off: %w", err)
	}
	if n > maxZip {
		return name, errors.New("the download is larger than the pinned release")
	}
	if got := hex.EncodeToString(h.Sum(nil)); !strings.EqualFold(got, in.SHA256) {
		return name, fmt.Errorf("%w (got %s)", ErrChecksum, got)
	}
	in.say("Download checked (SHA-256 matches %s).", Version)
	return name, nil
}

// progressWriter reports each quarter of the download.
type progressWriter struct {
	total, done int64
	next        int64 // next percentage to report
	say         func(string, ...any)
}

func (p *progressWriter) Write(b []byte) (int, error) {
	p.done += int64(len(b))
	if p.total > 0 {
		if p.next == 0 {
			p.next = 25
		}
		for p.next < 100 && p.done*100 >= p.total*p.next {
			p.say("Downloading... %d%%", p.next)
			p.next += 25
		}
	}
	return len(b), nil
}

// extractMono copies the Mono variant's TTFs out of the zip into dir.
func extractMono(zipPath, dir string) ([]string, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("the download is not a readable zip: %w", err)
	}
	defer zr.Close()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("could not create the font folder: %w", err)
	}
	var out []string
	for _, zf := range zr.File {
		base := path.Base(zf.Name)
		if !strings.HasPrefix(base, monoPrefix) || !strings.HasSuffix(base, ".ttf") || zf.FileInfo().IsDir() {
			continue
		}
		dst := filepath.Join(dir, base) // base only: nothing in the zip picks the folder
		if err := copyZipFile(zf, dst); err != nil {
			return out, err
		}
		out = append(out, dst)
	}
	if len(out) == 0 {
		return nil, errors.New("the zip has no " + monoPrefix + " fonts")
	}
	return out, nil
}

// fontValueName is the registry value name for a font file:
// "JetBrainsMonoNerdFontMono-BoldItalic.ttf" is
// "JetBrainsMono Nerd Font Mono Bold Italic (TrueType)".
func fontValueName(file string) string {
	style := strings.TrimSuffix(strings.TrimPrefix(file, monoPrefix), ".ttf")
	var sb strings.Builder
	for i, r := range style {
		if i > 0 && r >= 'A' && r <= 'Z' {
			sb.WriteByte(' ')
		}
		sb.WriteRune(r)
	}
	return FamilyName + " " + sb.String() + " (TrueType)"
}

func copyZipFile(zf *zip.File, dst string) error {
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	tmp := dst + ".part"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("could not write %s: %w", filepath.Base(dst), err)
	}
	_, err = io.Copy(f, io.LimitReader(rc, 64<<20))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return fmt.Errorf("could not write %s: %w", filepath.Base(dst), err)
	}
	return os.Rename(tmp, dst)
}
