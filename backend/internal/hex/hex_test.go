package hex

import "testing"

func TestBoardCells19(t *testing.T) {
	cells := BoardCells(2)
	if len(cells) != 19 {
		t.Fatalf("got %d cells", len(cells))
	}
	seen := map[Hex]bool{}
	for _, c := range cells {
		if !OnBoard(c, 2) {
			t.Fatalf("off board %+v", c)
		}
		if seen[c] {
			t.Fatalf("dup %+v", c)
		}
		seen[c] = true
	}
}

func TestNeighborWrap(t *testing.T) {
	h := Neighbor(Hex{0, 0}, 6)
	if h != (Hex{1, 0}) {
		t.Fatalf("%+v", h)
	}
}
