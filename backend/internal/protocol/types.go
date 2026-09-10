package protocol

const Version = 1

type Army string

const (
	ArmyRed    Army = "red"
	ArmyBlue   Army = "blue"
	ArmyGreen  Army = "green"
	ArmyYellow Army = "yellow"
)

type RoomPhase string

const (
	PhaseLobby RoomPhase = "lobby"
	PhaseMatch RoomPhase = "match"
)

type MatchPhase string

const (
	MatchPlaceHQ MatchPhase = "place_hq"
	MatchTurn    MatchPhase = "turn"
	MatchBattle  MatchPhase = "battle"
	MatchEnded   MatchPhase = "ended"
)

// Client → server

type CreateRoom struct {
	V    int    `json:"v"`
	Type string `json:"type"` // create_room
	Name string `json:"name"`
}

type JoinRoom struct {
	V    int    `json:"v"`
	Type string `json:"type"` // join_room
	Name string `json:"name"`
	Code string `json:"code"`
}

type Leave struct {
	V    int    `json:"v"`
	Type string `json:"type"` // leave
}

type SetArmy struct {
	V    int    `json:"v"`
	Type string `json:"type"` // set_army
	Army Army   `json:"army"`
}

type Ready struct {
	V     int    `json:"v"`
	Type  string `json:"type"` // ready
	Ready bool   `json:"ready"`
}

type Start struct {
	V    int    `json:"v"`
	Type string `json:"type"` // start
}

type Discard struct {
	V      int    `json:"v"`
	Type   string `json:"type"` // discard
	TileID string `json:"tileId"`
}

type Place struct {
	V      int    `json:"v"`
	Type   string `json:"type"` // place
	TileID string `json:"tileId"`
	Q      int    `json:"q"`
	R      int    `json:"r"`
	Facing int    `json:"facing"` // 0..5
}

type PlayInstant struct {
	V            int    `json:"v"`
	Type         string `json:"type"` // play_instant
	TileID       string `json:"tileId"`
	Q            *int   `json:"q,omitempty"`
	R            *int   `json:"r,omitempty"`
	Facing       *int   `json:"facing,omitempty"`
	TargetTileID string `json:"targetTileId,omitempty"`
}

type EndTurn struct {
	V    int    `json:"v"`
	Type string `json:"type"` // end_turn
}

type RedrawUnlucky struct {
	V    int    `json:"v"`
	Type string `json:"type"` // redraw_unlucky
}

type Rematch struct {
	V    int    `json:"v"`
	Type string `json:"type"` // rematch
}

type Reconnect struct {
	V           int    `json:"v"`
	Type        string `json:"type"` // reconnect
	Code        string `json:"code"`
	PlayerToken string `json:"playerToken"`
}

// Server → client

type ErrorMsg struct {
	V       int    `json:"v"`
	Type    string `json:"type"` // error
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Seat struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Army      Army   `json:"army,omitempty"`
	Ready     bool   `json:"ready"`
	Connected bool   `json:"connected"`
	Host      bool   `json:"host"`
}

type RoomState struct {
	V     int       `json:"v"`
	Type  string    `json:"type"` // room_state
	Code  string    `json:"code"`
	Phase RoomPhase `json:"phase"`
	You   Seat      `json:"you"`
	// Token is only for the recipient; used to reconnect.
	Token    string `json:"token"`
	Seats    []Seat `json:"seats"`
	CanStart bool   `json:"canStart"`
}

type Hex struct {
	Q int `json:"q"`
	R int `json:"r"`
}

type BoardTile struct {
	ID      string     `json:"id"`
	DefID   string     `json:"defId"`
	Kind    string     `json:"kind"`
	OwnerID string     `json:"ownerId"`
	Q       int        `json:"q"`
	R       int        `json:"r"`
	Facing  int        `json:"facing"`
	Wounds  int        `json:"wounds"`
	Edges   []EdgeMark `json:"edges,omitempty"`
}

// EdgeMark is an absolute board-side icon (0..5) for rendering.
type EdgeMark struct {
	Dir  int    `json:"dir"`
	Kind string `json:"kind"` // melee | ranged | net
}

type HandTile struct {
	ID    string `json:"id"`
	DefID string `json:"defId"`
	Kind  string `json:"kind"`
}

type PlayerView struct {
	ID           string     `json:"id"`
	Army         Army       `json:"army"`
	HQHP         int        `json:"hqHp"`
	Hand         []HandTile `json:"hand"`
	DeckCount    int        `json:"deckCount"`
	DiscardCount int        `json:"discardCount"`
}

type MatchResult struct {
	WinnerID string `json:"winnerId,omitempty"` // empty + Draw=true → draw
	Draw     bool   `json:"draw,omitempty"`
}

type MatchState struct {
	V                 int          `json:"v"`
	Type              string       `json:"type"` // match_state
	Phase             MatchPhase   `json:"phase"`
	TurnPlayerID      string       `json:"turnPlayerId,omitempty"`
	MustDiscard       bool         `json:"mustDiscard"`
	UnluckyAvailable  bool         `json:"unluckyAvailable"`
	Board             []BoardTile  `json:"board"`
	Players           []PlayerView `json:"players"`
	LegalHexes        []Hex        `json:"legalHexes,omitempty"`
	Result            *MatchResult `json:"result,omitempty"`
}
