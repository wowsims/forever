# Auras and dots

`spelldata.AuraConfig(row, opts...)` answers a `core.Aura` with the row's name as the label, its
`ActionID`, its duration and its stacks. The game files' -1 for a permanent aura becomes
`core.NeverExpires`. The stacks are `CumulativeAura` where the row states one and `ProcCharges`
otherwise, since the sim keeps stacks and charges in one field. A row that states no duration at all
resolves to 0, and core refuses to activate an aura with that duration. Such an aura needs
`Permanent()` or a duration the caller supplies.

`Label(name)` renames the aura, and two auras of the same spell on one unit need it. `Permanent()`
makes the aura last the iteration, whatever the row says.

```go
aura := warrior.RegisterAura(spelldata.AuraConfig(shieldBlockRank))
spelldata.ParseEffects(&warrior.Character, aura, shieldBlockRank)
```

`spelldata.DotConfig(row, effect, opts...)` answers a `core.DotConfig`: that aura, `TickLength` from
the effect's period, `NumberOfTicks` from the row's duration over that period, `BonusCoefficient` from
the effect's spell power share, and an `OnTick` that deals the effect's own `Average` at the caster's
level, on the stats and multipliers in force at the tick. `OnSnapshot` is nil. It panics where the
effect states no period or the row no duration, so a permanent aura's -1 never resolves into an absurd
number of ticks. A caller replaces `OnTick` for a heal or a value the game files do not state.

`row.TickOutcome(dot)` picks the tick's outcome from the row. The tick can crit where the game files
mark Periodic Can Crit, and it uses the magic hit table where the row's defense type is magic. A
dot's `OnTick` therefore never names an outcome itself.
