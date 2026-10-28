package priest

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	RegisterPriest()
	core.RegisterAgentFactory(
		proto.Player_HealerPriest{},
		proto.Spec_SpecHealerPriest,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
			return &healingPriest{Priest: New(character, SelfBuffs{}, options.TalentsString)}
		},
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_HealerPriest)
		},
	)
	common.RegisterAllEffects()
}

// A priest with the healing spells, for the healer rules: the healer spec's package imports this one,
// so the tests here cannot build its agent.
type healingPriest struct {
	*Priest
}

func (healer *healingPriest) Initialize() {
	healer.Priest.Initialize()
	healer.Priest.InitializeHealingSpells()
}

// The Disc preset from ui/specs/priest/healer/presets.ts.
const discTalents = "005005031330011531-335020030301"

func newDiscSim(t *testing.T) (*core.Simulation, *Priest) {
	t.Helper()
	player := &proto.Player{
		Name:          "Disc",
		Race:          proto.Race_RaceTroll,
		Class:         proto.Class_ClassPriest,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: discTalents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
		Spec: &proto.Player_HealerPriest{
			HealerPriest: &proto.HealerPriest{
				Options: &proto.HealerPriest_Options{ClassOptions: &proto.PriestOptions{}},
			},
		},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(10),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim, sim.Raid.Parties[0].Players[0].(*healingPriest).Priest
}

// The Shadow Testing build's gear, talents and rotation, which the UI's Shadow Testing gear set,
// Standard talents and Default rotation read too.
var shadowTestingBuild = core.GetTestBuildFromJSON(proto.Class_ClassPriest, "../../ui/specs/priest/dps/builds", "testing",
	core.ItemFilter{}, nil, nil)

// The UI's defaults: Flask of Supreme Power, Major Spellblasting Potion, Nightfin Soup, Demonic Rune,
// Brilliant Wizard Oil and Ez-Thro Dynamite II.
var defaultConsumables = &proto.ConsumesSpec{
	FlaskId:        13512,
	OffensivePotId: 250937,
	FoodId:         13931,
	ConjuredId:     12662,
	MhImbueId:      20749,
	ExplosiveId:    18588,
}

func TestShadowPriest(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassPriest,
			Race:       proto.Race_RaceTroll,
			OtherRaces: []proto.Race{proto.Race_RaceUndead, proto.Race_RaceDwarf, proto.Race_RaceNightElf},

			SpecOptions: core.SpecOptionsCombo{
				Label: "Shadow",
				SpecOptions: &proto.Player_DpsPriest{
					DpsPriest: &proto.DpsPriest{
						Options: &proto.DpsPriest_Options{
							ClassOptions: &proto.PriestOptions{
								// Begin the sim already in Shadowform so the opener
								// doesn't spend a GCD casting it.
								PreShadowform: true,
								Armor:         proto.PriestOptions_InnerFire,
							},
						},
					},
				},
			},

			GearSet:     shadowTestingBuild.GearSet,
			Talents:     shadowTestingBuild.Talents,
			Consumables: defaultConsumables,
			Profession1: proto.Profession_Enchanting,
			Profession2: proto.Profession_Tailoring,

			Rotation: core.RotationCombo{Label: "default", Rotation: shadowTestingBuild.Rotation.Rotation},

			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeStaff,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeOffHand,
				},
				ArmorType: proto.ArmorType_ArmorTypeCloth,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeWand,
				},
				// Blacklist melee enchants that appear on cloth-relevant slots but
				// are never used by casters.
				EnchantBlacklist: []int32{2673, 3225, 3273},
			},
		},
	}))
}

// The priest's own Prayer of Fortitude bids in the raid buff's category. Alone, the cast adds the
// row's stamina for the row's mana; Arcane Brilliance lifts the priest's mana pool above the cost. With the raid buff up, the cast is turned away, and the priest's
// copy, put up anyway as another priest's cast would, adds nothing.
func TestPrayerOfFortitudeCountsOnceWithTheRaidBuff(t *testing.T) {
	rank := spellData.PrayerOfFortitude.Highest()
	rowStamina := rank.EffectN(1).BaseValue()

	newSim := func(t *testing.T, raidBuffs *proto.RaidBuffs) (*core.Simulation, *Priest, *core.Spell) {
		t.Helper()
		player := &proto.Player{
			Name:      "Prayer",
			Race:      proto.Race_RaceTroll,
			Class:     proto.Class_ClassPriest,
			Equipment: shadowTestingBuild.GearSet.GearSet,
			Spec: &proto.Player_DpsPriest{
				DpsPriest: &proto.DpsPriest{
					Options: &proto.DpsPriest_Options{ClassOptions: &proto.PriestOptions{}},
				},
			},
			Rotation: &proto.APLRotation{},
		}
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, raidBuffs, &proto.Debuffs{}),
			Encounter:  core.MakeSingleTargetEncounter(0),
			SimOptions: &proto.SimOptions{RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()

		priest := sim.Raid.Parties[0].Players[0].(*Priest)
		spell := priest.GetSpell(core.ActionID{SpellID: rank.ID})
		if spell == nil {
			t.Fatalf("the priest has no %s %d", rank.Name, rank.ID)
		}
		return sim, priest, spell
	}

	t.Run("alone the cast adds the row's stamina for the row's mana", func(t *testing.T) {
		sim, priest, spell := newSim(t, &proto.RaidBuffs{ArcaneBrilliance: true})
		stamina, mana := priest.GetStats()[stats.Stamina], priest.CurrentMana()

		if !spell.Cast(sim, &priest.Unit) {
			t.Fatal("the priest could not cast Prayer of Fortitude with nothing in its category")
		}
		if got := priest.GetStats()[stats.Stamina] - stamina; got != rowStamina {
			t.Errorf("the cast added %v stamina, want the game files' %v", got, rowStamina)
		}
		if got := mana - priest.CurrentMana(); got != rank.Cost() {
			t.Errorf("the cast spent %v mana, want the game files' %v", got, rank.Cost())
		}
	})

	t.Run("the raid buff holds the category", func(t *testing.T) {
		sim, priest, spell := newSim(t, &proto.RaidBuffs{PrayerOfFortitude: true})
		stamina := priest.GetStats()[stats.Stamina]

		if spell.Cast(sim, &priest.Unit) {
			t.Error("the priest cast Prayer of Fortitude while the raid buff holds its category")
		}
		own := priest.GetAura("Prayer of Fortitude (Player)")
		if own == nil {
			t.Fatal("the priest carries no aura labelled \"Prayer of Fortitude (Player)\"")
		}
		own.Activate(sim)
		if got := priest.GetStats()[stats.Stamina]; got != stamina {
			t.Errorf("stamina is %v with both copies up, want the raid buff's %v counted once", got, stamina)
		}
	})
}

func TestDivineAegisMetricsOnlyRecordDirectCriticalHeals(t *testing.T) {
	sim, priest := newDiscSim(t)
	if priest.Talents.DivineAegis == 0 {
		t.Fatal("test talents do not enable Divine Aegis")
	}
	target := &priest.Unit
	spellForMask := func(mask int64) *core.Spell {
		for _, spell := range priest.Spellbook {
			if spell.Matches(mask) {
				return spell
			}
		}
		return nil
	}
	const divineAegisTriggeredID = 431624
	bindingHeal := spellForMask(PriestSpellBindingHeal)
	if bindingHeal == nil {
		t.Fatal("Binding Heal was not registered")
	}
	divineAegis := priest.GetSpell(core.ActionID{SpellID: divineAegisTriggeredID, Tag: bindingHeal.ActionID.SpellID})
	if divineAegis == nil {
		t.Fatal("Divine Aegis metrics were not registered for Binding Heal")
	}

	priest.OnHealDealt(sim, bindingHeal, &core.SpellResult{Target: target, Outcome: core.OutcomeCrit, Damage: 100})
	for sim.CurrentTime <= core.SpellBatchWindow && !sim.Step() {
	}
	if got := divineAegis.SpellMetrics[target.UnitIndex].TotalShielding; got <= 0 {
		t.Fatalf("Divine Aegis recorded %v shielding after a direct critical heal, want a positive absorb", got)
	}
	if got := divineAegis.SpellMetrics[target.UnitIndex].Hits; got != 1 {
		t.Fatalf("Divine Aegis recorded %d hits after one critical heal, want 1", got)
	}
	fillShield := func() {
		priest.OnHealDealt(sim, bindingHeal, &core.SpellResult{Target: target, Outcome: core.OutcomeCrit, Damage: 1e9})
		until := sim.CurrentTime + core.SpellBatchWindow
		for sim.CurrentTime <= until && !sim.Step() {
		}
	}
	fillShield()
	capacity := divineAegis.SpellMetrics[target.UnitIndex].TotalShielding
	fillShield()
	if got := divineAegis.SpellMetrics[target.UnitIndex].TotalShielding; got != capacity {
		t.Errorf("refreshing a full Divine Aegis shield changed credited capacity from %v to %v", capacity, got)
	}

	periodicSim, periodicPriest := newDiscSim(t)
	renew := func() *core.Spell {
		for _, spell := range periodicPriest.Spellbook {
			if spell.Matches(PriestSpellRenew) {
				return spell
			}
		}
		return nil
	}()
	if renew == nil {
		t.Fatal("Renew was not registered")
	}
	periodicTarget := &periodicPriest.Unit
	if periodicPriest.GetSpell(core.ActionID{SpellID: divineAegisTriggeredID, Tag: renew.ActionID.SpellID}) != nil {
		t.Error("Divine Aegis metrics were registered for Renew, which lands no direct heal")
	}
	periodicPriest.OnPeriodicHealDealt(periodicSim, renew, &core.SpellResult{Target: periodicTarget, Outcome: core.OutcomeCrit, Damage: 100})
	for periodicSim.CurrentTime <= core.SpellBatchWindow && !periodicSim.Step() {
	}
	for _, spell := range periodicPriest.Spellbook {
		if spell.ActionID.SpellID != divineAegisTriggeredID {
			continue
		}
		if got := spell.SpellMetrics[periodicTarget.UnitIndex].TotalShielding; got != 0 {
			t.Errorf("Divine Aegis %v recorded %v shielding after a periodic critical heal, want 0", spell.ActionID, got)
		}
	}
}

func TestPowerWordShieldRecordsAppliedAbsorbOnce(t *testing.T) {
	sim, priest := newDiscSim(t)
	target := &priest.Unit
	var spell *core.Spell
	for _, candidate := range priest.Spellbook {
		if candidate.Matches(PriestSpellPowerWordShield) {
			spell = candidate
		}
	}
	if spell == nil {
		t.Fatal("Power Word: Shield was not registered")
	}
	if !spell.Cast(sim, target) {
		t.Fatal("Power Word: Shield could not be cast on the priest")
	}
	metrics := spell.SpellMetrics[target.UnitIndex]
	if metrics.TotalShielding <= 0 {
		t.Fatalf("Power Word: Shield recorded %v shielding, want a positive applied absorb", metrics.TotalShielding)
	}
	if metrics.Hits != 1 {
		t.Errorf("Power Word: Shield recorded %d hits for one applied shield, want 1", metrics.Hits)
	}

	spell.SkipCastAndApplyEffects(sim, target)
	if got := spell.SpellMetrics[target.UnitIndex].TotalShielding; got != metrics.TotalShielding {
		t.Errorf("replacing an unused shield changed credited capacity from %v to %v", metrics.TotalShielding, got)
	}

	var lowerRank *core.Spell
	for _, candidate := range priest.Spellbook {
		if candidate.Matches(PriestSpellPowerWordShield) && candidate.Rank < spell.Rank {
			lowerRank = candidate
			break
		}
	}
	if lowerRank == nil {
		t.Fatal("no lower rank of Power Word: Shield was registered")
	}
	lowerRank.SkipCastAndApplyEffects(sim, target)
	if got := lowerRank.SpellMetrics[target.UnitIndex].TotalShielding; got != 0 {
		t.Errorf("replacing an unused shield with a weaker one credited %v capacity, want 0", got)
	}
}
