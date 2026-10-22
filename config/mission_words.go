package config

// mission_words.go is the words the mission board builds its offers from.
// Every age has the names of its four tiers, six places, four objectives for
// each tier and six twists. An offer's title is
//
//	<tier name>: <objective> <place>
//
// and its description is a twist, so the objectives end in "the" and the
// places are bare names ("lift a data core from the" + "Chrome Market").
// Twists fit any tier. All of it is placeholder text for the owner's copy
// pass: plain, short, and nothing before its age.

// The four tier names an age has, in the order of MissionWords.Tiers.
const (
	WordsShort = iota
	WordsLong
	WordsGrandMilitary
	WordsGrandCivil
)

// Road event weights are relative odds within a rolled event.
const (
	RoadStormWeight   = 3
	RoadHelpersWeight = 3
	RoadFindWeight    = 2
	RoadTollWeight    = 2
	RoadFoesWeight    = 2
	RoadAmbushWeight  = 2
	RoadFairWeight    = 2
	RoadRuinsWeight   = 1
)

// roadEventTemplates are short player-facing stories, selected by kind.
var roadEventTemplates = map[string][]string{
	"storm":       {"<weather> pinned the party down near the <place>. <time> lost.", "The party waited out <weather> by the <place>, losing <time>.", "<weather> turned the road near the <place> to trouble. The delay cost <time>."},
	"helpers":     {"<helper> showed the party a shorter way. <time> saved.", "A tip from <helper> cut <time> off the journey.", "The party took a shortcut with help from <helper>, saving <time>."},
	"find":        {"The party found <find>. The haul grows by <percent>%.", "A lucky find, <find>, adds <percent>% to the haul.", "<find> was waiting by the road. The haul grows by <percent>%."},
	"toll":        {"On the <place> road, <foe> blocked the way. The captain decided his hide was worth more than a fight and paid <gold> gold to pass.", "<foe> demanded <gold> gold at the <place>. The captain paid and kept moving.", "The party paid <gold> gold to <foe> for safe passage near the <place>."},
	"foes":        {"<foe> shadowed the party for days. The odds fall to <odds>%.", "The party spotted <foe> on the trail. Odds drop to <odds>%.", "<foe> made the road dangerous. Success odds are now <odds>%."},
	"ambush":      {"<foe> struck at dusk. <n> soldiers fell.", "<foe> caught the soldiers off guard. <n> fell in the ambush.", "At dusk, <foe> attacked. The party lost <n> soldiers."},
	"ambush_haul": {"<foe> made off with part of the haul.", "<foe> raided the supplies and took part of the haul.", "The party fought off <foe>, but some of the haul was lost."},
	"fair":        {"Fair skies and a good guide. The odds rise to <odds>%.", "A clear stretch lifted spirits. Success odds are now <odds>%.", "The party found its rhythm. Odds rise to <odds>%."},
	"ruins":       {"The party marked ruins off the road near the <place>. They are on the board now.", "An old ruin near the <place> is now marked on the board.", "The party found ruins near the <place>. A new offer is on the board."},
}

// RoadTemplate returns a kind's story template at index n.
func (MissionWords) RoadTemplate(kind string, n int) string {
	pool := roadEventTemplates[kind]
	if len(pool) == 0 {
		return ""
	}
	return pool[n%len(pool)]
}

// RoadTemplateCount is the number of story templates for kind.
func (MissionWords) RoadTemplateCount(kind string) int { return len(roadEventTemplates[kind]) }

// MissionWords is one age's words for the mission board.
type MissionWords struct {
	Age string
	// Tiers are the names of the short, long, grand military and grand civil
	// missions.
	Tiers  [4]string
	Places [6]string
	// The objectives of each tier: the verb phrases an offer's title is
	// built from. They end in "the", and a place follows.
	ShortObjectives    [4]string
	LongObjectives     [4]string
	MilitaryObjectives [4]string
	CivilObjectives    [4]string
	Twists             [6]string
	Weather            [3]string
	Foes               [3]string
	Helpers            [3]string
	Finds              [3]string
}

// ObjectivesFor is the four objectives of a tier, by the index of its name
// (WordsShort and the others).
func (w MissionWords) ObjectivesFor(tier int) [4]string {
	switch tier {
	case WordsLong:
		return w.LongObjectives
	case WordsGrandMilitary:
		return w.MilitaryObjectives
	case WordsGrandCivil:
		return w.CivilObjectives
	}
	return w.ShortObjectives
}

// MissionWordsFor is the words of an age; false for an age that has none.
func MissionWordsFor(age string) (MissionWords, bool) {
	for _, w := range missionWords {
		if w.Age == age {
			return w, true
		}
	}
	return MissionWords{}, false
}

// AllMissionWords is the words of every age, in age order.
func AllMissionWords() []MissionWords {
	out := make([]MissionWords, len(missionWords))
	copy(out, missionWords)
	return out
}

var missionWords = []MissionWords{
	{
		Age:     "primitive_age",
		Weather: [3]string{"a dust storm", "a flash flood", "a hailstorm"}, Foes: [3]string{"wolves", "rival clan", "raiders"}, Helpers: [3]string{"a friendly camp", "a shepherd", "a trading party"}, Finds: [3]string{"a bone cache", "a flint hoard", "a buried shrine"},
		Tiers:              [4]string{"Forage", "Hunt", "Raid", "Great Migration"},
		Places:             [6]string{"Dry Creek", "Birch Hollow", "Mud Flats", "Flint Ridge", "Fern Valley", "Long Meadow"},
		ShortObjectives:    [4]string{"gather roots along the", "pick berries near the", "follow animal tracks to the", "dig for grubs by the"},
		LongObjectives:     [4]string{"track a herd across the", "drive game toward the", "chase a mammoth over the", "stalk elk along the"},
		MilitaryObjectives: [4]string{"steal food from camps near the", "drive off strangers at the", "ambush hunters on the", "take back stolen fire from the"},
		CivilObjectives:    [4]string{"lead the tribe past the", "find a winter camp beyond the", "carry the fire across the", "settle the clan by the"},
		Twists: [6]string{
			"Rain washes out the first night's tracks.",
			"A pack of wolves trails the party.",
			"The fire goes out on the second night.",
			"Someone swears they heard a roar.",
			"The wind carries smoke from far off.",
			"Frost comes early this year.",
		},
	},
	{
		Age:     "stone_age",
		Weather: [3]string{"a dust storm", "a flash flood", "a hailstorm"}, Foes: [3]string{"bandits", "raiders", "a rival clan"}, Helpers: [3]string{"a friendly village", "a shepherd", "a trading post"}, Finds: [3]string{"a flint cache", "a stone hoard", "a buried shrine"},
		Tiers:              [4]string{"Forage", "Trek", "Raid", "Hill Fort Raising"},
		Places:             [6]string{"Salt Lick", "Flint Pits", "Cave Mouth", "Standing Stones", "Red Clay Bank", "Bone Hills"},
		ShortObjectives:    [4]string{"gather flint at the", "pick herbs around the", "scout game trails near the", "fish the pools below the"},
		LongObjectives:     [4]string{"trek to the far side of the", "follow the river past the", "trade hides at the", "map the high ground above the"},
		MilitaryObjectives: [4]string{"raid the rival clan at the", "burn the stockade by the", "take the stone quarry at the", "drive the raiders from the"},
		CivilObjectives:    [4]string{"raise stone walls on the", "haul great stones to the", "dig a ditch around the", "build the long house at the"},
		Twists: [6]string{
			"A rival clan's smoke rises in the distance.",
			"The elders say the signs are good.",
			"Someone twists an ankle on loose scree.",
			"The stone here rings like a bell.",
			"Cold nights keep the party close to the fire.",
			"A child's carving turns up along the path.",
		},
	},
	{
		Age:     "bronze_age",
		Weather: [3]string{"a dust storm", "a flash flood", "a hailstorm"}, Foes: [3]string{"bandits", "raiders", "a rival clan"}, Helpers: [3]string{"a friendly village", "a shepherd", "a trading post"}, Finds: [3]string{"an amber hoard", "a tin cache", "a buried shrine"},
		Tiers:              [4]string{"Scout", "Trade Trek", "Skirmish", "Trade Mission"},
		Places:             [6]string{"Copper Hills", "Tin Road", "Harbor Town", "Clay Terraces", "Amber Coast", "Cedar Pass"},
		ShortObjectives:    [4]string{"scout the road to the", "count the herds near the", "check the wells along the", "watch the border at the"},
		LongObjectives:     [4]string{"carry copper to the", "trade pots for tin at the", "follow the amber road to the", "escort a caravan through the"},
		MilitaryObjectives: [4]string{"skirmish with raiders at the", "hold the ford near the", "take the watchtower at the", "chase bandits out of the"},
		CivilObjectives:    [4]string{"open a trade house at the", "send gifts to the chiefs of the", "chart a trade route through the", "sign a pact at the"},
		Twists: [6]string{
			"A merchant offers a bad price and a good story.",
			"The tin seller keeps changing his mind.",
			"Smiths along the road want to see your bronze.",
			"A bridge is out, and the detour adds a day.",
			"Bandits were seen here last month.",
			"The harvest festival fills the road with carts.",
		},
	},
	{
		Age:     "iron_age",
		Weather: [3]string{"a hard rain", "a cold snap", "a windstorm"}, Foes: [3]string{"raiders", "warriors", "a rival clan"}, Helpers: [3]string{"a roadside village", "a shepherd", "a trading post"}, Finds: [3]string{"an iron cache", "a coin hoard", "a buried shrine"},
		Tiers:              [4]string{"Survey", "Caravan", "Siege", "Founding Expedition"},
		Places:             [6]string{"Ironhill", "Black Gate", "Salt Road", "Stone Bridge", "Border Keep", "Wolf Forest"},
		ShortObjectives:    [4]string{"survey the roads around the", "map the iron seams under the", "check the milestones near the", "measure the wall at the"},
		LongObjectives:     [4]string{"lead a caravan to the", "haul iron ore to the", "ride the long road to the", "carry grain through the"},
		MilitaryObjectives: [4]string{"lay siege to the", "break the gates of the", "starve out the keep at the", "storm the walls of the"},
		CivilObjectives:    [4]string{"found a town near the", "raise a fort beside the", "settle farmers at the", "dig a canal toward the"},
		Twists: [6]string{
			"A toll collector asks too many questions.",
			"The forge smoke can be seen for miles.",
			"Rain turns the road to mud.",
			"A deserter begs to join the party.",
			"The harvest has been poor along the way.",
			"An old beacon still burns on the hill.",
		},
	},
	{
		Age:     "classical_age",
		Weather: [3]string{"a sea squall", "a heat wave", "a hard rain"}, Foes: [3]string{"pirates", "mercenaries", "a rival city"}, Helpers: [3]string{"a village guide", "a merchant", "a way station"}, Finds: [3]string{"a coin hoard", "a bronze cache", "a buried shrine"},
		Tiers:              [4]string{"Survey", "Trade Voyage", "Legion March", "Embassy"},
		Places:             [6]string{"Marble Harbor", "Olive Coast", "Forum Road", "Hill of Oaks", "Amber Strait", "Great Aqueduct"},
		ShortObjectives:    [4]string{"survey the roads of the", "count the columns at the", "inspect the aqueduct near the", "chart the coast near the"},
		LongObjectives:     [4]string{"sail with wine to the", "carry marble to the", "trade oil and grain at the", "voyage around the cape of the"},
		MilitaryObjectives: [4]string{"march the legion to the", "besiege the walls of the", "hold the pass above the", "cut the supply road to the"},
		CivilObjectives:    [4]string{"send an embassy to the", "bring gifts to the senate of the", "host the games at the", "raise a theater at the"},
		Twists: [6]string{
			"A rhetorician argues with the party for a day.",
			"Omens are read at dawn and disagree.",
			"The magistrate insists on a long speech.",
			"A storm closes the harbor for two days.",
			"A scribe asks to record everything.",
			"Statues line the road like a crowd.",
		},
	},
	{
		Age:     "medieval_age",
		Weather: [3]string{"a snow squall", "a hard rain", "a windstorm"}, Foes: [3]string{"brigands", "soldiers", "a rival lord"}, Helpers: [3]string{"a village guide", "a ferryman", "an innkeeper"}, Finds: [3]string{"a silver hoard", "a locked chest", "a buried shrine"},
		Tiers:              [4]string{"Patrol", "Pilgrimage", "Siege", "Grand Pilgrimage"},
		Places:             [6]string{"Oak Abbey", "Grey Castle", "Pilgrim's Ford", "Wolf Moor", "King's Road", "Mill Town"},
		ShortObjectives:    [4]string{"patrol the lanes around the", "ride the bounds of the", "watch the road near the", "check the beacon fires at the"},
		LongObjectives:     [4]string{"walk the pilgrim road to the", "carry relics to the", "escort monks to the", "trade wool and tin at the"},
		MilitaryObjectives: [4]string{"lay siege to the", "storm the walls of the", "burn the siege camp at the", "hold the ford at the"},
		CivilObjectives:    [4]string{"lead the great pilgrimage to the", "raise a cathedral beside the", "found an abbey near the", "carry the king's gift to the"},
		Twists: [6]string{
			"A friar offers blessings for a small fee.",
			"Fog lies thick on the road.",
			"A minstrel sings about the party too early.",
			"The local lord demands a toll.",
			"Bells ring in the next valley.",
			"A bad harvest leaves the villages tense.",
		},
	},
	{
		Age:     "renaissance_age",
		Weather: [3]string{"a spring storm", "a hard rain", "a sea squall"}, Foes: [3]string{"privateers", "mercenaries", "a rival court"}, Helpers: [3]string{"a local guide", "a merchant", "a harbor master"}, Finds: [3]string{"a coin hoard", "a spice cache", "a buried shrine"},
		Tiers:              [4]string{"Exploration", "Voyage", "Campaign", "Grand Voyage"},
		Places:             [6]string{"Gilded Coast", "Spice Isles", "Painted Harbor", "Inkwell Bay", "Silk Quarter", "Far Meridian"},
		ShortObjectives:    [4]string{"explore the shore of the", "draw maps of the", "scout inland from the", "take soundings off the"},
		LongObjectives:     [4]string{"sail for the", "trade cloth and glass at the", "carry letters of credit to the", "seek a new passage past the"},
		MilitaryObjectives: [4]string{"march on the garrison at the", "besiege the fortress of the", "land troops at the", "drive the pirates from the"},
		CivilObjectives:    [4]string{"chart a western passage beyond the", "found a trading house on the", "outfit a great fleet for the", "bring scholars to the"},
		Twists: [6]string{
			"A patron demands the party name something after him.",
			"An astronomer swears the stars have moved.",
			"Pirates were seen off the coast.",
			"A painter asks to join the voyage.",
			"The compass drifts a few degrees.",
			"Printed pamphlets about the trip arrive before you do.",
		},
	},
	{
		Age:     "colonial_age",
		Weather: [3]string{"a monsoon", "a hard rain", "a sea squall"}, Foes: [3]string{"privateers", "raiders", "a rival company"}, Helpers: [3]string{"a local guide", "a trader", "a harbor master"}, Finds: [3]string{"a silver cache", "a spice hoard", "a buried shrine"},
		Tiers:              [4]string{"Survey", "Colony Run", "Campaign", "Found a Colony"},
		Places:             [6]string{"Cape Harbor", "New Bay", "Tobacco Coast", "Fur River", "Fort Haven", "Green Islands"},
		ShortObjectives:    [4]string{"survey the coast of the", "map the river mouth at the", "scout trails inland from the", "count the stores at the"},
		LongObjectives:     [4]string{"sail supplies to the", "carry furs from the", "trade with the settlers at the", "run the blockade to the"},
		MilitaryObjectives: [4]string{"march on the rival fort at the", "storm the harbor of the", "hold the colony line at the", "burn the stockade at the"},
		CivilObjectives:    [4]string{"found a colony on the", "settle families beside the", "raise a fort and a mill at the", "plant the first crops at the"},
		Twists: [6]string{
			"A fever runs through the ship's crew.",
			"Locals trade fish for iron needles.",
			"The charter is late and the ink is blotted.",
			"A storm blows the party off course.",
			"A map turns out to be wrong.",
			"Settlers argue about who owns the land.",
		},
	},
	{
		Age:     "industrial_age",
		Weather: [3]string{"a coal haze", "a hard rain", "a boiler burst"}, Foes: [3]string{"bandits", "rival workers", "a street gang"}, Helpers: [3]string{"a rail porter", "a station agent", "a trading post"}, Finds: [3]string{"a tool cache", "a coin hoard", "a buried store"},
		Tiers:              [4]string{"Prospect", "Rail Expedition", "Offensive", "Great Railway"},
		Places:             [6]string{"Coal Valley", "Iron Basin", "Canal Junction", "Smoke Hills", "Steam Harbor", "Mill Row"},
		ShortObjectives:    [4]string{"prospect for coal in the", "test the seams of the", "survey a rail line through the", "count the furnaces of the"},
		LongObjectives:     [4]string{"lay track to the", "haul ore across the", "run a freight line past the", "open a canal toward the"},
		MilitaryObjectives: [4]string{"launch an offensive against the", "shell the works at the", "cut the rail line to the", "take the foundry at the"},
		CivilObjectives:    [4]string{"build the great railway through the", "bridge the gorge of the", "lay a trunk line to the", "raise a station at the"},
		Twists: [6]string{
			"A boiler bursts and delays the whole line.",
			"The foreman wants more pay and more coal.",
			"Soot hangs over everything.",
			"An inventor begs the party to test his engine.",
			"A whistle sounds and the whole town stops.",
			"The telegraph wires hum with rumors.",
		},
	},
	{
		Age:     "victorian_age",
		Weather: [3]string{"a London fog", "a hard rain", "a gale"}, Foes: [3]string{"highwaymen", "rival agents", "a street gang"}, Helpers: [3]string{"a station porter", "a local guide", "a trading post"}, Finds: [3]string{"a brass cache", "a coin hoard", "a buried store"},
		Tiers:              [4]string{"Prospect", "Expedition", "Offensive", "Polar Expedition"},
		Places:             [6]string{"Ice Barrier", "Cape Frost", "Amber Falls", "Brass Harbor", "Gaslight Quarter", "Silver River"},
		ShortObjectives:    [4]string{"prospect for gold in the", "chart the foothills of the", "survey the coast near the", "gather specimens from the"},
		LongObjectives:     [4]string{"lead an expedition up the", "cross the unmapped lands of the", "trace the source of the", "bring back specimens from the"},
		MilitaryObjectives: [4]string{"launch an offensive on the", "bombard the fort at the", "advance on the capital of the", "hold the line at the"},
		CivilObjectives:    [4]string{"sail for the pole beyond the", "raise the great exhibition at the", "chart the ice around the", "winter over at the"},
		Twists: [6]string{
			"A naturalist insists on a day to study beetles.",
			"The tea runs out on the second day.",
			"A newspaper has sent a reporter along.",
			"The expedition's trunks are heavier than the map.",
			"Fog rolls in thick as soup.",
			"A telegram arrives with good news and bad news.",
		},
	},
	{
		Age:     "electric_age",
		Weather: [3]string{"a blackout", "a heat wave", "a thunderstorm"}, Foes: [3]string{"gangsters", "rival agents", "a street gang"}, Helpers: [3]string{"a station agent", "a local guide", "a relay operator"}, Finds: [3]string{"a parts cache", "a coin hoard", "a buried store"},
		Tiers:              [4]string{"Aerial Survey", "Expedition", "Offensive", "Grand Exhibition"},
		Places:             [6]string{"Sparkling Plains", "Lightning Ridge", "Dynamo Hill", "Copper Valley", "Lamp City", "Power Gorge"},
		ShortObjectives:    [4]string{"survey from the air above the", "photograph the", "check the power lines across the", "count the pylons of the"},
		LongObjectives:     [4]string{"lead an expedition across the", "string cables through the", "drive motor cars over the", "trace the grid to the"},
		MilitaryObjectives: [4]string{"advance with armor on the", "bomb the depots at the", "cut the telegraph to the", "break the line at the"},
		CivilObjectives:    [4]string{"light up the great exhibition at the", "stage a world's fair at the", "wire a whole city near the", "open a grand hall at the"},
		Twists: [6]string{
			"A generator coughs and the lights flicker.",
			"A crowd gathers to watch the machine.",
			"The radio crackles with a faint voice.",
			"A bulb shatters, then another.",
			"The inventor refuses to explain how it works.",
			"Dust blows across the airfield.",
		},
	},
	{
		Age:     "atomic_age",
		Weather: [3]string{"a radiation storm", "a blackout", "a dust storm"}, Foes: [3]string{"raiders", "rival patrol", "a street gang"}, Helpers: [3]string{"a field medic", "a relay operator", "a trading post"}, Finds: [3]string{"a supply cache", "a reactor part", "a buried vault"},
		Tiers:              [4]string{"Recon Flight", "Research Expedition", "Operation", "Manhattan-scale Project"},
		Places:             [6]string{"Desert Range", "Radar Hill", "Blue Lab", "Silver Coast", "Pine Ridge", "Test Flats"},
		ShortObjectives:    [4]string{"fly recon over the", "photograph the", "track weather over the", "map the coast by air near the"},
		LongObjectives:     [4]string{"lead a research expedition to the", "measure radiation near the", "carry instruments to the", "study the rocks of the"},
		MilitaryObjectives: [4]string{"launch an operation against the", "drop paratroops on the", "hold a bridgehead at the", "jam the radar at the"},
		CivilObjectives:    [4]string{"build a vast laboratory at the", "split the first atom at the", "raise a research city beside the", "fund a program for the"},
		Twists: [6]string{
			"A Geiger counter clicks louder than it should.",
			"The sirens test at noon and scare everyone.",
			"A secret memo goes astray.",
			"The coffee is terrible and the scientists are worse.",
			"Static fills the radio for an hour.",
			"An officer demands the party sign forms.",
		},
	},
	{
		Age:     "modern_age",
		Weather: [3]string{"a thunderstorm", "a flash flood", "a heat wave"}, Foes: [3]string{"smugglers", "a rival crew", "a street gang"}, Helpers: [3]string{"a local guide", "a relief station", "a trading post"}, Finds: [3]string{"a supply cache", "a rare mineral", "a buried vault"},
		Tiers:              [4]string{"Recon", "Field Op", "Strike", "Moonshot"},
		Places:             [6]string{"Harbor Terminal", "Border Ridge", "Silicon Park", "Salt Desert", "Grey Tower", "Ring Road"},
		ShortObjectives:    [4]string{"run recon on the", "scout the perimeter of the", "tail a convoy near the", "watch the airstrip at the"},
		LongObjectives:     [4]string{"run a field op in the", "extract a team from the", "escort a convoy across the", "survey the coast with drones near the"},
		MilitaryObjectives: [4]string{"launch a strike on the", "raid the depot at the", "take the airfield at the", "blockade the port of the"},
		CivilObjectives:    [4]string{"launch a rocket from the", "build a lunar base for the", "fund a moon mission at the", "ship a moon lander from the"},
		Twists: [6]string{
			"A phone buzzes at a bad moment.",
			"The satellite photo is a day old.",
			"Traffic jams the way out.",
			"A reporter has found the story first.",
			"The generator coughs twice and holds.",
			"Someone left the headlights on.",
		},
	},
	{
		Age:     "information_age",
		Weather: [3]string{"a thunderstorm", "a flash flood", "a heat wave"}, Foes: [3]string{"hackers", "a rival crew", "a street gang"}, Helpers: [3]string{"a local guide", "a relay station", "a trading post"}, Finds: [3]string{"a data cache", "a server part", "a buried drive"},
		Tiers:              [4]string{"Data Run", "Satellite Op", "Cyber Strike", "Global Launch"},
		Places:             [6]string{"Server Farm", "Fiber Trunk", "Data Haven", "Relay Hub", "Glass Campus", "Cable Landing"},
		ShortObjectives:    [4]string{"trace packets through the", "scrape records from the", "map the network of the", "test the firewall of the"},
		LongObjectives:     [4]string{"task a satellite over the", "lay a fiber line to the", "link the exchange at the", "sweep the airwaves above the"},
		MilitaryObjectives: [4]string{"launch a cyber strike on the", "crash the grid of the", "jam the uplink at the", "seize the servers of the"},
		CivilObjectives:    [4]string{"launch a global network from the", "open a world portal at the", "switch on the search engine at the", "stream the first broadcast from the"},
		Twists: [6]string{
			"The connection drops at the worst second.",
			"A modem screeches in the next room.",
			"An email chain gets longer and louder.",
			"The logs fill with strange pings.",
			"A coffee spill takes out one terminal.",
			"Someone is already in the system.",
		},
	},
	{
		Age:     "digital_age",
		Weather: [3]string{"a packet storm", "a power surge", "a network flood"}, Foes: [3]string{"rogue agents", "malware swarms", "a rival process"}, Helpers: [3]string{"a helper daemon", "a relay station", "a trusted node"}, Finds: [3]string{"an old server", "a data cache", "a buried backup"},
		Tiers:              [4]string{"Data Run", "Network Op", "Cyber Strike", "Network Founding"},
		Places:             [6]string{"Cloud Vault", "Packet Plaza", "Mirror Site", "Dark Port", "Deep Cache", "Open Grid"},
		ShortObjectives:    [4]string{"run a data sweep through the", "crawl the index of the", "audit the keys of the", "trace a signal into the"},
		LongObjectives:     [4]string{"run a network op through the", "migrate the archive to the", "tunnel under the", "sync the nodes of the"},
		MilitaryObjectives: [4]string{"launch a cyber strike on the", "take down the relays of the", "poison the cache of the", "lock out the admins of the"},
		CivilObjectives:    [4]string{"found a new network at the", "bring the whole city online at the", "raise a data commons in the", "launch a public ledger on the"},
		Twists: [6]string{
			"A server in another time zone wakes up.",
			"Latency spikes and nobody knows why.",
			"A patch breaks something unrelated.",
			"The cache holds more than it should.",
			"Two messages arrive with the same timestamp.",
			"A stranger on the channel offers help.",
		},
	},
	{
		Age:     "cyberpunk_age",
		Weather: [3]string{"a smog squall", "a power surge", "an acid rainstorm"}, Foes: [3]string{"a street gang", "corporate enforcers", "a rival crew"}, Helpers: [3]string{"a street medic", "a fixer", "a relay station"}, Finds: [3]string{"a data cache", "a chrome stash", "a buried server"},
		Tiers:              [4]string{"Recon", "Heist", "Assault", "Arcology Raising"},
		Places:             [6]string{"Chrome Market", "Neon Alley", "Rain Towers", "Static Docks", "Ash Arcade", "Black Plaza"},
		ShortObjectives:    [4]string{"scout the streets of the", "tap the cameras of the", "tail a courier through the", "scan the rooftops above the"},
		LongObjectives:     [4]string{"lift a data core from the", "break into the vault of the", "steal prototypes from the", "ghost a shipment out of the"},
		MilitaryObjectives: [4]string{"storm the corporate tower at the", "assault the checkpoint at the", "burn the security grid of the", "take the gang's hideout under the"},
		CivilObjectives:    [4]string{"raise a sealed arcology above the", "power a whole block in the", "wire a free city under the", "house the poor in a tower over the"},
		Twists: [6]string{
			"The guards change shift at dawn.",
			"Acid rain eats the paint off the van.",
			"A street doc offers a bad deal on an eye.",
			"A billboard shows the party's faces.",
			"The neon flickers in a pattern.",
			"A drone follows for three blocks.",
		},
	},
	{
		Age:     "fusion_age",
		Weather: [3]string{"a solar storm", "a plasma squall", "a coolant leak"}, Foes: [3]string{"raiders", "a rival crew", "rogue drones"}, Helpers: [3]string{"a station crew", "a relay station", "a trading post"}, Finds: [3]string{"a fuel cache", "a reactor part", "a buried module"},
		Tiers:              [4]string{"Sweep", "Salvage", "Assault", "Fusion City"},
		Places:             [6]string{"Reactor Ring", "Slag Fields", "Coolant Lake", "Burned Quarter", "Flux Yard", "Bright Dome"},
		ShortObjectives:    [4]string{"sweep the ruins around the", "scan the shielding of the", "check the coolant lines near the", "map the heat vents of the"},
		LongObjectives:     [4]string{"salvage cores from the", "pull superconductors out of the", "tow a hull from the", "strip a dead plant at the"},
		MilitaryObjectives: [4]string{"assault the rival plant at the", "breach the shield of the", "seize the control room of the", "cut power to the"},
		CivilObjectives:    [4]string{"light a fusion city at the", "build a sun in a bottle for the", "pipe cheap power to the", "raise a tower of glass beside the"},
		Twists: [6]string{
			"The reactor hums a note just off true.",
			"Hot spots drift across the map.",
			"A coolant leak turns the air sweet.",
			"Radiation badges tick over early.",
			"The lights dim and brighten on their own.",
			"Dust glows faintly at night.",
		},
	},
	{
		Age:     "space_age",
		Weather: [3]string{"a solar flare", "a meteor shower", "a radiation storm"}, Foes: [3]string{"pirates", "raiders", "a rival fleet"}, Helpers: [3]string{"a relay station", "a station crew", "a cargo hauler"}, Finds: [3]string{"a mineral cache", "an old probe", "a buried module"},
		Tiers:              [4]string{"Probe", "Orbital Mission", "Orbital Strike", "Colony Ship"},
		Places:             [6]string{"Low Orbit", "Moon Shadow", "Red Plain", "Ring of Ice", "Dust Belt", "Sunward Station"},
		ShortObjectives:    [4]string{"launch a probe past the", "scan the surface of the", "skim the dust of the", "map the craters of the"},
		LongObjectives:     [4]string{"fly an orbital mission around the", "dock with the wreck above the", "mine ice from the", "set a relay beyond the"},
		MilitaryObjectives: [4]string{"strike the enemy station above the", "knock out the satellite over the", "board the raider near the", "blockade the lanes of the"},
		CivilObjectives:    [4]string{"send a colony ship toward the", "raise a ring habitat over the", "seed gardens on the", "build a lunar town in the"},
		Twists: [6]string{
			"A micrometeor pings the hull.",
			"Mission control repeats the same order twice.",
			"A solar flare closes the window.",
			"Someone's cup floats away.",
			"The comm lag adds a quiet pause.",
			"A bolt shakes loose in the night.",
		},
	},
	{
		Age:     "interstellar_age",
		Weather: [3]string{"a solar flare", "a meteor shower", "a radiation storm"}, Foes: [3]string{"pirates", "raiders", "a rival fleet"}, Helpers: [3]string{"a relay station", "a colony crew", "a cargo hauler"}, Finds: [3]string{"an alien relic", "a mineral cache", "a lost probe"},
		Tiers:              [4]string{"Probe", "Deep-Space Mission", "Invasion", "Generation Ship"},
		Places:             [6]string{"Proxima Gate", "Cold Giant", "Long Dark", "Twin Suns", "Drift Station", "Ember Nebula"},
		ShortObjectives:    [4]string{"launch a probe toward the", "scan the planets of the", "sample the dust of the", "listen for signals from the"},
		LongObjectives:     [4]string{"fly a deep-space mission to the", "build a relay chain toward the", "chart the lanes beyond the", "mine comet ice near the"},
		MilitaryObjectives: [4]string{"invade the colony at the", "land troops on a world of the", "take the orbital yards at the", "break the garrison at the"},
		CivilObjectives:    [4]string{"launch a generation ship for the", "seed a new world near the", "carry a million sleepers to the", "raise a star colony by the"},
		Twists: [6]string{
			"The long dark is quiet and the crew is too.",
			"A star winks out on the charts.",
			"Cargo shifts and the ship hums.",
			"A message from home arrives years late.",
			"The sleepers dream the same dream.",
			"The engines purr and then cough.",
		},
	},
	{
		Age:     "galactic_age",
		Weather: [3]string{"a nebula squall", "a solar flare", "a meteor shower"}, Foes: [3]string{"pirates", "raiders", "a rival fleet"}, Helpers: [3]string{"a relay station", "a colony crew", "a trade envoy"}, Finds: [3]string{"an alien relic", "a star map", "a lost probe"},
		Tiers:              [4]string{"Survey Drone", "Star Voyage", "Invasion", "Stellar Settlement"},
		Places:             [6]string{"Orion Reach", "Crimson Rim", "Quiet Arm", "Core Gate", "Lantern Cluster", "Pale Spiral"},
		ShortObjectives:    [4]string{"send a survey drone through the", "count the stars of the", "scan the worlds of the", "map the lanes around the"},
		LongObjectives:     [4]string{"take a star voyage across the", "open a jump lane through the", "trade across the", "carry seed ships to the"},
		MilitaryObjectives: [4]string{"invade the fleet at the", "storm the throneworld of the", "blockade the star gates of the", "drive the armada from the"},
		CivilObjectives:    [4]string{"settle a thousand worlds in the", "raise a star ring around the", "found a stellar city in the", "gather a council at the"},
		Twists: [6]string{
			"A pulsar keeps perfect time on the ship's clock.",
			"A hail from an unknown fleet gets no reply.",
			"The map has a hole where a star should be.",
			"A comet crosses the lane right on schedule.",
			"An old beacon repeats a dead language.",
			"The jump leaves everyone a little dizzy.",
		},
	},
	{
		Age:     "quantum_age",
		Weather: [3]string{"a probability storm", "a phase shift", "a temporal squall"}, Foes: [3]string{"rogue agents", "a rival timeline", "a quantum swarm"}, Helpers: [3]string{"a relay mind", "a phase guide", "a trading node"}, Finds: [3]string{"a qubit cache", "a lost archive", "a buried simulation"},
		Tiers:              [4]string{"Phase Probe", "Rift Expedition", "Incursion", "Pocket Universe"},
		Places:             [6]string{"Fold Garden", "Mirror Gap", "Null Shore", "Still Point", "Hollow Spiral", "Echo Vault"},
		ShortObjectives:    [4]string{"phase a probe through the", "measure the noise of the", "sample a mirror from the", "trace the edges of the"},
		LongObjectives:     [4]string{"lead a rift expedition into the", "map the layers of the", "bring back a ghost from the", "hold open the door to the"},
		MilitaryObjectives: [4]string{"launch an incursion into the", "collapse the guard of the", "slip a strike team into the", "unravel the defenses of the"},
		CivilObjectives:    [4]string{"grow a pocket universe in the", "fold a new home inside the", "weave a stable bubble in the", "teach the rift to hold at the"},
		Twists: [6]string{
			"Everyone remembers the departure differently.",
			"A second copy of the map appears.",
			"Time stutters for about a second.",
			"The compass points at several places at once.",
			"A door appears and then thinks better of it.",
			"A familiar voice answers from nowhere.",
		},
	},
	{
		Age:     "transcendent_age",
		Weather: [3]string{"a reality storm", "a dimensional tide", "a time squall"}, Foes: [3]string{"rogue minds", "a rival cosmos", "a void swarm"}, Helpers: [3]string{"a guide intelligence", "a relay mind", "a trading nexus"}, Finds: [3]string{"a memory hoard", "a lost world", "a buried universe"},
		Tiers:              [4]string{"Glimpse", "Pilgrimage of Light", "Reality Incursion", "Ascension Rite"},
		Places:             [6]string{"First Dawn", "Silent Choir", "Open Horizon", "Garden of Hours", "Last Shore", "Shining Stair"},
		ShortObjectives:    [4]string{"catch a glimpse of the", "listen at the edge of the", "follow a thread of light to the", "gather echoes from the"},
		LongObjectives:     [4]string{"walk the pilgrimage of light to the", "carry a spark across the", "sing a path through the", "trace the light back to the"},
		MilitaryObjectives: [4]string{"break into the reality of the", "bend the law of the", "storm the veil around the", "unmake the wall of the"},
		CivilObjectives:    [4]string{"perform the ascension rite at the", "gather every voice at the", "open the final door at the", "raise the last choir at the"},
		Twists: [6]string{
			"The light feels warmer than it should.",
			"No one remembers who spoke first.",
			"Silence lasts a little too long.",
			"A thought arrives before the question.",
			"The ground seems to listen.",
			"A bell rings in no one's hand.",
		},
	},
}
