# Count-Driven City Synthesis (Citymap v2)

Supersedes the "one named marker per built type + MST road" citymap with a
procedurally-synthesized top-down city whose **form and density derive from the
player's actual building COUNTS and current AGE**.

Primitive/Stone = a small organic village of huts + dirt paths. Each later era
band grows denser and more structured — wider streets, gridded blocks, alleys,
gardens/parks, plazas, walls: ancient Mesopotamia → medieval → colonial/Victorian
→ modern → cyber → orbital. The **named landmark buildings** (Shrine, Gathering
Camp, …) remain as labeled hero structures embedded in the fabric — the city
grows around them.

## Model (new: citygen.go)

`cityPlan` — the render-ready output of the generator:
- `streets []street` — polylines with a width class (avenue/street/alley) + a
  paved color role. Terrain-routed: never cross water/mountain (reuse pathfind).
- `blocks  []block`  — parcels bounded by streets (rect for grid eras, loose
  hulls for organic).
- `lots    []lot`    — placed structures filling the blocks.

`lot`:
- rect `(x,y,w,h)`, `kind` (house | workshop | garden | plaza | wall | tower |
  landmark), `domain` (drives lineage color), `tier` (age tier → style),
  `label` ("" except for landmark lots, which carry the building name).

## Pipeline (per render — pure, deterministic from seed)

1. **gather**: built buildings → per type `{key, domain, category, tier, count}`;
   classify residential / production / civic-landmark / storage / wonder.
2. **size**: `cityScale = f(totalBuildings, era)` with sqrt/log scaling so large
   late-game counts still fit the canvas; sets city radius / block count /
   street density. Counts are REPRESENTATIVE, not 1:1 (12 huts → a house
   cluster, 400 buildings → a metropolis, both legible).
3. **streets**: the era street-pattern generator emits the network — organic
   walk | radial spokes | orthogonal grid | superblock grid | campus links |
   orbital rings — at `eraStyle` widths, terrain-routed around water.
4. **blocks**: derive parcels between streets.
5. **populate (COUNT-DRIVEN)**: distribute lots into blocks —
   - residential/housing counts → house lots (N scaled by count; style per era:
     hut → mudbrick → timber → rowhouse → tower → arcology),
   - production counts → workshop/factory lots in the production zone (era-zoned),
   - civic/faith/knowledge/wonder → landmark lots at prominent anchors
     (plaza / center / keep); these carry the label,
   - leftover block area → garden/park lots (`gardenRatio` grows with era),
   - era extras: walls + keep (medieval), formal parks (Victorian), plazas (civic).
6. **terrain-gate**: every lot on passable land; streets route around water
   (reuse terrainField + pathfind).

## eraStyle presets (per era band — one struct drives everything)

| era (age band)                       | street pattern      | main/alley w | garden | extras                    | house    |
|--------------------------------------|---------------------|--------------|--------|---------------------------|----------|
| organic (primitive, stone)           | random walk         | 1 / 0        | patches| —                         | hut      |
| ancient (bronze, iron, classical)    | radial + coarse grid| 2 / 1        | courts | central plaza / ziggurat  | mudbrick |
| castle (medieval, renaissance)       | winding + quarters  | 2 / 1        | kitchen| walls + keep, market sq   | timber   |
| zonedgrid (colonial, industrial, victorian) | strict zoned grid | 3 / 1  | formal | prod-left / res-right     | rowhouse |
| cityblocks (electric, atomic, modern)| superblock grid     | 4 / 2        | parks  | civic center              | tower    |
| campus (information, digital, cyberpunk, fusion) | hex campus + links | 2 / 1 | green | neon accents          | arcology |
| orbital (space, interstellar, galactic, quantum) | concentric rings | 2 / 1 | dome parks | central hub          | dome     |

(Era bands = the existing `eraForAge` index buckets in layout.go.)

## Rendering

terrain → streets (paved role, width) → block interiors (garden green / plaza
tone) → lots (2.5D volume via drawVolume; per-kind palette) → landmark labels
(existing overlay label pipeline). Theme-aware (every color via a theme role),
panic-safe, correct output size.

## Phasing

- **A**: framework (cityPlan / eraStyle / pipeline / count-driven populate) +
  `organic` village + `ancient` gridded city. Other eras fall back to a sane
  default preset. Route the render through the new generator (supersede the old
  per-era layout strategies). Tests.
- **B**: castle + zonedgrid + cityblocks.
- **C**: campus + orbital + per-category building sprites + wonder scaling.
- **D**: gardens/parks/alleys polish, density tuning, review refinements.

## Keep

- Named landmarks (one labeled hero per built type) — now embedded in the fabric.
- Terrain + land-gating + pathfind road/street routing.
- Theme-awareness, half-block render, correct-size, panic-safety.

## Worldmap faction cities (already working)

`DiplomacyManager.DiscoverFactions` reveals a civ at its `MinAge`; `worldCivs`
filters to `Discovered` and labels each dot with `FactionInfo.Name`;
`civSignature` re-renders on discovery. Future enhancement (optional, ties to
Trello NbT9WbNB / SGMjUNnd): give each faction a distinct generated **capital
city name** rather than reusing the faction name.
