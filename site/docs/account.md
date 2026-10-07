# Account & Recovery

AgeForge is a **local terminal game**. There is no server, login screen or cloud: everything lives on your own machine in the `data/` folder next to the `ageforge` binary. Your accounts are part of that. They are small files on disk that give your civilizations a stable identity and hold the progress you earn across games.

The game supports **multiple local accounts**, each in its own slot on disk, with one active at a time. This page explains how accounts and their data are laid out, how to switch between them, how to back them up (export) and bring them in (import), what the recovery code does and doesn't do, and how to carry an identity to a new machine.

---

## Multiple accounts, one at a time

You can keep **several accounts side by side** on the same machine. Each is its own slot on disk, and exactly **one is active** at any moment. The active account is the one whose saves you see in the [Load Game](saving-and-loading.md) browser and whose account-wide progress is in play. Switching accounts swaps both: a different identity and a different set of saves.

Each account lives in its own directory:

```
data/
├── active-account               # pointer to the currently active account
└── accounts/
    └── <account_id>/
        ├── account.json         # that account's identity and account-wide progress
        ├── badges.json          # that account's badges (once it has any)
        ├── settings.json        # that account's motion setting (once it is changed)
        └── saves/               # that account's game saves
```

The `data/active-account` pointer records which account is current. Each `data/accounts/<account_id>/` slot holds that account's `account.json`, its `badges.json` (see [Where badges are stored](#where-badges-are-stored)) and its own `saves/` folder, so saves are **per-account** and never mix between accounts. (See [Saving & Loading](saving-and-loading.md) for the save layout.)

**Each account only ever writes to its own slot.** Switching, importing or recovering never writes one account's data into another account's files, whichever account is active at the time.

### Switching during a game

A game belongs to the account it was started or loaded under. Its saves go to that account's slot, and its prestiges and age-ups count for that account only.

You can switch accounts during a game with `account switch <name>`. The game is saved to the account it belongs to, and you go back to the main menu, where you load one of the new account's games or start a new one. The game you left never carries on under the other account. (The **Accounts** panel on the main menu switches the same way; there is no game running there.)

### Upgrading from an older version

Older builds kept a single account at the top level (a flat `data/account.json` with saves in `data/saves/`). If the game finds that layout on first launch, it moves your account and its saves into their own `data/accounts/<id>/` slot and makes that account active. **Nothing is deleted**, and you don't have to do anything; your civilizations and unlocks come across intact. Just before the move, the game also copies your old flat data into `data/backups/pre-migration-<timestamp>/`, so your original files stay recoverable.

An account from before [badges](#badges) gets its badges the first time this version opens it. Nothing is lost and `account.json` is left as it was: the four old achievements become their badges, and the badges its record already proves (an age it has reached, prestiges it has made) are added. None of that is announced, and it happens once. Saves from before this version load as they are.

---

## Your account

The **first time you launch the game**, AgeForge asks you to **name your account**. The prompt comes up before the main menu, filled in with a suggested empire-style name you can keep, edit or reroll. When you submit, the game **asks you to confirm the name before creating the account**. The name *is* your identity, so a typo would quietly create a different, empty account. Choose **Create** to lock it in, or **Re-type** to fix it. The confirmation also reminds you to **re-enter the name exactly** to restore your account on another device. Whatever you settle on becomes your account's display name and your **identity**.

> **Your name *is* your identity.** The account ID is **derived from your name**: a SHA-256 hash of the normalized name, cut to a 32-character hex ID. Normalizing lowercases the name, trims it and collapses internal spacing before hashing, so `Imperium`, `imperium`, and the same name with stray surrounding spaces all give the *same* ID. The display name keeps your original casing. Because the ID comes from the name, **re-entering the exact same name on a new machine gives the exact same account ID**. That is the simplest way to restore your identity (see [Restoring on a new machine](#restoring-on-a-new-machine)).

The account holds two distinct things:

| Part | What it is |
|---|---|
| **Identity** | Your chosen name and the account ID derived from it |
| **Data** | Your earned account-wide progress (theme unlocks, lifetime stats, badges) and your prefs: your theme, map style, map glyphs, mini map setting and motion setting |

The split matters because the two halves are recovered very differently (see below). Your **identity** is carried by either your account name *or* the recovery code (both point at the same ID). The **data** is backed up separately with an account **export** (see [Exporting & importing accounts](#exporting-amp-importing-accounts)).

> **You choose the name once.** Because the ID comes from the name, picking a *different* name later creates a *different* identity; it does not rename the account. For that reason there is no in-game rename, so choose a name you're happy to keep. (If you want a second civilization to play in parallel, create a **new account**; see the Accounts panel below.)

---

## The Accounts panel

The main menu has an **Accounts** entry that opens a full-window panel listing **every local account** on this machine. For each one it shows:

- the **display name**,
- a **short ID** (the first part of the account ID),
- the **highest age** that account has ever reached,
- its **total prestiges** and how many **badges** it holds,
- a **current** marker on the account that's active right now, and
- a **modified** flag if that account's file was edited outside the game. The flag stays once it is set: saving the account again, deleting the flag from the file by hand, or exporting and importing the account does not clear it.

From the panel:

| Key | Action |
|---|---|
| `Enter` | **Switch** to the highlighted account (it becomes active; its saves show in the Load Game browser) |
| `n` | **New account**: name and create a fresh account alongside your existing ones |
| `e` | **Export** the highlighted account to a signed backup file |
| `b` | **Backup** the highlighted account: a full copy of its slot (`account.json`, `badges.json`, `settings.json` and `saves/`) to `data/backups/` (see [Backups](#backups)) |
| `i` | **Import** an account from a backup file |
| `r` | Show a **recovery code** for restoring an identity (see [The recovery code](#the-recovery-code)) |
| `w` | **Wipe** the highlighted account (permanent, behind a type-the-name confirm) |
| `Esc` | **Back** to the main menu |

This panel is where you manage accounts: switch between civilizations, create new ones, back them up, and wipe one behind the type-your-name confirm described under [Wiping an account](#wiping-an-account).

---

## Exporting & importing accounts

Each account has a backup of its **data** that is separate from its recovery code. The export carries your earned progress; the recovery code carries only your identity. They do different jobs and you keep them differently:

| Backup | What it carries | How to keep it |
|---|---|---|
| **Account export** | The account's ID, name, theme unlocks, lifetime stats, badges and prefs | Save the export file somewhere safe |
| **Recovery code** | Identity only (the account ID) | Write down the short `AGEF-…` string |

### Exporting

From the Accounts panel press `e`, or run:

```
account export
```

This writes a **signed backup of the active account**. The file is **bound to its account ID**: it carries the account's ID, name, theme unlocks, lifetime stats, badges (with the counts behind them) and prefs, and **the signature covers the ID**, so a backup can't be passed off as a different account's. By default it's written as `account-<id8>-export.json` inside that account's own slot (`data/accounts/<id>/`). To choose your own location, pass a path:

```
account export /path/to/my-ageforge-backup.json
```

Export after any big unlock. An export is a snapshot and does **not** update as you keep playing, so export again when you've earned something you'd hate to lose.

### Importing

On the same or a new machine, bring a backup in from the Accounts panel with `i`, or:

```
account import /path/to/my-ageforge-backup.json
```

Import goes by the **account ID inside the backup**, and it always lands in **that account's own slot**:

- If **no account with that ID exists** locally, import **creates** it from the backup.
- If that account **already exists** locally, import **merges** into it (the default; see below).

**Import never overwrites a *different* account.** Because it goes by the ID inside the file, importing a backup **can't replace your current account**. At worst it updates the account the backup belongs to, creating it if needed. After importing, the account shows up in the **Accounts** list, ready to switch to.

**Import never switches accounts**, from the panel or the `account import` command. A game in progress carries on under the account it belongs to. To play as the imported account, switch to it with `account switch <name>` or from the Accounts panel. A backup of the account you are using goes straight into it, and anything that account earned since its last autosave is kept.

By default the merge keeps the best of both:

- **Theme unlocks** are combined, so importing an old backup never *removes* a theme you've unlocked since.
- **Badges** are combined: you keep every badge either copy holds. For a badge both hold, the earlier date is kept.
- **Lifetime stats** and the counts behind your badges take the higher of the two values, so your bests never go down. Counts are never added together, so importing your own backup twice changes nothing.
- **Highest Age Ever** takes the later age of the two.
- **Active theme** keeps your current choice if you have one, otherwise it takes the backup's.

To overwrite that account's progress entirely with the backup instead, add `replace`:

```
account import /path/to/my-ageforge-backup.json replace
```

An export made by a version from before badges has no badges in it. Importing one never removes badges, with or without `replace`: it says nothing about them, so the account's badges stay as they are.

If the file is missing or has been tampered with, the import is refused with an error and your accounts are left unchanged. A backup of an account flagged **modified** keeps the flag, and passes it to the account it lands in.

> **No server, no copy elsewhere.** You can only recover progress you exported. If a machine's `data/` folder is gone and you never exported, that account's earned progress is gone too; there is no cloud copy to pull back.

---

## Backups

A **backup** is a full copy of an account's slot on disk: its `account.json`, `badges.json` and `settings.json` **plus a recursive copy of that slot's `saves/` folder**. It holds more than an [export](#exporting-amp-importing-accounts). An export writes only the account-wide progress (unlocks, lifetime stats, badges, prefs) into a single file and carries **no saves**, while a backup copies the whole slot, your games included.

The game makes a backup at three points:

- **Before any wipe.** Both the Accounts panel wipe (the type-the-name confirm) and the active-account wipe copy the slot first, so a wipe always leaves a recoverable copy behind. (The wipe still goes ahead if the backup fails.)
- **On export.** `account export` and the panel's Export (`e`) take a full backup right after writing the export file, and report its path.
- **On demand.** Run `account backup` to back up the **active** account, or press `b` in the **Accounts** panel to back up the **highlighted** one.

Backups live **under the data folder**, beside (not inside) `data/accounts/`:

```
data/
└── backups/
    └── <name>-<id8>-<timestamp>/
        ├── account.json
        ├── badges.json
        ├── settings.json
        └── saves/
```

`<name>` is the display name made safe for a file name, `<id8>` is the first 8 characters of the account ID, and `<timestamp>` is `YYYYMMDD-HHMMSS`. Because backups sit outside `data/accounts/`, **wiping an account's slot never deletes its backups.**

> **The game keeps the last 10 backups per account.** After each new backup, older backups for *that* account are deleted so only the **10 most recent** remain. Other accounts' backups are never touched.

### Restoring from a backup

There's no restore command. A backup is just files, so you put them back by hand: copy the backup folder's `account.json`, `badges.json`, `settings.json` (if it has one) and `saves/` back into that account's slot at `data/accounts/<id>/`. The `<id8>` in the backup folder name is the start of the full `<id>`; the full ID is the slot's directory name under `data/accounts/`.

---

## The recovery code

Run the `account` command at the `>` prompt to see the active account's short ID and its **recovery code**:

```
account
```

The recovery code looks like this:

```
AGEF-7Q2K-9X4M-ZJ31-…
```

It's an uppercase string in 4-character groups separated by dashes, about 41 characters long including the `AGEF-` prefix. It uses [Crockford base32](https://www.crockford.com/base32.html), which leaves out the easily confused letters `I`, `L`, `O` and `U`, so it copies cleanly when you write it on paper or read it aloud.

The code holds **only your account ID plus a checksum**.

- **The checksum catches typos.** If you mistype the code, recovery is refused with a "checksum failed" message instead of restoring the wrong account.
- **It is not a password.** The checksum guards against typos, not against other people. The code is an identifier, not a credential or a secret, and there is nothing to steal: account state is cosmetic.

### Identity vs progress: what the code does and doesn't carry

| | Restored by the recovery code? |
|---|---|
| Your **identity** (account ID) | **Yes** |
| Your earned **progress** (theme unlocks, lifetime stats, badges) | **No** |

The recovery code restores your **identity** across machines and reinstalls. It is **separate from a progress export and carries no progress**; the code is short because it holds only the identity.

To carry your earned progress between machines, use an account **export** (see [Exporting & importing accounts](#exporting-amp-importing-accounts)). The two work together: the code restores *who you are*, the export restores *what you've earned*. A full move to a new machine uses **both**.

---

## Lifetime stats & badges

Some progress is **account-wide**: it builds up across *every* game you play on that account and *every* prestige, not just your current run. It lives on the account, separate from the per-save Statistics that reset when you start over or prestige.

| Lifetime stat | What it tracks |
|---|---|
| **Total Prestiges** | Every prestige you've completed on this account, across all its games |
| **Civilizations Started** | Every civilization you've begun on this account: each new game, and the one that follows each prestige and each Succumb. Loading a game starts nothing |
| **Highest Age Ever** | The furthest age any of this account's civilizations has reached. It only goes *up* |

The **Stats** panel shows all three under **Lifetime (account)**. (Total Prestiges and Highest Age Ever also show for each account in the [Accounts panel](#the-accounts-panel), so you can compare your civilizations at a glance.)

The account also records each prestige under the age it was made from (`prestiges_by_age` in the stats of `account.json`), so an early taste (a prestige from the Medieval Age to the Atomic Age) can be told from a full run (the Modern Age or deeper); see [Early Tastes and Full Runs](prestige.md#early-tastes-and-full-runs). No panel shows it, and Total Prestiges still counts every prestige, tastes included. Prestiges made before the account kept this record count only in the total. An export carries it, and an import keeps the higher count for each age.

These stats update the moment you prestige or advance into a new age, and are saved to your account with the next autosave (and on a clean exit, or when you switch accounts), so a fresh prestige is never lost. They count for the account the game belongs to: a game started or loaded under one account never adds to another's.

### Badges

A **badge** is a permanent mark on your account for something you did in a game: reaching an age, building enough of something across all your runs, or pulling off something specific in one run. A badge is earned once and kept for good. It stays with your account through every prestige and new game, and it travels in an export.

Badges are separate from [milestones](milestones.md). A milestone belongs to one run: it pays a reward inside that run and starts over with the next. A badge belongs to the account and gives nothing inside a run, so the same game plays the same way whatever the account has earned.

**When you earn one**, the game shows a toast and writes one line to the log. The toast carries the badge in small, in its tier's colors, then its name, its tier and what it was for.

<figure class="screen" data-screen="badge-toast"><figcaption>A badge earned: the toast sits in the bar under the status line, and the log keeps the same words.</figcaption></figure>

**To see them**, open the badge case:

```
badges
```

#### The badge case

The case is a panel over the whole screen but the command bar. It shows every badge the account holds or can earn as a small badge in a grid, by family, with the selected badge at full size beside it.

<figure class="screen" data-screen="badge-case"><figcaption>The badge case. The top bar counts what is earned and gives the score and the title it holds; the line at the foot says how far the next rung is.</figcaption></figure>

A badge looks like what it is:

| It looks like | It is |
|---|---|
| A rounded box with a notch under it | **Bronze**, earned |
| A box with double lines | **Silver**, earned |
| A frame of half blocks | **Gold**, earned |
| The same frame with a star at each corner | **Platinum**, earned |
| A solid ring of blocks | **Legendary**, earned |
| A dashed box with a `?` | In sight, not earned yet. The detail says what it asks for |
| A solid slab | Hidden: a secret badge, or one about something you have not come across yet. A slab with `…` in it stands for all the hidden badges of its family, however many there are |
| A strike through the badge | Earned in a modified game. It adds no points |

The frame is the tier, so tiers are told apart by shape as well as by color: in every theme, in the themes that draw in one ink, and in the plain glyph set (`map glyphs ascii`), where the same shapes are drawn with ordinary characters. At full size a tier grows with its worth, from a three-row bronze box to a nine-row legendary ring. The glyph in the middle is the badge's emblem, and every emblem is a symbol the [Map](map.md) already draws: a lineage wears its map symbol, an age the town center of its era.

- **Tabs** run along the top: All, one for each family, and **Next**, which lists the badges you are closest to, nearest first. `Tab` steps through them.
- **Ladders** sit on a line of their own, lowest rung first, with the ladder's name and how far its next rung is beside them.
- **`Enter`** opens the selected badge's detail: the badge at full size with everything the case knows about it. `badges <name>` opens a detail directly.

<figure class="screen" data-screen="badge-detail"><figcaption>One badge's detail: its tier, rarity and points, what it asked for, when it was earned and in which game, and its place on its ladder.</figcaption></figure>

Platinum badges glint, legendary badges turn through their colors with a band of light crossing the rim, and a few special badges have a drawing of their own that moves. Nothing else in the case moves, and `motion off` holds all of it still (see [Motion](themes.md#motion)).

The keys are listed under [Badges](commands.md#badges) on the commands page and in the Help panel.

#### Titles

Your badge points add up to a **score**, and the score holds a **title**. The badge case shows it in its top bar, with the next title and what it asks for when the window is wide enough.

| Title | Score |
|---|---|
| Settler | 0 |
| Headman | 250 |
| Magistrate | 1,000 |
| Sovereign | 3,000 |
| Paragon | 6,000 |
| Eternal | 10,000 |
| Completionist | Every badge that counts, none of them earned in a modified game |

Past Settler, the status bar shows the title beside your account's name when the window has room for it. A title changes nothing in a game.

#### The plain list

The **Stats** panel (`stats`) lists the same badges as text under **Lifetime (account)**, and so does:

```
account badges
```

Both list every badge you may see, by group:

| Mark | Meaning |
|---|---|
| ★ | Earned |
| ☆ | Not earned yet. The line says what it asks for, and a badge that counts across runs shows how far along you are |
| ? | A secret badge. It shows a one-line hint and nothing else until you earn it |

The first line counts them: for example `3 of 9 earned, 15 points, ??? hidden`. A badge about something you have not come across yet (an age you have not seen named, for one) stays out of the list and out of the "of" number until you get there. The game does not say how many are hidden until an account has reached the last age.

Each badge has a tier, and a tier is worth points:

| Tier | Points |
|---|---|
| Bronze | 5 |
| Silver | 10 |
| Gold | 25 |
| Platinum | 50 |
| Legendary | 100 |

A few things about how badges are counted:

- **Building counts ignore selling and rebuilding.** A badge that counts buildings across your runs counts a copy only when it takes that building past the most you have built of it in the run, so selling a building and building it again adds nothing.
- **The four old achievements are badges now.** First Prestige, Serial Reincarnator, Age of Iron and Into the Modern Age carry over, and an account that had them keeps them.
- **The developer console does not block badges.** The developer console is a testing tool. A game it has changed is marked in its save, and the log says so once, but it still records to the account like any other game: if a badge's condition is met, the badge is earned. Unlocking the console earns a badge of its own.
- **A modified game marks what it earns.** A badge earned in a save that was edited outside the game is listed as earned in a modified game and adds no points. The same goes for a badge earned on an account whose `account.json` was edited, and for every badge in a `badges.json` that was edited.
- **A badge can give a theme.** Two [themes](themes.md#ambient-effects), Source and Glitch, come from secret badges. Earning the badge unlocks the theme on the account, for good.

### Where badges are stored

Badges are kept in their own file, `badges.json`, beside `account.json` in the account's slot (`data/accounts/<id>/`). It holds the badges you have earned, the counts behind the ones that count across runs, and the days you have played. The game writes it the first time the account has a badge to keep.

`account.json` did not change for badges, on purpose. If you also run a version of the game from before badges, it reads and writes `account.json` as it always did and leaves `badges.json` alone, so going back and forth between versions cannot mark an account as modified or cost it a badge. When this version next opens the account, it adds whatever the other version's play has proved since (an age reached, prestiges made) to your badges.

The file is signed, and the signature covers the account ID:

- A `badges.json` that was edited by hand, or copied in from another account, is flagged, and its badges are listed as earned in a modified game. That flag belongs to `badges.json` only; it never marks `account.json` as modified.
- If `badges.json` is deleted or cannot be read, the game rebuilds what `account.json` proves (ages reached, prestiges made). Dates, counts and badges with no record there are gone, so keep an [export](#exporting-amp-importing-accounts) or a [backup](#backups).
- An export carries the badge file inside it, and a backup copies it, so both bring your badges back.

### The settings file

The motion setting (`motion on` and `motion off`) is kept in a third file in the slot, `settings.json`. It is there for the same reason `badges.json` is: `account.json` keeps the shape older versions of the game know, so a newer setting goes beside it. The file holds a display preference and nothing you earn, so it is plain text and unsigned, and the game writes it only once you change the setting. If it is missing or cannot be read, motion is on. A [backup](#backups) copies it; an export does not carry it.

---

## Restoring on a new machine

There are **two ways** to restore your identity, because your name and your recovery code both lead to the same account ID. Either way, restoring your **identity** does not bring your earned progress with it; for that you also need an [account export](#exporting-amp-importing-accounts).

### The simplest way: re-enter your name

On a fresh machine, the first-run prompt asks you to name your account. **Type the exact same name you used before** and you get the exact same account ID, with no code needed. (Casing and surrounding spaces don't matter, but a different name is a different identity.)

### With the recovery code

On the new machine (or after a reinstall), run:

```
account recover AGEF-7Q2K-9X4M-ZJ31-…
```

This restores the identity in the code and switches to it. The account lands in **its own slot**:

- If that account is **already on this machine**, the game opens it as it is, progress included.
- If not, the game creates it with the ID only (no unlocks or stats; bring those back with an [import](#exporting-amp-importing-accounts)).

**Recovering never overwrites an account**, including the one you are using: that account keeps everything in its own slot, and you can switch back to it with `account switch <name>` or from the Accounts panel. If the account you are using holds any progress (theme unlocks, badges or lifetime stats), the command first says what it holds and asks you to run `account recover <code> confirm`. During a game, the game is saved to its account first and you go back to the main menu. Recovering the code of the account you are using does nothing.

**Recovery forgives copying mistakes:**

- Lowercase is fine.
- Extra spaces are ignored.
- The easily confused `I`/`L` and `1`, and `O` and `0`, are corrected automatically.

So if you wrote it down by hand and your `0` looks like an `O`, it still works.

---

## Wiping an account

If you want a clean slate for an account, with none of its old unlocks, stats or badges, you can **wipe** it.

**Wipe Account** is in the **Accounts panel** (press `w` on the highlighted account). It is not a typed command: deleting an account is permanent, so it sits behind a confirm.

1. **Read the warning.** It explains exactly what's about to happen.
2. **Type the account name.** You must type that account's display name *exactly*. Anything else (or pressing Esc) cancels with no changes.

On confirmation, the wipe **permanently deletes** that account's:

- **identity** (the account name and derived ID),
- **theme unlocks**,
- **lifetime stats**,
- **badges**, and
- **every save** in its slot.

**This cannot be undone in the game**, and no server keeps a copy. The game [backs up](#backups) the slot, saves included, to `data/backups/` first; restoring that copy is manual. Otherwise the old identity comes back only if you wrote down its [recovery code](#the-recovery-code) beforehand, and its earned progress only if you [exported it](#exporting-amp-importing-accounts) first.

> Typing `account wipe` at the `>` prompt doesn't wipe anything. It tells you where to find the wipe in the Accounts panel, the only place it happens.

---

## Recovery in short

With no server, **you recover your identity from a short code you write down**. **You recover progress only if you exported it first.** If a machine's `data/` folder is gone and the progress was never backed up, no server holds a copy. The recovery code restores who you are, not what you've earned.

That is also why the code is safe to share or lose: it identifies a cosmetic, local account and unlocks nothing worth guarding.

---

## See also

- [All Commands](commands.md): the full `account` command reference
- [Saving & Loading](saving-and-loading.md): how your *game saves* are stored per account under `data/accounts/<id>/saves/`
