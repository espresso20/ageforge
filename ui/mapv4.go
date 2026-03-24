package ui

import (
	"image"
	"image/color"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/espresso20/ageforge/game"
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
	{"primitive_age", v4rgb(0x1a5c0a), v4rgb(0x2d8c1a), v4rgb(0x4ab02a), v4rgb(0x2060c0), v4rgb(0x3da02a), false},
	{"stone_age", v4rgb(0x5a7a3a), v4rgb(0x7a9a4a), v4rgb(0x9ab85a), v4rgb(0x1a50a0), v4rgb(0x8aaa5a), false},
	{"bronze_age", v4rgb(0x6a7a2a), v4rgb(0x8a9a3a), v4rgb(0xaaba4a), v4rgb(0x1a50a0), v4rgb(0x9aaa4a), false},
	{"iron_age", v4rgb(0x4a6a2a), v4rgb(0x6a8a3a), v4rgb(0x8aaa4a), v4rgb(0x1a50a0), v4rgb(0x7a9a4a), false},
	{"classical_age", v4rgb(0x4a6030), v4rgb(0x6a8040), v4rgb(0x8aa050), v4rgb(0x1a4890), v4rgb(0x7a9050), false},
	{"medieval_age", v4rgb(0x3a5828), v4rgb(0x5a7838), v4rgb(0x7a9848), v4rgb(0x1a4080), v4rgb(0x6a8848), false},
	{"renaissance_age", v4rgb(0x384828), v4rgb(0x506838), v4rgb(0x688848), v4rgb(0x1a3870), v4rgb(0x607848), false},
	{"age_of_sail", v4rgb(0x304020), v4rgb(0x486030), v4rgb(0x607840), v4rgb(0x1a3868), v4rgb(0x587040), false},
	{"industrial_age", v4rgb(0x404038), v4rgb(0x585850), v4rgb(0x707068), v4rgb(0x183060), v4rgb(0x686860), false},
	{"gilded_age", v4rgb(0x383830), v4rgb(0x505048), v4rgb(0x686860), v4rgb(0x182858), v4rgb(0x606058), false},
	{"modern_age", v4rgb(0x282830), v4rgb(0x383840), v4rgb(0x484850), v4rgb(0x182050), v4rgb(0x484855), false},
	{"atomic_age", v4rgb(0x201820), v4rgb(0x302830), v4rgb(0x403840), v4rgb(0x201848), v4rgb(0x404048), false},
	{"space_age", v4rgb(0x100820), v4rgb(0x201830), v4rgb(0x302840), v4rgb(0x301060), v4rgb(0x302840), true},
	{"information_age", v4rgb(0x080818), v4rgb(0x181828), v4rgb(0x282838), v4rgb(0x280a58), v4rgb(0x282838), true},
	{"cyberpunk_age", v4rgb(0x060610), v4rgb(0x101020), v4rgb(0x1a1a30), v4rgb(0x5010a0), v4rgb(0x1a1a35), true},
	{"nanotech_age", v4rgb(0x040410), v4rgb(0x0c0c1c), v4rgb(0x14142c), v4rgb(0x4010a0), v4rgb(0x141430), true},
	{"fusion_age", v4rgb(0x040408), v4rgb(0x080818), v4rgb(0x101024), v4rgb(0x6010c0), v4rgb(0x10102a), true},
	{"singularity_age", v4rgb(0x020208), v4rgb(0x060614), v4rgb(0x0c0c20), v4rgb(0x7010d0), v4rgb(0x0c0c24), true},
	{"galactic_age", v4rgb(0x020206), v4rgb(0x040410), v4rgb(0x08081a), v4rgb(0x8010e0), v4rgb(0x08081e), true},
	{"cosmic_age", v4rgb(0x010106), v4rgb(0x03030e), v4rgb(0x060616), v4rgb(0x9010f0), v4rgb(0x060618), true},
	{"transcendent_age", v4rgb(0x010104), v4rgb(0x02020a), v4rgb(0x040410), v4rgb(0xa010ff), v4rgb(0x040412), true},
	{"divine_age", v4rgb(0x010102), v4rgb(0x020208), v4rgb(0x03030c), v4rgb(0xb020ff), v4rgb(0x03030e), true},
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

// v4PalacePixels returns a 16×16 palace sprite appropriate for the current age era.
func v4PalacePixels(era int) [16][16]uint32 {
	// Each era is a distinct palace design drawn inline
	switch era {
	case 0: // Primitive: chief's longhouse
		return [16][16]uint32{
			{0, 0, 0, 0, 0, 0x5a3a20, 0x5a3a20, 0x5a3a20, 0x5a3a20, 0x5a3a20, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0x5a3a20, 0x8b6a40, 0x8b6a40, 0x8b6a40, 0x8b6a40, 0x8b6a40, 0x5a3a20, 0, 0, 0, 0, 0},
			{0, 0, 0, 0x5a3a20, 0x8b6a40, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6a40, 0x5a3a20, 0, 0, 0, 0},
			{0, 0, 0x5a3a20, 0x8b6a40, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6a40, 0x5a3a20, 0, 0, 0, 0},
			{0, 0x5a3a20, 0x8b6a40, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6a40, 0x5a3a20, 0, 0, 0},
			{0x5a3a20, 0x8b6a40, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6a40, 0x5a3a20, 0, 0},
			{0x8b6040, 0xc8a060, 0xc8a060, 0xc8a060, 0x5a3a20, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x5a3a20, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6040, 0, 0},
			{0x8b6040, 0xc8a060, 0xc8a060, 0xc8a060, 0x5a3a20, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x5a3a20, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6040, 0, 0},
			{0x8b6040, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x5a3a20, 0x5a3a20, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6040, 0, 0},
			{0x8b6040, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x5a3a20, 0x5a3a20, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6040, 0, 0},
			{0x8b6040, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6040, 0, 0},
			{0x8b6040, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6040, 0, 0},
			{0x5a3a20, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x5a3a20, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
	case 1: // Classical/Medieval: stone castle
		return [16][16]uint32{
			{0, 0, 0, 0, 0, 0xd4a017, 0, 0, 0, 0xd4a017, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0xd4a017, 0xd4a017, 0xd4a017, 0, 0xd4a017, 0xd4a017, 0xd4a017, 0, 0, 0, 0, 0},
			{0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0, 0},
			{0, 0x908070, 0xd0c8b0, 0x908070, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0x908070, 0xd0c8b0, 0x908070, 0, 0},
			{0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xa0c0d0, 0xa0c0d0, 0xd0c8b0, 0xa0c0d0, 0xa0c0d0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0, 0},
			{0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xa0c0d0, 0xa0c0d0, 0xd0c8b0, 0xa0c0d0, 0xa0c0d0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0, 0},
			{0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0, 0},
			{0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd4a017, 0xd4a017, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0, 0},
			{0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd4a017, 0xd4a017, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0, 0},
			{0x8b1a1a, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0x8b1a1a, 0},
			{0x8b1a1a, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0x6b3a1a, 0x6b3a1a, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0x8b1a1a, 0},
			{0x8b1a1a, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0x6b3a1a, 0x6b3a1a, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0x8b1a1a, 0},
			{0x8b1a1a, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0x8b1a1a, 0},
			{0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
	case 2: // Industrial: Victorian town hall with clock tower
		return [16][16]uint32{
			{0, 0, 0, 0, 0, 0, 0, 0xc87941, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0x4a3a30, 0x8a6050, 0xc0c0a0, 0x8a6050, 0x4a3a30, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0x8a6050, 0x8a6050, 0xc0c0a0, 0x8a6050, 0x8a6050, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0, 0, 0, 0, 0, 0},
			{0, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0, 0},
			{0, 0x8a6050, 0xc0c0a0, 0xc0c0a0, 0x8a6050, 0x8a6050, 0xc0c0a0, 0xc0c0a0, 0xc0c0a0, 0x8a6050, 0x8a6050, 0xc0c0a0, 0xc0c0a0, 0x8a6050, 0, 0},
			{0, 0x8a6050, 0xc0c0a0, 0xc0c0a0, 0x8a6050, 0x8a6050, 0xc0c0a0, 0xc0c0a0, 0xc0c0a0, 0x8a6050, 0x8a6050, 0xc0c0a0, 0xc0c0a0, 0x8a6050, 0, 0},
			{0, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0, 0},
			{0, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x3a2a20, 0x3a2a20, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0, 0},
			{0, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x3a2a20, 0x3a2a20, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0, 0},
			{0, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0, 0},
			{0, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
	case 3: // Modern: glass capitol tower
		return [16][16]uint32{
			{0, 0, 0, 0, 0, 0, 0, 0xc0c8d0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0x808890, 0x80c0e0, 0x808890, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0x808890, 0x80c0e0, 0xe0f0ff, 0x80c0e0, 0x808890, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0x808890, 0x80c0e0, 0x80c0e0, 0xe0f0ff, 0x80c0e0, 0x80c0e0, 0x808890, 0, 0, 0, 0, 0},
			{0, 0, 0x505860, 0x505860, 0x808890, 0x808890, 0x808890, 0x808890, 0x808890, 0x808890, 0x808890, 0x505860, 0x505860, 0, 0, 0},
			{0, 0x505860, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x808890, 0x808890, 0x808890, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x505860, 0, 0},
			{0, 0x505860, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x808890, 0x808890, 0x808890, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x505860, 0, 0},
			{0, 0x505860, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x404850, 0x404850, 0x808890, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x505860, 0, 0},
			{0, 0x505860, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x404850, 0x404850, 0x808890, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x505860, 0, 0},
			{0, 0x505860, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x808890, 0x808890, 0x808890, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x505860, 0, 0},
			{0, 0x505860, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x808890, 0x808890, 0x808890, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x505860, 0, 0},
			{0, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0, 0},
			{0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
	case 4: // Digital/Cyber: neon megaplex
		return [16][16]uint32{
			{0, 0, 0, 0, 0, 0, 0, 0x40ffff, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0x101020, 0x40c0ff, 0x101020, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0x202030, 0x40c0ff, 0x40ffff, 0x40c0ff, 0x202030, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0x202030, 0x40c0ff, 0x202030, 0x202030, 0x202030, 0x40c0ff, 0x202030, 0, 0, 0, 0, 0},
			{0, 0, 0x101020, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x101020, 0, 0, 0},
			{0, 0x101020, 0x202030, 0x40c0ff, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x40c0ff, 0x202030, 0x101020, 0, 0},
			{0, 0x101020, 0x202030, 0x40c0ff, 0x202030, 0x40c0ff, 0x40c0ff, 0xff40c0, 0x40c0ff, 0x40c0ff, 0x202030, 0x40c0ff, 0x202030, 0x101020, 0, 0},
			{0, 0x101020, 0x202030, 0x40c0ff, 0x202030, 0x40c0ff, 0x40c0ff, 0xff40c0, 0x40c0ff, 0x40c0ff, 0x202030, 0x40c0ff, 0x202030, 0x101020, 0, 0},
			{0, 0x101020, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x101020, 0, 0},
			{0, 0x101020, 0x202030, 0x202030, 0x202030, 0x202030, 0x181828, 0x181828, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x101020, 0, 0},
			{0, 0x101020, 0x202030, 0x202030, 0x202030, 0x202030, 0x181828, 0x181828, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x101020, 0, 0},
			{0, 0x101020, 0x40c0ff, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x40c0ff, 0x101020, 0, 0},
			{0x40c0ff, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x40c0ff, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
	case 5: // Fusion/Space: orbital command spire
		return [16][16]uint32{
			{0, 0, 0, 0, 0, 0, 0, 0x80c0ff, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0x4080ff, 0xffe040, 0x4080ff, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0x0a1530, 0x4080ff, 0x80c0ff, 0x4080ff, 0x0a1530, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0x0a1530, 0x4080ff, 0x0a1530, 0x80c0ff, 0x0a1530, 0x4080ff, 0x0a1530, 0, 0, 0, 0, 0},
			{0, 0, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x80c0ff, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0, 0, 0},
			{0, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x0a1530, 0x80c0ff, 0x0a1530, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0, 0},
			{0, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x4080ff, 0x4080ff, 0xffe040, 0x4080ff, 0x4080ff, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0, 0},
			{0, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x4080ff, 0x4080ff, 0xffe040, 0x4080ff, 0x4080ff, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0, 0},
			{0, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x0a1530, 0x80c0ff, 0x0a1530, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0, 0},
			{0, 0x0a1530, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x80c0ff, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x0a1530, 0, 0},
			{0, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x80c0ff, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0, 0},
			{0, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x4080ff, 0x4080ff, 0x4080ff, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0, 0},
			{0x4080ff, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x4080ff, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
	default: // Era 6: Cosmic — glowing transcendent pillar
		return [16][16]uint32{
			{0, 0, 0, 0, 0, 0, 0, 0xff80ff, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0xffd040, 0xff80ff, 0xffd040, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0xc0a0ff, 0xff80ff, 0xffffff, 0xff80ff, 0xc0a0ff, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0xc0a0ff, 0xffd040, 0x080810, 0xffffff, 0x080810, 0xffd040, 0xc0a0ff, 0, 0, 0, 0, 0},
			{0, 0, 0, 0xc0a0ff, 0x080810, 0x080810, 0x080810, 0xffffff, 0x080810, 0x080810, 0x080810, 0xc0a0ff, 0, 0, 0, 0},
			{0, 0, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0xc0a0ff, 0xffffff, 0xc0a0ff, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0, 0, 0},
			{0, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0xffd040, 0xffd040, 0xffffff, 0xffd040, 0xffd040, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0, 0},
			{0, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0xffd040, 0xffd040, 0xffffff, 0xffd040, 0xffd040, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0, 0},
			{0, 0, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0xc0a0ff, 0xffffff, 0xc0a0ff, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0, 0, 0},
			{0, 0, 0, 0xc0a0ff, 0x080810, 0x080810, 0x080810, 0xffffff, 0x080810, 0x080810, 0x080810, 0xc0a0ff, 0, 0, 0, 0},
			{0, 0, 0, 0, 0xc0a0ff, 0xffd040, 0x080810, 0xffffff, 0x080810, 0xffd040, 0xc0a0ff, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0xc0a0ff, 0xff80ff, 0xffffff, 0xff80ff, 0xc0a0ff, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0xffd040, 0xffd040, 0xffd040, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
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

var (
	v4SpriteCacheMu sync.RWMutex
	v4SpriteCache   = map[string][16][16]uint32{}
)

// v4LoadBuildingSprite loads a 16×16 sprite from assets/sprites/buildings/<key>.png.
// Falls back to domain sprite if file not found. Results are cached.
func v4LoadBuildingSprite(key, domain string, isWonder, isStorage bool) [16][16]uint32 {
	v4SpriteCacheMu.RLock()
	if px, ok := v4SpriteCache[key]; ok {
		v4SpriteCacheMu.RUnlock()
		return px
	}
	v4SpriteCacheMu.RUnlock()

	// Try loading from disk
	path := filepath.Join("assets", "sprites", "buildings", key+".png")
	f, err := os.Open(path)
	if err == nil {
		defer f.Close()
		img, _, err := image.Decode(f)
		if err == nil {
			var px [16][16]uint32
			for row := 0; row < 16; row++ {
				for col := 0; col < 16; col++ {
					r, g, b, a := img.At(col, row).RGBA()
					if a > 0x8000 {
						px[row][col] = (uint32(r>>8) << 16) | (uint32(g>>8) << 8) | uint32(b>>8)
					}
				}
			}
			v4SpriteCacheMu.Lock()
			v4SpriteCache[key] = px
			v4SpriteCacheMu.Unlock()
			return px
		}
	}

	// Fallback to inline domain sprites
	var fallback [16][16]uint32
	if isWonder {
		fallback = v4SpriteWonderPixels
	} else if isStorage {
		fallback = v4SpriteStoragePixels
	} else if sp, ok := v4DomainSprites[domain]; ok {
		fallback = sp
	} else {
		fallback = v4SpriteCityPixels
	}
	v4SpriteCacheMu.Lock()
	v4SpriteCache[key] = fallback
	v4SpriteCacheMu.Unlock()
	return fallback
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
			img.SetRGBA(px+dstCol, py+dstRow, c)
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

// ── Road renderer ──────────────────────────────────────────────────────────

func v4DrawRoad(img *image.RGBA, x0, y0, x1, y1 int, roadClr color.RGBA) {
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	bounds := img.Bounds()
	for {
		if x0 >= bounds.Min.X && x0 < bounds.Max.X && y0 >= bounds.Min.Y && y0 < bounds.Max.Y {
			img.SetRGBA(x0, y0, roadClr)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// v4BuildingJitter returns a stable small perpendicular offset (-5..+5) for a building key.
// This breaks the perfectly-straight-line look without being random each frame.
func v4BuildingJitter(key string) int {
	h := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		h ^= uint32(key[i])
		h *= 16777619
	}
	return int(h%11) - 5 // -5 to +5
}

// ── Building list ──────────────────────────────────────────────────────────

type v4Building struct {
	key       string
	domain    string
	isWonder  bool
	isStorage bool
	count     int
}

func v4GetBuildings(state game.GameState) []v4Building {
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
		if len(key) >= 6 && key[len(key)-6:] == "wonder" {
			isWonder = true
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
	// Count total buildings built across all types
	totalBuilt := 0
	for _, bs := range state.Buildings {
		totalBuilt += bs.Count
	}
	// Show 1 sprite per 3 buildings built, minimum 1, max 200
	maxSprites := totalBuilt/3 + 1
	if maxSprites > 200 {
		maxSprites = 200
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].isWonder != result[j].isWonder {
			return result[i].isWonder
		}
		if result[i].domain != result[j].domain {
			return result[i].domain < result[j].domain
		}
		return result[i].key < result[j].key
	})
	if len(result) > maxSprites {
		result = result[:maxSprites]
	}
	return result
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
	G := uint32(0x1a6010)
	M := uint32(0x2d8c1a)
	L := uint32(0x4ab02a)
	T := uint32(0x5a3a20)
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
	D := uint32(0x1a4a10)
	M := uint32(0x2d6b1a)
	T := uint32(0x5a3a20)
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
	L := uint32(0xaaaaaa)
	G := uint32(0x7a7a7a)
	D := uint32(0x4a4a4a)
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
		return []v4TerrainFeature{v4FeatureJungleTree(), v4FeatureRock()}, 0.18
	case 1:
		return []v4TerrainFeature{v4FeaturePineTree(), v4FeatureCropField()}, 0.14
	case 2:
		return []v4TerrainFeature{v4FeatureCropField(), v4FeatureCityBlock()}, 0.12
	case 3:
		return []v4TerrainFeature{v4FeatureCityBlock(), v4FeatureCityBlock()}, 0.15
	case 4:
		return []v4TerrainFeature{v4FeatureNeonBlock(), v4FeatureCityBlock()}, 0.12
	case 5:
		return []v4TerrainFeature{v4FeatureEnergyCrystal(), v4FeatureEnergyCrystal()}, 0.10
	default:
		return []v4TerrainFeature{v4FeatureVoidParticle()}, 0.08
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

	// 1. FBM terrain background
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

	cx, cy := width/2, height/2

	// 2. Grid lines for advanced ages
	if pal.gridLines {
		v4DrawGridLines(img, pal)
	}

	// 2b. Scatter terrain features across the map
	{
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
					if ddx*ddx+ddy*ddy < 60*60 {
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

	// 3. River
	v4DrawRiver(img, pal, seed)

	// 4. City roads + buildings
	buildings := v4GetBuildings(state)
	roadClr := v4LerpColor(pal.bgMid, pal.bgLight, 0.4)

	baseAngles := []float64{0, 48, 93, 145, 198, 252, 310}
	numRoads := len(baseAngles)
	roadSpacing := 22.0 // base spacing at scale 1.0

	// --- Auto-zoom: compute all positions at scale 1.0, find max radius ---
	type bldPos struct {
		bx, by int
		pixels [16][16]uint32
	}
	var positions []bldPos
	maxRadius := 0.0

	for i, b := range buildings {
		roadIdx := i % numRoads
		slot := i/numRoads + 1
		deg := baseAngles[roadIdx]
		jitter := v4BuildingJitter(b.key)
		rad := deg * math.Pi / 180
		perpRad := rad + math.Pi/2
		dist := float64(slot) * roadSpacing
		bx := int(math.Round(dist*math.Cos(rad))) + int(math.Round(float64(jitter)*math.Cos(perpRad)))
		by := int(math.Round(dist*math.Sin(rad))) + int(math.Round(float64(jitter)*math.Sin(perpRad)))
		r := math.Sqrt(float64(bx*bx + by*by))
		if r > maxRadius {
			maxRadius = r
		}
		pixels := v4LoadBuildingSprite(b.key, b.domain, b.isWonder, b.isStorage)
		positions = append(positions, bldPos{bx, by, pixels})
	}

	// Scale so furthest building sits at 38% of half the shortest dimension,
	// leaving room for clearings and the city center sprite.
	// Minimum scale 0.20 (sprites become 3px), maximum 1.0.
	targetRadius := float64(v4Min(width, height)) * 0.38
	scale := 1.0
	if maxRadius > 0 && len(buildings) > 0 {
		scale = targetRadius / (maxRadius + roadSpacing) // +spacing for padding
		if scale > 1.0 {
			scale = 1.0
		}
		if scale < 0.20 {
			scale = 0.20
		}
	}
	spriteSize := int(math.Round(16 * scale))
	if spriteSize < 3 {
		spriteSize = 3
	}
	clearingR := int(math.Round(14 * scale))
	if clearingR < 4 {
		clearingR = 4
	}

	// Draw roads at scaled length
	if len(buildings) > 0 {
		maxSlot := (len(buildings)-1)/numRoads + 1
		for _, deg := range baseAngles {
			rad := deg * math.Pi / 180
			roadLen := (float64(maxSlot+1)*roadSpacing + 16) * scale
			ex := cx + int(math.Round(roadLen*math.Cos(rad)))
			ey := cy + int(math.Round(roadLen*math.Sin(rad)))
			v4DrawRoad(img, cx, cy, ex, ey, roadClr)
		}
	}

	// Draw clearings + sprites at scaled positions
	for _, pos := range positions {
		sbx := cx + int(math.Round(float64(pos.bx)*scale))
		sby := cy + int(math.Round(float64(pos.by)*scale))
		v4DrawClearingR(img, sbx, sby, pal.clearingColor, clearingR)
		v4DrawSpriteScaled(img, pos.pixels, sbx-spriteSize/2, sby-spriteSize/2, spriteSize)
	}

	// 5. Palace / city center — always at map center, evolves with age, drawn at 1.5× sprite size
	palaceSize := int(math.Round(float64(spriteSize) * 1.5))
	if palaceSize < 8 {
		palaceSize = 8
	}
	palaceEra := 0
	for i, p := range v4Palettes {
		if p.ageKey == state.Age {
			palaceEra = i * 6 / len(v4Palettes) // map palette index to era 0-6
			break
		}
	}
	v4DrawClearingR(img, cx, cy, pal.clearingColor, palaceSize+6)
	v4DrawSpriteScaled(img, v4PalacePixels(palaceEra), cx-palaceSize/2, cy-palaceSize/2, palaceSize)

	return img
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
