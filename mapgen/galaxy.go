package mapgen

import "github.com/espresso20/ageforge/world"

// The galaxy of the three space ages now lives in world/ (shape, arms, the
// home arm). These are the names the painters and tests already use.

const (
	turn      = world.Turn
	galRX     = world.GalRX
	galRYBase = world.GalRYBase
	galArmN   = world.GalArmN
)

type (
	GalaxyShape = world.GalaxyShape
	Galaxy      = world.Galaxy
)

var (
	sin4096   = world.Sin4096
	cos4096   = world.Cos4096
	bump      = world.Bump
	sq        = world.Sq
	NewGalaxy = world.NewGalaxy
)
