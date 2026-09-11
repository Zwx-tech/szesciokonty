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
	PassengerID  string `json:"passengerTileId,omitempty"`
}

type UseMobility struct {
	V           int    `json:"v"`
	Type        string `json:"type"` // use_mobility
	TileID      string `json:"tileId"`
	Q           *int   `json:"q,omitempty"`
	R           *int   `json:"r,omitempty"`
	Facing      *int   `json:"facing,omitempty"`
	PassengerID string `json:"passengerTileId,omitempty"`
}

type UseRecon struct {
	V    int    `json:"v"`
	Type string `json:"type"` // use_recon
}

type UseQuartermaster struct {
	V      int    `json:"v"`
	Type   string `json:"type"` // use_quartermaster
	TileID string `json:"tileId"`
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
	ID                   string     `json:"id"`
	DefID                string     `json:"defId"`
	Kind                 string     `json:"kind"`
	OwnerID              string     `json:"ownerId"`
	Q                    int        `json:"q"`
	R                    int        `json:"r"`
	Facing               int        `json:"facing"`
	Wounds               int        `json:"wounds"`
	HP                   int        `json:"hp,omitempty"`
	MaxHP                int        `json:"maxHp,omitempty"`
	Netted               bool       `json:"netted,omitempty"`
	EffectiveInitiatives []int      `json:"effectiveInitiatives,omitempty"`
	Edges                []EdgeMark `json:"edges,omitempty"`
	MobilityAvailable    bool       `json:"mobilityAvailable,omitempty"`
}

// EdgeMark is an absolute board-side icon (0..5) for rendering / inspect.
type EdgeMark struct {
	Dir      int    `json:"dir"`
	Kind     string `json:"kind"` // melee | ranged | net | armor | module_link
	Strength int    `json:"strength,omitempty"`
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
	Hand         []HandTile `json:"hand,omitempty"`
	HandCount    int        `json:"handCount"`
	DeckCount    int        `json:"deckCount"`
	DiscardCount int        `json:"discardCount"`
	Discard      []HandTile `json:"discard,omitempty"` // viewer-only
}

type MatchResult struct {
	WinnerID string `json:"winnerId,omitempty"` // empty + Draw=true → draw
	Draw     bool   `json:"draw,omitempty"`
}

// BattleStep is one initiative phase (or extra-attack pass) for client playback.
type BattleStep struct {
	Initiative int            `json:"initiative"` // -1 = extra attacks
	Label      string         `json:"label"`
	Board      []BoardTile    `json:"board"`
	HQHP       map[string]int `json:"hqHp"`
	Log        []string       `json:"log"`
}

type BattleReplay struct {
	Steps []BattleStep `json:"steps"`
}

type MatchState struct {
	V                      int           `json:"v"`
	Type                   string        `json:"type"` // match_state
	Phase                  MatchPhase    `json:"phase"`
	TurnPlayerID           string        `json:"turnPlayerId,omitempty"`
	MustDiscard            bool          `json:"mustDiscard"`
	UnluckyAvailable       bool          `json:"unluckyAvailable"`
	Board                  []BoardTile   `json:"board"`
	Players                []PlayerView  `json:"players"`
	LegalHexes             []Hex         `json:"legalHexes,omitempty"`
	Result                 *MatchResult  `json:"result,omitempty"`
	Log                    []string      `json:"log,omitempty"`
	EndMode                string        `json:"endMode,omitempty"`
	TieTurnsLeft           int           `json:"tieTurnsLeft,omitempty"`
	Replay                 *BattleReplay `json:"replay,omitempty"`
	ReconPeek              []string      `json:"reconPeek,omitempty"`
	ReconAvailable         bool          `json:"reconAvailable,omitempty"`
	QuartermasterAvailable bool          `json:"quartermasterAvailable,omitempty"`
}
