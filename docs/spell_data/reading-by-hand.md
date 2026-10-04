# Reading a row by hand

Rogue, warlock, mage, druid, priest, shaman and paladin read the store by hand. A class file names the
ladder it needs, picks a rank, and reads every field the registration wants straight into a plain
expression, with no resolver in between. In these seven classes' files **a row is a source of numbers
only**: they use no `spelldata.SpellConfig`, `AuraConfig`, `DotConfig`, `ProcTrigger`, `ParseEffects`
or `ParseStatic`. They build `core.SpellConfig` and `core.DotConfig` by hand, field by field. These are
the same shapes that [Building an ability](building-an-ability.md) and
[Auras and dots](auras-and-dots.md) describe for the resolvers, filled in without one.

For these classes the generated file is ladders too (see
[A class file is ladders](store.md#a-class-file-is-ladders)). `spellData.ShadowWordPain` is a
`spelldata.Ladder`, `.Highest()`/`.ByID(id)`/`.Rank(n)` answer a `*spelldata.Spell`, and `.Each(fn)`
walks every rank in declaration order:

```go
MindBlastRankMap.Each(func(_ int32, rank *spelldata.Spell) {
	priest.registerMindBlastSpell(rank, mindblastCDTimer)
})
```

With no resolver to read the row for the caller, a registration reads these values off it:

|                                                    |                                                                                                                                                                                         |
| -------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `rank.ID`, `rank.RankNumber()`                     | the `ActionID` and the `Rank` field a `core.SpellConfig` wants; `RankNumber()` is `NameSubtext_lang`'s "Rank N", 0 where the game files state none                                      |
| `rank.Cost()`                                      | the cost off the first power the row states, in the sim's units (rage already divided by ten); it answers a `float64`, so an `int32` field like `ManaCostOptions.FlatCost` needs a cast |
| `rank.GCD()`, `rank.CastTime()`                    | the global cooldown and the cast time                                                                                                                                                   |
| `max(rank.Cooldown(), rank.CategoryCooldown())`    | the row's own cooldown, or the one it shares with a category                                                                                                                            |
| `rank.Duration()`                                  | an aura or dot's length; the game files' -1 becomes `core.NeverExpires`                                                                                                                 |
| `rank.SpellSchool()`, `rank.DefenseTypeCore()`     | already in core's enums: the accessor converts the game files' byte                                                                                                                     |
| `float64(rank.MaxRange)`, `float64(rank.MinRange)` | the plain fields, in yards                                                                                                                                                              |

Straight off `sim/shaman/shocks.go`:

```go
func (shaman *Shaman) newShockSpellConfig(rank *spelldata.Spell, spellSchool core.SpellSchool, shockTimer *core.Timer) core.SpellConfig {
	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: rank.ID},
		SpellSchool: spellSchool,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagShamanSpell | SpellFlagShock | core.SpellFlagAPL | SpellFlagInstant,
		MaxRange:    float64(rank.MaxRange),

		ManaCost: core.ManaCostOptions{
			FlatCost: int32(rank.Cost()),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: rank.GCD()},
			CD: core.Cooldown{
				Timer:    shockTimer,
				Duration: max(rank.Cooldown(), rank.CategoryCooldown()),
			},
		},

		DamageMultiplier: 1,
		BonusCoefficient: rank.DamageEffect().Coeff(),
		ThreatMultiplier: 1,
	}
}
```

**An amount is `Average()`, or `Roll(sim)` where the class rolls it.** The
seven classes ported first read every amount as `Average`, both ends of a `(low, high)` alike, because
the literals they replaced never rolled. Paladin's literals did roll, so its damage and heals take
`Roll` and its per-hit and aura numbers take `Average`. A row can still carry a spread (Frostbolt's
`Variance` is 0.105), and `Roll(sim)`/`Min`/`Max` read it where a class asks for one. A ported number
reads `Average` so that it stays equal to the one it replaces, not because the data has no spread.

**Name the effect by role where one applies; by position where the row hides it behind another
effect.** `DamageEffect()`, `HealEffect()`, `EnergizeEffect()` and `PeriodicEffect()` answer the first
effect of their kind, which is right wherever a spell has only one. Hellfire's tick is an exception.
`PeriodicEffect()` would find the `A_PERIODIC_TRIGGER_SPELL` at position 1 (the trigger that fires the
Hellfire Effect spell each tick) and miss the damage tick itself at position 2:

```go
var hellfireRank = spellData.Hellfire.Highest()

// The tick sits at effect position 2: position 1 is the A_PERIODIC_TRIGGER_SPELL that fires the
// Hellfire Effect spell each tick, not the damage tick.
var hellfireTick = hellfireRank.EffectN(2)
var hellFireCoeff = hellfireTick.Coeff()
```

**When the description keeps a tick on another spell, read it off that spell, not the row that names
it.** Where the family has its own `XTriggered` ladder for the spell the trigger fires, that ladder is
the edge. Failing that, `row.Triggered()[0]` (an actual `TriggerID` edge) or `row.Refs()[0]` (a
tooltip reference) reach the same spell. A bare `spelldata.MustFind(id)` is the last resort, with a
comment naming why no other edge is unique. Hurricane's periodic damage sits on the spell its aura
triggers each tick, and the family's own `HurricaneTriggered` ladder names that spell. The tick length
still comes off Hurricane's own row, from the periodic dummy that carries no damage of its own:

```go
tickLength := hurricaneRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0).Period()

// Hurricane's periodic damage is the spell HurricaneTriggered casts each tick.
hurricaneTickSpell := spellData.HurricaneTriggered.Highest()
hurricaneTick := hurricaneTickSpell.DamageEffect()
```

**`EffectAt(n)` counts from 1 by position, like `EffectN`; it does not follow the game files'
`EffectIndex`.** `EffectAt(n+1)` lines up with game files index `n` only where the row's indices run
contiguously from 0. Where they do not, `Effect(aura, misc)` names the effect instead. Rogue's
`PuncturingWounds` talent needs `EffectAt`. Its two crit modifiers share an aura and misc, so
`Effect(aura, misc)` cannot tell them apart. They sit at positions 1 and 3, around the proc trigger at
position 2 (game files index 1), and Mutilate reads `EffectAt(3)`:

```go
// The proc trigger sits between the two crit modifiers, at effect position 2; the Mutilate
// crit bonus is the second of the two, at position 3, and they share an aura and misc so have
// to be indexed rather than named.
FloatValue: spellData.PuncturingWounds.EffectAt(3).ValueAt(rogue.Talents.PuncturingWounds),
```

**A `Ranked` family's rank can carry a per-level gain a `Ladder`'s per-rank readers do not add in.**
`EffectAt(n)` and `Effect(aura, misc)` answer base points across the ranks, which is right for a
`Talent`, whose curve states base points outright. A `Ranked` family (one spell per rank) can still
carry `EffectRealPointsPerLevel` on that effect. Then `ValueAt`/`FractionAt`/`MultiplierAt`/`TenthsAt`
read low wherever the game files scale it past the rank's `BaseLevel`. Read the rank's own effect
through `Average` instead of the ladder's per-rank reader:

```go
X.Rank(rank).EffectN(1).Average()   // not X.EffectAt(1).ValueAt(rank)
```

**A dot ticks on current stats, so there is no `OnSnapshot`.** `OnTick` reads the effect's own
`Average` straight into `CalcAndDealPeriodicDamage` (or `CalcAndDealPeriodicHealing`) every time it
fires, on whatever stats and multipliers are in force then. An input the cast fixes, such as combo
points about to be spent or a stack about to be consumed, is read where the cast reads it. It is kept
in a variable the `OnTick` closure captures and never re-read at the tick. Rip's combo points are read
in `ApplyEffects`, before `SpendComboPoints` runs them to zero. By the time `OnTick` runs, the count
would already be gone:

```go
var cp int32

// ...
	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
		if result.Landed() {
			cp = druid.ComboPoints()
			spell.Dot(target).Apply(sim)
			druid.SpendComboPoints(sim, spell.ComboPointMetrics())
		}
		spell.DealOutcome(sim, result)
	},
	OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		ap := dot.Spell.MeleeAttackPower(target)
		var tickDamage float64
		switch {
		case cp <= 3:
			tickDamage = 990 + 0.18*ap
		case cp == 4:
			tickDamage = 1272 + 0.24*ap
		default: // 5
			tickDamage = 1554 + 0.24*ap
		}
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, tickDamage/6, dot.OutcomeTick)
	},
```

**Hand numbers stay hand numbers: Go literals with a review comment.** A hand-supplied threat
number, PPM or coefficient keeps the same marker a resolver-built config carries. `sim/warrior/hamstring.go` writes it as an assignment on the config the resolver already
built:

```go
// TODO: Manual review needed -- the game files state no threat coefficient; 1 until measured in game.
config.ThreatMultiplier = 1
```

A hand-built `core.SpellConfig` has no resolver to build it first. It carries the same comment on the
same field, written as a struct-literal line:

```go
// TODO: Manual review needed -- the game files state no threat coefficient; 1 until measured in game.
ThreatMultiplier: 1,
```

## Worked examples

### Shadow Word: Pain's tick

`sim/priest/shadow_word_pain.go`:

```go
tick := rank.PeriodicEffect()
tickLength := tick.Period()

priest.RegisterSpell(core.SpellConfig{
	ActionID:       core.ActionID{SpellID: rank.ID},
	SpellSchool:    core.SpellSchoolShadow,
	DefenseType:    core.DefenseTypeMagic,
	ProcMask:       core.ProcMaskSpellDamage,
	Flags:          core.SpellFlagAPL,
	ClassSpellMask: PriestSpellShadowWordPain,
	Rank:           rank.RankNumber(),
	MaxRange:       float64(rank.MaxRange),

	ManaCost: core.ManaCostOptions{
		FlatCost: int32(rank.Cost()),
	},

	Cast: core.CastConfig{
		DefaultCast: core.Cast{GCD: rank.GCD()},
	},

	Dot: core.DotConfig{
		Aura: core.Aura{
			Label: fmt.Sprintf("ShadowWordPain-%d", rank.RankNumber()),
		},
		NumberOfTicks:       int32(rank.Duration() / tickLength),
		TickLength:          tickLength,
		AffectedByCastSpeed: false,
		BonusCoefficient:    tick.Coeff(),

		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			dot.Spell.CalcAndDealPeriodicDamage(sim, target, tick.Average(), dot.OutcomeTick)
		},
	},

	ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
		return spell.CalcPeriodicDamage(sim, target, tick.Average(), spell.OutcomeExpectedMagicHit)
	},
})
```

There is no `OnSnapshot` and no `useSnapshot` branch: the tick is `tick.Average()` wherever it is
asked for, cast or projected.

### Consecration's borrowed tick

Each rank's tooltip names one spell for both ticks, and the family's own `ConsecrationTriggered`
ladder carries those spells rank for rank (`Refs()[0]` reaches the same row). The two ticks sit on its
first two effects and are read by position, since both are school damage and `DamageEffect()` cannot
tell them apart. Effect 1 is the base tick. Effect 2 is the bonus its first four targets take, with its
own spell power share. The periodic dummy on Consecration's own row states no tick: its period times
the ticks, and its points are how many targets take the bonus. From `sim/paladin/consecration.go`:

```go
tickSpell := spellData.ConsecrationTriggered.Rank(n)
tick := tickSpell.EffectN(1)
bonus := tickSpell.EffectN(2)

dummy := rank.Effect(dbcenums.A_PERIODIC_DUMMY, 0)
bonusTargets := int(dummy.Average())
tickLength := dummy.Period()
numberOfTicks := int32(rank.Duration() / tickLength)

dealTick := func(sim *core.Simulation, dot *core.Dot) {
	for i, target := range sim.Encounter.ActiveTargetUnits {
		damage := tick.Average()
		if i < bonusTargets {
			damage += bonus.Average() + bonus.Coeff()*dot.Spell.BonusDamage(dot.Spell.Unit.AttackTables[target.UnitIndex])
		}
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, dot.OutcomeTickMagicHit)
	}
}
```

### A heal and a mana restore

```go
heal := rank.HealEffect().Average()      // both ends, where the class does not roll
mana := rank.EnergizeEffect().Average()  // 0 on a rank with no Energize effect at all
```

### Registering several ranks

Forever's Starfire tops out at rank 7. The spec's own subset ladder in `sim/druid/starfire.go`
therefore names ranks 6 and 7 by their ids and never a rank the table does not carry:

```go
var StarfireRankMap = spelldata.Ranked(spellData.Starfire.Rank(6).ID, spellData.Starfire.Rank(7).ID)
```

### An attack power coefficient

The game files state no attack power coefficient per combo point. Rupture's is a literal table indexed
by combo points, right beside the number it scales, in `sim/rogue/rupture.go`:

```go
func (rogue *Rogue) ruptureDamage(target *core.Unit, comboPoints int32, baseDamage float64, damagePerComboPoint float64) float64 {
	return baseDamage +
		damagePerComboPoint*float64(comboPoints) +
		[]float64{0, 0.01, 0.02, 0.03, 0.03, 0.03}[comboPoints]*rogue.Rupture.MeleeAttackPower(target)
}
```
