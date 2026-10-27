---
name: wowsims-spells
description: 'The map of the WoWSims Forever spell data skills. Use it when work touches game files data or the generators that write it and the owning skill is not yet clear: a spell, talent, aura, dot or proc built from its row, a consumable or item effect, an override, a regeneration, a rotation, a test, or deciding whether a sim number or the game files are wrong. Use it even when the task only says "fix this potion", "this talent is off" or "regenerate the DB", then load the skill it names.'
argument-hint: 'Describe the spell, talent, proc, consumable, item effect or generator task.'
---

# WoWSims Forever game files data

Text written here follows `wowsims-conventions`. This skill is the map: find the task under "Which skill" and load
that skill. The detail lives in `docs/spell_data/` (start at its README), `docs/consumables.md` and
`tools/spelldata/README.md`, which the subject skills route to.

## Where things live

| Path                | What it is                                                          |
| ------------------- | ------------------------------------------------------------------- |
| `sim/common/shared` | The builders that run consumables and item effects from their rows. |

## Which skill

| Task                                                                                            | Skill                                                             |
| ----------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| reading a rank, effect or value off the store, `go run ./tools/spelldata`, hand reads           | `wowsims-spell-store`                                             |
| a number looks wrong at rank N, or a value equals another rank's                                | `wowsims-spell-store` (its traps), then `wowsims-abilities-auras` |
| building an ability, aura, dot, talent, raid buff or debuff from its row                        | `wowsims-abilities-auras`                                         |
| a proc trigger, proc chance, ICD, charged buff, item or enchant proc, a measured rate           | `wowsims-procs`                                                   |
| a consumable, an item, set or enchant effect, a `db.json` entry, an item the request must carry | `wowsims-consumables-items`                                       |
| regenerating the store or the database, overrides, a database behind master, hotfixes           | `wowsims-db-pipeline`                                             |
| what core does around a cast: order, costs, scaling, targets, hooks, auras, hit tables, protos  | `wowsims-core-mechanics`                                          |
| a rotation, an APL value or action, a preset                                                    | `wowsims-apl`                                                     |
| the gates, `.results` files, where a test goes, printing a value at runtime                     | `wowsims-testing`                                                 |
| judging a diff, your own before a commit or someone else's pull request                         | `wowsims-review`                                                  |
| porting a whole class or working a `[Spec]` issue                                               | `wowsims-spec-implementation`                                     |

## When a sim number disagrees with the game files

The rows are a second opinion on every number the sim already had. Decide rather than assume, and when a
disagreement stays, say so at the call site.

- The sim is likely wrong when its value equals a different rank's or a sibling spell's, or breaks the
  die-sides ladder of the other ranks.
- The sim is likely right when the offset is uniform across unrelated abilities: that usually means the
  sim models something the game files' base excludes (a ranged-weapon component in every hunter cast
  time), not a run of slips. A coefficient read from another spell is not a disagreement either: a tick's
  coefficient belongs to the spell its description names.
- Where the row and the tooltip disagree, the tooltip can win (Flurry's haste ladder against its buff
  row's flat 30), and the comment names the wording it came from. An in-game measurement beats both the row and
  the tooltip; a rank whose value jumps against its sibling ranks may be a tooltip bug, so ask the owner before
  changing the sim to the row.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. The "You changed →
Update" table and the path check are in `references/maintenance.md`.

A ruling or trap is written down in the same change: a comment at the code when one file owns the rule, its subject
skill otherwise.
