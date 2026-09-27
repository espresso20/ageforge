package flavor

// HarbingerArrival — the age's harbinger has turned up. The log line above
// this one names the figure; these sentences are what the settlement saw,
// smelled and said while it happened, and the one detail about the figure a
// person would still mention a year later.
//
// The speaker changes every age, so most of the figure's own description lives
// in the per-age pools. The ungated pool and the era pools are about the
// settlement, because "he walked in from the dry country" is wrong for a chain
// email and "it arrived in forty inboxes" is wrong for an oracle.

func harbArrivalTemplates() []tmpl {
	out := pool("harb_arrival_any", erasAny, harbArrivalAny)
	out = append(out, pool("harb_arrival_grounded", erasGrounded, harbArrivalGrounded)...)
	out = append(out, pool("harb_arrival_late", erasLate, harbArrivalLate)...)
	out = append(out, pool("harb_arrival_ancient", erasAncient, harbArrivalAncient)...)
	out = append(out, pool("harb_arrival_feudal", erasFeudal, harbArrivalFeudal)...)
	out = append(out, pool("harb_arrival_industrial", erasIndustrial, harbArrivalIndustrial)...)
	out = append(out, pool("harb_arrival_digital", erasDigital, harbArrivalDigital)...)
	out = append(out, pool("harb_arrival_cosmic", erasCosmic, harbArrivalCosmic)...)
	out = append(out, agePools("harb_arrival_age", harbArrivalAges)...)
	return out
}

// harbArrivalAny fires in every age, so it describes the settlement and never
// the figure's body, clothes or medium.
var harbArrivalAny = []skel{
	{Text: "The children noticed first", Reg: rPlain, Topic: "family"},
	{Text: "Work stopped for the afternoon", Reg: rPlain, Topic: "work"},
	{Text: "The talk that evening was of nothing else", Reg: rPlain, Form: fOverheard, Topic: "rumour"},
	{Text: "An ordinary afternoon", Reg: rPlain, Topic: "time"},
	{Text: "Half the settlement came out to get a look at {subject}", Needs: needSubject, Reg: rPlain, Topic: "people"},
	{Text: "Everybody knows what it means when one of these turns up", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "Several people packed a bag that night and unpacked it in the morning, feeling foolish", Reg: rWry, Topic: "family"},
	{Text: "Three people said they had dreamed about {subject} the week before, and one of them had told people at the time, which has made her briefly important", Needs: needSubject, Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "An old woman said she had seen one of these before, as a girl, and that the last one had been right", Reg: rPlain, Form: fOverheard, Topic: "people"},
	{Text: "It is the sort of thing the old stories open with", Reg: rWry, Topic: "rumour"},
	{Text: "The first thing anybody wrote down about {subject} was the time of day", Needs: needSubject, Reg: rPlain, Form: fLedger, Topic: "time"},
	{Text: "Sleep was short", Reg: rPlain, Topic: "sleep"},
	{Text: "Two brothers who had not spoken in a year stood side by side to watch {subject} arrive, and afterwards neither of them mentioned it", Needs: needSubject, Reg: rWry, Topic: "family"},
	{Text: "Everyone found a reason to be outdoors", Reg: rWry, Topic: "people"},
	{Text: "Old Tamsin, who had been saying for a year that something was coming, was insufferable by nightfall, and her brother-in-law, who had spent the same year saying nothing ever would, went to bed early", Reg: rWry, Topic: "argument"},
	{Text: "The grumbling started before {subject} had finished arriving, mostly about the timing", Needs: needSubject, Reg: rJoke, Form: fComplaint, Topic: "argument"},
	{Text: "Mothers kept counting their children without meaning to, and the children let them", Reg: rPlain, Topic: "family"},
}

// harbArrivalGrounded is true of any age with lanes, fires and a sky people
// look at: the Stone Age to the Modern Age.
var harbArrivalGrounded = []skel{
	{Text: "The dogs howled until the small hours", Reg: rPlain, Topic: "animal"},
	{Text: "Doors were barred that night that had not been barred in years", Reg: rPlain, Topic: "building"},
	{Text: "At the end of the lane a woman who had lived there forty years began to pack her pots into a blanket, slowly, stopping to look at each one", Reg: rPlain, Topic: "family"},
	{Text: "Candles in every window", Reg: rPlain, Topic: "town"},
	{Text: "The crows came in from the fields at noon, which is the wrong time for crows", Reg: rWry, Topic: "omen"},
	{Text: "An old man who had lived through the last bad time carried his chair out into the road and sat in it with his hands on his knees, watching the sky, and would not be brought indoors to eat", Reg: rPlain, Topic: "people"},
	{Text: "Nobody would draw water alone after dark", Reg: rPlain, Topic: "ground"},
}

// harbArrivalLate is the digital and cosmic ages: screens, alerts, habitats.
var harbArrivalLate = []skel{
	{Text: "Every screen was showing it", Reg: rPlain, Topic: "machine"},
	{Text: "The alerts would not clear", Reg: rPlain, Form: fComplaint, Topic: "machine"},
	{Text: "The lights in the common areas were dimmed and left that way", Reg: rPlain, Topic: "building"},
	{Text: "Traffic on the local net tripled inside the first hour, and most of it was people asking each other whether they had seen it yet", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "For about a minute after it came, every system in the settlement froze while it tried to classify what it had just received, and the silence of the machines frightened people more than an alarm would have", Reg: rPlain, Topic: "machine"},
	{Text: "People kept their earpieces in while they slept", Reg: rPlain, Topic: "sleep"},
}

var harbArrivalAncient = []skel{
	{Text: "The drums stopped", Reg: rPlain, Topic: "noise"},
	{Text: "Ash on the wind all day", Reg: rPlain, Topic: "weather"},
	{Text: "The elders sat up late at the big fire", Reg: rPlain, Topic: "authority"},
	{Text: "The herd would not settle", Reg: rPlain, Topic: "animal"},
	{Text: "A goat was found dead in the morning with no mark on it", Reg: rPlain, Topic: "omen"},
	{Text: "Women brought the small children into the huts and sat in the doorways with their backs to the light", Reg: rPlain, Topic: "family"},
	{Text: "A boy saw a hawk carry a snake over the huts, and by evening everybody had seen it", Reg: rWry, Topic: "omen"},
	{Text: "The old men argued about whether this had happened in their grandfathers' time, settled nothing, and started again at the next meal", Reg: rWry, Form: fOverheard, Topic: "argument"},
	{Text: "The elders have forbidden the young men to make fun of it, which has made the young men worse", Reg: rJoke, Form: fNotice, Topic: "authority"},
	{Text: "The elders say nobody is to leave the camp alone until the moon turns", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Three goats, one dog and a child's cough, all put down to {subject} by the evening", Needs: needSubject, Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "The hunters say the herds moved off because of all this talk, and they say it loudly", Reg: rWry, Form: fComplaint, Topic: "animal"},
	{Text: "By the second night every family had buried something it could not afford to lose at the back of its hut, and every family knew where every other family had buried theirs, because the ground is hard and digging carries", Reg: rWry, Topic: "haul"},
	{Text: "The fires were built higher that night, and somebody sat up with each one", Reg: rPlain, Topic: "sleep"},
}

var harbArrivalFeudal = []skel{
	{Text: "The market emptied early", Reg: rPlain, Topic: "town"},
	{Text: "Masons put down their tools", Reg: rPlain, Topic: "work"},
	{Text: "The church was full by vespers", Reg: rPlain, Topic: "religion"},
	{Text: "Carts were turned round on the road outside the gate and went back the way they came", Reg: rPlain, Topic: "trade"},
	{Text: "The priest preached on the Flood without being asked to, and preached long", Reg: rWry, Topic: "religion"},
	{Text: "Every dog in the village barked at nothing through the night", Reg: rPlain, Topic: "animal"},
	{Text: "The miller sold three sacks of flour before noon and put his price up for the fourth", Reg: rWry, Form: fLedger, Topic: "trade"},
	{Text: "The steward has doubled the watch on the storehouse", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "The alewife says it is bad for trade, and she would know", Reg: rWry, Form: fComplaint, Topic: "trade"},
	{Text: "By supper the whole parish had heard, and by the next morning there were three versions of it going round the market, one of which had a comet in it that nobody had actually seen", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The bells were rung for no reason anyone would own to", Reg: rWry, Topic: "noise"},
	{Text: "A notice went up on the church door forbidding idle talk of the end of the world, and a crowd gathered to read it", Reg: rJoke, Form: fNotice, Topic: "paper"},
	{Text: "In every doorway the name of {subject} was being said by the time the lamps were lit", Needs: needSubject, Reg: rPlain, Topic: "name"},
	{Text: "The children followed {subject} as far as the churchyard wall and no further", Needs: needSubject, Reg: rPlain, Topic: "family"},
}

var harbArrivalIndustrial = []skel{
	{Text: "The shift whistle blew early", Reg: rPlain, Topic: "work"},
	{Text: "The pubs did good business", Reg: rWry, Topic: "trade"},
	{Text: "Men stood in knots on the corner reading over each other's shoulders", Reg: rPlain, Topic: "people"},
	{Text: "The foreman gave nobody the afternoon off", Reg: rWry, Form: fComplaint, Topic: "authority"},
	{Text: "Two shops on the high street had sold out of candles by teatime", Reg: rPlain, Form: fLedger, Topic: "trade"},
	{Text: "Chapel doors open all night", Reg: rPlain, Topic: "religion"},
	{Text: "The children played at the end of the world in the street until they were called in", Reg: rWry, Topic: "family"},
	{Text: "The town hall asked residents to remain calm and to stop telephoning the town hall", Reg: rJoke, Form: fNotice, Topic: "authority"},
	{Text: "The trains ran on time, which people found unsettling", Reg: rJoke, Topic: "machine"},
	{Text: "The talk in the canteen was of nothing else until the pudding came round", Reg: rWry, Form: fOverheard, Topic: "food"},
	{Text: "The insurance men were out on the doorsteps by teatime with their order books, going from house to house and doing a trade so brisk that the manager came down from head office to see it for himself", Reg: rWry, Topic: "money"},
	{Text: "REPENT was chalked on the warehouse wall by morning", Reg: rPlain, Topic: "message"},
	{Text: "The smoke from the stacks lay flat over the rooftops all day", Reg: rPlain, Topic: "weather"},
	{Text: "The landlord of the corner house put up a sign reading NO CREDIT UNTIL FURTHER NOTICE and was not ashamed of it", Reg: rWry, Form: fNotice, Topic: "money"},
}

var harbArrivalDigital = []skel{
	{Text: "Every phone in the building buzzed at once", Reg: rPlain, Topic: "machine"},
	{Text: "Nothing got done all day", Reg: rWry, Topic: "work"},
	{Text: "The servers ran hot all night", Reg: rPlain, Topic: "machine"},
	{Text: "A copy was printed out and pinned to the board in the break room, where it was read more than anything ever pinned there", Reg: rWry, Topic: "paper"},
	{Text: "The help desk logged four hundred tickets before lunch", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "The network slowed to a crawl as everyone tried to read it at once", Reg: rPlain, Topic: "machine"},
	{Text: "Management sent round a memo asking staff not to forward it", Reg: rJoke, Form: fNotice, Topic: "authority"},
	{Text: "People refreshed their feeds until their thumbs ached", Reg: rPlain, Topic: "people"},
	{Text: "The family group chat has been running without a break since lunch", Reg: rPlain, Form: fOverheard, Topic: "family"},
	{Text: "Several people deleted their apps and reinstalled them within the hour", Reg: rWry, Topic: "machine"},
	{Text: "On the ninth floor a man who had worked there for sixteen years stood up from his desk, put his coat on and went home without telling anyone, and his screen stayed lit all night", Reg: rPlain, Topic: "people"},
	{Text: "The complaints line was jammed", Reg: rPlain, Form: fComplaint, Topic: "message"},
	{Text: "The drone deliveries stopped", Reg: rPlain, Topic: "machine"},
	{Text: "Searches for {subject} outnumbered searches for everything else combined by the evening", Needs: needSubject, Reg: rPlain, Form: fLedger, Topic: "count"},
}

var harbArrivalCosmic = []skel{
	{Text: "Time felt thin", Reg: rPlain, Topic: "time"},
	{Text: "The deep antennas turned all at once", Reg: rPlain, Topic: "machine"},
	{Text: "The night watch logged it, checked the log, and logged it again", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "Crew stopped at the viewports and looked out at nothing", Reg: rPlain, Topic: "people"},
	{Text: "The hull ticked as it cooled", Reg: rPlain, Topic: "noise"},
	{Text: "The long-range relay reported a signal with no origin", Reg: rPlain, Topic: "message"},
	{Text: "It has been a long time since anything out there spoke first", Reg: rWry, Topic: "time"},
	{Text: "The station chaplain held an unscheduled service and every seat was taken", Reg: rPlain, Topic: "religion"},
	{Text: "A circular went round the decks confirming that the contact was real and asking crew to return to their duties", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "Out past the orbit of the last moon, where the dark is so complete that the instruments sometimes report their own noise as stars, something had turned toward the colony and begun, patiently, to speak", Reg: rPlain, Topic: "omen"},
	{Text: "The relay room has filed three complaints about the noise and been told to live with it", Reg: rWry, Form: fComplaint, Topic: "noise"},
	{Text: "The children on the habitat ring drew pictures of {subject} and stuck them to the viewports facing outward", Needs: needSubject, Reg: rPlain, Topic: "family"},
	{Text: "The stars looked exactly as they always had", Reg: rWry, Topic: "omen"},
	{Text: "Nothing on the scopes", Reg: rPlain, Form: fLedger, Topic: "machine"},
}
