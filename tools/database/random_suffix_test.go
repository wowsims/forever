package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

// A bonus list's allocations, per 10000 of the item's random property points, land on the stats their
// item mod names: spell power on spell damage and healing alike, healing and spell damage done each on
// its own, a mod named twice adds up, and health regeneration lands on HP5.
func TestRandomSuffixAllocations(t *testing.T) {
	type stats = map[proto.Stat]float64
	row := func(itemMod, allocation int) dbc.RandomSuffixStat {
		return dbc.RandomSuffixStat{ItemMod: itemMod, Allocation: allocation}
	}
	for _, tc := range []struct {
		name  string
		rows  []dbc.RandomSuffixStat
		wants stats
	}{
		{"of the Sorcerer: Intellect, Stamina and spell power",
			[]dbc.RandomSuffixStat{row(dbc.ITEM_MOD_INTELLECT, 4000), row(dbc.ITEM_MOD_STAMINA, 2600), row(dbc.ITEM_MOD_SPELL_POWER, 2300)},
			stats{proto.Stat_StatIntellect: 4000, proto.Stat_StatStamina: 2600, proto.Stat_StatSpellDamage: 2300, proto.Stat_StatHealingPower: 2300}},
		{"of Healing: healing and spell damage done apart",
			[]dbc.RandomSuffixStat{row(dbc.ITEM_MOD_SPELL_HEALING_DONE, 3760), row(dbc.ITEM_MOD_SPELL_DAMAGE_DONE, 1253)},
			stats{proto.Stat_StatHealingPower: 3760, proto.Stat_StatSpellDamage: 1253}},
		{"a mod named twice",
			[]dbc.RandomSuffixStat{row(dbc.ITEM_MOD_STAMINA, 500), row(dbc.ITEM_MOD_STAMINA, 700)},
			stats{proto.Stat_StatStamina: 1200}},
		{"of Regeneration: health regeneration alone",
			[]dbc.RandomSuffixStat{row(dbc.ITEM_MOD_HEALTH_REGEN, 1000)},
			stats{proto.Stat_StatHP5: 1000}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			suffix := dbc.RandomSuffix{ID: 990001, Name: tc.name, Stats: tc.rows}.ToProto()
			for stat, got := range suffix.Stats {
				if want := tc.wants[proto.Stat(stat)]; got != want {
					t.Errorf("%v: allocation %v, want %v", proto.Stat(stat), got, want)
				}
			}
		})
	}
}

// "of Eluding" carries 13669 as an equip spell, whose dodge percent the suffix adds whatever the item's
// points.
func TestRandomSuffixEquipSpellDodge(t *testing.T) {
	withDBCInputs(t)
	dbc.GetDBC()

	suffix := dbc.RandomSuffix{ID: 990002, Name: "of Eluding", EquipSpells: []int{13669}}.ToProto()
	want := equipAuraPoints(t, 13669, dbcenums.A_MOD_DODGE_PERCENT)
	got := 0.0
	if int(proto.PseudoStat_PseudoStatDodgePercent) < len(suffix.PseudoStats) {
		got = suffix.PseudoStats[proto.PseudoStat_PseudoStatDodgePercent]
	}
	if got != want || want == 0 {
		t.Errorf("dodge: got %v, want 13669's %v", got, want)
	}
	for stat, value := range suffix.Stats {
		if value != 0 {
			t.Errorf("%v: %v, want no flat stats", proto.Stat(stat), value)
		}
	}
}

// A bonus list with an ITEM_BONUS_REDUCED_CASTER_DPS row marks its suffix, and a weapon offering one
// carries the damage it deals with it: Brutal War Axe (3210) shows 43-66 with "of the Physician"
// (12911) in game against 58-88 on its own. A weapon offering none, and the armor the lower lists
// sit on, carry no such range.
func TestReducedCasterDpsSuffixDamage(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for id, want := range map[int]bool{12911: true, 12748: false} {
		if got := instance.RandomSuffix[id].ReducedCasterDps; got != want {
			t.Errorf("suffix %d reduces caster DPS: %v, want %v", id, got, want)
		}
	}

	axe := instance.Items[3210]
	scaling := axe.ToScaledUIItem(axe.ItemLevel).ScalingOptions[0]
	if scaling.WeaponDamageMin != 58 || scaling.WeaponDamageMax != 88 {
		t.Errorf("Brutal War Axe: damage %v-%v, want 58-88", scaling.WeaponDamageMin, scaling.WeaponDamageMax)
	}
	if reduced := scaling.ReducedCasterWeaponDamage; reduced == nil {
		t.Error("Brutal War Axe: no reduced caster damage")
	} else if reduced.WeaponDamageMin != 43 || reduced.WeaponDamageMax != 66 || reduced.UnroundedDps >= scaling.UnroundedDps {
		t.Errorf("Brutal War Axe: reduced caster damage %v-%v at %v DPS, want 43-66 under %v", reduced.WeaponDamageMin, reduced.WeaponDamageMax, reduced.UnroundedDps, scaling.UnroundedDps)
	}

	for _, id := range []int{281721, 12017} { // Restored Cleaver, Prismatic Band
		item := instance.Items[id]
		if reduced := item.ToScaledUIItem(item.ItemLevel).ScalingOptions[0].ReducedCasterWeaponDamage; reduced != nil {
			t.Errorf("%s: reduced caster damage %v, want none", item.Name, reduced)
		}
	}
}
