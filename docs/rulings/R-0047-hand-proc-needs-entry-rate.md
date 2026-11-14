# R-0047: A hand-written proc registers nothing while its item entry states no rate

- Skill: wowsims-procs
- Code: `sim/common/forever/items_weapons.go`, `newEntryRateItemEffect`; `tools/database/gen_effects.go`, `handProcStatesNoRate`
- Pull request: #905
- Evidence: not needed: a repo process rule

A hand-written proc that runs at the procs per minute on its item's entry registers nothing while the entry states none,
and no rate is made up for it: the item keeps its stats, and the generator lists it in the gear tab's missing-effect
notices until a rate reaches the entry.
