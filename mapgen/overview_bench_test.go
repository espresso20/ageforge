package mapgen

import "testing"

// BenchmarkOverviewSystem times the three squeezed views of the Sol system at
// a wide terminal's size.
func BenchmarkOverviewSystem(b *testing.B) {
	var gen Generator
	sc := gen.Draw(model(state("interstellar_age", 42, noon)), 0)
	for _, c := range []struct {
		name string
		mode SystemViewMode
	}{{"cosmos", SystemViewCosmos}, {"overview", SystemViewFull}, {"inner", SystemViewInner}} {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				sc.Anim = i
				_ = sc.overviewSystem(200, 55, c.mode, 0)
			}
		})
	}
}
