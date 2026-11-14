# EV-0012: Windfury Weapon's strikes log as 439440, a main-hand ability

- Claim: each Windfury Weapon strike is logged under 439440, a main-hand ability, and the rank's 16361 casts it by script with no row linking it to the imbue.
- Sources: combat log, two raw logs of Enhancement shamans (2026-10-05 and 2026-10-07); report codes not recorded
- Game build: not recorded
- Date: 2026-10-05 and 2026-10-07
- Pull request: #937

A swing that brought two strikes shows one Flametongue Totem proc, so the strikes never roll it, and none of 46 strikes
glanced. Flametongue Totem's row hears swings alone. The logs carry no other counts.
