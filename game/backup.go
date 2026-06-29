package game

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// backupsDirName is the subdirectory under the data ROOT that holds full account snapshots:
// <root>/backups/<name>-<id8>-<timestamp>/. It lives one level above the per-account slots
// (alongside <root>/accounts), so a wipe of <root>/accounts/<id>/ never touches the backups —
// the recoverable copy survives the very deletion it was taken to guard against.
const backupsDirName = "backups"

// backupTimestampLayout is the directory-name timestamp: zero-padded YYYYMMDD-HHMMSS, sortable
// lexicographically (later strings sort after earlier ones), filesystem-safe, and human-legible.
// Same-second collisions are disambiguated by a nanosecond suffix in BackupAccount.
const backupTimestampLayout = "20060102-150405"

// backupRetention is how many snapshots to keep PER ACCOUNT. After each backup the older ones
// for that same account (matched by its id8 infix) are pruned; other accounts' backups are
// never considered. Ten is enough to recover from a few bad wipes/exports without unbounded
// disk growth.
const backupRetention = 10

// BackupAccount snapshots the on-disk slot for account id into a fresh directory under
// <root>/backups/, returning the absolute path of the new backup dir. The snapshot is a COPY
// (the original slot is left untouched): account.json plus a recursive copy of the slot's
// saves/ subtree when it exists. It is the full-fidelity counterpart to ExportAccountByID
// (which serializes only the account's meta-progression, not the saves).
//
// The backup dir is named "<cleanname>-<id8>-<timestamp>": the FS-safe display name (resolved
// read-only from the slot's account.json, falling back to "account" when blank), the first 8
// id chars, and a YYYYMMDD-HHMMSS stamp. If a same-second backup already exists the stamp is
// suffixed with nanoseconds so two snapshots in the same second never collide.
//
// CRITICAL — it MUST NOT mutate the active account. The display name is resolved via
// loadAccountFromSlot (which keys off id, never calls setActiveAccountID), so backing up slot
// B can never repoint the active account at B. As belt-and-braces it snapshots
// getActiveAccountID() up front and restores it on return.
//
// Errors when the slot has no account.json ("no such account: <id>") — there is nothing to
// snapshot. After a successful copy it prunes this account's older backups down to
// backupRetention (best-effort: a prune failure does not fail the backup, since the snapshot
// itself already succeeded).
func BackupAccount(id string) (string, error) {
	// Belt-and-braces: assert (and restore) the active account around the whole operation. The
	// load below uses loadAccountFromSlot, which never touches the active id, so this is a
	// no-op guard that documents and future-proofs the invariant.
	activeBefore := getActiveAccountID()
	defer setActiveAccountID(activeBefore)

	slot := accountDir(id)
	srcAccount := filepath.Join(slot, accountFileName)
	if _, err := os.Stat(srcAccount); err != nil {
		// No account.json in the slot → nothing to snapshot (absent or unreadable).
		return "", fmt.Errorf("no such account: %s", id)
	}

	// Resolve the display name READ-ONLY for the dir name; fall back to "account" when the slot
	// is unparseable or the name is blank. cleanSaveName trims + collapses internal whitespace;
	// the slugger below strips any remaining path-hostile characters.
	name := "account"
	if acct, found, _ := loadAccountFromSlot(id); found && acct != nil {
		if cleaned := backupNameSlug(acct.DisplayName); cleaned != "" {
			name = cleaned
		}
	}

	id8 := id
	if len(id8) > 8 {
		id8 = id8[:8]
	}

	backupsRoot := filepath.Join(rootDataDir(), backupsDirName)
	if err := os.MkdirAll(backupsRoot, 0755); err != nil {
		return "", fmt.Errorf("failed to create backups root: %w", err)
	}

	// Build the dest dir, disambiguating a same-second collision with nanoseconds so two
	// backups taken inside one second never clobber each other.
	now := time.Now()
	dst := filepath.Join(backupsRoot, fmt.Sprintf("%s-%s-%s", name, id8, now.Format(backupTimestampLayout)))
	if _, err := os.Stat(dst); err == nil {
		dst = filepath.Join(backupsRoot, fmt.Sprintf("%s-%s-%s-%09d", name, id8, now.Format(backupTimestampLayout), now.Nanosecond()))
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup dir: %w", err)
	}

	// Copy account.json.
	if err := copyFile(srcAccount, filepath.Join(dst, accountFileName)); err != nil {
		return "", fmt.Errorf("failed to back up account.json: %w", err)
	}

	// Copy the slot's saves/ subtree if present (a slot may legitimately have none yet).
	srcSaves := filepath.Join(slot, "saves")
	if info, err := os.Stat(srcSaves); err == nil && info.IsDir() {
		if err := copyDir(srcSaves, filepath.Join(dst, "saves")); err != nil {
			return "", fmt.Errorf("failed to back up saves: %w", err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to stat saves for backup: %w", err)
	}

	// Prune older backups for THIS account. Best-effort: the snapshot already succeeded, so a
	// prune hiccup must not turn a good backup into an error.
	pruneAccountBackups(backupsRoot, id8)

	return dst, nil
}

// snapshotPreMigration takes a one-time, full COPY of the CURRENT flat data ROOT into a fresh
// dir <root>/backups/pre-migration-<YYYYMMDD-HHMMSS>/, returning its absolute path. It is the
// pre-Phase-A safety net: migrateLegacyAccountIfNeeded calls it right before it relocates the
// flat layout (account.json + saves/) into a slot, so the player's original on-disk state is
// recoverable even though the migration itself only os.Renames (atomic, non-destructive).
//
// It copies every TOP-LEVEL entry under the root EXCEPT "backups" and "accounts": "backups" is
// skipped so the walk never recurses into the snapshot target it is creating, and "accounts" is
// skipped because at pre-migration time it does not exist yet (the migration trigger requires
// its absence) — guarding the name keeps the snapshot honest if that ever changes. Everything
// else is captured: account.json, the saves/ tree, and any stray files (e.g. a user-made
// account-export.json) — i.e. the whole flat data/ that is about to move.
//
// It mirrors BackupAccount's timestamp discipline (backupTimestampLayout + a nanosecond suffix
// on a same-second collision) so two runs can never clobber one another. A missing or empty
// root is not an error — there is simply nothing to snapshot — so it returns ("", nil).
//
// Like BackupAccount it MUST NOT mutate the active account: it is pure file I/O keyed off the
// root, touching no pointer and calling no setActiveAccountID.
func snapshotPreMigration() (string, error) {
	root := rootDataDir()
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // no data root yet — nothing to snapshot
		}
		return "", fmt.Errorf("failed to read data root for pre-migration snapshot: %w", err)
	}

	// Gather the top-level entries to snapshot, skipping the two reserved names. Doing this
	// first lets us bail out (empty root, or only reserved entries) without creating an empty
	// snapshot dir.
	type rootEntry struct {
		name  string
		isDir bool
	}
	var toCopy []rootEntry
	for _, e := range entries {
		name := e.Name()
		if name == backupsDirName || name == accountsDirName {
			continue // never recurse into the snapshot target; accounts/ shouldn't exist yet
		}
		toCopy = append(toCopy, rootEntry{name: name, isDir: e.IsDir()})
	}
	if len(toCopy) == 0 {
		return "", nil // empty (or only reserved entries) — nothing to snapshot
	}

	backupsRoot := filepath.Join(root, backupsDirName)
	if err := os.MkdirAll(backupsRoot, 0755); err != nil {
		return "", fmt.Errorf("failed to create backups root: %w", err)
	}

	// Build the dest dir, disambiguating a same-second collision with nanoseconds so two
	// pre-migration snapshots taken inside one second never clobber each other.
	now := time.Now()
	dst := filepath.Join(backupsRoot, "pre-migration-"+now.Format(backupTimestampLayout))
	if _, err := os.Stat(dst); err == nil {
		dst = filepath.Join(backupsRoot, fmt.Sprintf("pre-migration-%s-%09d", now.Format(backupTimestampLayout), now.Nanosecond()))
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return "", fmt.Errorf("failed to create pre-migration snapshot dir: %w", err)
	}

	for _, e := range toCopy {
		src := filepath.Join(root, e.name)
		target := filepath.Join(dst, e.name)
		if e.isDir {
			if err := copyDir(src, target); err != nil {
				return "", fmt.Errorf("failed to snapshot %q before migration: %w", e.name, err)
			}
		} else if err := copyFile(src, target); err != nil {
			return "", fmt.Errorf("failed to snapshot %q before migration: %w", e.name, err)
		}
	}
	return dst, nil
}

// BackupAccount is the engine passthrough to the package-level game.BackupAccount: a full
// on-disk snapshot of the slot for account id. No ge.mu and no engine state touched — it is
// pure file I/O over the slots and the active account is never disturbed.
func (ge *GameEngine) BackupAccount(id string) (string, error) {
	return BackupAccount(id)
}

// backupNameSlug FS-safe-cleans a display name for use in a backup dir name. It first reuses
// cleanSaveName (trim + collapse internal whitespace), then replaces any character that is not
// a letter, digit, dash, or underscore with a dash, and finally collapses the spaces it kept
// into single dashes. The result never contains a path separator or other shell/FS-hostile
// character, so the backup dir name is always safe. Returns "" when nothing usable remains.
func backupNameSlug(name string) string {
	cleaned := cleanSaveName(name)
	if cleaned == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range cleaned {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		default:
			b.WriteByte('-')
		}
	}
	// Collapse runs of dashes and trim leading/trailing dashes for a tidy slug.
	slug := b.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return strings.Trim(slug, "-")
}

// pruneAccountBackups keeps only the newest backupRetention backup dirs whose name carries the
// "-<id8>-" infix for THIS account, os.RemoveAll-ing the rest. It matches by the id8 segment so
// it is robust to the cleanname prefix and never touches a DIFFERENT account's backups. Dirs
// are ordered by name (the trailing YYYYMMDD-HHMMSS[-nanos] stamp sorts chronologically), with
// ModTime as a tiebreaker, so "newest" is well-defined. Best-effort: errors are swallowed (a
// failed prune must not fail an already-successful backup).
func pruneAccountBackups(backupsRoot, id8 string) {
	entries, err := os.ReadDir(backupsRoot)
	if err != nil {
		return
	}
	infix := "-" + id8 + "-"
	type backupEntry struct {
		name    string
		modTime time.Time
	}
	var mine []backupEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// Match this account's snapshots by the id8 infix; the cleanname prefix varies and the
		// timestamp suffix follows, so the infix is the stable, account-specific key.
		if !strings.Contains(e.Name(), infix) {
			continue
		}
		info, infoErr := e.Info()
		mt := time.Time{}
		if infoErr == nil {
			mt = info.ModTime()
		}
		mine = append(mine, backupEntry{name: e.Name(), modTime: mt})
	}
	if len(mine) <= backupRetention {
		return
	}
	// Newest first: name descending (the trailing timestamp sorts chronologically), ModTime as
	// a tiebreaker for same-named edge cases.
	sort.Slice(mine, func(i, j int) bool {
		if mine[i].name != mine[j].name {
			return mine[i].name > mine[j].name
		}
		return mine[i].modTime.After(mine[j].modTime)
	})
	for _, old := range mine[backupRetention:] {
		_ = os.RemoveAll(filepath.Join(backupsRoot, old.name))
	}
}

// copyFile copies the file at src to dst, creating dst's parent dirs as needed and preserving
// the source's mode (defaulting to ~0644 when the source mode can't be read). It streams via
// io.Copy so large save files don't balloon memory.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	mode := os.FileMode(0644)
	if info, statErr := in.Stat(); statErr == nil {
		mode = info.Mode().Perm()
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// copyDir recursively copies the directory tree at src into dst, recreating the directory
// structure and copying every regular file via copyFile (which preserves modes). Symlinks and
// other irregular entries are skipped — a saves/ tree is plain files and dirs. It walks src
// with filepath.WalkDir and mirrors each entry under dst.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(src, path)
		if relErr != nil {
			return relErr
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if !d.Type().IsRegular() {
			return nil // skip symlinks / devices / sockets — saves are plain files
		}
		return copyFile(path, target)
	})
}
