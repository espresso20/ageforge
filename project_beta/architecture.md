# Technical Architecture — AgeForge Beta (Godot 4 + C#)

## Project Structure

```
project-ageforge/
├── project.godot                    # Godot project config
├── steam_appid.txt                  # Steam AppID (required at launch)
├── .gitignore
├── .github/
│   └── workflows/
│       ├── build.yml                # CI: build Windows/Mac/Linux
│       └── export.yml               # Steam export pipeline
├── addons/
│   └── godotsteam/                  # GodotSteam plugin (prebuilt GDExtension)
│       ├── godotsteam.gdextension
│       ├── libgodotsteam.windows.x86_64.dll
│       ├── libgodotsteam.linux.x86_64.so
│       └── libgodotsteam.macos.universal.dylib
├── src/
│   ├── core/                        # Ported game logic (from Go game/ package)
│   │   ├── Engine.cs                # GameEngine singleton (was engine.go)
│   │   ├── GameState.cs             # Pure data class (was state.go structs)
│   │   ├── Ticker.cs                # Delta accumulator tick loop
│   │   ├── EventBus.cs              # Godot-signal-based event bus
│   │   ├── Resources.cs             # Resource accumulation logic
│   │   ├── Buildings.cs             # Building state manager
│   │   ├── Workers.cs               # Worker assignment & production scaling
│   │   ├── Research.cs              # Tech tree unlocks
│   │   ├── Military.cs              # Army, raids, defense
│   │   ├── Trade.cs                 # Trade routes
│   │   ├── Diplomacy.cs             # Faction relations
│   │   ├── Prestige.cs              # Prestige / new game+
│   │   ├── Expeditions.cs           # Expedition system
│   │   ├── Milestones.cs            # 33 milestones, 5 chains
│   │   ├── Catastrophes.cs          # Catastrophe / epoch system
│   │   └── SaveSystem.cs            # JSON save/load + Steam cloud
│   ├── config/                      # Config loader (was config/ package)
│   │   ├── ConfigLoader.cs          # Reads JSON, populates static defs
│   │   ├── BuildingDef.cs           # C# equivalent of Go BuildingDef struct
│   │   ├── AgeDef.cs
│   │   ├── TechDef.cs
│   │   ├── MilestoneDef.cs
│   │   ├── WorkerDomainDef.cs
│   │   ├── ResourceDef.cs
│   │   └── EpochDef.cs
│   ├── ui/
│   │   ├── terminal/                # Terminal emulator UI
│   │   │   ├── TerminalPanel.cs     # Main terminal container controller
│   │   │   ├── CommandInput.cs      # LineEdit wrapper + command routing
│   │   │   ├── OutputLog.cs         # RichTextLabel scrollback buffer
│   │   │   ├── AutoComplete.cs      # Inline ghost text + popup completion
│   │   │   └── CommandRouter.cs     # Maps command strings → engine calls
│   │   ├── overlays/                # Overlay panels (tab equivalents)
│   │   │   ├── OverlayManager.cs    # Show/hide overlay by name
│   │   │   ├── ResearchOverlay.cs
│   │   │   ├── MilitaryOverlay.cs
│   │   │   ├── TradeOverlay.cs
│   │   │   ├── DiplomacyOverlay.cs
│   │   │   ├── StatsOverlay.cs
│   │   │   ├── WikiOverlay.cs
│   │   │   ├── WondersOverlay.cs
│   │   │   ├── WorkersOverlay.cs
│   │   │   ├── LogsOverlay.cs
│   │   │   └── EpochOverlay.cs
│   │   ├── map/                     # City map renderer
│   │   │   ├── CityMapController.cs # Subscribes to GameState changes
│   │   │   ├── CityMapRenderer.cs   # Places/removes building sprites
│   │   │   ├── BuildingSprite.cs    # Individual building node with animation
│   │   │   ├── RoadNetwork.cs       # Road spoke layout logic (ported from Go)
│   │   │   └── MapCamera.cs         # Auto-zoom Camera2D
│   │   ├── effects/
│   │   │   ├── AgeAdvancementScreen.cs  # Full-screen age cinematic
│   │   │   ├── CatastropheEffect.cs     # Screen shake, particle overlays
│   │   │   └── WonderReveal.cs          # Wonder completion cinematic
│   │   └── GameHUD.cs               # Status bar, speed controls, devmode badge
│   └── steam/
│       └── SteamManager.cs          # Steam init, achievements, cloud saves
├── data/                            # Game data (was Go config maps)
│   ├── ages.json
│   ├── resources.json
│   ├── techs.json
│   ├── milestones.json
│   ├── events.json
│   ├── expeditions.json
│   ├── trade_routes.json
│   ├── diplomacy.json
│   ├── epochs.json
│   ├── worker_domains.json
│   └── buildings/
│       ├── lineage_food.json
│       ├── lineage_wood.json
│       ├── lineage_stone.json
│       ├── lineage_iron.json
│       ├── lineage_military.json
│       ├── lineage_faith.json
│       ├── lineage_culture.json
│       ├── lineage_knowledge.json
│       ├── lineage_trade.json
│       ├── lineage_civic.json
│       ├── lineage_arcane.json
│       ├── lineage_industrial.json
│       ├── lineage_digital.json
│       └── storage_and_wonders.json
├── assets/
│   ├── fonts/
│   │   ├── JetBrainsMono-Regular.ttf
│   │   ├── JetBrainsMono-Bold.ttf
│   │   └── terminus.ttf              # Fallback bitmap font
│   ├── shaders/
│   │   ├── crt_terminal.gdshader     # Scanlines + vignette
│   │   ├── phosphor_glow.gdshader    # Text bloom
│   │   └── age_transition.gdshader   # Wipe effect for age advance
│   ├── sprites/
│   │   ├── buildings/                # Per-building sprite sheets
│   │   │   ├── gathering_camp.png
│   │   │   ├── lumber_mill.png
│   │   │   └── ...                   # One per building (284 total)
│   │   ├── particles/
│   │   │   ├── smoke.png
│   │   │   ├── sparkle.png
│   │   │   └── dust.png
│   │   └── backgrounds/
│   │       └── age_01_primitive.png  # Through age_22_galactic.png
│   ├── audio/
│   │   ├── ambient/
│   │   │   ├── age_01_forest.ogg
│   │   │   └── ...                   # One ambient loop per age
│   │   ├── sfx/
│   │   │   ├── keypress.ogg
│   │   │   ├── keypress_return.ogg
│   │   │   ├── error_beep.ogg
│   │   │   ├── build_complete.ogg
│   │   │   ├── age_advance.ogg
│   │   │   ├── catastrophe_alarm.ogg
│   │   │   └── wonder_complete.ogg
│   │   └── music/                    # Optional background music tracks
│   └── art/
│       ├── capsule_616x353.png       # Steam capsule art
│       ├── header_460x215.png        # Steam header
│       └── screenshots/
├── scenes/
│   ├── Main.tscn                    # Root scene, loads Boot then Game
│   ├── BootSplash.tscn              # "AGEFORGE v2.0" boot sequence
│   ├── GameScreen.tscn              # Main game (TerminalPanel + CityMap)
│   └── overlays/                    # One .tscn per overlay
│       ├── ResearchOverlay.tscn
│       ├── MilitaryOverlay.tscn
│       └── ...
└── export_presets.cfg               # Godot export configs for all platforms
```

---

## Core Systems Architecture

### GameEngine.cs — The Central Singleton

`GameEngine` is a Godot **Autoload** singleton, meaning it is instantiated at launch before any scene and persists for the entire session. It is the direct equivalent of `game/engine.go`.

```csharp
// src/core/Engine.cs
using Godot;
using System.Collections.Generic;

public partial class GameEngine : Node
{
    // Autoload singleton access
    public static GameEngine Instance { get; private set; }

    // Core state
    public GameState State { get; private set; }
    public EventBus Bus { get; private set; }

    // Subsystems
    private Ticker _ticker;
    private Buildings _buildings;
    private Workers _workers;
    private Research _research;
    private Trade _trade;
    private Diplomacy _diplomacy;
    private Prestige _prestige;
    private SaveSystem _saveSystem;

    // Speed control (equivalent to Go speedMultiplier)
    public float SpeedMultiplier { get; private set; } = 1.0f;
    public const float MinTickInterval = 0.2f; // 200ms

    public override void _Ready()
    {
        Instance = this;
        Bus = new EventBus();
        State = new GameState();
        _ticker = new Ticker(this);
        _buildings = new Buildings(State, Bus);
        _workers = new Workers(State, Bus);
        _research = new Research(State, Bus);
        // ... init other subsystems

        ConfigLoader.LoadAll();         // Load JSON data files
        _saveSystem = new SaveSystem(); // Try to load existing save
        if (!_saveSystem.TryLoad(ref State))
            State = GameState.NewGame();

        AddChild(_ticker);              // Ticker runs as a child node
    }

    public void SetSpeed(float multiplier)
    {
        SpeedMultiplier = Mathf.Max(0.1f, multiplier);
        _ticker.UpdateInterval(MinTickInterval / SpeedMultiplier);
    }
}
```

**Key differences from Go engine.go:**
- No mutex/RWMutex — Godot runs on a single main thread; the tick fires from `_Process`, never concurrent with input
- No goroutines — the tick loop is `_Process` delta accumulation (see Ticker.cs)
- No `sync.RWMutex` — subsystems read State freely; only `DoTick()` writes to it
- Autoload replaces global `var ge *GameEngine` pattern

---

### Ticker.cs — Delta Accumulator

The Go version uses a goroutine with `time.Sleep`. Godot uses `_Process(double delta)` which is called every frame with the elapsed time since the last frame.

```csharp
// src/core/Ticker.cs
using Godot;

public partial class Ticker : Node
{
    private double _accumulated = 0.0;
    private double _interval = 0.2; // 200ms default
    private GameEngine _engine;

    public Ticker(GameEngine engine) { _engine = engine; }

    public void UpdateInterval(double interval)
    {
        _interval = Math.Max(0.05, interval); // Floor at 50ms (20x speed cap)
    }

    public override void _Process(double delta)
    {
        _accumulated += delta;
        while (_accumulated >= _interval)
        {
            _accumulated -= _interval;
            _engine.DoTick();
        }
    }
}
```

**Why `while` not `if`:** If the game lags and two intervals pass in one frame, the `while` catches up with two ticks rather than skipping one. This matches the Go behavior where the ticker fires independently of render frames.

**Autosave:** `DoTick()` increments a counter; every 300 ticks (~60 seconds at 1x speed) it calls `SaveSystem.Save()` after the tick completes. No need for a separate goroutine.

---

### EventBus.cs — Godot Signals

The Go event bus uses `sync.RWMutex` and function callbacks. Godot has a built-in signal system that is single-threaded, type-safe in C#, and integrated with the editor.

```csharp
// src/core/EventBus.cs
using Godot;

public partial class EventBus : RefCounted
{
    // Declare signals using C# delegates
    [Signal] public delegate void ResourceChangedEventHandler(string key, double amount);
    [Signal] public delegate void BuildingBuiltEventHandler(string buildingKey, int count);
    [Signal] public delegate void BuildingDestroyedEventHandler(string buildingKey, int count);
    [Signal] public delegate void TechResearchedEventHandler(string techKey);
    [Signal] public delegate void AgeAdvancedEventHandler(string newAgeKey);
    [Signal] public delegate void WorkerAssignedEventHandler(string buildingKey, int count);
    [Signal] public delegate void CatastropheTriggeredEventHandler(string type);
    [Signal] public delegate void EpochChangedEventHandler(int epoch);
    [Signal] public delegate void MilestoneCompletedEventHandler(string milestoneKey);
    [Signal] public delegate void PrestigeEventHandler();
    [Signal] public delegate void LogMessageEventHandler(string message, string category);

    // Typed emit methods (called by engine subsystems)
    public void EmitBuildingBuilt(string key, int count) =>
        EmitSignal(SignalName.BuildingBuilt, key, count);

    public void EmitAgeAdvanced(string newAge) =>
        EmitSignal(SignalName.AgeAdvanced, newAge);

    public void EmitLog(string msg, string category = "info") =>
        EmitSignal(SignalName.LogMessage, msg, category);
}
```

**Connection pattern (in UI nodes):**
```csharp
// In TerminalPanel._Ready():
GameEngine.Instance.Bus.AgeAdvanced += OnAgeAdvanced;
GameEngine.Instance.Bus.LogMessage += outputLog.AppendLine;
```

**Critical difference from Go:** The Go bus runs handlers under the engine write lock, so calling `GetState()` inside a handler causes deadlock. In Godot, all signals are dispatched synchronously on the main thread. There is no lock to deadlock with. Handlers can freely read `GameEngine.Instance.State`.

---

### GameState.cs — Pure Data

`GameState` is a plain C# class (not a Godot `Node` or `Resource`). It holds all mutable game state and is designed to be JSON-serializable for saves.

```csharp
// src/core/GameState.cs
using System.Collections.Generic;
using System.Text.Json.Serialization;

public class GameState
{
    public string CurrentAgeKey { get; set; } = "primitive_age";
    public int CurrentEpoch { get; set; } = 1;
    public double Ticks { get; set; } = 0;
    public double TotalPlaytime { get; set; } = 0;
    public int PrestigeCount { get; set; } = 0;
    public double PrestigeLegacyBonus { get; set; } = 1.0;

    // Resources: key → amount (e.g., "food" → 142.5)
    public Dictionary<string, double> Resources { get; set; } = new();

    // Buildings: key → BuildingState
    public Dictionary<string, BuildingState> Buildings { get; set; } = new();

    // Workers: domain key → WorkerDomainState
    [JsonPropertyName("workers")]          // preserves save compat with Classic "villagers" key
    public Dictionary<string, WorkerDomainState> Workers { get; set; } = new();

    // Research
    public HashSet<string> ResearchedTechs { get; set; } = new();
    public string? CurrentResearch { get; set; } = null;
    public double ResearchProgress { get; set; } = 0;

    // Milestones
    public HashSet<string> CompletedMilestones { get; set; } = new();
    public Dictionary<string, double> MilestoneProgress { get; set; } = new();

    // Military
    public Dictionary<string, int> ArmyUnits { get; set; } = new();
    public int Defense { get; set; } = 0;

    // Trade
    public HashSet<string> ActiveTradeRoutes { get; set; } = new();

    // Diplomacy
    public Dictionary<string, double> FactionRelations { get; set; } = new();

    // Epoch / catastrophe
    public bool CatastrophePending { get; set; } = false;
    public string? PendingCatastropheType { get; set; }

    // Dev mode
    [JsonIgnore]
    public bool DevMode { get; set; } = false;

    public static GameState NewGame()
    {
        var s = new GameState();
        s.Resources["food"] = 10;
        s.Resources["wood"] = 5;
        // ... initial resource seeding
        return s;
    }
}

public class BuildingState
{
    public int Count { get; set; } = 0;
    public int WorkersAssigned { get; set; } = 0;
    public int WorkerCapacity { get; set; } = 0;
    public bool UnderConstruction { get; set; } = false;
    public double ConstructionProgress { get; set; } = 0;
}

public class WorkerDomainState
{
    public int Total { get; set; } = 0;
    public int Unassigned { get; set; } = 0;
    public Dictionary<string, int> Assignments { get; set; } = new();
}
```

---

### ConfigLoader.cs — JSON Data Files

Go config was Go code (maps, struct literals). Beta uses JSON files in `data/`, loaded at startup via Godot's `FileAccess`.

```csharp
// src/config/ConfigLoader.cs
using Godot;
using System.Collections.Generic;
using System.Text.Json;

public static class ConfigLoader
{
    public static Dictionary<string, BuildingDef> Buildings { get; private set; } = new();
    public static Dictionary<string, AgeDef> Ages { get; private set; } = new();
    public static Dictionary<string, TechDef> Techs { get; private set; } = new();
    // ... other defs

    public static void LoadAll()
    {
        Ages = LoadJson<Dictionary<string, AgeDef>>("res://data/ages.json");
        Techs = LoadJson<Dictionary<string, TechDef>>("res://data/techs.json");
        LoadBuildings(); // multiple files merged
        // ...
    }

    private static void LoadBuildings()
    {
        var files = new[] {
            "lineage_food", "lineage_wood", "lineage_stone", "lineage_iron",
            "lineage_military", "lineage_faith", "lineage_culture",
            "lineage_knowledge", "lineage_trade", "lineage_civic",
            "lineage_arcane", "lineage_industrial", "lineage_digital",
            "storage_and_wonders"
        };
        foreach (var file in files)
        {
            var batch = LoadJson<Dictionary<string, BuildingDef>>(
                $"res://data/buildings/{file}.json");
            foreach (var (k, v) in batch)
                Buildings[k] = v;
        }
    }

    private static T LoadJson<T>(string path)
    {
        using var f = FileAccess.Open(path, FileAccess.ModeFlags.Read);
        if (f == null) throw new System.Exception($"Config not found: {path}");
        var json = f.GetAsText();
        return JsonSerializer.Deserialize<T>(json,
            new JsonSerializerOptions { PropertyNameCaseInsensitive = true })!;
    }
}
```

---

## UI Architecture

### TerminalPanel — The Primary Interface

`TerminalPanel.tscn` is a Godot `PanelContainer` that occupies the left 60% of the game screen. It contains:

```
TerminalPanel (PanelContainer)
├── VBoxContainer
│   ├── OutputLog (RichTextLabel)     ← scrollback history
│   └── InputRow (HBoxContainer)
│       ├── PromptLabel (Label)       ← "> " in terminal color
│       ├── CommandInput (LineEdit)   ← player types here
│       └── GhostText (Label)        ← autocomplete ghost (gray)
└── ShaderMaterial → crt_terminal.gdshader
```

`RichTextLabel` supports BBCode natively, which maps directly to how Classic renders colored output:

```csharp
// OutputLog.cs
public void AppendLine(string message, string category = "info")
{
    string color = category switch {
        "error"   => "#ff5555",
        "success" => "#50fa7b",
        "warn"    => "#ffb86c",
        "age"     => "#bd93f9",
        "system"  => "#8be9fd",
        _         => "#f8f8f2"    // default terminal white
    };
    _label.AppendText($"[color={color}]{message}[/color]\n");
    // Auto-scroll to bottom
    _label.ScrollToLine(_label.GetLineCount() - 1);
}
```

**Scrollback buffer:** `RichTextLabel` handles this natively up to a configurable line limit. Classic's tview scrollback is replicated without additional code.

**Scrollback cap:** Set `RichTextLabel.MaxLinesVisible` or trim old lines periodically to prevent memory growth on long sessions.

---

### CommandInput.cs

```csharp
// src/ui/terminal/CommandInput.cs
using Godot;
using System.Collections.Generic;

public partial class CommandInput : LineEdit
{
    private List<string> _history = new();
    private int _historyPos = -1;

    public override void _Ready()
    {
        TextSubmitted += OnSubmit;
        TextChanged += OnTextChanged;
        // Don't clear on submit — we do it manually after routing
        ClearButtonEnabled = false;
    }

    private void OnSubmit(string text)
    {
        text = text.Trim();
        if (string.IsNullOrEmpty(text)) return;

        _history.Insert(0, text);
        if (_history.Count > 200) _history.RemoveAt(200);
        _historyPos = -1;

        Text = "";
        CommandRouter.Instance.Route(text);
    }

    public override void _Input(InputEvent @event)
    {
        if (@event is not InputEventKey key || !key.Pressed) return;

        if (key.Keycode == Key.Up && _history.Count > 0)
        {
            _historyPos = Math.Min(_historyPos + 1, _history.Count - 1);
            Text = _history[_historyPos];
            CaretColumn = Text.Length;
            AcceptEvent();
        }
        else if (key.Keycode == Key.Down)
        {
            _historyPos = Math.Max(_historyPos - 1, -1);
            Text = _historyPos >= 0 ? _history[_historyPos] : "";
            CaretColumn = Text.Length;
            AcceptEvent();
        }
        else if (key.Keycode == Key.Tab)
        {
            AutoComplete.Instance.AcceptSuggestion(this);
            AcceptEvent();
        }
    }

    private void OnTextChanged(string newText)
    {
        AutoComplete.Instance.Update(newText);
    }
}
```

---

### Overlay System

Each overlay is a `CanvasLayer` (renders above the game world) containing a `PanelContainer` styled with the terminal theme. They open/close via `OverlayManager`.

```csharp
// src/ui/overlays/OverlayManager.cs
public partial class OverlayManager : Node
{
    public static OverlayManager Instance { get; private set; }

    private Dictionary<string, Control> _overlays = new();
    private string? _currentOverlay = null;

    public void Show(string name)
    {
        if (_currentOverlay != null)
            _overlays[_currentOverlay].Visible = false;

        _overlays[name].Visible = true;
        _currentOverlay = name;
        _overlays[name].GrabFocus();
    }

    public void Hide()
    {
        if (_currentOverlay == null) return;
        _overlays[_currentOverlay].Visible = false;
        _currentOverlay = null;
        // Return focus to terminal input
        CommandInput.Instance.GrabFocus();
    }
}
```

Overlay panels use `RichTextLabel` with BBCode to render tables, just like Classic renders tview tables. The box-drawing characters (─, │, ┌, ┐, etc.) are Unicode and render perfectly in a monospace font. Existing Classic rendering code for overlays can be ported almost line-for-line with the BBCode color wrappers replaced by `[color=]` tags.

---

### City Map (CityMapRenderer)

The city map is a `SubViewport` embedded in the right 40% of the screen. This isolates its rendering from the terminal panel.

```
GameScreen (Control)
├── HSplitContainer
│   ├── TerminalPanel (left 60%)
│   └── SubViewportContainer (right 40%)
│       └── SubViewport
│           ├── TileMapLayer (road network)
│           ├── BuildingContainer (Node2D) ← building sprites added here
│           ├── ParticleContainer (Node2D) ← smoke, sparks, etc.
│           └── Camera2D (auto-zoom)
```

Building sprites are added/removed from `BuildingContainer` when the engine emits `BuildingBuilt`/`BuildingDestroyed` signals:

```csharp
// src/ui/map/CityMapRenderer.cs
public partial class CityMapRenderer : Node2D
{
    private Dictionary<string, BuildingSprite> _placed = new();

    public override void _Ready()
    {
        GameEngine.Instance.Bus.BuildingBuilt += OnBuildingBuilt;
        GameEngine.Instance.Bus.BuildingDestroyed += OnBuildingDestroyed;
        GameEngine.Instance.Bus.AgeAdvanced += OnAgeAdvanced;
    }

    private void OnBuildingBuilt(string key, int count)
    {
        if (_placed.ContainsKey(key)) return; // already on map

        var def = ConfigLoader.Buildings[key];
        var sprite = BuildingSprite.Instantiate(def);
        sprite.Position = RoadNetwork.GetSlotPosition(key);
        AddChild(sprite);
        _placed[key] = sprite;
        sprite.PlayConstructionAnimation();
    }
}
```

---

### Save System

```csharp
// src/core/SaveSystem.cs
using Godot;
using System.Text.Json;

public class SaveSystem
{
    private const string SavePath = "user://save_slot_{0}.json";
    private int _currentSlot = 0;

    public bool TryLoad(ref GameState state)
    {
        var path = string.Format(SavePath, _currentSlot);
        if (!FileAccess.FileExists(path)) return false;

        using var f = FileAccess.Open(path, FileAccess.ModeFlags.Read);
        var json = f.GetAsText();
        state = JsonSerializer.Deserialize<GameState>(json) ?? new GameState();
        return true;
    }

    public void Save(GameState state)
    {
        var json = JsonSerializer.Serialize(state,
            new JsonSerializerOptions { WriteIndented = false });

        using var f = FileAccess.Open(
            string.Format(SavePath, _currentSlot),
            FileAccess.ModeFlags.Write);
        f.StoreString(json);

        // Steam cloud save (if available)
        if (Steam.IsLoggedOn())
        {
            Steam.FileWrite($"save_slot_{_currentSlot}.json", json.ToUtf8Buffer());
        }
    }
}
```

`user://` maps to:
- Windows: `%APPDATA%/Godot/app_userdata/project-ageforge/`
- macOS: `~/Library/Application Support/Godot/app_userdata/project-ageforge/`
- Linux: `~/.local/share/godot/app_userdata/project-ageforge/`

---

## Performance Considerations

**Tick rate vs. frame rate:** The game ticks at configurable intervals (200ms baseline), not per-frame. Even at 10x speed (20ms ticks), the game is updating state 50 times per second. Godot renders at monitor refresh rate (60/120Hz). The UI only needs to reflect state changes, not update every frame.

**RichTextLabel scrollback:** Cap at 2000 lines. Beyond that, trim the top. This avoids memory growth over long sessions.

**City map sprite count:** 284 buildings maximum on map. Godot handles thousands of sprites at 60fps comfortably. No performance concern.

**Config loading:** All JSON loads on startup. After that, all config lookups are dictionary reads — O(1), no I/O. Same as Classic.

**Signal overhead:** Godot signals are fast for the volume AgeForge generates. Even in a catastrophic tick where 30 signals fire at once, the overhead is microseconds.

---

## Threading Model

Godot runs its main loop on a single thread. The game engine's `DoTick()` is called from `_Process` (main thread). All signal handlers fire on the main thread. No mutexes are needed.

If save serialization becomes a bottleneck (unlikely for AgeForge's state size), it can be moved to a `Thread` node:

```csharp
// Async save pattern (only needed if save takes >16ms, which it won't)
var thread = new Thread();
thread.Start(Callable.From(() => {
    var json = JsonSerializer.Serialize(State);
    // write to file
}));
```

For AgeForge's `GameState` size (~50KB JSON), synchronous save is fine.
