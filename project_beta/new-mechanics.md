# New Mechanics — AgeForge Beta

## Design Principle

Every new mechanic in Beta must pass one test: **does it enhance the feeling of operational command?** AgeForge is about managing a civilization as if it were a system you are administering. New mechanics should deepen that feeling — more output to parse, more commands to issue, more events to respond to — not dilute it with genre conventions borrowed from city builders or idle games.

The engine change is not an excuse to become a different kind of game. It is permission to become a more complete version of the game AgeForge always was.

---

## Visual Storytelling

### Age Advancement Cinematic

The age advancement splash in Classic is functional but minimal. Beta transforms it into a moment.

**Structure:**
1. The city map does not disappear — it continues rendering, but a dark overlay fades over it (80% opacity)
2. The terminal panel clears and a horizontal rule scrolls in
3. ASCII art of the new age name renders line by line, each line appearing with a 30ms delay (the "printing" effect)
4. Below the art: a brief flavor text paragraph describing the historical era
5. A list of unlocks scrolls in: new buildings, new resources, new research, new commands
6. A subtle border animates around the screen — the phosphor glow turns up for 2 seconds, then returns to normal
7. The ambient audio crossfades to the new age's track
8. Dismiss with any key or wait 6 seconds

Each age has unique flavor text. For example, the Industrial Age:

```
> Advancing civilization...
> [████████████████████████] 100%

  ██╗███╗   ██╗██████╗ ██╗   ██╗███████╗████████╗██████╗ ██╗ █████╗ ██╗
  ██║████╗  ██║██╔══██╗██║   ██║██╔════╝╚══██╔══╝██╔══██╗██║██╔══██╗██║
  ██║██╔██╗ ██║██║  ██║██║   ██║███████╗   ██║   ██████╔╝██║███████║██║
  ██║██║╚██╗██║██║  ██║██║   ██║╚════██║   ██║   ██╔══██╗██║██╔══██║██║
  ██║██║ ╚████║██████╔╝╚██████╔╝███████║   ██║   ██║  ██║██║██║  ██║███████╗
  ╚═╝╚═╝  ╚═══╝╚═════╝  ╚═════╝ ╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚══════╝

  ─────────────────────────────────────────────────────────────────────
  The furnaces never sleep. Coal and iron transform into machines that
  transform everything else. The world your ancestors built with their
  hands will now be built a thousand times faster by engines that never
  tire. The price of progress is written in smoke.
  ─────────────────────────────────────────────────────────────────────
  UNLOCKED:
    Buildings  → Coal Mine, Steam Engine, Factory, Rail Depot, Print Shop
    Resources  → coal, steel, oil
    Research   → Industrial Efficiency, Mass Production, Steam Power
    Commands   → deploy_rail, establish_factory
  ─────────────────────────────────────────────────────────────────────
  [Press any key to continue]
```

### City Skyline Transformation

As ages advance, the city map undergoes a visible transformation:
- Building sprite sets change (stone huts → brick buildings → steel towers → glass spires → energy constructs)
- Background image transitions (wilderness → farmland → town → city → megacity → orbital habitat)
- Road network expands (foot paths → cobblestone → paved → highway)
- Color palette shifts (warm natural tones → industrial grey → steel blue → plasma white)

The player watching their city transform across 22 ages is the long-form visual narrative of AgeForge. It requires no dialogue, no cutscenes — the building sprites and background art carry the story.

### Wonder Construction Time-Lapse

When a wonder completes construction, a 3-second animated sequence plays on the city map:
1. The wonder's building slot pulses with a golden light
2. The wonder sprite assembles from the bottom up (multi-frame construction animation)
3. A particle burst fires (gold sparkles, age-appropriate)
4. The city map camera briefly zooms toward the wonder, then zooms back

In the terminal panel simultaneously:
```
> WONDER COMPLETE: The Great Library
  Civilization benefit: +25% research speed
  +2 culture/tick
  Achievement unlocked: "Keeper of Knowledge"
```

### Building Animations

**Idle animations:**
| Building Type | Animation |
|--------------|-----------|
| Mill | Blades rotating at 4 FPS |
| Forge/Factory | Orange glow pulse, smoke particles |
| Farm | Gentle crop sway (shader effect on sprite) |
| Library/University | Floating book particle (subtle) |
| Cathedral/Temple | Window light flicker |
| Barracks/Fort | Guard patrol (small figure sprite cycling) |
| Observatory | Telescope rotation |
| Mine | Pickaxe swing (worker silhouette) |
| Trade Post | Goods crates appearing/disappearing |
| Wonders | Unique per wonder |

**Worker presence indicator:** Buildings with workers assigned show small silhouette sprites at the building entrance. Count scales with `WorkersAssigned`. An empty building looks abandoned. A fully staffed forge is a hive of activity.

---

## Steam-Specific Features

### Achievement System

AgeForge's existing milestone system maps perfectly to Steam achievements. The 33 milestones become 33 Steam achievements. Additional achievements:

**Age progression (22 achievements):**
- "Dawn of Fire" — Reach Stone Age
- "Forged in Bronze" — Reach Bronze Age
- ... one per age, named after the age

**Prestige achievements:**
- "New Blood" — First prestige
- "Eternal Cycle" — 5 prestiges
- "The Endless Engine" — 10 prestiges

**Challenge achievements (hidden until earned):**
- "Hermit Kingdom" — Reach Classical Age without any trade routes
- "Pacifist" — Reach Industrial Age without any military units
- "Speed of History" — Reach Iron Age within 30 minutes of real time
- "The Great Survivor" — Endure 5 catastrophes in a single run
- "Succumb and Transcend" — Succumb to a catastrophe and reach a higher age than your previous run
- "Wonder of Wonders" — Build all wonders available in a single age
- "Full Employment" — Have 100% worker assignment across all domains simultaneously
- "Technological Apex" — Research all 52 technologies in a single run

**Total: ~65 achievements.** This is a healthy number — enough to give dedicated players long-term goals without padding.

### Steam Trading Cards

One card per era (groups of 2-3 ages). 7 trading card designs, styled as terminal printouts from that era:

- **Card 1: Primitive** — A gather command printout with forest ASCII background
- **Card 2: Metalworking** — Forge output log with ore smelting stats
- **Card 3: Classical** — Senate trade manifest with faction names
- **Card 4: Medieval** — Castle garrison report
- **Card 5: Industrial** — Factory production manifest, carbon copy styling
- **Card 6: Modern** — Data center resource allocation printout
- **Card 7: Galactic** — Starship manifest in "SYSTEM TERMINAL" styling

Foil cards: same designs with the phosphor glow effect applied maximally — they genuinely look like glowing CRT screens.

### Steam Cloud Saves

Implemented via GodotSteam's `Steam.FileWrite()` API (see `architecture.md`). Up to 3 save slots synced. Save file size is ~50KB — trivially within Steam's cloud limit.

### Global Stats

Via Steamworks `ISteamUserStats`, track global aggregate stats displayed in the Stats overlay:
- "X% of AgeForge players have reached the Industrial Age"
- "X% have completed their first prestige"
- "Average playtime to first wonder: Y hours"
- "Most popular first tech researched: [tech name]"

These require defining stats in Steamworks developer dashboard. Display them in the in-game Stats overlay and in the Steam page's stats section.

### Steam Workshop (v1.1+, not Beta 1)

Allow players to create and share custom age/building configuration packs. A Workshop item is a JSON data pack that overrides specific building stats, age requirements, or resource rates. The game validates the pack on load (checking for required keys, valid ranges) and applies it as an overlay to the base config.

**What Workshop packs can modify:**
- Building costs and production rates
- Age advancement requirements
- Tech costs and effects
- New resource names (with custom display names)
- Alternate flavor text for ages

**What they cannot modify (hardcoded):**
- Save format
- Steam achievement triggers
- Engine architecture

This is deferred to v1.1. Beta 1 ships without Workshop.

---

## Enhanced Command Interface

### Inline Ghost Text Autocomplete

When the player types `bui`, the ghost text `ld` appears in gray after the cursor, completing to `build`. Pressing Tab accepts it. Typing continues to filter.

When there are multiple matches, a compact popup appears below the cursor:

```
> assign g▌
              gathering_camp   (food workers, 5 slots)
              granary          (storage building)
              guardhouse       (military workers, 3 slots)
```

The popup is navigable with Tab/Down arrow. It shows the building's domain and worker capacity inline — the player doesn't need to run `info gathering_camp` to remember what it takes.

### Command Syntax Hints

As the player types a command, a hint line appears above the input row showing the full syntax and available arguments:

```
  Syntax: assign <building> [count|all]
  Example: assign forge 3
> assign f▌
```

This replaces the need to type `help assign` repeatedly. It appears after the first word is typed and matches a known command.

### Multi-Line Paged Output

Some commands produce long output (e.g., `list buildings` with 284 buildings). Beta adds pagination:

```
> list buildings
Showing 1-20 of 284 buildings. [n]ext page, [p]rev, [f]ilter <term>

  gathering_camp    Food     primitive_age    0.50/tick    Cost: 10 wood
  hut               Housing  primitive_age    +2 pop cap   Cost: 5 wood
  ...
```

The `[n]`, `[p]`, `[f]` actions are typed commands, not mouse clicks. The terminal interface remains pure keyboard.

### Command Aliases

All 40+ commands can have user-defined aliases stored in `Settings`. For example:
```
> alias b build
> alias r research
> alias a assign
```

Common aliases are suggested in the tutorial and can be reset to defaults. This is a quality-of-life feature for power users.

---

## New Game Systems

### Diplomacy Events — Diplomatic Cables

Factions periodically send diplomatic cables to your civilization. These appear as terminal messages scrolling into the output log:

```
──────────────────────────────────────────────────────
INCOMING DIPLOMATIC CABLE — IRON THRONE FACTION
──────────────────────────────────────────────────────
To: [Your Civilization]
From: Chancellor Valdrin, Iron Throne

Our smiths have observed your forges with admiration.
We propose a mutual defense accord: share your iron
reserves with us, and our armies stand ready at your
borders. Decline, and we note your... independent
spirit.

Respond within 30 ticks.
  > accept treaty_iron_throne
  > decline treaty_iron_throne
  > counter treaty_iron_throne 50
──────────────────────────────────────────────────────
```

The cable includes response commands printed directly in the output. The player can type them or ignore the cable (declining by default after the timer expires). This is pure keyboard interaction — no buttons, no dialog boxes. The message looks like a terminal email.

**Implementation:** `DiplomacyEvent` struct with `Sender`, `Text`, `Timeout`, `Options []DiplomacyOption`. Each option maps to a command. The event fires via `Bus.EmitDiplomacyEvent()`.

**Frequency:** 1-3 cables per age. Not spam. Each one matters.

### Covert Operations — The `spy` Command

New in Beta: a spy command tree for intelligence and sabotage.

```
> spy intel iron_throne        — cost 50 gold, reveals their resource levels
> spy steal iron_throne tech   — cost 200 gold, 40% success rate, steals a researched tech
> spy sabotage iron_throne     — cost 300 gold, destroys their trade route with you
> spy status                   — shows active operations and outcomes
```

Results are delivered as terminal messages ("INTELLIGENCE REPORT — OPERATION IRONWATCH"):

```
INTELLIGENCE REPORT — OPERATION IRONWATCH
CLASSIFICATION: EYES ONLY
──────────────────────────────────────────
Target: Iron Throne Faction
Objective: Resource Assessment
Status: COMPLETE

FINDINGS:
  Iron reserves: 847 units (CRITICAL: stockpiling for military expansion)
  Gold reserves: 2,300 units
  Active military: 45 warriors, 12 cavalry
  Alert level: ELEVATED (consider delaying further operations)

Report filed by Agent Maren. Operation cost: 50 gold.
```

**Balance:** Spy operations cost gold (the universal currency from later ages). They have cooldowns. Getting caught (failed operations) damages faction relations and may trigger a retaliatory event.

**Why this fits the theme:** Intelligence is another form of operational data. You are reading reports. The information shapes your commands. It is entirely in the terminal aesthetic.

### City Districts

New command: `zone <area> <type>` — assign areas of the city map to functional zones:
- `zone north industrial` — northern sector focused on production
- `zone south residential` — southern sector for housing, increases population growth
- `zone east cultural` — cultural zone, bonus to culture output
- `zone west military` — military zone, faster training speed

Zones are visible on the city map as colored overlays (subtle transparent tints). They do not change the road network or building placement — they are designations that apply buffs.

**Why zones?** It adds a light strategic layer to city development without requiring drag-and-drop placement. The commands are typed. The effect is visible on the map. It deepens the city management dimension without abandoning the keyboard-first interface.

### Trade Convoy Visualization

When a trade route is active, the city map shows small animated convoy sprites moving along the road network from the civilization's border toward the trading partner's icon on the edge of the map. The convoy moves slowly, reaching the edge of the map over ~30 seconds, then restarting.

- Land routes: small cart sprite
- Sea routes (later ages): ship silhouette
- Air routes (modern ages): airplane sprite
- Digital routes (digital age): data packet icon (blinking dot moving along a glowing line)

This is pure visualization — no gameplay interaction required. It makes active trade routes tangibly visible and rewarding. When you have 8 active trade routes, the map is alive with movement.

### Wonder Reveal

When any wonder completes, a brief 4-second full-screen effect:

1. The city map camera zooms toward the wonder
2. A glowing border frames the wonder sprite
3. The terminal panel displays the wonder's name in large ASCII art with the glow shader at maximum
4. A distinctive sound plays (each wonder has a unique 2-second audio sting)
5. The effect fades; the game resumes

It is the one moment where the game interrupts normal play to say: **what you just built matters.**

---

## Monetization

### Base Game: $4.99

The impulse-buy price point is deliberate. AgeForge Beta is an indie game with niche appeal. The player acquisition strategy depends on:
1. Being cheap enough that any curious person buys it without research
2. Being good enough that those buyers recommend it to others
3. Being unique enough to generate organic "what is this" content on social media

At $4.99, the break-even for a small dev operation is a few thousand sales. At $9.99, the barrier to impulse purchase doubles and the player count (and thus Steam visibility) halves.

### Optional DLC (v1.1+)

DLC must follow the AgeForge aesthetic and design principles. No content gating, no FOMO.

**Cosmetic Terminal Themes Pack — $1.99**
Four additional color themes beyond the three base themes:
- Dracula Dark (purple/pink, based on the popular Dracula theme)
- Nord (cool arctic blues, based on the Nord color palette)
- Solarized (the classic Solarized terminal palette)
- Hacker Green (hyperintense green, minimal CRT effect — pure ANSI aesthetic)

**Extended Soundtrack — $2.99**
A full 22-track OST by a credited composer — one track per age, full-length versions of the ambient themes with additional instrumentation.

**Alternate History Pack — $3.99**
A content DLC adding an alternate history branch: instead of following the standard age progression, a "Cold War pivot" unlocks at the Modern Age, branching into:
- Surveillance State age (panopticon infrastructure, loyalty resources)
- Space Race age (orbital facilities, satellite networks)
- Collapse Reconstruction age (scarcity management, reclamation tech)

This is a content expansion, not cosmetic. It adds 3 new ages, 40+ new buildings, 8 new techs. It earns its price.

**No microtransactions.** No gacha. No season passes. No "Civilization Coins." AgeForge's audience is adult nerds with high disposable income and low tolerance for dark patterns. The way to monetize that audience is with quality content at fair prices.

### Free Updates (Forever)

The base game will receive balance updates, bug fixes, new events, and minor content additions indefinitely. This is standard practice for well-regarded indie games and is essential for Steam review health. Every update is an opportunity for Steam visibility (update news feeds) and review refresh.
