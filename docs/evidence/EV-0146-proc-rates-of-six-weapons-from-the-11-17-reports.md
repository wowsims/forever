# EV-0146: Proc rates of six weapons from the 2026-11-17 Warcraft Logs reports

- Claim: Killmaim procs at 1 and Waveslicer at 1.95 procs per minute; Julie's Dagger at about 1.05, Axe of the Deep
  Woods at about 1.7, Barman Shanker at about 2.05 and Blunderflame Bone Bow at about 2.1, too few wearers or too wide an
  interval to call final; each priced on each landed hit's weapon base speed.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section's table names; 66
  reports in all
- Game build: 1.60.1.70938 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-17
- Pull request: #PRNUM

## Method

EV-0136's method on its reports and 1,003 more listed up to 2026-11-17: a proc is an event of the item's proc spell,
counted once when it logs several events within 50 ms; the hits are the wearer's landed, non-periodic hits of the item's
hand (white swings, melee-defense abilities, seals and Windfury Totem strikes on the main hand; white swings on the off
hand; Auto Shot and ranged-defense hits on the ranged weapon); a dual wielder's swings are split between the hands by
swing timing; a run logged in two reports counts once and a stretch where the item was swapped away is dropped. The
interval is the wider of the per-hit Wilson interval and 2,000 bootstrap draws of whole wearers. A rate is measured when
5 or more wearers give it and its interval lies within 10% of it either side, and observed otherwise.

Julie's Dagger leaves out its one warrior run, since a warrior's off-hand Whirlwind rolls the dagger and is not counted
as a hit here (EV-0136); with it the rate reads 1.10. Axe of the Deep Woods reads 1.94 on its rogue and hunter runs and
1.50 on its warrior runs. Barman Shanker has only rogue wearers and Blunderflame Bone Bow only hunters.

## Killmaim (item 13016, spell 13318): 1, measured

700 procs on 12,890 landed hits, 32 runs: 1.02 [0.94, 1.10].

| Report           | Fights    | Class   | Hand      | Procs |  Hits | Rate |
| ---------------- | --------- | ------- | --------- | ----: | ----: | ---: |
| 9b1wvaKGpHdNJCZT | 2-4       | Paladin | main hand |    62 | 1,146 | 1.01 |
| pDRxq6wJynmhgdjL | 1-9       | Warrior | main hand |    57 |   944 | 1.13 |
| MrW6JCFZTDzdtYpH | 1-4       | Warrior | main hand |    40 |   910 | 0.82 |
| QhYmjJ4bHZXcgfk1 | 1-2, 4-27 | Shaman  | main hand |    42 |   889 | 0.89 |
| bdL8j9KGMR6r2xmk | 1-3       | Warrior | main hand |    41 |   736 | 1.04 |
| nyhMJLqBYxPpRaXf | 1-2       | Warrior | main hand |    44 |   572 | 1.44 |
| P16dx3VLmRyaDJZQ | 1-18      | Shaman  | main hand |    31 |   561 | 1.04 |
| G3VKdgR2aLAkNqWy | 1-3       | Shaman  | main hand |    34 |   556 | 1.15 |
| d7DFtXL2jfNa8Bw4 | 1-5       | Shaman  | main hand |    28 |   529 | 0.99 |
| ghzjGKVBvJbRZnMr | 1-2       | Shaman  | main hand |    34 |   458 | 1.39 |
| 3xgPkwfn8yQ9BqAW | 1-4       | Warrior | main hand |    29 |   449 | 1.21 |
| pqr4HxC3vWJhfG7B | 13-20     | Shaman  | main hand |    23 |   412 | 1.05 |
| hbW1zgMXyRv8JAVL | 1-6       | Shaman  | main hand |    28 |   403 | 1.30 |
| acHy1fNktP2m47pd | 1-2       | Shaman  | main hand |    15 |   398 | 0.71 |
| 8RFv6Xyr43dLpJ7h | 1-2       | Warrior | main hand |    24 |   371 | 1.21 |
| 2aVnTxFzfckdXpBy | 1-7       | Warrior | main hand |    17 |   351 | 0.91 |
| tMxrNWC6qPVzgwAv | 7-8       | Shaman  | main hand |    20 |   340 | 1.10 |
| zJ6C2148dTYfaXHb | 2-4       | Shaman  | main hand |    10 |   315 | 0.60 |
| jKJPF16ZnWLkrR2d | 1-2       | Shaman  | main hand |    17 |   315 | 1.01 |
| P7cV2HjLgxGpmZQk | 1-7       | Shaman  | main hand |    14 |   313 | 0.84 |
| gdKNPjVvcQt2LJnq | 1-3       | Shaman  | main hand |    15 |   304 | 0.93 |
| aQwNVCPHWmzDLXqk | 1-2       | Warrior | main hand |    16 |   221 | 1.36 |
| 3vDf6TMYBX9LK7bz | 1-2       | Shaman  | main hand |     6 |   216 | 0.52 |
| zhj69Dt2X1b3fvMT | 1         | Shaman  | main hand |     8 |   172 | 0.87 |
| grFA3BHxtCNyzR7v | 1-2       | Shaman  | main hand |    13 |   170 | 1.43 |
| fn2Dc941K63CNvmj | 3-4       | Warrior | main hand |     3 |   170 | 0.33 |
| QqKcMJwHjxNPkgRT | 1-3       | Shaman  | main hand |     7 |   152 | 0.86 |
| 7hNgRwMfkQZ2Yv4b | 1-3       | Hunter  | main hand |     5 |   138 | 0.68 |
| vJGT19YL26XFQpNB | 1-2       | Shaman  | main hand |     6 |   129 | 0.87 |
| 29A3dKwPgpRNnWvc | 7-11      | Shaman  | main hand |     6 |   125 | 0.90 |
| 8GW16QY9BmFHJXLb | 1         | Hunter  | main hand |     1 |    76 | 0.25 |
| 6v4wTWQkjb3Xd7mM | 6         | Warrior | main hand |     4 |    49 | 1.53 |

## Waveslicer (item 18324, spell 1302166): 1.95, measured

1,688 procs on 16,901 landed hits, 7 runs: 1.93 [1.84, 2.04].

| Report           | Fights | Class   | Hand      | Procs |  Hits | Rate |
| ---------------- | ------ | ------- | --------- | ----: | ----: | ---: |
| VYBjgJx8aZtRhmpy | 1-43   | Shaman  | main hand |   691 | 6,951 | 1.92 |
| AyTmpNtxB9jbnJ2H | 4-44   | Shaman  | main hand |   318 | 3,245 | 1.90 |
| q23Tz8hnZGkvbLx6 | 1-18   | Shaman  | main hand |   200 | 1,887 | 2.05 |
| NrpM7DJQBKWqbzVF | 1-7    | Paladin | main hand |   164 | 1,784 | 1.78 |
| DpVmH4WRxL72nGwd | 3      | Shaman  | main hand |   127 | 1,275 | 1.93 |
| H2Zg1jRm93PpTqK4 | 1-3    | Shaman  | main hand |   135 | 1,225 | 2.13 |
| KCYXdTfNgGHczbLw | 5-6    | Shaman  | main hand |    53 |   534 | 1.92 |

## Julie's Dagger (item 6660, spell 8348): 1.05, observed

218 procs on 9,524 landed hits, 9 runs, the warrior run left out (its off-hand Whirlwind rolls the dagger): 1.06 [0.92, 1.20].

| Report           | Fights  | Class              | Hand     | Procs |  Hits | Rate |
| ---------------- | ------- | ------------------ | -------- | ----: | ----: | ---: |
| JMnDtLGxf4FqCV6w | 1-3     | Rogue              | off hand |    42 | 2,252 | 0.86 |
| nyAx3XNDVmYWGdrh | 1-6     | Rogue              | off hand |    40 | 1,672 | 1.10 |
| 4nNqbfmrCT2Q9PpH | 1-2     | Rogue              | off hand |    29 | 1,345 | 0.99 |
| N3mrMbX8TJyz1pxq | 1-4     | Warrior (left out) | off hand |    38 | 1,212 | 1.45 |
| nCDZQNWKX3twHhyY | 1-8, 10 | Rogue              | off hand |    28 | 1,089 | 1.19 |
| 7qPHR6cLn3zGptvN | 1-3     | Rogue              | off hand |    26 | 1,052 | 1.14 |
| ApJhHfRkMDqFj3rb | 4-5     | Rogue              | off hand |    22 |   929 | 1.09 |
| dJa2G7PHZg9wt6QX | 1-4     | Rogue              | off hand |    11 |   604 | 0.84 |
| DmjBrYd7M9VZwCHF | 22      | Rogue              | off hand |    19 |   520 | 1.69 |
| 6kN78G2xfqy3vwTP | 6       | Rogue              | off hand |     1 |    57 | 0.80 |

## Axe of the Deep Woods (item 811, spell 18104): 1.7, observed

867 procs on 11,417 landed hits, 7 runs: 1.69 [1.55, 1.98].

| Report           | Fights                 | Class   | Hand      | Procs |  Hits | Rate |
| ---------------- | ---------------------- | ------- | --------- | ----: | ----: | ---: |
| pFH4amZPxkBhyTtV | 1-2, 4, 6-34           | Warrior | main hand |   246 | 3,635 | 1.50 |
| NB4ag2CcKryDT91x | 1-5                    | Warrior | main hand |   170 | 2,499 | 1.51 |
| zVAHZcQtTdbhMfLN | 7-8                    | Rogue   | main hand |   155 | 1,697 | 2.03 |
| xqnmW3BDpFb28cJ1 | 1-44, 46, 48-50, 52-53 | Hunter  | main hand |   133 | 1,653 | 1.79 |
| vZqVmLYQa6jwP3A9 | 2-16, 18-21            | Rogue   | main hand |    70 |   749 | 2.07 |
| 4ChPnBrVK2fJLMaX | 1-2, 4                 | Warrior | main hand |    39 |   592 | 1.46 |
| 9kb8p2jZ4YXmGQCc | 1                      | Warrior | main hand |    54 |   590 | 2.03 |

## Barman Shanker (item 12791, spell 1300787): 2.05, observed

681 procs on 9,923 landed hits, 5 runs: 2.06 [1.88, 2.23].

| Report           | Fights    | Class | Hand      | Procs |  Hits | Rate |
| ---------------- | --------- | ----- | --------- | ----: | ----: | ---: |
| GbNrpkt9fVKv74wQ | 1-2, 4-33 | Rogue | main hand |   185 | 2,758 | 2.01 |
| 9zxHaWbYT3gqDQvF | 12-18     | Rogue | main hand |   176 | 2,407 | 2.19 |
| zp9yVrtfTNZJCv7g | 5-7, 9-10 | Rogue | main hand |   143 | 2,010 | 2.13 |
| p6mVRBzvWMqnQZFc | 2-12      | Rogue | main hand |   108 | 1,902 | 1.70 |
| DmjBrYd7M9VZwCHF | 22        | Rogue | main hand |    69 |   844 | 2.45 |

## Blunderflame Bone Bow (item 275439, spell 1296730): 2.1, observed

376 procs on 3,874 landed hits, 6 runs: 2.08 [1.87, 2.37].

| Report           | Fights   | Class  | Hand   | Procs |  Hits | Rate |
| ---------------- | -------- | ------ | ------ | ----: | ----: | ---: |
| cL3KMWpnXyvDrqa2 | 1-38     | Hunter | ranged |   195 | 2,008 | 2.08 |
| HqYP86CdQZtkB2Gf | 5-10     | Hunter | ranged |   120 | 1,136 | 2.26 |
| x9v27CNzhRgMkXG6 | 9, 11-12 | Hunter | ranged |    23 |   290 | 1.70 |
| dMxyPfarzTc3tLbk | 6        | Hunter | ranged |     8 |   193 | 0.89 |
| QRCgnhck3KVq6vBT | 1-2      | Hunter | ranged |    15 |   133 | 2.42 |
| VK83G7MdBhLRFAfa | 1-42     | Hunter | ranged |    15 |   114 | 2.82 |
