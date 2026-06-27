# Saving & Loading

AgeForge keeps your civilization safe with named saves, a continuous autosave, and a full save browser. This page covers where saves live, the commands that manage them, and how to read the Load Game browser.

---

## Starting a new game

When you choose **New Game**, you're prompted to name your civilization — pre-filled with a randomly generated name. Press **Enter** to accept it, type your own, or press **Tab** to roll a fresh suggestion. That name becomes your active save, and the game autosaves into it from there.

---

## Where saves live

Save files are written to `./data/saves/*.json`, relative to the directory you launch the game from. One file per save. The `save`/`load` commands and the **Load Game** browser all read and write the same files, so a save made one way shows up the other.

---

## Commands

| Command | Description |
|---|---|
| `save [name]` | Save to a named slot. With no name, writes to the active save slot (see below) |
| `load [name]` | Load a named save |
| `saves` (or `save list`) | List all save files |
| `Esc` | Quick-save to your current (active) save |

```
save hero
save
load hero
saves
```

---

## Active save slot

The game remembers the last slot you explicitly `save`d to or `load`ed. A bare `save` (no name) re-writes **that** slot — it is not a generic dump.

- Until you've named a save or loaded one this session, a bare `save` defaults to the `autosave` slot.
- Once you `save hero` or `load hero`, a bare `save` thereafter targets `hero`.

The periodic autosave and `Esc` quick-save both write to your **active** save, continuously overwriting it. Your current game *is* the autosave — there's no separate slot quietly shadowing it.

---

## Autosave

The game autosaves periodically and whenever you press `Esc`, writing to your **active** save — so your current game is always kept up to date on disk. To preserve a specific point you don't want overwritten, save it under a new name (or duplicate it with `c` in the Load Game browser).

---

## The Load Game browser

From the main menu, choosing **Load Game** opens a save browser that lists every save in `./data/saves/`, most-recent first. Highlighting a save updates a **detail pane** showing everything you need to size up that save before loading it: its age and epoch, population, buildings, wonders, milestones, techs, soldiers, prestige, [morale](morale.md), a ⚠ warning if a catastrophe is pending, and the exact save time.

**Keys inside the browser:**

| Key | Action |
|---|---|
| `↑` / `↓` | Move the highlight between saves |
| `Enter` | Load the highlighted save |
| `d` | Delete the highlighted save (asks you to confirm first) |
| `r` | Rename the highlighted save |
| `c` | Duplicate the highlighted save |
| `Esc` | Return to the main menu |

---

## Save badges

Saves can carry a row tag in the list, explained on-screen in a bordered **Key** box:

| Tag | Meaning |
|---|---|
| ★ auto | The automatic save slot |
| ⚠ modified | The save file was edited outside the game (integrity check failed) |
| ⚠ corrupt | The file could not be read. It is still listed but dimmed, and cannot be loaded |

---

## Save integrity

Saves are signed. If a save file is edited outside the game, the integrity check fails and the file is flagged `⚠ modified` in the browser. A `⚠ corrupt` file is one the game couldn't read at all — it stays in the list, dimmed, but cannot be loaded.

---

## Tips

- **Name your important saves.** A bare `save` targets your active slot, but giving milestones their own names (`save iron_age_start`) keeps them out of harm's way.
- **Duplicate before risky moves.** Press `c` in the Load Game browser to clone a save before a prestige, catastrophe, or any decision you might want to undo.
- **Your active save is overwritten constantly.** Autosave keeps your current game up to date — but that means it's not a snapshot. To keep a run exactly as it is right now, duplicate it (`c`) or save it under a new name before risky moves.
