package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/vibe"
)

// TestEveryTextOfTheLogSheetIsWiredOrLeftAlone: a text of the log sheet that
// can follow VibeMode does. A new log line or refusal that is not written as a
// vibe call fails this test, and the fix is one command:
//
//	go run ./cmd/words vibe-wire
//
// (Texts that cannot follow the switch as they are written, or that were
// judged not to, are in vibe_left.go with the reason.)
func TestEveryTextOfTheLogSheetIsWiredOrLeftAlone(t *testing.T) {
	c := realCatalog(t)
	var loose []string
	wired := 0
	for _, r := range c.rows {
		if r.area != areaLog {
			continue
		}
		switch {
		case r.u.vibeID != "":
			wired++
		case c.wireVerdict(r).kind != wireNone:
			loose = append(loose, fmt.Sprintf("%s (%s:%d)", r.id, r.u.f.rel, r.u.line))
		}
	}
	if wired < 600 {
		t.Errorf("only %d texts of the log sheet are wired", wired)
	}
	if len(loose) > 0 {
		t.Errorf("%d texts of the log sheet can follow VibeMode and are not written as vibe calls; run `go run ./cmd/words vibe-wire`:\n  %s",
			len(loose), strings.Join(loose, "\n  "))
	}
}

// TestEveryWiredTextReadsPlainWhileVibeModeIsOff walks every vibe call in the
// source. Each is the text of its row of the export, word for word; with
// VibeMode off it reads as that text (and as fmt would put it together, for a
// format) even when the script has a line for the row; and with VibeMode on
// and no line for the row it reads the same. Nothing a player reads changes
// until a line is written and the switch is thrown.
func TestEveryWiredTextReadsPlainWhileVibeModeIsOff(t *testing.T) {
	c := realCatalog(t)
	calls := lineCalls(c.m)
	script := map[string]string{}
	for _, lc := range calls {
		script[lc.id] = "SCRIPT " + lc.id
	}
	was := vibe.On()
	defer vibe.Set(was)
	var none []any // spread into a format that is not a constant: vet flags such a call with no arguments at all
	read := func(lc lineCall) string {
		switch lc.kind {
		case "Sprintf":
			return vibe.Sprintf(lc.id, lc.plain, none...)
		case "Errorf":
			return vibe.Errorf(lc.id, lc.plain, none...).Error()
		}
		return vibe.Line(lc.id, lc.plain)
	}
	asFmt := func(lc lineCall) string {
		switch lc.kind {
		case "Sprintf":
			return fmt.Sprintf(lc.plain, none...)
		case "Errorf":
			return fmt.Errorf(lc.plain, none...).Error()
		}
		return lc.plain
	}
	n := 0
	for _, lc := range calls {
		where := fmt.Sprintf("%s:%d: vibe.%s(%q)", lc.file, lc.line, lc.kind, lc.id)
		if !lc.literal {
			t.Errorf("%s: the text is not a string literal", where)
			continue
		}
		row, ok := c.byID[lc.id]
		if !ok || row.current != lc.plain {
			t.Errorf("%s: is not the text of its row of the export", where)
			continue
		}
		n++
		want := asFmt(lc)

		restore := vibe.UseScriptForTest(script)
		vibe.Set(false)
		if got := read(lc); got != want {
			t.Errorf("%s: with VibeMode off and a line in the script, it reads %q, want %q", where, got, want)
		}
		restore()

		restore = vibe.UseScriptForTest(map[string]string{})
		vibe.Set(true)
		if got := read(lc); got != want || got == "" {
			t.Errorf("%s: with VibeMode on and no line in the script, it reads %q, want %q", where, got, want)
		}
		restore()
		vibe.Set(false)
	}
	if n < 600 {
		t.Errorf("only %d vibe calls walked", n)
	}
}

// TestVibeWireAddsTheImportToAFileWithNone: a file with no import block that
// gets a vibe call gets the vibe import with it, so it still builds; a file
// with imports keeps them, in the right group.
func TestVibeWireAddsTheImportToAFileWithNone(t *testing.T) {
	const call = "var greeting = vibe.Line(\"config.x.greeting.01\", \"Hello\")\n"
	for name, tc := range map[string]struct{ src, want string }{
		"none": {
			"package config\n\n// greeting says hello.\n" + call,
			"package config\n\nimport \"github.com/espresso20/ageforge/vibe\"\n\n// greeting says hello.\n" + call,
		},
		"none, a package doc": {
			"// Package config is the config.\npackage config\n\n" + call,
			"// Package config is the config.\npackage config\n\nimport \"github.com/espresso20/ageforge/vibe\"\n\n" + call,
		},
		"already there": {
			"package config\n\nimport \"github.com/espresso20/ageforge/vibe\"\n\n" + call,
			"package config\n\nimport \"github.com/espresso20/ageforge/vibe\"\n\n" + call,
		},
		"another import": {
			"package config\n\nimport \"strings\"\n\n" + call + "var _ = strings.ToUpper\n",
			"package config\n\nimport (\n\t\"strings\"\n\n\t\"github.com/espresso20/ageforge/vibe\"\n)\n\n" + call + "var _ = strings.ToUpper\n",
		},
	} {
		got, err := fixImports("config/x.go", []byte(tc.src))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if string(got) != tc.want {
			t.Errorf("%s: the file reads\n%s\nwant\n%s", name, got, tc.want)
		}
	}
}

// TestATableEntryCarriesItsRowIdAndCountsAsWired: the road event templates are a
// table made once, so each entry holds its own row id (vibeRow) and whoever
// draws one reads it through vibe.Line. The export uses the id, takes the
// id's own string for a key, and calls the row wired.
func TestATableEntryCarriesItsRowIdAndCountsAsWired(t *testing.T) {
	c := realCatalog(t)
	w := newVibeWiring(c)
	n := 0
	for _, r := range c.rows {
		if !strings.HasPrefix(r.id, "config.mission_words.roadEventTemplates.") {
			continue
		}
		n++
		if !r.u.vibeTable || r.u.vibeID != r.id {
			t.Errorf("%s is not read as a table entry carrying its own id", r.id)
		}
		if !w.wired(r.id) {
			t.Errorf("%s is not counted as wired", r.id)
		}
		if reason := c.wireVerdict(r).reason; forGood(reason) {
			t.Errorf("%s is left alone for good: %s", r.id, reason)
		}
		if r.area != areaLog {
			t.Errorf("%s is on the sheet %s", r.id, r.area.file)
		}
	}
	if n != 27 {
		t.Errorf("%d road event templates in the export, want 27", n)
	}
	for _, r := range c.rows {
		if strings.HasPrefix(r.id, "unclassified.") && strings.Contains(r.id, "roadTemplate") {
			t.Errorf("%s: the id of a table entry is read as text", r.id)
		}
	}
}
