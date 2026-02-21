package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/user/ageforge/config"
	"github.com/user/ageforge/game"
	"github.com/user/ageforge/ui"
)

// b is a shorthand to create a building entry with just a count —
// category/unlocked are filled in from config automatically.
func b(count int) int { return count }

// Demo ages — one representative per era, with era-appropriate buildings
var demos = []struct {
	Name      string
	AgeKey    string
	Buildings map[string]int // key → count (category looked up from config)
}{
	{
		Name:   "01_primitive_age",
		AgeKey: "primitive_age",
		Buildings: map[string]int{
			"hut": 5, "gathering_camp": 3, "woodcutter_camp": 2,
			"stash": 2, "firepit": 1, "altar": 1, "stone_pit": 1,
		},
	},
	{
		Name:   "02_bronze_age",
		AgeKey: "bronze_age",
		Buildings: map[string]int{
			"hut": 8, "farm": 5, "lumber_mill": 3, "quarry": 2, "mine": 2,
			"market": 1, "library": 1, "house": 3, "warehouse": 2,
			"gathering_camp": 4, "woodcutter_camp": 3, "stash": 3,
			"stonehenge": 1,
		},
	},
	{
		Name:   "03_classical_age",
		AgeKey: "classical_age",
		Buildings: map[string]int{
			"house": 10, "farm": 8, "lumber_mill": 4, "quarry": 3, "mine": 3,
			"market": 2, "library": 2, "warehouse": 3, "forum": 2,
			"aqueduct": 2, "amphitheater": 1, "classical_vault": 2,
			"barracks": 2, "colosseum": 1, "parthenon": 1,
		},
	},
	{
		Name:   "04_medieval_age",
		AgeKey: "medieval_age",
		Buildings: map[string]int{
			"manor": 6, "house": 12, "farm": 10, "lumber_mill": 5, "quarry": 4,
			"mine": 4, "coal_mine": 2, "smithy": 3, "market": 3,
			"university": 2, "cathedral": 1, "castle": 2, "keep": 2,
			"barracks": 3, "great_library": 1,
		},
	},
	{
		Name:   "05_industrial_age",
		AgeKey: "industrial_age",
		Buildings: map[string]int{
			"apartment": 8, "manor": 4, "factory": 6, "oil_well": 3,
			"coal_mine": 4, "mine": 5, "industrial_depot": 3, "bank": 2,
			"university": 2, "barracks": 3, "power_grid": 2, "telegraph": 1,
		},
	},
	{
		Name:   "06_electric_age",
		AgeKey: "electric_age",
		Buildings: map[string]int{
			"apartment": 12, "electric_mill": 4, "train_station": 3,
			"telephone_exchange": 2, "factory": 5, "power_grid": 3,
			"electric_warehouse": 3, "clocktower": 2, "victorian_vault": 2,
			"barracks": 3,
		},
	},
	{
		Name:   "07_modern_age",
		AgeKey: "modern_age",
		Buildings: map[string]int{
			"skyscraper": 8, "apartment": 15, "power_plant": 3,
			"research_lab": 3, "modern_depot": 4, "factory": 6,
			"reactor": 2, "bunker": 2, "missile_silo": 1, "space_program": 1,
		},
	},
	{
		Name:   "08_information_age",
		AgeKey: "information_age",
		Buildings: map[string]int{
			"skyscraper": 12, "apartment": 20, "server_farm": 5,
			"fiber_hub": 4, "media_center": 3, "info_vault": 3,
			"research_lab": 4, "power_plant": 4, "bunker": 2,
		},
	},
	{
		Name:   "09_cyberpunk_age",
		AgeKey: "cyberpunk_age",
		Buildings: map[string]int{
			"neon_tower": 10, "augmentation_clinic": 4, "black_market": 3,
			"cyber_vault": 3, "data_center": 5, "ai_lab": 3,
			"smart_grid": 4, "digital_archive": 2, "skyscraper": 15,
			"particle_accelerator": 1,
		},
	},
	{
		Name:   "10_fusion_age",
		AgeKey: "fusion_age",
		Buildings: map[string]int{
			"fusion_reactor": 5, "plasma_forge": 4, "maglev_station": 3,
			"fusion_vault": 3, "neon_tower": 12, "ai_lab": 4,
			"data_center": 6, "augmentation_clinic": 3,
		},
	},
	{
		Name:   "11_space_age",
		AgeKey: "space_age",
		Buildings: map[string]int{
			"launch_pad": 4, "space_station": 3, "orbital_habitat": 6,
			"orbital_depot": 3, "fusion_reactor": 4, "plasma_forge": 3,
			"maglev_station": 2, "dyson_scaffold": 1,
		},
	},
	{
		Name:   "12_galactic_age",
		AgeKey: "galactic_age",
		Buildings: map[string]int{
			"galactic_hub": 3, "antimatter_plant": 4, "megastructure": 2,
			"galactic_vault": 3, "warp_gate": 3, "colony_ship": 4,
			"star_forge": 3, "stellar_vault": 2, "orbital_habitat": 8,
			"space_station": 4,
		},
	},
	{
		Name:   "13_transcendent_age",
		AgeKey: "transcendent_age",
		Buildings: map[string]int{
			"quantum_computer": 5, "reality_engine": 4,
			"transcendence_beacon": 3, "quantum_vault": 3,
			"singularity_core": 1, "megastructure": 3,
			"galactic_hub": 4, "antimatter_plant": 5, "orbital_habitat": 10,
		},
	},
}

// buildMap converts simple key→count map to proper BuildingState map
// by looking up Category from config definitions.
func buildMap(simple map[string]int) map[string]game.BuildingState {
	defs := config.BuildingByKey()
	result := make(map[string]game.BuildingState, len(simple))
	for key, count := range simple {
		cat := ""
		if def, ok := defs[key]; ok {
			cat = def.Category
		}
		result[key] = game.BuildingState{
			Count:    count,
			Unlocked: true,
			Category: cat,
		}
	}
	return result
}

func main() {
	outDir := "map_demos"
	os.MkdirAll(outDir, 0o755)

	// Render at a good resolution for viewing
	const width = 480
	const height = 360

	for _, d := range demos {
		buildings := buildMap(d.Buildings)
		img := ui.GenerateMapImage(ui.MapGenConfig{
			Width:       width,
			Height:      height,
			DetailLevel: 1,
			Buildings:   buildings,
			AgeKey:      d.AgeKey,
		})

		path := filepath.Join(outDir, d.Name+".png")
		f, err := os.Create(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating %s: %v\n", path, err)
			continue
		}
		if err := png.Encode(f, img); err != nil {
			fmt.Fprintf(os.Stderr, "error encoding %s: %v\n", path, err)
		}
		f.Close()
		fmt.Printf("wrote %s\n", path)
	}

	fmt.Printf("\nDone! %d maps generated in %s/\n", len(demos), outDir)
}
