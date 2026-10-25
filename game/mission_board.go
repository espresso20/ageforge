package game

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/vibe"
)

// The mission board: three offers of each open tier, rolled up front. The
// Stone Era's board holds the short tier alone, the Iron Era's adds the long
// tier and the Steel Era's the grand one; a Veteran army sees the next tier an
// era early (config.MissionTiersOpen). The slots are always the open tiers in
// order, so a board holds 3, 6 or 9 offers.
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

// boardSizeFor is how many offers a board with tiers tiers open holds.
func boardSizeFor(tiers int) int { return tiers * config.BoardOffersPerTier }

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
	Age   string `json:"age"`
	Title string `json:"title"`
	// Place is the place the title names: the Expeditions panel puts the
	// offer's pin there and labels it.
	Place       string `json:"place,omitempty"`
	Description string `json:"description"`
	// Words are the words the three texts above were drawn from, by place in
	// their age's lists. An offer the game made holds them, and a save holds
	// them in place of the texts, so what a save says does not depend on the
	// script the game was read in (VibeMode); draw writes the texts again
	// from the words of the script in force. An offer with no Words is
	// written by hand (a test's) and its texts are what they are.
	Words OfferWords `json:"words,omitzero"`
	// LengthMinutes, SuccessChance and HaulMinutes are the rolls. The haul is
	// in minutes of the town's income; the resources are worked out at launch.
	LengthMinutes float64 `json:"length_minutes"`
	SuccessChance float64 `json:"success_chance"`
	HaulMinutes   float64 `json:"haul_minutes"`
	// Rare marks a Legendary offer.
	Rare        bool `json:"rare,omitempty"`
	ExpiresTick int  `json:"expires_tick,omitempty"`
}

// OfferWords are the indices an offer's texts were drawn at in its age's
// MissionWords: the objective of its tier, the place and the twist.
type OfferWords struct {
	// Made is set on the words of an offer the game made.
	Made      bool `json:"made"`
	Objective int  `json:"objective"`
	Place     int  `json:"place"`
	Twist     int  `json:"twist"`
	// Found, when not 0, is 1 and the place of the mission that found this
	// offer by the road (a ruins offer): its description names that place.
	Found int `json:"found,omitempty"`
}

// ruinsFound is what a ruins offer's description says before the place.
func ruinsFound() string {
	return vibe.Line("game.mission_board.ruinsFound.01", "Found by this mission near the ruins at ")
}

// ruinsFoundPlain is ruinsFound for an offer written by hand: it has no words
// to draw from, so a save holds its description as it was written, in plain
// English whatever script the game is read in.
const ruinsFoundPlain = "Found by this mission near the ruins at "

// wordAt is list[i], with an index a hand-edited save may have put out of range
// brought back into it.
func wordAt(list []string, i int) string {
	if len(list) == 0 {
		return ""
	}
	return list[((i%len(list))+len(list))%len(list)]
}

// draw writes the offer's Title, Place and Description from its Words, in the
// words of the script in force. An offer with no Words is left as it is.
func (o *MissionOffer) draw(words config.MissionWords) {
	w := o.Words
	if !w.Made {
		return
	}
	which := o.wordsKind()
	objectives := words.ObjectivesFor(which)
	place := wordAt(words.Places[:], w.Place)
	o.Place = place
	o.Title = fmt.Sprintf("%s: %s %s", words.Tiers[which], wordAt(objectives[:], w.Objective), place)
	if o.Rare {
		o.Title = config.RareTitleWord() + " " + o.Title
	}
	o.Description = wordAt(words.Twists[:], w.Twist)
	if w.Found > 0 {
		o.Description = ruinsFound() + wordAt(words.Places[:], w.Found-1) + "."
	}
}

// forSave is the offer as a save holds it: the words it was drawn from and
// none of the texts drawn.
func (o MissionOffer) forSave() MissionOffer {
	if o.Words.Made {
		o.Title, o.Place, o.Description = "", "", ""
	}
	return o
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

// ready reports whether the board is made for age with tiers tiers open.
func (b *MissionBoard) ready(age string, tiers int) bool {
	return b.Age == age && len(b.Slots) == boardSizeFor(tiers)
}

// rollRare decides whether the next offer is Legendary and moves the
// counters. The chance is always drawn, so the stream does not depend on the
// rules. rareFactor is the army rank's factor on the chance
// (config.ArmyRareFactor). Never two Legendary offers within config.RareOfferGap offers of
// each other (at least that many ordinary ones between), and every run of
// config.RareOfferFloor offers holds one (so after RareOfferFloor-1 ordinary
// ones in a row the next is Legendary). A board that has had no Legendary
// offer yet is held to the floor only.
func (b *MissionBoard) rollRare(rng *rand.Rand, rareFactor float64) bool {
	roll := rng.Float64()
	seenRare := b.SinceRare < b.Made
	rare := b.SinceRare >= config.RareOfferFloor-1 ||
		((!seenRare || b.SinceRare >= config.RareOfferGap) && roll < config.RareOfferChance*rareFactor)
	b.Made++
	if rare {
		b.SinceRare = 0
	} else {
		b.SinceRare++
	}
	return rare
}

// makeOffer rolls one offer of a tier for age, made at tick.
func (b *MissionBoard) makeOffer(rng *rand.Rand, words config.MissionWords, tier string, tick int, rareFactor float64) *MissionOffer {
	age := words.Age
	offer := &MissionOffer{Tier: tier, Age: age, ExpiresTick: tick + minutesToTicks(config.BoardExpireMinutes)}
	which := config.WordsShort
	switch tier {
	case TierLong:
		which = config.WordsLong
	case TierGrand:
		offer.Grand, which = GrandMilitary, config.WordsGrandMilitary
		if rng.Float64() < .5 {
			offer.Grand, which = GrandCivil, config.WordsGrandCivil
		}
	}
	objectives := words.ObjectivesFor(which)
	offer.Rare = b.rollRare(rng, rareFactor)
	offer.Words = OfferWords{
		Made:      true,
		Objective: rng.Intn(len(objectives)),
		Place:     rng.Intn(len(words.Places)),
		Twist:     rng.Intn(len(words.Twists)),
	}
	offer.draw(words)

	tuning, _ := config.MissionKindTuning(offer.Kind())
	base := tuning.BaseSuccess
	if tier == TierGrand && offer.Grand == GrandCivil {
		base = config.CivilGrandSuccess
	}
	length := tuning.MinMinutes + float64(rng.Float64()*(tuning.MaxMinutes-tuning.MinMinutes))
	success := base + (float64(rng.Float64()*0.2) - 0.1)
	success = math.Max(config.MissionSuccessMin, math.Min(config.MissionSuccessMax, success))
	haul := float64(length*tuning.Yield*0.6) + float64(length*tuning.Yield*rng.Float64())
	if offer.Rare {
		length *= config.RareLengthFactor
		success = math.Max(config.RareSuccessFloor, success-config.RareSuccessCut)
		haul *= config.RareHaulFactor
	}
	offer.LengthMinutes, offer.SuccessChance, offer.HaulMinutes = length, success, haul
	return offer
}

// regenerate throws the board away and makes new offers for age, in slot
// order, for tiers open tiers. Active missions are not touched.
func (b *MissionBoard) regenerate(rng *rand.Rand, words config.MissionWords, tick, tiers int, rareFactor float64) {
	b.Age = words.Age
	b.Slots = make([]boardSlot, boardSizeFor(tiers))
	for i := range b.Slots {
		b.Slots[i].Offer = b.makeOffer(rng, words, slotTier(i), tick, rareFactor)
	}
}

// resize fits a board made for age to tiers open tiers: a tier that closed
// takes its offers with it, a tier that opened gets new offers, in slot order.
func (b *MissionBoard) resize(rng *rand.Rand, words config.MissionWords, tick, tiers int, rareFactor float64) {
	size := boardSizeFor(tiers)
	if len(b.Slots) > size {
		b.Slots = b.Slots[:size]
	}
	for len(b.Slots) < size {
		b.Slots = append(b.Slots, boardSlot{Offer: b.makeOffer(rng, words, slotTier(len(b.Slots)), tick, rareFactor)})
	}
}

// refresh replaces the offers that expired, at once, and fills the slots
// whose wait for a new offer is over, in slot order.
func (b *MissionBoard) refresh(rng *rand.Rand, words config.MissionWords, tick int, rareFactor float64) {
	for i := range b.Slots {
		slot := &b.Slots[i]
		switch {
		case slot.Offer != nil && tick >= slot.Offer.ExpiresTick:
			slot.Offer = b.makeOffer(rng, words, slotTier(i), tick, rareFactor)
		case slot.Offer == nil && tick >= slot.RefillTick:
			slot.Offer = b.makeOffer(rng, words, slotTier(i), tick, rareFactor)
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
		return nil, vibe.Errorf("game.mission_board.slot.01", "there is no offer %d. The board has offers 1 to %d. Type expedition to see them", n, len(b.Slots))
	}
	slot := &b.Slots[n-1]
	if slot.Offer == nil {
		return nil, vibe.Errorf("game.mission_board.slot.02", "offer %d has been taken. A new one comes in a few minutes", n)
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
// for another age: a new age gets a whole new board. A board made for age
// is fitted to the tiers open now (the era and the army's rank): a tier that
// opens gets its offers, one that closes loses them. Under the engine's
// write lock.
func (mm *MilitaryManager) EnsureBoard(rng *rand.Rand, age string, tick int) {
	tiers := mm.TiersOpen(age)
	if mm.board.ready(age, tiers) {
		return
	}
	words, ok := mm.rules.MissionWords(age)
	if !ok {
		return
	}
	if mm.board.Age == age && len(mm.board.Slots) > 0 {
		mm.board.resize(rng, words, tick, tiers, mm.rareFactor())
		return
	}
	mm.board.regenerate(rng, words, tick, tiers, mm.rareFactor())
}

// RefreshBoard expires old offers and fills the slots that are due. Call
// EnsureBoard first.
func (mm *MilitaryManager) RefreshBoard(rng *rand.Rand, tick int) {
	words, ok := mm.rules.MissionWords(mm.board.Age)
	if !ok {
		return
	}
	mm.board.refresh(rng, words, tick, mm.rareFactor())
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
