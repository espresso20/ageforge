package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// arrivalSizes are the terminals the arrival screen is checked at.
var arrivalSizes = [][2]int{{80, 24}, {100, 30}, {120, 40}, {132, 43}, {144, 46}, {200, 60}, {90, 44}, {180, 26}}

// arrivalRig is an arrival screen on a stage: a clock the test holds and
// timers it fires by hand.
type arrivalRig struct {
	a     *arrival
	om    *OverlayManager
	clock time.Time
	// timers are the screen's timers, by how long each waits.
	timers map[time.Duration]func()
}

// fullSummary is an advance with upgrades on offer and buildings gone
// legacy, as a real one has.
func fullSummary() game.AgeAdvanceSummary {
	return game.AgeAdvanceSummary{
		BuildingsTransformed: []game.BuildingTransform{
			{OldName: "Elders' Hall", NewName: "Scriptorium", Count: 5},
			{OldName: "Forager Post", NewName: "Farm", Count: 4},
			{OldName: "Longhouse", NewName: "House", Count: 15},
			{OldName: "Standing Stones", NewName: "Altar", Count: 2},
			{OldName: "Stone Pit", NewName: "Quarry", Count: 5},
			{OldName: "War Camp", NewName: "Barracks", Count: 2},
			{OldName: "Woodcutter Camp", NewName: "Lumber Mill", Count: 4},
		},
		BuildingsLegacy: []string{"stash", "stone_camp"},
	}
}

// previousAge is the age before age, "" for the first.
func previousAge(age string) string {
	set := (&game.GameState{}).Ruleset()
	keys, at := set.AgeKeys(), set.Indexes()
	if i := at[age]; i > 0 {
		return keys[i-1]
	}
	return ""
}

// stagedArrival is the arrival screen for an advance from one age into the
// next, with fixture towns behind it. An advance that crosses into a new
// era is an epoch's, with an event.
func stagedArrival(t testing.TB, age string, tier mapmodel.GlyphTier, motion bool) *arrivalRig {
	t.Helper()
	return stageArrival(t, age, tier, motion, true, "")
}

// bareArrival is stagedArrival without the towns, whose pictures are most
// of what a frame costs to draw: the land is left as sky. Everything the
// screen itself lays out is as it is with them.
func bareArrival(t testing.TB, age string, tier mapmodel.GlyphTier, motion bool) *arrivalRig {
	t.Helper()
	return stageArrival(t, age, tier, motion, false, "")
}

// stageArrival: towns draws the towns' pictures, in the map style named
// ("" for the default).
func stageArrival(t testing.TB, age string, tier mapmodel.GlyphTier, motion, towns bool, style string) *arrivalRig {
	t.Helper()
	from := previousAge(age)
	set := (&game.GameState{}).Ruleset()
	epoch := from != "" && set.EraOf(from) != set.EraOf(age)
	var event game.EpochEventRecord
	if epoch {
		era, _ := set.Era(set.EraOf(age))
		event = game.EpochEventRecord{EpochKey: era.Key, EpochName: era.Name, EventKey: "golden_age", EventName: "A Golden Age", EventType: "good_major"}
	}
	r := &arrivalRig{om: &OverlayManager{}, clock: time.Unix(3_000_000, 0), timers: map[time.Duration]func(){}}
	a := newArrival(r.om, from, age, fullSummary(), epoch, event)
	a.now = func() time.Time { return r.clock }
	a.after = func(d time.Duration, fn func()) func() {
		r.timers[d] = fn
		return func() { delete(r.timers, d) }
	}
	cur := fixture.State(fixture.Options{Age: age, Seed: 7})
	var prev *game.GameState
	if from != "" {
		p := fixture.State(fixture.Options{Age: from, Seed: 7})
		prev = &p
	}
	if !towns {
		a.reg = nil
	}
	if style != "" {
		a.set.Style = style
	}
	a.feed(prev, &cur)
	a.set.Tier, a.set.Motion = tier, motion
	a.settings()
	a.begin()
	r.a = a
	return r
}

// at draws the screen at w by h, frames after it was begun.
func (r *arrivalRig) at(w, h, frame int) *mGrid {
	if a := r.a; a.sc != nil && (a.scKey[0] != w || a.scKey[1] != h) {
		a.frameGrid(w, h) // a new size starts the clock again from the frame it was at
	}
	r.clock = r.a.start.Add(time.Duration(frame-r.a.base) * mapAnimStep)
	return r.a.frameGrid(w, h)
}

// restart takes the screen back to its first frame, in a glyph set and
// with motion on or off, so one staged arrival serves many sizes (its two
// towns are what cost).
func (r *arrivalRig) restart(tier mapmodel.GlyphTier, motion bool) {
	a := r.a
	a.set.Tier, a.set.Motion = tier, motion
	a.settings()
	a.sc, a.info.pages = nil, nil
	a.stage, a.page, a.base = arrCelebrating, 0, 0
	a.quiet, a.start, a.keys, a.auto = time.Time{}, r.clock, 0, false
}

// TestArrivalDump writes the arrival screen's beats when ARRIVAL_DUMP names
// a folder: text for each, and one HTML page (arrival.html) in the themes
// ARRIVAL_DUMP_THEMES lists, for a look at them while working on them.
func TestArrivalDump(t *testing.T) {
	out := os.Getenv("ARRIVAL_DUMP")
	if out == "" {
		t.Skip("set ARRIVAL_DUMP to a folder to write the beats")
	}
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	themes := strings.Split(os.Getenv("ARRIVAL_DUMP_THEMES"), ",")
	if themes[0] == "" {
		themes = []string{theme.DefaultKey}
	}
	sizes := os.Getenv("ARRIVAL_DUMP_SIZES")
	if sizes == "" {
		sizes = "120x40,80x24"
	}
	var page strings.Builder
	page.WriteString("<!doctype html><meta charset=\"utf-8\"><title>Arrival</title><style>body{background:#222;color:#ddd;font:13px sans-serif}pre{font:10px/12px 'JetBrains Mono',Menlo,monospace;display:inline-block;margin:4px 12px 16px 0;vertical-align:top}h3{margin:10px 0 2px}</style>\n")
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
				age    string
				tier   mapmodel.GlyphTier
				motion bool
				frames []int
			}{
				{"bronze_age", mapmodel.TierUnicode, true, []int{0, 3, 5, 9, 14, 24, 30}},
				{"iron_age", mapmodel.TierUnicode, true, []int{2, 8, 13, 14, 18, 24, 31, 36, 50, 56}},
				{"transcendent_age", mapmodel.TierUnicode, true, []int{6, 24, 30}},
				{"bronze_age", mapmodel.TierUnicode, false, []int{0, 40}},
				{"iron_age", mapmodel.TierUnicode, false, []int{0, 60}},
				{"iron_age", mapmodel.TierASCII, true, []int{14, 36, 56}},
				{"renaissance_age", mapmodel.TierASCII, true, []int{14, 36, 56}},
			} {
				r := stagedArrival(t, c.age, c.tier, c.motion)
				for _, f := range c.frames {
					if !c.motion && f > 0 {
						r.a.inform()
					}
					g := r.at(w, h, f)
					name := fmt.Sprintf("%s_%s_%dx%d_motion%v_f%02d", key, c.age, w, h, c.motion, f)
					if c.tier == mapmodel.TierASCII {
						name += "_plain"
					}
					if err := os.WriteFile(filepath.Join(out, name+".txt"), []byte(gridText(g)), 0644); err != nil {
						t.Fatal(err)
					}
					fmt.Fprintf(&page, "<h3>%s</h3>\n%s", name, gridHTML(g, r.a.pal, r.a.view.plain))
				}
				r.a.close()
			}
		}
	}
	if err := os.WriteFile(filepath.Join(out, "arrival.html"), []byte(page.String()), 0644); err != nil {
		t.Fatal(err)
	}
}

// ---- the alphabet ----

// TestPixelAlphabetCoversEveryName: the alphabet is A to Z, the hyphen and
// the space, each seven rows of one width made of nothing but pixels; the
// six letters of the menu's wordmark are as they were; and every age's and
// every era's name in the game is set in it, and fits in iron on every
// terminal from 80 columns up.
func TestPixelAlphabetCoversEveryName(t *testing.T) {
	want := "ABCDEFGHIJKLMNOPQRSTUVWXYZ- "
	if len(menuFont) != len([]rune(want)) {
		t.Errorf("the alphabet has %d glyphs, want %d", len(menuFont), len([]rune(want)))
	}
	for _, r := range want {
		g, ok := menuFont[r]
		if !ok {
			t.Errorf("no glyph for %q", r)
			continue
		}
		wide := map[rune]int{'I': 4, 'M': 7, 'W': 7, '-': 4, ' ': 3}[r]
		if wide == 0 {
			wide = 6
		}
		lit := 0
		for y, row := range g {
			if len(row) != wide {
				t.Errorf("%q row %d is %d pixels wide, want %d", r, y, len(row), wide)
			}
			if strings.Trim(row, "#.") != "" {
				t.Errorf("%q row %d holds something that is not a pixel: %q", r, y, row)
			}
			lit += strings.Count(row, "#")
		}
		if (lit == 0) != (r == ' ') {
			t.Errorf("%q has %d pixels lit", r, lit)
		}
	}
	// The wordmark's letters, as the menu shipped them.
	for r, rows := range map[rune][7]string{
		'A': {".####.", "##..##", "##..##", "######", "##..##", "##..##", "##..##"},
		'G': {".#####", "##....", "##....", "##.###", "##..##", "##..##", ".#####"},
		'E': {"######", "##....", "##....", "#####.", "##....", "##....", "######"},
		'F': {"######", "##....", "##....", "#####.", "##....", "##....", "##...."},
		'O': {".####.", "##..##", "##..##", "##..##", "##..##", "##..##", ".####."},
		'R': {"#####.", "##..##", "##..##", "#####.", "##.##.", "##..##", "##..##"},
	} {
		if menuFont[r] != rows {
			t.Errorf("the wordmark's %q changed", r)
		}
	}
	if rows, ok := pixelWord("AGEFORGE"); !ok || rows != menuWord {
		t.Error("AGEFORGE set in the alphabet is not the menu's wordmark")
	}
	if _, ok := pixelWord("Stone Age 2"); ok {
		t.Error("a name with a character the alphabet lacks was set anyway")
	}

	set := (&game.GameState{}).Ruleset()
	var names []string
	for _, key := range set.AgeKeys() {
		def, _ := set.Age(key)
		names = append(names, def.Name)
	}
	for _, era := range set.Eras() {
		names = append(names, era.Name)
	}
	if len(names) < 29 {
		t.Fatalf("only %d names", len(names))
	}
	for _, name := range names {
		if _, ok := pixelWord(name); !ok {
			t.Errorf("%q cannot be set in the alphabet", name)
		}
		for _, size := range arrivalSizes {
			v := &arrivalView{age: name, era: name}
			L := arrivalLayoutFor(size[0], size[1], v)
			for what, p := range map[string]ironPlace{"an age": L.age, "an era": L.era} {
				if !p.ok {
					t.Errorf("%q as %s at %dx%d is not set in iron", name, what, size[0], size[1])
				}
				if p.x < 2 || p.x+p.w > size[0]-2 || p.y < 1 || p.y+p.h > size[1]-2 {
					t.Errorf("%q as %s at %dx%d sits at (%d,%d) %dx%d, off the screen's edge", name, what, size[0], size[1], p.x, p.y, p.w, p.h)
				}
			}
		}
	}
}

// ---- layout ----

// ironShort counts the cells of a name in iron that do not hold their
// piece of a letter on g: none, for a name that is whole.
func ironShort(g *mGrid, p ironPlace, plain bool) (short int) {
	if !p.ok {
		return -1
	}
	w := len(p.rows[0])
	holds := func(x, y int, r rune, bold bool) bool {
		if !g.in(x, y) {
			return false
		}
		c := g.c[y*g.w+x]
		return c.r == r && (!bold || c.bold)
	}
	for py, row := range p.rows {
		for dy := 0; dy < p.mode.sy; dy++ {
			y := p.y + py*p.mode.sy + dy
			if p.mode.half {
				for cx := 0; cx*2 < w; cx++ {
					l, r := row[cx*2] == '#', cx*2+1 < w && row[cx*2+1] == '#'
					switch {
					case l && !holds(p.x+cx, y, '▌', false), !l && r && !holds(p.x+cx, y, '▐', false):
						short++
					}
				}
				continue
			}
			for px := 0; px < w; px++ {
				if row[px] != '#' {
					continue
				}
				for dx := 0; dx < p.mode.sx; dx++ {
					x := p.x + px*p.mode.sx + dx
					if plain && !holds(x, y, '#', true) || !plain && !holds(x, y, '█', false) {
						short++
					}
				}
			}
		}
	}
	return short
}

// boxText is the text inside the information's box on a page: its rows
// joined, blanks squeezed.
func boxText(g *mGrid, info infoLayout) string {
	var rows []string
	for y := info.y + 1; y < info.y+info.h-1; y++ {
		row := []rune(g.row(y))
		rows = append(rows, strings.TrimSpace(string(row[info.x+1:info.x+info.w-1])))
	}
	return strings.Join(strings.Fields(strings.Join(rows, " ")), " ")
}

// checkArrivalInfo holds the information to the old splash's text: every
// line of it is on a page, whole, and nothing is cut off.
func checkArrivalInfo(t *testing.T, where string, r *arrivalRig, w, h int) {
	t.Helper()
	a := r.a
	if a.stage == arrCelebrating {
		a.inform()
	}
	g := r.at(w, h, a.sc.frame)
	if len(g.clipped) > 0 {
		t.Errorf("%s: text cut off at the edge: %q", where, g.clipped)
	}
	info := a.info
	if len(info.pages) == 0 {
		t.Fatalf("%s: the information has no page", where)
	}
	if info.x < 0 || info.y < 0 || info.x+info.w > w || info.y+info.h > h {
		t.Errorf("%s: the box (%d,%d %dx%d) is off the screen", where, info.x, info.y, info.w, info.h)
	}
	if info.named && info.y < a.sc.L.top {
		t.Errorf("%s: the box (from row %d) covers the lines under the name (to row %d)", where, info.y, a.sc.L.top)
	}
	if n := len(info.pages); n > 3 || w >= 100 && h >= 30 && n > 1 {
		t.Errorf("%s: the information takes %d pages", where, n)
	}
	all := ""
	for page := range info.pages {
		a.page = page
		g = r.at(w, h, a.sc.frame)
		if len(g.clipped) > 0 {
			t.Errorf("%s, page %d: text cut off at the edge: %q", where, page+1, g.clipped)
		}
		for _, at := range [][2]int{{info.x, info.y}, {info.x + info.w - 1, info.y}, {info.x, info.y + info.h - 1}, {info.x + info.w - 1, info.y + info.h - 1}} {
			if !strings.ContainsRune("┌┐└┘", g.c[at[1]*g.w+at[0]].r) {
				t.Errorf("%s, page %d: the box is broken at its corner (%d,%d)", where, page+1, at[0], at[1])
			}
		}
		for _, l := range info.pages[page] {
			if l.width() > info.w-2 {
				t.Errorf("%s, page %d: a line of %d cells in a box %d wide: %q", where, page+1, l.width(), info.w, l.plain())
			}
		}
		if len(info.pages[page])+2 > info.h {
			t.Errorf("%s, page %d: %d lines in a box of %d rows", where, page+1, len(info.pages[page]), info.h)
		}
		if len(info.pages) > 1 && !strings.Contains(g.row(info.y+info.h-1), fmt.Sprintf("%d/%d", page+1, len(info.pages))) {
			t.Errorf("%s, page %d: the page is not marked", where, page+1)
		}
		all += " " + boxText(g, info)
	}
	a.page = 0
	// What stands over the box, when the name does.
	screen := all
	if info.named {
		g = r.at(w, h, a.sc.frame)
		for y := 0; y < a.sc.L.top; y++ {
			screen += " " + strings.TrimSpace(g.row(y))
		}
		if short := ironShort(g, a.sc.L.age, a.view.plain); short != 0 {
			t.Errorf("%s: the name over the box is %d cells short", where, short)
		}
	}
	screen = strings.Join(strings.Fields(screen), " ")
	fold := func(s string) string {
		if a.view.plain {
			out := []rune(plainSafe(s))
			for i, r := range out {
				out[i] = mapmodel.Fold(r, mapmodel.TierASCII)
			}
			s = string(out)
		}
		return strings.Join(strings.Fields(s), " ")
	}
	if a.view.plain {
		out := []rune(screen)
		for i, r := range out {
			out[i] = mapmodel.Fold(r, mapmodel.TierASCII)
		}
		screen = string(out)
	}
	for _, l := range a.view.lines {
		text := fold(untag(l.text))
		switch l.kind {
		case skBlank, skRule:
			continue
		case skTitle:
			if info.named {
				continue // it is the name in iron
			}
		}
		if text != "" && !strings.Contains(screen, text) {
			t.Errorf("%s: a line of the old splash is not on the screen: %q\n%s", where, text, screen)
		}
	}
	if a.view.passed != "" && !strings.Contains(screen, fold(a.view.passed)) {
		t.Errorf("%s: the ages passed are not on the screen", where)
	}
}

// checkArrivalMoment holds a frame of the celebration to its layout.
func checkArrivalMoment(t *testing.T, where string, r *arrivalRig, w, h, frame int) {
	t.Helper()
	a := r.a
	g := r.at(w, h, frame)
	if a.stage != arrCelebrating {
		t.Fatalf("%s: frame %d is not of the celebration", where, frame)
	}
	if len(g.clipped) > 0 {
		t.Errorf("%s: text cut off at the edge: %q", where, g.clipped)
	}
	L := a.sc.L
	p := L.age
	if a.sc.inEra() {
		p = L.era
	}
	if short := ironShort(g, p, a.view.plain); short != 0 {
		t.Errorf("%s: the name is not whole in iron (%d)", where, short)
	}
	if p.tail != "" && !strings.Contains(g.row(p.y+p.h-1), spaced(strings.ToUpper(p.tail))) {
		t.Errorf("%s: the name's last word is not under it: %q", where, strings.TrimSpace(g.row(p.y+p.h-1)))
	}
	if !pageHas(g, a.view.prompt) {
		t.Errorf("%s: the line that says how to go on is not on the screen", where)
	}
	if a.sc.inEra() {
		if !pageHas(g, a.view.epochHeading) {
			t.Errorf("%s: the era's frame does not say it is a new epoch", where)
		}
		return
	}
	for i, line := range a.view.underLines(w - 6) {
		if !strings.Contains(g.row(L.under+i), line.text) {
			t.Errorf("%s: row %d should hold %q: %q", where, L.under+i, line.text, strings.TrimSpace(g.row(L.under+i)))
		}
	}
	if L.top+4 > h {
		t.Errorf("%s: the town has %d rows", where, h-L.top)
	}
	for _, chrome := range []string{"PgUp", "arrows move", "cursor hidden", "SETTLEMENT"} {
		if pageHas(g, chrome) {
			t.Errorf("%s: the map's own bars are on the screen (%q)", where, chrome)
		}
	}
}

// arrivalSweep is which terminals an age's arrival is checked at in a
// glyph set: every age at the five sizes the game is checked at (the plain
// set at its three smallest), and the ages that stretch the layout (the
// shortest name, the longest, and the advances that open an era) at all
// eight, the odd shapes among them.
func arrivalSweep(age string, tier mapmodel.GlyphTier) [][2]int {
	switch age {
	case "iron_age", "bronze_age", "renaissance_age", "interstellar_age", "transcendent_age":
		return arrivalSizes
	}
	if tier == mapmodel.TierASCII {
		return arrivalSizes[:3]
	}
	return arrivalSizes[:5]
}

// TestArrivalFitsEverySize plays the arrival into every age (an epoch's
// arrival for the six that open an era), at every size from 80x24 to
// 144x46 and three odd shapes, in both glyph sets, and fails on a name
// short of a letter, a line cut off or out of place, or a line of the old
// splash that is not on the information's pages. The towns are left out
// here (TestArrivalWithItsTowns draws them): nothing checked depends on
// them, and they are most of what a frame costs.
func TestArrivalFitsEverySize(t *testing.T) {
	set := (&game.GameState{}).Ruleset()
	capitals := 0
	for _, age := range set.AgeKeys()[1:] {
		r := bareArrival(t, age, mapmodel.TierUnicode, true)
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
			for _, size := range arrivalSweep(age, tier) {
				w, h := size[0], size[1]
				r.restart(tier, true)
				where := fmt.Sprintf("into the %s at %dx%d, %s glyphs", r.a.view.age, w, h, tier)
				if tier == mapmodel.TierASCII && w < 100 {
					// The plain set has no half cells: its narrow screens
					// have their own rules.
					if !checkArrivalPlainNarrow(t, where, r, w, h) && w == 80 {
						capitals++
					}
				} else {
					frames := []int{arrAgeStrike, arrAgeStrike + 6, arrAgeFrames - 1}
					if r.a.view.epoch {
						frames = []int{arrEraStrike, arrEraFrames + arrAgeStrike + 6}
					}
					for _, f := range frames {
						checkArrivalMoment(t, fmt.Sprintf("%s, frame %d", where, f), r, w, h, f)
					}
				}
				checkArrivalInfo(t, where+", the information", r, w, h)
				if t.Failed() {
					return
				}
			}
		}
		r.a.close()
	}
	if capitals != 2 {
		t.Errorf("%d ages are a line of capitals at 80x24 in the plain set, want the two longest", capitals)
	}
}

// TestArrivalWithItsTowns is the same check with the towns drawn, through
// the beats of an age's arrival and an epoch's at the five sizes: the town
// before, the two part way through the turn, the town after, and the town
// under the information, none of them bringing the map's own bars with it.
// The land always has its town, in either map style: where it has fewer
// rows than a style lays its picture out in (a small terminal, a quip that
// takes two lines, a name with its last word under it) it shows the foot
// of the picture, never an empty sky.
func TestArrivalWithItsTowns(t *testing.T) {
	for _, c := range []struct {
		age, style string
		sizes      [][2]int
	}{
		{"bronze_age", "", arrivalSizes[:5]},
		{"iron_age", "", arrivalSizes[:5]},
		// The cases with the least land: a quip on two lines, a plain name
		// with its last word under it, and the style that needs most rows.
		{"renaissance_age", "", arrivalSizes[:2]},
		{"bronze_age", "skyline", [][2]int{{80, 24}, {120, 40}}},
	} {
		r := stageArrival(t, c.age, mapmodel.TierUnicode, true, true, c.style)
		if r.a.oldTown == nil || r.a.newTown == nil {
			t.Fatalf("%s: no towns", c.age)
		}
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
			for _, size := range c.sizes {
				w, h := size[0], size[1]
				r.restart(tier, true)
				where := fmt.Sprintf("into the %s at %dx%d, %s glyphs, towns drawn (%s)", r.a.view.age, w, h, tier, r.a.set.Style)
				frames := []int{0, arrAgeStrike + 6, arrAgeFrames - 1}
				if r.a.view.epoch {
					frames = []int{0, arrEraStrike, arrEraFrames + arrAgeStrike + 6, arrEraFrames + arrAgeFrames - 1}
				}
				for _, f := range frames {
					if tier == mapmodel.TierASCII && w < 100 {
						if f != frames[len(frames)-1] {
							continue
						}
						checkArrivalPlainNarrow(t, where, r, w, h)
					} else {
						checkArrivalMoment(t, fmt.Sprintf("%s, frame %d", where, f), r, w, h, f)
					}
					// The town is there, under the name and clear of it:
					// two rows' worth of it at the least (a skyline is
					// mostly sky; stars and sparks alone are a few cells).
					g, L := r.at(w, h, f), r.a.sc.L
					from := L.top
					if r.a.sc.inEra() {
						from = max(from, L.era.y+L.era.h+1)
					}
					land := 0
					for y := from; y < h; y++ {
						land += len([]rune(strings.Join(strings.Fields(g.row(y)), "")))
					}
					if rows := h - from; rows < 4 || land < 2*w {
						t.Errorf("%s, frame %d: the land (%d rows) is all but empty: %d cells drawn", where, f, rows, land)
					}
				}
				checkArrivalInfo(t, where+", the information", r, w, h)
			}
		}
		r.a.close()
	}
}

// checkArrivalPlainNarrow: the plain glyph set has no half cells, so on a
// narrow screen a name too long for whole cells is set without its last
// word, which stands under it in capitals; the names too long even for
// that (Interstellar, Transcendent) are a line of capitals, and it reports
// false for those. The rest of the screen is as it is anywhere.
func checkArrivalPlainNarrow(t *testing.T, where string, r *arrivalRig, w, h int) (iron bool) {
	t.Helper()
	g := r.at(w, h, r.a.frames()-1)
	if len(g.clipped) > 0 {
		t.Errorf("%s: text cut off at the edge: %q", where, g.clipped)
	}
	p := r.a.sc.L.age
	switch {
	case p.ok && p.mode.half:
		t.Errorf("%s: the name is set in half cells", where)
	case p.ok:
		if short := ironShort(g, p, true); short != 0 {
			t.Errorf("%s: the name is %d cells short", where, short)
		}
		if p.tail != "" && !strings.Contains(g.row(p.y+p.h-1), spaced(strings.ToUpper(p.tail))) {
			t.Errorf("%s: the name's last word is not under it", where)
		}
	default:
		caps := strings.ToUpper(r.a.view.age)
		if !pageHas(g, caps) && !pageHas(g, spaced(caps)) {
			t.Errorf("%s: the name is not on the screen in capitals", where)
		}
	}
	for y := 0; y < g.h; y++ {
		for _, c := range g.row(y) {
			if c == '▌' || c == '▐' || c == '█' || c == '▀' {
				t.Fatalf("%s: row %d holds a block: %q", where, y, g.row(y))
			}
		}
	}
	return p.ok
}

// TestArrivalOnATerminalTooSmall: under 80x24 the screen is the name and
// the words, and no size at all makes it fail.
func TestArrivalOnATerminalTooSmall(t *testing.T) {
	for _, age := range []string{"bronze_age", "iron_age"} {
		r := bareArrival(t, age, mapmodel.TierUnicode, true)
		for _, w := range []int{1, 9, 30, 60, 79, 80} {
			for _, h := range []int{1, 4, 12, 23, 24} {
				r.restart(mapmodel.TierUnicode, true)
				for _, f := range []int{0, 14, 40} {
					g := r.at(w, h, f)
					caps := strings.ToUpper(r.a.sc.name().text)
					if r.a.stage == arrCelebrating && w >= 30 && h >= 4 && (w < menuMinW || h < menuMinH) && !pageHas(g, caps) && !pageHas(g, spaced(caps)) {
						t.Errorf("%s at %dx%d, frame %d: the name is not on the screen", age, w, h, f)
					}
				}
				r.a.key()
				r.at(w, h, 70)
			}
		}
		r.a.close()
	}
}

// ---- the script ----

// TestArrivalBeats holds the script to its beats: what happens at which
// frame, for an age and for an advance into a new epoch.
func TestArrivalBeats(t *testing.T) {
	if mapAnimStep != 125*time.Millisecond || arrivalFrames(false) != 26 || arrivalFrames(true) != 52 {
		t.Fatalf("the celebration is %d frames for an age and %d for an epoch, of %v", arrivalFrames(false), arrivalFrames(true), mapAnimStep)
	}
	// An age: the name heats, is struck with a burst of sparks at 0.625s,
	// the town turns over the 1.5s after, and the information follows at
	// 3.25s.
	r := stagedArrival(t, "bronze_age", mapmodel.TierUnicode, true)
	t.Cleanup(r.a.close)
	at := func(f int) *arrivalScene { r.at(120, 40, f); return r.a.sc }
	if sc := at(arrAgeStrike - 1); sc.heat <= 0 || sc.heat >= 1 || sc.strike > 0.01 || len(sc.sparks.a) != 0 || sc.turn != 0 {
		t.Errorf("before the strike: heat %.2f, strike %.2f, %d sparks, the town %.2f turned", sc.heat, sc.strike, len(sc.sparks.a), sc.turn)
	}
	if sc := at(arrAgeStrike); sc.heat != 1 || sc.strike < 0.8 || len(sc.sparks.a) < 50 || sc.turn != 0 {
		t.Errorf("at the strike: heat %.2f, strike %.2f, %d sparks, the town %.2f turned", sc.heat, sc.strike, len(sc.sparks.a), sc.turn)
	}
	if sc := at(arrAgeStrike + arrAgeDissolve/2); sc.turn <= 0.3 || sc.turn >= 0.7 {
		t.Errorf("half way through its turn the town is %.2f turned", sc.turn)
	}
	if sc := at(arrAgeStrike + arrAgeDissolve); sc.turn != 1 {
		t.Errorf("the town is %.2f turned when its turn is over", sc.turn)
	}
	if at(arrAgeFrames - 1); r.a.stage != arrCelebrating {
		t.Error("the celebration ended early")
	}
	if at(arrAgeFrames); r.a.stage != arrInforming {
		t.Error("the information did not follow the celebration")
	}
	if r.timers[3250*time.Millisecond] == nil || r.timers[arrivalHold] == nil || len(r.timers) != 2 {
		t.Errorf("an age's timers: %d, want one at 3.25s and one at %v", len(r.timers), arrivalHold)
	}

	// An epoch: two blows to build up (0.5s, 1s), the heavy one at 1.625s
	// with the whole screen lit, then the age's own beats from 3.25s and
	// the information at 6.5s.
	r = stagedArrival(t, "iron_age", mapmodel.TierUnicode, true)
	t.Cleanup(r.a.close)
	if sc := at(arrEraBlow1 - 1); !sc.inEra() || sc.strike > 0.01 || sc.flash != 0 || len(sc.sparks.a) != 0 {
		t.Errorf("before an epoch's first blow: strike %.2f, flash %.2f, %d sparks", sc.strike, sc.flash, len(sc.sparks.a))
	}
	first := 0
	if sc := at(arrEraBlow1); sc.strike < 0.3 || sc.flash != 0 || len(sc.sparks.a) < 10 {
		t.Errorf("an epoch's first blow: strike %.2f, flash %.2f, %d sparks", sc.strike, sc.flash, len(sc.sparks.a))
	} else {
		first = len(sc.sparks.a)
	}
	if sc := at(arrEraBlow2); sc.strike < 0.5 || sc.flash != 0 || len(sc.sparks.a) <= first {
		t.Errorf("an epoch's second blow: strike %.2f, flash %.2f, %d sparks after %d", sc.strike, sc.flash, len(sc.sparks.a), first)
	}
	if sc := at(arrEraStrike); sc.strike < 1 || sc.flash < 0.6 || len(sc.sparks.a) < 150 || sc.heat != 1 || sc.turn != 0 {
		t.Errorf("an epoch's heavy blow: strike %.2f, flash %.2f, %d sparks, heat %.2f", sc.strike, sc.flash, len(sc.sparks.a), sc.heat)
	}
	if sc := at(arrEraFrames - 1); !sc.inEra() || sc.name() != &sc.L.era {
		t.Error("the era's name gave way early")
	}
	if sc := at(arrEraFrames); sc.inEra() || sc.name() != &sc.L.age || sc.heat != 0 || sc.turn != 0 {
		t.Errorf("after the era the age's name starts cold, the town as it was: heat %.2f, turn %.2f", sc.heat, sc.turn)
	}
	if sc := at(arrEraFrames + arrAgeStrike); sc.heat != 1 || sc.strike < 0.8 || len(sc.sparks.a) < 50 {
		t.Errorf("the age's strike in an epoch's arrival: heat %.2f, strike %.2f, %d sparks", sc.heat, sc.strike, len(sc.sparks.a))
	}
	if sc := at(arrEraFrames + arrAgeStrike + arrAgeDissolve); sc.turn != 1 {
		t.Errorf("the town is %.2f turned when its turn is over", sc.turn)
	}
	if at(arrEraFrames + arrAgeFrames - 1); r.a.stage != arrCelebrating {
		t.Error("an epoch's celebration ended early")
	}
	if at(arrEraFrames + arrAgeFrames); r.a.stage != arrInforming {
		t.Error("the information did not follow an epoch's celebration")
	}
	if r.timers[6500*time.Millisecond] == nil || r.timers[arrivalHold] == nil || len(r.timers) != 2 {
		t.Errorf("an epoch's timers: %d, want one at 6.5s and one at %v", len(r.timers), arrivalHold)
	}
}

// litGround counts the cells of a frame that the flash has lit: a ground that is
// not the page's.
func litGround(g *mGrid, pal *menuPalette) (n int) {
	for _, c := range g.c {
		if c.hasBg && c.bg.Hex() != pal.bg.Hex() {
			n++
		}
	}
	return n
}

// TestAnEpochIsBiggerThanAnAge: an advance into a new epoch is told apart
// without reading a word. The era's name is twice the height of an age's
// and covers more of the screen, its celebration is twice as long, its
// heavy blow lights every cell of the screen where an age's strike lights
// none of the sky, and it throws more sparks.
func TestAnEpochIsBiggerThanAnAge(t *testing.T) {
	if arrivalFrames(true) < 2*arrivalFrames(false) {
		t.Errorf("an epoch's celebration is %d frames to an age's %d", arrivalFrames(true), arrivalFrames(false))
	}
	set := (&game.GameState{}).Ruleset()
	epochs := 0
	for _, age := range set.AgeKeys()[1:] {
		if set.EraOf(age) == set.EraOf(previousAge(age)) {
			continue
		}
		epochs++
		r := bareArrival(t, age, mapmodel.TierUnicode, true)
		if !r.a.view.epoch || r.a.view.era == "" || r.a.view.epochHeading == "" {
			t.Fatalf("the advance into %s is not an epoch's", age)
		}
		for _, size := range arrivalSizes {
			for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
				r.a.set.Tier = tier
				r.a.settings()
				L := arrivalLayoutFor(size[0], size[1], &r.a.view)
				where := fmt.Sprintf("%s at %dx%d, %s glyphs", r.a.view.era, size[0], size[1], tier)
				if !L.era.ok || L.era.mode.sy != 2 {
					t.Errorf("%s: the era's name is not set at twice the height (%+v)", where, L.era.mode)
				}
				if L.age.ok && L.era.w*L.era.h <= L.age.w*L.age.h {
					t.Errorf("%s: the era's name (%dx%d) is no bigger than the age's (%dx%d)", where, L.era.w, L.era.h, L.age.w, L.age.h)
				}
			}
		}
		r.a.close()
	}
	if epochs != 6 {
		t.Errorf("%d advances open an era, want 6", epochs)
	}

	era := stagedArrival(t, "iron_age", mapmodel.TierUnicode, true)
	age := stagedArrival(t, "bronze_age", mapmodel.TierUnicode, true)
	t.Cleanup(era.a.close)
	t.Cleanup(age.a.close)
	for _, size := range arrivalSizes {
		w, h := size[0], size[1]
		era.restart(mapmodel.TierUnicode, true)
		age.restart(mapmodel.TierUnicode, true)
		blow := era.at(w, h, arrEraStrike)
		if n := litGround(blow, era.a.pal); n != w*h {
			t.Errorf("at %dx%d an epoch's heavy blow lights %d of %d cells", w, h, n, w*h)
		}
		strike := age.at(w, h, arrAgeStrike)
		sky := 0
		for y := 0; y < age.a.sc.L.age.y; y++ {
			for x := 0; x < w; x++ {
				if c := strike.c[y*w+x]; c.hasBg && c.bg.Hex() != age.a.pal.bg.Hex() {
					sky++
				}
			}
		}
		if sky != 0 {
			t.Errorf("at %dx%d an age's strike lights %d cells of the sky", w, h, sky)
		}
		if e, a := len(era.a.sc.sparks.a), len(age.a.sc.sparks.a); e < 2*a {
			t.Errorf("at %dx%d an epoch's heavy blow throws %d sparks to an age's %d", w, h, e, a)
		}
	}
}

// ---- keys and timers ----

// arrivalShow is an overlay manager on a page stack, as the dashboard has
// one, with a clock and timers the test holds.
type arrivalShow struct {
	om     *OverlayManager
	pages  *tview.Pages
	clock  time.Time
	timers map[time.Duration]func()
	closed int // times the keyboard went back to the prompt
	// towns: draw the towns' pictures (they are most of what a screen
	// costs to make, and most tests here do not look at them).
	towns bool
}

func newArrivalShow() *arrivalShow {
	s := &arrivalShow{pages: tview.NewPages(), clock: time.Unix(3_000_000, 0), timers: map[time.Duration]func(){}}
	s.om = NewOverlayManager(s.pages, tview.NewApplication(), func() { s.closed++ })
	s.om.after = func(d time.Duration, fn func()) func() {
		s.timers[d] = fn
		return func() { delete(s.timers, d) }
	}
	return s
}

// advance shows the arrival the way the dashboard does: a refresh with the
// game as it was, the screen, then a refresh with the game as it is.
func (s *arrivalShow) advance(prev *game.GameState, cur game.GameState, epoch bool) *arrival {
	from := ""
	if prev != nil {
		s.om.Refresh(*prev)
		from = prev.Age
	}
	var event game.EpochEventRecord
	if epoch {
		set := cur.Ruleset()
		era, _ := set.Era(set.EraOf(cur.Age))
		event = game.EpochEventRecord{EpochKey: era.Key, EpochName: era.Name, EventKey: "golden_age", EventName: "A Golden Age", EventType: "good_major"}
	}
	ShowAgeSplashFull(s.om, from, cur.Age, fullSummary(), epoch, event)
	a := s.om.arrival
	a.now = func() time.Time { return s.clock }
	a.start = s.clock
	if !s.towns {
		a.reg = nil
	}
	s.om.Refresh(cur)
	return a
}

func (s *arrivalShow) press(a *arrival, k tcell.Key, r rune) {
	a.InputHandler()(tcell.NewEventKey(k, r, tcell.ModNone), func(tview.Primitive) {})
}

// stoneToBronze is the game before and after an advance into the Bronze
// Age, which opens no era.
func stoneToBronze() (*game.GameState, game.GameState, bool) {
	prev := fixture.State(fixture.Options{Age: "stone_age", Seed: 7})
	return &prev, fixture.State(fixture.Options{Age: "bronze_age", Seed: 7}), false
}

// TestArrivalKeys: any key moves on from the celebration to what the age
// opens, at once; a key held down does not run through both; the next
// press closes the screen and gives the keyboard back to the prompt. Esc
// is a key like any other. Closing twice is closing once.
func TestArrivalKeys(t *testing.T) {
	for _, key := range []struct {
		name string
		k    tcell.Key
		r    rune
	}{{"a letter", tcell.KeyRune, 'x'}, {"Enter", tcell.KeyEnter, 0}, {"space", tcell.KeyRune, ' '}, {"an arrow", tcell.KeyDown, 0}, {"Esc", tcell.KeyEsc, 0}} {
		s := newArrivalShow()
		s.towns = key.r == 'x'
		a := s.advance(stoneToBronze())
		press := func() {
			if key.k == tcell.KeyEsc {
				// Esc comes by way of the app, ahead of the dashboard.
				if !s.om.arrivalKey() {
					t.Errorf("%s: the screen did not take the key", key.name)
				}
				return
			}
			s.press(a, key.k, key.r)
		}
		if s.om.ActiveName() != arrivalPageName || !s.pages.HasPage(arrivalPageName) || a.stage != arrCelebrating {
			t.Fatalf("%s: the screen is not up on its celebration", key.name)
		}
		if s.towns && (a.oldTown == nil || a.newTown == nil || a.oldTown.age != "Stone Age" || a.newTown.age != "Bronze Age") {
			t.Fatalf("%s: the screen does not have the town before and after", key.name)
		}
		press()
		if a.stage != arrInforming || !s.pages.HasPage(arrivalPageName) || s.closed != 0 {
			t.Fatalf("%s: the first press did not move on to what the age opens (stage %d)", key.name, a.stage)
		}
		s.clock = s.clock.Add(arrivalKeyGap - time.Millisecond)
		press()
		if a.stage != arrInforming {
			t.Fatalf("%s: a key held down ran through the information", key.name)
		}
		s.clock = s.clock.Add(arrivalKeyGap)
		press()
		if a.stage != arrClosed || s.pages.HasPage(arrivalPageName) || s.om.ActiveName() != "" || s.om.arrival != nil {
			t.Fatalf("%s: the second press did not close the screen (stage %d)", key.name, a.stage)
		}
		if s.closed != 1 || len(s.timers) != 0 || a.stop != nil {
			t.Errorf("%s: closed %d times, %d timers left", key.name, s.closed, len(s.timers))
		}
		s.clock = s.clock.Add(time.Second)
		s.press(a, key.k, key.r)
		a.close()
		if s.om.arrivalKey() || s.closed != 1 {
			t.Errorf("%s: a screen that is closed answered a key", key.name)
		}
	}
}

// TestArrivalPages: where what the age opens takes more than one page, a
// key turns the page, and only the last page's key closes the screen.
func TestArrivalPages(t *testing.T) {
	set := (&game.GameState{}).Ruleset()
	for _, age := range set.AgeKeys()[1:] {
		prev := fixture.State(fixture.Options{Age: previousAge(age), Seed: 7})
		s := newArrivalShow()
		a := s.advance(&prev, fixture.State(fixture.Options{Age: age, Seed: 7}), set.EraOf(prev.Age) != set.EraOf(age))
		s.press(a, tcell.KeyEnter, 0)
		a.frameGrid(80, 24)
		pages := len(a.info.pages)
		if pages < 2 {
			a.close()
			continue
		}
		for p := 1; p < pages; p++ {
			s.clock = s.clock.Add(arrivalKeyGap)
			s.press(a, tcell.KeyEnter, 0)
			if a.stage != arrInforming || a.page != p {
				t.Fatalf("%s: a key on page %d of %d went to page %d, stage %d", age, p, pages, a.page+1, a.stage)
			}
			if g := a.frameGrid(80, 24); !pageHas(g, fmt.Sprintf("%d/%d", p+1, pages)) {
				t.Errorf("%s: page %d of %d is not marked", age, p+1, pages)
			}
		}
		s.clock = s.clock.Add(arrivalKeyGap)
		s.press(a, tcell.KeyEnter, 0)
		if a.stage != arrClosed {
			t.Fatalf("%s: a key on the last page did not close the screen", age)
		}
		return
	}
	t.Skip("nothing takes more than one page at 80x24")
}

// TestArrivalTimers: left alone, the celebration moves on to what the age
// opens by itself, and the screen closes by itself after arrivalHold. A
// timer that comes after a key has done its work does nothing. A key that
// lands just as the celebration ends by itself is let go by: it was pressed
// to skip the celebration, not to close what the age opens unread.
func TestArrivalTimers(t *testing.T) {
	s := newArrivalShow()
	a := s.advance(stoneToBronze())
	inform, hold := s.timers[time.Duration(arrivalFrames(false))*mapAnimStep], s.timers[arrivalHold]
	if inform == nil || hold == nil {
		t.Fatalf("the screen set %d timers", len(s.timers))
	}
	inform()
	if a.stage != arrInforming || !a.auto {
		t.Fatal("the celebration did not move on by itself")
	}
	inform()
	if a.stage != arrInforming || a.page != 0 {
		t.Error("the timer moved the information on")
	}
	s.press(a, tcell.KeyEnter, 0)
	s.clock = s.clock.Add(arrivalSettle - time.Millisecond)
	s.press(a, tcell.KeyEnter, 0)
	if a.stage != arrInforming || a.keys != 2 {
		t.Fatalf("a key as the celebration ended by itself closed the information (stage %d, %d keys)", a.stage, a.keys)
	}
	hold()
	if a.stage != arrClosed || s.pages.HasPage(arrivalPageName) || s.closed != 1 {
		t.Fatal("the screen did not close by itself")
	}
	hold()
	inform()
	if s.closed != 1 {
		t.Error("a timer after the screen closed did something")
	}

	// Once that moment has passed, a key closes it as any key does.
	s = newArrivalShow()
	a = s.advance(stoneToBronze())
	s.timers[time.Duration(arrivalFrames(false))*mapAnimStep]()
	s.clock = s.clock.Add(arrivalSettle)
	s.press(a, tcell.KeyEnter, 0)
	if a.stage != arrClosed {
		t.Error("a key after the celebration ended by itself did not close the information")
	}

	// A key that moved the celebration on is not the celebration ending by
	// itself, and the timer that follows changes nothing.
	s = newArrivalShow()
	a = s.advance(stoneToBronze())
	s.press(a, tcell.KeyEnter, 0)
	s.timers[time.Duration(arrivalFrames(false))*mapAnimStep]()
	if a.stage != arrInforming || a.auto || a.keys != 1 {
		t.Errorf("after a key: stage %d, by itself %v, %d keys", a.stage, a.auto, a.keys)
	}
	s.clock = s.clock.Add(arrivalKeyGap)
	s.press(a, tcell.KeyEnter, 0)
	if a.stage != arrClosed {
		t.Error("the second key did not close the screen")
	}
}

// TestArrivalGivesWay: the manager's Hide closes the screen whatever it is
// showing; another overlay opened over it takes its place and is not
// closed by it; and a second advance while the first's screen is up
// replaces it, so they never stack.
func TestArrivalGivesWay(t *testing.T) {
	s := newArrivalShow()
	a := s.advance(stoneToBronze())
	s.om.Hide()
	if a.stage != arrClosed || s.pages.HasPage(arrivalPageName) || s.om.arrival != nil || s.closed != 1 || len(s.timers) != 0 {
		t.Fatalf("Hide left the screen up: stage %d, closed %d, %d timers", a.stage, s.closed, len(s.timers))
	}

	s = newArrivalShow()
	s.om.Register("notes", "Notes", func(game.GameState, int) string { return "notes" })
	a = s.advance(stoneToBronze())
	_, cur, _ := stoneToBronze()
	if !s.om.Show("notes", cur) {
		t.Fatal("the other overlay did not open")
	}
	if a.stage != arrClosed || s.om.arrival != nil || len(s.timers) != 0 || s.pages.HasPage(arrivalPageName) {
		t.Errorf("another overlay opened over the screen and it stayed: stage %d, %d timers", a.stage, len(s.timers))
	}
	a.close()
	if s.om.ActiveName() != "notes" || !s.pages.HasPage("notes") || s.closed != 0 {
		t.Errorf("the screen closed the overlay that took its place (active %q)", s.om.ActiveName())
	}

	s = newArrivalShow()
	first := s.advance(stoneToBronze())
	bronze := fixture.State(fixture.Options{Age: "bronze_age", Seed: 7})
	second := s.advance(&bronze, fixture.State(fixture.Options{Age: "iron_age", Seed: 7}), true)
	if first.stage != arrClosed || first.stop != nil {
		t.Error("the first screen is still running under the second")
	}
	if s.om.arrival != second || second.stage != arrCelebrating || s.pages.GetPageCount() != 1 || s.closed != 0 {
		t.Errorf("two advances: %d pages, the keyboard went back to the prompt %d times", s.pages.GetPageCount(), s.closed)
	}
	if len(s.timers) != 2 || s.timers[time.Duration(arrivalFrames(true))*mapAnimStep] == nil {
		t.Errorf("two advances left %d timers, want the second screen's two", len(s.timers))
	}
	if second.view.age != "Iron Age" || !second.view.epoch || second.view.passed != "" {
		t.Errorf("the second screen is for %q (epoch %v, passed %q)", second.view.age, second.view.epoch, second.view.passed)
	}
	second.close()
}

// ---- motion off ----

// TestArrivalStill: with the motion setting off the celebration is one
// frame, there from the first draw and the same however long it is looked
// at: the name whole and hot, sparks in the air, the town already turned.
// An epoch's is the era's name with the screen lit, and names the age
// under it. No clock runs. The keys are the same.
func TestArrivalStill(t *testing.T) {
	for _, age := range []string{"bronze_age", "iron_age"} {
		for _, size := range [][2]int{{80, 24}, {120, 40}} {
			w, h := size[0], size[1]
			r := bareArrival(t, age, mapmodel.TierUnicode, false)
			where := fmt.Sprintf("into the %s at %dx%d", r.a.view.age, w, h)
			first := r.at(w, h, 0)
			sc := r.a.sc
			if r.a.stage != arrCelebrating || sc.turn != 1 || sc.heat != 1 || len(sc.sparks.a) < 20 {
				t.Errorf("%s: the still frame: stage %d, heat %.2f, the town %.2f turned, %d sparks", where, r.a.stage, sc.heat, sc.turn, len(sc.sparks.a))
			}
			p := sc.name()
			if short := ironShort(first, *p, false); short != 0 {
				t.Errorf("%s: the name is not whole in iron (%d)", where, short)
			}
			if c := first.c[p.y*w+p.x+firstLit(p)]; theme.ContrastRatio(c.fg, r.a.pal.bg) < 3 {
				t.Errorf("%s: the name is drawn at a contrast of %.2f", where, theme.ContrastRatio(c.fg, r.a.pal.bg))
			}
			if r.a.view.epoch {
				if p != &sc.L.era || !pageHas(first, r.a.view.epochHeading) || !pageHas(first, r.a.view.age) {
					t.Errorf("%s: an epoch's still frame is the era's name, its heading and the age's name", where)
				}
				if n := litGround(first, r.a.pal); n != w*h {
					t.Errorf("%s: an epoch's still frame lights %d of %d cells", where, n, w*h)
				}
			} else if !pageHas(first, r.a.view.desc) {
				t.Errorf("%s: the age's line is not under its name", where)
			}
			if len(first.clipped) > 0 || !pageHas(first, r.a.view.prompt) {
				t.Errorf("%s: cut off %q, or no line on how to go on", where, first.clipped)
			}
			for _, f := range []int{1, 9, 400} {
				later := r.at(w, h, f)
				if gridText(later) != gridText(first) || !sameInk(later, first) {
					t.Errorf("%s: the still frame changed after %d frames", where, f)
				}
			}

			// No clock runs behind a still frame.
			r.om.app = tview.NewApplication()
			sim := tcell.NewSimulationScreen("UTF-8")
			if err := sim.Init(); err != nil {
				t.Fatal(err)
			}
			sim.SetSize(w, h)
			r.a.SetRect(0, 0, w, h)
			r.a.Draw(theme.WrapScreen(sim))
			if r.a.stop != nil {
				t.Errorf("%s: a clock runs behind the still frame", where)
			}
			sim.Fini()

			// The keys are the same, and so is the information.
			r.a.key()
			info := r.at(w, h, 0)
			if r.a.stage != arrInforming || len(r.a.info.pages) == 0 {
				t.Fatalf("%s: a key on the still frame did not move on", where)
			}
			if r.a.info.named && ironShort(info, r.a.sc.L.age, false) != 0 {
				t.Errorf("%s: the age's name is not over the information", where)
			}
			if again := r.at(w, h, 400); gridText(again) != gridText(info) || !sameInk(again, info) {
				t.Errorf("%s: the information moves with motion off", where)
			}
			r.clock = r.clock.Add(arrivalKeyGap)
			r.a.key()
			if len(r.a.info.pages) == 1 && r.a.stage != arrClosed {
				t.Errorf("%s: the second key did not close the screen", where)
			}
			r.a.close()
		}
	}
}

// TestArrivalClock: a moving screen redraws on a clock of its own, started
// by its first draw and stopped when it closes.
func TestArrivalClock(t *testing.T) {
	r := stagedArrival(t, "bronze_age", mapmodel.TierUnicode, true)
	r.om.app = tview.NewApplication()
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(100, 30)
	r.a.SetRect(0, 0, 100, 30)
	if r.a.stop != nil {
		t.Fatal("the clock runs before the screen is drawn")
	}
	r.a.Draw(theme.WrapScreen(sim))
	stop := r.a.stop
	if stop == nil {
		t.Fatal("no clock runs behind the celebration")
	}
	r.a.Draw(theme.WrapScreen(sim))
	if r.a.stop != stop {
		t.Error("a second draw started a second clock")
	}
	r.a.close()
	if r.a.stop != nil {
		t.Error("the clock outlived the screen")
	}
	select {
	case <-stop:
	default:
		t.Error("the clock was not told to stop")
	}
	r.a.Draw(theme.WrapScreen(sim)) // a closed screen draws nothing and starts nothing
	if r.a.stop != nil {
		t.Error("a closed screen started a clock")
	}
}

// firstLit is the first lit pixel of a name's top row, in cells.
func firstLit(p *ironPlace) int {
	at := strings.IndexByte(p.rows[0], '#')
	if p.mode.half {
		return at / 2
	}
	return at * p.mode.sx
}

// sameInk reports whether two frames are coloured alike, cell for cell.
func sameInk(a, b *mGrid) bool {
	if len(a.c) != len(b.c) {
		return false
	}
	for i := range a.c {
		x, y := a.c[i], b.c[i]
		if x.r != y.r || x.fg.Hex() != y.fg.Hex() || x.hasBg != y.hasBg || x.hasBg && x.bg.Hex() != y.bg.Hex() || x.bold != y.bold {
			return false
		}
	}
	return true
}

// ---- away ----

// TestArrivalAfterBeingAway: when more than one age went by since the
// player last saw the game, there is one screen, for the furthest age.
// The town it starts from is the one they left, the ages passed on the way
// are listed, and if an era was entered on the way it is an epoch's
// arrival, with the event that era brought.
func TestArrivalAfterBeingAway(t *testing.T) {
	set := (&game.GameState{}).Ruleset()
	iron, _ := set.Era("iron_era")
	event := game.EpochEventRecord{EpochKey: iron.Key, EpochName: iron.Name, EventKey: "golden_age", EventName: "A Golden Age", EventType: "good_major"}

	// The save was in the Stone Age when it was loaded; nothing was seen
	// of it since, and it is in the Classical Age now. The dashboard saw
	// only the last hop, inside the Iron Era, so it says no epoch.
	cur := fixture.State(fixture.Options{Age: "classical_age", Seed: 7})
	cur.SessionStart = &game.SessionMark{Tick: 10, SavedAt: time.Unix(1_000, 0), Age: "stone_age"}
	cur.EpochEventHistory = []game.EpochEventRecord{event}
	s := newArrivalShow()
	s.towns = true
	ShowAgeSplashFull(s.om, "iron_age", "classical_age", fullSummary(), false, game.EpochEventRecord{})
	a := s.om.arrival
	a.now = func() time.Time { return s.clock }
	s.om.Refresh(cur)
	v := &a.view
	if v.age != "Classical Age" || v.passed != "Ages passed: Bronze Age, Iron Age" {
		t.Errorf("after being away: the screen is for %q, %q", v.age, v.passed)
	}
	if !v.epoch || v.era != iron.Name || a.event != event {
		t.Errorf("after being away through a new era: epoch %v, era %q, event %+v", v.epoch, v.era, a.event)
	}
	if a.from != "stone_age" || a.oldTown == nil || a.oldTown.age != "Stone Age" || a.newTown == nil || a.newTown.age != "Classical Age" {
		t.Error("after being away: the town does not turn from the age left into the age reached")
	}
	if s.pages.GetPageCount() != 1 {
		t.Errorf("after being away there are %d screens", s.pages.GetPageCount())
	}
	// The ages passed are on the celebration and on the information.
	rig := &arrivalRig{a: a, om: s.om, clock: s.clock, timers: s.timers}
	a.now = func() time.Time { return rig.clock }
	for _, size := range arrivalSizes {
		w, h := size[0], size[1]
		rig.restart(mapmodel.TierUnicode, true)
		g := rig.at(w, h, arrivalFrames(true)-1)
		where := fmt.Sprintf("after being away, at %dx%d", w, h)
		if a.stage != arrCelebrating || len(g.clipped) > 0 {
			t.Errorf("%s: stage %d, cut off %q", where, a.stage, g.clipped)
		}
		on := ""
		for y := 0; y < h; y++ {
			on += " " + strings.TrimSpace(g.row(y))
		}
		if !strings.Contains(strings.Join(strings.Fields(on), " "), v.passed) {
			t.Errorf("%s: the ages passed are not on the celebration", where)
		}
		if short := ironShort(g, a.sc.L.age, false); short != 0 {
			t.Errorf("%s: the name is not whole in iron (%d)", where, short)
		}
		checkArrivalInfo(t, where, rig, w, h)
	}
	a.close()

	// Seen all the way: the last snapshot is of this sitting, one age
	// back. Nothing was passed, and it is an age's arrival.
	prev := fixture.State(fixture.Options{Age: "iron_age", Seed: 7})
	prev.SessionStart = cur.SessionStart
	s = newArrivalShow()
	a = s.advance(&prev, cur, false)
	if a.view.passed != "" || a.view.epoch || a.from != "iron_age" {
		t.Errorf("an advance seen all the way: passed %q, epoch %v, from %s", a.view.passed, a.view.epoch, a.from)
	}
	a.close()

	// Two advances between two looks in one sitting: the last snapshot is
	// two ages back.
	prev = fixture.State(fixture.Options{Age: "bronze_age", Seed: 7})
	prev.SessionStart = cur.SessionStart
	s = newArrivalShow()
	a = s.advance(&prev, cur, false)
	if a.view.passed != "Ages passed: Iron Age" || !a.view.epoch || a.from != "bronze_age" {
		t.Errorf("two advances between two looks: passed %q, epoch %v, from %s", a.view.passed, a.view.epoch, a.from)
	}
	a.close()

	// A snapshot of another sitting (the save was loaded since) says
	// nothing about this one: the age the save was loaded in does.
	prev = fixture.State(fixture.Options{Age: "primitive_age", Seed: 7})
	prev.SessionStart = &game.SessionMark{Tick: 3, SavedAt: time.Unix(500, 0), Age: "primitive_age"}
	now := cur
	now.SessionStart = &game.SessionMark{Tick: 10, SavedAt: time.Unix(1_000, 0), Age: "iron_age"}
	s = newArrivalShow()
	a = s.advance(&prev, now, false)
	if a.view.passed != "" || a.view.epoch || a.from != "iron_age" {
		t.Errorf("a snapshot of another sitting was believed: passed %q, epoch %v, from %s", a.view.passed, a.view.epoch, a.from)
	}
	a.close()

	// With nothing to go on, it is the age before.
	bare := fixture.State(fixture.Options{Age: "classical_age", Seed: 7})
	bare.SessionStart = nil
	s = newArrivalShow()
	a = s.advance(nil, bare, false)
	if a.view.passed != "" || a.view.epoch || a.from != "iron_age" {
		t.Errorf("with nothing to go on: passed %q, epoch %v", a.view.passed, a.view.epoch)
	}
	a.close()
}

// ---- themes ----

// TestArrivalInEveryTheme draws an age's strike, an epoch's heavy blow and
// the information on a terminal in every theme and both glyph sets: the
// words can be read (a contrast of 3 or more on what is behind them), the
// name can, the page stands on the theme's ground, and the plain set shows
// nothing it does not have.
func TestArrivalInEveryTheme(t *testing.T) {
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	if n := len(theme.All()); n < 16 {
		t.Fatalf("%d themes", n)
	}
	const w, h = 100, 30
	age := stagedArrival(t, "bronze_age", mapmodel.TierUnicode, true)
	era := stagedArrival(t, "iron_age", mapmodel.TierUnicode, true)
	t.Cleanup(age.a.close)
	t.Cleanup(era.a.close)
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
			for _, shot := range []struct {
				name  string
				r     *arrivalRig
				frame int
				info  bool
			}{
				{"an age's strike", age, arrAgeStrike + 2, false},
				{"an age, turned", age, arrAgeFrames - 1, false},
				{"an age's information", age, 0, true},
				{"an epoch's heavy blow", era, arrEraStrike, false},
				{"an epoch, after the blow", era, arrEraStrike + 8, false},
				{"an epoch's information", era, 0, true},
			} {
				r := shot.r
				where := fmt.Sprintf("%s, %s, %s glyphs", th.Key, shot.name, tier)
				var g *mGrid
				if shot.info {
					r.a.key()
					g = r.at(w, h, r.a.frames())
				} else {
					r.restart(tier, true)
					g = r.at(w, h, shot.frame)
				}
				a := r.a
				sim.Clear()
				g.flush(theme.WrapScreen(sim), 0, 0, a.pal, a.view.plain)
				rows := make([]string, h)
				for y := range rows {
					row := make([]rune, w)
					for x := range row {
						row[x], _, _, _ = sim.GetContent(x, y)
						if a.view.plain && row[x] > 0x7e {
							t.Fatalf("%s: cell (%d,%d) holds %q, which the plain glyph set does not have", where, x, y, row[x])
						}
					}
					rows[y] = string(row)
				}
				if a.pal.bg.Hex() != th.Color(theme.RoleBackground).Hex() {
					t.Errorf("%s: the page is not on the theme's ground", where)
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
					if !a.view.plain {
						return s
					}
					out := []rune(plainSafe(s))
					for i, c := range out {
						out[i] = mapmodel.Fold(c, mapmodel.TierASCII)
					}
					return string(out)
				}
				if shot.info {
					for _, l := range a.info.pages[0] {
						if l.kind != skBlank && l.kind != skRule {
							read(fold(l.plain()))
						}
					}
				} else {
					read(a.view.prompt)
					p := a.sc.name()
					if a.sc.inEra() {
						read(a.view.epochHeading)
					} else {
						for _, l := range a.view.underLines(w - 6) {
							read(l.text)
						}
					}
					// The name on the page: a letter's cell is a block of
					// its ink (in the plain set, a ground of it with a mark
					// on it). The heavy blow's own frame is all light.
					x, y := p.x+firstLit(p), p.y
					_, _, st, _ := sim.GetContent(x, y)
					block, ground, _ := st.Decompose()
					if a.view.plain {
						block = ground
					}
					if k := theme.ContrastRatio(block, a.pal.bg); k < 3-0.02 && shot.frame != arrEraStrike {
						t.Errorf("%s: the name is drawn at a contrast of %.2f (%06x on %06x)", where, k, block.Hex(), a.pal.bg.Hex())
					}
				}
			}
		}
	}
}

// TestArrivalAmbientIsAnEpochs: a theme's own ambient effect plays in the
// sky of an epoch's celebration once the heavy blow has landed and its
// flash has gone, and nowhere else: not in an age's, not before the blow,
// not on the information, not with motion off.
func TestArrivalAmbientIsAnEpochs(t *testing.T) {
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	const w, h = 100, 30
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(w, h)
	// extra counts the cells the screen shows that the page did not draw.
	extra := func(r *arrivalRig, frames ...int) (n int) {
		r.a.SetRect(0, 0, w, h)
		for _, f := range frames {
			g := r.at(w, h, f)
			sim.Clear()
			r.a.Draw(theme.WrapScreen(sim))
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
	era := bareArrival(t, "iron_age", mapmodel.TierUnicode, true)
	age := bareArrival(t, "bronze_age", mapmodel.TierUnicode, true)
	t.Cleanup(era.a.close)
	t.Cleanup(age.a.close)
	// Once the flash has gone: every third frame, and the four in which
	// the era's name gives way (where the glitch bursts).
	after := []int{arrEraFrames, arrEraFrames + 1, arrEraFrames + 2, arrEraFrames + 3}
	for f := arrEraStrike + 12; f < arrivalFrames(true); f += 3 {
		after = append(after, f)
	}
	sort.Ints(after)
	with := 0
	for _, th := range theme.All() {
		if th.Effect == "" {
			continue
		}
		with++
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		era.restart(mapmodel.TierUnicode, true)
		if n := extra(era, 0, arrEraBlow2, arrEraStrike-1); n != 0 {
			t.Errorf("%s: the effect plays before an epoch's heavy blow (%d cells)", th.Key, n)
		}
		if n := extra(era, after...); n == 0 {
			t.Errorf("%s: the effect never plays in an epoch's celebration", th.Key)
		}
		era.a.key()
		if n := extra(era, arrivalFrames(true)+1); n != 0 {
			t.Errorf("%s: the effect plays on the information (%d cells)", th.Key, n)
		}
		age.restart(mapmodel.TierUnicode, true)
		if n := extra(age, 0, arrAgeStrike, arrAgeStrike+8, arrAgeFrames-1); n != 0 {
			t.Errorf("%s: the effect plays in an age's celebration (%d cells)", th.Key, n)
		}
		era.restart(mapmodel.TierUnicode, false)
		if n := extra(era, 0, 30); n != 0 {
			t.Errorf("%s: the effect plays with motion off (%d cells)", th.Key, n)
		}
	}
	if with == 0 {
		t.Fatal("no theme has an ambient effect")
	}
}
