# Auras and dots

`spelldata.AuraConfig(row, opts...)` answers a `core.Aura` with the row's name as the label, its
`ActionID`, its duration and its stacks. The game files' -1 for a permanent aura becomes
`core.NeverExpires`. The stacks are `CumulativeAura` where the row states one and `ProcCharges`
otherwise, since the sim keeps stacks and charges in one field. A row that states no duration at all
resolves to 0, and core refuses to activate an aura with that duration. Such an aura needs
`Permanent()` or a duration the caller supplies.

`Label(name)` renames the aura, and two auras of the same spell on one unit need it.
`RankLabel(rank, owner)` labels it by `rank.RankedLabel(owner)` ("Holy Fire Rank 8 - Priest (#1)"): a
caster registers every rank, and ranks that share one label share one aura, so one cast would tick
every rank's dot; the owner's label keeps two casters' debuffs on one target apart. A label built by
hand, such as a proc trigger's `Name`, takes `rank.RankedLabel(owner)` directly. `Permanent()` makes
the aura last the iteration, whatever the row says.

```go
aura := warrior.RegisterAura(spelldata.AuraConfig(shieldBlockRank))
spelldata.ParseEffects(&warrior.Character, aura, shieldBlockRank)
```

`spelldata.DotConfig(row, effect, opts...)` answers a `core.DotConfig`: that aura, `TickLength` from
the effect's period, `NumberOfTicks` from the row's duration over that period, `BonusCoefficient` from
the effect's spell power share, and an `OnTick` that deals one `Roll` of the effect at the caster's
level, on the stats and multipliers in force at the tick. Every tick rolls on its own where the effect
states a `Variance` (Caustic Smog Potion's 1252543 ticks 31 to 35); an effect that states none deals
its `Average`, and its ticks draw no random number. `OnSnapshot` is nil. It panics where the
effect states no period or the row no duration, so a permanent aura's -1 never resolves into an absurd
number of ticks. A caller replaces `OnTick` for a heal or a value the game files do not state.

A ranked row's dot also carries its name as `RankGroup` and the row's rank as `Rank`, and a caster
keeps one rank of a periodic on a unit; HoTs, self HoTs included, follow the same rule. A rank at or
above the one up ends it at once and starts fresh, so a cast of the same rank is a plain refresh. A
lower rank does not land while a higher one from the same caster is up, and the higher rank keeps
ticking on its own schedule and expires on time.

- A row whose only effect that does something is the periodic (Shadow Word: Pain, Renew, Corruption,
  Rend, and Rip and Rupture, whose dummy effects do nothing by themselves) sets
  `NotUnderHigherRank` in its `CastRequirement`: the lower rank is refused at the cast check with "a
  more powerful spell is active", so `CanCast` answers false and the cast spends no cost, global
  cooldown, cast time or cooldown. The check reads the dot the cast applies, `Spell.Dot(target)`,
  which follows RelatedDotSpell, else the caster's self periodic. A rank registered with `Proc()` has
  no cast requirement: it casts, and `Dot.Apply` skips its periodic.
- A row with any other effect casts as usual and every other effect lands; `Dot.Apply` skips only the
  periodic. Direct damage or a heal (Holy Fire), an energize (Garrote's combo point) and a second aura
  (Insect Swarm, Entangling Roots) all count; a dummy or script effect does not. A caller that
  touches the dot after `Apply`, to set stacks or read its expiry, checks `dot.IsActive()` first.

A ground effect (`IsAOE`, Flamestrike, Consecration) is one per caster: its dot keeps no rank, so a
cast of any rank, lower included, ends the one down and starts its own, and no rank is refused at the
cast.

The dot's aura is labelled with the ranked name, "Shadow Word: Pain Rank 2", unless an option labels
it, so each rank has an aura of its own; core panics where two ranks of a group share one. Two casters
never affect each other's ranks, and channels keep no group, since a new channel already ends the one
running. A dot built by hand joins a group only when its `core.DotConfig` sets `RankGroup` and `Rank`,
and its spell is refused at the cast only when its `core.CastRequirement` sets `NotUnderHigherRank`;
core compares the dots' `Rank`, never the spell's.

`row.TickOutcome(dot)` picks the tick's outcome from the row. A tick never misses: the dot's hit is
rolled once, by the cast or proc that applies it, and the dot goes on only where that roll lands. The
tick can crit where the game files mark Periodic Can Crit, on the spell crit where the dot spell's
defense type is magic and the physical crit otherwise, and takes its partial resist. A dot's `OnTick`
therefore never names an outcome itself, and a dot whose application deals no direct damage still
rolls the hit at application (`CalcOutcome` on the hit table, then `Apply` where it landed).
