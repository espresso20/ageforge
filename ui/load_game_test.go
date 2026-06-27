package ui

import (
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
)

func TestCommafy(t *testing.T) {
	cases := map[int]string{
		0:       "0",
		7:       "7",
		42:      "42",
		999:     "999",
		1000:    "1,000",
		12400:   "12,400",
		100000:  "100,000",
		1234567: "1,234,567",
		-1500:   "-1,500",
		-12:     "-12",
	}
	for in, want := range cases {
		if got := commafy(in); got != want {
			t.Errorf("commafy(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestRelativeTime(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name string
		t    time.Time
		want string
	}{
		{"zero", time.Time{}, "unknown"},
		{"just now", now.Add(-10 * time.Second), "just now"},
		{"minutes", now.Add(-5 * time.Minute), "5m ago"},
		{"hours", now.Add(-2 * time.Hour), "2h ago"},
		{"yesterday", now.Add(-30 * time.Hour), "yesterday"},
		{"days", now.Add(-72 * time.Hour), "3 days ago"},
		{"future", now.Add(10 * time.Minute), "just now"},
	}
	for _, c := range cases {
		if got := relativeTime(c.t); got != c.want {
			t.Errorf("relativeTime(%s) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestRowTagPrecedence(t *testing.T) {
	// Autosave wins even if modified/elite flags are also set.
	auto := game.SaveInfo{Name: game.AutosaveName, Modified: true, Elite: true}
	if got := rowTag(auto); got != "[gold]★ auto[-]" {
		t.Errorf("rowTag(autosave) = %q, want auto tag", got)
	}
	// Modified outranks elite for a player-named save.
	mod := game.SaveInfo{Name: "hero", Modified: true, Elite: true}
	if got := rowTag(mod); got != "[red]⚠ modified[-]" {
		t.Errorf("rowTag(modified) = %q, want modified tag", got)
	}
	// Elite when neither auto nor modified.
	elite := game.SaveInfo{Name: "hero", Elite: true}
	if got := rowTag(elite); got != "[gold]⭐ elite[-]" {
		t.Errorf("rowTag(elite) = %q, want elite tag", got)
	}
	// Plain save → no tag.
	plain := game.SaveInfo{Name: "hero"}
	if got := rowTag(plain); got != "" {
		t.Errorf("rowTag(plain) = %q, want empty", got)
	}
}

func TestDetailTextCorrupt(t *testing.T) {
	s := game.SaveInfo{Name: "broken", Corrupt: true, Timestamp: time.Now()}
	got := detailText(s)
	if want := "Corrupt save"; !contains(got, want) {
		t.Errorf("detailText(corrupt) = %q, should mention %q", got, want)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
