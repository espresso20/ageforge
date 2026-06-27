# Saving & Loading

AgeForge keeps your civilization safe with named saves, a continuous autosave, and a full save browser. This page covers where saves live, the commands that manage them, and how to read the Load Game browser.

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
| `Esc` | Quick-save to the `autosave` slot |

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

The periodic autosave (and `Esc`) always writes to its own separate `autosave` slot, regardless of which slot a bare `save` targets. A background save will never clobber your named file.

---

## Autosave

The game autosaves periodically and whenever you press `Esc`, always to the `autosave` slot. Treat it as a safety net rather than your primary save — name the runs you care about so a routine autosave can't overwrite them.

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
- **Autosave is your safety net, not your archive.** It's there if the game closes unexpectedly — but it gets overwritten constantly, so don't rely on it for runs you want to keep.
