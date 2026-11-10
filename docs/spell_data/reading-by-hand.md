# Reading a row by hand

A class reads the store by hand in the files that `grep -l 'spelldata.SpellConfig\|spelldata.Parse' sim/<class>/*.go`
does not list. A class file names the ladder it
needs, picks a rank, and reads every field the registration wants straight into a plain expression,
with no resolver in between. In those files **a row is a source of numbers only**: they use
no `spelldata.SpellConfig`, `AuraConfig`, `DotConfig`, `ProcTrigger`, `ParseEffects` or
`ParseStatic`. They build `core.SpellConfig` and `core.DotConfig` by hand, field by field. These are
the same shapes that [Building an ability](building-an-ability.md) and
[Auras and dots](auras-and-dots.md) describe for the resolvers, filled in without one.

For those files the generated file is ladders too (see
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
| `rank.ReagentCount(item)`                          | how many of that item each cast consumes, off `Reagents` (SpellReagents); 0 for an item it does not take                                                                                |
| `rank.GCD()`, `rank.CastTime()`                    | the global cooldown and the cast time                                                                                                                                                   |
| `max(rank.Cooldown(), rank.CategoryCooldown())`    | the row's own cooldown, or the one it shares with a category                                                                                                                            |
| `rank.Duration()`                                  | an aura or dot's length; the game files' -1 becomes `core.NeverExpires`                                                                                                                 |
| `rank.SpellSchool()`, `rank.DefenseTypeCore()`     | already in core's enums: the accessor converts the game files' byte                                                                                                                     |
| `float64(rank.MaxRange)`, `float64(rank.MinRange)` | the plain fields, in yards                                                                                                                                                              |

The mage's Frostbolt config in `sim/mage/frostbolt.go` takes the cost and the cast from its caller
and reads the class flags, the missile speed and the spell power share off the row (the registration
that calls it is still a stub):

```go
func (mage *Mage) frostBoltConfig(config core.SpellConfig) core.SpellConfig {
	return core.SpellConfig{
		ActionID:       config.ActionID,
		SpellSchool:    core.SpellSchoolFrost,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          config.Flags,
		ClassSpellMask: MageSpellFrostbolt,
		ClassFlags:     frostboltRank.ClassFlags,
		MissileSpeed:   float64(frostboltRank.Speed),

		ManaCost: config.ManaCost,
		Cast:     config.Cast,

		DamageMultiplier: config.DamageMultiplier,
		BonusCoefficient: frostboltRank.DamageEffect().Coeff(),
		ThreatMultiplier: 1,

		ApplyEffects: config.ApplyEffects,
	}
}
```

**An amount is `Average()`, or `Roll(sim)` where the class rolls it.** These classes read every
amount as `Average`, both ends of a `(low, high)` alike. A row can still carry a spread (Frostbolt's
`Variance` is 0.105), and `Roll(sim)`/`Min`/`Max` read it where a class asks for one. A ported number
reads `Average` so that it stays equal to the one it replaces, not because the data has no spread.

**Name the effect by role where one applies; by position where the row hides it behind another
effect.** `DamageEffect()`, `HealEffect()`, `EnergizeEffect()` and `PeriodicEffect()` answer the first
effect of their kind, which is right wherever a spell has only one. Hellfire's tick is an exception.
`PeriodicEffect()` would find the `A_PERIODIC_TRIGGER_SPELL` at position 1 (the trigger that fires the
Hellfire Effect spell each tick) and miss the damage tick itself at position 2. From
`sim/warlock/hellfire.go`:

```go
var hellfireRank = spellData.Hellfire.Highest()

// Effect 1 is the periodic trigger that fires the Hellfire Effect spell each tick; effect 2 is
// the damage tick.
var hellfireTick = hellfireRank.EffectN(2)
var hellFireCoeff = hellfireTick.Coeff()
```

**When the spell does its work through another spell, read that spell's row, not the row that names
it.** Where the family has its own `XTriggered` ladder for the spell the trigger fires, that ladder is
the edge. Failing that, `row.Triggered()[0]` (an actual `TriggerID` edge) or `row.Refs()[0]` (a
tooltip reference) reach the same spell. A bare `spelldata.MustFind(id)` is the last resort, with a
comment naming why no other edge is unique. Mutilate's two strikes are the spell its cast triggers,
and the family's own `MutilateTriggered` ladder names it, so `newMutilateHitSpell` in
`sim/rogue/talents_assassination.go` takes the strikes' class flags from there:

```go
ClassFlags:     spellData.MutilateTriggered.Highest().ClassFlags,
```

A periodic effect works the same way: where an aura triggers a damage spell each tick, the damage is
that spell's `DamageEffect()`, and the tick length still comes off the aura's own row, from the
periodic dummy that carries no damage of its own (`Effect(dbcenums.A_PERIODIC_DUMMY, 0).Period()`).

**`EffectAt(n)` counts from 1 by position, like `EffectN`; it does not follow the game files'
`EffectIndex`.** `EffectAt(n+1)` lines up with game files index `n` only where the row's indices run
contiguously from 0. Where they do not, or where two of a talent's effects share an aura and a misc
value, `EffectAt` takes the position. Otherwise `Effect(aura, misc)` names the effect, as the mage's
Wand Specialization does in `sim/mage/talents_arcane.go`:

```go
FloatValue: spellData.WandSpecialization.Effect(dbcenums.A_MOD_DAMAGE_PERCENT_DONE, schools).FractionAt(mage.Talents.WandSpecialization),
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
in a variable the `OnTick` closure captures and never re-read at the tick. A finisher reads its combo
points in `ApplyEffects`, before `SpendComboPoints` runs them to zero; by the time `OnTick` runs, the
count is gone. The tick itself is a package-level read of the row, as in `sim/druid/rip.go`:

```go
var ripRank = spellData.Rip.Highest()
var ripTick = ripRank.PeriodicEffect()
```

**A hand number is a Go literal with a review comment.** A hand-supplied threat number, PPM or
coefficient keeps the same marker a resolver-built config carries. `sim/warrior/hamstring.go` puts it
above the resolver call that writes the 1:

```go
// TODO: Manual review needed -- the game files state no threat coefficient; 1 until measured in game.
config := spelldata.SpellConfig(&warrior.Unit, hamstringRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
```

A hand-built `core.SpellConfig` has no resolver to build it first. It carries the same comment on the
same field, written as a struct-literal line:

```go
// TODO: Manual review needed -- the game files state no threat coefficient; 1 until measured in game.
ThreatMultiplier: 1,
```

## Worked examples

### Consecration's borrowed tick

Each rank's tooltip names one spell for both ticks, and the family's own `ConsecrationTriggered`
ladder carries those spells rank for rank (`Refs()[0]` reaches the same row). The two ticks sit on its
first two effects and are read by position, since both are school damage and `DamageEffect()` cannot
tell them apart. Effect 1 is the base tick. Effect 2 is the bonus its first four targets take, with its
own spell power share. The periodic dummy on Consecration's own row states no tick: its period times
the ticks, and its points are how many targets take the bonus. The paladin registers the spell
through `spelldata.SpellConfig`, but it reads the ticks by hand. From `sim/paladin/consecration.go`:

```go
tickSpell := spellData.ConsecrationTriggered.Rank(n)
tick := tickSpell.EffectN(1)
bonus := tickSpell.EffectN(2)

dummy := rank.Effect(dbcenums.A_PERIODIC_DUMMY, 0)
bonusTargets := int(dummy.Average())
tickLength := dummy.Period()
numberOfTicks := int32(rank.Duration() / tickLength)

// The bonus scales on its own coefficient, so it is added to the base damage here rather than
// through the dot's, which is the base tick's. Consecrated Ground marks the same targets.
dealTick := func(sim *core.Simulation, dot *core.Dot, firstTick bool) {
	for i, target := range sim.Encounter.ActiveTargetUnits {
		damage := tick.Average()
		if i < bonusTargets {
			damage += bonus.Average() + bonus.Coeff()*dot.Spell.BonusDamage(dot.Spell.Unit.AttackTables[target.UnitIndex])
			if firstTick && paladin.consecratedGroundAuras != nil {
				paladin.consecratedGroundAuras.Get(target).Activate(sim)
			}
		}
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, dot.Spell.OutcomeTickMagicCrit)
	}
}
```

There is no `OnSnapshot`: every tick reads `tick.Average()` on the stats in force when it lands.

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
