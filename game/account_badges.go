package game

import (
	"bytes"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
)

// account_badges.go is the account's side of the badges: where they are
// stored, how an account from before them gets its own, and how an event is
// judged against it. Like the lifetime-stat hooks, judging is in memory
// only: it takes the account's own lock, touches no file and calls nothing
// on the engine, so the engine may call it under its write lock. The write
// happens later, in FlushIfDirty.
//
// Storage. The badges, the counters behind them and the days played are in
// a file of their own, badges.json, beside account.json in the account's
// slot. account.json is not changed by any of this: it holds and signs
// exactly what it did before badges, so a build of the game from before
// them reads it, verifies it and writes it back as it always did, never
// sees badges.json, and cannot mark a healthy account as modified. When
// this build next loads the account it reconciles the two (ensureBadges):
// whatever the account's record proves, an achievement the older build
// earned included, is topped up in the badge file.
//
// The badge file is signed on its own, and the signature covers the account
// ID, so it cannot be carried to another account. An edited badge file is
// flagged (BadgesTampered) and its badges crossed; an edited account.json
// flags the account as it always did. Neither flags the other.

// badgeFileName is the badge file's name in an account's slot:
// <root>/accounts/<account_id>/badges.json.
const badgeFileName = "badges.json"

// badgeFileVersion is the badge file's own schema version.
const badgeFileVersion = 1

// badgeFile is badges.json. An export carries one whole, as its BadgeStore.
type badgeFile struct {
	Version int `json:"version"`
	// AccountID is the account the file belongs to. It is under the
	// signature: a badge file found in another account's slot is treated
	// as edited.
	AccountID string                 `json:"account_id"`
	Badges    map[string]BadgeEarned `json:"badges,omitempty"`
	Counters  map[string]float64     `json:"counters,omitempty"`
	Days      []string               `json:"days,omitempty"`
	// Tampered is the file's sticky tamper mark (Account.BadgesTampered).
	Tampered  bool   `json:"tampered,omitempty"`
	Signature string `json:"_sig,omitempty"`
}

// signBadgeFile is the HMAC of the badge file with its signature zeroed:
// the same construction as account.json's and a save's.
func signBadgeFile(b *badgeFile) string {
	payload := *b
	payload.Signature = ""
	data, _ := json.Marshal(&payload)
	return hmacSign(data, saveHMACKey)
}

// verifyBadgeFile reports whether a badge file is as the game signed it. A
// file with no signature is not: the game never wrote one unsigned.
func verifyBadgeFile(b *badgeFile) bool {
	return b.Signature != "" && hmac.Equal([]byte(b.Signature), []byte(signBadgeFile(b)))
}

// badgeFileLocked is the account's badges as a badge file, unsigned; nil
// when there is nothing to store. It shares the account's maps: marshal it
// before the lock is released. Callers hold a.mu (or own a).
func (a *Account) badgeFileLocked() *badgeFile {
	if len(a.Badges) == 0 && len(a.Counters) == 0 && len(a.Days) == 0 && !a.BadgesTampered {
		return nil
	}
	b := &badgeFile{
		Version:   badgeFileVersion,
		AccountID: a.AccountID,
		Badges:    a.Badges,
		Counters:  a.Counters,
		Days:      a.Days,
		Tampered:  a.BadgesTampered,
	}
	b.Signature = signBadgeFile(b)
	return b
}

// loadBadgeFile reads the badge file in dir into the account, after
// account.json has given it its ID. It never writes.
//
//   - No file: the account has no badge store yet (it is from before
//     badges, or has earned nothing). ensureBadges fills it from the record.
//   - A file as the game signed it, for this account: taken as it is.
//   - A file whose signature does not match, or signed for another account:
//     taken, flagged (BadgesTampered), and every badge in it crossed. The
//     next save writes the flag and the crosses, so they stick.
//   - A file that does not parse: left where it is and ignored. The next
//     save that has badges to write sets it aside as badges.json.corrupt.
func (a *Account) loadBadgeFile(dir string) {
	data, err := os.ReadFile(filepath.Join(dir, badgeFileName))
	if err != nil {
		a.badgeUnreadable = !os.IsNotExist(err)
		return
	}
	a.badgeDisk = data
	var b badgeFile
	if err := json.Unmarshal(data, &b); err != nil {
		a.badgeUnreadable = true
		return
	}
	a.Badges, a.Counters, a.Days = b.Badges, b.Counters, b.Days
	a.badgeRev++
	a.BadgesTampered = b.Tampered || b.AccountID != a.AccountID || !verifyBadgeFile(&b)
	if a.BadgesTampered {
		for key, e := range a.Badges {
			e.Flags |= BadgeFlagCrossed
			a.Badges[key] = e
		}
	}
}

// saveBadgeFile writes the badge file into dir, atomically, when the badges
// differ from what the file holds. An account with no badges and no file
// gets none. Callers hold a.mu (or own a); Save calls it.
func (a *Account) saveBadgeFile(dir string) error {
	b := a.badgeFileLocked()
	if b == nil {
		if a.badgeDisk == nil || a.badgeUnreadable {
			return nil // nothing to store, and nothing of ours on disk to clear
		}
		// The store was emptied (an import that replaced it): write it empty.
		b = &badgeFile{Version: badgeFileVersion, AccountID: a.AccountID}
		b.Signature = signBadgeFile(b)
	}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal badges: %w", err)
	}
	if !a.badgeUnreadable && bytes.Equal(data, a.badgeDisk) {
		return nil
	}
	path := filepath.Join(dir, badgeFileName)
	if a.badgeUnreadable {
		// Never overwrite a file we could not read: set it aside first.
		if err := os.Rename(path, path+".corrupt"); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("the badge file is damaged and could not be set aside: %w", err)
		}
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("failed to write badges: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("failed to finalize badges: %w", err)
	}
	a.badgeDisk, a.badgeUnreadable = data, false
	return nil
}

// exportBadgeFileLocked is the badge file an export carries: always one,
// signed, even when the account has no badges, so that an export with no
// badge store can only be one from before badges. Callers hold a.mu.
func (a *Account) exportBadgeFileLocked() *badgeFile {
	if b := a.badgeFileLocked(); b != nil {
		return b
	}
	b := &badgeFile{Version: badgeFileVersion, AccountID: a.AccountID}
	b.Signature = signBadgeFile(b)
	return b
}

// adoptBadgeFile is for an account that is being created in memory for a
// slot that may already hold a badge file (its account.json was lost or
// set aside, and the account is made again under the same name): the file
// is read and folded into whatever the new account carries, so the next
// save cannot write over badges it never loaded. Callers own a.
func (a *Account) adoptBadgeFile() {
	badges, counters, days, flagged := a.Badges, a.Counters, a.Days, a.BadgesTampered
	a.Badges, a.Counters, a.Days, a.BadgesTampered = nil, nil, nil, false
	a.loadBadgeFile(accountDir(a.AccountID))
	a.Badges = mergeBadges(a.Badges, badges)
	a.Counters = mergeCounters(a.Counters, counters)
	a.Days = mergeDays(a.Days, days)
	a.BadgesTampered = a.BadgesTampered || flagged
	a.badgeRev++
}

// takeBadgeStoreLocked folds the badge store an export carries into the
// account. A merge keeps every badge either copy holds (the earlier of
// two) and the larger of each counter; otherwise the account's store
// becomes the export's. An export from before badges carries none (b is
// nil): it says nothing about badges, so the account's are left as they
// are, replace or not. A flagged store flags the account's, and a flagged
// account's stays flagged. Callers hold a.mu (or own a).
func (a *Account) takeBadgeStoreLocked(b *badgeFile, merge bool) {
	if b == nil {
		return
	}
	a.BadgesTampered = a.BadgesTampered || b.Tampered
	a.badgeRev++
	if !merge {
		a.Badges = mergeBadges(b.Badges, nil)
		a.Counters = mergeCounters(b.Counters, nil)
		a.Days = mergeDays(b.Days, nil)
		return
	}
	a.Badges = mergeBadges(a.Badges, b.Badges)
	a.Counters = mergeCounters(a.Counters, b.Counters)
	a.Days = mergeDays(a.Days, b.Days)
}

// BadgeEarned is one earned badge as the badge file stores it.
type BadgeEarned struct {
	// At is when it was earned, in Unix seconds. 0 means before badges were
	// dated: it came over from an account achievement, or the account's
	// record already proved it.
	At int64 `json:"at"`
	// Run is the save it was earned in.
	Run string `json:"run,omitempty"`
	// Flags is a set of BadgeFlag bits.
	Flags int `json:"f,omitempty"`
}

// BadgeFlagCrossed marks a badge earned in a save edited outside the game,
// on an account whose file was, or held in a badge file that was. It shows
// struck through and adds no points.
const BadgeFlagCrossed = 1

// maxAccountDays is how many check-in days the account keeps: enough for
// any streak a badge asks for.
const maxAccountDays = 400

// earnedLocked reports whether the account holds the badge. Callers hold a.mu.
func (a *Account) earnedLocked(key string) bool {
	_, ok := a.Badges[key]
	return ok
}

// judge judges an event against the account: it moves the lifetime
// counters the event feeds and grants every badge it earns. In memory only.
func (a *Account) judge(book *badgeBook, ev Event, ctx badgeCtx) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.judgeLocked(book, ev, ctx, 0)
}

// maxBadgeChain bounds how far one badge may lead to another (a badge for
// earning badges) inside one judgment.
const maxBadgeChain = 8

func (a *Account) judgeLocked(book *badgeBook, ev Event, ctx badgeCtx, depth int) {
	a.countLocked(book, ev.Kind, ev.amount(), ctx, depth)
	if ev.Subject != "" {
		a.countLocked(book, ev.Kind+"."+ev.Subject, ev.amount(), ctx, depth)
	}
	for _, i := range book.byEvent[ev.Kind] {
		def := &book.defs[i]
		if a.earnedLocked(def.Key) || !a.meetsLocked(def, ev, ctx) {
			continue
		}
		a.grantLocked(book, def, ctx, false, depth)
	}
}

// countLocked adds n to a lifetime counter, if a badge keeps it, and grants
// the ladder rungs it now reaches.
func (a *Account) countLocked(book *badgeBook, name string, n float64, ctx badgeCtx, depth int) {
	if !book.counters[name] || n <= 0 {
		return
	}
	if a.Counters == nil {
		a.Counters = map[string]float64{}
	}
	a.Counters[name] += n
	a.dirty = true
	a.badgeRev++
	for _, i := range book.byCounter[name] {
		def := &book.defs[i]
		if a.Counters[name] < def.Threshold {
			break // sorted by threshold: the rest are higher
		}
		if !a.earnedLocked(def.Key) {
			a.grantLocked(book, def, ctx, false, depth)
		}
	}
}

// meetsLocked reports whether a Run or Moment badge's row holds for an event.
func (a *Account) meetsLocked(def *config.BadgeDef, ev Event, ctx badgeCtx) bool {
	if def.Subject != "" && !def.AnySubject && def.Subject != ev.Subject {
		return false
	}
	if def.InAge != "" && def.InAge != ctx.age {
		return false
	}
	if def.Scope == config.BadgeRun && def.Counter != "" {
		if v, _ := a.factLocked(def.Counter, ev, ctx); v < def.Threshold {
			return false
		}
	}
	for _, c := range def.When {
		v, known := a.factLocked(c.Fact, ev, ctx)
		switch c.Op {
		case config.BadgeAtLeast:
			if v < c.Value {
				return false
			}
		case config.BadgeAtMost:
			// "Never" needs the whole story: a fact the run cannot
			// vouch for does not pass.
			if !known || v > c.Value {
				return false
			}
		default:
			return false
		}
	}
	if def.Pred != "" && (ctx.pred == nil || !ctx.pred(def)) {
		return false
	}
	return true
}

// factLocked reads a fact by name: an attribute of the event being judged,
// a lifetime counter or a fact of the account's own from the account, and
// anything else from the run.
func (a *Account) factLocked(name string, ev Event, ctx badgeCtx) (float64, bool) {
	switch {
	case strings.HasPrefix(name, factEvent):
		v, ok := ev.Attrs[strings.TrimPrefix(name, factEvent)]
		return v, ok
	case strings.HasPrefix(name, factLife):
		return a.Counters[strings.TrimPrefix(name, factLife)], true
	case strings.HasPrefix(name, factAccount):
		return a.accountFactLocked(strings.TrimPrefix(name, factAccount))
	case strings.HasPrefix(name, factRun) && strings.HasSuffix(name, factAnySubject):
		// The run's tally for whatever this event happened to.
		name = strings.TrimSuffix(name, factAnySubject) + "." + ev.Subject
	}
	if ctx.fact == nil {
		return 0, false
	}
	return ctx.fact(name)
}

// accountFactLocked reads a fact the account itself knows:
//
//	days_within.<n>   the different days played among the n calendar days
//	                  that end on the last day played
func (a *Account) accountFactLocked(name string) (float64, bool) {
	if span, ok := strings.CutPrefix(name, "days_within."); ok {
		n, err := strconv.Atoi(span)
		if err != nil || n <= 0 || len(a.Days) == 0 {
			return 0, err == nil
		}
		last, err := time.Parse(accountDayLayout, a.Days[len(a.Days)-1])
		if err != nil {
			return 0, false
		}
		from := last.AddDate(0, 0, -(n - 1)).Format(accountDayLayout)
		count := 0
		for _, d := range a.Days {
			if d >= from {
				count++
			}
		}
		return float64(count), true
	}
	return 0, false
}

// grantLocked gives the account a badge. A silent grant is one the record
// already proved (migration, seeding): it is not dated and raises no toast.
// Earning a badge is an event of its own, so a badge for earning badges is
// a plain lifetime ladder.
func (a *Account) grantLocked(book *badgeBook, def *config.BadgeDef, ctx badgeCtx, silent bool, depth int) {
	if a.Badges == nil {
		a.Badges = map[string]BadgeEarned{}
	}
	e := BadgeEarned{}
	if !silent {
		e.At, e.Run = badgeClock().Unix(), ctx.run
	}
	if ctx.crossed || a.Tampered || a.BadgesTampered {
		e.Flags |= BadgeFlagCrossed
	}
	a.Badges[def.Key] = e
	a.dirty = true
	a.badgeRev++
	// A badge that was an account achievement before badges keeps its place in
	// that list too, so a build from before badges shows what this one earned.
	// Not a crossed one: account.json has no way to mark it, and the list is
	// what the badges are rebuilt from if the badge file is lost, so a crossed
	// badge written there would come back clean.
	if e.Flags&BadgeFlagCrossed == 0 {
		for _, old := range def.Aliases {
			if !slices.Contains(a.Achievements, old) {
				a.Achievements = append(a.Achievements, old)
			}
		}
	}
	// A theme the badge gives joins the account's unlocked themes.
	if t := def.Reward.Theme; t != "" && !a.hasThemeLocked(t) {
		a.Unlocks.Themes = append(a.Unlocks.Themes, t)
	}
	if !silent {
		a.pendingEarned = append(a.pendingEarned, def.Key)
	}
	if def.Integrity() || depth >= maxBadgeChain {
		return
	}
	set := ""
	if def.Set != "" {
		set = config.BadgeEvBadge + ".set." + def.Set
	}
	if silent {
		// A badge the record proved leads only to others it proves.
		a.countSilentLocked(book, config.BadgeEvBadge, depth)
		a.countSilentLocked(book, config.BadgeEvBadge+"."+def.Family, depth)
		if set != "" {
			a.countSilentLocked(book, set, depth)
		}
		return
	}
	if set != "" {
		a.countLocked(book, set, 1, ctx, depth+1)
	}
	top := 0.0
	if pos := book.rung[def.Key]; pos[1] > 0 && pos[0] == pos[1] {
		top = 1
	}
	a.judgeLocked(book, Event{Kind: config.BadgeEvBadge, Subject: def.Family, Attrs: map[string]float64{"top": top}}, ctx, depth+1)
}

// countSilentLocked is countLocked for a silent grant: what it reaches is
// granted silently too.
func (a *Account) countSilentLocked(book *badgeBook, name string, depth int) {
	if !book.counters[name] {
		return
	}
	if a.Counters == nil {
		a.Counters = map[string]float64{}
	}
	a.Counters[name]++
	a.badgeRev++
	for _, i := range book.byCounter[name] {
		def := &book.defs[i]
		if a.Counters[name] < def.Threshold {
			break
		}
		if !a.earnedLocked(def.Key) {
			a.grantLocked(book, def, badgeCtx{}, true, depth+1)
		}
	}
}

// badgeClock is the clock a badge is dated by: the wall clock, unless a
// test has pinned it (SetBadgeClockForTest).
var badgeClock = time.Now

// SetBadgeClockForTest dates the badges earned from now on by now instead
// of the wall clock, and returns a function that puts the wall clock back.
// The wiki's pictures are drawn with it, so the date on a badge in them is
// the same on every run.
func SetBadgeClockForTest(now func() time.Time) (restore func()) {
	prev := badgeClock
	badgeClock = now
	return func() { badgeClock = prev }
}

// drainEarned returns the keys of the badges earned since the last call.
func (a *Account) drainEarned() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := a.pendingEarned
	a.pendingEarned = nil
	return out
}

// drainCaughtUp returns how many badges reconciliation has given the
// account since the last call, and forgets the count.
func (a *Account) drainCaughtUp() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	n := a.pendingCaughtUp
	a.pendingCaughtUp = 0
	return n
}

// ensureBadges reconciles the account's badges with its record, under the
// ruleset's badges, and reports whether it changed anything. It is what
// gives an account from before badges its badge file, and what picks up
// whatever a build from before badges did to account.json since. It is
// safe to run on every load: a second run changes nothing.
//
//   - The four account achievements are their badges (each badge lists the
//     key it had as an alias). The list stays in account.json.
//   - The lifetime stats the account already kept seed the counters that
//     continue them (prestiges).
//   - What the record already proves is granted: the badge of every age up
//     to the highest reached, and every ladder rung a counter has passed.
//
// Everything it grants is silent: undated, and no toast of its own. However
// many it grants, the dashboard says so in one line (pendingCaughtUp): a
// ladder whose rungs were lowered can give an account a dozen at once. A
// badge granted on
// an account flagged as edited, or into a badge file flagged as edited, is
// crossed, like any other earned there; on a healthy account nothing is.
// In memory only; the caller flushes.
func (a *Account) ensureBadges(book *badgeBook) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.ensureBadgesLocked(book)
}

func (a *Account) ensureBadgesLocked(book *badgeBook) bool {
	changed := false
	// What this pass grants is announced in one line, however many there
	// are (pendingCaughtUp): counted off the badges held before and after,
	// so a badge one of them leads to is counted too.
	held := len(a.Badges)
	defer func() { a.pendingCaughtUp += max(len(a.Badges)-held, 0) }()
	grant := func(def *config.BadgeDef) {
		if def != nil && !a.earnedLocked(def.Key) {
			a.grantLocked(book, def, badgeCtx{}, true, 0)
			changed = true
		}
	}
	for _, old := range a.Achievements {
		if key, ok := book.set.BadgeForAlias(old); ok {
			grant(book.def(key))
		}
	}
	seed := func(name string, n float64) {
		if book.counters[name] && a.Counters[name] < n {
			if a.Counters == nil {
				a.Counters = map[string]float64{}
			}
			a.Counters[name] = n
			a.badgeRev++
			changed = true
		}
	}
	seed(config.BadgeEvPrestige, float64(a.Stats.TotalPrestiges))
	for _, age := range sortedKeys(a.Stats.PrestigesByAge) {
		seed(config.BadgeEvPrestige+"."+age, float64(a.Stats.PrestigesByAge[age]))
	}
	highest, reached := book.set.Index(a.Stats.HighestAge)
	for i := range book.defs {
		def := &book.defs[i]
		switch {
		case def.Integrity():
		case def.Scope == config.BadgeLifetime:
			if def.Counter != "" && a.Counters[def.Counter] >= def.Threshold {
				grant(def)
			}
		case def.Scope == config.BadgeMoment && def.Event == config.BadgeEvAgeReached &&
			def.Subject != "" && def.InAge == "" && len(def.When) == 0 && def.Pred == "":
			if pos, ok := book.set.Index(def.Subject); ok && reached && pos <= highest {
				grant(def)
			}
		}
	}
	// The badges already held count toward the badges for earning badges: an
	// account from before those existed kept none of their counters. A badge
	// this grants is held too, so it goes round until nothing is added.
	for range maxBadgeChain {
		held := map[string]float64{}
		for key := range a.Badges {
			def := book.def(key)
			if def == nil || def.Integrity() {
				continue
			}
			held[config.BadgeEvBadge]++
			held[config.BadgeEvBadge+"."+def.Family]++
			if def.Set != "" {
				held[config.BadgeEvBadge+".set."+def.Set]++
			}
		}
		grew := false
		for _, name := range sortedKeys(held) {
			if !book.counters[name] || a.Counters[name] >= held[name] {
				continue
			}
			if a.Counters == nil {
				a.Counters = map[string]float64{}
			}
			a.Counters[name] = held[name]
			a.badgeRev++
			grew, changed = true, true
			for _, i := range book.byCounter[name] {
				if def := &book.defs[i]; a.Counters[name] >= def.Threshold {
					grant(def)
				}
			}
		}
		if !grew {
			break
		}
	}
	// A theme an earned badge gives is unlocked: this is what gives the theme
	// to an account that earned the badge before it gave one.
	for _, key := range sortedKeys(a.Badges) {
		if def := book.def(key); def != nil {
			if t := def.Reward.Theme; t != "" && !a.hasThemeLocked(t) {
				a.Unlocks.Themes = append(a.Unlocks.Themes, t)
				changed = true
			}
		}
	}
	if changed {
		a.dirty = true
	}
	return changed
}

// noteDay records that the account was played on day ("2026-10-07") and
// returns whether it is new. The list is kept sorted and bounded.
func (a *Account) noteDay(day string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if day == "" || slices.Contains(a.Days, day) {
		return false
	}
	a.Days = append(a.Days, day)
	sort.Strings(a.Days)
	if len(a.Days) > maxAccountDays {
		a.Days = slices.Clone(a.Days[len(a.Days)-maxAccountDays:])
	}
	a.dirty = true
	return true
}

// mergeBadges folds another copy of the account's badges into a: every
// badge either side holds, keeping the earlier of two copies (undated is
// earliest).
func mergeBadges(a, b map[string]BadgeEarned) map[string]BadgeEarned {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	out := make(map[string]BadgeEarned, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if cur, ok := out[k]; !ok || v.At < cur.At {
			out[k] = v
		}
	}
	return out
}

// mergeCounters takes the larger of each counter. Not the sum: importing
// your own backup must not double anything.
func mergeCounters(a, b map[string]float64) map[string]float64 {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	out := make(map[string]float64, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v > out[k] {
			out[k] = v
		}
	}
	return out
}

// mergeDays is the union of two day lists, sorted and bounded.
func mergeDays(a, b []string) []string {
	out := unionStrings(a, b)
	if len(out) == 0 {
		return nil
	}
	sort.Strings(out)
	if len(out) > maxAccountDays {
		out = slices.Clone(out[len(out)-maxAccountDays:])
	}
	return out
}

// BadgeView is one badge as the player may see it. For a locked badge the
// spoiler rules still withhold, Hidden is set and the name and description
// are not in the view: Name is "???" and Desc is empty, or the hint for a
// secret badge.
type BadgeView struct {
	Key    string
	Family string
	Name   string
	Desc   string
	// Tier and Rarity in words; both "" for an integrity badge.
	Tier   string
	Rarity string
	Points int
	Earned bool
	// At is when it was earned; zero for a badge earned before badges
	// were dated. Run is the save it was earned in.
	At  time.Time
	Run string
	// Crossed: earned in an edited save or on an edited account. It adds
	// no points.
	Crossed bool
	Hidden  bool
	Secret  bool
	// Integrity badges are worth nothing, count toward nothing and are
	// listed only once earned.
	Integrity bool
	// Progress and Target are a counted goal's count and threshold, for a
	// lifetime badge that shows and is not earned. Target 0: no count.
	Progress float64
	Target   float64
	// Level is the tier as a number, for the frame the badge case draws;
	// config.BadgeNoTier for an integrity badge.
	Level config.BadgeTier
	// Emblem names the symbol the badge case draws (config.BadgeDef.Emblem).
	// A silhouette has none: a symbol would say what the badge is about.
	Emblem string
	// Ladder is the name of the ladder the badge is a rung of, Rung its
	// place on it from 1 and Rungs how many the ladder has. "" and 0 for
	// a badge on no ladder, and for a silhouette.
	Ladder string
	Rung   int
	Rungs  int
	// RewardTheme and RewardTitle are what the badge gives: a theme's key
	// and a title. Both "" for a silhouette.
	RewardTheme string
	RewardTitle string
}

// BadgeSummary is the totals of a badge list. Integrity badges are in none
// of them.
type BadgeSummary struct {
	// Earned is the badges earned; Shown the badges whose text shows
	// (earned or revealed); Hidden the rest.
	Earned int
	Shown  int
	Hidden int
	// HiddenCounted is false while the size of the hidden part is itself
	// withheld ("??? hidden"): until the account has reached the last age.
	HiddenCounted bool
	// Points is the score: every earned, uncrossed badge's points.
	Points int
	// Title is the title the score holds, NextTitle the one after it and
	// NextTitleAt the score that asks for ("" and 0 at the top). An
	// account that holds every badge that counts, none of them crossed,
	// holds a title of its own.
	Title       string
	NextTitle   string
	NextTitleAt int
	// TitleRank is the title's place among the titles: 0 for the one
	// every account starts with.
	TitleRank int
	// Titles is every title the account holds: the score's, then the ones
	// its badges give, in the catalog's order. Worn is the one it wears:
	// the one it chose, while it still holds it, else the score's.
	Titles []string
	Worn   string
}

// BadgeHiddenName is the name a silhouette lists under.
const BadgeHiddenName = "???"

// badgeView is one badge's view with its text shown, for a badge just
// earned. ok is false for a key the ruleset does not have.
func (a *Account) badgeView(book *badgeBook, key string) (BadgeView, bool) {
	def := book.def(key)
	if def == nil {
		return BadgeView{}, false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.viewLocked(book, def, true), true
}

// viewLocked builds a badge's view. shown says whether its text may show.
func (a *Account) viewLocked(book *badgeBook, def *config.BadgeDef, shown bool) BadgeView {
	v := BadgeView{
		Key:       def.Key,
		Family:    def.Family,
		Name:      BadgeHiddenName,
		Level:     def.Tier,
		Tier:      def.Tier.Name(),
		Rarity:    def.RarityName(),
		Points:    def.Points(),
		Secret:    def.Reveal.Kind == config.BadgeSecret,
		Integrity: def.Integrity(),
		Hidden:    !shown,
	}
	if e, ok := a.Badges[def.Key]; ok {
		v.Earned = true
		if e.At > 0 {
			v.At = time.Unix(e.At, 0)
		}
		v.Run = e.Run
		v.Crossed = e.Flags&BadgeFlagCrossed != 0
	}
	if !shown {
		// A silhouette: the tier shows (the frame is the tier), the text
		// does not. A secret badge gives its one-line hint.
		if v.Secret {
			v.Desc = def.Hint
		}
		return v
	}
	v.Name, v.Desc = def.Name, def.Desc
	v.Emblem = def.Emblem
	v.RewardTheme, v.RewardTitle = def.Reward.Theme, def.Reward.Title
	if !v.Earned && def.Scope == config.BadgeLifetime && def.Counter != "" && def.Threshold > 0 {
		v.Progress, v.Target = a.Counters[def.Counter], def.Threshold
	}
	if def.Ladder != "" {
		v.Ladder = def.Ladder
		pos := book.rung[def.Key]
		v.Rung, v.Rungs = pos[0], pos[1]
	}
	return v
}

// revealedLocked reports whether a locked badge's text may show.
func (a *Account) revealedLocked(def *config.BadgeDef, sight AgeSight) bool {
	switch def.Reveal.Kind {
	case config.BadgeVisible:
		return true
	case config.BadgeRevealAtAge:
		return sight.Reached(def.Reveal.Key)
	case config.BadgeRevealNextAge:
		return sight.SeenNext(def.Reveal.Key)
	case config.BadgeRevealOnCounter:
		return a.Counters[def.Reveal.Key] > 0
	case config.BadgeRevealOnBadge:
		return a.earnedLocked(def.Reveal.Key)
	}
	return false // secret, or a rule this version does not know
}

// badgeListCache is the last badge list badgeViews built and what it was
// built from: the account's badge revision, the ruleset's badges and the
// player's sight of the ages.
type badgeListCache struct {
	ok    bool
	rev   uint64
	book  *badgeBook
	sight AgeSight
	views []BadgeView
	sum   BadgeSummary
}

// badgeViews lists every badge of the ruleset as the player may see it, in
// catalog order, with the totals. An integrity badge is listed only once
// earned.
//
// The list is kept until the account's badges or counters change or the
// player sees further, so a snapshot that changed none of them costs
// nothing. The slice is shared between the callers that got it: read it,
// never write to it.
func (a *Account) badgeViews(book *badgeBook, sight AgeSight) ([]BadgeView, BadgeSummary) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if c := &a.badgeList; c.ok && c.rev == a.badgeRev && c.book == book && c.sight == sight {
		return c.views, c.sum
	}
	views, sum := a.buildBadgeViewsLocked(book, sight)
	a.badgeList = badgeListCache{ok: true, rev: a.badgeRev, book: book, sight: sight, views: views, sum: sum}
	return views, sum
}

// buildBadgeViewsLocked builds the list badgeViews keeps.
func (a *Account) buildBadgeViewsLocked(book *badgeBook, sight AgeSight) ([]BadgeView, BadgeSummary) {
	var sum BadgeSummary
	sum.HiddenCounted = sight.ReachedLast()
	out := make([]BadgeView, 0, len(book.defs))
	clean := 0
	for i := range book.defs {
		def := &book.defs[i]
		earned := a.earnedLocked(def.Key)
		if def.Integrity() {
			if earned {
				out = append(out, a.viewLocked(book, def, true))
			}
			continue
		}
		v := a.viewLocked(book, def, earned || a.revealedLocked(def, sight))
		out = append(out, v)
		switch {
		case v.Earned:
			sum.Earned++
			sum.Shown++
			if !v.Crossed {
				sum.Points += v.Points
				clean++
			}
		case v.Hidden:
			sum.Hidden++
		default:
			sum.Shown++
		}
	}
	countable := sum.Shown + sum.Hidden
	sum.Title, sum.TitleRank, sum.NextTitle, sum.NextTitleAt = book.set.BadgeTitle(sum.Points, countable > 0 && clean == countable)
	sum.Titles = append([]string{sum.Title}, a.badgeTitlesLocked(book)...)
	sum.Worn = sum.Title
	if want := a.settingsLocked().Title; want != "" && slices.Contains(sum.Titles, want) {
		sum.Worn = want
	}
	return out, sum
}

// badgeTitlesLocked is the titles the account's badges give it: a title is
// held with its badge, or with every badge of its set, and a crossed badge
// gives none.
func (a *Account) badgeTitlesLocked(book *badgeBook) []string {
	clean := func(key string) bool {
		e, ok := a.Badges[key]
		return ok && e.Flags&BadgeFlagCrossed == 0
	}
	var out []string
	for _, t := range book.set.BadgeWornTitles() {
		held := t.Badge != "" && clean(t.Badge)
		if t.Set != "" {
			keys := book.setOf[t.Set]
			held = len(keys) > 0
			for _, k := range keys {
				held = held && clean(k)
			}
		}
		if held {
			out = append(out, t.Title)
		}
	}
	return out
}

// countableBadges is how many of the account's badges count toward
// completion under book: earned, known to the ruleset and not integrity
// badges. With no book it is every badge the file holds, or its old
// achievements when those are more (a file no engine has brought up yet).
func (a *Account) countableBadges(book *badgeBook) int {
	if book == nil {
		return max(len(a.Badges), len(a.Achievements))
	}
	n := 0
	for key := range a.Badges {
		if def := book.def(key); def != nil && !def.Integrity() {
			n++
		}
	}
	return n
}

// EarnedBadges returns the keys of the badges the account holds, sorted.
func (a *Account) EarnedBadges() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return sortedKeys(a.Badges)
}
