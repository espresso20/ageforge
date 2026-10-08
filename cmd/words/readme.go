package main

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/flavor"
)

// readme.go writes the one-page guide that goes out with the sheets.

// flavorMoments ties each catalog to the flavor package's own name for it,
// so the guide can ask the real generator for assembled lines.
var flavorMoments = map[string]flavor.Moment{
	"exp_success":      flavor.ExpeditionSuccess,
	"exp_fail":         flavor.ExpeditionFailure,
	"enc_standoff":     flavor.EncounterStandoff,
	"enc_cap":          flavor.EncounterAtCapacity,
	"war_raid":         flavor.WarRaid,
	"harb_arrival":     flavor.HarbingerArrival,
	"harb_warn":        flavor.HarbingerWarning,
	"harb_appeased":    flavor.HarbingerAppeased,
	"harb_braced":      flavor.HarbingerBraced,
	"harb_invited":     flavor.HarbingerInvited,
	"harb_vindicated":  flavor.HarbingerVindicated,
	"harb_spared":      flavor.HarbingerSpared,
	"harb_discredited": flavor.HarbingerDiscredited,
	"harb_fulfilled":   flavor.HarbingerFulfilled,
	"run_end":          flavor.RunEnding,
}

// example is one line as the game assembles it.
type example struct {
	from *row   // the sentence it was made from
	line string // the line as the log shows it
}

// request builds the request the game would send for a catalog, in an early
// age where the catalog can happen at all.
func (c *catalog) request(mo *moment) (flavor.Request, string) {
	req := flavor.Request{Moment: flavorMoments[mo.prefix]}
	switch mo.about {
	case "expedition":
		req.Age = "bronze_age"
		req.Kind = "scouting"
		req.Resource, req.Amount = "food", 60
		if exps := config.BadgeExpeditions(); len(exps) > 0 {
			req.Subject = exps[0].Name
		}
	case "civilization":
		req.Age = "iron_age"
		if fs := config.BaseFactions(); len(fs) > 0 {
			req.Subject, req.Kind = fs[0].Name, fs[0].Personality
		}
		req.Resource, req.Amount = "food", 40
	case "risk":
		req.Age = "medieval_age"
		req.Kind = flavor.TierHigh
		for _, h := range config.Harbingers() {
			if h.Age == req.Age {
				req.Subject = h.Name
			}
		}
	default:
		req.Age = "modern_age"
	}
	return req, c.w.called(req.Age)
}

// examples asks the flavor package for lines of one catalog and returns up
// to n, preferring sentences that are filled in from a word list or a slot.
func (c *catalog) examples(mo *moment, n int) ([]example, string) {
	req, age := c.request(mo)
	stream := flavor.NewStream()
	rng := rand.New(rand.NewSource(1))
	var filledIn, plain []example
	seen := map[*row]bool{}
	for i := 0; i < 400 && len(filledIn) < n; i++ {
		res := stream.Generate(req, rng)
		r := c.fl.byTmpl[res.Template]
		if r == nil || seen[r] || res.Text == "" {
			continue
		}
		seen[r] = true
		ex := example{r, res.Text}
		if strings.ContainsAny(r.current, "~{") {
			filledIn = append(filledIn, ex)
		} else {
			plain = append(plain, ex)
		}
	}
	out := append(filledIn, plain...)
	if len(out) > n {
		out = out[:n]
	}
	return out, age
}

// readme returns the guide.
func readme(c *catalog) string {
	var b strings.Builder
	p := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }

	p("# The game's words, as spreadsheets")
	p("")
	p("Every line a player can read in AgeForge is a row in one of these files. Rewrite a line by typing your version next to it, then run the import and the game's source is changed to match. Nothing in this folder is the game itself. Run the export again whenever you like: what you have typed is carried over to the new sheets.")
	p("")
	p("## The files")
	p("")
	p("| File | What is in it | Rows | Words |")
	p("|---|---|---:|---:|")
	rows, words := 0, 0
	for _, a := range areas {
		n, w := 0, 0
		for _, r := range c.rows {
			if r.area == a {
				n++
				w += countWords(r.current)
			}
		}
		p("| `%s` | %s | %d | %d |", a.file, a.about, n, w)
		rows += n
		words += w
	}
	p("| | Total | %d | %d |", rows, words)
	p("")
	first := 0
	for _, r := range c.rows {
		if inFirstHours(r) {
			first++
		}
	}
	p("`%s` is where to start. It holds the %d rows a new player meets from the main menu through the %s, with what is on screen most of the time first: the main window and the menu, then the names of things, then descriptions and panels, then log lines and refusals, then one-off stories, and last the lines drawn from pools, where each line turns up rarely. Its rows are the same rows as in the other files, with the same `id`. Type a rewrite in either place and it counts. The save-name words and the rows of `%s` are left out of it.", firstHoursFile, first, c.w.ageNames[firstHoursAges-1], areaUnsure.file)
	p("")
	p("## The columns")
	p("")
	p("- `id`: the row's name. It says where the text lives in the source. Do not change it.")
	p("- `where`: when a player sees this line.")
	p("- `kind`: `voice` (descriptions, flavor, story), `message` (log lines, refusals, hints, help), `label` (menu entries, headings, short interface words) or `name` (the name of a building, tech, badge and so on).")
	p("- `age`: the age a player first meets the line in, numbered so the column sorts to the early game. `00 Main menu` is before the first age. For text in a panel or a command's reply it is the age the panel or command first works in, which can be earlier than the line itself. Blank when the tool cannot tell.")
	p("- `keep`: what your version must carry over unchanged (see the rules).")
	p("- `max`: the longest the line may be, in characters. Blank when nothing sets a limit.")
	p("- `current`: the line as it is now. Do not change it: the import uses it to check that the source still says what you were looking at.")
	p("- `yours`: your version. This is the only column you fill.")
	p("- `notes`: yours, for anything. The tool never reads it.")
	p("- `same_as`: when an earlier row has exactly the same text, that row's `id`.")
	p("")
	p("## The rules")
	p("")
	p("1. Leave `yours` blank to keep a line as it is.")
	p("2. Everything in `keep` must be in your version, each piece as many times as it is listed:")
	p("   - format verbs such as `%%s`, `%%d` and `%%.1f`, which the game replaces with a name or a number. They must also stay in the same order.")
	p("   - slots such as `{name}` and `{count}`, and the `~` of a story sentence, which a word list fills.")
	p("   - style tags such as `[red]`, `[gold::b]` and `[-]`. A tag colors what follows it and `[-]` ends it, so keep each pair around the words it should color.")
	p("   - key names such as `Enter` and `Esc`, and commands in quotes such as `'help'`.")
	p("   - glyphs such as `★` and `→`.")
	p("   - `\\n`, which stands for a line break inside the cell.")
	p("3. Stay within `max`.")
	p("4. Spaces at the start and end of a line are put back for you. Many short rows are pieces the game joins to a number or a name (`\" idle\"`), and the space is part of the join.")
	p("5. A row with a `same_as` and a blank `yours` takes the rewrite of the row it points to, so one rewrite of `Cancel` changes every `Cancel`. Fill its own `yours` to make it differ, or run the import with `--no-same` to switch this off.")
	p("6. The game's own wording rules apply to your version, and the import refuses a row that breaks one: no em dashes and no routine exclamation marks, US spelling, the glossary (worker, not villager; opinion, not standing), and for the story sentences the catalog's banned sentence shapes, its words that belong to one era only, and a length of 3 to 45 words.")
	p("7. Save each file as CSV in UTF-8, under its own name. The import reads every `.csv` in the folder.")
	p("")
	p("## Renaming something")
	p("")
	p("A `name` row is the name of a building, tech, badge, age and so on. The game looks things up by key, never by name, so a rename is safe for saves. It does show up in other places, which the import lists for you after it applies the row: wiki pages under `site/docs`, the landing page, and tests that quote the old name. The wiki's tech tables are regenerated for you. For a line that is not a name, the import lists the wiki pages that still quote it word for word. The other pages are yours to edit, and the wiki's pictures of the game are redrawn with `go test -tags mapcapture -run TestWriteSiteScreens ./ui`.")
	p("")
	p("Three kinds of name have a second copy that must match, and `same_as` keeps them together: an expedition's name (the badge case has its own copy), a theme's name (the same), and a title a badge gives. A trade route's name and description also decide how the route travels on the map: the map looks for words such as rail, ship, port, air and warp.")
	p("")
	p("## How the story lines are put together")
	p("")
	p("The story lines in the `flavor-*.csv` files are whole sentences. The game never joins two of them. For each line it picks one sentence from the pools that fit the moment and the age, and then:")
	p("")
	p("- if the sentence has a `~`, one entry of the word list named in `where` goes in its place (the entries are the `flavor.words.*` rows);")
	p("- a slot in braces is filled in: `{subject}` with the name of the expedition, civilization or harbinger, `{res}` with a resource, `{amt_res}` with an amount and its resource;")
	p("- a full stop is added, so a sentence is written without one.")
	p("")
	p("A sentence in an `any` pool can turn up in every age, so it must not name anything that belongs to one era. The other pools are held to their ages. Each catalog also has to keep its mix: about one sentence in five of 3 to 6 words, about one in ten over 33, and at least three in ten told flat with no joke. `go test ./flavor` checks the mix after an import, so a batch of rewrites that all get longer or all get shorter can fail it even when every row is fine.")
	p("")
	for _, mo := range moments {
		var pools, ofAge []string
		total, perAgeN := 0, 0
		for _, pl := range c.fl.order {
			if pl.m != mo || pl.size == 0 {
				continue
			}
			total += pl.size
			entry := fmt.Sprintf("`%s` (%s) %d", pl.bucket, poolName(c.w, pl), pl.size)
			if _, isAge := c.w.ageIdx[pl.bucket]; isAge {
				ofAge = append(ofAge, entry)
				perAgeN += pl.size
				continue
			}
			pools = append(pools, entry)
		}
		if len(ofAge) > 4 {
			pools = append(pools, fmt.Sprintf("and %d pools each held to one age and named for it (`stone_age`, `iron_age` and so on), %d sentences between them", len(ofAge), perAgeN))
		} else {
			pools = append(pools, ofAge...)
		}
		p("### %s", mo.title)
		p("")
		p("`flavor.%s.*` in `%s`: %d sentences, shown %s. Pools: %s.", mo.key, mo.area.file, total, mo.when, strings.Join(pools, "; "))
		p("")
		exs, age := c.examples(mo, 3)
		if len(exs) > 0 {
			p("Three lines as the game puts them together in the %s:", age)
			p("")
			for _, ex := range exs {
				p("- `%s` “%s” is shown as “%s”", ex.from.id, ex.from.current, ex.line)
			}
			p("")
		}
	}
	p("## Other text put together from pieces")
	p("")
	p("- Short rows whose `where` says “one piece of a line the game puts together” are joined in code to names and numbers. The `where` shows the whole line with … for what the game fills in. Rewrite the pieces of one line together.")
	p("- A building's description is followed by what the building does, built from the rows of `config.effect_text.*` (“+5 housing”, “storage for every resource”). A tech's card does the same from `config.tech_effects.*` and the `mechanic.*` rows.")
	p("- The refusal when a command waits for a tech is four rows of `feature.*` around the tech's name: “Trade routes need The Wheel first. Research it to start one.”")
	p("- A boon's line (`boon.*`) has slots for what it gives: `{pct}`, `{res}`, `{amt}`, `{n}`, `{ticks}`.")
	p("- A badge family's description (`badge_family.*`, `badge_ladder.*`) has slots the game fills for each badge: `{name}`, `{count}`.")
	p("- A suggested save name is two or three entries of `%s` side by side, and the map's name for your world is two or three syllables of `mapmodel.world.*`.", areaSaveNames.file)
	p("")
	p("## The commands")
	p("")
	p("```")
	p("go run ./cmd/words import --dry-run   # what would change and what would be refused; writes nothing")
	p("go run ./cmd/words import             # apply every filled row, then build the game")
	p("```")
	p("")
	p("The import prints each row it refuses and why, applies the rest, and builds the game. It does not run the tests: run `go test ./...` next. `go run ./cmd/words status` shows how much is rewritten and how much is left, and `go run ./cmd/words export` writes these files again from the source as it now stands, keeping what you have typed.")
	return b.String()
}

// poolName says which ages a pool fits, for the guide.
func poolName(w *world, pl *flavorPool) string {
	lo, hi := 1, len(w.ageKeys)
	switch {
	case len(pl.ages) > 0:
		lo, hi = len(w.ageKeys), 1
		for _, a := range pl.ages {
			if i := w.ageIdx[a]; i > 0 {
				lo, hi = min(lo, i), max(hi, i)
			}
		}
		if !contiguous(w, pl.ages) {
			var names []string
			for _, a := range pl.ages {
				names = append(names, strings.TrimSuffix(w.called(a), " Age"))
			}
			return joinAnd(names)
		}
	case len(pl.spans) > 0:
		lo, hi = len(w.ageKeys), 1
		for _, sp := range pl.spans {
			lo, hi = min(lo, flavorSpans[sp][0]), max(hi, flavorSpans[sp][1])
		}
	}
	short := func(i int) string { return strings.TrimSuffix(w.ageNames[i-1], " Age") }
	switch {
	case lo == 1 && hi == len(w.ageKeys):
		return "any age"
	case lo == hi:
		return short(lo)
	}
	return short(lo) + " to " + short(hi)
}
