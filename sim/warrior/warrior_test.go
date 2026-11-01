package warrior_test

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/warrior"
	"github.com/wowsims/forever/sim/warrior/dps"
	"github.com/wowsims/forever/sim/warrior/protection"
	googleProto "google.golang.org/protobuf/proto"
)

func init() {
	dps.RegisterDpsWarrior()
	protection.RegisterProtectionWarrior()
	common.RegisterAllEffects()
}

// The fixtures mirror the spec suites' Default* fixtures.
var furyTalents = "30305003-05053035052210501"
var armsTalents = "30305213032315201-0505003004"
var protectionTalents = "05--255533121331021351"

var dpsOptions = &proto.Player_DpsWarrior{
	DpsWarrior: &proto.DpsWarrior{
		Options: &proto.DpsWarrior_Options{
			ClassOptions: &proto.WarriorOptions{
				UseBattleShout: true,
				DefaultStance:  proto.WarriorStance_WarriorStanceBerserker,
			},
		},
	},
}

var protectionOptions = &proto.Player_ProtectionWarrior{
	ProtectionWarrior: &proto.ProtectionWarrior{
		Options: &proto.ProtectionWarrior_Options{
			ClassOptions: &proto.WarriorOptions{
				StartingRage:   100,
				UseBattleShout: true,
				DefaultStance:  proto.WarriorStance_WarriorStanceDefensive,
			},
		},
	},
}

func orcWarriorSim(t *testing.T, gearSet string, talents string) (*core.Simulation, *dps.DpsWarrior) {
	t.Helper()
	player := &proto.Player{
		Name:          "Weapon Scope",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Equipment:     core.GetGearSet("../../ui/specs/warrior/dps/gear_sets", gearSet).GearSet,
		TalentsString: talents,
		Spec:          dpsOptions,
		Rotation:      core.GetAplRotation("../../ui/specs/warrior/dps/apls", "fury").Rotation,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*dps.DpsWarrior)
}

func newWindowSim(t *testing.T) (*core.Simulation, *dps.DpsWarrior, *core.Unit) {
	t.Helper()

	player := &proto.Player{
		Name:          "Window",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: furyTalents,
		Spec:          dpsOptions,
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim, sim.Raid.Parties[0].Players[0].(*dps.DpsWarrior), sim.Encounter.AllTargetUnits[0]
}

const (
	mainHandSlot = 14
	crusader     = 1900
)

// A protection warrior with a main-hand Crusader, Windfury Totem in the party and the preraid shield.
func shieldTestSim(t *testing.T) (*core.Simulation, *core.Character, *core.Unit) {
	t.Helper()

	gear := googleProto.Clone(core.GetGearSet("../../ui/specs/warrior/protection/gear_sets", "preraid").GearSet).(*proto.EquipmentSpec)
	gear.Items[mainHandSlot].Enchant = crusader

	player := &proto.Player{
		Name:          "Shield",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Equipment:     gear,
		TalentsString: protectionTalents,
		Spec:          protectionOptions,
		Rotation:      core.GetAplRotation("../../ui/specs/warrior/protection/apls", "default").Rotation,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{WindfuryTotem: true},
			&proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim, sim.Raid.Parties[0].Players[0].GetCharacter(), sim.Encounter.AllTargetUnits[0]
}

func shieldSpells(t *testing.T, war *core.Character) map[string]*core.Spell {
	t.Helper()
	spells := map[string]*core.Spell{
		"Shield Bash": war.GetSpell(core.ActionID{SpellID: 1672}),
		"Shield Slam": war.GetSpell(core.ActionID{SpellID: 23925}),
	}
	for name, spell := range spells {
		if spell == nil {
			t.Fatalf("the warrior has no %s", name)
		}
	}
	return spells
}

// The trainer's top rank and the rank Manual of Battle Shout VII teaches.
const (
	battleShoutTrainerRank = 11551
	battleShoutBookRank    = 25289
)

// The warrior's own shout with three pieces of Battlegear of Wrath on, through
// the class option: it raises what either rank bids for its category by the 30
// spell 23563 states.
func TestWarriorShoutsForTheTierTwoBonus(t *testing.T) {
	bidOf := func(hasT2 bool, label string) float64 {
		t.Helper()

		player := &proto.Player{
			Name:          "Battlegear of Wrath",
			Race:          proto.Race_RaceOrc,
			Class:         proto.Class_ClassWarrior,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: furyTalents,
			Spec: &proto.Player_DpsWarrior{
				DpsWarrior: &proto.DpsWarrior{
					Options: &proto.DpsWarrior_Options{
						ClassOptions: &proto.WarriorOptions{
							UseBattleShout: true,
							DefaultStance:  proto.WarriorStance_WarriorStanceBerserker,
							HasBsT2:        hasT2,
						},
					},
				},
			},
		}

		env, _, _ := core.NewEnvironment(
			core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			core.MakeSingleTargetEncounter(0), false, true)

		war := env.Raid.Parties[0].Players[0].(*dps.DpsWarrior)
		aura := war.GetAura(label)
		if aura == nil {
			t.Fatalf("the ally aura array holds no aura labelled %q", label)
		}
		return aura.ExclusiveEffects[0].Priority
	}

	for _, rank := range []struct {
		spellID int32
		label   string
	}{
		{battleShoutTrainerRank, "Battle Shout Rank 6 (Player)"},
		{battleShoutBookRank, "Battle Shout Rank 7 (Player)"},
	} {
		bare, withSet := bidOf(false, rank.label), bidOf(true, rank.label)

		if want := buffs.BattleShoutRankValue(spelldata.MustFind(rank.spellID)); bare != want {
			t.Errorf("%s: a warrior without the set bids %v, want the game files' %v", rank.label, bare, want)
		}
		if want := bare + buffs.BattleShoutT2Bonus; withSet != want {
			t.Errorf("%s: a warrior wearing the set bids %v, want %v", rank.label, withSet, want)
		}
	}
}

// The class option decides which rank's own copy the character sheet shows; both ranks stay registered.
func TestWarriorSheetShowsTheShoutRankTheOptionPicks(t *testing.T) {
	for _, row := range []struct {
		name         string
		rank7        bool
		wantOnSheet  string
		wantOffSheet string
	}{
		{"off", false, "Battle Shout Rank 6 (Player)", "Battle Shout Rank 7 (Player)"},
		{"on", true, "Battle Shout Rank 7 (Player)", "Battle Shout Rank 6 (Player)"},
	} {
		t.Run(row.name, func(t *testing.T) {
			player := &proto.Player{
				Name:          "Sheet",
				Race:          proto.Race_RaceOrc,
				Class:         proto.Class_ClassWarrior,
				Equipment:     &proto.EquipmentSpec{},
				TalentsString: furyTalents,
				Spec: &proto.Player_DpsWarrior{
					DpsWarrior: &proto.DpsWarrior{
						Options: &proto.DpsWarrior_Options{
							ClassOptions: &proto.WarriorOptions{
								UseBattleShout:       true,
								DefaultStance:        proto.WarriorStance_WarriorStanceBerserker,
								HasBattleShoutRank_7: row.rank7,
							},
						},
					},
				},
			}
			env, _, _ := core.NewEnvironment(
				core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
				core.MakeSingleTargetEncounter(0), false, true)
			war := env.Raid.Parties[0].Players[0].(*dps.DpsWarrior)

			for label, want := range map[string]core.CharacterBuildPhase{
				row.wantOnSheet:  core.CharacterBuildPhaseBuffs,
				row.wantOffSheet: core.CharacterBuildPhaseNone,
			} {
				aura := war.GetAura(label)
				if aura == nil {
					t.Fatalf("the warrior carries no aura labelled %q", label)
				}
				if aura.BuildPhase != want {
					t.Errorf("%s has build phase %v, want %v", label, aura.BuildPhase, want)
				}
			}
		})
	}
}

// The shout the rotation casts by the rank's spell id.
func battleShoutSpell(t *testing.T, war *dps.DpsWarrior, spellID int32) *core.Spell {
	t.Helper()
	spell := war.GetSpell(core.ActionID{SpellID: spellID})
	if spell == nil {
		t.Fatalf("the warrior registers no Battle Shout %d", spellID)
	}
	return spell
}

// How the warrior treats the Battle Shout category. The buff belongs to the
// unit, so the raid's shout with the set and one from the warrior standing
// next to this one fill the same category. Casting a weaker
// shout the category would turn away wastes a global and rage, and so does
// refreshing a copy with most of its three minutes left.
func TestWarriorShoutsOnlyWhenTheCategoryWillTakeIt(t *testing.T) {
	warriorProto := func(name string, hasT2 bool) *proto.Player {
		return &proto.Player{
			Name:          name,
			Race:          proto.Race_RaceOrc,
			Class:         proto.Class_ClassWarrior,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: furyTalents,
			Spec: &proto.Player_DpsWarrior{
				DpsWarrior: &proto.DpsWarrior{
					Options: &proto.DpsWarrior_Options{
						ClassOptions: &proto.WarriorOptions{
							UseBattleShout: true,
							DefaultStance:  proto.WarriorStance_WarriorStanceBerserker,
							HasBsT2:        hasT2,
						},
					},
				},
			},
		}
	}

	build := func(raidBuffs *proto.RaidBuffs, players ...*proto.Player) *core.Environment {
		t.Helper()
		raid := &proto.Raid{
			Parties:          []*proto.Party{{Players: players, Buffs: &proto.PartyBuffs{}}},
			Buffs:            raidBuffs,
			Debuffs:          &proto.Debuffs{},
			NumActiveParties: 1,
		}
		env, _, _ := core.NewEnvironment(raid, core.MakeSingleTargetEncounter(0), false, true)
		return env
	}

	withSet := buffs.BattleShoutValue(0) + buffs.BattleShoutT2Bonus

	t.Run("the raid's Rank 7 with the set outbids either rank of a warrior without it", func(t *testing.T) {
		sim := &core.Simulation{}
		env := build(&proto.RaidBuffs{BattleShout: 3}, warriorProto("Bare", false))
		war := env.Raid.Parties[0].Players[0].(*dps.DpsWarrior)

		external := war.GetAura("Battle Shout Rank 7 (External)")
		if external == nil {
			t.Fatal("the raid's copy is not registered")
		}
		raidBid := buffs.BattleShoutRankValue(spelldata.MustFind(battleShoutBookRank)) + buffs.BattleShoutT2Bonus
		external.Activate(sim)
		if !external.IsActive() || external.ExclusiveEffects[0].Priority != raidBid {
			t.Fatalf("the raid's copy is active %v at %v, want it up at %v",
				external.IsActive(), external.ExclusiveEffects[0].Priority, raidBid)
		}

		for _, spellID := range []int32{battleShoutTrainerRank, battleShoutBookRank} {
			if battleShoutSpell(t, war, spellID).ExtraCastCondition(sim, &war.Unit) {
				t.Errorf("the warrior would shout %d for %v while a %v copy holds the category, wasting the global",
					spellID, buffs.BattleShoutRankValue(spelldata.MustFind(spellID)), raidBid)
			}
		}
	})

	// The copy holding the category is the warrior's own, shared with a raid
	// member wearing the set, so it bids the set's 30 more than this warrior is
	// worth. What decides the cast is how long it has left, not what it bids.
	for _, row := range []struct {
		name      string
		remaining time.Duration
		want      bool
	}{
		{"a warrior leaves its own copy alone while it has time to run", time.Second * 10, false},
		{"a warrior refreshes its own copy as it runs out", time.Second * 2, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			sim := &core.Simulation{}
			env := build(&proto.RaidBuffs{}, warriorProto("Set", true), warriorProto("Bare", false))
			bare := env.Raid.Parties[0].Players[1].(*dps.DpsWarrior)

			own := bare.GetAura("Battle Shout Rank 6 (Player)")
			if own == nil {
				t.Fatal("the warrior carries no aura labelled \"Battle Shout Rank 6 (Player)\"")
			}
			own.Activate(sim)
			if !own.IsActive() || own.ExclusiveEffects[0].Priority != withSet {
				t.Fatalf("the shared copy is active %v at %v, want it up at the set warrior's %v",
					own.IsActive(), own.ExclusiveEffects[0].Priority, withSet)
			}

			sim.CurrentTime = own.Duration - row.remaining
			if got := own.RemainingDuration(sim); got != row.remaining {
				t.Fatalf("the shared copy has %v left, want %v", got, row.remaining)
			}

			if got := battleShoutSpell(t, bare, battleShoutTrainerRank).ExtraCastCondition(sim, &bare.Unit); got != row.want {
				t.Errorf("the warrior would cast: %v, want %v with %v left and a %v threshold",
					got, row.want, row.remaining, warrior.ShoutExpirationThreshold)
			}
		})
	}
}

// A warrior that never shouts shares the raid's copy of the rank the raid's state names, which keeps
// the build phase the raid gave it so the character sheet shows it. The warrior registers the other
// rank's copy itself, and that one stays off the sheet.
func TestRaidShoutKeepsItsBuildPhaseOnAWarriorThatNeverShouts(t *testing.T) {
	const rank6, rank7 = "Battle Shout Rank 6 (External)", "Battle Shout Rank 7 (External)"
	for _, row := range []struct {
		state     int32
		raidLabel string
		ownLabel  string
	}{
		{1, rank6, rank7},
		{2, rank7, rank6},
		{3, rank7, rank6},
	} {
		player := &proto.Player{
			Name:          "Silent",
			Race:          proto.Race_RaceOrc,
			Class:         proto.Class_ClassWarrior,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: furyTalents,
			Spec: &proto.Player_DpsWarrior{
				DpsWarrior: &proto.DpsWarrior{
					Options: &proto.DpsWarrior_Options{
						ClassOptions: &proto.WarriorOptions{DefaultStance: proto.WarriorStance_WarriorStanceBerserker},
					},
				},
			},
		}
		env, _, _ := core.NewEnvironment(
			core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{BattleShout: row.state}, &proto.Debuffs{}),
			core.MakeSingleTargetEncounter(0), false, true)
		war := env.Raid.Parties[0].Players[0].(*dps.DpsWarrior)

		for _, want := range []struct {
			label string
			phase core.CharacterBuildPhase
		}{
			{row.raidLabel, core.CharacterBuildPhaseBuffs},
			{row.ownLabel, core.CharacterBuildPhaseNone},
		} {
			aura := war.GetAura(want.label)
			if aura == nil {
				t.Fatalf("state %d: the warrior carries no aura labelled %q", row.state, want.label)
			}
			if aura.BuildPhase != want.phase {
				t.Errorf("state %d: %q is in build phase %v, want %v", row.state, want.label, aura.BuildPhase, want.phase)
			}
		}
	}
}

// Sunder Armor is one debuff on the target: with a warrior in the raid, the raid's Sunder is the
// player's copy, which the warrior's cast stacks and the rotation reads by the spell's id, up to five,
// and the armor counts once.
func TestWarriorSunderStacksTheRaidsSunder(t *testing.T) {
	player := &proto.Player{
		Name:          "Sunder",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: furyTalents,
		Spec:          dpsOptions,
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{},
			&proto.Debuffs{SunderArmor: true}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	war := sim.Raid.Parties[0].Players[0].(*dps.DpsWarrior)
	target := sim.Encounter.AllTargetUnits[0]
	sunder := war.GetSpell(core.ActionID{SpellID: 11597})

	shared := target.GetAuraByID(core.ActionID{SpellID: 11597})
	if shared == nil {
		t.Fatalf("the rotation finds no Sunder Armor by its id on the target")
	}
	if shared.Label != "Sunder Armor (Player)" {
		t.Errorf("the raid's Sunder is labelled %q, want the player's copy %q", shared.Label, "Sunder Armor (Player)")
	}
	if own := war.SunderArmorAuras.Get(target); own != shared {
		t.Fatalf("the warrior stacks %q, want the raid's %q", own.Label, shared.Label)
	}

	armor := target.GetStats()[stats.Armor]
	shared.Activate(sim)
	shared.SetStacks(sim, 2)

	for range 50 {
		sunder.SkipCastAndApplyEffects(sim, target)
	}

	if got := shared.GetStacks(); got != 5 {
		t.Errorf("the raid's Sunder is at %d stacks after the warrior's casts, want 5", got)
	}
	if got, want := target.GetStats()[stats.Armor], armor-5*450; got != want {
		t.Errorf("armor is %v, want %v: five stacks of -450, counted once", got, want)
	}
	var sunders int
	for _, aura := range target.GetAurasWithTag(buffs.SunderArmorCategory) {
		if aura.ActionID.SpellID == 11597 {
			sunders++
		}
	}
	if sunders != 1 {
		t.Errorf("the target carries %d Sunder Armor auras, want the one both share", sunders)
	}
}

// A Retaliation counterattack is a main-hand special, so it spends no Flurry charge: a crit sets
// the charges to three, and anything else leaves the two the buff holds.
func TestRetaliationCounterattackSpendsNoFlurryCharge(t *testing.T) {
	sim, warrior := orcWarriorSim(t, "p1_fury", furyTalents)
	counterattack := warrior.GetSpell(core.ActionID{SpellID: 20240})
	flurry := warrior.GetAura("Flurry")
	if counterattack == nil || flurry == nil {
		t.Fatalf("counterattack %v, Flurry %v; want both registered", counterattack, flurry)
	}
	warrior.AddRage(sim, 100, warrior.NewRageMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionRageGain}))

	for range 20 {
		flurry.Activate(sim)
		flurry.SetStacks(sim, 2)
		if !counterattack.Cast(sim, warrior.CurrentTarget) {
			t.Fatal("the counterattack was not cast")
		}
		if got := flurry.GetStacks(); got == 1 {
			t.Fatalf("a counterattack left Flurry at %d charges, want 2, or 3 on a crit", got)
		}
	}
}

// Every white swing uses a Flurry charge whatever its outcome, and a Heroic Strike or Cleave uses
// one only when it misses.
func TestFlurryChargesFollowTheSwingOutcome(t *testing.T) {
	mainHand := core.ActionID{OtherID: proto.OtherAction_OtherActionAttack}.WithHand(proto.ItemSlot_ItemSlotMainHand)
	offHand := core.ActionID{OtherID: proto.OtherAction_OtherActionAttack}.WithHand(proto.ItemSlot_ItemSlotOffHand)
	heroicStrike := core.ActionID{SpellID: 11567}
	cleave := core.ActionID{SpellID: 20569}
	cases := []struct {
		name    string
		spell   core.ActionID
		outcome core.HitOutcome
		start   int32
		want    int32
	}{
		{"a missed main-hand swing uses a charge", mainHand, core.OutcomeMiss, 3, 2},
		{"a dodged main-hand swing uses a charge", mainHand, core.OutcomeDodge, 3, 2},
		{"a parried main-hand swing uses a charge", mainHand, core.OutcomeParry, 3, 2},
		{"a missed off-hand swing uses a charge", offHand, core.OutcomeMiss, 3, 2},
		{"a dodged off-hand swing uses a charge", offHand, core.OutcomeDodge, 3, 2},
		{"a parried off-hand swing uses a charge", offHand, core.OutcomeParry, 3, 2},
		{"a missed Heroic Strike uses a charge", heroicStrike, core.OutcomeMiss, 3, 2},
		{"a dodged Heroic Strike uses none", heroicStrike, core.OutcomeDodge, 3, 3},
		{"a parried Heroic Strike uses none", heroicStrike, core.OutcomeParry, 3, 3},
		{"a landed Heroic Strike uses none", heroicStrike, core.OutcomeHit, 3, 3},
		{"a missed Cleave uses a charge", cleave, core.OutcomeMiss, 3, 2},
		{"a dodged Cleave uses none", cleave, core.OutcomeDodge, 3, 3},
		{"a missed swing does not start Flurry", mainHand, core.OutcomeMiss, 0, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, warrior := orcWarriorSim(t, "p1_fury", furyTalents)
			flurry := warrior.GetAura("Flurry")
			spell := warrior.GetSpell(c.spell)
			if flurry == nil || spell == nil {
				t.Fatalf("Flurry %v, spell %v; want both registered", flurry, spell)
			}
			if c.start > 0 {
				flurry.Activate(sim)
				flurry.SetStacks(sim, c.start)
			}

			warrior.OnSpellHitDealt(sim, spell, &core.SpellResult{Target: warrior.CurrentTarget, Outcome: c.outcome})

			if got := flurry.GetStacks(); got != c.want {
				t.Errorf("Flurry holds %d charges, want %d", got, c.want)
			}
		})
	}
}

// Death Wish takes the global cooldown its row states and does not hold the rotation beyond it, so
// an off-GCD ability can follow it at once.
func TestDeathWishLeavesTheRotationFree(t *testing.T) {
	sim, warrior := orcWarriorSim(t, "p1_fury", furyTalents)
	deathWish := warrior.GetSpell(core.ActionID{SpellID: 12328})
	if deathWish == nil {
		t.Fatal("the warrior has no Death Wish")
	}
	if deathWish.DefaultCast.GCD != core.GCDDefault {
		t.Errorf("Death Wish takes a %v global cooldown, want %v", deathWish.DefaultCast.GCD, core.GCDDefault)
	}

	before := warrior.NextRotationActionAt()
	deathWish.SkipCastAndApplyEffects(sim, &warrior.Unit)
	if got := warrior.NextRotationActionAt(); got != before {
		t.Errorf("the rotation is held until %v after Death Wish, want %v", got, before)
	}
}

// Offensive State 1282735 hears the warrior's melee swings and abilities, and Defensive State 5301
// hears melee and ranged hits taken: a dodged swing opens the Overpower window and an avoided one the
// Revenge window, and a dodged spell or a landed swing opens neither.
func TestOverpowerAndRevengeWindowsOpenOnAvoidedMeleeHits(t *testing.T) {
	cases := []struct {
		name     string
		revenge  bool
		procMask core.ProcMask
		outcome  core.HitOutcome
		opens    bool
	}{
		{"a dodged swing opens the Overpower window", false, core.ProcMaskMeleeMHAuto, core.OutcomeDodge, true},
		{"a dodged spell does not open the Overpower window", false, core.ProcMaskSpellDamage, core.OutcomeDodge, false},
		{"a dodged swing opens the Revenge window", true, core.ProcMaskMeleeMHAuto, core.OutcomeDodge, true},
		{"a dodged spell does not open the Revenge window", true, core.ProcMaskSpellDamage, core.OutcomeDodge, false},
		{"a landed swing does not open the Revenge window", true, core.ProcMaskMeleeMHAuto, core.OutcomeHit, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, war, target := newWindowSim(t)

			window := war.OverpowerAura
			if c.revenge {
				window = war.GetAuraByID(core.ActionID{SpellID: 5302})
				if window == nil {
					t.Fatal("the warrior has no Revenge window")
				}
				war.OnSpellHitTaken(sim, &core.Spell{ProcMask: c.procMask, Unit: target},
					&core.SpellResult{Target: &war.Unit, Outcome: c.outcome})
			} else {
				war.OnSpellHitDealt(sim, &core.Spell{ProcMask: c.procMask, Unit: &war.Unit},
					&core.SpellResult{Target: target, Outcome: c.outcome})
			}

			if got := window.IsActive(); got != c.opens {
				t.Errorf("the window is up: %v, want %v", got, c.opens)
			}
		})
	}
}

// Furious Precision 3/3 gives the off-hand axe's attacks 10% hit and the main hand's none.
// Cruelty 5/5 gives each axe 5% crit, and a two-handed axe also takes Weaponmaster's 5%, whose aura is
// up.
func TestTalentsReachTheirWeapons(t *testing.T) {
	for _, c := range []struct {
		name     string
		gearSet  string
		talents  string
		mhCrit   float64
		ohHit    float64
		ohCrit   float64
		hasOH    bool
		wantAura string
	}{
		{"Fury dual-wields one-handed axes", "p1_fury", furyTalents, 5, 10, 5, true, ""},
		{"Arms swings a two-handed axe", "p1_arms", armsTalents, 10, 0, 0, false, "Weaponmaster (Axe/Polearm)"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, war := orcWarriorSim(t, c.gearSet, c.talents)

			mh := war.AutoAttacks.MHAuto()
			if mh.BonusHitPercent != 0 || mh.BonusCritPercent != c.mhCrit {
				t.Errorf("main-hand hit and crit = %0.2f%% and %0.2f%%, want 0%% and %0.2f%%", mh.BonusHitPercent, mh.BonusCritPercent, c.mhCrit)
			}
			if c.hasOH {
				oh := war.AutoAttacks.OHAuto()
				if oh.BonusHitPercent != c.ohHit || oh.BonusCritPercent != c.ohCrit {
					t.Errorf("off-hand hit and crit = %0.2f%% and %0.2f%%, want %0.2f%% and %0.2f%%", oh.BonusHitPercent, oh.BonusCritPercent, c.ohHit, c.ohCrit)
				}
			}
			if c.wantAura != "" {
				if aura := war.GetAura(c.wantAura); aura == nil || !aura.IsActive() {
					t.Errorf("%s is not up", c.wantAura)
				}
			}
		})
	}
}

// A shield's hit is an off-hand hit: it never procs Windfury Totem or the main-hand weapon's
// enchant, which a main-hand special does.
func TestShieldHitsNeverProcMainHandWeaponEffects(t *testing.T) {
	sim, war, target := shieldTestSim(t)
	mainHandSpecial := &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial}

	crusaderTrigger := war.GetAura("Enchant Weapon - Crusader")
	if crusaderTrigger == nil || crusaderTrigger.Dpm == nil {
		t.Fatal("the main hand's Crusader registered no proc manager")
	}
	rollsCrusader := func(procMask core.ProcMask) bool {
		for range 1000 {
			if crusaderTrigger.Dpm.Proc(sim, procMask, "Crusader Test") {
				return true
			}
		}
		return false
	}
	if !rollsCrusader(mainHandSpecial.ProcMask) {
		t.Fatal("1000 main-hand specials never procced Crusader, so the check below proves nothing")
	}

	windfuryTrigger := war.GetAura("Windfury Totem Trigger")
	windfury := war.GetAura("Windfury Totem (External)")
	if windfuryTrigger == nil || windfury == nil {
		t.Fatal("Windfury Totem registered no trigger or proc")
	}

	for name, spell := range shieldSpells(t, war) {
		if rollsCrusader(spell.ProcMask) {
			t.Errorf("%s procced the main hand's Crusader", name)
		}
		for i := 0; i < 200 && !windfury.IsActive(); i++ {
			windfuryTrigger.OnCastResult(windfuryTrigger, sim, spell, &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
		}
		if windfury.IsActive() {
			t.Errorf("%s procced Windfury Totem", name)
		}
	}

	for i := 0; i < 200 && !windfury.IsActive(); i++ {
		windfuryTrigger.OnCastResult(windfuryTrigger, sim, mainHandSpecial, &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
	}
	if !windfury.IsActive() {
		t.Error("200 main-hand specials never procced Windfury Totem, so the check above proves nothing")
	}
}

// Heroic Strike goes off on the next swing, which casts it directly, so the press that queues it is the
// manual cast. A press whose swing never fires still counts.
func TestHeroicStrikeCountsItsQueuingPressAsAManualCast(t *testing.T) {
	player := &proto.Player{
		Name:          "Press Count",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Equipment:     core.GetGearSet("../../ui/specs/warrior/dps/gear_sets", "p1_fury").GearSet,
		TalentsString: furyTalents,
		Spec:          dpsOptions,
		Rotation:      core.GetAplRotation("../../ui/specs/warrior/dps/apls", "fury").Rotation,
	}
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{Iterations: 5, RandomSeed: 1},
	})

	const heroicStrike = 25286 // The class book's rank, the one the Fury preset presses.
	for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
		if action.Id.GetSpellId() != heroicStrike || action.Id.Tag != 0 {
			continue
		}
		var presses, swings int32
		for _, target := range action.Targets {
			presses += target.ManualCasts
			swings += target.Casts
		}
		if swings == 0 || presses < swings {
			t.Errorf("Heroic Strike: ManualCasts = %d, Casts = %d, want at least as many presses as swings above 0", presses, swings)
		}
		return
	}
	t.Fatal("the result has no Heroic Strike action")
}

// An engineer warrior at 100 rage with no swing running, its Cleave queue spell and aura, and the
// explosive the consumables register.
func newThrowerSim(t *testing.T, consumes *proto.ConsumesSpec, explosiveID int32, queueDelay int32) (sim *core.Simulation, war *dps.DpsWarrior, queueCleave *core.Spell, queue *core.Aura, explosive *core.Spell) {
	t.Helper()
	player := &proto.Player{
		Name:          "Thrower",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Equipment:     core.GetGearSet("../../ui/specs/warrior/dps/gear_sets", "p1_fury").GearSet,
		TalentsString: furyTalents,
		Spec: &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{
			ClassOptions: &proto.WarriorOptions{
				UseBattleShout: true,
				DefaultStance:  proto.WarriorStance_WarriorStanceBerserker,
				QueueDelay:     queueDelay,
			},
		}}},
		Profession1: proto.Profession_Engineering,
		Consumables: consumes,
		Rotation:    &proto.APLRotation{},
	}
	sim = core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	war = sim.Raid.Parties[0].Players[0].(*dps.DpsWarrior)
	war.AutoAttacks.CancelAutoSwing(sim)
	war.AddRage(sim, 100, war.NewRageMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionRageGain}))

	queueCleave = war.GetSpell(war.Cleave.ActionID.WithTag(1))
	queue = war.GetAura("HS/Cleave Queue Aura-" + war.Cleave.ActionID.String())
	explosive = war.GetSpell(core.ActionID{ItemID: explosiveID})
	if queueCleave == nil || queue == nil || explosive == nil {
		t.Fatalf("queue spell %v, queue aura %v, explosive %v; want all registered", queueCleave, queue, explosive)
	}
	return sim, war, queueCleave, queue, explosive
}

// Throwing an explosive with a cast drops a Heroic Strike or Cleave queued for the next swing; the
// instant sapper charge keeps it.
func TestExplosiveDropsTheQueuedCleave(t *testing.T) {
	for _, tc := range []struct {
		name       string
		itemID     int32
		consumes   *proto.ConsumesSpec
		wantQueued bool
	}{
		{"Ez-Thro Dynamite II drops it", 18588, &proto.ConsumesSpec{ExplosiveId: 18588}, false},
		{"Goblin Sapper Charge keeps it", 10646, &proto.ConsumesSpec{ExtraIds: []int32{10646}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, war, queueCleave, queue, explosive := newThrowerSim(t, tc.consumes, tc.itemID, 0)
			if !queueCleave.Cast(sim, war.CurrentTarget) || !queue.IsActive() {
				t.Fatal("Cleave could not be queued")
			}
			if !explosive.Cast(sim, war.CurrentTarget) {
				t.Fatal("the explosive could not be thrown")
			}
			if got := queue.IsActive(); got != tc.wantQueued {
				t.Errorf("Cleave queued after the throw = %v, want %v", got, tc.wantQueued)
			}
		})
	}
}

// A Cleave press still waiting out the Queue Delay goes with the throw too, and Cleave can be queued
// again once the throw is done.
func TestExplosiveDropsTheCleavePressInsideTheQueueDelay(t *testing.T) {
	sim, war, queueCleave, queue, explosive := newThrowerSim(t, &proto.ConsumesSpec{ExplosiveId: 18588}, 18588, 250)
	if !queueCleave.Cast(sim, war.CurrentTarget) {
		t.Fatal("Cleave could not be pressed")
	}
	if !explosive.Cast(sim, war.CurrentTarget) {
		t.Fatal("the explosive could not be thrown")
	}
	for sim.CurrentTime < 2*time.Second {
		sim.Step()
	}
	if queue.IsActive() {
		t.Error("Cleave queued during the throw, want the press inside the Queue Delay dropped")
	}
	if !queueCleave.Cast(sim, war.CurrentTarget) {
		t.Error("Cleave could not be queued again after the throw")
	}
}
