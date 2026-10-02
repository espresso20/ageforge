package smoke

import (
	"time"

	"github.com/espresso20/ageforge/config"
)

// Pacing targets: how long a player should spend in each age at 1x game
// time, from entering it to entering the next. This table is the pacing
// contract the balance work aims at (the one-week curve: about 7 days of play
// to the Modern Age, the first prestige). It mirrors config.AgeTargets, which
// the game derives its economy from (the base curve × config.PacingStretch
// from the Bronze Age on): edit that one, then make this match
// (TestPacingTargetsMatchConfig). The report, the per-age timeouts and
// `-pacing enforce` all read this copy. See CONTRIBUTING.md, "Pacing targets".
var PacingTargets = map[string]time.Duration{
	"primitive_age":    15 * time.Minute,
	"stone_age":        45 * time.Minute,
	"bronze_age":       3*time.Hour + 54*time.Minute,
	"iron_age":         6*time.Hour + 30*time.Minute,
	"classical_age":    9*time.Hour + 6*time.Minute,
	"medieval_age":     11*time.Hour + 42*time.Minute,
	"renaissance_age":  15*time.Hour + 36*time.Minute,
	"colonial_age":     18*time.Hour + 12*time.Minute,
	"industrial_age":   20*time.Hour + 48*time.Minute,
	"victorian_age":    23*time.Hour + 24*time.Minute,
	"electric_age":     26 * time.Hour,
	"atomic_age":       31*time.Hour + 12*time.Minute,
	"modern_age":       31*time.Hour + 12*time.Minute,
	"information_age":  36*time.Hour + 24*time.Minute,
	"digital_age":      41*time.Hour + 36*time.Minute,
	"cyberpunk_age":    46*time.Hour + 48*time.Minute,
	"fusion_age":       52 * time.Hour,
	"space_age":        57*time.Hour + 12*time.Minute,
	"interstellar_age": 62*time.Hour + 24*time.Minute,
	"galactic_age":     62*time.Hour + 24*time.Minute,
	"quantum_age":      62*time.Hour + 24*time.Minute,
	// transcendent_age is the last age: nothing to pace.
}

// The first run to the Modern Age, for the greedy bot (the median across the
// seeds of the progression scenario): FirstRunLow to FirstRunHigh of 1x play.
// The targets sum to about 7 days; the bot, a near-perfect player, lands near
// 5.3 (build and research times are capped copies of hand-typed values, so
// they grow less than the targets). Under -pacing enforce a median outside the
// band fails the set, like an age outside its own band.
const (
	FirstRunLow  = 115*time.Hour + 12*time.Minute // 4.8 days
	FirstRunHigh = 148*time.Hour + 48*time.Minute // 6.2 days
)

// The tolerance band: an age passes when its time is within
// [PacingLow, PacingHigh] x its target.
const (
	PacingLow  = 0.5
	PacingHigh = 2.0
)

// Per-age timeout, derived from the table: TimeoutFactor x target, never
// less than TimeoutFloor. Ages without a target use DefaultAgeTimeout.
const (
	TimeoutFactor     = 4.0
	TimeoutFloor      = time.Hour
	DefaultAgeTimeout = 48 * time.Hour
)

// Pacing modes (-pacing).
const (
	// PacingReport reports every age as ok, slow or fast and never fails a
	// run over it; an age past its timeout is flagged and play continues.
	PacingReport = "report"
	// PacingEnforce fails a run on any age outside the band, and stops it at
	// an age's timeout.
	PacingEnforce = "enforce"
)

// Pacing verdicts.
const (
	VerdictOK   = "ok"
	VerdictSlow = "slow"
	VerdictFast = "fast"
	// VerdictNone is an age with no target (the final age).
	VerdictNone = ""
)

// Target returns the pacing target for age and whether it has one.
func Target(age string) (time.Duration, bool) {
	d, ok := PacingTargets[age]
	return d, ok
}

// Verdict grades secs spent in age against its target. An unfinished age
// can only be slow (already past the band) or undecided (VerdictNone).
func Verdict(age string, secs float64, finished bool) string {
	t, ok := Target(age)
	if !ok || t <= 0 {
		return VerdictNone
	}
	ratio := secs / t.Seconds()
	switch {
	case ratio > PacingHigh:
		return VerdictSlow
	case !finished:
		return VerdictNone
	case ratio < PacingLow:
		return VerdictFast
	default:
		return VerdictOK
	}
}

// AgeTimeout is how long a run may stay in age before the age counts as
// timed out: TimeoutFactor x its target, at least TimeoutFloor.
func AgeTimeout(age string) time.Duration {
	t, ok := Target(age)
	if !ok {
		return DefaultAgeTimeout
	}
	d := time.Duration(float64(t) * TimeoutFactor)
	if d < TimeoutFloor {
		d = TimeoutFloor
	}
	return d
}

// CumulativeTarget is the summed target from the Primitive Age up to (not
// including) age: the expected 1x time to reach it.
func CumulativeTarget(age string) time.Duration {
	var sum time.Duration
	for _, a := range config.AgeOrder() {
		if a == age {
			break
		}
		sum += PacingTargets[a]
	}
	return sum
}

// verdictMark renders a verdict for a report table.
func verdictMark(v string) string {
	switch v {
	case VerdictOK:
		return "✓"
	case VerdictSlow:
		return "slow"
	case VerdictFast:
		return "fast"
	default:
		return "-"
	}
}
