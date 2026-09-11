package match

import (
	"fmt"

	"github.com/Zwx-tech/szesciokonty/backend/internal/hex"
	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
	"github.com/Zwx-tech/szesciokonty/backend/internal/tile"
)

// hit is one simultaneous strike collected in an initiative phase.
type hit struct {
	attackerID string
	edge       int // absolute dir of the attack edge on the attacker
	targetID   string
	strength   int
	ranged     bool
	pierce     bool // line wound: ignore friendlies / hit all in line handled separately
}

func (m *Match) resolveBattle() {
	m.Phase = protocol.MatchBattle
	m.pendingDestroy = map[string]bool{}
	m.pendingReplay = &protocol.BattleReplay{}
	m.appendLog("Battle begins")
	m.beginStepLog()

	maxInit := 0
	for _, t := range m.Board {
		for _, v := range m.effectiveInits(t, m.nettedSet()) {
			if v > maxInit {
				maxInit = v
			}
		}
	}

	for phase := maxInit; phase >= 0; phase-- {
		m.resolvePhase(phase)
		m.recordBattleStep(phase, fmt.Sprintf("Initiative %d", phase))
		if m.bothHQDead() {
			break
		}
	}
	if !m.bothHQDead() {
		m.resolveExtraAttacks()
		m.recordBattleStep(-1, "Extra attacks")
	}

	m.finishBattle()
}

func (m *Match) finishBattle() {
	a, b := m.Players[0], m.Players[1]
	aDead, bDead := a.HQHP <= 0, b.HQHP <= 0
	if aDead || bDead {
		m.Phase = protocol.MatchEnded
		m.TurnPlayerID = ""
		switch {
		case aDead && bDead:
			m.Result = &protocol.MatchResult{Draw: true}
		case aDead:
			m.Result = &protocol.MatchResult{WinnerID: b.ID}
		default:
			m.Result = &protocol.MatchResult{WinnerID: a.ID}
		}
		return
	}

	if m.endMode == endFinalArmed || m.endMode == endTieBreak {
		m.endByScore()
		return
	}

	if m.endMode == endAwaitOpponent && m.TurnPlayerID == m.deckExhaustedBy {
		m.endMode = endFinalArmed
	}

	m.Phase = protocol.MatchTurn
	m.MustDiscard = false
	m.UnluckyAvailable = false
	m.advanceTurn()
}

func (m *Match) bothHQDead() bool {
	return m.Players[0].HQHP <= 0 && m.Players[1].HQHP <= 0
}

func (m *Match) resolvePhase(phase int) {
	netted := m.nettedSet()
	hits := m.collectHits(phase, netted, false)
	m.applyHits(hits, netted)
	m.endPhaseCleanup()
}

func (m *Match) resolveExtraAttacks() {
	netted := m.nettedSet()
	hits := m.collectHits(-1, netted, true)
	if len(hits) == 0 {
		return
	}
	m.applyHits(hits, netted)
	m.endPhaseCleanup()
}

func (m *Match) collectHits(phase int, netted map[string]bool, extraOnly bool) []hit {
	var hits []hit
	for _, atk := range m.Board {
		if netted[atk.ID] {
			continue
		}
		def := m.defOf(atk)
		if def == nil {
			continue
		}

		acts := false
		if extraOnly {
			acts = m.hasExtraAttack(atk, netted)
		} else {
			for _, v := range m.effectiveInits(atk, netted) {
				if v == phase {
					acts = true
					break
				}
			}
		}
		if !acts {
			continue
		}

		hits = append(hits, m.attacksFrom(atk, def)...)
	}
	return hits
}

func (m *Match) hasExtraAttack(t *BoardTile, netted map[string]bool) bool {
	def := m.defOf(t)
	if def == nil {
		return false
	}
	if def.Kind == tile.KindHQ {
		for _, c := range def.ComponentsOf(tile.CompHQAura) {
			for _, e := range c.Effects {
				if e.Kind == tile.EffectExtraAttack {
					return true
				}
			}
		}
	}
	return m.auraValue(t, tile.EffectMother, netted) > 0
}

func (m *Match) attacksFrom(atk *BoardTile, def *tile.Def) []hit {
	netted := m.nettedSet()
	var hits []hit
	meleeBonus := m.auraValue(atk, tile.EffectMeleeBonus, netted)
	rangedBonus := m.auraValue(atk, tile.EffectRangedBonus, netted)

	for _, c := range def.ComponentsOf(tile.CompAttack) {
		str := c.Strength
		ranged := c.AttackType == tile.AttackRanged
		if ranged {
			str += rangedBonus
		} else {
			str += meleeBonus
		}
		for _, rel := range c.Dirs {
			abs := (atk.Facing + rel) % 6
			hits = append(hits, m.hitsOnDir(atk, abs, str, ranged, false)...)
		}
	}

	for _, c := range def.ComponentsOf(tile.CompSpecial) {
		if c.ID != tile.SpecialLineWound {
			continue
		}
		str := c.Params["strength"]
		if str <= 0 {
			str = 1
		}
		str += rangedBonus
		abs := atk.Facing % 6
		hits = append(hits, m.hitsOnDir(atk, abs, str, true, true)...)
	}
	return hits
}

func (m *Match) hitsOnDir(atk *BoardTile, absDir, str int, ranged, pierce bool) []hit {
	if !ranged && !pierce {
		n := hex.Neighbor(hex.Hex{Q: atk.Q, R: atk.R}, absDir)
		var hits []hit
		for _, t := range m.tilesAt(n.Q, n.R) {
			if t.OwnerID == atk.OwnerID {
				continue
			}
			hits = append(hits, hit{atk.ID, absDir, t.ID, str, false, false})
		}
		return hits
	}

	var hits []hit
	h := hex.Hex{Q: atk.Q, R: atk.R}
	for {
		h = hex.Neighbor(h, absDir)
		if !hex.OnBoard(h, boardRadius) {
			break
		}
		tiles := m.tilesAt(h.Q, h.R)
		if len(tiles) == 0 {
			continue
		}
		if pierce {
			for _, t := range tiles {
				hits = append(hits, hit{atk.ID, absDir, t.ID, str, true, true})
			}
			continue
		}
		blocked := false
		for _, t := range tiles {
			if t.OwnerID == atk.OwnerID {
				blocked = true
				break
			}
		}
		if blocked {
			break
		}
		for _, t := range tiles {
			hits = append(hits, hit{atk.ID, absDir, t.ID, str, true, false})
		}
		break
	}
	return hits
}

func (m *Match) applyHits(hits []hit, netted map[string]bool) {
	hqDmg := map[string]int{}
	// medic: one absorb per medic tile this phase
	medicUsed := map[string]bool{}

	// group by target for medic choice: absorb largest strike first
	type applied struct {
		h   hit
		dmg int
	}
	var queue []applied
	for _, h := range hits {
		tgt := m.boardByID(h.targetID)
		atk := m.boardByID(h.attackerID)
		if tgt == nil || atk == nil {
			continue
		}
		if m.isHQ(atk) && m.isHQ(tgt) {
			continue
		}
		dmg := h.strength
		if h.ranged {
			skipArmor := atk != nil && m.playerHasScoper(atk.OwnerID)
			if !skipArmor {
				dmg -= m.armorReduce(tgt, h.edge)
				if dmg < 0 {
					dmg = 0
				}
			}
		}
		if dmg == 0 {
			continue
		}
		queue = append(queue, applied{h, dmg})
	}

	for _, a := range queue {
		tgt := m.boardByID(a.h.targetID)
		atk := m.boardByID(a.h.attackerID)
		if tgt == nil {
			continue
		}
		if m.tryBattleMedic(tgt, a.h.attackerID, medicUsed, netted) {
			m.appendLog("Init: medic absorbs hit on %s", shortDefID(tgt.DefID))
			continue
		}
		if m.isHQ(tgt) {
			hqDmg[tgt.OwnerID] += a.dmg
			atkName := "?"
			if atk != nil {
				atkName = shortDefID(atk.DefID)
			}
			m.appendLog("%s hits HQ for %d", atkName, a.dmg)
			continue
		}
		tgt.Wounds += a.dmg
		atkName := "?"
		if atk != nil {
			atkName = shortDefID(atk.DefID)
		}
		m.appendLog("%s hits %s for %d", atkName, shortDefID(tgt.DefID), a.dmg)
		if tgt.Wounds >= 1+m.toughness(tgt) {
			m.pendingDestroy[tgt.ID] = true
		}
	}

	for pid, dmg := range hqDmg {
		if p := m.player(pid); p != nil {
			p.HQHP -= dmg
			if p.HQHP < 0 {
				p.HQHP = 0
			}
		}
	}
}

func (m *Match) armorReduce(tgt *BoardTile, attackAbsDir int) int {
	def := m.defOf(tgt)
	if def == nil {
		return 0
	}
	// Attack arrives on the edge facing the attacker: opposite of travel dir.
	hitEdge := (attackAbsDir + 3) % 6
	for _, c := range def.ComponentsOf(tile.CompArmor) {
		for _, rel := range c.Dirs {
			if (tgt.Facing+rel)%6 == hitEdge {
				return 1
			}
		}
	}
	return 0
}

func (m *Match) tryBattleMedic(tgt *BoardTile, attackerID string, used map[string]bool, netted map[string]bool) bool {
	for _, mod := range m.Board {
		if mod.OwnerID != tgt.OwnerID || used[mod.ID] || netted[mod.ID] || m.pendingDestroy[mod.ID] {
			continue
		}
		def := m.defOf(mod)
		if def == nil || def.Kind != tile.KindModule {
			continue
		}
		has := false
		for _, c := range def.ComponentsOf(tile.CompModuleAura) {
			for _, e := range c.Effects {
				if e.Kind == tile.EffectMedic {
					has = true
				}
			}
		}
		if !has || !m.moduleLinksTo(mod, tgt) {
			continue
		}
		used[mod.ID] = true
		m.pendingDestroy[mod.ID] = true
		_ = attackerID // one enemy attack absorbed; choice is automatic
		return true
	}
	return false
}

func (m *Match) endPhaseCleanup() {
	for {
		var dead []*BoardTile
		for id := range m.pendingDestroy {
			if t := m.boardByID(id); t != nil {
				dead = append(dead, t)
			}
		}
		m.pendingDestroy = map[string]bool{}
		if len(dead) == 0 {
			return
		}
		for _, t := range dead {
			m.destroy(t)
			m.triggerDetonate(t)
		}
	}
}

func (m *Match) triggerDetonate(t *BoardTile) {
	def := m.defOf(t)
	if def == nil || !def.HasSpecial(tile.SpecialDetonate) {
		return
	}
	m.appendLog("%s detonates", shortDefID(t.DefID))
	for f := 0; f < 6; f++ {
		n := hex.Neighbor(hex.Hex{Q: t.Q, R: t.R}, f)
		for _, vic := range m.tilesAt(n.Q, n.R) {
			if m.isHQ(vic) {
				if p := m.player(vic.OwnerID); p != nil {
					p.HQHP--
					if p.HQHP < 0 {
						p.HQHP = 0
					}
				}
				continue
			}
			vic.Wounds++
			if vic.Wounds >= 1+m.toughness(vic) {
				m.pendingDestroy[vic.ID] = true
			}
		}
	}
}

// --- nets / bonuses / initiative ---

func (m *Match) nettedSet() map[string]bool {
	type edge struct{ from, to string }
	var nets []edge
	for _, src := range m.Board {
		def := m.defOf(src)
		if def == nil {
			continue
		}
		for _, c := range def.ComponentsOf(tile.CompNet) {
			for _, rel := range c.Dirs {
				abs := (src.Facing + rel) % 6
				n := hex.Neighbor(hex.Hex{Q: src.Q, R: src.R}, abs)
				for _, tgt := range m.tilesAt(n.Q, n.R) {
					if tgt.OwnerID == src.OwnerID {
						continue
					}
					nets = append(nets, edge{src.ID, tgt.ID})
				}
			}
		}
	}

	mutual := map[string]bool{}
	for _, e := range nets {
		for _, o := range nets {
			if e.from == o.to && e.to == o.from {
				mutual[e.from+"|"+e.to] = true
				mutual[o.from+"|"+o.to] = true
			}
		}
	}

	disabled := map[string]bool{}
	changed := true
	for changed {
		changed = false
		for _, e := range nets {
			if mutual[e.from+"|"+e.to] || disabled[e.from] {
				continue
			}
			if !disabled[e.to] {
				disabled[e.to] = true
				changed = true
			}
		}
	}
	return disabled
}

func (m *Match) effectiveInits(t *BoardTile, netted map[string]bool) []int {
	def := m.defOf(t)
	if def == nil || len(def.Initiatives) == 0 {
		return nil
	}
	bonus := m.auraValue(t, tile.EffectInitBonus, netted)
	penalty := m.saboteurPenalty(t, netted)
	out := make([]int, len(def.Initiatives))
	for i, v := range def.Initiatives {
		v = v + bonus - penalty
		if v < 0 {
			v = 0
		}
		out[i] = v
	}
	return out
}

func (m *Match) auraValue(t *BoardTile, kind tile.EffectKind, netted map[string]bool) int {
	sum := 0
	for _, src := range m.Board {
		if src.OwnerID != t.OwnerID || netted[src.ID] || src.ID == t.ID {
			continue
		}
		def := m.defOf(src)
		if def == nil {
			continue
		}
		switch def.Kind {
		case tile.KindModule:
			if !m.moduleLinksTo(src, t) {
				continue
			}
			sum += effectSum(def.ComponentsOf(tile.CompModuleAura), kind)
		case tile.KindHQ:
			if hexDist(src.Q, src.R, t.Q, t.R) != 1 {
				continue
			}
			sum += effectSum(def.ComponentsOf(tile.CompHQAura), kind)
		}
	}
	return sum
}

func effectSum(comps []tile.Component, kind tile.EffectKind) int {
	sum := 0
	for _, c := range comps {
		for _, e := range c.Effects {
			if e.Kind != kind {
				continue
			}
			if e.Value == 0 {
				sum++
			} else {
				sum += e.Value
			}
		}
	}
	return sum
}

func (m *Match) saboteurPenalty(t *BoardTile, netted map[string]bool) int {
	sum := 0
	for _, src := range m.Board {
		if src.OwnerID == t.OwnerID || netted[src.ID] {
			continue
		}
		def := m.defOf(src)
		if def == nil || def.Kind != tile.KindModule {
			continue
		}
		if !m.moduleLinksTo(src, t) {
			continue
		}
		for _, c := range def.ComponentsOf(tile.CompModuleAura) {
			for _, e := range c.Effects {
				if e.Kind == tile.EffectSaboteur {
					v := e.Value
					if v == 0 {
						v = 1
					}
					sum += v
				}
			}
		}
	}
	return sum
}
