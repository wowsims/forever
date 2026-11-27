# R-0057: A ProcChance column twice the tooltip's chance is the chance in the tooltip's area, halved elsewhere

- Skill: wowsims-procs
- Code: `sim/common/forever/items_trinkets.go`, `halvedOutside`, The Lion Horn of Stormwind 14557 and Uther's Strength
  11302
- Pull request: #1089
- Evidence: not needed: the game files state it

An item proc whose tooltip doubles its chance in an area, and whose ProcChance column holds twice the chance the
tooltip states, rolls the whole column in an encounter in one of the areas and half of it in any other encounter. The
Lion Horn of Stormwind's 20847 holds 10, which its tooltip prints as `${$h/2}%`, 5%, doubled in Strongholds and Cities.
Uther's Strength's 8397 holds 4 and states the 2% on its first effect, doubled in Wasteland and Haunted areas. The sim
fixes the encounter's areas for the run, so the trigger takes its chance when it is built, and the stats sheet lists the
item's doubling under each area.

A proc whose stated rate is the chance outside the area takes it as it is and doubles it in the area instead: Darkmoon
Card: Heroism's entry rate is its rate anywhere, twice that in Strongholds and Cities.

The owner ruled it for The Lion Horn of Stormwind on 2026-11-17 in #1089; Uther's Strength follows it (#1255).
