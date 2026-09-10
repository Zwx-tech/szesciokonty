package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
	"github.com/Zwx-tech/szesciokonty/backend/internal/tile"
)

type manifestArmy struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	File string `json:"file"`
}

type manifest struct {
	Armies map[string][]manifestArmy `json:"armies"`
}

func main() {
	outDir := filepath.Join("..", "frontend", "public", "tiles")
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fatal(err)
	}

	man := manifest{Armies: map[string][]manifestArmy{}}
	for _, army := range tile.Armies() {
		pack, err := tile.PackFor(army)
		if err != nil {
			fatal(err)
		}
		defs := pack.AllDefs()
		sort.Slice(defs, func(i, j int) bool {
			return defs[i].ID < defs[j].ID
		})
		entries := make([]manifestArmy, 0, len(defs))
		for _, def := range defs {
			svg := generateTileSvg(def, army)
			file := def.ID + ".svg"
			path := filepath.Join(outDir, file)
			if err := os.WriteFile(path, []byte(svg), 0o644); err != nil {
				fatal(err)
			}
			entries = append(entries, manifestArmy{
				ID: def.ID, Kind: string(def.Kind), File: file,
			})
		}
		man.Armies[string(army)] = entries
		fmt.Printf("%s: %d tiles\n", army, len(entries))
	}

	b, err := json.MarshalIndent(man, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "manifest.json"), append(b, '\n'), 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %s\n", outDir)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "gentiles: %v\n", err)
	os.Exit(1)
}

func generateTileSvg(def *tile.Def, army protocol.Army) string {
	fill := armyFill[army]
	strokeW := bodyStrokeWidth
	if def.Kind == tile.KindHQ {
		strokeW = hqStrokeWidth
	}

	tb := TileBuilder{}
	tb.New()
	tb.CreateBase(fill, strokeW)

	if def.Kind == tile.KindInstant {
		tb.CreateLabel(string(def.InstantKind()), instantFontSize)
		return tb.SvgString()
	}

	// Edge marks: attacks, nets, and module links (draw order via ComponentPriority).
	var tileComponents []tile.Component
	tileComponents = append(tileComponents, def.ComponentsOf(tile.CompNet)...)
	tileComponents = append(tileComponents, def.ComponentsOf(tile.CompAttack)...)
	tileComponents = append(tileComponents, def.ComponentsOf(tile.CompModuleLink)...)
	sort.Slice(tileComponents, func(i, j int) bool {
		return ComponentPriority(tileComponents[i]) < ComponentPriority(tileComponents[j])
	})

	for _, edge := range tileComponents {
		params := getComponentDrawParams(edge)
		if params.EdgeShape == ShapeHidden {
			continue
		}
		for _, side := range edge.Dirs {
			tb.CreateEdge(side, params)
		}
	}

	tb.CreateInitiatives(def.Initiatives)

	return tb.SvgString()
}

// ComponentPriority: lower numbers are drawn first (underneath).
func ComponentPriority(component tile.Component) int {
	switch component.Type {
	case tile.CompNet:
		return componentDrawOrder["net"]
	case tile.CompAttack:
		if component.AttackType == tile.AttackRanged {
			return componentDrawOrder["ranged"]
		}
		return componentDrawOrder["melee"]
	case tile.CompModuleLink:
		return componentDrawOrder["link"]
	default:
		return 999 // unknown component type; draw last
	}
}

func getComponentDrawParams(component tile.Component) ComponentDrawParams {
	switch component.Type {
	case tile.CompNet:
		return ComponentDrawParams{
			EdgeShape: ShapeTriangle,
			ScaleX:    netScaleX,
			ScaleY:    netScaleY,
			Fill:      componentFill["net"],
			Stroke:    edgeStroke,
			StrokeW:   edgeStrokeWidth,
		}
	case tile.CompAttack:
		if component.AttackType == tile.AttackRanged {
			return ComponentDrawParams{
				EdgeShape: ShapeTriangle,
				ScaleX:    rangeScaleX,
				ScaleY:    rangeScaleY,
				Fill:      componentFill["ranged"],
				Stroke:    edgeStroke,
				StrokeW:   edgeStrokeWidth,
			}
		}
		return ComponentDrawParams{
			EdgeShape: ShapeTriangle,
			ScaleX:    meleeScaleX,
			ScaleY:    meleeScaleY,
			Fill:      componentFill["melee"],
			Stroke:    edgeStroke,
			StrokeW:   edgeStrokeWidth,
		}
	case tile.CompModuleLink:
		return ComponentDrawParams{
			EdgeShape: ShapeLink,
			ScaleX:    linkScaleX,
			ScaleY:    linkScaleY,
			Fill:      componentFill["link"],
			Stroke:    edgeStroke,
			StrokeW:   edgeStrokeWidth,
		}
	default:
		return ComponentDrawParams{
			EdgeShape: ShapeHidden,
		}
	}
}
