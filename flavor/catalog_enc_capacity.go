package flavor

// EncounterAtCapacity — somebody arrived with an offer and the court could not
// take it.
//
// The log line above this one has already said so, so none of the sentences here
// say it again, and none of them is about fullness. They are about the people and
// the objects and the small administrative disasters that fullness produces: who
// signed for the second load, what has gone soft behind a chest, which visitor has
// quietly acquired a bed and a job, how much bread is going out, and the standing
// instruction that everybody knows and nobody wants to be the one to read aloud.
//
// The voice reference is catalog_exp_success.go — a quartermaster's book rather
// than a book of aphorisms. Lengths swing on purpose, from four words to forty,
// because the uniform rhythm of the two rejected versions of this file was the
// tell, independently of how good any single line was.

// --- slot banks --------------------------------------------------------------
//
// Two tight categories: the thing that arrived, and the place it was quietly
// put. Both are drawn on from the ungated pool, so both are free of era-coded
// nouns — a painted chest and a bad corner are as true of a longhouse as of an
// orbital station.

var encCapacityBanks = map[string][]string{
	// What turned up that nobody has anywhere to put.
	"enc_cap_gift": {
		"a crate of dried fruit", "a very large mirror", "two more caged birds",
		"a painted cabinet", "a bolt of heavy cloth", "a jar of something sweet",
		"another set of matched cups", "a life-size statue of a stranger",
		"a box of small carved figures", "three crates of wine",
		"a bundle of dried flowers", "a heavy glass bowl",
		"a bag of seed nobody recognises", "a cage of white rabbits",
		"a set of ceremonial knives", "an enormous woven hanging",
		"a case of foreign money", "a portrait of somebody important",
	},
	// Where it ended up. All noun phrases, all of them somewhere out of the way.
	"enc_cap_place": {
		"the back corner", "the room nobody uses", "the far end of the hall",
		"a cold room at the back", "the corridor outside", "the stairwell",
		"the space behind the door", "a cupboard that already sticks",
		"the corner where the light is bad", "the empty room beside the kitchen",
		"a locked room upstairs", "the alcove by the entrance",
		"the gap between two cabinets", "the side room with the low ceiling",
		"the landing halfway up the stairs", "the passage behind the curtain",
		"a room that used to be a store", "the shadow under the stairs",
	},
}

// --- the sentences -----------------------------------------------------------

// encCapacityTemplates is the EncounterAtCapacity skeleton set: one ungated pool
// that lands in any age, plus one pool per era bucket for the imagery that does
// not travel.
func encCapacityTemplates() []tmpl {
	out := pool("enc_cap_any", erasAny, encCapacityAny)
	out = append(out, pool("enc_cap_grounded", erasGrounded, encCapacityGrounded)...)
	out = append(out, pool("enc_cap_late", erasLate, encCapacityLate)...)
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
	// --- short. A fact, and then the sentence stops. ---
	{Text: "Nothing can get down the corridor", Reg: rPlain, Topic: "building"},
	{Text: "Seven chairs, fourteen visitors", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "Somebody signed for all of it", Reg: rPlain, Form: fNotice, Topic: "paper"},
	{Text: "The bird has stopped singing", Reg: rPlain, Topic: "animal"},
	{Text: "Nobody will open the tall crate", Reg: rPlain, Topic: "haul"},
	{Text: "Three arrivals before noon", Reg: rPlain, Form: fLedger, Topic: "time"},
	{Text: "The far room has been claimed", Reg: rPlain, Topic: "sleep"},
	{Text: "One name was written twice", Reg: rPlain, Topic: "name"},
	{Text: "Everything is stacked waist-high", Reg: rPlain, Topic: "haul"},
	{Text: "The interpreter has lost her voice", Reg: rPlain, Topic: "message"},
	{Text: "There is dust on ~ already", Slot: "enc_cap_gift", Reg: rPlain, Topic: "kit"},

	// --- mid. The working length: one thing, with enough detail to place it. ---
	{Text: "Gifts have been stacked along the corridor wall for over a week", Reg: rPlain, Topic: "haul"},
	{Text: "The man who keeps the list has run out of page", Reg: rPlain, Form: fComplaint, Topic: "paper"},
	{Text: "An envoy has started correcting the spelling on the notices", Reg: rPlain, Topic: "people"},
	{Text: "The animal that came this morning has eaten nothing it was offered", Reg: rPlain, Topic: "animal"},
	{Text: "Two delegations who dislike each other have been put in one room", Reg: rPlain, Topic: "argument"},
	{Text: "The cost of feeding everybody has been mentioned at the table", Reg: rPlain, Topic: "money"},
	{Text: "Something in the far corner has begun to smell of fish", Reg: rPlain, Topic: "smell"},
	{Text: "There is a standing instruction about this and nobody wants to break it first", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "One of the visitors has moved his things into an upstairs room", Reg: rPlain, Topic: "stranger"},
	{Text: "The woman on the door has started writing names on her hand", Reg: rPlain, Topic: "name"},
	{Text: "The children have been told to stay out of the far room", Reg: rPlain, Topic: "town"},
	{Text: "Somebody put a cloth over the big mirror because of the light", Reg: rPlain, Topic: "kit"},
	{Text: "Two of the gifts are the same gift, more or less", Reg: rJoke, Topic: "trade"},
	{Text: "The welcome speech has been rewritten for the fourth arrival", Reg: rPlain, Topic: "paper"},
	{Text: "The queue outside has doubled since morning and nobody has counted it", Reg: rPlain, Topic: "count"},
	{Text: "A gift arrived wrapped in cloth worth more than the gift", Reg: rWry, Topic: "trade"},
	{Text: "The rope across the cupboard door was somebody's idea of a solution", Reg: rWry, Topic: "kit"},
	{Text: "The pile by the door has taken ~ as well", Slot: "enc_cap_gift", Reg: rPlain, Topic: "haul"},
	{Text: "The newest thing to arrive has been carried straight to ~", Slot: "enc_cap_place", Reg: rPlain, Topic: "building"},
	{Text: "Nobody could think of anywhere better than ~", Slot: "enc_cap_place", Reg: rWry, Topic: "building"},
	{Text: "It has been decided to leave ~ where it is", Slot: "enc_cap_gift", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Everything that came in this week has gone into ~", Slot: "enc_cap_place", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "Whoever arranged the hall has put ~ behind a curtain", Slot: "enc_cap_gift", Reg: rWry, Topic: "building"},
	{Text: "The children have got at ~ and are being left to it", Slot: "enc_cap_gift", Reg: rWry, Topic: "town"},
	{Text: "Something soft has been left too long in ~", Slot: "enc_cap_place", Reg: rPlain, Topic: "smell"},

	// --- long. Circumstantial, subordinated, the way a chronicle actually runs. ---
	{Text: "The room set aside for guests has a man asleep in it who came with the second lot and has since attached himself to nobody in particular", Reg: rPlain, Topic: "sleep"},
	{Text: "The woman who keeps the list has asked for a second book, been told that one is being looked for, and started writing in the margins of the first", Reg: rPlain, Form: fComplaint, Topic: "paper"},
	{Text: "Nobody can say who signed for the second load, because the paper it was signed on has been used to wrap something, and the something has gone upstairs", Reg: rWry, Topic: "paper"},
	{Text: "For the second time this month somebody has moved ~ without being asked, and the second move was into a worse place than the first", Slot: "enc_cap_gift", Reg: rWry, Topic: "argument"},
	{Text: "The far end of a list somewhere has ~ written on it in a hand nobody recognises, and the item itself has not been seen since it came through the door", Slot: "enc_cap_gift", Reg: rWry, Form: fLedger, Topic: "paper"},

	// --- very long. Digressive, specific, a paragraph that forgot to stop. ---
	{Text: "Two of the waiting parties will not be in the same room, so one has the front and one has the back, and the passage between them has a chair in it with somebody sitting on it", Reg: rPlain, Topic: "argument"},
	{Text: "The third delegation this month came in while the second was still waiting to be seen, and the second has now been waiting long enough to have opinions about the third that it has been sharing with anybody who walks past", Reg: rWry, Topic: "time"},
	{Text: "The interpreter has worked every day since the first party arrived and has been thanked once, and this morning turned a long and formal greeting into a shorter and less formal one that nobody in the room queried", Reg: rWry, Topic: "message"},
	{Text: "One count of what has come in this month made it forty-one separate items, another the next day made it thirty-eight, and both numbers are written up on the same wall in the same hand", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "A woman came in on the fourth day with a small box and a long speech, was heard out, was given water and a chair, and is still in the chair, and the box is still on her knees", Reg: rPlain, Topic: "stranger"},

	// --- kind: mercantile ---
	{Text: "Samples went back unopened", Kinds: []string{"mercantile"}, Reg: rPlain, Form: fLedger, Topic: "trade"},
	{Text: "The traders brought samples and are leaving with the samples", Kinds: []string{"mercantile"}, Reg: rJoke, Topic: "trade"},

	// --- kind: peaceful ---
	{Text: "They brought flowers that will last about four days", Kinds: []string{"peaceful"}, Reg: rPlain, Topic: "kit"},
	{Text: "The visitors have been patient", Kinds: []string{"peaceful"}, Reg: rPlain, Topic: "people"},

	// --- kind: isolationist ---
	{Text: "They came a long way to stand in a doorway", Kinds: []string{"isolationist"}, Reg: rPlain, Topic: "stranger"},
	{Text: "The delegation refused food, refused chairs, refused the offer of a room, and has been standing in the same part of the hall since the middle of the morning", Kinds: []string{"isolationist"}, Reg: rPlain, Topic: "people"},

	// --- kind: aggressive ---
	{Text: "Their guards stand and stay standing", Kinds: []string{"aggressive"}, Reg: rPlain, Topic: "weapon"},
	{Text: "The escort left its weapons by the door after being asked four times", Kinds: []string{"aggressive"}, Reg: rPlain, Topic: "weapon"},

	// --- tone ---
	{Text: "Six matched cups, four sets", Tones: []Tone{Wry}, Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "Nobody has counted the chairs", Tones: []Tone{Wry}, Reg: rPlain, Topic: "count"},
	{Text: "The room has been named after what is in it", Tones: []Tone{Wry}, Reg: rJoke, Topic: "building"},
	{Text: "Three separate peoples have now given the same species of bird", Tones: []Tone{Wry}, Reg: rWry, Topic: "animal"},
	{Text: "Two people have worked out that a gift sent out last year has come back", Tones: []Tone{Wry}, Reg: rJoke, Topic: "trade"},
	{Text: "Somebody has begun keeping a second list of the things on the first list, on the grounds that the first list has stopped being a list and started being a wall", Tones: []Tone{Wry}, Reg: rJoke, Form: fLedger, Topic: "paper"},
	{Text: "A visitor asked where the gifts are kept and was taken to four different rooms by four different people, none of whom took him to the same room", Tones: []Tone{Wry}, Reg: rWry, Topic: "building"},

	// --- the visitor's name, in frames a singular-or-plural name survives ---
	{Text: "Word of it reached {subject} first", Needs: needSubject, Reg: rPlain, Topic: "rumour"},
	{Text: "Chairs were found for {subject}", Needs: needSubject, Reg: rPlain, Topic: "kit"},
	{Text: "Somebody waiting has folded the note from {subject} into a fan", Needs: needSubject, Reg: rWry, Topic: "paper"},
	{Text: "Nobody has explained to {subject} how long the waiting has been running", Needs: needSubject, Reg: rPlain, Topic: "time"},
	{Text: "The seating plan puts {subject} behind a pillar", Needs: needSubject, Reg: rWry, Topic: "argument"},
	{Text: "Everything sent by {subject} went to the same corner", Needs: needSubject, Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "An interpreter was found for {subject} and then sent away again", Needs: needSubject, Reg: rPlain, Topic: "people"},
	{Text: "The visitors' list has {subject} on it in two different hands", Needs: needSubject, Reg: rPlain, Form: fLedger, Topic: "name"},
	{Text: "The bill for feeding {subject} came up at the evening meal and was read out by somebody who had clearly been waiting all day to read it out", Needs: needSubject, Reg: rWry, Form: fComplaint, Topic: "money"},
	{Text: "The wrong line of the wrong list now reads {subject}, and it has been copied twice since by people who had no reason to doubt the first copy", Needs: needSubject, Reg: rPlain, Topic: "paper"},
	{Text: "The room at the back, the room beside it and the passage between the two have all been given over to the men who came in with {subject}, and there are still four of them sleeping in the hall", Needs: needSubject, Reg: rPlain, Topic: "sleep"},
	{Text: "Nobody has worked out what to do with what came from {subject}, so it has been left where it was set down, and people have begun walking round it as though it had always been there", Needs: needSubject, Reg: rWry, Topic: "haul"},
}

// encCapacityGrounded is set dressing that stops at the Modern Age: lamps, servants,
// chests, a council that sits before it is light. True in a manor, wrong on a
// station.
var encCapacityGrounded = []skel{
	{Text: "A basket of soft fruit was found behind a chest this morning", Reg: rPlain, Topic: "food"},
	{Text: "A door servant has learned to say {subject} without pausing in the middle, which took four days, and is the only thing about this week she intends to keep", Needs: needSubject, Reg: rWry, Topic: "name"},
	{Text: "A servant announced {subject} to a room that already had two delegations standing in it", Needs: needSubject, Reg: rPlain, Topic: "noise"},
	{Text: "A servant has counted the same stack four times this week", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "A servant was told to find room for ~", Slot: "enc_cap_gift", Reg: rPlain, Topic: "people"},
	{Text: "A treaty was read out in full to a room holding one servant, two waiting envoys from somewhere else entirely, and a child who had come in to look at the birds, and none of them was asked to leave", Kinds: []string{"peaceful"}, Reg: rWry, Form: fOverheard, Topic: "paper"},
	{Text: "An old woman has been sitting on a chest since first light", Reg: rPlain, Topic: "people"},
	{Text: "Four people have asked where the new chests go and none has been answered", Reg: rPlain, Form: fComplaint, Topic: "building"},
	{Text: "One of the sealed jars has been leaking down the inside of a chest", Reg: rPlain, Topic: "smell"},
	{Text: "One of the visitors has a bed, a chest, a place at the table and a job carrying water, and nobody can now remember which delegation brought him or whether he came with one at all", Reg: rWry, Topic: "stranger"},
	{Text: "Somebody has been sleeping in the far room since the third day, and the household believes he came with a delegation, and the delegations believe he lives here", Reg: rWry, Topic: "sleep"},
	{Text: "Somebody worked out that the visitors have eaten more in nine days than the household gets through in a month, and the figure has been repeated at every meal since", Reg: rWry, Form: fComplaint, Topic: "food"},
	{Text: "The best room in the house has been given to a crate, and the man who was moved out of it to make space has been sleeping on a bench in the passage and telling people about it", Tones: []Tone{Wry}, Reg: rWry, Topic: "sleep"},
	{Text: "The chest marked {subject} stays shut", Needs: needSubject, Reg: rPlain, Topic: "haul"},
	{Text: "The chests from the first day were pushed against the wall to make room for the second lot, and the second lot went in front of them", Reg: rPlain, Topic: "haul"},
	{Text: "The door servant has become the most consulted person in the house", Tones: []Tone{Wry}, Reg: rJoke, Topic: "authority"},
	{Text: "The household has been feeding between nine and fourteen extra people at every meal for nine days, and the woman who does the buying has stopped asking how long this goes on and has started buying for twenty", Reg: rPlain, Form: fComplaint, Topic: "food"},
	{Text: "The household now owns more ceremonial bowls than plates", Tones: []Tone{Wry}, Reg: rJoke, Topic: "kit"},
	{Text: "The interpreter was asked to attend at first light, waited through the whole morning without being called, and has been asked to attend again at first light tomorrow", Reg: rWry, Topic: "people"},
	{Text: "The merchants have set up outside the door to wait, have been there two days, and have begun selling to the household staff at a discount they describe as friendly", Kinds: []string{"mercantile"}, Reg: rWry, Topic: "money"},
	{Text: "The tall visitor has read every notice on the wall by the door, in order, and has begun asking the servants about the ones he does not understand", Reg: rPlain, Topic: "people"},
	{Text: "There is a rule about how many of these the household may take, everybody knows the rule, and the household went past it on the second day", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "Two more came at dusk", Reg: rPlain, Topic: "stranger"},
	{Text: "Two servants carried a crate from {subject} up the stairs, were told it should have gone down instead, and carried it down again without saying anything", Needs: needSubject, Reg: rPlain, Topic: "people"},
	{Text: "Two servants spent the whole morning shifting ~ from one side of a room to the other, and then shifting it back after somebody senior walked through", Slot: "enc_cap_gift", Reg: rWry, Topic: "haul"},
	{Text: "A jar that came in sealed has been opened, sniffed by four people in turn, and sealed again with a cloth and a length of string", Reg: rPlain, Topic: "smell"},

	// --- re-tagged out of the shared pool by the third late-age review: the
	// kitchen, the back door, the walk home. True up to the Modern Age. ---
	{Text: "The smell has reached the stairs", Reg: rPlain, Topic: "smell"},
	{Text: "The bread is running out", Reg: rPlain, Form: fComplaint, Topic: "food"},
	{Text: "Water is being carried upstairs", Reg: rPlain, Topic: "people"},
	{Text: "The big grey animal has been given a name by the kitchen staff", Reg: rJoke, Topic: "name"},
	{Text: "The kitchen has stopped asking how many are eating and started cooking for forty", Reg: rPlain, Form: fComplaint, Topic: "food"},
	{Text: "A man arrived at midday with a folded letter and an animal on a rope, and he has been offered water three times by three different people", Reg: rPlain, Topic: "stranger"},
	{Text: "Nobody knows what the animal eats, so it has been offered bread, fruit, boiled grain and a bowl of water, and it has taken the water", Reg: rPlain, Topic: "animal"},
	{Text: "The bread and the meat have been going out at twice the usual rate since the first party arrived, and the woman who buys both has started writing the numbers on the wall", Reg: rPlain, Form: fLedger, Topic: "money"},
	{Text: "An envoy who arrived before the cold weather now dresses like everybody else here, eats with the house, and was heard this morning complaining about the price of fish", Reg: rWry, Topic: "stranger"},
	{Text: "A boy was sent out to tell {subject} that the hall is being cleaned", Needs: needSubject, Reg: rWry, Form: fOverheard, Topic: "town"},
	{Text: "The gifts from {subject} came in three loads over two days, and the third load is still outside under a cloth because the second one has the doorway", Needs: needSubject, Reg: rPlain, Topic: "haul"},
}

// encCapacityLate is the digital and cosmic voice: the field notes have become
// logs, and the logs have started to notice how long everything takes.
var encCapacityLate = []skel{
	{Text: "The gift registry has run out of storage", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "One delegation has brought a gift that has not finished arriving", Reg: rJoke, Topic: "stranger"},
	{Text: "Protocol requires every gift to be acknowledged in person, and at current rates the acknowledgements will be finished some time after the last person able to give them has died", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "The guest quarters are full of envoys waiting to be told they can go home", Reg: rPlain, Topic: "people"},
	{Text: "An envoy asked what we needed and was shown the list of things we have been given", Reg: rWry, Topic: "trade"},
	{Text: "Storage is full", Reg: rPlain, Form: fNotice, Topic: "building"},
	{Text: "Every guest suite is taken", Reg: rPlain, Topic: "people"},
	{Text: "Further gifts cannot be accepted", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "The overflow room has overflowed", Reg: rPlain, Topic: "building"},
	{Text: "The welcome screen has a waiting list", Reg: rWry, Topic: "machine"},
	{Text: "Nine delegations, four interpreters", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The protocol officer has resigned again", Reg: rJoke, Topic: "people"},
	{Text: "Catering has stopped taking orders", Reg: rPlain, Topic: "food"},
	{Text: "The security scanner needs a rest", Reg: rWry, Topic: "kit"},
	{Text: "One envoy has been here a year", Reg: rPlain, Topic: "stranger"},
	{Text: "Every gift has to be logged and thanked for, and the thanking is where it has stopped", Reg: rWry, Topic: "authority"},
	{Text: "The embassy wing has been extended into the car park with temporary buildings", Reg: rPlain, Topic: "building"},
	{Text: "One delegation has offered to take some of the other delegations' gifts off our hands, for a fee", Reg: rWry, Topic: "trade"},
	{Text: "The interpreters are working in shifts and one of them has started dreaming in the visitors' language", Reg: rPlain, Topic: "people"},
	{Text: "An envoy who arrived to open talks has been here long enough to be invited to the office party", Reg: rJoke, Topic: "stranger"},
	{Text: "The scanner at the entrance has found the same undeclared item in four different gifts", Reg: rPlain, Topic: "machine"},
	{Text: "Thank-you letters are being drafted from a template, and the template has started to show", Reg: rWry, Topic: "paper"},
	{Text: "A robot cleaner has been going round and round a crate that arrived this morning and has not been moved", Reg: rPlain, Topic: "kit"},
	{Text: "The diplomatic calendar is now booked beyond the current government's term of office", Reg: rWry, Topic: "authority"},
	{Text: "Two envoys from rival delegations have been seated together at lunch three days running and have become friends", Reg: rPlain, Topic: "people"},
	{Text: "The cleaning staff would like to know which of the gifts are rubbish", Reg: rWry, Form: fComplaint, Topic: "people"},
	{Text: "A delegation arrived that no office had invited, and it was easier to accept their gift than to find out who they were", Reg: rPlain, Topic: "stranger"},
	{Text: "The customs duty on the gifts has been calculated and exceeds the budget of the office receiving them", Reg: rWry, Topic: "money"},
	{Text: "Gifts are being stored in the lift, which has been taken out of service for the purpose", Reg: rPlain, Topic: "building"},
	{Text: "Every delegation's name has been spelled wrong at least once on a lanyard", Reg: rWry, Topic: "name"},
	{Text: "Visiting hours for foreign envoys have been introduced and are being ignored", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "Every delegation brought food, and most of it needs refrigerating, and the fridges were full on Monday", Reg: rPlain, Topic: "food"},
	{Text: "The facial recognition at the entrance has started greeting the regular envoys by name", Reg: rJoke, Topic: "machine"},
	{Text: "An envoy has proposed that his delegation's gift be counted as an advance against next year's gift, and the finance office has spent a week trying to work out whether that is allowed", Reg: rWry, Topic: "trade"},
	{Text: "The head of protocol has been receiving delegations since seven and has now started receiving them standing up, to keep the meetings short", Reg: rPlain, Topic: "people"},
	{Text: "The gift register has grown so large that it now has its own register, listing which volume of the gift register each gift appears in", Reg: rWry, Topic: "paper"},
	{Text: "A second reception area has been opened in the old training suite, with a sign on the door and a coffee machine that works only for the first hour", Reg: rPlain, Topic: "building"},
	{Text: "One ambassador has been waiting so long for an audience that she has been appointed to a committee, and the committee has started meeting in the waiting room", Reg: rJoke, Topic: "stranger"},
	{Text: "The automated reply to foreign correspondence has been rewritten to sound less automated, and now three delegations have written back to the automated reply by name", Reg: rWry, Topic: "machine"},
	{Text: "The gift from one delegation is a machine for sorting gifts, and it has been unpacked and switched on, and it is the only thing in the building that looks happy", Reg: rPlain, Topic: "kit"},
	{Text: "Every new arrival is being told that the court is honoured and cannot receive them yet, and one delegation wants that in writing so that they can frame it", Reg: rWry, Topic: "authority"},
	{Text: "A child from one delegation has been taking the other delegations' children on tours of the building, and she now knows it better than half the staff", Reg: rPlain, Topic: "people"},
	{Text: "The insurance on the gifts now costs more per month than the building they are kept in", Reg: rWry, Topic: "money"},
	{Text: "The office that handles incoming delegations has asked for more staff, and the request has gone into the queue behind the delegations, and the delegations have begun asking whether that office has been told they are here", Reg: rWry, Topic: "authority"},
	{Text: "Every meeting room in the building has been given over to a delegation, and the staff who used to meet in them now meet standing in the corridor by the lifts, with the visiting envoys walking past and sometimes stopping to listen", Reg: rPlain, Topic: "building"},
	{Text: "The newest envoy arrived with a long speech of goodwill and a letter of introduction from a government that no longer exists, and he is being received with full honours and given the same slow answer as everyone else", Reg: rWry, Topic: "stranger"},
	{Text: "Two delegations have started trading their gifts to each other in the lobby while they wait, and one of them now holds the gift the other brought for us, and has offered to give it to us again", Reg: rJoke, Topic: "trade"},
	{Text: "The lobby sofas are all taken", Reg: rPlain, Topic: "people"},
	{Text: "The visitor app has crashed", Reg: rWry, Topic: "machine"},
	{Text: "The spare offices on the top floor have been furnished with borrowed chairs for envoys who arrived without appointments", Reg: rPlain, Topic: "building"},
	{Text: "The gift shop has started selling copies of the gifts", Reg: rWry, Topic: "trade"},
	{Text: "One delegation has set up its own reception desk inside ours and is now receiving the delegations that arrive after it", Reg: rPlain, Topic: "stranger"},

	// --- added by the third late-age review: offices, stations, the long view ---
	{Text: "Envoys are now waiting in the old mail room, among the parcels", Reg: rPlain, Topic: "building"},
	{Text: "Three envoys share one charger", Reg: rWry, Topic: "kit"},
	{Text: "Visitor passes are being written by hand now", Reg: rPlain, Topic: "paper"},
	{Text: "Every chair in the atrium has a delegation's coat on it", Reg: rPlain, Topic: "people"},
	{Text: "An envoy has been resending his greeting every hour on the hour", Reg: rPlain, Topic: "message"},
	{Text: "Every meeting room is booked until Friday, most of them by envoys who wanted somewhere to sit down", Reg: rWry, Topic: "time"},
	{Text: "The delegations have drawn up a rota for the one good sofa", Reg: rWry, Topic: "sleep"},
	{Text: "A crate marked fragile has served as a desk for three different envoys this week", Reg: rPlain, Topic: "haul"},
	{Text: "Protocol has run out of flags and is printing them on paper", Reg: rPlain, Topic: "authority"},
	{Text: "One delegation's gift is a live plant that needs watering, and the only person who waters it is one of their own envoys", Reg: rPlain, Topic: "stranger"},
	{Text: "Two envoys who arrived in the first week are now training the envoys who arrived this week", Reg: rWry, Topic: "people"},
	{Text: "The reception desk now hands out numbered tickets, and the envoys take them without complaint, and the ticket roll has been changed three times since Monday by a facilities man who keeps asking what is going on upstairs", Reg: rWry, Topic: "count"},
	{Text: "The cleaners have been given a list of the gifts they may throw away, and the list is blank", Reg: rWry, Topic: "authority"},
	{Text: "The envoys in the lobby are betting on who gets seen first", Reg: rPlain, Form: fOverheard, Topic: "rumour"},
	{Text: "Facilities would like the gifts moved off the emergency exits", Reg: rPlain, Form: fComplaint, Topic: "building"},
	{Text: "Envoys are asked not to charge their equipment from the emergency lighting", Reg: rWry, Form: fNotice, Topic: "machine"},
	{Text: "An envoy fell asleep in the lift and rode it between floors for most of the afternoon", Reg: rWry, Topic: "sleep"},
	{Text: "The trade delegation has started invoicing us for the waiting", Kinds: []string{"mercantile"}, Reg: rJoke, Topic: "money"},
	{Text: "They say they are happy to wait, and have brought books", Kinds: []string{"peaceful"}, Reg: rPlain, Topic: "people"},
	{Text: "Their security detail has taken over a whole corridor and put up signs of its own", Kinds: []string{"aggressive"}, Reg: rPlain, Topic: "border"},
	{Text: "They have refused the guest rooms and pitched a small tent in the atrium", Kinds: []string{"isolationist"}, Reg: rPlain, Topic: "stranger"},
	{Text: "The gift from one delegation is a book about patience", Tones: []Tone{Wry}, Reg: rJoke, Topic: "kit"},
}

// encCapacityAncient — the elders, the store-pit, hides drying for guests who
// show no sign of going home.
var encCapacityAncient = []skel{
	{Text: "The store-pit will not shut", Reg: rPlain, Topic: "building"},
	{Text: "Hides on every drying frame", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "Another herd came at dawn", Reg: rPlain, Topic: "animal"},
	{Text: "The elders have stopped counting", Reg: rPlain, Topic: "authority"},
	{Text: "Drums for the first three only", Reg: rWry, Topic: "noise"},
	{Text: "Two spears were planted outside a hut that has guests in it already", Reg: rPlain, Topic: "argument"},
	{Text: "The hides meant for winter are under sleeping strangers", Reg: rPlain, Topic: "sleep"},
	{Text: "A watch-fire was kept burning for people who came a day early", Reg: rPlain, Topic: "border"},
	{Text: "The spring below the camp is being drunk dry by visitors", Reg: rPlain, Topic: "ground"},
	{Text: "Somebody's hut now holds four carved things and one sleeping man", Reg: rWry, Form: fLedger, Topic: "building"},
	{Text: "The elders argued about who sits nearest the fire until it went out", Reg: rWry, Topic: "argument"},
	{Text: "An arrow was given as a token and has already been lost", Reg: rPlain, Topic: "kit"},
	{Text: "Three parties crossed the ford in two days and none has gone home", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "The old woman minding the visitors has asked to be given fewer visitors", Reg: rJoke, Form: fComplaint, Topic: "people"},
	{Text: "Smoke on the far ridge means another lot are walking in", Reg: rPlain, Topic: "border"},
	{Text: "A hide painted with strangers' marks hangs where nobody looks at it", Reg: rPlain, Topic: "map"},
	{Text: "The herd that came as a gift has eaten the grass down to the mud, and the boy who was told to mind it has been told to mind three more", Reg: rPlain, Topic: "animal"},
	{Text: "The elders have run out of ways to say wait, and the woman who does the saying has been doing it in the doorway since the light came up", Reg: rWry, Form: fOverheard, Topic: "authority"},
	{Text: "Spears from four different peoples are stacked against one wall, and every one of them was handed over with a speech that had to be listened to", Reg: rPlain, Topic: "weapon"},
	{Text: "The store-pit holds a great deal that nobody intends to eat, including two sacks of seed that came from people who do not grow anything here", Reg: rWry, Form: fLedger, Topic: "food"},
	{Text: "Guests have been sleeping under the hides put by for the cold months, and the woman who put them by has said her piece about it to three separate people", Reg: rPlain, Form: fComplaint, Topic: "sleep"},
	{Text: "The drums were sounded for the first three arrivals and for none of the six after that, which the sixth lot noticed and mentioned before they had sat down", Reg: rWry, Topic: "noise"},
	{Text: "One of the gifts was traded away the morning after it arrived, for a spear and a hide and a promise of two more hides, and the man who did it has answered four separate people differently about where it went", Reg: rWry, Topic: "trade"},
	{Text: "The people from the valley brought stone, stayed to watch what was done with the stone, were still watching when the next lot came up the road with more stone, and have now been given a hut of their own", Reg: rPlain, Topic: "stranger"},
}

// encCapacityFeudal — the steward, the gate, carts in the yard, heralds arguing
// about who speaks first.
var encCapacityFeudal = []skel{
	{Text: "The gate has been left open", Reg: rPlain, Topic: "border"},
	{Text: "Four carts, nobody unloading", Reg: rWry, Form: fLedger, Topic: "haul"},
	{Text: "The undercroft is full of presents", Reg: rPlain, Topic: "building"},
	{Text: "Wax on the wrong parchment", Reg: rPlain, Topic: "paper"},
	{Text: "Bells for the third delegation", Reg: rWry, Topic: "noise"},
	{Text: "The steward has stopped going down to meet new arrivals", Reg: rPlain, Topic: "authority"},
	{Text: "Two heralds arrived together and argued about which of them speaks first", Reg: rWry, Topic: "argument"},
	{Text: "The clerks have written the same welcome six times this month", Reg: rPlain, Form: fComplaint, Topic: "paper"},
	{Text: "A muster was called off so the yard could be cleared of wagons", Reg: rPlain, Topic: "town"},
	{Text: "The quartermaster has refused to sign for any more of it", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Riders keep arriving and the pickets have stopped bothering to count", Reg: rPlain, Topic: "border"},
	{Text: "The village has begun charging the waiting parties for bread", Reg: rWry, Topic: "money"},
	{Text: "An envoy has been given the steward's own room to sleep in", Reg: rPlain, Topic: "sleep"},
	{Text: "Supper was served twice because neither party would sit at one table", Reg: rPlain, Topic: "food"},
	{Text: "A horn was blown for a delegation that had already gone home", Reg: rJoke, Topic: "noise"},
	{Text: "Somebody's cart blocked the gate for most of the afternoon", Reg: rPlain, Topic: "machine"},
	{Text: "The clerk who keeps the visitors' list has asked in writing to be moved to other work, and the request went to the steward, who has it", Reg: rWry, Form: fComplaint, Topic: "paper"},
	{Text: "Sacks of foreign grain have been stacked against the storehouse wall under a cloth, and the cloth has been taken twice for other uses and put back once", Reg: rPlain, Topic: "haul"},
	{Text: "Parchment has run short on account of all the polite refusals, and the clerks are writing the newest ones on the backs of old muster lists", Reg: rWry, Topic: "paper"},
	{Text: "The militia turned out to carry chests up two flights of stairs, which was rather less than any of them had in mind when they answered the muster", Reg: rWry, Form: fComplaint, Topic: "people"},
	{Text: "Wagons have been coming up to the gate since before first light, and the gate has been left standing open on the grounds that shutting it and opening it is more work", Reg: rPlain, Topic: "border"},
	{Text: "Somebody regifted a cup at midsummer and it came back to the undercroft within the month, wrapped in better cloth and with a longer letter attached", Reg: rWry, Topic: "trade"},
	{Text: "A boy has been paid a coin a day to stand at the gate and tell arriving parties that the hall is being cleaned, which is the fourth day he has said it", Reg: rJoke, Form: fOverheard, Topic: "town"},
	{Text: "The banner hung for the first arrival was taken down for the second, put up again for the third because somebody thought it rude to leave it down, and has been up ever since through two more that nobody bothered to greet", Reg: rWry, Topic: "religion"},
}

// encCapacityIndustrial — the depot, the annexe, freight on the platform, and a
// telegram that arrived after the people it announced.
var encCapacityIndustrial = []skel{
	{Text: "The annexe roof is leaking", Reg: rPlain, Topic: "building"},
	{Text: "Freight on the platform, unclaimed", Reg: rPlain, Form: fLedger, Topic: "haul"},
	{Text: "Two more crates, no forms", Reg: rWry, Form: fLedger, Topic: "paper"},
	{Text: "The telephone has been unplugged", Reg: rJoke, Topic: "machine"},
	{Text: "Nobody signed the delivery book", Reg: rPlain, Topic: "count"},
	{Text: "The telegram announcing the delegation arrived a day after the delegation", Reg: rWry, Topic: "message"},
	{Text: "A whole siding has been given over to somebody's ceremonial engine", Reg: rPlain, Topic: "machine"},
	{Text: "The foreman was asked to find space and has declined in writing", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "Two men are on the payroll now purely to move boxes about", Reg: rPlain, Topic: "money"},
	{Text: "The depot office has run out of the yellow visitor forms", Reg: rPlain, Form: fComplaint, Topic: "paper"},
	{Text: "Shareholders have been told the reception went off extremely well", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "A wire came in asking whether the previous wire had come in", Reg: rJoke, Topic: "message"},
	{Text: "The lorries arrived at six and are still standing in the yard", Reg: rPlain, Topic: "machine"},
	{Text: "Three delegations have been booked into the same Thursday afternoon", Reg: rPlain, Form: fNotice, Topic: "time"},
	{Text: "A brass band was hired for the fourth arrival and stood down after one tune", Reg: rWry, Topic: "noise"},
	{Text: "Smoke from the visitors' cigars has filled the upper corridor", Reg: rPlain, Topic: "smell"},
	{Text: "Everything has been entered in triplicate, filed in three separate places, and the man who knows which three has been off sick since the second delegation came in", Reg: rWry, Form: fLedger, Topic: "paper"},
	{Text: "The warehouse manager has stopped answering the telephone, on the grounds that every call this week has been about somewhere to put something", Reg: rWry, Form: fComplaint, Topic: "building"},
	{Text: "Two crates of gifts have sat on the platform all week under a tarpaulin, and the station master has begun charging demurrage on them to somebody", Reg: rPlain, Topic: "money"},
	{Text: "The freight elevator has been reserved for gifts until Friday, which means the fourth floor is carrying everything up by hand and has said so loudly", Reg: rWry, Form: fComplaint, Topic: "machine"},
	{Text: "A porter has spent two days carrying the same chest between the annexe and the depot office, and nobody has told him to stop because nobody has noticed", Reg: rPlain, Topic: "people"},
	{Text: "The visitors' train was met by one man with an umbrella and a sheet of paper, and the sheet of paper had the wrong name on it in a very good hand", Reg: rWry, Topic: "name"},
	{Text: "The company photographer has now been booked for a fifth arrival, and has photographed four sets of visitors standing in front of the same crates, which have been in the annexe long enough to appear in all four", Reg: rWry, Topic: "town"},
	{Text: "A leak in the annexe roof has reached the stacked crates, and the man who reported it in March has produced the copy of the report he kept, and is going round the offices with it", Reg: rPlain, Form: fComplaint, Topic: "argument"},
}

// encCapacityDigital — the calendar, the lobby, badges, and a drone delivering
// flowers to a floor nobody works on.
var encCapacityDigital = []skel{
	{Text: "The visitor badges have run out", Reg: rPlain, Topic: "kit"},
	{Text: "Reception has four ambassadors waiting", Reg: rPlain, Topic: "count"},
	{Text: "The gift closet will not shut", Reg: rWry, Topic: "building"},
	{Text: "Nobody owns the visiting choir", Reg: rJoke, Topic: "noise"},
	{Text: "The lobby camera shows a queue", Reg: rPlain, Topic: "people"},
	{Text: "The delegation has been added to a calendar that has no gaps in it", Reg: rPlain, Form: fNotice, Topic: "time"},
	{Text: "A drone delivered flowers to a floor that nobody works on", Reg: rPlain, Topic: "machine"},
	{Text: "The channel for foreign relations has been muted by somebody in it", Reg: rWry, Topic: "message"},
	{Text: "Two analysts have been assigned to read the same greeting twice", Reg: rPlain, Topic: "paper"},
	{Text: "The network flagged the fourteenth identical invitation as spam", Reg: rWry, Topic: "message"},
	{Text: "A translation service has been running all morning on an empty room", Reg: rWry, Topic: "money"},
	{Text: "The conference room is booked out until the end of the quarter", Reg: rPlain, Form: fNotice, Topic: "building"},
	{Text: "Three envoys are in reception and one of them has fallen asleep", Reg: rPlain, Topic: "sleep"},
	{Text: "Somebody printed the wrong name on the welcome screen again", Reg: rPlain, Topic: "name"},
	{Text: "Two drones are circling the building waiting for landing clearance", Reg: rPlain, Topic: "border"},
	{Text: "The storage room downstairs now has a spreadsheet of its own", Reg: rWry, Form: fLedger, Topic: "haul"},
	{Text: "An assistant has been moving the same meeting for eleven weeks, and the last four messages about it have gone to a person who left in March", Reg: rPlain, Topic: "message"},
	{Text: "The uplink was left open through an entire welcome speech, and the far end has since asked, politely, for a copy of what was said afterwards", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "An analyst has been photographing the gifts for insurance, has got through sixty of them, and has been told that the insurance office wants them photographed against a plain wall", Reg: rPlain, Form: fLedger, Topic: "paper"},
	{Text: "The office manager has begun turning arrivals away politely at the door, and has developed a form of words for it that the whole floor can now recite", Reg: rWry, Form: fOverheard, Topic: "people"},
	{Text: "Four identical gift boxes arrived from four different embassies on the same afternoon, and three of them have been opened and one has been put in a cupboard", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "The security desk stopped logging deliveries at some point on Wednesday, and the log resumed on Thursday in a different hand, with a note about the missing day", Reg: rPlain, Form: fLedger, Topic: "paper"},
	{Text: "A crate of ceremonial hardware is under a desk on the second floor, and the person who sits at that desk has been working with their knees at an angle for nine days and has raised it with facilities three times", Reg: rPlain, Form: fComplaint, Topic: "argument"},
	{Text: "The parking outside has been taken by diplomatic vehicles since Monday, which is fine, except that the delivery entrance is behind them and the sandwich order has been left at the barrier three days running", Reg: rWry, Form: fComplaint, Topic: "food"},
	{Text: "The embassy inbox is full", Reg: rPlain, Topic: "message"},
	{Text: "The drone pad has a waiting list", Reg: rWry, Topic: "machine"},
	{Text: "Their gift is a subscription", Reg: rJoke, Topic: "message"},
	{Text: "An analyst has been assigned to track which delegation is angry about waiting, and in what order", Reg: rPlain, Topic: "people"},
	{Text: "The network is carrying so many diplomatic video calls that the rest of the building has gone back to phoning", Reg: rWry, Topic: "machine"},
	{Text: "One delegation's greeting has been sent by every channel at once, including the fax", Reg: rPlain, Topic: "message"},
	{Text: "A livestream of the queue outside the embassy has more followers than the embassy", Reg: rWry, Topic: "rumour"},
	{Text: "The gift from one delegation is a server rack, and it arrived switched on", Reg: rPlain, Topic: "machine"},
	{Text: "The after-action notes from last week's reception have been circulated as the plan for next week's", Reg: rWry, Topic: "paper"},
	{Text: "A drone carrying a gift has been hovering over the forecourt for six hours because the delivery app needs a signature and the person who can sign is in another meeting", Reg: rJoke, Topic: "machine"},
	{Text: "An intern has been put in charge of the gift database and has already found three gifts logged by two different people", Reg: rPlain, Topic: "people"},
	{Text: "The fusion plant sent over a gift of power credits and the accounts system has no field for it", Reg: rWry, Topic: "money"},
	{Text: "Every screen in reception is showing one delegation's welcome video on a loop, because it was uploaded first and the setting to change it is somewhere in a menu the receptionist has yet to find", Reg: rPlain, Topic: "machine"},
	{Text: "The embassy channel has been renamed by one of the delegations, and the new name is a greeting in their language that the analysts are still arguing about", Reg: rWry, Topic: "message"},
	{Text: "A delegation has sent a humanoid robot as its ambassador, and it has been waiting in reception with perfect patience since Tuesday, and the human envoys have started sitting next to it for company", Reg: rJoke, Topic: "kit"},

	// --- added by the third late-age review: the building, the city, the network ---
	{Text: "The courier drones are queueing", Reg: rPlain, Topic: "machine"},
	{Text: "Reception's wifi has given up", Reg: rPlain, Topic: "machine"},
	{Text: "The guest network password is unspellable", Reg: rWry, Topic: "kit"},
	{Text: "An ambassador has been on hold since lunch", Reg: rPlain, Topic: "message"},
	{Text: "The front desk has started a sweepstake on which delegation leaves first", Reg: rWry, Topic: "money"},
	{Text: "One envoy is livestreaming the wait to his audience at home", Reg: rWry, Topic: "rumour"},
	{Text: "The delegations' cars have taken the road outside and the food trucks have parked in behind them", Reg: rPlain, Topic: "town"},
	{Text: "The office printer jammed on the fortieth welcome letter", Reg: rPlain, Topic: "paper"},
	{Text: "One gift is a smart speaker that answers every question anyone asks in the lobby", Reg: rJoke, Topic: "kit"},
	{Text: "The gift scanner takes one item at a time, and the queue for it goes back past the lifts", Reg: rPlain, Topic: "haul"},
	{Text: "The fusion plant's liaison has been waiting in the lobby for so long that he now runs the lobby's coffee rota", Reg: rWry, Topic: "people"},
	{Text: "A junior press officer has been told to read every delegation's social accounts for signs of offence", Reg: rPlain, Topic: "rumour"},
	{Text: "Two envoys from delegations at war with each other are sharing a charging cable in the corner", Reg: rPlain, Topic: "stranger"},
	{Text: "The embassy liaison's out-of-office reply now runs to three paragraphs and a map of the building", Reg: rJoke, Topic: "message"},
	{Text: "A delegation has hired a lobbyist to lobby for them to be let into the lobby", Reg: rJoke, Topic: "trade"},
	{Text: "A food truck outside has named a sandwich after the delegation that has waited longest", Reg: rWry, Topic: "trade"},
	{Text: "Security has stopped scanning the gifts and started photographing them instead", Reg: rPlain, Topic: "authority"},
	{Text: "The server room is keeping the perishable gifts cold", Reg: rPlain, Topic: "food"},
	{Text: "The analysts say the third delegation's gift is listening", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The night guard wants to know which delegation owns the parrot", Reg: rWry, Form: fComplaint, Topic: "animal"},
	{Text: "Visiting envoys may use the staff canteen between two and three only", Reg: rPlain, Form: fNotice, Topic: "food"},
	{Text: "The embassy inbox is being answered in the order the mail arrived, and has reached last Tuesday", Reg: rWry, Form: fLedger, Topic: "message"},
	{Text: "The trade envoys set up a pop-up stall in the lobby and sold out by noon", Kinds: []string{"mercantile"}, Reg: rWry, Topic: "trade"},
	{Text: "Their bodyguards have not taken their sunglasses off indoors", Kinds: []string{"aggressive"}, Reg: rPlain, Topic: "weapon"},
	{Text: "The peace envoys have started volunteering on the front desk", Kinds: []string{"peaceful"}, Reg: rPlain, Topic: "people"},
	{Text: "One delegation has been waiting long enough to qualify for the staff pension scheme", Tones: []Tone{Wry}, Reg: rJoke, Topic: "money"},
}

// encCapacityCosmic — the dock schedule, the manifest, and a gift left in vacuum
// long enough to become a different shape.
var encCapacityCosmic = []skel{
	{Text: "The forward bay is diplomatic now", Reg: rWry, Topic: "building"},
	{Text: "Three freighters holding at the marker", Reg: rPlain, Form: fLedger, Topic: "border"},
	{Text: "Somebody welded a shelf up", Reg: rPlain, Topic: "kit"},
	{Text: "The relay has forty unanswered greetings", Reg: rPlain, Form: fLedger, Topic: "message"},
	{Text: "Nobody claimed the strapped crate", Reg: rPlain, Topic: "haul"},
	{Text: "The dock schedule has been rewritten twice this shift and once more since", Reg: rPlain, Form: fNotice, Topic: "time"},
	{Text: "A delegation has been in the transit lounge for nine hours", Reg: rPlain, Topic: "people"},
	{Text: "Something on the manifest is listed only as ceremonial", Reg: rWry, Form: fLedger, Topic: "paper"},
	{Text: "Two visiting envoys have settled into a maintenance corridor with their bags", Reg: rPlain, Topic: "stranger"},
	{Text: "A gift was left in vacuum and is now a different shape", Reg: rWry, Topic: "haul"},
	{Text: "The reactor deck runs warm, so the perishables went down there", Reg: rPlain, Topic: "food"},
	{Text: "An ambassador has memorised the whole station safety briefing by now", Reg: rWry, Topic: "paper"},
	{Text: "The dock crew have stopped asking anybody who they are", Reg: rPlain, Topic: "authority"},
	{Text: "A live animal came aboard and has been given its own cabin", Reg: rPlain, Topic: "animal"},
	{Text: "Four transponders are broadcasting the same friendly message at once", Reg: rPlain, Topic: "message"},
	{Text: "Somebody's ceremonial statue will not go through the hatch", Reg: rWry, Topic: "building"},
	{Text: "The airlock queue has become a matter of protocol, and two junior officers have spent the morning working out an order that offends the smallest number of people", Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "A freighter captain has been waiting behind three diplomatic ships since yesterday, and has been on the relay about it four times, and is being answered by a machine", Reg: rWry, Form: fComplaint, Topic: "message"},
	{Text: "The station gardens have been closed for a reception since the middle of last week, and the reception has been moved to next week without the gardens being reopened", Reg: rPlain, Topic: "building"},
	{Text: "The manifest lists ninety crates and describes none of them, and the officer who wrote it has been off shift since the third delegation docked", Reg: rPlain, Form: fLedger, Topic: "paper"},
	{Text: "Two delegations in orbit are refusing to come down in the order the dock has given them, and the dock has stopped offering an order and is now offering times", Reg: rWry, Topic: "argument"},
	{Text: "The hull outside the visitors' berth has been polished twice this month by a crew that has not been outside for any other reason in two years", Reg: rWry, Topic: "machine"},
	{Text: "Rations have been traded off for foreign delicacies that most of the crew will not eat, and the two who will eat them have become very popular at the wrong end of the day", Reg: rWry, Topic: "trade"},
	{Text: "A hologram of somebody's founder stands in a service corridor with its feet in a puddle of condensate, and it has been switched off twice by maintenance and switched back on twice by protocol staff", Reg: rJoke, Topic: "religion"},
	{Text: "More ships are still on approach", Reg: rPlain, Topic: "stranger"},
	{Text: "Their gifts left home centuries ago", Reg: rWry, Topic: "time"},
	{Text: "One envoy is a backup of another", Reg: rWry, Topic: "people"},
	{Text: "A gift has been placed in orbit because it will not fit through any door on the station", Reg: rPlain, Topic: "haul"},
	{Text: "The delegation that arrived this morning was sent in answer to a message we have no memory of sending", Reg: rWry, Topic: "time"},
	{Text: "The visiting crews have been given leave on the station and there is nowhere left for them to take it", Reg: rPlain, Topic: "people"},
	{Text: "One gift is still being uploaded, and it has been uploading since the start of the month", Reg: rWry, Topic: "stranger"},
	{Text: "The ship's mind has been asked to arrange the gifts by importance and would first like importance defined", Reg: rJoke, Topic: "machine"},
	{Text: "A delegation that set out with gifts for our grandparents has arrived, and has been told the gifts will be accepted on their behalf", Reg: rPlain, Topic: "time"},
	{Text: "The treaty that obliges us to receive every delegation was signed by a government that had not expected anyone to come this far", Reg: rWry, Topic: "authority"},
	{Text: "One of the visiting ships has no crew that anyone has seen, and its gift was delivered by the ship itself, carefully, through the cargo lock", Reg: rPlain, Topic: "stranger"},
	{Text: "An envoy from a civilisation that no longer exists at home wants to stay, and his application will be read at a meeting held once every ten years", Reg: rWry, Topic: "people"},
	{Text: "The storage ring is full, and the overflow is tethered outside in a long line that can be seen from the observation deck", Reg: rPlain, Topic: "haul"},
	{Text: "By the time we finish thanking this delegation, the civilisation that sent it will have become a different civilisation", Reg: rWry, Topic: "time"},
	{Text: "The oldest envoy on the station has been waiting for an audience since before the current administrator was born, has outlived three previous ones, and says she is in no hurry", Reg: rPlain, Topic: "people"},

	// --- added by the third late-age review: the ring, the fleet, and what is out past it ---
	{Text: "The docking ring is full", Reg: rPlain, Topic: "building"},
	{Text: "Diplomatic ships are circling the moon", Reg: rPlain, Topic: "border"},
	{Text: "Every guest berth has two envoys in it", Reg: rPlain, Topic: "sleep"},
	{Text: "An envoy's gift has been thawing in the cargo bay for a week", Reg: rPlain, Topic: "haul"},
	{Text: "The air budget has a line for delegations now", Reg: rWry, Form: fLedger, Topic: "count"},
	{Text: "A delegation arrived in cold sleep and has been left asleep until there is a room for it", Reg: rPlain, Topic: "sleep"},
	{Text: "A spacewalk crew has been sent out to untangle the gift tethers", Reg: rPlain, Topic: "kit"},
	{Text: "One envoy is a swarm and has been asked to wait in a single room", Reg: rJoke, Topic: "stranger"},
	{Text: "An ambassador made of light has been given a dark room to wait in, as a courtesy", Reg: rPlain, Topic: "stranger"},
	{Text: "The embassy module was docked at an angle that makes everyone inside it seasick", Reg: rWry, Topic: "building"},
	{Text: "Delegations from three star systems have to be greeted in the order their messages were sent, which was before any of the greeters were born", Reg: rWry, Topic: "time"},
	{Text: "The newest delegation arrived as a signal with no ship behind it, and has been living in the station's memory since, taking up a little more of it each day, and the archivists have been told to keep acknowledging it", Reg: rPlain, Topic: "stranger"},
	{Text: "The dock officer has learned a greeting in each envoy's language and gets most of them wrong", Reg: rWry, Topic: "message"},
	{Text: "One gift is a small moon", Reg: rJoke, Topic: "haul"},
	{Text: "The crews in the transit lounge say one of the envoy ships was here before the station was built", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The dock crew want lead-lined gloves for handling the gifts from the hot envoy", Reg: rWry, Form: fComplaint, Topic: "kit"},
	{Text: "Visitors are reminded that the observation deck is not a sleeping area", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Fourteen ships are holding for six free berths", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "A habitat on the far side of the ring has offered to take the overflow envoys at a nightly rate", Reg: rWry, Topic: "trade"},
	{Text: "The trade fleet has set up a market in orbit and is selling to the other delegations", Kinds: []string{"mercantile"}, Reg: rWry, Topic: "trade"},
	{Text: "Their warship has parked with its guns trained on the gift bay and has held that position for three days", Kinds: []string{"aggressive"}, Reg: rPlain, Topic: "weapon"},
	{Text: "Their delegation will not dock and speaks only through a small probe", Kinds: []string{"isolationist"}, Reg: rPlain, Topic: "stranger"},
	{Text: "The peace envoys have volunteered for hydroponics duty while they wait", Kinds: []string{"peaceful"}, Reg: rPlain, Topic: "food"},
	{Text: "The station's gift registry now wants its own seat at the protocol table", Tones: []Tone{Wry}, Reg: rJoke, Topic: "paper"},
	{Text: "The envoy from the dead world keeps asking whether his world has answered yet", Reg: rPlain, Topic: "stranger"},
}
