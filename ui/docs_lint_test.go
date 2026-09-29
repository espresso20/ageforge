package ui

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Docs lint: the wiki, the landing page and the READMEs are player text, and
// they drift from the game the same way in-game strings do. These tests pin
// the house style (no em dashes, no exclamation marks, one word per concept)
// and a list of facts that were wrong once and must not come back.
//
// The count claims ("22 ages", "301 buildings", the landing-page hero stats)
// are checked against config by the smoke docsync scenario, not here.

// playerDocs are the files a player reads.
func playerDocs(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob("../site/docs/*.md")
	if err != nil {
		t.Fatal(err)
	}
	return append(files, "../site/index.html", "../site/script.js", "../README.md")
}

// devDocs are prose for contributors. They follow the punctuation rules but
// may name Go identifiers, raw keys and code paths.
func devDocs(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob("../design-and-architecture/*.md")
	if err != nil {
		t.Fatal(err)
	}
	return append(files, "../CONTRIBUTING.md")
}

func readDoc(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func relDoc(path string) string { return filepath.ToSlash(strings.TrimPrefix(path, "../")) }

var (
	fencedCode  = regexp.MustCompile("(?s)```.*?```")
	inlineCode  = regexp.MustCompile("`[^`\n]*`")
	htmlComment = regexp.MustCompile(`(?s)<!--.*?-->`)
	htmlBlock   = regexp.MustCompile(`(?is)<(script|style)\b.*?</(script|style)>`)
	htmlCode    = regexp.MustCompile(`(?is)<code\b.*?</code\s*>`)
	mdImage     = regexp.MustCompile(`!\[`)
	doctype     = regexp.MustCompile(`(?i)<!doctype`)
)

// prose strips what is code, not prose: fenced blocks, inline code spans,
// HTML comments, <script>/<style>/<code> elements, image markers. Line
// structure is kept (stripped spans become spaces) so line numbers hold.
func prose(path, s string) string {
	blank := func(m string) string {
		return strings.Map(func(r rune) rune {
			if r == '\n' {
				return r
			}
			return ' '
		}, m)
	}
	switch filepath.Ext(path) {
	case ".js":
		// Only string literals are player text in a script.
		var b strings.Builder
		for _, line := range strings.Split(s, "\n") {
			for _, m := range jsString.FindAllString(line, -1) {
				b.WriteString(m)
				b.WriteByte(' ')
			}
			b.WriteByte('\n')
		}
		return b.String()
	case ".html":
		s = htmlBlock.ReplaceAllStringFunc(s, blank)
		s = htmlCode.ReplaceAllStringFunc(s, blank)
		s = doctype.ReplaceAllStringFunc(s, blank)
	default:
		s = fencedCode.ReplaceAllStringFunc(s, blank)
		s = inlineCode.ReplaceAllStringFunc(s, blank)
	}
	s = htmlComment.ReplaceAllStringFunc(s, blank)
	return mdImage.ReplaceAllString(s, " [")
}

var jsString = regexp.MustCompile(`"(?:[^"\\\n]|\\.)*"`)

func lineOf(s string, off int) int { return strings.Count(s[:off], "\n") + 1 }

// TestDocsNoEmDash: no em dashes in player or contributor docs, code blocks
// included (quoted game output follows the same rule as the game).
func TestDocsNoEmDash(t *testing.T) {
	for _, f := range append(playerDocs(t), devDocs(t)...) {
		body := readDoc(t, f)
		for i, line := range strings.Split(body, "\n") {
			if strings.Contains(line, "—") {
				t.Errorf("%s:%d: em dash (use a period, colon, comma or parentheses): %s", relDoc(f), i+1, strings.TrimSpace(line))
			}
		}
	}
}

// TestDocsNoExclamations: routine prose does not shout. Code, HTML comments
// and image markers are exempt.
func TestDocsNoExclamations(t *testing.T) {
	for _, f := range append(playerDocs(t), devDocs(t)...) {
		p := prose(f, readDoc(t, f))
		if filepath.Ext(f) == ".js" {
			p = strings.ReplaceAll(p, "!=", "  ")
		}
		for i, line := range strings.Split(p, "\n") {
			if strings.Contains(line, "!") {
				t.Errorf("%s:%d: exclamation mark in prose: %s", relDoc(f), i+1, strings.TrimSpace(line))
			}
		}
	}
}

// retiredDocText is wording the docs must not use: either a fact that was
// wrong or a word the glossary replaced. Patterns match prose only (code
// spans are stripped), case-insensitively.
var retiredDocText = []struct {
	re  *regexp.Regexp
	why string
}{
	// Glossary.
	{regexp.MustCompile(`(?i)\bvillagers?\b`), "the game says worker"},
	{regexp.MustCompile(`(?i)\bfavours?\b`), "timed civilization effects are boons and setbacks (and US spelling)"},
	{regexp.MustCompile(`(?i)\bfavor (deals?|kind)\b|\bFavor:`), "the deal kind that pays opinion is Goodwill"},
	{regexp.MustCompile(`(?i)(^|[^-\w])standings?\b`), "a civilization's attitude is opinion"},
	{regexp.MustCompile(`(?i)\bpop(ulation)? cap\b`), "the limit on workers is housing"},
	{regexp.MustCompile(`(?i)\bmilitary cap\b`), "not a stat the game has"},
	{regexp.MustCompile(`(?i)unlocks \+0\.5x`), "wonders raise the speed cap; nothing speeds up by itself"},
	{regexp.MustCompile(`production_all`), "write all production"},
	{regexp.MustCompile(`(?i)\b(harbour|colours?|centre|civilisation|defence|labour|theatre|organis|cancelled|fervour)`), "US spelling"},
	{regexp.MustCompile(`(?i)\b(economy|epoch|stats|trade|military) tab\b`), "the game has panels, not tabs"},
	// Stale facts.
	{regexp.MustCompile(`(?i)25\s*[-–]\s*75%.{0,20}no effect`), "morale is a continuous curve with no dead zone"},
	{regexp.MustCompile(`×1\.5 per tier`), "worker tiers scale food by 1.12, not 1.5"},
	{regexp.MustCompile(`(?i)advances? automatically`), "the player types advance"},
	{regexp.MustCompile(`(?i)quick-save`), "Esc closes a panel, or saves and returns to the main menu"},
	{regexp.MustCompile(`darwin-arm64|darwin-amd64`), "release assets are named ageforge-macos-*"},
	{regexp.MustCompile(`(?i)clear input`), "Ctrl+C has no handler; it quits"},
	{regexp.MustCompile(`library_of_congress`), "no such building"},
	{regexp.MustCompile(`(?i)reach colonial age`), "first contact comes from expeditions"},
	{regexp.MustCompile(`(?i)prestiiging`), "typo"},
	{regexp.MustCompile(`(?i)MIT License`), "LICENSE is the AgeForge Non-Commercial License"},
	{regexp.MustCompile(`(?i)relative to the directory you launch`), "data lives next to the binary"},
	// AI-writing tells.
	{regexp.MustCompile(`(?i)\b(revolutioni[sz]e|synergistic|synergy|flywheel|unleash|limitless|delve|tapestry)`), "stock word"},
}

func TestDocsNoRetiredText(t *testing.T) {
	for _, f := range playerDocs(t) {
		p := prose(f, readDoc(t, f))
		p = strings.ReplaceAll(p, "villagers.md", "workers-page") // the page's file name stays
		// Names that contain a retired word but mean something else.
		for _, name := range []string{"Standing Stones", "Standing Army"} {
			p = strings.ReplaceAll(p, name, "a game name")
		}
		for _, r := range retiredDocText {
			for _, m := range r.re.FindAllStringIndex(p, -1) {
				line := strings.Split(p, "\n")[lineOf(p, m[0])-1]
				t.Errorf("%s:%d: %q (%s): %s", relDoc(f), lineOf(p, m[0]), strings.TrimSpace(p[m[0]:m[1]]), r.why, strings.TrimSpace(line))
			}
		}
	}
}

// TestDocsNoGoIdentifiers: player docs describe effects, not code. No
// function calls, Go source paths or CamelCase identifiers.
var goIdent = []*regexp.Regexp{
	regexp.MustCompile(`\b[A-Z][A-Za-z]+\.?[A-Za-z]*\(\)`),                   // RecordTrade(), ge.Foo()
	regexp.MustCompile(`\b(game|ui|config|boon|flavor|theme)/[a-z_]+\.go\b`), // game/military.go
	regexp.MustCompile(`\b[A-Z][a-z]+(?:[A-Z][a-z]+)+\b`),                    // TradeBonus, EventManager
}

// camelOK are CamelCase words that are names, not code.
var camelOK = map[string]bool{"AgeForge": true, "PgUp": true, "PgDn": true, "GitHub": true, "PowerShell": true, "WezTerm": true, "MacOS": true, "JavaScript": true, "JetBrains": true}

func TestDocsNoGoIdentifiers(t *testing.T) {
	files, err := filepath.Glob("../site/docs/*.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range append(files, "../README.md") {
		raw := readDoc(t, f)
		raw = fencedCode.ReplaceAllString(raw, "")
		// Look inside inline code and prose alike: `CostScale` is still code.
		for _, re := range goIdent {
			for _, m := range re.FindAllStringIndex(raw, -1) {
				w := raw[m[0]:m[1]]
				if camelOK[w] {
					continue
				}
				t.Errorf("%s:%d: Go identifier %q in a player doc (describe the effect instead)", relDoc(f), lineOf(raw, m[0]), w)
			}
		}
	}
}
