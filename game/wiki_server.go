package game

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/user/ageforge/config"
)

const wikiPort = "7891"

// WikiServer hosts a local HTTP wiki for the game.
// It reads only from config (pure data), so it works without a running engine.
type WikiServer struct {
	mu      sync.Mutex
	running bool
	server  *http.Server
}

func NewWikiServer() *WikiServer { return &WikiServer{} }

func (ws *WikiServer) URL() string { return "http://localhost:" + wikiPort }

func (ws *WikiServer) IsRunning() bool {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.running
}

// Start launches the HTTP server in the background. Safe to call multiple times.
func (ws *WikiServer) Start() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.running {
		return nil
	}
	ln, err := net.Listen("tcp", ":"+wikiPort)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", wikiHome)
	mux.HandleFunc("/ages", wikiAges)
	mux.HandleFunc("/buildings", wikiBuildings)
	mux.HandleFunc("/resources", wikiResources)
	mux.HandleFunc("/techs", wikiTechs)
	mux.HandleFunc("/events", wikiEvents)
	mux.HandleFunc("/milestones", wikiMilestones)
	mux.HandleFunc("/prestige", wikiPrestige)
	mux.HandleFunc("/villagers", wikiVillagers)
	mux.HandleFunc("/trade", wikiTrade)
	ws.server = &http.Server{Handler: mux}
	ws.running = true
	go func() {
		_ = ws.server.Serve(ln)
		ws.mu.Lock()
		ws.running = false
		ws.mu.Unlock()
	}()
	return nil
}

// Stop shuts down the HTTP server.
func (ws *WikiServer) Stop() {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.server != nil {
		_ = ws.server.Close()
		ws.server = nil
	}
	ws.running = false
}

// OpenBrowser opens the wiki URL in the default system browser.
func (ws *WikiServer) OpenBrowser() {
	url := ws.URL()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

// ── HTML helpers ─────────────────────────────────────────────────────────────

const wikiCSS = `
* { box-sizing: border-box; }
body { background:#0d1117; color:#c9d1d9; font-family:'Segoe UI',Arial,sans-serif; margin:0; padding:0; line-height:1.5; }
nav { background:#161b22; padding:8px 20px; position:sticky; top:0; z-index:10; border-bottom:2px solid #f0a500; display:flex; align-items:center; flex-wrap:wrap; gap:4px; }
nav .logo { color:#f0a500; font-weight:bold; font-size:1.1em; margin-right:12px; text-decoration:none; }
nav a { color:#c9d1d9; text-decoration:none; padding:4px 10px; border-radius:4px; font-size:0.9em; }
nav a:hover { background:#21262d; color:#f0a500; }
.content { max-width:1300px; margin:0 auto; padding:20px 24px; }
h1 { color:#f0a500; border-bottom:2px solid #f0a500; padding-bottom:8px; margin-top:8px; }
h2 { color:#d29922; margin-top:32px; border-bottom:1px solid #30363d; padding-bottom:4px; }
h3 { color:#79c0ff; }
p { color:#8b949e; }
table { border-collapse:collapse; width:100%; margin-bottom:20px; font-size:0.875em; }
th { background:#21262d; color:#f0a500; padding:8px 10px; text-align:left; border:1px solid #30363d; white-space:nowrap; }
td { border:1px solid #30363d; padding:6px 10px; vertical-align:top; }
tr:nth-child(even) td { background:#111827; }
code { background:#21262d; border-radius:3px; padding:1px 6px; font-size:0.82em; color:#79c0ff; font-family:'Courier New',monospace; }
.badge { display:inline-block; border-radius:4px; padding:2px 7px; font-size:0.78em; font-weight:bold; margin:1px; white-space:nowrap; }
.gold   { background:#3d2600; color:#f0a500; border:1px solid #7d5000; }
.green  { background:#0a3622; color:#3fb950; border:1px solid #1a6032; }
.red    { background:#3a0a0a; color:#f85149; border:1px solid #6a1414; }
.blue   { background:#0d1b36; color:#79c0ff; border:1px solid #1d3060; }
.purple { background:#1a0a3a; color:#d2a8ff; border:1px solid #3a1a6a; }
.gray   { background:#21262d; color:#8b949e; border:1px solid #30363d; }
.teal   { background:#0a2f2a; color:#39d0c8; border:1px solid #1a5048; }
.orange { background:#3d1800; color:#f0956a; border:1px solid #7d3010; }
.desc   { color:#8b949e; font-size:0.85em; font-style:italic; }
.arrow  { color:#30363d; margin:0 4px; }
`

const wikiNav = `<nav>
  <a class="logo" href="/">🏛 AgeForge Wiki</a>
  <a href="/ages">Ages</a>
  <a href="/buildings">Buildings</a>
  <a href="/resources">Resources</a>
  <a href="/techs">Techs</a>
  <a href="/events">Events</a>
  <a href="/milestones">Milestones</a>
  <a href="/prestige">Prestige</a>
  <a href="/villagers">Villagers</a>
  <a href="/trade">Trade</a>
</nav>`

func wikiPage(title, body string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s — AgeForge Wiki</title>
<style>%s</style>
</head>
<body>
%s
<div class="content">%s</div>
</body>
</html>`, title, wikiCSS, wikiNav, body)
}

// ── Format helpers ────────────────────────────────────────────────────────────

func fmtEffect(e config.Effect) string {
	cls := "green"
	sign := "+"
	if e.Value < 0 {
		cls = "red"
		sign = ""
	}
	label := fmt.Sprintf("%s%.4g %s", sign, e.Value, e.Type)
	if e.Target != "" {
		label += " [" + e.Target + "]"
	}
	return fmt.Sprintf(`<span class="badge %s">%s</span>`, cls, label)
}

func fmtEffects(effects []config.Effect) string {
	if len(effects) == 0 {
		return `<span class="desc">—</span>`
	}
	var b strings.Builder
	for _, e := range effects {
		b.WriteString(fmtEffect(e))
	}
	return b.String()
}

func fmtCost(cost map[string]float64) string {
	if len(cost) == 0 {
		return "—"
	}
	keys := make([]string, 0, len(cost))
	for k := range cost {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("<code>%s</code> ×%.4g", k, cost[k]))
	}
	return strings.Join(parts, " ")
}

func fmtBuildReqs(reqs map[string]int) string {
	if len(reqs) == 0 {
		return "—"
	}
	keys := make([]string, 0, len(reqs))
	for k := range reqs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("<code>%s</code> ×%d", k, reqs[k]))
	}
	return strings.Join(parts, " ")
}

func sentimentBadge(s string) string {
	cls := map[string]string{
		"positive": "green",
		"negative": "red",
		"neutral":  "gray",
		"crisis":   "red",
		"boon":     "green",
	}
	c, ok := cls[s]
	if !ok {
		c = "gray"
	}
	return fmt.Sprintf(`<span class="badge %s">%s</span>`, c, s)
}

func catBadge(cat string) string {
	cls := map[string]string{
		"production": "green",
		"housing":    "teal",
		"research":   "blue",
		"military":   "red",
		"storage":    "gray",
		"wonder":     "purple",
	}
	c, ok := cls[cat]
	if !ok {
		c = "gray"
	}
	return fmt.Sprintf(`<span class="badge %s">%s</span>`, c, cat)
}

func ageBadge(age string) string {
	if age == "" {
		return `<span class="desc">—</span>`
	}
	return fmt.Sprintf(`<span class="badge gold">%s</span>`, age)
}

func codeOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return fmt.Sprintf("<code>%s</code>", s)
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// ── Page handlers ─────────────────────────────────────────────────────────────

func wikiHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	body := `<h1>AgeForge Wiki</h1>
<p>A complete in-game reference. All data is from the live game configuration.</p>
<table>
<tr><th>Section</th><th>Contents</th></tr>
<tr><td><a href="/ages">Ages</a></td><td>All 22 civilization ages — requirements and unlocks</td></tr>
<tr><td><a href="/buildings">Buildings</a></td><td>80 buildings (58 standard + 22 wonders) grouped by category</td></tr>
<tr><td><a href="/resources">Resources</a></td><td>21 resources — base storage and unlock age</td></tr>
<tr><td><a href="/techs">Techs</a></td><td>52 technologies grouped by age — prerequisites and bonuses</td></tr>
<tr><td><a href="/events">Events</a></td><td>Random events — sentiment, duration, effects</td></tr>
<tr><td><a href="/milestones">Milestones</a></td><td>33 milestones and 5 chains with title rewards</td></tr>
<tr><td><a href="/prestige">Prestige</a></td><td>9 prestige shop upgrades and tier costs</td></tr>
<tr><td><a href="/villagers">Villagers</a></td><td>8 villager types — gather rates and food costs</td></tr>
<tr><td><a href="/trade">Trade</a></td><td>15 trade routes, exchange rates, and 6 factions</td></tr>
</table>`
	fmt.Fprint(w, wikiPage("Home", body))
}

func wikiAges(w http.ResponseWriter, r *http.Request) {
	ages := config.Ages()
	var b strings.Builder
	b.WriteString(`<h1>Ages</h1>
<table>
<tr><th>#</th><th>Age</th><th>Key</th><th>Description</th><th>Resource Requirements</th><th>Building Requirements</th><th>Unlocks</th></tr>`)
	for _, a := range ages {
		// Build unlocks cell
		var unlockParts []string
		for _, bk := range a.UnlockBuildings {
			unlockParts = append(unlockParts, fmt.Sprintf(`<code>%s</code>`, bk))
		}
		for _, rk := range a.UnlockResources {
			unlockParts = append(unlockParts, fmt.Sprintf(`<span class="badge blue">res:%s</span>`, rk))
		}
		for _, vk := range a.UnlockVillagers {
			unlockParts = append(unlockParts, fmt.Sprintf(`<span class="badge teal">+%s</span>`, vk))
		}
		unlocks := strings.Join(unlockParts, " ")
		if unlocks == "" {
			unlocks = "—"
		}
		b.WriteString(fmt.Sprintf(`<tr>
<td>%d</td>
<td><b>%s</b></td>
<td><code>%s</code></td>
<td class="desc">%s</td>
<td>%s</td>
<td>%s</td>
<td>%s</td>
</tr>`,
			a.Order, a.Name, a.Key, a.Description,
			fmtCost(a.ResourceReqs),
			fmtBuildReqs(a.BuildingReqs),
			unlocks,
		))
	}
	b.WriteString("</table>")
	fmt.Fprint(w, wikiPage("Ages", b.String()))
}

func wikiBuildings(w http.ResponseWriter, r *http.Request) {
	buildings := config.BaseBuildings()
	catOrder := []string{"production", "housing", "research", "military", "storage", "wonder"}
	grouped := make(map[string][]config.BuildingDef)
	for _, bd := range buildings {
		grouped[bd.Category] = append(grouped[bd.Category], bd)
	}
	var b strings.Builder
	b.WriteString("<h1>Buildings</h1>")
	for _, cat := range catOrder {
		list, ok := grouped[cat]
		if !ok {
			continue
		}
		b.WriteString(fmt.Sprintf("<h2>%s %ss</h2>", catBadge(cat), capitalize(cat)))
		b.WriteString(`<table>
<tr><th>Name</th><th>Key</th><th>Age</th><th>Tech</th><th>Base Cost</th><th>Scale</th><th>Build Ticks</th><th>Effects</th><th>Description</th></tr>`)
		for _, bd := range list {
			nameCell := fmt.Sprintf("<b>%s</b>", bd.Name)
			if bd.MaxCount > 0 {
				nameCell += fmt.Sprintf(` <span class="badge gray">max %d</span>`, bd.MaxCount)
			}
			b.WriteString(fmt.Sprintf(`<tr>
<td>%s</td>
<td><code>%s</code></td>
<td>%s</td>
<td>%s</td>
<td>%s</td>
<td>%.2f×</td>
<td>%d</td>
<td>%s</td>
<td class="desc">%s</td>
</tr>`,
				nameCell, bd.Key,
				ageBadge(bd.RequiredAge),
				codeOrDash(bd.RequiredTech),
				fmtCost(bd.BaseCost),
				bd.CostScale, bd.BuildTicks,
				fmtEffects(bd.Effects),
				bd.Description,
			))
		}
		b.WriteString("</table>")
	}
	fmt.Fprint(w, wikiPage("Buildings", b.String()))
}

func wikiResources(w http.ResponseWriter, r *http.Request) {
	resources := config.BaseResources()
	var b strings.Builder
	b.WriteString(`<h1>Resources</h1>
<table>
<tr><th>Name</th><th>Key</th><th>Unlocked In</th><th>Base Storage</th><th>Description</th></tr>`)
	for _, res := range resources {
		b.WriteString(fmt.Sprintf(`<tr>
<td><b>%s</b></td>
<td><code>%s</code></td>
<td>%s</td>
<td>%.0f</td>
<td class="desc">%s</td>
</tr>`,
			res.Name, res.Key,
			ageBadge(res.Age),
			res.BaseStorage, res.Description,
		))
	}
	b.WriteString("</table>")
	fmt.Fprint(w, wikiPage("Resources", b.String()))
}

func wikiTechs(w http.ResponseWriter, r *http.Request) {
	ages := config.Ages()
	techsByAge := config.TechsByAge()
	var b strings.Builder
	b.WriteString("<h1>Technologies</h1>")
	for _, age := range ages {
		techs, ok := techsByAge[age.Key]
		if !ok {
			continue
		}
		b.WriteString(fmt.Sprintf("<h2>%s</h2>", age.Name))
		b.WriteString(`<table>
<tr><th>Name</th><th>Key</th><th>Cost</th><th>Ticks</th><th>Prerequisites</th><th>Effects</th><th>Description</th></tr>`)
		for _, t := range techs {
			prereqs := "—"
			if len(t.Prerequisites) > 0 {
				var ps []string
				for _, p := range t.Prerequisites {
					ps = append(ps, fmt.Sprintf("<code>%s</code>", p))
				}
				prereqs = strings.Join(ps, " ")
			}
			b.WriteString(fmt.Sprintf(`<tr>
<td><b>%s</b></td>
<td><code>%s</code></td>
<td>%.0f</td>
<td>%d</td>
<td>%s</td>
<td>%s</td>
<td class="desc">%s</td>
</tr>`,
				t.Name, t.Key,
				t.Cost, t.ResearchTicks,
				prereqs,
				fmtEffects(t.Effects),
				t.Description,
			))
		}
		b.WriteString("</table>")
	}
	fmt.Fprint(w, wikiPage("Technologies", b.String()))
}

func wikiEvents(w http.ResponseWriter, r *http.Request) {
	events := config.RandomEvents()
	var b strings.Builder
	b.WriteString(`<h1>Random Events</h1>
<table>
<tr><th>Name</th><th>Key</th><th>Sentiment</th><th>Min Age</th><th>Weight</th><th>Duration</th><th>Cooldown</th><th>Min Tick</th><th>Effects</th><th>Description</th></tr>`)
	for _, ev := range events {
		b.WriteString(fmt.Sprintf(`<tr>
<td><b>%s</b></td>
<td><code>%s</code></td>
<td>%s</td>
<td>%s</td>
<td>%d</td>
<td>%d</td>
<td>%d</td>
<td>%d</td>
<td>%s</td>
<td class="desc">%s</td>
</tr>`,
			ev.Name, ev.Key,
			sentimentBadge(ev.Sentiment),
			ageBadge(ev.MinAge),
			ev.Weight, ev.Duration, ev.Cooldown, ev.MinTick,
			fmtEffects(ev.Effects),
			ev.Description,
		))
	}
	b.WriteString("</table>")
	fmt.Fprint(w, wikiPage("Events", b.String()))
}

func wikiMilestones(w http.ResponseWriter, r *http.Request) {
	milestones := config.Milestones()
	chains := config.MilestoneChains()
	catOrder := config.MilestoneCategoryOrder()
	catNames := config.MilestoneCategoryNames()

	grouped := make(map[string][]config.MilestoneDef)
	for _, m := range milestones {
		grouped[m.Category] = append(grouped[m.Category], m)
	}

	var b strings.Builder
	b.WriteString("<h1>Milestones</h1>")
	for _, cat := range catOrder {
		list, ok := grouped[cat]
		if !ok {
			continue
		}
		name := catNames[cat]
		if name == "" {
			name = cat
		}
		b.WriteString(fmt.Sprintf("<h2>%s</h2>", name))
		b.WriteString(`<table>
<tr><th>Name</th><th>Key</th><th>Min Age</th><th>Hidden</th><th>Requirements</th><th>Rewards</th><th>Description</th></tr>`)
		for _, m := range list {
			var reqParts []string
			if m.MinAge != "" {
				reqParts = append(reqParts, ageBadge(m.MinAge))
			}
			if m.MinTick > 0 {
				reqParts = append(reqParts, fmt.Sprintf(`<span class="badge gray">tick≥%d</span>`, m.MinTick))
			}
			if m.MinPopulation > 0 {
				reqParts = append(reqParts, fmt.Sprintf(`<span class="badge teal">pop≥%d</span>`, m.MinPopulation))
			}
			if m.MinTechCount > 0 {
				reqParts = append(reqParts, fmt.Sprintf(`<span class="badge blue">techs≥%d</span>`, m.MinTechCount))
			}
			for _, rt := range m.RequiredTechs {
				reqParts = append(reqParts, fmt.Sprintf(`<code>%s</code>`, rt))
			}
			if len(m.MinResources) > 0 {
				reqParts = append(reqParts, fmtCost(m.MinResources))
			}
			if len(m.MinBuildings) > 0 {
				reqParts = append(reqParts, fmtBuildReqs(m.MinBuildings))
			}
			req := strings.Join(reqParts, " ")
			if req == "" {
				req = "—"
			}
			hidden := "—"
			if m.Hidden {
				hidden = `<span class="badge gray">hidden</span>`
			}
			b.WriteString(fmt.Sprintf(`<tr>
<td><b>%s</b></td>
<td><code>%s</code></td>
<td>%s</td>
<td>%s</td>
<td>%s</td>
<td>%s</td>
<td class="desc">%s</td>
</tr>`,
				m.Name, m.Key,
				ageBadge(m.MinAge),
				hidden, req,
				fmtEffects(m.Rewards),
				m.Description,
			))
		}
		b.WriteString("</table>")
	}

	// Chains
	b.WriteString(`<h2>Milestone Chains</h2>
<p>Complete all milestones in a chain to earn its title and a temporary tick-speed boost.</p>
<table>
<tr><th>Chain</th><th>Category</th><th>Title Reward</th><th>Speed Boost</th><th>Boost Duration</th><th>Milestones (in order)</th></tr>`)
	for _, c := range chains {
		var msList []string
		for _, mk := range c.MilestoneKeys {
			msList = append(msList, fmt.Sprintf("<code>%s</code>", mk))
		}
		b.WriteString(fmt.Sprintf(`<tr>
<td><b>%s</b></td>
<td>%s</td>
<td><span class="badge purple">%s</span></td>
<td><span class="badge green">+%.4g speed</span></td>
<td>%d ticks</td>
<td>%s</td>
</tr>`,
			c.Name, c.Category,
			c.Title,
			c.BoostValue, c.BoostDuration,
			strings.Join(msList, ` <span class="arrow">→</span> `),
		))
	}
	b.WriteString("</table>")
	fmt.Fprint(w, wikiPage("Milestones", b.String()))
}

func wikiPrestige(w http.ResponseWriter, r *http.Request) {
	upgrades := config.PrestigeUpgrades()
	var b strings.Builder
	b.WriteString(`<h1>Prestige</h1>
<p>Prestige resets your run from the Primitive Age, awarding Prestige Points to spend in the shop. Prestige is unlocked after reaching <b>Modern Age</b>.</p>
<table>
<tr><th>Name</th><th>Key</th><th>Effect</th><th>Per Tier</th><th>Max Tier</th><th>Costs (tier 1→N)</th><th>Description</th></tr>`)
	for _, u := range upgrades {
		costs := make([]string, len(u.Costs))
		for i, c := range u.Costs {
			costs[i] = fmt.Sprintf("%d", c)
		}
		effectDesc := fmt.Sprintf(`<code>%s</code> <span class="badge blue">%s</span>`, u.EffectKey, u.EffectType)
		b.WriteString(fmt.Sprintf(`<tr>
<td><b>%s</b></td>
<td><code>%s</code></td>
<td>%s</td>
<td><span class="badge green">+%.4g/tier</span></td>
<td>%d</td>
<td>%s</td>
<td class="desc">%s</td>
</tr>`,
			u.Name, u.Key,
			effectDesc,
			u.PerTier, u.MaxTier,
			strings.Join(costs, " → "),
			u.Description,
		))
	}
	b.WriteString("</table>")
	fmt.Fprint(w, wikiPage("Prestige", b.String()))
}

func wikiVillagers(w http.ResponseWriter, r *http.Request) {
	villagers := DefaultVillagerTypes()
	var b strings.Builder
	b.WriteString(`<h1>Villager Types</h1>
<p>Villagers are assigned to gather resources. Each type costs food per tick and has a base gather rate.</p>
<table>
<tr><th>Name</th><th>Key</th><th>Food/tick</th><th>Base Gather Rate</th><th>Can Gather</th></tr>`)
	for _, v := range villagers {
		gathers := "—"
		if len(v.CanGather) > 0 {
			var gs []string
			for _, g := range v.CanGather {
				gs = append(gs, fmt.Sprintf("<code>%s</code>", g))
			}
			gathers = strings.Join(gs, " ")
		}
		b.WriteString(fmt.Sprintf(`<tr>
<td><b>%s</b></td>
<td><code>%s</code></td>
<td>%.2f</td>
<td><span class="badge green">%.2f/tick</span></td>
<td>%s</td>
</tr>`,
			v.Name, v.Key,
			v.FoodCost, v.GatherRate,
			gathers,
		))
	}
	b.WriteString("</table>")
	fmt.Fprint(w, wikiPage("Villagers", b.String()))
}

func wikiTrade(w http.ResponseWriter, r *http.Request) {
	routes := config.BaseTradeRoutes()
	rates := config.BaseExchangeRates()
	factions := config.BaseFactions()
	var b strings.Builder
	b.WriteString("<h1>Trade</h1>")

	// Trade Routes
	b.WriteString(`<h2>Trade Routes</h2>
<table>
<tr><th>Name</th><th>Key</th><th>Min Age</th><th>Req Building</th><th>Min Count</th><th>Ticks/Run</th><th>Export</th><th>Import</th><th>Description</th></tr>`)
	for _, rt := range routes {
		b.WriteString(fmt.Sprintf(`<tr>
<td><b>%s</b></td>
<td><code>%s</code></td>
<td>%s</td>
<td>%s</td>
<td>%d</td>
<td>%d</td>
<td>%s</td>
<td>%s</td>
<td class="desc">%s</td>
</tr>`,
			rt.Name, rt.Key,
			ageBadge(rt.MinAge),
			codeOrDash(rt.RequiredBld),
			rt.MinCount, rt.TicksPerRun,
			fmtCost(rt.Export),
			fmtCost(rt.Import),
			rt.Description,
		))
	}
	b.WriteString("</table>")

	// Exchange Rates
	b.WriteString(`<h2>Exchange Rates</h2>
<table>
<tr><th>From</th><th>To</th><th>Base Rate</th><th>Min Age</th></tr>`)
	for _, er := range rates {
		b.WriteString(fmt.Sprintf(`<tr>
<td><code>%s</code></td>
<td><code>%s</code></td>
<td>%.4g</td>
<td>%s</td>
</tr>`,
			er.From, er.To,
			er.BaseRate,
			ageBadge(er.MinAge),
		))
	}
	b.WriteString("</table>")

	// Factions
	b.WriteString(`<h2>Factions</h2>
<p>Build diplomatic relations to unlock trade bonuses. Opinion starts at 0 and changes through diplomacy actions.</p>
<table>
<tr><th>Name</th><th>Key</th><th>Min Age</th><th>Specialty</th><th>Trade Bonus (Allied)</th><th>Description</th></tr>`)
	for _, f := range factions {
		b.WriteString(fmt.Sprintf(`<tr>
<td><b>%s</b></td>
<td><code>%s</code></td>
<td>%s</td>
<td><code>%s</code></td>
<td><span class="badge green">+%.4g</span></td>
<td class="desc">%s</td>
</tr>`,
			f.Name, f.Key,
			ageBadge(f.MinAge),
			f.Specialty,
			f.TradeBonus,
			f.Description,
		))
	}
	b.WriteString("</table>")
	fmt.Fprint(w, wikiPage("Trade", b.String()))
}
