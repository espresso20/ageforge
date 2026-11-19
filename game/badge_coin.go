package game

import (
	"slices"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// badgeCoinOf is the coin a badge of a coin family is the moment of: its type
// and the age it belongs to. Both are "" for a badge of any other family.
func badgeCoinOf(set *rules.Set, def *config.BadgeDef) (typ, age string) {
	if !slices.Contains(config.CoinTypes, def.Family) {
		return "", ""
	}
	typ = def.Family
	rest := strings.TrimPrefix(def.Key, typ+".")
	switch typ {
	case config.CoinWonder:
		if b, ok := set.Building(rest); ok {
			return typ, b.RequiredAge
		}
	case config.CoinHarbinger:
		for _, a := range set.AgeKeys() {
			if h, ok := set.Harbinger(a); ok && h.Key == strings.TrimSuffix(rest, ".met") {
				return typ, h.Age
			}
		}
	default:
		return typ, rest
	}
	return "", ""
}

// finestCoinLocked is the finest copy the account holds of the coin a badge is
// the moment of; the zero copy when there is none. Callers hold a.mu.
func (a *Account) finestCoinLocked(set *rules.Set, def *config.BadgeDef) CoinCopy {
	typ, age := badgeCoinOf(set, def)
	if s, ok := a.coins[coinKey(typ, age)]; ok && len(s.Copies) > 0 {
		return s.Copies[0]
	}
	return CoinCopy{}
}
