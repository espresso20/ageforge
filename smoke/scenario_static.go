package smoke

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// runStatic is the static scenario: every check that reads config alone, with
// a failure for each problem and a report section for each check.
func runStatic(res *Findings) {
	problems := StaticGates()
	var sb strings.Builder
	writeGates(&sb, problems)
	for _, g := range problems {
		res.fail("gate_"+g.Kind, "%s -> %s: %s %s (need %s)", g.From, g.To, g.Key, g.Resource, num(g.Need))
	}
	sp, rows := StaticStorage()
	for _, p := range sp {
		res.fail("storage_"+p.Rule, "%s", p)
	}
	mp, mr := StaticMilestones()
	for _, p := range mp {
		res.fail("milestone_"+p.Kind, "%s", p.Why)
	}
	ar := StaticAppeaseRules()
	for _, p := range ar {
		res.fail("appease_"+p.Rule, "%s", p)
	}
	br := StaticBraceRules()
	for _, p := range br {
		res.fail("brace_"+p.Rule, "%s", p)
	}
	fp, fr := StaticFaithStrength()
	for _, p := range fp {
		res.fail("faith_"+p.Rule, "%s", p)
	}
	cp, cr := StaticCultureStrength()
	for _, p := range cp {
		res.fail("culture_"+p.Rule, "%s", p)
	}
	pp, pr := StaticCulturePrices()
	for _, p := range pp {
		res.fail("culture_price_"+p.Rule, "%s", p)
	}
	dp := StaticDepth()
	for _, p := range dp {
		res.fail("depth_points", "%s", p)
	}
	rr := StaticResearch()
	research := 0
	for _, r := range rr {
		for _, p := range r.Problems() {
			research++
			res.fail("research_covenant", "%s", p)
		}
	}
	bp, breach, bfirst := staticBadgesFull()
	for _, p := range bp {
		res.fail("badge_"+p.Kind, "%s", p.Why)
	}
	res.Summary = fmt.Sprintf("%d gate problem(s); %d departure(s) from the storage check; %d milestone problem(s); %d Appease price(s) off the warning; %d Brace price(s) off their warning; %d faith band(s) out of reach; %d culture tier(s) out of reach; %d culture price(s) out of band; %d depth point problem(s); %d research problem(s); %d badge problem(s)", len(problems), len(sp), len(mp), len(ar), len(br), len(fp), len(cp), len(pp), len(dp), research, len(bp))
	res.section("Static gate check", "%s", strings.TrimPrefix(sb.String(), "\n## Static gate check\n\n"))
	var st strings.Builder
	writeStorage(&st, sp, rows)
	res.section("Storage check", "%s", st.String())
	var mf strings.Builder
	writeMilestones(&mf, mp, mr)
	res.section("Milestone feasibility", "%s", mf.String())
	var hf strings.Builder
	writeHarbingerPrices(&hf, ar, br)
	res.section("Harbinger prices against the warning", "%s", hf.String())
	var ff strings.Builder
	writeFaithStrength(&ff, fp, fr)
	res.section("Faith strength", "%s", ff.String())
	var cuf strings.Builder
	writeCultureStrength(&cuf, cp, cr)
	res.section("Culture strength", "%s", cuf.String())
	var cpf strings.Builder
	writeCulturePrices(&cpf, pp, pr)
	res.section("Culture prices", "%s", cpf.String())
	var df strings.Builder
	writeDepthStatic(&df, dp)
	res.section("Depth points", "%s", df.String())
	var rf strings.Builder
	writeResearch(&rf, rr)
	res.section("Research Covenant", "%s", rf.String())
	var lf strings.Builder
	writeFeatureLocks(&lf, StaticFeatureLocks())
	res.section("Feature locks", "%s", lf.String())
	// The model's all-production pool is held to what a game can hold.
	caps := StaticCaps()
	for _, r := range caps {
		if held := config.ProductionAllHeld[r.Age]; held > r.All()+1e-9 {
			res.fail("pool_model", "%s: the pacing model holds +%.0f%% all production, more than the +%.0f%% a player with every milestone, wonder and monument can", r.Age, held*100, r.All()*100)
		}
	}
	var cf strings.Builder
	writeCaps(&cf, caps)
	res.section("Bonus pools against their caps", "%s", cf.String())
	// The economy's properties, for the three reference players.
	props := StaticProperties()
	for _, p := range props.Problems() {
		res.fail("property", "%s", p)
	}
	var pf strings.Builder
	writeProperties(&pf, props)
	res.section("The economy's properties, by calculation", "%s", pf.String())
	var bf strings.Builder
	writeBadges(&bf, bp, breach, bfirst)
	res.section("Badge Covenant", "%s", strings.TrimPrefix(bf.String(), "\n### Badges\n\n"))
}
