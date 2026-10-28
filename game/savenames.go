package game

import (
	_ "embed"
	"math/rand"
	"strings"

	"github.com/espresso20/ageforge/pkg/wordlist"
)

// The words a suggested name is put together from are the lists of
// savenames.txt, which is built into the game. Each list is held by the
// variable of the same name below.
//
//go:embed savenames.txt
var saveNameFile string

// saveNameWords is the lists of savenames.txt by name.
var saveNameWords = readSaveNameWords(saveNameFile)

// readSaveNameWords reads the lists of a savenames.txt. An entry written
// (none) is an empty word. A file that does not parse stops the program at
// start, and TestSaveNameWordsAreSafe keeps one from reaching a build.
func readSaveNameWords(text string) map[string][]string {
	lists, err := wordlist.Parse(text)
	if err != nil {
		panic("game: savenames.txt: " + err.Error())
	}
	words := make(map[string][]string, len(lists))
	for _, list := range lists {
		for _, entry := range list.Entries {
			word := entry.Text
			if word == "(none)" {
				word = ""
			}
			words[list.Name()] = append(words[list.Name()], word)
		}
	}
	return words
}

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

// maybeThe returns "The " roughly half the time and "" the rest, so any
// template that opens with it also produces a bare "<Adj> <Noun>" form. This
// keeps the leading-article share from dominating the distribution.
// cleanSaveName trims, so a leading "" collapses cleanly.
func maybeThe() string {
	if rand.Intn(2) == 0 {
		return "The "
	}
	return ""
}

// polityPrefixes are varied "<X> of" phrasings for the "of" templates, so no
// single connective (the old hardcoded "Grand Duchy of") dominates. All entries
// are filesystem-safe: letters and single spaces only, no punctuation.
var polityPrefixes = saveNameWords["polityPrefixes"]

// --- EPIC: grand, dominion-flavoured. "The <Adj> <Noun>" / "<Root> <Suffix>".

var (
	epicAdjectives = saveNameWords["epicAdjectives"]
	epicNouns      = saveNameWords["epicNouns"]
	epicPlaceRoots = saveNameWords["epicPlaceRoots"]
	epicSuffixes   = saveNameWords["epicSuffixes"]
)

func generateEpicName() string {
	switch rand.Intn(5) {
	case 0:
		return maybeThe() + pick(epicAdjectives) + " " + pick(epicNouns)
	case 1:
		return pick(epicPlaceRoots) + " " + pick(epicSuffixes)
	case 2:
		return pick(epicAdjectives) + " " + pick(epicPlaceRoots)
	case 3:
		return pick(epicNouns) + " of " + pick(epicPlaceRoots)
	default:
		// e.g. "Dominion of Ironhold"
		return pick(polityPrefixes) + " " + pick(epicPlaceRoots)
	}
}

// --- MYTHIC: coined ancient-sounding names. No apostrophes (filesystem-safe).

var (
	mythicRoots      = saveNameWords["mythicRoots"]
	mythicConnectors = saveNameWords["mythicConnectors"]
	mythicSuffixes   = saveNameWords["mythicSuffixes"]
	mythicAdjectives = saveNameWords["mythicAdjectives"]
	mythicEpochNouns = saveNameWords["mythicEpochNouns"]
)

func generateMythicName() string {
	switch rand.Intn(3) {
	case 0:
		// Coined single word: <Root><connector><suffix>
		return pick(mythicRoots) + pick(mythicConnectors) + pick(mythicSuffixes)
	case 1:
		// Two coined words: <Root> <Root+suffix>
		return pick(mythicRoots) + " " + pick(mythicRoots) + pick(mythicSuffixes)
	default:
		// "[The] <Adj> <EpochNoun>" / "<Adj> <Root>"
		if rand.Intn(2) == 0 {
			return maybeThe() + pick(mythicAdjectives) + " " + pick(mythicEpochNouns)
		}
		return pick(mythicAdjectives) + " " + pick(mythicRoots)
	}
}

// --- WHIMSICAL: playful, low-stakes. "The <Adj> <Noun>" / "Grand Duchy of ...".

var (
	whimsicalWholeNames = saveNameWords["whimsicalWholeNames"]
	whimsicalAdjectives = saveNameWords["whimsicalAdjectives"]
	whimsicalNouns      = saveNameWords["whimsicalNouns"]
	whimsicalOfThings   = saveNameWords["whimsicalOfThings"]
)

func generateWhimsicalName() string {
	switch rand.Intn(5) {
	case 0:
		return pick(whimsicalWholeNames)
	case 1:
		return maybeThe() + pick(whimsicalAdjectives) + " " + pick(whimsicalNouns)
	case 2:
		// "Grand Duchy of" is now 1 of ~16 polity prefixes, not the only one.
		return pick(polityPrefixes) + " " + pick(whimsicalOfThings)
	case 3:
		return pick(whimsicalWholeNames) + " " + pick(whimsicalNouns)
	default:
		return pick(whimsicalAdjectives) + " " + pick(whimsicalWholeNames)
	}
}

// pick returns a uniformly random element from a non-empty slice.
func pick(bank []string) string {
	return bank[rand.Intn(len(bank))]
}
