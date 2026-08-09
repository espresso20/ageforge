package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// The Factions panel.
//
// This started life as a pure diplomacy screen (opinion bars and status labels)
// and grew into the single surface for everything the other civilizations are
// doing to you. Three things live here that used to live nowhere:
//
//   - LIVE FAVOURS & SETBACKS — the timed buffs and penalties an encounter
//     hands out. Before this they only appeared in the Statistics panel as
//     anonymous "active events", so a player had no way to know which civ was
//     responsible or that the effects were capped.
//   - THE GEOGRAPHIC SOCIETY — automatic expedition dispatch. A built Society
//     was previously invisible: nothing told you it existed, when it would fire
//     next, or that it was sitting starved.
//   - STRENGTH — FactionInfo.Strength has been on the snapshot since the civs
//     were written and has never once been drawn.
//
// It is registered under two overlay names ("factions" and the historical
// "diplomacy") pointing at this one provider, so both open the same view.
//
// Everything here is read-only over the snapshot and safe on a zero-value
// GameState — indexing nil maps and ranging nil slices both behave.

// factionsProvider renders the full Factions panel: live faction effects, the
// Geographic Society's automation status, a detail card per civilization you
// have met, and a compact roster of the ones you have not.
//
// w is the full terminal width; the overlay box is roughly 85% of it, and long
// text is truncated to fit because tview's soft-wrap does not carry colour tags
// across the break.
func factionsProvider(state game.GameState, w int) string {
	var sb strings.Builder

	factions := state.Diplomacy.Factions // may be nil; indexing nil maps is safe
	defs := config.BaseFactions()
	ages := config.AgeByKey()
	usable := panelUsableWidth(w)
	tally := tallyFactionEffects(state)

	met := 0
	var pending []config.FactionDef
	for _, def := range defs {
		if f, ok := factions[def.Key]; ok && f.Discovered {
			met++
			continue
		}
		pending = append(pending, def)
	}

	// ─── Header ───
	writeHeadedLine(&sb, usable, "gold", "═══ Factions ═══",
		fmt.Sprintf("%d met · %d undiscovered", met, len(pending)))
	sb.WriteString("\n")

	writeLiveFactionEffects(&sb, state, defs, usable, tally)
	writeGeographicSociety(&sb, state)

	// ─── Known Factions ───
	sb.WriteString("\n [yellow]── Known Factions ──[-]\n\n")
	if met == 0 {
		// The old copy here pointed at the Colonial Age and an Embassy. Both were
		// wrong: the first civ is reachable in the Bronze Age, and first contact
		// has been expedition-driven since the encounter engine landed.
		sb.WriteString(" [gray]You have not met anyone yet. Run scouting expeditions —[-]\n")
		sb.WriteString(" [gray]your scouts make first contact out in the field.[-]\n")
	}
	for _, def := range defs {
		f, ok := factions[def.Key]
		if !ok || !f.Discovered {
			continue
		}
		writeFactionCard(&sb, def, f, tally[def.Key], usable)
	}
	if met > 0 {
		// Embassies no longer gate first contact, but they are still how you court
		// a civ you have already met.
		sb.WriteString(" [gray]Tip: assign workers to an Embassy (Colonial) or Grand Embassy[-]\n")
		sb.WriteString(" [gray](Industrial) to passively raise opinion.[-]\n")
	}

	// ─── Not Yet Met ───
	writeUndiscoveredRoster(&sb, pending, ages, usable)

	sb.WriteString("\n [gray]Commands: diplomacy ally/rival/embargo/gift/neutral/tribute/raid <civ>[-]\n")

	return sb.String()
}

// factionEffectTally counts the live favours and setbacks attributable to one
// civilization.
type factionEffectTally struct {
	favours  int
	setbacks int
}

// tallyFactionEffects buckets the active events by the faction that granted
// them. Non-faction events (catastrophes, festivals, milestone boosts) return
// ok=false from the parser and are skipped.
func tallyFactionEffects(state game.GameState) map[string]factionEffectTally {
	out := make(map[string]factionEffectTally)
	for _, ev := range state.ActiveEvents {
		key, isBoon, ok := game.FactionKeyFromEventKey(ev.Key)
		if !ok {
			continue
		}
		t := out[key]
		if isBoon {
			t.favours++
		} else {
			t.setbacks++
		}
		out[key] = t
	}
	return out
}

// writeLiveFactionEffects renders the Live Favours & Setbacks section: every
// active event the encounter engine attributes to a civ, with its magnitude and
// wall-clock remainder, plus the occupancy of the two capacity pools.
//
// Workers on loan are listed here too. They are not events — they never expire
// on a timer the way a boon does — but they ARE a live effect another civ is
// having on your empire, and this is where a player looks for that.
func writeLiveFactionEffects(sb *strings.Builder, state game.GameState, defs []config.FactionDef, usable int, tally map[string]factionEffectTally) {
	favours, setbacks := 0, 0
	for _, t := range tally {
		favours += t.favours
		setbacks += t.setbacks
	}

	writeHeadedLine(sb, usable, "yellow", "── Live Favours & Setbacks ──",
		fmt.Sprintf("boons %d/%d · setbacks %d/%d",
			favours, game.MaxConcurrentFactionBoons, setbacks, game.MaxConcurrentFactionMaluses))
	sb.WriteString("\n")

	// Three columns: what it is, how big it is, how long it lasts. Fixed widths so
	// the magnitudes and the countdowns line up and the section can be scanned
	// vertically instead of read line by line.
	const magCol, timeCol = 16, 9
	descCol := usable - magCol - timeCol - 1
	if descCol < 24 {
		descCol = 24
	}

	names := make(map[string]config.FactionDef, len(defs))
	for _, def := range defs {
		names[def.Key] = def
	}

	wrote := false
	for _, ev := range state.ActiveEvents {
		key, isBoon, ok := game.FactionKeyFromEventKey(ev.Key)
		if !ok {
			continue
		}
		icon, iconColor := "✦", "gold"
		if !isBoon {
			icon, iconColor = "⚠", "red"
		}
		civ := key
		if def, found := names[key]; found {
			civ = def.Name
		} else if f, found := state.Diplomacy.Factions[key]; found && f.Name != "" {
			civ = f.Name
		}

		// Only the event name is trimmed to fit. Truncating the composed coloured
		// string instead would happily cut a "[gray]" tag in half.
		prefix := fmt.Sprintf("%s %s: ", icon, civ)
		evName := truncate(ev.Name, descCol-runeLen(prefix))
		magPlain, magColored := effectsSummary(ev.Effects)
		remain := formatTicks(ev.TicksLeft, state)

		fmt.Fprintf(sb, " [%s]%s[-] [cyan]%s[-]: %s%s%s%s%s[gray]%s[-]\n",
			iconColor, icon, civ, evName,
			columnGap(descCol-runeLen(prefix)-runeLen(evName)),
			magColored,
			columnGap(magCol-runeLen(magPlain)),
			columnGap(timeCol-runeLen(remain)-1), remain)
		wrote = true
	}

	// Lent workers — a live effect with no expiry clock.
	for _, def := range defs {
		f, ok := state.Diplomacy.Factions[def.Key]
		if !ok || f.LentWorkers <= 0 {
			continue
		}
		term := "temporary"
		if f.LentPerm {
			term = "permanent"
		}
		fmt.Fprintf(sb, " [green]↳ %d workers on loan from %s (%s)[-]\n", f.LentWorkers, def.Name, term)
		wrote = true
	}

	if !wrote {
		sb.WriteString(" [gray]No favours or setbacks in play.[-]\n")
		sb.WriteString(" [gray]Send scouting expeditions — encounters out in the field are[-]\n")
		sb.WriteString(" [gray]what earn a civilization's favour.[-]\n")
	}
}

// writeGeographicSociety renders the automation block in one of three states:
// nothing built, built-but-starved, or running.
func writeGeographicSociety(sb *strings.Builder, state game.GameState) {
	auto := state.Military.AutoExpedition

	sb.WriteString("\n [yellow]── Geographic Society ──[-]\n\n")

	if !auto.Active {
		sb.WriteString(" [gray]No Geographic Society. Build one (Industrial Age) to send[-]\n")
		sb.WriteString(" [gray]scouts out on standing orders.[-]\n")
		return
	}

	fill := 0
	if auto.Capacity > 0 {
		fill = auto.Assigned * 100 / auto.Capacity
	}
	fmt.Fprintf(sb, " Societies: [cyan]%d[-] · Staffed: [cyan]%d/%d[-] (%d%%) · Interval: %s\n",
		auto.Count, auto.Assigned, auto.Capacity, fill, formatTicks(auto.Interval, state))

	if auto.Starved {
		sb.WriteString(" [yellow]⚠ A dispatch is due but the party cannot be outfitted —[-]\n")
		sb.WriteString(" [yellow]  stores are short of the expedition cost.[-]\n")
		return
	}

	line := fmt.Sprintf(" Next dispatch in %s", formatTicks(auto.TicksLeft, state))
	if auto.Interval > 0 {
		// Fill runs the other way from the countdown: full bar means due now.
		line += "   " + ProgressBar(float64(auto.Interval-auto.TicksLeft), float64(auto.Interval), 20)
	}
	sb.WriteString(line + "\n")
	sb.WriteString(" [gray]Automated exploration running.[-]\n")
}

// writeFactionCard renders the detail block for one discovered civilization:
// identity, backstory, war banner, opinion bar, status and trade bonus, the
// distance to the next standing tier, lent workers, any live effects it is
// currently applying, and the commands available given its status.
//
// The body is unchanged from the original diplomacy panel apart from two
// additions — the strength rating and the live-effect line — so muscle memory
// and the existing assertions both survive.
func writeFactionCard(sb *strings.Builder, def config.FactionDef, f game.FactionInfo, t factionEffectTally, usable int) {
	// Strength lives on the snapshot, but a hand-built FactionInfo (tests, older
	// saves) can carry a zero, so fall back to the static definition.
	strength := f.Strength
	if strength <= 0 {
		strength = def.Strength
	}

	fmt.Fprintf(sb, " [cyan]%s[-]  [%s]%s[-]  [gray](%s)[-]  [gray]%s[-]\n",
		def.Name, personalityColor(def.Personality), def.Personality, def.Specialty, strengthStars(strength))

	// Backstory snippet, trimmed to the panel's usable width.
	if def.Backstory != "" {
		fmt.Fprintf(sb, "   [gray]%s[-]\n", truncate(def.Backstory, usable-4))
	}

	// War banner takes precedence — it's the headline state when active.
	if f.AtWar {
		sb.WriteString("   [red]⚔ AT WAR — raids incoming. 'diplomacy tribute " + def.Key + "' to sue for peace.[-]\n")
	}

	// Opinion bar across the -100..100 range. Clamp first to stay panic-free.
	opinion := f.Opinion
	if opinion > 100 {
		opinion = 100
	} else if opinion < -100 {
		opinion = -100
	}
	const barCells = 20
	filled := (opinion + 100) * barCells / 200
	if filled < 0 {
		filled = 0
	} else if filled > barCells {
		filled = barCells
	}
	opinionColor := "white"
	switch {
	case f.Opinion >= 50:
		opinionColor = "green"
	case f.Opinion >= 25:
		opinionColor = "cyan"
	case f.Opinion < 0:
		opinionColor = "red"
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barCells-filled)
	fmt.Fprintf(sb, "   Opinion: [%s]%4d[-]  [%s]%s[-]\n", opinionColor, f.Opinion, opinionColor, bar)

	// Status label, color-coded.
	statusColor := "white"
	switch f.Status {
	case "allied":
		statusColor = "green"
	case "friendly":
		statusColor = "cyan"
	case "rival":
		statusColor = "red"
	case "embargo":
		statusColor = "yellow"
	}
	// Active bonus + trade-rate modifier (only allied specialty trades get it).
	bonus := "[gray]no active bonus[-]"
	if f.Status == "allied" && f.TradeBonus > 0 {
		bonus = fmt.Sprintf("[green]+%.0f%% %s trades[-]", f.TradeBonus*100, f.Specialty)
	}
	fmt.Fprintf(sb, "   Status:  [%s][%s][-]  %s  [gray](%d trades)[-]\n",
		statusColor, f.Status, bonus, f.TradeCount)

	// Threshold indicator — distance to the next status tier.
	fmt.Fprintf(sb, "   %s\n", diplomacyThreshold(f.Status, f.Opinion))

	// Lent-worker status, if this civ has workers on loan with you.
	if f.LentWorkers > 0 {
		if f.LentPerm {
			fmt.Fprintf(sb, "   [green]↳ %d workers on loan (permanent)[-]\n", f.LentWorkers)
		} else {
			fmt.Fprintf(sb, "   [green]↳ %d workers on loan (temporary)[-]\n", f.LentWorkers)
		}
	}

	// Live effects this civ is currently applying, so the card and the section at
	// the top of the panel agree without the player having to cross-reference.
	if t.favours > 0 || t.setbacks > 0 {
		var parts []string
		if t.favours > 0 {
			parts = append(parts, fmt.Sprintf("[gold]✦ %d %s active[-]", t.favours, pluralize("favour", t.favours)))
		}
		if t.setbacks > 0 {
			parts = append(parts, fmt.Sprintf("[red]⚠ %d %s active[-]", t.setbacks, pluralize("setback", t.setbacks)))
		}
		fmt.Fprintf(sb, "   %s\n", strings.Join(parts, "  "))
	}

	// Action hint: the diplomacy commands available given current status.
	switch {
	case f.AtWar:
		fmt.Fprintf(sb, "   [gray]diplomacy tribute %s (sue for peace) · or wait them out[-]\n\n", def.Key)
	case f.Status == "allied":
		fmt.Fprintf(sb, "   [gray]diplomacy rival/embargo/neutral %s[-]\n\n", def.Key)
	default:
		fmt.Fprintf(sb, "   [gray]diplomacy gift %s (200g, +15) · ally/rival/embargo/neutral %s[-]\n\n",
			def.Key, def.Key)
	}
}

// maxRosterRows caps the compact undiscovered roster. Eleven civilizations at
// one line each still pushes the panel past a 24-row terminal once the sections
// above it are drawn, so the tail collapses into a count.
const maxRosterRows = 6

// writeUndiscoveredRoster renders the civilizations you have not met as one
// compact line each — name, strength, and the age that unlocks first contact —
// rather than the full-height teaser card the panel used to draw for every one
// of them.
//
// Strength sits in a fixed column BEFORE the truncatable detail so it survives
// on a narrow terminal; the specialty/personality tail is what gets clipped.
func writeUndiscoveredRoster(sb *strings.Builder, pending []config.FactionDef, ages map[string]config.AgeDef, usable int) {
	sb.WriteString("\n [yellow]── Not Yet Met ──[-]\n\n")

	if len(pending) == 0 {
		sb.WriteString(" [gray]Every civilization has been met.[-]\n")
		return
	}

	const nameCol = 18
	// 1 leading space + name column + space + 5 stars + 2 spaces.
	detailBudget := usable - (nameCol + 9)
	if detailBudget < 12 {
		detailBudget = 12
	}

	for i, def := range pending {
		if i >= maxRosterRows {
			fmt.Fprintf(sb, " [gray]… %d more[-]\n", len(pending)-maxRosterRows)
			break
		}
		ageName := def.MinAge
		if a, found := ages[def.MinAge]; found {
			ageName = a.Name
		}
		detail := truncate(fmt.Sprintf("??? reach %s · %s, %s", ageName, def.Specialty, def.Personality), detailBudget)
		fmt.Fprintf(sb, " [cyan]%s[-] [gray]%s[-]  [gray]%s[-]\n",
			padRight(truncate(def.Name, nameCol), nameCol), strengthStars(def.Strength), detail)
	}
}

// === shared formatting helpers ===

// panelUsableWidth converts the terminal width a provider is handed into the
// text width actually available inside the overlay box (roughly 85% of the
// terminal, less the border and padding). Clamped so a tiny or absent width
// still yields a sane budget rather than a negative one.
func panelUsableWidth(w int) int {
	usable := int(float64(w)*0.85) - 4
	if usable < 56 {
		usable = 56
	}
	if usable > 110 {
		usable = 110
	}
	return usable
}

// writeHeadedLine writes a section header with a right-aligned summary on the
// same row: "── Live Favours & Setbacks ──        boons 2/5 · setbacks 1/3".
// Both halves are passed uncoloured so the gap is computed from the width that
// actually prints, not from the length of the colour tags.
func writeHeadedLine(sb *strings.Builder, usable int, color, header, summary string) {
	fmt.Fprintf(sb, " [%s]%s[-]%s[gray]%s[-]\n",
		color, header, alignGap(usable, runeLen(header)+1, runeLen(summary)), summary)
}

// alignGap returns the spaces that push a trailing segment of width rightLen to
// the right edge of a total-width line whose leading segment is leftLen wide.
// Never narrower than two spaces, so the halves cannot collide.
func alignGap(total, leftLen, rightLen int) string {
	gap := total - leftLen - rightLen
	if gap < 2 {
		gap = 2
	}
	return strings.Repeat(" ", gap)
}

// columnGap pads out a fixed-width column. Always at least one space, so a value
// that overruns its column pushes the next one along instead of fusing with it.
func columnGap(n int) string {
	if n < 1 {
		return " "
	}
	return strings.Repeat(" ", n)
}

// runeLen counts display cells as runes. Every glyph the panel aligns against
// (box drawing, stars, the middot) is single-width, so runes are the right unit
// here and bytes are not.
func runeLen(s string) int { return len([]rune(s)) }

// strengthStars renders a 1-5 civ power rating as filled/hollow stars, clamped
// so an out-of-range or zero value still produces five cells.
func strengthStars(n int) string {
	if n < 0 {
		n = 0
	}
	if n > 5 {
		n = 5
	}
	return strings.Repeat("★", n) + strings.Repeat("☆", 5-n)
}

// pluralize appends an "s" to word when n is not 1.
func pluralize(word string, n int) string {
	if n == 1 {
		return word
	}
	return word + "s"
}

// effectsSummary renders every effect of an active event as one comma-joined
// magnitude string, returned twice: once plain (for width arithmetic) and once
// with sign colouring (for display). Effects with no renderable magnitude are
// skipped rather than printed as a bare name.
func effectsSummary(effects []game.EventEffectInfo) (plain, colored string) {
	var plains, coloreds []string
	for _, eff := range effects {
		p, c := effectMagnitude(eff)
		if p == "" {
			continue
		}
		plains = append(plains, p)
		coloreds = append(coloreds, c)
	}
	return strings.Join(plains, ", "), strings.Join(coloreds, ", ")
}

// effectMagnitude renders one active-event effect as a short magnitude —
// "+13% food", "+8% all prod", "+9% tick speed" — in plain and coloured form.
//
// The "<res>_rate" suffix case is the one that matters most here: it is the
// shape every faction specialty boon and setback arrives in, and without it a
// favour renders as a name with no number attached.
func effectMagnitude(eff game.EventEffectInfo) (plain, colored string) {
	switch {
	case eff.Type == "production":
		plain = fmt.Sprintf("%+.1f/t %s", eff.Value, eff.Target)
	case eff.Type == "production_all":
		plain = fmt.Sprintf("%+.0f%% all prod", eff.Value*100)
	case eff.Type == "tick_speed":
		plain = fmt.Sprintf("%+.0f%% tick speed", eff.Value*100)
	case strings.HasSuffix(eff.Type, "_rate"):
		plain = fmt.Sprintf("%+.0f%% %s", eff.Value*100, rateEffectLabel(eff))
	default:
		return "", ""
	}
	color := "green"
	if eff.Value < 0 {
		color = "red"
	}
	return plain, "[" + color + "]" + plain + "[-]"
}

// personalityColor maps a civilization personality to a tview color for the
// overlay header. Unknown personalities render gray (panic-safe default).
func personalityColor(personality string) string {
	switch personality {
	case "aggressive":
		return "red"
	case "peaceful":
		return "green"
	case "mercantile":
		return "yellow"
	case "isolationist":
		return "lightblue"
	default:
		return "gray"
	}
}

// truncate shortens s to at most max runes, appending an ellipsis when cut.
// Operates on runes so multibyte backstory text is never split mid-character.
func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}

// diplomacyThreshold renders the distance-to-next-tier indicator for a faction
// given its current status and opinion. Hostile statuses (rival/embargo) decay
// toward neutral, so they report "decaying" rather than a climb target.
func diplomacyThreshold(status string, opinion int) string {
	switch status {
	case "allied":
		return "[gray](maxed)[-]"
	case "rival", "embargo":
		return "[yellow](opinion decaying)[-]"
	}
	// neutral / friendly: climbing toward the next eligibility gate.
	switch {
	case opinion < 25:
		return fmt.Sprintf("[gray](+%d to friendly)[-]", 25-opinion)
	case opinion < 50:
		return fmt.Sprintf("[gray](+%d to ally-eligible)[-]", 50-opinion)
	default:
		return "[gray](ally-eligible — 500g)[-]"
	}
}
