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
        └── saves/               # that account's game saves
```

The `data/active-account` pointer records which account is current. Each `data/accounts/<account_id>/` slot holds that account's `account.json` and its own `saves/` folder, so saves are **per-account** and never mix between accounts. (See [Saving & Loading](saving-and-loading.md) for the save layout.)

### Upgrading from an older version

Older builds kept a single account at the top level (a flat `data/account.json` with saves in `data/saves/`). If the game finds that layout on first launch, it moves your account and its saves into their own `data/accounts/<id>/` slot and makes that account active. **Nothing is deleted**, and you don't have to do anything; your civilizations and unlocks come across intact. Just before the move, the game also copies your old flat data into `data/backups/pre-migration-<timestamp>/`, so your original files stay recoverable.

---

## Your account

The **first time you launch the game**, AgeForge asks you to **name your account**. The prompt comes up before the main menu, filled in with a suggested empire-style name you can keep, edit or reroll. When you submit, the game **asks you to confirm the name before creating the account**. The name *is* your identity, so a typo would quietly create a different, empty account. Choose **Create** to lock it in, or **Re-type** to fix it. The confirmation also reminds you to **re-enter the name exactly** to restore your account on another device. Whatever you settle on becomes your account's display name and your **identity**.

> **Your name *is* your identity.** The account ID is **derived from your name**: `sha256(normalize(name))`, taking the first 16 bytes as a 32-character hex ID. Normalizing lowercases the name, trims it and collapses internal spacing before hashing, so `Imperium`, `imperium`, and the same name with stray surrounding spaces all give the *same* ID. The display name keeps your original casing. Because the ID comes from the name, **re-entering the exact same name on a new machine gives the exact same account ID**. That is the simplest way to restore your identity (see [Restoring on a new machine](#restoring-on-a-new-machine)).

The account holds two distinct things:

| Part | What it is |
|---|---|
| **Identity** | Your chosen name and the account ID derived from it |
| **Data** | Your earned account-wide progress: theme unlocks, lifetime stats, achievements, prefs |

The split matters because the two halves are recovered very differently (see below). Your **identity** is carried by either your account name *or* the recovery code (both point at the same ID). The **data** is backed up separately with an account **export** (see [Exporting & importing accounts](#exporting--importing-accounts)).

> **You choose the name once.** Because the ID comes from the name, picking a *different* name later creates a *different* identity; it does not rename the account. For that reason there is no in-game rename, so choose a name you're happy to keep. (If you want a second civilization to play in parallel, create a **new account**; see the Accounts panel below.)

---

## The Accounts panel

The main menu has an **Accounts** entry that opens a full-window panel listing **every local account** on this machine. For each one it shows:

- the **display name**,
- a **short ID** (the first part of the account ID),
- the **highest age** that account has ever reached,
- its **total prestiges**,
- a **current** marker on the account that's active right now, and
- a **modified** flag if that account's file was edited outside the game.

From the panel:

| Key | Action |
|---|---|
| `Enter` | **Switch** to the highlighted account (it becomes active; its saves show in the Load Game browser) |
| `n` | **New account**: name and create a fresh account alongside your existing ones |
| `e` | **Export** the highlighted account to a signed backup file |
| `b` | **Backup** the highlighted account: a full copy of its slot (`account.json` + `saves/`) to `data/backups/` (see [Backups](#backups)) |
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
| **Account export** | The account's ID, name, theme unlocks, lifetime stats, achievements and prefs | Save the export file somewhere safe |
| **Recovery code** | Identity only (the account ID) | Write down the short `AGEF-…` string |

### Exporting

From the Accounts panel press `e`, or run:

```
account export
```

This writes a **signed backup of the active account**. The file is **bound to its account ID**: it carries the account's ID, name, theme unlocks, lifetime stats, achievements and prefs, and **the signature covers the ID**, so a backup can't be passed off as a different account's. By default it's written as `account-<id8>-export.json` inside that account's own slot (`data/accounts/<id>/`). To choose your own location, pass a path:

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

**Import never overwrites a *different* account.** Because it goes by the ID inside the file, importing a backup **can't replace your current account**. At worst it updates the account the backup belongs to, creating it if needed. After importing, the account shows up in the **Accounts** list, ready to switch to; the `account import` command also **switches to it** for you.

By default the merge keeps the best of both:

- **Theme unlocks** are combined, so importing an old backup never *removes* a theme you've unlocked since.
- **Achievements** are combined.
- **Lifetime stats** take the higher of the two values, so your bests never go down.
- **Active theme** keeps your current choice if you have one, otherwise it takes the backup's.

To overwrite that account's progress entirely with the backup instead, add `replace`:

```
account import /path/to/my-ageforge-backup.json replace
```

If the file is missing or has been tampered with, the import is refused with an error and your accounts are left unchanged.

> **No server, no copy elsewhere.** You can only recover progress you exported. If a machine's `data/` folder is gone and you never exported, that account's earned progress is gone too; there is no cloud copy to pull back.

---

## Backups

A **backup** is a full copy of an account's slot on disk: its `account.json` **plus a recursive copy of that slot's `saves/` folder**. It holds more than an [export](#exporting--importing-accounts). An export writes only the account-wide progress (unlocks, lifetime stats, achievements, prefs) into a single file and carries **no saves**, while a backup copies the whole slot, your games included.

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
        └── saves/
```

`<name>` is the display name made safe for a file name, `<id8>` is the first 8 characters of the account ID, and `<timestamp>` is `YYYYMMDD-HHMMSS`. Because backups sit outside `data/accounts/`, **wiping an account's slot never deletes its backups.**

> **The game keeps the last 10 backups per account.** After each new backup, older backups for *that* account are deleted so only the **10 most recent** remain. Other accounts' backups are never touched.

### Restoring from a backup

There's no restore command. A backup is just files, so you put them back by hand: copy the backup folder's `account.json` and `saves/` back into that account's slot at `data/accounts/<id>/`. The `<id8>` in the backup folder name is the start of the full `<id>`; the full ID is the slot's directory name under `data/accounts/`.

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
| Your earned **progress** (theme unlocks, lifetime stats, achievements) | **No** |

The recovery code restores your **identity** across machines and reinstalls. It is **separate from a progress export and carries no progress**; the code is short because it holds only the identity.

To carry your earned progress between machines, use an account **export** (see [Exporting & importing accounts](#exporting--importing-accounts)). The two work together: the code restores *who you are*, the export restores *what you've earned*. A full move to a new machine uses **both**.

---

## Lifetime stats & achievements

Some progress is **account-wide**: it builds up across *every* game you play on that account and *every* prestige, not just your current run. It lives on the account, separate from the per-save Statistics that reset when you start over or prestige.

| Lifetime stat | What it tracks |
|---|---|
| **Total Prestiges** | Every prestige you've completed on this account, across all its games |
| **Highest Age Ever** | The furthest age any of this account's civilizations has reached. It only goes *up* |

(These two also show for each account in the [Accounts panel](#the-accounts-panel), so you can compare your civilizations at a glance.)

**Achievements** are one-time, account-wide badges. Once unlocked, they stay unlocked; they stay with your account and travel in an export. The current set:

| Achievement | Unlocks when |
|---|---|
| **First Prestige** | You complete your first prestige |
| **Serial Reincarnator** | You reach 10 lifetime prestiges |
| **Age of Iron** | Any civilization reaches the Iron Age |
| **Into the Modern Age** | Any civilization reaches the Modern Age |

**Where to see them:** open the **Stats** panel (`stats`). Below the per-run Statistics there's a **Lifetime (Account)** section with your total prestiges, highest age ever reached, and the achievements you've unlocked. The game doesn't announce an achievement with a pop-up, so check the Stats panel to see what's unlocked.

These stats update the moment you prestige or advance into a new age, and are saved to your account with the next autosave (and on a clean exit), so a fresh prestige is never lost.

---

## Restoring on a new machine

There are **two ways** to restore your identity, because your name and your recovery code both lead to the same account ID. Either way, restoring your **identity** does not bring your earned progress with it; for that you also need an [account export](#exporting--importing-accounts).

### The simplest way: re-enter your name

On a fresh machine, the first-run prompt asks you to name your account. **Type the exact same name you used before** and you get the exact same account ID, with no code needed. (Casing and surrounding spaces don't matter, but a different name is a different identity.)

### With the recovery code

On the new machine (or after a reinstall), run:

```
account recover AGEF-7Q2K-9X4M-ZJ31-…
```

This restores the identity in the code. If the account currently on this machine already has theme unlocks, the command warns you first: recovering replaces the local identity, and the code doesn't carry those unlocks. Export them first if you want to keep them, then run `account recover <code> confirm` to go ahead.

**Recovery forgives copying mistakes:**

- Lowercase is fine.
- Extra spaces are ignored.
- The easily confused `I`/`L` and `1`, and `O` and `0`, are corrected automatically.

So if you wrote it down by hand and your `0` looks like an `O`, it still works.

---

## Wiping an account

If you want a clean slate for an account, with none of its old unlocks, stats or achievements, you can **wipe** it.

**Wipe Account** is in the **Accounts panel** (press `w` on the highlighted account). It is not a typed command: deleting an account is permanent, so it sits behind a confirm.

1. **Read the warning.** It explains exactly what's about to happen.
2. **Type the account name.** You must type that account's display name *exactly*. Anything else (or pressing Esc) cancels with no changes.

On confirmation, the wipe **permanently deletes** that account's:

- **identity** (the account name and derived ID),
- **theme unlocks**,
- **lifetime stats**, and
- **achievements**.

**This cannot be undone**, and no server keeps a copy. The old identity comes back only if you wrote down its [recovery code](#the-recovery-code) beforehand, and its earned progress only if you [exported it](#exporting--importing-accounts) first.

> Typing `account wipe` at the `>` prompt doesn't wipe anything. It tells you where to find the wipe in the Accounts panel, the only place it happens.

---

## Recovery in short

With no server, **you recover your identity from a short code you write down**. **You recover progress only if you exported it first.** If a machine's `data/` folder is gone and the progress was never backed up, no server holds a copy. The recovery code restores who you are, not what you've earned.

That is also why the code is safe to share or lose: it identifies a cosmetic, local account and unlocks nothing worth guarding.

---

## See also

- [All Commands](commands.md): the full `account` command reference
- [Saving & Loading](saving-and-loading.md): how your *game saves* are stored per account under `data/accounts/<id>/saves/`
