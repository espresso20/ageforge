package flavor

import (
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/vibe"
)

// The stories under VibeMode (vibe.go). A script that marks every text it is
// asked for stands in for the real one: taking the marks out gives the plain
// line back, so any change to the draw, the slots or the words around them
// shows.

const (
	markOpen  = "⟦"
	markClose = "⟧"
)

var unmark = strings.NewReplacer(markOpen, "", markClose, "")

func marked(id, plain string) string { return markOpen + plain + markClose }

// storyRequests are the requests the game builds, for every Moment at the ages
// the game plays them in (every age for most; the harbinger's own in short).
func storyRequests() []Request {
	var out []Request
	for _, m := range Moments() {
		for _, age := range config.AgeOrder() {
			out = append(out, gameRequests(m, age)...)
		}
	}
	return out
}

// TestEveryStoryTextHasARowByItsPlaceInTheCatalog: each sentence is
// flavor.<catalog>.<pool>.<n>, each word-list entry flavor.words.<list>.<n>,
// and none is given twice.
func TestEveryStoryTextHasARowByItsPlaceInTheCatalog(t *testing.T) {
	shape := regexp.MustCompile(`^flavor\.(words\.)?[a-z_]+(\.[a-z_]+)?\.[0-9]{3}$`)
	seen := map[string]bool{}
	sentences := 0
	for _, m := range Moments() {
		for _, tp := range templatesFor(m) {
			if !shape.MatchString(tp.Row) {
				t.Fatalf("template %s has the row %q", tp.ID, tp.Row)
			}
			if seen[tp.Row] {
				t.Fatalf("the row %s is given twice (template %s)", tp.Row, tp.ID)
			}
			seen[tp.Row] = true
			sentences++
		}
	}
	if sentences < 3700 {
		t.Errorf("%d sentences have a row, want at least 3,700", sentences)
	}
	words := 0
	for _, r := range VibeRows() {
		if !strings.HasPrefix(r.ID, "flavor.words.") {
			continue
		}
		if !shape.MatchString(r.ID) || seen[r.ID] {
			t.Fatalf("the word-list row %q is badly formed or given twice", r.ID)
		}
		seen[r.ID] = true
		words++
	}
	if words < 200 {
		t.Errorf("%d word-list entries have a row, want at least 200", words)
	}
	if got := len(VibeRows()); got != sentences+words {
		t.Errorf("VibeRows lists %d texts, the catalogs have %d", got, sentences+words)
	}

	// Spot checks, by hand, of the three shapes of pool name.
	byID := templateByID()
	for id, row := range map[string]string{
		"exp_success_any_0":   "flavor.expedition_success.any.001",
		"enc_cap_late_11":     "flavor.encounter_full.late.012",
		"run_end_plain_2":     "flavor.run_ending.plain.003",
		"harb_warn_anytier_0": "flavor.harbinger_warning.anytier.001",
	} {
		tp, ok := byID[id]
		if !ok {
			t.Errorf("no template %s", id)
		} else if tp.Row != row {
			t.Errorf("template %s has the row %s, want %s", id, tp.Row, row)
		}
	}
	// A batch of sentences for particular ages is a pool named by its first age.
	n := 0
	for _, m := range []Moment{HarbingerArrival, HarbingerWarning, HarbingerBraced, RunEnding} {
		for _, tp := range templatesFor(m) {
			rest, ok := strings.CutPrefix(tp.ID, map[Moment]string{
				HarbingerArrival: "harb_arrival_age_", HarbingerWarning: "harb_warn_age_", HarbingerBraced: "harb_braced_age_", RunEnding: "run_end_age_",
			}[m])
			if !ok {
				continue
			}
			n++
			if !strings.Contains(tp.Row, "."+tp.Ages[0]+".") || !strings.HasPrefix(rest, tp.Ages[0]+"_") {
				t.Fatalf("template %s (ages %v) has the row %s", tp.ID, tp.Ages, tp.Row)
			}
		}
	}
	if n < 100 {
		t.Errorf("only %d sentences of a per-age batch were checked", n)
	}
}

// TestStoriesAreThePlainCatalogWhileVibeModeIsOff: with the switch off, a
// script that rewrites every text does not reach a line.
func TestStoriesAreThePlainCatalogWhileVibeModeIsOff(t *testing.T) {
	vibe.Set(false)
	reqs := storyRequests()
	draw := func() []Result {
		out := make([]Result, 0, len(reqs))
		rng := rand.New(rand.NewSource(5))
		st := NewStream()
		for _, r := range reqs {
			out = append(out, st.Generate(r, rng))
		}
		return out
	}
	want := draw()
	restore := vibe.UseRewriterForTest(marked)
	got := draw()
	restore()
	if !reflect.DeepEqual(got, want) {
		t.Fatal("a script that rewrites every text changed a line with VibeMode off")
	}
	told := 0
	for _, r := range want {
		if r.Text != r.Plain {
			t.Fatalf("with VibeMode off the text %q is not the plain %q", r.Text, r.Plain)
		}
		if strings.Contains(r.Text, markOpen) || strings.Contains(r.Text, markClose) {
			t.Fatalf("a plain line has a mark in it: %q", r.Text)
		}
		if r.Text != "" {
			told++
		}
	}
	if told < len(reqs)/2 {
		t.Errorf("only %d of %d requests drew a line", told, len(reqs))
	}
}

// TestStoriesFollowTheScriptAndTheDrawDoesNot: with the switch on, every
// sentence and every word of a slot is the script's, and the draw is the draw
// it is with the switch off: the same sentence, the same word, the same number
// of draws from the generator, and the same memory in the Stream.
func TestStoriesFollowTheScriptAndTheDrawDoesNot(t *testing.T) {
	reqs := storyRequests()
	type outcome struct {
		results []Result
		next    int64
		state   *StreamState
	}
	draw := func(on bool) outcome {
		if on {
			defer vibe.UseRewriterForTest(marked)()
			vibe.Set(true)
			defer vibe.Set(false)
		}
		rng := rand.New(rand.NewSource(9))
		st := NewStream()
		var out outcome
		for _, r := range reqs {
			out.results = append(out.results, st.Generate(r, rng))
		}
		out.next, out.state = rng.Int63(), st.State()
		return out
	}
	plain, told := draw(false), draw(true)
	if told.next != plain.next {
		t.Error("the script changed how many draws the generator made")
	}
	if !reflect.DeepEqual(told.state, plain.state) {
		t.Error("the script changed what the Stream remembers")
	}
	slotted, lines := 0, 0
	for i, p := range plain.results {
		o := told.results[i]
		if o.Template != p.Template || o.Key() != p.Key() || o.slot != p.slot {
			t.Fatalf("request %d: the script changed the draw: %s %q vs %s %q", i, o.Template, o.Key(), p.Template, p.Key())
		}
		if o.Plain != p.Text {
			t.Fatalf("request %d: Plain reads %q with VibeMode on and %q with it off", i, o.Plain, p.Text)
		}
		if p.Text == "" {
			if o.Text != "" {
				t.Fatalf("request %d: the script gave a line to a request with none", i)
			}
			continue
		}
		lines++
		if got := unmark.Replace(o.Text); got != p.Text {
			t.Fatalf("request %d: the line under the script, marks out, reads\n  %q\nin plain English\n  %q", i, got, p.Text)
		}
		want := 1
		if p.slot != "" {
			want = 2 // the sentence and the word that fills its slot
			slotted++
		}
		// A frame that names a resource ("your ... stores") is the script's too.
		if n := strings.Count(o.Text, markOpen); n < want {
			t.Fatalf("request %d: %d of the line's texts are the script's, want %d: %q", i, n, want, o.Text)
		}
	}
	if lines < 1000 || slotted < 100 {
		t.Errorf("only %d lines drawn, %d of them with a slot", lines, slotted)
	}
}

// TestALineOfTheScriptTakesTheSlotItHas: the script's line for a sentence is
// cut at its slot and the drawn word goes in (the word the script has a line
// for, else its plain one). A line that cannot take the draw is not used.
func TestALineOfTheScriptTakesTheSlotItHas(t *testing.T) {
	var slotted, bare tmpl
	for _, tp := range templatesFor(ExpeditionFailure) {
		switch {
		case slotted.ID == "" && len(tp.Parts) == 3 && !strings.Contains(tp.sentence(), "{"):
			slotted = tp
		case bare.ID == "" && len(tp.Parts) == 1 && !strings.Contains(tp.sentence(), "{"):
			bare = tp
		}
	}
	if slotted.ID == "" || bare.ID == "" {
		t.Fatal("no sentence with a slot, or none without one")
	}
	bank := slotted.Parts[1].bank
	plainWord := banks[bank][3]
	vibed := func(script map[string]string, tp tmpl, draws []int) string {
		defer vibe.UseScriptForTest(script)()
		vibe.Set(true)
		defer vibe.Set(false)
		return tp.compose(draws, true)
	}
	plainSentence := slotted.compose([]int{3}, false)

	// The sentence and the word are both the script's.
	got := vibed(map[string]string{slotted.Row: "Yo ~ lol", wordRow(bank, 3): "the vibes"}, slotted, []int{3})
	if got != "Yo the vibes lol." {
		t.Errorf("a line with its slot reads %q", got)
	}
	// The word has no line: its plain word goes into the script's sentence.
	got = vibed(map[string]string{slotted.Row: "Yo ~ lol"}, slotted, []int{3})
	if got != "Yo "+plainWord+" lol." {
		t.Errorf("a sentence told with a plain word reads %q", got)
	}
	// The sentence has no line: the script's word goes into the plain sentence.
	got = vibed(map[string]string{wordRow(bank, 3): "the vibes"}, slotted, []int{3})
	head, tail, _ := strings.Cut(slotted.sentence(), slotMark)
	if got != head+"the vibes"+tail+"." {
		t.Errorf("a plain sentence told with a word of the script's reads %q", got)
	}
	// A line with no slot, or with two, cannot take the draw: the sentence is plain.
	for _, line := range []string{"Yo lol", "Yo ~ ~ lol"} {
		if got := vibed(map[string]string{slotted.Row: line}, slotted, []int{3}); got != plainSentence {
			t.Errorf("a line %q that cannot take the slot reads %q, want the plain sentence", line, got)
		}
	}
	// A sentence with no slot is told whole, and a line that has one is not used.
	if got := vibed(map[string]string{bare.Row: "Bruh"}, bare, nil); got != "Bruh." {
		t.Errorf("a sentence with no slot reads %q", got)
	}
	if got := vibed(map[string]string{bare.Row: "Bruh ~"}, bare, nil); got != bare.compose(nil, false) {
		t.Errorf("a line with a slot for a sentence with none reads %q", got)
	}
	// With the switch off, none of it is used.
	defer vibe.UseScriptForTest(map[string]string{slotted.Row: "Yo ~ lol", wordRow(bank, 3): "the vibes"})()
	vibe.Set(false)
	if got := slotted.compose([]int{3}, false); got != plainSentence {
		t.Errorf("with VibeMode off the plain reading is %q", got)
	}
}

// TestALineIsToldAgainFromItsKey: the key of a line names the sentence and the
// word drawn for it, and tells the same line again, in the words of whichever
// script is in force, without a draw. A key that names nothing is refused.
func TestALineIsToldAgainFromItsKey(t *testing.T) {
	reqs := storyRequests()
	rng := rand.New(rand.NewSource(3))
	st := NewStream()
	n := 0
	for i, req := range reqs {
		if i%7 != 0 {
			continue
		}
		res := st.Generate(req, rng)
		if res.Text == "" {
			if res.Key() != "" {
				t.Fatalf("an empty line has the key %q", res.Key())
			}
			continue
		}
		n++
		key := res.Key()
		for _, told := range []bool{false, true} {
			if got, ok := Retell(key, req, told); !ok || got != res.Text {
				t.Fatalf("%s: told again with VibeMode off (told %v) it reads %q (%v), want %q", key, told, got, ok, res.Text)
			}
		}
		func() {
			defer vibe.UseRewriterForTest(marked)()
			vibe.Set(true)
			defer vibe.Set(false)
			got, ok := Retell(key, req, true)
			if !ok || unmark.Replace(got) != res.Text || strings.Count(got, markOpen) < 1+len(res.draws) {
				t.Fatalf("%s: told again under the script it reads %q (%v), plain %q", key, got, ok, res.Text)
			}
			if plain, ok := Retell(key, req, false); !ok || plain != res.Text {
				t.Fatalf("%s: told again in plain English under the script it reads %q (%v), want %q", key, plain, ok, res.Text)
			}
			// What Generate told at the time is what the key tells now.
			told := NewStream()
			r2 := rand.New(rand.NewSource(int64(i)))
			again := told.Generate(req, r2)
			if again.Text == "" {
				return
			}
			if got, ok := Retell(again.Key(), req, true); !ok || got != again.Text {
				t.Fatalf("%s: told again it reads %q (%v), told at the time %q", again.Key(), got, ok, again.Text)
			}
		}()
	}
	if n < 300 {
		t.Errorf("only %d lines were told again", n)
	}
	var withSlot, noSlot string
	for _, tp := range templatesFor(ExpeditionSuccess) {
		switch {
		case withSlot == "" && len(tp.Parts) == 3:
			withSlot = tp.ID
		case noSlot == "" && len(tp.Parts) == 1:
			noSlot = tp.ID
		}
	}
	for _, key := range []string{
		"", "nonesuch_0", ":0",
		withSlot,            // a slot and no word for it
		withSlot + ":x",     // not a number
		withSlot + ":-1",    // before the first word
		withSlot + ":99999", // past the last
		withSlot + ":0,0",   // a word too many
		noSlot + ":0",       // a word for a sentence with no slot
	} {
		if got, ok := Retell(key, Request{}, true); ok {
			t.Errorf("the key %q was told again as %q", key, got)
		}
	}
	for _, key := range []string{noSlot, withSlot + ":0"} {
		if got, ok := Retell(key, Request{}, true); !ok || got == "" {
			t.Errorf("the key %q was refused", key)
		}
	}
}

// TestTheFramesThatNameAResourceAreTheScriptsToo: "your ... stores", "a haul
// of ..." and the singular of a count noun are words of the game's own, told in
// the script's when the line is, and plain in the Plain reading whatever the
// switch says.
func TestTheFramesThatNameAResourceAreTheScriptsToo(t *testing.T) {
	defer vibe.UseScriptForTest(map[string]string{
		"flavor.resource.phraseStores.01":  "yo ",
		"flavor.resource.phraseStores.02":  " pile",
		"flavor.resource.phraseHaul.01":    "a whole lotta ",
		"flavor.resource.countSingular.01": "bot",
	})()
	vibe.Set(true)
	defer vibe.Set(false)
	req := Request{Resource: "food", Amount: 1}
	if got := fillAs("{res_stores}; {res_haul}", req, true); got != "yo food pile; a whole lotta food" {
		t.Errorf("the frames under the script read %q", got)
	}
	if got := fillAs("{res_stores}; {res_haul}", req, false); got != "your food stores; a haul of food" {
		t.Errorf("the frames in plain English, with VibeMode on, read %q", got)
	}
	nano := Request{Resource: "nanobots", Amount: 1}
	if got := fillAs("{amt_res}", nano, true); got != "1 bot" {
		t.Errorf("the singular of a count noun under the script reads %q", got)
	}
	if got := fillAs("{amt_res}", nano, false); got != "1 nanobot" {
		t.Errorf("the singular of a count noun in plain English reads %q", got)
	}
	vibe.Set(false)
	if got := phraseStoresAs("food", false); got != "your food stores" {
		t.Errorf("phraseStoresAs reads %q with VibeMode off", got)
	}
}

// TestAResourceNamedInAStoryIsTheScriptsName: {res}, {res_stores}, {res_haul}
// and {amt_res} name the resource as the script does while VibeMode is on (the
// row the ruleset's definition reads), lowercased, and as config does
// otherwise, whatever the script holds; a resource the script has no name for
// keeps its own.
func TestAResourceNamedInAStoryIsTheScriptsName(t *testing.T) {
	defer vibe.UseScriptForTest(map[string]string{"resource.food.name": "Snacks"})()
	vibe.Set(true)
	defer vibe.Set(false)
	req := Request{Resource: "food", Amount: 12}
	const text = "{res}; {res_stores}; {res_haul}; {amt_res}"
	if got := fillAs(text, req, true); got != "snacks; your snacks stores; a haul of snacks; 12 snacks" {
		t.Errorf("the frames under the script read %q", got)
	}
	if got := fillAs(text, req, false); got != "food; your food stores; a haul of food; 12 food" {
		t.Errorf("the frames in plain English, with VibeMode on, read %q", got)
	}
	if got := fillAs("{res}", Request{Resource: "gold"}, true); got != "gold" {
		t.Errorf("a resource the script has no name for reads %q", got)
	}
	if got := fillAs("{res}", Request{}, true); got != "supplies" {
		t.Errorf("no resource reads %q", got)
	}
	vibe.Set(false)
	if got := fillAs(text, req, true); got != "food; your food stores; a haul of food; 12 food" {
		t.Errorf("with VibeMode off the frames read %q", got)
	}

	// Under a script that marks every text, the name is marked too, and with
	// the marks out the line is the plain one.
	defer vibe.UseRewriterForTest(marked)()
	vibe.Set(true)
	got := fillAs(text, req, true)
	if unmark.Replace(got) != fillAs(text, req, false) {
		t.Errorf("the frames under the marking script, marks out, read %q", unmark.Replace(got))
	}
	// the name four times, the two frames of {res_stores}, the one of {res_haul}
	if n := strings.Count(got, markOpen); n != 4+2+1 {
		t.Errorf("%d texts of the frames are the script's, want 7: %q", n, got)
	}
	if !strings.Contains(got, markOpen+"food"+markClose) {
		t.Errorf("the resource's name is not marked: %q", got)
	}
}
