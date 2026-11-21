package ui

import (
	"slices"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
)

// lessons.go is the table of lessons and the rule that picks the one the band
// shows. A lesson is shown once: a main-view lesson on the band above the
// Command box, a panel's own lesson in that panel. It is learned when its
// done-test passes, and the account keeps what it learned.

const lessonMain = "main"

// lessonCtx is what a lesson's tests read: the town, the panel open ("" for
// none), the counts the dashboard keeps, and, for a lesson that compares, the
// value when the band first showed it.
type lessonCtx struct {
	st      game.GameState
	open    string
	typed   int // commands typed at the prompt this sitting
	listed  int // panels opened by a click on the list this sitting
	fresh   int // panels found and not opened
	learned map[string]bool
	seen    bool
	base    float64
	nx      *nextCtx
	step    nextStep // the next step the card names
}

// lesson is one row of the table.
type lesson struct {
	key        string
	name, text rowid.Entry // the name and the one sentence
	clause     rowid.Entry // what learns it
	view       string      // lessonMain, or the panel it is taught in
	when       func(*lessonCtx) bool
	done       func(*lessonCtx) bool
	// measure, set for a lesson learned when a number rises, replaces done:
	// the lesson is learned when it passes the value it had when first shown.
	measure func(*lessonCtx) float64
	// target is what Show me takes the player to, a building card or a panel.
	target func(*lessonCtx) (card, panel string)
}

// learnedBy reports whether the lesson's done-test passes.
func (l lesson) learnedBy(c *lessonCtx) bool {
	if l.measure != nil {
		return c.seen && l.measure(c) > c.base
	}
	return l.done(c)
}

func (c *lessonCtx) next() *nextCtx {
	if c.nx == nil {
		c.nx = &nextCtx{st: c.st, cards: buildCards(c.st, buyOne), set: c.st.Ruleset()}
	}
	return c.nx
}

func (c *lessonCtx) built() int {
	n := len(c.st.BuildQueue)
	for _, b := range c.st.Buildings {
		n += b.Count
	}
	return n
}

func (c *lessonCtx) atWork() float64 {
	n := 0
	for _, b := range c.st.Buildings {
		n += b.WorkersAssigned
	}
	return float64(n)
}

func (c *lessonCtx) jobFree() bool {
	for _, b := range c.st.Buildings {
		if b.Count*b.WorkerCapacity > b.WorkersAssigned {
			return true
		}
	}
	return false
}

func (c *lessonCtx) ageIndex() float64 {
	return float64(slices.Index(c.st.Ruleset().AgeKeys(), c.st.Age))
}

func (c *lessonCtx) panelOpen(key string) bool { return c.open == key }

// leftPanel is the done-test of a lesson learned on leaving its panel.
func leftPanel(key string) func(*lessonCtx) bool {
	return func(c *lessonCtx) bool { return c.seen && c.open != key }
}

func (c *lessonCtx) researched(kind config.TechKind) bool {
	for _, t := range c.st.Research.Techs {
		if t.Researched && t.Kind == kind {
			return true
		}
	}
	return false
}

// cardTarget is a target that is the cheapest card meeting ok.
func cardTarget(ok func(game.BuildingState, config.BuildingDef) bool) func(*lessonCtx) (string, string) {
	return func(c *lessonCtx) (string, string) {
		card, _ := c.next().pick(ok)
		return card.key, ""
	}
}

func panelTarget(name string) func(*lessonCtx) (string, string) {
	return func(*lessonCtx) (string, string) { return "", name }
}

var (
	txtLessonNew   = rowid.E("ui.lessons.txtNew.01", "NEW")
	txtLessonLater = rowid.E("ui.lessons.txtLater.01", " Later ")
	txtLessonsHead = rowid.E("ui.lessons.txtHead.01", "Lessons")
)

// lessonTable is the 21 lessons, main view first, then the panels' own. Their
// words are placeholders.
var lessonTable = []lesson{
	{key: "gather", view: lessonMain,
		name:   rowid.E("ui.lessons.gather.01", "Gather"),
		text:   rowid.E("ui.lessons.gather.02", "Gather by hand to start the town."),
		clause: rowid.E("ui.lessons.gather.03", "learned when you gather"),
		when:   func(c *lessonCtx) bool { return strings.HasPrefix(c.step.cmd, "gather ") },
		done: func(c *lessonCtx) bool {
			for _, n := range c.st.Stats.TotalGathered {
				if n > 0 {
					return true
				}
			}
			return false
		}},
	{key: "first_building", view: lessonMain,
		name:   rowid.E("ui.lessons.first_building.01", "First building"),
		text:   rowid.E("ui.lessons.first_building.02", "Build to make the town grow by itself."),
		clause: rowid.E("ui.lessons.first_building.03", "learned when a building is started"),
		when: func(c *lessonCtx) bool {
			card, ok := c.next().pick(func(game.BuildingState, config.BuildingDef) bool { return true })
			return ok && c.built() == 0 && card.canBuild
		},
		done:   func(c *lessonCtx) bool { return c.built() > 0 },
		target: cardTarget(func(game.BuildingState, config.BuildingDef) bool { return true })},
	// Stock has no ceiling in the game, so no resource sits at a cap and this
	// one is never shown.
	{key: "storage", view: lessonMain,
		name:    rowid.E("ui.lessons.storage.01", "Storage"),
		text:    rowid.E("ui.lessons.storage.02", "A full store stops what comes in."),
		clause:  rowid.E("ui.lessons.storage.03", "learned when the cap rises"),
		when:    func(*lessonCtx) bool { return false },
		measure: func(*lessonCtx) float64 { return 0 }},
	{key: "homes", view: lessonMain,
		name:   rowid.E("ui.lessons.homes.01", "Homes"),
		text:   rowid.E("ui.lessons.homes.02", "People need homes before the town can grow."),
		clause: rowid.E("ui.lessons.homes.03", "learned when housing rises"),
		when: func(c *lessonCtx) bool {
			return c.st.Workers.MaxPop > 0 && c.st.Workers.TotalPop >= c.st.Workers.MaxPop
		},
		measure: func(c *lessonCtx) float64 { return float64(c.st.Workers.MaxPop) },
		target: cardTarget(func(_ game.BuildingState, def config.BuildingDef) bool {
			return slices.ContainsFunc(def.Effects, func(e config.Effect) bool {
				return e.Type == "capacity" && e.Target == "population" && e.Value > 0
			})
		})},
	{key: "jobs", view: lessonMain,
		name:    rowid.E("ui.lessons.jobs.01", "Jobs"),
		text:    rowid.E("ui.lessons.jobs.02", "Idle people need a building with jobs."),
		clause:  rowid.E("ui.lessons.jobs.03", "learned when more people work"),
		when:    func(c *lessonCtx) bool { return c.st.Workers.TotalIdle > 0 && !c.jobFree() },
		measure: (*lessonCtx).atWork,
		target:  cardTarget(func(bs game.BuildingState, _ config.BuildingDef) bool { return bs.WorkerCapacity > 0 })},
	{key: "research", view: lessonMain,
		name:   rowid.E("ui.lessons.research.01", "Research"),
		text:   rowid.E("ui.lessons.research.02", "Research gives the town new things to build."),
		clause: rowid.E("ui.lessons.research.03", "learned when a tech starts"),
		when: func(c *lessonCtx) bool {
			r := c.st.Research
			return r.CurrentTech == "" && r.TotalResearched == 0 && slices.ContainsFunc(sortedKeysOf(r.Techs), func(k string) bool { return r.Techs[k].Available })
		},
		done:   func(c *lessonCtx) bool { return c.st.Research.CurrentTech != "" || c.st.Research.TotalResearched > 0 },
		target: panelTarget("research")},
	{key: "command_line", view: lessonMain,
		name:   rowid.E("ui.lessons.command_line.01", "Command line"),
		text:   rowid.E("ui.lessons.command_line.02", "Type a command at the bottom and press Enter."),
		clause: rowid.E("ui.lessons.command_line.03", "learned when you type a command"),
		when:   func(c *lessonCtx) bool { return c.learned["first_building"] },
		done:   func(c *lessonCtx) bool { return c.typed > 0 }},
	{key: "panels", view: lessonMain,
		name:   rowid.E("ui.lessons.panels.01", "Panels"),
		text:   rowid.E("ui.lessons.panels.02", "A name on the list opens a panel."),
		clause: rowid.E("ui.lessons.panels.03", "learned when you open one from the list"),
		when:   func(c *lessonCtx) bool { return c.fresh > 0 },
		done:   func(c *lessonCtx) bool { return c.listed > 0 }},
	{key: "festivals", view: lessonMain,
		name:   rowid.E("ui.lessons.festivals.01", "Festivals"),
		text:   rowid.E("ui.lessons.festivals.02", "A festival spends culture to speed the town."),
		clause: rowid.E("ui.lessons.festivals.03", "learned when you hold one"),
		when: func(c *lessonCtx) bool {
			f := c.st.Festival
			return f.Ready && f.Cost > 0 && f.Culture >= f.Cost
		},
		done:   func(c *lessonCtx) bool { return c.st.Festival.CooldownLeft > 0 },
		target: panelTarget("people")},

	{key: "charter", view: "charter",
		name:    rowid.E("ui.lessons.charter.01", "Charter"),
		text:    rowid.E("ui.lessons.charter.02", "The charter lists what the next age asks."),
		clause:  rowid.E("ui.lessons.charter.03", "learned when the town advances"),
		when:    func(c *lessonCtx) bool { return c.panelOpen("charter") },
		measure: (*lessonCtx).ageIndex},
	{key: "keystones", view: "tree",
		name:   rowid.E("ui.lessons.keystones.01", "Keystones"),
		text:   rowid.E("ui.lessons.keystones.02", "A keystone tech opens a wonder."),
		clause: rowid.E("ui.lessons.keystones.03", "learned when you research one"),
		when:   func(c *lessonCtx) bool { return c.panelOpen("tree") },
		done:   func(c *lessonCtx) bool { return c.researched(config.TechKeystone) }},
	{key: "plan", view: "plan",
		name:   rowid.E("ui.lessons.plan.01", "Plan"),
		text:   rowid.E("ui.lessons.plan.02", "Queue builds and research to run in order."),
		clause: rowid.E("ui.lessons.plan.03", "learned when the queue holds an item"),
		when:   func(c *lessonCtx) bool { return c.panelOpen("plan") },
		done:   func(c *lessonCtx) bool { return len(c.st.Plan) > 0 }},
	{key: "wonders", view: "wonders",
		name:   rowid.E("ui.lessons.wonders.01", "Wonders"),
		text:   rowid.E("ui.lessons.wonders.02", "A wonder is banked toward over time."),
		clause: rowid.E("ui.lessons.wonders.03", "learned when you start one"),
		when:   func(c *lessonCtx) bool { return c.panelOpen("wonders") },
		done: func(c *lessonCtx) bool {
			for _, b := range c.st.Buildings {
				if b.Category == "wonder" && (b.Count > 0 || len(b.WonderBank) > 0) {
					return true
				}
			}
			return false
		}},
	{key: "trade", view: "trade",
		name:   rowid.E("ui.lessons.trade.01", "Trade"),
		text:   rowid.E("ui.lessons.trade.02", "A route swaps one resource for another."),
		clause: rowid.E("ui.lessons.trade.03", "learned when a route runs"),
		when:   func(c *lessonCtx) bool { return c.panelOpen("trade") },
		done:   func(c *lessonCtx) bool { return len(c.st.Trade.ActiveRoutes) > 0 }},
	// Dealing is a trade made, a gift or alliance (the opinion moves off its
	// start of 0), or a standing other than neutral.
	{key: "factions", view: "factions",
		name:   rowid.E("ui.lessons.factions.01", "Factions"),
		text:   rowid.E("ui.lessons.factions.02", "Other peoples trade, give and quarrel."),
		clause: rowid.E("ui.lessons.factions.03", "learned when you deal with one"),
		when:   func(c *lessonCtx) bool { return c.panelOpen("factions") },
		done: func(c *lessonCtx) bool {
			for _, f := range c.st.Diplomacy.Factions {
				if f.Discovered && (f.TradeCount > 0 || f.Opinion != 0 || f.Status != "neutral") {
					return true
				}
			}
			return false
		}},
	{key: "expedition", view: "expedition",
		name:   rowid.E("ui.lessons.expedition.01", "Expedition"),
		text:   rowid.E("ui.lessons.expedition.02", "A party sent out comes home with a haul."),
		clause: rowid.E("ui.lessons.expedition.03", "learned when a party is out or back"),
		when:   func(c *lessonCtx) bool { return c.panelOpen("expedition") },
		done: func(c *lessonCtx) bool {
			m := c.st.Military
			return len(m.Active) > 0 || m.CompletedCount > 0
		}},
	{key: "army", view: "army",
		name:   rowid.E("ui.lessons.army.01", "Army"),
		text:   rowid.E("ui.lessons.army.02", "Soldiers stand at posts and blunt a raid."),
		clause: rowid.E("ui.lessons.army.03", "learned when you have a soldier"),
		when:   func(c *lessonCtx) bool { return c.panelOpen("army") },
		done:   func(c *lessonCtx) bool { return c.st.Military.SoldierCount+c.st.Military.SoldiersAway > 0 }},
	{key: "harbinger", view: "harbinger",
		name:    rowid.E("ui.lessons.harbinger.01", "Harbinger"),
		text:    rowid.E("ui.lessons.harbinger.02", "A harbinger warns of what is coming."),
		clause:  rowid.E("ui.lessons.harbinger.03", "learned when it has passed"),
		when:    func(c *lessonCtx) bool { return c.panelOpen("harbinger") },
		measure: func(c *lessonCtx) float64 { return float64(len(c.st.HarbingerHistory)) }},
	// An epoch choice is a catastrophe answered: its record carries the
	// outcome, endured or succumbed.
	{key: "epoch", view: "epoch",
		name:   rowid.E("ui.lessons.epoch.01", "Epoch"),
		text:   rowid.E("ui.lessons.epoch.02", "An epoch ends in a choice for the town."),
		clause: rowid.E("ui.lessons.epoch.03", "learned when you choose"),
		when:   func(c *lessonCtx) bool { return c.panelOpen("epoch") },
		done: func(c *lessonCtx) bool {
			return slices.ContainsFunc(c.st.EpochEventHistory, func(r game.EpochEventRecord) bool {
				return r.Outcome == game.CatastropheEndured || r.Outcome == game.CatastropheSuccumbed
			})
		}},
	{key: "badges", view: "badges",
		name:   rowid.E("ui.lessons.badges.01", "Badges"),
		text:   rowid.E("ui.lessons.badges.02", "Badges are kept by the account."),
		clause: rowid.E("ui.lessons.badges.03", "learned when you leave the panel"),
		when:   func(c *lessonCtx) bool { return c.panelOpen("badges") },
		done:   leftPanel("badges")},
	{key: "milestones", view: "milestones",
		name:   rowid.E("ui.lessons.milestones.01", "Milestones"),
		text:   rowid.E("ui.lessons.milestones.02", "Milestones mark what the town has done."),
		clause: rowid.E("ui.lessons.milestones.03", "learned when you leave the panel"),
		when:   func(c *lessonCtx) bool { return c.panelOpen("milestones") },
		done:   leftPanel("milestones")},
}

// lessonBook is a sitting's record of the lessons: what the account has
// learned, what the band has shown (with the value a lesson compares with),
// and what was put off.
type lessonBook struct {
	learned map[string]bool
	seen    map[string]float64
	later   map[string]bool
	now     string // the lesson the band shows, "" for none
}

// lessonStep takes one look at the town: lessons whose done-test passes are
// learned (and the keys returned to save), then the lesson for the view that
// shows is chosen. A lesson already shown stays until learned or put off.
func lessonStep(b *lessonBook, c *lessonCtx) (newly []string) {
	if b.seen == nil {
		b.seen, b.later = map[string]float64{}, map[string]bool{}
	}
	c.learned = b.learned
	for _, l := range lessonTable {
		if b.learned[l.key] {
			continue
		}
		c.base, c.seen = 0, false
		if v, ok := b.seen[l.key]; ok {
			c.base, c.seen = v, true
		}
		if l.learnedBy(c) {
			b.learned[l.key] = true
			newly = append(newly, l.key)
		}
	}
	b.now = ""
	for _, l := range lessonTable {
		if b.learned[l.key] || b.later[l.key] || l.view != lessonMain && l.view != c.open || l.view == lessonMain && c.open != "" {
			continue
		}
		_, shown := b.seen[l.key]
		if !shown {
			c.seen = false
			if !l.when(c) {
				continue
			}
			if l.measure != nil {
				b.seen[l.key] = l.measure(c)
			} else {
				b.seen[l.key] = 0
			}
		}
		b.now = l.key
		break
	}
	return newly
}

func lessonOf(key string) (lesson, bool) {
	i := slices.IndexFunc(lessonTable, func(l lesson) bool { return l.key == key })
	if i < 0 {
		return lesson{}, false
	}
	return lessonTable[i], true
}

// bandCells is the band's one row, w cells wide. Short of width the clause
// goes first, then the sentence is cut. showMe says whether the button is
// drawn.
func bandCells(l lesson, w int, showMe bool) string {
	plate := newPlate(txtLessonNew.String())
	later := clickRegion(regionLater, theme.TagFgBgAttr(theme.RoleText, theme.RoleChip, "b")+txtLessonLater.String()+theme.Reset)
	btns, btnW := later, runeLen(txtLessonLater.String())
	if showMe {
		btns = clickRegion(regionLesson, theme.TagFgBgAttr(theme.RoleText, theme.RoleChip, "b")+txtShowMe.String()+theme.Reset) + " " + later
		btnW += runeLen(txtShowMe.String()) + 1
	}
	name := l.name.String()
	head := runeLen(txtLessonNew.String()) + 2 + 1 + runeLen(name) // the plate, a pad cell, the name
	text, clause := l.text.String(), l.clause.String()
	room := w - head - btnW - 3 // two between the sentence and the buttons, one spare
	switch {
	case runeLen(text)+2+runeLen(clause)+2 <= room:
	case runeLen(text) <= room:
		clause = ""
	case room >= 4:
		text, clause = fitText(text, room), ""
	default:
		text, clause = "", ""
	}
	used := head + btnW
	left := plate + " " + theme.Tag(theme.RoleBright) + name + "[-]"
	if text != "" {
		left += "  " + bodyText(text)
		used += 2 + runeLen(text)
	}
	right := ""
	if clause != "" {
		right = dimText(clause) + "  "
		used += runeLen(clause) + 2
	}
	return left + strings.Repeat(" ", max(w-used, 1)) + right + btns
}

// bandRows is the band's height in a window h rows tall: none when no lesson
// shows or the window is short.
func (d *Dashboard) bandRows(h int) int {
	if d.lessons.now == "" || h < bandMinRows {
		return 0
	}
	return 1
}

// bandMinRows is the window height below which the band stays away: the
// Panels list and the cards need the row more.
const bandMinRows = 30

const (
	regionLesson = "lessonshow"
	regionLater  = "lessonlater"
)

// onScreen says whether Show me would take the player nowhere: the card is
// already listed in the Buildings box.
func (b *buildBox) onScreen(key string) bool {
	i := slices.Index(b.order, key)
	return !b.older && i >= 0 && i < b.shown
}

// stepLessons looks at the town on the refresh path (the UI goroutine, never
// a Bus handler), saves what was learned and sets the band.
func (d *Dashboard) stepLessons(st game.GameState) {
	acct := d.engine.Account()
	if d.lessons.learned == nil {
		d.lessons.learned = map[string]bool{}
		if acct != nil {
			for _, k := range acct.LessonsLearned() {
				d.lessons.learned[k] = true
			}
		}
	}
	c := &lessonCtx{st: st, typed: d.lessonTyped, listed: d.lessonListed, fresh: d.panels.newCount(), step: d.next}
	if d.overlayMgr != nil {
		c.open = d.overlayMgr.ActiveName()
	}
	if newly := lessonStep(&d.lessons, c); len(newly) > 0 && acct != nil {
		_ = acct.LearnLessons(newly...)
	}
	d.lessonCard, d.lessonPanel = "", ""
	if l, ok := lessonOf(d.lessons.now); ok && l.target != nil {
		d.lessonCard, d.lessonPanel = l.target(c)
	}
	d.bandBox.changed()
}

// bandText is the band's row for its width.
func (d *Dashboard) bandText(w int) string {
	l, ok := lessonOf(d.lessons.now)
	if !ok {
		return ""
	}
	show := d.lessonPanel != "" || d.lessonCard != "" && !d.economyTab.buildBox.onScreen(d.lessonCard)
	return bandCells(l, w, show)
}

// bandClick is a click on the band: Show me, or Later.
func (d *Dashboard) bandClick(name string) {
	switch name {
	case regionLesson:
		d.showMe(d.lessonCard, d.lessonPanel)
	case regionLater:
		if d.lessons.later == nil {
			d.lessons.later = map[string]bool{}
		}
		d.lessons.later[d.lessons.now] = true
		if d.lastState != nil {
			d.stepLessons(*d.lastState)
		}
	}
}

// lessonKeys is the keys of the lessons learned, in table order.
func (d *Dashboard) lessonKeys() []string {
	var out []string
	for _, l := range lessonTable {
		if d.lessons.learned[l.key] {
			out = append(out, l.key)
		}
	}
	return out
}

// lessonsHelp is Help's first section: the lessons learned in full, the rest
// as a count.
func lessonsHelp(learned []string, width int) string {
	var sb strings.Builder
	sb.WriteString("[gold]═══ " + txtLessonsHead.String() + " ═══[-]\n")
	left := len(lessonTable)
	for _, k := range learned {
		if l, ok := lessonOf(k); ok {
			sb.WriteString(hangingRow("  [cyan]"+l.name.String()+"[-] - ", l.text.String(), width))
			left--
		}
	}
	if left > 0 {
		sb.WriteString("[gray]" + rowid.Sprintf("ui.lessons.help.01", "%d more to come.", left) + "[-]\n\n")
	} else {
		sb.WriteString("\n")
	}
	return sb.String()
}
