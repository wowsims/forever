# R-0056: A druid's form attacks roll chance-on-hit weapon procs as any other melee hit does

- Skill: wowsims-procs
- Code: `sim/core/aura_helpers.go`, `canProcFrom`, which treats a druid in a form like any other unit; Flame Wrath's
  `shared.DamageProcExtras[11809]` in `sim/common/forever/items_weapons.go`
- Pull request: #1280
- Evidence: EV-0022, EV-0119, EV-0122, EV-0151

In the sim a druid in Cat, Bear or Dire Bear Form rolls a chance-on-hit weapon on every landed melee hit: the white
swing and the form's attacks (Maul, Swipe, Claw, Rake, Shred, Ferocious Bite and the rest) alike. In game the form's
attacks roll none, and only the white swing does. The owner calls this a known game bug that will be fixed in game, so
the sim ignores it (owner ruling, 2026-11-27).

A rate measured on druids in a form counts their white swings alone, and most rates leave those druids out.
