# Citymap v3 — Top-Down Pixel-Art Living City (LOCKED SPEC)

Supersedes the Phase-A isometric / terrain-gated model after playtest. The
citymap is **NOT** a world/planet geo map and **NOT** isometric blocks planted on
terrain. It is a **top-down pixel-art city** — you look straight down at ROOFS,
streets, gardens and squares, like a top-down village/city game (Stardew,
top-down Zelda). It grows and restyles as the player's civ grows.

Every decision below is LOCKED from a design pass. Do NOT re-assume — if
something here is silent, ask, don't guess.

## Locked decisions
1. **View**: top-down, roofs-from-above. No isometric walls/sides (subtle shadow
   only). No world terrain, oceans, or continents on the city view.
2. **Ground**: neutral, ERA-TINTED — base color/texture shifts per era mood:
   earthy (primitive/ancient) → stone (medieval/renaissance) → concrete/asphalt
   (industrial/modern) → dark neon-grid (cyber) → pale metallic (space). No
   natural water. ALL greenery is BUILT (gardens/parks/squares/street-trees).
3. **Scale**: FILL-FRAME, densifying. The whole city always fits the ~200×100px
   canvas; more buildings → denser fabric + finer streets. Late age = packed
   metropolis. Auto-fit, no pan/zoom.
4. **Count fidelity**: NEAR 1:1 at low counts (24 huts ≈ 24 hut-roofs),
   SUB-LINEAR at high counts (hundreds → dense but legible, not N identical
   clones).
5. **Buildings**: each built type drawn as count-scaled TOP-DOWN ROOF sprites —
   query building counts → math → that many roofs of that type.
6. **Roof art**: age-realistic MATERIAL (thatch → clay tile → slate → flat modern
   → neon/metal) + a SUBTLE lineage tint so a temple reads different from a
   workshop. Soft drop-shadow (SE) for a hint of height. Muted & natural
   saturation, within the active theme.
7. **Labels**: KEY LANDMARKS ONLY (city center, wonders, a promoted hero when the
   civ has no civic building yet). Everything else unlabeled — read by roof
   shape/color.
8. **Layout**: PERSISTENT, grows in place. One deterministic layout per civ
   (stable seed). New buildings SLOT INTO the existing layout without moving the
   old ones (stable incremental placement — instance #N keeps its slot as #N+1 is
   added). The relative layout is stable; the fill-frame scale re-fits as it grows.
   Re-skins to the era on age-up; bones persist.
9. **City edge**: WALLS WHERE THEY FIT — walls+gates ring the built-up area in
   walled eras (ancient, medieval/renaissance); open ragged sprawl for industrial+.
10. **Streets**: ORGANIC → GRID → AVENUES. Winding dirt lanes (village) →
    tightening grid (classical/medieval) → wide boulevards + superblocks
    (modern/cyber). Per era.
11. **Growth**: MULTIPLE DISTRICTS but LOOSE — buildings gravitate toward
    same-kind clusters (residential / production / civic / market / garrison)
    that BLUR into each other, not hard-edged zones.
12. **Detail**: BALANCED living-city filler — gardens, squares, trees, wells,
    stalls, statues: alive and lived-in but never burying the buildings.
13. **Wonders**: DOMINANT CENTERPIECE — a large, ornate, unmistakable complex
    anchoring the city center, clearly the grandest thing on the map.
14. **Age-up**: GRADUAL RE-SKIN — roofs/streets/ground restyle to the new era;
    layout/bones persist.
15. **Minimap**: full `citymap` panel only for now (no dashboard minimap yet).

Cross-cutting: theme-aware (retint on theme switch), panic-safe, exact output
size, deterministic from the civ seed.

## Architecture

### Deterministic persistent layout
- `citySeed` = stable per civ (hash of display name / account), AGE-INDEPENDENT —
  the bones don't move across ages.
- Placement is a STABLE SEQUENCE: each building instance takes the next slot from
  a seeded space-filling sequence (golden-angle spiral / seeded Poisson) anchored
  on its district cluster, in a fixed deterministic order. Adding the 25th hut
  must NOT move the first 24.
- Fill-frame: compute the built-up footprint from total count (sqrt), then scale
  the whole plan to fill the canvas (roofs shrink as the city grows, staying
  legible). Relative positions stable; absolute scale re-fits.

### cityPlan (reuse + extend the Phase-A skeleton)
- `streets []street{pts, width, class(lane|street|avenue)}` — laid by the era
  street generator (NO terrain routing; no water).
- `districts []district{kind, center, members…}` — loose clusters.
- `lots []lot{x,y,w,h, kind, domain, tier, roofType, label}` where kind ∈
  {house, workshop, civic, market, garrison, landmark, wonder, garden, square,
  tree, prop, wall, gate}.
- ground = era-tinted fill (+ subtle texture).

### Pipeline (pure, deterministic)
1. gather built buildings → {key, domain, category, tier, count}; classify into
   district kinds.
2. size → target footprint from sqrt(total); fill-frame scale factor.
3. districts → seed loose district cluster centers (stable).
4. streets → era pattern (organic / grid / avenue) linking district centers + core.
5. populate (count-driven, near-1:1-low): each type emits count-scaled roof lots
   into its district cluster via the STABLE sequence; wonders → central dominant
   complex; leftover space → gardens/squares/trees/props at balanced density.
6. walls → if the era has walls, ring the built-up area with wall + gates.
7. (no terrain gate — neutral ground.)

### Rendering (top-down)
era-tinted ground (+texture) → streets (paved per class + era material) →
district ground accents (plaza stone, garden green) → ROOF SPRITES per lot
(top-down roof atlas: shape by roofType, material by era, subtle lineage tint,
soft SE drop-shadow; draw back-to-front so shadows layer) → props/trees →
walls/gates → landmark labels (overlay: city center, wonders, promoted hero).
Every color via theme roles; panic-safe; exact size.

### Top-down roof atlas (drawn top-down, filling the lot, ridge/texture hint + SE shadow)
- hut: small round/oval thatch roof, radial streaks.
- house: rectangle pitched roof — center ridge, two shaded slopes.
- rowhouse/longhouse: elongated ridge roof.
- temple/shrine: larger ornate symmetric roof + finial.
- market: open awning grid / stalls.
- workshop/factory: flat/low roof + chimney dots.
- tower/keep: small footprint, longer shadow, crenellation dots.
- civic/library: broad roof + rooftop detail.
- dome/observatory: circular roof + highlight.
- skyscraper: small footprint, LONG shadow, rooftop AC/helipad dots.
- arcology/cyber: angular neon-edged roof.
- wonder: large multi-part ornate complex (centerpiece).
Material palette by era (thatch/wood → clay tile → slate/lead → asphalt/steel →
glass/neon), muted, theme-derived; lineage tint blended ~15–25%.

### Era style presets (per eraForAge band)
| era band                         | ground        | streets              | walls          | roof material  | house    | wonder        |
|----------------------------------|---------------|----------------------|----------------|----------------|----------|---------------|
| organic (primitive, stone)       | earthy dirt+grass | winding dirt lanes | none           | thatch/wood    | hut      | (rare)        |
| ancient (bronze, iron, classical)| packed earth/stone | radial + coarse grid | mudbrick+gates | clay tile      | mudbrick | ziggurat      |
| castle (medieval, renaissance)   | cobble/stone  | winding + market sq  | stone+towers+gate | slate/tile  | timber   | cathedral     |
| industrial (colonial, industrial, victorian) | cobble→asphalt | strict grid | none (open)  | slate→tin      | rowhouse | expo hall     |
| modern (electric, atomic, modern)| asphalt       | wide avenues + superblocks | none     | flat modern    | tower    | skyscraper    |
| cyber (information, digital, cyberpunk, fusion) | dark neon-grid | boulevards + megablocks | none | neon/glass  | arcology | megastructure |
| space (space, interstellar, galactic, quantum) | pale metallic | ring/spoke arcs | dome ring    | metal/glass    | dome     | central spire |

## Phasing
- **V3-A (first cut, review)**: the whole top-down ENGINE (era-tinted ground,
  streets, stable persistent placement, fill-frame, top-down roof atlas w/ shadow,
  loose districts, balanced filler, landmark labels) + the PRIMITIVE/STONE village
  fully tuned. Other eras use a reasonable default preset. Drop world terrain +
  isometric volumes from the citymap render path. Tests.
- **V3-B**: ancient (mudbrick walls+gates, clay roofs, radial-grid, ziggurat) +
  castle (stone walls+towers, market square, cathedral).
- **V3-C**: industrial + modern (grid → avenues, rowhouse → tower, no walls).
- **V3-D**: cyber + space (neon / dome) + wonder-centerpiece polish + per-age props.
- **V3-E**: density / filler tuning + review refinements.

## Keep / drop
- KEEP: the cityPlan/eraStyle skeleton, count-scaling math, gatherBuildings +
  classify, the overlay label pipeline, theme-awareness, half-block render,
  panic-safety, correct-size.
- DROP from the CITYMAP path only: world terrain background, terrain-routed A*
  streets, isometric drawVolume, land-gating.
- The WORLDMAP keeps its terrain (approved, separate) — terrain.go stays; only the
  citymap stops calling it.

## Worldmap (approved, unchanged by this work)
"World map and its generator? much better." Faction cities already name on
discovery (`worldCivs` → `FactionInfo.Name`, gated on `Discovered`). Future
optional flavor (Trello NbT9WbNB / SGMjUNnd): distinct generated capital-city
names per faction.
