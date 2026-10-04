---
name: wowsims-spec-implementation
description: 'Use when picking up a [Spec] tracking issue for Forever: turning the datamined talent and ability checklist into a working spec, generating the class spell data, deciding what game files content is actually live, and verifying work the expected results cannot see.'
argument-hint: 'Name the spec and the issue number to work from.'
---

# Implementing a spec for Forever

Comments, docs, commit messages and PR text follow `wowsims-conventions`: "game files", not "client", no history in comments, no "golden".

## Scope

- Executing a `[Spec] Implement ...` issue, one checkbox per talent and per ability family.
- Producing the checklist for a spec that does not have an issue yet.

Not in scope: the spell data pipeline itself, which `wowsims-spells` covers, and item or enchant data.

## Where the numbers come from

The checklists are datamined, not hand-written. The extract lives outside this repo - it is working
tooling, not something the sim reads - so ask for its location rather than assuming a path. It
produces one JSON per class holding `abilities`, `families`, `talents` and a `meta` block naming the
game build and trait tree.

A `family` is an ability and all its ranks, keyed by its lowest rank's spell id. A checklist is one
box per family, not per rank: Heroic Strike has nine ranks and one box.

## Deciding what is actually live

**The Forever build carries Season of Discovery content that is not live.** This is the single
biggest source of wrong entries, and it is not obvious from the extract alone - the spells are in the
game files database, correctly extracted, and simply not part of the game being simulated.

The comparison against the SoD build is what settles it, per class: spell ids in both builds, ids in
SoD only, and ids in Forever only. Exclude the SoD-only ones. For warrior that removed an entire
Runes tree of `S03 - Tuning and Overrides Passive` entries, Dual Wield Specialization ranks 2-5,
eight stray Bloodthirst rows and Intervene - 13 of 54 families.

Anything the extract lists with no name and no description is the shape to check first; on warrior
every one of them was SoD-only.

## Traps

- **A new low rank shifts every rank above it.** Forever's Slam gains a rank 1 (`1240193`), so every
  older Slam id is one rank lower than it was while keeping its id. Match on spell id, never on rank.
- **Descriptions carry in-game colour codes.** Strip `|CFFFFFFFF` and `|R` and collapse whitespace, or
  the issue body renders as noise.
- **Comparing by name and comparing by id answer different questions.** A talent absent from another
  build's spell ids may still exist there under a different id. If the claim is "this is new", check
  the other build's data directly rather than inferring it from what the sim happens to register.
- **Counts in an issue body go stale.** They are a starting point to verify against the game files. Say
  what was filtered and why, so a re-run producing different numbers is explainable rather than
  alarming.

## Implementing

Read the class's generated spell data first - `sim/<class>/spell_data_auto_gen.go` - and take every
number the game files state from it rather than writing a literal. `docs/spell_data/` has the shapes,
the readers and the worked examples; `wowsims-spells` has the pipeline.

Where the sim and the game files disagree, decide rather than assume, and record the decision at the call
site. A literal that survives needs the reason on its line.

### Hit, crit and haste, and the weapon they need

Gear and auras give one `stats.HitRating`, `stats.CritRating` and `stats.HasteRating`, each reaching
physical and spell alike; there is no melee- or spell-only rating to write. "Weapon requirements and the
one rating" in `docs/spell_data/weapon-requirements.md` has the rule for a talent naming a weapon, the APIs and the
weapon-bound talents already audited for the classes still to port: rogue Precision, Hack and Slash and
Fist Weapon Specialization, Lethal Shots, shaman Thundering Strikes and the shaman Dual Wield
Specialization rune.

## Verifying

```
go test --tags=with_db ./sim/...          # the tag is required
git status --porcelain -- '*.results'      # empty unless a number was meant to move
go run ./tools/database/gen_spelldata      # regenerate; the diff must be empty
```

**A .results file that cannot move proves nothing.** Check that the suite actually exercises what was
changed before reading a pass as evidence - grep the .results for the spell id. The TBC paladin .results
carry no seal spell id at all, so every seal change "passed" while testing nothing; that port was
verified instead by dumping each constructed row against the literal it replaced, field by field,
which is what found the bugs.

**Before committing, run the `wowsims-review` pass over the diff**, ideally from a fresh agent: every number
and hand-built config against its row and tooltip, the rules the rows cannot show, and every finding
verified before it is fixed. The PR body carries its "checked, not findings" and "hand numbers that stay"
lists.

Read the summary lines rather than the exit code: `rtk`-wrapped `go test` has reported success for a
failing run. `.results.tmp` files are written on every run, so their presence means nothing either.
