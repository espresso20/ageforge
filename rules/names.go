package rules

import (
	"strings"

	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Kind says what a key names, for Name.
type Kind uint8

const (
	KindAge Kind = iota
	KindBuilding
	KindTech
	KindResource
	KindCiv
	KindRoute
	numKinds
)

func (s *Set) buildNames() {
	for k := range s.names {
		s.names[k] = map[string]string{}
	}
	for _, d := range s.ages {
		s.names[KindAge][d.Key] = d.Name
	}
	for _, d := range s.buildings {
		s.names[KindBuilding][d.Key] = d.Name
	}
	for _, d := range s.techs {
		s.names[KindTech][d.Key] = d.Name
	}
	for _, d := range s.resources {
		s.names[KindResource][d.Key] = d.Name
	}
	for _, d := range s.factions {
		s.names[KindCiv][d.Key] = d.Name
	}
	for _, d := range s.routes {
		s.names[KindRoute][d.Key] = d.Name
	}
}

// Name returns the display name of key ("Industrial Age", "Lumber Mill").
// A key the set has no name for reads as the key itself, spaced and
// capitalized: "iron_ore" becomes "Iron ore".
func (s *Set) Name(kind Kind, key string) string {
	if kind < numKinds {
		if n := s.names[kind][key]; n != "" {
			return n
		}
	}
	return textfmt.Capitalize(strings.ReplaceAll(key, "_", " "))
}

// Lineages that the wonder and monument rules single out.
const (
	lineageWonder    = "wonder"
	lineageMonuments = "monument"
)
