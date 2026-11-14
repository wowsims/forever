# EV-0012: Windfury Weapon's strikes log as 439440, and Shield Slam and Shield Bash use the main hand's weapon procs

- Claim: each Windfury Weapon strike is logged under 439440, a main-hand ability, and never rolls Flametongue Totem; the
  main-hand weapon's equip procs hear Shield Slam and Shield Bash.
- Sources: two raw combat logs of Enhancement shamans (2026-10-05 and 2026-10-07) and raw combat logs of a protection
  warrior at game build 1.60.1.70862; report codes not recorded
- Game build: not recorded for the shaman logs; 1.60.1.70862 for the warrior logs
- Date: 2026-10-05 and 2026-10-07 for the shaman logs; not recorded for the warrior logs
- Pull request: #937; the warrior log's build is in its commit 85f4100f62

## Windfury Weapon's strikes

A swing that brought two strikes shows one Flametongue Totem proc, so the strikes never roll it, and none of 46 strikes
glanced. Flametongue Totem's row hears swings alone. The rank's 16361 casts the strike by script, and no row links it to
the imbue. The logs carry no other counts.

## Shield Slam and Shield Bash

The main-hand Dreadfrost Saber proc fired on 54 of 80 isolated Shield Slams and 7 of 8 Shield Bashes, 184 of 237 white
swings, and on none of 74 Shield Block or 9 Battle Shout casts. The rule replaces the earlier off-hand rule, by the
owner's decision of 2026-11-10.
