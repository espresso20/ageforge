package game

import (
	"slices"
	"sort"
	"time"

	"github.com/espresso20/ageforge/config"
)

// account_badges.go is the account's side of the badges: what it stores
// (account.json version 2), how an older file becomes one, and how an event
// is judged against it. Like the lifetime-stat hooks, judging is in memory
// only: it takes the account's own lock, touches no file and calls nothing
// on the engine, so the engine may call it under its write lock. The write
// happens later, in FlushIfDirty.

// BadgeEarned is one earned badge as account.json stores it.
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
// or on an account whose file was. It shows struck through and adds no
// points.
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
	if ctx.clean {
		a.countLocked(book, ev.Kind, ev.amount(), ctx, depth)
		if ev.Subject != "" {
			a.countLocked(book, ev.Kind+"."+ev.Subject, ev.amount(), ctx, depth)
		}
	}
	for _, i := range book.byEvent[ev.Kind] {
		def := &book.defs[i]
		if !ctx.clean && !def.Integrity() {
			continue
		}
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
	if def.Subject != "" && def.Subject != ev.Subject {
		return false
	}
	if def.InAge != "" && def.InAge != ctx.age {
		return false
	}
	if def.Scope == config.BadgeRun && def.Counter != "" {
		if v, _ := a.factLocked(def.Counter, ctx); v < def.Threshold {
			return false
		}
	}
	for _, c := range def.When {
		v, known := a.factLocked(c.Fact, ctx)
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

// factLocked reads a fact by name: a lifetime counter from the account,
// anything else from the run.
func (a *Account) factLocked(name string, ctx badgeCtx) (float64, bool) {
	if len(name) > len(factLife) && name[:len(factLife)] == factLife {
		return a.Counters[name[len(factLife):]], true
	}
	if ctx.fact == nil {
		return 0, false
	}
	return ctx.fact(name)
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
		e.At, e.Run = time.Now().Unix(), ctx.run
	}
	if ctx.crossed || a.Tampered {
		e.Flags |= BadgeFlagCrossed
	}
	a.Badges[def.Key] = e
	a.dirty = true
	if !silent {
		a.pendingEarned = append(a.pendingEarned, def.Key)
	}
	if def.Integrity() || depth >= maxBadgeChain {
		return
	}
	next := ctx
	next.clean = true
	if silent {
		// A badge the record proved leads only to others it proves.
		a.countSilentLocked(book, config.BadgeEvBadge, depth)
		a.countSilentLocked(book, config.BadgeEvBadge+"."+def.Family, depth)
		return
	}
	a.judgeLocked(book, Event{Kind: config.BadgeEvBadge, Subject: def.Family}, next, depth+1)
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

// drainEarned returns the keys of the badges earned since the last call.
func (a *Account) drainEarned() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := a.pendingEarned
	a.pendingEarned = nil
	return out
}

// ensureBadges brings the account up to the ruleset's badges and reports
// whether it changed anything. It is what turns a version 1 file into a
// version 2 one, and it is safe to run on every load: a second run changes
// nothing.
//
//   - The four account achievements become their badges (each badge lists
//     the key it had as an alias). The old list stays in the file.
//   - The lifetime stats the account already kept seed the counters that
//     continue them (prestiges).
//   - What the record already proves is granted: the badge of every age up
//     to the highest reached, and every ladder rung a counter has passed.
//
// Everything it grants is silent: undated, and no toast. A badge granted on
// an account flagged as edited is crossed, like any other earned there.
// In memory only; the caller flushes.
func (a *Account) ensureBadges(book *badgeBook) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.ensureBadgesLocked(book)
}

func (a *Account) ensureBadgesLocked(book *badgeBook) bool {
	changed := false
	if a.Version < accountSchemaVersion {
		a.Version = accountSchemaVersion
		changed = true
	}
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
	return a.viewLocked(def, true), true
}

// viewLocked builds a badge's view. shown says whether its text may show.
func (a *Account) viewLocked(def *config.BadgeDef, shown bool) BadgeView {
	v := BadgeView{
		Key:       def.Key,
		Family:    def.Family,
		Name:      BadgeHiddenName,
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
	if !v.Earned && def.Scope == config.BadgeLifetime && def.Counter != "" && def.Threshold > 0 {
		v.Progress, v.Target = a.Counters[def.Counter], def.Threshold
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
	}
	return false // secret, or a rule this version does not know
}

// badgeViews lists every badge of the ruleset as the player may see it, in
// catalog order, with the totals. An integrity badge is listed only once
// earned.
func (a *Account) badgeViews(book *badgeBook, sight AgeSight) ([]BadgeView, BadgeSummary) {
	a.mu.Lock()
	defer a.mu.Unlock()
	var sum BadgeSummary
	sum.HiddenCounted = sight.ReachedLast()
	out := make([]BadgeView, 0, len(book.defs))
	for i := range book.defs {
		def := &book.defs[i]
		earned := a.earnedLocked(def.Key)
		if def.Integrity() {
			if earned {
				out = append(out, a.viewLocked(def, true))
			}
			continue
		}
		v := a.viewLocked(def, earned || a.revealedLocked(def, sight))
		out = append(out, v)
		switch {
		case v.Earned:
			sum.Earned++
			sum.Shown++
			if !v.Crossed {
				sum.Points += v.Points
			}
		case v.Hidden:
			sum.Hidden++
		default:
			sum.Shown++
		}
	}
	return out, sum
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
