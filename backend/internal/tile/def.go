package tile

import "github.com/Zwx-tech/szesciokonty/backend/internal/protocol"

type Kind string

const (
	KindHQ      Kind = "hq"
	KindWarrior Kind = "warrior"
	KindModule  Kind = "module"
	KindInstant Kind = "instant"
)

type CompType string

const (
	CompAttack     CompType = "attack"
	CompArmor      CompType = "armor"
	CompNet        CompType = "net"
	CompMobility   CompType = "mobility"
	CompModuleLink CompType = "module_link"
	CompModuleAura CompType = "module_aura"
	CompHQAura     CompType = "hq_aura"
	CompInstant    CompType = "instant"
	CompSpecial    CompType = "special"
)

type AttackType string

const (
	AttackMelee  AttackType = "melee"
	AttackRanged AttackType = "ranged"
)

type Instant string

const (
	InstantBattle    Instant = "battle"
	InstantMove      Instant = "move"
	InstantPush      Instant = "push"
	InstantSniper    Instant = "sniper"
	InstantAirStrike Instant = "airstrike"
	InstantGrenade   Instant = "grenade"
)

type EffectKind string

const (
	EffectInitBonus     EffectKind = "init_bonus"
	EffectMeleeBonus    EffectKind = "melee_bonus"
	EffectRangedBonus   EffectKind = "ranged_bonus"
	EffectMedic         EffectKind = "medic"
	EffectMother        EffectKind = "mother" // extra attack after last init
	EffectSaboteur      EffectKind = "saboteur"
	EffectScoper        EffectKind = "scoper"
	EffectRecon         EffectKind = "recon"
	EffectTransport     EffectKind = "transport"
	EffectQuartermaster EffectKind = "quartermaster"
	EffectExtraAttack   EffectKind = "extra_attack" // HQ: attack again after last init
)

// Special component ids for rare behaviors.
const (
	SpecialLineWound = "line_wound"
	SpecialDetonate  = "detonate"
	SpecialDualStack = "dual_stack"
	SpecialBlocker   = "blocker"
)

type Effect struct {
	Kind  EffectKind `json:"kind"`
	Value int        `json:"value,omitempty"`
}

// Component is a typed tile property. Unused fields stay empty.
type Component struct {
	Type CompType `json:"type"`

	AttackType AttackType `json:"attackType,omitempty"`
	Strength   int        `json:"strength,omitempty"`
	Dirs       []int      `json:"dirs,omitempty"`

	Instant Instant        `json:"instant,omitempty"`
	ID      string         `json:"id,omitempty"`
	Effects []Effect       `json:"effects,omitempty"`
	Params  map[string]int `json:"params,omitempty"`
}

// Def is an immutable tile prototype. Board/hand instances reference Def.ID.
type Def struct {
	ID          string      `json:"id"`
	Kind        Kind        `json:"kind"`
	Initiatives []int       `json:"initiatives,omitempty"`
	Toughness   int         `json:"toughness"` // extra HP beyond 1; HQ uses match HP track
	Components  []Component `json:"components"`
}

func (d *Def) IsInstant() bool { return d.Kind == KindInstant }
func (d *Def) IsUnit() bool {
	return d.Kind == KindHQ || d.Kind == KindWarrior || d.Kind == KindModule
}

func (d *Def) ComponentsOf(t CompType) []Component {
	var out []Component
	for _, c := range d.Components {
		if c.Type == t {
			out = append(out, c)
		}
	}
	return out
}

func (d *Def) HasComponent(t CompType) bool {
	for _, c := range d.Components {
		if c.Type == t {
			return true
		}
	}
	return false
}

func (d *Def) HasSpecial(id string) bool {
	for _, c := range d.ComponentsOf(CompSpecial) {
		if c.ID == id {
			return true
		}
	}
	return false
}

func (d *Def) InstantKind() Instant {
	for _, c := range d.ComponentsOf(CompInstant) {
		return c.Instant
	}
	return ""
}

type entry struct {
	def   *Def
	count int
}

type Pack struct {
	Army     protocol.Army
	HQ       *Def
	defs     map[string]*Def
	drawPile []entry
}

func (p *Pack) Def(id string) *Def { return p.defs[id] }

func (p *Pack) AllDefs() []*Def {
	out := make([]*Def, 0, len(p.defs))
	for _, d := range p.defs {
		out = append(out, d)
	}
	return out
}

func (p *Pack) DrawPileIDs() []string {
	out := make([]string, 0, 34)
	for _, e := range p.drawPile {
		for i := 0; i < e.count; i++ {
			out = append(out, e.def.ID)
		}
	}
	return out
}

func (p *Pack) TileCount() int {
	n := 1
	for _, e := range p.drawPile {
		n += e.count
	}
	return n
}

func newPack(army protocol.Army, hq *Def, rest []entry) *Pack {
	defs := map[string]*Def{hq.ID: hq}
	for _, e := range rest {
		defs[e.def.ID] = e.def
	}
	return &Pack{Army: army, HQ: hq, defs: defs, drawPile: rest}
}
