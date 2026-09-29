package skyline

import "github.com/espresso20/ageforge/mapmodel"

// catalog.go maps a building (lineage, period) onto a form and a nominal
// height in rows at full scale. This table is the only place that knows
// about lineages; everything else is geometry.

// form is an archetype with its nominal height.
type form struct {
	fn     archFn
	height float64 // rows at scale 1 (a 45-row terminal)
	smoke  bool    // a producer: its chimneys smoke only while staffed
}

func f(fn archFn, h float64) form  { return form{fn: fn, height: h} }
func fs(fn archFn, h float64) form { return form{fn: fn, height: h, smoke: true} }

// pick returns forms[i], clamped.
func pick(i int, forms ...form) form {
	if i < 0 {
		i = 0
	}
	if i >= len(forms) {
		i = len(forms) - 1
	}
	return forms[i]
}

// familyOf is the material family (and period band) of an age index:
// 0 primitive/stone, 1 bronze/iron, 2 classical, 3 medieval,
// 4 renaissance/colonial, 5 industrial/victorian, 6 electric/atomic,
// 7 modern..digital, 8 cyberpunk, 9 fusion/space, 10 cosmic.
func familyOf(age int) int {
	switch {
	case age <= 1:
		return 0
	case age <= 3:
		return 1
	case age == 4:
		return 2
	case age == 5:
		return 3
	case age <= 7:
		return 4
	case age <= 9:
		return 5
	case age <= 11:
		return 6
	case age <= 14:
		return 7
	case age == 15:
		return 8
	case age <= 17:
		return 9
	}
	return 10
}

// late indexes the post-cyberpunk forms by age (16 fusion … 21 transcendent).
func late(age int, forms ...form) form { return pick(age-16, forms...) }

func formFor(d *mapmodel.Def) form {
	a := d.Age
	b := familyOf(a)
	switch d.Lineage {
	case mapmodel.LinHousing:
		if b >= 9 {
			return late(a, f(aHabRing, 12), f(aSaucer, 14), f(aGenShip, 7), f(aArcology, 20), f(aTesseract, 10), f(aPillar, 22))
		}
		return pick(b, f(aHut, 4), f(aMudHouse, 5), f(aVilla, 7), f(aTimber, 7), f(aVilla, 9),
			f(aRowHouse, 8), f(aDeco, 14), f(aGlass, 18), f(aCyber, 24))
	case mapmodel.LinFood:
		switch {
		case b == 0:
			return f(aTent, 3)
		case b <= 6:
			if d.Key == "demesne" || d.Key == "estate_farm" {
				return f(aWindmill, 7)
			}
			return f(aFields, 4)
		case b <= 8:
			return f(aVertFarm, 14)
		}
		return late(a, f(aHabDome, 5), f(aVertFarm, 14), f(aHabDome, 6), f(aCrystal, 10), f(aArcology, 14))
	case mapmodel.LinWood:
		switch {
		case b <= 1:
			return f(aWoodCamp, 5)
		case b <= 4:
			return fs(aSmithy, 6)
		case d.Key == "oil_derrick" || d.Key == "oil_field" || d.Key == "oil_platform":
			return f(aDerrick, 9)
		case b <= 8:
			return fs(aFactory, 10)
		}
		return late(a, f(aVertFarm, 12), f(aHabDome, 6), f(aCrystal, 10), f(aArcology, 14), f(aTesseract, 9))
	case mapmodel.LinMines:
		switch {
		case b <= 2:
			return f(aStonesPit, 3)
		case b <= 5:
			return fs(aMineHead, 7)
		case b <= 8:
			return fs(aFactory, 10)
		}
		return late(a, f(aDerrick, 9), f(aDerrick, 10), f(aCrystal, 11), f(aCrystal, 13), f(aTesseract, 9))
	case mapmodel.LinKnowledge:
		switch {
		case b == 0:
			return f(aFirepit, 2)
		case a == 1:
			return f(aLonghouse, 4)
		case b == 1:
			return f(aMudHouse, 5)
		case b == 2:
			return f(aClassical, 7)
		case b == 3:
			return f(aCathedral, 11)
		case b <= 5:
			return f(aDomed, 11)
		case b == 6:
			return f(aDeco, 13)
		case b <= 8:
			return pick(b-7, f(aGlass, 16), f(aCyber, 20))
		}
		return late(a, f(aRingSpire, 18), f(aObservatory, 8), f(aRingSpire, 22), f(aCrystal, 16), f(aTesseract, 11))
	case mapmodel.LinFaith:
		switch {
		case a <= 1:
			return f(aStones, 3)
		case b == 1:
			return f(aZiggurat, 8)
		case b == 2:
			return f(aClassical, 8)
		case b <= 5:
			return f(aCathedral, 14)
		case b <= 7:
			return f(aDomed, 12)
		case b == 8:
			return f(aCyber, 18)
		}
		return late(a, f(aPrism, 12), f(aSaucer, 12), f(aPrism, 16), f(aCrystal, 16), f(aPillar, 20))
	case mapmodel.LinTrade:
		if b >= 9 {
			return late(a, f(aGlass, 24), f(aRingSpire, 22), f(aArcology, 20), f(aCrystal, 18), f(aTesseract, 11), f(aPillar, 24))
		}
		return pick(b, f(aMarket, 3), f(aMarket, 3), f(aClassical, 7), f(aTimber, 8), f(aDomed, 10),
			f(aDomed, 12), f(aDeco, 18), f(aGlass, 24), f(aCyber, 28))
	case mapmodel.LinMetal:
		switch {
		case b <= 4:
			return fs(aSmithy, 6)
		case b <= 7:
			return fs(aFactory, 11)
		case b <= 8:
			return fs(aFusion, 7)
		}
		return late(a, fs(aFusion, 7), fs(aFusion, 8), f(aRingSpire, 16), f(aCrystal, 14), f(aTesseract, 10))
	case mapmodel.LinEngineer:
		switch {
		case d.Key == "aqueduct":
			return f(aAqueduct, 5)
		case d.Key == "mill":
			return f(aWindmill, 7)
		case b <= 4:
			return fs(aSmithy, 6)
		case d.Key == "nuclear_plant":
			return fs(aCooling, 9)
		case b <= 6:
			return fs(aFactory, 11)
		case d.Key == "launch_complex":
			return fs(aLaunch, 15)
		case d.Key == "fusion_reactor":
			return f(aFusion, 7)
		case b <= 8:
			return f(aDataCenter, 5)
		}
		return late(a, f(aFusion, 7), fs(aLaunch, 15), f(aRingSpire, 20), f(aCrystal, 16), f(aTesseract, 11), f(aPillar, 24))
	case mapmodel.LinEnergy:
		switch d.Key {
		case "nuclear_reactor":
			return fs(aCooling, 9)
		case "solar_collector_array", "smart_energy_grid":
			return f(aSolar, 2)
		case "fusion_reactor_array":
			return f(aFusion, 7)
		case "power_generator", "quantum_battery_array":
			return f(aPylon, 8)
		}
		if b <= 6 {
			return fs(aFactory, 11)
		}
		if b <= 8 {
			return f(aDataCenter, 5)
		}
		return late(a, f(aFusion, 7), f(aSolar, 2), f(aDish, 6), f(aRingSpire, 20), f(aTesseract, 10))
	case mapmodel.LinCulture:
		if b >= 9 {
			return late(a, f(aHabDome, 5), f(aSaucer, 12), f(aSaucer, 16), f(aCrystal, 14), f(aPrism, 14))
		}
		return pick(b, f(aFirepit, 2), f(aFirepit, 2), f(aClassical, 7), f(aTimber, 8), f(aDomed, 11),
			f(aDomed, 11), f(aDish, 6), f(aGlass, 16), f(aCyber, 20))
	case mapmodel.LinHacker:
		if b <= 8 {
			return f(aDataCenter, 5)
		}
		return late(a, f(aDataCenter, 5), f(aDish, 6), f(aRingSpire, 18), f(aCrystal, 14), f(aTesseract, 10))
	case mapmodel.LinHarbor:
		if b <= 5 {
			return f(aWarehouse, 4)
		}
		return f(aCrane, 8)
	case mapmodel.LinMilitary:
		switch {
		case b == 0:
			return f(aTent, 3)
		case b == 1:
			return f(aTower, 6)
		case b <= 3:
			if d.Key == "castle_keep" {
				return f(aCastle, 9)
			}
			return f(aTower, 8)
		case b <= 5:
			return f(aCastle, 9)
		case b <= 8:
			return f(aBunker, 5)
		}
		return late(a, f(aBunker, 5), f(aHabDome, 6), f(aRingSpire, 16), f(aCrystal, 14), f(aTesseract, 10), f(aPillar, 20))
	case mapmodel.LinStorage:
		switch {
		case b == 0:
			return f(aPile, 2)
		case b <= 2:
			return f(aSilo, 4)
		case b <= 3:
			return f(aTower, 8)
		case b <= 6:
			return f(aWarehouse, 4)
		case b <= 8:
			return f(aVault, 4)
		}
		return late(a, f(aVault, 4), f(aHabDome, 5), f(aHabDome, 5), f(aCrystal, 7), f(aTesseract, 7))
	case mapmodel.LinMonument:
		if b >= 9 {
			return f(aPrism, 10)
		}
		return f(aObelisk, 8)
	case mapmodel.LinDiplomacy:
		if b >= 9 {
			return f(aSaucer, 12)
		}
		return f(aDomed, 10)
	}
	switch {
	case b <= 3:
		return f(aVilla, 6)
	case b <= 7:
		return f(aGlass, 14)
	}
	return f(aRingSpire, 18)
}
