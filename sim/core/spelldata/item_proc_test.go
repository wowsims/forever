package spelldata

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// One refusal per shape, against the rows the resolver's own tests are built from.
func TestItemProcUnsupported(t *testing.T) {
	withRows(t, procRows())

	for _, tc := range []struct {
		spellID      int32
		isWeaponProc bool
		want         []string
		why          string
	}{
		{2200, false, nil, "a chance on an effect of its own and a mask of melee hits"},
		{2000, false, []string{"the trigger carries a class mask"},
			"a column chance and a mask of melee hits, behind a filter the item cannot reproduce"},
		{2300, false, []string{"states no rate"}, "the 101 sentinel with no override to answer it"},
		{2400, false, nil, "the same sentinel, answered by an override into RPPM"},
		{2500, false, []string{"no callback in the proc mask", "no proc mask to measure its rate on"},
			"procs per minute with no hits to measure them on"},
		{2500, true, nil, "the same row on a weapon, where the weapon counts the hits"},
		{2100, true, []string{"states no rate"},
			"the 100 sentinel is no rate on a weapon proc: the game casts it off the hit, with no condition to meet"},
		{2100, false, nil, "the same sentinel on an aura, where the condition is the mask's"},
		{2600, false, []string{"named ability", "ProcTypeMask DEATH"},
			"a trigger restricted to one ability, and a bit the decoder models nothing for"},
		{2700, false, []string{"the trigger carries a class mask"},
			"an item worn by every class cannot carry one class's filter"},
		{2800, false, []string{"the trigger carries a class mask"},
			"a family no class owns still names a set of spells the listener cannot be narrowed to"},
		{2950, false, nil, "a block is an outcome the listener can be given, and the mask hears the melee hits it resolves from"},
	} {
		got := ItemProcUnsupportedWithRate(Find(tc.spellID), tc.isWeaponProc, 0, 0)
		if !slices.Equal(got, tc.want) {
			t.Errorf("spell %d (weapon proc %v): unsupported = %v, want %v - %s",
				tc.spellID, tc.isWeaponProc, got, tc.want, tc.why)
		}
	}

	if got := ItemProcUnsupportedWithRate(Find(999999), false, 0, 0); len(got) != 1 || got[0] != "no row in the store" {
		t.Errorf("unsupported = %v, want the missing row named", got)
	}
}

// A proc that deals its amount off its own trigger carries it on the A_PROC_TRIGGER_DAMAGE aura: that
// aura is the proc's damage effect, rolled by half its variance either side, 20 to 30 on 25 ± 0.4,
// and the row's RecoveryTime gives the listener no cooldown. The row's own damage effect stays empty.
func TestProcTriggerDamageAuraIsTheProcsDamageEffect(t *testing.T) {
	withRows(t, procRows())

	row := Find(2950)
	damage := row.ProcDamageEffect()
	if damage == NilEffect || damage.Aura != dbcenums.A_PROC_TRIGGER_DAMAGE {
		t.Fatalf("the proc's damage effect is %+v, want the A_PROC_TRIGGER_DAMAGE aura", damage)
	}
	if low, high := damage.Min(), damage.Max(); low != 20 || high != 30 {
		t.Errorf("the roll is %v to %v, want 20 to 30", low, high)
	}
	if row.DamageEffect() != NilEffect {
		t.Errorf("the row's damage effect is %+v, want none", row.DamageEffect())
	}

	trigger := ProcTrigger(testCharacter(), row, noopHandler)
	if row.Cooldown() == 0 {
		t.Fatal("the row states no RecoveryTime, so the ICD check below proves nothing")
	}
	if trigger.ICD != 0 || trigger.Outcome != core.OutcomeBlock || trigger.RequireDamageDealt {
		t.Errorf("ICD %v, outcome %v, requires damage %v; want none, a block, and no damage required",
			trigger.ICD, trigger.Outcome, trigger.RequireDamageDealt)
	}
}

// The sim refuses to register a proc whose row states no callback, and the audit has to see the same
// rows that way or the two would drift apart silently.
func TestItemProcUnsupportedCoversTheRegistrationsOwnRefusal(t *testing.T) {
	withRows(t, procRows())

	row := *Find(2000)
	row.ProcFlags = [2]uint32{0: dbcenums.PROC_FLAG_DEATH}
	if core.DecodeProcTypeMask(row.ProcFlags, row.ProcHint).Callback != core.CallbackEmpty {
		t.Fatal("the row the test is built on states a callback after all")
	}

	if !slices.Contains(ItemProcUnsupportedWithRate(&row, false, 0, 0), "no callback in the proc mask") {
		t.Error("a row the sim would not register was called supported")
	}
}

// A relic's trigger under its class's family keeps its mask as the relic's lock where the mask names a
// spell of the family in the store; under another family, on an item of no relic type, or naming no
// spell of its family, the mask is refused.
func TestRelicProcKeepsItsClassMask(t *testing.T) {
	withRows(t, append(procRows(), Spell{ID: 2993, Name: "Family Spell", ClassFlags: core.ClassFlags{Family: 7, Mask: [4]uint32{32}}}))
	row := Spell{ID: 2990, ProcChance: 10, ProcChanceSource: ProcChanceColumn,
		ProcFlags: [2]uint32{0: dbcenums.PROC_FLAG_DEAL_MELEE_SWING},
		Effects: []Effect{{SpellID: 2990, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_DUMMY,
			ClassFlags: core.ClassFlags{Family: 7, Mask: [4]uint32{32}}}}}
	classMask := []string{"the trigger carries a class mask"}

	if got := ItemProcUnsupportedWithRate(&row, false, 0, 0); !slices.Equal(got, classMask) {
		t.Errorf("on an item: unsupported = %v, want the class mask refused", got)
	}
	if got := ItemProcUnsupportedWithRate(&row, false, 0, 7); got != nil {
		t.Errorf("on a relic of the mask's family: unsupported = %v, want none", got)
	}
	if got := ItemProcUnsupportedWithRate(&row, false, 0, 11); !slices.Equal(got, classMask) {
		t.Errorf("on a relic of another family: unsupported = %v, want the class mask refused", got)
	}
	unfiled := row
	unfiled.Effects = []Effect{{SpellID: 2990, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_DUMMY,
		ClassFlags: core.ClassFlags{Family: 7, Mask: [4]uint32{1 << 20}}}}
	if got := ItemProcUnsupportedWithRate(&unfiled, false, 0, 7); !slices.Equal(got, []string{"the class mask names no spell of its family"}) {
		t.Errorf("on a relic with a mask under which no spell is filed: unsupported = %v, want it refused", got)
	}
}

// The spell a proc row fires: what its trigger aura names, the resource gain a dummy proc's tooltip
// names, and nothing where a dummy names none, two, or a spell that restores nothing.
func TestFiredSpell(t *testing.T) {
	gain := Spell{ID: 2993, Name: "Gain", Effects: []Effect{{SpellID: 2993, Type: dbcenums.E_ENERGIZE, BasePoints: 25,
		Target: [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_CASTER}}}}
	withRows(t, append(procRows(), gain))
	dummy := func(refs ...int32) *Spell {
		return &Spell{ID: 2991, RefIDs: refs, Effects: []Effect{{SpellID: 2991, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_DUMMY}}}
	}
	trigger := &Spell{ID: 2992, Effects: []Effect{{SpellID: 2992, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PROC_TRIGGER_SPELL, TriggerID: 2000}}}

	if got := trigger.FiredSpell(); got.ID != 2000 {
		t.Errorf("a trigger aura fires %d, want 2000", got.ID)
	}
	if got := dummy(2993).FiredSpell(); got.ID != 2993 {
		t.Errorf("a dummy naming one gain fires %d, want 2993", got.ID)
	}
	for name, row := range map[string]*Spell{"none": dummy(), "two": dummy(2993, 2000), "a spell restoring nothing": dummy(2000)} {
		if got := row.FiredSpell(); got != Nil {
			t.Errorf("a dummy naming %s fires %d, want nothing", name, got.ID)
		}
	}
}

// An enchant aura's procs-per-minute rate rolls on weapon hits only, so a mask of spells alone refuses
// it. An item's rate on the same mask is measured off the main hand and stands.
func TestEnchantAuraUnsupported(t *testing.T) {
	withRows(t, procRows())

	if got := EnchantAuraUnsupported(Find(2400)); got != nil {
		t.Errorf("unsupported = %v, want none for a rate on melee hits", got)
	}

	spellsOnly := *Find(2400)
	spellsOnly.ProcFlags = [2]uint32{0: dbcenums.PROC_FLAG_DEAL_HARMFUL_SPELL}
	if got := EnchantAuraUnsupported(&spellsOnly); !slices.Equal(got, []string{ReasonPPMHearsNoWeaponHits}) {
		t.Errorf("unsupported = %v, want the rate refused on a mask of spells", got)
	}
	if got := ItemProcUnsupportedWithRate(&spellsOnly, false, 0, 0); got != nil {
		t.Errorf("unsupported = %v, want an item's rate on spells to stand", got)
	}
}
