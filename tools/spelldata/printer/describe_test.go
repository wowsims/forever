package printer

import (
	"fmt"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// One row per wording branch, each a copy of the game files row named by its id with the effect the
// branch words, so a game build that moves a number does not move the words.
func TestEffectLines(t *testing.T) {
	caster := [2]dbcenums.ImplicitTarget{1, 0}
	enemy := [2]dbcenums.ImplicitTarget{6, 0}
	effect := func(row spelldata.Spell, e spelldata.Effect) spelldata.Spell {
		e.SpellID = row.ID
		row.Effects = []spelldata.Effect{e}
		return row
	}
	cases := []struct {
		name    string
		row     spelldata.Spell
		human   string
		literal string
	}{
		{
			name: "a bleed tick",
			row: effect(spelldata.Spell{ID: 11574, School: core.SpellSchoolPhysical, DurationMs: 21000},
				spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PERIODIC_DAMAGE, BasePoints: 21,
					BaseLevel: 60, APCoef: 0.02, PeriodMs: 3000, Target: enemy}),
			human:   "21 physical damage every 3 s to the enemy (7 ticks, +0.02 attack power)",
			literal: "E_APPLY_AURA A_PERIODIC_DAMAGE base=21 ap=0.02 period=3000ms target=[6,0]",
		},
		{
			name: "school damage that rolls and scales",
			row: effect(spelldata.Spell{ID: 116, School: core.SpellSchoolFrost},
				spelldata.Effect{Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 19, PPL: 0.5, Variance: 0.10526315867900848,
					BaseLevel: 4, MaxLevel: 8, SPCoef: 0.40700000524520874, Target: enemy}),
			human:   "20–22 frost damage to the enemy (at level 60, +0.407 spell power)",
			literal: "E_SCHOOL_DAMAGE base=19 ppl=0.5 variance=0.10526316 sp=0.407 target=[6,0]",
		},
		{
			name: "normalised weapon damage over an area",
			row: effect(spelldata.Spell{ID: 1680, School: core.SpellSchoolPhysical},
				spelldata.Effect{Type: dbcenums.E_NORMALIZED_WEAPON_DMG, BaseLevel: 36, SPCoef: 1,
					Target: [2]dbcenums.ImplicitTarget{22, 15}}),
			human:   "normalised weapon damage to every enemy around the caster",
			literal: "E_NORMALIZED_WEAPON_DMG base=0 sp=1 target=[22,15]",
		},
		{
			name: "normalised weapon damage with a flat bonus",
			row: effect(spelldata.Spell{ID: 19434, School: core.SpellSchoolPhysical},
				spelldata.Effect{Type: dbcenums.E_NORMALIZED_WEAPON_DMG, BasePoints: 20, BaseLevel: 20, SPCoef: 1,
					Target: enemy}),
			human:   "normalised weapon damage +20 to the enemy",
			literal: "E_NORMALIZED_WEAPON_DMG base=20 sp=1 target=[6,0]",
		},
		{
			name: "a proc that names its triggered spell",
			row: effect(spelldata.Spell{ID: 21838, School: core.SpellSchoolPhysical, DurationMs: -1},
				spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PROC_TRIGGER_SPELL, BaseLevel: 60,
					TriggerID: 29478, Target: caster}),
			human:   "casts 29478 Battlegear of Might on its trigger",
			literal: "E_APPLY_AURA A_PROC_TRIGGER_SPELL base=0 trigger=29478 target=[1,0]",
		},
		{
			name: "a percentage modifier naming its spells by class mask",
			row: effect(spelldata.Spell{ID: 16493, School: core.SpellSchoolPhysical, ClassFlags: core.ClassFlags{Family: 4}},
				spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_ADD_PCT_MODIFIER, BasePoints: 10,
					BaseLevel: 60, SPCoef: 1, Misc: 15,
					ClassFlags: core.ClassFlags{Family: 4, Mask: [4]uint32{0: 3999288558, 1: 33088, 2: 1, 3: 268436480}},
					Target:     caster}),
			human: "SPELLMOD_CRIT_DAMAGE_BONUS +10% on family 4 mask 0xee604cee,0x8140,0x1,0x10000400",
			literal: "E_APPLY_AURA A_ADD_PCT_MODIFIER base=10 sp=1 misc=15 " +
				"family=4 mask=0xee604cee,0x8140,0x1,0x10000400 target=[1,0]",
		},
		{
			name: "a flat cost modifier, which the game files state on a 0-1000 bar",
			row: effect(spelldata.Spell{ID: 12282, School: core.SpellSchoolPhysical, ClassFlags: core.ClassFlags{Family: 4}},
				spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_ADD_FLAT_MODIFIER, BasePoints: -10,
					BaseLevel: 1, Misc: 14, ClassFlags: core.ClassFlags{Family: 4, Mask: [4]uint32{0: 64}}, Target: caster}),
			human:   "SPELLMOD_COST -10 (-1 on a rage bar) on family 4 mask 0x40",
			literal: "E_APPLY_AURA A_ADD_FLAT_MODIFIER base=-10 misc=14 family=4 mask=0x40 target=[1,0]",
		},
		{
			name: "a talent the game files state nothing about",
			row: effect(spelldata.Spell{ID: 12295, School: core.SpellSchoolPhysical},
				spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_DUMMY, BasePoints: 15, BaseLevel: 1,
					Target: caster}),
			human:   "a dummy aura holding 15",
			literal: "E_APPLY_AURA A_DUMMY base=15 target=[1,0]",
		},
		{
			name: "a dummy tick, which states no amount at all",
			row: effect(spelldata.Spell{ID: 10, School: core.SpellSchoolFrost, DurationMs: 8000},
				spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PERIODIC_DUMMY, BaseLevel: 20,
					MaxLevel: 25, PeriodMs: 1000, Target: caster}),
			human:   "a dummy aura every 1 s (8 ticks)",
			literal: "E_APPLY_AURA A_PERIODIC_DUMMY base=0 period=1000ms target=[1,0]",
		},
		{
			name: "a heal that rolls",
			row: effect(spelldata.Spell{ID: 20007, School: core.SpellSchoolHoly, DurationMs: 15000},
				spelldata.Effect{Type: dbcenums.E_HEAL, BasePoints: 100, Variance: 0.5, Target: caster}),
			human:   "75–125 healing to the caster",
			literal: "E_HEAL base=100 variance=0.5 target=[1,0]",
		},
		{
			name: "a heal that rolls and scales",
			row: effect(spelldata.Spell{ID: 2050, School: core.SpellSchoolHoly},
				spelldata.Effect{Type: dbcenums.E_HEAL, BasePoints: 51, PPL: 0.8999999761581421,
					Variance: 0.19607843458652496, BaseLevel: 1, MaxLevel: 3, SPCoef: 0.42899999022483826, Amplitude: 1,
					Target: [2]dbcenums.ImplicitTarget{21, 0}}),
			human:   "47–58 healing to the friendly target (at level 60, +0.429 spell power)",
			literal: "E_HEAL base=51 ppl=0.9 variance=0.19607843 sp=0.429 amplitude=1 target=[21,0]",
		},
		{
			name: "rage off the game files' 0-1000 bar",
			row: effect(spelldata.Spell{ID: 100, School: core.SpellSchoolPhysical},
				spelldata.Effect{Type: dbcenums.E_ENERGIZE, BasePoints: 90, BaseLevel: 4, Misc: 1, Target: caster}),
			human:   "restores 9 rage to the caster",
			literal: "E_ENERGIZE base=90 misc=1 target=[1,0]",
		},
		{
			name: "a buff the caster hands to its party",
			row: effect(spelldata.Spell{ID: 25289, School: core.SpellSchoolPhysical, DurationMs: 180000},
				spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_ATTACK_POWER, BasePoints: 139,
					PPL: 0.6000000238418579, BaseLevel: 60, MaxLevel: 61, RadiusMax: 20,
					Target: [2]dbcenums.ImplicitTarget{20, 0}}),
			human:   "+139 attack power to the party around the caster (at level 60)",
			literal: "E_APPLY_AURA A_MOD_ATTACK_POWER base=139 ppl=0.6 radius=20 target=[20,0]",
		},
		{
			name: "crit on the attacks made with a melee weapon",
			row: effect(spelldata.Spell{ID: 12320, School: core.SpellSchoolPhysical,
				EquipClass: dbcenums.ITEM_CLASS_WEAPON, EquipSubclass: 173555},
				spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_WEAPON_CRIT_PERCENT, BasePoints: 5,
					BaseLevel: 1, Target: caster}),
			human:   "+5% physical crit, per weapon: only attacks made with a melee weapon",
			literal: "E_APPLY_AURA A_MOD_WEAPON_CRIT_PERCENT base=5 target=[1,0]",
		},
		{
			name: "hit on the off hand's attacks, by the hand bit",
			row: effect(spelldata.Spell{ID: 1323963, School: core.SpellSchoolPhysical,
				Attr:       [17]uint32{0: 0x1d0, 3: 0x1000000, 15: 0x2000},
				EquipClass: dbcenums.ITEM_CLASS_WEAPON, EquipSubclass: 41105, EquipInvType: 4194304},
				spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_HIT_CHANCE, BasePoints: 10,
					BaseLevel: 60, SPCoef: 1, Target: caster}),
			human:   "+10% physical hit, per weapon: only attacks made with a one-handed axe, one-handed mace, one-handed sword, fist weapon or dagger in the off hand",
			literal: "E_APPLY_AURA A_MOD_HIT_CHANCE base=10 sp=1 target=[1,0]",
		},
		{
			name: "crit on everything while a melee weapon is equipped",
			row: effect(spelldata.Spell{ID: 16255, School: core.SpellSchoolPhysical,
				EquipClass: dbcenums.ITEM_CLASS_WEAPON, EquipSubclass: 173555},
				spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_CRIT_PCT, BasePoints: 1,
					BaseLevel: 1, Target: caster}),
			human:   "+1% physical and spell crit, while a melee weapon is equipped: every spell and attack",
			literal: "E_APPLY_AURA A_MOD_CRIT_PCT base=1 target=[1,0]",
		},
		{
			name: "crit on everything with ranged weapons alone, which is per weapon",
			row: effect(spelldata.Spell{ID: 1310308, School: core.SpellSchoolPhysical,
				EquipClass: dbcenums.ITEM_CLASS_WEAPON, EquipSubclass: 262156},
				spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_CRIT_PCT, BasePoints: 2,
					BaseLevel: 60, Target: caster}),
			human:   "+2% physical crit, ranged weapons only: attacks made with a bow, gun or crossbow",
			literal: "E_APPLY_AURA A_MOD_CRIT_PCT base=2 target=[1,0]",
		},
		{
			name: "hit taken off the enemy, whose weapon requirement is the cast's",
			row: effect(spelldata.Spell{ID: 3043, School: core.SpellSchoolNature, DurationMs: 20000,
				EquipClass: dbcenums.ITEM_CLASS_WEAPON, EquipSubclass: 262156},
				spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_HIT_CHANCE, BasePoints: -2,
					BaseLevel: 22, SPCoef: 1, Target: enemy}),
			human:   "-2% physical hit to the enemy (the weapon requirement only gates casting)",
			literal: "E_APPLY_AURA A_MOD_HIT_CHANCE base=-2 sp=1 target=[6,0]",
		},
		{
			name: "hit rating off its melee, ranged and spell bits",
			row: effect(spelldata.Spell{ID: 1299734, School: core.SpellSchoolPhysical},
				spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_RATING, BasePoints: 5,
					BaseLevel: 60, Misc: 224, Target: caster}),
			human:   "+5 hit rating",
			literal: "E_APPLY_AURA A_MOD_RATING base=5 misc=224 target=[1,0]",
		},
		{
			name: "the expertise bit",
			row: effect(spelldata.Spell{ID: 1301076, School: core.SpellSchoolPhysical, ClassFlags: core.ClassFlags{Family: 7}},
				spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_RATING, BasePoints: 12,
					Misc: 8388608, Target: caster}),
			human:   "+12 expertise rating",
			literal: "E_APPLY_AURA A_MOD_RATING base=12 misc=8388608 target=[1,0]",
		},
		{
			name: "a shape no branch words",
			row: effect(spelldata.Spell{ID: 348, School: core.SpellSchoolFire, DurationMs: 15000},
				spelldata.Effect{Type: dbcenums.E_SCRIPT_EFFECT, BasePoints: 100, BaseLevel: 1, MaxLevel: 5,
					Target: enemy}),
			human:   unrecognised,
			literal: "E_SCRIPT_EFFECT base=100 target=[6,0]",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			lines := effectLines(&c.row)
			if len(lines) != 1 {
				t.Fatalf("a one-effect row words %d effects", len(lines))
			}
			if lines[0].Human != c.human {
				t.Errorf("human = %q, want %q", lines[0].Human, c.human)
			}
			if lines[0].Literal != c.literal {
				t.Errorf("literal = %q, want %q", lines[0].Literal, c.literal)
			}
		})
	}
}

func TestHeader(t *testing.T) {
	secs := func(ms int32) string { return column(float64(ms)/1000) + " s" }
	rage := func(s *spelldata.Spell) string { return column(s.PowerCost(dbcenums.POWER_RAGE)) + " rage" }
	rend := spelldata.MustFind(11574)
	whirlwind := spelldata.MustFind(1680)
	shieldBlock := spelldata.MustFind(2565)
	engrave := spelldata.MustFind(410021)

	cases := []struct {
		name string
		id   int32
		want []field
	}{
		{
			name: "a bleed with a stance and a weapon requirement",
			id:   11574,
			want: []field{
				{"school", "physical"},
				{"defense", "melee"},
				{"mechanic", "bleed"},
				{"duration", secs(rend.DurationMs)},
				{"gcd", secs(rend.GCDMs)},
				{"cost", rage(rend)},
				{"range", column(float64(rend.MaxRange)) + " yd"},
				{"stance", "Battle, Defensive"},
				{"equip", "needs a melee weapon in the main hand"},
				{"attrs", "refund on miss, periodic can crit"},
				{"labels", "25"},
			},
		},
		{
			name: "an area ability with a target cap",
			id:   1680,
			want: []field{
				{"school", "physical"},
				{"defense", "melee"},
				{"cooldown", secs(whirlwind.CategoryCooldownMs) + " (category)"},
				{"gcd", secs(whirlwind.GCDMs)},
				{"cost", rage(whirlwind)},
				{"stance", "Berserker"},
				{"equip", "needs a melee weapon"},
				{"targets", fmt.Sprintf("up to %d", whirlwind.MaxTargets)},
				{"labels", "25"},
			},
		},
		{
			name: "charges and a shield",
			id:   2565,
			want: []field{
				{"school", "physical"},
				{"duration", secs(shieldBlock.DurationMs)},
				{"cooldown", secs(shieldBlock.CooldownMs)},
				{"cost", rage(shieldBlock)},
				{"stance", "Defensive"},
				{"equip", "needs a shield"},
				{"charges", fmt.Sprint(shieldBlock.ProcCharges)},
				{"labels", "25"},
			},
		},
		{
			name: "one-handers in the off hand, by the hand bit",
			id:   23584,
			want: []field{
				{"school", "physical"},
				{"equip", "needs a one-handed axe, one-handed mace, one-handed sword, fist weapon or dagger in the off hand"},
				{"attrs", "passive"},
				{"labels", "25"},
			},
		},
		{
			name: "every ranged weapon but the wand, named one by one",
			id:   450403,
			want: []field{
				{"school", "physical"},
				{"equip", "needs a bow, gun, thrown or crossbow"},
				{"attrs", "passive"},
			},
		},
		{
			name: "an enchant naming the slots it takes",
			id:   410021,
			want: []field{
				{"school", "physical"},
				{"cast", secs(engrave.CastTimeMs)},
				{"equip", "needs cloth, leather, mail or plate in the chest or robe"},
				{"labels", "3071"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := headerFields(spelldata.MustFind(c.id))
			if !slices.Equal(got, c.want) {
				t.Errorf("header =\n%q\nwant\n%q", got, c.want)
			}
		})
	}

	if got := title(spelldata.MustFind(11574)); got != "11574 Rend (Rank 7)" {
		t.Errorf("title = %q", got)
	}
}

func TestProcSummary(t *testing.T) {
	cases := []struct {
		name string
		id   int32
		want string
	}{
		{
			name: "a chance the column really states",
			id:   21838,
			want: fmt.Sprintf("hears PROC_FLAG_TAKE_ANY_DAMAGE; %d%% chance", spelldata.MustFind(21838).ProcChance),
		},
		{
			name: "a 100 the tooltip states no chance beside",
			id:   12834,
			want: "hears PROC_FLAG_DEAL_MELEE_SWING, PROC_FLAG_DEAL_MELEE_ABILITY, " +
				"PROC_FLAG_DEAL_RANGED_ATTACK, PROC_FLAG_DEAL_RANGED_ABILITY, " +
				"PROC_FLAG_DEAL_HARMFUL_ABILITY, PROC_FLAG_DEAL_HARMFUL_SPELL; " +
				"no roll, it fires whenever its condition is met; tooltip says on a crit",
		},
		{
			name: "a chance an effect states, not the column",
			id:   11103,
			want: "hears PROC_FLAG_DEAL_MELEE_ABILITY, PROC_FLAG_DEAL_RANGED_ATTACK, " +
				"PROC_FLAG_DEAL_RANGED_ABILITY, PROC_FLAG_DEAL_HELPFUL_ABILITY, " +
				"PROC_FLAG_DEAL_HARMFUL_ABILITY, PROC_FLAG_DEAL_HELPFUL_SPELL, " +
				"PROC_FLAG_DEAL_HARMFUL_SPELL; " + column(spelldata.MustFind(11103).EffectN(1).BaseValue()) + "% chance, from effect 1; " +
				"tooltip says one named ability",
		},
		{
			name: "the internal cooldown is a header line, not part of the proc",
			id:   324,
			want: "hears PROC_FLAG_TAKE_MELEE_SWING, PROC_FLAG_TAKE_MELEE_ABILITY, " +
				"PROC_FLAG_TAKE_RANGED_ATTACK, PROC_FLAG_TAKE_RANGED_ABILITY, " +
				"PROC_FLAG_TAKE_HARMFUL_ABILITY, PROC_FLAG_TAKE_HARMFUL_SPELL; " +
				"no roll, it fires whenever its condition is met; tooltip says heals count",
		},
		{
			name: "the heartbeat bit alone is not a proc",
			id:   21849,
			want: "",
		},
		{
			name: "no proc columns at all",
			id:   11574,
			want: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := procSummary(spelldata.MustFind(c.id)); got != c.want {
				t.Errorf("procSummary = %q, want %q", got, c.want)
			}
		})
	}

	if got := refList(spelldata.MustFind(12834)); !slices.Equal(got, []string{"412609 Deep Wound"}) {
		t.Errorf("refList = %q", got)
	}
	if got := refList(spelldata.MustFind(11574)); len(got) != 0 {
		t.Errorf("refList = %q, want none", got)
	}
}

func TestLadderRefs(t *testing.T) {
	cases := []struct {
		name string
		id   int32
		want []string
	}{
		{
			name: "the top rank of a Ranked family",
			id:   11574,
			want: []string{"warrior spellData.Rend.Highest()"},
		},
		{
			name: "a rank below the top",
			id:   11572,
			want: []string{"warrior spellData.Rend.ByID(11572)"},
		},
		{
			name: "a trait-tree talent, whose ranks are one spell",
			id:   12282,
			want: []string{"warrior spellData.ImprovedHeroicStrike.Rank(n), n up to 3"},
		},
		{
			name: "an id two families carry",
			id:   12295,
			want: []string{
				"warrior spellData.ImprovedTacticalMastery.Rank(n), n up to 5",
				"warrior spellData.TacticalMasteryTriggered.Highest()",
			},
		},
		{
			name: "an item set bonus, which no class file names",
			id:   21838,
			want: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ladderRefs(c.id); !slices.Equal(got, c.want) {
				t.Errorf("ladderRefs(%d) = %q, want %q", c.id, got, c.want)
			}
		})
	}
}

func TestHighestRank(t *testing.T) {
	if s := highestRank(spelldata.ByName("Rend")); s.ID != 11574 {
		t.Errorf("Rend picks %d %s, want 11574 Rank 7", s.ID, s.Rank)
	}
	if s := highestRank(spelldata.ByName("Whirlwind")); s.ID != 1680 {
		t.Errorf("Whirlwind picks %d, want the warrior ability 1680", s.ID)
	}
}

func TestParseArgs(t *testing.T) {
	for _, args := range [][]string{{"11574", "-json"}, {"-json", "11574"}} {
		opts, err := parseArgs(args)
		if err != nil {
			t.Fatalf("parseArgs(%q): %v", args, err)
		}
		if opts.mode != "spell" || opts.arg != "11574" || !opts.json {
			t.Errorf("parseArgs(%q) = %+v", args, opts)
		}
	}
	if _, err := parseArgs(nil); err == nil {
		t.Error("parseArgs with no arguments wants a usage error")
	}
}

// A rating row bound to some attacks words what the parse converts: the store states none, so the rows
// are built here.
func TestScopedRatingLines(t *testing.T) {
	rating := func(mask int32, points float64) spelldata.Effect {
		return spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_RATING, Misc: mask, BasePoints: points,
			Target: [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_CASTER}}
	}
	cases := []struct {
		name string
		row  spelldata.Spell
		want string
	}{
		{"ranged hit rating alone", spelldata.Spell{Effects: []spelldata.Effect{rating(64, 20)}},
			"+20 hit rating, ranged attacks only, worth +2% hit"},
		{"ranged crit and haste rating", spelldata.Spell{Effects: []spelldata.Effect{rating(512|262144, 28)}},
			"+28 crit and haste rating, ranged attacks only, worth +2% crit; the haste rating is not read"},
		{"hit and crit rating with a dagger", spelldata.Spell{EquipClass: dbcenums.ITEM_CLASS_WEAPON,
			EquipSubclass: dbcenums.ITEM_SUBCLASS_BIT_WEAPON_DAGGERS, Effects: []spelldata.Effect{rating(32|256, 14)}},
			"+14 hit and crit rating, per weapon: only attacks made with a dagger, worth +1.4% hit and +1% crit"},
	}
	for _, c := range cases {
		if got := humanise(&c.row, &c.row.Effects[0]); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}
