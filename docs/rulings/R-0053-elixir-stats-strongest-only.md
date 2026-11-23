# R-0053: Each stat counts its strongest elixir, and five raid buffs overwrite the elixirs of their stat

- Skill: wowsims-consumables-items
- Code: `sim/common/shared/consumable_stats.go`, `consumableAura` with `spelldata.ExclusivePerStat(core.ElixirStatCategory)`, `spelldata.Exclusive(core.ElixirCritCategory, false)` and `StatsCountWithTheElixirs`, and for the scrolls `spelldata.ExclusivePerStat(core.StatBuffCategory)`; `sim/core/spelldata/parse_effects.go`, `Overwrites`; `tools/database/buffmanifest/buffs.go`, `ElixirCategory`; `tools/database/consumable_groups.go`, `rowGroupOf`; `ui/features/settings/model/consumables.ts`, `overwrittenTooltips`
- Pull request: #1185
- Evidence: EV-0185

Elixirs belong to no exclusive group: any set of them can be used, and the auras of all of them stay up.
Each stat they grant counts only the strongest elixir's amount, and only one elixir's 2% crit counts.
Juju Power 12451, Juju Might 12460 and Bogling Root 5206 count with the elixirs the same way. Flasks,
food and the other consumables stack with the elixirs.

Five raid buffs overwrite the elixirs of their stat, and the stronger of the raid buff and the
strongest elixir counts. Arcane Intellect does so for Intellect, Prayer of Spirit, the group form of
Divine Spirit and the sim's Spirit raid buff, for Spirit, and Power Word: Fortitude for Stamina: under
it, Elixir of Brute Force adds its 18 Strength and no Stamina. Moonkin Aura and Leader of the Pack,
which share one crit, do so for the elixirs' crit. No other raid buff overwrites an elixir: Battle
Shout and the totems stack with the elixirs, and so does every stat of Mark of the Wild.

A scroll counts only while it outbids the raid buff of its stat in core's `StatBuff` category: Arcane
Intellect, Prayer of Spirit or Power Word: Fortitude. A scroll stacks with the elixirs. With no raid
buff up, Elixir of Cunning's 25 Intellect and Scroll of Intellect IV's 16 add 41; under Arcane
Intellect's 31, neither adds any. Mark of the Wild stacks with the scrolls.

The sim overwrites an elixir by bidding the raid buff's stat in the elixirs' category without counting
it there (`spelldata.Overwrites`), and the raid buff counts by its own category alone. Two cases this
leaves open cannot arise. An elixir granting as much as the raid buff or more would count beside it,
and a scroll outbidding the raid buff would count while the raid buff, adding nothing, still kept the
elixirs out. Every elixir and scroll in the database grants less, which
`TestNoElixirOrScrollOutbidsTheRaidBuffOfItsStat` holds.
