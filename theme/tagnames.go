package theme

// roleTagNames is the semantic tag vocabulary: each role has a lowercase name
// that works as an inline color tag ([accent]…[-], [onaccent:accent]…).
// These are the names new code should write (via Tag/TagFgBg in tags.go).
var roleTagNames = [numRoles]string{
	RoleBackground:    "bg",
	RoleText:          "text",
	RoleDim:           "dim",
	RoleLabel:         "label",
	RoleAccent:        "accent",
	RoleHighlight:     "highlight",
	RolePositive:      "positive",
	RoleNegative:      "negative",
	RoleSelection:     "selection",
	RoleSurface:       "surface",
	RoleBorder:        "border",
	RoleSelectionText: "seltext",
	RoleBright:        "bright",
	RoleWarning:       "warning",
	RoleOnAccent:      "onaccent",
	RoleChip:          "chip",
	RoleOnNegative:    "onnegative",
	RoleComment:       "comment",
	RoleEffect:        "effect",
	RoleNew:           "new",
}

// tagName pairs a color name the theme owns with the role it resolves to.
type tagName struct {
	name string
	role Role
}

// legacyAliases are the pre-role color names the UI was written with (~900
// inline tags). They are remapped onto roles so every existing [gold]/[green]
// tag retints with no edit. They are part of the theme-owned vocabulary: the
// raw-color guard accepts them and rejects every other
// name or hex literal outside theme/.
//
// The kit's tag parser asks TagRole for a name first, so one of these always
// reads as its role's color in the active theme, never as the CSS color of
// the same name.
var legacyAliases = []tagName{
	{"gold", RoleAccent},
	{"gray", RoleDim},
	{"cyan", RoleLabel},
	{"green", RolePositive},
	{"red", RoleNegative},
	{"yellow", RoleHighlight},
	{"white", RoleText},
}

// remappedNames is every color name the theme owns (role names + legacy
// aliases). Exposed to tests via TagNames.
func remappedNames() []tagName {
	out := make([]tagName, 0, len(legacyAliases)+int(numRoles))
	out = append(out, legacyAliases...)
	for r := Role(0); r < numRoles; r++ {
		out = append(out, tagName{roleTagNames[r], r})
	}
	return out
}

// TagNames returns every inline-tag color name the theme owns, mapped to its
// role. The raw-color guard uses it as the allow-list.
func TagNames() map[string]Role {
	m := make(map[string]Role)
	for _, n := range remappedNames() {
		m[n.name] = n.role
	}
	return m
}
