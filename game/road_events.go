package game

import (
	"fmt"
	"github.com/espresso20/ageforge/detmath"
	"math"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

var roadKinds = weightedRoadKinds()

func weightedRoadKinds() []string {
	weights := []struct {
		kind string
		n    int
	}{
		{"storm", config.RoadStormWeight}, {"helpers", config.RoadHelpersWeight}, {"find", config.RoadFindWeight},
		{"toll", config.RoadTollWeight}, {"foes", config.RoadFoesWeight}, {"ambush", config.RoadAmbushWeight},
		{"fair", config.RoadFairWeight}, {"ruins", config.RoadRuinsWeight},
	}
	var out []string
	for _, w := range weights {
		for i := 0; i < w.n; i++ {
			out = append(out, w.kind)
		}
	}
	return out
}

func rollRoadEvents(rng *rand.Rand, offer MissionOffer, ticks int, words config.MissionWords) []RoadEvent {
	mean := offer.LengthMinutes / 40
	if offer.Tier == TierShort {
		mean = offer.LengthMinutes / 80
	}
	// Knuth Poisson draw; cap only protects pathological migrated saves.
	l, p, n := detmath.Exp(-mean), 1.0, 0
	for p > l && n < 100 {
		n++
		p *= rng.Float64()
	}
	n--
	if n < 0 {
		n = 0
	}
	out := make([]RoadEvent, n)
	for i := range out {
		at := int(math.Round(float64(ticks) * (.1 + float64(.8*rng.Float64()))))
		at = max(1, min(ticks, at))
		kind := roadKinds[rng.Intn(len(roadKinds))]
		out[i] = RoadEvent{
			Tick: at, Kind: kind,
			Story:       words.RoadTemplate(kind, rng.Intn(words.RoadTemplateCount(kind))),
			AuxStory:    words.RoadTemplate("foes", rng.Intn(words.RoadTemplateCount("foes"))),
			AmbushStory: words.RoadTemplate("ambush_haul", rng.Intn(words.RoadTemplateCount("ambush_haul"))),
			FoeIndex:    rng.Intn(3), HelperIndex: rng.Intn(3), WeatherIndex: rng.Intn(3), FindIndex: rng.Intn(3),
		}
	}
	return out
}

func (ge *GameEngine) fireRoadEvents(active *ActiveExpedition, elapsed int) {
	for i := range active.RoadEvents {
		e := &active.RoadEvents[i]
		if e.Fired || e.Tick > elapsed {
			continue
		}
		e.Fired = true
		told := ge.roadEventTold(active, e)
		words, _ := ge.rules.MissionWords(active.Age)
		story := tellRoadStory(words, active.roadPlace(), *e, told)
		// The save holds the story as told (the numbers in, the words of the
		// place and the people still marked), not the story as read.
		active.RoadTold = append(active.RoadTold, RoadTold{Event: i, Text: told})
		active.RoadEventsSeen = append(active.RoadEventsSeen, story)
		// An ambush costs the party soldiers; the other stories are not about people.
		var about LogTag
		if e.Kind == "ambush" {
			about = TagPeople | TagArmy | TagBad
		}
		ge.addLogAbout("event", about, story)
	}
}

// RoadTold is a road event's story as the save holds it: the story of the
// event numbered Event in RoadEvents, with what the road did written in
// (the time lost, the odds, the gold) and the words of the script left as
// <place>, <foe>, <helper>, <weather> and <find>, so the story can be read in
// whichever script the game is read in. Text is English, never VibeMode's.
type RoadTold struct {
	Event int    `json:"event"`
	Text  string `json:"text"`
}

// roadPlace is where the road event happens: the mission's place, else "road".
func (a *ActiveExpedition) roadPlace() string {
	if a.Place == "" {
		return "road"
	}
	return a.Place
}

// tellRoadStory fills the words of an age into a story told by roadEventTold:
// the place and the foe, helper, weather and find the event drew by index.
func tellRoadStory(words config.MissionWords, place string, e RoadEvent, told string) string {
	if vibe.On() {
		told = vibeRoadTold(words, e, told)
	}
	return strings.NewReplacer(
		"<place>", place,
		"<foe>", wordAt(words.Foes[:], e.FoeIndex),
		"<helper>", wordAt(words.Helpers[:], e.HelperIndex),
		"<weather>", wordAt(words.Weather[:], e.WeatherIndex),
		"<find>", wordAt(words.Finds[:], e.FindIndex),
	).Replace(told)
}

// applyRoadEvent carries a road event out and returns its story as read now.
func (ge *GameEngine) applyRoadEvent(a *ActiveExpedition, e *RoadEvent) string {
	words, _ := ge.rules.MissionWords(a.Age)
	return tellRoadStory(words, a.roadPlace(), *e, ge.roadEventTold(a, e))
}

// roadEventTold carries a road event out (it moves the mission's time, odds
// and loot, takes gold and soldiers, or posts an offer) and returns its story
// told with the numbers in and the words still to fill (RoadTold).
func (ge *GameEngine) roadEventTold(a *ActiveExpedition, e *RoadEvent) string {
	value := func(template, key, value string) string { return strings.ReplaceAll(template, key, value) }
	story := e.Story
	switch e.Kind {
	case "storm":
		delay := max(1, int(math.Ceil(float64(a.TotalTicks)*.15)))
		a.TicksLeft += delay
		story = value(story, "<time>", plainRoadDuration(delay))
	case "helpers":
		saved := max(1, int(math.Ceil(float64(a.TicksLeft)*.25)))
		a.TicksLeft = max(1, a.TicksLeft-saved)
		story = value(story, "<time>", plainRoadDuration(saved))
	case "find":
		factor := 1.4
		if a.Rare {
			factor = 1.8
		}
		scaleActiveRewards(a, factor)
		story = value(story, "<percent>", fmt.Sprint(int((factor-1)*100)))
	case "toll":
		gold := 2 * 60 / config.TickSeconds * ge.townIncome("gold")
		if ge.Resources.Get("gold") >= gold {
			ge.Resources.Remove("gold", gold)
			story = value(story, "<gold>", textfmt.Number(gold))
		} else {
			e.Kind = "foes"
			a.SuccessChance = math.Max(.05, a.SuccessChance-.08)
			story = e.AuxStory
			story = value(story, "<odds>", fmt.Sprint(int(a.SuccessChance*100)))
		}
	case "foes":
		a.SuccessChance = math.Max(.05, a.SuccessChance-.08)
		story = value(story, "<odds>", fmt.Sprint(int(a.SuccessChance*100)))
	case "fair":
		a.SuccessChance = math.Min(.95, a.SuccessChance+.06)
		story = value(story, "<odds>", fmt.Sprint(int(a.SuccessChance*100)))
	case "ambush":
		// An ambush falls on the party, whatever the mission: a tenth of it.
		lost := min(a.Soldiers, max(1, int(math.Ceil(float64(float64(a.Soldiers)*.10)))))
		ge.soldiersFall(lost)
		a.Soldiers -= lost
		story = value(story, "<n>", fmt.Sprint(lost))
		if a.Party == 0 && a.Soldiers == 0 {
			// No party went (an age with no army): the ambush takes haul only.
			scaleActiveRewards(a, .85)
			story = e.AmbushStory
		} else if !a.Military() {
			// A civil or scouting party also loses part of its haul.
			scaleActiveRewards(a, .85)
			story += " " + e.AmbushStory
		}
	case "ruins":
		if w, ok := ge.rules.MissionWords(a.Age); ok {
			var target = -1
			for i := range ge.Military.board.Slots {
				if ge.Military.board.Slots[i].Offer == nil {
					target = i
					break
				}
			}
			if target < 0 {
				target = 0
			}
			offer := ge.Military.board.makeOffer(ge.gameRNG(), w, TierShort, ge.tick, ge.Military.rareFactor())
			if a.Words.Made {
				offer.Words.Found = a.Words.Place + 1
				offer.draw(w)
			} else {
				// A mission written by hand: the offer's text is what it is.
				offer.Words, offer.Description = OfferWords{}, ruinsFoundPlain+a.roadPlace()+"."
			}
			ge.Military.board.Slots[target].Offer = offer
		}
	}
	return story
}

func scaleActiveRewards(a *ActiveExpedition, factor float64) {
	a.HaulScale = a.haulScale() * factor
	for _, m := range []map[string]float64{a.BaseRewards, a.SuccessRewards, a.FailureRewards} {
		for k := range m {
			m[k] *= factor
		}
	}
}

// haulScale is what the road has done to the loot: 1 until a road event scales it.
func (a *ActiveExpedition) haulScale() float64 {
	if a.HaulScale == 0 {
		return 1
	}
	return a.HaulScale
}

// plainRoadDuration is a span of the road as a story keeps it: in plain
// English, whatever script the game is read in (RoadTold).
func plainRoadDuration(ticks int) string {
	seconds := ticks * config.TickSeconds
	if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	}
	return fmt.Sprintf("%d minutes", (seconds+30)/60)
}

var roadDurationRe = regexp.MustCompile(`^(\d+) (seconds|minutes)$`)

// sayRoadDuration is a span of the road kept by plainRoadDuration, said in the
// script's words. Anything else is given back as it is.
func sayRoadDuration(kept string) string {
	m := roadDurationRe.FindStringSubmatch(kept)
	if m == nil {
		return kept
	}
	n, _ := strconv.Atoi(m[1])
	if m[2] == "seconds" {
		return vibe.Sprintf("game.road_events.expRoadDuration.01", "%d seconds", n)
	}
	return vibe.Sprintf("game.road_events.expRoadDuration.02", "%d minutes", n)
}

// roadSlotRe finds the numbers roadEventTold writes into a story's template.
var roadSlotRe = regexp.MustCompile(`<(?:time|percent|gold|odds|n)>`)

// vibeRoadTold tells a road event's story in the words of the script in force.
// told is the story as the save holds it: the event's templates (RoadEvent
// Story, AuxStory and AmbushStory, in plain English) with the numbers written
// in. Each template found in it is taken out and its line in the script put in
// its place, with the same numbers; the rest is left as it is, the words still
// to fill included. A template the script has no line for, or that is not in
// the story, changes nothing.
func vibeRoadTold(words config.MissionWords, e RoadEvent, told string) string {
	type span struct {
		lo, hi int
		text   string
	}
	var spans []span
	seen := map[string]bool{}
	for _, template := range []string{e.Story, e.AuxStory, e.AmbushStory} {
		if template == "" || seen[template] {
			continue
		}
		seen[template] = true
		line := words.RoadLine(template)
		if line == template {
			continue
		}
		slots := roadSlotRe.FindAllString(template, -1)
		var pattern strings.Builder
		for i, literal := range roadSlotRe.Split(template, -1) {
			pattern.WriteString(regexp.QuoteMeta(literal))
			if i < len(slots) {
				pattern.WriteString(`(.+?)`)
			}
		}
		at := regexp.MustCompile(pattern.String()).FindStringSubmatchIndex(told)
		if at == nil {
			continue
		}
		var fill []string
		for i, slot := range slots {
			value := told[at[2+2*i]:at[3+2*i]]
			if slot == "<time>" {
				value = sayRoadDuration(value)
			}
			fill = append(fill, slot, value)
		}
		spans = append(spans, span{at[0], at[1], strings.NewReplacer(fill...).Replace(line)})
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].lo < spans[j].lo })
	var out strings.Builder
	pos := 0
	for _, sp := range spans {
		if sp.lo < pos {
			continue
		}
		out.WriteString(told[pos:sp.lo])
		out.WriteString(sp.text)
		pos = sp.hi
	}
	out.WriteString(told[pos:])
	return out.String()
}
