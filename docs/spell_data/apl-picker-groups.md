# Groups in the APL picker

The APL cast-spell and aura pickers file each entry under one submenu. They try, in order:

1. the top level, for a shift spell of the unit and the aura of the same name;
2. a manual group the sim names (`MenuGroup`);
3. its category (`Cooldowns`, `Consumables`, `Items`, `Racials`, `Professions`, `External Buffs`, ...),
   except `Talents`;
4. its talent tree (`Holy`, `Arms`, ...);
5. the `Talents` category, for a talent aura with no tree;
6. `Spell` (cast-spell picker) or the top level (aura picker).

The cast-spell picker lists the shift spells first, then the talent trees, the manual groups, `Spell`, the
categories and the placeholders. The aura picker lists the shift auras, the talent trees, the manual groups
and the categories, then the remaining auras at the top level. The shift entries follow the sim's form
order and the trees the talent calculator's. The ranks of a spell sit behind its highest rank wherever it
is filed.

## Shift spells

The metadata's `shift_spells` are the unit's spells with `SpellFlagAPL` that shift it into a form, one per
form, in form id order: Battle, Defensive, Berserker for a warrior, the order of the stance bar in game. A
shift spell the APL cannot cast is left out. A spell config built by `spelldata.SpellConfig` carries the
row's `ShapeshiftForm()` as `ShiftsTo`, and a hand-built `core.SpellConfig` sets `ShiftsTo` itself.

The pickers list those spells at the top level ahead of every other entry, in that order, whatever their
group or category. An entry matches a shift spell by its name without the rank and its tag, so every rank
of a ranked shift spell sits there behind its highest rank, and so does the aura of the same name.

## Talent trees

An entry named after one of the build's talents files under that talent's tree. The metadata lists the
talents the unit has points in (`UnitMetadata.talents`, each at its talent spell id and held rank). The
pickers match an entry's name without the rank and its tag against each held talent's name, then look
the talent's spell id up as `spellId` in `ui/sim/talents/trees/<class>.json`. The entry files under the
`name` of the tree that talent sits in, as the JSON writes it. A pet or a target holds no talents, so its
entries find no tree.

Only an entry with no category or with the `Talents` category is matched. A talent entry with another
category, such as a talent's major cooldown under `Cooldowns`, keeps that category. A spell with no
matching talent stays under `Spell`, and a talent aura with none stays under `Talents`. That covers an aura
whose name differs from its talent's, such as Twist of Light's Echo auras.

## Manual groups

Set `MenuGroup` on a `core.SpellConfig` (or the `*core.Spell` it returns) or on a `core.Aura` to a
lowercase_underscore key, and add the label under `rotation_tab.apl.submenus.<key>`. The group wins over
the category and the talent tree; a shift spell and its aura stay at the top level whatever their group. An
aura with no group of its own takes the group of the spell that casts it (its `RelatedSelfBuff`, or the
spell with its id).

`spelldata.SpellConfig` sets no group. Class code that wants one sets `MenuGroup` on the returned config,
and the spell's aura follows:

```go
config := spelldata.SpellConfig(&warrior.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
config.MenuGroup = "my_group"
```

A hand-built `core.SpellConfig` or `core.Aura` sets `MenuGroup` itself.
