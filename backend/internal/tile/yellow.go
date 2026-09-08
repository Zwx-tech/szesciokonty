package tile

import "github.com/Zwx-tech/szesciokonty/backend/internal/protocol"

func yellowPack() *Pack {
	hq := def("yellow_hq", KindHQ, []int{0}, 0,
		melee(1, 0, 1, 2, 3, 4, 5),
		hqAura(eff(EffectMeleeBonus, 1)),
	)
	courier := warrior("yellow_courier", []int{3}, melee(1, 0), mobility())
	thumper := warrior("yellow_thumper", []int{2}, melee(1, 0))
	crew := warrior("yellow_crew", []int{2}, melee(1, 0))
	netter := warrior("yellow_netter", []int{1}, net(0))
	warden := def("yellow_warden", KindWarrior, []int{1}, 1, melee(1, 0))
	tempo := module("yellow_tempo", []int{0, 1, 2}, eff(EffectInitBonus, 1))
	netChief := warrior("yellow_net_chief", []int{2}, net(0), melee(1, 0))
	dual := warrior("yellow_dual", []int{3},
		ranged(1, 0),
		melee(1, 0),
		special(SpecialDualStack, nil),
	)
	chief := module("yellow_chief", []int{0, 1, 2},
		eff(EffectMeleeBonus, 1),
		eff(EffectInitBonus, 1),
	)
	converter := module("yellow_converter", []int{0, 1, 2}, Effect{Kind: EffectQuartermaster})
	officerA := module("yellow_officer_a", []int{0, 1, 2}, eff(EffectMeleeBonus, 1))
	officerB := module("yellow_officer_b", []int{0, 1, 2}, eff(EffectMeleeBonus, 1))
	carrier := module("yellow_carrier", []int{0, 1, 2}, Effect{Kind: EffectTransport})
	champion := def("yellow_champion", KindWarrior, []int{2}, 1, armor(0), melee(1, 0))

	return newPack(protocol.ArmyYellow, hq, []entry{
		{courier, 3},
		{thumper, 1},
		{crew, 4},
		{netter, 2},
		{warden, 1},
		{tempo, 1},
		{netChief, 1},
		{dual, 3},
		{chief, 1},
		{converter, 1},
		{officerA, 2},
		{officerB, 1},
		{carrier, 1},
		{champion, 1},
		{instant("yellow_battle", InstantBattle), 5},
		{instant("yellow_move", InstantMove), 3},
		{instant("yellow_shot", InstantSniper), 1},
		{instant("yellow_push", InstantPush), 2},
	})
}
