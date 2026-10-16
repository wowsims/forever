package warrior_test

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
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

// The warrior's own shout with three pieces of Battlegear of Wrath on, through
// the class option: it raises what the shout bids for its category by the 30
// spell 23563 states.
func TestWarriorShoutsForTheTierTwoBonus(t *testing.T) {
	bidOf := func(hasT2 bool) float64 {
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
		aura := war.GetAura("Battle Shout (Player)")
		if aura == nil {
			t.Fatal("the ally aura array holds no aura labelled \"Battle Shout (Player)\"")
		}
		return aura.ExclusiveEffects[0].Priority
	}

	bare, withSet := bidOf(false), bidOf(true)

	if want := buffs.BattleShoutValue(0); bare != want {
		t.Errorf("a warrior without the set bids %v, want the game files' %v", bare, want)
	}
	if want := bare + buffs.BattleShoutT2Bonus; withSet != want {
		t.Errorf("a warrior wearing the set bids %v, want %v", withSet, want)
	}
}

// How the warrior treats the Battle Shout category. The buff belongs to the
// unit, so a 169 shout from the raid and one from the warrior standing next to
// this one fill the same category. Casting a 139 shout the category would turn
// away wastes a global and rage, and so does refreshing a copy with most of its
// three minutes left.
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

	t.Run("an improved external shout outbids a warrior without the set", func(t *testing.T) {
		sim := &core.Simulation{}
		env := build(&proto.RaidBuffs{BattleShout: proto.TristateEffect_TristateEffectImproved}, warriorProto("Bare", false))
		war := env.Raid.Parties[0].Players[0].(*dps.DpsWarrior)

		external := war.GetAura("Battle Shout (External)")
		if external == nil {
			t.Fatal("the raid's copy is not registered")
		}
		external.Activate(sim)
		if !external.IsActive() || external.ExclusiveEffects[0].Priority != withSet {
			t.Fatalf("the raid's copy is active %v at %v, want it up at %v",
				external.IsActive(), external.ExclusiveEffects[0].Priority, withSet)
		}

		if war.BattleShout.ExtraCastCondition(sim, &war.Unit) {
			t.Error("the warrior would shout for 139 while a 169 copy holds the category, wasting the global")
		}
	})

	// The copy holding the category is the warrior's own, shared with a raid
	// member wearing the set, so it bids 169 while this warrior is worth 139.
	// What decides the cast is how long it has left, not what it bids.
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

			own := bare.GetAura("Battle Shout (Player)")
			if own == nil {
				t.Fatal("the warrior carries no aura labelled \"Battle Shout (Player)\"")
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

			if got := bare.BattleShout.ExtraCastCondition(sim, &bare.Unit); got != row.want {
				t.Errorf("the warrior would cast: %v, want %v with %v left and a %v threshold",
					got, row.want, row.remaining, warrior.ShoutExpirationThreshold)
			}
		})
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
				window = war.GetAuraByID(core.ActionID{SpellID: 25288})
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
// up. Savage is crit on every attack, not a mod on the axes.
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
