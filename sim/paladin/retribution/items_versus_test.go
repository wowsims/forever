//go:build with_db

package retribution

import (
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/paladin"
)

func paladinAgainstUndead(t *testing.T, equipped map[proto.ItemSlot]int32) (*core.Simulation, *core.Character) {
	t.Helper()
	return paladinAgainstUndeadTargets(t, equipped, 1)
}

func paladinAgainstUndeadTargets(t *testing.T, equipped map[proto.ItemSlot]int32, count int) (*core.Simulation, *core.Character) {
	t.Helper()
	targets := make([]*proto.Target, count)
	for i := range targets {
		targets[i] = &proto.Target{Name: "target", Level: 63, MobType: proto.MobType_MobTypeUndead}
	}
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	for slot, id := range equipped {
		items[slot].Id = id
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Paladin", Class: proto.Class_ClassPaladin, Race: proto.Race_RaceHuman,
				Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
					Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}}}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: &proto.ConsumesSpec{},
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Duration: 180, Targets: targets},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Justice Battlegear's 4-piece bonus 1301084 states A_MOD_MELEE_ATTACK_POWER_VERSUS against mask
// 32, Undead, and no ranged attack power.
func TestJusticeBattlegearFourPieceGrantsAttackPowerAgainstUndead(t *testing.T) {
	_, character := paladinAgainstUndead(t, map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotLegs:     273908,
		proto.ItemSlot_ItemSlotHands:    273909,
		proto.ItemSlot_ItemSlotShoulder: 273910,
		proto.ItemSlot_ItemSlotFeet:     273911,
	})
	table := character.AttackTables[character.CurrentTarget.UnitIndex]

	if got, want := table.MobTypeBonusStats[proto.MobType_MobTypeUndead], (stats.Stats{stats.AttackPower: spelldata.MustFind(1301084).EffectN(1).Average()}); got != want {
		t.Errorf("the bonus against Undead is %v, want %v", got, want)
	}
	if got := table.MobTypeBonusStats[proto.MobType_MobTypeDemon]; got != (stats.Stats{}) {
		t.Errorf("the bonus against Demons is %v, want none", got)
	}
}

// Tenets of the Silver Hand 249397's equip spell 1302545 states A_MOD_DAMAGE_DONE_VERSUS against
// mask 32: its percentage more damage against Undead, with crit damage following the damage.
func TestTenetsOfTheSilverHandMultipliesDamageAgainstUndead(t *testing.T) {
	_, without := paladinAgainstUndead(t, nil)
	_, with := paladinAgainstUndead(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotRanged: 249397})
	base := without.AttackTables[without.CurrentTarget.UnitIndex]
	table := with.AttackTables[with.CurrentTarget.UnitIndex]

	if got, want := table.DamageDealtMultiplier, base.DamageDealtMultiplier*(1+spelldata.MustFind(1302545).EffectN(1).Percent()); got != want {
		t.Errorf("the damage multiplier against Undead is %v, want %v", got, want)
	}
	if got, want := table.CritMultiplier, base.CritMultiplier; got != want {
		t.Errorf("the crit multiplier against Undead is %v, want %v", got, want)
	}
}

// The PvP gloves' equip spell 23300 states a flat amount on Judgement of the Crusader's Holy damage
// bonus, which every rank of the paladin's debuff carries.
func TestJudgementOfTheCrusaderGlovesRaiseHolyDamageTaken(t *testing.T) {
	holyTaken := func(hands int32) float64 {
		sim, character := paladinAgainstUndead(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotHands: hands})
		target := character.CurrentTarget
		for _, aura := range target.GetAurasWithTag(buffs.JudgementAuraTag) {
			if strings.HasPrefix(aura.Label, "Judgement of the Crusader") {
				aura.Activate(sim)
			}
		}
		return target.PseudoStats.SchoolBonusSpellDamage[stats.SchoolIndexHoly]
	}

	base := holyTaken(0)
	if base == 0 {
		t.Fatal("no Judgement of the Crusader debuff raised Holy damage taken")
	}
	increase := spelldata.MustFind(23300).EffectN(1).Average()
	for _, gloves := range []int32{16410, 16471, 23274} {
		if got, want := holyTaken(gloves), base+increase; got != want {
			t.Errorf("with gloves %d the debuff adds %v Holy damage taken, want %v", gloves, got, want)
		}
	}
}

// The Premier PvP gloves' equip spell 1293544 states a flat crit chance on family 10 mask 0x2000,
// every rank of Flash of Light and no other spell.
func TestPremierLuminousGlovesRaiseFlashOfLightCrit(t *testing.T) {
	bonusCrit := func(hands int32) (flashOfLight []float64, others []float64) {
		_, character := paladinAgainstUndead(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotHands: hands})
		for _, spell := range character.Spellbook {
			if spell.Matches(paladin.SpellMaskFlashOfLight) {
				flashOfLight = append(flashOfLight, spell.BonusCritPercent)
			} else {
				others = append(others, spell.BonusCritPercent)
			}
		}
		return flashOfLight, others
	}

	baseFlash, baseOthers := bonusCrit(0)
	if len(baseFlash) == 0 {
		t.Fatal("the paladin registers no Flash of Light")
	}
	crit := spelldata.MustFind(1293544).EffectN(1).Average()
	for _, gloves := range []int32{273329, 273363, 274250, 274256} {
		flash, others := bonusCrit(gloves)
		for i := range baseFlash {
			if got, want := flash[i], baseFlash[i]+crit; got != want {
				t.Errorf("with gloves %d Flash of Light rank %d has %v bonus crit, want %v", gloves, i+1, got, want)
			}
		}
		for i := range baseOthers {
			if others[i] != baseOthers[i] {
				t.Errorf("with gloves %d another spell's bonus crit moved from %v to %v", gloves, baseOthers[i], others[i])
			}
		}
	}
}

// Seal of Fury's hit and its judgement take the paladin's spell power but none of the spell damage
// against the target's creature type (EV-0143).
func TestSealOfFuryTakesNoSpellDamageAgainstTheTargetsType(t *testing.T) {
	for name, mask := range sealOfFuryHits {
		t.Run(name, func(t *testing.T) {
			damage := func(general, versus float64) float64 {
				damage, _ := twentyCasts(t, spellWithMask(mask), func(sim *core.Simulation, character *core.Character, _ *core.Spell) {
					character.AddStatDynamic(sim, stats.SpellDamage, general)
					character.AddMobTypeBonusStats(core.ActionID{SpellID: 1}, stats.Stats{stats.SpellDamage: versus}, proto.MobType_MobTypeUndead)
				})
				return damage
			}

			base := damage(0, 0)
			if got := damage(100, 0); got <= base {
				t.Errorf("under 100 spell power the casts deal %v, want more than the %v they deal without", got, base)
			}
			if got := damage(0, 100); got != base {
				t.Errorf("under 100 spell power against Undead the casts deal %v, want the %v they deal without", got, base)
			}
		})
	}
}
