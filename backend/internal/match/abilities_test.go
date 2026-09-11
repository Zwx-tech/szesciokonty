package match

import (
	"testing"

	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
	"github.com/Zwx-tech/szesciokonty/backend/internal/tile"
)

func abilityMatch(t *testing.T, a, b protocol.Army) *Match {
	t.Helper()
	m, err := New(Seat{ID: "a", Army: a}, Seat{ID: "b", Army: b})
	if err != nil {
		t.Fatal(err)
	}
	m.Phase = protocol.MatchTurn
	m.TurnPlayerID = "a"
	m.MustDiscard = false
	m.Board = map[string]*BoardTile{}
	m.mobilityUsed = map[string]bool{}
	put(m, "a", m.player("a").HQDefID, -2, 2, 0)
	put(m, "b", m.player("b").HQDefID, 2, -2, 0)
	return m
}

func TestMobilityMoveAndRotate(t *testing.T) {
	m := abilityMatch(t, protocol.ArmyGreen, protocol.ArmyRed)
	u := put(m, "a", "green_courier", 0, 0, 0)

	q, r := 1, 0
	if err := m.UseMobility("a", u.ID, &q, &r, nil, ""); err != nil {
		t.Fatal(err)
	}
	if u.Q != 1 || u.R != 0 {
		t.Fatalf("pos=%d,%d", u.Q, u.R)
	}
	if !m.mobilityUsed[u.ID] {
		t.Fatal("mobility should be marked used")
	}

	facing := 3
	if err := m.UseMobility("a", u.ID, nil, nil, &facing, ""); err != ErrNoMobility {
		t.Fatalf("double mobility want ErrNoMobility got %v", err)
	}

	m.mobilityUsed = map[string]bool{}
	if err := m.UseMobility("a", u.ID, nil, nil, &facing, ""); err != nil {
		t.Fatal(err)
	}
	if u.Facing != 3 {
		t.Fatalf("facing=%d", u.Facing)
	}
}

func TestMobilityNettedRejected(t *testing.T) {
	m := abilityMatch(t, protocol.ArmyGreen, protocol.ArmyBlue)
	u := put(m, "a", "green_courier", 0, 0, 0)
	put(m, "b", "blue_netter", 1, 0, 3) // nets toward courier

	q, r := 0, 1
	if err := m.UseMobility("a", u.ID, &q, &r, nil, ""); err != ErrNoMobility {
		t.Fatalf("want ErrNoMobility got %v", err)
	}
}

func TestBlockerPushRejected(t *testing.T) {
	m := abilityMatch(t, protocol.ArmyBlue, protocol.ArmyBlue)
	blocker := put(m, "b", "blue_bulwark", 1, 0, 0)
	put(m, "a", "blue_chaser", 0, 0, 0) // adjacent friendly for push validity

	dq, dr := 2, 0
	if err := m.playPush("a", blocker.ID, &dq, &dr); err != ErrBadTarget {
		t.Fatalf("want ErrBadTarget got %v", err)
	}
}

func TestBlockerOwnMoveRejected(t *testing.T) {
	m := abilityMatch(t, protocol.ArmyBlue, protocol.ArmyRed)
	blocker := put(m, "a", "blue_bulwark", 0, 0, 0)
	q, r := 1, 0
	if err := m.playMove("a", blocker.ID, &q, &r, nil, ""); err != ErrBadTarget {
		t.Fatalf("want ErrBadTarget got %v", err)
	}
}

func TestScoperSkipsRangedArmor(t *testing.T) {
	m := battleFixture(t, protocol.ArmyGreen, protocol.ArmyBlue)
	target := put(m, "b", "blue_plated_gunner", 1, 0, 3) // armor on edge facing west
	put(m, "a", "green_marksman", 0, 0, 0)               // ranged 1 toward (1,0)
	put(m, "a", "green_hijack", -1, 0, 0)                // link0 → marksman (scoper)

	m.resolveBattle()

	if m.boardByID(target.ID) != nil {
		t.Fatal("scoper should bypass armor so strength-1 ranged kills")
	}
}

func TestScoperSkipsSniperArmor(t *testing.T) {
	m := abilityMatch(t, protocol.ArmyGreen, protocol.ArmyBlue)
	put(m, "a", "green_hijack", 0, 0, 0)
	put(m, "a", "green_courier", 1, 0, 0) // linked for scoper
	target := put(m, "b", "blue_plated_gunner", 2, -1, 0)

	if !m.playerHasScoper("a") {
		t.Fatal("expected scoper active")
	}
	if err := m.playSniper("a", target.ID); err != nil {
		t.Fatal(err)
	}
	if m.boardByID(target.ID) != nil {
		t.Fatalf("scoper sniper should kill armored unit wounds=%d", target.Wounds)
	}
}

func TestSniperBlockedByArmorWithoutScoper(t *testing.T) {
	m := abilityMatch(t, protocol.ArmyGreen, protocol.ArmyBlue)
	target := put(m, "b", "blue_plated_gunner", 1, 0, 0)

	if err := m.playSniper("a", target.ID); err != nil {
		t.Fatal(err)
	}
	if m.boardByID(target.ID) == nil {
		t.Fatal("sniper without scoper should not pierce armor")
	}
}


func TestReconPeek(t *testing.T) {
	m := abilityMatch(t, protocol.ArmyGreen, protocol.ArmyRed)
	put(m, "a", "green_pathfinder", 0, 0, 0)
	put(m, "a", "green_courier", 1, 0, 0) // linked by pathfinder dirs 0..5

	opp := m.player("b")
	if len(opp.Deck) < 3 {
		t.Fatal("need deck")
	}
	if err := m.UseRecon("a"); err != nil {
		t.Fatal(err)
	}
	if len(m.reconPeek) != 3 {
		t.Fatalf("peek len=%d", len(m.reconPeek))
	}
	snap := m.Snapshot("a")
	if len(snap.ReconPeek) != 3 {
		t.Fatalf("snapshot peek=%d", len(snap.ReconPeek))
	}
	if len(m.Snapshot("b").ReconPeek) != 0 {
		t.Fatal("opponent should not see peek")
	}
	if err := m.UseRecon("a"); err != ErrNoRecon {
		t.Fatalf("double recon want ErrNoRecon got %v", err)
	}
}

func TestDualStackPlace(t *testing.T) {
	m := abilityMatch(t, protocol.ArmyYellow, protocol.ArmyRed)
	base := put(m, "a", "yellow_crew", 0, 0, 0)
	p := m.player("a")
	dual := TileInst{ID: "dual1", DefID: "yellow_dual"}
	p.Hand = append(p.Hand, dual)

	if err := m.Place("a", dual.ID, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	tiles := m.tilesAt(0, 0)
	if len(tiles) != 2 {
		t.Fatalf("want 2 tiles got %d", len(tiles))
	}
	_ = base
	if m.primaryAt(0, 0) == nil {
		t.Fatal("primary missing")
	}
}

func TestQuartermasterRecycle(t *testing.T) {
	m := abilityMatch(t, protocol.ArmyYellow, protocol.ArmyRed)
	put(m, "a", "yellow_converter", 0, 0, 0)
	put(m, "a", "yellow_crew", 1, 0, 0)

	p := m.player("a")
	recycled := TileInst{ID: "disc1", DefID: "yellow_thumper"}
	p.Discard = append(p.Discard, recycled)
	before := len(p.Deck)

	if err := m.UseQuartermaster("a", recycled.ID); err != nil {
		t.Fatal(err)
	}
	if len(p.Discard) != 0 {
		t.Fatal("should leave discard")
	}
	if len(p.Deck) != before+1 || p.Deck[len(p.Deck)-1].ID != recycled.ID {
		t.Fatal("should append to deck bottom")
	}
	snap := m.Snapshot("a")
	var me protocol.PlayerView
	for _, pv := range snap.Players {
		if pv.ID == "a" {
			me = pv
		}
	}
	if len(me.Discard) != 0 {
		t.Fatal("viewer discard should be empty")
	}
}

func TestTransportPassenger(t *testing.T) {
	m := abilityMatch(t, protocol.ArmyYellow, protocol.ArmyRed)
	// carrier links 0,1,2 → place at (0,0) facing 0 so link0=(1,0), link1=(0,1)
	put(m, "a", "yellow_carrier", 0, 0, 0)
	mover := put(m, "a", "yellow_crew", 1, 0, 0)
	pass := put(m, "a", "yellow_courier", 0, 1, 0)

	q, r := 2, 0
	if err := m.playMove("a", mover.ID, &q, &r, nil, pass.ID); err != nil {
		t.Fatal(err)
	}
	if mover.Q != 2 || mover.R != 0 {
		t.Fatalf("mover=%d,%d", mover.Q, mover.R)
	}
	if pass.Q != 1 || pass.R != 0 {
		t.Fatalf("passenger=%d,%d", pass.Q, pass.R)
	}
}

func TestTransportPassengerViaMobility(t *testing.T) {
	m := abilityMatch(t, protocol.ArmyYellow, protocol.ArmyRed)
	put(m, "a", "yellow_carrier", 0, 0, 0)
	mover := put(m, "a", "yellow_courier", 1, 0, 0) // has mobility
	pass := put(m, "a", "yellow_crew", 0, 1, 0)

	q, r := 2, 0
	if err := m.UseMobility("a", mover.ID, &q, &r, nil, pass.ID); err != nil {
		t.Fatal(err)
	}
	if pass.Q != 1 || pass.R != 0 {
		t.Fatalf("passenger=%d,%d", pass.Q, pass.R)
	}
}

func TestDualStackDefs(t *testing.T) {
	p, err := tile.PackFor(protocol.ArmyYellow)
	if err != nil {
		t.Fatal(err)
	}
	d := p.Def("yellow_dual")
	if d == nil || !d.HasSpecial(tile.SpecialDualStack) {
		t.Fatal("yellow_dual missing")
	}
}
