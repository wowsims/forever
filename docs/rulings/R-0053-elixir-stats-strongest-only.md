# R-0053: Each stat counts its strongest elixir, and five raid buffs overwrite the elixirs of their stat

Replaced by: R-0054

- Skill: wowsims-consumables-items
- Code: `sim/common/shared/consumable_stats.go`, `consumableAura` with `spelldata.ExclusivePerStat(core.ElixirStatCategory)`, `spelldata.Exclusive(core.ElixirCritCategory, false)` and `CountsWithElixirs`, and for the scrolls `spelldata.ExclusivePerStat(core.StatBuffCategory)`; `sim/core/spelldata/parse_effects.go`, `Overwrites`; `tools/database/buffmanifest/buffs.go`, `ElixirCategory`; `sim/core/consumes.go`, `selectedConsumableIDs`; `tools/database/consumable_groups.go`, `rowGroupOf` and `ElixirCritGroup`; `tools/database/overrides.go`, `ConsumablesCountWithElixirs`; `ui/features/settings/model/consumables.ts`, `overwrittenTooltips`
- Pull request: #1185
- Evidence: EV-0185

The settings offer one elixir per exclusive group, and the sim takes the first of each group the spec
names. An elixir's group is the first aura its on-use spell applies, but the five elixirs that grant
the 2% crit (the Mongoose, the Grizzly, the Owl, Sages and the Whale) share one Critical Strike group,
so each goes beside an elixir of its first stat: Elixir of the Grizzly beside Elixir of Greater
Strength, Elixir of the Mongoose beside Elixir of Greater Agility. Juju Power 12451, Juju Might 12460
and Bogling Root 5206 belong to no group and are picked beside the elixirs.

Across the groups, each stat the elixirs grant counts only the strongest elixir's amount, and only one
elixir's 2% crit counts: beside Elixir of the Mongoose, Elixir of Ferocity adds its Strength and no
Agility. The Jujus and Bogling Root count with the elixirs the same way, which their db.json
`counts_with_elixirs` marks: beside Elixir of Minor Force, Bogling Root adds nothing. Flasks, food and the other
consumables stack with the elixirs.

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
