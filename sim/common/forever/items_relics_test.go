//go:build with_db

package forever_test

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/testagent"
)

// The relic stand-ins' heals, under the class flags the store files the real heals under: Healing
// Touch 5185, which Idol of Longevity's 28847 names, and Regrowth 8936, which it does not; Lesser
// Healing Wave 8004, which Totem of Flowing Water's 28849 names, and Healing Wave 331.
var (
	stubHealingTouch      = core.ActionID{SpellID: 993105}
	stubRegrowth          = core.ActionID{SpellID: 993106}
	stubLesserHealingWave = core.ActionID{SpellID: 993107}
	stubHealingWave       = core.ActionID{SpellID: 993108}
)

func init() {
	testagent.Register(proto.Player_RestorationDruid{}, proto.Spec_SpecRestorationDruid, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_RestorationDruid)
	}, func(agent *testagent.Agent) {
		agent.EnableManaBar()
		agent.RegisterSpell(flatHeal(stubHealingTouch, 5185))
		agent.RegisterSpell(flatHeal(stubRegrowth, 8936))
	})
	testagent.Register(proto.Player_RestorationShaman{}, proto.Spec_SpecRestorationShaman, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_RestorationShaman)
	}, func(agent *testagent.Agent) {
		agent.EnableManaBar()
		agent.RegisterSpell(flatHeal(stubLesserHealingWave, 8004))
		agent.RegisterSpell(flatHeal(stubHealingWave, 331))
	})
}

// A heal of 100 under the class flags of the row it stands in for, costing nothing so the refund
// alone moves the mana.
func flatHeal(id core.ActionID, rowID int32) core.SpellConfig {
	return core.SpellConfig{
		ActionID:         id,
		SpellSchool:      core.SpellSchoolNature,
		ProcMask:         core.ProcMaskSpellHealing,
		Flags:            core.SpellFlagHelpful,
		ClassFlags:       spelldata.MustFind(rowID).ClassFlags,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, 100, spell.OutcomeHealing)
		},
	}
}

// Idol of Longevity 23004 refunds 28848's mana on each Healing Touch cast and nothing on a Regrowth;
// Totem of Flowing Water 23005 refunds 28850's on each Lesser Healing Wave cast and nothing on a
// Healing Wave. Each trigger carries its class's mask, read as the relic's lock on the spell it
// names, the totem's through the override that moves it onto Lesser Healing Wave's bit.
func TestRelicsRefundManaOnTheCastTheyName(t *testing.T) {
	for _, tc := range []struct {
		name           string
		class          proto.Class
		itemID, gainID int32
		named, other   core.ActionID
	}{
		{"Idol of Longevity", proto.Class_ClassDruid, 23004, 28848, stubHealingTouch, stubRegrowth},
		{"Totem of Flowing Water", proto.Class_ClassShaman, 23005, 28850, stubLesserHealingWave, stubHealingWave},
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
