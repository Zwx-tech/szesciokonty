package main

import "github.com/Zwx-tech/szesciokonty/backend/internal/protocol"

// Geometry
const (
	hexSize     = 40.0
	svgViewMin  = -52.0
	svgViewSize = 104.0
)

// Body strokes
const (
	bodyStroke      = "#111111"
	bodyStrokeWidth = 1.5
	hqStrokeWidth   = 2.5
)

// General edge constants
const (
	edgeStroke      = "#111111"
	edgeStrokeWidth = 1.0
)

// * Net constants
const (
	netScaleX = 0.08
	netScaleY = 0.2
)

// * Melee constants
const (
	meleeScaleX = 0.25
	meleeScaleY = 0.3
)

// * Ranged constants
const (
	rangeScaleX = 0.37
	rangeScaleY = 0.15
)

// * Link constants
const (
	linkScaleX = 0.37
	linkScaleY = 0.4
)

// * Barrier (armor) — curved band along the protected edge
const (
	barrierScaleX   = 0    // inset from each corner along the edge
	barrierScaleY   = 0.2  // inward thickness at the middle (fraction of radius)
	barrierOuterPad = 0.97 // pull outer curve slightly inside the hex stroke
	barrierStroke   = "#eeeeee"
	barrierStrokeW  = 0.5
)

// * Line wound (pierce) — long thin triangle, distinct from normal ranged
const (
	lineWoundScaleX      = 0.37
	lineWoundScaleY      = 0.15
	lineWoundStroke      = "#999999"
	lineWoundStrokeWidth = 2
)

// Label typography
const (
	labelFontFamily = "Georgia, serif"
	labelFill       = "#ffffff"
	instantFontSize = 11.0
	labelY          = 5.0
)

// Initiative badges (black circle + white number, as on physical Neuroshima Hex tiles).
// Badges sit along rays toward hex corners, between center and vertex.
const (
	initBadgeR           = 6.0
	initBadgeAlong       = 0.42 // 0 = center, 1 = hex corner
	initBadgeStartCorner = 3    // rear (west) when facing 0 = east; extras go clockwise
	initBadgeFill        = "#111111"
	initBadgeStroke      = "#111111"
	initTextFill         = "#ffffff"
	initFontFamily       = "Arial Black, Arial, Helvetica, sans-serif"
	initFontSize         = 9.0
	initFontWeight       = "700"
)

type EdgeShapeType string

const (
	ShapeTriangle EdgeShapeType = "triangle"
	ShapeHidden   EdgeShapeType = "hidden"
	ShapeLink     EdgeShapeType = "link"
	ShapeBarrier  EdgeShapeType = "barrier"
)

type ComponentDrawParams struct {
	EdgeShape EdgeShapeType
	ScaleX    float64
	ScaleY    float64
	Fill      string
	Stroke    string
	StrokeW   float64
	Dirs      []int // optional override when the component has no Dirs (e.g. line wound)
}

var armyFill = map[protocol.Army]string{
	protocol.ArmyRed:    "#aa3333",
	protocol.ArmyBlue:   "#3366aa",
	protocol.ArmyGreen:  "#339933",
	protocol.ArmyYellow: "#bbaa33",
}

var componentFill = map[string]string{
	"melee":      "#f0f0f0",
	"ranged":     "#f0f0f0",
	"net":        "#e8c84a",
	"link":       "#f0f0f0",
	"barrier":    "#222222",
	"line_wound": "#f0f0f0",
}

// Draw priority: lower drawn first (underneath).
var componentDrawOrder = map[string]int{
	"barrier":    0,
	"net":        1,
	"melee":      2,
	"ranged":     3,
	"line_wound": 4,
	"link":       5,
}
