---
name: wowsims-spec-implementation
description: 'Use when picking up a WoWSims Forever "[Spec] Implement ..." tracking issue, porting a class or spec onto the spell data store and resolvers, or producing the talent and ability checklist for a spec that has no issue yet: deciding which game files content is actually live in Forever, working the checklist family by family, keeping a long-running class branch merged with master, and proving the work with tests that can actually see it. Use it whenever the task names a class or spec to implement or port, even without the word "issue". Single spell, proc or consumable fixes outside a port are wowsims-spells.'
argument-hint: 'Name the spec and the issue number to work from.'
---

# Implementing a spec for Forever

Text written here follows `wowsims-conventions`. Read the router `wowsims-spells`, then the leaf each step
names. This skill is the part specific to a whole class: what belongs on the checklist, how to work it, and
how to know it is done.

## Where the checklist comes from

The checklists are datamined, not hand-written. The extract lives outside this repository (it is
tooling the sim never reads), so ask for its location rather than guessing a path. It writes one JSON per
class with `abilities`, `families`, `talents` and a `meta` block naming the game build and trait tree.

A family is an ability and all its ranks, keyed by its lowest rank's spell id, and the checklist has one
box per family: Heroic Strike has nine ranks and one box.

## Deciding what is actually live

The Forever build carries Season of Discovery content that is not live, and it is the biggest source of
wrong checklist entries: the spells are correctly extracted from the game files and simply not part of
the game being simulated. Compare the class against the SoD build by spell id (in both, SoD only,
Forever only) and drop the SoD-only ones. Entries with no name and no description are the first to
check; on the warrior every one of them was SoD-only, together with a whole Runes tree of
`S03 - Tuning and Overrides Passive` entries.

`gen_spelldata` already leaves out the rune abilities an `Engrave <slot> - <rune>` spell grants and
nothing else teaches, and the `// Not generated:` comment at the head of a class's
`spell_data_auto_gen.go` names them (`docs/spell_data/regenerating.md`). The issue's extract does not
filter them, so the comparison is still yours.

An ability is SoD-only when an `Engrave` spell reaches it (effect 54, enchant, aura 332 or trigger) and no
trainer row or chain from one does; id range and skill line prove nothing, `AcquireMethod 0` alone is not a
trainer row, and a live ability has a real learn level or a trait entry. Wowhead Forever's sources and stats
on SoD- and Classic-range items are inherited, not observed in Forever, so a page proves nothing; a new
Forever id, `envChange.status == "new"` or a Forever-range effect spell shows an item is live, and tooltip
stats with no game files row stay unconfirmed.

## Traps in the checklist

- **A new low rank shifts every rank above it.** Forever's Slam gains a rank 1 (1240193), so every older
  Slam id sits one rank lower while keeping its id. Match on spell id, never on rank number.
- **Descriptions carry in-game colour codes.** Strip `|CFFFFFFFF` and `|R` and collapse whitespace, or
  the issue body renders as noise.
- **Name and id answer different questions.** A talent missing from another build's ids may exist there
  under another id; check the other build's data before calling something new, rather than inferring it
  from what the sim registers.
- **An id is absent only when every table says so.** Check it against ItemSparse, Item, SpellName and
  SpellItemEnchantment together, and look for a rename: Curse of Agony and Curse of Doom are Bane of Agony
  and Bane of Doom, Faerie Fire (Feral) folded into Faerie Fire, Mangle is one Bear-only spell.
- **Counts in an issue body go stale.** Treat them as a starting point, and say what was filtered and
  why, so a re-run with different numbers is explainable.

## Working the checklist

Start from `sim/<class>/spell_data_auto_gen.go` and take every number the game files state from it
instead of a literal; a literal that survives carries its reason on the line. Whether the class reads
rows by hand or through the resolvers decides the idiom (`wowsims-spell-store`
`## Hand reads and resolvers`); moving a class onto the resolvers follows `docs/spell_data/porting.md`,
whose key steps are dumping the rows before touching a call site and isolating every .results move by
reverting one change. A talent naming a weapon, and the one hit, crit and haste rating, are
`wowsims-abilities-auras` `## A talent's modifiers`, including the weapon-bound talents of the classes
still to port. A stub's form, and when content is removed instead, is the silent no-op bullet of
`wowsims-review` step 2.

Where the sim and the game files disagree, decide rather than assume
(`wowsims-spells` `## When a sim number disagrees with the game files`), and record the decision at the
call site. A preset rotation follows `wowsims-apl` `## Timing`, and its cast counts are judged as
`## Presets` there says before its DPS is taken.

## Keeping the branch merged

A port lives on one branch while master moves under it: store and database conflicts are
`wowsims-db-pipeline` `## A database behind master`, and proto field numbers on a merge are
`wowsims-core-mechanics` `## Protos`.

## Verifying

```
go test --tags=with_db ./sim/<class>/... -count=1
git status --porcelain -- '*.results'      # empty unless a number was meant to move
make spelldata-check                       # the generated files match the database (needs wowsims.db)
```

Whether a suite can see the change is `wowsims-testing` `## .results discipline`, and the rest of proving
the work (the gates, `make update-tests`, where a test goes) is that skill too. Where the suite cannot see a
change, dump each constructed row against the literal it replaced, field by field; that is the check that
finds the bugs.

**Run the `wowsims-review` pass over your own diff before committing**, ideally from a fresh agent: every
number and hand-built config against its row and tooltip, the rules the rows cannot show, and every
finding verified before it is fixed. The PR body carries its "checked, not findings" and "hand numbers
that stay" lists.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. The "You changed →
Update" table and the path check are in `references/maintenance.md`.

A ruling, trap or workflow fact learned while using this skill is written down in the same change that uses it, where
`AGENTS.md` says: a comment at the code when one file owns the rule, this skill otherwise. A memory file is a pointer
to this skill, not a second copy.
