package game

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"time"

	"github.com/espresso20/ageforge/rowid"
)

// logentry.go: what a log line holds, how repeats fold into one line, and how
// the log is saved.

// Office is the part of the town a log line is about. Zero is no office: every
// line names one.
type Office uint8

const (
	OfficeBuilding Office = iota + 1
	OfficeResearch
	OfficePlan
	OfficePeople
	OfficeStores
	OfficeTrade
	OfficeDiplomacy
	OfficeArmy
	OfficeWorld
	OfficeGame
	officeCount = int(OfficeGame)
)

// Offices lists the ten offices in display order.
func Offices() []Office {
	out := make([]Office, officeCount)
	for i := range out {
		out[i] = Office(i + 1)
	}
	return out
}

var officeNames = [...]rowid.Entry{
	rowid.E("game.logentry.officeNames.01", "Building"), rowid.E("game.logentry.officeNames.02", "Research"),
	rowid.E("game.logentry.officeNames.03", "Plan"), rowid.E("game.logentry.officeNames.04", "People"),
	rowid.E("game.logentry.officeNames.05", "Stores"), rowid.E("game.logentry.officeNames.06", "Trade"),
	rowid.E("game.logentry.officeNames.07", "Diplomacy"), rowid.E("game.logentry.officeNames.08", "Army"),
	rowid.E("game.logentry.officeNames.09", "World"), rowid.E("game.logentry.officeNames.10", "Game"),
}

// Name is the office's name.
func (o Office) Name() string {
	if o < 1 || int(o) > officeCount {
		return ""
	}
	return officeNames[o-1].String()
}

// The three levels a line is shown at: each level shows its own lines and the
// ones above it.
const (
	LevelHeadline   = 1
	LevelTown       = 2
	LevelEverything = 3
)

const (
	// MaxLogSize is how many folded lines the log keeps and saves.
	MaxLogSize = 2000
	// foldTicks is how recent a line's newest member must be for a repeat to
	// join it.
	foldTicks = 70
	// foldMembers is how many members of a fold are kept whole.
	foldMembers = 12
)

// LogAmt is an amount of a resource a line paid or gained. Workers are the
// resource "workers".
type LogAmt struct {
	Res string  `json:"r"`
	N   float64 `json:"n"`
}

// LogMember is one line of a fold, kept whole.
type LogMember struct {
	At     int64    `json:"a"`
	Tick   int      `json:"k"`
	Text   string   `json:"m"`
	Paid   []LogAmt `json:"p,omitempty"`
	Gained []LogAmt `json:"g,omitempty"`
}

// LogEntry is one line of the log: when, what, what it cost and what it gave.
type LogEntry struct {
	Tick    int    `json:"k"`
	Message string `json:"m"`
	// Type is the category: "info", "success", "warning", "error", "event"
	// or LogRoutine. It decides the color.
	Type    string `json:"t"`
	At      int64  `json:"a"` // wall clock, unix seconds
	Office  Office `json:"o"`
	Level   int    `json:"l"`
	Trouble bool   `json:"tr,omitempty"`
	// Subject is the config key the line is about; ID is the row id of the
	// line's format, the key repeats fold by.
	Subject string   `json:"s,omitempty"`
	ID      string   `json:"i,omitempty"`
	Paid    []LogAmt `json:"p,omitempty"`
	Gained  []LogAmt `json:"g,omitempty"`
	// Count is how many lines this one stands for (0 or 1: itself), FirstAt
	// the time of the first of them and Members the newest of them.
	Count   int         `json:"c,omitempty"`
	FirstAt int64       `json:"f,omitempty"`
	Members []LogMember `json:"mm,omitempty"`
	// Seam marks the line a load writes to close the sitting before it: At is
	// the sitting's last line, FirstAt its first and Count its line count.
	Seam bool `json:"seam,omitempty"`
}

// Earlier is how many of the fold's lines are older than the kept members.
func (e LogEntry) Earlier() int { return max(0, e.Count-len(e.Members)) }

// logOpt sets a field of an entry as it is written.
type logOpt func(*logWrite)

type logWrite struct {
	e     LogEntry
	total func(gained float64) string
}

func logHead(w *logWrite)    { w.e.Level = LevelHeadline }
func logTrouble(w *logWrite) { w.e.Trouble = true }
func logID(id string) logOpt { return func(w *logWrite) { w.e.ID = id } }
func logSubj(s string) logOpt {
	return func(w *logWrite) { w.e.Subject = s }
}
func logPaid(a ...LogAmt) logOpt { return func(w *logWrite) { w.e.Paid = append(w.e.Paid, a...) } }
func logGot(a ...LogAmt) logOpt  { return func(w *logWrite) { w.e.Gained = append(w.e.Gained, a...) } }
func logLevel(l int) logOpt      { return func(w *logWrite) { w.e.Level = l } }
func logHeadIf(ok bool) logOpt {
	return func(w *logWrite) {
		if ok {
			w.e.Level = LevelHeadline
		}
	}
}

// logTotal gives a fold its own text: f is called with the summed first
// gained amount.
func logTotal(f func(gained float64) string) logOpt { return func(w *logWrite) { w.total = f } }

func logAs(t string) logOpt { return func(w *logWrite) { w.e.Type = t } }

// logNote is a line a manager hands back for the engine to write, with the
// fields it knows.
type logNote struct {
	text string
	opts []logOpt
}

func note(text string, o ...logOpt) logNote { return logNote{text, o} }

// logNotes writes the notes a manager returned, each as logType unless it
// says otherwise.
func (ge *GameEngine) logNotes(office Office, logType string, notes []logNote) {
	for _, n := range notes {
		ge.addLog(office, logType, n.text, n.opts...)
	}
}

// amts turns a resource map into amounts in key order.
func amts(m map[string]float64) []LogAmt {
	out := make([]LogAmt, 0, len(m))
	for _, k := range sortedKeys(m) {
		out = append(out, LogAmt{k, m[k]})
	}
	return out
}

func sumAmts(a, b []LogAmt) []LogAmt {
	out := append([]LogAmt(nil), a...)
	for _, x := range b {
		found := false
		for i := range out {
			if out[i].Res == x.Res {
				out[i].N += x.N
				found = true
				break
			}
		}
		if !found {
			out = append(out, x)
		}
	}
	return out
}

func (ge *GameEngine) clock() time.Time {
	if ge.now != nil {
		return ge.now()
	}
	return time.Now()
}

// SetClock replaces the clock the log stamps lines with (tests).
func (ge *GameEngine) SetClock(f func() time.Time) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.now = f
}

// addLog appends a log line (the lock held). logType is the category (see
// LogRoutine). A line is Everything when routine and Town otherwise, unless
// an option says Headline. A line with a row id joins an earlier line of the
// same office, subject and id whose newest member is at most foldTicks old.
func (ge *GameEngine) addLog(office Office, logType, message string, opts ...logOpt) {
	w := logWrite{e: LogEntry{Tick: ge.tick, Message: message, Type: logType, Office: office,
		Level: LevelTown, At: ge.clock().Unix()}}
	if logType == LogRoutine {
		w.e.Level = LevelEverything
	}
	for _, o := range opts {
		o(&w)
	}
	e := w.e
	if e.ID != "" {
		for i := len(ge.log) - 1; i >= 0 && e.Tick-ge.log[i].Tick <= foldTicks; i-- {
			o := ge.log[i]
			if o.Seam {
				break
			}
			if o.Office != e.Office || o.Subject != e.Subject || o.ID != e.ID {
				continue
			}
			ge.log = append(ge.log[:i], ge.log[i+1:]...)
			ge.log = append(ge.log, foldInto(o, e, w.total))
			return
		}
	}
	ge.log = append(ge.log, e)
	if len(ge.log) > MaxLogSize {
		ge.log = ge.log[len(ge.log)-MaxLogSize:]
	}
}

// foldInto is the earlier line o with the new line n joined to it.
func foldInto(o, n LogEntry, total func(float64) string) LogEntry {
	if o.Count < 1 {
		o.Count = 1
		o.FirstAt = o.At
		o.Members = []LogMember{{o.At, o.Tick, o.Message, o.Paid, o.Gained}}
	}
	o.Members = append(o.Members, LogMember{n.At, n.Tick, n.Message, n.Paid, n.Gained})
	if len(o.Members) > foldMembers {
		o.Members = append([]LogMember(nil), o.Members[len(o.Members)-foldMembers:]...)
	}
	o.Count++
	o.Paid, o.Gained = sumAmts(o.Paid, n.Paid), sumAmts(o.Gained, n.Gained)
	o.Tick, o.At, o.Message, o.Type = n.Tick, n.At, n.Message, n.Type
	if total != nil && len(o.Gained) > 0 {
		o.Message = total(o.Gained[0].N)
	}
	return o
}

// AddLog adds a line from outside the engine (thread-safe). The lines the
// interface writes are about the game itself or a refusal: level Everything
// unless headline.
func (ge *GameEngine) AddLog(office Office, level int, logType, message string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.addLog(office, logType, message, logLevel(level))
}

// GetLogs returns a copy of the full log (thread-safe)
func (ge *GameEngine) GetLogs() []LogEntry {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return append([]LogEntry(nil), ge.log...)
}

var txtSeam = rowid.E("game.logentry.logSeam.01", "A sitting ended.")

// logSeam closes the sitting before a load with one line: its first and last
// times and its line count. The label is made when it is drawn.
func (ge *GameEngine) logSeam() {
	n, first := 0, int64(0)
	for i := len(ge.log) - 1; i >= 0 && !ge.log[i].Seam; i-- {
		n++
		first = ge.log[i].At
	}
	if n == 0 {
		return
	}
	ge.log = append(ge.log, LogEntry{Tick: ge.tick, Message: txtSeam.String(), Type: "info", At: ge.log[len(ge.log)-1].At,
		Office: OfficeGame, Level: LevelHeadline, Seam: true, FirstAt: first, Count: n})
}

// logForSave is the log as one compact value: gzipped JSON in base64, so the
// indented save holds it on a single line.
func (ge *GameEngine) logForSave() string {
	if len(ge.log) == 0 {
		return ""
	}
	raw, _ := json.Marshal(ge.log)
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write(raw)
	_ = zw.Close()
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// logFromSave reads a saved log back; one that does not read is empty.
func logFromSave(s string) []LogEntry {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil
	}
	zr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil
	}
	raw, err := io.ReadAll(zr)
	if err != nil {
		return nil
	}
	var out []LogEntry
	if json.Unmarshal(raw, &out) != nil {
		return nil
	}
	return out[max(0, len(out)-MaxLogSize):]
}
