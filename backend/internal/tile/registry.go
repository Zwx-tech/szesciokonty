package tile

import (
	"fmt"

	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
)

var packs = map[protocol.Army]*Pack{}

func init() {
	for _, p := range []*Pack{redPack(), bluePack(), greenPack(), yellowPack()} {
		packs[p.Army] = p
	}
}

func PackFor(army protocol.Army) (*Pack, error) {
	p, ok := packs[army]
	if !ok {
		return nil, fmt.Errorf("unknown army %q", army)
	}
	return p, nil
}

func DefByID(id string) *Def {
	for _, p := range packs {
		if d := p.Def(id); d != nil {
			return d
		}
	}
	return nil
}

func Armies() []protocol.Army {
	return []protocol.Army{
		protocol.ArmyRed,
		protocol.ArmyBlue,
		protocol.ArmyGreen,
		protocol.ArmyYellow,
	}
}
