package nerdfont

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// fakeZip is a small stand-in for the release zip: two Mono TTFs, the NL
// and proportional variants the installer must skip, a readme, and an
// entry with a path that tries to leave the font folder.
func fakeZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range map[string]string{
		"JetBrainsMonoNerdFontMono-Regular.ttf":    "regular",
		"JetBrainsMonoNerdFontMono-BoldItalic.ttf": "bold italic",
		"JetBrainsMonoNLNerdFontMono-Regular.ttf":  "no ligatures",
		"JetBrainsMonoNerdFont-Regular.ttf":        "proportional icons",
		"README.md":                                "readme",
		"../../JetBrainsMonoNerdFontMono-Evil.ttf": "escapes",
	} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(body))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// fakeInstaller points an installer at a test server serving body, with a
// temp home for goos.
func fakeInstaller(t *testing.T, goos string, body []byte, checksum string) (*Installer, string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
	}))
	t.Cleanup(srv.Close)
	home := t.TempDir()
	in := &Installer{Client: srv.Client(), URL: srv.URL + "/JetBrainsMono.zip", SHA256: checksum, GOOS: goos,
		Home: home, TempDir: t.TempDir()}
	return in, home
}

func TestInstallPerOSPaths(t *testing.T) {
	body := fakeZip(t)
	for _, c := range []struct {
		goos string
		dir  func(home string) string
	}{
		{"darwin", func(h string) string { return filepath.Join(h, "Library", "Fonts") }},
		{"linux", func(h string) string { return filepath.Join(h, ".local", "share", "fonts") }},
		{"windows", func(h string) string { return filepath.Join(h, "LocalAppData", "Microsoft", "Windows", "Fonts") }},
	} {
		t.Run(c.goos, func(t *testing.T) {
			in, home := fakeInstaller(t, c.goos, body, sum(body))
			if c.goos == "windows" {
				in.LocalAppData = filepath.Join(home, "LocalAppData")
			}
			var cached bool
			in.FontCache = func(context.Context) error { cached = true; return nil }
			var registered []string
			in.Register = func(files []string) error { registered = files; return nil }
			var progress []string
			in.Progress = func(s string) { progress = append(progress, s) }

			res, err := in.Install(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			want := c.dir(home)
			if res.Dir != want {
				t.Errorf("font folder %s, want %s", res.Dir, want)
			}
			var got []string
			for _, f := range res.Files {
				if filepath.Dir(f) != want {
					t.Errorf("%s is outside the font folder", f)
				}
				got = append(got, filepath.Base(f))
			}
			sort.Strings(got)
			if strings.Join(got, ",") != "JetBrainsMonoNerdFontMono-BoldItalic.ttf,JetBrainsMonoNerdFontMono-Evil.ttf,JetBrainsMonoNerdFontMono-Regular.ttf" {
				t.Errorf("installed %v", got)
			}
			if b, err := os.ReadFile(filepath.Join(want, "JetBrainsMonoNerdFontMono-Regular.ttf")); err != nil || string(b) != "regular" {
				t.Errorf("Regular: %q, %v", b, err)
			}
			if _, err := os.Stat(filepath.Join(home, "JetBrainsMonoNerdFontMono-Evil.ttf")); err == nil {
				t.Error("a zip path escaped the font folder")
			}
			if cached != (c.goos == "linux") {
				t.Errorf("fc-cache ran = %v on %s", cached, c.goos)
			}
			if (len(registered) > 0) != (c.goos == "windows") || res.Registered != (c.goos == "windows") {
				t.Errorf("registry entries written = %v on %s", registered, c.goos)
			}
			if len(progress) == 0 {
				t.Error("no progress lines")
			}
			if left, _ := filepath.Glob(filepath.Join(in.TempDir, "*")); len(left) > 0 {
				t.Errorf("the download was left behind: %v", left)
			}
		})
	}
}

func TestInstallWindowsWithoutRegistryGivesManualStep(t *testing.T) {
	body := fakeZip(t)
	in, home := fakeInstaller(t, "windows", body, sum(body))
	in.LocalAppData = filepath.Join(home, "AppData", "Local")
	res, err := in.Install(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Registered || !strings.Contains(res.ManualStep, "Install") {
		t.Errorf("result %+v, want a manual step", res)
	}
}

func TestInstallLinuxXDGDataHome(t *testing.T) {
	body := fakeZip(t)
	in, home := fakeInstaller(t, "linux", body, sum(body))
	in.DataHome = filepath.Join(home, "xdg")
	res, err := in.Install(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Dir != filepath.Join(home, "xdg", "fonts") {
		t.Errorf("font folder %s", res.Dir)
	}
}

func TestInstallChecksumMismatchAborts(t *testing.T) {
	body := fakeZip(t)
	in, home := fakeInstaller(t, "darwin", body, strings.Repeat("0", 64))
	_, err := in.Install(context.Background())
	if !errors.Is(err, ErrChecksum) {
		t.Fatalf("err = %v, want ErrChecksum", err)
	}
	if _, err := os.Stat(filepath.Join(home, "Library", "Fonts")); !os.IsNotExist(err) {
		t.Error("the font folder was touched after a checksum mismatch")
	}
	if left, _ := filepath.Glob(filepath.Join(in.TempDir, "*")); len(left) > 0 {
		t.Errorf("the bad download was left behind: %v", left)
	}
}

func TestInstallNoNetwork(t *testing.T) {
	// A port nothing listens on: the connection is refused at once.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close()
	home := t.TempDir()
	in := &Installer{Client: http.DefaultClient, URL: "http://" + addr + "/JetBrainsMono.zip", SHA256: SHA256,
		GOOS: "darwin", Home: home, TempDir: t.TempDir()}
	_, err = in.Install(context.Background())
	if err == nil || !strings.Contains(err.Error(), "online") {
		t.Fatalf("err = %v, want a could-not-start error", err)
	}
	if _, err := os.Stat(filepath.Join(home, "Library")); !os.IsNotExist(err) {
		t.Error("the font folder was touched with no network")
	}
}

func TestInstallHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	in := &Installer{Client: srv.Client(), URL: srv.URL, SHA256: SHA256, GOOS: "linux", Home: t.TempDir(), TempDir: t.TempDir()}
	if _, err := in.Install(context.Background()); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("err = %v, want the HTTP status", err)
	}
}

func TestPinnedRelease(t *testing.T) {
	if !strings.HasPrefix(URL, "https://github.com/ryanoasis/nerd-fonts/releases/download/"+Version+"/") {
		t.Errorf("URL %s is not the official release for %s", URL, Version)
	}
	if b, err := hex.DecodeString(SHA256); err != nil || len(b) != 32 {
		t.Errorf("SHA256 %q is not a SHA-256", SHA256)
	}
}

func TestFontValueName(t *testing.T) {
	for file, want := range map[string]string{
		"JetBrainsMonoNerdFontMono-Regular.ttf":    "JetBrainsMono Nerd Font Mono Regular (TrueType)",
		"JetBrainsMonoNerdFontMono-BoldItalic.ttf": "JetBrainsMono Nerd Font Mono Bold Italic (TrueType)",
	} {
		if got := fontValueName(file); got != want {
			t.Errorf("%s: %q, want %q", file, got, want)
		}
	}
}
