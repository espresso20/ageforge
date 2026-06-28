package game

import (
	"math/rand"
	"strings"
)

// GenerateSaveName returns a procedural, human-friendly name for a new save
// slot. It picks one of three style families at random (epic / mythic /
// whimsical) and assembles a name from curated word banks.
//
// The result is always FILESYSTEM-SAFE: it becomes a save filename
// (<name>.json), so it contains letters and single spaces only — never an
// apostrophe, slash, dot, colon, quote, or other path-hostile character.
// Mythic names therefore use no apostrophes. Output is trimmed and has
// internal whitespace collapsed to single spaces.
//
// Names are cosmetic and MAY repeat across calls — there is no uniqueness
// guarantee. We use math/rand's global source (auto-seeded by modern Go),
// so callers do not need to seed anything.
func GenerateSaveName() string {
	switch rand.Intn(3) {
	case 0:
		return cleanSaveName(generateEpicName())
	case 1:
		return cleanSaveName(generateMythicName())
	default:
		return cleanSaveName(generateWhimsicalName())
	}
}

// cleanSaveName collapses internal whitespace to single spaces and trims the
// ends, guaranteeing the no-leading/trailing-space invariant the word-bank
// assembly relies on. The banks themselves never introduce path-hostile
// characters, so a strings.Fields round-trip is all the sanitising we need.
func cleanSaveName(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// --- EPIC: grand, dominion-flavoured. "The <Adj> <Noun>" / "<Root> <Suffix>".

var epicAdjectives = []string{
	"Gilded", "Verdant", "Azure", "Eternal", "Radiant", "Sovereign",
	"Ironclad", "Crimson", "Golden", "Boundless", "Towering", "Glorious",
	"Ascendant", "Resolute", "Imperial", "Hallowed",
	"Obsidian", "Argent", "Stalwart", "Undaunted", "Resplendent", "Granite",
	"Tempered", "Unbroken", "Stormborn", "Sunlit", "Vigilant", "Adamant",
	"Regal", "Lustrous", "Indomitable", "Valiant", "Majestic", "Triumphant",
	"Stoic", "Unyielding", "Dauntless", "Splendid", "Lofty", "Steadfast",
	"Burnished", "Emerald", "Sapphire", "Scarlet", "Vermilion", "Cobalt",
	"Auric", "Platinum", "Diamond", "Marble", "Bronze", "Silvered",
	"Crowned", "Ennobled", "Exalted", "Venerable", "August", "Illustrious",
	"Storied", "Fabled", "Legendary", "Renowned", "Vaunted", "Honored",
	"Bannered", "Plumed", "Mantled", "Sceptered", "Throned", "Halcyon",
	"Serene", "Tranquil", "Resounding", "Thunderous", "Roaring", "Soaring",
	"Cresting", "Surging", "Blazing", "Smoldering", "Kindled", "Flaring",
	"Frostbound", "Snowcrowned", "Glacial", "Windswept", "Stormwrought", "Sunward",
	"Dawnbright", "Twilit", "Starbound", "Moonlit", "Skyborne", "Highborn",
	"Stonefast", "Ironwrought", "Steelbound", "Anvilforged", "Hammerhewn", "Keenedged",
	"Fortified", "Embattled", "Warbound", "Shieldsworn", "Spearheaded", "Vanguarded",
	"Wardful", "Watchful", "Sentinel", "Guardian", "Bulwarked", "Ramparted",
	"Verdured", "Florid", "Bountiful", "Plentiful", "Abundant", "Fertile",
	"Flourishing", "Thriving", "Burgeoning", "Prospering", "Wealthy", "Opulent",
	"Lavish", "Sumptuous", "Bejeweled", "Ornate", "Adorned", "Festooned",
	"Hallowing", "Sanctified", "Blessed", "Consecrated", "Divine", "Celestial",
	"Empyrean", "Seraphic", "Beatific", "Luminous", "Effulgent", "Incandescent",
	"Brilliant", "Dazzling", "Shimmering", "Gleaming", "Glistening", "Glinting",
	"Mighty", "Puissant", "Formidable", "Redoubtable", "Commanding", "Dominant",
	"Paramount", "Supreme", "Foremost", "Preeminent", "Sovereignborn", "Crownward",
}

var epicNouns = []string{
	"Dominion", "Ascendancy", "Reach", "Compact", "Hegemony", "Empire",
	"Realm", "Bastion", "Sovereignty", "Accord", "Expanse", "Citadel",
	"Crown", "Vanguard", "Concord", "Mandate",
	"Protectorate", "Throne", "Aegis", "Pinnacle", "Stronghold", "Conclave",
	"Reign", "Dynasty", "Banner", "Keep", "Demesne", "Suzerainty",
	"Imperium", "Commonweal", "Federation", "Confederation", "Coalition", "Union",
	"Principality", "Margravate", "Marquisate", "Duchy", "Earldom", "Barony",
	"Palatinate", "Electorate", "Khanate", "Sultanate", "Caliphate", "Emirate",
	"Shogunate", "Magistracy", "Triumvirate", "Tetrarchy", "Oligarchy", "Synod",
	"Diet", "Senate", "Tribunal", "Curia", "Chancery", "Exarchate",
	"Stewardship", "Wardenry", "Marchland", "Borderland", "Frontier", "Holdfast",
	"Garrison", "Redoubt", "Fastness", "Rampart", "Bulwark", "Battlement",
	"Spire", "Tower", "Hold", "Hall", "Sanctum", "Sanctuary",
	"Cathedral", "Temple", "Shrine", "Reliquary", "Vault", "Treasury",
	"Foundry", "Forge", "Anvil", "Bastille", "Fortress", "Castellany",
	"Dominium", "Patrimony", "Legacy", "Inheritance", "Estate", "Seignory",
	"Concordat", "Covenant", "Charter", "Pact", "Treaty", "League",
	"Assembly", "Parliament", "Council", "Court", "Cabinet", "Bench",
	"Ascension", "Ascendance", "Sovereign", "Monarchy", "Kingdom", "Queendom",
	"Empery", "Realmhold", "Crownland", "Throneward", "Diadem", "Coronet",
	"Mantle", "Scepter", "Orb", "Regalia", "Standard", "Ensign",
	"Pennant", "Gonfalon", "Oriflamme", "Vexillum", "Eagle", "Lion",
	"Wyvern", "Griffon", "Phoenix", "Sentinel", "Warden", "Guardian",
	"Citadelle", "Acropolis", "Palace", "Manse", "Chateau", "Donjon",
	"Bailey", "Motte", "Barbican", "Gatehouse", "Watchtower", "Beacon",
	"Pillar", "Obelisk", "Monolith", "Colossus", "Bastionry", "Stronghall",
}

var epicPlaceRoots = []string{
	"Ironhold", "Sunspire", "Stormhall", "Goldmere", "Highreach", "Frosthold",
	"Emberfall", "Greywatch", "Brightspire", "Stonereach", "Dawnhold", "Ashmere",
	"Westmarch", "Thornwall", "Silvermoor", "Oakenford",
	"Frostpeak", "Emberforge", "Highcrag", "Dawnwatch", "Shadowfen", "Brightwater",
	"Thornkeep", "Oakenshield", "Silverpeak", "Wyrmrest", "Grimward", "Stormwatch",
	"Ironforge", "Goldspire", "Sunhaven", "Moonhold", "Starfall", "Skyhold",
	"Cloudreach", "Windhelm", "Stonehelm", "Ironhelm", "Steelkeep", "Bronzegate",
	"Silvergate", "Goldgate", "Irongate", "Stormgate", "Frostgate", "Emberkeep",
	"Ashenhold", "Cinderfall", "Flameheart", "Ashcroft", "Coalmere", "Smokeford",
	"Mistvale", "Foghollow", "Greymantle", "Greyhollow", "Duskwatch", "Nightwatch",
	"Twilighthold", "Dawnbreak", "Daybreak", "Sunbreak", "Morrowmere", "Eventide",
	"Westhaven", "Easthold", "Northwatch", "Southreach", "Farreach", "Longmarch",
	"Stillwater", "Clearwater", "Deepwater", "Whitewater", "Blackwater", "Bluewater",
	"Greenhollow", "Greenmere", "Greenvale", "Springvale", "Summervale", "Autumnvale",
	"Winterhold", "Snowhold", "Glacierhall", "Frostmoor", "Frostvale", "Coldcrag",
	"Highmoor", "Lowmoor", "Eastmoor", "Westmoor", "Fellmoor", "Bleakmoor",
	"Thornvale", "Briarhold", "Bramblewatch", "Hawthorne", "Elderwood", "Wildwood",
	"Oakheart", "Pinehold", "Cedarfall", "Ashwood", "Elmhollow", "Birchvale",
	"Stonewatch", "Stonegate", "Stonehaven", "Granitehold", "Marblehall", "Slatecrag",
	"Crystalspire", "Diamondhold", "Rubyfall", "Sapphirereach", "Emeraldvale", "Pearlhaven",
	"Wyrmwatch", "Drakehold", "Dragonreach", "Serpentmere", "Griffongate", "Phoenixfall",
	"Lionhold", "Eaglespire", "Ravenmoor", "Wolfden", "Bearclaw", "Stagheart",
	"Warhold", "Battleford", "Bloodkeep", "Ironclad", "Shieldwall", "Spearhold",
	"Banneret", "Bastionford", "Rampartmere", "Bulwarkhold", "Garrisonvale", "Redoubtwatch",
	"Hallowmere", "Sanctumvale", "Templehold", "Shrineford", "Sacredmoor", "Holyhaven",
	"Crownvale", "Throneford", "Regalhold", "Imperialspire", "Sovereignmere", "Royalwatch",
}

var epicSuffixes = []string{
	"Ascendancy", "Hegemony", "Dominion", "Reach", "Compact", "Accord",
	"Realm", "Empire", "Sovereignty", "Mandate", "Expanse", "Crown",
	"Protectorate", "Dominium", "Imperium", "Suzerainty", "Commonweal", "Concordat",
	"Confederation", "Federation", "Coalition", "Union", "Principality", "Demesne",
	"Marchland", "Borderland", "Frontier", "Holdfast", "Stewardship", "Wardenry",
	"Patrimony", "Seignory", "Palatinate", "Electorate", "Khanate", "Sultanate",
	"Caliphate", "Emirate", "Shogunate", "Margravate", "Marquisate", "Duchy",
	"Earldom", "Barony", "Triumvirate", "Tetrarchy", "Exarchate", "Magistracy",
	"Covenant", "Charter", "Pact", "League", "Treaty", "Alliance",
	"Synod", "Conclave", "Senate", "Tribunal", "Curia", "Diet",
	"Throne", "Reign", "Dynasty", "Monarchy", "Kingdom", "Queendom",
	"Empery", "Diadem", "Coronet", "Scepter", "Standard", "Banner",
	"Aegis", "Bastion", "Citadel", "Stronghold", "Fastness", "Redoubt",
	"Bulwark", "Rampart", "Garrison", "Keep", "Fortress", "Hold",
	"Sanctum", "Sanctuary", "Reliquary", "Cathedral", "Temple", "Shrine",
	"Spire", "Pinnacle", "Tower", "Hall", "Estate", "Manse",
	"Legacy", "Inheritance", "Lineage", "Bloodline", "Heritage", "Birthright",
	"Ascension", "Ascendance", "Dominance", "Supremacy", "Primacy", "Paramountcy",
	"Vanguard", "Sentinelship", "Wardship", "Watch", "Vigil", "Guard",
	"Mandatory", "Sovereignborn", "Crownland", "Realmhold", "Empireborn", "Thronehold",
	"Concord", "Entente", "Compactum", "Bond", "Oath", "Fealty",
	"Realmward", "Reachward", "Crownward", "Throneward", "Empireward", "Dominionward",
}

func generateEpicName() string {
	if rand.Intn(2) == 0 {
		return "The " + pick(epicAdjectives) + " " + pick(epicNouns)
	}
	return pick(epicPlaceRoots) + " " + pick(epicSuffixes)
}

// --- MYTHIC: coined ancient-sounding names. No apostrophes (filesystem-safe).

var mythicRoots = []string{
	"Vaelthar", "Kael", "Drovia", "Zhanggar", "Obsidian", "Aeon",
	"Myrr", "Thalor", "Eldros", "Korvath", "Nyxara", "Velmara",
	"Azhul", "Sorrowmere", "Caldris", "Ophis",
	"Azgaroth", "Velmoria", "Sythar", "Druumel", "Khaledon", "Ysmir",
	"Tarnaxis", "Oromel", "Vexar", "Quenthel", "Balgrym", "Theronax",
	"Maeldroth", "Vorhal", "Zephyrian", "Kaltheon", "Nymerion", "Sarvex",
	"Drathel", "Volmarr", "Ithragor", "Pyrethon", "Galmoreth", "Selvaria",
	"Morvath", "Caelum", "Tholrun", "Vandros", "Xerathon", "Olythar",
	"Brunmael", "Cendreth", "Dolmaris", "Eskandar", "Faltheon", "Grimvar",
	"Halveth", "Jorundel", "Khazros", "Lumeneth", "Mordreth", "Norvael",
	"Othrandil", "Praximar", "Quorvath", "Ravelune", "Solmera", "Talvoros",
	"Ulthane", "Varethon", "Wyldros", "Xalnareth", "Ymberon", "Zandrael",
	"Aldoreth", "Belmora", "Calderon", "Dravonis", "Emberthal", "Forlanis",
	"Garveth", "Helvanis", "Isgaroth", "Jeralune", "Kelthros", "Lovaris",
	"Maerlyn", "Nessoth", "Orvandel", "Phyrra", "Quelmir", "Rovathar",
	"Sundreth", "Therrak", "Ulvaris", "Veythar", "Wraithmoor", "Xenophar",
	"Yndaril", "Zorvath", "Aethros", "Branneth", "Corvethis", "Daromir",
	"Elnareth", "Fenrothi", "Gloamreth", "Hesperan", "Ilvareth", "Jandros",
	"Kovenar", "Lytheris", "Mournveil", "Nyxabar", "Oberyth", "Pendralis",
	"Quivaros", "Reldanis", "Sorvethel", "Thelmari", "Uveloth", "Vandareth",
	"Wystoria", "Xanthros", "Yveloth", "Zephanis", "Almareth", "Brimnor",
	"Cindaros", "Drennith", "Elgaroth", "Fyrelune", "Gravethor", "Halmaris",
	"Ironvael", "Jutheris", "Kaldemar", "Lorvanis", "Myrethon", "Nolvaris",
	"Ostrandel", "Pelloria", "Quenmaris", "Ralvethon", "Sythrael", "Tormalis",
	"Ulvendra", "Voskareth", "Werlathan", "Xolmira", "Ythanor", "Zelvaros",
}

var mythicConnectors = []string{
	"", "", "Drovia", "Vey", "Anar", "Mor", "Sael", "Thun",
	"Vora", "Eth", "Kah", "Ulm",
	"", "", "", "", "Dra", "Syl", "Bral", "Nor",
	"Zha", "Vael", "Tor", "Mel", "Gar", "Lor", "Pyr", "Ash",
	"Vor", "Drel", "Mae", "Sol", "Ther", "Vaen", "Kor", "Brin",
	"Dun", "Fal", "Gol", "Hal", "Jor", "Kel", "Lun", "Myr",
	"Nyl", "Oth", "Pel", "Quor", "Rav", "Sed", "Thal", "Ulv",
	"Ven", "Wyl", "Xan", "Yth", "Zor", "Alm", "Brim", "Cind",
	"Dren", "Elg", "Fyr", "Grav", "Halm", "Iron", "Jut", "Kald",
	"Lorv", "Myre", "Nolv", "Ostr", "Pell", "Quen", "Ralv", "Syth",
	"Torm", "Ulve", "Vosk", "Werl", "Xolm", "Ythan", "Zelv", "Aeth",
	"Bran", "Corv", "Dar", "Eln", "Fen", "Gloam", "Hesp", "Ilv",
	"Jand", "Kov", "Lyth", "Mourn", "Nyx", "Ober", "Pend", "Quiv",
	"Reld", "Sorv", "Thel", "Uvel", "Vand", "Wyst", "Xanth", "Yvel",
	"Zeph", "Aldo", "Belm", "Cald", "Drav", "Emb", "Forl", "Garv",
	"Helv", "Isg", "Jer", "Kelt", "Lov", "Maer", "Ness", "Orv",
}

var mythicSuffixes = []string{
	"um", "ia", "or", "eth", "an", "is", "ar", "yx",
	"ondor", "athar", "evar", "umir",
	"oth", "ael", "ys", "orin", "andril", "esh", "iel", "oria",
	"andor", "emar", "uvis", "elor", "anth", "ius", "eron", "alis",
	"oren", "yth", "asha", "endil", "ovar", "amir", "ethis", "oril",
	"ungar", "elis", "andra", "osha", "evik", "anor", "uvar", "ethel",
	"orum", "alon", "irith", "estra", "anos", "ydon", "elith", "ovan",
	"arion", "essa", "ulon", "andros", "ethar", "ivion", "oryl", "umbra",
	"avos", "endra", "ilith", "oron", "ascar", "emir", "uval", "anix",
	"orveth", "alia", "endral", "yssa", "omir", "athon", "elvar", "ondria",
	"avel", "ureth", "oryn", "anthe", "iros", "elune", "ovith", "armel",
	"endis", "ulis", "ariax", "ethon", "ovara", "anel", "irix", "oloth",
	"aveth", "endor", "ulara", "ymir", "ostra", "alune", "evos", "anira",
	"oreth", "elvis", "uvel", "ynia", "omar", "athel", "ervos", "ondil",
	"avia", "umbral", "orix", "aneth", "ilos", "estor", "uvian", "armion",
}

var mythicAdjectives = []string{
	"Obsidian", "Eternal", "Forgotten", "Shrouded", "Ancient", "Hollow",
	"Sunken", "Veiled", "Silent", "Endless", "First", "Pale",
	"Sundered", "Weeping", "Nameless", "Drowned", "Riven", "Fading",
	"Starless", "Moonless", "Sunless", "Lightless", "Shadowed", "Darkened",
	"Whispering", "Murmuring", "Sighing", "Mourning", "Grieving", "Sorrowing",
	"Withered", "Wilted", "Decaying", "Crumbling", "Ruined", "Shattered",
	"Broken", "Fractured", "Splintered", "Cracked", "Severed", "Cleaved",
	"Cursed", "Doomed", "Damned", "Forsaken", "Abandoned", "Desolate",
	"Barren", "Bleak", "Lonely", "Solitary", "Wandering", "Lost",
	"Buried", "Entombed", "Sepulchral", "Funereal", "Ghostly", "Spectral",
	"Wraithlike", "Phantom", "Haunted", "Hallowed", "Unhallowed", "Unmarked",
	"Wordless", "Voiceless", "Soundless", "Muted", "Hushed", "Stilled",
	"Frozen", "Glacial", "Frostbitten", "Icebound", "Snowbound", "Wintering",
	"Smoldering", "Charred", "Ashen", "Cindered", "Scorched", "Blackened",
	"Drowning", "Sinking", "Submerged", "Tidebound", "Stormwracked", "Windworn",
	"Timeworn", "Ageworn", "Weathered", "Eroded", "Faded", "Dimming",
	"Vanishing", "Dwindling", "Waning", "Dying", "Slumbering", "Sleeping",
	"Dreaming", "Drowsing", "Stirring", "Waking", "Risen", "Returning",
	"Undying", "Deathless", "Timeless", "Ageless", "Immortal", "Everlasting",
	"Primordial", "Primeval", "Antediluvian", "Olden", "Elder", "Foremost",
	"Hidden", "Secret", "Concealed", "Cloaked", "Masked", "Enshadowed",
	"Twilit", "Gloaming", "Duskbound", "Nightbound", "Starbound", "Voidtouched",
}

var mythicEpochNouns = []string{
	"Aeon", "Epoch", "Dominion", "Throne", "Expanse", "Covenant",
	"Sanctum", "Vault", "Spire", "Reach", "Hollow", "Marches",
	"Threshold", "Requiem", "Vigil", "Sepulchre", "Bastion", "Abyss",
	"Eon", "Era", "Age", "Cycle", "Dawn", "Dusk",
	"Twilight", "Gloaming", "Nightfall", "Daybreak", "Eventide", "Midnight",
	"Aurora", "Eclipse", "Equinox", "Solstice", "Zenith", "Nadir",
	"Apex", "Pinnacle", "Summit", "Crest", "Verge", "Brink",
	"Precipice", "Chasm", "Rift", "Void", "Maelstrom", "Tempest",
	"Cataclysm", "Reckoning", "Sundering", "Unmaking", "Awakening", "Ascension",
	"Communion", "Sacrament", "Litany", "Lament", "Dirge", "Threnody",
	"Elegy", "Psalm", "Canticle", "Liturgy", "Rite", "Ritual",
	"Mystery", "Enigma", "Riddle", "Secret", "Cipher", "Oracle",
	"Prophecy", "Augury", "Omen", "Portent", "Vision", "Revelation",
	"Sanctuary", "Refuge", "Haven", "Asylum", "Cloister", "Hermitage",
	"Reliquary", "Ossuary", "Catacomb", "Crypt", "Mausoleum", "Barrow",
	"Cairn", "Monument", "Memorial", "Monolith", "Obelisk", "Pillar",
	"Citadel", "Fortress", "Keep", "Tower", "Spireward", "Watchtower",
	"Beacon", "Lighthouse", "Pharos", "Lantern", "Flame", "Ember",
	"Wellspring", "Fountainhead", "Source", "Origin", "Genesis", "Inception",
	"Terminus", "Finality", "Ending", "Conclusion", "Culmination", "Apotheosis",
	"Marchland", "Borderland", "Wasteland", "Hinterland", "Frontier", "Periphery",
	"Demesne", "Domain", "Province", "Territory", "Realm", "Empire",
}

func generateMythicName() string {
	switch rand.Intn(3) {
	case 0:
		// Coined single word: <Root><connector><suffix>
		return pick(mythicRoots) + pick(mythicConnectors) + pick(mythicSuffixes)
	case 1:
		// Two coined words: <Root> <Root+suffix>
		return pick(mythicRoots) + " " + pick(mythicRoots) + pick(mythicSuffixes)
	default:
		// "The <Adj> <EpochNoun>" / "<Adj> <Root>"
		if rand.Intn(2) == 0 {
			return "The " + pick(mythicAdjectives) + " " + pick(mythicEpochNouns)
		}
		return pick(mythicAdjectives) + " " + pick(mythicRoots)
	}
}

// --- WHIMSICAL: playful, low-stakes. "The <Adj> <Noun>" / "Grand Duchy of ...".

var whimsicalWholeNames = []string{
	"Cluckington", "Snoozeburg", "Wobbleton", "Crumbshire", "Mittensgard",
	"Biscuitania", "Fluffhaven", "Pottersville", "Snackopolis", "Lazyfields",
	"Bumbleshire", "Quackmoor",
	"Snorington", "Puddlemarsh", "Grumbletown", "Wafflestead", "Drowsyvale", "Nibbleton",
	"Snickerdoodle", "Bumblewick", "Crumpetshire", "Doodleburg", "Snugglebottom", "Wigglesworth",
	"Pickletown", "Muffinford", "Toastington", "Buttonwood", "Dimplewick", "Fizzleburg",
	"Gigglemoor", "Hobnobbington", "Jellyford", "Kettleton", "Lollygag", "Marmaladeshire",
	"Noodlewick", "Oodlesburg", "Pumpernickel", "Quibbleton", "Rumblestead", "Squigglesworth",
	"Tumbleweed", "Umbridge", "Vexington", "Whifflemoor", "Yawnington", "Zigzagburg",
	"Bobblehead", "Chucklevale", "Dawdleburg", "Egglestone", "Fumbleford", "Gobbledygook",
	"Higgledy", "Jinglewick", "Knickknack", "Lumpington", "Mumbleshire", "Nuzzleburg",
	"Ploppleton", "Quirkwood", "Rambleton", "Scuttleburg", "Tickleshire", "Wugglemoor",
	"Bouncington", "Clumsyvale", "Dillydale", "Flopperton", "Gargleburg", "Hiccupshire",
	"Jumbleton", "Kerfufflewick", "Lollipopolis", "Moochville", "Nappington", "Plumpkin",
	"Snoozleton", "Tootleburg", "Waddlemoor", "Yodelshire", "Bamboozle", "Cuddleton",
	"Dunderwick", "Flibbertigibbet", "Grovelmoor", "Hodgepodge", "Jamboree", "Kibbleton",
	"Lazybones", "Munchington", "Noshville", "Piddleton", "Quaffmoor", "Snorewick",
	"Trundleburg", "Wobblegate", "Yummingdale", "Burpleton", "Chuckleburg", "Dozington",
	"Frumpwick", "Glumshire", "Hobbleton", "Jollyburg", "Kookville", "Lumberton",
	"Moodleshire", "Nubbington", "Plonkmoor", "Quirkburg", "Snortleton", "Twitchwick",
	"Wugglesburg", "Yawnmoor", "Bingleton", "Crinklewick", "Dribbleshire", "Fopperton",
	"Gloopville", "Hufflemoor", "Jibberburg", "Knobbington", "Loafshire", "Mopington",
}

var whimsicalAdjectives = []string{
	"Potato", "Mildly Cross", "Reluctant", "Comfy", "Sleepy", "Grumpy",
	"Snacking", "Wobbly", "Indignant", "Cozy", "Disgruntled", "Cheerful",
	"Vaguely Annoyed", "Perpetually Tired", "Mostly Harmless", "Slightly Confused", "Faintly Amused", "Quietly Smug",
	"Drowsy", "Peckish", "Cranky", "Bouncy", "Squishy", "Fluffy",
	"Wiggly", "Jiggly", "Floppy", "Lumpy", "Bumpy", "Clumsy",
	"Goofy", "Silly", "Loopy", "Dizzy", "Woozy", "Snoozy",
	"Mildly Confused", "Vaguely Hungry", "Slightly Damp", "Mostly Asleep", "Faintly Hopeful", "Quietly Determined",
	"Reasonably Calm", "Cautiously Optimistic", "Pleasantly Surprised", "Thoroughly Distracted", "Gently Baffled", "Politely Stubborn",
	"Toasty", "Munchy", "Crunchy", "Nibbly", "Doughy", "Buttery",
	"Pudgy", "Plump", "Roly", "Chubby", "Tubby", "Stout",
	"Yawning", "Dozing", "Lounging", "Loafing", "Dawdling", "Idling",
	"Grumbly", "Mumbly", "Bumbly", "Fumbly", "Stumbly", "Tumbly",
	"Snuggly", "Cuddly", "Huggable", "Squeezable", "Squashy", "Mushy",
	"Befuddled", "Bemused", "Bewildered", "Flustered", "Frazzled", "Discombobulated",
	"Hesitant", "Bashful", "Timid", "Sheepish", "Coy", "Demure",
	"Persnickety", "Finicky", "Fussy", "Particular", "Choosy", "Picky",
	"Whimsical", "Wistful", "Dreamy", "Daydreaming", "Absentminded", "Forgetful",
	"Mildly Peeved", "Quietly Hopeful", "Vaguely Suspicious", "Slightly Bored", "Mostly Content", "Faintly Curious",
	"Unbothered", "Unflappable", "Easygoing", "Laidback", "Mellow", "Chill",
	"Befuddledish", "Wobblesome", "Snoozeworthy", "Grouchy", "Chipper", "Jaunty",
}

var whimsicalNouns = []string{
	"Kingdom", "Empire", "Republic", "Federation", "Collective", "Commonwealth",
	"Confederacy", "League", "Syndicate", "Assembly", "Council", "Order",
	"Dominion", "Brigade", "Coalition", "Alliance", "Union", "Society",
	"Guild", "Fellowship", "Brotherhood", "Sisterhood", "Company", "Cohort",
	"Congress", "Parliament", "Senate", "Cabinet", "Committee", "Caucus",
	"Clan", "Tribe", "Horde", "Legion", "Battalion", "Regiment",
	"Squadron", "Platoon", "Garrison", "Militia", "Crew", "Posse",
	"Gang", "Crowd", "Bunch", "Bevy", "Flock", "Herd",
	"Pack", "Swarm", "Throng", "Cluster", "Gathering", "Conclave",
	"Coven", "Cabal", "Circle", "Ring", "Network", "Cooperative",
	"Conglomerate", "Consortium", "Cartel", "Trust", "Combine", "Bloc",
	"Faction", "Party", "Movement", "Cause", "Crusade", "Campaign",
	"Realm", "Domain", "Province", "Territory", "Principality", "Duchy",
	"Barony", "County", "Shire", "Hamlet", "Village", "Township",
	"Borough", "Municipality", "Commune", "Settlement", "Outpost", "Colony",
	"Dynasty", "Lineage", "House", "Estate", "Manor", "Holding",
	"Court", "Throne", "Crown", "Regime", "Administration", "Authority",
	"Bureau", "Agency", "Department", "Ministry", "Office", "Division",
	"Chapter", "Lodge", "Chamber", "Hall", "Assemblage", "Body",
	"Federacy", "Hegemony", "Sovereignty", "Protectorate", "Mandate", "Compact",
}

var whimsicalOfThings = []string{
	"Snacks", "Naps", "Mild Inconvenience", "Slightly Burnt Toast", "Spare Buttons",
	"Lost Socks", "Excellent Cheese", "Questionable Decisions", "Afternoon Tea",
	"Reasonable Doubt", "Comfortable Chairs", "Forgotten Errands",
	"Second Breakfast", "Unfinished Projects", "Tangled Headphones", "Warm Blankets",
	"Leftover Pizza", "Misplaced Keys", "Gentle Naps", "Cold Coffee",
	"Stale Crackers", "Crumpled Receipts", "Tepid Soup", "Lukewarm Tea",
	"Unread Emails", "Empty Promises", "Half Eaten Sandwiches", "Mismatched Socks",
	"Forgotten Passwords", "Dubious Leftovers", "Suspicious Casseroles", "Vague Intentions",
	"Postponed Chores", "Abandoned Hobbies", "Dusty Treadmills", "Untouched Salads",
	"Overdue Library Books", "Wilted Houseplants", "Tangled Cables", "Stray Crumbs",
	"Squeaky Floorboards", "Wobbly Tables", "Lopsided Cakes", "Melted Ice Cream",
	"Soggy Cereal", "Burnt Popcorn", "Flat Soda", "Stale Bread",
	"Crooked Pictures", "Loose Threads", "Errant Crumbs", "Lonely Cutlery",
	"Spare Change", "Pocket Lint", "Crumpled Napkins", "Forgotten Umbrellas",
	"Dubious Mushrooms", "Questionable Sushi", "Mystery Meat", "Suspicious Gravy",
	"Lukewarm Enthusiasm", "Mild Annoyance", "Faint Hope", "Modest Ambitions",
	"Reasonable Expectations", "Gentle Disappointment", "Quiet Desperation", "Polite Confusion",
	"Comfortable Silence", "Awkward Pauses", "Long Goodbyes", "Short Tempers",
	"Spilled Milk", "Burnt Toast", "Broken Pencils", "Dull Scissors",
	"Mismatched Gloves", "Odd Mittens", "Single Earrings", "Lost Buttons",
	"Cozy Sweaters", "Fuzzy Slippers", "Plump Cushions", "Soft Pillows",
	"Sleepy Cats", "Snoring Dogs", "Drowsy Hamsters", "Lazy Goldfish",
	"Idle Threats", "Empty Calories", "Hollow Victories", "Faint Praise",
	"Hasty Conclusions", "Wild Guesses", "Bold Assumptions", "Vague Plans",
	"Procrastinated Tasks", "Delayed Reactions", "Belated Apologies", "Premature Celebrations",
}

func generateWhimsicalName() string {
	switch rand.Intn(3) {
	case 0:
		return pick(whimsicalWholeNames)
	case 1:
		return "The " + pick(whimsicalAdjectives) + " " + pick(whimsicalNouns)
	default:
		return "Grand Duchy of " + pick(whimsicalOfThings)
	}
}

// pick returns a uniformly random element from a non-empty slice.
func pick(bank []string) string {
	return bank[rand.Intn(len(bank))]
}
