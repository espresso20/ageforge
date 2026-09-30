package flavor

// The verdicts, logged when the doom resolves (it strikes or passes by, or a
// false prophet's window runs out):
//
//   - HarbingerVindicated: the harbinger warned and the catastrophe came. The
//     caller sends Kind: KindFalseProphet when the warning had been invented,
//     which adds a few lines (ancient and feudal pools) about a fraud proved
//     right by chance.
//   - HarbingerSpared: a real harbinger warned and the roll went the player's
//     way. The danger was real; this is relief, and a little embarrassment.
//   - HarbingerDiscredited: a false prophet, found out. A false warning is
//     rolled once per epoch, at the epoch's first age, so it can reach the last
//     figure of the Stone, Iron and Steel Eras (the Steel Era's is the
//     Industrial Age's Newsboy, who falls back on the era-neutral pool).
//
// The catastrophe has its own text. These sentences are about the WARNING:
// who remembered it, who apologised for laughing, who wants their candles
// back. None of them describes the damage.

func harbVindicatedTemplates() []tmpl {
	out := pool("harb_vindicated_any", erasAny, harbVindicatedAny)
	out = append(out, pool("harb_vindicated_grounded", erasGrounded, harbVindicatedGrounded)...)
	out = append(out, pool("harb_vindicated_late", erasLate, harbVindicatedLate)...)
	out = append(out, pool("harb_vindicated_ancient", erasAncient, harbVindicatedAncient)...)
	out = append(out, pool("harb_vindicated_feudal", erasFeudal, harbVindicatedFeudal)...)
	out = append(out, pool("harb_vindicated_industrial", erasIndustrial, harbVindicatedIndustrial)...)
	out = append(out, pool("harb_vindicated_digital", erasDigital, harbVindicatedDigital)...)
	out = append(out, pool("harb_vindicated_cosmic", erasCosmic, harbVindicatedCosmic)...)
	return out
}

func harbSparedTemplates() []tmpl {
	out := pool("harb_spared_any", erasAny, harbSparedAny)
	out = append(out, pool("harb_spared_grounded", erasGrounded, harbSparedGrounded)...)
	out = append(out, pool("harb_spared_late", erasLate, harbSparedLate)...)
	out = append(out, pool("harb_spared_ancient", erasAncient, harbSparedAncient)...)
	out = append(out, pool("harb_spared_feudal", erasFeudal, harbSparedFeudal)...)
	out = append(out, pool("harb_spared_industrial", erasIndustrial, harbSparedIndustrial)...)
	out = append(out, pool("harb_spared_digital", erasDigital, harbSparedDigital)...)
	out = append(out, pool("harb_spared_cosmic", erasCosmic, harbSparedCosmic)...)
	return out
}

func harbDiscreditedTemplates() []tmpl {
	out := pool("harb_discredited_any", erasAny, harbDiscreditedAny)
	out = append(out, pool("harb_discredited_early", erasEarly, harbDiscreditedEarly)...)
	out = append(out, pool("harb_discredited_ancient", erasAncient, harbDiscreditedAncient)...)
	out = append(out, pool("harb_discredited_feudal", erasFeudal, harbDiscreditedFeudal)...)
	out = append(out, agePools("harb_discredited_age", harbDiscreditedAges)...)
	return out
}

// --- Vindicated ---------------------------------------------------------------

var harbVindicatedAny = []skel{
	{Text: "The laughing stopped", Reg: rPlain, Topic: "people"},
	{Text: "Everyone remembered every word", Reg: rPlain, Topic: "message"},
	{Text: "People apologised, awkwardly", Reg: rWry, Topic: "people"},
	{Text: "By noon nobody in the settlement could remember having mocked {subject}", Needs: needSubject, Reg: rWry, Topic: "argument"},
	{Text: "Every half-remembered word of the warning was suddenly remembered in full", Reg: rWry, Topic: "message"},
	{Text: "Several people claimed to have believed it from the start", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The warning is being taken apart for everything else it might have meant", Reg: rPlain, Topic: "omen"},
	{Text: "The warning was repeated now in a low voice, word for word, and the jokes that had been made about it were not repeated at all", Reg: rWry, Topic: "rumour"},
	{Text: "Nobody thanked {subject} in words, though people went and stood nearby, one or two at a time, all afternoon", Needs: needSubject, Reg: rPlain, Topic: "stranger"},
	{Text: "In the days afterwards the warning was retold until it had grown far more precise than it ever was, and the version the children tell now names the day and the hour, which the original never did", Reg: rWry, Topic: "time"},
	{Text: "Everyone is complaining that nobody listened, mostly to people who also did not listen", Reg: rWry, Form: fComplaint, Topic: "argument"},
	{Text: "Orders went out that the warning be kept and repeated as it was given", Reg: rPlain, Form: fNotice, Topic: "authority"},
}

var harbVindicatedGrounded = []skel{
	{Text: "The dogs had known", Reg: rPlain, Topic: "animal"},
	{Text: "Children who had played at the end of the world in the lanes were kept indoors", Reg: rPlain, Topic: "family"},
	{Text: "A woman who had called {subject} a madman went round with a basket of bread", Needs: needSubject, Reg: rPlain, Topic: "food"},
	{Text: "The omens were gone over again by the fire, one by one", Reg: rPlain, Topic: "omen"},
	{Text: "Down every lane people stopped each other to say that they had said so, and nobody was listening to anybody, and it went on until dark", Reg: rWry, Form: fOverheard, Topic: "argument"},
	{Text: "The old men who had sat in the road watching the sky were brought indoors at last, and given the best seat, and asked what else they had noticed", Reg: rPlain, Topic: "people"},
	{Text: "The chalk marks that had been put on the doors as a joke before the warning came true were left there, and nobody would scrub them off, and some of them are still there on the doors of houses whose owners have since moved away", Reg: rPlain, Topic: "building"},
	{Text: "Candles are to be lit for the warning every year on the day it came", Reg: rPlain, Form: fNotice, Topic: "religion"},
}

var harbVindicatedLate = []skel{
	{Text: "Search traffic spiked again", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "Every screen brought the warning back up", Reg: rPlain, Topic: "machine"},
	{Text: "The original warning was pinned to the top of every public board", Reg: rPlain, Topic: "message"},
	{Text: "Experts went back over the warning line by line", Reg: rPlain, Topic: "paper"},
	{Text: "The logs from before it happened were pulled and studied, and it turned out a surprising number of systems had flagged the danger and been overridden", Reg: rWry, Topic: "machine"},
	{Text: "Muted copies of the warning were dug back out of settings menus, and there it was, unchanged, where people had left it", Reg: rPlain, Topic: "message"},
	{Text: "For weeks afterwards the settlement's systems kept surfacing the warning unprompted, in reminders and summaries and at the top of every morning briefing, as though the machines, too, needed everyone to know that they had been told", Reg: rWry, Topic: "machine"},
	{Text: "People complain that the warning should have been louder", Reg: rWry, Form: fComplaint, Topic: "noise"},
}

var harbVindicatedAncient = []skel{
	{Text: "The songs began", Reg: rPlain, Topic: "noise"},
	{Text: "The elders were silent", Reg: rPlain, Topic: "authority"},
	{Text: "Offerings were made to {subject}", Needs: needSubject, Reg: rPlain, Topic: "religion"},
	{Text: "The young men carried the heaviest loads afterwards, and did not have to be told", Reg: rWry, Topic: "people"},
	{Text: "The warning was painted on the rock face so it would never be forgotten", Reg: rPlain, Topic: "message"},
	{Text: "By order of the elders {subject} is to be fed for as long as the camp has food", Needs: needSubject, Reg: rPlain, Form: fNotice, Topic: "food"},
	{Text: "The drums told it to the next camp over", Reg: rPlain, Topic: "noise"},
	{Text: "A child who had repeated the prophecy in a silly voice was made to repeat it properly", Reg: rWry, Topic: "family"},
	{Text: "Everybody now remembers some sign they saw beforehand, a bird or a cloud or a strange stone, and the signs get bigger each time", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The hunters went out to the place where the herd had grazed and found the ground as the prophecy had said it would be, and came back without speaking", Reg: rPlain, Topic: "ground"},
	{Text: "The old women went over every word of the warning at the fire that night, and at each word one of them nodded, and at the end the oldest said it had been a true one and there would be songs", Reg: rPlain, Topic: "people"},
	{Text: "The hunters complain that nobody told them properly", Reg: rWry, Form: fComplaint, Topic: "argument"},
	{Text: "Three huts, one pit and most of the herd, as the warning said", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The herd has been counted, and the count matches the prophecy", Reg: rPlain, Form: fLedger, Topic: "animal"},
	// a false prophet, proved right by chance
	{Text: "It came out later that {subject} had made the whole warning up, and the elders could not decide whether that made it better or worse", Kinds: tFalse, Needs: needSubject, Reg: rPlain, Topic: "argument"},
	{Text: "The fraud had guessed right, and nobody in the camp knew what to do about it", Kinds: tFalse, Reg: rWry, Topic: "stranger"},
	{Text: "The hunters hold that a liar who is right is still a liar, and the old women answer that a liar who is right is still right", Kinds: tFalse, Reg: rWry, Form: fOverheard, Topic: "rumour"},
}

var harbVindicatedFeudal = []skel{
	{Text: "The bells tolled", Reg: rPlain, Topic: "noise"},
	{Text: "The priest preached humility", Reg: rWry, Topic: "religion"},
	{Text: "Pilgrims came to see the place", Reg: rPlain, Topic: "stranger"},
	{Text: "The parish had masses said for {subject}", Needs: needSubject, Reg: rPlain, Topic: "religion"},
	{Text: "The words of the warning were painted over the church door", Reg: rPlain, Topic: "message"},
	{Text: "The alewife who had called it nonsense stood the whole tavern a round", Reg: rWry, Topic: "trade"},
	{Text: "The abbey's chronicle gave it a whole page, with drawings", Reg: rPlain, Topic: "paper"},
	{Text: "The tavern was empty on Sunday and the church was full", Reg: rPlain, Topic: "people"},
	{Text: "The bishop wrote to the king that the warning had been true in every particular, and asked for money to rebuild, and got some of it", Reg: rWry, Topic: "authority"},
	{Text: "Ballads about it were being sung in the market within the week, and the ballad-singers got the details wrong in ways that improved them", Reg: rWry, Topic: "rumour"},
	{Text: "The guilds argued for a month over whose members had heeded the warning soonest, and it was eventually settled by the mayor, who awarded the honour to the chandlers because they had sold the most candles in the days before it came", Reg: rJoke, Topic: "argument"},
	{Text: "The steward complains that everyone is blaming him for not doubling the watch", Reg: rWry, Form: fComplaint, Topic: "authority"},
	{Text: "It is decreed that the day shall be kept as a day of fasting from now on", Reg: rPlain, Form: fNotice, Topic: "food"},
	{Text: "In the tavern they say {subject} knew the hour and would not tell it", Needs: needSubject, Reg: rPlain, Form: fOverheard, Topic: "time"},
	// a false prophet, proved right by chance
	{Text: "The friars could not settle whether a false prophecy that comes true is still false, and wrote to the bishop", Kinds: tFalse, Reg: rPlain, Topic: "religion"},
	{Text: "The alehouse verdict is that {subject} invented every word of it and was then proved right by bad luck", Kinds: tFalse, Needs: needSubject, Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The chronicle records the warning as a fraud and the disaster as foretold, on facing pages", Kinds: tFalse, Reg: rJoke, Topic: "paper"},
}

var harbVindicatedIndustrial = []skel{
	{Text: "The papers ran a correction", Reg: rWry, Topic: "paper"},
	{Text: "The odds had been right", Reg: rPlain, Topic: "count"},
	{Text: "Statues were proposed", Reg: rWry, Topic: "town"},
	{Text: "Circulation doubled the next morning", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The editorial that had called the forecast alarmist was taken down from the office wall where it had been framed", Reg: rWry, Topic: "paper"},
	{Text: "The mayor gave a speech about heeding warnings, which was long", Reg: rJoke, Topic: "authority"},
	{Text: "Men stood outside the newspaper offices reading the old editions pinned in the window", Reg: rPlain, Topic: "people"},
	{Text: "The bookmakers who had given long odds against it closed for a week", Reg: rPlain, Topic: "money"},
	{Text: "People carried the clipping of the forecast in their wallets", Reg: rPlain, Topic: "paper"},
	{Text: "An inquiry was announced into why the published odds had been ignored, and the inquiry was expected to take two years and publish odds of its own", Reg: rWry, Topic: "authority"},
	{Text: "The insurance companies honoured the policies they had sold in the week before, and their chairmen gave interviews about the importance of listening to experts", Reg: rWry, Topic: "money"},
	{Text: "In the pubs that evening men told each other that they had read the figure in the paper and had believed it, and each of them had in fact turned to the racing page, and each of them knew the others had too", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The council complains that the forecast was not specific enough", Reg: rWry, Form: fComplaint, Topic: "authority"},
	{Text: "The forecast will now be printed on the front page by order of the council", Reg: rPlain, Form: fNotice, Topic: "paper"},
}

var harbVindicatedDigital = []skel{
	{Text: "The forecast went viral again", Reg: rPlain, Topic: "rumour"},
	{Text: "Old screenshots circulated everywhere", Reg: rPlain, Topic: "message"},
	{Text: "The algorithms promoted it again", Reg: rWry, Topic: "machine"},
	{Text: "The fact-check that had called it exaggerated was updated", Reg: rWry, Topic: "paper"},
	{Text: "Every comment that had mocked the warning was being quoted back at its author", Reg: rWry, Topic: "argument"},
	{Text: "The model's authors were invited onto every programme", Reg: rPlain, Topic: "people"},
	{Text: "People searched for the warning and found it had been right to the hour", Reg: rPlain, Topic: "time"},
	{Text: "The companies that had ignored their own risk scores held press conferences about lessons learned, and their share prices went up", Reg: rWry, Topic: "money"},
	{Text: "A petition to make the forecast's authors national heroes had a million signatures by morning", Reg: rPlain, Topic: "name"},
	{Text: "Somewhere in the servers the original forecast still sits with its timestamp from before, unaltered, and people keep going back to look at the timestamp as if it might have changed, and it never has", Reg: rPlain, Topic: "time"},
	{Text: "The comments say the warning should have been clearer", Reg: rWry, Form: fComplaint, Topic: "message"},
	{Text: "An inquiry into why the warnings were ignored has been announced and livestreamed", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "Downloads of the preparedness app, up nine hundredfold", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "Everyone at work is saying they took it seriously", Reg: rWry, Form: fOverheard, Topic: "work"},
}

var harbVindicatedCosmic = []skel{
	{Text: "The logs were sealed", Reg: rPlain, Topic: "paper"},
	{Text: "The chaplain said nothing", Reg: rPlain, Topic: "religion"},
	{Text: "The colony remembered", Reg: rPlain, Topic: "time"},
	{Text: "The original estimate was read aloud at the memorial", Reg: rPlain, Topic: "count"},
	{Text: "The navigation officer who had called it noise resigned", Reg: rPlain, Topic: "people"},
	{Text: "Every deck recorded its own account of the warning", Reg: rPlain, Topic: "message"},
	{Text: "The station's archive tagged the warning as verified", Reg: rPlain, Topic: "machine"},
	{Text: "People kept returning to the observation deck where they had watched the dark for signs, and stood there, and did not say what they were looking for now", Reg: rPlain, Topic: "omen"},
	{Text: "The estimate had been right to within its stated margin, and the engineers had the margin engraved on a plate by the airlock", Reg: rWry, Topic: "count"},
	{Text: "Out in the dark the thing that had been foretold has moved on or gone back to sleep, and the only evidence that the warning was ever given is the colony itself, smaller now, and the list of names read out at every shift change", Reg: rPlain, Topic: "name"},
	{Text: "Crew complain the warning was buried under routine traffic", Reg: rWry, Form: fComplaint, Topic: "message"},
	{Text: "All future warnings are to be relayed to every deck immediately", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Warning received, warning logged, warning overruled, warning correct", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "People on the ring say {subject} should have been listened to", Needs: needSubject, Reg: rPlain, Form: fOverheard, Topic: "rumour"},
}

// --- Spared -------------------------------------------------------------------

var harbSparedAny = []skel{
	{Text: "Nothing came in the end", Reg: rPlain, Topic: "time"},
	{Text: "People slept late", Reg: rPlain, Topic: "sleep"},
	{Text: "The relief was loud", Reg: rPlain, Topic: "noise"},
	{Text: "Nobody blamed {subject}, since the danger had been real", Needs: needSubject, Reg: rPlain, Topic: "stranger"},
	{Text: "The things that had been hidden were brought back out, slowly", Reg: rPlain, Topic: "haul"},
	{Text: "Goodbyes that had been said at doorways were not mentioned again, though everybody remembered them", Reg: rWry, Topic: "family"},
	{Text: "The children were let out, and ran, and kept running until they reached the far edge of the settlement and could run no further", Reg: rPlain, Topic: "family"},
	{Text: "The warning had been true and the danger had been real, and it passed by anyway, the way a storm sometimes passes to one side", Reg: rPlain, Topic: "weather"},
	{Text: "Hale, who had prepared harder than anyone, was a little disappointed, and would not have admitted it for anything", Reg: rWry, Topic: "people"},
	{Text: "For a long time afterwards people talked about the close call in the voice they use for things that nearly happened to someone else, and it was months before anyone was willing to say out loud how close it had been", Reg: rPlain, Topic: "rumour"},
	{Text: "Now people grumble about all the work that turned out to be unneeded", Reg: rWry, Form: fComplaint, Topic: "work"},
	{Text: "Word was given that everyone could stand down", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Nobody can agree on who to thank", Reg: rWry, Form: fOverheard, Topic: "religion"},
}

var harbSparedGrounded = []skel{
	{Text: "The dogs were let out", Reg: rPlain, Topic: "animal"},
	{Text: "The old men stopped watching the sky", Reg: rPlain, Topic: "people"},
	{Text: "The fires were let burn down to their ordinary size", Reg: rPlain, Topic: "sleep"},
	{Text: "Doors that had been barred were unbarred, and left open all day", Reg: rPlain, Topic: "building"},
	{Text: "The candles that had been lit in the windows were left to burn out on their own, and nobody wanted to be the first to blow one out", Reg: rPlain, Topic: "town"},
	{Text: "In the lane the neighbours who had not spoken in years found they had nothing to say to each other again, and were relieved about that too", Reg: rWry, Topic: "people"},
	{Text: "When the danger had gone past, the oldest woman in the place went down to the water and washed her face and hands, as her mother had done the last time, and a line of people followed her down and did likewise", Reg: rPlain, Topic: "religion"},
	{Text: "Word came that the lookouts could come down", Reg: rPlain, Form: fNotice, Topic: "authority"},
}

var harbSparedLate = []skel{
	{Text: "Alerts cleared, one by one", Reg: rPlain, Topic: "machine"},
	{Text: "The screens went back to normal", Reg: rPlain, Topic: "machine"},
	{Text: "The safe rooms emptied out over an afternoon", Reg: rPlain, Topic: "building"},
	{Text: "The risk estimate dropped back to zero in real time while people watched", Reg: rPlain, Topic: "count"},
	{Text: "People took their earpieces out for the first time in days", Reg: rPlain, Topic: "sleep"},
	{Text: "The systems that had been shut down came back online in order, and each one announced itself with a small cheerful chime that people found unbearable", Reg: rWry, Topic: "machine"},
	{Text: "For a day nobody could settle to anything, and output everywhere dropped to nothing while people sat around telling each other they were fine", Reg: rWry, Topic: "work"},
	{Text: "The settlement's logs recorded the moment the danger passed down to the millisecond, and people keep going back to read that one entry among millions", Reg: rPlain, Topic: "time"},
}

var harbSparedAncient = []skel{
	{Text: "The drums played fast", Reg: rPlain, Topic: "noise"},
	{Text: "Four pits opened and all four still good", Reg: rPlain, Form: fLedger, Topic: "food"},
	{Text: "The elders laughed", Reg: rPlain, Topic: "authority"},
	{Text: "The herd was let out onto the high pasture", Reg: rPlain, Topic: "animal"},
	{Text: "Families that had shared a fire went back to their own, a little sadly", Reg: rWry, Topic: "family"},
	{Text: "The young men went hunting to prove they had never been afraid", Reg: rWry, Topic: "people"},
	{Text: "The offerings were left where they were, in case", Reg: rWry, Topic: "religion"},
	{Text: "The children were washed of their protecting marks", Reg: rPlain, Topic: "family"},
	{Text: "The elders said the warning had been true and the sky had chosen to spare them, and everyone agreed not to ask the sky why", Reg: rWry, Topic: "omen"},
	{Text: "The women dug up the food they had buried and found most of it still good, and the feast that night used all of it", Reg: rPlain, Topic: "food"},
	{Text: "A boy climbed the tallest rock to watch the horizon where the danger was meant to come from, and sat there until dark, and came down to say the horizon was only the horizon, and was given the first piece of meat for saying so", Reg: rPlain, Topic: "border"},
	{Text: "The hunters complain that the herd got fat and lazy in the pen", Reg: rWry, Form: fComplaint, Topic: "animal"},
	{Text: "The elders say the fires can go down now", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "At the fires they say {subject} will be welcome back any time", Needs: needSubject, Reg: rPlain, Form: fOverheard, Topic: "stranger"},
}

var harbSparedFeudal = []skel{
	{Text: "The bells rang out", Reg: rPlain, Topic: "noise"},
	{Text: "The fair went ahead", Reg: rPlain, Topic: "trade"},
	{Text: "The gates were opened", Reg: rPlain, Topic: "building"},
	{Text: "The priest said a mass of thanksgiving and kept it short, because half the congregation had already left for the tavern", Reg: rWry, Topic: "religion"},
	{Text: "Carts came back through the north gate loaded with the household goods they had left with", Reg: rPlain, Topic: "trade"},
	{Text: "The lord's household came back from the country and pretended they had been hunting", Reg: rJoke, Topic: "authority"},
	{Text: "The watch was stood down, grumbling about lost sleep", Reg: rWry, Topic: "sleep"},
	{Text: "The guildhall voted to declare the danger over, and then argued for an hour about the wording", Reg: rJoke, Topic: "authority"},
	{Text: "Candles that had been bought for the end of the world were used for ordinary things all winter, and the chandlers had a lean spring", Reg: rWry, Topic: "trade"},
	{Text: "In the tavern that night the men who had sold their cows cheap before the danger passed tried to buy them back, and the men who had bought them would not sell, and there was nearly a fight, and then there was a fight", Reg: rWry, Topic: "argument"},
	{Text: "The miller complains that he is left with a loft full of flour nobody wants", Reg: rWry, Form: fComplaint, Topic: "food"},
	{Text: "Dancing on the green tonight, by order of the reeve", Reg: rPlain, Form: fNotice, Topic: "town"},
	{Text: "In the market they say the saints turned it aside", Reg: rPlain, Form: fOverheard, Topic: "religion"},
}

var harbSparedIndustrial = []skel{
	{Text: "Sandbags returned to the depot, eight thousand", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The odds fell away", Reg: rPlain, Topic: "count"},
	{Text: "Trains ran again", Reg: rPlain, Topic: "machine"},
	{Text: "Schools reopen on Monday", Reg: rPlain, Form: fNotice, Topic: "family"},
	{Text: "The papers printed the odds one last time, much smaller", Reg: rWry, Topic: "paper"},
	{Text: "The queues at the bank vanished overnight, and the manager, who had been sleeping in his office, went home for the first time in a week", Reg: rPlain, Topic: "money"},
	{Text: "The shelters were locked up again, with the tins still inside", Reg: rPlain, Topic: "building"},
	{Text: "The enlistment office sent its volunteers home with thanks", Reg: rPlain, Topic: "people"},
	{Text: "The bookmakers who had given short odds on disaster had nobody to pay, and threw a party", Reg: rWry, Topic: "money"},
	{Text: "A great deal of tinned food was eaten over the following year, and a generation of children grew up thinking tinned peaches were what you had on special occasions", Reg: rWry, Topic: "food"},
	{Text: "The council held a special meeting to congratulate itself on the preparations, which had not been needed, and voted unanimously to do it all again next time, and the minutes of the meeting were printed in the paper under the odds", Reg: rJoke, Topic: "authority"},
	{Text: "The shopkeepers complain that nobody wants tins any more", Reg: rWry, Form: fComplaint, Topic: "trade"},
	{Text: "The pubs have decided the figure was never that high", Reg: rWry, Form: fOverheard, Topic: "rumour"},
}

var harbSparedDigital = []skel{
	{Text: "The markets rallied hard", Reg: rPlain, Topic: "money"},
	{Text: "The alerts stopped", Reg: rPlain, Topic: "machine"},
	{Text: "Delivery slots reopened", Reg: rPlain, Topic: "trade"},
	{Text: "The probability dropped to nothing on every screen at once", Reg: rPlain, Topic: "count"},
	{Text: "Highways out of the city filled up again, going the other way", Reg: rWry, Topic: "ground"},
	{Text: "People tried to return their stockpiles, and the shops would not take them", Reg: rWry, Topic: "trade"},
	{Text: "The group chats filled with relief and then with photos of lunch", Reg: rWry, Topic: "food"},
	{Text: "Server farms spun back up in rolling sequence", Reg: rPlain, Topic: "machine"},
	{Text: "The companies that had sent staff home announced a return to the office with a speed that suggested they had been waiting for the excuse", Reg: rJoke, Topic: "work"},
	{Text: "The trending topic changed from the warning to the relief to a celebrity divorce inside a single afternoon", Reg: rWry, Topic: "rumour"},
	{Text: "When the danger had passed, the doors of the lower levels opened again in reverse order, and people came up blinking out of the car parks with their emergency kits over their shoulders", Reg: rPlain, Topic: "building"},
	{Text: "The comments complain the whole thing was overblown", Reg: rWry, Form: fComplaint, Topic: "argument"},
	{Text: "Offices will reopen on Monday, the email said", Reg: rPlain, Form: fNotice, Topic: "work"},
	{Text: "The office sweepstake on the date had no winner", Reg: rWry, Form: fLedger, Topic: "money"},
}

var harbSparedCosmic = []skel{
	{Text: "The shields came down", Reg: rPlain, Topic: "machine"},
	{Text: "The outer ring reopened", Reg: rPlain, Topic: "building"},
	{Text: "Every child off the shuttles and accounted for", Reg: rPlain, Form: fLedger, Topic: "family"},
	{Text: "All crew may resume normal duties", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "The estimate fell back to its usual whisper above zero", Reg: rPlain, Topic: "count"},
	{Text: "People went back to the viewports, and looked out, and saw stars", Reg: rPlain, Topic: "omen"},
	{Text: "The hull plating over the viewports was cut away again", Reg: rPlain, Topic: "building"},
	{Text: "The evacuation berths were unassigned without comment", Reg: rWry, Topic: "authority"},
	{Text: "Crews came in from the hull in their pressure suits and stood in the airlock for a long time before taking their helmets off", Reg: rPlain, Topic: "people"},
	{Text: "Whatever had been out there went on past the colony without turning, and the long-range instruments followed it until it was out of range, and nobody looked away until then", Reg: rPlain, Topic: "machine"},
	{Text: "The observation decks were unsealed and people crowded back in to look at the dark where it had been, and the dark looked exactly as it always had, and for a while that was the most beautiful thing any of them had ever seen", Reg: rPlain, Topic: "omen"},
	{Text: "Engineering complains about the wear on the shield emitters", Reg: rWry, Form: fComplaint, Topic: "machine"},
	{Text: "On the ring they are saying it was never going to hit", Reg: rWry, Form: fOverheard, Topic: "rumour"},
}

// --- Discredited ----------------------------------------------------------------

var harbDiscreditedAny = []skel{
	{Text: "The laughing started", Reg: rPlain, Topic: "people"},
	{Text: "All for nothing, then", Reg: rPlain, Topic: "time"},
	{Text: "People say they always had doubts", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "It seems that {subject} had never seen anything at all", Needs: needSubject, Reg: rPlain, Topic: "stranger"},
	{Text: "Everybody remembered the things they had given up and wanted them back", Reg: rWry, Topic: "haul"},
	{Text: "The warning, it emerged, had been made up from beginning to end", Reg: rPlain, Topic: "message"},
	{Text: "The believers kept indoors for a few days", Reg: rPlain, Topic: "people"},
	{Text: "Nobody can say what {subject} hoped to gain, and there are several theories, most of them involving food", Needs: needSubject, Reg: rJoke, Topic: "food"},
	{Text: "Bit by bit it came out that {subject} had given this warning in other places and been wrong there too", Needs: needSubject, Reg: rPlain, Topic: "rumour"},
	{Text: "The frightened were angry afterwards and the doubters were insufferable, and between them they made the next few days harder than any danger would have", Reg: rWry, Topic: "argument"},
	{Text: "People want to know who pays for everything that was done for nothing", Reg: rWry, Form: fComplaint, Topic: "money"},
	{Text: "It has been agreed that nobody will speak of it again", Reg: rJoke, Form: fNotice, Topic: "authority"},
}

// harbDiscreditedEarly spans the ancient and feudal buckets, where nearly all
// false warnings end (the Steel Era's last figure, in the Industrial Age, uses
// the era-neutral pool).
var harbDiscreditedEarly = []skel{
	{Text: "The offerings were taken back", Reg: rPlain, Topic: "religion"},
	{Text: "The fires were let go out in disgust", Reg: rPlain, Topic: "sleep"},
	{Text: "Children made up songs about {subject}, and the songs were rude", Needs: needSubject, Reg: rWry, Topic: "family"},
	{Text: "The buried things were dug up again, and the digging was not cheerful", Reg: rWry, Topic: "haul"},
	{Text: "The crowd that ran {subject} out of the settlement threw whatever came to hand, and some of it was food they could not spare", Needs: needSubject, Reg: rWry, Topic: "people"},
	{Text: "The old women said that the warning had been false but the fear had been real, and that the fear should count for something", Reg: rPlain, Form: fOverheard, Topic: "people"},
	{Text: "For a long time afterwards anyone who predicted anything at all, even rain, was looked at in a way that made them stop halfway through the sentence, and a whole season of weather went unforecast in the settlement because nobody dared", Reg: rJoke, Topic: "weather"},
	{Text: "A rotten turnip was thrown", Reg: rPlain, Topic: "food"},
	{Text: "Offerings returned to their owners, two goats short", Reg: rJoke, Form: fLedger, Topic: "count"},
	{Text: "Everyone who had repeated the warning to somebody else went round afterwards unsaying it, which took longer than saying it had", Reg: rWry, Topic: "rumour"},
	{Text: "A grandmother who had believed every prophet for sixty years announced that she would believe no more of them, and was believed by nobody", Reg: rJoke, Form: fOverheard, Topic: "people"},
	{Text: "The prophecy became a joke", Reg: rPlain, Topic: "rumour"},
}

var harbDiscreditedAncient = []skel{
	{Text: "Embarrassment all round among the elders", Reg: rWry, Topic: "authority"},
	{Text: "The drums played a mocking beat", Reg: rWry, Topic: "noise"},
	{Text: "The songs are unkind", Reg: rPlain, Topic: "noise"},
	{Text: "The ash marks were scrubbed off the doorposts", Reg: rPlain, Topic: "building"},
	{Text: "The herd was let out and came to no harm at all", Reg: rPlain, Topic: "animal"},
	{Text: "The sharpened spears were put away without anyone mentioning them", Reg: rWry, Topic: "weapon"},
	{Text: "Whom to listen to will be chosen more carefully in future, the elders say", Reg: rPlain, Form: fOverheard, Topic: "authority"},
	{Text: "The feathers and charms were thrown on the midden", Reg: rPlain, Topic: "religion"},
	{Text: "Three other camps, it turned out, had already driven {subject} out for doing exactly this", Needs: needSubject, Reg: rPlain, Topic: "stranger"},
	{Text: "The women dug up the grain pits they had moved further up the slope and moved them back down, and said a great deal while doing it", Reg: rWry, Topic: "food"},
	{Text: "The hunters who had stayed in for a moon went out on the first clear morning and found the herds had wandered far off, and walked three days to catch up, cursing {subject} by name the whole way", Needs: needSubject, Reg: rWry, Topic: "animal"},
	{Text: "Complaints from the hunters, who lost a moon of hunting to a liar", Reg: rWry, Form: fComplaint, Topic: "food"},
	{Text: "The elders have forbidden anyone to give food to wandering prophets", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Around the camp they say {subject} ate better than anyone that whole moon", Needs: needSubject, Reg: rJoke, Form: fOverheard, Topic: "food"},
	{Text: "One goat, three feathers and a great deal of dried meat, all wasted on a liar", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "The boy who had been sent to watch the pass came back after a moon and asked whether he could stop now", Reg: rWry, Topic: "border"},
	{Text: "The charms that had been tied to the children's wrists were cut off and burned, and some of the children cried because they had liked them", Reg: rPlain, Topic: "family"},
	{Text: "The herd went back out", Reg: rPlain, Topic: "animal"},
	{Text: "The elders sat at the big fire and argued long into the night about how they had been fooled, and decided at last that the prophet had been convincing, which was a comfort to nobody", Reg: rWry, Form: fOverheard, Topic: "argument"},
	{Text: "The prophet's name is no longer said", Reg: rPlain, Topic: "name"},
	{Text: "The standing stones where the offerings had been left were cleared, and the offerings shared out again, and there were arguments over whose had been whose", Reg: rWry, Topic: "religion"},
	{Text: "A hunter who had given up his best spear points as an offering tried to get them back from the stones and found them gone, and the whole camp has an opinion about who took them, and the prophet is one of the names", Reg: rWry, Topic: "weapon"},
	{Text: "Four moons wasted", Reg: rPlain, Form: fLedger, Topic: "time"},
}

var harbDiscreditedFeudal = []skel{
	{Text: "The bells were silent", Reg: rPlain, Topic: "noise"},
	{Text: "The bishop was displeased", Reg: rWry, Topic: "religion"},
	{Text: "Pilgrims asked for refunds", Reg: rJoke, Form: fComplaint, Topic: "money"},
	{Text: "The priest preached on false prophets, at length", Reg: rWry, Topic: "religion"},
	{Text: "The guild demanded its candle money back", Reg: rWry, Topic: "money"},
	{Text: "The carts that had fled north came back looking sheepish", Reg: rWry, Topic: "trade"},
	{Text: "The lord's bailiff went looking for {subject} with a warrant", Needs: needSubject, Reg: rPlain, Topic: "authority"},
	{Text: "The tavern had a new joke, and it was told all week", Reg: rWry, Topic: "rumour"},
	{Text: "Cattle sold cheap on the strength of the warning stayed sold, and the buyers laughed in the sellers' faces", Reg: rWry, Topic: "trade"},
	{Text: "The abbey's chronicle entered the whole affair under the heading of follies", Reg: rWry, Topic: "paper"},
	{Text: "The false warning was read out in the market square by order of the mayor, so that everybody could hear how foolish it sounded now, and it sounded exactly as frightening as it had the first time, and the reading was cut short", Reg: rWry, Topic: "town"},
	{Text: "The steward complains that he doubled the watch for a liar", Reg: rWry, Form: fComplaint, Topic: "authority"},
	{Text: "It is proclaimed that {subject} is not to be given lodging within the walls", Needs: needSubject, Reg: rPlain, Form: fNotice, Topic: "building"},
	{Text: "Gossip in the market has {subject} in the pay of the grain merchants", Needs: needSubject, Reg: rWry, Form: fOverheard, Topic: "trade"},
	{Text: "Candles returned to the chandler, forty; refunds given, none", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "The monks who had prayed in shifts for a week sent the town a bill for the prayers, and the town refused to pay for prayers against nothing", Reg: rJoke, Topic: "money"},
	{Text: "The friars preached that the false warning had been a test of belief, and that the town had failed it by believing, which confused everyone", Reg: rWry, Topic: "religion"},
	{Text: "The fair was rescheduled", Reg: rPlain, Form: fNotice, Topic: "trade"},
	{Text: "The chalk crosses that had been drawn on every door on the high street were scrubbed off by the next morning, each household doing its own and pretending it had never drawn one, and the stone was cleaner than it had been for years", Reg: rWry, Topic: "building"},
	{Text: "Masses of thanksgiving were said, somewhat sheepishly", Reg: rWry, Topic: "religion"},
	{Text: "The lord came back from the country, where he had fled before anyone else, and made a speech about steadiness in the face of rumour", Reg: rWry, Topic: "authority"},
	{Text: "Wills drawn up in the panic are being torn up now", Reg: rPlain, Topic: "paper"},
	{Text: "Twelve wills, torn up", Reg: rPlain, Form: fLedger, Topic: "paper"},
}

// harbDiscreditedAges holds the per-age Discredited lines. A false warning is
// rolled at an epoch's first age and repeated by every later figure, so the
// Steel Era's can end with the Industrial Age's Newsboy: a printed figure
// that turns out to have been invented.
var harbDiscreditedAges = []ageSet{
	{ages("industrial_age"), []skel{
		{Text: "The paper printed a retraction on page nine", Reg: rPlain, Topic: "paper"},
		{Text: "Nobody bought the late edition the next day", Reg: rPlain, Topic: "trade"},
		{Text: "The figure had been made up", Reg: rPlain, Topic: "count"},
		{Text: "The mills reopened on the Monday", Reg: rPlain, Topic: "work"},
		{Text: "The editor who printed the odds resigned, and was taken on a week later by a rival paper that wanted his readers", Reg: rWry, Topic: "authority"},
		{Text: "It turned out the odds had been set by a compositor with a grudge and a free afternoon", Reg: rJoke, Topic: "machine"},
		{Text: "Families brought their tinned goods back up from the cellars and did not say much about it", Reg: rPlain, Topic: "family"},
		{Text: "The newsboys found a different corner", Reg: rPlain, Topic: "town"},
		{Text: "Letters to the editor ran for a month, each angrier than the last", Reg: rWry, Topic: "message"},
		{Text: "The paper blamed the printers, and the printers blamed the paper", Reg: rWry, Topic: "argument"},
		{Text: "The factory owners who shut early want the lost wages paid back to them", Reg: rWry, Form: fComplaint, Topic: "money"},
		{Text: "The forecast printed in last week's late edition is withdrawn in full", Reg: rPlain, Form: fNotice, Topic: "authority"},
		{Text: "Extra coal, extra flour, extra candles, all bought for nothing", Reg: rWry, Form: fLedger, Topic: "haul"},
		{Text: "The insurance men sent round letters asking for the emergency premiums to be paid anyway", Reg: rWry, Topic: "money"},
		{Text: "In the pubs they say the whole thing was cooked up to sell papers", Reg: rPlain, Form: fOverheard, Topic: "rumour"},
		{Text: "The chapel emptied again", Reg: rPlain, Topic: "religion"},
		{Text: "A crowd stood outside the newspaper offices that evening and threw nothing worse than insults, which the editor printed the next morning in a column headed Readers Respond", Reg: rJoke, Topic: "people"},
		{Text: "Overnight a rude word appeared in chalk across the paper's front window, and the paper put a photograph of it on the front page", Reg: rJoke, Topic: "building"},
		{Text: "The trams ran full again by the end of the week", Reg: rPlain, Topic: "town"},
		{Text: "The mayor, who had ordered the schools shut, now says he never believed a word of it", Reg: rWry, Topic: "authority"},
		{Text: "The usual corner was taken over by a boy selling matches, and nobody saw {subject} there again", Needs: needSubject, Reg: rPlain, Topic: "stranger"},
		{Text: "Somebody paid {subject} to shout it, and the police would like to know who", Needs: needSubject, Reg: rWry, Topic: "authority"},
		{Text: "Word in the pubs is that {subject} sells a different paper now, on a different street, under a different name", Needs: needSubject, Reg: rJoke, Form: fOverheard, Topic: "stranger"},
		{Text: "The shopkeepers who put up their shutters say {subject} owes them a week's takings", Needs: needSubject, Reg: rWry, Form: fComplaint, Topic: "money"},
		{Text: "No one could say who had given {subject} the figure, and the paper did not care to find out", Needs: needSubject, Reg: rPlain, Topic: "message"},
		{Text: "The police took a statement from {subject}, who said only that the figure had seemed about right", Needs: needSubject, Reg: rWry, Topic: "authority"},
		{Text: "The last anyone saw of {subject}, the unsold late editions were going into the canal a bundle at a time", Needs: needSubject, Reg: rJoke, Topic: "paper"},
	}},
}
