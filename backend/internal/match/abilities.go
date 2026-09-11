package match

import (
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/Zwx-tech/szesciokonty/backend/internal/hex"
	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
	"github.com/Zwx-tech/szesciokonty/backend/internal/tile"
)

var (
	ErrNoMobility      = errors.New("mobility not available")
	ErrNoRecon         = errors.New("recon not available")
	ErrNoQuartermaster = errors.New("quartermaster not available")
	ErrBadPassenger    = errors.New("invalid transport passenger")
)

func (m *Match) UseMobility(playerID, tileID string, q, r *int, facing *int, passengerID string) error {
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
	t := m.boardByID(tileID)
	if t == nil || t.OwnerID != playerID {
		return ErrBadTarget
	}
	def := m.defOf(t)
	if def == nil || !def.HasComponent(tile.CompMobility) {
		return ErrNoMobility
	}
	if m.mobilityUsed[t.ID] {
		return ErrNoMobility
	}
	if m.nettedSet()[t.ID] {
		return ErrNoMobility
	}
	if def.HasSpecial(tile.SpecialBlocker) {
		return ErrBadTarget
	}

	oldQ, oldR := t.Q, t.R
	if err := m.relocateUnit(t, q, r, facing); err != nil {
		return err
	}
	m.mobilityUsed[t.ID] = true
	if err := m.tryTransportPassenger(playerID, t, oldQ, oldR, passengerID); err != nil {
		// roll back mobility move
		_ = m.relocateUnit(t, &oldQ, &oldR, nil)
		delete(m.mobilityUsed, t.ID)
		return err
	}
	m.appendLog("Mobility: %s", shortDefID(t.DefID))
	m.UnluckyAvailable = false
	return nil
}

func (m *Match) UseRecon(playerID string) error {
	if playerID != m.TurnPlayerID {
		return ErrNotYourTurn
	}
	if m.Phase != protocol.MatchTurn || m.MustDiscard {
		return ErrBadPhase
	}
	if !m.canUseRecon(playerID) {
		return ErrNoRecon
	}
	opp := m.other(playerID)
	n := 3
	if len(opp.Deck) < n {
		n = len(opp.Deck)
	}
	peek := make([]string, 0, n)
	// Sample without removing: shuffle indices into a temp pick of n.
	idxs := make([]int, len(opp.Deck))
	for i := range idxs {
		idxs[i] = i
	}
	shuffleInts(idxs)
	for i := 0; i < n; i++ {
		peek = append(peek, opp.Deck[idxs[i]].DefID)
	}
	m.reconPeek = peek
	m.reconPeekFor = playerID
	m.reconUsed = true
	m.appendLog("Recon peek")
	m.UnluckyAvailable = false
	return nil
}

func (m *Match) UseQuartermaster(playerID, tileID string) error {
	if playerID != m.TurnPlayerID {
		return ErrNotYourTurn
	}
	if m.Phase != protocol.MatchTurn || m.MustDiscard {
		return ErrBadPhase
	}
	if !m.canUseQuartermaster(playerID) {
		return ErrNoQuartermaster
	}
	m.clearReconPeek()
	p := m.player(playerID)
	var inst *TileInst
	var idx int
	for i := range p.Discard {
		if p.Discard[i].ID == tileID {
			inst = &p.Discard[i]
			idx = i
			break
		}
	}
	if inst == nil {
		return ErrBadTile
	}
	moved := *inst
	p.Discard = append(p.Discard[:idx], p.Discard[idx+1:]...)
	p.Deck = append(p.Deck, moved)
	m.quartermasterUsed = true
	m.appendLog("Quartermaster recycled %s", shortDefID(moved.DefID))
	m.UnluckyAvailable = false
	return nil
}

func (m *Match) canUseRecon(playerID string) bool {
	if playerID != m.TurnPlayerID || m.Phase != protocol.MatchTurn || m.MustDiscard || m.reconUsed {
		return false
	}
	return m.playerHasModuleEffect(playerID, tile.EffectRecon)
}

func (m *Match) canUseQuartermaster(playerID string) bool {
	if playerID != m.TurnPlayerID || m.Phase != protocol.MatchTurn || m.MustDiscard || m.quartermasterUsed {
		return false
	}
	return m.playerHasModuleEffect(playerID, tile.EffectQuartermaster)
}

func (m *Match) playerHasScoper(playerID string) bool {
	return m.playerHasModuleEffect(playerID, tile.EffectScoper)
}

func (m *Match) playerHasModuleEffect(playerID string, kind tile.EffectKind) bool {
	netted := m.nettedSet()
	for _, mod := range m.Board {
		if mod.OwnerID != playerID || netted[mod.ID] {
			continue
		}
		def := m.defOf(mod)
		if def == nil || def.Kind != tile.KindModule {
			continue
		}
		has := false
		for _, c := range def.ComponentsOf(tile.CompModuleAura) {
			for _, e := range c.Effects {
				if e.Kind == kind {
					has = true
				}
			}
		}
		if !has {
			continue
		}
		for _, t := range m.Board {
			if t.OwnerID == playerID && t.ID != mod.ID && m.moduleLinksTo(mod, t) {
				return true
			}
		}
	}
	return false
}

func (m *Match) relocateUnit(t *BoardTile, q, r *int, facing *int) error {
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
			if m.hexOccupied(newQ, newR) {
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
	t.Q, t.R, t.Facing = newQ, newR, newFacing
	return nil
}

func (m *Match) tryTransportPassenger(playerID string, moved *BoardTile, oldQ, oldR int, passengerID string) error {
	if passengerID == "" {
		return nil
	}
	if oldQ == moved.Q && oldR == moved.R {
		return ErrBadPassenger
	}
	if m.hexOccupied(oldQ, oldR) {
		return ErrBadPassenger
	}
	pass := m.boardByID(passengerID)
	if pass == nil || pass.OwnerID != playerID || pass.ID == moved.ID {
		return ErrBadPassenger
	}
	movedDef := m.defOf(moved)
	passDef := m.defOf(pass)
	if movedDef == nil || passDef == nil ||
		movedDef.Kind != tile.KindWarrior || passDef.Kind != tile.KindWarrior {
		return ErrBadPassenger
	}
	// Links are evaluated at the mover's vacated hex (pre-move), not the new one.
	if m.sharedTransportAt(playerID, oldQ, oldR, pass) == nil {
		return ErrBadPassenger
	}
	pass.Q, pass.R = oldQ, oldR
	m.appendLog("Transport moved %s", shortDefID(pass.DefID))
	return nil
}

func (m *Match) sharedTransportAt(playerID string, aq, ar int, b *BoardTile) *BoardTile {
	netted := m.nettedSet()
	for _, mod := range m.Board {
		if mod.OwnerID != playerID || netted[mod.ID] {
			continue
		}
		def := m.defOf(mod)
		if def == nil || def.Kind != tile.KindModule {
			continue
		}
		has := false
		for _, c := range def.ComponentsOf(tile.CompModuleAura) {
			for _, e := range c.Effects {
				if e.Kind == tile.EffectTransport {
					has = true
				}
			}
		}
		if !has {
			continue
		}
		if m.moduleLinksToHex(mod, aq, ar) && m.moduleLinksTo(mod, b) {
			return mod
		}
	}
	return nil
}

func (m *Match) moduleLinksToHex(mod *BoardTile, q, r int) bool {
	def := m.defOf(mod)
	if def == nil {
		return false
	}
	for _, c := range def.ComponentsOf(tile.CompModuleLink) {
		for _, d := range c.Dirs {
			abs := (mod.Facing + d) % 6
			n := hex.Neighbor(hex.Hex{Q: mod.Q, R: mod.R}, abs)
			if n.Q == q && n.R == r {
				return true
			}
		}
	}
	return false
}

func (m *Match) sharedTransportModule(playerID string, a, b *BoardTile) *BoardTile {
	return m.sharedTransportAt(playerID, a.Q, a.R, b)
}

func (m *Match) clearReconPeek() {
	m.reconPeek = nil
	m.reconPeekFor = ""
}

func shuffleInts(a []int) {
	for i := len(a) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			a[i], a[0] = a[0], a[i]
			continue
		}
		j := int(n.Int64())
		a[i], a[j] = a[j], a[i]
	}
}
