package ws

import (
	"log"
	"net/http"
	"sync"

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
			s.hub.rooms.Broadcast(r)
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
	s.hub.rooms.Broadcast(r)
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
		s.hub.rooms.Broadcast(rem)
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
	s.hub.rooms.Broadcast(r)
}

func (s *session) fail(err error) {
	code := "error"
	switch err {
	case room.ErrNotFound:
		code = "not_found"
	case room.ErrFull:
		code = "full"
	case room.ErrStarted:
		code = "started"
	case room.ErrNotInRoom:
		code = "not_in_room"
	case room.ErrNotHost:
		code = "not_host"
	case room.ErrBadArmy:
		code = "bad_army"
	case room.ErrCannotStart:
		code = "cannot_start"
	case room.ErrBadToken:
		code = "bad_token"
	case room.ErrBadName:
		code = "bad_name"
	}
	s.Send(protocol.NewError(code, err.Error()))
}
