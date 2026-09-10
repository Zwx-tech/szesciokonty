package match

import (
	"testing"

	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
)

func TestFinalBattleAfterOpponentTurn(t *testing.T) {
	m := battleFixture(t, protocol.ArmyRed, protocol.ArmyBlue)
	m.Phase = protocol.MatchTurn
	m.TurnPlayerID = "a"
	m.openingStep = 5
	m.player("a").Deck = nil
	m.player("b").Deck = m.player("b").Deck[:1]
	m.noteDeckExhausted(m.player("a"))
	if m.endMode != endAwaitOpponent {
		t.Fatal(m.endMode)
	}

	if err := m.EndTurn("a"); err != nil {
		t.Fatal(err)
	}
	if m.endMode != endFinalArmed || m.TurnPlayerID != "b" {
		t.Fatalf("mode=%d turn=%s", m.endMode, m.TurnPlayerID)
	}

	m.player("a").HQHP = 10
	m.player("b").HQHP = 7
	if err := m.EndTurn("b"); err != nil {
		t.Fatal(err)
	}
	if m.Phase != protocol.MatchEnded {
		t.Fatalf("phase=%s", m.Phase)
	}
	if m.Result == nil || m.Result.WinnerID != "a" {
		t.Fatalf("result=%v", m.Result)
	}
}

func TestFinalBattleTieBreakThenDraw(t *testing.T) {
	m := battleFixture(t, protocol.ArmyRed, protocol.ArmyBlue)
	m.Phase = protocol.MatchTurn
	m.TurnPlayerID = "b"
	m.openingStep = 5
	m.endMode = endFinalArmed
	m.deckExhaustedBy = "a"
	m.player("a").HQHP = 5
	m.player("b").HQHP = 5

	if err := m.EndTurn("b"); err != nil {
		t.Fatal(err)
	}
	if m.Phase != protocol.MatchTurn || m.endMode != endTieBreak || m.tieTurnsLeft != 2 {
		t.Fatalf("phase=%s mode=%d left=%d", m.Phase, m.endMode, m.tieTurnsLeft)
	}

	// two tie-break turns, then battle → still tied → draw
	m.MustDiscard = false
	m.player("a").Deck = nil
	m.player("b").Deck = nil
	if err := m.EndTurn(m.TurnPlayerID); err != nil {
		t.Fatal(err)
	}
	if m.tieTurnsLeft != 1 {
		t.Fatalf("left=%d", m.tieTurnsLeft)
	}
	m.MustDiscard = false
	if err := m.EndTurn(m.TurnPlayerID); err != nil {
		t.Fatal(err)
	}
	if m.Phase != protocol.MatchEnded || m.Result == nil || !m.Result.Draw {
		t.Fatalf("phase=%s result=%v", m.Phase, m.Result)
	}
}

func TestForfeit(t *testing.T) {
	m := battleFixture(t, protocol.ArmyRed, protocol.ArmyBlue)
	m.Phase = protocol.MatchTurn
	m.Forfeit("a")
	if m.Phase != protocol.MatchEnded || m.Result.WinnerID != "b" {
		t.Fatalf("%v %v", m.Phase, m.Result)
	}
}
