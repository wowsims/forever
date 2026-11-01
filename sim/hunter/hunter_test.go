package hunter

import (
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"

	"testing"
)

func init() {
	RegisterHunter()
	common.RegisterAllEffects()
}

func TestHunter(t *testing.T) {
	weaveRotation := core.GetAplRotation("../../ui/specs/hunter/dps/apls", "default")
	weaveRotation.Label = "weave"

	turretRotation := core.GetAplRotation("../../ui/specs/hunter/dps/apls", "default").Rotation
	stopWeaving(turretRotation)

	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassHunter,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceNightElf},
			GearSet:    core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1"),
			Talents:    DefaultMMTalents,
			OtherTalentSets: []core.TalentsCombo{
				{Label: "SV", Talents: DefaultSVTalents},
			},
			Consumables:      DefaultConsumables,
			SpecOptions:      core.SpecOptionsCombo{Label: "Default", SpecOptions: DefaultOptions},
			StartingDistance: 12,
			ReactionTimeMs:   250,
			Profession1:      proto.Profession_Engineering,
			Profession2:      proto.Profession_Blacksmithing,

			Rotation: core.GetAplRotation("../../ui/specs/hunter/dps/apls", "default"),
			OtherRotations: []core.RotationCombo{
				core.GetAplRotation("../../ui/specs/hunter/dps/apls", "clean_weave"),
				core.RotationWithSpellIds("default_book_ranks", weaveRotation, map[int32]int32{13555: 25295, 14322: 25296}),
			},

			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypeMail,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeBow,
					proto.RangedWeaponType_RangedWeaponTypeCrossbow,
					proto.RangedWeaponType_RangedWeaponTypeGun,
				},
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeAxe,
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypePolearm,
					proto.WeaponType_WeaponTypeStaff,
					proto.WeaponType_WeaponTypeSword,
				},
				HandTypes: []proto.HandType{
					proto.HandType_HandTypeMainHand,
					proto.HandType_HandTypeOffHand,
					proto.HandType_HandTypeOneHand,
					proto.HandType_HandTypeTwoHand,
				},
			},
		},
	}))
}

var DefaultOptions = &proto.Player_Hunter{
	Hunter: &proto.Hunter{
		Options: &proto.Hunter_Options{
			ClassOptions: &proto.HunterOptions{
				Ammo:        proto.HunterOptions_ThoriumHeadedArrow,
				QuiverBonus: proto.HunterOptions_AncientSinewWrappedLamina,
				PetType:     proto.HunterOptions_PetNone,
			},
		},
	},
}

// Every pet family, fully trained, on one race, talent set and rotation: the pet's own suite, so
// that each family's abilities run without multiplying the hunter's. The talents are Beast Mastery,
// the pet's tree, and the rotation sends two hawks for each Arcane Shot, the two spells sharing a cooldown.
func TestHunterPets(t *testing.T) {
	turretRotation := core.GetAplRotation("../../ui/specs/hunter/dps/apls", "default")
	turretRotation.Label = "turret"
	stopWeaving(turretRotation.Rotation)
	setVariable(turretRotation.Rotation, "Use Summon Hawk", "true")

	families := petOptions()
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:            proto.Class_ClassHunter,
			Race:             proto.Race_RaceOrc,
			GearSet:          core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1"),
			Talents:          DefaultBMTalents,
			Consumables:      DefaultConsumables,
			SpecOptions:      families[0],
			OtherSpecOptions: families[1:],
			StartingDistance: 12,
			ReactionTimeMs:   250,
			Rotation:         turretRotation,
			ItemFilter:       core.ItemFilter{ArmorType: proto.ArmorType_ArmorTypeMail},
		},
	}))
}

// Auto Shot needs the bow, gun or crossbow its row names: a thrown weapon in the ranged slot fires
// none and starts no windup, the p1 bow fires them.
func TestAutoShotNeedsABowGunOrCrossbow(t *testing.T) {
	const thrown = 20814 // Master's Throwing Dagger
	if item := core.GetItemByID(thrown); item == nil || item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeThrown {
		t.Fatalf("item %d is not a thrown weapon", thrown)
	}
	windup := "Casting " + core.ActionID{OtherID: proto.OtherAction_OtherActionShoot}.String()

	autoShots := func(rangedID int32) (casts int32, windups int) {
		gear := googleProto.Clone(core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1").GearSet).(*proto.EquipmentSpec)
		if rangedID != 0 {
			gear.Items[proto.ItemSlot_ItemSlotRanged] = &proto.ItemSpec{Id: rangedID}
		}
		player := &proto.Player{
			Name:               "Auto Shot",
			Race:               proto.Race_RaceOrc,
			Class:              proto.Class_ClassHunter,
			Equipment:          gear,
			TalentsString:      DefaultMMTalents,
			Spec:               DefaultOptions,
			Rotation:           &proto.APLRotation{},
			DistanceFromTarget: 20,
		}
		result := core.RunRaidSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  core.MakeSingleTargetEncounter(0),
			SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1, Debug: true},
		})
		if result.Error != nil {
			t.Fatalf("the sim failed: %s", result.Error.Message)
		}
		for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
			if action.Id.GetOtherId() == proto.OtherAction_OtherActionShoot {
				for _, target := range action.Targets {
					casts += target.Casts
				}
			}
		}
		return casts, strings.Count(result.Logs, windup)
	}

	if casts, windups := autoShots(0); casts == 0 || windups == 0 {
		t.Errorf("the p1 bow fired %d Auto Shots after %d windups, want some", casts, windups)
	}
	if casts, windups := autoShots(thrown); casts != 0 || windups != 0 {
		t.Errorf("a thrown weapon fired %d Auto Shots after %d windups, want none", casts, windups)
	}
}

// Sniper Shot's buff carries Auto Shot's reach past the 41 yards Hawk Eye gives it, for three Shots: a
// hunter at 45 yards fires no Auto Shot until Sniper Shot lands, then lands at most three per Sniper
// Shot, since a miss keeps its charge.
func TestSniperShotExtendsAutoShotRange(t *testing.T) {
	shots := func(castSniperShot bool) (sniperShots, autoShots, landed int32) {
		rotation := &proto.APLRotation{Type: proto.APLRotation_TypeAPL}
		if castSniperShot {
			rotation.PriorityList = []*proto.APLListItem{{Action: &proto.APLAction{Action: &proto.APLAction_CastSpell{
				CastSpell: &proto.APLActionCastSpell{SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: sniperShotRank.ID}}},
			}}}}
		}
		player := &proto.Player{
			Name:               "Sniper Shot",
			Race:               proto.Race_RaceOrc,
			Class:              proto.Class_ClassHunter,
			Equipment:          core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1").GearSet,
			TalentsString:      DefaultMMTalents,
			Spec:               DefaultOptions,
			Rotation:           rotation,
			DistanceFromTarget: 45,
		}
		result := core.RunRaidSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  core.MakeSingleTargetEncounter(0),
			SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
		})
		if result.Error != nil {
			t.Fatalf("the sim failed: %s", result.Error.Message)
		}
		for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
			for _, target := range action.Targets {
				switch {
				case action.Id.GetOtherId() == proto.OtherAction_OtherActionShoot:
					autoShots += target.Casts
					landed += target.Hits + target.Crits
				case action.Id.GetSpellId() == sniperShotRank.ID:
					sniperShots += target.Casts
				}
			}
		}
		return sniperShots, autoShots, landed
	}

	if _, autoShots, _ := shots(false); autoShots != 0 {
		t.Errorf("a hunter at 45 yards fired %d Auto Shots with no Sniper Shot, want none", autoShots)
	}
	sniperShots, _, landed := shots(true)
	if sniperShots == 0 || landed == 0 || landed > 3*sniperShots {
		t.Errorf("%d Sniper Shots let %d Auto Shots land from 45 yards, want some and at most 3 each", sniperShots, landed)
	}
}

// One sting per hunter on a target: a hunter's sting drops that hunter's other sting and leaves
// another hunter's. Two hunters' Scorpid Stings both stay up, and their -2% hit counts once.
func TestEachHunterKeepsOneStingPerTarget(t *testing.T) {
	player := func(name string) *proto.Player {
		return &proto.Player{Name: name, Race: proto.Race_RaceOrc, Class: proto.Class_ClassHunter,
			Equipment: &proto.EquipmentSpec{}, Spec: DefaultOptions, Rotation: &proto.APLRotation{}}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: &proto.Raid{
			Parties: []*proto.Party{{Players: []*proto.Player{player("A"), player("B")}, Buffs: &proto.PartyBuffs{}}},
			Buffs:   &proto.RaidBuffs{},
			Debuffs: &proto.Debuffs{},
		},
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{},
	}, simsignals.CreateSignals())
	sim.Reset()
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	other := sim.Raid.Parties[0].Players[1].(*Hunter)
	target := sim.Encounter.AllTargetUnits[0]

	serpent := hunter.GetSpell(core.ActionID{SpellID: serpentStingBookRank.ID}).Dot(target)
	scorpid := target.GetAura("Scorpid Sting-" + hunter.Label)
	otherScorpid := target.GetAura("Scorpid Sting-" + other.Label)

	// What a landed sting does: drop the hunter's other sting, then go up.
	dropOtherSting(sim, otherScorpid)
	otherScorpid.Activate(sim)
	dropOtherSting(sim, scorpid)
	scorpid.Activate(sim)
	dropOtherSting(sim, serpent.Aura)
	serpent.Apply(sim)
	if scorpid.IsActive() {
		t.Error("Serpent Sting landed beside the hunter's own Scorpid Sting, want it dropped")
	}

	dropOtherSting(sim, scorpid)
	scorpid.Activate(sim)
	if serpent.IsActive() {
		t.Error("Scorpid Sting landed beside the hunter's own Serpent Sting, want it dropped")
	}
	if !otherScorpid.IsActive() {
		t.Error("the hunter's stings dropped the other hunter's Scorpid Sting, want it on the target")
	}
	if got := target.GetStats()[stats.PhysicalHitPercent]; got != -2 {
		t.Errorf("two hunters' Scorpid Stings moved hit by %v, want -2 once", got)
	}
}

// The hunter's own Hunter's Mark carries the mark's tag, which Expose Prey reads before it opens
// Mongoose Bite, and gives every attacker its ranged attack power.
func TestTheHuntersOwnMarkCarriesTheMarkTag(t *testing.T) {
	player := &proto.Player{
		Name:          "Mark",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassHunter,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: DefaultSVTalents,
		Spec:          DefaultOptions,
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	target := sim.Encounter.AllTargetUnits[0]

	mark := hunter.GetSpell(core.ActionID{SpellID: huntersMarkRank.ID})
	if mark == nil {
		t.Fatal("the hunter has no Hunter's Mark")
	}
	mark.SkipCastAndApplyEffects(sim, target)
	if !target.HasActiveAuraWithTag(buffs.HuntersMarkCategory) {
		t.Error("the hunter's mark is up but no aura tagged HuntersMark is, so Expose Prey never sees it")
	}
	if got, want := target.PseudoStats.BonusRangedAttackPower, buffs.HuntersMarkValue(0); got != want {
		t.Errorf("the mark gives attackers %v ranged attack power, want %v", got, want)
	}
}

// A pet takes the hunter's ranged hit whole as melee hit, and 17/8 of it as spell hit.
func TestPetHitIsTheHuntersRangedHit(t *testing.T) {
	_, pet, _ := newPetSim(proto.HunterOptions_Cat, nil)

	inherited := pet.hunterOwner.makeStatInheritance()(stats.Stats{stats.PhysicalHitPercent: 3, stats.RangedHitPercent: 1})
	if got := inherited[stats.PhysicalHitPercent]; got != 4 {
		t.Errorf("pet melee hit %.4f%% from 4%% ranged hit, want 4%%", got)
	}
	if got := inherited[stats.SpellHitPercent]; got != 8.5 {
		t.Errorf("pet spell hit %.4f%% from 4%% ranged hit, want 8.5%%", got)
	}
}

// Lacerating Strikes refreshes the way Deep Wounds does: a bite while the bleed is up adds the unticked
// damage to the new bleed and keeps the first bite's tick timer.
func TestLaceratingStrikesRollsItsUntickedDamageIntoTheNextBite(t *testing.T) {
	player := &proto.Player{
		Name:          "Lacerate",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassHunter,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: DefaultSVTalents,
		Spec:          DefaultOptions,
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	target := sim.Encounter.AllTargetUnits[0]

	trigger := hunter.GetAura("Lacerating Strikes - Trigger")
	bite := hunter.GetSpell(core.ActionID{SpellID: mongooseBiteRank.ID})
	if trigger == nil || bite == nil {
		t.Fatal("the SV hunter has no Lacerating Strikes or no Mongoose Bite")
	}
	bleed := hunter.GetSpell(core.ActionID{SpellID: laceratingStrikesBleed.ID}).Dot(target)
	share := spellData.LaceratingStrikes.FractionAt(1)

	stepUntil := func(what string, done func() bool) {
		for range 1000 {
			if done() {
				return
			}
			sim.Step()
		}
		t.Fatalf("the sim never reached %s", what)
	}
	landBite := func(damage float64) {
		before := bleed.OutstandingDmg()
		trigger.OnSpellHitDealt(trigger, sim, bite, &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: damage})
		stepUntil("the bite's bleed", func() bool { return bleed.OutstandingDmg() != before })
	}

	landBite(700)
	ticks := bleed.RemainingTicks()
	stepUntil("the first tick", func() bool { return bleed.RemainingTicks() < ticks })
	owed := bleed.OutstandingDmg()
	nextTick := bleed.NextTickAt()

	landBite(700)
	if got, want := bleed.OutstandingDmg(), owed+700*share; math.Abs(got-want) > 1e-6 {
		t.Errorf("a second bite left %v owed on the bleed, want the %v unticked plus %v", got, owed, 700*share)
	}
	if got := bleed.RemainingTicks(); got != ticks {
		t.Errorf("a second bite left %d ticks on the bleed, want a full %d", got, ticks)
	}
	if got := bleed.NextTickAt(); got != nextTick {
		t.Errorf("a second bite moved the next tick to %v, want it kept at %v", got, nextTick)
	}
}

// Strider Kick's 30% speed lasts the row's 3 seconds, and a walk that outlasts it ends where it was
// going: a walk out to range started 2.5 seconds after the kick runs its first half second at 130%.
func TestStriderKickSpeedRunsOutMidWalk(t *testing.T) {
	player := &proto.Player{
		Name:               "Kick",
		Race:               proto.Race_RaceOrc,
		Class:              proto.Class_ClassHunter,
		Equipment:          &proto.EquipmentSpec{},
		TalentsString:      DefaultSVTalents,
		Spec:               DefaultOptions,
		Rotation:           &proto.APLRotation{},
		DistanceFromTarget: core.MaxMeleeRange,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	kick := hunter.GetSpell(core.ActionID{SpellID: striderKickRank.ID})
	if kick == nil {
		t.Fatal("the SV hunter has no Strider Kick")
	}

	stepTo := func(at time.Duration) {
		for sim.CurrentTime < at {
			sim.AddPendingAction(&core.PendingAction{NextActionAt: at, OnAction: func(*core.Simulation) {}})
			sim.Step()
		}
	}
	kick.SkipCastAndApplyEffects(sim, sim.Encounter.AllTargetUnits[0])
	if got, want := hunter.GetMovementSpeed(), 7*(1+striderKickRank.EffectN(3).Percent()); math.Abs(got-want) > 1e-9 {
		t.Fatalf("after Strider Kick the hunter runs %v yd/s, want %v", got, want)
	}

	start := sim.CurrentTime + 2500*time.Millisecond
	stepTo(start)
	const ranged = 15.5
	hunter.MoveTo(ranged, sim)
	for i := 0; hunter.Moving && i < 1000; i++ {
		sim.Step()
	}
	fast := 0.5 * 7 * (1 + striderKickRank.EffectN(3).Percent())
	want := start + 500*time.Millisecond + time.Duration((ranged-core.MaxMeleeRange-fast)/7*float64(time.Second))
	if math.Abs(hunter.DistanceFromTarget-ranged) > 1e-6 {
		t.Errorf("the walk out ended at %v yd, want %v", hunter.DistanceFromTarget, ranged)
	}
	if got := sim.CurrentTime; (got - want).Abs() > 20*time.Millisecond {
		t.Errorf("the walk out ended at %v, want about %v", got, want)
	}
}

// Enchant Boots - Minor Speed, the boots enchant Forever offers, runs the hunter 8% faster.
func TestMinorSpeedBootsRunFaster(t *testing.T) {
	gear := core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1").GearSet
	gear = googleProto.Clone(gear).(*proto.EquipmentSpec)
	gear.Items[proto.ItemSlot_ItemSlotFeet].Enchant = 911
	player := &proto.Player{
		Name:          "Boots",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassHunter,
		Equipment:     gear,
		TalentsString: DefaultMMTalents,
		Spec:          DefaultOptions,
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	if got := hunter.GetMovementSpeed(); math.Abs(got-7*1.08) > 1e-9 {
		t.Errorf("with Minor Speed on the boots the hunter runs %v yd/s, want %v", got, 7*1.08)
	}
}

// Survivalist's Discipline shortens the traps' shared cooldown by 20% a rank: a placed Explosive Trap
// holds Immolation Trap back 30, 24 or 18 seconds.
func TestSurvivalistsDisciplineShortensTheTrapCooldown(t *testing.T) {
	for points, want := range []time.Duration{30 * time.Second, 24 * time.Second, 18 * time.Second} {
		player := &proto.Player{
			Name:          "Traps",
			Race:          proto.Race_RaceOrc,
			Class:         proto.Class_ClassHunter,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "--" + strings.Repeat("0", 14) + strconv.Itoa(points),
			Spec:          DefaultOptions,
			Rotation:      &proto.APLRotation{},
		}
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  core.MakeSingleTargetEncounter(0),
			SimOptions: &proto.SimOptions{RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()
		hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
		target := sim.Encounter.AllTargetUnits[0]

		explosive := hunter.GetSpell(core.ActionID{SpellID: explosiveTrapRank.ID})
		immolation := hunter.GetSpell(core.ActionID{SpellID: immolationTrapRank.ID})
		if explosive == nil || immolation == nil {
			t.Fatal("the hunter has no Explosive Trap or no Immolation Trap")
		}
		if !explosive.Cast(sim, target) {
			t.Fatalf("with %d points, Explosive Trap could not be placed", points)
		}
		if got := immolation.ReadyAt() - sim.CurrentTime; got != want {
			t.Errorf("with %d points, a placed Explosive Trap held Immolation Trap back %v, want %v", points, got, want)
		}
	}
}

// An Auto Shot that comes due while the hunter moves winds up a reaction time after the move ends.
func TestAutoShotWindsUpWhenTheMoveEnds(t *testing.T) {
	player := &proto.Player{
		Name:               "Shot",
		Race:               proto.Race_RaceOrc,
		Class:              proto.Class_ClassHunter,
		Equipment:          core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1").GearSet,
		TalentsString:      DefaultMMTalents,
		Spec:               DefaultOptions,
		Rotation:           &proto.APLRotation{},
		DistanceFromTarget: 20,
		ReactionTimeMs:     150,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	for i := 0; sim.CurrentTime < time.Second && i < 10000; i++ {
		sim.Step()
	}

	due := hunter.AutoAttacks.NextRangedAttackAt()
	moveEnd := due + 300*time.Millisecond
	hunter.MoveDuration(moveEnd-sim.CurrentTime, sim)
	recast := moveEnd + hunter.ReactionTime
	for i := 0; sim.CurrentTime <= recast && i < 10000; i++ {
		sim.Step()
	}
	if got, want := hunter.AutoAttacks.NextRangedAttackAt(), recast+hunter.ApplyRangedSpeed(core.RangedSwingWindup); got != want {
		t.Errorf("a shot due at %v during a move that ended at %v fires at %v, want %v", due, moveEnd, got, want)
	}
}

// A hunter walking out of the deadzone who stops right on the minimum range is back in range only as
// the walk ends; a shot that came due on the way still waits a reaction time before it winds up.
func TestAutoShotWaitsAReactionWhenTheWalkEndsOnTheMinimumRange(t *testing.T) {
	player := &proto.Player{
		Name:               "Line",
		Race:               proto.Race_RaceOrc,
		Class:              proto.Class_ClassHunter,
		Equipment:          core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1").GearSet,
		TalentsString:      DefaultMMTalents,
		Spec:               DefaultOptions,
		Rotation:           &proto.APLRotation{},
		DistanceFromTarget: 20,
		ReactionTimeMs:     150,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	for i := 0; sim.CurrentTime < time.Second && i < 10000; i++ {
		sim.Step()
	}

	// Into the deadzone past melee range, where no melee swing restarts the shot, before the next
	// shot, then out to the minimum range, arriving after it came due.
	deadzone := core.MaxMeleeRange + 1
	hunter.MoveTo(deadzone, sim)
	for i := 0; hunter.Moving && i < 10000; i++ {
		sim.Step()
	}
	due := hunter.AutoAttacks.NextRangedAttackAt()
	hunter.MoveTo(core.MinRangedRange, sim)
	walkEnd := sim.CurrentTime + time.Duration((core.MinRangedRange-deadzone)/hunter.GetMovementSpeed()*float64(time.Second))
	if walkEnd <= due {
		t.Fatalf("the walk out ends at %v, before the shot comes due at %v", walkEnd, due)
	}
	recast := walkEnd + hunter.ReactionTime
	for i := 0; sim.CurrentTime <= recast && i < 10000; i++ {
		sim.Step()
	}
	if got, want := hunter.AutoAttacks.NextRangedAttackAt(), recast+hunter.ApplyRangedSpeed(core.RangedSwingWindup); (got - want).Abs() > time.Millisecond {
		t.Errorf("a walk out ending on the minimum range at %v fires its shot at %v, want %v", walkEnd, got, want)
	}
}

// A main-hand swing restarts Auto Shot: the next shot fires one full ranged swing after the melee
// swing, the windup included.
func TestMeleeSwingRestartsAutoShot(t *testing.T) {
	player := &proto.Player{
		Name:               "Swing",
		Race:               proto.Race_RaceOrc,
		Class:              proto.Class_ClassHunter,
		Equipment:          core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1").GearSet,
		TalentsString:      DefaultMMTalents,
		Spec:               DefaultOptions,
		Rotation:           &proto.APLRotation{},
		DistanceFromTarget: core.MaxMeleeRange,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	for i := 0; sim.CurrentTime < time.Second && i < 10000; i++ {
		sim.Step()
	}

	swing := hunter.AutoAttacks.MainhandSwingAt()
	for i := 0; hunter.AutoAttacks.MainhandPreviousSwing() < swing && i < 10000; i++ {
		sim.Step()
	}
	want := swing + hunter.AutoAttacks.RangedSwingSpeed() - hunter.ApplyRangedSpeed(core.RangedSwingWindup)
	if got := hunter.AutoAttacks.NextRangedAttackAt(); got != want {
		t.Errorf("after a melee swing at %v the next Auto Shot winds up at %v, want %v", swing, got, want)
	}
}

// Minor Speed and Strider Kick are both A_MOD_INCREASE_SPEED, of which only the highest counts: with
// both the hunter runs at Strider Kick's 30%, and back at the boots' 8% once the kick runs out.
func TestMinorSpeedDoesNotStackWithStriderKick(t *testing.T) {
	gear := googleProto.Clone(core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1").GearSet).(*proto.EquipmentSpec)
	gear.Items[proto.ItemSlot_ItemSlotFeet].Enchant = 911
	player := &proto.Player{
		Name:               "Boots and Kick",
		Race:               proto.Race_RaceOrc,
		Class:              proto.Class_ClassHunter,
		Equipment:          gear,
		TalentsString:      DefaultSVTalents,
		Spec:               DefaultOptions,
		Rotation:           &proto.APLRotation{},
		DistanceFromTarget: core.MaxMeleeRange,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	for i := 0; sim.CurrentTime < 0 && i < 10000; i++ {
		sim.Step()
	}

	if got := hunter.GetMovementSpeed(); math.Abs(got-7*1.08) > 1e-9 {
		t.Fatalf("with the boots alone the hunter runs %v yd/s, want %v", got, 7*1.08)
	}
	kick := hunter.GetSpell(core.ActionID{SpellID: striderKickRank.ID})
	kick.SkipCastAndApplyEffects(sim, sim.Encounter.AllTargetUnits[0])
	if got, want := hunter.GetMovementSpeed(), 7*(1+striderKickRank.EffectN(3).Percent()); math.Abs(got-want) > 1e-9 {
		t.Errorf("with the boots and Strider Kick the hunter runs %v yd/s, want Strider Kick's %v alone", got, want)
	}
	end := sim.CurrentTime + striderKickRank.Duration() + time.Millisecond
	for i := 0; sim.CurrentTime < end && i < 10000; i++ {
		sim.AddPendingAction(&core.PendingAction{NextActionAt: end, OnAction: func(*core.Simulation) {}})
		sim.Step()
	}
	if got := hunter.GetMovementSpeed(); math.Abs(got-7*1.08) > 1e-9 {
		t.Errorf("after Strider Kick runs out the hunter runs %v yd/s, want the boots' %v", got, 7*1.08)
	}
}

// The trainer's top rank of Aspect of the Hawk takes the book rank's place and procs Quick Shots under Deadly
// Aspects. The book rank of Serpent Sting replaces the trainer's, and the trainer's does not land under it.
func TestTrainerRanksBesideBookRanks(t *testing.T) {
	sim, hunter := newTalentSim("5", proto.HunterOptions_PetNone)
	target := hunter.CurrentTarget

	hawkAura := func(rank *spelldata.Spell) *core.Aura {
		return hunter.GetAuraByID(core.ActionID{SpellID: rank.ID})
	}
	castHawk := func(rank *spelldata.Spell) {
		hunter.GetSpell(core.ActionID{SpellID: rank.ID}).SkipCastAndApplyEffects(sim, target)
	}
	castHawk(aspectOfTheHawkBookRank)
	castHawk(aspectOfTheHawkTrainerRank)
	if !hawkAura(aspectOfTheHawkTrainerRank).IsActive() || hawkAura(aspectOfTheHawkBookRank).IsActive() {
		t.Error("the trainer's Aspect of the Hawk left the book rank up, want one aspect at a time")
	}

	quickShotsAura := hunter.GetAuraByID(core.ActionID{SpellID: quickShots.ID})
	for range 500 {
		if quickShotsAura.IsActive() {
			break
		}
		hunter.AutoAttacks.RangedAuto().SkipCastAndApplyEffects(sim, target)
	}
	if !quickShotsAura.IsActive() {
		t.Error("500 Auto Shots under the trainer's Aspect of the Hawk procced no Quick Shots")
	}

	trainerSting := hunter.GetSpell(core.ActionID{SpellID: serpentStingTrainerRank.ID}).Dot(target)
	bookSting := hunter.GetSpell(core.ActionID{SpellID: serpentStingBookRank.ID}).Dot(target)
	trainerSting.Apply(sim)
	bookSting.Apply(sim)
	if trainerSting.IsActive() || !bookSting.IsActive() {
		t.Error("the book rank's Serpent Sting left the trainer's up, want it to replace the trainer's")
	}
	trainerSting.Apply(sim)
	if trainerSting.IsActive() || !bookSting.IsActive() {
		t.Error("the trainer's Serpent Sting landed under the book rank, want it refused")
	}
}

// One combo per pet family, fully trained.
// The default rotation with its "Melee weave" variable turned off: the hunter stays at range.
func stopWeaving(rotation *proto.APLRotation) {
	setVariable(rotation, "Melee weave", "false")
}

func setVariable(rotation *proto.APLRotation, name string, value string) {
	for _, variable := range rotation.ValueVariables {
		if variable.Name == name {
			variable.Value = &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: value}}}
		}
	}
}

func petOptions() []core.SpecOptionsCombo {
	var combos []core.SpecOptionsCombo
	for _, value := range slices.Sorted(maps.Values(proto.HunterOptions_PetType_value)) {
		petType := proto.HunterOptions_PetType(value)
		if petType == proto.HunterOptions_PetNone {
			continue
		}
		combos = append(combos, core.SpecOptionsCombo{
			Label: petType.String(),
			SpecOptions: &proto.Player_Hunter{
				Hunter: &proto.Hunter{
					Options: &proto.Hunter_Options{
						ClassOptions: &proto.HunterOptions{
							PetType:        petType,
							PetUptime:      1,
							CobraReflexes:  true,
							PetAttackSpeed: proto.HunterOptions_FasterAttackII,
						},
					},
				},
			},
		})
	}
	return combos
}

var DefaultBMTalents = "5320001505101251-3050552"
var DefaultMMTalents = "-3053552511523051-5"
var DefaultSVTalents = "-00505505-500250030050220151"

var DefaultConsumables = &proto.ConsumesSpec{
	ElixirIds:      []int32{13452, 250341}, // Elixir of the Mongoose, Greater Mageblood Elixir
	FoodId:         250072,                 // Swiftstrike Steak
	OffensivePotId: 250943,                 // Major Frenzy Potion
	ConjuredId:     12662,                  // Demonic Rune
	ExplosiveId:    18588,                  // Ez-Thro Dynamite II
	// Goblin Sapper Charge, Scroll of Agility IV, Scroll of Strength IV
	ExtraIds: []int32{10646, 10309, 10310},
}
