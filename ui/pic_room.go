package ui

// pic_room.go works out the room a picture takes beside text. A picture's art
// pixels are real pixels, max(1, cell height/21) of them to one art pixel, so
// its width does not follow the cell's: the cells set aside for it come from
// its right edge and the cell's width, here and nowhere else.

// artUnit is the real pixels in one art pixel for a cell ch pixels tall.
func artUnit(ch int) int { return max(1, ch/21) }

// picRoom is the whole cells a picture takes when its right edge is right
// pixels from where it starts and the text beside it keeps gap pixels clear.
func picRoom(right, gap, cw int) int {
	if cw <= 0 {
		return 0
	}
	return (right + gap + cw - 1) / cw
}

const (
	// actionSignArtW is a sign on an action button, with the pixel its shadow
	// takes.
	actionSignArtW = 14
)

// actionSignCols is the cells at the left of an action button its sign takes
// where there are no pictures, and the least with them.
const actionSignCols = 4

// actionSignX is where a sign starts in an action button, in art pixels.
func actionSignX(cw, ch int) int { return cw * 7 / 5 / artUnit(ch) }

// actionSignRoom is the cells at the left of an action button that its sign
// takes, for cells of cw by ch pixels; actionSignCols without pictures.
func actionSignRoom(cw, ch int) int {
	if cw <= 0 || ch <= 0 {
		return actionSignCols
	}
	u := artUnit(ch)
	return max(actionSignCols, picRoom((actionSignX(cw, ch)+actionSignArtW)*u, u, cw))
}

// pictureCells is the cell size pictures are drawn at, zero where there are
// none (no window).
func pictureCells() (cw, ch int) {
	if pictureCell != nil {
		if w, h := pictureCell(); w > 0 && h > 0 {
			return w, h
		}
	}
	return 0, 0
}
