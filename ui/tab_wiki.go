package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/user/ageforge/config"
	"github.com/user/ageforge/game"
)

// WikiTab is a full help/wiki system with live game data
type WikiTab struct {
	root        *tview.Flex
	nav         *tview.TextView
	content     *tview.TextView
	pages       []wikiPage
	current     int
	lastRendered int
}

type wikiPage struct {
	title   string
	render  func(state game.GameState) string
}

// NewWikiTab creates the wiki tab
func NewWikiTab() *WikiTab {
	t := &WikiTab{lastRendered: -1}

	t.pages = []wikiPage{
		{title: "Overview", render: wikiOverview},
		{title: "Getting Started", render: wikiGettingStarted},
		{title: "Resources", render: wikiResources},
		{title: "Buildings", render: wikiBuildings},
		{title: "Villagers", render: wikiVillagers},
		{title: "Research", render: wikiResearch},
		{title: "Military", render: wikiMilitary},
		{title: "Ages", render: wikiAges},
		{title: "Events", render: wikiEvents},
		{title: "Prestige", render: wikiPrestige},
		{title: "Commands", render: wikiCommands},
		{title: "Tips & Strategy", render: wikiStrategy},
	}

	t.nav = tview.NewTextView().
		SetDynamicColors(true)
	t.nav.SetBorder(true).SetTitle(" Wiki ").SetTitleColor(ColorTitle)

	t.content = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	t.content.SetBorder(true)

	t.root = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(t.nav, 22, 0, false).
		AddItem(t.content, 0, 1, false)

	return t
}

// PrevPage moves to the previous wiki page
func (t *WikiTab) PrevPage() {
	if t.current > 0 {
		t.current--
		t.content.ScrollToBeginning()
	}
}

// NextPage moves to the next wiki page
func (t *WikiTab) NextPage() {
	if t.current < len(t.pages)-1 {
		t.current++
		t.content.ScrollToBeginning()
	}
}

// GoToPage jumps to a specific page by index
func (t *WikiTab) GoToPage(idx int) {
	if idx >= 0 && idx < len(t.pages) {
		t.current = idx
		t.content.ScrollToBeginning()
	}
}

// ScrollUp scrolls content up
func (t *WikiTab) ScrollUp() {
	row, col := t.content.GetScrollOffset()
	t.content.ScrollTo(row-10, col)
}

// ScrollDown scrolls content down
func (t *WikiTab) ScrollDown() {
	row, col := t.content.GetScrollOffset()
	t.content.ScrollTo(row+10, col)
}

// Root returns the root primitive
func (t *WikiTab) Root() tview.Primitive {
	return t.root
}

// Refresh updates the wiki with live game data
func (t *WikiTab) Refresh(state game.GameState) {
	// Update nav
	var nav strings.Builder
	for i, p := range t.pages {
		if i == t.current {
			fmt.Fprintf(&nav, " [black:gold] %d %s [-:-]\n", i+1, p.title)
		} else {
			fmt.Fprintf(&nav, " [gray]%d[-] %s\n", i+1, p.title)
		}
	}
	nav.WriteString("\n [gray]↑↓ to navigate[-]")
	nav.WriteString("\n [gray]PgUp/PgDn to scroll[-]")
	t.nav.SetText(nav.String())

	// Update content — only reset scroll when page changes
	pageChanged := t.current != t.lastRendered
	t.content.SetTitle(fmt.Sprintf(" %s ", t.pages[t.current].title))
	t.content.SetText(t.pages[t.current].render(state))
	if pageChanged {
		t.content.ScrollToBeginning()
		t.lastRendered = t.current
	}
}

// ===== Wiki Pages =====

func wikiOverview(_ game.GameState) string {
	return `[gold]AgeForge — A CLI Idle Empire Builder[-]

AgeForge is an idle/clicker game where you build a civilization
from nothing. Starting in the Primitive Age with just your
bare hands, you gather resources, build structures, recruit
villagers, and advance through 22 ages — from primitive
survival to galactic transcendence.

[gold]Core Loop[-]
  1. [cyan]Gather[-] resources manually
  2. [cyan]Build[-] structures for housing and production
  3. [cyan]Recruit[-] villagers and assign them to tasks
  4. [cyan]Research[-] technologies for permanent bonuses
  5. [cyan]Send expeditions[-] for loot and resources
  6. [cyan]Advance[-] to the next age when requirements are met

[gold]Key Concepts[-]
  • [yellow]Resources[-] are capped by storage. Build storage to hold more.
  • [yellow]Villagers[-] eat food every tick. Balance food workers vs others.
  • [yellow]Buildings[-] cost more each time (scaling costs).
  • [yellow]Ages[-] require both resources AND buildings to advance.
  • [yellow]Wonders[-] are unique buildings that take many ticks to build.
  • [yellow]Research[-] costs knowledge and unlocks permanent bonuses.
  • [yellow]Events[-] happen randomly — some help, some hurt.
  • [yellow]Milestones[-] reward permanent bonuses for achievements.
  • [yellow]Prestige[-] lets you reset at Medieval+ for permanent bonuses.

[gold]Game Speed[-]
  The game ticks every 2 seconds. Production and consumption
  happen each tick. This is designed as a long-term idle game —
  later ages take weeks or months to reach.

  Your game auto-saves when you press ESC. Use 'save' and
  'load' commands for manual saves.`
}

func wikiGettingStarted(_ game.GameState) string {
	return `[gold]Getting Started[-]

You begin in the [yellow]Primitive Age[-] with 15 food, 12 wood,
and nothing else. Here's your first steps:

[gold]Step 1: Gather Wood[-]
  Type: [cyan]gather wood[-]
  You get 3 wood per gather. You need more for your first stash.

[gold]Step 2: Build a Stash[-]
  Type: [cyan]build stash[-]
  Stashes provide +100 storage. You need more storage before
  you can hold enough resources for the Stone Age (500 food).

[gold]Step 3: Build a Hut[-]
  Type: [cyan]build hut[-]
  Huts provide +2 population capacity. You need housing
  before you can recruit villagers.

[gold]Step 4: Build an Altar[-]
  Type: [cyan]build altar[-]
  Altars slowly generate knowledge (+0.01/tick). You need
  200 knowledge and 5 altars for the Stone Age.

[gold]Step 5: Recruit Workers & Shamans[-]
  Type: [cyan]recruit worker[-] or [cyan]recruit shaman[-]
  Workers gather food/wood (0.35/tick). Shamans gather
  knowledge (0.08/tick). You need both to advance.

[gold]Step 6: Assign Villagers[-]
  Type: [cyan]assign worker food[-] or [cyan]assign shaman knowledge[-]
  Each worker eats 0.10 food/tick, shamans eat 0.2/tick.
  1 food worker can sustain 1 shaman with food to spare.

[gold]Step 7: Keep Building[-]
  Build more huts, stashes, and altars. Recruit and assign
  villagers. Watch the age progress bar — when all
  requirements turn [green]green[-], you'll advance!

[gold]The Food Balance[-]
  This is the most important early game concept:
  • Each worker eats [yellow]0.10 food/tick[-]
  • Each worker gathers [yellow]0.35 resource/tick[-]
  • So 1 food worker produces net +0.25 for others
  • 1 food worker covers ~1 shaman (0.20 food) or 2.5 workers

[gold]Storage Matters[-]
  Base storage is only [yellow]50[-] per resource. Resources stop
  accumulating when storage is full! Build Stashes
  (Primitive Age), Storage Pits (Stone Age), and Warehouses
  (Bronze Age) to increase caps.`
}

func wikiResources(state game.GameState) string {
	var sb strings.Builder
	sb.WriteString("[gold]Resources[-]\n\n")
	sb.WriteString("Resources are the core currency of the game. Each has\n")
	sb.WriteString("a storage cap — once full, production is wasted.\n\n")

	defs := config.BaseResources()
	for _, def := range defs {
		rs, exists := state.Resources[def.Key]
		sb.WriteString(fmt.Sprintf("[cyan]%s[-] [gray](%s)[-]\n", def.Name, def.Key))
		sb.WriteString(fmt.Sprintf("  %s\n", def.Description))
		sb.WriteString(fmt.Sprintf("  Unlocks in: [yellow]%s[-]\n", def.Age))
		if exists && rs.Unlocked {
			sb.WriteString(fmt.Sprintf("  [orange]Current: %.0f / %.0f  Rate: %s[-]\n",
				rs.Amount, rs.Storage, FormatRate(rs.Rate)))
		} else {
			sb.WriteString("  [gray]Not yet unlocked[-]\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("[gold]Storage[-]\n")
	sb.WriteString("  Base storage starts at 10-50 per resource.\n")
	sb.WriteString("  Build storage buildings to increase caps:\n")
	sb.WriteString("  • [cyan]Stash[-] (Primitive Age): +30 all resources\n")
	sb.WriteString("  • [cyan]Storage Pit[-] (Stone Age): +50 all resources\n")
	sb.WriteString("  • [cyan]Warehouse[-] (Bronze Age): +150 all resources\n")
	sb.WriteString("  • [cyan]Granary[-] (Iron Age): +200 food only\n")
	sb.WriteString("  Each age has a dedicated storage building with\n")
	sb.WriteString("  increasing capacity to handle scaling costs.\n")

	return sb.String()
}

func wikiBuildings(state game.GameState) string {
	var sb strings.Builder
	sb.WriteString("[gold]Buildings[-]\n\n")
	sb.WriteString("Buildings provide production, housing, storage, and more.\n")
	sb.WriteString("Each building costs more than the last (scaling costs).\n\n")

	// Group by age
	ages := config.AgeOrder()
	buildingDefs := config.BaseBuildings()
	byAge := make(map[string][]config.BuildingDef)
	for _, b := range buildingDefs {
		byAge[b.RequiredAge] = append(byAge[b.RequiredAge], b)
	}

	ageNames := config.AgeByKey()

	for _, ageKey := range ages {
		buildings, ok := byAge[ageKey]
		if !ok {
			continue
		}
		ageDef := ageNames[ageKey]
		sb.WriteString(fmt.Sprintf("[gold]── %s ──[-]\n", ageDef.Name))

		for _, b := range buildings {
			bs, exists := state.Buildings[b.Key]
			sb.WriteString(fmt.Sprintf("\n [cyan]%s[-] [gray](%s)[-]", b.Name, b.Key))
			if b.MaxCount > 0 {
				sb.WriteString(fmt.Sprintf(" [yellow]Max: %d[-]", b.MaxCount))
			}
			if b.Category == "wonder" {
				sb.WriteString(fmt.Sprintf(" [yellow]Wonder — %d ticks to build[-]", b.BuildTicks))
			}
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("   %s\n", b.Description))
			sb.WriteString(fmt.Sprintf("   Category: [yellow]%s[-]  Scale: %.0f%%\n",
				b.Category, (b.CostScale-1)*100))

			// Base cost
			costKeys := make([]string, 0, len(b.BaseCost))
			for k := range b.BaseCost {
				costKeys = append(costKeys, k)
			}
			sort.Strings(costKeys)
			costParts := make([]string, 0)
			for _, k := range costKeys {
				costParts = append(costParts, fmt.Sprintf("%s:%.0f", k, b.BaseCost[k]))
			}
			sb.WriteString(fmt.Sprintf("   Base cost: %s\n", strings.Join(costParts, " ")))

			// Effects
			for _, eff := range b.Effects {
				sb.WriteString(fmt.Sprintf("   Effect: [yellow]%s %s +%.1f[-]\n",
					eff.Type, eff.Target, eff.Value))
			}

			// Live data
			if exists && bs.Unlocked {
				sb.WriteString(fmt.Sprintf("   [orange]Built: %d  Next cost: %s[-]\n",
					bs.Count, FormatCost(bs.NextCost)))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func wikiVillagers(state game.GameState) string {
	var sb strings.Builder
	sb.WriteString("[gold]Villagers[-]\n\n")
	sb.WriteString("Villagers are your workforce. They consume food each tick\n")
	sb.WriteString("and can be assigned to gather resources.\n\n")

	v := state.Villagers
	sb.WriteString(fmt.Sprintf("[orange]Population: %d / %d  |  Idle: %d  |  Food drain: %.1f/tick[-]\n\n",
		v.TotalPop, v.MaxPop, v.TotalIdle, v.FoodDrain))

	types := game.DefaultVillagerTypes()
	for _, def := range types {
		sb.WriteString(fmt.Sprintf("[cyan]%s[-] [gray](%s)[-]\n", def.Name, def.Key))
		sb.WriteString(fmt.Sprintf("  Food cost: [yellow]%.2f/tick[-]\n", def.FoodCost))
		sb.WriteString(fmt.Sprintf("  Gather rate: [yellow]%.1f/tick[-] per assigned villager\n", def.GatherRate))
		if len(def.CanGather) > 0 {
			sb.WriteString(fmt.Sprintf("  Can gather: [yellow]%s[-]\n", strings.Join(def.CanGather, ", ")))
		} else {
			sb.WriteString("  [gray]Cannot gather resources (military unit)[-]\n")
		}

		// Live data
		if vt, ok := v.Types[def.Key]; ok && vt.Unlocked {
			sb.WriteString(fmt.Sprintf("  [orange]Count: %d  Idle: %d[-]\n", vt.Count, vt.IdleCount))
			aKeys := make([]string, 0, len(vt.Assignments))
			for k := range vt.Assignments {
				aKeys = append(aKeys, k)
			}
			sort.Strings(aKeys)
			for _, res := range aKeys {
				if vt.Assignments[res] > 0 {
					sb.WriteString(fmt.Sprintf("  [orange]  → %s: %d (producing %.1f/tick)[-]\n",
						res, vt.Assignments[res], float64(vt.Assignments[res])*def.GatherRate))
				}
			}
		} else {
			sb.WriteString("  [gray]Not yet unlocked[-]\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("[gold]Food Economy[-]\n")
	sb.WriteString("  Each worker eats 0.10 food/tick but gathers 0.35/tick.\n")
	sb.WriteString("  So 1 worker on food produces net +0.25 for others.\n")
	sb.WriteString("  That covers 1 shaman (0.20) or 2.5 other workers (0.10 each).\n\n")
	sb.WriteString("  Workers on food: covers this many others:\n")
	sb.WriteString("    1 food worker  → 1 shaman + 1 worker\n")
	sb.WriteString("    2 food workers → 2 shamans + 2 workers\n")
	sb.WriteString("    4 food workers → ~5 shamans or ~10 workers\n")

	return sb.String()
}

func wikiAges(state game.GameState) string {
	var sb strings.Builder
	sb.WriteString("[gold]Ages[-]\n\n")
	sb.WriteString("Advancing through ages unlocks new buildings, resources,\n")
	sb.WriteString("and villager types. Requirements include both resources\n")
	sb.WriteString("(which are consumed) and buildings (which must exist).\n\n")

	sb.WriteString(fmt.Sprintf("[orange]Current Age: %s[-]\n\n", state.AgeName))

	ages := config.Ages()
	for _, age := range ages {
		// Highlight current age
		marker := "  "
		if age.Key == state.Age {
			marker = "[orange]▸[-]"
		}
		// Check if reached
		reached := false
		for _, a := range state.Stats.AgesReached {
			if a == age.Key {
				reached = true
				break
			}
		}
		icon := "[gray]○[-]"
		if reached {
			icon = "[green]●[-]"
		}

		sb.WriteString(fmt.Sprintf("%s %s [cyan]%s[-]\n", marker, icon, age.Name))
		sb.WriteString(fmt.Sprintf("    %s\n", age.Description))

		// Requirements
		if len(age.ResourceReqs) > 0 || len(age.BuildingReqs) > 0 {
			sb.WriteString("    [gold]Requirements:[-]\n")
			rKeys := make([]string, 0, len(age.ResourceReqs))
			for k := range age.ResourceReqs {
				rKeys = append(rKeys, k)
			}
			sort.Strings(rKeys)
			for _, k := range rKeys {
				req := age.ResourceReqs[k]
				current := 0.0
				if rs, ok := state.Resources[k]; ok {
					current = rs.Amount
				}
				color := "red"
				if current >= req {
					color = "green"
				}
				sb.WriteString(fmt.Sprintf("      [%s]%s: %.0f / %.0f[-]\n", color, k, current, req))
			}
			bKeys := make([]string, 0, len(age.BuildingReqs))
			for k := range age.BuildingReqs {
				bKeys = append(bKeys, k)
			}
			sort.Strings(bKeys)
			for _, k := range bKeys {
				req := age.BuildingReqs[k]
				current := 0
				if bs, ok := state.Buildings[k]; ok {
					current = bs.Count
				}
				color := "red"
				if current >= req {
					color = "green"
				}
				sb.WriteString(fmt.Sprintf("      [%s]%s: %d / %d[-]\n", color, k, current, req))
			}
		}

		// Unlocks
		if len(age.UnlockBuildings) > 0 {
			sb.WriteString(fmt.Sprintf("    Unlocks buildings: [yellow]%s[-]\n",
				strings.Join(age.UnlockBuildings, ", ")))
		}
		if len(age.UnlockResources) > 0 {
			sb.WriteString(fmt.Sprintf("    Unlocks resources: [yellow]%s[-]\n",
				strings.Join(age.UnlockResources, ", ")))
		}
		if len(age.UnlockVillagers) > 0 {
			sb.WriteString(fmt.Sprintf("    Unlocks villagers: [yellow]%s[-]\n",
				strings.Join(age.UnlockVillagers, ", ")))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func wikiResearch(state game.GameState) string {
	var sb strings.Builder
	sb.WriteString("[gold]Research & Technology[-]\n\n")
	sb.WriteString("The tech tree allows you to research technologies that\n")
	sb.WriteString("provide permanent bonuses to your civilization.\n\n")

	sb.WriteString("[gold]How It Works[-]\n")
	sb.WriteString("  • Spend [yellow]knowledge[-] to start researching a tech\n")
	sb.WriteString("  • Research takes several ticks to complete\n")
	sb.WriteString("  • Only one tech can be researched at a time\n")
	sb.WriteString("  • Many techs have [yellow]prerequisites[-] that must be\n")
	sb.WriteString("    researched first\n")
	sb.WriteString("  • Each tech is locked to a minimum [yellow]age[-]\n\n")

	// Current research status
	if state.Research.CurrentTech != "" {
		fmt.Fprintf(&sb, "[orange]Currently Researching: %s (%d/%d ticks)[-]\n\n",
			state.Research.CurrentTechName,
			state.Research.TotalTicks-state.Research.TicksLeft,
			state.Research.TotalTicks)
	}
	fmt.Fprintf(&sb, "[orange]Total Researched: %d techs[-]\n\n", state.Research.TotalResearched)

	// List all techs by age
	techsByAge := config.TechsByAge()
	ageOrder := config.AgeOrder()
	ages := config.AgeByKey()

	for _, ageKey := range ageOrder {
		ageTechs, ok := techsByAge[ageKey]
		if !ok {
			continue
		}
		ageDef := ages[ageKey]

		sort.Slice(ageTechs, func(i, j int) bool {
			return ageTechs[i].Name < ageTechs[j].Name
		})

		sb.WriteString(fmt.Sprintf("[gold]── %s ──[-]\n", ageDef.Name))
		for _, tech := range ageTechs {
			ts, ok := state.Research.Techs[tech.Key]
			icon := "[gray]•[-]"
			if ok && ts.Researched {
				icon = "[green]✓[-]"
			} else if ok && ts.Available {
				icon = "[cyan]○[-]"
			}

			fmt.Fprintf(&sb, " %s [cyan]%s[-] [gray](%s)[-] — %.0f knowledge\n",
				icon, tech.Name, tech.Key, tech.Cost)
			fmt.Fprintf(&sb, "   %s\n", tech.Description)
			if len(tech.Prerequisites) > 0 {
				allTechs := config.TechByKey()
				var prereqNames []string
				for _, p := range tech.Prerequisites {
					if pd, ok := allTechs[p]; ok {
						prereqNames = append(prereqNames, pd.Name)
					}
				}
				fmt.Fprintf(&sb, "   [gray]Requires: %s[-]\n", strings.Join(prereqNames, ", "))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("[gold]Commands[-]\n")
	sb.WriteString("  [cyan]research[-] <tech_key>  — Start researching\n")
	sb.WriteString("  [cyan]research cancel[-]      — Cancel current research\n")
	sb.WriteString("  [cyan]research list[-]        — Show available techs\n")

	return sb.String()
}

func wikiMilitary(state game.GameState) string {
	var sb strings.Builder
	sb.WriteString("[gold]Military System[-]\n\n")
	sb.WriteString("Recruit soldiers and send them on expeditions to earn\n")
	sb.WriteString("loot and resources. Military becomes available in the\n")
	sb.WriteString("Bronze Age.\n\n")

	sb.WriteString("[gold]How It Works[-]\n")
	sb.WriteString("  • Build [yellow]Barracks[-] to unlock soldiers\n")
	sb.WriteString("  • [cyan]recruit soldier[-] to add military units\n")
	sb.WriteString("  • Soldiers eat food but don't gather resources\n")
	sb.WriteString("  • Send soldiers on [yellow]expeditions[-] for loot\n")
	sb.WriteString("  • Military bonuses from research improve success\n\n")

	mil := state.Military
	fmt.Fprintf(&sb, "[orange]Soldiers: %d  |  Defense: %.1f[-]\n", mil.SoldierCount, mil.DefenseRating)
	if mil.MilitaryBonus > 0 {
		fmt.Fprintf(&sb, "[orange]Military Bonus: +%.0f%%[-]\n", mil.MilitaryBonus*100)
	}
	if mil.ActiveExpedition != nil {
		fmt.Fprintf(&sb, "[orange]Active Expedition: %s (%d ticks left)[-]\n",
			mil.ActiveExpedition.Name, mil.ActiveExpedition.TicksLeft)
	}
	fmt.Fprintf(&sb, "[orange]Completed Expeditions: %d[-]\n\n", mil.CompletedCount)

	sb.WriteString("[gold]Available Expeditions[-]\n\n")
	for _, exp := range mil.Expeditions {
		diffColor := "green"
		if exp.Difficulty > 0.5 {
			diffColor = "red"
		} else if exp.Difficulty > 0.3 {
			diffColor = "yellow"
		}

		fmt.Fprintf(&sb, " [cyan]%s[-] [gray](%s)[-]\n", exp.Name, exp.Key)
		fmt.Fprintf(&sb, "   %s\n", exp.Description)
		fmt.Fprintf(&sb, "   Soldiers: %d  Duration: %d ticks  Difficulty: [%s]%.0f%%[-]\n\n",
			exp.SoldiersNeeded, exp.Duration, diffColor, exp.Difficulty*100)
	}

	if len(mil.Expeditions) == 0 {
		sb.WriteString("  [gray]Reach Bronze Age to unlock expeditions[-]\n\n")
	}

	sb.WriteString("[gold]Commands[-]\n")
	sb.WriteString("  [cyan]expedition[-] <key>   — Launch an expedition\n")
	sb.WriteString("  [cyan]expedition list[-]    — Show available expeditions\n")

	return sb.String()
}

func wikiEvents(state game.GameState) string {
	var sb strings.Builder
	sb.WriteString("[gold]Random Events & Milestones[-]\n\n")

	sb.WriteString("[gold]Random Events[-]\n")
	sb.WriteString("Events trigger randomly during gameplay. Some are\n")
	sb.WriteString("beneficial (bonus resources), some are harmful (lost\n")
	sb.WriteString("production), and some are mixed.\n\n")

	sb.WriteString("  • Events respect [yellow]age requirements[-]\n")
	sb.WriteString("  • Each event has a [yellow]cooldown[-] between occurrences\n")
	sb.WriteString("  • Timed events show in the [yellow]Active Events[-] panel\n")
	sb.WriteString("  • All events are logged in the game log\n\n")

	// Show active events
	if len(state.ActiveEvents) > 0 {
		sb.WriteString("[orange]Active Events:[-]\n")
		for _, evt := range state.ActiveEvents {
			fmt.Fprintf(&sb, "  [yellow]⚡ %s[-] (%d ticks left)\n", evt.Name, evt.TicksLeft)
		}
		sb.WriteString("\n")
	}

	// List event types
	events := config.RandomEvents()
	sb.WriteString("[gold]Event Types[-]\n\n")
	for _, evt := range events {
		color := "cyan"
		isNegative := false
		for _, eff := range evt.Effects {
			if eff.Type == "steal_resource" || eff.Value < 0 {
				isNegative = true
			}
		}
		if isNegative {
			color = "red"
		}
		fmt.Fprintf(&sb, " [%s]%s[-] [gray](%s+)[-]\n", color, evt.Name, evt.MinAge)
		fmt.Fprintf(&sb, "   %s\n\n", evt.Description)
	}

	// Milestones
	sb.WriteString("\n[gold]Milestones[-]\n\n")
	sb.WriteString("Milestones are permanent achievements that reward\n")
	sb.WriteString("bonuses when conditions are met.\n\n")

	ms := state.Milestones
	fmt.Fprintf(&sb, "[orange]Progress: %d / %d[-]\n\n", ms.CompletedCount, ms.TotalCount)

	mKeys := make([]string, 0, len(ms.Milestones))
	for k := range ms.Milestones {
		mKeys = append(mKeys, k)
	}
	sort.Strings(mKeys)
	for _, key := range mKeys {
		m := ms.Milestones[key]
		if m.Completed {
			fmt.Fprintf(&sb, " [green]✓ %s[-]\n", m.Name)
		} else {
			fmt.Fprintf(&sb, " [gray]○ %s[-]\n", m.Name)
		}
		fmt.Fprintf(&sb, "   %s\n\n", m.Description)
	}

	return sb.String()
}

func wikiPrestige(state game.GameState) string {
	var sb strings.Builder
	sb.WriteString("[gold]Prestige System[-]\n\n")
	sb.WriteString("Prestige is the endgame loop. Once you reach the\n")
	sb.WriteString("[yellow]Medieval Age[-] or later, you can prestige to reset\n")
	sb.WriteString("your game and earn permanent bonuses.\n\n")

	sb.WriteString("[gold]How It Works[-]\n")
	sb.WriteString("  1. Play until you reach [yellow]Medieval Age[-] or beyond\n")
	sb.WriteString("  2. Type [cyan]prestige confirm[-] to reset\n")
	sb.WriteString("  3. Earn [yellow]prestige points[-] based on progress\n")
	sb.WriteString("  4. Spend points in the [yellow]prestige shop[-]\n")
	sb.WriteString("  5. Start over with permanent bonuses!\n\n")

	sb.WriteString("[gold]What Gets Reset[-]\n")
	sb.WriteString("  [red]Wiped:[-] Resources, buildings, villagers, research,\n")
	sb.WriteString("  military, events, milestones, build queue, age\n")
	sb.WriteString("  [green]Kept:[-] Prestige level, points, purchased upgrades\n\n")

	sb.WriteString("[gold]Point Calculation[-]\n")
	sb.WriteString("  • Base: 1 point per age beyond Primitive\n")
	sb.WriteString("    (Medieval = 5, Modern = 12)\n")
	sb.WriteString("  • Bonus: +1 per 10 milestones completed\n")
	sb.WriteString("  • Bonus: +1 per 15 techs researched\n")
	sb.WriteString("  • Bonus: +1 per 50 buildings built\n")
	sb.WriteString("  • Diminishing returns at higher prestige levels\n\n")

	sb.WriteString("[gold]Passive Bonuses[-]\n")
	sb.WriteString("  Each prestige level gives:\n")
	sb.WriteString("  • [green]+2% production[-] to all resources\n")
	sb.WriteString("  • [green]+1% tick speed[-] (game runs faster!)\n")
	sb.WriteString("  These stack and apply automatically.\n\n")

	// Live status
	p := state.Prestige
	sb.WriteString("[gold]Your Prestige Status[-]\n")
	fmt.Fprintf(&sb, "  Level: [cyan]%d[-]\n", p.Level)
	fmt.Fprintf(&sb, "  Points: [cyan]%d[-] available / [cyan]%d[-] total earned\n", p.Available, p.TotalEarned)
	if p.PassiveBonus > 0 {
		fmt.Fprintf(&sb, "  Passive: [green]+%.0f%%[-] production\n", p.PassiveBonus*100)
	}
	if p.CanPrestige {
		fmt.Fprintf(&sb, "  [green]You can prestige now for %d points![-]\n", p.PendingPoints)
	} else {
		fmt.Fprintf(&sb, "  [yellow]Reach Medieval Age to prestige[-]\n")
	}
	sb.WriteString("\n")

	// Shop listing
	sb.WriteString("[gold]Prestige Shop[-]\n\n")
	for _, key := range []string{
		"gather_boost", "storage_bonus", "research_speed", "military_power",
		"starting_food", "starting_wood", "population_cap", "expedition_loot",
		"tick_speed",
	} {
		u, ok := p.Upgrades[key]
		if !ok {
			continue
		}
		icon := "[gray]○[-]"
		if u.Tier >= u.MaxTier {
			icon = "[green]★[-]"
		} else if u.Tier > 0 {
			icon = "[cyan]◆[-]"
		}
		costStr := "[gray]MAXED[-]"
		if u.NextCost > 0 {
			costStr = fmt.Sprintf("%d pts", u.NextCost)
		}
		fmt.Fprintf(&sb, " %s [cyan]%s[-] (%d/%d)\n", icon, u.Name, u.Tier, u.MaxTier)
		fmt.Fprintf(&sb, "   %s\n", u.Description)
		if u.Tier > 0 {
			fmt.Fprintf(&sb, "   Current: [green]%s[-]\n", u.Effect)
		}
		fmt.Fprintf(&sb, "   Next tier: %s\n\n", costStr)
	}

	sb.WriteString("[gold]Commands[-]\n")
	sb.WriteString("  [cyan]prestige[-]              — View prestige status\n")
	sb.WriteString("  [cyan]prestige confirm[-]      — Reset with prestige bonus\n")
	sb.WriteString("  [cyan]prestige shop[-]         — View upgrade shop\n")
	sb.WriteString("  [cyan]prestige buy[-] <key>    — Purchase an upgrade tier\n")

	return sb.String()
}

func wikiCommands(_ game.GameState) string {
	return `[gold]Commands[-]

All commands can be typed in the input bar at the bottom.
Most have single-letter shortcuts.

[gold]── Resource Gathering ──[-]

  [cyan]gather[-] <food|wood|stone> [amount]
  Shortcut: [cyan]g[-]
  Hand-gather food, wood, or stone. Default 3, max 5.
  Example: [yellow]gather food[-], [yellow]g wood 5[-], [yellow]g stone[-]

[gold]── Building ──[-]

  [cyan]build[-] <building_key>
  Shortcut: [cyan]b[-]
  Construct a building. Costs scale with each built.
  Type [yellow]build[-] with no args to see available buildings.
  Example: [yellow]build hut[-] or [yellow]b gathering_camp[-]

  Buildings with build time (wonders) are queued and
  complete after the required number of ticks.

[gold]── Villagers ──[-]

  [cyan]recruit[-] <type> [count]
  Shortcut: [cyan]r[-]
  Recruit villagers. Requires available population cap.
  Example: [yellow]recruit worker 3[-] or [yellow]r scholar[-]

  [cyan]assign[-] <type> <resource> [count]
  Shortcut: [cyan]a[-]
  Assign idle villagers to gather a resource.
  Example: [yellow]assign worker food 2[-] or [yellow]a worker wood[-]

  [cyan]unassign[-] <type> <resource> [count]
  Shortcut: [cyan]u[-]
  Remove villagers from a gathering assignment.
  Example: [yellow]unassign worker stone 1[-]

[gold]── Research ──[-]

  [cyan]research[-] <tech_key>
  Shortcut: [cyan]res[-]
  Start researching a technology. Costs knowledge.
  Example: [yellow]research tool_making[-] or [yellow]res fire_mastery[-]

  [cyan]research cancel[-]
  Cancel current research (no refund).

  [cyan]research list[-]
  Show available technologies.

[gold]── Military ──[-]

  [cyan]expedition[-] <key>
  Shortcut: [cyan]exp[-]
  Launch a military expedition. Requires soldiers.
  Example: [yellow]expedition scout_ruins[-] or [yellow]exp raid_bandits[-]

  [cyan]expedition list[-]
  Show available expeditions.

[gold]── Prestige ──[-]

  [cyan]prestige[-]
  View your prestige level, points, and bonuses.

  [cyan]prestige confirm[-]
  Reset the game and earn prestige points.
  Requires Medieval Age or later.

  [cyan]prestige shop[-]
  View available prestige upgrades and costs.

  [cyan]prestige buy[-] <upgrade_key>
  Purchase the next tier of a prestige upgrade.
  Example: [yellow]prestige buy gather_boost[-]

[gold]── Information ──[-]

  [cyan]status[-]        Detailed civilization overview
  Shortcut: [cyan]s[-]

  [cyan]help[-]          Command quick reference
  Shortcut: [cyan]h[-] or [cyan]?[-]

[gold]── Save/Load ──[-]

  [cyan]save[-] [name]    Save game (default: autosave)
  [cyan]load[-] [name]    Load game (default: autosave)

  Game auto-saves when you press ESC to return to menu.
  Saves are stored in data/saves/ as JSON files.

[gold]── Debug ──[-]

  [cyan]dump[-]
  Export all logs and engine state to a file for debugging.
  Creates a timestamped file in data/logs/.
  Example: [yellow]dump[-]

[gold]── Other ──[-]

  [cyan]quit[-]          Save and exit the game

[gold]── Navigation ──[-]

  F1-F5 / Tab    Switch between dashboard tabs
  Shift+Tab      Previous tab
  ↑↓ / 1-9       Navigate wiki pages (in Wiki tab)
  PgUp/PgDn      Scroll wiki content
  ESC            Auto-save and return to main menu`
}

func wikiStrategy(_ game.GameState) string {
	return `[gold]Tips & Strategy[-]

[gold]── Early Game (Primitive Age) ──[-]

  • Gather wood first — you need 5 for your first stash
  • Build stashes early! Base storage is 50 but you need
    1500 food for the Stone Age
  • Build altars to start generating knowledge — you
    need 200 knowledge and 5 altars for Stone Age
  • Recruit shamans and assign them to knowledge
  • Build huts, recruit workers, keep 1/3 on food
  • Don't recruit faster than you can feed

[gold]── Stone Age ──[-]

  • Build Storage Pits early — caps fill fast
  • Stone Pits are slow (0.1/tick) so build several
  • Firepits generate knowledge for Bronze Age
  • You need 1250 food, 750 stone, 250 knowledge for Bronze
  • That means LOTS of storage buildings first

[gold]── Bronze Age ──[-]

  • This is the big unlock — farms, mines, markets, etc.
  • Warehouses (+150 storage) are critical for scaling
  • Scholars are now available — assign to knowledge
  • Iron and gold open up new building options
  • Start saving for Iron Age requirements early

[gold]── Research ──[-]

  • Start researching [yellow]Tool Making[-] as soon as you have
    scholars generating knowledge
  • Research bonuses stack — prioritize production multipliers
  • [yellow]Agriculture[-] is huge — +0.5 food/tick permanently
  • Keep a steady knowledge income for continuous research

[gold]── Military ──[-]

  • Soldiers eat food but don't gather — balance carefully
  • Start with easier expeditions (Scout Ruins) to build loot
  • Military bonuses from research improve expedition success
  • Failed expeditions still give partial loot

[gold]── Prestige ──[-]

  • Don't prestige too early — push past Medieval for more points
  • [yellow]Starting Food/Wood[-] upgrades help early game the most
  • [yellow]Gather Boost[-] and [yellow]Research Speed[-] compound over time
  • The passive +2% production and +1% tick speed per level adds up
  • [yellow]Temporal Mastery[-] upgrade gives +5% tick speed per tier
  • Each prestige is faster than the last thanks to bonuses

[gold]── General Tips ──[-]

  • [yellow]Storage is the real gate[-] — you can't hold age
    requirements without building storage infrastructure
  • [yellow]Building costs scale[-] — your 10th hut costs much
    more than your 1st. Plan purchases carefully.
  • [yellow]Wonders are worth it[-] — they take many ticks but
    provide powerful bonuses
  • [yellow]Diversify production[-] — don't put all workers on
    one resource
  • [yellow]Watch for events[-] — random events can help or
    hurt. Active events show in the Stats tab.
  • [yellow]Chase milestones[-] — they give permanent bonuses
    that compound over time
  • [yellow]Check the Ages wiki[-] — requirements turn green
    as you meet them, plan ahead
  • [yellow]The game is idle[-] — leave it running and check
    back. Resources accumulate over time.
  • [yellow]Save often[-] — type 'save' before closing

[gold]── Production Math ──[-]

  Worker gather rate:  0.35 / tick (every 2 sec)
  Worker food cost:    0.10 / tick  (net food: +0.25)
  Shaman gather rate:  0.20 / tick (knowledge only)
  Shaman food cost:    0.20 / tick
  Scholar gather rate:  0.25 / tick
  Scholar food cost:   0.20 / tick
  Merchant gather rate: 0.30 / tick (gold, crypto)
  Merchant food cost:  0.20 / tick
  Engineer gather rate: 0.35 / tick (oil, electricity, data)
  Engineer food cost:  0.25 / tick
  Hacker gather rate:  0.40 / tick (data, crypto)
  Hacker food cost:    0.30 / tick
  Astronaut gather rate: 0.50 / tick (titanium, dark matter, plasma)
  Astronaut food cost: 0.40 / tick

  Per hour (1800 ticks):
    1 worker gathering:  630 resources
    1 worker food cost:  180 food
    1 building at 0.1/tick: 180 resources`
}
