package match

import (
	"testing"

	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
)

func TestBattleInstantEndsTurn(t *testing.T) {
	m := startedMatch(t)
	first := m.TurnPlayerID
	p := m.player(first)
	// inject battle into hand
	battleID := injectInstant(p, "battle")
	if err := m.PlayInstant(first, battleID, nil, nil, nil, "", ""); err != nil {
		t.Fatal(err)
	}
	if m.TurnPlayerID == first {
		t.Fatal("battle should end turn")
	}
	if handHas(p, battleID) {
		t.Fatal("battle should be discarded")
	}
}

func TestSniperKills(t *testing.T) {
	m := startedMatch(t)
	// green has sniper
	var sniperPlayer *Player
	for _, p := range m.Players {
		if p.Army == protocol.ArmyGreen {
			sniperPlayer = p
			break
		}
	}
	enemy := m.other(sniperPlayer.ID)
	// make it sniper player's turn
	m.TurnPlayerID = sniperPlayer.ID
	m.MustDiscard = false
	m.Phase = protocol.MatchTurn

	unitDef := ""
	for _, id := range enemy.Pack.DrawPileIDs() {
		d := enemy.Pack.Def(id)
		if d != nil && d.Kind == "warrior" && d.Toughness == 0 && !d.HasSpecial("blocker") {
			unitDef = id
			break
		}
	}
	enemyUnit := &BoardTile{ID: "e1", DefID: unitDef, OwnerID: enemy.ID, Q: -1, R: 1}
	m.setTile(enemyUnit)

	shot := injectInstant(sniperPlayer, "sniper")
	if err := m.PlayInstant(sniperPlayer.ID, shot, nil, nil, nil, enemyUnit.ID, ""); err != nil {
		t.Fatal(err)
	}
	if m.boardByID(enemyUnit.ID) != nil {
		t.Fatal("enemy should be destroyed")
	}
}

func TestMoveUnit(t *testing.T) {
	m := startedMatch(t)
	first := m.TurnPlayerID
	unit := &BoardTile{ID: "u1", DefID: m.player(first).HQDefID, OwnerID: first, Q: 0, R: 0, Facing: 0}
	// find existing HQ
	for _, t := range m.Board {
		if t.OwnerID == first {
			unit = t
			break
		}
	}
	q, r := unit.Q+1, unit.R
	// find empty neighbor
	for f := 0; f < 6; f++ {
		// use hex neighbor via board empty check
	}
	_ = q
	_ = r
	destQ, destR := -1, -1
	for _, c := range [][2]int{{1, 0}, {0, 1}, {-1, 1}, {-1, 0}, {0, -1}, {1, -1}} {
		nq, nr := unit.Q+c[0], unit.R+c[1]
		if !m.hexOccupied(nq, nr) {
			destQ, destR = nq, nr
			break
		}
	}
	if destQ == -1 && destR == -1 {
		t.Fatal("no empty neighbor")
	}
	p := m.player(first)
	move := injectInstant(p, "move")
	fq, fr, ff := destQ, destR, 3
	if err := m.PlayInstant(first, move, &fq, &fr, &ff, unit.ID, ""); err != nil {
		t.Fatal(err)
	}
	moved := m.boardByID(unit.ID)
	if moved == nil || moved.Q != destQ || moved.R != destR || moved.Facing != 3 {
		t.Fatalf("move failed: %+v", moved)
	}
}

func startedMatch(t *testing.T) *Match {
	t.Helper()
	m, err := New(
		Seat{ID: "a", Army: protocol.ArmyGreen},
		Seat{ID: "b", Army: protocol.ArmyRed},
	)
	if err != nil {
		t.Fatal(err)
	}
	first := m.TurnPlayerID
	second := m.other(first).ID
	_ = m.Place(first, "", 0, 0, 0)
	_ = m.Place(second, "", 1, 0, 0)
	return m
}

func injectInstant(p *Player, kind string) string {
	for _, d := range p.Pack.AllDefs() {
		if d.IsInstant() && string(d.InstantKind()) == kind {
			id := "inj-" + kind
			p.Hand = append(p.Hand, TileInst{ID: id, DefID: d.ID})
			return id
		}
	}
	// sniper might be named InstantSniper with kind sniper
	panic("no instant " + kind + " in pack")
}

func handHas(p *Player, id string) bool {
	for _, t := range p.Hand {
		if t.ID == id {
			return true
		}
	}
	return false
}
