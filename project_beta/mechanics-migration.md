# Mechanics Migration — AgeForge Classic to Beta

## Overview

The mechanical core of AgeForge is its most mature asset. 22 ages, 284 buildings, 52 techs, 33 milestones, and the prestige system have been balanced and debugged over many release cycles. The migration goal is to preserve every mechanic with zero regression while enabling the new systems described in `new-mechanics.md`.

This document maps every current system to its Beta equivalent.

---

## What Maps 1:1 (Logic Port Only)

These systems require no design changes. The port is a mechanical translation from Go to C# — same data structures, same formulas, same behavior.

### Resource Accumulation

**Classic:** `ge.doResourceTick()` iterates all buildings, computes rates based on worker scaling, adds to `GameState.Resources`.

**Beta:** `Resources.cs` — identical logic, called from `GameEngine.DoTick()`.

```go
// Classic (Go)
func workerScaledProduction(rate float64, assigned, capacity, count int) float64 {
    if capacity == 0 || count == 0 { return 0 }
    totalCap := capacity * count
    return rate * float64(count) * (0.20 + 0.80 * float64(assigned) / float64(totalCap))
}
```

```csharp
// Beta (C#) — identical formula
public static double WorkerScaledProduction(double rate, int assigned, int capacity, int count)
{
    if (capacity == 0 || count == 0) return 0;
    int totalCap = capacity * count;
    return rate * count * (0.20 + 0.80 * assigned / (double)totalCap);
}
```

The formula is identical. The only change is `float64` → `double` (same precision in C#) and Go's `float64()` cast → C#'s `(double)` cast.

### Building System

**Classic:** `Buildings` map in `GameState`, `BuildingState` struct with `Count`, `WorkersAssigned`, `WorkerCapacity`, `UnderConstruction`, `ConstructionProgress`.

**Beta:** `Dictionary<string, BuildingState>` in `GameState.cs`. All fields preserved. Build command flow:
1. Check resources sufficient (same cost check)
2. Check age requirement (same `def.AgeKey` check)
3. Check prerequisite techs (same `def.RequiresTechs` check)
4. Set `UnderConstruction = true`, `ConstructionProgress = 0`
5. Each tick: `ConstructionProgress += 1 / def.BuildTime`
6. When complete: `Count++`, `UnderConstruction = false`
7. Emit `BuildingBuilt` signal

Identical to Classic's behavior.

### Worker Assignment and Production Scaling

**Classic:** `AssignWorker(buildingKey, count)` — checks building exists, caps at capacity, updates `WorkerDomainState.Assignments[buildingKey]`.

**Beta:** `Workers.Assign(buildingKey, count)` — same checks, same formula. Domain auto-resolved from `def.WorkerDomain`.

The worker food drain formula (0.08/tick for food domain) and all domain-specific rules port directly.

### Research / Tech Tree

**Classic:** `ResearchedTechs HashSet[string]`, `CurrentResearch string`, `ResearchProgress float64`. Each tick adds `knowledgeRate` to progress until `def.Cost` reached.

**Beta:** Identical in `Research.cs`. The tech dependency graph (`def.Requires []string`) is checked on `research <tech>` command. All 52 tech definitions port from Go structs to `TechDef.cs` class.

### Trade Routes

**Classic:** `ActiveTradeRoutes HashSet[string]`. Each tick, active routes add/subtract resources at defined rates. Routes have `RequiresTech`, `RequiresAge`, `RequiresMilitary` gates.

**Beta:** `Trade.cs` — identical tick logic. Route definitions move from Go config to `data/trade_routes.json`.

### Diplomacy Factions

**Classic:** `FactionRelations map[string]float64`. Relation values shift based on events, player commands, and time. Faction bonuses apply when relations > threshold.

**Beta:** `Diplomacy.cs` — same model. All 6 faction definitions in `data/diplomacy.json`.

### Prestige System

**Classic:** On `prestige` command, validates age requirement (Age 22 or equivalent), records legacy bonus, resets `GameState` except `PrestigeCount` and `PrestigeLegacyBonus`, preserves bus.

**Beta:** `Prestige.cs` — same validation, same reset. The bus preservation pattern (do not `new EventBus()` on reset) is maintained because Godot signals still have their subscriber connections. `GameEngine.DoPrestige()` resets `State` fields selectively, same as Classic.

### Expeditions

**Classic:** `StartExpedition(key)` — costs resources, starts countdown, resolves on completion with reward roll.

**Beta:** `Expeditions.cs` — same state machine. All 15 expedition definitions in `data/expeditions.json`.

### Milestones

**Classic:** 33 milestones, 5 chains. Progress tracked in `MilestoneProgress[key]`. Completion grants title + tick speed boost via event.

**Beta:** `Milestones.cs` — identical. The "hidden until progress > 0.5 or preceding age" visibility rule preserved. Chain completion flow preserved.

### Catastrophe / Epoch System

**Classic:** 7 epochs (3 ages each). Faith/culture % of cap determines roll odds. Endure/Succumb/Defer choices. Endure = 20% buildings destroyed, Succumb = full reset with ruins + legacy, Defer = delay with penalty.

**Beta:** `Catastrophes.cs` — identical system. The roll formula, the faith/culture gates, the epoch progression, all port directly. The `CatastrophePending` flag and `PendingCatastropheType` in `GameState` are preserved.

### Save / Load

**Classic:** JSON via Go's `encoding/json`. Save path: `~/.ageforge/save.json`.

**Beta:** JSON via `System.Text.Json`. Save path: `user://save_slot_0.json` (Godot standard). JSON field names preserved for potential save migration utility. See note on `json:"villagers"` → `[JsonPropertyName("workers")]` compatibility.

---

## What Changes Significantly

### Tick System: Goroutine → Delta Accumulator

**Classic:** A goroutine runs `time.Sleep(tickInterval)` in a loop, then calls `ge.doTick()` under a write lock. Speed changes update `tickInterval`. Start/stop uses a `stopCh` channel.

**Beta:** The `Ticker` node accumulates delta in `_Process(double delta)`. When accumulated >= `_interval`, it fires `DoTick()`. No goroutine, no mutex needed. `SetSpeed(multiplier)` changes `_interval = MinTickInterval / multiplier`.

**Migration impact:** The autosave logic (fires every N ticks outside the lock in Classic) becomes a simple tick counter in `DoTick()` with no lock concerns.

### UI Rendering: tview → Godot Nodes

**Classic:** tview `TextView`, `InputField`, `Table`, `Pages` — all rendered to terminal cells.

**Beta:** `RichTextLabel` for output, `LineEdit` for input, `Container` nodes for layout. The content (the text being rendered) ports almost directly — the rendering primitives change.

**Overlay content strings:** Classic generates strings like:
```go
fmt.Sprintf("[yellow]%s[-] [%s]%s[-] Cost: %s\n", def.Name, color, status, costStr)
```

Beta generates BBCode for `RichTextLabel`:
```csharp
$"[color=#ffb86c]{def.Name}[/color] [{color}]{status}[/color] Cost: {costStr}\n"
```

The transformation is mechanical. The text content and layout logic are identical.

### Input: Blocking Readline → Godot LineEdit Signals

**Classic:** `ui.App.SetInputCapture()` intercepts key events. Commands submitted via `inputField.SetDoneFunc()`.

**Beta:** `CommandInput` wraps a `LineEdit`. `TextSubmitted` signal fires on Enter. `_Input()` handles Up/Down for history and Tab for completion. The signal-based model is cleaner and requires no special input capture hackery.

### Autocomplete: Custom tview → Godot AutoComplete Node

**Classic:** `ui/autocomplete.go` — custom completion logic, rendered into tview overlay.

**Beta:** `AutoComplete.cs` — same prefix-match completion logic (which commands are valid, which building/tech keys match the prefix). Rendered as either:
- Ghost text in a Label overlay on top of the `LineEdit` (inline completion, like VSCode's gray text)
- A `PopupMenu` below the cursor showing options

**Recommendation:** Ghost text for single-match completions, popup menu when there are 2+ matches. This is more sophisticated than Classic's implementation.

### Map Rendering: Half-Block Terminal Chars → Godot 2D Sprites

**Classic (mapv4):** The city map uses `▀`, `▄`, `█`, `░`, `▒`, `▓` half-block characters in a tview panel, algorithmically placed to simulate a city grid.

**Beta:** Real Godot `Sprite2D` nodes with actual pixel art. The road spoke layout algorithm (from Go's `mapv4.go`) ports to `RoadNetwork.cs` — same math, same radial spoke placement, but outputs `Vector2` positions for sprite placement instead of character coordinates.

The half-block approach was a creative solution to terminal constraints. In Beta it is replaced by proper sprites. The algorithm's core logic (radial spokes, distance-based density, age-appropriate building counts) is preserved because it produces a good city layout.

---

## New Mechanics Possible Only in Godot

### Animated Building Sprites

Every building has an idle animation. Mills spin their blades. Forges pulse with an orange glow. Cathedrals have animated windows. Farms have gentle crop-sway.

**Implementation:** Each building sprite sheet has a 4-8 frame idle animation. `AnimationPlayer` cycles through frames at 4-8 FPS. Construction animation (building "grows" from 0 scale to full scale over `buildTime` ticks) uses `Tween`.

```csharp
// BuildingSprite.cs
public void PlayConstructionAnimation()
{
    Scale = Vector2.Zero;
    var tween = CreateTween();
    tween.TweenProperty(this, "scale", Vector2.One, 0.5f)
         .SetTrans(Tween.TransitionType.Back)
         .SetEase(Tween.EaseType.Out);
    tween.TweenCallback(Callable.From(StartIdleAnimation));
}

public void StartIdleAnimation()
{
    _animPlayer.Play("idle");
}
```

### Particle Effects

**Smoke from productive buildings:** `GpuParticles2D` node on forges, factories, power plants. Particle count scales with `WorkersAssigned / WorkerCapacity`. A fully-staffed factory pours smoke; an empty one is cold.

**Sparkle from wonders:** Wonders have a permanent ambient particle effect (gold sparkles for Pyramid, blue light beams for Observatory, etc.)

**Catastrophe particles:** Earthquake → dust cloud `GpuParticles2D` with `explosion_dust.png`. Plague → green fog using `GpuParticles2D` with `fog_particle.png` and low emission angle spread.

### Screen Shake on Catastrophes

```csharp
// CatastropheEffect.cs
public static void ShakeScreen(float intensity, float duration)
{
    if (Settings.ReduceMotion) return;
    var tween = GameHUD.Instance.CreateTween();
    var camera = CityMap.Instance.Camera;
    float elapsed = 0f;
    while (elapsed < duration)
    {
        var offset = new Vector2(
            GD.Randf() * intensity - intensity / 2,
            GD.Randf() * intensity - intensity / 2);
        tween.TweenProperty(camera, "offset", offset, 0.05f);
        elapsed += 0.05f;
    }
    tween.TweenProperty(camera, "offset", Vector2.Zero, 0.1f);
}
```

### Real Audio

The entire ambient audio system described in `terminal-aesthetic.md` requires a real audio engine. Classic has no audio at all. Beta ships with:
- Keypress sounds
- Build complete sound
- Age advance sound
- Catastrophe alarm
- Wonder completion sound
- Looping ambient audio per age (crossfaded on age advance)

All via Godot's `AudioStreamPlayer` and `AudioBus` system. No external dependency.

### Age Advancement Cinematic

Classic shows an overlay panel for age advance. Beta shows a full-screen cinematic (described in `terminal-aesthetic.md`). The city map background transitions via a crossfade shader. Building sprites update in the background during the cinematic.

### Animated Resource Flow on City Map

Active trade routes show animated "resource packets" — small colored dots — moving along the road network from source buildings to target storage. These are `GpuParticles2D` with path-following behavior.

```csharp
// TradeFlowVisualizer.cs — fires when trade route is active
private void AnimateFlow(Vector2 from, Vector2 to, Color color)
{
    var particles = TradeParticleScene.Instantiate<GpuParticles2D>();
    particles.GlobalPosition = from;
    // Configure process material to emit toward `to`...
    CityMap.Instance.ParticleContainer.AddChild(particles);
}
```

### Steam Achievements

33 milestones map naturally to Steam achievements. Additionally:

| Achievement | Trigger |
|-------------|---------|
| "First Command" | Type first command |
| "Pioneer" | Complete first building |
| "Researcher" | Research first tech |
| "Age Advancement x22" | Reach each age (22 achievements) |
| "Prestige I-V" | First 5 prestige cycles |
| "Wonder Builder" | Complete any wonder |
| "Catastrophe Survivor" | Survive first catastrophe (Endure) |
| "Clean Slate" | Succumb to first catastrophe |
| "Empire of Ages" | Complete all 22 ages in one run |
| "Speed Runner" | Reach Age 10 within 2 hours |
| "Completionist" | Complete all 33 milestones |
| Milestone completions x33 | Each milestone |

```csharp
// Achievements.cs
public static void Unlock(string apiName)
{
    if (!SteamManager.Available) return;
    SteamManager.Instance.UnlockAchievement(apiName);
}

// Called from Milestones.cs:
GameEngine.Instance.Bus.MilestoneCompleted += (key) =>
    Achievements.Unlock($"milestone_{key}");
```

### Settings Screen

Classic has no settings screen (it's a terminal app). Beta has a full settings screen accessible via `settings` command or Escape menu:
- Resolution (windowed/fullscreen)
- Font size
- CRT shader intensity
- Color theme
- Volume controls (master, ambient, SFX)
- Accessibility options
- Key bindings (for future gamepad support)
- Save slot management

### Accessibility Options

Classic inherits terminal accessibility (screen reader compatibility, font scaling). Beta adds:
- Font size slider (11-20px)
- High contrast mode (removes glow, maximum color contrast)
- Reduce motion (disables screen shake, build animations, particle effects)
- Cursor size options
- Color blind mode (replaces semantic color scheme with pattern+symbol differentiation)

---

## What Gets Cut in Beta 1

These elements from Classic do not carry forward into Beta 1. They are either replaced by better implementations or removed as development artifacts.

### mapv1 / mapv2 / mapv3

The experimental map renderers were stepping stones to mapv4 (the half-block city map). In Beta, the city map is a real Godot 2D scene. The algorithm from mapv4 (radial spoke layout) is preserved in `RoadNetwork.cs`, but the half-block rendering is gone.

### Raw tview Terminal Rendering

The entire `ui/` package in Classic uses tview. In Beta, the `ui/` package equivalent is Godot scenes and C# scripts. Nothing from the tview import is carried forward.

### Half-Block Pixel Art Map

The `▀`, `▄`, `█` character art from mapv4 is replaced by real sprites. It was impressive for a terminal game. In a real game engine it would look like a regression.

### Go Build System

`go build`, `go test`, `Makefile` targets — all replaced by Godot's export system and a `dotnet test` equivalent for unit tests. The `make release-patch/minor/major` release process gets a new implementation using the Godot export pipeline + GitHub Actions.

### Direct Terminal Output

Classic uses `fmt.Fprintf` and tview's buffer. Beta uses `RichTextLabel.AppendText()`. The interface is different but the content format (command output strings) ports directly.

---

## Migration Difficulty by System

| System | Difficulty | Notes |
|--------|-----------|-------|
| Resource accumulation | Low | Pure math, direct port |
| Building system | Low | Struct → class, same logic |
| Worker system | Low | Direct port |
| Research tree | Low | Direct port |
| Trade routes | Low | Direct port |
| Diplomacy | Low | Direct port |
| Prestige | Low | Direct port |
| Expeditions | Low | Direct port |
| Milestones | Low | Direct port |
| Catastrophe/epoch | Medium | Same logic, new UI for Endure/Succumb choice |
| Save/load | Medium | Format change, migration utility needed |
| Command routing | Medium | tview input → Godot LineEdit |
| Tick system | Medium | Goroutine → delta accumulator |
| Autocomplete | Medium | Logic port + new render layer |
| Overlay panels | Medium | String content ports, render layer changes |
| Terminal panel | Medium | New from scratch in Godot |
| City map | High | Algorithm port + new sprite system |
| Audio system | High | New (Classic has none) |
| CRT shaders | High | New (Classic has none) |
| Steam integration | High | New (Classic has none) |
| Age cinematics | High | New (Classic has basic overlay) |
