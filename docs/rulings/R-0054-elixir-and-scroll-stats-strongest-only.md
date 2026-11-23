# R-0054: Each stat counts its strongest elixir or scroll, and five raid buffs overwrite them

- Skill: wowsims-consumables-items
- Code: `sim/common/shared/consumable_stats.go`, `consumableAura` with `spelldata.ExclusivePerStat(core.ElixirStatCategory)`, `spelldata.Exclusive(core.ElixirCritCategory, false)` and `CountsWithElixirs`; `sim/core/spelldata/parse_effects.go`, `Overwrites`; `tools/database/buffmanifest/buffs.go`, `ElixirCategory`; `sim/core/consumes.go`, `selectedConsumableIDs`; `tools/database/consumable_groups.go`, `rowGroupOf` and `ElixirCritGroup`; `tools/database/overrides.go`, `ConsumablesCountWithElixirs`; `ui/features/settings/model/consumables.ts`, `overwrittenTooltips` and `countsWithTheElixirs`
- Pull request: #1185, and this rule's scroll part in #TBD
- Evidence: EV-0185

The settings offer one elixir per exclusive group and one scroll per scroll group, and the sim takes the
first of each group the spec names. An elixir's group is the first aura its on-use spell applies, but the
five elixirs that grant the 2% crit (the Mongoose, the Grizzly, the Owl, Sages and the Whale) share one
Critical Strike group, so each goes beside an elixir of its first stat: Elixir of the Grizzly beside
Elixir of Greater Strength, Elixir of the Mongoose beside Elixir of Greater Agility. Juju Power 12451,
Juju Might 12460 and Bogling Root 5206 belong to no group and are picked beside the elixirs.

Across the groups, each stat the elixirs and scrolls grant counts only the strongest one's amount, and
only one elixir's 2% crit counts: beside Elixir of the Mongoose, Elixir of Ferocity adds its Strength and
no Agility. The Jujus and Bogling Root count with the elixirs the same way, which their db.json
`counts_with_elixirs` marks: beside Elixir of Minor Force, Bogling Root adds nothing. A scroll counts with
them by its type: beside Juju Power's 30 Strength, Scroll of Strength IV adds none of its 17, and beside
Elixir of Cunning's 25 Intellect, Scroll of Intellect IV adds none of its 16. Flasks, food and the other
consumables stack with the elixirs and scrolls.

The owner ruled the scroll part on 2026-11-23 after Juju Power removed Scroll of Strength's buff in game
(EV-0185). It counts every scroll with the elixirs, Scroll of Intellect beside Elixir of Cunning
included, which an earlier in-game check read as stacking.

Five raid buffs overwrite the elixirs and scrolls of their stat, and the stronger of the raid buff and
the strongest elixir or scroll counts. Arcane Intellect does so for Intellect, Prayer of Spirit, the
group form of Divine Spirit and the sim's Spirit raid buff, for Spirit, and Power Word: Fortitude for
Stamina: under it, Elixir of Brute Force adds its 18 Strength and no Stamina. Moonkin Aura and Leader of
the Pack, which share one crit, do so for the elixirs' crit. No other raid buff overwrites an elixir or a
scroll: Battle Shout and the totems stack with them, and so does every stat of Mark of the Wild.

The sim overwrites an elixir or a scroll by bidding the raid buff's stat in the elixirs' category without
counting it there (`spelldata.Overwrites`), and the raid buff counts by its own category alone. One case
this leaves open cannot arise: an elixir or scroll granting as much as the raid buff or more would count
beside it. Every elixir and scroll in the database grants less, which
`TestNoElixirOrScrollOutbidsTheRaidBuffOfItsStat` holds.
