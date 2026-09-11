package match

import (
	"github.com/Zwx-tech/szesciokonty/backend/internal/hex"
	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
	"github.com/Zwx-tech/szesciokonty/backend/internal/tile"
)

func (m *Match) Place(playerID, tileID string, q, r, facing int) error {
	if playerID != m.TurnPlayerID {
		return ErrNotYourTurn
	}
	if facing < 0 || facing > 5 {
		return ErrBadFacing
	}
	if !hex.OnBoard(hex.Hex{Q: q, R: r}, boardRadius) {
		return ErrBadHex
	}

	switch m.Phase {
	case protocol.MatchPlaceHQ:
		return m.placeHQ(playerID, q, r, facing)
	case protocol.MatchTurn:
		return m.placeUnit(playerID, tileID, q, r, facing)
	default:
		return ErrBadPhase
	}
}

func (m *Match) placeHQ(playerID string, q, r, facing int) error {
	p := m.player(playerID)
	if p == nil {
		return ErrNotYourTurn
	}
	if m.hexOccupied(q, r) {
		return ErrOccupied
	}
	for _, t := range m.Board {
		if t.OwnerID == playerID && t.DefID == p.HQDefID {
			return ErrHQDone
		}
	}
	m.setTile(&BoardTile{
		ID: newID(), DefID: p.HQDefID, OwnerID: playerID,
		Q: q, R: r, Facing: facing,
	})

	other := m.other(playerID)
	otherHasHQ := false
	for _, t := range m.Board {
		if t.OwnerID == other.ID && t.DefID == other.HQDefID {
			otherHasHQ = true
			break
		}
	}
	if !otherHasHQ {
		m.TurnPlayerID = other.ID
		return nil
	}

	// both HQs placed → opening turns
	m.Phase = protocol.MatchTurn
	m.TurnPlayerID = m.FirstPlayerID
	m.openingStep = 0
	m.beginTurn()
	return nil
}

func (m *Match) placeUnit(playerID, tileID string, q, r, facing int) error {
	if m.MustDiscard {
		return ErrMustDiscard
	}
	m.clearReconPeek()
	p := m.player(playerID)
	inst, idx := takeHand(p, tileID)
	if inst == nil {
		return ErrBadTile
	}
	def := p.Pack.Def(inst.DefID)
	if def == nil || !def.IsUnit() || def.Kind == tile.KindHQ {
		rest := append([]TileInst{*inst}, p.Hand[idx:]...)
		p.Hand = append(p.Hand[:idx], rest...)
		return ErrNotUnit
	}
	if m.hexOccupied(q, r) && !m.canDualStackPlace(playerID, def, q, r) {
		putBack(p, *inst, idx)
		return ErrOccupied
	}

	m.setTile(&BoardTile{
		ID: inst.ID, DefID: inst.DefID, OwnerID: playerID,
		Q: q, R: r, Facing: facing,
	})
	m.UnluckyAvailable = false

	if m.allHexesOccupied() {
		m.resolveBattle()
	}
	return nil
}

func (m *Match) Discard(playerID, tileID string) error {
	if playerID != m.TurnPlayerID {
		return ErrNotYourTurn
	}
	if m.Phase != protocol.MatchTurn {
		return ErrBadPhase
	}
	m.clearReconPeek()
	p := m.player(playerID)
	inst, _ := takeHand(p, tileID)
	if inst == nil {
		return ErrBadTile
	}
	p.Discard = append(p.Discard, *inst)
	if m.MustDiscard {
		m.MustDiscard = false
	}
	m.UnluckyAvailable = false
	return nil
}

func (m *Match) EndTurn(playerID string) error {
	if playerID != m.TurnPlayerID {
		return ErrNotYourTurn
	}
	if m.Phase != protocol.MatchTurn {
		return ErrBadPhase
	}
	if m.MustDiscard {
		return ErrMustDiscard
	}
	m.afterPlayerEndedTurn(playerID)
	return nil
}

func (m *Match) RedrawUnlucky(playerID string) error {
	if playerID != m.TurnPlayerID {
		return ErrNotYourTurn
	}
	if m.Phase != protocol.MatchTurn || !m.UnluckyAvailable {
		return ErrNoUnlucky
	}
	p := m.player(playerID)
	drawn := make(map[string]bool, len(m.drawnIDs))
	for _, id := range m.drawnIDs {
		drawn[id] = true
	}
	kept := p.Hand[:0]
	for _, t := range p.Hand {
		if drawn[t.ID] {
			p.Discard = append(p.Discard, t)
		} else {
			kept = append(kept, t)
		}
	}
	p.Hand = kept
	m.drawnIDs = nil
	m.UnluckyAvailable = false
	m.drawUpTo(p, 3)
	m.checkUnlucky(p)
	return nil
}

func (m *Match) beginTurn() {
	p := m.player(m.TurnPlayerID)
	m.MustDiscard = false
	m.UnluckyAvailable = false
	m.drawnIDs = nil
	m.mobilityUsed = make(map[string]bool)
	m.reconUsed = false
	m.quartermasterUsed = false
	m.reconPeek = nil
	m.reconPeekFor = ""

	target := 3
	switch m.openingStep {
	case 0:
		target = 1
	case 1:
		target = 2
	}

	before := len(p.Hand)
	m.drawUpTo(p, target)
	drew := len(p.Hand) - before
	for i := before; i < len(p.Hand); i++ {
		m.drawnIDs = append(m.drawnIDs, p.Hand[i].ID)
	}

	if m.openingStep >= 2 && len(p.Hand) == 3 && drew > 0 {
		// normal turn filled to 3 → mandatory discard (unless deck emptied mid-draw with <3)
		m.MustDiscard = true
	}
	if m.openingStep >= 2 && drew == 3 {
		m.checkUnlucky(p)
	}
}

func (m *Match) advanceTurn() {
	m.openingStep++
	other := m.other(m.TurnPlayerID)
	m.TurnPlayerID = other.ID
	m.beginTurn()
}

func (m *Match) drawUpTo(p *Player, n int) {
	for len(p.Hand) < n && len(p.Deck) > 0 {
		t := p.Deck[0]
		p.Deck = p.Deck[1:]
		p.Hand = append(p.Hand, t)
		m.noteDeckExhausted(p)
	}
}

func (m *Match) checkUnlucky(p *Player) {
	if len(m.drawnIDs) != 3 {
		return
	}
	for _, id := range m.drawnIDs {
		t := handByID(p, id)
		if t == nil {
			return
		}
		def := p.Pack.Def(t.DefID)
		if def == nil || !def.IsInstant() {
			return
		}
	}
	m.UnluckyAvailable = true
}

func (m *Match) other(id string) *Player {
	for _, p := range m.Players {
		if p.ID != id {
			return p
		}
	}
	return m.Players[0]
}

func takeHand(p *Player, tileID string) (*TileInst, int) {
	for i, t := range p.Hand {
		if t.ID == tileID {
			inst := t
			p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
			return &inst, i
		}
	}
	return nil, -1
}

func handByID(p *Player, id string) *TileInst {
	for i := range p.Hand {
		if p.Hand[i].ID == id {
			return &p.Hand[i]
		}
	}
	return nil
}
