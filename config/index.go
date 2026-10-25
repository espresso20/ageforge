package config

// indexBy maps every row to its key. When two rows share a key the later one
// wins, as it does for every table the ByKey lookups are built from.
func indexBy[T any](rows []T, key func(T) string) map[string]T {
	m := make(map[string]T, len(rows))
	for _, r := range rows {
		m[key(r)] = r
	}
	return m
}
