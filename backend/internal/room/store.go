package room

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Zwx-tech/szesciokonty/backend/internal/match"
	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
)

var (
	ErrNotFound    = errors.New("room not found")
	ErrFull        = errors.New("room full")
	ErrStarted     = errors.New("game already started")
	ErrNotInRoom   = errors.New("not in a room")
	ErrNotHost     = errors.New("only host can start")
	ErrBadArmy     = errors.New("invalid or taken army")
	ErrCannotStart = errors.New("cannot start yet")
	ErrBadToken    = errors.New("invalid reconnect token")
	ErrBadName     = errors.New("invalid name")
	ErrNoRematch   = errors.New("rematch not available")
)

const DisconnectGrace = 30 * time.Second

type Client interface {
	Send(msg any)
}

type Player struct {
	ID        string
	Token     string
	Name      string
	Army      protocol.Army
	Ready     bool
	Connected bool
	Host      bool
	Client    Client
}

type Room struct {
	Code  string
	Phase protocol.RoomPhase
	Host  *Player
	Guest *Player
	Match *match.Match
}

type Store struct {
	mu       sync.Mutex
	rooms    map[string]*Room
	forfeits map[string]*time.Timer // key: code|playerID
}

func NewStore() *Store {
	return &Store{
		rooms:    make(map[string]*Room),
		forfeits: make(map[string]*time.Timer),
	}
}

func (s *Store) Create(name string, client Client) (*Room, *Player, error) {
	name, err := cleanName(name)
	if err != nil {
		return nil, nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	code := s.uniqueCode()
	p := newPlayer(name, true, client)
	r := &Room{Code: code, Phase: protocol.PhaseLobby, Host: p}
	s.rooms[code] = r
	return r, p, nil
}

func (s *Store) Join(code, name string, client Client) (*Room, *Player, error) {
	name, err := cleanName(name)
	if err != nil {
		return nil, nil, err
	}
	code = normalizeCode(code)

	s.mu.Lock()
	defer s.mu.Unlock()

	r, ok := s.rooms[code]
	if !ok {
		return nil, nil, ErrNotFound
	}
	if r.Phase != protocol.PhaseLobby {
		return nil, nil, ErrStarted
	}
	if r.Guest != nil {
		return nil, nil, ErrFull
	}

	p := newPlayer(name, false, client)
	r.Guest = p
	return r, p, nil
}

func (s *Store) Reconnect(code, token string, client Client) (*Room, *Player, error) {
	code = normalizeCode(code)

	s.mu.Lock()
	defer s.mu.Unlock()

	r, ok := s.rooms[code]
	if !ok {
		return nil, nil, ErrNotFound
	}
	p := r.playerByToken(token)
	if p == nil {
		return nil, nil, ErrBadToken
	}
	s.cancelForfeitLocked(code, p.ID)
	p.Connected = true
	p.Client = client
	return r, p, nil
}

func (s *Store) Leave(code, playerID string) (closed bool, remaining *Room, err error) {
	code = normalizeCode(code)

	s.mu.Lock()
	defer s.mu.Unlock()

	r, ok := s.rooms[code]
	if !ok {
		return false, nil, ErrNotFound
	}
	p := r.playerByID(playerID)
	if p == nil {
		return false, nil, ErrNotInRoom
	}

	s.cancelForfeitLocked(code, playerID)

	if r.Match != nil && r.Match.Phase != protocol.MatchEnded {
		r.Match.Forfeit(playerID)
		s.broadcastMatchLocked(r)
	}

	if p.Host {
		s.notify(r, protocol.NewError("room_closed", "host left"))
		s.clearForfeitsLocked(code)
		delete(s.rooms, code)
		return true, nil, nil
	}

	r.Guest = nil
	r.Host.Ready = false
	if r.Phase == protocol.PhaseMatch && r.Match != nil && r.Match.Phase == protocol.MatchEnded {
		// keep ended match for host result view
		return false, r, nil
	}
	if r.Phase == protocol.PhaseMatch {
		r.Phase = protocol.PhaseLobby
		r.Match = nil
	}
	return false, r, nil
}

func (s *Store) Disconnect(code, playerID string) *Room {
	code = normalizeCode(code)

	s.mu.Lock()
	defer s.mu.Unlock()

	r, ok := s.rooms[code]
	if !ok {
		return nil
	}
	p := r.playerByID(playerID)
	if p == nil {
		return nil
	}
	p.Connected = false
	p.Client = nil
	if r.Match != nil && r.Match.Phase != protocol.MatchEnded {
		s.scheduleForfeitLocked(code, playerID)
	}
	return r
}

func (s *Store) SetArmy(code, playerID string, army protocol.Army) (*Room, error) {
	if !validArmy(army) {
		return nil, ErrBadArmy
	}
	code = normalizeCode(code)

	s.mu.Lock()
	defer s.mu.Unlock()

	r, p, err := s.seat(code, playerID)
	if err != nil {
		return nil, err
	}
	if r.Phase != protocol.PhaseLobby {
		return nil, ErrStarted
	}
	other := r.other(p)
	if other != nil && other.Army == army {
		return nil, ErrBadArmy
	}
	p.Army = army
	p.Ready = false
	return r, nil
}

func (s *Store) SetReady(code, playerID string, ready bool) (*Room, error) {
	code = normalizeCode(code)

	s.mu.Lock()
	defer s.mu.Unlock()

	r, p, err := s.seat(code, playerID)
	if err != nil {
		return nil, err
	}
	if r.Phase != protocol.PhaseLobby {
		return nil, ErrStarted
	}
	if ready && p.Army == "" {
		return nil, ErrCannotStart
	}
	p.Ready = ready
	return r, nil
}

func (s *Store) Start(code, playerID string) (*Room, error) {
	code = normalizeCode(code)

	s.mu.Lock()
	defer s.mu.Unlock()

	r, p, err := s.seat(code, playerID)
	if err != nil {
		return nil, err
	}
	if !p.Host {
		return nil, ErrNotHost
	}
	if r.Phase != protocol.PhaseLobby {
		return nil, ErrStarted
	}
	if !r.canStart() {
		return nil, ErrCannotStart
	}
	m, err := match.New(
		match.Seat{ID: r.Host.ID, Army: r.Host.Army},
		match.Seat{ID: r.Guest.ID, Army: r.Guest.Army},
	)
	if err != nil {
		return nil, err
	}
	r.Match = m
	r.Phase = protocol.PhaseMatch
	return r, nil
}

func (s *Store) MatchPlayInstant(code, playerID, tileID string, q, r *int, facing *int, targetTileID string) (*Room, error) {
	return s.withMatch(code, func(m *match.Match) error {
		return m.PlayInstant(playerID, tileID, q, r, facing, targetTileID)
	})
}

func (s *Store) MatchDiscard(code, playerID, tileID string) (*Room, error) {
	return s.withMatch(code, func(m *match.Match) error {
		return m.Discard(playerID, tileID)
	})
}

func (s *Store) MatchPlace(code, playerID, tileID string, q, r, facing int) (*Room, error) {
	return s.withMatch(code, func(m *match.Match) error {
		return m.Place(playerID, tileID, q, r, facing)
	})
}

func (s *Store) MatchEndTurn(code, playerID string) (*Room, error) {
	return s.withMatch(code, func(m *match.Match) error {
		return m.EndTurn(playerID)
	})
}

func (s *Store) MatchRedrawUnlucky(code, playerID string) (*Room, error) {
	return s.withMatch(code, func(m *match.Match) error {
		return m.RedrawUnlucky(playerID)
	})
}

func (s *Store) Rematch(code, playerID string) (*Room, error) {
	code = normalizeCode(code)
	s.mu.Lock()
	defer s.mu.Unlock()

	r, p, err := s.seat(code, playerID)
	if err != nil {
		return nil, err
	}
	_ = p
	if r.Phase != protocol.PhaseMatch || r.Match == nil || r.Match.Phase != protocol.MatchEnded {
		return nil, ErrNoRematch
	}
	r.Match = nil
	r.Phase = protocol.PhaseLobby
	r.Host.Ready = false
	if r.Guest != nil {
		r.Guest.Ready = false
	}
	return r, nil
}

func forfeitKey(code, playerID string) string {
	return code + "|" + playerID
}

func (s *Store) scheduleForfeitLocked(code, playerID string) {
	key := forfeitKey(code, playerID)
	if t := s.forfeits[key]; t != nil {
		t.Stop()
	}
	s.forfeits[key] = time.AfterFunc(DisconnectGrace, func() {
		s.applyForfeit(code, playerID)
	})
}

func (s *Store) cancelForfeitLocked(code, playerID string) {
	key := forfeitKey(code, playerID)
	if t := s.forfeits[key]; t != nil {
		t.Stop()
		delete(s.forfeits, key)
	}
}

func (s *Store) clearForfeitsLocked(code string) {
	prefix := code + "|"
	for k, t := range s.forfeits {
		if strings.HasPrefix(k, prefix) {
			t.Stop()
			delete(s.forfeits, k)
		}
	}
}

func (s *Store) applyForfeit(code, playerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := forfeitKey(code, playerID)
	delete(s.forfeits, key)

	r, ok := s.rooms[code]
	if !ok || r.Match == nil || r.Match.Phase == protocol.MatchEnded {
		return
	}
	p := r.playerByID(playerID)
	if p == nil || p.Connected {
		return
	}
	r.Match.Forfeit(playerID)
	s.broadcastLocked(r)
	s.broadcastMatchLocked(r)
}

func (s *Store) withMatch(code string, fn func(*match.Match) error) (*Room, error) {
	code = normalizeCode(code)
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[code]
	if !ok {
		return nil, ErrNotFound
	}
	if room.Match == nil {
		return nil, ErrNotInRoom
	}
	if room.Match.Phase == protocol.MatchEnded {
		return nil, match.ErrBadPhase
	}
	if err := fn(room.Match); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *Store) Broadcast(r *Room) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.broadcastLocked(r)
}

func (s *Store) BroadcastAll(r *Room) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.broadcastLocked(r)
	s.broadcastMatchLocked(r)
}

func (s *Store) broadcastMatchLocked(r *Room) {
	if r.Match == nil {
		return
	}
	msg := r.Match.Snapshot()
	for _, p := range r.players() {
		if p.Client == nil {
			continue
		}
		p.Client.Send(msg)
	}
}

func (s *Store) broadcastLocked(r *Room) {
	for _, p := range r.players() {
		if p.Client == nil {
			continue
		}
		p.Client.Send(r.stateFor(p))
	}
}

func (s *Store) notify(r *Room, msg any) {
	for _, p := range r.players() {
		if p.Client == nil {
			continue
		}
		p.Client.Send(msg)
	}
}

func (s *Store) seat(code, playerID string) (*Room, *Player, error) {
	r, ok := s.rooms[code]
	if !ok {
		return nil, nil, ErrNotFound
	}
	p := r.playerByID(playerID)
	if p == nil {
		return nil, nil, ErrNotInRoom
	}
	return r, p, nil
}

func (s *Store) uniqueCode() string {
	for {
		code := randomCode(5)
		if _, ok := s.rooms[code]; !ok {
			return code
		}
	}
}

func (r *Room) canStart() bool {
	return r.Guest != nil &&
		r.Host.Army != "" && r.Guest.Army != "" &&
		r.Host.Ready && r.Guest.Ready &&
		r.Host.Army != r.Guest.Army
}

func (r *Room) stateFor(p *Player) protocol.RoomState {
	seats := make([]protocol.Seat, 0, 2)
	for _, s := range r.players() {
		seats = append(seats, seatView(s))
	}
	return protocol.RoomState{
		V:        protocol.Version,
		Type:     protocol.TypeRoomState,
		Code:     r.Code,
		Phase:    r.Phase,
		You:      seatView(p),
		Token:    p.Token,
		Seats:    seats,
		CanStart: r.canStart(),
	}
}

func seatView(p *Player) protocol.Seat {
	return protocol.Seat{
		ID:        p.ID,
		Name:      p.Name,
		Army:      p.Army,
		Ready:     p.Ready,
		Connected: p.Connected,
		Host:      p.Host,
	}
}

func (r *Room) players() []*Player {
	out := make([]*Player, 0, 2)
	if r.Host != nil {
		out = append(out, r.Host)
	}
	if r.Guest != nil {
		out = append(out, r.Guest)
	}
	return out
}

func (r *Room) playerByID(id string) *Player {
	for _, p := range r.players() {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (r *Room) playerByToken(token string) *Player {
	for _, p := range r.players() {
		if p.Token == token {
			return p
		}
	}
	return nil
}

func (r *Room) other(p *Player) *Player {
	if r.Host != nil && r.Host.ID == p.ID {
		return r.Guest
	}
	return r.Host
}

func newPlayer(name string, host bool, client Client) *Player {
	return &Player{
		ID:        randomID(),
		Token:     randomID(),
		Name:      name,
		Connected: true,
		Host:      host,
		Client:    client,
	}
}

func cleanName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 16 {
		return "", ErrBadName
	}
	return name, nil
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func validArmy(a protocol.Army) bool {
	switch a {
	case protocol.ArmyRed, protocol.ArmyBlue, protocol.ArmyGreen, protocol.ArmyYellow:
		return true
	default:
		return false
	}
}

func randomID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func randomCode(n int) string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = codeAlphabet[int(b[i])%len(codeAlphabet)]
	}
	return string(out)
}
