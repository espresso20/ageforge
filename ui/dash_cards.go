package ui

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/window/kit"
)

// dash_cards.go is the model of the dashboard's Buildings box: a card for
// each building the age offers, what it costs for the amount the player buys
// at once, whether it can be built and, when it cannot, the one thing that
// stops it. The box (dash_cards_box.go) draws and orders them.

// buyChoice is how many copies a Build button builds: one, ten, or as many as
// can be afforded now.
type buyChoice int

const (
	buyOne buyChoice = iota
	buyTen
	buyMax
	buyChoices
)

// buyWords are the keys' labels, in the order of buyChoice, and buyCounts the
// count the build command takes for each (none for one copy).
var (
	buyWords  = [buyChoices]string{"x1", "x10", "max"}
	buyCounts = [buyChoices]string{"", "10", "max"}
	// buyArgs is what the buy command takes for each.
	buyArgs = [buyChoices]string{"1", "10", "max"}
)

// The other words the box says.
var (
	txtBuy      = rowid.E("ui.dash_cards.txtBuy.01", "Buy")
	txtBuild    = rowid.E("ui.dash_cards.txtBuild.01", "Build")
	txtNoIncome = rowid.E("ui.dash_cards.txtNoIncome.01", "no income")
	// txtLimit is the card's reason at its count limit.
	txtLimit = rowid.E("ui.dash_cards.txtLimit.01", "Building limit reached.")
	// The detail box's buttons, and the row that counts the older buildings.
	txtSell    = rowid.E("ui.dash_cards.txtSell.01", "Sell")
	txtUpgrade = rowid.E("ui.dash_cards.txtUpgrade.01", "Upgrade")
	txtOlder   = rowid.E("ui.dash_cards.txtOlder.01", "older buildings")
	// txtNextStep is the plate on the card of the next step.
	txtNextStep = rowid.E("ui.dash_cards.txtNextStep.01", " next step ")
)

// nearShare is the share of a price in hand from which a card that cannot be
// built yet is listed among the near ones.
const nearShare = 0.5

// The groups a card is listed in.
const (
	groupBuild = iota // can be built now
	groupNear         // at least half the price in hand
	groupLater        // the rest, and what a lock or a limit holds
)

// costPart is one resource of a price: its text, whether the player holds
// enough of it, and what it is made of (the resource, held and needed).
type costPart struct {
	text       string
	ok         bool
	res        string
	have, need float64
}

// priceTip is the tooltip of a price: a line a resource, with what is held and
// what is needed, the short ones marked with the sign the reason line uses.
func priceTip(price []costPart) kit.TipText {
	var t kit.TipText
	for _, p := range price {
		mark := "  "
		if !p.ok {
			mark = "! "
		}
		t.Lines = append(t.Lines, mark+rowid.Sprintf("ui.dash_cards.priceTip.01", "%s: have %s, need %s", game.ResourceName(p.res), textfmt.Number(p.have), textfmt.Number(p.need)))
	}
	return t
}

// buildCard is one building as the box lists it.
type buildCard struct {
	key, name string
	owned     int
	out       string // what one copy makes, per second
	price     []costPart
	// frac is how much of the price of one copy is in hand, 0 to 1 (the gauge).
	frac float64
	// share is the part of what the player holds that one copy takes: the
	// order of the cards that can be built, cheapest first.
	share    float64
	canBuild bool
	group    int
	// reason is what stops the card, and wait the time until it can be built
	// ("" when none is known, or the card is not waiting on a resource).
	reason, wait string
	// cmd is the command a Build button carries.
	cmd string
	// older marks a card of an older building, which has no Build button.
	older bool
}

// command is a command line of words.
func command(words ...string) string { return strings.Join(words, " ") }

// buildCmd is the command that builds copies of key by choice.
func buildCmd(key string, c buyChoice) string {
	if n := buyCounts[c]; n != "" {
		return command(buildLine(key), n)
	}
	return buildLine(key)
}

// cardOutput is what a building's description says after its first sentence,
// what a copy makes: "Roman engineering feats improve iron output. +21.6
// iron/s (5 workers)." is "+21.6 iron/s (5 workers)".
func cardOutput(desc string) string {
	if i := strings.Index(desc, ". "); i >= 0 {
		desc = desc[i+2:]
	}
	return textfmt.Capitalize(strings.TrimSuffix(strings.TrimSpace(desc), "."))
}

// buildCards lists the buildings of the current age as cards, in no
// particular order: the unlocked ones, and the ones a tech still locks.
// Wonders are not listed; they have their own panel.
func buildCards(state game.GameState, buy buyChoice) []buildCard {
	set := state.Ruleset()
	var cards []buildCard
	for _, key := range sortedKeysOf(state.Buildings) {
		bs := state.Buildings[key]
		if bs.AgeKey != state.Age || bs.Category == "wonder" || bs.IsLegacy || (!bs.Unlocked && bs.LockTech == "") {
			continue
		}
		def, ok := set.Building(key)
		if !ok {
			continue
		}
		c := buildCard{key: key, name: bs.Name, owned: bs.Count,
			out: cardOutput(buildingText(bs))}
		queued := 0
		for _, q := range state.BuildQueue {
			if q.Key == key {
				queued++
			}
		}
		limit := bs.AtMaxCount || (def.MaxCount > 0 && bs.Count+queued >= def.MaxCount)
		// What one copy takes, and what the amount chosen takes.
		one := bs.NextCost
		n := 1
		switch buy {
		case buyTen:
			n = 10
		case buyMax:
			n = max(1, state.MaxBuildable(key))
		}
		shown := one
		if n > 1 {
			if cost, ok := state.BuildBatch(key, n); ok {
				shown = cost
			}
		}
		for _, res := range sortedKeysOf(shown) {
			if shown[res] == 0 {
				continue
			}
			c.price = append(c.price, costPart{text: game.Amount(shown[res], res), ok: state.Resources[res].Amount >= shown[res], res: res, have: state.Resources[res].Amount, need: shown[res]})
		}
		c.frac, c.share = priceHeld(state, one)
		c.canBuild = bs.CanBuild && !limit
		c.cmd = buildCmd(key, buy)
		switch {
		case !bs.Unlocked:
			c.reason, c.group = lockedNeedsLine(bs), groupLater
		case limit:
			c.reason, c.group = txtLimit.String(), groupLater
			if why := state.QueuedRefusal(key); why != "" {
				c.reason = why
			}
		case c.canBuild:
			c.group = groupBuild
		default:
			c.reason, c.wait = shortOf(state, one)
			c.group = groupLater
			if c.frac >= nearShare {
				c.group = groupNear
			}
		}
		cards = append(cards, c)
	}
	return cards
}

// olderCards lists the older buildings the town still holds as cards with no
// price and no Build button.
func olderCards(state game.GameState) []buildCard {
	set := state.Ruleset()
	var cards []buildCard
	for _, key := range sortedKeysOf(state.Buildings) {
		bs := state.Buildings[key]
		if _, ok := set.Building(key); !ok || !bs.IsLegacy || bs.Count == 0 || bs.Category == "wonder" {
			continue
		}
		cards = append(cards, buildCard{key: key, name: bs.Name, owned: bs.Count,
			out: cardOutput(buildingText(bs)), older: true, group: groupBuild})
	}
	return cards
}

// priceHeld is how much of a price the player holds: frac, the mean over its
// resources of the share held (1 for a resource held in full), and share, the
// mean part of the stock the price takes.
func priceHeld(state game.GameState, cost map[string]float64) (frac, share float64) {
	n := 0
	for res, need := range cost {
		if need <= 0 {
			continue
		}
		have := state.Resources[res].Amount
		frac += math.Min(1, have/need)
		share += need / math.Max(have, 1)
		n++
	}
	if n == 0 {
		return 1, 0
	}
	return frac / float64(n), share / float64(n)
}

// maxWaitSeconds is the longest wait worth printing; a rate that gives a longer
// one counts as no income.
const maxWaitSeconds = 365 * 24 * 3600

// shortOf is what stops a price being paid: the resource the player is
// furthest from affording, as "short of 44.1K gold", and how long at the
// current net rate until it is held ("in 1m 44s"), or "no income" when that
// resource is not rising. Both "" when nothing is short.
func shortOf(state game.GameState, cost map[string]float64) (reason, wait string) {
	var worst string
	var worstWait, worstDeficit float64
	for _, res := range sortedKeysOf(cost) {
		deficit := cost[res] - state.Resources[res].Amount
		if deficit <= 0 {
			continue
		}
		w := math.Inf(1)
		if rate := state.Resources[res].Rate; rate > 0 {
			// A rate that is all but zero is no income: a year is the
			// longest wait worth printing.
			if w = deficit * tickSeconds(state.TickIntervalMs) / rate; w > maxWaitSeconds {
				w = math.Inf(1)
			}
		}
		// The furthest: the longest wait, and among those with no income the
		// biggest part of the price missing.
		if worst == "" || w > worstWait || (w == worstWait && deficit/cost[res] > worstDeficit) {
			worst, worstWait, worstDeficit = res, w, deficit/cost[res]
		}
	}
	if worst == "" {
		return "", ""
	}
	reason = rowid.Sprintf("ui.dash_cards.shortOf.01", "short of %s", game.Amount(cost[worst]-state.Resources[worst].Amount, worst))
	if math.IsInf(worstWait, 1) {
		return reason, txtNoIncome.String()
	}
	return reason, rowid.Sprintf("ui.dash_cards.shortOf.02", "in %s", textfmt.Duration(time.Duration(worstWait*float64(time.Second))))
}

// orderCards sorts cards into the order the box lists them: the ones that can
// be built, cheapest first, then the near ones, closest first, then the
// rest, closest first. It returns the keys and how many are in the first two
// groups (the ones listed before the quiet row). The card of the Next step
// (first) leads whatever group it is in.
func orderCards(cards []buildCard, first string) (keys []string, shown int) {
	sorted := append([]buildCard(nil), cards...)
	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		if a.key != b.key && (a.key == first || b.key == first) && first != "" {
			return a.key == first
		}
		if a.group != b.group {
			return a.group < b.group
		}
		switch {
		case a.group == groupBuild && a.share != b.share:
			return a.share < b.share
		case a.group != groupBuild && a.frac != b.frac:
			return a.frac > b.frac
		}
		return a.name < b.name
	})
	for _, c := range sorted {
		keys = append(keys, c.key)
		if c.group != groupLater || c.key == first {
			shown++
		}
	}
	return keys, shown
}
