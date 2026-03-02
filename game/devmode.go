package game

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// devKeyHash is the SHA256 of the developer passphrase.
const devKeyHash = "ef3d0375543ef25277bf004b685160c420813e13047396914538e9803f4ead4f"

// DevModeActive is true once the developer passphrase has been accepted.
// It is never persisted to disk.
var DevModeActive bool

// DevGodMode disables resource costs and skips build queues when true.
var DevGodMode bool

// CheckDevKey returns true and activates dev mode if the passphrase matches.
func CheckDevKey(input string) bool {
	h := sha256.Sum256([]byte(strings.TrimSpace(input)))
	got := fmt.Sprintf("%x", h)
	if got == devKeyHash {
		DevModeActive = true
		return true
	}
	return false
}

// DevExecCommand runs a developer command against the engine.
// Returns a result message. Commands:
//
//	/age <key>         — jump to any age
//	/fill              — fill all resources to storage cap
//	/give <res> <n>    — add amount to resource
//	/techs             — unlock all techs available up to current age
//	/build <key>       — instantly place one building
//	/prestige <n>      — set prestige level (0-9)
//	/speed <n>         — set tick speed multiplier
//	/god               — toggle godmode (free costs, instant builds)
//	/ages              — list all age keys
func DevExecCommand(cmd string, ge *GameEngine) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}
	switch strings.ToLower(parts[0]) {

	case "/age":
		if len(parts) < 2 {
			return "usage: /age <age_key>"
		}
		key := parts[1]
		// Validate key exists
		found := false
		for _, a := range config.Ages() {
			if a.Key == key {
				found = true
				break
			}
		}
		if !found {
			return "unknown age: " + key
		}
		ge.mu.Lock()
		ge.age = key
		ge.applyAgeUnlocks(key)
		ge.mu.Unlock()
		return "jumped to " + key

	case "/fill":
		ge.mu.Lock()
		snap := ge.Resources.GetAllStorage()
		for res, cap := range snap {
			ge.Resources.Add(res, cap)
		}
		ge.mu.Unlock()
		return "all resources filled to cap"

	case "/give":
		if len(parts) < 3 {
			return "usage: /give <resource> <amount>"
		}
		var amount float64
		if _, err := fmt.Sscanf(parts[2], "%f", &amount); err != nil {
			return "invalid amount: " + parts[2]
		}
		ge.mu.Lock()
		ge.Resources.Add(parts[1], amount)
		ge.mu.Unlock()
		return fmt.Sprintf("gave %.0f %s", amount, parts[1])

	case "/techs":
		ge.mu.Lock()
		ageOrder := ge.progress.GetAgeOrder()
		currentOrder := ageOrder[ge.age]
		count := 0
		for _, def := range config.Technologies() {
			if ge.Research.researched[def.Key] {
				continue
			}
			if ageOrder[def.Age] <= currentOrder {
				ge.Research.researched[def.Key] = true
				count++
			}
		}
		ge.mu.Unlock()
		return fmt.Sprintf("unlocked %d techs", count)

	case "/build":
		if len(parts) < 2 {
			return "usage: /build <building_key>"
		}
		key := parts[1]
		ge.mu.Lock()
		if _, ok := ge.Buildings.defs[key]; !ok {
			ge.mu.Unlock()
			return "unknown building: " + key
		}
		ge.Buildings.UnlockBuilding(key)
		ge.Buildings.counts[key]++
		ge.mu.Unlock()
		return "built " + key

	case "/prestige":
		if len(parts) < 2 {
			return "usage: /prestige <level 0-9>"
		}
		var level int
		if _, err := fmt.Sscanf(parts[1], "%d", &level); err != nil || level < 0 || level > 9 {
			return "level must be 0-9"
		}
		ge.mu.Lock()
		ge.Prestige.level = level
		ge.mu.Unlock()
		return fmt.Sprintf("prestige level set to %d", level)

	case "/speed":
		if len(parts) < 2 {
			return "usage: /speed <multiplier>"
		}
		var mult float64
		if _, err := fmt.Sscanf(parts[1], "%f", &mult); err != nil || mult <= 0 {
			return "multiplier must be > 0"
		}
		ge.mu.Lock()
		ge.speedMultiplier = mult
		ge.mu.Unlock()
		return fmt.Sprintf("tick speed set to %.0fx", mult)

	case "/god":
		DevGodMode = !DevGodMode
		if DevGodMode {
			return "godmode ON — free costs, instant builds"
		}
		return "godmode OFF"

	case "/ages":
		var names []string
		for _, a := range config.Ages() {
			names = append(names, a.Key)
		}
		return strings.Join(names, "  ")

	default:
		return "unknown command: " + parts[0] + "  (try /age /fill /give /techs /build /prestige /speed /god /ages)"
	}
}
