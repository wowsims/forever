# Traps in the game files' rows

_Source of truth:_ the rows themselves, which `go run ./tools/spelldata <id>` prints, and `discoverLadders` in
`tools/database/gen_spell_data.go`, which decides which rows become a family's ranks.

## Ladders and ranks

- **A ladder is never keyed by name alone.** Forever re-issues a rank as a granted copy (`AcquireMethod` 3 beside the
  trained 0 or 2), one name can carry two parallel ladders (Holy Shock damage and heal, Renew), `SupercedesSpell` is 0
  almost everywhere, and `SkillLineAbility` rows come duplicated.
- **A rank exists in Forever when it has a `SkillLineAbility` row and the Wowhead Forever tooltip names it.** Battle
  Shout 25289 and Heroic Strike 25286, level-60 ranks with high ids, are trained rows.
- **A book rank's `SkillLineAbility` row looks like a trained one.** Shadow Bolt 25307, Immolate 25309 and Corruption
  25311 are taught by the level-60 Grimoire of Shadow Bolt X, Immolate VIII and Corruption VII items, not by the
  trainer. The item's `ItemEffect` row with `TriggerType` 6 names the rank it teaches, which tells a book rank from a
  trained one. A second level-60 rank, or a rank off the ladder's level step, is the hint to check.
- **Class lines hold rows that are not class spells.** Poisons is skill line 40 in CategoryID 9 with the rank numeral
  in the name, and talent passives sit in class lines as trained rows with `BaseLevel` 1; a real trained passive is
  taught past level 1.
- **The summons have ladders** because `discoverLadders` takes the "Summon" subtext, and the store carries
  `SpellReagents` (`Spell.Reagents`, `ReagentCount`). Pet stats and the Warlock Pet Scaling amounts are not in the
  game files and stay hand-kept (PR #331).

## Talents

- **A trait-tree talent is one spell plus a curve row** (`spelldata.Talent(id, maxRanks)`). Check rank values against
  the live trait tree's TraitDefinitionEffectPoints and CurvePoint: a talent can also sit in an inactive tree copy, and
  a triggered buff row's BasePoints is rank 1 only.
- **A talent's active spell carries `ClassMask` 0.** The only link from talent to ability is the name: a ranked spell
  under the talent's name is the ability, an unranked one is the talent's own aura or twin.

## Wowhead against the row

- **`Effect.Variance` is the spread the game rolls.** Wowhead's Forever pages print only the flat base, so
  `Effect.Roll` stays; the shape of each roll is in `docs/spell_data/store.md`.
- **Wowhead Forever's `Value:` is `EffectBasePoints + 1`.** Its tooltip numbers scale to `SpellLevels.MaxLevel`, not
  60, and its rolled ranges do not match the game, so a range comes from an in-game tooltip at a stated level (PR
  #161).
