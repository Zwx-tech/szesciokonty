package match

import (
	"errors"

	"github.com/Zwx-tech/szesciokonty/backend/internal/hex"
	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
	"github.com/Zwx-tech/szesciokonty/backend/internal/tile"
)

var (
	ErrNotInstant = errors.New("not an instant")
	ErrBadTarget  = errors.New("invalid target")
	ErrNoBattle   = errors.New("battle tile not allowed")
)

func (m *Match) PlayInstant(playerID, tileID string, q, r *int, facing *int, targetTileID string) error {
	if playerID != m.TurnPlayerID {
		return ErrNotYourTurn
	}
	if m.Phase != protocol.MatchTurn {
		return ErrBadPhase
	}
	if m.MustDiscard {
		return ErrMustDiscard
	}

	p := m.player(playerID)
	inst, idx := takeHand(p, tileID)
	if inst == nil {
		return ErrBadTile
	}
	def := p.Pack.Def(inst.DefID)
	if def == nil || !def.IsInstant() {
		putBack(p, *inst, idx)
		return ErrNotInstant
	}

	var err error
	switch def.InstantKind() {
	case tile.InstantBattle:
		err = m.playBattle()
	case tile.InstantMove:
		err = m.playMove(playerID, targetTileID, q, r, facing)
	case tile.InstantPush:
		err = m.playPush(playerID, targetTileID, q, r)
	case tile.InstantSniper:
		err = m.playSniper(playerID, targetTileID)
	case tile.InstantAirStrike:
		err = m.playAirStrike(q, r)
	case tile.InstantGrenade:
		err = m.playGrenade(playerID, targetTileID)
	default:
		err = ErrNotInstant
	}
	if err != nil {
		putBack(p, *inst, idx)
		return err
	}

	p.Discard = append(p.Discard, *inst)
	m.UnluckyAvailable = false
	return nil
}

func (m *Match) playBattle() error {
	if m.anyDeckEmpty() {
		return ErrNoBattle
	}
	m.resolveBattle()
	return nil
}

func (m *Match) playMove(playerID, targetID string, q, r *int, facing *int) error {
	t := m.boardByID(targetID)
	if t == nil || t.OwnerID != playerID {
		return ErrBadTarget
	}

	newQ, newR := t.Q, t.R
	if q != nil && r != nil {
		newQ, newR = *q, *r
		if !hex.OnBoard(hex.Hex{Q: newQ, R: newR}, boardRadius) {
			return ErrBadHex
		}
		if newQ != t.Q || newR != t.R {
			if hexDist(t.Q, t.R, newQ, newR) != 1 {
				return ErrBadHex
			}
			if _, taken := m.Board[key(newQ, newR)]; taken {
				return ErrOccupied
			}
		}
	}
	newFacing := t.Facing
	if facing != nil {
		if *facing < 0 || *facing > 5 {
			return ErrBadFacing
		}
		newFacing = *facing
	}

	delete(m.Board, key(t.Q, t.R))
	t.Q, t.R, t.Facing = newQ, newR, newFacing
	m.Board[key(newQ, newR)] = t
	return nil
}

func (m *Match) playPush(playerID, targetID string, q, r *int) error {
	if q == nil || r == nil {
		return ErrBadHex
	}
	t := m.boardByID(targetID)
	if t == nil || t.OwnerID == playerID {
		return ErrBadTarget
	}
	if !m.adjacentToFriendly(playerID, t.Q, t.R) {
		return ErrBadTarget
	}
	dest := hex.Hex{Q: *q, R: *r}
	if !hex.OnBoard(dest, boardRadius) {
		return ErrBadHex
	}
	if hexDist(t.Q, t.R, dest.Q, dest.R) != 1 {
		return ErrBadHex
	}
	if _, taken := m.Board[key(dest.Q, dest.R)]; taken {
		return ErrOccupied
	}

	delete(m.Board, key(t.Q, t.R))
	t.Q, t.R = dest.Q, dest.R
	m.Board[key(dest.Q, dest.R)] = t
	return nil
}

func (m *Match) playSniper(playerID, targetID string) error {
	t := m.boardByID(targetID)
	if t == nil || t.OwnerID == playerID || m.isHQ(t) {
		return ErrBadTarget
	}
	m.wound(t, 1)
	return nil
}

func (m *Match) playAirStrike(q, r *int) error {
	if q == nil || r == nil {
		return ErrBadHex
	}
	center := hex.Hex{Q: *q, R: *r}
	area := []hex.Hex{center}
	for f := 0; f < 6; f++ {
		area = append(area, hex.Neighbor(center, f))
	}
	for _, h := range area {
		if !hex.OnBoard(h, boardRadius) {
			return ErrBadHex
		}
	}
	for _, h := range area {
		t := m.Board[key(h.Q, h.R)]
		if t == nil || m.isHQ(t) {
			continue
		}
		m.wound(t, 1)
	}
	return nil
}

func (m *Match) playGrenade(playerID, targetID string) error {
	t := m.boardByID(targetID)
	if t == nil || t.OwnerID == playerID || m.isHQ(t) {
		return ErrBadTarget
	}
	hq := m.hqTile(playerID)
	if hq == nil || hexDist(hq.Q, hq.R, t.Q, t.R) != 1 {
		return ErrBadTarget
	}
	m.wound(t, 99)
	return nil
}

func (m *Match) wound(t *BoardTile, n int) {
	if m.tryMedic(t) {
		return
	}
	t.Wounds += n
	if t.Wounds >= 1+m.toughness(t) {
		m.destroy(t)
	}
}

func (m *Match) tryMedic(t *BoardTile) bool {
	for _, mod := range m.Board {
		if mod.OwnerID != t.OwnerID || mod.ID == t.ID {
			continue
		}
		def := m.defOf(mod)
		if def == nil || def.Kind != tile.KindModule {
			continue
		}
		hasMedic := false
		for _, c := range def.ComponentsOf(tile.CompModuleAura) {
			for _, e := range c.Effects {
				if e.Kind == tile.EffectMedic {
					hasMedic = true
				}
			}
		}
		if !hasMedic || !m.moduleLinksTo(mod, t) {
			continue
		}
		m.destroy(mod)
		return true
	}
	return false
}

func (m *Match) moduleLinksTo(mod, target *BoardTile) bool {
	def := m.defOf(mod)
	if def == nil {
		return false
	}
	for _, c := range def.ComponentsOf(tile.CompModuleLink) {
		for _, d := range c.Dirs {
			abs := (mod.Facing + d) % 6
			n := hex.Neighbor(hex.Hex{Q: mod.Q, R: mod.R}, abs)
			if n.Q == target.Q && n.R == target.R {
				return true
			}
		}
	}
	return false
}

func (m *Match) destroy(t *BoardTile) {
	delete(m.Board, key(t.Q, t.R))
	if p := m.player(t.OwnerID); p != nil {
		p.Discard = append(p.Discard, TileInst{ID: t.ID, DefID: t.DefID})
	}
}

func (m *Match) toughness(t *BoardTile) int {
	if d := m.defOf(t); d != nil {
		return d.Toughness
	}
	return 0
}

func (m *Match) defOf(t *BoardTile) *tile.Def {
	p := m.player(t.OwnerID)
	if p == nil {
		return nil
	}
	return p.Pack.Def(t.DefID)
}

func (m *Match) isHQ(t *BoardTile) bool {
	p := m.player(t.OwnerID)
	return p != nil && t.DefID == p.HQDefID
}

func (m *Match) hqTile(playerID string) *BoardTile {
	p := m.player(playerID)
	if p == nil {
		return nil
	}
	for _, t := range m.Board {
		if t.OwnerID == playerID && t.DefID == p.HQDefID {
			return t
		}
	}
	return nil
}

func (m *Match) boardByID(id string) *BoardTile {
	for _, t := range m.Board {
		if t.ID == id {
			return t
		}
	}
	return nil
}

func (m *Match) adjacentToFriendly(playerID string, q, r int) bool {
	for _, t := range m.Board {
		if t.OwnerID == playerID && hexDist(t.Q, t.R, q, r) == 1 {
			return true
		}
	}
	return false
}

func (m *Match) anyDeckEmpty() bool {
	for _, p := range m.Players {
		if len(p.Deck) == 0 {
			return true
		}
	}
	return false
}

func hexDist(q1, r1, q2, r2 int) int {
	s1 := -q1 - r1
	s2 := -q2 - r2
	return (absInt(q1-q2) + absInt(r1-r2) + absInt(s1-s2)) / 2
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func putBack(p *Player, inst TileInst, idx int) {
	if idx < 0 || idx > len(p.Hand) {
		p.Hand = append(p.Hand, inst)
		return
	}
	rest := append([]TileInst{inst}, p.Hand[idx:]...)
	p.Hand = append(p.Hand[:idx], rest...)
}
