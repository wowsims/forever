---
name: wowsims-abilities-auras
description: 'Use when WoWSims Forever Go code builds an ability, aura, dot, talent, passive, raid buff or debuff from its spell row: calling spelldata.SpellConfig, AuraConfig, DotConfig, ParseStatic or ParseEffects, choosing Melee, Magic, Proc, Tag, RankLabel, FromCaster or ShowAsSchoolHit, wiring what SPELLDATA_REPORT=1 says the parse skipped, porting a talent or a weapon-bound hit or crit talent, or registering a buff manifest row. Use it too when a stance or weapon check sits in ExtraCastCondition, a sub-second GCD clamps back to one second, a talent reads the wrong effect, a per-point literal disagrees with the ladder, two ranks or two casters share one aura, a debuff category keeps only one debuff, or the row and the tooltip disagree on one spell. Reading a rank or effect off the store is wowsims-spell-store; a proc or trigger is wowsims-procs; consumables and item effects are wowsims-consumables-items; hit tables, costs and cast order are wowsims-core-mechanics; the map of every spell skill is wowsims-spells.'
---

# wowsims-abilities-auras

The resolvers build an ability, an aura, a dot or a talent's modifiers from the row, and the caller adds only what no
row states: `ApplyEffects`, the proc mask, a condition the game files lack. The rule this skill exists for: never
restate the row in code. A stance check, a duration or a per-rank literal typed beside a resolver agrees with the row
until the data moves, and then it is wrong with nothing to say so. The rulings the code follows are in
`references/rulings.md`. Text here follows `wowsims-conventions`; the pages below hold the detail.

## Where things live

| Path                                          | What it is                                                                        |
| --------------------------------------------- | --------------------------------------------------------------------------------- |
| `sim/core/spelldata/resolve_spell.go`         | `SpellConfig` and its `SpellOpt`s                                                 |
| `sim/core/spelldata/resolve_aura.go`          | `AuraConfig`, `DotConfig` and their `AuraOpt`s, `PeriodicDamageTick`              |
| `sim/core/spelldata/parse_effects.go`         | `ParseStatic`, `ParseEffects`, `DryRun`, their `ParseOpt`s and the skipped report |
| `sim/core/spelldata/parse_effects_table.go`   | which aura and misc value becomes which sim kind                                  |
| `sim/core/spelldata/weapon_scope.go`          | `WeaponScope`, the rule a weapon-bound hit, crit or damage aura follows           |
| `sim/core/shapeshift.go`                      | `CastRequirement`, which core checks on every cast                                |
| `sim/core/dbcenums/spellmods.go`              | the `SPELLMOD_*` names                                                            |
| `sim/core/buffs/`                             | the generated buff and debuff constructors, `newDebuff`, `newCasterDebuff`        |
| `tools/database/buffmanifest/`                | one row per buff and debuff proto field                                           |
| `ui/sim/talents/trees/`                       | the talent trees, which give a talent's row and points                            |
| `tools/database/overrides/spell_overrides.go` | numbers the game files do not state, such as an attack power share                |

| You are…                                                                                | Open                                     |
| --------------------------------------------------------------------------------------- | ---------------------------------------- |
| registering an ability, picking its options, or asking where it may be cast             | `docs/spell_data/building-an-ability.md` |
| building an aura or a dot, a rank group, a tick's outcome                               | `docs/spell_data/auras-and-dots.md`      |
| parsing a talent, passive or buff, `FromCaster`, `ShowAsSchoolHit`, the skipped report  | `docs/spell_data/talents-and-auras.md`   |
| wiring a hit, crit or damage talent bound to a weapon                                   | `docs/spell_data/weapon-requirements.md` |
| adding a raid buff or debuff, a manifest row, a driver, totems on pets                  | `docs/spell_data/buffs-and-debuffs.md`   |
| chasing a value that is off: class flags, a parsed row doubling a hand line, `SPCoef` 1 | `docs/spell_data/pitfalls.md`            |
| deciding which number wins for one spell: a form rule, a tooltip bug                    | `references/rulings.md`                  |
| reviewing a spell built beside its row: hand literals, scaling, targets, requirements   | `references/traps.md`                    |

## An ability

The options, the fields the row fills and the three shapes a registration takes are in `building-an-ability.md`.

- **`CastRequirement` is never restated in `ExtraCastCondition`.** `SpellConfig` fills the form, stance, caster aura,
  weapon, shield, creature type and out-of-combat state, and core refuses the cast. `ExtraCastCondition` holds only
  what no row forbids; a refusal the row does not state is a defect too. An aura that lifts a requirement sets
  `CastRequirement.CasterForm` on gain and clears it on expiry. Cases:
  `.github/skills/wowsims-review/references/defects.md` §What the row forbids.
- **A sub-second GCD is named twice.** `Cast.GCDTime` floors a GCD at `core.GCDMin`, one second, unless the cast sets
  `GCDMin`, so a spell with a shorter GCD sets both `GCD` and `GCDMin`.
- **`Magic()` already sets the row's `DirectVariance`.** Code registered through it passes `Average()` or
  `DirectAmount` to the damage call, never `Roll(sim)`, or the damage spreads twice (PR #267).
- **A cast row plus a bolt row casts the bolt as its own spell.** On Penance, hit and crit talents name the bolt row's
  class flags and the cost talent names the cast row's.
- **Shadowform (form 28) is a stance in the game files** (`stanceForms` bit 27), and its tooltip bars only healing
  spells.
- Which row's Always Hit bit a cast reads is the first ruling under §Rolls and outcomes in `references/rulings.md`.

## An aura or dot

Labels, rank groups, the default tick and its outcome are in `auras-and-dots.md`; a dot kept on an `A_PERIODIC_DUMMY`
is in `building-an-ability.md` §When not to take a default.

- **An unranked per-target debuff takes `spelldata.Label(name + " - " + priest.Label)`**; without the caster in its
  label, two players of the class panic on registration.
- **`FromCaster` slots are numbered across the class.** Shadow Weaving holds slot 0 of 2 and the Discipline
  `AttachDDBC` slot 1 (`sim/priest/talents_shadow.go`, `sim/priest/talents_discipline.go`), so a new per-caster debuff
  raises the total in every call.

## A talent's modifiers

What the parse attaches, what it skips and its options are in `talents-and-auras.md`; a weapon-bound talent's rule
per aura is in `weapon-requirements.md`.

- **Pick a talent's effect by what the call site does, never by which number fits.** A talent can state the same
  1/2/3 ladder on two effects (a stat bonus and a crit reduction); the mod's `Kind` decides which one is read.
- **Read the ladder; never multiply rank 1.** Migrating a per-point literal buys provenance, rarely accuracy: most
  percent talents scale linearly, and the few that do not are why the rows are worth reading.
- **The `SPELLMOD_*` names are read off the talents that use them**, because the game files ship no list. A value no
  talent names means the data changed, not that a name is missing.
- **Talent rows and points come from `ui/sim/talents/trees/*.json`**, never the `Talent` table in
  `tools/database/wowsims.db`, which holds a different tree layout.
- **"By up to" damage taken is coefficient-scaled.** `A_MOD_DAMAGE_TAKEN` on a magic mask worded "by up to" is spell
  damage against the target, scaled by each spell's coefficient (`SchoolBonusSpellDamage`), but `ParseEffects` files
  it as flat `BonusSpellDamageTaken`; build such a set bonus or debuff by hand (PR #283).

What a class still wires by hand is what the parse skipped; print it from the class's tests:

```
SPELLDATA_REPORT=1 go test --tags=with_db -v -count=1 ./sim/<class>/ | grep '^spelldata: unparsed' | sort -u
```

## Buffs and debuffs

The manifest, the kinds, the apply order and the drivers are in `buffs-and-debuffs.md`.

- **`NewExclusiveEffect` sets `category.SingleAura` on every call**, so every registrant of a debuff category passes
  false or the last one turns the category back to one aura. A class or pet copy of a debuff uses the generated
  `buffs.NewSpellData*Aura` or `newCasterDebuff`, with the tag set before `RegisterAura` (PR #159).
- **Issue #36 is the checklist of every buff and debuff.** A change to `tools/database/buffmanifest/`,
  `sim/core/buffs/`, a class cast behind a listed buff or the game files re-audits it. A box is ticked once the
  raid-input aura works with no stub or TODO left; a class cast's own TODO (threat) does not hold it open.

## When the sim and the row disagree

The decision rule is in `wowsims-spells`, `## When a sim number disagrees with the game files`; apply it one spell at a
time and leave the outcome at the call site. Flurry is the case: talent 12319's tooltip names its own ladder as the
haste ("$m1% for your next $12966n swings"), so the sim reads that ladder over buff row 12966's flat 30, and the
comment in `sim/warrior/talents_fury.go` quotes the wording.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. The commands that
re-derive each claim, the "You changed → Update" table and the path check are in `references/maintenance.md`.

A ruling, trap or workflow fact learned while using this skill is written down in the same change that uses it, where
`AGENTS.md` says: a comment at the code when one file owns the rule, this skill or its `references/rulings.md` otherwise. A memory file is a pointer
to this skill, not a second copy.
