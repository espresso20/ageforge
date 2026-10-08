package main

import (
	"fmt"
	"sort"
	"strings"
)

// rows.go turns the text units of the source into the rows of the sheets.

// The four kinds of row.
const (
	kindVoice   = "voice"   // descriptions, flavor, story
	kindMessage = "message" // log lines, refusals, hints, help
	kindLabel   = "label"   // menu entries, headings, short interface words
	kindName    = "name"    // the name of a building, tech, badge and so on
)

// row is one line of a sheet, and the literal it stands for.
type row struct {
	id      string
	where   string
	kind    string
	age     int // position in the age order from 1; 0 for before the first age; -1 when not known
	max     int // longest the line may be, in characters; 0 for no limit
	current string
	sameAs  string

	area   *area
	u      *unit
	twins  []*unit // literals the code compares this text with, rewritten along with it
	unsure bool    // the reader could not confirm a player sees it
	seen   int     // how often a player meets it: lower is more often
	rules  ruleSet // which of the game's copy rules a rewrite must pass
	pool   int     // for a line drawn from a pool, how many lines share it
	eras   []string
}

// area is one sheet.
type area struct {
	file  string // the sheet's file name
	title string // what the README calls it
	about string // one sentence on what it holds
}

var (
	areaFlavorExp     = &area{"flavor-expeditions.csv", "Expedition stories", "The line of story the log adds when an expedition comes back, well or badly, and the word lists those lines draw on."}
	areaFlavorEnc     = &area{"flavor-encounters.csv", "Encounter stories", "The line of story the log adds when an expedition meets another civilization and nothing comes of it: a standoff with one you are at war with, or a court with no room for another boon."}
	areaFlavorRaid    = &area{"flavor-raids.csv", "Raid stories", "The line of story the log adds when a civilization at war with you raids."}
	areaFlavorHarbIn  = &area{"flavor-harbinger-warnings.csv", "Harbinger stories: arrival and warning", "What the log says when an age's harbinger arrives and gives the warning."}
	areaFlavorHarbDo  = &area{"flavor-harbinger-choices.csv", "Harbinger stories: appease, brace, invite", "What the log says after you appease, brace or invite."}
	areaFlavorHarbOut = &area{"flavor-harbinger-outcomes.csv", "Harbinger stories: how it ended", "What the log says when the warning comes true, comes to nothing, turns out false, or an invited doom arrives."}
	areaFlavorEnd     = &area{"flavor-run-endings.csv", "Run endings", "The line of story the log adds at every prestige."}
	areaBuildings     = &area{"buildings.csv", "Buildings", "Names, descriptions and flavor lines of the buildings."}
	areaTechs         = &area{"techs.csv", "Techs", "Names, card codes and descriptions of the techs, the lanes of the tree and what techs unlock."}
	areaBadges        = &area{"badges.csv", "Badges", "Names, descriptions and hints of the badges, their ladders and the titles they give."}
	areaMilestones    = &area{"milestones.csv", "Milestones", "Names, goals and quips of the milestones and their chains, and the titles they give."}
	areaAges          = &area{"wonders-and-ages.csv", "Wonders, ages and eras", "Names, descriptions and quips of the ages and eras, the wonders, the resources and the worker classes."}
	areaEvents        = &area{"events-and-catastrophes.csv", "Events and catastrophes", "Random events, era events, awakenings, the catastrophes, the Last Passage and the harbingers themselves."}
	areaWorld         = &area{"trade-and-civilizations.csv", "Trade, civilizations and prestige", "Trade routes, the other civilizations, boons and setbacks, the prestige shop and the short lines the log adds to routine news."}
	areaMap           = &area{"map.csv", "The map", "What the map says: the things you can inspect on it, its headers and hints, and the themes."}
	areaScreens       = &area{"screens-and-help.csv", "Screens and help", "The main menu, the panels, the help text, the prompts and the windows that pop up."}
	areaLog           = &area{"log-lines-and-refusals.csv", "Log lines and refusals", "What the game log says when something happens, and what it says when it will not do what you typed."}
	areaSaveNames     = &area{"save-names.csv", "Save names", "The words the game puts together when it suggests a name for a new save."}
	areaUnsure        = &area{"unsure.csv", "Not sure a player sees these", "Text the tool could not confirm ever reaches the screen. Rewrite a row here only once you know it does."}
)

// areas is every sheet, in the order the export writes and counts them.
var areas = []*area{
	areaScreens, areaLog, areaBuildings, areaTechs, areaAges, areaMilestones, areaBadges,
	areaEvents, areaWorld, areaMap,
	areaFlavorExp, areaFlavorEnc, areaFlavorRaid, areaFlavorHarbIn, areaFlavorHarbDo, areaFlavorHarbOut, areaFlavorEnd,
	areaSaveNames, areaUnsure,
}

// ruleSet names the groups of copy rules a rewrite is held to.
type ruleSet int

const (
	rulesUI      ruleSet = 1 << iota // the wording guards over ui, game and boon
	rulesConfig                      // the style guard over config's player text
	rulesDocs                        // the wiki's rules, for text the wiki quotes
	rulesNoMark                      // no square brackets and no % (written into a format the game owns)
	rulesSkel                        // a flavor catalog sentence
	rulesBank                        // a flavor catalog word-list entry
	rulesSaveNam                     // letters and single spaces only
)

// left is a unit the export leaves out, and why.
type left struct {
	u   *unit
	why string
}

// catalog is everything the reader found.
type catalog struct {
	m        *module
	w        *world
	fl       *flavorReader
	rows     []*row
	commands map[string]bool // the words a command can start with
	cmdWords map[string]bool // the subcommands and argument words that can follow
	rules    *copyRules      // the game's copy rules, read from its lint tests
	matched  map[string]bool // "package\x00text" for every text a package compares or looks up
	left     []left
	byID     map[string]*row
}

// read builds the catalog from the source under root.
func read(root string) (*catalog, error) {
	m, err := loadModule(root)
	if err != nil {
		return nil, err
	}
	rules, err := loadRules(root)
	if err != nil {
		return nil, err
	}
	c := &catalog{m: m, w: sharedWorld(), byID: map[string]*row{}, rules: rules}
	fl := newFlavorReader(m)
	c.fl = fl
	c.readCommands()
	// What each package matches on, so a word it also compares is not
	// mistaken for text.
	c.matched = map[string]bool{}
	perFile := map[*srcFile][]*unit{}
	for _, f := range m.files {
		perFile[f] = m.units(f)
		for _, u := range perFile[f] {
			switch u.near {
			case "mapkey", "compare", "case", "index":
				c.matched[f.dir+"\x00"+u.text] = true
			}
		}
	}
	for _, f := range m.files {
		units := perFile[f]
		cr := newCodeReader(c, f, units)
		for _, u := range units {
			var r *row
			var why string
			switch {
			case u.near == "mapkey" || u.near == "compare" || u.near == "case" || u.near == "index":
				why = whyMatched
			case f.mode == modeFlavor:
				r, why = fl.read(c, u)
			case f.mode == modeData && len(u.structs) > 0:
				r, why = readField(c, u)
			}
			if r == nil && why == "" {
				r, why = cr.read(u)
			}
			if r == nil {
				c.left = append(c.left, left{u, why})
				continue
			}
			r.u = u
			r.current = u.text
			if r.unsure {
				r.area = areaUnsure
			}
			c.rows = append(c.rows, r)
		}
	}
	fl.bankRows(c)
	c.link()
	if err := c.finish(); err != nil {
		return nil, err
	}
	return c, nil
}

// link attaches to each row the literals in the same file that the code
// matches its text with: a button's label and the "if chosen == label" that
// reads it, a line's "Wonder: " and the HasPrefix that looks for it. A
// rewrite changes both.
func (c *catalog) link() {
	byFileText := map[string]*row{}
	for _, r := range c.rows {
		k := r.u.f.rel + "\x00" + r.current
		if _, ok := byFileText[k]; !ok {
			byFileText[k] = r
		}
	}
	for _, l := range c.left {
		if (l.why != whyMatched && l.why != whyOp) || keyLike(l.u.text) {
			continue
		}
		if r, ok := byFileText[l.u.f.rel+"\x00"+l.u.text]; ok {
			r.twins = append(r.twins, l.u)
		}
	}
}

// finish orders the rows, checks the ids and fills same_as.
func (c *catalog) finish() error {
	order := map[*area]int{}
	for i, a := range areas {
		order[a] = i
	}
	sort.SliceStable(c.rows, func(i, j int) bool {
		a, b := c.rows[i], c.rows[j]
		if a.area != b.area {
			return order[a.area] < order[b.area]
		}
		return false // source order within a sheet
	})
	first := map[string]string{}
	for _, r := range c.rows {
		if r.area == nil {
			return fmt.Errorf("%s has no sheet", r.id)
		}
		if prev, dup := c.byID[r.id]; dup {
			return fmt.Errorf("two rows share the id %s (%s:%d and %s:%d)", r.id, prev.u.f.rel, prev.u.line, r.u.f.rel, r.u.line)
		}
		c.byID[r.id] = r
		if r.rules&rulesSkel != 0 {
			if c.rules.otherLines == nil {
				c.rules.otherLines = map[string]string{}
			}
			c.rules.otherLines[r.current] = r.id
		}
		if id, ok := first[r.current]; ok {
			r.sameAs = id
		} else {
			first[r.current] = r.id
		}
	}
	return nil
}

// ageText writes the age column: the age's place in the order and its name,
// so the column sorts to the early game.
func (c *catalog) ageText(r *row) string {
	switch {
	case r.age < 0:
		return ""
	case r.age == 0:
		return "00 Main menu"
	}
	return fmt.Sprintf("%02d %s", r.age, c.w.ageNames[r.age-1])
}

// idSafe makes a piece of an id out of free text: lower case, with anything
// that is not a letter, digit, dash or underscore turned into an underscore.
func idSafe(s string) string {
	var b strings.Builder
	gap := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r >= 'A' && r <= 'Z', r == '-':
			if gap && b.Len() > 0 {
				b.WriteByte('_')
			}
			gap = false
			b.WriteRune(r)
		default:
			gap = true
		}
	}
	return b.String()
}
