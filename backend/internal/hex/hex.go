package hex

// Axial helpers shared conceptually with the frontend (q, r). Pointy-top facing 0 = +q.

type Hex struct {
	Q, R int
}

func (h Hex) S() int { return -h.Q - h.R }

var Dirs = [6]Hex{
	{1, 0},
	{1, -1},
	{0, -1},
	{-1, 0},
	{-1, 1},
	{0, 1},
}

func Neighbor(h Hex, facing int) Hex {
	d := Dirs[(facing%6+6)%6]
	return Hex{h.Q + d.Q, h.R + d.R}
}

func OnBoard(h Hex, radius int) bool {
	return abs(h.Q) <= radius && abs(h.R) <= radius && abs(h.S()) <= radius
}

func BoardCells(radius int) []Hex {
	cells := make([]Hex, 0, 1+3*radius*(radius+1))
	for q := -radius; q <= radius; q++ {
		r1 := max(-radius, -q-radius)
		r2 := min(radius, -q+radius)
		for r := r1; r <= r2; r++ {
			cells = append(cells, Hex{q, r})
		}
	}
	return cells
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
