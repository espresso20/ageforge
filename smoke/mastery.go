package smoke

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// Era Mastery in the smoke suite (Pacing v2, PR 5).
//
//   - Presets start a run as a returning player (applyPreset): the veteran
//     has mastery 10 through the Space Age and a record in the Interstellar
//     Age, the returning player mastery 2 through the Information Age, 1 on
//     the Digital Age and a record in the Cyberpunk Age.
//   - Ages are graded against their target ÷ k (VerdictK). On known ground
//     the Primitive and Stone Ages are graded together instead
//     (VeteranEarlyMax): at k = 4 they are minutes long, and the bot's
//     decision interval and the one-tick floors bind there, not the game.
//   - The veteran scenario plays the veteran preset: to the Iron Age on one
//     seed per PR, to a Modern Age prestige on three seeds nightly, held to
//     VeteranRunLow to VeteranRunHigh.
//   - Push cycles (Config.PushCycles): cycle 2 plays for as long as cycle 1
//     took, then prestiges. It must cover cycle 1's ages LaterRunMinSpeedup
//     times faster and end at least LaterRunMinDepth ages deeper
//     (LaterRunRow).

// Presets (Config.Preset, -preset).
const (
	PresetVeteran   = "veteran"
	PresetReturning = "returning"
)

// PresetNames lists the presets for flag help.
func PresetNames() []string { return []string{PresetVeteran, PresetReturning} }

// presetMastery is a preset's mastery map and record.
func presetMastery(preset string) (map[string]int, string, bool) {
	m := map[string]int{}
	set := func(through string, level int) {
		for _, a := range config.AgeOrder() {
			if _, ok := m[a]; !ok {
				m[a] = level
			}
			if a == through {
				return
			}
		}
	}
	switch preset {
	case PresetVeteran:
		set("space_age", config.MasteryCap)
		return m, "interstellar_age", true
	case PresetReturning:
		set("information_age", 2)
		set("digital_age", 1)
		return m, "cyberpunk_age", true
	}
	return nil, "", false
}

// applyPreset gives ge the preset's mastery and record ("" leaves it new).
func applyPreset(ge *game.GameEngine, preset string) {
	if m, rec, ok := presetMastery(preset); ok {
		ge.SetMasteryForTest(m, rec)
	}
}

// ValidPreset reports whether preset is "" or a known preset.
func ValidPreset(preset string) bool {
	_, _, ok := presetMastery(preset)
	return preset == "" || ok
}

// VerdictK grades secs spent in age at Era Mastery speed k against its
// target ÷ k (k ≤ 1 is the frontier: Verdict). On known ground the
// Primitive and Stone Ages get no verdict of their own (VeteranEarlyMax
// grades them together).
func VerdictK(age string, secs float64, finished bool, k float64) string {
	if k <= 1 {
		return Verdict(age, secs, finished)
	}
	if earlyAge(age) {
		return VerdictNone
	}
	return Verdict(age, float64(secs*k), finished)
}

// earlyAge reports whether age is one VeteranEarlyMax grades.
func earlyAge(age string) bool { return age == "primitive_age" || age == "stone_age" }

// VeteranEarlyMax is the most the Primitive and Stone Ages together may take
// on known ground (the median across seeds, cycle 1).
const VeteranEarlyMax = time.Hour

// The veteran preset's run to the Modern Age (cycle 1, median across seeds):
// about 1.25 days measured in the plan's stand-in, at k = 4.16.
const (
	VeteranRunLow  = 26*time.Hour + 24*time.Minute // 1.1 days
	VeteranRunHigh = 36 * time.Hour                // 1.5 days
)

// firstRunBand is the band cfg's first run to the Modern Age is held to, and
// whether it is graded at all: the one-week band for a new player, the
// veteran band for the veteran preset, none for the returning preset
// (reported only).
func firstRunBand(cfg Config) (lo, hi time.Duration, graded bool) {
	switch cfg.Preset {
	case PresetVeteran:
		return VeteranRunLow, VeteranRunHigh, true
	case PresetReturning:
		return 0, 0, false
	}
	return FirstRunLow, FirstRunHigh, true
}

// EarlyRow is the Primitive and Stone Ages together on known ground (cycle
// 1), across seeds.
type EarlyRow struct {
	Samples    int     `json:"samples"`
	MedianSecs float64 `json:"median_seconds"`
	MaxSecs    float64 `json:"max_seconds"`
	Verdict    string  `json:"verdict"`
}

// newEarly grades the Primitive and Stone Ages of cycle 1 when they ran on
// known ground; nil when they did not (a new player) or no run finished
// them.
func newEarly(runs []*RunResult) *EarlyRow {
	var got []float64
	for _, r := range runs {
		secs, n, known := 0.0, 0, false
		for _, a := range r.Ages {
			if a.Cycle != 1 || !earlyAge(a.Age) || a.Unfinished {
				continue
			}
			secs += a.Seconds
			n++
			known = known || a.K > 1
		}
		if n == 2 && known {
			got = append(got, secs)
		}
	}
	if len(got) == 0 {
		return nil
	}
	_, med, hi := spread(got)
	row := &EarlyRow{Samples: len(got), MedianSecs: med, MaxSecs: hi, Verdict: VerdictOK}
	if med > VeteranEarlyMax.Seconds() {
		row.Verdict = VerdictSlow
	}
	return row
}

// LaterRunMinSpeedup and LaterRunMinDepth are the push cycle's bar: cycle 2
// covers cycle 1's ages at least this many times faster, and ends at least
// this many ages deeper (medians across seeds).
const (
	LaterRunMinSpeedup = 1.9
	LaterRunMinDepth   = 1
)

// LaterRunSeed is one seed's cycle 1 against its cycle 2 push.
type LaterRunSeed struct {
	Seed int64 `json:"seed"`
	// Cycle1Secs is cycle 1's time to the age it prestiged from (Cycle1Age);
	// Cycle2Secs is cycle 2's time to reach that same age (-1 if it never
	// did). Cycle2Age is where cycle 2 ended.
	Cycle1Age  string  `json:"cycle1_age"`
	Cycle1Secs float64 `json:"cycle1_seconds"`
	Cycle2Secs float64 `json:"cycle2_seconds"`
	Cycle2Age  string  `json:"cycle2_age"`
	Speedup    float64 `json:"speedup"`
	Depth      int     `json:"depth"`
}

// LaterRunRow grades the push cycles across seeds.
type LaterRunRow struct {
	Seeds         []LaterRunSeed `json:"seeds"`
	MedianSpeedup float64        `json:"median_speedup"`
	MedianDepth   float64        `json:"median_depth"`
	Failed        bool           `json:"failed,omitempty"`
}

// newLaterRun compares each run's cycle 1 with its cycle 2; nil when no run
// finished a first cycle.
func newLaterRun(runs []*RunResult, order map[string]int) *LaterRunRow {
	row := &LaterRunRow{}
	var speeds, depths []float64
	for _, r := range runs {
		if len(r.Cycles) == 0 {
			continue
		}
		c1 := r.Cycles[0]
		s := LaterRunSeed{Seed: r.Seed, Cycle1Age: c1.FinalAge, Cycle1Secs: c1.Seconds, Cycle2Secs: -1}
		// Cycle 2's time to cycle 1's last age, and where it ended.
		c2, reached := 0.0, false
		for _, a := range r.Ages {
			if a.Cycle != 2 {
				continue
			}
			if order[a.Age] >= order[c1.FinalAge] {
				reached = true
				break
			}
			c2 += a.Seconds
		}
		if reached {
			s.Cycle2Secs = c2
			if c2 > 0 {
				s.Speedup = c1.Seconds / c2
			}
		}
		s.Cycle2Age = r.FinalAge
		if len(r.Cycles) > 1 {
			s.Cycle2Age = r.Cycles[1].FinalAge
		}
		s.Depth = order[s.Cycle2Age] - order[c1.FinalAge]
		row.Seeds = append(row.Seeds, s)
		speeds = append(speeds, s.Speedup)
		depths = append(depths, float64(s.Depth))
	}
	if len(row.Seeds) == 0 {
		return nil
	}
	_, row.MedianSpeedup, _ = spread(speeds)
	_, row.MedianDepth, _ = spread(depths)
	row.Failed = row.MedianSpeedup < LaterRunMinSpeedup || row.MedianDepth < LaterRunMinDepth
	return row
}

// writeLaterRun renders the push-cycle table.
func (s *Summary) writeLaterRun(sb *strings.Builder) {
	l := s.LaterRun
	if l == nil {
		return
	}
	fmt.Fprintf(sb, "\nLater runs (cycle 2 plays as long as cycle 1, then prestiges): median %.2fx faster over cycle 1's ages (want %gx or more), median %+.1f ages deeper (want %+d or more): %s.\n\n",
		l.MedianSpeedup, LaterRunMinSpeedup, l.MedianDepth, LaterRunMinDepth, verdictMark(map[bool]string{true: VerdictSlow, false: VerdictOK}[l.Failed]))
	sb.WriteString("| seed | cycle 1 to | cycle 1 | cycle 2 to the same age | speed-up | cycle 2 ended in | deeper by |\n|---|---|---|---|---|---|---|\n")
	for _, x := range l.Seeds {
		c2 := "never"
		if x.Cycle2Secs >= 0 {
			c2 = days(x.Cycle2Secs)
		}
		fmt.Fprintf(sb, "| %d | %s | %s | %s | %.2fx | %s | %+d |\n", x.Seed, x.Cycle1Age, days(x.Cycle1Secs), c2, x.Speedup, x.Cycle2Age, x.Depth)
	}
}

// writeEarly renders the known-ground Primitive and Stone line.
func (s *Summary) writeEarly(sb *strings.Builder) {
	e := s.Early
	if e == nil {
		return
	}
	fmt.Fprintf(sb, "\nPrimitive and Stone Ages on known ground: median %s (max %s, %d seeds), limit %s: %s.\n",
		dur(e.MedianSecs), dur(e.MaxSecs), e.Samples, dur(VeteranEarlyMax.Seconds()), verdictMark(e.Verdict))
}

// veteranConfig is the veteran scenario's run for the tier: to the Iron Age
// on one seed per PR (seconds of wall time), to a Modern Age prestige on
// three seeds nightly.
func veteranConfig(e *Env) (Config, []int64) {
	cfg := e.Base
	cfg.Preset = PresetVeteran
	cfg.Cycles, cfg.FinalAge, cfg.PushCycles = 1, "", false
	if e.full() {
		cfg.MaxSim = 200 * time.Hour
		return cfg, e.seeds(3)
	}
	cfg.StopAge, cfg.MaxSim = "iron_age", 50*time.Hour
	return cfg, e.seeds(1)
}

func runVeteran(e *Env, res *Result) {
	cfg, seeds := veteranConfig(e)
	sum := runBotSet(e, res, "veteran", "veteran", cfg, seeds)
	var sb strings.Builder
	sum.writePacingTable(&sb)
	sum.writeEarly(&sb)
	sum.writeFirstRun(&sb)
	res.section("Veteran pacing (mastery 10 through the Space Age, record Interstellar)", "%s", sb.String())
	var ends []string
	for _, r := range sum.Runs {
		ends = append(ends, fmt.Sprintf("seed %d %s at %s", r.Seed, r.Outcome, r.FinalAge))
	}
	sort.Strings(ends)
	res.Summary = strings.Join(ends, "; ")
	if f := sum.FirstRun; f != nil && f.MedianSecs >= 0 {
		res.Summary += fmt.Sprintf("; to the Modern Age %s (%s)", days(f.MedianSecs), verdictMark(f.Verdict))
	}
	if x := sum.Early; x != nil {
		res.Summary += fmt.Sprintf("; Primitive and Stone %s (%s)", dur(x.MedianSecs), verdictMark(x.Verdict))
	}
	// The legacy kit after a scripted prestige, and the shop refund: a
	// broken kit or refund fails every PR, not just the nightly.
	var steps []string
	seed := e.SeedBase
	check := func(name string, ok bool, format string, args ...interface{}) bool {
		if ok {
			steps = append(steps, fmt.Sprintf("| %s | ok |", name))
			return true
		}
		msg := fmt.Sprintf(format, args...)
		steps = append(steps, fmt.Sprintf("| %s | %s |", name, cell(msg)))
		f := res.fail("veteran_"+strings.ReplaceAll(name, " ", "_"), "%s", msg)
		f.Seed, f.Repro = seed, fmt.Sprintf("go run ./cmd/smoke -scenario veteran -seed-base %d -v", seed)
		return false
	}
	prestigeKitHooked(e, check)
	prestigeRefundHooked(check)
	res.section("Legacy kit and shop refund (test hooks, not play)", "| step | result |\n|---|---|\n%s", strings.Join(steps, "\n"))
}
