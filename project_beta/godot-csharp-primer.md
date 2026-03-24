# Godot 4 + C# Primer for Go Developers

## Overview

This document is a translation guide for a developer who knows Go well and is learning Godot 4 with C#. It covers the Go patterns you use daily and their direct equivalents in Godot C#. It is not a comprehensive Godot tutorial — it is the minimum viable knowledge to start porting AgeForge.

**Recommended learning path before starting the port:**
1. Read this document fully
2. Complete the official Godot 4 "Your First 2D Game" tutorial (~2 hours): https://docs.godotengine.org/en/stable/getting_started/first_2d_game/index.html
3. Read the Godot C# basics page: https://docs.godotengine.org/en/stable/tutorials/scripting/c_sharp/index.html
4. Then start Phase 1

---

## Core Mental Models

### The Node Tree

Every Godot object that exists in your game is a **Node** in a tree. The scene editor shows this tree visually. When you run the game, Godot traverses the tree calling `_Ready()` (initialization) on every node, then `_Process(delta)` (per-frame update) on every node that implements it.

```
Main (Node)
├── GameEngine (Node)           ← Autoload singleton
│   ├── Ticker (Node)           ← Fires DoTick() via _Process
│   └── EventBus (RefCounted)   ← Not a Node, just a C# object owned by Engine
├── GameScreen (Control)
│   ├── TerminalPanel (PanelContainer)
│   │   ├── OutputLog (RichTextLabel)
│   │   └── CommandInput (LineEdit)
│   └── CityMapViewport (SubViewportContainer)
│       └── SubViewport
│           └── CityMapRenderer (Node2D)
```

Nodes are instantiated either in the editor (drag to scene tree) or in code:
```csharp
var ticker = new Ticker();
AddChild(ticker); // Ticker is now part of this node's subtree
```

### Scenes

A **Scene** is a saved node tree. `Main.tscn` is the root scene. `TerminalPanel.tscn` is a sub-scene you can instantiate multiple times. Think of scenes as composable templates — like Go structs, but visual and hierarchical.

```csharp
// Instantiate a scene from code:
var overlayScene = GD.Load<PackedScene>("res://scenes/overlays/ResearchOverlay.tscn");
var overlay = overlayScene.Instantiate<ResearchOverlay>();
AddChild(overlay);
```

### Autoload Singletons

Godot has a concept called **Autoload** — specific scripts/scenes that are instantiated at launch and persist for the entire session, accessible from anywhere. This is how `GameEngine` works in Beta.

Register an autoload in `Project → Project Settings → Autoload`:
```
Name: GameEngine
Path: res://src/core/Engine.cs
```

Access it from any script:
```csharp
var state = GameEngine.Instance.State;
```

This replaces Go's package-level variables (which AgeForge Classic avoids via dependency injection — the Autoload pattern is the Godot-idiomatic equivalent of a passed dependency).

---

## Go → C# Pattern Translations

### Goroutines → `_Process` / `Timer` Nodes

**Go:**
```go
func (ge *GameEngine) Start() {
    go func() {
        for {
            select {
            case <-ge.stopCh:
                return
            case <-time.After(ge.tickInterval):
                ge.doTick()
            }
        }
    }()
}
```

**C# (Godot):**
```csharp
// Ticker.cs — inherits Node, added as child of GameEngine
public partial class Ticker : Node
{
    private double _accumulated = 0.0;
    private double _interval = 0.2;

    public override void _Process(double delta)
    {
        _accumulated += delta;
        while (_accumulated >= _interval)
        {
            _accumulated -= _interval;
            GameEngine.Instance.DoTick();
        }
    }
}
```

There is no goroutine. `_Process(double delta)` is called by Godot every frame (60fps). The accumulator pattern replicates a fixed-interval timer without goroutines. No `select`, no channel, no `stopCh`. To "stop" the ticker, call `SetProcess(false)`:

```csharp
public void Stop() => SetProcess(false);
public void Start() => SetProcess(true);
```

**For one-shot delays:** Use a `Timer` node:
```csharp
var timer = new Timer();
timer.WaitTime = 2.0f;
timer.OneShot = true;
timer.Timeout += OnTimerFired;
AddChild(timer);
timer.Start();
```

**For repeated timers:** Same but `OneShot = false`. Or just use the `_Process` accumulator.

---

### Go Channels → Godot Signals

Go channels decouple producers from consumers. Godot signals do the same thing on the main thread.

**Go (event bus with channels/callbacks):**
```go
type EventBus struct {
    handlers map[string][]func(payload interface{})
    mu       sync.RWMutex
}

func (eb *EventBus) Subscribe(event string, handler func(interface{})) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    eb.handlers[event] = append(eb.handlers[event], handler)
}

func (eb *EventBus) Emit(event string, payload interface{}) {
    eb.mu.RLock()
    defer eb.mu.RUnlock()
    for _, h := range eb.handlers[event] {
        h(payload) // runs under lock — DEADLOCK DANGER
    }
}
```

**C# (Godot signals):**
```csharp
public partial class EventBus : RefCounted
{
    // Declare signal with typed parameters
    [Signal] public delegate void BuildingBuiltEventHandler(string key, int count);
    [Signal] public delegate void AgeAdvancedEventHandler(string ageKey);

    // Emit (no lock needed — single-threaded)
    public void EmitBuildingBuilt(string key, int count) =>
        EmitSignal(SignalName.BuildingBuilt, key, count);
}

// Subscribe from another script:
GameEngine.Instance.Bus.BuildingBuilt += OnBuildingBuilt;

// Handler:
private void OnBuildingBuilt(string key, int count) { /* ... */ }
```

**Key difference from Go:** Godot signals are dispatched synchronously on the main thread. No lock. No deadlock danger. You can freely read `GameEngine.Instance.State` inside a signal handler — unlike Classic where calling `GetState()` inside a bus handler causes a deadlock.

**Disconnecting signals:** If a node subscribes to a signal in `_Ready()`, it should unsubscribe in `_ExitTree()` to prevent use-after-free:
```csharp
public override void _ExitTree()
{
    GameEngine.Instance.Bus.BuildingBuilt -= OnBuildingBuilt;
}
```

---

### Go Structs → C# Classes

Go structs are value types. C# classes are reference types. For game state, this distinction matters.

**Go:**
```go
type BuildingState struct {
    Count            int
    WorkersAssigned  int
    WorkerCapacity   int
    UnderConstruction bool
    ConstructionProgress float64
}
```

**C#:**
```csharp
public class BuildingState
{
    public int Count { get; set; } = 0;
    public int WorkersAssigned { get; set; } = 0;
    public int WorkerCapacity { get; set; } = 0;
    public bool UnderConstruction { get; set; } = false;
    public double ConstructionProgress { get; set; } = 0.0;
}
```

Properties with `{ get; set; }` behave like public fields for JSON serialization and general use. Use them instead of public fields to allow future validation logic without API changes.

**When to use `Resource` vs plain class:**
- Use `Resource` (inherits `Godot.Resource`) for objects that need to be saved/loaded as Godot assets via the editor (like `ThemeSettings`, `AudioConfig`)
- Use plain C# classes for pure game state (`BuildingState`, `GameState`) — simpler to JSON-serialize
- Use `Node` for objects that exist in the scene tree and need `_Process`, `_Ready`, signals

---

### Go Maps → C# `Dictionary<TKey, TValue>`

**Go:**
```go
resources := map[string]float64{
    "food": 100.0,
    "wood": 50.0,
}

amount := resources["food"]                    // panics if missing
amount, ok := resources["food"]               // safe two-value form
resources["stone"] = 0                        // add/update
delete(resources, "stone")                    // remove
```

**C#:**
```csharp
var resources = new Dictionary<string, double>
{
    ["food"] = 100.0,
    ["wood"] = 50.0,
};

double amount = resources["food"];            // throws if missing
bool ok = resources.TryGetValue("food", out double val); // safe form
resources["stone"] = 0;                      // add/update
resources.Remove("stone");                   // remove
resources.ContainsKey("iron");               // equivalent to _, ok := map[key]

// Iteration:
foreach (var (key, value) in resources)
{
    GD.Print($"{key}: {value}");
}
```

**Important:** C# `Dictionary` preserves insertion order in .NET 5+ (implementation detail, not guaranteed by spec). Go maps do not. If order matters, use `List<KeyValuePair<string, double>>` or `SortedDictionary<string, double>`.

---

### Go Interfaces → C# Interfaces

**Go:**
```go
type Tickable interface {
    OnTick(state *GameState) error
}

type Building struct { ... }
func (b *Building) OnTick(state *GameState) error { ... }
```

**C#:**
```csharp
public interface ITickable
{
    void OnTick(GameState state);
}

public class Buildings : ITickable
{
    public void OnTick(GameState state) { /* ... */ }
}
```

The concepts are identical. The syntax differs slightly: Go uses implicit interface satisfaction (any type with matching methods satisfies the interface), C# requires explicit `class Foo : IBar` declaration. The explicit declaration is clearer.

---

### Go Error Returns → C# Exceptions (or Result Pattern)

**Go:**
```go
func (b *Buildings) Build(state *GameState, key string) error {
    def, ok := config.BuildingByKey(key)
    if !ok {
        return fmt.Errorf("unknown building: %s", key)
    }
    if !canAfford(state, def.Cost) {
        return fmt.Errorf("insufficient resources")
    }
    // ... build it
    return nil
}
```

**C# (exceptions — idiomatic C#):**
```csharp
public void Build(GameState state, string key)
{
    if (!ConfigLoader.Buildings.TryGetValue(key, out var def))
        throw new ArgumentException($"Unknown building: {key}");
    if (!CanAfford(state, def.Cost))
        throw new InvalidOperationException("Insufficient resources");
    // ... build it
}

// Caller:
try { _buildings.Build(State, key); }
catch (Exception e) { GameEngine.Instance.Bus.EmitLog(e.Message, "error"); }
```

**C# (Result pattern — closer to Go):**
```csharp
public record BuildResult(bool Success, string? Error = null);

public BuildResult TryBuild(GameState state, string key)
{
    if (!ConfigLoader.Buildings.TryGetValue(key, out var def))
        return new BuildResult(false, $"Unknown building: {key}");
    if (!CanAfford(state, def.Cost))
        return new BuildResult(false, "Insufficient resources");
    // ... build it
    return new BuildResult(true);
}

// Caller (no try/catch needed):
var result = _buildings.TryBuild(State, key);
if (!result.Success)
    GameEngine.Instance.Bus.EmitLog(result.Error!, "error");
```

**Recommendation for AgeForge Beta:** Use the Result pattern for game command dispatch (build, research, assign). These are expected failure paths (bad input from the player) and should not throw exceptions. Use exceptions for programmer errors (missing config keys, null references that should never happen).

---

### Go `init()` → Godot `_Ready()`

**Go:**
```go
var allBuildings []BuildingDef

func init() {
    allBuildings = loadBuildingsFromConfig()
}
```

**C# (Godot):**
```csharp
public partial class ConfigLoader : Node  // or just a static class
{
    public static Dictionary<string, BuildingDef> Buildings { get; private set; } = new();

    public override void _Ready()
    {
        // Called when the node enters the scene tree
        LoadAll();
    }

    public static void LoadAll()
    {
        Buildings = LoadJson<Dictionary<string, BuildingDef>>("res://data/buildings/...");
        // ...
    }
}
```

For config that doesn't need to be a Node, use a static class with an explicit `LoadAll()` call from `GameEngine._Ready()`. This is cleaner than relying on `_Ready()` ordering across different nodes.

---

### File I/O: `os.Open` → `FileAccess.Open`

**Go:**
```go
f, err := os.Open("saves/game.json")
if err != nil { return err }
defer f.Close()
data, err := io.ReadAll(f)
```

**C# (Godot):**
```csharp
// Use Godot's FileAccess for game files (works in exports correctly)
using var f = FileAccess.Open("user://save_slot_0.json", FileAccess.ModeFlags.Read);
if (f == null)
{
    GD.PrintErr($"File not found: {FileAccess.GetOpenError()}");
    return false;
}
string json = f.GetAsText();
// f disposed automatically by `using`
```

**Path conventions:**
- `res://` — the game's installation directory (read-only in exports). Use for config, assets.
- `user://` — the user's save data directory (read-write). Use for saves, settings.

**Do not use `System.IO.File`** for game files — it doesn't understand `res://` or `user://` paths and breaks in packaged exports.

---

### JSON: `encoding/json` → `System.Text.Json`

**Go:**
```go
data, _ := json.Marshal(state)
err := json.Unmarshal(data, &state)
```

**C#:**
```csharp
using System.Text.Json;
using System.Text.Json.Serialization;

// Serialize
string json = JsonSerializer.Serialize(state, new JsonSerializerOptions
{
    WriteIndented = false,
    PropertyNamingPolicy = JsonNamingPolicy.CamelCase
});

// Deserialize
var state = JsonSerializer.Deserialize<GameState>(json,
    new JsonSerializerOptions { PropertyNameCaseInsensitive = true });
```

**JSON field name control:**
```csharp
public class WorkerState
{
    [JsonPropertyName("villagers")]  // preserve Classic save format
    public Dictionary<string, WorkerDomainState> Workers { get; set; } = new();

    [JsonIgnore]                     // don't serialize DevMode
    public bool DevMode { get; set; } = false;
}
```

---

## Key Godot Concepts for AgeForge

### `_Process` vs `_PhysicsProcess`

- `_Process(double delta)` — called every render frame (~60fps). Use for: ticker, UI updates, visual animations.
- `_PhysicsProcess(double delta)` — called at a fixed rate (default 60Hz, independent of render). Use for: physics simulations, collision detection.

AgeForge Beta uses `_Process` only. There is no physics.

### `Callable`

Godot's way to hold a reference to a method for deferred calls or connecting signals:

```csharp
// Deferred call (runs next frame, not immediately)
CallDeferred(Callable.From(DoSomething));

// Pass as argument:
timer.Timeout += Callable.From(() => GD.Print("fired")).AsDelegate<Action>();
```

You will encounter `Callable` most often when connecting signals from code or when deferring calls to the next frame to avoid modifying the scene tree during `_Process`.

### `Tween`

Godot's built-in animation utility. Replaces manual lerp loops:

```csharp
// Animate a node's scale from 0 to 1 over 0.5 seconds, with a bounce ease
var tween = CreateTween();
tween.TweenProperty(node, "scale", Vector2.One, 0.5f)
     .SetTrans(Tween.TransitionType.Back)
     .SetEase(Tween.EaseType.Out);
```

Tweens are fire-and-forget (the tween object handles itself). You can chain them with `.Then()` or `.Parallel()`.

### Scene Instancing

Scenes can be instantiated at runtime. Use this for building sprites, overlay panels, particle effects:

```csharp
private PackedScene _buildingScene = GD.Load<PackedScene>("res://scenes/BuildingSprite.tscn");

private void SpawnBuilding(string key, Vector2 position)
{
    var sprite = _buildingScene.Instantiate<BuildingSprite>();
    sprite.Position = position;
    sprite.SetBuilding(key);
    _buildingContainer.AddChild(sprite);
}
```

### GD.Print / GD.PrintErr

Equivalent to `fmt.Println` / `log.Println`. Output appears in the Godot editor console:

```csharp
GD.Print($"Building built: {key}");     // equivalent to fmt.Printf
GD.PrintErr($"Config load failed: {path}"); // appears in red in editor
```

In exports (shipped game), output goes to a log file in `user://logs/`.

---

## Differences in Philosophy

### No Goroutines — Single-Threaded by Default

Godot's main loop is single-threaded. `_Process` runs on the main thread. Signals fire on the main thread. This means:

- No deadlocks from mutexes
- No race conditions
- No need for `sync.RWMutex`
- No `go func()` patterns

The tradeoff: if a single `_Process` call takes too long (more than ~8ms at 120fps), the game stutters. AgeForge's `DoTick()` is fast (pure data math on <10KB of state), so this is not a concern.

If you ever need true background work (e.g., a large file operation), Godot has `Thread` nodes. For AgeForge Beta, you will not need them.

### Properties vs Fields

Go uses exported fields directly. C# convention uses properties. For game state:

```csharp
// Preferred: property
public int Count { get; set; } = 0;

// Acceptable: public field (simpler but less flexible)
public int Count = 0;
```

`System.Text.Json` serializes both. The property form allows adding a getter/setter body later without changing callers. Prefer properties for `GameState` and def classes.

### `partial` Classes

Godot C# requires `partial class` declarations for Node scripts because Godot generates additional code in a separate partial file. Always include `partial` when inheriting from a Godot type:

```csharp
public partial class CommandInput : LineEdit { }  // correct
public class CommandInput : LineEdit { }           // will fail to compile
```

For pure C# classes that don't inherit from Godot types, `partial` is not required:
```csharp
public class GameState { }  // no partial needed — not a Godot type
```

---

## Quick Reference Card

| Go | C# (Godot) |
|----|-----------|
| `goroutine` | `_Process` delta accumulator |
| `time.Sleep(200ms)` | Delta accumulator in `Ticker` |
| `chan T` | `[Signal] delegate`, `EmitSignal` |
| `sync.RWMutex` | Not needed (single-threaded) |
| `struct Foo {}` | `public class Foo {}` |
| `map[string]float64` | `Dictionary<string, double>` |
| `interface Foo {}` | `public interface IFoo {}` |
| `error` return | `Exception` or Result record |
| `init()` | `_Ready()` or explicit `LoadAll()` |
| `fmt.Println` | `GD.Print()` |
| `os.Open(path)` | `FileAccess.Open("user://...", ...)` |
| `json.Marshal` | `JsonSerializer.Serialize` |
| `json.Unmarshal` | `JsonSerializer.Deserialize<T>` |
| `defer f.Close()` | `using var f = ...` (IDisposable) |
| Global var | `Autoload` singleton |
| Package | `namespace` |
| `go test` | `dotnet test` |
| `go build` | Godot editor export |

---

## Common Gotchas

**1. `null` is not `nil`**
C# uses `null`. Go uses `nil`. Same concept, but C# has nullable reference types (enabled in modern projects). You'll see `string?` (nullable string) vs `string` (non-null). The compiler warns on potential null dereferences. Take these warnings seriously.

**2. `string.Format` vs Go fmt verbs**
Go: `fmt.Sprintf("value: %.2f", amount)` → C#: `$"value: {amount:F2}"` (string interpolation) or `string.Format("value: {0:F2}", amount)`. Interpolation is preferred.

**3. Integer division**
Go: `int / int = int` (truncates). C#: same. But `double / int` in C# requires the int to be cast: `(double)count / total` not `count / total` when you want a fractional result.

**4. `foreach` doesn't support index**
Go: `for i, v := range slice`. C#: `foreach` has no index. Use `for (int i = 0; i < list.Count; i++)` when you need the index.

**5. Scene tree modifications during `_Process`**
Don't call `AddChild()` or `RemoveChild()` from inside `_Process`. Use `CallDeferred(...)` to defer it to the end of the frame. This is a common Godot gotcha.

**6. `res://` paths are case-sensitive on Linux**
On Windows and Mac, `res://Assets/Fonts/JetBrainsMono.ttf` and `res://assets/fonts/JetBrainsMono.ttf` both work. On Linux, only the exact case works. Use consistent lowercase paths everywhere.
