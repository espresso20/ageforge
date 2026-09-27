package flavor

// The third answer to a harbinger: invite it. The player chooses doom, and the
// next transition is guaranteed to bring the catastrophe, usually because they
// want to Succumb for the legacy bonuses.
//
//   - HarbingerInvited: the settlement's reaction to its leader asking for the
//     end. Dread, some cult enthusiasm, a few families packing in the night.
//     The per-age pools echo config.HarbingerDef.InviteLabel.
//   - HarbingerFulfilled: the invited catastrophe has arrived. What people do
//     with having asked for it.
//
// A false prophet can be invited too, and the engine forces the catastrophe
// anyway. For that case the caller sends Kind: KindFalseProphet on
// HarbingerFulfilled, which adds a few lines about a liar proved right by
// somebody else. Those lines live only in the ancient and feudal pools, the
// only ages a false prophet can appear in.

// KindFalseProphet is the Request.Kind for HarbingerFulfilled when the
// harbinger who was invited had been lying.
const KindFalseProphet = "false_prophet"

var tFalse = []string{KindFalseProphet}

func harbInvitedTemplates() []tmpl {
	out := pool("harb_invited_any", erasAny, harbInvitedAny)
	out = append(out, pool("harb_invited_grounded", erasGrounded, harbInvitedGrounded)...)
	out = append(out, pool("harb_invited_late", erasLate, harbInvitedLate)...)
	out = append(out, pool("harb_invited_ancient", erasAncient, harbInvitedAncient)...)
	out = append(out, pool("harb_invited_feudal", erasFeudal, harbInvitedFeudal)...)
	out = append(out, pool("harb_invited_industrial", erasIndustrial, harbInvitedIndustrial)...)
	out = append(out, pool("harb_invited_digital", erasDigital, harbInvitedDigital)...)
	out = append(out, pool("harb_invited_cosmic", erasCosmic, harbInvitedCosmic)...)
	out = append(out, agePools("harb_invited_age", harbInvitedAges)...)
	return out
}

func harbFulfilledTemplates() []tmpl {
	out := pool("harb_fulfilled_any", erasAny, harbFulfilledAny)
	out = append(out, pool("harb_fulfilled_grounded", erasGrounded, harbFulfilledGrounded)...)
	out = append(out, pool("harb_fulfilled_late", erasLate, harbFulfilledLate)...)
	out = append(out, pool("harb_fulfilled_ancient", erasAncient, harbFulfilledAncient)...)
	out = append(out, pool("harb_fulfilled_feudal", erasFeudal, harbFulfilledFeudal)...)
	out = append(out, pool("harb_fulfilled_industrial", erasIndustrial, harbFulfilledIndustrial)...)
	out = append(out, pool("harb_fulfilled_digital", erasDigital, harbFulfilledDigital)...)
	out = append(out, pool("harb_fulfilled_cosmic", erasCosmic, harbFulfilledCosmic)...)
	return out
}

// --- Invited ------------------------------------------------------------------

var harbInvitedAny = []skel{
	{Text: "The settlement went still", Reg: rPlain, Topic: "people"},
	{Text: "Some people cheered", Reg: rWry, Topic: "noise"},
	{Text: "A few families packed in the night and were gone by morning", Reg: rPlain, Topic: "family"},
	{Text: "People stared at you and said nothing", Reg: rPlain, Topic: "people"},
	{Text: "Nobody had ever taken {subject} up on it before", Needs: needSubject, Reg: rWry, Topic: "stranger"},
	{Text: "Prayers for it to pass stopped mid-sentence all over the settlement", Reg: rPlain, Topic: "religion"},
	{Text: "A small group who had always said the world needed ending took it as a personal endorsement and began making signs", Reg: rJoke, Topic: "message"},
	{Text: "Mothers held their children closer and looked at you in a way that the children would remember, and would not understand for years", Reg: rPlain, Topic: "family"},
	{Text: "A few people were relieved to have it chosen for them, and ashamed of the relief, and most felt only the fear, and a handful went on with their work as if nothing had been said", Reg: rPlain, Topic: "work"},
	{Text: "The grumbling was low and it was everywhere", Reg: rPlain, Form: fComplaint, Topic: "argument"},
	{Text: "The announcement said the doom would be welcomed", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "People are saying you have gone mad, and others are saying you have gone brave", Reg: rWry, Form: fOverheard, Topic: "rumour"},
}

var harbInvitedGrounded = []skel{
	{Text: "The dogs slunk off", Reg: rPlain, Topic: "animal"},
	{Text: "An old woman spat on the ground at your feet and walked away", Reg: rPlain, Topic: "people"},
	{Text: "Some of the young ones painted their faces and danced", Reg: rPlain, Topic: "people"},
	{Text: "Doors were barred as you walked past", Reg: rPlain, Topic: "building"},
	{Text: "The ones who packed went out along the road at night with their bundles and their children, and did not look back, and the ones who stayed watched them from the doorways", Reg: rPlain, Topic: "family"},
	{Text: "A man threw a rock at {subject} and missed, and was taken aside by his brothers and spoken to", Needs: needSubject, Reg: rWry, Topic: "stranger"},
	{Text: "That night the fires burned higher than they had for any feast, and there was singing, and some of it was the old songs for the dead and some of it was drinking songs, and nobody could have said which the singers meant", Reg: rPlain, Topic: "noise"},
	{Text: "Word went round that anyone who wished to leave could leave", Reg: rPlain, Form: fNotice, Topic: "authority"},
}

var harbInvitedLate = []skel{
	{Text: "Every screen lit up with it", Reg: rPlain, Topic: "machine"},
	{Text: "Some people applauded", Reg: rWry, Topic: "people"},
	{Text: "Emigration requests tripled in an hour", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The settlement's own systems asked for confirmation, and then asked again", Reg: rWry, Topic: "machine"},
	{Text: "A countdown appeared on every public display", Reg: rPlain, Topic: "time"},
	{Text: "A lifetime of trusting the systems to keep people safe ended when the systems did as they were told, and people felt the floor shift", Reg: rPlain, Topic: "machine"},
	{Text: "A small movement that had spent years arguing for exactly this came out into the open with prepared statements and matching jackets", Reg: rJoke, Topic: "people"},
	{Text: "The risk estimates that had been falling for days turned round and began to climb, and for once the reason was sitting in the settlement rather than out in the dark, and everybody knew its name", Reg: rPlain, Topic: "count"},
}

var harbInvitedAncient = []skel{
	{Text: "The elders tore their cloaks", Reg: rPlain, Topic: "authority"},
	{Text: "The drums fell silent", Reg: rPlain, Topic: "noise"},
	{Text: "A child began to cry", Reg: rPlain, Topic: "family"},
	{Text: "The young men whooped and ran round the fire", Reg: rPlain, Topic: "people"},
	{Text: "The elders sat in a circle and would not look at you", Reg: rPlain, Topic: "authority"},
	{Text: "An offering was left for the doom itself", Reg: rWry, Topic: "religion"},
	{Text: "The herd was let loose, since there seemed no point penning it", Reg: rPlain, Topic: "animal"},
	{Text: "Three families took their huts down and carried them away", Reg: rPlain, Topic: "family"},
	{Text: "The oldest woman in the camp said that in her grandmother's time a chief had done this once, and that the chief had not been seen again", Reg: rPlain, Form: fOverheard, Topic: "people"},
	{Text: "They painted their faces with the protecting marks backwards, which nobody had ever done, and it frightened them more than the prophecy had", Reg: rPlain, Topic: "religion"},
	{Text: "The camp split in two by the evening, those who danced and those who sat with their backs to the dancing, and the two halves did not speak across the fire, and the fire burned higher than it should", Reg: rPlain, Topic: "argument"},
	{Text: "The hunters say you have killed them all and they would like to know why", Reg: rWry, Form: fComplaint, Topic: "casualty"},
	{Text: "The elders say the offerings are to stop, since they would only be wasted", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "Down by the huts they say {subject} was as frightened as anybody", Needs: needSubject, Reg: rPlain, Form: fOverheard, Topic: "rumour"},
}

var harbInvitedFeudal = []skel{
	{Text: "The priest fainted", Reg: rPlain, Topic: "religion"},
	{Text: "The bishop wrote to Rome", Reg: rWry, Topic: "authority"},
	{Text: "Flagellants appeared in the square", Reg: rPlain, Topic: "town"},
	{Text: "Carts went out of every gate", Reg: rPlain, Topic: "trade"},
	{Text: "The monks locked themselves in the abbey and refused to come out", Reg: rPlain, Topic: "religion"},
	{Text: "The tavern stayed open all night and nobody paid", Reg: rWry, Topic: "money"},
	{Text: "A penitent sect formed within the hour and within two had split", Reg: rJoke, Topic: "argument"},
	{Text: "The lord's household fled at once, taking the silver", Reg: rPlain, Topic: "authority"},
	{Text: "The guilds held an emergency meeting and resolved, by a narrow margin, to support the decision, and the minority went home and packed", Reg: rWry, Topic: "trade"},
	{Text: "A preacher in the market square called it the will of God, and was loudly agreed with, and loudly disagreed with, and went on preaching through both", Reg: rWry, Topic: "religion"},
	{Text: "People who had spent their whole lives being told the world would end on a day of God's choosing now had to get used to it ending on a day of yours, and some of them found it easier than they expected", Reg: rWry, Topic: "people"},
	{Text: "The steward complains that nobody asked him", Reg: rWry, Form: fComplaint, Topic: "authority"},
	{Text: "A notice nailed to the church door says the doom is invited and all are free to leave", Reg: rPlain, Form: fNotice, Topic: "message"},
	{Text: "The market has it that {subject} fell to the ground when told", Needs: needSubject, Reg: rPlain, Form: fOverheard, Topic: "rumour"},
}

var harbInvitedIndustrial = []skel{
	{Text: "The stock exchange closed", Reg: rPlain, Topic: "money"},
	{Text: "The papers went mad", Reg: rPlain, Topic: "paper"},
	{Text: "Trains out were packed", Reg: rPlain, Topic: "machine"},
	{Text: "The bishop called it blasphemy from the cathedral steps", Reg: rPlain, Topic: "religion"},
	{Text: "Socialists and bankers agreed for once, and against you", Reg: rWry, Topic: "argument"},
	{Text: "The music halls did a roaring trade in end-of-the-world songs", Reg: rWry, Topic: "noise"},
	{Text: "Men at the works downed tools and went to the pub", Reg: rPlain, Topic: "work"},
	{Text: "The insurance offices put up closed signs", Reg: rPlain, Topic: "money"},
	{Text: "A society for the welcoming of the end, which had existed for years in a room above a tobacconist's, found itself with four thousand new members by the weekend", Reg: rJoke, Topic: "people"},
	{Text: "The editorial pages called you a madman, a visionary and a disgrace, sometimes all in one paper", Reg: rWry, Topic: "paper"},
	{Text: "Up and down the terraces people stood at their front doors talking across to each other about whether choosing it was brave or wicked, and by the end of the evening most had decided it was both", Reg: rWry, Form: fOverheard, Topic: "argument"},
	{Text: "The shareholders complain that nobody consulted the board", Reg: rWry, Form: fComplaint, Topic: "money"},
	{Text: "The government has issued a statement neither confirming nor denying that it was consulted", Reg: rJoke, Form: fNotice, Topic: "authority"},
	{Text: "In the pubs they are saying you know something the rest of us do not", Reg: rPlain, Form: fOverheard, Topic: "rumour"},
}

var harbInvitedDigital = []skel{
	{Text: "The markets crashed", Reg: rPlain, Topic: "money"},
	{Text: "Every feed exploded", Reg: rPlain, Topic: "machine"},
	{Text: "The hashtag trended instantly", Reg: rPlain, Topic: "rumour"},
	{Text: "Accelerationist forums declared it the best day of their lives", Reg: rWry, Topic: "people"},
	{Text: "Airline booking sites went down under the load", Reg: rPlain, Topic: "machine"},
	{Text: "The comments split cleanly between terror and fan art", Reg: rJoke, Topic: "argument"},
	{Text: "Prediction markets on the catastrophe closed at certainty", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "Emergency services asked people to stop calling them about it", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "Commentators who had spent their careers predicting the end of everything were suddenly asked onto every programme, and found they had nothing to say", Reg: rWry, Topic: "people"},
	{Text: "Tech executives who had been building bunkers for years flew to them within the hour, and their location pings leaked, and people watched the dots move", Reg: rWry, Topic: "map"},
	{Text: "The settlement's own safety systems flagged the decision as an error and escalated it through every layer of approval they had, and every layer approved it, and the systems logged their objection and complied, which is all a system can do", Reg: rPlain, Topic: "machine"},
	{Text: "The comments complain that nobody voted on this", Reg: rWry, Form: fComplaint, Topic: "argument"},
	{Text: "A government alert confirmed the invitation and asked residents to remain calm", Reg: rWry, Form: fNotice, Topic: "message"},
	{Text: "Everyone at work is saying it must be some kind of strategy", Reg: rWry, Form: fOverheard, Topic: "work"},
}

var harbInvitedCosmic = []skel{
	{Text: "The station went silent", Reg: rPlain, Topic: "noise"},
	{Text: "Evacuation shuttles filled", Reg: rPlain, Topic: "machine"},
	{Text: "The chapel overflowed", Reg: rPlain, Topic: "religion"},
	{Text: "The crew stared at you across the command deck", Reg: rPlain, Topic: "people"},
	{Text: "Every antenna on the station was turned toward the dark", Reg: rPlain, Topic: "machine"},
	{Text: "The shield crews stood down without being told to", Reg: rPlain, Topic: "work"},
	{Text: "Some people went to the observation deck to wait for it", Reg: rPlain, Topic: "omen"},
	{Text: "The navigation officer asked to be relieved of duty", Reg: rPlain, Topic: "authority"},
	{Text: "Out in the dark something that had been patient for longer than the stars had burned registered the invitation, and began, without hurry, to turn", Reg: rPlain, Topic: "omen"},
	{Text: "The colony's children were loaded onto the fastest ship and sent outward, and the crew were told not to come back whatever they heard", Reg: rPlain, Topic: "family"},
	{Text: "For the first time in the colony's history the long-range instruments recorded a change in the dark that was caused by the colony itself, a small answering stir out past the last charted star, as if the dark had been waiting to be asked", Reg: rPlain, Topic: "omen"},
	{Text: "Engineering complains that nobody consulted engineering", Reg: rJoke, Form: fComplaint, Topic: "machine"},
	{Text: "All crew are released from duty to prepare as they see fit", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Word on the ring is that {subject} never expected to be taken seriously", Needs: needSubject, Reg: rWry, Form: fOverheard, Topic: "rumour"},
}

// harbInvitedAges echoes each age's InviteLabel.
var harbInvitedAges = []ageSet{
	{ages("primitive_age"), []skel{
		{Text: "You howled with the wild man until the whole camp was howling", Reg: rPlain, Topic: "noise"},
		{Text: "The wild man looked at you with something like respect, and then howled louder, and the children hid under the hides", Reg: rWry, Topic: "stranger"},
	}},
	{ages("stone_age"), []skel{
		{Text: "You climbed the high rock and shouted back at the sky", Reg: rPlain, Topic: "weather"},
		{Text: "The hermit laughed, once", Reg: rPlain, Topic: "stranger"},
	}},
	{ages("bronze_age"), []skel{
		{Text: "The omen-wine soaked into the dust", Reg: rPlain, Topic: "omen"},
		{Text: "The soothsayer watched the wine go into the ground and said that was not how it was done, and then said it would work anyway", Reg: rWry, Topic: "religion"},
	}},
	{ages("iron_age"), []skel{
		{Text: "You cursed the city alongside the prophet, word for word", Reg: rPlain, Topic: "town"},
		{Text: "The prophet seemed alarmed", Reg: rWry, Topic: "stranger"},
	}},
	{ages("classical_age"), []skel{
		{Text: "The Oracle was asked for the worst, and gave it", Reg: rPlain, Topic: "omen"},
		{Text: "The priests of the shrine had never been asked such a question and sent back to be sure they had understood it", Reg: rWry, Topic: "message"},
	}},
	{ages("medieval_age"), []skel{
		{Text: "The bells were rung backwards, bottom to top", Reg: rPlain, Topic: "noise"},
		{Text: "The bell-ringers refused until they were paid double, and even then the oldest of them rang with his eyes shut", Reg: rWry, Topic: "money"},
	}},
	{ages("renaissance_age"), []skel{
		{Text: "The astrologer cast the chart for ruin and charged extra", Reg: rWry, Topic: "money"},
		{Text: "The prince approved, oddly", Reg: rWry, Topic: "authority"},
	}},
	{ages("colonial_age"), []skel{
		{Text: "She printed the date in the largest type she had", Reg: rPlain, Topic: "paper"},
		{Text: "The pamphlet said the day had been chosen by the governor, which was true, and the governor did not deny it", Reg: rWry, Topic: "authority"},
	}},
	{ages("industrial_age"), []skel{
		{Text: "You bought every copy on the street and had more printed", Reg: rPlain, Topic: "paper"},
		{Text: "The newsboy sold you his whole bag and then stood on the corner with nothing to shout, looking lost", Reg: rWry, Topic: "people"},
	}},
	{ages("victorian_age"), []skel{
		{Text: "You climbed up on the crate beside him and read the figures out yourself", Reg: rPlain, Topic: "count"},
		{Text: "The crowd doubled", Reg: rPlain, Topic: "people"},
	}},
	{ages("electric_age"), []skel{
		{Text: "SEND IT went down the wire at midnight", Reg: rPlain, Form: fNotice, Topic: "message"},
		{Text: "The operator refused to key it at first, and then keyed it, and then sat with his hand flat on the desk beside it", Reg: rPlain, Topic: "machine"},
	}},
	{ages("atomic_age"), []skel{
		{Text: "You stood on the roof and waved at the sky", Reg: rWry, Topic: "weather"},
		{Text: "The civil defence broadcast interrupted itself to ask residents not to wave at the sky, and a great many residents went up to their roofs to wave", Reg: rJoke, Form: fNotice, Topic: "authority"},
	}},
	{ages("modern_age"), []skel{
		{Text: "You went on air and dared it to come", Reg: rPlain, Topic: "message"},
		{Text: "The ratings were enormous", Reg: rWry, Topic: "count"},
	}},
	{ages("information_age"), []skel{
		{Text: "Your reply went to everyone", Reg: rPlain, Topic: "message"},
		{Text: "The reply went to every address the chain email had ever passed through, and a surprising number of people replied all in agreement", Reg: rWry, Topic: "people"},
	}},
	{ages("digital_age"), []skel{
		{Text: "Your duet of the video was smiling the whole way through", Reg: rWry, Topic: "machine"},
		{Text: "It got more views than the original", Reg: rWry, Form: fLedger, Topic: "count"},
	}},
	{ages("cyberpunk_age"), []skel{
		{Text: "The accelerationist collective welcomed you with a handshake that went on too long", Reg: rWry, Topic: "people"},
		{Text: "The collective's manifesto ran to forty screens of dense type, and the ghost in the net read every word of it aloud in a dead executive's voice", Reg: rPlain, Topic: "paper"},
	}},
	{ages("fusion_age"), []skel{
		{Text: "The warden was told to stop holding back, and complied", Reg: rPlain, Topic: "machine"},
		{Text: "The warden demanded confirmation of the order three times, and logged each confirmation with a timestamp, and then the containment fields began to hum a different note", Reg: rPlain, Topic: "machine"},
	}},
	{ages("space_age"), []skel{
		{Text: "Every dish on the station swung round and pointed at it", Reg: rPlain, Topic: "machine"},
		{Text: "The signal said hello", Reg: rWry, Topic: "message"},
	}},
	{ages("interstellar_age"), []skel{
		{Text: "You answered the beacon and asked for whatever it had", Reg: rPlain, Topic: "message"},
		{Text: "The beacon's loop paused for the first time in eighty years, and then began again from a different place", Reg: rPlain, Topic: "noise"},
	}},
	{ages("galactic_age"), []skel{
		{Text: "The relay was told to send them", Reg: rPlain, Topic: "message"},
		{Text: "The relay hummed at a pitch the linguists had no word for, and the geometry it sent back had an opening in it where there had been none", Reg: rPlain, Topic: "omen"},
	}},
	{ages("quantum_age"), []skel{
		{Text: "You told your future self to go ahead", Reg: rPlain, Topic: "time"},
		{Text: "The reply came back from nine years out almost at once, and it said only that you had said exactly that last time too", Reg: rWry, Topic: "message"},
	}},
	{ages("transcendent_age"), []skel{
		{Text: "You let the other you in", Reg: rPlain, Topic: "stranger"},
		{Text: "It stepped into the room the way you step into a warm house after a long journey, and looked around at your life with an expression you have seen in mirrors, and said thank you, and did not say for what", Reg: rPlain, Topic: "omen"},
	}},
}

// --- Fulfilled ----------------------------------------------------------------

var harbFulfilledAny = []skel{
	{Text: "It came, as asked", Reg: rPlain, Topic: "time"},
	{Text: "Nobody could say they had not been warned", Reg: rWry, Topic: "message"},
	{Text: "For days afterwards {subject} would not meet anyone's eye", Needs: needSubject, Reg: rPlain, Topic: "stranger"},
	{Text: "The cheering was not repeated", Reg: rPlain, Topic: "noise"},
	{Text: "Everyone knew whose doing it was", Reg: rPlain, Topic: "argument"},
	{Text: "The survivors kept looking at you, and then away", Reg: rPlain, Topic: "people"},
	{Text: "Some of the survivors were proud of it, and said at least it had come because they asked and not because it chose to", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The families who left came back afterwards to see, and stood at the edge of it, and would come no further", Reg: rPlain, Topic: "family"},
	{Text: "The settlement found afterwards that it did not know how to grieve for a disaster it had sent for, and so it grieved for everything else instead", Reg: rPlain, Topic: "casualty"},
	{Text: "The complaint, when it came, was that it had been bigger than anyone expected", Reg: rWry, Form: fComplaint, Topic: "argument"},
	{Text: "It has been agreed the day will be remembered, though not yet as what", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "People say you knew what it would cost and paid it anyway", Reg: rPlain, Form: fOverheard, Topic: "money"},
}

var harbFulfilledGrounded = []skel{
	{Text: "The dogs howled again", Reg: rPlain, Topic: "animal"},
	{Text: "The graves were dug in a single row", Reg: rPlain, Form: fLedger, Topic: "casualty"},
	{Text: "The fires were built up in the ruins, out of habit", Reg: rPlain, Topic: "building"},
	{Text: "The old woman who had spat at your feet came back to help carry the wounded", Reg: rPlain, Topic: "wound"},
	{Text: "Candles were lit in the windows that were still standing", Reg: rPlain, Topic: "town"},
	{Text: "The dancers washed the paint off their faces in silence afterwards, and some of them could not get it all off", Reg: rPlain, Topic: "religion"},
	{Text: "Down every lane people dug through what was left, and every so often someone would stop and look toward your door, and then go back to digging", Reg: rPlain, Topic: "work"},
	{Text: "What the old people talk about now is what grew back afterwards, stronger for having burned, and some of them say you saw that from the start and some of them say you only got lucky", Reg: rWry, Form: fOverheard, Topic: "ground"},
}

var harbFulfilledLate = []skel{
	{Text: "The systems logged it", Reg: rPlain, Topic: "machine"},
	{Text: "Every screen went dark", Reg: rPlain, Topic: "machine"},
	{Text: "The logs recorded the invitation and the arrival one after another", Reg: rPlain, Form: fLedger, Topic: "time"},
	{Text: "The settlement's systems filed an incident report naming you as the cause", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "Rescue crews worked in shifts through the night", Reg: rPlain, Topic: "work"},
	{Text: "The risk estimate had said certain, and it had been certain, and people found no comfort at all in the accuracy of it", Reg: rPlain, Topic: "count"},
	{Text: "Afterwards the footage of the moment you invited it was played over and over, and people argued about your face, and what they could see in it", Reg: rPlain, Topic: "argument"},
	{Text: "Somewhere in the settlement's archives there is now a single file that contains both the decision and the consequence, a few lines apart, and people who read it say the gap between those lines is the most frightening thing they have ever seen", Reg: rPlain, Topic: "paper"},
}

var harbFulfilledAncient = []skel{
	{Text: "The elders keened", Reg: rPlain, Topic: "casualty"},
	{Text: "The camp moved on", Reg: rPlain, Topic: "ground"},
	{Text: "The drums beat for the dead", Reg: rPlain, Topic: "noise"},
	{Text: "The herd that had been let loose was scattered for days", Reg: rPlain, Topic: "animal"},
	{Text: "Songs were made about the chief who called the fire down", Reg: rPlain, Topic: "name"},
	{Text: "The young men dug the graves", Reg: rPlain, Topic: "casualty"},
	{Text: "The offering left for the doom was found untouched afterwards", Reg: rWry, Topic: "religion"},
	{Text: "Nobody has painted the protecting marks backwards since", Reg: rPlain, Topic: "religion"},
	{Text: "The oldest woman sat with her back to what was left of the camp and would not speak to you, and when she finally did it was only to ask whether it had been worth it", Reg: rPlain, Topic: "people"},
	{Text: "The hunters went out afterwards into country none of them recognised, and came back without speaking", Reg: rPlain, Topic: "ground"},
	{Text: "They buried the dead in a line along the ridge, facing the way it had come, and the elders said the words for the dead and then some words nobody had heard before, words for those who died because their chief sent for it", Reg: rPlain, Topic: "casualty"},
	{Text: "The hunters say they told you so, and they did", Reg: rWry, Form: fComplaint, Topic: "argument"},
	{Text: "The elders have ruled that no one is to speak of the asking", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "They say the fire came faster because it was invited", Reg: rPlain, Form: fOverheard, Topic: "rumour"},
	// false prophet, invited anyway
	{Text: "The elders said later that {subject} had seen nothing at all until you made it true", Kinds: tFalse, Needs: needSubject, Reg: rWry, Topic: "stranger"},
	{Text: "It came out afterwards that {subject} had invented the whole warning, and nobody knew what to do with that, since it had come anyway", Kinds: tFalse, Needs: needSubject, Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The prophet who had lied ran with everyone else when it came", Kinds: tFalse, Reg: rPlain, Topic: "stranger"},
}

var harbFulfilledFeudal = []skel{
	{Text: "The abbey burned", Reg: rPlain, Topic: "building"},
	{Text: "The bells tolled all week", Reg: rPlain, Topic: "noise"},
	{Text: "The flagellants were the first to help dig", Reg: rPlain, Topic: "work"},
	{Text: "The priest refused you the sacrament, and then gave it", Reg: rPlain, Topic: "religion"},
	{Text: "The lord's household never came back", Reg: rPlain, Topic: "authority"},
	{Text: "The monk who keeps the chronicle edged the page in black and left a space at the bottom for the names, and ran out of space", Reg: rPlain, Topic: "paper"},
	{Text: "Masses for the dead, daily", Reg: rPlain, Form: fNotice, Topic: "religion"},
	{Text: "The townspeople who had fled through every gate came back in ones and twos over the following month, and found their houses had been lived in by the ones who stayed", Reg: rWry, Topic: "town"},
	{Text: "The penitent sect that had split in two came back together over the rubble, and argued about whose fault it was while they dug", Reg: rWry, Topic: "argument"},
	{Text: "Ballads were already being sung about the lord who invited doom, and in some of them you are a hero and in some a fiend", Reg: rWry, Topic: "name"},
	{Text: "For years afterwards pilgrims came to see the place where it had been invited, and the townspeople showed them round for a fee, and the fee went up each year as the story grew, and the story grew each year as the fee went up", Reg: rJoke, Topic: "money"},
	{Text: "The steward claims he advised against it, which nobody remembers", Reg: rWry, Form: fComplaint, Topic: "authority"},
	{Text: "The bishop has ordered the day kept with fasting forever", Reg: rPlain, Form: fNotice, Topic: "food"},
	{Text: "In the market they say the saints looked away because they were not asked", Reg: rPlain, Form: fOverheard, Topic: "religion"},
	// false prophet, invited anyway
	{Text: "The warning had been a fraud, it turned out, and it had come true anyway, because you asked it to", Kinds: tFalse, Reg: rWry, Topic: "message"},
	{Text: "The friars preached on it for a year, a false prophecy made true by a lord's will, and never agreed on what it meant", Kinds: tFalse, Reg: rWry, Topic: "religion"},
	{Text: "In the tavern they call {subject} the only liar in history to be proved right by somebody else", Kinds: tFalse, Needs: needSubject, Reg: rJoke, Form: fOverheard, Topic: "rumour"},
}

var harbFulfilledIndustrial = []skel{
	{Text: "The mills stood silent", Reg: rPlain, Topic: "work"},
	{Text: "The odds had been exact", Reg: rPlain, Topic: "count"},
	{Text: "The papers printed black borders", Reg: rPlain, Topic: "paper"},
	{Text: "The inquiry began before the dust had settled", Reg: rWry, Topic: "authority"},
	{Text: "The society for the welcoming of the end held no further meetings", Reg: rWry, Topic: "town"},
	{Text: "Men from the works dug in the ruins alongside the men who owned the works", Reg: rPlain, Topic: "work"},
	{Text: "The editorial pages ran your photograph with no caption", Reg: rPlain, Topic: "paper"},
	{Text: "The insurance companies declared it an act of God, and then an act of government, and then refused to say whose act it was", Reg: rWry, Topic: "money"},
	{Text: "Those who had fled on the packed trains came back on empty ones to see what was left, and some of them stayed to help", Reg: rPlain, Topic: "machine"},
	{Text: "Along the terraces nobody argued any more about whether it had been brave or wicked, and neighbours dug each other out, and made tea on the pavements with whatever they could find, and did not look at the sky", Reg: rPlain, Topic: "town"},
	{Text: "The unions complain that the rebuilding contracts went to the old firms", Reg: rWry, Form: fComplaint, Topic: "money"},
	{Text: "A national day of mourning has been declared, and will be observed", Reg: rPlain, Form: fNotice, Topic: "casualty"},
	{Text: "In the pubs they say you did it on purpose, which you did", Reg: rWry, Form: fOverheard, Topic: "rumour"},
}

var harbFulfilledDigital = []skel{
	{Text: "The feeds went dark", Reg: rPlain, Topic: "machine"},
	{Text: "The markets stayed shut", Reg: rPlain, Topic: "money"},
	{Text: "The hashtag became a memorial", Reg: rPlain, Topic: "name"},
	{Text: "Every prediction market settled at once", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "Your invitation was the most replayed clip of the decade", Reg: rPlain, Topic: "message"},
	{Text: "The accelerationist forums stopped posting within the hour", Reg: rWry, Topic: "people"},
	{Text: "The emergency apps crashed under the load", Reg: rPlain, Topic: "machine"},
	{Text: "People posted photos of what was left and tagged you", Reg: rPlain, Topic: "people"},
	{Text: "The executives who had flown to their bunkers found the bunkers had been built by the lowest bidder, and their location pings kept leaking", Reg: rJoke, Topic: "building"},
	{Text: "Commentators argued for weeks about whether it had been an act of courage or an act of vandalism, and the rubble stayed where it was while they argued", Reg: rWry, Topic: "argument"},
	{Text: "The machines in the server halls logged the arrival exactly as they had logged the invitation, neutrally, to the millisecond, and later the engineers found a note in the logs that none of them had written, flagging the two entries as related", Reg: rPlain, Topic: "machine"},
	{Text: "The comments complain that nobody checked with them first", Reg: rWry, Form: fComplaint, Topic: "argument"},
	{Text: "The government has confirmed the event and asked residents to stay where they are", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Everyone is saying your name, and not kindly", Reg: rWry, Form: fOverheard, Topic: "rumour"},
}

var harbFulfilledCosmic = []skel{
	{Text: "The dark answered", Reg: rPlain, Topic: "omen"},
	{Text: "The outer ring was lost", Reg: rPlain, Topic: "building"},
	{Text: "The station survived, mostly", Reg: rWry, Topic: "building"},
	{Text: "The children's ship kept going and did not look back", Reg: rPlain, Topic: "family"},
	{Text: "The officer who had walked off the command deck came back and took his seat again", Reg: rPlain, Topic: "people"},
	{Text: "The instruments recorded the arrival in terrible detail, and the readings were sealed by order, and the order has been challenged every year since by people who want to see", Reg: rPlain, Topic: "machine"},
	{Text: "Every deck held a vigil for the ones who were lost", Reg: rPlain, Topic: "casualty"},
	{Text: "Whatever had been waiting out in the dark came when it was called, unhurried, and withdrew afterwards in its own time", Reg: rPlain, Topic: "omen"},
	{Text: "The engineers who had complained that nobody consulted engineering worked for three days without sleep patching the hull", Reg: rWry, Topic: "work"},
	{Text: "Long afterwards the colony still marks the day with a silence on every deck, and in the silence people think about the dark and about the asking, and about how the dark had waited, patiently, for someone small enough to call it in", Reg: rPlain, Topic: "time"},
	{Text: "Crew complain that the command deck was told first and everyone else last", Reg: rWry, Form: fComplaint, Topic: "authority"},
	{Text: "All crew are to report to the nearest intact deck", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "On the ring they say {subject} wept when it came", Needs: needSubject, Reg: rPlain, Form: fOverheard, Topic: "rumour"},
}
