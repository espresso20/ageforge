package config

import (
	"maps"
	"slices"

	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/words"
)

// The lexicon: the game's general words, each with its own name in every age.
// A scholar in the Bronze Age is a Scribe; in the Cyberpunk Age, a Netrunner.
// Game text that names a kind of person reads the current age's word through
// GameState.Term, never a hard-coded one.
//
// A term is a key ("workers") with a singular and a plural in each age. Adding
// a term later is one entry in each age's row of lexiconRows, nothing else:
// a term with no entry in an age (or no row for the age) falls back to its
// plain English word (plainTerms), and a key nobody knows falls back to the
// key itself.
//
// The people terms are the four kinds of person in the town: workers (they
// have a job), the reserve (they live in the homes and have no job yet),
// soldiers and scholars. The names are placeholders; the copy is the
// owner's to rewrite.

// Word is one term in one age: its singular and its plural.
type Word struct {
	One, Many string
}

// The people term keys.
const (
	TermWorkers  = "workers"
	TermReserve  = "reserve"
	TermSoldiers = "soldiers"
	TermScholars = "scholars"
)

// PeopleTerms are the people term keys, in the order the game lists them.
func PeopleTerms() []string {
	return []string{TermWorkers, TermReserve, TermSoldiers, TermScholars}
}

// t is a lexicon entry written short: t("Gatherer", "Gatherers").
func t(one, many string) Word { return Word{One: one, Many: many} }

// plainTerms are the plain English words, used where an age has no entry.
var plainTerms = map[string]Word{
	TermWorkers:   t("worker", "workers"),
	TermReserve:   t("villager", "villagers"),
	TermSoldiers:  t("soldier", "soldiers"),
	TermScholars:  t("scholar", "scholars"),
	"army_rank_0": t("Green", "Green"), "army_rank_1": t("Seasoned", "Seasoned"),
	"army_rank_2": t("Veteran", "Veteran"), "army_rank_3": t("Elite", "Elite"),
	"army_rank_4": t("Legendary", "Legendary"),
}

// lexiconRow is one age's words.
type lexiconRow struct {
	age   string
	terms map[string]Word
}

// lexiconRows is the lexicon, one row per age, in the order of ages.
var lexiconRows = []lexiconRow{
	{"primitive_age", map[string]Word{
		TermWorkers: t("Gatherer", "Gatherers"), TermReserve: t("Clan member", "Clanfolk"),
		TermSoldiers: t("Hunter", "Hunters"), TermScholars: t("Elder", "Elders")}},
	{"stone_age", map[string]Word{
		TermWorkers: t("Laborer", "Laborers"), TermReserve: t("Villager", "Villagers"),
		TermSoldiers: t("Warrior", "Warriors"), TermScholars: t("Storyteller", "Storytellers")}},
	{"bronze_age", map[string]Word{
		TermWorkers: t("Craftsperson", "Craftsfolk"), TermReserve: t("Villager", "Villagers"),
		TermSoldiers: t("Spearman", "Spearmen"), TermScholars: t("Scribe", "Scribes")}},
	{"iron_age", map[string]Word{
		TermWorkers: t("Laborer", "Laborers"), TermReserve: t("Townsperson", "Townsfolk"),
		TermSoldiers: t("Swordsman", "Swordsmen"), TermScholars: t("Sage", "Sages")}},
	{"classical_age", map[string]Word{
		TermWorkers: t("Freeman", "Freemen"), TermReserve: t("Citizen", "Citizens"),
		TermSoldiers: t("Hoplite", "Hoplites"), TermScholars: t("Philosopher", "Philosophers")}},
	{"medieval_age", map[string]Word{
		TermWorkers: t("Serf", "Serfs"), TermReserve: t("Commoner", "Commoners"),
		TermSoldiers: t("Man-at-arms", "Men-at-arms"), TermScholars: t("Monk", "Monks")}},
	{"renaissance_age", map[string]Word{
		TermWorkers: t("Guildsman", "Guildsmen"), TermReserve: t("Burgher", "Burghers"),
		TermSoldiers: t("Musketeer", "Musketeers"), TermScholars: t("Scholar", "Scholars")}},
	{"colonial_age", map[string]Word{
		TermWorkers: t("Settler", "Settlers"), TermReserve: t("Colonist", "Colonists"),
		TermSoldiers: t("Regular", "Regulars"), TermScholars: t("Naturalist", "Naturalists")}},
	{"industrial_age", map[string]Word{
		TermWorkers: t("Mill hand", "Mill hands"), TermReserve: t("Townsperson", "Townsfolk"),
		TermSoldiers: t("Rifleman", "Riflemen"), TermScholars: t("Engineer", "Engineers")}},
	{"victorian_age", map[string]Word{
		TermWorkers: t("Laborer", "Laborers"), TermReserve: t("Citizen", "Citizens"),
		TermSoldiers: t("Infantryman", "Infantry"), TermScholars: t("Scientist", "Scientists")}},
	{"electric_age", map[string]Word{
		TermWorkers: t("Lineman", "Linemen"), TermReserve: t("Citizen", "Citizens"),
		TermSoldiers: t("Infantryman", "Infantry"), TermScholars: t("Inventor", "Inventors")}},
	{"atomic_age", map[string]Word{
		TermWorkers: t("Worker", "Workers"), TermReserve: t("Civilian", "Civilians"),
		TermSoldiers: t("Trooper", "Troopers"), TermScholars: t("Physicist", "Physicists")}},
	{"modern_age", map[string]Word{
		TermWorkers: t("Employee", "Employees"), TermReserve: t("Civilian", "Civilians"),
		TermSoldiers: t("Soldier", "Soldiers"), TermScholars: t("Researcher", "Researchers")}},
	{"information_age", map[string]Word{
		TermWorkers: t("Staff member", "Staff"), TermReserve: t("Civilian", "Civilians"),
		TermSoldiers: t("Operator", "Operators"), TermScholars: t("Analyst", "Analysts")}},
	{"digital_age", map[string]Word{
		TermWorkers: t("Coder", "Coders"), TermReserve: t("Netizen", "Netizens"),
		TermSoldiers: t("Cyber trooper", "Cyber troops"), TermScholars: t("Data scientist", "Data scientists")}},
	{"cyberpunk_age", map[string]Word{
		TermWorkers: t("Wage slave", "Wage slaves"), TermReserve: t("Street dweller", "Street folk"),
		TermSoldiers: t("Merc", "Mercs"), TermScholars: t("Netrunner", "Netrunners")}},
	{"fusion_age", map[string]Word{
		TermWorkers: t("Technician", "Technicians"), TermReserve: t("Resident", "Residents"),
		TermSoldiers: t("Augment", "Augments"), TermScholars: t("Theorist", "Theorists")}},
	{"space_age", map[string]Word{
		TermWorkers: t("Crew member", "Crew"), TermReserve: t("Colonist", "Colonists"),
		TermSoldiers: t("Marine", "Marines"), TermScholars: t("Astronomer", "Astronomers")}},
	{"interstellar_age", map[string]Word{
		TermWorkers: t("Spacer", "Spacers"), TermReserve: t("Settler", "Settlers"),
		TermSoldiers: t("Ranger", "Rangers"), TermScholars: t("Xenologist", "Xenologists")}},
	{"galactic_age", map[string]Word{
		TermWorkers: t("Hand", "Hands"), TermReserve: t("Citizen", "Citizens"),
		TermSoldiers: t("Legionnaire", "Legionnaires"), TermScholars: t("Archivist", "Archivists")}},
	{"quantum_age", map[string]Word{
		TermWorkers: t("Operator", "Operators"), TermReserve: t("Entangled one", "Entangled"),
		TermSoldiers: t("Warden", "Wardens"), TermScholars: t("Quantum mind", "Quantum minds")}},
	{"transcendent_age", map[string]Word{
		TermWorkers: t("Shaper", "Shapers"), TermReserve: t("Soul", "Souls"),
		TermSoldiers: t("Paragon", "Paragons"), TermScholars: t("Oracle", "Oracles")}},
}

// Term is the word key has in age: its singular and plural. A term the age
// has no entry for gives its plain English word; a key with none gives the
// key itself. With VibeMode on, a word the script has a line for reads as
// that line (see LexiconCells for the ids).
func Term(age, key string) Word {
	w, scope := plainTerm(age, key)
	if scope != "" && vibe.On() {
		if s, ok := vibe.Text(lexiconID(scope, key, "one")); ok {
			w.One = s
		}
		if s, ok := vibe.Text(lexiconID(scope, key, "many")); ok {
			w.Many = s
		}
	}
	return w
}

// plainTerm is the word as the lexicon has it, and the scope its VibeMode
// rows are named by: the age when the age has its own entry, "plain" when
// the word is the plain English one, "" for a key nobody knows.
func plainTerm(age, key string) (Word, string) {
	for _, row := range lexiconRows {
		if row.age == age {
			if w, ok := row.terms[key]; ok {
				return w, age
			}
			break
		}
	}
	if w, ok := plainTerms[key]; ok {
		return w, lexiconPlain
	}
	return Word{One: key, Many: key}, ""
}

// lexiconPlain is the scope of the plain English words in a row id.
const lexiconPlain = "plain"

// lexiconID is the row id of one form of a term: lexicon.<age>.<key>.one,
// or .many. The plain English words are lexicon.plain.<key>.one.
func lexiconID(scope, key, form string) string {
	return "lexicon." + scope + "." + key + "." + form
}

// LexiconCells lists every word of the lexicon as a text row: the singular
// and plural of each term in each age, and then the plain English words.
// They are the lines VibeMode's script may replace, by the ids it gives.
func LexiconCells() []words.Cell {
	var cells []words.Cell
	add := func(scope, key string, w Word) {
		cells = append(cells,
			words.Cell{ID: lexiconID(scope, key, "one"), Field: "config.Word.One", Text: w.One},
			words.Cell{ID: lexiconID(scope, key, "many"), Field: "config.Word.Many", Text: w.Many})
	}
	for _, row := range lexiconRows {
		for _, k := range slices.Sorted(maps.Keys(row.terms)) {
			add(row.age, k, row.terms[k])
		}
	}
	for _, k := range slices.Sorted(maps.Keys(plainTerms)) {
		add(lexiconPlain, k, plainTerms[k])
	}
	return cells
}

// TermsFor is every term the lexicon holds for age, by key; a term the age
// has no entry for is left out (Term answers it with the plain word).
func TermsFor(age string) map[string]Word {
	out := map[string]Word{}
	for _, row := range lexiconRows {
		if row.age == age {
			for k, w := range row.terms {
				out[k] = w
			}
		}
	}
	return out
}
