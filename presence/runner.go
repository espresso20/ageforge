package presence

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Discord allows one activity update every 15 seconds.
const (
	defaultInterval = 15 * time.Second
	defaultRetry    = 30 * time.Second
	ioTimeout       = time.Second
	readyTimeout    = 2 * time.Second
	closeWait       = 2 * time.Second
)

// Options sets up a Runner.
type Options struct {
	// AppID is the Discord application id. Empty turns the Runner off: it never dials.
	AppID string
	// Activity returns what to show now, or nil to show nothing. It runs on the Runner's
	// own goroutine, so it may read the engine's state.
	Activity func() *Activity
	// Dial opens the connection to Discord. Nil dials the local Discord client.
	Dial func() (io.ReadWriteCloser, error)
	// Interval and Retry are how often the activity is checked and how often a lost
	// connection is retried. Zero takes Discord's 15s and 30s.
	Interval, Retry time.Duration
}

// Runner keeps the profile current until Close.
type Runner struct {
	opts    Options
	pid     int
	done    chan struct{}
	stopped chan struct{}
	once    sync.Once
	poke    chan struct{}
	seq     int
}

// Start begins showing the activity. It returns at once; with no AppID (or no Activity
// function) it returns a Runner that does nothing.
func Start(o Options) *Runner {
	r := &Runner{opts: o, pid: os.Getpid(), done: make(chan struct{}), poke: make(chan struct{}, 1), stopped: make(chan struct{})}
	if o.AppID == "" || o.Activity == nil {
		close(r.stopped)
		return r
	}
	if o.Dial == nil {
		r.opts.Dial = dialIPC
	}
	if o.Interval <= 0 {
		r.opts.Interval = defaultInterval
	}
	if o.Retry <= 0 {
		r.opts.Retry = defaultRetry
	}
	go r.loop()
	return r
}

// Close clears the activity and closes the connection. It is safe to call more than once,
// and on a nil Runner.
func (r *Runner) Close() {
	if r == nil {
		return
	}
	r.once.Do(func() { close(r.done) })
	select {
	case <-r.stopped:
	case <-time.After(closeWait):
	}
}

// Refresh asks for the activity to be checked now rather than at the next tick, so a
// change of setting shows at once. It never blocks and is safe on a nil Runner.
func (r *Runner) Refresh() {
	if r == nil {
		return
	}
	select {
	case r.poke <- struct{}{}:
	default:
	}
}

func (r *Runner) loop() {
	defer close(r.stopped)
	var (
		s       *session
		sent    bool
		last    string
		lastTry time.Time
	)
	drop := func() {
		if s != nil {
			s.close()
		}
		s, sent = nil, false
	}
	tick := func() {
		act := r.opts.Activity()
		if s != nil && s.isDead() {
			drop()
		}
		if s == nil {
			if act == nil || (!lastTry.IsZero() && time.Since(lastTry) < r.opts.Retry) {
				return
			}
			lastTry = time.Now()
			ns, err := r.connect()
			if err != nil {
				return
			}
			s = ns
		}
		key := activityKey(act)
		if sent && key == last {
			return
		}
		if err := r.send(s, act); err != nil {
			drop()
			return
		}
		sent, last = true, key
	}

	tick()
	t := time.NewTicker(r.opts.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			tick()
		case <-r.poke:
			tick()
		case <-r.done:
			if s != nil {
				if !s.isDead() && sent && last != activityKey(nil) {
					_ = r.send(s, nil)
				}
				s.close()
			}
			return
		}
	}
}

func activityKey(a *Activity) string {
	if a == nil {
		return ""
	}
	b, _ := json.Marshal(a.wire())
	return string(b)
}

// session is one connection to Discord.
type session struct {
	rw    io.ReadWriteCloser
	wmu   sync.Mutex
	ready chan struct{}
	dead  chan struct{}
	once  sync.Once
	rdy   sync.Once
}

func (s *session) isDead() bool {
	select {
	case <-s.dead:
		return true
	default:
		return false
	}
}

func (s *session) close() {
	s.once.Do(func() { close(s.dead) })
	_ = s.rw.Close()
}

// write sends bytes with a short deadline where the connection has one.
func (s *session) write(b []byte) error {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	if d, ok := s.rw.(interface{ SetWriteDeadline(time.Time) error }); ok {
		_ = d.SetWriteDeadline(time.Now().Add(ioTimeout))
	}
	_, err := s.rw.Write(b)
	return err
}

// read answers the client until the connection ends: replies are read and dropped, a
// ping is answered, a close ends the session.
func (s *session) read() {
	defer s.close()
	for {
		f, err := readFrame(s.rw)
		if err != nil {
			return
		}
		switch f.Op {
		case opClose:
			return
		case opPing:
			if s.write(encodeFrame(opPong, f.Data)) != nil {
				return
			}
		default:
			s.rdy.Do(func() { close(s.ready) })
		}
	}
}

// connect dials and shakes hands, and waits for Discord's ready reply.
func (r *Runner) connect() (*session, error) {
	rw, err := r.opts.Dial()
	if err != nil {
		return nil, err
	}
	s := &session{rw: rw, ready: make(chan struct{}), dead: make(chan struct{})}
	go s.read()
	if err := s.write(handshakeFrame(r.opts.AppID)); err != nil {
		s.close()
		return nil, err
	}
	select {
	case <-s.ready:
		return s, nil
	case <-s.dead:
		return nil, fmt.Errorf("presence: Discord closed the connection")
	case <-time.After(readyTimeout):
		s.close()
		return nil, fmt.Errorf("presence: Discord did not answer the handshake")
	}
}

func (r *Runner) send(s *session, a *Activity) error {
	r.seq++
	nonce := fmt.Sprintf("%d-%d", time.Now().UnixNano(), r.seq)
	return s.write(activityFrame(r.pid, a, nonce))
}
