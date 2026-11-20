package planboard

import (
	"math"
	"time"
)

// The ruler's ladder: a mark at each of these, cut at the first one past the
// finish.
var (
	ladH = []float64{0, 1. / 60, 1. / 6, 1, 6, 24, 72, 168, 504}
	// LadderNames are the marks' labels, in the order of the ladder.
	LadderNames = []string{"now", "1m", "10m", "1h", "6h", "1d", "3d", "1w", "3w"}
)

// axis puts a duration on the sheet: a step of the ladder is as wide as the
// starts and ends inside it ask for (14 cells at least, 5 when empty), and
// inside a step the scale is linear.
type axis struct{ xs []int } // pixel of each mark

func newAxis(marks []time.Duration, finish time.Duration, x0, cells, cw int) axis {
	top := 1
	for top < len(ladH)-1 && ladH[top] <= finish.Hours() {
		top++
	}
	step := func(d time.Duration) int {
		i := 0
		for i < top-1 && d.Hours() > ladH[i+1] {
			i++
		}
		return i
	}
	cnt := make([]int, top)
	tot := 0
	for _, d := range append(marks, finish) {
		cnt[step(d)]++
		tot++
	}
	w := make([]int, top)
	for i, c := range cnt {
		w[i] = 5
		if c > 0 {
			w[i] = 14
		}
	}
	sum := 0
	for _, v := range w {
		sum += v
	}
	if extra := cells - sum; extra > 0 {
		last := 0
		for i, c := range cnt {
			if c > 0 {
				w[i] += extra * c / tot
				last = i
			}
		}
		sum = 0
		for _, v := range w {
			sum += v
		}
		w[last] += cells - sum
	}
	for sum = 0; ; sum = 0 { // too narrow: shave the widest steps
		for _, v := range w {
			sum += v
		}
		if sum <= cells {
			break
		}
		wi := 0
		for i, v := range w {
			if v > w[wi] {
				wi = i
			}
		}
		w[wi]--
	}
	a := axis{[]int{x0 * cw}}
	for _, v := range w {
		a.xs = append(a.xs, a.xs[len(a.xs)-1]+v*cw)
	}
	return a
}

func (a axis) x(d time.Duration) int {
	h, top := d.Hours(), len(a.xs)-1
	i := 0
	for i < top-1 && h > ladH[i+1] {
		i++
	}
	f := math.Max(0, math.Min(1, (h-ladH[i])/(ladH[i+1]-ladH[i])))
	return a.xs[i] + int(f*float64(a.xs[i+1]-a.xs[i]))
}

func (a axis) end() int { return a.xs[len(a.xs)-1] }
