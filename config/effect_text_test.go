package config

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Guard tests for player-facing config text. They pin what descriptions claim
// to what the data does, so a balance pass or a new entry cannot quietly
// reintroduce a number that is off by 10^6 or an effect that does not exist.

var digitRe = regexp.MustCompile(`[0-9]`)

// TestDurationText pins the wall-clock format to the UI's formatTicks shape.
func TestDurationText(t *testing.T) {
	cases := map[int]string{0: "~0s", 5: "~10s", 30: "~1m", 144: "~4m 48s", 180: "~6m", 10000: "~5h 33m", 50000: "~1d 3h"}
	for ticks, want := range cases {
		if got := DurationText(ticks); got != want {
			t.Errorf("DurationText(%d) = %q, want %q", ticks, got, want)
		}
	}
}

// TestBuildingDescriptionsComeFromEffects: every building's mechanical
// sentence is the one built from its runtime Effects, and the hand-written
// flavor in front of it quotes no numbers of its own (those would go stale).
func TestBuildingDescriptionsComeFromEffects(t *testing.T) {
	for _, d := range BaseBuildings() {
		txt := buildingEffectText(d)
		if !strings.HasSuffix(d.Description, txt) {
			t.Errorf("%s: description %q does not end with its effect text %q", d.Key, d.Description, txt)
			continue
		}
		prose := strings.TrimSpace(strings.TrimSuffix(d.Description, txt))
		// A building the engine drives without an Effect (the Geographic
		// Society) may state its staffing by hand.
		prose = strings.Replace(prose, "("+strconv.Itoa(d.WorkerCapacity)+" workers)", "", 1)
		if digitRe.MatchString(prose) || strings.Contains(prose, "/tick") {
			t.Errorf("%s: flavor %q quotes numbers; put them in Effects instead", d.Key, prose)
		}
		for _, e := range d.Effects {
			if e.Type == "production" && e.Value > 0 {
				want := "+" + FormatAmount(e.Value) + " " + ResourceLabel(e.Target) + "/tick"
				if !strings.Contains(d.Description, want) {
					t.Errorf("%s: description %q is missing %q", d.Key, d.Description, want)
				}
			}
		}
		// Game speed is fixed at 1x: wonders no longer raise a speed cap, so
		// no description may promise one.
		if d.Category == "wonder" && strings.Contains(strings.ToLower(d.Description), "speed cap") {
			t.Errorf("%s: wonder description mentions a speed cap, which no longer exists: %q", d.Key, d.Description)
		}
	}
}

// eventFragments returns the phrases an event's LogMessage must contain for
// each effect, lowercased.
func eventFragments(e EventDef) (frags []string, hasSteal, hasRate bool) {
	for _, eff := range e.Effects {
		switch eff.Type {
		case "instant_resource":
			frags = append(frags, "+"+FormatAmount(eff.Value)+" "+ResourceLabel(eff.Target))
		case "steal_resource":
			hasSteal = true
			frags = append(frags, FormatAmount(eff.Value)+" "+ResourceLabel(eff.Target))
		case "production":
			hasRate = true
			frags = append(frags, strings.ToLower(RateText(eff.Target, eff.Value)))
		case "worker_loss":
			frags = append(frags, FormatPercent(eff.Value)+" of your workers")
		}
	}
	for i := range frags {
		frags[i] = strings.ToLower(frags[i])
	}
	return frags, hasSteal, hasRate
}

// raidEvents are the raid-type events the Army PR owns; their text is left to
// it, so they are exempt from the strict effect check below.
var raidEvents = map[string]bool{"tribal_raid": true, "bandit_raid": true, "pirate_attack": true}

// bannedEffectClaims are phrases event text used for effects no event has.
var bannedEffectClaims = []string{
	"doubled", "twice as", "no disasters", "building costs", "military",
	"population up", "production up", "production down", "research speed",
	"research up", "all trade routes", "ticks",
}

// TestEventTextMatchesEffects: an event's log line states every effect it
// applies, with the real amount and duration, and no number it does not have.
func TestEventTextMatchesEffects(t *testing.T) {
	for _, e := range allEventDefs() {
		if raidEvents[e.Key] {
			continue
		}
		msg := strings.ToLower(e.LogMessage)
		frags, hasSteal, hasRate := eventFragments(e)
		if hasRate && e.Duration > 0 {
			frags = append(frags, "for "+DurationText(e.Duration))
		}
		sort.Slice(frags, func(i, j int) bool { return len(frags[i]) > len(frags[j]) })
		rest := msg
		for _, f := range frags {
			if !strings.Contains(rest, f) {
				t.Errorf("%s: log %q is missing %q", e.Key, e.LogMessage, f)
				continue
			}
			rest = strings.Replace(rest, f, "", 1)
		}
		if hasSteal && !strings.Contains(msg, "up to") {
			t.Errorf("%s: a loss is capped at what you hold; say %q: %q", e.Key, "up to", e.LogMessage)
		}
		if digitRe.MatchString(rest) {
			t.Errorf("%s: log %q quotes a number no effect backs (left over: %q)", e.Key, e.LogMessage, rest)
		}
		if digitRe.MatchString(e.Description) {
			t.Errorf("%s: description %q quotes numbers", e.Key, e.Description)
		}
		for _, b := range bannedEffectClaims {
			if strings.Contains(msg, b) || strings.Contains(strings.ToLower(e.Description), b) {
				t.Errorf("%s: text claims %q, which no event effect does: %q / %q", e.Key, b, e.LogMessage, e.Description)
			}
		}
	}
}

// TestAwakeningTextMatchesEffects: each awakening states its boost, amount and
// duration.
func TestAwakeningTextMatchesEffects(t *testing.T) {
	for _, a := range Awakenings() {
		txt := strings.ToLower(a.FlavorText)
		want := []string{"for " + DurationText(a.Duration)}
		for _, e := range a.Effects {
			switch e.Type {
			case "production":
				want = append(want, strings.ToLower(RateText(e.Target, e.Value)))
			case "production_all":
				want = append(want, "all production +"+FormatPercent(e.Value))
			default:
				t.Errorf("%s: effect type %q has no wording here; add it", a.Key, e.Type)
			}
		}
		for _, w := range want {
			if !strings.Contains(txt, w) {
				t.Errorf("%s: flavor %q is missing %q", a.Key, a.FlavorText, w)
			}
		}
		for _, b := range bannedEffectClaims {
			if strings.Contains(txt, b) {
				t.Errorf("%s: flavor claims %q: %q", a.Key, b, a.FlavorText)
			}
		}
	}
}

// TestPrestigeDescriptionsMatchPerTier: the per-tier figure in the shop is
// the one the upgrade applies.
func TestPrestigeDescriptionsMatchPerTier(t *testing.T) {
	for _, u := range PrestigeUpgrades() {
		var want string
		switch u.EffectType {
		case "legacy":
			continue // a one-tier kit item: no per-tier figure
		case "rate_bonus":
			want = "+" + FormatPercent(u.PerTier)
		default:
			want = "+" + FormatAmount(u.PerTier)
		}
		if !strings.HasPrefix(u.Description, want+" ") {
			t.Errorf("%s: description %q should start with %q", u.Key, u.Description, want)
		}
	}
}

var milestoneNumRe = regexp.MustCompile(`[0-9][0-9,]*`)

// TestMilestoneDescriptionNumbers: every number in a milestone's description
// is one of its conditions, and play-time conditions read as wall-clock time.
func TestMilestoneDescriptionNumbers(t *testing.T) {
	for _, m := range Milestones() {
		desc := m.Description
		if m.MinTick > 0 {
			d := DurationText(m.MinTick)
			if !strings.Contains(desc, d) {
				t.Errorf("%s: %q should state the play time as %q", m.Key, desc, d)
			}
			desc = strings.Replace(desc, d, "", 1)
		}
		ok := map[float64]bool{}
		for _, v := range m.MinBuildings {
			ok[float64(v)] = true
		}
		for _, v := range m.MinResources {
			ok[v] = true
		}
		for _, v := range []int{m.MinPopulation, m.MinTechCount, m.MinBuildingSum.Count, m.MinTotalBuilt,
			m.MinSoldiersTrained, m.MinWonders, m.MinKnowledgeWorkers} {
			if v > 0 {
				ok[float64(v)] = true
			}
		}
		for _, n := range milestoneNumRe.FindAllString(desc, -1) {
			v, err := strconv.ParseFloat(strings.ReplaceAll(n, ",", ""), 64)
			if err != nil || !ok[v] {
				t.Errorf("%s: %q says %s, which is not one of its conditions", m.Key, m.Description, n)
			}
		}
		if strings.Contains(strings.ToLower(desc), " all ") {
			t.Errorf("%s: %q says \"all\"; name the count instead", m.Key, m.Description)
		}
	}
}

// TestTradeRouteDescriptionsNameGiveAndGet: a route description reads from the
// player's side, "<what you send> for <what you get>", naming every export
// before "for" and every import after it.
func TestTradeRouteDescriptionsNameGiveAndGet(t *testing.T) {
	for _, r := range BaseTradeRoutes() {
		d := strings.ToLower(r.Description)
		i := strings.Index(d, " for ")
		if i < 0 {
			t.Errorf("%s: %q has no \"for\" between what you send and what you get", r.Key, r.Description)
			continue
		}
		give, get := d[:i], d[i:]
		for res := range r.Export {
			if !strings.Contains(give, ResourceLabel(res)) {
				t.Errorf("%s: %q should name the export %q before \"for\"", r.Key, r.Description, ResourceLabel(res))
			}
		}
		for res := range r.Import {
			if !strings.Contains(get, ResourceLabel(res)) {
				t.Errorf("%s: %q should name the import %q after \"for\"", r.Key, r.Description, ResourceLabel(res))
			}
		}
	}
}

// TestTradeRouteMinAgeCoversBuilding: a route is not offered before its
// required building exists.
func TestTradeRouteMinAgeCoversBuilding(t *testing.T) {
	order := map[string]int{}
	for i, a := range AgeOrder() {
		order[a] = i
	}
	bld := BuildingByKey()
	for _, r := range BaseTradeRoutes() {
		b, ok := bld[r.RequiredBld]
		if !ok {
			t.Errorf("%s: unknown required building %q", r.Key, r.RequiredBld)
			continue
		}
		if order[r.MinAge] < order[b.RequiredAge] {
			t.Errorf("%s: opens in %s, %s arrives in %s", r.Key, r.MinAge, r.RequiredBld, b.RequiredAge)
		}
	}
}

// playerText collects every player-facing string config owns, labelled.
func playerText() map[string]string {
	out := map[string]string{}
	add := func(label, s string) {
		if s != "" {
			out[label] = s
		}
	}
	for _, d := range BaseBuildings() {
		add("building "+d.Key+" name", d.Name)
		add("building "+d.Key+" description", d.Description)
		add("building "+d.Key+" flavor", d.Flavor)
	}
	for _, x := range Technologies() {
		add("tech "+x.Key+" name", x.Name)
		add("tech "+x.Key+" description", x.Description)
	}
	for _, x := range BaseResources() {
		add("resource "+x.Key+" description", x.Description)
	}
	for _, x := range Ages() {
		add("age "+x.Key+" description", x.Description)
		add("age "+x.Key+" quip", x.Quip)
	}
	for _, x := range Epochs() {
		add("epoch "+x.Key+" description", x.Description)
		name, flavor := CatastropheInfo(x.Key)
		add("catastrophe "+x.Key+" name", name)
		add("catastrophe "+x.Key+" flavor", flavor)
	}
	for _, x := range EpochEventByKey() {
		add("epoch event "+x.Key+" name", x.Name)
		add("epoch event "+x.Key+" flavor", x.FlavorText)
	}
	for _, x := range allEventDefs() {
		if raidEvents[x.Key] {
			continue
		}
		add("event "+x.Key+" name", x.Name)
		add("event "+x.Key+" description", x.Description)
		add("event "+x.Key+" log", x.LogMessage)
	}
	for _, x := range Awakenings() {
		add("awakening "+x.Key+" name", x.Name)
		add("awakening "+x.Key+" flavor", x.FlavorText)
	}
	for _, x := range Milestones() {
		add("milestone "+x.Key+" name", x.Name)
		add("milestone "+x.Key+" description", x.Description)
		add("milestone "+x.Key+" flavor", x.Flavor)
	}
	for _, x := range MilestoneChains() {
		add("chain "+x.Key+" name", x.Name)
		add("chain "+x.Key+" flavor", x.Flavor)
	}
	for _, x := range PrestigeUpgrades() {
		add("prestige "+x.Key+" name", x.Name)
		add("prestige "+x.Key+" description", x.Description)
	}
	for _, x := range BaseTradeRoutes() {
		add("route "+x.Key+" name", x.Name)
		add("route "+x.Key+" description", x.Description)
	}
	for _, x := range BaseFactions() {
		add("civ "+x.Key+" backstory", x.Backstory)
		add("civ "+x.Key+" description", x.Description)
	}
	for _, x := range WorkerClasses() {
		add("worker class "+x.Domain+"/"+x.AgeKey, x.ClassName)
	}
	for _, x := range Harbingers() {
		add("harbinger "+x.Key+" name", x.Name)
		add("harbinger "+x.Key+" description", x.Description)
	}
	for _, m := range LogFlavorMoments() {
		for i, q := range logFlavorPools[m] {
			add("log flavor "+m+" "+strconv.Itoa(i), q)
		}
	}
	return out
}

var (
	rawKeyRe = regexp.MustCompile(`\b[a-z]+_[a-z_]+\b`)
	// styleBans are glossary and tone rules for player text: US spelling,
	// glossary terms, and stock words.
	styleBans = []string{
		"colour", "favour", "honour", "harbour", "centre", "theatre", "defence",
		"civilisation", "labour", "organis", "maximis", "fervour",
		"neighbour", "travelling", "realise", "recognise", "crystallise", "decentralise", "optimise", "optimising",
		"pop cap", "population cap", "military cap", "culture cap", "unlocks +",
		"gather rate", "standings", "villager",
		"transform", "revolutioni", "realm", "unleash", "limitless",
		"changes everything", "frontiers await", "ascension",
	}
)

// TestConfigPlayerTextStyle: config text uses no em or en dashes, no
// exclamation marks, no raw snake_case keys, and none of the banned spellings,
// retired terms or stock words.
func TestConfigPlayerTextStyle(t *testing.T) {
	text := playerText()
	labels := make([]string, 0, len(text))
	for l := range text {
		labels = append(labels, l)
	}
	sort.Strings(labels)
	for _, l := range labels {
		s := text[l]
		low := strings.ToLower(s)
		if strings.ContainsAny(s, "—–") {
			t.Errorf("%s: dash in %q", l, s)
		}
		if strings.Contains(s, "!") {
			t.Errorf("%s: exclamation mark in %q", l, s)
		}
		if k := rawKeyRe.FindString(s); k != "" {
			t.Errorf("%s: raw key %q in %q", l, k, s)
		}
		for _, b := range styleBans {
			if strings.Contains(low, b) {
				t.Errorf("%s: %q in %q", l, b, s)
			}
		}
	}
}
