package textfmt

import (
	"testing"
	"time"
)

func TestNumber(t *testing.T) {
	cases := map[float64]string{
		0:          "0",
		950:        "950",
		12.5:       "12.5",
		0.04:       "0.04",
		-3:         "-3",
		1000:       "1K",
		1500:       "1.5K",
		0.7111:     "0.711",
		44.83:      "44.8",
		999600:     "1M",
		12500:      "12.5K",
		966000:     "966K",
		1234567:    "1.23M",
		876000000:  "876M",
		2.5e9:      "2.5B",
		999.96:     "1K",
		1.5e15:     "1.5Q",
		-1234567.0: "-1.23M",
		7.73e18:    "7.73Qi",
		1.5e18:     "1.5Qi",
		2.5e19:     "25Qi",
		-1.18e19:   "-11.8Qi",
	}
	for in, want := range cases {
		if got := Number(in); got != want {
			t.Errorf("Number(%v) = %q, want %q", in, got, want)
		}
	}
}

// TestNumberUnitBoundaries: every unit starts at its own power of a thousand,
// the amount just under it still reads in the unit below, and a figure that
// rounds up to 1000 of a unit rolls over to 1 of the next.
func TestNumberUnitBoundaries(t *testing.T) {
	want := []struct {
		power  float64
		suffix string
	}{
		{1e3, "K"}, {1e6, "M"}, {1e9, "B"}, {1e12, "T"}, {1e15, "Q"},
		{1e18, "Qi"}, {1e21, "Sx"}, {1e24, "Sp"}, {1e27, "Oc"}, {1e30, "No"}, {1e33, "Dc"},
	}
	if len(want) != len(units) {
		t.Fatalf("the table has %d units, the test knows %d", len(units), len(want))
	}
	prev := ""
	for i, w := range want {
		if units[i].threshold != w.power || units[i].suffix != w.suffix {
			t.Errorf("unit %d is %v %q, want %v %q", i, units[i].threshold, units[i].suffix, w.power, w.suffix)
		}
		if got := Number(w.power); got != "1"+w.suffix {
			t.Errorf("Number(%v) = %q, want %q", w.power, got, "1"+w.suffix)
		}
		if got := Number(999 * w.power); got != "999"+w.suffix {
			t.Errorf("Number(999 * %v) = %q, want %q", w.power, got, "999"+w.suffix)
		}
		if got := Number(w.power * 1.5); got != "1.5"+w.suffix {
			t.Errorf("Number(1.5 * %v) = %q, want %q", w.power, got, "1.5"+w.suffix)
		}
		if i > 0 {
			// Just under the unit: 999.6 of the unit below rounds to 1000
			// and moves up, 999.4 of it stays.
			if got := Number(999.6 * want[i-1].power); got != "1"+w.suffix {
				t.Errorf("Number(999.6 * %v) = %q, want %q", want[i-1].power, got, "1"+w.suffix)
			}
			if got := Number(999.4 * want[i-1].power); got != "999"+want[i-1].suffix {
				t.Errorf("Number(999.4 * %v) = %q, want %q", want[i-1].power, got, "999"+want[i-1].suffix)
			}
		}
		if prev != "" && prev == w.suffix {
			t.Errorf("two units share the suffix %q", w.suffix)
		}
		prev = w.suffix
	}
	if got := Number(-2.5e21); got != "-2.5Sx" {
		t.Errorf("Number(-2.5e21) = %q", got)
	}
}

// TestMaxNumber: MaxNumber is the last value that prints in three digits, one
// step over it is the first that does not, and nothing under it ever prints
// more than three digits before its unit, whatever the unit.
func TestMaxNumber(t *testing.T) {
	max := MaxNumber()
	if got := Number(max); got != "999Dc" {
		t.Errorf("Number(MaxNumber()) = %q, want 999Dc", got)
	}
	if got := Number(1000 * units[len(units)-1].threshold); got != "1000Dc" {
		t.Errorf("Number(1000Dc) = %q, want 1000Dc (the first value past the table)", got)
	}
	// The late game's stores reach about 25 quintillion; the table runs
	// well past that, by at least three units.
	const topStore = 2.5e19
	if max < topStore*1e9 {
		t.Errorf("MaxNumber() = %v is not three units past the late game's top store %v", max, topStore)
	}
	for v := 0.5; v <= max; v *= 1.37 {
		got := Number(v)
		digits := 0
		for _, r := range got {
			if r == '.' {
				break
			}
			if r >= '0' && r <= '9' {
				digits++
			}
		}
		if digits > 3 {
			t.Fatalf("Number(%v) = %q, more than three digits before the unit", v, got)
		}
	}
}

func TestRate(t *testing.T) {
	cases := map[float64]string{
		0:       "+0/tick",
		3.25:    "+3.25/tick",
		-0.12:   "-0.12/tick",
		0.004:   "+0.004/tick",
		1500:    "+1.5K/tick",
		50:      "+50/tick",
		12.5:    "+12.5/tick",
		-2000.0: "-2K/tick",
	}
	for in, want := range cases {
		if got := Rate(in); got != want {
			t.Errorf("Rate(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestCountAndPlural(t *testing.T) {
	if got := Count(1, "item", "items"); got != "1 item" {
		t.Errorf("got %q", got)
	}
	if got := Count(3, "worker", "workers"); got != "3 workers" {
		t.Errorf("got %q", got)
	}
	if got := Count(0, "worker", "workers"); got != "0 workers" {
		t.Errorf("got %q", got)
	}
}

func TestDurationAndTicks(t *testing.T) {
	if got := Duration(284 * time.Second); got != "4m 44s" {
		t.Errorf("got %q", got)
	}
	if got := Ticks(120, time.Second); got != "~2m" {
		t.Errorf("got %q", got)
	}
	if got := Duration(0); got != "0s" {
		t.Errorf("got %q", got)
	}
}

func TestSentenceAndList(t *testing.T) {
	if got := Sentence("unknown resource: x"); got != "Unknown resource: x." {
		t.Errorf("got %q", got)
	}
	if got := List([]string{"a", "b", "c"}); got != "a, b and c" {
		t.Errorf("got %q", got)
	}
}
