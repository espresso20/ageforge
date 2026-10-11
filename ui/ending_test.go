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

var _ = tcell.KeyEnter
