package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

func tallyByKey(t *testing.T, key string) config.TallyDef {
	t.Helper()
	for _, td := range rules.Core().Tallies() {
		if td.Key == key {
			return td
		}
	}
	t.Fatalf("no tally %s", key)
	return config.TallyDef{}
}

// A tally's coin rises with its count, and a lower count never takes a level back.
func TestTallyCoinRisesAndNeverFalls(t *testing.T) {
	isolateAccountDir(t)
	a := testAccount(t)
	book := coreBook()
	td := tallyByKey(t, "lineage.housing")
	ctx := badgeCtx{}

	a.mu.Lock()
	a.Counters = map[string]float64{td.Counter: td.Rungs[0]}
	a.tallyLocked(book, td, ctx, false, 0)
	a.mu.Unlock()
	if l, _, f := a.TallyCoin(td.Key); l != 1 || f != int(td.Rungs[0]) {
		t.Fatalf("first rung: level %d figure %d, want 1 and %d", l, f, int(td.Rungs[0]))
	}

	a.mu.Lock()
	a.Counters[td.Counter] = td.Rungs[2]
	a.tallyLocked(book, td, ctx, false, 0)
	a.mu.Unlock()
	want := td.Level(td.Rungs[2])
	if l, _, _ := a.TallyCoin(td.Key); l != want {
		t.Fatalf("third rung: level %d, want %d", l, want)
	}

	a.mu.Lock()
	a.Counters[td.Counter] = td.Rungs[0]
	a.tallyLocked(book, td, ctx, false, 0)
	a.mu.Unlock()
	if l, _, _ := a.TallyCoin(td.Key); l != want {
		t.Errorf("a lower count took the coin to level %d, want it kept at %d", l, want)
	}
	if n := len(a.CoinStackOf(config.CoinTally, td.Key).Copies); n != 1 {
		t.Errorf("the tally has %d records, want one", n)
	}
}

// A domain tally counts the best number staffed at once, taken from a census.
func TestTallyDomainTakesThePeakFromACensus(t *testing.T) {
	isolateAccountDir(t)
	a := testAccount(t)
	book := coreBook()
	td := tallyByKey(t, "domain.food")
	census := func(n float64) {
		a.mu.Lock()
		defer a.mu.Unlock()
		a.peaksLocked(book, Event{Kind: config.BadgeEvCensus, Attrs: map[string]float64{"staffed.food": n}}, badgeCtx{}, 0)
	}
	census(td.Rungs[1])
	census(1)
	if l, _, f := a.TallyCoin(td.Key); l != td.Level(td.Rungs[1]) || f != int(td.Rungs[1]) {
		t.Errorf("peak coin: level %d figure %d, want level %d figure %d", l, f, td.Level(td.Rungs[1]), int(td.Rungs[1]))
	}
}

// A rung reached in play is announced like a badge; a catch-up from the record is not,
// and a coin the record already holds at its finest level gives the badges its strike gives.
func TestTallyRungIsAnnouncedAndCatchUpIsSilent(t *testing.T) {
	isolateAccountDir(t)
	a := testAccount(t)
	book := coreBook()
	td := tallyByKey(t, "lineage.housing")
	a.mu.Lock()
	a.Counters = map[string]float64{td.Counter: td.Rungs[0]}
	a.tallyLocked(book, td, badgeCtx{}, true, 0)
	a.mu.Unlock()
	if got := a.drainTallies(); len(got) != 0 {
		t.Fatalf("a silent strike was announced: %v", got)
	}
	a.mu.Lock()
	a.Counters[td.Counter] = td.Rungs[1]
	a.tallyLocked(book, td, badgeCtx{}, false, 0)
	a.mu.Unlock()
	got := a.drainTallies()
	if len(got) != 1 || got[0] != td.Key {
		t.Fatalf("announced %v, want [%s]", got, td.Key)
	}
	v, ok := a.tallyView(book, td.Key)
	if !ok || !v.Earned || v.Name != td.Name || v.Tier == "" {
		t.Errorf("view %+v ok %v", v, ok)
	}

	a.mu.Lock()
	a.Counters[td.Counter] = td.Rungs[len(td.Rungs)-1]
	a.tallyLocked(book, td, badgeCtx{}, true, 0)
	a.ensureBadgesLocked(book)
	a.mu.Unlock()
	if !a.earnedLocked("special.full_set") {
		t.Error("a lineage coin held at its finest level did not give Full Set on catch-up")
	}
}
