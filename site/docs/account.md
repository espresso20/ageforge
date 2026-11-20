# Accounts & Recovery

AgeForge is a local game. There is no server, login screen or cloud. Everything lives on your machine, in the game's data folder: `data/` beside the `ageforge` program in a build from source, or your user config folder in a release build (see [Updating](getting-started.md#updating)). Your accounts are small files in that folder. They give your civilizations a stable identity and hold what you earn across games: theme unlocks, lifetime stats, badges and your display settings.

You can keep several accounts on one machine, with one active at a time. This page covers how accounts are laid out on disk, how to switch between them, how to export, back up and import them, what the recovery code does and does not do, and how to carry an identity to another machine.

---

## Multiple accounts, one at a time

Each account has its own slot on disk, and exactly one account is active at any moment. The active account's saves are the ones in the [Load Game](saving-and-loading.md) browser, and its progress is the progress in play. Switching accounts changes both.

```
data/
├── active-account                  # which account is active
├── accounts/
│   └── <account_id>/
│       ├── account.json            # identity, theme unlocks, lifetime stats, theme, map glyphs, mini map
│       ├── badges.json             # badges and the counts behind them (written with every save of the account)
│       ├── runs.json               # the run archive: a record of each run that ended
│       ├── settings.json           # motion, worn title and current game
│       ├── account-<id8>-export.json   # where `account export` writes by default
│       └── saves/                  # this account's game saves
└── backups/
    ├── <name>-<id8>-<timestamp>/   # full copies made by `account backup`
    └── <id>.before-4.0/            # an account from before 4.0 whose name was used again
```

`<account_id>` is 32 characters of hex, and `<id8>` is its first 8. Saves are per account and never mix. See [Saving & Loading](saving-and-loading.md) for the save layout.

Each account writes only to its own slot. Switching, importing or recovering never writes one account's data into another account's files, whichever account is active at the time.

### Switching during a game

A game belongs to the account it was started or loaded under. Its saves go to that account's slot, and its prestiges and age-ups count for that account only.

`account switch <name>` switches accounts during a game. The name must belong to an account on this machine, and `account list` shows them. The game is saved to the account it belongs to, and you go back to the main menu, which shows the other account's current game. Continue opens it, or you can load another of its games or start a new one. The game you left never carries on under the other account. The Accounts panel on the main menu switches the same way, with no game running.

### Accounts from before 4.0

The game no longer moves or converts old accounts. Every account the game saves carries a format number, 4, in `account.json`. An account without it was written before 4.0, played or not, and the game refuses it with the line saves from before 4.0 get: "This was saved before 4.0 and can no longer be loaded." It is not listed in the Accounts panel and is left on disk as it is. To play, make an account: if you give it the same name as the refused one, the game moves the old account's whole folder to `data/backups/<id>.before-4.0`, starts a new empty account under that name, and says so in the log. An export from before [badges](#badges) is refused too.

Each time the game loads an account, it adds the badges the account's record already proves: the badge of every age the account has reached, and ladder rungs whose counts the account has passed. It reports them in one toast and one log line, such as `12 badges earned from your earlier play.`

Five themes come with age badges (see [Themes from badges](badges.md#themes-from-badges)). If an account has reached one of those ages without the theme, the game adds the badge and the theme when it loads the account.


---

## Your account

The first time you launch the game, it asks you to name your account. The prompt comes up before the main menu, with a suggested name already in the box. Tab suggests another, Enter submits, and Esc accepts what is in the box. A name can be up to 64 characters. It cannot contain `/`, `\` or `..`, and it cannot start with a dot.

The game then asks you to confirm: "Create your AgeForge account as ...?" Choose Create to lock the name in, or Re-type to fix it. The name is your identity, and a typo would quietly create a different, empty account. The confirmation reminds you to enter the name exactly as written to restore your account on another device.

> Your name is your identity. The account ID is derived from your name: a SHA-256 hash of the normalized name, cut to a 32-character hex ID. Normalizing lowercases the name, trims it and collapses inner spacing before hashing. `Imperium`, `imperium` and the same name with stray spaces around it all give the same ID. The display name keeps your original casing. Entering the exact same name on another machine gives the exact same account ID. That is the simplest way to restore your identity (see [Restoring on a new machine](#restoring-on-a-new-machine)).

The account holds two things:

| Part | What it is |
|---|---|
| Identity | Your chosen name and the account ID derived from it |
| Data | Your earned progress (theme unlocks, lifetime stats, badges) and your settings (see [Account settings](#account-settings)) |

The two are recovered in different ways. Your identity is carried by your account name or by the recovery code, which both point at the same ID. Your data is backed up with an account export (see [Exporting & importing accounts](#exporting-amp-importing-accounts)).

You choose the name once. A different name makes a different identity, so there is no in-game rename. If you want a second civilization to play in parallel, create another account with `n` in the Accounts panel. It starts empty and takes nothing from the account you were using.

---

## The Accounts panel

Accounts on the main menu (`a`) opens a full-screen panel that lists every local account on this machine. Each row in the list shows:

- the display name,
- a short ID (the first 8 characters of the account ID),
- a `(current)` marker on the active account, and
- a `⚠ modified` flag if that account's `account.json` was edited outside the game. The flag stays once it is set. Saving the account again, deleting the flag from the file by hand, or exporting and importing the account does not clear it.

The detail pane beside the list shows the selected account's full ID and recovery code, its highest age, its total prestiges, its badge count and its status. `account list` prints a text version of the list.

| Key | Action |
|---|---|
| `Enter` | Switch to the selected account. It becomes active, and its saves show in the Load Game browser |
| `n` | Name and create a new account alongside your existing ones |
| `e` | Export the selected account to a signed backup file |
| `b` | Back up the selected account: a full copy of its slot to `data/backups/` (see [Backups](#backups)) |
| `i` | Import an account from an export file. The panel always merges |
| `r` | Show the selected account's recovery code (see [The recovery code](#the-recovery-code)) |
| `x` | Delete every save of the selected account, after a confirm. The account itself is kept (see [Deleting saves](saving-and-loading.md#deleting-saves)) |
| `w` | Wipe the selected account (permanent, behind a type-the-name confirm) |
| `Esc` or `q` | Back to the main menu |

In the window, a click on an account selects it.

On a narrow window the keys are set over two lines. For wiping, see [Wiping an account](#wiping-an-account).

---

## Exporting & importing accounts

An account has two kinds of backup, and they do different jobs.

| Backup | What it carries | How to keep it |
|---|---|---|
| Account export | The account's ID, name, theme unlocks, lifetime stats, badges, [run archive](#run-archive), theme, map glyphs and mini map setting | Save the export file somewhere safe |
| Recovery code | Identity only (the account ID) | Write down the short `AGEF-…` string |

### Exporting

From the Accounts panel press `e`, or run:

```
account export
```

This writes a signed backup of the active account. The file is bound to its account ID: the signature covers the ID, so a backup cannot be passed off as another account's. It also carries the badges with the counts behind them, and the modified flag if the account has one. By default the file is `account-<id8>-export.json` in that account's own slot (`data/accounts/<id>/`). To choose a location, pass a path:

```
account export /path/to/my-ageforge-backup.json
```

An export also takes a full backup of the account (see [Backups](#backups)) and reports its path. If that backup fails, the export still stands.

An export is a snapshot and does not update as you keep playing, so export again after anything you would hate to lose. It does not carry `settings.json`, which holds motion, your worn title and your current game.

### Importing

On the same machine or another one, bring an export in from the Accounts panel with `i`, or run:

```
account import /path/to/my-ageforge-backup.json
```

Import goes by the account ID inside the file, and it always lands in that account's own slot:

- If no account with that ID exists on this machine, import creates it from the file.
- If that account already exists, import merges into it. This is the default.

Import never overwrites a different account. At worst it updates the account the file belongs to, creating it if needed. Afterward the account shows up in the Accounts list, ready to switch to.

Import never switches accounts, from the panel or from the command. A game in progress carries on under the account it belongs to. To play as the imported account, switch to it with `account switch <name>` or from the panel. A backup of the account you are using goes straight into it, and anything it earned since its last autosave is kept.

A merge keeps the best of both copies:

- Theme unlocks are combined, so an old backup never removes a theme you unlocked since.
- Badges are combined. You keep every badge either copy holds, and for a badge both hold, the earlier date is kept.
- Lifetime stats and the counts behind your badges take the higher of the two values. Counts are never added together, so importing your own backup twice changes nothing.
- Highest age takes the later age of the two.
- Theme, map glyphs and the mini map setting keep your choice if you have made one. Otherwise they take the file's.

To overwrite that account's progress entirely with the file, add `replace`:

```
account import /path/to/my-ageforge-backup.json replace
```

The import is refused, and your accounts are left unchanged, if the file cannot be read, if it was edited or damaged since it was made (its badge section included), if it has no badge section (an export from before 4.0), or if it has no usable account ID. A file from an account flagged modified keeps the flag and passes it to the account it lands in.

> With no server there is no copy elsewhere. You can only recover progress you exported. If a machine's `data/` folder is gone and you never exported, that account's earned progress is gone too.

---

## Backups

A backup is a full copy of an account's slot: `account.json`, `badges.json`, `runs.json` and `settings.json` (each when the slot has it) plus a copy of the slot's whole `saves/` folder. It holds more than an [export](#exporting-amp-importing-accounts). An export is one file of account-wide progress with no saves, and a backup includes your games.

The game makes a backup at three points:

- Before any wipe, so a wipe always leaves a recoverable copy. The wipe still goes ahead if the backup fails.
- On export. `account export` and the panel's `e` take a full backup right after writing the export file.
- On demand. `account backup` backs up the active account, and `b` in the Accounts panel backs up the selected one.

Backups live under the data folder, beside `data/accounts/` and not inside it:

```
data/
└── backups/
    └── <name>-<id8>-<timestamp>/
        ├── account.json
        ├── badges.json
        ├── runs.json
        ├── settings.json
        └── saves/
```

`<name>` is the display name made safe for a file name, `<id8>` is the first 8 characters of the account ID, and `<timestamp>` is `YYYYMMDD-HHMMSS`. Because backups sit outside `data/accounts/`, wiping an account's slot never deletes its backups.

The game keeps the 10 most recent backups per account. After each new backup, it deletes that account's older ones until 10 remain. Other accounts' backups are never touched.

### Restoring from a backup

There is no restore command. A backup is plain files, so you put them back by hand. Copy the backup folder's `account.json`, `badges.json`, `runs.json` and `settings.json` (each if it has one) and `saves/` into that account's slot at `data/accounts/<id>/`. The `<id8>` in the backup folder's name is the start of the full ID, and the full ID is the slot's folder name under `data/accounts/`.

---

## The recovery code

Run `account show code` at the `>` prompt to see the active account's recovery code, and nothing else, in a small window that `Esc` closes. The code is not written to the game log.

```
account show code
```

The bare `account` command shows the short ID and the backup help, and points to this command. The Accounts panel shows the selected account's code in its detail pane, and `r` shows it in a window.

The recovery code looks like this:

```
AGEF-7Q2K-9X4M-ZJ31-…
```

It is an uppercase string of 41 characters, including the `AGEF-` prefix, in groups of four separated by dashes. It uses [Crockford base32](https://www.crockford.com/base32.html), which leaves out the easily confused letters `I`, `L`, `O` and `U`, so it copies cleanly when you write it on paper or read it aloud.

The code holds only your account ID and a checksum.

- The checksum catches typos. If you mistype the code, recovery is refused with a message that a character is probably wrong, instead of restoring the wrong account.
- It is not a password. The checksum guards against typos, not against other people. The code is an identifier and not a secret.

### Identity vs progress: what the code does and doesn't carry

| | Restored by the recovery code? |
|---|---|
| Your identity (account ID) | Yes |
| Your earned progress (theme unlocks, lifetime stats, badges) | No |

The code restores your identity across machines and reinstalls. It carries no progress, which is why it is short. To carry progress between machines, use an account export (see [Exporting & importing accounts](#exporting-amp-importing-accounts)). The code restores who you are, and the export restores what you have earned. A full move to another machine uses both.

---

## Lifetime stats & badges

Some progress is account-wide. It builds up across every game you play on that account and every prestige, not just your current run. It lives on the account, apart from the per-save Statistics that start over with a new game or a prestige.

| Lifetime stat | What it tracks |
|---|---|
| Total Prestiges | Every prestige you have completed on this account, across all its games |
| Civilizations Started | Every civilization you have begun on this account: each new game, and the one that follows each prestige and each Succumb. Loading a game starts nothing |
| Highest Age Ever | The furthest age any of this account's civilizations has reached. It only goes up |

The Stats panel's Lifetime part is the account's record board. At the top it shows the account's name, its title and a bar toward the next title, with the counts of prestiges, civilizations started and badges earned. Below, a ladder has one step for each age in order: the steps for ages the account has reached are lit in the age's color, a crown marks the highest, and the pointer, moved with Left and Right or by a click, shows the fastest time to that age and the run and date that set it. Under that are the four best single-run counts, and a grid of the worker domains, each with the best tier earned in it and how many of its badges are held. The [Accounts panel](#the-accounts-panel) shows the prestiges and highest age for each account, so you can compare your civilizations.

The account also records each prestige under the age it was made from (`prestiges_by_age` in the stats of `account.json`), so an early taste (a prestige from the Medieval Age to the Atomic Age) can be told from a full run (the Modern Age or deeper). See [Early tastes and full runs](prestige.md#early-tastes-and-full-runs). No panel shows this, and Total Prestiges counts every prestige, tastes included. Prestiges made before the account kept this record count only in the total. An export carries it, and an import keeps the higher count for each age.

The account also keeps two kinds of record across runs (`fastest_age` and `best_run` in the stats of `account.json`): the fewest game ticks it has taken a run to enter each age, and the best single-run figures for buildings built, workers recruited, techs researched and expeditions and campaigns done. Each record stores when it was set and in which save. Age times update as you enter an age, and the run figures when you enter an age and when a run ends (a prestige, a Succumb or a new game). The Stats panel's Lifetime part shows them. An export carries them, and an import keeps the faster time and the higher figure.

These stats update the moment you prestige or advance into an age. They are saved to your account with the next autosave, on a clean exit, and when you switch accounts, so a fresh prestige is never lost. They count for the account the game belongs to. A game started or loaded under one account never adds to another's.

### Badges

A badge is a permanent mark on your account for something you did in a game: reaching an age, building enough of something across all your runs, or pulling off something specific in one run. A badge is earned once and kept for good. It stays with your account through every prestige and every new game, and it travels in an export. The [Badges](badges.md) page lists the families, the ladders and their rungs, the titles and the themes that badges give. This section covers the Ranks screen and how badges are kept.

Badges are separate from [milestones](milestones.md). A milestone belongs to one run, pays a reward inside that run and starts over with the next. A badge belongs to the account and gives nothing inside a run, so the same game plays the same way whatever the account has earned.

When you earn a badge, the log gets one line, such as `Badge earned: Rock Solid (bronze). Reach the Stone Age.` A badge that gives a theme adds a second log line. Badges your account is given all at once when a game loads (rungs whose counts it had already passed) get one notice in the header that says how many.

To see your badges, open the Ranks screen:

```
badges
```

The main menu's Badges entry (`b`) opens the same screen with no game loaded, and shows your count and your title on the menu itself.

#### Ranks

Every badge belongs to one worker domain, and each domain has a rank built from the points of the badges earned in it. Points are account-wide: they persist across runs and prestige. A badge is worth its tier's points (bronze 5, silver 10, gold 25, platinum 50, legendary 100); a badge earned in a modified game adds none, and neither does an integrity badge. The ranks are placeholders for now: Apprentice, Journeyman, Adept, Expert, Master and Grandmaster, at 25, 100, 190, 320, 460 and 700 points.

**Each rank adds 2% to the output of that domain's workers**, so Grandmaster adds 12%. Adept also puts the domain's title beside the town's name, on the map's top row and on the Epoch screen. The Grandmaster theme color is not in yet.

The screen fills the window except for the command bar, so the prompt keeps working while it is open. The heading and three tabs sit at the top: Ranks, All badges and Recent. A count line gives the badges earned, the points and the ranks held.

- Ranks lists the worker domains, the highest rank first. Each row shows the rank, the points, a bar toward the next rank and how many badges the domain has earned. Wide windows add a detail column for the picked domain: its points to the next rank, its badges by tier and the badges nearest to earned.
- All badges lists every badge in sight by domain; `Enter` on a domain lists only its badges.
- Recent lists the earned badges, the latest first.
- Keys: `↑` `↓` pick, `Enter` opens a domain, `Tab` cycles the tabs, `PgUp` `PgDn` page, `Esc` steps back and then closes. A click on a tab or a row does the same, and a double click opens a domain.

The keys are listed under [Badges](commands.md#badges) on the commands page and in the Help panel.

#### Titles

Your badge points add up to a score, and the score holds a title. Some badges give a title of their own. `title` lists the ones you hold, `title <name>` wears one, and `title default` goes back to the score's. Once you wear anything but Settler, the header's top edge shows the title beside your account's name when the window has room for it. A title changes nothing in a game. [Titles](badges.md#titles) lists them all with the scores they ask for.

Each badge has a tier, and each tier is worth points. The tiers, their points and how badges are earned are on the [Badges](badges.md) page. A badge about something you have not come across yet stays out of the lists and out of the counts until you get there.

### Where badges are stored

Badges are kept in their own file, `badges.json`, beside `account.json` in the account's slot (`data/accounts/<id>/`). It holds the badges you have earned, the counts behind the ones that count across runs, and the days you have played. The game writes it every time it saves the account, even with no badges yet. A badge the game has retired stays in the file. It no longer shows and no longer scores.

`account.json` keeps the shape it had before badges, on purpose: it holds what it always held, and the badges are in `badges.json` beside it. When this version opens an account, it adds the badge of any age the account has reached that it does not hold yet.

The file is signed, and the signature covers the account ID. A `badges.json` that was edited by hand, or copied in from another account, is flagged, and its badges are listed as earned in a modified game. That flag belongs to `badges.json` only. It never marks `account.json` as modified.

If `badges.json` is deleted, the game treats it like an edited one: the account still loads, a new empty `badges.json` is started and flagged, and every badge earned from then on is listed as earned in a modified game. The badges that were in the file are gone, so keep an [export](#exporting-amp-importing-accounts) or a [backup](#backups). A file that cannot be read is set aside the next time the account is saved. If `account.json` is lost and `badges.json` stays, an account made again under the same name, or recovered with its code, takes the badges back. An export carries the badge file inside it, and a backup copies it, so both bring your badges back.

---

## Account settings

Some settings belong to the account, not to a save. Switching accounts swaps them, and a new account starts with the defaults.

| Setting | Command | Default | Kept in |
|---|---|---|---|
| Theme | `theme <key>` | Forge | `account.json` |
| Map glyphs | `map glyphs ascii` or `unicode` | unicode | `account.json` |
| Mini map | `minimap on` or `minimap off` | on | `account.json` |
| Discord status | The Settings page, row **Discord status** | on | `account.json` |
| Motion | `motion on` or `motion off` | on | `settings.json` |
| Worn title | `title <name>` | The title your score holds | `settings.json` |
| Current game | `m` in the Load Game browser | The game you played last | `settings.json` |

**Discord status** shows your age and town on your Discord profile while the Discord desktop app is running: "Playing AgeForge", then the age and its year, then your town and epoch (and your prestige level once you have one). It talks to the Discord app on your own machine and sends nothing anywhere else. If Discord is not running, nothing happens. Turn it off on the Settings page and the profile clears at once.

The Map's zoom is not saved with the account. With no account loaded, the display settings last for the session only.

### The settings file

`settings.json` holds the motion setting, the title you chose to wear (`title`), and which game the main menu's Continue opens: the game you played last, and your [main game](saving-and-loading.md#continue-and-your-main-game) if you marked one. It sits beside `account.json` for the same reason `badges.json` does: `account.json` keeps the shape older versions of the game know, so a newer setting goes in a file of its own.

The file holds preferences and nothing you earn, so it is plain text and unsigned. The game writes it when one of its settings changes and the first time you play a game. If it is missing or cannot be read, motion is on, you wear the title your score holds, and Continue opens the save written most recently. A title is worn only while the account holds it, so writing one into the file by hand gives nothing. A [backup](#backups) copies this file, and an export does not.

---

## Edited and damaged files

The game signs the files that carry your progress: saves, `account.json`, `badges.json` and exports. The signature is a check, not a lock. A save or an account file that fails the check still loads, and the game marks what came from it as modified. An export that fails it is refused.

| File | If it was edited outside the game | What you see |
|---|---|---|
| A save | The game loads it and marks the game modified. The saves you make from it carry the mark, and a new game starts clean | A red line at the top of the game screen (one of six, such as `✦ FORGED SCROLLS ✦`), `⚠ modified` in the Load Game browser, and every badge earned in that game is crossed |
| `account.json` | The account is flagged and stays flagged. Exports carry the flag | `⚠ modified` in the Accounts panel and `account list`, and every badge earned on that account is crossed |
| `badges.json` | The file is flagged, and so are all its badges. The flag never marks `account.json` | Every badge in the file is crossed |
| An export | The import is refused and nothing changes | An error message |

A crossed badge is listed as earned in a modified game. It adds no points and gives no title. The theme it gives, if any, still unlocks.

The developer console is a testing tool and does not block badges. A game it has changed is marked in its save, and the log says so once, but it still records to the account like any other game. Unlocking the console earns a badge of its own.

A file the game cannot read at all is set aside and never overwritten. An unreadable `account.json` becomes `account.json.corrupt`, and the game asks you to name the account again, as on a first launch. An unreadable `badges.json` becomes `badges.json.corrupt` the next time the game has badges to write.

---

## Restoring on a new machine

There are two ways to restore your identity, because your name and your recovery code both lead to the same account ID. Either way, restoring your identity does not bring your earned progress with it. For that you also need an [account export](#exporting-amp-importing-accounts).

### The simplest way: re-enter your name

On a fresh machine, the first-run prompt asks you to name your account. Type the exact same name you used before and you get the exact same account ID, with no code needed. Casing and spaces around the name do not matter, but a different name is a different identity.

### With the recovery code

On the other machine, or after a reinstall, run:

```
account recover AGEF-7Q2K-9X4M-ZJ31-…
```

This restores the identity in the code and switches to it. The account lands in its own slot:

- If that account is already on this machine, the game opens it as it is, progress included.
- If not, the game creates it with the ID only, with no unlocks or stats. Bring those back with an [import](#exporting-amp-importing-accounts).

Recovering never overwrites an account, the one you are using included. That account keeps everything in its own slot, and you can switch back to it with `account switch <name>` or from the Accounts panel. If the account you are using holds any progress (theme unlocks, badges or lifetime stats), the command first says what it holds and asks you to run `account recover <code> confirm`. During a game, the game is saved to its account first and you go back to the main menu. Recovering the code of the account you are using does nothing.

Recovery forgives copying mistakes:

- Lowercase is fine.
- Spaces and dashes are ignored.
- The easily confused `I` and `L` are read as `1`, `O` is read as `0`, and `U` is read as `V`.

So if you wrote it down by hand and your `0` looks like an `O`, it still works.

---

## Wiping an account

If you want a clean slate for an account, with none of its old unlocks, stats or badges, you can wipe it.

Wiping is in the Accounts panel: select the account and press `w`. It is not a typed command. Deleting an account is permanent, so it sits behind two steps:

1. Read the warning. It says what is about to be deleted. Choose Wipe it to go on, or Keep it to stop.
2. Type the account's display name exactly, with the same casing. Anything else, or pressing Esc, cancels with no changes.

On confirmation, the wipe permanently deletes that account's:

- identity (the account name and derived ID),
- theme unlocks,
- lifetime stats,
- badges, and
- every save in its slot.

This cannot be undone in the game, and no server keeps a copy. The game [backs up](#backups) the slot, saves included, to `data/backups/` first and shows the backup's path. Restoring that copy is manual. Otherwise the old identity comes back only if you wrote down its [recovery code](#the-recovery-code) beforehand, and its earned progress only if you [exported it](#exporting-amp-importing-accounts) first. If you wipe your last account, the game asks you to name a new one.

> Typing `account wipe` at the `>` prompt does not wipe anything. It tells you where to find the wipe in the Accounts panel, the only place it happens.

---

## Run archive

The account keeps a record of each run that ends: by a prestige, by a Succumb, when a new game is started over it, or when the account's saves are deleted while it is loaded. A record holds the run's id, the save it was played in, when it started and ended, its last tick, the age it reached, its prestige level, how it ended, and its counts of buildings built, workers recruited, techs researched and expeditions done. A prestige is always recorded. Any other run is recorded once it has about 300 ticks behind it.

The 20 newest records, and the best run (the highest age, then the fewest ticks), also keep a thinned copy of the run's history: ten measures at up to 256 points each, with its age, wonder, catastrophe and harbinger marks. The account keeps up to 200 records and drops the oldest past that. No panel shows them yet.

A run left for a new game is still in its save. If that save is loaded and the run ends later, the new record replaces the old one. Saves branched from the same game share one run id, so they share one record.

The records are in `runs.json`, a signed file beside `account.json`. An export carries them, an import merges them by run id (the copy that ran longer is kept), and a backup copies the file. If the file is missing, damaged or edited, the archive starts again empty and nothing else about the account changes.

---

## Recovery in short

With no server, you recover your identity from a short code you write down. You recover progress only if you exported it first. If a machine's `data/` folder is gone and the progress was never backed up, no server holds a copy. The recovery code restores who you are, not what you have earned.

That is also why the code is safe to share or lose. It names an account and unlocks nothing.

---

## See also

- [All Commands](commands.md): the full `account` command reference
- [Saving & Loading](saving-and-loading.md): how your game saves are stored per account under `data/accounts/<id>/saves/`
- [Badges](badges.md): the badge families, ladders, titles and themes
