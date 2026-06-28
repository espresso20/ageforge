# Accounts & Player Identity

**Status:** Design proposal (implementation-ready)
**Author:** engineering
**Related:** `theming.md` (theme unlocks are the first consumer of this system), `game/save.go`, `game/devmode.go`

---

## 0. TL;DR — the recommendation up front

Build a **local account file as the single source of truth, plus a short human-readable recovery code that encodes ONLY the identity.** Concretely:

- `data/account.json` — small, signed, identity + meta-progression state (theme unlocks, lifetime stats, achievements). One per machine, auto-created on first run.
- A **recovery code** (e.g. `AGEF-7Q2K-9X4M-ZJ31`) that encodes the account ID + a checksum + optional display name. This is the *only* thing a player needs to write down to keep their **identity** stable across machines/reinstalls. It is tiny because it carries identity, not data.
- **Progress data** (unlocks/stats) is exported/imported as a separate signed blob (`account-export.json` or a long base64 string), because — with no server — data cannot be reconstituted from nothing.
- Reuse the **existing HMAC + atomic-write patterns** from `game/save.go`. Do not invent a new integrity scheme.

The hard truth, stated once and honestly: **there is no server, so genuine cross-machine recovery requires the player to keep a backup.** We can make that backup small (a code for identity) and easy (one-button export for data). We cannot make data appear from nothing.

---

## 1. Problem statement

### 1.1 What we have today

AgeForge is a **local** terminal idle game. Everything lives on the player's machine. There is no backend, no network identity, no login. On disk we have exactly two things (confirmed by inspecting the tree and the code):

- `data/saves/*.json` — one file per run. Resolved by `saveDirectory()` in `game/save.go`: binary-relative `data/saves/` via `os.Executable()` + `EvalSymlinks`, with a CWD-relative `data/saves` legacy fallback in `savePath()`.
- `data/logs/` — log dumps (CWD-relative `data/logs`, created ad hoc in `ui/input.go`).

There is **no** `settings.json`, `account.json`, `preferences`, or `profile` file, and no loader for one. Verified by grep across `*.go`. This is greenfield.

Each save (`GameSave` in `game/save.go`) is self-contained: it carries its own resources, buildings, workers, prestige, badges, lineage parent (`ParentName`), etc. State is **per-save**, not **per-player**. There is no concept of "the player" that spans saves.

### 1.2 Why we now need a per-player identity

`theming.md` introduces **theme unlocks** as meta-progression: a player unlocks a theme by reaching some milestone, and that theme should then be available in **every** save — new games, branches, all of them. The same shape applies to the meta-progression we already know is coming:

- **Cosmetics** (themes, palettes, splash variants)
- **Lifetime / cross-save stats** ("total prestiges across all civilizations", "highest age ever reached")
- **Achievements** (one-time, account-wide, not per-run)

None of these belong in a `GameSave`. If theme unlocks lived in the save, then:

- Starting a new game would reset your unlocks (wrong — they're earned once).
- A player with three civilizations would have three divergent unlock sets (wrong — there's one player).
- Deleting a save would delete cosmetics you earned (wrong).

So we need a persistent **player identity** that all saves attach to, with a per-player state bag that the rest of the game reads from.

### 1.3 Constraints (non-negotiable)

1. **Local-only.** No server, no network calls, no cloud dependency for core function.
2. **Low friction.** First run must "just work" — no forced login, no mandatory naming wall before the player can play.
3. **Consistent with existing patterns.** Reuse HMAC signing + atomic temp-file+rename. Do not break the cheater/elite badge system or the `forgeMasterKey` easter egg.
4. **Backward compatible.** Existing players have no account; the game must create one transparently and adopt their existing saves.

---

## 2. Evaluation — A vs B vs Hybrid

The crux of this whole design is one distinction, so name it explicitly:

> **IDENTITY** is *who you are* — a small, stable token (an account ID, maybe a display name). It is cheap to back up and cheap to derive.
>
> **DATA** is *what you've earned* — unlock state, lifetime stats, achievements. It grows over time, is not derivable from anything, and must physically live somewhere.

Every approach below succeeds or fails on how it treats these two things. The recurring trap is conflating them — trying to make identity carry data, or assuming data is recoverable just because identity is.

### 2.1 Option A — local account-ID file

**Shape:** On first run, generate a stable ID (UUIDv4). Write it to `data/account.json`. Key all unlocks/prefs to it. Done.

```jsonc
// data/account.json (Option A, minimal)
{
  "account_id": "a1b2c3d4-...-uuid",
  "display_name": "Adam",
  "created": "2026-06-27T12:00:00Z",
  "unlocks": { "themes": ["amber_crt", "obsidian"] },
  "_sig": "hmac-sha256-hex"
}
```

**First-run creation:** Auto-generate silently. Optionally offer a display-name prompt, but never block on it — default to a generated name (we already have a cosmetic name generator in `game/savenames.go`) or empty.

**Backup/restore:** Copy the file. Restore = drop it back.

**Machine transfer:** Copy `data/account.json` to the new machine's data dir.

**Tamper:** Sign it with HMAC (same `saveHMACKey` scheme). On mismatch, we *don't* brick the player — we flag it (mirrors the cheater badge: cosmetic consequence, not a lockout), because the unlock state is not security-critical.

**Failure modes — and this is where A hurts:**
- **File deleted → identity AND unlocks both gone.** There is no backup of either. The player is a brand-new account; every theme is re-locked. This is the worst property of pure A: a single file is the only copy of *both* halves, and losing it loses everything at once.
- **Multi-machine:** No story at all. Copying the file by hand works but is undiscoverable and error-prone (wrong directory, overwrites the new machine's progress). There's no compact thing to "write down."
- **Multiple profiles on one machine:** Possible (multiple account files), but the file is a fixed path, so you'd need a profile-switcher. Out of scope for v1 but the schema shouldn't preclude it.

**Verdict:** A is the right *storage substrate* — a local signed file is exactly what we want for DATA. But as the *whole* solution it has no backup story beyond "copy this file," which is precisely the friction we're trying to avoid and the failure mode we're trying to soften.

### 2.2 Option B — seed / passphrase-derived identity

**Shape:** Player picks a name + passphrase. We derive a deterministic account ID: `account_id = KDF(passphrase, salt)`. Backup = "remember your passphrase." On a new machine, re-enter it and your ID regenerates.

**This addresses the user's instinct that "cramming all the data into a seed is bad" — and it's correct to be suspicious. Here's the precise reasoning:**

A seed can only ever derive things that are *pure functions of the seed*. An account ID is exactly that: `id = KDF(seed)`. Deterministic, tiny, reproducible anywhere. **So a seed can carry IDENTITY perfectly.**

But unlocks/stats/achievements are **not** a function of the seed — they're a function of *everything you did over months of play*. There is no `KDF` that turns "amber_crt, obsidian, 14 prestiges, highest age = bronze" back out of a passphrase. So:

- **Can the seed carry only identity?** Yes, cleanly. But then on a fresh machine you get your *ID* back and **all your DATA is gone** — same data-loss problem as A's deleted file, just relocated. Re-deriving the identity doesn't re-derive the unlocks, because the unlocks were never in the seed.
- **What would it take to make a seed restore actual progress?** You'd have to *encode the data into the exportable code itself* — turn the "seed" into an export blob: `base64(compress(unlock_state + stats + checksum))`. The moment you do that, it stops being a memorable passphrase and becomes a **long opaque string that grows with progress.** A handful of theme unlocks is maybe 40–80 chars; add lifetime stats and achievements and you're at hundreds of characters of base64. Nobody writes that on a sticky note. You've reinvented "export a file" but worse, because you've disguised a data blob as a "seed" and set the expectation that it's memorable.

**Other B problems:**
- **Collisions / security:** A weak passphrase = a guessable/duplicable account ID. To resist this you need a real KDF (argon2/scrypt) and ideally a per-install salt — but a per-install salt that lives only on the machine reintroduces the "lose the machine, lose the salt" problem, and a *fixed* salt makes the whole namespace brute-forceable. For a cosmetic-unlock system this is a lot of crypto ceremony for low stakes, and it pushes friction onto the player (choose and remember a strong passphrase up front).
- **Friction:** B forces an identity decision *before* the player has any reason to care, violating constraint #2.

**Verdict:** B is genuinely good at the *identity* half and genuinely bad at the *data* half, and pretending otherwise (by stuffing data into the seed) is the anti-pattern the user already smelled. The useful kernel of B — "a small derivable/portable identity token" — survives into the hybrid as the **recovery code**.

### 2.3 Option C — Hybrid (recommended), plus rejected alternatives

**The hybrid takes the best half of each:**

- From A: a **local signed file** as the source of truth for DATA (and the live identity). Zero friction, works offline, fits our existing patterns.
- From B: a **small, portable identity token** — but instead of deriving the ID *from* a passphrase, we generate the ID and *encode it into* a short recovery code. The code is small precisely because it carries identity only.

So:

| Concern | Lives where | Backup mechanism |
|---|---|---|
| **Identity** (account ID, display name) | `data/account.json` (live copy) | **Recovery code** — short, write-downable |
| **Data** (unlocks, stats, achievements) | `data/account.json` (live copy) | **Export blob/file** — explicit, larger, one-button |

This keeps the user's IDENTITY-vs-DATA separation honest: the *small* backup (recovery code) restores identity; the *bigger* backup (export) restores data; and we never lie about a passphrase magically restoring progress.

**Rejected alternatives (judged honestly):**

- **Cloud / GitHub Gist sync.** Tempting — it's the only thing that gives true automatic cross-machine recovery. Rejected for v1 because it breaks constraint #1 (local-only), adds auth/secrets/network-failure handling, and turns a cosmetic-unlock feature into an infrastructure project. Worth keeping as a *future* opt-in layered on top of the export blob (the export format becomes the sync payload), but not now.
- **Plain file export/import with no recovery code.** This is the hybrid minus the small identity token. It works, but the only backup is "copy a JSON file," which is the same undiscoverable friction as A. The recovery code is cheap to add and is what makes "I reinstalled / new laptop" not feel hostile. Keep both.
- **Tie identity into save HMAC as the *only* mechanism.** We *will* attribute saves to an account (§5), but identity can't live solely inside saves — that's the per-save trap from §1.2.

> **v1 caveat — one fixed `account.json` path means no multi-profile.** Because v1 resolves a single `data/account.json`, two players sharing one install (or two OS users on one machine pointed at the same data dir) **collide on the same account file and merge progress — last writer wins.** There is no per-player separation in v1. True multi-profile (`data/accounts/<id>.json` + a switcher) is deferred future work (see §7 and §9); the `version` field leaves room for it without a breaking change.

---

## 3. Recommended design

**Decision: Hybrid (Option C).** Local signed account file as source of truth + short recovery code for identity backup + explicit export/import for data backup. Cloud sync explicitly deferred.

### 3.1 Where it lives

A new **account directory**, resolved exactly like saves are. We factor the binary-relative resolution out of `saveDirectory()` into a shared `dataDirectory()` so accounts, saves, and (eventually) logs share one root and one fallback rule:

```
<dir-of-binary>/data/
  ├── account.json        ← NEW: identity + meta-progression (signed, atomic)
  ├── saves/*.json
  └── logs/
```

Resolution rule (mirrors `savePath`): prefer binary-relative `data/account.json`; if absent, check CWD-relative `data/account.json` (legacy/dev-run fallback); default to binary-relative for new writes.

### 3.2 Unified file vs split file

We considered `account.json` (identity) + `settings.json` (prefs/unlocks) as two files. **Recommendation: one file, `data/account.json`,** with internal sections. Rationale:

- One file = one signature, one atomic write, one backup unit. Two files doubles the integrity surface and invites a half-restored state (identity from machine A, unlocks from machine B).
- Prefs and unlocks are both "per-player," so they share a lifecycle.
- If non-account-scoped *machine* prefs ever appear (e.g. terminal-specific render tweaks that should NOT travel with the account), *those* get their own unsigned `data/prefs.json`. Until such a pref exists, don't create the file.

### 3.3 Schema

```jsonc
// data/account.json
{
  "version": 1,                         // schema version, for migrations
  "account_id": "f3a1c9e2b7d44a18...", // 128-bit random, hex (16 bytes)
  "display_name": "Adam",               // optional; "" is valid
  "created": "2026-06-27T12:00:00Z",
  "last_seen": "2026-06-27T18:30:00Z",

  // --- meta-progression (DATA) ---
  "unlocks": {
    "themes": ["amber_crt", "obsidian"] // theme keys; see theming.md
    // future: "palettes": [...], "splash_variants": [...]
  },
  "stats": {                            // lifetime, cross-save aggregates
    "total_prestiges": 14,
    "highest_age": "bronze_age",
    "civilizations_started": 7,
    "saves_completed": 2
  },
  "achievements": ["first_prestige", "reached_iron"],

  // --- preferences (travel with the account) ---
  "prefs": {
    "active_theme": "obsidian"
  },

  // --- integrity (same scheme as saves) ---
  "_sig": "hmac-sha256-hex"
}
```

Notes:
- `account_id` is **16 random bytes (crypto/rand) as hex**, not a v4 UUID specifically — but UUIDv4 is fine if a dep is already present. 128 bits makes collisions a non-issue and is what the recovery code encodes.
- All new fields are optional / zero-valued on read, so schema additions stay backward compatible (same discipline as `GameSave`).
- `version` lets us migrate the file later without guessing.

### 3.4 Integrity — reuse, don't reinvent

Sign with the **same HMAC-SHA256 construction** as saves:

- Zero `_sig`, `json.Marshal`, HMAC with `saveHMACKey`, hex-encode, store in `_sig`. (Identical to `signSave` — factor it into a shared helper, e.g. `hmacSign(payload []byte, key string)`, so saves and accounts call one function.)
- On load: if `_sig` is empty → treat as legacy/benign (don't punish). If present and mismatched → set a cosmetic `tampered` flag on the in-memory account (mirrors `cheaterBadge`), but **still load it** — unlock state is not security-critical, and a hard failure would be a worse experience than a tampered-themes notice.
- We deliberately do **not** add a `forgeMasterKey` proof to the account file. The elite easter egg is a *save* concept; it stays in saves untouched. (See §5.)

Write with the **same atomic pattern**: temp file + `os.Rename`, `0644`, `MkdirAll(dir, 0755)` — copy `SaveGame`'s body.

### 3.5 The recovery code (identity backup)

Encode `account_id` + a checksum (+ optional truncated display-name hint) into a grouped, human-readable string:

```
AGEF-7Q2K-9X4M-ZJ31-...   (Crockford base32, dash-grouped, uppercase)
```

- **Payload:** 16-byte `account_id` + 2-byte CRC-ish checksum. Base32 of 18 bytes ≈ 29 chars + dashes. Short enough to write down, long enough to not collide.
- **Crockford base32** avoids ambiguous chars (no `I/L/O/U`), so transcription is robust.
- The checksum lets import reject a typo'd code with a clear message instead of silently creating a wrong account.
- **What it does:** re-creates an `account.json` with the *same account_id* on a new machine (or after deletion). Identity restored. Saves signed/attributed to that ID re-associate (§5).
- **What it does NOT do:** restore unlocks/stats. Those are DATA — see §3.6. We must say this plainly in the UI when showing the code: *"This restores your identity, not your progress. Export your progress separately to keep it."*

> **INVARIANT — re-association goes through `SaveGame`, never an in-place JSON edit.**
> "Saves re-associate" sounds innocent, but the naive implementation is dangerous. When a restored identity re-claims its saves, **the new `account_id` must be written by re-marshalling and re-saving the save through the normal `SaveGame` path** so the HMAC `_sig` recomputes over the new payload. **Never** open a save file, splice `account_id` into the JSON, and write it back: `verifySave` re-hashes the *entire* payload, so a save with a freshly-edited `account_id` but a stale `_sig` fails verification → `sigValid=false` → `CheaterBadge=true`. That stamps a false "⚠ modified" badge on a legitimately-imported save. Same invariant applies to migration stamping (§6). One door in: `SaveGame`.

- **The recovery code is a convenience identifier, not a credential.** `account.json` is signed with `saveHMACKey`, a *constant compiled into every binary*. Anyone can therefore craft a validly-signed `account.json` for *any* account ID they like — the signature proves "this file came from an AgeForge build," not "this file belongs to you." The checksum in the code guards against *typos*, not *forgery*. This is acceptable precisely because account state is **cosmetic, not security-critical**: there is nothing to steal and no advantage to forging someone's ID. We do not pretend otherwise, and we never build a feature that assumes the code is a secret.

### 3.6 Progress export / import (data backup)

A one-button **Export Progress** writes a signed blob — either a file (`account-export-<date>.json`, byte-for-byte an `account.json` minus the live timestamps) or a copyable long base64 string for the "paste it somewhere" crowd.

- Import validates the signature, then either *merges* (union of unlocks, max of stats) or *replaces* (prompt). Merge is the safer default — re-importing an old backup shouldn't *remove* a theme you've since unlocked.
- This is honest about size: the export grows with your progress. That's fine — it's a *file/blob*, not something you memorize. This is exactly the thing Option B got wrong by disguising it as a "seed."

---

## 4. Backup / restore / machine-transfer UX

| Scenario | Player does | Result |
|---|---|---|
| **Same machine, normal play** | nothing | `account.json` is the live source of truth; auto-updated. |
| **Back up identity** | copies the **recovery code** (one screen, ~29 chars) | Can re-create the same account ID anywhere. |
| **Back up progress** | clicks **Export Progress** → saves a file / copies a blob | Can restore unlocks + stats anywhere. |
| **New machine (full transfer)** | enters recovery code → imports export blob | Identity + data both restored. |
| **New machine (identity only)** | enters recovery code | Same account ID; **unlocks start empty** — re-earnable, and re-importing later will merge. |
| **Reinstall, kept nothing** | — | Fresh account. Honest loss. |

**The honest line, in the doc and in the UI:** with no server, *identity* recovery is a 29-char code, but *progress* recovery requires you to have exported it. We make export one click and remind the player to do it after big unlocks; we cannot resurrect data that was never backed up.

---

## 5. Anti-tamper stance & relationship to existing systems

- **Cheater/elite badges (saves) are untouched.** The account file gets its *own* cosmetic `tampered` flag using the *same* HMAC helper, but with no `forgeMasterKey` proof and no lockout. We are consistent with the existing philosophy stated in `save.go`: the HMAC is integrity signalling, *"cosmetic, not security-critical."*
- **The `forgeMasterKey` easter egg stays a save-only concept.** We do not extend it to accounts, do not move the key, do not change `verifySave`. An elite save remains elite regardless of account.
- **Optional save→account attribution.** We can add an optional `AccountID string \`json:"account_id,omitempty"\`` to `GameSave`, stamped on save and covered by the existing save signature automatically (it's part of the marshalled payload `signSave` already hashes). This lets the Load Game browser show "this save belongs to account X" and lets a restored identity re-claim its saves. It is *additive and optional*: empty on legacy saves (`omitempty` keeps bytes identical), and a mismatch is informational, never blocking. No change to the signing/verifying logic is required — just a new field.
- **The stamp must go through `SaveGame`, full stop.** Whether the `account_id` is written by migration (§6) or by recovery re-association (§3.5), it MUST be applied by re-saving through the normal `SaveGame` path so `_sig` recomputes — **never** by editing the save's JSON in place. See the called-out invariant in §3.5: a stale signature flips `sigValid=false` → `CheaterBadge=true` → a false "⚠ modified" badge on a legitimate save.
- **Elite saves survive stamping — but only because stamping goes through `SaveGame`.** A save carrying an elite `_proof` (the `forgeMasterKey` easter egg) is re-signed when `account_id` is stamped onto it. Because the same `SaveGame` flow that recomputes `_sig` also recomputes the elite proof, elite status is preserved across attribution. This is *another* reason the in-place-edit shortcut is forbidden: editing the JSON directly would leave `_proof` stale and silently strip the elite badge.
- **The forged-account caveat (A4) applies here too.** Save attribution is informational, not authoritative: because `saveHMACKey` is a known constant (see §3.5), an `account_id` on a save proves nothing about ownership. Treat it as a hint for the Load browser, never as access control.

---

## 6. Migration — existing players

Existing players have saves but no account. On startup:

1. `LoadOrCreateAccount()` resolves `data/account.json`.
2. **Absent →** generate a new `account_id`, write a fresh signed `account.json`. Silent; no prompt. The player keeps playing; a non-blocking toast can mention "your account was created — back up your recovery code in Settings" but must not gate play (constraint #2).
3. **Present →** load + verify; set `tampered` flag on mismatch; update `last_seen`; rewrite atomically.
4. **Attaching existing saves:** when an un-attributed save (`account_id == ""`) is *next saved*, stamp it with the current account ID. We do **not** mass-rewrite all saves on first run (that would invalidate nothing — the sig recomputes — but it's needless I/O and churns mtimes the Load browser sorts on). Lazy attribution on next write is enough; the browser treats `""` as "this account" by default.

> **INVARIANT (restated for migration) — stamp via `SaveGame`, never edit JSON in place.** Lazy stamp-on-save is safe *only because the stamp rides the normal `SaveGame` write*, which re-marshals the payload and recomputes `_sig` (and the elite `_proof`, §5). The lazy timing is what makes it cheap; the `SaveGame` path is what makes it correct. Do not "optimize" this by reading a save, splicing in `account_id`, and writing the bytes back — `verifySave` re-hashes the whole payload, so a stale `_sig` yields `sigValid=false` → `CheaterBadge=true` → a false "⚠ modified" badge. This is the single most important implementation rule in this doc. Full statement in §3.5.

No existing save format field is removed or repurposed. The only save-side change is the additive optional `account_id`.

---

## 7. Failure modes & how the design handles each

| Failure | Without this design | With the hybrid |
|---|---|---|
| **`account.json` deleted** | (n/a today) | Identity lost *unless* recovery code kept; unlocks lost *unless* export kept. Auto-creates a fresh account so the game still runs. We mitigate by surfacing the recovery code prominently and nudging export after unlocks. |
| **Multi-machine** | none | Recovery code re-creates identity; export blob carries data. Manual but small and discoverable. |
| **Multiple profiles, one machine** | none | v1: single `account.json`. Schema/`version` field leaves room for a future `data/accounts/<id>.json` layout + switcher without a breaking change. Not built in v1. |
| **Tampered account file** | (n/a) | Loads with a cosmetic `tampered` flag (mirrors cheater badge); no lockout. |
| **Corrupt / unparseable account file** | (n/a) | Treat like a corrupt save: don't crash. Back up the bad file to `account.json.corrupt`, create a fresh account, log it. Player can recover via code+export. |
| **Two machines edit then both `account.json` copied around** | none | Last-write-wins on the file; import *merges* unlocks (union) and *maxes* stats to limit accidental loss. |

---

## 8. Integration API — how the rest of the game uses it

A small package (e.g. `account`) owns the file and exposes a narrow surface. The UI and theming code talk only to this — they never touch the file directly. Mirrors how `game` owns saves.

```go
package account

// Loaded once at startup; held by the app, passed explicitly (no globals — per CLAUDE.md).
type Account struct { /* fields from §3.3 */ }

func LoadOrCreate() (*Account, error)        // §6 startup path
func (a *Account) Save() error               // signed + atomic, reuses save helpers

// Unlocks (theming.md is the first caller)
func (a *Account) HasTheme(key string) bool
func (a *Account) UnlockTheme(key string) (newly bool, err error) // persists; returns true if it wasn't already unlocked
func (a *Account) UnlockedThemes() []string

// Prefs
func (a *Account) ActiveTheme() string
func (a *Account) SetActiveTheme(key string) error

// Lifetime stats (engine calls these at prestige/age-up; debounce writes)
func (a *Account) RecordPrestige()
func (a *Account) RecordAgeReached(ageKey string)

// Identity backup / restore
func (a *Account) RecoveryCode() string                 // §3.5
func ImportRecoveryCode(code string) (*Account, error)  // re-creates identity
func (a *Account) ExportProgress() ([]byte, error)      // §3.6 signed blob
func ImportProgress(blob []byte, merge bool) error
```

**Theming hook:** the theme picker calls `HasTheme`/`UnlockedThemes` to decide what's selectable, and the engine calls `UnlockTheme` when an unlock condition is met (the condition logic lives in `theming.md`). Because the account is loaded once and passed explicitly, no lock-acquiring call leaks into a Bus handler (respects the documented Bus deadlock rule).

**Write cadence:** `Save()` after meaningful state changes (an unlock, a prefs change, a prestige), not every tick. Like autosave, the file write happens outside any engine lock.

---

## 9. Phased implementation plan

Each phase is independently shippable and testable. Sized for one self-contained subagent each per project conventions.

- **Phase 1 — Storage substrate.**
  Factor `dataDirectory()` out of `saveDirectory()`. New `account` package: `Account` struct, `LoadOrCreate`, `Save` (reusing a shared `hmacSign` helper extracted from `signSave`, plus the temp+rename write). Tamper/corrupt handling. Tests: create → load → verify sig → tamper → flag set; corrupt → fresh account + `.corrupt` backup.

- **Phase 2 — Startup wiring + migration.**
  Call `LoadOrCreate` at app boot; thread the `*Account` through to the UI. Add optional `account_id` to `GameSave`; lazy-stamp on next save. Non-blocking first-run toast. Tests: legacy save loads unchanged; next save gains `account_id`; sig still valid.

- **Phase 3 — Unlock API + theming integration.**
  `HasTheme`/`UnlockTheme`/`UnlockedThemes`/`ActiveTheme`/`SetActiveTheme` (signatures in §8). Wire the theme picker (per `theming.md`) to the account. Tests: unlock persists across restart and across a *new* save.
  **Cross-doc dependency:** this phase *unblocks* `theming.md` Phase 2 (account-wide persistence), which is gated on exactly this unlock API. `theming.md` Phase 1 deliberately ships against a *stub* of this API and has **no** dependency on the account system, so the two tracks can proceed in parallel and only converge here. Named in both plans (`theming.md` §10) so neither can be scheduled to deadlock the other.

- **Phase 4 — Recovery code.**
  Crockford base32 encode/decode of `account_id` + checksum; `RecoveryCode()` / `ImportRecoveryCode`. Settings screen shows the code + the honest "identity, not progress" copy. Tests: round-trip; typo → checksum rejection; import re-creates same ID.

- **Phase 5 — Export / import progress.**
  `ExportProgress` (signed blob/file) + `ImportProgress(merge)`. Settings buttons. Tests: export → wipe unlocks → import(merge) restores; import doesn't drop newer unlocks.

- **Phase 6 — Lifetime stats + achievements.**
  Engine hooks (`RecordPrestige`, `RecordAgeReached`, …); a small achievements table. Surface lifetime stats in the Stats tab.

- **Deferred / future (not in this plan):** multi-profile switcher (`data/accounts/<id>.json`); opt-in cloud/gist sync built *on top of* the Phase-5 export format. Both are non-breaking thanks to the `version` field and the narrow API.

- **Docs:** every phase updates the wiki per the project rule — at minimum `site/docs/commands.md` (new Settings/account actions) and a new account/recovery page; `theming.md` cross-links the unlock API.

---

## Appendix — grounding notes (code as-built)

- No account/settings/preferences file or loader exists today (grep-confirmed). Greenfield.
- `game/save.go`: `saveDirectory()` = binary-relative `data/saves` via `os.Executable()`+`EvalSymlinks`; `savePath()` falls back to CWD-relative `data/saves`. `signSave` zeroes `_sig`/`_proof`, marshals, HMAC-SHA256 under `saveHMACKey`, hex. `verifySave` treats unsigned as benign-valid. Atomic writes via temp+`Rename` (`SaveGame`, `reparentSaveFile`, `DuplicateSave`). `GameSave` carries `ParentName` (lineage) and `CheaterBadge`/`EliteBadge`.
- `game/devmode.go`: `forgeMasterKey` signs the *save signature* for the elite badge (the easter egg). Deliberate; left untouched by this design.
- Names go through `ValidateSaveName`/`sanitizeSaveName`. Module: `github.com/espresso20/ageforge`, Go 1.24.
