package mapview

import (
	"math"
	"testing"
	"time"

	"github.com/espresso20/ageforge/world"
)

func transcendent(t *testing.T) *View {
	t.Helper()
	v := New()
	v.Set(build(t, world.AgeTranscendent), 1, testPal)
	v.bod.wg.Wait()
	if v.open != 0 || v.ch[0].f.Kind != world.FrameWeb {
		t.Fatalf("open %d, first frame %v: the web should be open", v.open, v.ch[0].f.Kind)
	}
	return v
}

func TestWebIsLockedAtQuantum(t *testing.T) {
	v := New()
	v.Set(build(t, world.AgeQuantum), 1, testPal)
	if v.ch[0].f.Kind != world.FrameWeb || v.open != 1 {
		t.Fatalf("first frame %v, open %d: the web should sit locked above the Local Group", v.ch[0].f.Kind, v.open)
	}
	if world.OpensAt(world.FrameWeb) != world.AgeTranscendent {
		t.Errorf("the web opens at %q", world.OpensAt(world.FrameWeb))
	}
	o := v.objAt(0, "local_group")
	z, ok := zFor(v, 0)
	if o == nil || !ok {
		t.Fatal("no web node or zoom")
	}
	at(v, z, 0, o.Pos)
	x, y := screenOf(v, 0, o)
	v.Click(int(math.Round(x)), int(math.Round(y)), tw, th)
	if s := v.state(); s.Lock != world.AgeTranscendent {
		t.Errorf("a locked web object says %q opens it", s.Lock)
	}
	v.Focus(int(math.Round(x)), int(math.Round(y)), tw, th)
	if v.ch[v.m].f.Kind != world.FrameWeb {
		t.Error("a locked frame was entered")
	}
}

func TestWebIsUnreachableBelowQuantum(t *testing.T) {
	v := New()
	v.Set(build(t, world.AgeGalactic), 1, testPal)
	for _, l := range v.ch {
		if l.f.Kind == world.FrameWeb {
			t.Fatal("the web is in the chain at Galactic")
		}
	}
	v.Do(Cmd{Op: ZoomOut})
	for range 400 {
		v.Do(Cmd{Op: ZoomOut})
		v.Update(100 * time.Millisecond)
	}
	if v.ch[v.m].f.Kind == world.FrameWeb {
		t.Error("zoomed out into the web")
	}
}

func TestHomeAndZoomOutReachTheWeb(t *testing.T) {
	v := transcendent(t)
	v.Do(Cmd{Op: Home})
	if v.ch[v.m].f.Kind != world.FrameWeb {
		t.Errorf("home shows %v", v.ch[v.m].f.Kind)
	}
	lg := v.ch.idx(world.FrameLocalGroup)
	z, ok := zFor(v, lg)
	if !ok {
		t.Fatal("no zoom inside the Local Group")
	}
	at(v, z, lg, world.Vec{})
	for range 600 {
		v.Do(Cmd{Op: ZoomOut})
		v.Update(100 * time.Millisecond)
	}
	if v.ch[v.m].f.Kind != world.FrameWeb {
		t.Errorf("zooming out from the Local Group ends in %v", v.ch[v.m].f.Kind)
	}
}

func TestLocalGroupNodeDivesAndOthersDoNot(t *testing.T) {
	v := transcendent(t)
	z, ok := zFor(v, 0)
	if !ok {
		t.Fatal("no zoom inside the web")
	}
	node := v.objAt(0, "local_group")
	at(v, z, 0, node.Pos)
	x, y := screenOf(v, 0, node)
	v.Focus(int(math.Round(x)), int(math.Round(y)), tw, th)
	if lo := v.ch[1].lo; v.tz > lo+1e-9 {
		t.Errorf("target zoom %.2f is above the Local Group's start %.2f: no dive", v.tz, lo)
	}
	v = transcendent(t)
	var far *world.Object
	for i := range v.ch[0].f.Objects {
		if o := &v.ch[0].f.Objects[i]; o.Kind == world.KindCluster && o.Key != "local_group" {
			far = o
			break
		}
	}
	at(v, z, 0, far.Pos)
	x, y = screenOf(v, 0, far)
	v.Focus(int(math.Round(x)), int(math.Round(y)), tw, th)
	if v.tz != z {
		t.Errorf("a far cluster moved the zoom %.2f to %.2f: it should only be selected", z, v.tz)
	}
	if v.sel.key != far.Key {
		t.Errorf("selected %q, want %q", v.sel.key, far.Key)
	}
}

func TestClickingAClusterSettlesIt(t *testing.T) {
	v := transcendent(t)
	z, _ := zFor(v, 0)
	var far *world.Object
	for i := range v.ch[0].f.Objects {
		if o := &v.ch[0].f.Objects[i]; o.Kind == world.KindCluster && o.Key != "local_group" {
			far = o
			break
		}
	}
	if far.Doubt == 0 || v.isSettled(far) {
		t.Fatal("a cluster should start uncertain")
	}
	at(v, z, 0, far.Pos)
	x, y := screenOf(v, 0, far)
	v.Click(int(math.Round(x)), int(math.Round(y)), tw, th)
	if got := v.TakeSettled(); len(got) != 1 || got[0] != far.Key {
		t.Fatalf("a click settled %v, want [%s]", got, far.Key)
	}
}

func TestClickingAStrandSelectsItsLine(t *testing.T) {
	v := transcendent(t)
	z, _ := zFor(v, 0)
	var s *world.Object
	for i := range v.ch[0].f.Objects {
		if o := &v.ch[0].f.Objects[i]; o.Kind == world.KindStrand {
			s = o
			break
		}
	}
	at(v, z, 0, s.Pos)
	x, y := screenOf(v, 0, s)
	v.Click(int(math.Round(x)), int(math.Round(y)), tw, th)
	if got := v.objAt(v.sel.fi, v.sel.key); got == nil || got.Kind != world.KindStrand {
		t.Errorf("a click on a strand selected %v", got)
	}
}

func TestWebFrameIsQuick(t *testing.T) {
	v := transcendent(t)
	v.Do(Cmd{Op: Home})
	v.Render(1200, 800)
	start := time.Now()
	const n = 20
	for range n {
		v.dirty = true
		v.Render(1200, 800)
	}
	per := time.Since(start) / n
	t.Logf("web frame: %v", per)
	if per > 4*time.Millisecond {
		t.Errorf("a web frame takes %v at 1200 px wide", per)
	}
}
