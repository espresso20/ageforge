package game

import (
	"math/rand"
	"strings"
)

// GenerateSaveName returns a procedural, human-friendly name for a new save
// slot. It picks one of three style families at random (epic / mythic /
// whimsical) and assembles a name from curated word banks.
//
// The result is always FILESYSTEM-SAFE: it becomes a save filename
// (<name>.json), so it contains letters and single spaces only — never an
// apostrophe, slash, dot, colon, quote, or other path-hostile character.
// Mythic names therefore use no apostrophes. Output is trimmed and has
// internal whitespace collapsed to single spaces.
//
// Names are cosmetic and MAY repeat across calls — there is no uniqueness
// guarantee. We use math/rand's global source (auto-seeded by modern Go),
// so callers do not need to seed anything.
func GenerateSaveName() string {
	switch rand.Intn(3) {
	case 0:
		return cleanSaveName(generateEpicName())
	case 1:
		return cleanSaveName(generateMythicName())
	default:
		return cleanSaveName(generateWhimsicalName())
	}
}

// cleanSaveName collapses internal whitespace to single spaces and trims the
// ends, guaranteeing the no-leading/trailing-space invariant the word-bank
// assembly relies on. The banks themselves never introduce path-hostile
// characters, so a strings.Fields round-trip is all the sanitising we need.
func cleanSaveName(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// --- EPIC: grand, dominion-flavoured. "The <Adj> <Noun>" / "<Root> <Suffix>".

var epicAdjectives = []string{
	"Gilded", "Verdant", "Azure", "Eternal", "Radiant", "Sovereign",
	"Ironclad", "Crimson", "Golden", "Boundless", "Towering", "Glorious",
	"Ascendant", "Resolute", "Imperial", "Hallowed",
}

var epicNouns = []string{
	"Dominion", "Ascendancy", "Reach", "Compact", "Hegemony", "Empire",
	"Realm", "Bastion", "Sovereignty", "Accord", "Expanse", "Citadel",
	"Crown", "Vanguard", "Concord", "Mandate",
}

var epicPlaceRoots = []string{
	"Ironhold", "Sunspire", "Stormhall", "Goldmere", "Highreach", "Frosthold",
	"Emberfall", "Greywatch", "Brightspire", "Stonereach", "Dawnhold", "Ashmere",
	"Westmarch", "Thornwall", "Silvermoor", "Oakenford",
}

var epicSuffixes = []string{
	"Ascendancy", "Hegemony", "Dominion", "Reach", "Compact", "Accord",
	"Realm", "Empire", "Sovereignty", "Mandate", "Expanse", "Crown",
}

func generateEpicName() string {
	if rand.Intn(2) == 0 {
		return "The " + pick(epicAdjectives) + " " + pick(epicNouns)
	}
	return pick(epicPlaceRoots) + " " + pick(epicSuffixes)
}

// --- MYTHIC: coined ancient-sounding names. No apostrophes (filesystem-safe).

var mythicRoots = []string{
	"Vaelthar", "Kael", "Drovia", "Zhanggar", "Obsidian", "Aeon",
	"Myrr", "Thalor", "Eldros", "Korvath", "Nyxara", "Velmara",
	"Azhul", "Sorrowmere", "Caldris", "Ophis",
}

var mythicConnectors = []string{
	"", "", "Drovia", "Vey", "Anar", "Mor", "Sael", "Thun",
	"Vora", "Eth", "Kah", "Ulm",
}

var mythicSuffixes = []string{
	"um", "ia", "or", "eth", "an", "is", "ar", "yx",
	"ondor", "athar", "evar", "umir",
}

var mythicAdjectives = []string{
	"Obsidian", "Eternal", "Forgotten", "Shrouded", "Ancient", "Hollow",
	"Sunken", "Veiled", "Silent", "Endless", "First", "Pale",
}

var mythicEpochNouns = []string{
	"Aeon", "Epoch", "Dominion", "Throne", "Expanse", "Covenant",
	"Sanctum", "Vault", "Spire", "Reach", "Hollow", "Marches",
}

func generateMythicName() string {
	switch rand.Intn(3) {
	case 0:
		// Coined single word: <Root><connector><suffix>
		return pick(mythicRoots) + pick(mythicConnectors) + pick(mythicSuffixes)
	case 1:
		// Two coined words: <Root> <Root+suffix>
		return pick(mythicRoots) + " " + pick(mythicRoots) + pick(mythicSuffixes)
	default:
		// "The <Adj> <EpochNoun>" / "<Adj> <Root>"
		if rand.Intn(2) == 0 {
			return "The " + pick(mythicAdjectives) + " " + pick(mythicEpochNouns)
		}
		return pick(mythicAdjectives) + " " + pick(mythicRoots)
	}
}

// --- WHIMSICAL: playful, low-stakes. "The <Adj> <Noun>" / "Grand Duchy of ...".

var whimsicalWholeNames = []string{
	"Cluckington", "Snoozeburg", "Wobbleton", "Crumbshire", "Mittensgard",
	"Biscuitania", "Fluffhaven", "Pottersville", "Snackopolis", "Lazyfields",
	"Bumbleshire", "Quackmoor",
}

var whimsicalAdjectives = []string{
	"Potato", "Mildly Cross", "Reluctant", "Comfy", "Sleepy", "Grumpy",
	"Snacking", "Wobbly", "Indignant", "Cozy", "Disgruntled", "Cheerful",
}

var whimsicalNouns = []string{
	"Kingdom", "Empire", "Republic", "Federation", "Collective", "Commonwealth",
	"Confederacy", "League", "Syndicate", "Assembly", "Council", "Order",
}

var whimsicalOfThings = []string{
	"Snacks", "Naps", "Mild Inconvenience", "Slightly Burnt Toast", "Spare Buttons",
	"Lost Socks", "Excellent Cheese", "Questionable Decisions", "Afternoon Tea",
	"Reasonable Doubt", "Comfortable Chairs", "Forgotten Errands",
}

func generateWhimsicalName() string {
	switch rand.Intn(3) {
	case 0:
		return pick(whimsicalWholeNames)
	case 1:
		return "The " + pick(whimsicalAdjectives) + " " + pick(whimsicalNouns)
	default:
		return "Grand Duchy of " + pick(whimsicalOfThings)
	}
}

// pick returns a uniformly random element from a non-empty slice.
func pick(bank []string) string {
	return bank[rand.Intn(len(bank))]
}
