//go:build !windows

package presence

import (
	"encoding/json"
	"io"
	"net"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeDiscord is a unix socket that speaks enough of the protocol: it answers the
// handshake and records every frame it is sent.
type fakeDiscord struct {
	ln     net.Listener
	mu     sync.Mutex
	frames []frame
	dials  atomic.Int32
}

func newFakeDiscord(t *testing.T) *fakeDiscord {
	t.Helper()
	ln, err := net.Listen("unix", filepath.Join(t.TempDir(), "discord-ipc-0"))
	if err != nil {
		t.Skipf("no unix sockets here: %v", err)
	}
	f := &fakeDiscord{ln: ln}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go f.serve(c)
		}
	}()
	return f
}

func (f *fakeDiscord) serve(c net.Conn) {
	defer c.Close()
	for {
		fr, err := readFrame(c)
		if err != nil {
			return
		}
		f.mu.Lock()
		f.frames = append(f.frames, fr)
		f.mu.Unlock()
		if fr.Op == opHandshake {
			c.Write(encodeFrame(opFrame, []byte(`{"cmd":"DISPATCH","evt":"READY"}`)))
		} else {
			c.Write(encodeFrame(opFrame, []byte(`{"cmd":"SET_ACTIVITY"}`)))
		}
	}
}

func (f *fakeDiscord) dial() (io.ReadWriteCloser, error) {
	f.dials.Add(1)
	return net.Dial("unix", f.ln.Addr().String())
}

func (f *fakeDiscord) snapshot() []frame {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]frame(nil), f.frames...)
}

// activities lists the SET_ACTIVITY frames as the details they carry ("" for a clear).
func (f *fakeDiscord) activities() []string {
	var out []string
	for _, fr := range f.snapshot() {
		if fr.Op != opFrame {
			continue
		}
		var m struct {
			Args struct {
				Activity *struct {
					Details string `json:"details"`
				} `json:"activity"`
			} `json:"args"`
		}
		if json.Unmarshal(fr.Data, &m) != nil {
			continue
		}
		if m.Args.Activity == nil {
			out = append(out, "")
		} else {
			out = append(out, m.Args.Activity.Details)
		}
	}
	return out
}

func waitFor(t *testing.T, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestRunnerAgainstFakeServer(t *testing.T) {
	srv := newFakeDiscord(t)
	var mu sync.Mutex
	details := "one"
	r := Start(Options{
		AppID: "555",
		Activity: func() *Activity {
			mu.Lock()
			defer mu.Unlock()
			return &Activity{Details: details}
		},
		Dial:     srv.dial,
		Interval: 10 * time.Millisecond,
		Retry:    10 * time.Millisecond,
	})
	waitFor(t, "first activity", func() bool { return equal(srv.activities(), []string{"one"}) })

	// The handshake came first and carried the id.
	first := srv.snapshot()[0]
	if first.Op != opHandshake || string(first.Data) != `{"client_id":"555","v":1}` {
		t.Errorf("first frame = %d %s", first.Op, first.Data)
	}

	// Several ticks with nothing changed send nothing more.
	time.Sleep(100 * time.Millisecond)
	if got := srv.activities(); !equal(got, []string{"one"}) {
		t.Errorf("an unchanged activity was resent: %v", got)
	}

	mu.Lock()
	details = "two"
	mu.Unlock()
	waitFor(t, "changed activity", func() bool { return equal(srv.activities(), []string{"one", "two"}) })

	r.Close()
	waitFor(t, "clear on close", func() bool { return equal(srv.activities(), []string{"one", "two", ""}) })
	r.Close() // twice is fine
	if n := srv.dials.Load(); n != 1 {
		t.Errorf("dialled %d times, want 1", n)
	}
}

func TestRunnerNilActivityClearsAndRefresh(t *testing.T) {
	srv := newFakeDiscord(t)
	var on atomic.Bool
	on.Store(true)
	r := Start(Options{
		AppID: "555",
		Activity: func() *Activity {
			if on.Load() {
				return &Activity{Details: "x"}
			}
			return nil
		},
		Dial:     srv.dial,
		Interval: time.Hour, // only Refresh drives it after the first tick
		Retry:    time.Hour,
	})
	defer r.Close()
	waitFor(t, "shown", func() bool { return equal(srv.activities(), []string{"x"}) })
	on.Store(false)
	r.Refresh()
	waitFor(t, "cleared", func() bool { return equal(srv.activities(), []string{"x", ""}) })
}

func TestRunnerWithoutDiscordIsSilentAndRetries(t *testing.T) {
	var dials atomic.Int32
	r := Start(Options{
		AppID:    "555",
		Activity: func() *Activity { return &Activity{Details: "x"} },
		Dial: func() (io.ReadWriteCloser, error) {
			dials.Add(1)
			return nil, io.ErrClosedPipe
		},
		Interval: 5 * time.Millisecond,
		Retry:    40 * time.Millisecond,
	})
	time.Sleep(150 * time.Millisecond)
	r.Close()
	if n := dials.Load(); n < 2 || n > 5 {
		t.Errorf("dialled %d times in 150ms with a 40ms retry", n)
	}
}

func TestEmptyAppIDNeverDials(t *testing.T) {
	var dials atomic.Int32
	r := Start(Options{
		AppID:    "",
		Activity: func() *Activity { return &Activity{Details: "x"} },
		Dial: func() (io.ReadWriteCloser, error) {
			dials.Add(1)
			return nil, io.ErrClosedPipe
		},
		Interval: time.Millisecond,
	})
	time.Sleep(30 * time.Millisecond)
	r.Refresh()
	r.Close()
	if dials.Load() != 0 {
		t.Error("presence dialled with no application id")
	}
}

func TestRunnerOffSettingNeverDials(t *testing.T) {
	var dials atomic.Int32
	r := Start(Options{
		AppID:    "555",
		Activity: func() *Activity { return nil },
		Dial: func() (io.ReadWriteCloser, error) {
			dials.Add(1)
			return nil, io.ErrClosedPipe
		},
		Interval: time.Millisecond,
	})
	time.Sleep(30 * time.Millisecond)
	r.Close()
	if dials.Load() != 0 {
		t.Error("presence dialled with nothing to show")
	}
}
