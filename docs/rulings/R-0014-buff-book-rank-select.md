# R-0014: A buff a class book raises is a rank select with 0 as off

- Skill: wowsims-abilities-auras
- Code: `tools/database/buffmanifest/buffs.go`, `sim/core/buffs/meta.go`, `sim/core/buffs/priest.go`, `ui/features/inputs/icon_inputs.ts`
- Pull request: #794
- Evidence: none

A raid or individual buff that a class book raises is a rank select: an int32 where 0 is off and 1 the trainer's top
rank, listed in the manifest row's `Ranks`, from which the generators render the proto field, the apply block and the
settings input. The totems and greater blessings take a select. A shaman's own totem reads the row of the rank it casts.
The four group inputs, `mark_of_the_wild`, `power_word_fortitude`, `arcane_intellect` and `shadow_protection`, are bools
that read the single-target spell's top rank, since its amounts match the group spell's top rank (Mark of the Wild +16
stats and 385 armor; Power Word: Fortitude 70; Arcane Intellect 31; Shadow Protection 60). A mage in the raid sets
`arcane_intellect`. A priest's Prayer of Fortitude and Prayer of Shadow Protection read their own rows and bid against
the raid's input of the same stat, so the two count once. Prayer of Spirit stays a bool, since no item teaches it.
