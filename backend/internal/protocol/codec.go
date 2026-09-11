package protocol

import (
	"encoding/json"
	"fmt"
)

const (
	TypeCreateRoom       = "create_room"
	TypeJoinRoom         = "join_room"
	TypeLeave            = "leave"
	TypeSetArmy          = "set_army"
	TypeReady            = "ready"
	TypeStart            = "start"
	TypeDiscard          = "discard"
	TypePlace            = "place"
	TypePlayInstant      = "play_instant"
	TypeUseMobility      = "use_mobility"
	TypeUseRecon         = "use_recon"
	TypeUseQuartermaster = "use_quartermaster"
	TypeEndTurn          = "end_turn"
	TypeRedrawUnlucky    = "redraw_unlucky"
	TypeRematch          = "rematch"
	TypeReconnect        = "reconnect"
	TypeError            = "error"
	TypeRoomState        = "room_state"
	TypeMatchState       = "match_state"
)

type envelope struct {
	V    int    `json:"v"`
	Type string `json:"type"`
}

// ClientMessage is any decoded client intent.
type ClientMessage any

func DecodeClient(data []byte) (ClientMessage, error) {
	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	if env.V != Version {
		return nil, fmt.Errorf("unsupported version %d", env.V)
	}

	switch env.Type {
	case TypeCreateRoom:
		var m CreateRoom
		return m, json.Unmarshal(data, &m)
	case TypeJoinRoom:
		var m JoinRoom
		return m, json.Unmarshal(data, &m)
	case TypeLeave:
		var m Leave
		return m, json.Unmarshal(data, &m)
	case TypeSetArmy:
		var m SetArmy
		return m, json.Unmarshal(data, &m)
	case TypeReady:
		var m Ready
		return m, json.Unmarshal(data, &m)
	case TypeStart:
		var m Start
		return m, json.Unmarshal(data, &m)
	case TypeDiscard:
		var m Discard
		return m, json.Unmarshal(data, &m)
	case TypePlace:
		var m Place
		return m, json.Unmarshal(data, &m)
	case TypePlayInstant:
		var m PlayInstant
		return m, json.Unmarshal(data, &m)
	case TypeUseMobility:
		var m UseMobility
		return m, json.Unmarshal(data, &m)
	case TypeUseRecon:
		var m UseRecon
		return m, json.Unmarshal(data, &m)
	case TypeUseQuartermaster:
		var m UseQuartermaster
		return m, json.Unmarshal(data, &m)
	case TypeEndTurn:
		var m EndTurn
		return m, json.Unmarshal(data, &m)
	case TypeRedrawUnlucky:
		var m RedrawUnlucky
		return m, json.Unmarshal(data, &m)
	case TypeRematch:
		var m Rematch
		return m, json.Unmarshal(data, &m)
	case TypeReconnect:
		var m Reconnect
		return m, json.Unmarshal(data, &m)
	default:
		return nil, fmt.Errorf("unknown type %q", env.Type)
	}
}

func Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

func NewError(code, message string) ErrorMsg {
	return ErrorMsg{V: Version, Type: TypeError, Code: code, Message: message}
}
