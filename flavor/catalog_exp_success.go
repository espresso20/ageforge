package flavor

// ExpeditionSuccess — a party was sent out and it worked.
//
// The log line directly above this one already says "<Name> succeeded! Gained
// loot." So none of the sentences here say that it worked. They say what came
// back, who did not, what broke, what the map says now, what the town did about
// it, and who is already arguing over the share. See catalog.go for the rules.
//
// This file is also the VOICE REFERENCE for the other four Moments. What it is
// trying to be is a quartermaster's book rather than a book of aphorisms:
// circumstantial, often mundane, occasionally digressive, and funny by accident
// about one line in seven. The lengths swing on purpose — a four-word line next
// to a thirty-eight-word one — because uniform rhythm is what made the two
// previous versions of this catalog read as machine-written.

// --- slot banks --------------------------------------------------------------
//
// Noun phrases, lowercase, one tight category per bank. A slot swaps the detail
// in a sentence; it never changes what the sentence is about. Bank fragments
// inherit the eras of every skeleton that references them, so anything drawn on
// from the ungated pool has to be as true of a hunting party as of a boarding
// crew.

var expSuccessBanks = map[string][]string{
	// What they carried in.
	"exp_success_haul": {
		"the heavy cases", "the bundles", "the crates", "the heavy bags", "the sealed boxes",
		"the packs", "the whole load", "the last two bundles", "the roped bundles",
		"the smaller cases", "the long crates", "the unopened bags",
		"the padded boxes", "the split bags", "the folded canvas", "the heavy end of it",
		"the burnt crates", "the last of the packs",
	},
	// A small physical thing that goes wrong or gets lost.
	"exp_success_kit": {
		"lamp", "rope", "knife", "good coat", "water skin", "spare boot",
		"cooking pot", "tin cup", "small hammer", "water bottle", "belt knife",
		"good blanket", "long rope", "brass whistle", "walking staff",
		"canvas satchel", "spare pin", "sharpening stone",
	},
}

// --- the sentences -----------------------------------------------------------

// expSuccessTemplates is the ExpeditionSuccess skeleton set: one ungated pool
// that lands in any age, plus one pool per era bucket for the imagery that does
// not travel.
func expSuccessTemplates() []tmpl {
	out := pool("exp_success_any", erasAny, expSuccessAny)
	out = append(out, pool("exp_success_grounded", erasGrounded, expSuccessGrounded)...)
	out = append(out, pool("exp_success_late", erasLate, expSuccessLate)...)
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
	// --- short. A fact, and then the sentence stops. ---
	{Text: "Two are limping", Reg: rPlain, Topic: "wound"},
	{Text: "Nine in, eleven counted", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "Serit slept eleven hours", Reg: rPlain, Topic: "name"},
	{Text: "Everyone wants a share", Reg: rPlain, Topic: "argument"},
	{Text: "It rained the whole way", Reg: rPlain, Topic: "weather"},
	{Text: "The cook is singing badly", Reg: rPlain, Topic: "people"},
	{Text: "Word travelled ahead of them", Reg: rPlain, Topic: "rumour"},
	{Text: "The rope held", Reg: rPlain, Topic: "kit"},
	{Text: "A stranger walked in behind them", Reg: rPlain, Topic: "stranger"},
	{Text: "Three teeth between four men", Reg: rWry, Form: fLedger, Topic: "wound"},
	{Text: "The bread had gone green", Reg: rPlain, Topic: "food"},
	{Text: "Two names are gone", Reg: rPlain, Topic: "casualty"},
	{Text: "Salt got into the water", Reg: rPlain, Topic: "food"},
	{Text: "Half of them are asleep sitting up", Reg: rPlain, Topic: "sleep"},
	{Text: "The cook was told nothing", Reg: rPlain, Form: fComplaint, Topic: "food"},
	{Text: "A child is charging admission", Reg: rJoke, Topic: "town"},
	{Text: "Kel is not speaking to Serit", Reg: rPlain, Topic: "argument"},
	{Text: "Boots off, then arguing", Reg: rWry, Topic: "people"},
	{Text: "Somebody has bled on this", Reg: rPlain, Topic: "wound"},

	// --- mid. The working length: one thing, with enough detail to place it. ---
	{Text: "The youngest one has not stopped talking since the moment he got in", Reg: rWry, Topic: "people"},
	{Text: "The whole lot came in at dusk without a word to anyone they passed, ate what was left of the evening meal cold, and were asleep before the light had properly gone", Reg: rPlain, Topic: "food"},
	{Text: "One of them came back barefoot and will not discuss it", Reg: rWry, Topic: "people"},
	{Text: "There is a prisoner in the back room who will not give a name", Reg: rPlain, Topic: "stranger"},
	{Text: "They came back a day early, which has upset the kitchen enormously", Reg: rWry, Form: fComplaint, Topic: "food"},
	{Text: "The oldest of them sat down in the doorway and would not get up", Reg: rPlain, Topic: "people"},
	{Text: "Two of them have new scars and two versions of how they got them", Reg: rWry, Topic: "wound"},
	{Text: "The list of names is shorter than it was when they set out", Reg: rPlain, Form: fLedger, Topic: "casualty"},
	{Text: "Two were buried on the way out, at a bend in the water that nobody thought to mark, and the man who dug for them cannot now say where it was", Reg: rPlain, Topic: "casualty"},
	{Text: "One of them has decided to stay out there, and they let him", Reg: rPlain, Topic: "casualty"},
	{Text: "The load came in whole, and so did ~, which nobody had written down on the way out", Slot: "exp_success_haul", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "Somebody counted ~ twice and got two answers", Slot: "exp_success_haul", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "There is blood on ~ and nobody is saying whose", Slot: "exp_success_haul", Reg: rPlain, Topic: "haul"},
	{Text: "The same man has claimed ~ three separate times today", Slot: "exp_success_haul", Reg: rWry, Topic: "argument"},
	{Text: "Water got into ~ on the last night out", Slot: "exp_success_haul", Reg: rPlain, Topic: "haul"},
	{Text: "Two men are sitting on ~ and will not move off it", Slot: "exp_success_haul", Reg: rWry, Topic: "argument"},
	{Text: "They lost the second ~ on the first night and managed without", Slot: "exp_success_kit", Reg: rPlain, Topic: "kit"},
	{Text: "Directions were bought off a local for a ~", Slot: "exp_success_kit", Reg: rPlain, Topic: "trade"},
	{Text: "A child has been given the broken ~ to play with", Slot: "exp_success_kit", Reg: rPlain, Topic: "town"},
	{Text: "Somebody left a ~ behind on purpose, or so the story goes", Slot: "exp_success_kit", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The only thing that came home whole is a ~", Slot: "exp_success_kit", Reg: rWry, Topic: "kit"},
	{Text: "Everything is spread across the floor and being sorted by size", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "The map is longer than it was on the morning they set out, and one whole corner of it is now taken up with a coastline that two of them insist is there and the third says he never saw", Reg: rPlain, Topic: "map"},
	{Text: "They found water where the old marks said there was none", Reg: rPlain, Topic: "map"},
	{Text: "A marker was left at the crossing for whoever comes next, cut in at about the height of a man, which several people have already pointed out is the wrong height for when the water is up", Reg: rPlain, Topic: "map"},
	{Text: "The far side of the hills has a name now, and it is rude", Reg: rJoke, Topic: "map"},
	{Text: "A child has already made up a song about the whole business", Reg: rWry, Topic: "town"},
	{Text: "The children were sent to bed and listened from the stairs anyway", Reg: rPlain, Topic: "town"},
	{Text: "There is a name being repeated tonight that nobody knew this morning", Reg: rPlain, Topic: "name"},
	{Text: "The old woman who told them where to look would like paying", Reg: rPlain, Topic: "money"},
	{Text: "Something followed them for two days and then it stopped", Reg: rPlain, Topic: "ground"},
	{Text: "The rain started an hour after the last of them was indoors", Reg: rPlain, Topic: "weather"},
	{Text: "The crossing was frozen and they walked over without looking down", Reg: rPlain, Topic: "ground"},
	{Text: "The water was worse than the fighting, according to all four of them", Reg: rPlain, Form: fOverheard, Topic: "ground"},
	{Text: "The route home was twice the length of the route out and no explanation has been offered by anybody who walked it", Reg: rPlain, Topic: "ground"},
	{Text: "The story got better every time it was told on the walk back", Reg: rWry, Topic: "rumour"},
	{Text: "A woman named Serit did most of the work and none of the talking", Reg: rWry, Topic: "name"},
	{Text: "The dried meat ran out two days short of home and they walked the rest of it hungry, which is being mentioned rather more often than the distance covered", Reg: rPlain, Topic: "food"},
	{Text: "They walked the last mile home without singing, and it was noticed", Reg: rPlain, Topic: "sleep"},
	{Text: "Kel wants it written down that he said the hills were passable", Reg: rWry, Form: fComplaint, Topic: "paper"},
	{Text: "One of them will not go near water now and will not say why", Reg: rPlain, Topic: "people"},
	{Text: "The wound on the tall one's arm is being described as nothing much", Reg: rWry, Topic: "wound"},
	{Text: "Two of them came back wearing boots that belong to somebody else", Reg: rWry, Topic: "people"},
	{Text: "The whole lot of them went straight to sleep in the wrong beds", Reg: rPlain, Topic: "sleep"},
	{Text: "Somebody has promised half of it away before it was even counted", Reg: rWry, Topic: "money"},
	{Text: "Three of them are arguing about who saw the place first", Reg: rPlain, Topic: "argument"},
	{Text: "There is a bruise on the tall one shaped exactly like a boot", Reg: rJoke, Topic: "wound"},
	{Text: "One lost a tooth out there and considers it a fair swap", Reg: rWry, Topic: "wound"},
	{Text: "Every door in the place has been left standing open since noon", Reg: rPlain, Topic: "building"},
	{Text: "Anyone who wants the story told again will have to pay for it", Reg: rJoke, Form: fNotice, Topic: "rumour"},
	{Text: "Nobody has slept and nobody looks likely to start", Reg: rPlain, Topic: "sleep"},
	{Text: "Something came back with them that no one can identify", Reg: rJoke, Topic: "animal"},

	// --- long. Circumstantial, subordinated, the way a chronicle actually runs. ---
	{Text: "Serit walked the last of it on a bad ankle and would not be carried, so the whole lot arrived after dark and found the kitchen shut", Reg: rPlain, Topic: "name"},
	{Text: "They were three days late because of the water and then two days early once it dropped, so nobody was waiting when they finally came in", Reg: rPlain, Topic: "time"},
	{Text: "There was an argument about the share before anyone had washed, another one after the evening meal, and a third this morning involving two people who did not go", Reg: rWry, Topic: "argument"},
	{Text: "Somebody put a mark on the doorframe for every day they were out, then stopped on the ninth, and has not gone back to finish it", Reg: rPlain, Topic: "time"},
	{Text: "The heavy stuff went into the back room, the light stuff into the front, and one bundle that nobody will claim is sitting in the passage", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "Two of them have been sat in the corner since they got in, eating without speaking, and the third keeps getting up to check the door", Reg: rPlain, Topic: "people"},
	{Text: "The old man who gave them the crossing wants a share, and he is not wrong to ask, and there is nobody who wants to be the one to tell him no", Reg: rPlain, Topic: "money"},
	{Text: "A boy walked out three miles to meet them on the way in and walked the last of it holding somebody's pack for him", Reg: rPlain, Topic: "town"},
	{Text: "There is a list on the wall of what came back and a shorter list beside it of what was promised, and the two are being compared by strangers all afternoon", Reg: rWry, Form: fLedger, Topic: "paper"},
	{Text: "It took four of them to get the heavy end through the door and it will take six to get it out again when somebody decides where it goes", Reg: rWry, Topic: "haul"},
	{Text: "They have been asked to describe the country beyond the hills and so far have offered a hand gesture, a long silence, and the word steep", Reg: rWry, Form: fOverheard, Topic: "map"},
	{Text: "Nobody wrote down the day they crossed the water, so the count of days out is being reconstructed from what people remember eating", Reg: rWry, Form: fLedger, Topic: "paper"},
	{Text: "Two of them have gone straight back out to fetch the load that was cached at the halfway point, on the grounds that somebody else will take it if they wait", Reg: rPlain, Topic: "haul"},
	{Text: "The tall one has told everybody that she is fine, has said it while sitting down, and has now been carried indoors by two people she is still telling", Reg: rWry, Topic: "wound"},
	{Text: "An old man has been at the door since first light asking after a name that is on neither list, and he has had both lists read out to him slowly", Reg: rPlain, Topic: "casualty"},
	{Text: "The rope, the good lamp and both spare knives went over the side at the crossing, and the man who let go of them has been reminded of it at every meal since", Reg: rWry, Topic: "kit"},
	{Text: "They found a place where the water comes out of the rock cold enough to hurt, and drank until they were sick, and have talked about very little else since", Reg: rPlain, Topic: "ground"},
	{Text: "Everything that came in wet is hanging up, everything that came in broken is in a pile by the door, and everything that came in whole has already been taken by somebody", Reg: rWry, Form: fLedger, Topic: "haul"},
	{Text: "A woman none of them recognised walked along with them for half a day, said nothing anybody could understand, and turned off at the water without looking back", Reg: rPlain, Topic: "stranger"},
	{Text: "Kel has explained the route to four separate people and drawn it out each time on the back of the same sheet, and no two of the drawings agree", Reg: rWry, Topic: "map"},
	{Text: "The last four miles were done in the dark because nobody wanted to make camp that close to home, which everyone now agrees was a poor decision", Reg: rPlain, Topic: "ground"},
	{Text: "Somebody has been going round asking what each thing is worth, and getting a different number from every person, and writing all of them down", Reg: rWry, Form: fLedger, Topic: "money"},

	// --- very long. Digressive, specific, a paragraph that forgot to stop. ---
	{Text: "The count was done at the door by two people who did not agree, then again in the back room by one who did not care, and the number written up came from a boy nobody had asked", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "Serit has been asked what the country beyond the water is like and has now said, at some length and to four different people, that it is much the same as here except that the birds are wrong", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The evening meal was late because the kitchen had planned for twelve and cooked for twelve, and fifteen came in, two of whom nobody could name, and the whole thing was stretched with water and a great deal of bread", Reg: rPlain, Topic: "food"},
	{Text: "Word had gone round three days ago that they were all dead, on the authority of a man who heard it from a man who was not there, and that man has found a reason to be out of town since noon", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The heaviest of what came in has been left in the passage where it was dropped, because moving it needs four people and every one of the four has found something else that badly needs doing", Reg: rWry, Topic: "haul"},
	{Text: "They walked past the place they meant to stop at, then past the next one, and made the whole last stretch in one go, and the reason given for this is that the first one had a smell about it", Reg: rPlain, Topic: "ground"},
	{Text: "A woman has been waiting by the door since the middle of the afternoon with a clean shirt for one of them, and he walked straight past her to the food", Reg: rPlain, Topic: "town"},
	{Text: "Everything came back through the same door within the space of an hour, in no order at all, and the sorting of it has been handed to a boy of about ten who has taken to the work with more seriousness than anyone expected", Reg: rWry, Form: fLedger, Topic: "haul"},

	// --- kind: scouting ---
	{Text: "Three days out, two back", Kinds: []string{"scouting"}, Reg: rPlain, Form: fLedger, Topic: "time"},
	{Text: "They walked a good deal further than they were told to", Kinds: []string{"scouting"}, Reg: rWry, Topic: "map"},
	{Text: "The new marks on the map are in a shaky hand", Kinds: []string{"scouting"}, Reg: rPlain, Topic: "map"},
	{Text: "They counted the lights on the far side and stopped at forty", Kinds: []string{"scouting"}, Reg: rPlain, Topic: "border"},
	{Text: "Two of them went up the ridge to see what was on the other side, came back down without saying anything, and went up again the next morning with the third", Kinds: []string{"scouting"}, Reg: rPlain, Topic: "map"},
	{Text: "The sketch of the far bank is good enough to work from, which is a considerable improvement on the last one, which was a circle", Kinds: []string{"scouting"}, Reg: rWry, Topic: "map"},
	{Text: "Nobody has been that far and come home still talking", Kinds: []string{"scouting"}, Reg: rWry, Topic: "map"},

	// --- kind: military ---
	{Text: "The fighting was short and the walk home took nine days", Kinds: []string{"military"}, Reg: rPlain, Topic: "weapon"},
	{Text: "They took the place at dawn and were gone before noon", Kinds: []string{"military"}, Reg: rPlain, Topic: "weapon"},
	{Text: "One of them will not put the blade down yet", Kinds: []string{"military"}, Reg: rPlain, Topic: "weapon"},
	{Text: "The wounded came in first", Kinds: []string{"military"}, Reg: rPlain, Topic: "wound"},
	{Text: "Two dead, nine home", Kinds: []string{"military"}, Reg: rPlain, Form: fLedger, Topic: "casualty"},
	{Text: "There was a great deal of shouting on the way in and none of it was orders", Kinds: []string{"military"}, Reg: rWry, Topic: "noise"},
	{Text: "The blades have all been cleaned and put away and the man who did it has been sitting looking at the wall ever since", Kinds: []string{"military"}, Reg: rPlain, Topic: "weapon"},

	// --- tone ---
	{Text: "The drink is already gone", Tones: []Tone{Triumphant}, Reg: rWry, Topic: "food"},
	{Text: "Somebody has decided this deserves a statue", Tones: []Tone{Triumphant}, Reg: rJoke, Topic: "town"},
	{Text: "There is dancing and a great deal of it is very bad", Tones: []Tone{Triumphant}, Reg: rJoke, Topic: "town"},
	{Text: "Two people who have never met are toasting each other by the fire, and one of them was not here this morning and cannot say who invited him", Tones: []Tone{Triumphant}, Reg: rWry, Topic: "stranger"},
	{Text: "The noise has been going since dusk and shows no sign of stopping", Tones: []Tone{Triumphant}, Reg: rPlain, Topic: "noise"},
	{Text: "Nobody has admitted that the plan was mostly an accident", Tones: []Tone{Wry}, Reg: rJoke, Topic: "rumour"},
	{Text: "The version being told tonight has a sea monster in it", Tones: []Tone{Wry}, Reg: rJoke, Form: fOverheard, Topic: "rumour"},
	{Text: "Somebody has worked out that if you count the days generously and the losses charitably, this was the best thing anyone here has done in a decade", Tones: []Tone{Wry}, Reg: rJoke, Form: fLedger, Topic: "count"},
	{Text: "It was mostly luck and the two people who know that are keeping quiet", Tones: []Tone{Wry}, Reg: rWry, Topic: "rumour"},
	{Text: "The men who stayed behind have been telling everyone what they would have done", Tones: []Tone{Wry}, Reg: rJoke, Form: fOverheard, Topic: "town"},

	// --- the order title, in a frame a verb-led name survives ---
	{Text: "The orders marked {subject} came back signed and filthy", Needs: needSubject, Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "Somebody filed the {subject} order under things that worked", Needs: needSubject, Reg: rWry, Form: fNotice, Topic: "paper"},
	{Text: "A page titled {subject} now has three names crossed out on it", Needs: needSubject, Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "One lamp, one knife, filed under {subject}", Needs: needSubject, Reg: rWry, Form: fLedger, Topic: "paper"},
	{Text: "The line that reads {subject} has a thumbprint on it now", Needs: needSubject, Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "Everything filed under {subject} came home muddy, including the paper it was written on, which somebody had folded into a boot", Needs: needSubject, Reg: rWry, Form: fNotice, Topic: "paper"},
	{Text: "The orders marked {subject} were read out once at the start, then carried nine days in a wet pack, and can no longer be read out at all", Needs: needSubject, Reg: rWry, Form: fNotice, Topic: "paper"},

	// --- the resource, as a thing in a room rather than as a payout ---
	{Text: "Nobody wants to sit up guarding the {res} tonight", Needs: needRes, Reg: rPlain, Topic: "money"},
	{Text: "Two people have already argued about the {res} this evening", Needs: needRes, Reg: rPlain, Topic: "argument"},
	{Text: "The whole town knew about the {res} before the council did", Needs: needRes, Reg: rWry, Topic: "rumour"},
	{Text: "Whoever found the trail wants a bigger share of the {res}", Needs: needRes, Reg: rPlain, Topic: "money"},
	{Text: "A little of the {res} has already gone missing", Needs: needRes, Reg: rPlain, Topic: "money"},
	{Text: "Everything is being counted twice and the {res} three times", Needs: needRes, Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "Whatever room the {res} went into first turned out to be damp, so it has been moved once already, and the second room has a lock somebody has lost the key to", Needs: needRes, Reg: rPlain, Topic: "building"},
	{Text: "Nobody around here expected {res_stores} to be this full", Needs: needRes | needMassRes, Reg: rPlain, Topic: "money"},
	{Text: "The count came to {amt_res}, twice, by two different people", Needs: needRes | needAmount, Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The number being repeated tonight is {amt_res}", Needs: needRes | needAmount, Reg: rPlain, Form: fOverheard, Topic: "count"},
}

// expSuccessGrounded is the part of the old ungated pool whose imagery is a town
// with lanes, dogs and a washing line: true from the Stone Age to the
// Modern Age, and wrong on a station. Moved here after a corpus review.
var expSuccessGrounded = []skel{
	{Text: "The dog came back fat", Reg: rPlain, Topic: "animal"},
	{Text: "There was a dog with them on the way out that was not with them on the way back, and a different dog with them on the way back that was not with them on the way out, and nobody involved finds this worth discussing", Reg: rJoke, Topic: "animal"},
	{Text: "The story as it stands this evening involves a night crossing, a wolf, a broken bridge and a stranger who gave them directions and would not take payment, and only one of those four things happened", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "Everything stinks of wet leather", Reg: rPlain, Topic: "smell"},
	{Text: "Somebody chalked {amt_res} on the wall by the door and somebody else has already rubbed out the last figure and written a bigger one", Needs: needRes | needAmount, Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "The dogs have taken a great interest in ~", Slot: "exp_success_haul", Reg: rWry, Topic: "animal"},
	{Text: "They brought a dog back with them and the dog is staying", Reg: rPlain, Topic: "animal"},
	{Text: "The washing is out and half of it will never be clean again", Reg: rPlain, Topic: "people"},
	{Text: "The first count at the door came to nine sacks and the second came to twelve, and the woman who took the second one has gone to bed", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "The younger of the two brothers has told the story four times and it has acquired a river, a night crossing and a wolf that were not in the first version", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "They came in the back way to avoid the fuss, which worked for about four minutes until the dogs started", Reg: rWry, Topic: "animal"},
	{Text: "The blankets are all wet, the heat has been on since noon, and the room now smells of scorched cloth and everyone has stopped mentioning it", Reg: rPlain, Topic: "smell"},
	{Text: "Whoever packed the water skins packed nine of them for thirteen people, which was survivable in the cold weeks and will be remembered for years", Reg: rWry, Form: fComplaint, Topic: "kit"},
	{Text: "Mud on everything", Reg: rPlain, Topic: "ground"},
	{Text: "One of them lost a boot in the mud on the second day, walked the remaining nine days with rags on that foot, and has refused every offer of a replacement pair on the grounds that he has got used to it", Reg: rWry, Topic: "people"},
}

// expSuccessLate is the digital and cosmic voice: the field notes have become
// logs, and the logs have started to notice how long everything takes.
var expSuccessLate = []skel{
	{Text: "They came back with more readings than anyone alive will have time to look at", Reg: rWry, Topic: "haul"},
	{Text: "The recovered samples have been filed in a vault designed to outlast everyone with the clearance to open it", Reg: rWry, Form: fLedger, Topic: "paper"},
	{Text: "All nine are home", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "One of them keeps asking what day it is, and each time he is told he writes it on the back of his hand, and by the evening meal both hands were full and he had started on his wrist", Reg: rWry, Topic: "time"},
	{Text: "The map of the region they surveyed is complete and accurate to the metre, and there is nothing on it worth going back for", Reg: rWry, Topic: "map"},
	{Text: "Their suits went for cleaning and the cleaners sent one of them back with a question", Reg: rJoke, Topic: "kit"},
	{Text: "Welcome-home drinks are at the usual time in the usual place, and attendance will be noted", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "The youngest of them has asked to see a window", Reg: rPlain, Topic: "people"},
}

// expSuccessAncient — fires, herds, hides, spears, the elders.
var expSuccessAncient = []skel{
	{Text: "Two spears came home broken", Reg: rPlain, Topic: "weapon"},
	{Text: "The hides are stiff with salt", Reg: rPlain, Topic: "haul"},
	{Text: "The store-pit is full", Reg: rPlain, Form: fLedger, Topic: "money"},
	{Text: "Drums until dawn", Reg: rPlain, Topic: "noise"},
	{Text: "The ford was low", Reg: rPlain, Topic: "ground"},
	{Text: "The elders looked at what came back and said nothing useful", Reg: rWry, Topic: "authority"},
	{Text: "They drove a herd back and not all of it is theirs", Reg: rWry, Topic: "animal"},
	{Text: "Somebody has planted a spear outside a doorway that is not his", Reg: rWry, Topic: "argument"},
	{Text: "The watch-fire was kept lit three nights for nothing", Reg: rPlain, Form: fComplaint, Topic: "border"},
	{Text: "An arrow came back in a shoulder, still whole", Reg: rPlain, Topic: "wound"},
	{Text: "They found a spring on the far side and drank it low", Reg: rPlain, Topic: "ground"},
	{Text: "The valley beyond has better grass and worse people", Reg: rWry, Topic: "map"},
	{Text: "The elders want the spears counted before anybody eats", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "A new hut is going up before the mud has dried", Reg: rPlain, Topic: "building"},
	{Text: "They brought back stone that rings when you strike it", Reg: rPlain, Topic: "haul"},
	{Text: "A hide came back with marks on it that nobody can read", Reg: rPlain, Topic: "map"},
	{Text: "The old man who knows the ford has been proved right again and is being insufferable about it in the way of a man who has waited years", Reg: rWry, Topic: "name"},
	{Text: "The drums stopped when the last of them was counted in, and started again an hour later for no reason anybody has explained", Reg: rPlain, Topic: "noise"},
	{Text: "A boy carried a spear the whole way and never once used it, and has been carrying it around the huts all evening in case anybody asks", Reg: rWry, Topic: "weapon"},
	{Text: "The herd lost two on the way and the boy who was watching them has been crying since the middle of the afternoon", Reg: rPlain, Topic: "animal"},
	{Text: "Three of them walked the last stretch under one hide, which kept the rain off two of them, and there has been discussion about which two", Reg: rWry, Topic: "weather"},
	{Text: "The hides go to the women who worked the last lot, the spears to whoever carried them, and the stone to nobody yet because the elders have not finished arguing", Reg: rPlain, Form: fLedger, Topic: "argument"},
	{Text: "Smoke on the ridge turned out to be somebody else's trouble, which was established by two of them walking most of a day toward it and most of a day back", Reg: rPlain, Topic: "border"},
	{Text: "The harvest can wait, apparently, because everybody who should be out in it is instead standing around the store-pit looking at hides as though they might change", Reg: rWry, Form: fComplaint, Topic: "town"},
}

// expSuccessFeudal — carts, gates, bells, clerks, stewards, musters.
var expSuccessFeudal = []skel{
	{Text: "The gate was opened early", Reg: rPlain, Topic: "border"},
	{Text: "One cart came back on three wheels", Reg: rPlain, Topic: "machine"},
	{Text: "The bells went at dusk", Reg: rPlain, Topic: "noise"},
	{Text: "Wax across the only clean parchment", Reg: rPlain, Form: fComplaint, Topic: "paper"},
	{Text: "The undercroft is full", Reg: rPlain, Form: fLedger, Topic: "money"},
	{Text: "The quartermaster has counted it and wants it counted again", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "A rider went ahead with the news and beat them in by an hour", Reg: rPlain, Topic: "message"},
	{Text: "The clerks are still writing and supper is going cold", Reg: rPlain, Form: fComplaint, Topic: "paper"},
	{Text: "One banner came home in pieces and is being sewn tonight", Reg: rPlain, Topic: "kit"},
	{Text: "The militia turned out for nothing and are pretending otherwise", Reg: rWry, Topic: "town"},
	{Text: "The wagons are in the yard and nobody will unload them", Reg: rWry, Form: fComplaint, Topic: "haul"},
	{Text: "A herald has already made the story longer than it was", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "They came through the village at dusk and woke all of it", Reg: rPlain, Topic: "town"},
	{Text: "The pickets saw them coming and rang nothing at all", Reg: rPlain, Topic: "border"},
	{Text: "The muster roll is two names shorter than it was", Reg: rPlain, Form: fLedger, Topic: "casualty"},
	{Text: "The steward wants the tally before the men want their supper", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "A horn went up at the gate and half of it was flat", Reg: rJoke, Topic: "noise"},
	{Text: "The village priest has claimed a share and is being ignored", Reg: rWry, Topic: "religion"},
	{Text: "The clerks have been writing since noon, and the steward has sent to the monastery for more ink and a boy who can spell", Reg: rWry, Form: fComplaint, Topic: "paper"},
	{Text: "A cart wheel came off in the yard, at walking pace, in front of about thirty people, and the carter has not been allowed to forget any part of it", Reg: rJoke, Topic: "machine"},
	{Text: "Two riders came in muddy to the waist and have offered no account of how, and the steward has stopped asking because the answers keep changing", Reg: rWry, Topic: "people"},
	{Text: "The storehouse doors have stood open since noon and there is a boy sitting on a stool beside them who was told to watch and has been watching very hard", Reg: rPlain, Topic: "building"},
	{Text: "The harvest stood untouched all afternoon because everybody who should have been in it was up on the wall watching the road, and the steward has said his piece about that", Reg: rWry, Form: fComplaint, Topic: "town"},
	{Text: "The bells go again at first light, by order of nobody anyone can name, and the boy who rang them last night is being watched by several people at once", Reg: rWry, Form: fNotice, Topic: "noise"},
}

// expSuccessIndustrial — depots, sidings, foremen, telegrams, triplicate.
var expSuccessIndustrial = []skel{
	{Text: "The depot clock has stopped", Reg: rPlain, Topic: "time"},
	{Text: "Two crates went missing", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "Mud to the axles", Reg: rPlain, Topic: "machine"},
	{Text: "The payroll is short again", Reg: rPlain, Form: fComplaint, Topic: "money"},
	{Text: "Everything is in triplicate", Reg: rWry, Form: fNotice, Topic: "paper"},
	{Text: "The telegram arrived a full day before the men did", Reg: rPlain, Topic: "message"},
	{Text: "Freight came in on the evening train and sat on the platform", Reg: rPlain, Topic: "haul"},
	{Text: "The foreman has signed for all of it and gone home", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "Somebody rang the office telephone at two in the morning", Reg: rPlain, Topic: "message"},
	{Text: "There is a wire on the foreman's desk that nobody wants to open", Reg: rPlain, Topic: "message"},
	{Text: "The annexe is full and they are stacking it in the corridor", Reg: rPlain, Topic: "building"},
	{Text: "The shareholders will hear a considerably tidier version", Reg: rWry, Topic: "rumour"},
	{Text: "The men came off the train singing and were told to stop", Reg: rWry, Topic: "noise"},
	{Text: "The warehouse roof leaks and it all went under there anyway", Reg: rWry, Form: fComplaint, Topic: "building"},
	{Text: "The telegraph operator has been awake for two days", Reg: rPlain, Topic: "sleep"},
	{Text: "The depot cat has moved into the new crates", Reg: rWry, Topic: "animal"},
	{Text: "A photograph was taken and everybody in it is squinting", Reg: rWry, Topic: "town"},
	{Text: "The foreman's tally and the office tally differ by one crate, and both men have now written to the other explaining, at length, why the difference is not theirs", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "A rail truck was left uncoupled at the top of the siding and rolled down into the fence at a walking pace, and the fence has been mended and the matter has not", Reg: rWry, Topic: "machine"},
	{Text: "They telegraphed ahead and the message arrived so badly garbled that the yard spent an afternoon preparing for eleven horses that were never mentioned in the original", Reg: rJoke, Form: fNotice, Topic: "message"},
	{Text: "Two of the column came home on stretchers and walked off them at the platform, which the doctor has described as a matter he intends to take up with both of them", Reg: rWry, Topic: "wound"},
	{Text: "The night shift was told to go home and worked through anyway, and the men who did it have spent the morning making sure everybody knows they did", Reg: rWry, Topic: "people"},
	{Text: "The siding is blocked until somebody moves the empty trucks, and the man who moves the empty trucks is on the evening train, which cannot get in because the siding is blocked", Reg: rJoke, Form: fComplaint, Topic: "machine"},
	{Text: "There is a new lock on the warehouse door and three keys were cut for it, and as of this evening two of the three are unaccounted for and the third is in a drawer nobody can open", Reg: rWry, Form: fLedger, Topic: "building"},
}

// expSuccessDigital — uplinks, drones, feeds, analysts, channels.
var expSuccessDigital = []skel{
	{Text: "The uplink held", Reg: rPlain, Topic: "machine"},
	{Text: "The drone came home with its camera facing backwards", Reg: rPlain, Form: fLedger, Topic: "machine"},
	{Text: "The channel logs are gone", Reg: rPlain, Topic: "paper"},
	{Text: "The floor has been online since the weekend", Reg: rPlain, Topic: "sleep"},
	{Text: "Coffee on the console", Reg: rPlain, Topic: "smell"},
	{Text: "The feed cut at the worst moment and came back at the dullest", Reg: rWry, Topic: "machine"},
	{Text: "Two analysts are still arguing over the same forty seconds", Reg: rWry, Topic: "argument"},
	{Text: "The after-action report is longer than the operation was", Reg: rWry, Form: fNotice, Topic: "paper"},
	{Text: "Somebody left the channel open and the whole floor listened in", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The network went down for six minutes and nobody noticed", Reg: rPlain, Topic: "machine"},
	{Text: "A junior analyst spotted it first and will never stop saying so", Reg: rWry, Topic: "people"},
	{Text: "The feed from the second team is still buffering, hours later", Reg: rPlain, Form: fComplaint, Topic: "machine"},
	{Text: "One drone came home by a route it had not been given", Reg: rWry, Topic: "machine"},
	{Text: "The uplink logs show a gap that nobody wants to explain", Reg: rPlain, Topic: "paper"},
	{Text: "An analyst went home, slept four hours, and came back in", Reg: rPlain, Topic: "sleep"},
	{Text: "The drone has been given a name and the name has stuck", Reg: rWry, Topic: "name"},
	{Text: "The night channel was silent for two hours and then very loud", Reg: rPlain, Topic: "noise"},
	{Text: "The drone footage has been watched sixty-odd times this evening, mostly by people who were not on the operation and have opinions about how it went", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The after-action briefing has been moved twice, once because the room was taken and once because the man giving it has been asleep at his desk since noon", Reg: rWry, Form: fNotice, Topic: "sleep"},
	{Text: "Two laptops came back and one of them still switches on, which is the better ratio of the two operations this month and is being mentioned as though it were an achievement", Reg: rWry, Form: fLedger, Topic: "machine"},
	{Text: "Someone's badge is still on the desk where they left it and the desk has been walked past all afternoon by people who have decided it is not their business", Reg: rPlain, Topic: "casualty"},
	{Text: "The uplink dish took a stone at some point in the second week and carried on working, and the engineer who bolted it together has said several times that she is not surprised", Reg: rWry, Topic: "machine"},
	{Text: "A second drone was sent out to find the first one, found it, and then both of them came home together at a speed the flight log describes as unusual", Reg: rWry, Form: fLedger, Topic: "machine"},
	{Text: "The channel is full of people asking what happened, and the four people who could answer are all asleep, and the two people who are answering were not there", Reg: rWry, Form: fOverheard, Topic: "rumour"},
}

// expSuccessCosmic — hulls, bays, relays, transponders, airlocks.
var expSuccessCosmic = []skel{
	{Text: "The hull is scored down one side", Reg: rPlain, Topic: "machine"},
	{Text: "Bay three is full", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "Vacuum frost on everything", Reg: rPlain, Topic: "haul"},
	{Text: "The new scrape along the outer seam already has a nickname", Reg: rPlain, Topic: "name"},
	{Text: "The reactor is still ticking", Reg: rPlain, Topic: "machine"},
	{Text: "They came out of orbit hot and the dock crew swore at them", Reg: rWry, Topic: "authority"},
	{Text: "The transponder was dead for eleven hours and then it was fine", Reg: rPlain, Topic: "message"},
	{Text: "Something is loose behind a bulkhead and nobody can find it", Reg: rPlain, Topic: "machine"},
	{Text: "The airlock cycled twice on the way in", Reg: rPlain, Topic: "machine"},
	{Text: "The relay picked them up before the watch officer did", Reg: rWry, Topic: "border"},
	{Text: "A freighter crew saw them pass and logged it as debris", Reg: rWry, Form: fLedger, Topic: "message"},
	{Text: "Two of the crew have not taken the suits off yet", Reg: rPlain, Topic: "people"},
	{Text: "The manifest and the bay count differ by one crate", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The dock lights were left burning all night for them", Reg: rPlain, Topic: "town"},
	{Text: "The bay doors froze half open and were beaten with a wrench", Reg: rWry, Topic: "machine"},
	{Text: "They spent the last day of the run on emergency air", Reg: rPlain, Topic: "wound"},
	{Text: "The transponder code came back wrong and was accepted anyway, because the officer on watch recognised the approach and could not be bothered with the paperwork at that hour", Reg: rWry, Form: fNotice, Topic: "border"},
	{Text: "A bulkhead panel came away in somebody's hand three days out and has been held in place since by a strap, a wedge and an agreement not to look at it", Reg: rWry, Topic: "machine"},
	{Text: "Orbital watch has logged the approach as clean, which the four people who were on the deck at the time have declined, politely and separately, to comment on", Reg: rWry, Form: fNotice, Topic: "border"},
	{Text: "Everything in the forward bay smells of scorched insulation and the smell has got into the corridor, the mess and, according to one of them, the food", Reg: rPlain, Form: fComplaint, Topic: "smell"},
	{Text: "The dock crew found a boot in the cargo net during the unload and it has been sitting on the ledge by the airlock since, and nobody has claimed it", Reg: rWry, Topic: "kit"},
	{Text: "The hull sensors are reporting the ship as sound while the ship makes a noise every ninety seconds that the engineer has started writing down the timings of", Reg: rWry, Topic: "machine"},
	{Text: "A repair patch made of tape and a cut-down panel is holding a whole compartment, and the inspector who looked at it this morning signed the sheet and left the remarks box empty", Reg: rWry, Topic: "machine"},
	{Text: "The freighter that lent them fuel at the halfway point would like it back, has said so through the relay four times since yesterday, and is being answered by nobody", Reg: rWry, Form: fNotice, Topic: "message"},
}
