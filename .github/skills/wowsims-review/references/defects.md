# Defects a spell review looks for

Each section is one defect class from `SKILL.md`, the cases reviews of class ports found, and the fix that
reads the row. The cases name the row and the numbers, so each one can be checked again with
`go run ./tools/spelldata <id>`.

## A number the row states, typed by hand

- **Vigor 14983:** `spellData.Vigor.ValueAt(rank)` already answers 5 and 10. A second `5.0 * rank` on top
  made the maximum energy 120 at rank 2. The fix is `100.0 + spellData.Vigor.ValueAt(rank)`.
- **Expose Armor 11198:** the 450 armor per combo point is effect 1's `PointsPerResource` (-450).
- **Quietus 1310728:** two `A_DUMMY` effects still carry the numbers, 2/4/6/8/10 and 35. An `A_DUMMY` row
  is unparsed, not empty: read `EffectAt(n).FractionAt(rank)` and `EffectN(n).BaseValue()` off it.
- **Murder 14158:** `A_MOD_DAMAGE_DONE_VERSUS` misc 80 is Giant | Humanoid, not Dragonkin, and it is a
  parse row: `spelldata.ParseStatic(&rogue.Character, spellData.Murder.Rank(rank))`.

## Hand code the resolvers already build

- **Blade Flurry 13877:** 15 s, 2 min, 25 energy and `A_MOD_MELEE_HASTE_3` 20 were all typed by hand.
  `spelldata.SpellConfig` and `AuraConfig` plus `ParseEffects` build the same spell and restore the class
  flags the hand config dropped.
- **Slice and Dice:** the haste came from `EffectAt(2).FractionAt(2)`, a literal rank; `ParseEffects` on the
  rank already bound attaches it.
- **Feral `comboPointCost`, `refundOnMiss`, Enrage's ticks:** a loop over `rank.Powers` where
  `rank.Power(dbcenums.POWER_COMBO_POINTS)` answers; a second `RefundsOnMiss` check where `SpellConfig`
  already set the refund from `MissRefund()`; a hand `PeriodicAction` where `spelldata.DotConfig(rank, periodic)`
  works out the ticks (`sim/common/shared/spell_data_energize.go` does it for an `A_PERIODIC_ENERGIZE` row).
- **Rake, Pounce and Lacerate `OnTick`:** with an attack power coefficient of 0, each equals
  `DotConfig`'s default tick and can go.

## A placeholder that is not an override

- **Feral attack power coefficients:** five Go constants, with a PR body saying the generator refuses them
  once the game files state them. Only an override does that. Rip 9896's `A_PERIODIC_DAMAGE` takes an
  `APCoefPeriodic` override and Ferocious Bite 22829's `E_SCHOOL_DAMAGE` an `APCoefDirect` one, read with
  `APCoeff()`; the multiply by combo points stays in code.

## The wrong rank, the wrong effect

- **`Highest()` is the ladder's last entry, not its top rank.** `spellData.MutilateTriggered` lists
  399960 ... 1310706 in id order, and 1310706 is Rank 1 with 23 damage where the rank-4 hands carry 67. Each
  hand's row is the cast row's own trigger: `mutilateCastRank.EffectN(2).Trigger()` (main hand, 1241586) and
  `EffectN(3).Trigger()` (off hand, 1241590).
- **Hack and Slash:** the proc chance is effect 3, not effect 1 (the Dagger and Fist crit), and the ICD is
  the row's 200 ms, not 500: `EffectAt(3).FractionAt(rank)` and `Rank(rank).ICD()`. The two effects share a
  1..5 ladder, so nothing moved; the code still named the wrong one.
- **Furor 17056:** the tooltip prices the energy with `$m2`, effect 2; the code read effect 1, the rage
  chance, whose curve happens to match.
- **Heart of the Wild 17003:** the tooltip's `$s1` is Intellect and `$s3` cat Strength, and the code read
  each from the other's effect; both hold 2/4/6/8/10, so nothing moved. An `A_MOD_TOTAL_STAT_PERCENTAGE`
  effect names its stats in the second misc value where it sets one, a bitmask (8 Intellect, 4 Stamina, 1
  Strength), and the first value then does not count: 17003 states 0 and 8. `Effect.TotalStatPercentStats()`
  reads it, as the parse and `go run ./tools/spelldata` do; a hand read of `Misc` gets the wrong stat.
- **Initiative 13976:** 33/67/100% from effect 1 and a hit listener. A hand trigger with no chance fired at
  every rank, and on `CallbackOnCastComplete`, which has no result, so a dodged Garrote still awarded the
  point. `spelldata.ProcTrigger` on the rank decodes both.

## Core's order

- **Clearcasting and the attack that procs it: checked, not a finding.** `cast.go` runs `applyEffects`,
  then `Unit.OnCastComplete`, which read as though a Claw that procs Omen of Clarity would consume the
  Clearcasting it just gained. It does not: the Omen trigger comes from `spelldata.ProcTrigger`, which leaves
  `TriggerImmediately` false, so the handler runs one `SpellBatchWindow` later (`aura_helpers.go`), after
  `OnCastComplete`. Over 2000 stepped Claws, all 60 Clearcasting gains survived the Claw that procced them.
  The same order is a real defect for a hand-built trigger that sets `TriggerImmediately`. This was posted
  as a finding from reading the code and withdrawn once a runtime probe contradicted it.
- **`CallbackOnCastComplete` has no result** (`aura_helpers.go`), so `Outcome: OutcomeLanded` on such a
  trigger is never consulted.

## What a cast spends, and when

- **Ferocious Bite** must not use one success condition for both resources. The beta report in
  [PR #149](https://github.com/wowsims/forever/pull/149#discussion_r4178336479) says a failed Bite leaves
  zero energy. Spend extra energy on failure too, without refunding the base cost, while retaining
  combo points. The runtime regression checks a dodged 5-point Bite at 100 energy leaves 5 points and
  zero energy; that test verifies the implementation of the report, not the server behavior itself.
- **Cutthroat 462707** carries one charge, and nothing consumed it, so one proc allowed every Ambush the
  energy paid for in its 10 s.

## What scales

- **A flat weapon bonus is added before the weapon percent.** Ravage 9867 states +98 and 350%, and its
  tooltip reads "350% damage plus 343", which is 98 x 3.5. Shred, Ravage and Claw added the flat bonus
  after the multiplier and lost 44, 245 and 11.5 per hit.
- **Lacerate** ticked for the row's 15 at every stack count, and priced its direct hit on the count after
  the new application, where the tooltip says "per existing application".
- **Rupture 11275:** effect 1's `PointsPerResource` 4.73 was dropped, so a 5-point Rupture ticked for 35
  instead of 58.65 before attack power.
- **Furor:** the energy counted every second out of Cat Form, bear time included, and capped at maximum
  energy instead of the talent's cap; with Furor untaken the cat kept whatever the energy bar held, since the
  bar ticks in every form.

## How many targets

- **Demoralizing Roar 9898:** effect 1 targets `TARGET_UNIT_SRC_AREA_ENEMY` (`EffectN(1).HitsAnArea()` is
  true), and only `CurrentTarget` got the debuff.
- **Berserk 417141:** "strike up to $s3 targets" is 3 in all; the code hit 1 + 3.

## What the row forbids

- **Prowl 9913:** attribute word 0 has `0x10000000`, not usable in combat. It was recast on its 10 s
  category cooldown, and the .results showed 18 Prowls, 9 Pounces and 9 Ravages in 180 s.
- **Stealth's form was set and never cleared,** so the stance requirement on Ambush and Garrote was always
  met or never. Tie `ShapeshiftForm` to the aura's gain and expiry.
- **Faerie Fire 9907:** an `ExtraCastCondition` refused a recast while the druid's own debuff was up, which
  the row does not forbid, so the rotation could never refresh it early.
- **Cutthroat 462707 is `A_MOD_IGNORE_SHAPESHIFT` on Ambush,** which the parse does not carry: the aura sets
  `Ambush.CastRequirement.CasterForm` on gain and clears it on expiry, rather than a stealth check in
  `ExtraCastCondition`.

## Hooks

- **Maul in Cat Form:** the cat spec registered Maul but never set `ReplaceMHSwing`, so a queued Maul never
  swung, spent no rage, and its queue aura then refused every later queue.
- **`BreakStealth` missing** from Hemorrhage, Slice and Dice and Venom, unlike every other attack, so a cast
  out of Vanish left the rogue with no auto attacks.
- **Prowl broken twice on leaving Cat Form:** `forms.go` and `prowl.go` both hooked the form's expiry. The
  `IsActive` check in `breakProwl` hid it; one hook goes.
- **Form-breaking consumables re-wrapped on every reset:** the druid wrapped each major cooldown's
  `ShouldActivate` in `Reset`, which allocates a closure per cooldown per iteration and depends on
  `majorCooldownManager.reset` copying fresh cooldowns first. It also missed cooldowns with user timings,
  which `shouldActivateHelper` (`major_cooldown.go`) returns on before it reaches `ShouldActivate`. A rule
  that belongs to every class goes in core once.

## Whose state

- **Rend and Tear on Swipe and Primal Bite:** the bonus against a bleeding target was worked out once from
  the primary target and dealt to every target hit. Each target's damage reads its own bleed.
- **Soul Siphon:** `GetAurasWithTag` returns every caster's dots on the target, so another warlock's Agony
  counted.
- **Amplify Curse:** a map read for a target it had no key for answered 0, so an Agony applied without its
  `ApplyEffects` ticked for nothing.

## Silent no-ops and orphans

- **Enrage 5229** lost its armor penalty ("reduces base armor by 27% in Bear Form and 16% in Dire Bear
  Form"), and **Frenzied Regeneration 22842** registered an aura that spends no rage and heals nothing, both
  with no `// TODO:`.
- **Soul Harvesting and Fel Concentration** had their stubs deleted, so points in them did nothing and the
  markers were gone.
- **Orphans:** `SpellIsPassive` lost its only caller, and three empty `register*` functions were left
  uncalled; warlock masks named `WarlockSpellCurseOfAgony`, which nothing registers, and missed Bane of
  Agony.

## One copy of each rule

- **Drain Life, Drain Soul and Wrack** repeated the same `OnTick`, `ApplyEffects` and `ExpectedTickDamage`,
  apart from the tick effect; one builder taking the rank and its tick keeps them from drifting.
- **Rip's combo-point map** was set to nil in `Reset` and allocated again by the first Rip of each iteration;
  every entry is written before it is read, so one map built at registration does.

## Proc triggers

The `wowsims-spells` traps cover the two that silence a row-built trigger: a trigger named after its row
collides with an `AuraConfig` buff of the same label, and a periodic row decodes to `ProcMaskSpellDamage`,
which a bleed tick does not carry.
