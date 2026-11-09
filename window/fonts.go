package window

import (
	"bytes"
	"embed"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/sfnt"
)

//go:embed fonts/*.ttf fonts/*.otf
var fontFiles embed.FS

// Font is the face the player picks for the whole window.
type Font string

const (
	FontMono  Font = "mono"  // JetBrains Mono, drawn smooth at any size
	FontPixel Font = "pixel" // Departure Mono, drawn at whole multiples of its pixel size
)

// pixelEm is the size, in pixels, Departure Mono is drawn on: one em is 11 of
// its pixels, so only multiples of 11 keep the pixels square and sharp.
const pixelEm = 11

// fontFile is one font on disk: loaded and parsed the first time a rune needs it.
type fontFile struct {
	path string
	once sync.Once
	sf   *sfnt.Font
	src  *text.GoTextFaceSource
}

func (f *fontFile) load() {
	f.once.Do(func() {
		b, err := fontFiles.ReadFile(f.path)
		if err != nil {
			return
		}
		if f.sf, err = sfnt.Parse(b); err != nil {
			f.sf = nil
			return
		}
		f.src, _ = text.NewGoTextFaceSource(bytes.NewReader(b))
	})
}

// has reports whether the font has a glyph for r.
func (f *fontFile) has(r rune) bool {
	f.load()
	if f.sf == nil || f.src == nil {
		return false
	}
	var buf sfnt.Buffer
	g, err := f.sf.GlyphIndex(&buf, r)
	return err == nil && g != 0
}

func ff(name string) *fontFile { return &fontFile{path: "fonts/" + name} }

var (
	jbRegular  = ff("JetBrainsMono-Regular.ttf")
	jbBold     = ff("JetBrainsMono-Bold.ttf")
	departure  = ff("DepartureMono-Regular.otf")
	symbols2   = ff("NotoSansSymbols2-Regular.ttf")
	symbols    = ff("NotoSansSymbols.ttf")
	mathFont   = ff("NotoSansMath-Regular.ttf")
	canadian   = ff("NotoSansCanadianAboriginal.ttf")
	japanese   = ff("NotoSansJP-Regular.ttf")
	emoji      = ff("NotoEmoji-Regular.ttf")
	extra      = ff("NotoSansExtra.ttf")
	fallbacks  = []*fontFile{symbols2, symbols, mathFont, canadian, japanese, emoji, extra}
	chainMono  = append([]*fontFile{jbRegular}, fallbacks...)
	chainBold  = append([]*fontFile{jbBold, jbRegular}, fallbacks...)
	chainPixel = append([]*fontFile{departure, jbRegular}, fallbacks...)
)

// faces picks the font for each rune. The first font in the chain with a glyph
// wins; a rune none has is remembered as missing.
type faces struct {
	mu    sync.Mutex
	picks map[faceKey]*fontFile
}

type faceKey struct {
	font Font
	bold bool
	r    rune
}

func newFaces() *faces { return &faces{picks: map[faceKey]*fontFile{}} }

func chainFor(f Font, bold bool) []*fontFile {
	switch {
	case f == FontPixel:
		return chainPixel
	case bold:
		return chainBold
	}
	return chainMono
}

// pick returns the font that draws r, or nil when no font has it.
func (fc *faces) pick(f Font, bold bool, r rune) *fontFile {
	k := faceKey{f, bold, r}
	fc.mu.Lock()
	defer fc.mu.Unlock()
	if p, ok := fc.picks[k]; ok {
		return p
	}
	var found *fontFile
	for _, c := range chainFor(f, bold) {
		if c.has(r) {
			found = c
			break
		}
	}
	fc.picks[k] = found
	return found
}

// missing lists the runes among rs that no font in the chain has.
func (fc *faces) missing(f Font, rs []rune) []rune {
	var out []rune
	for _, r := range rs {
		if !isProcedural(r) && fc.pick(f, false, r) == nil {
			out = append(out, r)
		}
	}
	return out
}
