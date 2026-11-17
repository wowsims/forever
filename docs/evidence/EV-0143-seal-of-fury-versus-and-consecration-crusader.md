# EV-0143: Seal and Judgement of Fury take no spell damage against a creature type, and Consecration's marked tick takes Judgement of the Crusader after its done multipliers

- Claim: Seal of Fury's hit and Judgement of Fury take none of an item's spell damage against the target's creature type
  (Courier's Clasp 279833, +10 against Humanoids), and both take Judgement of the Crusader at their coefficient after
  the done multipliers. Consecration's marked tick takes that spell damage at the bonus damage's coefficient, and
  Judgement of the Crusader at the same coefficient outside its base, where no done multiplier scales it; the tick on
  an unmarked enemy takes neither.
- Sources: combat log, a level 60 Protection paladin wearing Courier's Clasp in 17 of 17 combatant info lines, on Scarlet
  Humanoids and Undead in Stratholme, 2026-11-13 14:26 to 16:12 player's local time (22:26 to 00:12 UTC); Warcraft Logs
  reports gBDMLv8J1K7nqkpd, PYQf86cTrZ1GjdKx and 7tLY1pVK9wRWPABX, one clasp-wearing paladin each, every fight in the
  paladin's own events; combat logs, four logs of a Retribution paladin with Vengeance, levels 17 to 30, 2026-11-07 to
  2026-11-09; combat logs, every paladin damage line in 150 local logs from 2026-09-26 to 2026-11-16, for Judgement of
  the Crusader.
- Game build: 1.60.1.70938 for the Stratholme log, before the 2026-11-14 02:00 UTC hotfix; the hotfix it ran on is not
  recorded. The game build and hotfix of the other logs and reports are not recorded.
- Date: 2026-09-26 to 2026-11-16
- Pull request: #TBD

## Creature types

The game files' Creature table holds only companions and pets, so a target's type comes from its NPC tooltip. An
Enhancement shaman in the same Stratholme group, wearing the same clasp, takes the bonus on the Scarlet targets, so the
game treats them as Humanoids. The log's spell power field never holds the bonus: across the paladin's own lines whose
target switched between a Humanoid and an Undead, it stayed the same 40 times and changed once.

## Seal of Fury and Judgement of Fury against the clasp

A Seal of Fury hit fits a reading when floor(35 + 0.1 × (spell power + extra)) equals its base amount, with spell power
read from the paladin's own line at most 10 s before the hit. A hit discriminates when +0 and +10 predict different
amounts. Undead are the control.

| Row                       | Humanoid, fit +10               | Undead, fit +10 |
| ------------------------- | ------------------------------- | --------------- |
| Seal of Fury              | 0 of 437, 0.0% (0 to 0.9%)      | 0 of 591        |
| Consecration, marked tick | 266 of 266, 100% (98.6 to 100%) | 0 of 30         |

Intervals are 95% Wilson intervals. The marked tick's model is floor(39 + 0.042 × spell power), the flat tick and the
bonus together.

Judgement of Fury rolls too wide for exact fits. A linear fit of its base amount on Humanoids minus the other types, at
the same target level and with spell power and attack power held, reads +0.61 ± 1.24 damage against +4.3 expected at 0.429 (one paladin,
277 hits, 130 on Humanoids).

In the Warcraft Logs reports, each non-crit hit with no block, resist or absorb on a Humanoid, off the paladin's own
Judgement of the Crusader targets and while the clasp is worn, is compared with the mean of the same ability's hits on
other types within 60 s:

| Report           | Row               | Hits on Humanoids | Compared | Difference   | +10 expects |
| ---------------- | ----------------- | ----------------: | -------: | ------------ | ----------- |
| gBDMLv8J1K7nqkpd | Seal of Fury      |               922 |      598 | +0.20 ± 0.19 | +1.0        |
| PYQf86cTrZ1GjdKx | Seal of Fury      |                71 |       71 | +0.09 ± 0.16 | +1.0        |
| 7tLY1pVK9wRWPABX | Judgement of Fury |               282 |       48 | −1.20 ± 1.44 | +4.3        |

A hit is compared when the same ability hit another type within the window.

The Judgement of Fury reading rests on one paladin, so it is the weaker of the two.

## Judgement of the Crusader

k is the damage a hit gains per point of Holy damage the target takes extra: its amount minus its base amount, over
Judgement of the Crusader's points on the target, non-crits only. A crit's k over a hit's tells whether a crit scales it.

| Row                       | Hits under the judgement | k              | k in the base | Crit k over hit k |
| ------------------------- | -----------------------: | -------------- | ------------- | ----------------- |
| Seal of Fury              |                    1,108 | 0.102 ± 0.001  | 0             | 2.04              |
| Judgement of Fury         |                      233 | 0.441 ± 0.005  | 0             | 1.95              |
| Consecration, marked tick |   November logs, per log | 0.036 to 0.046 | 0             | 1.47              |
| Consecration, flat tick   |                          | 0.000          | 0             |                   |

The marked tick's range covers 37 of the 43 November rows, one per log and tick rank, from 2026-11-06 to 2026-11-16.
The other six read −0.005 (38 ticks, every one under another paladin's judgement), 0.030 ± 0.008 (32), 0.035 ± 0.001
(175), 0.047 ± 0.006 (7), 0.048 ± 0.007 (16) and 0.065 ± 0.001 (33). A few casters read 1.26 to 2.5 times the row on
every source at once, probably from a talent that raises the judgement on the target; that is not checked.

Consecration's k sits outside the base amount, which holds the caster's spell power, the versus bonus and the done
multipliers. In the Retribution paladin's four logs, Vengeance raises the marked tick's base by ×1.082 and leaves k
where it is: 0.041 ± 0.001 over 61 ticks without it and 0.041 ± 0.001 over 393 with it, where a scaled k would read
0.044. October logs, on an older game build, read the marked tick's k at 0.084 to 0.111.
