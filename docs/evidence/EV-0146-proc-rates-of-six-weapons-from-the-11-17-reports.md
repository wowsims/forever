# EV-0146: Proc rates of three weapons from the 2026-11-17 Warcraft Logs reports

- Claim: Axe of the Deep Woods procs at about 1.7, Barman Shanker at about 2.05 and Blunderflame Bone Bow at about 2.1
  procs per minute, too few wearers, too wide an interval or wearers of one class only to call final; each priced on
  each landed hit's weapon base speed. Barman Shanker's Rend deals twice its damage when applied from behind the target.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section's table names; 66
  reports in all
- Game build: 1.60.1.70938 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-17
- Pull request: #1091

## Method

EV-0136's method on its reports and 1,003 more listed up to 2026-11-17: a proc is an event of the item's proc spell,
counted once when it logs several events within 50 ms; the hits are the wearer's landed, non-periodic hits of the item's
hand (white swings, melee-defense abilities, seals and Windfury Totem strikes on the main hand; white swings on the off
hand; Auto Shot and ranged-defense hits on the ranged weapon); a dual wielder's swings are split between the hands by
swing timing; a run logged in two reports counts once and a stretch where the item was swapped away is dropped. The
interval is the wider of the per-hit Wilson interval and 2,000 bootstrap draws of whole wearers. A rate is measured when
5 or more wearers of more than one class give it and its interval lies within 10% of it either side, and observed
otherwise. The hits a dual wielder's split gives are fractional, so each table's rows are rounded and sum to within a
few hits of its total.

Julie's Dagger leaves out its one warrior run, since a warrior's off-hand Whirlwind rolls the dagger and is not counted
as a hit here (EV-0136). Axe of the Deep Woods reads 1.94 on its rogue and hunter runs and 1.55 on its warrior runs. Barman Shanker has only rogue wearers and Blunderflame Bone Bow only hunters.

## Axe of the Deep Woods (item 811, spell 18104): 1.7, observed

867 procs on 11,417 landed hits, 6 wearers, 7 runs: 1.69 [1.55, 1.98].

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

681 procs on 9,923 landed hits, 5 wearers, 5 runs: 2.06 [1.88, 2.23]. Observed, since every wearer is a rogue.

Its Rend ticks before mitigation fall in two groups, 12 to 14 and 26 to 28: 669 and 761 ticks in GbNrpkt9fVKv74wQ,
9zxHaWbYT3gqDQvF and zp9yVrtfTNZJCv7g. The doubled ticks are the Rends applied from behind the target, as 1300787's
second effect states.

| Report           | Fights    | Class | Hand      | Procs |  Hits | Rate |
| ---------------- | --------- | ----- | --------- | ----: | ----: | ---: |
| GbNrpkt9fVKv74wQ | 1-2, 4-33 | Rogue | main hand |   185 | 2,758 | 2.01 |
| 9zxHaWbYT3gqDQvF | 12-18     | Rogue | main hand |   176 | 2,407 | 2.19 |
| zp9yVrtfTNZJCv7g | 5-7, 9-10 | Rogue | main hand |   143 | 2,010 | 2.13 |
| p6mVRBzvWMqnQZFc | 2-12      | Rogue | main hand |   108 | 1,902 | 1.70 |
| DmjBrYd7M9VZwCHF | 22        | Rogue | main hand |    69 |   844 | 2.45 |

## Blunderflame Bone Bow (item 275439, spell 1296730): 2.1, observed

376 procs on 3,874 landed hits, 5 wearers, 6 runs: 2.08 [1.87, 2.37].

| Report           | Fights   | Class  | Hand   | Procs |  Hits | Rate |
| ---------------- | -------- | ------ | ------ | ----: | ----: | ---: |
| cL3KMWpnXyvDrqa2 | 1-38     | Hunter | ranged |   195 | 2,008 | 2.08 |
| HqYP86CdQZtkB2Gf | 5-10     | Hunter | ranged |   120 | 1,136 | 2.26 |
| x9v27CNzhRgMkXG6 | 9, 11-12 | Hunter | ranged |    23 |   290 | 1.70 |
| dMxyPfarzTc3tLbk | 6        | Hunter | ranged |     8 |   193 | 0.89 |
| QRCgnhck3KVq6vBT | 1-2      | Hunter | ranged |    15 |   133 | 2.42 |
| VK83G7MdBhLRFAfa | 1-42     | Hunter | ranged |    15 |   114 | 2.82 |
