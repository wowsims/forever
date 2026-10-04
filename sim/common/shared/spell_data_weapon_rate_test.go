package shared

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Two weapons stating their own procs-per-minute rate on the same trigger spell each get a manager
// at their rate: Wound 16405 at 3 PPM on one, 1 PPM on the other, read off the item's effect entry
// when the shape passes none.
func TestWeaponProcRateComesFromTheItemEffect(t *testing.T) {
	const trigger int32 = 16405
	for _, tc := range []struct {
		itemID int32
		ppm    float64
	}{
		{992601, 3},
		{992602, 1},
	} {
		withAuraItem(tc.itemID, proto.ItemType_ItemTypeWeapon, &proto.ItemEffect{BuffId: trigger,
			Effect: &proto.ItemEffect_Proc{Proc: &proto.ProcEffect{ProcRate: &proto.ProcEffect_Ppm{Ppm: tc.ppm}}}})

		agent := newTestAgent()
		character := agent.character
		character.Equipment = core.ProtoToEquipment(&proto.EquipmentSpec{Items: testHands(&proto.ItemSpec{Id: tc.itemID}, &proto.ItemSpec{})})
		character.EnableAutoAttacks(agent, core.AutoAttackOptions{
			MainHand:       character.WeaponFromMainHand(),
			AutoSwingMelee: true,
		})

		cfg := SpellDataProc{Name: "Test Wound Weapon", ItemID: tc.itemID, TriggerSpellID: trigger, BuffSpellID: trigger, IsWeaponProc: true}
		config := spellDataProcListener(character, cfg, cfg.effectSource(), spelldata.MustFind(trigger))
		if config.ProcChance != 0 || config.DPM == nil {
			t.Fatalf("item %d: chance %v, manager %v; want a manager and no flat chance", tc.itemID, config.ProcChance, config.DPM)
		}
		want := testOneHander(tc.itemID).WeaponSpeed * tc.ppm / 60
		if got := config.DPM.Chance(core.ProcMaskMeleeMHAuto, nil); math.Abs(got-want) > 1e-9 {
			t.Errorf("item %d at %v PPM rolls %v per main-hand swing, want %v", tc.itemID, tc.ppm, got, want)
		}
	}
}
