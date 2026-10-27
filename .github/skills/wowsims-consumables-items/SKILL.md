---
name: wowsims-consumables-items
description: 'Use when WoWSims Forever work touches the item, enchant or consumable database or the effects registered for it: a db.json or db.bin entry; a potion, flask, elixir, food, scroll, imbue or explosive (routing, what ships, potion classes, exclusive groups, mob type and zone stats); a hand-written item, set, trinket or enchant effect and the folder it goes in; random suffixes, caster weapon stats or swap-set imbues. Use it too when an item works in tests but not on the dev server or in the browser, core.GetItemByID answers nil, the sim panics with "Cannot add multiple effects for one item", a regeneration ships consumables nobody chose, a potion sits in the wrong picker, or an enchant id answers the wrong row. Running gen_db, overrides and a database behind master are wowsims-db-pipeline; item procs are wowsims-procs; reading a row is wowsims-spell-store; a buff from a row is wowsims-abilities-auras; tests are wowsims-testing; the bulk sim request is wowsims-bulk-sim; the spell skills map is wowsims-spells.'
---

# wowsims-consumables-items

An item, enchant or consumable reaches the sim in two halves: an entry in the database `gen_db` writes, and an
effect registered for its id. The generator writes both wherever a builder in `sim/common/shared` can run the rows; a
hand-written effect covers the rest. The rule this skill exists for: what the sim reads must be in the database it
loads, and outside a `with_db` build that is only what the request carries. Rulings are in `references/rulings.md`.
Text here follows `wowsims-conventions`; `docs/consumables.md` holds the consumable pipeline in detail.

## Where things live

| Path                                                  | What it is                                                                                                                   |
| ----------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| `assets/database/db.json`, `assets/database/db.bin`   | one `proto.UIDatabase` written twice by `gen_db`; nothing derives one from the other                                         |
| `assets/database/leftover_db.json`, `leftover_db.bin` | the items and enchants the simmable filters leave out                                                                        |
| `assets/database/loader.go`                           | embeds `db.bin` and `leftover_db.bin`; only a `with_db` build loads them into the sim's tables (`sim/core/database_load.go`) |
| `ui/sim/proto/database.ts`                            | the UI fetches `db.json` (`READ_JSON`) and the leftovers when a gear set needs them                                          |
| `tools/database/gen_db/main.go`                       | `gen_db`, and the order of the steps below                                                                                   |
| `tools/database/gen_consumables.go`                   | consumable routing and what ships; groups, unparsed stats and potion classes sit beside it                                   |
| `tools/database/gen_effects.go`                       | item, enchant and set effect generation and the missing-effects list                                                         |
| `tools/database/overrides.go`                         | item, enchant and consumable overrides, allow and deny lists, `ItemSetClasses`                                               |
| `sim/core/consumes.go`, `sim/core/item_effects.go`    | the consumable registry and the item and enchant registries                                                                  |
| `sim/common/forever/`                                 | generated registrations (`*_auto_gen.go`) and hand-written shared effects (`items_*.go`, `enchants.go`, `consumables_*.go`)  |
| `sim/<class>/item_<kind>.go`                          | hand-written effects only one class can use                                                                                  |
| `ui/sim/constants/missing_effects_auto_gen.ts`        | items, enchants and set bonuses no code registers an effect for                                                              |
| `ui/features/settings/model/consumables_auto_gen.ts`  | one named export per shipped consumable, which presets name consumables through                                              |

| You are…                                                                       | Open                                                                                                                                                  |
| ------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| adding, removing or correcting a consumable, or one that ships unasked         | `docs/consumables.md` §How a consumable gets into db.json, §Adding a consumable the rows do not support, §Removing a consumable, §Correcting a number |
| deciding which consumables stack                                               | `docs/consumables.md` §Exclusive groups                                                                                                               |
| a consumable in the wrong picker or the wrong order                            | `docs/consumables.md` §What the settings offer, then "Conditions are data" below                                                                      |
| writing a consumable's effect by hand                                          | `docs/consumables.md` §Writing an effect by hand: the Natural flasks                                                                                  |
| registering an item, enchant or set proc                                       | `docs/spell_data/procs.md` §Item and enchant procs, then `wowsims-procs`                                                                              |
| running `gen_db`, adding an override, or working from a database behind master | `wowsims-db-pipeline` `## Regenerating`, `## Overrides`, `## A database behind master`                                                                |
| deciding elixir or holiday stacking, an item's folder, a caster weapon's stats | `references/rulings.md`                                                                                                                               |
| reading an item, enchant, suffix or set row from the game files                | `references/item-data.md`                                                                                                                             |
| testing an item or consumable effect                                           | `wowsims-testing` `## What a test asserts`                                                                                                            |

## The pipeline

`gen_db -gen=db` writes the database and the registrations in one pass, after the numbered `.sql` overrides; running
it is `wowsims-db-pipeline` `## Regenerating`. For items, enchants and consumables the steps run in this order:

1. items and enchants from the game files;
2. every consumable routed and shipped (`RouteConsumable`, `ShipsConsumable`), then `ConsumableOverrides` merged;
3. `AddUnparsedStats`, then `AddPotionClasses`, which reads the lists the first one adds;
4. `ItemOverrides`, then `ApplyItemSetClasses`, which replaces the allowlist of every item in a listed set, then
   `EnchantOverrides`;
5. the global filters (`KeepEnchant`, `KeepConsumable`), the consumable icons, then the leftovers split off;
6. the generators: item effects, enchant effects, consumable registrations, the consumables TS file, the
   missing-effects file, the random property points.

`db.json` and `db.bin` are written at the very end, so a run that stops partway leaves the old database beside new
registrations.

- **A new parse row ships consumables.** `Effect.ParsedAura()` turns true for every misc value of a newly parsed
  aura, and routing reads that as supported, so the next regeneration can add imbues and stat auras nobody chose.
  Diff the generated consumables and check each new one against its row.

## Conditions are data

What an on-use grants only against some creature types, or only in some zones, is a list on the consumable:
`mob_type_stats` and `zone_stats` in `proto/db.proto`. How each is filled and read is `docs/consumables.md` §Stats
db.json lists outside the parse and §What the settings offer, which holds the EP order.

- **A conditional stat never goes into `stats`.** The sim reads the rows, never the lists, so a list that is wrong
  moves the settings and no sim number.

## Hand-written effects

A consumable's hand-written effect, and when it replaces the generated one, is `docs/consumables.md` §How the effect
is registered. Items and enchants work differently:

- **An item or enchant takes one effect.** `core.NewItemEffect` and `core.NewEnchantEffect` panic on a second
  registration for one id. The generator skips an id that already has a hand-written effect (`implementedByHand` and
  `core.HasEnchantEffect` in `tools/database/gen_effects.go`), which works because hand-written effects register in
  `init()` and the generated ones from `common.RegisterAllEffects`. A new hand-written effect therefore lands with the
  regenerated `sim/common/forever/` files in one change; until the generator runs, the sim panics on the duplicate.
  The first `gen_db` run still panics, since its binary was compiled from the old generated file (`sim.RegisterAll`
  runs after `GenerateItemEffects` rewrites the files and before `db.json` is written), and a second run completes.
- **The folder follows who can use the item**, as `references/rulings.md` states. A class test binary registers only
  its own package, so a shared set left in a class folder never reaches the other classes' AllItems suites;
  `TestGearSetsHoldOnlyTheirClassItems` in `sim/item_class_allowlist_test.go` fails a preset wearing another class's
  locked item.
- **Three aura lists route the generated item auras** (`tools/database/gen_effects.go`). `itemAuraKinds` is shared
  by on-use and equip spells, so an aura type added there, a class-mask modifier say, flips every equip item that
  states it into the item aura shape; equip-only types go in `equipAuraKinds` and on-use-only ones in
  `onUseAuraKinds`. `refuseEquipBesideAnotherEffect` refuses an equip aura on an item that already registers an
  effect and lists it as missing.
- **Shapes the builders already cover.** Damage shields attach through `buffs.AttachDamageShield`; an on-use that
  lowers threat registers as `CooldownTypeUnknown` and is never used on its own; `shared.AbsorbBreakDamage` lists the
  absorbs that deal damage when broken; a flat block value on an equip spell (Glyph of Deflection) is a stat through
  `ParseStatEffect`.

## The request database

- **Only `with_db` builds load every item.** The dev server and the WASM know only the items in the request's
  `SimDatabase`, so `core.GetItemByID` answers nil there for an item nobody wears, while every test passes. Code that
  reads an unworn item (the party Darkspear's swing speed in `sim/core/buffs/drivers.go`) needs the UI to add it to
  the request: `PARTY_BUFF_ITEM_IDS` in `ui/sim/proto/buff_items.ts`. Check such code against a build without the
  tag. How consumables reach the request is `docs/consumables.md` §The browser path; the bulk sim's is `wowsims-bulk-sim`.
- **A `with_db` build skips an effect for an id the database lacks** (`core.MissingItemEffects`), so an effect
  written for an item the game build does not ship registers nothing and fails nothing.

## Traps

- **An imbue check reads the main set and the swap set.** The swap set's imbues are their own `ConsumesSpec` fields,
  so a class effect keyed to an imbue checks the hand's temporary enchant on both (`carriesImbue` in
  `sim/rogue/poisons.go`, `ItemSwap.GetUnequippedItemBySlot`).
- **Zero means none on item scalars.** `ScalingItemProperties.unrounded_dps` is 0 for an item with no weapon table
  row and is read into `core.Item.UnroundedDPS` as it stands; the rule for every proto field is
  `wowsims-core-mechanics` `## Protos`.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. The commands that
re-derive each claim, the "You changed → Update" table and the path check are in `references/maintenance.md`.

A ruling, trap or workflow fact learned while using this skill is written down in the same change that uses it, where
`AGENTS.md` says: a comment at the code when one file owns the rule, this skill or its `references/rulings.md` otherwise. A memory file is a pointer
to this skill, not a second copy.
