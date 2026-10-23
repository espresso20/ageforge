package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// Wonder overflow is gone with the storage caps: `wonder overflow` is no
// command, the usage line does not offer it, and the Wonders panel has no
// switch.
func TestWonderOverflowCommandIsGone(t *testing.T) {
	engine := game.NewGameEngine()
	res := HandleCommand("wonder overflow", engine)
	if res.Type != "error" || !strings.Contains(res.Message, "Unknown wonder command") || strings.Contains(res.Message, "overflow [") {
		t.Errorf("wonder overflow = %+v, want the unknown-command error without an overflow usage", res)
	}
	if strings.Contains(wondersProvider(engine.GetState(), 100), "Overflow") {
		t.Error("the Wonders panel still shows an overflow switch")
	}
	for _, c := range NewAutoCompleter(engine)("wonder overflow") {
		t.Errorf("autocomplete still offers %q", c)
	}
}
