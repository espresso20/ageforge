package flavor

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/espresso20/ageforge/config"
)

// generator_test.go holds the end-to-end contract tests for the generator as the
// GAME uses it: every one of the 22 ages, every Moment, the tones and kinds the
// engine adapters actually send, and the Stream the engine logs through. The
// catalog-authoring lints (burstiness, register quotas, AI-tell shapes) live in
// flavor_test.go; the tests here are about what a player can be shown.

// gameKinds are the Request.Kind values the engine adapters send: expedition
// categories for the expedition Moments, faction personalities for the rest.
var gameKinds = map[Moment][]string{
	ExpeditionSuccess:   {"scouting", "military"},
	ExpeditionFailure:   {"scouting", "military"},
	EncounterStandoff:   {"peaceful", "aggressive", "mercantile", "isolationist"},
	EncounterAtCapacity: {"peaceful", "aggressive", "mercantile", "isolationist"},
	WarRaid:             {"peaceful", "aggressive", "mercantile", "isolationist"},
	// Harbinger Moments: the warning takes the risk tier, a fulfilled invite
	// may carry the false-prophet kind, and the rest take no kind at all.
	HarbingerArrival:     {""},
	HarbingerWarning:     {TierHigh, TierMedium, TierLow, TierNone},
	HarbingerAppeased:    {""},
	HarbingerBraced:      {""},
	HarbingerVindicated:  {""},
	HarbingerSpared:      {""},
	HarbingerDiscredited: {""},
	HarbingerInvited:     {""},
	HarbingerFulfilled:   {"", KindFalseProphet},
	// RunEnding takes no kind; its Subject is the harbinger in the Cosmic Era.
	RunEnding: {""},
}

// gameRequests returns the requests the engine can realistically build for a
// Moment at an age: every tone crossed with the adapter's kinds, with the subject
// and resource shapes that call site uses, plus the bare zero-ish request.
func gameRequests(m Moment, age string) []Request {
	out := []Request{{Moment: m, Age: age}}
	for _, tone := range allTones {
		for _, kind := range gameKinds[m] {
			switch m {
			case ExpeditionSuccess, ExpeditionFailure:
				for _, subj := range []string{"Scout Party", "Raid Bandit Camp", "Quantum Incursion"} {
					out = append(out,
						Request{Moment: m, Tone: tone, Age: age, Kind: kind, Subject: subj},
						Request{Moment: m, Tone: tone, Age: age, Kind: kind, Subject: subj, Resource: "gold", Amount: 340},
						Request{Moment: m, Tone: tone, Age: age, Kind: kind, Subject: subj, Resource: "soldiers", Amount: 1},
						Request{Moment: m, Tone: tone, Age: age, Kind: kind, Subject: subj, Resource: "dark_matter_crystals", Amount: 12},
					)
				}
			case HarbingerArrival, HarbingerWarning, HarbingerAppeased, HarbingerBraced,
				HarbingerVindicated, HarbingerSpared, HarbingerDiscredited,
				HarbingerInvited, HarbingerFulfilled, RunEnding:
				for _, subj := range []string{gameSubject(m, age), ""} {
					out = append(out, Request{Moment: m, Tone: tone, Age: age, Kind: kind, Subject: subj})
				}
			case WarRaid:
				for _, subj := range []string{"Merchant Guild", "Void Reavers"} {
					for _, res := range []string{"food", "gold", "nanobots"} {
						out = append(out, Request{Moment: m, Tone: tone, Age: age, Kind: kind, Subject: subj, Resource: res})
					}
				}
			default:
				for _, subj := range []string{"Merchant Guild", "Void Reavers", "Riverlands Tribes"} {
					out = append(out, Request{Moment: m, Tone: tone, Age: age, Kind: kind, Subject: subj})
				}
			}
		}
	}
	return out
}

// maxLineRunes caps a finished line for the log panel it lands in. A raid line is
// appended to a ~50-rune mechanical line, and an encounter line follows a
// "✖ <Civ Name>: " prefix, so a flavour line past this reads as a paragraph in a
// log that is meant to be scanned. The authored ceiling is 45 words (TestBurstiness);
// this is the same promise measured after substitution.
const maxLineRunes = 260

var (
	// doubledPunct catches two pieces of punctuation with nothing between them
	// that a person would not type: ",," / ".," / ";." / ". ." and friends.
	doubledPunct = regexp.MustCompile(`[,;:]\s*[,;:.!?]|[.!?]\s*[,;:]|\.\s*\.|!!|\?\?`)
	// spaceBeforePunct catches " ," / " ." left behind by an empty substitution.
	spaceBeforePunct = regexp.MustCompile(`\s[,;:.!?]`)
	// badArticleA catches "a" before a vowel sound the slot supplied ("a old
	// lamp"); the exceptions are the consonant-sound vowel spellings English has.
	badArticleA = regexp.MustCompile(`\b[Aa] ([aeiouAEIOU]\w*)`)
	// badArticleAn catches "an" before a consonant ("an rope").
	badArticleAn = regexp.MustCompile(`\b[Aa]n ([b-df-gj-np-tv-zB-DF-GJ-NP-TV-Z]\w*)`)
	// sentenceBreak finds each sentence boundary inside a line so every sentence,
	// not just the first, can be held to the capital-letter rule.
	sentenceBreak = regexp.MustCompile(`[.!?]["”']?\s+(\S)`)
	// consonantVowel are vowel-initial words that take "a" ("a unit", "a one").
	consonantVowel = regexp.MustCompile(`(?i)^(u[nst]i|use|usu|one|once|eu|ur[ai])`)
)

// hygiene returns a description of the first thing wrong with a finished line, or
// "" if it is clean. It is the whole output contract in one place: markup-free,
// placeholder-free, trimmed, single-spaced, capitalised at every sentence start,
// terminally punctuated, no doubled punctuation, grammatical articles, no
// stuttered word, and short enough for the log.
func hygiene(line string) string {
	switch {
	case line == "":
		return "empty"
	case strings.ContainsAny(line, "[]%"):
		return "markup or printf directive"
	case strings.ContainsAny(line, "{}~"):
		return "unfilled placeholder or slot"
	case line != strings.TrimSpace(line):
		return "leading or trailing whitespace"
	case strings.Contains(line, "  "):
		return "double space"
	case utf8.RuneCountInString(line) > maxLineRunes:
		return fmt.Sprintf("%d runes, over the %d cap", utf8.RuneCountInString(line), maxLineRunes)
	}
	first, _ := utf8.DecodeRuneInString(line)
	if !unicode.IsUpper(first) && !unicode.IsDigit(first) {
		return "does not start with a capital"
	}
	last, _ := utf8.DecodeLastRuneInString(line)
	if !strings.ContainsRune(".!?", last) {
		return "no terminal punctuation"
	}
	if m := doubledPunct.FindString(line); m != "" {
		return fmt.Sprintf("doubled punctuation %q", m)
	}
	if m := spaceBeforePunct.FindString(line); m != "" {
		return fmt.Sprintf("space before punctuation %q", m)
	}
	for _, sm := range sentenceBreak.FindAllStringSubmatch(line, -1) {
		r, _ := utf8.DecodeRuneInString(sm[1])
		if !unicode.IsUpper(r) && !unicode.IsDigit(r) && !strings.ContainsRune(`"“'`, r) {
			return fmt.Sprintf("sentence starting %q is not capitalised", sm[1])
		}
	}
	for _, sm := range badArticleA.FindAllStringSubmatch(line, -1) {
		if !consonantVowel.MatchString(sm[1]) {
			return fmt.Sprintf("article %q", sm[0])
		}
	}
	if m := badArticleAn.FindString(line); m != "" && !strings.HasPrefix(strings.ToLower(m[3:]), "hour") {
		return fmt.Sprintf("article %q", m)
	}
	words := strings.Fields(strings.ToLower(line))
	for i := 1; i < len(words); i++ {
		if words[i] == words[i-1] && words[i] != "that" && words[i] != "had" {
			return fmt.Sprintf("stuttered word %q", words[i])
		}
	}
	return ""
}

// TestHygieneCatchesKnownDefects tests the tester: a hygiene check that passes
// everything proves nothing. Each case is a defect the generator could plausibly
// produce, and the first is the one this suite actually caught in the catalog
// ("the second ~" drawing "second lamp" from the kit bank).
func TestHygieneCatchesKnownDefects(t *testing.T) {
	bad := []string{
		"They lost the second second lamp on the first night.",
		"",
		" Leading space.",
		"Trailing space. ",
		"Two  spaces.",
		"lowercase start.",
		"No terminal punctuation",
		"Doubled,, comma.",
		"Stray comma before stop,.",
		"Space before stop .",
		"First sentence. second sentence is lowercase.",
		"They left a old lamp behind.",
		"They left an rope behind.",
		"An {res} placeholder.",
		"A ~ slot.",
		"A [red]tag[-] leaked.",
		"A 12% leak.",
		strings.Repeat("Word ", 60) + "end.",
	}
	for _, s := range bad {
		if hygiene(s) == "" {
			t.Errorf("hygiene passed a defective line: %q", s)
		}
	}
	good := []string{
		"The dog came back fat.",
		"Nine in, eleven counted.",
		"Two did well. The rest brought maps.",
		"A unit of twelve walked in behind them.",
		"They waited an hour at the crossing.",
		"12 soldiers are asleep in the barn.",
		"He said that that was the end of it.",
	}
	for _, s := range good {
		if p := hygiene(s); p != "" {
			t.Errorf("hygiene rejected a clean line %q: %s", s, p)
		}
	}
}

// foreignMarkers returns a regexp matching every era-coded word that is NOT at
// home in e, built from the same eraMarkers table the authoring lint uses.
func foreignMarkers(e era) *regexp.Regexp {
	var words []string
	for word, homes := range eraMarkers {
		if !atHome(homes, e) {
			words = append(words, regexp.QuoteMeta(word))
		}
	}
	sort.Strings(words)
	return regexp.MustCompile(`(?i)\b(?:` + strings.Join(words, "|") + `)\b`)
}

// callerText strips the parts of a line that came from the CALLER rather than
// from the catalog — the subject and the resource label — so an era scan does not
// blame the generator for an expedition called "Orbital Strike".
func callerText(line string, req Request) string {
	if req.Subject != "" {
		line = strings.ReplaceAll(line, req.Subject, " ")
	}
	if req.Resource != "" {
		line = strings.ReplaceAll(line, resourceLabel(req.Resource), " ")
		line = strings.ReplaceAll(line, singularLabel(req.Resource), " ")
	}
	return line
}

// singularLabel is the n == 1 form of a count-noun resource label, which the
// quantity frame can emit ("1 soldier"). Mass nouns return their label.
func singularLabel(key string) string {
	return strings.TrimSpace(phraseQuantity(key, 1)[1:])
}

// TestEveryAgeEveryMomentYieldsCleanLines is the coverage matrix the task is
// named for: all 22 ages x every registered Moment x every tone x every kind the
// engine sends, each drawn through a Stream many times, and every single line has
// to be non-empty, pass the full hygiene contract, and carry no word from a
// foreign era. The era is taken from eraOf, so an age that moves bucket in
// config moves the expectation with it.
func TestEveryAgeEveryMomentYieldsCleanLines(t *testing.T) {
	ages := config.AgeOrder()
	if len(ages) != 22 {
		t.Fatalf("config.AgeOrder() has %d ages; this matrix was written for 22 — "+
			"check eraOf's bucket boundaries before updating the count", len(ages))
	}
	draws := 6
	if testing.Short() {
		draws = 2
	}
	lines, longest := 0, ""
	for _, age := range ages {
		e, bounded := eraOf(age)
		if !bounded {
			t.Fatalf("age %q is not bounded to an era", age)
		}
		foreign := foreignMarkers(e)
		for _, m := range Moments() {
			if isHarbinger(m) {
				continue // TestHarbingerLinesAreClean runs this matrix for them, per tier
			}
			rng := rand.New(rand.NewSource(int64(len(age))*131 + int64(m)))
			st := NewStream()
			for _, req := range gameRequests(m, age) {
				for i := 0; i < draws; i++ {
					line := st.Line(req, rng)
					lines++
					if problem := hygiene(line); problem != "" {
						t.Fatalf("%v at %s (%+v): %s\n    %q", m, age, req, problem, line)
					}
					if hit := foreign.FindString(callerText(line, req)); hit != "" {
						t.Fatalf("era bleed: %v at %s (%s era) produced %q, which carries %q",
							m, age, eraName(e), line, hit)
					}
					if len(line) > len(longest) {
						longest = line
					}
				}
			}
		}
	}
	t.Logf("%d lines across %d ages x %d moments, all clean; longest is %d runes:\n    %q",
		lines, len(ages), len(Moments()), utf8.RuneCountInString(longest), longest)
}

// TestHygieneOverFuzzedRequests pushes the hygiene contract past realistic
// requests into hostile ones: unknown ages and kinds, every resource key config
// ships (mass and count), amounts that round to zero or one or overflow, and
// subjects of every shape. Nothing a caller can hand in may produce a dirty line.
func TestHygieneOverFuzzedRequests(t *testing.T) {
	var keys []string
	for _, def := range config.BaseResources() {
		keys = append(keys, def.Key)
	}
	amounts := []float64{0, 0.4, 0.6, 1, 2, 1e9, -5}
	subjects := append([]string{""}, realExpeditionNames...)
	subjects = append(subjects, realFactionNames...)
	rng := rand.New(rand.NewSource(424242))
	ages := append(config.AgeOrder(), "", "not_an_age")
	n := 0
	for _, m := range Moments() {
		// The harbinger Moments read Age, Subject and (for two of them) Kind and
		// nothing else, so a quarter of the draws covers their hostile surface.
		draws := 20000
		if isRare(m) {
			draws = 5000
		}
		for i := 0; i < draws; i++ {
			req := Request{
				Moment:   m,
				Tone:     allTones[rng.Intn(len(allTones))],
				Age:      ages[rng.Intn(len(ages))],
				Kind:     []string{"", "scouting", "military", "peaceful", "aggressive", "mercantile", "isolationist", "Nonsense"}[rng.Intn(8)],
				Subject:  subjects[rng.Intn(len(subjects))],
				Resource: keys[rng.Intn(len(keys))],
				Amount:   amounts[rng.Intn(len(amounts))],
				Count:    rng.Intn(3),
				Ticks:    rng.Intn(3) * 450,
			}
			line := Line(req, rng)
			n++
			if problem := hygiene(line); problem != "" {
				t.Fatalf("%v (%+v): %s\n    %q", m, req, problem, line)
			}
		}
	}
	t.Logf("%d fuzzed lines, all clean", n)
}

// TestSeedsReproduceAndDiverge pins determinism for EVERY Moment in every era, on
// both paths: the raw generator and a Stream. Same seed, same request, same
// history ⇒ the identical run of lines. Different seed ⇒ a materially different
// run (not merely one line off), which proves the rng is driving the choice.
func TestSeedsReproduceAndDiverge(t *testing.T) {
	const n = 300
	run := func(req Request, seed int64, stream bool) []string {
		rng := rand.New(rand.NewSource(seed))
		st := NewStream()
		out := make([]string, n)
		for i := range out {
			if stream {
				out[i] = st.Line(req, rng)
			} else {
				out[i] = Line(req, rng)
			}
		}
		return out
	}
	for _, m := range Moments() {
		for _, bucket := range eraBuckets {
			req := Request{Moment: m, Age: eraAges[bucket.era], Kind: gameKinds[m][0], Subject: "Void Reavers", Resource: "gold", Amount: 90}
			for _, stream := range []bool{false, true} {
				a, b, c := run(req, 77, stream), run(req, 77, stream), run(req, 78, stream)
				differ := 0
				for i := range a {
					if a[i] != b[i] {
						t.Fatalf("%v %s stream=%v: line %d differs under the same seed:\n  %q\n  %q",
							m, bucket.suffix, stream, i, a[i], b[i])
					}
					if a[i] != c[i] {
						differ++
					}
				}
				if differ < n*8/10 {
					t.Errorf("%v %s stream=%v: seeds 77 and 78 agree on %d of %d lines; want them to diverge",
						m, bucket.suffix, stream, n-differ, n)
				}
			}
		}
	}
}

// TestStreamNeverRepeatsInsideItsWindow is the repeat-avoidance guarantee as a
// player would measure it. For every Moment at every age, 10,000 consecutive
// lines through one Stream (the engine's path) must contain:
//
//   - ZERO skeleton repeats inside the last streamMemory lines, and so zero
//     identical lines inside that window. This is a hard promise, not a rate:
//     every era-eligible pool is wider than the window (TestSkeletonFloors), and
//     the Stream falls back to a direct pick from the unseen part of the pool
//     when its retries run out.
//   - adjacent same-topic lines at most 2% of the time (the topic ring is a soft
//     preference, so this is a rate).
//   - near-total coverage: at least 95% of the skeletons the age can reach
//     actually show up, so avoidance is not bought by starving part of the pool.
//   - a distinct-line floor: at least as many distinct finished lines as there are
//     reachable skeletons, i.e. slots add variety on top rather than replacing it.
func TestStreamNeverRepeatsInsideItsWindow(t *testing.T) {
	// 10,000 lines at the representative age of each era bucket, 2,000 at every
	// other age: the window property is per-draw, so the extra depth buys little
	// once each bucket has had its long run.
	deep := map[string]bool{}
	for _, a := range eraAges {
		deep[a] = true
	}
	topics := topicIndex()
	for _, age := range config.AgeOrder() {
		for _, m := range Moments() {
			if !harbReachable(m, age) {
				continue // a false prophet cannot be discredited where none exist
			}
			// Harbinger pools are a few dozen sentences; 2,000 draws already
			// cycle each one about sixty times, so the deep run adds only time.
			draws := 2000
			if deep[age] && !testing.Short() && !isRare(m) {
				draws = 10000
			}
			req := Request{Moment: m, Age: age, Kind: gameKinds[m][0], Tone: Neutral, Subject: gameSubject(m, age)}
			pool := eligible(req)
			if len(pool) <= streamMemory {
				t.Fatalf("%v at %s: only %d eligible skeletons, not wider than the %d-line window",
					m, age, len(pool), streamMemory)
			}
			rng := rand.New(rand.NewSource(int64(m)*1_000_003 + int64(len(age))))
			st := NewStream()
			window := make([]string, 0, streamMemory)
			lastLine := make([]string, 0, streamMemory)
			seenSk := map[string]bool{}
			seenLine := map[string]bool{}
			adjacentTopic, prevTopic := 0, ""
			for i := 0; i < draws; i++ {
				res := st.Generate(req, rng)
				for j, id := range window {
					if id == res.Template {
						t.Fatalf("%v at %s: skeleton %s repeated after %d lines (window %d), draw %d:\n    %q",
							m, age, id, len(window)-j, streamMemory, i, res.Text)
					}
					if lastLine[j] == res.Text {
						t.Fatalf("%v at %s: identical line inside the window at draw %d: %q", m, age, i, res.Text)
					}
				}
				if len(window) == streamMemory {
					window, lastLine = window[1:], lastLine[1:]
				}
				window, lastLine = append(window, res.Template), append(lastLine, res.Text)
				seenSk[res.Template] = true
				seenLine[res.Text] = true
				if tp := topics[res.Template]; tp != "" && tp == prevTopic {
					adjacentTopic++
				} else {
					prevTopic = tp
				}
			}
			if rate := float64(adjacentTopic) / float64(draws); rate > 0.02 {
				t.Errorf("%v at %s: %.1f%% of lines share a topic with the line before; want <= 2%%",
					m, age, rate*100)
			}
			if cov := float64(len(seenSk)) / float64(len(pool)); cov < 0.95 {
				t.Errorf("%v at %s: only %d of %d reachable skeletons appeared in %d lines (%.0f%%); want >= 95%%",
					m, age, len(seenSk), len(pool), draws, cov*100)
			}
			if len(seenLine) < len(pool) {
				t.Errorf("%v at %s: %d distinct lines from a pool of %d skeletons; want at least one line per skeleton",
					m, age, len(seenLine), len(pool))
			}
		}
	}
}

// TestSharedStreamAcrossMoments mirrors the engine exactly: ONE Stream for the
// whole log, fed expedition, encounter and raid requests in an interleaved
// order. The skeleton window is shared, so no sentence may come back inside it
// whichever Moment asked for it.
func TestSharedStreamAcrossMoments(t *testing.T) {
	rng := rand.New(rand.NewSource(9001))
	st := NewStream()
	var window []string
	for i := 0; i < 20000; i++ {
		m := Moments()[rng.Intn(len(Moments()))]
		age := config.AgeOrder()[(i/1000)%22]
		kinds := gameKinds[m]
		res := st.Generate(Request{Moment: m, Age: age, Kind: kinds[rng.Intn(len(kinds))], Subject: "Merchant Guild"}, rng)
		for _, id := range window {
			if id == res.Template {
				t.Fatalf("draw %d: skeleton %s came back inside the shared window", i, id)
			}
		}
		window = append(window, res.Template)
		if len(window) > streamMemory {
			window = window[1:]
		}
	}
}

// TestNarrowPoolDegradesGracefully pins the other half of the Stream contract: a
// pool NARROWER than the window must still produce a line every time rather than
// loop or go silent. No real age hits this; a future Moment with a thin catalog
// or an over-constrained request might.
func TestNarrowPoolDegradesGracefully(t *testing.T) {
	// A Wry, need-heavy request narrows the pool less than you might hope, so
	// build the narrow case directly: prefill the ring with every eligible id.
	req := Request{Moment: WarRaid, Age: "stone_age", Kind: "aggressive"}
	pool := eligible(req)
	st := NewStream()
	st.recent = make([]string, len(pool)+1)
	for i, tpl := range pool {
		st.recent[i] = tpl.ID
	}
	rng := rand.New(rand.NewSource(3))
	for i := 0; i < 50; i++ {
		if line := st.Line(req, rng); line == "" {
			t.Fatalf("draw %d: exhausted Stream returned an empty line", i)
		}
	}
	var nilStream *Stream
	if nilStream.Line(req, rng) == "" {
		t.Fatal("nil Stream returned an empty line; it should fall back to Generate")
	}
}

// TestConcurrentUse is the lock-safety test; run it with -race. The catalog, the
// banks and the lazily built indexes (templateCache, topicIndex, ageIndex) are
// shared read-only state, and first use of each races with every other first
// use. Generate itself is pure, so many goroutines with their own rngs and their
// own Streams must be able to draw at once. A Stream is NOT safe for concurrent
// use — the engine only touches its one under the write lock — so the shared
// Stream here is driven under a mutex exactly as the engine drives it.
func TestConcurrentUse(t *testing.T) {
	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		shared = NewStream()
		srng   = rand.New(rand.NewSource(1))
	)
	errs := make(chan string, 64)
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(int64(g)))
			own := NewStream()
			ages := config.AgeOrder()
			for i := 0; i < 500; i++ {
				m := Moments()[(g+i)%len(Moments())]
				req := Request{Moment: m, Age: ages[(g*7+i)%len(ages)], Kind: gameKinds[m][i%len(gameKinds[m])]}
				if Line(req, rng) == "" || own.Line(req, rng) == "" {
					errs <- fmt.Sprintf("goroutine %d draw %d: empty line", g, i)
					return
				}
				_ = Signatures(m)
				mu.Lock()
				line := shared.Line(req, srng)
				mu.Unlock()
				if line == "" {
					errs <- fmt.Sprintf("goroutine %d draw %d: empty line from shared Stream", g, i)
					return
				}
			}
		}(g)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}

// BenchmarkStreamLine is the per-line cost of the engine's path. It used to
// rebuild the whole catalog and the age table on every draw.
func BenchmarkStreamLine(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	st := NewStream()
	req := Request{Moment: ExpeditionSuccess, Age: "medieval_age", Kind: "military", Subject: "Scout Party", Resource: "gold", Amount: 50}
	for i := 0; i < b.N; i++ {
		_ = st.Line(req, rng)
	}
}

// --- the humanizer pass, frozen into tests ----------------------------------

// TestHouseTics holds the catalog's OWN machine tells to a ceiling. The shapes
// TestNoAITells bans are the generic ones; a bulk read of a generated corpus
// found a second layer that is specific to this catalog, where one writer
// reached for the same device until it became a signature:
//
//   - "eleven" as the default funny number (28 lines before the pass),
//   - the "asked twice ... both times" repetition gag (54),
//   - "very" and "quietly" as padding ("has gone very quiet", five times),
//   - "nobody wants to be the one who",
//   - anonymous somebody/nobody actors in about one line in five.
//
// Each ceiling sits a little above the post-pass count, so a new line can use
// the device and a new pool cannot bring the habit back.
func TestHouseTics(t *testing.T) {
	type tic struct {
		name string
		re   *regexp.Regexp
		max  int
	}
	tics := []tic{
		{"the number eleven", regexp.MustCompile(`(?i)\beleven\b`), 9},
		{"the twice gag", regexp.MustCompile(`(?i)\btwice\b`), 26},
		{"very", regexp.MustCompile(`(?i)\bvery\b`), 10},
		{"quietly", regexp.MustCompile(`(?i)\bquietly\b`), 0},
		{"wants to be the one", regexp.MustCompile(`(?i)\bwants? to be the (one|person)\b`), 2},
		{"gone quiet", regexp.MustCompile(`(?i)\b(gone|went|been) (very )?quiet\b`), 2},
		{"asked N times", regexp.MustCompile(`(?i)\b(asked|told|reminded|answered|thanked)\b[^.]{0,50}\b(twice|three times|four times|five times|each time|both times)\b`), 5},
		// Found by the second, late-age corpus review: "has asked" had become
		// every late-era character's reaction, a committee was the punchline
		// of six lines, and the restored-from-backup dread was losing its
		// force by the ninth use.
		{"has asked", regexp.MustCompile(`(?i)\b(has|have|had) asked\b`), 24},
		{"committee", regexp.MustCompile(`(?i)\b(committee|working group)\b`), 3},
		{"restored from backup", regexp.MustCompile(`(?i)\b(restored|restoration|backup of|from backup|backups were)\b`), 8},
		{"the ship's mind", regexp.MustCompile(`(?i)\bship's mind\b`), 5},
		{"the same", regexp.MustCompile(`(?i)\bthe same\b`), 88},
		// Found by the humanizer pass over the harbinger Moments. The first
		// draft leaned on an ironic reversal ("the people who had laughed
		// were quiet now") in 39 sentences, six times the rate of the rest of
		// the catalog; ended long lines on an "as if" simile; filled time with
		// "for a long time"; announced things in the passive ("It has been
		// proclaimed"); and turned observations into sayings ("there is a kind
		// of relief in", "it is one thing to").
		{"the people who had", regexp.MustCompile(`(?i)\b(people|those|ones|men|women|man|woman) who (had|have)\b`), 26},
		{"ends on as if", regexp.MustCompile(`(?i)\bas (if|though)\b[^,]*$`), 17},
		{"for a long time", regexp.MustCompile(`(?i)\bfor a (long time|while)\b`), 10},
		{"a great many", regexp.MustCompile(`(?i)\ba great (many|deal)\b`), 16},
		{"it has been proclaimed", regexp.MustCompile(`(?i)^it (has been|is|was) (agreed|ordered|proclaimed|decreed|announced)`), 6},
		{"a kind of", regexp.MustCompile(`(?i)\b(there is|there was) a (strange )?kind of\b|\bit is one thing to\b|\bis a kind of\b`), 0},
	}
	anon := regexp.MustCompile(`(?i)\b(somebody|someone|nobody|no one)\b`)
	const maxAnonShare = 0.15

	counts := make([]int, len(tics))
	total, anonLines := 0, 0
	for _, m := range Moments() {
		for _, tpl := range templatesFor(m) {
			text := authoredText(tpl)
			total++
			for i, x := range tics {
				if x.re.MatchString(text) {
					counts[i]++
				}
			}
			if anon.MatchString(text) {
				anonLines++
			}
		}
	}
	for i, x := range tics {
		if counts[i] > x.max {
			t.Errorf("%q appears in %d skeletons; the ceiling is %d", x.name, counts[i], x.max)
		}
		t.Logf("%-22s %3d (ceiling %d)", x.name, counts[i], x.max)
	}
	if share := float64(anonLines) / float64(total); share > maxAnonShare {
		t.Errorf("%.1f%% of skeletons hand the action to somebody/nobody; the ceiling is %.0f%% — name a person",
			share*100, maxAnonShare*100)
	}
	t.Logf("anonymous actor share %.1f%% of %d skeletons", float64(anonLines)/float64(total)*100, total)
}

// TestNoNearDuplicateSkeletons catches the same sentence written twice with a
// noun swapped: "The depot clock has stopped" in one Moment and "The depot clock
// stopped" in another, "One boot came home" / "One crate came home" / "One drone
// came home" across three eras. A Stream cannot see these as repeats, and a
// player reads them as the machine cycling nouns through one frame. Measured as
// the overlap of content words between every pair of skeletons in the catalog.
func TestNoNearDuplicateSkeletons(t *testing.T) {
	const maxOverlap = 0.55
	stop := map[string]bool{}
	for _, w := range strings.Fields("the a an of to and in was is it they them that on for with at from by as he his her she had has have one two were be are not this but or out who all what their there been no so when if back its into up over than where which being about since") {
		stop[w] = true
	}
	word := regexp.MustCompile(`[a-z']+`)
	type entry struct {
		id    string
		text  string
		words map[string]bool
	}
	var all []entry
	for _, m := range Moments() {
		for _, tpl := range templatesFor(m) {
			text := strings.TrimSpace(authoredText(tpl))
			ws := map[string]bool{}
			for _, w := range word.FindAllString(strings.ToLower(text), -1) {
				if len(w) > 2 && !stop[w] {
					ws[w] = true
				}
			}
			if len(ws) >= 3 {
				all = append(all, entry{tpl.ID, text, ws})
			}
		}
	}
	for i := 0; i < len(all); i++ {
		for j := i + 1; j < len(all); j++ {
			inter := 0
			for w := range all[i].words {
				if all[j].words[w] {
					inter++
				}
			}
			union := len(all[i].words) + len(all[j].words) - inter
			if o := float64(inter) / float64(union); o > maxOverlap {
				t.Errorf("near-duplicate skeletons (%.0f%% shared content words):\n    %s %q\n    %s %q",
					o*100, all[i].id, all[i].text, all[j].id, all[j].text)
			}
		}
	}
}

// TestAbstractResourcesAreNeverGoods pins needTangible end to end: a resource
// you cannot carry may appear as a figure ("120 knowledge") and nowhere else.
// Before the guard, expedition rewards and faction specialties produced
// "Nobody wants to sit up guarding the knowledge tonight" and "There is a smell
// of data on the back stairs".
func TestAbstractResourcesAreNeverGoods(t *testing.T) {
	if len(abstractResource) == 0 {
		t.Fatal("abstractResource is empty")
	}
	for key := range abstractResource {
		label := resourceLabel(key)
		figure := regexp.MustCompile(`\d+ ` + regexp.QuoteMeta(label))
		bare := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(label) + `\b`)
		for _, m := range Moments() {
			rng := rand.New(rand.NewSource(int64(len(key)) + int64(m)))
			for _, age := range config.AgeOrder() {
				for i := 0; i < 40; i++ {
					req := Request{Moment: m, Age: age, Tone: allTones[i%4], Kind: gameKinds[m][i%len(gameKinds[m])],
						Subject: "Tech Consortium", Resource: key, Amount: 120}
					line := Line(req, rng)
					if bare.MatchString(figure.ReplaceAllString(line, "")) {
						t.Fatalf("%v at %s treats %q as goods: %q", m, age, key, line)
					}
				}
			}
		}
	}
}
