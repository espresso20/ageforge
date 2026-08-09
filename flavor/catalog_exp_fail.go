package flavor

// ExpeditionFailure — a party was sent out and it went badly.
//
// The log line directly above this one already says "<Name> failed! Partial loot
// recovered." So none of the sentences here say that it failed, and none of them
// weigh the loss against the recovery. They say who did not come home, what broke,
// who is being blamed, what the map got wrong, and what the families were told.
// See catalog.go for the rules.

// --- slot banks --------------------------------------------------------------
//
// Noun phrases, lowercase, one tight category per bank, no era-coded nouns —
// both banks are drawn by ungated sentences, so both must land in every age.

var expFailBanks = map[string][]string{
	// A carried thing that did not come home. Definite noun phrases, so the frame
	// supplies only the verb.
	"exp_fail_kit": {
		"the spare rope", "the good lamp", "the map case", "the tinder box",
		"the big cook pot", "the second pack", "the drinking water", "the heavy axe",
		"the spare boots", "the medicine box", "the climbing line", "the long knife",
		"the folded bedding", "the last dry wood", "the signal mirror",
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
	// --- who did not come back ---
	{Text: "Nine went out. Seven came back and one of those is carried"},
	{Text: "Two names have been scratched off the list and not replaced"},
	{Text: "They buried one on the way out and one on the way back"},
	{Text: "Nobody has said the cook's name out loud since they got in"},
	{Text: "The woman who led them out is not among those who came in"},
	{Text: "The count at the door came to three fewer than went out"},
	{Text: "Somebody is still out there and the search stopped at dusk"},
	{Text: "They left two behind and could not go back for them"},

	// --- wounds ---
	{Text: "The bandage came off in the doorway and nobody looked"},
	{Text: "Two of them will not walk properly again and know it"},
	{Text: "Nobody wants to look too long at ~", Slot: "exp_fail_wound"},
	{Text: "They keep changing the wrapping on ~ and saying nothing", Slot: "exp_fail_wound"},
	{Text: "There is an argument about who is to blame for ~", Slot: "exp_fail_wound"},
	{Text: "A child asked about ~ and was taken outside", Slot: "exp_fail_wound"},
	{Text: "Everyone has an opinion about ~ and none of them agrees", Slot: "exp_fail_wound"},
	{Text: "Nothing was said all evening about ~", Slot: "exp_fail_wound"},
	{Text: "The blood on the floor was mopped before anyone asked about it"},

	// --- what broke, what was left behind ---
	{Text: "Nobody went back for ~ and nobody has said so", Slot: "exp_fail_kit"},
	{Text: "They came home without ~ and would not discuss it", Slot: "exp_fail_kit"},
	{Text: "The river took ~ on the second morning", Slot: "exp_fail_kit"},
	{Text: "Somebody carried ~ the whole way for nothing", Slot: "exp_fail_kit"},
	{Text: "Somebody swears ~ went out on the pack that morning", Slot: "exp_fail_kit"},
	{Text: "Half a day was lost going back for ~", Slot: "exp_fail_kit"},
	{Text: "Everything they set out with fits into one small pile now"},
	{Text: "The rope was cut, not broken, and that is the whole problem"},

	// --- a count that does not add up ---
	{Text: "The count was done three times and came out wrong three times"},
	{Text: "Two lists were made and neither one agrees with the other"},

	// --- the town, the families ---
	{Text: "A woman has been waiting at the door since first light"},
	{Text: "The children were kept indoors and told nothing at all"},
	{Text: "One mother has been told. The other has not been found"},
	{Text: "There is a house with the shutters closed at the end of the row"},
	{Text: "The cooking went ahead because it had already been started"},

	// --- burial ---
	{Text: "They dug in hard ground for most of an afternoon"},
	{Text: "The digging was finished before anyone came out to watch"},
	{Text: "Three rocks were set in a row above the water line"},

	// --- what the map got wrong ---
	{Text: "The map was wrong about the crossing and two people paid for it"},
	{Text: "A whole line on the map has been crossed out and left blank"},
	{Text: "The route everyone swore by is marked bad now in fresh ink"},
	{Text: "Whoever drew that map is going to be found and spoken to"},

	// --- rumour and blame ---
	{Text: "The story changes depending on which of them is telling it"},
	{Text: "Two versions are going around and neither one is kind"},
	{Text: "Blame has settled on the quietest one, who is not arguing"},
	{Text: "An argument started at the door and has not stopped since"},
	{Text: "Nobody has blamed the weather yet, which takes some discipline"},
	{Text: "The man who said it would be easy has gone very quiet"},

	// --- silence ---
	{Text: "The one who planned it has not spoken since the door shut"},
	{Text: "Dinner was eaten in a silence nobody tried to break"},
	{Text: "Three of them sat down inside and have not moved in an hour"},

	// --- animals ---
	{Text: "The dog came back alone, two days before anyone else did"},
	{Text: "Only one of the pack animals made it up the last slope"},
	{Text: "A dog that was not theirs followed them home and stayed"},

	// --- weather ---
	{Text: "The rain did not stop for six days and neither did they"},
	{Text: "It froze on the third night and nobody was ready for it"},
	{Text: "The wind took the shelter apart at about the worst hour"},

	// --- names, and one boy ---
	{Text: "A man called Orrek turned back early and has not been forgiven"},
	{Text: "The boy who begged to go is not talking about it now"},
	{Text: "A sealed box came all that way and had nothing inside it"},

	// --- doors, walls, boots, water ---
	{Text: "Someone put a fist through the door frame on the way in"},
	{Text: "The wall by the steps has a new hole at head height"},
	{Text: "Boots came off at the door and nobody has moved since"},
	{Text: "The water ran out a day earlier than the plan allowed"},

	// --- the confident, afterwards ---
	{Text: "The plan was good. Everything that happened to it was not"},
	{Text: "Three people have already explained how they would have done it"},

	// --- kind: scouting ---
	{Text: "They went four days out and turned back on the fifth", Kinds: []string{"scouting"}},
	{Text: "The far side was never reached and will have to be tried again", Kinds: []string{"scouting"}},
	{Text: "Whatever is out past the tree line stayed out past it", Kinds: []string{"scouting"}},
	{Text: "The new marks stop halfway across and start again nowhere", Kinds: []string{"scouting"}},
	{Text: "Two went ahead to look and only one came back to say", Kinds: []string{"scouting"}},

	// --- kind: military ---
	{Text: "The fighting went on longer than the food did", Kinds: []string{"military"}},
	{Text: "They held the ground for a night and gave it back at dawn", Kinds: []string{"military"}},
	{Text: "Four wounded, one dead, and nothing to show for either", Kinds: []string{"military"}},
	{Text: "Whoever they met was better at it and knew the ground", Kinds: []string{"military"}},
	{Text: "The blades came home dull and nobody is cleaning them tonight", Kinds: []string{"military"}},

	// --- tone ---
	{Text: "Nobody has counted the dead out loud yet, and nobody will", Tones: []Tone{Grim}},
	{Text: "The list on the wall is being read by people who cannot read", Tones: []Tone{Grim}},
	{Text: "There are more empty places at the table than full ones", Tones: []Tone{Grim}},
	{Text: "Everyone who stayed home has an excellent theory about it", Tones: []Tone{Wry}},
	{Text: "The plan is being called ambitious by the man who wrote it", Tones: []Tone{Wry}},

	// --- the order title, in a frame a verb-led name survives ---
	{Text: "A page marked {subject} has been turned face down", Needs: needSubject},
	{Text: "Whoever wrote the {subject} order is not in the room", Needs: needSubject},
	{Text: "Somebody has stopped reading at the line that reads {subject}", Needs: needSubject},
	{Text: "Nothing filed under {subject} came back the way it left", Needs: needSubject},
	{Text: "The board still shows the {subject} order as out", Needs: needSubject},
	{Text: "Two men refused the {subject} order and are quiet about it", Needs: needSubject},

	// --- the resource, as a thing in a room rather than as a payout ---
	{Text: "Whatever came back of the {res} fits in one hand", Needs: needRes},
	{Text: "Somebody has put what is left of the {res} in one box", Needs: needRes},
	{Text: "There is not enough {res} here to argue over", Needs: needRes},
	{Text: "Nobody is guarding the {res} tonight because nobody needs to", Needs: needRes},
	{Text: "They set down {res_haul} and nobody moved to help", Needs: needRes | needMassRes},
	{Text: "What is left of {res_stores} sits in the corner untouched", Needs: needRes | needMassRes},
	{Text: "All of it together came to {amt_res} and no more", Needs: needRes | needAmount},
	{Text: "They walked back with {amt_res} and a man on a plank", Needs: needRes | needAmount},
}

// expFailAncient — fires, herds, hides, spears, the elders.
var expFailAncient = []skel{
	{Text: "The elders were told first and have said nothing since"},
	{Text: "Two spears came back and four went out"},
	{Text: "The herd was scattered and only a few were driven home"},
	{Text: "A hide was used to carry someone the last part of the way"},
	{Text: "The hut at the end of the row has nobody in it tonight"},
	{Text: "The drums did not go up and everyone understood why"},
	{Text: "The watch-fire was let go out and nobody has relit it"},
	{Text: "The store-pit will not last the cold at this rate"},
	{Text: "An arrow was cut out of a shoulder by the fire"},
	{Text: "The spring they were promised was dry when they reached it"},
	{Text: "The ford ran higher than anyone remembered and took two"},
	{Text: "Smoke on the far ridge was the last thing three of them saw"},
	{Text: "The valley they crossed has better ways in than the one used"},
	{Text: "The old road washed out behind them and doubled the walk"},
	{Text: "The harvest will be short two pairs of hands this year"},
	{Text: "Nobody has told the elders how many spears came back"},
	{Text: "A boy walked the whole valley and came home with one hide"},
	{Text: "The dogs would not go near what was carried in"},
	{Text: "Three of them walked in wearing hides that were not theirs"},
	{Text: "A spear was left standing in the ground where they stopped"},
	{Text: "Smoke over the camp went sideways all night and choked everyone"},
	{Text: "The column came back in ones and twos over two days"},
	{Text: "Nobody has beaten the drums and nobody is going to"},
	{Text: "The hides they went out for are still lying in the wet"},
	{Text: "An old woman walked out to the road and came back alone"},
	{Text: "The store-pit was opened early, which is its own bad news"},
	{Text: "The elders have started arguing about the spring already"},
	{Text: "A spear point was found in a pack and nobody claims it"},
}

// expFailFeudal — carts, gates, bells, clerks, stewards, musters.
var expFailFeudal = []skel{
	{Text: "The gate was opened at night and shut again very fast"},
	{Text: "One cart came back. Two did not, and neither did the drivers"},
	{Text: "The bells were not rung and the village noticed that"},
	{Text: "The quartermaster has stopped counting and gone to sit down"},
	{Text: "A rider came in ahead with news nobody wanted to hear"},
	{Text: "The clerks have written the same three names four times"},
	{Text: "A banner came back folded around something and stayed folded"},
	{Text: "The militia were called out and stood in the rain for nothing"},
	{Text: "The storehouse doors stayed shut and the steward kept the key"},
	{Text: "Wax was melted for seals that will not be needed now"},
	{Text: "Two wagons are in the yard with the covers still tied down"},
	{Text: "The herald has been told to say very little for once"},
	{Text: "Supper was laid for more than sat down to it"},
	{Text: "A picket ran in ahead of them without being told to"},
	{Text: "The undercroft has room again, which nobody wanted to notice"},
	{Text: "A horn went up at the gate and then stopped halfway"},
	{Text: "The muster roll has three lines drawn through it"},
	{Text: "The ford took a cart and everything that was on it"},
	{Text: "The spring rain turned the road to a trench past the mill"},
	{Text: "Smoke over the valley for two days, and none of it planned"},
	{Text: "The harvest hands were pulled off the fields to dig instead"},
	{Text: "Parchment has been used up on letters to three villages"},
	{Text: "An arrow was still in the cart when it rolled through the gate"},
	{Text: "The village priest was fetched before the steward was"},
	{Text: "The column came home at walking pace with the wagons empty"},
	{Text: "A clerk fainted in the yard and was carried inside"},
	{Text: "The bell rope has been taken down until somebody decides"},
	{Text: "Riders went out again at first light and found nothing"},
}

// expFailIndustrial — depots, sidings, foremen, telegrams, triplicate.
var expFailIndustrial = []skel{
	{Text: "The telegram came in before the men and said very little"},
	{Text: "The depot platform was empty at the hour they were due"},
	{Text: "The foreman signed for what came back and did not look up"},
	{Text: "Freight arrived on the late train with three crates missing"},
	{Text: "A wire went out to head office and has not been answered"},
	{Text: "The lorries came back with one windscreen gone entirely"},
	{Text: "The payroll has fewer names on it than it did last week"},
	{Text: "It is being written up in triplicate and the wording is disputed"},
	{Text: "The siding was kept clear all evening for a train that came late"},
	{Text: "The annexe is being used for something nobody wants to name"},
	{Text: "The warehouse doors were closed early and the men sent home"},
	{Text: "The shareholders will be told in a fortnight and told gently"},
	{Text: "The telephone rang twice in the night and was not picked up"},
	{Text: "Smoke over the yard all day and none of it from the works"},
	{Text: "A lorry is still down in the valley with its axle gone"},
	{Text: "A column of men walked back from the works in the dark"},
	{Text: "The rail out of the depot is bent where the truck came off"},
	{Text: "The telegraph office read it twice and asked for a repeat"},
	{Text: "Two men off the train went straight to the infirmary"},
	{Text: "The foreman's list and the doctor's list do not agree"},
	{Text: "Somebody telegraphed the families before the office decided to"},
	{Text: "The night shift was stood down and stayed on anyway"},
	{Text: "A crate came off the platform and split in front of everyone"},
	{Text: "The warehouse book has a page torn out of the middle"},
	{Text: "There is a wire on the desk that nobody has opened"},
	{Text: "The freight that did come in was left where it landed"},
	{Text: "A lorry engine ran all night outside the annexe for warmth"},
	{Text: "The depot cat has gone missing too, which is being mentioned"},
}

// expFailDigital — uplinks, drones, feeds, analysts, channels.
var expFailDigital = []skel{
	{Text: "The uplink dropped at the worst possible moment and stayed down"},
	{Text: "A drone came home. The people it was watching did not"},
	{Text: "The feed shows eleven seconds that nobody will play again"},
	{Text: "Two analysts have been taken off the floor and sent home"},
	{Text: "The after-action meeting was cancelled and quietly rescheduled"},
	{Text: "Somebody muted the channel rather than listen to the rest"},
	{Text: "The network held perfectly, which somehow makes it worse"},
	{Text: "Three drones went out and one came back on its own"},
	{Text: "An analyst has been staring at the same frame for an hour"},
	{Text: "The last message on the channel is four letters long"},
	{Text: "Nobody has closed the map that is still up on the wall"},
	{Text: "The badge on the desk belongs to somebody not coming in"},
	{Text: "A junior was sent for coffee and did not come back upstairs"},
	{Text: "The footage stops before the part everyone wants to see"},
	{Text: "The uplink logs end mid-sentence and pick up an hour later"},
	{Text: "Two laptops came back and neither one switches on now"},
	{Text: "Somebody has been rewatching it since two in the morning"},
	{Text: "The drone camera survived. Almost nothing else did"},
	{Text: "The team lead has not left the room since the feed cut"},
	{Text: "An after-action draft is open on a screen nobody is reading"},
	{Text: "The channel went quiet in a way that carries down the hall"},
	{Text: "Someone unplugged the wall screen rather than watch it loop"},
	{Text: "The network flagged it as routine forty minutes too late"},
	{Text: "A drone is still out there transmitting to nobody at all"},
	{Text: "The road cameras caught them going out and never coming back"},
	{Text: "Half the floor is standing around one desk saying nothing"},
	{Text: "The analysts stopped talking when the second list came through"},
	{Text: "Somebody's chair has not been moved and nobody will sit in it"},
}

// expFailCosmic — hulls, bays, relays, transponders, airlocks.
var expFailCosmic = []skel{
	{Text: "The hull came back with a hole nobody wants to stand near"},
	{Text: "Cargo bay two is empty and the doors are still open"},
	{Text: "They came out of orbit slow and crooked and nobody cheered"},
	{Text: "The transponder answered late, and answered wrong"},
	{Text: "A bulkhead is buckled inward and has been welded shut for now"},
	{Text: "The airlock cycled once with one fewer than went out"},
	{Text: "The relay lost them for nine hours and found them once"},
	{Text: "A freighter passed close enough to see it and did not stop"},
	{Text: "The reactor was shut down hard and is still cooling"},
	{Text: "Two suits came back empty and were carried in anyway"},
	{Text: "The manifest lists things that are not on the ship"},
	{Text: "There is vacuum frost on a boot in the forward locker"},
	{Text: "The dock crew stopped talking when the ramp came down"},
	{Text: "Somebody scratched two names into the plate beside the hatch"},
	{Text: "The bay doors would not close and were left that way"},
	{Text: "Orbital watch logged the approach and said nothing about it"},
	{Text: "The hull sensors are lying and everyone knows which ones"},
	{Text: "They were on emergency air for the last eleven hours"},
	{Text: "The relay logs have a gap and the gap is the whole story"},
	{Text: "A repair patch let go on the way in and took a panel"},
	{Text: "The reactor alarm has been silenced rather than answered"},
	{Text: "Another crew towed them the last part of the way in"},
	{Text: "Somebody has written the wrong number on the bay door"},
	{Text: "The manifest and the people aboard do not come to one total"},
	{Text: "Nobody has cleaned the forward bay and nobody has been asked"},
	{Text: "The transponder was answering on a dead crew's code"},
	{Text: "Two of them will not go out the airlock again, and said so"},
	{Text: "The dock lights were left on for a ship that came in dark"},
}
