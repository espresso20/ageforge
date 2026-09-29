package flavor

import (
	"go/parser"
	"go/token"
	"math"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
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

// eraAges names one representative age per bucket, for the tests that need to
// stand inside an era rather than reason about it.
var eraAges = map[era]string{
	eraAncient:    "stone_age",
	eraFeudal:     "renaissance_age",
	eraIndustrial: "atomic_age",
	eraDigital:    "cyberpunk_age",
	eraCosmic:     "quantum_age",
}

// fuzzRequests returns a broad cross-product of realistic requests for a Moment:
// every tone, both expedition kinds and a couple of personalities, several ages
// across all five eras, mass and count resources (with and without an amount),
// and both a subject and no subject.
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

// authoredUnit is one piece of prose a person wrote, plus where it lives and
// which eras it can fire in. The catalog holds prose in two places now — the
// SENTENCES (literal parts of a skeleton) and the NOUN BANKS their slots draw
// from — and every content lint has to see both or it only checks half the file.
type authoredUnit struct {
	text  string
	where string
	eras  map[era]bool
}

// authoredProse walks every registered Moment and yields every authored string
// in the catalog with the era set it can reach.
func authoredProse() []authoredUnit {
	var out []authoredUnit
	bankEras := map[string]map[era]bool{}
	for _, m := range Moments() {
		for _, tpl := range templatesFor(m) {
			eras := map[era]bool{}
			if len(tpl.Eras) == 0 {
				for _, bucket := range eraBuckets {
					eras[bucket.era] = true
				}
			} else {
				for _, e := range tpl.Eras {
					eras[e] = true
				}
			}
			out = append(out, authoredUnit{text: authoredText(tpl), where: tpl.ID, eras: eras})
			for _, p := range tpl.Parts {
				if p.bank == "" {
					continue
				}
				if bankEras[p.bank] == nil {
					bankEras[p.bank] = map[era]bool{}
				}
				for e := range eras {
					bankEras[p.bank][e] = true
				}
			}
		}
	}
	for name, eras := range bankEras {
		for _, frag := range banks[name] {
			out = append(out, authoredUnit{text: frag, where: "bank " + name, eras: eras})
		}
	}
	return out
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
			if strings.ContainsAny(line, "{}~") {
				t.Fatalf("%v: unfilled placeholder or slot in %q", m, line)
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
			if strings.ContainsAny(res.Text, "{}[]%~") {
				t.Fatalf("%v: zero-Request line is not clean: %q", m, res.Text)
			}
			seen[res.Text] = true
		}
		// The distinct-line floor is for the per-line Moments. The harbinger
		// Moments fire a few times a run and are sized by TestHarbingerPoolFloors;
		// a tiered warning with no tier deliberately reaches only its small
		// tier-neutral set.
		if !isRare(m) && len(seen) < 150 {
			t.Errorf("%v: zero Request produced only %d distinct lines in 500 draws; want >= 150", m, len(seen))
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
// FIXED number of draws (skeleton + one per slot) — the property that lets a
// caller share one rng across systems without the prose stream desynchronising
// everything downstream of it.
func TestGenerateUsesOnlyPassedRNG(t *testing.T) {
	for _, m := range Moments() {
		req := Request{Moment: m}
		gen := rand.New(rand.NewSource(7))
		res := Generate(req, gen)
		if res.Text == "" {
			t.Fatalf("%v: empty text", m)
		}
		manual := rand.New(rand.NewSource(7))
		p := eligible(req)
		pick := p[manual.Intn(len(p))]
		if pick.ID != res.Template {
			t.Fatalf("%v: skeleton pick is not the first draw (got %q, replayed %q)", m, res.Template, pick.ID)
		}
	}
}

// --- THE quality gate: how many separate things can it SAY -----------------

// TestSkeletonFloors is the metric that replaced Capacity as the quality bar.
//
// Capacity multiplies bank sizes together and so reports a number in the
// thousands for a Moment a player experiences as seven sentences on a loop:
// perceived repetition is governed by the SMALLEST pool a line draws from, and
// for a one-sentence-per-line catalog that pool is the sentences themselves.
// The floor that matters is therefore per ERA-ELIGIBLE POOL — the sentences an
// age in one bucket can actually reach with no tone, kind, subject or resource
// to widen it, which is the narrowest a real run ever gets.
func TestSkeletonFloors(t *testing.T) {
	const perEraFloor = 70

	for _, m := range Moments() {
		if isRare(m) {
			continue // sized for how often they fire: see TestHarbingerPoolFloors, TestRunEndingPoolFloors
		}
		total := DistinctSkeletons(m)
		if total == 0 {
			t.Fatalf("%v has no skeletons", m)
		}
		t.Logf("%-20v DistinctSkeletons = %3d   (Capacity = %d)", m, total, Capacity(m))
		for _, bucket := range eraBuckets {
			n := len(eligible(Request{Moment: m, Age: eraAges[bucket.era]}))
			if n < perEraFloor {
				t.Errorf("%v in the %s era can reach only %d skeletons; want >= %d — "+
					"that bucket needs more authored SENTENCES, not more nouns in a bank",
					m, bucket.suffix, n, perEraFloor)
			}
			t.Logf("    %-11s era-eligible skeletons = %3d", bucket.suffix, n)
		}
	}
	if DistinctSkeletons(MomentUnknown) != 0 {
		t.Errorf("DistinctSkeletons(MomentUnknown) = %d, want 0", DistinctSkeletons(MomentUnknown))
	}
}

// TestRepetitionInAStream is the test that would have caught the original
// failure: it reads a Moment the way a player does, as a run of consecutive
// lines, and asserts the catalog does not run out.
//
// It measures both paths, and asserts different things about each, because they
// have different jobs.
//
//   - The RAW generator is memoryless by contract — Generate is pure given
//     (req, rng), which is what lets a seed reproduce a whole run's prose. For a
//     uniform draw of 200 samples from a pool of N the per-skeleton count is
//     Poisson(200/N), so a max of 5 or 6 is simply what the arithmetic produces at
//     any catalog size a person can write: the expected number of skeletons seen
//     5+ times only falls below one around N = 400 PER ERA PER MOMENT. So the raw
//     numbers are logged, and what is asserted about them is DEPTH — 60 distinct
//     skeletons and 70 distinct finished lines out of 200 — which is the thing
//     more authoring actually buys.
//
//   - The STREAM is the path the engine uses, and it has a memory, so it can be
//     held to a real ceiling. Recency suppression, not catalog size, is what
//     removes the complaint this rewrite came from: the same sentence three times
//     inside twenty-four lines.
//
// The firing-rate gate on the game side widens the window further: at roughly one
// resolution in three, 200 flavour lines is several hours of play.
func TestRepetitionInAStream(t *testing.T) {
	const (
		draws            = 200
		wantDistinct     = 60
		wantDistinctLine = 70
		maxStreamRepeats = 5
		// seeds is how many independent runs the Stream-versus-raw comparison
		// is summed over. One seed of 200 draws put the comparison at the mercy
		// of which slot fills that seed happened to hit; five make it a
		// measurement.
		seeds = 5
	)
	for _, m := range Moments() {
		if isRare(m) {
			continue // a few lines a run, not two hundred: see TestHarbingerPoolFloors
		}
		for _, bucket := range eraBuckets {
			age := eraAges[bucket.era]
			var rawSk, rawLn, stSk, stLn int
			for k := 0; k < seeds; k++ {
				seed := int64(m)*104729 + int64(bucket.era) + int64(k)*7919
				rng := rand.New(rand.NewSource(seed))
				bySkeleton := map[string]int{}
				byLine := map[string]int{}
				worst := 0
				for i := 0; i < draws; i++ {
					res := Generate(Request{Moment: m, Age: age}, rng)
					bySkeleton[res.Template]++
					byLine[res.Text]++
					if bySkeleton[res.Template] > worst {
						worst = bySkeleton[res.Template]
					}
				}
				if len(bySkeleton) < wantDistinct {
					t.Errorf("%v at %s seed %d: only %d distinct skeletons in %d lines; want >= %d",
						m, age, k, len(bySkeleton), draws, wantDistinct)
				}
				if len(byLine) < wantDistinctLine {
					t.Errorf("%v at %s seed %d: only %d distinct finished lines in %d; want >= %d",
						m, age, k, len(byLine), draws, wantDistinctLine)
				}
				// No assertion on the RAW max on purpose: a uniform memoryless draw
				// puts it wherever the seed feels like. The same draws as a PLAYER
				// meets them, through a Stream, can be held to a real ceiling.
				srng := rand.New(rand.NewSource(seed))
				st := NewStream()
				streamSk := map[string]int{}
				streamLines := map[string]int{}
				streamWorst := 0
				for i := 0; i < draws; i++ {
					res := st.Generate(Request{Moment: m, Age: age}, srng)
					streamSk[res.Template]++
					streamLines[res.Text]++
					if streamSk[res.Template] > streamWorst {
						streamWorst = streamSk[res.Template]
					}
				}
				// The ceiling is asserted on the first seed only, as it always was:
				// with a 32-line window a skeleton can legitimately recur every 33
				// lines, so across many seeds one of them eventually shows six.
				if k == 0 && streamWorst > maxStreamRepeats {
					t.Errorf("%v at %s seed %d: through a Stream, one skeleton still repeated %d times in %d lines; want <= %d",
						m, age, k, streamWorst, draws, maxStreamRepeats)
				}
				rawSk += len(bySkeleton)
				rawLn += len(byLine)
				stSk += len(streamSk)
				stLn += len(streamLines)
				if k == 0 {
					t.Logf("%-20v %-11s raw: %3d skeletons / %3d lines / worst %d    stream: %3d skeletons / %3d lines / worst %d",
						m, bucket.suffix, len(bySkeleton), len(byLine), worst,
						len(streamSk), len(streamLines), streamWorst)
				}
			}
			// The Stream must not be WORSE than the raw generator. Distinct
			// SKELETONS is held strictly, summed over the seeds: that is what a
			// player perceives as repetition. Distinct finished LINES gets ten
			// percent of slack, and it is earned: a Stream also suppresses a
			// repeated noun-phrase draw, and the redraw often lands on a slotless
			// sentence, which trades a noun variant for a wider spread of
			// sentences.
			if stSk < rawSk {
				t.Errorf("%v at %s: over %d seeds the Stream produced FEWER distinct skeletons (%d) than the raw generator (%d)",
					m, age, seeds, stSk, rawSk)
			}
			if stLn*10 < rawLn*9 {
				t.Errorf("%v at %s: over %d seeds the Stream produced materially fewer distinct lines (%d) than the raw generator (%d)",
					m, age, seeds, stLn, rawLn)
			}
		}
	}
}

// --- the shape of an authored sentence -------------------------------------

// TestSkeletonShape is what makes Signatures() trustworthy, and it is a much
// smaller claim than the old anchor-bank invariant it replaced: a skeleton is one
// sentence, it opens with authored literal text, and it carries at most one slot.
// Everything before the slot therefore reaches the output verbatim and is the
// line's signature.
func TestSkeletonShape(t *testing.T) {
	for _, m := range Moments() {
		for _, tpl := range templatesFor(m) {
			slots := 0
			for _, p := range tpl.Parts {
				if p.bank != "" {
					slots++
				}
			}
			if slots > 1 {
				t.Errorf("%v: skeleton %q has %d slots; a sentence may carry at most one, "+
					"because two independently drawn phrases in one sentence is the "+
					"assembly this catalog exists to avoid", m, tpl.ID, slots)
			}
			if len(tpl.Parts) == 0 || tpl.Parts[0].bank != "" {
				t.Errorf("%v: skeleton %q does not open on authored text", m, tpl.ID)
				continue
			}
			sig := anchorOf(tpl)
			if len(sig) < 12 {
				t.Errorf("%v: skeleton %q opens with only %q before its first slot or "+
					"placeholder; want >= 12 characters so the signature is distinctive",
					m, tpl.ID, sig)
			}
			if strings.ContainsAny(sig, "{}~") {
				t.Errorf("%v: signature %q of skeleton %q is not placeholder-free", m, sig, tpl.ID)
			}
		}
	}
}

// --- C1: BURSTINESS, the gate both previous catalogs would have failed ------

// The diagnosis this test exists to encode.
//
// Two rewrites of this catalog were rejected as machine-written. The first joined
// clauses and averaged 12-25 words a line. The second deleted the clause joiner,
// wrote every line as one authored sentence, and capped length at 14 words: mean
// 10.3, longest 14. It read exactly as artificial as the first.
//
// The shared cause is not word choice. It is BURSTINESS — variance in sentence
// length. Human prose swings from four words to forty and the swing is purposeful:
// short at the moment of emphasis, long while a thought is still being qualified.
// A generator clusters in a narrow band. Both catalogs had a narrow band; the only
// thing that changed between them was WHERE the band sat. Uniform rhythm is the
// tell, independently of how good any single line is.
//
// So the cap is gone and the distribution is the assertion. A Moment must populate
// four length bands in roughly the shares below, and the standard deviation of its
// sentence lengths must clear 7 words. The old catalog scored SD 2.1.
var lengthBands = []struct {
	name   string
	lo, hi int
	want   float64
}{
	{"very short  3-6", 3, 6, 0.20},
	{"mid         7-16", 7, 16, 0.45},
	{"long       17-32", 17, 32, 0.25},
	{"very long  33-45", 33, 45, 0.10},
}

const (
	// minSentenceSD is the burstiness floor, in words. Pass 2 scored 2.1.
	minSentenceSD = 7.0
	// bandTolerance is how far a band's share may sit from its target, in
	// percentage points.
	bandTolerance = 0.07
	minWords      = 3
	maxWords      = 45
)

func TestBurstiness(t *testing.T) {
	for _, m := range Moments() {
		tpls := templatesFor(m)
		lens := make([]int, 0, len(tpls))
		bands := make([]int, len(lengthBands))
		for _, tpl := range tpls {
			w := wordCount(authoredText(tpl))
			lens = append(lens, w)
			switch {
			case w < minWords:
				t.Errorf("%v: skeleton %q is only %d words — %q", m, tpl.ID, w, authoredText(tpl))
			case w > maxWords:
				t.Errorf("%v: skeleton %q is %d words, over the %d ceiling — %q", m, tpl.ID, w, maxWords, authoredText(tpl))
			}
			for i, band := range lengthBands {
				if w >= band.lo && w <= band.hi {
					bands[i]++
				}
			}
		}
		n := float64(len(lens))
		mean, sd := meanSD(lens)
		if sd < minSentenceSD {
			t.Errorf("%v: sentence-length SD is %.2f, want >= %.1f — the lines are all the same "+
				"LENGTH, which is the tell that sank both previous catalogs regardless of how "+
				"good the individual sentences were", m, sd, minSentenceSD)
		}
		for i, band := range lengthBands {
			got := float64(bands[i]) / n
			if got < band.want-bandTolerance || got > band.want+bandTolerance {
				t.Errorf("%v: the %q band holds %.0f%% of skeletons (%d/%d), want %.0f%% ± %.0f — "+
					"a missing band is a missing rhythm", m, band.name, got*100, bands[i],
					len(tpls), band.want*100, bandTolerance*100)
			}
		}
		t.Logf("%-20v n=%3d  mean %5.1f  SD %5.2f   %s", m, len(tpls), mean, sd, bandReport(bands, n))
	}
}

// meanSD returns the mean and population standard deviation of a length sample.
func meanSD(xs []int) (float64, float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	var sum float64
	for _, x := range xs {
		sum += float64(x)
	}
	mean := sum / float64(len(xs))
	var ss float64
	for _, x := range xs {
		d := float64(x) - mean
		ss += d * d
	}
	return mean, math.Sqrt(ss / float64(len(xs)))
}

// bandReport renders the band shares for the log line.
func bandReport(bands []int, n float64) string {
	var sb strings.Builder
	for i, band := range lengthBands {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(band.name)
		sb.WriteString(" ")
		sb.WriteString(strconv.Itoa(int(float64(bands[i])/n*100 + 0.5)))
		sb.WriteString("%")
	}
	return sb.String()
}

// --- C2: the named AI signatures, banned by lint ---------------------------

// TestNoAITells fails on the sentence SHAPES that read as machine-written
// independently of their content. Every one of them builds to a weighted final
// clause — the reader is handed a setup and then a verdict — and a catalog
// saturated with them reads as a catalog of aphorisms rather than as a chronicle.
//
// They are hard to see while authoring precisely because they feel like good
// writing. Pass 2 was full of them and none was noticed until a reviewer read
// forty lines in a row.
func TestNoAITells(t *testing.T) {
	type tell struct {
		name string
		re   *regexp.Regexp
	}
	tells := []tell{
		{
			"two-beat negation parallel (\"X. Y did not.\")",
			regexp.MustCompile(`(?i)\.\s+[^.]*\b(did not|didn't|does not|doesn't|do not|don't|was not|wasn't|were not|weren't|is not|isn't|are not|aren't|has not|hasn't|have not|haven't|had not|hadn't|will not|won't|would not|wouldn't|could not|couldn't|cannot|can't)\s*$`),
		},
		{
			"false contrast (\"not X, but Y\")",
			regexp.MustCompile(`(?i)\bnot\b[^.]{0,45}\bbut\b`),
		},
		{
			"false contrast (\"not X — Y\")",
			regexp.MustCompile(`(?i)\bnot\b[^.]{0,45}—`),
		},
		{
			"appositive negation (\"cut, not broken,\")",
			regexp.MustCompile(`(?i),\s*not\s+[\w']+\s*,`),
		},
		{
			"withheld payload (\"isn't X, it's Y\")",
			regexp.MustCompile(`(?i)\b(isn't|is not|it's not|was not|wasn't|were not|weren't)\b[^.]{0,45},\s*(it's|it is|its|that's|that is|they're|they are)\b`),
		},
		{
			"\"not only / not just / so much as\"",
			regexp.MustCompile(`(?i)\b(not (only|just|merely|simply|so much)|so much as|less a \w+ than)\b`),
		},
		{
			"trailing verdict clause",
			regexp.MustCompile(`(?i)\b(and that (is|was) that|and there it is|and so it goes|make of that what you will|speaks for itself|says everything|tells you everything|draw your own)\b`),
		},
		{
			"trailing verdict noun (\"which is the whole problem\")",
			regexp.MustCompile(`(?i),\s*(and that|which)\s+(is|was)\s+(the\s+)?(point|problem|whole|trouble|joke|reason|end|worst|best|lesson|price|cost|shape|size|story)\b`),
		},
		{
			"negation triad used for rhythm",
			regexp.MustCompile(`(?i)\b(no|nothing|not one|never)\b[^.,]{1,20},\s*(no|nothing|not one|never)\b[^.,]{1,20},\s*(and\s+)?(no|nothing|not one|never)\b`),
		},
		{
			"\"No X. No Y.\" opener",
			regexp.MustCompile(`(?i)^\s*(no|nothing|never|none)\b[^.]*\.\s*(no|nothing|never|none)\b`),
		},
		{
			"\", and then nothing\"",
			regexp.MustCompile(`(?i),\s*and then nothing\b`),
		},
	}

	emdash, total := 0, 0
	perMomentDash := map[Moment]int{}
	perMomentTotal := map[Moment]int{}
	for _, m := range Moments() {
		for _, tpl := range templatesFor(m) {
			text := strings.TrimSpace(authoredText(tpl))
			total++
			perMomentTotal[m]++
			for _, x := range tells {
				if x.re.MatchString(text) {
					t.Errorf("%v %s is a %s:\n    %q", m, tpl.ID, x.name, text)
				}
			}
			if strings.Contains(text, ":") {
				t.Errorf("%v %s uses a colon, which in this catalog only ever introduced a "+
					"verdict (\"two ropes and one broken hand: the whole cost\"):\n    %q", m, tpl.ID, text)
			}
			if strings.Contains(text, "—") {
				emdash++
				perMomentDash[m]++
			}
			// A bare fragment followed by a judgment: "Two dead, and both of them
			// were careless." The head is too short to be a clause, so the sentence
			// exists only to deliver its tail.
			if head, tail, ok := strings.Cut(text, ","); ok {
				tailStart := strings.ToLower(strings.Fields(strings.TrimSpace(tail) + " x")[0])
				switch tailStart {
				case "and", "but", "which", "both", "all", "though", "so":
					if wordCount(head) <= 3 {
						t.Errorf("%v %s is a fragment plus a judgment — the opening %q is too "+
							"short to be a clause, so the line exists only to deliver its "+
							"verdict:\n    %q", m, tpl.ID, strings.TrimSpace(head), text)
					}
				}
			}
		}
	}

	const maxDashShare = 0.08
	if share := float64(emdash) / float64(total); share > maxDashShare {
		t.Errorf("%d of %d skeletons (%.1f%%) contain an em dash; the cap is %.0f%% — "+
			"in volume it is a tell on its own", emdash, total, share*100, maxDashShare*100)
	}
	for _, m := range Moments() {
		share := float64(perMomentDash[m]) / float64(perMomentTotal[m])
		if share > maxDashShare {
			t.Errorf("%v: %.1f%% of skeletons contain an em dash; the cap is %.0f%%", m, share*100, maxDashShare*100)
		}
	}
	t.Logf("%d skeletons checked against %d tells; %d carry an em dash (%.1f%%)",
		total, len(tells), emdash, float64(emdash)/float64(total)*100)
}

// --- C3: register quotas ---------------------------------------------------

// TestRegisterQuotas is the structural answer to single-voice convergence. An LLM
// told to write good flavour text converges on terse ironic understatement, and no
// amount of instruction to "vary the tone" survives two hundred lines. So the mix
// is declared per sentence and asserted here.
//
//   - at least 30% rPlain: no irony, no wit, a circumstantial fact and nothing
//     else. Both previous catalogs had approximately none of this — pass 2 replaced
//     loud irony with quiet irony rather than with none.
//   - at least 15% non-narrative: a ledger line, a grumble, an overheard fragment,
//     a posted notice. Not everything a log says has to be narration.
//   - at most 20% reaching for a joke.
func TestRegisterQuotas(t *testing.T) {
	const (
		minPlain    = 0.30
		maxJoke     = 0.20
		minNonNarr  = 0.15
		minEachForm = 3
	)
	regName := map[register]string{regUnset: "UNSET", rPlain: "plain", rWry: "wry", rJoke: "joke"}
	formName := map[form]string{fNarr: "narration", fLedger: "ledger", fComplaint: "complaint", fOverheard: "overheard", fNotice: "notice"}

	for _, m := range Moments() {
		tpls := templatesFor(m)
		regs := map[register]int{}
		forms := map[form]int{}
		for _, tpl := range tpls {
			if tpl.Reg == regUnset {
				t.Errorf("%v %s has no register tag — every sentence must declare Reg, so a "+
					"forgotten tag cannot vote in the quota: %q", m, tpl.ID, authoredText(tpl))
			}
			regs[tpl.Reg]++
			forms[tpl.Form]++
		}
		n := float64(len(tpls))
		plain := float64(regs[rPlain]) / n
		joke := float64(regs[rJoke]) / n
		nonNarr := float64(len(tpls)-forms[fNarr]) / n

		if plain < minPlain {
			t.Errorf("%v: only %.0f%% of skeletons are rPlain, want >= %.0f%% — a catalog where "+
				"every line has an angle is exhausting and none of the angles land",
				m, plain*100, minPlain*100)
		}
		if joke > maxJoke {
			t.Errorf("%v: %.0f%% of skeletons reach for a joke, the cap is %.0f%%", m, joke*100, maxJoke*100)
		}
		if nonNarr < minNonNarr {
			t.Errorf("%v: only %.0f%% of skeletons are non-narrative, want >= %.0f%% — add ledger "+
				"lines, complaints, overheard fragments and notices", m, nonNarr*100, minNonNarr*100)
		}
		for _, f := range []form{fLedger, fComplaint, fOverheard, fNotice} {
			if forms[f] < minEachForm {
				t.Errorf("%v: only %d %s skeletons; want >= %d so the form is actually reachable",
					m, forms[f], formName[f], minEachForm)
			}
		}
		t.Logf("%-20v plain %2.0f%%  wry %2.0f%%  joke %2.0f%%   |  narration %2.0f%%  ledger %d  complaint %d  overheard %d  notice %d",
			m, plain*100, float64(regs[rWry])/n*100, joke*100, float64(forms[fNarr])/n*100,
			forms[fLedger], forms[fComplaint], forms[fOverheard], forms[fNotice])
		_ = regName
	}
}

// TestTopicVocabulary keeps the Stream's subject-matter suppression honest: every
// sentence declares what it is ABOUT, from a closed vocabulary, so a typo cannot
// split one topic into two that never suppress each other.
func TestTopicVocabulary(t *testing.T) {
	used := map[string]int{}
	for _, m := range Moments() {
		byTopic := map[string]int{}
		for _, tpl := range templatesFor(m) {
			switch {
			case tpl.Topic == "":
				t.Errorf("%v %s has no Topic — the Stream cannot tell it apart from the line "+
					"before it: %q", m, tpl.ID, authoredText(tpl))
			case !knownTopics[tpl.Topic]:
				t.Errorf("%v %s has topic %q, which is not in knownTopics", m, tpl.ID, tpl.Topic)
			}
			byTopic[tpl.Topic]++
			used[tpl.Topic]++
		}
		if len(byTopic) < 14 {
			t.Errorf("%v spreads over only %d topics; want >= 14 or the topic ring starves", m, len(byTopic))
		}
	}
	for topic := range knownTopics {
		if used[topic] == 0 {
			t.Logf("note: topic %q is declared but unused", topic)
		}
	}
}

// TestOpenerVariety catches the other uniformity a length histogram cannot see:
// forty lines that all begin "Nobody has" read as one voice even when every one of
// them is a different length.
func TestOpenerVariety(t *testing.T) {
	const (
		maxWordShare   = 0.08
		maxBigramShare = 0.04
	)
	articles := map[string]bool{"the": true, "a": true, "an": true}
	for _, m := range Moments() {
		tpls := templatesFor(m)
		n := float64(len(tpls))
		words, bigrams := map[string]int{}, map[string]int{}
		for _, tpl := range tpls {
			f := strings.Fields(strings.ToLower(strings.Trim(authoredText(tpl), " ")))
			if len(f) == 0 {
				continue
			}
			words[f[0]]++
			if len(f) > 1 {
				bigrams[f[0]+" "+f[1]]++
			}
		}
		for w, c := range words {
			if articles[w] {
				continue
			}
			if share := float64(c) / n; share > maxWordShare {
				t.Errorf("%v: %.0f%% of skeletons open with %q (%d of %d); the cap is %.0f%%",
					m, share*100, w, c, len(tpls), maxWordShare*100)
			}
		}
		for bg, c := range bigrams {
			if share := float64(c) / n; share > maxBigramShare {
				t.Errorf("%v: %.0f%% of skeletons open with %q (%d of %d); the cap is %.0f%%",
					m, share*100, bg, c, len(tpls), maxBigramShare*100)
			}
		}
	}
}

// TestNoOutcomeRestating enforces the orthogonality rule. The caller prints the
// mechanical line — "Scout Party succeeded. Loot: 60 food, 60 wood, 20 stone." — immediately above
// this one, so a flavour line that also says the venture paid out is padding
// dressed as prose. It is a literal grep because the failure mode was literal:
// the old catalog leaned on a small set of accounting idioms and they were the
// phrases players saw three times a screen.
func TestNoOutcomeRestating(t *testing.T) {
	banned := []string{
		"pays out", "paid out", "closes out", "in profit", "the accounting",
		"goes into the record", "for the record", "the record",
		"the venture", "the undertaking", "the expedition",
		"all told", "on balance", "not as badly", "could have been worse",
		"worse than it could", "tally up", "the ledger", "breaks even",
		"says nothing, loudly", "comes to nothing", "at capacity",
		"a good one", "a bad one",
	}
	for _, u := range authoredProse() {
		low := strings.ToLower(u.text)
		for _, b := range banned {
			if strings.Contains(low, b) {
				t.Errorf("%s restates the mechanical outcome: %q contains the banned phrase %q",
					u.where, strings.TrimSpace(u.text), b)
			}
		}
		if strings.HasPrefix(u.text, "The party ") || strings.Contains(u.text, ". The party ") {
			t.Errorf("%s uses \"the party\" as a bare subject: %q — name someone or something instead",
				u.where, strings.TrimSpace(u.text))
		}
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
	// One alternation per check rather than one regexp per label: the same
	// matches, at a fraction of the cost once fourteen Moments feed it.
	var plural, article []string
	for key, class := range resourceClass {
		label := regexp.QuoteMeta(resourceLabel(key))
		plural = append(plural, label)
		if class == massNoun {
			article = append(article, label)
		}
	}
	sort.Strings(plural)
	sort.Strings(article)
	badPlural := []*regexp.Regexp{regexp.MustCompile(`(?i)\b(?:` + strings.Join(plural, "|") + `)s\b`)}
	badArticle := []*regexp.Regexp{regexp.MustCompile(`(?i)\ba (?:` + strings.Join(article, "|") + `)\b`)}

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
	if got := phraseBare(""); got != "supplies" {
		t.Errorf("phraseBare(\"\") = %q, want the generic fallback", got)
	}
}

// TestMassOnlyFramesNeverReachCountNouns is the structural half of the
// countability guarantee: the {res_stores} / {res_haul} frames are gated on
// needMassRes, so no skeleton that uses them can be picked for "soldiers".
func TestMassOnlyFramesNeverReachCountNouns(t *testing.T) {
	massOnly := []string{"{res_stores}", "{res_haul}"}
	for _, m := range Moments() {
		for _, tpl := range templatesFor(m) {
			text := authoredText(tpl)
			for _, p := range tpl.Parts {
				if p.bank != "" {
					text += " " + strings.Join(banks[p.bank], " ")
				}
			}
			for _, f := range massOnly {
				if strings.Contains(text, f) && tpl.Needs&needMassRes == 0 {
					t.Errorf("skeleton %q uses the mass-only frame %s but does not declare needMassRes", tpl.ID, f)
				}
			}
		}
	}
	for _, m := range Moments() {
		for _, key := range []string{"soldiers", "nanobots", "dark_matter_crystals"} {
			for _, tpl := range eligible(Request{Moment: m, Resource: key, Amount: 5}) {
				if tpl.Needs&needMassRes != 0 {
					t.Errorf("%v: skeleton %q is eligible for count noun %q", m, tpl.ID, key)
				}
			}
		}
	}
}

// --- the two grammar traps -------------------------------------------------

// TestVerbLedExpeditionSubjects drives every REAL expedition name through both
// expedition Moments. Half of them are verb-led ("Conquer Territory"), so a
// sentence that used the name as a subject would produce nonsense; this asserts
// the output stays clean and that the name, when used, survives verbatim.
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
				if strings.ContainsAny(res.Text, "{}[]%~") || strings.Contains(res.Text, "  ") {
					t.Fatalf("%v/%q: malformed line %q", m, name, res.Text)
				}
				if strings.Contains(res.Text, name) {
					usedSubject = true
				}
			}
			if !usedSubject {
				t.Errorf("%v: expedition name %q never appeared in 400 draws — the subject skeletons are unreachable", m, name)
			}
		}
	}
}

// TestSubjectFrames is a static lint on the authored prose, and the durable half
// of the grammar guarantee: it constrains what a future author may write.
//
//   - Expedition moments: {subject} is an ORDER TITLE and must sit inside one of
//     the title frames, never in a bare subject position.
//   - Faction moments: {subject} is a proper noun of unknown grammatical NUMBER,
//     so it may never be followed by a copula or auxiliary — use a past-tense
//     verb or an object position instead.
func TestSubjectFrames(t *testing.T) {
	titleFrames := []string{
		"the {subject} order", "marked {subject}", "titled {subject}",
		"filed under {subject}", "reads {subject}",
	}
	copulas := []string{
		"{subject} is", "{subject} are", "{subject} was", "{subject} were",
		"{subject} has", "{subject} have", "{subject} do", "{subject} does",
	}
	expeditionMoments := map[Moment]bool{ExpeditionSuccess: true, ExpeditionFailure: true}

	for _, m := range Moments() {
		texts := []string{}
		for _, tpl := range templatesFor(m) {
			texts = append(texts, authoredText(tpl))
			for _, p := range tpl.Parts {
				if p.bank != "" {
					texts = append(texts, banks[p.bank]...)
				}
			}
		}
		for _, frag := range texts {
			if !strings.Contains(frag, "{subject}") {
				continue
			}
			if expeditionMoments[m] {
				ok := false
				low := strings.ToLower(frag)
				for _, f := range titleFrames {
					if strings.Contains(low, f) {
						ok = true
						break
					}
				}
				if !ok {
					t.Errorf("%v: %q uses {subject} outside a title frame — "+
						"expedition names are verb-led and cannot be sentence subjects", m, frag)
				}
				continue
			}
			// Harbinger names are all singular ("the Oracle", "your future
			// self"; TestHarbingerLabels in config pins their shape), so a
			// copula after one is grammatical. RunEnding's Subject is one too.
			if isRare(m) {
				continue
			}
			for _, c := range copulas {
				if strings.Contains(frag, c) {
					t.Errorf("%v: %q puts a copula after {subject} — faction names "+
						"have inconsistent grammatical number; use past tense or an object position", m, frag)
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
				if res.Text == "" || strings.ContainsAny(res.Text, "{}[]%~") {
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

// TestBanksAreSane holds the noun banks to their one job: a bank is a set of
// interchangeable lowercase NOUN PHRASES of a single tight category, which is the
// only kind of substitution that reads reliably. Anything that opens a sentence,
// carries punctuation, or hides a placeholder is not a noun phrase.
func TestBanksAreSane(t *testing.T) {
	used := map[string]bool{}
	for _, m := range Moments() {
		for _, tpl := range templatesFor(m) {
			for _, p := range tpl.Parts {
				if p.bank != "" {
					used[p.bank] = true
				}
			}
		}
	}
	for name, bank := range banks {
		if !used[name] {
			t.Errorf("bank %q is never referenced by a skeleton", name)
		}
		if len(bank) < 6 {
			t.Errorf("bank %q has only %d fragment(s); want >= 6 so a slot is worth having", name, len(bank))
		}
		seen := map[string]bool{}
		for i, frag := range bank {
			switch {
			case strings.TrimSpace(frag) == "":
				t.Errorf("bank %q fragment %d is blank", name, i)
			case frag != strings.TrimSpace(frag):
				t.Errorf("bank %q fragment %d has edge whitespace: %q", name, i, frag)
			case strings.ContainsAny(frag, "[]%{}~"):
				t.Errorf("bank %q fragment %d is not a plain noun phrase: %q", name, i, frag)
			case strings.HasSuffix(frag, ".") || strings.HasSuffix(frag, ","):
				t.Errorf("bank %q fragment %d ends in punctuation: %q", name, i, frag)
			case seen[frag]:
				t.Errorf("bank %q has a duplicate fragment: %q", name, frag)
			}
			seen[frag] = true
			if first := frag[:1]; first != strings.ToLower(first) {
				t.Errorf("bank %q fragment %q is capitalized — a slot sits mid-sentence", name, frag)
			}
		}
	}
}

// TestTemplatesAreWellFormed checks the skeleton set itself: unique ids, real
// bank references, non-empty parts, no duplicated prose, and a declared need for
// every placeholder a sentence can actually emit.
func TestTemplatesAreWellFormed(t *testing.T) {
	seenID := map[string]Moment{}
	seenText := map[string]string{}
	for _, m := range Moments() {
		tpls := templatesFor(m)
		if len(tpls) < 8 {
			t.Errorf("%v has only %d skeletons; want >= 8", m, len(tpls))
		}
		for _, tpl := range tpls {
			if tpl.ID == "" {
				t.Errorf("%v has a skeleton with no id", m)
			}
			if prev, dup := seenID[tpl.ID]; dup {
				t.Errorf("skeleton id %q is used by both %v and %v", tpl.ID, prev, m)
			}
			seenID[tpl.ID] = m
			text := authoredText(tpl)
			if prev, dup := seenText[text]; dup {
				t.Errorf("skeleton %q duplicates the prose of %q: %q", tpl.ID, prev, text)
			}
			seenText[text] = tpl.ID
			if len(tpl.Parts) == 0 {
				t.Errorf("skeleton %q has no parts", tpl.ID)
			}
			if tpl.capacity() == 0 {
				t.Errorf("skeleton %q has zero capacity — a referenced bank is missing or empty", tpl.ID)
			}
			for _, p := range tpl.Parts {
				if p.bank == "" {
					continue
				}
				if _, ok := banks[p.bank]; !ok {
					t.Errorf("skeleton %q references unknown bank %q", tpl.ID, p.bank)
				}
			}
			joined := text
			for _, p := range tpl.Parts {
				if p.bank != "" {
					joined += "\x00" + strings.Join(banks[p.bank], "\x00")
				}
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
					t.Errorf("skeleton %q can emit %s but does not declare the matching need", tpl.ID, c.ph)
				}
			}
		}
	}
}

// --- era bleed -------------------------------------------------------------

// eraMarkers maps an ERA-CODED word to the era buckets it is at home in. A word
// with more than one home — a ford is as ancient as it is feudal, a road exists
// in every age that has ground — is only foreign outside all of them.
//
// This is the durable half of the anachronism guarantee, and it matters MORE now
// than it did: rule 2 of the catalog asks authors to name concrete things, and
// concrete things are era-bound almost by definition. An ungated sentence fires
// in every era, so it may contain no marker at all; an era-gated sentence may
// contain only markers at home in its own bucket.
var eraMarkers = map[string][]era{
	// --- ancient: fires, herds, hides, the elders ---------------------------
	"elders":     {eraAncient},
	"spear":      {eraAncient},
	"spears":     {eraAncient},
	"herd":       {eraAncient},
	"herds":      {eraAncient},
	"hide":       {eraAncient},
	"hides":      {eraAncient},
	"hut":        {eraAncient},
	"huts":       {eraAncient},
	"drums":      {eraAncient},
	"watch-fire": {eraAncient},
	"store-pit":  {eraAncient},
	"arrow":      {eraAncient, eraFeudal},
	"spring":     {eraAncient, eraFeudal},
	"ford":       {eraAncient, eraFeudal},

	// --- feudal: carts, gates, bells, clerks, musters ------------------------
	"cart":          {eraFeudal},
	"carts":         {eraFeudal},
	"gate":          {eraFeudal},
	"gates":         {eraFeudal},
	"bell":          {eraFeudal},
	"bells":         {eraFeudal},
	"horn":          {eraFeudal},
	"horns":         {eraFeudal},
	"militia":       {eraFeudal},
	"muster":        {eraFeudal},
	"quartermaster": {eraFeudal},
	"clerk":         {eraFeudal},
	"clerks":        {eraFeudal},
	"banner":        {eraFeudal},
	"banners":       {eraFeudal},
	"rider":         {eraFeudal},
	"riders":        {eraFeudal},
	"wagon":         {eraFeudal},
	"wagons":        {eraFeudal},
	"steward":       {eraFeudal},
	"stewards":      {eraFeudal},
	"herald":        {eraFeudal},
	"heralds":       {eraFeudal},
	"parchment":     {eraFeudal},
	"undercroft":    {eraFeudal},
	"picket":        {eraFeudal},
	"pickets":       {eraFeudal},
	"village":       {eraFeudal},
	"villages":      {eraFeudal},
	"supper":        {eraFeudal},
	"storehouse":    {eraFeudal},
	"wax":           {eraFeudal},

	// --- industrial: depots, sidings, the telegraph --------------------------
	"telegraph":    {eraIndustrial},
	"telegraphed":  {eraIndustrial},
	"telegram":     {eraIndustrial},
	"depot":        {eraIndustrial},
	"foreman":      {eraIndustrial},
	"siding":       {eraIndustrial},
	"freight":      {eraIndustrial},
	"lorries":      {eraIndustrial},
	"telephone":    {eraIndustrial},
	"platform":     {eraIndustrial},
	"payroll":      {eraIndustrial},
	"shareholders": {eraIndustrial},
	"triplicate":   {eraIndustrial},
	"annexe":       {eraIndustrial},
	"warehouse":    {eraIndustrial},
	"rail":         {eraIndustrial},
	"train":        {eraIndustrial},
	"wire":         {eraIndustrial},

	// --- digital: feeds, uplinks, drones, analysts ---------------------------
	"uplink":       {eraDigital},
	"drone":        {eraDigital},
	"drones":       {eraDigital},
	"feed":         {eraDigital},
	"analyst":      {eraDigital},
	"analysts":     {eraDigital},
	"network":      {eraDigital},
	"after-action": {eraDigital},
	"channel":      {eraDigital},
	"channels":     {eraDigital},

	// --- cosmic: holds, bays, hulls, relays ----------------------------------
	"hull":         {eraCosmic},
	"hulls":        {eraCosmic},
	"bay":          {eraCosmic},
	"bays":         {eraCosmic},
	"orbit":        {eraCosmic},
	"orbital":      {eraCosmic},
	"transponder":  {eraCosmic},
	"transponders": {eraCosmic},
	"vacuum":       {eraCosmic},
	"bulkhead":     {eraCosmic},
	"bulkheads":    {eraCosmic},
	"dock":         {eraCosmic},
	"relay":        {eraCosmic},
	"relays":       {eraCosmic},
	"manifest":     {eraCosmic},
	"freighter":    {eraCosmic},
	// The Fusion Age is in the digital bucket and its harbinger is the plant's
	// own safety system, so reactors are at home there too.
	"reactor": {eraDigital, eraCosmic},
	"airlock": {eraCosmic},

	// --- spans: at home in the ages that have ground, nowhere else -----------
	"road":    {eraAncient, eraFeudal, eraIndustrial, eraDigital},
	"valley":  {eraAncient, eraFeudal, eraIndustrial},
	"column":  {eraAncient, eraFeudal, eraIndustrial},
	"columns": {eraAncient, eraFeudal, eraIndustrial},
	"smoke":   {eraAncient, eraFeudal, eraIndustrial},
	"harvest": {eraAncient, eraFeudal},
	"yard":    {eraFeudal, eraIndustrial},

	// --- spans: the domestic furniture of a town with lanes and a washing
	// line. Added after a corpus review found the Transcendent Age narrated by
	// a baker with a shovel, dogs under the wall and washing left out: none of
	// these words was marked, so the ungated pool quietly filled up with them.
	"dog":      {eraAncient, eraFeudal, eraIndustrial},
	"dogs":     {eraAncient, eraFeudal, eraIndustrial},
	"leather":  {eraAncient, eraFeudal, eraIndustrial},
	"lane":     {eraAncient, eraFeudal, eraIndustrial},
	"lanes":    {eraAncient, eraFeudal, eraIndustrial},
	"washing":  {eraAncient, eraFeudal, eraIndustrial},
	"baker":    {eraAncient, eraFeudal, eraIndustrial},
	"bucket":   {eraAncient, eraFeudal, eraIndustrial},
	"shovel":   {eraAncient, eraFeudal, eraIndustrial},
	"chalk":    {eraAncient, eraFeudal, eraIndustrial},
	"chalked":  {eraAncient, eraFeudal, eraIndustrial},
	"roof":     {eraAncient, eraFeudal, eraIndustrial},
	"roofs":    {eraAncient, eraFeudal, eraIndustrial},
	"draught":  {eraAncient, eraFeudal, eraIndustrial},
	"loft":     {eraAncient, eraFeudal, eraIndustrial},
	"animals":  {eraAncient, eraFeudal, eraIndustrial},
	"wolf":     {eraAncient, eraFeudal, eraIndustrial},
	"sacks":    {eraAncient, eraFeudal, eraIndustrial},
	"notches":  {eraAncient, eraFeudal, eraIndustrial},
	"graves":   {eraAncient, eraFeudal, eraIndustrial},
	"shelter":  {eraAncient, eraFeudal, eraIndustrial},
	"blankets": {eraAncient, eraFeudal, eraIndustrial},
	"spade":    {eraAncient, eraFeudal, eraIndustrial},
	"barrows":  {eraAncient, eraFeudal, eraIndustrial},
	"barrow":   {eraAncient, eraFeudal, eraIndustrial},
	"mud":      {eraAncient, eraFeudal, eraIndustrial},

	// --- spans: pre-modern set dressing. A second corpus review of the late
	// ages still found a medieval village under the uplinks: door bars, lamps,
	// servants, a council sitting before it was light, a map case. Anything that
	// lights, bars, serves or keeps time by the sun stops at the Modern Age.
	// lateSetDressing in late_test.go pins that these stay marked.
	"bar":                 {eraAncient, eraFeudal, eraIndustrial},
	"barred":              {eraAncient, eraFeudal, eraIndustrial},
	"lamp":                {eraAncient, eraFeudal, eraIndustrial},
	"lamps":               {eraAncient, eraFeudal, eraIndustrial},
	"lantern":             {eraAncient, eraFeudal, eraIndustrial},
	"lanterns":            {eraAncient, eraFeudal, eraIndustrial},
	"candle":              {eraAncient, eraFeudal, eraIndustrial},
	"candles":             {eraAncient, eraFeudal, eraIndustrial},
	"torch":               {eraAncient, eraFeudal, eraIndustrial},
	"torches":             {eraAncient, eraFeudal, eraIndustrial},
	"servant":             {eraAncient, eraFeudal, eraIndustrial},
	"servants":            {eraAncient, eraFeudal, eraIndustrial},
	"household":           {eraAncient, eraFeudal, eraIndustrial},
	"council":             {eraAncient, eraFeudal, eraIndustrial},
	"the well":            {eraAncient, eraFeudal, eraIndustrial},
	"map case":            {eraAncient, eraFeudal, eraIndustrial},
	"first light":         {eraAncient, eraFeudal, eraIndustrial},
	"dawn":                {eraAncient, eraFeudal, eraIndustrial},
	"dusk":                {eraAncient, eraFeudal, eraIndustrial},
	"before it was light": {eraAncient, eraFeudal, eraIndustrial},
	"by dark":             {eraAncient, eraFeudal, eraIndustrial},
	"after dark":          {eraAncient, eraFeudal, eraIndustrial},
	"outer houses":        {eraAncient, eraFeudal, eraIndustrial},
	"runner":              {eraAncient, eraFeudal, eraIndustrial},
	"salt":                {eraAncient, eraFeudal, eraIndustrial},
	"chest":               {eraAncient, eraFeudal, eraIndustrial},
	"chests":              {eraAncient, eraFeudal, eraIndustrial},
	"basket":              {eraAncient, eraFeudal, eraIndustrial},
	"baskets":             {eraAncient, eraFeudal, eraIndustrial},
	"bench":               {eraAncient, eraFeudal, eraIndustrial},
	"plank":               {eraAncient, eraFeudal, eraIndustrial},
	"planks":              {eraAncient, eraFeudal, eraIndustrial},
	"coin":                {eraAncient, eraFeudal, eraIndustrial},
	"by the fire":         {eraAncient, eraFeudal, eraIndustrial},
	"ink":                 {eraAncient, eraFeudal, eraIndustrial},
	"water skin":          {eraAncient, eraFeudal, eraIndustrial},
	"water skins":         {eraAncient, eraFeudal, eraIndustrial},
	"walking staff":       {eraAncient, eraFeudal, eraIndustrial},
	"sharpening stone":    {eraAncient, eraFeudal, eraIndustrial},
	"cooking pot":         {eraAncient, eraFeudal, eraIndustrial},

	// Overland travel on foot: the set dressing of a party that walks there.
	// A Neon Heist does not cross a river at the second camp.
	"river":     {eraAncient, eraFeudal, eraIndustrial},
	"hills":     {eraAncient, eraFeudal, eraIndustrial},
	"ridge":     {eraAncient, eraFeudal, eraIndustrial},
	"tree line": {eraAncient, eraFeudal, eraIndustrial},
	"camp":      {eraAncient, eraFeudal, eraIndustrial},
	"crossing":  {eraAncient, eraFeudal, eraIndustrial},
	"miles":     {eraAncient, eraFeudal, eraIndustrial},
	"mile":      {eraAncient, eraFeudal, eraIndustrial},
	"far bank":  {eraAncient, eraFeudal, eraIndustrial},
	"slope":     {eraAncient, eraFeudal, eraIndustrial},
	"walk home": {eraAncient, eraFeudal, eraIndustrial},
	"walk back": {eraAncient, eraFeudal, eraIndustrial},
	"trail":     {eraAncient, eraFeudal, eraIndustrial},

	// The river that "the water" means in a walking chronicle, and the other
	// phrases of a trek. Found in the second late-age corpus review: a
	// Transcendent crew turning back at the water, a Cyberpunk team buying
	// directions off a local.
	"back at the water":     {eraAncient, eraFeudal, eraIndustrial},
	"off at the water":      {eraAncient, eraFeudal, eraIndustrial},
	"by the water":          {eraAncient, eraFeudal, eraIndustrial},
	"past the water":        {eraAncient, eraFeudal, eraIndustrial},
	"beyond the water":      {eraAncient, eraFeudal, eraIndustrial},
	"into the water":        {eraAncient, eraFeudal, eraIndustrial},
	"crossed the water":     {eraAncient, eraFeudal, eraIndustrial},
	"bend in the water":     {eraAncient, eraFeudal, eraIndustrial},
	"because of the water":  {eraAncient, eraFeudal, eraIndustrial},
	"best of the water":     {eraAncient, eraFeudal, eraIndustrial},
	"got the water wrong":   {eraAncient, eraFeudal, eraIndustrial},
	"water they were":       {eraAncient, eraFeudal, eraIndustrial},
	"near water":            {eraAncient, eraFeudal, eraIndustrial},
	"out of the rock":       {eraAncient, eraFeudal, eraIndustrial},
	"wash water":            {eraAncient, eraFeudal, eraIndustrial},
	"old marks":             {eraAncient, eraFeudal, eraIndustrial},
	"dried meat":            {eraAncient, eraFeudal, eraIndustrial},
	"wet pack":              {eraAncient, eraFeudal, eraIndustrial},
	"off a local":           {eraAncient, eraFeudal, eraIndustrial},
	"the blade":             {eraAncient, eraFeudal, eraIndustrial},
	"walked with them":      {eraAncient, eraFeudal, eraIndustrial},
	"walked along with":     {eraAncient, eraFeudal, eraIndustrial},
	"walked a long way":     {eraAncient, eraFeudal, eraIndustrial},
	"walked it":             {eraAncient, eraFeudal, eraIndustrial},
	"walked the last":       {eraAncient, eraFeudal, eraIndustrial},
	"walked past the place": {eraAncient, eraFeudal, eraIndustrial},
	"last night out":        {eraAncient, eraFeudal, eraIndustrial},
	"rained the whole way":  {eraAncient, eraFeudal, eraIndustrial},
	"rain for six":          {eraAncient, eraFeudal, eraIndustrial},
	"paces":                 {eraAncient, eraFeudal, eraIndustrial},
	"paced":                 {eraAncient, eraFeudal, eraIndustrial},
	"no moon":               {eraAncient, eraFeudal, eraIndustrial},
	"in the dirt":           {eraAncient, eraFeudal, eraIndustrial},
	"sweeps her step":       {eraAncient, eraFeudal, eraIndustrial},
	"up on the wall":        {eraAncient, eraFeudal, eraIndustrial},
	"length of string":      {eraAncient, eraFeudal, eraIndustrial},
	"barrels of oil":        {eraAncient, eraFeudal, eraIndustrial},
	"stone bowl":            {eraAncient, eraFeudal, eraIndustrial},
	"sixth night":           {eraAncient, eraFeudal, eraIndustrial},
	"blades":                {eraAncient, eraFeudal, eraIndustrial},
	"dragged":               {eraAncient, eraFeudal, eraIndustrial},
	"carried nine days":     {eraAncient, eraFeudal, eraIndustrial},
	"medicine box":          {eraAncient, eraFeudal, eraIndustrial},
	"three packs":           {eraAncient, eraFeudal, eraIndustrial},

	// The house around the people: bread on the table, a doorframe to mark,
	// a cellar to hide in, a goat in the yard. A third late-age review found
	// the Information Age onward still narrated by a kitchen with a back door,
	// because none of these was marked and the shared pools filled up with
	// them. They stop at the Modern Age like the rest of the domestic set.
	"bread":       {eraAncient, eraFeudal, eraIndustrial},
	"doorframe":   {eraAncient, eraFeudal, eraIndustrial},
	"doorframes":  {eraAncient, eraFeudal, eraIndustrial},
	"doorstep":    {eraAncient, eraFeudal, eraIndustrial},
	"back stairs": {eraAncient, eraFeudal, eraIndustrial},
	"hearth":      {eraAncient, eraFeudal, eraIndustrial},
	"firewood":    {eraAncient, eraFeudal, eraIndustrial},
	"cellar":      {eraAncient, eraFeudal, eraIndustrial},
	"pantry":      {eraAncient, eraFeudal, eraIndustrial},
	"sack":        {eraAncient, eraFeudal, eraIndustrial},
	"buckets":     {eraAncient, eraFeudal, eraIndustrial},
	"barefoot":    {eraAncient, eraFeudal, eraIndustrial},
	"horse":       {eraAncient, eraFeudal, eraIndustrial},
	"horses":      {eraAncient, eraFeudal, eraIndustrial},
	"cattle":      {eraAncient, eraFeudal, eraIndustrial},
	"goat":        {eraAncient, eraFeudal, eraIndustrial},
	"goats":       {eraAncient, eraFeudal, eraIndustrial},
}

// eraName gives a bucket a readable name for failure messages.
func eraName(e era) string {
	for _, bucket := range eraBuckets {
		if bucket.era == e {
			return bucket.suffix
		}
	}
	return "era(?)"
}

// markerRE builds the word-boundary matcher for one marker.
func markerRE(word string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
}

// atHome reports whether a marker is at home in era e.
func atHome(homes []era, e era) bool {
	for _, h := range homes {
		if h == e {
			return true
		}
	}
	return false
}

// TestNoEraBleedInAuthoredText is the static half of the guard: a lint over every
// authored string in the catalog — the sentences AND the noun banks — that
// catches an anachronism at the exact line carrying it.
func TestNoEraBleedInAuthoredText(t *testing.T) {
	prose := authoredProse()
	// Compiled once: the catalog has grown past three thousand strings, and
	// compiling each marker per string was most of this test's time.
	res := make(map[string]*regexp.Regexp, len(eraMarkers))
	for word := range eraMarkers {
		res[word] = markerRE(word)
	}
	for _, u := range prose {
		for word, homes := range eraMarkers {
			if !res[word].MatchString(u.text) {
				continue
			}
			for e := range u.eras {
				if atHome(homes, e) {
					continue
				}
				t.Errorf("era bleed: %s can fire in the %s era but %q carries the era-coded "+
					"word %q — rewrite it age-agnostic, or move it into the %s pool",
					u.where, eraName(e), strings.TrimSpace(u.text), word, eraName(homes[0]))
			}
		}
	}
	t.Logf("checked %d authored strings against %d era markers", len(prose), len(eraMarkers))
}

// TestNoEraBleedInOutput is the end-to-end half: generate a few thousand lines per
// Moment at a representative age in EVERY bucket and assert no line carries a word
// from a foreign era. This is the test that would have failed on the real space-age
// output that prompted the original fix — carts, gates, valleys and militia
// narrating an orbital strike.
//
// The Subject and the resource label are stripped before scanning, because both
// reach the output verbatim from the CALLER: "Orbital Strike" is an expedition
// name, not a fragment this package authored.
func TestNoEraBleedInOutput(t *testing.T) {
	ages := []struct {
		age string
		era era
	}{
		{"primitive_age", eraAncient},
		{"medieval_age", eraFeudal},
		{"electric_age", eraIndustrial},
		{"digital_age", eraDigital},
		{"galactic_age", eraCosmic},
	}
	tones := allTones
	kinds := []string{"", "scouting", "military", "aggressive", "mercantile"}
	subjects := []string{"", "Raid Bandit Camp", "Naval Expedition", "Orbital Strike", "Merchant Guild", "Void Reavers"}
	resources := []string{"", "food", "gold", "soldiers", "quantum_flux"}

	foreign := map[era]*regexp.Regexp{}
	for _, bucket := range eraBuckets {
		var words []string
		for word, homes := range eraMarkers {
			if !atHome(homes, bucket.era) {
				words = append(words, regexp.QuoteMeta(word))
			}
		}
		sort.Strings(words) // stable pattern, so a failure reproduces
		foreign[bucket.era] = regexp.MustCompile(`(?i)\b(?:` + strings.Join(words, "|") + `)\b`)
	}

	lines := 0
	for _, m := range Moments() {
		if isHarbinger(m) {
			// Era bleed for these is checked at every one of the 22 ages, with
			// the real roster names, by TestHarbingerLinesAreClean. Their lines
			// ignore Kind (bar the tier), resource and amount, so this matrix
			// would only repeat those draws five hundred times over.
			continue
		}
		for _, a := range ages {
			rng := rand.New(rand.NewSource(int64(m)*7717 + int64(a.era)))
			for _, tone := range tones {
				for _, kind := range kinds {
					for _, subj := range subjects {
						for _, key := range resources {
							for i := 0; i < 8; i++ {
								req := Request{
									Moment: m, Tone: tone, Age: a.age, Kind: kind,
									Subject: subj, Resource: key, Amount: 137,
								}
								line := Line(req, rng)
								lines++
								scan := line
								if subj != "" {
									scan = strings.ReplaceAll(scan, subj, " ")
								}
								if key != "" {
									scan = strings.ReplaceAll(scan, resourceLabel(key), " ")
								}
								if hit := foreign[a.era].FindString(scan); hit != "" {
									t.Fatalf("era bleed in %v at %s (%s era): %q contains %q, "+
										"which is at home in the %s era only", m, a.age,
										eraName(a.era), line, hit,
										eraName(eraMarkers[strings.ToLower(hit)][0]))
								}
							}
						}
					}
				}
			}
		}
	}
	t.Logf("fuzzed %d lines across %d moments and %d eras with no era bleed", lines, len(Moments()), len(ages))
}

// TestEveryEraHasItsOwnVoice pins the other side of the fix: it is not enough to
// scrub the anachronisms out, every bucket must have real authored prose of its
// own or the late ages just get a blander version of the same pool.
func TestEveryEraHasItsOwnVoice(t *testing.T) {
	const (
		wantEraSkeletons = 20
		wantDistinct     = 220
	)
	for _, m := range Moments() {
		if isRare(m) {
			// Their era voice is held per AGE by TestHarbingerSpeakerFit and
			// TestLateEraVoiceShare; the 2000-draw depth floor does not apply.
			continue
		}
		for _, bucket := range eraBuckets {
			age := eraAges[bucket.era]
			p := eligible(Request{Moment: m, Age: age})
			gated := 0
			for _, tpl := range p {
				if len(tpl.Eras) > 0 {
					gated++
				}
			}
			if gated < wantEraSkeletons {
				t.Errorf("%v: only %d skeletons are specific to the %s era; want >= %d so those "+
					"ages do not borrow the age-agnostic pool for their whole voice",
					m, gated, bucket.suffix, wantEraSkeletons)
			}

			rng := rand.New(rand.NewSource(int64(m)*3313 + int64(bucket.era)))
			seen := map[string]bool{}
			hitEra := 0
			for i := 0; i < 2000; i++ {
				res := Generate(Request{Moment: m, Age: age}, rng)
				seen[res.Text] = true
				if strings.Contains(res.Template, "_"+bucket.suffix+"_") {
					hitEra++
				}
			}
			if hitEra == 0 {
				t.Errorf("%v: no line in 2000 draws at %s used the %s pool", m, age, bucket.suffix)
			}
			if len(seen) < wantDistinct {
				t.Errorf("%v at %s: only %d distinct lines in 2000 draws; want >= %d", m, age, len(seen), wantDistinct)
			}
			t.Logf("%-20v %-11s %3d era skeletons, %4d distinct lines / 2000, %4d era-voiced",
				m, bucket.suffix, gated, len(seen), hitEra)
		}
	}
}

// TestSignaturesClassifyEveryLine is the end-to-end version of the signature
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

// TestToneAndKindBiasSelection proves the three shaping inputs actually change
// the prose rather than being decorative. It works off template PREDICATES rather
// than hard-coded ids, so re-authoring the catalog cannot silently disarm it.
func TestToneAndKindBiasSelection(t *testing.T) {
	reach := func(req Request, seed int64, n int) map[string]bool {
		rng := rand.New(rand.NewSource(seed))
		out := map[string]bool{}
		for i := 0; i < n; i++ {
			out[Generate(req, rng).Template] = true
		}
		return out
	}
	find := func(m Moment, pred func(tmpl) bool) string {
		for _, tpl := range templatesFor(m) {
			if pred(tpl) {
				return tpl.ID
			}
		}
		return ""
	}
	hasTone := func(want Tone) func(tmpl) bool {
		return func(tp tmpl) bool {
			for _, x := range tp.Tones {
				if x == want {
					return true
				}
			}
			return false
		}
	}
	hasKind := func(want string) func(tmpl) bool {
		return func(tp tmpl) bool {
			for _, x := range tp.Kinds {
				if x == want {
					return true
				}
			}
			return false
		}
	}
	hasEra := func(want era) func(tmpl) bool {
		return func(tp tmpl) bool {
			return len(tp.Eras) == 1 && tp.Eras[0] == want
		}
	}

	m := ExpeditionSuccess
	base := Request{Moment: m}
	neutral := reach(base, 3, 3000)
	for id := range neutral {
		for _, tpl := range templatesFor(m) {
			if tpl.ID == id && (len(tpl.Tones) > 0 || len(tpl.Kinds) > 0) {
				t.Errorf("skeleton %q is tone- or kind-gated but fired on a bare Request", id)
			}
		}
	}

	triumphID := find(m, hasTone(Triumphant))
	if triumphID == "" {
		t.Fatal("ExpeditionSuccess has no Triumphant-gated skeleton")
	}
	triumphant := base
	triumphant.Tone = Triumphant
	if !reach(triumphant, 3, 3000)[triumphID] {
		t.Errorf("the Triumphant skeleton %q never fired under Triumphant", triumphID)
	}

	scoutID := find(m, hasKind("scouting"))
	if scoutID == "" {
		t.Fatal("ExpeditionSuccess has no scouting-gated skeleton")
	}
	scouting := base
	scouting.Kind = "SCOUTING" // case-insensitive
	if !reach(scouting, 3, 3000)[scoutID] {
		t.Errorf("the scouting skeleton %q never fired for Kind=scouting", scoutID)
	}

	ancientID := find(m, hasEra(eraAncient))
	cosmicID := find(m, hasEra(eraCosmic))
	if ancientID == "" || cosmicID == "" {
		t.Fatal("ExpeditionSuccess is missing an ancient or a cosmic pool")
	}
	early := base
	early.Age = "primitive_age"
	earlyReach := reach(early, 3, 3000)
	if earlyReach[cosmicID] {
		t.Error("a cosmic-era skeleton fired in the primitive age")
	}
	if !earlyReach[ancientID] {
		t.Error("the ancient-era skeleton never fired in the primitive age")
	}
	late := base
	late.Age = "transcendent_age"
	lateReach := reach(late, 3, 3000)
	if !lateReach[cosmicID] {
		t.Error("the cosmic-era skeleton never fired in the transcendent age")
	}
	if lateReach[ancientID] {
		t.Error("an early-era skeleton fired in the transcendent age")
	}
	unset := reach(base, 3, 3000)
	if !unset[cosmicID] || !unset[ancientID] {
		t.Error("an unset Age should not exclude era-specific skeletons")
	}
}

// --- read it as a stream ---------------------------------------------------

// TestReadTheStream prints CONSECUTIVE lines exactly as a player meets them, in
// one age, off one rng. No assertion here can judge prose; the point is that a
// reviewer reads thirty lines in a row and sees whether the catalog holds up.
// Run with -v.
//
// What to look for, in order of severity: two adjacent lines about the same
// thing, a run of lines the same LENGTH, a false-contrast construction, three wry
// observations in a row with no flat line between them, or a line that just says
// the thing succeeded again.
func TestReadTheStream(t *testing.T) {
	streams := []struct {
		label string
		n     int
		req   Request
	}{
		{"ExpeditionSuccess — medieval age, military, gold", 40, Request{
			Moment: ExpeditionSuccess, Tone: Triumphant, Age: "medieval_age",
			Kind: "military", Subject: "Siege Enemy Castle", Resource: "gold", Amount: 340,
		}},
		{"ExpeditionFailure — iron age, scouting", 20, Request{
			Moment: ExpeditionFailure, Age: "iron_age", Kind: "scouting",
			Subject: "Scout Nearby Ruins", Resource: "iron", Amount: 30,
		}},
		{"WarRaid — galactic age, aggressive", 12, Request{
			Moment: WarRaid, Tone: Grim, Age: "galactic_age",
			Kind: "aggressive", Subject: "Void Reavers", Resource: "antimatter",
		}},
		{"EncounterStandoff — cyberpunk age", 12, Request{
			Moment: EncounterStandoff, Age: "cyberpunk_age",
			Kind: "aggressive", Subject: "Shadow Syndicate",
		}},
		{"EncounterAtCapacity — classical age, mercantile", 12, Request{
			Moment: EncounterAtCapacity, Tone: Wry, Age: "classical_age",
			Kind: "mercantile", Subject: "Merchant Guild",
		}},
		{"ExpeditionSuccess — primitive age, zero request otherwise", 12, Request{
			Moment: ExpeditionSuccess, Age: "primitive_age",
		}},
		{"ExpeditionSuccess — transcendent age", 12, Request{
			Moment: ExpeditionSuccess, Age: "transcendent_age",
		}},
	}
	for _, s := range streams {
		rng := rand.New(rand.NewSource(20260809))
		st := NewStream()
		t.Logf("--- %s ---", s.label)
		for i := 0; i < s.n; i++ {
			t.Logf("  %2d  %s", i+1, st.Line(s.req, rng))
		}
	}
}
