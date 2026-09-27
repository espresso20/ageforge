package flavor

// HarbingerWarning — the warning itself, and what the settlement did with it.
// Request.Kind carries the risk tier (TierNone .. TierHigh); see harbinger.go.
//
// The tier is the whole point of this Moment. Before the Industrial Age the
// player has no number to read, so the prose is the only gauge they get, and a
// "none" warning must never sound like a "high" one. Every sentence here is
// tagged with the tiers it honestly fits: one tier, or an adjacent pair when a
// reaction reads the same at both (people going back to work fits "nothing to
// fear" and "a little to fear" alike). A small tier-neutral pool keeps a zero
// Request grammatical and is kept small on purpose, so that a caller who sends
// a tier gets that tier's voice almost every time.
//
// Precision follows config.HarbingerDef.ForecastPrecision. Ancient and feudal
// pools are vague: omens, cries, charts. From the Industrial Age the medium
// publishes the odds, so those pools talk about the figure, the odds, the
// estimate, and never print one. TestHarbingerPrecisionVocabulary holds the
// early pools to it.
//
// The quotable copy (the crier's cry, the headline, the leaked memo) lives in
// the per-age pools in catalog_harb_warning_ages.go, because it belongs to one
// speaker. The pools below are the settlement's side.

func harbWarningTemplates() []tmpl {
	out := pool("harb_warn_any", erasAny, harbWarnAny)
	out = append(out, pool("harb_warn_anytier", erasAny, harbWarnAnyTier)...)
	out = append(out, pool("harb_warn_ancient", erasAncient, harbWarnAncient)...)
	out = append(out, pool("harb_warn_feudal", erasFeudal, harbWarnFeudal)...)
	out = append(out, pool("harb_warn_industrial", erasIndustrial, harbWarnIndustrial)...)
	out = append(out, pool("harb_warn_digital", erasDigital, harbWarnDigital)...)
	out = append(out, pool("harb_warn_cosmic", erasCosmic, harbWarnCosmic)...)
	out = append(out, agePools("harb_warn_age", harbWarnAges)...)
	return out
}

// harbWarnAny is tier-neutral and era-neutral: a reaction that reads the same
// whether the warning was idle or dire, in a cave or on a ring station.
var harbWarnAny = []skel{
	{Text: "Everybody stopped to listen", Reg: rPlain, Topic: "noise"},
	{Text: "The children were sent indoors first, out of habit", Reg: rPlain, Topic: "family"},
	{Text: "The children asked for it to be said again", Kinds: tNoneLow, Reg: rPlain, Topic: "family"},
	{Text: "It was talked over at every meal that evening", Kinds: tLowMed, Reg: rPlain, Topic: "food"},
	{Text: "The listeners went home in twos and threes", Reg: rPlain, Topic: "town"},
	{Text: "By evening the words had passed from mouth to mouth so many times that the version the last family heard had a different ending, and that family slept better or worse than everybody else accordingly", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "A few people waited afterwards in case {subject} had anything to add", Needs: needSubject, Reg: rPlain, Topic: "stranger"},
}

// harbWarnAnyTier fits every era but only some tiers.
var harbWarnAnyTier = []skel{
	// none
	{Text: "Nothing to fear this time", Kinds: tNone, Reg: rPlain, Topic: "omen"},
	{Text: "A few people seemed disappointed", Kinds: tNone, Reg: rWry, Topic: "people"},
	{Text: "By the next day it had mostly been forgotten", Kinds: tNone, Reg: rPlain, Topic: "time"},
	{Text: "All the signs were good, or at least dull", Kinds: tNone, Reg: rWry, Topic: "omen"},
	{Text: "The warning, when it came, was that there was nothing to warn about", Kinds: tNone, Reg: rWry, Topic: "message"},
	{Text: "After looking the place over from end to end, {subject} could find nothing worth frightening anyone with", Kinds: tNone, Needs: needSubject, Reg: rWry, Topic: "stranger"},
	{Text: "Having come all this way to foretell a disaster, {subject} was left with the job of saying that none was due, and did it with the air of someone delivering bad news", Kinds: tNone, Needs: needSubject, Reg: rJoke, Topic: "stranger"},
	// none or low
	{Text: "One man laughed out loud and was looked at", Kinds: tNoneLow, Reg: rWry, Topic: "argument"},
	{Text: "Most people went back to work within the hour", Kinds: tNoneLow, Reg: rPlain, Topic: "work"},
	{Text: "Plans went ahead", Kinds: tNoneLow, Reg: rPlain, Topic: "time"},
	{Text: "The worry was gone by the next meal", Kinds: tNoneLow, Reg: rWry, Topic: "time"},
	// low
	{Text: "It was a warning of the mild kind, the kind people half hear", Kinds: tLow, Reg: rWry, Topic: "message"},
	{Text: "A few families moved their valuables somewhere safer", Kinds: tLow, Reg: rPlain, Topic: "family"},
	{Text: "People said it was probably nothing, and kept saying it all evening", Kinds: tLow, Reg: rWry, Form: fOverheard, Topic: "rumour"},
	// low or medium
	{Text: "The unease lingered into the next day", Kinds: tLowMed, Reg: rPlain, Topic: "time"},
	{Text: "Here and there, people began putting things aside", Kinds: tLowMed, Reg: rPlain, Topic: "haul"},
	{Text: "Nobody laughed at {subject}", Kinds: tLowMed, Needs: needSubject, Reg: rPlain, Topic: "people"},
	// medium
	{Text: "Voices were kept low for the rest of the day", Kinds: tMed, Reg: rPlain, Topic: "noise"},
	{Text: "People checked on their neighbours without being asked to", Kinds: tMed, Reg: rPlain, Topic: "message"},
	{Text: "Families sat up talking about where they would go if they had to", Kinds: tMed, Reg: rPlain, Topic: "family"},
	{Text: "Two families had started carrying their things to the strongest building before the warning was finished, and a third was waiting to see what the first two did", Kinds: tMed, Reg: rWry, Topic: "people"},
	// medium or high
	{Text: "Old people nodded", Kinds: tMedHigh, Reg: rPlain, Topic: "people"},
	{Text: "Sleep came hard", Kinds: tMedHigh, Reg: rPlain, Topic: "sleep"},
	{Text: "Doors were shut earlier than usual", Kinds: tMedHigh, Reg: rPlain, Topic: "building"},
	{Text: "Mothers kept their children within reach all evening", Kinds: tMedHigh, Reg: rPlain, Topic: "family"},
	// high
	{Text: "It was terror, plainly", Kinds: tHigh, Reg: rPlain, Topic: "omen"},
	{Text: "People held their neighbours in the street", Kinds: tHigh, Reg: rPlain, Topic: "town"},
	{Text: "The oldest people went round saying goodbye to the youngest", Kinds: tHigh, Reg: rPlain, Topic: "family"},
	{Text: "People began saying goodbye to each other at doorways, in the ordinary way, as if it were any evening", Kinds: tHigh, Reg: rPlain, Topic: "time"},
	{Text: "Old Maud, who has never believed a warning in her life, sat down and cried", Kinds: tHigh, Reg: rPlain, Topic: "name"},
	{Text: "It said the end was close, and it said so plainly", Kinds: tHigh, Reg: rPlain, Topic: "message"},
	{Text: "At no point did {subject} try to soften any of it", Kinds: tHigh, Needs: needSubject, Reg: rPlain, Topic: "stranger"},
	{Text: "Grown men sat down where they stood and put their hands over their faces, and their families stood round them not knowing what to do, because nobody had ever seen them like that", Kinds: tHigh, Reg: rPlain, Topic: "people"},
}

var harbWarnAncient = []skel{
	// none
	{Text: "No omens worth the name", Kinds: tNone, Reg: rPlain, Topic: "omen"},
	{Text: "The sky stayed where it was", Kinds: tNone, Reg: rPlain, Topic: "weather"},
	{Text: "Hunters went out as usual", Kinds: tNone, Reg: rPlain, Topic: "work"},
	{Text: "The elders listened, found nothing in it, and went back to the fire", Kinds: tNone, Reg: rPlain, Topic: "authority"},
	{Text: "It was a long way to come to say that the sky would stay up", Kinds: tNone, Reg: rWry, Topic: "stranger"},
	{Text: "The children had made a game of the prophecy by nightfall", Kinds: tNone, Reg: rWry, Topic: "family"},
	{Text: "Some of the young men said they had known all along, which they had not", Kinds: tNone, Reg: rJoke, Form: fOverheard, Topic: "people"},
	{Text: "A woman who had buried her good knife dug it up again", Kinds: tNone, Reg: rWry, Topic: "haul"},
	{Text: "The elders asked three different ways, the way you ask a child whether it has really washed, and got the one answer each time, that nothing was coming, and let it go and went to eat", Kinds: tNone, Reg: rWry, Topic: "authority"},
	// none or low
	{Text: "The drums started up again after dark", Kinds: tNoneLow, Reg: rPlain, Topic: "noise"},
	{Text: "The herd stayed near", Kinds: tNoneLow, Reg: rPlain, Topic: "animal"},
	{Text: "The warning was mostly about the weather, and the weather is always bad", Kinds: tNoneLow, Reg: rJoke, Topic: "weather"},
	{Text: "The elders said to keep the fires going and not to worry, in that order", Kinds: tNoneLow, Reg: rWry, Form: fNotice, Topic: "authority"},
	// low
	{Text: "Something about a dry season, perhaps", Kinds: tLow, Reg: rPlain, Form: fOverheard, Topic: "weather"},
	{Text: "A few families dug their grain pits out and moved them further up the slope, grumbling the whole time and saying it was only to be safe, and then stood around the new pits looking pleased with themselves", Kinds: tLow, Reg: rPlain, Topic: "food"},
	{Text: "Charms on the children's wrists", Kinds: tLow, Reg: rPlain, Topic: "religion"},
	{Text: "The warning was of a bad season, the kind the old people remember a few of", Kinds: tLow, Reg: rPlain, Topic: "time"},
	{Text: "Hunters were told to stay within sight of the camp for a moon", Kinds: tLow, Reg: rPlain, Form: fNotice, Topic: "authority"},
	// low or medium
	{Text: "The elders sent the fastest boy up to watch the pass, with a skin of water and orders to run back without stopping", Kinds: tLowMed, Reg: rPlain, Topic: "border"},
	{Text: "Offerings began to appear at the standing stones without anybody being told to leave them", Kinds: tLowMed, Reg: rWry, Topic: "religion"},
	{Text: "The old men watched the sky", Kinds: tLowMed, Reg: rPlain, Topic: "omen"},
	{Text: "Everything that could be dried was being dried", Kinds: tLowMed, Reg: rPlain, Topic: "food"},
	// medium
	{Text: "The prophecy spoke of the ground shaking and the herds running", Kinds: tMed, Reg: rPlain, Form: fOverheard, Topic: "omen"},
	{Text: "The women began to sort the stores into what could be carried and what would be left", Kinds: tMed, Reg: rPlain, Topic: "haul"},
	{Text: "On the elders' order the drums were silent that night", Kinds: tMed, Reg: rPlain, Form: fNotice, Topic: "noise"},
	{Text: "Families that had quarrelled for a generation over a stolen pig shared a fire that night, and passed each other food, and the pig was never mentioned, though everyone around that fire was thinking about it", Kinds: tMed, Reg: rWry, Topic: "family"},
	{Text: "It spoke of a season when the rivers would run the wrong colour and the animals would come down out of the hills to die among the huts, and the old women, who had heard such things before, did not argue with it", Kinds: tMed, Reg: rPlain, Topic: "omen"},
	// medium or high
	{Text: "The fires burned all night", Kinds: tMedHigh, Reg: rPlain, Topic: "sleep"},
	{Text: "No child was let out of sight", Kinds: tMedHigh, Reg: rPlain, Topic: "family"},
	{Text: "The hunters stayed in", Kinds: tMedHigh, Reg: rPlain, Topic: "work"},
	{Text: "The men sharpened their spears without knowing against what", Kinds: tMedHigh, Reg: rWry, Topic: "weapon"},
	// high
	{Text: "The elders cut their hair", Kinds: tHigh, Reg: rPlain, Topic: "religion"},
	{Text: "Nobody ate that night", Kinds: tHigh, Reg: rPlain, Topic: "food"},
	{Text: "The sky will burn, it said, and the ground will open", Kinds: tHigh, Reg: rPlain, Form: fOverheard, Topic: "omen"},
	{Text: "Women keened at the fire all night", Kinds: tHigh, Reg: rPlain, Topic: "noise"},
	{Text: "The herd was driven into the gully and the gully was walled with brush", Kinds: tHigh, Reg: rPlain, Topic: "animal"},
	{Text: "A man walked out of the camp that night with his family and did not come back", Kinds: tHigh, Reg: rPlain, Topic: "people"},
	{Text: "The drums beat until the drummers' hands bled", Kinds: tHigh, Reg: rPlain, Topic: "noise"},
	{Text: "Every child was painted with the protecting marks, even the babies", Kinds: tHigh, Reg: rPlain, Topic: "religion"},
	{Text: "It said the fire in the sky would come down to the ground, and that what it touched would never grow again, and afterwards the only sound in the camp was the fire", Kinds: tHigh, Reg: rPlain, Topic: "omen"},
}

var harbWarnFeudal = []skel{
	// none
	{Text: "Market day went ahead", Kinds: tNone, Reg: rPlain, Topic: "trade"},
	{Text: "The alewife reopened by noon", Kinds: tNone, Reg: rPlain, Topic: "trade"},
	{Text: "The reeve had the extra watch stood down", Kinds: tNone, Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "The priest announced from the pulpit that nothing was coming, and looked relieved", Kinds: tNone, Reg: rPlain, Topic: "religion"},
	{Text: "Children ran after the carts shouting that the world was not ending after all", Kinds: tNone, Reg: rWry, Topic: "family"},
	{Text: "Brother Anselm, who keeps the abbey's chronicle, gave it a single line and went back to the price of eels", Kinds: tNone, Reg: rJoke, Topic: "paper"},
	{Text: "The bailiff complains that he doubled the watch for nothing and will be asking the lord who pays for it", Kinds: tNone, Reg: rWry, Form: fComplaint, Topic: "money"},
	{Text: "There was a certain amount of grumbling in the tavern that evening from men who had sold their second cow on the strength of the first rumour, and nobody bought them a drink", Kinds: tNone, Reg: rWry, Form: fComplaint, Topic: "trade"},
	// none or low
	{Text: "The bells rang only for vespers", Kinds: tNoneLow, Reg: rPlain, Topic: "noise"},
	{Text: "The steward changed nothing", Kinds: tNoneLow, Reg: rPlain, Topic: "authority"},
	{Text: "The parish heard it out politely, the way it hears a long sermon", Kinds: tNoneLow, Reg: rWry, Topic: "religion"},
	{Text: "Carts left for the fair on the usual morning", Kinds: tNoneLow, Reg: rPlain, Topic: "trade"},
	// low
	{Text: "Something about a wet summer and a poor harvest, perhaps", Kinds: tLow, Reg: rPlain, Form: fOverheard, Topic: "weather"},
	{Text: "A few households laid in an extra sack of flour", Kinds: tLow, Reg: rPlain, Topic: "food"},
	{Text: "The priest added a line to the Sunday prayers, asking for protection from whatever it was, and the congregation said it louder than the rest", Kinds: tLow, Reg: rPlain, Topic: "religion"},
	{Text: "The miller raised his price a little", Kinds: tLow, Reg: rWry, Topic: "trade"},
	{Text: "Pilgrims were told to keep to the roads", Kinds: tLow, Reg: rPlain, Form: fNotice, Topic: "authority"},
	// low or medium
	{Text: "Candles sold well at the church door", Kinds: tLowMed, Reg: rWry, Topic: "trade"},
	{Text: "The mayor had the gates checked", Kinds: tLowMed, Reg: rPlain, Topic: "building"},
	{Text: "Wills were drawn up that had been put off for years", Kinds: tLowMed, Reg: rPlain, Topic: "paper"},
	{Text: "The miller's wife began putting flour by in the loft", Kinds: tLowMed, Reg: rPlain, Topic: "food"},
	// medium
	{Text: "The village sat up late", Kinds: tMed, Reg: rPlain, Topic: "sleep"},
	{Text: "Men went up on the walls to look at the road, and came down, and went up again", Kinds: tMed, Reg: rPlain, Topic: "border"},
	{Text: "The priest heard confessions until midnight and then sent for a second priest from the next parish, who arrived at two in the morning in his nightshirt and a borrowed cloak and heard confessions until it was light", Kinds: tMed, Reg: rPlain, Topic: "religion"},
	{Text: "Carts full of household goods were seen leaving by the north gate", Kinds: tMed, Reg: rPlain, Topic: "trade"},
	{Text: "The guild masters met and agreed to meet again", Kinds: tMed, Reg: rJoke, Topic: "authority"},
	// medium or high
	{Text: "The church stayed full", Kinds: tMedHigh, Reg: rPlain, Topic: "religion"},
	{Text: "Every door on the high street had a chalk cross on it by morning", Kinds: tMedHigh, Reg: rPlain, Topic: "building"},
	{Text: "No carts came in from the east", Kinds: tMedHigh, Reg: rPlain, Topic: "border"},
	{Text: "The watch was doubled, then doubled again", Kinds: tMedHigh, Reg: rPlain, Form: fNotice, Topic: "authority"},
	// high
	{Text: "The bells rang all night", Kinds: tHigh, Reg: rPlain, Topic: "noise"},
	{Text: "The fair was cancelled", Kinds: tHigh, Reg: rPlain, Form: fNotice, Topic: "trade"},
	{Text: "Wives kissed their husbands at the door, in the street, where anyone could see", Kinds: tHigh, Reg: rPlain, Topic: "family"},
	{Text: "The lord's household left before dawn with every cart they owned", Kinds: tHigh, Reg: rPlain, Topic: "authority"},
	{Text: "Monks walked the walls with the relics, chanting", Kinds: tHigh, Reg: rPlain, Topic: "religion"},
	{Text: "A woman in the market sold her whole stall for a horse", Kinds: tHigh, Reg: rPlain, Topic: "trade"},
	{Text: "The gates were shut at noon and the keys carried up to the castle", Kinds: tHigh, Reg: rPlain, Topic: "building"},
	{Text: "Old Wat the smith was at the altar before the priest", Kinds: tHigh, Reg: rWry, Topic: "people"},
	{Text: "By nightfall the road north was a single slow line of carts and handbarrows and people carrying their children, and the ones who stayed behind stood in their doorways and watched it go without saying anything to each other", Kinds: tHigh, Reg: rPlain, Topic: "people"},
}

var harbWarnIndustrial = []skel{
	// none
	{Text: "Business as usual", Kinds: tNone, Reg: rPlain, Topic: "work"},
	{Text: "The papers ran it on page nine, under the cattle prices", Kinds: tNone, Reg: rWry, Topic: "paper"},
	{Text: "The published odds were long enough that nobody read past the headline", Kinds: tNone, Reg: rWry, Topic: "paper"},
	{Text: "The foreman docked anyone who stopped to talk about it", Kinds: tNone, Reg: rWry, Topic: "authority"},
	{Text: "The bookmakers stopped taking bets on it by noon", Kinds: tNone, Reg: rPlain, Topic: "money"},
	{Text: "The figure was so small that the printers had to set it in the smallest type they had", Kinds: tNone, Reg: rJoke, Topic: "paper"},
	{Text: "The usual Wednesday turnout at chapel", Kinds: tNone, Reg: rPlain, Form: fLedger, Topic: "religion"},
	{Text: "The insurance office on the high street, which had put an extra man on the counter in anticipation, sent him home at lunchtime, and he stood outside for a while in his good collar not knowing what to do with the afternoon", Kinds: tNone, Reg: rWry, Topic: "money"},
	// none or low
	{Text: "The trams kept running", Kinds: tNoneLow, Reg: rPlain, Topic: "machine"},
	{Text: "The odds were printed, read, and forgotten by the second cup of tea", Kinds: tNoneLow, Reg: rWry, Topic: "time"},
	{Text: "Nobody at the works took the afternoon off", Kinds: tNoneLow, Reg: rPlain, Topic: "work"},
	{Text: "The shops kept their usual hours", Kinds: tNoneLow, Reg: rPlain, Topic: "trade"},
	// low
	{Text: "Insurance enquiries were up, a little", Kinds: tLow, Reg: rPlain, Form: fLedger, Topic: "money"},
	{Text: "The odds were small but they were printed, and printed odds have a way of sticking", Kinds: tLow, Reg: rWry, Topic: "paper"},
	{Text: "A few people bought an extra tin of something", Kinds: tLow, Reg: rPlain, Topic: "food"},
	{Text: "The figure was discussed in the pub in the tone usually kept for the weather", Kinds: tLow, Reg: rWry, Form: fOverheard, Topic: "weather"},
	{Text: "The council issued a statement saying it was monitoring the situation", Kinds: tLow, Reg: rJoke, Form: fNotice, Topic: "authority"},
	// low or medium
	{Text: "Queues formed outside the grocer's", Kinds: tLowMed, Reg: rPlain, Topic: "trade"},
	{Text: "People did sums on the backs of envelopes and did not like the answers", Kinds: tLowMed, Reg: rWry, Topic: "count"},
	{Text: "The churches reported larger congregations than at Easter", Kinds: tLowMed, Reg: rPlain, Form: fLedger, Topic: "religion"},
	{Text: "The price of tinned food crept up through the week", Kinds: tLowMed, Reg: rPlain, Topic: "trade"},
	// medium
	{Text: "The figure was the first thing anybody said to anybody that morning", Kinds: tMed, Reg: rPlain, Form: fOverheard, Topic: "count"},
	{Text: "Firms in the city began moving their books out to the country, and on the roads north there were vans piled high with ledgers and filing cabinets and junior staff sitting on top holding them steady", Kinds: tMed, Reg: rPlain, Topic: "work"},
	{Text: "The bank had a queue round the corner by ten, and the manager came out at half past to say there was money for everyone, which made it longer", Kinds: tMed, Reg: rPlain, Topic: "money"},
	{Text: "The odds were printed in a black box on the front page, which the papers had never done before", Kinds: tMed, Reg: rPlain, Topic: "paper"},
	{Text: "Families argued over dinner about whether the figure was high or merely unpleasant", Kinds: tMed, Reg: rWry, Topic: "family"},
	// medium or high
	{Text: "The number was worse than yesterday's", Kinds: tMedHigh, Reg: rPlain, Topic: "count"},
	{Text: "Men stood outside the newspaper offices waiting for the next edition", Kinds: tMedHigh, Reg: rPlain, Topic: "town"},
	{Text: "Schools closed early", Kinds: tMedHigh, Reg: rPlain, Form: fNotice, Topic: "family"},
	{Text: "Every hardware shop in town sold out of rope and candles", Kinds: tMedHigh, Reg: rPlain, Form: fLedger, Topic: "trade"},
	// high
	{Text: "The figure was printed in red", Kinds: tHigh, Reg: rPlain, Topic: "paper"},
	{Text: "Nobody went in to work", Kinds: tHigh, Reg: rPlain, Topic: "work"},
	{Text: "Crowds gathered at the stations and the stations could not take them", Kinds: tHigh, Reg: rPlain, Topic: "machine"},
	{Text: "The mayor spoke from the town hall steps and could not be heard over the crowd", Kinds: tHigh, Reg: rPlain, Topic: "authority"},
	{Text: "The factories closed at noon without saying when they would reopen", Kinds: tHigh, Reg: rPlain, Form: fNotice, Topic: "work"},
	{Text: "The odds were on every front page in the country and people read them standing in the street, one after another, and handed the paper on", Kinds: tHigh, Reg: rPlain, Topic: "paper"},
	{Text: "Men queued outside the enlistment office without anyone having called for them", Kinds: tHigh, Reg: rPlain, Topic: "people"},
	{Text: "The churches kept their doors open through the night and there were people in every pew", Kinds: tHigh, Reg: rPlain, Topic: "religion"},
	{Text: "In the terraces behind the works people took their good china down from the dressers and wrapped it in newspaper, piece by piece, and packed it into tea chests, as if the china were the thing that needed saving", Kinds: tHigh, Reg: rWry, Topic: "family"},
}

var harbWarnDigital = []skel{
	// none
	{Text: "The risk score came back green", Kinds: tNone, Reg: rPlain, Topic: "machine"},
	{Text: "The markets barely moved", Kinds: tNone, Reg: rPlain, Topic: "money"},
	{Text: "The number was so low it rounded to nothing", Kinds: tNone, Reg: rPlain, Topic: "count"},
	{Text: "The trending topic was gone by dinner", Kinds: tNone, Reg: rPlain, Topic: "rumour"},
	{Text: "It was a meme within the hour, and a stale one by evening", Kinds: tNone, Reg: rWry, Topic: "rumour"},
	{Text: "The official feed posted a reassurance and turned the comments off", Kinds: tNone, Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "People went back to arguing about other things", Kinds: tNone, Reg: rPlain, Topic: "argument"},
	{Text: "A man who had spent the previous night stockpiling bottled water and batteries posted a long thread explaining why he did not regret it, and it got more attention than the forecast itself had", Kinds: tNone, Reg: rJoke, Form: fComplaint, Topic: "people"},
	// none or low
	{Text: "Calendars stayed full", Kinds: tNoneLow, Reg: rPlain, Topic: "time"},
	{Text: "Office chatter moved on by lunch", Kinds: tNoneLow, Reg: rPlain, Topic: "work"},
	{Text: "The probability was published, noted and scrolled past", Kinds: tNoneLow, Reg: rWry, Topic: "message"},
	{Text: "The drone deliveries kept coming", Kinds: tNoneLow, Reg: rPlain, Topic: "machine"},
	// low
	{Text: "Sales of power banks ticked up", Kinds: tLow, Reg: rPlain, Form: fLedger, Topic: "trade"},
	{Text: "The probability was low enough to joke about and high enough to check again", Kinds: tLow, Reg: rWry, Topic: "count"},
	{Text: "A few people downloaded offline maps, just in case", Kinds: tLow, Reg: rPlain, Topic: "map"},
	{Text: "The risk score went amber on a handful of dashboards", Kinds: tLow, Reg: rPlain, Topic: "machine"},
	{Text: "An advisory went out asking people to keep their devices charged", Kinds: tLow, Reg: rPlain, Form: fNotice, Topic: "authority"},
	// low or medium
	{Text: "The number kept getting refreshed", Kinds: tLowMed, Reg: rPlain, Topic: "machine"},
	{Text: "People set alerts for any change in the forecast and then lay awake waiting for their phones to buzz, and when they did, it was usually friends asking whether they had seen it", Kinds: tLowMed, Reg: rPlain, Topic: "time"},
	{Text: "The group chats filled up with links to preparedness guides, most of them written by people selling something, and people read them anyway, in bed, with the brightness turned all the way down so as not to wake anyone", Kinds: tLowMed, Reg: rPlain, Topic: "message"},
	{Text: "Delivery slots sold out", Kinds: tLowMed, Reg: rPlain, Form: fLedger, Topic: "trade"},
	// medium
	{Text: "The probability was the first thing on every screen that morning", Kinds: tMed, Reg: rPlain, Topic: "machine"},
	{Text: "Companies told staff to work from home and to back up everything", Kinds: tMed, Reg: rPlain, Form: fNotice, Topic: "work"},
	{Text: "The supermarkets rationed water to two packs a customer", Kinds: tMed, Reg: rPlain, Form: fNotice, Topic: "food"},
	{Text: "Offline copies of everything were being made, just in case", Kinds: tMed, Reg: rPlain, Topic: "paper"},
	{Text: "The analysts argued about the model on every channel, and the model did not care", Kinds: tMed, Reg: rWry, Topic: "argument"},
	// medium or high
	{Text: "Nobody could get through to anybody", Kinds: tMedHigh, Reg: rPlain, Form: fComplaint, Topic: "message"},
	{Text: "Emergency apps topped the download charts", Kinds: tMedHigh, Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The official feeds stopped posting jokes", Kinds: tMedHigh, Reg: rWry, Topic: "authority"},
	{Text: "The networks slowed under the weight of everyone checking", Kinds: tMedHigh, Reg: rPlain, Topic: "machine"},
	// high
	{Text: "The markets halted", Kinds: tHigh, Reg: rPlain, Form: fNotice, Topic: "money"},
	{Text: "The number was red and it was large", Kinds: tHigh, Reg: rPlain, Topic: "count"},
	{Text: "The highways out of the city locked solid inside an hour", Kinds: tHigh, Reg: rPlain, Topic: "ground"},
	{Text: "Every phone in the city sounded the emergency tone at once", Kinds: tHigh, Reg: rPlain, Topic: "noise"},
	{Text: "Server farms began shutting down in rolling sequence to protect the hardware", Kinds: tHigh, Reg: rPlain, Topic: "machine"},
	{Text: "Families rang relatives they had not spoken to in years", Kinds: tHigh, Reg: rPlain, Topic: "family"},
	{Text: "The probability was read out on every screen in the city by a voice that had clearly been told to sound calm and could not manage it", Kinds: tHigh, Reg: rPlain, Topic: "noise"},
	{Text: "Nobody believed it and everybody acted on it", Kinds: tHigh, Reg: rWry, Topic: "people"},
	{Text: "In the office towers people left their desks without logging off, and the screens went on glowing on the empty floors all night, showing that one number to nobody", Kinds: tHigh, Reg: rPlain, Topic: "machine"},
}

var harbWarnCosmic = []skel{
	// none
	{Text: "Cargo schedules resumed", Kinds: tNone, Reg: rPlain, Topic: "trade"},
	{Text: "The crews stood down", Kinds: tNone, Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "No anomaly on any band", Kinds: tNone, Reg: rPlain, Form: fLedger, Topic: "machine"},
	{Text: "The long-range estimate came back clean", Kinds: tNone, Reg: rPlain, Topic: "machine"},
	{Text: "The number was small enough that the navigation officer called it noise", Kinds: tNone, Reg: rPlain, Topic: "count"},
	{Text: "The deep watch went back to its usual silence", Kinds: tNone, Reg: rPlain, Topic: "noise"},
	{Text: "The ring's chaplain gave a short service of thanks and then a longer one about vigilance, because he had prepared it", Kinds: tNone, Reg: rJoke, Topic: "religion"},
	{Text: "For a whole shift the observation decks were crowded with people staring out at the dark for whatever had been expected, and the dark stayed exactly as dark as it had always been, and by the next shift they had gone", Kinds: tNone, Reg: rPlain, Topic: "omen"},
	// none or low
	{Text: "Shipping kept to its routes", Kinds: tNoneLow, Reg: rPlain, Topic: "trade"},
	{Text: "The probability sat just above zero, where it usually sits", Kinds: tNoneLow, Reg: rPlain, Topic: "count"},
	{Text: "The shield crews finished their shift and went to eat", Kinds: tNoneLow, Reg: rPlain, Topic: "work"},
	{Text: "Children on the ring were told it was nothing and believed it", Kinds: tNoneLow, Reg: rPlain, Topic: "family"},
	// low
	{Text: "The estimate showed a small risk, the kind the station lives with", Kinds: tLow, Reg: rPlain, Topic: "time"},
	{Text: "Maintenance moved the hull inspection forward a week", Kinds: tLow, Reg: rPlain, Form: fNotice, Topic: "machine"},
	{Text: "A few people checked where their assigned evacuation pods were, and one or two walked down to look at them, and came back feeling slightly foolish", Kinds: tLow, Reg: rPlain, Topic: "people"},
	{Text: "The probability was low, and the dark is large", Kinds: tLow, Reg: rWry, Topic: "omen"},
	{Text: "The spare air scrubbers were brought up from storage", Kinds: tLow, Reg: rPlain, Topic: "kit"},
	// low or medium
	{Text: "All sensor time went outward", Kinds: tLowMed, Reg: rPlain, Topic: "machine"},
	{Text: "The station ran an evacuation drill that people took seriously", Kinds: tLowMed, Reg: rPlain, Topic: "building"},
	{Text: "The long-range estimate was updated every hour and read every minute", Kinds: tLowMed, Reg: rWry, Topic: "time"},
	{Text: "Crew began carrying their pressure suits folded over one arm between shifts, into the mess and the showers and the chapel, and nobody commented on it", Kinds: tLowMed, Reg: rPlain, Topic: "kit"},
	// medium
	{Text: "The estimate was high enough to be read aloud at every shift change", Kinds: tMed, Reg: rPlain, Topic: "count"},
	{Text: "The outer habitats were ordered to seal their bulkheads at night", Kinds: tMed, Reg: rPlain, Form: fNotice, Topic: "building"},
	{Text: "Families moved their bedding into the inner ring, where the hull is thickest, and slept in the corridors outside the hydroponics bays among the smell of wet soil, and the children thought it was a holiday", Kinds: tMed, Reg: rPlain, Topic: "family"},
	{Text: "The probability was a real number now, and people treated it as one", Kinds: tMed, Reg: rPlain, Topic: "people"},
	{Text: "Every spare berth on the evacuation shuttles was assigned by lottery", Kinds: tMed, Reg: rPlain, Form: fNotice, Topic: "authority"},
	// medium or high
	{Text: "The chapel was never empty", Kinds: tMedHigh, Reg: rPlain, Topic: "religion"},
	{Text: "The night watch stopped talking", Kinds: tMedHigh, Reg: rPlain, Topic: "noise"},
	{Text: "The shields were run up to full and kept there", Kinds: tMedHigh, Reg: rPlain, Topic: "machine"},
	{Text: "People stopped looking out of the viewports", Kinds: tMedHigh, Reg: rPlain, Topic: "omen"},
	// high
	{Text: "The outer ring was abandoned", Kinds: tHigh, Reg: rPlain, Topic: "building"},
	{Text: "Evacuation shuttles began loading children first", Kinds: tHigh, Reg: rPlain, Form: fNotice, Topic: "family"},
	{Text: "The estimate was read out and then the comm line went dead for a full minute", Kinds: tHigh, Reg: rPlain, Topic: "message"},
	{Text: "The probability was nearly certain and still rising", Kinds: tHigh, Reg: rPlain, Topic: "count"},
	{Text: "People held hands in the corridors with strangers", Kinds: tHigh, Reg: rPlain, Topic: "stranger"},
	{Text: "The observation decks were sealed, on the grounds that nobody should have to watch it come", Kinds: tHigh, Reg: rPlain, Topic: "building"},
	{Text: "Every clock on the station had become a countdown", Kinds: tHigh, Reg: rPlain, Topic: "time"},
	{Text: "The number kept climbing", Kinds: tHigh, Reg: rPlain, Topic: "count"},
	{Text: "In the great dark past the last charted star, something that had been patient for longer than there have been stars had finally turned its attention toward the colony, and the estimate on every screen was simply the colony's way of saying so", Kinds: tHigh, Reg: rPlain, Topic: "omen"},
}
