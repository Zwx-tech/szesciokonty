package match

import "github.com/Zwx-tech/szesciokonty/backend/internal/protocol"

type endMode int

const (
	endNone          endMode = iota
	endAwaitOpponent         // last tile drawn; finish turn, then opponent plays once
	endFinalArmed            // next battle is the Final Battle (or force on opponent end-turn)
	endTieBreak              // one extra turn each, then one more battle
)

func (m *Match) Forfeit(loserID string) {
	if m.Phase == protocol.MatchEnded {
		return
	}
	winner := m.other(loserID)
	m.Phase = protocol.MatchEnded
	m.TurnPlayerID = ""
	m.Result = &protocol.MatchResult{WinnerID: winner.ID}
}

func (m *Match) noteDeckExhausted(p *Player) {
	if len(p.Deck) > 0 || m.deckExhaustedBy != "" {
		return
	}
	m.deckExhaustedBy = p.ID
	if m.endMode == endNone {
		m.endMode = endAwaitOpponent
	}
}

func (m *Match) afterPlayerEndedTurn(playerID string) {
	switch m.endMode {
	case endAwaitOpponent:
		if playerID == m.deckExhaustedBy {
			m.endMode = endFinalArmed
		}
		m.advanceTurn()
	case endFinalArmed:
		if playerID != m.deckExhaustedBy {
			m.resolveBattle()
			return
		}
		m.advanceTurn()
	case endTieBreak:
		m.tieTurnsLeft--
		if m.tieTurnsLeft <= 0 {
			m.resolveBattle()
			return
		}
		m.advanceTurn()
	default:
		m.advanceTurn()
	}
}

func (m *Match) endByScore() {
	a, b := m.Players[0], m.Players[1]
	switch {
	case a.HQHP > b.HQHP:
		m.Phase = protocol.MatchEnded
		m.TurnPlayerID = ""
		m.Result = &protocol.MatchResult{WinnerID: a.ID}
	case b.HQHP > a.HQHP:
		m.Phase = protocol.MatchEnded
		m.TurnPlayerID = ""
		m.Result = &protocol.MatchResult{WinnerID: b.ID}
	case m.endMode == endTieBreak:
		m.Phase = protocol.MatchEnded
		m.TurnPlayerID = ""
		m.Result = &protocol.MatchResult{Draw: true}
	default:
		// Tie after Final Battle → one turn each, then another battle.
		m.endMode = endTieBreak
		m.tieTurnsLeft = 2
		m.Phase = protocol.MatchTurn
		m.MustDiscard = false
		m.UnluckyAvailable = false
		m.advanceTurn()
	}
}
