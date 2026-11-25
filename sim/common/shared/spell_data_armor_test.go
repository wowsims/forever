package shared

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// A spell built from its row ticks through all, a tenth or none of the target's armor as the row's
// attributes say: the warrior's Rend sets both bits and ignores armor, a weapon's Rend sets only the
// first and meets a tenth, and Wound and Bigger Rock to the Face set neither and meet it whole, a bleed
// or not.
func TestPeriodicPhysicalRowsMeetTheArmorTheirAttributesName(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{shaman("Wearer", nil)},
			Buffs:   &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 60, Stats: stats.Stats{stats.Armor: 3000}.ToProtoArray()}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	character := auraTesterAt(sim, 0).GetCharacter()
	target := character.CurrentTarget
	attackTable := character.AttackTables[target.UnitIndex]
	tenth := target.Armor() / 10
	if tenth <= 0 {
		t.Fatal("the target has no armor")
	}
	whole := attackTable.GetArmorDamageModifier(nil)

	for _, tc := range []struct {
		name                    string
		id                      int32
		ignores, tenth, isBleed bool
		want                    float64
	}{
		{"the warrior's Rend", 772, true, true, true, 1},
		{"Gatorbite Axe's Rend", 21949, false, true, true, 1 - tenth/(tenth+400+85*float64(character.Level))},
		{"Wound", 17407, false, false, false, whole},
		{"Bigger Rock to the Face", 1306816, false, false, true, whole},
	} {
		row := spelldata.MustFind(tc.id)
		if row.SpellSchool() != core.SpellSchoolPhysical || row.PeriodicDamageEffect() == spelldata.NilEffect ||
			row.HasAttr(dbcenums.ATTR_INDEX_EX_6, dbcenums.ATTR_EX_6_TICKS_IGNORE_ARMOR) != tc.ignores ||
			row.HasAttr(dbcenums.ATTR_INDEX_BASE, dbcenums.ATTR_TICKS_MEET_A_TENTH_OF_ARMOR) != tc.tenth ||
			row.AppliesBleed() != tc.isBleed {
			t.Fatalf("%s's row is no physical periodic damage with the attributes and mechanic this case needs", tc.name)
		}
		spell := character.RegisterSpell(spelldata.SpellConfig(&character.Unit, row, spelldata.Melee(core.ProcMaskEmpty), spelldata.Proc()))
		if got, _ := spell.ResistanceMultiplier(sim, true, attackTable); !core.WithinToleranceFloat64(tc.want, got, 1e-9) {
			t.Errorf("%s's ticks keep %v of their damage against armor, want %v", tc.name, got, tc.want)
		}
	}
}
