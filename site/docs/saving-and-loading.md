# Saving & Loading

AgeForge keeps your civilization safe with named saves, a continuous autosave, and a full save browser. This page covers where saves live, the commands that manage them, and how to read the Load Game browser.

---

## Starting a new game

When you choose **New Game**, you're prompted to name your civilization — pre-filled with a randomly generated name. Press **Enter** to accept it, type your own, or press **Tab** to roll a fresh suggestion. That name becomes your active save, and the game autosaves into it from there.

---

## Where saves live

Save files are written to **your active account's** saves folder — `data/accounts/<account_id>/saves/*.json`, relative to the directory you launch the game from. One file per save. The `save`/`load` commands and the **Load Game** browser all read and write the same files, so a save made one way shows up the other.

Saves are **per-account**: each account keeps its own `saves/` folder, so switching accounts changes which saves you see, and one account's saves never mix with another's. See [Account & Recovery](account.md) for how accounts and their slots are laid out.

> **Upgrading from an older version?** Earlier builds kept a single flat `data/saves/` folder. On first launch the game **migrates that layout automatically and non-destructively** — your existing saves move into your account's `data/accounts/<id>/saves/` slot, and **nothing is deleted**. You don't have to do anything.

---

## Commands

| Command | Description |
|---|---|
| `save` | Opens an **Overwrite / Branch** prompt for your current run (see below) |
| `save <name>` | **Branches** a new save with that name off your current run; autosave then follows it |
| `load` | Opens the **Load Game** browser (your lineage tree) to pick which save/branch to load |
| `load <name>` | Loads that named save directly |
| `saves` (or `save list`) | List all save files |
| `Esc` | Quick-save to your current (active) save |

```
save           # prompt: Overwrite this run, or Branch a new save?
save hero      # branch a new save named "hero" off the current run
load           # open the Load Game browser to pick a save/branch
load hero      # load the save named "hero" directly
saves
```

A bare `load` (no name) **opens the Load Game browser** — the lineage tree of every save — so you choose which save/branch to load instead of having a slot picked for you. You can open it mid-game; pressing `Esc` in the browser returns you to your current run without loading anything. `load <name>` skips the browser and loads that save directly.

---

## Active save slot

The game remembers the last slot you explicitly saved to or `load`ed — your **active** save.

- Until you've named a save or loaded one this session, the active slot defaults to `autosave`.
- Once you branch to `hero` or `load hero`, the active slot is `hero` and autosave follows it.

The periodic autosave and `Esc` quick-save both write to your **active** save, continuously overwriting it. Your current game *is* the autosave — there's no separate slot quietly shadowing it.

---

## Branching your save

A bare `save` opens a prompt with two choices:

- **Overwrite** — write your current run to the active slot now (the same thing autosave and `Esc` do, on demand).
- **Branch new** — fork a brand-new save. You're given a generated name (editable; **Tab** rolls a fresh one, **Enter** confirms). The new save's *parent* is the save you branched from, and **autosave switches to follow the new branch** — so the old save is left frozen exactly at the branch point.

`save <name>` skips the prompt and branches straight to that name.

This is the way to preserve a moment without stopping play: branch before a prestige, a risky catastrophe, or any decision you might want to revisit. Your old save stays as it was; you keep playing on the new branch. Branched saves are ordinary files — they appear in the `saves` list and the Load Game browser alongside everything else. (The save names must be valid and unique; branching to a name that's already taken is rejected.)

---

## Autosave

The game autosaves periodically and whenever you press `Esc`, writing to your **active** save — so your current game is always kept up to date on disk. To preserve a specific point you don't want overwritten, save it under a new name (or duplicate it with `c` in the Load Game browser).

---

## The Load Game browser

Choosing **Load Game** from the main menu — or typing a bare `load` mid-game — opens a save browser that lists every save belonging to your **active account** (under `data/accounts/<id>/saves/`). Highlighting a save updates a **detail pane** showing everything you need to size up that save before loading it: its age and epoch, population, buildings, wonders, milestones, techs, soldiers, prestige, [morale](morale.md), its account attribution (*this account* / *another account* / *pre-account*), a ⚠ warning if a catastrophe is pending, the exact save time, and — for branched saves — a **Branched from** line naming the save it forked off. Opened mid-game, `Esc` returns you to your current run without loading anything.

**The lineage tree.** Saves aren't shown as a flat list — they're arranged as a **lineage tree**. When you [branch](saving-and-loading.md#branching-your-save) a new save off your current run, it appears **indented beneath its parent** with tree connectors (`├─`, `└─`), so you can see at a glance which saves descend from which. Top-level roots (saves you started fresh, plus any orphans) are ordered most-recent first, and each parent's children are likewise ordered most-recent first.

A **● active** marker shows which save your game is currently autosaving into — the save the periodic autosave and `Esc` quick-save follow.

If a save's parent has been **deleted**, the child can no longer point at it, so it becomes an **orphan** and is promoted to the top level of the tree (its detail pane marks the lost parent as *detached*). **Renaming** a save, by contrast, **keeps its children attached** — they're automatically re-parented to the new name (and re-signed, so they don't load flagged as modified), so the lineage follows the rename intact. Children that are themselves flagged as *modified* are left untouched, so renaming never launders a tampered save's badge.

**Keys inside the browser:**

| Key | Action |
|---|---|
| `↑` / `↓` | Move the highlight between saves |
| `Enter` | Load the highlighted save |
| `d` | Delete the highlighted save (asks you to confirm first) |
| `r` | Rename the highlighted save |
| `c` | Duplicate the highlighted save |
| `Esc` | Return to where you opened it from — the main menu, or your current run if opened mid-game |

---

## Save badges

Saves can carry a row tag in the list, explained on-screen in a bordered **Key** box:

| Tag | Meaning |
|---|---|
| ★ auto | The automatic save slot |
| ● active | The save your game is autosaving into (the active slot the autosave follows) |
| ⚠ modified | The save file was edited outside the game (integrity check failed) |
| ⚠ corrupt | The file could not be read. It is still listed but dimmed, and cannot be loaded |

---

## Save integrity

Saves are signed. If a save file is edited outside the game, the integrity check fails and the file is flagged `⚠ modified` in the browser. A `⚠ corrupt` file is one the game couldn't read at all — it stays in the list, dimmed, but cannot be loaded.

---

## Tips

- **Branch at milestones.** `save iron_age_start` forks a new save at that moment and moves autosave onto it, freezing the old run where it was — your snapshot can't be overwritten.
- **Duplicate before risky moves.** Press `c` in the Load Game browser to clone a save before a prestige, catastrophe, or any decision you might want to undo. (Branching does much the same, but keeps you playing on the new copy rather than the old one.)
- **Your active save is overwritten constantly.** Autosave keeps your current game up to date — but that means it's not a snapshot. To keep a run exactly as it is right now, duplicate it (`c`) or save it under a new name before risky moves.
