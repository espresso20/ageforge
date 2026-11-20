package ui

import (
	"github.com/espresso20/ageforge/window/kit"
	"image"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// pickCard picks the card of key with a click on its first row.
func pickCard(t *testing.T, d *Dashboard, st *stage, key string) {
	t.Helper()
	box := d.economyTab.buildBox
	pc := placedOf(t, d, key)
	x, y, _, _ := box.GetRect()
	st.click(x+3, y+pc.y)
	if box.picked != key {
		t.Fatalf("a click picked %q, want %q", box.picked, key)
	}
	d.refresh()
	st.paint()
}

func TestOnlyThePickedCardHasThePrimaryButton(t *testing.T) {
	d, st, _ := cardTown(t)
	box := d.economyTab.buildBox
	pickCard(t, d, st, "hut")
	three, rest := 0, 0
	for _, pc := range box.placed {
		k := box.cardKeys.keys[pc.build]
		switch {
		case k.r.Dy() == actionRows:
			three++
			if pc.card.key != "hut" || k.kind != keyPrimary {
				t.Errorf("%s has a three-row button of kind %v", pc.card.key, k.kind)
			}
		case k.r.Dy() == 1 && k.kind == keyAccent:
			rest++
		default:
			t.Errorf("%s has a button %v high of kind %v", pc.card.key, k.r.Dy(), k.kind)
		}
	}
	if three != 1 || rest != len(box.placed)-1 {
		t.Errorf("%d primary buttons and %d one-row keys for %d cards", three, rest, len(box.placed))
	}
}

func TestGatherButtonsStopAfterTheFirstTwoAges(t *testing.T) {
	d, _, _ := medievalTown(t, 160, 45)
	for _, k := range d.economyTab.buildBox.strip.keys {
		if strings.HasPrefix(k.cmd, "gather ") {
			t.Errorf("the Medieval Age strip has %q", k.cmd)
		}
	}
}

func TestThePickFollowsTheBuildingWhenTheListIsSortedAgain(t *testing.T) {
	d, st, eng := cardTown(t)
	box := d.economyTab.buildBox
	pickCard(t, d, st, "hut")
	for i := 0; i < 3; i++ {
		before := ownedAndQueued(eng, box.picked)
		runs := len(d.cmdHistory)
		d.submitInput()
		if len(d.cmdHistory) != runs+1 || !strings.HasPrefix(d.cmdHistory[runs], "build ") || ownedAndQueued(eng, strings.TrimPrefix(d.cmdHistory[runs], "build ")) <= before-1 {
			t.Fatalf("Enter on the empty prompt ran %v", d.cmdHistory[runs:])
		}
		d.refresh()
		st.paint()
		if box.picked != "hut" {
			t.Fatalf("after build %d the pick is %q, want hut", i+1, box.picked)
		}
		if at := slices.Index(box.order, box.picked); at != box.pickAt {
			t.Errorf("pick place %d, the card is at %d", box.pickAt, at)
		}
	}
	// Sorted in another order, the same card stays picked.
	slices.Reverse(box.order)
	box.resort = true
	box.recompute()
	if box.picked != "hut" || box.order[box.pickAt] != "hut" {
		t.Errorf("after a re-sort the pick is %q at %d", box.picked, box.pickAt)
	}
	// A card that left the list passes the pick to the one in its place.
	box.order = slices.DeleteFunc(box.order, func(k string) bool { return k == "hut" })
	box.settlePick()
	if box.picked == "hut" || box.picked == "" {
		t.Errorf("the pick is %q after its card left", box.picked)
	}
}

func TestDetailButtonsRunTheirCommandsAndOffOnesRunNothing(t *testing.T) {
	d, st, eng := cardTown(t)
	box := d.economyTab.buildBox
	d.runForTest("build hut")
	d.refresh()
	pickCard(t, d, st, "hut")
	if !box.detail.ok || box.det == nil {
		t.Fatal("the picked card has no detail box")
	}
	want := map[string]bool{"sell hut": false, "upgrade hut 1": false}
	for _, k := range box.det.keys {
		if _, ok := want[k.cmd]; ok {
			want[k.cmd] = true
		}
	}
	if !want["sell hut"] {
		t.Errorf("the detail has no sell button: %v", want)
	}
	// An off button runs nothing: the Upgrade of a building that is not an
	// older kind.
	for _, k := range box.det.keys {
		if k.cmd == "upgrade hut 1" {
			if !k.off {
				t.Error("Upgrade is on for a building that has no newer kind")
			}
			p := centre(d, k.r)
			runs := len(d.cmdHistory)
			st.click(p.X, p.Y)
			if len(d.cmdHistory) != runs {
				t.Errorf("an off Upgrade ran %v", d.cmdHistory[runs:])
			}
		}
	}
	if box.detail.upWhy == "" {
		t.Error("an off Upgrade gives no reason")
	}
	if text, bad := box.detailNote(xkey{}, false); !bad || text != box.detail.upWhy {
		t.Errorf("at rest the note is %q (bad %v), want the Upgrade's reason %q", text, bad, box.detail.upWhy)
	}
	_ = eng
}

// An off Sell runs nothing and says why, at rest and under the pointer.
func TestAnOffSellRunsNothingAndShowsItsReason(t *testing.T) {
	d, st, _ := cardTown(t)
	box := d.economyTab.buildBox
	d.runForTest("build hut")
	d.refresh()
	pickCard(t, d, st, "hut")
	var sell xkey
	for _, k := range box.det.keys {
		if strings.HasPrefix(k.cmd, "sell ") {
			sell = k
		}
	}
	if sell.cmd == "" || !sell.off {
		t.Fatalf("Sell is %+v, want a key that is off in this age", sell)
	}
	if box.detail.sellWhy == "" {
		t.Fatal("an off Sell gives no reason")
	}
	p := centre(d, sell.r)
	runs := len(d.cmdHistory)
	st.click(p.X, p.Y)
	if len(d.cmdHistory) != runs {
		t.Errorf("an off Sell ran %v", d.cmdHistory[runs:])
	}
	if text, bad := box.detailNote(sell, true); !bad || text != box.detail.sellWhy {
		t.Errorf("under the pointer the note is %q (bad %v), want %q", text, bad, box.detail.sellWhy)
	}
}

// A long list keeps the detail while a few cards still show beside it.
func TestALongListKeepsTheDetail(t *testing.T) {
	d, st, eng := medievalTown(t, 160, 58)
	box := d.economyTab.buildBox
	setStock(eng, 1e9)
	d.refresh()
	st.paint()
	keys := box.order
	if len(keys) == 0 {
		t.Fatal("no cards")
	}
	pickCard(t, d, st, placedKey(box))
	_, _, w, _ := box.GetRect()
	long := 0
	defer func() {
		if long == 0 {
			t.Error("no height made a list longer than the box")
		}
	}()
	for h := 30; h >= 20; h-- {
		box.drawAt(w, h, 0)
		if box.shownItems >= box.items {
			continue
		}
		long++
		box.draw(w, h)
		if h >= 26 && box.det == nil {
			t.Errorf("height %d: a long list lost the detail (%d cards shown)", h, box.shownItems)
		}
		if box.det != nil && box.shownItems < detailKeepCards {
			t.Errorf("height %d: %d cards beside the detail, want at least %d", h, box.shownItems, detailKeepCards)
		}
	}
}

func placedKey(b *buildBox) string { return b.placed[0].card.key }

func TestTheDetailGivesWayBeforeCardsAreHidden(t *testing.T) {
	d, st, _ := cardTown(t)
	box := d.economyTab.buildBox
	pickCard(t, d, st, "hut")
	if box.det == nil {
		t.Fatal("a tall box has no detail")
	}
	_, _, w, _ := box.GetRect()
	sawDetail, sawGone := false, false
	for h := 40; h >= 6; h-- {
		box.drawAt(w, h, 0)
		base, total := box.shownItems, box.items
		box.draw(w, h)
		if box.shownItems < base && (base >= total || box.shownItems < detailKeepCards) {
			t.Errorf("at height %d the detail hid cards: %d shown, %d without it of %d", h, box.shownItems, base, total)
		}
		if box.det != nil && base >= total && box.shownItems < base {
			t.Errorf("at height %d the detail cost a card of a list that fits", h)
		}
		sawDetail = sawDetail || box.det != nil
		sawGone = sawGone || box.det == nil
	}
	if !sawDetail || !sawGone {
		t.Errorf("the detail showed %v and gave way %v over the heights", sawDetail, sawGone)
	}
}

func TestOlderBuildingsRowListAndWayBack(t *testing.T) {
	d, st, eng := cardTown(t)
	box := d.economyTab.buildBox
	setStock(eng, 1e9)
	if err := eng.EnterAgeForTest("medieval_age"); err != nil {
		t.Fatal(err)
	}
	// After the advance: entering the era rolls an epoch event, and a great
	// fire destroys up to eight buildings, which would take all three huts.
	eng.SetTownForTest(map[string]int{"hut": 3}, 0)
	d.refresh()
	for d.overlayMgr != nil && d.overlayMgr.HasActive() {
		d.overlayMgr.Hide()
	}
	d.refresh()
	st.paint()
	if box.olderN == 0 || box.olderRow < 0 {
		t.Fatalf("the box counts %d older kinds on row %d", box.olderN, box.olderRow)
	}
	n := box.olderN
	if box.closeOlder() {
		t.Fatal("closeOlder reported a list that is not open")
	}
	pt := centre(d, image.Rect(2, box.olderRow, 3, box.olderRow+1))
	st.click(pt.X, pt.Y)
	st.paint()
	if !box.older || len(box.placed) != n {
		t.Fatalf("a click on the row shows %d cards (older %v), want %d", len(box.placed), box.older, n)
	}
	for _, pc := range box.placed {
		if !pc.card.older || pc.build >= 0 {
			t.Errorf("%s is listed with a Build button (older %v, build %d, olderN %d)", pc.card.key, pc.card.older, pc.build, n)
		}
	}
	pickCard(t, d, st, box.placed[0].card.key)
	if !box.detail.ok {
		t.Error("an older building shows no detail")
	}
	up := false
	for _, k := range box.det.keys {
		up = up || (k.cmd == "upgrade "+box.picked+" 1" && !k.off)
	}
	if !up {
		t.Errorf("the older %s has no Upgrade on (%q)", box.picked, box.detail.upWhy)
	}
	if !box.closeOlder() || box.older {
		t.Error("Esc does not return to the build list")
	}
	d.refresh()
	st.paint()
	for _, pc := range box.placed {
		if pc.card.older {
			t.Errorf("%s is still listed as older", pc.card.key)
		}
	}
}

func TestWorkersKeysRunAssignAndUnassign(t *testing.T) {
	d, st, eng := cardTown(t)
	box := d.economyTab.buildBox
	eng.SetTownForTest(map[string]int{"gathering_camp": 2, "hut": 4}, 3)
	d.refresh()
	st.paint()
	pickCard(t, d, st, "gathering_camp")
	if box.detail.workers == "" {
		t.Skip("the building has no workers line")
	}
	got := map[string]bool{}
	for _, k := range box.det.keys {
		got[k.cmd] = true
	}
	if !got["assign gathering_camp"] || !got["unassign gathering_camp"] {
		t.Errorf("the detail's keys carry %v", got)
	}
	// The plus key runs assign and no other command.
	for _, k := range box.det.keys {
		if k.cmd == "assign gathering_camp" && !k.off {
			p := centre(d, k.r)
			runs := len(d.cmdHistory)
			st.click(p.X, p.Y)
			if len(d.cmdHistory) != runs+1 || d.cmdHistory[runs] != "assign gathering_camp" {
				t.Errorf("a click on plus ran %v", d.cmdHistory[runs:])
			}
		}
	}
}

func TestSellRunsWithTheBuyAmountAndAnOffOneShowsItsReason(t *testing.T) {
	d, st, eng := cardTown(t)
	box := d.economyTab.buildBox
	eng.SetTownForTest(map[string]int{"hut": 12}, 0)
	d.refresh()
	st.paint()
	pickCard(t, d, st, "hut")
	for c, want := range map[buyChoice]string{buyOne: "sell hut", buyTen: "sell hut 10", buyMax: "sell hut 12"} {
		box.buy = c
		d.refresh()
		box.recompute()
		st.paint()
		found := false
		for _, k := range box.det.keys {
			if strings.HasPrefix(k.cmd, "sell ") {
				found = true
				if k.cmd != want {
					t.Errorf("buy %v: Sell carries %q, want %q", c, k.cmd, want)
				}
			}
		}
		if !found {
			t.Errorf("buy %v: no Sell key", c)
		}
	}
	// In the first age the engine refuses a sale, and the detail says why.
	box.buy = buyOne
	d.refresh()
	box.recompute()
	st.paint()
	if box.detail.sellOff && box.detail.sellWhy == "" {
		t.Error("an off Sell shows no reason")
	}
}

func TestAUniqueBuildingUnderConstructionIsBlocked(t *testing.T) {
	d, st, eng := cardTown(t)
	if err := eng.EnterAgeForTest("classical_age"); err != nil {
		t.Fatal(err)
	}
	setStock(eng, 1e9)
	d.refresh()
	for d.overlayMgr != nil && d.overlayMgr.HasActive() {
		d.overlayMgr.Hide()
	}
	if err := eng.BuildBuilding("cultural_obelisk"); err != nil {
		t.Fatal(err)
	}
	d.refresh()
	st.paint()
	var c buildCard
	for _, x := range d.economyTab.buildBox.cards {
		if x.key == "cultural_obelisk" {
			c = x
		}
	}
	if c.canBuild || !strings.Contains(c.reason, "under construction") {
		t.Errorf("a queued unique building has canBuild %v and reason %q", c.canBuild, c.reason)
	}
}

func TestTheResourceRowUnderThePointerIsLit(t *testing.T) {
	d, st, _ := cardTown(t)
	v := d.economyTab.resourceTV
	x, y, w, _ := v.GetInnerRect()
	st.paint()
	row := -1
	for i, l := range strings.Split(v.raw, "\n") {
		if strings.TrimSpace(kit.Strip(l)) != "" {
			row = i
			break
		}
	}
	if row < 0 {
		t.Fatal("the Resources box has no text")
	}
	bg := func() any { return st.buf.At(x+1, y+row).Bg }
	before := bg()
	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: x + 1, Y: y + row})
	st.paint()
	if bg() == before {
		t.Error("the row under the pointer is not lit")
	}
	st.mouse(kit.Mouse{Kind: kit.MouseLeave})
	st.paint()
	if bg() != before {
		t.Error("the row stays lit after the pointer leaves")
	}
	_ = w
}

// The question and the command name the same count, and it never exceeds the
// copies held: the engine sells no more than the town has.
func TestSellCountClampsToOwned(t *testing.T) {
	cases := []struct {
		buy   buyChoice
		owned int
		want  int
	}{{buyOne, 5, 1}, {buyTen, 5, 5}, {buyTen, 30, 10}, {buyMax, 7, 7}}
	for _, c := range cases {
		if got := sellCount(c.buy, c.owned); got != c.want {
			t.Errorf("sellCount(%v, %d) = %d, want %d", c.buy, c.owned, got, c.want)
		}
		cmd := sellCommand("farm", c.buy, c.owned)
		wantCmd := command(verbSell, "farm")
		if c.want > 1 {
			wantCmd = command(verbSell, "farm", strconv.Itoa(c.want))
		}
		if cmd != wantCmd {
			t.Errorf("sellCommand(%v, %d) = %q, want %q", c.buy, c.owned, cmd, wantCmd)
		}
	}
}

// bufRows is the text of the buffer, a string a row with the blanks trimmed
// off the end.
func bufRows(b *kit.Buffer) []string {
	rows := make([]string, b.H)
	for y := range rows {
		var sb strings.Builder
		for x := 0; x < b.W; x++ {
			sb.WriteRune(b.Cells[y*b.W+x].R)
		}
		rows[y] = strings.TrimRight(sb.String(), " ")
	}
	return rows
}

// detailTown is a first-age town on a w by h screen with the Gathering Camp
// picked, the detail's rows as text from its title to its bottom edge, and
// the box's detail keys.
func detailTown(t *testing.T, w, h int) (*Dashboard, *stage, []string) {
	t.Helper()
	d, st, _ := cardTown(t)
	st.p.SetRect(0, 0, w, h)
	st.root.Resize(w, h)
	d.refresh()
	st.paint()
	pickCard(t, d, st, "gathering_camp")
	box := d.economyTab.buildBox
	if box.det == nil {
		t.Fatalf("%dx%d: no detail", w, h)
	}
	bx, by, bw, bh := box.GetRect()
	var rows []string
	for y := by; y < by+bh-1; y++ {
		r := []rune(bufRows(&st.buf)[y])
		if bx+bw > len(r) {
			r = append(r, []rune(strings.Repeat(" ", bx+bw-len(r)))...)
		}
		rows = append(rows, string(r[bx:bx+bw]))
	}
	title := -1
	for y, r := range rows {
		if strings.Contains(r, box.detail.bs.Name+" x") {
			title = y
		}
	}
	if title < 0 {
		t.Fatal("no title row")
	}
	return d, st, rows[title:]
}

// Where the Buildings box has the rows (the foot strip takes 14 of the middle
// column's, so a 157-wide window needs 58 rows) every word of the detail's
// notice and description is on screen and no row of it ends in an ellipsis;
// the buttons are where they are drawn. A shorter box gives lines up in order
// (TestTheDetailGivesUpLinesInOrder).
func TestTheDetailReadsInFullWhereTheBoxHasRoom(t *testing.T) {
	d, st, rows := detailTown(t, 157, 58)
	box := d.economyTab.buildBox
	det := box.detail
	text := strings.Join(rows, " ")
	for _, r := range rows {
		if strings.Contains(r, "…") {
			t.Errorf("a row of the detail is cut: %q", r)
		}
	}
	for _, s := range []string{det.upWhy, buildingText(det.bs), det.bs.Flavor} {
		for _, word := range strings.Fields(s) {
			if !strings.Contains(text, word) {
				t.Errorf("the word %q of %q is not on screen", word, s)
			}
		}
	}
	// The buttons still hit where they are drawn: the pointer on Sell shows
	// the reason of Sell in full.
	for _, k := range box.det.keys {
		if k.cmd != det.sellCmd {
			continue
		}
		p := centre(d, k.r)
		st.mouse(kit.Mouse{Kind: kit.MouseMove, X: p.X, Y: p.Y})
		st.paint()
		got := strings.Join(bufRows(&st.buf), " ")
		for _, word := range strings.Fields(det.sellWhy) {
			if !strings.Contains(got, word) {
				t.Errorf("under the pointer the word %q of Sell's reason is not on screen", word)
			}
		}
		return
	}
	t.Error("no Sell key")
}

// When the rows run out the flavour line goes first, then the description is
// held to two rows with an ellipsis on the second; the title, the notice and
// the workers stay.
func TestTheDetailGivesUpLinesInOrder(t *testing.T) {
	d, _, _ := detailTown(t, 157, 46)
	box := d.economyTab.buildBox
	_, _, bw, _ := box.GetRect()
	p := box.planDetail(bw)
	if len(p.desc) < 2 || len(p.flav) == 0 || p.noteN == 0 || !p.workers {
		t.Fatalf("the test building has plan %+v", p)
	}
	base := 1 + p.noteN + 1
	full := p.fit(p.rows())
	if len(full.flav) == 0 || len(full.desc) != len(p.desc) {
		t.Errorf("with room for all, the plan lost lines: %+v", full)
	}
	noFlav := p.fit(p.rows() - 1)
	if len(noFlav.flav) != 0 || len(noFlav.desc) != len(p.desc) {
		t.Errorf("one row short, the flavour goes first: %+v", noFlav)
	}
	for rows := p.rows() - 1 - len(p.desc); rows >= base; rows-- {
		got := p.fit(rows)
		want := min(rows-base, 2)
		if len(got.flav) != 0 || len(got.desc) != want || got.noteN != p.noteN || !got.workers {
			t.Errorf("%d rows: %d description rows, want %d", rows, len(got.desc), want)
		}
	}
	if two := p.fit(base + 2); len(p.desc) > 2 && !strings.HasSuffix(two.desc[1], "…") {
		t.Errorf("a description held to two rows ends %q, want an ellipsis", two.desc[1])
	}
}

// In a wide window the detail is the four text rows in six frame rows.
func TestTheDetailInAWideWindowIsFourRows(t *testing.T) {
	d, _, rows := detailTown(t, 279, 74)
	box := d.economyTab.buildBox
	det := box.detail
	if len(rows) != detailRows-1 {
		t.Fatalf("the detail is %d rows from its title down, want %d: %q", len(rows), detailRows-1, rows)
	}
	want := []string{det.bs.Name + " x", buildingText(det.bs), "workers: up to 3 each", det.bs.Flavor}
	for i, s := range want {
		if !strings.Contains(rows[i], s) {
			t.Errorf("row %d is %q, want it to hold %q", i, rows[i], s)
		}
	}
	if !strings.Contains(rows[0], det.upWhy) {
		t.Errorf("the title row %q lacks the note %q", rows[0], det.upWhy)
	}
}
