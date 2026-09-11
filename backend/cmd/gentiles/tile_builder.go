package main

import (
	"fmt"
	"strings"

	"github.com/Zwx-tech/szesciokonty/backend/internal/hex"
)

type TileBuilder struct {
	b strings.Builder
}

func xmlEscape(s string) string {
	r := strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&quot;",
	)
	return r.Replace(s)
}

func (tb *TileBuilder) New() {
	tb.b.Reset()

	//* add XML headers
	tb.b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	fmt.Fprintf(&tb.b,
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="%.0f %.0f %.0f %.0f" width="%.0f" height="%.0f">`,
		svgViewMin, svgViewMin, svgViewSize, svgViewSize, svgViewSize, svgViewSize,
	)
	tb.b.WriteString("\n")
}

func (tb *TileBuilder) CreateBase(fill string, strokeW float64) {
	corners := hex.HexCorners(hexSize)
	tb.b.WriteString(`  <polygon points="`)
	tb.b.WriteString(hex.PolyPoints(corners))
	tb.b.WriteString(`" fill="`)
	tb.b.WriteString(fill)
	fmt.Fprintf(&tb.b, `" stroke="%s" stroke-width="%.1f"/>`, bodyStroke, strokeW)
	tb.b.WriteString("\n")
}

/*
	Draws triangle on the specified edge

ScaleX: Procentage of triangle base width (0.0 = no base, 1.0 = full edge width)
ScaleY: Percentage of triangle height (0.0 = no height, 1.0 = full hex radius)
*/
func (tb *TileBuilder) CreateTriangleEdge(dir int, scaleX float64, scaleY float64, fill string, stroke string, strokeW float64) {
	dir = hex.Mod6(dir)
	corners := hex.HexCorners(hexSize)
	a := corners[dir]
	b := corners[(dir+1)%6]

	inv := 1 - scaleX
	ax := a.X*inv + b.X*scaleX
	ay := a.Y*inv + b.Y*scaleX
	bx := a.X*scaleX + b.X*inv
	by := a.Y*scaleX + b.Y*inv

	tip := hex.DirPixel(dir, hexSize*scaleY)

	fmt.Fprintf(&tb.b,
		`  <polygon points="%.2f,%.2f %.2f,%.2f %.2f,%.2f" fill="%s" fill-opacity="1.0" stroke="%s" stroke-width="%.1f"/>`+"\n",
		ax, ay, bx, by, tip.X, tip.Y, fill, stroke, strokeW,
	)
}

func (tb *TileBuilder) CreateLinkEdge(dir int, scaleX float64, scaleY float64, fill string, stroke string, strokeW float64) {
	dir = hex.Mod6(dir)
	corners := hex.HexCorners(hexSize)
	a := corners[dir]
	b := corners[(dir+1)%6]

	inv := 1 - scaleX
	ax := a.X*inv + b.X*scaleX
	ay := a.Y*inv + b.Y*scaleX
	bx := a.X*scaleX + b.X*inv
	by := a.Y*scaleX + b.Y*inv

	tx := ((ax + bx) / 2) * scaleY
	ty := ((ay + by) / 2) * scaleY

	cx := ax - tx
	cy := ay - ty

	dx := bx - tx
	dy := by - ty

	fmt.Fprintf(&tb.b,
		`  <polygon points="%.2f,%.2f %.2f,%.2f %.2f,%.2f %.2f,%.2f" fill="%s" fill-opacity="1.0" stroke="%s" stroke-width="%.1f"/>`+"\n",
		ax, ay, bx, by, dx, dy, cx, cy, fill, stroke, strokeW,
	)
}

/*
	Draws an armor/barrier band along the specified edge (Neuroshima Hex armor icon).
	Uses a closed <path> with quadratic beziers — outer arc along the hex edge,
	inner arc bowed toward the center — matching the original curved shield mark.

ScaleX: inset from each corner along the edge (0 = full edge, 0.5 = point)
ScaleY: inward thickness at mid-edge as a fraction of hex radius
*/
func (tb *TileBuilder) CreateBarrierEdge(dir int, scaleX float64, scaleY float64, fill string, stroke string, strokeW float64) {
	dir = hex.Mod6(dir)
	corners := hex.HexCorners(hexSize)
	a := corners[dir]
	b := corners[(dir+1)%6]

	inv := 1 - scaleX
	// Endpoints of the band along the hex edge.
	e1x := a.X*inv + b.X*scaleX
	e1y := a.Y*inv + b.Y*scaleX
	e2x := a.X*scaleX + b.X*inv
	e2y := a.Y*scaleX + b.Y*inv

	// Outer endpoints / control, slightly inside the tile body.
	o1x, o1y := e1x*barrierOuterPad, e1y*barrierOuterPad
	o2x, o2y := e2x*barrierOuterPad, e2y*barrierOuterPad
	omx := (a.X + b.X) / 2 * barrierOuterPad
	omy := (a.Y + b.Y) / 2 * barrierOuterPad

	// Inner endpoints / control: ends a bit less deep, middle bows toward center.
	innerEnd := 1 - scaleY*0.75
	innerMid := 1 - scaleY
	i1x, i1y := e1x*innerEnd, e1y*innerEnd
	i2x, i2y := e2x*innerEnd, e2y*innerEnd
	imx := (a.X + b.X) / 2 * innerMid
	imy := (a.Y + b.Y) / 2 * innerMid

	// M o1 → Q outer mid → o2 → L i2 → Q inner mid → i1 → Z
	fmt.Fprintf(&tb.b,
		`  <path d="M %.2f,%.2f Q %.2f,%.2f %.2f,%.2f L %.2f,%.2f Q %.2f,%.2f %.2f,%.2f Z" fill="%s" fill-opacity="1.0" stroke="%s" stroke-width="%.1f"/>`+"\n",
		o1x, o1y,
		omx, omy, o2x, o2y,
		i2x, i2y,
		imx, imy, i1x, i1y,
		fill, stroke, strokeW,
	)
}

func (tb *TileBuilder) CreateLabel(text string, fontSize float64) {
	fmt.Fprintf(&tb.b,
		`  <text x="0" y="%.0f" text-anchor="middle" font-family="%s" font-size="%.0f" fill="%s">%s</text>`+"\n",
		labelY, labelFontFamily, fontSize, labelFill, xmlEscape(text),
	)
}

// CreateInitiatives draws Neuroshima-style initiative badges: black circles with
// white numbers. Each badge sits between the tile center and a hex corner
// (rear side when facing 0), matching physical tile layout.
func (tb *TileBuilder) CreateInitiatives(values []int) {
	if len(values) == 0 {
		return
	}
	corners := hex.HexCorners(hexSize)
	for i, v := range values {
		c := corners[hex.Mod6(initBadgeStartCorner+i)]
		x := c.X * initBadgeAlong
		y := c.Y * initBadgeAlong
		fmt.Fprintf(&tb.b,
			`  <circle cx="%.2f" cy="%.2f" r="%.2f" fill="%s" stroke="%s" stroke-width="1"/>`+"\n",
			x, y, initBadgeR, initBadgeFill, initBadgeStroke,
		)
		fmt.Fprintf(&tb.b,
			`  <text x="%.2f" y="%.2f" text-anchor="middle" dominant-baseline="central" font-family="%s" font-size="%.0f" font-weight="%s" fill="%s">%d</text>`+"\n",
			x, y, initFontFamily, initFontSize, initFontWeight, initTextFill, v,
		)
	}
}

func (tb *TileBuilder) CreateEdge(dir int, params ComponentDrawParams) {
	switch params.EdgeShape {
	case ShapeTriangle:
		tb.CreateTriangleEdge(dir, params.ScaleX, params.ScaleY, params.Fill, params.Stroke, params.StrokeW)
	case ShapeLink:
		tb.CreateLinkEdge(dir, params.ScaleX, params.ScaleY, params.Fill, params.Stroke, params.StrokeW)
	case ShapeBarrier:
		tb.CreateBarrierEdge(dir, params.ScaleX, params.ScaleY, params.Fill, params.Stroke, params.StrokeW)
	case ShapeHidden:
		// do nothing
	default:
		// unknown component type; skip
	}
}

func (tb *TileBuilder) SvgString() string {
	tb.b.WriteString(`</svg>`)
	tb.b.WriteString("\n")
	return tb.b.String()
}
