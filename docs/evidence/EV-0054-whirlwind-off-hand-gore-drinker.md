# EV-0054: Whirlwind's off-hand strike, Gore Drinker, Weaponmaster and Sweeping Strikes

- Claim: Whirlwind's off-hand strike starts Flurry on a crit and spends no charge, Gore Drinker spends a charge and heals
  once per landed Whirlwind or Cleave strike, Weaponmaster (Sword) rolls once a two-handed Whirlwind cast, and Sweeping
  Strikes brings one main-hand hit a Whirlwind.
- Sources: Forever combat logs of Fury and Arms warriors; local logs of 2026-11-10 and 2026-11-11 hold 97 Whirlwind casts
  by eight dual-wield Fury warriors with Flurry talented; report codes not recorded
- Game build: not recorded
- Date: 2026-11-10 and 2026-11-11 for the Fury logs; not recorded for the rest
- Pull request: #962

In Forever logs the off-hand strike is its own spell, 462891 "Whirlwind", logged in the same millisecond as the 1680
cast, at about 0.6 times the main hand's damage.

| Whirlwind cast      | Casts          | Flurry gained at the cast | Within 130 ms |
| ------------------- | -------------- | ------------------------- | ------------- |
| Main hand crit only | 6              | 6                         | 6             |
| Both hands crit     | 4              | 3                         | 4             |
| Off hand crit only  | 18 (6 players) | 16                        | 18            |
| No crit             | 42             | 1                         | 2             |

An off-hand-only cast has no white swing or other physical crit within 12 ms. 12 casts with Flurry up, a landed off-hand
strike that did not crit and no white swing within 300 ms lost no charge, nor did 4 dodged or parried Whirlwind
strikes; every charge spent near a Whirlwind lines up with a white swing.

| Case                                        | Logs                                                                                                          |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| Gore Drinker                                | one charge and heal per landed strike, each hand and each target, capped by the charges left; see below       |
| Weaponmaster (Sword), main hand             | 7 extra attacks in 100 two-handed Whirlwinds; by landed targets 1: 0 of 7, 2: 1 of 26, 3: 5 of 44, 4: 1 of 23 |
| Sweeping Strikes, main hand                 | exactly one hit and one charge at 2 to 4 targets in 43 casts                                                  |
| Weaponmaster and Sweeping Strikes, off hand | no dual-wield Arms warrior in any log                                                                         |
| Improved Whirlwind                          | both hands +27% on the closest target (two warriors); no talent: +0.4%                                        |
| Deep Wounds                                 | 4 fresh applications from off-hand-only crits                                                                 |
| Unbridled Wrath, Windfury Totem             | no Whirlwind roll: 0 of 75, 0 of 363                                                                          |
| Off-hand misses                             | 5 misses in about 2,090 off-hand strikes, with Furious Precision 3/3 on every dual-wielder                    |

Gore Drinker's per-strike rule fits 76 of 82 clean Whirlwind casts, where one per cast fits only the 25 where the two
coincide. On Cleave, 2 landed strikes gave 2 heals (13 casts at 3 charges, 10 at 2). Two main-hand and two off-hand
strikes landed at 3 charges gave 3 heals, and a dodged main hand with one landed off-hand strike gave 1.

Which target's hit takes the Weaponmaster roll is not measured. That Sweeping Strikes' own hits never roll Weaponmaster
is recorded only as a statement in the comment at the registration, with no count.
