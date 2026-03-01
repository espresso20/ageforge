package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	updaterAPIURL     = "https://api.github.com/repos/espresso20/ageforge/releases/latest"
	updaterReleaseURL = "https://github.com/espresso20/ageforge/releases/latest/download"
)

// UpdateResult holds the outcome of a version check.
type UpdateResult struct {
	CurrentVersion string
	LatestVersion  string
	IsNewer        bool
	BinaryName     string // platform-specific filename, e.g. ageforge-macos-arm64
}

// CheckLatest fetches the latest GitHub release tag and compares it to currentVersion.
func CheckLatest(currentVersion string) (UpdateResult, error) {
	if currentVersion == "dev" {
		return UpdateResult{}, fmt.Errorf("dev build — update check not available")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", updaterAPIURL, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ageforge-updater")

	resp, err := client.Do(req)
	if err != nil {
		return UpdateResult{}, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return UpdateResult{}, fmt.Errorf("GitHub returned HTTP %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return UpdateResult{}, fmt.Errorf("parse error: %w", err)
	}

	latest := release.TagName
	return UpdateResult{
		CurrentVersion: currentVersion,
		LatestVersion:  latest,
		IsNewer:        semverGreater(latest, currentVersion),
		BinaryName:     updaterPlatformBinary(),
	}, nil
}

// DownloadAndInstall downloads the new binary, verifies its SHA256 checksum, and
// atomically replaces the running binary (Unix) or saves it alongside (Windows).
// Returns a human-readable success message or an error.
func DownloadAndInstall(result UpdateResult) (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot locate executable: %w", err)
	}

	downloadURL := updaterReleaseURL + "/" + result.BinaryName
	checksumURL := updaterReleaseURL + "/SHA256SUMS.txt"

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Stream to temp file
	tmp, err := os.CreateTemp("", "ageforge-update-*")
	if err != nil {
		return "", fmt.Errorf("cannot create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("write error: %w", err)
	}
	tmp.Close()

	// Verify SHA256 against published checksums
	if err := updaterVerifyChecksum(tmpPath, checksumURL, result.BinaryName); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("checksum mismatch: %w", err)
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("chmod failed: %w", err)
	}

	// Windows: cannot overwrite a running .exe — save alongside instead
	if runtime.GOOS == "windows" {
		newPath := strings.TrimSuffix(exePath, ".exe") + "-" + result.LatestVersion + ".exe"
		if err := os.Rename(tmpPath, newPath); err != nil {
			os.Remove(tmpPath)
			return "", fmt.Errorf("save failed: %w", err)
		}
		return fmt.Sprintf(
			"Saved as:\n%s\n\nReplace the current .exe to apply the update.",
			newPath,
		), nil
	}

	// Unix: atomic rename swap with backup
	backupPath := exePath + ".old"
	_ = os.Rename(exePath, backupPath)
	if err := os.Rename(tmpPath, exePath); err != nil {
		_ = os.Rename(backupPath, exePath) // restore on failure
		os.Remove(tmpPath)
		return "", fmt.Errorf("install failed: %w", err)
	}
	_ = os.Remove(backupPath)

	return fmt.Sprintf("Updated to %s\n\nRestart the game to play the new version.", result.LatestVersion), nil
}

// updaterPlatformBinary returns the release asset filename for the current OS/arch.
func updaterPlatformBinary() string {
	osName := runtime.GOOS
	if osName == "darwin" {
		osName = "macos"
	}
	name := fmt.Sprintf("ageforge-%s-%s", osName, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

// semverGreater reports whether semver string a is strictly greater than b.
func semverGreater(a, b string) bool {
	parse := func(s string) [3]int {
		s = strings.TrimPrefix(s, "v")
		parts := strings.SplitN(s, ".", 3)
		for len(parts) < 3 {
			parts = append(parts, "0")
		}
		var v [3]int
		for i := 0; i < 3; i++ {
			v[i], _ = strconv.Atoi(parts[i])
		}
		return v
	}
	av, bv := parse(a), parse(b)
	for i := 0; i < 3; i++ {
		if av[i] != bv[i] {
			return av[i] > bv[i]
		}
	}
	return false
}

// updaterVerifyChecksum fetches SHA256SUMS.txt and verifies filePath against it.
// Silently passes if the checksum file is unavailable or the binary isn't listed.
func updaterVerifyChecksum(filePath, checksumURL, binaryName string) error {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(checksumURL)
	if err != nil {
		return nil // non-fatal: skip if unavailable
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	expected := ""
	for _, line := range strings.Split(string(body), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == binaryName {
			expected = parts[0]
			break
		}
	}
	if expected == "" {
		return nil // no entry for this binary, skip
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := hex.EncodeToString(h.Sum(nil))
	if actual != expected {
		return fmt.Errorf("expected %.8s…, got %.8s…", expected, actual)
	}
	return nil
}
