package flavor

import (
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// harbinger_test.go holds the tests specific to the nine harbinger Moments.
// They share every catalog lint in flavor_test.go and generator_test.go (hygiene,
// era bleed, AI tells, register quotas, burstiness, house tics, near-duplicates)
// but not the per-line Moments' SIZE floors, and the reason is how often they
// fire.
//
// # Why these Moments have their own pool floor
//
// The floors in TestSkeletonFloors and friends (70 era-eligible skeletons,
// 60 distinct in 200 draws) are sized for Moments that fire hundreds of times a
// run: a raid log, an expedition log. A harbinger fires once per epoch
// transition, which is at most six times a run, each time in a different age,
// and each firing writes one line per Moment. A player meets a given harbinger
// pool once a run. Writing seventy sentences per era bucket per Moment to guard
// against a repeat inside two hundred consecutive draws would buy nothing a
// player could notice, and the extra lines would be generic ones, which is the
// failure this catalog exists to avoid.
//
// What does matter, and what is held here instead:
//   - every pool a real request can reach (every age, every tier, with and
//     without a subject) is WIDER than the Stream's skeleton window, so the
//     no-repeat-inside-the-window promise holds for these Moments exactly as it
//     does for the others (TestStreamNeverRepeatsInsideItsWindow runs over them
//     unchanged);
//   - a tier-carrying warning is almost entirely that tier's voice;
//   - the speaker's own words never leak into another age;
//   - pre-Industrial prose never talks about odds, and Industrial-and-later
//     warnings do.

// harbingerMoments returns the harbinger Moments in registration order.
func harbingerMoments() []Moment {
	return []Moment{
		HarbingerArrival, HarbingerWarning, HarbingerAppeased, HarbingerBraced,
		HarbingerVindicated, HarbingerSpared, HarbingerDiscredited,
		HarbingerInvited, HarbingerFulfilled,
	}
}

// isHarbinger reports whether m is a harbinger Moment.
func isHarbinger(m Moment) bool {
	for _, h := range harbingerMoments() {
		if h == m {
			return true
		}
	}
	return false
}

// harbName is the roster name for an age, which is what the game sends as
// Request.Subject for every harbinger Moment.
func harbName(t testing.TB, age string) string {
	h, ok := config.HarbingerFor(age)
	if !ok {
		t.Fatalf("no harbinger for %s", age)
	}
	return h.Name
}

// gameSubject is the Subject the engine sends for a Moment at an age: the
// harbinger's name for the harbinger Moments, nothing otherwise.
func gameSubject(m Moment, age string) string {
	if !isHarbinger(m) {
		return ""
	}
	if h, ok := config.HarbingerFor(age); ok {
		return h.Name
	}
	return ""
}

// falseProphetsPossible reports whether an age's harbinger can be repeating a
// false warning. The game rolls once per epoch, with the chance of the epoch's
// first age, and every later figure of that epoch repeats the result.
func falseProphetsPossible(age string) bool {
	ep, ok := config.EpochByKey()[config.EpochForAge(age)]
	if !ok || len(ep.Ages) == 0 {
		return false
	}
	h, ok := config.HarbingerFor(ep.Ages[0])
	return ok && h.FalseProphetChance > 0
}

// harbReachable reports whether the game can ever ask for m at age. Only
// HarbingerDiscredited is restricted: it needs a false warning, which only the
// Stone, Iron and Steel Eras can roll.
func harbReachable(m Moment, age string) bool {
	if m == HarbingerDiscredited {
		return falseProphetsPossible(age)
	}
	return true
}

// harbKinds returns the Request.Kind values the game sends for m at age.
func harbKinds(m Moment, age string) []string {
	switch m {
	case HarbingerWarning:
		return HarbingerTiers()
	case HarbingerFulfilled, HarbingerVindicated, HarbingerDiscredited:
		if falseProphetsPossible(age) {
			return []string{"", KindFalseProphet}
		}
		return []string{""}
	default:
		return []string{""}
	}
}

// TestHarbingerPoolFloors is the size floor for these Moments. See the file
// comment for why it is not TestSkeletonFloors's.
func TestHarbingerPoolFloors(t *testing.T) {
	const (
		// withSubject is the floor for the request the game actually sends. It
		// is streamMemory+1: the pool must be wider than the Stream's window.
		withSubject = streamMemory + 1
		// bareFloor is the floor with no Subject: every {subject} line drops out.
		// It only has to keep a careless caller well away from a repeat.
		bareFloor = 24
		// tierShare is the minimum share of a tiered warning pool that is
		// written for that tier (or its neighbour), not tier-neutral.
		tierShare = 0.75
	)
	for _, m := range harbingerMoments() {
		smallest, where := 1<<30, ""
		for _, age := range config.AgeOrder() {
			if !harbReachable(m, age) {
				continue
			}
			for _, kind := range harbKinds(m, age) {
				req := Request{Moment: m, Age: age, Kind: kind, Subject: harbName(t, age)}
				pool := eligible(req)
				if len(pool) < withSubject {
					t.Errorf("%v at %s kind %q: %d eligible skeletons; want >= %d (wider than the Stream window)",
						m, age, kind, len(pool), withSubject)
				}
				if len(pool) < smallest {
					smallest, where = len(pool), age+"/"+kind
				}
				req.Subject = ""
				if n := len(eligible(req)); n < bareFloor {
					t.Errorf("%v at %s kind %q with no subject: %d eligible skeletons; want >= %d", m, age, kind, n, bareFloor)
				}
				if m == HarbingerWarning {
					tiered := 0
					for _, tp := range pool {
						if len(tp.Kinds) > 0 {
							tiered++
						}
					}
					if share := float64(tiered) / float64(len(pool)); share < tierShare {
						t.Errorf("%v at %s tier %q: only %.0f%% of the pool is tier-specific; want >= %.0f%%",
							m, age, kind, share*100, tierShare*100)
					}
				}
			}
		}
		t.Logf("%-21v %4d skeletons; smallest reachable pool %d (%s)", m, DistinctSkeletons(m), smallest, where)
	}
}

// harbRequests is the realistic request set for a harbinger Moment at an age:
// every kind the game sends, every tone, with the roster name and without it,
// plus hostile fields the Moment must shrug off (a resource, an amount, a count).
func harbRequests(t testing.TB, m Moment, age string) []Request {
	var out []Request
	for _, kind := range harbKinds(m, age) {
		for _, tone := range allTones {
			for _, subj := range []string{harbName(t, age), ""} {
				out = append(out,
					Request{Moment: m, Age: age, Kind: kind, Tone: tone, Subject: subj},
					Request{Moment: m, Age: age, Kind: kind, Tone: tone, Subject: subj, Resource: "gold", Amount: 512, Count: 7, Ticks: 30},
				)
			}
		}
	}
	return out
}

var digitRE = regexp.MustCompile(`[0-9]`)

// TestHarbingerLinesAreClean is the coverage matrix for the new Moments: every
// Moment, every one of the 22 ages, every tier, every tone, through a Stream,
// and every line must pass the hygiene contract, carry no word from a foreign
// era, and carry no digit. The odds are shown by the UI; a digit in the prose
// would sooner or later disagree with it.
func TestHarbingerLinesAreClean(t *testing.T) {
	draws := 6
	if testing.Short() {
		draws = 2
	}
	lines := 0
	for _, age := range config.AgeOrder() {
		e, _ := eraOf(age)
		foreign := foreignMarkers(e)
		for _, m := range harbingerMoments() {
			rng := rand.New(rand.NewSource(int64(len(age))*977 + int64(m)))
			st := NewStream()
			for _, req := range harbRequests(t, m, age) {
				for i := 0; i < draws; i++ {
					line := st.Line(req, rng)
					lines++
					if p := hygiene(line); p != "" {
						t.Fatalf("%v at %s (%+v): %s\n    %q", m, age, req, p, line)
					}
					if digitRE.MatchString(line) {
						t.Fatalf("%v at %s (%+v): a digit in a harbinger line: %q", m, age, req, line)
					}
					if hit := foreign.FindString(callerText(line, req)); hit != "" {
						t.Fatalf("era bleed: %v at %s produced %q, which carries %q", m, age, line, hit)
					}
				}
			}
		}
	}
	t.Logf("%d harbinger lines across %d ages, all clean", lines, len(config.AgeOrder()))
}

// TestHarbingerTemplatesUseNoFigures pins the digit rule at the source: no
// harbinger sentence may use a placeholder that renders a number or a resource.
func TestHarbingerTemplatesUseNoFigures(t *testing.T) {
	banned := []string{"{amt}", "{amt_res}", "{n}", "{ticks}", "{res}", "{res_stores}", "{res_haul}"}
	for _, m := range harbingerMoments() {
		for _, tp := range templatesFor(m) {
			text := authoredText(tp)
			for _, p := range tp.Parts {
				if p.bank != "" {
					text += " " + strings.Join(banks[p.bank], " ")
				}
			}
			if digitRE.MatchString(text) {
				t.Errorf("%s carries a digit: %q", tp.ID, text)
			}
			for _, b := range banned {
				if strings.Contains(text, b) {
					t.Errorf("%s uses %s; harbinger lines carry no figures", tp.ID, b)
				}
			}
			if i := strings.Index(text, "{subject}"); i >= 0 {
				before := strings.TrimRight(text[:i], " ")
				if before == "" || strings.HasSuffix(before, ".") || strings.HasSuffix(before, "“") {
					t.Errorf("%s opens a sentence with {subject}, but roster names open lowercase: %q", tp.ID, text)
				}
			}
		}
	}
}

// precisionWords is the vocabulary of a published forecast. None of it belongs
// in an age whose harbinger is vague.
var precisionWords = regexp.MustCompile(`(?i)\b(odds|percent|percentage|per cent|probability|probable|statistic\w*|decimal|forecast\w*|figures?|estimate\w*|numbers?|risk score|model)\b`)

// TestHarbingerPrecisionVocabulary holds ForecastPrecision in the prose. Before
// the Industrial Age nothing a harbinger Moment can say mentions odds, figures
// or estimates. From the Industrial Age on, every tier of the warning can reach
// at least two sentences that refer to the published figure, so a player who
// has the number on screen also hears it talked about.
func TestHarbingerPrecisionVocabulary(t *testing.T) {
	for _, age := range config.AgeOrder() {
		h, _ := config.HarbingerFor(age)
		for _, m := range harbingerMoments() {
			seen := map[string]bool{}
			for _, kind := range harbKinds(m, age) {
				for _, tp := range eligible(Request{Moment: m, Age: age, Kind: kind, Subject: h.Name}) {
					if seen[tp.ID] {
						continue
					}
					seen[tp.ID] = true
					if h.ForecastPrecision == config.ForecastVague {
						if w := precisionWords.FindString(authoredText(tp)); w != "" {
							t.Errorf("%s can fire in %s, whose harbinger is vague, but says %q: %q",
								tp.ID, age, w, authoredText(tp))
						}
					}
				}
			}
		}
		if h.ForecastPrecision != config.ForecastNumeric {
			continue
		}
		for _, tier := range HarbingerTiers() {
			n := 0
			for _, tp := range eligible(Request{Moment: HarbingerWarning, Age: age, Kind: tier, Subject: h.Name}) {
				if precisionWords.MatchString(authoredText(tp)) {
					n++
				}
			}
			if n < 2 {
				t.Errorf("HarbingerWarning at %s tier %q reaches only %d sentences about the published figure; want >= 2",
					age, tier, n)
			}
		}
	}
}

// speakerWords is the vocabulary that belongs to one speaker. A harbinger line
// reachable in any other age may not use it: the wild man does not mention
// newspapers and the viral video does not mention scrolls. It only covers the
// harbinger Moments, whose per-age pools are the reason Ages gating exists.
var speakerWords = map[string][]string{
	"primitive_age":    {"wild man"},
	"stone_age":        {"hermit"},
	"bronze_age":       {"soothsayer", "knucklebones", "sparrows?"},
	"iron_age":         {"desert", "locusts"},
	"classical_age":    {"cleft rock"},
	"medieval_age":     {"crier", "oyez", "hear ye"},
	"renaissance_age":  {"astrologer", "astrolabe", "horoscope", "conjunctions?"},
	"colonial_age":     {"pamphlets?", "pamphleteer", "coffee house"},
	"industrial_age":   {"newsboy", "read all about it"},
	"victorian_age":    {"soap crate", "soapbox", "sandwich board"},
	"electric_age":     {"telegraph", "dispatch(es)?", "operator"},
	"atomic_age":       {"civil defence", "wireless", "announcer"},
	"modern_age":       {"anchor", "autocue"},
	"information_age":  {"chain email", "inbox(es)?", "forwarded"},
	"digital_age":      {"video", "views", "thumbnail"},
	"cyberpunk_age":    {"ghost", "arcology"},
	"fusion_age":       {"warden"},
	"space_age":        {"monitoring station", "monitor's"},
	"interstellar_age": {"beacon('s)?"},
	"galactic_age":     {"linguists"},
	"quantum_age":      {"future self", "nine years"},
	"transcendent_age": {"unmade", "branch(es)? that ended"},
}

// TestHarbingerSpeakerFit is the per-age half of the era-bleed guard. The era
// buckets are too coarse for a cast that changes every age (a newsboy and a
// civil-defence broadcast share a bucket), so each speaker's vocabulary is
// checked against every age it is NOT at home in.
func TestHarbingerSpeakerFit(t *testing.T) {
	res := map[string]*regexp.Regexp{}
	for age, words := range speakerWords {
		res[age] = regexp.MustCompile(`(?i)\b(` + strings.Join(words, "|") + `)\b`)
	}
	order := config.AgeOrder()
	for _, m := range harbingerMoments() {
		for _, tp := range templatesFor(m) {
			text := authoredText(tp)
			for _, p := range tp.Parts {
				if p.bank != "" {
					text += " " + strings.Join(banks[p.bank], " ")
				}
			}
			for _, age := range order {
				if !eraOK(tp, mustEra(t, age)) || !ageOK(tp, age) {
					continue // not reachable here
				}
				for owner, re := range res {
					if owner == age {
						continue
					}
					if w := re.FindString(text); w != "" && !speakerAlsoAt(res, age, w) {
						t.Errorf("%s can fire in %s but uses %q, which belongs to the %s harbinger: %q",
							tp.ID, age, w, owner, text)
					}
				}
			}
		}
	}
}

// speakerAlsoAt reports whether w is also one of age's own speaker words (a
// word two speakers share is fine in either age).
func speakerAlsoAt(res map[string]*regexp.Regexp, age, w string) bool {
	re, ok := res[age]
	return ok && re.MatchString(w)
}

func mustEra(t testing.TB, age string) era {
	e, ok := eraOf(age)
	if !ok {
		t.Fatalf("age %q has no era", age)
	}
	return e
}

// TestHarbingerFalseProphetLines pins the KindFalseProphet lines: they exist
// only in HarbingerFulfilled and HarbingerVindicated, only in ages that can
// have a false prophet, and a plain request never reaches them.
func TestHarbingerFalseProphetLines(t *testing.T) {
	found := 0
	for _, m := range Moments() {
		for _, tp := range templatesFor(m) {
			isFalse := false
			for _, k := range tp.Kinds {
				if k == KindFalseProphet {
					isFalse = true
				}
			}
			if !isFalse {
				continue
			}
			found++
			if m != HarbingerFulfilled && m != HarbingerVindicated {
				t.Errorf("%s is a false-prophet line outside HarbingerFulfilled and HarbingerVindicated", tp.ID)
			}
			for _, age := range config.AgeOrder() {
				if eraOK(tp, mustEra(t, age)) && ageOK(tp, age) && !falseProphetsPossible(age) {
					t.Errorf("%s can fire in %s, which has no false prophets", tp.ID, age)
				}
			}
		}
	}
	if found == 0 {
		t.Fatal("no KindFalseProphet lines in the catalog")
	}
	for _, m := range []Moment{HarbingerFulfilled, HarbingerVindicated} {
		for _, tp := range eligible(Request{Moment: m, Age: "bronze_age", Subject: "the Soothsayer"}) {
			if len(tp.Kinds) > 0 {
				t.Errorf("%s fired on a plain %s request", tp.ID, m)
			}
		}
	}
}

// TestHarbingerDeterminism pins the purity contract for the new Moments on both
// paths: same seed and request, same lines; a different seed, different lines.
func TestHarbingerDeterminism(t *testing.T) {
	const n = 120
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
	for _, m := range harbingerMoments() {
		for _, age := range []string{"primitive_age", "medieval_age", "atomic_age", "cyberpunk_age", "transcendent_age"} {
			if !harbReachable(m, age) {
				continue
			}
			kinds := harbKinds(m, age)
			req := Request{Moment: m, Age: age, Kind: kinds[len(kinds)-1], Subject: harbName(t, age)}
			for _, stream := range []bool{false, true} {
				a, b, c := run(req, 4242, stream), run(req, 4242, stream), run(req, 4243, stream)
				differ := 0
				for i := range a {
					if a[i] != b[i] {
						t.Fatalf("%v %s stream=%v: line %d differs under the same seed", m, age, stream, i)
					}
					if a[i] != c[i] {
						differ++
					}
				}
				if differ < n*8/10 {
					t.Errorf("%v %s stream=%v: seeds 4242 and 4243 agree on %d of %d lines", m, age, stream, n-differ, n)
				}
			}
		}
	}
}

// TestHarbingerTiersSoundDifferent is a cheap check that the tiers are not
// decorative: at every age the none-tier and high-tier warning pools share
// nothing but the small tier-neutral set.
func TestHarbingerTiersSoundDifferent(t *testing.T) {
	for _, age := range config.AgeOrder() {
		name := harbName(t, age)
		none := map[string]bool{}
		for _, tp := range eligible(Request{Moment: HarbingerWarning, Age: age, Kind: TierNone, Subject: name}) {
			none[tp.ID] = true
		}
		for _, tp := range eligible(Request{Moment: HarbingerWarning, Age: age, Kind: TierHigh, Subject: name}) {
			if none[tp.ID] && len(tp.Kinds) > 0 {
				t.Errorf("%s at %s is reachable from both the none and the high tier", tp.ID, age)
			}
		}
	}
}
