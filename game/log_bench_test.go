package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// The log and the refusals are hot paths: every line and every refusal is
// built from a format. These benchmarks hold the cost of that, so a change to
// how the lines are written can be measured against the numbers before it.

// BenchmarkLogLine measures a command that writes two log lines (one
// debug, one routine), each built from a format.
func BenchmarkLogLine(b *testing.B) {
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
	ge := NewGameEngine()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ge.RecruitWorker("worker", config.Million); err == nil {
			b.Fatal("a recruit of a million was not refused")
		}
	}
}
