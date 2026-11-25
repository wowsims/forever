# EV-0198: Sniper Shot's charge goes at the Shot's cast, a miss included

- Claim: Sniper Shot's buff (1310687, 1310785, 1310786) loses one charge at the cast of each Shot its range modifier
  names (Auto Shot, Arcane Shot, Aimed Shot, Multi-Shot, Concussive Shot, Scatter Shot), whether the Shot hits or
  misses, and one for each Multi-Shot however many enemies it hits. Serpent Sting, which the modifier does not name,
  spends none.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the streams on disk of the hunters that gained the buff:
  202 streams in 145 reports, too many to list
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20
  18:23 UTC); the range spans hotfix push 114313 (2026-11-20 23:47 UTC); the hotfix each player ran is not recorded
- Date: 2026-11-14 to 2026-11-24, the last report starting at 11:04 UTC
- Pull request: #1233

## Method

Inside each window of the buff, a Shot's cast event counts the buff's charge removals (a stack removed or the buff
removed) within 5 ms of it. A cast with another Shot cast within 5 ms is left out, as the two share the removals (313
casts; 235 of them saw two removals, one for each Shot). A Shot's result is its first damage or miss event within 1.5 s of the cast, and "no result" where none follows:
Concussive Shot deals no damage, and an enemy can die first. A Multi-Shot's targets are its damage events within
200 ms of its first.

## Counts

Casts that spent one charge, of the casts.

| Shot            | Landed      | Missed | No result |
| --------------- | ----------- | ------ | --------- |
| Auto Shot       | 3,894/3,896 | 25/25  | 195/195   |
| Arcane Shot     | 1,471/1,472 | 52/53  | 50/50     |
| Aimed Shot      | 684/685     | 18/18  | 11/11     |
| Concussive Shot |             | 5/5    | 16/16     |
| Scatter Shot    | 1/1         |        | 1/1       |
| Serpent Sting   |             | 0/33   | 0/426     |

Every Multi-Shot spent exactly one charge: 3 of 3 with no target hit, 121 of 121 with one, 271 of 271 with two and 328
of 328 with three. No Serpent Sting spent one.

By build:

| Build | Named Shots, landed or no result | Named Shots, missed | Multi-Shot | Serpent Sting |
| ----- | -------------------------------: | ------------------: | ---------: | ------------: |
| 70938 |                      1,481/1,482 |               20/20 |    207/207 |         0/123 |
| 70959 |                      1,474/1,476 |               26/27 |    127/127 |          0/95 |
| 71037 |                      3,368/3,369 |               54/54 |    389/389 |         0/241 |
