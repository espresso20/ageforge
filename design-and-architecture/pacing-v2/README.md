# Pacing v2: design docs

The spec for Pacing v2: a one-week first run, faster replays of ages already completed, and away play that costs nothing the player didn't choose. Kept as written: this folder is an archive of the design, not a page that tracks the code. When the code and a doc here disagree, the code wins and `design-and-architecture/economy.md` and the wiki (`site/docs/`) describe it.

| File | What it is | Status |
|------|-----------|--------|
| [prestige-audit.md](prestige-audit.md) | Audit of the prestige loop and storage caps at `54c171d` (2026-09-29): what points earn and buy, why the shop barely moves pacing, caps under multipliers, the cost of a slower game, and three redesign options. | Source for the plan. The owner chose Option C, Era Mastery. Its measurements predate the one-week curve (#166) and the retired `speed` setting (#156). |
| [plan.md](plan.md) | Implementation plan written 2026-09-30: seven PRs, the targets the smoke suite enforces, save migration and refund, every clock that needs re-timing, risks and open questions. | In progress. A status note at the top lists what has shipped (PRs 1, 2 trimmed and 4), what is in flight (PR 3) and what remains (5, 6, 7), with the owner's decisions. |
