package world

import "sync"

// FrameKind says what a frame holds.
type FrameKind uint8

const (
	FrameLocalGroup    FrameKind = iota // the void between galaxies (kly)
	FrameGalaxy                         // the home galaxy, seeded, with the people's territory (ly)
	FrameRegion                         // the Orion arm sector round Sol (ly)
	FrameNeighbourhood                  // Sol's nearest stars (ly)
	FrameSystem                         // a solar system (AU)
	FramePlanet                         // the home world's neighbourhood (planet radii)
)

// Unit is what a frame's numbers are measured in.
type Unit uint8

const (
	UnitKly         Unit = iota // thousands of light-years
	UnitLy                      // light-years
	UnitAU                      // astronomical units
	UnitPlanetRadii             // radii of the frame's planet
)

// Vec is a position or size in a frame's own units.
type Vec struct{ X, Y float64 }

// Band is the zoom band a frame is meant to be seen in: the width of the
// view, in the frame's units, from where the frame takes over from its
// parent (Max, zoomed out) to where its children take over (Min, zoomed in).
type Band struct{ Min, Max float64 }

// Frame is one level of the nested sky. Its Pos is where it hangs in the
// parent frame, in the parent's units; its Radius and everything in Objects
// are in its own.
type Frame struct {
	Kind     FrameKind
	Key      string // stable id, also the label key of the frame's centre
	Unit     Unit
	Anchor   string  // key of the parent's object this frame hangs on ("" its centre)
	Pos      Vec     // anchor in the parent's units, at phase zero for a moving anchor
	Radius   float64 // the extent of the frame, in its own units
	Band     Band
	Schem    bool    // positions are schematic, not to scale (the Local Group)
	Inflate  float64 // a frame drawn this many times its unit's true size (0 or 1: true size): the home close-up, so its planet has a face at the handoff
	Children []*Frame
	Objects  []Object
}

// ObjKind is what an object is.
type ObjKind uint8

const (
	KindGalaxy  ObjKind = iota // a galaxy, with its look
	KindFeature                // an arm feature: a nebula, a cluster, a knot, the arm's band
	KindStar                   // a star
	KindSector                 // a lit system of one building line's sector
	KindRival                  // a system of a civilization met
	KindRift                   // a cell of the crack a harbinger opens
	KindLane                   // a warp lane between two colony stars
	KindColony                 // a near star the people have lit
	KindPlanet                 // a planet
	KindMoon                   // a moon
	KindBelt                   // an asteroid or Kuiper belt
	KindComet                  // a comet
	KindStation                // a station, module or mine
	KindGate                   // the warp gate
	KindShip                   // a ship
)

// Form is what a feature looks like: the arm's band, a cluster of stars, a
// glowing or dark cloud, a remnant, a knot of the arm or the people's reach.
type Form uint8

const (
	FormNone      Form = iota
	FormBand           // the arm's band of faint stars and dust
	FormCluster        // an open cluster: bright points
	FormEmission       // an emission nebula: a pink cloud with dark filaments
	FormDark           // a dark nebula: a silhouette against the stars
	FormRemnant        // a supernova remnant: a shell of filaments
	FormKnot           // a small star-forming knot of the arm
	FormTerritory      // the people's reach, a ring round Sol
	FormZone           // a sector of the galaxy: its pick zone, drawn by the disc's tint
)

// Object is anything the map shows. Not every field means something for every
// kind; the comments on the fields say which kinds read them.
type Object struct {
	Kind ObjKind
	Form Form   // KindFeature: how it looks
	Key  string // unique within the frame
	// Label and Inspect are vibe keys: the name the map writes by the object
	// and the text the inspect panel shows. Empty means unlabelled/unpickable.
	Label, Inspect string
	Pos            Vec     // in the frame's units
	Size           float64 // radius, in the frame's units
	Tone           Tone
	Weight         uint8
	Gated          string // what game state lit it: "" for seeded, else a short reason

	Line  string // sector, rival, colony, lane, station: the building line
	Count int    // colony, sector: buildings; station: copies
	Tier  int    // colony: 0..3
	To    Vec    // lane: the far end (Pos is the near one)
	Flow  int    // lane: traffic 1..3

	Galaxy *GalaxyLook  // KindGalaxy
	Star   *Look        // KindStar, KindColony
	Body   *Body        // KindPlanet, KindMoon, KindBelt, KindComet
	Shape  *GalaxyShape // the seeded galaxy's arms and tilt (the home galaxy in the Local Group and Galaxy frames)
	Seed   int64        // with Shape: the save's seed, which the arms' winding reads
	Path   []Vec        // a polyline: the home arm, core to rim (KindFeature)
	Pick   bool         // the player can select it
	Held   bool         // a sector or generated star: the people hold it, so a star may be entered
	Terr   *Zones       // KindGalaxy of the home galaxy: the territory the disc is tinted with
}

// Body is the orbit and make-up of a planet, moon, belt or comet.
type Body struct {
	Spec     PlanetSpec // the face (planets)
	Parent   string     // key of what it goes round ("" the frame's centre)
	Orbit    float64    // radius of the orbit, in the frame's units
	Period   int        // ticks of the animation clock for one turn
	Phase    int        // where it starts, in 4096ths of a turn
	Rings    bool
	Moons    int
	Kind     string  // "rocky", "gas", "ice", "dwarf", "moon", "belt", "comet"
	Earths   float64 // physical radius, in Earth radii (planets); Object.Size is the art size
	StationN int     // stations in its close-up
	Settled  int     // colonies of the people on it (generated systems)
}

// World is the whole sky for one save.
type World struct {
	Seed int64
	Age  string
	Root *Frame // the outermost frame of the age's gate: the locked one if there is one (see GateOf)
	Key  uint64 // ChangeKey of what it was built from

	mu  sync.Mutex        // guards sys
	sys map[string]*Frame // generated near-star systems, built on first ask
}

// Find returns the frame of kind k at or below f, or nil.
func (f *Frame) Find(k FrameKind) *Frame {
	if f == nil {
		return nil
	}
	if f.Kind == k {
		return f
	}
	for _, c := range f.Children {
		if r := c.Find(k); r != nil {
			return r
		}
	}
	return nil
}

// Count is how many objects of kind k the frame and its children hold.
func (f *Frame) Count(k ObjKind) int {
	n := 0
	for _, o := range f.Objects {
		if o.Kind == k {
			n++
		}
	}
	for _, c := range f.Children {
		n += c.Count(k)
	}
	return n
}
