package flavor

// EncounterAtCapacity — somebody arrived with an offer and the court could not
// take it.
//
// The log line above this one has already said so. None of the sentences here
// repeat it. They say where the gift got put, who has been sitting in the hall
// since dawn, which animal nobody will claim, what has gone soft in the corner,
// and which name has been spelled two different ways on the same page. The
// comedy is bureaucratic saturation, and it works best reported flat — see
// catalog.go, rule 3.

// --- slot banks --------------------------------------------------------------
//
// Two tight categories: the thing that arrived, and the place it was quietly
// put. Both are ungated, so both are free of era-coded nouns — a crate and a
// back corner are as true of a longhouse as of an orbital station.

var encCapacityBanks = map[string][]string{
	// What turned up that nobody has anywhere to put.
	"enc_cap_gift": {
		"a crate of dried fruit", "a very large mirror", "two more caged birds",
		"a painted chest", "a bolt of heavy cloth", "a jar of something sweet",
		"another set of matched cups", "a life-size statue of a stranger",
		"a box of small carved animals", "three barrels of oil",
		"a bundle of dried flowers", "a heavy stone bowl",
		"a sack of seed nobody recognises", "a cage of white rabbits",
		"a set of ceremonial knives", "an enormous woven basket",
		"a chest of foreign coins", "a portrait of somebody important",
	},
	// Where it ended up. All noun phrases, all of them somewhere out of the way.
	"enc_cap_place": {
		"the back corner", "the room nobody uses", "the far end of the hall",
		"a cold room at the back", "the corridor outside", "the stairwell",
		"the space behind the door", "a cupboard that already sticks",
		"the corner where the light is bad", "the empty room beside the kitchen",
		"a locked room upstairs", "the alcove by the entrance",
		"the gap between two chests", "the side room with the low ceiling",
		"the landing halfway up the stairs", "the passage behind the curtain",
		"a room that used to be a pantry", "the shadow under the stairs",
	},
}

// --- the sentences -----------------------------------------------------------

// encCapacityTemplates is the EncounterAtCapacity skeleton set: one ungated pool
// that lands in any age, plus one pool per era bucket for the imagery that does
// not travel.
func encCapacityTemplates() []tmpl {
	out := pool("enc_cap_any", erasAny, encCapacityAny)
	out = append(out, pool("enc_cap_ancient", erasAncient, encCapacityAncient)...)
	out = append(out, pool("enc_cap_feudal", erasFeudal, encCapacityFeudal)...)
	out = append(out, pool("enc_cap_industrial", erasIndustrial, encCapacityIndustrial)...)
	out = append(out, pool("enc_cap_digital", erasDigital, encCapacityDigital)...)
	out = append(out, pool("enc_cap_cosmic", erasCosmic, encCapacityCosmic)...)
	return out
}

// encCapacityAny fires in every age, so nothing in it may name an era-coded
// object. Rooms, doors, chairs, lists, servants, animals and children are
// timeless; stewards, telegrams and airlocks are not.
var encCapacityAny = []skel{
	// --- where it ended up ---
	{Text: "The pile by the door has taken ~ as well", Slot: "enc_cap_gift"},
	{Text: "The newest gift has been carried straight to ~", Slot: "enc_cap_place"},
	{Text: "Nobody could think of anywhere better than ~", Slot: "enc_cap_place"},
	{Text: "Two servants spent the morning shifting ~ twice", Slot: "enc_cap_gift"},
	{Text: "The council has agreed to leave ~ alone for now", Slot: "enc_cap_gift"},
	{Text: "It has ended up in ~ with everything else", Slot: "enc_cap_place"},
	{Text: "Everything that arrived this week is in ~", Slot: "enc_cap_place"},
	{Text: "Whoever arranged the hall has hidden ~ behind a curtain", Slot: "enc_cap_gift"},
	{Text: "For the second time this month, somebody moved ~", Slot: "enc_cap_gift"},
	{Text: "A servant was told to find room for ~", Slot: "enc_cap_gift"},

	// --- who has been waiting, and for how long ---
	{Text: "The envoy has been sitting in the same chair since dawn"},
	{Text: "Four delegations are waiting and there are three chairs"},
	{Text: "The man from the south has been here eleven days"},
	{Text: "Someone has been offered water four times this afternoon"},
	{Text: "A delegation arrived at noon and has not been greeted"},
	{Text: "The waiting room has a smell to it now"},
	{Text: "Three parties are waiting in one room and ignoring each other"},

	// --- a visitor with nothing to do ---
	{Text: "An envoy has learned the names of all the door servants"},
	{Text: "The visiting minister has taken up carving, badly"},
	{Text: "Their second man has started drawing the ceiling"},
	{Text: "Two envoys have begun playing a game with pebbles"},
	{Text: "The tall visitor has read every name on the wall"},
	{Text: "Somebody's guest has been asleep since the middle of the morning"},
	{Text: "One of the visitors has taken to eating with the servants"},
	{Text: "The envoy who came in winter now dresses like everyone else"},

	// --- the household ---
	{Text: "A servant has been counting the same gifts all week"},
	{Text: "The woman who sweeps has asked where she is meant to sweep"},
	{Text: "Two of the kitchen staff are refusing to go in there"},
	{Text: "The door cannot be closed while the crate stands there"},
	{Text: "Somebody must be paid to sit with the animals overnight"},
	{Text: "The cupboard has been shut with a rope"},
	{Text: "The children have found ~ and are not being stopped", Slot: "enc_cap_gift"},

	// --- the same thing, again ---
	{Text: "This is the fourth identical bowl and nobody has said so"},
	{Text: "The two chests are the same chest, more or less"},
	{Text: "There are now six of a thing nobody wanted one of"},
	{Text: "Someone noticed the pattern matches the one from last month"},

	// --- livestock nobody asked for ---
	{Text: "The white bird has been given to a child to hold"},
	{Text: "Two goats have been tied outside and are eating the flowers"},
	{Text: "Nobody has decided who is responsible for the large cat"},
	{Text: "The animal arrived with a letter of introduction and a name"},
	{Text: "Something in a crate has begun making noise at night"},

	// --- things going off ---
	{Text: "A basket of something has gone soft in ~", Slot: "enc_cap_place"},
	{Text: "The fruit that came first has stopped being fruit"},
	{Text: "The smell in the side room has been reported twice"},
	{Text: "Nobody opened the sealed jars. Now nobody will"},
	{Text: "There is dust on ~ already, one day after arrival", Slot: "enc_cap_gift"},

	// --- nothing to translate ---
	{Text: "The translator has not been needed for nine days"},
	{Text: "Two interpreters have taught each other their own languages"},
	{Text: "The man who speaks four tongues has been given a broom"},

	// --- the wrong name on the page ---
	{Text: "The name on the second document is spelled two ways"},
	{Text: "Someone wrote the wrong title and it has been copied since"},
	{Text: "The list has a name on it that nobody can pronounce"},

	// --- seating ---
	{Text: "There has been an argument about who sits nearer the fire"},
	{Text: "The seating was settled at midday and undone by evening"},
	{Text: "Two delegations refuse to be seated on the same side"},

	// --- back where it came from ---
	{Text: "The gift given last month has come back with a new label"},
	{Text: "Somebody recognised their own mark on the underside"},
	{Text: "A gift went out this morning and came back by evening"},

	// --- it will not fit ---
	{Text: "The thing in the entrance will not go through the doorway"},
	{Text: "Four men measured the door and then measured it again"},
	{Text: "They have taken the hinges off and it still will not pass"},

	// --- the queue and the calendar ---
	{Text: "There is a queue outside and it has doubled since morning"},
	{Text: "The next free day on the schedule is in autumn"},
	{Text: "A child has been playing with something valuable for an hour"},

	// --- kind: mercantile ---
	{Text: "The traders came with samples and left with the samples", Kinds: []string{"mercantile"}},
	{Text: "A price was mentioned twice and ignored both times", Kinds: []string{"mercantile"}},
	{Text: "The merchants have set up in the courtyard to wait", Kinds: []string{"mercantile"}},
	{Text: "Somebody was offered a discount for taking two of them", Kinds: []string{"mercantile"}},

	// --- kind: peaceful ---
	{Text: "They brought flowers, which will last about four days", Kinds: []string{"peaceful"}},
	{Text: "The visitors have been very patient and remain very patient", Kinds: []string{"peaceful"}},
	{Text: "A treaty was read aloud to a mostly empty room", Kinds: []string{"peaceful"}},

	// --- kind: isolationist ---
	{Text: "They came a long way to stand in a doorway", Kinds: []string{"isolationist"}},
	{Text: "The delegation refused food and has been standing all day", Kinds: []string{"isolationist"}},
	{Text: "Nobody expected them and nobody knows how to seat them", Kinds: []string{"isolationist"}},

	// --- tone ---
	{Text: "The court is now the largest owner of ceremonial bowls", Tones: []Tone{Wry}},
	{Text: "Three separate peoples have gifted the same species of bird", Tones: []Tone{Wry}},
	{Text: "A room has been named after the things stored in it", Tones: []Tone{Wry}},
	{Text: "The doorman has become the most powerful person here", Tones: []Tone{Wry}},

	// --- the visitor's name, in frames a singular-or-plural name survives ---
	{Text: "A servant announced {subject} to a room already full", Needs: needSubject},
	{Text: "Somebody folded the note from {subject} into a fan", Needs: needSubject},
	{Text: "Nobody has explained to {subject} how long the wait runs", Needs: needSubject},
	{Text: "The seating plan puts {subject} behind a pillar", Needs: needSubject},
	{Text: "Everything sent by {subject} went to the same corner", Needs: needSubject},
	{Text: "An interpreter was found for {subject} and then sent away", Needs: needSubject},
}

// encCapacityAncient — the elders, the store-pit, hides drying for guests who
// will not leave.
var encCapacityAncient = []skel{
	{Text: "The elders have run out of ways to say wait"},
	{Text: "Two more hides were laid on the pile by the store-pit"},
	{Text: "A herd was driven in as a gift and has eaten the grass"},
	{Text: "Somebody's hut now holds three carved things and one sleeping man"},
	{Text: "The spears they sent are stacked against the wrong wall"},
	{Text: "Visitors from beyond the ford have been waiting since the last rain"},
	{Text: "The drums were sounded for the first three arrivals only"},
	{Text: "A watch-fire was kept burning for guests who came a day early"},
	{Text: "The store-pit is full of things nobody intends to eat"},
	{Text: "An arrow was given as a token and has been lost"},
	{Text: "The people from the valley brought stone and stayed to watch it"},
	{Text: "Smoke on the far ridge means another lot are coming"},
	{Text: "The elders argued about seating until the fire went out"},
	{Text: "Three parties came up the road together and will not separate"},
	{Text: "The harvest was left standing while everyone dealt with visitors"},
	{Text: "A spear was planted outside the wrong hut again"},
	{Text: "Two herds were promised and only one of them arrived"},
	{Text: "Somebody has been drying hides for guests who never leave"},
	{Text: "The spring below the camp is being used by strangers"},
	{Text: "An old woman has been given the visitors to mind"},
	{Text: "The hut set aside for guests has guests in it already"},
	{Text: "They came down the valley in a column and stopped short"},
	{Text: "The elders sent a boy out to count who is waiting"},
	{Text: "A hide painted with strangers' marks hangs where nobody looks"},
	{Text: "The ford was crossed by four parties in two days"},
	{Text: "Someone traded a gift away for a spear and said nothing"},
	{Text: "The drums stopped. The waiting did not"},
	{Text: "Guests have been sleeping under the hides meant for winter"},
}

// encCapacityFeudal — the steward, the gate, carts in the yard, heralds arguing
// about who speaks first.
var encCapacityFeudal = []skel{
	{Text: "The steward has stopped opening the gate for new arrivals"},
	{Text: "Four carts of gifts are standing in the yard unloaded"},
	{Text: "A herald has been waiting on the steps since the bells rang"},
	{Text: "The clerks have written the same welcome six times this month"},
	{Text: "Two riders arrived at supper and were shown to a bench"},
	{Text: "The undercroft holds nothing but presents and one furious cat"},
	{Text: "Somebody sealed the wrong parchment with the wrong wax"},
	{Text: "A banner was hung and taken down within the hour"},
	{Text: "The quartermaster has refused to sign for any more of it"},
	{Text: "Wagons keep arriving and the gate is being left open"},
	{Text: "Sacks of foreign grain have been stacked against the storehouse wall"},
	{Text: "The bells rang for the third delegation and nobody looked up"},
	{Text: "A muster was called off so the yard could be cleared"},
	{Text: "The village has begun charging the waiting parties for bread"},
	{Text: "Somebody's cart blocked the gate for most of the afternoon"},
	{Text: "The clerk who keeps the list has asked to be moved"},
	{Text: "Two heralds arrived together and argued about who speaks first"},
	{Text: "A horn was blown for a delegation that had already left"},
	{Text: "The militia have been used to carry chests up stairs"},
	{Text: "Parchment has run short because of all the polite refusals"},
	{Text: "The road to the gate is lined with waiting wagons"},
	{Text: "An envoy has been given the steward's own room to sleep in"},
	{Text: "The pickets counted eleven strangers on the road before noon"},
	{Text: "Somebody regifted a cup and it came back within the month"},
	{Text: "The yard smells of horses that belong to nobody here"},
	{Text: "Supper was served twice. Neither party would share a table"},
	{Text: "The harvest carts cannot get in past the visiting ones"},
	{Text: "A boy has been paid to say the hall is being cleaned"},
}

// encCapacityIndustrial — the depot, the annexe, freight on the platform, a
// telegram that arrived after the people it announced.
var encCapacityIndustrial = []skel{
	{Text: "The telegram announcing the delegation arrived after the delegation"},
	{Text: "Freight from three separate consulates is stacked in the annexe"},
	{Text: "The foreman has been asked to make space and has declined"},
	{Text: "Two crates of gifts have sat on the platform all week"},
	{Text: "The depot office has run out of forms for visitors"},
	{Text: "Somebody telephoned ahead and nobody wrote it down"},
	{Text: "The warehouse manager has stopped answering the telephone"},
	{Text: "A whole siding has been given over to somebody's ceremonial engine"},
	{Text: "The visitors' train was met by one man with an umbrella"},
	{Text: "Everything has been recorded in triplicate and filed nowhere useful"},
	{Text: "The payroll now includes two men who only move boxes"},
	{Text: "Shareholders have been told the reception went extremely well"},
	{Text: "A wire arrived asking whether the last one arrived"},
	{Text: "The lorries came at six and are still in the yard"},
	{Text: "Three delegations are booked into the same afternoon"},
	{Text: "Somebody has painted a number on the door of the annexe"},
	{Text: "The telegraph office is sending more apologies than anything else"},
	{Text: "A porter has been carrying the same chest back and forth"},
	{Text: "The waiting room has been repainted twice and refurnished once"},
	{Text: "Smoke from the visitors' cigars has filled the upper corridor"},
	{Text: "Two interpreters are on the payroll and neither has spoken today"},
	{Text: "The freight elevator has been reserved for gifts until Friday"},
	{Text: "A brass band was hired and then quietly stood down"},
	{Text: "The road outside the depot is blocked by parked vehicles"},
	{Text: "Nobody has opened the crate marked with foreign lettering"},
	{Text: "The company photographer has been booked for a fifth arrival"},
	{Text: "A visiting official has learned to work the goods lift"},
	{Text: "A leak in the annexe roof has reached the stacked crates"},
}

// encCapacityDigital — the calendar, the lobby, badges, a drone delivering
// flowers to a floor nobody works on.
var encCapacityDigital = []skel{
	{Text: "The delegation has been added to a calendar that has no gaps"},
	{Text: "Four ambassadors are sitting in a lobby designed for two"},
	{Text: "The visitor badges have run out for the second time"},
	{Text: "A drone delivered flowers to a floor nobody works on"},
	{Text: "The channel for foreign relations has muted itself"},
	{Text: "Somebody left the uplink open through an entire welcome speech"},
	{Text: "Two analysts have been assigned to read the same greeting"},
	{Text: "The gift closet on the third floor will not shut"},
	{Text: "An after-action note on the last visit was never written"},
	{Text: "The network flagged the fourteenth identical invitation as spam"},
	{Text: "A translation service has been running all morning on an empty room"},
	{Text: "The conference room has been booked out until next quarter"},
	{Text: "Somebody's assistant has been rescheduling the same meeting since March"},
	{Text: "Three envoys are waiting in reception and one has fallen asleep"},
	{Text: "The feed from the lobby camera shows a very long queue"},
	{Text: "A crate of ceremonial hardware is under a desk downstairs"},
	{Text: "The office manager has begun turning arrivals away politely"},
	{Text: "Nobody knows whose responsibility the visiting choir has become"},
	{Text: "The parking outside has been taken by diplomatic vehicles"},
	{Text: "An analyst has been photographing the gifts for insurance purposes"},
	{Text: "Two drones are circling the roof waiting for landing clearance"},
	{Text: "The intercom has been used four times to ask about seating"},
	{Text: "Somebody printed the wrong name on the welcome screen again"},
	{Text: "The storage room now has a spreadsheet of its own"},
	{Text: "A delegation has been shown the roof terrace instead"},
	{Text: "The road outside is closed for a motorcade nobody requested"},
	{Text: "Four identical baskets arrived from four different embassies"},
	{Text: "The security desk has stopped logging deliveries. Nobody objected"},
}

// encCapacityCosmic — the dock schedule, the manifest, a gift left in vacuum
// long enough to become a different shape.
var encCapacityCosmic = []skel{
	{Text: "The forward bay has been given over to diplomatic gifts"},
	{Text: "Three freighters are holding at the outer marker awaiting a berth"},
	{Text: "The dock schedule has been rewritten twice this shift"},
	{Text: "A delegation has been in the transit lounge for nine hours"},
	{Text: "Something on the manifest is listed only as ceremonial"},
	{Text: "The airlock queue has become a matter of protocol"},
	{Text: "Two visiting envoys have settled into a maintenance corridor"},
	{Text: "Nobody has claimed the crate strapped behind the bulkhead panel"},
	{Text: "The station relay has forty unanswered greetings on it"},
	{Text: "A gift was left in vacuum and is now a different shape"},
	{Text: "The reactor deck is warmer, so the perishables went there"},
	{Text: "Four transponders are broadcasting the same friendly message"},
	{Text: "An ambassador has memorised the whole station safety briefing"},
	{Text: "The hull outside the visitors' berth has been polished twice"},
	{Text: "Somebody welded a shelf to a bulkhead to hold the gifts"},
	{Text: "The dock crew have stopped asking who anyone is"},
	{Text: "A live animal came aboard and has been given its own cabin"},
	{Text: "Two delegations in orbit are refusing to descend in order"},
	{Text: "The manifest lists ninety crates and describes none of them"},
	{Text: "An orbital station meant for sixty is holding rather more"},
	{Text: "The airlock cycled for a delegation that changed its mind"},
	{Text: "Somebody's ceremonial statue does not fit through the hatch"},
	{Text: "The relay logs show the same welcome sent eleven times"},
	{Text: "A freighter captain has been waiting behind three diplomatic ships"},
	{Text: "The station gardens have been closed for a reception since Tuesday"},
	{Text: "Rations have been swapped for foreign delicacies nobody enjoys"},
	{Text: "The visiting party has been assigned bunks in the old dock office"},
	{Text: "A hologram of somebody's founder stands in a service corridor"},
}
