package printer

import (
	"bytes"
	"go/constant"
	"go/parser"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func TestWeaponAndRatingConstants(t *testing.T) {
	for text, want := range map[string]string{
		"stats.HitRating":                        "17",
		"stats.CritRating":                       "18",
		"stats.HasteRating":                      "19",
		"core.HitRatingPerHitPercent":            "10",
		"core.CritRatingPerCritPercent":          "14",
		"core.HasteRatingPerHastePercent":        "10",
		"core.SpellMod_BonusHit_Percent":         "4096",
		"dbcenums.ITEM_CLASS_WEAPON":             "2",
		"dbcenums.ITEM_SUBCLASS_BIT_WEAPON_FIST": "8192",
		"dbcenums.INVTYPE_BIT_WEAPONOFFHAND":     "4194304",
		"dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_POLEARM": "67",
	} {
		expr, err := parser.ParseExpr(text)
		if err != nil {
			t.Fatal(err)
		}
		value, err := evalConst(expr)
		if err != nil {
			t.Errorf("%s: %v", text, err)
			continue
		}
		if got := constant.ToInt(value).String(); got != want {
			t.Errorf("%s = %s, want %s", text, got, want)
		}
	}
}

const weaponRowsGo = `package warrior

var cruelty = spelldata.MustFind(12320)
var thunderingStrikes = spelldata.MustFind(16255)
var rangedScope = spelldata.MustFind(1310308)
var dualWield = spelldata.MustFind(23584)
var furiousPrecision = spelldata.MustFind(1323963)
var scorpidSting = spelldata.MustFind(3043)
var twoHandedSpec = spelldata.MustFind(12163)
`

func TestHoverWeaponRules(t *testing.T) {
	percent := func(id int32) string { return signedColumn(spelldata.MustFind(id).EffectN(1).Mean()) + "%" }
	for needle, want := range map[string]string{
		"MustFind(12320)":   "**1** · " + percent(12320) + " physical crit, per weapon: only attacks made with a melee weapon  \n",
		"MustFind(16255)":   "**1** · " + percent(16255) + " physical and spell crit, while a melee weapon is equipped: every spell and attack  \n",
		"MustFind(1310308)": "**1** · " + percent(1310308) + " physical crit, ranged weapons only: attacks made with a bow, gun or crossbow  \n",
		"MustFind(1323963)": "**1** · " + percent(1323963) + " physical hit, per weapon: only attacks made with a one-handed axe, one-handed mace, one-handed sword, fist weapon or dagger in the off hand  \n",
		"MustFind(3043)":    "**1** · " + percent(3043) + " physical hit to the enemy (the weapon requirement only gates casting)  \n",
		"MustFind(12163)":   "**1** · " + percent(12163) + " physical damage done and on the melee-defense spells of the other schools, while a two-handed axe, two-handed mace, polearm, two-handed sword, staff, bit 0x1000 or spear is equipped  \n",
	} {
		wantHover(t, weaponRowsGo, needle, 9, want)
	}
	wantHover(t, weaponRowsGo, "MustFind(23584)", 9,
		"| **equip** | needs a one-handed axe, one-handed mace, one-handed sword, fist weapon or dagger in the off hand |")
}

func TestExprWeaponRow(t *testing.T) {
	got := evalJSON(t, "spelldata.MustFind(3043).Effect(dbcenums.A_MOD_HIT_CHANCE, 0).BaseValue()")
	want := column(spelldata.MustFind(3043).Effect(dbcenums.A_MOD_HIT_CHANCE, 0).BaseValue())
	if got.Value != want || got.Trail != "spelldata.MustFind(3043).Effect(54, 0).BaseValue()" {
		t.Errorf("the chain reads %s = %s", got.Trail, got.Value)
	}
	if human := got.Effects[readEffect(got)-1].Human; !strings.Contains(human, "only gates casting") {
		t.Errorf("the effect read is worded %q", human)
	}

	if got := evalJSON(t, "spelldata.MustFind(12320).EquipSubclass"); got.Value != "173555" {
		t.Errorf("Cruelty's subclass mask reads %s", got.Value)
	}
}

func TestConfigSpellModConfig(t *testing.T) {
	var out bytes.Buffer
	literal := "core.SpellModConfig{Kind: core.SpellMod_BonusHit_Percent, FloatValue: 2, ProcMask: core.ProcMaskMelee, " +
		"EquipSubclass: dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_2H, EquipHand: core.ProcMaskMeleeOH, FromRating: true}"
	if err := Run([]string{"-config", literal}, &out); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{
		"Kind           SpellMod_BonusHit_Percent",
		"EquipSubclass  384 (one-handed sword or two-handed sword)",
		"EquipHand      ProcMaskMeleeOHAuto | ProcMaskMeleeOHSpecial",
		"FromRating     true",
		"per weapon: only attacks made with a one-handed sword or two-handed sword in the off hand. " +
			"Magic-defense spells and pets never match; an item swap re-filters it. The sheet counts it as rating.",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("-config lacks %q:\n%s", want, text)
		}
	}
}

const weaponmasterGo = `package warrior

func (warrior *Warrior) registerWeaponmaster() {
	rank := warrior.Talents.Weaponmaster
	critAura.AttachSpellMod(core.SpellModConfig{
		Kind:          core.SpellMod_BonusCrit_Percent,
		FloatValue:    spellData.Weaponmaster.EffectAt(1).ValueAt(rank),
		ProcMask:      core.ProcMaskMelee,
		EquipSubclass: dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_POLEARM,
	})
	rangedOnly := core.SpellModConfig{Kind: core.SpellMod_BonusCrit_Percent, ProcMask: core.ProcMaskRanged, EquipSubclass: 262156}
}
`

func TestHoverSpellModConfig(t *testing.T) {
	wantHover(t, weaponmasterGo, "AttachSpellMod(core.SpellModConfig", 22,
		"| **Kind** | `SpellMod_BonusCrit_Percent` | `core.SpellMod_BonusCrit_Percent` |",
		"| **EquipSubclass** | `67 (one-handed axe, two-handed axe or polearm)` | "+
			"`dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE \\| dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE \\| dbcenums.ITEM_SUBCLASS_BIT_WEAPON_POLEARM` |",
		"per weapon: only attacks made with a one-handed axe, two-handed axe or polearm.",
		"unevaluated: `FloatValue: spellData.Weaponmaster.EffectAt(1).ValueAt(rank)` (rank is not a literal")
	wantHover(t, weaponmasterGo, "SpellModConfig{Kind", 3,
		"| **EquipSubclass** | `262156 (bow, gun or crossbow)` | `262156` |",
		"ranged weapons only: attacks made with a bow, gun or crossbow.")
}
