package main

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// code_special.go holds what the code reader has to be told: what each
// source file is to a player (for the where sentence, the sheet and the
// age), and the handful of places where text is built by a helper, lives in
// a table of its own, or looks like prose and is not.

// noteFor returns what a source file is to a player.
func noteFor(f *srcFile) fileNote {
	n, ok := fileNotes[f.rel]
	if !ok {
		n = fileNote{place: "in " + f.rel}
	}
	if n.seen == 0 {
		n.seen = seenRegular
	}
	if n.area == nil {
		switch {
		case f.dir == "ui" || f.dir == ".":
			n.area = areaScreens
		case strings.HasPrefix(f.dir, "ui/mapstyle") || f.dir == "mapmodel" || f.dir == "theme":
			n.area = areaMap
		default:
			n.area = areaLog
		}
	}
	return n
}

// fileNotes says, for each file that holds text in code, where a player
// meets that text. A file missing here still exports; its rows say which
// file they came from.
var fileNotes = map[string]fileNote{
	// ----- the engine: log lines and refusals -----
	"game/account.go":           {place: "about accounts, recovery codes and progress exports", age: "menu"},
	"game/account_badges.go":    {place: "about the account's badges"},
	"game/account_settings.go":  {place: "about an account's settings and titles", age: "menu"},
	"game/auto_expedition.go":   {place: "about expeditions the game sends for you", age: "primitive_age", seen: seenNews},
	"game/backup.go":            {place: "about account backups", age: "menu"},
	"game/badges.go":            {place: "about badges", age: "primitive_age", seen: seenNews},
	"game/badge_hooks.go":       {place: "about badges"},
	"game/buildings.go":         {place: "about buildings", age: "primitive_age", seen: seenNews},
	"game/bus.go":               {skip: whyKey},
	"game/caps.go":              {place: "about the limits on bonuses", age: "primitive_age"},
	"game/catastrophe.go":       {place: "about catastrophes", age: "iron_age", seen: seenRare},
	"game/culture.go":           {place: "about culture and festivals", age: "bronze_age", seen: seenNews},
	"game/current_game.go":      {place: "about which save is the current game", age: "menu"},
	"game/deals.go":             {place: "about trade deals with other civilizations", age: "bronze_age", seen: seenNews},
	"game/defense.go":           {place: "about raids and your army's defense", age: "bronze_age", seen: seenNews},
	"game/devcmd.go":            {skip: whyDev},
	"game/devmode.go":           {skip: whyDev},
	"game/digest.go":            {place: "in the summary of what happened while you were away", age: "primitive_age"},
	"game/diplomacy.go":         {place: "about other civilizations", age: "bronze_age", seen: seenNews},
	"game/encounters.go":        {place: "about meeting other civilizations on expeditions", age: "bronze_age", seen: seenNews},
	"game/engine.go":            {place: "about the run itself: gathering, building, workers, research, ages, events, wonders and prestige", age: "primitive_age", seen: seenNews},
	"game/events.go":            {place: "about events", age: "primitive_age", seen: seenNews},
	"game/expedition_flavor.go": {place: "in the story lines about expeditions", age: "primitive_age"},
	"game/faction_boon.go":      {place: "about boons and setbacks from other civilizations", age: "bronze_age", seen: seenNews},
	"game/faith.go":             {place: "about faith", age: "stone_age", seen: seenNews},
	"game/fate.go":              {place: "about fated dooms", age: "iron_age", seen: seenRare},
	"game/features.go":          {place: "about commands that wait for a tech", age: "primitive_age", seen: seenNews},
	"game/harbinger.go":         {place: "about the harbinger", age: "iron_age", seen: seenRare},
	"game/history.go":           {place: "in the history of past runs", age: "medieval_age", seen: seenRare},
	"game/last_passage.go":      {place: "about the Last Passage", age: "interstellar_age", seen: seenRare},
	"game/legacy.go":            {place: "about the legacy kit", age: "medieval_age", seen: seenRare},
	"game/mastery.go":           {place: "about Era Mastery", age: "medieval_age", seen: seenRare},
	"game/milestones.go":        {place: "about milestones", age: "primitive_age", seen: seenNews},
	"game/military.go":          {place: "about expeditions, campaigns and the army", age: "primitive_age", seen: seenNews},
	"game/names.go":             {place: "where the game names a thing it has no name for", age: "primitive_age"},
	"game/overflow.go":          {place: "about wonder overflow", age: "stone_age", seen: seenNews},
	"game/plan.go":              {place: "about the build plan", age: "primitive_age", seen: seenNews},
	"game/plan_advance.go":      {place: "about the build plan's advance item", age: "primitive_age", seen: seenNews},
	"game/plan_deal.go":         {place: "about the build plan's deal items", age: "bronze_age", seen: seenNews},
	"game/plan_research.go":     {place: "about techs in the build plan", age: "primitive_age", seen: seenNews},
	"game/plan_trade.go":        {place: "about the build plan's trade items", age: "bronze_age", seen: seenNews},
	"game/prestige.go":          {place: "about prestige", age: "medieval_age", seen: seenRare},
	"game/progress.go":          {place: "about advancing to the next age", age: "primitive_age", seen: seenNews},
	"game/research.go":          {place: "about research", age: "primitive_age", seen: seenNews},
	"game/save.go":              {place: "about saving and loading", age: "menu"},
	"game/session_mark.go":      {place: "about the session"},
	"game/shares.go":            {place: "about the roster and auto-recruit", age: "primitive_age", seen: seenNews},
	"game/sim.go":               {place: "about the time the game ran while you were away", age: "primitive_age"},
	"game/spoilers.go":          {place: "that stands in for the name of something not met yet", age: "primitive_age"},
	"game/stats.go":             {place: "in the run's statistics", age: "primitive_age"},
	"game/tech_layer.go":        {place: "about what techs change", age: "primitive_age"},
	"game/trade.go":             {place: "about the market and trade routes", age: "bronze_age", seen: seenNews},
	"game/types.go":             {place: "in the game's state"},
	"game/updater.go":           {place: "about checking for and installing an update", age: "menu", seen: seenRare},
	"game/villagers.go":         {place: "about workers", age: "primitive_age", seen: seenNews},
	"game/wording.go":           {place: "in the Army's text", age: "primitive_age"},
	"boon/boon.go":              {place: "in the names of a boon's kind", age: "bronze_age"},
	"boon/roll.go":              {place: "in the line the log writes for a boon or setback", age: "bronze_age", area: areaWorld},
	"boon/catalog.go":           {place: "in the boons and setbacks other civilizations leave", age: "bronze_age", area: areaWorld},
	"flavor/resource.go":        {place: "in the way the story lines name a resource (the {res}, {res_stores} and {res_haul} slots)", area: areaFlavorExp, age: "primitive_age", seen: seenPool},
	"flavor/flavor.go":          {place: "in the flavor package's names for its own settings"},
	"pkg/textfmt/textfmt.go":    {place: "wherever the game writes a list, a count or a rate", age: "primitive_age", seen: seenAlways},
	"rules/badges.go":           {place: "in the descriptions of the tech badges", area: areaBadges, seen: seenRegular},
	"main.go":                   {place: "printed in the terminal when the game starts or exits", age: "menu", seen: seenRare},

	// ----- config: text built in code -----
	"config/effect_text.go":    {place: "in what a building's description says it does", area: areaBuildings, age: "primitive_age", seen: seenOften},
	"config/tech_effects.go":   {place: "in what a tech's card says it does", area: areaTechs, age: "primitive_age", seen: seenOften},
	"config/mechanics.go":      {place: "in what a tech's card says it does", area: areaTechs, age: "primitive_age"},
	"config/feature_locks.go":  {place: "in the refusal when a command waits for a tech", area: areaTechs, age: "primitive_age", seen: seenNews},
	"config/milestones.go":     {place: "in the Milestones panel", area: areaMilestones, age: "primitive_age"},
	"config/badge_catalog.go":  {place: "in the badge case", area: areaBadges},
	"config/badge_specials.go": {place: "in the badge case", area: areaBadges},
	"config/badges.go":         {place: "in the badge case", area: areaBadges},
	"config/epochs.go":         {place: "about eras and catastrophes", area: areaEvents, age: "iron_age"},
	"config/events.go":         {place: "about events", area: areaEvents},
	"config/tech_tree.go":      {place: "in the Research panel", area: areaTechs, age: "primitive_age"},
	"config/tech_checks.go":    {skip: whyDev},
	"config/pacing.go":         {skip: whyDev},
	"config/workers.go":        {place: "in the Workers panel", area: areaAges, age: "primitive_age"},

	// ----- the map -----
	"mapmodel/catalog.go":     {place: "on the map, naming a kind of building", age: "primitive_age"},
	"mapmodel/clock.go":       {place: "on the map, naming the weather or time of day", age: "primitive_age"},
	"mapmodel/flows.go":       {place: "in the map's list of what is stuck", age: "primitive_age"},
	"mapmodel/model.go":       {place: "on the map", age: "primitive_age"},
	"mapmodel/recap.go":       {place: "in the map's news of what changed since your last visit", age: "primitive_age"},
	"mapmodel/world.go":       {place: "in the name the map makes up for your world (it is put together from these syllables)", age: "primitive_age", seen: seenPool},
	"mapmodel/commands.go":    {place: "in the line the map shows about a harbinger", age: "iron_age"},
	"mapmodel/fingerprint.go": {skip: whyDev},
	"mapmodel/glyphs.go":      {place: "naming the map's glyph sets", age: "primitive_age"},
	"mapmodel/sky.go":         {place: "on the map in the space ages", age: "space_age"},
	"mapmodel/movers.go":      {place: "on the map", age: "primitive_age"},
	"mapmodel/city.go":        {place: "on the map of the late-game city", age: "modern_age"},
	"theme/theme.go":          {place: "in the theme picker", age: "primitive_age"},
	"theme/contrast.go":       {place: "in the theme picker's color-vision check", age: "primitive_age"},
	"theme/palette.go":        {skip: whyDev},

	"ui/mapstyle/roguelike/chrome.go":        {place: "in the frame of the roguelike map: its header, its hints and its list of what is stuck", age: "primitive_age", seen: seenOften},
	"ui/mapstyle/roguelike/city.go":          {place: "on the signs of the roguelike map's late-game city", age: "cyberpunk_age", seen: seenPool},
	"ui/mapstyle/roguelike/citylines.go":     {place: "when you inspect a rail or tether in the roguelike map's late-game city", age: "modern_age", seen: seenRare},
	"ui/mapstyle/roguelike/cityview.go":      {place: "when you inspect the roguelike map's late-game city", age: "modern_age", seen: seenRare},
	"ui/mapstyle/roguelike/glyph.go":         {place: "in the roguelike map's legend", age: "primitive_age"},
	"ui/mapstyle/roguelike/inspect.go":       {place: "when you inspect something on the roguelike map", age: "primitive_age"},
	"ui/mapstyle/roguelike/plates.go":        {place: "in the decoration of the roguelike map (its title plate changes with the era)", age: "primitive_age", seen: seenRare},
	"ui/mapstyle/roguelike/scene.go":         {place: "on the roguelike map's danger dial", age: "iron_age", seen: seenRare},
	"ui/mapstyle/roguelike/space_chrome.go":  {place: "in the frame of the roguelike map in the space ages", age: "space_age", seen: seenOften},
	"ui/mapstyle/roguelike/space_deep.go":    {place: "when you inspect something on the roguelike map in deep space", age: "interstellar_age", seen: seenRare},
	"ui/mapstyle/roguelike/space_galaxy.go":  {place: "when you inspect something on the roguelike map of the galaxy", age: "galactic_age", seen: seenRare},
	"ui/mapstyle/roguelike/space_inspect.go": {place: "when you inspect something on the roguelike map in the space ages", age: "space_age", seen: seenRare},
	"ui/mapstyle/roguelike/space_legend.go":  {place: "in the roguelike map's legend in the space ages", age: "space_age"},
	"ui/mapstyle/roguelike/space_mandala.go": {place: "when you inspect the roguelike map of the Transcendent Age", age: "transcendent_age", seen: seenRare},
	"ui/mapstyle/roguelike/space_orbit.go":   {place: "when you inspect your station on the roguelike map", age: "space_age", seen: seenRare},
	"ui/mapstyle/roguelike/space_quantum.go": {place: "when you inspect the roguelike map of the Quantum Age", age: "quantum_age", seen: seenRare},
	"ui/mapstyle/roguelike/space_visitor.go": {place: "when you inspect a visitor on the roguelike map in the space ages", age: "space_age", seen: seenRare},
	"ui/mapstyle/roguelike/traffic.go":       {place: "when you inspect a railway on the roguelike map", age: "industrial_age", seen: seenRare},
	"ui/mapstyle/roguelike/view.go":          {place: "naming the roguelike map style and its zoom levels", age: "primitive_age"},
	"ui/mapstyle/roguelike/visitor.go":       {place: "when you inspect a visitor on the roguelike map", age: "primitive_age", seen: seenRare},
	"ui/mapstyle/skyline/chrome.go":          {place: "in the frame of the skyline map: its header, its hints and its status line", age: "primitive_age", seen: seenOften},
	"ui/mapstyle/skyline/city.go":            {place: "on the signs of the skyline map's late-game city", age: "cyberpunk_age", seen: seenPool},
	"ui/mapstyle/skyline/compact.go":         {place: "on the mini map in the main window", age: "primitive_age", seen: seenAlways},
	"ui/mapstyle/skyline/inspect.go":         {place: "when you inspect something on the skyline map", age: "primitive_age"},
	"ui/mapstyle/skyline/orbit.go":           {place: "in the frame of the skyline map in the space ages", age: "space_age", seen: seenOften},
	"ui/mapstyle/skyline/orbit_compact.go":   {place: "on the mini map in the space ages", age: "space_age", seen: seenAlways},
	"ui/mapstyle/skyline/orbit_mandala.go":   {place: "on the skyline map of the Transcendent Age", age: "transcendent_age", seen: seenRare},
	"ui/mapstyle/skyline/scene.go":           {place: "on the markers of the skyline map", age: "primitive_age"},
	"ui/mapstyle/skyline/skyline.go":         {place: "naming the skyline map style", age: "primitive_age"},
	"ui/mapstyle/skyline/wonders.go":         {place: "on a wonder drawn on the skyline map", age: "primitive_age", seen: seenRare},

	// ----- screens -----
	"ui/accounts_panel.go":       {place: "in the Accounts panel of the main menu", age: "menu", seen: seenRegular},
	"ui/age_splash.go":           {place: "on the splash shown when you reach a new age", age: "stone_age", seen: seenRare, voice: true},
	"ui/ancient_memory_modal.go": {place: "in the Ancient Memory offer after a prestige", age: "medieval_age", seen: seenRare, voice: true},
	"ui/app.go":                  {place: "when the game starts", age: "menu", seen: seenAlways},
	"ui/arrival.go":              {place: "on the screen shown when you reach a new age", age: "stone_age", seen: seenRare},
	"ui/arrival_page.go":         {place: "on the screen shown when you reach a new age", age: "stone_age", seen: seenRare},
	"ui/badge_art_legend.go":     {skip: whyArt},
	"ui/badge_art_special.go":    {skip: whyArt},
	"ui/badge_case.go":           {place: "in the badge case", age: "primitive_age"},
	"ui/badge_emblems.go":        {skip: whyArt},
	"ui/badge_list.go":           {place: "in the list of badges on the Stats panel and in the badge commands", age: "primitive_age"},
	"ui/badge_panel.go":          {place: "in the badge case and the toast when a badge is earned", age: "primitive_age"},
	"ui/braille.go":              {skip: whyArt},
	"ui/catastrophe_modal.go":    {place: "in the window where you choose to endure or succumb to a catastrophe", age: "iron_age", seen: seenRare, voice: true},
	"ui/command_input.go":        {place: "at the command prompt", age: "primitive_age", seen: seenAlways},
	"ui/commands.go":             {place: "in the Help panel", age: "primitive_age", seen: seenRegular},
	"ui/dashboard.go":            {place: "in the main window", age: "primitive_age", seen: seenAlways},
	"ui/ending.go":               {place: "on the film shown when a run ends in a prestige or a fall", age: "medieval_age", seen: seenRare},
	"ui/ending_page.go":          {place: "on the film shown when a run ends in a prestige or a fall", age: "medieval_age", seen: seenRare},
	"ui/fit.go":                  {place: "in the main window", age: "primitive_age", seen: seenAlways},
	"ui/icons.go":                {place: "in the icons check (the icons command)", age: "primitive_age", seen: seenRare},
	"ui/input.go":                {place: "in the reply to a command you typed", age: "primitive_age", seen: seenNews},
	"ui/keystone.go":             {place: "where a panel says a wonder waits for its tech", age: "primitive_age"},
	"ui/last_passage.go":         {place: "in the Last Passage window", age: "interstellar_age", seen: seenRare, voice: true},
	"ui/load_game.go":            {place: "in the Load Game screen", age: "menu", seen: seenRegular},
	"ui/log_routing.go":          {place: "in the log", age: "primitive_age", seen: seenAlways},
	"ui/map_mini.go":             {place: "on the mini map in the main window", age: "primitive_age", seen: seenAlways},
	"ui/map_panel.go":            {place: "in the Map panel", age: "primitive_age"},
	"ui/map_settings.go":         {place: "in the reply to a map setting command", age: "primitive_age"},
	"ui/mastery.go":              {place: "where the panels show Era Mastery", age: "medieval_age", seen: seenRare},
	"ui/medal.go":                {skip: whyArt},
	"ui/menu.go":                 {place: "in the main menu", age: "menu", seen: seenAlways},
	"ui/menu_page.go":            {place: "in the main menu", age: "menu", seen: seenAlways},
	"ui/menu_scene.go":           {place: "in the main menu", age: "menu", seen: seenAlways},
	"ui/menu_town.go":            {place: "in the main menu", age: "menu", seen: seenAlways},
	"ui/menu_update.go":          {place: "in the main menu's Check for updates", age: "menu", seen: seenRare},
	"ui/newgame_modal.go":        {place: "in the windows that ask for a name: a new game's and a new account's", age: "menu", seen: seenAlways},
	"ui/overlay.go":              {place: "in the frame of every panel", age: "primitive_age", seen: seenAlways},
	"ui/overlay_buildings.go":    {place: "in the Buildings panel", age: "primitive_age", seen: seenOften},
	"ui/overlay_epoch.go":        {place: "in the Epoch panel", age: "primitive_age"},
	"ui/overlay_expeditions.go":  {place: "in the Expeditions panel", age: "primitive_age"},
	"ui/overlay_factions.go":     {place: "in the Factions panel", age: "bronze_age"},
	"ui/overlay_harbinger.go":    {place: "in the Harbinger panel", age: "iron_age", seen: seenRare},
	"ui/overlay_help.go":         {place: "in the Help panel", age: "primitive_age"},
	"ui/overlay_history.go":      {place: "in the History panel", age: "medieval_age", seen: seenRare},
	"ui/overlay_logs.go":         {place: "in the Logs panel", age: "primitive_age"},
	"ui/overlay_milestones.go":   {place: "in the Milestones panel", age: "primitive_age"},
	"ui/overlay_military.go":     {place: "in the Army panel", age: "primitive_age"},
	"ui/overlay_plan.go":         {place: "in the Plan panel", age: "primitive_age"},
	"ui/overlay_research.go":     {place: "where a panel says what a tech does", age: "primitive_age"},
	"ui/overlay_stats.go":        {place: "in the Stats panel", age: "primitive_age"},
	"ui/overlay_trade.go":        {place: "in the Trade panel", age: "bronze_age"},
	"ui/overlay_wonders.go":      {place: "in the Wonders panel", age: "stone_age"},
	"ui/overlay_workers.go":      {place: "in the Workers panel", age: "primitive_age"},
	"ui/prestige_kit.go":         {place: "on the prestige screens: depth points and the legacy kit", age: "medieval_age", seen: seenRare},
	"ui/research_panel.go":       {place: "in the Research panel", age: "primitive_age"},
	"ui/research_tree.go":        {place: "in the Research panel's tech tree", age: "primitive_age"},
	"ui/resources_box.go":        {place: "in the Resources box of the main window", age: "primitive_age", seen: seenAlways},
	"ui/save_modal.go":           {place: "in the window that asks whether to overwrite a save", age: "menu", seen: seenRegular},
	"ui/splash_canvas.go":        {skip: whyArt},
	"ui/spoilers.go":             {place: "that stands in for the name of something not met yet", age: "primitive_age"},
	"ui/suggest.go":              {place: "at the command prompt", age: "primitive_age", seen: seenAlways},
	"ui/tab_economy.go":          {place: "in the main window's Buildings list and its Getting Started guide", age: "primitive_age", seen: seenAlways},
	"ui/theme_account.go":        {place: "about themes", age: "primitive_age"},
	"ui/theme_picker.go":         {place: "in the theme picker", age: "primitive_age"},
	"ui/theme_unlock.go":         {place: "when a badge unlocks a theme", age: "bronze_age", seen: seenRare},
	"ui/theme_widgets.go":        {skip: "widget chrome, not sentences"},
	"ui/toast.go":                {place: "in a toast", age: "primitive_age"},
	"ui/villager_panel.go":       {place: "in the Workers panel", age: "primitive_age"},
	"ui/wonder_gallery.go":       {place: "in the Wonders panel", age: "stone_age"},
	"ui/wonder_icon.go":          {skip: whyArt},
}

// skipDecls are functions and variables whose strings no player reads, by
// "file:name".
var skipDecls = map[string]string{
	"config/tech_effects.go:Check":                whyDev,
	"ui/commands.go:devCommands":                  whyDev,
	"config/tech_tree.go:TechTreeProblems":        whyDev,
	"config/badge_catalog.go:mastery":             "a note for developers: how a badge is proved or judged",
	"rules/badges.go:BadgeProblems":               whyDev,
	"mapmodel/commands.go:BuildingCommand":        whySyntax,
	"mapmodel/commands.go:FactionCommand":         whySyntax,
	"mapmodel/commands.go:IdleCommand":            whySyntax,
	"mapmodel/commands.go:RouteCommand":           whySyntax,
	"mapmodel/commands.go:WonderCommand":          whySyntax,
	"ui/mapstyle/skyline/wonders.go:wonderSprite": whyArt,
}

// newRow starts a row a special rule fills.
func newRow(id, kind, where string) *row {
	return &row{id: id, kind: kind, where: where, age: -1}
}

// special handles the places in code that the general reader would get
// wrong or name badly. done is false when the unit is not one of them.
func (cr *codeReader) special(u *unit) (r *row, why string, done bool) {
	c, w := cr.c, cr.c.w
	rel := cr.f.rel
	if why, ok := skipDecls[rel+":"+u.decl]; ok {
		return nil, why, true
	}
	s := u.st()
	cs := u.call()
	inCall := len(u.calls) > 0 && cs.out == 0 // the call is closer than any struct literal
	arg := func(i int) string {
		if i < len(cs.args) {
			return cs.args[i]
		}
		return ""
	}

	// A String method names a value for whoever prints it.
	if u.fn == "String" && u.recv != "" {
		if strings.Contains(u.text, "(?)") {
			return nil, whyDev, true
		}
		switch u.recv {
		case "Tone", "Moment", "Polarity", "Kind", "Role":
			return nil, whyDev, true
		}
	}

	switch rel {
	case "ui/commands.go":
		path := func() string {
			var names []string
			for i := len(u.structs) - 1; i >= 0; i-- {
				if u.structs[i].typ == "ui.Command" && u.structs[i].sib["Name"] != "" {
					names = append(names, u.structs[i].sib["Name"])
				}
			}
			return strings.Join(names, ".")
		}
		helpRow := func(id, form string) *row {
			cr.n[id]++
			if cr.n[id] > 1 {
				id = fmt.Sprintf("%s.%d", id, cr.n[id])
			}
			r := newRow(id, kindMessage, "What the Help panel says the command “"+form+"” does.")
			if u.chain != nil {
				r.where += " It is one piece of that line, which reads: “" + oneLine(u.chain.pattern) + "” (… is a number the game fills in)."
			}
			return r
		}
		switch {
		case inCall && cs.fn == "sub" && cs.arg == 2:
			return helpRow("ui.help.commands."+idDots(strings.TrimPrefix(path()+"."+arg(0), ".")), arg(1)), "", true
		case inCall && cs.fn == "confirmYes" && cs.arg == 1:
			return helpRow("ui.help.commands."+idSafe(arg(0)), arg(0)), "", true
		case inCall && cs.fn == "panel" && cs.arg == 1:
			return newRow("ui.help.panels."+idSafe(arg(0)), kindMessage, "What the Help panel's list of panels says the command “"+arg(0)+"” opens."), "", true
		case inCall:
			return nil, "", false
		case s.typ == "ui.Usage" && s.field == "Text":
			name := path()
			if len(s.coll) > 0 && s.coll[0].index > 0 {
				name += fmt.Sprintf(".%d", s.coll[0].index+1)
			}
			return helpRow("ui.help.commands."+idDots(name), s.sib["Form"]), "", true
		case s.typ == "ui.Usage":
			return nil, whySyntax, true
		case s.typ == "ui.Command" && s.field == "Panel":
			return newRow("ui.help.panels."+idSafe(s.sib["Name"]), kindMessage, "What the Help panel's list of panels says the command “"+s.sib["Name"]+"” opens."), "", true
		case s.typ == "ui.Command" || s.typ == "ui.Arg":
			return nil, whySyntax, true
		}

	case "game/savenames.go":
		if !hasWord(u.text) {
			return nil, whyNoWords, true
		}
		list := u.decl
		cr.n[list]++
		r := newRow(fmt.Sprintf("save_name.%s.%03d", list, cr.n[list]), kindVoice,
			"One of the words and phrases the game puts together when it suggests a name for a new save (the list "+list+"). The name becomes a file name: letters and single spaces only.")
		if u.fn != "" {
			r.where = "A joining word the game puts between the pieces of a suggested save name. The name becomes a file name: letters and single spaces only."
		}
		r.area, r.rules, r.seen, r.age = areaSaveNames, rulesSaveNam, seenPool, 0
		return r, "", true

	case "game/military.go":
		if s.typ == "game.ExpeditionDef" && !inCall {
			key := s.sib["Key"]
			switch s.field {
			case "Name":
				r := newRow("expedition."+idSafe(key)+".name", kindName, "The name of an expedition or campaign: in the Expeditions and Army panels, in the log, and as the {subject} of the story lines about it.")
				r.area, r.age, r.seen = areaWorld, w.ageOf(s.sib["MinAge"]), seenRegular
				return r, "", true
			case "Description":
				r := newRow("expedition."+idSafe(key)+".description", kindVoice, "The description of “"+s.sib["Name"]+"” in the Expeditions and Army panels.")
				r.area, r.age, r.seen = areaWorld, w.ageOf(s.sib["MinAge"]), seenRegular
				return r, "", true
			}
			return nil, whyKey, true
		}
	}

	if cr.f.dir == "config" {
		switch {
		case u.decl == "buildingFlavor" && u.near == "elem":
			r := c.fromRule(fieldRule{id: "building.{k}.flavor", kind: kindVoice, area: areaBuildings, age: "{k}", seen: seenRegular,
				where: "The line in italics under the description of {name:k} in the Buildings list and the Buildings panel."}, u)
			if w.wonder[u.coll[0].key] {
				r.area = areaAges
				r.where = strings.Replace(r.where, "in the Buildings list and the Buildings panel", "in the Wonders panel", 1)
			}
			return r, "", true
		case u.decl == "logFlavorPools" && u.near == "elem":
			when := map[string]string{
				"building_complete": "when a building is finished", "research_done": "when a tech is researched",
				"age_advance": "when you advance an age", "starvation": "when workers starve",
				"starvation_ended": "when a famine ends", "catastrophe_survived": "after you endure a catastrophe",
			}
			key := u.coll[len(u.coll)-1].key
			r := c.fromRule(fieldRule{id: "log_flavor.{k}.{n}", kind: kindVoice, area: areaWorld, seen: seenPool, rules: rulesNoMark, age: "primitive_age",
				where: "One of the short lines the log sometimes adds " + when[key] + ". It has to fit every age, the first and the last alike."}, u)
			if key == "catastrophe_survived" {
				r.age = w.ageOf("iron_age")
			}
			return r, "", true
		case u.decl == "CatastropheInfo" && u.near == "return":
			era := u.caseKey
			if _, ok := w.eraName[era]; !ok {
				return nil, whyDev, true // the fallback for an era that does not exist
			}
			var r *row
			if u.ret == 0 {
				r = newRow("catastrophe."+era+".name", kindName, "The name of the catastrophe that can end the "+w.eraName[era]+": in the Harbinger panel, the window where you choose, the log and the badges.")
			} else {
				r = newRow("catastrophe."+era+".text", kindVoice, "The story the window tells when the catastrophe of the "+w.eraName[era]+" strikes and you must choose to endure or succumb.")
			}
			r.area, r.seen, r.age = areaEvents, seenRare, w.keyAge["catastrophe:"+era]
			r.rules = rulesConfig | rulesDocs
			if u.ret == 1 && !config.CatastropheAllowed(era) {
				r.unsure = true
				r.where += " No catastrophe can strike in the " + w.eraName[era] + ", so as things stand this story is never shown."
			}
			return r, "", true
		case u.decl == "LastPassageInfo" && u.near == "return":
			r := newRow("last_passage.name", kindName, "The name of the Last Passage, the end of the Cosmic Era: in the Harbinger panel, its window, the log and the badges.")
			if u.ret == 1 {
				r = newRow("last_passage.text", kindVoice, "The story the Last Passage's window tells.")
			}
			r.area, r.seen, r.age = areaEvents, seenRare, w.ageOf("interstellar_age")
			r.rules = rulesConfig | rulesDocs
			return r, "", true
		case inCall && cs.fn == "wc" && cs.arg == 2:
			r := newRow("worker_class."+idSafe(arg(0))+"."+idSafe(arg(1)), kindName, "What the Workers panel calls the workers of the "+strings.ReplaceAll(arg(0), "_", " ")+" domain from the "+w.called(arg(1))+" on.")
			r.area, r.seen, r.age = areaAges, seenRegular, w.ageOf(arg(1))
			r.rules = rulesConfig | rulesDocs
			return r, "", true
		case len(u.calls) > 0 && cs.fn == "ladder" && cs.arg == 6 && u.near == "elem":
			r := newRow(fmt.Sprintf("badge_ladder.%s.rung.%d", idSafe(arg(0)), u.coll[0].index+1), kindName,
				fmt.Sprintf("The name of badge %d of the ladder “%s”: in the badge case, and in the toast and log line when it is earned.", u.coll[0].index+1, arg(1)))
			r.area, r.seen, r.rules = areaBadges, seenRegular, rulesConfig|rulesDocs
			return r, "", true
		case inCall && cs.fn == "ladder":
			key := idSafe(arg(0))
			var r *row
			switch cs.arg {
			case 1:
				r = newRow("badge_ladder."+key+".heading", kindLabel, "The heading the badge case puts over the ladder “"+arg(1)+"”.")
			case 3:
				r = newRow("badge_ladder."+key+".description", kindVoice, "What the badge case says each badge of the ladder “"+arg(1)+"” is for. {count} is the number that rung asks for.")
			case 4:
				r = newRow("badge_ladder."+key+".first", kindVoice, "What the badge case says the first badge of the ladder “"+arg(1)+"” is for, when it asks for just one.")
			default:
				return nil, whyKey, true
			}
			r.area, r.seen, r.rules = areaBadges, seenRegular, rulesConfig|rulesDocs
			return r, "", true
		case inCall && cs.fn == "mapLook" && (cs.arg == 1 || cs.arg == 2):
			r := newRow("badge."+idSafe(arg(0))+".name", kindName, "The name of a badge: in the badge case, and in the toast and log line when it is earned.")
			if cs.arg == 2 {
				r = newRow("badge."+idSafe(arg(0))+".description", kindVoice, "What the badge case says the badge “"+arg(1)+"” is for.")
			}
			r.area, r.seen, r.rules = areaBadges, seenRegular, rulesConfig|rulesDocs
			return r, "", true
		case inCall && (cs.fn == "Occurs" || cs.fn == "BotProof" || cs.fn == "StaticProof" || cs.fn == "never" || cs.fn == "atLeast" ||
			strings.HasPrefix(cs.fn, "RevealUntil") || cs.fn == "mapLook" || cs.fn == "bad"):
			return nil, "a note or key for developers: how a badge is proved or judged", true
		case u.decl == "MilestoneCategoryNames" && u.near == "elem":
			r := c.fromRule(fieldRule{id: "milestone_category.{k}", kind: kindLabel, area: areaMilestones, seen: seenRegular, age: "primitive_age",
				where: "The heading of a group of milestones in the Milestones panel."}, u)
			return r, "", true
		case u.decl == "BadgeCompleteTitle":
			r := newRow("badge_title.complete", kindName, "The title an account can wear once it holds every counted badge.")
			r.area, r.seen, r.rules = areaBadges, seenRare, rulesConfig|rulesDocs
			return r, "", true
		}
	}

	switch {
	case rel == "mapmodel/catalog.go" && u.decl == "LineageNames" && u.near == "elem":
		r := c.fromRule(fieldRule{id: "map.lineage.{k}", kind: kindLabel, area: areaMap, seen: seenRegular, age: "primitive_age",
			where: "What the map calls a kind of building, in its legend and in the line about a district."}, u)
		return r, "", true
	case rel == "rules/names.go" && strings.HasPrefix(u.decl, "Count"):
		return nil, "a name the tests count the wiki's numbers by", true
	}
	return nil, "", false
}

// idDots makes an id out of a dotted path, keeping the dots.
func idDots(s string) string {
	parts := strings.Split(s, ".")
	for i, p := range parts {
		parts[i] = idSafe(p)
	}
	return strings.Join(parts, ".")
}
