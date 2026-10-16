# Porting a class onto the resolvers

1. **Start from the ladders.** Every class file is already one ladder per family. The druid,
   priest, rogue and warlock read them by hand, the way [Reading a row by hand](reading-by-hand.md)
   describes; grep the class folder for `spelldata.SpellConfig` and `spelldata.Parse` to see where a
   class stands. This checklist covers the move from hand reads onto the resolvers (`SpellConfig`,
   `AuraConfig`, `DotConfig`, `ParseEffects`, `ProcTrigger`), with the warrior's files as its examples.
2. **Dump the rows before touching a call site**, three ways: the effects at the rank taken, which
   registered spells each modifier's `EffectSpellClassMask` names, and the whole `ProcTrigger` the row
   decodes to. With the mask dump, you know a .results move before you run it.
3. **Pass `Ladder.Rank(n)`, never `MustFind(id)`,** for anything the talent tree prices: a trait
   talent's per-rank numbers live in the curve, and the base row's effect can read 0.
4. **Check the cooldown categories.** A category cooldown resolves to `Unit.CategoryTimer`, which is
   one map per unit and the same one consumables and racials use for categories 4, 30, 1141, 1153 and 1190. A class row stating one of those would share a cooldown with a potion. Two of the class's own
   abilities sharing a category share the timer, which is what happens in game: the warrior's Revenge
   and Overpower run off 65, Shield Bash and Pummel off 88, and Mortal Strike, Bloodthirst and Shield
   Slam off 971.
5. **Read what the resolver picked up.** `SpellFlagHelpful` follows the first effect's target and flips
   `IsFriendly` in the APL editor, so an attack whose first effect is a self side-effect needs it
   cleared. `Rank` comes from "Rank N" and flips `HasRanks` there. `IgnoreHaste` follows the physical
   school, not the defense type.
6. **A row with a `Variance` rolls.** Use `Roll(sim)` where the game files state a spread; the
   extra draw reshuffles every later roll, which is a .results move with a known cause.
7. **Parse before you activate.** `ParseEffects` attaches on gain, so an aura already up when the parse
   runs loses the rows that need a `Simulation`.
8. **Compare value by value before deleting a hand modifier.** A mod that lands in a different bucket
   (a school pseudo-stat where the sim had a per-spell mod) is the same number in a different place,
   and a DPS diff will not show it. Diff the `.results` as well as the DPS numbers.
9. **`RequireDamageDealt` defaults true.** A listener that fires on a dodge, a parry, a miss or a block
   needs it false, which the decoder has already done where the tooltip named the outcome and the row
   carries `ProcHintOutcomeTaken`; clear it by hand only where the wording yielded no hint. The outcome
   itself is always the caller's.
10. **Rename a trigger whose driver and buff share a name**, and split a row that decodes to hits dealt
    _and_ taken into two listeners with a condition each.
11. **Say which source won.** Where the row and the tooltip disagree, the decision goes in a comment
    that names the tooltip's wording. Where a number stays by hand, its review marker stays with it.
12. **Isolate every .results move** by reverting exactly one change and re-running. Two causes in one
    commit look like one inexplicable cause.
13. **Replace the class masks last.** A handler calling `spell.Matches` takes `spellData.X.Contains(spell)`;
    a `ProcTrigger` narrowed by `ClassSpellMask` takes the same call as its `ExtraCondition`, which runs
    before the proc roll and so draws no extra random number. A sim-only sub-spell with no row of its own
    is compared by its `*core.Spell` (Sweeping Strikes' copies), and Whirlwind's off-hand strike is
    `Whirlwind.Contains(spell)` with an off-hand proc mask. Never the row's `ClassFlags`: they share bits
    across abilities.
14. **A gear proc takes three calls:** `ProcTrigger` for the listener, `AuraConfig` for the buff
    it applies and `ParseEffects` for what that buff does. `sim/warrior/items.go` uses this idiom for
    every set bonus and item buff the class owns.
15. **`Melee` and `Magic` put the spell in the rotation; `Flags` does not.** A stance-gated
    ability takes `Flags` so the APL editor does not offer it. A config that writes
    `ThreatMultiplier = 1` or `FlatThreatBonus = 0` after the resolver repeats what the resolver
    already wrote; keep it only where its review marker says the number is still to be measured.
