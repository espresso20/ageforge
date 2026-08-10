package flavor

// EncounterStandoff — contact with a civilization we are at war with, and both
// sides walked away from it.
//
// The log line directly above this one already says the encounter produced
// nothing, so none of these sentences report that. They report the afternoon:
// who was on watch and what he thought he saw, what got shouted and whether
// anybody understood a word of it, the interpreter, the body handed back, the
// dog that had to be fetched, what was eaten standing up, what the officers
// wrote down afterwards, and how much the story has grown since. See catalog.go
// for the rules.
//
// The texture this file is aiming at is the quartermaster's book of
// catalog_exp_success.go rather than a book of aphorisms: circumstantial, often
// mundane, occasionally digressive, and funny by accident about one line in
// fifteen. The lengths swing on purpose — a four-word line beside a forty-word
// one — because uniform rhythm is what made the two previous versions of this
// catalog read as machine-written.

// --- slot banks --------------------------------------------------------------
//
// Two categories, both true in every age, because the sentences that draw them
// are ungated: a measured gap, and a small act. Lowercase noun phrases, no
// terminal punctuation, nothing era-coded.

var encStandoffBanks = map[string][]string{
	// The gap between the two lines, as somebody afterwards described it.
	"enc_standoff_distance": {
		"twenty paces", "forty paces", "sixty paces", "a hundred paces",
		"two hundred paces", "a dozen paces", "a good ten paces",
		"a careful fifty paces", "eighty paces of open ground",
		"thirty paces of bad footing", "a short throw", "a long throw",
		"ninety paces or so", "the length of a shout",
		"the distance a voice carries", "an arm's length",
		"the width of the water", "a hundred and fifty paces",
	},
	// The small acts. Any one of them was the whole of what was exchanged.
	"enc_standoff_gesture": {
		"a raised hand", "a nod nobody returned", "a flat open palm",
		"a bow of the head", "two fingers lifted", "a slow wave", "a shrug",
		"a stiff nod", "a long look", "a step backwards", "a shaken head",
		"a hand held out empty", "a whistle with no answer", "a mouthed word",
		"a hand raised and dropped", "a knuckle to the forehead",
		"a wave that was almost friendly", "a lowered weapon",
	},
}

// --- the sentences -----------------------------------------------------------

// encStandoffTemplates is the EncounterStandoff skeleton set: one ungated pool
// that lands in any age, plus one pool per era bucket for the imagery that does
// not travel.
func encStandoffTemplates() []tmpl {
	out := pool("enc_standoff_any", erasAny, encStandoffAny)
	out = append(out, pool("enc_standoff_ancient", erasAncient, encStandoffAncient)...)
	out = append(out, pool("enc_standoff_feudal", erasFeudal, encStandoffFeudal)...)
	out = append(out, pool("enc_standoff_industrial", erasIndustrial, encStandoffIndustrial)...)
	out = append(out, pool("enc_standoff_digital", erasDigital, encStandoffDigital)...)
	out = append(out, pool("enc_standoff_cosmic", erasCosmic, encStandoffCosmic)...)
	return out
}

// encStandoffAny fires in every age. Watching, waiting, shouting, water,
// weather, children, dogs and being lied to about it afterwards are timeless;
// gates, sidings and airlocks are not.
var encStandoffAny = []skel{
	// --- short. A fact, and then the sentence stops. ---
	{Text: "Neither side ate first", Reg: rPlain, Topic: "food"},
	{Text: "The waiting took four hours", Reg: rPlain, Topic: "time"},
	{Text: "Nobody sat down once", Reg: rPlain, Topic: "people"},
	{Text: "Somebody's dog went over", Reg: rPlain, Topic: "animal"},
	{Text: "Water was shared, grudgingly", Reg: rWry, Topic: "ground"},
	{Text: "The wind never dropped", Reg: rPlain, Topic: "weather"},
	{Text: "Two names were shouted across", Reg: rPlain, Topic: "name"},
	{Text: "Nine of us, forty of them", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The interpreter was useless", Reg: rWry, Topic: "message"},
	{Text: "Bread was eaten standing up", Reg: rPlain, Topic: "food"},
	{Text: "A body was handed back", Reg: rPlain, Topic: "casualty"},
	{Text: "Everyone stayed on their feet", Reg: rPlain, Topic: "people"},
	{Text: "The far side sat down", Reg: rPlain, Topic: "border"},
	{Text: "Nobody was told anything", Reg: rPlain, Form: fComplaint, Topic: "authority"},
	{Text: "What crossed the gap was ~", Slot: "enc_standoff_gesture", Reg: rPlain, Topic: "people"},

	// --- mid. The working length: one thing, with enough detail to place it. ---
	{Text: "The man on watch counted them twice and got two different numbers", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "Whatever was shouted from the far side, nobody here could make it out", Reg: rPlain, Topic: "noise"},
	{Text: "Our interpreter got three words of it and guessed at the rest", Reg: rWry, Topic: "message"},
	{Text: "They handed a body over at the halfway point and stepped back", Reg: rPlain, Topic: "casualty"},
	{Text: "Both sides used the same water and took turns about it", Reg: rPlain, Topic: "ground"},
	{Text: "A dog crossed the open ground and had to be fetched back", Reg: rPlain, Topic: "animal"},
	{Text: "The waiting went on so long that people began sitting down", Reg: rPlain, Topic: "time"},
	{Text: "A quiet swap was made in the middle and both men have denied it", Reg: rWry, Topic: "trade"},
	{Text: "It rained on both of them for the whole of the afternoon", Reg: rPlain, Topic: "weather"},
	{Text: "The story going round tonight has twice as much shouting in it", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The officers wrote down the hour it started and very little else", Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "One of ours went for his weapon and was sat on immediately", Reg: rPlain, Topic: "weapon"},
	{Text: "Cold food, eaten standing, with everybody watching everybody else", Reg: rPlain, Topic: "food"},
	{Text: "A name was shouted and a man near the front went white", Reg: rPlain, Topic: "name"},
	{Text: "The children were kept back and watched the whole thing anyway", Reg: rPlain, Topic: "town"},
	{Text: "Nobody on either side wanted to be the one who moved", Reg: rPlain, Topic: "people"},
	{Text: "Estimates of their numbers went from sixty to ninety and were then left alone", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "An old woman on our side shouted something rude in their language", Reg: rWry, Topic: "noise"},
	{Text: "Two of theirs were recognised and neither name has been said aloud", Reg: rPlain, Topic: "name"},
	{Text: "The ground between them is churned to mud from all the standing", Reg: rPlain, Topic: "ground"},
	{Text: "Somebody has been telling it as a great victory since about noon", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "Whatever they had been eating carried across on the wind for two hours", Reg: rWry, Topic: "smell"},
	{Text: "The men who held the left have complained about the ground ever since", Reg: rPlain, Form: fComplaint, Topic: "ground"},
	{Text: "The cook was given no warning and says so to anyone who passes", Reg: rWry, Form: fComplaint, Topic: "food"},
	{Text: "Sleep was poor afterwards and the ones who say otherwise are lying", Reg: rWry, Topic: "sleep"},
	{Text: "A bet was taken on which side would leave first and nobody has paid", Reg: rWry, Topic: "money"},
	{Text: "The far side had better boots and everybody here noticed that", Reg: rWry, Topic: "kit"},
	{Text: "Somebody's mother came out to look and was walked back indoors", Reg: rPlain, Topic: "town"},
	{Text: "There was one patch of shade and neither side went near it", Reg: rPlain, Topic: "ground"},
	{Text: "Both lines stopped at ~ and stayed there", Slot: "enc_standoff_distance", Reg: rPlain, Topic: "border"},
	{Text: "The gap was paced out afterwards at ~", Slot: "enc_standoff_distance", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "Nobody on our side came closer than ~ all afternoon", Slot: "enc_standoff_distance", Reg: rPlain, Topic: "border"},
	{Text: "A boy sent out with water walked ~ and came back the same way", Slot: "enc_standoff_distance", Reg: rPlain, Topic: "people"},
	{Text: "Their oldest man managed ~ before turning away", Slot: "enc_standoff_gesture", Reg: rPlain, Topic: "stranger"},
	{Text: "One of ours offered ~ and has been talked about for it since", Slot: "enc_standoff_gesture", Reg: rWry, Topic: "argument"},
	{Text: "The whole of it came down to ~ and a great deal of standing about", Slot: "enc_standoff_gesture", Reg: rWry, Topic: "people"},
	{Text: "It ended with ~ from somebody nobody on this side can name", Slot: "enc_standoff_gesture", Reg: rPlain, Topic: "stranger"},

	// --- long. Circumstantial, subordinated, the way a chronicle actually runs. ---
	{Text: "The man on the left of our line had been on his feet since before dawn and gave an account afterwards that had the far side arriving from two different directions", Reg: rPlain, Topic: "people"},
	{Text: "The shouting went back and forth for a quarter of an hour before anybody worked out that the two sides were using different words for the same river", Reg: rWry, Topic: "message"},
	{Text: "A body was carried out to the middle by four of theirs and left on the grass, and two of ours went and got it without being told to", Reg: rPlain, Topic: "casualty"},
	{Text: "Both sides drank from the same water within an hour of each other, upstream and down, and there has been an argument since about who got the better end", Reg: rWry, Topic: "ground"},
	{Text: "The dog belongs to somebody in the second row and it went straight across, and a boy of about nine had to walk out into the open and carry it back", Reg: rPlain, Topic: "animal"},
	{Text: "Food had not been thought about at all, so the whole afternoon was got through on what happened to be in people's pockets", Reg: rPlain, Topic: "food"},
	{Text: "A trade was done quietly at the left-hand end while everybody else was watching the middle, and the two men involved have each denied it twice", Reg: rWry, Form: fOverheard, Topic: "trade"},
	{Text: "The weather turned about halfway through and the rain came in sideways, and neither side would give the other the satisfaction of moving into the lee of the hill", Reg: rWry, Topic: "weather"},
	{Text: "The version being told in the town tonight has three times the numbers, a river crossing, and a speech that nobody who was there remembers hearing", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "What went into writing afterwards was the hour it began, the hour it ended, an estimate of their strength, and one line about the ground being poor", Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "A name was shouted across the gap by somebody on their side, and a woman four ranks back in ours answered it before she could stop herself", Reg: rPlain, Topic: "name"},
	{Text: "There were children on the slope behind us the entire time, sent away twice, and back inside the quarter hour on both occasions", Reg: rWry, Topic: "town"},
	{Text: "One man counted the far side at eighty and another at two hundred and forty, and the two of them were still going at it after dark", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "An old man walked out from their line with both hands open and stood in the middle a long while before anybody would come and hear him", Reg: rPlain, Topic: "stranger"},
	{Text: "The ones who were nearest have been asked what was said and have given four different answers, none of which agree on the language it was in", Reg: rPlain, Form: fOverheard, Topic: "message"},
	{Text: "The two front ranks settled at ~, which held for the rest of the day and was measured afterwards by a man who had nothing else to do", Slot: "enc_standoff_distance", Reg: rWry, Topic: "count"},
	{Text: "Shouting carried fine across ~, and the insults were understood well enough on both sides for two men to be told to stop", Slot: "enc_standoff_distance", Reg: rWry, Topic: "noise"},
	{Text: "A woman at their front gave ~, held it a moment, and dropped her arm, and the men behind her began packing up before she had turned round", Slot: "enc_standoff_gesture", Reg: rPlain, Topic: "people"},
	{Text: "Somebody near the middle answered with ~ and has been asked since whether that was wise, and has said each time that it seemed reasonable at the hour", Slot: "enc_standoff_gesture", Reg: rWry, Topic: "argument"},

	// --- very long. Digressive, specific, a paragraph that forgot to stop. ---
	{Text: "Our interpreter learned the language forty years ago from a trader who has been dead for twenty of them, and she caught perhaps half of what was shouted across, and she will only admit to a third of that", Reg: rWry, Topic: "message"},
	{Text: "The dog went over at some point in the second hour, was fed by three separate people on the far side, and had to be walked back by a boy sent out with both hands in the air and no idea what he was doing", Reg: rWry, Topic: "animal"},
	{Text: "There is a man in the second row who put his hand up and shouted something that could have started the whole thing, and he has been asked about it by four separate people and has stopped answering his door", Reg: rPlain, Topic: "argument"},
	{Text: "What got eaten in those four hours was two loaves passed down the line, a bag of dried fruit belonging to somebody who has not been thanked for it, and the water that had been meant for the walk home", Reg: rPlain, Form: fLedger, Topic: "food"},
	{Text: "The account written up afterwards runs to nine lines, six of which are about the ground and the weather, and the other three say that contact was made at the hour stated and that the far side withdrew at its own pace", Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "The story has been told in the town four times since last night and has acquired, in order, a river, a wolf, an insult in a language nobody there speaks, and a man who walked out alone into the middle and came back grinning", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "Nobody here has much to say about the far side's numbers, their weapons or their intentions, and everybody here has a very great deal to say about how long they were made to stand in one place with wet feet", Reg: rWry, Form: fComplaint, Topic: "people"},

	// --- kind: aggressive ---
	{Text: "Blood was wanted", Kinds: []string{"aggressive"}, Reg: rPlain, Topic: "weapon"},
	{Text: "The front rank on their side came in fast and stopped short", Kinds: []string{"aggressive"}, Reg: rPlain, Topic: "weapon"},
	{Text: "One of them spat and left it at that", Kinds: []string{"aggressive"}, Reg: rWry, Topic: "argument"},
	{Text: "They had come out for a fight and stood about for four hours instead, and the man who brought them out has been hearing about it since", Kinds: []string{"aggressive"}, Reg: rWry, Topic: "argument"},
	{Text: "A captain over there drew a weapon, held it up over his head so that everybody on both sides of the ground could see it, and then put it away again without saying one word to anybody", Kinds: []string{"aggressive"}, Reg: rPlain, Topic: "weapon"},

	// --- kind: mercantile ---
	{Text: "They tried to sell us something halfway through", Kinds: []string{"mercantile"}, Reg: rJoke, Topic: "trade"},
	{Text: "A price was called out across the gap", Kinds: []string{"mercantile"}, Reg: rWry, Form: fOverheard, Topic: "trade"},
	{Text: "Their side counted our numbers twice, carefully, the way a man counts out coin before he hands it over", Kinds: []string{"mercantile"}, Reg: rWry, Form: fLedger, Topic: "count"},

	// --- kind: isolationist ---
	{Text: "Not one of them spoke the whole time", Kinds: []string{"isolationist"}, Reg: rPlain, Topic: "noise"},
	{Text: "They wanted us gone more than they wanted us dead", Kinds: []string{"isolationist"}, Reg: rPlain, Topic: "border"},
	{Text: "A marker was put down at the edge of the ground they claim, and every one of them walked back behind it and stood there until we had gone", Kinds: []string{"isolationist"}, Reg: rPlain, Topic: "border"},

	// --- tone ---
	{Text: "Two armies stood in the rain being scrupulously polite", Tones: []Tone{Wry}, Reg: rJoke, Topic: "weather"},
	{Text: "The most dangerous moment of the whole afternoon was a sneeze", Tones: []Tone{Wry}, Reg: rJoke, Topic: "noise"},
	{Text: "Both sides have marched off telling it as a win", Tones: []Tone{Wry}, Reg: rWry, Topic: "rumour"},
	{Text: "It was ended by everybody getting bored at roughly the same time, which took rather longer than anyone will admit to in the morning", Tones: []Tone{Wry}, Reg: rWry, Topic: "time"},
	{Text: "Somebody has worked out that if you count the standing about as fighting, this was the longest engagement anybody here has ever been in", Tones: []Tone{Wry}, Reg: rJoke, Form: fLedger, Topic: "count"},
	{Text: "Nobody sang on the way back", Tones: []Tone{Grim}, Reg: rPlain, Topic: "sleep"},
	{Text: "A body came over and nobody knew the face", Tones: []Tone{Grim}, Reg: rPlain, Topic: "casualty"},
	{Text: "The wounded man they let through died in the night", Tones: []Tone{Grim}, Reg: rPlain, Topic: "wound"},
	{Text: "The two men who were told to hold the left have not eaten since and will not be spoken to, and one of them has been sat in the same place since dark", Tones: []Tone{Grim}, Reg: rPlain, Topic: "people"},
	{Text: "There is drink out and somebody has already fallen over", Tones: []Tone{Triumphant}, Reg: rJoke, Topic: "town"},

	// --- the enemy, in frames a mixed-number name survives ---
	{Text: "Somebody shouted the name {subject} across the gap and got no answer to it", Needs: needSubject, Reg: rPlain, Topic: "name"},
	{Text: "Nobody here had ever stood that close to {subject} before", Needs: needSubject, Reg: rPlain, Topic: "stranger"},
	{Text: "The men spent the evening arguing about what {subject} wanted out of it", Needs: needSubject, Reg: rWry, Topic: "argument"},
	{Text: "The long war with {subject} stopped for one afternoon and started again by dark", Needs: needSubject, Reg: rWry, Topic: "border"},
	{Text: "Two of ours walked out to meet {subject} and walked back at the same pace", Needs: needSubject, Reg: rPlain, Topic: "people"},
	{Text: "Somebody drew a picture of {subject} in the dirt, unkindly, and it was rubbed out before the officers came past", Needs: needSubject, Reg: rWry, Topic: "argument"},
	{Text: "The order went round not to fire on {subject} without a word from the front, and it went round twice more before anybody believed it had come from where it said", Needs: needSubject, Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "There is an old man here who fought {subject} thirty years ago in a different place, and he walked the length of our line during it, looking, and he has never said what he was looking for", Needs: needSubject, Reg: rPlain, Topic: "name"},

	// --- the resource, as a thing in a room rather than as a payout ---
	{Text: "Somebody offered {res} across the gap and got a shaken head for it", Needs: needRes, Reg: rPlain, Topic: "trade"},
	{Text: "Everything got carried up to the line, including the {res}, and carried back down again unopened", Needs: needRes, Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "Nobody wanted the job of guarding the {res} while all that was going on", Needs: needRes, Reg: rPlain, Topic: "money"},
	{Text: "There has been an argument since about who was meant to be watching the {res}", Needs: needRes, Reg: rPlain, Topic: "argument"},
	{Text: "The far side asked after the {res} by name, which nobody here has an explanation for", Needs: needRes, Reg: rPlain, Topic: "rumour"},
	{Text: "The store where the {res} is kept was left open the whole time and nothing at all went missing from it", Needs: needRes, Reg: rWry, Topic: "building"},
	{Text: "Two men were pulled off the {res} and put on the line, and neither of them has been put back, and both have mentioned it", Needs: needRes, Reg: rWry, Form: fComplaint, Topic: "money"},
	{Text: "The figure going round tonight is {amt_res}", Needs: needRes | needAmount, Reg: rPlain, Form: fOverheard, Topic: "count"},
	{Text: "Somebody has priced the whole afternoon at {amt_res} in lost work", Needs: needRes | needAmount, Reg: rWry, Form: fLedger, Topic: "money"},
	{Text: "A man on our side has worked out that standing about for four hours cost the town something like {amt_res}, and he has told nine people, and two of them have asked him to do the sum again", Needs: needRes | needAmount, Reg: rJoke, Form: fLedger, Topic: "money"},
}

// encStandoffAncient — elders, spears, herds, hides, the ford, the watch-fire.
var encStandoffAncient = []skel{
	{Text: "The drums stopped early", Reg: rPlain, Topic: "noise"},
	{Text: "Two spears, planted, left", Reg: rPlain, Form: fLedger, Topic: "weapon"},
	{Text: "Both herds grazed the same slope", Reg: rPlain, Topic: "animal"},
	{Text: "The ford ran low", Reg: rPlain, Topic: "ground"},
	{Text: "Elders talked all night", Reg: rPlain, Topic: "authority"},
	{Text: "An arrow was nocked and put away again without being drawn", Reg: rPlain, Topic: "weapon"},
	{Text: "Both sides watered at the ford and neither hurried about it", Reg: rPlain, Topic: "ground"},
	{Text: "The watch-fire was fed all night by two boys who saw nothing", Reg: rPlain, Topic: "border"},
	{Text: "Somebody rolled a hide out on the grass and sat on it, watching", Reg: rPlain, Topic: "people"},
	{Text: "The spears on the far side were better made and every man here saw that", Reg: rWry, Topic: "weapon"},
	{Text: "A spear was thrown short on purpose and left where it landed", Reg: rPlain, Topic: "weapon"},
	{Text: "The elders want the whole thing described to them again in the morning", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Nobody went near the store-pit while they were stood out there", Reg: rPlain, Topic: "building"},
	{Text: "An old woman shouted at them from behind our spears and was heard", Reg: rWry, Topic: "noise"},
	{Text: "Smoke from their fires and ours went straight up all afternoon", Reg: rPlain, Topic: "weather"},
	{Text: "The boys who fed the watch-fire want it known that nobody relieved them", Reg: rWry, Form: fComplaint, Topic: "border"},
	{Text: "The herds were driven apart before anything could start, which took most of an hour and involved a good deal of shouting at animals rather than at people", Reg: rWry, Topic: "animal"},
	{Text: "One of the elders walked out alone to the middle of the ground, stood there a while, and came back at a considerably slower pace than he went out", Reg: rPlain, Topic: "authority"},
	{Text: "A hide was carried across as a gift and handed straight back, and the man who carried it has been asked four times what was said to him", Reg: rPlain, Topic: "trade"},
	{Text: "The harvest was standing in the valley the whole time and neither lot went into it, which the old men have been remarking on since with something like surprise", Reg: rWry, Topic: "town"},
	{Text: "Their dogs and ours worked out where the line was long before the men did, and settled it between themselves at the water without anybody getting bitten", Reg: rWry, Topic: "animal"},
	{Text: "Frost on the grass at first light, two lines of men standing in it, and a boy of about ten going up and down our side with a water skin", Reg: rPlain, Topic: "weather"},
	{Text: "The elders have talked about it for three nights and have arrived at the position that the spears should have gone in on the first afternoon, which was not the position any of them held on the first afternoon", Reg: rJoke, Topic: "authority"},
	{Text: "A spear came down beside the ford in the second hour, thrown by a young man on their side who was taken by the arm and walked back through his own line, and nothing else went across all day", Reg: rPlain, Topic: "weapon"},
}

// encStandoffFeudal — gates, bells, heralds, pickets, stewards, the muster.
var encStandoffFeudal = []skel{
	{Text: "The gate stayed shut", Reg: rPlain, Topic: "border"},
	{Text: "One bell, then nothing", Reg: rPlain, Topic: "noise"},
	{Text: "Supper was late", Reg: rPlain, Form: fComplaint, Topic: "food"},
	{Text: "Both banners came down", Reg: rPlain, Topic: "border"},
	{Text: "The muster was stood down", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "A herald went out, said his piece, and came back very thirsty", Reg: rJoke, Topic: "message"},
	{Text: "Two riders met at the ford and talked about the weather", Reg: rPlain, Topic: "weather"},
	{Text: "The clerks have written it up and made it duller than it was", Reg: rWry, Form: fNotice, Topic: "paper"},
	{Text: "A cart of grain was let through the middle and nobody touched it", Reg: rPlain, Topic: "trade"},
	{Text: "Their steward and ours discussed the harvest, of all things", Reg: rJoke, Topic: "trade"},
	{Text: "The pickets on both sides could see each other all night", Reg: rPlain, Topic: "border"},
	{Text: "Somebody sent a jug over and it came back empty", Reg: rJoke, Topic: "food"},
	{Text: "The village between the two camps was left alone by both", Reg: rPlain, Topic: "town"},
	{Text: "A horn went up on their side and got no answer", Reg: rPlain, Topic: "noise"},
	{Text: "The militia went home to the harvest and were glad to", Reg: rWry, Topic: "town"},
	{Text: "An arrow stuck in the gate post and has been left there", Reg: rPlain, Topic: "weapon"},
	{Text: "The quartermaster counted their column at four hundred, stopped counting, and wrote down four hundred, and has been asked about it twice since by men who counted more", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "A boy carried a message across the ground on foot because no rider would go, and he was given bread on the far side and sent back with nothing written down", Reg: rPlain, Topic: "message"},
	{Text: "Wax was melted for a seal on an agreement neither steward believed in, and the parchment it was pressed into has already been folded into somebody's boot", Reg: rWry, Form: fNotice, Topic: "paper"},
	{Text: "The storehouse doors were shut for the first time since the spring, and the man with the key had to be found, and he was asleep in the undercroft", Reg: rPlain, Topic: "building"},
	{Text: "Both columns ate supper within sight of one another in the yard beyond the ford, at about the same hour, in silence apart from the carts being unloaded", Reg: rPlain, Topic: "food"},
	{Text: "The militia want it written down that they turned out on time and stood in the wet for five hours without one of them being told what for", Reg: rWry, Form: fComplaint, Topic: "paper"},
	{Text: "The bells were rung once at about noon by a boy who had been told to ring them if anything happened, and were then not rung again, and the boy has spent the whole morning explaining himself to three different people", Reg: rPlain, Topic: "noise"},
	{Text: "A clerk read out the terms from parchment in a voice nobody past the second rank could hear, and read them again louder when the militia complained, and by the third reading their whole column had turned round and started up the road", Reg: rWry, Form: fNotice, Topic: "paper"},
}

// encStandoffIndustrial — depots, sidings, foremen, telegrams, triplicate.
var encStandoffIndustrial = []skel{
	{Text: "The depot cat stayed indoors", Reg: rPlain, Topic: "animal"},
	{Text: "Payroll was late again", Reg: rPlain, Form: fComplaint, Topic: "money"},
	{Text: "Two lorries idled all afternoon", Reg: rPlain, Form: fLedger, Topic: "machine"},
	{Text: "Nobody rang the telephone", Reg: rPlain, Topic: "message"},
	{Text: "The wire was cut", Reg: rPlain, Topic: "message"},
	{Text: "A freight train went through between the two lines, slowly, twice", Reg: rPlain, Topic: "machine"},
	{Text: "The foreman walked over, shook a hand, and walked straight back", Reg: rPlain, Topic: "authority"},
	{Text: "Both work parties stood about on the platform for most of the morning", Reg: rPlain, Topic: "people"},
	{Text: "A report went in triplicate and was mislaid by the afternoon", Reg: rWry, Form: fNotice, Topic: "paper"},
	{Text: "The men on both sides smoked in the yard about ten feet apart", Reg: rPlain, Topic: "people"},
	{Text: "The night shift watched the whole thing from the annexe roof", Reg: rPlain, Topic: "building"},
	{Text: "Somebody telegraphed the entire affair in about nine words", Reg: rWry, Form: fNotice, Topic: "message"},
	{Text: "A version fit for the shareholders is already being drafted upstairs", Reg: rWry, Form: fNotice, Topic: "paper"},
	{Text: "The rail line between the two camps stayed open the whole week", Reg: rPlain, Topic: "border"},
	{Text: "The works whistle went and both sides jumped for no reason", Reg: rJoke, Topic: "noise"},
	{Text: "Their foreman and ours had worked the same siding years back", Reg: rPlain, Topic: "name"},
	{Text: "The night shift want to know who is paying for the extra hours", Reg: rWry, Form: fComplaint, Topic: "money"},
	{Text: "A crate of tools changed hands at the siding and was carried straight back over within the hour by a man who had been told to stop being helpful", Reg: rWry, Topic: "trade"},
	{Text: "The freight was left standing on the siding for four hours and guarded by men from both sides who did not speak to one another once", Reg: rPlain, Topic: "haul"},
	{Text: "Nobody in the depot office will put a name to who moved first, and the telegram that went out at six says only that a meeting took place", Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "Two columns halted a street apart in the rain and neither advanced, and the trams went between them on the hour as though nothing at all were happening", Reg: rWry, Topic: "machine"},
	{Text: "A photograph was taken from the annexe roof and came out badly blurred, and the men in the front of it have all identified themselves anyway", Reg: rJoke, Topic: "town"},
	{Text: "The warehouse doors were padlocked at eleven in the morning and the key went into the foreman's coat pocket, and he walked out to the line and stood there until dark, and the doors were still locked when the night shift came on", Reg: rPlain, Topic: "building"},
	{Text: "The wages office has worked out what four hours of two hundred men standing about in a wet yard comes to, and has put the figure on the foreman's desk, and the foreman has put a mug on top of it", Reg: rJoke, Form: fLedger, Topic: "money"},
}

// encStandoffDigital — feeds, uplinks, drones, analysts, the open channel.
var encStandoffDigital = []skel{
	{Text: "The uplink stayed open", Reg: rPlain, Topic: "machine"},
	{Text: "Two drones, one hour", Reg: rPlain, Form: fLedger, Topic: "machine"},
	{Text: "Nobody closed the channel", Reg: rPlain, Topic: "message"},
	{Text: "The feed shows nothing", Reg: rPlain, Topic: "machine"},
	{Text: "An analyst called it close", Reg: rWry, Topic: "people"},
	{Text: "Their drone and ours held the same height for eleven minutes", Reg: rPlain, Topic: "machine"},
	{Text: "Somebody opened a channel, breathed into it, and closed it again", Reg: rWry, Topic: "noise"},
	{Text: "The after-action briefing has one line in it and a lot of white space", Reg: rWry, Form: fNotice, Topic: "paper"},
	{Text: "The network flagged the whole thing as an engagement, which is generous", Reg: rJoke, Form: fNotice, Topic: "paper"},
	{Text: "Two vehicles idled facing each other for most of the morning", Reg: rPlain, Topic: "machine"},
	{Text: "An analyst has been asked to account for the silence, twice", Reg: rPlain, Form: fComplaint, Topic: "paper"},
	{Text: "Both sets of comms ended up on one channel by accident", Reg: rWry, Topic: "message"},
	{Text: "Wind, boots, and forty minutes of nothing on the uplink", Reg: rPlain, Topic: "noise"},
	{Text: "Somebody streamed the whole afternoon and got very few viewers", Reg: rJoke, Topic: "rumour"},
	{Text: "The road between the two positions stayed empty all night", Reg: rPlain, Topic: "border"},
	{Text: "One analyst noticed that both sides were carrying the same rifles", Reg: rPlain, Topic: "weapon"},
	{Text: "The channel logs are twenty minutes of somebody breathing and four seconds of a man saying a name that has been played back so often the audio is going", Reg: rWry, Form: fOverheard, Topic: "message"},
	{Text: "The feed cut for four seconds in the second hour and came back to exactly the same picture, which several people have found more unsettling than the rest of it", Reg: rWry, Topic: "machine"},
	{Text: "Nobody on the network wants to be the one who logged it, so it has been logged three times with three different timestamps by three different people", Reg: rWry, Form: fNotice, Topic: "paper"},
	{Text: "The drone footage has been reviewed frame by frame by people who were not there, and shows two groups of people standing still, and one man scratching his neck", Reg: rWry, Topic: "rumour"},
	{Text: "Somebody's phone rang in the middle of it, loudly, on both sides of the ground, and the man it belonged to has not been allowed to forget any part of that", Reg: rJoke, Topic: "noise"},
	{Text: "Two of the ground team have put in about the state of the boots issued to them, which is now more paperwork than the afternoon itself generated", Reg: rJoke, Form: fComplaint, Topic: "kit"},
	{Text: "A drone was flown low over their position and flown back again, and the pilot who did it has been spoken to by two people who outrank her and by one who does not, and she has said the same thing to all three", Reg: rWry, Topic: "authority"},
	{Text: "The uplink was open for the whole four hours and carries, in total, one sentence from the man on the left flank about the state of his boots, and the rest is wind", Reg: rWry, Form: fOverheard, Topic: "kit"},
}

// encStandoffCosmic — hulls, bays, transponders, the airlock, the relay.
var encStandoffCosmic = []skel{
	{Text: "Two hulls, no movement", Reg: rPlain, Form: fLedger, Topic: "machine"},
	{Text: "The relay logged it", Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "Their reactor ran hot", Reg: rPlain, Topic: "machine"},
	{Text: "The airlock never cycled", Reg: rPlain, Topic: "machine"},
	{Text: "Four hundred metres, exactly", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "Their transponder answered ours and then went quiet for an hour", Reg: rPlain, Topic: "message"},
	{Text: "Both ships held orbit and neither turned to face the other", Reg: rPlain, Topic: "machine"},
	{Text: "A freighter went between them and the whole moment was over", Reg: rWry, Topic: "machine"},
	{Text: "The manifest they sent across was a list of nothing much", Reg: rWry, Form: fLedger, Topic: "paper"},
	{Text: "A hand went up in their forward bay, at something none of ours could see", Reg: rWry, Topic: "people"},
	{Text: "The dock crew were stood down and told to wait longer", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Two crews looked at each other through the same stretch of vacuum", Reg: rPlain, Topic: "people"},
	{Text: "Their hull markings were read out and copied down wrong", Reg: rPlain, Form: fLedger, Topic: "paper"},
	{Text: "A shuttle went halfway across and turned back on its own", Reg: rPlain, Topic: "machine"},
	{Text: "The bay doors opened and closed with nothing leaving either ship", Reg: rPlain, Topic: "machine"},
	{Text: "Orbital watch has entered it as a routine crossing", Reg: rWry, Form: fNotice, Topic: "border"},
	{Text: "The dock crew want to know why they were held over for six hours", Reg: rWry, Form: fComplaint, Topic: "authority"},
	{Text: "A dead crewman was sent across in a sealed bag and their side took him in without a word on the relay, and the bag came back empty an hour later", Reg: rPlain, Topic: "casualty"},
	{Text: "Their pilot matched our vector to the metre and held it there for most of the watch, which the officer at the helm has described as the worst three hours of her career", Reg: rWry, Topic: "machine"},
	{Text: "The reactor alarm went off in the second hour and everybody on this side assumed the worst, and it turned out to be a sensor that has been going for a fortnight", Reg: rWry, Topic: "machine"},
	{Text: "Nothing crossed the gap in four hours except one very poor joke on an open frequency, which their side did not laugh at and ours has been repeating since", Reg: rJoke, Form: fOverheard, Topic: "noise"},
	{Text: "A boarding party got as far as the tube, stood in it fully suited for twenty minutes, and came back through the airlock without anybody saying why", Reg: rPlain, Topic: "people"},
	{Text: "The transponder codes were exchanged, logged, checked against the list, found to be twelve years out of date on their side, and passed anyway by an officer who has since written four hundred words about why", Reg: rWry, Form: fNotice, Topic: "paper"},
	{Text: "Somebody in the forward bay scratched a name into the frame beside the port during the second hour, and it has been found, and nobody has painted over it, and there is an argument going on about whether it should stay", Reg: rPlain, Topic: "name"},
}
