package ui

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"sort"
	"strings"
	"sync"

	gameAssets "github.com/espresso20/ageforge/assets"
	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/sprites"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// MapV4 renders the player's capital city as a progressively growing pixel-art image
// using 16×16 domain sprites, half-block terminal rendering, and age-shifting terrain.
type MapV4 struct {
	mu             sync.Mutex
	state          game.GameState
	cachedImg      *image.RGBA
	cachedW        int
	cachedH        int
	cachedAge      string
	cachedBldCount int
	cachedEra      int
}

// NewMapV4 creates a new MapV4 instance.
func NewMapV4() *MapV4 { return &MapV4{} }

// ── Age palette ────────────────────────────────────────────────────────────

type v4AgePalette struct {
	ageKey        string
	bgDark        color.RGBA
	bgMid         color.RGBA
	bgLight       color.RGBA
	riverColor    color.RGBA
	clearingColor color.RGBA
	gridLines     bool
}

func v4rgb(hex uint32) color.RGBA {
	return color.RGBA{uint8(hex >> 16), uint8(hex >> 8), uint8(hex), 255}
}

var v4Palettes = []v4AgePalette{
	{"primitive_age", v4rgb(0x1a5c0a), v4rgb(0x2d8c1a), v4rgb(0x4ab02a), v4rgb(0x2060c0), v4rgb(0x8aba40), false},
	{"stone_age", v4rgb(0x5a7a3a), v4rgb(0x7a9a4a), v4rgb(0x9ab85a), v4rgb(0x1a50a0), v4rgb(0xb8cc70), false},
	{"bronze_age", v4rgb(0x6a7a2a), v4rgb(0x8a9a3a), v4rgb(0xaaba4a), v4rgb(0x1a50a0), v4rgb(0xc8c860), false},
	{"iron_age", v4rgb(0x4a6a2a), v4rgb(0x6a8a3a), v4rgb(0x8aaa4a), v4rgb(0x1a50a0), v4rgb(0xb0c860), false},
	{"classical_age", v4rgb(0x4a6030), v4rgb(0x6a8040), v4rgb(0x8aa050), v4rgb(0x1a4890), v4rgb(0xb0b870), false},
	{"medieval_age", v4rgb(0x3a5828), v4rgb(0x5a7838), v4rgb(0x7a9848), v4rgb(0x1a4080), v4rgb(0xa8b868), false},
	{"renaissance_age", v4rgb(0x384828), v4rgb(0x506838), v4rgb(0x688848), v4rgb(0x1a3870), v4rgb(0xa0b060), false},
	{"age_of_sail", v4rgb(0x304020), v4rgb(0x486030), v4rgb(0x607840), v4rgb(0x1a3868), v4rgb(0x98a858), false},
	{"industrial_age", v4rgb(0x404038), v4rgb(0x585850), v4rgb(0x707068), v4rgb(0x183060), v4rgb(0x888870), false},
	{"gilded_age", v4rgb(0x383830), v4rgb(0x505048), v4rgb(0x686860), v4rgb(0x182858), v4rgb(0x808068), false},
	{"modern_age", v4rgb(0x282830), v4rgb(0x383840), v4rgb(0x484850), v4rgb(0x182050), v4rgb(0x686870), false},
	{"atomic_age", v4rgb(0x201820), v4rgb(0x302830), v4rgb(0x403840), v4rgb(0x201848), v4rgb(0x605860), false},
	{"space_age", v4rgb(0x100820), v4rgb(0x201830), v4rgb(0x302840), v4rgb(0x301060), v4rgb(0x504860), true},
	{"information_age", v4rgb(0x080818), v4rgb(0x181828), v4rgb(0x282838), v4rgb(0x280a58), v4rgb(0x484858), true},
	{"cyberpunk_age", v4rgb(0x060610), v4rgb(0x101020), v4rgb(0x1a1a30), v4rgb(0x5010a0), v4rgb(0x303050), true},
	{"nanotech_age", v4rgb(0x040410), v4rgb(0x0c0c1c), v4rgb(0x14142c), v4rgb(0x4010a0), v4rgb(0x282848), true},
	{"fusion_age", v4rgb(0x040408), v4rgb(0x080818), v4rgb(0x101024), v4rgb(0x6010c0), v4rgb(0x202038), true},
	{"singularity_age", v4rgb(0x020208), v4rgb(0x060614), v4rgb(0x0c0c20), v4rgb(0x7010d0), v4rgb(0x181830), true},
	{"galactic_age", v4rgb(0x020206), v4rgb(0x040410), v4rgb(0x08081a), v4rgb(0x8010e0), v4rgb(0x101028), true},
	{"cosmic_age", v4rgb(0x010106), v4rgb(0x03030e), v4rgb(0x060616), v4rgb(0x9010f0), v4rgb(0x0c0c20), true},
	{"transcendent_age", v4rgb(0x010104), v4rgb(0x02020a), v4rgb(0x040410), v4rgb(0xa010ff), v4rgb(0x0a0a18), true},
	{"divine_age", v4rgb(0x010102), v4rgb(0x020208), v4rgb(0x03030c), v4rgb(0xb020ff), v4rgb(0x080810), true},
}

func v4GetPalette(ageKey string) v4AgePalette {
	for _, p := range v4Palettes {
		if p.ageKey == ageKey {
			return p
		}
	}
	return v4Palettes[0]
}

// ── Sprite pixel arrays ────────────────────────────────────────────────────

// food color constants
const (
	v4fG = 0xe8c840 // grain
	v4fS = 0x6a4a20 // stem
	v4fB = 0x4a3010 // base
)

// wood color constants
const (
	v4wD = 0x2d6b1a // dark green
	v4wM = 0x3d8b2a // mid green
	v4wT = 0x6b3a1a // trunk
	v4wR = 0x8b5a2b // root
)

// stone color constants
const (
	v4stL = 0xaaaaaa // light gray
	v4stG = 0x7a7a7a // mid gray
	v4stD = 0x4a4a4a // dark gray
	v4stH = 0xcccccc // highlight
)

// iron color constants
const (
	v4irI = 0x4a4a5a // dark iron
	v4irL = 0x7a7a8a // light iron
	v4irS = 0x2a2a3a // shadow
)

// knowledge color constants
const (
	v4kC = 0x8b4513 // brown cover
	v4kP = 0xf0f0e0 // page
	v4kL = 0xd0d0c0 // page shadow
	v4kG = 0xd4a017 // gold spine
)

// military color constants
const (
	v4miS = 0xc0c0c0 // silver sword
	v4miB = 0x8b1a1a // red shield
	v4miE = 0xd4a017 // gold emblem
	v4miW = 0xf0f0f0 // sword edge
)

// faith color constants
const (
	v4faW = 0xd0c8b0 // stone wall
	v4faD = 0x908070 // dark stone
	v4faG = 0xd4a017 // gold cross
	v4faR = 0x8b1a1a // red door
)

// culture color constants
const (
	v4cuH = 0xf5c080 // happy mask
	v4cuS = 0x8080c0 // sad mask
	v4cuG = 0xd4a017 // gold trim
	v4cuK = 0x303030 // dark outline
)

// trade color constants
const (
	v4trG = 0xd4a017 // gold
	v4trS = 0xc0c0c0 // silver pan
	v4trC = 0xe8c840 // coin
	v4trB = 0x6b3a1a // brown post
)

// energy color constants
const (
	v4enY = 0xffe040 // yellow bolt
	v4enO = 0xff8c00 // orange glow
	v4enW = 0xffffff // white core
	v4enG = 0x888888 // gray tower
)

// metallurgy color constants
const (
	v4meG = 0x888888 // gray
	v4meL = 0xbbbbbb // light
	v4meD = 0x555555 // dark
	v4meH = 0xdddddd // highlight
)

// advanced color constants
const (
	v4adC = 0x40c0ff // cyan
	v4adB = 0x0080c0 // blue
	v4adW = 0xffffff // white
	v4adD = 0x004080 // dark blue
	v4adP = 0x80e0ff // pale
)

// city_center color constants
const (
	v4ccW = 0xd0c8b0 // wall
	v4ccD = 0x908070 // dark stone
	v4ccG = 0xd4a017 // gold flag
	v4ccR = 0x8b1a1a // red flag
	v4ccB = 0x303030 // battlements
)

// wonder color constants
const (
	v4wnS = 0xc8a840 // sandstone
	v4wnL = 0xe8c860 // light face
	v4wnD = 0xa08030 // shadow
	v4wnG = 0xd4a017 // gold cap
)

// storage color constants
const (
	v4stW  = 0x8b5a2b // wood brown
	v4stLw = 0xc87941 // light wood
	v4stR  = 0x8b1a1a // red roof
	v4stDw = 0x5a3010 // dark wood
	v4stO  = 0xd4a017 // gold hinge
)

func v4SpriteFood() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, v4fG, 0, 0, v4fG, 0, 0, v4fG, 0, 0, 0, 0, 0, 0},
		{0, 0, v4fG, v4fG, v4fG, 0, v4fG, v4fG, 0, v4fG, v4fG, 0, 0, 0, 0, 0},
		{0, 0, v4fG, v4fG, v4fG, v4fG, v4fG, v4fG, v4fG, v4fG, v4fG, 0, 0, 0, 0, 0},
		{0, 0, 0, v4fG, v4fG, v4fG, v4fG, v4fG, v4fG, v4fG, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4fG, v4fG, v4fG, v4fG, v4fG, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4fS, 0, v4fS, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4fS, v4fS, v4fS, v4fS, v4fS, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4fS, v4fS, v4fS, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4fS, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4fS, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4fS, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4fS, v4fS, v4fS, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4fS, v4fS, v4fS, v4fS, v4fS, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, v4fB, v4fB, v4fB, v4fB, v4fB, v4fB, v4fB, 0, 0, 0, 0, 0, 0},
		{0, 0, v4fB, v4fB, v4fB, v4fB, v4fB, v4fB, v4fB, v4fB, v4fB, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4SpriteWood() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, v4wM, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4wM, v4wM, v4wM, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4wM, v4wM, v4wM, v4wM, v4wM, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4wD, v4wM, v4wM, v4wM, v4wM, v4wM, v4wD, 0, 0, 0, 0, 0},
		{0, 0, 0, v4wD, v4wD, v4wM, v4wM, v4wM, v4wM, v4wM, v4wD, v4wD, 0, 0, 0, 0},
		{0, 0, 0, 0, v4wM, v4wM, v4wM, v4wM, v4wM, v4wM, v4wM, v4wM, v4wM, 0, 0, 0},
		{0, 0, v4wM, v4wM, v4wM, v4wM, v4wM, v4wM, v4wM, v4wM, v4wM, v4wM, v4wM, v4wM, 0, 0},
		{0, v4wM, v4wM, v4wM, v4wM, v4wD, v4wM, v4wM, v4wM, v4wM, v4wD, v4wM, v4wM, v4wM, v4wM, 0},
		{v4wM, v4wM, v4wM, v4wM, v4wD, v4wD, v4wM, v4wM, v4wM, v4wM, v4wD, v4wD, v4wM, v4wM, v4wM, v4wM},
		{0, 0, 0, 0, 0, 0, 0, v4wT, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4wT, v4wT, v4wT, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4wT, v4wT, v4wT, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4wR, v4wT, v4wT, v4wT, v4wR, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4wR, v4wR, v4wT, v4wT, v4wT, v4wR, v4wR, 0, 0, 0, 0, 0},
		{0, 0, 0, v4wR, v4wR, v4wR, v4wR, v4wR, v4wR, v4wR, v4wR, v4wR, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4SpriteStone() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4stG, v4stG, v4stG, v4stG, v4stG, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4stG, v4stG, v4stH, v4stH, v4stL, v4stG, v4stG, 0, 0, 0, 0, 0},
		{0, 0, 0, v4stG, v4stG, v4stH, v4stH, v4stL, v4stL, v4stL, v4stG, v4stG, 0, 0, 0, 0},
		{0, 0, v4stG, v4stG, v4stH, v4stH, v4stL, v4stL, v4stL, v4stL, v4stL, v4stG, v4stG, 0, 0, 0},
		{0, 0, v4stG, v4stG, v4stH, v4stL, v4stL, v4stG, v4stG, v4stL, v4stL, v4stG, v4stG, 0, 0, 0},
		{0, 0, v4stG, v4stG, v4stL, v4stL, v4stG, v4stD, v4stD, v4stG, v4stL, v4stG, v4stG, 0, 0, 0},
		{0, v4stG, v4stG, v4stL, v4stL, v4stG, v4stD, v4stD, v4stD, v4stD, v4stG, v4stL, v4stG, v4stG, 0, 0},
		{0, v4stG, v4stG, v4stL, v4stG, v4stD, v4stD, v4stD, v4stD, v4stD, v4stD, v4stG, v4stL, v4stG, v4stG, 0},
		{v4stG, v4stG, v4stL, v4stG, v4stD, v4stD, v4stD, v4stG, v4stG, v4stD, v4stD, v4stD, v4stG, v4stL, v4stG, v4stG},
		{v4stG, v4stL, v4stG, v4stD, v4stD, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stD, v4stD, v4stG, v4stL, v4stG},
		{v4stG, v4stG, v4stD, v4stD, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stD, v4stD, v4stG, v4stG},
		{v4stG, v4stD, v4stD, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stD, v4stD, v4stG},
		{v4stD, v4stD, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stD, v4stD},
		{v4stD, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stG, v4stD},
		{v4stD, v4stD, v4stD, v4stD, v4stD, v4stD, v4stD, v4stD, v4stD, v4stD, v4stD, v4stD, v4stD, v4stD, v4stD, v4stD},
	}
}

func v4SpriteIron() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, 0, 0, 0, 0, 0},
		{0, 0, 0, v4irI, v4irI, v4irL, v4irL, v4irL, v4irL, v4irI, v4irI, v4irI, 0, 0, 0, 0},
		{0, 0, 0, v4irI, v4irI, v4irL, v4irL, v4irL, v4irL, v4irI, v4irI, v4irI, 0, 0, 0, 0},
		{0, 0, 0, 0, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4irI, v4irI, v4irI, v4irI, v4irI, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, 0, 0, 0, 0, 0},
		{0, 0, 0, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, 0, 0, 0, 0},
		{0, 0, v4irI, v4irI, v4irI, v4irL, v4irL, v4irL, v4irL, v4irI, v4irI, v4irI, v4irI, 0, 0, 0},
		{0, 0, v4irI, v4irI, v4irL, v4irL, v4irL, v4irL, v4irL, v4irL, v4irI, v4irI, v4irI, 0, 0, 0},
		{0, 0, v4irI, v4irI, v4irL, v4irL, v4irL, v4irL, v4irL, v4irL, v4irI, v4irI, v4irI, 0, 0, 0},
		{0, 0, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, 0, 0, 0},
		{0, 0, 0, v4irS, v4irI, v4irI, 0, 0, 0, v4irI, v4irI, v4irS, 0, 0, 0, 0},
		{0, 0, v4irS, v4irS, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irS, v4irS, 0, 0, 0},
		{0, v4irS, v4irS, v4irS, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irI, v4irS, v4irS, v4irS, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4SpriteKnowledge() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, v4kC, v4kC, v4kC, v4kC, v4kC, v4kG, v4kC, v4kC, v4kC, v4kC, v4kC, v4kC, 0, 0},
		{0, v4kC, v4kC, v4kP, v4kP, v4kP, v4kC, v4kG, v4kC, v4kP, v4kP, v4kP, v4kP, v4kC, v4kC, 0},
		{0, v4kC, v4kP, v4kP, v4kP, v4kP, v4kC, v4kG, v4kC, v4kP, v4kP, v4kP, v4kP, v4kP, v4kC, 0},
		{0, v4kC, v4kP, v4kP, v4kP, v4kP, v4kC, v4kG, v4kC, v4kP, v4kP, v4kP, v4kP, v4kP, v4kC, 0},
		{0, v4kC, v4kP, v4kP, v4kL, v4kP, v4kC, v4kG, v4kC, v4kP, v4kP, v4kL, v4kP, v4kP, v4kC, 0},
		{0, v4kC, v4kP, v4kP, v4kL, v4kP, v4kC, v4kG, v4kC, v4kP, v4kP, v4kL, v4kP, v4kP, v4kC, 0},
		{0, v4kC, v4kP, v4kP, v4kL, v4kP, v4kC, v4kG, v4kC, v4kP, v4kP, v4kL, v4kP, v4kP, v4kC, 0},
		{0, v4kC, v4kP, v4kP, v4kL, v4kP, v4kC, v4kG, v4kC, v4kP, v4kP, v4kL, v4kP, v4kP, v4kC, 0},
		{0, v4kC, v4kP, v4kP, v4kP, v4kP, v4kC, v4kG, v4kC, v4kP, v4kP, v4kP, v4kP, v4kP, v4kC, 0},
		{0, v4kC, v4kP, v4kP, v4kP, v4kP, v4kC, v4kG, v4kC, v4kP, v4kP, v4kP, v4kP, v4kP, v4kC, 0},
		{0, v4kC, v4kC, v4kC, v4kC, v4kC, v4kC, v4kG, v4kC, v4kC, v4kC, v4kC, v4kC, v4kC, v4kC, 0},
		{0, 0, v4kC, v4kC, v4kC, v4kC, v4kC, v4kG, v4kC, v4kC, v4kC, v4kC, v4kC, 0, 0, 0},
		{0, 0, 0, v4kC, v4kC, v4kC, v4kC, v4kG, v4kC, v4kC, v4kC, v4kC, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, v4kG, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4SpriteMilitary() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, v4miW, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4miW, v4miS, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4miW, v4miS, v4miS, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, v4miB, v4miB, v4miB, v4miW, v4miS, v4miS, v4miB, v4miB, v4miB, 0, 0, 0, 0, 0},
		{0, v4miB, v4miB, v4miB, v4miB, v4miB, v4miS, v4miS, v4miB, v4miB, v4miB, v4miB, 0, 0, 0, 0},
		{0, v4miB, v4miB, v4miB, v4miB, v4miB, v4miS, v4miS, v4miB, v4miB, v4miB, v4miB, 0, 0, 0, 0},
		{0, v4miB, v4miB, v4miB, v4miB, v4miE, v4miE, v4miE, v4miB, v4miB, v4miB, v4miB, 0, 0, 0, 0},
		{0, v4miB, v4miB, v4miB, v4miB, v4miE, v4miE, v4miE, v4miB, v4miB, v4miB, v4miB, 0, 0, 0, 0},
		{0, v4miB, v4miB, v4miB, v4miB, v4miE, v4miE, v4miE, v4miB, v4miB, v4miB, v4miB, 0, 0, 0, 0},
		{0, v4miB, v4miB, v4miB, v4miB, v4miB, v4miB, v4miB, v4miB, v4miB, v4miB, v4miB, 0, 0, 0, 0},
		{0, 0, v4miB, v4miB, v4miB, v4miB, v4miB, v4miB, v4miB, v4miB, v4miB, 0, 0, 0, 0, 0},
		{0, 0, 0, v4miB, v4miB, v4miB, v4miB, v4miB, v4miB, v4miB, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4miB, v4miB, v4miB, v4miB, v4miB, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4miB, v4miB, v4miB, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4miB, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4SpriteFaith() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, v4faG, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4faG, v4faG, v4faG, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4faG, v4faG, v4faG, v4faG, v4faG, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4faW, v4faG, v4faG, v4faG, v4faW, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4faW, v4faW, v4faG, v4faW, v4faG, v4faW, v4faW, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, 0, 0, 0, 0, 0},
		{0, 0, 0, v4faW, v4faW, v4faD, v4faW, v4faW, v4faW, v4faD, v4faW, v4faW, 0, 0, 0, 0},
		{0, 0, 0, v4faW, v4faW, v4faD, v4faW, v4faW, v4faW, v4faD, v4faW, v4faW, 0, 0, 0, 0},
		{0, 0, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, 0, 0, 0},
		{0, 0, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, 0, 0, 0},
		{0, 0, v4faW, v4faW, v4faD, v4faW, v4faW, v4faW, v4faW, v4faW, v4faD, v4faW, v4faW, 0, 0, 0},
		{0, 0, v4faW, v4faW, v4faD, v4faW, v4faW, v4faR, v4faR, v4faW, v4faD, v4faW, v4faW, 0, 0, 0},
		{0, 0, v4faW, v4faW, v4faW, v4faW, v4faW, v4faR, v4faR, v4faW, v4faW, v4faW, v4faW, 0, 0, 0},
		{0, 0, v4faW, v4faW, v4faW, v4faW, v4faW, v4faR, v4faR, v4faW, v4faW, v4faW, v4faW, 0, 0, 0},
		{0, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, 0, 0},
		{v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW, v4faW},
	}
}

func v4SpriteCulture() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, v4cuG, v4cuG, v4cuG, v4cuG, 0, 0, 0, 0, v4cuG, v4cuG, v4cuG, v4cuG, v4cuG, 0, 0},
		{0, v4cuG, v4cuH, v4cuH, v4cuG, 0, 0, 0, v4cuG, v4cuG, v4cuS, v4cuS, v4cuS, v4cuG, v4cuG, 0},
		{v4cuG, v4cuH, v4cuH, v4cuH, v4cuH, v4cuG, 0, v4cuG, v4cuS, v4cuS, v4cuS, v4cuS, v4cuS, v4cuS, v4cuG, 0},
		{v4cuG, v4cuH, v4cuK, 0, v4cuK, v4cuH, v4cuG, v4cuG, v4cuS, v4cuK, 0, 0, v4cuK, v4cuS, v4cuG, 0},
		{v4cuG, v4cuH, v4cuH, v4cuH, v4cuH, v4cuH, v4cuG, v4cuG, v4cuS, v4cuS, v4cuS, v4cuS, v4cuS, v4cuS, v4cuG, 0},
		{v4cuG, v4cuH, v4cuK, v4cuK, v4cuK, v4cuH, v4cuG, v4cuG, v4cuS, v4cuK, v4cuK, v4cuK, v4cuS, v4cuS, v4cuG, 0},
		{v4cuG, v4cuH, v4cuH, v4cuK, v4cuH, v4cuH, v4cuG, v4cuG, v4cuS, v4cuS, v4cuK, v4cuS, v4cuS, v4cuS, v4cuG, 0},
		{0, v4cuG, v4cuH, v4cuH, v4cuH, v4cuG, 0, v4cuG, v4cuS, v4cuS, v4cuS, v4cuS, v4cuS, v4cuG, v4cuG, 0},
		{0, 0, v4cuG, v4cuH, v4cuG, 0, 0, 0, v4cuG, v4cuS, v4cuS, v4cuS, v4cuG, v4cuG, 0, 0},
		{0, 0, v4cuG, v4cuG, v4cuG, 0, 0, 0, 0, v4cuG, v4cuG, v4cuG, v4cuG, 0, 0, 0},
		{0, 0, 0, v4cuG, 0, 0, 0, 0, 0, 0, v4cuG, 0, 0, 0, 0, 0},
		{0, 0, v4cuG, v4cuG, v4cuG, v4cuG, v4cuG, v4cuG, v4cuG, v4cuG, v4cuG, v4cuG, 0, 0, 0, 0},
		{0, 0, 0, v4cuG, v4cuG, v4cuG, v4cuG, v4cuG, v4cuG, v4cuG, v4cuG, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4SpriteTrade() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, v4trG, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, v4trG, v4trG, v4trG, v4trG, v4trG, v4trG, v4trG, v4trG, v4trG, v4trG, v4trG, v4trG, v4trG, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, v4trG, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, v4trG, 0, 0, 0, v4trG, v4trG, v4trG, 0, 0, 0, v4trG, 0, 0, 0},
		{0, v4trG, 0, v4trG, 0, v4trG, 0, v4trG, 0, v4trG, 0, v4trG, 0, v4trG, 0, 0},
		{v4trG, 0, 0, 0, v4trG, 0, 0, v4trG, 0, 0, v4trG, 0, 0, 0, v4trG, 0},
		{v4trS, v4trS, v4trS, v4trS, v4trS, v4trS, 0, v4trG, 0, v4trS, v4trS, v4trS, v4trS, v4trS, v4trS, 0},
		{v4trS, v4trC, v4trC, v4trC, v4trC, v4trS, 0, v4trB, 0, v4trS, v4trC, v4trC, v4trC, v4trC, v4trS, 0},
		{v4trS, v4trC, v4trC, v4trC, v4trC, v4trS, 0, v4trB, 0, v4trS, v4trC, v4trC, v4trC, v4trC, v4trS, 0},
		{v4trS, v4trS, v4trS, v4trS, v4trS, v4trS, 0, v4trB, 0, v4trS, v4trS, v4trS, v4trS, v4trS, v4trS, 0},
		{0, 0, 0, 0, 0, 0, 0, v4trB, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, v4trB, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4trB, v4trB, v4trB, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4trB, v4trB, v4trB, v4trB, v4trB, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4trB, v4trB, v4trB, v4trB, v4trB, v4trB, v4trB, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4SpriteEnergy() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, v4enW, v4enW, v4enW, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4enW, v4enY, v4enY, v4enO, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4enW, v4enY, v4enY, v4enY, v4enO, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, v4enW, v4enY, v4enY, v4enY, v4enO, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, v4enW, v4enY, v4enY, v4enY, v4enY, v4enY, v4enY, v4enO, 0, 0, 0, 0, 0, 0},
		{0, v4enW, v4enY, v4enY, v4enY, v4enY, v4enY, v4enY, v4enY, v4enY, v4enO, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4enW, v4enY, v4enY, v4enY, v4enO, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4enW, v4enY, v4enO, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4enW, v4enO, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4enW, v4enY, v4enO, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4enW, v4enY, v4enY, v4enY, v4enO, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, v4enG, v4enG, v4enG, v4enG, v4enG, v4enG, v4enG, v4enG, v4enG, v4enG, 0, 0, 0, 0},
		{0, 0, 0, v4enG, v4enG, 0, 0, 0, 0, v4enG, v4enG, 0, 0, 0, 0, 0},
		{0, 0, 0, v4enG, v4enG, 0, 0, 0, 0, v4enG, v4enG, 0, 0, 0, 0, 0},
		{0, 0, v4enG, v4enG, v4enG, v4enG, 0, 0, v4enG, v4enG, v4enG, v4enG, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4SpriteMetallurgy() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, v4meD, v4meD, v4meD, v4meD, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4meD, v4meG, v4meG, v4meG, v4meG, v4meD, 0, 0, 0, 0, 0, 0},
		{0, 0, v4meD, v4meD, v4meG, v4meL, v4meH, v4meL, v4meG, v4meG, v4meD, v4meD, 0, 0, 0, 0},
		{0, v4meD, v4meG, v4meG, v4meG, v4meH, v4meH, v4meL, v4meG, v4meG, v4meG, v4meG, v4meD, 0, 0, 0},
		{v4meD, v4meG, v4meG, v4meL, v4meH, v4meH, v4meH, v4meH, v4meH, v4meL, v4meG, v4meG, v4meG, v4meD, 0, 0},
		{v4meD, v4meG, v4meL, v4meH, v4meH, v4meH, v4meH, v4meH, v4meH, v4meH, v4meL, v4meG, v4meG, v4meD, 0, 0},
		{v4meD, v4meG, v4meG, v4meH, v4meH, v4meG, v4meG, v4meG, v4meG, v4meH, v4meH, v4meG, v4meG, v4meD, 0, 0},
		{v4meD, v4meG, v4meL, v4meH, v4meH, v4meG, v4meD, v4meD, v4meG, v4meH, v4meH, v4meL, v4meG, v4meD, 0, 0},
		{v4meD, v4meG, v4meL, v4meH, v4meH, v4meG, v4meD, v4meD, v4meG, v4meH, v4meH, v4meL, v4meG, v4meD, 0, 0},
		{v4meD, v4meG, v4meG, v4meH, v4meH, v4meG, v4meG, v4meG, v4meG, v4meH, v4meH, v4meG, v4meG, v4meD, 0, 0},
		{v4meD, v4meG, v4meL, v4meH, v4meH, v4meH, v4meH, v4meH, v4meH, v4meH, v4meL, v4meG, v4meG, v4meD, 0, 0},
		{v4meD, v4meG, v4meG, v4meL, v4meH, v4meH, v4meH, v4meH, v4meH, v4meL, v4meG, v4meG, v4meG, v4meD, 0, 0},
		{0, v4meD, v4meG, v4meG, v4meG, v4meH, v4meH, v4meL, v4meG, v4meG, v4meG, v4meG, v4meD, 0, 0, 0},
		{0, 0, v4meD, v4meD, v4meG, v4meL, v4meH, v4meL, v4meG, v4meG, v4meD, v4meD, 0, 0, 0, 0},
		{0, 0, 0, 0, v4meD, v4meG, v4meG, v4meG, v4meG, v4meD, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4meD, v4meD, v4meD, v4meD, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4SpriteAdvanced() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, v4adW, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4adW, v4adC, v4adW, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4adW, v4adC, v4adC, v4adC, v4adW, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4adW, v4adP, v4adC, v4adW, v4adC, v4adC, v4adW, 0, 0, 0, 0, 0},
		{0, 0, 0, v4adW, v4adP, v4adP, v4adC, v4adW, v4adC, v4adC, v4adC, v4adW, 0, 0, 0, 0},
		{0, 0, v4adW, v4adP, v4adP, v4adB, v4adC, v4adW, v4adC, v4adB, v4adC, v4adC, v4adW, 0, 0, 0},
		{0, v4adW, v4adP, v4adP, v4adB, v4adB, v4adC, v4adW, v4adC, v4adB, v4adB, v4adC, v4adW, 0, 0, 0},
		{v4adW, v4adP, v4adP, v4adB, v4adB, v4adD, v4adC, v4adW, v4adC, v4adD, v4adB, v4adB, v4adC, v4adW, 0, 0},
		{0, v4adW, v4adP, v4adP, v4adB, v4adB, v4adC, v4adW, v4adC, v4adB, v4adB, v4adC, v4adW, 0, 0, 0},
		{0, 0, v4adW, v4adP, v4adP, v4adB, v4adC, v4adW, v4adC, v4adB, v4adC, v4adW, 0, 0, 0, 0},
		{0, 0, 0, v4adW, v4adP, v4adP, v4adC, v4adW, v4adC, v4adP, v4adW, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4adW, v4adP, v4adC, v4adW, v4adC, v4adW, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4adW, v4adC, v4adW, v4adC, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4adW, v4adC, v4adW, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, v4adW, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4SpriteCityCenter() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, v4ccR, v4ccR, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, v4ccG, v4ccR, v4ccR, v4ccG, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, v4ccG, v4ccG, v4ccG, v4ccG, v4ccG, v4ccG, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{v4ccB, v4ccW, v4ccB, v4ccW, v4ccB, v4ccW, v4ccB, v4ccW, v4ccB, v4ccW, v4ccB, v4ccW, v4ccB, v4ccW, v4ccB, v4ccW},
		{v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW},
		{v4ccW, v4ccW, v4ccD, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccD, v4ccW, v4ccW, v4ccW, v4ccW},
		{v4ccW, v4ccW, v4ccD, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccD, v4ccW, v4ccW, v4ccW, v4ccW},
		{v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW},
		{v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW},
		{v4ccW, v4ccW, v4ccD, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccD, v4ccW, v4ccW, v4ccW, v4ccW},
		{v4ccW, v4ccW, v4ccD, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccD, v4ccW, v4ccW, v4ccW, v4ccW},
		{v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW},
		{v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW},
		{v4ccW, v4ccW, v4ccD, v4ccW, v4ccR, v4ccR, v4ccW, v4ccW, v4ccW, v4ccR, v4ccR, v4ccD, v4ccW, v4ccW, v4ccW, v4ccW},
		{v4ccW, v4ccW, v4ccD, v4ccW, v4ccR, v4ccR, v4ccW, v4ccW, v4ccW, v4ccR, v4ccR, v4ccD, v4ccW, v4ccW, v4ccW, v4ccW},
		{v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW, v4ccW},
	}
}

// v4PalacePixels returns a 16×16 palace sprite for the given age key.
// Each of the 22 ages has a unique design drawn using sprites.Canvas helpers.
func v4PalacePixels(ageKey string) [16][16]uint32 {
	switch ageKey {
	case "primitive_age":
		// Chief's roundhouse — circular clay hut with thatched roof and fire pit
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x3A2E1A) // muddy ground
		// Circular wall using octagonal approximation
		c.FillRect(4, 2, 8, 12, 0x7A5A38) // main body
		c.FillRect(2, 4, 12, 8, 0x7A5A38) // widen middle
		// Thatched roof (darker brown cone suggestion)
		c.FillRect(5, 2, 6, 2, 0x5A4020)
		c.FillRect(4, 3, 8, 2, 0x5A4020)
		c.Hline(3, 4, 10, 0x6A5030)
		// Wall shading — right/bottom slightly darker
		c.Vline(13, 4, 8, 0x5A4228)
		c.Hline(4, 13, 8, 0x5A4228)
		// Central fire pit
		c.Set(7, 7, 0xFF6020)
		c.Set(8, 7, 0xFF8040)
		c.Set(7, 8, 0xCC4010)
		c.Set(8, 8, 0xFF6020)
		// Door gap (south face)
		c.Set(7, 13, 0x2A1E0A)
		c.Set(8, 13, 0x2A1E0A)
		return c.Pixels

	case "stone_age":
		// Stone circle — Stonehenge-style ring of standing stones
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x202020)
		// Ring of standing stones around center (8,8)
		stones := [][2]int{
			{8, 1}, {12, 3}, {14, 7}, {13, 12}, {8, 14}, {3, 12}, {2, 7}, {4, 3},
			{10, 2}, {14, 10}, {6, 14}, {2, 10},
		}
		for _, s := range stones {
			c.FillRect(s[0]-1, s[1]-1, 2, 3, 0xA09080)
		}
		// Central altar stone
		c.FillRect(6, 6, 4, 4, 0xC0B090)
		c.Set(7, 7, 0xD8C8A8)
		c.Set(8, 7, 0xD8C8A8)
		return c.Pixels

	case "bronze_age":
		// Step pyramid / ziggurat — 4 concentric squares
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x303020)
		c.FillRect(1, 8, 14, 7, 0xC8943C)  // base level
		c.FillRect(3, 6, 10, 5, 0xDBA84A)  // level 2
		c.FillRect(5, 4, 6, 4, 0xEFBE5A)   // level 3
		c.FillRect(6, 2, 4, 4, 0xFFD870)   // top level / summit
		c.Set(7, 1, 0xFFE890)              // capstone
		c.Set(8, 1, 0xFFE890)
		// Shadow edges
		c.Vline(1, 8, 7, 0x906830)
		c.Vline(14, 8, 7, 0x906830)
		c.Hline(1, 14, 14, 0x705028)
		c.Vline(3, 6, 5, 0xAA7830)
		c.Vline(12, 6, 5, 0xAA7830)
		c.Vline(5, 4, 4, 0xC09040)
		c.Vline(10, 4, 4, 0xC09040)
		return c.Pixels

	case "iron_age":
		// Hillfort / motte-and-bailey — earthwork ring with keep
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x506030)  // grass
		// Outer earthwork ring
		c.FillRect(1, 1, 14, 14, 0x6B5030)
		c.FillRect(2, 2, 12, 12, 0x8B6B40)  // inner courtyard
		// Ditch around outer ring
		c.Hline(1, 1, 14, 0x4A3820)
		c.Hline(1, 14, 14, 0x4A3820)
		c.Vline(1, 1, 14, 0x4A3820)
		c.Vline(14, 1, 14, 0x4A3820)
		// Inner keep tower
		c.FillRect(5, 5, 6, 6, 0x505050)
		c.FillRect(6, 6, 4, 4, 0x686868)
		// Battlements on keep
		c.Set(5, 4, 0x505050)
		c.Set(7, 4, 0x505050)
		c.Set(9, 4, 0x505050)
		c.Set(11, 4, 0x505050)
		// Gate opening south
		c.FillRect(7, 10, 2, 1, 0x303030)
		return c.Pixels

	case "classical_age":
		// Greek/Roman temple — column peristyle with cella
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0xF0ECD8)  // marble floor
		// Outer colonnade columns
		for col := 1; col <= 14; col += 2 {
			c.Set(col, 1, 0xD0C8A8)
			c.Set(col, 2, 0xD0C8A8)
		}
		for col := 1; col <= 14; col += 2 {
			c.Set(col, 12, 0xD0C8A8)
			c.Set(col, 13, 0xD0C8A8)
		}
		for row := 1; row <= 13; row += 2 {
			c.Set(1, row, 0xD0C8A8)
			c.Set(14, row, 0xD0C8A8)
		}
		// Inner cella
		c.FillRect(4, 4, 8, 8, 0xE0D8C0)
		c.Hline(4, 4, 8, 0xB8B098)
		c.Hline(4, 11, 8, 0xB8B098)
		c.Vline(4, 4, 8, 0xB8B098)
		c.Vline(11, 4, 8, 0xB8B098)
		// Gold roof trim
		c.Hline(0, 0, 16, 0xD4A820)
		c.Hline(0, 15, 16, 0xD4A820)
		// Doorway
		c.FillRect(7, 9, 2, 3, 0xA09070)
		return c.Pixels

	case "medieval_age":
		// Gothic cathedral — cross-shaped floor plan
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x706860)  // stone ground
		// Cross nave: vertical (north-south)
		c.FillRect(5, 0, 6, 16, 0xC0B8A0)
		// Cross transept: horizontal (east-west)
		c.FillRect(0, 5, 16, 6, 0xC0B8A0)
		// Corner tower squares
		c.FillRect(0, 0, 4, 4, 0x908070)
		c.FillRect(12, 0, 4, 4, 0x908070)
		c.FillRect(0, 12, 4, 4, 0x908070)
		c.FillRect(12, 12, 4, 4, 0x908070)
		// Rose window: gold dot at crossing center
		c.Set(7, 7, 0xD4A820)
		c.Set(8, 7, 0xD4A820)
		c.Set(7, 8, 0xD4A820)
		c.Set(8, 8, 0xD4A820)
		// Stone vaulting lines
		c.Hline(5, 7, 6, 0xA09880)
		c.Vline(7, 5, 6, 0xA09880)
		// Tower details
		c.Set(1, 1, 0xA09880)
		c.Set(2, 1, 0xA09880)
		c.Set(13, 1, 0xA09880)
		c.Set(14, 1, 0xA09880)
		return c.Pixels

	case "renaissance_age":
		// Baroque palace — U-shape with inner courtyard and dome
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x90C060)  // garden courtyard
		// U-shaped palace wings
		c.FillRect(0, 0, 4, 14, 0xE8D498)   // left wing
		c.FillRect(12, 0, 4, 14, 0xE8D498)  // right wing
		c.FillRect(0, 0, 16, 5, 0xE8D498)   // top connecting wing
		// Central dome at top center
		c.Dome(7, 4, 3, 0xCCA030)
		c.Set(7, 1, 0xFFD860)
		c.Set(8, 1, 0xFFD860)
		// Windows on wings
		c.Set(1, 6, 0xC0A870)
		c.Set(1, 9, 0xC0A870)
		c.Set(14, 6, 0xC0A870)
		c.Set(14, 9, 0xC0A870)
		// Courtyard garden path
		c.Hline(4, 10, 8, 0xC8B878)
		c.Vline(7, 5, 9, 0xC8B878)
		c.Vline(8, 5, 9, 0xC8B878)
		return c.Pixels

	case "age_of_sail":
		// Colonial harbour master's hall — brick wings, central clock tower
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x507040)
		// Main brick building
		c.FillRect(1, 4, 14, 10, 0xA06840)
		// White central facade
		c.FillRect(5, 3, 6, 11, 0xF0E8D0)
		// Clock tower
		c.FillRect(6, 0, 4, 5, 0xC89858)
		c.Set(7, 1, 0xF0F0F0)
		c.Set(8, 1, 0xF0F0F0)
		// Columns
		c.Vline(5, 3, 11, 0xD8C898)
		c.Vline(10, 3, 11, 0xD8C898)
		// Windows in wings
		c.Set(2, 6, 0x80A0B8)
		c.Set(2, 9, 0x80A0B8)
		c.Set(13, 6, 0x80A0B8)
		c.Set(13, 9, 0x80A0B8)
		// Door
		c.FillRect(7, 11, 2, 3, 0x603820)
		// Ground path
		c.Hline(5, 14, 6, 0xB0A870)
		return c.Pixels

	case "industrial_age":
		// Victorian city hall + clock tower
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x484840)
		// Main brick building
		c.FillRect(1, 5, 14, 10, 0xA84030)
		// Central clock tower rising from center
		c.FillRect(6, 0, 4, 7, 0x806050)
		// Clock face
		c.Set(7, 2, 0xF0F0F0)
		c.Set(8, 2, 0xF0F0F0)
		c.Set(7, 3, 0xF0F0F0)
		c.Set(8, 3, 0xF0F0F0)
		// Windows
		c.FillRect(2, 7, 2, 3, 0x8090A0)
		c.FillRect(12, 7, 2, 3, 0x8090A0)
		c.FillRect(5, 7, 2, 3, 0x8090A0)
		c.FillRect(9, 7, 2, 3, 0x8090A0)
		// Smokestack at corner
		c.FillRect(13, 2, 2, 5, 0x404040)
		c.Set(13, 1, 0x606060)
		c.Set(14, 1, 0x606060)
		// Foundation
		c.Hline(1, 14, 14, 0x703020)
		return c.Pixels

	case "gilded_age":
		// Gilded era grand opera house — ornate symmetrical facade
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x504840)
		// Main building
		c.FillRect(1, 4, 14, 11, 0xD4B880)
		// Grand dome at center
		c.Dome(7, 5, 4, 0xD4A820)
		c.Set(7, 1, 0xFFD040)
		c.Set(8, 1, 0xFFD040)
		// Colonnade
		for col := 2; col <= 13; col += 2 {
			c.Vline(col, 4, 3, 0xC8A870)
		}
		// Arched windows
		c.Set(3, 8, 0x90C0D8)
		c.Set(3, 9, 0x90C0D8)
		c.Set(12, 8, 0x90C0D8)
		c.Set(12, 9, 0x90C0D8)
		c.Set(6, 8, 0x90C0D8)
		c.Set(6, 9, 0x90C0D8)
		c.Set(9, 8, 0x90C0D8)
		c.Set(9, 9, 0x90C0D8)
		// Grand entrance steps
		c.Hline(5, 14, 6, 0xB89860)
		c.Hline(6, 15, 4, 0xB89860)
		return c.Pixels

	case "electric_age":
		// Art Deco civic tower — stepped pyramid with gold spire
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x303028)
		// Setback levels (bottom to top, wider to narrower)
		c.FillRect(1, 12, 14, 3, 0xF0E8C8)
		c.FillRect(2, 9, 12, 4, 0xE8E0C0)
		c.FillRect(3, 6, 10, 4, 0xE0D8B8)
		c.FillRect(4, 3, 8, 4, 0xD8D0B0)
		c.FillRect(5, 1, 6, 3, 0xD4A020)   // gold upper level
		c.FillRect(7, 0, 2, 2, 0xFFD040)   // gold spire
		// Geometric ornament lines
		c.Hline(1, 11, 14, 0xC8B890)
		c.Hline(2, 8, 12, 0xC8B890)
		c.Hline(3, 5, 10, 0xC8B890)
		c.Hline(4, 2, 8, 0xD4A820)
		// Vertical ornament
		c.Vline(1, 12, 3, 0xC0B888)
		c.Vline(14, 12, 3, 0xC0B888)
		return c.Pixels

	case "modern_age":
		// Glass skyscraper complex — dark glass, towers, reflecting pool
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x203038)
		// Central tall tower
		c.FillRect(5, 0, 6, 13, 0x304858)
		// Flanking towers
		c.FillRect(0, 3, 4, 10, 0x283848)
		c.FillRect(12, 3, 4, 10, 0x283848)
		// Blue glass highlight strips on main tower
		c.Vline(5, 0, 13, 0x60A0C0)
		c.Vline(10, 0, 13, 0x60A0C0)
		c.Hline(5, 4, 6, 0x4888A8)
		c.Hline(5, 8, 6, 0x4888A8)
		// Glass highlights on side towers
		c.Vline(0, 3, 10, 0x4880A0)
		c.Vline(15, 3, 10, 0x4880A0)
		// Reflecting pool at base
		c.FillRect(3, 13, 10, 3, 0x4080A0)
		c.Hline(4, 13, 8, 0x60B0C8)
		return c.Pixels

	case "atomic_age":
		// Modernist government building — horizontal slab on piloti
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0xB0A890)  // plaza
		// Piloti columns
		for col := 2; col <= 13; col += 3 {
			c.Vline(col, 8, 4, 0x908878)
		}
		// Horizontal slab
		c.FillRect(0, 4, 16, 5, 0xD0C8B8)
		// Glass curtain wall strip
		c.FillRect(0, 5, 16, 3, 0x8090A0)
		c.Hline(0, 5, 16, 0xA0B0C0)
		c.Hline(0, 7, 16, 0xA0B0C0)
		// Vertical mullions
		for col := 2; col <= 14; col += 2 {
			c.Vline(col, 5, 3, 0x708090)
		}
		// Flat roof
		c.Hline(0, 4, 16, 0xC0B8A8)
		// Entry overhang
		c.FillRect(5, 12, 6, 2, 0xC0B8A8)
		return c.Pixels

	case "information_age":
		// Tech campus — campus ring road, central hub, data center blocks
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0xD8D8D0)
		// Campus ring road
		c.Hline(1, 2, 14, 0xB0A8A0)
		c.Hline(1, 13, 14, 0xB0A8A0)
		c.Vline(1, 2, 12, 0xB0A8A0)
		c.Vline(14, 2, 12, 0xB0A8A0)
		// Central hub building
		c.FillRect(5, 5, 6, 6, 0x6090B0)
		c.FillRect(6, 6, 4, 4, 0x80B0C8)
		c.Set(7, 7, 0xC0E0F0)
		c.Set(8, 7, 0xC0E0F0)
		// Data center blocks in corners
		c.FillRect(2, 3, 2, 2, 0x506070)
		c.FillRect(12, 3, 2, 2, 0x506070)
		c.FillRect(2, 11, 2, 2, 0x506070)
		c.FillRect(12, 11, 2, 2, 0x506070)
		// Roads to hub
		c.Hline(3, 7, 2, 0xC0B8B0)
		c.Hline(11, 7, 2, 0xC0B8B0)
		c.Vline(7, 3, 2, 0xC0B8B0)
		c.Vline(7, 11, 2, 0xC0B8B0)
		return c.Pixels

	case "digital_age":
		// Server farm / data citadel — grid of server blocks with LED strips
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x303038)
		// Grid of server block rows
		for row := 1; row <= 13; row += 3 {
			c.FillRect(1, row, 14, 2, 0x202028)
			c.Hline(1, row+2, 14, 0x2060FF)  // blue LED strip between rows
		}
		// Server rack columns
		for col := 2; col <= 13; col += 3 {
			c.Vline(col, 1, 13, 0x282830)
		}
		// Central node
		c.FillRect(6, 6, 4, 4, 0x183048)
		c.Set(7, 7, 0x40C0FF)
		c.Set(8, 7, 0x40C0FF)
		c.Set(7, 8, 0x40C0FF)
		c.Set(8, 8, 0x40C0FF)
		// Status LEDs along edge
		for row := 1; row <= 13; row += 3 {
			c.Set(1, row, 0x00FF80)
			c.Set(14, row, 0x00FF80)
		}
		return c.Pixels

	case "cyberpunk_age":
		// Neon megaplex tower — near-black with neon stripes
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x101018)
		// Central tower block
		c.FillRect(4, 0, 8, 14, 0x181820)
		// Neon stripe running up center
		c.Vline(7, 0, 14, 0xFF2060)
		c.Vline(8, 0, 14, 0xFF2060)
		// Neon cyan horizontal bands
		c.Hline(4, 3, 8, 0x00FFEE)
		c.Hline(4, 7, 8, 0x00FFEE)
		c.Hline(4, 11, 8, 0x00FFEE)
		// Side buildings
		c.FillRect(0, 3, 3, 10, 0x141420)
		c.FillRect(13, 3, 3, 10, 0x141420)
		// Purple glow corner accents
		c.Set(0, 3, 0x8020FF)
		c.Set(2, 3, 0x8020FF)
		c.Set(13, 3, 0x8020FF)
		c.Set(15, 3, 0x8020FF)
		c.Set(0, 12, 0x8020FF)
		c.Set(2, 12, 0x8020FF)
		c.Set(13, 12, 0x8020FF)
		c.Set(15, 12, 0x8020FF)
		// Neon base glow
		c.Hline(4, 13, 8, 0xFF2060)
		c.FillRect(2, 14, 12, 2, 0x0A0810)
		c.Hline(3, 14, 10, 0x8020FF)
		return c.Pixels

	case "nanotech_age":
		// Nanotech research nexus — molecular lattice structure
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x080810)
		// Hexagonal lattice suggestion
		c.Hline(3, 4, 10, 0x2040A0)
		c.Hline(3, 8, 10, 0x2040A0)
		c.Hline(3, 12, 10, 0x2040A0)
		c.Vline(3, 4, 9, 0x2040A0)
		c.Vline(8, 4, 9, 0x2040A0)
		c.Vline(13, 4, 9, 0x2040A0)
		// Node points
		for _, pt := range [][2]int{{3, 4}, {8, 4}, {13, 4}, {3, 8}, {8, 8}, {13, 8}, {3, 12}, {8, 12}, {13, 12}} {
			c.Set(pt[0], pt[1], 0x60C0FF)
		}
		// Central research core
		c.FillRect(5, 5, 6, 6, 0x101830)
		c.Set(7, 7, 0x80FFFF)
		c.Set(8, 7, 0x80FFFF)
		c.Set(7, 8, 0x80FFFF)
		c.Set(8, 8, 0x80FFFF)
		// Diagonal struts
		c.Set(5, 5, 0x4060C0)
		c.Set(10, 5, 0x4060C0)
		c.Set(5, 10, 0x4060C0)
		c.Set(10, 10, 0x4060C0)
		return c.Pixels

	case "fusion_age":
		// Fusion reactor command center — circular reactor ring with plasma
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x0A0A20)
		// Outer reactor housing ring
		c.Hline(3, 2, 10, 0x2040FF)
		c.Hline(3, 13, 10, 0x2040FF)
		c.Vline(2, 3, 10, 0x2040FF)
		c.Vline(13, 3, 10, 0x2040FF)
		c.Set(3, 3, 0x2040FF)
		c.Set(12, 3, 0x2040FF)
		c.Set(3, 12, 0x2040FF)
		c.Set(12, 12, 0x2040FF)
		// Inner reactor ring
		c.Hline(5, 4, 6, 0x4070FF)
		c.Hline(5, 11, 6, 0x4070FF)
		c.Vline(4, 5, 6, 0x4070FF)
		c.Vline(11, 5, 6, 0x4070FF)
		// Energy core
		c.FillRect(6, 6, 4, 4, 0x102040)
		c.Set(7, 7, 0x80C0FF)
		c.Set(8, 7, 0x80C0FF)
		c.Set(7, 8, 0x80C0FF)
		c.Set(8, 8, 0x80C0FF)
		// Plasma conduit lines radiating out
		c.Vline(7, 0, 3, 0x4080FF)
		c.Vline(8, 0, 3, 0x4080FF)
		c.Vline(7, 13, 3, 0x4080FF)
		c.Vline(8, 13, 3, 0x4080FF)
		c.Hline(0, 7, 3, 0x4080FF)
		c.Hline(0, 8, 3, 0x4080FF)
		c.Hline(13, 7, 3, 0x4080FF)
		c.Hline(13, 8, 3, 0x4080FF)
		return c.Pixels

	case "space_age":
		// Launch command complex — launch pad with rocket
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x808080)  // concrete
		// Launch pad rectangle
		c.FillRect(4, 8, 8, 6, 0x606060)
		// Blast deflector at base
		c.FillRect(3, 12, 10, 2, 0x404040)
		c.Hline(3, 11, 10, 0x505050)
		// Rocket silhouette on pad
		c.FillRect(7, 1, 2, 9, 0xD0D0D0)
		// Nose cone
		c.Set(7, 0, 0xE0E0E0)
		c.Set(8, 0, 0xE0E0E0)
		// Fins
		c.Set(5, 8, 0xB0B0B0)
		c.Set(6, 8, 0xB0B0B0)
		c.Set(9, 8, 0xB0B0B0)
		c.Set(10, 8, 0xB0B0B0)
		// Flame
		c.Set(7, 10, 0xFF8020)
		c.Set(8, 10, 0xFF8020)
		c.Set(7, 11, 0xFF4000)
		c.Set(8, 11, 0xFF4000)
		// Service towers
		c.Vline(2, 3, 10, 0x707070)
		c.Vline(13, 3, 10, 0x707070)
		c.Hline(2, 5, 5, 0x707070)
		c.Hline(9, 5, 5, 0x707070)
		return c.Pixels

	case "interstellar_age":
		// Deep space station — ring station with solar panels
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x080808)  // void
		// Outer ring
		c.Hline(3, 2, 10, 0x2050A0)
		c.Hline(3, 13, 10, 0x2050A0)
		c.Vline(2, 3, 10, 0x2050A0)
		c.Vline(13, 3, 10, 0x2050A0)
		c.Set(3, 3, 0x2050A0)
		c.Set(12, 3, 0x2050A0)
		c.Set(3, 12, 0x2050A0)
		c.Set(12, 12, 0x2050A0)
		// Inner ring
		c.Hline(5, 4, 6, 0x4080D0)
		c.Hline(5, 11, 6, 0x4080D0)
		c.Vline(4, 5, 6, 0x4080D0)
		c.Vline(11, 5, 6, 0x4080D0)
		// Hub at center
		c.FillRect(6, 6, 4, 4, 0x203060)
		c.Set(7, 7, 0x80C0FF)
		c.Set(8, 8, 0x80C0FF)
		// Solar panel arrays extending from ring
		c.FillRect(0, 6, 2, 4, 0xFFCC00)
		c.FillRect(14, 6, 2, 4, 0xFFCC00)
		c.FillRect(6, 0, 4, 2, 0xFFCC00)
		c.FillRect(6, 14, 4, 2, 0xFFCC00)
		// Panel dividers
		c.Vline(1, 6, 4, 0xCC9900)
		c.Vline(14, 6, 4, 0xCC9900)
		c.Hline(6, 1, 4, 0xCC9900)
		c.Hline(6, 14, 4, 0xCC9900)
		return c.Pixels

	case "galactic_age":
		// Alien crystalline palace — star pattern of crystal spires
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x050510)
		// Teal crystal spires in star pattern
		// Center spire
		c.Vline(7, 3, 10, 0x20D0A0)
		c.Vline(8, 3, 10, 0x20D0A0)
		c.Set(7, 2, 0x40FFC0)
		c.Set(8, 2, 0x40FFC0)
		// Cross spires
		c.Hline(3, 7, 10, 0x20D0A0)
		c.Hline(3, 8, 10, 0x20D0A0)
		// Diagonal crystal lines
		c.Set(4, 4, 0x20D0A0)
		c.Set(5, 5, 0x20D0A0)
		c.Set(10, 4, 0x20D0A0)
		c.Set(11, 5, 0x20D0A0)
		c.Set(4, 11, 0x20D0A0)
		c.Set(5, 10, 0x20D0A0)
		c.Set(10, 11, 0x20D0A0)
		c.Set(11, 10, 0x20D0A0)
		// Violet energy core
		c.FillRect(6, 6, 4, 4, 0x100820)
		c.Set(7, 7, 0xA020FF)
		c.Set(8, 7, 0xA020FF)
		c.Set(7, 8, 0xA020FF)
		c.Set(8, 8, 0xA020FF)
		// Bioluminescent accents
		c.Set(3, 3, 0x40FFA0)
		c.Set(12, 3, 0x40FFA0)
		c.Set(3, 12, 0x40FFA0)
		c.Set(12, 12, 0x40FFA0)
		return c.Pixels

	case "quantum_age":
		// Quantum probability nexus — concentric dotted rings, bright core
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x020208)
		// Outer uncertainty halo — alternating dots
		outerPts := [][2]int{
			{7, 0}, {9, 0}, {12, 1}, {14, 3}, {15, 5}, {15, 8}, {15, 11},
			{14, 13}, {12, 14}, {9, 15}, {7, 15}, {4, 14}, {2, 13}, {1, 11},
			{0, 8}, {0, 5}, {1, 3}, {4, 1},
		}
		for i, pt := range outerPts {
			if i%2 == 0 {
				c.Set(pt[0], pt[1], 0x8040FF)
			} else {
				c.Set(pt[0], pt[1], 0xFF40C0)
			}
		}
		// Inner halo
		innerPts := [][2]int{
			{7, 3}, {9, 3}, {11, 5}, {12, 7}, {12, 9}, {11, 11},
			{9, 12}, {7, 12}, {5, 11}, {4, 9}, {4, 7}, {5, 5},
		}
		for i, pt := range innerPts {
			if i%2 == 0 {
				c.Set(pt[0], pt[1], 0xFF40C0)
			} else {
				c.Set(pt[0], pt[1], 0x8040FF)
			}
		}
		// Phase lines
		c.Hline(4, 7, 8, 0x4040A0)
		c.Vline(7, 4, 8, 0x4040A0)
		// Quantum core — bright white center
		c.FillRect(6, 6, 4, 4, 0x080820)
		c.Set(7, 7, 0xFFFFFF)
		c.Set(8, 7, 0xFFFFFF)
		c.Set(7, 8, 0xFFFFFF)
		c.Set(8, 8, 0xFFFFFF)
		return c.Pixels

	case "singularity_age":
		// Singularity convergence point — spiraling light collapse
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x020206)
		// Outer spiral suggestion
		c.Hline(2, 2, 12, 0x6020C0)
		c.Vline(13, 2, 12, 0x6020C0)
		c.Hline(2, 13, 12, 0x6020C0)
		c.Vline(2, 2, 12, 0x6020C0)
		// Mid ring
		c.Hline(4, 4, 8, 0x8040E0)
		c.Vline(11, 4, 8, 0x8040E0)
		c.Hline(4, 11, 8, 0x8040E0)
		c.Vline(4, 4, 8, 0x8040E0)
		// Inner collapse
		c.FillRect(6, 6, 4, 4, 0x200440)
		c.Set(7, 7, 0xFFFFFF)
		c.Set(8, 7, 0xFFFF80)
		c.Set(7, 8, 0xFFFF80)
		c.Set(8, 8, 0xFFFFFF)
		// Radiant beams
		c.Set(7, 0, 0xC060FF)
		c.Set(8, 0, 0xC060FF)
		c.Set(0, 7, 0xC060FF)
		c.Set(0, 8, 0xC060FF)
		c.Set(15, 7, 0xC060FF)
		c.Set(15, 8, 0xC060FF)
		c.Set(7, 15, 0xC060FF)
		c.Set(8, 15, 0xC060FF)
		return c.Pixels

	default:
		// transcendent_age / divine_age / cosmic_age — transcendent light pillar
		var c sprites.Canvas
		// Gold radiance expanding outward
		c.FillRect(0, 0, 16, 16, 0xE0C0FF)   // pale violet outer glow
		c.FillRect(2, 2, 12, 12, 0xFFE080)   // gold radiance halo
		c.FillRect(4, 4, 8, 8, 0xFFF0B0)     // brighter inner halo
		c.FillRect(5, 0, 6, 16, 0xFFFFFF)    // white light pillar vertical
		c.FillRect(0, 5, 16, 6, 0xFFFFFF)    // white light pillar horizontal
		// Intensify center
		c.FillRect(6, 4, 4, 8, 0xFFFFFF)
		c.FillRect(4, 6, 8, 4, 0xFFFFFF)
		c.Set(7, 7, 0xFFFFFF)
		c.Set(8, 7, 0xFFFFFF)
		c.Set(7, 8, 0xFFFFFF)
		c.Set(8, 8, 0xFFFFFF)
		return c.Pixels
	}
}

func v4SpriteWonder() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, v4wnG, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4wnG, v4wnG, v4wnG, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4wnL, v4wnS, v4wnD, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4wnL, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, v4wnD, 0, 0, 0, 0, 0},
		{0, 0, 0, v4wnL, v4wnL, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, v4wnD, v4wnD, 0, 0, 0, 0},
		{0, 0, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, 0, 0, 0},
		{0, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, 0, 0},
		{v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, 0},
		{v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, 0},
		{v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, 0},
		{v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, 0},
		{v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4SpriteStorage() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, 0, 0, 0, 0},
		{0, 0, 0, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, 0, 0, 0},
		{0, 0, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, 0, 0},
		{0, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, v4stR, 0},
		{0, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, 0},
		{0, v4stW, v4stLw, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stLw, v4stW, v4stW, 0},
		{0, v4stW, v4stLw, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stLw, v4stW, v4stW, 0},
		{0, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, 0},
		{0, v4stW, v4stW, v4stDw, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stDw, v4stW, v4stW, v4stW, 0},
		{0, v4stW, v4stW, v4stDw, v4stW, v4stO, v4stW, v4stW, v4stO, v4stW, v4stDw, v4stW, v4stW, v4stW, 0, 0},
		{0, v4stW, v4stW, v4stDw, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stDw, v4stW, v4stW, v4stW, 0, 0},
		{0, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, 0, 0},
		{0, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, 0, 0},
		{0, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, 0, 0},
		{0, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, v4stW, 0, 0},
	}
}

// Domain → sprite mapping
var v4DomainSprites = map[string][16][16]uint32{
	"food":       v4SpriteFood(),
	"wood":       v4SpriteWood(),
	"stone":      v4SpriteStone(),
	"iron":       v4SpriteIron(),
	"knowledge":  v4SpriteKnowledge(),
	"military":   v4SpriteMilitary(),
	"faith":      v4SpriteFaith(),
	"culture":    v4SpriteCulture(),
	"trade":      v4SpriteTrade(),
	"energy":     v4SpriteEnergy(),
	"metallurgy": v4SpriteMetallurgy(),
	"advanced":   v4SpriteAdvanced(),
}

var v4SpriteWonderPixels  = v4SpriteWonder()
var v4SpriteStoragePixels = v4SpriteStorage()
var v4SpriteCityPixels    = v4SpriteCityCenter()

// v4WonderSpriteByKey returns a unique 16×16 sprite for each of the 22 age wonders.
// Falls back to the generic pyramid if the key is unrecognised.
func v4WonderSpriteByKey(key string) [16][16]uint32 {
	switch key {
	case "sacred_grove":
		// Primitive age — circle of standing stones with a tree canopy
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x1A2A10)
		c.Dome(7, 6, 5, 0x3A7A20)
		c.Dome(7, 5, 3, 0x4A9A30)
		c.Set(7, 2, 0x60C040)
		c.Vline(7, 10, 3, 0x6B4020)
		c.Vline(8, 10, 3, 0x6B4020)
		stones := [][2]int{{2, 12}, {5, 14}, {10, 14}, {13, 12}, {14, 9}, {13, 6}, {2, 9}}
		for _, s := range stones {
			c.FillRect(s[0], s[1], 2, 2, 0xA09080)
		}
		return c.Pixels

	case "great_monolith":
		// Stone age — megalith row of standing stones
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x303030)
		c.Hline(0, 12, 16, 0x505050)
		posts := []int{1, 4, 7, 10, 13}
		for _, x := range posts {
			c.FillRect(x, 3, 2, 10, 0xA09080)
			c.Hline(x, 3, 2, 0xC0B0A0)
		}
		c.FillRect(4, 2, 5, 2, 0xC0B0A0)
		return c.Pixels

	case "stonehenge":
		// Bronze age — Stonehenge-style trilithon ring
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x202018)
		ring := [][2]int{{7, 1}, {11, 2}, {14, 5}, {14, 10}, {11, 13}, {7, 14}, {3, 13}, {2, 10}, {2, 5}, {4, 2}}
		for _, s := range ring {
			c.FillRect(s[0]-1, s[1]-1, 2, 3, 0x8C7C6C)
		}
		c.FillRect(5, 5, 2, 6, 0xA09080)
		c.FillRect(9, 5, 2, 6, 0xA09080)
		c.FillRect(4, 4, 8, 2, 0xC0B0A0)
		c.FillRect(6, 10, 4, 2, 0xD0C0B0)
		return c.Pixels

	case "colosseum":
		// Classical age — oval ring of arched walls
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x201808)
		c.Dome(7, 8, 7, 0xC8A868)
		c.Dome(7, 8, 5, 0x201808)
		c.FillRect(1, 10, 14, 5, 0xC8A868)
		c.FillRect(3, 10, 10, 3, 0x201808)
		for x := 1; x < 14; x += 2 {
			c.Set(x, 10, 0x201808)
		}
		return c.Pixels

	case "parthenon":
		// Iron age — Greek temple with columns and pediment
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x181810)
		c.FillRect(1, 13, 14, 2, 0xD0C898)
		c.FillRect(0, 12, 16, 1, 0xB8B080)
		for x := 1; x < 15; x += 2 {
			c.Vline(x, 6, 7, 0xE8E0C0)
		}
		c.FillRect(1, 5, 14, 2, 0xC8C090)
		c.FillRect(3, 3, 10, 2, 0xD0C898)
		c.FillRect(5, 2, 6, 1, 0xD0C898)
		c.Set(7, 1, 0xE8E0C0)
		c.Set(8, 1, 0xE8E0C0)
		return c.Pixels

	case "great_library":
		// Medieval age — library with arched windows
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x181818)
		c.FillRect(1, 13, 14, 2, 0x909090)
		c.FillRect(2, 5, 12, 9, 0xA0A0B0)
		c.FillRect(1, 3, 14, 3, 0x808090)
		c.FillRect(3, 1, 10, 3, 0x909090)
		c.Set(7, 0, 0xC0C0C0)
		c.Set(8, 0, 0xC0C0C0)
		for _, x := range []int{3, 7, 11} {
			c.FillRect(x, 7, 2, 4, 0x404060)
			c.Set(x, 6, 0x404060)
			c.Set(x+1, 6, 0x404060)
		}
		c.Hline(3, 9, 10, 0x808080)
		c.Hline(3, 11, 10, 0x808080)
		return c.Pixels

	case "sistine_chapel":
		// Renaissance age — chapel dome with cross
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x181818)
		c.FillRect(1, 9, 14, 6, 0xC0B890)
		c.FillRect(4, 6, 8, 4, 0xA8A080)
		c.Dome(7, 6, 4, 0xD8D0A8)
		c.FillRect(6, 1, 4, 3, 0xB0A880)
		c.Set(7, 0, 0xE8E0C0)
		c.Set(8, 0, 0xE8E0C0)
		c.Set(7, 4, 0x303028)
		c.Set(8, 4, 0x303028)
		return c.Pixels

	case "grand_lighthouse":
		// Colonial age — tall lighthouse with light beam
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x001830)
		c.FillRect(0, 13, 16, 3, 0x1040A0)
		c.Hline(0, 12, 16, 0x2060C0)
		c.FillRect(5, 3, 6, 10, 0xD0C8B0)
		c.FillRect(6, 1, 4, 3, 0xC0B8A0)
		c.FillRect(5, 0, 6, 2, 0xFFFF80)
		c.Set(4, 1, 0xFFFF40)
		c.Set(11, 1, 0xFFFF40)
		c.Set(1, 2, 0xFFFF80)
		c.Set(2, 1, 0xFFFF80)
		c.FillRect(3, 12, 10, 2, 0xA09080)
		return c.Pixels

	case "crystal_palace":
		// Industrial age — glass grid building
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x101820)
		c.FillRect(0, 2, 16, 4, 0x80C8E8)
		c.Dome(7, 4, 5, 0xA0E0F8)
		for x := 0; x < 16; x += 3 {
			c.Vline(x, 2, 12, 0x607080)
		}
		for y := 2; y < 14; y += 3 {
			c.Hline(0, y, 16, 0x607080)
		}
		c.FillRect(1, 3, 2, 2, 0xB8E8F8)
		c.FillRect(4, 3, 2, 2, 0xB8E8F8)
		c.FillRect(7, 3, 2, 2, 0xB8E8F8)
		c.FillRect(10, 3, 2, 2, 0xB8E8F8)
		c.FillRect(13, 3, 2, 2, 0xB8E8F8)
		c.FillRect(0, 13, 16, 3, 0x506070)
		return c.Pixels

	case "eiffel_tower":
		// Victorian age — Eiffel Tower silhouette
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x101018)
		c.FillRect(1, 10, 3, 5, 0x808070)
		c.FillRect(12, 10, 3, 5, 0x808070)
		c.FillRect(4, 8, 3, 4, 0x909080)
		c.FillRect(9, 8, 3, 4, 0x909080)
		c.Set(4, 10, 0x101018)
		c.Set(11, 10, 0x101018)
		c.FillRect(5, 5, 6, 4, 0xA0A090)
		c.FillRect(7, 1, 2, 5, 0xC0C0B0)
		c.Set(7, 0, 0xD8D8C8)
		c.Set(8, 0, 0xD8D8C8)
		c.Hline(0, 14, 16, 0x606060)
		return c.Pixels

	case "hoover_dam":
		// Electric age — dam wall with water
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x101820)
		c.FillRect(0, 0, 16, 6, 0x203050)
		c.FillRect(2, 3, 12, 10, 0x909090)
		c.FillRect(3, 4, 10, 9, 0xA0A0A0)
		c.FillRect(0, 0, 16, 4, 0x1050A0)
		c.Hline(0, 3, 16, 0x2080D0)
		c.FillRect(0, 12, 16, 4, 0x1040A0)
		c.Vline(4, 5, 7, 0x606060)
		c.Vline(7, 5, 7, 0x606060)
		c.Vline(10, 5, 7, 0x606060)
		c.Hline(3, 11, 10, 0xFFCC20)
		return c.Pixels

	case "particle_accelerator":
		// Atomic age — circular ring accelerator
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x080810)
		c.Dome(7, 8, 6, 0x4060A0)
		c.Dome(7, 8, 4, 0x080810)
		c.Dome(7, 8, 3, 0x2040C0)
		c.Dome(7, 8, 1, 0x080810)
		c.Set(7, 8, 0x80C0FF)
		c.Set(8, 8, 0x80C0FF)
		c.Hline(0, 8, 2, 0x6080FF)
		c.Hline(14, 8, 2, 0x6080FF)
		c.Vline(7, 0, 2, 0x6080FF)
		c.Vline(7, 14, 2, 0x6080FF)
		return c.Pixels

	case "space_program":
		// Space age — rocket on launch pad
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x050510)
		c.Set(2, 1, 0xFFFFFF)
		c.Set(6, 3, 0xFFFFFF)
		c.Set(11, 0, 0xFFFFFF)
		c.Set(14, 4, 0xFFFFFF)
		c.Vline(2, 4, 10, 0x707070)
		c.Hline(2, 6, 3, 0x707070)
		c.Hline(2, 9, 3, 0x707070)
		c.FillRect(6, 3, 4, 9, 0xD0D0D0)
		c.FillRect(7, 1, 2, 3, 0xE0E0E0)
		c.Set(7, 0, 0xC0C0C0)
		c.Set(8, 0, 0xC0C0C0)
		c.FillRect(5, 10, 2, 3, 0xA0A0B0)
		c.FillRect(9, 10, 2, 3, 0xA0A0B0)
		c.Hline(6, 12, 4, 0xFF8020)
		c.Hline(7, 13, 2, 0xFFAA40)
		c.FillRect(4, 13, 8, 2, 0x606070)
		return c.Pixels

	case "global_network":
		// Information age — data center server rack pattern
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x0A0A18)
		c.FillRect(1, 1, 6, 14, 0x303048)
		c.FillRect(9, 1, 6, 14, 0x303048)
		for y := 2; y < 14; y += 2 {
			c.Hline(2, y, 4, 0x404860)
			c.Set(5, y, 0x00FF80)
			c.Hline(10, y, 4, 0x404860)
			c.Set(13, y, 0x0080FF)
		}
		c.Hline(7, 4, 2, 0x00FFFF)
		c.Hline(7, 8, 2, 0x00FFFF)
		c.Hline(7, 12, 2, 0x00FFFF)
		return c.Pixels

	case "world_simulation":
		// Cyberpunk age — neon sprawl cityscape
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x080810)
		c.FillRect(0, 8, 3, 8, 0x202028)
		c.FillRect(4, 6, 2, 10, 0x202028)
		c.FillRect(7, 4, 3, 12, 0x202028)
		c.FillRect(11, 7, 2, 9, 0x202028)
		c.FillRect(13, 5, 3, 11, 0x202028)
		c.Hline(0, 9, 16, 0xCC00FF)
		c.Hline(0, 12, 16, 0x00FFFF)
		c.Hline(0, 14, 16, 0xFF0080)
		c.Set(1, 10, 0xFF40FF)
		c.Set(5, 8, 0x40FFFF)
		c.Set(8, 6, 0xFF40FF)
		c.Set(12, 9, 0x40FFFF)
		c.Set(14, 7, 0xFF40FF)
		return c.Pixels

	case "neon_citadel":
		// Digital age — quantum hexagonal pattern
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x040814)
		hexPts := [][2]int{
			{7, 1}, {11, 3}, {13, 7}, {11, 11}, {7, 13}, {3, 11}, {1, 7}, {3, 3},
		}
		for _, p := range hexPts {
			c.Set(p[0], p[1], 0xA0C0FF)
			if p[0]+1 < 16 {
				c.Set(p[0]+1, p[1], 0xA0C0FF)
			}
		}
		inner := [][2]int{{7, 4}, {10, 6}, {10, 10}, {7, 12}, {4, 10}, {4, 6}}
		for _, p := range inner {
			c.Set(p[0], p[1], 0x60A0FF)
		}
		c.FillRect(6, 6, 4, 4, 0x2040A0)
		c.Set(7, 7, 0xE0F0FF)
		c.Set(8, 7, 0xE0F0FF)
		c.Set(7, 8, 0xE0F0FF)
		c.Set(8, 8, 0xE0F0FF)
		return c.Pixels

	case "stellar_cradle":
		// Fusion age — torus ring with plasma core
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x020208)
		c.Dome(7, 8, 6, 0x2060C0)
		c.Dome(7, 8, 4, 0x020208)
		c.Dome(7, 8, 3, 0xFF6020)
		c.Dome(7, 8, 1, 0xFFB040)
		c.Set(7, 8, 0xFFFF80)
		c.Set(8, 8, 0xFFFF80)
		c.Hline(2, 8, 3, 0x4080D0)
		c.Hline(11, 8, 3, 0x4080D0)
		return c.Pixels

	case "dyson_scaffold":
		// Interstellar age — elongated silver ark ship
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x020208)
		c.Set(1, 2, 0xFFFFFF)
		c.Set(5, 0, 0xFFFFFF)
		c.Set(12, 3, 0xFFFFFF)
		c.Set(14, 1, 0xFFFFFF)
		c.FillRect(1, 6, 14, 4, 0xA0B0C0)
		c.FillRect(2, 5, 12, 6, 0xB0C0D0)
		c.Set(0, 7, 0x808090)
		c.Set(0, 8, 0x808090)
		c.FillRect(12, 4, 3, 2, 0x607080)
		c.FillRect(12, 10, 3, 2, 0x607080)
		c.Set(15, 5, 0x40C0FF)
		c.Set(15, 6, 0x40C0FF)
		c.Set(15, 9, 0x40C0FF)
		c.Set(15, 10, 0x40C0FF)
		c.Hline(4, 7, 7, 0x6080FF)
		return c.Pixels

	case "warp_nexus":
		// Galactic age — dyson sphere orange/gold ring
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x020208)
		c.Set(1, 1, 0xFFFFFF)
		c.Set(13, 2, 0xFFFFFF)
		c.Set(3, 13, 0xFFFFFF)
		c.Set(14, 12, 0xFFFFFF)
		c.Dome(7, 8, 6, 0xC87020)
		c.Dome(7, 8, 4, 0x020208)
		c.Dome(7, 8, 3, 0xFFAA20)
		c.FillRect(5, 6, 6, 6, 0xFFCC40)
		c.Set(7, 7, 0xFFFF80)
		c.Set(8, 7, 0xFFFF80)
		c.Set(7, 8, 0xFFFF80)
		c.Set(8, 8, 0xFFFF80)
		c.Set(7, 2, 0xE09030)
		c.Set(7, 14, 0xE09030)
		c.Set(1, 8, 0xE09030)
		c.Set(14, 8, 0xE09030)
		return c.Pixels

	case "cosmic_beacon":
		// Quantum age — reality anchor deep purple with gold geometric star
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x0A0018)
		c.Vline(7, 0, 3, 0xFFCC20)
		c.Vline(7, 13, 3, 0xFFCC20)
		c.Hline(0, 7, 3, 0xFFCC20)
		c.Hline(13, 7, 3, 0xFFCC20)
		c.Set(3, 3, 0xFFAA10)
		c.Set(4, 4, 0xFFAA10)
		c.Set(11, 3, 0xFFAA10)
		c.Set(12, 4, 0xFFAA10)
		c.Set(3, 12, 0xFFAA10)
		c.Set(4, 11, 0xFFAA10)
		c.Set(11, 12, 0xFFAA10)
		c.Set(12, 11, 0xFFAA10)
		c.FillRect(4, 4, 8, 8, 0x400080)
		c.FillRect(5, 5, 6, 6, 0x600090)
		c.Set(7, 7, 0xFFCC20)
		c.Set(8, 7, 0xFFCC20)
		c.Set(7, 8, 0xFFCC20)
		c.Set(8, 8, 0xFFCC20)
		return c.Pixels

	case "reality_anchor":
		// Quantum age — reality anchor with radiating chain lines
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x0A0018)
		c.Vline(7, 0, 16, 0x5000A0)
		c.Hline(0, 7, 16, 0x5000A0)
		c.Vline(8, 0, 16, 0x5000A0)
		c.Hline(0, 8, 16, 0x5000A0)
		c.Set(7, 1, 0xFFCC20)
		c.Set(7, 14, 0xFFCC20)
		c.Set(1, 7, 0xFFCC20)
		c.Set(14, 7, 0xFFCC20)
		c.Set(4, 4, 0xCC8810)
		c.Set(11, 4, 0xCC8810)
		c.Set(4, 11, 0xCC8810)
		c.Set(11, 11, 0xCC8810)
		c.FillRect(5, 5, 6, 6, 0x300060)
		c.FillRect(6, 6, 4, 4, 0x500090)
		c.Set(7, 7, 0xE0A020)
		c.Set(8, 7, 0xE0A020)
		c.Set(7, 8, 0xE0A020)
		c.Set(8, 8, 0xE0A020)
		return c.Pixels

	case "singularity_core":
		// Transcendent age — world tree with luminous tip
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x040C04)
		c.FillRect(6, 13, 4, 3, 0x5A3010)
		c.Set(4, 14, 0x3A2008)
		c.Set(5, 15, 0x3A2008)
		c.Set(10, 14, 0x3A2008)
		c.Set(11, 15, 0x3A2008)
		c.FillRect(7, 7, 2, 7, 0x6B4020)
		c.Set(5, 8, 0x5A3810)
		c.Set(6, 7, 0x5A3810)
		c.Set(9, 7, 0x5A3810)
		c.Set(10, 8, 0x5A3810)
		c.Dome(7, 9, 4, 0x206820)
		c.Dome(7, 7, 3, 0x30A030)
		c.Dome(7, 5, 3, 0x50C040)
		c.Dome(7, 3, 2, 0x70E060)
		c.Set(7, 1, 0xC0FF80)
		c.Set(8, 1, 0xC0FF80)
		c.Set(7, 0, 0xFFFFCC)
		return c.Pixels

	default:
		return v4SpriteWonderPixels
	}
}

var (
	v4SpriteCacheMu sync.RWMutex
	v4SpriteCache   = map[string][16][16]uint32{}
)

// v4BgCache caches pre-scaled PNG background pixel rows for each age key.
// Key format: "<ageKey>:<width>x<height>". Cleared when age changes.
var (
	v4BgCacheMu sync.RWMutex
	v4BgCache   = map[string][]uint32{}
)

// v4LoadBgPixels loads assets/maps/<ageKey>.png and scales it to w×h using
// nearest-neighbour resampling. Returns nil if the file doesn't exist or
// fails to decode. Results are cached; cache is cleared on age transition.
func v4LoadBgPixels(ageKey string, w, h int) []uint32 {
	// Build a compact string key that encodes age + dimensions.
	cacheKey := ageKey + ":" + v4itoa(w) + "x" + v4itoa(h)

	v4BgCacheMu.RLock()
	if px, ok := v4BgCache[cacheKey]; ok {
		v4BgCacheMu.RUnlock()
		return px
	}
	v4BgCacheMu.RUnlock()

	path := "maps/" + ageKey + ".png"
	f, err := gameAssets.FS.Open(path)
	if err != nil {
		return nil // file not found — fall back to FBM
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return nil // corrupt / unsupported — fall back to FBM
	}

	sw := src.Bounds().Dx()
	sh := src.Bounds().Dy()
	ox := src.Bounds().Min.X
	oy := src.Bounds().Min.Y

	// Area-averaging (box filter) downscale: average all source pixels that map
	// to each destination pixel. Eliminates the aliasing from nearest-neighbour
	// when downscaling a large PNG to terminal dimensions.
	px := make([]uint32, w*h)
	for dy := 0; dy < h; dy++ {
		sy0 := oy + dy*sh/h
		sy1 := oy + (dy+1)*sh/h
		if sy1 <= sy0 {
			sy1 = sy0 + 1
		}
		for dx := 0; dx < w; dx++ {
			sx0 := ox + dx*sw/w
			sx1 := ox + (dx+1)*sw/w
			if sx1 <= sx0 {
				sx1 = sx0 + 1
			}
			var rSum, gSum, bSum, n uint32
			for sy := sy0; sy < sy1; sy++ {
				for sx := sx0; sx < sx1; sx++ {
					r, g, b, _ := src.At(sx, sy).RGBA()
					rSum += r >> 8
					gSum += g >> 8
					bSum += b >> 8
					n++
				}
			}
			px[dy*w+dx] = ((rSum/n) << 16) | ((gSum/n) << 8) | (bSum / n)
		}
	}

	v4BgCacheMu.Lock()
	v4BgCache[cacheKey] = px
	v4BgCacheMu.Unlock()
	return px
}

// v4itoa converts a non-negative int to its decimal string without fmt.
func v4itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}

// v4BuildingLookup maps building key → [2]int{lineage, tier} for in-memory sprite generation.
var v4BuildingLookup = func() map[string][2]int {
	m := make(map[string][2]int, len(sprites.AllBuildings))
	for _, b := range sprites.AllBuildings {
		m[b.Key] = [2]int{b.Lineage, b.Tier}
	}
	return m
}()

// v4StableHash returns a stable FNV-1a hash for a string, used to pick sprite variations.
func v4StableHash(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

// v4LoadBuildingSprite returns a 16×16 sprite for a building.
// It first tries to load a pre-rendered PNG from assets/sprites/buildings/<key>_var<N>.png.
// If the file is not found or cannot be decoded, it falls back to in-memory generation.
// Results are cached per (key, ageKey) pair so age-advancement triggers fresh generation.
func v4LoadBuildingSprite(key, domain, ageKey string, isWonder, isStorage bool) [16][16]uint32 {
	cacheKey := key + ":" + ageKey
	v4SpriteCacheMu.RLock()
	if px, ok := v4SpriteCache[cacheKey]; ok {
		v4SpriteCacheMu.RUnlock()
		return px
	}
	v4SpriteCacheMu.RUnlock()

	var px [16][16]uint32

	// Try loading a pre-rendered PNG sprite first.
	variation := int(v4StableHash(key)) % 3
	var pngSuffix string
	switch variation {
	case 1:
		pngSuffix = "_v2"
	case 2:
		pngSuffix = "_v3"
	default:
		pngSuffix = ""
	}
	pngPath := "sprites/buildings/" + key + pngSuffix + ".png"
	if f, err := gameAssets.FS.Open(pngPath); err == nil {
		img, _, decodeErr := image.Decode(f)
		f.Close()
		if decodeErr == nil {
			bounds := img.Bounds()
			w := bounds.Max.X - bounds.Min.X
			h := bounds.Max.Y - bounds.Min.Y
			if w >= 16 && h >= 16 {
				for row := 0; row < 16; row++ {
					for col := 0; col < 16; col++ {
						r, g, b, _ := img.At(bounds.Min.X+col, bounds.Min.Y+row).RGBA()
						px[row][col] = (uint32(r>>8) << 16) | (uint32(g>>8) << 8) | uint32(b>>8)
					}
				}
				v4SpriteCacheMu.Lock()
				v4SpriteCache[cacheKey] = px
				v4SpriteCacheMu.Unlock()
				return px
			}
		}
	}

	// Fall back to in-memory generation.
	if lt, ok := v4BuildingLookup[key]; ok {
		px = sprites.GenerateBuildingSprite(lt[0], lt[1], ageKey, 0)
	} else if isWonder {
		px = v4WonderSpriteByKey(key)
	} else if isStorage {
		px = v4SpriteStoragePixels
	} else if sp, ok := v4DomainSprites[domain]; ok {
		px = sp
	} else {
		px = v4SpriteCityPixels
	}

	v4SpriteCacheMu.Lock()
	v4SpriteCache[cacheKey] = px
	v4SpriteCacheMu.Unlock()
	return px
}

// ── Sprite renderer ────────────────────────────────────────────────────────

func v4DrawSprite(img *image.RGBA, pixels [16][16]uint32, px, py int) {
	for row := 0; row < 16; row++ {
		for col := 0; col < 16; col++ {
			v := pixels[row][col]
			if v == 0 {
				continue
			}
			c := color.RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 255}
			img.SetRGBA(px+col, py+row, c)
		}
	}
}

// v4DrawSpriteScaled draws a 16×16 sprite scaled to spriteSize×spriteSize using nearest-neighbor.
func v4DrawSpriteScaled(img *image.RGBA, pixels [16][16]uint32, px, py, spriteSize int) {
	if spriteSize <= 0 {
		return
	}
	for dstRow := 0; dstRow < spriteSize; dstRow++ {
		for dstCol := 0; dstCol < spriteSize; dstCol++ {
			srcRow := dstRow * 16 / spriteSize
			srcCol := dstCol * 16 / spriteSize
			v := pixels[srcRow][srcCol]
			if v == 0 {
				continue
			}
			c := color.RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 255}
			x, y := px+dstCol, py+dstRow
			if x >= 0 && x < img.Bounds().Max.X && y >= 0 && y < img.Bounds().Max.Y {
				img.SetRGBA(x, y, c)
			}
		}
	}
	// 1px shadow outline below and to the right for depth/contrast
	shadowClr := color.RGBA{0, 0, 0, 120}
	for dstRow := 0; dstRow < spriteSize; dstRow++ {
		for dstCol := 0; dstCol < spriteSize; dstCol++ {
			srcRow := dstRow * 16 / spriteSize
			srcCol := dstCol * 16 / spriteSize
			v := pixels[srcRow][srcCol]
			if v == 0 {
				continue
			}
			// Check if pixel below is transparent/outside → draw shadow below it
			belowRow := (dstRow + 1) * 16 / spriteSize
			belowV := uint32(0)
			if belowRow < 16 {
				belowV = pixels[belowRow][srcCol]
			}
			if belowV == 0 {
				x, y := px+dstCol, py+dstRow+1
				if x >= 0 && x < img.Bounds().Max.X && y >= 0 && y < img.Bounds().Max.Y {
					existing := img.RGBAAt(x, y)
					blended := color.RGBA{
						R: uint8((int(existing.R)*2 + int(shadowClr.R)) / 3),
						G: uint8((int(existing.G)*2 + int(shadowClr.G)) / 3),
						B: uint8((int(existing.B)*2 + int(shadowClr.B)) / 3),
						A: 255,
					}
					img.SetRGBA(x, y, blended)
				}
			}
		}
	}
}

// ── Clearing halo ──────────────────────────────────────────────────────────

func v4DrawClearingR(img *image.RGBA, cx, cy int, clr color.RGBA, r int) {
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > float64(r) {
				continue
			}
			alpha := float64(1.0) - dist/float64(r)
			alpha = alpha * alpha * 0.6
			x, y := cx+dx, cy+dy
			if x < 0 || y < 0 || x >= img.Bounds().Max.X || y >= img.Bounds().Max.Y {
				continue
			}
			base := img.RGBAAt(x, y)
			blended := color.RGBA{
				R: uint8(float64(base.R)*(1-alpha) + float64(clr.R)*alpha),
				G: uint8(float64(base.G)*(1-alpha) + float64(clr.G)*alpha),
				B: uint8(float64(base.B)*(1-alpha) + float64(clr.B)*alpha),
				A: 255,
			}
			img.SetRGBA(x, y, blended)
		}
	}
}

// ── Organic city slot generation ───────────────────────────────────────────

const (
	v4SlotSpacing  = 28 // pixels between jittered-grid cells (fallback default)
	v4JitterAmount = 10 // max random offset per slot in each axis (fallback default)
	v4MinDist      = 18 // minimum distance between two placed buildings (fallback default)
	v4CenterRadius = 20 // reserved radius around palace — no buildings placed here
	wonderStripH   = 20 // pixel height of the wonder row reserved at the bottom of the map
)

// v4DensityParams holds layout parameters for the city at a given era and growth stage.
type v4DensityParams struct {
	slotSpacing int     // pixels between grid slots
	minDist     int     // minimum pixels between placed sprites
	jitterAmt   int     // max random offset per slot
	spriteScale float64 // multiplier on top of auto-zoom scale
	maxPerType  int     // max sprites per building type
	maxTotal    int     // hard cap on total sprites
}

// v4CityDensity returns layout parameters based on era + totalBuilt.
// Era 0 is a sparse primitive settlement; era 6 is an alien megacity.
// Within each era, every 20 buildings built tightens spacing by 1px (floor at 60% of base).
func v4CityDensity(era int, totalBuilt int) v4DensityParams {
	type eraBase struct {
		spacing, minDist, jitter int
		spriteScale              float64
		maxPerType, maxTotal     int
	}
	bases := [7]eraBase{
		{44, 32, 16, 1.0, 6, 60},   // era 0: primitive — sparse village
		{34, 24, 13, 1.1, 7, 80},   // era 1: classical/medieval — small town
		{26, 18, 10, 1.2, 8, 100},  // era 2: renaissance/industrial — market town
		{20, 13, 8, 1.3, 9, 130},   // era 3: modern/atomic — dense city
		{15, 9, 6, 1.4, 10, 160},   // era 4: space/cyber — megacity
		{11, 6, 4, 1.5, 12, 200},   // era 5: nanotech/fusion — ultra-dense
		{7, 3, 2, 1.6, 15, 250},    // era 6: galactic/cosmic/divine — alien megacity
	}
	if era < 0 {
		era = 0
	}
	if era > 6 {
		era = 6
	}
	b := bases[era]

	// Within-era growth: every 20 buildings built, tighten spacing by 1px (floor at 60% of base)
	growthSteps := totalBuilt / 20
	minSpacing := int(float64(b.spacing) * 0.6)
	spacing := b.spacing - growthSteps
	if spacing < minSpacing {
		spacing = minSpacing
	}

	minDist := b.minDist - growthSteps/2
	minMinDist := int(float64(b.minDist) * 0.6)
	if minDist < minMinDist {
		minDist = minMinDist
	}

	return v4DensityParams{
		slotSpacing: spacing,
		minDist:     minDist,
		jitterAmt:   b.jitter,
		spriteScale: b.spriteScale,
		maxPerType:  b.maxPerType,
		maxTotal:    b.maxTotal,
	}
}

// v4SlotJitter returns a deterministic jitter offset in [-jitterAmt, +jitterAmt]
// for a grid cell (gx, gy) and a salt value, using integer hashing.
func v4SlotJitter(gx, gy, salt int) (jx, jy int) {
	const jRange = 2*v4JitterAmount + 1
	hx := gx*73856093 ^ gy*19349663 ^ salt*83492791
	hy := gx*19349663 ^ gy*73856093 ^ salt*41728397
	// force positive modulo
	hxMod := hx % jRange
	if hxMod < 0 {
		hxMod += jRange
	}
	hyMod := hy % jRange
	if hyMod < 0 {
		hyMod += jRange
	}
	return hxMod - v4JitterAmount, hyMod - v4JitterAmount
}

// v4BuildingSlots generates candidate building positions using a jittered grid
// expanding outward from (cx, cy). Slots are sorted by distance from center
// (ascending) so buildings fill from center outward.
// Returns positions as offsets from center (not absolute pixel coords).
// slotSpacing controls grid cell size; jitterAmt controls max per-cell random offset.
func v4BuildingSlots(width, height, cx, cy, slotSpacing, jitterAmt int) [][2]int {
	const gridRange = 10 // -10 to +10 in each axis → 441 cells
	const seed = 7

	type slot struct {
		ox, oy int
		dist2  int
	}
	var candidates []slot

	for gy := -gridRange; gy <= gridRange; gy++ {
		for gx := -gridRange; gx <= gridRange; gx++ {
			// Skip center — palace lives there
			if gx == 0 && gy == 0 {
				continue
			}
			worldX := gx * slotSpacing
			worldY := gy * slotSpacing
			jx, jy := v4SlotJitter(gx, gy, seed)
			// Clamp jitter to jitterAmt (v4SlotJitter uses the const, so we rescale)
			if jitterAmt != v4JitterAmount && v4JitterAmount > 0 {
				jx = jx * jitterAmt / v4JitterAmount
				jy = jy * jitterAmt / v4JitterAmount
			}
			ox := worldX + jx
			oy := worldY + jy

			// Skip if too close to palace
			d2center := ox*ox + oy*oy
			if d2center < v4CenterRadius*v4CenterRadius {
				continue
			}

			// Skip if outside image bounds (also reserve bottom wonder strip)
			absX := cx + ox
			absY := cy + oy
			if absX < 8 || absX >= width-8 || absY < 8 || absY >= height-wonderStripH-8 {
				continue
			}

			candidates = append(candidates, slot{ox, oy, d2center})
		}
	}

	// Sort by distance from center ascending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].dist2 < candidates[j].dist2
	})

	result := make([][2]int, len(candidates))
	for i, c := range candidates {
		result[i] = [2]int{c.ox, c.oy}
	}
	return result
}

// v4AssignSlots picks the first N available slots from the candidate pool,
// ensuring no two placed buildings are within minDist pixels of each other.
func v4AssignSlots(slots [][2]int, n, minDist int) [][2]int {
	var placed [][2]int
	for _, s := range slots {
		if len(placed) >= n {
			break
		}
		ok := true
		for _, p := range placed {
			dx := s[0] - p[0]
			dy := s[1] - p[1]
			if dx*dx+dy*dy < minDist*minDist {
				ok = false
				break
			}
		}
		if ok {
			placed = append(placed, s)
		}
	}
	return placed
}

// ── Building list ──────────────────────────────────────────────────────────

type v4Building struct {
	key       string
	domain    string
	isWonder  bool
	isStorage bool
	count     int
}

func v4GetBuildings(state game.GameState, params v4DensityParams) []v4Building {
	var result []v4Building
	for key, bs := range state.Buildings {
		if bs.Count == 0 {
			continue
		}
		// Skip hidden/meta buildings that don't make sense as city sprites
		if key == "stash" || key == "ruins" {
			continue
		}
		isWonder := false
		isStorage := false
		def, defOk := config.BuildingByKey()[key]
		if defOk && def.Category == "wonder" {
			isWonder = true
		} else if len(key) >= 6 && key[len(key)-6:] == "wonder" {
			isWonder = true
		}
		// Only render buildings that belong to the current age.
		// If RequiredAge is empty (untagged building), include it as a fallback.
		if defOk && def.RequiredAge != "" && def.RequiredAge != state.Age {
			continue
		}
		// Wonders are rendered separately in the bottom strip — exclude from city slots.
		if isWonder {
			continue
		}
		for _, kw := range []string{"storage", "warehouse", "granary", "silo", "vault", "barn", "stockpile", "depot", "repository"} {
			if strings.Contains(key, kw) {
				isStorage = true
			}
		}
		result = append(result, v4Building{
			key:       key,
			domain:    bs.WorkerDomain,
			isWonder:  isWonder,
			isStorage: isStorage,
			count:     bs.Count,
		})
	}
	// Expand: one sprite entry per building instance (Count copies per type).
	// Use density params for per-type and total caps.
	maxPerType := params.maxPerType
	maxTotal := params.maxTotal
	var expanded []v4Building
	sort.Slice(result, func(i, j int) bool {
		if result[i].isWonder != result[j].isWonder {
			return result[i].isWonder
		}
		if result[i].domain != result[j].domain {
			return result[i].domain < result[j].domain
		}
		return result[i].key < result[j].key
	})
	for _, b := range result {
		copies := b.count
		if copies > maxPerType {
			copies = maxPerType
		}
		for i := 0; i < copies; i++ {
			expanded = append(expanded, b)
			if len(expanded) >= maxTotal {
				return expanded
			}
		}
	}
	return expanded
}

// ── Noise (renamed from v2 to avoid conflicts) ─────────────────────────────

func v4Hash(x, y uint32) uint32 {
	h := x*2654435761 ^ y*2246822519
	h ^= h >> 16
	h *= 0x45d9f3b
	h ^= h >> 16
	return h
}

func v4Noise(fx, fy float64, seed uint32) float64 {
	x0 := int(math.Floor(fx))
	y0 := int(math.Floor(fy))
	x1 := x0 + 1
	y1 := y0 + 1

	tx := fx - math.Floor(fx)
	ty := fy - math.Floor(fy)
	tx = tx * tx * (3 - 2*tx)
	ty = ty * ty * (3 - 2*ty)

	v00 := float64(v4Hash(uint32(x0)+seed, uint32(y0)+seed)) / float64(^uint32(0))
	v10 := float64(v4Hash(uint32(x1)+seed, uint32(y0)+seed)) / float64(^uint32(0))
	v01 := float64(v4Hash(uint32(x0)+seed, uint32(y1)+seed)) / float64(^uint32(0))
	v11 := float64(v4Hash(uint32(x1)+seed, uint32(y1)+seed)) / float64(^uint32(0))

	return v00*(1-tx)*(1-ty) + v10*tx*(1-ty) + v01*(1-tx)*ty + v11*tx*ty
}

func v4FBM(x, y float64, seed uint32) float64 {
	n := v4Noise(x/40.0, y/40.0, seed) * 0.5
	n += v4Noise(x/20.0, y/20.0, seed+100) * 0.3
	n += v4Noise(x/10.0, y/10.0, seed+200) * 0.2
	return n
}

// ── Grid lines ─────────────────────────────────────────────────────────────

func v4Min8(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func v4Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── Terrain feature sprites (8×8, 0 = transparent) ─────────────────────────

// v4TerrainFeature is a small pixel art terrain decoration (8×8, 0=transparent)
type v4TerrainFeature [8][8]uint32

func v4FeatureJungleTree() v4TerrainFeature {
	G := uint32(0x0d3a08)
	M := uint32(0x1a5c10)
	L := uint32(0x2d7a1a)
	T := uint32(0x3d2010)
	return v4TerrainFeature{
		{0, 0, 0, M, 0, 0, 0, 0},
		{0, 0, M, L, M, 0, 0, 0},
		{0, M, L, L, L, M, 0, 0},
		{G, M, L, L, L, M, G, 0},
		{0, G, M, L, M, G, 0, 0},
		{0, 0, 0, T, 0, 0, 0, 0},
		{0, 0, 0, T, 0, 0, 0, 0},
		{0, 0, T, T, T, 0, 0, 0},
	}
}

func v4FeaturePineTree() v4TerrainFeature {
	D := uint32(0x0a2a08)
	M := uint32(0x164810)
	T := uint32(0x3d2010)
	return v4TerrainFeature{
		{0, 0, 0, M, 0, 0, 0, 0},
		{0, 0, D, M, D, 0, 0, 0},
		{0, D, M, M, M, D, 0, 0},
		{0, 0, D, M, D, 0, 0, 0},
		{D, M, M, M, M, M, D, 0},
		{0, 0, 0, T, 0, 0, 0, 0},
		{0, 0, 0, T, 0, 0, 0, 0},
		{0, 0, T, T, T, 0, 0, 0},
	}
}

func v4FeatureRock() v4TerrainFeature {
	L := uint32(0xbbbbbb)
	G := uint32(0x888888)
	D := uint32(0x444444)
	return v4TerrainFeature{
		{0, 0, G, G, G, 0, 0, 0},
		{0, G, L, L, G, G, 0, 0},
		{G, L, L, G, G, G, G, 0},
		{G, L, G, G, G, G, G, 0},
		{G, G, G, G, G, G, D, 0},
		{0, G, G, G, G, D, 0, 0},
		{0, 0, D, D, D, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4FeatureCropField() v4TerrainFeature {
	C := uint32(0xe8c840)
	Gr := uint32(0x6a8a2a)
	return v4TerrainFeature{
		{Gr, Gr, Gr, Gr, Gr, Gr, Gr, Gr},
		{C, 0, C, 0, C, 0, C, 0},
		{C, C, C, C, C, C, C, C},
		{Gr, Gr, Gr, Gr, Gr, Gr, Gr, Gr},
		{C, 0, C, 0, C, 0, C, 0},
		{C, C, C, C, C, C, C, C},
		{Gr, Gr, Gr, Gr, Gr, Gr, Gr, Gr},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4FeatureCityBlock() v4TerrainFeature {
	W := uint32(0x808890)
	Gl := uint32(0x80c0e0)
	D := uint32(0x404850)
	return v4TerrainFeature{
		{D, D, D, D, D, D, D, D},
		{D, W, Gl, W, Gl, W, Gl, D},
		{D, W, Gl, W, Gl, W, Gl, D},
		{D, W, W, W, W, W, W, D},
		{D, W, Gl, W, Gl, W, Gl, D},
		{D, W, Gl, W, Gl, W, Gl, D},
		{D, W, W, W, W, W, W, D},
		{D, D, D, D, D, D, D, D},
	}
}

func v4FeatureNeonBlock() v4TerrainFeature {
	B := uint32(0x101020)
	C := uint32(0x40c0ff)
	P := uint32(0xff40c0)
	return v4TerrainFeature{
		{B, C, B, B, B, B, C, B},
		{C, B, B, B, B, B, B, C},
		{B, B, P, B, B, P, B, B},
		{B, B, B, B, B, B, B, B},
		{B, B, P, B, B, P, B, B},
		{C, B, B, B, B, B, B, C},
		{B, C, B, B, B, B, C, B},
		{B, B, C, C, C, C, B, B},
	}
}

func v4FeatureEnergyCrystal() v4TerrainFeature {
	Gb := uint32(0x4080ff)
	L := uint32(0x80c0ff)
	W := uint32(0xffffff)
	return v4TerrainFeature{
		{0, 0, 0, W, 0, 0, 0, 0},
		{0, 0, L, Gb, L, 0, 0, 0},
		{0, L, Gb, L, Gb, L, 0, 0},
		{L, Gb, L, W, L, Gb, L, 0},
		{0, L, Gb, L, Gb, L, 0, 0},
		{0, 0, L, Gb, L, 0, 0, 0},
		{0, 0, 0, Gb, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4FeatureVoidParticle() v4TerrainFeature {
	P := uint32(0xc0a0ff)
	Gd := uint32(0xffd040)
	return v4TerrainFeature{
		{0, 0, 0, P, 0, 0, 0, 0},
		{0, 0, P, 0, P, 0, 0, 0},
		{0, P, 0, Gd, 0, P, 0, 0},
		{P, 0, Gd, 0, Gd, 0, P, 0},
		{0, P, 0, Gd, 0, P, 0, 0},
		{0, 0, P, 0, P, 0, 0, 0},
		{0, 0, 0, P, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func v4DrawTerrainFeature(img *image.RGBA, feat v4TerrainFeature, px, py int) {
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			v := feat[row][col]
			if v == 0 {
				continue
			}
			x, y := px+col, py+row
			if x < 0 || y < 0 || x >= img.Bounds().Max.X || y >= img.Bounds().Max.Y {
				continue
			}
			img.SetRGBA(x, y, color.RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 255})
		}
	}
}

// v4FeatureSetForEra returns features to scatter for a given era index and density.
func v4FeatureSetForEra(era int) (features []v4TerrainFeature, density float64) {
	switch era {
	case 0:
		return []v4TerrainFeature{v4FeatureJungleTree(), v4FeatureRock()}, 0.45
	case 1:
		return []v4TerrainFeature{v4FeaturePineTree(), v4FeatureCropField()}, 0.38
	case 2:
		return []v4TerrainFeature{v4FeatureCropField(), v4FeatureCityBlock()}, 0.32
	case 3:
		return []v4TerrainFeature{v4FeatureCityBlock(), v4FeatureCityBlock()}, 0.38
	case 4:
		return []v4TerrainFeature{v4FeatureNeonBlock(), v4FeatureCityBlock()}, 0.32
	case 5:
		return []v4TerrainFeature{v4FeatureEnergyCrystal(), v4FeatureEnergyCrystal()}, 0.28
	default:
		return []v4TerrainFeature{v4FeatureVoidParticle()}, 0.22
	}
}

func v4DrawGridLines(img *image.RGBA, pal v4AgePalette) {
	bounds := img.Bounds()
	gridClr := color.RGBA{
		R: uint8(v4Min8(int(pal.bgMid.R)+20, 255)),
		G: uint8(v4Min8(int(pal.bgMid.G)+20, 255)),
		B: uint8(v4Min8(int(pal.bgMid.B)+30, 255)),
		A: 255,
	}
	for x := 0; x < bounds.Max.X; x += 16 {
		for y := 0; y < bounds.Max.Y; y++ {
			img.SetRGBA(x, y, gridClr)
		}
	}
	for y := 0; y < bounds.Max.Y; y += 16 {
		for x := 0; x < bounds.Max.X; x++ {
			img.SetRGBA(x, y, gridClr)
		}
	}
}

// ── River ──────────────────────────────────────────────────────────────────

func v4DrawRiver(img *image.RGBA, pal v4AgePalette, seed uint32) {
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	riverClr := pal.riverColor
	x := w / 5
	for y := 0; y < h; y++ {
		noise := v4Hash(seed, uint32(y/8)) % 7
		x += int(noise) - 3
		if x < 4 {
			x = 4
		}
		if x > w/3 {
			x = w / 3
		}
		for rx := x - 2; rx <= x+2; rx++ {
			if rx >= 0 && rx < w {
				img.SetRGBA(rx, y, riverClr)
			}
		}
	}
}

// ── Color helpers ──────────────────────────────────────────────────────────

func v4LerpColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: 255,
	}
}

// ── Main image generator ───────────────────────────────────────────────────

func v4GenerateImage(state game.GameState, width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	pal := v4GetPalette(state.Age)
	seed := uint32(42)

	// 1. Background: PNG image for the age if available, else FBM terrain.
	bgPixels := v4LoadBgPixels(state.Age, width, height)
	if bgPixels != nil {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				packed := bgPixels[y*width+x]
				img.SetRGBA(x, y, color.RGBA{
					R: uint8(packed >> 16),
					G: uint8(packed >> 8),
					B: uint8(packed),
					A: 255,
				})
			}
		}
	} else {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				n := v4FBM(float64(x), float64(y), seed)
				var c color.RGBA
				if n < 0.35 {
					t := n / 0.35
					c = v4LerpColor(pal.bgDark, pal.bgMid, t)
				} else {
					t := (n - 0.35) / 0.65
					c = v4LerpColor(pal.bgMid, pal.bgLight, t)
				}
				img.SetRGBA(x, y, c)
			}
		}
	}

	cx, cy := width/2, height/2

	// 2. Grid lines for advanced ages
	if pal.gridLines {
		v4DrawGridLines(img, pal)
	}

	// 2b. Scatter terrain features across the map (skip when background PNG is loaded)
	if bgPixels == nil {
		era := 0
		for i, p := range v4Palettes {
			if p.ageKey == state.Age {
				era = i * 6 / len(v4Palettes)
				break
			}
		}
		features, density := v4FeatureSetForEra(era)
		if len(features) > 0 {
			step := 24
			for gy := 0; gy < height-8; gy += step {
				for gx := 0; gx < width-8; gx += step {
					n := v4FBM(float64(gx+4), float64(gy+4), seed+99)
					if n > density {
						continue
					}
					ddx := gx + 4 - cx
					ddy := gy + 4 - cy
					if ddx*ddx+ddy*ddy < 40*40 {
						continue
					}
					if gx < width/5+10 {
						continue
					}
					fi := int(v4Hash(seed+77, uint32(gx*1000+gy))) % len(features)
					jx := int(v4Hash(seed+11, uint32(gx+gy*3))%uint32(step/2)) - step/4
					jy := int(v4Hash(seed+22, uint32(gx*2+gy))%uint32(step/2)) - step/4
					v4DrawTerrainFeature(img, features[fi], gx+jx, gy+jy)
				}
			}
		}
	}

	// 3. River (skip when background PNG is loaded — it already contains a river)
	if bgPixels == nil {
		v4DrawRiver(img, pal, seed)
	}

	// 4. Organic jittered-grid city layout — no road lines

	// Compute total buildings built for within-era density growth.
	totalBuilt := 0
	for _, bs := range state.Buildings {
		totalBuilt += bs.Count
	}
	era := sprites.GetEra(state.Age)
	density := v4CityDensity(era, totalBuilt)

	buildings := v4GetBuildings(state, density)

	// Generate candidate slots (offsets from center) sorted by distance ascending.
	allSlots := v4BuildingSlots(width, height, cx, cy, density.slotSpacing, density.jitterAmt)

	// Assign one slot per building, enforcing minimum spacing between placed sprites.
	assignedOffsets := v4AssignSlots(allSlots, len(buildings), density.minDist)

	// --- Auto-zoom: find max radius among assigned slot offsets ---
	type bldPos struct {
		bx, by int
		pixels [16][16]uint32
	}
	var positions []bldPos
	maxRadius := 0.0

	for i, b := range buildings {
		var ox, oy int
		if i < len(assignedOffsets) {
			ox = assignedOffsets[i][0]
			oy = assignedOffsets[i][1]
		}
		r := math.Sqrt(float64(ox*ox + oy*oy))
		if r > maxRadius {
			maxRadius = r
		}
		pixels := v4LoadBuildingSprite(b.key, b.domain, state.Age, b.isWonder, b.isStorage)
		positions = append(positions, bldPos{ox, oy, pixels})
	}

	// Scale so furthest building sits at 40% of half the shortest dimension.
	// Clamp between 0.15 and 0.5.
	targetRadius := float64(v4Min(width, height)) * 0.40
	scale := 0.8
	if maxRadius > 0 && len(buildings) > 0 {
		scale = targetRadius / (maxRadius + float64(v4SlotSpacing)) // padding
		if scale > 0.8 {
			scale = 0.8
		}
		if scale < 0.15 {
			scale = 0.15
		}
	}
	// Sprite size is fixed — decoupled from zoom scale so downscaling positions
	// never also shrinks sprites below their native 16×16 resolution.
	// density.spriteScale grows 1.0→1.6 across eras so late-game cities look bigger.
	spriteSize := int(math.Round(16 * density.spriteScale))
	if spriteSize < 16 {
		spriteSize = 16
	}
	clearingR := 12 // fixed small clearing halo (reduced from ~14*scale)
	if clearingR < 4 {
		clearingR = 4
	}

	// Draw clearings + sprites at scaled positions (no road lines)
	for _, pos := range positions {
		sbx := cx + int(math.Round(float64(pos.bx)*scale))
		sby := cy + int(math.Round(float64(pos.by)*scale))
		v4DrawClearingR(img, sbx, sby, pal.clearingColor, clearingR)
		v4DrawSpriteScaled(img, pos.pixels, sbx-spriteSize/2, sby-spriteSize/2, spriteSize)
	}

	// 5. Palace / city center — fixed 16px, small clearing so it doesn't dominate
	palaceSize := 16
	v4DrawClearingR(img, cx, cy, pal.clearingColor, palaceSize+2)
	v4DrawSpriteScaled(img, v4PalacePixels(state.Age), cx-palaceSize/2, cy-palaceSize/2, palaceSize)

	// 6. Wonder strip — separator line + wonder sprites across the bottom band.
	v4DrawWonderStrip(img, state, width, height)

	return img
}

// v4DrawWonderStrip draws a separator line and one 16×16 wonder sprite per built wonder
// evenly spaced across the bottom wonderStripH pixels of img.
func v4DrawWonderStrip(img *image.RGBA, state game.GameState, width, height int) {
	// Separator line at the top of the strip.
	sepY := height - wonderStripH - 1
	if sepY >= 0 {
		sepColor := color.RGBA{0x2a, 0x2a, 0x2a, 0xff}
		for x := 0; x < width; x++ {
			img.SetRGBA(x, sepY, sepColor)
		}
	}

	// Collect built wonders in stable key order.
	bldByKey := config.BuildingByKey()
	type builtWonder struct{ key string }
	var wonders []builtWonder
	for key, bs := range state.Buildings {
		if bs.Count == 0 {
			continue
		}
		if def, ok := bldByKey[key]; ok && def.Category == "wonder" {
			wonders = append(wonders, builtWonder{key})
		}
	}
	if len(wonders) == 0 {
		return
	}
	sort.Slice(wonders, func(i, j int) bool { return wonders[i].key < wonders[j].key })

	// Render each wonder sprite centred within its horizontal slot.
	n := len(wonders)
	xStep := width / n
	sy := height - wonderStripH + 2 // top-left y for 16×16 sprite

	for i, w := range wonders {
		sprite := v4LoadBuildingSprite(w.key, "", state.Age, true, false)
		sx := i*xStep + xStep/2 - 8
		for py := 0; py < 16; py++ {
			for px := 0; px < 16; px++ {
				packed := sprite[py][px]
				a := uint8(packed >> 24)
				// Sprites stored without alpha channel use 0 as transparent sentinel.
				if packed == 0 && a == 0 {
					continue
				}
				r := uint8(packed >> 16)
				g := uint8(packed >> 8)
				b := uint8(packed)
				dx := sx + px
				dy := sy + py
				if dx < 0 || dx >= width || dy < 0 || dy >= height {
					continue
				}
				img.SetRGBA(dx, dy, color.RGBA{r, g, b, 255})
			}
		}
	}
}

// WonderSpriteIcon returns a 2-character tview color-tagged thumbnail for the wonder
// identified by key, suitable for prepending to wonder names in overlay text views.
// Each character is a '▄' half-block: BG = upper sample pixel, FG = lower sample pixel.
func WonderSpriteIcon(key string) string {
	px := v4WonderSpriteByKey(key)
	c := func(row, col int) (r, g, b uint8) {
		packed := px[row][col]
		return uint8(packed >> 16), uint8(packed >> 8), uint8(packed)
	}
	r1u, g1u, b1u := c(4, 4)
	r1l, g1l, b1l := c(12, 4)
	r2u, g2u, b2u := c(4, 12)
	r2l, g2l, b2l := c(12, 12)
	ch1 := fmt.Sprintf("[#%02x%02x%02x:#%02x%02x%02x]▄[-:-]", r1l, g1l, b1l, r1u, g1u, b1u)
	ch2 := fmt.Sprintf("[#%02x%02x%02x:#%02x%02x%02x]▄[-:-]", r2l, g2l, b2l, r2u, g2u, b2u)
	return ch1 + ch2
}

// ── Build / Refresh ────────────────────────────────────────────────────────

// Build returns a tview.Primitive that renders the city map using half-block characters.
func (m *MapV4) Build(state game.GameState) tview.Primitive {
	m.mu.Lock()
	m.state = state
	m.mu.Unlock()

	box := tview.NewBox().SetBorder(false)
	box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		if width <= 0 || height <= 0 {
			return x, y, width, height
		}

		m.mu.Lock()
		st := m.state
		m.mu.Unlock()

		imgW := width
		imgH := height * 2

		bldCount := v4CountBuildings(st)
		m.mu.Lock()
		if m.cachedAge != st.Age {
			v4SpriteCacheMu.Lock()
			v4SpriteCache = map[string][16][16]uint32{}
			v4SpriteCacheMu.Unlock()
			v4BgCacheMu.Lock()
			v4BgCache = map[string][]uint32{}
			v4BgCacheMu.Unlock()
		}
		needRegen := m.cachedImg == nil || m.cachedW != imgW || m.cachedH != imgH ||
			m.cachedAge != st.Age || m.cachedBldCount != bldCount
		if needRegen {
			m.cachedImg = v4GenerateImage(st, imgW, imgH)
			m.cachedW = imgW
			m.cachedH = imgH
			m.cachedAge = st.Age
			m.cachedBldCount = bldCount
		}
		img := m.cachedImg
		m.mu.Unlock()

		renderHalfBlock(screen, img, x, y, width, height)
		return x, y, width, height
	})
	return box
}

// renderHalfBlock renders img onto screen using half-block characters (▄).
// Each terminal row represents 2 pixel rows: upper pixel = BG color, lower = FG color.
func renderHalfBlock(screen tcell.Screen, img *image.RGBA, offX, offY, termW, termH int) {
	for row := 0; row < termH; row++ {
		for col := 0; col < termW; col++ {
			upperC := img.RGBAAt(col, row*2)
			lowerC := img.RGBAAt(col, row*2+1)
			bgColor := tcell.NewRGBColor(int32(upperC.R), int32(upperC.G), int32(upperC.B))
			fgColor := tcell.NewRGBColor(int32(lowerC.R), int32(lowerC.G), int32(lowerC.B))
			screen.SetContent(offX+col, offY+row, '▄', nil,
				tcell.StyleDefault.Background(bgColor).Foreground(fgColor))
		}
	}
}

// Refresh updates the stored game state. Safe to call from any goroutine.
func (m *MapV4) Refresh(state game.GameState) {
	m.mu.Lock()
	m.state = state
	m.mu.Unlock()
}

func v4CountBuildings(state game.GameState) int {
	n := 0
	for _, bs := range state.Buildings {
		if bs.Count > 0 {
			n++
		}
	}
	return n
}
