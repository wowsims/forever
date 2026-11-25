# EV-0196: Flat Physical damage done lands before armor

- Claim: Fury of the Bogling 5665 (Bogling Root, item 5206; aura 13, +1 Physical damage done) adds its 1 to the
  holder's white swings and Auto Shots before the target's armor, inside the hit's unmitigatedAmount, and to weapon
  specials the same way. Thunder Clap, a spell with flat base damage, takes none of it. A crit is consistent with the 1
  doubled; one warrior's Rend ticks show no 1 a tick.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the 50 indexed reports whose abilities list 5665; the
  holders with hits while it is on, in the reports §Reports lists
- Game build: 1.60.1.70862 to 1.60.1.71091 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-11 to 2026-11-25
- Pull request: #TBD

## Method

Per holder, ability and fight, with the holder's other buffs the same, the hits with the buff on are set against those
with it off. A hit counts as on only when both the Buffs events of 5665 and the event's `buffs` string say so, and as
off only when both say off (the string drops every aura for part of one report); 763 of 5,983 events are left out for
it. The holder's pet sits in the same fetch and is left out of every count but the control. Normal hits only (no
absorb, block or overkill), none within 50 ms of the buff's own events.

Two reads separate before and after armor. The residual amount - k x unmitigatedAmount, with k the median share
through armor off the buff for the same holder and target game id, stays at 0 when the 1 sits in unmitigatedAmount
and reads 1 when it lands after armor. The shift of unmitigatedAmount itself, on minus off, reads 1 when the 1 lands
before armor.

## Results

| Test                                             | Off / on  | Result                                                     |
| ------------------------------------------------ | --------- | ---------------------------------------------------------- |
| White swing residual                             | 609 / 224 | 0.027 off, 0.028 on                                        |
| Auto Shot residual                               | 524 / 154 | -0.05 off, -0.10 on                                        |
| Weapon special residual                          | 392 / 178 | -0.29 off, -0.31 on                                        |
| White swing unmitigatedAmount shift, 9 groups    | 422 / 354 | +1.04 +- 0.18; each group's minimum and 10th percentile +1 |
| Auto Shot unmitigatedAmount shift, 7 groups      | 501 / 456 | +1.02 +- 0.21                                              |
| White swings and Auto Shots, 16 groups           | 864 / 754 | +1.10 +- 0.14                                              |
| Weapon specials, 12 groups                       | 285 / 226 | +0.79 +- 0.14                                              |
| Wing Clip, one hunter                            | 77 / 55   | 3 and 5 off, 4 and 6 on                                    |
| Thunder Clap, 6 groups                           | 84 / 89   | +0.28 +- 0.11; 18 and 20 on and off, never 19              |
| Crit residual (amount - 1.998 x k x unmitigated) | 72 / 40   | 0.105 off, 0.29 on                                         |
| Crit unmitigatedAmount shift                     | 43 / 44   | +0.26 +- 0.51                                              |
| Rend ticks, one warrior                          | 19 / 22   | 10.89 off, 11.14 on (+0.25 +- 0.16)                        |
| Pet swings split by the owner's buff (control)   | 8 groups  | -0.33 +- 0.16                                              |
| Arcane Shot, Thorns, Retribution Aura (control)  | 495 / 444 | -0.19 +- 0.15 and +0.06 +- 0.04                            |

The residual holds at 0 on whites, Auto Shots and specials, so nothing lands after armor, and the minimum white swing
moves by exactly 1 before armor. Among the specials, Maul, Revenge and Claw read +1.36, +1.22 and +0.92, Raptor Strike
+0.47 and Multi-Shot -0.60, each +- 0.56 to 0.64. A crit residual that stays at twice k x unmitigatedAmount rules out a
1 added after the crit doubling and before armor (+0.7, about 3 sigma); the crit shift itself cannot tell a doubled 1
from half of it. A 1 spread over a Rend's ticks (about 0.14 a tick) cannot be told from none.

## Reports

The reports with counted hits: 6zxcNCPR2fvh4YDV, AGWNMHFKkbZ1q7gX, FhfTQGMzrKLbVPBg, JcQhKD97NWxL2B3r,
KGrXTJq98zmcdhBW, NBZdqrT9FWa4MYzL, W26nJZxcYmvVFkQp, ah4jpXRfKY7FrPdG, dDht8R2MfLXpxZkj, g34kZxhKDz1TYBpj,
kMpXbdzAyJm3VW46, kgpXBdMVvtFh98Qj, tpBKMqQmcyJ3DCY6, ybwV1azLfYMqHBDn, ycZLWXxQ4jk21w3q
