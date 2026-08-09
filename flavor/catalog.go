package flavor

// catalog.go is the DATA half of this package: the authored fragment banks and
// the sentence structures that compose them. Everything here is plain data, so
// widening a Moment's prose is an edit in this file and nowhere else.
//
// # How to read a Moment
//
// Each Moment has one or more ANCHOR banks — the clause that states the outcome —
// plus supporting banks:
//
//	*_core       ANCHOR. Capitalized, sentence-initial, no terminal punctuation.
//	*_core_*     ANCHOR variants gated by Kind, Era, or Tone.
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
// reference. Fragments carry no terminal punctuation; the template supplies it.
var banks = map[string][]string{

	// ========================================================================
	// EXPEDITION SUCCESS
	// ========================================================================

	// ANCHOR — kind-neutral, early/industrial eras.
	"exp_success_core": {
		"The party comes back heavier than it left",
		"The column straggles in at dusk, loaded and unbothered",
		"Everything the venture set out for is now sitting in the yard",
		"The return party is met at the gate by a clerk with a very long list",
		"The undertaking closes out in profit, and the record will say so",
		"The party returns intact, which is more than the last one managed",
		"Your people come home with full packs and a great deal to say about it",
		"The venture pays out, to nobody's greater surprise than the people who went",
		"What came back fills more carts than anyone budgeted for",
		"The party arrives home on schedule, which the quartermaster finds suspicious",
		"The work is done and the proof of it is stacked by the gate",
		"The expedition ends the way everyone hoped and nobody expected",
		"They come back tired, filthy, and entirely unwilling to be modest about it",
		"The venture concludes well, provided nobody asks about the middle of it",
	},
	// ANCHOR — the same beat in the late ages, where nobody has a gate warden.
	"exp_success_core_late": {
		"The team returns on schedule and the manifest checks out",
		"Everything logged as recoverable was recovered",
		"The mission closes out clean, which the operations desk finds unnerving",
		"The crew comes back with full holds and unspent contingency",
		"The survey pays for itself twice over before the debrief ends",
		"The return is quiet and the cargo seals are intact",
		"The expedition files its report early, and nobody has an explanation for that",
		"What the mission set out to acquire is now on the inventory",
	},
	// ANCHOR — scouting only.
	"exp_success_core_scout": {
		"Your scouts return with the map redrawn and their boots ruined",
		"The scouting party comes home early, which is either skill or luck",
		"The reconnaissance holds up: everything they promised was where they said",
		"Your scouts walked further than they were ordered to and came back better for it",
		"The party returns with a map, a rumour, and something heavy in a sack",
		"The advance party is home, and the far country is a little less far",
	},
	// ANCHOR — military only.
	"exp_success_core_war": {
		"The campaign closes with the banners still up and the ranks mostly full",
		"The war party returns under its own colours, which is the whole trick",
		"The column comes home in step, which suggests the fighting went their way",
		"The campaign ends decisively enough that nobody argues about the details",
		"Your soldiers return with spoils, minor wounds, and expanded stories",
		"The muster returns thinner than it left and considerably richer",
	},
	"exp_success_lead": {
		"Word runs ahead of the party by two days, as it always does",
		"The gate is opened early on the strength of a rumour",
		"Nobody at home admits to having doubted it",
		"The bells are rung, briefly, by someone who was not asked to",
		"A crowd forms at the gate, mostly to see what is in the carts",
		"The road was long and the return is louder than the departure",
		"The council had already drafted the disappointing announcement",
		"Somebody has been keeping a tally, and it is a good one",
	},
	"exp_success_tail_c": {
		"the quartermaster is already complaining about where to put it",
		"the storehouse doors are propped open for the afternoon",
		"nobody has yet volunteered to do the counting",
		"the tally is disputed twice before supper",
		"three separate accounts of the journey are already in circulation",
		"the clerks are told to stop asking where it all came from",
		"the road home is described as far worse than it was",
		"somebody is being carried on shoulders who did very little",
		"the celebration is scheduled before the inventory is finished",
		"the ledger is updated with visible satisfaction",
		"the story grows a little in each retelling",
		"the gate is kept open until the last cart is through",
	},
	"exp_success_tail_s": {
		"The council will take credit for it within the week",
		"A feast is proposed and immediately over-budgeted",
		"The storehouse keeper asks, politely, for a bigger storehouse",
		"Two of the party are already volunteering for the next one",
		"Somebody suggests a monument and somebody else suggests a bath",
		"The record will be tidied before it is read aloud",
		"The roads go quiet again by morning",
		"A song about it is begun and, mercifully, abandoned",
		"The maps are corrected, and the corrections are argued over",
		"The whole affair is declared to have been the plan all along",
	},
	"exp_success_tail_triumph": {
		"The whole valley hears about it by nightfall",
		"It is, by every measure anyone at home is using, a triumph",
		"Nothing about the day will be understated in the retelling",
		"The banners come out and stay out",
		"Even the clerks are cheerful, which is unsettling",
		"The gate is left open late, on purpose",
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
		"the ledger entry marked {subject} is ruled off with some ceremony",
		"what the muster-roll calls {subject} is done with at last",
		"the {subject} venture goes into the record as a good one",
		"the clerks file {subject} under matters concluded",
		"the orders read {subject}, and for once the orders were followed",
		"the {subject} business ends better than it was planned",
		"the standing order titled {subject} is struck from the board",
	},
	"exp_success_res_any": {
		"the carts are full of {res} and very little else",
		"there is {res} enough to make the clerks nervous",
		"most of what came back is {res}",
	},
	"exp_success_res_amt": {
		"the tally comes to {amt_res}, give or take a cart",
		"the ledger gains a line reading {amt_res}",
		"somebody counts {amt_res} and somebody else counts differently",
	},
	"exp_success_res_mass": {
		"{res_haul} is unloaded before the party has finished dismounting",
		"{res_stores} will hold through the season now",
		"there is {res_haul} to be argued over by morning",
	},

	// ========================================================================
	// EXPEDITION FAILURE
	// ========================================================================

	"exp_fail_core": {
		"The party comes back lighter than it left",
		"The venture ends short of everything it was sent for",
		"What returns is a fraction of what was promised",
		"The column comes home the long way, quietly",
		"The expedition is written up as a partial success by someone generous",
		"The party returns, and the returning is most of the achievement",
		"The undertaking fails in the ordinary way: slowly, and then all at once",
		"They come back with less than they carried out",
		"The venture goes badly, though not as badly as it could have",
		"The party is home, the packs are light, and nobody is talking",
		"The attempt does not survive contact with the country it was sent to",
		"The return is early, which is never a good sign",
		"Something went wrong out there and the accounts of it do not agree",
		"The expedition ends without ceremony and with fewer carts",
	},
	"exp_fail_core_late": {
		"The mission returns degraded and under-provisioned",
		"The manifest and the cargo do not agree, and the cargo wins",
		"The operation is closed out as partially recovered",
		"The crew comes home with the hull intact and very little in it",
		"The survey returns early, having found the country uncooperative",
		"The mission ends inside acceptable losses, by a definition written afterwards",
		"Half of what was logged outbound is not logged inbound",
		"The debrief takes longer than the expedition did",
	},
	"exp_fail_core_scout": {
		"Your scouts come back with a shorter map than they left with",
		"The scouting party loses the trail and, for two days, itself",
		"The reconnaissance returns with rumours and no confirmations",
		"Your scouts report that the country is worse than the maps allowed for",
		"The advance party turns back at a river nobody had drawn",
		"The scouts come home apologetic, which is at least honest",
	},
	"exp_fail_core_war": {
		"The campaign ends with the banners furled early",
		"The war party withdraws in reasonable order, which is the best that can be said",
		"The column comes home out of step and short of its complement",
		"The muster returns thinner and no wealthier",
		"Your soldiers give ground, and then give a great deal of explanation",
		"The campaign is broken off before it becomes a disaster worth naming",
	},
	"exp_fail_lead": {
		"The gate is opened without any bells at all",
		"Word arrives ahead of the party and gets quieter as it travels",
		"The council prepares two announcements and uses the shorter one",
		"The crowd at the gate thins once the carts are counted",
		"Nobody at home admits to having predicted this either",
		"The road home was the hardest part, by every account given",
		"A clerk is sent to the gate and comes back with very little to write",
		"The watch reports the party's return without further comment",
	},
	"exp_fail_tail_c": {
		"the tally is finished in under a minute",
		"the storehouse doors stay shut",
		"nobody asks for the full account twice",
		"the map is amended in three places and believed in none",
		"the blame is distributed evenly and accepted by nobody",
		"the quartermaster says nothing, loudly",
		"the road is given more credit for it than the enemy",
		"the record is written carefully and filed quickly",
		"the party is fed and not questioned",
		"the survivors are allowed to keep their version",
		"somebody proposes a second attempt and is not thanked",
		"the clerks find room in the ledger for a very small entry",
	},
	"exp_fail_tail_s": {
		"The council decides the plan was sound and the country unreasonable",
		"A shorter route is proposed by someone who has never walked it",
		"The maps are blamed, and then quietly corrected",
		"Nobody volunteers for the next one before supper",
		"The whole affair is described, in hindsight, as a reconnaissance",
		"A report is commissioned and will be read by no one",
		"The storehouse keeper is relieved on one count and worried on several",
		"It is agreed the expedition was unlucky rather than badly planned",
		"The next order is drafted with a great deal more caution",
		"The roads go quiet and stay that way",
	},
	"exp_fail_tail_grim": {
		"Not everyone who left is accounted for",
		"The names are read out and the list is longer than the tally",
		"There are two funerals and no feast",
		"The road out will be walked more carefully next time, by fewer people",
		"Nobody sings anything",
		"The season turns and the gap in the ranks stays where it is",
	},
	"exp_fail_subject_c": {
		"the {subject} order is closed without comment",
		"the ledger entry marked {subject} is ruled off in a plainer hand",
		"what the muster-roll calls {subject} is quietly abandoned",
		"the {subject} venture goes into the record as instructive",
		"the clerks file {subject} under matters best not revisited",
		"the orders read {subject}, and the country read otherwise",
		"the {subject} business ends earlier than it was meant to",
		"the standing order titled {subject} is struck from the board anyway",
	},
	"exp_fail_res_any": {
		"what {res} came back would not fill a corner",
		"there is some {res} in the carts and not much else",
		"the carts hold {res} and a great deal of empty air",
	},
	"exp_fail_res_amt": {
		"the tally stops at {amt_res}",
		"the ledger gains one thin line reading {amt_res}",
		"somebody counts {amt_res} and then counts it again, hopefully",
	},
	"exp_fail_res_mass": {
		"{res_stores} will not notice the difference",
		"there is barely {res_haul} to show for the season",
		"{res_stores} are exactly where they were",
	},

	// ========================================================================
	// ENCOUNTER — STANDOFF (contact with a civ at war, no harm done)
	// ========================================================================

	"enc_standoff_core": {
		"Your scouts and theirs see each other across the valley, and both parties withdraw",
		"An enemy column shadows the expedition most of a day, then breaks off",
		"Contact with the enemy, brief and inconclusive",
		"Two patrols meet at a ford, take each other's measure, and go around",
		"The parties pass within sight and neither commits to anything",
		"A watch-fire is spotted, answered, and left alone by both sides",
		"There is a long afternoon of looking at one another from a safe distance",
		"Both sides find the other at the same moment and both decline the invitation",
		"The enemy pickets are counted from a ridge and then avoided entirely",
		"Someone shouts across the water and nobody shouts back",
		"The expedition finds the enemy camp, considers it at length, and walks past",
		"A scout of theirs and a scout of yours share a road for an hour without agreeing to",
		"The two columns spend the day marching parallel and pretending not to",
		"An arrow is loosed by somebody nervous and lands in nothing at all",
	},
	"enc_standoff_lead": {
		"The border country is quiet in the way that means occupied",
		"There is smoke on the far ridge and it is not yours",
		"The war has not reached this valley yet, only its scouts",
		"Both sides have business here and neither will say what",
		"The road is wide enough for two parties who do not wish to meet",
		"Nothing has been agreed since the fighting started, least of all here",
	},
	"enc_standoff_tail_c": {
		"nothing is gained and nothing is lost",
		"your people come home with nothing but the sighting",
		"the sighting is reported and filed",
		"both accounts of it will be exaggerated later",
		"the maps gain one small mark and no explanation",
		"the party keeps walking and does not look back twice",
		"everyone involved reports having been perfectly calm",
		"the encounter costs a day and nothing else",
		"the incident is entered in the log with admirable brevity",
		"no one draws anything, which takes some doing",
	},
	"enc_standoff_tail_s": {
		"The war continues elsewhere, on schedule",
		"Both parties will call it restraint when they report it",
		"The valley is left to itself again by evening",
		"It is not peace, but it will do for the afternoon",
		"The report is one line long and entirely accurate",
		"Somebody will be commended for this and it will not be the right person",
		"The distance between the two columns is never once discussed",
		"Nothing about it changes anything, which is rather the point",
	},
	"enc_standoff_tail_wry": {
		"Both sides will describe their own restraint as strategy",
		"It is the most agreeable thing either army has done all season",
		"The report writes itself and says nothing",
		"History will not record it, which is fair",
		"Everyone goes home to a supper they have earned by walking",
		"The war is briefly, accidentally, civil",
	},
	"enc_standoff_subject_c": {
		"the party comes within sight of the {subject} and no closer",
		"there is no word exchanged with the {subject} at all",
		"the encounter with the {subject} ends the way it began",
		"your scouts count the banners of the {subject} and withdraw",
		"nothing passes between your people and the {subject}",
		"a message is not sent to the {subject}, on reflection",
	},

	// ========================================================================
	// ENCOUNTER — AT CAPACITY (the court cannot hold another favour)
	// ========================================================================

	"enc_cap_core": {
		"Your stores are already thick with foreign gifts, and the envoys are sent home with their crates unopened",
		"The ledger of outstanding favours is full",
		"Your court can carry no more obligations this season",
		"The gifts are admired at the gate and declined at the door",
		"There is nowhere left to put a favour, let alone another one",
		"The steward refuses the crates on the grounds that the last lot are still unpacked",
		"The envoys are fed, thanked, and walked back to the road",
		"Your household is already obliged in more directions than it can face",
		"The offer is made handsomely and turned down politely",
		"Every shelf that matters is spoken for",
		"The party returns with courtesies and very little else",
		"There are too many debts of gratitude outstanding to take on another",
		"The crates go back down the road under the same wax seals they arrived with",
		"Generosity arrives at a bad moment and is asked to wait",
	},
	"enc_cap_lead": {
		"The season has been an unusually friendly one",
		"Foreign goodwill has become a storage problem",
		"The court has spent the year accepting things",
		"There is a queue of envoys and a shortage of shelves",
		"The stewards have been keeping count and they are unhappy about it",
		"Nobody planned for this much kindness",
	},
	"enc_cap_tail_c": {
		"the crates are never opened",
		"the steward is quietly relieved",
		"the envoys take it better than expected",
		"the courtesies are exchanged at length and mean nothing",
		"nobody is offended, which takes effort on both sides",
		"the party comes home with stories instead of goods",
		"the ledger is closed with some firmness",
		"the gifts are admired and then reloaded",
		"the road back is described as pleasant",
		"the whole exchange takes an afternoon and produces a receipt",
	},
	"enc_cap_tail_s": {
		"The stewards will remember this the next time they are asked for room",
		"It is agreed that the timing was nobody's fault",
		"A larger storehouse is proposed for the fourth time this year",
		"The envoys are invited back at a less crowded moment",
		"Somebody suggests giving something away and is ignored",
		"The court's reputation for generosity survives intact",
		"The matter is recorded as a courtesy call and left there",
		"Everyone parts on good terms and slightly worse tempers",
	},
	"enc_cap_tail_wry": {
		"Prosperity is turning out to have logistics",
		"It is a good problem and it is still a problem",
		"The steward's position is vindicated and nobody enjoys it",
		"Somewhere a shelf is being built, slowly",
		"The court is rich in gestures and short of floor",
		"Being owed too much is a novel complaint",
	},
	"enc_cap_subject_c": {
		"the party comes home with the goodwill of the {subject} and nothing heavier",
		"there is no room left for anything from the {subject}",
		"your stewards decline the offer from the {subject}, with real regret",
		"your stewards send word to the {subject} that the timing is poor",
		"nothing further is taken from the {subject} this season",
		"the arrangement with the {subject} can wait until there is room",
	},

	// ========================================================================
	// WAR RAID
	// ========================================================================

	"war_raid_core": {
		"They came at dawn, took what they came for, and were gone before the horns",
		"The raiders were in and out before the watch had finished shouting",
		"The border villages are counted and one of them is short",
		"A raiding party crosses at the shallow ford and does not linger",
		"The attack is over by the time the militia is dressed",
		"They came for the stores and they knew exactly where the stores were",
		"The raid is brief, efficient, and infuriatingly well informed",
		"Smoke goes up on the eastern road and is answered too late",
		"The column that hit the outer holdings was gone before anyone agreed on its size",
		"The watch reports the raid promptly, having watched it",
		"The gate held, and nothing behind the gate did",
		"They take what they can carry and burn a little of what they cannot",
		"The raiders leave the way they came, unhurried",
		"It is over quickly, as these things are",
	},
	"war_raid_lead": {
		"The war has settled into a rhythm and this is the loud part of it",
		"There is nothing surprising about it any more, which is its own insult",
		"The border goes quiet for a stretch, and then it does not",
		"The outer holdings have learned to keep the carts loaded",
		"Word of the crossing arrives with the raiders",
		"It is the season for this and everyone knows it",
		"The militia was assembled, briefly, and for nothing",
		"The bells go up along the valley in the wrong order",
	},
	"war_raid_tail_c": {
		"the ledger is corrected downward before noon",
		"the watch is doubled, several hours late",
		"nobody in the outer holdings is surprised",
		"the road is repaired and the fence is not",
		"the militia arrives in time to look at the tracks",
		"the accounting is done twice, hopefully",
		"the stewards ask, again, for a wall",
		"the survivors are unimpressed by the response",
		"the loss is entered and the entry is not read aloud",
		"somebody counts the tracks and gives up",
		"the gate wardens are questioned and have very little to add",
		"an inquiry is announced and forgotten by evening",
	},
	"war_raid_tail_s": {
		"The council calls it a probe and moves on",
		"A stronger wall is proposed and costed and shelved",
		"The war continues without either side saying much about it",
		"The outer holdings will want an answer by spring",
		"Retaliation is discussed at length by people who will not be going",
		"The border is redrawn on a map and nowhere else",
		"It will happen again on roughly the same schedule",
		"The garrison commander writes a very short report",
		"Nobody is blamed, which surprises the people who were there",
		"The bells are repaired before the fences are",
	},
	"war_raid_tail_grim": {
		"The outer holdings bury what they can find",
		"There is a house on the east road that will not be rebuilt",
		"The tally of the missing is kept separately",
		"Somebody is still walking the fields at dusk, counting",
		"The bells are not rung again that night",
		"The war stops being an argument and becomes a fact",
	},
	"war_raid_subject_c": {
		"the raiders wore the colours of the {subject}",
		"there is no doubt at all that it was the {subject}",
		"the tracks lead back toward the {subject}",
		"the war with the {subject} arrives at the outer holdings",
		"nothing is sent to the {subject} in reply, yet",
		"the raid is credited to the {subject} within the hour",
		"your stewards blame the {subject} and, for once, are right",
		"a rider is dispatched toward the {subject} and turns back at the border",
	},
	"war_raid_res_any": {
		"they took {res} and left the rest scattered",
		"it was the {res} they wanted and the {res} they took",
		"the {res} went first and fastest",
		"they knew exactly where to find the {res}",
	},
	"war_raid_res_mass": {
		"{res_stores} are lighter than they were at dawn",
		"{res_haul} goes over the border in somebody else's carts",
		"{res_stores} will be short until the season turns",
		"they left {res_stores} standing open",
	},
}

// --- templates ---------------------------------------------------------------

// expSuccessTemplates is the ExpeditionSuccess structure set.
func expSuccessTemplates() []tmpl {
	return []tmpl{
		{ID: "exp_success_bare", Parts: []part{b("exp_success_core"), lit(".")}},
		{ID: "exp_success_and", Parts: []part{b("exp_success_core"), lit(", and "), b("exp_success_tail_c"), lit(".")}},
		{ID: "exp_success_dash", Parts: []part{b("exp_success_core"), lit(" — "), b("exp_success_tail_c"), lit(".")}},
		{ID: "exp_success_then", Parts: []part{b("exp_success_core"), lit(". "), b("exp_success_tail_s"), lit(".")}},
		{ID: "exp_success_lead_bare", Parts: []part{b("exp_success_lead"), lit(". "), b("exp_success_core"), lit(".")}},
		{ID: "exp_success_lead_then", Parts: []part{b("exp_success_lead"), lit(". "), b("exp_success_core"), lit(". "), b("exp_success_tail_s"), lit(".")}},

		// Late-era anchors: same beats, no gate wardens.
		{ID: "exp_success_late_bare", Parts: []part{b("exp_success_core_late"), lit(".")},
			Eras: []era{eraFuture}},
		{ID: "exp_success_late_then", Parts: []part{b("exp_success_core_late"), lit(". "), b("exp_success_tail_s"), lit(".")},
			Eras: []era{eraFuture}},
		{ID: "exp_success_late_and", Parts: []part{b("exp_success_core_late"), lit(", and "), b("exp_success_tail_c"), lit(".")},
			Eras: []era{eraFuture}},

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
	}
}

// expFailTemplates is the ExpeditionFailure structure set.
func expFailTemplates() []tmpl {
	return []tmpl{
		{ID: "exp_fail_bare", Parts: []part{b("exp_fail_core"), lit(".")}},
		{ID: "exp_fail_and", Parts: []part{b("exp_fail_core"), lit(", and "), b("exp_fail_tail_c"), lit(".")}},
		{ID: "exp_fail_dash", Parts: []part{b("exp_fail_core"), lit(" — "), b("exp_fail_tail_c"), lit(".")}},
		{ID: "exp_fail_then", Parts: []part{b("exp_fail_core"), lit(". "), b("exp_fail_tail_s"), lit(".")}},
		{ID: "exp_fail_lead_bare", Parts: []part{b("exp_fail_lead"), lit(". "), b("exp_fail_core"), lit(".")}},
		{ID: "exp_fail_lead_then", Parts: []part{b("exp_fail_lead"), lit(". "), b("exp_fail_core"), lit(". "), b("exp_fail_tail_s"), lit(".")}},

		{ID: "exp_fail_late_bare", Parts: []part{b("exp_fail_core_late"), lit(".")},
			Eras: []era{eraFuture}},
		{ID: "exp_fail_late_then", Parts: []part{b("exp_fail_core_late"), lit(". "), b("exp_fail_tail_s"), lit(".")},
			Eras: []era{eraFuture}},
		{ID: "exp_fail_late_and", Parts: []part{b("exp_fail_core_late"), lit(", and "), b("exp_fail_tail_c"), lit(".")},
			Eras: []era{eraFuture}},

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
	}
}

// encStandoffTemplates is the EncounterStandoff structure set.
func encStandoffTemplates() []tmpl {
	return []tmpl{
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
	}
}

// encCapacityTemplates is the EncounterAtCapacity structure set.
func encCapacityTemplates() []tmpl {
	return []tmpl{
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
	}
}

// warRaidTemplates is the WarRaid structure set.
func warRaidTemplates() []tmpl {
	return []tmpl{
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
	}
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
		return []string{"exp_success_core", "exp_success_core_late", "exp_success_core_scout", "exp_success_core_war"}
	case ExpeditionFailure:
		return []string{"exp_fail_core", "exp_fail_core_late", "exp_fail_core_scout", "exp_fail_core_war"}
	case EncounterStandoff:
		return []string{"enc_standoff_core"}
	case EncounterAtCapacity:
		return []string{"enc_cap_core"}
	case WarRaid:
		return []string{"war_raid_core"}
	default:
		return nil
	}
}
