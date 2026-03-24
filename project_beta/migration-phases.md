# Migration Phases — AgeForge Classic to Beta

## Overview

The migration is estimated at 24 weeks for a single developer working part-time (~20 hours/week) or 12-14 weeks at full-time pace. The phases are designed so that each one ends with a playable, testable milestone — no phase ends with the game broken.

AgeForge Classic continues active development in parallel throughout all phases. Bug fixes and balance patches land on Classic first; they feed into Beta's implementation as informed design decisions.

**Total estimated hours:** ~480 hours (20 hrs/week × 24 weeks)

---

## Phase 0: Repo + Tooling Setup
**Duration:** Week 1 (20 hours)
**Goal:** Working development environment with CI/CD and project scaffold

### Tasks

**Day 1-2: Environment setup**
- Install Godot 4.x (latest stable, as of this writing: 4.4) from https://godotengine.org/
- Install .NET 8 SDK (`dotnet --version` confirms)
- Install JetBrains Rider or configure VS Code with C# Dev Kit + Godot extension
- Verify that Godot 4 C# template compiles: `File → New Project → "C# .NET" template`
- Create new **private** GitHub repo: `github.com/espresso20/project-ageforge`
- Set up `.gitignore` for Godot: https://github.com/github/gitignore/blob/main/Godot.gitignore

**Day 3: GodotSteam integration**
- Download Steamworks SDK from https://partner.steamgames.com/ (requires Steam developer account — create if needed)
- Download GodotSteam prebuilt GDExtension for Godot 4 + .NET from https://github.com/GodotSteam/GodotSteam
- Follow integration guide: place in `addons/godotsteam/`, verify `steam_appid.txt` is present
- Verify Steam initializes without error (use test AppID 480 = Spacewar for early dev)
- **NOTE:** Do not ship with AppID 480. You need a real AppID. Register the game on Steamworks (free, requires $100 Steam Direct fee — do this now so the AppID is available for CI builds)

**Day 4: CI/CD**
- GitHub Actions workflow: build Godot project for Windows, Mac, Linux
- Export templates: download Godot's export templates for all three platforms
- CI fires on every push to `main` branch
- Artifacts: three zip files (one per platform), uploaded to GitHub Releases on version tags

```yaml
# .github/workflows/build.yml
name: Build
on: [push, pull_request]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup .NET
        uses: actions/setup-dotnet@v4
        with: { dotnet-version: '8.0.x' }
      - name: Download Godot
        run: |
          wget -q https://github.com/godotengine/godot/releases/download/4.4-stable/Godot_v4.4-stable_linux.x86_64.zip
          unzip -q Godot_v4.4-stable_linux.x86_64.zip
      - name: Export Windows
        run: ./Godot_v4.4-stable_linux.x86_64 --headless --export-release "Windows Desktop" ./build/project-ageforge.exe
      - name: Export Linux
        run: ./Godot_v4.4-stable_linux.x86_64 --headless --export-release "Linux/X11" ./build/project-ageforge.x86_64
      - uses: actions/upload-artifact@v4
        with: { name: builds, path: ./build/ }
```

**Day 5: Project scaffold**
- Create all directories from `architecture.md` project structure (empty files with stubs)
- Copy JetBrains Mono font files into `assets/fonts/`
- Create `project.godot` with C# enabled, monospace font as default
- Create `Main.tscn`, `BootSplash.tscn`, `GameScreen.tscn` as empty scenes
- Verify project opens in Godot editor without errors
- First commit: "feat: project scaffold"

### Milestone
Godot project opens in editor. CI builds green. All three platform exports produce binaries (blank window). GodotSteam initializes without crash.

---

## Phase 1: Core Engine Port to C#
**Duration:** Weeks 2-4 (60 hours)
**Goal:** All game logic in C#, unit-tested, no UI

### Week 2: Data Structures and Config

**Config export (from Classic):**
- Write a Go script (one-off, not part of Classic codebase) that serializes all config maps to JSON:
  - `config/ages.go` → `data/ages.json`
  - `config/buildings_lineage_*.go` → `data/buildings/lineage_*.json` (13 files)
  - `config/buildings.go` (storage/wonders) → `data/buildings/storage_and_wonders.json`
  - `config/techs.go` → `data/techs.json`
  - `config/milestones.go` → `data/milestones.json`
  - `config/events.go` → `data/events.json`
  - `config/expeditions.go` → `data/expeditions.json`
  - `config/trade_routes.go` → `data/trade_routes.json`
  - `config/diplomacy.go` → `data/diplomacy.json`
  - `config/epochs.go` → `data/epochs.json`
  - `config/workers.go` → `data/worker_domains.json`
- Place all JSON files in `data/`
- Write `ConfigLoader.cs` (see `architecture.md`)
- Write C# classes for each def type: `BuildingDef.cs`, `AgeDef.cs`, etc.
- Unit test: `ConfigLoader.LoadAll()` loads all 284 buildings without error

**GameState.cs:**
- Port all structs from Classic's state files
- Verify JSON round-trip: serialize → deserialize → all fields match
- Unit test: `GameState.NewGame()` produces expected initial state

### Week 3: Core Subsystems

Port these systems in order (each depends on previous):

1. `EventBus.cs` — signals (no test needed, purely structural)
2. `Resources.cs` — `WorkerScaledProduction()` formula
3. `Buildings.cs` — `Build()`, `Destroy()`, construction tick logic
4. `Workers.cs` — `Assign()`, `AssignAll()`, `Unassign()`, `UnassignAll()`, `Recruit()`
5. `Research.cs` — `StartResearch()`, research tick, unlock on complete
6. `Trade.cs` — `ActivateRoute()`, `DeactivateRoute()`, trade tick
7. `Diplomacy.cs` — faction relation shifts, bonus thresholds
8. `Military.cs` — unit recruitment, raid/defense calculations
9. `Expeditions.cs` — `StartExpedition()`, outcome resolution
10. `Milestones.cs` — progress tracking, completion detection, chain bonuses
11. `Catastrophes.cs` — roll logic, Endure/Succumb/Defer effects
12. `Prestige.cs` — validation, state reset, legacy bonus

**Unit test targets:** Every formula function gets a test. Every state transition gets a test. Use xUnit or NUnit (standard .NET test frameworks, both work with `dotnet test`).

```csharp
// Tests/ResourcesTests.cs
[Fact]
public void WorkerScaledProduction_FullStaff_ReturnsFullRate()
{
    double result = Resources.WorkerScaledProduction(1.0, 10, 10, 1);
    Assert.Equal(1.0, result);
}

[Fact]
public void WorkerScaledProduction_NoWorkers_ReturnsMinimum()
{
    double result = Resources.WorkerScaledProduction(1.0, 0, 10, 1);
    Assert.Equal(0.20, result); // 20% minimum production
}
```

### Week 4: Engine + Ticker Integration

- `Engine.cs` — wire all subsystems, implement `DoTick()`
- `Ticker.cs` — delta accumulator
- `SaveSystem.cs` — JSON serialize/deserialize `GameState`
- Integration test: create `GameEngine`, run 1000 ticks, verify resources accumulated correctly, verify autosave fires at tick 300

### Milestone
`dotnet test` passes 100% of ~80 unit tests. `GameEngine.DoTick()` runs 1000 ticks in under 50ms. `SaveSystem.Save()` + `TryLoad()` round-trips correctly.

---

## Phase 2: Terminal UI Shell
**Duration:** Weeks 5-7 (60 hours)
**Goal:** Playable via terminal in Godot. Feature parity with Classic's core commands.

### Week 5: Terminal Panel

- `TerminalPanel.tscn` — layout with `RichTextLabel` + `LineEdit` (see `architecture.md`)
- `OutputLog.cs` — `AppendLine()` with BBCode color support, scrollback to 2000 lines
- `CommandInput.cs` — text submission, command history (Up/Down), Tab completion hook
- CRT shader applied to `TerminalPanel`: scanlines + vignette at default settings
- Color theme system: `ThemeManager.cs`, default Amber theme
- Boot sequence: `BootSplash.tscn` with 3-4 second text scroll

### Week 6: Command Router

- `CommandRouter.cs` — maps command strings to engine calls, same command list as Classic
- Port all ~40 commands: `build`, `research`, `assign`, `unassign`, `recruit`, `gather`, `trade`, `spy`, `zone`, `list`, `info`, `status`, `help`, `settings`, `save`, `prestige`, etc.
- Error output (red), success output (green), info output (amber/default)
- `help` command renders command list with syntax hints
- `info <building>` renders building details

### Week 7: Autocomplete + Polish

- `AutoComplete.cs` — prefix match, ghost text inline, popup for multi-match
- Command syntax hint bar (above input, shows full syntax as player types)
- Per-command argument completion (e.g., `assign` completes to built building keys)
- Keypress sound effects hooked up (placeholder `.ogg` files acceptable at this stage)

### Milestone
Open game → boot sequence → type `build gathering_camp` → see construction progress → tick completes → see resource output. All Classic commands functional. Command history and autocomplete working.

---

## Phase 3: All Overlays
**Duration:** Weeks 8-10 (60 hours)
**Goal:** Full feature parity with Classic. All overlays functional.

### Overlay Architecture

Each overlay follows the same pattern:
1. `OverlayName.tscn` — a `CanvasLayer` containing a `PanelContainer` with terminal styling
2. `OverlayNameController.cs` — subscribes to relevant `EventBus` signals, updates `RichTextLabel` content
3. Command `overlay <name>` (or shorthand like `research`, `military`, `trade`) opens the overlay
4. Escape closes the current overlay

**Week 8:** Research, Workers, Stats overlays
**Week 9:** Military, Trade, Diplomacy, Wonders overlays
**Week 10:** Logs, Epoch, Wiki, History overlays + overlay manager polish

### Overlay Content

Port the rendering logic from Classic's tview tab renderers. The content strings (building lists, tech trees, faction tables) are largely identical. Replace tview color codes with BBCode `[color=]` tags.

Key polish items:
- Box-drawing character borders on all overlays (rendered as monospace text, not Godot UI borders)
- Scroll support in overlays with long content
- Keyboard navigation within overlays where applicable (e.g., Up/Down to scroll research list)

### Milestone
Full game playable via terminal + overlays. All 10 overlays functional. Feature parity with Classic confirmed by running the same game session and reaching the Bronze Age.

---

## Phase 4: City Map
**Duration:** Weeks 11-14 (80 hours)
**Goal:** Live city map visible and reactive to all game state changes.

### Week 11: Scene Setup + Road Network

- `GameScreen.tscn` — `HSplitContainer` with terminal (60%) and `SubViewportContainer` (40%)
- `SubViewport` with `Camera2D`, `TileMapLayer` for road tiles
- `RoadNetwork.cs` — port `mapv4.go`'s radial spoke algorithm; output `Vector2` positions
- Road rendering using `TileMapLayer` with road tile sprites (create simple placeholder tiles)
- `MapCamera.cs` — auto-zoom based on total building count (more buildings = zoom out)
- Scroll wheel zoom override

### Week 12: Building Sprites

- Source or commission placeholder building sprites (32x32 or 48x48 pixel art)
- One sprite per building type (can start with one sprite per category: farm, mine, mill, storage, military, wonder)
- `BuildingSprite.cs` — construction animation + idle animation stub
- `CityMapRenderer.cs` — subscribe to `BuildingBuilt`/`BuildingDestroyed`, place/remove sprites
- Verify: typing `build gathering_camp` places a sprite at the correct road network position

### Week 13: Backgrounds + Age Transitions

- Create or source 22 age background images (768×432 or 1024×576 pixel art)
- Backgrounds progress from wilderness → city → megacity → orbital
- Age background transitions on `AgeAdvanced` signal: crossfade via `AnimationPlayer`
- Update road tile style per age (dirt path → cobblestone → asphalt)

### Week 14: Animations + Particles

- Idle animation for mill, forge, farm (4-8 frame loops)
- Construction animation (scale up from 0 with easing)
- Building destruction animation (particles burst + sprite fades)
- Smoke `GpuParticles2D` on industrial buildings (scales with worker assignment)
- Wonder completion particle burst
- Age advance cinematic integration with city map (overlay dims, then new background fades in)

### Milestone
City map is visible alongside terminal. Buildings appear when constructed. Age transitions update background and road style. Smoke particles on forges and factories. Camera auto-zooms. Scroll wheel zoom works.

---

## Phase 5: Polish + Audio
**Duration:** Weeks 15-17 (60 hours)
**Goal:** Full audio implementation. Settings screen. Age advancement cinematic complete.

### Week 15: Audio

- Source all required audio files (ambient loops × 22, SFX × 8, ambient variants)
- `AudioManager.cs` singleton: `PlaySfx(name, pitch)`, `PlayAmbient(ageKey)`, `CrossfadeAmbient(fromKey, toKey, duration)`
- Connect all audio triggers:
  - `CommandInput._Input()` → keypress click
  - `BuildingBuilt` signal → build complete ding
  - `AgeAdvanced` signal → age advance fanfare + ambient crossfade
  - `CatastropheTriggered` signal → alarm SFX
  - `MilestoneCompleted` signal → achievement ding
  - Invalid command → error beep
- `AudioBus` layout: Master → Ambient (25% default), Master → SFX (75% default)

### Week 16: Settings Screen

- `settings` command opens settings overlay
- Controls: font size, CRT intensity, phosphor glow, color theme, ambient volume, SFX volume, reduce motion, high contrast
- Settings persisted to `user://settings.json`
- Apply settings immediately (no "apply" button — sliders are live)

### Week 17: Age Advancement Cinematic

- Full-screen cinematic: dark overlay → ASCII art age name → flavor text → unlock list → dismiss
- ASCII art for all 22 ages (can use figlet-generated text, adjusted manually)
- Phosphor glow shader at maximum during cinematic
- Boot sequence fully polished with actual game data (real save slot info, real loaded counts)
- Splash screen art (`assets/art/boot_logo.png`)

### Milestone
Audio plays correctly for all triggers. Ambient audio crossfades on age advance. Settings screen functional and persistent. Age advancement cinematic plays with full art and audio.

---

## Phase 6: Steam Integration
**Duration:** Weeks 18-20 (60 hours)
**Goal:** Steam working. Achievements, cloud saves, overlay functional. Ready for Steamworks review.

### Week 18: Achievements

- Define all ~65 achievements in Steamworks developer dashboard (achievement API names, display names, descriptions, icons)
- Implement `Achievements.cs` — `Unlock(apiName)` calls `SteamManager`
- Connect all achievement triggers to game events:
  - First command → "First Command"
  - Each age advance → age-specific achievement
  - Milestone completions → milestone achievement
  - Prestige count thresholds → prestige achievements
  - Challenge conditions → hidden achievements
- Verify in Steam client: achievements unlock and appear in Steam overlay

### Week 19: Cloud Saves + Steam Stats

- `SaveSystem.cs` — add `Steam.FileWrite()` call after every local save
- Load from Steam cloud on first launch if local save absent
- Conflict resolution: if cloud save is newer, prompt player to choose
- Steam stats: define 10 global stats in Steamworks dashboard (ages reached, builds completed, prestige count, etc.)
- Report stats on `SteamManager.StoreStats()` (called after every achievement unlock)

### Week 20: Final Steam Integration

- Steam overlay (`Shift+Tab` in-game): verify overlay appears and doesn't crash
- Rich Presence: `Steam.SetRichPresence()` showing current age and prestige count
  - Display: "Playing AgeForge Beta • Bronze Age • 0 prestiges"
- Steam Input API: optional, allows controller mapping (even if the game is keyboard-only, defining a default action set future-proofs controller support)
- Build and submit to Steamworks for store page review (store page requires at least 5 screenshots, a trailer, and capsule art)

### Milestone
Steam achievements unlock in Steam client. Cloud saves sync. Steam overlay functional. Steam rich presence shows current game state. Store page live in "Coming Soon" mode.

---

## Phase 7: Beta Testing + Launch
**Duration:** Weeks 21-24 (80 hours)
**Goal:** Public release on Steam.

### Week 21: Internal Testing

- Full playthroughs: reach Age 22 twice (new game, and continue from existing Classic save via migration tool)
- Save migration utility: `ClassicSaveConverter.cs` reads Classic's JSON format, outputs Beta-compatible save
- Stress test: 10 hours of game time, verify no memory leaks, no tick drift, no save corruption
- Bug bash: all overlays, all commands, all edge cases

### Week 22: Steam Next Fest Preparation

**Steam Next Fest** (occurs multiple times per year — target the one ~2 months after development completes). Requirements:
- Submit game to Next Fest at least 3 weeks before it starts
- A publicly visible demo is required — build a demo version (first 3 ages, save disabled)
- Demo build CI pipeline separate from main game
- Trailer: 60-90 second screen recording of the terminal game + city map, with CRT filter applied in post (OBS recording → DaVinci Resolve → add CRT overlay). Music: the Age 1 ambient track. Show: boot sequence, typing commands, city growing, age advance cinematic, catastrophe.

### Week 23: Press Kit + Store Page Polish

**Press kit contents:**
- 10 high-resolution screenshots (1920×1080): boot sequence, various ages, overlays, city map at different scales, catastrophe, wonder completion, age advance cinematic
- Trailer (from Week 22)
- Fact sheet: developer name, release date, price, platform, languages, features
- "About" text (150 words): the terminal civilization builder pitch
- Contact email

**Store page checklist:**
- Short description: "Build a civilization from the command line. 22 ages. 284 buildings. Type your way from primitive settlement to galactic empire."
- Long description: full feature overview with screenshots embedded
- Capsule art (616×353): glowing amber terminal window with city map visible
- Header art (460×215)
- 5 feature bullets
- System requirements (minimum: integrated graphics, 4GB RAM, 500MB storage)
- Tags: Idle, Strategy, City Builder, Text-Based, Simulation

### Week 24: Launch

**Launch checklist:**
- Final build tested on Windows, Mac, Linux
- All achievements working in production Steam environment (not development)
- Cloud saves syncing
- Store page live and approved by Valve
- Set launch date (Tuesday is the best day for Steam releases — avoids weekend noise)
- Discord announcement, social media
- Classic repo updated with link to Beta Steam page
- Classic README updated: "AgeForge Classic is the terminal original — always free at ageforge.io. AgeForge (Godot) is the full Steam release."
- Launch at $4.99

### Post-Launch (Week 25+)

- Monitor reviews and bug reports daily for first 2 weeks
- Patch 1.0.1 within 1 week of launch (fixes only, no new features)
- Content update roadmap published on Steam
- Begin Workshop implementation (v1.1)
- Begin alternate history DLC content (v1.1 or v1.2)

---

## Parallel Track: AgeForge Classic

Classic is not abandoned during Beta development. It continues to receive:
- Bug fixes and balance patches (on its own `master` branch)
- Player feedback informs Beta's design decisions
- New features that are trivially implementable in Classic but whose Beta version needs more thought

When Beta launches, Classic gets:
- A note in its README pointing to Beta's Steam page
- A final "1.x" release tag
- Classic stays permanently free at **ageforge.io** — the website is its home, not Steam

Classic and Beta coexist. They serve different audiences and validate each other.

---

## Risk Register

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| GodotSteam incompatibility with Godot version | Low | High | Pin Godot version; test GodotSteam integration in Phase 0 |
| Steam Direct $100 fee delays AppID | Low | Medium | Submit immediately in Phase 0; don't wait |
| C# hot-reload instability | Medium | Low | Accept cold-reload workflow; not a blocker |
| Building sprite sourcing takes too long | Medium | Medium | Use geometric placeholder sprites in Phases 1-4; commission art in Phase 5 |
| Config JSON export from Go has edge cases | Low | Medium | Validate all 284 buildings load correctly in Phase 1 unit tests |
| Classic save format incompatibility | Low | Medium | Build converter in Phase 7; document breaking changes |
| Audio sourcing | Medium | Medium | Source CC0 audio from Freesound.org early (no license issues, free) |
| Godot export template signing (Mac notarization) | Medium | High | Research Apple notarization for Godot exports early; budget extra time in Phase 6 |
| Valve review rejection | Low | Medium | Review Steamworks guidelines in Phase 0; don't submit experimental builds |
