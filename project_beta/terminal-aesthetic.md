# Terminal Aesthetic Design — AgeForge Beta

## Why the Terminal Aesthetic Is the Brand

AgeForge is not a city builder with a quirky UI. It is a terminal application where civilization happens to be the data you're managing. When you type `assign forge 5` and watch the output scroll in amber text, you are not playing a game that looks like a terminal. You are running a civilization on a computer that happens to be a game.

That distinction is the entire brand.

Steam has hundreds of city builders. Factorio, Rimworld, Dwarf Fortress, Frostpunk, Anno — players have infinite choices for "build a thing and watch it grow." AgeForge does not compete with those games. It competes with them for the same 15 minutes of attention, and it wins by offering something they cannot: the feeling of operational command. You are not clicking on sprites. You are issuing directives.

The terminal aesthetic is not a limitation inherited from the original implementation. It is the one irreplaceable element of AgeForge's identity. Everything in Beta's visual design exists to serve and amplify that identity.

### What "terminal aesthetic" means in practice

- **Monospace font, always.** Every character of game output uses the same width. Tables align because of font choice, not layout engines.
- **Color is semantic, not decorative.** Green means good. Red means bad. Amber is system output. Blue is your commands. Cyan is categories. Purple is ages. This is a convention from terminal software, and players learn it within minutes.
- **Text is the content.** There can be icons and progress bars but they should mirror using a terminal asthetic for the most part. A building's production is written as `0.50/tick`. A worker's status is `5/10 assigned`. The numbers are the interface.
- **Commands are the input method.** Always. Even in Beta with a real game engine, the game is controlled by typing. The mouse exists for scrolling and clicking buttons in overlays.
- **The prompt never disappears.** The `>` cursor is always present, always blinking. The game is always waiting for your next command.

---

## Implementing the Terminal Aesthetic in Godot

### Font Selection

The font is the most important visual decision. It needs to be:
1. Monospace (required for table alignment)
2. Legible at small sizes (players may run small windows)
3. Free to use in a commercial release (OFL or equivalent)
4. Character-complete for Unicode box-drawing (─ │ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼ and all block elements)

**Primary recommendation: JetBrains Mono**
- License: SIL Open Font License 1.1 (commercial use allowed)
- Excellent legibility at 13px+
- Complete Unicode coverage including box-drawing and block elements
- Distinguishes 0/O and l/I/1 clearly
- Download: https://www.jetbrains.com/legalnotice/

**Secondary recommendation: Iosevka**
- License: SIL Open Font License 1.1
- Narrower than JetBrains Mono — fits more text per line
- Many stylistic variants (Iosevka Term is optimized for terminal use)
- Download: https://typeof.net/Iosevka/

**Fallback: Terminus (bitmap)**
- Pure bitmap font, no anti-aliasing — the "real terminal" look
- Slightly harder to read at small sizes but extremely authentic
- Good for an optional "retro" theme setting

**Godot font setup:**
```
Project Settings → Theme → Default Font → JetBrainsMono-Regular.ttf
Font size: 14px (adjustable in settings)
Antialiasing: LCD subpixel (for screen readability at 14px)
```

For the terminal panel specifically, override the font in `TerminalPanel.tscn`:
```
RichTextLabel → Theme Overrides → Fonts → Normal Font → JetBrainsMono-Regular.ttf
RichTextLabel → Theme Overrides → Fonts → Bold Font → JetBrainsMono-Bold.ttf
```

---

### Color Themes

Offer three terminal color themes, selectable in Settings. All themes use the same semantic color mapping — only the palette values change.

**Theme: Amber (default)**
Modeled after the Amber phosphor displays of 1970s-80s terminals (VT100 era).

```
Background:     #0d0a00
Text (normal):  #ffb000
Text (dim):     #7a5400
Cursor:         #ffb000
Success/good:   #ffd050
Error/bad:      #ff4400
Warning:        #ff8800
System/info:    #cc8800
Age/special:    #ffffff
Selection bg:   #3a2800
```

**Theme: Phosphor Green**
The classic "green screen" — 1980s PC monitor aesthetic.

```
Background:     #000d00
Text (normal):  #00cc00
Text (dim):     #005000
Cursor:         #00ff00
Success/good:   #00ff44
Error/bad:      #ff2200
Warning:        #aaff00
System/info:    #009900
Age/special:    #aaffaa
Selection bg:   #003300
```

**Theme: Blue Steel**
Cold, modern, datacenter aesthetic. Fits the later industrial/digital ages.

```
Background:     #000814
Text (normal):  #8ecae6
Text (dim):     #264653
Cursor:         #aef0ff
Success/good:   #52b788
Error/bad:      #e63946
Warning:        #ffd166
System/info:    #4cc9f0
Age/special:    #e2cfff
Selection bg:   #1a3a4a
```

Store the active theme as a `Resource` and apply it globally via `Theme` in Godot's theme system:
```csharp
// In GameHUD._Ready():
ThemeManager.Apply(Settings.ColorTheme);
```

---

### CRT Shader

The CRT scanline and phosphor glow effect transforms a clean flat-colored `RichTextLabel` into something that feels like it's running on an actual CRT monitor. This is the single most important visual effect in Beta. and should an optional asthetic that can be toggled on and off.

Apply this shader as a `ShaderMaterial` on the `TerminalPanel`'s `PanelContainer`.

**`assets/shaders/crt_terminal.gdshader`:**

```glsl
shader_type canvas_item;

// Scanline strength: 0.0 = no scanlines, 1.0 = full black alternating
uniform float scanline_strength : hint_range(0.0, 1.0) = 0.25;

// Vignette strength: darkening at edges
uniform float vignette_strength : hint_range(0.0, 1.0) = 0.35;

// Curvature: barrel distortion to simulate curved CRT glass
uniform float curvature : hint_range(0.0, 0.1) = 0.02;

// Phosphor glow radius (screen-space blur approximation)
uniform float glow_strength : hint_range(0.0, 1.0) = 0.15;

// Color tint — set to theme's base text color
uniform vec3 phosphor_tint : source_color = vec3(1.0, 0.69, 0.0); // amber default

void fragment() {
    vec2 uv = UV;

    // --- Barrel distortion ---
    vec2 curved = uv - 0.5;
    float dist = dot(curved, curved);
    curved *= 1.0 + dist * curvature * 4.0;
    curved += 0.5;
    // Clip outside distorted area
    if (curved.x < 0.0 || curved.x > 1.0 || curved.y < 0.0 || curved.y > 1.0) {
        COLOR = vec4(0.0, 0.0, 0.0, 1.0);
        return;
    }

    // --- Sample base texture ---
    vec4 col = texture(TEXTURE, curved);

    // --- Scanlines ---
    // Every other pixel row is darkened
    float screen_y = curved.y * float(textureSize(TEXTURE, 0).y);
    float scanline = sin(screen_y * 3.14159) * 0.5 + 0.5;
    scanline = mix(1.0, scanline, scanline_strength);
    col.rgb *= scanline;

    // --- Phosphor glow (approximated as slight color bleed) ---
    // Sample neighbors and add a fraction of their color
    vec2 px = 1.0 / vec2(textureSize(TEXTURE, 0));
    vec4 glow = texture(TEXTURE, curved + vec2(px.x * 2.0, 0.0))
              + texture(TEXTURE, curved - vec2(px.x * 2.0, 0.0))
              + texture(TEXTURE, curved + vec2(0.0, px.y * 2.0))
              + texture(TEXTURE, curved - vec2(0.0, px.y * 2.0));
    glow *= 0.25 * glow_strength;
    col.rgb += glow.rgb;

    // --- Phosphor tint ---
    // Bias bright areas toward the phosphor color
    float brightness = dot(col.rgb, vec3(0.299, 0.587, 0.114));
    col.rgb = mix(col.rgb, col.rgb * phosphor_tint, brightness * 0.3);

    // --- Vignette ---
    vec2 vig = curved - 0.5;
    float vignette = 1.0 - dot(vig * 1.5, vig * 1.5);
    vignette = clamp(vignette, 0.0, 1.0);
    vignette = mix(1.0, pow(vignette, 0.5), vignette_strength);
    col.rgb *= vignette;

    COLOR = col;
}
```

**Shader parameters by theme:**

| Theme | `scanline_strength` | `vignette_strength` | `curvature` | `phosphor_tint` |
|-------|--------------------|--------------------|-------------|-----------------|
| Amber | 0.25 | 0.35 | 0.02 | (1.0, 0.69, 0.0) |
| Green | 0.20 | 0.30 | 0.02 | (0.0, 0.8, 0.0) |
| Blue Steel | 0.10 | 0.20 | 0.00 | (0.55, 0.79, 0.9) |

The Blue Steel theme has no curvature — it reads as a modern flat-panel monitor. Amber and Green are retrofuturist.

**Performance note:** This shader runs on the GPU. At 1080p, it costs approximately 0.3ms per frame on integrated graphics — negligible.

---

### Phosphor Glow on Text (Separate Shader)

The CRT shader adds a mild glow to the entire panel. For text specifically, a separate approach makes the text appear to emit light:

In Godot 4, `RichTextLabel` supports a `CanvasItem → Material → ShaderMaterial`. Apply a second shader there specifically for text glow:

```glsl
// assets/shaders/phosphor_glow.gdshader
shader_type canvas_item;
uniform float glow_radius : hint_range(1.0, 8.0) = 3.0;
uniform float glow_intensity : hint_range(0.0, 2.0) = 0.8;
uniform vec3 glow_color : source_color = vec3(1.0, 0.7, 0.0);

void fragment() {
    vec4 base = texture(TEXTURE, UV);
    vec2 px = 1.0 / vec2(textureSize(TEXTURE, 0));

    float glow = 0.0;
    for (float x = -glow_radius; x <= glow_radius; x += 1.0) {
        for (float y = -glow_radius; y <= glow_radius; y += 1.0) {
            glow += texture(TEXTURE, UV + vec2(x, y) * px).a;
        }
    }
    glow /= (glow_radius * 2.0 + 1.0) * (glow_radius * 2.0 + 1.0);
    glow *= glow_intensity;

    vec4 glow_col = vec4(glow_color * glow, glow * 0.8);
    COLOR = base + glow_col * (1.0 - base.a);
}
```

**Note:** Apply this shader only at the "Accessible" and "Vibrant" intensity settings. Provide a "Minimal" setting (no glow) for players who prefer flat rendering or have photosensitivity concerns.

---

### Cursor Blink

The blinking cursor is a subtle but powerful signal that the game is waiting. Godot's `LineEdit` has a built-in cursor blink rate. Override it to match the theme:

```csharp
// CommandInput._Ready()
CaretBlinkEnabled = true;
CaretBlinkInterval = 0.5f; // 500ms blink — classic terminal rate
// Style the caret to match the phosphor color
AddThemeColorOverride("caret_color", ThemeManager.CurrentTheme.CursorColor);
```

The `>` prompt label should also pulse subtly:

```csharp
// In TerminalPanel, animate prompt label with a Tween
var tween = CreateTween().SetLoops();
tween.TweenProperty(_promptLabel, "modulate:a", 0.5f, 0.8f);
tween.TweenProperty(_promptLabel, "modulate:a", 1.0f, 0.8f);
```

---

### Boot Sequence

On game start, before the game state loads, display a terminal boot sequence. This is pure theater — it takes 3-4 seconds and sets the mood perfectly. the version number would be the correct number for the current live version in steam. and the loading would be based on the current number of buildings, tech, milses and other things you have currently unlocked.

```
AGEFORGE SYSTEMS v2.0.0
Copyright (C) 2024-2026 — All rights reserved

Initializing core subsystems...
  [OK] Resource management module
  [OK] Population dynamics engine
  [OK] Event propagation bus
  [OK] Epoch scheduler
  [OK] Catastrophe prediction model

Loading civilization database...
  284 building definitions loaded
   52 technology records found
   33 milestone objectives indexed
   22 historical ages mapped

Connecting to save state...
  [OK] Slot 0: Civilization "Ember Reach" — Age: Iron Age — Ticks: 147,293

BOOT COMPLETE. Type 'help' for commands.
>
```

Each line appears with a slight delay (20-50ms) to simulate text streaming in. Use a `Timer` and an array of lines:

```csharp
// BootSplash.cs
private string[] _bootLines = { /* ... */ };
private int _lineIndex = 0;

public void StartBoot()
{
    var timer = new Timer();
    timer.WaitTime = 0.04f;
    timer.Timeout += PrintNextLine;
    AddChild(timer);
    timer.Start();
}

private void PrintNextLine()
{
    if (_lineIndex >= _bootLines.Length)
    {
        GetTree().ChangeSceneToFile("res://scenes/GameScreen.tscn");
        return;
    }
    _outputLabel.AppendText(_bootLines[_lineIndex++] + "\n");
}
```

Pressing any key after the first few lines skips to the game immediately (the same behavior as Classic's age splash keypress fix from the v3.2.0 release).

---

## The Duality: Terminal + Real Game

The screen is split 60/40:

```
┌────────────────────────────────────────────────────────────┐
│                                                            │
│  TERMINAL PANEL (60%)          │  CITY MAP (40%)          │
│                                │                          │
│  > build lumber_mill           │  [animated city          │
│  Building started: Lumber Mill │   sprites react to       │
│  [lumber_mill] 0/1 built       │   your commands]         │
│  > assign lumber_mill 3        │                          │
│  Assigned 3 woodcutters        │  Camera auto-zooms       │
│  Production: +1.5 wood/tick    │  as city grows           │
│  > _                           │                          │
│                                │  Age 3: Bronze Age       │
│                                │  Population: 47          │
└────────────────────────────────────────────────────────────┘
```

The city map is always visible. It is not a separate screen you navigate to. It is a live view of the civilization you are managing. When you type `build forge`, a forge appears on the map. When a catastrophe strikes, buildings crumble. When you advance an age, the entire city transitions — new building sprites, new background, new ambient audio.

**The map does not need player attention.** It is ambient information. Players who want to ignore it can. Players who love it can zoom in with the scroll wheel or the `zoom` command.

---

## Overlay Panels as Terminal Windows

When the player types `research` or opens the Research overlay, it appears as a floating terminal window over the city map. It uses box-drawing characters for its border — not Godot UI borders:

```
┌─────────────────── RESEARCH ────────────────────────────────┐
│ Current: Pottery [████████░░] 80% (12 ticks remaining)      │
│                                                              │
│ AVAILABLE                                                    │
│ ──────────────────────────────────────────────────────────   │
│  [x] Pottery          +20% food storage     Cost: 50 know   │
│  [ ] Bronze Working   Unlocks forge          Cost: 100 know  │
│  [ ] Masonry          +15% stone storage     Cost: 75 know   │
│  [ ] Writing          Unlocks scribes        Cost: 120 know  │
│                                                              │
│ Type 'research <tech>' to begin research.                    │
│ [Esc] close                                                  │
└──────────────────────────────────────────────────────────────┘
```

The border is rendered as a `Label` or via `RichTextLabel`'s `[font]` tag using JetBrains Mono. All characters are Unicode box-drawing — they look identical to what Classic renders in tview today.

This means overlay content code from Classic (which generates these strings) ports almost directly to Beta with minimal changes.

---

## Age Advancement Cinematic

When the player advances an age, the game shows a full-screen terminal cinematic:

1. The city map fades to black
2. A loading bar appears: `Advancing civilization... [████████████████] 100%`
3. Large ASCII art of the age name fades in with the phosphor glow shader active
4. A brief text description of the age scrolls in
5. Key unlocks are listed: new buildings, new resources, new commands
6. Ambient audio crossfades to the new age's theme
7. Press any key or wait 5 seconds to dismiss
8. City map fades back in with new building sprites and background

**Example ASCII art for Bronze Age:**

```
  ██████╗ ██████╗  ██████╗ ███╗   ██╗███████╗███████╗
  ██╔══██╗██╔══██╗██╔═══██╗████╗  ██║╚══███╔╝██╔════╝
  ██████╔╝██████╔╝██║   ██║██╔██╗ ██║  ███╔╝ █████╗
  ██╔══██╗██╔══██╗██║   ██║██║╚██╗██║ ███╔╝  ██╔══╝
  ██████╔╝██║  ██║╚██████╔╝██║ ╚████║███████╗███████╗
  ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝╚══════╝

       ██╗    ██╗ ██████╗ ██████╗ ██╗  ██╗███████╗
       ██║    ██║██╔═══██╗██╔══██╗██║ ██╔╝██╔════╝
       ██║ █╗ ██║██║   ██║██████╔╝█████╔╝ ███████╗
       ██║███╗██║██║   ██║██╔══██╗██╔═██╗ ╚════██║
       ╚███╔███╔╝╚██████╔╝██║  ██║██║  ██╗███████║
        ╚══╝╚══╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝

  ─────────────────────────────────────────────────────
  The age of metal dawns. Bronze unlocks new tools,
  new weapons, and new ambitions.
  ─────────────────────────────────────────────────────
  NEW: Forge, Bronze Mine, Armory, Barracks
  NEW: bronze, tin resources
  NEW: recruit warrior, deploy army commands
  ─────────────────────────────────────────────────────
  [Press any key to continue]
```

The glow shader on this screen can be turned up higher than normal — this is a moment, and it should feel like one.

---

## Sound Design

Sound design is the most underrated part of terminal aesthetics. The right audio makes a text interface feel physical and present.

### Keypress Sounds

Every keystroke triggers a brief mechanical keyboard click. Use a small set of 4-6 variations (slightly different pitch/timing) played randomly to avoid repetition. Source: free CC0 mechanical keyboard samples from Freesound.org.

Volume should be low (~40% of master) and adjustable in settings. Some players will turn this off. Most will leave it on after the first five minutes.

`Enter` gets a distinct sound — heavier click, slightly longer decay. This is the command being submitted.

```csharp
// CommandInput.cs — key sound integration
public override void _Input(InputEvent @event)
{
    if (@event is InputEventKey key && key.Pressed && !key.Echo)
    {
        if (key.Keycode == Key.Enter || key.Keycode == Key.KpEnter)
            AudioManager.PlaySfx("keypress_return");
        else
            AudioManager.PlaySfx("keypress", RandomPitch(0.95f, 1.05f));
    }
}
```

### Error Beep

Classic terminal behavior: invalid command → terminal beep. In Beta, an invalid command plays a brief low-frequency beep (like the PC speaker BIOS beep) AND prints the error in red. It should feel like the machine rejected your command.

### Building Completion

A satisfying "ding" — think old terminal notification bell. Bright, short, a hint of reverb. Plays when `UnderConstruction` transitions to `false`.

### Age Advancement Fanfare

A short (5-8 second) ambient piece that plays once during the age advancement cinematic. Not a full music track — more like a chime. Should be abstract and atmospheric. The progression from primitive (wooden drum hit + wind) through industrial (steam whistle + industrial hum) to galactic (crystalline synth chord) mirrors the game's arc.

### Ambient Audio

Each age has a looping ambient track that plays at low volume (~25% master) during normal play. These crossfade smoothly when the age advances.

**Age audio arc:**
| Ages | Sound | Character |
|------|-------|-----------|
| 1-3 (Primitive, Stone) | Forest, distant river, birds | Quiet, organic |
| 4-6 (Bronze, Iron, Classical) | Wind, distant smithing, crowd murmur | Growing activity |
| 7-9 (Medieval, Feudal, Renaissance) | Church bells, market sounds, workshop clatter | Layered, complex |
| 10-12 (Industrial, Steam, Victorian) | Steam hiss, factory rhythm, machinery | Mechanical, rhythmic |
| 13-15 (Modern, WW era, Atomic) | Electrical hum, radio static, machinery | Tense, electric |
| 16-18 (Space, Cyber, Digital) | Server room hum, digital beeps, cooling fans | Cold, precise |
| 19-22 (Post-human, Galactic) | Deep synthesizer drone, occasional harmonic tone | Vast, alien |

All audio files should be `.ogg` format for Godot's native streaming (smaller than `.wav`, no license issues like `.mp3`).

---

## Accessibility Settings

Provide these options in the Settings screen:

| Setting | Default | Notes |
|---------|---------|-------|
| Font size | 14px | Range: 11-20px |
| CRT shader intensity | Medium | Off / Low / Medium / High |
| Phosphor glow | On | Toggle |
| Scanline strength | 25% | 0-100% slider |
| Color theme | Amber | Amber / Green / Blue |
| Keypress sounds | On | Toggle |
| Ambient audio volume | 25% | 0-100% |
| Cursor blink rate | 500ms | Slow / Normal / Fast / Off |
| High contrast mode | Off | Removes glow, increases contrast ratio |
| Reduce motion | Off | Disables screen shake, transition animations |

The "Reduce motion" setting is important. AgeForge's catastrophes trigger screen shake. Players with vestibular disorders need to turn that off without losing the rest of the experience.
