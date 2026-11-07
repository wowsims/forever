package database

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
)

// Moves to the repository root and skips without the game files DBC inputs, which are not in the repo.
func withDBCInputs(t *testing.T) {
	t.Helper()

	inRepositoryRoot(t)
	if _, err := os.Stat("./assets/db_inputs/dbc/spells.json"); err != nil {
		t.Skip("no game files DBC inputs in assets/db_inputs/dbc - run `make db` from a local WoW install to enable this test")
	}
}

// Which slots of a real enchant reach the generator, in which shape and at which chance. Every
// entry names the reason it is refused, or none where it registers.
func TestEnchantProcRouting(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	type want struct {
		trigger int
		damage  bool
		reason  string
	}

	for _, tc := range []struct {
		effectID int
		name     string
		want     []want
	}{
		{36, "Fiery Blaze: Effect 1 states the tooltip's chance", []want{{6297, true, ""}}},
		{803, "Fiery Weapon: Effect 1 states none", []want{{13897, true, spelldata.ReasonStatesNoRate}}},
		{1899, "Unholy Weapon: 'often' is a rate the game files do not carry", []want{{20006, false, spelldata.ReasonStatesNoRate}}},
		{1900, "Crusader: the combat spell only, the aura 458112 applies the same Holy Strength", []want{{20007, false, spelldata.ReasonStatesNoRate}}},
		{7940, "Grand Crusader: the same shape as Crusader", []want{{1231124, false, spelldata.ReasonStatesNoRate}}},
		{7210, "Dismantle: two auras, 100 beside 'sometimes' and a chance, both hitting mechanicals only", []want{
			{435467, true, spelldata.ReasonStatesNoRate},
			{442206, true, "mechanicals"},
		}},
		{7223, "Retricutioner: a damage shield, kept up as an equip aura", []want{{435901, false, ""}}},
		{7941, "Grand Arcanist: spell power and healing register, the mana beside them is not a stat", []want{{1231152, false, ""}}},
		{8216, "Insight: a chance, a buff that multiplies Spirit", []want{{1248758, false, ""}}},
		{8217, "Revelation: 100 beside 'a chance to trigger'", []want{{1248806, false, spelldata.ReasonStatesNoRate}}},
		{8721, "Recovery: 100 beside a cooldown, a heal on the wearer's attack dodged or parried", []want{{1248761, false, ""}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			enchant, ok := instance.EnchantsByEffectId[tc.effectID]
			if !ok {
				t.Fatalf("enchant %d is not in the enchant inputs", tc.effectID)
			}
			slots := enchant.ProcSlots()
			got := routeEnchantProcs(slots, instance)
			if len(got) != len(tc.want) {
				t.Fatalf("%d routings, want %d", len(got), len(tc.want))
			}

			for i, w := range tc.want {
				r := got[i]
				chance, wantChance := 0, 0
				if r.IsWeaponProc {
					chance = int(spelldata.Find(int32(r.TriggerSpellID)).ProcChance)
					for _, slot := range slots {
						if slot.IsCombatSpell && slot.SpellID == r.TriggerSpellID {
							wantChance = slot.ChancePct
						}
					}
				}
				if damage := r.Shape == ShapeDamage; r.TriggerSpellID != w.trigger || damage != w.damage || chance != wantChance {
					t.Errorf("routing %d: trigger %d, damage %v, combat chance %d%%; want %d, %v, the slot's %d%%",
						i, r.TriggerSpellID, damage, chance, w.trigger, w.damage, wantChance)
				}
				if w.reason == "" && !r.Supported() {
					t.Errorf("routing %d is refused (%s), want it registered", i, r.Reason())
				}
				if tc.effectID == 7223 && r.Shape != ShapeEquip {
					t.Errorf("routing %d took shape %v, want the equip aura", i, r.Shape)
				}
				if w.reason != "" && !strings.Contains(r.Reason(), w.reason) {
					t.Errorf("routing %d: reason %q, want it to name %q", i, r.Reason(), w.reason)
				}
			}
		})
	}
}

// An equip slot whose row states a proc trigger beside an equip aura kind stays a proc, the way an
// item's does: a trigger is added to Retricutioner's 435901 in the game files for the test, and the
// slot leaves the equip aura branch.
func TestEnchantEquipSlotWithATriggerStaysAProc(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	slot := dbc.EnchantProcSlot{SpellID: 435901}
	if r := routeEnchantSlot(slot, instance); r.Shape != ShapeEquip {
		t.Fatalf("Retricutioner's slot took shape %v, want the equip aura", r.Shape)
	}

	effects := instance.SpellEffects[435901]
	effects[9] = dbc.SpellEffect{EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_PROC_TRIGGER_SPELL, EffectTriggerSpell: 23781}
	t.Cleanup(func() { delete(effects, 9) })
	if r := routeEnchantSlot(slot, instance); r.Shape == ShapeEquip {
		t.Errorf("a slot whose row states a trigger took the equip aura shape, want it read as a proc")
	}
}

// A PPM override answers "states no rate" on the spell the slot is routed through - the combat spell
// of an Effect 1 slot, the equip aura of an Effect 3 one - and leaves every other refusal in place.
// On an equip aura that hears only spells the rate is refused in turn: it would never roll.
func TestEnchantProcRoutingTakesAPPMOverride(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		effectID   int
		name       string
		overrideOn int32
		refusedFor string
	}{
		{803, "Fiery Weapon: Effect 1, a damage combat spell", 13897, ""},
		{1899, "Unholy Weapon: Effect 1, a combat spell granting a buff", 20006, ""},
		{8217, "Revelation: Effect 3, the aura's 100 beside 'a chance to trigger', heard on spells only", 1248806,
			spelldata.ReasonPPMHearsNoWeaponHits},
	} {
		t.Run(tc.name, func(t *testing.T) {
			enchant := instance.EnchantsByEffectId[tc.effectID]
			route := func() *ProcRouting {
				t.Helper()
				got := routeEnchantProcs(enchant.ProcSlots(), instance)
				if len(got) != 1 || got[0].TriggerSpellID != int(tc.overrideOn) {
					t.Fatalf("routings %v, want one through %d", got, tc.overrideOn)
				}
				return got[0]
			}

			before := route()
			if !statesNoRate(before) {
				t.Fatalf("reason %q names no missing rate to answer", before.Reason())
			}

			withPPMOverride(t, tc.overrideOn, 2)

			after := route()
			if tc.refusedFor != "" && !slices.Contains(after.Unsupported, tc.refusedFor) {
				t.Errorf("unsupported = %v with the override, want it to name %q", after.Unsupported, tc.refusedFor)
			}
			want := slices.DeleteFunc(slices.Clone(before.Unsupported), func(reason string) bool {
				return reason == spelldata.ReasonStatesNoRate
			})
			got := slices.DeleteFunc(slices.Clone(after.Unsupported), func(reason string) bool {
				return reason == tc.refusedFor
			})
			if !slices.Equal(got, want) {
				t.Errorf("unsupported = %v with the override, want %v beside %q", after.Unsupported, want, tc.refusedFor)
			}
		})
	}
}

// Insight's effect entry resolves no flat stats, so the slot applies its buff 1299796 as auras, whose
// row multiplies Spirit.
func TestEnchantPercentStatBuffIsTheAppliedSpell(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	enchant := instance.EnchantsByEffectId[8216]
	got := routeEnchantProcs(enchant.ProcSlots(), instance)
	if len(got) != 1 || got[0].Shape != ShapeAura || got[0].TriggerSpellID != 1248758 || got[0].BuffSpellID != 1299796 {
		t.Fatalf("routings %v, want one aura from 1248758 to 1299796", got)
	}
	if !got[0].Supported() {
		t.Fatalf("refused (%s), want it registered", got[0].Reason())
	}

	buff := spelldata.Find(1299796)
	if e := buff.FirstAura(dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE); e == spelldata.NilEffect || e.Misc2 != 16 {
		t.Errorf("1299796 states %+v, want A_MOD_TOTAL_STAT_PERCENTAGE on the stat mask 16 (Spirit)", e)
	}
}

// Recovery's equip aura 1248761 carries no description; its grant 1248760 reads "trigger Recovery when
// you are Parried or Dodged". The row carries that outcome, a ProcCategoryRecovery and its ProcChance
// of 100, and the routing casts the E_HEAL_PCT heal 1248759 it applies.
func TestRecoveryHealsOnTheWearersAttackDodgedOrParried(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	enchant := instance.EnchantsByEffectId[8721]
	got := routeEnchantProcs(enchant.ProcSlots(), instance)
	if len(got) != 1 {
		t.Fatalf("%d routings, want 1", len(got))
	}
	r := got[0]

	if !r.Supported() {
		t.Fatalf("refused (%s), want it registered", r.Reason())
	}
	if r.Shape != ShapeHeal || r.TriggerSpellID != 1248761 || r.BuffSpellID != 1248759 {
		t.Errorf("shape %v, trigger %d, buff %d; want a heal from 1248761 casting 1248759",
			r.Shape, r.TriggerSpellID, r.BuffSpellID)
	}
	trigger := spelldata.Find(1248761)
	if want := core.ProcHintAttackDodged | core.ProcHintAttackParried; trigger.ProcHint != want {
		t.Errorf("hint %q, want %q", formatProcHint(trigger.ProcHint), formatProcHint(want))
	}

	decoded := core.DecodeProcTypeMask(trigger.ProcFlags, trigger.ProcHint)
	if decoded.Callback != core.CallbackOnSpellHitDealt {
		t.Errorf("callback %v, want the wearer's own hits dealt", decoded.Callback)
	}
	if decoded.Outcome != core.OutcomeDodge|core.OutcomeParry || decoded.RequireDamageDealt {
		t.Errorf("outcome %d, require damage %v; want dodge or parry with no damage dealt",
			decoded.Outcome, decoded.RequireDamageDealt)
	}
	if trigger.StatedChance() != 1 || trigger.ICD() <= 0 {
		t.Errorf("chance %v, ICD %v; want every time, with a cooldown between", trigger.StatedChance(), trigger.ICD())
	}

	heal := spelldata.Find(1248759).ProcHealEffect()
	if heal.Type != dbcenums.E_HEAL_PCT || heal.Percent() <= 0 {
		t.Errorf("heal effect %v at %v, want E_HEAL_PCT at a share of maximum health", heal.Type, heal.Percent())
	}
}

// The store row as a PPM override writes it, restored when the test ends.
func withPPMOverride(t *testing.T, spellID int32, ppm float32) {
	t.Helper()
	row := spelldata.Find(spellID)
	if row == spelldata.Nil {
		t.Fatalf("spell %d is not in the store", spellID)
	}

	original := *row
	overridden := original
	overridden.RPPM = ppm
	overridden.ProcChanceSource, overridden.ProcChanceEffect = spelldata.ProcChancePPM, 0
	*row = overridden
	t.Cleanup(func() { *row = original })
}

// Recovery's tooltip states a cooldown with "more often than", which is not the "often" of a rate.
func TestTooltipRateWording(t *testing.T) {
	for tooltip, want := range map[string]bool{
		"Permanently enchant a melee weapon to often strike for 40 additional fire damage.":                   true,
		"Permanently enchant a Weapon to cause all spells and attacks to sometimes deal 76 additional damage": true,
		"Permanently enchant a Melee Weapon to have a chance to trigger Revelation when a non-periodic spell": true,
		"healing you for 5% of your maximum health. Cannot occur more often than once every 10 sec.":          false,
	} {
		if got := tooltipStatesAnUnknownRate(tooltip); got != want {
			t.Errorf("%q: %v, want %v", tooltip, got, want)
		}
	}
}
