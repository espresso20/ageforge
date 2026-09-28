package flavor

// WarRaid — a civilization you are at war with came over the wall and took
// something.
//
// The log line directly above this one already names the raider, names the
// resource and gives the number. So nothing here says what was taken or how
// much. These sentences are the morning after: what burnt and what stood next
// to it, who saw them coming and who slept through it, the state of the outer
// wall, what they left in the mud, the counting, somebody's child who is
// missing and somebody's child who is not, and the rebuilding that started
// before the ash went cold. See catalog.go for the rules.
//
// This is the grimmest Moment in the catalog and it carries the smallest
// humour budget of the five — under one line in twelve reaches for anything,
// and the flat lines are what make those land. Grim is not the same as
// melodramatic. The strongest sentences here are the plainest: what a person
// did with their hands the next morning.

// --- slot banks --------------------------------------------------------------
//
// Noun phrases, lowercase, one tight category per bank, no era-coded nouns:
// both are drawn by ungated sentences and so have to be true of a stone camp
// and of a ring station alike. That rules out more than the lint catches —
// nothing here is a cart, a hearth or a window shutter.

var warRaidBanks = map[string][]string{
	// Things they broke getting in, getting out, or apparently for the exercise.
	"war_raid_damage": {
		"the outer wall", "the north door", "the water store", "the door frame",
		"the light over the entrance", "the long ladder",
		"the lock on the store room", "the lock on the main door",
		"the ceiling of the passage", "the covered way down to the water",
		"the seals around the inner door", "the lights along the passage",
		"the steps down to the store", "the panel beside the entrance",
		"the wall at the far end", "the floor of the front room",
		"the hinges on both doors",
	},
	// Small things they dropped, abandoned or left on purpose. Objects only —
	// something a person can pick up, carry inside and put on a table.
	"war_raid_trace": {
		"a broken knife", "a boot with no laces", "a torn strip of cloth",
		"a coil of good rope", "a child's shoe", "a torn glove",
		"an empty water flask", "a helmet with the strap cut",
		"a knife with the tip snapped off", "a bundle nobody will open",
		"a length of chain", "a bag of somebody else's food",
		"a flask still half full", "a strap with teeth marks in it",
		"a bloodied rag", "a cap two sizes too big", "a ring on a string",
		"a burnt glove",
	},
}

// --- the sentences -----------------------------------------------------------

// warRaidTemplates is the WarRaid skeleton set: one ungated pool that lands in
// any age, plus one pool per era bucket for the imagery that does not travel.
func warRaidTemplates() []tmpl {
	out := pool("war_raid_any", erasAny, warRaidAny)
	out = append(out, pool("war_raid_grounded", erasGrounded, warRaidGrounded)...)
	out = append(out, pool("war_raid_late", erasLate, warRaidLate)...)
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
	// --- short. A fact, and then the sentence stops. ---
	{Text: "The outer wall is open", Reg: rPlain, Topic: "building"},
	{Text: "Two children have not been found", Reg: rPlain, Topic: "casualty"},
	{Text: "It rained on and off", Reg: rPlain, Topic: "weather"},
	{Text: "Sen has gone to sit with her sister", Reg: rPlain, Topic: "name"},
	{Text: "Everything smells of wet ash", Reg: rPlain, Topic: "smell"},
	{Text: "Somebody bit through a lip", Reg: rPlain, Topic: "wound"},
	{Text: "A child drew the men", Reg: rPlain, Topic: "people"},
	{Text: "It was loud, then quiet", Reg: rPlain, Topic: "noise"},
	{Text: "Four doors need new frames", Reg: rPlain, Form: fLedger, Topic: "building"},

	// --- mid, with a slot. What they broke on the way through. ---
	{Text: "The mending starts with ~ and gets worse from there", Slot: "war_raid_damage", Reg: rPlain, Topic: "building"},
	{Text: "Two men worked on ~ until it was too dark to see", Slot: "war_raid_damage", Reg: rPlain, Topic: "people"},
	{Text: "Nobody looked at ~ until the middle of the morning", Slot: "war_raid_damage", Reg: rPlain, Topic: "time"},
	{Text: "The children have been kept away from ~", Slot: "war_raid_damage", Reg: rPlain, Topic: "town"},
	{Text: "What was done to ~ took time and both hands", Slot: "war_raid_damage", Reg: rWry, Topic: "argument"},
	{Text: "Somebody has already started on ~ without being asked", Slot: "war_raid_damage", Reg: rWry, Topic: "building"},
	{Text: "There is an argument going on about who pays for ~", Slot: "war_raid_damage", Reg: rWry, Form: fComplaint, Topic: "money"},
	{Text: "Nothing can be carried in or out until somebody sees to ~", Slot: "war_raid_damage", Reg: rWry, Form: fComplaint, Topic: "haul"},

	// --- mid, with a slot. What they left in the mud. ---
	{Text: "In the doorway they left ~ and went on", Slot: "war_raid_trace", Reg: rPlain, Topic: "haul"},
	{Text: "Nobody has claimed ~ and it has been sitting there two days", Slot: "war_raid_trace", Reg: rPlain, Topic: "kit"},

	// --- mid. The working length: one thing, with enough detail to place it. ---
	{Text: "Everyone says they were asleep and about half of them were", Reg: rWry, Form: fOverheard, Topic: "sleep"},
	{Text: "A girl of about seven bit one of them on the hand", Reg: rPlain, Topic: "people"},
	{Text: "A mark was left on the door and it is meant to be read", Reg: rWry, Topic: "message"},
	{Text: "The quiet afterwards went on for most of an hour", Reg: rWry, Topic: "noise"},
	{Text: "The people two doors along are packing what they have left", Reg: rPlain, Topic: "town"},
	{Text: "The story going round already has forty men in it", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "A man called Hest went out to look and has not come back", Reg: rPlain, Topic: "name"},

	// --- long. Circumstantial, subordinated, the way a chronicle runs. ---
	{Text: "On the table by the door there is ~, and nobody has touched it since it was carried in", Slot: "war_raid_trace", Reg: rPlain, Topic: "kit"},
	{Text: "Two people who were nowhere near the door that night have resigned from things they had appointed themselves to, and a third is expected to follow", Reg: rJoke, Topic: "authority"},
	{Text: "There is talk of going out after them, and the talk has names in it now, and the older ones have started sitting in on it", Reg: rWry, Form: fOverheard, Topic: "argument"},
	{Text: "The blame has settled, without anybody deciding it, on the one who sleeps nearest the door, and he has stopped coming out at meals", Reg: rWry, Topic: "argument"},

	// --- very long. Digressive, specific, a paragraph that forgot to stop. ---
	{Text: "The count of what is gone was started twice and abandoned twice, once because the woman doing it was called away and once because the man who took over could not read her marks, and it is being started again this evening", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "The men who came up from the far end an hour too late have spent all day explaining where they were, and the explanation has got shorter each time, and nobody has asked for it since noon", Reg: rWry, Form: fComplaint, Topic: "people"},

	// --- kind: aggressive ---
	{Text: "It happened in broad daylight", Kinds: []string{"aggressive"}, Reg: rPlain, Topic: "time"},
	{Text: "Nothing about the way they came in was quiet or quick", Kinds: []string{"aggressive"}, Reg: rPlain, Topic: "noise"},
	{Text: "They stood in the open long enough to be counted", Kinds: []string{"aggressive"}, Reg: rWry, Topic: "border"},

	// --- kind: isolationist ---
	{Text: "Nobody here knew their faces", Kinds: []string{"isolationist"}, Reg: rPlain, Topic: "stranger"},
	{Text: "They spoke to nobody and took only what two men could carry", Kinds: []string{"isolationist"}, Reg: rWry, Topic: "haul"},
	{Text: "None of the old people can put a name to the marks on the door", Kinds: []string{"isolationist"}, Reg: rPlain, Form: fOverheard, Topic: "message"},
	{Text: "They went through the place without once raising their voices, which is being talked about more than the damage is", Kinds: []string{"isolationist"}, Reg: rWry, Form: fOverheard, Topic: "rumour"},

	// --- tone ---
	{Text: "The blood has been washed twice", Tones: []Tone{Grim}, Reg: rPlain, Topic: "wound"},
	{Text: "Two of the dead were laid out in the same room as the living", Tones: []Tone{Grim}, Reg: rPlain, Topic: "casualty"},

	// --- the raider's name, in frames a plural or singular name survives ---
	{Text: "Somebody named {subject} out loud", Needs: needSubject, Reg: rPlain, Topic: "name"},
	{Text: "The talk is all of {subject}", Needs: needSubject, Reg: rPlain, Form: fOverheard, Topic: "rumour"},
	{Text: "A name has been scratched into the door frame, and the name is {subject}", Needs: needSubject, Reg: rPlain, Topic: "message"},
	{Text: "The children have learned to spell {subject} this week", Needs: needSubject, Reg: rWry, Topic: "town"},
	{Text: "Nobody here ever traded with {subject}, and nobody here will", Needs: needSubject, Reg: rWry, Topic: "trade"},
	{Text: "Somebody in the back room keeps saying {subject} over and over", Needs: needSubject, Reg: rWry, Form: fOverheard, Topic: "noise"},
	{Text: "The old quarrel with {subject} got a great deal worse last night", Needs: needSubject, Reg: rPlain, Topic: "argument"},
	{Text: "Two of the wounded said {subject} before they went under, and a third said a word nobody here knows, and it has been written down as best anyone could manage", Needs: needSubject, Reg: rPlain, Topic: "wound"},

	// --- what was taken, as a thing that used to sit in a room ---
	{Text: "Whoever came in knew where to find the {res}", Needs: needRes, Reg: rPlain, Topic: "haul"},
	{Text: "Everything else got stepped over on the way to the {res}", Needs: needRes, Reg: rPlain, Topic: "haul"},
	{Text: "The floor where the {res} used to sit has been swept and swept again, by a man who was told to stop an hour ago", Needs: needRes, Reg: rWry, Topic: "building"},
	{Text: "Somebody stood in front of {res_stores} with both arms out, and got put on the floor for it, and has been telling the story since the middle of the morning with the arms still out", Needs: needRes | needMassRes, Reg: rWry, Topic: "people"},
	{Text: "The number being said out loud this evening is {amt_res}, and the number the two who did the counting wrote down is not that number, and the difference is being argued about in the back room by people who counted nothing", Needs: needRes | needAmount, Reg: rWry, Form: fLedger, Topic: "count"},
}

// warRaidGrounded is the part of the old ungated pool whose imagery is a town
// with lanes, dogs and a washing line: true from the Stone Age to the
// Modern Age, and wrong on a station. Moved here after a corpus review.
var warRaidGrounded = []skel{
	{Text: "Whoever pointed them at the {res} lives here, eats here, and was standing in the lane this morning with everybody else", Needs: needRes, Reg: rWry, Topic: "argument"},
	{Text: "A boy has chalked {amt_res} on the wall by the door because he heard it said, and two people have rubbed it out, and it has gone back up both times in a bigger hand", Needs: needRes | needAmount, Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "Two names are chalked up", Reg: rWry, Form: fLedger, Topic: "name"},
	{Text: "Half the animals came home", Reg: rPlain, Topic: "animal"},
	{Text: "The dogs turned up ~ under the wall at first light", Slot: "war_raid_trace", Reg: rPlain, Topic: "animal"},
	{Text: "Halfway up the lane somebody picked up ~", Slot: "war_raid_trace", Reg: rPlain, Topic: "kit"},
	{Text: "The dogs went off an hour before anyone got up", Reg: rPlain, Topic: "animal"},
	{Text: "A woman down the lane counted them going past and got to nine", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The grey dog is dead and the children have been told", Reg: rPlain, Topic: "animal"},
	{Text: "The baker went at them with a shovel and got hurt for it", Reg: rPlain, Topic: "wound"},
	{Text: "A boy who was in the loft the whole time has described the men twice and both descriptions are of the same man", Reg: rPlain, Form: fOverheard, Topic: "people"},
	{Text: "Somebody has been going from house to house with a stick of chalk writing down what each place has lost, and has been let in everywhere", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "Every roof on the water side wants work before the cold comes, and the man who does roofs is one of the two who have not been found", Reg: rWry, Form: fComplaint, Topic: "building"},
	{Text: "The washing was out on the line when it started and it is still out on the line, and the woman it belongs to has not been back in the house", Reg: rPlain, Topic: "town"},
	{Text: "An old man came down to the water at first light with a bucket, filled it, looked at the far bank for a while, and carried the bucket back up", Reg: rPlain, Topic: "ground"},
	{Text: "Two of ours went down in the lane and one of theirs went down with them, and ours have been carried in and washed and laid out, and the third is still lying where he fell", Reg: rPlain, Topic: "casualty"},
	{Text: "Word had gone round by the middle of the morning that they had come through the water side, and by evening that they had come down the lane, and by now that somebody here had walked out and shown them the way in", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "A woman has been at the door since first light asking whether anybody saw her son go past, and she has asked everybody on the row, and two of them have started avoiding the lane so as not to be asked again", Reg: rPlain, Topic: "casualty"},
	{Text: "They made no attempt at all to come in unseen, and walked out the same way, and one of them stopped in the lane to drink", Kinds: []string{"aggressive"}, Reg: rWry, Topic: "people"},
	{Text: "There was a man at the end of the lane who did nothing for the whole of it except stand and watch the rest work, and four separate people have described him, and the four descriptions agree", Kinds: []string{"aggressive"}, Reg: rPlain, Topic: "stranger"},
	{Text: "The lane is still blocked", Reg: rPlain, Form: fComplaint, Topic: "ground"},
	{Text: "Three graves and one is small", Tones: []Tone{Grim}, Reg: rPlain, Form: fLedger, Topic: "casualty"},
	{Text: "The children were kept in the back room all morning while the lane was cleared, and were let out at noon, and went straight to the wall to look", Tones: []Tone{Grim}, Reg: rPlain, Topic: "town"},
	{Text: "There is a version going round in which the whole thing lasted an hour and there were thirty of them, and a shorter version told by the two who were actually out in the lane, and the long one is winning", Tones: []Tone{Wry}, Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "A boy will not give up ~ he found in the mud", Slot: "war_raid_trace", Reg: rWry, Topic: "people"},
	{Text: "They left one of their own face down in the mud", Reg: rPlain, Topic: "casualty"},
	{Text: "Nine barrows of rubble so far", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The digging is being done by whoever can still hold a spade", Tones: []Tone{Grim}, Reg: rWry, Topic: "ground"},
	{Text: "A woman has been sitting with a body since dawn", Tones: []Tone{Grim}, Reg: rPlain, Topic: "casualty"},
	{Text: "The council sat before it was light and by the middle of the day had agreed to meet again", Reg: rJoke, Topic: "authority"},
	{Text: "The lamps were lit all night and are lit now, and nobody has said out loud that they will be lit tomorrow as well", Reg: rWry, Topic: "sleep"},
	{Text: "The well was cleaned out and the water is being carried up from further along, which will go on until somebody can say the well is fit to drink from", Reg: rPlain, Topic: "ground"},
	{Text: "The wounded were brought in and the doors were barred behind them", Tones: []Tone{Grim}, Reg: rPlain, Topic: "wound"},
	{Text: "Third night with the lamps lit", Reg: rPlain, Topic: "sleep"},
	{Text: "Whoever let them in knew which way the bar lifted", Reg: rWry, Topic: "argument"},
	{Text: "Word got to the next place along the water before first light", Reg: rPlain, Form: fOverheard, Topic: "rumour"},
	{Text: "Blood in the wash water", Reg: rPlain, Topic: "wound"},
	{Text: "It smells of burning down by the water and will for days", Reg: rPlain, Topic: "smell"},
	{Text: "Somebody cut a shape into the wall down by the water", Reg: rPlain, Topic: "message"},
	{Text: "Whoever they were, they walked a long way to get here", Kinds: []string{"isolationist"}, Reg: rWry, Topic: "ground"},
	{Text: "The woman who washes the dead has kept back ~ and given it to the boy who sweeps her step", Slot: "war_raid_trace", Reg: rPlain, Topic: "casualty"},
	{Text: "There was no moon at all, which they will have known about, because the two nights before it were bright enough to read by", Reg: rWry, Topic: "weather"},
	{Text: "The man who is keeping watch tonight kept watch last night and the night before, and has been told to go and lie down by everybody who passes the ladder, and is still up on the wall", Tones: []Tone{Grim}, Reg: rWry, Topic: "sleep"},

	// --- re-tagged out of the shared pool by the third late-age review: the
	// kitchen, the back door, the walk home. True up to the Modern Age. ---
	{Text: "The bread was left outside", Reg: rPlain, Topic: "food"},
	{Text: "The man on the wall saw them and stayed where he was", Reg: rWry, Topic: "border"},
	{Text: "The back door was open from the inside", Reg: rPlain, Topic: "building"},
	{Text: "Two houses on the corner were missed entirely", Reg: rPlain, Topic: "people"},
	{Text: "The wall goes up again in the morning, higher this time, and the man who said it was high enough two years ago has been reminded of that by everybody carrying stone", Reg: rWry, Form: fComplaint, Topic: "argument"},
	{Text: "Everything that could be moved indoors was moved indoors before the light came up, and most of it has been carried back out since because the houses are full", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "The wind was up all evening and covered most of the noise, so the first anybody knew of it was a door going in", Reg: rPlain, Topic: "weather"},
	{Text: "The smallest of the children was under the floor the whole time and did not make a sound, and has been carried about by three different people today, and has been asleep for most of it", Reg: rPlain, Topic: "sleep"},
	{Text: "The rebuilding was started before the ash had gone cold, by two men who had been up all night and were not asked to, and there is now a good deal of new timber lying in the wrong place", Reg: rWry, Topic: "building"},
	{Text: "A line of spilled {res} runs down the back stairs and along the passage, and it will be there a while yet", Needs: needRes | needMassRes, Reg: rPlain, Topic: "smell"},
	{Text: "Two of them walked out of the low door carrying {res_haul} between them at no particular speed, and one of them looked back at the house before he went, and four people saw that and none of them describe him the same way", Needs: needRes | needMassRes, Reg: rWry, Topic: "haul"},
	{Text: "Nobody had thought to move the {res} indoors, and the two people who told everyone to move it have not said so again, which everybody has noticed", Needs: needRes, Reg: rWry, Topic: "argument"},
	{Text: "The people who were loudest about the wall being high enough have gone indoors, and the people who were quiet about it are being kind, mostly", Tones: []Tone{Wry}, Reg: rWry, Topic: "argument"},
}

// warRaidLate is the digital and cosmic voice: the field notes have become
// logs, and the logs have started to notice how long everything takes.
var warRaidLate = []skel{
	{Text: "The damage report is still loading", Reg: rWry, Topic: "machine"},
	{Text: "They took the backups first, which means they knew where the backups were", Reg: rPlain, Topic: "machine"},
	{Text: "Four sections are sealed and the people inside them have been told help is coming, which is true in the sense that it is scheduled", Reg: rWry, Topic: "people"},
	{Text: "Every screen in the sector went dark at the same moment and came back a minute later showing the correct time", Reg: rPlain, Topic: "machine"},
	{Text: "The casualty list is being updated live", Reg: rPlain, Form: fLedger, Topic: "casualty"},
	{Text: "Emergency lighting is on in the residential levels and will stay on, and the children have already stopped asking when the proper lights are coming back", Reg: rPlain, Topic: "town"},
	{Text: "Survivors are asked to report to the nearest intact terminal", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "The insurers have described the attack as an act of war, which the people who were in it could have told them for free", Reg: rJoke, Topic: "money"},
	{Text: "Working out what was taken will take longer than the raid did, and the people doing it need more people", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "The security doors opened for them", Reg: rPlain, Topic: "machine"},
	{Text: "The sirens ran all night", Reg: rPlain, Topic: "noise"},
	{Text: "Triage took until morning", Reg: rPlain, Form: fLedger, Topic: "casualty"},
	{Text: "Every camera recorded the wrong corridor", Reg: rWry, Topic: "machine"},
	{Text: "The evacuation drill worked", Reg: rPlain, Topic: "people"},
	{Text: "Level four has no power", Reg: rPlain, Topic: "building"},
	{Text: "All access codes have been changed", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "They used a maintenance code that was supposed to have been retired years ago", Reg: rPlain, Topic: "machine"},
	{Text: "The medics set up in the lobby and the lobby has not been a lobby since", Reg: rWry, Topic: "people"},
	{Text: "The emergency shutters came down on the wrong side of half the residents", Reg: rPlain, Topic: "building"},
	{Text: "The security review has found that every procedure was followed and that the procedures were the problem", Reg: rWry, Topic: "authority"},
	{Text: "The backup generators came on and ran the lights and nothing else", Reg: rPlain, Topic: "machine"},
	{Text: "The injured were sorted with colour-coded tags until the tags ran out", Reg: rPlain, Topic: "wound"},
	{Text: "The insurers want a list of what was taken, itemised, with receipts", Reg: rWry, Form: fComplaint, Topic: "money"},
	{Text: "A security guard held one door for twenty minutes and has asked not to be thanked in public", Reg: rPlain, Topic: "people"},
	{Text: "The alarm system logged every breach and sent the alerts to an address that closed last year", Reg: rPlain, Topic: "machine"},
	{Text: "The official count is lower than the count going round the residential levels", Reg: rWry, Topic: "rumour"},
	{Text: "Whatever they came for was on the lower levels, and they knew which lift to take", Reg: rPlain, Topic: "haul"},
	{Text: "Families from the damaged section are sleeping in the sports hall", Reg: rPlain, Topic: "sleep"},
	{Text: "The facial recognition system identified every one of them as a member of staff", Reg: rWry, Topic: "machine"},
	{Text: "The list of the missing is updated every hour on a screen in the main hall, and the crowd in front of it gets smaller each time a name comes off", Reg: rPlain, Topic: "casualty"},
	{Text: "A minister arrived by evening, was shown the damage, said that lessons would be learned, and was driven back out through the breach the raiders used", Reg: rWry, Topic: "authority"},
	{Text: "The door logs show them entering at one end of the complex and leaving at the other in under twenty minutes, stopping only where the stores were", Reg: rPlain, Topic: "machine"},
	{Text: "Every resident has been sent a form for reporting losses, and the form was written for a burglary, and it has a box for the make of the bicycle", Reg: rWry, Form: fNotice, Topic: "paper"},
	{Text: "A teenager who hid in a ventilation duct the whole time has been interviewed four times and gets the same details right every time, which has made him the most useful witness they have", Reg: rPlain, Topic: "people"},
	{Text: "The breach in the outer wall is being covered with temporary panels, and children have already started drawing on them", Reg: rPlain, Topic: "building"},
	{Text: "The automated defences spent the raid guarding the parts of the complex that were not being attacked, as they had been programmed to, and the engineer who wrote that programming has been at a terminal ever since going through it line by line", Reg: rWry, Topic: "machine"},
	{Text: "The canteen stayed open all night because the woman who runs it would not close it, and by morning she had fed the medics, the guards and the families from the damaged section, and she has gone home now and is not to be called", Reg: rPlain, Topic: "people"},
	{Text: "The lifts are locked down", Reg: rPlain, Topic: "machine"},
	{Text: "Two guards are unaccounted for", Reg: rPlain, Topic: "people"},
	{Text: "Emergency powers were declared and then read", Reg: rWry, Topic: "authority"},
	{Text: "The atrium glass is gone", Reg: rPlain, Topic: "building"},
	{Text: "The power came back in the wrong order and the doors on the east side stayed sealed until noon", Reg: rPlain, Topic: "machine"},
	{Text: "The residents' forum has already named the person who let them in, and has named four different people", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The field hospital in the car park has run out of beds and started using the reclining chairs from the staff lounge", Reg: rPlain, Topic: "wound"},
	{Text: "Repairs have been estimated, and the estimate has been sent back with a request for a smaller one", Reg: rWry, Topic: "money"},
	{Text: "A child was found asleep under a desk in the admin block the next morning, having slept through the whole of it", Reg: rPlain, Topic: "people"},
	{Text: "The night staff want to know why the raid alert reached the day staff first", Reg: rWry, Form: fComplaint, Topic: "people"},
	{Text: "The sprinklers put out the fire in the stores and flooded the level below, where the spare equipment was kept", Reg: rPlain, Topic: "machine"},
	{Text: "The names of the dead were read out over the public address system at noon, and every corridor in the complex stopped to listen", Reg: rPlain, Topic: "casualty"},

	// --- added by the third late-age review: offices, stations, the long view ---
	{Text: "The shutters are still down", Reg: rPlain, Topic: "building"},
	{Text: "Every door log has a gap in it", Reg: rPlain, Topic: "paper"},
	{Text: "Two stairwells are sealed with tape", Reg: rPlain, Topic: "building"},
	{Text: "The lost property office has been turned over to the belongings of the missing", Reg: rPlain, Topic: "casualty"},
	{Text: "Everyone was handed a foil blanket and most of them are still wearing one", Reg: rPlain, Topic: "people"},
	{Text: "The blood bank ran low by midnight and a queue formed in the corridor to give more", Reg: rPlain, Topic: "wound"},
	{Text: "Insurance assessors arrived before the fire crews had finished", Reg: rWry, Topic: "money"},
	{Text: "The one lift that still works has a queue and a guard", Reg: rPlain, Topic: "machine"},
	{Text: "The names of the missing went up on the whiteboard in the canteen, and they have been rubbed out one at a time through the day as people turned up, and the last three have been circled", Reg: rPlain, Topic: "name"},
	{Text: "The rumour on the residential levels is that they had a floor plan", Reg: rPlain, Form: fOverheard, Topic: "rumour"},
	{Text: "Residents are asked not to post photographs of the breach", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Four doors breached, two stores emptied, one guard still in surgery", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The maintenance crew are on their thirtieth hour and have started writing the time on every work order they sign", Reg: rWry, Form: fComplaint, Topic: "people"},
	{Text: "They set off every alarm on purpose on the way out", Kinds: []string{"aggressive"}, Reg: rWry, Topic: "noise"},
	{Text: "They spoke to nobody and their suits carried no markings", Kinds: []string{"isolationist"}, Reg: rPlain, Topic: "stranger"},
	{Text: "The morgue ran out of drawers and borrowed a cold store", Tones: []Tone{Grim}, Reg: rPlain, Topic: "casualty"},
	{Text: "A woman from the damaged section has been going door to door on the upper levels with a photograph of her son on her phone, and every door has opened for her, and nobody behind any of them has seen him", Tones: []Tone{Grim}, Reg: rPlain, Topic: "casualty"},
}

// warRaidAncient — the elders, the herd, the store-pit, the watch-fire.
var warRaidAncient = []skel{
	{Text: "Two huts are down", Reg: rPlain, Topic: "building"},
	{Text: "The store-pit was emptied by hand", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "The drums went up too late", Reg: rWry, Topic: "noise"},
	{Text: "Smoke on the ridge until noon", Reg: rPlain, Topic: "border"},
	{Text: "Nobody goes past the racks", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "They drove off half the herd and killed what would not walk", Reg: rPlain, Topic: "animal"},
	{Text: "A spear was left standing in the ground by the store-pit", Reg: rPlain, Topic: "weapon"},
	{Text: "Somebody let the watch-fire go out well before midnight", Reg: rWry, Form: fComplaint, Topic: "border"},
	{Text: "An arrow came through the roof of the wrong hut", Reg: rWry, Topic: "wound"},
	{Text: "They crossed at the ford, which everyone had said they would not", Reg: rWry, Topic: "ground"},
	{Text: "The hides put out to dry are gone or trampled into the mud", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "The spring below the camp has blood in it and is being left alone", Reg: rPlain, Topic: "ground"},
	{Text: "A boy went after them with a spear and was carried back", Reg: rPlain, Topic: "wound"},
	{Text: "Half the harvest was still in the ground, which saved it", Reg: rWry, Topic: "food"},
	{Text: "The valley people watched them go past and lit nothing at all", Reg: rWry, Topic: "border"},
	{Text: "A hide with a hand-mark on it was tied to the post by the path", Reg: rPlain, Topic: "message"},
	{Text: "The elders have been sitting since before light and have sent for the old woman who remembers the last time, and she is coming up the slope now", Reg: rWry, Topic: "authority"},
	{Text: "Water has been carried up from the ford all morning by whoever could be spared, and the two women doing most of it have not been spelled off once", Reg: rWry, Form: fComplaint, Topic: "people"},
	{Text: "The column that came through was longer than the camp is wide, which two people have now said out loud and nobody has argued with", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "An old woman stood in her doorway with a spear held wrong and they went round her, and she has been standing in the same place most of the day", Reg: rPlain, Topic: "people"},
	{Text: "The tracks go up the valley and stop where the ground turns hard, and two men followed them that far and came back and have not said much since", Reg: rPlain, Topic: "ground"},
	{Text: "Somebody has cut marks into the post beside the store-pit, one for each of ours, and a boy has been sent to ask the elders whether that is allowed", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "Three huts stand open with nothing in them, and the families are in with other families, and the old man who lived alone in the fourth has gone up to sit by the store-pit", Reg: rPlain, Topic: "town"},
	{Text: "They fired the drying racks on the way out and the dogs went in every direction, and half of them were back before dawn and the rest are still somewhere up on the slope", Reg: rPlain, Topic: "animal"},
}

// warRaidFeudal — the gate, the bells, the muster, the steward's arithmetic.
var warRaidFeudal = []skel{
	{Text: "The gate stood open", Reg: rPlain, Topic: "border"},
	{Text: "Supper went cold in the hall", Reg: rWry, Form: fComplaint, Topic: "food"},
	{Text: "Two carts are gone", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "The bells rang late", Reg: rPlain, Topic: "noise"},
	{Text: "Wax still soft on the letter", Reg: rPlain, Topic: "paper"},
	{Text: "Riders were through the village before the bells started", Reg: rWry, Topic: "town"},
	{Text: "The storehouse doors were taken off their hinges and left in the yard", Reg: rPlain, Topic: "building"},
	{Text: "Two pickets were found tied up and unhurt and are being asked about it", Reg: rWry, Topic: "border"},
	{Text: "The steward has counted what is left and started again from the top", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "A clerk was hurt carrying the rolls out and will keep the hand", Reg: rPlain, Topic: "wound"},
	{Text: "Nobody called the militia until the whole of it was over", Reg: rPlain, Form: fComplaint, Topic: "authority"},
	{Text: "The muster is called for dawn and nobody expects much of it", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "An arrow went through the shutter and stuck in the beam", Reg: rPlain, Topic: "weapon"},
	{Text: "Somebody has hung a torn banner over the gate arch", Reg: rWry, Topic: "town"},
	{Text: "The undercroft was found open and most of the way empty", Reg: rPlain, Topic: "building"},
	{Text: "The column came up the road past the ford in good order", Reg: rPlain, Topic: "ground"},
	{Text: "A herald came the next morning with something on parchment that the steward read once, folded, and put inside his coat without saying a word about it", Reg: rWry, Topic: "message"},
	{Text: "The village below lost most of its roofs and blames the gate for it, and two families have already carried their things up the road to relatives", Reg: rWry, Topic: "town"},
	{Text: "The quartermaster will not begin counting until the yard is cleared, and the yard cannot be cleared until the carts are mended, and the carter is at the ford", Reg: rJoke, Form: fComplaint, Topic: "count"},
	{Text: "The gate keeper is dead and his son has been round the whole village asking who had the bar up last, and has been answered politely at every door", Reg: rWry, Topic: "argument"},
	{Text: "Smoke stood over the harvest fields until the rain came in the afternoon, and the men who went out to beat it down came back black to the elbow", Reg: rPlain, Topic: "weather"},
	{Text: "The bells go again at first light, by order of the steward, and the boy who rings them has been given a stool and told to stay by the rope", Reg: rPlain, Form: fNotice, Topic: "noise"},
	{Text: "Two wagons were taken and a third went into the ditch at the bend, and it has been got out and stands in the yard with a broken axle and a great deal of opinion around it", Reg: rJoke, Topic: "machine"},
	{Text: "The horns went up in the wrong order, which had the militia running to the ford while the riders were already inside the yard, and the man who blew them has been stood down and is sitting in the hall", Reg: rWry, Topic: "noise"},
}

// warRaidIndustrial — the siding, the warehouse, the wire, the night shift.
var warRaidIndustrial = []skel{
	{Text: "The telephone line was cut first", Reg: rPlain, Topic: "message"},
	{Text: "Two men have not come in", Reg: rPlain, Form: fLedger, Topic: "casualty"},
	{Text: "A bullet went through the depot clock at twenty past four", Reg: rPlain, Topic: "time"},
	{Text: "Smoke over the yard till dawn", Reg: rPlain, Topic: "weather"},
	{Text: "Freight left lying in the mud", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "The way in was along the rail siding, on foot", Reg: rPlain, Topic: "ground"},
	{Text: "The warehouse doors were cut and the lock left hanging on the chain", Reg: rPlain, Topic: "building"},
	{Text: "A night watchman is in hospital and will keep the eye", Reg: rPlain, Topic: "wound"},
	{Text: "The foreman was on the platform for all of it", Reg: rWry, Topic: "authority"},
	{Text: "Two lorries went out after them and came back with nothing", Reg: rWry, Topic: "machine"},
	{Text: "The payroll safe was opened with something heavy and unhurried", Reg: rWry, Topic: "money"},
	{Text: "The wire to the next station was down the whole night", Reg: rPlain, Topic: "message"},
	{Text: "It will be written up in triplicate and read by nobody", Reg: rJoke, Form: fNotice, Topic: "paper"},
	{Text: "Somebody telegraphed the wrong office and lost forty minutes doing it", Reg: rPlain, Form: fComplaint, Topic: "message"},
	{Text: "Nobody has seen the depot cat since Tuesday night", Reg: rWry, Topic: "animal"},
	{Text: "They knew which shed held the good stock and walked past two others", Reg: rWry, Topic: "haul"},
	{Text: "The office clock stopped at the minute the wall took the blow, and three men have written that time down separately and two of the three agree", Reg: rWry, Topic: "time"},
	{Text: "Men are sleeping in the warehouse tonight with the lamps left burning, and the foreman has said nothing about the cost of the oil", Reg: rWry, Topic: "sleep"},
	{Text: "The valley road was blocked behind them with a felled tree, and the two men sent to clear it took until noon and came back saying the cuts were fresh", Reg: rPlain, Topic: "ground"},
	{Text: "The shareholders will be told about the fence and about the lock, and the meeting at which they are told has been moved back a week", Reg: rJoke, Form: fNotice, Topic: "authority"},
	{Text: "A column of them walked out along the rail line in the dark and nobody thought to stop the goods train, which went through at half past four and saw nothing", Reg: rWry, Topic: "border"},
	{Text: "The annexe roof is holed in two places and the paper under it is ruined, and the two women who keep the files have been drying sheets on the office stove", Reg: rWry, Topic: "paper"},
	{Text: "The telegram went out an hour after they had gone, and the reply came back in the morning telling the yard to secure the fence, which is where the fence used to be", Reg: rWry, Topic: "message"},
	{Text: "The foreman's account and the watchman's account differ on the number of them and on the hour, and both men have been asked to write it out again, and both have written out the same thing again", Reg: rWry, Form: fLedger, Topic: "count"},
}

// warRaidDigital — the feed, the uplink, the drones, the after-action.
var warRaidDigital = []skel{
	{Text: "The feed went dark early", Reg: rPlain, Topic: "machine"},
	{Text: "Power first, then the doors", Reg: rPlain, Topic: "building"},
	{Text: "Two guards are in hospital", Reg: rPlain, Form: fLedger, Topic: "wound"},
	{Text: "The night analyst is still up", Reg: rPlain, Topic: "sleep"},
	{Text: "Nobody wants tomorrow's after-action", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "Two drones went up late and found dust and a cold engine", Reg: rPlain, Topic: "machine"},
	{Text: "The uplink was cut at the pole outside the fence", Reg: rPlain, Topic: "machine"},
	{Text: "An analyst watched the whole of it and could do nothing", Reg: rPlain, Topic: "people"},
	{Text: "The network is back and half the cameras are still down", Reg: rWry, Topic: "machine"},
	{Text: "Somebody put the footage out before the shift had ended", Reg: rWry, Form: fComplaint, Topic: "rumour"},
	{Text: "The channel filled up with people who were nowhere near it", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The service road was used both ways, with every light off", Reg: rPlain, Topic: "ground"},
	{Text: "The last clean image is a shoulder and half a face", Reg: rPlain, Topic: "message"},
	{Text: "The badge reader logged forty entries inside one minute", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "They knew the camera angles better than the guards did", Reg: rWry, Topic: "border"},
	{Text: "Three doors opened on credentials that expired last year", Reg: rWry, Form: fLedger, Topic: "authority"},
	{Text: "The drones are grounded until somebody accounts for the eleven minutes, and the person who has to account for them went home at four and has not been reached", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "A junior analyst called it in four minutes before it started and was told to hold, and has since had to say exactly what she saw to everybody above her", Reg: rWry, Topic: "people"},
	{Text: "Half the floor has watched the same nine seconds of footage since the morning, and the room has agreed on the number of them and on nothing else", Reg: rWry, Form: fOverheard, Topic: "count"},
	{Text: "Somebody's phone is still transmitting from inside a bag somewhere on the far side of the fence, and it has been listened to for six hours and has given up two coughs", Reg: rWry, Topic: "noise"},
	{Text: "The network logged the whole thing as a maintenance window, which is what it was told to log, and the person who scheduled the window is on leave until Thursday", Reg: rWry, Topic: "machine"},
	{Text: "A drone found the vehicles four hours later on a track behind the reservoir, all three of them burnt out and parked neatly, one behind the other", Reg: rWry, Topic: "machine"},
	{Text: "The channel logs stop for nine minutes and start again clean, and the two people who could say why are the two people who were on the desk, and both of them have been sent home to sleep", Reg: rWry, Form: fLedger, Topic: "paper"},
	{Text: "The uplink came back on its own at ten to six and the first thing through it was a maintenance ticket somebody had filed the evening before about a door sensor on the east side, which has now been read by about forty people", Reg: rJoke, Form: fNotice, Topic: "paper"},
	{Text: "The drones were hacked first", Reg: rPlain, Topic: "machine"},
	{Text: "The raid trended before it ended", Reg: rWry, Topic: "rumour"},
	{Text: "Every feed cut at once", Reg: rPlain, Topic: "machine"},
	{Text: "The accounts were drained in seconds and the bank is investigating", Reg: rWry, Topic: "money"},
	{Text: "Their malware is still somewhere on the network and the analysts are hunting it floor by floor", Reg: rPlain, Topic: "machine"},
	{Text: "An analyst watched the whole thing on a feed from home and called it in before the building's own systems noticed", Reg: rWry, Topic: "people"},
	{Text: "The drones that were supposed to defend the compound were turned around and used against it", Reg: rPlain, Topic: "machine"},
	{Text: "The raiders left a message on every screen in the building, and it has a spelling mistake in it, and the analysts are treating the spelling mistake as a clue", Reg: rJoke, Topic: "message"},
	{Text: "The fusion plant tripped to safe mode when the grid dropped, and the city went dark for an hour", Reg: rPlain, Topic: "machine"},
	{Text: "The after-action review has been asked to explain how an attack this large came through a network this closely watched, and would like more time", Reg: rWry, Topic: "paper"},
	{Text: "People filmed it on their phones from the flats across the street, and the footage is better than anything the security system kept", Reg: rPlain, Topic: "people"},
	{Text: "Their implants made them invisible to our cameras and easy to see for a cleaner with a mop, who passed them in the service corridor and took them for contractors", Reg: rWry, Topic: "machine"},

	// --- added by the third late-age review: the building, the city, the network ---
	{Text: "Their drones jammed ours", Reg: rPlain, Topic: "machine"},
	{Text: "The fusion plant was never touched", Reg: rPlain, Topic: "building"},
	{Text: "The lobby feed shows them holding the door for each other", Reg: rWry, Topic: "people"},
	{Text: "The police arrived in time to photograph the tyre marks", Reg: rWry, Topic: "authority"},
	{Text: "An analyst traced their vehicles to a rented garage across the city and found it swept clean", Reg: rPlain, Topic: "map"},
	{Text: "The street outside is closed and covered in broken glass", Reg: rPlain, Topic: "town"},
	{Text: "People in the flats opposite are still posting clips, and the analysts are going through every one of them frame by frame", Reg: rPlain, Topic: "rumour"},
	{Text: "Their malware emptied the stores database and left the canteen menu alone", Reg: rWry, Topic: "machine"},
	{Text: "The hospital across the road took the wounded, then the families, then the press, and wants the network team to do something about the press", Reg: rWry, Topic: "wound"},
	{Text: "The security company that runs the doors has issued a statement saying its system performed as designed, and the analysts have spent the day reading the design documents, which run to four hundred pages and mention the east door once", Reg: rWry, Topic: "authority"},
	{Text: "Two of the night staff slept under their desks", Reg: rPlain, Topic: "sleep"},
	{Text: "The word in the office is that one of the raiders used to work here", Reg: rPlain, Form: fOverheard, Topic: "rumour"},
	{Text: "All staff must re-enrol their fingerprints by Friday", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "Six of the forty cameras were working, and none of the six faced the door they used", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "The cleaners were told to wait for the forensics team and have been waiting since six", Reg: rPlain, Form: fComplaint, Topic: "people"},
	{Text: "They shot out every streetlight on the way in", Kinds: []string{"aggressive"}, Reg: rPlain, Topic: "weapon"},
	{Text: "The office has put flowers on two empty desks", Tones: []Tone{Grim}, Reg: rPlain, Topic: "casualty"},
}

// warRaidCosmic — the hull, the bay, the airlock, the dock crew.
var warRaidCosmic = []skel{
	{Text: "The reactor was left alone", Reg: rWry, Topic: "machine"},
	{Text: "Bay three is open to space", Reg: rPlain, Topic: "building"},
	{Text: "Two of the dock crew died", Reg: rPlain, Topic: "casualty"},
	{Text: "Frost down the port corridor", Reg: rPlain, Topic: "building"},
	{Text: "The airlock seals are cut", Reg: rPlain, Topic: "machine"},
	{Text: "They came in with the transponder dead and nobody challenged it", Reg: rPlain, Topic: "border"},
	{Text: "The cargo bay doors were cut through from the outside", Reg: rPlain, Topic: "building"},
	{Text: "There is a hole in the hull the size of a table", Reg: rPlain, Topic: "machine"},
	{Text: "A bulkhead went inward and the section is still sealed", Reg: rPlain, Topic: "machine"},
	{Text: "Nothing in the bay matches what the manifest has on it", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "They went out through the aft airlock and left it standing open", Reg: rPlain, Topic: "machine"},
	{Text: "The relay logged them and flagged nothing at all", Reg: rWry, Topic: "message"},
	{Text: "Vacuum took the forward compartment before anyone reached the doors", Reg: rPlain, Topic: "casualty"},
	{Text: "A freighter three hours out saw them and stayed quiet", Reg: rWry, Topic: "stranger"},
	{Text: "Orbital watch had them and lost them behind the moon", Reg: rWry, Topic: "border"},
	{Text: "Whoever opened the inner door had a code that worked", Reg: rWry, Topic: "argument"},
	{Text: "The hull plating outside the bay is scored in long parallel lines, which the engineer says was a cutting head and the second engineer says was something else", Reg: rWry, Topic: "argument"},
	{Text: "Air is being rationed until the bay is sealed, and the notice about it went up at four in the morning and has been read by everybody and argued with by most", Reg: rPlain, Form: fNotice, Topic: "food"},
	{Text: "The transponder code they used belonged to a ship broken up eleven years ago, and the woman who found that out has not left the terminal since she found it", Reg: rWry, Topic: "message"},
	{Text: "Two relays went quiet for a minute and came back clean, and the log of that minute is the same log as the minute before it, character for character", Reg: rWry, Topic: "paper"},
	{Text: "A child was found sealed into a storage locker on the lower level, breathing, and would not come out for the first two people who opened it", Reg: rPlain, Topic: "people"},
	{Text: "The manifest was altered before anybody thought to look at it, and the alteration is neat and in the right hand and dated three days before any of this", Reg: rWry, Form: fLedger, Topic: "paper"},
	{Text: "Somebody was in the bay when it went and did not get out, and the dock crew have been welding along the seam all night with the lights up and the talk down", Reg: rPlain, Topic: "casualty"},
	{Text: "Half the ring has no pressure and the other half has questions, and the dock crew have answered the same one about forty times, which is whether the doors were opened from inside or cut from outside", Reg: rWry, Form: fOverheard, Topic: "argument"},
	{Text: "Three sections vented", Reg: rPlain, Topic: "wound"},
	{Text: "The dead will be restored from backup", Reg: rWry, Topic: "people"},
	{Text: "They took the whole cargo ring", Reg: rPlain, Topic: "haul"},
	{Text: "The raid lasted ninety seconds, local", Reg: rWry, Topic: "time"},
	{Text: "The station is running at half spin", Reg: rPlain, Topic: "machine"},
	{Text: "The bodies recovered from outside are being brought in one at a time through the small lock", Reg: rPlain, Topic: "casualty"},
	{Text: "Some of the restored have asked to be told how they died, and some have asked to be spared it", Reg: rWry, Topic: "people"},
	{Text: "Their ships matched our spin before they fired, which means they had studied the station for a long time", Reg: rPlain, Topic: "machine"},
	{Text: "The station administrator asked the ship's mind for an honest estimate of the next raid and has kept the answer to herself", Reg: rWry, Topic: "authority"},
	{Text: "Everyone on the damaged ring has been moved inward, and the inner rings are now so crowded that people are sleeping in the corridors under the emergency lighting", Reg: rPlain, Topic: "people"},
	{Text: "The attack was launched from so far out that the ships that made it are still accelerating away from something they finished weeks ago", Reg: rWry, Topic: "time"},
	{Text: "The raiders took nothing anyone can identify and left a small device bolted to the outer hull, and it has been counting down in a notation no one on the station can read", Reg: rPlain, Topic: "stranger"},

	// --- added by the third late-age review: the ring, the fleet, and what is out past it ---
	{Text: "The ring is still wobbling", Reg: rPlain, Topic: "machine"},
	{Text: "Pressure alarms all night", Reg: rPlain, Topic: "noise"},
	{Text: "They took the seed vault", Reg: rPlain, Topic: "haul"},
	{Text: "Hydroponics lost a whole tier to decompression", Reg: rPlain, Topic: "food"},
	{Text: "The medical bay has been treating cold burns since the second watch", Reg: rPlain, Topic: "wound"},
	{Text: "Their boarding craft left scorch rings on the hull that look like handprints", Reg: rPlain, Topic: "message"},
	{Text: "The station's children have been moved to the core, where the gravity is lighter, and have been told it is a holiday", Reg: rWry, Topic: "town"},
	{Text: "A suit with nobody in it was found walking the outer hull on its last instructions", Reg: rPlain, Topic: "stranger"},
	{Text: "The raiders came in through a cargo lock welded shut thirty years ago, and the weld was cut from the inside, and the security chief wants everyone who has lived aboard longer than that interviewed", Reg: rPlain, Topic: "border"},
	{Text: "The comms mast took a hit and the station has been talking to the fleet by flashing its docking lights", Reg: rWry, Topic: "message"},
	{Text: "The air smells of the backup scrubbers", Reg: rPlain, Topic: "smell"},
	{Text: "The crews say the raiders kept their helmets on the whole time, and some say there was nothing inside them", Reg: rPlain, Form: fOverheard, Topic: "rumour"},
	{Text: "Walking outside the inner ring is suspended until the hull crews say otherwise", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Three decks vented, forty sealed in, all recovered alive", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The hull crews have dug out last month's request for more patrols and pinned it to the mess door", Reg: rWry, Form: fComplaint, Topic: "argument"},
	{Text: "Their ships stayed in view long after the raid, lit up and slow", Kinds: []string{"aggressive"}, Reg: rWry, Topic: "border"},
	{Text: "The dead are being kept in the cold hold until the burial ship comes", Tones: []Tone{Grim}, Reg: rPlain, Topic: "casualty"},
	{Text: "Their boarders wore no insignia and spoke in hand signals only", Kinds: []string{"isolationist"}, Reg: rPlain, Topic: "stranger"},
}
