package main

import "github.com/espresso20/ageforge/config"

// catalog.go maps a building (lineage, period) onto an archetype and a
// nominal height in rows at full scale. This table is the only place that
// knows about lineages; everything else is geometry.

type form struct {
	fn     archFn
	height float64 // nominal rows at scale 1 (a 45-row terminal)
	smoke  bool    // a producer: its chimneys smoke only while staffed
}

func f(fn archFn, h float64) form     { return form{fn: fn, height: h} }
func fs(fn archFn, h float64) form    { return form{fn: fn, height: h, smoke: true} }
func byAge(a int, forms ...form) form { return forms[min(a, len(forms)-1)] }
func ageBand(a int, cut []int) int {
	for i, c := range cut {
		if a < c {
			return i
		}
	}
	return len(cut)
}

func formFor(def config.BuildingDef, age int) form {
	lin := def.LineageKey
	if def.Category == "storage" {
		lin = "storage"
	}
	if def.Category == "monument" {
		lin = "monument"
	}
	// period bands: 0 prim/stone, 1 bronze/iron, 2 classical, 3 medieval,
	// 4 ren/colonial, 5 industrial/victorian, 6 electric/atomic,
	// 7 modern..digital, 8 cyberpunk, 9 fusion/space, 10 cosmic
	b := familyOf(age)
	switch lin {
	case "housing":
		return byAge(b, f(aHut, 4), f(aMudHouse, 5), f(aVilla, 7), f(aTimber, 7), f(aVilla, 9),
			f(aRowHouse, 8), f(aDeco, 14), f(aGlass, 18), f(aCyber, 24), f(aArcology, 20), f(aSaucer, 12))
	case "food":
		switch {
		case b == 0:
			return f(aTent, 3)
		case b <= 6:
			if def.Key == "demesne" || def.Key == "estate_farm" {
				return f(aWindmill, 7)
			}
			return f(aFields, 4)
		case b <= 8:
			return f(aVertFarm, 14)
		default:
			return f(aHabDome, 5)
		}
	case "organic_extraction":
		switch {
		case b <= 1:
			return f(aWoodCamp, 5)
		case b <= 4:
			return fs(aSmithy, 6)
		case def.Key == "oil_derrick" || def.Key == "oil_field" || def.Key == "oil_platform":
			return f(aDerrick, 9)
		case b <= 6:
			return fs(aFactory, 10)
		default:
			return fs(aFactory, 10)
		}
	case "geological_extraction":
		switch {
		case b <= 2:
			return f(aStonesPit, 3)
		case b <= 5:
			return fs(aMineHead, 7)
		default:
			return fs(aFactory, 10)
		}
	case "knowledge":
		switch {
		case b == 0:
			return f(aFirepit, 2)
		case age == 1:
			return f(aLonghouse, 4)
		case b == 1:
			return f(aMudHouse, 5)
		case b == 2:
			return f(aClassical, 7)
		case b == 3:
			return f(aCathedral, 11)
		case b == 4:
			return f(aDomed, 11)
		case b == 5:
			return f(aDomed, 12)
		case b == 6:
			return f(aDeco, 13)
		case age == 17:
			return f(aDish, 6)
		default:
			return byAge(b-7, f(aGlass, 16), f(aCyber, 20), f(aHabDome, 5), f(aSpire, 22))
		}
	case "faith":
		switch {
		case age == 0:
			return f(aStones, 3)
		case age == 1:
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
		default:
			return f(aSpire, 20)
		}
	case "trade":
		return byAge(b, f(aMarket, 3), f(aMarket, 3), f(aClassical, 7), f(aTimber, 8), f(aDomed, 10),
			f(aDomed, 12), f(aDeco, 18), f(aGlass, 24), f(aCyber, 28), f(aGlass, 26), f(aArcology, 22))
	case "metallurgy":
		switch {
		case b <= 4:
			return fs(aSmithy, 6)
		case b <= 7:
			return fs(aFactory, 11)
		case b <= 9:
			return fs(aFusion, 7)
		default:
			return f(aSpire, 18)
		}
	case "engineering":
		switch {
		case b <= 2:
			if def.Key == "aqueduct" {
				return f(aAqueduct, 5)
			}
			return fs(aSmithy, 6)
		case def.Key == "mill":
			return f(aWindmill, 7)
		case b <= 4:
			return fs(aSmithy, 6)
		case def.Key == "nuclear_plant":
			return fs(aCooling, 9)
		case b <= 6:
			return fs(aFactory, 11)
		case def.Key == "launch_complex":
			return fs(aLaunch, 15)
		case def.Key == "fusion_reactor":
			return f(aFusion, 7)
		case b <= 8:
			return f(aDataCenter, 5)
		default:
			return f(aSpire, 20)
		}
	case "energy":
		switch def.Key {
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
		return f(aSpire, 22)
	case "culture_arts":
		return byAge(b, f(aFirepit, 2), f(aFirepit, 2), f(aClassical, 7), f(aTimber, 8), f(aDomed, 11),
			f(aDomed, 11), f(aDish, 6), f(aGlass, 16), f(aCyber, 20), f(aHabDome, 5), f(aSaucer, 16))
	case "hacker":
		if b <= 8 {
			return f(aDataCenter, 5)
		}
		return f(aSpire, 20)
	case "harbor":
		if b <= 5 {
			return f(aWarehouse, 4)
		}
		return f(aCrane, 8)
	case "military":
		switch {
		case b == 0:
			return f(aTent, 3)
		case b == 1:
			return f(aTower, 6)
		case b <= 3:
			if def.Key == "castle_keep" {
				return f(aCastle, 9)
			}
			return f(aTower, 8)
		case b <= 5:
			return f(aCastle, 9)
		case b <= 8:
			return f(aBunker, 5)
		default:
			return f(aHabDome, 5)
		}
	case "storage":
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
		default:
			return f(aHabDome, 5)
		}
	case "monument":
		return f(aObelisk, 8)
	case "diplomacy":
		return f(aDomed, 10)
	}
	switch {
	case b <= 3:
		return f(aVilla, 6)
	case b <= 7:
		return f(aGlass, 14)
	default:
		return f(aSpire, 18)
	}
}

func aStonesPit(r *rnd, h int) *Sprite {
	w := r.rng(6, 9)
	s := newSprite(w, 3)
	for x := 0; x < w; x++ {
		hh := 1 + int(hash(x, int(r.s))%2)
		if x == 0 || x == w-1 {
			hh = 1
		}
		for y := 3 - hh; y < 3; y++ {
			slot := SRock
			if x > w/2 {
				slot = SRockDark
			}
			s.put(x, y, '█', slot, SNone)
		}
		if hh == 1 {
			s.put(x, 1, '▄', SRock, SNone)
		}
	}
	s.put(w/2, 1, '╥', STrunk, SNone) // a hoist
	s.put(w/2, 0, '╓', STrunk, SNone)
	return s
}

func aMineHead(r *rnd, h int) *Sprite {
	s := newSprite(9, 7)
	// headframe
	for y := 1; y < 7; y++ {
		s.put(1, y, '╱', SMetal, SNone)
		s.put(3, y, '║', SMetal, SNone)
	}
	s.put(3, 0, 'O', SMetal, SNone)
	s.hline(1, 3, 1, '═', SMetal, SNone)
	s.body(4, 4, 5, 3)
	s.hline(4, 8, 3, '▄', SRoof, SNone)
	s.put(7, 2, '█', SWallDark, SNone)
	s.put(7, 3, '█', SWallDark, SNone)
	s.Smoke = append(s.Smoke, pt{7, 1})
	s.put(5, 5, '■', SWin, SWall)
	s.put(1, 6, '▄', SRock, SNone)
	s.put(0, 6, '▄', SRockDark, SNone)
	return s
}

func aAqueduct(r *rnd, h int) *Sprite {
	w := r.rng(11, 15)
	s := newSprite(w, 5)
	s.hline(0, w-1, 1, '█', STrim, SNone)
	s.hline(0, w-1, 0, '▄', STrim, SNone)
	for x := 0; x < w; x++ {
		if x%3 == 0 {
			s.vline(x, 2, 4, '█', STrim, SNone)
		} else {
			s.put(x, 2, '▀', STrim, SNone)
		}
	}
	s.hline(0, w-1, 1, '▄', SWater, STrim)
	return s
}
