package flavor

// EncounterStandoff — we met a civilization we are at war with, and nobody
// swung.
//
// The log line above this one already says an encounter happened and produced
// nothing, so none of these sentences report that. They report the afternoon:
// who moved first, whose blade stayed up, what got shouted across the gap, the
// dog that walked between the lines, the body handed back, and the two armed
// crowds being awkwardly courteous at a careful distance. See catalog.go for the
// rules.

// --- slot banks --------------------------------------------------------------
//
// Two categories, both true in every age, because the sentences that draw them
// are ungated: a measured gap, and a small act. Lowercase, no terminal
// punctuation, nothing era-coded.

var encStandoffBanks = map[string][]string{
	// The gap between the two lines, as somebody afterwards described it.
	"enc_standoff_distance": {
		"twenty paces", "forty paces", "a hundred paces", "two hundred paces",
		"a dozen paces", "a good ten paces", "a careful fifty paces",
		"eighty paces of open ground", "thirty paces of wet grass",
		"a good throw", "a short throw", "half a field",
		"the width of the stream", "the width of the ditch", "the breadth of the water",
		"the length of a shout", "the distance a voice carries", "an arm's length",
	},
	// The small acts. All of them are the whole of what was exchanged.
	"enc_standoff_gesture": {
		"a raised hand", "a nod nobody returned", "a flat open palm",
		"a bow of the head", "two fingers lifted", "a slow wave", "a shrug",
		"a stiff nod", "a lowered blade", "a long look", "a step backwards",
		"a shaken head", "a hand off the belt", "a whistle with no answer",
		"a mouthed word", "a hand raised and dropped", "a knuckle to the forehead",
		"a wave that was almost friendly",
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

// encStandoffAny fires in every age. Distance, water, hands, weather, dogs,
// children and silence are timeless; gates, sidings and airlocks are not.
var encStandoffAny = []skel{
	// --- who moved, and who did not ---
	{Text: "Neither line moved for the better part of an hour"},
	{Text: "The man on the left took one step and thought again"},
	{Text: "Somebody's foot slipped and both sides flinched at once"},
	{Text: "Their front rank sat down. Ours had no idea what to do"},
	{Text: "One of ours walked forward six steps and then walked back"},
	{Text: "A woman at the front of their line yawned"},
	{Text: "Each side waited for the other one to leave first"},
	{Text: "The two lines stood there until the light went"},
	{Text: "Everyone stayed exactly where they were and got very cold"},
	{Text: "Nobody wanted to be the first to sit down"},

	// --- the measured gap ---
	{Text: "Both lines stopped at ~ and neither of them moved", Slot: "enc_standoff_distance"},
	{Text: "Neither side came closer than ~ for the whole afternoon", Slot: "enc_standoff_distance"},
	{Text: "Our line kept ~ between itself and theirs", Slot: "enc_standoff_distance"},
	{Text: "The closest anyone got was ~, and that was clumsiness", Slot: "enc_standoff_distance"},
	{Text: "The shouting carried fine across ~, which surprised nobody", Slot: "enc_standoff_distance"},
	{Text: "Somebody paced the gap afterwards and called it ~", Slot: "enc_standoff_distance"},

	// --- the weapons ---
	{Text: "Nobody put a blade away and nobody drew one either"},
	{Text: "Their blades stayed up the whole time. So did ours"},
	{Text: "One of them lowered a weapon and was told to raise it"},
	{Text: "A man at the back kept sharpening something, slowly"},
	{Text: "Somebody's hand did not leave his belt the entire time"},
	{Text: "Two of ours had already put their weapons down, quietly"},
	{Text: "An old soldier on our side never once looked up"},

	// --- what was shouted ---
	{Text: "Somebody shouted an insult and nobody bothered to answer"},
	{Text: "A single word was shouted twice and then dropped"},
	{Text: "The insults were competent. Nobody improved on them"},
	{Text: "Their side shouted first and ran out of things to say"},
	{Text: "An old joke about our mothers was shouted and ignored"},
	{Text: "The insults ran out at roughly the same time on both sides"},
	{Text: "One man on their side laughed, and that carried"},

	// --- water and ground ---
	{Text: "Both sides drank from the same stream, twenty steps apart"},
	{Text: "The stream was low and both of them needed it more than a fight"},
	{Text: "Ours took water first and theirs waited without complaining"},
	{Text: "There was one good patch of shade and neither side took it"},
	{Text: "It rained on both of them equally and improved nothing"},

	// --- weather ---
	{Text: "The wind came up and everyone pretended not to be cold"},
	{Text: "Fog came in and both lines quietly lost sight of each other"},
	{Text: "The sun went behind the hill and that settled it"},
	{Text: "Hail started, and two hundred people stood in it anyway"},

	// --- a body handed back ---
	{Text: "They gave back a body and took nothing for the trouble"},
	{Text: "One of their dead was carried halfway out and set down"},
	{Text: "A dead man changed hands and both sides stepped away"},
	{Text: "Somebody handed over a dead man's knife without being asked"},

	// --- animals and children ---
	{Text: "A dog crossed between the lines and was fed by both"},
	{Text: "The dog went over to their side and did not come back"},
	{Text: "A child on their side waved. Nobody waved back"},
	{Text: "Somebody's goat wandered across and was politely returned"},
	{Text: "Two children were sent away and watched from the slope"},

	// --- accidental courtesy ---
	{Text: "What passed between the two lines was ~ and little else", Slot: "enc_standoff_gesture"},
	{Text: "The whole meeting came down to ~ and a great deal of standing", Slot: "enc_standoff_gesture"},
	{Text: "Their leader managed ~ before turning around and leaving", Slot: "enc_standoff_gesture"},
	{Text: "Somebody on our side offered ~ and immediately regretted it", Slot: "enc_standoff_gesture"},
	{Text: "It ended with ~ from a man nobody knows", Slot: "enc_standoff_gesture"},
	{Text: "An old man on the far side gave ~ and turned away", Slot: "enc_standoff_gesture"},

	// --- somebody recognised ---
	{Text: "One of theirs was recognised. Nobody said the name aloud"},
	{Text: "Two cousins found each other across the lines and said nothing"},
	{Text: "A man over there used to live here. He kept quiet"},
	{Text: "Somebody recognised a face and decided against mentioning it"},

	// --- a wounded man let through ---
	{Text: "They let a wounded man walk through and nobody stopped him"},
	{Text: "A limping woman was allowed past and given water"},
	{Text: "The badly hurt one was carried between the lines untouched"},

	// --- a trade attempted ---
	{Text: "Someone offered salt. It was refused with great courtesy"},
	{Text: "A trade was proposed, considered a while, and declined"},
	{Text: "They wanted to swap prisoners and neither side had any"},

	// --- a boundary, and the withdrawal ---
	{Text: "A stake was driven into the ground and left standing"},
	{Text: "Both sides agreed where the ground stops belonging to them"},
	{Text: "They backed away first, slowly, and without turning around"},
	{Text: "Our lot left at dusk. Theirs left later, to make a point"},
	{Text: "Neither group would go until the other one started going"},

	// --- afterwards ---
	{Text: "A promise was made that neither side intends to keep"},
	{Text: "Somebody swore this would not happen twice. It will"},
	{Text: "The story told tonight has more shouting in it than there was"},
	{Text: "Everyone walked home describing it as a victory"},
	{Text: "Nobody slept well and nobody would say that out loud"},
	{Text: "A name was learned and then written down wrong"},

	// --- kind: aggressive ---
	{Text: "Their whole line wanted a fight and did not get one", Kinds: []string{"aggressive"}},
	{Text: "One of them spat on the ground and left it at that", Kinds: []string{"aggressive"}},
	{Text: "They came looking for blood and went home without any", Kinds: []string{"aggressive"}},
	{Text: "Their captain drew a blade, held it up, and put it away", Kinds: []string{"aggressive"}},

	// --- kind: mercantile ---
	{Text: "They tried to sell us something in the middle of it", Kinds: []string{"mercantile"}},
	{Text: "Their side counted our numbers the way a man counts coins", Kinds: []string{"mercantile"}},
	{Text: "Someone over there quoted a price and got laughed at", Kinds: []string{"mercantile"}},

	// --- kind: isolationist ---
	{Text: "They wanted us gone rather more than they wanted us dead", Kinds: []string{"isolationist"}},
	{Text: "Their people put a marker down and walked back behind it", Kinds: []string{"isolationist"}},
	{Text: "Not one of them spoke. That was apparently the message", Kinds: []string{"isolationist"}},

	// --- tone ---
	{Text: "Two armies stood in the rain being extremely polite", Tones: []Tone{Wry}},
	{Text: "The most dangerous moment of the day was a sneeze", Tones: []Tone{Wry}},
	{Text: "Nobody won anything and both sides marched off like winners", Tones: []Tone{Wry}},
	{Text: "The whole thing was resolved by everyone getting bored", Tones: []Tone{Wry}},

	// --- the enemy, in frames a mixed-number name survives ---
	{Text: "Somebody shouted the name {subject} across the gap and stopped", Needs: needSubject},
	{Text: "Nobody here has ever stood that close to {subject} before", Needs: needSubject},
	{Text: "The men spent the evening arguing about what {subject} wanted", Needs: needSubject},
	{Text: "The long war with {subject} paused for one whole afternoon", Needs: needSubject},
	{Text: "Two of ours walked out to meet {subject} and came back", Needs: needSubject},
	{Text: "Somebody drew a picture of {subject} in the dirt, unkindly", Needs: needSubject},
}

// encStandoffAncient — elders, spears, herds, the ford, the watch-fire.
var encStandoffAncient = []skel{
	{Text: "The elders wanted a fight and did not get to have one"},
	{Text: "Two spears were planted in the ground and left there"},
	{Text: "Their herd and ours grazed the same slope all afternoon"},
	{Text: "An arrow was nocked, held, and then quietly put away"},
	{Text: "Both sides watered at the ford and took turns doing it"},
	{Text: "The watch-fire on the ridge burned all night for nothing"},
	{Text: "Somebody rolled a hide out and sat on it, watching"},
	{Text: "The drums started and then somebody sensibly stopped them"},
	{Text: "They stood on the far bank of the ford until dusk"},
	{Text: "A boy carried water out and came back with the cup"},
	{Text: "Nobody went near the store-pit while they were watching"},
	{Text: "Smoke from their fires went straight up, the same as ours"},
	{Text: "Their spears were better made and everyone here noticed"},
	{Text: "A spear was thrown short on purpose and nobody moved"},
	{Text: "The valley was wide enough for both of them, barely"},
	{Text: "Each side kept to its own end of the spring"},
	{Text: "One of the elders walked out alone and came back slower"},
	{Text: "The huts were left standing and everyone made a point of it"},
	{Text: "Two herds were driven apart before anything could start"},
	{Text: "The old road between the camps stayed empty all day"},
	{Text: "Someone drove a stake in beside the road and walked off"},
	{Text: "Their column came down the valley and stopped halfway"},
	{Text: "The harvest was standing and neither side trampled it"},
	{Text: "An old woman shouted from behind their spears and was heard"},
	{Text: "A hide was traded for nothing and handed straight back"},
	{Text: "Frost on the grass, and two lines of men standing in it"},
	{Text: "Their dogs and ours settled the border before the men did"},
	{Text: "The elders talked about it for three nights afterwards"},
}

// encStandoffFeudal — gates, bells, heralds, pickets, the muster.
var encStandoffFeudal = []skel{
	{Text: "A herald went out, said his piece, and came back thirsty"},
	{Text: "Their banner came down the road and stopped at the ford"},
	{Text: "The gate stayed shut and the militia stayed behind it"},
	{Text: "Two riders met in the middle and talked about the weather"},
	{Text: "The bells were rung once and then somebody thought better"},
	{Text: "Their steward and ours discussed the harvest, of all things"},
	{Text: "A horn went up on their side and got no answer at all"},
	{Text: "The muster was called off before anyone reached the yard"},
	{Text: "Their pickets and ours could see each other all night"},
	{Text: "A cart of grain was let through and nobody touched it"},
	{Text: "The clerks wrote it all down and made it sound duller"},
	{Text: "Somebody sent a jug of drink over and it came back empty"},
	{Text: "The village between them was left alone by both sides"},
	{Text: "Their column halted at the edge of the yard and waited"},
	{Text: "A wagon was turned around without a word being said"},
	{Text: "The quartermaster counted their numbers and stopped at four hundred"},
	{Text: "Both sides ate supper within sight of one another"},
	{Text: "The storehouse doors were shut for the first time in months"},
	{Text: "A boy carried a message across and was given bread"},
	{Text: "Their banners came down and ours came down a moment later"},
	{Text: "The road past the ford was left open by agreement"},
	{Text: "Someone pressed a seal into wax for a promise nobody wanted"},
	{Text: "The militia went home to the harvest and were glad to"},
	{Text: "An arrow stuck in the gate post and stayed there"},
	{Text: "Smoke from their camp drifted over the valley all evening"},
	{Text: "The undercroft was stocked for a siege that never came"},
	{Text: "Riders came out from both sides and neither drew a blade"},
	{Text: "A clerk read from parchment until everyone stopped listening"},
}

// encStandoffIndustrial — depots, sidings, the works whistle, triplicate.
var encStandoffIndustrial = []skel{
	{Text: "The telegram went out before anyone knew what to put in it"},
	{Text: "Two work parties met at the siding and stood about"},
	{Text: "A freight train went past between the two lines, slowly"},
	{Text: "The foreman walked over, shook a hand, and walked back"},
	{Text: "Both sides waited on the platform for a train neither wanted"},
	{Text: "The telephone in the depot office rang and nobody picked up"},
	{Text: "A report was filed in triplicate and immediately mislaid"},
	{Text: "The lorries were left running the whole time, wasting fuel"},
	{Text: "Their men and ours smoked in the yard, ten feet apart"},
	{Text: "The warehouse doors were shut and a padlock went on them"},
	{Text: "Somebody cut the wire and then spliced it back together"},
	{Text: "The night shift watched from the annexe roof and said little"},
	{Text: "A works whistle went and both sides jumped for no reason"},
	{Text: "The depot clock was the only thing that moved for an hour"},
	{Text: "Payroll was late and the men had worse things to think about"},
	{Text: "Their foreman and ours had worked together once, years back"},
	{Text: "The rail line between the two camps stayed open all week"},
	{Text: "Somebody telegraphed the whole affair in about nine words"},
	{Text: "A version fit for the shareholders is already being drafted"},
	{Text: "Smoke from their stacks and ours met somewhere over the valley"},
	{Text: "Two columns halted a street apart and neither advanced"},
	{Text: "The road out of the works was left clear for them"},
	{Text: "A crate of tools changed hands and was carried straight back"},
	{Text: "The freight was left on the siding and guarded by both"},
	{Text: "Nobody in the office will put a name to who moved first"},
	{Text: "The men were told to stand easy and stood anything but"},
	{Text: "Somebody took a photograph and it came out badly blurred"},
	{Text: "The train was held twenty minutes to let them all pass"},
}

// encStandoffDigital — feeds, uplinks, drones, analysts, the open channel.
var encStandoffDigital = []skel{
	{Text: "The drones circled each other for eleven minutes and left"},
	{Text: "Two analysts watched the same feed and disagreed politely"},
	{Text: "The uplink stayed open and neither side said a word on it"},
	{Text: "Somebody opened a channel, breathed into it, and closed it"},
	{Text: "The after-action briefing has one line in it and much white space"},
	{Text: "A drone was flown low over them and flown back again"},
	{Text: "The feed shows two groups of people not moving, for an hour"},
	{Text: "Nobody on the network wants to be the one who logged it"},
	{Text: "Their drone and ours hovered at the same height, watching"},
	{Text: "The channel logs are twenty minutes of somebody breathing"},
	{Text: "An analyst called it a near miss and went to get coffee"},
	{Text: "The road between the two positions stayed empty all night"},
	{Text: "Somebody's phone rang and both sides heard it clearly"},
	{Text: "The network flagged it as an engagement, which is generous"},
	{Text: "Two drones came down at once and neither was shot at"},
	{Text: "The feed cut for four seconds and came back to the same"},
	{Text: "An analyst has been asked to explain the silence, twice"},
	{Text: "Their comms and ours were on one channel by accident"},
	{Text: "Wind and boots, and forty minutes of nothing on the uplink"},
	{Text: "Somebody streamed the whole thing and got very few viewers"},
	{Text: "The drone footage has been reviewed and shows nobody blinking"},
	{Text: "One of the analysts noticed both sides carried the same rifles"},
	{Text: "The after-action report is one page and mostly headings"},
	{Text: "Nobody has closed the channel and nobody is talking on it"},
	{Text: "Two vehicles idled facing each other for most of the morning"},
	{Text: "The network went quiet, which the analysts found unsettling"},
	{Text: "A camera on their drone was pointed straight at ours"},
	{Text: "The uplink dropped for a moment and everyone stayed put"},
}

// encStandoffCosmic — hulls, bays, transponders, the airlock, the relay.
var encStandoffCosmic = []skel{
	{Text: "Two hulls sat a kilometre apart and did nothing for hours"},
	{Text: "Their transponder answered ours and then went quiet again"},
	{Text: "Both ships held orbit and neither of them turned to face"},
	{Text: "The relay logged the meeting as a routine crossing"},
	{Text: "Somebody opened an airlock, looked out, and closed it again"},
	{Text: "Their reactor was running hot and everyone could see it"},
	{Text: "The dock crew were stood down and told to wait longer"},
	{Text: "A freighter passed between them and the moment was over"},
	{Text: "Nobody moved a gun and nobody moved out of the way"},
	{Text: "The manifest they sent over was a list of nothing useful"},
	{Text: "Two crews stared through the same stretch of vacuum, briefly"},
	{Text: "Their hull markings were read aloud and copied down wrong"},
	{Text: "The bulkhead lights were dimmed on both ships at once"},
	{Text: "A shuttle went halfway across and turned back on its own"},
	{Text: "The transponders exchanged codes and neither side liked it"},
	{Text: "Somebody in the forward bay waved. It is not clear at what"},
	{Text: "The orbital watch called it an approach and left it there"},
	{Text: "Both ships vented at the same moment, entirely by chance"},
	{Text: "One of their bays opened and closed without anything leaving"},
	{Text: "The relay picked up a conversation that never quite began"},
	{Text: "Their pilot matched our vector exactly, which was unnerving"},
	{Text: "The reactor alarm went off and everyone assumed the worst"},
	{Text: "A dead crewman was sent across and their side took him in"},
	{Text: "Nothing crossed the gap except a very poor radio joke"},
	{Text: "The dock was cleared for a fight that did not arrive"},
	{Text: "A boarding party got as far as the tube and stopped"},
	{Text: "Their airlock stayed shut for the whole of the encounter"},
	{Text: "The vacuum between them measured four hundred metres, exactly"},
}
