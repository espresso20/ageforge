# The City Map

Open it with `citymap` (or `map`, or the `m` tab shortcut). The **City Map** is a procedural, top-down pixel-art rendering of your civilization as one living settlement — you look straight down at the roofs, streets and squares of a city built from your **actual building counts**. It is:

- **Deterministic** — the same civilization always draws the same city. Layout is stable and grows in place; new buildings slot into the existing fabric rather than reshuffling it.
- **Theme-aware** — every colour is pulled from your active [theme](themes.md), so switching themes retints the whole city instantly.
- **Panic-safe** — the renderer never crashes the game; a bad draw degrades gracefully instead of taking the tab down.
- **Gradual** — the city re-skins to the current era as you advance. Nothing snaps; roofs, ground, walls and props restyle age by age while the bones of the city persist.

There is no world terrain here — biomes and neighbours live on the [World Map](commands.md#world-map) (`worldmap`). On the City Map every green thing (gardens, ponds, street-trees) is **built**, not landscape.

---

## A distinct look for every age

The headline: the city's **visual identity evolves per age**. All 22 ages have their own roof materials, ground surface, house forms, walls (or lack of them), wonder centerpiece, and density — a thatch-hut hamlet and a neon megablock are unmistakably different places, and every age in between is its own step.

The tables below group the 22 ages by their **7 epochs**. Each row is a one-line sketch of what that age's city looks like from above.

### ◈ Stone Era

| Age | The look |
|---|---|
| **Primitive** | A thatch-hut village on earthy dirt, winding lanes, no walls. |
| **Stone** | A rocky grey highland of thatch huts and sparse trees, around a megalithic stone-circle monument. |
| **Bronze** | Warm terracotta clay-tile roofs over mudbrick houses, ringed by a mudbrick curtain wall, anchored by a stepped ziggurat. |

### ⚔ Iron Era

| Age | The look |
|---|---|
| **Iron** | Cooler grey-tinged clay roofs behind a brown timber palisade, a fortified keep with a watchtower at the heart. |
| **Classical** | White-stone houses under terracotta roofs, columns and pale marble paving inside stone walls, crowned by a columned temple. |
| **Medieval** | Blue-grey slate roofs over timber houses on cobbled streets, stone walls with towers, and a cathedral. |

### ⚙ Steel Era

| Age | The look |
|---|---|
| **Renaissance** | Ornate cream ashlar stone under a great dome, wrapped in an angular star-fort wall. |
| **Colonial** | Warm brick-red terraced rowhouses on dirt lanes, a timber palisade fort, and a statehouse. |
| **Industrial** | Grimy red brick and dull tin roofs over sooty ground, smokestacks and a great factory — walls gone, open sprawl. |

### ⚡ Electric Era

| Age | The look |
|---|---|
| **Victorian** | Warm brownstone rowhouses on stone pavers, threaded with gas-lit green parks. |
| **Electric** | Pale art-deco concrete flat blocks along wide avenues, lit by a warm electric glow, topped by a setback tower. |
| **Atomic** | Cooler pastel midcentury concrete-and-steel — cleaner, airier, suburban. |

### ▣ Digital Era

| Age | The look |
|---|---|
| **Modern** | Cool blue-grey glass-and-steel skyscrapers along steel avenues. |
| **Information** | Denser glass towers, colder in tone, studded with low data-center blocks. |
| **Digital** | Darker, sleek glass with the first neon — cyan and magenta — bleeding into the streets. |

### ◉ Neon Era

| Age | The look |
|---|---|
| **Cyberpunk** | Dark, dense megablocks drowning in saturated neon, hung with holograms. |
| **Fusion** | Clean, bright white minimalist towers around pale-cyan glowing reactor cores. |
| **Space** | Pale metallic domes beside a rocket launchpad. |

### ✦ Cosmic Era

| Age | The look |
|---|---|
| **Interstellar** | A dark starfield deck bristling with pale metallic spires. |
| **Galactic** | A lit metallic megastation with grand concentric orbital rings around a glowing hub. |
| **Quantum** | A dark deck scattered with iridescent (shifting cyan/magenta/gold) crystal nodes under a crystal lattice. |
| **Transcendent** | A luminous, ethereal field of soft light-forms rising toward an ascension of pure light — the final age. |

---

## Landmarks & wonders

Only key landmarks are **labelled**, as soft pill banners that stay readable over any roof — the City Center, your built [wonders](wonders.md), and a promoted hero building when you have no civic building yet. The rest of your city reads by roof shape and colour rather than a wall of text.

The **dominant wonder anchors the city center**: whichever wonder you've built looms largest and the town hugs it, drawn as an era-appropriate centerpiece (a ziggurat in the ancient ages, a cathedral or keep in the medieval ages, a reactor core or orbital ring in the far future). Build more wonders and the grandest one holds the middle.
