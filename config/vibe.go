package config

// vibe.go: what the game makes of a line a script swaps in.
//
// VibeMode's script replaces the authored line of a definition's text field.
// Almost always that line is what a player reads. Three fields are built
// from the authored line when the tables load, and a line swapped in later
// has to be built the same way, or the numbers they carry would be lost.

// ComposeText is the text a player reads when line is the authored line of
// the field of owner (a definition, "config.BuildingDef.Description"):
//
//   - a building's Description is followed by what the building does, in
//     numbers ("+1 food/tick (3 workers).");
//   - the FlavorText of an era event or an awakening has its "{dur}" written
//     in as how long the effect lasts.
//
// Any other field reads as the line.
func ComposeText(owner any, field, line string) string {
	switch d := owner.(type) {
	case BuildingDef:
		if field == "config.BuildingDef.Description" {
			if fx := buildingEffectText(d); fx != "" && line != "" {
				return line + " " + fx
			}
		}
	case EpochEventDef:
		if field == "config.EpochEventDef.FlavorText" {
			return withDuration(line, d.Duration)
		}
	case AwakeningDef:
		if field == "config.AwakeningDef.FlavorText" {
			return withDuration(line, d.Duration)
		}
	}
	return line
}
