# EV-0125: Blackblade of Shahram's rate and what Shahram casts

- Claim: Blackblade of Shahram 12592 summons Shahram (16602) about 1.1 times a minute on landed main-hand hits. Each
  Shahram casts exactly one of his six spells, starts its hardcast 1.03 to 1.83 s after the summon, and every copy of
  his Will or Blessing runs on its own timer.
- Sources: public reports on forever.warcraftlogs.com, two paladins wearing the sword in the main hand:
  NXMZV9xjvWpB7GtP fight 13 and D463FZBX1C9maJ7b fight 7.
- Game build: 1.60.1.70938; the hotfix state and relogs were not recorded.
- Date: 2026-11-13 (NXMZV9xjvWpB7GtP) and 2026-11-14 (D463FZBX1C9maJ7b)
- Pull request: #1051

A proc is one summon line of 16602 by the wearer. A hit is a landed damage line of the wearer on an enemy (hit, crit,
block, glance or crush), or a melee cast that lands with no damage line, counted at its debuff. The sword's chain and
the ticks of damage over time are left out. Hits that land within 100 ms of each other form one cluster, since a white
swing and the Seal of Righteousness hit it triggers share a millisecond. A proc belongs to the cluster whose first hit
lies between 400 ms before the proc and 25 ms after it: the summon is logged at the cast line of a melee cast, 17 ms before
its damage line, and about 17 ms after a white swing. The rate is the chance per roll times 60 over the sword's 3.5 s
speed.

## Which hits roll

Ranged-type hits (Judgement of Righteousness) roll nothing, 0 of 95 isolated ones, as EV-0119 found. A Judgement of
Command rolls twice, as EV-0119 found. A Seal of Righteousness hit rolls the sword as well: clusters of a white swing
and its seal hit procced 12 times in 98 (12.2%), single white swings 10 times in 168 (6.0%), and a fit with the seal
rolling explains the clusters better than one without it (log-likelihood -277.6 against -282.0).

## Hits

| Report           | Procs | White swings | Melee casts | Judgement of Command | Windfury Totem strikes | Seal of Righteousness | Ranged-type |
| ---------------- | ----- | ------------ | ----------- | -------------------- | ---------------------- | --------------------- | ----------- |
| D463FZBX1C9maJ7b | 43    | 231          | 156         | 65                   | 50                     | 178                   | 93          |
| NXMZV9xjvWpB7GtP | 59    | 256          | 177         | 101                  | 55                     | 110                   | 46          |

## Rates

| Cut                                                                     | Procs / rolls            | Chance             | PPM [95%]         |
| ----------------------------------------------------------------------- | ------------------------ | ------------------ | ----------------- |
| White swings only                                                       | 102 / 487                | 20.9%              | 3.59              |
| White swings, melee casts, Judgement of Command and Windfury, once each | 102 / 1,091              | 9.35%              | 1.60              |
| Every hit that rolls, the seal's included, Judgement of Command twice   | 102 / 1,545              | 6.60%              | 1.13 [0.94, 1.36] |
| The same rolls, fitted over every cluster                               | 102 procs, 899 clusters  | 6.55% [5.35, 7.85] | 1.12 [0.92, 1.35] |
| Isolated hits only, fitted                                              | 27 / 337 hits, 436 rolls | 6.30% [4.30, 8.80] | 1.08 [0.74, 1.51] |

The first two rows are upper bounds: they leave out hits that roll. The third is the lower bound, and reads 0.99 and
1.26 for the two wearers alone. The fit over every cluster counts a cluster of k rolls as one chance of
1 - (1 - p)^k, so two procs from one cluster are not double counted. The isolated cut keeps hits with no other hit
within 400 ms either side. The rate is 1.1, which both fits round to at the nearest 0.05 step. Two wearers leave the
interval wide, so the item keeps the gear tab's notice for an unvalidated rate.

## Shahram's spells

Every one of the 102 summons cast one spell and nothing else.

| Spell                     | Picks |
| ------------------------- | ----- |
| Flames of Shahram 16596   | 33    |
| Will of Shahram 16598     | 30    |
| Might of Shahram 16600    | 13    |
| Blessing of Shahram 16599 | 9     |
| Fist of Shahram 16601     | 9     |
| Curse of Shahram 16597    | 8     |

- The hardcast begins 1,029 to 1,825 ms after the summon, median 1,421 ms, and runs 984 to 1,028 ms, median 1,002 ms.
- Flames hit 0 to 7 enemies a cast. Of 87 hit lines none crit, 41 were partly resisted, and the largest amount before
  resistance was 158.
- Will lands on the wearer alone. Each copy is removed 19,979 to 20,028 ms after its own application, with no refresh
  and no stack count, and two copies were up at once.
- Blessing ticks 50 mana and 50 health every 5 s on the party around Shahram. Two Blessings cast 13.0 s apart both
  ticked, each on its own 5 s beat.
- Fist lands on Shahram himself, 9 of 9 times.
