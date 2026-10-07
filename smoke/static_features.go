package smoke

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// The feature locks (config.FeatureLocks): the commands that wait for a
// tech. A lock is live when its tech is in the tree. A lock whose tech is
// not there yet is inert, its command open, and it switches on the day a
// tech with that key is added, with nothing else to change. This check lists
// which is which, so a content change that adds such a tech shows up here
// as a lock that moved from waiting to live.

// FeatureLockRow is one feature lock as the static report lists it.
type FeatureLockRow struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	// Tech is the key of the tech that opens the command; TechName and
	// TechAge are empty while the tech is not in the tree.
	Tech     string `json:"tech"`
	TechName string `json:"tech_name,omitempty"`
	TechAge  string `json:"tech_age,omitempty"`
	// Live reports that the tech is in the tree and the lock applies.
	Live bool `json:"live"`
}

// StaticFeatureLocks lists every feature lock of the core ruleset, in the
// order the tree opens them.
func StaticFeatureLocks() []FeatureLockRow {
	set := rules.Core()
	var out []FeatureLockRow
	for _, d := range set.FeatureLocks() {
		row := FeatureLockRow{Key: d.Key, Name: d.Name, Tech: d.Tech}
		if t, ok := set.Tech(d.Tech); ok {
			row.TechName, row.TechAge, row.Live = t.Name, t.Age, true
		}
		out = append(out, row)
	}
	return out
}

// FeatureLocksWaiting is the keys of the locks whose tech is not in the tree
// yet, in order.
func FeatureLocksWaiting(rows []FeatureLockRow) []string {
	var out []string
	for _, r := range rows {
		if !r.Live {
			out = append(out, r.Key)
		}
	}
	return out
}

// writeFeatureLocks renders the feature locks.
func writeFeatureLocks(sb *strings.Builder, rows []FeatureLockRow) {
	sb.WriteString("The commands that wait for a tech (config.FeatureLocks). A lock is live when its tech is in the tree: the command is refused until the tech is researched, with the tech's name in the refusal. A lock that waits for its tech is inert and its command is open; adding a tech with that key switches the lock on, with nothing else to change. A run that already used a command keeps it either way.\n\n")
	sb.WriteString("| command | opened by | tech's age | lock |\n|---|---|---|---|\n")
	for _, r := range rows {
		tech, age, state := "`"+r.Tech+"`", "-", "waits for its tech (open)"
		if r.Live {
			tech, age, state = r.TechName, ageName(r.TechAge), "live"
		}
		fmt.Fprintf(sb, "| %s | %s | %s | %s |\n", r.Name, tech, age, state)
	}
	if waiting := FeatureLocksWaiting(rows); len(waiting) > 0 {
		fmt.Fprintf(sb, "\n%d of %d locks wait for their tech: %s.\n", len(waiting), len(rows), strings.Join(waiting, ", "))
	}
}

// ageName is an age's display name without " Age" ("Bronze").
func ageName(key string) string {
	a, ok := config.AgeByKey()[key]
	if !ok {
		return key
	}
	return strings.TrimSuffix(a.Name, " Age")
}
