package ui

import (
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// look takes one look at st, for a book that has learned the given keys.
func look(b *lessonBook, st game.GameState, open string, typed int) []string {
	return lessonStep(b, &lessonCtx{st: st, open: open, typed: typed})
}

func newBook(learned ...string) *lessonBook {
	b := &lessonBook{learned: map[string]bool{}}
	for _, k := range learned {
		b.learned[k] = true
	}
	return b
}

func TestLessonTableIsTwentyOne(t *testing.T) {
	seen := map[string]bool{}
	for _, l := range lessonTable {
		if seen[l.key] || l.when == nil || (l.done == nil) == (l.measure == nil) {
			t.Errorf("lesson %q is repeated or lacks a test", l.key)
		}
		seen[l.key] = true
	}
	if len(lessonTable) != 21 {
		t.Errorf("%d lessons, want 21", len(lessonTable))
	}
}

func TestLessonGatherAndTyped(t *testing.T) {
	st := quiet()
	b := newBook()
	look(b, st, "", 0)
	if b.now != "gather" {
		t.Fatalf("a new town shows %q, want gather", b.now)
	}
	st.Stats.TotalGathered = map[string]float64{"food": 1}
	look(b, st, "", 0)
	if !b.learned["gather"] || b.now == "gather" {
		t.Errorf("gathering did not learn it: learned %v, now %q", b.learned, b.now)
	}
	// Command line waits for the first building, then is learned on a typed line.
	look(b, st, "", 0)
	if b.now == "command_line" {
		t.Error("command_line shows before a building is learned")
	}
	b.learned["first_building"] = true
	look(b, st, "", 0)
	if b.now != "command_line" {
		t.Fatalf("now %q, want command_line", b.now)
	}
	look(b, st, "", 1)
	if !b.learned["command_line"] {
		t.Error("a typed line did not learn command_line")
	}
}

func TestLessonRisesFromTheValueWhenShown(t *testing.T) {
	st := quiet()
	st.Workers.TotalPop, st.Workers.MaxPop = 5, 5
	b := newBook("gather", "first_building", "storage")
	look(b, st, "", 0)
	if b.now != "homes" {
		t.Fatalf("now %q, want homes", b.now)
	}
	look(b, st, "", 0)
	if b.learned["homes"] {
		t.Error("homes learned with the housing where it was")
	}
	st.Workers.MaxPop = 9
	look(b, st, "", 0)
	if !b.learned["homes"] {
		t.Error("homes not learned when the housing rose")
	}
}

func TestLessonAlreadyDoneIsSavedUnseen(t *testing.T) {
	st := quiet()
	st.Stats.TotalGathered = map[string]float64{"wood": 4}
	b := newBook()
	newly := look(b, st, "", 0)
	if !slices.Contains(newly, "gather") || b.now == "gather" {
		t.Errorf("gather newly %v, now %q: want it saved without showing", newly, b.now)
	}
	if _, shown := b.seen["gather"]; shown {
		t.Error("a lesson that was done was marked shown")
	}
}

func TestLessonLaterAndOrder(t *testing.T) {
	st := quiet()
	b := newBook()
	look(b, st, "", 0)
	if b.now != "gather" {
		t.Fatalf("now %q", b.now)
	}
	b.later["gather"] = true
	look(b, st, "", 0)
	if b.now == "gather" {
		t.Error("a lesson put off was shown again")
	}
	// A panel's lesson shows only in its panel, and main lessons not there.
	look(b, st, "plan", 0)
	if b.now != "plan" {
		t.Errorf("in Plan now %q, want plan", b.now)
	}
	look(b, st, "", 0)
	if b.now == "plan" {
		t.Error("the plan lesson shows on the main view")
	}
}

func TestLessonDoneTests(t *testing.T) {
	st := quiet()
	st.Festival.Ready, st.Festival.Cost, st.Festival.Culture = true, 10, 10
	b := newBook("gather", "first_building", "storage", "homes", "jobs", "research", "command_line", "panels")
	look(b, st, "", 0)
	if b.now != "festivals" {
		t.Fatalf("now %q, want festivals", b.now)
	}
	st.Festival.CooldownLeft = 5
	look(b, st, "", 0)
	if !b.learned["festivals"] {
		t.Error("a festival held did not learn the lesson")
	}
}

func TestBandDropsClauseThenCutsSentence(t *testing.T) {
	l, _ := lessonOf("gather")
	wide := untag(bandCells(l, 160, true))
	for _, want := range []string{"NEW", "Gather", l.text.String(), l.clause.String(), "Show me", "Later"} {
		if !strings.Contains(wide, want) {
			t.Errorf("wide band lacks %q: %q", want, wide)
		}
	}
	if runeLen(wide) != 160 {
		t.Errorf("band is %d cells, want 160", runeLen(wide))
	}
	mid := untag(bandCells(l, 70, true))
	if strings.Contains(mid, l.clause.String()) || !strings.Contains(mid, l.text.String()) {
		t.Errorf("clause should drop first: %q", mid)
	}
	narrow := untag(bandCells(l, 44, true))
	if strings.Contains(narrow, l.text.String()) || !strings.Contains(narrow, "Later") || runeLen(narrow) > 44 {
		t.Errorf("sentence should be cut: %q", narrow)
	}
	if strings.Contains(untag(bandCells(l, 160, false)), "Show me") {
		t.Error("Show me drawn with no target")
	}
}

func TestHelpLessonsSection(t *testing.T) {
	none := untag(helpProvider(nil, game.GameState{}, 120))
	if !strings.HasPrefix(none, "═══ Lessons") || !strings.Contains(none, "21 more to come") {
		t.Errorf("Help should open with the lessons and a count: %.80q", none)
	}
	got := untag(helpProvider([]string{"gather"}, game.GameState{}, 120))
	l, _ := lessonOf("gather")
	if !strings.Contains(got, l.text.String()) || !strings.Contains(got, "20 more to come") {
		t.Error("a learned lesson should show in full and the rest as a count")
	}
	if strings.Contains(got, "New here?") {
		t.Error("the onboarding hint is still in Help")
	}
}

func TestLessonsRoundTripAndBand(t *testing.T) {
	d, st, acct := accountDash(t)
	d.lessons = lessonBook{}
	d.refresh()
	st.paint()
	if d.lessons.now == "" {
		t.Skipf("no lesson shows in this town (now %q)", d.lessons.now)
	}
	now := d.lessons.now
	if _, _, _, h := d.bandBox.GetRect(); h != 1 || d.bandRows(24) != 0 {
		t.Errorf("band is %d rows tall, and %d at 24 rows: want 1 and 0", h, d.bandRows(24))
	}
	if !strings.Contains(untag(d.bandBox.GetText(false)), "Later") {
		t.Errorf("band row lacks Later: %q", d.bandBox.GetText(false))
	}
	d.bandClick(regionLater)
	if d.lessons.now == now {
		t.Error("Later did not hide the lesson")
	}
	must(t, acct.LearnLessons("gather", "homes", "gather"))
	back, _, err := game.LoadAccount()
	must(t, err)
	got := back.LessonsLearned()
	if c := func(k string) int { return strings.Count(" "+strings.Join(got, " ")+" ", " "+k+" ") }; c("gather") != 1 || c("homes") != 1 {
		t.Errorf("lessons read back as %v", got)
	}
}
