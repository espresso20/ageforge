package config

import "strings"

// A badge belongs to one worker domain. The rule is fixed: the badge's
// family decides, and its subject where the family is about a line, a
// domain, a resource, an age or a ladder. A badge no rule names resolves to
// "", which the tests refuse.

var ageDomain = map[string]string{
	"stone_age": "food", "bronze_age": "metallurgy", "iron_age": "metallurgy",
	"classical_age": "masonry", "medieval_age": "military", "renaissance_age": "trade",
	"colonial_age": "trade", "industrial_age": "engineering", "victorian_age": "energy",
	"electric_age": "energy", "atomic_age": "energy", "modern_age": "engineering",
	"information_age": "hacker", "digital_age": "hacker", "cyberpunk_age": "hacker",
	"space_age": "astronaut", "interstellar_age": "astronaut", "galactic_age": "astronaut",
	"fusion_age": "astronaut", "quantum_age": "astronaut", "transcendent_age": "astronaut",
	"primitive_age": "food",
}

var resourceDomain = map[string]string{
	"food": "food", "wood": "lumber", "stone": "masonry", "marble": "masonry",
	"faith": "faith", "culture": "faith", "gold": "trade", "soldiers": "military",
	"iron": "metallurgy", "iron_ore": "metallurgy", "steel": "metallurgy",
	"titanium": "metallurgy", "titanium_ore": "metallurgy",
	"coal": "energy", "oil": "energy", "electricity": "energy", "uranium": "energy",
	"plasma": "energy", "data": "hacker", "crypto": "hacker", "nanobots": "hacker",
	"quantum_flux": "hacker", "antimatter": "astronaut", "dark_matter": "astronaut",
	"dark_matter_crystals": "astronaut",
}

var lineageDomain = map[string]string{
	"food": "food", "organic_extraction": "lumber", "geological_extraction": "masonry",
	"faith": "faith", "trade": "trade", "harbor": "trade", "military": "military",
	"engineering": "engineering", "metallurgy": "metallurgy", "energy": "energy",
	"hacker": "hacker", "housing": "masonry", "storage": "masonry",
	"culture_arts": "faith", "knowledge": "engineering",
}

// ladderDomain is keyed by the ladder's name, the middle of "ladder.<name>.<n>".
var ladderDomain = map[string]string{
	"prestiges": "faith", "endured": "faith", "succumbed": "faith", "threads": "faith",
	"appeased": "faith", "braced": "faith", "invited": "faith", "exposed": "faith",
	"festivals": "faith", "memories": "faith", "wonders": "masonry", "upgrades": "masonry",
	"techs": "engineering", "milestones": "engineering", "chains": "engineering",
	"plan": "engineering", "expeditions": "trade", "deals": "trade", "blackmarket": "trade",
	"raids": "military", "hardships": "military", "returns": "food",
}

// specialDomain is keyed by the part of the key after "special.".
var specialDomain = map[string]string{
	"hut_hoarder": "masonry", "zoning_board": "masonry", "every_nook": "masonry",
	"absentee_landlord": "masonry", "twenty_two_wonders": "masonry", "old_bones": "masonry",
	"nothing_to_endure": "faith", "family_heirlooms": "faith", "asked_for_it": "faith",
	"tithe_to_a_liar": "faith", "self_fulfilling": "faith", "prepared_for_nothing": "faith",
	"full_chorus": "faith", "selective_hearing": "faith", "collector": "faith",
	"curator": "faith", "archivist": "faith", "museum_piece": "faith", "full_set": "faith",
	"rogues_gallery": "faith", "connoisseur_of_endings": "faith", "unkillable": "faith",
	"every_day_a_holiday": "faith", "once_in_a_legend": "faith",
	"hands_on": "lumber", "artisanal": "lumber",
	"early_riser": "food", "fashionably_late": "food", "while_you_were_out": "food",
	"maximum_leave": "food", "regular": "food", "standing_appointment": "food",
	"long_game": "food", "portion_control": "food", "grumbling": "food", "everything_full": "food",
	"sandbagged": "military", "bare_hands": "military", "standing_guard": "military",
	"walked_out_whole": "military", "not_today": "military", "protection_money": "military",
	"belt_and_braces": "military", "character_building": "military",
	"express_lane": "trade", "signed_the_guest_book": "trade", "paid_in_full": "trade",
	"knock_knock": "trade", "cold_shoulder": "trade", "popular": "trade",
	"everybodys_friend": "trade", "regifting": "trade", "loyal_customer": "trade",
	"no_questions_asked": "trade", "small_world": "trade", "liquidation_sale": "trade",
	"lucky_break": "trade", "hand_in_the_cookie_jar": "trade", "creative_accounting": "trade",
	"chain_reaction": "engineering", "six_for_six": "engineering", "box_ticker": "engineering",
	"forgetful": "hacker", "deja_vu": "hacker", "written_down_somewhere": "hacker",
	"touched_by_the_source": "hacker",
	"heirloom_universe":     "astronaut", "we_are_not_alone": "astronaut",
}

// Domain is the worker domain the badge counts toward, one of WorkerDomains,
// or "" for a badge no rule names.
func (d BadgeDef) Domain() string {
	switch d.Family {
	case "domain":
		return d.Subject
	case "age":
		return ageDomain[d.Subject]
	case "lineage":
		return lineageDomain[d.Subject]
	case "resource":
		return resourceDomain[d.Subject]
	case "ladder":
		return ladderDomain[keyPart(d.Key, 1)]
	case "special":
		return specialDomain[keyPart(d.Key, 1)]
	case "civ", "expedition":
		return "trade"
	case "harbinger", "doom", "awakening", "theme":
		return "faith"
	case "wonder", "map", "maximalist":
		return "masonry"
	case "techs", "swift":
		return "engineering"
	}
	return ""
}

// keyPart is the n-th dot-separated part of a badge key ("" past the end).
func keyPart(key string, n int) string {
	if parts := strings.Split(key, "."); n < len(parts) {
		return parts[n]
	}
	return ""
}

// RankTitles are the six ranks a domain earns, lowest first. Rank 0 has no
// title. The wording is a placeholder.
var RankTitles = [6]string{"Apprentice", "Journeyman", "Adept", "Expert", "Master", "Grandmaster"}

// RankThresholds are the domain points each rank takes, lowest first.
var RankThresholds = [6]int{25, 100, 190, 320, 460, 700}

// RankBonus is the output a domain's workers gain per rank.
const RankBonus = 0.02

// RankFor is the rank, 0 to 6, that points reach.
func RankFor(points int) int {
	r := 0
	for _, t := range RankThresholds {
		if points < t {
			break
		}
		r++
	}
	return r
}
