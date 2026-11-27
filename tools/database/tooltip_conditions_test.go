package database

import (
	"reflect"
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/tooltip"
)

// Defensive Stance 71 is known to every warrior and up in its own form; Mortal Strike's trainer rank 21551 is known
// with the talent 12294 it goes back to; an Engraving rune, Engrave Gloves - Mutilate 400094, is known to no one.
func TestTooltipConditionsReadTheClassFormAndTalent(t *testing.T) {
	helper := openBuffTestDB(t)
	withDBCInputs(t)
	instance := dbc.GetDBC()

	var form int32
	for _, effect := range instance.SpellEffectsInOrder(71) {
		if effect.EffectAura == dbcenums.A_MOD_SHAPESHIFT {
			form = int32(effect.EffectMiscValues[0])
		}
	}
	if form == 0 {
		t.Fatal("Defensive Stance shifts into no form")
	}

	conditions, err := tooltipConditions(helper, instance, map[int32]int32{12294: 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		what string
		got  tooltip.Expr
		want tooltip.Expr
	}{
		{"knowing Defensive Stance", conditions.Known[71], tooltip.PlayerOfClass(int32(proto.Class_ClassWarrior))},
		{"standing in Defensive Stance", conditions.Auras[71], tooltip.PlayerInForm(form)},
		{"knowing Mortal Strike's trainer rank", conditions.Known[21551], tooltip.TalentTaken(12294)},
		{"knowing an Engraving rune", conditions.Known[400094], nil},
	} {
		if !reflect.DeepEqual(tc.got, tc.want) {
			t.Errorf("%s reads %v, want %v", tc.what, tc.got, tc.want)
		}
	}
}
