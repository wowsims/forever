package retribution

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/paladin"
)

// The swing the tests land, and the time they let pass after one: the batch window the seal effects
// land in and the talent's proc cooldown, so the next swing is heard.
var unsealedWrathSwing = core.ActionID{OtherID: 1}

func passUnsealedWrathCooldown(sim *core.Simulation) {
	for end := sim.CurrentTime + spelldata.MustFind(1310735).ICD() + 2*core.SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
	}
}

// A paladin with Unsealed Wrath and no rotation or auto attacks, a main-hand weapon to swap to, a
// swing that always lands, and a count of the Seal of Righteousness hits cast so far under one tag.
func unsealedWrathPaladin(t *testing.T) (sim *core.Simulation, pal *paladin.Paladin, swing func(), righteousnessHits func(tag int32) int32) {
	t.Helper()
	return unsealedWrathPaladinUnder(t, &proto.RaidBuffs{})
}

func unsealedWrathPaladinUnder(t *testing.T, raidBuffs *proto.RaidBuffs) (sim *core.Simulation, pal *paladin.Paladin, swing func(), righteousnessHits func(tag int32) int32) {
	t.Helper()
	player := &proto.Player{
		Name:           "Unsealed Wrath",
		Race:           proto.Race_RaceHuman,
		Class:          proto.Class_ClassPaladin,
		Equipment:      core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "default").GearSet,
		TalentsString:  UnsealedWrathTalents,
		EnableItemSwap: true,
		ItemSwap:       core.GetItemSwapGearSet("../../../ui/specs/paladin/retribution/gear_sets", "holy_shock_item_swap").ItemSwap,
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
		}},
		Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	sim = core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, raidBuffs, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pal = sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
	pal.AutoAttacks.CancelAutoSwing(sim)
	if !pal.Talents.UnsealedWrath {
		t.Fatal("the talents do not take Unsealed Wrath")
	}

	righteousnessHits = func(tag int32) int32 {
		var casts int32
		for _, spell := range pal.Spellbook {
			if spell.Matches(paladin.SpellMaskSealOfRighteousnessProc) && spell.ActionID.Tag == tag {
				casts += spell.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts
			}
		}
		return casts
	}

	strike := pal.RegisterSpell(core.SpellConfig{
		ActionID:         unsealedWrathSwing,
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         core.ProcMaskMeleeMHAuto,
		Flags:            core.SpellFlagMeleeMetrics,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		},
	})
	swing = func() {
		strike.SkipCastAndApplyEffects(sim, pal.CurrentTarget)
		passUnsealedWrathCooldown(sim)
	}
	return sim, pal, swing, righteousnessHits
}

func castSeal(t *testing.T, sim *core.Simulation, pal *paladin.Paladin, id int32) {
	t.Helper()
	spell := pal.GetSpell(core.ActionID{SpellID: id})
	if spell == nil {
		t.Fatalf("seal %d is not registered", id)
	}
	spell.SkipCastAndApplyEffects(sim, pal.CurrentTarget)
}

const windfuryTotemStrike = 10610 // The strike Windfury Totem's proc casts.

// Every seal with a per-hit effect hears Windfury Totem's strike as one more white swing.
func TestSealsHearWindfuryTotemsStrike(t *testing.T) {
	for _, c := range []struct {
		name string
		seal int32
		proc int64
	}{
		{"Seal of Command", paladin.SealOfCommandRankMap.Highest().ID, paladin.SpellMaskSealOfCommandProc},
		{"Seal of Fury", paladin.SealOfFuryRankMap.Highest().ID, paladin.SpellMaskSealOfFuryProc},
		{"Seal of Righteousness", paladin.SealOfRighteousnessRankMap.Highest().ID, paladin.SpellMaskSealOfRighteousnessProc},
		{"Seal of Wisdom", paladin.SealOfWisdomRankMap.Highest().ID, paladin.SpellMaskSealOfWisdomProc},
		{"Seal of Light", paladin.SealOfLightRankMap.Highest().ID, paladin.SpellMaskSealOfLightProc},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, pal, _, _ := unsealedWrathPaladinUnder(t, &proto.RaidBuffs{WindfuryTotem: true})
			strike := pal.GetSpell(core.ActionID{SpellID: windfuryTotemStrike})
			if strike == nil {
				t.Fatal("the paladin has no Windfury Totem strike")
			}
			procs := func() int32 {
				var casts int32
				for _, spell := range pal.Spellbook {
					if spell.Matches(c.proc) && spell.ActionID.Tag == 0 {
						casts += spell.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts + spell.SpellMetrics[pal.UnitIndex].Casts
					}
				}
				return casts
			}

			// Each strike comes after the seal's own proc cooldown, under a fresh seal.
			for range 100 {
				castSeal(t, sim, pal, c.seal)
				pal.OnSpellHitDealt(sim, strike, &core.SpellResult{Target: pal.CurrentTarget, Outcome: core.OutcomeHit, Damage: 100})
				for end := sim.CurrentTime + 1100*time.Millisecond; sim.CurrentTime <= end && !sim.Step(); {
				}
			}
			if procs() == 0 {
				t.Errorf("100 landed Windfury Totem strikes under %s procced it none, want some", c.name)
			}
		})
	}
}

// Windfury Totem's strike neither stores nor spends an Echo, though the seals hear it as a white
// swing: Unsealed Wrath hears white swings only. Each landed swing stores one copy whether or not a
// strike follows.
func TestUnsealedWrathStoresNoCopyFromWindfuryTotemsStrike(t *testing.T) {
	sim, pal, swing, _ := unsealedWrathPaladinUnder(t, &proto.RaidBuffs{WindfuryTotem: true})
	strike := pal.GetSpell(core.ActionID{SpellID: windfuryTotemStrike})
	echo := pal.GetAuraByID(core.ActionID{SpellID: 1311704})
	if strike == nil || echo == nil {
		t.Fatal("the paladin has no Windfury Totem strike or no Echo of Righteousness")
	}
	var struck int
	for range 300 {
		castSeal(t, sim, pal, paladin.SealOfRighteousnessRankMap.Highest().ID)
		echo.Deactivate(sim)
		before := strike.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts
		swing()
		if strike.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts > before {
			struck++
		}
		if got := echo.GetStacks(); got != 1 {
			t.Fatalf("a swing stored %d copies of the seal, want 1", got)
		}
	}
	if struck == 0 {
		t.Fatal("300 swings brought no Windfury Totem strike, so the check above proves nothing")
	}
}

// Each swing under a seal stores one copy of it, up to the Echo's 3 stacks.
func TestUnsealedWrathStoresUpToThreeCopies(t *testing.T) {
	sim, pal, swing, _ := unsealedWrathPaladin(t)
	castSeal(t, sim, pal, paladin.SealOfRighteousnessRankMap.Highest().ID)

	echo := pal.GetAuraByID(core.ActionID{SpellID: 1311704})
	if echo == nil {
		t.Fatal("Echo of Righteousness is not registered")
	}
	for i, want := range []int32{1, 2, 3, 3} {
		swing()
		if got := echo.GetStacks(); got != want {
			t.Fatalf("after swing %d Echo of Righteousness has %d stacks, want %d", i+1, got, want)
		}
	}
}

// A second swing in the instant of the first stores no copy, which is how an extra attack adds none;
// a swing after the talent's proc cooldown stores one again.
func TestUnsealedWrathStoresNoCopyFromASwingInsideItsProcCooldown(t *testing.T) {
	sim, pal, swing, _ := unsealedWrathPaladin(t)
	castSeal(t, sim, pal, paladin.SealOfRighteousnessRankMap.Highest().ID)
	echo := pal.GetAuraByID(core.ActionID{SpellID: 1311704})
	strike := pal.GetSpell(unsealedWrathSwing)

	strike.SkipCastAndApplyEffects(sim, pal.CurrentTarget)
	strike.SkipCastAndApplyEffects(sim, pal.CurrentTarget)
	if got := echo.GetStacks(); got != 1 {
		t.Fatalf("two swings in one instant stored %d copies of the seal, want 1", got)
	}

	passUnsealedWrathCooldown(sim)
	swing()
	if got := echo.GetStacks(); got != 2 {
		t.Errorf("a swing after the proc cooldown left %d copies of the seal, want 2", got)
	}
}

// A hit that lands under another seal spends every stored copy, each one a Seal of Righteousness hit
// tagged with Echo of Righteousness, and stores a copy of the new seal.
func TestUnsealedWrathSpendsTheStoredCopiesUnderAnotherSeal(t *testing.T) {
	sim, pal, swing, righteousnessHits := unsealedWrathPaladin(t)
	castSeal(t, sim, pal, paladin.SealOfRighteousnessRankMap.Highest().ID)
	for range 3 {
		swing()
	}

	castSeal(t, sim, pal, paladin.SealOfFuryRankMap.Highest().ID)
	before, beforeUntagged := righteousnessHits(1311704), righteousnessHits(0)
	swing()

	if got := righteousnessHits(1311704) - before; got != 3 {
		t.Errorf("the hit under Seal of Fury replayed %d Seal of Righteousness hits tagged with Echo of Righteousness, want 3", got)
	}
	if got := righteousnessHits(0) - beforeUntagged; got != 0 {
		t.Errorf("the hit under Seal of Fury cast %d untagged Seal of Righteousness hits, want 0", got)
	}
	if pal.GetAuraByID(core.ActionID{SpellID: 1311704}).IsActive() {
		t.Error("Echo of Righteousness stays up after its copies are spent")
	}
	if got := pal.GetAuraByID(core.ActionID{SpellID: 1311701}).GetStacks(); got != 1 {
		t.Errorf("Echo of Fury has %d stacks, want 1", got)
	}
}

// Swapping the main-hand weapon drops every Echo, so a hit under another seal afterwards spends no
// stored copy.
func TestUnsealedWrathDropsTheEchoesOnAMainHandSwap(t *testing.T) {
	sim, pal, swing, righteousnessHits := unsealedWrathPaladin(t)
	castSeal(t, sim, pal, paladin.SealOfRighteousnessRankMap.Highest().ID)
	for range 3 {
		swing()
	}
	echo := pal.GetAuraByID(core.ActionID{SpellID: 1311704})
	if echo.GetStacks() != 3 {
		t.Fatalf("Echo of Righteousness has %d stacks before the swap, want 3", echo.GetStacks())
	}

	mainHand := pal.MainHand().ID
	pal.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
	if pal.MainHand().ID == mainHand {
		t.Fatalf("the swap left main-hand weapon %d in place", mainHand)
	}
	if echo.IsActive() {
		t.Error("Echo of Righteousness stays up through a main-hand swap")
	}

	castSeal(t, sim, pal, paladin.SealOfFuryRankMap.Highest().ID)
	before := righteousnessHits(1311704)
	swing()
	if got := righteousnessHits(1311704) - before; got != 0 {
		t.Errorf("the hit under Seal of Fury after the swap replayed %d Seal of Righteousness hits, want 0", got)
	}
}

// Seal of Light and Seal of Wisdom are stored as well, and each spent copy of Seal of Wisdom rolls the
// seal's own chance to restore mana.
func TestUnsealedWrathStoresLightAndWisdom(t *testing.T) {
	sim, pal, swing, _ := unsealedWrathPaladin(t)
	echoOfLight := pal.GetAuraByID(core.ActionID{SpellID: 1328198})
	echoOfWisdom := pal.GetAuraByID(core.ActionID{SpellID: 1328206})
	if echoOfLight == nil || echoOfWisdom == nil {
		t.Fatal("Echo of Light or Echo of Wisdom is not registered")
	}

	castSeal(t, sim, pal, paladin.SealOfLightRankMap.Highest().ID)
	swing()
	if got := echoOfLight.GetStacks(); got != 1 {
		t.Fatalf("a swing under Seal of Light left Echo of Light at %d stacks, want 1", got)
	}

	castSeal(t, sim, pal, paladin.SealOfWisdomRankMap.Highest().ID)
	for range 3 {
		swing()
	}
	if echoOfLight.IsActive() {
		t.Error("Echo of Light stays up after a hit under Seal of Wisdom")
	}
	if got := echoOfWisdom.GetStacks(); got != 3 {
		t.Fatalf("three swings under Seal of Wisdom left Echo of Wisdom at %d stacks, want 3", got)
	}

	wisdomRestores := func() int32 {
		var casts int32
		for _, spell := range pal.Spellbook {
			if spell.Matches(paladin.SpellMaskSealOfWisdomProc) {
				for _, metrics := range spell.SpellMetrics {
					casts += metrics.Casts
				}
			}
		}
		return casts
	}
	// Only the hit under Seal of Righteousness is counted, so every restore is a spent copy's.
	const spends = 100
	var restores int32
	for i := range spends {
		if i > 0 {
			castSeal(t, sim, pal, paladin.SealOfWisdomRankMap.Highest().ID)
			for range 3 {
				swing()
			}
		}
		castSeal(t, sim, pal, paladin.SealOfRighteousnessRankMap.Highest().ID)
		before := wisdomRestores()
		swing()
		restores += wisdomRestores() - before
	}
	if restores == 0 || restores == 3*spends {
		t.Errorf("%d spends of three Seal of Wisdom copies restored mana %d times, want a roll per copy", spends, restores)
	}
}

// The ladder's first Seal of Righteousness deals its hits as its own spell, not as the level 1
// starter seal's.
func TestSealOfRighteousnessRankOneHitsAsItsOwnSpell(t *testing.T) {
	const want = 25741 // 21084's damage spell, "Rank 1.5" in the game files
	sim, pal, swing, _ := unsealedWrathPaladin(t)
	rank := paladin.SealOfRighteousnessRankMap.Rank(1)
	castSeal(t, sim, pal, rank.ID)
	for range 20 {
		swing()
	}

	hitsBy := map[int32]int32{}
	for _, spell := range pal.Spellbook {
		if spell.Matches(paladin.SpellMaskSealOfRighteousnessProc) {
			if casts := spell.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts; casts > 0 {
				hitsBy[spell.ActionID.SpellID] += casts
			}
		}
	}
	if len(hitsBy) != 1 || hitsBy[want] == 0 {
		t.Errorf("seal %d hits as %v (spell id: hits), want spell %d alone", rank.ID, hitsBy, want)
	}
}
