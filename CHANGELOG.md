# Changelog

All notable changes to AgeForge are documented here.

---

## [Unreleased]

### Added
- **Build plan** (`plan`): queue builds, techs, trades and an advance, and the game starts each one, in order, as the resources come in, while you play and while you are away. Items are paid for when they start, not when you add them. A waiting item holds back its next price from the items below it, so the order is your priority, but items below still start with what it doesn't need. Techs start one at a time in plan order (a research queue). `plan trade <from> <to> [amount]` sells as the resource comes in once the market has recovered from the last sale; `plan advance` advances as soon as the age is ready, and the next age's buildings and techs can be planned to wait for it. Items that can never start drop out with a log line, copies the plan builds are staffed from idle workers, and the plan is saved and cleared by prestige, Succumb and a new game. The Plan panel shows each item's price and status, with keys to reorder and remove. See the new Build Plan wiki page.
- **Offline catch-up runs in one-minute steps**: the plan starts items as the day's income pays for them, the caps apply along the way, and construction and research finish while you are away (they used to wait for you to come back). A day away takes milliseconds; with nothing planned or under construction it pays what it always did.
- **Wonder overflow** (`wonder overflow on|off`, on by default): production a full store would waste goes into the current age's wonder bank, up to what the wonder still needs, during play and offline. Shown in the Wonders overlay and saved.
- **The Last Passage**: the Cosmic Era's passage is now prestige itself. Confirming prestige in the Cosmic Era rolls once with the epoch-catastrophe odds (18% / 15% / 12% by faith fill, ×0.6 per level of Appease, certain if invited). If nothing comes, the verdict is Spared and prestige completes. If it comes, prestige waits on a choice in the catastrophe-modal style (✦ The Last Passage): Esc closes it, a ☄ LAST PASSAGE status-bar badge shows, and a bare `catastrophe` reopens it. Only prestige is blocked, and the choice is saved with your game. **Endure** completes the prestige keeping 50% of the run's prestige points (70% / 85% braced), rounded down. **Succumb** completes it with no points from the run and grants the Cosmic Legacy. Both add a civilization-log entry and count in the Survived / Succumbed tallies. Prestige from before the Cosmic Era is unchanged. `prestige` shows the chance and the warning figure, `prestige confirm` explains both outcomes, and the `catastrophe` outlook and Epoch tab show the next passage.
- **Cosmic Legacy**: Succumbing to the Last Passage grants a one-time, permanent +10% production, shown as "Cosmic Legacy" in the Stats overlay (Active Multipliers and Legacy Bonuses). It survives every prestige and Succumb; only a full wipe clears it. Once held, Succumb at the Last Passage is closed.
- **Cosmic Era harbinger thread**: the Distress Beacon, the Elder Relay, your future self and your unmade self now appear, warning of the Last Passage from the Interstellar Age on, so all 22 roster figures speak. Odds are published, there are no false prophets, and Appease / Brace levels carry across the handoffs. Appease costs 1.2B faith + 19B culture and Brace 1.56T dark matter + 75.6B titanium (level 2 double), the same in every Cosmic age. In the Cosmic Era, Brace raises Endure's points share and Invite is how to choose the Cosmic Legacy on purpose.
- **Closing lines at prestige**: every prestige now logs one closing flavor line in the voice of the age the run ended in. Modern through Space Age runs wind down plainly; Cosmic Era runs end with the age's harbinger present.
- Dev console: `/lastpassage` makes the Last Passage pending (final epoch only), for testing.
- **The Harbinger**: every epoch whose transition can bring a catastrophe (Stone to Neon Era) now has a harbinger thread, from its first age until the transition. Each age's figure takes up the warning in turn (the Stone Era goes the Wild Man, the Hermit, the Soothsayer), and a new game, Succumb or prestige is greeted on the first tick. Each figure brings a log entry, a toast and a ⚑ HARBINGER status-bar badge; nothing blocks and nothing expires. Open the panel with `harbinger` (or `harb`). Answers belong to the passage and carry across figures, and the price is the same in every age of the epoch. **Appease** costs a quarter of what a moderate faith economy makes over the epoch's ages, in faith, and the same in culture from the Steel Era on, with level 2 at double, and multiplies the real catastrophe chance by 0.6 per level. **Brace** costs 12% (24%) of the most the epoch still asks of each core resource, so an Endure destroys 15% (10%) of buildings instead of 20% and keeps 30% (45%) of stored resources instead of 15%. **Invite** is free and guarantees the catastrophe, for a deliberate Succumb. Stone, Iron and Steel Era threads may be false prophets (8/64, 5/64, 2/64), repeating one false claim through every figure; from the Industrial Age on the panel prints the odds. The transition records a verdict (Fulfilled, Vindicated, Spared or Discredited), and the Epoch tab lists each thread's chain of figures. Saved with your game.
- **Harbinger content (data and prose only; the mechanic ships separately)**: a roster of 22 harbingers, one per age, from the Wild Man to your own unmade self, each with its own Appease, Brace and Invite action names, a forecast that turns from omens to published odds at the Industrial Age, and a false-prophet chance that falls to zero by then. About 1,500 new hand-written flavor sentences cover the harbinger's arrival, a warning in four levels of fear, appeasing, bracing, inviting the doom, and each outcome, written in the speaker's own voice and never printing a number.
- Two map views: `citymap` (your settlement, was `map`; `map` still works) and a new `worldmap` — a Game-of-Life-style world of settlements with your civilization and the diplomacy civs called out as labeled, relationship-colored dots, theme-aware.
- **Humor & personality layer (buildings, ages, events, milestones, logs)** — buildings now carry an optional cosmetic *flavor* line, rendered as a dim italic note beneath the functional description in both the Buildings overlay and the Economy tab (costs, rates, and worker counts are untouched — flavor is purely additive). All 300 buildings got a line, age-appropriately voiced: primitive-age field notes ("Four walls and the ambition of someday having five.") drifting toward late-game existential dread ("Stores resources in a superposition of 'we have it' and 'oh no.'"). Every one of the 22 ages also gained a one-liner shown on the age-advance splash beneath the formal title, scaling from caveman observation to cosmic dread. The pass now reaches the rest of the game text: **all 61 random / epoch events** have rewritten player-facing log lines ("A wandering merchant arrives, smelling of cabbage and opportunity." / "Your most useless citizen is now a thought leader.") with the mechanical effect summary preserved — the joke rides alongside the numbers, never replaces them. **Every milestone and chain** (77 milestones, 6 chains) gained a *Flavor* quip shown in the completion log beside the reward ("You built a Wonder. Your neighbors are impressed. One is drafting a strongly worded letter."). And a **rotating pool of log flavor** fires on notable moments — a building finishing, an age turning over, a research breakthrough, the onset and end of famine, surviving a catastrophe — drawn at random per event so the same moment reads a little differently each time, and rate-limited on high-volume moments so the tick log never turns to wallpaper. Expedition results, faction standoffs, encounters turned away by a full court, and war raids go further: they draw from a **procedural flavor generator** instead of fixed three-line banks — about 1,800 hand-written sentences, split by era so a Stone Age raid and a Quantum Age raid read like different places (from the Information Age on, at least half of what you read is written for your era), fed through one stream per log that never repeats a sentence inside a 32-line window. A war raid's log line no longer reprints the raiding civilization's whole backstory every time.
- **Full Diplomacy System — Civilization Encounters** — the six NPC factions become the backbone of an **11-civilization roster** spanning every epoch. Each civ has a **personality** (peaceful / aggressive / mercantile / isolationist), a **backstory**, and a **strength** rating. Civilizations are met through **first-contact events** (age/epoch-gated; the founding civs appear in the early eras, more each epoch). Opinion now **drifts by personality** — aggressive civs trend hostile, peaceful trend friendly, mercantile warm to your trade activity, isolationists hover near neutral. **Peaceful, high-opinion civs lend you workers** (temporary loans, or permanent above opinion 80). **War** is event-driven (no combat): a civ declares war only when its opinion is below -75 *and* a provocation threshold is crossed (raid their trade route via `diplomacy raid`, or embargo them — two provocations trips it), then launches periodic resource **raids** scaled to its strength. Make peace with `diplomacy tribute <civ>` (gold + culture) or by waiting them out. The Diplomacy overlay now shows personality, backstory, war banners, and lent-worker status; new civ state (discovery, opinion, war, provocations, lent batches) persists in saves and resets per prestige run.
- **Embassy buildings & Diplomacy overlay** — two new diplomacy buildings: the **Embassy** (Colonial Age, costs gold + iron) and the **Grand Embassy** (Industrial Age, costs gold + steel, twice the rate). Staffed with trade-domain workers, they passively generate opinion each tick toward all of your non-hostile factions (scaled by worker fill, capped at +100). A new **Diplomacy overlay** — opened by typing `diplomacy` (or `dip`) with no arguments — lists every civilization with opinion bars, color-coded status, the active trade-rate bonus, a threshold indicator (e.g. *+8 to friendly*), and the actions available per civ; `diplomacy <action> <civ>` still performs the action directly.
- **Theme system** — eleven switchable themes, dark and light, with a live picker and a `theme` command. Standard: Forge (dark, default) and Daylight (light). Accessibility, unlocked from the start: Deuteranopia-safe, Protanopia-safe, High Contrast, and High Contrast Light (light, AAA text contrast, blue/orange gains and losses with ▲/▼). Five flavor themes (Parchment, Bronze, Cyberpunk, Monochrome, Cosmic) unlock by reaching milestones. Each theme describes its full surface — canvas, panels, borders, selection, text tiers, good/warn/bad — and paints its own background, so light themes read correctly on dark terminals and vice versa. The picker groups themes under Standard / Accessibility / Unlockable, tags each row light/dark, accessible, or locked, and shows a details pane with palette swatches and a self-contained sample panel in the candidate theme's colors. The city and world maps are light-aware (the city map re-keys for a light page; space-age maps stay dark). Active theme and unlocks persist per account.
- **Multiple accounts** — each account gets its own data slot (`data/accounts/<id>/`), with a one-time, non-destructive migration of existing data. A start-screen Accounts panel lists every account and lets you switch, create, export, import, back up, recover, and wipe. Name-derived identity with `AGEF-` recovery codes; signed, ID-bound export/import that lands in its own slot and can't clobber another account; lifetime stats and achievements.
- **Account backups** — a full-slot snapshot (account.json + saves) is taken before a wipe, on export, and on demand (`account backup` or the panel), keeping the newest ten; plus a one-time pre-migration snapshot of the old data layout.
- **Save lineage and Load Game browser** — procedurally-named saves you can branch into new lines, shown as a lineage tree; the browser adds delete / rename / duplicate, richer save metadata, and account attribution.
- **Multiplier resolver** — a single engine resolver that every bonus source emits into; the Active Multipliers panel renders from its breakdown, color-coded by sign with every contributing source shown.
- **Morale rework** — morale is now a managed two-way resource with restoring buildings, a continuous curve, a history graph, and banded displays; faith production rate lifts morale.
- **Expeditions and Army, split** — scouting Expeditions and military Campaigns are now separate systems; soldiers are a real produced/stored resource, with concurrent per-category expedition slots.
- **Help panel** — `help` opens a panel instead of dumping inline text.
- **Main-screen UI overhaul** — a framed command bar, a scannable ✓/✗ age-progress strip, a cleaner log, lifted secondary-text contrast, an early-game onboarding panel, and rebalanced panel widths.
- Nanobots are a real resource now — a Modern-age producer building (Nano Foundry, +80 nanobots/tick), 3 nanobot techs (Nanofabrication cuts build costs −8%, Medical Nanobots boosts population and food, Self-Replication ramps nanobot output), and several digital/cyberpunk buildings now cost nanobots to construct.
- Culture has sinks now — Cultural Monuments (spend culture for a permanent production bonus) and a `festival` command (spend a lump of culture for a temporary multi-resource buff); prestige gates remain the primary long-term sink.
- **Trade System Expansion** — six new trade routes fill the colonial → industrial gap (`mercantile_convoy`, `triangular_trade`, `tea_clippers`, `coal_barges`, `cotton_exchange`, `steamship_line`), bringing the total to **21**. A new **harbour lineage** (`harbor` → `harbor_authority` → `seaport` → `container_terminal` → `logistics_hub`, Colonial → Digital, trade-domain workers) boosts the income of **every active trade route** (+5% to +25% per tier, stacking). **Trade disruption** ties into diplomacy: routes whose imports include a resource specialised in by a civilization you're **at war with or have embargoed** earn nothing until the conflict ends — flagged in the log and the Trade overlay, and resuming automatically on peace. A new **black market** (`blackmarket` / `trade black <resource>`, Colonial Age+) lets you spend a lump of **culture** on a high-risk deal: a 55% chance of a large resource haul (2.5× the stake) or losing the culture, on an ~8-minute cooldown. The Trade chain capstone **Maritime Empire** extends the chain to **6** milestones.
- **Age Awakenings** — a one-time deterministic boost fires the first time you enter each epoch's signature age (7 epochs, 7 awakenings): Pottery Mastery, Discovery of Metallurgy, Steam Breakthrough, Electrification, Information Age Dawns, Cybernetic Awakening, and First Contact Signal. Each grants a modest temporary production bonus, fires at most once per prestige run, persists across save/load, and resets on prestige so the next run can earn them again.
- **Ancient Civilization Memory** — early in a new prestige run (Primitive or Stone age), a ~40% chance offers a cache of your now-extinct predecessor: accept to research one random age-appropriate tech free of prerequisites, age gate, and knowledge cost — but at half research speed (2× ticks). One cache per run; declining or reloading won't re-roll it, and the reachable tech tier scales with prestige level (one extra age of reach per two levels). Requires prestige level ≥ 1 and resets each run.

### Balance
- **Storage holds about an hour and a half of production (the Storage Covenant).** Every age's full storage must hold 1.5 hours of what a moderate economy of that age makes of each resource it builds with; the ages that fell short got more storage per copy: Storage Pit 600 → 2,200, Warehouse 4K → 11K, Granary 15K → 26K, Classical Vault 100K → 110K, Keep 400K → 410K, Renaissance Vault 500K → 2.7M, Colonial Warehouse 10M → 33M, Industrial Depot 50M → 170M, Victorian Vault 350M → 1.1B, Info Vault 260B → 790B. Stores in the Renaissance to Victorian Ages used to fill in 5–15 minutes. With the build plan and overflow, a player who checks in every hour reaches the first prestige in about 2.8 days (was 7.3), every 3 hours in about 3.9 (was 17.4), every 8 hours in about 7.2 (was 44.5).
- **The trading post no longer costs gold.** It costs 32K stone and 15K iron (was also 8.8K gold) and makes 48.8 gold/tick fully staffed (was 65.2; its rate follows its price). It is the Iron Age's only gold producer and its only trade building, so a player who skipped the Bronze Age market had no way to get gold in the Iron Age at all. See Fixed.
- **Harbinger Appease is priced off what faith buildings make.** It cost 15% of the passage storage, and with the shorter ages faith almost never reached it: the smoke-test bot could not afford it in most epochs. Level 1 now costs a quarter of what a player who keeps five staffed copies of each faith building makes over the epoch's ages at their target lengths (with the production bonus of the techs and wonders held by then), in faith, and the same for culture from the Steel Era on; level 2 is double, and the price is still the same in every age of the epoch. Level 1 prices: Stone Era 59 faith (was 12,000), Iron 5,400 (33,000), Steel 74K faith and 770K culture (2.25M each), Electric 1.2M and 16M (70.5M each), Digital 12M and 180M (147B each), Neon 130M and 2B (46.5B each), Cosmic 1.2B and 19B (46.5B each). The bot, building toward it, now affords level 1 a quarter to half of the way through each thread before the Digital Era (earlier after, when saved faith carries over) and level 2 before the passage in every epoch. See the Harbinger wiki page.
- **Space Age back on pace (2.1x its target, now 1.3x).** The Orbital Refinery makes titanium (it made dark matter, which only unlocks in the Interstellar Age) and no longer costs titanium itself (its 88T went to plasma), so the Space Age has a titanium producer that can start the supply. Most of the gain came from the smoke-test bot, which traded away the plasma it was saving for the next producer every time titanium, which nothing made, was its slowest target.
- **Wonders fit in storage.** The Sistine Chapel costs 24M stone and 24M gold (was 27M and 19M; 27M stone was more than any Renaissance warehouse holds) and the World Simulation 34T steel, 20T electricity and 1T data (was 54T, 6.4T and 970B; 54T steel was more than Digital Age storage holds). Wonders are banked a deposit at a time, but each part now fits in one full store.
- The "Iron Forged" milestone grants 40 iron instead of 40 coal, which is locked until the Renaissance.
- **Pacing rebalance: about 3 days to the first prestige.** Each age now has a target length at 1x (Primitive 15 minutes, Stone 45 minutes, Bronze 1.5 hours, then 2.5, 3.5, 4.5, 6, 7, 8, 9, 10 and 12 hours to the Atomic Age; 12 to 24 hours for each age after), and the economy is derived from it. Before, production grew 2x per age while prices grew 5-8x, so from the Iron Age on a new building took days to pay for itself, and no run reached the Modern Age. The smoke-test bot now reaches it in about 2.4 days on every seed (it used to stall in the Colonial Age after 160 days). See the Buildings, Trade and Ages wiki pages for every number.
  - *Production:* every producer's output of a construction resource is set so that, fully staffed, it earns back its first copy's price in its age's payback time: 1/16 of the age's target in the Primitive Age, rising to about 1/4 in the Renaissance and 2/3 in the Space Age, since later ages also run on every older building. Most rates went up by orders of magnitude (the Renaissance mill makes 676 steel a tick, not 1.6). Food, faith, culture and soldiers keep hand-set rates.
  - *Market:* any two construction resources of your current age now trade at their price parity less a 20% fee, including pairs that were never listed (steel for titanium, data for crypto). This is where stone comes from after the Bronze Age, iron after the Medieval Age, steel from the Modern Age on, and titanium and crypto: nothing in those ages produces them. Listed pairs involving food, faith, culture, or knowledge before it is a building material keep their fixed rates.
  - *Wonders* each cost the same share of their age's economy (40 price units, keeping each wonder's resource mix): the early wonders cost far less, the Particle Accelerator less, the Warp Nexus and Cosmic Beacon more. Sacred Grove 500 food (was 5K), Great Monolith 1.5K food (was 10K), Sistine Chapel 20K faith (was 6M), and the Stellar Cradle no longer costs uranium, which nothing in the Fusion Age produces.
  - *Build and research times* are capped at 1/6 of the age's target (wonders included; the Steam Turbine took 28 hours and the Server Farm a week) and 1/48 for storage, which you buy many times an age. Research is capped at 1/8 of its age's target.
  - *Early game:* gathering camp 1.0 food/tick (was 0.5), forager post 1.5 (was 1.0); story circle 0.2, elders' hall 0.6, scriptorium 2.0, agora 1.6, library 3.2 knowledge/tick (were 0.05 to 0.8). Stone Age needs 1K food, 1K wood, 150 knowledge and 10 huts (was 16K, 10K, 2.8K, 20); Bronze Age 4K food, 8K wood, 4K stone, 1.5K knowledge and 15 longhouses (was 30K, 16K, 8K, 10K, 40).
  - *Requirements sized to what you can make:* the Renaissance asks 880 steel and 8K faith (was 3.5K and 44K: no Medieval building makes steel, and faith comes only from flat faith-building rates); Steel Forging makes 0.25 steel/tick (was 0.1).
  - *Gates evened out:* Classical needs 12 agoras and 10 trading posts (were 8, 5), Colonial 8 exchanges, universities and art studios (5, 3, 5), Industrial 8 plantations and 10 ports (5, 8), Interstellar 15 orbital habitats (20).
  - *Harbinger:* with shorter ages Appease's old faith price was rarely affordable before the passage (repriced below); Brace is affordable within hours.
- Flattened building cost curves and raised age-advancement requirements.
- Capped age-transition carryover to a starter head-start.
- `gather` yield raised 10 → 25 and disabled past the Medieval age.
- Storage lineage: every tier stays affordable to its build cap — the cost of copy N never outruns the storage those copies provide (the old stash-deadlock class of bug, lineage-wide). Storage buildings cap at 25 copies (stash 50), and several under-provisioned vault caps were raised. Guarded by a regression test.
- **Every age can be completed again.** The Stone Age could not be finished (the 50th longhouse cost more wood than Stone Age storage holds), and neither could any age from the Iron Age on, for four reasons, now fixed and guarded by a test that runs on every build (the "Gate Covenant", see the Ages wiki page):
  - *Requirement counts:* the last required copy of each required building now costs at most half the storage you can build by then. Bronze needs 40 longhouses (was 50), Medieval 15 libraries (was 20), Electric 10 steam turbines (was 20), Atomic 15 electric arc furnaces (was 20), Modern 15 nuclear reactors and 15 bunker complexes (were 30 each), and every age from Information to Transcendent had counts of 50 to 500 cut to 10 to 30.
  - *Older buildings:* requirements no longer name a building from an earlier age, which you can't build any more and may have upgraded away. Classical asks for hunting lodges and trading posts instead of barracks and markets, Medieval for military academies instead of barracks, Renaissance for guildhalls instead of markets, Electric for Bessemer plants instead of steel mills, Atomic for power stations instead of steam works, Galactic for antimatter forges instead of orbital refineries, and Quantum for stellar metallurgy instead of antimatter forges. Industrial no longer asks for market gardens, and Victorian no longer asks for oil, which nothing could produce before it.
  - *Resources with no way in:* the smithy no longer costs iron (it is the Bronze Age's only iron source), the mill no longer costs steel, the ironworks, smelter and forge no longer cost coal (which unlocks in the Renaissance), the Nuclear Extraction Plant no longer costs uranium, the Crypto Exchange, Cyber Shrine and Logistics Hub no longer cost crypto, the Monument of Ages no longer costs titanium, and gold and data can be exchanged from the Modern Age (was Information).
  - *Cosmic storage:* the Stellar, Galactic and Quantum Vaults hold +2Q, +20Q and +200Q per copy (were +500T, +2Q, +10Q).
- **Storage no longer upgrades.** Reaching the Stone Age used to offer `upgrade stash`, which traded a stash you can never rebuild for one of the 25 storage pit slots and lowered the most you could ever store. Upgrading more than two of 32 stashes made the Bronze Age's food requirement impossible. Storage buildings are no longer offered as upgrades, any upgrade stops at the target's build cap, and trying to build an older age's storage says so instead of pointing at `upgrade`.
- **Upgraded markets still trade.** The market exchange now opens with any trade building (market, trading post, merchant quarter, ...), not just a market or port, so upgrading your markets no longer shuts it until the Colonial Age.
- Milestone chains rebalanced: completion speed-boosts normalized across all six chains (Military was 18× weaker than Settlement — now in line); the Trade chain expanded 3 → 6 milestones (5 via the milestone revamp, then a sixth — Maritime Empire — with the Trade System Expansion); and Military/Trade/Scholar capstones now pay out broad `production_all` instead of domain-only bonuses, so finishing any chain helps your whole economy.

### Changed
- **The command prompt completes as you type.** The best completion of the line shows in dim text after the cursor (`adv` shows `advance`); `Tab` or `→` takes it, and `Tab` again cycles through the others. The suggestion dropdown is gone, and `↑`/`↓` stay command history. `Enter` runs a whole command as typed; an unfinished line runs its completion (`adv` + `Enter` advances), except for commands that can't be undone (`sell`, `dismiss`, `research cancel`, `plan clear`, `load <name>`, `prestige confirm yes`, `festival confirm yes`, `harbinger invite`, `diplomacy raid`, `quit`, `account switch`/`import`/`recover`/`wipe`), which are filled in for a second `Enter`. See [The prompt](site/docs/commands.md#the-prompt).
- **Completions follow the game**: `build` offers only what you can build in this age, affordable first; `research` the techs you can start, affordable first; `plan build` and `plan research` the same, then the next age's; `assign` your built buildings, free slots first; `trade` what the market buys and sells; `theme` your unlocked themes.
- **One command list.** Completion, the Help panel and the smoke suite's docs and fuzz checks now all read one command registry (`ui/commands.go`), and a test holds it to the command handler both ways, so a command can no longer be missing from completion or help.
- **Map: complete rewrite (`ui/citymap`)** — the 2,734-line MapV4 monolith and its **54MB** of embedded realistic-photo backgrounds + dead sprite PNGs are gone, replaced by a light, fully **theme-aware** city map that retints live with the active theme. It is built in three layers: **(P1)** procedural, theme-tinted terrain (a self-contained FBM elevation field — water → lowland → highland → peak — with a faint per-age hue shift) streamed through half-blocks; **(P2)** per-age layout strategies (organic scatter → hub-and-spoke → castle + quarters → zoned grid → city blocks → campus clusters → orbital rings), theme-tinted roads, and your **actual buildings** drawn as **2.5D volumes** (lit roof + shaded wall + drop shadow) — **one named marker per built building type** (a Shrine, a Gathering Camp, a Story Circle — your real buildings, not aggregated resource districts), clustered with their lineage-mates and colored by lineage so each domain reads as a same-colored neighborhood whose silhouette evolves by age; **(P3)** the systems-weave + polish — your **active trade routes** drawn as dashed lanes fanning out to the map border, the discovered **diplomacy civilizations** ringed around the edges as relationship-colored markers (allies green, rivals/embargo red, at-war bright red with a "!"), each building **labeled with its own name** (collision-limited so the map stays legible, with each lineage cluster's most prominent building named first), a corner **title** ("Empire — current age"), and light per-era flourishes (orbital starfields, cyber neon edge-glow, industrial smoke). Labels and civ markers are crisp theme-colored **text** stamped on top of the soft half-block terrain (a hybrid render model), and every color — terrain, structure, lanes, and overlay text — resolves from the active theme at draw time, so a theme switch retints the whole map. The full P1–P3 citymap rewrite; binary −54MB.
- **Map: biome terrain + terrain-aware meandering roads (`ui/citymap`, P4)** — the flat elevation-band terrain is now a real **biome map**. A second independent FBM **moisture** field joins the elevation field, and the two classify every pixel into one of eight biomes — **deep water, shallow water, beach/sand, grassland, forest, hills/rock, mountain, snow/peak** — each colored from the **active theme** (water from the dim role toward blue, forest from the background toward green, grassland a lighter land tone, hills/mountain toward a dim gray, snow toward the bright text role, sand a warm light tone), so the whole biome map still **retints on a theme switch** and keeps its faint per-age hue shift. Each biome carries a **passability** flag (deep/shallow water + mountain/snow are impassable, all land is passable), and the passability grid is built once per render and reused. **Roads are no longer straight radii** — instead of Bresenham lines from each building to the centre that sliced through lakes and peaks, every road is now routed with **A\* pathfinding** over a downsampled (3 px/node) cost grid built from the biomes: water/mountain is blocked, open land is cheap, a clearance penalty keeps routes off shorelines, and a per-node noise jitter plus a smoothing pass make even open-terrain roads gently **wander** — so roads bend **around** water and mountains and read as streets, not spokes (boxed-in buildings fall back to a direct line). **Building placement is terrain-aware too**: any building (or the centre marker) whose slot lands on water or a mountain is **nudged to the nearest passable cell** before its volume, label, and road are drawn — nothing sits in a lake or on a peak, while the per-era layout and lineage clustering are otherwise unchanged. Pathfinding is computed only when the cached image regenerates (age/size/buildings/theme change), so it stays cheap. The central marker's label is renamed **"Capital" → "City Center"** to match the city view (same prominent marker).
- **Catastrophes: no more Defer.** When a catastrophe hits, nothing is destroyed until you choose Endure or Succumb, and the game keeps running meanwhile. Esc closes the choice so you can look around; a status-bar badge shows it is pending, and a bare `catastrophe` command reopens it. A save with a pending catastrophe shows the choice again on load.
- **Catastrophes: a pending catastrophe blocks `advance` and prestige** until you choose, with a message pointing at `catastrophe`.
- **Catastrophes: none before the Iron Era.** Epoch transitions into earlier epochs never roll one, so a run sees at most 6 (one per epoch from Iron to Cosmic).
- **Catastrophes: Succumb's Ancient Knowledge stacks and is permanent** — +25% research speed per distinct epoch succumbed, derived from your legacy flags, so it survives save/load, Succumb and prestige. It shows as "Legacy" in Active Multipliers. Old saves have the stored copy converted on load.
- **Catastrophes: ruins are capped at 24.** Past the cap the lowest-value ruins (earliest age, then lowest output) crumble first; old saves are trimmed on load.
- **Catastrophes: seeded randomness.** Epoch event rolls, destroyed buildings and ruins now come from the run's seeded generator over a fixed-order pool, so the same seed gives the same outcome. The Ancient Memory and black-market rolls also fall back to the seeded generator.
- **Catastrophe odds are visible**: the bare `catastrophe` command and the Epoch tab show the chance for the next transition (12–18% by faith).

### Fixed
- **`wonder bank food all` did nothing.** The command only knew `wonder collect`, so `wonder bank …` quietly showed the bank status and deposited nothing. `bank` now works as `collect`; the amount can be a number, `all` or `max` (as much as the wonder still needs, up to what you have), or left off (the same as `all`); and `wonder bank all` banks every resource the wonder needs. Each deposit reports what actually went in (it used to echo the amount typed even when the bank took less), and a refusal says why: the wonder is built, it doesn't need that resource, that part is full, you have none, or you have less than the amount. Autocomplete offers the resources the wonder still needs and the amount words.
- **Same seed, same run on every machine.** A run played on Apple Silicon (or any arm64) drifted from the same run on an x86 PC within a few hundred ticks, so a save could play on differently after moving between them and local smoke numbers didn't match CI. The arm64 compiler fused multiply-adds that x86 rounds in two steps, and `math.Pow`, `Exp` and `Log` return different last bits per architecture. Simulation code now rounds every such product and uses its own portable `Pow`, `Log10` and friends (package `detmath`); x86 runs play out exactly as before. A new CI workflow plays the same seeds on Linux x86, Linux arm64 and macOS arm64 and fails if they differ.
- **Completion drift.** The completion list was kept by hand apart from the command handler and kept falling behind it (it lacked `diplomacy tribute`, `raid` and `acct`, and offered `research list list`). It is now the command registry, held to the handler by a test.
- **Help** said `gather` takes at most 10; it takes 25. `quit` is now listed in Help and the Commands page.
- **Iron Age gold trap.** Reaching the Iron Age without a Bronze Age market left no way to get gold: the trading post, the only gold producer and the only trade building (the market needs one), cost gold, and so did agoras, legion forts, temples and the Classical Age's requirement. The trading post is now priced in stone and iron only. The automated age-gate check (the "Gate Covenant") now asks whether every resource a gate needs can be had by a player who skipped every building no earlier gate required, so a trap of this kind fails the build; it found no others.
- **Commands needed Enter twice.** While the suggestion list was open, the first Enter only took the suggestion (`advance` became `advance `), so `advance`, `status`, `research list` and the rest ran on the second press. Enter now runs what you typed on the first press; Tab still takes the highlighted suggestion.
- **`build` queues several storage copies.** Building a storage building a second time while the first copy was under construction was refused ("already under construction"), though `build <storage> 5` queued five. Now each `build` queues another copy, up to the building's cap. Only unique buildings still refuse a second copy.
- **Keys were case-sensitive in some commands.** `build Hut` worked but `sell Hut` and `dismiss Hut` said the building was unknown, and so did mixed-case keys in `research`, `expedition`, `campaign`, `diplomacy`, `trade route start|stop` and `prestige buy`. Every command now reads its building, tech, resource, civilization, theme, expedition, route and upgrade keys in any case.
- **The Commands page's shortcut table was wrong.** It listed `e`, `m`, `w` and `l`, which are not commands, and gave `r`, `s` and `t` as panels when they are `recruit`, `status` and `trade`. It now lists the real short names, and the docs check holds it to the command registry. How to Play and the Workers page no longer point at those keys either.

### Removed
- **`catastrophe invoke`.** Deliberately facing a catastrophe (to Succumb for a legacy bonus) moves to the Harbinger's Invite. The command only ever worked in the Stone Era anyway, where it made a free invoke-and-Succumb loop. Developers can still force one with the dev console's `/catastrophe`.

### Fixed
- A wonder bank could never fill when its rounded price came out as 999999.9999999999 (the Parthenon's stone): the cost rounding now multiplies by an exact power of ten, and a bank within 0.001 of its price counts as full, as the deposit command already assumed.
- Building descriptions now always show the building's real production rate.
- The `prestige` status said to reach the Medieval Age to prestige; it now says the Modern Age.
- Theme picker no longer deadlocks the app on arrow/Esc.
- Danger dialogs (wipe data, wipe account, delete save) were unreadable on some themes — the account-wipe panel drew red text on its red fill. They now read on every theme.
- Keycaps and several stray colors ignored the active theme.
- Procedural save names: flattened the distribution (no more "Grand Duchy" clustering) and added ~10× more names across every bank.
- Account Wipe modal shows the exact name to type and has a cleaner confirm UI.
- Save-name modal: readable input contrast, opaque background, sized to content.
- Load Game: bare `load` opens the browser; renaming re-parents child saves; footer hotkeys render as keycaps.
- Modifiers: negative production debuffs apply correctly and build-cost modifiers are wired in.
- Stash is buildable again (its first-copy cost had exceeded the base wood cap).
- Tick loop and UI snapshot no longer rebuild the building table (and other config tables) every tick; a late-game tick is ~13x faster.
- Age-advance splash no longer freezes the game when an epoch catastrophe rolls on the same advance; the catastrophe modal now opens after the splash is dismissed.
- Catastrophe and Ancient Memory modal buttons show their labels (they rendered blank) and their keyboard shortcuts.
- Catastrophe Endure counted wonders when working out the 20% of buildings to destroy (while never destroying them); it now counts only destroyable buildings.
- Catastrophe Endure left workers assigned to buildings it destroyed; they now return to the idle pool before the 25% worker loss.
- A random catastrophe roll could overwrite one that was still pending, and the new one's modal never appeared. That can no longer happen.
- The catastrophe toast never fired (no event was published); it now shows when a catastrophe strikes.
- The catastrophe modal covered the whole dashboard with blank space and had see-through rows; it is now a box sized to its content, floating over the visible dashboard.
- The Epoch tab showed a pending or overwritten catastrophe as "Survived"; it now shows Survived, Succumbed, Pending, or "outcome not recorded" for old saves.
- The Stats panel counted succumbs as survived; Survived now means Endured, and Succumbed counts every succumb.
- `recruit 9223372036854775807` overflowed the population check and left the population near -9.2 quintillion. Counts are now whole numbers from 1 to 1,000,000 and amounts must be positive numbers; anything else is refused with the usage line (`trade food wood NaN` and `wonder collect food NaN` used to turn a resource into NaN).
- Loading a game dropped building upgrades offered in an earlier age (a lineage with no tier in the current age lost its offer, and `upgrade gathering_camp` said there was nothing to upgrade). Saves now keep the offers.
- Loading an Iron Age save unlocked coal, which normal play unlocks in the Renaissance Age.
- Loading a game restarted its random events, expeditions and encounters from the run's first roll. A loaded game now carries on the same random stream it was saved with.
- Research bonuses could come back a hair off after a load (0.7999999999999999 instead of 0.8).
- After a prestige, the storage upgrade didn't apply until the first tick, and starting resources beyond the base storage were cut off (Starting Food tier 3 gave 35 food, not 75).
- `dump` wrote its file to a `data/logs` folder relative to wherever the game was launched; it now goes in `logs/` in your account's data folder.
- Upgrading only some copies of a building left all its workers on the copies that remained, more than they have slots for. Workers over the remaining slots now move to the upgraded building, or go idle if it's full.
- Loading a game reset the festival and black-market cooldowns, and a prestige left them counting against the old run's ticks. Both are saved now and reset on prestige.
- Selling a storage building left stock above the new, lower cap for anything not being produced.
- Autocomplete offers `diplomacy tribute`, `diplomacy raid`, `account wipe` and the `acct` alias, and no longer re-offers an argument you've already typed (`research list list`).

---

## [v3.6.4] — 2026-03-25

### Added
- add buildings overlay showing full age history

### Fixed
- reconstruct pending upgrades on load so 'upgrade' works after save/reload
- correct age advancement requirements to reference current-age buildings
- city map shows all-age buildings rendered in current age style
- lock build command to current age only
- economy tab Buildings panel only shows current age buildings

---

## [v3.6.3] — 2026-03-24

### Fixed
- resolve undefined variable compile errors in overlay_stats (#32)

---

## [v3.6.2] — 2026-03-24

### Fixed
- expose permanentBonuses in GameState and fix Active Multipliers display (#27)
- correct energy and metallurgy lineage production rates on upgrade (#28)
- city map only renders buildings from the current age (#29)
- unique 16x16 sprites for all 22 age wonders (#30)
- wonder multiplier bonuses now appear in Active Multipliers (#31)

---

## [v3.6.1] — 2026-03-24

### Fixed
- embed assets/maps and assets/sprites into binary

---

## [v3.6.0] — 2026-03-24

### Added
- map rendering experiments (mapv1 / mapv2 / mapv3) (#26)

---

## [v3.5.0] — 2026-03-23

### Added
- player-driven building upgrade system (#24)

### Fixed
- prevent electricity regression when upgrading energy lineage tiers 4 and 9 (#25)
- resolve CSP violation on changelog page

---

## [v3.4.0] — 2026-03-23

### Fixed
- changelog page — fetch CHANGELOG.md instead of GitHub Releases API (#23)
- add morale panel to workers overlay (#22)

---

## [v3.3.0] — 2026-03-18

### Added
- morale system — worker output multiplier (#21)
- declutter main screen — workers/wonder panels to overlays (#20)

---

## [v3.2.5] — 2026-03-17

### Added
- wonder completion required to advance ages (#19)

### Fixed
- remove full-screen dev console, route dev commands through main input
- age splash keypress regression and faith preserved on age advance

---

## [v3.2.4] — 2026-03-17

### Added
- dynamic overlay width + fix all providers to accept terminal size
- civilization history overlay with braille line graphs

---

## [v3.2.3] — 2026-03-17

### Fixed
- building cost scaling now correct for batch, max, and queued builds

---

## [v3.2.2] — 2026-03-16

### Fixed
- three UI display fixes — wonder rates, building names, speed multiplier

---

## [v3.2.1] — 2026-03-16

### Fixed
- four config bug fixes — event collision, tech rates, cost cliff, titanium age

---

## [v3.2.0] — 2026-03-16

### Fixed
- age advancement splash ignores keypresses until auto-dismiss

---

## [v3.1.0] — 2026-03-11

### Fixed
- wiki update and fix exponential worker growth bug

---

## [v3.0.0] — 2026-03-09

### Added
- expand 33 → 74 milestones and harden all thresholds
- fix map building scale, density throttle, and 249-building sprite coverage
- revamp map system with age-appropriate city layouts and building sprites
- add research as a panel view and command option
- expressive resource bars with fill-level colour and status glyphs
- implement worker_loss effect + ended message damage summary
- worker panel grouping, map fix, stats rate breakdown
- replace F-tab system with command-summoned overlay panels
- purple fill color for all progress bars
- rewrite worker panel with per-domain assignment groups
- add floating overlay framework + milestones panel
- collapse 12 domain pools into single generic worker pool
- simplify assign/unassign/recruit command API
- rename Villager→Worker across UI layer
- rename Villager→Worker across game core and engine
- strip legacy alias system from WorkerManager
- comprehensive autocomplete revamp
- population screen redesign
- add command history ring buffer
- rewrite autocomplete to use domain keys
- Phase 13 — epoch-exclusive regular events
- Phase 11e+11f — Epoch tab (F10) + Stats tab epoch/legacy fields
- Phase 11c — age advance modal transformation summary + epoch reveal
- Phase 11a+11b — economy tab worker display + villager panel rewrite
- Phase 11d — culture progress bar + faith threshold indicator
- Phase 10 economy redesign — full 13-lineage building content overhaul
- implement catastrophe system (Phase 9 economy redesign)
- implement epoch system (Phase 8 economy redesign)
- Phase 7 economy redesign — age transition building transformation pass
- Phase 6 economy redesign — worker-building coupling engine
- Phase 5 economy redesign — config foundation data structures
- update Savefile system with a new configuration
- add logo to website nav, hero, footer, and og:image

### Fixed
- refactor changes for wiki and add updated screenshots
- remove all F-key references, fix commands at a glance and controls
- radical threshold hardening + fix blank progress bars
- fix milestone progress flickering, missing categories, and autocomplete gaps
- correct military lineage tier order, bad upgrade, and wonder description
- research_speed permanentBonus now actually reduces research duration
- restore building scale=2, remove primitive-era wonder circle artifact
- fix map city circular blob — remove anchor-walk clustering and expand zone
- fix map building shadow artifact and add era-aware city label
- wire master_builder and scholars_haven to correct game state
- worker panel always resolves class name for any domain+age
- Scholar's Haven now requires 5 knowledge workers assigned
- war_machine no longer fires in primitive age
- capture ESC on overlay TextView to close panel
- recruit log shows worker(s) not domain key; fix dev tab focus
- assign/unassign gracefully skip legacy domain prefix arg
- recruit autocomplete shows class names; save-compat; class name resolution
- worker alias, wood camp timing, UI panel layout
- Phase 17a — stone_camp and hunting_lodge age gating
- Phase 14 — worker assignment, gather cap, domain key alignment
- remove dashboard bleed-through on wipe confirmation
- three ticker freeze bugs
- resolve save directory relative to binary, not CWD
- fix docsify security for mobile use

### Balance
- full age pass renaissance→quantum — rates, times, food costs
- city spread ratio sized to 20
- medieval age pass — rates, build times, food cost
- classical age pass — rates, build times, food cost
- iron age pass — rates, build times, food cost
- bronze age pass — rates, build times, food cost
- stone age pass — buildings, rates, costs
- wood camp rate up, food drain nudged higher
- stash cost up, gathering camp cheaper, food drain reduced
- Phase 17b — fix primitive age economy (food rate vs drain)
- game balances, ages and research

### Changed
- delete dead tab_*.go files, migrate helpers to overlays
- simplify recruit command — remove domain arg
- replace DefaultWorkerTypes usage with live worker state
- remove all backward-compat shims
- remove assign/unassign backward-compat shims
- split buildings_new*.go into per-lineage files

### Other
- Change floor value for buildings to NOT produce if no workers in the building

---

## [v2.5.2] — 2026-03-02

### Fixed
- package declaration typo in config/buildings.go (confi → config)

---

## [v2.5.1] — 2026-03-02

### Fixed
- clarify bullet point prompt wording in commit helper
- commit helper adds balance type, length enforcement, and bullet body

### Changed
- stash now has max count of 50, all buildings in primitive take longer to build, altar production raised from .004 to .008 knowledge

---

## [v2.5.0] — 2026-03-02

### Added
- manual age advancement — type 'advance' when ready
- add interactive conventional commit helper (make commit)

---

## [v2.4.7] — 2026-03-02

### Added
- show version + async update badge

---

## [v2.4.6] — 2026-03-01

### Fixed
- rich Discord embeds + fix blank release notes

### Other
- patch release notes

### Fixed
- release workflow: awk now skips [Unreleased] section and targets versioned entries only — release notes no longer blank
- release workflow: removed redundant github-release-to-discord.yml (GITHUB_TOKEN releases don't trigger release: published in other workflows)
- release workflow: cleaned up dead commented-out SethCohen job
- discord notification: rich embed with per-section fields (Added/Fixed/Changed/Other), thumbnail, timestamp, and download link

---

## [v2.4.5] — 2026-03-01

### Other
- new release notes updates

---

## [v2.4.4] — 2026-03-01

### Fixed
- replace Python heredoc with jq+curl for Discord notification

---

## [v2.4.3] — 2026-03-01

### Fixed
- Discord notify step - pass values via env vars not heredoc interpolation

---

## [v2.4.2] — 2026-03-01

### Added
- screenshot lightbox on click + smaller cards showing 2 at a time

### Fixed
- move Discord notification into release.yml as a final step

### Other
- update screen shots mechanism
- new screenshots

---

## [v2.4.1] — 2026-03-01

### Fixed
- explicitly mark GitHub releases as published to trigger Discord webhook
- screenshot caption renders below image, not overlapping it

---

## [v2.4.0] — 2026-03-01

### Added
- add auto-discovering screenshots carousel to site

### Fixed
- show correct precision for small resource rates in economy tab
- build max now correctly queues multiple buildings with a MaxCount
- update screenshots section heading and subtitle

### Other
- add webook url for discord connect
- add screenshot

---

## [v2.3.0] — 2026-03-01

### Added
- auto-generate release notes from commits + changelog page on site

---

## [v2.2.0] — 2026-03-01

---

## [v2.1.0] — 2026-03-01

---

## [v2.0.0] — 2026-03-01

---

## [v1.1.0] — 2026-02-27

---

## [v1.0.0] — 2026-02-26

### Added
- 22 playable ages from Primitive to Transcendent
- 80 buildings (58 standard + 22 wonders) with scaling costs
- 52 researches with prerequisites and permanent bonuses
- 33 milestones across 5 chains with titles and speed boosts
- 21 resource types with rates, storage caps, and breakdowns
- 8 villager types with food drain and assignment system
- 15 military expeditions with risk/reward mechanics
- 28 random events with sentiment streaks and timed effects
- 15 trade routes with supply/demand pressure
- 6 diplomacy factions with opinion and status tracking
- Prestige system with 9 upgrades across 5 tiers
- In-game wiki server (port 7891) with 10 reference pages
- 9 UI tabs: Economy, Research, Military, Trade, Stats, Wiki, Map, Wonders, Logs
- Save/load system with multiple named slots
- Full keyboard navigation and command parser

---

## [1.0.1] — 2025-02-19

### Fixed
- Minor balance adjustments to early-game resource rates
- Hut cost and pop cap rebalanced for smoother ramp

---

## [1.0.0] — 2025-02-18

### Added
- Initial public release
- Core idle loop with tick-based resource production
- Primitive and Stone Age content
- Basic building queue and villager assignment
