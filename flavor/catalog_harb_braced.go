package flavor

// HarbingerBraced — the player spent resources to soften the blow if it comes.
// The log line says what was spent; these sentences are the digging, stacking,
// sealing and sandbagging, and the arguments about who got the heavy end.
//
// The per-age pools echo config.HarbingerDef.BraceLabel for that age.

func harbBracedTemplates() []tmpl {
	out := pool("harb_braced_any", erasAny, harbBracedAny)
	out = append(out, pool("harb_braced_grounded", erasGrounded, harbBracedGrounded)...)
	out = append(out, pool("harb_braced_late", erasLate, harbBracedLate)...)
	out = append(out, pool("harb_braced_ancient", erasAncient, harbBracedAncient)...)
	out = append(out, pool("harb_braced_feudal", erasFeudal, harbBracedFeudal)...)
	out = append(out, pool("harb_braced_industrial", erasIndustrial, harbBracedIndustrial)...)
	out = append(out, pool("harb_braced_digital", erasDigital, harbBracedDigital)...)
	out = append(out, pool("harb_braced_cosmic", erasCosmic, harbBracedCosmic)...)
	out = append(out, agePools("harb_braced_age", harbBracedAges)...)
	return out
}

var harbBracedAny = []skel{
	{Text: "Nobody walked anywhere empty-handed", Reg: rPlain, Topic: "haul"},
	{Text: "The work went on all night", Reg: rPlain, Topic: "work"},
	{Text: "The strongest things were put in front of the weakest", Reg: rPlain, Topic: "building"},
	{Text: "Every pair of hands was put to use, including some that were more trouble than help", Reg: rWry, Topic: "people"},
	{Text: "Lists were made of what would be needed first", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The old argued with the young about how it had been done last time", Reg: rWry, Form: fOverheard, Topic: "argument"},
	{Text: "Children were given small jobs to keep them out of the way", Reg: rPlain, Topic: "family"},
	{Text: "All through the preparations {subject} gave no advice at all", Needs: needSubject, Reg: rWry, Topic: "stranger"},
	{Text: "People who had never lifted anything heavier than a cup found themselves carrying loads down to the stores and back again until their hands blistered", Reg: rPlain, Topic: "wound"},
	{Text: "By the end there was nothing left to do but wait, and nobody was good at that", Reg: rWry, Topic: "time"},
	{Text: "It was the kind of work that goes on long after anyone can remember why it started, so that by the third day people were stacking things out of habit and had to be stopped by someone who could still count", Reg: rWry, Topic: "work"},
	{Text: "The grumbling was about who had been given the heavy end", Reg: rWry, Form: fComplaint, Topic: "argument"},
	{Text: "Orders went round that nothing was to be wasted", Reg: rPlain, Form: fNotice, Topic: "authority"},
}

var harbBracedGrounded = []skel{
	{Text: "Roofs were weighted down with rocks", Reg: rPlain, Topic: "building"},
	{Text: "The animals were brought in close", Reg: rPlain, Topic: "animal"},
	{Text: "Water was stored in everything that would hold it", Reg: rPlain, Topic: "ground"},
	{Text: "Blankets were piled in the middle of every floor", Reg: rPlain, Topic: "sleep"},
	{Text: "Every spade in the place was in use from morning to night, and the ones who had no spade dug with their hands", Reg: rPlain, Topic: "work"},
	{Text: "The dogs followed the work from place to place and got under everyone's feet, and nobody had the heart to shut them in", Reg: rWry, Topic: "animal"},
	{Text: "The whole settlement moved its food and its seed and its old people into the strongest building it had, and then stood looking at the strongest building it had and wondering whether it was strong enough, and then went back for more stone", Reg: rWry, Topic: "building"},
	{Text: "The mud got into everything", Reg: rPlain, Form: fComplaint, Topic: "ground"},
}

var harbBracedLate = []skel{
	{Text: "Backups ran all night", Reg: rPlain, Topic: "machine"},
	{Text: "Redundant systems came online", Reg: rPlain, Topic: "machine"},
	{Text: "Every spare battery was charged and stacked", Reg: rPlain, Topic: "kit"},
	{Text: "Emergency rations were issued to every residence", Reg: rPlain, Form: fNotice, Topic: "food"},
	{Text: "The safe rooms were stocked and checked", Reg: rPlain, Topic: "building"},
	{Text: "Everything was copied to three separate sites", Reg: rPlain, Form: fLedger, Topic: "machine"},
	{Text: "Automated systems worked through the preparations faster than people could follow them, and people followed anyway, checking what the machines had already checked", Reg: rWry, Topic: "machine"},
	{Text: "In the last hours every nonessential process in the settlement was shut down to free power for the ones that would matter, and the lights dimmed floor by floor until the place felt like it was holding its breath", Reg: rPlain, Topic: "building"},
}

var harbBracedAncient = []skel{
	{Text: "The herd was penned", Reg: rPlain, Topic: "animal"},
	{Text: "Pits were dug and filled", Reg: rPlain, Topic: "ground"},
	{Text: "Stones were rolled across the cave mouths", Reg: rPlain, Topic: "building"},
	{Text: "Every hide in the camp was stretched over the huts and weighted down", Reg: rPlain, Topic: "building"},
	{Text: "Dried meat was stacked in the deepest pit and covered with earth", Reg: rPlain, Topic: "food"},
	{Text: "The elders marked out where each family would go when it came", Reg: rPlain, Topic: "family"},
	{Text: "Children gathered firewood until there was none left within a day's walk", Reg: rPlain, Topic: "work"},
	{Text: "The men dug a ditch round the whole camp, which took four days and a good deal of shouting, and filled it with thorn", Reg: rWry, Topic: "ground"},
	{Text: "Water was carried up from the spring in skins and bowls and anything that would hold it, and poured into a pit lined with clay", Reg: rPlain, Topic: "ground"},
	{Text: "The women took the smallest children and the oldest people up to the high cave with enough food for a moon, and the men walled the entrance behind them with stones and left a gap at the top the width of a hand", Reg: rPlain, Topic: "family"},
	{Text: "The hunters complain that nobody is hunting", Reg: rWry, Form: fComplaint, Topic: "food"},
	{Text: "The elders say nothing is to be eaten from the pits until they give the word", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Four pits of meat, two of roots, one of seed", Reg: rPlain, Form: fLedger, Topic: "count"},
}

var harbBracedFeudal = []skel{
	{Text: "The granary was filled", Reg: rPlain, Topic: "food"},
	{Text: "The walls were shored up", Reg: rPlain, Topic: "building"},
	{Text: "Barrels in every cellar", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "The masons worked by torchlight", Reg: rPlain, Topic: "work"},
	{Text: "Carts brought stone from the quarry all week", Reg: rPlain, Topic: "haul"},
	{Text: "The castle cellars were cleared to make room for grain", Reg: rPlain, Topic: "food"},
	{Text: "Every well in the town was covered with planks", Reg: rPlain, Topic: "ground"},
	{Text: "The lord ordered every household to lay in a month of flour and to send one man a day to work on the walls, and the carts ran day and night", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Timber was brought down from the forest faster than it could be seasoned, and the carpenters complained about it the whole time they were using it", Reg: rWry, Form: fComplaint, Topic: "building"},
	{Text: "The cathedral treasure was carried down into the crypt and bricked up behind a false wall, and the three men who did the bricking were made to swear on the relics never to say where, and one of them has been drinking ever since", Reg: rWry, Topic: "religion"},
	{Text: "The steward has ordered every cellar in the town inspected", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Two hundred barrels of salt fish in the undercroft", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The masons say the wall will hold, and they say it without looking at the wall", Reg: rWry, Form: fOverheard, Topic: "building"},
}

var harbBracedIndustrial = []skel{
	{Text: "Sandbags went up", Reg: rPlain, Topic: "building"},
	{Text: "Coal was stockpiled", Reg: rPlain, Topic: "haul"},
	{Text: "Hospitals cleared their wards and waited", Reg: rPlain, Topic: "wound"},
	{Text: "Cellars up and down the street were cleared and whitewashed", Reg: rPlain, Topic: "building"},
	{Text: "The mills were shut and their windows boarded", Reg: rPlain, Topic: "work"},
	{Text: "Tinned food was bought by the crate", Reg: rPlain, Topic: "food"},
	{Text: "The fire brigade drilled in the square every morning", Reg: rPlain, Topic: "town"},
	{Text: "Volunteers filled sandbags on the embankment from first light until the lamps came on, and the sandbags went up along the river wall in a line a mile long", Reg: rPlain, Topic: "people"},
	{Text: "The railway company ran extra trains to carry coal and flour into the city, and the trains came in so heavily loaded the platforms creaked", Reg: rPlain, Topic: "machine"},
	{Text: "The pawnshops did a roaring trade as families turned whatever they had into tins and candles and rope, and the pawnbroker on the corner, who had seen a few of these, bought nothing he could not sell back to them afterwards", Reg: rWry, Topic: "money"},
	{Text: "The ratepayers complain about the cost of the sandbags", Reg: rWry, Form: fComplaint, Topic: "money"},
	{Text: "Householders are advised to keep a bucket of sand on every landing", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Ten thousand sandbags filled, and four thousand placed", Reg: rPlain, Form: fLedger, Topic: "count"},
}

var harbBracedDigital = []skel{
	{Text: "Generators were fuelled", Reg: rPlain, Topic: "machine"},
	{Text: "Drones ferried medicine all night", Reg: rPlain, Topic: "wound"},
	{Text: "The server centres went to island mode", Reg: rPlain, Topic: "machine"},
	{Text: "Critical systems were mirrored to three continents", Reg: rPlain, Topic: "machine"},
	{Text: "Everyone downloaded everything they might need offline", Reg: rWry, Topic: "people"},
	{Text: "The water utility filled every reservoir to the brim", Reg: rPlain, Topic: "ground"},
	{Text: "The supermarkets restocked overnight and emptied by noon", Reg: rPlain, Topic: "trade"},
	{Text: "Offices sent staff home with laptops and instructions to keep them charged and, for once, to write their passwords down on paper", Reg: rWry, Form: fNotice, Topic: "work"},
	{Text: "Hospitals postponed everything that could be postponed and filled their corridors with beds, and the beds sat empty and made up, waiting", Reg: rPlain, Topic: "wound"},
	{Text: "In the towers the building management system sealed the lower levels floor by floor, running its checks in a calm voice over the speakers, and residents stood in the corridors with their go-bags listening to it count down the doors", Reg: rPlain, Topic: "building"},
	{Text: "The comments complain that the drills are making everyone late", Reg: rWry, Form: fComplaint, Topic: "time"},
	{Text: "Residents are asked to keep their devices charged and their go-bags by the door", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Three days of water per person, allocated by app", Reg: rPlain, Form: fLedger, Topic: "count"},
}

var harbBracedCosmic = []skel{
	{Text: "The shields were doubled", Reg: rPlain, Topic: "machine"},
	{Text: "Bulkheads were sealed", Reg: rPlain, Topic: "building"},
	{Text: "Every airlock on the station was cycled and checked", Reg: rPlain, Topic: "machine"},
	{Text: "The outer ring was stripped of anything that could be moved", Reg: rPlain, Topic: "haul"},
	{Text: "Reserve air was pumped into every tank", Reg: rPlain, Topic: "kit"},
	{Text: "Hull plating was welded over the viewports", Reg: rPlain, Topic: "building"},
	{Text: "Crews worked in pressure suits for three shifts straight, reinforcing the hull from outside, and came in with their faces grey and went straight back out", Reg: rPlain, Topic: "work"},
	{Text: "The seed vault was moved to the station's core, where it would be the last thing to go", Reg: rPlain, Topic: "food"},
	{Text: "Every ship in dock was fuelled and crewed and kept ready to leave, and people who walked past the docking bays on their way to work looked at the lit hatches and did the arithmetic of how many berths there were and how many people", Reg: rPlain, Topic: "count"},
	{Text: "Hydroponics complains that the shield load is starving their lights", Reg: rWry, Form: fComplaint, Topic: "food"},
	{Text: "All crew are to carry a pressure suit at all times until further notice", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Shields at full, reserve air at full, morale at half", Reg: rJoke, Form: fLedger, Topic: "count"},
	{Text: "The station hummed with load", Reg: rPlain, Topic: "noise"},
}

// harbBracedAges echoes each age's BraceLabel.
var harbBracedAges = []ageSet{
	{ages("primitive_age"), []skel{
		{Text: "Every family dug a pit and filled it", Reg: rPlain, Topic: "ground"},
		{Text: "Everyone hid a little extra", Reg: rWry, Topic: "haul"},
	}},
	{ages("stone_age"), []skel{
		{Text: "The cave mouth was walled to the height of a man", Reg: rPlain, Topic: "building"},
		{Text: "They walled up the cave mouth with the biggest stones they could move and packed the gaps with moss and mud, and the hermit watched them do it", Reg: rPlain, Topic: "work"},
	}},
	{ages("bronze_age"), []skel{
		{Text: "The grain went into jars sealed with pitch and clay", Reg: rPlain, Topic: "food"},
		{Text: "Potters worked through the night", Reg: rPlain, Topic: "work"},
	}},
	{ages("iron_age"), []skel{
		{Text: "The walls went up a course higher, then another", Reg: rPlain, Topic: "building"},
		{Text: "Every man in the city carried stone for the walls for a week, and the priests carried stone too, which was remarked on", Reg: rWry, Topic: "religion"},
	}},
	{ages("classical_age"), []skel{
		{Text: "The citadel was stocked with oil, grain and wine for a year", Reg: rPlain, Form: fLedger, Topic: "food"},
		{Text: "Amphorae filled the citadel's cellars", Reg: rPlain, Topic: "haul"},
	}},
	{ages("medieval_age"), []skel{
		{Text: "Timber props went up against every weak stretch of wall", Reg: rPlain, Topic: "building"},
		{Text: "The masons found three places where the wall had been hollow for a century and filled them, and said it was a mercy nobody had leaned on it", Reg: rWry, Form: fOverheard, Topic: "building"},
	}},
	{ages("renaissance_age"), []skel{
		{Text: "The bastions were packed with stores", Reg: rPlain, Topic: "haul"},
		{Text: "The engineers who designed the new bastions came back to see them used, and walked along the top making notes, which annoyed the soldiers", Reg: rWry, Topic: "people"},
	}},
	{ages("colonial_age"), []skel{
		{Text: "Powder and flour went into the magazine", Reg: rPlain, Topic: "haul"},
		{Text: "The powder was kept dry and the flour was kept separate, after what happened last time", Reg: rWry, Form: fOverheard, Topic: "weapon"},
	}},
	{ages("industrial_age"), []skel{
		{Text: "The mills were braced with iron girders", Reg: rPlain, Topic: "building"},
		{Text: "Engineers went through every mill with chalk, marking the walls that would come down first, and the chalk marks were still there years later", Reg: rPlain, Topic: "work"},
	}},
	{ages("victorian_age"), []skel{
		{Text: "Tinned goods and sandbags, in every hall", Reg: rPlain, Form: fLedger, Topic: "haul"},
		{Text: "The grocers were emptied of tins by Tuesday", Reg: rPlain, Topic: "trade"},
	}},
	{ages("electric_age"), []skel{
		{Text: "Emergency cables ran down every main street", Reg: rPlain, Topic: "machine"},
		{Text: "The electricians worked through the night wiring spare lamps into the hospitals and the telephone exchanges, and at dawn the whole city flickered once as they tested it", Reg: rPlain, Topic: "machine"},
	}},
	{ages("atomic_age"), []skel{
		{Text: "The fallout shelters were stocked with ~", Slot: "harb_shelter_industrial", Reg: rPlain, Topic: "haul"},
		{Text: "Families carried their own supplies down into the shelter under the town hall, a tin of this and a jar of that, and the shelter marshal wrote every item in a book and stacked it along the wall by type", Reg: rPlain, Form: fLedger, Topic: "count"},
	}},
	{ages("modern_age"), []skel{
		{Text: "Substations were ringed with sandbags", Reg: rPlain, Topic: "machine"},
		{Text: "Crews reinforced the transmission towers across the whole region", Reg: rPlain, Topic: "work"},
	}},
	{ages("information_age"), []skel{
		{Text: "Everything was backed up to tape", Reg: rPlain, Topic: "machine"},
		{Text: "The tape drives ran all night in the basement, and in the morning there were shelves of cartridges labelled in marker pen and nobody entirely sure what was on them", Reg: rWry, Topic: "machine"},
	}},
	{ages("digital_age"), []skel{
		{Text: "The bottled water was gone from every shop by noon", Reg: rPlain, Topic: "trade"},
		{Text: "Pallets of water went into every spare room", Reg: rPlain, Topic: "haul"},
	}},
	{ages("cyberpunk_age"), []skel{
		{Text: "The grid was cut off from the net", Reg: rPlain, Topic: "machine"},
		{Text: "Technicians pulled the cables out by hand, one by one, until the grid was sealed off from the net and running on its own, deaf and blind and safe", Reg: rPlain, Topic: "machine"},
	}},
	{ages("fusion_age"), []skel{
		{Text: "Every spare watt went to containment", Reg: rPlain, Topic: "machine"},
		{Text: "The containment fields hummed at a pitch that set teeth on edge", Reg: rPlain, Topic: "noise"},
	}},
	{ages("space_age"), []skel{
		{Text: "The orbital shields were hardened", Reg: rPlain, Topic: "machine"},
		{Text: "Shield crews worked outside in shifts, bolting reinforcing plates to the orbital shields, and the station's lights dimmed every time the new sections came online", Reg: rPlain, Topic: "work"},
	}},
	{ages("interstellar_age"), []skel{
		{Text: "The outposts were pulled back inside the shield", Reg: rPlain, Topic: "border"},
		{Text: "The last outpost to come in was a mining crew who had been out for three years and did not want to leave, and they came in anyway, with their samples", Reg: rWry, Topic: "people"},
	}},
	{ages("galactic_age"), []skel{
		{Text: "The fleet folded into the nebula", Reg: rPlain, Topic: "machine"},
		{Text: "The fleet slipped into the nebula one ship at a time, running dark, and the station watched them go until the last drive glow faded into the gas", Reg: rPlain, Topic: "machine"},
	}},
	{ages("quantum_age"), []skel{
		{Text: "The timeline was branched, and the branches hedged", Reg: rPlain, Topic: "time"},
		{Text: "Your future self had suggested hedging, and you branched the timeline three ways, and now there are versions of you preparing for things this version will never see", Reg: rWry, Topic: "time"},
	}},
	{ages("transcendent_age"), []skel{
		{Text: "You anchored yourself to this branch", Reg: rPlain, Topic: "time"},
		{Text: "You held onto this reality the way you would hold onto a railing in a storm, knowing that the other you had held onto theirs just as hard, and that it had not been enough for them", Reg: rPlain, Topic: "omen"},
	}},
}
