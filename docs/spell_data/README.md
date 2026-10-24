# Spell Data

The sim reads its numbers from the game files instead of hand-transcribed literals. They come out of
one store, `sim/core/spelldata`: a generated file that holds every spell the sim can reach, and
resolvers that turn a row into a spell config, an aura, a dot, a talent's modifiers or a proc listener.
The warrior, hunter, mage, paladin and shaman build on the resolvers; the druid, priest, rogue and
warlock read rows by hand, one value at a time, with no resolver in between. Grep the class folder for
`spelldata.SpellConfig` and `spelldata.Parse` to see which way the code you touch goes.

- [The store](store.md)
- [Building an ability](building-an-ability.md)
- [Groups in the APL picker](apl-picker-groups.md)
- [Auras and dots](auras-and-dots.md)
- [Talents and auras: ParseStatic and ParseEffects](talents-and-auras.md)
- [Weapon requirements and the one rating](weapon-requirements.md)
- [Procs](procs.md)
- [Overrides](overrides.md)
- [Reading a row by hand](reading-by-hand.md)
- [Regenerating and checking](regenerating.md)
    - [Taking a new game build one spell data group at a time](partial-updates.md)
- [Porting a class onto the resolvers](porting.md)
- [Pitfalls](pitfalls.md)
- [Buffs and debuffs](buffs-and-debuffs.md)

The accessor, spell config, aura and ladder snippets in these pages are pinned by
`sim/core/spelldata/example_test.go`, which runs them against the committed store. The snippets that
need a character (the parses and the proc triggers) are read off the warrior files they name.
