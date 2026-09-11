package match

import (
	"testing"

	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
	"github.com/Zwx-tech/szesciokonty/backend/internal/tile"
)

func battleFixture(t *testing.T, a, b protocol.Army) *Match {
	t.Helper()
	m, err := New(Seat{ID: "a", Army: a}, Seat{ID: "b", Army: b})
	if err != nil {
		t.Fatal(err)
	}
	m.Phase = protocol.MatchTurn
	m.Board = map[string]*BoardTile{}
	// Park HQs off the fight so auras do not skew fixtures.
	put(m, "a", m.player("a").HQDefID, -2, 2, 0)
	put(m, "b", m.player("b").HQDefID, 2, -2, 0)
	return m
}

func put(m *Match, owner, defID string, q, r, facing int) *BoardTile {
	bt := &BoardTile{
		ID: newID(), DefID: defID, OwnerID: owner,
		Q: q, R: r, Facing: facing,
	}
	m.setTile(bt)
	return bt
}

func TestBattleMutualMeleeDestroy(t *testing.T) {
	m := battleFixture(t, protocol.ArmyRed, protocol.ArmyRed)
	fa := put(m, "a", "red_fighter", 0, 0, 0)
	fb := put(m, "b", "red_fighter", 1, 0, 3)

	m.resolveBattle()

	if m.boardByID(fa.ID) != nil || m.boardByID(fb.ID) != nil {
		t.Fatal("both fighters should die simultaneously")
	}
	if m.Phase != protocol.MatchTurn {
		t.Fatalf("phase %s", m.Phase)
	}
}

func TestBattleReplaySteps(t *testing.T) {
	m := battleFixture(t, protocol.ArmyRed, protocol.ArmyRed)
	// Both fighters init 2 — mutual melee at initiative 2.
	fa := put(m, "a", "red_fighter", 0, 0, 0)
	fb := put(m, "b", "red_fighter", 1, 0, 3)

	m.resolveBattle()

	if m.pendingReplay == nil || len(m.pendingReplay.Steps) == 0 {
		t.Fatal("expected battle replay steps")
	}
	step := m.pendingReplay.Steps[0]
	if step.Initiative != 2 {
		t.Fatalf("first step initiative=%d want 2", step.Initiative)
	}
	if step.Label != "Initiative 2" {
		t.Fatalf("label=%q", step.Label)
	}
	// After simultaneous destroy, board step should not still list the fighters.
	for _, bt := range step.Board {
		if bt.ID == fa.ID || bt.ID == fb.ID {
			t.Fatal("fighters should be gone in post-phase snapshot")
		}
	}
	if len(step.Log) == 0 {
		t.Fatal("step should include hit/destroy log lines")
	}

	snap := m.Snapshot("a")
	if snap.Replay == nil || len(snap.Replay.Steps) == 0 {
		t.Fatal("snapshot should include replay")
	}
	m.ClearReplay()
	snap2 := m.Snapshot("a")
	if snap2.Replay != nil {
		t.Fatal("replay should clear")
	}
}

func TestBattleReplayMultiInit(t *testing.T) {
	m := battleFixture(t, protocol.ArmyGreen, protocol.ArmyGreen)
	// green_heavy has inits 1,2 — acts twice; soft target dies on first hit.
	put(m, "a", "green_heavy", 0, 0, 0)
	target := put(m, "b", "green_courier", 1, 0, 3)

	m.resolveBattle()

	if m.pendingReplay == nil {
		t.Fatal("nil replay")
	}
	inits := make([]int, 0, len(m.pendingReplay.Steps))
	for _, s := range m.pendingReplay.Steps {
		inits = append(inits, s.Initiative)
	}
	if len(inits) < 1 {
		t.Fatalf("steps=%v", inits)
	}
	// Highest initiative first.
	if inits[0] < 2 {
		t.Fatalf("expected init 2 first, got %v", inits)
	}
	_ = target
}

func TestBattleArmorBlocksRanged1(t *testing.T) {
	m := battleFixture(t, protocol.ArmyBlue, protocol.ArmyBlue)
	target := put(m, "a", "blue_plated_gunner", 0, 0, 0)
	put(m, "b", "blue_gunner", 1, 0, 3)

	m.resolveBattle()

	if m.boardByID(target.ID) == nil {
		t.Fatal("armor should block strength-1 ranged")
	}
	if target.Wounds != 0 {
		t.Fatalf("wounds=%d", target.Wounds)
	}
}

func TestBattleNetPreventsAttack(t *testing.T) {
	m := battleFixture(t, protocol.ArmyRed, protocol.ArmyBlue)
	victim := put(m, "a", "red_fighter", 0, 0, 0)
	put(m, "b", "blue_netter", 1, 0, 3)

	m.resolveBattle()

	if m.boardByID(victim.ID) == nil {
		t.Fatal("netted fighter should not die")
	}
}

func TestBattleModuleMeleeBonus(t *testing.T) {
	m := battleFixture(t, protocol.ArmyRed, protocol.ArmyRed)
	put(m, "a", "red_officer", 0, 0, 0) // links 0 → (1,0)
	put(m, "a", "red_fighter", 1, 0, 0)
	target := put(m, "b", "red_plated", 2, 0, 3)

	m.resolveBattle()

	if m.boardByID(target.ID) != nil {
		t.Fatal("melee+1 should destroy toughness-1 target")
	}
}

func TestBattleHQNeverDamagesHQ(t *testing.T) {
	m := battleFixture(t, protocol.ArmyRed, protocol.ArmyRed)
	// move HQs adjacent facing each other
	for _, bt := range m.Board {
		if m.isHQ(bt) {
			m.removeTile(bt)
		}
	}
	put(m, "a", "red_hq", 0, 0, 0)
	put(m, "b", "red_hq", 1, 0, 3)

	m.resolveBattle()

	if m.Players[0].HQHP != hqHP || m.Players[1].HQHP != hqHP {
		t.Fatalf("hq hp changed: %d %d", m.Players[0].HQHP, m.Players[1].HQHP)
	}
}

func TestBattleHQDamageEndsGame(t *testing.T) {
	m := battleFixture(t, protocol.ArmyRed, protocol.ArmyRed)
	for _, bt := range m.Board {
		if bt.OwnerID == "b" && m.isHQ(bt) {
			m.removeTile(bt)
		}
	}
	put(m, "b", "red_hq", 2, 0, 0)
	m.player("b").HQHP = 1
	put(m, "a", "red_bruiser", 1, 0, 0)

	m.resolveBattle()

	if m.Phase != protocol.MatchEnded {
		t.Fatalf("phase=%s", m.Phase)
	}
	if m.Result == nil || m.Result.WinnerID != "a" {
		t.Fatalf("result=%v", m.Result)
	}
}

func TestBattleToughnessSurvivesOneWound(t *testing.T) {
	m := battleFixture(t, protocol.ArmyRed, protocol.ArmyRed)
	target := put(m, "b", "red_plated", 1, 0, 3)
	put(m, "a", "red_fighter", 0, 0, 0)

	m.resolveBattle()

	if m.boardByID(target.ID) == nil {
		t.Fatal("toughness 1 should survive one wound")
	}
	if target.Wounds != 1 {
		t.Fatalf("wounds=%d", target.Wounds)
	}
}

func TestBattleMedicAbsorbs(t *testing.T) {
	m := battleFixture(t, protocol.ArmyRed, protocol.ArmyRed)
	fighter := put(m, "a", "red_fighter", 0, 0, 0)
	medic := put(m, "a", "red_medic", -1, 0, 0) // link0 → fighter
	put(m, "b", "red_fighter", 1, 0, 3)

	m.resolveBattle()

	if m.boardByID(fighter.ID) == nil {
		t.Fatal("medic should save fighter")
	}
	if m.boardByID(medic.ID) != nil {
		t.Fatal("medic should be discarded")
	}
}

func TestPackDefsExist(t *testing.T) {
	for _, army := range []protocol.Army{protocol.ArmyRed, protocol.ArmyBlue, protocol.ArmyGreen, protocol.ArmyYellow} {
		p, err := tile.PackFor(army)
		if err != nil {
			t.Fatal(err)
		}
		if p.Def(p.HQ.ID) == nil {
			t.Fatal(army)
		}
	}
}
