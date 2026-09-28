package main

// catalog.go is the authored prose. It follows the flavor package's contract:
// the unit is a finished sentence somebody wrote; a sentence has at most one
// slot ({x}), which only ever takes a noun phrase; nothing is joined clause to
// clause; output carries no markup; lengths are deliberately uneven and most
// lines are flat. Mechanics (counts, names, numbers) are printed separately, by
// the room's "Here:" block, never in these sentences.

// establish is the first sentence of a room: what the place IS in this epoch.
// Indexed [place][epoch]; epochs follow config.Epochs() order.
var establish = map[PlaceKey][7]string{
	Square: {
		"A patch of grass worn to dirt by people standing about on it.",
		"Paving stones, a speaking platform, and a great many opinions.",
		"The piazza is cobbled in a pattern the architect was very pleased with.",
		"A civic plaza under electric lamps that hum at a pitch only the dogs mind.",
		"The concourse is glass on three sides and advertising on the fourth.",
		"The atrium climbs forty storeys to a ceiling painted to look like weather.",
		"The ring curves up and away in both directions until it meets itself overhead.",
	},
	Homes: {
		"Huts, close together, for warmth and for gossip.",
		"Narrow lanes of houses lean toward each other over the street.",
		"Terraces of brick houses, each with a step somebody scrubs every morning.",
		"Row after row of identical houses, told apart by their curtains.",
		"Apartment blocks stand in a grid, each one lit a window at a time.",
		"Housing stacks rise in pods, bolted on wherever there was room for one more.",
		"The habitat rings turn slowly, and the people inside have stopped noticing.",
	},
	Fields: {
		"Berry scrub and digging ground, picked over by people who know which roots are safe.",
		"Strip fields run out from the walls, each strip farmed by somebody's cousin.",
		"Plantations stretch in straight rows to the horizon.",
		"The farm belt is flat, fenced, and loud with machinery at harvest.",
		"Climate-sealed agricultural domes cover what used to be the fields.",
		"The vat halls smell of yeast and warm plastic.",
		"The hydroponic bays glow a pink that is good for plants and bad for sleep.",
	},
	Woods: {
		"Old trees, and the stumps of trees that were closer to the huts.",
		"Timber yards stack cut logs in the order they were felled.",
		"Coal pits smoke along the hillside where the forest used to be.",
		"Derricks nod over the oil fields like birds that have found something.",
		"The refinery towers flare off whatever they cannot sell.",
		"Nanovats hum in rows, growing what the forests used to.",
		"The matter weavers spin feedstock out of dust and patience.",
	},
	Quarry: {
		"Shallow pits where the good flint comes up.",
		"The quarry face is cut in steps, each one a year of work.",
		"Mine heads and spoil heaps.",
		"The uranium works sit behind two fences and a sign nobody reads twice.",
		"Deep shafts go down further than anyone here has been.",
		"The crystal mines ring faintly when the drills stop.",
		"Tethered asteroids hang off the station on long cables, being eaten from the inside.",
	},
	Forge: {
		"Clay smelting pits, black at the rim.",
		"The forge quarter rings with hammers from first light.",
		"Foundries pour iron into sand under a roof of permanent smoke.",
		"Arc furnaces light the whole district blue when they strike.",
		"Alloy plants run clean and quiet, which the old smiths find suspicious.",
		"The dark refinery keeps its windows painted over.",
		"The star forge draws a thread of plasma from something very far away.",
	},
	Works: {
		"A smithy lane where anything broken gets taken to be looked at.",
		"Workshops and water channels, and a mill wheel that never quite stops.",
		"Mills with long belts running from one wheel to the next.",
		"Power yards full of transformers and men in rubber boots.",
		"The grid is a room full of cabinets, and all of them are warm.",
		"The augment foundry fits people with parts they did not know they needed.",
		"The launch complex is gantries, fuel lines, and a countdown that never reaches zero.",
	},
	Temple: {
		"A shrine of stacked stones with offerings left in the gaps.",
		"The temple mount is steep enough that nobody arrives at the top in a bad mood.",
		"Cathedral spires over a close of quiet lawns.",
		"Chapels line the road, each one a different shade of the same belief.",
		"Meditation gardens, raked gravel, and a waiting list.",
		"The sanctuary is lit in neon, and the prayers are logged.",
		"The cloister faces the void through a window a kilometre wide.",
	},
	Academy: {
		"A ring of logs around a fire where the old people tell it the way it happened.",
		"Library steps, where scholars argue in the sun and read in the shade.",
		"The university has four courtyards and one opinion per student.",
		"The institute's windows are full of chalkboards and the chalkboards are full.",
		"The campus has a lawn nobody is allowed to walk on.",
		"The server halls are cold on purpose.",
		"The observatory deck looks out at stars that have been catalogued twice.",
	},
	Market: {
		"Flat stones where people lay out what they have and point at what they want.",
		"Guild row is all signboards and scales.",
		"The exchange floor shouts all day and whispers after dark.",
		"The financial district builds tall to look down on the rest of the city.",
		"The venture quarter is glass, coffee, and pitch decks.",
		"The night market sells what the day market will not admit to.",
		"The bazaar ring trades in things that have not been invented yet.",
	},
	Harbour: {
		"A pebble landing where the canoes are pulled up.",
		"The quay is stacked with barrels and the gulls are organised.",
		"The harbour is a forest of masts in a calm year.",
		"Cranes and warehouses line the docks, and the water tastes of coal.",
		"Container stacks, colour-coded, moved by machines that do not look up.",
		"The logistics hub routes cargo it will never see.",
		"The docking spindle holds ships at arm's length and takes their paperwork.",
	},
	Stores: {
		"Stash pits, lined with bark and covered with a flat rock.",
		"Granaries on stilts, out of reach of the rats, mostly.",
		"Warehouses, numbered, with a clerk at each door.",
		"The depot runs on a railway siding and a ledger the size of a door.",
		"The archive vaults keep everything, including the things they were asked to delete.",
		"Cyber vaults, air-gapped, guarded by a man who is also air-gapped.",
		"The orbital depots hang in a slow line, each one a warehouse with its own gravity.",
	},
	Barracks: {
		"A war camp: spears in a rack and a fire nobody lets go out.",
		"The barracks yard is swept daily, by the recruits who made it dirty.",
		"The fort sits low behind earthworks, pointed at nothing in particular.",
		"The garrison drills on a parade ground painted with white lines.",
		"Command post antennae on a roof, and nobody allowed on the roof.",
		"The drone pens buzz like a hive that has read the manual.",
		"The fleet yards berth warships too large to see all at once.",
	},
	Wonders: {
		"A rise of ground the whole settlement agrees is special.",
		"The wonder walk is lined with the things your people built to be remembered.",
		"The grand promenade runs past every monument you have, in the order you built them.",
		"The exposition grounds were laid out for a fair and never taken down.",
		"The monument mall is a long lawn with a great deal of history at either end.",
		"The citadel steps climb past the old wonders to the new one.",
		"The beacon spire rises from the ring into nothing.",
	},
	Gate: {
		"The track out of the settlement, trodden flat by everyone who ever left.",
		"The town gate stands open, with a guard who knows everyone's business.",
		"The toll road runs out between milestones and a toll-keeper's hut.",
		"The railway station has a clock everybody sets their watch by.",
		"The interchange ties six roads into a knot and lets the traffic sort it out.",
		"The skyport lands anything with wings and several things without.",
		"The jump gate sits dark between departures, a ring of nothing held open.",
	},
}

// pathLines describe a place that has nothing in it yet but the road.
var pathLines = []string{
	"Nothing is built here yet; the road just runs through.",
	"Open ground. Somebody has paced it out and gone away again.",
	"A bare stretch between the places people actually go.",
	"Weeds, a surveyor's stake, and the road.",
}

// density: how full the place is, relative to how much it could hold.
var density = map[string][]string{
	"sparse": {
		"It is quiet.",
		"There is room here for a great deal more.",
		"A single path does for the whole place.",
		"Most of the ground is still waiting for a reason.",
	},
	"busy": {
		"People cross it with somewhere to be.",
		"It is busy in the ordinary way.",
		"Every building here has a neighbour close enough to borrow from.",
		"The traffic has worn a groove down the middle.",
	},
	"crowded": {
		"You have to turn sideways to get through.",
		"It is packed, and the newest buildings have been fitted into the gaps between the old.",
		"There is no more room, and they are building anyway.",
		"Nobody here has seen the ground in years.",
	},
}

// Condition sentences. {x} is the only slot; it takes a noun phrase.
var (
	lineUnderstaffed = []string{
		"Half the benches are empty; there are not enough hands for the work.",
		"Tools lie where they were put down, waiting for somebody to pick them up.",
		"Some doors here are shut in the middle of the day.",
	}
	lineStaffed = []string{
		"Every station is taken.",
		"Nobody here is standing still.",
		"The work goes on in shifts, and the shifts overlap.",
	}
	lineLegacy = []string{
		"The old {x} is still in use, and nobody has had the heart to pull it down.",
		"Between the new buildings stands the old {x}, doing what it always did.",
		"The {x} dates from an older age and looks it.",
	}
	lineRuins = []string{
		"The shell of the old {x} still stands where the catastrophe left it.",
		"Scorched footings mark where the {x} used to be.",
		"The ruined {x} is used for storage now, and for courage.",
	}
	lineNew = []string{
		"The {x} is new; the scaffolding is still leaning against it.",
		"People keep stopping to look at the {x}.",
		"Fresh sawdust and mortar around the {x}.",
	}
	lineWar = []string{
		"The road is barricaded, and the watch counts everyone twice.",
		"Carts come in from the road faster than they go out.",
		"Sentries walk the line with their eyes on the horizon.",
	}
	linePeace = []string{
		"Travellers come and go without being asked their business.",
		"The road is open and nobody is watching it very hard.",
	}
	lineExpedition = []string{
		"The {x} left by this road and has not come back yet.",
		"Fresh tracks lead out, the ones the {x} made.",
	}
	lineRouteBusy = []string{
		"Crates stamped for the {x} are stacked waiting for the next sailing.",
		"The {x} keeps the cranes busy.",
	}
	lineRouteBlocked = []string{
		"The {x} has not come in, and the crates for it are gathering dust.",
	}
	lineIdle = []string{
		"Idle hands sit on the steps waiting to be told what to do.",
		"A crowd with nothing to do has gathered, as crowds do.",
	}
	lineMoraleHigh = []string{
		"Somebody is singing.",
		"People greet each other by name.",
	}
	lineMoraleLow = []string{
		"People keep their heads down and their voices lower.",
		"There is a sullen quiet over the place.",
	}
	lineHarbinger = []string{
		"{x} stands in the middle of it all, and people walk wide around.",
		"{x} has taken the best spot, and nobody has asked for it back.",
	}
	lineWonderBuilding = []string{
		"Scaffolding climbs the {x}, which is not finished and already famous.",
		"The {x} is going up, one argued-over stone at a time.",
	}
	lineReady = []string{
		"Everyone seems to be looking at the road out.",
		"There is a feeling of packed bags about the place.",
	}
	lineEvent = []string{
		"Everyone here is talking about the {x}.",
		"The {x} has put the place in a strange mood.",
	}
)

// sky: time of day, per register (0 ancient, 1 industrial, 2 future). The
// cosmic register replaces day with the station's light cycle.
var sky = map[string][3]string{
	"dawn":      {"The sun is just up and the smoke from the cook fires goes straight up.", "Dawn, and the lamplighters are putting the lamps out.", "Morning cycle; the lights come up a shade at a time."},
	"morning":   {"Morning. The shadows are long and pointing west.", "A grey morning with the whistles going.", "Mid-morning, by the clocks."},
	"noon":      {"Noon, and the heat has sent everyone sensible into the shade.", "Noon. The bells and the whistles argue about it.", "Noon, if anyone still checks."},
	"afternoon": {"The afternoon has gone slow and golden.", "Afternoon, and the smoke has settled into the streets.", "Afternoon shift."},
	"dusk":      {"Dusk. Somebody is calling children in for the night.", "Dusk, and the gas lamps come on one at a time down the street.", "Evening cycle; the signs outshine the sky."},
	"night":     {"Night. The fires are banked and the stars are very close.", "Night, and the lit windows are the only way to tell where the streets are.", "Night cycle. The city does not sleep so much as dim."},
}

var weather = map[string][3]string{
	"clear":    {"The sky is clear.", "A rare clear day.", "Visibility is good."},
	"cloud":    {"Clouds are coming in from the west.", "The sky is the colour of the chimneys.", "Low cloud sits on the tall towers."},
	"rain":     {"It is raining, lightly and without conviction.", "It is raining soot.", "Rain, and every sign reflected in it twice."},
	"wind":     {"A hard wind is pulling at the thatch.", "The wind carries the smell of the works across the whole town.", "Wind in the canyons between the towers."},
	"fog":      {"Fog has come up from the river.", "A yellow fog you could lean on.", "Fog, and the drones are flying on instruments."},
	"dry":      {"The ground is cracked; it has not rained in weeks.", "Dust on everything; the drought has not broken.", "The air is dry enough to crack lips."},
	"solar":    {"", "", "A solar storm has the hull lights flickering."},
	"quiet":    {"", "", "Outside the windows, nothing, in every direction."},
	"meteor":   {"", "", "A meteor shower drags bright lines across the dark side of the ring."},
	"festival": {"Bunting hangs across the paths.", "Bunting hangs across the street.", "Somebody has programmed the lights for a festival."},
}
