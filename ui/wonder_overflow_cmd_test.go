package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

func TestWonderOverflowCommand(t *testing.T) {
	engine := game.NewGameEngine()
	if !engine.WonderOverflow() {
		t.Fatal("overflow should start on")
	}
	if res := HandleCommand("wonder overflow off", engine); res.Type == "error" || engine.WonderOverflow() {
		t.Fatalf("wonder overflow off: %+v, overflow still %v", res, engine.WonderOverflow())
	}
	if res := HandleCommand("wonder overflow", engine); !strings.Contains(res.Message, "off") {
		t.Errorf("bare wonder overflow = %q, want it to say off", res.Message)
	}
	if res := HandleCommand("wonder overflow ON", engine); res.Type == "error" || !engine.WonderOverflow() {
		t.Fatalf("wonder overflow ON: %+v", res)
	}
	if res := HandleCommand("wonder overflow maybe", engine); res.Type != "error" {
		t.Errorf("wonder overflow maybe = %+v, want a usage error", res)
	}
	// The Wonders panel shows the switch.
	st := engine.GetState()
	if !strings.Contains(wondersProvider(st, 100), "Overflow") {
		t.Error("the Wonders panel doesn't show the overflow switch")
	}
	got := NewAutoCompleter(engine)("wonder overflow o")
	if strings.Join(got, ",") != "wonder overflow off,wonder overflow on" {
		t.Errorf("autocomplete = %v", got)
	}
}
