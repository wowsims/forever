package spelldata

import (
	"math"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestMobTypesReadsTheCreatureTypeMask(t *testing.T) {
	for mask, want := range map[int32][]proto.MobType{
		4:   {proto.MobType_MobTypeDemon},
		32:  {proto.MobType_MobTypeUndead},
		36:  {proto.MobType_MobTypeDemon, proto.MobType_MobTypeUndead},
		128: nil,
		256: {proto.MobType_MobTypeMechanical},
	} {
		if got := MobTypes(mask); !slices.Equal(got, want) {
			t.Errorf("MobTypes(%d) = %v, want %v", mask, got, want)
		}
	}
}

func TestAVersusAuraFollowsItsAuraOnTheAttackTables(t *testing.T) {
	withRows(t, parseRows())
	sim := &core.Simulation{}
	character := parseWarrior()
	undead := core.NewAttackTable(&character.Unit, &core.Unit{Type: core.EnemyUnit, MobType: proto.MobType_MobTypeUndead})
	character.AttackTables = []*core.AttackTable{undead}

	row := oneEffectRow(dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS, 32, 99)
	aura := character.RegisterAura(AuraConfig(row))
	ParseEffects(character, aura, row)

	aura.Activate(sim)
	if got := undead.VersusStats[stats.AttackPower]; got != 99 {
		t.Errorf("attack power against Undead while up: %v, want 99", got)
	}
	aura.Deactivate(sim)
	if got := undead.VersusStats[stats.AttackPower]; got != 0 {
		t.Errorf("attack power against Undead after expiry: %v, want 0", got)
	}

}

// The static path runs while the unit is being built, before the attack tables exist, so it puts
// the bonus on them once the environment is finalized.
func TestAStaticVersusRowWaitsForTheAttackTables(t *testing.T) {
	withRows(t, parseRows())
	row := oneEffectRow(dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS, 32, 99)
	undeadTable := func(character *core.Character) *core.AttackTable {
		table := core.NewAttackTable(&character.Unit, &core.Unit{Type: core.EnemyUnit, MobType: proto.MobType_MobTypeUndead})
		character.AttackTables = []*core.AttackTable{table}
		return table
	}

	building := parseWarrior()
	building.Env = &core.Environment{}
	table := undeadTable(building)
	parsed := ParseStatic(building, row)
	if kinds := appliedKinds(parsed); !slices.Equal(kinds, []string{"mob-type-stat AttackPower"}) {
		t.Fatalf("the static path attached %v, want the bonus against Undead", kinds)
	}
	if got := table.VersusStats[stats.AttackPower]; got != 0 {
		t.Errorf("attack power against Undead before the environment is finalized: %v, want 0", got)
	}

	finalized := parseWarrior()
	finalized.Env = &core.Environment{State: core.Finalized}
	table = undeadTable(finalized)
	ParseStatic(finalized, row)
	if got := table.VersusStats[stats.AttackPower]; got != 99 {
		t.Errorf("attack power against Undead once finalized: %v, want 99", got)
	}
}

// Per stat, a bonus against one creature type bids apart from the same stat against another, and
// competes only with the same stat against the same mask.
func TestAVersusAuraBidsPerCreatureTypeMask(t *testing.T) {
	sim := &core.Simulation{}
	character := parseWarrior()
	undead := core.NewAttackTable(&character.Unit, &core.Unit{Type: core.EnemyUnit, MobType: proto.MobType_MobTypeUndead})
	demon := core.NewAttackTable(&character.Unit, &core.Unit{Type: core.EnemyUnit, MobType: proto.MobType_MobTypeDemon})
	character.AttackTables = []*core.AttackTable{undead, demon}

	register := func(label string, mask int32, points float64) *core.Aura {
		row := buffRow(7070, auraEffect(dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS, mask, points))
		aura := character.RegisterAura(AuraConfig(row, Label(label)))
		ParseEffects(character, aura, row, ExclusivePerStat("StatBuff"))
		return aura
	}
	againstUndead := register("Undead", 32, 99)
	againstDemons := register("Demons", 4, 185)
	weakerAgainstUndead := register("Weaker Undead", 32, 36)

	if a, b := categoryNames(againstUndead), categoryNames(againstDemons); slices.Equal(a, b) {
		t.Fatalf("masks 32 and 4 bid in the same category %v", a)
	}
	if a, b := categoryNames(againstUndead), categoryNames(weakerAgainstUndead); !slices.Equal(a, b) {
		t.Fatalf("two bonuses against mask 32 bid in %v and %v, want one category", a, b)
	}

	againstUndead.Activate(sim)
	againstDemons.Activate(sim)
	weakerAgainstUndead.Activate(sim)
	if got := undead.VersusStats[stats.AttackPower]; got != 99 {
		t.Errorf("attack power against Undead: %v, want the stronger 99 alone", got)
	}
	if got := demon.VersusStats[stats.AttackPower]; got != 185 {
		t.Errorf("attack power against Demons: %v, want 185", got)
	}
}

func TestAVersusAuraGainAndExpireDoNotAllocate(t *testing.T) {
	sim := &core.Simulation{}
	character := parseWarrior()
	character.AttackTables = []*core.AttackTable{
		core.NewAttackTable(&character.Unit, &core.Unit{Type: core.EnemyUnit, MobType: proto.MobType_MobTypeUndead}),
		core.NewAttackTable(&character.Unit, &core.Unit{Type: core.EnemyUnit, MobType: proto.MobType_MobTypeDemon}),
	}

	for id, effect := range map[int32]Effect{
		7071: auraEffect(dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS, 32, 99),
		7072: auraEffect(dbcenums.A_MOD_DAMAGE_DONE_VERSUS, 32, 1),
	} {
		row := buffRow(id, effect)
		aura := character.RegisterAura(AuraConfig(row))
		ParseEffects(character, aura, row)
		if allocs := testing.AllocsPerRun(100, func() {
			aura.Activate(sim)
			aura.Deactivate(sim)
		}); allocs != 0 {
			t.Errorf("%v: a gain and an expiry allocate %v times, want 0", effect.Aura, allocs)
		}
	}
}

// A_MOD_DAMAGE_DONE_VERSUS multiplies the damage dealt to the masked mob types while its aura is up,
// and divides it back out on expiry; crit damage follows the damage rather than being multiplied
// again. A row stating 1 against mask 32, Undead.
func TestADamageVersusAuraFollowsItsAuraOnTheAttackTables(t *testing.T) {
	sim := &core.Simulation{}
	character := parseWarrior()
	undead := core.NewAttackTable(&character.Unit, &core.Unit{Type: core.EnemyUnit, MobType: proto.MobType_MobTypeUndead})
	demon := core.NewAttackTable(&character.Unit, &core.Unit{Type: core.EnemyUnit, MobType: proto.MobType_MobTypeDemon})
	character.AttackTables = []*core.AttackTable{undead, demon}
	crit := undead.CritMultiplier

	row := oneEffectRow(dbcenums.A_MOD_DAMAGE_DONE_VERSUS, 32, 1)
	aura := character.RegisterAura(AuraConfig(row))
	ParseEffects(character, aura, row)

	aura.Activate(sim)
	if got := undead.DamageDealtMultiplier; got != 1.01 {
		t.Errorf("damage against Undead while up: %v, want 1.01", got)
	}
	if got := demon.DamageDealtMultiplier; got != 1 {
		t.Errorf("damage against Demons while up: %v, want 1", got)
	}
	if got := undead.CritMultiplier; got != crit {
		t.Errorf("crit multiplier against Undead while up: %v, want %v", got, crit)
	}
	aura.Deactivate(sim)
	if got := undead.DamageDealtMultiplier; got != 1 {
		t.Errorf("damage against Undead after expiry: %v, want 1", got)
	}

	static := parseWarrior()
	static.Env = &core.Environment{State: core.Finalized}
	table := core.NewAttackTable(&static.Unit, &core.Unit{Type: core.EnemyUnit, MobType: proto.MobType_MobTypeUndead})
	static.AttackTables = []*core.AttackTable{table}
	ParseStatic(static, row)
	if got := table.DamageDealtMultiplier; got != 1.01 {
		t.Errorf("damage against Undead from the static path: %v, want 1.01", got)
	}
}

// A_MOD_CRIT_PERCENT_VERSUS multiplies the crit damage dealt to the masked mob types on top of the
// damage, which effect 1 of Humanoid Slaying 19153 raises, while its aura is up.
func TestACritDamageVersusRowMultipliesTheCritOnTheAttackTables(t *testing.T) {
	withGeneratedStore(t)

	static := parseWarrior()
	static.Env = &core.Environment{State: core.Finalized}
	humanoid := core.NewAttackTable(&static.Unit, &core.Unit{Type: core.EnemyUnit, MobType: proto.MobType_MobTypeHumanoid})
	undead := core.NewAttackTable(&static.Unit, &core.Unit{Type: core.EnemyUnit, MobType: proto.MobType_MobTypeUndead})
	static.AttackTables = []*core.AttackTable{humanoid, undead}

	humanoidSlaying := MustFind(19153)
	parsed := ParseStatic(static, humanoidSlaying)
	if kinds := appliedKinds(parsed); !slices.Equal(kinds, []string{"damage-dealt-versus", "crit-damage-versus"}) {
		t.Fatalf("Humanoid Slaying attached %v, want the damage and the crit damage against Humanoids", kinds)
	}
	if got, want := humanoid.DamageDealtMultiplier, 1+humanoidSlaying.EffectN(1).Percent(); math.Abs(got-want) > 1e-9 {
		t.Errorf("damage against Humanoids: %v, want %v", got, want)
	}
	if got, want := humanoid.CritMultiplier, 1+humanoidSlaying.EffectN(2).Percent(); math.Abs(got-want) > 1e-9 {
		t.Errorf("crit multiplier against Humanoids: %v, want %v", got, want)
	}
	if undead.DamageDealtMultiplier != 1 || undead.CritMultiplier != 1 {
		t.Errorf("against Undead: damage %v, crit %v, want both untouched", undead.DamageDealtMultiplier, undead.CritMultiplier)
	}

	sim := &core.Simulation{}
	character := parseWarrior()
	table := core.NewAttackTable(&character.Unit, &core.Unit{Type: core.EnemyUnit, MobType: proto.MobType_MobTypeHumanoid})
	character.AttackTables = []*core.AttackTable{table}
	row := oneEffectRow(dbcenums.A_MOD_CRIT_PERCENT_VERSUS, 64, 3)
	aura := character.RegisterAura(AuraConfig(row))
	ParseEffects(character, aura, row)

	aura.Activate(sim)
	if got := table.CritMultiplier; math.Abs(got-1.03) > 1e-9 {
		t.Errorf("crit multiplier against Humanoids while up: %v, want 1.03", got)
	}
	if got := table.DamageDealtMultiplier; got != 1 {
		t.Errorf("damage against Humanoids while up: %v, want 1", got)
	}
	aura.Deactivate(sim)
	if got := table.CritMultiplier; math.Abs(got-1) > 1e-9 {
		t.Errorf("crit multiplier against Humanoids after expiry: %v, want 1", got)
	}
}

// A_SPELL_AURA_MOD_ARMOR_PENETRATION_PCT is the share of every defender's armor the unit's attacks
// ignore: Serrated Blades 14171 at 3 of 3 ranks ignores the share its rank states, once the attack
// tables exist.
func TestAnArmorPenetrationRowIgnoresAShareOfTheArmor(t *testing.T) {
	withGeneratedStore(t)
	serratedBlades := Talent(14171, 3).Rank(3)
	armoredTable := func(character *core.Character) *core.AttackTable {
		table := core.NewAttackTable(&character.Unit, &core.Unit{Type: core.EnemyUnit})
		character.AttackTables = []*core.AttackTable{table}
		return table
	}

	building := parseWarrior()
	building.Env = &core.Environment{}
	table := armoredTable(building)
	parsed := ParseStatic(building, serratedBlades, Effects(1))
	if kinds := appliedKinds(parsed); !slices.Equal(kinds, []string{"armor-ignore"}) {
		t.Fatalf("Serrated Blades attached %v, want the armor it ignores", kinds)
	}
	if got := table.ArmorIgnoreFactor; got != 0 {
		t.Errorf("armor ignored before the environment is finalized: %v, want 0", got)
	}

	finalized := parseWarrior()
	finalized.Env = &core.Environment{State: core.Finalized}
	table = armoredTable(finalized)
	ParseStatic(finalized, serratedBlades, Effects(1))
	if got, want := table.ArmorIgnoreFactor, serratedBlades.EffectN(1).Percent(); math.Abs(got-want) > 1e-9 {
		t.Errorf("armor ignored once finalized: %v, want %v", got, want)
	}

	sim := &core.Simulation{}
	character := parseWarrior()
	table = armoredTable(character)
	row := oneEffectRow(dbcenums.A_SPELL_AURA_MOD_ARMOR_PENETRATION_PCT, 0, 9)
	aura := character.RegisterAura(AuraConfig(row))
	ParseEffects(character, aura, row)
	aura.Activate(sim)
	if got := table.ArmorIgnoreFactor; math.Abs(got-0.09) > 1e-9 {
		t.Errorf("armor ignored while up: %v, want 0.09", got)
	}
	aura.Deactivate(sim)
	if got := table.ArmorIgnoreFactor; math.Abs(got) > 1e-9 {
		t.Errorf("armor ignored after expiry: %v, want 0", got)
	}
}
