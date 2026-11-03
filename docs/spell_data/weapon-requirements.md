# Weapon requirements and the one rating

The game files name the weapon a spell needs in `SpellEquippedItems`: an item class (2 for a weapon, 4 for
armor and shields), a subclass mask and an inventory-type mask. The store carries them as `EquipClass`,
`EquipSubclass` and `EquipInvType`. The hand comes from `ATTR_EX_3_REQ_MAIN_HAND` or
`ATTR_EX_3_REQ_OFF_HAND`. On an ability, the requirement is what it takes to cast it. On a hit or crit
aura it also decides which attacks the bonus reaches, and there the same columns mean different things
per aura.

## One rating

Forever gear carries the game files' generic hit, crit and haste rating, and the sim keeps one of each:
`stats.HitRating`, `stats.CritRating` and `stats.HasteRating`. Each feeds both the physical and the spell
percent, and haste rating speeds melee, ranged and casting alike. The conversions are
`core.HitRatingPerHitPercent` (10), `core.CritRatingPerCritPercent` (14) and
`core.HasteRatingPerHastePercent` (10). No rating stat is melee-, ranged- or spell-only. A bonus that
reaches only one channel is a percent (`stats.PhysicalHitPercent`, `stats.SpellCritPercent`) or a mod on
some attacks.

`A_MOD_RATING` names its ratings as a mask of the game files' combat ratings, and `spelldata.RatingStats`
reads that mask bit by bit. The melee and spell bits of one rating name the same stat, and its ranged bit
adds nothing beside them. 8388608 is `ExpertiseRating`, and defense, dodge, parry and block have a bit
each. Two kinds of row go to `spelldata.ScopedRatingStats` instead: a row that names a weapon, and a row
whose mask holds only the ranged bits 64, 512 and 262144 (`spelldata.RangedBitsOnly`). For these the aura
parse turns the hit and crit among the ratings into the percent they are worth, as a mod on those attacks
tagged `FromRating`. Haste has no mod that reaches some attacks and not others, so the aura parse leaves
it unread, along with the other ratings. An item's equip spell whose mask holds only ranged bits is
folded in the generator instead: `dbc.RangedRatingPercents` turns its hit, crit and haste into the
ranged hit, crit and haste percents. An equip spell whose rating aura names a weapon is not folded; it
is applied at run time.

## The rule per aura

`spelldata.WeaponScope(row, effect)` answers the rule the parse applies to a hit, crit, rating or damage
done effect, and `go run ./tools/spelldata <id>` prints that rule in words on each effect line.

| aura                                                  | the row requires                                                  | the parse attaches                                                                                                                                        | rows                                                                                                                                                                               |
| ----------------------------------------------------- | ----------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 52 `A_MOD_WEAPON_CRIT_PERCENT`, 54 `A_MOD_HIT_CHANCE` | a weapon                                                          | per weapon: `SpellMod_BonusCrit_Percent` or `SpellMod_BonusHit_Percent` on the matching weapon's attacks                                                  | Cruelty 12320, Conviction 20117, Dual Wield Specialization 23584 (the off hand, by its hand bit), Fist Weapon Specialization 13707, Lethal Shots 450403, Ranged Hit Bonus +3 22780 |
| 290 `A_MOD_CRIT_PCT`                                  | a weapon not ranged alone, or one named by inventory type or hand | while equipped: physical and spell crit on every spell and attack while a matching weapon is in any weapon slot                                           | Thundering Strikes 16255                                                                                                                                                           |
| 290 `A_MOD_CRIT_PCT`                                  | bows, guns, crossbows, thrown or wands alone, by subclass         | per weapon: a crit mod on ranged attacks                                                                                                                  | Ranged Crit Bonus +02 1310308                                                                                                                                                      |
| 189 `A_MOD_RATING`                                    | a weapon, or no weapon and a mask of ranged bits alone            | per weapon, or on ranged attacks: the hit and crit it is worth, `FromRating`                                                                              |                                                                                                                                                                                    |
| 52, 54, 290, 189, 79                                  | a weapon, on an aura that reaches another unit                    | the stat, on that unit: the requirement only gates the cast                                                                                               | Scorpid Sting 3043                                                                                                                                                                 |
| 52, 54, 290                                           | a shield (class 4)                                                | the stat: a shield row is never scoped                                                                                                                    |                                                                                                                                                                                    |
| 55, 57 and the haste auras                            | anything                                                          | the stat: spell hit, spell crit and haste are never scoped                                                                                                | Precision 13705's spell hit                                                                                                                                                        |
| 79 `A_MOD_DAMAGE_PERCENT_DONE`                        | a weapon not ranged alone, or one named by inventory type or hand | while equipped: the school multiplier, and with a melee weapon a `SpellMod_DamageDone_Pct` on the melee-defense spells of the schools the mask leaves out | Two-Handed Weapon Specialization 20111 and 12163                                                                                                                                   |
| 79 `A_MOD_DAMAGE_PERCENT_DONE`                        | bows, guns, crossbows, thrown or wands alone, by subclass         | per weapon: a `SpellMod_DamageDone_Pct` of the mask's schools on ranged attacks                                                                           | Ranged Weapon Specialization 19507, Wand Specialization 6057                                                                                                                       |

A per-weapon mod carries no class flags, so it reaches the auto attacks whatever family the talent
belongs to.

A damage done row held while equipped follows weapon swaps, so the talent needs no `Conditional` of
its own. `Applied` names the melee-defense part as `damage-dealt-melee` beside the school multiplier.
That part reaches a spell whose `SchoolIndex` reads a school the row leaves out (`MultiplierSchool`),
so a melee-defense spell of several schools takes the bonus once.

**Read the row's tooltip before scoping; 290 is ambiguous.** "Improves your chance to get a critical
strike with all spells and attacks" is crit on everything while the weapon is equipped, which is what
the parse makes of a melee-weapon 290 (Thundering Strikes). "Increases your chance to get a critical
strike with Axes and Polearms" is per weapon, although the game files also state it as a 290, on the
Weaponmaster 12700 and Hack and Slash 13706 rows. The live Weaponmaster 1290261 and Hack and Slash 13960
carry that crit as an `A_DUMMY` branch, so each needs the crit built by hand as a per-weapon mod. Where
the parse and the tooltip disagree, the tooltip wins, and the call site names the tooltip's wording.

## Which API

| the effect                                                              | write                                                                                                          |
| ----------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| a per-weapon bonus the row states as 52, 54 or a ranged-only 290        | `ParseStatic` or `ParseEffects`; `Applied` names a `SpellMod_BonusHit_Percent` or `SpellMod_BonusCrit_Percent` |
| a per-weapon bonus the row does not state, or states as a 290 exception | `AddStaticMod(core.SpellModConfig{...})` with the Equip fields                                                 |
| a per-weapon bonus that shows as an aura                                | `AttachSpellMod` on that aura, with the Equip fields                                                           |
| stats on every spell and attack while a weapon is equipped              | `character.NewEquipGatedStats(req, bonus)`, which the parse's 290 path uses                                    |
| a temporary enchant's per-weapon effects                                | `spelldata.OnHand(...)`, which binds them to the hand carrying the enchant                                     |

Weaponmaster (`sim/warrior/talents_arms.go`) keeps its crit visible as an aura. `critOn` says whether
either hand holds an axe or a polearm when the character is built:

```go
critAura := warrior.RegisterAura(core.Aura{
	Label:      "Weaponmaster (Axe/Polearm)",
	ActionID:   actionID.WithTag(1),
	Duration:   core.NeverExpires,
	BuildPhase: core.Ternary(critOn, core.CharacterBuildPhaseTalents, core.CharacterBuildPhaseNone),
}).AttachSpellMod(core.SpellModConfig{
	Kind:          core.SpellMod_BonusCrit_Percent,
	FloatValue:    spellData.Weaponmaster.EffectAt(1).ValueAt(rank),
	ProcMask:      core.ProcMaskMelee,
	EquipSubclass: dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_POLEARM,
})
```

Stats held while a weapon is equipped (the shape the parse's 290 path builds, in
`sim/core/spelldata/parse_effects_table.go`) follow the weapon slots through item swaps.
`SetLevel(nil, 1)` adds them as the character is built, and a level of 0 takes them off:

```go
req := core.EquipRequirement{SubclassMask: core.WeaponSubclassMask(proto.WeaponType_WeaponTypeMace)}
character.NewEquipGatedStats(req, stats.Stats{stats.PhysicalCritPercent: 1, stats.SpellCritPercent: 1}).SetLevel(nil, 1)
```

The Equip fields are the game files' own masks. `EquipSubclass` takes `dbcenums.ITEM_SUBCLASS_BIT_WEAPON_*`
bits, written out in a `SpellModConfig` literal so `-config` and the hover can read them. In running
code, `core.WeaponSubclassMask(types...)` builds the same mask from `proto.WeaponType`s, one- and
two-handed alike. `EquipInvType` takes `dbcenums.INVTYPE_BIT_*` bits, and `EquipHand` takes the
`ProcMaskMeleeMH` or `ProcMaskMeleeOH` that a hand bit names.
`core.EquipRequirement{SubclassMask, InvTypeMask, Hand}` holds the same, and `Equipped(character)` answers
whether a weapon meeting it sits in any weapon slot. An item's own bits are `(*Item).WeaponSubclassBit()`
and the slot's inventory-type bit. A one-hander also counts as the main- or off-hand weapon of the hand
holding it.

## What core does with the Equip fields

- `shouldApply` checks the weapon behind the spell's proc mask: the main hand's for a main-hand bit, else
  the off hand's, else the ranged slot's. A spell with no weapon bit never matches.
- A magic-defense spell never matches, because it rolls on the spell tables, which no weapon's bonus
  reaches. Thunder Clap takes no Cruelty.
- A pet never matches, and neither does a shield or a held item in the off hand.
- An item swap re-filters every weapon-bound mod.
- The character sheet reads the mods through `ScopedPercents(probe)` for its main-hand, off-hand and ranged
  hit and crit rows, and shows the part tagged `FromRating` (`HitFromRating`, `CritFromRating`) as rating. A
  mod counts there only when it names no `ClassMask`, `ClassFlags`, school or spell flag.

`-config` reads a `core.SpellModConfig` literal the same way and reports the weapons its Equip fields
name and the rule they get. A hover on `SpellModConfig` in a class file shows the same.

## The weapon-bound talents of the classes still to port

| talent                                  | what it is                                                                                                                                                                    |
| --------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| rogue Precision 13705                   | effect 1, a 54 with melee weapons, is per weapon; effect 2, a 55, is spell hit on every spell. `ParseStatic` does both                                                        |
| rogue Hack and Slash 13960              | the dagger and fist crit is an `A_DUMMY`: a per-weapon crit mod by hand, `EquipSubclass: dbcenums.ITEM_SUBCLASS_BIT_WEAPON_DAGGERS \| dbcenums.ITEM_SUBCLASS_BIT_WEAPON_FIST` |
| rogue Fist Weapon Specialization 13707  | a 52 with fist weapons, per weapon. `ParseStatic`                                                                                                                             |
| Lethal Shots 450403                     | a 52 with ranged weapons, per weapon on ranged attacks. `ParseStatic`                                                                                                         |
| shaman Thundering Strikes 16255         | a 290 with melee weapons, while equipped. `ParseStatic`                                                                                                                       |
| shaman Dual Wield Specialization 408496 | the rune's hit is an `A_DUMMY` and the row names no weapon: spell and melee hit gated on dual wielding, `NewEquipGatedStats` with `Hand: core.ProcMaskMeleeOH`                |

A requirement of `Hand: core.ProcMaskMeleeOH` alone is met while the off hand holds a weapon, that is,
while the character dual wields.
