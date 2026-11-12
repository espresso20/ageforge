package world

// GalaxyLook is how a galaxy is drawn: the parameters of its arms. Moved
// from mapgen's wide views; the Local Group lays the Milky Way, Andromeda and
// Triangulum with these.
type GalaxyLook struct {
	R, Q, Phi int // radius, the squash of its tilt out of 256, its angle (4096ths of a turn)
	Arms      int
	Spin      int // turns of the arms, in 256ths
	Core, Bar int
	Dust      int
	Stars     int
	Spurs     bool // two fainter, shorter arms between the main ones
	Irr       bool // an irregular dwarf: clumps, no arms (a bar if Bar is set)
}

// The three galaxies the Local Group draws with arms (radius is filled by the
// layout).
var (
	LookMilkyWay   = GalaxyLook{Q: 256, Phi: 0, Arms: 2, Spin: 200, Core: 150, Bar: 130, Dust: 200, Stars: 1, Spurs: true}
	LookAndromeda  = GalaxyLook{Q: 85, Phi: -Turn / 12, Arms: 2, Spin: 256, Core: 170, Dust: 230, Stars: 1}
	LookTriangulum = GalaxyLook{Q: 256, Phi: 0, Arms: 2, Spin: 200, Core: 90, Dust: 0, Stars: 1}
)

// ArmXY is the point at t (0..255 along it) of the arm that leaves at angle
// off, round (cx, cy); with a bar the arms leave from its ends. It also
// returns the radius.
func ArmXY(g GalaxyLook, cx, cy, off, t int) (int, int, int) {
	lo := 30
	if g.Bar > 0 {
		off += Turn / 7
		lo = 97
	}
	th := off + t*g.Spin*Turn/65536
	r := g.R * (lo + (256-lo)*t*t/65536) / 256 // wide pitch outside, like a log spiral
	u, v := r*Cos4096(th)/1024, r*Sin4096(th)/1024*g.Q/256
	cs, sn := Cos4096(g.Phi), Sin4096(g.Phi)
	return cx + (u*cs-v*sn)/1024, cy + (u*sn+v*cs)/1024, r
}

// SkyObject is a named thing of the wide views.
type SkyObject struct {
	Key string
	Ly  int // light-years from Sol; 0 for the Milky Way, which has none
	Lon int // galactic longitude, degrees (arm objects)
}

// SkyObjects are the selectable objects: the arm sector's first, then the
// void's. A map thing's Ref is the index plus one.
var SkyObjects = [...]SkyObject{
	{"orion_nebula", 1344, 209}, {"horsehead_nebula", 1375, 206}, {"pleiades", 444, 167}, {"hyades", 153, 180},
	{"crab_nebula", 6500, 185}, {"eagle_nebula", 7000, 17}, {"lagoon_nebula", 4100, 6}, {"carina_nebula", 8500, 287},
	{"milky_way", 0, 0}, {"large_magellanic_cloud", 160000, 0}, {"small_magellanic_cloud", 200000, 0},
	{"andromeda", 2500000, 0}, {"triangulum", 2700000, 0},
}

const (
	ArmObjects = 8    // the first SkyObjects that stand in the Orion arm
	ArmReachLy = 8500 // the arm sector's reach, in light-years
	ArmBandDeg = 20   // the arm's direction on the screen
	ArmSkew    = 6    // degrees the Horsehead is drawn aside so it does not sit on the Orion Nebula
)

// ArmRadii is the drawn size of each arm object, in the layout's units.
var ArmRadii = [ArmObjects]int{7, 4, 4, 4, 5, 7, 7, 10}
