package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

var _ = mapstyle.Frame{}

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
	a.feed(prev, &cur)
	a.set.Tier, a.set.Motion = tier, motion
	a.settings()
	a.begin()
	r.a = a
	return r
}

// at draws the screen at w by h, frames after it was begun.
func (r *arrivalRig) at(w, h, frame int) *mGrid {
	if a := r.a; a.sc == nil || a.scKey[0] != w || a.scKey[1] != h {
		a.frameGrid(w, h) // the screen's clock starts at its first draw at a size
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
	a.lastKey, a.start = time.Time{}, r.clock
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

// TestArrivalFitsEverySize plays the arrival into every age (an epoch's
// arrival for the six that open an era), at every size from 80x24 to
// 144x46 and three odd shapes, in both glyph sets, and fails on a name
// short of a letter, a line cut off or out of place, or a line of the old
// splash that is not on the information's pages.
func TestArrivalFitsEverySize(t *testing.T) {
	set := (&game.GameState{}).Ruleset()
	capitals := 0
	for _, age := range set.AgeKeys()[1:] {
		r := stagedArrival(t, age, mapmodel.TierUnicode, true)
		for _, size := range arrivalSizes {
			w, h := size[0], size[1]
			for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
				r.restart(tier, true)
				where := fmt.Sprintf("into the %s at %dx%d, %s glyphs", r.a.view.age, w, h, tier)
				if tier == mapmodel.TierASCII && w < 100 {
					// The plain set has no half cells: its narrow screens
					// have their own rules.
					if !checkArrivalPlainNarrow(t, where, r, w, h) && w == 80 {
						capitals++
					}
				} else {
					frames := []int{0, arrAgeStrike, arrAgeStrike + 6, arrAgeFrames - 1}
					if r.a.view.epoch {
						frames = []int{0, arrEraStrike, arrEraFrames - 1, arrEraFrames + arrAgeStrike, arrEraFrames + arrAgeStrike + 6}
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
		r := stagedArrival(t, age, mapmodel.TierUnicode, true)
		for _, w := range []int{1, 9, 30, 60, 79, 80} {
			for _, h := range []int{1, 4, 12, 23, 24} {
				r.restart(mapmodel.TierUnicode, true)
				for _, f := range []int{0, 14, 40} {
					g := r.at(w, h, f)
					caps := strings.ToUpper(r.a.sc.name().text)
					if w >= 30 && h >= 4 && (w < menuMinW || h < menuMinH) && !pageHas(g, caps) && !pageHas(g, spaced(caps)) {
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
