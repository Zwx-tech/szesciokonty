package hex

import (
	"fmt"
	"math"
	"strings"
)

// Axial helpers shared conceptually with the frontend (q, r). Pointy-top facing 0 = +q.

type Hex struct {
	Q, R int
}

func (h Hex) S() int { return -h.Q - h.R }

type Pt struct{ X, Y float64 }

var Dirs = [6]Hex{
	{1, 0},
	{0, 1},
	{-1, 1},
	{-1, 0},
	{0, -1},
	{1, -1},
}

// Pointy-top axial → pixel (matches frontend hexToPixel).
const sqrt3 = 1.7320508075688772

// Pointy-top pixel offset for facing dir, matching frontend hexToPixel(HEX_DIRS[dir], size).
func DirPixel(dir int, size float64) Pt {
	q, r := Dirs[Mod6(dir)].Q, Dirs[Mod6(dir)].R
	return Pt{
		X: size * (sqrt3*float64(q) + (sqrt3/2)*float64(r)),
		Y: size * ((3.0 / 2.0) * float64(r)),
	}
}

func HexCorners(size float64) [6]Pt {
	var out [6]Pt
	for i := 0; i < 6; i++ {
		angle := (math.Pi / 180) * float64(60*i-30)
		out[i] = Pt{X: size * math.Cos(angle), Y: size * math.Sin(angle)}
	}
	return out
}

func PolyPoints(corners [6]Pt) string {
	parts := make([]string, 6)
	for i, c := range corners {
		parts[i] = fmt.Sprintf("%.2f,%.2f", c.X, c.Y)
	}
	return strings.Join(parts, " ")
}

func Mod6(n int) int {
	n %= 6
	if n < 0 {
		n += 6
	}
	return n
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
