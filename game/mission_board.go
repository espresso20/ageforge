package game

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/espresso20/ageforge/config"
)

// The mission board: nine offers, three of each tier, rolled up front.
//
// Every offer holds what a launch used to roll (its length, its odds and its
// haul in minutes of income), so the player picks between offers. Costs and
// the haul in resources are still worked out at launch, from the age.
//
// All randomness comes from the run's seeded engine RNG (the one catastrophes
// and encounters use), in a fixed order: a new grand offer draws its kind,
// then every offer draws its rarity, its objective, its place, its twist,
// its length, its odds and its haul. Nothing here iterates a map to decide
// anything random, so a save replays the same board.

// The tiers of the board, and the kinds of grand offer.
const (
	TierShort = "short"
	TierLong  = "long"
	TierGrand = "grand"

	GrandMilitary = "military"
	GrandCivil    = "civil"
)

// boardTiers is the tiers in board order: short, long, grand. The board's
// slots follow it, config.BoardOffersPerTier to a tier.
var boardTiers = []string{TierShort, TierLong, TierGrand}

// boardSize is how many offers the board holds.
var boardSize = len(boardTiers) * config.BoardOffersPerTier

// tierKind is the kind of mission a tier's offers are tuned as.
func tierKind(tier string) string {
	switch tier {
	case TierLong:
		return config.MissionExpedition
	case TierGrand:
		return config.MissionCampaign
	}
	return config.MissionScout
}

// kindTier is the tier of a mission kind.
func kindTier(kind string) string {
	switch kind {
	case config.MissionExpedition:
		return TierLong
	case config.MissionCampaign:
		return TierGrand
	}
	return TierShort
}

// slotTier is the tier of board slot i (0-based).
func slotTier(i int) string { return boardTiers[i/config.BoardOffersPerTier] }

// minutesToTicks is minutes of game time in ticks.
func minutesToTicks(minutes float64) int {
	return int(math.Round(minutes * 60 / config.TickSeconds))
}

// MissionOffer is one offer on the board: everything about a mission that
// was rolled when it was made.
type MissionOffer struct {
	Tier string `json:"tier"`
	// Grand is GrandMilitary or GrandCivil for a grand offer, "" otherwise.
	Grand string `json:"grand,omitempty"`
	// Age is the age the offer was made in; its era is the one a badge for
	// the mission is earned in.
	Age         string `json:"age"`
	Title       string `json:"title"`
	Description string `json:"description"`
	// LengthMinutes, SuccessChance and HaulMinutes are the rolls. The haul is
	// in minutes of the town's income; the resources are worked out at launch.
	LengthMinutes float64 `json:"length_minutes"`
	SuccessChance float64 `json:"success_chance"`
	HaulMinutes   float64 `json:"haul_minutes"`
	// Rare marks a Legendary offer.
	Rare        bool `json:"rare,omitempty"`
	ExpiresTick int  `json:"expires_tick,omitempty"`
}

// Military reports whether the offer is a military grand mission: the only
// kind that costs soldiers and counts the army.
func (o MissionOffer) Military() bool { return o.Tier == TierGrand && o.Grand == GrandMilitary }

// Kind is the config kind the offer is tuned as: scout, expedition or
// campaign.
func (o MissionOffer) Kind() string { return tierKind(o.Tier) }

// Category is ExpeditionMilitary for a military grand offer and
// ExpeditionScouting for every other. It picks the techs' terms and the
// encounter odds.
func (o MissionOffer) Category() string {
	if o.Military() {
		return ExpeditionMilitary
	}
	return ExpeditionScouting
}

// CostMinutes is how many minutes of income the offer costs: of soldier
// income for a military offer, of the age's two most-charged materials for
// any other.
func (o MissionOffer) CostMinutes() float64 {
	if o.Tier == TierGrand && o.Grand == GrandCivil {
		return config.CivilGrandCostMinutes
	}
	t, _ := config.MissionKindTuning(o.Kind())
	return t.CostMinutes
}

// boardSlot is one place on the board: an offer, or the tick a taken
// offer's replacement is made.
type boardSlot struct {
	Offer      *MissionOffer `json:"offer,omitempty"`
	RefillTick int           `json:"refill_tick,omitempty"`
}

// MissionBoard is the board, the counters its rare rule runs on, and the age
// it was made for.
type MissionBoard struct {
	Age   string      `json:"age,omitempty"`
	Slots []boardSlot `json:"slots,omitempty"`
	// Made counts every offer ever made on this board; SinceRare the ordinary
	// offers made since the last Legendary one.
	Made      int `json:"made,omitempty"`
	SinceRare int `json:"since_rare,omitempty"`
}

// ready reports whether the board is made for age.
func (b *MissionBoard) ready(age string) bool {
	return b.Age == age && len(b.Slots) == boardSize
}

// rollRare decides whether the next offer is Legendary and moves the
// counters. The chance is always drawn, so the stream does not depend on the
// rules. Never two Legendary offers within config.RareOfferGap offers of
// each other (at least that many ordinary ones between), and the next after
// config.RareOfferFloor ordinary ones in a row is one. A board that has had
// no Legendary offer yet is held to the floor only.
func (b *MissionBoard) rollRare(rng *rand.Rand) bool {
	roll := rng.Float64()
	seenRare := b.SinceRare < b.Made
	rare := b.SinceRare >= config.RareOfferFloor ||
		((!seenRare || b.SinceRare >= config.RareOfferGap) && roll < config.RareOfferChance)
	b.Made++
	if rare {
		b.SinceRare = 0
	} else {
		b.SinceRare++
	}
	return rare
}

// makeOffer rolls one offer of a tier for age, made at tick.
func (b *MissionBoard) makeOffer(rng *rand.Rand, tier, age string, tick int) *MissionOffer {
	words, ok := config.MissionWordsFor(age)
	if !ok {
		return nil
	}
	offer := &MissionOffer{Tier: tier, Age: age, ExpiresTick: tick + minutesToTicks(config.BoardExpireMinutes)}
	name := words.Tiers[config.WordsShort]
	objectives := words.Objectives[config.WordsShort]
	switch tier {
	case TierLong:
		name, objectives = words.Tiers[config.WordsLong], words.Objectives[config.WordsLong]
	case TierGrand:
		offer.Grand = GrandMilitary
		name, objectives = words.Tiers[config.WordsGrandMilitary], words.Objectives[config.WordsGrandMilitary]
		if rng.Float64() < .5 {
			offer.Grand = GrandCivil
			name, objectives = words.Tiers[config.WordsGrandCivil], words.Objectives[config.WordsGrandCivil]
		}
	}
	offer.Rare = b.rollRare(rng)
	objective := objectives[rng.Intn(len(objectives))]
	place := words.Places[rng.Intn(len(words.Places))]
	offer.Description = words.Twists[rng.Intn(len(words.Twists))]
	offer.Title = fmt.Sprintf("%s: %s %s", name, objective, place)

	tuning, _ := config.MissionKindTuning(offer.Kind())
	base := tuning.BaseSuccess
	if tier == TierGrand && offer.Grand == GrandCivil {
		base = config.CivilGrandSuccess
	}
	length := tuning.MinMinutes + rng.Float64()*(tuning.MaxMinutes-tuning.MinMinutes)
	success := base + (rng.Float64()*0.2 - 0.1)
	success = math.Max(config.MissionSuccessMin, math.Min(config.MissionSuccessMax, success))
	haul := length*tuning.Yield*0.6 + length*tuning.Yield*rng.Float64()
	if offer.Rare {
		length *= config.RareLengthFactor
		success = math.Max(config.RareSuccessFloor, success-config.RareSuccessCut)
		haul *= config.RareHaulFactor
		offer.Title = config.RareTitleWord + " " + offer.Title
	}
	offer.LengthMinutes, offer.SuccessChance, offer.HaulMinutes = length, success, haul
	return offer
}

// regenerate throws the board away and makes nine new offers for age, in
// slot order. Active missions are not touched.
func (b *MissionBoard) regenerate(rng *rand.Rand, age string, tick int) {
	b.Age = age
	b.Slots = make([]boardSlot, boardSize)
	for i := range b.Slots {
		b.Slots[i].Offer = b.makeOffer(rng, slotTier(i), age, tick)
	}
}

// refresh replaces the offers that expired, at once, and fills the slots
// whose wait for a new offer is over, in slot order.
func (b *MissionBoard) refresh(rng *rand.Rand, tick int) {
	for i := range b.Slots {
		slot := &b.Slots[i]
		switch {
		case slot.Offer != nil && tick >= slot.Offer.ExpiresTick:
			slot.Offer = b.makeOffer(rng, slotTier(i), b.Age, tick)
		case slot.Offer == nil && tick >= slot.RefillTick:
			slot.Offer = b.makeOffer(rng, slotTier(i), b.Age, tick)
		}
	}
}

// take removes offer n (1-based) from the board and sets its slot to be
// filled again after config.BoardRefillMinutes. It returns the offer, or why
// there is none.
func (b *MissionBoard) take(n, tick int) (*MissionOffer, error) {
	slot, err := b.slot(n)
	if err != nil {
		return nil, err
	}
	offer := slot.Offer
	slot.Offer = nil
	slot.RefillTick = tick + minutesToTicks(config.BoardRefillMinutes)
	return offer, nil
}

// slot is board slot n (1-based), or why there is no such offer.
func (b *MissionBoard) slot(n int) (*boardSlot, error) {
	if n < 1 || n > len(b.Slots) {
		return nil, fmt.Errorf("there is no offer %d. The board has offers 1 to %d. Type expedition to see them", n, boardSize)
	}
	slot := &b.Slots[n-1]
	if slot.Offer == nil {
		return nil, fmt.Errorf("offer %d has been taken. A new one comes in a few minutes", n)
	}
	return slot, nil
}

// peek is offer n (1-based) without taking it.
func (b *MissionBoard) peek(n int) (*MissionOffer, error) {
	slot, err := b.slot(n)
	if err != nil {
		return nil, err
	}
	return slot.Offer, nil
}

// clone is a deep copy of the board, for saving.
func (b *MissionBoard) clone() MissionBoard {
	out := *b
	out.Slots = make([]boardSlot, len(b.Slots))
	for i, slot := range b.Slots {
		out.Slots[i] = slot
		if slot.Offer != nil {
			offer := *slot.Offer
			out.Slots[i].Offer = &offer
		}
	}
	return out
}

// EnsureBoard makes the board for age when it is not made yet, or is made
// for another age: a new age gets a whole new board. Under the engine's
// write lock.
func (mm *MilitaryManager) EnsureBoard(rng *rand.Rand, age string, tick int) {
	if !mm.board.ready(age) {
		mm.board.regenerate(rng, age, tick)
	}
}

// RefreshBoard expires old offers and fills the slots that are due. Call
// EnsureBoard first.
func (mm *MilitaryManager) RefreshBoard(rng *rand.Rand, tick int) {
	mm.board.refresh(rng, tick)
}

// BoardOffer is offer n (1-based) on the board, or why there is none.
func (mm *MilitaryManager) BoardOffer(n int) (MissionOffer, error) {
	offer, err := mm.board.peek(n)
	if err != nil {
		return MissionOffer{}, err
	}
	return *offer, nil
}

// armyTerm is what the army adds to a military grand mission's odds: 0.10 for
// each unit of mission power over the age's typical power, within 15 points
// either way. power is the army on the age's yardstick (missionPower).
func armyTerm(power, typical float64) float64 {
	if typical <= 0 {
		return 0
	}
	return math.Max(-0.15, math.Min(0.15, 0.10*(power/typical-1)))
}
