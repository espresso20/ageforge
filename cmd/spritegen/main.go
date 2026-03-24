// cmd/spritegen/main.go — generates 16×16 pixel art domain icons + sprite sheet + preview
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// rgba converts a packed 0xRRGGBB value to color.RGBA (alpha=255), or transparent if 0.
func rgba(v uint32) color.RGBA {
	if v == 0 {
		return color.RGBA{0, 0, 0, 0}
	}
	return color.RGBA{
		R: uint8((v >> 16) & 0xff),
		G: uint8((v >> 8) & 0xff),
		B: uint8(v & 0xff),
		A: 255,
	}
}

type sprite struct {
	name   string
	pixels [16][16]uint32
}

// ── Color palette ─────────────────────────────────────────────────────────────

// food: wheat bundle
const (
	fG = 0xe8c840 // grain
	fS = 0x6a4a20 // stem
	fB = 0x4a3010 // base
)

// wood: pine tree
const (
	wD = 0x2d6b1a // dark green
	wM = 0x3d8b2a // mid green
	wT = 0x6b3a1a // trunk
	wR = 0x8b5a2b // root
)

// stone: rocky cliff
const (
	stL = 0xaaaaaa // light gray
	stG = 0x7a7a7a // mid gray
	stD = 0x4a4a4a // dark gray
	stH = 0xcccccc // highlight
)

// iron: anvil
const (
	irI = 0x4a4a5a // dark iron
	irL = 0x7a7a8a // light iron
	irS = 0x2a2a3a // shadow
)

// knowledge: open book
const (
	kC = 0x8b4513 // brown cover
	kP = 0xf0f0e0 // page
	kL = 0xd0d0c0 // page shadow
	kG = 0xd4a017 // gold spine
)

// military: shield with sword
const (
	miS = 0xc0c0c0 // silver sword
	miB = 0x8b1a1a // red shield
	miE = 0xd4a017 // gold emblem
	miW = 0xf0f0f0 // sword edge
)

// faith: gothic church
const (
	faW = 0xd0c8b0 // stone wall
	faD = 0x908070 // dark stone
	faG = 0xd4a017 // gold cross
	faR = 0x8b1a1a // red door
)

// culture: theater masks
const (
	cuH = 0xf5c080 // happy mask
	cuS = 0x8080c0 // sad mask
	cuG = 0xd4a017 // gold trim
	cuK = 0x303030 // dark outline
)

// trade: merchant scales
const (
	trG = 0xd4a017 // gold
	trS = 0xc0c0c0 // silver pan
	trC = 0xe8c840 // coin
	trB = 0x6b3a1a // brown post
)

// energy: lightning bolt / power tower
const (
	enY = 0xffe040 // yellow bolt
	enO = 0xff8c00 // orange glow
	enW = 0xffffff // white core
	enG = 0x888888 // gray tower
)

// metallurgy: large gear
const (
	meG = 0x888888 // gray
	meL = 0xbbbbbb // light
	meD = 0x555555 // dark
	meH = 0xdddddd // highlight
)

// advanced: blue crystal
const (
	adC = 0x40c0ff // cyan
	adB = 0x0080c0 // blue
	adW = 0xffffff // white
	adD = 0x004080 // dark blue
	adP = 0x80e0ff // pale
)

// city_center: medieval castle keep
const (
	ccW = 0xd0c8b0 // wall
	ccD = 0x908070 // dark stone
	ccG = 0xd4a017 // gold flag
	ccR = 0x8b1a1a // red flag
	ccB = 0x303030 // battlements
)

// wonder: pyramid
const (
	wnS = 0xc8a840 // sandstone
	wnL = 0xe8c860 // light face
	wnD = 0xa08030 // shadow
	wnG = 0xd4a017 // gold cap
)

// storage: wooden barn
const (
	stW = 0x8b5a2b // wood brown
	stLw = 0xc87941 // light wood
	stR = 0x8b1a1a // red roof
	stDw = 0x5a3010 // dark wood
	stO = 0xd4a017 // gold hinge
)

// ── Sprite definitions ────────────────────────────────────────────────────────

func spriteFood() [16][16]uint32 {
	_ = fB // used in rows 13-14
	return [16][16]uint32{
		{0, 0, 0, fG, 0, 0, fG, 0, 0, fG, 0, 0, 0, 0, 0, 0},
		{0, 0, fG, fG, fG, 0, fG, fG, 0, fG, fG, 0, 0, 0, 0, 0},
		{0, 0, fG, fG, fG, fG, fG, fG, fG, fG, fG, 0, 0, 0, 0, 0},
		{0, 0, 0, fG, fG, fG, fG, fG, fG, fG, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, fG, fG, fG, fG, fG, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, fS, 0, fS, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, fS, fS, fS, fS, fS, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, fS, fS, fS, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, fS, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, fS, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, fS, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, fS, fS, fS, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, fS, fS, fS, fS, fS, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, fB, fB, fB, fB, fB, fB, fB, 0, 0, 0, 0, 0, 0},
		{0, 0, fB, fB, fB, fB, fB, fB, fB, fB, fB, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func spriteWood() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, wM, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, wM, wM, wM, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, wM, wM, wM, wM, wM, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, wD, wM, wM, wM, wM, wM, wD, 0, 0, 0, 0, 0},
		{0, 0, 0, wD, wD, wM, wM, wM, wM, wM, wD, wD, 0, 0, 0, 0},
		{0, 0, 0, 0, wM, wM, wM, wM, wM, wM, wM, wM, wM, 0, 0, 0},
		{0, 0, wM, wM, wM, wM, wM, wM, wM, wM, wM, wM, wM, wM, 0, 0},
		{0, wM, wM, wM, wM, wD, wM, wM, wM, wM, wD, wM, wM, wM, wM, 0},
		{wM, wM, wM, wM, wD, wD, wM, wM, wM, wM, wD, wD, wM, wM, wM, wM},
		{0, 0, 0, 0, 0, 0, 0, wT, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, wT, wT, wT, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, wT, wT, wT, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, wR, wT, wT, wT, wR, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, wR, wR, wT, wT, wT, wR, wR, 0, 0, 0, 0, 0},
		{0, 0, 0, wR, wR, wR, wR, wR, wR, wR, wR, wR, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func spriteStone() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, stG, stG, stG, stG, stG, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, stG, stG, stH, stH, stL, stG, stG, 0, 0, 0, 0, 0},
		{0, 0, 0, stG, stG, stH, stH, stL, stL, stL, stG, stG, 0, 0, 0, 0},
		{0, 0, stG, stG, stH, stH, stL, stL, stL, stL, stL, stG, stG, 0, 0, 0},
		{0, 0, stG, stG, stH, stL, stL, stG, stG, stL, stL, stG, stG, 0, 0, 0},
		{0, 0, stG, stG, stL, stL, stG, stD, stD, stG, stL, stG, stG, 0, 0, 0},
		{0, stG, stG, stL, stL, stG, stD, stD, stD, stD, stG, stL, stG, stG, 0, 0},
		{0, stG, stG, stL, stG, stD, stD, stD, stD, stD, stD, stG, stL, stG, stG, 0},
		{stG, stG, stL, stG, stD, stD, stD, stG, stG, stD, stD, stD, stG, stL, stG, stG},
		{stG, stL, stG, stD, stD, stG, stG, stG, stG, stG, stG, stD, stD, stG, stL, stG},
		{stG, stG, stD, stD, stG, stG, stG, stG, stG, stG, stG, stG, stD, stD, stG, stG},
		{stG, stD, stD, stG, stG, stG, stG, stG, stG, stG, stG, stG, stG, stD, stD, stG},
		{stD, stD, stG, stG, stG, stG, stG, stG, stG, stG, stG, stG, stG, stG, stD, stD},
		{stD, stG, stG, stG, stG, stG, stG, stG, stG, stG, stG, stG, stG, stG, stG, stD},
		{stD, stD, stD, stD, stD, stD, stD, stD, stD, stD, stD, stD, stD, stD, stD, stD},
	}
}

func spriteIron() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, irI, irI, irI, irI, irI, irI, irI, 0, 0, 0, 0, 0},
		{0, 0, 0, irI, irI, irL, irL, irL, irL, irI, irI, irI, 0, 0, 0, 0},
		{0, 0, 0, irI, irI, irL, irL, irL, irL, irI, irI, irI, 0, 0, 0, 0},
		{0, 0, 0, 0, irI, irI, irI, irI, irI, irI, irI, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, irI, irI, irI, irI, irI, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, irI, irI, irI, irI, irI, irI, irI, 0, 0, 0, 0, 0},
		{0, 0, 0, irI, irI, irI, irI, irI, irI, irI, irI, irI, 0, 0, 0, 0},
		{0, 0, irI, irI, irI, irL, irL, irL, irL, irI, irI, irI, irI, 0, 0, 0},
		{0, 0, irI, irI, irL, irL, irL, irL, irL, irL, irI, irI, irI, 0, 0, 0},
		{0, 0, irI, irI, irL, irL, irL, irL, irL, irL, irI, irI, irI, 0, 0, 0},
		{0, 0, irI, irI, irI, irI, irI, irI, irI, irI, irI, irI, irI, 0, 0, 0},
		{0, 0, 0, irS, irI, irI, 0, 0, 0, irI, irI, irS, 0, 0, 0, 0},
		{0, 0, irS, irS, irI, irI, irI, irI, irI, irI, irI, irS, irS, 0, 0, 0},
		{0, irS, irS, irS, irI, irI, irI, irI, irI, irI, irI, irS, irS, irS, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func spriteKnowledge() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, kC, kC, kC, kC, kC, kG, kC, kC, kC, kC, kC, kC, 0, 0},
		{0, kC, kC, kP, kP, kP, kC, kG, kC, kP, kP, kP, kP, kC, kC, 0},
		{0, kC, kP, kP, kP, kP, kC, kG, kC, kP, kP, kP, kP, kP, kC, 0},
		{0, kC, kP, kP, kP, kP, kC, kG, kC, kP, kP, kP, kP, kP, kC, 0},
		{0, kC, kP, kP, kL, kP, kC, kG, kC, kP, kP, kL, kP, kP, kC, 0},
		{0, kC, kP, kP, kL, kP, kC, kG, kC, kP, kP, kL, kP, kP, kC, 0},
		{0, kC, kP, kP, kL, kP, kC, kG, kC, kP, kP, kL, kP, kP, kC, 0},
		{0, kC, kP, kP, kL, kP, kC, kG, kC, kP, kP, kL, kP, kP, kC, 0},
		{0, kC, kP, kP, kP, kP, kC, kG, kC, kP, kP, kP, kP, kP, kC, 0},
		{0, kC, kP, kP, kP, kP, kC, kG, kC, kP, kP, kP, kP, kP, kC, 0},
		{0, kC, kC, kC, kC, kC, kC, kG, kC, kC, kC, kC, kC, kC, kC, 0},
		{0, 0, kC, kC, kC, kC, kC, kG, kC, kC, kC, kC, kC, 0, 0, 0},
		{0, 0, 0, kC, kC, kC, kC, kG, kC, kC, kC, kC, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, kG, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func spriteMilitary() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, miW, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, miW, miS, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, miW, miS, miS, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, miB, miB, miB, miW, miS, miS, miB, miB, miB, 0, 0, 0, 0, 0},
		{0, miB, miB, miB, miB, miB, miS, miS, miB, miB, miB, miB, 0, 0, 0, 0},
		{0, miB, miB, miB, miB, miB, miS, miS, miB, miB, miB, miB, 0, 0, 0, 0},
		{0, miB, miB, miB, miB, miE, miE, miE, miB, miB, miB, miB, 0, 0, 0, 0},
		{0, miB, miB, miB, miB, miE, miE, miE, miB, miB, miB, miB, 0, 0, 0, 0},
		{0, miB, miB, miB, miB, miE, miE, miE, miB, miB, miB, miB, 0, 0, 0, 0},
		{0, miB, miB, miB, miB, miB, miB, miB, miB, miB, miB, miB, 0, 0, 0, 0},
		{0, 0, miB, miB, miB, miB, miB, miB, miB, miB, miB, 0, 0, 0, 0, 0},
		{0, 0, 0, miB, miB, miB, miB, miB, miB, miB, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, miB, miB, miB, miB, miB, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, miB, miB, miB, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, miB, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func spriteFaith() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, faG, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, faG, faG, faG, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, faG, faG, faG, faG, faG, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, faW, faG, faG, faG, faW, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, faW, faW, faG, faW, faG, faW, faW, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, faW, faW, faW, faW, faW, faW, faW, 0, 0, 0, 0, 0},
		{0, 0, 0, faW, faW, faD, faW, faW, faW, faD, faW, faW, 0, 0, 0, 0},
		{0, 0, 0, faW, faW, faD, faW, faW, faW, faD, faW, faW, 0, 0, 0, 0},
		{0, 0, faW, faW, faW, faW, faW, faW, faW, faW, faW, faW, faW, 0, 0, 0},
		{0, 0, faW, faW, faW, faW, faW, faW, faW, faW, faW, faW, faW, 0, 0, 0},
		{0, 0, faW, faW, faD, faW, faW, faW, faW, faW, faD, faW, faW, 0, 0, 0},
		{0, 0, faW, faW, faD, faW, faW, faR, faR, faW, faD, faW, faW, 0, 0, 0},
		{0, 0, faW, faW, faW, faW, faW, faR, faR, faW, faW, faW, faW, 0, 0, 0},
		{0, 0, faW, faW, faW, faW, faW, faR, faR, faW, faW, faW, faW, 0, 0, 0},
		{0, faW, faW, faW, faW, faW, faW, faW, faW, faW, faW, faW, faW, faW, 0, 0},
		{faW, faW, faW, faW, faW, faW, faW, faW, faW, faW, faW, faW, faW, faW, faW, faW},
	}
}

func spriteCulture() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, cuG, cuG, cuG, cuG, 0, 0, 0, 0, cuG, cuG, cuG, cuG, cuG, 0, 0},
		{0, cuG, cuH, cuH, cuG, 0, 0, 0, cuG, cuG, cuS, cuS, cuS, cuG, cuG, 0},
		{cuG, cuH, cuH, cuH, cuH, cuG, 0, cuG, cuS, cuS, cuS, cuS, cuS, cuS, cuG, 0},
		{cuG, cuH, cuK, 0, cuK, cuH, cuG, cuG, cuS, cuK, 0, 0, cuK, cuS, cuG, 0},
		{cuG, cuH, cuH, cuH, cuH, cuH, cuG, cuG, cuS, cuS, cuS, cuS, cuS, cuS, cuG, 0},
		{cuG, cuH, cuK, cuK, cuK, cuH, cuG, cuG, cuS, cuK, cuK, cuK, cuS, cuS, cuG, 0},
		{cuG, cuH, cuH, cuK, cuH, cuH, cuG, cuG, cuS, cuS, cuK, cuS, cuS, cuS, cuG, 0},
		{0, cuG, cuH, cuH, cuH, cuG, 0, cuG, cuS, cuS, cuS, cuS, cuS, cuG, cuG, 0},
		{0, 0, cuG, cuH, cuG, 0, 0, 0, cuG, cuS, cuS, cuS, cuG, cuG, 0, 0},
		{0, 0, cuG, cuG, cuG, 0, 0, 0, 0, cuG, cuG, cuG, cuG, 0, 0, 0},
		{0, 0, 0, cuG, 0, 0, 0, 0, 0, 0, cuG, 0, 0, 0, 0, 0},
		{0, 0, cuG, cuG, cuG, cuG, cuG, cuG, cuG, cuG, cuG, cuG, 0, 0, 0, 0},
		{0, 0, 0, cuG, cuG, cuG, cuG, cuG, cuG, cuG, cuG, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func spriteTrade() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, trG, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, trG, trG, trG, trG, trG, trG, trG, trG, trG, trG, trG, trG, trG, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, trG, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, trG, 0, 0, 0, trG, trG, trG, 0, 0, 0, trG, 0, 0, 0},
		{0, trG, 0, trG, 0, trG, 0, trG, 0, trG, 0, trG, 0, trG, 0, 0},
		{trG, 0, 0, 0, trG, 0, 0, trG, 0, 0, trG, 0, 0, 0, trG, 0},
		{trS, trS, trS, trS, trS, trS, 0, trG, 0, trS, trS, trS, trS, trS, trS, 0},
		{trS, trC, trC, trC, trC, trS, 0, trB, 0, trS, trC, trC, trC, trC, trS, 0},
		{trS, trC, trC, trC, trC, trS, 0, trB, 0, trS, trC, trC, trC, trC, trS, 0},
		{trS, trS, trS, trS, trS, trS, 0, trB, 0, trS, trS, trS, trS, trS, trS, 0},
		{0, 0, 0, 0, 0, 0, 0, trB, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, trB, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, trB, trB, trB, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, trB, trB, trB, trB, trB, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, trB, trB, trB, trB, trB, trB, trB, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func spriteEnergy() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, enW, enW, enW, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, enW, enY, enY, enO, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, enW, enY, enY, enY, enO, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, enW, enY, enY, enY, enO, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, enW, enY, enY, enY, enY, enY, enY, enO, 0, 0, 0, 0, 0, 0},
		{0, enW, enY, enY, enY, enY, enY, enY, enY, enY, enO, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, enW, enY, enY, enY, enO, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, enW, enY, enO, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, enW, enO, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, enW, enY, enO, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, enW, enY, enY, enY, enO, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, enG, enG, enG, enG, enG, enG, enG, enG, enG, enG, 0, 0, 0, 0},
		{0, 0, 0, enG, enG, 0, 0, 0, 0, enG, enG, 0, 0, 0, 0, 0},
		{0, 0, 0, enG, enG, 0, 0, 0, 0, enG, enG, 0, 0, 0, 0, 0},
		{0, 0, enG, enG, enG, enG, 0, 0, enG, enG, enG, enG, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func spriteMetallurgy() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, meD, meD, meD, meD, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, meD, meG, meG, meG, meG, meD, 0, 0, 0, 0, 0, 0},
		{0, 0, meD, meD, meG, meL, meH, meL, meG, meG, meD, meD, 0, 0, 0, 0},
		{0, meD, meG, meG, meG, meH, meH, meL, meG, meG, meG, meG, meD, 0, 0, 0},
		{meD, meG, meG, meL, meH, meH, meH, meH, meH, meL, meG, meG, meG, meD, 0, 0},
		{meD, meG, meL, meH, meH, meH, meH, meH, meH, meH, meL, meG, meG, meD, 0, 0},
		{meD, meG, meG, meH, meH, meG, meG, meG, meG, meH, meH, meG, meG, meD, 0, 0},
		{meD, meG, meL, meH, meH, meG, meD, meD, meG, meH, meH, meL, meG, meD, 0, 0},
		{meD, meG, meL, meH, meH, meG, meD, meD, meG, meH, meH, meL, meG, meD, 0, 0},
		{meD, meG, meG, meH, meH, meG, meG, meG, meG, meH, meH, meG, meG, meD, 0, 0},
		{meD, meG, meL, meH, meH, meH, meH, meH, meH, meH, meL, meG, meG, meD, 0, 0},
		{meD, meG, meG, meL, meH, meH, meH, meH, meH, meL, meG, meG, meG, meD, 0, 0},
		{0, meD, meG, meG, meG, meH, meH, meL, meG, meG, meG, meG, meD, 0, 0, 0},
		{0, 0, meD, meD, meG, meL, meH, meL, meG, meG, meD, meD, 0, 0, 0, 0},
		{0, 0, 0, 0, meD, meG, meG, meG, meG, meD, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, meD, meD, meD, meD, 0, 0, 0, 0, 0, 0, 0},
	}
}

func spriteAdvanced() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, adW, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, adW, adC, adW, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, adW, adC, adC, adC, adW, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, adW, adP, adC, adW, adC, adC, adW, 0, 0, 0, 0, 0},
		{0, 0, 0, adW, adP, adP, adC, adW, adC, adC, adC, adW, 0, 0, 0, 0},
		{0, 0, adW, adP, adP, adB, adC, adW, adC, adB, adC, adC, adW, 0, 0, 0},
		{0, adW, adP, adP, adB, adB, adC, adW, adC, adB, adB, adC, adW, 0, 0, 0},
		{adW, adP, adP, adB, adB, adD, adC, adW, adC, adD, adB, adB, adC, adW, 0, 0},
		{0, adW, adP, adP, adB, adB, adC, adW, adC, adB, adB, adC, adW, 0, 0, 0},
		{0, 0, adW, adP, adP, adB, adC, adW, adC, adB, adC, adW, 0, 0, 0, 0},
		{0, 0, 0, adW, adP, adP, adC, adW, adC, adP, adW, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, adW, adP, adC, adW, adC, adW, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, adW, adC, adW, adC, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, adW, adC, adW, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, adW, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func spriteCityCenter() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, ccR, ccR, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, ccG, ccR, ccR, ccG, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, ccG, ccG, ccG, ccG, ccG, ccG, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{ccB, ccW, ccB, ccW, ccB, ccW, ccB, ccW, ccB, ccW, ccB, ccW, ccB, ccW, ccB, ccW},
		{ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW},
		{ccW, ccW, ccD, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccD, ccW, ccW, ccW, ccW},
		{ccW, ccW, ccD, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccD, ccW, ccW, ccW, ccW},
		{ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW},
		{ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW},
		{ccW, ccW, ccD, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccD, ccW, ccW, ccW, ccW},
		{ccW, ccW, ccD, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccD, ccW, ccW, ccW, ccW},
		{ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW},
		{ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW},
		{ccW, ccW, ccD, ccW, ccR, ccR, ccW, ccW, ccW, ccR, ccR, ccD, ccW, ccW, ccW, ccW},
		{ccW, ccW, ccD, ccW, ccR, ccR, ccW, ccW, ccW, ccR, ccR, ccD, ccW, ccW, ccW, ccW},
		{ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW, ccW},
	}
}

func spriteWonder() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, wnG, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, wnG, wnG, wnG, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, wnL, wnS, wnD, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, wnL, wnL, wnS, wnD, wnD, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, wnL, wnL, wnL, wnS, wnD, wnD, wnD, 0, 0, 0, 0, 0},
		{0, 0, 0, wnL, wnL, wnL, wnL, wnS, wnD, wnD, wnD, wnD, 0, 0, 0, 0},
		{0, 0, wnL, wnL, wnL, wnL, wnL, wnS, wnD, wnD, wnD, wnD, wnD, 0, 0, 0},
		{0, wnL, wnL, wnL, wnL, wnL, wnL, wnS, wnD, wnD, wnD, wnD, wnD, wnD, 0, 0},
		{wnL, wnL, wnL, wnL, wnL, wnL, wnL, wnS, wnD, wnD, wnD, wnD, wnD, wnD, wnD, 0},
		{wnL, wnL, wnL, wnL, wnL, wnL, wnL, wnS, wnD, wnD, wnD, wnD, wnD, wnD, wnD, 0},
		{wnL, wnL, wnL, wnL, wnL, wnL, wnL, wnS, wnD, wnD, wnD, wnD, wnD, wnD, wnD, 0},
		{wnS, wnS, wnS, wnS, wnS, wnS, wnS, wnS, wnS, wnS, wnS, wnS, wnS, wnS, wnS, 0},
		{wnD, wnD, wnD, wnD, wnD, wnD, wnD, wnD, wnD, wnD, wnD, wnD, wnD, wnD, wnD, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func spriteStorage() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, stR, stR, stR, stR, stR, stR, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, stR, stR, stR, stR, stR, stR, stR, stR, 0, 0, 0, 0},
		{0, 0, 0, stR, stR, stR, stR, stR, stR, stR, stR, stR, stR, 0, 0, 0},
		{0, 0, stR, stR, stR, stR, stR, stR, stR, stR, stR, stR, stR, stR, 0, 0},
		{0, stR, stR, stR, stR, stR, stR, stR, stR, stR, stR, stR, stR, stR, stR, 0},
		{0, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, 0},
		{0, stW, stLw, stW, stW, stW, stW, stW, stW, stW, stW, stW, stLw, stW, stW, 0},
		{0, stW, stLw, stW, stW, stW, stW, stW, stW, stW, stW, stW, stLw, stW, stW, 0},
		{0, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, 0},
		{0, stW, stW, stDw, stW, stW, stW, stW, stW, stW, stW, stDw, stW, stW, stW, 0},
		{0, stW, stW, stDw, stW, stO, stW, stW, stO, stW, stDw, stW, stW, stW, 0, 0},
		{0, stW, stW, stDw, stW, stW, stW, stW, stW, stW, stDw, stW, stW, stW, 0, 0},
		{0, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, 0, 0},
		{0, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, 0, 0},
		{0, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, 0, 0},
		{0, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, stW, 0, 0},
	}
}

// ── Sprite table ──────────────────────────────────────────────────────────────

var domainSprites = []sprite{
	{name: "food", pixels: spriteFood()},
	{name: "wood", pixels: spriteWood()},
	{name: "stone", pixels: spriteStone()},
	{name: "iron", pixels: spriteIron()},
	{name: "knowledge", pixels: spriteKnowledge()},
	{name: "military", pixels: spriteMilitary()},
	{name: "faith", pixels: spriteFaith()},
	{name: "culture", pixels: spriteCulture()},
	{name: "trade", pixels: spriteTrade()},
	{name: "energy", pixels: spriteEnergy()},
	{name: "metallurgy", pixels: spriteMetallurgy()},
	{name: "advanced", pixels: spriteAdvanced()},
	{name: "city_center", pixels: spriteCityCenter()},
	{name: "wonder", pixels: spriteWonder()},
	{name: "storage", pixels: spriteStorage()},
}

// ── Image helpers ─────────────────────────────────────────────────────────────

func spriteToImage(s sprite) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			c := rgba(s.pixels[y][x])
			img.SetNRGBA(x, y, color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A})
		}
	}
	return img
}

func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// buildSpriteSheet places all sprites in a single row: width = n×16, height = 16.
func buildSpriteSheet(imgs []*image.NRGBA) *image.NRGBA {
	n := len(imgs)
	sheet := image.NewNRGBA(image.Rect(0, 0, n*16, 16))
	for i, img := range imgs {
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				sheet.SetNRGBA(i*16+x, y, img.NRGBAAt(x, y))
			}
		}
	}
	return sheet
}

const (
	previewScale  = 6  // each pixel → 6×6 → 96px per sprite
	previewPerRow = 5
	previewGap    = 4
	spriteDisplay = 16 * previewScale // 96px
)

// buildPreview scales each sprite 6× and arranges them 5 per row with 4px gaps.
func buildPreview(imgs []*image.NRGBA) *image.NRGBA {
	n := len(imgs)
	rows := (n + previewPerRow - 1) / previewPerRow

	cellW := spriteDisplay + previewGap
	cellH := spriteDisplay + previewGap

	totalW := previewPerRow*cellW - previewGap
	totalH := rows*cellH - previewGap

	preview := image.NewNRGBA(image.Rect(0, 0, totalW, totalH))

	for i, img := range imgs {
		col := i % previewPerRow
		row := i / previewPerRow
		offX := col * cellW
		offY := row * cellH

		for py := 0; py < 16; py++ {
			for px := 0; px < 16; px++ {
				c := img.NRGBAAt(px, py)
				for sy := 0; sy < previewScale; sy++ {
					for sx := 0; sx < previewScale; sx++ {
						preview.SetNRGBA(offX+px*previewScale+sx, offY+py*previewScale+sy, c)
					}
				}
			}
		}
	}
	return preview
}

func main() {
	outDir := "assets/sprites"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output dir: %v\n", err)
		os.Exit(1)
	}

	// Generate per-building sprites for all 284 buildings (3 variations each)
	if err := GenerateAllBuildingSprites("assets/sprites/buildings"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate building sprites: %v\n", err)
		os.Exit(1)
	}

	imgs := make([]*image.NRGBA, len(domainSprites))

	for i, s := range domainSprites {
		img := spriteToImage(s)
		imgs[i] = img
		path := fmt.Sprintf("%s/%s.png", outDir, s.name)
		if err := savePNG(path, img); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save %s: %v\n", path, err)
			os.Exit(1)
		}
	}

	// Sprite sheet: all 15 sprites in a single row (240×16).
	sheet := buildSpriteSheet(imgs)
	if err := savePNG(outDir+"/spritesheet.png", sheet); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save spritesheet: %v\n", err)
		os.Exit(1)
	}

	// Preview: 5 per row, scaled 6× (96px each), with 4px gaps.
	preview := buildPreview(imgs)
	if err := savePNG(outDir+"/preview.png", preview); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save preview: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %d sprites → %s/\n", len(domainSprites), outDir)
	fmt.Printf("  spritesheet.png  %dx16 px\n", len(domainSprites)*16)
	fmt.Printf("  preview.png      scaled 6× grid (%d per row, 96px each)\n", previewPerRow)
	for _, s := range domainSprites {
		fmt.Printf("  %s.png\n", s.name)
	}
}
