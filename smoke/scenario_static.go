package smoke

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// sectionFrom adds a report section whose body is what write renders, less
// the heading write opens with (trim; "" for none), which the section's own
// title replaces.
func sectionFrom(res *Findings, title, trim string, write func(sb *strings.Builder)) {
	var sb strings.Builder
	write(&sb)
	res.section(title, "%s", strings.TrimPrefix(sb.String(), trim))
}

// failEach records a failure for every problem: check names the check it
// broke, msg says what is wrong.
func failEach[T any](res *Findings, problems []T, check, msg func(T) string) {
	for _, p := range problems {
		res.fail(check(p), "%s", msg(p))
	}
}

// say is a problem as the report prints it: its String.
func say[T any](p T) string { return fmt.Sprint(p) }

// runStatic is the static scenario: every check that reads config alone, with
// a failure for each problem and a report section for each check.
func runStatic(res *Findings) {
	problems := StaticGates()
	failEach(res, problems, func(g GateProblem) string { return "gate_" + g.Kind }, func(g GateProblem) string {
		return fmt.Sprintf("%s -> %s: %s %s (need %s)", g.From, g.To, g.Key, g.Resource, num(g.Need))
	})
	sp, rows := StaticStorage()
	failEach(res, sp, func(p StorageProblem) string { return "storage_" + p.Rule }, say[StorageProblem])
	mp, mr := StaticMilestones()
	failEach(res, mp, func(p MilestoneProblem) string { return "milestone_" + p.Kind }, func(p MilestoneProblem) string { return p.Why })
	ar := StaticAppeaseRules()
	failEach(res, ar, func(p AppeaseRuleProblem) string { return "appease_" + p.Rule }, say[AppeaseRuleProblem])
	br := StaticBraceRules()
	failEach(res, br, func(p BraceRuleProblem) string { return "brace_" + p.Rule }, say[BraceRuleProblem])
	fp, fr := StaticFaithStrength()
	failEach(res, fp, func(p FaithStrengthProblem) string { return "faith_" + p.Rule }, say[FaithStrengthProblem])
	cp, cr := StaticCultureStrength()
	failEach(res, cp, func(p CultureStrengthProblem) string { return "culture_" + p.Rule }, say[CultureStrengthProblem])
	pp, pr := StaticCulturePrices()
	failEach(res, pp, func(p CulturePriceProblem) string { return "culture_price_" + p.Rule }, say[CulturePriceProblem])
	dp := StaticDepth()
	failEach(res, dp, func(string) string { return "depth_points" }, say[string])
	rr := StaticResearch()
	var research []string
	for _, r := range rr {
		research = append(research, r.Problems()...)
	}
	failEach(res, research, func(string) string { return "research_covenant" }, say[string])
	bp, breach, bfirst := staticBadgesFull()
	failEach(res, bp, func(p BadgeProblem) string { return "badge_" + p.Kind }, func(p BadgeProblem) string { return p.Why })
	res.Summary = fmt.Sprintf("%d gate problem(s); %d departure(s) from the storage check; %d milestone problem(s); %d Appease price(s) off the warning; %d Brace price(s) off their warning; %d faith band(s) out of reach; %d culture tier(s) out of reach; %d culture price(s) out of band; %d depth point problem(s); %d research problem(s); %d badge problem(s)", len(problems), len(sp), len(mp), len(ar), len(br), len(fp), len(cp), len(pp), len(dp), len(research), len(bp))
	sectionFrom(res, "Static gate check", "\n## Static gate check\n\n", func(sb *strings.Builder) { writeGates(sb, problems) })
	sectionFrom(res, "Storage check", "", func(sb *strings.Builder) { writeStorage(sb, sp, rows) })
	sectionFrom(res, "Milestone feasibility", "", func(sb *strings.Builder) { writeMilestones(sb, mp, mr) })
	sectionFrom(res, "Harbinger prices against the warning", "", func(sb *strings.Builder) { writeHarbingerPrices(sb, ar, br) })
	sectionFrom(res, "Faith strength", "", func(sb *strings.Builder) { writeFaithStrength(sb, fp, fr) })
	sectionFrom(res, "Culture strength", "", func(sb *strings.Builder) { writeCultureStrength(sb, cp, cr) })
	sectionFrom(res, "Culture prices", "", func(sb *strings.Builder) { writeCulturePrices(sb, pp, pr) })
	sectionFrom(res, "Depth points", "", func(sb *strings.Builder) { writeDepthStatic(sb, dp) })
	sectionFrom(res, "Research Covenant", "", func(sb *strings.Builder) { writeResearch(sb, rr) })
	sectionFrom(res, "Feature locks", "", func(sb *strings.Builder) { writeFeatureLocks(sb, StaticFeatureLocks()) })
	// The model's all-production pool is held to what a game can hold.
	caps := StaticCaps()
	for _, r := range caps {
		if held := config.ProductionAllHeld[r.Age]; held > r.All()+1e-9 {
			res.fail("pool_model", "%s: the pacing model holds +%.0f%% all production, more than the +%.0f%% a player with every milestone, wonder and monument can", r.Age, held*100, r.All()*100)
		}
	}
	sectionFrom(res, "Bonus pools against their caps", "", func(sb *strings.Builder) { writeCaps(sb, caps) })
	// The economy's properties, for the three reference players.
	props := StaticProperties()
	failEach(res, props.Problems(), func(string) string { return "property" }, say[string])
	sectionFrom(res, "The economy's properties, by calculation", "", func(sb *strings.Builder) { writeProperties(sb, props) })
	sectionFrom(res, "Badge Covenant", "\n### Badges\n\n", func(sb *strings.Builder) { writeBadges(sb, bp, breach, bfirst) })
}
