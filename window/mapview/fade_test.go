package mapview

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"image"
	"sort"
	"testing"
	"time"

	"github.com/espresso20/ageforge/world"
)

// fadeFrames times Render across the zoom band where frame kind over fades in
// over the one above it. It returns the sorted frame times.
func fadeFrames(v *View, over world.FrameKind, hash hash.Hash) []time.Duration {
	w, h := 1200/ArtScale(1200), 675/ArtScale(1200)
	var ts []time.Duration
	for i, c := range v.ch {
		if c.f.Kind != over || i == 0 {
			continue
		}
		for k := 0; k <= 200; k++ {
			z := c.hi - (c.hi-c.lo)*float64(k)/200
			v.z, v.tz = z, z
			v.m = v.ch.masterAt(z, look{})
			v.centre()
			v.Update(0)
			if u, o, _ := v.ch.layersAt(z, v.look()); u == o {
				continue
			}
			best := time.Hour // the best of five: a busy machine adds time, never takes it away
			var img *image.RGBA
			for r := 0; r < 5; r++ {
				v.dirty = true
				t0 := time.Now()
				img = v.Render(w, h)
				best = min(best, time.Since(t0))
			}
			ts = append(ts, best)
			if hash != nil {
				hash.Write(img.Pix)
			}
		}
	}
	sort.Slice(ts, func(a, b int) bool { return ts[a] < ts[b] })
	return ts
}

// BenchmarkFade reports the median and worst crossfade frame (the best of five
// renders each) per band at 1200 px wide. The budget is 4 ms a frame.
func BenchmarkFade(b *testing.B) {
	v := fadeView(b)
	for name, k := range map[string]world.FrameKind{"galaxy_over_local_group": world.FrameGalaxy, "region_over_galaxy": world.FrameRegion} {
		b.Run(name, func(b *testing.B) {
			var ts []time.Duration
			for i := 0; i < b.N; i++ {
				ts = fadeFrames(v, k, nil)
			}
			b.ReportMetric(float64(ts[len(ts)/2])/1e6, "median_ms")
			b.ReportMetric(float64(ts[len(ts)-1])/1e6, "worst_ms")
		})
	}
}

// TestFadePictureIsPinned: the crossfade pixels do not drift (a golden hash of
// every frame of both bands).
func TestFadePictureIsPinned(t *testing.T) {
	h := sha256.New()
	v := fadeView(t)
	fadeFrames(v, world.FrameGalaxy, h)
	fadeFrames(v, world.FrameRegion, h)
	if got := hex.EncodeToString(h.Sum(nil)); got != fadeGolden {
		t.Errorf("fade picture changed: %s", got)
	}
}

const fadeGolden = "3a14fa4961f8b81b2eacbd8e44c1f2fee0da414b74dcd5e6c67a18173f687c4d"

func fadeView(tb testing.TB) *View {
	v := New()
	v.Set(build(tb, world.AgeTranscendent), 1, testPal)
	v.bod.wg.Wait()
	return v
}
