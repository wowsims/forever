# EV-0213: Sweeping Strikes' normalized strike rolls a crit of its own and is never avoided

- Claim: The normalized strike Sweeping Strikes brings from a Whirlwind (26654) is never missed, dodged or parried and
  rolls a crit of its own, at about the warrior's melee crit chance, whether or not the main-hand Whirlwind strike that
  brought it crit.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the warrior streams on disk that log both a dual-wield
  Whirlwind (the off-hand strike 462891) and a Sweeping Strikes hit (26654 or 462890), 21 streams of 20 warriors in the
  21 reports under Reports, all their fights
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-24
- Pull request: not yet opened

## Method

Each Whirlwind cast (1680) of these warriors, in two-handed fights and dual-wield fights alike, with the warrior's own
events in the 2 ms on either side of the cast. Every 26654 event in that window counts, whatever its hit type, so a
miss, dodge or parry would show as one. The rate beside it is the crit share of the landed main-hand Whirlwind strikes
of the same casts.

## Numbers

| Report build | 26654 strikes | Crits | Crit share | The casts' landed main-hand strikes |
| ------------ | ------------- | ----- | ---------- | ----------------------------------- |
| 70938        | 201           | 33    | 16.4%      | 76 of 518, 14.7%                    |
| 70959        | 249           | 46    | 18.5%      | 183 of 787, 23.3%                   |
| 71037        | 104           | 31    | 29.8%      | 66 of 307, 21.5%                    |
| All          | 554           | 110   | 19.9%      | 325 of 1,612, 20.2%                 |

The 554 strikes log 444 hits and 110 crits: no miss, dodge, parry, block or glance.

| The cast's main-hand Whirlwind strikes | 26654 strikes | Crits       |
| -------------------------------------- | ------------- | ----------- |
| At least one crit                      | 242           | 51, 21.1%   |
| None crit                              | 312           | 59, 18.9%   |

The strike's crit does not follow the strike that brought it, unlike Sweeping Strikes' copied hit 12723, which carries
its hit's crit and never crits itself (EV-0182). Over every Whirlwind cast of these warriors, 1,413 of 7,285 landed
main-hand strikes crit (19.4%).

The strike Sweeping Strikes brings from an Execute below 20% health is not in these logs.

## Reports

16GhDNRmVn8KWjBT, 6q1mWrtM7zXFckvp, 782JNDc1pgGRZxVC, 8Zp12yQdxz67MaDL, AcP9Y3qbw2G7fxgn, bMF6ZAPf7BQLakgn,
cbkDQfXNPGj3tnWT, cZ9B6Ay3RDTWgq4K, DJ436qAXGnBHLdTa, DQ62GWmAajLh4zbN, gDtFwpGPLM2TyZKB, gKdFBvXjnTfDLHzx,
GMrxRanAp9fJ46bQ, gvWZG1cjrChKwdtz, mPMpkAgjbtR7HBXD, MwVH1KQmak73NPd8, NWDCd6xf9AXK17yv, t3g6TX91AGmz8yhr,
tnyh4dxq6pNXQzLw, y6FZRHQnckD8YjAC, ZFXTPc8fz3RHjKM9.
