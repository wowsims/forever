package database

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/database/overrides"
)

// A chance-on-hit item whose spell is a heal over time on the wearer routes to the heal shape. Julie's
// Dagger registers on the rate its override gives it; Reforged Spear has none, and while the game files
// state no rate for it stays commented out.
func TestChanceOnHitHotRoutesAsAHeal(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		itemID  int
		spellID int32
	}{
		{6660, 8348},      // Julie's Dagger: Julie's Blessing
		{275645, 1297357}, // Reforged Spear
	} {
		item := instance.Items[tc.itemID]
		parsed := item.ToUIItem()
		effects := dbc.MergeItemEffectsForAllStates(parsed)
		i := slices.IndexFunc(effects, func(e *proto.ItemEffect) bool { return e.BuffId == tc.spellID })
		if i < 0 {
			t.Fatalf("item %d carries no effect on %d", tc.itemID, tc.spellID)
		}

		groups := map[string]Group{}
		wantUnsupported, ok := rateOutcome(t, tc.itemID, tc.spellID, TryParseProcEffect(parsed, effects[i], instance, groups))
		if !ok {
			continue
		}
		entry := groups["Procs"].Entries[0]
		r := entry.Proc
		if r.Shape != ShapeHeal || !r.IsWeaponProc || r.TriggerSpellID != int(tc.spellID) || r.BuffSpellID != int(tc.spellID) {
			t.Errorf("item %d: shape %v, weapon proc %v, trigger %d, buff %d; want a weapon proc healing through %d",
				tc.itemID, r.Shape, r.IsWeaponProc, r.TriggerSpellID, r.BuffSpellID, tc.spellID)
		}
		if !slices.Equal(r.Unsupported, wantUnsupported) {
			t.Errorf("item %d refused for %q, want %q", tc.itemID, r.Unsupported, wantUnsupported)
		}

		e := spelldata.Find(tc.spellID).ProcHealEffect()
		if e.Aura != dbcenums.A_PERIODIC_HEAL || e.Target[0] != dbcenums.TARGET_UNIT_CASTER {
			t.Errorf("%d heals through aura %v on target %d; want A_PERIODIC_HEAL on the caster", tc.spellID, e.Aura, e.Target[0])
		}
	}
}

// What a weapon's chance-on-hit parse must answer. A rate from a weapon rate override, or one the
// trigger's row states, registers it unlisted. With neither it is refused for the missing rate alone
// and listed as missing. A row's 100 or 101 is no rate on a weapon proc: the game casts the spell off
// the hit without consulting the row. Answers the refusal reasons to expect, or false after reporting
// a parse that answered otherwise.
func rateOutcome(t *testing.T, itemID int, spellID int32, got EffectParseResult) ([]string, bool) {
	t.Helper()
	row := spelldata.Find(spellID)
	rated := overrides.WeaponProcPPM(int32(itemID), spellID) > 0
	stated := row.RPPM > 0 || (row.ProcChanceSource != spelldata.ProcChanceAlways && row.StatedChance() != 0)
	_, listed := missingEffectsMap["ItemEffects"][int32(itemID)]
	switch {
	case (rated || stated) && got == EffectParseResultSuccess && !listed:
		return nil, true
	case !rated && !stated && got == EffectParseResultRefused && listed:
		return []string{spelldata.ReasonStatesNoRate}, true
	}
	t.Errorf("item %d parsed as %v, listed as missing %v; rated by an override %v, by its row %v", itemID, got, listed, rated, stated)
	return nil, false
}

// A chance-on-hit spell that grants extra attacks routes to the extra-attack shape: Flurry Axe 871's
// 18797 grants the wearer extra attacks, at the override's 1.9 PPM, so it registers.
func TestChanceOnHitExtraAttackRoutes(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	const itemID, spellID = 871, 18797
	parsed, effect := parsedItem(t, instance, itemID, spellID)

	groups := map[string]Group{}
	if got := TryParseProcEffect(parsed, effect, instance, groups); got != EffectParseResultSuccess {
		t.Fatalf("item %d parsed as %v, want %v", itemID, got, EffectParseResultSuccess)
	}
	r := groups["Procs"].Entries[0].Proc
	if r.Shape != ShapeExtraAttack || !r.IsWeaponProc || r.TriggerSpellID != spellID || r.BuffSpellID != 0 || len(r.Unsupported) != 0 {
		t.Errorf("item %d: shape %v, weapon proc %v, trigger %d, buff %d, unsupported %q; want a weapon proc granting extra attacks through %d",
			itemID, r.Shape, r.IsWeaponProc, r.TriggerSpellID, r.BuffSpellID, r.Unsupported, spellID)
	}
}

// A chance-on-hit spell dealing damage over time alone routes to the damage shape, cast through its own
// row. An item the weapon rate overrides give a rate registers on it; any other is refused for the
// missing rate alone, while the game files state none, and listed as missing. The same rows at a rate
// on the item register.
func TestChanceOnHitDamageOverTimeRoutesAsDamage(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		itemID  int
		spellID int32
	}{
		{10772, 18075},    // Glutton's Cleaver: a bleed
		{12777, 16898},    // Blazing Rapier: Fire
		{12795, 13318},    // Blood Talon: Rend, a bleed
		{13035, 17511},    // Serpent Slicer: a poison
		{17002, 17511},    // Ichor Spitter: the same poison
		{15814, 16406},    // Hameya's Slayer: a bleed
		{274963, 1295744}, // Rot-Covered Harpoon: a disease
		{282561, 1314040}, // Siegebreaker's Blaster: Fire
		{285102, 1317583}, // Darkmist Dirk: a poison
	} {
		s := spelldata.Find(tc.spellID)
		if s.DamageEffect() != spelldata.NilEffect || s.PeriodicDamageEffect() == spelldata.NilEffect {
			t.Errorf("%d deals direct damage or none over time", tc.spellID)
			continue
		}

		parsed, effect := parsedItem(t, instance, tc.itemID, tc.spellID)
		groups := map[string]Group{}
		wantUnsupported, ok := rateOutcome(t, tc.itemID, tc.spellID, TryParseProcEffect(parsed, effect, instance, groups))
		if !ok {
			continue
		}
		r := groups["Procs"].Entries[0].Proc
		if r.Shape != ShapeDamage || !r.IsWeaponProc || r.TriggerSpellID != int(tc.spellID) || r.BuffSpellID != int(tc.spellID) ||
			r.ProcConstructor() != "NewSpellDataDamageProc" {
			t.Errorf("item %d: shape %v, weapon proc %v, trigger %d, buff %d; want a weapon proc dealing %d's damage",
				tc.itemID, r.Shape, r.IsWeaponProc, r.TriggerSpellID, r.BuffSpellID, tc.spellID)
		}
		if !slices.Equal(r.Unsupported, wantUnsupported) {
			t.Errorf("item %d refused for %q, want %q", tc.itemID, r.Unsupported, wantUnsupported)
		}

		rated := routeProc(int(tc.spellID), int(tc.spellID), true, 1)
		rated.as(ShapeDamage, tc.spellID)
		if !rated.Supported() {
			t.Errorf("item %d at 1 PPM refused for %q, want registered", tc.itemID, rated.Unsupported)
		}
	}
}

// A trigger's ProcChance column reaches the item's proc only where it is a roll, below 100: Green Whelp
// Armor 7375's 9160 rolls its column, while Storm Gauntlets 12632's 16615 states the 100 or more that
// fires on its own condition and leaves the proc stating no chance.
func TestItemProcCarriesOnlyARolledChance(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		itemID  int
		buffID  int32
		trigger int
		rolls   bool
	}{
		{7375, 9159, 9160, true},
		{12632, 16614, 16615, false},
	} {
		column := float64(instance.Spells[tc.trigger].ProcChance)
		if rolls := column > 0 && column < 100; rolls != tc.rolls {
			t.Fatalf("%d states ProcChance %v, outside the case the test covers", tc.trigger, column)
		}
		want := 0.0
		if tc.rolls {
			want = column / 100
		}
		_, effect := parsedItem(t, instance, tc.itemID, tc.buffID)
		if got := effect.GetProc().GetProcChance(); got != want {
			t.Errorf("item %d's proc states chance %v, want %v", tc.itemID, got, want)
		}
	}
}

// Two weapons sharing a trigger spell each carry their own rate on their item effect: The Needler and
// Hanzo Sword both cast Wound 16405, at 3 and 1 PPM, and each registers on its own.
func TestWeaponsSharingATriggerKeepTheirOwnRate(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		itemID int
		ppm    float64
	}{
		{13060, 3},
		{8190, 1},
	} {
		parsed, effect := parsedItem(t, instance, tc.itemID, 16405)
		if got := effect.GetProc().GetPpm(); got != tc.ppm {
			t.Errorf("item %d's effect states %v PPM, want %v", tc.itemID, got, tc.ppm)
		}

		groups := map[string]Group{}
		if got := TryParseProcEffect(parsed, effect, instance, groups); got != EffectParseResultSuccess {
			t.Errorf("item %d parsed as %v, want %v", tc.itemID, got, EffectParseResultSuccess)
			continue
		}
		r := groups["Procs"].Entries[0].Proc
		if r.ItemPPM != tc.ppm || len(r.Unsupported) != 0 || !strings.Contains(r.Summary, fmt.Sprintf("%v ppm on the item", tc.ppm)) {
			t.Errorf("item %d routed with %v PPM, unsupported %q, summary %q", tc.itemID, r.ItemPPM, r.Unsupported, r.Summary)
		}
	}
}

// Every hand-kept weapon rate names an effect the game files hang on that item for that spell, and
// the database writes the rate onto the item: a mistyped id would otherwise leave the weapon refused
// as stating no rate, with the rate sitting unread.
func TestWeaponProcRowsNameAnItemEffect(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, row := range overrides.WeaponProcs {
		if !slices.ContainsFunc(instance.ItemEffectsByParentID[int(row.ItemID)], func(e dbc.ItemEffect) bool {
			return e.SpellID == int(row.SpellID)
		}) {
			t.Errorf("item %d carries no effect casting %d", row.ItemID, row.SpellID)
			continue
		}

		item := instance.Items[int(row.ItemID)]
		parsed := item.ToUIItem()
		effects := dbc.MergeItemEffectsForAllStates(parsed)
		if !slices.ContainsFunc(effects, func(e *proto.ItemEffect) bool { return e.GetProc().GetPpm() == row.PPM }) {
			t.Errorf("item %d's effects state no %v PPM", row.ItemID, row.PPM)
		}
	}
}

// An on-use with no stats routes from the spell it casts: damage on the enemy it is used on or a heal
// on the wearer or a friendly target registers, a heal on anyone else or a spell that does neither is
// refused with the reason and listed.
func TestOnUseRoutesFromTheSpellItCasts(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		itemID  int
		spellID int32
		want    EffectParseResult
		damage  bool
		heal    bool
		reasons []string
	}{
		{219345, 443265, EffectParseResultSuccess, true, false, nil}, // Infernal Lasso: A_PERIODIC_DAMAGE, and a root it leaves out
		{16768, 20631, EffectParseResultSuccess, false, true, nil},   // Furbolg Medicine Pouch: A_PERIODIC_HEAL on the wearer
		{18637, 23064, EffectParseResultSuccess, false, true, nil},   // Major Recombobulator: E_HEAL on a friendly target
		{14153, 18386, EffectParseResultRefused, false, true,
			[]string{"the heal lands on implicit target 5, not the wearer"}}, // Robe of the Void, on the pet
		{7734, 14537, EffectParseResultRefused, false, false,
			[]string{"14537 deals no damage and heals no one (E_DUMMY)"}}, // Six Demon Bag
	} {
		item := instance.Items[tc.itemID]
		parsed := item.ToUIItem()
		effects := dbc.MergeItemEffectsForAllStates(parsed)
		i := slices.IndexFunc(effects, func(e *proto.ItemEffect) bool { return e.BuffId == tc.spellID && e.GetOnUse() != nil })
		if i < 0 {
			t.Fatalf("item %d carries no on-use effect on %d", tc.itemID, tc.spellID)
		}

		groups := map[string]Group{}
		if got := TryParseOnUseEffect(parsed, effects[i], instance, groups); got != tc.want {
			t.Errorf("item %d parsed as %v, want %v", tc.itemID, got, tc.want)
			continue
		}

		var entry *Entry
		for _, grp := range groups {
			entry = grp.Entries[0]
		}
		r := entry.Proc
		if r == nil || r.TriggerSpellID != int(tc.spellID) || (r.Shape == ShapeDamage) != tc.damage || (r.Shape == ShapeHeal) != tc.heal ||
			entry.Supported != (tc.want == EffectParseResultSuccess) {
			t.Errorf("item %d: routing %+v, supported %v; want spell %d, damage %v, heal %v", tc.itemID, r, entry.Supported, tc.spellID, tc.damage, tc.heal)
			continue
		}
		if !slices.Equal(r.Unsupported, tc.reasons) {
			t.Errorf("item %d refused for %q, want %q", tc.itemID, r.Unsupported, tc.reasons)
		}
	}
}

// An on-use whose spell raises the wearer's speeds registers as a speed buff, every speed effect of
// its row modelled. Manual Crowd Pummeler is below MIN_EFFECT_ILVL and registers through its
// EffectFloorExemptions entry.
func TestOnUseSpeedBuffRoutesAsASpeedBuff(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		itemID  int
		spellID int32
		summary string
	}{
		{19339, 23723, "on use: 23723 (A_MOD_CASTING_SPEED_NOT_STACK)"},                      // Mind Quickening Gem
		{19343, 23733, "on use: 23733 (A_MOD_CASTING_SPEED_NOT_STACK, A_MOD_MELEE_HASTE_3)"}, // Scrolls of Blinding Light
		{22954, 28866, "on use: 28866 (A_MOD_MELEE_HASTE_3, A_MOD_RANGED_HASTE)"},            // Kiss of the Spider
		{9449, 13494, "on use: 13494 (A_MOD_MELEE_HASTE_3)"},                                 // Manual Crowd Pummeler
	} {
		item := instance.Items[tc.itemID]
		parsed := item.ToUIItem()
		effects := dbc.MergeItemEffectsForAllStates(parsed)
		i := slices.IndexFunc(effects, func(e *proto.ItemEffect) bool { return e.BuffId == tc.spellID && e.GetOnUse() != nil })
		if i < 0 {
			t.Fatalf("item %d carries no on-use effect on %d", tc.itemID, tc.spellID)
		}

		groups := map[string]Group{}
		if got := TryParseOnUseEffect(parsed, effects[i], instance, groups); got != EffectParseResultSuccess {
			t.Errorf("item %d parsed as %v, want registered", tc.itemID, got)
			continue
		}

		entries := groups["Speed"].Entries
		if len(entries) != 1 {
			t.Fatalf("item %d: %d entries in the Speed group, want 1", tc.itemID, len(entries))
		}
		r := entries[0].Proc
		if !entries[0].Supported || r.Shape != ShapeSpeed || r.TriggerSpellID != int(tc.spellID) ||
			r.OnUseConstructor() != "NewSpellDataSpeedOnUse" || r.Summary != tc.summary {
			t.Errorf("item %d: routing %+v, supported %v; want %d registered through NewSpellDataSpeedOnUse with %q",
				tc.itemID, r, entries[0].Supported, tc.spellID, tc.summary)
		}
	}
}

// An on-use whose spell keeps an A_PERIODIC_TRIGGER_SPELL aura on the wearer, casting a row that deals
// damage in an area or cone, registers as periodic damage, every period of the aura for the spell's
// duration: Thunderbrew's Boot Flask's 12257 casts the cone 12278 and leaves its E_TRIGGER_SPELL of
// Strong Alcohol 11009 out; Cloak of Fire's 18364 casts the area 18374.
func TestOnUsePeriodicTriggerDamageRoutesAsPeriodicDamage(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		itemID      int
		spellID     int32
		onUse, cast string
	}{
		{744, 12257, "on use: 12257 (A_PERIODIC_TRIGGER_SPELL); not simulated: E_TRIGGER_SPELL", "12278 (E_SCHOOL_DAMAGE)"},
		{14134, 18364, "on use: 18364 (A_PERIODIC_TRIGGER_SPELL)", "18374 (E_SCHOOL_DAMAGE)"},
	} {
		row := spelldata.MustFind(tc.spellID)
		period := row.FirstAura(dbcenums.A_PERIODIC_TRIGGER_SPELL).Period()
		if period <= 0 || row.Duration() <= 0 {
			t.Fatalf("%d states a period of %v for %v, want both", tc.spellID, period, row.Duration())
		}
		summary := fmt.Sprintf("%s; every %v for %v: %s", tc.onUse, period, row.Duration(), tc.cast)

		item := instance.Items[tc.itemID]
		parsed := item.ToUIItem()
		effects := dbc.MergeItemEffectsForAllStates(parsed)
		i := slices.IndexFunc(effects, func(e *proto.ItemEffect) bool { return e.BuffId == tc.spellID && e.GetOnUse() != nil })
		if i < 0 {
			t.Fatalf("item %d carries no on-use effect on %d", tc.itemID, tc.spellID)
		}

		groups := map[string]Group{}
		if got := TryParseOnUseEffect(parsed, effects[i], instance, groups); got != EffectParseResultSuccess {
			t.Errorf("item %d parsed as %v, want registered", tc.itemID, got)
			continue
		}

		entries := groups["Damage"].Entries
		if len(entries) != 1 {
			t.Fatalf("item %d: %d entries in the Damage group, want 1", tc.itemID, len(entries))
		}
		r := entries[0].Proc
		if !entries[0].Supported || r.Shape != ShapePeriodicDamage || r.TriggerSpellID != int(tc.spellID) ||
			r.OnUseConstructor() != "NewSpellDataPeriodicDamageOnUse" || r.Summary != summary {
			t.Errorf("item %d: routing %+v, supported %v; want %d registered through NewSpellDataPeriodicDamageOnUse with %q",
				tc.itemID, r, entries[0].Supported, tc.spellID, summary)
		}
	}
}

// An on-use whose spell applies a hit or crit percent, which no item stat holds, registers as an
// aura: Gri'lek's Charm of Valor's 24498 is spell crit alone, and Weakness Analyzer's 1291101 drops at
// the first crit, a charge the hits its row names spend. The spell modifiers, the threat and the damage
// shield an on-use states route the same way, and a movement slow beside an enemy attack speed slow is
// left out of what it models: Cold Basilisk Eye's 1139.
func TestOnUsePercentBuffRoutesAsAnAura(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		itemID  int
		spellID int32
		summary string
		reasons []string
	}{
		{19952, 24498, "on use: 24498 (A_MOD_SPELL_CRIT_CHANCE)", nil},                                      // Gri'lek's Charm of Valor
		{272438, 1291101, "on use: 1291101 (A_MOD_CRIT_PCT)", nil},                                          // Weakness Analyzer
		{19959, 24544, "on use: 24544 (A_ADD_FLAT_MODIFIER, A_ADD_PCT_MODIFIER)", nil},                      // Hazza'rah's Charm of Magic
		{23001, 28862, "on use: 28862 (A_MOD_THREAT)", nil},                                                 // Eye of Diminution
		{21488, 26168, "on use: 26168 (A_DAMAGE_SHIELD)", nil},                                              // Fetish of Chitinous Spikes
		{13164, 17275, "on use: 17275 (A_DAMAGE_SHIELD, A_MOD_RESISTANCE)", nil},                            // Heart of the Scale
		{5079, 1139, "on use: 1139 (A_MOD_MELEE_RANGED_HASTE_2); not simulated: A_MOD_DECREASE_SPEED", nil}, // Cold Basilisk Eye
	} {
		item := instance.Items[tc.itemID]
		parsed := item.ToUIItem()
		effects := dbc.MergeItemEffectsForAllStates(parsed)
		i := slices.IndexFunc(effects, func(e *proto.ItemEffect) bool { return e.BuffId == tc.spellID && e.GetOnUse() != nil })
		if i < 0 {
			t.Fatalf("item %d carries no on-use effect on %d", tc.itemID, tc.spellID)
		}

		want, group := EffectParseResultSuccess, "Auras"
		if tc.reasons != nil {
			want, group = EffectParseResultRefused, ""
		}
		groups := map[string]Group{}
		if got := TryParseOnUseEffect(parsed, effects[i], instance, groups); got != want {
			t.Errorf("item %d parsed as %v, want %v", tc.itemID, got, want)
			continue
		}

		entries := groups[group].Entries
		if len(entries) != 1 {
			t.Fatalf("item %d: %d entries in the %q group, want 1", tc.itemID, len(entries), group)
		}
		r := entries[0].Proc
		if r.Shape != ShapeAura || r.TriggerSpellID != int(tc.spellID) || r.OnUseConstructor() != "NewSpellDataAuraOnUse" ||
			r.Summary != tc.summary || !slices.Equal(r.Unsupported, tc.reasons) {
			t.Errorf("item %d: routing %+v; want %d through NewSpellDataAuraOnUse with %q, refused for %q",
				tc.itemID, r, tc.spellID, tc.summary, tc.reasons)
		}
	}
}

// Glyph of Deflection 23040's 28773 states its block value as A_MOD_BLOCK_VALUE_FLAT, which the item
// effect carries as a stat and the flat on-use grants. Blazing Emblem 2802's 13744 absorbs Fire beside
// its Fire resistance, so it registers as an absorb carrying the resistance, and is not listed as
// missing. Heart of the Scale 13164's 17275 states a damage shield beside its Fire resistance, so it
// registers as an aura carrying both.
func TestOnUseStatBuffWithWhatItsBuffStatesBeside(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	glyph, glyphEffect := parsedItem(t, instance, 23040, 28773)
	blockValue := spelldata.MustFind(28773).FirstAura(dbcenums.A_MOD_BLOCK_VALUE_FLAT).BaseValue()
	if got := dbc.EffectStats(glyphEffect)[int32(proto.Stat_StatBlockValue)]; got != blockValue || got == 0 {
		t.Errorf("Glyph of Deflection's effect carries %v block value, want 28773's %v", got, blockValue)
	}
	groups := map[string]Group{}
	if got := TryParseOnUseEffect(glyph, glyphEffect, instance, groups); got != EffectParseResultSuccess {
		t.Errorf("Glyph of Deflection parsed as %v, want the stat on-use registered", got)
	} else if entries := groups["BlockValue"].Entries; len(entries) != 1 || !entries[0].Supported || entries[0].Proc != nil {
		t.Errorf("Glyph of Deflection's BlockValue entries %+v, want one flat on-use", entries)
	}

	emblem, emblemEffect := parsedItem(t, instance, 2802, 13744)
	groups = map[string]Group{}
	if got := TryParseOnUseEffect(emblem, emblemEffect, instance, groups); got != EffectParseResultSuccess {
		t.Fatalf("Blazing Emblem parsed as %v, want the absorb on-use registered", got)
	}
	fireResistance := spelldata.MustFind(13744).FirstAura(dbcenums.A_MOD_RESISTANCE).BaseValue()
	if got := dbc.EffectStats(emblemEffect)[int32(proto.Stat_StatFireResistance)]; got != fireResistance || got == 0 {
		t.Errorf("Blazing Emblem's effect carries %v Fire resistance, want 13744's %v", got, fireResistance)
	}
	if entries := groups["Absorbs"].Entries; len(entries) != 1 || !entries[0].Supported ||
		entries[0].Proc.OnUseConstructor() != "NewSpellDataAbsorbOnUse" ||
		entries[0].Proc.Summary != "on use: 13744 (A_SCHOOL_ABSORB, A_MOD_RESISTANCE)" {
		t.Errorf("Blazing Emblem's Absorbs entries %+v, want one absorb on-use carrying the resistance", entries)
	}
	if _, listed := missingEffectsMap["ItemEffects"][2802]; listed {
		t.Errorf("2802 is listed as missing an effect")
	}

	heart, heartEffect := parsedItem(t, instance, 13164, 17275)
	groups = map[string]Group{}
	if got := TryParseOnUseEffect(heart, heartEffect, instance, groups); got != EffectParseResultSuccess {
		t.Fatalf("Heart of the Scale parsed as %v, want registered", got)
	}
	if entries := groups["Auras"].Entries; len(entries) != 1 || !entries[0].Supported ||
		entries[0].Proc.OnUseConstructor() != "NewSpellDataAuraOnUse" {
		t.Errorf("Heart of the Scale's Auras entries %+v, want one aura on-use", entries)
	}
}

// An item registers one effect, so an equip aura beside another effect of the item that registers is
// refused and listed, and the other kept: Sulfuras 17182's damage shield beside its Fireball proc. An
// equip aura alone, Cloak of Flames 3475's, stays.
func TestEquipAuraBesideAnotherEffectIsRefused(t *testing.T) {
	equip := func(itemID int) *Entry {
		return &Entry{Variants: []*Variant{{ID: itemID, Name: "Test", SpellID: 21142}}, Supported: true,
			Proc: &ProcRouting{TriggerSpellID: 21142, Shape: ShapeEquip}}
	}
	sulfurasShield, cloakShield := equip(17182), equip(3475)
	fireball := &Entry{Variants: []*Variant{{ID: 17182, Name: "Test", SpellID: 21162}}, Supported: true,
		Proc: &ProcRouting{TriggerSpellID: 21162, Shape: ShapeDamage}}

	refuseEquipBesideAnotherEffect(map[string]Group{}, map[string]Group{
		"Equip": {Name: "Equip", Entries: []*Entry{sulfurasShield, cloakShield}},
		"Procs": {Name: "Procs", Entries: []*Entry{fireball}},
	})

	if sulfurasShield.Supported || !slices.Equal(sulfurasShield.Proc.Unsupported, []string{"another effect of the item registers"}) {
		t.Errorf("Sulfuras' damage shield supported %v, refused for %q; want refused beside its proc",
			sulfurasShield.Supported, sulfurasShield.Proc.Unsupported)
	}
	if !fireball.Supported || !cloakShield.Supported {
		t.Errorf("Sulfuras' proc supported %v, Cloak of Flames' shield supported %v; want both kept", fireball.Supported, cloakShield.Supported)
	}
	if _, listed := missingEffectsMap["ItemEffects"][17182]; !listed {
		t.Errorf("17182 is not listed as missing an effect")
	}
}

// Aegis of Preservation 19345 registers its 500 armor and lists the heal on every hit taken, 23781,
// that its buff 23780 procs.
func TestOnUseStatBuffListsTheProcItCarries(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	item := instance.Items[19345]
	parsed := item.ToUIItem()
	effects := dbc.MergeItemEffectsForAllStates(parsed)
	groups := map[string]Group{}
	if got := TryParseOnUseEffect(parsed, effects[0], instance, groups); got != EffectParseResultSuccess {
		t.Fatalf("parsed as %v, want the stat on-use registered", got)
	}

	entry := groups["Armor"].Entries[0]
	if want := "the proc the buff carries, 23781 (E_HEAL)"; !entry.Supported || entry.NotSimulated != want {
		t.Errorf("supported %v, not simulated %q; want registered with %q", entry.Supported, entry.NotSimulated, want)
	}
	if _, listed := missingEffectsMap["ItemEffects"][19345]; !listed {
		t.Errorf("19345 is not listed as missing an effect")
	}
}

// A trigger restricted to one named ability is a shape no ProcTypeMask states, so the wording is the
// only evidence there is. The rows below are the game files' own, raw and rendered: the store reads
// the raw description and the item generator the rendered tooltip, and both go through this
// matcher.
func TestNamedAbilityWording(t *testing.T) {
	for _, tc := range []struct {
		named bool
		text  string
		why   string
	}{
		{true, "Your casts of $?s2060[Greater Heal]?s5185[Healing Touch][Greater Heal, Healing Touch] in combat grant up to $1249119s1 increased healing.",
			"Eternal Power 1249118, as the game files write it"},
		{true, "Your casts of Greater Heal in combat grant up to 40 increased healing and up to 13 increased damage for 15s.",
			"the same tooltip, rendered"},
		{true, "Your Backstab has a $m1% chance to cause your next Ambush to not require Stealth.",
			`"has" after a named ability`},
		{true, "Heals from your Earth Shield have a $s1% chance to make your next cast time heal instant cast.",
			`"have" after a named ability`},
		{true, "When you cast Chain Heal or Riptide you gain $s1 charges of Tidal Waves.",
			"a cast of one named ability"},
		{true, "Your Shock spells have a chance to deal extra damage.", "the wording the matcher already read"},
		{false, "When you cast a healing spell, gain Mana equal to $m1% of the base cost of the spell.",
			"no ability is named"},
		{false, "20% chance to regain 100 mana when you cast a Judgement.",
			"a lowercase article keeps the class of spell general"},
		{false, "Multiple casts of Divine Light do not accumulate this shield.",
			"not the caster's own casts, and not a trigger clause"},
		{false, "Your melee attacks have a chance to deal extra damage.",
			"no ability is named"},
		{false, "2% chance on successful spellcast to increase your Spirit by 150 for 15s.",
			"Darkmoon Card: Blue Dragon, which fires off any cast"},
	} {
		if got := procTooltipHints(tc.text).Matches(core.ProcHintNamedAbility); got != tc.named {
			t.Errorf("named ability = %v, want %v for %s:\n  %s", got, tc.named, tc.why, tc.text)
		}
	}
}

// The wearer's own attack dodged or parried is a trigger only in a condition clause. The same words
// state a magnitude on every expertise row, and "when you parry" is an attack the wearer takes.
func TestAttackAvoidedWording(t *testing.T) {
	const avoided = core.ProcHintAttackDodged | core.ProcHintAttackParried

	for _, tc := range []struct {
		want core.ProcHint
		text string
		why  string
	}{
		{avoided, "Permanently enchant a Melee Weapon to trigger Recovery when you are Parried or Dodged, healing you for 5% of your maximum health. Cannot occur more often than once every 10 sec.",
			"Recovery's grant 1248760, rendered"},
		{core.ProcHintAttackParried, "Whenever your melee attacks are parried, gain 10 rage.",
			"one outcome named, one bit"},
		{0, "Reduces the chance for your attacks to be dodged or parried by $s1%.",
			"Increased Expertise 1213288, a magnitude"},
		{0, "Your Taunt ability never misses, and your chance to be Dodged or Parried is reduced by $s1%.",
			"the Naxxramas tank 2P 1219540, a magnitude"},
		{0, "Instantly overpower the enemy, causing weapon damage plus $s1.  Only useable after the target dodges.  The Overpower cannot be blocked, dodged or parried.",
			"Overpower 7384, which states what it cannot be"},
		{0, "Your Shield Slam deals $s1% increased threat and its cooldown is reset if it is Dodged, Parried, or Blocked.",
			"TAQ tank 4P 1214162, one named ability rather than the wearer's attacks"},
		{0, "When you parry an attack, gain 10 rage.",
			"the wearer parrying an attack it takes"},
	} {
		if got := procTooltipHints(tc.text) & avoided; got != tc.want {
			t.Errorf("hint = %q, want %q for %s:\n  %s", formatProcHint(got), formatProcHint(tc.want), tc.why, tc.text)
		}
	}
}

// An enchant whose equip spell names a melee weapon subset registers the spell as an equip aura, kept
// up while the enchant is worn. 430391 (Hit Weapon +2) states A_MOD_HIT_CHANCE on the melee weapons
// but not spears, which no enchant stat can hold; its row is freed of its area group for the test. An
// enchant that already registers a proc lists it.
func TestWeaponScopedEnchantRegistersAnEquipAura(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	const effectID, spellID = 990991, 430391
	row := spelldata.MustFind(spellID)
	areas := row.RequiredAreas
	row.RequiredAreas = 0
	t.Cleanup(func() { row.RequiredAreas = areas })
	enchant := dbc.Enchant{EffectId: effectID, Name: "Test Hit Enchant",
		Effects: []int{dbc.ITEM_ENCHANTMENT_EQUIP_SPELL}, EffectArgs: []int{spellID}}
	parsed := &proto.UIEnchant{EffectId: effectID, Name: enchant.Name}
	grants := map[int]*dbc.SpellEffect{effectID: {SpellID: spellID}}
	if dbc.EquipSpellFolds(spellID) {
		t.Fatalf("%d folds into the enchant stats; the test needs one left to run time", spellID)
	}

	groups := map[string]Group{}
	storeWeaponScopedEnchant(instance, enchant, parsed, grants, groups)
	entries := groups["Enchants"].Entries
	if len(entries) != 1 {
		t.Fatalf("%d entries in the Enchants group, want 1", len(entries))
	}
	r := entries[0].Proc
	if r.Shape != ShapeEquip || r.TriggerSpellID != spellID || r.ProcConstructor() != "NewSpellDataEquipAura" ||
		entries[0].Variants[0].ID != effectID || !entries[0].Supported {
		t.Errorf("routing %+v for enchant %d, supported %v; want %d registered through NewSpellDataEquipAura",
			r, entries[0].Variants[0].ID, entries[0].Supported, spellID)
	}

	registered := map[string]Group{"Enchants": {Name: "Enchants", Entries: []*Entry{{
		Variants: []*Variant{{ID: effectID}}, Proc: &ProcRouting{}, Supported: true}}}}
	storeWeaponScopedEnchant(instance, enchant, parsed, grants, registered)
	if second := registered["Enchants"].Entries; len(second) != 2 || second[1].Supported {
		t.Errorf("an enchant that already registers should list the equip aura, got %d entries", len(second))
	}
}
