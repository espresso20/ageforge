package flavor

// catalog.go is the DATA half of this package: the authored fragment banks and
// the sentence structures that compose them. Everything here is plain data, so
// widening a Moment's prose is an edit in this file (or catalog_eras.go) and
// nowhere else.
//
// # The two kinds of fragment, and why the split matters
//
// A fragment is legitimate in exactly one of two ways:
//
//  1. AGE-AGNOSTIC. No era-coded physical noun anywhere in it. It has to land in
//     the Stone Age and the Quantum Age alike, which is the standard
//     config/log_flavor.go already sets for the fixed pools. People, orders,
//     records, the return itself, reputation, morale, counting, complaining,
//     taking credit — all of that is timeless. Carts, gates, bells, militia,
//     quartermasters, valleys and marching columns are not. These fragments live
//     in the banks BELOW and fire in every age.
//
//  2. ERA-GATED. Vivid, specific, and eligible only inside its era bucket. These
//     live in catalog_eras.go, are named *_core_<era> / *_tail_c_<era>, and are
//     only ever referenced by templates carrying a matching Eras constraint.
//
// There is no third kind. A cart in an ungated bank is a bug — it is how the
// space age ended up being narrated by gate wardens — and TestNoEraBleed fails
// the build for it.
//
// # How to read a Moment
//
// Each Moment has one or more ANCHOR banks — the clause that states the outcome —
// plus supporting banks:
//
//	*_core       ANCHOR. Capitalized, sentence-initial, no terminal punctuation.
//	*_core_*     ANCHOR variants gated by Kind or Era.
//	*_lead       Capitalized scene-setter that PRECEDES the core as its own sentence.
//	*_tail_c     lowercase continuation, joined with ", and " / " — " / "; ".
//	*_tail_s     Capitalized sentence that FOLLOWS the core.
//	*_subject_c  lowercase continuation that uses {subject} in a safe frame.
//	*_res_*      lowercase continuation that names the resource.
//
// Every template must draw one slot from an anchor bank; TestAnchorInvariant
// enforces it, and Signatures() returns those banks so a text-only consumer can
// classify a generated line.
//
// # The two grammar traps, and how the frames dodge them
//
//  1. EXPEDITION SUBJECTS ARE VERB-LED. The roster is half order titles ("Raid
//     Bandit Camp", "Conquer Territory") and half noun phrases ("Scout Party",
//     "Naval Expedition"), so "the {subject} returns" is broken for half of them.
//     Every expedition fragment therefore treats {subject} as the TITLE OF AN
//     ORDER — "the {subject} order", "filed under {subject}", "the orders read
//     {subject}" — which reads correctly for both shapes. TestSubjectFrames pins
//     the allowed frames.
//
//  2. FACTION NAMES HAVE INCONSISTENT NUMBER. "Void Reavers" is plural,
//     "Merchant Guild" is singular, and both are on the same roster, so no present
//     -tense verb can agree with all of them. Faction fragments use {subject} as
//     an OBJECT ("toward the {subject}", "with the {subject}") or with a PAST
//     -tense verb, which is number-agnostic in English. Never a copula.
//     TestSubjectFrames denies "{subject} is/are/was/were/has/have/does/do".
//
// # Voice
//
// Dry, in-world, one or two short sentences. The joke is on the world — its
// clerks, its councils, its logistics — never on the player and never on the
// mechanic. Nothing here mentions ticks, rolls, slots, or caps.

// banks holds every authored fragment bank, keyed by the name templates
// reference: the age-agnostic banks in this file plus the era-gated banks in
// catalog_eras.go. Fragments carry no terminal punctuation; the template
// supplies it.
var banks = mergeBanks(neutralBanks, eraBanks)

// mergeBanks folds the bank tables into one map. A key collision is an authoring
// mistake that would silently discard a whole bank, so it panics at init rather
// than shipping a half-empty pool.
func mergeBanks(tables ...map[string][]string) map[string][]string {
	out := make(map[string][]string)
	for _, t := range tables {
		for name, bank := range t {
			if _, dup := out[name]; dup {
				panic("flavor: duplicate bank name " + name)
			}
			out[name] = bank
		}
	}
	return out
}

// neutralBanks are the AGE-AGNOSTIC banks. Nothing in here may name a cart, a
// gate, a bell, a militia, a quartermaster, a valley, a marching column, a
// harvest, or anything else that pins a sentence to one stretch of history. If a
// line would read oddly aboard a ship in the Galactic Age, it belongs in
// catalog_eras.go instead.
var neutralBanks = map[string][]string{

	// ========================================================================
	// EXPEDITION SUCCESS
	// ========================================================================

	// ANCHOR — kind-neutral, every age.
	"exp_success_core": {
		"The party comes back heavier than it left",
		"Everything the venture set out for is accounted for and put away",
		"The undertaking closes out in profit, and the record will say so",
		"The party returns intact, which is more than the last one managed",
		"Your people come home with full packs and a great deal to say about it",
		"The venture pays out, to nobody's greater surprise than the people who went",
		"More came back than there was ever any room for",
		"The party arrives home exactly when it said it would, which is the least believable part",
		"The work is done and the proof of it is stacked where everyone can walk past it",
		"The expedition ends the way everyone hoped and nobody expected",
		"They come back tired, filthy, and entirely unwilling to be modest about it",
		"The venture concludes well, provided nobody asks about the middle of it",
		"Everything that went out comes back, and a good deal besides",
		"The accounting comes out in your favour and stubbornly stays that way",
		"The return is uneventful, which the people who went will not be mentioning",
	},
	// ANCHOR — scouting only. Scouts, maps and rumours travel well between ages.
	"exp_success_core_scout": {
		"Your scouts return with the map redrawn and themselves half ruined",
		"The scouting party comes home early, which is either skill or luck",
		"The reconnaissance holds up: everything they promised was where they said",
		"Your scouts went further than they were ordered to and came back better for it",
		"The party returns with a map, a rumour, and something heavy nobody will identify",
		"The advance party is home, and the far country is a little less far",
	},
	// ANCHOR — military only. Soldiers, campaigns and spoils likewise.
	"exp_success_core_war": {
		"The campaign closes with the ranks mostly full and the objective taken",
		"The war party returns under its own command, which is the whole trick",
		"The force comes home in good order, which suggests the fighting went their way",
		"The campaign ends decisively enough that nobody argues about the details",
		"Your soldiers return with spoils, minor wounds, and expanded stories",
		"The expedition returns thinner than it left and considerably richer",
	},
	"exp_success_lead": {
		"Word runs ahead of the party by two days, as it always does",
		"Nobody at home admits to having doubted it",
		"The way out was long and the return is louder than the departure",
		"The council had already drafted the disappointing announcement",
		"Somebody has been keeping a tally, and it is a good one",
		"Nobody at home was expecting them for another week",
		"The news arrives first, arrives wrong, and is corrected at volume",
		"There had been a great deal of quiet doubt about this one",
	},
	"exp_success_tail_c": {
		"the stores are thrown open for the afternoon and stay open",
		"nobody has yet volunteered to do the counting",
		"the tally is disputed twice before anyone has eaten",
		"three separate accounts of the journey are already in circulation",
		"everyone is told to stop asking where it all came from",
		"the way home is described as far worse than it was",
		"somebody is being carried on shoulders who did very little",
		"the celebration is scheduled before the counting is finished",
		"the accounts are updated with visible satisfaction",
		"the story grows a little in each retelling",
		"the estimate is quietly revised upward and nobody objects",
		"two people are already claiming it was their idea",
	},
	"exp_success_tail_s": {
		"The council will take credit for it within the week",
		"A celebration is proposed and immediately over-budgeted",
		"Whoever is responsible for the stores asks, politely, for bigger stores",
		"Two of the party are already volunteering for the next one",
		"Somebody suggests a monument and somebody else suggests a bath",
		"The record will be tidied before it is read aloud",
		"Everything goes quiet again by morning",
		"A song about it is begun and, mercifully, abandoned",
		"The maps are corrected, and the corrections are argued over",
		"The whole affair is declared to have been the plan all along",
		"Nobody mentions the part of the plan that did not happen",
	},
	"exp_success_tail_triumph": {
		"Everyone hears about it, and everyone hears a different version",
		"It is, by every measure anyone at home is using, a triumph",
		"Nothing about the day will be understated in the retelling",
		"The celebrating starts early and declines to stop",
		"Even the people who keep the accounts are cheerful, which is unsettling",
		"Nobody is sent to bed at a reasonable hour",
		"Not one person present is prepared to call it luck",
	},
	"exp_success_tail_wry": {
		"The plan is now described as having been careful",
		"Everyone present has always believed in it",
		"The margin of success is not examined closely",
		"The council accepts the outcome and the credit together",
		"It is agreed the risk was minimal, in hindsight",
		"The one dissenting voice is offered a drink",
	},
	"exp_success_subject_c": {
		"the {subject} order is marked closed",
		"the record marked {subject} is ruled off with some ceremony",
		"what the standing order calls {subject} is done with at last",
		"the {subject} venture goes into the record as a good one",
		"somebody will file {subject} under matters concluded",
		"the orders read {subject}, and for once the orders were followed",
		"the {subject} business ends better than it was planned",
		"the standing order titled {subject} is struck from the board",
	},
	"exp_success_res_any": {
		"the whole return is {res} and very little else",
		"there is {res} enough to make the people counting it nervous",
		"most of what came back is {res}",
	},
	"exp_success_res_amt": {
		"the tally comes to {amt_res}, give or take an argument",
		"the record gains a line reading {amt_res}",
		"somebody counts {amt_res} and somebody else counts differently",
	},
	"exp_success_res_mass": {
		"{res_haul} is unloaded before the party has finished arriving",
		"{res_stores} will hold for a good while now",
		"there is {res_haul} to be argued over by morning",
	},

	// ========================================================================
	// EXPEDITION FAILURE
	// ========================================================================

	"exp_fail_core": {
		"The party comes back lighter than it left",
		"The venture ends short of everything it was sent for",
		"What returns is a fraction of what was promised",
		"The expedition is written up as a partial success by someone generous",
		"The party returns, and the returning is most of the achievement",
		"The undertaking fails in the ordinary way: slowly, and then all at once",
		"They come back with less than they carried out",
		"The venture goes badly, though not as badly as it could have",
		"The party is home, the packs are light, and nobody is talking",
		"The attempt does not survive its first contact with the real thing",
		"The return is early, which is never a good sign",
		"Something went wrong out there and the accounts of it do not agree",
		"The expedition ends without ceremony and with a great deal less than it was sent for",
		"Whatever was out there declines to be brought back",
		"The party comes home and the only thing that arrives intact is the party",
	},
	"exp_fail_core_scout": {
		"Your scouts come back with a shorter map than they left with",
		"The scouting party loses the trail and, for two days, itself",
		"The reconnaissance returns with rumours and no confirmations",
		"Your scouts report that everything out there is worse than the maps allowed for",
		"The advance party turns back at something nobody had thought to draw",
		"The scouts come home apologetic, which is at least honest",
	},
	"exp_fail_core_war": {
		"The campaign ends early and nobody present calls it a withdrawal",
		"The war party withdraws in reasonable order, which is the best that can be said",
		"The force comes home out of order and short of its complement",
		"The expedition returns thinner and no wealthier",
		"Your soldiers give ground, and then give a great deal of explanation",
		"The campaign is broken off before it becomes a disaster worth naming",
	},
	"exp_fail_lead": {
		"Word arrives ahead of the party and gets quieter as it travels",
		"The council prepares two announcements and uses the shorter one",
		"Nobody at home admits to having predicted this either",
		"The way home was the hardest part, by every account given",
		"The return is reported without further comment",
		"There was no announcement prepared for this outcome",
		"The people who stayed behind can tell from a distance",
		"Nothing about the return is loud",
	},
	"exp_fail_tail_c": {
		"the tally is finished in under a minute",
		"the stores stay shut",
		"nobody asks for the full account twice",
		"the map is amended in three places and believed in none",
		"the blame is distributed evenly and accepted by nobody",
		"the people who keep the accounts say nothing, loudly",
		"the distance is given more credit for it than the enemy",
		"the record is written carefully and filed quickly",
		"the party is fed and not questioned",
		"the survivors are allowed to keep their version",
		"somebody proposes a second attempt and is not thanked",
		"room is found in the record for a very small entry",
	},
	"exp_fail_tail_s": {
		"The council decides the plan was sound and everything else unreasonable",
		"A shorter route is proposed by someone who has never travelled it",
		"The maps are blamed, and then quietly corrected",
		"Nobody volunteers for the next one for several days",
		"The whole affair is described, in hindsight, as a reconnaissance",
		"A report is commissioned and will be read by no one",
		"Whoever keeps the stores is relieved on one count and worried on several",
		"It is agreed the expedition was unlucky rather than badly planned",
		"The next order is drafted with a great deal more caution",
		"Everything goes quiet and stays that way",
	},
	"exp_fail_tail_grim": {
		"Not everyone who left is accounted for",
		"The names are read out and the list is longer than the tally",
		"There are two funerals and no celebration",
		"The way out will be taken more carefully next time, by fewer people",
		"Nobody sings anything",
		"Time passes and the gap in the ranks stays where it is",
	},
	"exp_fail_subject_c": {
		"the {subject} order is closed without comment",
		"the record marked {subject} is ruled off without ceremony",
		"what the standing order calls {subject} is quietly abandoned",
		"the {subject} venture goes into the record as instructive",
		"somebody will file {subject} under matters best not revisited",
		"the orders read {subject}, and events read otherwise",
		"the {subject} business ends earlier than it was meant to",
		"the standing order titled {subject} is struck from the board anyway",
	},
	"exp_fail_res_any": {
		"what {res} came back would not fill a corner",
		"there is some {res} in what returned and not much else",
		"there is {res} enough to be embarrassing and no more",
	},
	"exp_fail_res_amt": {
		"the tally stops at {amt_res}",
		"the record gains one thin line reading {amt_res}",
		"somebody counts {amt_res} and then counts it again, hopefully",
	},
	"exp_fail_res_mass": {
		"{res_stores} will not notice the difference",
		"there is barely {res_haul} to show for all of it",
		"{res_stores} are exactly where they were",
	},

	// ========================================================================
	// ENCOUNTER — STANDOFF (contact with a civ at war, no harm done)
	// ========================================================================

	"enc_standoff_core": {
		"Your scouts and theirs see each other at a distance, and both parties withdraw",
		"An enemy force shadows the expedition most of a day, then breaks off",
		"Contact with the enemy, brief and inconclusive",
		"Two patrols meet by accident, take each other's measure, and go around",
		"The parties pass within sight and neither commits to anything",
		"A signal goes up, is answered, and comes to precisely nothing",
		"There is a long afternoon of looking at one another from a safe distance",
		"Both sides find the other at the same moment and both decline the invitation",
		"The enemy positions are counted from somewhere safe and then avoided entirely",
		"Someone signals across the gap and nobody signals back",
		"The expedition finds the enemy at rest, considers it at length, and moves on",
		"A scout of theirs and a scout of yours share a route for an hour without agreeing to",
		"The two forces spend the day moving parallel and pretending not to",
		"Somebody nervous does something regrettable and it comes to nothing at all",
		"Neither side does the one thing that would have started it",
		"The two parties spend an hour deciding not to be the one who begins",
	},
	"enc_standoff_lead": {
		"The border country is quiet in the way that means occupied",
		"There is something on the far side of it and it is not yours",
		"The war has not reached this place yet, only its scouts",
		"Both sides have business here and neither will say what",
		"There is room enough here for two parties who do not wish to meet",
		"Nothing has been agreed since the fighting started, least of all here",
		"Everyone out here is a long way from anyone who could give an order",
	},
	"enc_standoff_tail_c": {
		"nothing is gained and nothing is lost",
		"your people come home with nothing but the sighting",
		"the sighting is reported and filed",
		"both accounts of it will be exaggerated later",
		"the maps gain one small mark and no explanation",
		"the party keeps moving and does not look back twice",
		"everyone involved reports having been perfectly calm",
		"the encounter costs a day and nothing else",
		"the incident is entered in the log with admirable brevity",
		"no one draws anything, which takes some doing",
	},
	"enc_standoff_tail_s": {
		"The war continues elsewhere, on schedule",
		"Both parties will call it restraint when they report it",
		"The place is left to itself again by evening",
		"It is not peace, but it will do for the afternoon",
		"The report is one line long and entirely accurate",
		"Somebody will be commended for this and it will not be the right person",
		"The distance between the two parties is never once discussed",
		"Nothing about it changes anything, which is rather the point",
	},
	"enc_standoff_tail_wry": {
		"Both sides will describe their own restraint as strategy",
		"It is the most agreeable thing either army has done all year",
		"The report writes itself and says nothing",
		"History will not record it, which is fair",
		"Everyone goes home to a meal they have earned by not fighting",
		"The war is briefly, accidentally, civil",
	},
	"enc_standoff_subject_c": {
		"the party comes within sight of the {subject} and no closer",
		"there is no word exchanged with the {subject} at all",
		"the encounter with the {subject} ends the way it began",
		"your scouts count the strength of the {subject} and withdraw",
		"nothing passes between your people and the {subject}",
		"a message is not sent to the {subject}, on reflection",
	},

	// ========================================================================
	// ENCOUNTER — AT CAPACITY (there is no room for another favour)
	// ========================================================================

	"enc_cap_core": {
		"Your stores are already thick with foreign gifts, and the envoys are sent back with their crates unopened",
		"The list of outstanding favours will not take another name",
		"There is no room to carry another obligation this year",
		"The gifts are admired on arrival and declined immediately afterwards",
		"There is nowhere left to put a favour, let alone another one",
		"Whoever keeps the stores refuses the delivery, the last one being still unpacked",
		"The envoys are fed, thanked, and sent back the way they came",
		"You are already obliged in more directions than you can comfortably face",
		"The offer is made handsomely and turned down politely",
		"Every space that matters is spoken for twice over",
		"The party returns with courtesies and very little else",
		"There are too many debts of gratitude outstanding to take on another",
		"The gifts go back where they came from with their seals unbroken",
		"Generosity arrives at a bad moment and is asked to wait",
		"The generosity is real, the timing is terrible, and the answer is no",
	},
	"enc_cap_lead": {
		"It has been an unusually friendly year",
		"Foreign goodwill has become a storage problem",
		"You have spent the year accepting things",
		"There is a queue of envoys and a shortage of room",
		"Whoever keeps the count has been keeping it, and is unhappy about it",
		"Nobody planned for this much kindness",
	},
	"enc_cap_tail_c": {
		"the crates are never opened",
		"the people who would have to store it are quietly relieved",
		"the envoys take it better than expected",
		"the courtesies are exchanged at length and mean nothing",
		"nobody is offended, which takes effort on both sides",
		"the party comes home with stories instead of goods",
		"the record is closed with some firmness",
		"the gifts are admired and then loaded straight back up",
		"the journey back is described as pleasant",
		"the whole exchange takes an afternoon and produces a receipt",
	},
	"enc_cap_tail_s": {
		"This will be remembered the next time anyone asks for room",
		"It is agreed that the timing was nobody's fault",
		"More storage is proposed for the fourth time this year",
		"The envoys are invited back at a less crowded moment",
		"Somebody suggests giving something away and is ignored",
		"Your reputation for generosity survives intact",
		"The matter is recorded as a courtesy call and left there",
		"Everyone parts on good terms and slightly worse tempers",
	},
	"enc_cap_tail_wry": {
		"Prosperity is turning out to have logistics",
		"It is a good problem and it is still a problem",
		"The cautious position is vindicated and nobody enjoys it",
		"Somewhere, more room is being made, slowly",
		"You are rich in gestures and short of floor",
		"Being owed too much is a novel sort of complaint",
	},
	"enc_cap_subject_c": {
		"the party comes home with the goodwill of the {subject} and nothing heavier",
		"there is no room left for anything from the {subject}",
		"your people decline the offer from the {subject}, with real regret",
		"word is sent to the {subject} that the timing is poor",
		"nothing further is taken from the {subject} this year",
		"the arrangement with the {subject} can wait until there is room",
	},

	// ========================================================================
	// WAR RAID
	// ========================================================================

	"war_raid_core": {
		"They came at first light, took what they came for, and were gone before anyone was awake",
		"The raiders were in and out before the watch had finished shouting",
		"The outlying settlements are counted and one of them is short",
		"A raiding party crosses where nobody was watching and does not linger",
		"The attack is over by the time anyone is ready to answer it",
		"They came for the stores and they knew exactly where the stores were",
		"The raid is brief, efficient, and infuriatingly well informed",
		"The alarm goes up on the eastern edge and is answered far too late",
		"The force that hit the outer holdings was gone before anyone agreed on its size",
		"The watch reports the raid promptly, having watched it",
		"The defences held, and nothing behind the defences did",
		"They take what they can carry and burn a little of what they cannot",
		"The raiders leave the way they came, unhurried",
		"It is over quickly, as these things are",
		"Whoever planned it worked from a list, and the list was accurate",
		"The raid is answered promptly by people arriving from much too far away",
	},
	"war_raid_lead": {
		"The war has settled into a rhythm and this is the loud part of it",
		"There is nothing surprising about it any more, which is its own insult",
		"The border goes quiet for a stretch, and then it does not",
		"The outer holdings have learned to keep everything packed",
		"Word of the crossing arrives with the raiders",
		"This has happened often enough to have a routine",
		"A defence was assembled, briefly, and for nothing",
		"The alarm goes up along the frontier in entirely the wrong order",
	},
	"war_raid_tail_c": {
		"the accounts are corrected downward before noon",
		"the watch is doubled, several hours late",
		"nobody in the outer holdings is surprised",
		"what was broken is counted and none of it is mended",
		"the response arrives in time to look at where they went",
		"the accounting is done twice, hopefully",
		"somebody asks, again, for a wall",
		"the survivors are unimpressed by the response",
		"the loss is entered and the entry is not read aloud",
		"somebody tries to count them and gives up",
		"the people on watch are questioned and have very little to add",
		"an inquiry is announced and forgotten by evening",
	},
	"war_raid_tail_s": {
		"The council calls it a probe and moves on",
		"A stronger wall is proposed and costed and shelved",
		"The war continues without either side saying much about it",
		"The outer holdings will want an answer before long",
		"Retaliation is discussed at length by people who will not be going",
		"The border is redrawn on a map and nowhere else",
		"It will happen again on roughly the same schedule",
		"The garrison commander writes a very short report",
		"Nobody is blamed, which surprises the people who were there",
		"The alarms are repaired long before the fences are",
	},
	"war_raid_tail_grim": {
		"The outer holdings bury what they can find",
		"There is a house on the eastern edge that will not be rebuilt",
		"The tally of the missing is kept separately",
		"Somebody is still out there at dusk, counting",
		"Nothing is sounded again that night",
		"The war stops being an argument and becomes a fact",
	},
	"war_raid_subject_c": {
		"the raiders wore the colours of the {subject}",
		"there is no doubt at all that it was the {subject}",
		"everything points back toward the {subject}",
		"the war with the {subject} arrives at the outer holdings",
		"nothing is sent to the {subject} in reply, yet",
		"the raid is credited to the {subject} within the hour",
		"everyone blames the {subject} and, for once, is right",
		"a message is sent toward the {subject} and turns back at the border",
	},
	"war_raid_res_any": {
		"they took {res} and left the rest scattered",
		"it was the {res} they wanted and the {res} they took",
		"the {res} went first and fastest",
		"they knew exactly where to find the {res}",
	},
	"war_raid_res_mass": {
		"{res_stores} are lighter than they were at first light",
		"{res_haul} goes over the border in somebody else's hands",
		"{res_stores} will be short for a while yet",
		"they left {res_stores} standing open",
	},
}

// --- templates ---------------------------------------------------------------

// eraBucket pairs an era with the suffix its bank names use. The order is fixed
// so generated template sets are stable, which the determinism contract needs.
type eraBucket struct {
	era    era
	suffix string
}

var eraBuckets = []eraBucket{
	{eraAncient, "ancient"},
	{eraFeudal, "feudal"},
	{eraIndustrial, "industrial"},
	{eraDigital, "digital"},
	{eraCosmic, "cosmic"},
}

// eraTemplates expands one Moment's era-gated structure set: six shapes per
// bucket, each anchored on that bucket's core bank. Two are continued by the
// bucket's OWN tail; the other four pair the era anchor with the Moment's
// age-agnostic lead and follow-ons, which both widens the cross product and keeps
// the era voice from tipping over into pastiche.
//
// Six shapes is also what puts era-flavoured lines at roughly one in four of a
// fully-specified request's pool. Fewer and the late ages read as the neutral
// pool with a coat of paint; many more and every sentence starts shouting about
// its century.
//
// It is a data expansion, not logic — the alternative is a hundred and fifty
// near-identical literals, which is harder to read and easier to get wrong.
func eraTemplates(prefix string) []tmpl {
	out := make([]tmpl, 0, len(eraBuckets)*6)
	for _, bucket := range eraBuckets {
		var (
			core  = prefix + "_core_" + bucket.suffix
			tailC = prefix + "_tail_c_" + bucket.suffix
			lead  = prefix + "_lead"
			nTail = prefix + "_tail_c"
			tailS = prefix + "_tail_s"
			id    = prefix + "_" + bucket.suffix
			eras  = []era{bucket.era}
		)
		out = append(out,
			tmpl{ID: id + "_bare", Eras: eras, Parts: []part{b(core), lit(".")}},
			tmpl{ID: id + "_and", Eras: eras, Parts: []part{b(core), lit(", and "), b(tailC), lit(".")}},
			tmpl{ID: id + "_dash", Eras: eras, Parts: []part{b(core), lit(" — "), b(tailC), lit(".")}},
			tmpl{ID: id + "_then", Eras: eras, Parts: []part{b(core), lit(". "), b(tailS), lit(".")}},
			tmpl{ID: id + "_semi", Eras: eras, Parts: []part{b(core), lit("; "), b(nTail), lit(".")}},
			tmpl{ID: id + "_lead", Eras: eras, Parts: []part{b(lead), lit(". "), b(core), lit(".")}},
		)
	}
	return out
}

// eraCoreBanks names one Moment's per-era anchor banks, for anchorBanksFor.
func eraCoreBanks(prefix string) []string {
	out := make([]string, 0, len(eraBuckets))
	for _, bucket := range eraBuckets {
		out = append(out, prefix+"_core_"+bucket.suffix)
	}
	return out
}

// expSuccessTemplates is the ExpeditionSuccess structure set.
func expSuccessTemplates() []tmpl {
	return append([]tmpl{
		{ID: "exp_success_bare", Parts: []part{b("exp_success_core"), lit(".")}},
		{ID: "exp_success_and", Parts: []part{b("exp_success_core"), lit(", and "), b("exp_success_tail_c"), lit(".")}},
		{ID: "exp_success_dash", Parts: []part{b("exp_success_core"), lit(" — "), b("exp_success_tail_c"), lit(".")}},
		{ID: "exp_success_then", Parts: []part{b("exp_success_core"), lit(". "), b("exp_success_tail_s"), lit(".")}},
		{ID: "exp_success_lead_bare", Parts: []part{b("exp_success_lead"), lit(". "), b("exp_success_core"), lit(".")}},
		{ID: "exp_success_lead_then", Parts: []part{b("exp_success_lead"), lit(". "), b("exp_success_core"), lit(". "), b("exp_success_tail_s"), lit(".")}},

		// Kind-specific anchors.
		{ID: "exp_success_scout_then", Kinds: []string{"scouting"},
			Parts: []part{b("exp_success_core_scout"), lit(". "), b("exp_success_tail_s"), lit(".")}},
		{ID: "exp_success_scout_and", Kinds: []string{"scouting"},
			Parts: []part{b("exp_success_core_scout"), lit(", and "), b("exp_success_tail_c"), lit(".")}},
		{ID: "exp_success_war_then", Kinds: []string{"military"},
			Parts: []part{b("exp_success_core_war"), lit(". "), b("exp_success_tail_s"), lit(".")}},
		{ID: "exp_success_war_dash", Kinds: []string{"military"},
			Parts: []part{b("exp_success_core_war"), lit(" — "), b("exp_success_tail_c"), lit(".")}},

		// Tone-specific tails.
		{ID: "exp_success_triumph", Tones: []Tone{Triumphant},
			Parts: []part{b("exp_success_core"), lit(". "), b("exp_success_tail_triumph"), lit(".")}},
		{ID: "exp_success_wry", Tones: []Tone{Wry},
			Parts: []part{b("exp_success_core"), lit(". "), b("exp_success_tail_wry"), lit(".")}},

		// Subject (order title) and resource frames.
		{ID: "exp_success_subject", Needs: needSubject,
			Parts: []part{b("exp_success_core"), lit("; "), b("exp_success_subject_c"), lit(".")}},
		{ID: "exp_success_subject_then", Needs: needSubject,
			Parts: []part{b("exp_success_core"), lit("; "), b("exp_success_subject_c"), lit(". "), b("exp_success_tail_s"), lit(".")}},
		{ID: "exp_success_res", Needs: needRes,
			Parts: []part{b("exp_success_core"), lit("; "), b("exp_success_res_any"), lit(".")}},
		{ID: "exp_success_res_amt", Needs: needRes | needAmount,
			Parts: []part{b("exp_success_core"), lit(", and "), b("exp_success_res_amt"), lit(".")}},
		{ID: "exp_success_res_mass", Needs: needRes | needMassRes,
			Parts: []part{b("exp_success_core"), lit(" — "), b("exp_success_res_mass"), lit(".")}},
	}, eraTemplates("exp_success")...)
}

// expFailTemplates is the ExpeditionFailure structure set.
func expFailTemplates() []tmpl {
	return append([]tmpl{
		{ID: "exp_fail_bare", Parts: []part{b("exp_fail_core"), lit(".")}},
		{ID: "exp_fail_and", Parts: []part{b("exp_fail_core"), lit(", and "), b("exp_fail_tail_c"), lit(".")}},
		{ID: "exp_fail_dash", Parts: []part{b("exp_fail_core"), lit(" — "), b("exp_fail_tail_c"), lit(".")}},
		{ID: "exp_fail_then", Parts: []part{b("exp_fail_core"), lit(". "), b("exp_fail_tail_s"), lit(".")}},
		{ID: "exp_fail_lead_bare", Parts: []part{b("exp_fail_lead"), lit(". "), b("exp_fail_core"), lit(".")}},
		{ID: "exp_fail_lead_then", Parts: []part{b("exp_fail_lead"), lit(". "), b("exp_fail_core"), lit(". "), b("exp_fail_tail_s"), lit(".")}},

		{ID: "exp_fail_scout_then", Kinds: []string{"scouting"},
			Parts: []part{b("exp_fail_core_scout"), lit(". "), b("exp_fail_tail_s"), lit(".")}},
		{ID: "exp_fail_scout_and", Kinds: []string{"scouting"},
			Parts: []part{b("exp_fail_core_scout"), lit(", and "), b("exp_fail_tail_c"), lit(".")}},
		{ID: "exp_fail_war_then", Kinds: []string{"military"},
			Parts: []part{b("exp_fail_core_war"), lit(". "), b("exp_fail_tail_s"), lit(".")}},
		{ID: "exp_fail_war_dash", Kinds: []string{"military"},
			Parts: []part{b("exp_fail_core_war"), lit(" — "), b("exp_fail_tail_c"), lit(".")}},

		{ID: "exp_fail_grim", Tones: []Tone{Grim},
			Parts: []part{b("exp_fail_core"), lit(". "), b("exp_fail_tail_grim"), lit(".")}},

		{ID: "exp_fail_subject", Needs: needSubject,
			Parts: []part{b("exp_fail_core"), lit("; "), b("exp_fail_subject_c"), lit(".")}},
		{ID: "exp_fail_subject_then", Needs: needSubject,
			Parts: []part{b("exp_fail_core"), lit("; "), b("exp_fail_subject_c"), lit(". "), b("exp_fail_tail_s"), lit(".")}},
		{ID: "exp_fail_res", Needs: needRes,
			Parts: []part{b("exp_fail_core"), lit("; "), b("exp_fail_res_any"), lit(".")}},
		{ID: "exp_fail_res_amt", Needs: needRes | needAmount,
			Parts: []part{b("exp_fail_core"), lit(", and "), b("exp_fail_res_amt"), lit(".")}},
		{ID: "exp_fail_res_mass", Needs: needRes | needMassRes,
			Parts: []part{b("exp_fail_core"), lit(" — "), b("exp_fail_res_mass"), lit(".")}},
	}, eraTemplates("exp_fail")...)
}

// encStandoffTemplates is the EncounterStandoff structure set.
func encStandoffTemplates() []tmpl {
	return append([]tmpl{
		{ID: "enc_standoff_bare", Parts: []part{b("enc_standoff_core"), lit(".")}},
		{ID: "enc_standoff_and", Parts: []part{b("enc_standoff_core"), lit(", and "), b("enc_standoff_tail_c"), lit(".")}},
		{ID: "enc_standoff_dash", Parts: []part{b("enc_standoff_core"), lit(" — "), b("enc_standoff_tail_c"), lit(".")}},
		{ID: "enc_standoff_semi", Parts: []part{b("enc_standoff_core"), lit("; "), b("enc_standoff_tail_c"), lit(".")}},
		{ID: "enc_standoff_then", Parts: []part{b("enc_standoff_core"), lit(". "), b("enc_standoff_tail_s"), lit(".")}},
		{ID: "enc_standoff_lead_bare", Parts: []part{b("enc_standoff_lead"), lit(". "), b("enc_standoff_core"), lit(".")}},
		{ID: "enc_standoff_lead_and", Parts: []part{b("enc_standoff_lead"), lit(". "), b("enc_standoff_core"), lit(", and "), b("enc_standoff_tail_c"), lit(".")}},
		{ID: "enc_standoff_lead_then", Parts: []part{b("enc_standoff_lead"), lit(". "), b("enc_standoff_core"), lit(". "), b("enc_standoff_tail_s"), lit(".")}},
		{ID: "enc_standoff_wry", Tones: []Tone{Wry},
			Parts: []part{b("enc_standoff_core"), lit(". "), b("enc_standoff_tail_wry"), lit(".")}},
		{ID: "enc_standoff_subject", Needs: needSubject,
			Parts: []part{b("enc_standoff_core"), lit("; "), b("enc_standoff_subject_c"), lit(".")}},
		{ID: "enc_standoff_subject_then", Needs: needSubject,
			Parts: []part{b("enc_standoff_core"), lit("; "), b("enc_standoff_subject_c"), lit(". "), b("enc_standoff_tail_s"), lit(".")}},
	}, eraTemplates("enc_standoff")...)
}

// encCapacityTemplates is the EncounterAtCapacity structure set.
func encCapacityTemplates() []tmpl {
	return append([]tmpl{
		{ID: "enc_cap_bare", Parts: []part{b("enc_cap_core"), lit(".")}},
		{ID: "enc_cap_and", Parts: []part{b("enc_cap_core"), lit(", and "), b("enc_cap_tail_c"), lit(".")}},
		{ID: "enc_cap_dash", Parts: []part{b("enc_cap_core"), lit(" — "), b("enc_cap_tail_c"), lit(".")}},
		{ID: "enc_cap_semi", Parts: []part{b("enc_cap_core"), lit("; "), b("enc_cap_tail_c"), lit(".")}},
		{ID: "enc_cap_then", Parts: []part{b("enc_cap_core"), lit(". "), b("enc_cap_tail_s"), lit(".")}},
		{ID: "enc_cap_lead_bare", Parts: []part{b("enc_cap_lead"), lit(". "), b("enc_cap_core"), lit(".")}},
		{ID: "enc_cap_lead_and", Parts: []part{b("enc_cap_lead"), lit(". "), b("enc_cap_core"), lit(", and "), b("enc_cap_tail_c"), lit(".")}},
		{ID: "enc_cap_lead_then", Parts: []part{b("enc_cap_lead"), lit(". "), b("enc_cap_core"), lit(". "), b("enc_cap_tail_s"), lit(".")}},
		{ID: "enc_cap_wry", Tones: []Tone{Wry},
			Parts: []part{b("enc_cap_core"), lit(". "), b("enc_cap_tail_wry"), lit(".")}},
		{ID: "enc_cap_subject", Needs: needSubject,
			Parts: []part{b("enc_cap_core"), lit("; "), b("enc_cap_subject_c"), lit(".")}},
		{ID: "enc_cap_subject_then", Needs: needSubject,
			Parts: []part{b("enc_cap_core"), lit("; "), b("enc_cap_subject_c"), lit(". "), b("enc_cap_tail_s"), lit(".")}},
	}, eraTemplates("enc_cap")...)
}

// warRaidTemplates is the WarRaid structure set.
func warRaidTemplates() []tmpl {
	return append([]tmpl{
		{ID: "war_raid_bare", Parts: []part{b("war_raid_core"), lit(".")}},
		{ID: "war_raid_and", Parts: []part{b("war_raid_core"), lit(", and "), b("war_raid_tail_c"), lit(".")}},
		{ID: "war_raid_dash", Parts: []part{b("war_raid_core"), lit(" — "), b("war_raid_tail_c"), lit(".")}},
		{ID: "war_raid_semi", Parts: []part{b("war_raid_core"), lit("; "), b("war_raid_tail_c"), lit(".")}},
		{ID: "war_raid_then", Parts: []part{b("war_raid_core"), lit(". "), b("war_raid_tail_s"), lit(".")}},
		{ID: "war_raid_lead_bare", Parts: []part{b("war_raid_lead"), lit(". "), b("war_raid_core"), lit(".")}},
		{ID: "war_raid_lead_then", Parts: []part{b("war_raid_lead"), lit(". "), b("war_raid_core"), lit(". "), b("war_raid_tail_s"), lit(".")}},
		{ID: "war_raid_grim", Tones: []Tone{Grim},
			Parts: []part{b("war_raid_core"), lit(". "), b("war_raid_tail_grim"), lit(".")}},
		{ID: "war_raid_subject", Needs: needSubject,
			Parts: []part{b("war_raid_core"), lit("; "), b("war_raid_subject_c"), lit(".")}},
		{ID: "war_raid_subject_then", Needs: needSubject,
			Parts: []part{b("war_raid_core"), lit("; "), b("war_raid_subject_c"), lit(". "), b("war_raid_tail_s"), lit(".")}},
		{ID: "war_raid_res", Needs: needRes,
			Parts: []part{b("war_raid_core"), lit("; "), b("war_raid_res_any"), lit(".")}},
		{ID: "war_raid_res_then", Needs: needRes,
			Parts: []part{b("war_raid_core"), lit("; "), b("war_raid_res_any"), lit(". "), b("war_raid_tail_s"), lit(".")}},
		{ID: "war_raid_res_mass", Needs: needRes | needMassRes,
			Parts: []part{b("war_raid_core"), lit(" — "), b("war_raid_res_mass"), lit(".")}},
	}, eraTemplates("war_raid")...)
}

// templatesFor returns a Moment's template set in a stable order. An unregistered
// Moment returns nil, which is what makes Generate yield an empty Result for it.
func templatesFor(m Moment) []tmpl {
	switch m {
	case ExpeditionSuccess:
		return expSuccessTemplates()
	case ExpeditionFailure:
		return expFailTemplates()
	case EncounterStandoff:
		return encStandoffTemplates()
	case EncounterAtCapacity:
		return encCapacityTemplates()
	case WarRaid:
		return warRaidTemplates()
	default:
		return nil
	}
}

// anchorBanksFor returns the bank names whose fragments are GUARANTEED to appear
// verbatim in a Moment's output — every template draws exactly one slot from one
// of these. Signatures() is built from them; see the contract there.
func anchorBanksFor(m Moment) []string {
	switch m {
	case ExpeditionSuccess:
		return append([]string{"exp_success_core", "exp_success_core_scout", "exp_success_core_war"},
			eraCoreBanks("exp_success")...)
	case ExpeditionFailure:
		return append([]string{"exp_fail_core", "exp_fail_core_scout", "exp_fail_core_war"},
			eraCoreBanks("exp_fail")...)
	case EncounterStandoff:
		return append([]string{"enc_standoff_core"}, eraCoreBanks("enc_standoff")...)
	case EncounterAtCapacity:
		return append([]string{"enc_cap_core"}, eraCoreBanks("enc_cap")...)
	case WarRaid:
		return append([]string{"war_raid_core"}, eraCoreBanks("war_raid")...)
	default:
		return nil
	}
}
