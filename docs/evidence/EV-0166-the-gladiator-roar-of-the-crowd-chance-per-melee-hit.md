# EV-0166: The Gladiator's Roar of the Crowd is a chance per landed melee hit

- Claim: Roar of the Crowd (The Gladiator 4-piece bonus, 1301123) procs at about 7.2% per landed melee hit, white
  swings, Windfury strikes and Stormstrike alike, a chance per hit and not 5 PPM. The row's base of 5 is the haste the
  tooltip prints (`$s1%`) and may also be the chance: 7.2% sits between a flat 5% and the 10% the tooltip gives with no
  ally near, and this log cannot separate the two. One wearer gives the rate, so it is observed.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000): aNh3Dt8kbZHL9cBM fight 1 (the one indexed wearer with
  4 or more pieces, wearing 5); xcGfXYvTD6bpg3qt fights 1 to 18, a second shaman whose report carries no gear, as a
  check only
- Game build: 1.60.1.70959 or 1.60.1.71037 by report date; the hotfix the player ran is not recorded
- Date: 2026-11-18
- Pull request: #1146

## Method

Only events the wearer is the source of count (totems left out). The wearer is an Enhancement shaman with Dreadforge
Retaliator, a 3.7 s two-hander, and the set worn in every fight. A proc is an applybuff or refreshbuff of 1301124 on
the wearer. The buff logs 15 to 50 ms after the hit that rolled it, so a hit gets credit for a proc 20 ms before to 120
ms after it.

- All hits: every landed white swing, Windfury Weapon strike, Windfury Totem strike, Stormstrike and Cleave. Row mask
  20 hears melee swings and melee abilities. Lone hits show both kinds proccing: white swings, Windfury strikes and
  Stormstrike each start a proc with no other hit near it.
- Lone white swings: landed white swings with no other melee event of the wearer within 200 ms. These count no
  roll table, but the sample is small.
- PPM is procs over the summed weapon speed of the hits over 60, for comparison only.

The tooltip says attacks are 2 times as likely to proc with no ally within 20 yards. The log does not record ally
positions, and both reports are group dungeon runs, so the two cases cannot be told apart. The rate is the
with-allies chance, or a mix of both.

## Rate

| Count                                | Procs / hits | Chance [95%, Poisson] | PPM at 3.7 s |
| ------------------------------------ | -----------: | --------------------- | -----------: |
| All hits, aNh3Dt8kbZHL9cBM           |     40 / 557 | 7.18% [5.13, 9.78]    |         1.16 |
| Lone white swings, aNh3Dt8kbZHL9cBM  |     14 / 141 | 9.93% [5.43, 16.66]   |         1.61 |
| All hits, xcGfXYvTD6bpg3qt (no gear) |       5 / 88 | 5.68% [1.84, 13.26]   |            - |

5 PPM on a 3.7 s weapon would be 30.8% per hit, far outside every interval. A flat 5% sits just below the all-hits
interval and a flat 10% just above it; a 5% chance doubled on some of the hits would read in between. Three pairs of procs log at the same millisecond on two Windfury strikes; they count as two procs.
