package flavor

// WarRaid — a civilization you are at war with came over the wall and took
// something.
//
// The log line directly above this one already names the raider, names the
// resource and gives the number. So nothing here says what was taken or how
// much. These sentences are the morning after: what they wrecked getting in,
// who watched and did nothing, which door gave, whose dog is dead, what a child
// says she saw, what got left in the mud on the way out, and who the town has
// decided to blame. See catalog.go for the rules.
//
// A raid is a bad day. It is played straight — the occasional dry line only
// lands because the four either side of it are flat.

// --- slot banks --------------------------------------------------------------
//
// Noun phrases, lowercase, one tight category per bank, no era-coded nouns: both
// are drawn by ungated sentences and so have to be true in a stone camp and on a
// ring station alike.

var warRaidBanks = map[string][]string{
	// Things they broke getting in, getting out, or apparently for the exercise.
	"war_raid_damage": {
		"the well rope", "the north door", "the water butt", "the long fence",
		"the kitchen roof", "the outer wall", "the drying racks", "the footbridge",
		"the door frame", "the grain sacks", "the big cook pot", "the ladder to the loft",
		"the shutters on the front", "the woodpile", "the rain barrel", "the back stairs",
		"the lock on the inner door", "the fence posts along the lane",
	},
	// Small things they dropped, abandoned or left on purpose. Objects only —
	// something a person can pick up, carry inside and put on a table.
	"war_raid_trace": {
		"a broken knife", "a boot with no laces", "a torn strip of cloth",
		"a coil of good rope", "a child's shoe", "a torn glove", "an empty water skin",
		"a helmet with the strap cut", "a knife with the tip snapped off",
		"a bundle nobody will open", "a length of chain", "a bag of somebody else's food",
		"a flask still half full", "a leather strap with teeth marks", "a bloodied rag",
		"a cap two sizes too big", "a ring on a string", "a burnt-out lamp",
	},
}

// --- the sentences -----------------------------------------------------------

// warRaidTemplates is the WarRaid skeleton set: one ungated pool that lands in
// any age, plus one pool per era bucket for the imagery that does not travel.
func warRaidTemplates() []tmpl {
	out := pool("war_raid_any", erasAny, warRaidAny)
	out = append(out, pool("war_raid_ancient", erasAncient, warRaidAncient)...)
	out = append(out, pool("war_raid_feudal", erasFeudal, warRaidFeudal)...)
	out = append(out, pool("war_raid_industrial", erasIndustrial, warRaidIndustrial)...)
	out = append(out, pool("war_raid_digital", erasDigital, warRaidDigital)...)
	out = append(out, pool("war_raid_cosmic", erasCosmic, warRaidCosmic)...)
	return out
}

// warRaidAny fires in every age, so nothing in it may name an era-coded object.
// Doors, walls, water, rope, blood, dogs, children, names, counting and bad
// weather are true of every age that has people in it; gates, sidings, feeds and
// airlocks are not.
var warRaidAny = []skel{
	// --- what they broke on the way through ---
	{Text: "The repairs start with ~ and get worse from there", Slot: "war_raid_damage"},
	{Text: "Somebody put a shoulder through ~ on the way out", Slot: "war_raid_damage"},
	{Text: "Nobody noticed ~ was broken until the morning", Slot: "war_raid_damage"},
	{Text: "The first thing anyone mended was ~", Slot: "war_raid_damage"},
	{Text: "Two men spent the night patching ~ in the rain", Slot: "war_raid_damage"},
	{Text: "What they did to ~ took deliberate effort", Slot: "war_raid_damage"},
	{Text: "The children are being kept away from ~", Slot: "war_raid_damage"},

	// --- what they left behind ---
	{Text: "In the doorway they left ~ and did not come back", Slot: "war_raid_trace"},
	{Text: "Whoever came through left ~ where the fire had been", Slot: "war_raid_trace"},
	{Text: "The dogs found ~ under the wall at first light", Slot: "war_raid_trace"},
	{Text: "A boy is refusing to give up ~ he found", Slot: "war_raid_trace"},
	{Text: "Somebody has kept ~ and will not say why", Slot: "war_raid_trace"},
	{Text: "Halfway up the lane somebody found ~", Slot: "war_raid_trace"},

	// --- who saw them, and did nothing ---
	{Text: "The man on the wall saw them coming and said nothing"},
	{Text: "Three people watched from an upstairs window and stayed there"},
	{Text: "Everyone claims to have been asleep, which cannot be true"},
	{Text: "The dogs barked for an hour before anyone got up"},
	{Text: "A woman down the lane counted them going past"},

	// --- a door ---
	{Text: "The back door was open from the inside"},
	{Text: "Whoever let them in used a key"},
	{Text: "The door held. The frame did not"},
	{Text: "They went through the same door twice, in and out"},

	// --- an animal ---
	{Text: "The grey dog is dead and the children know"},
	{Text: "Somebody's dog followed them out and has not come home"},
	{Text: "The animals were let loose and are still being counted"},

	// --- what the children say ---
	{Text: "A child says there were eleven of them. There were six"},
	{Text: "The smallest child hid under the floor and kept quiet"},
	{Text: "A boy who saw everything has stopped speaking entirely"},
	{Text: "The children have started playing a game about it"},
	{Text: "One of the girls bit a man and is proud of it"},

	// --- who fought back ---
	{Text: "The baker went at them with a shovel and lost"},
	{Text: "An old man got one of them with a rock"},
	{Text: "Two of ours are hurt and one of theirs is dead"},
	{Text: "The woman next door would not step aside for them"},
	{Text: "One of ours went down swinging and got back up"},

	// --- a body ---
	{Text: "They left one of their own dead in the mud"},
	{Text: "The dead man they left has no marks on his hands"},

	// --- a message, a mark ---
	{Text: "Someone cut a shape into the wall by the water"},
	{Text: "A message was left and nobody here can read it"},
	{Text: "They wrote a name on the wall and misspelled it"},
	{Text: "The mark on the door is meant to be found again"},

	// --- the sound and the smell afterwards ---
	{Text: "The quiet afterwards lasted most of an hour"},
	{Text: "It smells of burning down by the water and will for days"},
	{Text: "Nobody slept after. The dogs kept starting up again"},
	{Text: "There was shouting, then nothing, then shouting further off"},

	// --- what the neighbours are saying ---
	{Text: "The people two doors down are packing to leave"},
	{Text: "Word reached the next place along the water before dawn"},
	{Text: "Everyone has a theory and none of them agree"},
	{Text: "The story going around already has forty men in it"},

	// --- a name ---
	{Text: "A man called Hest is missing and nobody saw him go"},
	{Text: "The one who gave the orders had a scar and a limp"},
	{Text: "Somebody heard a name shouted and keeps repeating it"},
	{Text: "A woman named Alder pulled two people out of the water"},

	// --- what got repaired first ---
	{Text: "The well was cleaned out before anything else got touched"},
	{Text: "The wall goes up again tomorrow, higher this time"},
	{Text: "They fixed the water before they buried anyone"},

	// --- a grudge ---
	{Text: "The young men have opinions and are being kept busy"},
	{Text: "There is talk of going after them, and it is serious"},
	{Text: "Nobody here forgets anything. That is the local trouble"},

	// --- the weather that night ---
	{Text: "It rained hard, which is why the tracks are clear"},
	{Text: "There was no moon, which they will have counted on"},
	{Text: "The wind was up and covered most of the noise"},
	{Text: "It was the coldest night of the year and they still came"},

	// --- who is being blamed ---
	{Text: "The blame has settled on the man who sleeps by the door"},
	{Text: "Two people have already resigned from things they invented"},
	{Text: "The council met before light and got nowhere by noon"},
	{Text: "Somebody will answer for the door being left unbarred"},

	// --- counting, and the dark ---
	{Text: "The counting will take longer than the mending"},
	{Text: "Everything got moved indoors before it was properly light"},
	{Text: "The lamps stayed lit all night and are still lit"},

	// --- kind: aggressive ---
	{Text: "They came in daylight, which is a kind of statement", Kinds: []string{"aggressive"}},
	{Text: "They did not bother hiding the way they came in", Kinds: []string{"aggressive"}},
	{Text: "Nothing about this was quiet or quick or careful", Kinds: []string{"aggressive"}},
	{Text: "They stayed long enough to be sure of being seen", Kinds: []string{"aggressive"}},

	// --- kind: isolationist ---
	{Text: "Nobody here has ever seen one of them up close", Kinds: []string{"isolationist"}},
	{Text: "They came a very long way to do this", Kinds: []string{"isolationist"}},
	{Text: "They spoke to nobody and took nothing they could not carry", Kinds: []string{"isolationist"}},

	// --- tone: grim ---
	{Text: "There are three graves being dug and one is small", Tones: []Tone{Grim}},
	{Text: "The blood by the well has been washed twice already", Tones: []Tone{Grim}},
	{Text: "The wounded were carried in and the doors were barred", Tones: []Tone{Grim}},

	// --- the raider's name, in frames a plural or singular name survives ---
	{Text: "A name has been scratched into the door: {subject}", Needs: needSubject},
	{Text: "The children have learned to spell {subject} this week", Needs: needSubject},
	{Text: "Nobody here traded with {subject} anyway, and nobody will", Needs: needSubject},
	{Text: "Somebody keeps saying {subject} like a curse", Needs: needSubject},
	{Text: "The old feud with {subject} got worse tonight", Needs: needSubject},
	{Text: "Two of the wounded named {subject} before they slept", Needs: needSubject},

	// --- what was taken, as a thing that used to sit in a room ---
	{Text: "They went straight for the {res} and knew where it was", Needs: needRes},
	{Text: "There is a hole in the floor where the {res} sat", Needs: needRes},
	{Text: "Nobody had moved the {res} indoors, and nobody will forget that", Needs: needRes},
	{Text: "The doors on the {res} store were cut, not forced", Needs: needRes},
	{Text: "Whoever pointed them at the {res} lives here", Needs: needRes},
	{Text: "Everything else got stepped over on the way to the {res}", Needs: needRes},
	{Text: "The smell of {res} is still on the stairs", Needs: needRes},
	{Text: "Somebody stood in front of {res_stores} and got hit for it", Needs: needRes | needMassRes},
	{Text: "Two of them walked out with {res_haul} and no hurry", Needs: needRes | needMassRes},
}

// warRaidAncient — the elders, the herd, the store-pit, the watch-fire.
var warRaidAncient = []skel{
	{Text: "The elders have been sitting since before light"},
	{Text: "Two huts burned and a third was pulled down to stop it"},
	{Text: "A spear was left standing in the ground by the store-pit"},
	{Text: "They drove off half the herd and killed what would not walk"},
	{Text: "The store-pit was opened with a stone and emptied by hand"},
	{Text: "Somebody let the watch-fire go out before midnight"},
	{Text: "The drums started too late to be any use"},
	{Text: "An arrow came through the roof of the wrong hut"},
	{Text: "They crossed at the ford, which everyone said they would not"},
	{Text: "The hides put out to dry are gone or trampled"},
	{Text: "The ridge was full of smoke by the time anyone looked"},
	{Text: "The old spring below the camp has blood in it"},
	{Text: "A boy went after them with a spear and was carried back"},
	{Text: "They came up the old road and left the same way"},
	{Text: "Half the harvest was still in the ground, which saved it"},
	{Text: "The valley people saw them pass and lit nothing"},
	{Text: "One of them died on the slope and was left where he fell"},
	{Text: "Nobody will speak against the elders while they are deciding"},
	{Text: "A hide with a hand-mark on it was nailed to a post"},
	{Text: "They fired the drying racks and the dogs scattered"},
	{Text: "The column that came through was longer than the camp"},
	{Text: "Water was carried up from the ford all morning"},
	{Text: "The herd dogs are dead and the camp is very quiet"},
	{Text: "Three huts stand open with nothing left inside them"},
	{Text: "An old woman held a spear at her door and they went round"},
	{Text: "The tracks lead up the valley and stop at hard ground"},
	{Text: "Someone has cut marks into the post by the store-pit"},
	{Text: "The smoke has gone but the smell has not"},
}

// warRaidFeudal — the gate, the bells, the muster, the steward's arithmetic.
var warRaidFeudal = []skel{
	{Text: "The gate stood open and nobody has explained how"},
	{Text: "Riders came through the lower village before the bells started"},
	{Text: "The bells rang while they were already leaving"},
	{Text: "Two carts were taken and a third was tipped into the ditch"},
	{Text: "The storehouse doors were cut off their hinges"},
	{Text: "Two pickets were found tied up and unhurt"},
	{Text: "The steward has counted what is left and gone very quiet"},
	{Text: "A clerk was hurt trying to carry the rolls out"},
	{Text: "Nobody called the militia until it was already over"},
	{Text: "Supper went cold in the hall while everyone stood outside"},
	{Text: "Somebody has hung a torn banner over the gate"},
	{Text: "The undercroft was found open and mostly empty"},
	{Text: "An arrow went through the shutter and into the wall"},
	{Text: "The muster is called for dawn and nobody expects much"},
	{Text: "Wax was still soft on the letter they did not take"},
	{Text: "They came up the road past the ford in good order"},
	{Text: "The quartermaster wants the yard cleared before he counts anything"},
	{Text: "A herald arrived afterwards with something insulting on parchment"},
	{Text: "Two wagons are missing and one is in the river"},
	{Text: "The horns went up in the wrong order and confused everyone"},
	{Text: "The village below lost its roofs and blames us for it"},
	{Text: "Smoke over the harvest fields until it rained"},
	{Text: "One of the riders dropped a glove in the yard"},
	{Text: "The gate keeper is dead and his son is asking questions"},
	{Text: "They took the carts and left the animals in the traces"},
	{Text: "A banner was left in the mud, which will mean something"},
	{Text: "The spring below the wall runs red and will for days"},
	{Text: "Nobody rang the bell at the ford crossing"},
}

// warRaidIndustrial — the siding, the warehouse, the wire, the night shift.
var warRaidIndustrial = []skel{
	{Text: "The telegram went out an hour after they had gone"},
	{Text: "They came in along the rail siding on foot"},
	{Text: "The warehouse doors were cut and the lock left hanging"},
	{Text: "A night watchman is in hospital and will not lose the eye"},
	{Text: "The foreman was on the platform and did nothing at all"},
	{Text: "Freight was pushed off the flatbed and left in the mud"},
	{Text: "The telephone line was cut before anything else happened"},
	{Text: "Two lorries went out after them and came back empty"},
	{Text: "The payroll safe was opened with something heavy"},
	{Text: "The depot office has a boot print on the door"},
	{Text: "There was smoke over the yard until the rain came"},
	{Text: "The wire to the next station was down all night"},
	{Text: "A train was held two hours and nobody said why"},
	{Text: "They cut the fence by the siding and walked straight in"},
	{Text: "The annexe roof is holed and the paperwork is ruined"},
	{Text: "It will be written up in triplicate and read by nobody"},
	{Text: "The shareholders will be told about the fence"},
	{Text: "The foreman's tally and the watchman's account do not agree"},
	{Text: "Men are sleeping in the warehouse tonight with the lamps lit"},
	{Text: "The valley road was blocked with a felled tree behind them"},
	{Text: "Somebody telegraphed the wrong office and lost forty minutes"},
	{Text: "Nobody has seen the depot cat since Tuesday night"},
	{Text: "The platform lamps were smashed on the way out"},
	{Text: "A column of them walked out along the rail line"},
	{Text: "The company will not be telling the newspapers"},
	{Text: "Two men from the night shift have not come in"},
	{Text: "The office clock stopped when the wall took the blow"},
	{Text: "They knew which shed held the good stock"},
}

// warRaidDigital — the feed, the uplink, the drones, the after-action.
var warRaidDigital = []skel{
	{Text: "The feed went dark ninety seconds before they arrived"},
	{Text: "Two drones went up late and found nothing but dust"},
	{Text: "The uplink was cut at the pole outside the fence"},
	{Text: "An analyst watched the whole thing and could not stop it"},
	{Text: "The network is back and half the cameras are not"},
	{Text: "Somebody posted the footage before the after-action started"},
	{Text: "The channel filled with people who were not there"},
	{Text: "They came down the service road with the lights off"},
	{Text: "A drone found the vehicles abandoned four hours later"},
	{Text: "The night analyst has been awake since it started"},
	{Text: "Power went first, then the doors, then everything else"},
	{Text: "The last clean image is a shoulder and half a face"},
	{Text: "Two of the guards are in hospital and one is not"},
	{Text: "Nobody wants to chair the after-action meeting tomorrow"},
	{Text: "The badge reader logged forty entries in one minute"},
	{Text: "They knew the camera angles better than the guards did"},
	{Text: "Someone on the network is selling the footage already"},
	{Text: "The drones are grounded until somebody explains the gap"},
	{Text: "A junior analyst called it in and was told to wait"},
	{Text: "The channel logs stop for eleven minutes and resume clean"},
	{Text: "Half the floor is watching the same nine seconds"},
	{Text: "The uplink came back up on its own at dawn"},
	{Text: "Somebody's phone is still transmitting from inside a bag"},
	{Text: "The network flagged it as a maintenance window"},
	{Text: "Three doors were opened with credentials that expired last year"},
	{Text: "The feed of the far fence has been dead for months"},
	{Text: "Every screen in the room is showing the same corridor"},
	{Text: "An analyst has drawn the route on a paper map"},
}

// warRaidCosmic — the hull, the bay, the airlock, the dock crew.
var warRaidCosmic = []skel{
	{Text: "They came in with the transponder dead and nobody challenged it"},
	{Text: "The cargo bay doors were cut, not opened"},
	{Text: "Two of the dock crew are dead and one is missing"},
	{Text: "The hull has a hole in it the size of a table"},
	{Text: "A bulkhead was blown inward and the section is still sealed"},
	{Text: "Nothing in the bay matches what the manifest says"},
	{Text: "They went out through the aft airlock and left it open"},
	{Text: "The relay logged them and flagged nothing at all"},
	{Text: "Vacuum took the forward compartment before anyone reached the doors"},
	{Text: "The reactor was left alone, which is the one mercy"},
	{Text: "A freighter three hours out saw them and said nothing"},
	{Text: "Orbital watch had them and lost them behind the moon"},
	{Text: "The airlock seals are cut and the spares are gone"},
	{Text: "Somebody was in the bay and did not get out"},
	{Text: "The dock lights were off, which they should not have been"},
	{Text: "Whoever opened the inner door had a working code"},
	{Text: "The hull plating outside the bay is scored in lines"},
	{Text: "Two relays went quiet for a minute and came back clean"},
	{Text: "The reactor alarm ran the whole time and nobody heard it"},
	{Text: "There is frost on everything down the port corridor"},
	{Text: "A child was found sealed in a locker, breathing"},
	{Text: "The transponder code they used belonged to a scrapped ship"},
	{Text: "Somebody wrote on the bulkhead in grease and left it"},
	{Text: "The dock crew are welding through the night shift"},
	{Text: "Air is being rationed until the bay is sealed again"},
	{Text: "They left an airlock cycling and it ran for an hour"},
	{Text: "The manifest was altered before anyone thought to check"},
	{Text: "Half the ring has no pressure and the rest has questions"},
}
