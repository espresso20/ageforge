package rules

import (
	"maps"
	"sort"
	"strings"
	"unicode"

	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
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

// The quantities Counts reports, as the docs word them.
const (
	CountAges        = "ages"
	CountBuildings   = "buildings"
	CountTechs       = "technologies"
	CountMilestones  = "milestones"
	CountChains      = "milestone chains"
	CountResources   = "resources"
	CountEras        = "epochs"
	CountDomains     = "worker domains"
	CountShopItems   = "prestige shop items"
	CountLineages    = "lineages"
	CountLines       = "building lines"
	CountWonders     = "wonders"
	CountCivs        = "civilizations"
	CountRoutes      = "trade routes"
	CountLineageBld  = "lineage buildings"
	CountStorage     = "storage buildings"
	CountMonuments   = "cultural monuments"
	CountStandalone  = "standalone buildings"
	lineageStorage   = "storage"
	lineageWonder    = "wonder"
	lineageMonuments = "monument"
)

func (s *Set) buildCounts() {
	c := map[string]int{
		CountAges: len(s.ages), CountBuildings: len(s.buildings), CountTechs: len(s.techs),
		CountMilestones: len(s.milestones), CountChains: len(s.chains), CountResources: len(s.resources),
		CountEras: len(s.eras), CountDomains: len(s.domains), CountShopItems: len(s.shop),
		CountCivs: len(s.factions), CountRoutes: len(s.routes),
	}
	lineages, lines := map[string]bool{}, map[string]bool{}
	for _, d := range s.buildings {
		switch {
		case d.Category == "wonder":
			c[CountWonders]++
		case d.Category == "storage":
			c[CountStorage]++
		case d.Category == "monument":
			c[CountMonuments]++
		case d.LineageKey != "" && d.LineageKey != lineageStorage && d.LineageKey != lineageWonder && d.LineageKey != lineageMonuments:
			c[CountLineageBld]++
			lineages[d.LineageKey] = true
			lines[d.Line()] = true
		default:
			c[CountStandalone]++
		}
	}
	c[CountLineages] = len(lineages)
	c[CountLines] = len(lines)
	s.counts = c
}

// Counts returns the numbers the docs quote about the game: how many ages,
// buildings, technologies and so on the set defines, keyed by the Count
// constants.
func (s *Set) Counts() map[string]int { return maps.Clone(s.counts) }

// Slug is a name the way it is typed at the command line: lower case, every
// run of anything but letters and digits one underscore ("Crash Pad" is
// "crash_pad"), the same shape as the config keys.
func Slug(name string) string {
	var b strings.Builder
	gap := false
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if gap && b.Len() > 0 {
				b.WriteByte('_')
			}
			gap = false
			b.WriteRune(r)
			continue
		}
		gap = true
	}
	return b.String()
}

// Typed is what a player types for key while VibeMode is on: the slug of the
// name the screen shows. With VibeMode off, or for a key the set has no name
// for, it is the key as it is.
func (s *Set) Typed(kind Kind, key string) string {
	if s == nil || !vibe.On() || kind >= numKinds {
		return key
	}
	if n := s.names[kind][key]; n != "" {
		if t := Slug(n); t != "" {
			return t
		}
	}
	return key
}

// KeyFor reads what a player typed for a thing of kind: the key, the slug of
// the name this set shows, or the slug of the plain name. With prefix, a word
// that starts a name's slug and no other's ("crash" for "Crash Pad") names
// it too. ok is false when nothing fits, or when a prefix fits two.
func (s *Set) KeyFor(kind Kind, typed string, prefix bool) (key string, ok bool) {
	if kind >= numKinds {
		return "", false
	}
	t := Slug(typed)
	if t == "" {
		return "", false
	}
	if _, isKey := s.names[kind][t]; isKey {
		return t, true
	}
	sets := []*Set{s}
	if plain := plainCore.get(); plain != s {
		sets = append(sets, plain)
	}
	keys := make([]string, 0, len(s.names[kind]))
	for k := range s.names[kind] {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, set := range sets {
		for _, k := range keys {
			if Slug(set.names[kind][k]) == t {
				return k, true
			}
		}
	}
	if !prefix {
		return "", false
	}
	found := ""
	for _, set := range sets {
		for _, k := range keys {
			if strings.HasPrefix(Slug(set.names[kind][k]), t+"_") && k != found {
				if found != "" {
					return "", false
				}
				found = k
			}
		}
	}
	return found, found != ""
}
