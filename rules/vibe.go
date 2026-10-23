package rules

import (
	"sync"
	"sync/atomic"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/words"
)

// vibe.go: the VibeMode overlay.
//
// VibeMode's script is a line per row id. Vibed puts it into a Source: every
// text of a definition that has a row in the script is swapped for the
// script's line, and every other text, key and number is left as it was. The
// result compiles to a ruleset like any other, so the whole game reads the
// script through the set it plays by, with no second code path.

// Vibed returns src with every text VibeMode's script has a line for put in.
// It does not change src or the tables of package config it was built from.
func Vibed(src Source) Source {
	out, _ := words.Rewrite(src, func(c words.Cell) (string, bool) {
		line, ok := vibe.Text(c.ID)
		if !ok {
			return "", false
		}
		return config.ComposeText(c.Owner, c.Field, line), true
	})
	return out
}

// lazySet is a set built the first time it is asked for.
type lazySet struct {
	once  sync.Once
	build func() *Set
	set   atomic.Pointer[Set]
}

func (l *lazySet) get() *Set {
	l.once.Do(func() { l.set.Store(l.build()) })
	return l.set.Load()
}

// is reports whether s is this set, without building it.
func (l *lazySet) is(s *Set) bool { return s != nil && l.set.Load() == s }

var (
	plainCore = &lazySet{build: func() *Set { return Compile(FromConfig()) }}
	vibeCore  = &lazySet{build: func() *Set { return Compile(Vibed(FromConfig())) }}
)

// CoreFor is the core set in plain English (vibe false) or in VibeMode's
// script (vibe true), each built on first use and shared from then on. Both
// are made from the same tables, so they differ in words alone.
func CoreFor(vibeOn bool) *Set {
	if vibeOn {
		return vibeCore.get()
	}
	return plainCore.get()
}

// IsCore reports whether s is one of the two core sets: a game on one of
// them follows the VibeMode switch; a game on a set of its own does not.
func IsCore(s *Set) bool { return plainCore.is(s) || vibeCore.is(s) }
