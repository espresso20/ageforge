package world

import "github.com/espresso20/ageforge/mapmodel"

// Whole-number maths shared by the sky: the same seed must give the same
// bits on every machine, so nothing here touches a float. Angles are 4096ths
// of a turn and sin comes from a table.

// sinTable is sin over a quarter-turn-by-16 grid: entry i is sin(i/256 of a
// turn) scaled by 1024; the last entry closes the circle.
var sinTable = [257]int{
	0, 25, 50, 75, 100, 125, 150, 175, 200, 224, 249, 273, 297, 321, 345, 369,
	392, 415, 438, 460, 483, 505, 526, 548, 569, 590, 610, 630, 650, 669, 688, 706,
	724, 742, 759, 775, 792, 807, 822, 837, 851, 865, 878, 891, 903, 915, 926, 936,
	946, 955, 964, 972, 980, 987, 993, 999, 1004, 1009, 1013, 1016, 1019, 1021, 1023, 1024,
	1024, 1024, 1023, 1021, 1019, 1016, 1013, 1009, 1004, 999, 993, 987, 980, 972, 964, 955,
	946, 936, 926, 915, 903, 891, 878, 865, 851, 837, 822, 807, 792, 775, 759, 742,
	724, 706, 688, 669, 650, 630, 610, 590, 569, 548, 526, 505, 483, 460, 438, 415,
	392, 369, 345, 321, 297, 273, 249, 224, 200, 175, 150, 125, 100, 75, 50, 25,
	0, -25, -50, -75, -100, -125, -150, -175, -200, -224, -249, -273, -297, -321, -345, -369,
	-392, -415, -438, -460, -483, -505, -526, -548, -569, -590, -610, -630, -650, -669, -688, -706,
	-724, -742, -759, -775, -792, -807, -822, -837, -851, -865, -878, -891, -903, -915, -926, -936,
	-946, -955, -964, -972, -980, -987, -993, -999, -1004, -1009, -1013, -1016, -1019, -1021, -1023, -1024,
	-1024, -1024, -1023, -1021, -1019, -1016, -1013, -1009, -1004, -999, -993, -987, -980, -972, -964, -955,
	-946, -936, -926, -915, -903, -891, -878, -865, -851, -837, -822, -807, -792, -775, -759, -742,
	-724, -706, -688, -669, -650, -630, -610, -590, -569, -548, -526, -505, -483, -460, -438, -415,
	-392, -369, -345, -321, -297, -273, -249, -224, -200, -175, -150, -125, -100, -75, -50, -25,
	0,
}

// Turn is a full circle in the angle unit.
const Turn = 4096

// Sin4096 is sin(a) scaled by 1024, a in 4096ths of a turn.
func Sin4096(a int) int {
	a &= Turn - 1
	i, f := a>>4, a&15
	return sinTable[i] + (sinTable[i+1]-sinTable[i])*f/16
}

// Cos4096 is cos(a) scaled by 1024.
func Cos4096(a int) int { return Sin4096(a + Turn/4) }

// Roll is a stable pick in [0, n) from the hash of vals.
func Roll(n int, vals ...int64) int { return int(mapmodel.Hash(vals...) % uint64(n)) }

// Isqrt is the integer square root.
func Isqrt(n int) int {
	if n <= 0 {
		return 0
	}
	r := n
	for r*r > n {
		r = (r + n/r) / 2
	}
	for (r+1)*(r+1) <= n {
		r++
	}
	return r
}

// Sq is v squared.
func Sq(v int) int { return v * v }

// Bump is a smooth hill: 256 at d2 of nothing, none at rad2.
func Bump(d2, rad2 int) int {
	if d2 >= rad2 {
		return 0
	}
	t := 256 - d2*256/rad2
	return t * t >> 8
}
