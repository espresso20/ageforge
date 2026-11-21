package main

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/espresso20/ageforge/config"
)

// This file applies the game's copy rules to a rewrite. The word lists and
// patterns live in rules.go, copied from the lint tests that were removed with
// the test suite. What each rule does with its list (which text it applies to,
// what counts as a hit) is stated here, a few lines apiece.

// pattern is a regular expression from a lint, with what the lint says
// about it.
type pattern struct {
	re   *regexp.Regexp
	note string
}

// copyRules are the lists and patterns the copy rules apply.
type copyRules struct {
	// the wording rules over ui, game and boon.
	retiredTerms []pattern
	rawAmount    *regexp.Regexp
	bang         *regexp.Regexp
	exemptFiles  map[string]bool // files the wording rules skip
	readBack     []string        // text the game parses back out of old saves
	// the stale copy rule over ui.
	retiredCopy []pattern
	// the wiki's rule.
	retiredDoc []pattern
	// the style rules over config's text, and the event text rule.
	styleBans  []string
	rawKey     *regexp.Regexp
	verbLeft   *regexp.Regexp
	maxLine    int
	otherLines map[string]string // every catalog sentence -> its id, to catch a duplicate
}

// loadRules builds the copy rules from rules.go.
func loadRules() (*copyRules, error) {
	cr := &copyRules{
		retiredTerms: retiredTerms,
		rawAmount:    rawAmountRe,
		bang:         bangRe,
		exemptFiles:  map[string]bool{},
		readBack:     copyExemptLiterals,
		retiredCopy:  retiredCopy,
		retiredDoc:   retiredDoc,
		styleBans:    styleBans,
		rawKey:       rawKeyRe,
		verbLeft:     verbLeftRe,
		maxLine:      maxLineRunes,
	}
	for rel := range copyExemptFiles {
		cr.exemptFiles[rel] = true
	}
	return cr, nil
}

// check returns what is wrong with a rewrite of a row under the game's copy
// rules, one sentence a problem. A rule is skipped when the original text
// already breaks it, so old text is not a reason to refuse new text.
func (cr *copyRules) check(r *row, yours string) []string {
	var out []string
	// hit adds a problem when the rewrite trips a rule the original passes.
	hit := func(trips func(s string) bool, problem string) {
		if trips(yours) && !trips(r.current) {
			out = append(out, problem)
		}
	}
	has := func(sub string) func(string) bool {
		return func(s string) bool { return strings.Contains(s, sub) }
	}
	matches := func(re *regexp.Regexp) func(string) bool {
		return func(s string) bool { return re.MatchString(s) }
	}

	if r.rules&rulesUI != 0 {
		hit(has("—"), "has an em dash: use a period, colon, comma or parentheses")
		hit(matches(cr.bang), "has an exclamation mark: routine events are stated, not shouted")
		hit(matches(cr.rawAmount), "prints an amount with a bare number verb")
		if strings.Contains(yours, " ") {
			for _, p := range cr.retiredTerms {
				hit(matches(p.re), fmt.Sprintf("uses a retired term (“%s”): the game says %s", p.re.FindString(yours), p.note))
			}
		}
		if r.u != nil && strings.HasPrefix(r.u.f.rel, "ui/") {
			for _, p := range cr.retiredCopy {
				hit(matches(p.re), "brings back retired text: "+p.note)
			}
		}
	}
	if r.rules&rulesConfig != 0 {
		hit(func(s string) bool { return strings.ContainsAny(s, "—–") }, "has a dash: use a period, colon, comma or parentheses")
		hit(has("!"), "has an exclamation mark")
		hit(matches(cr.rawKey), fmt.Sprintf("has a raw key in it (“%s”): write the name a player sees", cr.rawKey.FindString(yours)))
		for _, b := range cr.styleBans {
			hit(func(s string) bool { return strings.Contains(strings.ToLower(s), b) }, fmt.Sprintf("uses “%s”, which the style guard for game data bans (US spelling, the glossary, stock words)", b))
		}
	}
	if r.rules&rulesDocs != 0 {
		for _, p := range cr.retiredDoc {
			hit(matches(p.re), fmt.Sprintf("uses “%s”, which the wiki's lint refuses, and the wiki quotes this text: %s", p.re.FindString(yours), p.note))
		}
	}
	if r.rules&rulesNoMark != 0 {
		hit(func(s string) bool { return strings.ContainsAny(s, "[]") }, "has a square bracket, which this line may not carry (the game wraps it in a style tag of its own)")
		hit(matches(cr.verbLeft), "has a % directive, which this line may not carry")
	}
	if r.rules&rulesSaveNam != 0 {
		hit(func(s string) bool {
			for _, c := range strings.TrimSpace(s) {
				if !unicode.IsLetter(c) && c != ' ' {
					return true
				}
			}
			return strings.Contains(strings.TrimSpace(s), "  ")
		}, "has something other than letters and single spaces, and a save's name becomes a file name")
	}
	out = append(out, checkTech(r, yours)...)
	return out
}

// checkTech applies the tech tree's rule for a card code: two to
// config.TechCodeMax capital letters, no two techs alike. A tech that sets
// no code takes one from its name, so a rename can change it.
func checkTech(r *row, yours string) []string {
	if r.u == nil || r.u.st().typ != "config.TechDef" {
		return nil
	}
	s := r.u.st()
	var code string
	switch {
	case s.field == "Code":
		code = strings.TrimSpace(yours)
	case s.field == "Name" && s.sib["Code"] == "":
		code = config.TechCodeFor(strings.TrimSpace(yours))
	default:
		return nil
	}
	valid := len(code) >= 2 && len(code) <= config.TechCodeMax
	for _, c := range code {
		if c < 'A' || c > 'Z' {
			valid = false
		}
	}
	if !valid {
		if s.field == "Name" {
			return []string{fmt.Sprintf("would give the tech the card code “%s” (a tech with no code of its own takes it from the first word of its name): a code is 2 to %d capital letters", code, config.TechCodeMax)}
		}
		return []string{fmt.Sprintf("is not a card code: 2 to %d capital letters", config.TechCodeMax)}
	}
	for _, t := range config.Technologies() {
		if t.Key != s.sib["Key"] && t.Code == code {
			if s.field == "Name" {
				return []string{fmt.Sprintf("would give the tech the card code %s, which %s already has (a tech with no code of its own takes it from the first word of its name)", code, t.Name)}
			}
			return []string{fmt.Sprintf("is the card code %s already has", t.Name)}
		}
	}
	return nil
}
