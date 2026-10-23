package game

import (
	"fmt"
	"testing"
)

// The merge helpers size their result from the two inputs. These tests pin what they
// return (CodeQL's go/allocation-size-overflow flagged the old len(a)+len(b) hints) and
// that a large pair of overlapping copies, such as a backup of your own account, is
// folded without reserving the sum of both.

func daysRange(from, to int) []string {
	out := make([]string, 0, to-from)
	for i := from; i < to; i++ {
		out = append(out, fmt.Sprintf("day-%06d", i))
	}
	return out
}

func TestUnionStringsKeepsOrderAndDedupes(t *testing.T) {
	got := unionStrings([]string{"x", "y", "x"}, []string{"y", "z", "x", "w"})
	want := []string{"x", "y", "z", "w"}
	if len(got) != len(want) {
		t.Fatalf("union = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("union = %v, want %v", got, want)
		}
	}
	if out := unionStrings(nil, nil); len(out) != 0 {
		t.Fatalf("union of nothing = %v, want empty", out)
	}
}

func TestUnionStringsOverlappingCopiesDoNotReserveTheSum(t *testing.T) {
	const n = 100_000
	a, b := daysRange(0, n), daysRange(0, n)
	got := unionStrings(a, b)
	if len(got) != n {
		t.Fatalf("union of two identical %d-entry lists has %d entries", n, len(got))
	}
	if cap(got) >= 2*n {
		t.Fatalf("capacity %d reserves the sum of both inputs (%d)", cap(got), 2*n)
	}
}

func TestUnionStringsDisjointLargeInputs(t *testing.T) {
	const n = 50_000
	got := unionStrings(daysRange(0, n), daysRange(n, 2*n))
	if len(got) != 2*n {
		t.Fatalf("union of two disjoint %d-entry lists has %d entries, want %d", n, len(got), 2*n)
	}
	if got[0] != "day-000000" || got[len(got)-1] != fmt.Sprintf("day-%06d", 2*n-1) {
		t.Fatalf("union lost its order: first %q last %q", got[0], got[len(got)-1])
	}
}

func TestMergeDaysIsBoundedAfterTheUnion(t *testing.T) {
	// Far more days than the account keeps, from both copies: the newest maxAccountDays stay.
	got := mergeDays(daysRange(0, 5_000), daysRange(2_500, 7_500))
	if len(got) != maxAccountDays {
		t.Fatalf("merged days = %d, want %d", len(got), maxAccountDays)
	}
	if want := fmt.Sprintf("day-%06d", 7_499); got[len(got)-1] != want {
		t.Fatalf("newest day = %q, want %q", got[len(got)-1], want)
	}
	if mergeDays(nil, nil) != nil {
		t.Fatalf("merging no days should give nil")
	}
}

func TestMergeBadgesKeepsTheEarlierCopy(t *testing.T) {
	a := map[string]BadgeEarned{"one": {At: 200}, "both": {At: 300}}
	b := map[string]BadgeEarned{"two": {At: 50}, "both": {At: 100, Count: 0}}
	got := mergeBadges(a, b)
	if len(got) != 3 {
		t.Fatalf("merged badges = %v, want three", got)
	}
	if got["both"].At != 100 {
		t.Fatalf("both kept At=%d, want the earlier 100", got["both"].At)
	}
	if mergeBadges(nil, nil) != nil {
		t.Fatalf("merging no badges should give nil")
	}
}

func TestMergeBadgesLargeOverlappingCopies(t *testing.T) {
	const n = 50_000
	a := make(map[string]BadgeEarned, n)
	b := make(map[string]BadgeEarned, n)
	for i := 0; i < n; i++ {
		k := fmt.Sprintf("badge-%d", i)
		a[k] = BadgeEarned{At: int64(i + 10)}
		b[k] = BadgeEarned{At: int64(i + 5)}
	}
	got := mergeBadges(a, b)
	if len(got) != n {
		t.Fatalf("merged %d badges, want %d", len(got), n)
	}
	if got["badge-7"].At != 12 {
		t.Fatalf("badge-7 At = %d, want the earlier 12", got["badge-7"].At)
	}
}

func TestMergeCountersTakesTheLargerNotTheSum(t *testing.T) {
	got := mergeCounters(map[string]float64{"a": 3, "b": 9}, map[string]float64{"a": 5, "c": 1})
	if len(got) != 3 || got["a"] != 5 || got["b"] != 9 || got["c"] != 1 {
		t.Fatalf("merged counters = %v", got)
	}
	if mergeCounters(nil, nil) != nil {
		t.Fatalf("merging no counters should give nil")
	}
}
