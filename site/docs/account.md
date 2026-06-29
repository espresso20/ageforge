# Account & Recovery

AgeForge is a **local terminal game**. There is no server, no login screen, no cloud — everything lives on your own machine under `./data/`, relative to the directory you launch the game from. Your accounts are part of that: small files on disk that give your civilizations a stable identity and hold your earned meta-progression.

The game supports **multiple local accounts**, each in its own slot on disk, with one active at a time. This page explains how accounts and their data are laid out, how to switch between them, how to back them up (export) and bring them in (import), what the recovery code does (and, just as important, what it doesn't), and how to carry an identity to a new machine.

---

## Multiple accounts, one at a time

You can keep **several accounts side by side** on the same machine. Each is its own slot on disk, and exactly **one is active** at any moment — the active account is the one whose saves you see in the [Load Game](saving-and-loading.md) browser and whose meta-progression is in play. Switching accounts swaps both: a different identity, and a different set of saves.

Each account lives under its own directory:

```
data/
├── active-account               # pointer to the currently active account
└── accounts/
    └── <account_id>/
        ├── account.json         # that account's identity + meta-progression
        └── saves/               # that account's game saves
```

The `data/active-account` pointer records which account is current. Each `data/accounts/<account_id>/` slot holds that account's `account.json` and its own `saves/` folder — so saves are **per-account** and never mix between accounts. (See [Saving & Loading](saving-and-loading.md) for the save layout.)

### Upgrading from an older version

If you're coming from an older build that kept a single account at the top level (a flat `data/account.json` with saves in `data/saves/`), the game **migrates you automatically on first launch** — and **non-destructively**. Your existing account and its saves are moved into their own `data/accounts/<id>/` slot, that account is set as active, and **nothing is deleted**. You don't have to do anything; your civilizations and unlocks come across intact. As a one-time safety net, the game also snapshots your old flat data into `data/backups/pre-migration-<timestamp>/` just before the move, so your original state stays recoverable if you ever need it.

---

## Your account

The **first time you launch the game**, AgeForge asks you to **name your account**. The prompt comes up before the main menu, pre-filled with a suggested empire-style name you can keep, edit, or reroll. When you submit, the game **asks you to confirm the name before creating the account** — because the name *is* your identity, a typo would otherwise silently mint a different, empty account. Choose **Create** to lock it in, or **Re-type** to fix it. The confirmation also reminds you to **re-enter the name exactly** to restore your account on another device. Whatever you settle on becomes your account's display name — and, just as importantly, your **identity**.

> **Your name *is* your identity.** The account ID is **derived from your name** — specifically `sha256(normalize(name))`, taking the first 16 bytes as a 32-character hex ID. "Normalize" means the name is lowercased, trimmed, and internal spacing collapsed before hashing, so `Imperium`, `imperium`, and the same name with stray surrounding spaces all derive the *same* ID. The display name keeps your original casing. Because the ID is name-derived, **re-entering the exact same name on a new machine regenerates the exact same account ID** — that's the simplest way to restore your identity (see [Restoring on a new machine](#restoring-on-a-new-machine)).

The account holds two distinct things:

| Part | What it is |
|---|---|
| **Identity** | Your chosen name and the account ID derived from it |
| **Data** | Your earned meta-progression — theme unlocks, lifetime stats, achievements, prefs |

The split matters, because the two halves are recovered very differently (see below). Your **identity** is carried by either your account name *or* the recovery code (both point at the same ID); the **data** is backed up separately with an account **export** (see [Exporting & importing accounts](#exporting--importing-accounts)).

> **Naming is chosen-once.** Because the ID is derived from the name, picking a *different* name later mints a *different* identity — it does not rename the account in place. There's no in-game rename for this reason. Choose a name you're happy to keep. (If you want a second civilization to play in parallel, that's exactly what a **new account** is for — see the Accounts panel below.)

---

## The Accounts panel

The main menu has an **Accounts** entry that opens a full-screen panel listing **every local account** on this machine. For each one it shows:

- the **display name**,
- a **short ID** (the first chunk of the account ID),
- the **highest age** that account has ever reached,
- its **total prestiges**,
- a **current** marker on the account that's active right now, and
- a **modified** flag if that account's file was tampered with (edited outside the game).

From the panel:

| Key | Action |
|---|---|
| `Enter` | **Switch** to the highlighted account (it becomes active; its saves load in the Load Game browser) |
| `n` | **New account** — name and create a fresh account alongside your existing ones |
| `e` | **Export** the highlighted account to a signed backup file |
| `b` | **Backup** the highlighted account — a full snapshot of its slot (`account.json` + `saves/`) to `data/backups/` (see [Backups](#backups)) |
| `i` | **Import** an account from a backup file |
| `r` | Show a **recovery code** for restoring an identity (see [The recovery code](#the-recovery-code)) |
| `w` | **Wipe** the highlighted account (permanent — behind a type-the-name confirm) |
| `Esc` | **Back** to the main menu |

This is the hub for managing accounts: it's where you switch between civilizations, spin up new ones, and back them up. (**Wipe Account** used to sit directly on the main menu — it now lives here, behind the same type-your-name confirm described under [Wiping an account](#wiping-an-account).)

---

## Exporting & importing accounts

Each account has a backup of its **data** that is separate from its recovery code. The export carries your earned progress; the recovery code carries only your identity. They do different jobs and you keep them differently:

| Backup | What it carries | How to keep it |
|---|---|---|
| **Account export** | The account's id, name, theme unlocks, lifetime stats, achievements, and prefs | Save the export file somewhere safe |
| **Recovery code** | Identity only (the account ID) | Write down the short `AGEF-…` string |

### Exporting

From the Accounts panel press `e`, or run:

```
account export
```

This writes a **signed backup of the active account**. The blob is **bound to its account ID**: it carries the account's id, name, theme unlocks, lifetime stats, achievements, and prefs, and **the signature covers the id**, so a backup can't be quietly re-attributed to a different account. By default it's written as `account-<id8>-export.json` inside that account's own slot (`data/accounts/<id>/`). To choose your own location, pass a path:

```
account export /path/to/my-ageforge-backup.json
```

Export after any big unlock. It's a one-shot snapshot — it does **not** auto-update as you keep playing, so re-export when you've earned something you'd hate to lose.

### Importing

On the same or a new machine, bring a backup in from the Accounts panel with `i`, or:

```
account import /path/to/my-ageforge-backup.json
```

Import is keyed by the **account ID embedded in the backup**, and it always lands in **that account's own slot**:

- If **no account with that ID exists** locally, import **creates** it from the backup.
- If that account **already exists** locally, import **merges** into it (the default — see below).

Crucially, **import never overwrites a *different* account.** Because it's keyed by the embedded ID, importing a backup **can't clobber your current account** — at worst it updates the account the backup belongs to, creating it if need be. After importing, the account shows up in the **Accounts** list, ready to switch to; the `account import` command also **switches to it** for you.

By default the merge is the **safe** choice:

- **Theme unlocks** are unioned — importing an old backup never *removes* a theme you've unlocked since.
- **Achievements** are unioned.
- **Lifetime stats** take the higher of the two values, so your bests never regress.
- **Active theme** keeps your current choice if you have one, otherwise adopts the backup's.

To overwrite that account's progress wholesale with the backup instead, add `replace`:

```
account import /path/to/my-ageforge-backup.json replace
```

If the file is missing or has been tampered with, the import is rejected with a clear error and your accounts are left unchanged — corrupt data never gets imported.

> **No server, so no magic.** Progress recovery only works if you exported it first. If a machine's `./data/` is gone and you never exported, that account's earned progress is gone too — there's no cloud copy to pull back.

---

## Backups

A **backup** is a full on-disk snapshot of an account's slot: a copy of its `account.json` **plus a recursive copy of that slot's `saves/` folder**. It's a heavier, more complete thing than an [export](#exporting--importing-accounts) — an export serializes only the account's meta-progression (unlocks, lifetime stats, achievements, prefs) into a single blob and carries **no saves**, whereas a backup captures the whole slot, your games included.

Backups happen on **three triggers**:

- **Before any wipe.** Both the Accounts-panel wipe (the type-the-name confirm) and the active-account wipe snapshot the slot first, so a wipe always leaves a recoverable copy behind. (The wipe still proceeds even if the backup somehow fails.)
- **On export.** `account export` and the panel's Export (`e`) take a full backup right after writing the export blob, and report its path.
- **On demand.** Run `account backup` to snapshot the **active** account, or press `b` in the **Accounts** panel to snapshot the **highlighted** one.

Backups live **under the data root**, alongside (not inside) `data/accounts/`:

```
data/
└── backups/
    └── <name>-<id8>-<timestamp>/
        ├── account.json
        └── saves/
```

`<name>` is the filesystem-safe display name, `<id8>` is the first 8 characters of the account ID, and `<timestamp>` is `YYYYMMDD-HHMMSS`. Because backups sit outside `data/accounts/`, **wiping an account's slot never deletes its backups.**

> **Retention: the last 10 per account.** After each new backup, older backups for *that* account are pruned automatically — only the **10 most recent** are kept. Other accounts' backups are never touched.

### Restoring from a backup

There's no restore command — a backup is just files, so you put them back by hand. Copy the contents of a backup folder — its `account.json` and its `saves/` — back into that account's slot at `data/accounts/<id>/`. The `<id8>` in the backup folder name is the short prefix of the full `<id>`; the full ID is the slot directory name under `data/accounts/`.

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

It's an uppercase string, grouped into 4-character chunks separated by dashes, around 41 characters long including the `AGEF-` prefix. It's encoded in [Crockford base32](https://www.crockford.com/base32.html), which deliberately omits the ambiguous letters `I`, `L`, `O`, and `U` — so it transcribes cleanly when you write it on paper or read it aloud.

The code encodes **only your account ID plus a checksum**. Nothing else.

- **The checksum is a typo guard.** If you mistype the code, recovery is rejected with a clear "checksum failed" message instead of silently restoring the wrong account.
- **It is not a password.** The checksum guards against typos, not against other people. The code is a convenience identifier, not a credential or a secret — there is nothing to steal. Account state is cosmetic.

### Identity vs progress — what the code does and doesn't carry

This is the part to read carefully.

| | Restored by the recovery code? |
|---|---|
| Your **identity** (account ID) | **Yes** |
| Your earned **progress** (theme unlocks, lifetime stats, achievements) | **No** |

The recovery code restores your **identity** across machines and reinstalls. It is **separate from a progress export and carries no progress** — the code is small precisely because it carries identity, not data.

To carry your earned progress between machines, use an account **export** (see [Exporting & importing accounts](#exporting--importing-accounts)). The two are complementary: the code resurrects *who you are*, the export resurrects *what you've done*. A full transfer to a new machine uses **both**.

---

## Lifetime stats & achievements

Some progress is **account-wide and cross-save** — it accumulates across *every* game you play on that account and *every* prestige, not just your current run. This lives on the account, separate from the per-save Statistics that reset when you start over or prestige.

| Lifetime stat | What it tracks |
|---|---|
| **Total Prestiges** | Every prestige you've completed on this account, across all its games |
| **Highest Age Ever** | The furthest age any of this account's civilizations has reached — it only ever goes *up* |

(These two are also surfaced per-account in the [Accounts panel](#the-accounts-panel), so you can compare your civilizations at a glance.)

**Achievements** are one-time, account-wide badges. Once unlocked, they stay unlocked — they ride with your account and travel in an export. The current set:

| Achievement | Unlocks when |
|---|---|
| **First Prestige** | You complete your first prestige |
| **Serial Reincarnator** | You reach 10 lifetime prestiges |
| **Age of Iron** | Any civilization reaches the Iron Age |
| **Into the Modern Age** | Any civilization reaches the Modern Age |

**Where to see them:** open the **Stats** overlay. Below the per-run Statistics there's a **Lifetime (Account)** section showing your total prestiges, highest age ever reached, and the achievements you've unlocked. (Reaching a milestone is recorded silently — there's no pop-up; check the Stats overlay to see what's unlocked.)

These stats update the moment you prestige or advance into a new age, and are saved to your account alongside the next autosave (and on a clean exit), so a fresh prestige is never lost.

---

## Restoring on a new machine

There are **two ways** to restore your identity, because your identity has two equivalent forms — your name and your recovery code both resolve to the same account ID. Either way, restoring **identity** does not bring your earned progress with it; for that you also need an [account export](#exporting--importing-accounts).

### The simplest way: re-enter your name

On a fresh machine, the first-run prompt asks you to name your account. **Type the exact same name you used before** and you'll regenerate the exact same account ID — no code required. (Remember normalization: casing and surrounding spaces don't matter, but a *meaningfully* different name is a different identity.)

### The precise way: recover from a code

Alternatively, on the new machine (or after a reinstall), run:

```
account recover AGEF-7Q2K-9X4M-ZJ31-…
```

This restores the identity encoded in the code.

**Recovery is lenient about transcription.** You don't have to get the formatting perfect:

- Lowercase is fine.
- Extra spaces are ignored.
- The ambiguous swaps `I`/`L` ↔ `1` and `O` ↔ `0` are corrected automatically.

So if you wrote it down by hand and your `0` looks like an `O`, it'll still resolve.

---

## Wiping an account

If you want a genuine clean slate for an account — none of its old unlocks, stats, or achievements — you can **wipe** it.

**Wipe Account** lives in the **Accounts panel** (press `w` on the highlighted account). It is deliberately *not* a plain typed command: deleting an account is permanent, so it sits behind a confirm.

1. **Confirm the stakes.** A warning explains exactly what's about to happen.
2. **Type the account name.** You must type that account's display name *exactly* — an exact match is the only thing that proceeds. Anything else (or pressing Esc) aborts with no changes.

On confirmation, the wipe **permanently deletes** that account's:

- **identity** (the account name and derived ID),
- **theme unlocks**,
- **lifetime stats**, and
- **achievements**.

**This cannot be undone** — there is no server backup. The old identity is only recoverable if you wrote down its [recovery code](#the-recovery-code) beforehand, and its earned progress only if you'd [exported it](#exporting--importing-accounts) first.

> Typing `account wipe` at the `>` prompt won't perform the wipe — it just points you at the Accounts panel, because the destructive action only happens behind the type-the-name confirm.

---

## The honest truth

With no server, **identity recovery is a short code you write down** — that's the whole trick. But **progress recovery requires you to have exported it first.** We can't make data appear from nothing: if a machine's `./data/` is gone and the progress was never backed up, there's no server holding a copy to pull it back from. The recovery code resurrects who you are, not what you've done.

That's also why the code is safe to share or lose track of — it's an identifier for a cosmetic, local account, not a key to anything worth guarding.

---

## See also

- [All Commands](commands.md) — the full `account` command reference
- [Saving & Loading](saving-and-loading.md) — how your *game saves* are stored per-account under `data/accounts/<id>/saves/`
