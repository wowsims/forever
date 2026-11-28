package paladin_test

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/paladin"
	"github.com/wowsims/forever/sim/paladin/retribution"
)

// A neck no game files row holds, whose effect gives its wearer the listeners below.
const listenerNeck int32 = 9_990_001

// How often each kind of listener heard a main-hand melee ability, counted by the neck's wearer.
type tally struct {
	castResults        int // Every cast result, landed or not.
	weaponProc         int // A landed hit, as a weapon proc hears it.
	hitListener        int // A landed hit, as any other proc hears it.
	castResultListener int // A landed cast result, as any other proc hears it.
}

var heard tally

func init() {
	retribution.RegisterRetributionPaladin()
	common.RegisterAllEffects()

	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{
		Id: listenerNeck, Name: "Listener Neck", Type: proto.ItemType_ItemTypeNeck,
		ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}},
	}}})
	core.NewItemEffect(listenerNeck, func(agent core.Agent) {
		listen := func(name string, count *int, trigger core.ProcTrigger) {
			trigger.Name = name
			trigger.ProcMask = core.ProcMaskMeleeMHSpecial
			trigger.TriggerImmediately = true
			trigger.Handler = func(_ *core.Simulation, _ *core.Spell, _ *core.SpellResult) { *count++ }
			agent.GetCharacter().MakeProcTriggerAura(trigger)
		}
		listen("Every Cast Result", &heard.castResults, core.ProcTrigger{Callback: core.CallbackOnCastResult})
		listen("Weapon Proc Listener", &heard.weaponProc, core.ProcTrigger{Callback: core.CallbackOnSpellHitDealt, Outcome: core.OutcomeLanded, IsWeaponProc: true})
		listen("Hit Listener", &heard.hitListener, core.ProcTrigger{Callback: core.CallbackOnSpellHitDealt, Outcome: core.OutcomeLanded})
		listen("Cast Result Listener", &heard.castResultListener, core.ProcTrigger{Callback: core.CallbackOnCastResult, Outcome: core.OutcomeLanded})
	})
}

// The Retribution suite's Unsealed Wrath build.
var retributionTalents = "050250002-501-05233011201301351"

// A Retribution paladin who wears the listener neck and nothing else, with no rotation and no auto
// attacks, on one target.
func listeningPaladin(t *testing.T) (*core.Simulation, *paladin.Paladin, *core.Unit) {
	t.Helper()
	sim, pal, target := paladinWearing(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotNeck: listenerNeck})
	heard = tally{}
	return sim, pal, target
}

// A Retribution paladin who wears the items and nothing else, with no rotation and no auto attacks,
// on one target.
func paladinWearing(t *testing.T, worn map[proto.ItemSlot]int32) (*core.Simulation, *paladin.Paladin, *core.Unit) {
	t.Helper()
	return paladinWearingAgainst(t, worn, core.MakeSingleTargetEncounter(0), proto.Race_RaceHuman)
}

// The same paladin of the race against the encounter's targets.
func paladinWearingAgainst(t *testing.T, worn map[proto.ItemSlot]int32, encounter *proto.Encounter, race proto.Race) (*core.Simulation, *paladin.Paladin, *core.Unit) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for slot := range items {
		items[slot] = &proto.ItemSpec{Id: worn[proto.ItemSlot(slot)]}
	}
	player := &proto.Player{
		Name:          "Paladin",
		Race:          race,
		Class:         proto.Class_ClassPaladin,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: retributionTalents,
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
		}},
		Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pal := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
	pal.AutoAttacks.CancelAutoSwing(sim)
	return sim, pal, pal.CurrentTarget
}

// A landed Judgement of Command rolls a weapon proc once, on its damage (EV-0155). A listener that is
// no weapon proc hears the damage alone, and a judgement that misses rolls none of them (EV-0156).
func TestJudgementOfCommandRollsAWeaponProcOnce(t *testing.T) {
	const casts = 50
	for _, c := range []struct {
		name  string
		lands bool
		want  tally
	}{
		{"a judgement that lands", true, tally{castResults: casts, weaponProc: casts, hitListener: casts, castResultListener: casts}},
		{"a judgement that misses", false, tally{castResults: casts}},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, pal, target := listeningPaladin(t)
			seal := pal.GetSpell(core.ActionID{SpellID: paladin.SealOfCommandRankMap.Highest().ID})
			if seal == nil {
				t.Fatal("the paladin has no Seal of Command")
			}
			seal.SkipCastAndApplyEffects(sim, target)

			if c.lands {
				pal.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
			} else {
				for _, spell := range pal.Spellbook {
					if spell.Matches(paladin.SpellMaskJudgementOfCommand) {
						pal.AttackTables[target.UnitIndex].Hand(spell).BaseMissChance = 2
					}
				}
			}
			for range casts {
				pal.Judgement.SkipCastAndApplyEffects(sim, target)
			}

			// The dummy's landing is no cast result to an aura proc, so each cast shows once there.
			if heard != c.want {
				t.Errorf("%d Judgement casts were heard as %+v, want %+v", casts, heard, c.want)
			}
		})
	}
}

// Battlegear of Eternal Justice's 3-piece bonus restores the mana of its spell 26975 on its row's
// share of Judgement casts, once a cast at most.
func TestBattlegearOfEternalJusticeRestoresManaOnJudgement(t *testing.T) {
	gain := spelldata.MustFind(26975).EnergizeEffect().BaseValue()
	chance := spelldata.MustFind(26135).StatedChance()
	if gain <= 0 || chance <= 0 {
		t.Fatalf("the set restores %v mana at a chance of %v, want both above 0", gain, chance)
	}
	const casts = 1000

	// How many Judgement casts restored the mana once, and how many more than once.
	restores := func(t *testing.T, worn map[proto.ItemSlot]int32) (once, more int) {
		t.Helper()
		sim, pal, target := paladinWearing(t, worn)
		seal := pal.GetSpell(core.ActionID{SpellID: paladin.SealOfCommandRankMap.Highest().ID})
		if seal == nil {
			t.Fatal("the paladin has no Seal of Command")
		}
		metrics := pal.NewManaMetrics(core.ActionID{SpellID: 1})
		for range casts {
			seal.SkipCastAndApplyEffects(sim, target)
			pal.SpendMana(sim, pal.CurrentMana(), metrics)
			pal.AddMana(sim, pal.MaxMana()/2, metrics)
			pal.Judgement.CD.Reset()
			pal.SetGCDTimer(sim, sim.CurrentTime)
			if !pal.Judgement.Cast(sim, target) {
				t.Fatal("Judgement did not cast")
			}
			before := pal.CurrentMana()
			stepFor(t, sim, 5*core.SpellBatchWindow)
			switch regained := pal.CurrentMana() - before; {
			case regained >= 2*gain-1e-6:
				more++
			case regained >= gain-1e-6:
				once++
			}
		}
		return once, more
	}

	once, more := restores(t, map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotBack:     21397,
		proto.ItemSlot_ItemSlotFinger1:  21396,
		proto.ItemSlot_ItemSlotMainHand: 21395,
	})
	if got := float64(once) / casts; math.Abs(got-chance) > 0.05 {
		t.Errorf("with the set, %v of Judgement casts restored the mana, want about %v", got, chance)
	}
	if more != 0 {
		t.Errorf("with the set, %d Judgement casts restored the mana more than once, want none", more)
	}
	if once, more := restores(t, nil); once+more != 0 {
		t.Errorf("without the set, %d Judgement casts restored the mana, want none", once+more)
	}
}

// A white swing that crits adds its Vengeance stack after the hits it procs (EV-0224).
func TestSealOfRighteousnessHitIsPricedOnVengeanceBeforeTheCrit(t *testing.T) {
	// The damage of the Seal of Righteousness hit a white swing procs under Vengeance at the stacks
	// given. The seal's hit never crits, so it adds no stack of its own and the amounts compare.
	sealHit := func(t *testing.T, stacks int32, crit bool) float64 {
		t.Helper()
		sim, pal, target := paladinWearing(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: 21395})
		vengeance := pal.GetAuraByID(core.ActionID{SpellID: 20050})
		if vengeance == nil {
			t.Fatal("the paladin has no Vengeance")
		}
		seal := pal.GetSpell(core.ActionID{SpellID: paladin.SealOfRighteousnessRankMap.Highest().ID})
		if seal == nil {
			t.Fatal("the paladin has no Seal of Righteousness")
		}
		seal.SkipCastAndApplyEffects(sim, target)

		var sealHits []*core.Spell
		for _, spell := range pal.Spellbook {
			if spell.Matches(paladin.SpellMaskSealOfRighteousnessProc) && spell.ActionID.Tag == 0 {
				spell.BonusCritPercent = -1000
				sealHits = append(sealHits, spell)
			}
		}
		dealt := func() (total float64) {
			for _, spell := range sealHits {
				total += spell.SpellMetrics[target.UnitIndex].TotalDamage
			}
			return total
		}

		pal.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
		if crit {
			pal.AddStatDynamic(sim, stats.PhysicalCritPercent, 100)
		} else {
			pal.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		}

		swing := pal.AutoAttacks.MHAuto()
		for range 100 {
			vengeance.Deactivate(sim)
			if stacks > 0 {
				vengeance.Activate(sim)
				vengeance.SetStacks(sim, stacks)
			}
			before, crits := dealt(), swing.SpellMetrics[target.UnitIndex].Crits
			if !swing.Cast(sim, target) {
				t.Fatal("the white swing did not cast")
			}
			stepFor(t, sim, 5*core.SpellBatchWindow)

			hit := dealt() - before
			critted := swing.SpellMetrics[target.UnitIndex].Crits > crits
			if hit <= 0 || critted != crit {
				continue
			}
			if crit && vengeance.GetStacks() != stacks+1 {
				t.Fatalf("a crit from %d Vengeance stacks left %d, want %d", stacks, vengeance.GetStacks(), stacks+1)
			}
			return hit
		}
		t.Fatalf("no white swing in 100 procced a Seal of Righteousness hit with crit %v", crit)
		return 0
	}

	for _, stacks := range []int32{0, 3} {
		atStacks := sealHit(t, stacks, false)
		if above := sealHit(t, stacks+1, false); above <= atStacks {
			t.Fatalf("a Seal of Righteousness hit at %d Vengeance stacks deals %.3f, at %d %.3f, want more with the stack", stacks+1, above, stacks, atStacks)
		}
		if fromCrit := sealHit(t, stacks, true); math.Abs(fromCrit-atStacks) > 1e-6 {
			t.Errorf("a crit from %d Vengeance stacks procs a Seal of Righteousness hit of %.3f, want %.3f, the hit at %d stacks", stacks, fromCrit, atStacks, stacks)
		}
	}
}

// A Consecration cast rolls Draconic Infused Emblem once for each enemy, and its ticks roll none (EV-0115).
func TestConsecrationCastRollsDraconicInfusedEmblemPerEnemy(t *testing.T) {
	const casts, enemies = 2000, 5
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Duration = casts * 9
	for range enemies - 1 {
		encounter.Targets = append(encounter.Targets, core.NewDefaultTarget())
	}
	sim, pal, target := paladinWearingAgainst(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: 22268}, encounter, proto.Race_RaceHuman)
	buff := pal.GetAuraByID(core.ActionID{SpellID: 1318930})
	if buff == nil {
		t.Fatal("22268 registers no 1318930")
	}
	consecration := pal.GetSpell(core.ActionID{SpellID: paladin.ConsecrationRankMap.Highest().ID})
	if consecration == nil {
		t.Fatal("the paladin has no Consecration")
	}
	pal.AddStatDynamic(sim, stats.SpellHitPercent, 100)

	// The proc is deferred by the spell batch window, one pending action each, so each step deals at most
	// one, and the buff comes down before the next roll can refresh it.
	var atCast, onTicks int
	for range casts {
		start := sim.CurrentTime
		consecration.SkipCastAndApplyEffects(sim, target)
		for consecration.AOEDot().IsActive() || sim.CurrentTime < start+100*time.Millisecond {
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
		t.Errorf("%d Consecration casts on %d enemies procced it %d times at the cast, want about %.0f", casts, enemies, atCast, mean)
	}
	if onTicks != 0 {
		t.Errorf("Consecration ticks procced it %d times, want none", onTicks)
	}
}

// An Undead paladin's Consecration cast sets off no Touch of the Grave (EV-0115).
func TestConsecrationCastSetsOffNoTouchOfTheGrave(t *testing.T) {
	const casts = 100
	for _, enemies := range []int{1, 5} {
		encounter := core.MakeSingleTargetEncounter(0)
		encounter.Duration = casts
		for range enemies - 1 {
			encounter.Targets = append(encounter.Targets, core.NewDefaultTarget())
		}
		sim, pal, target := paladinWearingAgainst(t, nil, encounter, proto.Race_RaceUndead)
		drain := pal.GetSpell(core.ActionID{SpellID: 1260198})
		if drain == nil {
			t.Fatal("the Undead paladin has no Touch of the Grave drain")
		}
		consecration := pal.GetSpell(core.ActionID{SpellID: paladin.ConsecrationRankMap.Highest().ID})
		if consecration == nil {
			t.Fatal("the paladin has no Consecration")
		}
		pal.AddStatDynamic(sim, stats.SpellHitPercent, 100)

		// The ground is taken away half a second after each cast, before its first tick, so only the
		// cast's landings could set off the drain.
		for range casts {
			consecration.SkipCastAndApplyEffects(sim, target)
			sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{
				DoAt:     sim.CurrentTime + 500*time.Millisecond,
				OnAction: consecration.AOEDot().Deactivate,
			}))
			for consecration.AOEDot().IsActive() {
				if sim.Step() {
					t.Fatal("the encounter ended before the casts did")
				}
			}
		}

		drains := 0
		for _, metrics := range drain.SpellMetrics {
			drains += int(metrics.Casts)
		}
		if drains != 0 {
			t.Errorf("on %d enemies at a time, %d Consecration casts drained %d times, want none", enemies, casts, drains)
		}
	}
}

// Bolt's Charged Eyeball hears neither Consecration's cast nor its ticks. A white hit that lands still
// casts the shock (EV-0226).
func TestConsecrationSetsOffNoBoltsChargedEyeball(t *testing.T) {
	const casts, swings, enemies = 500, 400, 3
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Duration = casts*13 + swings*4
	for range enemies - 1 {
		encounter.Targets = append(encounter.Targets, core.NewDefaultTarget())
	}
	sim, pal, target := paladinWearingAgainst(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: 280032}, encounter, proto.Race_RaceHuman)
	shock := pal.GetSpell(core.ActionID{SpellID: 1315996})
	if shock == nil {
		t.Fatal("280032 registers no 1315996")
	}
	consecration := pal.GetSpell(core.ActionID{SpellID: paladin.ConsecrationRankMap.Highest().ID})
	if consecration == nil {
		t.Fatal("the paladin has no Consecration")
	}
	pal.AddStatDynamic(sim, stats.SpellHitPercent, 100)
	shocks := func() int {
		n := 0
		for _, metrics := range shock.SpellMetrics {
			n += int(metrics.Casts)
		}
		return n
	}
	for range casts {
		consecration.SkipCastAndApplyEffects(sim, target)
		for consecration.AOEDot().IsActive() {
			if sim.Step() {
				t.Fatal("the encounter ended before the casts did")
			}
		}
		stepFor(t, sim, 4*time.Second) // past the trinket's 3 s cooldown, so every landing and tick could roll
	}
	if got := shocks(); got != 0 {
		t.Errorf("%d Consecration casts on %d enemies cast %d shocks, want none", casts, enemies, got)
	}

	swing := pal.AutoAttacks.MHAuto()
	for range swings {
		pal.OnSpellHitDealt(sim, swing, &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 1})
		stepFor(t, sim, 4*time.Second) // past the trinket's 3 s cooldown
	}
	if shocks() == 0 {
		t.Errorf("%d landed white hits cast no shock, want some", swings)
	}
}

// A Consecration tick that crits gives no Vengeance. A white swing that crits gives it (EV-0226).
func TestConsecrationCritGivesNoVengeance(t *testing.T) {
	sim, pal, target := paladinWearing(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: 21395})
	vengeance := pal.GetAuraByID(core.ActionID{SpellID: 20050})
	consecration := pal.GetSpell(core.ActionID{SpellID: paladin.ConsecrationRankMap.Highest().ID})
	if vengeance == nil || consecration == nil || consecration.RelatedDotSpell == nil {
		t.Fatal("the paladin has no Vengeance or no Consecration ticks")
	}
	pal.AddStatDynamic(sim, stats.SpellHitPercent, 100)
	pal.AddStatDynamic(sim, stats.SpellCritPercent, 100)

	consecration.SkipCastAndApplyEffects(sim, target)
	for consecration.AOEDot().IsActive() {
		if sim.Step() {
			t.Fatal("the encounter ended before the ticks did")
		}
		if vengeance.IsActive() {
			t.Fatal("a Consecration tick that crit gave Vengeance, want none")
		}
	}
	if consecration.RelatedDotSpell.SpellMetrics[target.UnitIndex].CritTicks == 0 {
		t.Fatal("no Consecration tick crit, so the check proves nothing")
	}

	pal.OnSpellHitDealt(sim, pal.AutoAttacks.MHAuto(), &core.SpellResult{Target: target, Outcome: core.OutcomeCrit, Damage: 1})
	stepFor(t, sim, 5*core.SpellBatchWindow)
	if !vengeance.IsActive() {
		t.Error("a white swing that crit gave no Vengeance")
	}
}

// Runs the sim for d, failing the test if the encounter ends first. An empty action at the end keeps the queue from
// running dry.
func stepFor(t *testing.T, sim *core.Simulation, d time.Duration) {
	t.Helper()
	end := sim.CurrentTime + d
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{DoAt: end, OnAction: func(*core.Simulation) {}}))
	for sim.CurrentTime < end {
		if sim.Step() {
			t.Fatal("the encounter ended early")
		}
	}
}
