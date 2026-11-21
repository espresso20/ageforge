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
		s.named[k] = map[string]string{}
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
	for k, names := range s.names {
		for key, name := range names {
			n := plainName(strings.ToLower(name))
			if other, taken := s.named[k][n]; taken && other != key {
				key = "" // a name two things share finds neither
			}
			s.named[k][n] = key
		}
	}
}

// plainName is a name as a key would spell it: lower case, an underscore
// for each space or hyphen, no other punctuation.
var plainName = strings.NewReplacer(" ", "_", "-", "_", "'", "", "’", "", ".", "", ",", "").Replace

// KeyNamed returns the key that name spells: a key itself, or a display
// name, in any case and with spaces or hyphens between its words ("Lumber
// Mill", "lumber-mill" and "lumber_mill" are all lumber_mill). A key wins
// over another thing's name.
func (s *Set) KeyNamed(kind Kind, name string) (string, bool) {
	if kind >= numKinds {
		return "", false
	}
	n := plainName(strings.ToLower(name))
	if _, ok := s.names[kind][n]; ok {
		return n, true
	}
	key := s.named[kind][n]
	return key, key != ""
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
