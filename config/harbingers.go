package config

import "sync"

// harbingers.go — the per-age roster of harbingers: the figure who turns up
// some while before a fated doom strikes and says how worried to be (and, from
// the Classical Age on, roughly when). This file is data only. The mechanic
// (the hidden fate roll, when a harbinger appears, what Appease / Brace /
// Invite cost and do, how false prophets come) lives on the game side
// (game/fate.go, game/harbinger.go), and the prose lives in the flavor
// package, whose harbinger Moments take Name as Request.Subject.

// ForecastPrecision is how much a harbinger's medium can tell you.
type ForecastPrecision int

const (
	// ForecastVague is omens, cries and charts: the warning's tone is the only
	// gauge of the risk. Every age before the Industrial Age.
	ForecastVague ForecastPrecision = iota
	// ForecastNumeric is a medium that publishes the odds: the paper prints
	// them, the wire carries them, the screen shows them. The UI shows the
	// number; flavor lines refer to it and never print it.
	ForecastNumeric
)

// ForecastTiming is how much a harbinger can say about WHEN the doom strikes.
type ForecastTiming int

const (
	// TimingNone is a figure who gives no timing at all: doom is coming, and
	// that is everything. Every age before the Classical Age.
	TimingNone ForecastTiming = iota
	// TimingAge is a figure who can tell whether the doom falls in the current
	// age ("before this age is out") or later in the era ("before the era
	// ends"). From the Classical Age on.
	TimingAge
)

// HarbingerDef is one age's harbinger.
type HarbingerDef struct {
	// Age is the config age key this harbinger belongs to.
	Age string
	// Key is a stable snake_case identifier, unique across the roster.
	Key string
	// Name is the display name, a singular noun phrase that opens lowercase
	// ("the Town Crier", "your future self") so it reads mid-sentence. The
	// flavor package drops it into sentences as-is.
	Name string
	// Description is a one-line, player-facing summary of who this is.
	Description string
	// AppeaseLabel names the action that spends faith and culture to lower the
	// odds, in the age's own terms.
	AppeaseLabel string
	// BraceLabel names the action that spends resources to soften an Endure.
	BraceLabel string
	// InviteLabel names the action that chooses doom: the next transition is
	// guaranteed to bring the catastrophe. For players who want to Succumb on
	// purpose for the legacy bonuses. Deliberately a little unhinged.
	InviteLabel string
	// ForecastPrecision is vague before the Industrial Age and numeric from it.
	ForecastPrecision ForecastPrecision
	// ForecastTiming is TimingNone before the Classical Age and TimingAge
	// from it. See timingFor.
	ForecastTiming ForecastTiming
	// FalseProphetChance is the chance that an era entered at this age with no
	// doom fated gets a false prophet anyway: a warning that reads exactly like
	// a real one, revealed once its foretold window passes without a doom.
	// Only an era's first age rolls it. See falseProphetChance for the curve.
	FalseProphetChance float64
}

// falseProphetChance is the curve behind HarbingerDef.FalseProphetChance: a
// straight line from 1/8 at the Primitive Age (index 0) down to exactly zero
// at the Industrial Age (index 8), in steps of 1/64.
//
// Why a straight line, and why it ends where it does. The chance of being lied
// to should fall as the settlement's ways of checking a claim improve, and
// the one step change in those ways is the Industrial Age, when the medium
// starts printing the odds: a printed figure can be wrong, but it cannot be a
// fraud the way a wild man's vision can. So zero lands there exactly, which is
// also where ForecastPrecision flips, and the two can never disagree about
// whether a warning is checkable. Between the endpoints nothing justifies a
// curve over a line, and a line keeps each age's number legible: 1/64 per age,
// exactly representable, so tests compare it with == rather than a tolerance.
//
// 1/8 at the start is a cap, not a target. Only an era's first age rolls it,
// and only when nothing is fated there: one run in eight meets a hollow
// warning in the Stone Era (where nothing can be fated), fewer in the Iron and
// Steel Eras. Enough that a player learns to doubt the Wild Man, rare enough
// that Appease is still usually money well spent. Colonial ends at 1/64, the
// "trending to zero" the Industrial cut-off finishes.
func falseProphetChance(ageIndex int) float64 {
	const industrialIndex = 8
	if ageIndex >= industrialIndex {
		return 0
	}
	return float64(industrialIndex-ageIndex) / 64
}

// precisionFor is ForecastNumeric from the Industrial Age (index 8) on.
func precisionFor(ageIndex int) ForecastPrecision {
	if ageIndex >= 8 {
		return ForecastNumeric
	}
	return ForecastVague
}

// timingFor is TimingAge from the Classical Age (index 4) on.
//
// Why there. The first era that can be fated is the Iron Era, and its opening
// figure, the Desert Prophet, still gives no timing, so the first doom a new
// player meets is as mysterious as the design wants it. The Oracle is the
// first figure whose whole trade is prophecy, so she is the first to name a
// time. Everything before her (the Wild Man, the Hermit, the Soothsayer, who
// only ever come as false prophets now, and the Desert Prophet) says doom is
// coming and nothing about when.
func timingFor(ageIndex int) ForecastTiming {
	if ageIndex >= 4 {
		return TimingAge
	}
	return TimingNone
}

// harbingerRoster is the hand-written part of each entry, in age order. The
// derived fields (precision, false-prophet chance) are filled in by
// buildHarbingers from the age's position, so they cannot drift from the curve.
var harbingerRoster = []HarbingerDef{
	{
		Age: "primitive_age", Key: "wild_man", Name: "the Wild Man",
		Description:  "A man who lives past the last fire walks in from the wilderness, gray with ash, to say what he has seen.",
		AppeaseLabel: "Leave offerings at the stones",
		BraceLabel:   "Dig in and hoard",
		InviteLabel:  "Howl with the Wild Man",
	},
	{
		Age: "stone_age", Key: "hermit", Name: "the Hermit",
		Description:  "Comes down from the high caves once in a generation, and never with good news.",
		AppeaseLabel: "Carve his sign on the cave wall",
		BraceLabel:   "Wall up the cave mouth",
		InviteLabel:  "Climb the high rock and shout back",
	},
	{
		Age: "bronze_age", Key: "soothsayer", Name: "the Soothsayer",
		Description:  "Reads the future in knucklebones, sparrows and goat livers, and wants paying before and after.",
		AppeaseLabel: "Sacrifice a goat at the altar",
		BraceLabel:   "Seal the grain in jars",
		InviteLabel:  "Pour the omen-wine on the ground",
	},
	{
		Age: "iron_age", Key: "desert_prophet", Name: "the Desert Prophet",
		Description:  "Walks in from the dry country with sand in his beard and one message for the city.",
		AppeaseLabel: "Fast and wear sackcloth",
		BraceLabel:   "Raise the walls a course higher",
		InviteLabel:  "Curse the city alongside him",
	},
	{
		Age: "classical_age", Key: "oracle", Name: "the Oracle",
		Description:  "Speaks from the smoke over the cleft rock, through priests who charge by the question.",
		AppeaseLabel: "Send rich gifts to the shrine",
		BraceLabel:   "Provision the citadel",
		InviteLabel:  "Ask the Oracle for the worst",
	},
	{
		Age: "medieval_age", Key: "town_crier", Name: "the Town Crier",
		Description:  "Rings his bell at the market cross and reads out doom in the voice he uses for tolls.",
		AppeaseLabel: "Pay the monks to pray",
		BraceLabel:   "Shore up the walls",
		InviteLabel:  "Ring the bells backwards",
	},
	{
		Age: "renaissance_age", Key: "court_astrologer", Name: "the Court Astrologer",
		Description:  "Casts the prince's horoscope, and lately the prince's horoscope has been bad for everyone.",
		AppeaseLabel: "Commission a votive altarpiece",
		BraceLabel:   "Lay in stores behind the bastions",
		InviteLabel:  "Have the astrologer cast for ruin",
	},
	{
		Age: "colonial_age", Key: "pamphleteer", Name: "the Pamphleteer",
		Description:  "Prints doom on cheap paper and sells it outside the coffee house for a penny.",
		AppeaseLabel: "Proclaim a day of fasting and prayer",
		BraceLabel:   "Stockpile powder and flour",
		InviteLabel:  "Pay her to print the date",
	},
	{
		Age: "industrial_age", Key: "newsboy", Name: "the Newsboy",
		Description:  "Shouts the late edition from the corner, and the late edition has the odds printed on it.",
		AppeaseLabel: "Sponsor a revival meeting",
		BraceLabel:   "Reinforce the mills",
		InviteLabel:  "Buy every copy and print the rest yourself",
	},
	{
		Age: "victorian_age", Key: "soapbox_doomsayer", Name: "the Doomsayer",
		Description:  "Stands on a soap crate by the park railings with a sandwich board and a table of figures.",
		AppeaseLabel: "Hold a national day of prayer",
		BraceLabel:   "Lay in tinned goods and sandbags",
		InviteLabel:  "Climb onto the soapbox with him",
	},
	{
		Age: "electric_age", Key: "telegraph", Name: "the Telegraph",
		Description:  "Chatters all night at the post office with dispatches from stations that have stopped answering.",
		AppeaseLabel: "Fund a mass revival tour",
		BraceLabel:   "Wire the city for emergency power",
		InviteLabel:  "Wire back SEND IT",
	},
	{
		Age: "atomic_age", Key: "civil_defence_broadcast", Name: "the Civil Defense Broadcast",
		Description:  "Three long notes on every wireless, then a calm voice reading the odds from a card.",
		AppeaseLabel: "Fund the early-warning network",
		BraceLabel:   "Stock the fallout shelters",
		InviteLabel:  "Stand on the roof and wave at the sky",
	},
	{
		Age: "modern_age", Key: "evening_news", Name: "the Evening News",
		Description:  "Leads with it at six, with a graphic, an expert and an anchor trying not to look worried.",
		AppeaseLabel: "Run a national unity telethon",
		BraceLabel:   "Harden the power grid",
		InviteLabel:  "Go on air and dare it",
	},
	{
		Age: "information_age", Key: "chain_email", Name: "the Chain Email",
		Description:  "Forward this to ten people or it happens to you. It has a spreadsheet attached.",
		AppeaseLabel: "Forward it to everyone you know",
		BraceLabel:   "Back up everything to tape",
		InviteLabel:  "Reply all and say bring it on",
	},
	{
		Age: "digital_age", Key: "viral_video", Name: "the Viral Video",
		Description:  "Shaky, portrait, a million views by lunch, and the odds on a whiteboard at the end.",
		AppeaseLabel: "Fund a feel-good hashtag campaign",
		BraceLabel:   "Buy up the bottled water",
		InviteLabel:  "Duet the video, smiling",
	},
	{
		Age: "cyberpunk_age", Key: "net_ghost", Name: "the Ghost in the Net",
		Description:  "A dead corporate AI that leaks internal risk memos through the net, glitching on every third word.",
		AppeaseLabel: "Bribe the oracle AI",
		BraceLabel:   "Air-gap the grid",
		InviteLabel:  "Join the accelerationist collective",
	},
	{
		Age: "fusion_age", Key: "reactor_warden", Name: "the Reactor Warden",
		Description:  "The plant's safety intelligence, which has never before spoken outside a scheduled drill.",
		AppeaseLabel: "Hold a vigil in the containment hall",
		BraceLabel:   "Shunt power to the containment fields",
		InviteLabel:  "Tell the Warden to stop holding back",
	},
	{
		Age: "space_age", Key: "deep_space_monitor", Name: "the Deep Space Monitor",
		Description:  "A station behind the moon that has watched one patch of sky for forty years, and has just marked a packet urgent.",
		AppeaseLabel: "Broadcast a peace hymn into the dark",
		BraceLabel:   "Harden the orbital shields",
		InviteLabel:  "Point every dish at it and wave",
	},
	{
		Age: "interstellar_age", Key: "distress_beacon", Name: "the Distress Beacon",
		Description:  "Still looping from a colony that went silent eighty years ago, and the loop has changed.",
		AppeaseLabel: "Hold a vigil for the lost colony",
		BraceLabel:   "Pull the outposts back behind the shield",
		InviteLabel:  "Answer the beacon and ask for it",
	},
	{
		Age: "galactic_age", Key: "elder_relay", Name: "the Elder Relay",
		Description:  "An alien relay older than the species that found it, speaking in geometry for the first time in an age.",
		AppeaseLabel: "Answer the relay in the old tongue",
		BraceLabel:   "Fold the fleet into the nebula",
		InviteLabel:  "Tell the relay to send them",
	},
	{
		Age: "quantum_age", Key: "future_self", Name: "your future self",
		Description:  "A message in your handwriting, stamped nine years from now, that knows your passcode.",
		AppeaseLabel: "Keep the promises you made yourself",
		BraceLabel:   "Branch the timeline and hedge",
		InviteLabel:  "Tell your future self to go ahead",
	},
	{
		Age: "transcendent_age", Key: "unmade_self", Name: "your unmade self",
		Description:  "A version of you from a branch that ended, come to see whether this one ends the same way.",
		AppeaseLabel: "Mourn the branches that ended",
		BraceLabel:   "Anchor yourself to this reality",
		InviteLabel:  "Let the other you in",
	},
}

// harbingerTables is the roster, built once. Harbingers and HarbingerFor are
// called from the transition path and from the UI; neither should rebuild the
// age table or the index on every call (see the tick-path rebuild fix in
// PR #108), and the roster is immutable after init anyway.
var harbingerTables = sync.OnceValue(buildHarbingers)

type harbingerIndex struct {
	list  []HarbingerDef
	byAge map[string]HarbingerDef
}

// buildHarbingers fills the derived fields from each age's position in
// AgeOrder. A roster entry for an age config does not know is a programming
// error and panics at first use rather than shipping a harbinger that can
// never appear.
func buildHarbingers() harbingerIndex {
	pos := make(map[string]int)
	for i, k := range AgeOrder() {
		pos[k] = i
	}
	idx := harbingerIndex{
		list:  make([]HarbingerDef, 0, len(harbingerRoster)),
		byAge: make(map[string]HarbingerDef, len(harbingerRoster)),
	}
	for _, h := range harbingerRoster {
		i, ok := pos[h.Age]
		if !ok {
			panic("config: harbinger roster names unknown age " + h.Age)
		}
		h.ForecastPrecision = precisionFor(i)
		h.ForecastTiming = timingFor(i)
		h.FalseProphetChance = falseProphetChance(i)
		idx.list = append(idx.list, h)
		idx.byAge[h.Age] = h
	}
	return idx
}

// Harbingers returns the full roster in age order. The slice is a fresh copy;
// callers may modify it.
func Harbingers() []HarbingerDef {
	src := harbingerTables().list
	out := make([]HarbingerDef, len(src))
	copy(out, src)
	return out
}

// HarbingerFor returns the harbinger for an age key, and false for an unknown
// age. HarbingerDef holds only values, so the returned copy is safe to keep.
func HarbingerFor(age string) (HarbingerDef, bool) {
	h, ok := harbingerTables().byAge[age]
	return h, ok
}
