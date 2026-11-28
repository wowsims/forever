package retribution

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/paladin"
	goproto "google.golang.org/protobuf/proto"
)

var noConvictionTalents = "050250002-5-05203301101301351"      // The Unsealed Wrath build without its Conviction points.
var noTwoHandedSpecTalents = "050250002-501-05233011201001351" // The Unsealed Wrath build without its Two-Handed Weapon Specialization points.

func convictionPaladin(t *testing.T, talents string) *paladin.Paladin {
	t.Helper()
	return paladinWearing(t, talents, core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "default").GearSet)
}

func paladinWearing(t *testing.T, talents string, equipment *proto.EquipmentSpec) *paladin.Paladin {
	t.Helper()
	player := &proto.Player{
		Name:          "Conviction",
		Race:          proto.Race_RaceUndead,
		Class:         proto.Class_ClassPaladin,
		Equipment:     equipment,
		TalentsString: talents,
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
		}},
		Rotation: core.GetAplRotation("../../../ui/specs/paladin/retribution/apls", "default").Rotation,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
}

// Conviction 3/3 gives 6% crit to every spell and attack while the two-handed axe is equipped.
func TestConvictionGivesEverySpellAndAttackCrit(t *testing.T) {
	pal, base := convictionPaladin(t, UnsealedWrathTalents), convictionPaladin(t, noConvictionTalents)
	if pal.Talents.Conviction != 3 || base.Talents.Conviction != 0 {
		t.Fatalf("the talents take Conviction %d/3 and %d/3, want 3/3 and none", pal.Talents.Conviction, base.Talents.Conviction)
	}

	for _, spell := range []struct {
		name string
		id   core.ActionID
	}{
		{"the main-hand auto attack", pal.AutoAttacks.MHAuto().ActionID},
		{"Holy Strike", core.ActionID{SpellID: paladin.HolyStrikeRankMap.Highest().ID}},
		{"Hammer of Wrath", core.ActionID{SpellID: paladin.HammerOfWrathRankMap.Highest().ID}},
	} {
		with, without := pal.GetSpell(spell.id), base.GetSpell(spell.id)
		if with == nil || without == nil {
			t.Fatalf("%s is not registered", spell.name)
		}
		got := with.PhysicalCritChance(pal.AttackTables[pal.CurrentTarget.UnitIndex]) -
			without.PhysicalCritChance(base.AttackTables[base.CurrentTarget.UnitIndex])
		if with.DefenseType == core.DefenseTypeMagic {
			got = with.SpellCritChance(pal.CurrentTarget) - without.SpellCritChance(base.CurrentTarget)
		}
		if math.Abs(got-0.06) > 1e-9 {
			t.Errorf("Conviction gives %s %.4f crit, want 0.06", spell.name, got)
		}
	}
}

// Two-Handed Weapon Specialization gives every Holy damage spell the bonus the physical attacks take
// while a two-hander is equipped, whatever its defense type.
func TestTwoHandedWeaponSpecializationHolyReach(t *testing.T) {
	pal, base := convictionPaladin(t, UnsealedWrathTalents), convictionPaladin(t, noTwoHandedSpecTalents)
	if pal.Talents.TwoHandedWeaponSpecialization == 0 || base.Talents.TwoHandedWeaponSpecialization != 0 ||
		pal.MainHand().HandType != proto.HandType_HandTypeTwoHand {
		t.Fatal("want a two-hander, on one paladin with Two-Handed Weapon Specialization and one without")
	}
	bonus := pal.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] /
		base.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]
	if bonus <= 1 {
		t.Fatalf("the talent multiplies physical damage by %v, want its bonus", bonus)
	}

	equipment := goproto.Clone(core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "default").GearSet).(*proto.EquipmentSpec)
	equipment.Items[proto.ItemSlot_ItemSlotMainHand] =
		core.GetGearSet("../../../ui/specs/paladin/protection/gear_sets", "default").GearSet.Items[proto.ItemSlot_ItemSlotMainHand]
	oneHanded, oneHandedBase := paladinWearing(t, UnsealedWrathTalents, equipment), paladinWearing(t, noTwoHandedSpecTalents, equipment)
	if oneHanded.MainHand().HandType == proto.HandType_HandTypeTwoHand {
		t.Fatal("the Protection set's main hand is a two-hander, want a one-hander")
	}

	damage := func(p *paladin.Paladin, spell *core.Spell) float64 {
		return spell.AttackerDamageMultiplier(p.AttackTables[p.CurrentTarget.UnitIndex], false)
	}
	healing := func(_ *paladin.Paladin, spell *core.Spell) float64 { return spell.CasterHealingMultiplier() }
	for _, tc := range []struct {
		name  string
		mask  int64
		id    core.ActionID
		heals bool
	}{
		{name: "Holy Strike", mask: paladin.SpellMaskHolyStrike},
		{name: "the Seal of Righteousness hit", mask: paladin.SpellMaskSealOfRighteousnessProc},
		{name: "Judgement of Command", mask: paladin.SpellMaskJudgementOfCommand},
		{name: "Judgement of Righteousness", mask: paladin.SpellMaskJudgementOfRighteousness},
		{name: "Hammer of Wrath", mask: paladin.SpellMaskHammerOfWrath},
		{name: "Consecration", mask: paladin.SpellMaskConsecration},
		{name: "Exorcism", mask: paladin.SpellMaskExorcism},
		{name: "Holy Wrath", mask: paladin.SpellMaskHolyWrath},
		{name: "Retribution Aura", id: core.ActionID{SpellID: paladin.RetributionAuraRankMap.Highest().ID, Tag: 2}},
		{name: "Holy Light", mask: paladin.SpellMaskHolyLight, heals: true},
	} {
		multiplier, want := damage, bonus
		if tc.heals {
			multiplier, want = healing, 1
		}
		for _, wearing := range []struct {
			hand       string
			with, base *paladin.Paladin
			want       float64
		}{
			{"a two-hander", pal, base, want},
			{"a one-hander", oneHanded, oneHandedBase, 1},
		} {
			checked := 0
			for _, spell := range wearing.with.Spellbook {
				if !(spell.Matches(tc.mask) || tc.mask == 0 && spell.ActionID.SameAction(tc.id)) {
					continue
				}
				without := wearing.base.GetSpell(spell.ActionID)
				if without == nil {
					t.Fatalf("%s: %v is not registered without the talent", tc.name, spell.ActionID)
				}
				checked++
				if got := multiplier(wearing.with, spell) / multiplier(wearing.base, without); math.Abs(got-wearing.want) > 1e-9 {
					t.Errorf("%s %v with %s: the talent multiplies it by %v, want %v", tc.name, spell.ActionID, wearing.hand, got, wearing.want)
				}
			}
			if checked == 0 {
				t.Errorf("%s with %s: no spell is registered", tc.name, wearing.hand)
			}
		}
	}
}
