# EV-0217: Book of the Dead's Summoned Skeleton (11200), level 60

- Claim: Book of the Dead's Summon Skeleton 17490 puts up one Summoned Skeleton for 60 s behind a 15 minute cooldown;
  for a level 60 owner it has the level 60 creature health and armor and swings every 2.0 s for 2 s of the level 60
  creature auto attack dps, 15% either side, before armor.
- Sources: two combat logs, one level 60 Troll warlock (Shadow Bolt 11661 and Demon Armor 11735, both learned at 60),
  Winterspring and Stratholme
- Game build: not recorded
- Date: 2026-11-25, 03:11 and 20:44 local time
- Pull request: not recorded

The logs hold 11 summons and 215 swings. Health, armor and attack power come from the skeleton's own advanced info on
its swing lines, and the swing damage from the `baseAmount` field, the damage before the target's armor. The owner's
level is read from the ranks they cast, since the trailing field of the advanced info is an item level.

| Reading      | Measured                                                                                                                                                                   | Sim (level 60)                       |
| ------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------ |
| Count        | 1 per summon                                                                                                                                                               | 1, from the row's base points        |
| Health       | 4365 on every summon, the level 60 creature health                                                                                                                         | 1.0 of the creature health           |
| Armor        | 3566 on every summon, the level 60 creature armor                                                                                                                          | 1.0 of the creature armor            |
| Attack power | 253 on every summon                                                                                                                                                        | none; the swing share holds it       |
| Swing speed  | 181 of 204 gaps between swings at 1.95 to 2.05 s, 2 at 1.92 and 1.93 s, the rest longer as the skeleton moved between targets                                              | 2.0 s                                |
| Swing        | 104 to 139 before armor over 145 normal hits, mean 120.6 (119.0 to 122.2), standard deviation 9.8; 2 s of the creature dps gives 120.9, and 15% either side 102.8 to 139.0 | 1.0 of 2 s of the creature dps, ±15% |
| Duration     | last swing 58.7 to 60.3 s after the summon on the 8 summons that kept a target                                                                                             | the row's 60 s                       |
| Cooldown     | "Item is not ready yet" 9 min 53 s and 12 min 31 s after a use; shortest gap between uses 15 min 26 s                                                                      | the item effect's 15 minutes         |

Outcomes over the 215 swings: 168 hits, 23 dodges, 10 misses, 8 glancing blows, 4 parries and 2 crits. Fifteen of the
dodges came in the one summon spent on Postmaster Malown and his Undead Postmen.

27 landed swings are left out of the swing row: every one on the Undead Postman and Postmaster Malown, who put
Wailing Dead 7713 and Curse of Weakness 12741 on the skeleton and pulled its hits as low as 61, and every one after
Demoralizing Roar 15971 or Ghoul Plague 16458 landed on it. Demoralizing Roar dropped the skeleton's attack power from
253 to 120, and its three hits after that averaged 108, so the attack power counts toward the swing; at 253 the share
above holds it.
