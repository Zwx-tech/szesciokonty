package match

import (
	"crypto/rand"
	enchex "encoding/hex"
	"errors"
	"math/big"
	"strconv"

	"github.com/Zwx-tech/szesciokonty/backend/internal/hex"
	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
	"github.com/Zwx-tech/szesciokonty/backend/internal/tile"
)

var (
	ErrNotYourTurn = errors.New("not your turn")
	ErrBadPhase    = errors.New("invalid phase")
	ErrMustDiscard = errors.New("must discard first")
	ErrBadTile     = errors.New("invalid tile")
	ErrBadHex      = errors.New("invalid hex")
	ErrOccupied    = errors.New("hex occupied")
	ErrBadFacing   = errors.New("invalid facing")
	ErrNoUnlucky   = errors.New("unlucky redraw not available")
	ErrNotUnit     = errors.New("tile is not a unit")
	ErrHQDone      = errors.New("hq already placed")
)

const boardRadius = 2
const hqHP = 20

type TileInst struct {
	ID    string
	DefID string
}

type BoardTile struct {
	ID      string
	DefID   string
	OwnerID string
	Q, R    int
	Facing  int
	Wounds  int
}

type Player struct {
	ID      string
	Army    protocol.Army
	HQHP    int
	HQDefID string
	Deck    []TileInst
	Hand    []TileInst
	Discard []TileInst
	Pack    *tile.Pack
}

type Match struct {
	Phase            protocol.MatchPhase
	Players          [2]*Player
	Board            map[string]*BoardTile
	TurnPlayerID     string
	FirstPlayerID    string
	MustDiscard      bool
	UnluckyAvailable bool
	Result           *protocol.MatchResult
	openingStep      int
	drawnIDs         []string
	pendingDestroy   map[string]bool
	endMode          endMode
	deckExhaustedBy  string
	tieTurnsLeft     int
}

type Seat struct {
	ID   string
	Army protocol.Army
}

func New(a, b Seat) (*Match, error) {
	pa, err := newPlayer(a)
	if err != nil {
		return nil, err
	}
	pb, err := newPlayer(b)
	if err != nil {
		return nil, err
	}

	first := 0
	n, err := rand.Int(rand.Reader, big.NewInt(2))
	if err != nil {
		return nil, err
	}
	if n.Int64() == 1 {
		first = 1
	}
	players := [2]*Player{pa, pb}
	return &Match{
		Phase:         protocol.MatchPlaceHQ,
		Players:       players,
		Board:         make(map[string]*BoardTile),
		TurnPlayerID:  players[first].ID,
		FirstPlayerID: players[first].ID,
	}, nil
}

func newPlayer(s Seat) (*Player, error) {
	pack, err := tile.PackFor(s.Army)
	if err != nil {
		return nil, err
	}
	ids := pack.DrawPileIDs()
	shuffle(ids)
	deck := make([]TileInst, len(ids))
	for i, defID := range ids {
		deck[i] = TileInst{ID: newID(), DefID: defID}
	}
	return &Player{
		ID:      s.ID,
		Army:    s.Army,
		HQHP:    hqHP,
		HQDefID: pack.HQ.ID,
		Deck:    deck,
		Pack:    pack,
	}, nil
}

func (m *Match) player(id string) *Player {
	for _, p := range m.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (m *Match) Snapshot() protocol.MatchState {
	board := make([]protocol.BoardTile, 0, len(m.Board))
	for _, t := range m.Board {
		kind := ""
		var edges []protocol.EdgeMark
		if p := m.player(t.OwnerID); p != nil {
			if d := p.Pack.Def(t.DefID); d != nil {
				kind = string(d.Kind)
				edges = edgeMarks(d, t.Facing)
			}
		}
		board = append(board, protocol.BoardTile{
			ID: t.ID, DefID: t.DefID, Kind: kind, OwnerID: t.OwnerID,
			Q: t.Q, R: t.R, Facing: t.Facing, Wounds: t.Wounds, Edges: edges,
		})
	}
	players := make([]protocol.PlayerView, 0, 2)
	for _, p := range m.Players {
		hand := make([]protocol.HandTile, len(p.Hand))
		for i, t := range p.Hand {
			kind := ""
			if d := p.Pack.Def(t.DefID); d != nil {
				kind = string(d.Kind)
			}
			hand[i] = protocol.HandTile{ID: t.ID, DefID: t.DefID, Kind: kind}
		}
		players = append(players, protocol.PlayerView{
			ID: p.ID, Army: p.Army, HQHP: p.HQHP,
			Hand: hand, DeckCount: len(p.Deck), DiscardCount: len(p.Discard),
		})
	}
	return protocol.MatchState{
		V:                protocol.Version,
		Type:             protocol.TypeMatchState,
		Phase:            m.Phase,
		TurnPlayerID:     m.TurnPlayerID,
		MustDiscard:      m.MustDiscard,
		UnluckyAvailable: m.UnluckyAvailable,
		Board:            board,
		Players:          players,
		LegalHexes:       m.legalHexes(),
		Result:           m.Result,
	}
}

func edgeMarks(d *tile.Def, facing int) []protocol.EdgeMark {
	var out []protocol.EdgeMark
	for _, c := range d.ComponentsOf(tile.CompAttack) {
		kind := "melee"
		if c.AttackType == tile.AttackRanged {
			kind = "ranged"
		}
		for _, rel := range c.Dirs {
			out = append(out, protocol.EdgeMark{
				Dir:  (facing + rel) % 6,
				Kind: kind,
			})
		}
	}
	for _, c := range d.ComponentsOf(tile.CompNet) {
		for _, rel := range c.Dirs {
			out = append(out, protocol.EdgeMark{
				Dir:  (facing + rel) % 6,
				Kind: "net",
			})
		}
	}
	return out
}

func (m *Match) legalHexes() []protocol.Hex {
	if m.Phase != protocol.MatchPlaceHQ && m.Phase != protocol.MatchTurn {
		return nil
	}
	out := make([]protocol.Hex, 0, 19)
	for _, c := range hex.BoardCells(boardRadius) {
		if _, ok := m.Board[key(c.Q, c.R)]; !ok {
			out = append(out, protocol.Hex{Q: c.Q, R: c.R})
		}
	}
	return out
}

func key(q, r int) string {
	return strconv.Itoa(q) + "," + strconv.Itoa(r)
}

func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return enchex.EncodeToString(b[:])
}

func shuffle(ids []string) {
	for i := len(ids) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := i
		if err == nil {
			j = int(n.Int64())
		}
		ids[i], ids[j] = ids[j], ids[i]
	}
}
