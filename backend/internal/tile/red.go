package tile

import "github.com/Zwx-tech/szesciokonty/backend/internal/protocol"

func redPack() *Pack {
	hq := def("red_hq", KindHQ, []int{0}, 0,
		melee(1, 0, 1, 2, 3, 4, 5),
		hqAura(eff(EffectInitBonus, 1)),
	)
	fighter := warrior("red_fighter", []int{2}, melee(1, 0))
	slasher := warrior("red_slasher", []int{3}, melee(1, 0, 1))
	netter := warrior("red_netter", []int{2}, net(0), melee(3, 0))
	plated := def("red_plated", KindWarrior, []int{1}, 1, armor(0, 1), melee(1, 0))
	bruiser := warrior("red_bruiser", []int{2}, melee(2, 0))
	stalker := warrior("red_stalker", []int{3}, ranged(1, 0), mobility())
	medic := module("red_medic", []int{0, 1, 2}, Effect{Kind: EffectMedic})
	officer := module("red_officer", []int{0, 1, 2}, eff(EffectMeleeBonus, 1))
	officerHeavy := def("red_officer_heavy", KindModule, nil, 1,
		moduleLink(0, 1, 2, 3),
		moduleAura(eff(EffectMeleeBonus, 1)),
	)
	tempo := module("red_tempo", []int{0, 1, 2}, eff(EffectInitBonus, 1))

	return newPack(protocol.ArmyRed, hq, []entry{
		{fighter, 6},
		{slasher, 4},
		{netter, 2},
		{plated, 1},
		{bruiser, 2},
		{stalker, 2},
		{medic, 1},
		{officer, 2},
		{officerHeavy, 1},
		{tempo, 2},
		{instant("red_battle", InstantBattle), 6},
		{instant("red_move", InstantMove), 4},
		{instant("red_grenade", InstantGrenade), 1},
	})
}
