package database

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Edits the store's row of a spell for the test, restoring it afterwards.
func editRow(t *testing.T, id int32, edit func(*spelldata.Spell)) {
	t.Helper()
	row := spelldata.MustFind(id)
	original := *row
	edit(row)
	t.Cleanup(func() { *row = original })
}

func TestPotionClassesFollowTheirEffectsAndStats(t *testing.T) {
	const host int32 = 17531 // Any row in the store: each case states its own effects.
	onCaster := [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_CASTER}
	energize := func(power dbcenums.PowerType) spelldata.Effect {
		return spelldata.Effect{Type: dbcenums.E_ENERGIZE, Misc: int32(power), Target: onCaster}
	}
	heal := spelldata.Effect{Type: dbcenums.E_HEAL, Target: onCaster}
	absorb := spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_SCHOOL_ABSORB, Target: onCaster}
	damage := spelldata.Effect{Type: dbcenums.E_SCHOOL_DAMAGE, Target: [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_TARGET_ENEMY}}
	aura := func(kind dbcenums.EffectAuraType, misc int32) spelldata.Effect {
		return spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: kind, Misc: misc, Target: onCaster}
	}
	const other = proto.ConsumableType_ConsumableTypeOther
	const (
		resource  = proto.PotionClass_PotionClassResource
		defensive = proto.PotionClass_PotionClassDefensive
		offensive = proto.PotionClass_PotionClassOffensive
	)

	for _, c := range []struct {
		name    string
		kind    proto.ConsumableType
		effects []spelldata.Effect
		granted stats.Stats
		want    []proto.PotionClass
	}{
		{"mana", proto.ConsumableType_ConsumableTypePotion, []spelldata.Effect{energize(dbcenums.POWER_MANA)}, stats.Stats{}, []proto.PotionClass{resource}},
		{"a heal and mana", proto.ConsumableType_ConsumableTypePotion, []spelldata.Effect{heal, energize(dbcenums.POWER_MANA)}, stats.Stats{}, []proto.PotionClass{resource, defensive}},
		{"an absorb", proto.ConsumableType_ConsumableTypePotion, []spelldata.Effect{absorb}, stats.Stats{}, []proto.PotionClass{defensive}},
		{"rage and strength", proto.ConsumableType_ConsumableTypePotion, []spelldata.Effect{energize(dbcenums.POWER_RAGE)}, stats.Stats{stats.Strength: 1}, []proto.PotionClass{resource, offensive}},
		{"armor", proto.ConsumableType_ConsumableTypePotion, nil, stats.Stats{stats.Armor: 1}, []proto.PotionClass{defensive}},
		{"spell damage", proto.ConsumableType_ConsumableTypePotion, nil, stats.Stats{stats.SpellDamage: 1}, []proto.PotionClass{offensive}},
		{"damage to an enemy", proto.ConsumableType_ConsumableTypePotion, []spelldata.Effect{damage}, stats.Stats{}, []proto.PotionClass{offensive}},
		{"a flask", proto.ConsumableType_ConsumableTypeFlask, nil, stats.Stats{stats.Health: 1}, nil},
		{"a miscellaneous absorb", other, []spelldata.Effect{absorb}, stats.Stats{}, []proto.PotionClass{defensive}},
		{"miscellaneous Stamina, for its health", other, nil, stats.Stats{stats.Stamina: 1}, []proto.PotionClass{defensive}},
		{"miscellaneous resistances", other, nil, stats.Stats{stats.FireResistance: 1}, []proto.PotionClass{defensive}},
		{"miscellaneous dodge chance", other, []spelldata.Effect{aura(dbcenums.A_MOD_DODGE_PERCENT, 0)}, stats.Stats{}, []proto.PotionClass{defensive}},
		{"miscellaneous Defense skill", other, []spelldata.Effect{aura(dbcenums.A_MOD_SKILL, 95)}, stats.Stats{}, []proto.PotionClass{defensive}},
		{"another miscellaneous skill", other, []spelldata.Effect{aura(dbcenums.A_MOD_SKILL, 43)}, stats.Stats{}, []proto.PotionClass{offensive}},
		{"miscellaneous agility", other, nil, stats.Stats{stats.Agility: 1}, []proto.PotionClass{offensive}},
		{"miscellaneous melee haste the stats leave out", other, []spelldata.Effect{aura(dbcenums.A_MOD_MELEE_HASTE_3, 0)}, stats.Stats{}, []proto.PotionClass{offensive}},
	} {
		t.Run(c.name, func(t *testing.T) {
			editRow(t, host, func(s *spelldata.Spell) { s.Effects = c.effects })
			db := NewWowDatabase()
			db.Consumables[1] = &proto.Consumable{Id: 1, Type: c.kind, Stats: c.granted.ToProtoArray(),
				ItemEffects: []*proto.ItemEffect{{BuffId: host}}}
			if err := AddPotionClasses(db); err != nil {
				t.Fatal(err)
			}
			if got := db.Consumables[1].PotionClasses; !slices.Equal(got, c.want) {
				t.Errorf("classes %v, want %v", got, c.want)
			}
		})
	}
}

func TestPotionClassesCountMobTypeStats(t *testing.T) {
	const host int32 = 17531
	editRow(t, host, func(s *spelldata.Spell) { s.Effects = nil })
	db := NewWowDatabase()
	db.Consumables[1] = &proto.Consumable{Id: 1, Type: proto.ConsumableType_ConsumableTypePotion,
		MobTypeStats: []*proto.MobTypeStats{{MobTypes: []proto.MobType{proto.MobType_MobTypeDemon}, Stats: map[int32]float64{int32(stats.AttackPower): 1}}},
		ItemEffects:  []*proto.ItemEffect{{BuffId: host}}}
	if err := AddPotionClasses(db); err != nil {
		t.Fatal(err)
	}
	if got, want := db.Consumables[1].PotionClasses, []proto.PotionClass{proto.PotionClass_PotionClassOffensive}; !slices.Equal(got, want) {
		t.Errorf("classes %v, want %v", got, want)
	}
}
