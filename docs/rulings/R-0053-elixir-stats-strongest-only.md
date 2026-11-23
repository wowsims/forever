# R-0053: Every elixir counts, but each stat only at its strongest elixir's amount

- Skill: wowsims-consumables-items
- Code: `sim/common/shared/consumable_stats.go`, `consumableAura` with `spelldata.ExclusivePerStat(elixirStatCategory)` (the category `"ElixirStat"`), `spelldata.Exclusive(elixirCritCategory, false)` (the category `"ElixirCrit"`) and `statsCountWithTheElixirs`; `tools/database/consumable_groups.go`, `rowGroupOf`
- Pull request: not recorded
- Evidence: EV-0185

Elixirs belong to no exclusive group: any set of them can be used, and the auras of all of them stay up.
Each stat they grant counts only the strongest elixir's amount, and only one elixir's 2% crit counts.
Juju Power 12451, Juju Might 12460 and Bogling Root 5206 count with the elixirs the same way. A
scroll's stat counts only while it outbids both the raid buffs and the elixirs of the same stat: a
stronger raid buff or elixir overwrites it, and the scroll then adds nothing while the raid buff and the
elixir each keep their stat. Flasks, food and the other consumables stack with the elixirs.
