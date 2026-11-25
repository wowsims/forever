package priest

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	RegisterPriest()
	common.RegisterAllEffects()
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
// row's stamina for the row's mana; Arcane Intellect lifts the priest's mana pool above the cost. With the raid buff up, the cast is turned away, and the priest's
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
		sim, priest, spell := newSim(t, &proto.RaidBuffs{ArcaneIntellect: true})
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
		sim, priest, spell := newSim(t, &proto.RaidBuffs{PowerWordFortitude: true})
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

// Spirit Tap comes with a killing blow or the death of an enemy with the priest's Vampiric Embrace on it, and not
// with an unafflicted death the raid gets no kill for. An afflicted death always grants it, once; any other killing
// blow rolls the rank's chance.
func TestSpiritTapComesWithAKillingBlowOrAnAfflictedDeath(t *testing.T) {
	newSim := func(t *testing.T, talents string, points int32) (*core.Simulation, *Priest, *core.Aura) {
		t.Helper()
		encounter := core.MakeSingleTargetEncounter(0)
		encounter.Targets = []*proto.Target{core.FreshDefaultTargetConfig()}
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid: core.SinglePlayerRaidProto(&proto.Player{
				Name:          "Spirit Tap",
				Race:          proto.Race_RaceTroll,
				Class:         proto.Class_ClassPriest,
				Equipment:     shadowTestingBuild.GearSet.GearSet,
				TalentsString: talents,
				Spec: &proto.Player_DpsPriest{
					DpsPriest: &proto.DpsPriest{
						Options: &proto.DpsPriest_Options{ClassOptions: &proto.PriestOptions{}},
					},
				},
				Rotation: &proto.APLRotation{},
			}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  encounter,
			SimOptions: &proto.SimOptions{RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()

		priest := sim.Raid.Parties[0].Players[0].(*Priest)
		if priest.Talents.SpiritTap != points || priest.VampiricEmbrace == nil {
			t.Fatalf("the build has Spirit Tap %d and Vampiric Embrace %v, want %d and the spell",
				priest.Talents.SpiritTap, priest.VampiricEmbrace, points)
		}
		buff := priest.GetAura(spellData.SpiritTapTriggered.Highest().Name)
		if buff == nil {
			t.Fatal("the priest carries no Spirit Tap buff")
		}
		return sim, priest, buff
	}
	afflict := func(sim *core.Simulation, priest *Priest, target *core.Target) {
		for _, auras := range priest.VampiricEmbrace.RelatedAuraArrays {
			auras.Get(&target.Unit).Activate(sim)
		}
	}

	t.Run("each death grants once at the full rank", func(t *testing.T) {
		// Spirit Tap 5/5, where the row's chance is certain, and Vampiric Embrace in the Shadow tree.
		sim, priest, buff := newSim(t, "--00500000001", 5)
		target := sim.Encounter.AllTargets[0]
		grants := 0
		sim.Log = func(message string, vals ...any) {
			line := fmt.Sprintf(message, vals...)
			if strings.Contains(line, buff.ActionID.String()) && (strings.Contains(line, "Aura gained") || strings.Contains(line, "Aura refreshed")) {
				grants++
			}
		}

		for _, c := range []struct {
			name      string
			afflicted bool
			killShare float64
			grants    int
		}{
			{"an unafflicted death with no kill credit", false, 0, 0},
			{"an unafflicted killing blow", false, 1, 1},
			{"an afflicted death with no kill credit", true, 0, 1},
			{"an afflicted killing blow", true, 1, 1},
		} {
			buff.Deactivate(sim)
			target.Enable(sim)
			if c.afflicted {
				afflict(sim, priest, target)
			}
			grants = 0

			target.Die(sim, c.killShare)
			if grants != c.grants || buff.IsActive() != (c.grants > 0) {
				t.Errorf("%s granted Spirit Tap %d times and left it up %v, want %d", c.name, grants, buff.IsActive(), c.grants)
			}
		}
	})

	t.Run("a killing blow rolls the rank's chance and an afflicted death always grants", func(t *testing.T) {
		// Spirit Tap 1/5 and Vampiric Embrace in the Shadow tree.
		const deaths = 1000
		chance := spellData.SpiritTap.EffectAt(1).FractionAt(1)
		sim, priest, buff := newSim(t, "--00100000001", 1)
		target := sim.Encounter.AllTargets[0]

		for _, c := range []struct {
			name      string
			afflicted bool
			killShare float64
			want      float64
		}{
			{"a killing blow", false, 1, chance},
			{"an afflicted death", true, 0, 1},
		} {
			grants := 0
			for range deaths {
				buff.Deactivate(sim)
				target.Enable(sim)
				if c.afflicted {
					afflict(sim, priest, target)
				}
				target.Die(sim, c.killShare)
				if buff.IsActive() {
					grants++
				}
			}
			if got := float64(grants) / deaths; math.Abs(got-c.want) > c.want/2 {
				t.Errorf("%s granted Spirit Tap on %v of deaths, want about %v", c.name, got, c.want)
			}
		}
	})
}

// Vampiric Embrace heals each party member the row's share of the priest's Shadow damage, and the priest that
// share raised by the second effect's percent.
func TestVampiricEmbraceHealsThePriestMoreThanTheParty(t *testing.T) {
	priestPlayer := func(name, talents string) *proto.Player {
		return &proto.Player{
			Name:          name,
			Race:          proto.Race_RaceTroll,
			Class:         proto.Class_ClassPriest,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: talents,
			Spec: &proto.Player_DpsPriest{
				DpsPriest: &proto.DpsPriest{
					Options: &proto.DpsPriest_Options{ClassOptions: &proto.PriestOptions{}},
				},
			},
			Rotation: &proto.APLRotation{},
		}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{
			// Vampiric Embrace in the Shadow tree.
			priestPlayer("Vampiric Embrace", "--00000000001"),
			priestPlayer("Party Member", ""),
		}, Buffs: &proto.PartyBuffs{}}}, Buffs: &proto.RaidBuffs{}, Debuffs: &proto.Debuffs{}},
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	priest, member := sim.Raid.Parties[0].Players[0].(*Priest), sim.Raid.Parties[0].Players[1].(*Priest)
	rank := spellData.VampiricEmbrace.Highest()
	blastRank := spellData.MindBlast.Highest()
	heal := priest.GetSpell(core.ActionID{SpellID: 15290})
	mindBlast := priest.GetSpell(core.ActionID{SpellID: blastRank.ID})
	if priest.VampiricEmbrace == nil || heal == nil || mindBlast == nil {
		t.Fatalf("the priest has Vampiric Embrace %v, its heal %v and %s %d %v, want all three",
			priest.VampiricEmbrace, heal, blastRank.Name, blastRank.ID, mindBlast)
	}
	target := sim.Encounter.AllTargetUnits[0]
	for _, auras := range priest.VampiricEmbrace.RelatedAuraArrays {
		auras.Get(target).Activate(sim)
	}

	for range 10 {
		mindBlast.SkipCastAndApplyEffects(sim, target)
	}
	for end := sim.CurrentTime + time.Second; sim.CurrentTime < end && !sim.Step(); {
	}

	damage := mindBlast.SpellMetrics[target.UnitIndex].TotalDamage
	own, party := heal.SpellMetrics[priest.UnitIndex].TotalHealing, heal.SpellMetrics[member.UnitIndex].TotalHealing
	if want := damage * rank.EffectN(1).Percent(); party == 0 || math.Abs(party-want) > 1e-6*want {
		t.Errorf("the party member was healed %v from %v Shadow damage, want the row's share %v", party, damage, want)
	}
	if want := party * (1 + rank.EffectN(2).Percent()); math.Abs(own-want) > 1e-6*want {
		t.Errorf("the priest was healed %v where the party member got %v, want %v", own, party, want)
	}
}

// Early Demise reads the health of each Shadow Word: Death's own target: the crit bonus is off for a cast
// while the target is above its threshold and on once the target is below it.
func TestEarlyDemiseFollowsTheCastsTarget(t *testing.T) {
	encounter := core.MakeSingleTargetEncounter(0)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Name:      "Early Demise",
			Race:      proto.Race_RaceTroll,
			Class:     proto.Class_ClassPriest,
			Equipment: &proto.EquipmentSpec{},
			// Early Demise 2/2 in the Shadow tree.
			TalentsString: "--000000000000002",
			Spec: &proto.Player_DpsPriest{
				DpsPriest: &proto.DpsPriest{
					Options: &proto.DpsPriest_Options{ClassOptions: &proto.PriestOptions{}},
				},
			},
			Rotation: &proto.APLRotation{},
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	priest := sim.Raid.Parties[0].Players[0].(*Priest)
	if priest.Talents.EarlyDemise != 2 {
		t.Fatalf("the build has Early Demise %d, want 2", priest.Talents.EarlyDemise)
	}
	rank := spellData.ShadowWordDeath.Highest()
	spell := priest.GetSpell(core.ActionID{SpellID: rank.ID})
	if spell == nil {
		t.Fatalf("the priest has no %s %d", rank.Name, rank.ID)
	}
	target := sim.Encounter.AllTargetUnits[0]
	base := spell.BonusCritPercent
	bonus := spellData.EarlyDemise.EffectAt(1).ValueAt(2)

	spell.SkipCastAndApplyEffects(sim, target)
	if got := spell.BonusCritPercent - base; math.Abs(got) > 1e-9 {
		t.Errorf("Shadow Word: Death at a target above 20%% has %v bonus crit, want 0", got)
	}
	for !target.IsExecutePhase20() && !sim.Step() {
	}
	spell.SkipCastAndApplyEffects(sim, target)
	if got := spell.BonusCritPercent - base; math.Abs(got-bonus) > 1e-9 {
		t.Errorf("Shadow Word: Death at a target below 20%% has %v bonus crit, want %v", got, bonus)
	}
}

// Starshards' ticks roll their partial resist against the target's whole average resist, where other
// channels' ticks roll against a tenth of it (EV-0142).
func TestStarshardsTicksMeetTheTargetsFullResistance(t *testing.T) {
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Duration = 3600
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Name:      "Starshards",
			Race:      proto.Race_RaceNightElf,
			Class:     proto.Class_ClassPriest,
			Equipment: &proto.EquipmentSpec{},
			Spec: &proto.Player_DpsPriest{
				DpsPriest: &proto.DpsPriest{
					Options: &proto.DpsPriest_Options{ClassOptions: &proto.PriestOptions{}},
				},
			},
			Rotation: &proto.APLRotation{},
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	priest := sim.Raid.Parties[0].Players[0].(*Priest)
	target := sim.Encounter.AllTargetUnits[0]
	// An average resist of 300 / (300 + 8 * 60 + 19) leaves no unresisted step at the full share, and
	// about seven in ten unresisted at a tenth.
	target.AddStatDynamic(sim, stats.ArcaneResistance, 300)
	rank := StarshardsRankMap.Highest()
	starshards := priest.GetSpell(core.ActionID{SpellID: rank.ID})
	if starshards == nil {
		t.Fatalf("the priest has no %s %d", rank.Name, rank.ID)
	}
	tick := starshards.ActionID.String() + " tick "
	manaMetrics := priest.NewManaMetrics(starshards.ActionID)
	landed, unresisted := 0, 0
	sim.Log = func(message string, vals ...any) {
		line := fmt.Sprintf(message, vals...)
		if !strings.Contains(line, tick) {
			return
		}
		landed++
		if !strings.Contains(line, "% Resist)") {
			unresisted++
		}
	}

	for range 40 {
		priest.AddMana(sim, priest.MaxMana(), manaMetrics)
		if !starshards.Cast(sim, target) {
			t.Fatalf("Starshards did not cast at %v", sim.CurrentTime)
		}
		for end := sim.CurrentTime + 31*time.Second; sim.CurrentTime < end && !sim.Step(); {
		}
	}

	if landed < 100 || unresisted != 0 {
		t.Errorf("%d of %d Starshards ticks took no partial resist, want none of at least 100", unresisted, landed)
	}
}

// Penance's heal bolts crit with Revelation's charge up and leave it: every heal of the channel crits and
// the charge outlasts the channel (EV-0149).
func TestPenanceHealsCritAndLeaveRevelationsCharge(t *testing.T) {
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Duration = 3600
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for slot := range items {
		items[slot] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 1613, Enchant: 8217}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Name:          "Penance",
			Race:          proto.Race_RaceTroll,
			Class:         proto.Class_ClassPriest,
			Equipment:     &proto.EquipmentSpec{Items: items},
			TalentsString: "000000000000001",
			Spec: &proto.Player_DpsPriest{
				DpsPriest: &proto.DpsPriest{
					Options: &proto.DpsPriest_Options{ClassOptions: &proto.PriestOptions{}},
				},
			},
			Rotation: &proto.APLRotation{},
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	priest := sim.Raid.Parties[0].Players[0].(*Priest)
	healBolt := spellData.Penance.Highest().Refs()[1].ID
	channel := priest.GetSpell(core.ActionID{SpellID: healBolt, Tag: 1})
	bolt := priest.GetSpell(core.ActionID{SpellID: healBolt})
	buff := priest.GetAuraByID(core.ActionID{SpellID: 1323377})
	if !priest.Talents.Penance || channel == nil || bolt == nil || buff == nil {
		t.Fatalf("Penance %v, its heal %v and bolt %v, Revelation %v; want all", priest.Talents.Penance, channel, bolt, buff)
	}

	buff.Activate(sim)
	buff.SetStacks(sim, buff.MaxStacks)
	if !channel.Cast(sim, &priest.Unit) {
		t.Fatal("Penance's heal did not cast")
	}
	for end := sim.CurrentTime + 5*time.Second; sim.CurrentTime < end && !sim.Step(); {
	}
	hits, crits := int32(0), int32(0)
	for _, metrics := range bolt.SpellMetrics {
		hits += metrics.Hits
		crits += metrics.Crits
	}
	if hits != 0 || crits < 2 {
		t.Errorf("Penance's heals with Revelation up landed %d without a crit and %d crits, want every one of at least 2 to crit", hits, crits)
	}
	if !buff.IsActive() {
		t.Error("Penance's heals spent Revelation's charge")
	}
}
