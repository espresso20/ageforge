# Map concepts

In September 2026 the live map was rebuilt a seventh time without ever feeling right, so we prototyped six different directions in a "Map Lab" and compared them as real terminal captures.

**Chosen:** `roguelike` (a glyph world with an inspect cursor, and the default) and `skyline` (an ANSI-art panorama). Players switch between them. Both are built on one shared map model, with a thin renderer per style. The work starts with the `feat/espresso/map-v2-core` PR (phase 1) and is followed by integration (phase 2), which replaces `ui/citymap`.

**Saved for later:** the four concepts below were strong but not chosen for now. Their specs are kept here so any of them can come back as another selectable style. Once the shared map model exists, a new style only needs a renderer that reads the model and implements the style interface.

| Concept | Idea | Place or information | Prototype branch |
|---|---|---|---|
| [mud](mud.md) | The settlement as a small text world you walk through (`look`, `go`, `visit`), with generated room prose and a box-drawn road map | Place | `lab/map-mud` |
| [organism](organism.md) | The civilization as one living tree: limbs are lineages, the trunk is population, and roots grow through strata of history | Information | `lab/map-organism` |
| [topology](topology.md) | The economy as a live system diagram, with packets flowing from producers through a resource bus to services and other civilizations | Information | `lab/map-topology` |
| [atlas](atlas.md) | The known world as a typeset atlas that gains a new cartographic plate each epoch, from charcoal on hide to a galactic star chart | Place (world) | `lab/map-atlas` |

Ideas from these already feed the chosen styles:
- **atlas:** its epoch plates style roguelike's region zoom.
- **topology:** its flows view becomes a toggle on the roguelike map.
- **mud:** its "while you were away" digest becomes the map's news line.

## Reading the specs

Each spec is the prototype's own DESIGN.md, copied from its `lab/map-*` branch. It covers the core idea, how it meets the map requirements checklist, how it looks in each epoch, light and dark themes, small terminals, performance, integration, the estimated effort to productionize it, and risks. The prototypes build and run from their branches:

```bash
git checkout lab/map-organism
go run ./lab/organism
```

Each branch's DESIGN.md gives its exact flags. The branches hold the working code, generated game states and captures.

[history-and-requirements.md](history-and-requirements.md) is the research behind the lab. It gives the timeline of all seven earlier map iterations, why none of them felt right, and the requirements checklist any map for this game should meet. Read it before building a new style.
