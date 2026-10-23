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
	var none []any // a format is put together with no arguments, as vet wants of a format it cannot see
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
