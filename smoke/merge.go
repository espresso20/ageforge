package smoke

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// ReadSession loads a session report (report.json).
func ReadSession(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &s, nil
}

// configFromJSON is the inverse of the ConfigJSON a summary records, as far
// as grading and repro commands need it.
func configFromJSON(c ConfigJSON) Config {
	cfg := DefaultConfig()
	cfg.Seeds = c.Seeds
	cfg.Catastrophe, cfg.Harbinger, cfg.PrestigeAge = c.Catastrophe, c.Harbinger, c.PrestigeAge
	cfg.Cycles, cfg.FinalAge = c.Cycles, c.FinalAge
	cfg.DecideEvery, cfg.CheckEvery = c.DecideEvery, c.CheckEvery
	cfg.SoftlockSpan = time.Duration(c.SoftlockSecs * float64(time.Second))
	cfg.AgeTimeout = time.Duration(c.AgeTimeout * float64(time.Second))
	cfg.MaxSim = time.Duration(c.MaxSimSecs * float64(time.Second))
	cfg.Pacing, cfg.Style, cfg.LastPassage = c.Pacing, c.Style, c.LastPassage
	return cfg
}

// MergeProgression pools the progression runs of several sessions (the deep
// tier's one-seed CI jobs) into one session and grades them again under
// e.Pacing, so the pacing median, and its enforcement, is across every seed.
// The static gate check is run again (it reads config alone). Every part
// must hold a progression result, and no seed may appear twice.
func MergeProgression(e *Env, parts []*Session) (*Session, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("nothing to merge")
	}
	var runs []*RunResult
	var cfgJSON *ConfigJSON
	seen := map[int64]bool{}
	started := parts[0].Started
	var wall int64
	for i, p := range parts {
		var prog *Summary
		for _, r := range p.Scenarios {
			if r.Name == "progression" && len(r.Progression) > 0 {
				prog = r.Progression[0]
			}
		}
		if prog == nil {
			return nil, fmt.Errorf("report %d has no progression runs", i+1)
		}
		if cfgJSON == nil {
			c := prog.Config
			cfgJSON = &c
		} else if prog.Config.PrestigeAge != cfgJSON.PrestigeAge || prog.Config.Cycles != cfgJSON.Cycles || prog.Config.FinalAge != cfgJSON.FinalAge {
			return nil, fmt.Errorf("report %d played a different run (prestige %q, %d cycles, final %q) from the first (%q, %d, %q)",
				i+1, prog.Config.PrestigeAge, prog.Config.Cycles, prog.Config.FinalAge, cfgJSON.PrestigeAge, cfgJSON.Cycles, cfgJSON.FinalAge)
		}
		for _, r := range prog.Runs {
			if seen[r.Seed] {
				return nil, fmt.Errorf("seed %d appears in more than one report", r.Seed)
			}
			seen[r.Seed] = true
			runs = append(runs, r)
		}
		if p.Started.Before(started) {
			started = p.Started
		}
		wall = max(wall, p.WallMs)
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].Seed < runs[j].Seed })
	cfg := configFromJSON(*cfgJSON)
	cfg.Pacing = e.Pacing
	cfg.Seeds = nil
	for _, r := range runs {
		cfg.Seeds = append(cfg.Seeds, r.Seed)
	}

	sess := &Session{Tier: parts[0].Tier, Pacing: e.Pacing, Started: started, WallMs: wall}
	static := &Result{Name: "static"}
	runStatic(e, static)
	prog := &Result{Name: "progression"}
	for _, p := range parts {
		for _, r := range p.Scenarios {
			if r.Name == "progression" {
				prog.WallMs = max(prog.WallMs, r.WallMs)
			}
		}
	}
	describeProgression(prog, foldBotSet(e, prog, "progression", "progression", cfg, started, runs))
	prog.Summary = fmt.Sprintf("merged from %d report(s): %s", len(parts), prog.Summary)
	for _, r := range []*Result{static, prog} {
		r.Status = StatusPass
		if len(r.Failures) > 0 {
			r.Status = StatusFail
			sess.Failed = true
		}
		sess.Scenarios = append(sess.Scenarios, r)
	}
	return sess, nil
}
