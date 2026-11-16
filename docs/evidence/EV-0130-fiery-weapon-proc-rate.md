# EV-0130: Fiery Weapon's proc rate on two two-handed weapons

- Claim: Enchant Weapon - Fiery Weapon 803 rolls its strike 13897 at 6 procs per minute on the enchanted weapon's speed,
  on the hits that roll a chance-on-hit weapon proc.
- Sources: combat log, Retribution paladin, levels 46 and 47, with the enchant on a two-handed mace of speed 3.4 and on
  a two-handed sword of speed 3.9; three logs: the open world, and twice the open world followed by a five-player
  dungeon under a rank 2 Windfury Totem
- Game build: 1.60.1.70938, after the hotfix pushes up to 2026-11-15 06:22 UTC that EV-0119 names, logged in after them
- Date: 2026-11-15
- Pull request: #1055

A unit is the paladin's melee-type hits, avoided hits and proc lines with no gap over 150 ms between neighbours. Its
rolls are one for a white swing, a Holy Strike, a Seal of Command hit, each Seal of Righteousness hit and a Windfury
Totem strike, and two for a Judgement of Command (EV-0119). Units with an avoided hit or a killing blow are left out.
The chance is the maximum-likelihood binomial chance per roll, the rate is that chance times 60 over the weapon's
speed, and the interval is the 95% likelihood interval.

The rate is read on the units with no line of the weapon's own proc, since a landed weapon proc brings one more roll at
a lower chance (the last table).

| Main hand                                 | Units | Rolls | Strikes | Chance per roll | Procs per minute    |
| ----------------------------------------- | ----- | ----- | ------- | --------------- | ------------------- |
| Speed 3.4, units with no weapon proc line | 918   | 1,809 | 626     | 34.60%          | 6.11 (5.72 to 6.50) |
| Speed 3.9, units with no weapon proc line | 352   | 729   | 267     | 36.63%          | 5.63 (5.10 to 6.18) |
| Speed 3.4, every unit                     | 1,164 | 2,492 | 859     | 34.31%          | 6.05 (5.73 to 6.39) |
| Speed 3.9, every unit                     | 466   | 1,051 | 414     | 39.11%          | 6.02 (5.57 to 6.47) |

Both intervals hold 6 procs per minute, the vanilla carry-over value. The two speeds do not tell a rate per minute from
a flat chance: on the units with no weapon proc line the two chances differ by 2.0 points, where 6 procs per minute
asks for 5.0 and a flat chance for none, with a standard error of 2.1. One-handers and the off hand were not measured.

Of 1,401 strikes in the three logs none missed and none was fully resisted, and 2 were reflected. The paladin's other
spells missed nothing either (Consecration 0 of 1,996 ticks), while the party's spells missed 59 times in 1,697 on the
same enemies, so the strike's own roll to hit is not shown.

A landed proc of the weapon itself (the mace's bleed 1306816, the sword's strike 12731) brings a further strike: 9 of
1,833 units show more strikes or weapon procs than their rolls allow, each with the weapon's proc in it. Fitted as one
more roll for each landed weapon proc, with the base chance from the first table:

| Main hand | Units with a landed weapon proc | Landed weapon procs | Chance of the further roll | Base chance | A further roll at the base chance, 2 x log-likelihood |
| --------- | ------------------------------- | ------------------- | -------------------------- | ----------- | ----------------------------------------------------- |
| Speed 3.4 | 215                             | 247                 | 7.9% (1.9 to 14.5)         | 34.60%      | 47.9 worse than none                                  |
| Speed 3.9 | 103                             | 116                 | 18.6% (5.5 to 32.8)        | 36.63%      | 17.6 worse than none                                  |

Among units of one roll, 3 of 59 with a landed weapon proc show two strikes and none of 501 without one does.
