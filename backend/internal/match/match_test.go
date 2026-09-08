package match

import (
	"testing"

	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
	"github.com/Zwx-tech/szesciokonty/backend/internal/tile"
)

func TestHQThenOpeningDraw(t *testing.T) {
	m, err := New(
		Seat{ID: "a", Army: protocol.ArmyRed},
		Seat{ID: "b", Army: protocol.ArmyBlue},
	)
	if err != nil {
		t.Fatal(err)
	}
	if m.Phase != protocol.MatchPlaceHQ {
		t.Fatal(m.Phase)
	}
	first := m.TurnPlayerID
	second := m.other(first).ID

	if err := m.Place(first, "", 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	if m.TurnPlayerID != second {
		t.Fatal("expected second to place HQ")
	}
	if err := m.Place(second, "", 1, 0, 0); err != nil {
		t.Fatal(err)
	}
	if m.Phase != protocol.MatchTurn {
		t.Fatal(m.Phase)
	}
	if m.TurnPlayerID != first {
		t.Fatal("first should open")
	}
	p := m.player(first)
	if len(p.Hand) != 1 {
		t.Fatalf("opening hand want 1 got %d", len(p.Hand))
	}
	if m.MustDiscard {
		t.Fatal("no discard on opening draw 1")
	}
}

func TestNormalDiscardAndPlace(t *testing.T) {
	m, err := New(
		Seat{ID: "a", Army: protocol.ArmyRed},
		Seat{ID: "b", Army: protocol.ArmyGreen},
	)
	if err != nil {
		t.Fatal(err)
	}
	first := m.TurnPlayerID
	second := m.other(first).ID
	_ = m.Place(first, "", 0, 0, 0)
	_ = m.Place(second, "", 1, 0, 0)

	_ = m.EndTurn(first)
	_ = m.EndTurn(second)

	p := m.player(first)
	if len(p.Hand) != 3 || !m.MustDiscard {
		t.Fatalf("hand=%d mustDiscard=%v", len(p.Hand), m.MustDiscard)
	}
	if err := m.Place(first, p.Hand[0].ID, 0, 1, 0); err != ErrMustDiscard {
		t.Fatalf("want must discard, got %v", err)
	}
	disc := p.Hand[0].ID
	if err := m.Discard(first, disc); err != nil {
		t.Fatal(err)
	}
	if m.MustDiscard {
		t.Fatal("discard cleared")
	}

	placed := false
	for _, h := range append([]TileInst{}, p.Hand...) {
		def := p.Pack.Def(h.DefID)
		if def == nil || !def.IsUnit() || def.Kind == tile.KindHQ {
			continue
		}
		if err := m.Place(first, h.ID, -1, 0, 0); err != nil {
			t.Fatal(err)
		}
		placed = true
		break
	}
	if !placed {
		// hand may be all instants — discard until a unit or end turn
		for len(p.Hand) > 0 {
			h := p.Hand[0]
			def := p.Pack.Def(h.DefID)
			if def != nil && def.IsUnit() && def.Kind != tile.KindHQ {
				if err := m.Place(first, h.ID, -1, 0, 0); err != nil {
					t.Fatal(err)
				}
				placed = true
				break
			}
			_ = m.Discard(first, h.ID)
		}
	}
	if err := m.EndTurn(first); err != nil {
		t.Fatal(err)
	}
	if m.TurnPlayerID != second {
		t.Fatal("turn should pass")
	}
}
