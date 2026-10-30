package ui

import (
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// lateHallState is a run well on its way: it has been through every age up
// to the one named, with a few badges earned.
func lateHallState(t *testing.T, age string) game.GameState {
	t.Helper()
	engine := game.NewGameEngine()
	must(t, engine.SummonHarbingerForTest(age))
	st := engine.GetState()
	st.Stats.AgesReached = nil
	st.Stats.AgeLog = nil
	for i, a := range st.Ruleset().Ages() {
		st.Stats.AgesReached = append(st.Stats.AgesReached, a.Key)
		st.Stats.AgeLog = append(st.Stats.AgeLog, game.AgeEntry{Age: a.Key, Tick: i * 900})
		if a.Key == age {
			break
		}
	}
	st.CatastrophesEndured = 2
	st.AccountStats = &game.AccountStatsView{Badges: []game.BadgeView{
		{Key: "a", Name: "Common One", Rarity: "common", Points: 1, Earned: true},
		{Key: "b", Name: "Epic Two", Rarity: "epic", Points: 20, Earned: true, Emblem: "star"},
		{Key: "c", Name: "Rare Three", Rarity: "rare", Points: 10, Earned: true},
		{Key: "d", Name: "Legend Four", Rarity: "legendary", Points: 50, Earned: true},
		{Key: "e", Name: "Uncommon Five", Rarity: "uncommon", Points: 3, Earned: true},
		{Key: "f", Name: "Unearned Six", Rarity: "mythic", Points: 99},
		{Key: "g", Name: "Common Seven", Rarity: "common", Points: 1, Earned: true},
	}}
	return st
}

// unseenAges are the names of the ages the player may not see yet.
func unseenAges(st game.GameState) []string {
	sight := game.SightOf(&st)
	var out []string
	for _, a := range st.Ruleset().Ages() {
		if !sight.Age(a.Key) {
			out = append(out, a.Name)
		}
	}
	return out
}

func TestEpochHallFrontPage(t *testing.T) {
	early := game.NewGameEngine().GetState()
	late := lateHallState(t, "galactic_age")
	for name, st := range map[string]game.GameState{"early": early, "late": late} {
		hidden := unseenAges(st)
		if len(hidden) == 0 {
			t.Fatalf("%s: setup: nothing is unseen", name)
		}
		for _, dim := range [][2]int{{200, 50}, {80, 24}} {
			txt := untag(hallFrontText(st, dim[0], dim[1], 0))
			if strings.Contains(txt, "…") {
				t.Errorf("%s %dx%d: prose cut off:\n%s", name, dim[0], dim[1], txt)
			}
			for _, a := range hidden {
				if strings.Contains(txt, a) {
					t.Errorf("%s %dx%d: names the unseen age %q", name, dim[0], dim[1], a)
				}
			}
			wantAll(t, name, txt, "Highlights", "Ages", "Top badges", "Feats", "more to come")
		}
	}
	txt := untag(hallFrontText(early, 120, 40, 0))
	wantAll(t, "early", txt, "No badges earned yet.")
	txt = untag(hallFrontText(late, 120, 40, 0))
	wantAll(t, "late", txt, "Legend Four", "Epic Two", "Rare Three", "Uncommon Five", "Common One")
	if strings.Contains(txt, "Unearned Six") || strings.Contains(txt, "Common Seven") {
		t.Errorf("late: shows more than the five rarest earned:\n%s", txt)
	}
	if strings.Index(txt, "Legend Four") > strings.Index(txt, "Epic Two") {
		t.Errorf("late: badges are not rarest first")
	}
}

func TestEpochHallTabSwitchesPages(t *testing.T) {
	h := &epochHall{}
	st := game.NewGameEngine().GetState()
	front := h.provider(st, 120)
	if !h.handleKey(wkey(tcell.KeyTab, 0)) || !h.details {
		t.Fatal("Tab did not open the details page")
	}
	if details := h.provider(st, 120); details == front || details != epochProvider(st, 120) {
		t.Error("the details page is not the old panel")
	}
	h.handleKey(wkey(tcell.KeyTab, 0))
	if h.details || h.provider(st, 120) != front {
		t.Error("Tab did not return to the front page")
	}
	h.handleKey(wkey(tcell.KeyTab, 0))
	h.open()
	if h.details {
		t.Error("open should start on the front page")
	}
	if h.handleKey(wkey(tcell.KeyDown, 0)) {
		t.Error("scroll keys belong to the text view")
	}
}

func TestEpochCardCommandWritesFile(t *testing.T) {
	dir := t.TempDir()
	defer game.SetDataDirForTest(dir)()
	var copied string
	was := epochCardCopier
	epochCardCopier = func(p string) error { copied = p; return nil }
	defer func() { epochCardCopier = was }()
	engine := game.NewGameEngine()
	res := HandleCommand("epoch card", engine)
	if res.Type != "success" || !strings.Contains(res.Message, "epoch-card.png") {
		t.Fatalf("reply = %+v", res)
	}
	if _, err := os.Stat(copied); err != nil || filepath.Base(copied) != "epoch-card.png" {
		t.Errorf("the copier was given %q: %v", copied, err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(copied), "epoch-card.txt")); err == nil {
		t.Error("the text card is written again")
	}
	if res := HandleCommand("epoch", engine); res.OverlayName != "epoch" {
		t.Errorf("bare epoch should open the screen: %+v", res)
	}
}

// firstBox is the span of the first line that opens a box (a tile of the
// totals row): its first and last used column.
func firstBox(t *testing.T, txt string) (first, last int) {
	t.Helper()
	for _, l := range strings.Split(untag(txt), "\n") {
		if !strings.Contains(l, "╭") {
			continue
		}
		// Stars may share the row; the tiles are from the first corner to
		// the last.
		rs := []rune(l)
		first, last = -1, 0
		for i, r := range rs {
			if r == '╭' && first < 0 {
				first = i
			}
			if r == '╮' {
				last = i
			}
		}
		return first, last
	}
	t.Fatal("no tile row on the page")
	return 0, 0
}

func TestEpochHallTotalsAreCentred(t *testing.T) {
	late := lateHallState(t, "galactic_age")
	for name, st := range map[string]game.GameState{"early": game.NewGameEngine().GetState(), "late": late} {
		for _, dim := range [][2]int{{200, 50}, {400, 100}, {100, 30}} {
			w := overlayTextWidth(dim[0])
			first, last := firstBox(t, hallFrontText(st, dim[0], dim[1], 0))
			// The row's middle against the page's, in half cells.
			if d := first + last + 1 - w; d < -4 || d > 4 {
				t.Errorf("%s %dx%d: totals row spans %d..%d on a page %d wide, off centre", name, dim[0], dim[1], first, last, w)
			}
		}
	}
}

func TestEpochHallBannerHasManyColours(t *testing.T) {
	for _, mo := range []bool{false, true} {
		c := newHallCtx(196, 50, 0, hallMotion{on: mo, t: 3})
		c.banner("Stone Era", 0, 1)
		seen := map[any]bool{}
		for _, cell := range c.g.c {
			if cell.r == '█' || cell.r == '▌' || cell.r == '▐' {
				seen[cell.fg] = true
			}
		}
		if len(seen) < 2 {
			t.Errorf("motion %v: the banner is one colour (%d)", mo, len(seen))
		}
	}
}

func TestEpochHallMotion(t *testing.T) {
	st := lateHallState(t, "galactic_age")
	a := hallFront(st, 200, 50, 0, hallMotion{t: 1})
	b := hallFront(st, 200, 50, 0, hallMotion{t: 97.3})
	if a != b {
		t.Error("with motion off the page changed between two frames")
	}
	if on := hallFront(st, 200, 50, 0, hallMotion{on: true, t: 1}); on == hallFront(st, 200, 50, 0, hallMotion{on: true, t: 5}) {
		t.Error("with motion on the page does not move")
	}
}

func TestEpochHallPlainTierIsASCII(t *testing.T) {
	st := lateHallState(t, "galactic_age")
	for _, r := range untag(hallFrontText(st, 200, 50, mapmodel.TierASCII)) {
		if r > 0x7e && r != '\n' {
			t.Fatalf("the plain tier drew %q", r)
		}
	}
}

func TestEpochHallMonochromeHasNoHue(t *testing.T) {
	was := theme.Active().Key
	defer func() { _ = theme.SetActive(was) }()
	if err := theme.SetActive("monochrome"); err != nil {
		t.Skip(err)
	}
	hexRe := regexp.MustCompile(`\[#([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})[:\]]`)
	for _, m := range hexRe.FindAllStringSubmatch(hallFrontText(lateHallState(t, "galactic_age"), 200, 50, 0), -1) {
		if m[1] != m[2] || m[2] != m[3] {
			t.Fatalf("a hue in Monochrome: #%s%s%s", m[1], m[2], m[3])
		}
	}
}

// TestEpochHallDump writes the page's text for the render script when
// HALL_DUMP names a directory.
func TestEpochHallDump(t *testing.T) {
	dir := os.Getenv("HALL_DUMP")
	if dir == "" {
		t.Skip("HALL_DUMP not set")
	}
	must(t, os.MkdirAll(dir, 0o755))
	early := game.NewGameEngine().GetState()
	early.AccountStats = &game.AccountStatsView{Badges: []game.BadgeView{
		{Key: "a", Name: "First Fire", Rarity: "common", Points: 1, Earned: true, Level: config.BadgeBronze},
		{Key: "b", Name: "Hands Full", Rarity: "uncommon", Points: 3, Earned: true, Level: config.BadgeSilver},
		{Key: "c", Name: "Quarry Master", Rarity: "rare", Points: 10, Earned: true, Level: config.BadgeGold},
	}}
	late := lateHallState(t, "galactic_age")
	for i, b := range late.AccountStats.Badges {
		late.AccountStats.Badges[i].Level = []config.BadgeTier{config.BadgeBronze, config.BadgePlatinum, config.BadgeGold, config.BadgeLegendary, config.BadgeSilver, config.BadgeLegendary, config.BadgeBronze}[i]
		_ = b
	}
	early.Workers.TotalPop = 14
	early.Research.TotalResearched = 6
	late.Workers.TotalPop = 48210
	late.Research.TotalResearched = 171
	late.Prestige.Level = 7
	late.Buildings = map[string]game.BuildingState{"x": {Count: 1284}}
	for name, st := range map[string]game.GameState{"early": early, "late": late} {
		for _, dim := range [][2]int{{400, 100}, {200, 50}, {100, 30}} {
			txt := hallFront(st, dim[0], dim[1], 0, hallMotion{on: true, t: 2})
			must(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s-%dx%d.txt", name, dim[0], dim[1])), []byte(txt), 0o644))
		}
	}
}

func TestEpochHallBigPointIsLit(t *testing.T) {
	for _, sx := range []int{1, 2} {
		rows, ok := bigNumber("48.2K", sx)
		if !ok {
			t.Fatalf("sx %d: no glyph for 48.2K", sx)
		}
		// Two digits and their gaps come first, then the point.
		at := 2*3*sx + 2*sx
		base := []rune(rows[2])
		if base[at] != '▀' {
			t.Errorf("sx %d: no lit cell for the point at %d on the baseline: %q", sx, at, rows[2])
		}
		if n := len([]rune(rows[0])); n != len(base) {
			t.Errorf("sx %d: rows differ in width", sx)
		}
	}
	// On the plain tier the point is a '#'.
	c := newHallCtx(196, 50, mapmodel.TierASCII, hallMotion{})
	c.y = 1
	c.tiles([]hallStat{{value: "48.2K", big: "48.2K", label: "Population"}})
	txt := untag(c.render())
	if strings.ContainsAny(txt, "█▀▄") {
		t.Errorf("plain tile still has block glyphs:\n%s", txt)
	}
}

func TestEpochHallTilesShareAHeightAndTheBarSitsInside(t *testing.T) {
	late := lateHallState(t, "galactic_age")
	for _, dim := range [][2]int{{200, 50}, {100, 30}, {100, 24}} {
		lines := strings.Split(untag(hallFrontText(late, dim[0], dim[1], 0)), "\n")
		top, bottom := -1, -1
		for i, l := range lines {
			if top < 0 && strings.Contains(l, "╭") {
				top = i
			}
			if top >= 0 && strings.Contains(l, "╰") {
				bottom = i
				break
			}
		}
		if top < 0 || bottom < 0 {
			t.Fatalf("%v: no tile row", dim)
		}
		if strings.ContainsAny(lines[bottom], "━╌") {
			t.Errorf("%v: progress bar is on the border row: %q", dim, lines[bottom])
		}
		if !strings.ContainsAny(lines[bottom-1], "━╌") {
			t.Errorf("%v: no progress bar just above the border: %q", dim, lines[bottom-1])
		}
	}
}

func TestEpochHallCardKey(t *testing.T) {
	var made int
	now := time.Unix(1000, 0)
	var res epochCardResult
	var cerr error
	h := &epochHall{
		now:   func() time.Time { return now },
		state: func() game.GameState { return game.GameState{} },
		card: func(game.GameState) (epochCardResult, error) {
			made++
			return res, cerr
		},
	}
	press := func(r rune) bool { return h.handleKey(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)) }

	res = epochCardResult{Path: "/x/epoch-card.png", Copied: true}
	if !press('c') || made != 1 {
		t.Fatalf("c did not make the card (made %d)", made)
	}
	if got := h.notice(); got != "Picture copied. Paste it into Discord." {
		t.Errorf("copied notice = %q", got)
	}
	now = now.Add(hallNoteFor + time.Second)
	if h.notice() != "" {
		t.Error("the notice did not expire")
	}

	res = epochCardResult{Path: "/x/epoch-card.png"}
	press('c')
	if got := h.notice(); got != "Picture saved to /x/epoch-card.png." {
		t.Errorf("file-only notice = %q", got)
	}
	if !press('x') || h.notice() != "" {
		t.Error("the next key did not clear the notice")
	}
	if press('x') {
		t.Error("an idle key was taken")
	}

	cerr = errors.New("Could not save the picture: disk full.")
	press('c')
	if h.notice() != cerr.Error() {
		t.Errorf("error notice = %q", h.notice())
	}
}

func TestMakeEpochCardClipboard(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	st := game.NewGameEngine().GetState()
	var got string
	res, err := makeEpochCard(st, func(path string) error { got = path; return nil })
	if err != nil || !res.Copied {
		t.Fatalf("copy ok: %+v, %v", res, err)
	}
	if got != res.Path || filepath.Ext(got) != ".png" {
		t.Errorf("the copier was given %q, the file is %q", got, res.Path)
	}
	f, err := os.Open(res.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("the file is not a PNG: %v", err)
	}
	if b := img.Bounds(); b.Dx() != epochPicCols*shotCellW+2*shotPad || (b.Dy()-2*shotPad)%shotCellH != 0 || b.Dy() < epochPicMinRows*shotCellH {
		t.Errorf("picture is %v", b)
	}
	res, err = makeEpochCard(st, func(string) error { return errNoClipboard })
	if err != nil || res.Copied || res.Path == "" {
		t.Errorf("clipboard failing: %+v, %v", res, err)
	}
}

func TestClipboardImageCommandsQuoteThePath(t *testing.T) {
	path := `/Users/a b/it's "x"/epoch-card.png`
	mac := clipboardImageCommands("darwin", path)[0].argv
	if mac[0] != "osascript" || !strings.Contains(mac[2], `/Users/a b/it's \"x\"/epoch-card.png`) || !strings.Contains(mac[2], "PNGf") {
		t.Errorf("macOS command: %q", mac)
	}
	win := clipboardImageCommands("windows", path)[0].argv
	if win[0] != "powershell" || !strings.Contains(win[len(win)-1], `it''s "x"`) {
		t.Errorf("Windows command: %q", win)
	}
	lin := clipboardImageCommands("linux", path)
	if lin[0].argv[0] != "wl-copy" || !lin[0].stdin || lin[1].argv[0] != "xclip" || lin[1].argv[len(lin[1].argv)-1] != path {
		t.Errorf("Linux commands: %+v", lin)
	}
}

func TestCopyImageWithMissingTool(t *testing.T) {
	if err := copyImageWith([]imageCommand{{argv: []string{"definitely-not-a-tool"}}}, "x"); err != errNoClipboard {
		t.Errorf("missing tool: %v", err)
	}
}

func TestEpochHallCardKeyHandsTheCopierAPNGPath(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	var path string
	var copyErr error
	now := time.Unix(1000, 0)
	h := &epochHall{
		now:   func() time.Time { return now },
		state: func() game.GameState { return game.NewGameEngine().GetState() },
		card: func(st game.GameState) (epochCardResult, error) {
			return makeEpochCard(st, func(p string) error { path = p; return copyErr })
		},
	}
	press := func() { h.handleKey(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone)) }
	press()
	if filepath.Ext(path) != ".png" || h.notice() != "Picture copied. Paste it into Discord." {
		t.Errorf("path %q, notice %q", path, h.notice())
	}
	copyErr = errNoClipboard
	press()
	if want := "Picture saved to " + path + "."; h.notice() != want {
		t.Errorf("notice %q, want %q", h.notice(), want)
	}
}
