package main

import (
	"testing"

	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// BenchmarkBuild is the per-snapshot cost (runs only when the state changes).
func BenchmarkBuild(b *testing.B) {
	mv, err := loadView("states", "galactic_age", 7)
	if err != nil {
		b.Skip(err)
	}
	w := NewWorld(7, worldW, worldH)
	p := NewPlan(w)
	for i := 0; i < b.N; i++ {
		Build(w, p, mv, false)
	}
}

// BenchmarkFrame is the per-refresh cost of one 160x48 frame.
func BenchmarkFrame(b *testing.B) {
	mv, err := loadView("states", "galactic_age", 7)
	if err != nil {
		b.Skip(err)
	}
	_ = theme.SetActive("forge")
	w := NewWorld(7, worldW, worldH)
	v := NewView(Build(w, NewPlan(w), mv, false))
	scr := tcell.NewSimulationScreen("UTF-8")
	_ = scr.Init()
	scr.SetSize(160, 48)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Frame = i
		v.Draw(scr, 160, 48)
	}
}

// BenchmarkWorld is the one-off cost of generating terrain and the plan.
func BenchmarkWorld(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewPlan(NewWorld(7, worldW, worldH))
	}
}
