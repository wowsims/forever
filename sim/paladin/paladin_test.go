package paladin_test

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
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
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for slot := range items {
		items[slot] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotNeck] = &proto.ItemSpec{Id: listenerNeck}
	player := &proto.Player{
		Name:          "Paladin",
		Race:          proto.Race_RaceHuman,
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
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pal := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
	pal.AutoAttacks.CancelAutoSwing(sim)
	heard = tally{}
	return sim, pal, pal.CurrentTarget
}

// A landed Judgement of Command rolls a weapon proc twice: once as its dummy lands and once on its
// damage. A listener that is no weapon proc hears the damage alone, on the hit and on the cast result,
// and a judgement that misses rolls none of them (EV-0156).
func TestJudgementOfCommandRollsAWeaponProcTwice(t *testing.T) {
	const casts = 50
	for _, c := range []struct {
		name  string
		lands bool
		want  tally
	}{
		{"a judgement that lands", true, tally{castResults: casts, weaponProc: 2 * casts, hitListener: casts, castResultListener: casts}},
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
