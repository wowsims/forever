# Pitfalls

**Class flags are not an ability's identity.** Filtering Blood Craze's Bloodthirst listener by
Bloodthirst's `ClassFlags` fired it on Mortal Strike too and moved multi-target HPS from 0.16 to 15. Ask
the ladder: `spellData.Bloodthirst.Contains(spell)`.

**A hand-built `core.SpellConfig` carries no class flags unless it sets them.** `spelldata.SpellConfig`
fills `ClassFlags` from the row, but a config written out by hand needs `ClassFlags: rank.ClassFlags`.
Without them every modifier, trigger and charge the game files state by class flags skips the spell, and
nothing reports it: Revelation's paladin buff 1323419 gave Holy Light no crit and was never spent.

**The data contains ranks the game never grants.** Fireball 38692 and Frostbolt 38697 are rank 14
entries at level 70 whose `SkillLineAbility` rows and spell attributes are byte-for-byte
indistinguishable from the real rank 13s, so the generator cannot filter them. Reaching for
`Highest()` on those families silently casts a spell that does not exist. When that happened during the
mage port, it cost 1.2% DPS. Use `ByID`.

**A parsed row doubles the hand line that already states it.** When the paladin parses Shield
Specialization, the `BlockValueMultiplier` line in `sim/paladin/talents_protection.go` goes, since
`A_MOD_BLOCK_VALUE_PCT` is a row. Holy Power's crit mods in `talents_holy.go` go the same way. The
`agiApDep` the druid's Cat Form stub keeps commented out must not come back beside a parsed Cat Form
(Passive) 3025, which states the same Agility share. Last Stand 12976 is the opposite case: it stays
hand-read, because its 30 is a percentage and the parse reads `A_MOD_INCREASE_HEALTH` as flat
health.

**A parsed school hit talent shows on the sheet only when asked.** Shadow Focus, Holy Precision and
Divine Precision state their hit on class flags, so a plain `ParseStatic` raises the hit of their
spells and leaves the character sheet's school hit rows where they were. Pass
`spelldata.ShowAsSchoolHit` with the schools the tooltip names. Writing the hit into
`PseudoStats.SchoolBonusHitChance` instead drops the row's spell list, and Divine Precision would
lose Holy Strike's melee roll. See [Talents and auras](talents-and-auras.md).

**A rank can carry nothing in a role.** Lay on Hands rank 1 restores no mana, while ranks 2-4 do, so
`EnergizeEffect()` answers `NilEffect` for rank 1, and every accessor reads that as zero.

**Never delete a generated file before regenerating it.** The class package stops compiling, and then
`gen_db`, which imports the sim, cannot build either. Regenerate over the existing file, or run
`git checkout HEAD -- <path>` to get it back.

**A melee ability states its bonus as weapon damage, not school damage.** Sinister Strike's +98 is
`E_NORMALIZED_WEAPON_DMG`, and Holy Strike's flat part sits there too. `DamageEffect()` answers the
first weapon or school damage effect alike, so the read is the same either way. An
`E_WEAPON_PERCENT_DAMAGE` on the same spell scales the flat part as well as the weapon damage: Holy
Strike rank 1's 12 ± 1.5 at 25% shows in game as 2 to 4.

**Rage is stored on a 0-1000 bar.** Heroic Strike's cost reads 150 where the player sees 15.
`PowerCost` and `Cost()` divide it through. Mana, energy and focus need no conversion. An `Energize`
effect that restores rage is still in tenths, which `Tenths()` reads.

**A new game files table needs a settings line.** `SpellCastTimes` was missing from
`generator-settings.json` and cast times read zero until it was added and `make db` re-run. Adding a
table is one line; the extractor needs no code.

**`EffectN` counts positions, `Effect.Index` is the game files' number.** `EffectN(1)` is the first effect
the row carries, whatever index the game files gave it. On 46 rows the stated index is not the position
the effect sits at. Reading a store row with the game files' number in hand is off by one wherever the
two agree, and wrong in a different way where they do not.

**`Average` keeps a damage or heal amount's fraction and floors every other, in float32.** In game
the amount a hit or heal deals rounds up with a chance equal to its fraction, so it averages the exact
value; a stat, a resource gain, an absorb or a threat is the whole number the tooltip shows. An effect
whose `EffectBasePointsF` is fractional (about one in 55) keeps its fraction under the per-level gain
either way, and `Min`/`Max` stay whole. Reading `BaseValue()` and doing the arithmetic in float64
gives a different answer on those rows.

**A talent's curve can disagree with its own row.** A `Talent` rank takes its base points from the
curve, and on some talents the curve states a positive number where the row and the tooltip state a
reduction. The ladder keeps the curve's sign, which is what applies in game, as [The store](store.md)
describes; compare `Highest()` against `MustFind(id)` to find such a talent.

**A row with more than one effect has to name the one it means.** `Ladder.ValueAt` panics rather than
read the first of several, and `Effect(aura, misc)` panics when two effects match. Both follow the same
rule: a caller says which effect it means with `EffectAt(n)` or `EffectN(n)`.

**A hand-built `Effect` scales from its own `BaseLevel` and `MaxLevel`.** The generator stamps both onto
every effect so `AverageAt(level)` needs no lookup. An `Effect` literal in a test or a hand-written config
that leaves them zero scales `PPL` from level 0 up to the caster's level. Copy them from the spell's row,
or set `PPL` to 0.

**`MustFind` at package init fails at startup, not at the call site.** That is intended: a regeneration
that drops or renumbers an id stops the sim with the id in the message. It also means a package-level
`var` reaching for a spell this build does not carry takes the whole package down, including tests that
never touch that spell.

**`SPCoef` of exactly 1 is the column's filler.** 2,113 of the store's 9,651 effects carry it, among
them weapon-damage effects, speed auras and shapeshifts. No physical-school row in the game files states
a fractional coefficient at all. `Magic()` and `DotConfig` hand the row's coefficient to core as it
stands, so a physical row read through either would take spell power per hit or per tick. Read the
tooltip before you believe a coefficient of 1.

**`Spell.ProcChance` is not always a chance.** 100 and 101 are the game files' "fires whenever its own
condition is met", and the tooltip says whether the column means a chance or that condition.
`ProcChanceSource`, baked in at generation, holds the answer, and `spelldata.ProcTrigger` reads it
instead of the column.

**A proc row with no stated rate panics when the trigger is built.** `ProcChancePPM` with no override
behind it means the game files state nothing anywhere, so `spelldata.ProcTrigger` demands a `PPM()` and
does not build a listener that fires on every hit.
