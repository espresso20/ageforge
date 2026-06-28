# Account & Recovery

AgeForge is a **local terminal game**. There is no server, no login screen, no cloud — everything lives on your own machine under `./data/`, relative to the directory you launch the game from. Your account is part of that: a small file on disk that gives your civilization a stable identity and holds your earned meta-progression.

This page explains what an account is, what the recovery code does (and, just as important, what it doesn't), and how to carry your identity to a new machine.

---

## Your account

The **first time you launch the game**, AgeForge asks you to **name your account**. The prompt comes up before the main menu, pre-filled with a suggested empire-style name you can keep, edit, or reroll. When you submit, the game **asks you to confirm the name before creating the account** — because the name *is* your identity, a typo would otherwise silently mint a different, empty account. Choose **Create** to lock it in, or **Re-type** to fix it. The confirmation also reminds you to **re-enter the name exactly** to restore your account on another device. Whatever you settle on becomes your account's display name — and, just as importantly, your **identity**. The account is then stored at `./data/account.json`. One account per machine, or more precisely one per data directory.

> **Your name *is* your identity.** The account ID is **derived from your name** — specifically `sha256(normalize(name))`, taking the first 16 bytes as a 32-character hex ID. "Normalize" means the name is lowercased, trimmed, and internal spacing collapsed before hashing, so `Imperium`, `imperium`, and the same name with stray surrounding spaces all derive the *same* ID. The display name keeps your original casing. Because the ID is name-derived, **re-entering the exact same name on a new machine regenerates the exact same account ID** — that's the simplest way to restore your identity (see [Restoring on a new machine](#restoring-on-a-new-machine)).

The account holds two distinct things:

| Part | What it is |
|---|---|
| **Identity** | Your chosen name and the account ID derived from it |
| **Data** | Your earned meta-progression — theme unlocks, lifetime stats, achievements |

The split matters, because the two halves are recovered very differently (see below). Your **identity** is carried by either your account name *or* the recovery code (both point at the same ID); the **data** is backed up separately with **progress export** (see [Backing up your progress](#backing-up-your-progress)).

> **Naming is chosen-once.** Because the ID is derived from the name, picking a *different* name later mints a *different* identity — it does not rename the account in place. There's no in-game rename for this reason. Choose a name you're happy to keep.

---

## The recovery code

Run the `account` command at the `>` prompt to see your account's short ID and its **recovery code**:

```
account
```

The recovery code looks like this:

```
AGEF-7Q2K-9X4M-ZJ31-…
```

It's an uppercase string, grouped into 4-character chunks separated by dashes, around 41 characters long including the `AGEF-` prefix. It's encoded in [Crockford base32](https://www.crockford.com/base32.html), which deliberately omits the ambiguous letters `I`, `L`, `O`, and `U` — so it transcribes cleanly when you write it on paper or read it aloud.

The code encodes **only your account ID plus a checksum**. Nothing else.

- **The checksum is a typo guard.** If you mistype the code, the import is rejected with a clear "checksum failed" message instead of silently creating the wrong account.
- **It is not a password.** The checksum guards against typos, not against other people. The code is a convenience identifier, not a credential or a secret — there is nothing to steal. Account state is cosmetic.

---

## What it restores (and what it doesn't)

This is the part to read carefully.

| | Restored by the recovery code? |
|---|---|
| Your **identity** (account ID) | **Yes** |
| Your earned **progress** (theme unlocks, lifetime stats, achievements) | **No** |

The recovery code restores your **identity** across machines and reinstalls. It does **not** restore your earned progress, because that progress is separate **data** — and the code is small precisely because it carries identity, not data.

To carry your progress between machines, back it up via **progress export** — see the next section.

---

## Lifetime stats & achievements

Some progress is **account-wide and cross-save** — it accumulates across *every* game you play and *every* prestige, not just your current run. This lives on your account, separate from the per-save Statistics that reset when you start over or prestige.

| Lifetime stat | What it tracks |
|---|---|
| **Total Prestiges** | Every prestige you've ever completed, across all games |
| **Highest Age Ever** | The furthest age any of your civilizations has reached — it only ever goes *up* |

**Achievements** are one-time, account-wide badges. Once unlocked, they stay unlocked — they ride with your account and travel in a progress export. The current set:

| Achievement | Unlocks when |
|---|---|
| **First Prestige** | You complete your first prestige |
| **Serial Reincarnator** | You reach 10 lifetime prestiges |
| **Age of Iron** | Any civilization reaches the Iron Age |
| **Into the Modern Age** | Any civilization reaches the Modern Age |

**Where to see them:** open the **Stats** overlay. Below the per-run Statistics there's a **Lifetime (Account)** section showing your total prestiges, highest age ever reached, and the achievements you've unlocked. (Reaching a milestone is recorded silently — there's no pop-up; check the Stats overlay to see what's unlocked.)

These stats update the moment you prestige or advance into a new age, and are saved to your account alongside the next autosave (and on a clean exit), so a fresh prestige is never lost.

---

## Backing up your progress

Your account has **two backups, and they do different jobs**:

| Backup | What it carries | How to keep it |
|---|---|---|
| **Recovery code** | Identity only (your account ID) | Write down the short `AGEF-…` string |
| **Progress export** | Data only (theme unlocks, lifetime stats, achievements, prefs) | Save the export file somewhere safe |

They are deliberately separate. The code is small because it carries identity, not data; the export grows with your progress because it carries your earned data, not identity. Neither one replaces the other — a full transfer to a new machine uses **both**.

### Exporting

Run:

```
account export
```

This writes a signed `account-export.json` next to your account file (under `./data/`). To choose your own location, pass a path:

```
account export /path/to/my-ageforge-backup.json
```

Export after any big unlock. It's a one-shot snapshot — it does **not** auto-update as you keep playing, so re-export when you've earned something you'd hate to lose.

### Importing

On the same or a new machine, restore from a backup file:

```
account import /path/to/my-ageforge-backup.json
```

By default this **merges** the backup into your current account, which is the safe choice:

- **Theme unlocks** are unioned — importing an old backup never *removes* a theme you've unlocked since.
- **Achievements** are unioned.
- **Lifetime stats** take the higher of the two values, so your bests never regress.
- **Active theme** keeps your current choice if you have one, otherwise adopts the backup's.

To overwrite your current progress wholesale with the backup instead, add `replace`:

```
account import /path/to/my-ageforge-backup.json replace
```

If the file is missing or has been tampered with, the import is rejected with a clear error and your account is left unchanged — corrupt data never gets imported.

> **No server, so no magic.** Progress recovery only works if you exported it first. The recovery code resurrects who you are; the export resurrects what you've done. If a machine's `./data/` is gone and you never exported, that progress is gone too — there's no cloud copy to pull back.

---

## Restoring on a new machine

There are **two ways** to restore your identity, because your identity has two equivalent forms — your name and your recovery code both resolve to the same account ID.

### The simplest way: re-enter your name

On a fresh machine, the first-run prompt asks you to name your account. **Type the exact same name you used before** and you'll regenerate the exact same account ID — no code required. (Remember normalization: casing and surrounding spaces don't matter, but a *meaningfully* different name is a different identity.) This restores **identity only**, not your earned progress — for that you still need a progress export.

### The precise way: recover from a code

Alternatively, on the new machine (or after a reinstall), run:

```
account recover AGEF-7Q2K-9X4M-ZJ31-…
```

This replaces the local account's identity with the one encoded in the code.

**Recovery is lenient about transcription.** You don't have to get the formatting perfect:

- Lowercase is fine.
- Extra spaces are ignored.
- The ambiguous swaps `I`/`L` ↔ `1` and `O` ↔ `0` are corrected automatically.

So if you wrote it down by hand and your `0` looks like an `O`, it'll still import.

### Overwrite guard

Recovering **replaces** the local identity. If the local account already has **unlocked progress**, `account recover` won't blow it away silently — it warns you and asks you to confirm:

```
account recover AGEF-… confirm
```

The guard exists because the code does **not** carry your unlocks. Recovering points this machine at a different identity; the local progress doesn't travel with it. The `confirm` keyword is your acknowledgement that you understand that before proceeding.

---

## The honest truth

With no server, **identity recovery is a short code you write down** — that's the whole trick. But **progress recovery requires you to have exported it first.** We can't make data appear from nothing: if a machine's `./data/` is gone and the progress was never backed up, there's no server holding a copy to pull it back from. The recovery code resurrects who you are, not what you've done.

That's also why the code is safe to share or lose track of — it's an identifier for a cosmetic, local account, not a key to anything worth guarding.

---

## See also

- [All Commands](commands.md) — the full `account` / `account recover` command reference
- [Saving & Loading](saving-and-loading.md) — how your *game saves* (a separate thing from your account) are stored under `./data/saves/`
