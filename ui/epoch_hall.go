package ui

import (
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// epoch_hall.go is the Epoch screen: a page to open and show off a run. Its
// front page has the current epoch's name in iron (the arrival screen's
// lettering, in the epoch's colour), the town under it, the run's headline
// figures, the ages done and ahead, the account's five rarest badges and the
// feats the game already keeps. Tab turns to the details page, which is the
// panel this screen replaced (epochProvider). The Discord card (epoch_card.go)
// is the same facts as plain text.
//
// Everything here follows the no-leak rule of the old panel: an age the
// player has not seen is counted, never named.

// epochHall is the screen's state: which page it is on, and what it reads
// from the dashboard.
type epochHall struct {
	details bool
	// tier is the glyph tier the player draws in (the badge marks and the
	// plain set); height is the screen's rows. Either may be nil.
	tier   func() mapmodel.GlyphTier
	height func() int
}

// open starts the screen on the front page.
func (h *epochHall) open() { h.details = false }

// back is Esc: the screen always closes.
func (h *epochHall) back() bool { return false }

// handleKey turns the page on Tab and reports whether it took the key; the
// rest (scrolling) is the text view's.
func (h *epochHall) handleKey(ev *tcell.EventKey) bool {
	if ev.Key() != tcell.KeyTab && ev.Key() != tcell.KeyBacktab {
		return false
	}
	h.details = !h.details
	return true
}

// footer is the keys on the screen's foot.
func (h *epochHall) footer() [][2]string {
	turn := "details"
	if h.details {
		turn = vibe.Line("ui.epoch_hall.footer.02", "front page")
	}
	return [][2]string{{"↑↓", "scroll"}, {"PgUp PgDn", vibe.Line("ui.epoch_hall.footer.05", "page")}, {"Tab", turn}, {"Esc", "close"}}
}

func (h *epochHall) glyphTier() mapmodel.GlyphTier {
	if h.tier == nil {
		return mapmodel.TierUnicode
	}
	return h.tier()
}

func (h *epochHall) rows() int {
	if h.height == nil {
		return 40
	}
	return h.height()
}

// provider is the screen's text: the front page, or the details.
func (h *epochHall) provider(state game.GameState, screenW int) string {
	if h.details {
		return epochProvider(state, screenW)
	}
	return hallFrontText(state, screenW, h.rows(), h.glyphTier())
}

// ---- what the screen says ----

// hallStat is one headline figure.
type hallStat struct{ value, label string }

// hallBadgeCount is how many badges the front page shows.
const hallBadgeCount = 5

// hallView is the facts of the screen, shared by the front page and the
// card so that the two say the same.
type hallView struct {
	// name is the epoch's name, icon its mark and color its tview color name.
	name, icon, color string
	town, age, year   string
	stats             []hallStat
	// done are the ages behind the town, cur the age it is in, ahead the
	// ages ahead the player may see named, and unseen how many more there
	// are that it may not.
	done, ahead []string
	cur         string
	unseen      int
	totalAges   int
	badges      []game.BadgeView
	// feats are the run's feats as tview text with color tags.
	feats []string
}

// newHallView reads the screen's facts from a state.
func newHallView(state game.GameState) hallView {
	set := state.Ruleset()
	v := hallView{name: state.EpochName, icon: state.EpochIcon, color: "white"}
	if ep, ok := set.Era(state.EpochKey); ok {
		v.name, v.icon, v.color = ep.Name, ep.Icon, ep.Color
	}
	if v.name == "" {
		v.name = game.AgeName(state.Age)
	}
	v.town = mapmodel.TownNameOf(state.Seed)
	v.cur = game.AgeName(state.Age)
	v.age = v.cur
	if yr, ok := set.AgeYear(state.Age); ok {
		v.year = set.AgeYearLabel(state.Age)
		if v.year == "" {
			v.year = config.YearLabel(yr)
		}
	}

	buildings := 0
	for _, b := range state.Buildings {
		buildings += b.Count
	}
	v.stats = []hallStat{
		{textfmt.Int(state.Workers.TotalPop), vibe.Line("ui.epoch_hall.newHallView.01", "Population")},
		{textfmt.Int(buildings), vibe.Line("ui.epoch_hall.newHallView.02", "Buildings")},
		{vibe.Sprintf("ui.epoch_hall.newHallView.03", "%s of %s", textfmt.Int(state.Research.TotalResearched), textfmt.Int(len(set.Techs()))), vibe.Line("ui.epoch_hall.newHallView.04", "Techs")},
		{textfmt.Int(len(builtWonders(state))), vibe.Line("ui.epoch_hall.newHallView.05", "Wonders")},
		{textfmt.Int(state.Prestige.Level), vibe.Line("ui.epoch_hall.newHallView.06", "Prestige")},
	}

	// The ages: those the run has been through, the one it is in, and the
	// ones ahead, named only where the player may see them.
	sight := game.SightOf(&state)
	for _, a := range state.Stats.AgesReached {
		if a != state.Age {
			v.done = append(v.done, game.AgeName(a))
		}
	}
	ages := set.Ages()
	v.totalAges = len(ages)
	at := -1
	for i, a := range ages {
		if a.Key == state.Age {
			at = i
		}
	}
	if at >= 0 {
		for _, a := range ages[at+1:] {
			if sight.Age(a.Key) {
				v.ahead = append(v.ahead, a.Name)
			} else {
				v.unseen++
			}
		}
	}

	v.badges = rarestBadges(state, hallBadgeCount)
	v.feats = hallFeats(state)
	return v
}

// badgeRarityRank orders the rarities in words, common lowest. An unknown
// word ranks below them all.
func badgeRarityRank(word string) int {
	for r := config.BadgeMythic; r >= config.BadgeCommon; r-- {
		if r.Name() == word {
			return int(r)
		}
	}
	return 0
}

// rarestBadges are the n rarest badges the account has earned and the
// player may see: rarity first, then the points they are worth, then the
// one earned first. Integrity badges are not trophies.
func rarestBadges(state game.GameState, n int) []game.BadgeView {
	if state.AccountStats == nil {
		return nil
	}
	var got []game.BadgeView
	for _, b := range state.AccountStats.Badges {
		if b.Earned && !b.Hidden && !b.Integrity {
			got = append(got, b)
		}
	}
	sort.SliceStable(got, func(i, j int) bool {
		a, b := got[i], got[j]
		if ra, rb := badgeRarityRank(a.Rarity), badgeRarityRank(b.Rarity); ra != rb {
			return ra > rb
		}
		if a.Points != b.Points {
			return a.Points > b.Points
		}
		if !a.At.Equal(b.At) {
			if a.At.IsZero() != b.At.IsZero() {
				return b.At.IsZero()
			}
			return a.At.Before(b.At)
		}
		return a.Key < b.Key
	})
	if len(got) > n {
		got = got[:n]
	}
	return got
}

// hallFeats are the run's feats, from what the game already keeps: the
// catastrophes met, how each harbinger ended, the milestones and the
// fastest age. Each is a line of tview text; none is a doom not yet foretold.
func hallFeats(state game.GameState) []string {
	var feats []string
	if n := state.CatastrophesEndured + state.CatastrophesSuccumbed; n > 0 {
		feats = append(feats, vibe.Sprintf("ui.epoch_hall.hallFeats.01", "Catastrophes met: %d, endured %d, succumbed %d",
			n, state.CatastrophesEndured, state.CatastrophesSuccumbed))
	}
	for _, r := range state.HarbingerHistory {
		feats = append(feats, harbingerRecordText(state.Ruleset(), r))
	}
	if ms := state.Milestones; ms.CompletedCount > 0 {
		feats = append(feats, vibe.Sprintf("ui.epoch_hall.hallFeats.02", "Milestones reached: %d of %d", ms.CompletedCount, ms.TotalCount))
	}
	if age, ticks, ok := fastestAge(state); ok {
		feats = append(feats, vibe.Sprintf("ui.epoch_hall.hallFeats.03", "Fastest age: %s, in %s", age, game.DurationText(ticks, tickInterval(state))))
	}
	return feats
}

// fastestAge is the age of this run that took the fewest ticks to pass, of
// those passed (the current one is not done). ok is false before the first.
func fastestAge(state game.GameState) (name string, ticks int, ok bool) {
	log := state.Stats.AgeLog
	for i := 0; i+1 < len(log); i++ {
		if d := log[i+1].Tick - log[i].Tick; !ok || d < ticks {
			name, ticks, ok = game.AgeName(log[i].Age), d, true
		}
	}
	return name, ticks, ok
}

// ---- the front page ----

// hallMark picks a mark for the Unicode set or the plain one.
func hallMark(plain bool, uni, ascii string) string {
	if plain {
		return ascii
	}
	return uni
}

// hallFrontText is the front page for a screen screenW cells wide and rows
// tall.
func hallFrontText(state game.GameState, screenW, rows int, tier mapmodel.GlyphTier) string {
	w := overlayTextWidth(screenW)
	plain := tier == mapmodel.TierASCII
	v := newHallView(state)
	tag := theme.NameTag(v.color)
	var sb strings.Builder

	// The banner.
	sb.WriteString("\n")
	for _, l := range hallBanner(v.name, tag, w, rows, plain) {
		sb.WriteString(l + "\n")
	}
	sb.WriteString("\n")
	under := []string{lit(v.town), lit(v.age)}
	if v.year != "" {
		under = append(under, lit(v.year))
	}
	sep := " " + hallMark(plain, "·", "-") + " "
	line := strings.TrimSpace(v.icon + " " + strings.Join(under, sep))
	if plain {
		line = strings.Join(under, sep)
	}
	sb.WriteString(centreTagged(tag+"[::b]"+line+"[-:-:-]", visibleLen(line), w) + "\n\n")

	// The highlights.
	sb.WriteString(hallHeading(vibe.Line("ui.epoch_hall.hallFrontText.01", "Highlights")) + "\n\n")
	for _, l := range hallTiles(v.stats, w) {
		sb.WriteString(l + "\n")
	}
	sb.WriteString("\n")

	// The ages.
	sb.WriteString(hallHeading(vibe.Line("ui.epoch_hall.hallFrontText.02", "Ages")) + "\n\n")
	done, here, next := hallMark(plain, "✓", "+"), hallMark(plain, "◆", "*"), hallMark(plain, "○", "o")
	var toks []string
	for _, a := range v.done {
		toks = append(toks, theme.Paint(theme.RolePositive, done)+" "+lit(a))
	}
	toks = append(toks, tag+here+"[-] "+theme.Paint(theme.RoleBright, lit(v.cur)))
	for _, a := range v.ahead {
		toks = append(toks, theme.Paint(theme.RoleDim, next+" "+lit(a)))
	}
	if v.unseen > 0 {
		toks = append(toks, theme.Paint(theme.RoleDim, next+" "+vibe.Sprintf("ui.epoch_hall.hallFrontText.03", "%d more to come", v.unseen)))
	}
	for _, l := range hallFlow(toks, "   ", w) {
		sb.WriteString(l + "\n")
	}
	sb.WriteString("\n")

	// The badges.
	sb.WriteString(hallHeading(vibe.Line("ui.epoch_hall.hallFrontText.04", "Top badges")) + "\n\n")
	if len(v.badges) == 0 {
		sb.WriteString(theme.Paint(theme.RoleDim, vibe.Line("ui.epoch_hall.hallFrontText.05", "No badges earned yet.")) + "\n")
	}
	for _, b := range v.badges {
		sb.WriteString(badgeMark(b, tier) + " " + theme.Paint(theme.RoleBright, lit(b.Name)))
		if b.Tier != "" || b.Rarity != "" {
			sb.WriteString(theme.Paint(theme.RoleDim, " ("+lit(strings.Join(nonEmpty(b.Tier, b.Rarity), ", "))+")"))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// The feats.
	sb.WriteString(hallHeading(vibe.Line("ui.epoch_hall.hallFrontText.06", "Feats")) + "\n\n")
	if len(v.feats) == 0 {
		sb.WriteString(theme.Paint(theme.RoleDim, vibe.Line("ui.epoch_hall.hallFrontText.07", "No feats yet. They come with the ages.")) + "\n")
	}
	bullet := hallMark(plain, "·", "-")
	for _, f := range v.feats {
		sb.WriteString(" " + bullet + " " + f + "\n")
	}
	sb.WriteString("\n" + theme.Paint(theme.RoleDim, vibe.Line("ui.epoch_hall.hallFrontText.08", "Type epoch card to save a card you can paste into Discord.")) + "\n")
	return sb.String()
}

func nonEmpty(ss ...string) []string {
	var out []string
	for _, s := range ss {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func hallHeading(title string) string {
	return theme.Paint(theme.RoleAccent, "── "+title+" ──")
}

// centreTagged pads s (n cells wide on screen) to the middle of w cells.
func centreTagged(s string, n, w int) string {
	if n >= w {
		return s
	}
	return strings.Repeat(" ", (w-n)/2) + s
}

// hallFlow lays tokens out in rows at most w cells wide, sep between two on
// a row, never breaking inside a token.
func hallFlow(toks []string, sep string, w int) []string {
	var rows []string
	row, n := "", 0
	for _, t := range toks {
		tn := visibleLen(t)
		if n > 0 && n+len(sep)+tn > w {
			rows = append(rows, row)
			row, n = "", 0
		}
		if n > 0 {
			row += sep
			n += len(sep)
		}
		row += t
		n += tn
	}
	if n > 0 {
		rows = append(rows, row)
	}
	return rows
}

// hallTiles lays the headline figures out as columns: each figure over its
// label, as many to a row as w holds, then another row.
func hallTiles(stats []hallStat, w int) []string {
	cell := 0
	for _, s := range stats {
		cell = max(cell, runeLen(s.value), runeLen(s.label))
	}
	cell += 3
	per := max(w/cell, 1)
	var out []string
	for i := 0; i < len(stats); i += per {
		var vals, labels strings.Builder
		for _, s := range stats[i:min(i+per, len(stats))] {
			vals.WriteString(theme.Paint(theme.RoleHighlight, lit(s.value)) + strings.Repeat(" ", cell-runeLen(s.value)))
			labels.WriteString(theme.Paint(theme.RoleDim, s.label) + strings.Repeat(" ", cell-runeLen(s.label)))
		}
		out = append(out, strings.TrimRight(vals.String(), " "), strings.TrimRight(labels.String(), " "))
		if i+per < len(stats) {
			out = append(out, "")
		}
	}
	return out
}

// hallBanner is the epoch's name in iron, in the epoch's color, centred in w
// cells and at most the rows the screen can spare. A name that does not fit
// in iron is one line of spaced capitals.
func hallBanner(name, tag string, w, rows int, plain bool) []string {
	tag = strings.TrimSuffix(tag, "]") + "::b]"
	limit := 8
	if rows >= 36 {
		limit = 16
	}
	big := []ironMode{{sx: 2, sy: 2}, {sx: 1, sy: 2}}
	modes := []ironMode{{sx: 2, sy: 1}, {sx: 1, sy: 1}}
	if !plain {
		modes = append(modes, ironMode{sy: 1, half: true})
	}
	p := placeIron(name, append(big, modes...), w, limit)
	if !p.ok {
		text := strings.ToUpper(name)
		if s := spaced(text); runeLen(s) <= w {
			text = s
		}
		return []string{centreTagged(tag+lit(text)+"[-:-:-]", runeLen(text), w)}
	}
	g := newMGrid(p.w, p.h)
	drawIron(g, newMenuPalette(theme.Active()), p.rows, 0, 0, p.mode, func(int, int) float64 { return 0.7 }, false)
	if p.tail != "" {
		t := spaced(strings.ToUpper(p.tail))
		g.text(max((p.w-runeLen(t))/2, 0), p.h-1, t, theme.Color(theme.RoleText), true)
	}
	pad := strings.Repeat(" ", max((w-p.w)/2, 0))
	out := make([]string, 0, g.h)
	for y := 0; y < g.h; y++ {
		var sb strings.Builder
		sb.WriteString(pad)
		inRun := false
		for x := 0; x < g.w; x++ {
			r := g.c[y*g.w+x].r
			if r == ' ' {
				if inRun {
					sb.WriteString("[-:-:-]")
					inRun = false
				}
				sb.WriteRune(' ')
				continue
			}
			if !inRun {
				sb.WriteString(tag)
				inRun = true
			}
			if plain {
				r = '#'
			}
			sb.WriteRune(r)
		}
		if inRun {
			sb.WriteString("[-:-:-]")
		}
		out = append(out, strings.TrimRight(sb.String(), " "))
	}
	return out
}
