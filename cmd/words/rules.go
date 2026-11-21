package main

import "regexp"

// The game's copy rules, as the import applies them. They were read from the
// lint tests that the test suite held (ui/copy_guard_test.go, ui/stale_copy_test.go,
// ui/docs_lint_test.go, config/effect_text_test.go, config/event_flavor_test.go)
// before those tests were removed. A term retired here is refused by the
// import the same day; add a rule here when the copy needs one.

// copyExemptFiles are skipped by the wording rules over ui, game and boon.
// The value says why.
var copyExemptFiles = map[string]string{
	"game/devmode.go":           "dev console",
	"game/devcmd.go":            "dev console",
	"game/updater.go":           "self-update plumbing, not game text",
	"ui/splash_canvas.go":       "art",
	"ui/wonder_icon.go":         "art",
	"game/savenames.go":         "generated save names",
	"ui/theme_widgets.go":       "widget chrome",
	"game/expedition_flavor.go": "flavor catalog glue",
}

// copyExemptLiterals are literal substrings exempt from every wording rule.
// The civilization-log markers in game/catastrophe.go are parsed back out of
// saved history, so rewording them would break the stats of existing saves.
var copyExemptLiterals = []string{
	" — Endured ",
	" — Succumbed to ",
}

// retiredTerms are glossary variants the wording audit replaced. Matching is
// case-insensitive on whole words inside literals that read as prose (they
// contain a space), so map keys and identifiers are not flagged.
var retiredTerms = []pattern{
	{regexp.MustCompile(`(?i)\bstandings\b|%d standing\b|[0-9] standing\b|faction standing`), "opinion"},
	{regexp.MustCompile(`(?i)\bfavours?\b`), "boon (or Goodwill for the deal kind)"},
	{regexp.MustCompile(`(?i)\bvillagers?\b`), "worker"},
	{regexp.MustCompile(`(?i)\bcolour`), "color"},
	{regexp.MustCompile(`(?i)\bharbour`), "harbor"},
	{regexp.MustCompile(`(?i)\bcentre`), "center"},
	{regexp.MustCompile(`(?i)\bcivilisation`), "civilization"},
	{regexp.MustCompile(`(?i)\bdefence\b`), "defense"},
	{regexp.MustCompile(`(?i)\blabour\b`), "labor"},
	{regexp.MustCompile(`(?i)\bcancelled\b`), "canceled"},
	{regexp.MustCompile(`(?i)\bpop cap\b`), "housing"},
	{regexp.MustCompile(`(?i)\bhouse left\b`), "housing left"},
	{regexp.MustCompile(`\bproduction_all\b`), "all production"},
	{regexp.MustCompile(`\d/t\b`), "/tick"},
	{regexp.MustCompile(`(?i)\bmilitary cap\b`), "soldier storage"},
	{regexp.MustCompile(`(?i)\bspeed cap\b|unlocks \+0\.5x`), "nothing: there is no speed setting, the game runs at 1x and wonders raise no cap"},
	{regexp.MustCompile(`\bESC\b`), "Esc (e.g. \"(Esc to close)\")"},
	{regexp.MustCompile(`\(s\)`), "a plural helper (textfmt.Count)"},
}

// rawAmountRe finds a bare float verb followed by a resource slot, the shape
// of an amount printed without the shared formatter ("%.0f %s").
var rawAmountRe = regexp.MustCompile(`%[+]?\.[0-9]f %s`)

// bangRe finds an exclamation mark that ends a sentence or a clause.
var bangRe = regexp.MustCompile(`[A-Za-z0-9)\]]!(\s|$|\[|")`)

// retiredCopy is player-facing text that was wrong and has been removed. The
// note says why, so nobody restores it from an old branch or doc. The text is
// matched case-insensitively.
var retiredCopy = []pattern{
	{regexp.MustCompile(`(?i)` + regexp.QuoteMeta("Reach Colonial Age to discover factions")), "first contact comes from expeditions, and the Bronze Age civ is met long before Colonial"},
	{regexp.MustCompile(`(?i)` + regexp.QuoteMeta("Colonial Age and build an Embassy")), "embassies raise opinion with civs already met; they never gated first contact"},
}

// retiredDoc is wording the wiki must not use: either a fact that was wrong or
// a word the glossary replaced. Patterns match prose only (code spans are
// stripped), case-insensitively.
var retiredDoc = []pattern{
	// Glossary.
	{regexp.MustCompile(`(?i)\bvillagers?\b`), "the game says worker"},
	{regexp.MustCompile(`(?i)\bfavours?\b`), "timed civilization effects are boons and setbacks (and US spelling)"},
	{regexp.MustCompile(`(?i)\bfavor (deals?|kind)\b|\bFavor:`), "the deal kind that pays opinion is Goodwill"},
	{regexp.MustCompile(`(?i)(^|[^-\w])standings?\b`), "a civilization's attitude is opinion"},
	{regexp.MustCompile(`(?i)\bpop(ulation)? cap\b`), "the limit on workers is housing"},
	{regexp.MustCompile(`(?i)\bmilitary cap\b`), "not a stat the game has"},
	{regexp.MustCompile(`(?i)\bspeed cap\b|unlocks \+0\.5x`), "there is no speed setting: the game runs at 1x and wonders raise no cap"},
	{regexp.MustCompile(`production_all`), "write all production"},
	{regexp.MustCompile(`(?i)\b(harbour|colours?|centre|civilisation|defence|labour|theatre|organis|cancelled|fervour)`), "US spelling"},
	{regexp.MustCompile(`(?i)\b(economy|epoch|stats|trade|military) tab\b`), "the game has panels, not tabs"},
	// Stale facts.
	{regexp.MustCompile(`(?i)25\s*[-–]\s*75%.{0,20}no effect`), "morale is a continuous curve with no dead zone"},
	{regexp.MustCompile(`×1\.5 per tier`), "worker tiers scale food by 1.12, not 1.5"},
	{regexp.MustCompile(`(?i)advances? automatically`), "the player types advance"},
	{regexp.MustCompile(`(?i)quick-save`), "Esc closes a panel, or saves and returns to the main menu"},
	{regexp.MustCompile(`darwin-arm64|darwin-amd64`), "release assets are named ageforge-macos-*"},
	{regexp.MustCompile(`(?i)clear input`), "Ctrl+C has no handler; it quits"},
	{regexp.MustCompile(`library_of_congress`), "no such building"},
	{regexp.MustCompile(`(?i)reach colonial age`), "first contact comes from expeditions"},
	{regexp.MustCompile(`(?i)prestiiging`), "typo"},
	{regexp.MustCompile(`(?i)MIT License`), "LICENSE is the AgeForge Non-Commercial License"},
	{regexp.MustCompile(`(?i)relative to the directory you launch`), "data lives next to the binary"},
	// AI-writing tells.
	{regexp.MustCompile(`(?i)\b(revolutioni[sz]e|synergistic|synergy|flywheel|unleash|limitless|delve|tapestry)`), "stock word"},
}

// rawKeyRe finds a snake_case key left in player text.
var rawKeyRe = regexp.MustCompile(`\b[a-z]+_[a-z_]+\b`)

// styleBans are glossary and tone rules for game data text: US spelling,
// glossary terms, and stock words.
var styleBans = []string{
	"colour", "favour", "honour", "harbour", "centre", "theatre", "defence",
	"civilisation", "labour", "organis", "maximis", "fervour",
	"neighbour", "travelling", "realise", "recognise", "crystallise", "decentralise", "optimise", "optimising",
	"pop cap", "population cap", "military cap", "culture cap", "unlocks +",
	"gather rate", "standings", "villager",
	"transform", "revolutioni", "realm", "unleash", "limitless",
	"changes everything", "frontiers await", "ascension",
}

// verbLeftRe matches a leftover printf verb ("%s", "%d"); a percentage such as
// "15%" is fine.
var verbLeftRe = regexp.MustCompile(`%[a-zA-Z]`)

// maxLineRunes caps a finished line for the log panel it lands in.
const maxLineRunes = 260
