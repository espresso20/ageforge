package game

import (
	"slices"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/platform/steam"
)

// tally.go keeps the lifetime tallies: each is one coin, a single record in
// the collection (level, when, count) that is only ever replaced by a finer one.

// peakCounter is the counter that holds a domain tally's best count.
func peakCounter(t config.TallyDef) string { return "peak." + t.Key }

// tallyCounters is every account counter a tally reads.
func tallyCounters(t config.TallyDef) []string {
	switch {
	case t.Peak:
		return []string{peakCounter(t)}
	case len(t.Subjects) > 0 && t.Counter == "":
		return t.Subjects
	case len(t.Subjects) > 0:
		out := make([]string, len(t.Subjects))
		for i, s := range t.Subjects {
			out[i] = t.Counter + "." + s
		}
		return out
	}
	return []string{t.Counter}
}

// tallyValue is what the account's counters say a tally stands at.
func (a *Account) tallyValue(t config.TallyDef) float64 {
	if len(t.Subjects) == 0 {
		return a.Counters[tallyCounters(t)[0]]
	}
	n := 0.0
	for i, c := range tallyCounters(t) {
		if a.Counters[c] >= t.MinOf(i) {
			n++
		}
	}
	return n
}

// TallyCoin is the coin of a tally: its level (0 when not struck), when, and
// the count it was struck at.
func (a *Account) TallyCoin(key string) (level int, at int64, figure int) {
	if s := a.CoinStackOf(config.CoinTally, key); len(s.Copies) > 0 {
		return s.Copies[0].Level, s.Copies[0].At, s.Copies[0].Figure
	}
	return
}

// tallyLocked strikes a tally's coin if its count has reached a finer level.
// A silent strike is a catch-up from the record: it is not dated and gives
// no event.
func (a *Account) tallyLocked(book *badgeBook, t config.TallyDef, ctx badgeCtx, silent bool, depth int) {
	v := a.tallyValue(t)
	lvl := t.Level(v)
	if lvl == 0 {
		return
	}
	c := config.CoinTally
	cp := CoinCopy{Type: c, Age: t.Key, Level: lvl, Figure: int(v)}
	if !silent {
		cp.At = badgeClock().Unix()
	}
	if ctx.crossed || a.Tampered || a.BadgesTampered {
		cp.Flags = BadgeFlagCrossed
	}
	if _, struck := a.strikeCoinLocked(cp); !struck {
		return
	}
	a.badgeRev++
	rung := 0
	for _, r := range t.Rungs {
		if v >= r {
			rung++
		}
	}
	steam.Unlock(t.Key + "." + strconv.Itoa(rung))
	for _, th := range book.set.BadgeThemes() {
		if k, l, ok := book.tallies.Gift(th.Badge); ok && k == t.Key && lvl >= l && !slices.Contains(a.Unlocks.Themes, th.Key) {
			a.Unlocks.Themes = append(a.Unlocks.Themes, th.Key)
		}
	}
	if lvl == config.CoinLevels && !silent && depth < maxBadgeChain {
		top := 0.0
		if strings.HasPrefix(t.Key, "lineage.") {
			top = 1
		}
		a.judgeLocked(book, Event{Kind: config.BadgeEvTallyTop, Subject: t.Key, Attrs: map[string]float64{"lineage": top}}, ctx, depth+1)
	}
}

// tallyCountLocked re-reads the tallies that count on a counter.
func (a *Account) tallyCountLocked(book *badgeBook, counter string, ctx badgeCtx, depth int) {
	for _, i := range book.tallyOf[counter] {
		a.tallyLocked(book, book.tallies[i], ctx, false, depth)
	}
}

// peaksLocked raises the domain tallies' best counts from a census.
func (a *Account) peaksLocked(book *badgeBook, ev Event, ctx badgeCtx, depth int) {
	for _, t := range book.tallies {
		if !t.Peak {
			continue
		}
		name := peakCounter(t)
		if v := ev.Attrs[factStaffed+strings.TrimPrefix(t.Counter, factStaffed)]; v > a.Counters[name] {
			if a.Counters == nil {
				a.Counters = map[string]float64{}
			}
			a.Counters[name] = v
			a.dirty = true
			a.tallyLocked(book, t, ctx, false, depth)
		}
	}
}

// tallyLevelOf is a tally's level, for the titles it gives, 0 when its coin is crossed. The lock is held.
func (a *Account) tallyLevelOf(key string) int {
	if s, ok := a.coins[coinKey(config.CoinTally, key)]; ok && len(s.Copies) > 0 && s.Copies[0].Flags&BadgeFlagCrossed == 0 {
		return s.Copies[0].Level
	}
	return 0
}
