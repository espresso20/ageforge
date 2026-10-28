package mapstyle

import (
	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/vibe"
)

// agents.go says what the cursor says on the traffic: a cart, a car, a ship,
// each what it is, with a line that one of its kind keeps.

// movers are the traffic of the engine's ages by name: the title, and two
// lines to choose from.
var movers = map[string][3]vibe.Entry{
	"walker":          {vibe.E("map.mover.walker.title", "Worker"), vibe.E("map.mover.walker.line.01", "On the way to work."), vibe.E("map.mover.walker.line.02", "Heading home after a shift.")},
	"ox_cart":         {vibe.E("map.mover.ox_cart.title", "Ox cart"), vibe.E("map.mover.ox_cart.line.01", "An ox cart, creaking to market."), vibe.E("map.mover.ox_cart.line.02", "An ox cart, loaded with grain.")},
	"wagon":           {vibe.E("map.mover.wagon.title", "Wagon"), vibe.E("map.mover.wagon.line.01", "A wagon, hauling timber."), vibe.E("map.mover.wagon.line.02", "A covered wagon, bound for market.")},
	"early_car":       {vibe.E("map.mover.early_car.title", "Motor car"), vibe.E("map.mover.early_car.line.01", "An early motor car, sputtering along."), vibe.E("map.mover.early_car.line.02", "An early motor car, frightening the horses.")},
	"car":             {vibe.E("map.mover.car.title", "Car"), vibe.E("map.mover.car.line.01", "A car, stuck behind a truck."), vibe.E("map.mover.car.line.02", "A car, on the school run.")},
	"hovercar":        {vibe.E("map.mover.hovercar.title", "Hovercar"), vibe.E("map.mover.hovercar.line.01", "A hovercar, skimming the street."), vibe.E("map.mover.hovercar.line.02", "A hovercar, cutting corners.")},
	"shuttle":         {vibe.E("map.mover.shuttle.title", "Shuttle"), vibe.E("map.mover.shuttle.line.01", "A shuttle, climbing to orbit."), vibe.E("map.mover.shuttle.line.02", "A shuttle, outbound.")},
	"mining_drone":    {vibe.E("map.mover.mining_drone.title", "Mining drone"), vibe.E("map.mover.mining_drone.line.01", "A mining drone, hauling ore from the belt."), vibe.E("map.mover.mining_drone.line.02", "A mining drone, out for another rock.")},
	"generation_ship": {vibe.E("map.mover.generation_ship.title", "Generation ship"), vibe.E("map.mover.generation_ship.line.01", "A generation ship: its crew's grandchildren will land."), vibe.E("map.mover.generation_ship.line.02", "A generation ship, a town in a hull.")},
	"starship":        {vibe.E("map.mover.starship.title", "Starship"), vibe.E("map.mover.starship.line.01", "A starship, about to jump to warp."), vibe.E("map.mover.starship.line.02", "A starship, home from the far side of the arm.")},
	"phase_ship":      {vibe.E("map.mover.phase_ship.title", "Phase ship"), vibe.E("map.mover.phase_ship.line.01", "A phase ship, in several places at once."), vibe.E("map.mover.phase_ship.line.02", "A phase ship, arriving before it left.")},
	"mote":            {vibe.E("map.mover.mote.title", "Mote of light"), vibe.E("map.mover.mote.line.01", "A mote of light, drifting home."), vibe.E("map.mover.mote.line.02", "A mote of light. It was a city once.")},
}

// describeAgent is the traffic under the cursor.
func describeAgent(th mapgen.Thing) (Inspection, bool) {
	words, ok := movers[th.Key]
	if !ok {
		return Inspection{}, false
	}
	return Inspection{Title: words[0].String(), Lines: []string{words[1+th.Ref%2].String()}}, true
}
