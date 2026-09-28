package smoke

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui"
)

// The docsync scenario checks the player-facing numbers and the command
// reference against the game itself:
//   - every "N ages / buildings / technologies / ..." claim in the landing
//     page (hero stats included), the README and the wiki's front page, and
//     in the opening lines of every wiki page, against config counts;
//   - the lineage count and table rows in buildings.md;
//   - commands.md against the command registry (ui.Commands, which
//     ui.TestRegistryMatchesDispatcher holds to the command handler): every
//     command and subcommand in the registry is documented, every documented
//     command exists, and every documented literal subcommand is one the
//     registry lists.

// Quantity keys.
const (
	qAges        = "ages"
	qBuildings   = "buildings"
	qTechs       = "technologies"
	qMilestones  = "milestones"
	qChains      = "milestone chains"
	qResources   = "resources"
	qEpochs      = "epochs"
	qThemes      = "themes"
	qDomains     = "worker domains"
	qUpgrades    = "prestige upgrades"
	qLineages    = "lineages"
	qWonders     = "wonders"
	qFactions    = "civilizations"
	qRoutes      = "trade routes"
	qExpeditions = "expeditions"
	qLineageBld  = "lineage buildings"
	qStorage     = "storage buildings"
	qMonuments   = "cultural monuments"
	qStandalone  = "standalone buildings"
)

// nonProductionLineages are lineage keys that are not counted as lineages
// in the docs: storage, wonders and monuments are building groups, not
// production lines.
var nonProductionLineages = map[string]bool{"storage": true, "wonder": true, "monument": true}

// GameCounts is every number the docs quote, from config.
func GameCounts() map[string]int {
	c := map[string]int{
		qAges: len(config.Ages()), qBuildings: len(config.BaseBuildings()), qTechs: len(config.Technologies()),
		qMilestones: len(config.Milestones()), qChains: len(config.MilestoneChains()), qResources: len(config.BaseResources()),
		qEpochs: len(config.Epochs()), qThemes: len(theme.All()), qDomains: len(config.WorkerDomains()),
		qUpgrades: len(config.PrestigeUpgrades()), qFactions: len(config.BaseFactions()), qRoutes: len(config.BaseTradeRoutes()),
	}
	lineages := map[string]bool{}
	for _, d := range config.BaseBuildings() {
		switch {
		case d.Category == "wonder":
			c[qWonders]++
		case d.Category == "storage":
			c[qStorage]++
		case d.Category == "monument":
			c[qMonuments]++
		case d.LineageKey != "" && !nonProductionLineages[d.LineageKey]:
			c[qLineageBld]++
			lineages[d.LineageKey] = true
		default:
			c[qStandalone]++
		}
	}
	c[qLineages] = len(lineages)
	order := map[string]int{}
	for i, a := range config.AgeOrder() {
		order[a] = i
	}
	exps := map[string]bool{}
	mm := game.NewMilitaryManager()
	for _, a := range config.AgeOrder() {
		for _, x := range mm.GetAvailableExpeditions(a, order) {
			exps[x.Key] = true
		}
	}
	c[qExpeditions] = len(exps)
	return c
}

// claimWord maps the words after a number to a quantity, with the smallest
// number that counts as a claim about the whole game (so "every 3 ages" or
// "8 buildings" in passing is not one).
var claimWords = []struct {
	re  *regexp.Regexp
	q   string
	min int
}{
	{regexp.MustCompile(`(?i)\b(\d+)[- ]lineage buildings\b`), qLineageBld, 50},
	{regexp.MustCompile(`(?i)\b(\d+)[- ](?:production )?lineages?\b`), qLineages, 5},
	{regexp.MustCompile(`(?i)\b(\d+) (?:total )?buildings\b`), qBuildings, 100},
	{regexp.MustCompile(`(?i)\b(\d+) (?:technologies|techs)\b`), qTechs, 20},
	{regexp.MustCompile(`(?i)\b(\d+) milestones\b`), qMilestones, 20},
	{regexp.MustCompile(`(?i)milestones\W+(\d+) achievements\b`), qMilestones, 20},
	{regexp.MustCompile(`(?i)\b(\d+) resources\b`), qResources, 10},
	{regexp.MustCompile(`(?i)\b(\d+) ages\b`), qAges, 10},
	{regexp.MustCompile(`(?i)\b(\d+) epochs\b`), qEpochs, 5},
	{regexp.MustCompile(`(?i)\b(\d+) themes\b`), qThemes, 5},
	{regexp.MustCompile(`(?i)\b(\d+) (?:worker )?domains\b`), qDomains, 5},
	{regexp.MustCompile(`(?i)\b(\d+) (?:prestige )?upgrades\b`), qUpgrades, 5},
	{regexp.MustCompile(`(?i)\b(\d+) wonders\b`), qWonders, 15},
	{regexp.MustCompile(`(?i)\b(\d+) chains\b`), qChains, 3},
	{regexp.MustCompile(`(?i)\b(\d+) expeditions\b`), qExpeditions, 5},
	{regexp.MustCompile(`(?i)\b(\d+) trade routes\b`), qRoutes, 5},
	{regexp.MustCompile(`(?i)\b(\d+)-civ(?:ilization)?\b`), qFactions, 5},
	{regexp.MustCompile(`(?i)\b(\d+) storage\b`), qStorage, 10},
	{regexp.MustCompile(`(?i)\b(\d+) cultural monuments\b`), qMonuments, 2},
	{regexp.MustCompile(`(?i)\b(\d+) (?:administrative|standalone)\b`), qStandalone, 2},
}

// heroStat is one landing-page hero number: <span class="hstat-n">N</span><span class="hstat-l">Label</span>.
var heroStat = regexp.MustCompile(`(?s)<span class="hstat-n">([^<]+)</span\s*>\s*<span class="hstat-l">([^<]+)</span>`)

var heroLabels = map[string]string{"ages": qAges, "buildings": qBuildings, "techs": qTechs, "technologies": qTechs,
	"milestones": qMilestones, "resources": qResources, "themes": qThemes, "epochs": qEpochs}

// docClaim is one number the docs state.
type docClaim struct {
	file string
	line int
	q    string
	got  int
	text string
}

// headlineScope says which lines of which files are swept for claims:
// whole files for the landing page and the READMEs, the opening lines of
// every other wiki page.
func headlineFiles(root string) map[string]int {
	out := map[string]int{"site/index.html": 0, "README.md": 0, "site/docs/README.md": 0}
	pages, _ := filepath.Glob(filepath.Join(root, "site/docs/*.md"))
	for _, p := range pages {
		rel, _ := filepath.Rel(root, p)
		if _, ok := out[rel]; !ok && !strings.HasPrefix(filepath.Base(p), "_") {
			out[rel] = 5
		}
	}
	return out
}

func docClaims(root string) ([]docClaim, error) {
	var out []docClaim
	files := headlineFiles(root)
	for _, rel := range sortedKeys(files) {
		raw, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(raw), "\n")
		for i, line := range lines {
			if n := files[rel]; n > 0 && i >= n {
				break
			}
			taken := map[int]bool{} // match offsets already claimed by a longer pattern
			for _, w := range claimWords {
				for _, m := range w.re.FindAllStringSubmatchIndex(line, -1) {
					if taken[m[2]] {
						continue
					}
					v, _ := strconv.Atoi(line[m[2]:m[3]])
					if v < w.min {
						continue
					}
					taken[m[2]] = true
					out = append(out, docClaim{file: rel, line: i + 1, q: w.q, got: v, text: strings.TrimSpace(line[m[0]:m[1]])})
				}
			}
		}
		if rel == "site/index.html" {
			for _, m := range heroStat.FindAllStringSubmatchIndex(string(raw), -1) {
				label := strings.ToLower(strings.TrimSpace(string(raw[m[4]:m[5]])))
				q, ok := heroLabels[label]
				if !ok {
					continue
				}
				v, err := strconv.Atoi(strings.TrimSpace(string(raw[m[2]:m[3]])))
				if err != nil {
					continue // "∞" and the like
				}
				line := strings.Count(string(raw[:m[0]]), "\n") + 1
				out = append(out, docClaim{file: rel, line: line, q: q, got: v, text: "hero stat " + label})
			}
		}
	}
	return out, nil
}

// lineageTable finds buildings.md's "The N Production Lineages" heading and
// counts its table rows.
var lineageHeading = regexp.MustCompile(`(?m)^## The (\d+) Production Lineages`)

func lineageTable(root string) (heading, rows int, err error) {
	raw, err := os.ReadFile(filepath.Join(root, "site/docs/buildings.md"))
	if err != nil {
		return 0, 0, err
	}
	s := string(raw)
	m := lineageHeading.FindStringSubmatchIndex(s)
	if m == nil {
		return 0, 0, fmt.Errorf("no \"## The N Production Lineages\" heading in buildings.md")
	}
	heading, _ = strconv.Atoi(s[m[2]:m[3]])
	for _, line := range strings.Split(s[m[1]:], "\n")[1:] {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## ") {
			break
		}
		if len(line) > 2 && line[0] == '|' && line[2] >= '0' && line[2] <= '9' {
			rows++
		}
	}
	return heading, rows, nil
}

// registryCommands reads the command registry (ui.Commands): every command
// (its name, then its aliases) and the words each takes right after its name.
func registryCommands() (cmds [][]string, accepts map[string]map[string]bool) {
	accepts = map[string]map[string]bool{}
	for _, c := range ui.Commands() {
		cmds = append(cmds, c.Names)
		accepts[c.Names[0]] = c.Accepts
	}
	return cmds, accepts
}

// dynamicWords is every game key a command can take that is not a
// subcommand: resources, buildings, techs, themes, upgrades, factions,
// routes, expeditions.
func dynamicWords() map[string]bool {
	out := map[string]bool{}
	for _, r := range config.BaseResources() {
		out[r.Key] = true
	}
	for _, b := range config.BaseBuildings() {
		out[b.Key] = true
	}
	for _, t := range config.Technologies() {
		out[t.Key] = true
	}
	for _, t := range theme.All() {
		out[t.Key] = true
	}
	for _, u := range config.PrestigeUpgrades() {
		out[u.Key] = true
	}
	for _, f := range config.BaseFactions() {
		out[f.Key] = true
	}
	for _, r := range config.BaseTradeRoutes() {
		out[r.Key] = true
	}
	order := map[string]int{}
	for i, a := range config.AgeOrder() {
		order[a] = i
	}
	mm := game.NewMilitaryManager()
	for _, a := range config.AgeOrder() {
		for _, x := range mm.GetAvailableExpeditions(a, order) {
			out[x.Key] = true
		}
	}
	return out
}

// offeredSubcommands is what the registry lists after each command, two
// words deep ("trade route start"), keeping keywords and dropping game keys
// and numbers (gather's food, wood and stone are resources, not
// subcommands).
func offeredSubcommands() map[string][]string {
	dyn := dynamicWords()
	keyword := func(w string) bool {
		if _, err := strconv.ParseFloat(strings.TrimSuffix(w, "x"), 64); err == nil {
			return false
		}
		return !dyn[w]
	}
	out := map[string][]string{}
	for _, c := range ui.Commands() {
		var subs []string
		for _, s := range c.Subs {
			ok := true
			for _, w := range strings.Fields(s) {
				ok = ok && keyword(w)
			}
			if ok {
				subs = append(subs, s)
			}
		}
		sort.Strings(subs)
		out[c.Names[0]] = subs
	}
	return out
}

// documentedForms returns the command spans in the first column of every
// "| Command | ... |" table in commands.md, plus "(or `x`)" aliases.
var codeSpan = regexp.MustCompile("`([^`]+)`")

func documentedForms(md string) (forms []string, all string) {
	inCmdTable := false
	for _, line := range strings.Split(md, "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "| Command |"):
			inCmdTable = true
			continue
		case !strings.HasPrefix(t, "|"):
			inCmdTable = false
			continue
		case !inCmdTable || strings.HasPrefix(t, "|---"):
			continue
		}
		// Split cells on unescaped pipes only: `[count\|max]` is one cell.
		row := strings.ReplaceAll(strings.TrimPrefix(t, "|"), `\|`, "\x00")
		first := strings.SplitN(row, "|", 2)[0]
		for _, m := range codeSpan.FindAllStringSubmatch(first, -1) {
			forms = append(forms, strings.TrimSpace(strings.ReplaceAll(m[1], "\x00", "|")))
		}
	}
	return forms, md
}

// requiredForms are command shapes with no keyword to find them by.
var requiredForms = []struct{ cmd, prefix, what string }{
	{"trade", "`trade <", "the resource exchange, trade <from> <to> <amount>"},
}

func runDocsync(e *Env, res *Result) {
	if e.RepoRoot == "" {
		res.fail("docsync_unavailable", "repository root unknown; run from the repo")
		return
	}
	counts := GameCounts()
	claims, err := docClaims(e.RepoRoot)
	if err != nil {
		res.fail("docsync_read", "%v", err)
		return
	}
	var rows []string
	wrong := 0
	for _, c := range claims {
		want := counts[c.q]
		mark := "✓"
		if c.got != want {
			mark = fmt.Sprintf("✗ (config: %d)", want)
			wrong++
			res.fail("doc_count", "%s:%d says %q but config has %d %s", c.file, c.line, c.text, want, c.q)
		}
		rows = append(rows, fmt.Sprintf("| %s:%d | %s | %d | %s |", c.file, c.line, cell(c.text), c.got, mark))
	}
	heading, tableRows, err := lineageTable(e.RepoRoot)
	if err != nil {
		res.fail("doc_lineages", "%v", err)
	} else {
		if heading != counts[qLineages] {
			res.fail("doc_lineages", "site/docs/buildings.md heading says %d production lineages but config has %d", heading, counts[qLineages])
		}
		if tableRows != counts[qLineages] {
			res.fail("doc_lineages", "site/docs/buildings.md's lineage table has %d rows but config has %d lineages", tableRows, counts[qLineages])
		}
	}

	cmds, accepts := registryCommands()
	mdRaw, err := os.ReadFile(filepath.Join(e.RepoRoot, "site/docs/commands.md"))
	if err != nil {
		res.fail("docsync_read", "%v", err)
		return
	}
	forms, md := documentedForms(string(mdRaw))
	documented := func(form string) bool {
		return strings.Contains(md, "`"+form+"`") || strings.Contains(md, "`"+form+" ")
	}
	alias := map[string]string{}
	for _, names := range cmds {
		for _, n := range names {
			alias[n] = names[0]
		}
	}
	var missing, aliasesMissing []string
	for _, names := range cmds {
		if !documented(names[0]) {
			missing = append(missing, names[0])
			res.fail("doc_command_missing", "commands.md has no `%s` (a registered command)", names[0])
		}
		for _, a := range names[1:] {
			if !documented(a) {
				aliasesMissing = append(aliasesMissing, a)
			}
		}
	}
	offered := offeredSubcommands()
	// A subcommand is documented when some documented form of the command
	// (or an alias) names its words in order, as literals or as choices in
	// brackets: `recruit [count|max]` documents recruit max.
	docTokens := map[string][][]string{}
	for _, form := range forms {
		toks := strings.FieldsFunc(strings.ToLower(form), func(r rune) bool {
			return r == ' ' || r == '[' || r == ']' || r == '|' || r == '\\' || r == '<' || r == '>'
		})
		if len(toks) > 0 {
			if p, ok := alias[toks[0]]; ok {
				docTokens[p] = append(docTokens[p], toks[1:])
			}
		}
	}
	subDocumented := func(cmd, sub string) bool {
		want := strings.Fields(sub)
		for _, toks := range docTokens[cmd] {
			i := 0
			for _, t := range toks {
				if i < len(want) && t == want[i] {
					i++
				}
			}
			if i == len(want) {
				return true
			}
		}
		return false
	}
	for _, names := range cmds {
		for _, sub := range offered[names[0]] {
			if ok := subDocumented(names[0], sub); !ok {
				res.fail("doc_subcommand_missing", "commands.md has no `%s %s` (the command registry lists it)", names[0], sub)
			}
		}
	}
	for _, rf := range requiredForms {
		if !strings.Contains(md, rf.prefix) {
			res.fail("doc_form_missing", "commands.md has no row for %s", rf.what)
		}
	}
	for _, form := range forms {
		if form == "" || !(form[0] >= 'a' && form[0] <= 'z') {
			continue // a key (`Esc`, `↑`), not a typed command
		}
		f := strings.Fields(strings.ToLower(form))
		primary, ok := alias[f[0]]
		if !ok {
			res.fail("doc_command_unknown", "commands.md documents `%s`, but %q is not in the command registry", form, f[0])
			continue
		}
		if len(f) < 2 || !isLiteralWord(f[1]) {
			continue
		}
		acc := accepts[primary]
		off := map[string]bool{}
		for _, s := range offered[primary] {
			off[strings.Fields(s)[0]] = true
		}
		if !acc[f[1]] && !off[f[1]] {
			res.fail("doc_subcommand_unknown", "commands.md documents `%s`, but the registry's `%s` does not take %q", form, primary, f[1])
		}
	}
	res.Summary = fmt.Sprintf("%d number claim(s) checked, %d wrong; %d command(s), %d documented form(s)", len(claims), wrong, len(cmds), len(forms))
	var cs []string
	for _, k := range sortedKeys(counts) {
		cs = append(cs, fmt.Sprintf("%s %d", k, counts[k]))
	}
	res.section("Config counts", "%s", strings.Join(cs, ", "))
	res.section("Number claims", "| where | text | says | verdict |\n|---|---|---|---|\n%s", strings.Join(rows, "\n"))
	if len(aliasesMissing) > 0 {
		res.section("Undocumented aliases (not failing)", "%s", strings.Join(aliasesMissing, ", "))
	}
}

// isLiteralWord is a plain lowercase word, not a placeholder like <key>,
// [count|all] or a number.
func isLiteralWord(s string) bool {
	if s == "" || strings.ContainsAny(s, "<>[]|()") {
		return false
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r == '_') {
			return false
		}
	}
	return true
}
