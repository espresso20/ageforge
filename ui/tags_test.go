package ui

import (
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

// visible is s as the player sees it: the tags the game means read, and
// gone.
func visible(s string) string { return kit.Strip(s) }

func TestSafeTagsKeepsBracketedWords(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Usage: sell <building> [count]", "Usage: sell <building> [count]"},
		{"Usage: account import <path> [replace]", "Usage: account import <path> [replace]"},
		{"   Status:  [green][allied][-]  +20% iron", "   Status:  [allied]  +20% iron"},
		{"[gold]Progress exported:[-] /tmp/x.json", "Progress exported: /tmp/x.json"},
		{"[accent:chip:b] Enter [-:-:-] ok", " Enter  ok"},
		{"[#ff8800]hex[-] and [X] and [i]", "hex and [X] and [i]"},
		{"[amt|all] and [3] stay", "[amt|all] and [3] stay"},
		{"already escaped [count[]", "already escaped [count]"},
		{"[gray](current)[-]", "(current)"},
	}
	for _, c := range cases {
		got := visible(safeTags(c.in))
		if got != c.want {
			t.Errorf("safeTags(%q) renders %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSafeTagsIdempotent(t *testing.T) {
	in := "Usage: plan build <building> [count] [gold]x[-]"
	once := safeTags(in)
	if twice := safeTags(once); twice != once {
		t.Errorf("safeTags not idempotent: %q then %q", once, twice)
	}
}

func TestLitEscapesEverything(t *testing.T) {
	if got := visible(lit("[gold]not a color[-]")); got != "[gold]not a color[-]" {
		t.Errorf("lit rendered %q", got)
	}
}
