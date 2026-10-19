# Building an ability

`spelldata.SpellConfig(unit, row, opts...)` fills the `core.SpellConfig` fields the row states, and
leaves everything else to the caller:

```go
var pummelRank = spellData.Pummel.ByID(6554)

config := spelldata.SpellConfig(&warrior.Unit, pummelRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) { /* ... */ }
warrior.RegisterSpell(config)
```

The row fills the `ActionID`, `Rank` (from "Rank 4", which is what flips `HasRanks` in the APL UI),
`SpellSchool`, `DefenseType`, `ClassFlags`, `MissileSpeed`, `MissileMinDuration` and
`MinRange`/`MaxRange`. A spell that is not passive (no `Proc()`), states no range and has an effect
centred on the caster or a cone from it gets that effect's radius as its `MaxRange`, after the
options: Arcane Explosion and Thunder Clap reach only as far as they hit, and Hellfire, whose damage
sits on a triggered row, sets it by hand. A row with Fixed Travel Time states its flight time in seconds as its Speed,
so it gets no `MissileSpeed` and that time as its `MissileMinDuration`, or its own MinDuration where
that is longer: Dragonfire Potion's 1252554 lands 1 s after use at any distance. It fills the cast:
the cast time, the GCD where the row sits in the global cooldown category, its own cooldown and the
category one it shares. It fills the cost out of the first bar it states, with rage divided off the
0-1000 bar, and the refund the Discount Power On Miss attribute states. `CastRequirement` gets the
forms, caster auras, weapon, shield and out-of-combat state the game files require (see below). That is
why Pummel needs no Berserker Stance condition and Shield Bash needs no shield check. `Flags` gets
what the attributes and targets say:
`SpellFlagPassiveSpell`, `SpellFlagChanneled`, `SpellFlagSuppressWeaponProcs` and `SpellFlagHelpful`.

A bleed row also sets the damage and threat multipliers to 1.

The row does not fill `ApplyEffects`, `ProcMask`, the multipliers on any other row, `ClassSpellMask`,
`ExtraCastCondition`, `Dot`, `RelatedSelfBuff` or the threat numbers the game files do not carry. Nor
does it fill `MaxTargets`, which has no `SpellConfig` field at all, so a caller that caps an area effect
reads `row.MaxTargets` itself. The same goes for `RequiredAreas`, the area group a spell only works in.
`row.AreaType()` names the kind of terrain it stands for (Forest and Grassland, Mountainous, ...), and a
caller gates on `sim.Encounter.InArea(row.AreaType())`. A group that is a single zone reads as
`AreaTypeUnknown`. A unit
is needed for the cooldown timers, so a config is built where the sim has a character rather than at
package init.

|               |                                                                                                                                                                       |
| ------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Melee(mask)` | the proc mask, `SpellFlagMeleeMetrics` and `SpellFlagAPL`, damage and threat multipliers of 1, and `IgnoreHaste`                                                      |
| `Magic(mask)` | the proc mask, `SpellFlagAPL`, the multipliers, and `BonusCoefficient` from the damage effect's spell power share, or the heal's where the spell damages nothing      |
| `Proc()`      | `SpellFlagPassiveSpell` and `SpellFlagNoOnCastComplete`, clears `SpellFlagAPL`, and empties the cast (cast time, GCD, cooldowns), every cost and the cast requirement |
| `Flags(f)`    | ors in flags the game files do not state                                                                                                                              |
| `Tag(n)`      | splits one spell id into several actions                                                                                                                              |

The row is filled first and the options run on top in the order given, so an option sees what the row
put there.

## Where a spell can be cast

`core.CastRequirement` is what the game files require of the caster's form, auras and equipment, of
the target's creature type, and of the ranks of the periodic the cast applies. Core refuses any cast,
`CanCast` or `CanQueue` that breaks it ("wrong form", "missing caster aura", "excluded caster aura",
"missing equipped item", "in combat", "wrong target creature type", "a more powerful spell is
active"). It reads seven game files sources:
`SpellShapeshift`'s mask and exclude mask (`StanceMask`, `StanceExclude`), `SpellAuraRestrictions`
(`CasterAura`, `ExcludeCasterAura`; Tiger's Fury states its Cat Form requirement there, not in a mask),
the not-shapeshifted and castable-in-caster-form attribute bits, `SpellShapeshiftForm`'s stance flag
(`dbcenums.ShapeshiftForm.IsStance`, generated into `sim/core/dbcenums/forms_auto_gen.go`),
`SpellEquippedItems` with the `ATTR_EX_3_REQ_MAIN_HAND` and `ATTR_EX_3_REQ_OFF_HAND` hand bits, the
not-in-combat attribute bit (`ATTR_NOT_IN_COMBAT`, Charge's), and `SpellTargetRestrictions`'
`TargetCreatureType`.

For the caster's form `f`: an excluded form refuses; a form in the mask allows; in a shapeshift that is
not a stance (Cat, Bear) a spell with a mask or the not-shapeshifted bit is refused; with no form or in
a stance (the warrior's, Moonkin, Tree) a spell with a mask is refused unless it is castable in caster
form. A required caster aura must then be active on the caster, matched by `ActionID.SpellID`.

The equipment is read as the spell is cast, so it follows item swaps. A weapon row (item class 2) is
`CastRequirement.Weapon`, a `core.EquipRequirement`. Each hand that a hand bit names must hold a weapon
its subclass and inventory-type masks match (Spearing Strike 1310222 a two-hander in the main hand,
Mutilate a dagger in both). A row with no hand bit is met by any weapon slot (Aimed Shot's bow, gun or
crossbow in the ranged slot). An armor row (item class 4) naming the shield subclass is
`CastRequirement.Shield`, which a shield in the off hand meets; a shield is the only armor a castable
spell asks for. A unit that is not a player meets neither, and a player with no weapon meets no weapon
requirement. A spell
with the not-in-combat bit is refused from the start of the encounter on, so Charge is a prepull
action.

`TargetCreatureType` is a mask of creature types, bit `1 << (type - 1)`, which `spelldata.MobTypes`
turns into `proto.MobType`s; `CastRequirement.TargetMobTypes` holds them as bits `1 << proto.MobType`.
It limits the cast only when the row's first effect targets `TARGET_UNIT_TARGET_ENEMY`, the enemy the
spell is cast on: Chastise's 64 allows a Humanoid, Exorcism's 36 an Undead or a Demon. Elsewhere the
mask names some other unit and the cast is not limited: Bestial Wrath's Beast is the pet it lands on,
and Holy Wrath's Undead and Demon are the enemies its area hits, which the spell's effects pick
themselves. A target of unknown mob type is in no mask and is refused. A weapon swing checks no target.

`CastRequirement.NotUnderHigherRank` is set on a ranked row whose only effect that does something is
the periodic, an effect with a tick period; a dummy or script effect does not count. The cast is
refused while a higher rank of the dot it applies (`Spell.Dot(target)`, which follows
RelatedDotSpell, else the caster's self periodic) from the same caster is up, as
[Auras and dots](auras-and-dots.md) sets out. `Proc()` empties the cast requirement, so a rank
registered with it casts and only its `Dot.Apply` skips.

The class does nothing beyond reporting its form in `Unit.ShapeshiftForm` (the warrior's `setStance`,
the druid's `setForm`). A unit that sets `Unit.AutoUnshift` (the druid, to `ClearForm`) can cast a spell
that its form refuses but no form allows: it leaves the form first, once, after every other check has
passed. That holds in a stance-type form too: Healing Touch in Moonkin Form and Wrath in Tree of Life
leave the form.

- A class that reads the store by hand takes `row.CastRequirement()`.
- A hand-built spell with no row states it by hand as a `core.CastRequirement` literal:
  `Forms: dbcenums.FORM_BATTLE_STANCE.Mask()`, with `ExcludedForms`, `CasterForm` and the attribute
  bits beside it.
- A talent that swaps a spell on the action bar (`A_OVERRIDE_ACTIONBAR_SPELLS`) swaps its row too:
  Vanguard's Charge takes `chargeRank.OverriddenBy(spellData.Vanguard.Highest()).CastRequirement()`,
  whose row allows Defensive Stance.
- A form spell (subtext "Shapeshift") is a single-rank ladder like a warrior stance, so
  `spellData.CatForm` exists.

## The three shapes a registration takes

1. **The row, plus what no row states.** `sim/warrior/pummel.go` and `sim/warrior/slam.go`: the
   resolver, then the cast condition and `ApplyEffects`.
2. **The row, then an override of one of its fields.** `sim/warrior/hamstring.go` pins the threat
   numbers with their review marker, because the game files state no threat coefficient. Write the
   override after the call, on its own line, with the reason beside it.
3. **A hand-written `core.SpellConfig` reading the row's values.** The Sweeping Strikes hit spell in
   `sim/warrior/talents_arms.go` is a spell the sim registers and the store does not carry, so it is
   built by hand out of the parent's numbers.

## When not to take a default

- **A dot the game files keep on an `A_PERIODIC_DUMMY` stays hand-written.** `PeriodicEffect()` does not
  answer a dummy, and Deep Wounds' dummy states a spell power coefficient of 1 that `DotConfig` would
  put on a weapon-damage tick. Its spell config still resolves; only the dot is written by hand
  (`sim/warrior/talents_arms.go`).
- **A sub-spell whose casts the sim reports does not take `Proc()`.** The option marks the spell
  passive and empties its cast, cooldown and cost, and the metrics aggregator counts no cast for a
  passive spell. Blood Craze's heal takes `Proc()` on its own row, `BloodCrazeTriggered`
  (`sim/warrior/talents_fury.go`). Whirlwind's off-hand strike takes it on the Whirlwind row itself,
  with `Melee(ProcMaskMeleeOHSpecial)`, `Proc()` and `Tag(2)`. That is what lets a spell with a cast time
  and a cooldown of its own lend its row to a sub-spell that has neither, with only `ApplyEffects`
  written by hand (`sim/warrior/whirlwind.go`). Deep Wounds (`sim/warrior/talents_arms.go`)
  and Retaliation's counterattack (`sim/warrior/retaliation.go`) instead name the flags they need by
  hand: the bleed takes `SpellFlagNoOnCastComplete` with `SpellFlagIgnoreResists` and
  `SpellFlagProc`, and the counterattack takes `SpellFlagMeleeMetrics`.
- **A sim-only copy of a spell carries the parent's `ClassFlags`.** Sweeping Strikes' hit spell and its
  normalized-attack copy (`sim/warrior/talents_arms.go`) have no row of their own, and without the
  parent's family mask the talents that name Sweeping Strikes would not reach them.
