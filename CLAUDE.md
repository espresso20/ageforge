# AgeForge - Project Conventions

## Overview
CLI idle/clicker empire builder game built with Go + tview/tcell. 22 ages, 301 buildings (14 lineages), 94 techs, 77 milestones, 7 epochs, catastrophe system with harbingers and the Last Passage, 12 worker domains, 26 resources, trade, diplomacy, prestige. Pacing targets live in `config/pacing.go` (about a week to first prestige).

## Conventions
- Package names: lowercase, single word (config, game, ui)
- Config keys: snake_case strings ("lumber_mill", "stone_age")
- Use `float64` for resource amounts, `int` for building counts
- Error handling: return errors up, log at boundaries
- No global state; pass dependencies explicitly
- Bus handlers run under engine write lock — never call GetState() or other lock-acquiring methods inside them

## Work Tracking (Trello + GitHub)
- The backlog lives on the public Trello board: https://trello.com/b/tf31C2cz/ageforge (lists: Refactor, Bugs, Features, Balance, Later Enhancements, Doing, Done). There is no TODO.md / DONE.md.
- Read it without auth: `curl -sL https://trello.com/b/tf31C2cz.json` returns lists, cards, labels and checklists. Writing to the board needs Trello auth that isn't set up, so tell Adam which cards to move rather than assuming they moved.
- Reference cards by shortLink in PR titles, e.g. `feat(trade): ... (card iKgfDClK)`.
- Before starting a feature, check the Done list and recent merged PRs (`gh pr list --state merged`) so you don't redo shipped work.
- The current state lives in: open PRs (`gh pr list`), CHANGELOG.md `[Unreleased]`, and the nightly smoke report (see Testing). Record newly discovered follow-ups in the PR description or ask Adam to add a card; don't hold them only in context.

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
make smoke        # fast smoke tier (runs on every PR, pacing enforced)
make smoke-full   # full tier (nightly)
```

## Website & Wiki Sync Rule
**Any change to the game MUST also update the appropriate pages in `site/`.**
- Balance change (building cost, resource rate, tech effect, age requirement) → update `site/docs/` wiki page for that system
- New building/tech/age/resource/wonder/expedition → add it to the relevant wiki page AND the landing page if it affects headline stats
- Mechanic change (new command, removed command, renamed key) → update `site/docs/commands.md` and any affected wiki page
- The landing page `site/index.html` hero stats must stay accurate
- Use a subagent for the wiki update if it spans more than 2 files

## Subagent Usage — Token Efficiency (CRITICAL)
This project is large. Burning main context on file reads is wasteful. Follow these rules strictly:

**What goes in subagents (NOT main context):**
- All implementation work for any sub-phase: reading files + writing code + running tests
- Any task touching more than 2 files
- All wiki/docs writes
- Any bulk read of config or source files

**What stays in main context:**
- Architecture decisions and reasoning
- Writing self-contained subagent prompts
- Small targeted edits (1-2 lines in a file already read this session)
- Reviewing subagent summaries and deciding next steps

**Subagent rules:**
- Use `subagent_type=general-purpose` for read+write tasks (implementation, docs)
- Use `subagent_type=Explore` for read-only research across multiple files
- Every subagent prompt must be **fully self-contained**: list exact files to read, exact changes needed, verify with `go build ./...` and `go test ./...`, return a summary of changes made
- Spawn parallel subagents for independent work (e.g. Phase 11a + 11b simultaneously; Phase 12 docs in 3-4 parallel batches)

**Never read these files in main context** (too large — always delegate):
- `config/buildings_new*.go`, `config/ages.go`, `config/buildings.go`
- `game/engine.go`, `game/villagers.go`, `game/buildings.go`
- `ui/tab_economy.go`, `ui/dashboard.go`
- Any file over ~200 lines unless doing a single targeted Grep lookup

**Per-phase workflow:**
1. Write a self-contained subagent prompt with exact spec
2. Spawn subagent — it reads, implements, tests, reports back
3. Review summary in main context
4. Commit if clean; open a PR and note any Trello cards to move
