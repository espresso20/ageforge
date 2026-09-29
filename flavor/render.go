package flavor

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// The template model. A template is DATA: an ordered list of parts, where a part
// is either a literal connective the template owns ("; ", " — ", ".") or a
// reference to a named fragment bank. Rendering draws one authored fragment per
// bank part and concatenates. Nothing transforms case and nothing rewrites the
// inside of a fragment, which is what makes Signatures() a reliable classifier:
// an authored fragment reaches the output verbatim.
//
// Casing convention, enforced by authorship rather than code:
//   - fragments that open a sentence are Capitalized and carry NO terminal
//     punctuation (the template supplies "." or the connective).
//   - fragments that continue a sentence after ", and " / " — " / "; " are
//     lowercase and also carry no terminal punctuation.
//
// That convention is why there is no capitalize() in this file. Adding one would
// break the verbatim guarantee.

// part is one slot in a template: a literal (bank == "") or a bank draw.
type part struct {
	lit  string
	bank string
}

// lit builds a literal part.
func lit(s string) part { return part{lit: s} }

// b builds a bank-reference part.
func b(name string) part { return part{bank: name} }

// need is a bitmask of Request fields a template requires. A template whose
// needs are not met is not eligible, which is how "0 = omit from the sentence"
// is enforced: there is no code path that renders a zero amount, because no
// amount-bearing template can be selected without one.
type need uint16

const (
	// needSubject requires a non-empty Request.Subject.
	needSubject need = 1 << iota
	// needRes requires a non-empty Request.Resource.
	needRes
	// needMassRes requires a resource that is a MASS noun (implies needRes).
	// Templates using the {res_stores} / {res_haul} frames declare this.
	needMassRes
	// needAmount requires Request.Amount to round to 1 or more.
	needAmount
	// needCount requires Request.Count >= 1.
	needCount
	// needTicks requires Request.Ticks >= 1.
	needTicks
	// needTangible requires a resource you can carry, guard, stack or smell.
	// Every skeleton that names a resource as a physical thing ({res},
	// {res_stores}, {res_haul}) gets it automatically in skel.template, so no
	// author can forget it; {amt_res} is a figure and works for anything.
	// Without it the catalog produced "Nobody wants to sit up guarding the
	// knowledge tonight" and "There is a smell of data on the back stairs".
	needTangible
)

// era is a technological bucket used to keep anachronisms out of the prose: gate
// wardens and militia bells belong to the feudal ages, cargo seals and telemetry
// to the cosmic ones. Age "" or an unknown age imposes NO era bound, which is the
// permissive default a zero Request relies on.
//
// Five buckets, not three. Three put the Stone Age and the Colonial Age in the
// same voice at one end and the Information Age and the Transcendent Age in the
// same voice at the other, which is how carts and gate wardens ended up narrating
// orbital strikes. The buckets below each cover a span whose imagery genuinely
// holds together.
//
// Era gating is SEASONING, not the main course. The bulk of every Moment's prose
// lives in ungated banks that are authored to be genuinely age-agnostic — the
// standard config/log_flavor.go sets, where "a single line has to land in the
// Stone Age and the Quantum Age alike". An era bank exists only for the lines
// whose imagery IS the joke, and those lines never leave their bucket.
type era int

const (
	eraAncient    era = iota // primitive_age .. classical_age
	eraFeudal                // medieval_age .. colonial_age
	eraIndustrial            // industrial_age .. modern_age
	eraDigital               // information_age .. fusion_age
	eraCosmic                // space_age .. transcendent_age
)

// tmpl is one sentence structure.
type tmpl struct {
	// ID is a STABLE identifier surfaced as Result.Template. Never reuse an id
	// for different prose — downstream classification keys off it.
	ID string
	// Needs are the Request fields this template requires.
	Needs need
	// Tones restricts eligibility to these tones; nil means any tone.
	Tones []Tone
	// Kinds restricts eligibility to these Request.Kind values (compared
	// lowercased); nil means any kind.
	Kinds []string
	// Eras restricts eligibility to these eras; nil means any era. An unknown or
	// empty Request.Age also matches everything.
	Eras []era
	// Ages narrows eligibility further, to these config age keys; nil means any
	// age in Eras. It exists for the harbinger Moments, whose speaker changes
	// every age: a newsboy's headline and a civil-defence broadcast share an
	// era bucket and nothing else. Like Eras, an unknown or empty Request.Age
	// matches everything. Set it through agePool, which also fills Eras so the
	// era-bleed lint sees the right bucket.
	Ages []string
	// Reg, Form and Topic carry the authored sentence's register tags through to
	// the tests that assert the mix, and Topic through to Stream. They never
	// affect eligibility — a quota is a property of the CATALOG, not a filter on
	// a draw; filtering on it would make some sentences unreachable.
	Reg   register
	Form  form
	Topic string
	// Parts is the ordered slot list.
	Parts []part
}

// capacity is the number of distinct lines this template can produce: the product
// of its bank sizes (1 for a literal-only template). A reference to a missing
// bank contributes 0, which makes the omission loudly visible in Capacity().
func (t tmpl) capacity() int {
	n := 1
	for _, p := range t.Parts {
		if p.bank == "" {
			continue
		}
		n *= len(banks[p.bank])
	}
	return n
}

// Generate produces one line for req. It is PURE given (req, rng): the draws are
// a template pick followed by one pick per bank slot, left to right, always in
// that order, so a seed reproduces the whole prose stream. It never touches
// package-level rand and never seeds anything.
//
// An unregistered Moment, or a Moment with no eligible template, returns the zero
// Result (empty Text) rather than panicking.
func Generate(req Request, rng *rand.Rand) Result {
	pool := eligible(req)
	if len(pool) == 0 {
		return Result{Moment: req.Moment}
	}
	// Draw 1: the structure. Draws 2..n happen in render.
	return render(pool[rng.Intn(len(pool))], req, rng)
}

// render fills one chosen template: one authored fragment per bank slot, left to
// right, then placeholder substitution. Split out of Generate so a Stream's
// last-resort pick can render a template it chose itself.
func render(t tmpl, req Request, rng *rand.Rand) Result {
	var sb strings.Builder
	var drew string
	for _, p := range t.Parts {
		if p.bank == "" {
			sb.WriteString(p.lit)
			continue
		}
		bank := banks[p.bank]
		if len(bank) == 0 {
			continue // an empty bank is a catalog bug; a test catches it
		}
		drew = bank[rng.Intn(len(bank))]
		sb.WriteString(drew)
	}

	return Result{
		Text:     fill(sb.String(), req),
		Moment:   req.Moment,
		Template: t.ID,
		slot:     drew,
	}
}

// eligible returns the templates of req.Moment whose needs, tone, kind, and era
// constraints are all satisfied, in catalog order (so the pick is deterministic
// given rng's state).
func eligible(req Request) []tmpl {
	all := templatesFor(req.Moment)
	if len(all) == 0 {
		return nil
	}
	// A bounded age reads its Moment's templates already filtered by era and
	// age (see ageCache), in catalog order, so the result is the list the
	// full scan would build, from a slice a fraction of the size.
	if byAge, ok := ageCache()[req.Moment][req.Age]; ok {
		all = byAge
	}
	have := satisfied(req)
	kind := strings.ToLower(strings.TrimSpace(req.Kind))

	out := make([]tmpl, 0, len(all))
	for _, t := range all {
		if t.Needs&^have != 0 {
			continue // some required field is missing
		}
		if !toneOK(t, req.Tone) || !kindOK(t, kind) {
			continue
		}
		out = append(out, t)
	}
	return out
}

// ageCache holds, per Moment and per config age key, the Moment's templates
// that the age's era and age gates admit, in catalog order. Built once, like
// templateCache. An empty or unknown age has no entry, and eligible falls
// back to the whole catalog, which is the permissive default.
//
// It exists because the harbinger Moments carry per-age pools: HarbingerWarning
// has over five hundred templates of which any one age can reach about a tenth,
// and scanning all of them on every draw (thirteen times a line through a
// Stream) was most of the cost of a line.
var ageCache = sync.OnceValue(func() map[Moment]map[string][]tmpl {
	out := make(map[Moment]map[string][]tmpl, len(Moments()))
	for _, m := range Moments() {
		byAge := make(map[string][]tmpl, len(ageIndex()))
		for age := range ageIndex() {
			e, _ := eraOf(age)
			var list []tmpl
			for _, t := range templatesFor(m) {
				if eraOK(t, e) && ageOK(t, age) {
					list = append(list, t)
				}
			}
			byAge[age] = list
		}
		out[m] = byAge
	}
	return out
})

// satisfied is the bitmask of needs req can actually meet.
func satisfied(req Request) need {
	var have need
	if strings.TrimSpace(req.Subject) != "" {
		have |= needSubject
	}
	if req.Resource != "" {
		have |= needRes
		if isMass(req.Resource) {
			have |= needMassRes
		}
		if !isAbstract(req.Resource) {
			have |= needTangible
		}
	}
	if roundAmount(req.Amount) >= 1 {
		have |= needAmount
	}
	if req.Count >= 1 {
		have |= needCount
	}
	if req.Ticks >= 1 {
		have |= needTicks
	}
	return have
}

// toneOK reports whether a template accepts this tone (nil Tones = any).
func toneOK(t tmpl, tone Tone) bool {
	if len(t.Tones) == 0 {
		return true
	}
	for _, x := range t.Tones {
		if x == tone {
			return true
		}
	}
	return false
}

// kindOK reports whether a template accepts this (already lowercased) kind. A
// template with no Kinds accepts anything; a template WITH Kinds never matches an
// empty kind, so a zero Request only ever sees the unconstrained templates.
func kindOK(t tmpl, kind string) bool {
	if len(t.Kinds) == 0 {
		return true
	}
	if kind == "" {
		return false
	}
	for _, x := range t.Kinds {
		if x == kind {
			return true
		}
	}
	return false
}

// eraOK reports whether a template accepts this era (nil Eras = any).
func eraOK(t tmpl, e era) bool {
	if len(t.Eras) == 0 {
		return true
	}
	for _, x := range t.Eras {
		if x == e {
			return true
		}
	}
	return false
}

// ageOK reports whether a template accepts this age key (nil Ages = any).
func ageOK(t tmpl, age string) bool {
	if len(t.Ages) == 0 {
		return true
	}
	for _, a := range t.Ages {
		if a == age {
			return true
		}
	}
	return false
}

// eraOf maps a config age key to its era bucket. bounded is false for an empty or
// unknown age, which lifts the era constraint entirely — the permissive default,
// and the reason a zero Request can still reach every template.
func eraOf(age string) (era, bool) {
	if age == "" {
		return eraAncient, false
	}
	idx, ok := ageIndex()[age]
	switch {
	case !ok:
		return eraAncient, false
	case idx <= 4: // primitive_age .. classical_age
		return eraAncient, true
	case idx <= 7: // medieval_age .. colonial_age
		return eraFeudal, true
	case idx <= 12: // industrial_age .. modern_age
		return eraIndustrial, true
	case idx <= 16: // information_age .. fusion_age
		return eraDigital, true
	default: // space_age .. transcendent_age
		return eraCosmic, true
	}
}

// ageIndex maps each config age key to its position in config.AgeOrder(). Built
// once: AgeOrder rebuilds the whole age table on every call, and eraOf runs on
// every draw.
var ageIndex = sync.OnceValue(func() map[string]int {
	order := config.AgeOrder()
	out := make(map[string]int, len(order))
	for i, k := range order {
		out[k] = i
	}
	return out
})

// fill substitutes the placeholders in an assembled sentence. Every placeholder
// this package understands is listed here; anything unrecognised is left alone
// (and caught by TestNoUnfilledPlaceholders).
//
// Placeholders:
//
//	{subject}    Request.Subject verbatim — only ever appears inside an authored
//	             frame that tolerates a verb-led order title or a proper noun of
//	             uncertain grammatical number (see catalog.go).
//	{res}        bare resource label: "food", "soldiers"
//	{res_stores} MASS ONLY: "your food stores"
//	{res_haul}   MASS ONLY: "a haul of food"
//	{amt_res}    inflected quantity: "40 food", "1 soldier", "12 soldiers"
//	{amt}        Request.Amount rounded, K/M formatted ("1.23M")
//	{n}          Request.Count, bare
//	{ticks}      Request.Ticks, bare
func fill(s string, req Request) string {
	if !strings.ContainsRune(s, '{') {
		return s
	}
	amt := roundAmount(req.Amount)
	return strings.NewReplacer(
		"{subject}", req.Subject,
		"{res}", phraseBare(req.Resource),
		"{res_stores}", phraseStores(req.Resource),
		"{res_haul}", phraseHaul(req.Resource),
		"{amt_res}", phraseQuantity(req.Resource, amt),
		"{amt}", textfmt.Int(amt),
		"{n}", textfmt.Int(req.Count),
		"{ticks}", strconv.Itoa(req.Ticks),
	).Replace(s)
}

// roundAmount rounds a resource amount to the nearest whole unit, clamping
// negatives to 0 (a negative amount is a caller bug, not a sentence).
func roundAmount(a float64) int {
	if a <= 0 || math.IsNaN(a) {
		return 0
	}
	if math.IsInf(a, 1) {
		return math.MaxInt32
	}
	return int(math.Round(a))
}
