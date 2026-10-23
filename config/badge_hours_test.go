package config

import (
	"fmt"
	"strings"
	"testing"
)

// TestEverythingFullSaysTheHoursItAsks: the badge's text names the hours the
// census holds it to (MaximalistHours), and the badge is still in the book.
func TestEverythingFullSaysTheHoursItAsks(t *testing.T) {
	for _, b := range Badges() {
		if b.Key != "special.everything_full" {
			continue
		}
		if want := fmt.Sprintf("%d hours of your income", MaximalistHours); !strings.Contains(b.Desc, want) {
			t.Errorf("Everything Full says %q, want it to ask %q", b.Desc, want)
		}
		return
	}
	t.Fatal("no special.everything_full badge")
}
