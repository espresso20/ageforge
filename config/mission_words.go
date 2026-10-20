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

// MissionWords is one age's words for the mission board.
type MissionWords struct {
	Age string
	// Tiers are the names of the short, long, grand military and grand civil
	// missions.
	Tiers      [4]string
	Places     [6]string
	Objectives [4][4]string
	Twists     [6]string
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
		Age:    "primitive_age",
		Tiers:  [4]string{"Forage", "Hunt", "Raid", "Great Migration"},
		Places: [6]string{"Dry Creek", "Birch Hollow", "Mud Flats", "Flint Ridge", "Fern Valley", "Long Meadow"},
		Objectives: [4][4]string{
			{"gather roots along the", "pick berries near the", "follow animal tracks to the", "dig for grubs by the"},
			{"track a herd across the", "drive game toward the", "chase a mammoth over the", "stalk elk along the"},
			{"steal food from camps near the", "drive off strangers at the", "ambush hunters on the", "take back stolen fire from the"},
			{"lead the tribe past the", "find a winter camp beyond the", "carry the fire across the", "settle the clan by the"},
		},
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
		Age:    "stone_age",
		Tiers:  [4]string{"Forage", "Trek", "Raid", "Hill Fort Raising"},
		Places: [6]string{"Salt Lick", "Flint Pits", "Cave Mouth", "Standing Stones", "Red Clay Bank", "Bone Hills"},
		Objectives: [4][4]string{
			{"gather flint at the", "pick herbs around the", "scout game trails near the", "fish the pools below the"},
			{"trek to the far side of the", "follow the river past the", "trade hides at the", "map the high ground above the"},
			{"raid the rival clan at the", "burn the stockade by the", "take the stone quarry at the", "drive the raiders from the"},
			{"raise stone walls on the", "haul great stones to the", "dig a ditch around the", "build the long house at the"},
		},
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
		Age:    "bronze_age",
		Tiers:  [4]string{"Scout", "Trade Trek", "Skirmish", "Trade Mission"},
		Places: [6]string{"Copper Hills", "Tin Road", "Harbor Town", "Clay Terraces", "Amber Coast", "Cedar Pass"},
		Objectives: [4][4]string{
			{"scout the road to the", "count the herds near the", "check the wells along the", "watch the border at the"},
			{"carry copper to the", "trade pots for tin at the", "follow the amber road to the", "escort a caravan through the"},
			{"skirmish with raiders at the", "hold the ford near the", "take the watchtower at the", "chase bandits out of the"},
			{"open a trade house at the", "send gifts to the chiefs of the", "chart a trade route through the", "sign a pact at the"},
		},
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
		Age:    "iron_age",
		Tiers:  [4]string{"Survey", "Caravan", "Siege", "Founding Expedition"},
		Places: [6]string{"Ironhill", "Black Gate", "Salt Road", "Stone Bridge", "Border Keep", "Wolf Forest"},
		Objectives: [4][4]string{
			{"survey the roads around the", "map the iron seams under the", "check the milestones near the", "measure the wall at the"},
			{"lead a caravan to the", "haul iron ore to the", "ride the long road to the", "carry grain through the"},
			{"lay siege to the", "break the gates of the", "starve out the keep at the", "storm the walls of the"},
			{"found a town near the", "raise a fort beside the", "settle farmers at the", "dig a canal toward the"},
		},
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
		Age:    "classical_age",
		Tiers:  [4]string{"Survey", "Trade Voyage", "Legion March", "Embassy"},
		Places: [6]string{"Marble Harbor", "Olive Coast", "Forum Road", "Hill of Oaks", "Amber Strait", "Great Aqueduct"},
		Objectives: [4][4]string{
			{"survey the roads of the", "count the columns at the", "inspect the aqueduct near the", "chart the coast near the"},
			{"sail with wine to the", "carry marble to the", "trade oil and grain at the", "voyage around the cape of the"},
			{"march the legion to the", "besiege the walls of the", "hold the pass above the", "cut the supply road to the"},
			{"send an embassy to the", "bring gifts to the senate of the", "host the games at the", "raise a theater at the"},
		},
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
		Age:    "medieval_age",
		Tiers:  [4]string{"Patrol", "Pilgrimage", "Siege", "Grand Pilgrimage"},
		Places: [6]string{"Oak Abbey", "Grey Castle", "Pilgrim's Ford", "Wolf Moor", "King's Road", "Mill Town"},
		Objectives: [4][4]string{
			{"patrol the lanes around the", "ride the bounds of the", "watch the road near the", "check the beacon fires at the"},
			{"walk the pilgrim road to the", "carry relics to the", "escort monks to the", "trade wool and tin at the"},
			{"lay siege to the", "storm the walls of the", "burn the siege camp at the", "hold the ford at the"},
			{"lead the great pilgrimage to the", "raise a cathedral beside the", "found an abbey near the", "carry the king's gift to the"},
		},
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
		Age:    "renaissance_age",
		Tiers:  [4]string{"Exploration", "Voyage", "Campaign", "Grand Voyage"},
		Places: [6]string{"Gilded Coast", "Spice Isles", "Painted Harbor", "Inkwell Bay", "Silk Quarter", "Far Meridian"},
		Objectives: [4][4]string{
			{"explore the shore of the", "draw maps of the", "scout inland from the", "take soundings off the"},
			{"sail for the", "trade cloth and glass at the", "carry letters of credit to the", "seek a new passage past the"},
			{"march on the garrison at the", "besiege the fortress of the", "land troops at the", "drive the pirates from the"},
			{"chart a western passage beyond the", "found a trading house on the", "outfit a great fleet for the", "bring scholars to the"},
		},
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
		Age:    "colonial_age",
		Tiers:  [4]string{"Survey", "Colony Run", "Campaign", "Found a Colony"},
		Places: [6]string{"Cape Harbor", "New Bay", "Tobacco Coast", "Fur River", "Fort Haven", "Green Islands"},
		Objectives: [4][4]string{
			{"survey the coast of the", "map the river mouth at the", "scout trails inland from the", "count the stores at the"},
			{"sail supplies to the", "carry furs from the", "trade with the settlers at the", "run the blockade to the"},
			{"march on the rival fort at the", "storm the harbor of the", "hold the colony line at the", "burn the stockade at the"},
			{"found a colony on the", "settle families beside the", "raise a fort and a mill at the", "plant the first crops at the"},
		},
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
		Age:    "industrial_age",
		Tiers:  [4]string{"Prospect", "Rail Expedition", "Offensive", "Great Railway"},
		Places: [6]string{"Coal Valley", "Iron Basin", "Canal Junction", "Smoke Hills", "Steam Harbor", "Mill Row"},
		Objectives: [4][4]string{
			{"prospect for coal in the", "test the seams of the", "survey a rail line through the", "count the furnaces of the"},
			{"lay track to the", "haul ore across the", "run a freight line past the", "open a canal toward the"},
			{"launch an offensive against the", "shell the works at the", "cut the rail line to the", "take the foundry at the"},
			{"build the great railway through the", "bridge the gorge of the", "lay a trunk line to the", "raise a station at the"},
		},
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
		Age:    "victorian_age",
		Tiers:  [4]string{"Prospect", "Expedition", "Offensive", "Polar Expedition"},
		Places: [6]string{"Ice Barrier", "Cape Frost", "Amber Falls", "Brass Harbor", "Gaslight Quarter", "Silver River"},
		Objectives: [4][4]string{
			{"prospect for gold in the", "chart the foothills of the", "survey the coast near the", "gather specimens from the"},
			{"lead an expedition up the", "cross the unmapped", "trace the source of the", "bring back specimens from the"},
			{"launch an offensive on the", "bombard the fort at the", "advance on the capital of the", "hold the line at the"},
			{"sail for the pole beyond the", "raise the great exhibition at the", "chart the ice around the", "winter over at the"},
		},
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
		Age:    "electric_age",
		Tiers:  [4]string{"Aerial Survey", "Expedition", "Offensive", "Grand Exhibition"},
		Places: [6]string{"Sparkling Plains", "Lightning Ridge", "Dynamo Hill", "Copper Valley", "Lamp City", "Power Gorge"},
		Objectives: [4][4]string{
			{"survey from the air above the", "photograph the", "check the power lines across the", "count the pylons of the"},
			{"lead an expedition across the", "string cables through the", "drive motor cars over the", "trace the grid to the"},
			{"advance with armor on the", "bomb the depots at the", "cut the telegraph to the", "break the line at the"},
			{"light up the great exhibition at the", "stage a world's fair at the", "wire a whole city near the", "open a grand hall at the"},
		},
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
		Age:    "atomic_age",
		Tiers:  [4]string{"Recon Flight", "Research Expedition", "Operation", "Manhattan-scale Project"},
		Places: [6]string{"Desert Range", "Radar Hill", "Blue Lab", "Silver Coast", "Pine Ridge", "Test Flats"},
		Objectives: [4][4]string{
			{"fly recon over the", "photograph the", "track weather over the", "map the coast by air near the"},
			{"lead a research expedition to the", "measure radiation near the", "carry instruments to the", "study the rocks of the"},
			{"launch an operation against the", "drop paratroops on the", "hold a bridgehead at the", "jam the radar at the"},
			{"build a vast laboratory at the", "split the first atom at the", "raise a research city beside the", "fund a program for the"},
		},
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
		Age:    "modern_age",
		Tiers:  [4]string{"Recon", "Field Op", "Strike", "Moonshot"},
		Places: [6]string{"Harbor Terminal", "Border Ridge", "Silicon Park", "Salt Desert", "Grey Tower", "Ring Road"},
		Objectives: [4][4]string{
			{"run recon on the", "scout the perimeter of the", "tail a convoy near the", "watch the airstrip at the"},
			{"run a field op in the", "extract a team from the", "escort a convoy across the", "survey the coast with drones near the"},
			{"launch a strike on the", "raid the depot at the", "take the airfield at the", "blockade the port of the"},
			{"launch a rocket from the", "build a lunar base for the", "fund a moon mission at the", "ship a moon lander from the"},
		},
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
		Age:    "information_age",
		Tiers:  [4]string{"Data Run", "Satellite Op", "Cyber Strike", "Global Launch"},
		Places: [6]string{"Server Farm", "Fiber Trunk", "Data Haven", "Relay Hub", "Glass Campus", "Cable Landing"},
		Objectives: [4][4]string{
			{"trace packets through the", "scrape records from the", "map the network of the", "test the firewall of the"},
			{"task a satellite over the", "lay a fiber line to the", "link the exchange at the", "sweep the airwaves above the"},
			{"launch a cyber strike on the", "crash the grid of the", "jam the uplink at the", "seize the servers of the"},
			{"launch a global network from the", "open a world portal at the", "switch on the search engine at the", "stream the first broadcast from the"},
		},
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
		Age:    "digital_age",
		Tiers:  [4]string{"Data Run", "Network Op", "Cyber Strike", "Network Founding"},
		Places: [6]string{"Cloud Vault", "Packet Plaza", "Mirror Site", "Dark Port", "Deep Cache", "Open Grid"},
		Objectives: [4][4]string{
			{"run a data sweep through the", "crawl the index of the", "audit the keys of the", "trace a signal into the"},
			{"run a network op through the", "migrate the archive to the", "tunnel under the", "sync the nodes of the"},
			{"launch a cyber strike on the", "take down the relays of the", "poison the cache of the", "lock out the admins of the"},
			{"found a new network at the", "bring the whole city online at the", "raise a data commons in the", "launch a public ledger on the"},
		},
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
		Age:    "cyberpunk_age",
		Tiers:  [4]string{"Recon", "Heist", "Assault", "Arcology Raising"},
		Places: [6]string{"Chrome Market", "Neon Alley", "Rain Towers", "Static Docks", "Ash Arcade", "Black Plaza"},
		Objectives: [4][4]string{
			{"scout the streets of the", "tap the cameras of the", "tail a courier through the", "scan the rooftops above the"},
			{"lift a data core from the", "break into the vault of the", "steal prototypes from the", "ghost a shipment out of the"},
			{"storm the corporate tower at the", "assault the checkpoint at the", "burn the security grid of the", "take the gang's hideout under the"},
			{"raise a sealed arcology above the", "power a whole block in the", "wire a free city under the", "house the poor in a tower over the"},
		},
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
		Age:    "fusion_age",
		Tiers:  [4]string{"Sweep", "Salvage", "Assault", "Fusion City"},
		Places: [6]string{"Reactor Ring", "Slag Fields", "Coolant Lake", "Burned Quarter", "Flux Yard", "Bright Dome"},
		Objectives: [4][4]string{
			{"sweep the ruins around the", "scan the shielding of the", "check the coolant lines near the", "map the heat vents of the"},
			{"salvage cores from the", "pull superconductors out of the", "tow a hull from the", "strip a dead plant at the"},
			{"assault the rival plant at the", "breach the shield of the", "seize the control room of the", "cut power to the"},
			{"light a fusion city at the", "build a sun in a bottle for the", "pipe cheap power to the", "raise a tower of glass beside the"},
		},
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
		Age:    "space_age",
		Tiers:  [4]string{"Probe", "Orbital Mission", "Orbital Strike", "Colony Ship"},
		Places: [6]string{"Low Orbit", "Moon Shadow", "Red Plain", "Ring of Ice", "Dust Belt", "Sunward Station"},
		Objectives: [4][4]string{
			{"launch a probe past the", "scan the surface of the", "skim the dust of the", "map the craters of the"},
			{"fly an orbital mission around the", "dock with the wreck above the", "mine ice from the", "set a relay beyond the"},
			{"strike the enemy station above the", "knock out the satellite over the", "board the raider near the", "blockade the lanes of the"},
			{"send a colony ship toward the", "raise a ring habitat over the", "seed gardens on the", "build a lunar town in the"},
		},
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
		Age:    "interstellar_age",
		Tiers:  [4]string{"Probe", "Deep-Space Mission", "Invasion", "Generation Ship"},
		Places: [6]string{"Proxima Gate", "Cold Giant", "Long Dark", "Twin Suns", "Drift Station", "Ember Nebula"},
		Objectives: [4][4]string{
			{"launch a probe toward the", "scan the planets of the", "sample the dust of the", "listen for signals from the"},
			{"fly a deep-space mission to the", "build a relay chain toward the", "chart the lanes beyond the", "mine comet ice near the"},
			{"invade the colony at the", "land troops on a world of the", "take the orbital yards at the", "break the garrison at the"},
			{"launch a generation ship for the", "seed a new world near the", "carry a million sleepers to the", "raise a star colony by the"},
		},
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
		Age:    "galactic_age",
		Tiers:  [4]string{"Survey Drone", "Star Voyage", "Invasion", "Stellar Settlement"},
		Places: [6]string{"Orion Reach", "Crimson Rim", "Quiet Arm", "Core Gate", "Lantern Cluster", "Pale Spiral"},
		Objectives: [4][4]string{
			{"send a survey drone through the", "count the stars of the", "scan the worlds of the", "map the lanes around the"},
			{"take a star voyage across the", "open a jump lane through the", "trade across the", "carry seed ships to the"},
			{"invade the fleet at the", "storm the throneworld of the", "blockade the star gates of the", "drive the armada from the"},
			{"settle a thousand worlds in the", "raise a star ring around the", "found a stellar city in the", "gather a council at the"},
		},
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
		Age:    "quantum_age",
		Tiers:  [4]string{"Phase Probe", "Rift Expedition", "Incursion", "Pocket Universe"},
		Places: [6]string{"Fold Garden", "Mirror Gap", "Null Shore", "Still Point", "Hollow Spiral", "Echo Vault"},
		Objectives: [4][4]string{
			{"phase a probe through the", "measure the noise of the", "sample a mirror from the", "trace the edges of the"},
			{"lead a rift expedition into the", "map the layers of the", "bring back a ghost from the", "hold open the door to the"},
			{"launch an incursion into the", "collapse the guard of the", "slip a strike team into the", "unravel the defenses of the"},
			{"grow a pocket universe in the", "fold a new home inside the", "weave a stable bubble in the", "teach the rift to hold at the"},
		},
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
		Age:    "transcendent_age",
		Tiers:  [4]string{"Glimpse", "Pilgrimage of Light", "Reality Incursion", "Ascension Rite"},
		Places: [6]string{"First Dawn", "Silent Choir", "Open Horizon", "Garden of Hours", "Last Shore", "Shining Stair"},
		Objectives: [4][4]string{
			{"catch a glimpse of the", "listen at the edge of the", "follow a thread of light to the", "gather echoes from the"},
			{"walk the pilgrimage of light to the", "carry a spark across the", "sing a path through the", "trace the light back to the"},
			{"break into the reality of the", "bend the law of the", "storm the veil around the", "unmake the wall of the"},
			{"perform the ascension rite at the", "gather every voice at the", "open the final door at the", "raise the last choir at the"},
		},
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
