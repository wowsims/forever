# EV-0217: Book of the Dead's Summoned Skeleton (11200), level 60

- Claim: Book of the Dead's Summon Skeleton 17490 puts up one Summoned Skeleton for 60 s behind a 15 minute cooldown;
  for a level 60 owner it has the level 60 creature health and armor, takes no buff from players, and swings every
  2.0 s for 2 s of the level 60 creature auto attack dps, 14% either side, before armor. It never crits a target of
  level 60 or above.
- Sources: two combat logs of one level 60 Troll warlock (Shadow Bolt 11661 and Demon Armor 11735, both learned at 60),
  Winterspring and Stratholme; and every report of the Warcraft Logs index that shows the skeleton with a level 60
  owner (CombatantInfo talent points plus 9), 65 reports after two that log the same runs as another are left out,
  every fight: 291bwhHd8fzcPrxj 2F9bfTDrYXzwP6xN 2P4jmkMtNyJ3fYGp 2jphqGxf6y3Xd9RC 32jcraACMpBvTYHP 39BfqGnADJ1WYkb2
  3FwNGQ47tpmbP2gD 3kwcDX7JjVPdT6Mf 78wWZfmgrJYzFXvD 7KQd1Lnpqmzjfwv3 7tbWfYqJFd4vzL3p 9AhqFv3TzPjGfwgc
  AnTqJHWLhpYa4G9v AwjKkm6rcg7bDYCW CqyAYjZfgTXt6c17 DBtbX3zxp4NGTwm2 Df8YB9wqMdrQRKWt FtZN1DxThCy9LaAm
  G7X1CkT8z6PrfjWM GfyxbaQhZvL8dT9H JNTYBgF8W9tp2afq JNghP4dV8KR2pqna KpcADW3GwYqXzgM2 L8JK2cFz4mRN16Mx
  Lap9jz17btDX8M3R Lr12yNYCdD3a7HQW LzgbyJPQF2XDcq8C MDtnVjJwryA2CFR9 NP4aFz217GMDqcZv PJL3nYkF4hHAZ8jb
  PyjdHqGZ319xFvMf QrjXL7FYdBf2VmWJ TkfYVFdKBth9C3mR WdR6JhPmXq2tk3GA XLJwNW71prmaTVKP aH4JcBQkzRm8yA32
  aTKzBPYNHLb7Fw8A aWLTAv36Kx9HVPkY baQwDWF6mhCxpvVZ c3qNjGmB7x9akpgr cbTt1zxajFfNVPDA cvRt64Ldnx9wjXFa
  dwjtaQyAR98XB1nF fKDHdqbZNL9jkGnV hKMXQCcvzDY2PgRq hMk4XPnp2wAT81bZ hWXfqG7n4TC6kgRP hZBrCMNw9Tn6zWJf
  jDHFByYtnmZdrCbW jRxcftL1nPkDq7hz kYmZf7TpVjXWrwK6 p8HfCyDW2gqdvKXx pc7RGq4AK2fL9g8V qrJRyaw31zhH8gPY
  rXNYR7VkQd4xW1FC tWvzCryG1XKLMpwc vpRD61MkQH2gPGJK wXMtrHy7ap3hLD1q watv9ZMA2q3jGXgH xNqdkFAXbJwyHhcp
  y72Cf4pTxQ1WFbaV yKgMtJbmLrWPFVpG yfwhbHv8JBVXdGCF zWFnYbDkx8dM3yLw ztnPqVJdCk9YwZBN
- Game build: not recorded for the combat logs; 1.60.1.70938 to 1.60.1.71139 by the report dates, hotfixes not recorded
- Date: 2026-11-25 for the combat logs; 2026-11-14 to 2026-11-26 for the reports
- Pull request: #1250

The combat logs hold 11 summons and 215 swings; the reports 240 summons by 37 owners and 4,162 swings. Health, armor
and attack power come from the skeleton's own advanced info on its swing lines in the combat logs; the reports do not
carry them. The swing damage is the damage before the target's armor (`baseAmount` in the combat logs,
`unmitigatedAmount` in the reports), and in the reports the target's level is the one its resources carry on the hit.
The owner's level in the combat logs is read from the ranks they cast, since the trailing field of the advanced info
is an item level.

The swing row counts normal hits on targets of level 60 and below. Hits on higher targets are left out, and so are
hits while a debuff that lowers the skeleton's damage was on it: Demoralizing Shout 13730 and 16244, Demoralizing Roar
15971, Maggot Slime 1302838, Curse of Weakness 12741, Ghoul Plague 16458, Wailing Dead 7713, Cripple 11443 and Shrink 7289. The combat logs' Undead Postman and Postmaster Malown fight is left out whole, since Wailing Dead and Curse of
Weakness came and went through it.

| Reading       | Combat logs                                                                                           | Reports                                                                                                                                                                                  | Sim (level 60)                 |
| ------------- | ----------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| Count         | 1 per summon                                                                                          |                                                                                                                                                                                          | 1, from the row's base points  |
| Health, armor | 4365 and 3566 on every summon, the level 60 creature health and armor                                 |                                                                                                                                                                                          | 1.0 of each                    |
| Attack power  | 253 on every swing, 120 under Demoralizing Roar                                                       |                                                                                                                                                                                          | none; the swing share holds it |
| Buffs         | none from players                                                                                     | none from players; one report has Black March Blessing 23120 from no source                                                                                                              | none                           |
| Swing speed   | 181 of 204 gaps between swings at 1.95 to 2.05 s, the rest longer as the skeleton moved               |                                                                                                                                                                                          | 2.0 s                          |
| Swing         | 104 to 137 over 129 hits, mean 120.2 (118.5 to 121.8)                                                 | 104 to 137 over 2,260 hits, about 68 at each value, none at 138 or above; 16 more at 101 to 103 in Maraudon and Scholomance, unexplained; mean of all 2,276 is 119.96 (119.56 to 120.37) | 120.9 ± 14%: 103.98 to 137.83  |
| Duration      | last swing 58.7 to 60.3 s after the summon on the 8 summons that kept a target                        |                                                                                                                                                                                          | the row's 60 s                 |
| Cooldown      | "Item is not ready yet" 9 min 53 s and 12 min 31 s after a use; shortest gap between uses 15 min 26 s |                                                                                                                                                                                          | the item effect's 15 minutes   |

The game writes the damage as a whole number. 120.9 ± 14% rounded down gives every value from 104 to 137, with 137
about four fifths as often as the rest and 138 never, which is what the reports show (69 hits at 104, 53 at 137, none
at 138). With ±15% about 64 of the reports' hits would read 138 or more and about 77 below 104.

Attack power counts toward the swing: the 120 hits in the reports under Demoralizing Shout or Maggot Slime alone, which
lower attack power, average 105.2.

## Attack table

Every swing in the reports, by the target's level, in per cent.

| Target level | Swings | Hit  | Crit | Glancing | Miss | Dodge | Parry | Block |
| ------------ | ------ | ---- | ---- | -------- | ---- | ----- | ----- | ----- |
| below 50     | 181    | 96.1 | 3.9  | 0        | 0    | 0     | 0     | 0     |
| 50 to 59     | 1,794  | 92.6 | 0.9  | 0.1      | 3.0  | 2.4   | 0.8   | 0.2   |
| 60           | 760    | 76.4 | 0    | 8.9      | 6.8  | 5.0   | 1.7   | 1.1   |
| 61           | 270    | 68.5 | 0    | 22.6     | 4.8  | 3.3   | 0.7   | 0     |
| 62           | 914    | 56.3 | 0    | 33.2     | 5.0  | 4.8   | 0.7   | 0     |
| 63           | 138    | 50.7 | 0    | 36.2     | 8.0  | 5.1   | 0     | 0     |

The sim's skeleton attacks from behind and never crits. Against one target, 60 summons gave hit, glancing, miss and
dodge rates of 79.8, 10.2, 4.7 and 5.4 at level 60; 67.9, 20.6, 5.3 and 6.2 at 61; 56.4, 30.1, 5.9 and 7.5 at 62; and
44.8, 39.7, 8.8 and 6.8 at 63.

The combat logs' 215 swings read 168 hits, 23 dodges, 10 misses, 8 glancing blows, 4 parries and 2 crits; 18 of the
dodges came in the Undead Postman and Postmaster Malown fight.
