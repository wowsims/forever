# EV-0056: Bloodthrill procs from main-hand white swings, Heroic Strike and Cleave only

- Claim: Bloodthrill opens Overpower on a main-hand white swing and on the on-next-swing Heroic Strike and Cleave
  against a target with the warrior's Rend, and on no other special.
- Sources: Forever combat logs: every raw log with bystander warriors (24 gains), and one dual-wield warrior (main hand
  69 to 89 every 2.6 s, off hand 19 to 24 every 1.6 s); report codes not recorded
- Game build: not recorded
- Date: 2026-10-17 for the dual-wield warrior; not recorded for the rest
- Pull request: #534

Bloodthrill's row has a ProcTypeMask of word 0 `0x14` (melee swing and melee ability) and word 1 `0x20` (main hand
only), so it would admit every main-hand special.

| Source                                      | Procs                                                           | No procs                                                                                                                                                                                                                                                               |
| ------------------------------------------- | --------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| All raw logs, bystander warriors (24 gains) | white swing 19, Heroic Strike 4, Cleave 1                       | on their own Rended targets: Hamstring 0 of 8 landed, Overpower 0 of 9, Execute 0 of 7, Thunder Clap 0 of 13, Victory Rush 0 of 6, Sweeping Strikes 0 of 6, Deep Wound ticks 0 of 30; about 3.8 expected at the white, Heroic Strike and Cleave rate (P(0) about 0.02) |
| Dual-wield warrior                          | main hand 6 of 175 landed on Rended targets (3.4%, rank 1's 4%) | off hand 0 of 200 (6.9 expected, P(0) 0.001); Hamstring 0 of 91 landed (3.1 expected, P(0) 0.04)                                                                                                                                                                       |
