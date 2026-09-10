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
)

type ComponentDrawParams struct {
	EdgeShape EdgeShapeType
	ScaleX    float64
	ScaleY    float64
	Fill      string
	Stroke    string
	StrokeW   float64
}

var armyFill = map[protocol.Army]string{
	protocol.ArmyRed:    "#aa3333",
	protocol.ArmyBlue:   "#3366aa",
	protocol.ArmyGreen:  "#339933",
	protocol.ArmyYellow: "#bbaa33",
}

var componentFill = map[string]string{
	"melee":  "#f0f0f0",
	"ranged": "#f0f0f0",
	"net":    "#e8c84a",
	"link":   "#f0f0f0",
}

// Draw priority: nets under ranged under melee.
var componentDrawOrder = map[string]int{
	"net":    0,
	"ranged": 1,
	"melee":  2,
	"link":   3,
}
