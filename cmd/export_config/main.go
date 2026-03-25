// export_config serializes all static game config data to JSON files
// for consumption by the Godot 4 C# project.
//
// Usage:
//
//	go run ./cmd/export_config --out /path/to/project-ageforge/data
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/espresso20/ageforge/config"
)

func main() {
	out := flag.String("out", "../../data", "output directory for JSON files")
	flag.Parse()

	outDir, err := filepath.Abs(*out)
	if err != nil {
		fatalf("resolving output dir: %v", err)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fatalf("creating output dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(outDir, "buildings"), 0755); err != nil {
		fatalf("creating buildings dir: %v", err)
	}

	fmt.Printf("Exporting config to: %s\n", outDir)

	write(outDir, "ages.json", config.Ages())
	write(outDir, "resources.json", config.BaseResources())
	write(outDir, "techs.json", config.Technologies())
	write(outDir, "milestones.json", config.Milestones())
	write(outDir, "milestone_chains.json", config.MilestoneChains())
	write(outDir, "milestone_titles.json", config.MilestoneTitles())
	write(outDir, "epochs.json", config.Epochs())
	write(outDir, "events.json", config.RandomEvents())
	write(outDir, "trade_routes.json", config.BaseTradeRoutes())
	write(outDir, "exchange_rates.json", config.BaseExchangeRates())
	write(outDir, "factions.json", config.BaseFactions())
	write(outDir, "worker_classes.json", config.WorkerClasses())
	write(outDir, "building_upgrades.json", config.BuildingUpgrades())
	write(outDir, "prestige_upgrades.json", config.PrestigeUpgrades())

	// Buildings: one file per lineage + storage/wonders
	writeBuildingLineages(outDir)

	fmt.Println("Done.")
}

// writeBuildingLineages splits production buildings by lineage and writes
// storage/wonder buildings to a separate file — matching the architecture.md
// project structure.
func writeBuildingLineages(outDir string) {
	production := config.NewProductionBuildings()
	allBuildings := config.BaseBuildings() // production + storage + wonders combined

	// Index production keys so we can separate out storage/wonders
	productionKeys := make(map[string]bool, len(production))
	for _, b := range production {
		productionKeys[b.Key] = true
	}

	// Split production buildings by lineage
	byLineage := make(map[string][]config.BuildingDef)
	for _, b := range production {
		key := b.LineageKey
		if key == "" {
			key = "misc"
		}
		byLineage[key] = append(byLineage[key], b)
	}
	for lineage, defs := range byLineage {
		write(filepath.Join(outDir, "buildings"), "lineage_"+lineage+".json", defs)
	}

	// Storage and wonders are everything in BaseBuildings() not in production
	var storageAndWonders []config.BuildingDef
	for _, b := range allBuildings {
		if !productionKeys[b.Key] {
			storageAndWonders = append(storageAndWonders, b)
		}
	}
	write(filepath.Join(outDir, "buildings"), "storage_and_wonders.json", storageAndWonders)

	// Combined map keyed by building key — convenient for ConfigLoader.cs
	combined := make(map[string]config.BuildingDef, len(allBuildings))
	for _, b := range allBuildings {
		combined[b.Key] = b
	}
	write(outDir, "all_buildings.json", combined)

	fmt.Printf("  buildings: %d production + %d storage/wonder = %d total\n",
		len(production), len(storageAndWonders), len(allBuildings))
}

func write(dir, filename string, v any) {
	path := filepath.Join(dir, filename)
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fatalf("marshaling %s: %v", filename, err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fatalf("writing %s: %v", path, err)
	}
	fmt.Printf("  wrote %s (%d bytes)\n", filename, len(data))
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "export_config: "+format+"\n", args...)
	os.Exit(1)
}
