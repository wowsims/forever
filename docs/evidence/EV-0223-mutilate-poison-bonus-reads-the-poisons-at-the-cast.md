# EV-0223: Mutilate's poison bonus reads the target's poisons at the cast

- Claim: both Mutilate strikes take the poison bonus from the poisons on the target at the cast; a poison that either
  strike's poison proc puts up during the cast raises neither strike. The main-hand strike logs before the off-hand
  strike.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every indexed stream with a Mutilate strike: 253 rogue
  units (one rogue in one report) in 251 streams of 251 reports, too many to list
- Game build: 1.60.1.70938 to 1.60.1.71139 by report date (70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19
  UTC, 71037 from 2026-11-20 18:23 UTC, 71091 from 2026-11-24 15:21 UTC in the US and 2026-11-25 02:19 UTC in the EU,
  71139 from 2026-11-25 19:14 UTC); 166 of the 251 reports start after hotfix push 114313 (2026-11-20 23:47 UTC), and
  the reading is the same on both sides of it; the hotfix each player ran is not recorded
- Date: 2026-11-14 to 2026-11-26 (report starts 2026-11-14 04:17 to 2026-11-26 14:38 UTC; the census reads reports from
  2026-10-14 to 2026-11-27, and the first Mutilate is in a report of 2026-11-14)
- Pull request: #

## Method

A unit is one rogue in one report whose own events hold a Mutilate main-hand strike (1241586) or off-hand strike
(1241590). The poison state on a target comes from the rogue's own aura events of every ability whose name holds
"Poison", keyed by target and instance and ending 20 s after its last event. Instant Poison puts up no aura, so the
state holds the debuff poisons, Deadly Poison and Wound Poison among them.

Each strike is grouped by the state before the cast's first strike, the off-hand or main-hand strike on the same target
that logs first within 50 ms:

- poisoned before the cast: a poison up before the first strike, and none put up within 50 ms before it;
- clean, a poison put up within 50 ms before the cast's first strike: the poison went up just before the cast, so it is
  up at the cast;
- clean, a poison put up and logged after the strike, within 50 ms: the cast's own poison procs;
- clean, no poison put up within 50 ms.

Only normal hits count. Each one is divided by the unit's median normal hit of the same strike in the last group, for
units with 5 or more of them (252 units for the main hand). A median at 1.000 reads no bonus, and one near 1.19 reads
the bonus. Weapon damage spreads each group, so the share of hits at 1.1 or above stands beside the median: a group
where some hits take the bonus shows a larger share than the clean group.

Scripts and outputs: `requests/forever-aura-order-1127` in the log parser's folder, `specials.py` with
`specials_out.txt` for the main hand and the log order, `mut_hands.py` with `mut_hands_out.txt` for both strikes and
the split by hotfix push 114313, and `mut_shank.py` with `mut_shank_out.txt` for the units with off-hand Instant Poison.

## Log order

Of the 89,303 main-hand strike events, the main hand logs first in 88,676 casts, the off hand first in 624 (622 before
hotfix push 114313, 2 after) and no off-hand strike logs within 50 ms in 3. The poison a cast puts up logs after both
strikes: of the 6,018 off-hand normal hits whose cast put a poison up, it logs between the two strikes in 1.

## Readings

Normal hits over the unit's clean median, all reports:

| Strike    | State at the cast                                   |      n | Median | Share at 1.1 or above |
| --------- | --------------------------------------------------- | -----: | -----: | --------------------: |
| Main hand | clean, no poison put up                             | 21,200 |  1.000 |                 0.203 |
| Main hand | clean, a poison put up after the strike             |  5,949 |  1.000 |                 0.204 |
| Main hand | clean, a poison put up within 50 ms before the cast |    184 |  1.179 |                 0.766 |
| Main hand | poisoned before the cast                            | 29,802 |  1.188 |                 0.774 |
| Off hand  | clean, no poison put up                             | 20,548 |  1.000 |                 0.195 |
| Off hand  | clean, a poison put up after the strike             |  6,017 |  1.000 |                 0.196 |
| Off hand  | clean, a poison put up within 50 ms before the cast |    205 |  1.190 |                 0.771 |
| Off hand  | poisoned before the cast                            | 29,728 |  1.188 |                 0.807 |

On both sides of hotfix push 114313 (reports starting before 2026-11-20 23:47 UTC: 85 reports, 87 units; after it: 166
reports, 166 units):

| Strike    | State at the cast                       | Before: n | Median | After: n | Median |
| --------- | --------------------------------------- | --------: | -----: | -------: | -----: |
| Main hand | clean, no poison put up                 |     7,770 |  1.000 |   13,430 |  1.000 |
| Main hand | clean, a poison put up after the strike |     2,692 |  1.000 |    3,257 |  1.005 |
| Main hand | poisoned before the cast                |    11,346 |  1.185 |   18,456 |  1.189 |
| Off hand  | clean, no poison put up                 |     7,335 |  1.000 |   13,213 |  1.000 |
| Off hand  | clean, a poison put up after the strike |     2,750 |  1.000 |    3,267 |  1.000 |
| Off hand  | poisoned before the cast                |    11,435 |  1.186 |   18,293 |  1.188 |

A poison the cast puts up leaves the strike's median at 1.000 and its share at 1.1 or above at the clean group's, on
both strikes. The main hand logs first, so it is the off hand that would see a poison the main-hand strike put up. The
log does not say which hand carries which poison, but Shank strikes with the off hand alone (EV-0155): a unit whose
Instant Poison follows 10% or more of its 30 or more Shanks within 50 ms carries Instant Poison on the off hand, and its
debuff poison on the main hand. The rate falls into two groups, 0.00 to 0.08 and 0.13 to 0.61, and 60 counted units are
in the upper one. On these 60, a poison put up in the cast comes from the main-hand strike, and the off hand still
reads no bonus:

| Strike    | State at the cast                       |     n | Median | Share at 1.1 or above |
| --------- | --------------------------------------- | ----: | -----: | --------------------: |
| Main hand | clean, no poison put up                 | 6,919 |  1.000 |                 0.214 |
| Main hand | clean, a poison put up after the strike | 1,010 |  1.000 |                 0.196 |
| Off hand  | clean, no poison put up                 | 6,785 |  1.000 |                 0.204 |
| Off hand  | clean, a poison put up after the strike | 1,027 |  1.000 |                 0.199 |
