# EV-0125: Blackblade of Shahram's rate and what Shahram casts

- Claim: Blackblade of Shahram 12592 summons Shahram (16602) 1.0 times a minute on landed main-hand hits. Each
  Shahram casts exactly one of his six spells, picked about 7:5:2:2:2:2 (Flames, Will, Might, Blessing, Fist, Curse),
  and starts its hardcast about 1.0 to 1.8 s after the summon. Will lands on the wearer alone, Blessing on the players
  of the wearer's party and never on a pet, Flames, Curse and Might on enemies, and Fist on Shahram himself; every copy
  of his Will or Blessing runs on its own timer. Flames can miss and crit; the sim rolls it on the wearer's spell hit
  and spell crit, an owner ruling.
- Sources: public reports on forever.warcraftlogs.com, two paladins wearing the sword in the main hand:
  NXMZV9xjvWpB7GtP fight 13 and D463FZBX1C9maJ7b fight 7.
- Game build: 1.60.1.70938; the hotfix state and relogs were not recorded.
- Date: 2026-11-13 (NXMZV9xjvWpB7GtP) and 2026-11-14 (D463FZBX1C9maJ7b)
- Pull request: #1051
- Later count: Blackblade of Shahram's rate and Shahram's spells, §Count of the 2026-11-13 to 2026-11-24 reports

A proc is one summon line of 16602 by the wearer. A hit is a landed damage line of the wearer on an enemy (hit, crit,
block, glance or crush), or a melee cast that lands with no damage line, counted at its debuff. The sword's chain and
the ticks of damage over time are left out. Hits that land within 100 ms of each other form one cluster, since a white
swing and the Seal of Righteousness hit it triggers share a millisecond. A proc belongs to the cluster whose first hit
lies between 400 ms before the proc and 25 ms after it: the summon is logged at the cast line of a melee cast, 17 ms
before its damage line, and about 17 ms after a white swing.

## Which hits roll

Ranged-type hits (Judgement of Righteousness) roll nothing, 0 of 95 isolated ones, and a Judgement of Command rolled
twice up to hotfix push 114399 (2026-11-25), as EV-0119 found, and rolls once from that push (EV-0155). A Seal of
Righteousness hit rolls the sword as well: clusters of a white swing and its seal hit procced 12 times in 98 (12.2%),
single white swings 10 times in 168 (6.0%), and a fit with the seal rolling explains the clusters better than one
without it (log-likelihood -277.6 against -282.0).

## Rate

- Count: §Count of the 2026-11-13 to 2026-11-24 reports
- Proc: weapon proc 16602
- Rate: 1 PPM, measured

1,332 procs on 23,438 rolling hits, 13 wearers (10 warriors and 3 hunters), 20 runs, in the main hand: all hits 0.974
[0.919, 1.028]. The rate is all hits, which the one-hand white and specials rates agree with. One-hand white: 1.06
[0.91, 1.23], 170 procs on 2,755 lone units, 13 wearers. Specials: 0.98 [0.80, 1.22], 80 procs on 1,404 lone units, 13
wearers; Rend landings give 33 procs on 499 of them. Two-roll abilities, priced as one roll: 0.81, 12 procs on 255 lone
units (Execute 12 on 249 and Mongoose Bite 0 on 6), left out of the rate. Direct: window [-20, 100] ms, median 0 ms,
98th percentile 0 ms, 244 isolated hits.

The count joins summons within 50 ms of each other into one proc, as it joins the events of any proc. Counting each
summon as a proc of its own adds 46 procs, 28 that log at the same millisecond as another summon and 18 that log 1 to 41
ms after one, and gives all hits 1.008 [0.953, 1.060] on 1,378 procs. Every group of summons at one millisecond lies
beside as many rolling hits or more. Paladins are left out of the rate; their lone white swings, a check that sets no
rate, read 1.07 [0.94, 1.21], 113 procs on 1,817 lone units, 29 wearers.

| Class   | Wearers | Runs | Procs |   Hits |  PPM | PPM, each summon a proc |
| ------- | ------: | ---: | ----: | -----: | ---: | ----------------------: |
| Warrior |      10 |   16 | 1,288 | 22,808 | 0.97 |                    1.00 |
| Hunter  |       3 |    4 |    44 |    630 | 1.20 |                    1.20 |

## Shahram's spells

- Count: §Count of the 2026-11-13 to 2026-11-24 reports
- Weights: 7:5:2:2:2:2, observed
- Flames: rolls the wearer's spell hit and spell crit, an owner ruling

A pick is the spell of the begincast Shahram logs 0.5 to 3 s after his summon. 3,689 of the 3,754 summons have one, and
each Shahram casts that spell and nothing else. Of the 65 summons without one, 56 log nothing more from that Shahram
within 4 s, 8 an application of Will and 1 of Blessing, and 28 fall within 1.5 s of a fight's end.

| Spell                     | Picks | Share | 95% interval   |
| ------------------------- | ----: | ----: | -------------- |
| Flames of Shahram 16596   | 1,272 | 34.5% | [33.0%, 36.0%] |
| Will of Shahram 16598     |   943 | 25.6% | [24.2%, 27.0%] |
| Might of Shahram 16600    |   403 | 10.9% | [10.0%, 12.0%] |
| Blessing of Shahram 16599 |   370 | 10.0% | [9.1%, 11.0%]  |
| Fist of Shahram 16601     |   353 |  9.6% | [8.7%, 10.6%]  |
| Curse of Shahram 16597    |   348 |  9.4% | [8.5%, 10.4%]  |

The picks come from 56 wearers (34 paladins, 18 warriors and 4 hunters) in 63 reports. Weights of 3:3:1:1:1:1 do not
fit them (chi-square 53.9, p = 2e-10): Flames comes up more often than Will (p = 3e-12), and the four other spells
equally often (p = 0.17). The weights 7:5:2:2:2:2 are a round ratio the owner chose to fit these counts (chi-square
5.78, p = 0.33). The picks do not differ between the 45 wearers with 8 picks or more
(p = 0.43), between the reports on 70938, 70959 and 71037 (p = 0.84), or before and from 2026-11-23 (p = 0.19).

- The hardcast begins 917 to 2,570 ms after the summon, median 1,419 ms, 2nd to 98th percentile 1,017 to 1,804 ms, and
  3,506 of the 3,689 begin within 1,030 to 1,830 ms. It runs a median 1,003 ms, 2nd to 98th percentile 974 to 1,058 ms.
- Flames lands on the enemies around Shahram, 0 to 26 a cast: 3,238 of its 3,246 damage lines are on creatures and their
  pets, and the other 8 on players and a player's pet whom the wearer's side damages and never heals or buffs. Of the
  3,246 lines, 18 are misses, 83 immune and 3,145 land: 2,314 hits, 812 partly resisted hits, 16 crits and 3 partly
  resisted crits, so 19 of the 3,145 crit. Normal hits deal up to 182 (median 125), and 14 of the 16 crits with no
  partial resist deal 151 to 231, within 1.5 times the row's 90 to 160. The sim rolls Flames on the wearer's spell hit
  and spell crit, for the spell crit multiplier; that is the owner's ruling, not a fit to these counts.
- Curse and Might land on the enemies around Shahram: all 863 of Curse's applications and 957 of Might's 961 are on
  creatures and their pets, and the other 4 of Might's on players in one report, all of whom the wearer's side
  damages. 2 of the 4 land on a player the wearer's side also heals and blesses.
- Will lands on the wearer alone, 937 applications. Each copy is removed a median 20,005 ms after its own application,
  2nd to 98th percentile 19,946 to 20,043 ms, with no refresh and no stack count. A copy was applied 144 times with one
  other copy up and 19 times with two.
- Blessing lands on the wearer and up to 4 other players of the wearer's party: 366 applications on the wearer, 1,359
  on other players and 132 on pets. Each pet's copy is removed within about 10 ms (median 10 ms, 98th percentile 35 ms)
  and never ticks. In 35 of the 38 fights whose roster holds more than 5 players, a wearer's Blessings reach no more
  than 4 other players between them. Each player's copy ticks 50 mana and 50 health every 5 s for 20 s, and two copies
  on one player tick side by side, each on its own beat: 102 times a second copy landed while the first was up.
- Fist lands on Shahram himself, 352 times.

## Count of the 2026-11-13 to 2026-11-24 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every report on disk that holds a summon of 16602. The
  rate runs and their fights: 1JLq46yg7wYNhan9 17-28; 2x7ZjvdbRwFMWKTP 3-45; 3X9ychCWZz1vTVxA 9-41; 4VtCK6Qqr1MD2hpc 4;
  ARwPXzLJk78HgVQq 1; CV1DwP4Bct23AzGY 42; JRtbDP4L3BdazHrV 1-2; PC9bgYJBDN8cMnAT 4; WFMjbyKYzvfNkPhH 6-14;
  WdbKG4RkmjxA7YV3 1-3; WwAp7qrmHYNDPzfg 4-9; YAQ1TL6rbFck3JNC 4; aDHTPGq6htJFx8Nw 1-8; bAJFX6m7ftYWCKHP 1-4;
  bR9LQtgmKwZ3r6kW 1-25; n2HkhcYFrLK1tfZj 1-2; nqDL2hKfQB4yvjgJ 1-18; pHdqPxhAYLQTwya7 1-2; qtNCmzfYQwAnRMkZ 1-9;
  tmHdvqxYj1z28PFy 1-9. The picks, every fight with a summon of: 1JLq46yg7wYNhan9, 2q1MandA4HXxL7Ym,
  2x7ZjvdbRwFMWKTP, 37YWanvb6KqcxPh1, 3X9ychCWZz1vTVxA, 4f2ygdYCzFAvqc6K, 4MDFjtr2ngHbWK7d, 4VtCK6Qqr1MD2hpc,
  67fqCHN9ZManpxrk, 6xygv7QnbRjLXGVf, 74h9PQwZNdxGpbCk, 9HTwnpy1qjJ4rDVx, 9jxRWGhpcFbZ14A6, A6qcQ9jtTNXKbvYh,
  aDHTPGq6htJFx8Nw, ARwPXzLJk78HgVQq, bAJFX6m7ftYWCKHP, bR9LQtgmKwZ3r6kW, cMPjHwfyp3zLtJ6C, CqPk8KgFWYTMnw3D,
  CV1DwP4Bct23AzGY, D463FZBX1C9maJ7b, daV6M4rtT7bcD1vm, Gjd4M6kTqVPXY7cL, gpXqDYBNbxKc1y7W, HqwZG7rfy6dg9RL2,
  jAqXz8PN92ZbG3nc, JRtbDP4L3BdazHrV, KYRHMqa28bVmk3QC, LYw76XG1CFHQNdJz, m4dCL16tqnN2D3hW, MazXZJYPrAkG7vWm,
  MDk4wyGC1dbrY8cH, n2HkhcYFrLK1tfZj, Nm6gT8bzpRGHJqfA, nqDL2hKfQB4yvjgJ, nW8CDGVNP3xhtAa6, NXMZV9xjvWpB7GtP,
  PafbczYWBMjgHCG7, PC9bgYJBDN8cMnAT, pHdqPxhAYLQTwya7, pHxtWjvhNAGq4Dc6, qtNCmzfYQwAnRMkZ, qzRy8Y1pWbwdBAkg,
  RWBahpJkAgVPGNFx, t3g6TX91AGmz8yhr, tmHdvqxYj1z28PFy, TQHpwRCAYtLXZKcG, VCJMzXad3Q6vbjRk, vH9PnJDV81wrthRj,
  vwxP1TaLzG8A4m7h, w1T3Z2hQHpkPcAxY, WdbKG4RkmjxA7YV3, WFMjbyKYzvfNkPhH, WwAp7qrmHYNDPzfg, xhzFqGmY3P4Nr2QR,
  YAQ1TL6rbFck3JNC, yBHqgjMPA2hGzCpx, YLdrGB6ygJ3aqxkD, Z6Xm92K3zgfrqM84, z9dafy3kFWZK4Gb8, ZaJPCDqQYvt37TAr and
  zhBG1Z4HA7Lm63FR
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC). Of the 63 pick reports 9 are on 70938, 13 on 70959 and 41 on 71037, and of the 20 rate reports 4, 4 and 12; 39
  pick reports and 11 rate reports start after hotfix push 114313 (2026-11-20 23:47 UTC). The sim's database is
  1.60.1.71037, push 114366; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-24
- Pull request: #1195

The rate is EV-0153 §Method run on Blackblade of Shahram over every wearer stream on disk, once with summons within 50
ms of each other as one proc and once with each summon as a proc of its own; the picks and their targets are read from
the same reports, matched to each Shahram by his instance. Reports VtjvpP8bwcgCYTXF and NdTXJzkWHLZrVQYm log the same
run as 1JLq46yg7wYNhan9, every summon of theirs matching one of its own about 3 s apart, and count once: both leave the
picks, and NdTXJzkWHLZrVQYm, the one among the rate runs, leaves the rate. Report YLdrGB6ygJ3aqxkD logs each event of
its wearers about seven times, and each counts once. Scripts and outputs: `requests/forever-shahram-1124` in the log
parser's folder.
