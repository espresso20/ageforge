package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
)

// endTimer is a timer the film set, held by the test.
type endTimer struct {
	d    time.Duration
	fn   func()
	dead bool
}

// endingRig is the film on a stage: an overlay manager on a page stack, a
// clock the test holds and timers it fires by hand.
type endingRig struct {
	e      *ending
	om     *OverlayManager
	pages  *tview.Pages
	clock  time.Time
	timers []*endTimer
	closed int // times the keyboard went back to the prompt
}

// live is the timers that have not been called off, soonest first.
func (r *endingRig) live() (out []*endTimer) {
	for _, t := range r.timers {
		if !t.dead {
			out = append(out, t)
		}
	}
	return out
}

// sampleEnding is the record of a run's ending as the engine writes one,
// with lines of every kind it reports (their wording is the fixture's own).
func sampleEnding(kind string, level int) game.RunEnding {
	end := game.RunEnding{Kind: kind, Age: "medieval_age", Level: level, Points: 13, Full: 13, Badges: 4}
	say := func(k, lvl, text string) {
		end.Lines = append(end.Lines, game.RunEndingLine{Kind: k, Level: lvl, Text: text})
	}
	switch kind {
	case game.RunEndEndured:
		end.Age, end.Points, end.Full, end.Catastrophe = "interstellar_age", 1200, 2000, "The Last Passage"
		say(game.EndLineVerdict, "warning", "☄ Endure: The Last Passage. The stars go out one by one, and the fleet holds its course.")
		say(game.EndLineVerdict, "warning", "  You keep 60% of this run's prestige points: 1200 of 2000.")
	case game.RunEndSuccumbed:
		end.Age, end.Points, end.Full, end.Catastrophe = "interstellar_age", 0, 2000, "The Last Passage"
		say(game.EndLineVerdict, "event", "☄ Succumb: The Last Passage took everything this run had. No prestige points from it.")
	case game.RunEndSpared:
		end.Age, end.Points, end.Full = "interstellar_age", 2000, 2000
	case game.RunEndFallen:
		end.Age, end.Points, end.Full, end.Catastrophe = "classical_age", 0, 0, "The Great Plague"
		say(game.EndLineVerdict, "event", "☄ The Great Plague: civilization has fallen. A new dawn.")
		say(game.EndLineLegacy, "success", "Iron Era legacy bonus (permanent): iron production +10%.")
		say(game.EndLineKnowledge, "success", "Ancient Knowledge: research time -5% for each epoch succumbed in, now -5% (permanent).")
	}
	if kind != game.RunEndFallen {
		say(game.EndLineVoice, "info", "  [gray]The bells rang the town out, and the last cart left by the east road.[-]")
		say(game.EndLineComplete, "success", fmt.Sprintf("Prestige complete. Level %d, %d prestige points earned.", level, end.Points))
	}
	if kind == game.RunEndPrestige && level == 1 {
		say(game.EndLineEarly, "info", "That was an early taste: a prestige from the Medieval Age paid 13 prestige points, for the 5 ages the run completed. Going deeper pays far more: each era's ages are worth 3 times the era before, and a run to the Modern Age pays 1,093 prestige points.")
	}
	if kind == game.RunEndSuccumbed {
		say(game.EndLineLegacy, "success", "✦ Cosmic Legacy: all production +10%, permanent. It survives every prestige and every fall.")
	}
	say(game.EndLineMastery, "info", "Era Mastery: the Primitive Age to the Classical Age gained a mastery level each and will run faster.")
	say(game.EndLineGround, "info", "Known ground: the Primitive Age runs 1.5x faster (mastery 1).")
	if level > 2 {
		say(game.EndLineKnowledge, "info", "Ancient Knowledge from 2 epochs you succumbed in: research time -10%.")
		say(game.EndLineRuins, "info", "Ruins carried forward from past civilizations: 3 types.")
		say(game.EndLineKit, "info", "Worker Shares: your shares carry over. food 40%, wood 30%, knowledge 30%.")
	}
	return end
}

// stagedEnding is the film of an ending on a stage, shown the way the
// dashboard shows it: a refresh with the run that ended, the film, then a
// refresh with the new run. towns draws the towns' pictures (most of what
// a frame costs; most tests do not look at them).
func stagedEnding(t testing.TB, end game.RunEnding, tier mapmodel.GlyphTier, motion, towns bool) *endingRig {
	t.Helper()
	r := &endingRig{pages: tview.NewPages(), clock: time.Unix(3_000_000, 0)}
	r.om = NewOverlayManager(r.pages, tview.NewApplication(), func() { r.closed++ })
	r.om.after = func(d time.Duration, fn func()) func() {
		tm := &endTimer{d: d, fn: fn}
		r.timers = append(r.timers, tm)
		return func() { tm.dead = true }
	}
	old := fixture.State(fixture.Options{Age: end.Age, Seed: 7})
	cur := fixture.State(fixture.Options{Age: "primitive_age", Seed: 7})
	cur.Prestige.Level = end.Level
	r.om.Refresh(old)
	ShowRunEnding(r.om, end)
	e := r.om.film
	e.now = func() time.Time { return r.clock }
	e.start = r.clock
	if !towns {
		e.reg = nil
	}
	r.om.Refresh(cur)
	e.set.Tier, e.set.Motion = tier, motion
	e.settings()
	e.sc = nil
	e.arm()
	r.e = e
	return r
}

// restart takes the film back to its first frame, in a glyph set and with
// motion on or off, so one staged film serves many sizes.
func (r *endingRig) restart(tier mapmodel.GlyphTier, motion bool) {
	e := r.e
	e.set.Tier, e.set.Motion = tier, motion
	e.settings()
	e.sc, e.base, e.start, e.quiet, e.keys = nil, 0, r.clock, time.Time{}, 0
	e.arm()
}

// at draws the film at w by h at a frame.
func (r *endingRig) at(w, h, frame int) *mGrid {
	r.clock = r.e.start.Add(time.Duration(frame-r.e.base) * mapAnimStep)
	return r.e.frameGrid(w, h)
}

// TestEndingDump writes the film's beats when ENDING_DUMP names a folder:
// text for each, and one HTML page (ending.html) in the themes
// ENDING_DUMP_THEMES lists, for a look at them while working on them.
func TestEndingDump(t *testing.T) {
	out := os.Getenv("ENDING_DUMP")
	if out == "" {
		t.Skip("set ENDING_DUMP to a folder to write the beats")
	}
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	themes := strings.Split(os.Getenv("ENDING_DUMP_THEMES"), ",")
	if themes[0] == "" {
		themes = []string{theme.DefaultKey}
	}
	sizes := os.Getenv("ENDING_DUMP_SIZES")
	if sizes == "" {
		sizes = "120x40,80x24"
	}
	var page strings.Builder
	page.WriteString("<!doctype html><meta charset=\"utf-8\"><title>Ending</title><style>body{background:#222;color:#ddd;font:13px sans-serif}pre{font:10px/12px 'JetBrains Mono',Menlo,monospace;display:inline-block;margin:4px 12px 16px 0;vertical-align:top}h3{margin:10px 0 2px}</style>\n")
	for _, key := range themes {
		if err := theme.SetActive(key); err != nil {
			t.Fatal(err)
		}
		for _, sz := range strings.Split(sizes, ",") {
			var w, h int
			if _, err := fmt.Sscanf(sz, "%dx%d", &w, &h); err != nil {
				t.Fatal(err)
			}
			for _, c := range []struct {
				name   string
				end    game.RunEnding
				tier   mapmodel.GlyphTier
				motion bool
			}{
				{"first", sampleEnding(game.RunEndPrestige, 1), mapmodel.TierUnicode, true},
				{"seventh", sampleEnding(game.RunEndPrestige, 7), mapmodel.TierUnicode, true},
				{"endured", sampleEnding(game.RunEndEndured, 3), mapmodel.TierUnicode, true},
				{"fallen", sampleEnding(game.RunEndFallen, 2), mapmodel.TierUnicode, true},
				{"still", sampleEnding(game.RunEndPrestige, 3), mapmodel.TierUnicode, false},
				{"plain", sampleEnding(game.RunEndPrestige, 3), mapmodel.TierASCII, true},
			} {
				r := stagedEnding(t, c.end, c.tier, c.motion, true)
				r.at(w, h, 0)
				var frames []int
				for _, s := range r.e.sc.spans {
					switch s.beat {
					case beatEnding:
						frames = append(frames, s.from+2, s.from+8, s.from+14, s.from+19)
					case beatReckoning:
						frames = append(frames, s.from+endLead+1, s.from+s.frames/2, s.end()-2)
					case beatStrike:
						frames = append(frames, s.from+endBlow2, s.from+endBlow, s.from+endBlow+2, s.from+endBlow+6, s.end()-1)
					case beatBeginning:
						frames = append(frames, s.from, s.from+3, s.from+10, s.end()-1)
					}
					if !c.motion {
						frames = append(frames[:len(frames)-len(frames)%100], s.from)
					}
				}
				if !c.motion {
					frames = nil
					for _, s := range r.e.sc.spans {
						frames = append(frames, s.from)
					}
				}
				for _, f := range frames {
					g := r.at(w, h, f)
					name := fmt.Sprintf("%s_%s_%dx%d_f%03d", key, c.name, w, h, f)
					if err := os.WriteFile(filepath.Join(out, name+".txt"), []byte(gridText(g)), 0644); err != nil {
						t.Fatal(err)
					}
					fmt.Fprintf(&page, "<h3>%s</h3>\n%s", name, gridHTML(g, r.e.pal, r.e.view.plain))
				}
				r.e.close()
			}
		}
	}
	if err := os.WriteFile(filepath.Join(out, "ending.html"), []byte(page.String()), 0644); err != nil {
		t.Fatal(err)
	}
}

// ---- what it says ----

// endFold is a line as the plain glyph set shows it.
func endFold(s string, plain bool) string {
	if plain {
		out := []rune(plainSafe(s))
		for i, r := range out {
			out[i] = mapmodel.Fold(r, mapmodel.TierASCII)
		}
		s = string(out)
	}
	return strings.Join(strings.Fields(s), " ")
}

// endScreen is everything on a frame, rows joined and blanks squeezed.
func endScreen(g *mGrid, plain bool) string {
	var rows []string
	for y := 0; y < g.h; y++ {
		rows = append(rows, g.row(y))
	}
	return endFold(strings.Join(rows, " "), plain)
}

// endSpanOf is the film's span for a beat, ok false when it has none.
func endSpanOf(sc *endScene, beat int) (endSpan, bool) {
	for _, s := range sc.spans {
		if s.beat == beat {
			return s, true
		}
	}
	return endSpan{}, false
}

// endings is the endings the film is checked for, by name.
func endings() map[string]game.RunEnding {
	return map[string]game.RunEnding{
		"a first prestige":            sampleEnding(game.RunEndPrestige, 1),
		"a seventh prestige":          sampleEnding(game.RunEndPrestige, 7),
		"the Last Passage, spared":    sampleEnding(game.RunEndSpared, 4),
		"the Last Passage, endured":   sampleEnding(game.RunEndEndured, 3),
		"the Last Passage, succumbed": sampleEnding(game.RunEndSuccumbed, 3),
		"a fall to a catastrophe":     sampleEnding(game.RunEndFallen, 2),
	}
}

// checkEndingFrames plays one staged film at a size and holds every beat
// to its layout: nothing cut off, the line that says how to go on, and
// what the beat is there to say.
func checkEndingFrames(t *testing.T, where string, r *endingRig, w, h int) {
	t.Helper()
	e := r.e
	v := &e.view
	plain := v.plain
	frame := func(f int) *mGrid {
		g := r.at(w, h, f)
		if len(g.clipped) > 0 {
			t.Errorf("%s, frame %d: text cut off at the edge: %q", where, f, g.clipped)
		}
		if !pageHas(g, v.prompt) {
			t.Errorf("%s, frame %d: the line that says how to go on is not on the screen", where, f)
		}
		return g
	}
	r.at(w, h, 0)
	sc := e.sc
	L := sc.L

	// The ending: the age and the run's closing line over the land.
	s, _ := endSpanOf(sc, beatEnding)
	for _, f := range []int{s.from, s.from + endBurnHold + 6, s.end() - 1} {
		g := frame(f)
		if caps := strings.ToUpper(v.age); !strings.Contains(g.row(1), caps) && !strings.Contains(g.row(1), spaced(caps)) {
			t.Errorf("%s, frame %d: the age is not named over the town: %q", where, f, strings.TrimSpace(g.row(1)))
		}
		for i, line := range L.voice {
			if !strings.Contains(g.row(3+i), line) {
				t.Errorf("%s, frame %d: the run's closing line is not under the age", where, f)
			}
		}
	}
	if g := frame(s.end() - 1); g.c[L.ey*w+L.ex].r != emberGlyph(plain) {
		t.Errorf("%s: the ending does not end on its ember", where)
	}

	// The reckoning: every line is on the screen whole when it has
	// landed and counted up, and the last frame holds the last of them.
	s, _ = endSpanOf(sc, beatReckoning)
	e.sc = nil
	r.at(w, h, s.from)
	sc = e.sc
	for i := range v.lines {
		g := frame(s.from + sc.reck.at[i] + endCountUp)
		last := len(sc.reck.rows)
		if i+1 < len(sc.reck.first) {
			last = sc.reck.first[i+1]
		}
		for _, row := range sc.reck.rows[sc.reck.first[i]:last] {
			if want := endFold(row.plain(), plain); want != "" && !strings.Contains(endScreen(g, plain), want) {
				t.Errorf("%s: line %d of the reckoning is not whole on the screen when it lands: %q", where, i+1, want)
			}
		}
		if i > 0 && sc.reck.at[i] <= sc.reck.at[i-1] {
			t.Errorf("%s: line %d lands with the line before it", where, i+1)
		}
	}
	if g := frame(s.end() - 1); g.c[L.ey*w+L.ex].r != emberGlyph(plain) {
		t.Errorf("%s: the reckoning lost its ember", where)
	}
	for _, row := range sc.reck.rows {
		if row.width() > L.lineW {
			t.Errorf("%s: a row of the reckoning is %d cells wide, over %d", where, row.width(), L.lineW)
		}
	}

	// The strike, a prestige's only: the word whole in iron, a star for
	// every prestige, and the game's own line for it.
	if s, ok := endSpanOf(sc, beatStrike); ok != v.prestige {
		t.Errorf("%s: a strike %v for a prestige %v", where, ok, v.prestige)
	} else if ok {
		for _, f := range []int{s.from, s.from + endBlow3, s.from + endBlow + 3, s.end() - 1} {
			g := frame(f)
			if short := ironShort(g, L.word, plain); short != 0 {
				t.Errorf("%s, frame %d: the word is not whole in iron (%d)", where, f, short)
			}
		}
		g := frame(s.from + endBlow + 3)
		star := "★"
		if plain {
			star = "*"
		}
		if n, want := strings.Count(g.row(L.stars), star), min(v.end.Level, (w-8)/2); n != want {
			t.Errorf("%s: %d stars under the word, want %d", where, n, want)
		}
		if on := endScreen(g, plain); !strings.Contains(on, endFold(v.complete, plain)) {
			t.Errorf("%s: the game's line for the prestige is not under the word", where)
		}
		if p := L.word; p.x < 2 || p.x+p.w > w-2 || p.y < 1 || L.capRow+len(L.caption) > L.hint {
			t.Errorf("%s: the word (%d,%d %dx%d) and its lines do not fit", where, p.x, p.y, p.w, p.h)
		}
	}

	// The beginning: the new town named, under the forged name for a
	// prestige.
	s, _ = endSpanOf(sc, beatBeginning)
	for _, f := range []int{s.from, s.from + 6, s.end() - 1} {
		g := frame(f)
		if capt := fitOption(v.captions, w-4); capt == "" || !strings.Contains(endFold(g.row(L.dawnCap), plain), endFold(capt, plain)) {
			t.Errorf("%s, frame %d: the new town is not named: %q", where, f, strings.TrimSpace(g.row(L.dawnCap)))
		}
		mark := 0
		for y := L.wy; y < L.wy+menuWordH; y++ {
			mark += strings.Count(g.row(y), "█") + strings.Count(g.row(y), "#")
		}
		if v.prestige != (mark > 100) {
			t.Errorf("%s, frame %d: the forged name has %d cells, for a prestige %v", where, f, mark, v.prestige)
		}
	}
}

// TestEndingFitsEverySize plays the film of every kind of ending at every
// size from 80x24 to 144x46 and three odd shapes, in both glyph sets.
func TestEndingFitsEverySize(t *testing.T) {
	for name, end := range endings() {
		r := stagedEnding(t, end, mapmodel.TierUnicode, true, false)
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
			sizes := arrivalSizes
			if tier == mapmodel.TierASCII && end.Level != 7 {
				sizes = arrivalSizes[:3]
			}
			for _, size := range sizes {
				r.restart(tier, true)
				checkEndingFrames(t, fmt.Sprintf("%s at %dx%d, %s glyphs", name, size[0], size[1], tier), r, size[0], size[1])
				if t.Failed() {
					return
				}
			}
		}
		r.e.close()
	}
}

// TestEndingWithItsTowns: the same with the towns drawn, small terminal
// and large. The town that burns fills the land at the start and is gone
// at the end; the new land is there in the beginning.
func TestEndingWithItsTowns(t *testing.T) {
	land := func(g *mGrid, from int) (n int) {
		for y := from; y < g.h-3; y++ {
			n += len([]rune(strings.Join(strings.Fields(g.row(y)), "")))
		}
		return n
	}
	for _, kind := range []string{game.RunEndPrestige, game.RunEndFallen} {
		r := stagedEnding(t, sampleEnding(kind, 3), mapmodel.TierUnicode, true, true)
		if r.e.oldTown == nil || r.e.newTown == nil {
			t.Fatalf("%s: no towns", kind)
		}
		for _, size := range [][2]int{{80, 24}, {120, 40}} {
			w, h := size[0], size[1]
			for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
				r.restart(tier, true)
				where := fmt.Sprintf("%s at %dx%d, %s glyphs, towns drawn", kind, w, h, tier)
				checkEndingFrames(t, where, r, w, h)
				r.restart(tier, true)
				first := r.at(w, h, 0)
				L := r.e.sc.L
				s, _ := endSpanOf(r.e.sc, beatEnding)
				if n := land(first, L.top); n < 2*w {
					t.Errorf("%s: the town that was built is not on the screen (%d cells)", where, n)
				}
				mid := r.at(w, h, s.from+endBurnHold+8)
				if a, b := land(mid, L.top), land(first, L.top); a >= b {
					t.Errorf("%s: the town is not burning down (%d cells of %d)", where, a, b)
				}
				s, _ = endSpanOf(r.e.sc, beatBeginning)
				if n := land(r.at(w, h, s.end()-1), L.dawnTop); n < 2*w {
					t.Errorf("%s: the new land is not on the screen (%d cells)", where, n)
				}
			}
		}
		r.e.close()
	}
}

// TestEndingOnATerminalTooSmall: under 80x24 each beat is a line, and no
// size at all makes the film fail.
func TestEndingOnATerminalTooSmall(t *testing.T) {
	r := stagedEnding(t, sampleEnding(game.RunEndPrestige, 3), mapmodel.TierUnicode, true, false)
	for _, w := range []int{1, 9, 30, 79, 80} {
		for _, h := range []int{1, 4, 12, 23, 24} {
			for _, motion := range []bool{true, false} {
				r.restart(mapmodel.TierUnicode, motion)
				for f := 0; f < 120; f += 13 {
					r.at(w, h, f)
				}
			}
		}
	}
	r.e.close()
}

// ---- the script ----

// TestEndingBeats holds the film to its script: the beats in order, each
// as long as it should be, and what happens in each.
func TestEndingBeats(t *testing.T) {
	r := stagedEnding(t, sampleEnding(game.RunEndPrestige, 3), mapmodel.TierUnicode, true, false)
	t.Cleanup(r.e.close)
	r.at(120, 40, 0)
	sc := r.e.sc
	var beats []int
	for _, s := range sc.spans {
		beats = append(beats, s.beat)
	}
	if fmt.Sprint(beats) != fmt.Sprint([]int{beatEnding, beatReckoning, beatStrike, beatBeginning}) {
		t.Fatalf("a prestige's beats: %v", beats)
	}
	if sc.spans[0].frames != endBurnFrames || sc.spans[2].frames != endStrikeFrames || sc.spans[3].frames != endDawnFrames {
		t.Errorf("the beats are %d, %d, %d and %d frames", sc.spans[0].frames, sc.spans[1].frames, sc.spans[2].frames, sc.spans[3].frames)
	}
	// Ten to thirteen seconds, whatever the run earned.
	for name, end := range endings() {
		v := endViewFor(end, r.e.pal)
		spans := endSpans(v.prestige, true, endReckFor(&v, r.e.pal, 76).frames)
		total := spans[len(spans)-1].end()
		lo, hi := 80, 106
		if !v.prestige {
			lo, hi = 56, 88 // a fall has no strike
		}
		if total < lo || total > hi {
			t.Errorf("%s: the film is %d frames (%.1fs), want %d to %d", name, total, float64(total)/8, lo, hi)
		}
	}
	at := func(f int) *endScene { r.at(120, 40, f); return r.e.sc }

	// The ending: the town stands, then burns from its edges, and the
	// tail is the ember's alone.
	if p := burn(endBurnHold); p != 0 {
		t.Errorf("the fire starts before the town has stood its moment (%.2f)", p)
	}
	if p := burn(endBurnFrames - endBurnTail); p < 1.2 {
		t.Errorf("the fire has not burnt the town out by the tail (%.2f)", p)
	}
	if sc := at(endBurnHold + 8); len(sc.sparks.a) == 0 {
		t.Error("a burning town throws no sparks")
	}

	// The reckoning: lines land one at a time, each with sparks.
	s := sc.spans[1]
	if sc := at(s.from + endLead - 1); sc.landed != 0 {
		t.Errorf("%d lines landed before the reckoning's lead was over", sc.landed)
	}
	for i := range sc.reck.at {
		if got := at(s.from + sc.reck.at[i]); got.landed != i+1 || got.strike < 0.5 {
			t.Errorf("line %d: %d landed, strike %.2f", i+1, got.landed, got.strike)
		}
	}

	// The strike: three blows build up, each harder, then the heavy one.
	s = sc.spans[2]
	if sc := at(s.from + endBlow1 - 1); sc.flash != 0 || sc.strike > 0.01 || sc.ring >= 0 {
		t.Errorf("before the first blow: flash %.2f, strike %.2f", sc.flash, sc.strike)
	}
	last, sparks := 0.0, 0
	for i, f := range []int{endBlow1, endBlow2, endBlow3, endBlow} {
		sc := at(s.from + f)
		if sc.flash <= last || len(sc.sparks.a) <= sparks {
			t.Errorf("blow %d is no harder than the one before: flash %.2f after %.2f, %d sparks after %d", i+1, sc.flash, last, len(sc.sparks.a), sparks)
		}
		last, sparks = sc.flash, len(sc.sparks.a)
	}
	if sc := at(s.from + endBlow); sc.heat != 1 || sc.ring < 0 || sc.shake == 0 {
		t.Errorf("the heavy blow: heat %.2f, ring %.1f, shake %d", sc.heat, sc.ring, sc.shake)
	}
	if sc := at(s.from + endBlow + 2); sc.shake != 0 || sc.ring < 3 {
		t.Errorf("after the heavy blow: shake %d, ring %.1f", sc.shake, sc.ring)
	}

	// The beginning comes out of the light.
	s = sc.spans[3]
	if sc := at(s.from); sc.flash < 0.5 || len(sc.sparks.a) > 4 {
		t.Errorf("the beginning: flash %.2f, %d sparks", sc.flash, len(sc.sparks.a))
	}
	if sc := at(s.end() - 1); sc.flash > 0.02 {
		t.Errorf("the beginning ends still lit (%.2f)", sc.flash)
	}

	// A fall has no strike and no sparks.
	f := stagedEnding(t, sampleEnding(game.RunEndFallen, 2), mapmodel.TierUnicode, true, false)
	t.Cleanup(f.e.close)
	f.at(120, 40, 0)
	if _, ok := endSpanOf(f.e.sc, beatStrike); ok || len(f.e.sc.spans) != 3 {
		t.Errorf("a fall's beats: %+v", f.e.sc.spans)
	}
	for frame := 0; frame < f.e.sc.total(); frame += 3 {
		f.at(120, 40, frame)
		if s, _ := f.e.sc.span(frame); s.beat != beatBeginning && len(f.e.sc.sparks.a) != 0 {
			t.Fatalf("a fall throws sparks at frame %d", frame)
		}
	}
}

// TestAPrestigeIsBiggerThanAnEpoch: the strike is the largest moment in
// the game. Its word covers more of the screen than an era's name does at
// every size, its blow lights the screen harder and longer, and it throws
// more sparks.
func TestAPrestigeIsBiggerThanAnEpoch(t *testing.T) {
	era := bareArrival(t, "iron_age", mapmodel.TierUnicode, true)
	film := stagedEnding(t, sampleEnding(game.RunEndPrestige, 1), mapmodel.TierUnicode, true, false)
	t.Cleanup(era.a.close)
	t.Cleanup(film.e.close)
	for _, size := range arrivalSizes {
		w, h := size[0], size[1]
		era.restart(mapmodel.TierUnicode, true)
		film.restart(mapmodel.TierUnicode, true)
		blow := era.at(w, h, arrEraStrike)
		eraFlash, eraSparks := era.a.sc.flash, len(era.a.sc.sparks.a)
		film.at(w, h, 0)
		s, _ := endSpanOf(film.e.sc, beatStrike)
		g := film.at(w, h, s.from+endBlow)
		sc := film.e.sc
		where := fmt.Sprintf("at %dx%d", w, h)
		if p, q := sc.L.word, era.a.sc.L.era; !p.ok || p.w*p.h < q.w*q.h || p.h < q.h {
			t.Errorf("%s: the word (%dx%d) is smaller than an era's name (%dx%d)", where, p.w, p.h, q.w, q.h)
		}
		if n := litGround(g, film.e.pal); n != w*h || litGround(blow, era.a.pal) != w*h {
			t.Errorf("%s: the heavy blow lights %d of %d cells", where, n, w*h)
		}
		if sc.flash <= eraFlash || len(sc.sparks.a) <= eraSparks {
			t.Errorf("%s: flash %.2f to an epoch's %.2f, %d sparks to its %d", where, sc.flash, eraFlash, len(sc.sparks.a), eraSparks)
		}
		// And the light lasts longer.
		if film.at(w, h, s.from+endBlow+4); film.e.sc.flash <= 0.1 {
			t.Errorf("%s: the light is gone half a second after the blow", where)
		}
	}
	if endStrikeFrames <= arrEraFrames-arrEraStrike {
		t.Errorf("the strike is %d frames", endStrikeFrames)
	}
}

// TestEndingGrowsWithPrestige: a later prestige is not the first one
// played again. There is a star more for each, more sparks at the blow and
// at each line, and what the run earned is its own.
func TestEndingGrowsWithPrestige(t *testing.T) {
	first := stagedEnding(t, sampleEnding(game.RunEndPrestige, 1), mapmodel.TierUnicode, true, false)
	later := stagedEnding(t, sampleEnding(game.RunEndPrestige, 7), mapmodel.TierUnicode, true, false)
	t.Cleanup(first.e.close)
	t.Cleanup(later.e.close)
	shot := func(r *endingRig) (stars, sparks, lineSparks int, reck endSpan) {
		r.at(120, 40, 0)
		reck, _ = endSpanOf(r.e.sc, beatReckoning)
		r.at(120, 40, reck.from+r.e.sc.reck.at[0])
		lineSparks = len(r.e.sc.sparks.a)
		s, _ := endSpanOf(r.e.sc, beatStrike)
		g := r.at(120, 40, s.from+endBlow)
		return strings.Count(g.row(r.e.sc.L.stars), "★"), len(r.e.sc.sparks.a), lineSparks, reck
	}
	s1, k1, l1, r1 := shot(first)
	s7, k7, l7, r7 := shot(later)
	if s1 != 1 || s7 != 7 {
		t.Errorf("stars: %d for the first prestige, %d for the seventh", s1, s7)
	}
	if k7 <= k1 || l7 <= l1 {
		t.Errorf("sparks: %d and %d at the first prestige, %d and %d at the seventh", k1, l1, k7, l7)
	}
	if len(later.e.view.lines) <= len(first.e.view.lines)-1 || r7.frames == 0 || r1.frames == 0 {
		t.Errorf("the reckoning: %d lines then %d", len(first.e.view.lines), len(later.e.view.lines))
	}
}

// TestEndingVerdict: what a catastrophe or the Last Passage did comes
// first in the reckoning, stands alone a while, and lands in silence. The
// celebration does not start over it.
func TestEndingVerdict(t *testing.T) {
	for _, kind := range []string{game.RunEndEndured, game.RunEndSuccumbed, game.RunEndFallen} {
		end := sampleEnding(kind, 3)
		r := stagedEnding(t, end, mapmodel.TierUnicode, true, false)
		v := &r.e.view
		verdicts := 0
		for _, l := range v.lines {
			if l.kind == elVerdict {
				verdicts++
			}
		}
		if verdicts == 0 || v.lines[0].kind != elVerdict || v.lines[verdicts].kind == elVerdict {
			t.Fatalf("%s: the verdict is not first: %d of %d lines", kind, verdicts, len(v.lines))
		}
		r.at(120, 40, 0)
		sc := r.e.sc
		s, _ := endSpanOf(sc, beatReckoning)
		if gap := sc.reck.at[verdicts] - sc.reck.at[verdicts-1]; gap != endVerdictHold {
			t.Errorf("%s: the verdict stands alone for %d frames, want %d", kind, gap, endVerdictHold)
		}
		// All through the verdict's time nothing sparks and nothing else is
		// said.
		for f := s.from; f < s.from+sc.reck.at[verdicts]; f++ {
			g := r.at(120, 40, f)
			if n := len(r.e.sc.sparks.a); n != 0 && f > s.from+endLead {
				t.Fatalf("%s: %d sparks over the verdict at frame %d", kind, n, f)
			}
			if pageHas(g, v.age) && f >= s.from+endLead {
				t.Fatalf("%s: the reckoning went on over the verdict at frame %d", kind, f)
			}
		}
		g := r.at(120, 40, s.from+sc.reck.at[verdicts-1]+1)
		if on := endScreen(g, false); !strings.Contains(on, endFold(r.e.sc.reck.rows[0].plain(), false)) {
			t.Errorf("%s: the verdict is not on the screen", kind)
		}
		// A prestige the Last Passage took the points of says so once: no
		// line counts up to nothing.
		for _, l := range v.lines {
			if kind == game.RunEndSuccumbed && l.kind == elCount && l.count == 0 {
				t.Errorf("%s: a line counts up to nothing", kind)
			}
		}
		r.e.close()
	}
}

// ---- keys and timers ----

// TestEndingKeys: any key skips to the next beat, at once; a key held
// down does not run through two; on the last beat a key closes the film
// and gives the keyboard back to the prompt. Esc is a key like any other.
func TestEndingKeys(t *testing.T) {
	for _, esc := range []bool{false, true} {
		r := stagedEnding(t, sampleEnding(game.RunEndPrestige, 3), mapmodel.TierUnicode, true, false)
		e := r.e
		press := func() {
			if esc {
				if !r.om.arrivalKey() {
					t.Fatal("the film did not take Esc")
				}
				return
			}
			e.InputHandler()(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone), func(tview.Primitive) {})
		}
		if r.om.ActiveName() != endingPageName || !r.pages.HasPage(endingPageName) || !r.om.FullScreenUp() {
			t.Fatal("the film is not up")
		}
		r.at(100, 30, 3)
		for _, want := range []int{beatReckoning, beatStrike, beatBeginning} {
			press()
			r.at(100, 30, e.pos())
			if s, f := e.sc.span(e.sc.frame); s.beat != want || f != 0 || e.beat().beat != want || len(e.sc.sparks.a) > 0 && want != beatBeginning {
				t.Fatalf("esc %v: a key went to beat %d frame %d, want the start of beat %d", esc, s.beat, f, want)
			}
			r.clock = r.clock.Add(arrivalKeyGap - time.Millisecond)
			press()
			if e.beat().beat != want {
				t.Fatalf("esc %v: a key held down ran through beat %d", esc, want)
			}
			r.clock = r.clock.Add(arrivalKeyGap)
		}
		if e.closed || r.closed != 0 {
			t.Fatal("the film closed before its last beat")
		}
		press()
		if !e.closed || r.pages.HasPage(endingPageName) || r.om.film != nil || r.om.ActiveName() != "" || r.closed != 1 || len(r.live()) != 0 || r.om.FullScreenUp() {
			t.Fatalf("esc %v: a key on the last beat did not close the film (%d timers left)", esc, len(r.live()))
		}
		r.clock = r.clock.Add(time.Second)
		press2 := e.keys
		e.key()
		e.close()
		if r.closed != 1 || e.keys != press2+1 || r.om.arrivalKey() {
			t.Error("a closed film answered a key")
		}
	}
}

// TestEndingPlaysThrough: left alone the film plays every beat and closes
// by itself when the last is over; its timers are a redraw as each beat
// ends and the close.
func TestEndingPlaysThrough(t *testing.T) {
	for _, motion := range []bool{true, false} {
		r := stagedEnding(t, sampleEnding(game.RunEndPrestige, 3), mapmodel.TierUnicode, motion, false)
		e := r.e
		r.at(100, 30, 0)
		spans := e.sc.spans
		live := r.live()
		if len(live) != len(spans) {
			t.Fatalf("motion %v: %d timers for %d beats", motion, len(live), len(spans))
		}
		for i, s := range spans {
			if want := time.Duration(s.end()) * mapAnimStep; live[i].d != want {
				t.Errorf("motion %v: timer %d waits %v, want %v", motion, i, live[i].d, want)
			}
			if !motion && s.frames != endStillFrames {
				t.Errorf("a still beat is %d frames", s.frames)
			}
		}
		for i, tm := range live {
			if e.closed {
				t.Fatalf("motion %v: the film closed before its last beat", motion)
			}
			r.clock = e.start.Add(tm.d)
			tm.fn()
			if i < len(spans)-1 {
				if g := r.at(100, 30, e.pos()); e.beat().beat != spans[i+1].beat || g == nil {
					t.Errorf("motion %v: after beat %d the film is in beat %d", motion, i, e.beat().beat)
				}
			}
		}
		if !e.closed || r.closed != 1 || r.pages.HasPage(endingPageName) {
			t.Errorf("motion %v: the film did not close by itself", motion)
		}
		live[len(live)-1].fn()
		if r.closed != 1 {
			t.Error("the close timer closed the film twice")
		}
	}
}

// TestEndingStill: with the motion setting off the film is one still
// frame a beat, each the same however long it is looked at, with no clock
// behind it, and a key moves to the next.
func TestEndingStill(t *testing.T) {
	r := stagedEnding(t, sampleEnding(game.RunEndPrestige, 3), mapmodel.TierUnicode, false, true)
	e := r.e
	t.Cleanup(e.close)
	const w, h = 100, 30
	r.om.app = tview.NewApplication()
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(w, h)
	e.SetRect(0, 0, w, h)
	var stills []string
	for i := 0; i < 4; i++ {
		first := r.at(w, h, i*endStillFrames)
		if s, _ := e.sc.span(e.sc.frame); s.beat != i || s.frames != endStillFrames {
			t.Fatalf("still %d is beat %d of %d frames", i, s.beat, s.frames)
		}
		for _, f := range []int{1, 9, endStillFrames - 1} {
			if later := r.at(w, h, i*endStillFrames+f); gridText(later) != gridText(first) || !sameInk(later, first) {
				t.Errorf("still %d changed after %d frames", i, f)
			}
		}
		if len(first.clipped) > 0 {
			t.Errorf("still %d: text cut off: %q", i, first.clipped)
		}
		stills = append(stills, gridText(first))
		e.Draw(theme.WrapScreen(sim))
		if e.stop != nil {
			t.Errorf("a clock runs behind still %d", i)
		}
	}
	for i := 1; i < len(stills); i++ {
		if stills[i] == stills[i-1] {
			t.Errorf("stills %d and %d are the same frame", i-1, i)
		}
	}
	// What each still shows: the town standing with the fire at its edges,
	// the reckoning all told, the word struck on a lit screen, the new
	// fire under the forged name.
	g := r.at(w, h, 0)
	if !pageHas(g, spaced(strings.ToUpper(e.view.age))) {
		t.Error("the first still does not name the age")
	}
	g = r.at(w, h, endStillFrames)
	for _, row := range e.sc.reck.rows {
		if want := endFold(row.plain(), false); !strings.Contains(endScreen(g, false), want) {
			t.Errorf("the reckoning's still is short of %q", want)
		}
	}
	g = r.at(w, h, 2*endStillFrames)
	if ironShort(g, e.sc.L.word, false) != 0 || litGround(g, e.pal) != w*h || !pageHas(g, e.view.complete) {
		t.Error("the strike's still is not the word whole on a lit screen with the game's line")
	}
	// Keys move through the stills.
	r.restart(mapmodel.TierUnicode, false)
	for i := 1; i < 4; i++ {
		e.key()
		if r.at(w, h, e.pos()); e.beat().beat != i {
			t.Fatalf("a key on still %d went to beat %d", i-1, e.beat().beat)
		}
		r.clock = r.clock.Add(arrivalKeyGap)
	}
	e.key()
	if !e.closed {
		t.Error("a key on the last still did not close the film")
	}
}

// TestEndingGivesWay: the manager's Hide closes the film; another overlay
// opened over it takes its place; an arrival and a film never stack.
func TestEndingGivesWay(t *testing.T) {
	r := stagedEnding(t, sampleEnding(game.RunEndPrestige, 2), mapmodel.TierUnicode, true, false)
	r.om.Hide()
	if !r.e.closed || r.pages.HasPage(endingPageName) || r.om.film != nil || r.closed != 1 || len(r.live()) != 0 {
		t.Fatalf("Hide left the film up (%d timers)", len(r.live()))
	}

	r = stagedEnding(t, sampleEnding(game.RunEndPrestige, 2), mapmodel.TierUnicode, true, false)
	first := r.e
	ShowRunEnding(r.om, sampleEnding(game.RunEndFallen, 2))
	if !first.closed || r.om.film == first || r.om.film == nil || r.pages.GetPageCount() != 1 || r.closed != 0 {
		t.Errorf("a second film did not take the first's place (%d pages)", r.pages.GetPageCount())
	}
	second := r.om.film
	ShowAgeSplashFull(r.om, "primitive_age", "stone_age", game.AgeAdvanceSummary{}, false, game.EpochEventRecord{})
	if !second.closed || r.om.film != nil || r.om.arrival == nil || r.om.ActiveName() != arrivalPageName || r.pages.GetPageCount() != 1 {
		t.Errorf("an arrival did not take the film's place (active %q)", r.om.ActiveName())
	}
	arrival := r.om.arrival
	ShowRunEnding(r.om, sampleEnding(game.RunEndPrestige, 3))
	if arrival.stage != arrClosed || r.om.arrival != nil || r.om.film == nil || r.om.ActiveName() != endingPageName {
		t.Errorf("a film did not take the arrival's place (active %q)", r.om.ActiveName())
	}
	r.om.film.close()
}

// ---- themes ----

// TestEndingInEveryTheme draws a frame of every beat on a terminal in
// every theme and both glyph sets: the words can be read (a contrast of 3
// or more on what is behind them), the word in iron stands off the page,
// and the plain set shows nothing it does not have.
func TestEndingInEveryTheme(t *testing.T) {
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	const w, h = 100, 30
	r := stagedEnding(t, sampleEnding(game.RunEndEndured, 5), mapmodel.TierUnicode, true, false)
	t.Cleanup(r.e.close)
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(w, h)
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
			r.restart(tier, true)
			r.at(w, h, 0)
			e := r.e
			plain := e.view.plain
			for _, sp := range e.sc.spans {
				f := sp.end() - 1
				if sp.beat == beatStrike {
					f = sp.from + endBlow + 2 // the screen still lit
				}
				g := r.at(w, h, f)
				where := fmt.Sprintf("%s, beat %d, %s glyphs", th.Key, sp.beat, tier)
				sim.Clear()
				g.flush(theme.WrapScreen(sim), 0, 0, e.pal, plain)
				rows := make([]string, h)
				for y := range rows {
					row := make([]rune, w)
					for x := range row {
						row[x], _, _, _ = sim.GetContent(x, y)
						if plain && row[x] > 0x7e {
							t.Fatalf("%s: cell (%d,%d) holds %q, which the plain glyph set does not have", where, x, y, row[x])
						}
					}
					rows[y] = string(row)
				}
				read := func(word string) {
					if word = strings.TrimSpace(word); word == "" {
						return
					}
					for y := 0; y < h; y++ {
						at := strings.Index(rows[y], word)
						if at < 0 {
							continue
						}
						x0 := len([]rune(rows[y][:at]))
						for i, c := range []rune(word) {
							if c == ' ' {
								continue
							}
							_, _, st, _ := sim.GetContent(x0+i, y)
							fg, bg, _ := st.Decompose()
							if k := theme.ContrastRatio(fg, bg); k < 3-0.02 {
								t.Errorf("%s: %q is drawn at a contrast of %.2f (%06x on %06x)", where, word, k, fg.Hex(), bg.Hex())
								return
							}
						}
						return
					}
					t.Errorf("%s: %q is not on the screen", where, word)
				}
				fold := func(s string) string {
					if !plain {
						return s
					}
					out := []rune(plainSafe(s))
					for i, c := range out {
						out[i] = mapmodel.Fold(c, mapmodel.TierASCII)
					}
					return string(out)
				}
				read(e.view.prompt)
				switch sp.beat {
				case beatEnding:
					read(spaced(strings.ToUpper(e.view.age)))
					for _, line := range e.sc.L.voice {
						read(line)
					}
				case beatReckoning:
					for _, row := range e.sc.reck.rows {
						read(fold(row.plain()))
					}
				case beatStrike:
					for _, line := range e.sc.L.caption {
						read(line)
					}
					p := e.sc.L.word
					_, _, st, _ := sim.GetContent(p.x+firstLit(&p), p.y)
					block, ground, _ := st.Decompose()
					if plain {
						block = ground
					}
					if k := theme.ContrastRatio(block, e.pal.bg); k < 3-0.02 {
						t.Errorf("%s: the word is drawn at a contrast of %.2f on the page", where, k)
					}
				case beatBeginning:
					read(fold(fitOption(e.view.captions, w-4)))
				}
			}
		}
	}
}

// TestEndingAmbientIsThePrestiges: a theme's own ambient effect plays from
// the heavy blow to the film's end, once the light lets it through, and
// nowhere else: not while the town burns, not over the reckoning, not in
// a fall's film, not with motion off.
func TestEndingAmbientIsThePrestiges(t *testing.T) {
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	const w, h = 100, 30
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(w, h)
	extra := func(r *endingRig, frames ...int) (n int) {
		r.e.SetRect(0, 0, w, h)
		for _, f := range frames {
			g := r.at(w, h, f)
			sim.Clear()
			r.e.Draw(theme.WrapScreen(sim))
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					if c, _, _, _ := sim.GetContent(x, y); c != g.c[y*w+x].r && !(c == ' ' && g.c[y*w+x].r == 0) {
						n++
					}
				}
			}
		}
		return n
	}
	film := stagedEnding(t, sampleEnding(game.RunEndPrestige, 2), mapmodel.TierUnicode, true, false)
	fall := stagedEnding(t, sampleEnding(game.RunEndFallen, 2), mapmodel.TierUnicode, true, false)
	t.Cleanup(film.e.close)
	t.Cleanup(fall.e.close)
	with := 0
	for _, th := range theme.All() {
		if th.Effect == "" {
			continue
		}
		with++
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		film.restart(mapmodel.TierUnicode, true)
		film.at(w, h, 0)
		strike, _ := endSpanOf(film.e.sc, beatStrike)
		var before, after []int
		for f := 0; f < strike.from+endBlow; f += 5 {
			before = append(before, f)
		}
		for f := strike.from + endBlow + 10; f < film.e.sc.total(); f++ {
			after = append(after, f)
		}
		if n := extra(film, before...); n != 0 {
			t.Errorf("%s: the effect plays before the heavy blow (%d cells)", th.Key, n)
		}
		if n := extra(film, after...); n == 0 {
			t.Errorf("%s: the effect never plays after the heavy blow", th.Key)
		}
		fall.restart(mapmodel.TierUnicode, true)
		fall.at(w, h, 0)
		var all []int
		for f := 0; f < fall.e.sc.total(); f += 4 {
			all = append(all, f)
		}
		if n := extra(fall, all...); n != 0 {
			t.Errorf("%s: the effect plays in a fall's film (%d cells)", th.Key, n)
		}
		film.restart(mapmodel.TierUnicode, false)
		if n := extra(film, 0, endStillFrames, 2*endStillFrames, 3*endStillFrames); n != 0 {
			t.Errorf("%s: the effect plays with motion off (%d cells)", th.Key, n)
		}
	}
	if with == 0 {
		t.Fatal("no theme has an ambient effect")
	}
}
