package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/vibe"
)

// The log and the refusals are the paths VibeMode reaches into: every line
// and every refusal is built from a format. These benchmarks hold the cost of
// that with VibeMode off (the shipped default), so a change to how the lines
// are written can be measured against the numbers before it.

// BenchmarkLogLine measures a command that writes two log lines (one
// debug, one routine), each built from a format.
func BenchmarkLogLine(b *testing.B) {
	vibeOffB(b)
	ge := NewGameEngine()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ge.GatherResource("food", 1); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRefusal measures a command that is refused with a formatted
// message.
func BenchmarkRefusal(b *testing.B) {
	vibeOffB(b)
	ge := NewGameEngine()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ge.BuildMultiple("hut", 0); err == nil {
			b.Fatal("a build of none was not refused")
		}
	}
}

// BenchmarkRefusalSeveralValues measures a refusal that formats a count
// with its word and two numbers into its message.
func BenchmarkRefusalSeveralValues(b *testing.B) {
	vibeOffB(b)
	ge := NewGameEngine()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ge.RecruitWorker("worker", config.Million); err == nil {
			b.Fatal("a recruit of a million was not refused")
		}
	}
}

// vibeOffB puts the VibeMode switch off for a benchmark, and back.
func vibeOffB(tb testing.TB) {
	tb.Helper()
	was := vibe.On()
	tb.Cleanup(func() { vibe.Set(was) })
	vibe.Set(false)
}
