package kit_test

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// tview_test.go holds the kit to the text views the game's strings were
// written for: the same tagged string must come out as the same cells. It
// goes when tview does.

// corpus is tagged text of every kind the game writes, and the corner cases
// of the syntax.
var corpus = []string{
	"plain text",
	"[gold]Title[-] and [green]gain[-], [red]loss[-]",
	"[accent]role[-] [dim]names[-] [highlight::b]bold[-:-:-] back",
	"[onaccent:accent:b] Enter [text:chip:b] Keep [-:-:-] after",
	"[#ff8800]hex[-] [:#002244]ground[:-] [#00ff00:#220022:u]both[-:-:-]",
	"[::b]bold[::B] [::i]italic[::I] [::u]under[::U] [::d]dim[::D] [::r]rev[::R] [::s]strike[::-]",
	"[::bu]two[::-] [red::b]fg and attr[-::-]",
	"escaped [gold[] tag, [Enter[] key, [[] and [] and [",
	"a [word] nobody knows is still a tag",
	"[notacolor]unknown name[-] [1abc] digit first [a b] space inside",
	"[#12345]short hex [#1234567]long hex [#gggggg]not hex",
	"[red[[] double escape [red[[[] triple",
	"[:] empty parts [::] three [:::] four [-] reset [-:-] two [-:-:-] three",
	"[red][green][blue]stacked[-] tags, [red][blue[] then escape",
	"[\"region\"]inside a region[\"\"] outside",
	"[red:blue:b:https://example.com]link part[-:-:-:-] done",
	"unclosed [red and [#ff0000 and [::b",
	"wide 你好世界 glyphs and [red]彩色[-] mixed in",
	"emoji 🎉 and a keycap 1️⃣ and a family 👨‍👩‍👧 end",
	"combining é ä and zero​width",
	"line one\nline two\n\nline four",
	"ends with a break\n",
	"\nstarts with one",
	"a rather long sentence that has to wrap around a narrow box more than once to be laid out",
	"supercalifragilisticexpialidocious is longer than the box and wraps anywhere",
	"trailing   spaces   between    words   that   wrap",
	"[gold]colors [green]that [red]cross [cyan]a [yellow]wrap [white]point [gray]stay [-]right",
	"hyphen-ated words and slash/separated/paths break where they may",
	"  leading spaces and a [gold]tag[-] at the  end [gold]",
	"─── box ┌─┐ │x│ └─┘ ▁▂▃▄▅▆▇█ ░▒▓",
	"",
}

type diffCell struct {
	r      rune
	comb   string
	fg, bg color.RGBA
	attr   kit.Attr
}

func (c diffCell) String() string {
	return fmt.Sprintf("%q%q fg %v bg %v attr %b", c.r, c.comb, c.fg, c.bg, c.attr)
}

func plain(c tcell.Color, role theme.Role) color.RGBA {
	if !c.Valid() {
		return theme.RGBA(role)
	}
	r, g, b := c.RGB()
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

// viaTview is the cells a text view paints s as.
func viaTview(t *testing.T, s string, w, h int, wrap kit.Wrap, align kit.Align) []diffCell {
	t.Helper()
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(w, h)
	scr := theme.WrapScreen(sim)
	tv := tview.NewTextView().SetDynamicColors(true).SetRegions(true).
		SetWrap(wrap != kit.WrapNone).SetWordWrap(wrap == kit.WrapWord).
		SetTextAlign([]int{tview.AlignLeft, tview.AlignCenter, tview.AlignRight}[align])
	tv.SetText(s)
	tv.SetRect(0, 0, w, h)
	tv.Draw(scr)
	scr.Show()
	cells, cw, _ := sim.GetContents()
	out := make([]diffCell, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := cells[y*cw+x]
			fg, bg, at := c.Style.Decompose()
			d := diffCell{r: ' ', fg: plain(fg, theme.RoleText), bg: plain(bg, theme.RoleBackground)}
			if len(c.Runes) > 0 && c.Runes[0] != 0 {
				d.r, d.comb = c.Runes[0], string(c.Runes[1:])
			}
			for bit, a := range map[tcell.AttrMask]kit.Attr{
				tcell.AttrBold: kit.Bold, tcell.AttrItalic: kit.Italic, tcell.AttrDim: kit.Dim,
				tcell.AttrReverse: kit.Reverse, tcell.AttrStrikeThrough: kit.Strike, tcell.AttrBlink: kit.Blink,
			} {
				if at&bit != 0 {
					d.attr |= a
				}
			}
			out[y*w+x] = d
		}
	}
	return out
}

// viaKit is the cells the kit paints s as.
func viaKit(s string, w, h int, wrap kit.Wrap, align kit.Align) (*kit.Buffer, []diffCell) {
	tx := kit.NewText(s)
	tx.Wrap, tx.Align = wrap, align
	root := kit.NewRoot(tx)
	root.Resize(w, h)
	b := kit.NewBuffer(w, h)
	root.Paint(b)
	out := make([]diffCell, w*h)
	for i, c := range b.Cells {
		out[i] = diffCell{r: c.R, comb: c.Comb, fg: c.Fg, bg: c.Bg, attr: c.Attr}
	}
	return b, out
}

func TestTextPaintsAsTheTextViewDid(t *testing.T) {
	const h = 14
	for _, s := range corpus {
		for _, w := range []int{80, 31, 12, 5} {
			for _, wrap := range []kit.Wrap{kit.WrapNone, kit.WrapChar, kit.WrapWord} {
				if wrap != kit.WrapNone && strings.Contains(s, "[]") {
					// The text view wraps wrongly around an escaped tag: it
					// forgets the break after it, and prints the swallowed
					// bracket when a line starts there. The kit does neither
					// (see TestEscapedTagsWrapLikeTheirText).
					continue
				}
				for _, align := range []kit.Align{kit.AlignLeft, kit.AlignCenter, kit.AlignRight} {
					want := viaTview(t, s, w, h, wrap, align)
					buf, got := viaKit(s, w, h, wrap, align)
					for i := range want {
						if buf.Cells[i].Cont {
							continue // the right half of a wide glyph: tview leaves what was there
						}
						g, x := got[i], want[i]
						g.comb, x.comb = "", "" // tcell drops the second half of a flag; the kit keeps clusters whole
						// The text view loses track of underline (its attribute
						// reset leaves the line on), so that is the kit's own test.
						g.attr &^= kit.Underline
						if i%w == w-1 && g.r == ' ' && uniseg.StringWidth(string(x.r)) == 2 {
							continue // it lets a wide glyph hang over the right edge; the kit leaves a blank
						}
						if g.r == ' ' && x.r == ' ' {
							g.fg, x.fg, g.attr, x.attr = color.RGBA{}, color.RGBA{}, 0, 0 // a blank's ink does not show
						}
						if g != x {
							t.Errorf("%q at %dx%d wrap %d align %d: cell %d,%d\n  kit   %v\n  tview %v\n%s", s, w, h, wrap, align, i%w, i/w, g, x, strings.Join(buf.Lines(), "|\n"))
							break
						}
					}
				}
			}
		}
	}
}

func TestTheTextHelpersMatch(t *testing.T) {
	for _, s := range corpus {
		if got, want := kit.Width(s), tview.TaggedStringWidth(s); got != want {
			t.Errorf("Width(%q) = %d, the text view counts %d", s, got, want)
		}
		if got, want := kit.Escape(s), tview.Escape(s); got != want {
			t.Errorf("Escape(%q) = %q, want %q", s, got, want)
		}
		// A tag with a fourth part (a link) is not something Escape knows.
		if got := kit.Strip(kit.Escape(s)); got != s && !strings.Contains(s, ":::") && !strings.Contains(s, "https") {
			t.Errorf("Strip(Escape(%q)) = %q: an escaped string prints as written", s, got)
		}
		for _, w := range []int{80, 31, 12, 5} {
			if strings.Contains(s, "[]") {
				continue // see TestTextPaintsAsTheTextViewDid
			}
			got, want := kit.WordWrap(s, w), tview.WordWrap(s, w)
			if fmt.Sprintf("%q", got) != fmt.Sprintf("%q", want) {
				t.Errorf("WordWrap(%q, %d)\n  kit   %q\n  tview %q", s, w, got, want)
			}
		}
	}
}
