package database

// The override table is hand-kept data with a rule per field: the generator has to refuse a value
// the game files have started stating themselves, and a table nobody can regenerate is a table nobody
// trusts. These run the merge on synthetic rows, so they need no game files database.

import (
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/tools/database/overrides"
)

func TestOverridesBakeOntoTheRow(t *testing.T) {
	rows := []storeSpell{
		{ID: 10, Effects: []storeEffect{
			{Index: 0, Type: dbcenums.E_SCHOOL_DAMAGE},
			{Index: 1, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PERIODIC_DAMAGE},
			{Index: 2, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE},
		}},
	}

	list := []overrides.Override{
		{SpellID: 10, Field: overrides.PPM, Value: 1.5, Reason: "measured"},
		{SpellID: 10, Field: overrides.FlatThreat, Value: 200, Reason: "measured"},
		{SpellID: 10, Field: overrides.APCoefDirect, Value: 0.2, Reason: "measured"},
		{SpellID: 10, Field: overrides.APCoefPeriodic, Value: 0.1, Reason: "measured"},
		{SpellID: 10, Field: overrides.SPCoefDirect, Value: 0.3, Reason: "measured"},
		{SpellID: 10, Field: overrides.DurationMs, Value: 15000, Reason: "measured"},
		{SpellID: 10, Field: overrides.ICDMs, Value: 10000, Reason: "measured"},
		{SpellID: 10, Field: overrides.Hint, Value: float64(core.ProcHintCrit), Reason: "measured"},
		{SpellID: 10, Field: overrides.StatMask, Value: 16, Reason: "tooltip"},
	}
	if err := applyOverrides(rows, list); err != nil {
		t.Fatalf("baking the overrides: %v", err)
	}

	row := rows[0]
	if row.RPPM != 1.5 || row.FlatThreat != 200 || row.DurationMs != 15000 || row.ICDMs != 10000 {
		t.Errorf("the row reads RPPM %v, flat threat %v, duration %d, internal cooldown %d",
			row.RPPM, row.FlatThreat, row.DurationMs, row.ICDMs)
	}
	if row.Effects[0].APCoef != 0.2 || row.Effects[1].APCoef != 0.1 {
		t.Errorf("the effects read %v direct and %v periodic attack power",
			row.Effects[0].APCoef, row.Effects[1].APCoef)
	}
	if row.Effects[0].SPCoef != 0.3 || row.Effects[1].SPCoef != 0 {
		t.Errorf("the effects read %v direct and %v periodic spell power",
			row.Effects[0].SPCoef, row.Effects[1].SPCoef)
	}
	if row.ProcChanceSource != procChancePPM {
		t.Errorf("a PPM override leaves the chance source at %s", row.ProcChanceSource)
	}
	if row.ProcHint != core.ProcHintCrit {
		t.Errorf("the row reads hint %d", row.ProcHint)
	}
	if row.Effects[2].Misc2 != 16 {
		t.Errorf("the stat percentage reads mask %d", row.Effects[2].Misc2)
	}
	if len(row.overrideNotes) != len(list) {
		t.Errorf("the row states %d override reasons for %d overrides", len(row.overrideNotes), len(list))
	}
}

func TestThreatAPCoefficientBakesOntoTheThreatEffect(t *testing.T) {
	rows := []storeSpell{{ID: 10, Effects: []storeEffect{
		{Index: 0, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_RESISTANCE_PCT},
		{Index: 1, Type: dbcenums.E_THREAT},
	}}}
	list := []overrides.Override{{SpellID: 10, Field: overrides.APCoefThreat, Value: 0.05, Reason: "measured"}}

	if err := applyOverrides(rows, list); err != nil {
		t.Fatalf("baking the override: %v", err)
	}
	if rows[0].Effects[0].APCoef != 0 || rows[0].Effects[1].APCoef != 0.05 {
		t.Errorf("the effects read %v and %v attack power", rows[0].Effects[0].APCoef, rows[0].Effects[1].APCoef)
	}
}

func TestThreatPerSpellLevelBakesTheRanksFlatThreat(t *testing.T) {
	rows := []storeSpell{{ID: 10, SpellLevel: 16}, {ID: 11, SpellLevel: 24}}
	list := []overrides.Override{
		{SpellID: 10, Field: overrides.ThreatPerSpellLevel, Value: 2.5, Reason: "measured"},
		{SpellID: 11, Field: overrides.ThreatPerSpellLevel, Value: 2.5, Reason: "measured"},
	}

	if err := applyOverrides(rows, list); err != nil {
		t.Fatalf("baking the overrides: %v", err)
	}
	if rows[0].FlatThreat != 40 || rows[1].FlatThreat != 60 {
		t.Errorf("the ranks read %v and %v flat threat", rows[0].FlatThreat, rows[1].FlatThreat)
	}
}

// A proc chance override states the percentage the tooltip does not, and the store keeps the column
// as the source of it.
func TestProcChanceOverrideBecomesTheColumn(t *testing.T) {
	rows := []storeSpell{{ID: 10, ProcChanceSource: procChanceAlways, ProcChance: 101}}
	list := []overrides.Override{{SpellID: 10, Field: overrides.ProcChancePct, Value: 15, Reason: "measured"}}

	if err := applyOverrides(rows, list); err != nil {
		t.Fatalf("baking the override: %v", err)
	}
	if rows[0].ProcChance != 15 || rows[0].ProcChanceSource != procChanceColumn {
		t.Errorf("the row reads chance %d from %s", rows[0].ProcChance, rows[0].ProcChanceSource)
	}
}

func TestOverridesRefuseWhatTheGameFilesState(t *testing.T) {
	for _, tc := range []struct {
		name  string
		rows  []storeSpell
		list  []overrides.Override
		wants string
	}{
		{
			name:  "a spell the store does not carry",
			rows:  []storeSpell{{ID: 10}},
			list:  []overrides.Override{{SpellID: 11, Field: overrides.PPM, Value: 1, Reason: "measured"}},
			wants: "which the store does not carry",
		},
		{
			name:  "no reason",
			rows:  []storeSpell{{ID: 10}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.PPM, Value: 1}},
			wants: "states no reason",
		},
		{
			name: "the same field twice",
			rows: []storeSpell{{ID: 10}},
			list: []overrides.Override{
				{SpellID: 10, Field: overrides.PPM, Value: 1, Reason: "measured"},
				{SpellID: 10, Field: overrides.PPM, Value: 2, Reason: "measured again"},
			},
			wants: "two PPM overrides",
		},
		{
			name:  "a game files procs-per-minute row",
			rows:  []storeSpell{{ID: 10, procsPerMinuteID: 7}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.PPM, Value: 1, Reason: "measured"}},
			wants: "now states SpellProcsPerMinuteID 7",
		},
		{
			name:  "a game files threat effect",
			rows:  []storeSpell{{ID: 10, Effects: []storeEffect{{Index: 1, Type: dbcenums.E_THREAT}}}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.FlatThreat, Value: 200, Reason: "measured"}},
			wants: "now states threat on effect 1",
		},
		{
			name: "a game files attack-power coefficient",
			rows: []storeSpell{{ID: 10, Effects: []storeEffect{
				{Index: 0, Type: dbcenums.E_SCHOOL_DAMAGE, APCoef: 0.5},
			}}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.APCoefDirect, Value: 0.2, Reason: "measured"}},
			wants: "now states 0.5 attack power on its direct effect",
		},
		{
			name: "a game files spell-power coefficient",
			rows: []storeSpell{{ID: 10, Effects: []storeEffect{
				{Index: 0, Type: dbcenums.E_SCHOOL_DAMAGE, SPCoef: 0.5},
			}}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.SPCoefDirect, Value: 0.2, Reason: "measured"}},
			wants: "now states 0.5 spell power on its direct effect",
		},
		{
			name:  "no direct effect to carry a spell-power coefficient",
			rows:  []storeSpell{{ID: 10}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.SPCoefDirect, Value: 0.2, Reason: "measured"}},
			wants: "no direct effect to carry a spell-power coefficient",
		},
		{
			name:  "no effect to carry the coefficient",
			rows:  []storeSpell{{ID: 10}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.APCoefPeriodic, Value: 0.2, Reason: "measured"}},
			wants: "no periodic effect",
		},
		{
			name: "a game files threat coefficient",
			rows: []storeSpell{{ID: 10, Effects: []storeEffect{
				{Index: 1, Type: dbcenums.E_THREAT, APCoef: 0.1},
			}}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.APCoefThreat, Value: 0.05, Reason: "measured"}},
			wants: "now states 0.1 attack power on its threat effect",
		},
		{
			name:  "no threat effect to carry the coefficient",
			rows:  []storeSpell{{ID: 10}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.APCoefThreat, Value: 0.05, Reason: "measured"}},
			wants: "no threat effect",
		},
		{
			name:  "a per-level threat on a game files threat effect",
			rows:  []storeSpell{{ID: 10, SpellLevel: 16, Effects: []storeEffect{{Index: 1, Type: dbcenums.E_THREAT}}}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.ThreatPerSpellLevel, Value: 2.5, Reason: "measured"}},
			wants: "now states threat on effect 1",
		},
		{
			name: "a flat and a per-level threat on one spell",
			rows: []storeSpell{{ID: 10, SpellLevel: 16}},
			list: []overrides.Override{
				{SpellID: 10, Field: overrides.FlatThreat, Value: 200, Reason: "measured"},
				{SpellID: 10, Field: overrides.ThreatPerSpellLevel, Value: 2.5, Reason: "measured"},
			},
			wants: "two flat threat overrides",
		},
		{
			name:  "a chance the tooltip states",
			rows:  []storeSpell{{ID: 10, tooltipStatesChance: true}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.ProcChancePct, Value: 15, Reason: "measured"}},
			wants: "states its own proc chance in the tooltip",
		},
		{
			name:  "a chance that is no percentage",
			rows:  []storeSpell{{ID: 10}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.ProcChancePct, Value: 150, Reason: "measured"}},
			wants: "no whole percentage",
		},
		{
			name:  "a game files duration",
			rows:  []storeSpell{{ID: 10, DurationMs: 8000}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.DurationMs, Value: 15000, Reason: "measured"}},
			wants: "now states a duration of 8000 ms",
		},
		{
			name:  "a game files internal cooldown",
			rows:  []storeSpell{{ID: 10, ICDMs: 3000}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.ICDMs, Value: 10000, Reason: "measured"}},
			wants: "now states an internal cooldown of 3000 ms",
		},
		{
			name: "a game files stat",
			rows: []storeSpell{{ID: 10, Effects: []storeEffect{
				{Index: 0, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE, Misc2: 8},
			}}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.StatMask, Value: 16, Reason: "tooltip"}},
			wants: "now states misc 0 and 8",
		},
		{
			name:  "no stat percentage to carry the mask",
			rows:  []storeSpell{{ID: 10}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.StatMask, Value: 16, Reason: "tooltip"}},
			wants: "no A_MOD_TOTAL_STAT_PERCENTAGE effect",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := applyOverrides(tc.rows, tc.list)
			if err == nil {
				t.Fatalf("the merge accepted %s", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wants) {
				t.Errorf("the error reads %q, which does not name %q", err, tc.wants)
			}
		})
	}
}

// Every shipped override has to name a field the merge knows, so a row added with a field the merge
// forgot fails here rather than at the next regeneration.
func TestShippedOverridesAreWellFormed(t *testing.T) {
	seen := map[int32]bool{}
	for _, o := range overrides.Spells {
		if o.Reason == "" || o.Source == "" {
			t.Errorf("spell %d's %s override states reason %q and source %q", o.SpellID, o.Field, o.Reason, o.Source)
		}
		if o.Field.String() == "unknown field" {
			t.Errorf("spell %d carries an override of field %d", o.SpellID, uint8(o.Field))
		}
		seen[o.SpellID] = true
	}
	if len(seen) == 0 {
		t.Error("the override table is empty, so nothing above checked anything")
	}

	for _, extra := range overrides.ExtraSpells {
		if extra.Reason == "" || extra.Source == "" {
			t.Errorf("extra spell %d states reason %q and source %q", extra.SpellID, extra.Reason, extra.Source)
		}
	}

	for _, link := range overrides.HandTriggers {
		if link.Reason == "" || link.Source == "" {
			t.Errorf("the hand trigger %d -> %d states reason %q and source %q",
				link.Spell, link.Triggers, link.Reason, link.Source)
		}
	}

	for _, proc := range overrides.WeaponProcs {
		if proc.Reason == "" || proc.Source == "" {
			t.Errorf("item %d's rate on %d states reason %q and source %q", proc.ItemID, proc.SpellID, proc.Reason, proc.Source)
		}
	}

	for _, proc := range overrides.EnchantProcs {
		if proc.Reason == "" || proc.Source == "" || proc.Kind.String() == "unknown rate kind" || proc.Value <= 0 {
			t.Errorf("enchant %d's rate on %d is a %s of %v and states reason %q and source %q",
				proc.EnchantID, proc.SpellID, proc.Kind, proc.Value, proc.Reason, proc.Source)
		}
	}
}

func TestAreaBonusesBakeOntoTheRow(t *testing.T) {
	rows := []storeSpell{{ID: 10}, {ID: 11}}
	list := []overrides.AreaBonus{
		{SpellID: 10, Groups: []int32{9163, 9202}, Multiplier: 2, DurationMultiplier: 1, Reason: "tooltip", Source: "tooltip"},
	}
	if err := applyAreaBonuses(rows, list); err != nil {
		t.Fatalf("%v", err)
	}
	if got := rows[0]; len(got.AreaBonusGroups) != 2 || got.AreaBonusGroups[1] != 9202 || got.AreaMultiplier != 2 || got.AreaDurationMultiplier != 1 {
		t.Errorf("row 10 carries %v x%v (duration x%v)", got.AreaBonusGroups, got.AreaMultiplier, got.AreaDurationMultiplier)
	}
	if len(rows[0].overrideNotes) != 1 || !strings.Contains(rows[0].overrideNotes[0], "AreaBonus x2") {
		t.Errorf("row 10's notes read %v", rows[0].overrideNotes)
	}
	if rows[1].AreaMultiplier != 0 || len(rows[1].overrideNotes) != 0 {
		t.Errorf("row 11 was written to: %v", rows[1])
	}

	for _, tc := range []struct {
		name  string
		rows  []storeSpell
		list  []overrides.AreaBonus
		wants string
	}{
		{
			name:  "no reason",
			rows:  []storeSpell{{ID: 10}},
			list:  []overrides.AreaBonus{{SpellID: 10, Groups: []int32{9161}, Multiplier: 2, DurationMultiplier: 1}},
			wants: "states no reason",
		},
		{
			name:  "a spell the store does not carry",
			rows:  []storeSpell{{ID: 10}},
			list:  []overrides.AreaBonus{{SpellID: 12, Groups: []int32{9161}, Multiplier: 2, DurationMultiplier: 1, Reason: "tooltip"}},
			wants: "which the store does not carry",
		},
		{
			name: "two bonuses on one spell",
			rows: []storeSpell{{ID: 10}},
			list: []overrides.AreaBonus{
				{SpellID: 10, Groups: []int32{9161}, Multiplier: 2, DurationMultiplier: 1, Reason: "tooltip"},
				{SpellID: 10, Groups: []int32{9165}, Multiplier: 3, DurationMultiplier: 1, Reason: "tooltip"},
			},
			wants: "two area bonuses",
		},
		{
			name:  "no area group",
			rows:  []storeSpell{{ID: 10}},
			list:  []overrides.AreaBonus{{SpellID: 10, Multiplier: 2, DurationMultiplier: 1, Reason: "tooltip"}},
			wants: "names no area group",
		},
		{
			name:  "a row that now states the group itself",
			rows:  []storeSpell{{ID: 10, RequiredAreas: 9161}},
			list:  []overrides.AreaBonus{{SpellID: 10, Groups: []int32{9161}, Multiplier: 2, DurationMultiplier: 1, Reason: "tooltip"}},
			wants: "now states area group 9161 itself",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := applyAreaBonuses(tc.rows, tc.list)
			if err == nil {
				t.Fatalf("the merge accepted %s", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wants) {
				t.Errorf("the error reads %q, which does not name %q", err, tc.wants)
			}
		})
	}
}

func TestShippedAreaBonusesAreWellFormed(t *testing.T) {
	if len(overrides.AreaBonuses) == 0 {
		t.Fatal("the area bonus table is empty, so nothing below checked anything")
	}
	for _, b := range overrides.AreaBonuses {
		if b.Reason == "" || b.Source == "" {
			t.Errorf("spell %d's area bonus states reason %q and source %q", b.SpellID, b.Reason, b.Source)
		}
		if len(b.Groups) == 0 || b.Multiplier <= 0 || b.DurationMultiplier <= 0 {
			t.Errorf("spell %d's area bonus reads %v x%v (duration x%v)", b.SpellID, b.Groups, b.Multiplier, b.DurationMultiplier)
		}
	}
}
