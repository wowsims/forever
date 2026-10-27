# Talents and auras: ParseStatic and ParseEffects

A talent, a passive and a buff are all the same thing to the game files: an aura whose `E_APPLY_AURA`
effects say what it does. The parse walks those effects and attaches each one to the sim: a `SpellMod`
for the two modifier auras, a stat buff, a stat multiplier, a share of one stat added to another, a
pseudo-stat multiplier or one of the speed multipliers. A port names the row it reads and leaves what
the row contains to the parse.

```go
spelldata.ParseStatic(&warrior.Character, spellData.Cruelty.Rank(warrior.Talents.Cruelty))
```

`ParseStatic` applies at once and never comes off, which is how a talent or a class passive behaves.
A bonus against a creature type and the share of armor the unit ignores live on the attack tables.
Those tables exist once the environment is finalized, so the static path sets both values from a
post-finalize effect.
`ParseEffects(character, aura, row, opts...)` makes the same attachments follow an aura: they turn on
when it is gained, off when it expires, and scale with its stacks where the row states
`CumulativeAura`.

The game files' own class flags limit a modifier effect. `EffectSpellClassMask` decides which
registered spells the mod reaches, so a talent touches exactly the spells the game files name it for.
An effect that names no spells at all, on a row of a class family, reaches that whole family; that is
what the game files mean by an unmasked class modifier.

**Call `ParseEffects` before the aura activates.** It attaches through `ApplyOnGain` and
`ApplyOnExpire`. The parse catches up an aura that is already active the same way core's `Attach`
helpers do, with one exception. The rows that need a `Simulation` to act (the cast, melee, ranged and
attack speed multipliers and the energy regen multiplier) stay off until the aura is applied again.

**An aura that several ranks share takes a modifier once**, however many of those ranks the modifier's
mask names, because the mod is attached to the aura and not to each spell pointing at it.

|                              |                                                                                                                     |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| `Effects(1, 2)`              | only these effects, counted from 1 by position the way `EffectN` counts                                             |
| `SkipEffects(3)`             | everything but these                                                                                                |
| `Conditional(fn)`            | a condition every attachment is gated on, read on gain and whenever the caller calls `Refresh(sim)`                 |
| `IgnoreStacks()`             | the values do not follow the aura's stacks, which is what a row stating charges rather than cumulative stacks means |
| `FromCaster(unit, i, total)` | the unit that cast a debuff, for the rows that reach only its attacks, in slot `i` of `total` on its attack table   |
| `ShowAsSchoolHit(schools)`   | the schools whose hit the character sheet shows the row's `SPELLMOD_RESIST_MISS_CHANCE` hit in                      |

**A debuff that reaches only its caster's attacks** sits on the target and is parsed onto the target's
aura with `FromCaster`: Shadow Weaving's shadow damage (`A_MOD_SCHOOL_MASK_DAMAGE_FROM_CASTER`, by the
school mask in its misc value), Hemorrhage's Rupture and Wrack's other shadow dots
(`A_MOD_SPELL_DAMAGE_FROM_CASTER`) and Winter's Chill's crit on Frostbolt and Ice Lance
(`A_MOD_CRIT_CHANCE_FOR_CASTER_WITH_ABILITIES`). The last two pick their spells by the effect's class
flags, as a modifier does. Each is a percentage applied once per stack, so Shadow Weaving at three
stacks is 6%. The parse holds one slot on the caster's attack table against the target. Its damage
effects multiply into one `DamageDoneByCasterExtraMultiplier` handler and its crit effects add into one
`CritChanceByCasterBonus` handler; core's `EnableDamageDoneByCaster` and `EnableCritChanceByCaster`
switch these on while an effect is up. The class numbers those slots as it numbers its own
`AttachDDBC` calls. Each parse owns its slot, so two debuffs that a class parses take two indexes.
Another caster's attacks on the same target are untouched. So is every spell of a raid debuff's
caster, which the sim does not model; without `FromCaster` the rows are skipped. The character
argument is nil because the aura belongs to the target, and the parse reads the row at the caster's
level.

```go
spelldata.ParseEffects(nil, swAura, spellData.ShadowWeavingTriggered.Highest(),
	spelldata.FromCaster(&priest.Unit, 0, 2))
```

Defensive Stance uses both shapes at once (`sim/warrior/stances.go`). The stance passive is parsed onto
the stance aura. Defiance is parsed onto the same aura with a `Conditional` for the shield: its row
states the shield, the parse does not scope to it, and `Refresh` re-reads the condition on an off-hand
swap.

The call answers a `*Parsed`. `Applied` names each attachment, or each part of one that has parts (the
stats of a multi-stat row, the melee-defense part beside a weapon damage done row). For each it gives
the effect, the sim kind it became and the value in the sim's own units. `Skipped` holds the aura
effects the table has no row for, which are exactly what the port still has to wire by hand.
`SPELLDATA_REPORT=1` prints each of those on the console with its position, the game files' name for
its aura, its misc value and its amount. Improved Slam (`sim/warrior/talents_arms.go`) is the ordinary
case: the cast time and global cooldown mods attach, and the five rank-swap effects behind them have no
sim kind and are reported.

## What a value becomes

The table writes the sim's own units, which differ from stat to stat:

|                                                                     | the sim stores                                        | the parse writes                                                       |
| ------------------------------------------------------------------- | ----------------------------------------------------- | ---------------------------------------------------------------------- |
| Strength, Agility, Stamina, Intellect, Spirit                       | flat points                                           | the value                                                              |
| Health, Armor, the five resistances                                 | flat points                                           | the value                                                              |
| PhysicalDamage, SpellDamage and the per-school damage stats         | flat power                                            | the value                                                              |
| HealingPower, AttackPower, RangedAttackPower                        | flat power                                            | the value                                                              |
| MP5                                                                 | mana per five seconds                                 | the value                                                              |
| PhysicalHitPercent, SpellHitPercent                                 | percentage points                                     | the value                                                              |
| PhysicalCritPercent, SpellCritPercent                               | percentage points                                     | the value                                                              |
| crit chance on the spells a modifier names, flat or percentage      | `SpellMod_BonusCrit_Percent`, percentage points       | the value                                                              |
| hit chance on the spells a modifier names                           | `SpellMod_BonusHit_Percent`, percentage points        | the value                                                              |
| BlockPercent, DodgePercent, ParryPercent, ExpertisePercent          | percentage points                                     | the value                                                              |
| HitRating, CritRating, HasteRating and the other ratings            | rating                                                | the value, for each rating the misc value's mask names                 |
| a rating bound to a weapon, or a mask of ranged bits alone          | a hit or crit percent mod on those attacks            | value / `HitRatingPerHitPercent` or `CritRatingPerCritPercent`         |
| `PseudoStats.BonusHealingTaken`                                     | flat healing                                          | the value                                                              |
| `PseudoStats.ReducedPhysicalHitTakenChance`                         | percentage points                                     | the value with its sign turned                                         |
| `PseudoStats.SpiritRegenRateCasting`                                | a share of spirit regen while casting, capped at 1    | the value over 100                                                     |
| `PseudoStats.BonusRangedDamageTaken`                                | flat damage on each physical ranged attack            | the value, where the school mask names physical                        |
| `PseudoStats.MeleeDamageTakenMultiplier`                            | a multiplier on each melee attack, of any school      | `1 + value/100`                                                        |
| ArmorPenetration, SpellPiercing                                     | flat points of the target's armor or resistance       | the value's size, whichever sign the row states it with                |
| `PseudoStats.SchoolSpellPiercing`                                   | flat points of one school's resistance                | the value's size, for each school the misc value's mask names          |
| AttackPower, RangedAttackPower, SpellDamage against a creature type | flat power on each attack table's `MobTypeBonusStats` | the value, for each mob type the misc value's creature type mask names |
| damage dealt against a creature type                                | `DamageDealtMultiplier` of each masked table          | `1 + value/100`                                                        |
| crit damage dealt against a creature type                           | `CritMultiplier` of each masked table                 | `1 + value/100`                                                        |
| the share of armor ignored                                          | `ArmorIgnoreFactor` of each attack table, a fraction  | the value over 100                                                     |
| a flat proc chance modifier                                         | percentage points on the trigger's chance             | the value                                                              |
| a percentage proc chance modifier                                   | a fraction: the chance times one plus it              | the value over 100                                                     |
| damage taken from the caster, by school or by class flags           | the caster's `DamageDoneByCasterExtraMultiplier` slot | `1 + stacks x value/100`                                               |
| crit chance for the caster, by class flags                          | the caster's `CritChanceByCasterBonus` slot           | stacks x value, in percentage points                                   |
| every pseudo-stat multiplier and stat dependency                    | a multiplier near 1                                   | `1 + value/100`                                                        |
| armor, (ranged) attack power, spell damage or healing from a stat   | a stat dependency's ratio                             | the value over 100                                                     |
| the rage bar's maximum (`A_MOD_MAX_POWER` misc 1)                   | rage                                                  | the value over 10                                                      |
| the energy bar's maximum (misc 3, and `A_MOD_INCREASE_ENERGY`)      | energy                                                | the value                                                              |
| the maximum mana (misc 0 of either)                                 | Mana, flat points                                     | the value                                                              |
| a percentage GCD modifier (`SPELLMOD_GLOBAL_COOLDOWN`)              | `spell.GCDMultiplier`, summed from 1                  | the value over 100                                                     |
| a percentage range modifier (`SPELLMOD_RANGE`)                      | `spell.RangeMultiplier`, summed from 1                | the value over 100                                                     |

Block, dodge and parry are percentage points like hit and crit: core divides them by 100 where it rolls
them (`GetBlockFromRating`), so the game files' 5 is stored as 5. `TestParseStaticStatConventions` pins
one row of each family against what core reads back. [Weapon requirements and the one rating](weapon-requirements.md)
covers how a rating mask is read and which hit and crit rows become mods on one weapon's attacks.

A spell school's hit talent states `SPELLMOD_RESIST_MISS_CHANCE` on the class flags of that school's
spells: Arcane Focus 11222, Elemental Precision 29438 (Fire and Frost), Shadow Focus 15260, Holy
Precision 1309957 and Divine Precision 1310904. Parse it, and the mod raises the hit of exactly the
spells the flags name, on whichever table they roll: Divine Precision reaches Holy Strike's melee roll.
Class flags carry no school, so the character sheet cannot place that hit in a school's row on its own.
For a talent whose spells are the class's main spells of a school, the class passes the schools its
tooltip names, `spelldata.ShowAsSchoolHit(core.SpellSchoolFire | core.SpellSchoolFrost)`. The mod then
carries them as `ShowAsSchoolHit`, and the sheet adds the hit to those schools' rows while the mod is
active: from the talents stage on for a `ParseStatic` in `ApplyTalents`, and while the aura is up for a
`ParseEffects`. It changes no spell the mod reaches, so the row is the hit of the talent's spells: a
spell of the school the flags miss, such as Holy Shield's damage under Divine Precision, keeps the base
spell hit, and the reforge optimizer's hit caps read the row as it stands. A row that attaches no hit
modifier panics with the option; a Nil row, an untaken rank, parses nothing. `PseudoStats.SchoolBonusHitChance`
is the bonus stats' school hit, which reaches every spell of the school with a `ClassSpellMask` or class
flags and no melee roll. A talent written into it by hand loses its row's spell list, and one written in
beside the parse counts its hit twice.

Armor has two flat auras on misc 1. `A_MOD_RESISTANCE` (22) is armor like any other, and
`A_MOD_ARMOR_NO_MODS` (674) is the one the class armor buffs and debuffs state: Mark of the Wild,
Devotion Aura, Sunder Armor, Expose Armor, Faerie Fire and Curse of Recklessness. Armor on 674 is the
`ArmorNoMods` stat, on players, pets and enemies alike. Every unit adds it to Armor through one
unmultiplied dependency (`AddUnmultipliedStatDependency`), which the percent armor modifiers leave
whole: they scale the rest of Armor and the dependency adds ArmorNoMods after. The equip scaling of
armor scales Armor alone, so it never reaches ArmorNoMods either. The row takes misc 1 alone and skips
any other mask with a reason, and `spelldata.ArmorNoModsStats` is the one rule the parser and the
item generator read it by. The item database writes 674 armor as ArmorNoMods for item equip spells,
item buffs, enchants and consumables, so the shared builders and the item swap carry it in the stats
like any other stat. Its EP is a weight of its own, which the stat weights measure whenever they
measure Armor.

Armor has two percentage auras. `A_MOD_BASE_RESISTANCE_PCT` scales the armor from items, which is what
its tooltip states, and `A_MOD_RESISTANCE_PCT` multiplies the whole stat but the 674 armor:
Inspiration's 8 is 8% more armor, whatever else it came from.

`A_MOD_MAX_HEALTH` is flat health, which is how the bear forms state it: Dire Bear Form's passive is
1240 health at 60, and `A_MOD_INCREASE_HEALTH` is flat health too. Last Stand 12976 states its 30% on
`A_MOD_INCREASE_HEALTH` and a script applies it as a percentage, so the warrior reads it by hand and
does not parse the row.

`A_MOD_ATTACK_POWER_OF_STAT_PERCENT` (598) and `A_MOD_RANGED_ATTACK_POWER_OF_STAT_PERCENT` (580) add a
share of the misc value's stat to attack power: Cat Form (Passive) 3025 states melee attack power as
100% of Agility, Careful Aim 1223984 both as a share of Intellect. `dbcenums` names them after those
rows, and 635 `A_IGNORE_FOG_OF_WAR` after Enchanted Flare 1221416, which has nothing to model.

Innervate 29166 states both mana regen auras. `A_MOD_MANA_REGEN_INTERRUPT` adds its share to
`SpiritRegenRateCasting`, and core caps the sum at the full spirit regen, so Innervate's 100% on top
of Reverence's 30% gives the regen the character has while not casting. `A_MOD_POWER_REGEN_PERCENT`
multiplies `SpiritRegenMultiplier` on the mana bar (Innervate's +400%) and the energy regen speed on an
energy bar. Innervate's regen comes from the generated aura. No flag forces full spirit regen, and its
driver keeps only the timing and the metrics. The energy row asks the unit for its bar at parse time and is
skipped on a unit without one, so a rogue enables the bar before it parses Adrenaline Rush 13750.

A maximum power row raises the bar its misc value names, through `SetMaximumRage` or
`SetMaximumEnergy`. A raised maximum grants nothing. When an aura carrying one drops, the power above
the old maximum is lost and counted against the aura's metrics. A reset starts the rage bar at
the starting rage the options asked for, clamped to the maximum as it stands then. Mana is the Mana
stat, which the sim keeps as the maximum, so misc 0 is a flat stat buff on any unit: the Flask of
Distilled Wisdom's 17627 is 2000 Mana. The generator's `ParseStatEffect` reads the same misc 0 as Mana
for the consumable and item stats it bakes. A unit without the rage or energy bar the row names skips
it.

Whether the gnome's Expansive Mind scales a raised maximum depends on where the maximum is raised. The
racial multiplies the maximum rage and energy by 1.05 once, in `applyRaceEffects`, which runs after the
class constructor and before `ApplyTalents`. So Boundless Rage is parsed in `NewWarrior` straight after
`EnableRageBar`, and the rogue reads Vigor by hand into the `MaxEnergy` it enables the bar with. Parsed
with the talents, either would miss the racial.

A percentage GCD modifier scales `DefaultCast.GCD` the moment it toggles, after the flat GCD mods, so
every reader of the default cast sees the same GCD. Haste applies at the cast, and the GCD minimum
still floors what is left. Nature's Grace's -10 takes a 1.5 second GCD to 1.35 seconds.

A bonus against a creature type bids in a per-stat category under its stat and its mask
(`AttackPowerVersus32` for attack power against Undead), so it competes only with the same stat against
the same creature types.

Two aura names read as something else. `A_MOD_CRIT_PERCENT_VERSUS` is crit damage against the creature
types, not crit chance: Humanoid Slaying 19153 states it next to the same 3% of damage dealt.
`A_SPELL_AURA_MOD_ARMOR_PENETRATION_PCT` is the share of the defender's armor ignored, not a rating:
Serrated Blades 14171 ignores 9% at three ranks.

Target resistance by mask 1 becomes `ArmorPenetration`. A mask of all five resistances becomes
`SpellPiercing`, and a mask of some spell schools becomes `SchoolSpellPiercing` for each of them. A mask
of armor and a spell school is skipped (`core.SpellSchool.PiercingStat`). The per-school kind bids under
its mask: `SchoolSpellPiercing4` for the Baleflame imbue's fire and `SchoolSpellPiercing16` for
Balefrost's frost. Core adds a spell's school piercing to `SpellPiercing`; a holy or multi-school spell
has no resistance stat to pierce.

A percentage `SPELLMOD_DAMAGE` raises only direct hits and direct heals (`SpellMod_DirectDamageDone_Pct`),
and `SPELLMOD_DOT` raises only the ticks (`SpellMod_DotDamageDone_Pct`). A talent that raises both a
spell and its dot therefore states the percentage twice, and each effect attaches its own mod (Improved
Serpent Sting 19464, Improved Moonfire 16821). `SPELLMOD_ALL_EFFECTS` raises both
(`SpellMod_DamageDone_Flat`).

Two percentage rows of `SPELLMOD_DAMAGE`, `SPELLMOD_DOT` or `SPELLMOD_COST` on one spell multiply in
game, each as its own factor: Improved Wrath and Moonglow on a Wrath costing 120 Mana leave 45, not 30.
The sim gives each of those rows its own factor: a hit's applies after the additive bucket that
`SPELLMOD_ALL_EFFECTS` sums into, and a cost's goes into `SpellMod_PowerCost_Pct`. A cost row that takes the
whole cost or more (the Clearcasting family) makes the spell free while it is up, whatever the other
cost rows are, and gives the cost back exactly when it drops.

## What the table does not take

Anything the table has no row for is skipped and reported; the parse does not guess a value for it. That covers the dummies the game files use
for "a script does this", the proc-driver auras (a proc is a `ProcTrigger`, not a modifier), the
crowd-control auras (stun, fear, root, silence) and `A_MOD_SKILL` on a profession skill; Defense reads
as defense rating and a weapon skill adds to the skill the player states. Those are named in the report
rather than numbered, so a port can see at a glance which ones are its own work.

The shape of the parse also skips some rows the table does know. A value the sim keeps as one number
per unit cannot be applied once per stack, and the static path has neither an aura to follow nor a
`Simulation` to answer a later `Refresh` with:

|                                     | skips                                                                                                                     |
| ----------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| a stacking aura (`MaxStack` > 0)    | the stat multipliers and shares, the equipment scaling, the pseudo-stat and speed multipliers and the cooldown multiplier |
| `ParseStatic`                       | the speed and energy regen multipliers, which need the `Simulation` an aura's gain hands over                             |
| `ParseStatic` with `Conditional`    | the stat multipliers and shares and the equipment scaling as well                                                         |
| an aura on a unit with no character | the equipment scaling and the stats and damage held while a weapon is equipped, whose helpers are a character's           |
| a parse without `FromCaster`        | the damage and crit a debuff adds to its caster's attacks alone, and `ParseStatic` those too                              |

`IgnoreStacks()` clears the first row of that table for a spell whose column counts charges rather
than cumulative stacks.
