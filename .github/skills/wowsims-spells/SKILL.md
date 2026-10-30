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
| naming a spell, aura, item or enchant in a `.siml` file, the SIML name table, `make siml-names` | `wowsims-siml`                                                    |

## Names in SIML files

Registering a spell adds its name to the SIML name table, `sim/siml/names.json`, when you run `make siml-names` (or `make siml`); builds read the committed table and never regenerate it, so commit the table with the spell. The rotations under `ui/specs` name spells and auras by this table, which the generator writes from what the sim registers: each spell id and tag gets a name from its game name (Hunter's Mark is `hunters_mark`) and lists the specs that register it. `TestTableKeepsItsEntries` in `sim/siml/gen`, which `go test --tags=with_db ./sim/...` runs, logs the names the table lacks and fails only when a committed entry comes back changed. Deploy and release build from the committed table, so a spell left out of it has no name on the site, and prints as a raw ref, until the table is regenerated and committed.

The table is append-only. A name keeps its spell id, tag and rank whatever the sim registers later, and a regeneration changes only its spec lists (`registeredBy`, `icdOnlyBy`) and its `uncastable` bit. A spell registered under another id or tag therefore gets a new name, with a qualifier because the old entry keeps the plain one (`sim/siml/SPEC.md` §7.4 lists them). The old name stays pinned, the generator reports it (`stale: warrior.spell.eye_for_an_eye is pinned to spell 9799; the sim registers spell 25997`), and a rotation that still writes it fails the repository check, `siml fmt -check ui/specs`. A class spell whose plain name a shared entry already holds (a gear proc named after a class spell) gets a qualified name too, and the class's files keep meaning the shared spell by the plain one; the generator reports it on a `qualified:` line, and running it with the `-repin` that line names on the committed table (`git checkout sim/siml/names.json` first, when `make siml-names` has written the new entry) gives the class the plain name. `go run --tags=with_db ./sim/siml/cmd/simlnames -w -repin <class>.<namespace>.<name>` is the only way a name changes meaning: it names the entry again from what the sim registers now, prints the old and new meaning, lists the `ui/specs` files that use the name, and adds a `<name>: was …, now …` line to the repin record, `sim/siml/gen/repinned.txt`, which is committed with the table. Without `-w` it prints the table instead of writing it. The plain name of a spell goes to the one the builds register castable (`core.SpellFlagAPL`, what the site's rotation editor offers), and a cast of a spell no build lets a rotation cast is a `not-castable` warning in the parser. On a pull request CI runs `go run ./sim/siml/cmd/simlnames -base HEAD^1`, which fails on any name of the base branch's table that means something else now, unless the pull request adds the name's line to the record. A name the generator added for a spell that is registered no more, and that `HEAD`'s table does not hold, it drops again on its next run, so work in progress leaves no name behind once it is undone.

In a `.siml` file, the SIML language server (`wowsimcli siml lsp`, which `tools/vscode-apl` starts) shows a spell's row on a hover over its name, as `tools/spelldata -lsp` does in Go.

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
