package ui

import (
	"strconv"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// panelWidth is the terminal width these tests render at. Wide enough that the
// compact roster is not clipped, so assertions can look for whole tokens.
const panelWidth = 100

// lineContaining returns the first line of out that contains needle, or "" if
// no line does. Used to assert on ONE line rather than on the whole panel, which
// is how the width and one-line-per-civ claims actually get tested.
func lineContaining(out, needle string) string {
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, needle) {
			return line
		}
	}
	return ""
}

// TestFactionsProvider_RendersAllFactions verifies the panel renders every
// canonical faction (discovered as a card, undiscovered as a compact roster
// line), shows status + opinion, and does not panic on a mixed-status snapshot.
func TestFactionsProvider_RendersAllFactions(t *testing.T) {
	// Build a snapshot covering each status branch plus an undiscovered faction.
	statuses := []string{"neutral", "friendly", "allied", "rival", "embargo"}
	factions := make(map[string]game.FactionInfo)
	defs := config.BaseFactions()
	for i, def := range defs {
		if i == len(defs)-1 {
			// Leave the last faction undiscovered to exercise the roster path.
			factions[def.Key] = game.FactionInfo{Name: def.Name, Specialty: def.Specialty, Discovered: false}
			continue
		}
		factions[def.Key] = game.FactionInfo{
			Name:       def.Name,
			Specialty:  def.Specialty,
			Discovered: true,
			Opinion:    -100 + i*45, // spans negative → positive across factions
			Status:     statuses[i%len(statuses)],
			TradeBonus: def.TradeBonus,
			TradeCount: i * 3,
			Strength:   def.Strength,
		}
	}

	state := game.GameState{Diplomacy: game.DiplomacyState{Factions: factions}}

	out := factionsProvider(state, panelWidth)
	if out == "" {
		t.Fatal("factionsProvider returned empty output")
	}

	// Every faction name should appear (discovered or not).
	for _, def := range defs {
		if !strings.Contains(out, def.Name) {
			t.Errorf("overlay output missing faction %q", def.Name)
		}
	}
	// Section headers, status labels and the undiscovered teaser marker.
	for _, want := range []string{
		"Factions", "Opinion", "Status", "allied",
		"Live Favours & Setbacks", "Geographic Society", "Known Factions", "Not Yet Met", "???",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("overlay output missing expected token %q", want)
		}
	}
}

// TestFactionsProvider_EmptyState confirms the provider is safe on a zero-value
// state (nil factions map, no events, no Society) and that the onboarding hint
// points at expeditions rather than the old — and doubly wrong — claim that
// diplomacy starts with a Colonial Age Embassy.
func TestFactionsProvider_EmptyState(t *testing.T) {
	out := factionsProvider(game.GameState{}, panelWidth)
	if out == "" {
		t.Fatal("factionsProvider returned empty output on zero state")
	}
	if !strings.Contains(out, "scouting expeditions") {
		t.Errorf("empty-state overlay should point at scouting expeditions, got:\n%s", out)
	}
	if strings.Contains(out, "Colonial Age and build an Embassy") {
		t.Error("empty-state overlay still carries the stale embassy-gated-diplomacy copy")
	}
	// All eleven civs are unmet, so the header counts and the roster tail agree.
	total := len(config.BaseFactions())
	if !strings.Contains(out, "0 met") {
		t.Errorf("header should report 0 met on a zero state, got:\n%s", out)
	}
	if !strings.Contains(out, "… ") {
		t.Errorf("roster of %d unmet civs should collapse with a '… N more' tail, got:\n%s", total, out)
	}
}

// TestFactionsProvider_RendersExpandedCivData renders a snapshot exercising the
// civilization detail fields — personality, backstory, war banner, and
// lent-worker status — across the full roster.
func TestFactionsProvider_RendersExpandedCivData(t *testing.T) {
	defs := config.BaseFactions()
	factions := make(map[string]game.FactionInfo)
	for i, def := range defs {
		info := game.FactionInfo{
			Name:        def.Name,
			Specialty:   def.Specialty,
			Personality: def.Personality,
			Backstory:   def.Backstory,
			Discovered:  true,
			Opinion:     20,
			Status:      "neutral",
		}
		switch i % 3 {
		case 0:
			info.AtWar = true
			info.Opinion = -90
			info.Status = "embargo"
		case 1:
			info.LentWorkers = 5
			info.LentPerm = true
			info.Status = "allied"
			info.Opinion = 85
		}
		factions[def.Key] = info
	}

	out := factionsProvider(game.GameState{Diplomacy: game.DiplomacyState{Factions: factions}}, panelWidth)
	if out == "" {
		t.Fatal("factionsProvider returned empty output for expanded roster")
	}
	// Personality labels, a war banner, and lent-worker status should all surface.
	for _, want := range []string{"AT WAR", "on loan", "tribute"} {
		if !strings.Contains(out, want) {
			t.Errorf("expanded overlay missing expected token %q", want)
		}
	}
	// Each civ's personality string should render somewhere.
	for _, def := range defs {
		if !strings.Contains(out, def.Personality) {
			t.Errorf("overlay missing personality %q for civ %q", def.Personality, def.Name)
		}
	}
	// Lent workers are a live faction effect, so they are also summarised at the
	// top of the panel with the civ that lent them.
	if !strings.Contains(out, "on loan from") {
		t.Errorf("lent workers should also appear in the live-effects section, got:\n%s", out)
	}
}

// TestFactionsProvider_LiveFavoursAndSetbacks checks that faction-attributed
// active events are split into favours and setbacks, credited to the right civ,
// and rendered with both a magnitude and a wall-clock remainder — plus the
// capacity counters that tell a player when a pool is full.
func TestFactionsProvider_LiveFavoursAndSetbacks(t *testing.T) {
	state := game.GameState{
		TickIntervalMs: 2000, // 142 ticks → ~4m 44s
		Diplomacy: game.DiplomacyState{Factions: map[string]game.FactionInfo{
			"riverlands_tribes": {Name: "Riverlands Tribes", Discovered: true, Status: "friendly", Opinion: 40, Strength: 1},
			"ironhold_clans":    {Name: "Ironhold Clans", Discovered: true, Status: "rival", Opinion: -30, Strength: 3},
		}},
		ActiveEvents: []game.ActiveEventState{
			{
				Name: "Specialty Windfall", Key: "faction_boon_riverlands_tribes", TicksLeft: 142,
				Effects: []game.EventEffectInfo{{Type: "food_rate", Target: "food", Value: 0.13}},
			},
			{
				Name: "Industrious Spell", Key: "faction_boon_merchant_guild", TicksLeft: 48,
				Effects: []game.EventEffectInfo{{Type: "production_all", Value: 0.08}},
			},
			{
				Name: "Cursed Relic", Key: "faction_malus_ironhold_clans", TicksLeft: 20,
				Effects: []game.EventEffectInfo{{Type: "iron_rate", Target: "iron", Value: -0.11}},
			},
			// A non-faction event must not be attributed to anyone.
			{
				Name: "Reconstruction Effort", Key: "reconstruction", TicksLeft: 216,
				Effects: []game.EventEffectInfo{{Type: "production_all", Value: -0.10}},
			},
		},
	}

	out := factionsProvider(state, panelWidth)

	// Favours: named, credited, quantified, and counted down in wall-clock.
	favour := lineContaining(out, "Specialty Windfall")
	if favour == "" {
		t.Fatalf("live favour missing from panel:\n%s", out)
	}
	for _, want := range []string{"✦", "Riverlands Tribes", "+13% food", "~4m 44s"} {
		if !strings.Contains(favour, want) {
			t.Errorf("favour line %q missing %q", favour, want)
		}
	}
	if global := lineContaining(out, "Industrious Spell"); !strings.Contains(global, "+8% all prod") {
		t.Errorf("production_all favour should render its magnitude, got %q", global)
	}

	// Setbacks: the other icon, the negative magnitude, the right civ.
	setback := lineContaining(out, "Cursed Relic")
	if setback == "" {
		t.Fatalf("live setback missing from panel:\n%s", out)
	}
	for _, want := range []string{"⚠", "Ironhold Clans", "-11% iron", "[red]"} {
		if !strings.Contains(setback, want) {
			t.Errorf("setback line %q missing %q", setback, want)
		}
	}

	// Capacity counters, straight off the exported caps.
	if !strings.Contains(out, "boons 2/5") || !strings.Contains(out, "setbacks 1/3") {
		t.Errorf("panel should report pool occupancy against the exported caps, got:\n%s", out)
	}

	// The catastrophe event is not anyone's doing and must not be listed here.
	if strings.Contains(out, "Reconstruction Effort") {
		t.Error("a non-faction active event was attributed to a civilization")
	}

	// The civ cards echo their own live effects.
	if !strings.Contains(out, "✦ 1 favour active") {
		t.Errorf("the granting civ's card should flag its live favour, got:\n%s", out)
	}
	if !strings.Contains(out, "⚠ 1 setback active") {
		t.Errorf("the afflicting civ's card should flag its live setback, got:\n%s", out)
	}
}

// TestFactionsProvider_LiveEffectsEmptyState checks the gray two-liner that
// stands in when nothing is running — a statement plus how to earn one.
func TestFactionsProvider_LiveEffectsEmptyState(t *testing.T) {
	out := factionsProvider(game.GameState{}, panelWidth)
	if !strings.Contains(out, "No favours or setbacks in play") {
		t.Errorf("empty live-effects section missing its statement, got:\n%s", out)
	}
	if !strings.Contains(out, "boons 0/5") {
		t.Errorf("empty live-effects section should still show pool occupancy, got:\n%s", out)
	}
}

// TestFactionsProvider_GeographicSocietyStates walks the automation block
// through all three of its states: nothing built, built-but-starved, running.
func TestFactionsProvider_GeographicSocietyStates(t *testing.T) {
	t.Run("not built", func(t *testing.T) {
		out := factionsProvider(game.GameState{}, panelWidth)
		if !strings.Contains(out, "No Geographic Society") {
			t.Errorf("missing the not-built hint, got:\n%s", out)
		}
		if !strings.Contains(out, "Industrial Age") {
			t.Error("the not-built hint should name the age that unlocks the Society")
		}
		if strings.Contains(out, "Next dispatch") {
			t.Error("no Society built, yet the panel is counting down to a dispatch")
		}
	})

	t.Run("starved", func(t *testing.T) {
		state := game.GameState{TickIntervalMs: 2000}
		state.Military.AutoExpedition = game.AutoExpeditionState{
			Active: true, Starved: true, TicksLeft: 0, Interval: 300,
			Count: 1, Assigned: 2, Capacity: 5,
		}
		out := factionsProvider(state, panelWidth)
		if !strings.Contains(out, "cannot be outfitted") {
			t.Errorf("starved Society should warn that it can't afford a party, got:\n%s", out)
		}
		if !strings.Contains(out, "[yellow]") {
			t.Error("the starved warning should be yellow")
		}
		if strings.Contains(out, "Next dispatch") {
			t.Error("a starved Society has no meaningful countdown to show")
		}
	})

	t.Run("running", func(t *testing.T) {
		state := game.GameState{TickIntervalMs: 2000}
		state.Military.AutoExpedition = game.AutoExpeditionState{
			Active: true, TicksLeft: 77, Interval: 192,
			Count: 2, Assigned: 7, Capacity: 10,
		}
		out := factionsProvider(state, panelWidth)
		line := lineContaining(out, "Societies:")
		for _, want := range []string{"Societies: [cyan]2", "7/10", "(70%)", "~6m 24s"} {
			if !strings.Contains(line, want) {
				t.Errorf("Society status line %q missing %q", line, want)
			}
		}
		next := lineContaining(out, "Next dispatch")
		if !strings.Contains(next, "~2m 34s") {
			t.Errorf("dispatch countdown should be wall-clock, got %q", next)
		}
		// Progress bar runs the opposite way from the countdown: 77 of 192 ticks
		// left means 115/192 elapsed, so the bar is majority-filled.
		if !strings.Contains(next, "█") || !strings.Contains(next, "░") {
			t.Errorf("dispatch countdown should carry a partial progress bar, got %q", next)
		}
		if !strings.Contains(out, "Automated exploration running") {
			t.Error("running Society should say so")
		}
	})
}

// TestFactionsProvider_UndiscoveredRosterIsCompact verifies unmet civs collapse
// to one line each — the whole point of the roster, since eleven full-height
// teasers overflow an 80x24 terminal — carrying the age teaser, specialty,
// personality and strength, with a count for the tail.
func TestFactionsProvider_UndiscoveredRosterIsCompact(t *testing.T) {
	out := factionsProvider(game.GameState{}, panelWidth)

	// One line per civ, at most maxRosterRows of them, then a count.
	rows := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "??? reach ") {
			rows++
		}
	}
	if rows != maxRosterRows {
		t.Errorf("roster rendered %d rows, want the cap of %d", rows, maxRosterRows)
	}
	total := len(config.BaseFactions())
	wantTail := "… " + strconv.Itoa(total-maxRosterRows) + " more"
	if !strings.Contains(out, wantTail) {
		t.Errorf("roster missing the %q tail, got:\n%s", wantTail, out)
	}

	// The first unmet civ is the Bronze Age one, and its line carries everything.
	first := lineContaining(out, "Riverlands Tribes")
	for _, want := range []string{"★☆☆☆☆", "??? reach Bronze Age", "food", "peaceful"} {
		if !strings.Contains(first, want) {
			t.Errorf("roster line %q missing %q", first, want)
		}
	}
}

// TestFactionsProvider_RendersStrength confirms the civ power rating — on the
// snapshot since the civs were written, drawn nowhere until now — reaches the
// screen for met and unmet civs alike, and that a snapshot carrying no strength
// falls back to the definition rather than printing five hollow stars.
func TestFactionsProvider_RendersStrength(t *testing.T) {
	state := game.GameState{Diplomacy: game.DiplomacyState{Factions: map[string]game.FactionInfo{
		// Strength deliberately left at zero: an older save or a hand-built
		// snapshot must still show the civ's real rating.
		"void_reavers": {Name: "Void Reavers", Discovered: true, Status: "neutral"},
	}}}
	out := factionsProvider(state, panelWidth)

	card := lineContaining(out, "Void Reavers")
	if !strings.Contains(card, "★★★★★") {
		t.Errorf("Void Reavers are Strength 5; card line %q should show five filled stars", card)
	}
	// And an unmet mid-strength civ on the roster.
	if line := lineContaining(out, "Merchant Guild"); !strings.Contains(line, "★★☆☆☆") {
		t.Errorf("Merchant Guild are Strength 2; roster line %q should show two filled stars", line)
	}
}

// TestStrengthStars covers the clamp so an out-of-range rating still renders
// five cells instead of a negative Repeat (which panics).
func TestStrengthStars(t *testing.T) {
	for _, c := range []struct {
		in   int
		want string
	}{
		{0, "☆☆☆☆☆"},
		{1, "★☆☆☆☆"},
		{5, "★★★★★"},
		{-3, "☆☆☆☆☆"},
		{9, "★★★★★"},
	} {
		if got := strengthStars(c.in); got != c.want {
			t.Errorf("strengthStars(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestEffectMagnitude covers each renderable effect shape plus the unknown type
// that must render nothing rather than a bare name with no number.
func TestEffectMagnitude(t *testing.T) {
	cases := []struct {
		eff       game.EventEffectInfo
		want      string
		wantColor string
	}{
		{game.EventEffectInfo{Type: "food_rate", Target: "food", Value: 0.13}, "+13% food", "green"},
		{game.EventEffectInfo{Type: "iron_rate", Target: "iron", Value: -0.11}, "-11% iron", "red"},
		{game.EventEffectInfo{Type: "gather_rate", Value: 0.20}, "+20% gather", "green"},
		{game.EventEffectInfo{Type: "production_all", Value: 0.08}, "+8% all prod", "green"},
		{game.EventEffectInfo{Type: "tick_speed", Value: 0.09}, "+9% tick speed", "green"},
		{game.EventEffectInfo{Type: "production", Target: "wood", Value: 2.5}, "+2.5/t wood", "green"},
		{game.EventEffectInfo{Type: "something_else", Value: 1}, "", ""},
	}
	for _, c := range cases {
		plain, colored := effectMagnitude(c.eff)
		if plain != c.want {
			t.Errorf("effectMagnitude(%+v) plain = %q, want %q", c.eff, plain, c.want)
		}
		if c.want == "" {
			if colored != "" {
				t.Errorf("unrenderable effect should produce no coloured form, got %q", colored)
			}
			continue
		}
		if !strings.Contains(colored, "["+c.wantColor+"]") {
			t.Errorf("effectMagnitude(%+v) coloured = %q, want %s", c.eff, colored, c.wantColor)
		}
	}
}

// TestFactionsProvider_RespectsWidth checks the provider actually uses the width
// it is handed instead of discarding it — narrow terminals must not emit lines
// that soft-wrap, because tview's wrap drops the colour tags across the break.
func TestFactionsProvider_RespectsWidth(t *testing.T) {
	factions := make(map[string]game.FactionInfo)
	for _, def := range config.BaseFactions() {
		factions[def.Key] = game.FactionInfo{
			Name: def.Name, Specialty: def.Specialty, Personality: def.Personality,
			Backstory: def.Backstory, Discovered: true, Status: "neutral", Opinion: 10,
			Strength: def.Strength,
		}
	}
	state := game.GameState{Diplomacy: game.DiplomacyState{Factions: factions}}

	narrow := factionsProvider(state, 80)
	wide := factionsProvider(state, 160)
	if narrow == wide {
		t.Fatal("provider ignores its width parameter — narrow and wide renders are identical")
	}
	// The backstory is the long line; at 80 columns it must be cut.
	if !strings.Contains(narrow, "…") {
		t.Error("no truncation at 80 columns — long text will soft-wrap and lose its colour tags")
	}
}

// TestFactionsProvider_NoPanicOnDegenerateState throws the shapes that a real
// snapshot can legitimately produce and that arithmetic in this panel could trip
// over: a zero-capacity Society, a zero interval, out-of-range opinion.
func TestFactionsProvider_NoPanicOnDegenerateState(t *testing.T) {
	state := game.GameState{
		Diplomacy: game.DiplomacyState{Factions: map[string]game.FactionInfo{
			"riverlands_tribes": {Name: "Riverlands Tribes", Discovered: true, Opinion: 9999, Status: ""},
			"ironhold_clans":    {Name: "Ironhold Clans", Discovered: true, Opinion: -9999, Status: "rival"},
		}},
		ActiveEvents: []game.ActiveEventState{
			{Name: "Nameless", Key: "faction_boon_", TicksLeft: 5},
			{Name: "Unknown Civ", Key: "faction_boon_who_are_they", TicksLeft: 5},
		},
	}
	state.Military.AutoExpedition = game.AutoExpeditionState{Active: true, Capacity: 0, Interval: 0}

	if out := factionsProvider(state, 0); out == "" {
		t.Fatal("degenerate state produced no output")
	}
	// An event whose key has an empty faction suffix is not attributable and must
	// be dropped, not credited to a civ with no name.
	out := factionsProvider(state, panelWidth)
	if strings.Contains(out, "Nameless") {
		t.Error("an unattributable faction event was rendered anyway")
	}
	// An unrecognised faction key still renders, falling back to the raw key.
	if !strings.Contains(out, "who_are_they") {
		t.Errorf("unknown faction key should fall back to the key itself, got:\n%s", out)
	}
}

// TestFactionsOverlayWiring checks the four places the panel has to be named for
// a player to reach it. The strings must agree exactly: the sidebar highlight
// compares its entry against the REGISTERED overlay name, so a near-miss shows
// the entry but never lights it up.
func TestFactionsOverlayWiring(t *testing.T) {
	engine := game.NewGameEngine()

	// 1. The bare command routes to the overlay.
	if res := HandleCommand("factions", engine); res.OverlayName != "factions" {
		t.Errorf(`HandleCommand("factions") opened %q, want "factions"`, res.OverlayName)
	}
	// 2. The historical aliases open the same panel, under the PRIMARY name — the
	// sidebar can only highlight the entry it is handed.
	for _, alias := range []string{"diplomacy", "dip"} {
		if res := HandleCommand(alias, engine); res.OverlayName != "factions" {
			t.Errorf(`HandleCommand(%q) opened %q, want "factions"`, alias, res.OverlayName)
		}
	}
	// Arguments still route to the diplomacy actions rather than the panel.
	if res := HandleCommand("diplomacy gift riverlands_tribes", engine); res.OverlayName != "" {
		t.Errorf("diplomacy with arguments opened overlay %q, want an action result", res.OverlayName)
	}
	// 3. The sidebar lists it, and highlights it when it is the active overlay.
	sidebar := buildSidebarText("factions")
	if !strings.Contains(sidebar, "factions") {
		t.Errorf("sidebar is missing the factions entry:\n%s", sidebar)
	}
	if !strings.Contains(sidebar, "[black:gold] factions") {
		t.Errorf("sidebar entry does not highlight under the registered overlay name:\n%s", sidebar)
	}
	// 4. Autocomplete offers it.
	found := false
	for _, c := range commands {
		if c == "factions" {
			found = true
			break
		}
	}
	if !found {
		t.Error("autocomplete command list is missing \"factions\"")
	}
	// 5. Help advertises it.
	if help := helpProvider(game.GameState{}, panelWidth); !strings.Contains(help, "factions") {
		t.Error("help panel does not advertise the factions command")
	}
}

// TestExpeditionsProvider_ShowsSocietyStatus verifies the scouting surface says
// whether scouting is being done for you. It was silent on automation entirely,
// so a built Geographic Society was invisible on the one panel about scouting.
func TestExpeditionsProvider_ShowsSocietyStatus(t *testing.T) {
	t.Run("not built", func(t *testing.T) {
		out := expeditionsProvider(game.GameState{}, panelWidth)
		if !strings.Contains(out, "Geographic Society (Industrial Age)") {
			t.Errorf("expeditions panel should hint at the Society, got:\n%s", out)
		}
	})
	t.Run("running", func(t *testing.T) {
		state := game.GameState{TickIntervalMs: 2000}
		state.Military.AutoExpedition = game.AutoExpeditionState{
			Active: true, TicksLeft: 77, Interval: 192, Count: 1, Assigned: 3, Capacity: 5,
		}
		line := lineContaining(expeditionsProvider(state, panelWidth), "Geographic Society")
		if !strings.Contains(line, "~2m 34s") {
			t.Errorf("expeditions panel should show the next dispatch in wall-clock, got %q", line)
		}
	})
	t.Run("starved", func(t *testing.T) {
		state := game.GameState{TickIntervalMs: 2000}
		state.Military.AutoExpedition = game.AutoExpeditionState{
			Active: true, Starved: true, Interval: 192, Count: 1,
		}
		line := lineContaining(expeditionsProvider(state, panelWidth), "Geographic Society")
		if !strings.Contains(line, "too thin") {
			t.Errorf("expeditions panel should flag a starved Society, got %q", line)
		}
	})
}

// TestDiplomacyThreshold covers the distance-to-next-tier indicator branches.
func TestDiplomacyThreshold(t *testing.T) {
	cases := []struct {
		status  string
		opinion int
		want    string
	}{
		{"neutral", 0, "to friendly"},
		{"friendly", 30, "to ally-eligible"},
		{"friendly", 55, "ally-eligible"},
		{"allied", 100, "maxed"},
		{"rival", -10, "decaying"},
		{"embargo", -50, "decaying"},
	}
	for _, c := range cases {
		got := diplomacyThreshold(c.status, c.opinion)
		if !strings.Contains(got, c.want) {
			t.Errorf("diplomacyThreshold(%q, %d) = %q, want it to contain %q",
				c.status, c.opinion, got, c.want)
		}
	}
}
