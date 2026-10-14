package config

import "math"

// LatePriceStep is the per-age price step from the Modern Age on. Prices in
// the Modern Age stay at their authored value; each following age costs this
// factor more. Outputs do not move with this curve.
const LatePriceStep = 1.2

// LatePriceStartAge is the first age on the late-price curve.
const LatePriceStartAge = "modern_age"

// LatePriceMultiplier is the price multiplier for age. It is one before and
// in the Modern Age, and LatePriceStep for every age after it.
func LatePriceMultiplier(age string) float64 {
	positions := AgePositions(AgeOrder())
	start, ok := positions[LatePriceStartAge]
	if !ok {
		return 1
	}
	pos, ok := positions[age]
	if !ok || pos <= start {
		return 1
	}
	return math.Pow(LatePriceStep, float64(pos-start))
}

// LatePriceTargetTicks is an analytical pacing target for a price-derived
// rule. It is not a game clock: it describes the longer stay created by the
// late-price curve while production remains unchanged.
func LatePriceTargetTicks(age string) float64 {
	return AgeTargetTicks(age) * LatePriceMultiplier(age)
}

func applyLatePriceClimb(defs []BuildingDef) {
	for i := range defs {
		c := LatePriceMultiplier(defs[i].RequiredAge)
		if c == 1 {
			continue
		}
		for resource, price := range defs[i].BaseCost {
			defs[i].BaseCost[resource] = price * c
		}
	}
}

// rederiveLatePriceStorage keeps storage on its existing price rule after the
// late-age curve changes its inputs. Storage is expressed in resources, so it
// has to be written back into the runtime table after the prices are scaled.
func rederiveLatePriceStorage(defs []BuildingDef) {
	base := 0.0
	for _, resource := range BaseResources() {
		if resource.Key == "wood" {
			base = resource.BaseStorage
			break
		}
	}
	sizes, _ := StorageRuleSizes(defs, Ages(), base)
	for i := range defs {
		if size, ok := sizes[defs[i].Key]; ok {
			for j := range defs[i].Effects {
				if defs[i].Effects[j].Type == "storage" && defs[i].Effects[j].Target == "all" {
					defs[i].Effects[j].Value = size
				}
			}
		}
	}
}
