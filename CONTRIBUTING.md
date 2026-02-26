# Contributing to AgeForge

Thanks for your interest in contributing! AgeForge is a terminal-based idle empire builder written in Go.

## Getting Started

```bash
git clone https://github.com/espresso20/ageforge.git
cd ageforge
go build -o ageforge .
./ageforge
```

Run the tests:

```bash
go test ./...
```

## Project Structure

| Directory | Purpose |
|---|---|
| `config/` | Data definitions — resources, buildings, techs, ages, events. Pure data, no logic. |
| `game/` | Game engine, managers, tick loop. No UI imports. |
| `ui/` | tview-based TUI. Reads `GameState` snapshots only. |
| `main.go` | Entry point, wires engine + UI. |

## How to Contribute

1. **Fork** the repo and create a branch: `git checkout -b feat/your-feature`
2. **Make your changes** — keep them focused and minimal
3. **Run tests** — `go test ./...` must pass
4. **Open a PR** with a clear description of what you changed and why

## Guidelines

- Keep PRs small and focused. One thing per PR.
- Config-driven content (new buildings, techs, resources) lives in `config/` as data structs.
- Game logic goes in `game/`, UI rendering in `ui/`. Don't mix them.
- No global state — pass dependencies explicitly.
- If you're adding a new game system, add it to the event bus so other systems can react.

## Reporting Bugs

Open a [GitHub Issue](https://github.com/espresso20/ageforge/issues) with:
- What you were doing
- What you expected to happen
- What actually happened
- Your OS and terminal

## Questions

Join the [Discord](https://discord.gg/EPvyd5vjpj) for discussion.
