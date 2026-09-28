package main

// describe.go composes what a room says: the paragraph (authored sentences
// picked by the real state), the mechanical blocks under it (Here, Also here,
// Exits), and the other text screens (examine, survey, while-you-were-away).

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/flavor"
	"github.com/espresso20/ageforge/theme"
)

// ticksPerDay: one in-world day is 1200 ticks (40 minutes at 1x).
const ticksPerDay = 1200

// Hour of day for the state; tick 0 is seven in the morning.
func (c *City) Hour() float64 {
	return math.Mod(7+float64(c.St.Tick%ticksPerDay)*24/ticksPerDay, 24)
}

func (c *City) TimeWord() string {
	h := c.Hour()
	switch {
	case h < 5.5:
		return "night"
	case h < 7.5:
		return "dawn"
	case h < 11:
		return "morning"
	case h < 14:
		return "noon"
	case h < 18:
		return "afternoon"
	case h < 20:
		return "dusk"
	}
	return "night"
}

// Weather is cosmetic: seeded by the day, bent by active events.
func (c *City) Weather() string {
	for _, e := range c.St.ActiveEvents {
		k := strings.ToLower(e.Key + " " + e.Name)
		switch {
		case strings.Contains(k, "drought"):
			return "dry"
		case strings.Contains(k, "festival") || strings.Contains(k, "celebrat"):
			return "festival"
		}
	}
	day := itoa(c.St.Tick / ticksPerDay)
	v := hash(itoa(int(c.St.Seed)), day) % 10
	if c.Orbital {
		return [...]string{"quiet", "quiet", "quiet", "quiet", "quiet", "solar", "solar", "meteor", "quiet", "meteor"}[v]
	}
	return [...]string{"clear", "clear", "clear", "clear", "cloud", "cloud", "rain", "rain", "wind", "fog"}[v]
}

func (c *City) Scene(frame int) Scene {
	return Scene{Band: band(c.Epoch), Hour: c.Hour(), Weather: c.Weather(), Harbinger: c.St.Harbinger != nil, Frame: frame, Orbital: c.Orbital}
}

// pick chooses a sentence from list, stable for about ten minutes of play, so
// a second look says the same thing and a look after a while does not.
func (c *City) pick(list []string, salt string) string {
	if len(list) == 0 {
		return ""
	}
	return list[hash(salt, c.St.Age, itoa(c.St.Tick/300))%uint32(len(list))]
}

func fill(s, x string) string { return strings.Replace(s, "{x}", x, 1) }

func capFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Span is a run of text in one role.
type Span struct {
	Text string
	Role theme.Role
	Bold bool
}

// Para is a paragraph of spans; the renderer wraps it.
type Para []Span

func plain(s string, r theme.Role) Para { return Para{{Text: s, Role: r}} }

// recentBuilds maps building name → tick of its last completion in the log.
func (c *City) recentBuilds() map[string]int {
	out := map[string]int{}
	for _, l := range c.St.Log {
		if i := strings.Index(l.Message, "Build complete: "); i >= 0 {
			rest := l.Message[i+len("Build complete: "):]
			if j := strings.Index(rest, " (count"); j >= 0 {
				rest = rest[:j]
			}
			out[rest] = l.Tick
		}
	}
	return out
}

// resourcePlace maps a resource to the place that makes it.
func (c *City) resourcePlace(res string) (PlaceKey, bool) {
	for _, d := range c.defs {
		if d.OutputResource == res {
			if k, ok := c.lineHome[d.LineageKey]; ok {
				if _, here := c.Places[k]; here {
					return k, true
				}
			}
		}
	}
	return "", false
}

func (c *City) atWar() []string {
	var out []string
	for _, f := range c.sortedFactions() {
		if f.AtWar {
			out = append(out, f.Name)
		}
	}
	return out
}

type factionRow struct {
	Key                       string
	Name, Status, Personality string
	Opinion                   int
	AtWar                     bool
	Deals                     int
}

func (c *City) sortedFactions() []factionRow {
	var out []factionRow
	for k, f := range c.St.Diplomacy.Factions {
		if !f.Discovered {
			continue
		}
		out = append(out, factionRow{Key: k, Name: f.Name, Status: f.Status, Personality: f.Personality, Opinion: f.Opinion, AtWar: f.AtWar, Deals: len(f.Deals)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

// Describe is the room paragraph.
func (c *City) Describe(p *Place, since *City) Para {
	var s []string
	add := func(line string) {
		if line != "" {
			s = append(s, line)
		}
	}
	salt := string(p.Key)
	if p.Path {
		add(capFirst(p.Name) + ".")
		add(c.pick(pathLines, salt))
	} else {
		add(establish[p.Key][c.Epoch])
		if c.AgeIdx == 0 && p.Key == Square {
			s[0] = "A clearing with a fire in it, which is most of what the settlement is so far."
		}
	}

	total := p.Total()
	tier := "busy"
	switch {
	case total <= 3:
		tier = "sparse"
	case total >= 25:
		tier = "crowded"
	}
	if !p.Path && p.Key != Square && p.Key != Gate && p.Key != Wonders {
		add(c.pick(density[tier], salt+"d"))
	}

	// Conditions, most important first; at most three.
	var cond []string
	recent := c.recentBuilds()
	switch p.Key {
	case Square:
		if h := c.St.Harbinger; h != nil {
			cond = append(cond, capFirst(fill(c.pick(lineHarbinger, salt+"h"), h.Name)))
			if len(h.Lines) > 0 {
				cond = append(cond, h.Lines[hash(h.Key, itoa(c.St.Tick/300))%uint32(len(h.Lines))])
			}
		}
		if c.St.Workers.TotalIdle > 0 {
			cond = append(cond, c.pick(lineIdle, salt+"i"))
		}
	case Gate:
		if war := c.atWar(); len(war) > 0 {
			cond = append(cond, c.pick(lineWar, salt+"w"))
		} else if len(c.sortedFactions()) > 0 {
			cond = append(cond, c.pick(linePeace, salt+"p"))
		}
		if sc := c.St.Military.ActiveScout; sc != nil {
			cond = append(cond, fill(c.pick(lineExpedition, salt+"e"), sc.Name))
		} else if sc := c.St.Military.ActiveMilitary; sc != nil {
			cond = append(cond, fill(c.pick(lineExpedition, salt+"e"), sc.Name))
		}
		if l := c.lastExpeditionLine(); l != "" {
			cond = append(cond, l)
		}
		if c.St.AgeReady {
			cond = append(cond, c.pick(lineReady, salt+"r"))
		}
	case Wonders:
		if w := c.St.CurrentAgeWonderName; w != "" {
			for _, q := range c.St.BuildQueue {
				if q.Name == w {
					cond = append(cond, fill(c.pick(lineWonderBuilding, salt+"w"), w))
				}
			}
		}
	case Homes:
		switch {
		case c.St.Morale >= 0.8:
			cond = append(cond, c.pick(lineMoraleHigh, salt+"m"))
		case c.St.Morale <= 0.3:
			cond = append(cond, c.pick(lineMoraleLow, salt+"m"))
		}
	}
	if (p.Key == Harbour) || (p.Key == Market && c.Places[Harbour] == nil) {
		for _, r := range c.St.Trade.ActiveRoutes {
			if r.Disrupted {
				cond = append(cond, fill(c.pick(lineRouteBlocked, salt+"rb"), r.Name))
			} else {
				cond = append(cond, fill(c.pick(lineRouteBusy, salt+"r"), r.Name))
			}
			break
		}
	}
	for _, e := range c.St.ActiveEvents {
		for _, ef := range e.Effects {
			if k, ok := c.resourcePlace(ef.Target); ok && k == p.Key {
				cond = append(cond, fill(c.pick(lineEvent, salt+"ev"), e.Name))
				break
			}
		}
	}
	for _, h := range p.Holdings {
		if t, ok := recent[h.Name]; ok && c.St.Tick-t < 900 && h.Count > 0 {
			cond = append(cond, fill(c.pick(lineNew, salt+"n"), strings.ToLower(h.Name)))
			break
		}
	}
	for _, h := range p.Holdings {
		if h.Ruins > 0 {
			cond = append(cond, fill(c.pick(lineRuins, salt+"ru"), strings.ToLower(h.Name)))
			break
		}
	}
	if st := p.Staffing(); st >= 0 && tier != "sparse" {
		if st < 0.6 {
			cond = append(cond, c.pick(lineUnderstaffed, salt+"u"))
		} else if st >= 0.99 && hash(salt, c.St.Age)%2 == 0 {
			cond = append(cond, c.pick(lineStaffed, salt+"s"))
		}
	}
	var legacy []Holding
	for _, h := range p.Holdings {
		if h.Legacy && h.Count > 0 {
			legacy = append(legacy, h)
		}
	}
	if len(legacy) > 0 && len(legacy) < len(p.Holdings) {
		h := legacy[hash(salt, c.St.Age, "leg")%uint32(len(legacy))]
		cond = append(cond, fill(c.pick(lineLegacy, salt+"l"), strings.ToLower(h.Name)))
	}
	if len(cond) > 3 {
		cond = cond[:3]
	}
	s = append(s, cond...)

	// Sky, unless you are indoors (the orbital rooms are all indoors).
	b := band(c.Epoch)
	reg := 0
	if b >= 1 {
		reg = 1
	}
	if b >= 2 {
		reg = 2
	}
	if sk := sky[c.TimeWord()][reg]; sk != "" {
		add(sk)
	}
	if w := weather[c.Weather()][reg]; w != "" && c.Weather() != "clear" {
		add(w)
	}
	para := Para{{Text: strings.Join(s, " "), Role: theme.RoleText}}
	return para
}

// lastExpeditionLine narrates the most recent returned expedition through the
// game's own flavor generator, so the gate speaks in the house voice.
func (c *City) lastExpeditionLine() string {
	for i := len(c.St.Log) - 1; i >= 0; i-- {
		l := c.St.Log[i]
		if c.St.Tick-l.Tick > 1200 {
			break
		}
		m := l.Message
		var mom flavor.Moment
		switch {
		case strings.Contains(m, "Expedition") && strings.Contains(m, "succeeded"):
			mom = flavor.ExpeditionSuccess
		case strings.Contains(m, "Expedition") && (strings.Contains(m, "failed") || strings.Contains(m, "partial")):
			mom = flavor.ExpeditionFailure
		default:
			continue
		}
		return flavor.Line(flavor.Request{Moment: mom, Age: c.St.Age}, newRand(int64(hash(m, itoa(l.Tick)))))
	}
	return ""
}

// HereLine is the mechanical list of what stands in the room.
func (c *City) HereLine(p *Place) Para {
	if len(p.Holdings) == 0 {
		return Para{{Text: "Here: ", Role: theme.RoleLabel, Bold: true}, {Text: "nothing yet.", Role: theme.RoleDim}}
	}
	out := Para{{Text: "Here: ", Role: theme.RoleLabel, Bold: true}}
	for i, h := range p.Holdings {
		if i > 0 {
			out = append(out, Span{Text: " · ", Role: theme.RoleDim})
		}
		if h.Count > 0 {
			role := theme.RoleBright
			if h.Legacy {
				role = theme.RoleText
			}
			out = append(out, Span{Text: h.Name, Role: role})
			if h.Category != "wonder" {
				out = append(out, Span{Text: " ×" + itoa(h.Count), Role: theme.RoleHighlight})
			}
			if h.Legacy {
				out = append(out, Span{Text: " (old)", Role: theme.RoleDim})
			}
		}
		if h.Ruins > 0 {
			if h.Count == 0 {
				out = append(out, Span{Text: h.Name, Role: theme.RoleText})
			}
			out = append(out, Span{Text: fmt.Sprintf(" %d ruined", h.Ruins), Role: theme.RoleNegative})
		}
	}
	w, cp := 0, 0
	for _, h := range p.Holdings {
		w += h.Workers
		cp += h.Capacity
	}
	if cp > 0 {
		role := theme.RolePositive
		if float64(w) < 0.6*float64(cp) {
			role = theme.RoleWarning
		}
		out = append(out, Span{Text: "   Hands: ", Role: theme.RoleLabel, Bold: true}, Span{Text: fmt.Sprintf("%d/%d", w, cp), Role: role})
	}
	return out
}

// bar is a five-cell fill gauge; its shape reads without colour.
func bar(f float64) string {
	n := int(math.Round(f * 5))
	n = clamp(n, 0, 5)
	return strings.Repeat("▰", n) + strings.Repeat("▱", 5-n)
}

func rate(r float64) string {
	sign := "+"
	if r < 0 {
		sign = "-"
		r = -r
	}
	switch {
	case r >= 1e4:
		return sign + compact(r) + "/t"
	case r >= 100:
		return fmt.Sprintf("%s%.0f/t", sign, r)
	}
	return fmt.Sprintf("%s%.1f/t", sign, r)
}

// Ledger is the room's working numbers: what it makes and how full that is,
// or, for the places that make nothing, what they are for. It is the reason
// to walk somewhere rather than open a panel: the numbers sit where the work is.
func (c *City) Ledger(p *Place) []Para {
	var out []Para
	label := func(s string) Span { return Span{Text: s + ": ", Role: theme.RoleLabel, Bold: true} }
	resLine := func(key string) []Span {
		r, ok := c.St.Resources[key]
		if !ok {
			return nil
		}
		role := theme.RolePositive
		if r.Rate < 0 {
			role = theme.RoleNegative
		}
		sp := []Span{{Text: r.Name + " ", Role: theme.RoleBright}, {Text: rate(r.Rate), Role: role}}
		if r.Storage > 0 {
			f := r.Amount / r.Storage
			brole := theme.RoleDim
			if f >= 0.99 {
				brole = theme.RoleWarning
			}
			sp = append(sp, Span{Text: " " + bar(f), Role: brole})
			if f >= 0.99 {
				sp = append(sp, Span{Text: " full", Role: theme.RoleWarning})
			}
		}
		return sp
	}
	// What this place makes.
	var made []string
	seen := map[string]bool{}
	for _, h := range p.Holdings {
		if h.Count == 0 {
			continue
		}
		if res := c.defs[h.Key].OutputResource; res != "" && !seen[res] {
			seen[res] = true
			made = append(made, res)
		}
	}
	if len(made) > 0 && p.Key != Stores {
		row := Para{label("Makes")}
		for i, res := range made {
			if i == 3 {
				break
			}
			if i > 0 {
				row = append(row, Span{Text: "   ", Role: theme.RoleDim})
			}
			row = append(row, resLine(res)...)
		}
		out = append(out, row)
	}
	switch p.Key {
	case Square:
		w := c.St.Workers
		row := Para{label("Folk"), {Text: fmt.Sprintf("%d", w.TotalPop), Role: theme.RoleHighlight}, {Text: " (room for " + compact(float64(w.MaxPop)) + ")", Role: theme.RoleDim}}
		if w.TotalIdle > 0 {
			row = append(row, Span{Text: fmt.Sprintf("   %d idle", w.TotalIdle), Role: theme.RoleWarning})
		}
		row = append(row, Span{Text: "   morale ", Role: theme.RoleDim}, Span{Text: bar(c.St.Morale), Role: theme.RoleHighlight})
		out = append(out, row)
	case Stores:
		type fill struct {
			name string
			f    float64
		}
		var fs []fill
		for _, r := range c.St.Resources {
			if r.Unlocked && r.Storage > 0 && r.Amount > 0 {
				fs = append(fs, fill{r.Name, r.Amount / r.Storage})
			}
		}
		sort.Slice(fs, func(i, j int) bool {
			if fs[i].f != fs[j].f {
				return fs[i].f > fs[j].f
			}
			return fs[i].name < fs[j].name
		})
		row := Para{label("Shelves")}
		for i, f := range fs {
			if i == 5 {
				break
			}
			if i > 0 {
				row = append(row, Span{Text: "  ", Role: theme.RoleDim})
			}
			role := theme.RoleDim
			if f.f >= 0.99 {
				role = theme.RoleWarning
			}
			row = append(row, Span{Text: f.name + " ", Role: theme.RoleText}, Span{Text: bar(f.f), Role: role})
		}
		out = append(out, row)
	case Academy:
		if r := c.St.Research; r.CurrentTech != "" {
			done := 0.0
			if r.TotalTicks > 0 {
				done = 1 - float64(r.TicksLeft)/float64(r.TotalTicks)
			}
			out = append(out, Para{label("Studying"), {Text: r.CurrentTechName + " ", Role: theme.RoleBright}, {Text: bar(done), Role: theme.RoleHighlight}, {Text: fmt.Sprintf(" %d ticks left", r.TicksLeft), Role: theme.RoleDim}})
		}
	case Wonders:
		for _, q := range c.St.BuildQueue {
			if q.Name == c.St.CurrentAgeWonderName && q.TotalTicks > 0 {
				out = append(out, Para{label("Raising"), {Text: q.Name + " ", Role: theme.RoleBright}, {Text: bar(1 - float64(q.TicksLeft)/float64(q.TotalTicks)), Role: theme.RoleHighlight}})
			}
		}
		if len(out) == 0 && c.St.CurrentAgeWonderName != "" {
			out = append(out, Para{label("Planned"), {Text: c.St.CurrentAgeWonderName, Role: theme.RoleBright}, {Text: " (this age's wonder, not yet begun)", Role: theme.RoleDim}})
		}
	case Barracks:
		m := c.St.Military
		out = append(out, Para{label("Garrison"), {Text: compact(float64(m.SoldierCount)) + " soldiers", Role: theme.RoleHighlight}, {Text: fmt.Sprintf("   defence %.0f", m.DefenseRating), Role: theme.RoleDim}})
	case Gate:
		if sc := c.St.Military.ActiveScout; sc != nil {
			out = append(out, Para{label("Away"), {Text: sc.Name, Role: theme.RoleBright}, {Text: fmt.Sprintf(", back in %d ticks", sc.TicksLeft), Role: theme.RoleDim}})
		}
		if n := len(c.St.Trade.ActiveRoutes); n > 0 && c.Places[Harbour] == nil {
			out = append(out, Para{label("Caravans"), {Text: fmt.Sprintf("%d route(s) on the road", n), Role: theme.RoleText}})
		}
	case Harbour:
		for _, r := range c.St.Trade.ActiveRoutes {
			role := theme.RolePositive
			state := fmt.Sprintf("%d sailings", r.CyclesDone)
			if r.Disrupted {
				role, state = theme.RoleNegative, "blockaded"
			}
			out = append(out, Para{label("Route"), {Text: r.Name + " ", Role: theme.RoleBright}, {Text: state, Role: role}})
		}
	}
	return out
}

// AlsoLine lists who, rather than what, is here.
func (c *City) AlsoLine(p *Place) Para {
	var who []string
	switch p.Key {
	case Square:
		if h := c.St.Harbinger; h != nil {
			who = append(who, h.Name)
		}
		if n := c.St.Workers.TotalIdle; n > 0 {
			who = append(who, fmt.Sprintf("%d idle workers", n))
		}
	case Gate:
		for _, f := range c.sortedFactions() {
			switch {
			case f.AtWar:
				who = append(who, "raiders of "+f.Name)
			case f.Opinion >= 50:
				who = append(who, "an envoy of "+f.Name)
			}
		}
		if sc := c.St.Military.ActiveScout; sc != nil {
			who = append(who, fmt.Sprintf("tracks of the %s (away, %d ticks)", sc.Name, sc.TicksLeft))
		}
	case Harbour, Market:
		if p.Key == Harbour || c.Places[Harbour] == nil {
			for _, r := range c.St.Trade.ActiveRoutes {
				who = append(who, "cargo for the "+r.Name)
			}
		}
	case Barracks:
		if n := c.St.Military.SoldierCount; n > 0 {
			who = append(who, fmt.Sprintf("%s soldiers", compact(float64(n))))
		}
	}
	if len(who) == 0 {
		return nil
	}
	return Para{{Text: "Also here: ", Role: theme.RoleLabel, Bold: true}, {Text: strings.Join(who, ", ") + ".", Role: theme.RoleText}}
}

// ExitsLine lists the roads out.
func (c *City) ExitsLine(p *Place) Para {
	out := Para{{Text: "Exits: ", Role: theme.RoleLabel, Bold: true}}
	first := true
	for d := North; d <= West; d++ {
		k, ok := p.Exits[d]
		if !ok {
			continue
		}
		if !first {
			out = append(out, Span{Text: " · ", Role: theme.RoleDim})
		}
		first = false
		to := c.Places[k]
		role := theme.RoleText
		if to.Path {
			role = theme.RoleDim
		}
		out = append(out, Span{Text: c.DirWord(d), Role: theme.RoleAccent, Bold: true}, Span{Text: " " + to.Name, Role: role})
	}
	if p.Key == Gate {
		if len(c.sortedFactions()) > 0 {
			out = append(out, Span{Text: " · ", Role: theme.RoleDim}, Span{Text: "out", Role: theme.RoleAccent, Bold: true}, Span{Text: " the wide world", Role: theme.RoleText})
		}
	}
	return out
}

// Examine is a building's card.
func (c *City) Examine(p *Place, key string) []Para {
	for _, h := range p.Holdings {
		if h.Key != key {
			continue
		}
		var out []Para
		title := Para{{Text: h.Name, Role: theme.RoleAccent, Bold: true}}
		if h.Category != "wonder" {
			title = append(title, Span{Text: fmt.Sprintf("  ×%d", h.Count), Role: theme.RoleHighlight})
		}
		title = append(title, Span{Text: "  in " + p.Name, Role: theme.RoleDim})
		out = append(out, title)
		if h.Flavor != "" {
			out = append(out, plain(h.Flavor, theme.RoleText))
		}
		if h.Desc != "" {
			out = append(out, plain(h.Desc, theme.RoleDim))
		}
		var facts []string
		if h.Capacity > 0 {
			facts = append(facts, fmt.Sprintf("hands %d/%d", h.Workers, h.Capacity))
		}
		if h.Legacy {
			facts = append(facts, "old: still works, cannot be built any more")
		}
		if h.Ruins > 0 {
			facts = append(facts, fmt.Sprintf("%d in ruins (half output, no workers)", h.Ruins))
		}
		facts = append(facts, "key "+h.Key)
		out = append(out, plain(strings.Join(facts, " · "), theme.RoleLabel))
		return out
	}
	return nil
}

// Survey is the whole settlement on one screen, place by place.
func (c *City) Survey() []Para {
	out := []Para{{{Text: "You climb somewhere high and look out over " + c.Places[Square].Name + " and everything around it.", Role: theme.RoleText}}}
	for _, k := range c.Order {
		p := c.Places[k]
		row := Para{{Text: glyphFor(k) + " ", Role: theme.RoleAccent}, {Text: padRight(capFirst(p.Name), 24), Role: theme.RoleBright}}
		switch {
		case p.Path:
			row = append(row, Span{Text: "open ground", Role: theme.RoleDim})
		case len(p.Holdings) == 0:
			row = append(row, Span{Text: "—", Role: theme.RoleDim})
		default:
			h := p.Holdings[0]
			row = append(row, Span{Text: fmt.Sprintf("%4d built", p.Total()), Role: theme.RoleHighlight}, Span{Text: "  newest: " + h.Name, Role: theme.RoleText})
			if st := p.Staffing(); st >= 0 && st < 0.6 {
				row = append(row, Span{Text: "  short-handed", Role: theme.RoleWarning})
			}
			if p.Ruins() > 0 {
				row = append(row, Span{Text: fmt.Sprintf("  %d ruined", p.Ruins()), Role: theme.RoleNegative})
			}
		}
		out = append(out, row)
	}
	return out
}

// World is the room beyond the gate: the civilizations you know of.
func (c *City) World() []Para {
	fs := c.sortedFactions()
	out := []Para{{{Text: "The wide world", Role: theme.RoleAccent, Bold: true}, {Text: "  beyond " + c.Places[Gate].Name, Role: theme.RoleDim}}}
	if len(fs) == 0 {
		return append(out, plain("The road runs out into country nobody here has mapped. Send someone.", theme.RoleText))
	}
	out = append(out, plain("The road forks, and every fork has somebody at the end of it.", theme.RoleText))
	for i, f := range fs {
		bearing := c.DirWord(Dir(hash(f.Key) % 4))
		days := 2 + int(hash(f.Key, "d")%9)
		role, mood := theme.RoleText, f.Status
		switch {
		case f.AtWar:
			role, mood = theme.RoleNegative, "at war"
		case f.Opinion >= 60:
			role = theme.RolePositive
		}
		out = append(out, Para{
			{Text: fmt.Sprintf("%2d  ", i+1), Role: theme.RoleDim},
			{Text: padRight(f.Name, 26), Role: theme.RoleBright},
			{Text: padRight(fmt.Sprintf("%d %s %s", days, [4]string{"days", "days", "hours", "jumps"}[band(c.Epoch)], bearing), 22), Role: theme.RoleDim},
			{Text: padRight(mood, 10), Role: role},
			{Text: fmt.Sprintf("opinion %+d  %s", f.Opinion, f.Personality), Role: theme.RoleLabel},
		})
	}
	return out
}

func compact(v float64) string {
	switch {
	case v >= 1e9:
		return fmt.Sprintf("%.1fB", v/1e9)
	case v >= 1e6:
		return fmt.Sprintf("%.1fM", v/1e6)
	case v >= 1e4:
		return fmt.Sprintf("%.0fk", v/1e3)
	}
	return fmt.Sprintf("%.0f", v)
}

func padRight(s string, n int) string {
	l := len([]rune(s))
	if l >= n {
		return s + " "
	}
	return s + strings.Repeat(" ", n-l)
}

// ---- Idle return -----------------------------------------------------------

// Change is one thing that differs between two looks at the city.
type Change struct {
	Place PlaceKey
	Text  Para
}

// Diff lists what changed from prev to c, most notable first.
func (c *City) Diff(prev *City) (away string, changes []Change) {
	ms := (c.St.Tick - prev.St.Tick) * max(c.St.TickIntervalMs, 1)
	d := durationWords(ms)
	if prev.St.Age != c.St.Age {
		changes = append(changes, Change{Square, Para{{Text: prev.St.AgeName + " gave way to the " + c.St.AgeName + ".", Role: theme.RoleAccent, Bold: true}}})
	}
	for _, k := range c.Order {
		p := c.Places[k]
		old := prev.Places[k]
		if old == nil || (old.Path && !p.Path) {
			if !p.Path {
				changes = append(changes, Change{k, Para{{Text: "New: ", Role: theme.RolePositive, Bold: true}, {Text: capFirst(p.Name), Role: theme.RoleBright}, {Text: fmt.Sprintf(" (%d built)", p.Total()), Role: theme.RoleDim}}})
			}
			continue
		}
		if old.Name != p.Name {
			changes = append(changes, Change{k, Para{{Text: capFirst(old.Name), Role: theme.RoleText}, {Text: " is ", Role: theme.RoleDim}, {Text: p.Name, Role: theme.RoleBright}, {Text: " now.", Role: theme.RoleDim}}})
		}
		was := map[string]int{}
		wasR := map[string]int{}
		for _, h := range old.Holdings {
			was[h.Key] = h.Count
			wasR[h.Key] = h.Ruins
		}
		var parts []string
		for _, h := range p.Holdings {
			if dn := h.Count - was[h.Key]; dn > 0 {
				parts = append(parts, fmt.Sprintf("+%d %s", dn, h.Name))
			}
		}
		if len(parts) > 0 {
			if len(parts) > 3 {
				parts = append(parts[:3], fmt.Sprintf("%d more", len(parts)-3))
			}
			changes = append(changes, Change{k, Para{{Text: capFirst(p.Name) + ": ", Role: theme.RoleLabel}, {Text: strings.Join(parts, ", "), Role: theme.RolePositive}}})
		}
		for _, h := range p.Holdings {
			if dr := h.Ruins - wasR[h.Key]; dr > 0 {
				changes = append(changes, Change{k, Para{{Text: capFirst(p.Name) + ": ", Role: theme.RoleLabel}, {Text: fmt.Sprintf("%d %s lost to ruin", dr, h.Name), Role: theme.RoleNegative}}})
			}
		}
	}
	if c.St.Harbinger != nil && (prev.St.Harbinger == nil || prev.St.Harbinger.Key != c.St.Harbinger.Key) {
		changes = append(changes, Change{Square, Para{{Text: capFirst(c.St.Harbinger.Name) + " has come to " + c.Places[Square].Name + ".", Role: theme.RoleWarning}}})
	}
	// The log's own headlines, oldest first, a handful.
	var news []string
	for _, l := range c.St.Log {
		if l.Tick <= prev.St.Tick || l.Type == "debug" || l.Type == "info" {
			continue
		}
		news = append(news, strings.TrimSpace(l.Message))
	}
	if len(news) > 4 {
		news = news[len(news)-4:]
	}
	for _, n := range news {
		changes = append(changes, Change{"", Para{{Text: "· " + n, Role: theme.RoleDim}}})
	}
	return d, changes
}

func durationWords(ms int) string {
	m := ms / 60000
	if m < 60 {
		return fmt.Sprintf("%d minutes", m)
	}
	if m%60 == 0 {
		return fmt.Sprintf("%dh", m/60)
	}
	return fmt.Sprintf("%dh %02dm", m/60, m%60)
}

// ChangedPlaces reports which places differ since prev (the minimap's marks).
func (c *City) ChangedPlaces(prev *City) map[PlaceKey]bool {
	out := map[PlaceKey]bool{}
	if prev == nil {
		return out
	}
	_, ch := c.Diff(prev)
	for _, x := range ch {
		if x.Place != "" {
			out[x.Place] = true
		}
	}
	return out
}
