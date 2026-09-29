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
	}
	for in, want := range cases {
		if got := Number(in); got != want {
			t.Errorf("Number(%v) = %q, want %q", in, got, want)
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
