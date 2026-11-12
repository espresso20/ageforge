package mapview

import (
	"bytes"
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

// TestFadeMaskedPaintMatchesFull: painting only the pixels each layer keeps
// gives the same picture as painting both layers whole and dithering between
// them, on every frame of both bands. (A golden hash would differ by CPU:
// arm64 fuses multiply-adds.)
func TestFadeMaskedPaintMatchesFull(t *testing.T) {
	v := fadeView(t)
	w, h := 1200/ArtScale(1200), 675/ArtScale(1200)
	n := 0
	for i, c := range v.ch {
		if (c.f.Kind != world.FrameGalaxy && c.f.Kind != world.FrameRegion) || i == 0 {
			continue
		}
		for k := 0; k <= 40; k++ {
			z := c.hi - (c.hi-c.lo)*float64(k)/40
			v.z, v.tz = z, z
			v.m = v.ch.masterAt(z, look{})
			v.centre()
			v.Update(0)
			under, over, f := v.ch.layersAt(z, v.look())
			if under == over {
				continue
			}
			v.dirty = true
			got := append([]byte(nil), v.Render(w, h).Pix...)
			snapped := v.snapped(w)
			want := image.NewRGBA(image.Rect(0, 0, w, h))
			top := image.NewRGBA(want.Rect)
			v.paint(want, under, snapped, nil)
			v.paint(top, over, snapped, nil)
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					if Bayer(x, y) < f {
						o := want.PixOffset(x, y)
						copy(want.Pix[o:o+4], top.Pix[o:o+4])
					}
				}
			}
			v.overlay(want, snapped)
			if !bytes.Equal(got, want.Pix) {
				t.Fatalf("frame %d of %v at z=%.3f differs from the full paint", k, c.f.Kind, z)
			}
			n++
		}
	}
	if n == 0 {
		t.Fatal("no crossfade frames found")
	}
}

func fadeView(tb testing.TB) *View {
	v := New()
	v.Set(build(tb, world.AgeTranscendent), 1, testPal)
	v.bod.wg.Wait()
	return v
}
