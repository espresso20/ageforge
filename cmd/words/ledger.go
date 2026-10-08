package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ledger.go keeps count of what has been rewritten. The ledger is a small
// file in the repository: for each row the import has applied, the row's id
// and a hash of the text it was given. status compares it with the source,
// so it needs no sheets.

// ledgerPath is where the ledger lives, from the repository root.
const ledgerPath = "words/ledger.json"

// hashText is the hash the ledger stores for a row's text.
func hashText(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:8])
}

func loadLedger(root string) (map[string]string, error) {
	led := map[string]string{}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(ledgerPath)))
	if os.IsNotExist(err) {
		return led, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &led); err != nil {
		return nil, fmt.Errorf("%s is not readable: %w", ledgerPath, err)
	}
	return led, nil
}

func saveLedger(root string, led map[string]string) error {
	data, err := json.MarshalIndent(led, "", "  ") // keys come out sorted
	if err != nil {
		return err
	}
	path := filepath.Join(root, filepath.FromSlash(ledgerPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// tally is the count for one sheet.
type tally struct {
	lines, words         int
	doneLines, doneWords int // rewritten, and still reading as rewritten
	changed              int // rewritten once, and different in the source now
}

// count works out the tallies by sheet from the source and the ledger.
func count(c *catalog, led map[string]string) (by map[*area]*tally, total tally, gone int) {
	by = map[*area]*tally{}
	for _, a := range areas {
		by[a] = &tally{}
	}
	seen := map[string]bool{}
	for _, r := range c.rows {
		t := by[r.area]
		w := countWords(r.current)
		t.lines++
		t.words += w
		if h, ok := led[r.id]; ok {
			seen[r.id] = true
			if h == hashText(r.current) {
				t.doneLines++
				t.doneWords += w
			} else {
				t.changed++
			}
		}
	}
	for _, a := range areas {
		t := by[a]
		total.lines += t.lines
		total.words += t.words
		total.doneLines += t.doneLines
		total.doneWords += t.doneWords
		total.changed += t.changed
	}
	for id := range led {
		if !seen[id] {
			gone++
		}
	}
	return by, total, gone
}

// status prints how much is rewritten and how much is left.
func status(root string, out io.Writer) error {
	c, err := read(root)
	if err != nil {
		return err
	}
	led, err := loadLedger(root)
	if err != nil {
		return err
	}
	by, total, gone := count(c, led)
	fmt.Fprintf(out, "%-32s %7s %8s   %9s %8s   %9s %8s   %s\n", "sheet", "lines", "words", "rewritten", "words", "remaining", "words", "changed since")
	line := func(name string, t tally) {
		fmt.Fprintf(out, "%-32s %7d %8d   %9d %8d   %9d %8d   %d\n", name, t.lines, t.words, t.doneLines, t.doneWords, t.lines-t.doneLines, t.words-t.doneWords, t.changed)
	}
	for _, a := range areas {
		line(a.file, *by[a])
	}
	line("total", total)
	if total.changed > 0 {
		fmt.Fprintf(out, "\n%d rewritten lines read differently in the source now than when they were imported (changed since). They count as remaining.\n", total.changed)
	}
	if gone > 0 {
		fmt.Fprintf(out, "%d rewritten lines are no longer in the source at the id they had.\n", gone)
	}
	return nil
}
