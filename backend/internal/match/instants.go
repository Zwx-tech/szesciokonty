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

func (m *Match) PlayInstant(playerID, tileID string, q, r *int, facing *int, targetTileID, passengerTileID string) error {
	if playerID != m.TurnPlayerID {
		return ErrNotYourTurn
	}
	if m.Phase != protocol.MatchTurn {
		return ErrBadPhase
	}
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
	if def == nil || !def.IsInstant() {
		putBack(p, *inst, idx)
		return ErrNotInstant
	}

	var err error
	switch def.InstantKind() {
	case tile.InstantBattle:
		err = m.playBattle()
	case tile.InstantMove:
		err = m.playMove(playerID, targetTileID, q, r, facing, passengerTileID)
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
	m.appendLog("Played %s", shortDefID(inst.DefID))
	return nil
}

func (m *Match) playBattle() error {
	if m.anyDeckEmpty() {
		return ErrNoBattle
	}
	m.resolveBattle()
	return nil
}

func (m *Match) playMove(playerID, targetID string, q, r *int, facing *int, passengerTileID string) error {
	t := m.boardByID(targetID)
	if t == nil || t.OwnerID != playerID {
		return ErrBadTarget
	}
	if def := m.defOf(t); def != nil && def.HasSpecial(tile.SpecialBlocker) {
		return ErrBadTarget
	}

	oldQ, oldR, oldFacing := t.Q, t.R, t.Facing
	if err := m.relocateUnit(t, q, r, facing); err != nil {
		return err
	}
	if err := m.tryTransportPassenger(playerID, t, oldQ, oldR, passengerTileID); err != nil {
		t.Q, t.R, t.Facing = oldQ, oldR, oldFacing
		return err
	}
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
	if def := m.defOf(t); def != nil && def.HasSpecial(tile.SpecialBlocker) {
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
	if m.hexOccupied(dest.Q, dest.R) {
		return ErrOccupied
	}

	m.removeTile(t)
	t.Q, t.R = dest.Q, dest.R
	m.setTile(t)
	return nil
}

func (m *Match) playSniper(playerID, targetID string) error {
	t := m.boardByID(targetID)
	if t == nil || t.OwnerID == playerID || m.isHQ(t) {
		return ErrBadTarget
	}
	dmg := 1
	if !m.playerHasScoper(playerID) {
		dmg -= m.anyArmorReduce(t)
	}
	if dmg > 0 {
		m.wound(t, dmg)
	}
	return nil
}

func (m *Match) anyArmorReduce(t *BoardTile) int {
	def := m.defOf(t)
	if def == nil {
		return 0
	}
	if len(def.ComponentsOf(tile.CompArmor)) > 0 {
		return 1
	}
	return 0
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
		for _, t := range m.tilesAt(h.Q, h.R) {
			if m.isHQ(t) {
				continue
			}
			m.wound(t, 1)
		}
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
		m.appendLog("Medic absorbs wound on %s", shortDefID(t.DefID))
		return
	}
	t.Wounds += n
	m.appendLog("%s takes %d wound(s)", shortDefID(t.DefID), n)
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
	m.appendLog("%s destroyed", shortDefID(t.DefID))
	m.removeTile(t)
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
	return m.Board[id]
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
