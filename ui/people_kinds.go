package ui

import (
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// people_kinds.go is the four kinds of person in the town, as the People box
// and the People panel list them: workers (they have a job), the reserve
// (they live in the homes and wait for one), soldiers and scholars. Each is
// called by the age's own word (game.GameState.Term).

// peopleKind is one kind of person.
type peopleKind struct {
	key   string // config.TermWorkers and the rest
	one   string // the age's singular, as the lexicon writes it
	name  string // the age's plural: "Craftsfolk"
	count int
	// out is the soldiers away on missions (the soldiers only).
	out int
	// posts is every post the soldiers have; more than count when posts wait
	// on food (the soldiers only, 0 when none do).
	posts int
	// role colours the kind in the People box and in the panel's bar.
	role theme.Role
	// kind is the panel's colour kind for the same role.
	kind xk
}

// peopleKinds lists the four kinds in the order the game shows them. The
// soldiers are the soldiers at their posts (Workers.Soldiers: the military
// buildings' own), with those on missions counted apart.
func peopleKinds(st game.GameState) []peopleKind {
	ws := st.Workers
	out := st.Military.SoldiersAway
	mk := func(key string, count int, role theme.Role, kind xk) peopleKind {
		w := st.Term(key)
		return peopleKind{key: key, one: w.One, name: w.Many, count: count, role: role, kind: kind}
	}
	kinds := []peopleKind{
		mk(config.TermWorkers, ws.Working(), theme.RoleHighlight, xWork),
		mk(config.TermReserve, ws.Reserve(), theme.RolePositive, xRes),
		mk(config.TermSoldiers, ws.Soldiers, theme.RoleNegative, xArmy),
		mk(config.TermScholars, ws.Scholars, theme.RoleLabel, xLore),
	}
	kinds[2].out = out
	if st.Military.SoldierCap > ws.Soldiers {
		kinds[2].posts = st.Military.SoldierCap
	}
	return kinds
}

// txtOut is said the same way on several screens, so it has one id.
var txtOut = vibe.E("ui.people_kinds.boxLine.01", " out")

// boxLine is the kind's line in the People box w cells wide (0: any width):
// its name on the left, its count on the right, and, for the soldiers, "·
// N out" after the count when there is room.
func (k peopleKind) boxLine(w int) string {
	count := textfmt.Number(float64(k.count))
	if k.posts > k.count {
		count += "/" + textfmt.Number(float64(k.posts))
	}
	if k.out > 0 {
		// Only when the whole name still fits beside it.
		for _, with := range []string{
			count + " · " + textfmt.Number(float64(k.out)) + txtOut.String(),
			count + " +" + textfmt.Number(float64(k.out)) + txtOut.String(),
		} {
			if w <= 0 || runeLen(k.name)+1+runeLen(with) <= w {
				count = with
				break
			}
		}
	}
	name := k.name
	if w > 0 {
		name = fitText(name, max(1, w-runeLen(count)-1))
	}
	pad := " "
	if w > 0 {
		pad = strings.Repeat(" ", max(1, w-runeLen(name)-runeLen(count)))
	}
	return theme.Tag(k.role) + name + "[-]" + pad + "[white]" + count + "[-]"
}
