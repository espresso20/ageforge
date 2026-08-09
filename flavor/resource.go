package flavor

import (
	"strconv"
	"strings"
	"sync"

	"github.com/espresso20/ageforge/config"
)

// Resource countability — the part of this package that stops the generator
// producing "foods", "a knowledge", or "1 soldiers".
//
// config.ResourceDef carries a display Name and nothing about grammar, and the
// roster is a mixed bag: most resources are MASS nouns (Food, Wood, Knowledge,
// Stone, Dark Matter, Quantum Flux), and three are already-plural COUNT nouns
// (Soldiers, Nanobots, Dark Matter Crystals). The only pluralizer that exists
// elsewhere in the tree naively appends "s", which is exactly the failure mode
// this table is here to prevent.
//
// So the countability lives HERE, next to the sentences that depend on it, and
// every key config ships must be classified — TestEveryResourceClassified fails
// the build if a new resource lands without an entry.

// nounClass is a resource label's grammatical class.
type nounClass int

const (
	// massNoun is uncountable: "food", "your food stores", "a haul of food".
	// NEVER "a food", never "foods", never "3 foods".
	massNoun nounClass = iota
	// countNoun is an already-PLURAL count noun: "soldiers", "nanobots". It takes
	// a number directly ("12 soldiers") and has a singular form for the n == 1
	// case ("1 soldier"). It must never take the mass frames ("your soldiers
	// stores" is not a sentence), which is why those frames are gated behind the
	// needMassRes template requirement.
	countNoun
)

// resourceClass classifies every key in config.BaseResources(). Unlisted keys
// default to massNoun, which is the safe direction to be wrong in: the mass
// frames never pluralize and never take an article.
var resourceClass = map[string]nounClass{
	// --- mass nouns (uncountable) -------------------------------------------
	"food":         massNoun,
	"wood":         massNoun,
	"knowledge":    massNoun,
	"faith":        massNoun,
	"stone":        massNoun,
	"iron":         massNoun,
	"gold":         massNoun,
	"coal":         massNoun,
	"marble":       massNoun,
	"iron_ore":     massNoun,
	"steel":        massNoun,
	"culture":      massNoun,
	"oil":          massNoun,
	"electricity":  massNoun,
	"uranium":      massNoun,
	"data":         massNoun,
	"titanium_ore": massNoun,
	"crypto":       massNoun,
	"plasma":       massNoun,
	"titanium":     massNoun,
	"dark_matter":  massNoun,
	"antimatter":   massNoun,
	"quantum_flux": massNoun,

	// --- count nouns (already plural in their display name) ------------------
	"soldiers":             countNoun,
	"nanobots":             countNoun,
	"dark_matter_crystals": countNoun,
}

// countSingular holds the SINGULAR label for each count noun, for the "1 X" case.
// Mass nouns need no entry — they do not inflect.
var countSingular = map[string]string{
	"soldiers":             "soldier",
	"nanobots":             "nanobot",
	"dark_matter_crystals": "dark matter crystal",
}

// labelOverride replaces the config display name where lowercasing it does not
// give the phrase we want mid-sentence. Empty today; kept as the documented seam
// so a future awkward display name is fixed here rather than in the sentences.
var labelOverride = map[string]string{}

// labelsOnce guards the lazily-built key→label map. config.ResourceByKey()
// rebuilds its map on every call, and the generator can run thousands of times a
// run, so the lookup is built once.
var (
	labelsOnce sync.Once
	labelCache map[string]string
)

// resourceLabel returns the plain, mid-sentence label for a resource key:
// the config display name lowercased ("Dark Matter Crystals" → "dark matter
// crystals"), an override if one is registered, or the key with underscores
// turned into spaces when config does not know the key at all.
//
// An empty key yields the generic "supplies", so a template that references a
// resource still reads if a caller forgets to set one.
func resourceLabel(key string) string {
	if key == "" {
		return "supplies"
	}
	labelsOnce.Do(func() {
		defs := config.ResourceByKey()
		labelCache = make(map[string]string, len(defs))
		for k, d := range defs {
			labelCache[k] = strings.ToLower(d.Name)
		}
	})
	if v, ok := labelOverride[key]; ok {
		return v
	}
	if v, ok := labelCache[key]; ok {
		return v
	}
	return strings.ReplaceAll(key, "_", " ")
}

// classOf returns a key's noun class, defaulting to massNoun for anything
// unclassified (see resourceClass).
func classOf(key string) nounClass {
	if c, ok := resourceClass[key]; ok {
		return c
	}
	return massNoun
}

// isMass reports whether a key is a mass noun. Templates that use the mass-only
// frames ({res_stores}, {res_haul}) declare needMassRes and are filtered out for
// count nouns rather than contorting the phrasing.
func isMass(key string) bool { return classOf(key) == massNoun }

// --- the grammatical frames -------------------------------------------------
//
// Four frames, each rendered from a placeholder in the templates. Two are safe
// for BOTH classes; two are mass-only and gated by needMassRes.

// phraseBare is "{res}": the bare label. Mass and count both work as a bare noun
// phrase ("took food", "took soldiers").
func phraseBare(key string) string { return resourceLabel(key) }

// phraseStores is "{res_stores}": MASS ONLY. "your food stores" — a plural head
// noun, so it takes a plural verb ("your food stores are lighter").
func phraseStores(key string) string { return "your " + resourceLabel(key) + " stores" }

// phraseHaul is "{res_haul}": MASS ONLY. "a haul of food" — the indefinite
// article attaches to "haul", never to the mass noun itself.
func phraseHaul(key string) string { return "a haul of " + resourceLabel(key) }

// phraseQuantity is "{amt_res}": a number and a label, correctly inflected.
// Mass nouns never inflect ("40 food"); count nouns take the singular at exactly
// one ("1 soldier") and the already-plural label otherwise ("12 soldiers").
func phraseQuantity(key string, n int) string {
	label := resourceLabel(key)
	if classOf(key) == countNoun && n == 1 {
		if s, ok := countSingular[key]; ok {
			label = s
		}
	}
	return strconv.Itoa(n) + " " + label
}
