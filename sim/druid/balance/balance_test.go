package balance

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/siml"
)

func init() {
	RegisterBalanceDruid()
	common.RegisterAllEffects()
}

// The Balance page's default build, all from defaults.siml: race, professions, spec options, talents, gear,
// rotation, consumables, distance, timings, buffs, debuffs, target dummies and encounter.
var defaultBuild = readDefaultBuild()

// The Moonkin talents defaults.siml sets.
var DefaultTalents = defaultBuild.Talents

func readDefaultBuild() core.CharacterSuiteConfig {
	f, err := siml.ReadFile("../../../ui/specs/druid/balance/defaults.siml")
	if err != nil {
		panic(err)
	}
	settings, player := f.Settings, f.Settings.GetPlayer()
	return core.CharacterSuiteConfig{
		Class:              proto.Class_ClassDruid,
		Race:               player.GetRace(),
		Profession1:        player.GetProfession1(),
		Profession2:        player.GetProfession2(),
		GearSet:            core.GearSetCombo{Label: "default", GearSet: f.Equipment()},
		SpecOptions:        core.SpecOptionsCombo{Label: "Standard", SpecOptions: player.GetSpec()},
		Talents:            player.GetTalentsString(),
		Rotation:           core.RotationCombo{Label: "default", Rotation: f.Rotation()},
		Encounter:          core.EncounterCombo{Label: "default", Encounter: settings.GetEncounter()},
		StartingDistance:   player.GetDistanceFromTarget(),
		ReactionTimeMs:     player.GetReactionTimeMs(),
		ChannelClipDelayMs: player.GetChannelClipDelayMs(),
		Consumables:        player.GetConsumables(),
		IndividualBuffs:    player.GetBuffs(),
		PartyBuffs:         settings.GetPartyBuffs(),
		RaidBuffs:          settings.GetRaidBuffs(),
		Debuffs:            settings.GetDebuffs(),
		Cooldowns:          player.GetCooldowns(),
		TargetDummies:      settings.GetTargetDummies(),
	}
}

func TestBalance(t *testing.T) {
	config := defaultBuild
	config.OtherRaces = []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceTauren}
	config.ItemFilter = core.ItemFilter{
		WeaponTypes:       DefaultWeaponTypes,
		ArmorType:         DefaultArmorType,
		RangedWeaponTypes: DefaultRangedWeaponTypes,
	}
	config.EPReferenceStat = proto.Stat_StatSpellDamage
	config.StatsToWeigh = []proto.Stat{
		proto.Stat_StatIntellect,
		proto.Stat_StatSpellDamage,
		proto.Stat_StatArcaneDamage,
		proto.Stat_StatNatureDamage,
		proto.Stat_StatSpellPiercing,
		proto.Stat_StatHitRating,
		proto.Stat_StatCritRating,
		proto.Stat_StatHasteRating,
		proto.Stat_StatMP5,
		proto.Stat_StatSpirit,
	}
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{config}))
}

// A Moonkin build without Dreamstate, and the same build with one point and with all five in it.
const (
	noDreamstateTalents   = "5232220115201301--5053"
	dreamstateOneTalents  = "5232220115201311--5053"
	fullDreamstateTalents = "5232220115201351--5053"
)

const DefaultArmorType = proto.ArmorType_ArmorTypeLeather

var DefaultWeaponTypes = []proto.WeaponType{
	proto.WeaponType_WeaponTypeDagger,
	proto.WeaponType_WeaponTypeMace,
	proto.WeaponType_WeaponTypeStaff,
	proto.WeaponType_WeaponTypeOffHand,
}

var DefaultRangedWeaponTypes = []proto.RangedWeaponType{
	proto.RangedWeaponType_RangedWeaponTypeIdol,
}

// Someone else still spends the cast and the cooldown; without Dreamstate the caster gains none of the mana.
func TestGiftedInnervateCastsWithoutTheMana(t *testing.T) {
	for _, c := range []struct {
		name     string
		gift     bool
		wantMana bool
	}{
		{"self", false, true},
		{"someone else", true, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, druid := balanceDruid(t, noDreamstateTalents, c.gift, 0, 1)
			spell := druid.GetSpell(core.ActionID{SpellID: 29166})
			if spell == nil {
				t.Fatal("Innervate is not registered")
			}

			before := druid.ManaRegenPerSecondWhileNotCasting()
			dummy := sim.Raid.GetFirstTargetDummy()
			if dummy == nil {
				t.Fatal("Balance has no dummy")
			}

			target := sim.Encounter.ActiveTargetUnits[0]
			if !spell.Cast(sim, target) {
				t.Fatal("Innervate did not cast")
			}
			if spell.IsReady(sim) {
				t.Error("Innervate is still ready")
			}
			if got := spell.SpellMetrics[target.UnitIndex].Casts; got != 1 {
				t.Errorf("Innervate casts = %d, want 1", got)
			}

			active := druid.HasActiveAuraWithTag(buffs.InnervatesCategory)
			gotMana := druid.ManaRegenPerSecondWhileNotCasting() != before
			if active != c.wantMana || gotMana != c.wantMana {
				t.Errorf("mana buff active = %v, regen changed = %v, want %v", active, gotMana, c.wantMana)
			}
			if gotDummy := dummy.HasActiveAuraWithTag(buffs.InnervatesCategory); gotDummy == c.wantMana {
				t.Errorf("dummy Innervate = %v, want %v", gotDummy, !c.wantMana)
			}
		})
	}

	t.Run("someone else with an incoming innervate", func(t *testing.T) {
		sim, druid := balanceDruid(t, noDreamstateTalents, true, 1, 1)
		spell := druid.GetSpell(core.ActionID{SpellID: 29166})
		if spell == nil {
			t.Fatal("Innervate is not registered")
		}
		incoming := druid.GetAura("Innervates (External)")
		if incoming == nil {
			t.Fatal("the incoming Innervate is not registered")
		}
		incoming.Activate(sim)
		if !druid.HasActiveAuraWithTag(buffs.InnervatesCategory) {
			t.Fatal("the incoming Innervate is not up")
		}
		target := sim.Encounter.ActiveTargetUnits[0]
		if !spell.Cast(sim, target) {
			t.Fatal("Innervate did not cast")
		}
		if spell.IsReady(sim) {
			t.Error("Innervate is still ready")
		}
		if got := spell.SpellMetrics[target.UnitIndex].Casts; got != 1 {
			t.Errorf("Innervate casts = %d, want 1", got)
		}
		dummy := sim.Raid.GetFirstTargetDummy()
		if dummy == nil {
			t.Fatal("Balance has no dummy")
		}
		if !dummy.HasActiveAuraWithTag(buffs.InnervatesCategory) {
			t.Error("the dummy has no Innervate")
		}
	})

	t.Run("cast at the dummy", func(t *testing.T) {
		sim, druid := balanceDruid(t, noDreamstateTalents, false, 0, 1)
		spell := druid.GetSpell(core.ActionID{SpellID: 29166})
		if spell == nil {
			t.Fatal("Innervate is not registered")
		}
		dummy := sim.Raid.GetFirstTargetDummy()
		if dummy == nil {
			t.Fatal("Balance has no dummy")
		}
		before := druid.ManaRegenPerSecondWhileNotCasting()
		if !spell.Cast(sim, &dummy.Unit) {
			t.Fatal("Innervate did not cast")
		}
		if druid.HasActiveAuraWithTag(buffs.InnervatesCategory) {
			t.Error("the caster has Innervate")
		}
		if druid.ManaRegenPerSecondWhileNotCasting() != before {
			t.Error("the caster's regen changed")
		}
		if !dummy.HasActiveAuraWithTag(buffs.InnervatesCategory) {
			t.Error("the dummy has no Innervate")
		}
	})
}

// Dreamstate gives the caster a share of an Innervate cast on someone else: all of it at full rank, less at
// one point, and nothing extra on a self-cast.
func TestDreamstateSharesAGiftedInnervate(t *testing.T) {
	gain := func(t *testing.T, talents string, gift bool) float64 {
		sim, druid := balanceDruid(t, talents, gift, 0, 1)
		spell := druid.GetSpell(core.ActionID{SpellID: 29166})
		if spell == nil {
			t.Fatal("Innervate is not registered")
		}
		before := druid.ManaRegenPerSecondWhileNotCasting()
		if !spell.Cast(sim, sim.Encounter.ActiveTargetUnits[0]) {
			t.Fatal("Innervate did not cast")
		}
		if gift && !sim.Raid.GetFirstTargetDummy().HasActiveAuraWithTag(buffs.InnervatesCategory) {
			t.Error("the dummy has no Innervate")
		}
		return druid.ManaRegenPerSecondWhileNotCasting() - before
	}

	self := gain(t, noDreamstateTalents, false)
	if self <= 0 {
		t.Fatalf("a self-cast Innervate gains %v regen, want more than 0", self)
	}
	if got := gain(t, fullDreamstateTalents, true); math.Abs(got-self) > 1e-9 {
		t.Errorf("full Dreamstate gains %v regen from a gift, want the self-cast's %v", got, self)
	}
	if got := gain(t, dreamstateOneTalents, true); got <= 0 || got >= self {
		t.Errorf("one point of Dreamstate gains %v regen from a gift, want between 0 and %v", got, self)
	}
	if got := gain(t, fullDreamstateTalents, false); math.Abs(got-self) > 1e-9 {
		t.Errorf("a self-cast with Dreamstate gains %v regen, want the self-cast's %v", got, self)
	}
}

func balanceDruid(t *testing.T, talents string, gift bool, incoming int32, targets int) (*core.Simulation, *core.Character) {
	t.Helper()
	target := &proto.UnitReference{}
	if gift {
		target = &proto.UnitReference{Type: proto.UnitReference_Player, Index: 1}
	}
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassDruid,
		Race:          proto.Race_RaceNightElf,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
		Buffs:         &proto.IndividualBuffs{Innervates: incoming},
	}, &proto.Player_BalanceDruid{
		BalanceDruid: &proto.BalanceDruid{
			Options: &proto.BalanceDruid_Options{
				ClassOptions: &proto.DruidOptions{InnervateTarget: target},
			},
		},
	})
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.TargetDummies = 1
	encounter := core.MakeSingleTargetEncounter(0)
	for len(encounter.Targets) < targets {
		encounter.Targets = append(encounter.Targets, core.NewDefaultTarget())
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1, UseLabeledRands: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// A Balance druid in a fresh sim with this many targets; a duration other than 0 resets the sim to run that long.
func newBalanceDruid(t *testing.T, targets int, duration time.Duration) (*core.Simulation, *BalanceDruid) {
	t.Helper()
	sim, _ := balanceDruid(t, DefaultTalents, false, 0, targets)
	if duration != 0 {
		sim.Cleanup()
		sim.BaseDuration = duration
		sim.Reset()
	}
	return sim, sim.Raid.Parties[0].Players[0].(*BalanceDruid)
}

// Steps the sim until the clock reaches until.
func stepUntil(sim *core.Simulation, until time.Duration) {
	for sim.CurrentTime < until && !sim.Step() {
	}
}

// How many of casts casts end with Clearcasting active 100 ms later. The buff makes each cast free;
// settle waits out the proc's cooldown.
func countOmenProcs(sim *core.Simulation, druid *BalanceDruid, casts int, cast, settle func(start time.Duration)) int {
	procs := 0
	for range casts {
		druid.ClearcastingAura.Activate(sim)
		start := sim.CurrentTime
		cast(start)
		stepUntil(sim, start+100*time.Millisecond)
		if druid.ClearcastingAura.IsActive() {
			procs++
		}
		settle(start)
	}
	return procs
}

// Hurricane's ticks roll no Omen of Clarity, and their crits grant no Nature's Grace.
func TestHurricaneTicksRollNoOmenAndGrantNoNaturesGrace(t *testing.T) {
	t.Run("Omen of Clarity", func(t *testing.T) {
		sim, druid := newBalanceDruid(t, 3, 0)
		for range 500 {
			// The step clears the proc's cooldown, which a successful roll starts on the spot.
			sim.CurrentTime += time.Minute
			druid.Hurricane.RelatedDotSpell.Cast(sim, druid.CurrentTarget)
			if !druid.OmenOfClarityAura.Icd.IsReady(sim) {
				t.Fatal("a Hurricane tick rolled Omen of Clarity")
			}
		}
	})

	t.Run("Nature's Grace", func(t *testing.T) {
		sim, druid := newBalanceDruid(t, 1, 0)
		naturesGrace := druid.GetAura("Nature's Grace")
		if naturesGrace == nil {
			t.Fatal("Nature's Grace is not registered")
		}
		druid.AddStatDynamic(sim, stats.SpellCritPercent, 100)
		tick := druid.Hurricane.RelatedDotSpell
		start := sim.CurrentTime
		tick.Cast(sim, druid.CurrentTarget)
		stepUntil(sim, start+time.Second)
		if tick.SpellMetrics[druid.CurrentTarget.UnitIndex].Crits == 0 {
			t.Fatal("the tick did not crit")
		}
		if naturesGrace.IsActive() {
			t.Error("a Hurricane tick crit granted Nature's Grace")
		}
	})
}

// A Hurricane tick is cast by the channel, so it is a proc, heard only by listeners that can proc from
// procs, unless its row is flagged Not a Proc.
func TestHurricaneTickIsAProcAsItsRowStates(t *testing.T) {
	_, druid := newBalanceDruid(t, 1, 0)
	tick := druid.Hurricane.RelatedDotSpell
	row := spelldata.MustFind(tick.ActionID.SpellID)
	if got := tick.Flags.Matches(core.SpellFlagProc); got != row.IsAProc() {
		t.Errorf("the Hurricane tick is a proc: %v, want %v as its row states", got, row.IsAProc())
	}
}

// The Hurricane cast rolls Omen of Clarity once for each target its area reaches, after it spends
// Clearcasting, so a cast that spends the buff can come out with a new one.
func TestHurricaneCastRollsOmenPerTarget(t *testing.T) {
	const casts = 400
	procsOn := func(targets int) int {
		sim, druid := newBalanceDruid(t, targets, casts*time.Minute)
		return countOmenProcs(sim, druid, casts, func(_ time.Duration) {
			if !druid.Hurricane.Cast(sim, druid.CurrentTarget) {
				t.Fatal("Hurricane did not cast")
			}
		}, func(_ time.Duration) {
			// The channel outlasts the proc's cooldown.
			for druid.Hurricane.AOEDot().IsActive() && !sim.Step() {
			}
		})
	}

	one, five := procsOn(1), procsOn(5)
	if one == 0 {
		t.Fatalf("a Hurricane cast that spends Clearcasting on one target never comes out with a new one in %d casts", casts)
	}
	if float64(five) < 2.5*float64(one) {
		t.Errorf("Hurricane procs Omen of Clarity %d of %d casts on five targets, %d on one; want about four times as often", five, casts, one)
	}
}

// A Hurricane cast rolls Draconic Infused Emblem 22268 once per enemy, its ticks none (EV-0115).
func TestHurricaneCastRollsDraconicInfusedEmblemPerEnemy(t *testing.T) {
	const casts, enemies = 2000, 5
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotTrinket1+1)
	for slot := range items {
		items[slot] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotTrinket1].Id = 22268
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassDruid,
		Race:          proto.Race_RaceNightElf,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: DefaultTalents,
		Rotation:      &proto.APLRotation{},
	}, &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{Options: &proto.BalanceDruid_Options{ClassOptions: &proto.DruidOptions{}}}})
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Duration = casts * 11
	for range enemies - 1 {
		encounter.Targets = append(encounter.Targets, core.NewDefaultTarget())
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	druid := sim.Raid.Parties[0].Players[0].(*BalanceDruid)
	buff := druid.GetAuraByID(core.ActionID{SpellID: 1318930})
	if buff == nil {
		t.Fatal("22268 registers no 1318930")
	}

	// The proc is deferred by the spell batch window, one pending action each, so each step deals at most
	// one, and the buff comes down before the next roll can refresh it.
	var atCast, onTicks int
	for range casts {
		start := sim.CurrentTime
		druid.Hurricane.SkipCastAndApplyEffects(sim, druid.CurrentTarget)
		for druid.Hurricane.AOEDot().IsActive() || sim.CurrentTime < start+100*time.Millisecond {
			if sim.Step() {
				t.Fatal("the encounter ended before the casts did")
			}
			if buff.IsActive() {
				if sim.CurrentTime < start+100*time.Millisecond {
					atCast++
				} else {
					onTicks++
				}
				buff.Deactivate(sim)
			}
		}
	}

	chance := spelldata.MustFind(1318931).StatedChance()
	if chance <= 0 || chance >= 1 {
		t.Fatalf("1318931 states a chance of %v, want one between none and certain", chance)
	}
	rolls := float64(casts * enemies)
	if mean, deviation := chance*rolls, math.Sqrt(chance*(1-chance)*rolls); math.Abs(float64(atCast)-mean) > 4*deviation {
		t.Errorf("%d Hurricane casts on %d enemies procced it %d times at the cast, want about %.0f", casts, enemies, atCast, mean)
	}
	if onTicks != 0 {
		t.Errorf("Hurricane ticks procced it %d times, want none", onTicks)
	}
}

// A helpful cast such as Thorns rolls Omen of Clarity in Moonkin Form, after it spends Clearcasting;
// in caster form it rolls nothing, and neither does a shift into Moonkin Form.
func TestHelpfulCastsRollOmenInMoonkinFormOnly(t *testing.T) {
	const casts = 400
	newDruid := func() (*core.Simulation, *BalanceDruid) {
		return newBalanceDruid(t, 1, casts*10*time.Second)
	}
	procsIn := func(moonkin bool) int {
		sim, druid := newDruid()
		if !moonkin {
			druid.ClearForm(sim)
		}
		return countOmenProcs(sim, druid, casts, func(_ time.Duration) {
			if !druid.Thorns.Cast(sim, &druid.Unit) {
				t.Fatal("Thorns did not cast")
			}
		}, func(start time.Duration) {
			// The wait outlasts the proc's cooldown.
			stepUntil(sim, start+10*time.Second)
		})
	}

	if got := procsIn(true); got == 0 {
		t.Errorf("Thorns never procs Omen of Clarity in Moonkin Form in %d casts", casts)
	}
	if got := procsIn(false); got != 0 {
		t.Errorf("Thorns procs Omen of Clarity %d of %d casts in caster form, want none", got, casts)
	}

	sim, druid := newDruid()
	// A free shift, so the mana lasts; a roll would start the proc's cooldown on the spot.
	druid.MoonkinForm.Cost.PercentModifier = 0
	for range casts {
		druid.ClearForm(sim)
		start := sim.CurrentTime
		if !druid.MoonkinForm.Cast(sim, &druid.Unit) {
			t.Fatal("Moonkin Form did not cast")
		}
		if !druid.OmenOfClarityAura.Icd.IsReady(sim) {
			t.Fatal("a shift into Moonkin Form rolled Omen of Clarity")
		}
		stepUntil(sim, start+10*time.Second)
	}
}
