package config

import "math"

// Era Mastery (Pacing v2, PR 5): each age remembers how many runs completed
// it, and an age you know runs faster. An age's mastery m is the number of
// runs that completed it and then ended, in a prestige or a Succumb, capped
// at MasteryCap; it runs k times
// faster, k = 1 + √m: production × k (after the pools), storage × k, build
// and research times ÷ k. The frontier (m = 0) runs at 1x. Events, raids and
// the other real-clock timers are never divided.
//
// Catch-up: an age CatchUpGap or more behind the player's record (the deepest
// age ever entered) runs at CatchUpK, or at its own k if that is higher, so a
// veteran's early epochs take hours, not days.
const (
	// MasteryCap is the most mastery an age can hold.
	MasteryCap = 10
	// CatchUpGap is how many ages behind the record an age must be for
	// catch-up to apply.
	CatchUpGap = 6
	// CatchUpK is the speed catch-up gives. A tuning constant: k = 4 brings
	// a veteran's first seven ages to about ten hours.
	CatchUpK = 4.0
)

// masteryK is k for m = 0..MasteryCap: 1, 2, 2.41, ... 4.16. math.Sqrt is
// correctly rounded on every CPU, so the table is the same everywhere.
var masteryK = func() [MasteryCap + 1]float64 {
	var t [MasteryCap + 1]float64
	for m := range t {
		t[m] = 1 + math.Sqrt(float64(m))
	}
	return t
}()

// ClampMastery holds m to 0..MasteryCap.
func ClampMastery(m int) int {
	return min(max(m, 0), MasteryCap)
}

// MasteryK is the speed of an age with mastery m: 1 + √m, with m clamped to
// 0..MasteryCap.
func MasteryK(m int) float64 {
	return masteryK[ClampMastery(m)]
}

// AgeSpeed is the speed of an age with mastery m that sits behind ages
// behind the record: MasteryK(m), raised to CatchUpK when behind is at least
// CatchUpGap.
func AgeSpeed(m, behind int) float64 {
	k := MasteryK(m)
	if behind >= CatchUpGap && k < CatchUpK {
		return CatchUpK
	}
	return k
}
