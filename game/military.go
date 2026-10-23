package game

import (
	"fmt"
	"maps"
	"math"
	"math/rand"
	"slices"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// Expedition categories. Every mission costs resources, but only a military
// grand mission costs soldiers and counts the army: that one is
// ExpeditionMilitary, and short, long and civil grand missions are
// ExpeditionScouting. The category picks the techs' terms on a mission and the
// odds of an encounter when it comes back.
const (
	ExpeditionScouting = "scouting"
	ExpeditionMilitary = "military"
)

// failedLootShare is the share of an expedition's or campaign's loot that
// comes back when it fails.
const failedLootShare = 0.3

// ActiveExpedition is a mission under way: a copy of the offer it was sent
// from, what it cost, and how long is left.
type ActiveExpedition struct {
	MissionOffer
	// Soldiers is the party marching now (fewer than Party after an ambush);
	// Party the soldiers it set out with. Both are away from their posts.
	Soldiers   int `json:"soldiers,omitempty"`
	Party      int `json:"party,omitempty"`
	LaunchTick int `json:"launch_tick,omitempty"`
	TicksLeft  int `json:"ticks_left"`
	// TotalTicks is the mission's whole length in ticks when it was sent.
	TotalTicks     int                `json:"total_ticks,omitempty"`
	ElapsedTicks   int                `json:"elapsed_ticks,omitempty"`
	BaseRewards    map[string]float64 `json:"base_rewards,omitempty"`
	SuccessRewards map[string]float64 `json:"success_rewards,omitempty"`
	FailureRewards map[string]float64 `json:"failure_rewards,omitempty"`
	Cost           map[string]float64 `json:"cost,omitempty"`
	RoadEvents     []RoadEvent        `json:"road_events,omitempty"`
	// RoadEventsSeen is the stories of the road events that fired, as read
	// now. A save holds them as RoadTold (the same stories with the words of
	// the script left to fill) and leaves this out, unless the mission was
	// written by hand and has no RoadTold.
	RoadEventsSeen []string   `json:"road_events_seen,omitempty"`
	RoadTold       []RoadTold `json:"road_told,omitempty"`
	// LuckSeed seeds the mission's own random stream, drawn from the game's
	// RNG at launch; LuckDraws is how far that stream has been drawn, so a
	// load resumes it where it stood. Luck is the first draw: the success
	// roll, made against the odds the mission has at its resolution.
	LuckSeed  int64   `json:"luck_seed,omitempty"`
	LuckDraws uint64  `json:"luck_draws,omitempty"`
	Luck      float64 `json:"luck,omitempty"`
	// LaunchChance is the odds the mission was sent with: SuccessChance moves
	// with the road, this does not. HaulScale is what the road did to the
	// loot, the product of its scalings (0 is none). Both are read by the
	// Expedition Report's field report.
	LaunchChance float64 `json:"launch_chance,omitempty"`
	HaulScale    float64 `json:"haul_scale,omitempty"`
	luckSrc      *countingSource
	luckRNG      *rand.Rand
}

// luckStream is the mission's own random stream. It is built on first use from
// LuckSeed, skipped forward to LuckDraws after a load, so every draw about the
// mission (the encounter on return, a ruins offer) comes from its own stream
// and not from the game's shared one, wherever in the run the mission is
// processed.
func (a *ActiveExpedition) luckStream() *rand.Rand {
	if a.luckRNG == nil {
		a.luckSrc = newCountingSource(a.LuckSeed)
		a.luckSrc.skip(min(a.LuckDraws, maxRNGReplay))
		a.luckRNG = rand.New(a.luckSrc)
	}
	return a.luckRNG
}

// luckDraws is how far the mission's stream has been drawn, for a save.
func (a *ActiveExpedition) luckDraws() uint64 {
	if a.luckSrc != nil {
		return a.luckSrc.draws
	}
	return a.LuckDraws
}

// withLuck runs fn with the mission's own stream standing in for the game's
// seeded RNG, so the draws fn makes (the encounter, a boon, a ruins offer) are
// the mission's. Call under the write lock.
func (ge *GameEngine) withLuck(rng *rand.Rand, fn func()) {
	saved := ge.rng
	ge.rng = rng
	defer func() { ge.rng = saved }()
	fn()
}

// RoadEvent is a deterministic launch-time event, ordered by mission tick.
type RoadEvent struct {
	Tick         int    `json:"tick"`
	Kind         string `json:"kind"`
	Story        string `json:"story,omitempty"`
	AuxStory     string `json:"aux_story,omitempty"`
	AmbushStory  string `json:"ambush_story,omitempty"`
	FoeIndex     int    `json:"foe_index,omitempty"`
	HelperIndex  int    `json:"helper_index,omitempty"`
	WeatherIndex int    `json:"weather_index,omitempty"`
	FindIndex    int    `json:"find_index,omitempty"`
	Fired        bool   `json:"fired,omitempty"`
}

// MilitaryManager handles the mission board, the missions under way and the
// defense rating.
//
// Any number of missions may be under way, of any tier: the army is the only
// limit (every mission takes a party of soldiers out of the army at home, and
// the engine checks that and the other costs before LaunchOffer is called).
// active holds them in the order they were sent.
type MilitaryManager struct {
	rules          *rules.Set
	active         []*ActiveExpedition
	board          MissionBoard
	completedCount int
	// experience is the army's experience (army.go).
	experience float64
	// minTiers keeps that many tiers of the board open whatever the era and
	// the rank say. Only tests set it, to plant an offer of a tier the era
	// has not opened.
	minTiers      int
	totalLoot     map[string]float64
	defenseRating float64
	// returns is the returns log, newest first (mission_returns.go);
	// returnsMade counts every mission that has come home this run and
	// returnsSeen how many of those the Expedition Report has shown.
	returns                  []MissionReturn
	returnsMade, returnsSeen int
	// scoutTime is the techs' term on the time a short, long or civil
	// mission takes (config.MechanicExpeditionTicks). The engine sets it
	// (SetScoutTime); 0 reads as 1. Military grand missions keep their time.
	scoutTime float64
	// campaignTime and campaignPay are the techs' terms on the time a
	// military grand mission takes and on what it brings back
	// (config.MechanicCampaignTicks, config.MechanicCampaignReward). The
	// engine sets them (SetCampaignTerms); 0 reads as 1.
	campaignTime, campaignPay float64
}

// SetScoutTime sets the techs' term on a short, long or civil mission's time.
func (mm *MilitaryManager) SetScoutTime(f float64) { mm.scoutTime = f }

// SetCampaignTerms sets the techs' terms on a military grand mission's time
// and on what it brings back.
func (mm *MilitaryManager) SetCampaignTerms(time, pay float64) {
	mm.campaignTime, mm.campaignPay = time, pay
}

// missionTicks is ticks with the techs' cut for a category.
func (mm *MilitaryManager) missionTicks(category string, ticks int) int {
	if category == ExpeditionScouting {
		return techTimeTicks(ticks, mm.scoutTime)
	}
	return techTimeTicks(ticks, mm.campaignTime)
}

// CampaignPay is amount of a campaign's loot with the techs' share on top:
// amount itself with none.
func (mm *MilitaryManager) CampaignPay(amount float64) float64 {
	if mm.campaignPay <= 1 {
		return amount
	}
	return float64(amount * mm.campaignPay)
}

// NewMilitaryManager creates a military manager on the core ruleset.
func NewMilitaryManager() *MilitaryManager { return NewMilitaryManagerWith(rules.Core()) }

// Rebind moves the manager onto set, whose clocks time its missions.
func (mm *MilitaryManager) Rebind(set *rules.Set) {
	mm.rules = set
	mm.redraw()
}

// redraw writes again every text the manager holds that was drawn from the
// words of a script: the offers on the board, the missions under way (their
// titles, places and road stories) and the returns log. Their words are held
// by index, so the texts follow the script in force; the numbers are not
// touched. It runs when the manager moves onto a ruleset and when a save is
// loaded.
func (mm *MilitaryManager) redraw() {
	if mm.rules == nil {
		return
	}
	for i := range mm.board.Slots {
		if o := mm.board.Slots[i].Offer; o != nil {
			if words, ok := mm.rules.MissionWords(o.Age); ok {
				o.draw(words)
			}
		}
	}
	for _, a := range mm.active {
		words, ok := mm.rules.MissionWords(a.Age)
		if !ok {
			continue
		}
		a.MissionOffer.draw(words)
		if len(a.RoadTold) > 0 {
			a.RoadEventsSeen = a.RoadEventsSeen[:0:0]
			for _, rt := range a.RoadTold {
				if rt.Event >= 0 && rt.Event < len(a.RoadEvents) {
					a.RoadEventsSeen = append(a.RoadEventsSeen, tellRoadStory(words, a.roadPlace(), a.RoadEvents[rt.Event], rt.Text))
				}
			}
		}
	}
	for i := range mm.returns {
		mm.returns[i].draw(mm)
	}
}

// NewMilitaryManagerWith creates a military manager on set.
func NewMilitaryManagerWith(set *rules.Set) *MilitaryManager {
	return &MilitaryManager{
		rules:     set,
		totalLoot: make(map[string]float64),
	}
}

// LaunchOffer puts a mission under way, with its party of soldiers, and
// returns it. The offer's SuccessChance is the chance the mission is sent with
// (the engine has already counted the army). The party and the resource costs
// are validated and deducted by the engine BEFORE this is called.
func (mm *MilitaryManager) LaunchOffer(offer MissionOffer, tick, soldiers int, cost, baseHaul map[string]float64) (*ActiveExpedition, error) {
	ticks := mm.missionTicks(offer.Category(), minutesToTicks(offer.LengthMinutes))
	if ticks < 1 {
		ticks = 1
	}
	a := &ActiveExpedition{
		MissionOffer: offer,
		Soldiers:     soldiers,
		Party:        soldiers,
		LaunchTick:   tick,
		TicksLeft:    ticks,
		TotalTicks:   ticks,
		BaseRewards:  maps.Clone(baseHaul),
		Cost:         maps.Clone(cost),
		LaunchChance: offer.SuccessChance,
	}
	mm.active = append(mm.active, a)
	return a, nil
}

// ActiveByTier returns the oldest mission under way in a tier, or nil. The
// returned pointer is the live manager state.
func (mm *MilitaryManager) ActiveByTier(tier string) *ActiveExpedition {
	for _, a := range mm.active {
		if a.Tier == tier {
			return a
		}
	}
	return nil
}

// ActiveOf returns every mission under way in a tier, oldest first.
func (mm *MilitaryManager) ActiveOf(tier string) []*ActiveExpedition {
	var out []*ActiveExpedition
	for _, a := range mm.active {
		if a.Tier == tier {
			out = append(out, a)
		}
	}
	return out
}

// ActiveAll returns every mission under way, in the order they were sent.
func (mm *MilitaryManager) ActiveAll() []*ActiveExpedition { return slices.Clone(mm.active) }

// HasActive reports whether any mission is currently under way.
func (mm *MilitaryManager) HasActive() bool { return len(mm.active) > 0 }

// SoldiersAway is the soldiers out on missions: the parties marching.
func (mm *MilitaryManager) SoldiersAway() int {
	away := 0
	for _, a := range mm.active {
		away += a.Soldiers
	}
	return away
}

// ExpeditionResult holds the rewards + player-facing message for one mission
// that resolved this tick. Category and Success describe HOW it resolved so the
// engine can weight a faction encounter on resolution (see
// GameEngine.rollExpeditionEncounter).
//
// Name identifies WHICH mission resolved. The varied flavour line is generated
// ENGINE-side (MilitaryManager has no rng of its own for prose; ge.rng does,
// and the prose stream must come off the same seeded source as everything
// else) — see GameEngine.expeditionFlavorLine. Message stays the mechanical
// line; flavour rides alongside it, never in place of it.
type ExpeditionResult struct {
	Rewards      map[string]float64
	Message      string
	Name         string // the offer's title
	Category     string // ExpeditionScouting or ExpeditionMilitary
	Success      bool   // true if the mission succeeded
	SoldiersSent int
	// SoldiersFell is the soldiers who fell on the way home (the failure's or
	// the success's losses, drawn from the mission's own stream); Party the
	// party that set out and Marching the soldiers still with it on return.
	SoldiersFell, Party, Marching int
	// Offer is the offer the mission was sent from: its tier, age and
	// Legendary mark are what the badges read.
	Offer MissionOffer
	// Experience is what the army's experience really changed by when the
	// mission came home (negative on a loss, 0 at the floor); Rank and
	// RankBefore are the army's rank index after and before.
	Experience       float64
	Rank, RankBefore int
	RoadEvents       int
	RoadStories      []string
	// luck is the mission's own stream, for the encounter on return.
	luck *rand.Rand
}

// Party losses. A failure costs a share of the party that comes home, a
// success only sometimes; a Legendary mission's failure costs more. Every draw
// is the mission's own.
const (
	lossFailLow, lossFailHigh         = 0.4, 0.7
	lossRareFailLow, lossRareFailHigh = 0.6, 0.9
	lossWinChance                     = 0.30
	lossWinMax                        = 0.15
)

// missionLosses is the soldiers of a returning party who fell: on a failure
// ceil(party x U(0.4, 0.7)) (Legendary U(0.6, 0.9)); on a success, with a 30%
// chance, ceil(party x U(0, 0.15)). It draws from rng a fixed number of times
// for the outcome: one on a failure, two on a success.
func missionLosses(rng *rand.Rand, party int, success, rare bool) int {
	if party <= 0 {
		return 0
	}
	share := 0.0
	if success {
		hit := rng.Float64() < lossWinChance
		amount := float64(lossWinMax * rng.Float64())
		if hit {
			share = amount
		}
	} else {
		lo, hi := lossFailLow, lossFailHigh
		if rare {
			lo, hi = lossRareFailLow, lossRareFailHigh
		}
		share = lo + float64((hi-lo)*rng.Float64())
	}
	return min(party, int(math.Ceil(float64(float64(party)*share))))
}

// Tick advances every mission under way by one tick and returns a result for
// each that resolved this tick, in the order they were sent. Missions tick
// down and resolve independently.
//
// The success roll is the mission's own SuccessChance (rolled when the offer
// was made; the army was counted when it was sent). expeditionBonus is not
// used here: the loot was sized at launch. A failure differs from a success in
// the loot (30%) and in the soldiers who fall (missionLosses).
//
// The roll is the mission's own draw from its luck stream, taken at launch
// (ActiveExpedition.Luck): it succeeds when that draw is under the odds it
// has now, so a road event that moves the odds still counts.
func (mm *MilitaryManager) Tick(missionPower, expeditionBonus float64) []ExpeditionResult {
	var results []ExpeditionResult
	kept := mm.active[:0:0]
	for _, active := range mm.active {
		if res, ok := mm.tickMission(active); ok {
			results = append(results, res)
			continue
		}
		kept = append(kept, active)
	}
	mm.active = kept
	return results
}

// tickMission advances one mission. ok is true only on the tick it resolves.
func (mm *MilitaryManager) tickMission(active *ActiveExpedition) (ExpeditionResult, bool) {
	active.TicksLeft--
	if active.TicksLeft > 0 {
		return ExpeditionResult{}, false
	}

	success := active.Luck < active.SuccessChance
	rewards := make(map[string]float64)
	var message string
	if success {
		for res, amount := range active.SuccessRewards {
			rewards[res] = amount
			mm.totalLoot[res] += amount
		}
		message = fmt.Sprintf("%s succeeded. Loot: %s.", active.Title, amountsText(rewards))
	} else {
		for res, amount := range active.FailureRewards {
			rewards[res] = amount
			mm.totalLoot[res] += amount
		}
		message = fmt.Sprintf("%s failed. You kept %.0f%% of the loot: %s.", active.Title, failedLootShare*100, amountsText(rewards))
	}
	fell := missionLosses(active.luckStream(), active.Soldiers, success, active.Rare)
	gained, before, after := mm.gainExperience(config.ArmyExperienceFor(active.Kind(), success))

	mm.completedCount++
	return ExpeditionResult{
		Rewards:      rewards,
		Message:      message,
		Name:         active.Title,
		Category:     active.Category(),
		Success:      success,
		SoldiersSent: active.Soldiers,
		SoldiersFell: fell,
		Party:        active.Party,
		Marching:     active.Soldiers,
		Offer:        active.MissionOffer,
		Experience:   gained,
		Rank:         after,
		RankBefore:   before,
		RoadEvents:   len(active.RoadEventsSeen),
		RoadStories:  slices.Clone(active.RoadEventsSeen),
		luck:         active.luckStream(),
	}, true
}

// CalculateDefense calculates defense rating from soldiers and bonuses
func (mm *MilitaryManager) CalculateDefense(soldierCount int, militaryBonus float64) float64 {
	base := float64(soldierCount) * 2.0
	return base * (1.0 + militaryBonus)
}

// Snapshot returns military state for UI.
//
// soldierCount is the soldiers at home. The engine fills the board's offers (it knows the costs and the stores) and
// the automatic dispatch view.
func (mm *MilitaryManager) Snapshot(soldierCount int, militaryBonus, expeditionBonus float64) MilitaryState {
	snap := func(a *ActiveExpedition) *ExpeditionSnapshot {
		out := snapshotActive(a, expeditionBonus)
		if out != nil {
			out.Label = mm.Label(a.MissionOffer)
		}
		return out
	}
	var all []*ExpeditionSnapshot
	for _, a := range mm.active {
		all = append(all, snap(a))
	}
	activeShort, activeLong, activeGrand := snap(mm.ActiveByTier(TierShort)), snap(mm.ActiveByTier(TierLong)), snap(mm.ActiveByTier(TierGrand))

	// ActiveScout and ActiveMilitary are the two slots the older panels read:
	// the short (else the long) mission, and the grand one.
	activeScout := activeShort
	if activeScout == nil {
		activeScout = activeLong
	}

	loot := make(map[string]float64)
	for k, v := range mm.totalLoot {
		loot[k] = v
	}

	return MilitaryState{
		SoldierCount:    soldierCount,
		DefenseRating:   mm.CalculateDefense(soldierCount, militaryBonus),
		MilitaryBonus:   militaryBonus,
		ExpeditionBonus: expeditionBonus,
		Active:          all,
		SoldiersAway:    mm.SoldiersAway(),
		ActiveShort:     activeShort,
		ActiveLong:      activeLong,
		ActiveGrand:     activeGrand,
		ActiveScout:     activeScout,
		ActiveMilitary:  activeGrand,
		CompletedCount:  mm.completedCount,
		Experience:      mm.experience,
		Rank:            mm.Rank(),
		TotalLoot:       loot,
		Returns:         mm.returnsCopy(),
		ReturnsMade:     mm.returnsMade,
		ReturnsSeen:     mm.returnsSeen,
	}
}

// snapshotActive converts a live ActiveExpedition into a UI snapshot, or nil.
func snapshotActive(active *ActiveExpedition, expeditionBonus float64) *ExpeditionSnapshot {
	if active == nil {
		return nil
	}
	success := maps.Clone(active.SuccessRewards)
	failure := maps.Clone(active.FailureRewards)
	if success == nil {
		success = scaleLoot(active.BaseRewards, 1+expeditionBonus)
	}
	if failure == nil {
		failure = scaleLoot(active.BaseRewards, failedLootShare)
	}
	total := active.TotalTicks
	if total < active.TicksLeft {
		total = max(active.TicksLeft, minutesToTicks(active.LengthMinutes))
	}
	return &ExpeditionSnapshot{
		Name:          active.Title,
		Place:         active.Place,
		ExpiresTick:   active.ExpiresTick,
		TotalTicks:    total,
		Tier:          active.Tier,
		Grand:         active.Grand,
		Rare:          active.Rare,
		Soldiers:      active.Soldiers,
		Party:         max(active.Party, active.Soldiers),
		TicksLeft:     active.TicksLeft,
		LengthMinutes: active.LengthMinutes,
		SuccessChance: active.SuccessChance,
		HaulMinutes:   active.HaulMinutes,
		SuccessHaul:   success,
		FailureHaul:   failure,
		RoadStories:   append([]string(nil), active.RoadEventsSeen...),
		RoadKinds:     firedRoadKinds(active),
		LaunchChance:  active.LaunchChance,
		HaulScale:     active.haulScale(),
	}
}

// firedRoadKinds is the kind of each road event that has fired, in the order
// RoadEventsSeen holds their stories.
func firedRoadKinds(active *ActiveExpedition) []string {
	var kinds []string
	for _, e := range active.RoadEvents {
		if e.Fired {
			kinds = append(kinds, e.Kind)
		}
	}
	return kinds
}

func scaleLoot(in map[string]float64, scale float64) map[string]float64 {
	out := make(map[string]float64, len(in))
	for key, value := range in {
		out[key] = value * scale
	}
	return out
}

// LoadState restores military state from save: the missions under way, the
// board and its counters, and the totals.
func (mm *MilitaryManager) LoadState(save MilitarySave) {
	mm.active = nil
	for _, entry := range save.Active {
		if entry.Mission == nil {
			continue
		}
		mission := *entry.Mission
		mission.luckSrc, mission.luckRNG = nil, nil
		mission.RoadEvents = append([]RoadEvent(nil), entry.Mission.RoadEvents...)
		mission.RoadEventsSeen = append([]string(nil), entry.Mission.RoadEventsSeen...)
		mission.RoadTold = append([]RoadTold(nil), entry.Mission.RoadTold...)
		mm.active = append(mm.active, &mission)
	}
	mm.board = MissionBoard{}
	if save.Board != nil {
		mm.board = save.Board.clone()
	}
	mm.completedCount = save.CompletedCount
	mm.experience = max(0, save.Experience)
	mm.returns = slices.Clone(save.Returns)
	for i := range mm.returns {
		mm.returns[i].Loot = maps.Clone(mm.returns[i].Loot)
	}
	mm.returnsMade, mm.returnsSeen = save.ReturnsMade, save.ReturnsSeen
	if save.TotalLoot != nil {
		mm.totalLoot = save.TotalLoot
	}
	mm.redraw()
}

// ActiveForSave returns deep copies of the missions under way, as a list
// keyed by slot (the tier), in the order they were sent.
func (mm *MilitaryManager) ActiveForSave() []ActiveSlotSave {
	var out []ActiveSlotSave
	for _, active := range mm.active {
		c := *active
		c.BaseRewards, c.SuccessRewards = maps.Clone(active.BaseRewards), maps.Clone(active.SuccessRewards)
		c.FailureRewards, c.Cost = maps.Clone(active.FailureRewards), maps.Clone(active.Cost)
		c.RoadEvents = append([]RoadEvent(nil), active.RoadEvents...)
		c.RoadEventsSeen = append([]string(nil), active.RoadEventsSeen...)
		c.RoadTold = append([]RoadTold(nil), active.RoadTold...)
		// A save holds the words a text was drawn from, not the text, so it
		// reads the same whichever script the game was read in (VibeMode).
		c.MissionOffer = c.MissionOffer.forSave()
		if len(c.RoadTold) > 0 {
			c.RoadEventsSeen = nil
		}
		// The copy starts its own stream from the seed and draw count, so it
		// never shares the live mission's stream.
		c.LuckDraws, c.luckSrc, c.luckRNG = active.luckDraws(), nil, nil
		out = append(out, ActiveSlotSave{Slot: active.Tier, Mission: &c})
	}
	return out
}

// BoardForSave returns a deep copy of the board, or nil before one is made.
func (mm *MilitaryManager) BoardForSave() *MissionBoard {
	if mm.board.Age == "" {
		return nil
	}
	c := mm.board.clone()
	for i := range c.Slots {
		if c.Slots[i].Offer != nil {
			offer := c.Slots[i].Offer.forSave()
			c.Slots[i].Offer = &offer
		}
	}
	return &c
}

// returnsCopy is the returns log, copied so that what it hands out shares
// nothing with the manager (the loot maps included).
func (mm *MilitaryManager) returnsCopy() []MissionReturn {
	out := slices.Clone(mm.returns)
	for i := range out {
		out[i].Loot = maps.Clone(out[i].Loot)
	}
	return out
}

// ReturnsForSave is the returns log as a save holds it.
func (mm *MilitaryManager) ReturnsForSave() []MissionReturn {
	out := mm.returnsCopy()
	for i := range out {
		out[i] = out[i].forSave()
	}
	return out
}
