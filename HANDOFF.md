# Handoff — Age splash freeze fix (`claude/fix-age-splash-freeze-aUfxI`)

Status: fix committed and pushed; **PR not yet opened, not merged** (the session that
did the work had no GitHub PR tooling and was not permitted to push to `master`).

## Next steps for whoever picks this up

1. Open a PR from `claude/fix-age-splash-freeze-aUfxI` → `master` using the PR text below.
2. **Squash-merge.** The branch contains a wrong first attempt (`d37d158`) followed by its
   revert inside `32624a7`; squashing keeps that out of `master` history.
3. Delete this `HANDOFF.md` in the squash (or right after) — it's a handoff note, not docs.
4. Optionally file/fix the related UI bugs listed at the bottom.

Verify before merging:

```bash
go build ./... && go vet ./...
go test -short ./...                                   # fast suite
go test -race -run Repro -v -timeout 300s ./ui         # ~95s end-to-end regression tests
```

---

## PR text (ready to paste)

**Title:** fix: age splash freeze when an epoch catastrophe rolls on the same advance

### Summary
An age advance that crosses an epoch boundary can roll a catastrophe in the same call
(`advanceAge` → `rollEpochEvent` sets `PendingCatastrophe` while publishing
`EventAgeAdvanced`). The next `refresh()` then showed the age splash **and** stacked the
catastrophe modal on top of it:

- **Keys went to the modal.** "Press any key to continue" did nothing; the modal only reacts
  to e/s/d/Tab/Enter, and Enter silently chose ENDURE.
- **After 20s the modal became unreachable.** The splash's auto-dismiss
  (`OverlayManager.Hide` → `onClose`) moved focus to the command input *underneath* the
  still-visible modal. Nothing on screen could take keys; only Esc (back to menu) or
  Ctrl-C escaped — players killed the terminal.

It was **not** `SetBeforeDrawFunc` (PR #17), a focus-routing bug, or a deadlock — the tview
event loop stayed responsive in every run of the harness.

### Changes
- `ui/dashboard.go` `refresh()`: don't show the catastrophe modal while the age splash is
  active. It appears on the next refresh (≤500ms) after the splash is dismissed.
- `ui/age_splash.go`: revert of `d37d158`. That commit's theory was wrong (in tview v0.42 a
  Flex's input capture *does* fire for keys routed to its focused child), and making
  `titleTV` non-focusable meant a Pages re-focus landed on a bare Flex with no key handler.
- `ui/age_splash_repro_test.go`: end-to-end regression tests driving the real `App` on a
  `tcell.SimulationScreen` — all 21 typed `advance`s, forced catastrophe + advance with
  Defer/Endure, and idling past the 20s auto-dismiss. Skipped with `-short` (~95s).

### Test results
| Test | Before fix | After fix |
|---|---|---|
| `TestReproAgeSplashAllAges` | FAIL when an epoch catastrophe rolls | PASS |
| `TestReproAgeSplashWithCatastrophe` (Defer / Endure) | FAIL | PASS |
| `TestReproAgeSplashCatastropheAutoDismiss` | FAIL (modal unreachable) | PASS |

`go build`, `go vet`, `go test -short ./...` and the Repro tests with `-race` all pass.

---

## Related UI bugs found (not fixed in this branch)

- **Catastrophe buttons have no text.** `"[ENDURE]"` / `"[SUCCUMB]"` in
  `ui/catastrophe_modal.go` are parsed as tview colour tags. Wrap labels in `tview.Escape()`.
  The e/s/d shortcuts are also never shown.
- **Catastrophe modal blanks the screen.** It's padded with opaque `tview.NewBox()` spacers;
  the bottom rows let the dashboard bleed through.
- **`advance` needs two Enters** — autocomplete consumes the first.
- **`/speed` data race.** The dev command writes `speedMultiplier` without the lock that
  `getTickInterval` reads under (flagged by `-race`).
- **Pre-existing `gofmt` drift** in `ui/catastrophe_modal.go`, `ui/dashboard.go`,
  `ui/overlay_history.go`, `ui/overlay_workers.go`, `ui/villager_panel.go`,
  `ui/wonder_gallery.go`.
