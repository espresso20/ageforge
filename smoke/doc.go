// Package smoke holds AgeForge's smoke suite. What `go test ./smoke` runs reads
// the game's tables and drives the engine directly: the static checks (the
// Gate Covenant, storage, milestones, harbinger prices, faith and culture
// strength, depth points, research, badges, lineages, caps), the reference
// players and the economy's properties, the docs check (docsync), the accounts
// scenario and the walk of the ordinary player's purchases through every age.
//
// The test bot is parked. The bot that plays the game end to end, the
// scenarios built on it (progression, saveload, offline, fuzz, perf, styles,
// idle, prestige), their reports and cmd/smoke build only with the smokebot
// build tag; CONTRIBUTING.md, "Smoke testing", says how to run them.
package smoke
