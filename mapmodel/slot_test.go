package mapmodel

import "testing"

// TestSkySlotsNeverCollide: a (copy, type) maps to one slot, and no two share
// one, however many types an age has.
func TestSkySlotsNeverCollide(t *testing.T) {
	for _, nt := range []int{1, 7, 30, 43} {
		type spot struct{ col, row int }
		seen := map[spot][2]int{}
		for tt := 0; tt < nt; tt++ {
			for c := 0; c < SkyMaxCopies; c++ {
				col, row := skySlot(c, tt, nt)
				if o, dup := seen[spot{col, row}]; dup {
					t.Fatalf("%d types: copy %d of type %d and copy %d of type %d share slot %d/%d", nt, c, tt, o[0], o[1], col, row)
				}
				seen[spot{col, row}] = [2]int{c, tt}
				if row < 0 || row > 2 {
					t.Fatalf("slot row %d", row)
				}
			}
		}
	}
}
