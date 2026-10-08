package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// names.go is what the import does after a name is rewritten. The game
// looks things up by key, so the rename itself is complete; but the old
// name is still written in places the import does not edit: the wiki, the
// landing page and tests. It lists them. The one set of wiki tables that
// has a generator, it regenerates.

// wikiGenerator is the command that rewrites the wiki's tech tables from
// config (config/tech_wiki_test.go).
var wikiGenerator = []string{"go", "test", "./config", "-run", "TestTechWikiTables", "-count=1"}

// quoters are the places a name can be written outside the game's source.
type quoter struct {
	rel  string
	text string
	kind string // "wiki", "generated" or "test"
}

// loadQuoters reads the wiki, the landing page, the README and every test.
func loadQuoters(root string) []quoter {
	var out []quoter
	add := func(rel, kind string) {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err == nil {
			out = append(out, quoter{rel, string(data), kind})
		}
	}
	add("README.md", "wiki")
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			switch {
			case rel == ".":
			case strings.HasPrefix(d.Name(), "."), d.Name() == "node_modules", rel == "words", rel == "data", rel == "dist":
				return filepath.SkipDir
			}
			return nil
		}
		switch {
		case strings.HasPrefix(rel, "site/docs/screens/") && strings.HasSuffix(rel, ".json"):
			add(rel, "generated")
		case strings.HasPrefix(rel, "site/") && (strings.HasSuffix(rel, ".md") || strings.HasSuffix(rel, ".html") || strings.HasSuffix(rel, ".js")):
			add(rel, "wiki")
		case strings.HasSuffix(rel, "_test.go"):
			add(rel, "test")
		}
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].rel < out[j].rel })
	return out
}

// followUps reports, for each name the import changed, where the old name
// is still written, and regenerates the wiki's tech tables.
func followUps(c *catalog, root string, changes []change, o importOptions, out io.Writer) {
	var renamed []change
	tables := false
	for _, ch := range changes {
		if ch.r.kind == kindName {
			renamed = append(renamed, ch)
		}
		switch strings.SplitN(ch.r.id, ".", 2)[0] {
		case "tech", "building", "age":
			tables = true
		}
	}
	if len(renamed) > 0 {
		quoters := loadQuoters(root)
		fmt.Fprintln(out, "\nNames changed. The old name is still written in these places, which the import does not edit:")
		for _, ch := range renamed {
			old := strings.TrimSpace(ch.r.current)
			fmt.Fprintf(out, "  %s: “%s” is now “%s”\n", ch.r.id, old, strings.TrimSpace(ch.text))
			if len([]rune(old)) < 3 {
				fmt.Fprintln(out, "      (too short a name to search for)")
				continue
			}
			re := regexp.MustCompile(`(^|[^\pL\pN_])` + regexp.QuoteMeta(old) + `($|[^\pL\pN_])`)
			found := 0
			for _, q := range quoters {
				if n := len(re.FindAllStringIndex(q.text, -1)); n > 0 {
					found++
					if found <= 12 {
						fmt.Fprintf(out, "      %-9s %s (%d)\n", q.kind, q.rel, n)
					}
				}
			}
			switch {
			case found == 0:
				fmt.Fprintln(out, "      nowhere else")
			case found > 12:
				fmt.Fprintf(out, "      and %d more files\n", found-12)
			}
		}
		fmt.Fprintln(out, "  wiki: a page under site/, to edit by hand. generated: a picture of the game, redrawn with")
		fmt.Fprintln(out, "  go test -tags mapcapture -run TestWriteSiteScreens ./ui. test: a test that quotes the name.")
	}
	// Any other rewritten line that the wiki quotes word for word.
	if stale := wikiQuotes(root, changes); len(stale) > 0 {
		fmt.Fprintln(out, "\nThe wiki still quotes these lines as they were. The import does not edit it:")
		for _, line := range stale {
			fmt.Fprintln(out, "  "+line)
		}
	}
	if !tables {
		return
	}
	switch {
	case o.dryRun:
		fmt.Fprintln(out, "\nThe wiki's tech tables (site/docs/technologies.md) would be regenerated.")
	case o.regenerate:
		cmd := exec.Command(wikiGenerator[0], wikiGenerator[1:]...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "UPDATE_WIKI=1")
		var buf bytes.Buffer
		cmd.Stdout, cmd.Stderr = &buf, &buf
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(out, "\nThe wiki's tech tables could not be regenerated (UPDATE_WIKI=1 %s):\n%s\n", strings.Join(wikiGenerator, " "), buf.String())
			return
		}
		fmt.Fprintln(out, "\nRegenerated the wiki's tech tables (site/docs/technologies.md).")
	}
}

// wikiQuotes lists the wiki pages that still hold the old text of a
// rewritten line that is not a name, a page a line: the page, how many
// lines, and the first few ids. Only lines long enough to be found
// reliably (twelve characters) are looked for.
func wikiQuotes(root string, changes []change) []string {
	var lines []change
	for _, ch := range changes {
		if ch.r.kind != kindName && len([]rune(strings.TrimSpace(ch.r.current))) >= 12 {
			lines = append(lines, ch)
		}
	}
	if len(lines) == 0 {
		return nil
	}
	var out []string
	for _, q := range loadQuoters(root) {
		if q.kind != "wiki" {
			continue
		}
		var ids []string
		for _, ch := range lines {
			if strings.Contains(q.text, strings.TrimSpace(ch.r.current)) {
				ids = append(ids, ch.r.id)
			}
		}
		switch {
		case len(ids) == 0:
		case len(ids) <= 3:
			out = append(out, fmt.Sprintf("%s: %s", q.rel, strings.Join(ids, ", ")))
		default:
			out = append(out, fmt.Sprintf("%s: %d lines (%s, ...)", q.rel, len(ids), strings.Join(ids[:3], ", ")))
		}
	}
	return out
}
