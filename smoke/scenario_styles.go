package smoke

import (
	"fmt"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// Play styles for the styles scenario (and -style on the command line).
const (
	StyleGreedy    = "greedy"
	StyleIdle      = "idle"
	StyleHarbinger = "harbinger"
	StyleSuccumb   = "succumber"
	StyleCosmic    = "cosmic"
)

// IdleCheckIn is how often the idle style looks at the game: decisions only
// happen at check-ins, with ticks running in between.
const IdleCheckIn = 3 * time.Hour

// StyleNames lists the styles in report order.
func StyleNames() []string {
	return []string{StyleGreedy, StyleIdle, StyleHarbinger, StyleSuccumb, StyleCosmic}
}

// firstLastPassageAge is the first age whose prestige can bring the Last
// Passage: the first age of the final epoch.
func firstLastPassageAge() string {
	for _, a := range config.AgeOrder() {
		if config.IsFinalEpoch(config.EpochForAge(a)) {
			return a
		}
	}
	return ""
}

// ApplyStyle turns base into the named style. Unknown names are an error.
func ApplyStyle(base Config, style string) (Config, error) {
	c := base
	c.Style = style
	switch style {
	case StyleGreedy:
	case StyleIdle:
		every := int(IdleCheckIn / game.BaseTickInterval)
		c.DecideEvery, c.CheckEvery = every, every
		// Caps fill between check-ins, so "no progress" needs a longer window.
		c.SoftlockSpan = 4 * IdleCheckIn
	case StyleHarbinger:
		c.Harbinger = HarbingerBoth
	case StyleSuccumb:
		c.Catastrophe = "succumb"
	case StyleCosmic:
		c.InviteCosmic = true
		c.LastPassage = "succumb"
		c.PrestigeAge = firstLastPassageAge()
	default:
		return c, fmt.Errorf("unknown style %q (want %s)", style, strings.Join(StyleNames(), ", "))
	}
	return c, nil
}

func runStyles(e *Env, res *Result) {
	styles := StyleNames()
	if e.Style != "" {
		styles = []string{e.Style}
	}
	seeds := e.seeds(2)
	var rows []string
	for _, style := range styles {
		base := e.Base
		base.Cycles, base.MaxSim = 1, 500*time.Hour
		if style == StyleCosmic {
			base.MaxSim = 1000 * time.Hour
		}
		cfg, err := ApplyStyle(base, style)
		if err != nil {
			res.fail("config", "%v", err)
			continue
		}
		sum := runBotSet(e, res, "style-"+style, "styles", cfg, seeds)
		rows = append(rows, styleRow(style, sum))
	}
	res.Summary = fmt.Sprintf("%d style(s) x %d seed(s)", len(styles), len(seeds))
	res.section("Outcomes per style", "| style | outcomes | furthest age | first prestige (median) | ages on target | anomalies | succumbed | cosmic legacy |\n|---|---|---|---|---|---|---|---|\n%s\n\nPer-style pacing and events are in style-<name>.md next to this report.",
		strings.Join(rows, "\n"))
	for _, sum := range res.Progression {
		var sb strings.Builder
		sum.writePacingTable(&sb)
		res.section("Pacing: "+strings.TrimPrefix(sum.Mode, "style-"), "%s", sb.String())
	}
}

func styleRow(style string, sum *Summary) string {
	outcomes := map[string]int{}
	furthest, fi := "", -1
	order := map[string]int{}
	for i, a := range config.AgeOrder() {
		order[a] = i
	}
	var prestige []float64
	succumbed, legacy := 0, 0
	for _, r := range sum.Runs {
		outcomes[r.Outcome]++
		if order[r.FinalAge] > fi {
			fi, furthest = order[r.FinalAge], r.FinalAge
		}
		for _, c := range r.Cycles {
			if c.Cycle == 1 {
				prestige = append(prestige, c.Seconds)
			}
		}
		succumbed += r.Stats.CatastrophesSuccumbed
		if r.CosmicLegacy {
			legacy++
		}
	}
	onTarget, graded := 0, 0
	for _, p := range sum.Pacing {
		if p.Verdict != VerdictNone {
			graded++
			if p.Verdict == VerdictOK {
				onTarget++
			}
		}
	}
	first := "not reached"
	if len(prestige) > 0 {
		_, med, _ := spread(prestige)
		first = fmt.Sprintf("%s (%d/%d seeds)", dur(med), len(prestige), len(sum.Runs))
	}
	return fmt.Sprintf("| %s | %s | %s | %s | %d/%d | %d | %d | %d/%d |", style, countStr(outcomes), furthest, first,
		onTarget, graded, sum.Anomalies, succumbed, legacy, len(sum.Runs))
}
