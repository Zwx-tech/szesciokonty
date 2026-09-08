package tile

func attack(t AttackType, str int, dirs ...int) Component {
	return Component{Type: CompAttack, AttackType: t, Strength: str, Dirs: dirs}
}

func melee(str int, dirs ...int) Component {
	return attack(AttackMelee, str, dirs...)
}

func ranged(str int, dirs ...int) Component {
	return attack(AttackRanged, str, dirs...)
}

func armor(dirs ...int) Component {
	return Component{Type: CompArmor, Dirs: dirs}
}

func net(dirs ...int) Component {
	return Component{Type: CompNet, Dirs: dirs}
}

func mobility() Component {
	return Component{Type: CompMobility}
}

func moduleLink(dirs ...int) Component {
	return Component{Type: CompModuleLink, Dirs: dirs}
}

func moduleAura(effects ...Effect) Component {
	return Component{Type: CompModuleAura, Effects: effects}
}

func hqAura(effects ...Effect) Component {
	return Component{Type: CompHQAura, Effects: effects}
}

func instantComp(kind Instant) Component {
	return Component{Type: CompInstant, Instant: kind}
}

func special(id string, params map[string]int) Component {
	return Component{Type: CompSpecial, ID: id, Params: params}
}

func eff(kind EffectKind, value int) Effect {
	return Effect{Kind: kind, Value: value}
}

func def(id string, kind Kind, inits []int, toughness int, comps ...Component) *Def {
	return &Def{ID: id, Kind: kind, Initiatives: inits, Toughness: toughness, Components: comps}
}

func warrior(id string, inits []int, comps ...Component) *Def {
	return def(id, KindWarrior, inits, 0, comps...)
}

func module(id string, linkDirs []int, effects ...Effect) *Def {
	comps := []Component{moduleLink(linkDirs...), moduleAura(effects...)}
	return def(id, KindModule, nil, 0, comps...)
}

func instant(id string, kind Instant) *Def {
	return def(id, KindInstant, nil, 0, instantComp(kind))
}
