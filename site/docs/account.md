# Account & Recovery

AgeForge is a **local terminal game**. There is no server, no login screen, no cloud — everything lives on your own machine under `./data/`, relative to the directory you launch the game from. Your account is part of that: a small file on disk that gives your civilization a stable identity and holds your earned meta-progression.

This page explains what an account is, what the recovery code does (and, just as important, what it doesn't), and how to carry your identity to a new machine.

---

## Your account

Your account is created **automatically on first run** and stored at `./data/account.json`. There's no signup, no prompt, and it never blocks you from playing — it's just there the first time you launch the game. One account per machine, or more precisely one per data directory.

The account holds two distinct things:

| Part | What it is |
|---|---|
| **Identity** | A stable account ID, and optionally a display name |
| **Data** | Your earned meta-progression — theme unlocks, lifetime stats, achievements |

The split matters, because the two halves are recovered very differently (see below). The **data** side is still being built out — lifetime stats and unlocks are tracked, but exporting and importing your progress between machines is **coming in a later update**.

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

To carry your progress between machines you'll back it up via **progress export**, which is **coming in a later update**. Until then, progress lives only in `./data/` on the machine that earned it.

---

## Restoring on a new machine

On the new machine (or after a reinstall), run:

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
