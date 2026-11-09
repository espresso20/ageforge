// Package world is the sky: everything in space the map shows, from the Local
// Group down to the home planet's moons and stations, as nested frames with
// their own units. It decides what is there; it draws nothing. It is pure: no
// ebiten, tview, tcell, theme or ui. Colours are Tone roles, never RGB.
//
// Frames (unit; zoom band as view width in the unit):
//
//	LocalGroup   kly   400..6000     schematic: galaxies not to scale
//	  Galaxy     ly    15000..140000   the seeded home galaxy, 50000 ly radius
//	    Region   ly    400..18000      the Orion arm round Sol, 8500 ly reach
//	      Neighbourhood ly 3..60       stars within 13 ly of Sol
//	        System AU  0.5..140        Sol's system; also the root in space/interstellar ages
//	          Planet radii 1.5..800    the home world, its moon, stations
//
// Star ages (galactic, quantum, transcendent) root at the Local Group; space
// and interstellar ages root at the System; ground ages have no root.
//
// Objects (kind: source, gate):
//
//	galaxy     Milky Way, Andromeda, Triangulum, the Clouds: fixed look; always (Local Group)
//	galaxy     far galaxies: seeded, 24 candidates minus overlaps; always, faint
//	galaxy     the home galaxy's disc and arms: seeded shape (PickShape); always (Galaxy)
//	feature    home arm path, arm band, pink knots (11), named nebulae (8): fixed table + seed; always
//	feature    territory tint: grows with lit colony count; needs a lineage standing
//	star       Sol, the 19 near stars: fixed catalogue; always (Region, Neighbourhood, System)
//	sector     one per lit system of a building line's sector: seeded cells, count from the line's buildings; lineage Count>0
//	rival      systems of a civilization: slot picked by key, count from strength; civ Discovered
//	rift       cells of the crack across the galaxy: seeded path; a harbinger is present
//	colony     a near star lit by a line: ColonyStars picks it; lineage Count>0
//	lane       warp lane between colonies and Sol, traffic 1..3 by the smaller count; two or more lit
//	planet     eight planets (real AU, art size, seeded phase), Pluto: fixed table; always (System)
//	belt       asteroid belt, Kuiper belt: fixed; always
//	comet      one comet on a 2800-tick cycle: seeded phase; always
//	gate       warp gate at 35 AU: wonder warp_nexus built, or any star-age system view
//	moon       the Moon: seeded phase; always (Planet)
//	station    2..3 stations, the 10 satellite swarm, 2..3 mega stations: seeded; always
//	station    one module per building line standing: lineage Count>0, space_age only
//	ship       the starship: seeded; always
//
// Not modelled (painter-only today): the space-age ring city's parcel layout,
// the module hull art, nebula recipes, the quantum age's blinking ghost
// systems and the transcendent age's halo ring.
package world
