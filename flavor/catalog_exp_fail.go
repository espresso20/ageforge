package flavor

// ExpeditionFailure — a party was sent out and it went badly.
//
// The log line directly above this one already says "<Name> failed! Partial loot
// recovered." So none of the sentences here say that it failed, and none of them
// weigh the loss against the recovery. They say who did not come home, what broke,
// who is being blamed, what the map got wrong, and what the families were told.
// See catalog.go for the rules, and catalog_exp_success.go for the voice.
//
// Grim is not the same as melodramatic, and this file leans on the flat register
// harder than the others do. The strongest line available here is usually a plain
// account of what somebody did with their hands the following morning.

// --- slot banks --------------------------------------------------------------
//
// Noun phrases, lowercase, one tight category per bank, no era-coded nouns —
// both banks are drawn by ungated sentences, so both must land in every age.

var expFailBanks = map[string][]string{
	// A carried thing that did not come home. Definite noun phrases, so the frame
	// supplies only the verb.
	"exp_fail_kit": {
		"the spare rope", "the good lamp", "the map case", "the spare gloves",
		"the big cook pot", "the second pack", "the drinking water", "the heavy axe",
		"the spare boots", "the medicine box", "the climbing line", "the long knife",
		"the folded bedding", "the last dry clothes", "the signal mirror",
		"the digging tools", "the salt", "the last of the food",
	},
	// An injury, named the way a room names one: by the part of the body everyone
	// is trying not to look at.
	"exp_fail_wound": {
		"the bad hand", "the wrapped ribs", "the split lip", "the swollen knee",
		"the burnt forearm", "the shoulder that will not lift",
		"the eye that has not opened", "the cut under the jaw", "the leg that drags",
		"the bandaged foot", "the arm in the sling", "the broken wrist",
		"the gash across the back", "the ankle nobody will look at",
		"the two missing fingers", "the hand that will not close",
		"the bite on the calf", "the tooth that came out on the way",
	},
}

// --- the sentences -----------------------------------------------------------

// expFailTemplates is the ExpeditionFailure skeleton set: one ungated pool that
// lands in any age, plus one pool per era bucket for the imagery that does not
// travel.
func expFailTemplates() []tmpl {
	out := pool("exp_fail_any", erasAny, expFailAny)
	out = append(out, pool("exp_fail_ancient", erasAncient, expFailAncient)...)
	out = append(out, pool("exp_fail_feudal", erasFeudal, expFailFeudal)...)
	out = append(out, pool("exp_fail_industrial", erasIndustrial, expFailIndustrial)...)
	out = append(out, pool("exp_fail_digital", erasDigital, expFailDigital)...)
	out = append(out, pool("exp_fail_cosmic", erasCosmic, expFailCosmic)...)
	return out
}

// expFailAny fires in every age, so nothing in it may name an era-coded object.
// Counting, arguing, burying, waiting at a door and blaming the quiet one are
// timeless; carts, sidings and airlocks are not.
var expFailAny = []skel{
	// --- short ---
	{Text: "Seven back of nine", Reg: rPlain, Form: fLedger, Topic: "casualty"},
	{Text: "The medicine box is gone", Reg: rPlain, Topic: "kit"},
	{Text: "Nobody has eaten", Reg: rPlain, Topic: "food"},
	{Text: "Both stretchers came in full", Reg: rPlain, Topic: "wound"},
	{Text: "It took eleven days", Reg: rPlain, Topic: "time"},
	{Text: "Kel has stopped talking", Reg: rPlain, Topic: "name"},
	{Text: "The water was bad", Reg: rPlain, Topic: "food"},
	{Text: "Rain for six days", Reg: rPlain, Topic: "weather"},
	{Text: "One boot came home", Reg: rWry, Form: fLedger, Topic: "kit"},
	{Text: "There will be a hearing", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "The dog is missing", Reg: rPlain, Topic: "animal"},
	{Text: "Somebody's mother is at the door", Reg: rPlain, Topic: "town"},
	{Text: "Three packs, one man", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "The wounded came in last", Reg: rPlain, Topic: "wound"},
	{Text: "Serit has said her piece", Reg: rPlain, Topic: "name"},
	{Text: "The light went early", Reg: rPlain, Topic: "weather"},
	{Text: "Everyone is very quiet", Reg: rPlain, Topic: "people"},
	{Text: "Two graves, no marker", Reg: rPlain, Form: fLedger, Topic: "casualty"},
	{Text: "The salt is ruined", Reg: rPlain, Topic: "food"},
	{Text: "Word came ahead of them", Reg: rPlain, Topic: "message"},
	{Text: "Blame is going round", Reg: rWry, Form: fOverheard, Topic: "argument"},
	{Text: "Nothing was found out there", Reg: rPlain, Topic: "map"},

	// --- mid ---
	{Text: "Two names have been scratched off the list and left unreplaced", Reg: rPlain, Form: fLedger, Topic: "casualty"},
	{Text: "Nine went out and the man carried in makes seven", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "A woman has been at the door since first light and will be told soon", Reg: rPlain, Topic: "town"},
	{Text: "The youngest of them has been sitting outside since they got in", Reg: rPlain, Topic: "people"},
	{Text: "Somebody has taken the boots off the dead and it was the right call", Reg: rWry, Topic: "kit"},
	{Text: "They walked away from ~ at the second camp and never went back", Slot: "exp_fail_kit", Reg: rPlain, Topic: "kit"},
	{Text: "The river took ~ on the second morning", Slot: "exp_fail_kit", Reg: rPlain, Topic: "ground"},
	{Text: "Somebody traded ~ for directions that turned out to be wrong", Slot: "exp_fail_kit", Reg: rWry, Topic: "trade"},
	{Text: "Whoever was carrying ~ has not been asked about it yet", Slot: "exp_fail_kit", Reg: rPlain, Topic: "argument"},
	{Text: "Nobody wants to look too long at ~", Slot: "exp_fail_wound", Reg: rPlain, Topic: "wound"},
	{Text: "The dressing on ~ has been changed four times today", Slot: "exp_fail_wound", Reg: rPlain, Topic: "wound"},
	{Text: "There is an argument about who should have carried ~", Slot: "exp_fail_kit", Reg: rWry, Topic: "argument"},
	{Text: "A child keeps asking about ~ and keeps being sent out", Slot: "exp_fail_wound", Reg: rPlain, Topic: "town"},
	{Text: "The map got the water wrong and everything after it followed", Reg: rPlain, Topic: "map"},
	{Text: "The marks on the far side stop halfway and start again nowhere", Reg: rPlain, Topic: "map"},
	{Text: "Whoever drew the last stretch of the map has not come forward", Reg: rWry, Topic: "map"},
	{Text: "The route back was longer than the route out by three days", Reg: rPlain, Topic: "ground"},
	{Text: "The food ran out on the eighth day and the water on the ninth", Reg: rPlain, Form: fLedger, Topic: "food"},
	{Text: "They ate what they had and then they ate what they had left", Reg: rWry, Topic: "food"},
	{Text: "A man in the back room will not have his boots taken off", Reg: rPlain, Topic: "wound"},
	{Text: "The one who wanted to turn back on the fourth day has been very quiet", Reg: rWry, Topic: "argument"},
	{Text: "Two of them have already asked to be sent out again", Reg: rWry, Topic: "people"},
	{Text: "The washing has been done twice and some of it has been burnt", Reg: rPlain, Topic: "smell"},
	{Text: "A boy has been sent round the doors with the names", Reg: rPlain, Form: fNotice, Topic: "message"},
	{Text: "Nobody has been able to say where the third one went", Reg: rPlain, Topic: "casualty"},
	{Text: "The count of what came home was done once and left at that", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "What came back has been put in the small room and locked in", Reg: rPlain, Topic: "haul"},
	{Text: "Somebody is being blamed and it is the quietest of them", Reg: rWry, Topic: "argument"},
	{Text: "The story is being told badly and by the wrong people", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "There is an account going round that has them very much braver", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The families were told before the council was, which is being looked into", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "A door has been shut on the far side of the house all evening", Reg: rPlain, Topic: "building"},
	{Text: "The light out on the ridge was seen from home and taken for something else", Reg: rPlain, Topic: "border"},
	{Text: "Two of them came in without their packs and would not say where", Reg: rPlain, Topic: "haul"},
	{Text: "A name was cut into the rock at the last camp and left there", Reg: rPlain, Topic: "name"},
	{Text: "The ground out there is worse than anybody was told", Reg: rPlain, Form: fComplaint, Topic: "ground"},
	{Text: "Whoever gave them the directions has left the district", Reg: rWry, Topic: "stranger"},
	{Text: "The ones who managed to sleep have said they would rather have stayed awake", Reg: rPlain, Topic: "sleep"},
	{Text: "The wind took the shelter apart at about the worst hour", Reg: rPlain, Topic: "weather"},
	{Text: "They walked the last stretch in silence and have kept it up", Reg: rPlain, Topic: "people"},
	{Text: "A man walked in behind them who has not been asked his business", Reg: rPlain, Topic: "stranger"},
	{Text: "The one who was supposed to be watching has been left alone about it", Reg: rPlain, Topic: "argument"},
	{Text: "Something in the packs smells and nobody is naming it", Reg: rPlain, Topic: "smell"},
	{Text: "Somebody did the count twice and got the same bad number", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "Half the town heard before the other half and it went badly", Reg: rWry, Topic: "rumour"},
	{Text: "An animal got into the food stores on the fifth night", Reg: rPlain, Topic: "animal"},
	{Text: "The best of the water was found the day after they turned back", Reg: rJoke, Topic: "ground"},
	{Text: "Two of them are already telling it as though they were elsewhere", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The last of the light was spent looking for one man", Reg: rPlain, Topic: "casualty"},
	{Text: "Nobody has taken the blankets off the two by the wall", Reg: rPlain, Topic: "casualty"},

	{Text: "Somebody has been sent back out for ~ and has not been told to hurry", Slot: "exp_fail_kit", Reg: rWry, Topic: "kit"},
	{Text: "A price has been put on ~ by somebody who was not there", Slot: "exp_fail_kit", Reg: rWry, Form: fLedger, Topic: "money"},
	{Text: "The woman doing the bandaging has looked at ~ twice and said nothing either time", Slot: "exp_fail_wound", Reg: rPlain, Topic: "people"},
	{Text: "Two people have given different accounts of how they came by ~", Slot: "exp_fail_wound", Reg: rPlain, Form: fOverheard, Topic: "rumour"},
	{Text: "What went into the water with ~ has been listed out on a sheet and pinned up by the door, and people keep stopping to read it on their way past", Slot: "exp_fail_kit", Reg: rPlain, Form: fLedger, Topic: "paper"},

	// --- long ---
	{Text: "The names were read out once in the morning by a man who got two of them wrong, and there has been an argument about that since which is easier to have than the other one", Reg: rWry, Form: fNotice, Topic: "casualty"},
	{Text: "They turned back at the water because it was up, waited two days for it to drop, and it dropped on the third, by which time the food was already a problem", Reg: rPlain, Topic: "ground"},
	{Text: "The man who mapped the crossing has been asked to come and explain it, has agreed to come, and has now sent word that he is unwell", Reg: rWry, Topic: "map"},
	{Text: "Somebody has been round every door in the row already, which was kindly meant and has arrived a good deal faster than the truth of it", Reg: rPlain, Topic: "town"},
	{Text: "Serit came in last, walked past everybody without stopping, and has been sitting with the youngest one's kit in her lap since the middle of the afternoon", Reg: rPlain, Topic: "name"},
	{Text: "There is a boy of about nine standing at the corner watching the door, and he has been sent home twice, and he has come back both times", Reg: rPlain, Topic: "town"},
	{Text: "The pack that came home heaviest turned out to hold one man's whole share of the water, carried nine days after the man himself had stopped needing it", Reg: rPlain, Topic: "haul"},
	{Text: "What is left has been laid out on the floor and is being written down by somebody who keeps having to stop and start the line again", Reg: rPlain, Form: fLedger, Topic: "paper"},
	{Text: "The wounded were brought in on planks cut from the shelter, which is why nobody slept dry for the last four nights, and which nobody has complained about", Reg: rPlain, Topic: "wound"},
	{Text: "Two of them have given accounts of the same afternoon that agree about the weather and about nothing else at all", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The order to turn back was given a full day after the point at which everybody now says it should have been given, and the man who gave it has said as much himself", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "A woman came in from the outer houses with bread and left it on the step without knocking, and it has been sitting there since because nobody wants to be the one to bring it in", Reg: rPlain, Topic: "food"},
	{Text: "The count of what came home is nine items on a list of forty, and the man reading the list aloud stopped at about the twentieth", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "Everything that could be carried was carried, everything that could be dragged was dragged, and a good deal of it was put down again within sight of home", Reg: rPlain, Topic: "haul"},
	{Text: "The one who kept the days went on marking them right to the end, so there is a stick with eleven notches on it in a pack that came home with somebody else", Reg: rPlain, Topic: "time"},
	{Text: "The blame has settled on the youngest for now, on the grounds that he is the youngest, and two people who were there have said so out loud and been ignored", Reg: rWry, Topic: "argument"},
	{Text: "There is a swelling on one of them that was described as nothing on the fourth day, as manageable on the seventh, and is now being looked at by three people at once", Reg: rPlain, Topic: "wound"},
	{Text: "Word went round the outer houses first, then the middle ones, and reached the family last, which is how it always goes and is always described afterwards as unfortunate", Reg: rPlain, Form: fComplaint, Topic: "rumour"},
	{Text: "The last camp was made in a bad place because the good place was an hour further on and nobody had an hour left in them", Reg: rPlain, Topic: "ground"},
	{Text: "A stranger walked with them for two days on the way back, shared what he had, took nothing, and turned off before they were close enough to be seen from home", Reg: rPlain, Topic: "stranger"},
	{Text: "The good lamp went over at the crossing along with the man holding it, and the lamp has been mentioned rather more often this evening than the man", Reg: rWry, Topic: "kit"},
	{Text: "Somebody has written the names on the wall by the door in a hand that gets steadily worse toward the bottom of the list", Reg: rPlain, Topic: "paper"},
	{Text: "The animals that went out with them came back without their loads, arriving in ones and twos over most of a day, and one of them has still not turned up", Reg: rPlain, Topic: "animal"},
	{Text: "There was food waiting for eleven and seven came in, and the woman who cooked it has been standing over the rest of it for an hour without touching it", Reg: rPlain, Topic: "food"},
	{Text: "The route was argued over before they left, in front of witnesses, and everybody who argued for it has spent the evening explaining that they argued for it conditionally", Reg: rWry, Topic: "argument"},
	{Text: "Two of them will not go into the back room where the others are, and have been standing in the passage for long enough that somebody has brought them chairs", Reg: rPlain, Topic: "building"},
	{Text: "The weather turned on the sixth night and stayed turned, and everything after that in every account anybody has given begins with a note about the cold", Reg: rPlain, Topic: "weather"},

	// --- very long ---
	{Text: "The list of what went out runs to two full sides and the list of what came home fits comfortably on the back of the second, and the man who wrote both of them has left them on the table and gone outside", Reg: rPlain, Form: fLedger, Topic: "paper"},
	{Text: "A decision was made on the fifth day that everybody agreed with at the time, that two of them now say they never agreed with, and that the man who made it cannot be asked about, being one of the ones still out there", Reg: rWry, Topic: "argument"},
	{Text: "A woman has been sitting on the step since the afternoon with a bag of her son's clothes in her lap, and she has been told twice that there is nothing yet, and both times she has thanked the person and stayed put", Reg: rPlain, Topic: "town"},
	{Text: "The water they were counting on turned out to be a day further along than the marks said, which was survivable, and the water after that turned out not to be there at all, which was the part that did the damage", Reg: rPlain, Topic: "map"},
	{Text: "Somebody in the outer houses has been telling people all evening that he said from the start it would go this way, and four separate people have now told him to his face that he said the opposite in front of them", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The wounded were laid out along the wall in the order they came in rather than the order anybody would have chosen, so the two worst are at the far end where the draught is, and moving them has been discussed twice and not done", Reg: rPlain, Topic: "wound"},
	{Text: "What came home fits in one small room and has been counted, written down, counted again by somebody who did not trust the first count, and written down a second time on the same sheet in a different hand", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The dog that went out with them came back on its own two days ahead, which everybody understood perfectly well at the time and which nobody said anything about until the party itself came in", Reg: rPlain, Topic: "animal"},
	{Text: "There is a man sitting by the door who has been asked four times whether he wants anything to eat, has said no four times, and is now being brought something anyway by a woman who has stopped asking", Reg: rPlain, Topic: "food"},
	{Text: "The order was read to them before they went, and the paper it was written on came home in a wet pack with the ink gone, and it has been laid out flat to dry on the table with a weight on each corner", Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "Two accounts of the last afternoon have now been given, one by a man who was there for all of it and one by a man who was there for the first hour, and it is the second account that has gone round the town", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The eldest of them has been out at the wall since they got in, doing a job that did not need doing, and three people have gone out to fetch him in and all three have come back alone and said he is busy", Reg: rPlain, Topic: "people"},

	// --- kind: scouting ---
	{Text: "Four days out, turned on the fifth", Kinds: []string{"scouting"}, Reg: rPlain, Form: fLedger, Topic: "time"},
	{Text: "The far side was never reached and will have to be tried again", Kinds: []string{"scouting"}, Reg: rPlain, Topic: "map"},
	{Text: "Whatever is out past the tree line stayed out past it", Kinds: []string{"scouting"}, Reg: rPlain, Topic: "map"},
	{Text: "The sketch stops in the middle of a line", Kinds: []string{"scouting"}, Reg: rPlain, Topic: "paper"},
	{Text: "Two went ahead to look at the crossing and one came back to describe it, and the description is the only thing anybody has to work from", Kinds: []string{"scouting"}, Reg: rPlain, Topic: "map"},
	{Text: "The country past the water was never seen properly because the weather sat on it for four days and then they had no days left to spend", Kinds: []string{"scouting"}, Reg: rPlain, Topic: "weather"},

	// --- kind: military ---
	{Text: "The fighting went on longer than the food", Kinds: []string{"military"}, Reg: rPlain, Topic: "weapon"},
	{Text: "They held the ground for a night and gave it back at dawn", Kinds: []string{"military"}, Reg: rPlain, Topic: "weapon"},
	{Text: "Four wounded and one dead", Kinds: []string{"military"}, Reg: rPlain, Form: fLedger, Topic: "casualty"},
	{Text: "Whoever they met was better at it and knew the ground", Kinds: []string{"military"}, Reg: rPlain, Topic: "border"},
	{Text: "The blades came home dull and nobody is cleaning them tonight", Kinds: []string{"military"}, Reg: rPlain, Topic: "weapon"},
	{Text: "The order was to take the place and hold it, and they took it at about noon and held it until roughly the middle of the night, and the second half of that has been left out of most of the tellings", Kinds: []string{"military"}, Reg: rWry, Form: fOverheard, Topic: "rumour"},

	// --- tone ---
	{Text: "The dead have not been counted aloud yet", Tones: []Tone{Grim}, Reg: rPlain, Topic: "casualty"},
	{Text: "The list on the wall is being read by people who cannot read", Tones: []Tone{Grim}, Reg: rPlain, Topic: "town"},
	{Text: "There are more empty places at the table than full ones", Tones: []Tone{Grim}, Reg: rPlain, Topic: "food"},
	{Text: "The two by the wall have been left where they were put, under a sheet, and everybody coming through the room has slowed down at the same point on the floor and then walked on", Tones: []Tone{Grim}, Reg: rPlain, Topic: "casualty"},
	{Text: "Everyone who stayed home has an excellent theory", Tones: []Tone{Wry}, Reg: rJoke, Form: fOverheard, Topic: "rumour"},
	{Text: "The plan is being called ambitious by the man who wrote it", Tones: []Tone{Wry}, Reg: rJoke, Topic: "argument"},
	{Text: "A great many people have remembered this evening that they had reservations at the time, and between them they have produced more reservations than there were people in the room when it was decided", Tones: []Tone{Wry}, Reg: rJoke, Form: fOverheard, Topic: "argument"},
	{Text: "Somebody has already worked out whose fault it was", Tones: []Tone{Wry}, Reg: rJoke, Topic: "argument"},
	{Text: "The council will meet about it on the day after next, which gives everybody two clear days to arrive at the same account of events", Tones: []Tone{Wry}, Reg: rJoke, Form: fNotice, Topic: "authority"},

	// --- the order title, in a frame a verb-led name survives ---
	{Text: "A page marked {subject} has been turned face down", Needs: needSubject, Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "Whoever wrote the {subject} order is not in the room", Needs: needSubject, Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "Somebody stopped reading at the line that reads {subject}", Needs: needSubject, Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "Nothing filed under {subject} came home the way it left", Needs: needSubject, Reg: rPlain, Form: fNotice, Topic: "haul"},
	{Text: "The board still shows the {subject} order as out", Needs: needSubject, Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "Two men refused the {subject} order before it went out, were noted down as having refused it, and have spent the evening being avoided by people who would rather not talk to them just now", Needs: needSubject, Reg: rWry, Form: fNotice, Topic: "argument"},
	{Text: "The whole file marked {subject} has been put back in the box it came out of, and the box has gone up on the high shelf, and somebody has written the date on the lid", Needs: needSubject, Reg: rPlain, Form: fNotice, Topic: "paper"},

	// --- the resource, as a thing in a room ---
	{Text: "Whatever came back of the {res} fits in one hand", Needs: needRes, Reg: rPlain, Topic: "haul"},
	{Text: "What is left of the {res} has gone into one box", Needs: needRes, Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "There is too little {res} here to argue over", Needs: needRes, Reg: rPlain, Topic: "money"},
	{Text: "Nobody is guarding the {res} tonight", Needs: needRes, Reg: rPlain, Topic: "money"},
	{Text: "Whatever came home of the {res} was set down in the middle of the floor by the first man through the door, and it has not been moved since, and people have been walking round it all evening", Needs: needRes, Reg: rPlain, Topic: "building"},
	{Text: "They set down {res_haul} and nobody moved to help", Needs: needRes | needMassRes, Reg: rPlain, Topic: "haul"},
	{Text: "Nothing has been done with {res_stores} since it was set down", Needs: needRes | needMassRes, Reg: rPlain, Topic: "money"},
	{Text: "All of it together came to {amt_res}", Needs: needRes | needAmount, Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "They walked back with {amt_res} and a man on a plank", Needs: needRes | needAmount, Reg: rPlain, Topic: "wound"},
	{Text: "The figure written up is {amt_res}, and the man who wrote it has been asked twice whether that is the whole of it, and has said yes both times without looking up", Needs: needRes | needAmount, Reg: rPlain, Form: fLedger, Topic: "count"},
}

// expFailAncient — fires, herds, hides, spears, the elders.
var expFailAncient = []skel{
	{Text: "Four spears came home", Reg: rPlain, Form: fLedger, Topic: "weapon"},
	{Text: "The store-pit will not last", Reg: rWry, Topic: "food"},
	{Text: "The hides came home ruined", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "The elders have not spoken", Reg: rWry, Topic: "authority"},
	{Text: "Drums went up at dusk", Reg: rPlain, Topic: "noise"},
	{Text: "The ford ran higher than anyone remembered and took two of them", Reg: rPlain, Topic: "ground"},
	{Text: "An arrow came out of the trees and nobody saw the hand", Reg: rPlain, Topic: "border"},
	{Text: "The herd was lost on the second night and never found again", Reg: rPlain, Topic: "animal"},
	{Text: "A hide was left over the two of them where they fell", Reg: rPlain, Topic: "casualty"},
	{Text: "The watch-fire was not lit on the night it mattered", Reg: rPlain, Form: fComplaint, Topic: "border"},
	{Text: "They came back along the low ground because the ridge was watched", Reg: rPlain, Topic: "ground"},
	{Text: "The elders want the spears counted and nobody has started", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "There is a hut standing empty at the end of the row", Reg: rPlain, Topic: "building"},
	{Text: "The spring on the far side turned out to be dry, which two of them had said it would be before they set out and have not repeated since", Reg: rWry, Topic: "map"},
	{Text: "The old man who read the ridge for them read it wrong, and he knows it, and he has been sitting outside his own doorway since the middle of the afternoon", Reg: rPlain, Topic: "name"},
	{Text: "A boy carried a spear the whole way out and dropped it in the water on the way back, and has been telling everybody about the water rather than the spear", Reg: rWry, Topic: "weapon"},
	{Text: "The drums stopped early and have not started again, and the man who beats them has been asked twice and has said he will when there is something to beat them for", Reg: rPlain, Topic: "noise"},
	{Text: "Two hides came home and one of them is over a man who cannot be moved yet, so the count of hides is being given as one by anybody who is asked", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "Smoke went up on the ridge behind them for two days running and nobody has explained whose it was", Reg: rPlain, Topic: "border"},
	{Text: "The harvest was left standing while everybody watched the low ground, and it is still standing, and somebody will have to say something about it tomorrow", Reg: rPlain, Form: fComplaint, Topic: "town"},
	{Text: "They buried one where he fell because the ground was soft there and it was not soft anywhere else for most of a day in either direction", Reg: rPlain, Topic: "casualty"},
	{Text: "The elders have called everybody to the fire at first light, which has been taken by most people to mean the same thing, and by two people to mean something worse", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "A spear was left planted at the place they turned back from, which somebody did on purpose and nobody has criticised", Reg: rPlain, Topic: "weapon"},
	{Text: "The valley they were told to cross turned out to hold water most of the way across it, which the man who told them about it has now been reminded of by four separate people", Reg: rWry, Topic: "map"},
}

// expFailFeudal — carts, gates, bells, clerks, stewards, musters.
var expFailFeudal = []skel{
	{Text: "One cart came back", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "The bells were not rung", Reg: rWry, Topic: "noise"},
	{Text: "The muster roll is short", Reg: rWry, Form: fLedger, Topic: "casualty"},
	{Text: "Two banners came home wet", Reg: rPlain, Topic: "kit"},
	{Text: "The steward has said nothing", Reg: rWry, Topic: "authority"},
	{Text: "The gate was opened quietly and shut quietly behind them", Reg: rPlain, Topic: "border"},
	{Text: "The ford took a cart and everything that was on it", Reg: rPlain, Topic: "ground"},
	{Text: "A rider went out to meet them and came back at a walk", Reg: rPlain, Topic: "message"},
	{Text: "The clerks have written it all down and put it away", Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "Supper was laid for the whole company and half of it is untouched", Reg: rPlain, Topic: "food"},
	{Text: "A herald was sent ahead with better news than turned out to be true", Reg: rWry, Form: fNotice, Topic: "message"},
	{Text: "The village heard the wagons before it heard anything else", Reg: rPlain, Topic: "town"},
	{Text: "The undercroft has room in it that nobody wanted to have", Reg: rWry, Topic: "building"},
	{Text: "The pickets on the wall saw them at first light and rang nothing, and there has been a good deal of discussion since about whether that was a kindness", Reg: rPlain, Topic: "border"},
	{Text: "The steward has had the muster roll on his table since noon with a candle beside it and has not yet drawn a line through anything", Reg: rPlain, Form: fLedger, Topic: "paper"},
	{Text: "Wax has been spilled across the fair copy, which the clerk noticed an hour later, and the whole of it is being written out again by lamplight", Reg: rPlain, Form: fComplaint, Topic: "paper"},
	{Text: "A cart was left at the crossing with a wheel off and its load in the water, and two men have gone out with rope to see what can be had back", Reg: rPlain, Topic: "haul"},
	{Text: "The village priest has been at three doors already this evening and has two more to do before he can go home", Reg: rPlain, Topic: "religion"},
	{Text: "The militia turned out at the gate to bring them in and stood there a long while, and by the time the last of them was through the line had thinned out to about six", Reg: rPlain, Topic: "town"},
	{Text: "Somebody rang the bells at the wrong hour out of habit, realised halfway through, and stopped, which was worse than either finishing or not starting", Reg: rWry, Topic: "noise"},
	{Text: "The road past the ford is cut again and the long way round adds two days, which is being explained to the steward by a man who has explained it before", Reg: rWry, Form: fComplaint, Topic: "ground"},
	{Text: "A banner came home in a state that means it will have to be replaced rather than mended, and the woman who made the first one has offered to make the second", Reg: rPlain, Topic: "kit"},
	{Text: "The quartermaster has the count and will not give it out until the steward has had it, which is correct and is making him unpopular in the yard", Reg: rWry, Form: fNotice, Topic: "count"},
	{Text: "Two riders have been sent back out along the road with lamps to look for anybody still walking, and they were told to turn round at midnight and have not", Reg: rPlain, Topic: "casualty"},
}

// expFailIndustrial — depots, sidings, foremen, telegrams, triplicate.
var expFailIndustrial = []skel{
	{Text: "The telegram was short", Reg: rWry, Form: fNotice, Topic: "message"},
	{Text: "One crate came home", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "The payroll will be adjusted", Reg: rWry, Form: fNotice, Topic: "money"},
	{Text: "Nobody signed the freight book", Reg: rWry, Form: fLedger, Topic: "paper"},
	{Text: "The depot was kept open", Reg: rWry, Topic: "building"},
	{Text: "The foreman signed for what came back and went outside", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Freight came in on the late train with two men riding on it", Reg: rPlain, Topic: "haul"},
	{Text: "The warehouse book has a page torn out of the middle", Reg: rPlain, Form: fLedger, Topic: "paper"},
	{Text: "A wire went out at four and was answered at eleven", Reg: rPlain, Form: fNotice, Topic: "message"},
	{Text: "The office telephone has been ringing since the news reached town", Reg: rPlain, Topic: "message"},
	{Text: "It has been written up in triplicate and filed in a hurry", Reg: rJoke, Form: fNotice, Topic: "paper"},
	{Text: "The lorries came in one behind the other with their lamps still lit", Reg: rPlain, Topic: "machine"},
	{Text: "The night shift stayed on without being asked", Reg: rPlain, Topic: "people"},
	{Text: "The company will hear a version of this that has been through three offices, and by the time it gets there the weather in it will be considerably worse", Reg: rWry, Form: fNotice, Topic: "rumour"},
	{Text: "There is a wire on the desk that came in this afternoon and has been left where it landed, and two people have picked it up and put it down again unopened", Reg: rPlain, Form: fNotice, Topic: "message"},
	{Text: "The platform was cleared for the eleven o'clock and the eleven o'clock came in with the lamps down, which everybody standing on it understood before the doors opened", Reg: rPlain, Topic: "message"},
	{Text: "The foreman's tally and the office tally agree for once, which has been remarked on by two men who have not thought it through", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "A crate that came back has been opened, gone through, closed again and put in the annexe, and nobody has been able to say what was supposed to be in it", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "The men who came off the train were told to go home and paid for the full week, and about half of them went to the yard instead and have been there since", Reg: rPlain, Topic: "people"},
	{Text: "The siding has been kept clear all evening for a second train that was mentioned once in a garbled wire and that nobody has since been able to confirm", Reg: rPlain, Form: fComplaint, Topic: "machine"},
	{Text: "A photograph was taken at the platform when they went out, three weeks ago, by a man from the town paper, and somebody has taken it down off the office wall", Reg: rPlain, Topic: "town"},
	{Text: "The depot clock stopped at some point in the afternoon and nobody has wound it, and the whole building has been working off a pocket watch on the counter", Reg: rPlain, Topic: "time"},
	{Text: "The shareholders' letter is due on Friday and the man who writes it has been sitting with a blank sheet since six, and has twice asked what the weather was like out there", Reg: rWry, Form: fNotice, Topic: "paper"},
	{Text: "Two of them came in on stretchers and one walked off his at the platform, and the doctor has written that down in a way that makes clear what he thinks of it", Reg: rWry, Topic: "wound"},
}

// expFailDigital — uplinks, drones, feeds, analysts, channels.
var expFailDigital = []skel{
	{Text: "The uplink dropped twice", Reg: rPlain, Topic: "machine"},
	{Text: "One drone came home", Reg: rPlain, Form: fLedger, Topic: "machine"},
	{Text: "The channel has gone quiet", Reg: rWry, Topic: "noise"},
	{Text: "Nobody has left the building", Reg: rWry, Topic: "sleep"},
	{Text: "The feed ran out at dusk", Reg: rPlain, Topic: "machine"},
	{Text: "The after-action report has been started four times", Reg: rJoke, Form: fNotice, Topic: "paper"},
	{Text: "An analyst has watched the same nine seconds all afternoon", Reg: rPlain, Topic: "people"},
	{Text: "The network held perfectly and carried every bit of it home", Reg: rWry, Topic: "machine"},
	{Text: "Somebody closed the channel and the floor went very quiet", Reg: rPlain, Topic: "noise"},
	{Text: "A badge has been left on a desk that its owner will not want back", Reg: rPlain, Topic: "casualty"},
	{Text: "The uplink logs have been pulled and sent upstairs", Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "A drone came back on one rotor and landed on the roof by itself", Reg: rPlain, Topic: "machine"},
	{Text: "The feed cut at the point everybody has since wanted to see", Reg: rPlain, Topic: "machine"},
	{Text: "Two analysts have gone through the last hour of it frame by frame and have come out of the room with the same account, which is the first thing all evening that anybody has been able to rely on", Reg: rPlain, Topic: "paper"},
	{Text: "The channel logs from the second team stop mid-sentence, and the sentence they stop in the middle of has been read aloud once and has not been read aloud again", Reg: rPlain, Form: fOverheard, Topic: "message"},
	{Text: "A junior analyst flagged the ground conditions four days ago in a note that went to three people, and one of the three has spent this evening finding out which of the other two read it", Reg: rWry, Form: fNotice, Topic: "argument"},
	{Text: "The after-action briefing has been put back to the morning because the man who has to give it has been on the floor since before it happened and cannot yet be got to stop working", Reg: rPlain, Form: fNotice, Topic: "sleep"},
	{Text: "There is a drone still out there somewhere on its last power, sending the same position every forty seconds, and the room has stopped pretending it is going to move", Reg: rPlain, Topic: "machine"},
	{Text: "The coffee has been going since about four and there is a cup on every desk on the floor, and most of them are cold, and none of them have been thrown out", Reg: rPlain, Topic: "food"},
	{Text: "Someone patched a drone with tape three weeks ago and it is one of the two that came home, and the engineer who did it has not said anything about it and is not going to", Reg: rWry, Topic: "machine"},
	{Text: "The feed from the lead element is being replayed on the wall screen for the fourth time, and the room has quietly reorganised itself so that fewer people are facing it", Reg: rPlain, Topic: "people"},
	{Text: "A note has gone round the floor asking people to keep the channel clear, which nobody needed telling, and which everybody has read twice", Reg: rPlain, Form: fNotice, Topic: "message"},
	{Text: "The network flagged the whole thing as routine at the time and is still flagging it as routine now, and somebody has been given the job of going in and changing that", Reg: rWry, Form: fNotice, Topic: "paper"},
	{Text: "Two people who were not on the operation have been on the floor since the middle of the evening asking what happened, and they have been answered politely, and they are still asking", Reg: rWry, Form: fOverheard, Topic: "rumour"},
}

// expFailCosmic — hulls, bays, relays, transponders, airlocks.
var expFailCosmic = []skel{
	{Text: "The hull is opened along one seam", Reg: rPlain, Topic: "machine"},
	{Text: "Bay two came home empty", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "The relay heard nothing", Reg: rWry, Topic: "message"},
	{Text: "Three suits came back", Reg: rPlain, Form: fLedger, Topic: "kit"},
	{Text: "The airlock has been cycling", Reg: rWry, Topic: "machine"},
	{Text: "They came in on emergency air with the reactor down", Reg: rPlain, Topic: "machine"},
	{Text: "The transponder was answering right up to the last hour", Reg: rWry, Topic: "message"},
	{Text: "A repair patch let go on the way in and took a panel with it", Reg: rPlain, Topic: "machine"},
	{Text: "The dock crew came out and then stopped where they were", Reg: rPlain, Topic: "town"},
	{Text: "The manifest has more on it than the bay does", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "Vacuum got into the forward compartment on the second day", Reg: rPlain, Topic: "casualty"},
	{Text: "Somebody has left a name scratched into a bulkhead plate", Reg: rPlain, Topic: "name"},
	{Text: "Orbital watch logged the approach and said nothing over the relay", Reg: rPlain, Form: fNotice, Topic: "border"},
	{Text: "The relay picked up a carrier on the third day with nothing on it, held it for two hours, lost it, and has been listening on the same frequency since without being told to", Reg: rPlain, Topic: "message"},
	{Text: "The bay doors were opened before the ship was properly down because everybody on the deck could already see how many suits were standing up inside", Reg: rPlain, Topic: "machine"},
	{Text: "A freighter crew answered the distress and got there eleven hours after it went out, which was as fast as anybody could have done it and is being said out loud a good deal this evening", Reg: rPlain, Topic: "stranger"},
	{Text: "The reactor was run past every limit on the way home and it held, and the engineer who ran it that way has been sitting in the corridor outside the bay since they docked", Reg: rPlain, Topic: "machine"},
	{Text: "There is a boot in the cargo net that has been in the cargo net since the unload, and the dock crew have worked round it all evening", Reg: rPlain, Topic: "kit"},
	{Text: "The hull sensors have been reporting the same fault every ninety seconds since they came in, and somebody has muted the panel rather than answer it", Reg: rWry, Topic: "machine"},
	{Text: "The manifest was signed at the far end by a man whose name is on the other list as well, and the two lists have been put side by side on the counter and left there", Reg: rPlain, Form: fLedger, Topic: "paper"},
	{Text: "Two of the crew have not come out of the suits yet and have been sitting in them in the bay for most of an hour, and the medical officer has decided to let that run", Reg: rPlain, Topic: "people"},
	{Text: "Everything in the forward bay is scorched down one side and the smell has got into the corridor, and the door at the end of it has been shut and taped", Reg: rPlain, Topic: "smell"},
	{Text: "The dock lights were left burning for the whole of the last watch and were still burning when they came in, which the officer who ordered it has been thanked for by three people", Reg: rPlain, Topic: "town"},
	{Text: "A crate came home sealed, with a number on it that matches nothing on either bay list, and it has been put in the corner of the office and nobody has opened it", Reg: rPlain, Form: fLedger, Topic: "haul"},
}
