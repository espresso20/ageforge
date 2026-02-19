# AgeForge - Project Conventions

## Overview
CLI idle/clicker empire builder game built with Go + tview/tcell.

## Architecture
- **config/** - Data definitions (resources, buildings, techs, ages). Pure data, no logic.
- **game/** - Game engine, managers, tick loop. No UI imports.
- **ui/** - tview-based TUI. Reads GameState snapshots, never touches engine internals.
- **main.go** - Entry point, wires engine + UI.

## Patterns
- **Event Bus**: Systems communicate via `game.EventBus` (pub/sub). Events: BuildingCompleted, ResearchDone, AgeAdvanced, etc.
- **Config-Driven**: All content defined as data in config/, referenced by string keys.
- **Manager Pattern**: Each system (resources, buildings, villagers, research, military) has a manager struct with clear API.
- **GameState Snapshot**: `engine.GetState()` returns a read-only snapshot for UI consumption.

## Conventions
- Package names: lowercase, single word (config, game, ui)
- Config keys: snake_case strings ("lumber_mill", "stone_age")
- Use `float64` for resource amounts, `int` for building counts
- Error handling: return errors up, log at boundaries
- No global state; pass dependencies explicitly

## Build & Run
```bash
go build -o ageforge .
./ageforge
# or
go run main.go
```

## Testing
```bash
go test ./...
```
