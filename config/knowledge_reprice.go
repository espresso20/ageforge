package config

// The prices of the buildings that were priced in knowledge, as they stood
// before knowledge stopped being a price: the other parts of each (the
// knowledge part is gone). The age's price levels are the medians of the
// first-copy prices; the rule that took knowledge out of a price multiplied a
// building's other parts so that its price in those units stayed what it was
// (TestRepricedBuildingsKeepTheirUnits), and the levels themselves must not
// move because six of an age's buildings were re-priced. So the levels read
// these amounts for these buildings (priceLevels) and the amounts in the
// tables for every other.
var repricedFrom = map[string]map[string]float64{
	"demesne":                 {"stone": 540000, "gold": 180000},
	"market_garden":           {"gold": 1800000, "steel": 600000},
	"stonemasons_guild":       {"stone": 480000, "gold": 150000},
	"iron_mine":               {"gold": 1500000, "steel": 540000},
	"monastery_library":       {"stone": 540000, "gold": 180000},
	"university":              {"gold": 1800000, "steel": 600000},
	"natural_philosophy_hall": {"gold": 1.1e07, "steel": 4500000},
	"sawmill":                 {"stone": 480000, "gold": 170000},
	"coal_mine":               {"gold": 1700000, "steel": 600000},
	"cathedral":               {"stone": 600000, "gold": 200000},
	"basilica":                {"gold": 2000000, "stone": 1200000},
	"mission":                 {"gold": 9000000, "stone": 6000000},
	"guildhall":               {"stone": 1200000, "gold": 410000},
	"exchange":                {"gold": 4100000, "steel": 1500000},
	"art_studio":              {"gold": 2000000, "stone": 900000},
	"great_library":           {"stone": 5600000, "gold": 4320000},
}

// repricedUnits is the price of each of those buildings, in the age's price
// units, with its knowledge part: the price the other parts were multiplied to
// keep (TestRepricedBuildingsKeepTheirUnits holds the tables to it, to the
// three figures a price is written to).
var repricedUnits = map[string]float64{
	"demesne": 3.3333, "market_garden": 3.25, "stonemasons_guild": 2.7222, "iron_mine": 2.6928,
	"monastery_library": 3.0, "university": 3.0, "natural_philosophy_hall": 2.6071, "sawmill": 2.8333,
	"coal_mine": 2.9444, "cathedral": 3.2222, "basilica": 3.6944, "mission": 2.8929,
	"guildhall": 7.1667, "exchange": 7.2361, "art_studio": 3.1111,
}
