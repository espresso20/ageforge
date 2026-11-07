package window

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings are what the window remembers between runs. They are the machine's,
// not an account's, so they sit in the game's data folder beside the accounts.
type Settings struct {
	Width      int  `json:"width"`       // window size, in device-independent pixels
	Height     int  `json:"height"`      //
	Fullscreen bool `json:"fullscreen"`  //
	Font       Font `json:"font"`        // the face
	FontSize   int  `json:"font_size"`   // mono size, 10 to 32
	PixelScale int  `json:"pixel_scale"` // the pixel font's multiple of its own pixel
}

const (
	minFontSize     = 10
	maxFontSize     = 32
	defaultFontSize = 16
	settingsFile    = "window.json"
)

func defaultSettings() Settings {
	return Settings{Font: FontMono, FontSize: defaultFontSize}
}

// clean puts every field in range. scale is the display's device scale.
func (s Settings) clean(scale float64) Settings {
	if s.Font != FontPixel {
		s.Font = FontMono
	}
	s.FontSize = min(max(s.FontSize, minFontSize), maxFontSize)
	lo, hi := pixelRange(scale)
	if s.PixelScale == 0 {
		s.PixelScale = pixelDefault(scale)
	}
	s.PixelScale = min(max(s.PixelScale, lo), hi)
	return s
}

// pixelRange is the multiples of the pixel font's pixel that land between 10
// and 32 points (a little over, so the smallest step is never out of reach) on
// a display with this device scale.
func pixelRange(scale float64) (lo, hi int) {
	lo = max(1, int(float64(minFontSize)*scale/pixelEm+0.5))
	hi = max(lo, int(float64(maxFontSize+2)*scale/pixelEm))
	return lo, hi
}

// pixelDefault is the multiple nearest 16 points.
func pixelDefault(scale float64) int {
	lo, hi := pixelRange(scale)
	return min(max(int(float64(defaultFontSize)*scale/pixelEm+0.5), lo), hi)
}

func settingsPath(dir string) string { return filepath.Join(dir, settingsFile) }

// loadSettings reads the saved settings; a missing or unreadable file is the defaults.
func loadSettings(dir string) Settings {
	s := defaultSettings()
	b, err := os.ReadFile(settingsPath(dir))
	if err != nil {
		return s
	}
	if json.Unmarshal(b, &s) != nil {
		return defaultSettings()
	}
	return s
}

// saveSettings writes the settings; a failure is not worth stopping the game for.
func saveSettings(dir string, s Settings) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath(dir), b, 0o644)
}

// gridSize is how many cells fit a window of w by h pixels.
func gridSize(w, h, cellW, cellH int) (cols, rows int) {
	if cellW < 1 || cellH < 1 {
		return 1, 1
	}
	return max(1, w/cellW), max(1, h/cellH)
}
