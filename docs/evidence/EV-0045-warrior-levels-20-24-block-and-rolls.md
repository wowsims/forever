# EV-0045: Warrior threat, block value and physical rolls, levels 20 to 24

- Claim: Cleave Rank 1 adds 30 threat beyond its damage, Master of Defense's rage makes 5 threat a rage in either
  stance, a partial block takes the sheet's block value floored, the flat physical bonus follows the row's coefficient,
  and a physical hit rolls its base points by half the Variance onto a whole number.
- Sources: two raw combat logs, a ThreatLog and a StatLedger of a warrior at levels 20 to 24; a roll width check across
  every raw Forever log
- Game build: not recorded
- Date: 2026-11-07
- Pull request: #906

## Threat

From 3,126 ThreatLog updates on 370 mobs:

| Reading                   | Measured                                                                                                        |
| ------------------------- | --------------------------------------------------------------------------------------------------------------- |
| Victory Rush's heal       | 0.5 x the health restored x the stance, split between the mobs in the fight: 19 of 21 windows within ±1 of that |
| Master of Defense rage    | clean Defensive windows read 4.99 [4.92, 5.05] (n 12), single source windows 25 and never 45                    |
| Cleave Rank 1 flat threat | 30.0 on one target (n 2), 14.95 per target on two (n 18)                                                        |

With Victory Rush's heal at 0 the median residual is +29.8. For Master of Defense, the unscaled model fits all 35
fights within 5%.

## Damage

These readings come from the combat logs and the sheet. Every partial block (540 of 540) takes exactly the sheet's
block value, floor(shield + 0.05 x Strength - 1); a 6 shield at Strength 54 to 59 blocks 7, where rounding gives 8.

With Elixir of Minor Force (+1), Shield Bash Rank 1 (60/60), Hamstring (30/30) and Revenge take the flat physical
damage bonus; Thunder Clap (31/31), Victory Rush (15/15) and Rend ticks do not. Their rows state coefficient 1 and 0.

Revenge Rank 1's 447 hits (base 22, Variance 0.154) fit a physical roll of round(uniform 22 ± 1.69), landing on 20 to
24; 19 to 25 is rejected (19: 3 against 60.7, 25: 0 against 60.7).

| Hit      | 20   | 21    | 22    | 23    | 24   |
| -------- | ---- | ----- | ----- | ----- | ---- |
| Counted  | 27   | 133   | 126   | 116   | 23   |
| Expected | 24.1 | 125.6 | 125.6 | 125.6 | 24.1 |

Across every raw log, hunter pet Claw and Bite show exactly 2h + 1 values in all 35 clean groups
(3,312 hits), never 4h + 1. Eviscerate's 30 groups all stay inside the half width, and the full width loses by 131.6 log
likelihood. Ferocious Bite and Shield Slam have too little data.

## Not measured

Whether the block floor applies to the Strength part alone or to the whole is not measured: the two differ only under a
block value multiplier. The logs also leave unmeasured whether healing modifiers apply to Victory Rush's heal, whether Heroic
Strike and Cleave take the flat bonus (inconclusive), Defiance ranks 2 and 3, and Master of Defense without a shield and on spell avoids (one parried Ignite gave two procs). The non-warrior
spells that lose the flat bonus under the rule (Rake's opening hit, Summon Hawk, Hanzo Sword's Wound, Everlook
Pathcarver's on-use) were not measured themselves. Shield Bash Rank 3 (row 1672) states 0, so it no longer takes the
bonus; only Rank 1 was measured. The pull request closed #907 with the physical roll width findings and the raw log
excerpts.
