# AgeForge Beta — From Terminal to Steam

## Vision Statement

AgeForge Beta is the evolution of AgeForge Classic into a proper Steam-released game. It preserves everything that makes AgeForge distinctive — the command-line interface, the terminal aesthetic, the depth of a 22-age civilization simulation — while unlocking the full potential of a real game engine: animated city maps, real audio, Steam achievements, and a cinematic experience that terminal emulators can never provide.

The core promise: **it still feels like you're running a civilization from a command line.** That's not a limitation we're overcoming. That's the brand.

---

## Why This Matters

### The Steam Opportunity

AgeForge Classic runs in a terminal. It's incredible for the audience that finds it, but it is invisible to Steam's 130 million active users. A Steam release at even a modest $4.99 price point, with proper capsule art and a 60-second trailer showing a glowing amber terminal controlling a growing city, reaches an audience orders of magnitude larger.

The "command-line civilization builder" niche on Steam is essentially empty. Dwarf Fortress has ASCII. Caves of Qud has ASCII. Nobody has a sleek, modern terminal aesthetic with a real game behind it. AgeForge Beta owns that niche.

### Richer Mechanics

The current Go/tview architecture has served well through 22 ages, 284 buildings, and a full prestige system — but it has hard ceilings. There is no way to play real audio, animate sprites, run shaders, or integrate Steam achievements from a terminal application. The moment you want a forge to glow when workers are assigned, or a catastrophe to shake the screen, or a wonder reveal to fill the display with a glowing ASCII monument — you need a game engine.

Godot 4 removes those ceilings without imposing royalties, account requirements, or bloat.

### The Terminal Aesthetic Is the Differentiator

AgeForge does not need to become another city builder with isometric sprites and a resource bar at the top. That market is saturated. What it needs to become is the game that makes players feel like a systems administrator who accidentally built an empire — and that feeling lives entirely in the terminal interface. Godot lets us enhance that feeling, not abandon it.

---

## The Core Promise (Expanded)

- You type commands. Always. `build lumber_mill`. `research pottery`. `assign forge 5`. The keyboard is the controller.
- The terminal panel occupies 60% of the screen and is the primary interface. It renders with a CRT shader, phosphor glow, and a blinking cursor.
- The other 40% is a live city map — a real Godot 2D scene — that reacts in real time to every command you type. Buildings appear. Roads extend. The skyline changes between ages.
- Overlay panels (research tree, army roster, trade routes) open as floating terminal windows with Unicode box-drawing borders. They look like tmux panes.
- Age advancement plays a brief full-screen cinematic: scrolling text, ASCII art age name with glow effect, ambient audio crossfade.
- Every mechanic from Classic is present. Nothing is dumbed down.

---

## Documents in This Folder

| File | Contents |
|------|----------|
| `engine-selection.md` | Full comparison of Godot, Unity, Unreal, Ebiten, and hybrid approaches. Explains why Godot 4 + C# is the right choice. |
| `architecture.md` | Complete technical architecture for the Godot 4 C# version. Project structure, core systems, UI architecture, tick system, save system. |
| `terminal-aesthetic.md` | Deep design doc on preserving and elevating the terminal look in Godot. CRT shader code, font choices, color themes, sound design. |
| `mechanics-migration.md` | Detailed mapping of every current mechanic to its Godot equivalent. What ports 1:1, what changes, what becomes possible for the first time. |
| `new-mechanics.md` | Vision for new systems unlocked by the engine change. Visual storytelling, Steam features, new commands, city districts, monetization. |
| `migration-phases.md` | Realistic phased plan with weekly milestones from repo setup to Steam launch. ~24 weeks total. |
| `godot-csharp-primer.md` | Practical translation guide for a Go developer moving to Godot C#. Every Go pattern mapped to its C# equivalent. |

---

## The Key Decision: Godot 4 with C#

This decision warrants its own brief treatment here before the full breakdown in `engine-selection.md`.

**Why not GDScript?**
GDScript is excellent for small games and Godot-native developers. For AgeForge Beta, which involves porting ~8,000 lines of Go game logic, GDScript's dynamic typing is a liability. The Go codebase is statically typed, struct-heavy, and interface-driven. C# maps onto that thinking almost directly. Every Go `struct` becomes a C# `class`. Every Go `interface` becomes a C# `interface`. Every `map[string]float64` becomes a `Dictionary<string, double>`. The translation is mechanical. GDScript would require rethinking idioms that don't need rethinking.

**Why not Unity?**
Unity is technically capable but carries three problems: the 2023 runtime fee controversy damaged developer trust and the terms can still change, it's significantly heavier than needed for a 2D game with terminal aesthetics, and it requires a Unity account and license management. Godot is genuinely free — MIT license, no royalties, no accounts, no usage tiers. For an indie game at $4.99, every dollar of margin matters.

**Why not Unreal?**
Unreal Engine is built for 3D photorealism. Using it for a terminal-aesthetic 2D game is like using a Formula 1 car to drive to the grocery store. C++ is a significant step up in complexity from Go. The 5% royalty kicks in above $1M revenue. The binary size alone (~1GB runtime) is disqualifying for a game whose appeal is that it feels minimal and precise.

**Why not keep Go + add a renderer?**
The tempting option. Ebiten (a 2D Go game library) could theoretically render the terminal panel and a city map. But it has no scene editor, no Steam SDK integration, no audio engine worth using, no shader pipeline, no UI framework. Every system that Godot provides for free (signal bus, scene instancing, animation player, audio streams, shader language) would have to be built from scratch in Go. The Go codebase would expand by 50,000 lines and still not match Godot's capabilities. The game logic in Go was never the hard part — it's pure data transformation with no external dependencies. Porting it to C# is three weeks of mechanical work. Rebuilding Godot in Go is three years.

**Why not a Hybrid (Go backend + Godot frontend)?**
Architecturally elegant, but practically painful. Two codebases, IPC protocol design, packaging a Go binary inside a Godot export, debugging across a process boundary. Steam does not understand "my game is actually two processes." The Go game logic has no state that Godot needs to observe other than `GameState` — and that port to C# is straightforward. There is no reason to maintain the complexity of cross-process communication when the data transformation logic is the least interesting part of the system.

**Godot 4 + C# is right because:**
1. No royalties, no fees, no accounts
2. C# is Go-adjacent for a statically-typed, struct-heavy codebase
3. GodotSteam plugin is mature and actively maintained
4. Godot 4's rendering pipeline supports the CRT shaders needed for the aesthetic
5. Ships standalone executables for Windows, Mac, Linux — no runtime required
6. ~50MB binary base (vs Unity's ~200MB+)
7. The Godot community is large, growing, and indie-friendly

---

## What AgeForge Classic Becomes

AgeForge Classic (this repo) does not die when Beta ships. It continues as the "purist" version — the one that runs in an actual terminal, requires no GPU, works over SSH. Classic stays **permanently free at ageforge.io** — that's its home. No Steam page needed for Classic; the website is the distribution channel. Classic feeds design ideas into Beta. The two coexist.

Beta is not a replacement. It's an amplification.
