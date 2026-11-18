package game

import (
	"cmp"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
)

// collection.go keeps the coin collection on the account: one copy of a coin
// for each run that earned it, stacked by type and age.
//
// Storage. The copies are in a file of their own, collection.json, beside
// badges.json and runs.json, signed like the run archive and bound to the
// account's ID. A file that is missing, unreadable, signed wrongly or signed
// for another account reads as empty and is written again at the next strike:
// the collection is a record, and it never stops an account from loading.
// The collection starts empty; nothing is carried over from the badges.
//
// Striking is in memory only (it takes the account's own lock, does no I/O
// and never calls the engine), so the engine may strike under its write lock.
// The write happens in Save, from FlushIfDirty, outside that lock.

const (
	collectionFileName = "collection.json"
	// collectionVersion is the file's own schema version.
	collectionVersion = 1
	// coinStackMax is how many copies a stack keeps.
	coinStackMax = 24
)

// CoinCopy is one coin: the thing done in an age, in one run. The JSON keys
// are short because the file holds many of them.
type CoinCopy struct {
	Type  string `json:"t"`
	Age   string `json:"a"`
	Level int    `json:"l"`
	Run   string `json:"r"`
	At    int64  `json:"at"`
	// Figure is what set the level: the ticks spent in the age when it was struck.
	Figure int `json:"f"`
	// Seed is stable for a copy: see coinSeed.
	Seed uint32 `json:"s"`
	// Flags is a set of BadgeFlag bits, set as a badge's are: crossed when the
	// strike was made on an edited run or account.
	Flags int `json:"fl,omitempty"`
}

// CoinStack is every copy of one type in one age, finest first. Count is every
// copy ever struck, which the cap on Copies does not lower.
type CoinStack struct {
	Type   string     `json:"t"`
	Age    string     `json:"a"`
	Count  int        `json:"n"`
	Copies []CoinCopy `json:"c"`
}

func (s CoinStack) clone() CoinStack {
	s.Copies = slices.Clone(s.Copies)
	return s
}

// coinSeed is a 32-bit FNV-1a hash of the account, run, type and age: the same
// inputs always give the same seed.
func coinSeed(account, run, typ, age string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(strings.Join([]string{account, run, typ, age}, "\x00")))
	return h.Sum32()
}

func coinKey(typ, age string) string { return typ + "|" + age }

// coinOrder sorts copies finest first, then earliest, then by run.
func coinOrder(a, b CoinCopy) int {
	return cmp.Or(cmp.Compare(b.Level, a.Level), cmp.Compare(a.At, b.At), cmp.Compare(a.Run, b.Run))
}

// capCoins trims a stack to the most it keeps: the lowest level goes first,
// the oldest first among equals.
func capCoins(s *CoinStack) {
	slices.SortStableFunc(s.Copies, coinOrder)
	for len(s.Copies) > coinStackMax {
		drop := len(s.Copies) - 1
		for i := len(s.Copies) - 1; i >= 0 && s.Copies[i].Level == s.Copies[drop].Level; i-- {
			if s.Copies[i].At <= s.Copies[drop].At {
				drop = i
			}
		}
		s.Copies = slices.Delete(s.Copies, drop, drop+1)
	}
	s.Count = max(s.Count, len(s.Copies))
}

// strikeCoinLocked adds a copy to its stack: one copy for each run, and a
// second strike in a run keeps the finer (an equal level keeps the first).
// It returns the copy now held for the run and whether this strike changed
// the stack. Callers hold a.mu.
func (a *Account) strikeCoinLocked(c CoinCopy) (CoinCopy, bool) {
	if !config.IsCoinType(c.Type) || c.Age == "" || c.Level < 1 || c.Level > config.CoinLevels {
		return CoinCopy{}, false
	}
	key := coinKey(c.Type, c.Age)
	s, ok := a.coins[key]
	if !ok {
		s = CoinStack{Type: c.Type, Age: c.Age}
	}
	s = s.clone()
	for i, held := range s.Copies {
		if held.Run != c.Run {
			continue
		}
		if c.Level <= held.Level {
			return held, false
		}
		s.Copies[i] = c
		capCoins(&s)
		a.putCoinStackLocked(key, s)
		return c, true
	}
	s.Copies = append(s.Copies, c)
	s.Count++
	capCoins(&s)
	a.putCoinStackLocked(key, s)
	return c, slices.Contains(s.Copies, c)
}

func (a *Account) putCoinStackLocked(key string, s CoinStack) {
	if a.coins == nil {
		a.coins = map[string]CoinStack{}
	}
	a.coins[key] = s
	a.coinsChanged, a.dirty = true, true
}

// StrikeCoin strikes a coin of typ for age in run, at the level that ticksInAge
// ticks (at the age's speed) earn. It returns the copy held for the run and
// whether the strike added a copy or bettered one. crossed is the run's own
// edited-save mark, as a badge's. IN-MEMORY ONLY, like the
// Record* hooks: the engine calls it under its write lock.
func (a *Account) StrikeCoin(typ, age string, ticksInAge int, speed float64, run string, at time.Time, crossed bool) (CoinCopy, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	flags := 0
	if crossed || a.Tampered || a.BadgesTampered {
		flags = BadgeFlagCrossed
	}
	return a.strikeCoinLocked(CoinCopy{
		Flags:  flags,
		Type:   typ,
		Age:    age,
		Level:  config.CoinLevel(age, float64(ticksInAge), speed),
		Run:    run,
		At:     at.Unix(),
		Figure: ticksInAge,
		Seed:   coinSeed(a.AccountID, run, typ, age),
	})
}

// coinHeld reports whether the run already holds a coin of typ for age.
func (a *Account) coinHeld(typ, age, run string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, c := range a.coins[coinKey(typ, age)].Copies {
		if c.Run == run {
			return true
		}
	}
	return false
}

// CoinStackOf is a copy of the stack of typ in age (empty when nothing is struck).
func (a *Account) CoinStackOf(typ, age string) CoinStack {
	a.mu.Lock()
	defer a.mu.Unlock()
	s, ok := a.coins[coinKey(typ, age)]
	if !ok {
		return CoinStack{Type: typ, Age: age}
	}
	return s.clone()
}

// CoinStacksOfAge is the stacks the account holds for age, in the order of
// config.CoinTypes.
func (a *Account) CoinStacksOfAge(age string) []CoinStack {
	a.mu.Lock()
	defer a.mu.Unlock()
	var out []CoinStack
	for _, t := range config.CoinTypes {
		if s, ok := a.coins[coinKey(t, age)]; ok {
			out = append(out, s.clone())
		}
	}
	return out
}

// CoinBest is the finest level in the stack of typ in age (0 when empty).
func (a *Account) CoinBest(typ, age string) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	best := 0
	for _, c := range a.coins[coinKey(typ, age)].Copies {
		best = max(best, c.Level)
	}
	return best
}

// collectionFile is collection.json. An export carries one whole, as its CollectionStore.
type collectionFile struct {
	Version int `json:"version"`
	// AccountID is under the signature: a file found in another account's slot
	// reads as empty.
	AccountID string      `json:"account_id"`
	Stacks    []CoinStack `json:"stacks,omitempty"`
	Signature string      `json:"_sig,omitempty"`
}

func signCollection(f *collectionFile) string {
	payload := *f
	payload.Signature = ""
	data, _ := json.Marshal(&payload)
	return hmacSign(data, saveHMACKey)
}

func verifyCollection(f *collectionFile) bool {
	return f.Signature != "" && hmac.Equal([]byte(f.Signature), []byte(signCollection(f)))
}

// collectionLocked is the account's coins as a signed file, stacks in key
// order. It shares the copies: marshal it before the lock is released.
// Callers hold a.mu (or own a).
func (a *Account) collectionLocked() *collectionFile {
	f := &collectionFile{Version: collectionVersion, AccountID: a.AccountID}
	for _, k := range slices.Sorted(maps.Keys(a.coins)) {
		f.Stacks = append(f.Stacks, a.coins[k])
	}
	f.Signature = signCollection(f)
	return f
}

// coinsFromFile is the stacks of a file as the account holds them: the cap
// applied and anything the game could not have struck left out.
func coinsFromFile(f *collectionFile) map[string]CoinStack {
	out := map[string]CoinStack{}
	for _, s := range f.Stacks {
		if !config.IsCoinType(s.Type) || s.Age == "" {
			continue
		}
		key := coinKey(s.Type, s.Age)
		merged := out[key]
		merged.Type, merged.Age = s.Type, s.Age
		merged.Count = max(merged.Count, s.Count)
		merged.Copies = append(merged.Copies, s.Copies...)
		out[key] = merged
	}
	for k, s := range out {
		out[k] = foldCoinCopies(s, nil)
	}
	return out
}

// foldCoinCopies unions the copies of s and more by run (the finer copy wins,
// the held one when equal), then applies the cap.
func foldCoinCopies(s CoinStack, more []CoinCopy) CoinStack {
	byRun := map[string]CoinCopy{}
	for _, c := range slices.Concat(s.Copies, more) {
		if c.Level < 1 || c.Level > config.CoinLevels {
			continue
		}
		if held, ok := byRun[c.Run]; !ok || c.Level > held.Level || (c.Level == held.Level && c.At < held.At) {
			byRun[c.Run] = c
		}
	}
	s.Copies = slices.Collect(maps.Values(byRun))
	capCoins(&s)
	return s
}

// loadCollectionFile reads the account's coins from dir. It never writes, and
// a file that is missing, unreadable or not as the game signed it for this
// account leaves the account with none.
func (a *Account) loadCollectionFile(dir string) {
	a.coins = nil
	data, err := os.ReadFile(filepath.Join(dir, collectionFileName))
	if err != nil {
		return
	}
	var f collectionFile
	if json.Unmarshal(data, &f) != nil || f.AccountID != a.AccountID || !verifyCollection(&f) {
		return
	}
	a.coins = coinsFromFile(&f)
}

// saveCollectionFile writes the coins into dir, atomically, when they changed
// since they were read or last written. Callers hold a.mu (or own a); Save
// calls it.
func (a *Account) saveCollectionFile(dir string) error {
	if !a.coinsChanged {
		return nil
	}
	data, err := json.Marshal(a.collectionLocked())
	if err != nil {
		return fmt.Errorf("failed to marshal collection: %w", err)
	}
	if err := writeFileAtomic(filepath.Join(dir, collectionFileName), data); err != nil {
		return fmt.Errorf("failed to write collection: %w", err)
	}
	a.coinsChanged = false
	return nil
}

// takeCollectionLocked folds the coins an export carries into the account. A
// merge unions them by type, age and run (the finer copy wins) and then
// applies the cap; the count of copies ever struck is the larger of the two,
// so importing a backup of the same account adds nothing. Otherwise the
// export's become the account's. An export with none leaves the account's as
// they are. Callers hold a.mu (or own a).
func (a *Account) takeCollectionLocked(f *collectionFile, merge bool) {
	if f == nil {
		return
	}
	theirs := coinsFromFile(f)
	if !merge {
		a.coins = theirs
	} else {
		if a.coins == nil {
			a.coins = map[string]CoinStack{}
		}
		for k, t := range theirs {
			mine := a.coins[k].clone()
			mine.Type, mine.Age = t.Type, t.Age
			count := max(mine.Count, t.Count)
			mine = foldCoinCopies(mine, t.Copies)
			mine.Count = max(count, len(mine.Copies))
			a.coins[k] = mine
		}
	}
	a.coinsChanged, a.dirty = true, true
}

// mergeCoinStacks is the union of two sets of stacks by type, age and run,
// the finer copy winning, with the cap applied. Neither argument is changed.
func mergeCoinStacks(a, b map[string]CoinStack) map[string]CoinStack {
	if len(a) == 0 && len(b) == 0 {
		return a
	}
	out := map[string]CoinStack{}
	for _, m := range []map[string]CoinStack{a, b} {
		for k, s := range m {
			held := out[k]
			held.Type, held.Age = s.Type, s.Age
			count := max(held.Count, s.Count)
			held = foldCoinCopies(held, s.Copies)
			held.Count = max(count, len(held.Copies))
			out[k] = held
		}
	}
	return out
}
