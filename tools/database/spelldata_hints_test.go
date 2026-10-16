package database

// The proc shapes docs/spell_data/procs.md names, read off the game files database the way the store's rows
// are. Which shape a spell has is a reading of its tooltip, and a reading is what silently changes
// when a matcher is widened: Flurry states "$m1%" for its attack speed and Unbridled Wrath states
// "$m1% chance" for its rage, and only the second is a roll.
//
// Skips when tools/database/wowsims.db is absent, which is why CI is unaffected.

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/tools/database/overrides"
)

var clientTables struct {
	once   sync.Once
	tables *spellTables
	err    error
}

func clientSpellTables(t *testing.T) *spellTables {
	t.Helper()
	DatabasePath = "wowsims.db"
	if _, err := os.Stat(DatabasePath); err != nil {
		t.Skipf("no game files database at %s - run `make db` from a local WoW install to enable this gate", DatabasePath)
	}

	clientTables.once.Do(func() {
		helper, err := NewDBHelper()
		if err != nil {
			clientTables.err = fmt.Errorf("opening %s: %w", DatabasePath, err)
			return
		}
		defer helper.Close()

		if clientTables.tables, err = loadSpellTables(helper.db); err != nil {
			clientTables.err = fmt.Errorf("loading the spell tables: %w", err)
		}
	})
	if clientTables.err != nil {
		t.Fatal(clientTables.err)
	}
	return clientTables.tables
}

func TestProcShapeOfNamedSpells(t *testing.T) {
	tables := clientSpellTables(t)

	for _, want := range []struct {
		id      int32
		name    string
		source  storeProcChanceSource
		effect  int8
		why     string
		hint    core.ProcHint
		hintWhy string
	}{
		{12317, "Enrage", procChanceColumn, 0, `the tooltip's "$h%" is the ProcChance column`,
			core.ProcHintDamage, `"any damaging attack" is damage wording, which the taken mask it sits beside ignores`},
		{12322, "Unbridled Wrath", procChanceEffectN, 1, `"$m1% chance to generate Rage" is effect 1's ladder`,
			0, `"when you deal melee damage with a weapon" is the mask's own shape`},
		{12298, "Shield Specialization", procChanceEffectN, 2, `"$m2% chance to generate Rage", past the "$s1%" block value`,
			core.ProcHintOutcomeTaken, `"when you Block" is an outcome no ProcTypeMask has a bit for`},
		{12319, "Flurry", procChanceAlways, 0, `"$m1%" is the attack speed, and a crit is the condition`,
			0, `"after dealing a melee critical strike" is the crit the caller reads off the result, since the same listener has to see the white hits that spend a charge`},
		{12834, "Deep Wounds", procChanceAlways, 0, `"$m1%" is the share of weapon damage, and a crit is the condition`,
			core.ProcHintCrit, `"Your critical strikes" names the trigger`},
		{30160, "Elemental Devastation", procChanceAlways, 0, `"$m1%" is the melee crit it grants, and a crit is the condition`,
			core.ProcHintCrit | core.ProcHintDamage, `"Your Critical Strikes", capitalized, names the trigger, and "direct damage spells" restricts it to damage`},
		{16928, "Armor Shatter", procChanceAlways, 0, "the 101 sentinel; the rate is Annihilator's, on its item effect",
			0, "the tooltip is the armor reduction, not a trigger"},
		{1308935, "Striking", procChanceColumn, 0, "no tooltip at all, so the column is all there is",
			0, "no tooltip to read a hint off"},
		{23548, "Battlegear of Wrath, the parry grant", procChanceColumn, 0, `the tooltip's "$h%" is the ProcChance column`,
			core.ProcHintOutcomeTaken, `"after a block" is an outcome no ProcTypeMask has a bit for`},
		{23547, "Battlegear of Wrath, the parry buff", procChanceAlways, 0, "the 100 in the column is the game files' no-roll",
			0, "the row ships no tooltip, so the parry it is spent on stays the caller's"},
		{29626, "Shadowbolt Volley", procChancePPM, 0, `a column of 100 next to "Chance to strike" is a rate the game files keep elsewhere`,
			0, "the wording states a rate, not a trigger the mask cannot carry"},
		{22618, "Force Reactive Disk", procChanceAlways, 0, `the "chance of damaging the shield" is the shield's durability, in a sentence naming nothing the charge fires on; the charge itself fires on every block`,
			core.ProcHintDamage, `"when the shield blocks" is the shield's own condition rather than an outcome of the wearer's, and "damaging the shield" is damage wording the taken mask ignores`},
		{432042, "Tidal Waves", procChanceAlways, 0, `"increases the critical effect chance of your Lesser Healing Wave" is a magnitude, not a rate`,
			core.ProcHintCastTrigger | core.ProcHintCrit | core.ProcHintNamedAbility,
			`"When you cast Chain Heal or Riptide" names the cast, the ability and the critical effect it grants`},
		{467889, "T2 Shaman Tank 2P", procChanceAlways, 0, `"grants increased chance to Block" is a magnitude, not a rate`,
			0, `"until you Block an attack" is what spends the buff, not what triggers it`},
		{1226977, "Scarlet Enclave Elemental 4P", procChanceAlways, 0, `"Chance to trigger Overload increased by an additional" is a magnitude, not a rate`,
			core.ProcHintDamage, `"When Lightning Bolt or Chain Lightning deal damage" restricts the trigger to damage; the rest modifies another spell's chance`},
		{440529, "Resourcefulness", procChanceEffectN, 2, `"a $m2% chance" is effect 2's ladder`,
			0, `"grants your melee and ranged ability Critical Strikes" names the crit after the verb, which no hint reads; the caller sets the outcome`},
		{23689, "Darkmoon Card: Heroism", procChancePPM, 0, `"Sometimes heals" beside the 100 is a rate the game files keep elsewhere`,
			core.ProcHintHeals | core.ProcHintDamage, `"heals bearer" is heal wording, and "when damaging an enemy" restricts the trigger to damage`},
		{1248761, "Recovery, an enchant's equip aura", procChanceAlways, 0, `the row ships no tooltip, and its grant 1248760's "Cannot occur more often than" is a cooldown, not a rate`,
			core.ProcHintAttackDodged | core.ProcHintAttackParried, `the grant reads "when you are Parried or Dodged"`},
		{1248806, "Revelation, an enchant's equip aura", procChancePPM, 0, `its grant 1248805's "a chance to trigger Revelation" beside the 100 is a rate the game files keep elsewhere`,
			0, "the grant's crit wording names which hits feed the proc, which is the row's own mask to say"},
		{1248758, "Insight, an enchant's equip aura", procChanceColumn, 0, "the column's 35 is the roll",
			0, `the grant's "when you cast a spell" names which hits feed the proc, which is the row's own mask to say`},
		{28200, "Talisman of Ascendance's stacking aura", procChanceAlways, 0, "the 100 in the column is the game files' no-roll",
			core.ProcHintHeals | core.ProcHintDamage, `"Your next 5 damage or healing spells" restricts the trigger to damage and names healing without being restricted to it`},
		{9452, "Vindication", procChanceAlways, 0, "the 100 in the column is the game files' no-roll",
			core.ProcHintDamage, `"your damaging melee attacks" restricts the trigger to damage`},
		{14076, "Improved Sap", procChanceEffectN, 1, `"$s1% chance" is effect 1's ladder`,
			core.ProcHintNamedAbility, `"your Sap ability" names the ability and no damage`},
		{7711, "Blazefury Medallion", procChanceAlways, 0, "the 100 in the column is the game files' no-roll",
			core.ProcHintDamage, "an override rules the proc to hits with a damage component"},
	} {
		t.Run(want.name, func(t *testing.T) {
			rows := []storeSpell{tables.row(want.id)}
			applyTooltipHints(tables, &rows[0])
			if err := applyOverrides(rows, overridesOf(want.id)); err != nil {
				t.Fatalf("baking the overrides: %v", err)
			}
			row := rows[0]

			if row.ProcChanceSource != want.source || row.ProcChanceEffect != want.effect {
				t.Errorf("spell %d reads %s effect %d, and %s, so it is %s effect %d",
					want.id, row.ProcChanceSource, row.ProcChanceEffect, want.why, want.source, want.effect)
			}

			if row.ProcHint != want.hint {
				t.Errorf("spell %d reads hint %d, and %s, so it is %d",
					want.id, row.ProcHint, want.hintWhy, want.hint)
			}

			// An effect the chance sits on has to be one the sim can reach by that number, since
			// EffectN counts positions and the game files' EffectIndex has gaps.
			if row.ProcChanceEffect != 0 {
				position := int(row.ProcChanceEffect) - 1
				if position >= len(row.Effects) {
					t.Fatalf("spell %d names effect %d of %d", want.id, row.ProcChanceEffect, len(row.Effects))
				}
				if !isProcEffect(&row.Effects[position]) {
					t.Errorf("spell %d's effect %d carries aura %d, which is no proc aura",
						want.id, row.ProcChanceEffect, row.Effects[position].Aura)
				}
			}
		})
	}
}

// A combat spell's roll is the one its enchantments state, whatever its own column says.
func TestEnchantChanceOfNamedCombatSpells(t *testing.T) {
	tables := clientSpellTables(t)

	fieryBlaze := tables.row(6297)
	stated, ok := tables.EnchantChances[6297]
	if !ok || stated.Chance == 0 {
		t.Fatalf("Fiery Blaze 6297: enchantments state no chance for it (%v)", stated)
	}
	applyTooltipHints(tables, &fieryBlaze)
	if err := applyEnchantChance(tables, &fieryBlaze); err != nil {
		t.Fatalf("Fiery Blaze 6297: %v", err)
	}
	if fieryBlaze.ProcChance != stated.Chance || fieryBlaze.ProcChanceSource != procChanceColumn || fieryBlaze.ProcChanceEffect != 0 {
		t.Errorf("Fiery Blaze 6297: ProcChance %d from %s effect %d, want enchantments %v's %d from the column",
			fieryBlaze.ProcChance, fieryBlaze.ProcChanceSource, fieryBlaze.ProcChanceEffect, stated.Enchants, stated.Chance)
	}

	for _, want := range []struct {
		id     int32
		name   string
		chance uint8
	}{
		{11398, "Mind-numbing Poison III, 20 on enchantments 643 and 8696 beside a column of 101", 20},
		{8516, "Windfury Totem, 20 on enchantment 1783 beside a column of 100", 20},
	} {
		row := tables.row(want.id)
		applyTooltipHints(tables, &row)
		if err := applyEnchantChance(tables, &row); err != nil {
			t.Fatalf("%s: %v", want.name, err)
		}
		if row.ProcChance != want.chance || row.ProcChanceSource != procChanceColumn || row.ProcChanceEffect != 0 {
			t.Errorf("%s: ProcChance %d from %s effect %d, want %d from the column",
				want.name, row.ProcChance, row.ProcChanceSource, row.ProcChanceEffect, want.chance)
		}
	}

	row := tables.row(1248758)
	applyTooltipHints(tables, &row)
	own := row.ProcChance
	if err := applyEnchantChance(tables, &row); err != nil || row.ProcChance != own || len(row.overrideNotes) != 0 {
		t.Errorf("Insight's equip aura 1248758 reads %d%% (%v, notes %v), want its own %d untouched",
			row.ProcChance, err, row.overrideNotes, own)
	}
}

func TestSplitDamageOfNamedSpells(t *testing.T) {
	tables := clientSpellTables(t)

	for _, want := range []struct {
		id     int32
		name   string
		splits bool
	}{
		{1295270, "Prototype Pathcarver", true},
		{26789, "Shard of the Fallen Star", true},
		{24340, "Meteor", true},
		{6297, "Fiery Blaze", false},
		{21179, "Chain Lightning", false},
	} {
		row := tables.row(want.id)
		applyTooltipHints(tables, &row)
		if row.SplitsDamage != want.splits {
			t.Errorf("%s %d reads SplitsDamage %v, want %v", want.name, want.id, row.SplitsDamage, want.splits)
		}
	}
}

func overridesOf(id int32) []overrides.Override {
	var mine []overrides.Override
	for _, o := range overrides.Spells {
		if o.SpellID == id {
			mine = append(mine, o)
		}
	}
	return mine
}
