package smoke

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/espresso20/ageforge/game"
)

// The saveload scenario plays a seed and, at one checkpoint per age reached
// (checkpointDelay ticks into the age), saves through the real save path.
// The uninterrupted run carries on; a fresh engine loads the save and plays
// the same N ticks with a fresh bot. Both must end bit-identical. A load
// must also reproduce the state it was saved from, re-save to the same
// bytes, and pass the signature check.

// checkpointDelay is how far into an age the checkpoint is taken, so the
// state has an age's worth of queue, research and events in flight.
const checkpointDelay = 300

type checkpoint struct {
	idx      int
	age      string
	cycle    int
	ticks    int // runner ticks at the save
	sim      time.Duration
	gameTick int
	file     string
	endTicks int
	atSave   game.GameState
	atEnd    *game.GameState
	// b is the engine the save was loaded into, right after the save (a
	// save 5s or older gets offline catch-up on load); reload is its verdict.
	b      *game.GameEngine
	reload string
}

// saveloadSkip lists GameState paths that legitimately differ between an
// uninterrupted run and a reloaded one: whether the engine has a save slot,
// wall-clock play time, the log (not saved; compared after the checkpoint
// separately) and the last age advance's summary (a one-shot for the age
// splash, not saved).
func saveloadSkip(path string) bool {
	switch path {
	case "SaveExists", "Stats.PlayTime", "Log", "LastAgeAdvanceSummary":
		return true
	}
	return false
}

// reloadSkip is saveloadSkip plus the resource rates: a live snapshot holds
// the rates the last tick computed before morale and the like moved, while
// LoadGame recomputes them from the saved state. The next tick agrees
// again, and the continued-play comparison covers them. Military.SoldierRate
// is the soldiers resource's rate under another name.
func reloadSkip(path string) bool {
	if saveloadSkip(path) || path == "Military.SoldierRate" {
		return true
	}
	if strings.HasPrefix(path, "Resources[") {
		return strings.HasSuffix(path, "].Rate") || strings.Contains(path, "].Breakdown")
	}
	return false
}

// postLog is the log entries written after tick t (the log is not saved, so
// only what happened after the checkpoint can match).
func postLog(st game.GameState, t int) []game.LogEntry {
	var out []game.LogEntry
	for _, l := range st.Log {
		if l.Tick > t {
			out = append(out, l)
		}
	}
	return out
}

// saveJSONSkip lists save-file keys that change on every write.
var saveJSONSkip = map[string]bool{"timestamp": true, "_sig": true, "_proof": true}

func runSaveload(e *Env, res *Result) {
	seeds, cps, n := e.seeds(1), 2, 2000
	if e.full() {
		seeds, cps, n = e.seeds(2), 4, 10000
	}
	n -= n % e.Base.DecideEvery
	type out struct {
		rows  []string
		fails []Finding
		warns []Finding
	}
	results := runSeeds(e, seeds, func(seed int64) out {
		var o out
		rows, fails, warns := saveloadSeed(e, seed, cps, n)
		o.rows, o.fails, o.warns = rows, fails, warns
		return o
	})
	var rows []string
	for _, o := range results {
		rows = append(rows, o.rows...)
		res.Failures = append(res.Failures, o.fails...)
		res.Warnings = append(res.Warnings, o.warns...)
	}
	res.Summary = fmt.Sprintf("%d checkpoint(s) over %d seed(s), %d ticks of continued play each", len(rows), len(seeds), n)
	res.section("Checkpoints", "| seed | checkpoint | age | game tick | reload | continued %d ticks |\n|---|---|---|---|---|---|\n%s",
		n, strings.Join(rows, "\n"))
}

func saveloadSeed(e *Env, seed int64, want, n int) (rows []string, fails, warns []Finding) {
	fail := func(check, repro, format string, args ...interface{}) {
		fails = append(fails, Finding{Check: check, Seed: seed, Message: fmt.Sprintf(format, args...), Repro: repro})
	}
	warn := func(check, repro, format string, args ...interface{}) {
		warns = append(warns, Finding{Check: check, Seed: seed, Message: fmt.Sprintf(format, args...), Repro: repro})
	}
	repro := fmt.Sprintf("go run ./cmd/smoke -scenario saveload -seed-base %d -seeds 1 -v", seed)
	dir := filepath.Join(game.DataDir(), "saves")

	// Run A: the uninterrupted game, saving at each checkpoint.
	var cps []*checkpoint
	cfg := e.Base
	cfg.Cycles, cfg.MaxSim = 1, 2000*time.Hour
	lastAge := ""
	cfg.hook = func(r *runner) bool {
		for _, cp := range cps {
			if cp.atEnd == nil && r.ticks == cp.endTicks {
				st := deepCopy(r.ge.GetState())
				cp.atEnd = &st
			}
		}
		if len(cps) < want && r.age != lastAge && r.ticks-r.ageT0 >= checkpointDelay {
			lastAge = r.age
			st := deepCopy(r.ge.GetState())
			cp := &checkpoint{idx: len(cps) + 1, age: r.age, cycle: r.cycle, ticks: r.ticks, sim: r.sim,
				gameTick: st.Tick, file: fmt.Sprintf("saveload-%d-%d", seed, len(cps)+1), endTicks: r.ticks + n, atSave: st}
			if err := r.ge.SaveGame(cp.file); err != nil {
				fail("save_error", repro, "saving checkpoint %d (%s): %v", cp.idx, cp.age, err)
				return true
			}
			cps = append(cps, cp)
			// Load at once: LoadGame gives a save 5s or older offline
			// catch-up, which would read as a divergence.
			reloadCheckpoint(cp, dir, repro, fail, warn)
		}
		if len(cps) == want {
			done := true
			for _, cp := range cps {
				done = done && cp.atEnd != nil
			}
			return done
		}
		return false
	}
	a := Run(cfg, seed)
	for _, an := range a.Anomalies {
		fails = append(fails, Finding{Check: an.Kind + "/" + an.Check, Seed: seed, Message: "uninterrupted run: " + an.Message, Detail: an.Dump, Repro: repro})
	}
	if len(cps) < want {
		warns = append(warns, Finding{Check: "checkpoints", Seed: seed,
			Message: fmt.Sprintf("only %d of %d checkpoints: the run ended (%s at %s) before reaching more ages", len(cps), want, a.Outcome, a.FinalAge)})
	}

	for _, cp := range cps {
		cont := "-"
		if cp.b != nil {
			cont = continueCheckpoint(e, seed, cp, repro, fail, warn)
		}
		rows = append(rows, fmt.Sprintf("| %d | %d | %s | %d | %s | %s |", seed, cp.idx, cp.age, cp.gameTick, cp.reload, cont))
	}
	return rows, fails, warns
}

type reporter func(check, repro, format string, args ...interface{})

// reloadCheckpoint checks the save just written: its signature, that a
// tampered copy fails it, that it loads into a fresh engine as the state it
// was saved from, and that re-saving writes the same data. It leaves the
// loaded engine in cp.b and the verdict in cp.reload.
func reloadCheckpoint(cp *checkpoint, dir, repro string, fail, warn reporter) {
	reload := "ok"
	defer func() { cp.reload = reload }()
	path := filepath.Join(dir, cp.file+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		fail("save_missing", repro, "checkpoint %d: %v", cp.idx, err)
		reload = "missing"
		return
	}
	if cheater, _ := game.PeekSaveBadges(cp.file); cheater {
		fail("save_signature", repro, "checkpoint %d (%s): a save straight from SaveGame fails its own signature check", cp.idx, cp.age)
		reload = "bad signature"
	}
	// The check must also catch tampering: one changed number flips the badge.
	tampered := strings.Replace(string(raw), `"tick": `, `"tick": 1`, 1)
	if err := os.WriteFile(filepath.Join(dir, cp.file+"-tampered.json"), []byte(tampered), 0o644); err == nil {
		if cheater, _ := game.PeekSaveBadges(cp.file + "-tampered"); !cheater {
			fail("save_signature_blind", repro, "checkpoint %d: a save with its tick edited still passes the signature check", cp.idx)
		}
	}

	var b *game.GameEngine
	var stB game.GameState
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				fail("load_panic", repro, "checkpoint %d (%s): LoadGame panicked: %v\n%s", cp.idx, cp.age, rec, debug.Stack())
				b = nil
			}
		}()
		b = game.NewGameEngine()
		if err := b.LoadGame(cp.file); err != nil {
			fail("load_error", repro, "checkpoint %d (%s): LoadGame: %v", cp.idx, cp.age, err)
			b = nil
			return
		}
		stB = b.GetState()
	}()
	if b == nil {
		reload = "load failed"
		return
	}
	cp.b = b
	if stB.CheaterBadge {
		fail("save_signature", repro, "checkpoint %d (%s): the reloaded game carries the cheater badge", cp.idx, cp.age)
		reload = "cheater badge"
	}
	if d := firstDiff(cp.atSave, stB, reloadSkip); d != "" {
		fail("reload_divergence", repro, "checkpoint %d (%s, game tick %d): the loaded state differs from the saved one at %s", cp.idx, cp.age, cp.gameTick, d)
		reload = "differs"
	}
	resave := cp.file + "-resave"
	if err := b.SaveGame(resave); err != nil {
		fail("save_error", repro, "re-saving checkpoint %d: %v", cp.idx, err)
	} else {
		d, orderOnly := diffSaveFiles(path, filepath.Join(dir, resave+".json"))
		if d != "" {
			fail("resave_divergence", repro, "checkpoint %d (%s): saving the loaded game writes different data at %s", cp.idx, cp.age, d)
			if reload == "ok" {
				reload = "re-save differs"
			}
		}
		for _, p := range orderOnly {
			warn("save_map_order", repro, "checkpoint %d (%s): the save writes %s in a different order each time (a set serialized in map order), so two saves of the same game differ byte for byte", cp.idx, cp.age, p)
		}
	}
}

// continueCheckpoint plays the loaded engine on for the checkpoint's N
// ticks and compares it with the uninterrupted run. It returns the verdict
// for the table.
func continueCheckpoint(e *Env, seed int64, cp *checkpoint, repro string, fail, warn reporter) string {
	cont := "ok"
	if cp.atEnd == nil {
		return "not reached"
	}
	endB, perr := continueFrom(e, seed, cp, cp.b, nil)
	if perr != "" {
		fail("continue_panic", repro, "checkpoint %d (%s): the loaded game broke while continuing: %s", cp.idx, cp.age, perr)
		return "panicked"
	}
	d := firstDiff(*cp.atEnd, endB, saveloadSkip)
	if d == "" {
		d = firstDiff(postLog(*cp.atEnd, cp.gameTick), postLog(endB, cp.gameTick), nil)
		if d != "" {
			d = "Log (entries after the checkpoint)" + strings.TrimPrefix(d, "(root)")
		}
	}
	if d == "" {
		return cont
	}
	cont = "differs"
	f := fmt.Sprintf("checkpoint %d (%s, game tick %d): %d ticks after loading, the game differs from the uninterrupted run; first divergence at %s",
		cp.idx, cp.age, cp.gameTick, cp.endTicks-cp.ticks, d)
	if why := explainDivergence(e, seed, cp, endB); why != "" {
		f += ". " + why
	}
	fail("continue_divergence", repro, "%s", f)
	return cont
}

// continueFrom plays the checkpoint's N ticks on ge (a loaded engine, or nil
// to replay the seed from scratch up to the checkpoint first) with a fresh
// bot, exactly as the uninterrupted run's loop does, and returns the end
// state. before, if set, runs on the engine at the checkpoint.
func continueFrom(e *Env, seed int64, cp *checkpoint, ge *game.GameEngine, before func(*game.GameEngine)) (end game.GameState, panicked string) {
	cfg := e.Base
	cfg.Cycles, cfg.MaxSim = 1, 2000*time.Hour
	var r *runner
	if ge == nil {
		ge = game.NewGameEngine()
		ge.SeedRNG(seed)
		r = newRunner(cfg, seed, ge)
		r.enterAge(ge.GetState())
	} else {
		r = newRunner(cfg, seed, ge)
		r.ticks, r.sim, r.cycle = cp.ticks, cp.sim, cp.cycle
		r.enterAge(ge.GetState())
	}
	armed := false
	r.cfg.hook = func(r *runner) bool {
		if r.ticks == cp.ticks && !armed {
			armed = true
			if before != nil {
				before(r.ge)
			}
		}
		if r.ticks == cp.endTicks {
			end = deepCopy(r.ge.GetState())
			return true
		}
		return false
	}
	defer func() {
		if rec := recover(); rec != nil {
			panicked = fmt.Sprintf("%v\n%s", rec, debug.Stack())
		}
	}()
	for !r.step() {
	}
	return end, ""
}

// explainDivergence tells a save that loses the RNG position apart from a
// save that loses state: it replays the seed to the checkpoint, restarts the
// RNG there from the seed (what LoadGame did before saves carried the stream
// position), and plays on. If that matches the loaded run, everything saved
// came back except the stream position.
func explainDivergence(e *Env, seed int64, cp *checkpoint, loaded game.GameState) string {
	ctl, p := continueFrom(e, seed, cp, nil, func(ge *game.GameEngine) { ge.SeedRNG(ge.Seed()) })
	if p != "" {
		return ""
	}
	// The replay is a different engine, so its start time differs, and its
	// stream positions count from the restart.
	skip := func(path string) bool {
		return saveloadSkip(path) || path == "Stats.GameStarted" || path == "RNGDraws" || path == "QuipDraws"
	}
	if firstDiff(ctl, loaded, skip) == "" && firstDiff(postLog(ctl, cp.gameTick), postLog(loaded, cp.gameTick), nil) == "" {
		return "A replay that restarts the RNG from the seed at the checkpoint matches the loaded run exactly, so the save restores every field but not the RNG stream position: LoadGame must replay the saved rng_draws/quip_draws"
	}
	return "A replay that restarts the RNG at the checkpoint does not match the loaded run either, so the save loses state beyond the RNG position"
}

// diffSaveFiles compares two save files, ignoring the write timestamp and
// signatures. Lists of strings that hold the same items in a different
// order are not a difference, but come back in orderOnly: the save wrote a
// set in map order.
func diffSaveFiles(a, b string) (d string, orderOnly []string) {
	ra, err := os.ReadFile(a)
	if err != nil {
		return err.Error(), nil
	}
	rb, err := os.ReadFile(b)
	if err != nil {
		return err.Error(), nil
	}
	ta, err := jsonTree(ra)
	if err != nil {
		return "parse: " + err.Error(), nil
	}
	tb, err := jsonTree(rb)
	if err != nil {
		return "parse: " + err.Error(), nil
	}
	for _, t := range []interface{}{ta, tb} {
		if m, ok := t.(map[string]interface{}); ok {
			for k := range saveJSONSkip {
				delete(m, k)
			}
		}
	}
	orderOnly = sortStringSets(ta, tb, "")
	return firstDiff(ta, tb, nil), orderOnly
}

// sortStringSets sorts, in both trees, every list of strings found at the
// same path, and returns the paths where only the order differed.
func sortStringSets(a, b interface{}, path string) []string {
	var out []string
	switch x := a.(type) {
	case map[string]interface{}:
		y, ok := b.(map[string]interface{})
		if !ok {
			return nil
		}
		for _, k := range sortedKeys(x) {
			out = append(out, sortStringSets(x[k], y[k], join(path, k))...)
		}
	case []interface{}:
		y, ok := b.([]interface{})
		if !ok || len(x) != len(y) {
			return nil
		}
		sx, okx := stringList(x)
		sy, oky := stringList(y)
		if !okx || !oky {
			for i := range x {
				out = append(out, sortStringSets(x[i], y[i], fmt.Sprintf("%s[%d]", path, i))...)
			}
			return out
		}
		before := strings.Join(sx, "\x00") != strings.Join(sy, "\x00")
		sort.Strings(sx)
		sort.Strings(sy)
		if before && strings.Join(sx, "\x00") == strings.Join(sy, "\x00") {
			out = append(out, path)
		}
		for i := range x {
			x[i], y[i] = sx[i], sy[i]
		}
	}
	return out
}

func stringList(xs []interface{}) ([]string, bool) {
	out := make([]string, len(xs))
	for i, v := range xs {
		s, ok := v.(string)
		if !ok {
			return nil, false
		}
		out[i] = s
	}
	return out, true
}
