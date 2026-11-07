//go:build with_db

package forever_test

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/testagent"
)

// The relic stand-ins' heals, under the class flags the relics' triggers name: Healing Touch is the
// druid family's bit 5, which Idol of Longevity's 28847 names, and Regrowth bit 6, which it does not;
// Lesser Healing Wave is the shaman family's bit 5, which Totem of Flowing Water's 28849 names.
var (
	stubHealingTouch      = core.ActionID{SpellID: 993105}
	stubRegrowth          = core.ActionID{SpellID: 993106}
	stubLesserHealingWave = core.ActionID{SpellID: 993107}
)

func init() {
	testagent.Register(proto.Player_RestorationDruid{}, proto.Spec_SpecRestorationDruid, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_RestorationDruid)
	}, func(agent *testagent.Agent) {
		agent.EnableManaBar()
		agent.RegisterSpell(flatHeal(stubHealingTouch, 7, 5))
		agent.RegisterSpell(flatHeal(stubRegrowth, 7, 6))
	})
	testagent.Register(proto.Player_RestorationShaman{}, proto.Spec_SpecRestorationShaman, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_RestorationShaman)
	}, func(agent *testagent.Agent) {
		agent.EnableManaBar()
		agent.RegisterSpell(flatHeal(stubLesserHealingWave, 11, 5))
	})
}

// A heal of 100 under the family's bit, costing nothing so the refund alone moves the mana.
func flatHeal(id core.ActionID, family int32, bit uint) core.SpellConfig {
	return core.SpellConfig{
		ActionID:         id,
		SpellSchool:      core.SpellSchoolNature,
		ProcMask:         core.ProcMaskSpellHealing,
		Flags:            core.SpellFlagHelpful,
		ClassFlags:       core.ClassFlags{Family: family, Mask: [4]uint32{1 << bit}},
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, 100, spell.OutcomeHealing)
		},
	}
}

// Idol of Longevity 23004 refunds 28848's mana on each Healing Touch cast and nothing on another heal;
// Totem of Flowing Water 23005 refunds 28850's on each Lesser Healing Wave cast. Each trigger carries
// its class's mask, read as the relic's lock on the spell it names.
func TestRelicsRefundManaOnTheCastTheyName(t *testing.T) {
	for _, tc := range []struct {
		name           string
		class          proto.Class
		itemID, gainID int32
		named, other   core.ActionID
	}{
		{"Idol of Longevity", proto.Class_ClassDruid, 23004, 28848, stubHealingTouch, stubRegrowth},
		{"Totem of Flowing Water", proto.Class_ClassShaman, 23005, 28850, stubLesserHealingWave, core.ActionID{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gain := spelldata.MustFind(tc.gainID).EnergizeEffect().Average()
			sim, character := itemWearer(t, tc.class, proto.ItemSlot_ItemSlotRanged, tc.itemID)
			character.SpendMana(sim, 4*gain, character.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen}))

			before := character.CurrentMana()
			castHeal(t, sim, character, character.GetSpell(tc.named))
			if got := character.CurrentMana() - before; gain <= 0 || got != gain {
				t.Errorf("a %s cast restored %v mana, want the row's %v", tc.named, got, gain)
			}
			castHeal(t, sim, character, character.GetSpell(tc.named))
			if got := character.CurrentMana() - before; got != 2*gain {
				t.Errorf("a second cast left the mana restored at %v, want %v", got, 2*gain)
			}
			if tc.other.IsEmptyAction() {
				return
			}
			castHeal(t, sim, character, character.GetSpell(tc.other))
			if got := character.CurrentMana() - before; got != 2*gain {
				t.Errorf("after a %s cast the mana restored is %v, want %v still", tc.other, got, 2*gain)
			}
		})
	}
}

// A cast on the caster, and the spell batch window the listener resolves after.
func castHeal(t *testing.T, sim *core.Simulation, character *core.Character, spell *core.Spell) {
	t.Helper()
	if spell == nil {
		t.Fatal("the stand-in registers no such heal")
	}
	if !spell.Cast(sim, &character.Unit) {
		t.Fatalf("%s failed to cast", spell.ActionID)
	}
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
}
