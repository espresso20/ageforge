package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// apply.go is the import: it reads the sheets, decides row by row whether a
// rewrite can go in, and writes the ones that can into the string literals
// they came from.

type importOptions struct {
	dryRun     bool // print what would change and what would be refused; write nothing
	noSame     bool // do not carry a rewrite to the rows whose same_as points at it
	build      bool // build the module after applying
	regenerate bool // rewrite the wiki tables that have a generator
}

// entry is one filled row of one sheet.
type entry struct {
	file           string
	id             string
	current, yours string
}

// change is a rewrite that passed every check.
type change struct {
	r     *row
	text  string // the new text, with the original's edge spaces
	file  string // the sheet it came from
	after string // the id whose rewrite it took over, for a same_as row
}

// refusal is a row the import will not apply, and why.
type refusal struct {
	file, id, why string
}

// result is what an import decided.
type result struct {
	changes   []change
	refused   []refusal
	unchanged int
	settled   map[string]string // id -> text, for rows whose rewrite is already in the source
}

// blank reports whether a yours cell is empty.
func blank(s string) bool { return strings.TrimSpace(s) == "" }

// decide reads the sheets in dir and sorts their filled rows into changes,
// refusals and rows that need nothing.
func decide(c *catalog, dir string, noSame bool) (*result, error) {
	names, err := sheetFiles(dir)
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no sheets in %s: run the export first", dir)
	}
	var order []string           // ids with a filled yours, in sheet order
	byID := map[string][]entry{} // every filled row, by id
	type follower struct{ file, id, current, leader string }
	var followers []follower
	for _, n := range names {
		recs, err := readSheet(filepath.Join(dir, n))
		if err != nil {
			return nil, err
		}
		for _, rec := range recs {
			id := rec[colID]
			if blank(rec[colYours]) {
				if rec[colSameAs] != "" {
					followers = append(followers, follower{n, id, rec[colCurrent], strings.TrimSpace(rec[colSameAs])})
				}
				continue
			}
			if _, seen := byID[id]; !seen {
				order = append(order, id)
			}
			byID[id] = append(byID[id], entry{n, id, rec[colCurrent], rec[colYours]})
		}
	}

	res := &result{settled: map[string]string{}}
	accepted := map[string]string{} // id -> the yours that stands for it
	refusedID := map[string]bool{}
	refuse := func(file, id, why string) {
		res.refused = append(res.refused, refusal{file, id, why})
		refusedID[id] = true
	}
	// try runs every check on one rewrite of one row.
	try := func(e entry, after string) {
		r, ok := c.byID[e.id]
		if !ok {
			refuse(e.file, e.id, "this id is no longer in the source (the line was removed or moved): export again and carry the rewrite over")
			return
		}
		text := restoreEdges(r.current, e.yours)
		if text == r.current {
			// Either the rewrite matches the line, or it is already applied.
			res.unchanged++
			if e.current != r.current {
				res.settled[e.id] = text
			}
			accepted[e.id] = e.yours
			return
		}
		if e.current != r.current {
			refuse(e.file, e.id, fmt.Sprintf("the line has changed in the source since the export; it now reads %s", strconv.Quote(r.current)))
			return
		}
		problems := checkKeep(r, text)
		if n := len([]rune(text)); r.max > 0 && n > r.max {
			problems = append(problems, fmt.Sprintf("is %d characters long and the most is %d", n, r.max))
		}
		problems = append(problems, c.rules.check(r, text)...)
		if len(problems) > 0 {
			refuse(e.file, e.id, strings.Join(problems, "; "))
			return
		}
		res.changes = append(res.changes, change{r: r, text: text, file: e.file, after: after})
		accepted[e.id] = e.yours
	}

	for _, id := range order {
		es := byID[id]
		clash := false
		for _, e := range es[1:] {
			if strings.TrimSpace(e.yours) != strings.TrimSpace(es[0].yours) {
				clash = true
			}
		}
		if clash {
			var where []string
			for _, e := range es {
				where = append(where, fmt.Sprintf("%s has %s", e.file, strconv.Quote(strings.TrimSpace(e.yours))))
			}
			for _, e := range es {
				refuse(e.file, id, "filled differently in two files ("+strings.Join(where, ", ")+"): make them the same or clear one")
			}
			continue
		}
		try(es[0], "")
	}

	// A blank row whose same_as points at a rewritten row takes that rewrite.
	if !noSame {
		done := map[string]bool{}
		for _, f := range followers {
			if done[f.id] || len(byID[f.id]) > 0 {
				continue // already handled, or filled in another sheet
			}
			yours, ok := accepted[f.leader]
			if !ok || refusedID[f.leader] {
				continue
			}
			done[f.id] = true
			try(entry{f.file, f.id, f.current, yours}, f.leader)
		}
	}
	return res, nil
}

// checkKeep holds a rewrite to the pieces of the original it must carry
// over: the same format verbs in the same order, and the same slots, style
// tags, key names, quoted commands, line breaks and glyphs.
func checkKeep(r *row, yours string) []string {
	slotMark := r.rules&rulesSkel != 0
	was, now := tokensOf(r.current, slotMark), tokensOf(yours, slotMark)
	var out []string

	verbs := func(ts []kept) []string {
		var v []string
		for _, t := range ts {
			if t.kind == tokVerb {
				v = append(v, t.text)
			}
		}
		return v
	}
	if a, b := verbs(was), verbs(now); strings.Join(a, " ") != strings.Join(b, " ") {
		switch {
		case len(a) == 0:
			out = append(out, fmt.Sprintf("has a format verb (%s) and the original has none", strings.Join(b, " ")))
		case sameSet(a, b):
			out = append(out, fmt.Sprintf("has its format verbs out of order: they must stay %s", strings.Join(a, " then ")))
		default:
			out = append(out, fmt.Sprintf("must keep the format verbs %s, in that order, and has %s", strings.Join(a, " "), orNone(b)))
		}
	}
	count := func(ts []kept) map[string]int {
		m := map[string]int{}
		for _, t := range ts {
			if t.kind != tokVerb {
				m[show(t)]++
			}
		}
		return m
	}
	a, b := count(was), count(now)
	var names []string
	for k := range a {
		names = append(names, k)
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			names = append(names, k)
		}
	}
	sort.Strings(names)
	for _, k := range names {
		switch {
		case b[k] < a[k] && b[k] == 0:
			out = append(out, fmt.Sprintf("is missing %s", k))
		case b[k] < a[k]:
			out = append(out, fmt.Sprintf("has %s %d times and must have it %d times", k, b[k], a[k]))
		case b[k] > a[k] && a[k] == 0:
			out = append(out, fmt.Sprintf("has %s, which the original does not", k))
		case b[k] > a[k]:
			out = append(out, fmt.Sprintf("has %s %d times and the original has it %d times", k, b[k], a[k]))
		}
	}
	if p := unbalanced(now); p != "" && unbalanced(was) != p {
		out = append(out, "has unbalanced style tags: "+p)
	}
	return out
}

// show writes a kept piece for a message.
func show(t kept) string {
	if t.kind == tokBreak {
		return `a line break (\n)`
	}
	return t.text
}

func orNone(v []string) string {
	if len(v) == 0 {
		return "none"
	}
	return strings.Join(v, " ")
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x, y := append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	return strings.Join(x, "\x00") == strings.Join(y, "\x00")
}

// unbalanced says what is wrong with the order of a line's style tags, or
// "" when each opening tag is closed and nothing is closed twice. A tag such
// as [red] opens a style and [-] closes it.
func unbalanced(ts []kept) string {
	open := false
	for _, t := range ts {
		if t.kind != tokTag || !isStyleTag(t.text) {
			continue
		}
		closes := strings.Trim(t.text, "[]-:") == ""
		switch {
		case closes && !open:
			return "a closing " + t.text + " comes before any opening tag"
		case closes:
			open = false
		default:
			open = true
		}
	}
	if open {
		return "a style is opened and never closed with [-]"
	}
	return ""
}

// quote writes text as a Go string literal in the style of the literal it
// replaces: between backquotes when that one was and the text allows it.
func quote(u *unit, text string) string {
	if u.raw && !strings.ContainsAny(text, "`\r") {
		return "`" + text + "`"
	}
	return strconv.Quote(text)
}

// rewrite returns the new contents of the source files the changes touch.
func rewrite(c *catalog, changes []change) (map[string][]byte, error) {
	type edit struct {
		start, end int
		text       string
	}
	edits := map[string][]edit{}
	for _, ch := range changes {
		edits[ch.r.u.f.rel] = append(edits[ch.r.u.f.rel], edit{ch.r.u.start, ch.r.u.end, quote(ch.r.u, ch.text)})
		for _, tw := range ch.r.twins {
			edits[tw.f.rel] = append(edits[tw.f.rel], edit{tw.start, tw.end, quote(tw, ch.text)})
		}
	}
	out := map[string][]byte{}
	for rel, es := range edits {
		sort.Slice(es, func(i, j int) bool { return es[i].start > es[j].start })
		src := append([]byte(nil), c.m.byRel[rel].src...)
		last := len(src) + 1
		for _, e := range es {
			if e.end > last {
				return nil, fmt.Errorf("%s: two rewrites overlap", rel)
			}
			src = append(src[:e.start], append([]byte(e.text), src[e.end:]...)...)
			last = e.start
		}
		formatted, err := format.Source(src)
		if err != nil {
			return nil, fmt.Errorf("%s would not parse after the rewrite: %w", rel, err)
		}
		out[rel] = formatted
	}
	return out, nil
}

// importSheets applies the sheets in dir to the source under root.
func importSheets(root, dir string, o importOptions, out io.Writer) error {
	c, err := read(root)
	if err != nil {
		return err
	}
	res, err := decide(c, dir, o.noSame)
	if err != nil {
		return err
	}
	files, err := rewrite(c, res.changes)
	if err != nil {
		return err
	}

	verb := "Applied"
	if o.dryRun {
		verb = "Would apply"
	}
	for _, ch := range res.changes {
		note := ""
		if ch.after != "" {
			note = "  (same as " + ch.after + ")"
		}
		fmt.Fprintf(out, "%s  %s%s\n    was: %s\n    now: %s\n", verb, ch.r.id, note, strconv.Quote(ch.r.current), strconv.Quote(ch.text))
	}
	for _, rf := range res.refused {
		fmt.Fprintf(out, "Refused  %s  (%s)\n    %s\n", rf.id, rf.file, rf.why)
	}

	if !o.dryRun {
		rels := make([]string, 0, len(files))
		for rel := range files {
			rels = append(rels, rel)
		}
		sort.Strings(rels)
		for _, rel := range rels {
			if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(rel)), files[rel], 0o644); err != nil {
				return err
			}
		}
		led, err := loadLedger(root)
		if err != nil {
			return err
		}
		for _, ch := range res.changes {
			led[ch.r.id] = hashText(ch.text)
		}
		for id, text := range res.settled {
			led[id] = hashText(text)
		}
		if len(res.changes) > 0 || len(res.settled) > 0 {
			if err := saveLedger(root, led); err != nil {
				return err
			}
		}
	}

	followUps(c, root, res.changes, o, out)

	if o.build && !o.dryRun && len(res.changes) > 0 {
		cmd := exec.Command("go", "build", "./...")
		cmd.Dir = root
		var buf bytes.Buffer
		cmd.Stdout, cmd.Stderr = &buf, &buf
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(out, "\nThe game does not build after the import:\n%s\n", buf.String())
			return fmt.Errorf("go build ./... failed")
		}
		fmt.Fprintln(out, "\nThe game builds.")
	}

	summary := "applied"
	if o.dryRun {
		summary = "would apply"
	}
	fmt.Fprintf(out, "\n%s %d, refused %d, unchanged %d\n", summary, len(res.changes), len(res.refused), res.unchanged)
	if !o.dryRun && len(res.changes) > 0 {
		fmt.Fprintln(out, "Next: go test ./...")
	}
	return nil
}
