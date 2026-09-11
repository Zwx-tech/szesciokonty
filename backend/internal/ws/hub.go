package ws

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/Zwx-tech/szesciokonty/backend/internal/match"
	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
	"github.com/Zwx-tech/szesciokonty/backend/internal/room"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

type Hub struct {
	rooms *room.Store
}

func NewHub() *Hub {
	return &Hub{rooms: room.NewStore()}
}

func (h *Hub) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade: %v", err)
			return
		}
		s := &session{hub: h, conn: conn}
		defer s.close()

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := s.handle(data); err != nil {
				log.Printf("ws handle: %v", err)
				return
			}
		}
	})
}

type session struct {
	hub      *Hub
	conn     *websocket.Conn
	writeMu  sync.Mutex
	roomCode string
	playerID string
}

func (s *session) Send(msg any) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	b, err := protocol.Encode(msg)
	if err != nil {
		return
	}
	_ = s.conn.WriteMessage(websocket.TextMessage, b)
}

func (s *session) close() {
	if s.roomCode != "" && s.playerID != "" {
		if r := s.hub.rooms.Disconnect(s.roomCode, s.playerID); r != nil {
			s.hub.rooms.BroadcastAll(r)
		}
	}
	_ = s.conn.Close()
}

func (s *session) handle(data []byte) error {
	msg, err := protocol.DecodeClient(data)
	if err != nil {
		s.Send(protocol.NewError("bad_request", err.Error()))
		return nil
	}

	switch m := msg.(type) {
	case protocol.CreateRoom:
		s.create(m.Name)
	case protocol.JoinRoom:
		s.join(m.Name, m.Code)
	case protocol.Reconnect:
		s.reconnect(m.Code, m.PlayerToken)
	case protocol.Leave:
		s.leave()
	case protocol.SetArmy:
		s.setArmy(m.Army)
	case protocol.Ready:
		s.setReady(m.Ready)
	case protocol.Start:
		s.start()
	case protocol.Discard:
		s.discard(m.TileID)
	case protocol.Place:
		s.place(m.TileID, m.Q, m.R, m.Facing)
	case protocol.EndTurn:
		s.endTurn()
	case protocol.RedrawUnlucky:
		s.redrawUnlucky()
	case protocol.PlayInstant:
		s.playInstant(m)
	case protocol.UseMobility:
		s.useMobility(m)
	case protocol.UseRecon:
		s.useRecon()
	case protocol.UseQuartermaster:
		s.useQuartermaster(m.TileID)
	case protocol.Rematch:
		s.rematch()
	default:
		s.Send(protocol.NewError("not_implemented", "handler not wired yet"))
	}
	return nil
}

func (s *session) create(name string) {
	if s.playerID != "" {
		s.Send(protocol.NewError("already_in_room", "leave before creating another"))
		return
	}
	r, p, err := s.hub.rooms.Create(name, s)
	if err != nil {
		s.fail(err)
		return
	}
	s.roomCode, s.playerID = r.Code, p.ID
	s.hub.rooms.Broadcast(r)
}

func (s *session) join(name, code string) {
	if s.playerID != "" {
		s.Send(protocol.NewError("already_in_room", "leave before joining another"))
		return
	}
	r, p, err := s.hub.rooms.Join(code, name, s)
	if err != nil {
		s.fail(err)
		return
	}
	s.roomCode, s.playerID = r.Code, p.ID
	s.hub.rooms.Broadcast(r)
}

func (s *session) reconnect(code, token string) {
	if s.playerID != "" {
		s.Send(protocol.NewError("already_in_room", "already seated"))
		return
	}
	r, p, err := s.hub.rooms.Reconnect(code, token, s)
	if err != nil {
		s.fail(err)
		return
	}
	s.roomCode, s.playerID = r.Code, p.ID
	s.hub.rooms.BroadcastAll(r)
}

func (s *session) leave() {
	if s.playerID == "" {
		s.Send(protocol.NewError("not_in_room", room.ErrNotInRoom.Error()))
		return
	}
	code, id := s.roomCode, s.playerID
	s.roomCode, s.playerID = "", ""
	closed, rem, err := s.hub.rooms.Leave(code, id)
	if err != nil {
		s.fail(err)
		return
	}
	if !closed && rem != nil {
		s.hub.rooms.BroadcastAll(rem)
	}
}

func (s *session) setArmy(army protocol.Army) {
	r, err := s.hub.rooms.SetArmy(s.roomCode, s.playerID, army)
	if err != nil {
		s.fail(err)
		return
	}
	s.hub.rooms.Broadcast(r)
}

func (s *session) setReady(ready bool) {
	r, err := s.hub.rooms.SetReady(s.roomCode, s.playerID, ready)
	if err != nil {
		s.fail(err)
		return
	}
	s.hub.rooms.Broadcast(r)
}

func (s *session) start() {
	r, err := s.hub.rooms.Start(s.roomCode, s.playerID)
	if err != nil {
		s.fail(err)
		return
	}
	s.hub.rooms.BroadcastAll(r)
}

func (s *session) discard(tileID string) {
	r, err := s.hub.rooms.MatchDiscard(s.roomCode, s.playerID, tileID)
	if err != nil {
		s.fail(err)
		return
	}
	s.hub.rooms.BroadcastAll(r)
}

func (s *session) place(tileID string, q, rCoord, facing int) {
	roomRef, err := s.hub.rooms.MatchPlace(s.roomCode, s.playerID, tileID, q, rCoord, facing)
	if err != nil {
		s.fail(err)
		return
	}
	s.hub.rooms.BroadcastAll(roomRef)
}

func (s *session) endTurn() {
	r, err := s.hub.rooms.MatchEndTurn(s.roomCode, s.playerID)
	if err != nil {
		s.fail(err)
		return
	}
	s.hub.rooms.BroadcastAll(r)
}

func (s *session) redrawUnlucky() {
	r, err := s.hub.rooms.MatchRedrawUnlucky(s.roomCode, s.playerID)
	if err != nil {
		s.fail(err)
		return
	}
	s.hub.rooms.BroadcastAll(r)
}

func (s *session) playInstant(m protocol.PlayInstant) {
	r, err := s.hub.rooms.MatchPlayInstant(
		s.roomCode, s.playerID, m.TileID, m.Q, m.R, m.Facing, m.TargetTileID, m.PassengerID,
	)
	if err != nil {
		s.fail(err)
		return
	}
	s.hub.rooms.BroadcastAll(r)
}

func (s *session) useMobility(m protocol.UseMobility) {
	r, err := s.hub.rooms.MatchUseMobility(
		s.roomCode, s.playerID, m.TileID, m.Q, m.R, m.Facing, m.PassengerID,
	)
	if err != nil {
		s.fail(err)
		return
	}
	s.hub.rooms.BroadcastAll(r)
}

func (s *session) useRecon() {
	r, err := s.hub.rooms.MatchUseRecon(s.roomCode, s.playerID)
	if err != nil {
		s.fail(err)
		return
	}
	s.hub.rooms.BroadcastAll(r)
}

func (s *session) useQuartermaster(tileID string) {
	r, err := s.hub.rooms.MatchUseQuartermaster(s.roomCode, s.playerID, tileID)
	if err != nil {
		s.fail(err)
		return
	}
	s.hub.rooms.BroadcastAll(r)
}

func (s *session) rematch() {
	r, err := s.hub.rooms.Rematch(s.roomCode, s.playerID)
	if err != nil {
		s.fail(err)
		return
	}
	s.hub.rooms.Broadcast(r)
}

func (s *session) fail(err error) {
	code := "error"
	switch {
	case errors.Is(err, room.ErrNotFound):
		code = "not_found"
	case errors.Is(err, room.ErrFull):
		code = "full"
	case errors.Is(err, room.ErrStarted):
		code = "started"
	case errors.Is(err, room.ErrNotInRoom):
		code = "not_in_room"
	case errors.Is(err, room.ErrNotHost):
		code = "not_host"
	case errors.Is(err, room.ErrBadArmy):
		code = "bad_army"
	case errors.Is(err, room.ErrCannotStart):
		code = "cannot_start"
	case errors.Is(err, room.ErrBadToken):
		code = "bad_token"
	case errors.Is(err, room.ErrBadName):
		code = "bad_name"
	case errors.Is(err, match.ErrNotYourTurn):
		code = "not_your_turn"
	case errors.Is(err, match.ErrBadPhase):
		code = "bad_phase"
	case errors.Is(err, match.ErrMustDiscard):
		code = "must_discard"
	case errors.Is(err, match.ErrBadTile):
		code = "bad_tile"
	case errors.Is(err, match.ErrBadHex):
		code = "bad_hex"
	case errors.Is(err, match.ErrOccupied):
		code = "occupied"
	case errors.Is(err, match.ErrBadFacing):
		code = "bad_facing"
	case errors.Is(err, match.ErrNoUnlucky):
		code = "no_unlucky"
	case errors.Is(err, match.ErrNotUnit):
		code = "not_unit"
	case errors.Is(err, match.ErrHQDone):
		code = "hq_done"
	case errors.Is(err, match.ErrNotInstant):
		code = "not_instant"
	case errors.Is(err, match.ErrBadTarget):
		code = "bad_target"
	case errors.Is(err, match.ErrNoBattle):
		code = "no_battle"
	case errors.Is(err, match.ErrNoMobility):
		code = "no_mobility"
	case errors.Is(err, match.ErrNoRecon):
		code = "no_recon"
	case errors.Is(err, match.ErrNoQuartermaster):
		code = "no_quartermaster"
	case errors.Is(err, match.ErrBadPassenger):
		code = "bad_passenger"
	case errors.Is(err, room.ErrNoRematch):
		code = "no_rematch"
	}
	s.Send(protocol.NewError(code, err.Error()))
}
