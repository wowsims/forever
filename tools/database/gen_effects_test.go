package database

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/database/overrides"
)

// A chance-on-hit item whose spell is a heal over time on the wearer routes to the heal shape. Neither
// row states a rate: 8348's ProcChance is the 100 that means no roll, 1297357 carries no aura options
// at all, and "Chance on hit" beside neither names a percentage. Julie's Dagger registers on the
// override's 1 PPM; Reforged Spear stays commented out.
func TestChanceOnHitHotRoutesAsAHeal(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		itemID   int
		spellID  int32
		rated    bool
		perTick  float64
		duration time.Duration
	}{
		{6660, 8348, true, 13, 12 * time.Second},       // Julie's Dagger: Julie's Blessing
		{275645, 1297357, false, 26, 14 * time.Second}, // Reforged Spear
	} {
		item := instance.Items[tc.itemID]
		parsed := item.ToUIItem()
		parsed.ItemEffects = dbc.MergeItemEffectsForAllStates(parsed)
		i := slices.IndexFunc(parsed.ItemEffects, func(e *proto.ItemEffect) bool { return e.BuffId == tc.spellID })
		if i < 0 {
			t.Fatalf("item %d carries no effect on %d", tc.itemID, tc.spellID)
		}

		groups := map[string]Group{}
		wantResult, wantUnsupported := EffectParseResultRefused, []string{spelldata.ReasonStatesNoRate}
		if tc.rated {
			wantResult, wantUnsupported = EffectParseResultSuccess, nil
		}
		if got := TryParseProcEffect(parsed, parsed.ItemEffects[i], instance, groups); got != wantResult {
			t.Errorf("item %d parsed as %v, want %v", tc.itemID, got, wantResult)
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

		heal := spelldata.Find(tc.spellID)
		e := heal.ProcHealEffect()
		if e.Aura != dbcenums.A_PERIODIC_HEAL || e.Target[0] != dbcenums.TARGET_UNIT_CASTER ||
			e.Average() != tc.perTick || e.Period() != 2*time.Second || heal.Duration() != tc.duration {
			t.Errorf("%d heals through aura %v on target %d, %v every %v for %v; want A_PERIODIC_HEAL on the caster, %v every 2s for %v",
				tc.spellID, e.Aura, e.Target[0], e.Average(), e.Period(), heal.Duration(), tc.perTick, tc.duration)
		}
	}
}

// A chance-on-hit spell that grants extra attacks routes to the extra-attack shape: Flurry Axe 871's
// 18797 grants the wearer one, at the override's 1.9 PPM, so it registers.
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
	if got := spelldata.Find(spellID).ExtraAttacksEffect().BasePoints; got != 1 {
		t.Errorf("%d grants %v extra attacks, want 1", spellID, got)
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
		parsed.ItemEffects = dbc.MergeItemEffectsForAllStates(parsed)
		if !slices.ContainsFunc(parsed.ItemEffects, func(e *proto.ItemEffect) bool { return e.GetProc().GetPpm() == row.PPM }) {
			t.Errorf("item %d's effects state no %v PPM", row.ItemID, row.PPM)
		}
	}
}

// An on-use with no stats routes from the spell it casts: damage on the enemy it is used on or a heal
// on the wearer registers, a heal on anyone else or a spell that does neither is refused with the
// reason and listed.
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
		{18637, 23064, EffectParseResultRefused, false, true,
			[]string{"the heal lands on implicit target 21, not the wearer"}}, // Major Recombobulator
		{14153, 18386, EffectParseResultRefused, false, true,
			[]string{"the heal lands on implicit target 5, not the wearer"}}, // Robe of the Void, on the pet
		{7734, 14537, EffectParseResultRefused, false, false,
			[]string{"14537 deals no damage and heals no one (E_DUMMY)"}}, // Six Demon Bag
	} {
		item := instance.Items[tc.itemID]
		parsed := item.ToUIItem()
		parsed.ItemEffects = dbc.MergeItemEffectsForAllStates(parsed)
		i := slices.IndexFunc(parsed.ItemEffects, func(e *proto.ItemEffect) bool { return e.BuffId == tc.spellID && e.GetOnUse() != nil })
		if i < 0 {
			t.Fatalf("item %d carries no on-use effect on %d", tc.itemID, tc.spellID)
		}

		groups := map[string]Group{}
		if got := TryParseOnUseEffect(parsed, parsed.ItemEffects[i], instance, groups); got != tc.want {
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
// its row modelled.
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
	} {
		item := instance.Items[tc.itemID]
		parsed := item.ToUIItem()
		parsed.ItemEffects = dbc.MergeItemEffectsForAllStates(parsed)
		i := slices.IndexFunc(parsed.ItemEffects, func(e *proto.ItemEffect) bool { return e.BuffId == tc.spellID && e.GetOnUse() != nil })
		if i < 0 {
			t.Fatalf("item %d carries no on-use effect on %d", tc.itemID, tc.spellID)
		}

		groups := map[string]Group{}
		if got := TryParseOnUseEffect(parsed, parsed.ItemEffects[i], instance, groups); got != EffectParseResultSuccess {
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

// An on-use whose spell applies a hit or crit percent, which no item stat holds, registers as an
// aura: Gri'lek's Charm of Valor's 24498 is spell crit alone. Weakness Analyzer's 1291101 drops at the
// first crit, a charge the aura does not spend, and is refused.
func TestOnUsePercentBuffRoutesAsAnAura(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		itemID  int
		spellID int32
		summary string
		reasons []string
	}{
		{19952, 24498, "on use: 24498 (A_MOD_SPELL_CRIT_CHANCE)", nil}, // Gri'lek's Charm of Valor
		{272438, 1291101, "on use: 1291101 (A_MOD_CRIT_PCT)", // Weakness Analyzer
			[]string{"the buff drops after ProcCharges 1, which the aura does not spend"}},
	} {
		item := instance.Items[tc.itemID]
		parsed := item.ToUIItem()
		parsed.ItemEffects = dbc.MergeItemEffectsForAllStates(parsed)
		i := slices.IndexFunc(parsed.ItemEffects, func(e *proto.ItemEffect) bool { return e.BuffId == tc.spellID && e.GetOnUse() != nil })
		if i < 0 {
			t.Fatalf("item %d carries no on-use effect on %d", tc.itemID, tc.spellID)
		}

		want, group := EffectParseResultSuccess, "Auras"
		if tc.reasons != nil {
			want, group = EffectParseResultRefused, ""
		}
		groups := map[string]Group{}
		if got := TryParseOnUseEffect(parsed, parsed.ItemEffects[i], instance, groups); got != want {
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

// Aegis of Preservation 19345 registers its 500 armor and lists the heal on every hit taken, 23781,
// that its buff 23780 procs.
func TestOnUseStatBuffListsTheProcItCarries(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	item := instance.Items[19345]
	parsed := item.ToUIItem()
	parsed.ItemEffects = dbc.MergeItemEffectsForAllStates(parsed)
	groups := map[string]Group{}
	if got := TryParseOnUseEffect(parsed, parsed.ItemEffects[0], instance, groups); got != EffectParseResultSuccess {
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
