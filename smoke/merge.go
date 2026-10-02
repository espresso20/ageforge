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

// MergeSessions merges the sessions of several CI jobs into one: the deep
// tier's one-seed jobs and the nightly's shards. The progression runs of
// every part that holds them are pooled and graded again under e.Pacing, so
// the pacing median, and its enforcement, is across every seed; the static
// check is run again when a part held it or progression (it reads config
// alone); and every other scenario's result is carried over as its job
// reported it (its tables stay in that job's artifact: sections are not in
// report.json). No seed may appear twice, no other scenario may come from two
// parts, and the progression parts must have played the same run.
func MergeSessions(e *Env, parts []*Session) (*Session, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("nothing to merge")
	}
	var runs []*RunResult
	var cfgJSON *ConfigJSON
	seen := map[int64]bool{}
	carried := map[string]*Result{}
	carriedFrom := map[string]int{}
	static := false
	started := parts[0].Started
	var wall, progWall int64
	for i, p := range parts {
		for _, r := range p.Scenarios {
			switch {
			case r.Name == "static":
				static = true
			case r.Name == "progression" && len(r.Progression) > 0:
				static = true
				prog := r.Progression[0]
				progWall = max(progWall, r.WallMs)
				if cfgJSON == nil {
					c := prog.Config
					cfgJSON = &c
				} else if prog.Config.PrestigeAge != cfgJSON.PrestigeAge || prog.Config.Cycles != cfgJSON.Cycles || prog.Config.FinalAge != cfgJSON.FinalAge {
					return nil, fmt.Errorf("report %d played a different run (prestige %q, %d cycles, final %q) from the first (%q, %d, %q)",
						i+1, prog.Config.PrestigeAge, prog.Config.Cycles, prog.Config.FinalAge, cfgJSON.PrestigeAge, cfgJSON.Cycles, cfgJSON.FinalAge)
				}
				for _, run := range prog.Runs {
					if seen[run.Seed] {
						return nil, fmt.Errorf("seed %d appears in more than one report", run.Seed)
					}
					seen[run.Seed] = true
					runs = append(runs, run)
				}
			case r.Name == "progression":
				return nil, fmt.Errorf("report %d has a progression result with no runs", i+1)
			default:
				if j, dup := carriedFrom[r.Name]; dup {
					return nil, fmt.Errorf("scenario %s appears in reports %d and %d", r.Name, j, i+1)
				}
				carried[r.Name], carriedFrom[r.Name] = r, i+1
			}
		}
		if p.Started.Before(started) {
			started = p.Started
		}
		wall = max(wall, p.WallMs)
	}

	sess := &Session{Tier: parts[0].Tier, Pacing: e.Pacing, Started: started, WallMs: wall}
	for _, sc := range Scenarios() {
		var r *Result
		switch {
		case sc.Name == "static" && static:
			r = &Result{Name: "static"}
			runStatic(e, r)
		case sc.Name == "progression" && len(runs) > 0:
			sort.Slice(runs, func(i, j int) bool { return runs[i].Seed < runs[j].Seed })
			cfg := configFromJSON(*cfgJSON)
			cfg.Pacing = e.Pacing
			cfg.Seeds = nil
			for _, run := range runs {
				cfg.Seeds = append(cfg.Seeds, run.Seed)
			}
			r = &Result{Name: "progression", WallMs: progWall}
			describeProgression(r, foldBotSet(e, r, "progression", "progression", cfg, started, runs))
			r.Summary = fmt.Sprintf("merged from %d report(s): %s", len(parts), r.Summary)
		case carried[sc.Name] != nil:
			r = carried[sc.Name]
			r.section("Details", "Run in its own CI job (report %d of the merge): its tables and per-seed files are in that job's artifact.", carriedFrom[sc.Name])
			if r.Status == StatusFail {
				sess.Failed = true
			}
			sess.Scenarios = append(sess.Scenarios, r)
			continue
		default:
			continue
		}
		r.Status = StatusPass
		if len(r.Failures) > 0 {
			r.Status = StatusFail
			sess.Failed = true
		}
		sess.Scenarios = append(sess.Scenarios, r)
	}
	if len(sess.Scenarios) == 0 {
		return nil, fmt.Errorf("the %d report(s) hold no scenarios", len(parts))
	}
	return sess, nil
}
