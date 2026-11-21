# EV-0130: Fiery Weapon's proc rate on two two-handed weapons

- Claim: Enchant Weapon - Fiery Weapon 803 rolls its strike 13897 at 6 procs per minute on two two-handed weapons, on
  the hits that roll a chance-on-hit weapon proc. Their two speeds do not tell procs per minute from a flat chance.
- Sources: combat log, Retribution paladin, levels 46 and 47, with the enchant on a two-handed mace of speed 3.4 and on
  a two-handed sword of speed 3.9; three logs: the open world, and twice the open world followed by a five-player
  dungeon under a rank 2 Windfury Totem
- Game build: 1.60.1.70938, after the hotfix pushes up to 2026-11-15 06:22 UTC that EV-0119 names, logged in after them
- Date: 2026-11-15
- Pull request: #1056

A unit is the paladin's melee-type hits, avoided hits and proc lines with no gap over 150 ms between neighbours. Its
rolls are one for a white swing, a Holy Strike, a Seal of Command hit, each Seal of Righteousness hit and a Windfury
Totem strike, and two for a Judgement of Command (EV-0119). Units with an avoided hit or a killing blow are left out.
The chance is the maximum-likelihood binomial chance per roll, the rate is that chance times 60 over the weapon's speed,
and the interval is the 95% likelihood interval.

The rate is read on the units with no line of the weapon's own proc, since a landed weapon proc can bring one more roll
(the last table). In the rows over every unit, a unit with more strikes than rolls counts as full.

| Main hand                                 | Units | Rolls | Strikes counted | Chance per roll | Procs per minute    |
| ----------------------------------------- | ----- | ----- | --------------- | --------------- | ------------------- |
| Speed 3.4, units with no weapon proc line | 918   | 1,809 | 626             | 34.60%          | 6.11 (5.72 to 6.50) |
| Speed 3.9, units with no weapon proc line | 352   | 729   | 267             | 36.63%          | 5.63 (5.10 to 6.18) |
| Speed 3.4, every unit                     | 1,164 | 2,492 | 855 of 859      | 34.31%          | 6.05 (5.73 to 6.39) |
| Speed 3.9, every unit                     | 466   | 1,051 | 411 of 414      | 39.11%          | 6.02 (5.57 to 6.47) |

Both intervals hold 6 procs per minute, the vanilla carry-over value. The two speeds do not tell a rate per minute from
a flat chance: on the units with no weapon proc line the two chances differ by 2.0 points, where 6 procs per minute asks
for 5.0 and a flat chance for none, with a standard error of 2.1. One-handers and the off hand were not measured.

Of the 1,401 strikes in the three logs, in a counted unit or not, none missed and none was fully resisted, and 2 were
reflected. The count does not show whether the strike can miss. The paladin's spell hit modifier read 1 in the ledger,
taken as 1%, and 1,080 of the strikes were on targets three or more levels below the paladin, where the sim's spell miss
is 0 as well. The rest:

| Target's level less the paladin's | Strikes | Missed | The sim's spell miss at that hit |
| --------------------------------- | ------- | ------ | -------------------------------- |
| -2                                | 167     | 0      | 1%                               |
| -1                                | 69      | 0      | 2%                               |
| 0                                 | 47      | 0      | 3%                               |
| +1                                | 38      | 0      | 4%                               |

On these 321 the sim's table asks for 6.0 misses, and none comes 0.2% of the time. The paladin's Consecration, also a
magic spell with no Always Hit on its row, missed none of 576 ticks at the same four differences, where the table asks
for 9.3. No target was more than one level above the paladin.

Other players' magic spells with no Always Hit missed 60 times in 2,974 direct hits in the same logs, unevenly. Of the
seven casters with 100 hits or more, one missed 28 of 791, two missed 1 each of 395 and 100, and four missed none of
their 636. The log does not carry a player's level, so their counts cannot be laid against the sim's table.

A landed proc of the weapon itself (the mace's bleed 1306816, the sword's strike 12731) brings a further strike: 9 of
1,833 units, those with a killing blow among them, show more strikes or weapon procs than their rolls allow, each with
the weapon's proc in it. The further roll's chance is fitted on the units with a landed weapon proc, one more roll for
each, with the base chance from the first table and a bootstrap interval over units. The last column refits one chance
for every roll over every unit, with and without the further roll, and gives how much worse the fit with it is, as twice
the log-likelihood.

| Main hand | Units with a landed weapon proc | Landed weapon procs | Chance of the further roll | Base chance | A further roll at the base chance |
| --------- | ------------------------------- | ------------------- | -------------------------- | ----------- | --------------------------------- |
| Speed 3.4 | 215                             | 247                 | 7.9% (1.9 to 14.5)         | 34.60%      | 47.9 worse                        |
| Speed 3.9 | 103                             | 116                 | 18.6% (5.5 to 32.8)        | 36.63%      | 17.6 worse                        |

Among units of one roll, 3 of 59 with a landed weapon proc show two strikes and none of 501 without one does.

These chances pool every landed weapon proc, whatever cast it. One that a white swing cast brings one more roll at about
the base chance, and one that anything else cast brings none.
