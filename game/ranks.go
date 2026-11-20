package game

import "github.com/espresso20/ageforge/config"

// rankCache is the account's domain points, kept for one badge revision.
type rankCache struct {
	ok     bool
	rev    uint64
	book   *badgeBook
	points map[string]int
}

// domainPoints is the points the account's earned badges give each worker
// domain. They are account-wide: they persist across runs and prestige, and a
// dev command never takes one away. An integrity badge gives none.
func (a *Account) domainPoints(book *badgeBook) map[string]int {
	a.mu.Lock()
	defer a.mu.Unlock()
	if c := &a.ranks; c.ok && c.rev == a.badgeRev && c.book == book {
		return c.points
	}
	pts := make(map[string]int, len(config.WorkerDomains()))
	for key := range a.Badges {
		if def := book.def(key); def != nil {
			pts[def.Domain()] += def.Points()
		}
	}
	a.ranks = rankCache{ok: true, rev: a.badgeRev, book: book, points: pts}
	return pts
}

// DomainRank is the rank, 0 to 6, the live account has reached in a worker
// domain. It takes the engine lock only if the caller does not hold it, so
// call it from outside the engine.
func (ge *GameEngine) DomainRank(domain string) int {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.domainRankLocked(domain)
}

// domainRankLocked is DomainRank for a caller that holds ge.mu.
func (ge *GameEngine) domainRankLocked(domain string) int {
	if ge.account == nil || ge.badges == nil {
		return 0
	}
	return config.RankFor(ge.account.domainPoints(ge.badges)[domain])
}

// rankOutputBonus is the share of a domain's worker output its rank adds.
// The caller holds ge.mu.
func (ge *GameEngine) rankOutputBonus(domain string) float64 {
	return config.RankBonus * float64(ge.domainRankLocked(domain))
}
