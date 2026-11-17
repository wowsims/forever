# EV-0139: Deep Wounds ticks from the average damage of the weapon that crit

- Claim: an off-hand crit starts Deep Wounds, and the bleed is sized from the off hand's average weapon damage without
  the off-hand damage penalty; main-hand swings, main-hand abilities and Windfury Totem's strike size it from the main
  hand. Attack Power does not count.
- Sources: combat logs. A level 60 dual-wield Fury warrior in an Upper Blackrock Spire group with Windfury Totem
  (2026-11-16); a level 30 dual-wield warrior in Duskwood (2026-10-13); bystander warriors in the logger's own logs
  2026-11-08 to 2026-11-16.
- Game build: 1.60.1 as the log header states it; the build number and hotfix were not recorded.
- Date: 2026-10-13 and 2026-11-08 to 2026-11-16
- Pull request: #1077

`SWING_DAMAGE` has no hand field, so a swing's hand comes from its base damage: a dual wielder's swings fall in two
clusters, and the `SWING_MISSED` lines, which do carry the off-hand flag, confirm the warrior dual-wields. A crit counts
when its target had no Deep Wound from that warrior, since a refresh logs no aura event, and when the crit did not kill.
A crit's target has to be in log range: a swing whose `SWING_DAMAGE_LANDED` twin is missing is left out, because the
log drops the target's aura events with it.

## Which crits start the bleed

`SPELL_AURA_APPLIED` 412609 within 150 ms of the crit, on targets without the bleed:

| Crit                  | Level 60 warrior, 2026-11-16 | Other dual wielders, 2026-11-12 to 11-14 |
| --------------------- | ---------------------------- | ---------------------------------------- |
| Off-hand swing        | 17/17                        | 4/4                                      |
| Main-hand swing       | 16/16                        | 4/4                                      |
| Windfury Totem strike | 7/7                          |                                          |
| Whirlwind             | 76/76                        | 12/12                                    |

Off-hand swing crits with no other crit by the same warrior on that target within 150 ms: 16/16 and 5/5 for the
strike. Bystander warriors' one-hand and two-hand swing crits: 183/183.

## How large the ticks are

The first tick of each application that no later crit refreshed before that tick. The level 60 warrior's weapons, from
the log's `COMBATANT_INFO`: main hand 80–149 (average 114.5), off hand 73–136 (average 104.5). Deep Wounds 3/3 is 60%
of the average over 4 ticks; 17.2 logs as 17 and 15.7 as 15. Whirlwind's strikes split by base damage, the off-hand
strike's below 300.

| Started by            | n   | Logged first tick   | Main hand's share | Critting weapon's share | Off hand with its 50% penalty |
| --------------------- | --- | ------------------- | ----------------- | ----------------------- | ----------------------------- |
| Main-hand swing       | 10  | 17 (one 18, one 26) | 17.2              | 17.2                    |                               |
| Cleave                | 7   | 17–18               | 17.2              | 17.2                    |                               |
| Windfury Totem strike | 1   | 17                  | 17.2              | 17.2                    |                               |
| Off-hand swing        | 10  | 14–16, median 15    | 17.2              | 15.7                    | 7.8                           |
| Whirlwind, main hand  | 16  | 17 (13 of 16)       | 17.2              | 17.2                    |                               |
| Whirlwind, off hand   | 24  | 15–16 (22 of 24)    | 17.2              | 15.7                    | 7.8                           |

When both of a Whirlwind's strikes crit in one batch the two amounts add up: the first tick reads 32 (6 of 8) or 36,
against 17.2 + 15.7 = 32.9. The main-hand swing that ticked 26 is not explained.

The level 30 warrior (main hand average 46.5, off hand 4.7): main-hand crits tick 6 (n 26), off-hand crits 1 (n 14 of
15). The main-hand rule would give 6 for both. The off hand's share is 0.7 a tick, which logs as 1, so how the game
rounds a tick below 1 is not settled.
