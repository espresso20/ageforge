package game

import (
	"cmp"
	"crypto/hmac"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"
)

// run_archive.go files a small copy of every finished run on the account, so a
// screen can lay past runs beside the current one.
//
// Storage. The records are in a file of their own, runs.json, beside account.json
// and badges.json in the account's slot, signed like the badge file and bound to
// the account's ID. Nothing in account.json changes. A file that is missing,
// unreadable, signed wrongly or signed for another account reads as empty and is
// written again at the next filing: the archive is a record, and it never stops
// an account from loading.
//
// Filing is in memory only (Account.fileRun takes the account's own lock, does
// no I/O and never calls the engine), so the engine may file a run under its
// write lock. The write happens in Save, from FlushIfDirty, outside that lock.

const (
	runArchiveFileName = "runs.json"
	// runArchiveVersion is the file's own schema version.
	runArchiveVersion = 1
	// runArchiveMinBuckets is how many run-tier buckets a run needs before one
	// that did not end in a prestige is worth filing.
	runArchiveMinBuckets = 30
	// runArchiveBuckets is the most buckets a filed history keeps.
	runArchiveBuckets = 256
	// runArchiveFull is how many of the newest records keep their history; the
	// best run keeps its own as well.
	runArchiveFull = 20
	// runArchiveMax is how many records the account keeps.
	runArchiveMax = 200
)

// How a filed run ended.
const (
	RunPrestige  = "prestige"  // the run was prestiged from (the Last Passage's endings too)
	RunSuccumbed = "succumbed" // the civilization fell to a catastrophe
	RunLeft      = "left"      // the game was wiped or a new one started over it
)

// runArchiveCols are the history columns a filed run keeps; runArchiveMarks the
// kinds of mark.
var (
	runArchiveCols  = []string{"population", "rate.food", "know_rate", "faith", "morale", "prod_all", "tick_speed", "buildings", "techs", "expeditions"}
	runArchiveMarks = []string{MarkAge, MarkWonder, MarkCatastrophe, MarkHarbinger}
)

// RunRecord is one finished run as the account keeps it.
type RunRecord struct {
	ID string `json:"id"`
	// Run is the save the run was played in, as the records name a run.
	Run string `json:"run,omitempty"`
	// Started and Ended are wall-clock Unix seconds; Tick is the run's last tick.
	Started int64  `json:"started"`
	Ended   int64  `json:"ended"`
	Tick    int    `json:"tick"`
	Outcome string `json:"outcome"`
	// Age is the highest age the run reached.
	Age         string `json:"age"`
	Built       int    `json:"built"`
	Recruited   int    `json:"recruited"`
	Researched  int    `json:"researched"`
	Expeditions int    `json:"expeditions"`
	// Level is the prestige level the run was played at.
	Level int `json:"level"`
	// History is the run's thinned history (runArchiveCols, runArchiveMarks, at
	// most runArchiveBuckets buckets); only the run tier is filled. nil once the
	// record is older than the account keeps histories for.
	History *RunHistory `json:"history,omitempty"`
}

func (r RunRecord) clone() RunRecord {
	r.History = r.History.Clone()
	return r
}

// newRunID is eight random bytes as hex.
func newRunID() string {
	var b [8]byte
	if _, err := crand.Read(b[:]); err != nil {
		binary.BigEndian.PutUint64(b[:], uint64(time.Now().UnixNano()))
	}
	return hex.EncodeToString(b[:])
}

// archived is the history as a finished run files it: the run tier alone, cut
// to the kept columns and thinned, with the kept kinds of mark.
func (h *RunHistory) archived() *RunHistory {
	out := NewRunHistory()
	run := h.Run
	out.Run = HistoryTier{Span: run.Span, N: run.N, Fill: run.Fill, End: run.End, Cols: map[string][]float32{}}
	for _, k := range runArchiveCols {
		if c, ok := run.Cols[k]; ok {
			out.Run.Cols[k] = slices.Clone(c)
		}
	}
	out.Run.thin(runArchiveBuckets)
	for _, m := range h.Marks {
		if slices.Contains(runArchiveMarks, m.Kind) {
			out.Marks = append(out.Marks, m)
		}
	}
	return out
}

// fileRunLocked files the run in memory as ended, at the level it was played at,
// unless it is too short to be worth keeping (a prestige is always kept). A run
// that does not record to the account is not filed. Callers hold the write lock.
func (ge *GameEngine) fileRunLocked(outcome string, level int) {
	acct := ge.accountForRecordsLocked()
	if acct == nil || (outcome != RunPrestige && ge.history.Run.N < runArchiveMinBuckets) {
		return
	}
	acct.fileRun(RunRecord{
		ID:          ge.runID,
		Run:         ge.activeSaveName,
		Started:     ge.Stats.GameStarted.Unix(),
		Ended:       badgeClock().Unix(),
		Tick:        ge.tick,
		Outcome:     outcome,
		Age:         ge.age,
		Built:       ge.Stats.TotalBuilt,
		Recruited:   ge.Stats.TotalRecruited,
		Researched:  len(ge.Research.researched),
		Expeditions: ge.Military.completedCount,
		Level:       level,
		History:     ge.history.archived(),
	})
}

// RunArchiveView is the read-only projection of the account's filed runs: a copy
// the caller may keep or change.
type RunArchiveView struct {
	// RunID is the run in play now.
	RunID string
	// Runs is the filed runs, newest first.
	Runs []RunRecord
}

// RunArchive returns the filed runs, newest first, and the id of the current run.
func (ge *GameEngine) RunArchive() RunArchiveView {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	v := RunArchiveView{RunID: ge.runID}
	if ge.account != nil {
		v.Runs = ge.account.runRecords()
	}
	return v
}

// runArchiveFile is runs.json. An export carries one whole, as its RunStore.
type runArchiveFile struct {
	Version int `json:"version"`
	// AccountID is under the signature: a file found in another account's slot
	// reads as empty.
	AccountID string      `json:"account_id"`
	Runs      []RunRecord `json:"runs,omitempty"`
	Signature string      `json:"_sig,omitempty"`
}

func signRunArchive(f *runArchiveFile) string {
	payload := *f
	payload.Signature = ""
	data, _ := json.Marshal(&payload)
	return hmacSign(data, saveHMACKey)
}

func verifyRunArchive(f *runArchiveFile) bool {
	return f.Signature != "" && hmac.Equal([]byte(f.Signature), []byte(signRunArchive(f)))
}

// runArchiveLocked is the account's records as a signed file. It shares the
// records: marshal it before the lock is released. Callers hold a.mu (or own a).
func (a *Account) runArchiveLocked() *runArchiveFile {
	f := &runArchiveFile{Version: runArchiveVersion, AccountID: a.AccountID, Runs: a.runs}
	f.Signature = signRunArchive(f)
	return f
}

// loadRunFile reads the account's records from dir. It never writes, and a file
// that is missing, unreadable or not as the game signed it for this account
// leaves the account with none.
func (a *Account) loadRunFile(dir string) {
	a.runs = nil
	data, err := os.ReadFile(filepath.Join(dir, runArchiveFileName))
	if err != nil {
		return
	}
	var f runArchiveFile
	if json.Unmarshal(data, &f) != nil || f.AccountID != a.AccountID || !verifyRunArchive(&f) {
		return
	}
	a.runs = limitRuns(f.Runs)
}

// saveRunFile writes the records into dir, atomically, when they changed since
// they were read or last written. Callers hold a.mu (or own a); Save calls it.
func (a *Account) saveRunFile(dir string) error {
	if !a.runsChanged {
		return nil
	}
	data, err := json.Marshal(a.runArchiveLocked())
	if err != nil {
		return fmt.Errorf("failed to marshal runs: %w", err)
	}
	if err := writeFileAtomic(filepath.Join(dir, runArchiveFileName), data); err != nil {
		return fmt.Errorf("failed to write runs: %w", err)
	}
	a.runsChanged = false
	return nil
}

// fileRun adds a finished run to the account (one with the same ID is replaced)
// and marks the account for the next flush. IN-MEMORY ONLY, like the Record*
// hooks: the engine calls it under its write lock.
func (a *Account) fileRun(r RunRecord) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.runs = limitRuns(foldRuns(a.runs, []RunRecord{r}, true))
	a.runsChanged, a.dirty = true, true
}

// runRecords is a copy of the account's records, newest first.
func (a *Account) runRecords() []RunRecord {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]RunRecord, len(a.runs))
	for i, r := range a.runs {
		out[i] = r.clone()
	}
	return out
}

// takeRunStoreLocked folds the records an export carries into the account: a
// merge unions them by run ID, otherwise the export's become the account's.
// An export with none leaves the account's as they are. Callers hold a.mu (or own a).
func (a *Account) takeRunStoreLocked(f *runArchiveFile, merge bool) {
	if f == nil {
		return
	}
	if merge {
		a.runs = limitRuns(foldRuns(a.runs, f.Runs, false))
	} else {
		a.runs = limitRuns(foldRuns(nil, f.Runs, false))
	}
	a.runsChanged, a.dirty = true, true
}

// foldRuns adds in to have by run ID. On a clash the incoming record wins when
// replace is set, otherwise the one with the larger end tick (the held one on a
// tie). The records are copied, so neither side is shared with the result.
func foldRuns(have, in []RunRecord, replace bool) []RunRecord {
	out := make([]RunRecord, 0, len(have)+len(in))
	for _, r := range have {
		out = append(out, r.clone())
	}
	for _, r := range in {
		if i := slices.IndexFunc(out, func(x RunRecord) bool { return x.ID == r.ID }); i < 0 {
			out = append(out, r.clone())
		} else if replace || r.Tick > out[i].Tick {
			out[i] = r.clone()
		}
	}
	return out
}

// limitRuns orders the records newest first, keeps the newest runArchiveMax,
// and strips the history from all but the newest runArchiveFull and the best run.
func limitRuns(runs []RunRecord) []RunRecord {
	runs = slices.Clone(runs)
	slices.SortStableFunc(runs, func(a, b RunRecord) int {
		if a.Ended != b.Ended {
			return cmp.Compare(b.Ended, a.Ended)
		}
		return cmp.Compare(b.ID, a.ID)
	})
	if len(runs) > runArchiveMax {
		runs = runs[:runArchiveMax]
	}
	best := -1
	for i, r := range runs {
		if best < 0 || betterRun(r, runs[best]) {
			best = i
		}
	}
	for i := runArchiveFull; i < len(runs); i++ {
		if i != best {
			runs[i].History = nil
		}
	}
	return runs
}

// betterRun reports whether a is a better run than b: the higher age, then the
// fewer ticks.
func betterRun(a, b RunRecord) bool {
	if oa, ob := ageOrderOf(a.Age), ageOrderOf(b.Age); oa != ob {
		return oa > ob
	}
	return a.Tick < b.Tick
}
