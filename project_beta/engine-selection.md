# Engine Selection — AgeForge Beta

## Evaluation Criteria

| Criterion | Weight | Notes |
|-----------|--------|-------|
| 2D rendering quality | High | CRT shaders, RichText, sprite animation |
| Steam integration | High | Must work day-one without hacks |
| Go developer ramp-up | Medium | C# and Go are closer than any other pair |
| Binary size / performance | Medium | Terminal game, not AAA |
| Cost / royalties | High | Indie margins matter at $4.99 |
| Long-term stability | Medium | Will still be maintained in 5 years |
| Scene editor / tooling | Medium | Reduces manual layout code |

---

## Option A: Godot 4 + C# (RECOMMENDED)

### Overview
Godot 4 is a fully open-source game engine (MIT license) with a built-in 2D renderer, animation system, audio engine, shader pipeline, and UI framework. C# support in Godot 4 uses .NET 6/8 via the Mono runtime, with first-class IDE support from JetBrains Rider and VS Code.

### Pros

**No royalties, ever.** Godot is MIT licensed. There is no revenue threshold, no per-seat fee, no account requirement. The engine source code is on GitHub. If Godot's maintainers dissolve tomorrow, the code still works.

**C# maps cleanly onto Go.** Both languages are statically typed, garbage-collected, struct/class-driven, and interface-based. A Go developer reading C# for the first time will recognize 80% of the idioms within a day. The mechanical port of `GameState`, `Buildings`, `Workers`, and `Resources` from Go structs to C# classes is tedious but not difficult.

**GodotSteam is mature.** The GodotSteam plugin (https://godotsteam.com) is a well-maintained open-source plugin wrapping the full Steamworks SDK. It supports achievements, cloud saves, leaderboards, overlay, input, and DLC. As of Godot 4.2+, it ships prebuilt binaries. Integration is adding the plugin and calling `Steam.Init()` in an autoload.

**Excellent 2D pipeline.** Godot's 2D renderer uses a canvas system with proper z-ordering, SubViewports, Camera2D with zoom/pan, particle systems, and a full shader language (GLSL-compatible). Everything the city map needs is native.

**Shader support for the terminal aesthetic.** Godot's `CanvasShader` lets you attach GLSL-compatible fragment shaders to any `Control` or `Sprite2D`. The CRT scanline + phosphor glow effect is ~30 lines of shader code. This cannot be done in a terminal at all.

**Small binary.** A Godot 4 + C# export for a 2D game is approximately 45–60MB on Windows (including .NET runtime). For Linux it's smaller. This is acceptable for Steam.

**Active community and documentation.** Godot has first-rate official documentation, an active Discord, and a large Reddit community. The "Godot 4 + C#" combination has significantly more resources than it did two years ago.

### Cons

**C# in Godot is still maturing.** GDScript gets features first. Some editor integrations (like live script reloading) work better with GDScript than C#. Hot-reload of C# scripts requires specific project settings and doesn't always work cleanly. This is a workflow annoyance, not a blocker.

**Smaller C# community than GDScript.** Most Godot tutorials use GDScript. You'll often need to mentally translate example code. The official C# API docs are complete but sparse on examples.

**GDNative deprecation.** GDNative (the old C++ extension system) is replaced by GDExtension in Godot 4. If you ever want to call into a C library directly, the path is GDExtension, not GDNative. Not relevant for this project but worth knowing.

### Steam Integration Detail

```
1. Download Steamworks SDK from https://partner.steamgames.com/
2. Download GodotSteam prebuilt for Godot 4.x + .NET from https://godotsteam.com
3. Replace Godot editor binary with GodotSteam build (or use as a GDExtension)
4. Add steam_appid.txt to project root with your AppID
5. In an Autoload singleton:
```

```csharp
// SteamManager.cs (Autoload)
public partial class SteamManager : Node
{
    public override void _Ready()
    {
        var result = Steam.Init(YOUR_APP_ID);
        if (result["status"] as int? != 1)
            GD.PrintErr("Steam failed to initialize");
    }

    public void UnlockAchievement(string apiName)
    {
        Steam.SetAchievement(apiName);
        Steam.StoreStats();
    }
}
```

### Verdict
Godot 4 + C# is the right choice. The only real cost is the C# Godot maturity gap, and that gap is acceptable for a game that doesn't need cutting-edge engine features.

---

## Option B: Godot 4 + GDScript

### Overview
GDScript is Godot's native scripting language — Python-like syntax, dynamically typed (with optional type hints), interpreted at runtime.

### Pros
- Fastest iteration speed in Godot (hot-reload works perfectly)
- Largest Godot tutorial base uses GDScript
- No .NET runtime dependency — smaller binary, simpler export
- Better editor integration (syntax highlighting, debugger, profiler all native)

### Cons

**Dynamic typing is a regression for this codebase.** AgeForge's game logic is a dense web of `map[string]float64`, `BuildingDef` structs, and interface-based event handlers. In GDScript, this becomes `Dictionary` with untyped values and duck-typed function calls. Bugs that the Go compiler catches at compile time become runtime errors. For a game with 284 buildings and 52 techs loaded from config, that's a real liability.

**A Go developer will find it jarring.** GDScript syntax reads like Python. Go developers are accustomed to explicit types, explicit interfaces, and no implicit `self`. The context switch adds weeks of productivity loss.

**Performance ceiling is lower.** GDScript is interpreted. For a simple idle game this doesn't matter much, but AgeForge's tick system touches every building's production calculation every 200ms. With 284 buildings that's a lot of dictionary lookups. C# compiles to native IL and is 5-10x faster for this kind of work.

### Verdict
Good choice if you were building a small game from scratch with no prior codebase. Wrong choice for porting 8,000 lines of Go simulation logic.

---

## Option C: Unity + C#

### Overview
Unity is the dominant game engine for indie 2D and 3D games, with a massive community, extensive asset store, and C# as its scripting language.

### Pros
- Largest game development community on Earth
- Asset Store has essentially everything (fonts, shaders, audio middleware)
- C# is familiar — same language as Godot option A
- Unity UI Toolkit (UIElements) is mature for custom UI
- Cinemachine, Timeline, and other Unity packages are polished

### Cons

**The runtime fee controversy is unresolved.** In 2023 Unity announced (then partially walked back) a retroactive per-install fee. The terms can still change. For a small indie game at $4.99, the risk that Unity changes its terms again and affects your margin is non-zero. This is a real business risk, not a theoretical one.

**Unity is heavy for this scope.** Unity's overhead for a 2D game is substantial. The default blank project compiles to ~200MB+ before you add any content. The editor is slower than Godot's on comparable hardware. For a game that's 95% text rendering and 5% sprite animation, this overhead is wasteful.

**No free Steam integration.** Unity does not ship with Steamworks integration. You use Steamworks.NET (free, community-maintained) or Facepunch.Steamworks (also free). These work but add an extra dependency and integration surface.

**Requires a Unity account.** Unity requires account creation and license management even for the free Personal tier. This is a minor friction but emblematic of the trust problem.

**Overkill.** Unity's 3D rendering pipeline, physics engine, and animation rigging system are irrelevant for AgeForge Beta. You're paying the complexity cost without getting the capability benefit.

### Verdict
A valid technical choice but a poor business choice. The runtime fee controversy alone is disqualifying for a long-lived indie game. Godot provides 95% of the capability at 10% of the complexity.

---

## Option D: Unreal Engine 5

### Overview
Unreal Engine is Epic Games' flagship engine, renowned for photorealistic 3D rendering, used in AAA titles and increasingly in indie 3D games.

### Pros
- Unmatched visual fidelity (irrelevant for AgeForge)
- Nanite and Lumen for next-gen rendering (irrelevant for AgeForge)
- Blueprints visual scripting (C++ alternative, but not great for porting Go logic)
- Free until $1M revenue

### Cons

**Completely wrong tool.** Unreal is designed for 3D games with real-time global illumination, complex physics, and cinematic rendering. AgeForge Beta is a 2D game with a terminal interface and a sprite city map. Using Unreal for this is comically disproportionate.

**C++ is a step up in complexity from Go.** Go developers transitioning to C++ face manual memory management, header/implementation file splits, complex build systems (CMake/Premake), and template metaprogramming. The learning curve is months, not weeks.

**5% royalty over $1M.** Unlikely to matter for AgeForge Beta's launch trajectory, but the royalty model is philosophically misaligned with the indie ethos of the project.

**Binary bloat.** An Unreal project's base export is measured in gigabytes. A 2GB install for a terminal-aesthetic idle game would be a meme.

**No native terminal rendering.** Unreal's UMG (Unreal Motion Graphics) UI system is designed for HUD elements in 3D games, not terminal-style text interfaces. Implementing a convincing CRT terminal in UMG would require significant custom work that Godot handles trivially.

### Verdict
Not suitable for this project in any dimension. Eliminated.

---

## Option E: Keep Go + Custom Renderer (Ebiten / SDL2)

### Overview
Ebiten is a 2D game library for Go (https://ebitengine.org). SDL2 bindings (via cgo) provide lower-level rendering. The idea: keep all game logic in Go, add a rendering layer.

### Pros

**Reuse all existing Go code.** The `game/` package, all of `config/`, the event bus, the tick system — none of it changes. This is genuinely appealing.

**Go is the known quantity.** No new language, no new engine, no ramp-up time.

**Ebiten is capable.** Ebiten has shipped real games (including one that reached Steam). It handles 2D sprites, text rendering, shaders (via Kage, Ebiten's shader language), audio (via Oto), and input.

**Simpler distribution.** `go build` produces a single binary. No engine runtime, no .NET.

### Cons

**No Steamworks SDK.** This is the hard blocker. Steam integration from Go requires cgo bindings to the Steamworks C++ SDK. These exist (e.g., `pkg/steam` wrappers) but are unmaintained community projects. Getting achievements, cloud saves, and the Steam overlay working reliably across Windows/Mac/Linux with cgo is a significant engineering effort with ongoing maintenance risk. Godot has GodotSteam, which is actively maintained by people who care about exactly this problem.

**No scene editor.** Every UI layout is code. Every overlay panel, every button position, every font size — all hand-coded. Godot's scene editor lets you drag and resize and see changes immediately. The difference in iteration speed for UI work is enormous.

**No built-in audio engine worth using.** Ebiten uses Oto for audio, which is low-level PCM streaming. There is no concept of audio buses, spatial audio, or easy looping with crossfade. Implementing "ambient audio that crossfades between ages" in Oto requires hundreds of lines of audio pipeline code. Godot's `AudioStreamPlayer` does this in 5 lines.

**No animation system.** Sprite animations in Ebiten require manual frame management. Godot's `AnimationPlayer` node handles complex multi-track animations, including shader parameter animation, with zero code.

**The rendering ceiling is lower.** Kage (Ebiten's shader language) is capable but not as mature as Godot's GLSL-compatible shader pipeline. CRT effects are achievable but require more work.

**Ebiten's community is small.** Compared to Godot or Unity, Ebiten's community is a fraction the size. Finding solutions to obscure problems takes longer.

### The Honest Assessment
This option is attractive because it minimizes rewriting. But the Go game logic rewrite to C# is estimated at 3 weeks of mechanical work. The Ebiten route saves those 3 weeks but costs them back in Steam integration alone — plus ongoing costs in audio, UI layout, animation, and editor tooling. The net is negative.

### Verdict
Viable for a proof-of-concept or a game that explicitly doesn't need Steam. Not the right choice when Steam integration is a primary goal.

---

## Option F: Hybrid — Go Backend + Godot Frontend

### Overview
The Go engine (game logic, tick system, GameState) runs as a local subprocess or goroutine library. Godot connects to it via WebSocket, named pipe, or shared memory, receiving state updates and sending command strings.

### Pros

**Maximum logic reuse.** The entire `game/` package stays in Go, unchanged. Godot only handles rendering and input.

**Clean separation of concerns.** Game logic is pure Go, tested with `go test`. UI is pure Godot. Each can be developed independently.

**Godot gets Steam.** The Godot process is the "game" from Steam's perspective. GodotSteam works normally.

### Cons

**Two codebases.** Every feature addition touches both Go and Godot. A new command needs: Go handler, Go test, IPC protocol message definition, Godot command router, Godot UI update. That's 5 touch points per feature instead of 2.

**IPC protocol design is non-trivial.** What format does `GameState` serialize to? JSON is safe but slow for 200ms ticks. Protobuf is fast but adds a build dependency. How do you handle the tick state pushing 25 resource values 5 times per second over a socket? What happens if the Go process crashes? How do you restart it without losing the Godot session?

**Steam packaging is awkward.** Steam expects one executable. Shipping a game that secretly launches a Go subprocess requires either a launcher wrapper or embedding the Go binary as a resource, both of which add complexity to the build pipeline.

**Debugging is harder.** An error in resource accumulation could be in Go or in the serialization/deserialization layer or in Godot's state representation. Crossing a process boundary doubles the debugging surface.

**The game logic port is not the hard work.** The honest accounting: porting Go game logic to C# is ~3 weeks of mechanical, low-risk work. Designing, implementing, testing, and maintaining an IPC protocol is ongoing, high-risk work. The hybrid saves 3 weeks and costs significantly more over the project lifetime.

### Verdict
Architecturally interesting but practically painful. The one scenario where it makes sense is if the Go game logic were extremely complex to port (e.g., uses CGo, has complex concurrency patterns, uses Go-specific libraries). AgeForge's game logic is pure data transformation with no external dependencies. The port is straightforward.

---

## Final Comparison Table

| Criterion | Godot 4 C# | Godot 4 GDScript | Unity C# | Unreal | Go + Ebiten | Hybrid |
|-----------|-----------|-----------------|----------|--------|------------|--------|
| Steam integration | Excellent (GodotSteam) | Excellent | Good (3rd party) | Good | Poor (unmaintained) | Good |
| Terminal aesthetic | Excellent | Excellent | Good | Poor | Good | Excellent |
| Go developer ramp-up | Fast (2 weeks) | Medium (3 weeks) | Fast (2 weeks) | Slow (months) | None | None |
| Binary size | ~55MB | ~30MB | ~200MB | ~2GB | ~20MB | ~75MB |
| Royalties | None | None | None* | 5% >$1M | None | None |
| Audio engine | Excellent | Excellent | Excellent | Excellent | Poor | Excellent |
| Scene editor | Yes | Yes | Yes | Yes | No | Yes |
| Long-term trust | High (MIT) | High (MIT) | Medium | Medium | High | High |
| Sprite animation | Excellent | Excellent | Excellent | Excellent | Manual | Excellent |
| Shaders (CRT) | Excellent | Excellent | Good | Good | Good | Excellent |
| **Overall** | **Best** | Good | Good | Poor | Fair | Fair |

*Unity's terms have changed before and may change again.

---

## Recommendation

**Godot 4 with C# (.NET 8) is the correct engine for AgeForge Beta.**

The decision is driven by: zero royalties, mature Steam integration via GodotSteam, C#'s close relationship to Go for the port work, an excellent 2D rendering pipeline for the city map and CRT effects, and strong long-term stability from an MIT-licensed open-source project. The only meaningful downside — C# in Godot being slightly less mature than GDScript — does not affect any feature AgeForge Beta needs.

Start here. Port methodically. The terminal aesthetic will be better in Godot than it ever was in a real terminal.
