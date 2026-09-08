package tile

import "github.com/Zwx-tech/szesciokonty/backend/internal/protocol"

func bluePack() *Pack {
	hq := def("blue_hq", KindHQ, []int{0}, 0,
		melee(1, 0, 1, 2, 3, 4, 5),
		hqAura(eff(EffectRangedBonus, 1)),
	)
	bulwark := def("blue_bulwark", KindWarrior, nil, 1,
		armor(0, 1, 5),
		special(SpecialBlocker, nil),
	)
	gunner := warrior("blue_gunner", []int{2}, ranged(1, 0))
	piercer := def("blue_piercer", KindWarrior, []int{1}, 1,
		special(SpecialLineWound, map[string]int{"strength": 1}),
	)
	bastion := def("blue_bastion", KindWarrior, []int{1}, 1,
		armor(0, 1, 5),
		ranged(1, 0),
		melee(1, 0),
	)
	chaser := warrior("blue_chaser", []int{2}, melee(1, 0))
	sentry := def("blue_sentry", KindWarrior, []int{1}, 1, ranged(1, 0))
	platedChaser := warrior("blue_plated_chaser", []int{2}, armor(0), melee(1, 0))
	platedGunner := warrior("blue_plated_gunner", []int{1}, armor(0), ranged(1, 0))
	bomber := def("blue_bomber", KindWarrior, []int{2}, 1,
		melee(1, 0),
		special(SpecialDetonate, nil),
	)
	blade := warrior("blue_blade", []int{3}, melee(1, 0))
	nexus := module("blue_nexus", []int{0, 1, 2},
		eff(EffectMeleeBonus, 1),
		eff(EffectRangedBonus, 1),
	)
	netter := warrior("blue_netter", []int{1}, net(0))
	burst := def("blue_burst", KindWarrior, []int{1, 2}, 1, ranged(1, 0))
	medic := module("blue_medic", []int{0, 1, 2}, Effect{Kind: EffectMedic})
	catalyst := module("blue_catalyst", []int{0, 1, 2}, Effect{Kind: EffectMother})
	spotter := warrior("blue_spotter", []int{1}, ranged(1, 0))
	officer := module("blue_officer", []int{0, 1, 2}, eff(EffectRangedBonus, 1))
	tempo := module("blue_tempo", []int{0, 1, 2}, eff(EffectInitBonus, 1))

	return newPack(protocol.ArmyBlue, hq, []entry{
		{bulwark, 2},
		{gunner, 2},
		{piercer, 1},
		{bastion, 1},
		{chaser, 2},
		{sentry, 1},
		{platedChaser, 2},
		{platedGunner, 1},
		{bomber, 1},
		{blade, 1},
		{nexus, 1},
		{netter, 1},
		{burst, 1},
		{medic, 2},
		{catalyst, 1},
		{spotter, 1},
		{officer, 1},
		{tempo, 1},
		{instant("blue_battle", InstantBattle), 4},
		{instant("blue_move", InstantMove), 1},
		{instant("blue_push", InstantPush), 5},
		{instant("blue_blast", InstantAirStrike), 1},
	})
}
