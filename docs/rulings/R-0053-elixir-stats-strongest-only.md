# R-0053: Every elixir counts, but each stat only at its strongest elixir's amount

- Skill: wowsims-consumables-items
- Code: `sim/common/shared/consumable_stats.go`, `consumableAura` with `spelldata.ExclusivePerStat(elixirStatCategory)` (the category `"ElixirStat"`), `spelldata.Exclusive(elixirCritCategory, false)` (the category `"ElixirCrit"`), `StatsCountWithTheElixirs`, and for the scrolls `spelldata.ExclusivePerStatBehind(core.StatBuffCategory)`; `sim/core/spelldata/parse_effects_exclusive.go`, `bidPerStatBehind`; `tools/database/consumable_groups.go`, `rowGroupOf`; `ui/features/settings/model/consumables.ts`, `overwrittenTooltips`
- Pull request: #1185
- Evidence: EV-0185

Elixirs belong to no exclusive group: any set of them can be used, and the auras of all of them stay up.
Each stat they grant counts only the strongest elixir's amount, and only one elixir's 2% crit counts.
Juju Power 12451, Juju Might 12460 and Bogling Root 5206 count with the elixirs the same way. Flasks,
food and the other consumables stack with the elixirs.

The scrolls follow the owner's ruling, which no evidence file holds. A scroll's stat counts only while it
outbids both the elixirs and the StatBuff raid buffs of the same stat (Arcane Intellect, Prayer of Spirit
and Power Word: Fortitude). A stronger raid buff or elixir overwrites it, and the scroll then adds
nothing while the raid buff and the elixir each keep their stat. Mark of the Wild stacks with the
scrolls.
