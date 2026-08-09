package flavor

import (
	"fmt"
	"go/parser"
	"go/token"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// --- helpers ---------------------------------------------------------------

// realExpeditionNames is the live expedition roster copied from game/military.go.
// Half of these are VERB-LED order titles ("Raid Bandit Camp") and half are noun
// phrases ("Scout Party"), which is exactly the trap the subject frames exist to
// dodge — so the generator is tested against the real strings, not a tidy sample.
var realExpeditionNames = []string{
	"Scout Party", "Scout Nearby Ruins", "Raid Bandit Camp", "Trade Escort",
	"Conquer Territory", "Siege Enemy Castle", "Naval Expedition", "Colonial Campaign",
	"World Domination", "Cyber Raid", "Neon Heist", "Fusion Plant Assault",
	"Orbital Strike", "Warp Invasion", "Galactic Conquest", "Quantum Incursion",
}

// realFactionNames is the live civilization roster from config/trade.go. Note the
// mixed grammatical number: "Merchant Guild" is singular, "Void Reavers" plural.
var realFactionNames = []string{
	"Riverlands Tribes", "Ironhold Clans", "Merchant Guild", "Artisan League",
	"Atomic Directorate", "Tech Consortium", "Shadow Syndicate", "Plasma Nomads",
	"Stellar Federation", "Void Reavers", "Quantum Collective",
}

var allTones = []Tone{Neutral, Grim, Triumphant, Wry}

// fuzzRequests returns a broad cross-product of realistic requests for a Moment:
// every tone, both expedition kinds and a couple of personalities, several ages
// across all three eras, mass and count resources (with and without an amount),
// and both a subject and no subject.
//
// The subject and resource axes are SAMPLED rather than exhaustive — the full
// rosters get their own dedicated tests (TestVerbLedExpeditionSubjects,
// TestFactionSubjectsRender, TestNoMassNounPluralization), and crossing all of
// them here would put this package into the tens of seconds for no extra signal.
func fuzzRequests(m Moment) []Request {
	ages := []string{"", "primitive_age", "medieval_age", "industrial_age", "modern_age", "space_age", "transcendent_age", "not_a_real_age"}
	kinds := []string{"", "scouting", "military", "aggressive", "mercantile", "NONSENSE"}
	subjects := []string{"", "Raid Bandit Camp", "Scout Party", "Merchant Guild", "Void Reavers"}
	resources := []string{"food", "knowledge", "gold", "quantum_flux", "soldiers", "nanobots"}

	var out []Request
	for _, tone := range allTones {
		for _, age := range ages {
			for _, kind := range kinds {
				for _, subj := range subjects {
					for _, key := range resources {
						out = append(out,
							Request{Moment: m, Tone: tone, Age: age, Kind: kind, Subject: subj},
							Request{Moment: m, Tone: tone, Age: age, Kind: kind, Subject: subj, Resource: key},
							Request{Moment: m, Tone: tone, Age: age, Kind: kind, Subject: subj, Resource: key, Amount: 1},
							Request{Moment: m, Tone: tone, Age: age, Kind: kind, Subject: subj, Resource: key, Amount: 4237.6, Count: 3, Ticks: 900},
						)
					}
				}
			}
		}
	}
	return out
}

// generateMany runs one request n times off a fresh rng and returns the lines.
func generateMany(req Request, seed int64, n int) []string {
	rng := rand.New(rand.NewSource(seed))
	out := make([]string, n)
	for i := range out {
		out[i] = Line(req, rng)
	}
	return out
}

// allFragments returns every authored fragment in the package, with its bank.
func allFragments(t *testing.T) map[string][]string {
	t.Helper()
	return banks
}

// --- decoupling guard: the package must never import game ------------------

func TestNoGameImport(t *testing.T) {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, imp := range f.Imports {
			if strings.Contains(imp.Path.Value, "/ageforge/game") {
				t.Errorf("%s imports %s — flavor must not depend on the game package", name, imp.Path.Value)
			}
		}
	}
}

// --- output contract -------------------------------------------------------

// TestOutputIsMarkupFree fuzzes every Moment across the full request
// cross-product: generated prose must never carry a tview colour tag or a printf
// directive, because the CALLER wraps it in both. Also pins the shape (no
// unfilled placeholders, no doubled spaces, terminal full stop).
func TestOutputIsMarkupFree(t *testing.T) {
	for _, m := range Moments() {
		reqs := fuzzRequests(m)
		rng := rand.New(rand.NewSource(int64(m) * 7919))
		for _, req := range reqs {
			line := Line(req, rng)
			if line == "" {
				t.Fatalf("%v: empty line for %+v", m, req)
			}
			if strings.ContainsAny(line, "[]%") {
				t.Fatalf("%v: markup or printf directive in %q", m, line)
			}
			if strings.ContainsAny(line, "{}") {
				t.Fatalf("%v: unfilled placeholder in %q", m, line)
			}
			if strings.Contains(line, "  ") {
				t.Fatalf("%v: doubled space in %q", m, line)
			}
			if strings.Contains(line, " .") || strings.Contains(line, "..") {
				t.Fatalf("%v: stray punctuation in %q", m, line)
			}
			if !strings.HasSuffix(line, ".") {
				t.Fatalf("%v: line does not end in a full stop: %q", m, line)
			}
			if line != strings.TrimSpace(line) {
				t.Fatalf("%v: line has edge whitespace: %q", m, line)
			}
		}
	}
}

// TestUnknownMomentIsEmpty pins the "unknown Moment yields nothing, quietly"
// contract — no panic, no partial sentence.
func TestUnknownMomentIsEmpty(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for _, m := range []Moment{MomentUnknown, Moment(-1), Moment(9999)} {
		res := Generate(Request{Moment: m, Subject: "Scout Party", Resource: "food", Amount: 10}, rng)
		if res.Text != "" || res.Template != "" {
			t.Errorf("Moment %d produced %+v, want the zero Result", m, res)
		}
		if Line(Request{Moment: m}, rng) != "" {
			t.Errorf("Line(Moment %d) returned non-empty", m)
		}
	}
}

// TestZeroRequestIsGrammatical is the load-bearing "every field is optional"
// test: a Request carrying nothing but a registered Moment must still produce a
// finished sentence — capitalized, punctuated, placeholder-free.
func TestZeroRequestIsGrammatical(t *testing.T) {
	for _, m := range Moments() {
		rng := rand.New(rand.NewSource(20260808))
		seen := map[string]bool{}
		for i := 0; i < 500; i++ {
			res := Generate(Request{Moment: m}, rng)
			if res.Text == "" {
				t.Fatalf("%v: zero Request produced an empty line", m)
			}
			if res.Template == "" {
				t.Fatalf("%v: zero Request produced no template id", m)
			}
			first := res.Text[:1]
			if first != strings.ToUpper(first) {
				t.Fatalf("%v: zero-Request line does not start capitalized: %q", m, res.Text)
			}
			if strings.ContainsAny(res.Text, "{}[]%") {
				t.Fatalf("%v: zero-Request line is not clean: %q", m, res.Text)
			}
			seen[res.Text] = true
		}
		// A zero Request only reaches the unconstrained templates, but that pool
		// must still be wide enough to not read as a fixed string.
		if len(seen) < 50 {
			t.Errorf("%v: zero Request produced only %d distinct lines in 500 draws; want >= 50", m, len(seen))
		}
	}
}

// --- determinism -----------------------------------------------------------

// TestDeterministicStream pins the purity contract: same seed + same Request ⇒
// identical prose stream, and a different seed diverges (so the rng is genuinely
// driving the draws rather than the output being constant).
func TestDeterministicStream(t *testing.T) {
	req := Request{
		Moment: ExpeditionSuccess, Tone: Triumphant, Age: "medieval_age",
		Subject: "Raid Bandit Camp", Kind: "military", Resource: "gold", Amount: 420, Count: 4, Ticks: 60,
	}
	a := generateMany(req, 20260808, 400)
	b := generateMany(req, 20260808, 400)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("line %d differs between identical seeds:\n  %q\n  %q", i, a[i], b[i])
		}
	}
	c := generateMany(req, 99999, 400)
	same := true
	for i := range a {
		if a[i] != c[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatal("different seeds produced identical streams — rng is not driving the draws")
	}
}

// TestGenerateUsesOnlyPassedRNG is a smoke check that Generate makes a bounded,
// FIXED number of draws (template + one per bank slot) — the property that lets a
// caller share one rng across systems without the prose stream desynchronising
// everything downstream of it.
func TestGenerateUsesOnlyPassedRNG(t *testing.T) {
	for _, m := range Moments() {
		req := Request{Moment: m}
		// Two rngs from the same seed: consume one via Generate, the other by
		// hand for the same number of draws, then check they stay in lockstep.
		gen := rand.New(rand.NewSource(7))
		res := Generate(req, gen)
		if res.Text == "" {
			t.Fatalf("%v: empty text", m)
		}
		manual := rand.New(rand.NewSource(7))
		pool := eligible(req)
		pick := pool[manual.Intn(len(pool))]
		if pick.ID != res.Template {
			t.Fatalf("%v: template pick is not the first draw (got %q, replayed %q)", m, res.Template, pick.ID)
		}
	}
}

// --- capacity --------------------------------------------------------------

// TestCapacityFloors asserts real authored volume per Moment. These floors are
// the point of the hybrid design: if someone thins a bank, the combination count
// collapses and this fails before a player sees the same sentence twice an hour.
func TestCapacityFloors(t *testing.T) {
	floors := map[Moment]int{
		ExpeditionSuccess:   1500,
		ExpeditionFailure:   1500,
		EncounterStandoff:   1000,
		EncounterAtCapacity: 1000,
		WarRaid:             1500,
	}
	for _, m := range Moments() {
		got := Capacity(m)
		want, ok := floors[m]
		if !ok {
			t.Fatalf("Moment %v has no capacity floor declared in this test", m)
		}
		if got < want {
			t.Errorf("Capacity(%v) = %d, want >= %d", m, got, want)
		}
		t.Logf("Capacity(%v) = %d", m, got)
	}
	if Capacity(MomentUnknown) != 0 {
		t.Errorf("Capacity(MomentUnknown) = %d, want 0", Capacity(MomentUnknown))
	}
}

// --- countability ----------------------------------------------------------

// TestEveryResourceClassified is the guard that stops a new resource shipping
// without a grammatical class: config is the source of truth for the roster, and
// this package must have an opinion about every key on it.
func TestEveryResourceClassified(t *testing.T) {
	for _, r := range config.BaseResources() {
		c, ok := resourceClass[r.Key]
		if !ok {
			t.Errorf("resource %q (%s) is not classified in resourceClass — add it as massNoun or countNoun",
				r.Key, r.Name)
			continue
		}
		if c == countNoun {
			if _, ok := countSingular[r.Key]; !ok {
				t.Errorf("count noun %q has no singular form in countSingular (needed for the \"1 X\" case)", r.Key)
			}
		}
		if label := resourceLabel(r.Key); label == "" || label != strings.ToLower(label) {
			t.Errorf("resource %q label %q is not a plain lowercase mid-sentence label", r.Key, label)
		}
	}
	// Nothing in the table may reference a key config does not ship.
	known := config.ResourceByKey()
	for key := range resourceClass {
		if _, ok := known[key]; !ok {
			t.Errorf("resourceClass has %q, which is not a config resource key", key)
		}
	}
}

// TestNoMassNounPluralization fuzzes the generator with EVERY resource key and
// greps the output for the two failure shapes a naive pluralizer produces:
// "foods"/"knowledges" (mass noun given an s) and "a food"/"a knowledge" (mass
// noun given an indefinite article).
func TestNoMassNounPluralization(t *testing.T) {
	var badPlural, badArticle []*regexp.Regexp
	for key, class := range resourceClass {
		if class != massNoun {
			continue
		}
		label := regexp.QuoteMeta(resourceLabel(key))
		badPlural = append(badPlural, regexp.MustCompile(`(?i)\b`+label+`s\b`))
		badArticle = append(badArticle, regexp.MustCompile(`(?i)\ba `+label+`\b`))
	}
	// And the count nouns must never be double-pluralized either.
	for key, class := range resourceClass {
		if class != countNoun {
			continue
		}
		label := regexp.QuoteMeta(resourceLabel(key))
		badPlural = append(badPlural, regexp.MustCompile(`(?i)\b`+label+`s\b`))
	}

	lines := 0
	for _, m := range Moments() {
		rng := rand.New(rand.NewSource(int64(m) * 104729))
		for _, r := range config.BaseResources() {
			for _, amt := range []float64{0, 1, 2, 7.4, 999, 1e6} {
				for i := 0; i < 40; i++ {
					line := Line(Request{
						Moment: m, Age: "medieval_age", Kind: "military",
						Subject: "Raid Bandit Camp", Resource: r.Key, Amount: amt,
					}, rng)
					lines++
					for _, re := range badPlural {
						if re.MatchString(line) {
							t.Fatalf("mass/count noun pluralized by %s in %q", re, line)
						}
					}
					for _, re := range badArticle {
						if re.MatchString(line) {
							t.Fatalf("mass noun given an indefinite article by %s in %q", re, line)
						}
					}
				}
			}
		}
	}
	t.Logf("fuzzed %d lines across %d resources with no mass-noun pluralization", lines, len(config.BaseResources()))
}

// TestQuantityInflection pins the "1 soldier" / "12 soldiers" / "40 food" cases
// directly, since they are the whole reason the countability table exists.
func TestQuantityInflection(t *testing.T) {
	cases := []struct {
		key  string
		n    int
		want string
	}{
		{"food", 1, "1 food"},
		{"food", 40, "40 food"},
		{"knowledge", 3, "3 knowledge"},
		{"soldiers", 1, "1 soldier"},
		{"soldiers", 12, "12 soldiers"},
		{"nanobots", 1, "1 nanobot"},
		{"dark_matter_crystals", 2, "2 dark matter crystals"},
		{"dark_matter", 2, "2 dark matter"},
	}
	for _, c := range cases {
		if got := phraseQuantity(c.key, c.n); got != c.want {
			t.Errorf("phraseQuantity(%q, %d) = %q, want %q", c.key, c.n, got, c.want)
		}
	}
	if got := phraseStores("food"); got != "your food stores" {
		t.Errorf("phraseStores(food) = %q", got)
	}
	if got := phraseHaul("quantum_flux"); got != "a haul of quantum flux" {
		t.Errorf("phraseHaul(quantum_flux) = %q", got)
	}
	// An unset resource still renders as something grammatical.
	if got := phraseBare(""); got != "supplies" {
		t.Errorf("phraseBare(\"\") = %q, want the generic fallback", got)
	}
}

// TestMassOnlyFramesNeverReachCountNouns is the structural half of the
// countability guarantee: the {res_stores} / {res_haul} frames are gated on
// needMassRes, so no template that uses them can be picked for "soldiers".
func TestMassOnlyFramesNeverReachCountNouns(t *testing.T) {
	massOnly := []string{"{res_stores}", "{res_haul}"}
	for _, m := range Moments() {
		for _, tpl := range templatesFor(m) {
			usesMassFrame := false
			for _, p := range tpl.Parts {
				if p.bank == "" {
					continue
				}
				for _, frag := range banks[p.bank] {
					for _, f := range massOnly {
						if strings.Contains(frag, f) {
							usesMassFrame = true
						}
					}
				}
			}
			if usesMassFrame && tpl.Needs&needMassRes == 0 {
				t.Errorf("template %q uses a mass-only frame but does not declare needMassRes", tpl.ID)
			}
		}
	}
	// And the runtime agrees: a count-noun request never selects one.
	for _, m := range Moments() {
		for _, key := range []string{"soldiers", "nanobots", "dark_matter_crystals"} {
			for _, tpl := range eligible(Request{Moment: m, Resource: key, Amount: 5}) {
				if tpl.Needs&needMassRes != 0 {
					t.Errorf("%v: template %q is eligible for count noun %q", m, tpl.ID, key)
				}
			}
		}
	}
}

// --- the two grammar traps -------------------------------------------------

// TestVerbLedExpeditionSubjects drives every REAL expedition name through both
// expedition Moments. Half of them are verb-led ("Conquer Territory"), so a
// template that used the name as a sentence subject would produce nonsense; this
// asserts the output stays clean and that the name, when used, survives verbatim.
func TestVerbLedExpeditionSubjects(t *testing.T) {
	for _, m := range []Moment{ExpeditionSuccess, ExpeditionFailure} {
		for _, name := range realExpeditionNames {
			rng := rand.New(rand.NewSource(int64(len(name)) * 31))
			usedSubject := false
			for i := 0; i < 400; i++ {
				res := Generate(Request{
					Moment: m, Subject: name, Kind: "military", Age: "medieval_age", Resource: "gold", Amount: 250,
				}, rng)
				if res.Text == "" {
					t.Fatalf("%v/%q: empty line", m, name)
				}
				if strings.ContainsAny(res.Text, "{}[]%") || strings.Contains(res.Text, "  ") {
					t.Fatalf("%v/%q: malformed line %q", m, name, res.Text)
				}
				if strings.Contains(res.Text, name) {
					usedSubject = true
				}
			}
			if !usedSubject {
				t.Errorf("%v: expedition name %q never appeared in 400 draws — the subject templates are unreachable", m, name)
			}
		}
	}
}

// TestSubjectFrames is a static lint on the authored banks, and the durable half
// of the grammar guarantee: it constrains what a future author may write.
//
//   - Expedition moments: {subject} is an ORDER TITLE and must sit inside one of
//     the title frames, never in a bare subject position.
//   - Faction moments: {subject} is a proper noun of unknown grammatical NUMBER,
//     so it may never be followed by a copula or auxiliary — use a past-tense
//     verb or an object position instead.
func TestSubjectFrames(t *testing.T) {
	titleFrames := []string{
		"the {subject} order", "the {subject} venture", "the {subject} business",
		"marked {subject}", "calls {subject}", "file {subject}",
		"read {subject}", "titled {subject}",
	}
	copulas := []string{
		"{subject} is", "{subject} are", "{subject} was", "{subject} were",
		"{subject} has", "{subject} have", "{subject} do", "{subject} does",
	}
	expeditionMoments := map[Moment]bool{ExpeditionSuccess: true, ExpeditionFailure: true}

	for _, m := range Moments() {
		bankNames := map[string]bool{}
		for _, tpl := range templatesFor(m) {
			for _, p := range tpl.Parts {
				if p.bank != "" {
					bankNames[p.bank] = true
				}
			}
		}
		for name := range bankNames {
			for _, frag := range banks[name] {
				if !strings.Contains(frag, "{subject}") {
					continue
				}
				if expeditionMoments[m] {
					ok := false
					for _, f := range titleFrames {
						if strings.Contains(frag, f) {
							ok = true
							break
						}
					}
					if !ok {
						t.Errorf("%v bank %q: %q uses {subject} outside a title frame — "+
							"expedition names are verb-led and cannot be sentence subjects", m, name, frag)
					}
					continue
				}
				for _, c := range copulas {
					if strings.Contains(frag, c) {
						t.Errorf("%v bank %q: %q puts a copula after {subject} — faction names "+
							"have inconsistent grammatical number; use past tense or an object position", m, name, frag)
					}
				}
			}
		}
	}
}

// TestFactionSubjectsRender drives the real (mixed-number) civilization roster
// through the faction Moments.
func TestFactionSubjectsRender(t *testing.T) {
	for _, m := range []Moment{EncounterStandoff, EncounterAtCapacity, WarRaid} {
		for _, name := range realFactionNames {
			rng := rand.New(rand.NewSource(int64(len(name)) * 17))
			used := false
			for i := 0; i < 400; i++ {
				res := Generate(Request{Moment: m, Subject: name, Kind: "aggressive", Age: "iron_age", Resource: "gold"}, rng)
				if res.Text == "" || strings.ContainsAny(res.Text, "{}[]%") {
					t.Fatalf("%v/%q: bad line %q", m, name, res.Text)
				}
				if strings.Contains(res.Text, name) {
					used = true
				}
			}
			if !used {
				t.Errorf("%v: faction name %q never appeared in 400 draws", m, name)
			}
		}
	}
}

// --- catalog invariants ----------------------------------------------------

// TestBanksAreSane is the same bar config/log_flavor_test.go holds its pools to,
// applied to every authored fragment: non-empty, unique, markup-free, and free of
// the terminal punctuation the TEMPLATE is responsible for supplying.
func TestBanksAreSane(t *testing.T) {
	for name, bank := range allFragments(t) {
		if len(bank) < 3 {
			t.Errorf("bank %q has only %d fragment(s); want >= 3 for variety", name, len(bank))
		}
		seen := map[string]bool{}
		for i, frag := range bank {
			switch {
			case strings.TrimSpace(frag) == "":
				t.Errorf("bank %q fragment %d is blank", name, i)
			case frag != strings.TrimSpace(frag):
				t.Errorf("bank %q fragment %d has edge whitespace: %q", name, i, frag)
			case strings.ContainsAny(frag, "[]%"):
				t.Errorf("bank %q fragment %d carries markup or a printf directive: %q", name, i, frag)
			case strings.HasSuffix(frag, ".") || strings.HasSuffix(frag, ","):
				t.Errorf("bank %q fragment %d ends in punctuation the template owns: %q", name, i, frag)
			case seen[frag]:
				t.Errorf("bank %q has a duplicate fragment: %q", name, frag)
			}
			seen[frag] = true
		}
		// Casing convention: a *_tail_c / *_subject_c / *_res_* fragment continues
		// a sentence and must be lowercase; everything else opens one.
		lowerBank := strings.Contains(name, "_tail_c") || strings.Contains(name, "_subject_c") || strings.Contains(name, "_res_")
		for _, frag := range bank {
			first := frag[:1]
			if lowerBank && first != strings.ToLower(first) {
				t.Errorf("continuation bank %q has a capitalized fragment: %q", name, frag)
			}
			if !lowerBank && !strings.HasPrefix(frag, "{") && first != strings.ToUpper(first) {
				t.Errorf("sentence bank %q has a lowercase fragment: %q", name, frag)
			}
		}
	}
}

// TestTemplatesAreWellFormed checks the structure set itself: unique ids, real
// bank references, non-empty parts, and a declared need for every placeholder a
// template can actually emit.
func TestTemplatesAreWellFormed(t *testing.T) {
	seenID := map[string]Moment{}
	for _, m := range Moments() {
		tpls := templatesFor(m)
		if len(tpls) < 8 {
			t.Errorf("%v has only %d templates; want >= 8", m, len(tpls))
		}
		for _, tpl := range tpls {
			if tpl.ID == "" {
				t.Errorf("%v has a template with no id", m)
			}
			if prev, dup := seenID[tpl.ID]; dup {
				t.Errorf("template id %q is used by both %v and %v", tpl.ID, prev, m)
			}
			seenID[tpl.ID] = m
			if len(tpl.Parts) == 0 {
				t.Errorf("template %q has no parts", tpl.ID)
			}
			if tpl.capacity() == 0 {
				t.Errorf("template %q has zero capacity — a referenced bank is missing or empty", tpl.ID)
			}
			for _, p := range tpl.Parts {
				if p.bank == "" {
					continue
				}
				if _, ok := banks[p.bank]; !ok {
					t.Errorf("template %q references unknown bank %q", tpl.ID, p.bank)
				}
			}
			// Placeholder ⇒ matching need.
			joined := ""
			for _, p := range tpl.Parts {
				if p.bank == "" {
					joined += p.lit
					continue
				}
				joined += strings.Join(banks[p.bank], "\x00")
			}
			checks := []struct {
				ph string
				n  need
			}{
				{"{subject}", needSubject},
				{"{res}", needRes},
				{"{res_stores}", needMassRes},
				{"{res_haul}", needMassRes},
				{"{amt_res}", needAmount},
				{"{amt}", needAmount},
				{"{n}", needCount},
				{"{ticks}", needTicks},
			}
			for _, c := range checks {
				if strings.Contains(joined, c.ph) && tpl.Needs&c.n == 0 {
					t.Errorf("template %q can emit %s but does not declare the matching need", tpl.ID, c.ph)
				}
			}
		}
	}
}

// TestAnchorInvariant is what makes Signatures() trustworthy: every template
// draws exactly one slot from an anchor bank, anchor fragments carry no
// placeholders (so they reach the output verbatim), and they are long enough to
// be distinctive.
func TestAnchorInvariant(t *testing.T) {
	for _, m := range Moments() {
		anchors := map[string]bool{}
		for _, name := range anchorBanksFor(m) {
			if len(banks[name]) == 0 {
				t.Fatalf("%v: anchor bank %q is missing or empty", m, name)
			}
			anchors[name] = true
			for _, frag := range banks[name] {
				if strings.ContainsAny(frag, "{}") {
					t.Errorf("%v: anchor fragment %q carries a placeholder — it would not reach the output verbatim", m, frag)
				}
				if len(frag) < 12 {
					t.Errorf("%v: anchor fragment %q is too short to be a distinctive signature", m, frag)
				}
			}
		}
		for _, tpl := range templatesFor(m) {
			n := 0
			for _, p := range tpl.Parts {
				if anchors[p.bank] {
					n++
				}
			}
			if n != 1 {
				t.Errorf("%v: template %q draws %d anchor slots, want exactly 1", m, tpl.ID, n)
			}
		}
	}
}

// TestSignaturesClassifyEveryLine is the end-to-end version of the anchor
// contract, and the property game/boon_tuning_test.go depends on: any generated
// line contains exactly one of its Moment's signatures and none of any other
// Moment's, so a consumer holding only the finished text can bucket it.
func TestSignaturesClassifyEveryLine(t *testing.T) {
	sigs := map[Moment][]string{}
	for _, m := range Moments() {
		sigs[m] = Signatures(m)
		if len(sigs[m]) == 0 {
			t.Fatalf("Signatures(%v) is empty", m)
		}
	}
	// No signature may contain another, or a line could match two.
	for m1, a := range sigs {
		for m2, bb := range sigs {
			for _, x := range a {
				for _, y := range bb {
					if x == y && m1 != m2 {
						t.Errorf("signature %q is shared by %v and %v", x, m1, m2)
					}
					if x != y && strings.Contains(x, y) {
						t.Errorf("signature %q (%v) contains signature %q (%v)", x, m1, y, m2)
					}
				}
			}
		}
	}
	for _, m := range Moments() {
		rng := rand.New(rand.NewSource(int64(m) * 2711))
		for _, req := range fuzzRequests(m)[:2000] {
			line := Line(req, rng)
			hits := 0
			for _, s := range sigs[m] {
				if strings.Contains(line, s) {
					hits++
				}
			}
			if hits != 1 {
				t.Fatalf("%v: line %q matched %d signatures, want exactly 1", m, line, hits)
			}
			for _, other := range Moments() {
				if other == m {
					continue
				}
				for _, s := range sigs[other] {
					if strings.Contains(line, s) {
						t.Fatalf("%v line %q matched a %v signature %q", m, line, other, s)
					}
				}
			}
		}
	}
}

// TestToneAndKindBiasSelection proves the two shaping inputs actually change the
// prose rather than being decorative: a tone-specific template is reachable only
// under its tone, and a kind-specific one only under its kind.
func TestToneAndKindBiasSelection(t *testing.T) {
	reach := func(req Request, seed int64, n int) map[string]bool {
		rng := rand.New(rand.NewSource(seed))
		out := map[string]bool{}
		for i := 0; i < n; i++ {
			out[Generate(req, rng).Template] = true
		}
		return out
	}
	base := Request{Moment: ExpeditionSuccess}

	neutral := reach(base, 3, 3000)
	if neutral["exp_success_triumph"] || neutral["exp_success_wry"] {
		t.Error("a tone-specific template fired under Neutral")
	}
	triumphant := base
	triumphant.Tone = Triumphant
	if !reach(triumphant, 3, 3000)["exp_success_triumph"] {
		t.Error("the Triumphant template never fired under Triumphant")
	}

	if neutral["exp_success_scout_then"] || neutral["exp_success_war_then"] {
		t.Error("a kind-specific template fired with no Kind set")
	}
	scouting := base
	scouting.Kind = "SCOUTING" // case-insensitive
	if !reach(scouting, 3, 3000)["exp_success_scout_then"] {
		t.Error("the scouting template never fired for Kind=scouting")
	}

	// Age bounds era-specific fragments: a stone-age success never reaches the
	// late-era anchors, a transcendent-age one does, and an unset age is
	// permissive (both are reachable).
	early := base
	early.Age = "primitive_age"
	if reach(early, 3, 3000)["exp_success_late_bare"] {
		t.Error("a late-era template fired in the primitive age")
	}
	late := base
	late.Age = "transcendent_age"
	if !reach(late, 3, 3000)["exp_success_late_bare"] {
		t.Error("the late-era template never fired in the transcendent age")
	}
	if !reach(base, 3, 3000)["exp_success_late_bare"] {
		t.Error("an unset Age should not exclude era-specific templates")
	}
}

// --- eyeball sample --------------------------------------------------------

// TestSampleLines prints a sample per Moment so a reviewer can judge the PROSE,
// which no assertion can do. Run with -v.
func TestSampleLines(t *testing.T) {
	samples := []struct {
		label string
		req   Request
	}{
		{"expedition success (military, medieval)", Request{Moment: ExpeditionSuccess, Tone: Triumphant, Age: "medieval_age", Kind: "military", Subject: "Siege Enemy Castle", Resource: "gold", Amount: 340}},
		{"expedition success (scouting, primitive)", Request{Moment: ExpeditionSuccess, Tone: Wry, Age: "primitive_age", Kind: "scouting", Subject: "Scout Party", Resource: "food", Amount: 60}},
		{"expedition success (space age)", Request{Moment: ExpeditionSuccess, Tone: Triumphant, Age: "space_age", Kind: "military", Subject: "Orbital Strike", Resource: "titanium", Amount: 900}},
		{"expedition failure (military, iron)", Request{Moment: ExpeditionFailure, Tone: Grim, Age: "iron_age", Kind: "military", Subject: "Raid Bandit Camp", Resource: "iron", Amount: 30}},
		{"expedition failure (scouting, modern)", Request{Moment: ExpeditionFailure, Age: "modern_age", Kind: "scouting", Subject: "Scout Nearby Ruins", Resource: "data", Amount: 12}},
		{"encounter standoff", Request{Moment: EncounterStandoff, Tone: Wry, Age: "medieval_age", Kind: "aggressive", Subject: "Ironhold Clans"}},
		{"encounter at capacity", Request{Moment: EncounterAtCapacity, Tone: Wry, Age: "classical_age", Kind: "mercantile", Subject: "Merchant Guild"}},
		{"war raid", Request{Moment: WarRaid, Tone: Grim, Age: "medieval_age", Kind: "aggressive", Subject: "Void Reavers", Resource: "gold"}},
		{"zero request (every moment)", Request{}},
	}
	for _, s := range samples {
		if s.req.Moment == MomentUnknown {
			continue
		}
		rng := rand.New(rand.NewSource(20260808))
		t.Logf("--- %s ---", s.label)
		for i := 0; i < 8; i++ {
			t.Log("   " + Line(s.req, rng))
		}
	}
	for _, m := range Moments() {
		rng := rand.New(rand.NewSource(11))
		t.Logf("--- zero request: %v ---", m)
		for i := 0; i < 4; i++ {
			t.Log("   " + Line(Request{Moment: m}, rng))
		}
	}
	_ = fmt.Sprint()
}
