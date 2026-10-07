package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
)

// The Factions panel.
//
// This started life as a pure diplomacy screen (opinion bars and status labels)
// and grew into the single surface for everything the other civilizations are
// doing to you. Three things live here that used to live nowhere:
//
//   - BOONS AND SETBACKS: the timed buffs and penalties an encounter
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
	defs := state.Ruleset().Factions()
	usable := panelUsableWidth(w)
	tally := tallyFactionEffects(state)

	met, pending := 0, 0
	for _, def := range defs {
		if f, ok := factions[def.Key]; ok && f.Discovered {
			met++
			continue
		}
		pending++
	}

	// ─── Header ───
	writeHeadedLine(&sb, usable, "gold", "═══ Factions ═══",
		fmt.Sprintf("%d met · %d undiscovered", met, pending))
	sb.WriteString("\n")

	writeLiveFactionEffects(&sb, state, defs, usable, tally)
	writeGeographicSociety(&sb, state)

	// ─── Known Factions ───
	sb.WriteString("\n [yellow]── Civilizations you have met ──[-]\n\n")
	if met == 0 {
		// The old copy here pointed at the Colonial Age and an Embassy. Both were
		// wrong: the first civ is reachable in the Bronze Age, and first contact
		// has been expedition-driven since the encounter engine landed.
		sb.WriteString(" [gray]You have not met anyone yet. Your scouts make first contact[-]\n")
		sb.WriteString(" [gray]out in the field, so send scouting expeditions.[-]\n")
	}
	for _, def := range defs {
		f, ok := factions[def.Key]
		if !ok || !f.Discovered {
			continue
		}
		writeFactionCard(&sb, def, f, tally[def.Key], usable, state)
	}
	if met > 0 {
		// Embassies no longer gate first contact, but they are still how you court
		// a civ you have already met. Their ages are named only once the player
		// can see them (spoilers.go).
		sb.WriteString(theme.Paint(theme.RoleDim, fmt.Sprintf(" Tip: assign workers to an Embassy (%s) or a Grand Embassy",
			ageRef(state, buildingAge(state, "embassy", "colonial_age")))) + "\n")
		sb.WriteString(theme.Paint(theme.RoleDim, fmt.Sprintf(" (%s) to raise opinion over time.",
			ageRef(state, buildingAge(state, "grand_embassy", "industrial_age")))) + "\n")
	}

	// ─── Not Yet Met ───
	writeUndiscoveredRoster(&sb, pending)

	sb.WriteString("\n [gray]Commands: diplomacy ally/rival/embargo/gift/neutral/tribute/raid <civ>[-]\n")
	sb.WriteString(" [gray]Deals: diplomacy deals <civ> · diplomacy accept <civ> <n> · plan deal <civ> <n>[-]\n")

	return sb.String()
}

// factionEffectTally counts the live boons and setbacks attributable to one
// civilization.
type factionEffectTally struct {
	boons    int
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
			t.boons++
		} else {
			t.setbacks++
		}
		out[key] = t
	}
	return out
}

// writeLiveFactionEffects renders the Boons and setbacks section: every
// active event the encounter engine attributes to a civ, with its magnitude and
// wall-clock remainder, plus the occupancy of the two capacity pools.
//
// Workers on loan are listed here too. Diplomatic loans are not events and
// have no clock shown, but they ARE a live effect another civ is having on
// your empire, and this is where a player looks for that. Crews a boon lent
// are listed one by one with who sent them, how many and the time left.
func writeLiveFactionEffects(sb *strings.Builder, state game.GameState, defs []config.FactionDef, usable int, tally map[string]factionEffectTally) {
	boons, setbacks := 0, 0
	for _, t := range tally {
		boons += t.boons
		setbacks += t.setbacks
	}

	writeHeadedLine(sb, usable, "yellow", "── Boons and setbacks ──",
		fmt.Sprintf("boons %d/%d · setbacks %d/%d",
			boons, game.MaxConcurrentFactionBoons, setbacks, game.MaxConcurrentFactionMaluses))
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
		// string instead would happily cut a "[gray]" tag in half. A boon's own
		// name already starts with the civilization's ("Ironhold Clans:
		// Specialty Windfall"), which the line has just said.
		prefix := fmt.Sprintf("%s %s: ", icon, civ)
		evName := truncate(strings.TrimPrefix(ev.Name, civ+": "), descCol-runeLen(prefix))
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

	// Lent workers: a live effect with no expiry clock.
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

	// Boon crews: workers a boon lent for a set time (Extra Hands), one row
	// per crew with its own countdown, laid out like the boons above. They
	// take no boon slot, so they are not counted in the header.
	for _, crew := range state.Diplomacy.BoonCrews {
		civ := textfmt.Capitalize(civRef(state, crew.FactionKey))
		prefix := fmt.Sprintf("↳ %s: ", civ)
		desc := truncate("crew on loan", descCol-runeLen(prefix))
		count := textfmt.Count(crew.Count, "worker", "workers")
		remain := formatTicks(crew.TicksLeft, state)
		fmt.Fprintf(sb, " [green]↳[-] [cyan]%s[-]: %s%s[green]%s[-]%s%s[gray]%s[-]\n",
			civ, desc,
			columnGap(descCol-runeLen(prefix)-runeLen(desc)),
			count,
			columnGap(magCol-runeLen(count)),
			columnGap(timeCol-runeLen(remain)-1), remain)
		wrote = true
	}

	if !wrote {
		sb.WriteString(" [gray]No boons or setbacks in play.[-]\n")
		sb.WriteString(" [gray]Civilizations grant boons after encounters on scouting expeditions.[-]\n")
	}
}

// writeGeographicSociety renders the automation block in one of three states:
// nothing built, built-but-starved, or running.
func writeGeographicSociety(sb *strings.Builder, state game.GameState) {
	auto := state.Military.AutoExpedition

	sb.WriteString("\n [yellow]── Geographic Society ──[-]\n\n")

	if !auto.Active {
		sb.WriteString(theme.Paint(theme.RoleDim, fmt.Sprintf(" No Geographic Society. %s brings one, and it sends",
			ageRefCap(state, buildingAge(state, "geographic_society", "industrial_age")))) + "\n")
		sb.WriteString(theme.Paint(theme.RoleDim, " scouts out on standing orders.") + "\n")
		return
	}

	fill := 0
	if auto.Capacity > 0 {
		fill = auto.Assigned * 100 / auto.Capacity
	}
	fmt.Fprintf(sb, " Societies: [cyan]%d[-] · Staffed: [cyan]%d/%d[-] (%d%%) · Sends a party every %s\n",
		auto.Count, auto.Assigned, auto.Capacity, fill, formatTicks(auto.Interval, state))

	if auto.Starved {
		sb.WriteString(" [yellow]⚠ A party is due, but you cannot pay the expedition cost.[-]\n")
		sb.WriteString(" [yellow]  It goes out once you have the resources.[-]\n")
		return
	}

	line := fmt.Sprintf(" Next dispatch in %s", formatTicks(auto.TicksLeft, state))
	if auto.Interval > 0 {
		// Fill runs the other way from the countdown: full bar means due now.
		line += "   " + ProgressBar(float64(auto.Interval-auto.TicksLeft), float64(auto.Interval), 20)
	}
	sb.WriteString(line + "\n")
	sb.WriteString(" [gray]The Society is running: parties go out on their own.[-]\n")
}

// writeFactionCard renders the detail block for one discovered civilization:
// identity, backstory, war banner, opinion bar, status and trade bonus, the
// distance to the next standing tier, lent workers, any live effects it is
// currently applying, and the commands available given its status.
//
// The body is unchanged from the original diplomacy panel apart from two
// additions — the strength rating and the live-effect line — so muscle memory
// and the existing assertions both survive.
func writeFactionCard(sb *strings.Builder, def config.FactionDef, f game.FactionInfo, t factionEffectTally, usable int, state game.GameState) {
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
		sb.WriteString("   [red]⚔ At war: expect raids. Sue for peace with: diplomacy tribute " + def.Key + "[-]\n")
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
		bonus = fmt.Sprintf("[green]%s %s trades[-]", textfmt.SignedPercent(f.TradeBonus), game.ResourceName(f.Specialty))
	}
	fmt.Fprintf(sb, "   Status: [%s]%s[-]  %s  [gray](%s done)[-]\n",
		statusColor, f.Status, bonus, textfmt.Count(f.TradeCount, "trade", "trades"))

	// Threshold indicator: distance to the next status tier.
	fmt.Fprintf(sb, "   %s\n", diplomacyThreshold(f.Status, f.Opinion, def.Key))

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
	if t.boons > 0 || t.setbacks > 0 {
		var parts []string
		if t.boons > 0 {
			parts = append(parts, fmt.Sprintf("[gold]✦ %s active[-]", textfmt.Count(t.boons, "boon", "boons")))
		}
		if t.setbacks > 0 {
			parts = append(parts, fmt.Sprintf("[red]⚠ %s active[-]", textfmt.Count(t.setbacks, "setback", "setbacks")))
		}
		fmt.Fprintf(sb, "   %s\n", strings.Join(parts, "  "))
	}

	// Trade deals: what this civ offers now, or why it offers nothing.
	writeFactionDeals(sb, f, state, usable)

	// Action hint: the diplomacy commands available given current status.
	switch {
	case f.AtWar:
		fmt.Fprintf(sb, "   [gray]diplomacy tribute %s (sue for peace) · or wait them out[-]\n\n", def.Key)
	case f.Status == "allied":
		fmt.Fprintf(sb, "   [gray]diplomacy rival/embargo/neutral %s[-]\n\n", def.Key)
	default:
		fmt.Fprintf(sb, "   [gray]Send a gift: %s for +%d opinion (diplomacy gift %s) · ally/rival/embargo/neutral %s[-]\n\n",
			game.Amount(state.Diplomacy.GiftCost, "gold"), game.GiftOpinion, def.Key, def.Key)
	}
}

// writeUndiscoveredRoster says how many civilizations are still out there,
// and nothing more: no names, ages, specialties or personalities. They are
// the player's to discover (the no-spoiler rule, spoilers.go).
func writeUndiscoveredRoster(sb *strings.Builder, pending int) {
	sb.WriteString("\n [yellow]── Not yet met ──[-]\n\n")
	if pending == 0 {
		sb.WriteString(" [gray]Every civilization has been met.[-]\n")
		return
	}
	them := "them"
	if pending == 1 {
		them = "it"
	}
	sb.WriteString(theme.Paint(theme.RoleDim, fmt.Sprintf(" %s not yet discovered. Send expeditions to find %s.",
		textfmt.Count(pending, "civilization", "civilizations"), them)) + "\n")
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
// same row: "── Boons and setbacks ──        boons 2/5 · setbacks 1/3".
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

// effectMagnitude renders one active-event effect as a short magnitude
// ("+13% food production", "+8% all production", "+0.5 food/tick") in plain
// and colored form.
//
// The "<res>_rate" suffix case is the one that matters most here: it is the
// shape every civilization specialty boon and setback arrives in, and without
// it a boon renders as a name with no number attached.
func effectMagnitude(eff game.EventEffectInfo) (plain, colored string) {
	switch {
	case eff.Type == "production":
		plain = rateNumber(eff.Value) + " " + game.ResourceName(eff.Target) + "/tick"
	case eff.Type == "production_all":
		plain = textfmt.SignedPercent(eff.Value) + " all production"
	case eff.Type == "tick_speed":
		plain = textfmt.SignedPercent(eff.Value) + " game speed"
	case strings.HasSuffix(eff.Type, "_rate"):
		plain = textfmt.SignedPercent(eff.Value) + " " + rateEffectLabel(eff)
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

// diplomacyThreshold renders the distance-to-next-tier indicator for a
// civilization given its current status and opinion. Hostile statuses
// (rival/embargo) decay toward neutral, so they report "decaying" rather than
// a climb target. key is the civ key the ally command takes.
func diplomacyThreshold(status string, opinion int, key string) string {
	switch status {
	case "allied":
		return "[gray](maxed)[-]"
	case "rival", "embargo":
		return "[yellow](opinion decaying)[-]"
	}
	// neutral / friendly: climbing toward the next eligibility gate.
	switch {
	case opinion < 25:
		return fmt.Sprintf("[gray](+%d opinion to friendly)[-]", 25-opinion)
	case opinion < game.AllyOpinion:
		return fmt.Sprintf("[gray](+%d opinion to ally-eligible)[-]", game.AllyOpinion-opinion)
	default:
		return fmt.Sprintf("[gray](can ally: diplomacy ally %s, %s)[-]", key, game.Amount(game.AllyCost, "gold"))
	}
}

// writeFactionDeals renders a civ's trade deals on its card: one numbered
// line per offer (the number `diplomacy accept` takes), with how much better
// than the market it pays, or why the civ offers nothing. Theme roles only.
func writeFactionDeals(sb *strings.Builder, f game.FactionInfo, state game.GameState, usable int) {
	if f.DealsBlocked != "" {
		fmt.Fprintf(sb, "   %s\n", theme.Paint(theme.RoleWarning, "Deals: none while they are "+f.DealsBlocked+"."))
		return
	}
	if len(f.Deals) == 0 {
		fmt.Fprintf(sb, "   %s\n", theme.Paint(theme.RoleDim, "Deals: nothing to offer right now."))
		return
	}
	fmt.Fprintf(sb, "   %s %s\n", theme.Paint(theme.RoleLabel, "Deals:"),
		theme.Paint(theme.RoleDim, "new offers in "+formatTicks(f.DealRefreshIn, state)))
	for _, d := range f.Deals {
		fmt.Fprintf(sb, "    %s\n", dealLine(d, state, usable-4))
	}
}

// dealLine renders one offer from the player's side, in the words of
// game.DealTerms: "1. Buy: give 516 wood → get 618 food   +20% vs market".
// What you can't pay yet is in the Negative role; a taken offer is dim.
// Resource keys go in as display names, so "iron_ore" reads "iron ore".
func dealLine(d game.DealInfo, state game.GameState, width int) string {
	// Raw keys go in; game.DealGets and DealTerms name the resources, so the
	// painted skeleton below and plain use the same words ("iron ore").
	give := FormatNumber(d.GiveAmt) + " " + game.ResourceName(d.Give)
	get := game.DealGets(d.Get, d.GetAmt, d.Standing, FormatNumber)
	note := ""
	switch {
	case d.Taken:
		note = "taken"
	case d.Edge > 0:
		note = fmt.Sprintf("%+.0f%% vs market", d.Edge*100)
	case d.Kind == game.DealRare:
		note = "next age's goods"
	case d.Get != "":
		note = "not sold at the market"
	}
	plain := fmt.Sprintf("%d. %s", d.Num, game.DealTerms(d.Kind, d.Give, d.GiveAmt, d.Get, d.GetAmt, d.Standing, FormatNumber))
	if d.Taken {
		return theme.Paint(theme.RoleDim, truncate(plain+"  "+note, width))
	}
	giveRole := theme.RoleHighlight
	if state.Resources[d.Give].Amount < d.GiveAmt {
		giveRole = theme.RoleNegative
	}
	gap := columnGap(width - runeLen(plain) - runeLen(note))
	if runeLen(plain)+runeLen(note)+1 > width {
		note, gap = "", ""
	}
	// Same words as plain, painted: the numbers-and-kind skeleton of DealTerms.
	return fmt.Sprintf("%s %s %s %s %s %s%s%s",
		theme.Paint(theme.RoleLabel, fmt.Sprintf("%d.", d.Num)),
		theme.Paint(theme.RoleAccent, game.DealKindLabel(d.Kind)+":"),
		theme.Paint(theme.RoleDim, "give"), theme.Paint(giveRole, give),
		theme.Paint(theme.RoleDim, "→ get"), theme.Paint(theme.RolePositive, get),
		gap, theme.Paint(theme.RoleDim, note))
}
