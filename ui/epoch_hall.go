package ui

import (
	"sort"
	"strings"
	"time"

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
// is a picture of this page.
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
	// motion says whether things move (the motion setting) and now is the
	// clock. Either may be nil: still, and the wall clock.
	motion func() bool
	now    func() time.Time
	// state is the game as the picture reads it, and card makes the card (the
	// dashboard gives makeEpochCard with the real clipboard; tests give their
	// own). Either may be nil, and then c does nothing.
	state func() game.GameState
	card  func(game.GameState) (epochCardResult, error)
	// note is the line the last card left on the foot, until noteUntil.
	note      string
	noteUntil time.Time
}

// hallNoteFor is how long the card's confirmation stays on the foot.
const hallNoteFor = 5 * time.Second

func (h *epochHall) clock() time.Time {
	if h.now != nil {
		return h.now()
	}
	return time.Now()
}

// notice is the confirmation line for the foot, "" once it has gone.
func (h *epochHall) notice() string {
	if h.note == "" || !h.clock().Before(h.noteUntil) {
		return ""
	}
	return h.note
}

// makeCard is the c key: save the card, copy it, and say which on the foot.
func (h *epochHall) makeCard() {
	if h.state == nil || h.card == nil {
		return
	}
	res, err := h.card(h.state())
	switch {
	case err != nil:
		h.note = err.Error()
	case res.Copied:
		h.note = vibe.Line("ui.epoch_hall.makeCard.01", "Picture copied. Paste it into Discord.")
	default:
		h.note = vibe.Sprintf("ui.epoch_hall.makeCard.02", "Picture saved to %s.", res.Path)
	}
	h.noteUntil = h.clock().Add(hallNoteFor)
}

// open starts the screen on the front page.
func (h *epochHall) open() { h.details = false }

// back is Esc: the screen always closes.
func (h *epochHall) back() bool { return false }

// handleKey turns the page on Tab and reports whether it took the key; the
// rest (scrolling) is the text view's.
func (h *epochHall) handleKey(ev *tcell.EventKey) bool {
	had := h.note != ""
	h.note = ""
	if ev.Key() == tcell.KeyRune && (ev.Rune() == 'c' || ev.Rune() == 'C') {
		h.makeCard()
		return true
	}
	if ev.Key() != tcell.KeyTab && ev.Key() != tcell.KeyBacktab {
		return had
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
	mo := hallMotion{}
	if h.motion != nil && h.motion() {
		now := time.Now
		if h.now != nil {
			now = h.now
		}
		mo = hallMotion{on: true, t: float64(now().UnixMilli()) / 1000}
	}
	return hallFront(state, screenW, h.rows(), h.glyphTier(), mo)
}

// ---- what the screen says ----

// hallStat is one headline figure.
type hallStat struct {
	// value is the figure as the card says it, big the number the tile
	// draws large, and label what it counts. hasBar marks a figure that is
	// a share of a whole (frac of it done).
	value, label, big string
	hasBar            bool
	frac              float64
}

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
	// doneEp and curEp are the epoch (by order) of each age behind the town
	// and of the one it is in; order is the town's epoch; seed is the run's.
	doneEp       []int
	curEp, order int
	seed         int64
	cur          string
	unseen       int
	totalAges    int
	badges       []game.BadgeView
	// feats are the run's feats as tview text with color tags, featKinds the
	// kind of each (featCatastrophe and so on).
	feats     []string
	featKinds []int
}

// newHallView reads the screen's facts from a state.
func newHallView(state game.GameState) hallView {
	set := state.Ruleset()
	v := hallView{name: state.EpochName, icon: state.EpochIcon, color: "white"}
	if ep, ok := set.Era(state.EpochKey); ok {
		v.name, v.icon, v.color, v.order = ep.Name, ep.Icon, ep.Color, ep.Order
	}
	v.seed = state.Seed
	epochOf := func(age string) int {
		for _, ep := range set.Eras() {
			for _, a := range ep.Ages {
				if a == age {
					return ep.Order
				}
			}
		}
		return v.order
	}
	v.curEp = epochOf(state.Age)
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
	stat := func(n int, label string) hallStat {
		return hallStat{value: textfmt.Int(n), big: textfmt.Int(n), label: label}
	}
	techs := stat(state.Research.TotalResearched, vibe.Line("ui.epoch_hall.newHallView.04", "Techs"))
	techs.value = vibe.Sprintf("ui.epoch_hall.newHallView.03", "%s of %s", textfmt.Int(state.Research.TotalResearched), textfmt.Int(len(set.Techs())))
	techs.hasBar = true
	if total := len(set.Techs()); total > 0 {
		techs.frac = float64(state.Research.TotalResearched) / float64(total)
	}
	v.stats = []hallStat{
		stat(state.Workers.TotalPop, vibe.Line("ui.epoch_hall.newHallView.01", "Population")),
		stat(buildings, vibe.Line("ui.epoch_hall.newHallView.02", "Buildings")),
		techs,
		stat(len(builtWonders(state)), vibe.Line("ui.epoch_hall.newHallView.05", "Wonders")),
		stat(state.Prestige.Level, vibe.Line("ui.epoch_hall.newHallView.06", "Prestige")),
	}

	// The ages: those the run has been through, the one it is in, and the
	// ones ahead, named only where the player may see them.
	sight := game.SightOf(&state)
	for _, a := range state.Stats.AgesReached {
		if a != state.Age {
			v.done = append(v.done, game.AgeName(a))
			v.doneEp = append(v.doneEp, epochOf(a))
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
	for _, f := range hallFeats(state) {
		v.feats = append(v.feats, f.text)
		v.featKinds = append(v.featKinds, f.kind)
	}
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
func hallFeats(state game.GameState) []hallFeat {
	var feats []hallFeat
	if n := state.CatastrophesEndured + state.CatastrophesSuccumbed; n > 0 {
		feats = append(feats, hallFeat{featCatastrophe, vibe.Sprintf("ui.epoch_hall.hallFeats.01", "Catastrophes met: %d, endured %d, succumbed %d",
			n, state.CatastrophesEndured, state.CatastrophesSuccumbed)})
	}
	for _, r := range state.HarbingerHistory {
		feats = append(feats, hallFeat{featHarbinger, harbingerRecordText(state.Ruleset(), r)})
	}
	if ms := state.Milestones; ms.CompletedCount > 0 {
		feats = append(feats, hallFeat{featMilestone, vibe.Sprintf("ui.epoch_hall.hallFeats.02", "Milestones reached: %d of %d", ms.CompletedCount, ms.TotalCount)})
	}
	if age, ticks, ok := fastestAge(state); ok {
		feats = append(feats, hallFeat{featFastest, vibe.Sprintf("ui.epoch_hall.hallFeats.03", "Fastest age: %s, in %s", age, game.DurationText(ticks, tickInterval(state)))})
	}
	return feats
}

// hallFeat is one feat: its kind and its text (tview, with color tags).
type hallFeat struct {
	kind int
	text string
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
// tall, as a still frame.
func hallFrontText(state game.GameState, screenW, rows int, tier mapmodel.GlyphTier) string {
	return hallFront(state, screenW, rows, tier, hallMotion{})
}

// hallFront is the front page at a moment of the clock. Every block is
// centred as a unit on the width of the screen.
func hallFront(state game.GameState, screenW, rows int, tier mapmodel.GlyphTier, mo hallMotion) string {
	return hallFrontPage(state, screenW, rows, tier, mo, true)
}

// hallFrontPage is hallFront with or without the line that says which key
// copies the picture (the picture itself leaves it out).
func hallFrontPage(state game.GameState, screenW, rows int, tier mapmodel.GlyphTier, mo hallMotion, hint bool) string {
	v := newHallView(state)
	c := hallCompose(v, screenW, rows, tier, mo, 1, hint)
	// On a tall screen the page sits a little below the top, not hard
	// against it, with the stars all round.
	if spare := rows - 4 - c.end; spare > 3 {
		c = hallCompose(v, screenW, rows, tier, mo, 1+min(spare/3, 10), hint)
	}
	return c.render()
}

// hallCompose lays the page out with its banner starting at row top.
func hallCompose(v hallView, screenW, rows int, tier mapmodel.GlyphTier, mo hallMotion, top int, hint bool) *hallCtx {
	w := overlayTextWidth(screenW)
	c := newHallCtx(w, rows, tier, mo)
	stops := c.epochInks(v.order)

	// The banner and the lines under it.
	c.y = top
	c.banner(v.name, v.order, v.seed)
	under := []string{v.town, v.age}
	if v.year != "" {
		under = append(under, v.year)
	}
	line := strings.Join(under, " "+c.mark("·", "-")+" ")
	if !c.plain {
		line = strings.TrimSpace(v.icon + " " + line)
	}
	c.paragraph(line, c.ink(stops[2]))
	c.y += c.gap

	c.heading(vibe.Line("ui.epoch_hall.hallFrontText.01", "Highlights"))
	c.tiles(v.stats)
	c.y += c.gap - 1

	c.heading(vibe.Line("ui.epoch_hall.hallFrontText.02", "Ages"))
	c.track(v)
	c.y += c.gap

	c.heading(vibe.Line("ui.epoch_hall.hallFrontText.04", "Top badges"))
	c.badges(v.badges)
	c.y += c.gap - 1

	c.heading(vibe.Line("ui.epoch_hall.hallFrontText.06", "Feats"))
	c.feats(v)

	if hint {
		c.paragraph(vibe.Line("ui.epoch_hall.hallFrontText.08", "Press c to copy a picture for Discord."), c.role(theme.RoleDim))
	}
	c.end = c.y
	c.sky(v.seed)
	return c
}
