package game

import (
	"fmt"
	"github.com/espresso20/ageforge/detmath"
	"math"
	"math/rand"
	"strings"

	"github.com/espresso20/ageforge/config"
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
		story := ge.applyRoadEvent(active, e)
		active.RoadEventsSeen = append(active.RoadEventsSeen, story)
		ge.addLog("event", story)
	}
}

func (ge *GameEngine) applyRoadEvent(a *ActiveExpedition, e *RoadEvent) string {
	words, _ := ge.rules.MissionWords(a.Age)
	place := a.Place
	if place == "" {
		place = "road"
	}
	foe, helper, weather, find := words.Foes[e.FoeIndex], words.Helpers[e.HelperIndex], words.Weather[e.WeatherIndex], words.Finds[e.FindIndex]
	value := func(template, key, value string) string { return strings.ReplaceAll(template, key, value) }
	story := e.Story
	story = value(story, "<place>", place)
	story = value(story, "<foe>", foe)
	story = value(story, "<helper>", helper)
	story = value(story, "<weather>", weather)
	story = value(story, "<find>", find)
	switch e.Kind {
	case "storm":
		delay := max(1, int(math.Ceil(float64(a.TotalTicks)*.15)))
		a.TicksLeft += delay
		story = value(story, "<time>", expRoadDuration(delay))
	case "helpers":
		saved := max(1, int(math.Ceil(float64(a.TicksLeft)*.25)))
		a.TicksLeft = max(1, a.TicksLeft-saved)
		story = value(story, "<time>", expRoadDuration(saved))
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
			story = value(story, "<gold>", strings.TrimSuffix(Amount(gold, "gold"), " gold"))
		} else {
			e.Kind = "foes"
			a.SuccessChance = math.Max(.05, a.SuccessChance-.08)
			story = e.AuxStory
			story = value(story, "<place>", place)
			story = value(story, "<foe>", foe)
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
		if !a.Military() {
			// A civil or scouting party also loses part of its haul.
			scaleActiveRewards(a, .85)
			story += " " + value(e.AmbushStory, "<foe>", foe)
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
			offer := ge.Military.board.makeOffer(ge.gameRNG(), w, TierShort, ge.tick)
			offer.Description = "Found by this mission near the ruins at " + place + "."
			ge.Military.board.Slots[target].Offer = offer
		}
	}
	return story
}

func scaleActiveRewards(a *ActiveExpedition, factor float64) {
	for _, m := range []map[string]float64{a.BaseRewards, a.SuccessRewards, a.FailureRewards} {
		for k := range m {
			m[k] *= factor
		}
	}
}

func expRoadDuration(ticks int) string {
	seconds := ticks * config.TickSeconds
	if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	}
	minutes := (seconds + 30) / 60
	return fmt.Sprintf("%d minutes", minutes)
}
