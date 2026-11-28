package tooltip

import (
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/tools/database/dbc"
)

func Test_WhenInvalidTernaryGiven_ThenProperlyApplyFixes(t *testing.T) {
	provider := fixtureProvider(fixtureRow{
		spell:   dbc.Spell{ID: 101, Duration: 10000},
		effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_MOD_DECREASE_SPEED}},
	})
	renderTooltipTest(t, provider, 101,
		"Damage every ${$101d3/10}.2 seconds$?$w1!=0[ and movement slowed by $w1%][].",
		"Damage every 1.00 seconds.",
	)
}

func Test_WhenLConditionGiven_ThenProperlyEvaluate(t *testing.T) {
	for _, c := range []struct {
		removes float64
		want    string
	}{
		{1, "Purges the enemy target, removing 1 magic effect."},
		{2, "Purges the enemy target, removing 2 magic effects."},
	} {
		provider := fixtureProvider(fixtureRow{
			spell:   dbc.Spell{ID: 101},
			effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_DISPEL, EffectBasePoints: c.removes}},
		})
		renderTooltipTest(t, provider, 101, "Purges the enemy target, removing $s1 magic $leffect:effects;.", c.want)
	}
}

func Test_WhenGivenDurationShorter60_ThenRenderSeconds(t *testing.T) {
	provider := fixtureProvider(fixtureRow{spell: dbc.Spell{ID: 101, Duration: 3000}})
	renderTooltipTest(t, provider, 101, "Stuns target for $d.", "Stuns target for 3s.")
}

func Test_WhenGivenDurationLongerThan2Hours_ThenRnderHrs(t *testing.T) {
	provider := fixtureProvider(fixtureRow{
		spell:   dbc.Spell{ID: 101, Duration: 3 * 60 * 60 * 1000},
		effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_MOD_INCREASE_MOUNTED_SPEED, EffectBasePoints: 10}},
	})
	renderTooltipTest(t, provider, 101,
		"Increases mounted movement speed by $s1% for $d, stacking with all other mounted movement speed bonuses.",
		"Increases mounted movement speed by 10% for 3hrs, stacking with all other mounted movement speed bonuses.",
	)
}

// Uppercase $H is the proc chance as much as lowercase $h.
func Test_WhenUppercaseProcChanceGiven_ThenRenderTheChance(t *testing.T) {
	provider := fixtureProvider(
		fixtureRow{spell: dbc.Spell{ID: 101, ProcChance: 5, ProcCategoryRecovery: 10000}},
		fixtureRow{spell: dbc.Spell{ID: 102, Duration: 3000}},
	)
	renderTooltipTest(t, provider, 101,
		"When struck in combat has a $H% chance to make you invulnerable to melee damage for $102d. This effect can only occur once every $proccooldown sec.",
		"When struck in combat has a 5% chance to make you invulnerable to melee damage for 3s. This effect can only occur once every 10 sec.",
	)
}

// A reference into another spell renders the amount that spell states, with no spell power on top.
func Test_WhenOtherSpellsDamageReferenced_ThenRenderItsStatedAmount(t *testing.T) {
	provider := fixtureProvider(
		fixtureRow{spell: dbc.Spell{ID: 101, Duration: 10000}},
		fixtureRow{
			spell:   dbc.Spell{ID: 102},
			effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_SCHOOL_DAMAGE, EffectBasePoints: 200, EffectBonusCoefficient: 1}},
		},
	)
	renderTooltipTest(t, provider, 101,
		"Infuses you with Arcane energy, causing your next Arcane Shot fired within $d to detonate at the target. The Arcane Detonation will deal $102s1 damage to enemies near the target.",
		"Infuses you with Arcane energy, causing your next Arcane Shot fired within 10s to detonate at the target. The Arcane Detonation will deal 200 damage to enemies near the target.",
	)
}

// A computed value is unsigned like a plain one, since the tooltip writes its own words for a reduction.
func Test_WhenNegativeValueComputed_ThenRenderItUnsigned(t *testing.T) {
	provider := fixtureProvider(fixtureRow{
		spell:   dbc.Spell{ID: 101},
		effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_ADD_FLAT_MODIFIER, EffectBasePoints: -30}},
	})
	renderTooltipTest(t, provider, 101, "Hamstring Rage cost reduced by $/10;s1.", "Hamstring Rage cost reduced by 3.0.")
}

type fixtureRow struct {
	spell   dbc.Spell
	effects []dbc.SpellEffect
}

// The game files' data provider the item generator renders with, over these rows alone.
func fixtureProvider(rows ...fixtureRow) DBCTooltipDataProvider {
	db := dbc.NewDBC()
	for _, r := range rows {
		db.Spells[int(r.spell.ID)] = r.spell
		effects := map[int]dbc.SpellEffect{}
		for _, e := range r.effects {
			effects[e.EffectIndex] = e
		}
		db.SpellEffects[int(r.spell.ID)] = effects
	}
	return DBCTooltipDataProvider{DBC: db}
}

func renderTooltipTest(t *testing.T, dataProvider TooltipDataProvider, spellId int64, description string, expectedDescription string) {
	t.Helper()
	tp, error := ParseTooltip(description, dataProvider, spellId)

	if error != nil {
		t.Fatal(error)
	}

	if got := tp.String(); got != expectedDescription {
		t.Errorf("%q renders %q, want %q", description, got, expectedDescription)
	}
}
