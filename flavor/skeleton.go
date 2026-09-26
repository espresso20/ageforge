package flavor

import "strings"

// The SKELETON model — the unit of authorship in this package.
//
// A skeleton is ONE authored sentence. Not a clause, not a fragment, not a half
// of something the renderer welds to another half: a finished line a person
// wrote and read back. It may carry AT MOST ONE slot, and that slot is always a
// NOUN PHRASE drawn from a bank of the same tight category, because same-category
// noun substitution is the one substitution that always reads:
//
//	"They came back with ~."   ← slot: a haul
//	"The ~ did not survive the crossing."  ← slot: a piece of kit
//
// What this model deliberately CANNOT express is the thing that produced
// "The quartermaster looks at what returned and asks nothing at all, and the
// quartermaster says nothing, loudly." — two independently drawn clauses joined
// by a connective the machine chose. There is no semantic model behind such a
// join, so it double-mentions subjects and welds unrelated beats, and no amount
// of extra fragments fixes it. Joining clauses is not supported here. It is not
// a missing feature.
//
// The consequence for the quality metric is the whole point. Perceived repetition
// is governed by the SMALLEST bank a line draws from, never by the product of the
// bank sizes, so the number worth asserting a floor on is DistinctSkeletons — how
// many separate sentences a Moment can say — and not Capacity.

// slotMark is the single slot marker inside an authored sentence. A tilde never
// occurs in English prose, so it needs no escaping and reads as a gap on the page.
const slotMark = "~"

// --- register and form: the tags that keep the catalog off one note ----------
//
// Two rewrites of this catalog failed the same way. Both were told to write
// "good flavour text" and both converged on ONE voice — terse ironic
// understatement — because that is the attractor the instruction lands in. The
// prose was fine sentence by sentence and unreadable forty lines at a time.
//
// The fix is not a better adjective in a doc comment. It is a QUOTA, declared per
// sentence and asserted in a test, so the mix is structural. Two axes:
//
//	register — how much wit the sentence carries. Most of a catalog must be rPlain:
//	           a flat circumstantial fact, reported without a raised eyebrow. Comedy
//	           needs a straight man and the straight man is 30% of the roster.
//	form     — whether the sentence is NARRATION at all. A ledger line, a grumble, a
//	           fragment of overheard speech and a posted notice are different kinds
//	           of text, not different phrasings of the same one.
//
// register has NO usable zero value on purpose: an author who forgets the tag gets
// a test failure rather than a silent vote for the quota's cheapest bucket.
type register int

const (
	// regUnset is the zero value and is invalid — TestRegisterQuotas rejects it.
	regUnset register = iota
	// rPlain carries no irony and no wit. It reports a circumstantial fact and
	// stops. If a reader could smile at it, it is not rPlain.
	rPlain
	// rWry carries a dry angle — the sentence knows something it is not saying —
	// without reaching for a laugh.
	rWry
	// rJoke reaches for a laugh. Capped at one line in five.
	rJoke
)

// form is what KIND of text a sentence is. The zero value is narration, which is
// the safe default: the quota on non-narrative forms is a FLOOR, so a forgotten
// tag makes the test harder to pass, never easier.
type form int

const (
	// fNarr is plain narration: something happened, here it is. The default.
	fNarr form = iota
	// fLedger is an inventory line, a tally, a manifest entry, a count.
	fLedger
	// fComplaint is somebody grumbling, in their own register.
	fComplaint
	// fOverheard is a fragment of speech, a rumour, what is being said.
	fOverheard
	// fNotice is administrative: an order, a posted note, a rule, a form.
	fNotice
)

// knownTopics is the SUBJECT-MATTER vocabulary. Every skeleton declares one, and
// Stream suppresses a topic that fired in the last few lines — which is the fix
// for the residual complaint that two adjacent lines were both about the dog. The
// recency ring filtered skeleton identity, which cannot see that two different
// sentences are about the same thing.
//
// A closed vocabulary rather than free text, so a typo cannot quietly fragment a
// topic into two that never suppress each other (TestTopicVocabulary).
var knownTopics = map[string]bool{
	"animal": true, "argument": true, "authority": true, "border": true,
	"building": true, "casualty": true, "count": true, "food": true,
	"ground": true, "haul": true, "kit": true, "machine": true,
	"map": true, "message": true, "money": true, "name": true,
	"noise": true, "paper": true, "people": true, "religion": true,
	"rumour": true, "sleep": true, "smell": true, "stranger": true,
	"time": true, "town": true, "trade": true, "weapon": true,
	"weather": true, "wound": true,
}

// skel is one authored sentence plus its eligibility rules. Text carries NO
// terminal full stop — template() supplies it, so a line can never ship without
// one and an author can never accidentally ship two.
type skel struct {
	// Text is the sentence, with at most one slotMark. Internal punctuation is
	// fine ("Two did not come back~ the rest brought maps" is not how this works,
	// but "Two did not come back. The rest brought maps" is one authored unit and
	// perfectly legal). No terminal full stop.
	Text string
	// Slot names the noun-phrase bank filling this sentence's slotMark. Empty
	// when Text has no slot, which is the common case and the preferred one.
	Slot string
	// Needs are the Request fields the sentence requires — the mechanism that
	// makes "0 = omit from the sentence" structural rather than conditional.
	Needs need
	// Tones restricts the sentence to these tones; nil means any.
	Tones []Tone
	// Kinds restricts the sentence to these Request.Kind values (lowercased);
	// nil means any.
	Kinds []string
	// Reg is how much wit the sentence carries. REQUIRED — the zero value fails
	// TestRegisterQuotas, because a forgotten tag must not vote in the quota.
	Reg register
	// Form is what kind of text the sentence is. Zero means narration.
	Form form
	// Topic is the sentence's subject matter, from knownTopics. Stream uses it to
	// stop two adjacent lines both being about the dog.
	Topic string
}

// template converts an authored sentence into the renderer's template form.
//
// The parts always begin with a LITERAL, which is what makes the signature
// contract cheap: a skeleton's leading run of authored text reaches the output
// verbatim, so it is the line's identity for anyone holding only the string
// (see anchorOf and Signatures).
func (s skel) template(id string, eras []era) tmpl {
	t := tmpl{
		ID: id, Needs: s.Needs, Tones: s.Tones, Kinds: s.Kinds, Eras: eras,
		Reg: s.Reg, Form: s.Form, Topic: s.Topic,
	}
	if namesResourceAsGoods(s.Text) {
		t.Needs |= needRes | needTangible
	}
	if s.Slot == "" {
		t.Parts = []part{lit(s.Text + ".")}
		return t
	}
	head, tail, _ := strings.Cut(s.Text, slotMark)
	t.Parts = []part{lit(head), b(s.Slot), lit(tail + ".")}
	return t
}

// namesResourceAsGoods reports whether a sentence uses one of the frames that
// treat the resource as a physical load: bare {res}, {res_stores}, {res_haul}.
// {amt_res} alone does not count; "120 knowledge" is a figure, not a sack.
func namesResourceAsGoods(text string) bool {
	return strings.Contains(text, "{res}") ||
		strings.Contains(text, "{res_stores}") ||
		strings.Contains(text, "{res_haul}")
}

// pool converts a batch of authored sentences into templates under one era
// gating, numbering their ids from prefix.
//
// Ids are POSITIONAL within a pool. Nothing persists them — Result.Template is
// read by tests and by nothing that survives a process — so an insert renumbering
// its neighbours is harmless. The stable handle a text-only consumer wants is the
// signature, which is derived from the prose itself and moves with it.
func pool(prefix string, eras []era, ss []skel) []tmpl {
	out := make([]tmpl, 0, len(ss))
	for i, s := range ss {
		out = append(out, s.template(prefix+"_"+itoa(i), eras))
	}
	return out
}

// itoa is strconv.Itoa without the import, for ids that are always small.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [4]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// anchorOf returns a template's SIGNATURE: the leading run of authored text that
// is guaranteed to reach the output verbatim.
//
// Every skeleton starts with a literal, and nothing in the render path rewrites
// the inside of a literal or transforms case, so everything up to the first slot
// or placeholder survives into the finished line character for character. That
// run is therefore an enumerable handle on the line for a consumer that only has
// the string — game/boon_tuning_test.go classifies encounter outcomes with it.
//
// Returns "" for a template that opens on a bank draw, which the catalog forbids
// (TestSkeletonShape).
func anchorOf(t tmpl) string {
	if len(t.Parts) == 0 || t.Parts[0].bank != "" {
		return ""
	}
	s := t.Parts[0].lit
	if i := strings.IndexByte(s, '{'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimRight(s, " ")
}

// authoredText returns a template's prose with its slot and placeholders elided —
// what the sentence says, independent of what fills it. Used by the era-bleed lint
// and the length lint, both of which care about the words an author chose.
func authoredText(t tmpl) string {
	var sb strings.Builder
	for _, p := range t.Parts {
		if p.bank != "" {
			sb.WriteString(" " + slotMark + " ")
			continue
		}
		sb.WriteString(p.lit)
	}
	return sb.String()
}

// wordCount counts words in an authored sentence, treating a slot or a
// placeholder as exactly one word. It is the measure the length cap is asserted
// on: how long the line reads, not how long the substitution happens to make it.
func wordCount(s string) int {
	s = strings.ReplaceAll(s, slotMark, "x")
	for _, ph := range []string{
		"{subject}", "{res_stores}", "{res_haul}", "{amt_res}", "{res}", "{amt}", "{n}", "{ticks}",
	} {
		s = strings.ReplaceAll(s, ph, "x")
	}
	return len(strings.Fields(s))
}
