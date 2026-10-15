package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

func prestigeCueTestState() game.GameState {
	return game.GameState{
		Age:            "bronze_age",
		TickIntervalMs: 2000,
		PrestigeCue: game.PrestigeCue{
			Ready: true, Age: "bronze_age", FinishTicks: 30, FreshTicks: 12,
			Multiplier: 1.10, NextMultiplier: 1.25,
		},
	}
}

func TestPrestigeCueNoticeTextFitsCommonWidths(t *testing.T) {
	state := prestigeCueTestState()
	for _, width := range []int{80, 100} {
		text := prestigeCueText(state, width)
		for _, want := range []string{
			"Time to prestige?",
			"Finishing the Bronze Age takes about ~1m at today's income.",
			"A new run with x1.25 production would be back here and past it in about ~24s.",
			"Type prestige to see what you keep.",
		} {
			if !strings.Contains(strings.Join(strings.Fields(text), " "), want) {
				t.Errorf("at width %d notice missing %q:\n%s", width, want, text)
			}
		}
		max := width*17/20 - 4
		for i, line := range strings.Split(text, "\n") {
			if len(line) > max {
				t.Errorf("at width %d line %d is %d cells, max %d: %q", width, i+1, len(line), max, line)
			}
		}
	}
}

func TestPrestigeCueNoticeVisibility(t *testing.T) {
	state := prestigeCueTestState()
	if !shouldShowPrestigeCue(state, false) {
		t.Fatal("ready unseen cue should show")
	}
	state.PrestigeCue.Seen = true
	if shouldShowPrestigeCue(state, false) {
		t.Fatal("seen cue should not show")
	}
	state.PrestigeCue.Seen = false
	state.PrestigeCue.Ready = false
	if shouldShowPrestigeCue(state, false) {
		t.Fatal("not-ready cue should not show")
	}
	state.PrestigeCue.Ready = true
	if shouldShowPrestigeCue(state, true) {
		t.Fatal("cue should wait behind an active page")
	}
}

func TestStatsPrestigeCueLineOnlyWhenReady(t *testing.T) {
	state := prestigeCueTestState()
	text := statsProvider(state, 0)
	want := "A new run would pass this age in about ~24s; finishing it takes about ~1m at today's income."
	if !strings.Contains(text, want) {
		t.Fatalf("stats missing cue line %q:\n%s", want, text)
	}
	state.PrestigeCue.Ready = false
	if got := statsProvider(state, 0); strings.Contains(got, "A new run would pass this age") {
		t.Fatalf("stats showed cue while not ready:\n%s", got)
	}
}
