package tile

import (
	"testing"

	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
)

func TestPacksComplete(t *testing.T) {
	for _, army := range Armies() {
		p, err := PackFor(army)
		if err != nil {
			t.Fatal(err)
		}
		if p.TileCount() != 35 {
			t.Fatalf("%s: want 35 tiles, got %d", army, p.TileCount())
		}
		pile := p.DrawPileIDs()
		if len(pile) != 34 {
			t.Fatalf("%s: draw pile %d", army, len(pile))
		}
		if p.HQ == nil || p.HQ.Kind != KindHQ || !p.HQ.HasComponent(CompHQAura) {
			t.Fatalf("%s: bad HQ", army)
		}
		for _, id := range pile {
			d := p.Def(id)
			if d == nil {
				t.Fatalf("%s: missing def %s", army, id)
			}
			if d.Kind == KindHQ {
				t.Fatalf("%s: HQ in draw pile", army)
			}
			if d.Kind == KindInstant && d.InstantKind() == "" {
				t.Fatalf("%s: instant %s missing kind", army, id)
			}
			if d.Kind == KindWarrior && len(d.Initiatives) == 0 && !d.HasSpecial(SpecialBlocker) {
				t.Fatalf("%s: warrior %s has no initiative", army, id)
			}
			if len(d.Components) == 0 && d.Kind != KindWarrior {
				t.Fatalf("%s: %s has no components", army, id)
			}
		}
	}
}

func TestNettersCompose(t *testing.T) {
	red, _ := PackFor(protocol.ArmyRed)
	nf := red.Def("red_netter")
	if !nf.HasComponent(CompNet) || len(nf.ComponentsOf(CompAttack)) != 1 {
		t.Fatalf("red netter: %+v", nf.Components)
	}
	if nf.ComponentsOf(CompAttack)[0].Strength != 3 {
		t.Fatal("expected melee 3")
	}

	yellow, _ := PackFor(protocol.ArmyYellow)
	nm := yellow.Def("yellow_net_chief")
	if !nm.HasComponent(CompNet) || nm.ComponentsOf(CompAttack)[0].Strength != 1 {
		t.Fatalf("yellow net chief: %+v", nm.Components)
	}

	blue, _ := PackFor(protocol.ArmyBlue)
	bn := blue.Def("blue_netter")
	if !bn.HasComponent(CompNet) || bn.HasComponent(CompAttack) {
		t.Fatalf("blue netter should be net-only: %+v", bn.Components)
	}
}

func TestUniqueArmies(t *testing.T) {
	seen := map[protocol.Army]bool{}
	for _, a := range Armies() {
		if seen[a] {
			t.Fatal("dup army")
		}
		seen[a] = true
	}
	if len(seen) != 4 {
		t.Fatal(len(seen))
	}
}
