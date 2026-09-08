package tile

import "github.com/Zwx-tech/szesciokonty/backend/internal/protocol"

func greenPack() *Pack {
	hq := def("green_hq", KindHQ, []int{0}, 0,
		melee(1, 0, 1, 2, 3, 4, 5),
		hqAura(Effect{Kind: EffectExtraAttack}),
	)
	courier := warrior("green_courier", []int{2}, melee(1, 0), mobility())
	heavy := warrior("green_heavy", []int{1, 2}, ranged(1, 0))
	marksman := warrior("green_marksman", []int{3}, ranged(1, 0))
	blaster := warrior("green_blaster", []int{2}, ranged(1, 0))
	rover := warrior("green_rover", []int{2, 3},
		ranged(1, 0),
		melee(1, 0),
		mobility(),
	)
	bruiser := warrior("green_bruiser", []int{2}, melee(1, 0))
	disruptor := module("green_disruptor", []int{0, 1, 2}, eff(EffectSaboteur, 1))
	pathfinder := module("green_pathfinder", []int{0, 1, 2, 3, 4, 5}, Effect{Kind: EffectRecon})
	medic := module("green_medic", []int{0, 1, 2}, Effect{Kind: EffectMedic})
	officer := module("green_officer", []int{0, 1, 2}, eff(EffectRangedBonus, 1))
	hijack := module("green_hijack", []int{0}, Effect{Kind: EffectScoper})
	tempo := module("green_tempo", []int{0, 1, 2}, eff(EffectInitBonus, 1))

	return newPack(protocol.ArmyGreen, hq, []entry{
		{courier, 2},
		{heavy, 1},
		{marksman, 5},
		{blaster, 2},
		{rover, 1},
		{bruiser, 1},
		{disruptor, 1},
		{pathfinder, 1},
		{medic, 2},
		{officer, 1},
		{hijack, 1},
		{tempo, 2},
		{instant("green_battle", InstantBattle), 6},
		{instant("green_move", InstantMove), 7},
		{instant("green_shot", InstantSniper), 1},
	})
}
