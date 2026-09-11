package match

import (
	"crypto/rand"
	enchex "encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

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
	Board            map[string]*BoardTile // keyed by tile ID
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
	log              []string
	pendingReplay    *protocol.BattleReplay
	stepLogStart     int

	mobilityUsed      map[string]bool
	reconUsed         bool
	quartermasterUsed bool
	reconPeekFor      string
	reconPeek         []string
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
		mobilityUsed:  make(map[string]bool),
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

func (m *Match) Snapshot(viewerID string) protocol.MatchState {
	board := m.boardView(viewerID)
	players := make([]protocol.PlayerView, 0, 2)
	for _, p := range m.Players {
		handCount := len(p.Hand)
		var hand []protocol.HandTile
		var discard []protocol.HandTile
		if p.ID == viewerID {
			hand = make([]protocol.HandTile, len(p.Hand))
			for i, t := range p.Hand {
				kind := ""
				if d := p.Pack.Def(t.DefID); d != nil {
					kind = string(d.Kind)
				}
				hand[i] = protocol.HandTile{ID: t.ID, DefID: t.DefID, Kind: kind}
			}
			discard = make([]protocol.HandTile, len(p.Discard))
			for i, t := range p.Discard {
				kind := ""
				if d := p.Pack.Def(t.DefID); d != nil {
					kind = string(d.Kind)
				}
				discard[i] = protocol.HandTile{ID: t.ID, DefID: t.DefID, Kind: kind}
			}
		}
		players = append(players, protocol.PlayerView{
			ID: p.ID, Army: p.Army, HQHP: p.HQHP,
			Hand: hand, HandCount: handCount,
			DeckCount: len(p.Deck), DiscardCount: len(p.Discard),
			Discard: discard,
		})
	}
	logCopy := append([]string(nil), m.log...)
	var peek []string
	if viewerID == m.reconPeekFor && len(m.reconPeek) > 0 {
		peek = append([]string(nil), m.reconPeek...)
	}
	return protocol.MatchState{
		V:                      protocol.Version,
		Type:                   protocol.TypeMatchState,
		Phase:                  m.Phase,
		TurnPlayerID:           m.TurnPlayerID,
		MustDiscard:            m.MustDiscard,
		UnluckyAvailable:       m.UnluckyAvailable,
		Board:                  board,
		Players:                players,
		LegalHexes:             m.legalHexes(),
		Result:                 m.Result,
		Log:                    logCopy,
		EndMode:                endModeString(m.endMode),
		TieTurnsLeft:           m.tieTurnsLeft,
		Replay:                 m.pendingReplay,
		ReconPeek:              peek,
		ReconAvailable:         m.canUseRecon(viewerID),
		QuartermasterAvailable: m.canUseQuartermaster(viewerID),
	}
}

func (m *Match) ClearReplay() {
	m.pendingReplay = nil
}

func (m *Match) boardView(viewerID string) []protocol.BoardTile {
	netted := m.nettedSet()
	board := make([]protocol.BoardTile, 0, len(m.Board))
	for _, t := range m.Board {
		kind := ""
		var edges []protocol.EdgeMark
		maxHP, hp := 0, 0
		var effInits []int
		mobilityAvail := false
		if p := m.player(t.OwnerID); p != nil {
			if d := p.Pack.Def(t.DefID); d != nil {
				kind = string(d.Kind)
				edges = edgeMarks(d, t.Facing)
				if d.Kind != tile.KindHQ {
					maxHP = 1 + d.Toughness
					hp = maxHP - t.Wounds
					if hp < 0 {
						hp = 0
					}
				}
				effInits = m.effectiveInits(t, netted)
				if viewerID == m.TurnPlayerID && t.OwnerID == viewerID &&
					d.HasComponent(tile.CompMobility) && !m.mobilityUsed[t.ID] && !netted[t.ID] &&
					m.Phase == protocol.MatchTurn && !m.MustDiscard {
					mobilityAvail = true
				}
			}
		}
		board = append(board, protocol.BoardTile{
			ID: t.ID, DefID: t.DefID, Kind: kind, OwnerID: t.OwnerID,
			Q: t.Q, R: t.R, Facing: t.Facing, Wounds: t.Wounds,
			HP: hp, MaxHP: maxHP, Netted: netted[t.ID],
			EffectiveInitiatives: effInits, Edges: edges,
			MobilityAvailable: mobilityAvail,
		})
	}
	return board
}

func (m *Match) hqHPView() map[string]int {
	out := make(map[string]int, 2)
	for _, p := range m.Players {
		out[p.ID] = p.HQHP
	}
	return out
}

func (m *Match) beginStepLog() {
	m.stepLogStart = len(m.log)
}

func (m *Match) takeStepLog() []string {
	if m.stepLogStart > len(m.log) {
		m.stepLogStart = 0
	}
	out := append([]string(nil), m.log[m.stepLogStart:]...)
	m.stepLogStart = len(m.log)
	return out
}

func (m *Match) recordBattleStep(initiative int, label string) {
	logs := m.takeStepLog()
	if len(logs) == 0 {
		return
	}
	if m.pendingReplay == nil {
		m.pendingReplay = &protocol.BattleReplay{}
	}
	m.pendingReplay.Steps = append(m.pendingReplay.Steps, protocol.BattleStep{
		Initiative: initiative,
		Label:      label,
		Board:      m.boardView(""),
		HQHP:       m.hqHPView(),
		Log:        logs,
	})
}

func endModeString(m endMode) string {
	switch m {
	case endAwaitOpponent:
		return "await_opponent"
	case endFinalArmed:
		return "final_armed"
	case endTieBreak:
		return "tie_break"
	default:
		return ""
	}
}

func (m *Match) appendLog(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	m.log = append(m.log, msg)
	const maxLog = 20
	if len(m.log) > maxLog {
		m.log = m.log[len(m.log)-maxLog:]
	}
}

func shortDefID(defID string) string {
	for _, prefix := range []string{"red_", "blue_", "green_", "yellow_"} {
		if strings.HasPrefix(defID, prefix) {
			return strings.TrimPrefix(defID, prefix)
		}
	}
	return defID
}

func edgeMarks(d *tile.Def, facing int) []protocol.EdgeMark {
	var out []protocol.EdgeMark
	for _, c := range d.ComponentsOf(tile.CompArmor) {
		for _, rel := range c.Dirs {
			out = append(out, protocol.EdgeMark{
				Dir: (facing + rel) % 6, Kind: "armor",
			})
		}
	}
	for _, c := range d.ComponentsOf(tile.CompNet) {
		for _, rel := range c.Dirs {
			out = append(out, protocol.EdgeMark{
				Dir: (facing + rel) % 6, Kind: "net",
			})
		}
	}
	for _, c := range d.ComponentsOf(tile.CompAttack) {
		kind := "melee"
		if c.AttackType == tile.AttackRanged {
			kind = "ranged"
		}
		for _, rel := range c.Dirs {
			out = append(out, protocol.EdgeMark{
				Dir: (facing + rel) % 6, Kind: kind, Strength: c.Strength,
			})
		}
	}
	for _, c := range d.ComponentsOf(tile.CompModuleLink) {
		for _, rel := range c.Dirs {
			out = append(out, protocol.EdgeMark{
				Dir: (facing + rel) % 6, Kind: "module_link",
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
		if !m.hexOccupied(c.Q, c.R) {
			out = append(out, protocol.Hex{Q: c.Q, R: c.R})
		}
	}
	// Dual-stack placement targets when current player holds a dual_stack tile.
	p := m.player(m.TurnPlayerID)
	if p != nil && m.Phase == protocol.MatchTurn {
		hasDual := false
		for _, h := range p.Hand {
			if d := p.Pack.Def(h.DefID); d != nil && d.HasSpecial(tile.SpecialDualStack) {
				hasDual = true
				break
			}
		}
		if hasDual {
			seen := map[string]bool{}
			for _, h := range out {
				seen[key(h.Q, h.R)] = true
			}
			for _, t := range m.Board {
				if t.OwnerID != p.ID || m.isHQ(t) {
					continue
				}
				d := m.defOf(t)
				if d == nil || d.Kind != tile.KindWarrior || d.HasSpecial(tile.SpecialDualStack) {
					continue
				}
				if len(m.tilesAt(t.Q, t.R)) != 1 {
					continue
				}
				k := key(t.Q, t.R)
				if seen[k] {
					continue
				}
				seen[k] = true
				out = append(out, protocol.Hex{Q: t.Q, R: t.R})
			}
		}
	}
	return out
}

func (m *Match) tilesAt(q, r int) []*BoardTile {
	var out []*BoardTile
	for _, t := range m.Board {
		if t.Q == q && t.R == r {
			out = append(out, t)
		}
	}
	return out
}

func (m *Match) primaryAt(q, r int) *BoardTile {
	tiles := m.tilesAt(q, r)
	if len(tiles) == 0 {
		return nil
	}
	for _, t := range tiles {
		if d := m.defOf(t); d != nil && !d.HasSpecial(tile.SpecialDualStack) {
			return t
		}
	}
	return tiles[0]
}

func (m *Match) hexOccupied(q, r int) bool {
	return len(m.tilesAt(q, r)) > 0
}

func (m *Match) setTile(t *BoardTile) {
	m.Board[t.ID] = t
}

func (m *Match) removeTile(t *BoardTile) {
	delete(m.Board, t.ID)
}

func (m *Match) allHexesOccupied() bool {
	for _, c := range hex.BoardCells(boardRadius) {
		if !m.hexOccupied(c.Q, c.R) {
			return false
		}
	}
	return true
}

func (m *Match) canDualStackPlace(playerID string, placing *tile.Def, q, r int) bool {
	if !placing.HasSpecial(tile.SpecialDualStack) {
		return false
	}
	existing := m.tilesAt(q, r)
	if len(existing) != 1 {
		return false
	}
	other := existing[0]
	if other.OwnerID != playerID || m.isHQ(other) {
		return false
	}
	otherDef := m.defOf(other)
	if otherDef == nil || otherDef.Kind != tile.KindWarrior {
		return false
	}
	if otherDef.HasSpecial(tile.SpecialDualStack) {
		return false
	}
	return true
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
