package flavor

// ExpeditionSuccess — a party was sent out and it worked.
//
// The log line directly above this one already says "<Name> succeeded! Gained
// loot." So none of the sentences here say that it worked. They say what came
// back, who did not, what broke, what the map says now, what the town did about
// it, and who is already arguing over the share. See catalog.go for the rules.

// --- slot banks --------------------------------------------------------------
//
// Noun phrases, lowercase, one tight category per bank. A slot swaps the detail
// in a sentence; it never changes what the sentence is about.

var expSuccessBanks = map[string][]string{
	// What they carried in. Age-agnostic on purpose: sacks and bundles are as
	// true of a hunting party as of a boarding crew.
	"exp_success_haul": {
		"the sacks", "the bundles", "the crates", "the heavy bags", "the sealed boxes",
		"the packs", "the whole load", "the last two bundles", "the roped bundles",
		"the smaller sacks", "the long crates", "the unopened bags",
		"the padded boxes", "the split sacks", "the folded canvas", "the heavy end of it",
		"the burnt crates", "the last of the packs",
	},
	// A small physical thing that goes wrong or gets lost. Same rule.
	"exp_success_kit": {
		"lamp", "rope", "knife", "good coat", "water skin", "spare boot",
		"cooking pot", "second lamp", "tally stick", "flint", "belt knife",
		"good blanket", "long rope", "brass whistle", "walking staff",
		"leather satchel", "spare pin", "sharpening stone",
	},
}

// --- the sentences -----------------------------------------------------------

// expSuccessTemplates is the ExpeditionSuccess skeleton set: one ungated pool
// that lands in any age, plus one pool per era bucket for the imagery that does
// not travel.
func expSuccessTemplates() []tmpl {
	out := pool("exp_success_any", erasAny, expSuccessAny)
	out = append(out, pool("exp_success_ancient", erasAncient, expSuccessAncient)...)
	out = append(out, pool("exp_success_feudal", erasFeudal, expSuccessFeudal)...)
	out = append(out, pool("exp_success_industrial", erasIndustrial, expSuccessIndustrial)...)
	out = append(out, pool("exp_success_digital", erasDigital, expSuccessDigital)...)
	out = append(out, pool("exp_success_cosmic", erasCosmic, expSuccessCosmic)...)
	return out
}

// expSuccessAny fires in every age, so nothing in it may name an era-coded
// object. People, counting, arguing, sleeping, weather, names and dogs are
// timeless; carts, uplinks and bulkheads are not.
var expSuccessAny = []skel{
	// --- who came back, and in what state ---
	{Text: "Two of them came back wearing someone else's boots"},
	{Text: "The youngest one has not stopped talking since"},
	{Text: "Nobody has slept. Nobody intends to"},
	{Text: "Half of them are asleep sitting up"},
	{Text: "The oldest of them sat down and would not get up"},
	{Text: "Two of them have new scars and matching stories"},
	{Text: "One came back barefoot and will not explain the boots"},
	{Text: "The tall one has a bruise the exact shape of a boot"},
	{Text: "They walked in at dusk, quiet, and ate everything cold"},
	{Text: "One of them will not go near water now, and will not say why"},
	{Text: "The wound on the tall one's arm is being described as nothing"},
	{Text: "They came back a day early, which has upset the cooks"},
	{Text: "One of them lost a tooth and considers it a fair trade"},

	// --- who did not ---
	{Text: "The knife came back. The man who carried it did not"},
	{Text: "They buried two on the way out and did not mark the place"},
	{Text: "There is a list of names and it is shorter than it was"},
	{Text: "One of them has decided to stay out there. They let him"},
	{Text: "Two did not come back. The rest brought maps"},

	// --- what came in ---
	{Text: "They came back with ~ and a limp", Slot: "exp_success_haul"},
	{Text: "Someone counted ~ twice and got two answers", Slot: "exp_success_haul"},
	{Text: "There is blood on ~ and nobody is saying whose", Slot: "exp_success_haul"},
	{Text: "Somebody has already claimed ~ three separate times", Slot: "exp_success_haul"},
	{Text: "Everything is spread across the floor being counted"},
	{Text: "There is a prisoner here who will not give a name"},
	{Text: "A stranger walked in behind them and has not been asked to leave"},
	{Text: "They brought back a dog. The dog stays"},
	{Text: "The dog that went with them came back fatter"},

	// --- what broke or was lost ---
	{Text: "They lost the second ~ on the first night and managed anyway", Slot: "exp_success_kit"},
	{Text: "The last anyone saw of the ~ was two nights ago", Slot: "exp_success_kit"},
	{Text: "Directions were bought with a ~ and everyone calls it fair", Slot: "exp_success_kit"},
	{Text: "The rope held. That surprised everyone involved"},
	{Text: "Two ropes and one broken hand: the whole cost"},
	{Text: "The rain found ~ before the roof did", Slot: "exp_success_haul"},
	{Text: "Water got into ~ on the last night out", Slot: "exp_success_haul"},
	{Text: "Nobody will admit to opening ~ on the way home", Slot: "exp_success_haul"},
	{Text: "The dogs were very interested in ~", Slot: "exp_success_haul"},
	{Text: "Two men are sitting on ~ and will not move", Slot: "exp_success_haul"},
	{Text: "One of them came home without a ~ and will not say why", Slot: "exp_success_kit"},
	{Text: "There is an argument going on about a missing ~", Slot: "exp_success_kit"},
	{Text: "The only thing that came back whole is a ~", Slot: "exp_success_kit"},
	{Text: "A child has been given the broken ~ to play with", Slot: "exp_success_kit"},
	{Text: "Somebody left a ~ behind on purpose, they say", Slot: "exp_success_kit"},

	// --- what the map says now ---
	{Text: "The map is longer than it was this morning"},
	{Text: "They found water where the old maps said there was none"},
	{Text: "They marked the safe crossing and drew a face beside it"},
	{Text: "They left a marker at the crossing so the next lot can find it"},
	{Text: "The far side of the hills has a name now"},

	// --- the town ---
	{Text: "A child has already made up a song about it"},
	{Text: "The children were sent to bed and listened from the stairs anyway"},
	{Text: "Somebody's brother is a hero now and will be unbearable about it"},
	{Text: "There is a name being repeated tonight that nobody knew this morning"},
	{Text: "The old woman who told them where to look wants paying"},
	{Text: "Somebody has already promised half of it away"},
	{Text: "Three of them are arguing about who saw it first"},

	// --- the world, briefly ---
	{Text: "Something followed them for two days and then stopped"},
	{Text: "The rain started an hour after they were under a roof"},
	{Text: "The crossing was frozen. They walked over it and did not look down"},
	{Text: "The water was worse than the fighting, they say"},
	{Text: "They came back the long way and will not say why"},
	{Text: "Nobody sang on the walk home. That is not a bad sign"},
	{Text: "The story got better every time it was told on the way back"},
	{Text: "A woman named Serit did most of the work and none of the talking"},
	{Text: "They ate the last of the dried meat two days out and walked hungry"},

	// --- kind: scouting ---
	{Text: "Three days out, two days back, and a better map", Kinds: []string{"scouting"}},
	{Text: "They walked further than anyone told them to", Kinds: []string{"scouting"}},
	{Text: "The new marks on the map are in a very shaky hand", Kinds: []string{"scouting"}},
	{Text: "Nobody has been that far and come home still talking", Kinds: []string{"scouting"}},
	{Text: "They counted the far watchfires and stopped at forty", Kinds: []string{"scouting"}},

	// --- kind: military ---
	{Text: "The fighting was short and the walking was long", Kinds: []string{"military"}},
	{Text: "Two dead, and both of them were careless", Kinds: []string{"military"}},
	{Text: "They took the place at dawn and left before noon", Kinds: []string{"military"}},
	{Text: "One of them will not put the blade down yet", Kinds: []string{"military"}},
	{Text: "The wounded came in first, which is how you know it went well", Kinds: []string{"military"}},

	// --- tone ---
	{Text: "The drink is already gone and the night is young", Tones: []Tone{Triumphant}},
	{Text: "Somebody has decided this deserves a statue", Tones: []Tone{Triumphant}},
	{Text: "There is dancing, and a great deal of it is very bad", Tones: []Tone{Triumphant}},
	{Text: "Nobody has admitted the plan was mostly an accident", Tones: []Tone{Wry}},
	{Text: "The version told tonight has a sea monster in it", Tones: []Tone{Wry}},

	// --- the order title, in a frame a verb-led name survives ---
	{Text: "The orders marked {subject} came back signed and filthy", Needs: needSubject},
	{Text: "Somebody filed the {subject} order under things that worked", Needs: needSubject},
	{Text: "A page titled {subject} now has three names crossed out", Needs: needSubject},
	{Text: "One lamp and a good knife, filed under {subject}", Needs: needSubject},
	{Text: "Everything filed under {subject} came home muddy", Needs: needSubject},
	{Text: "The line that reads {subject} has a thumbprint on it now", Needs: needSubject},

	// --- the resource, as a thing in a room rather than as a payout ---
	{Text: "Nobody wants to sit up guarding the {res} tonight", Needs: needRes},
	{Text: "Two people have already argued about the {res} tonight", Needs: needRes},
	{Text: "The whole town knew about the {res} before the council did", Needs: needRes},
	{Text: "Whoever found the trail wants a bigger share of the {res}", Needs: needRes},
	{Text: "Somebody has already stolen a little of the {res}", Needs: needRes},
	{Text: "Everything is being counted twice, the {res} three times", Needs: needRes},
	{Text: "Nobody around here expected {res_stores} to be this full", Needs: needRes | needMassRes},
	{Text: "The count came to {amt_res}, twice, by two different people", Needs: needRes | needAmount},
	{Text: "The number everyone is repeating tonight is {amt_res}", Needs: needRes | needAmount},
}

// expSuccessAncient — fires, herds, hides, spears, the elders.
var expSuccessAncient = []skel{
	{Text: "The elders looked at what came back and said nothing useful"},
	{Text: "Two spears came home broken and one came home bloody"},
	{Text: "They drove a herd back with them, and not all of it is theirs"},
	{Text: "The hides are stiff with salt and will need working"},
	{Text: "Someone put a spear point through the doorway of the wrong hut"},
	{Text: "The drums went all night and the dogs hated every minute"},
	{Text: "The watch-fire was kept lit three nights for nothing"},
	{Text: "The store-pit is full enough to argue over"},
	{Text: "An arrow came back in a shoulder, still whole"},
	{Text: "They found a spring on the far side and drank it low"},
	{Text: "The ford was low enough to cross without swimming"},
	{Text: "Smoke on the ridge turned out to be someone else's trouble"},
	{Text: "The valley beyond has better grass and worse people"},
	{Text: "They came back along the old cut road and it held"},
	{Text: "Three of them walked the whole way under one hide"},
	{Text: "The elders want the spears counted before anyone eats"},
	{Text: "A boy carried a spear the whole way and never once used it"},
	{Text: "The herd lost two on the road and nobody blames the boy"},
	{Text: "There is a new hut going up before the mud has dried"},
	{Text: "They brought back stone that rings when you strike it"},
	{Text: "The harvest can wait. Everyone is out looking at what came in"},
	{Text: "A hide came back with marks on it that nobody can read"},
	{Text: "The old man who knows the ford has been proved right again"},
	{Text: "The drums stopped when the last of them was counted in"},
	{Text: "The cook-fire smoke went up straight, which is a good sign"},
	{Text: "A dog followed them home out of the valley and will not leave"},
	{Text: "They walked the column back at half speed and still arrived early"},
	{Text: "Somebody's spear is planted outside a doorway that is not his"},
}

// expSuccessFeudal — carts, gates, bells, clerks, stewards, musters.
var expSuccessFeudal = []skel{
	{Text: "The gate was opened early, which the steward will hear about"},
	{Text: "Two carts came back loaded and one came back on three wheels"},
	{Text: "The bells were rung by a boy who was not asked to"},
	{Text: "The quartermaster has counted it and wants it counted again"},
	{Text: "A rider went ahead with the news and beat them by an hour"},
	{Text: "The clerks are still writing and supper is going cold"},
	{Text: "One banner came home in pieces and is being sewn tonight"},
	{Text: "The militia turned out for nothing and are pretending otherwise"},
	{Text: "The storehouse doors have stood open since noon"},
	{Text: "Somebody spilled wax across the only clean parchment"},
	{Text: "The wagons are in the yard and nobody will unload them"},
	{Text: "A herald has already made the story longer than it was"},
	{Text: "They came through the village at dusk and woke all of it"},
	{Text: "The pickets saw them coming and rang nothing at all"},
	{Text: "The undercroft is fuller than the steward likes to admit"},
	{Text: "A horn went up at the gate and half of it was flat"},
	{Text: "The muster roll is two names shorter than it was"},
	{Text: "The road past the ford is passable again, barely"},
	{Text: "A cart wheel broke in the yard and everyone found it funny"},
	{Text: "The steward wants the tally before the men want their supper"},
	{Text: "Two riders came in muddy to the waist with no explanation"},
	{Text: "The clerks have run out of parchment and are using the backs"},
	{Text: "Somebody's banner is hanging over the gate without permission"},
	{Text: "The village priest has claimed a share and is being ignored"},
	{Text: "The harvest stood untouched while everyone watched the road"},
	{Text: "A wagon came in with a goat on it that nobody ordered"},
	{Text: "The bells go again at first light, apparently"},
	{Text: "Smoke over the valley all week, and none of it ours"},
}

// expSuccessIndustrial — depots, sidings, foremen, telegrams, triplicate.
var expSuccessIndustrial = []skel{
	{Text: "The telegram arrived a full day before the men did"},
	{Text: "Freight came in on the evening train and sat on the platform"},
	{Text: "The foreman has signed for all of it and gone home"},
	{Text: "Two crates went missing between the siding and the warehouse"},
	{Text: "The depot clock stopped at four and nobody has wound it"},
	{Text: "Somebody rang the office telephone at two in the morning"},
	{Text: "The payroll is short and the explanation is longer than the list"},
	{Text: "It has been written up in triplicate and filed in the wrong drawer"},
	{Text: "The lorries came back with mud to the axles"},
	{Text: "There is a wire on the foreman's desk nobody wants to open"},
	{Text: "The annexe is full and they are stacking it in the corridor"},
	{Text: "The shareholders will hear a considerably tidier version"},
	{Text: "A rail truck was left uncoupled and rolled into the fence"},
	{Text: "The men came off the train singing and were told to stop"},
	{Text: "The warehouse roof leaks and it all went under there anyway"},
	{Text: "The telegraph operator has been awake for two days"},
	{Text: "Smoke over the yard until midnight, and the neighbours complained"},
	{Text: "The valley road is cut, so everything came the long way round"},
	{Text: "Somebody's boots are still on the platform where he left them"},
	{Text: "The depot cat has moved into the new crates"},
	{Text: "The foreman's tally and the office tally differ by one crate"},
	{Text: "They telegraphed ahead and the message arrived badly garbled"},
	{Text: "The freight was signed for by a man who does not work here"},
	{Text: "Two of the column came home on stretchers and walked off them"},
	{Text: "There is a new lock on the warehouse door and three keys"},
	{Text: "The night shift was told to go home and worked through anyway"},
	{Text: "The siding is blocked until somebody moves the empty trucks"},
	{Text: "A photograph was taken. Everyone in it is squinting"},
}

// expSuccessDigital — uplinks, drones, feeds, analysts, channels.
var expSuccessDigital = []skel{
	{Text: "The uplink held for the whole run, which nobody expected"},
	{Text: "A drone came home with one rotor and a great many opinions"},
	{Text: "The feed cut at the worst moment and returned at the dullest"},
	{Text: "Two analysts are still arguing over the same forty seconds"},
	{Text: "The after-action report is longer than the operation was"},
	{Text: "Somebody left the channel open and the whole floor listened in"},
	{Text: "The network went down for six minutes and nobody noticed"},
	{Text: "The drones came back clean. The people did not"},
	{Text: "A junior analyst spotted it first and will never stop saying so"},
	{Text: "The feed from the second team is still buffering, hours later"},
	{Text: "Half the team has not been off the network since Tuesday"},
	{Text: "The channel logs have been pulled and will be quietly edited"},
	{Text: "Someone patched a drone with tape and it flew itself home"},
	{Text: "The uplink logs show a gap nobody wants to explain"},
	{Text: "There is coffee on the console and nobody will admit whose"},
	{Text: "The after-action briefing has been moved twice already"},
	{Text: "An analyst went home, slept four hours, and came back in"},
	{Text: "The drone footage has been watched sixty times and counting"},
	{Text: "Somebody named the drone. The name has stuck"},
	{Text: "The night channel was silent for two hours and then very loud"},
	{Text: "The network flagged the whole thing as routine, which it was not"},
	{Text: "Two laptops came back and one of them still switches on"},
	{Text: "The feed shows a hand, a doorway, and then nothing at all"},
	{Text: "The analysts have gone quiet, which is how they celebrate"},
	{Text: "Someone's badge is still on the desk and its owner is not"},
	{Text: "The uplink dish took a stone and carried on working"},
	{Text: "The channel is full of people asking what happened"},
	{Text: "A second drone went to find the first and both came home"},
}

// expSuccessCosmic — hulls, bays, relays, transponders, airlocks.
var expSuccessCosmic = []skel{
	{Text: "The hull came back scored down one side and holding"},
	{Text: "Cargo bay three is full and the manifest says it is empty"},
	{Text: "They came out of orbit hot and the dock crew swore at them"},
	{Text: "The transponder was dead for eleven hours and then it was not"},
	{Text: "Something is loose behind a bulkhead and nobody can find it"},
	{Text: "The airlock cycled twice on the way in. Nobody has asked why"},
	{Text: "The relay picked them up before the watch officer did"},
	{Text: "A freighter crew saw them pass and logged it as debris"},
	{Text: "The reactor ran hot the whole way and still ticks as it cools"},
	{Text: "Two of the crew have not taken the suits off yet"},
	{Text: "The manifest and the bay count differ by one crate"},
	{Text: "There is vacuum frost on everything they carried in"},
	{Text: "The dock lights were left burning all night for them"},
	{Text: "Somebody scratched a name into the hull plate near the seam"},
	{Text: "The bay doors froze half open and were beaten with a wrench"},
	{Text: "The transponder code came back wrong and was accepted anyway"},
	{Text: "A bulkhead panel came away in someone's hand on the way home"},
	{Text: "They spent the last day of the run on emergency air"},
	{Text: "The relay logs have a gap about the length of a nap"},
	{Text: "One of them will not go back out and has said so aloud"},
	{Text: "Orbital watch called it a clean approach. It was not"},
	{Text: "Everything in the forward bay smells of scorched insulation"},
	{Text: "The dock crew found a boot in the cargo net"},
	{Text: "They came in on the wrong vector and nobody corrected them"},
	{Text: "The hull sensors say the ship is fine. The ship disagrees"},
	{Text: "A repair patch made of tape is holding a whole compartment"},
	{Text: "The freighter that lent them fuel would like it back"},
	{Text: "The reactor alarm went off once they were safely docked"},
}
