package rules

import "github.com/espresso20/ageforge/config"

// RankTitles is the six rank titles, lowest first.
func (s *Set) RankTitles() [6]string { return config.RankTitles }

// RankThresholds is the badge points each rank starts at, lowest first.
func (s *Set) RankThresholds() [6]int { return config.RankThresholds }

// DomainOpens is the name of the age a worker domain's first class arrives in,
// "" for a domain with no workers.
func (s *Set) DomainOpens(domain string) string {
	for _, w := range config.WorkerClasses() {
		if w.Domain == domain {
			return s.ageByKey[w.AgeKey].Name
		}
	}
	return ""
}
