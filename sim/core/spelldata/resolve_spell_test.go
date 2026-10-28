package spelldata

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// The category two of the rows below share, as SpellCategories.Category states it.
const testCategory int16 = 971

// Rows in the shapes the resolvers have to answer for: the four power bars, the three cooldown
// shapes, and the attributes and targets the flags come from.
func resolverRows() []Spell {
	return []Spell{
		{
			ID: 100, Name: "Rage Melee", Rank: "Rank 3", School: 1, DefenseType: 2,
			Attr:       [17]uint32{dbcenums.ATTR_INDEX_EX_1: dbcenums.ATTR_EX_1_DISCOUNT_POWER_ON_MISS},
			CooldownMs: 6000, GCDMs: 1500, StartRecoveryCategory: 133, MaxRange: 5,
			ClassFlags: core.ClassFlags{Family: 4, Mask: [4]uint32{0: 0x20}},
			Effects: []Effect{
				{SpellID: 100, Type: dbcenums.E_WEAPON_PERCENT_DAMAGE, BasePoints: 100, Target: [2]dbcenums.ImplicitTarget{6, 0}},
			},
			Powers: []Power{{Type: 1, Cost: 300}},
		},
		{
			ID: 200, Name: "Category Only", School: 1, DefenseType: 2,
			CategoryCooldownMs: 6000, GCDMs: 1500, StartRecoveryCategory: 133, Category: testCategory,
			Effects: []Effect{{SpellID: 200, Type: dbcenums.E_SCHOOL_DAMAGE, Target: [2]dbcenums.ImplicitTarget{6, 0}}},
			Powers:  []Power{{Type: 1, Cost: 150}},
		},
		{
			ID: 250, Name: "Physical Shout", School: 1, DefenseType: 1,
			GCDMs: 1500, StartRecoveryCategory: 133,
			Effects: []Effect{{SpellID: 250, Type: dbcenums.E_APPLY_AURA, Target: [2]dbcenums.ImplicitTarget{20, 0}}},
			Powers:  []Power{{Type: 1, Cost: 100}},
		},
		{
			ID: 300, Name: "Own And Category", School: 1, DefenseType: 2,
			CooldownMs: 30000, CategoryCooldownMs: 6000, GCDMs: 1500, StartRecoveryCategory: 133,
			Category: testCategory,
			Effects:  []Effect{{SpellID: 300, Type: dbcenums.E_SCHOOL_DAMAGE, Target: [2]dbcenums.ImplicitTarget{6, 0}}},
		},
		{
			ID: 350, Name: "Categoryless Cooldown", School: 1, DefenseType: 2,
			CategoryCooldownMs: 8000, GCDMs: 1500, StartRecoveryCategory: 133,
			Effects: []Effect{{SpellID: 350, Type: dbcenums.E_SCHOOL_DAMAGE, Target: [2]dbcenums.ImplicitTarget{6, 0}}},
		},
		{
			ID: 400, Name: "Mana Caster", Rank: "Rank 7", School: 16, DefenseType: 1, Speed: 24,
			CastTimeMs: 2500, GCDMs: 1500, StartRecoveryCategory: 133, MinRange: 8, MaxRange: 30,
			Effects: []Effect{
				{SpellID: 400, Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 500, SPCoef: 0.814,
					Target: [2]dbcenums.ImplicitTarget{6, 0}},
			},
			Powers: []Power{{Type: 0, Cost: 425, CostPct: 19}},
		},
		{
			ID: 500, Name: "Passive Aura", Rank: "Passive",
			Attr:    [17]uint32{dbcenums.ATTR_INDEX_BASE: dbcenums.ATTR_PASSIVE},
			Effects: []Effect{{SpellID: 500, Type: dbcenums.E_APPLY_AURA, Target: [2]dbcenums.ImplicitTarget{1, 0}}},
		},
		{
			ID: 600, Name: "Energy Channel", School: 8, DefenseType: 1,
			Attr: [17]uint32{
				dbcenums.ATTR_INDEX_EX_1: dbcenums.ATTR_EX_1_IS_CHANNELLED | dbcenums.ATTR_EX_1_DISCOUNT_POWER_ON_MISS,
				dbcenums.ATTR_INDEX_EX_4: dbcenums.ATTR_EX_4_SUPPRESS_WEAPON_PROCS,
			},
			DurationMs: 6000, GCDMs: 1000, StartRecoveryCategory: 133,
			Effects: []Effect{{SpellID: 600, Type: dbcenums.E_APPLY_AURA, Target: [2]dbcenums.ImplicitTarget{6, 0}}},
			Powers:  []Power{{Type: 3, Cost: 60}},
		},
		{
			ID: 700, Name: "Raid Heal", School: 2, DefenseType: 1,
			Effects: []Effect{{SpellID: 700, Type: dbcenums.E_HEAL, BasePoints: 900, SPCoef: 0.6,
				Target: [2]dbcenums.ImplicitTarget{57, 0}}},
			Powers: []Power{{Type: 2, Cost: 40}},
		},
		{
			ID: 750, Name: "Percent Cost Form", School: 1,
			Effects: []Effect{{SpellID: 750, Type: dbcenums.E_APPLY_AURA, Target: [2]dbcenums.ImplicitTarget{1, 0}}},
			Powers:  []Power{{Type: 0, CostPct: 4}},
		},
		{
			ID: 800, Name: "Bleed", School: 1, DefenseType: 2, SpellLevel: 50, DurationMs: 21000,
			Mechanic:              dbcenums.MECHANIC_BLEED,
			GCDMs:                 1500,
			StartRecoveryCategory: 133,
			MaxStack:              5,
			Effects: []Effect{
				{SpellID: 800, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PERIODIC_DAMAGE,
					BasePoints: 70, PPL: 1, BaseLevel: 50, PeriodMs: 3000, SPCoef: 0.1, Target: [2]dbcenums.ImplicitTarget{6, 0}},
			},
			Powers: []Power{{Type: 1, Cost: 100}},
		},
		{
			ID: 900, Name: "Permanent Charge Aura", DurationMs: -1, ProcCharges: 3,
			Effects: []Effect{{SpellID: 900, Type: dbcenums.E_APPLY_AURA, Target: [2]dbcenums.ImplicitTarget{1, 0}}},
		},
		{
			ID: 1000, Name: "Tickless Aura", DurationMs: 10000,
			Effects: []Effect{{SpellID: 1000, Type: dbcenums.E_APPLY_AURA, Target: [2]dbcenums.ImplicitTarget{1, 0}}},
		},
	}
}

// The timers are all a resolved config needs off the unit, and both are created on demand.
func testUnit() *core.Unit {
	return &core.Unit{}
}

// A row flagged Fixed Travel Time states its flight time in seconds as its Speed, so its missile has no
// speed and flies that long at any distance, unless its MinDuration is longer.
func TestAFixedTravelTimeRowFliesForItsSpeedInSeconds(t *testing.T) {
	var fixed [17]uint32
	fixed[dbcenums.ATTR_INDEX_EX_9] = dbcenums.ATTR_EX_9_FIXED_TRAVEL_TIME
	withRows(t, []Spell{
		{ID: 1, Name: "Fixed", Speed: 4, Attr: fixed},
		{ID: 2, Name: "Fixed Under Its Minimum", Speed: 4, MinDuration: 5, Attr: fixed},
		{ID: 3, Name: "Flying", Speed: 4, MinDuration: 1},
	})

	for _, tc := range []struct {
		id      int32
		speed   float64
		minimum time.Duration
	}{
		{1, 0, 4 * time.Second},
		{2, 0, 5 * time.Second},
		{3, 4, time.Second},
	} {
		config := SpellConfig(testUnit(), Find(tc.id))
		if config.MissileSpeed != tc.speed || config.MissileMinDuration != tc.minimum {
			t.Errorf("%s: MissileSpeed %v, MissileMinDuration %v; want %v and %v", Find(tc.id).Name,
				config.MissileSpeed, config.MissileMinDuration, tc.speed, tc.minimum)
		}
	}
}

func TestSpellConfigFromARow(t *testing.T) {
	withRows(t, resolverRows())
	config := SpellConfig(testUnit(), Find(400))

	if config.ActionID != (core.ActionID{SpellID: 400}) {
		t.Errorf("ActionID = %v, want spell 400", config.ActionID)
	}
	if config.Rank != 7 {
		t.Errorf("Rank = %d, want 7", config.Rank)
	}
	if config.SpellSchool != core.SpellSchoolFrost {
		t.Errorf("SpellSchool = %v, want frost", config.SpellSchool)
	}
	if config.DefenseType != core.DefenseTypeMagic {
		t.Errorf("DefenseType = %v, want magic", config.DefenseType)
	}
	if config.MissileSpeed != 24 {
		t.Errorf("MissileSpeed = %v, want 24", config.MissileSpeed)
	}
	if config.MinRange != 8 || config.MaxRange != 30 {
		t.Errorf("range = %v-%v, want 8-30", config.MinRange, config.MaxRange)
	}
	if config.Cast.DefaultCast.CastTime != time.Millisecond*2500 {
		t.Errorf("cast time = %v, want 2.5s", config.Cast.DefaultCast.CastTime)
	}
	if config.Cast.DefaultCast.GCD != time.Millisecond*1500 {
		t.Errorf("GCD = %v, want 1.5s", config.Cast.DefaultCast.GCD)
	}
	if config.Cast.IgnoreHaste {
		t.Error("a magic spell should not ignore haste")
	}
	if config.ManaCost.FlatCost != 425 || config.ManaCost.BaseCostPercent != 19 {
		t.Errorf("mana cost = %d flat, %v%%, want 425 and 19", config.ManaCost.FlatCost,
			config.ManaCost.BaseCostPercent)
	}
	if config.Flags != 0 {
		t.Errorf("flags = %v, want none", config.Flags)
	}
}

// A rank the game files do not state answers 0, which is what the APL UI reads as "no ranks".
func TestSpellConfigRank(t *testing.T) {
	withRows(t, resolverRows())

	if got := SpellConfig(testUnit(), Find(100)).Rank; got != 3 {
		t.Errorf("Rank of a ranked spell = %d, want 3", got)
	}
	if got := SpellConfig(testUnit(), Find(500)).Rank; got != 0 {
		t.Errorf("Rank of a spell whose subtext is a word = %d, want 0", got)
	}
	if got := SpellConfig(testUnit(), Find(600)).Rank; got != 0 {
		t.Errorf("Rank of a spell with no subtext = %d, want 0", got)
	}
}

func TestSpellConfigBleedMultipliers(t *testing.T) {
	withRows(t, resolverRows())

	config := SpellConfig(testUnit(), Find(800))
	if config.DamageMultiplier != 1 || config.ThreatMultiplier != 1 {
		t.Errorf("multipliers = %v damage, %v threat, want 1 and 1 off the game files' bleed mechanic",
			config.DamageMultiplier, config.ThreatMultiplier)
	}
	if config := SpellConfig(testUnit(), Find(100)); config.DamageMultiplier != 0 {
		t.Errorf("damage multiplier = %v, want none on a spell without the mechanic", config.DamageMultiplier)
	}
}

func TestSpellConfigRageCostAndRefund(t *testing.T) {
	withRows(t, resolverRows())

	rage := SpellConfig(testUnit(), Find(100)).RageCost
	if rage.Cost != 30 {
		t.Errorf("rage cost = %d, want 30 off the game files' 0-1000 bar", rage.Cost)
	}
	if rage.Refund != 0.8 {
		t.Errorf("rage refund = %v, want 0.8", rage.Refund)
	}

	if refund := SpellConfig(testUnit(), Find(200)).RageCost.Refund; refund != 0 {
		t.Errorf("rage refund without the game files' flag = %v, want 0", refund)
	}
}

func TestSpellConfigEnergyAndFocusCosts(t *testing.T) {
	withRows(t, resolverRows())

	energy := SpellConfig(testUnit(), Find(600)).EnergyCost
	if energy.Cost != 60 || energy.Refund != 0.8 {
		t.Errorf("energy cost = %d, refund %v, want 60 and 0.8", energy.Cost, energy.Refund)
	}

	focus := SpellConfig(testUnit(), Find(700)).FocusCost
	if focus.Cost != 40 || focus.Refund != 0 {
		t.Errorf("focus cost = %d, refund %v, want 40 and 0", focus.Cost, focus.Refund)
	}
}

// A spell the row states no cost for leaves every cost option empty.
func TestSpellConfigWithoutAPowerRow(t *testing.T) {
	withRows(t, resolverRows())
	config := SpellConfig(testUnit(), Find(300))

	if config.ManaCost.FlatCost != 0 || config.RageCost.Cost != 0 || config.EnergyCost.Cost != 0 ||
		config.FocusCost.Cost != 0 {
		t.Errorf("a row with no power row resolved to a cost: %+v", config)
	}
}

func TestSpellConfigOwnCooldown(t *testing.T) {
	withRows(t, resolverRows())
	unit := testUnit()
	config := SpellConfig(unit, Find(100))

	if config.Cast.CD.Timer == nil {
		t.Fatal("a spell with its own cooldown got no timer")
	}
	if config.Cast.CD.Timer == unit.CategoryTimer(int32(testCategory)) {
		t.Error("a spell with its own cooldown should not run off a category timer")
	}
	if config.Cast.CD.Duration != time.Second*6 {
		t.Errorf("cooldown = %v, want 6s", config.Cast.CD.Duration)
	}
	if config.Cast.SharedCD.Timer != nil {
		t.Error("a spell with no category cooldown got a shared cooldown")
	}
}

// Two spells in one category run off the unit's timer for that category.
func TestSpellConfigCategoryCooldownSharesTheTimer(t *testing.T) {
	withRows(t, resolverRows())
	unit := testUnit()
	config := SpellConfig(unit, Find(200))

	if config.Cast.CD.Timer != unit.CategoryTimer(int32(testCategory)) {
		t.Error("a category cooldown did not run off the unit's timer for that category")
	}
	if config.Cast.CD.Duration != time.Second*6 {
		t.Errorf("category cooldown = %v, want 6s", config.Cast.CD.Duration)
	}
	if config.Cast.SharedCD.Timer != nil {
		t.Error("a spell with only a category cooldown got a shared cooldown too")
	}
}

// A category cooldown with no category to share is the spell's own recovery time.
func TestSpellConfigCategoryCooldownWithoutACategory(t *testing.T) {
	withRows(t, resolverRows())
	unit := testUnit()
	config := SpellConfig(unit, Find(350))

	if config.Cast.CD.Timer == nil {
		t.Fatal("a categoryless category cooldown got no timer")
	}
	if config.Cast.CD.Timer == unit.CategoryTimer(0) {
		t.Error("a categoryless category cooldown landed on the timer for category 0")
	}
	if config.Cast.CD.Duration != time.Second*8 {
		t.Errorf("cooldown = %v, want 8s", config.Cast.CD.Duration)
	}
}

// A spell with both keeps its own cooldown and shares the category's.
func TestSpellConfigBothCooldowns(t *testing.T) {
	withRows(t, resolverRows())
	unit := testUnit()
	config := SpellConfig(unit, Find(300))

	if config.Cast.CD.Timer == nil || config.Cast.CD.Timer == unit.CategoryTimer(int32(testCategory)) {
		t.Error("the spell's own cooldown did not get a timer of its own")
	}
	if config.Cast.CD.Duration != time.Second*30 {
		t.Errorf("cooldown = %v, want 30s", config.Cast.CD.Duration)
	}
	if config.Cast.SharedCD.Timer != unit.CategoryTimer(int32(testCategory)) {
		t.Error("the shared cooldown did not run off the unit's timer for the category")
	}
	if config.Cast.SharedCD.Duration != time.Second*6 {
		t.Errorf("shared cooldown = %v, want 6s", config.Cast.SharedCD.Duration)
	}
}

// Only the game files' global cooldown category spends the GCD.
func TestSpellConfigGCDGate(t *testing.T) {
	withRows(t, resolverRows())

	if got := SpellConfig(testUnit(), Find(100)).Cast.DefaultCast.GCD; got != time.Millisecond*1500 {
		t.Errorf("GCD of a row in category 133 = %v, want 1.5s", got)
	}
	if got := SpellConfig(testUnit(), Find(700)).Cast.DefaultCast.GCD; got != 0 {
		t.Errorf("GCD of a row outside category 133 = %v, want none", got)
	}
}

// A cast that spends a resource without a GCD or a cast time reads as a cast rather than as the empty
// one core takes for a proc, whether the row states the cost flat or as a share of the bar.
func TestSpellConfigMarksACostedOffGCDCastNonEmpty(t *testing.T) {
	withRows(t, resolverRows())

	if !SpellConfig(testUnit(), Find(700)).Cast.DefaultCast.NonEmpty {
		t.Error("a focus cost with no GCD resolved to an empty cast")
	}
	if !SpellConfig(testUnit(), Find(750)).Cast.DefaultCast.NonEmpty {
		t.Error("a cost stated only as a percentage of the bar resolved to an empty cast")
	}
	if SpellConfig(testUnit(), Find(500)).Cast.DefaultCast.NonEmpty {
		t.Error("a passive with no cost should be left an empty cast")
	}
}

// Haste shortens a cast and its GCD by school, not by the hit table the row files the spell under:
// the game files' shouts are physical spells with the magic defense type.
func TestSpellConfigIgnoreHasteFollowsTheSchool(t *testing.T) {
	withRows(t, resolverRows())

	if !SpellConfig(testUnit(), Find(250)).Cast.IgnoreHaste {
		t.Error("a physical spell with the magic defense type did not ignore haste")
	}
	if !SpellConfig(testUnit(), Find(100)).Cast.IgnoreHaste {
		t.Error("a physical ability did not ignore haste")
	}
	if SpellConfig(testUnit(), Find(400)).Cast.IgnoreHaste {
		t.Error("a frost spell ignored haste")
	}
}

func TestSpellConfigFlagsFromTheRow(t *testing.T) {
	withRows(t, resolverRows())

	passive := SpellConfig(testUnit(), Find(500)).Flags
	if !passive.Matches(core.SpellFlagPassiveSpell) {
		t.Error("a passive row did not resolve to SpellFlagPassiveSpell")
	}
	if !passive.Matches(core.SpellFlagHelpful) {
		t.Error("an effect targeting the caster did not resolve to SpellFlagHelpful")
	}

	channel := SpellConfig(testUnit(), Find(600)).Flags
	if !channel.Matches(core.SpellFlagChanneled) {
		t.Error("a channeled row did not resolve to SpellFlagChanneled")
	}
	if !channel.Matches(core.SpellFlagSuppressWeaponProcs) {
		t.Error("a row with Suppress Weapon Procs did not resolve to the flag")
	}
	if channel.Matches(core.SpellFlagHelpful) {
		t.Error("an effect targeting an enemy resolved to SpellFlagHelpful")
	}

	if !SpellConfig(testUnit(), Find(700)).Flags.Matches(core.SpellFlagHelpful) {
		t.Error("an effect targeting a raid member did not resolve to SpellFlagHelpful")
	}
}

func TestSpellConfigMeleeOption(t *testing.T) {
	withRows(t, resolverRows())
	config := SpellConfig(testUnit(), Find(100), Melee(core.ProcMaskMeleeMHSpecial))

	if config.ProcMask != core.ProcMaskMeleeMHSpecial {
		t.Errorf("proc mask = %v, want the melee main-hand special mask", config.ProcMask)
	}
	if !config.Flags.Matches(core.SpellFlagMeleeMetrics) || !config.Flags.Matches(core.SpellFlagAPL) {
		t.Errorf("flags = %v, want the melee metrics and APL flags", config.Flags)
	}
	if config.DamageMultiplier != 1 || config.ThreatMultiplier != 1 {
		t.Errorf("multipliers = %v and %v, want 1 and 1", config.DamageMultiplier, config.ThreatMultiplier)
	}
	if !config.Cast.IgnoreHaste {
		t.Error("a melee ability should ignore haste")
	}
	if config.BonusCoefficient != 0 {
		t.Errorf("bonus coefficient = %v, want none on a melee ability", config.BonusCoefficient)
	}
}

func TestSpellConfigMagicOption(t *testing.T) {
	withRows(t, resolverRows())
	config := SpellConfig(testUnit(), Find(400), Magic(core.ProcMaskSpellDamage))

	if config.ProcMask != core.ProcMaskSpellDamage {
		t.Errorf("proc mask = %v, want the spell damage mask", config.ProcMask)
	}
	if !config.Flags.Matches(core.SpellFlagAPL) || config.Flags.Matches(core.SpellFlagMeleeMetrics) {
		t.Errorf("flags = %v, want the APL flag and no melee metrics", config.Flags)
	}
	if config.BonusCoefficient != 0.814 {
		t.Errorf("bonus coefficient = %v, want the damage effect's 0.814", config.BonusCoefficient)
	}

	// A spell with nothing but a heal takes its coefficient off the heal.
	if got := SpellConfig(testUnit(), Find(700), Magic(core.ProcMaskSpellHealing)).BonusCoefficient; got != 0.6 {
		t.Errorf("bonus coefficient of a heal = %v, want 0.6", got)
	}
}

func TestSpellConfigNoThreatRow(t *testing.T) {
	withRows(t, []Spell{{
		ID: 1100, Name: "Silent Bolt", School: 8, DefenseType: 1,
		Attr:    [17]uint32{dbcenums.ATTR_INDEX_EX_1: dbcenums.ATTR_EX_1_NO_THREAT},
		Effects: []Effect{{SpellID: 1100, Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 100, Target: [2]dbcenums.ImplicitTarget{6, 0}}},
	}})

	if got := SpellConfig(testUnit(), Find(1100), Magic(core.ProcMaskSpellDamage)).ThreatMultiplier; got != 0 {
		t.Errorf("Magic threat multiplier = %v, want 0 on a row with ATTR_EX_1_NO_THREAT", got)
	}
	if got := SpellConfig(testUnit(), Find(1100), Melee(core.ProcMaskMeleeMHSpecial)).ThreatMultiplier; got != 0 {
		t.Errorf("Melee threat multiplier = %v, want 0 on a row with ATTR_EX_1_NO_THREAT", got)
	}
}

// Area damage drops the row's range to the target: around the caster it reaches its radius, as a
// cast and not as a proc, and around the point the triggering spell chose it reaches anywhere.
func TestSpellConfigAreaAroundACasterOrPointReachesItsRadiusOrAnywhere(t *testing.T) {
	for _, tc := range []struct {
		centre       dbcenums.ImplicitTarget
		cast, asProc float64
	}{
		{dbcenums.TARGET_DEST_CASTER, 8, 0},
		{dbcenums.TARGET_SRC_CASTER, 8, 0},
		{dbcenums.TARGET_DEST_DEST, 0, 0},
	} {
		withRows(t, []Spell{{
			ID: 1200, Name: "Nova", School: 4, DefenseType: 1, MaxRange: 30,
			Effects: []Effect{{SpellID: 1200, Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 100, RadiusMax: 8,
				Target: [2]dbcenums.ImplicitTarget{tc.centre, dbcenums.TARGET_UNIT_DEST_AREA_ENEMY}}},
		}})
		if got := SpellConfig(testUnit(), Find(1200)).MaxRange; got != tc.cast {
			t.Errorf("max range of area damage centred on target %d = %v, want %v", tc.centre, got, tc.cast)
		}
		if got := SpellConfig(testUnit(), Find(1200), Proc()).MaxRange; got != tc.asProc {
			t.Errorf("max range of a proc's area damage centred on target %d = %v, want %v", tc.centre, got, tc.asProc)
		}
	}
}

// The rows' own areas: a cast centred on the caster or a cone from it reaches its radius, from
// whichever radius slot the row states it in, even where the row states a range, and through the row
// its periodic trigger aura fires. A proc of one has no range, and a DEST_DEST area, a cone on a row
// with a range of its own and a spell with no area keep what they had.
func TestSpellConfigReachOfTheGameFilesAreas(t *testing.T) {
	withGeneratedStore(t)

	for _, tc := range []struct {
		id, areaRow int32
		effect      int
	}{
		{10202, 10202, 1}, // Arcane Explosion, [22,15], radius in the second slot
		{10161, 10161, 2}, // Cone of Cold, a cone
		{20924, 20924, 1}, // Consecration, a dummy at [18,16]
		{18818, 18818, 1}, // Flamestrike, [18,16], states 30 yd
		{11684, 11682, 1}, // Hellfire, whose periodic trigger aura fires Hellfire Effect
	} {
		row := MustFind(tc.id)
		radius := float64(MustFind(tc.areaRow).EffectN(tc.effect).RadiusMax)
		if radius == 0 {
			t.Fatalf("%d effect %d states no radius", tc.areaRow, tc.effect)
		}
		if got := SpellConfig(testUnit(), row).MaxRange; got != radius {
			t.Errorf("%d %s cast reaches %v yd, want its %v yd radius", tc.id, row.Name, got, radius)
		}
		if got := SpellConfig(testUnit(), row, Proc()).MaxRange; got != 0 {
			t.Errorf("%d %s as a proc reaches %v yd, want no range", tc.id, row.Name, got)
		}
	}

	t.Run("a cone on a row with a range of its own", func(t *testing.T) {
		forkedLightning := Find(11828) // a 45 yd cone on a row stating 30 yd
		if forkedLightning == Nil {
			t.Skip("spell 11828 is not in the store; this case runs again once the live game files hold it")
		}
		if got := SpellConfig(testUnit(), forkedLightning).MaxRange; got != 30 {
			t.Errorf("%d %s reaches %v yd, want its row's own 30", forkedLightning.ID, forkedLightning.Name, got)
		}
	})

	fireNova := MustFind(408428) // around the totem's point (DEST_DEST)
	if fireNova.MaxRange == 0 {
		t.Fatalf("%d %s states no range of its own", fireNova.ID, fireNova.Name)
	}
	if got := SpellConfig(testUnit(), fireNova).MaxRange; got != 0 {
		t.Errorf("%d %s reaches %v yd, want no range", fireNova.ID, fireNova.Name, got)
	}
	if got := SpellConfig(testUnit(), MustFind(20511)).MaxRange; got != 10 {
		t.Errorf("Intimidating Shout reaches %v yd, want its row's own 10", got)
	}
}

func TestSpellConfigProcFlagsAndTag(t *testing.T) {
	withRows(t, resolverRows())
	config := SpellConfig(testUnit(), Find(100), Proc(), Flags(core.SpellFlagIgnoreResists), Tag(2))

	if !config.Flags.Matches(core.SpellFlagPassiveSpell) ||
		!config.Flags.Matches(core.SpellFlagNoOnCastComplete) {
		t.Errorf("flags = %v, want the passive and no-on-cast-complete flags", config.Flags)
	}
	if config.Flags.Matches(core.SpellFlagNoMetrics) {
		t.Error("a proc sub-spell keeps its metrics")
	}
	if !config.Flags.Matches(core.SpellFlagIgnoreResists) {
		t.Error("the caller's own flag was not kept")
	}
	if config.ActionID.Tag != 2 {
		t.Errorf("tag = %d, want 2", config.ActionID.Tag)
	}
	if config.RageCost.Cost != 0 || config.Cast.DefaultCast.GCD != 0 || config.Cast.CD.Timer != nil {
		t.Errorf("a proc sub-spell has no cost, GCD or cooldown of its own, got %d rage, %v GCD, timer %v",
			config.RageCost.Cost, config.Cast.DefaultCast.GCD, config.Cast.CD.Timer)
	}
}

// Melee and Magic put a spell in the rotation, and Proc takes the sub-spell they resolve back out of
// it: the flags the warrior's own proc sub-spells carry today.
func TestSpellConfigProcLeavesTheRotation(t *testing.T) {
	withRows(t, resolverRows())
	config := SpellConfig(testUnit(), Find(100), Melee(core.ProcMaskMeleeMHSpecial), Proc())

	want := core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete
	if config.Flags != want {
		t.Errorf("flags = %v, want %v", config.Flags, want)
	}
}

// SpellEquippedItems and the not-in-combat attribute go into the cast requirement: Spearing Strike
// 1310222's two-handers in the main hand, Shield Bash 1672's shield, Charge 11578's out of combat.
// Armor that is not a shield asks nothing of a cast.
func TestSpellConfigCastRequirementFromEquipment(t *testing.T) {
	twoHanders := int32(dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_2H |
		dbcenums.ITEM_SUBCLASS_BIT_WEAPON_POLEARM | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_2H | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_STAFF)
	withRows(t, []Spell{
		{ID: 2, Name: "Leather", EquipClass: dbcenums.ITEM_CLASS_ARMOR, EquipSubclass: 0x4},
		{ID: 1672, Name: "Shield Bash", EquipClass: dbcenums.ITEM_CLASS_ARMOR, EquipSubclass: dbcenums.ITEM_SUBCLASS_BIT_ARMOR_SHIELD},
		{ID: 11578, Name: "Charge", Attr: [17]uint32{dbcenums.ATTR_INDEX_BASE: dbcenums.ATTR_NOT_IN_COMBAT}},
		{ID: 1310222, Name: "Spearing Strike", EquipClass: dbcenums.ITEM_CLASS_WEAPON, EquipSubclass: twoHanders,
			Attr: [17]uint32{dbcenums.ATTR_INDEX_EX_3: dbcenums.ATTR_EX_3_REQ_MAIN_HAND}},
	})

	for _, c := range []struct {
		id   int32
		want core.CastRequirement
	}{
		{1310222, core.CastRequirement{Weapon: core.EquipRequirement{SubclassMask: twoHanders, Hand: core.ProcMaskMeleeMH}}},
		{1672, core.CastRequirement{Shield: true}},
		{11578, core.CastRequirement{NotInCombat: true}},
		{2, core.CastRequirement{}},
	} {
		if got := SpellConfig(testUnit(), Find(c.id)).CastRequirement; got != c.want {
			t.Errorf("%s's cast requirement = %+v, want %+v", Find(c.id).Name, got, c.want)
		}
	}
}
