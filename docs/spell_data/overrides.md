# Overrides

`tools/database/overrides/spell_overrides.go` holds the numbers the store carries that the game files
do not state. One row per number:

```go
{16928, PPM, 1, "Annihilator: ProcChance 101 sentinel; TBC 1 PPM, unverified on Forever", "tbc-carryover"},
```

The next reader weighs the number against `Reason` and `Source`, and the generator refuses a row without
a reason. `Source` names where the number came from: `tbc-carryover`, `tooltip`, `wcl:<report>/<fight>`,
`issue #N`, `combat log, <spell and rank>, <date>`. A `wcl:`, `issue #` or `combat log` source is a
measurement on Forever; a carry-over or an `estimate` is listed as unvalidated on the gear it reaches.

Every field names a rule the generator checks before it writes the value, so a stale override lasts
until the next regeneration at most:

|                                   |                                                                                | stale when                               |
| --------------------------------- | ------------------------------------------------------------------------------ | ---------------------------------------- |
| `PPM`                             | procs per minute, onto `Spell.RPPM`                                            | the row states a `SpellProcsPerMinuteID` |
| `FlatThreat`                      | threat the tooltip states without a number                                     | the spell gains a threat effect          |
| `ThreatPerSpellLevel`             | flat threat of the value times the rank's spell level, onto `Spell.FlatThreat` | the spell gains a threat effect          |
| `APCoefDirect` / `APCoefPeriodic` | an attack power coefficient the game files keep in script                      | the game files state one on that effect  |
| `APCoefThreat`                    | the attack power share of the threat effect                                    | the game files state one on it           |
| `SPCoefDirect`                    | a spell power coefficient the direct effect does not state                     | the game files state one on that effect  |
| `ProcChancePct`                   | a whole percentage onto `Spell.ProcChance`                                     | the tooltip states a chance of its own   |
| `DurationMs`                      | an aura duration                                                               | the game files state a `SpellDuration`   |
| `ICDMs`                           | the internal cooldown between two procs, onto `Spell.ICDMs`                    | the row states a `ProcCategoryRecovery`  |
| `Hint`                            | the `ProcHint` bits the tooltip's wording did not yield                        | -                                        |

The generator also refuses two cases that have nothing to do with the field: an override naming a spell
the store does not carry, and two overrides of the same field on one spell. Every applied override
leaves an `// override: <field> <value> -- <reason>` comment on the row it wrote to, so the generated
store shows which numbers are not the game files'.

An item or enchant proc refused as `states no rate` loses that refusal once a `PPM` override sits on
the spell it is routed through (the `TriggerSpellID` of its commented-out registration, the id on its
`// trigger N` line). Where that was its only refusal, the proc registers. The spell it is routed
through depends on the kind of proc:

- A combat enchant (Effect 1) or a chance-on-hit item effect: the combat spell itself (Fiery Weapon's
  13897, Unholy Weapon's 20006), not the spell that grants the enchant. The rate is measured on the
  hits of the weapon carrying it (`NewDynamicLegacyProcForEnchant(id, ppm, 0)`, or `...ForWeapon` for an
  item).
- An equip aura (Effect 3): the aura carrying the proc trigger (Revelation's 1248806), not the spell it
  triggers (1248808) nor the grant (1248805). The rate is measured on the aura's own proc mask, so an
  aura whose flags decode to no mask stays refused as `no proc mask to measure its rate on`.

A procs-per-minute enchant proc rolls on weapon hits only. Its mask keeps its melee and ranged bits,
and a spell or heal hit never procs it. `dpmForMask` strips the other bits for every enchant.
`spelldata.EnchantAuraUnsupported` refuses a rate on an equip aura that hears only spells, as
`a procs-per-minute rate hears no weapon hits in this mask`, so the aura stays listed and does not
register a proc that would never fire. An enchant the game does let spells proc is an exception to
state there, by enchant; there is none today. A stated chance (Fiery Blaze 36's 15%, Insight 8216's
35%) is not a rate, and it hears what its row says.

Revelation 8217 stays listed. Its rate is scripted, and the game files do not state it. Trigger
1248806's `ProcChance` 100 is the sentinel beside the tooltip's "a chance", and its effect entry
resolves no stats from 1248808.

A rate follows the weapon the enchant sits on. A weapon enchant, ranged ones included
(`NewDynamicLegacyProcForEnchantWithMask`), rolls only on the hand carrying it, at that weapon's speed,
and an item swap moves it with the weapon. An enchant on something other than a weapon (armor, a cloak,
a shield or a held-in-off-hand item) is priced off the main hand (`NewLegacyPPMManager`, which prices
every mask but the off hand's and the ranged one's at the main hand's speed).

Where a combat spell and an aura apply the same spell (Crusader), the slot whose row states a rate is
the one kept, so putting the override on the combat spell keeps the combat spell. After adding a row,
run `gen_spelldata` and then `gen_db`, which classifies the procs out of the store compiled into it.
`TestEnchantProcRoutingTakesAPPMOverride` in `tools/database` and `TestSpellDataProcTakesAPPMOverride`
in `sim/common/shared` pin both halves on those three rows with a rate that exists only in the test.
`TestEnchantAuraPPMFollowsItsWeapon` beside the latter pins the routing by slot, and
`TestEnchantPPMFollowsTheEnchantedWeapon` in `sim/core` pins the move across an item swap.

`overrides.AreaBonuses` is the second table in the same file, for an effect whose tooltip says it is
doubled in some kind of area while the game files state no companion row for it ("This effect is doubled
in Volcanic areas" on Molten Fury). A row names the AreaGroup ids (any one of them counts), the factor
on the effect's amounts and the factor on its duration. These go onto `Spell.AreaBonusGroups`,
`AreaMultiplier` and `AreaDurationMultiplier`. `row.AreaBonus(&character.Env.Encounter)` reads them
against the encounter's area set. The override is stale once the row states a `RequiredAreasID` of its
own. The on-use stat actives that `shared.NewSimpleStatActive` registers read it; a proc chance the
tooltip says is doubled has no reader yet.

A consumable's spells take these overrides like any other spell's; the lists that add, remove or
replace a consumable itself are in [Consumables](../consumables.md).

`tools/database/overrides/extra_spells.go` is another hand-kept list: spells the generator
force-includes although no class, item, enchant or set reaches them. It is empty today. Each entry
states a reason and a source like an override, and the generator refuses one without a reason. The
extras are added while the store renders, so adding one without regenerating fails the check rather
than shipping a store that does not match its source.

`overrides.HandTriggers` (`tools/database/overrides/hand_triggers.go`) links a spell to one a
server-side handler casts off it, which no effect, `$<id>` token or skill line names. The store adds
each link to `Triggered()` and `Drivers()` beside the edges the effects state, and the closure follows
it, so the linked spell is carried. Caustic Smog Potion's cloud 1252526 creates an area trigger and
nothing else, and the link to 1252543 is what lets `shared.DamageCloudOf` find the damage it puts on
the enemies inside. An entry states a reason and a source, and the generator refuses one without a
reason. It also refuses an entry the game files have caught up with: once an effect of the spell
states the linked spell as its `TriggerID`, the regeneration fails with `the hand trigger ... is
stale` until the entry is removed.
