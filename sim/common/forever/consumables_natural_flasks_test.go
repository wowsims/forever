//go:build with_db

package forever_test

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/core/testagent"
)

func init() {
	testagent.RegisterAsDpsWarrior()
}

func warriorWithFlask(t *testing.T, flaskID int32) *core.Character {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotTrinket2+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Flasked", Class: proto.Class_ClassWarrior, Race: proto.Race_RaceOrc,
				Spec: &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{
					ClassOptions: &proto.WarriorOptions{DefaultStance: proto.WarriorStance_WarriorStanceBerserker}}}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: &proto.ConsumesSpec{FlaskId: flaskID},
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Each flask's 1293740-1293743 states Stamina in its first effect, and in its A_DUMMY the zone bonus percent.
func TestNaturalFlasksGrantStaminaAndTheirZoneBonus(t *testing.T) {
	bare := warriorWithFlask(t, 0)
	for _, tc := range []struct {
		itemID   int32
		bonusTo  []stats.Stat
		hastened bool
	}{
		{274273, []stats.Stat{stats.PhysicalHitPercent, stats.SpellHitPercent}, false},
		{274274, []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent}, false},
		{274275, []stats.Stat{stats.ExpertisePercent}, false},
		{274276, nil, true},
	} {
		flasked := warriorWithFlask(t, tc.itemID)
		if flasked.GetAura(core.GetConsumableByID(tc.itemID).Name) == nil {
			t.Errorf("%d put up no aura", tc.itemID)
			continue
		}
		row := spelldata.MustFind(core.GetConsumableByID(tc.itemID).OnUse().BuffId)
		stamina, percent := row.EffectN(1).Average(), row.EffectN(2).Average()
		if stamina == 0 || percent == 0 {
			t.Fatalf("%d: %d states %v Stamina and a %v%% zone bonus, want both", tc.itemID, row.ID, stamina, percent)
		}
		var bonus stats.Stats
		for _, stat := range tc.bonusTo {
			bonus[stat] = percent
		}
		speed := 1.0
		if tc.hastened {
			speed = 1 + percent/100
		}

		gained := flasked.GetStats().Subtract(bare.GetStats())
		if gained[stats.Stamina] != stamina {
			t.Errorf("%d: Stamina rose by %v, want %v", tc.itemID, gained[stats.Stamina], stamina)
		}
		for _, stat := range []stats.Stat{stats.PhysicalHitPercent, stats.SpellHitPercent, stats.PhysicalCritPercent,
			stats.SpellCritPercent, stats.ExpertiseRating, stats.ExpertisePercent} {
			if math.Abs(gained[stat]-bonus[stat]) > 1e-9 {
				t.Errorf("%d: %s rose by %v, want %v", tc.itemID, stat.StatName(), gained[stat], bonus[stat])
			}
		}

		for name, got := range map[string]float64{
			"attack speed": flasked.PseudoStats.AttackSpeedMultiplier / bare.PseudoStats.AttackSpeedMultiplier,
			"cast speed":   flasked.PseudoStats.CastSpeedMultiplier / bare.PseudoStats.CastSpeedMultiplier,
		} {
			if math.Abs(got-speed) > 1e-9 {
				t.Errorf("%d: %s is %v times the bare warrior's, want %v", tc.itemID, name, got, speed)
			}
		}
	}
}
