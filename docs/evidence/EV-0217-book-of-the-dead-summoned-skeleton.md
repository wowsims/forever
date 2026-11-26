# EV-0217: Book of the Dead's Summoned Skeleton (11200), level 60

- Claim: Book of the Dead's Summon Skeleton 17490 puts up one Summoned Skeleton for 60 s behind a 15 minute cooldown;
  for a level 60 owner it has the level 60 creature health and armor, takes no buff from players, and swings every
  2.0 s for 2 s of the level 60 creature auto attack dps, 14% either side, before armor, never critting a target of
  level 60 or above.
- Sources: two combat logs of one level 60 Troll warlock (Shadow Bolt 11661 and Demon Armor 11735, both learned at 60;
  the trailing field of the advanced info is an item level), Winterspring and Stratholme; and every report of the
  Warcraft Logs index that shows the skeleton with a level 60 owner (CombatantInfo talent points plus 9), 65 reports
  after two that log the same runs as another are left out,
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

The swing row counts normal hits on targets of level 60 and below. Hits on higher targets are left out, and so are hits
while a debuff that lowers the skeleton's damage was on it: Demoralizing Shout 13730 and 16244, Demoralizing Roar 15971,
Maggot Slime 1302838, Curse of Weakness 12741, Ghoul Plague 16458, Wailing Dead 7713, Cripple 11443 and Shrink 7289.
Attack power counts toward the swing: the 120 hits in the reports under Demoralizing Shout or Maggot Slime alone, which
lower attack power, average 105.2. The combat logs' Undead Postman and Postmaster Malown fight is left out whole, since
Wailing Dead and Curse of Weakness came and went through it.

| Reading       | Measured                                                                                                                                                                                                                  | Sim (level 60)                 |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| Count         | 1 per summon                                                                                                                                                                                                              | 1, from the row's base points  |
| Health, armor | 4365 and 3566 on every summon in the combat logs, the level 60 creature health and armor                                                                                                                                  | 1.0 of each                    |
| Attack power  | 253 on every swing in the combat logs, 120 under Demoralizing Roar                                                                                                                                                        | none; the swing share holds it |
| Buffs         | none from players; one report has Black March Blessing 23120 from no source                                                                                                                                               | none                           |
| Swing speed   | 181 of 204 gaps between swings in the combat logs at 1.95 to 2.05 s, the rest longer as the skeleton moved                                                                                                                | 2.0 s                          |
| Swing         | combat logs: 104 to 137 over 129 hits, mean 120.2 (118.5 to 121.8); reports: 104 to 137 over 2,260 hits, about 68 at each value, none at 138 or above, and 16 more at 101 to 103 in Maraudon and Scholomance, unexplained | 120.9 ± 14%                    |
| Duration      | last swing 58.7 to 60.3 s after the summon on the 8 summons in the combat logs that kept a target                                                                                                                         | the row's 60 s                 |
| Cooldown      | "Item is not ready yet" 9 min 53 s and 12 min 31 s after a use; shortest gap between uses 15 min 26 s                                                                                                                     | the item effect's 15 minutes   |

The game writes the damage as a whole number. 120.9 ± 14% is 103.98 to 137.83, which rounded down gives 104 to 137
evenly, 103 on about one hit in 1,700, 137 about four fifths as often as the rest and 138 never; the reports show 69
hits at 104, 53 at 137 and none at 138. With ±15% about 64 of the reports' hits would read 138 or more and about 77
below 104. Rounded down, the fit averages about 120.4; the sim does not round, so its hits average 120.9, about 0.75%
above the reports' mean of 119.96 (119.56 to 120.37) over all 2,276 hits, and a lower share would leave out the hits
at 137.

## Attack table

Per cent of swings by the target's level: every swing in the reports, and the sim against one target over 60 summons.
The combat logs' 215 swings read 168 hits, 23 dodges (18 of them in the Undead Postman and Postmaster Malown fight),
10 misses, 8 glancing blows, 4 parries and 2 crits.

| Target level | Source  | Swings | Hit  | Crit | Glancing | Miss | Dodge | Parry | Block |
| ------------ | ------- | ------ | ---- | ---- | -------- | ---- | ----- | ----- | ----- |
| below 50     | reports | 181    | 96.1 | 3.9  | 0        | 0    | 0     | 0     | 0     |
| 50 to 59     | reports | 1,794  | 92.6 | 0.9  | 0.1      | 3.0  | 2.4   | 0.8   | 0.2   |
| 60           | reports | 760    | 76.4 | 0    | 8.9      | 6.8  | 5.0   | 1.7   | 1.1   |
| 60           | sim     | 1,860  | 79.8 | 0    | 10.2     | 4.7  | 5.4   | 0     | 0     |
| 61           | reports | 270    | 68.5 | 0    | 22.6     | 4.8  | 3.3   | 0.7   | 0     |
| 61           | sim     | 1,860  | 67.9 | 0    | 20.6     | 5.3  | 6.2   | 0     | 0     |
| 62           | reports | 914    | 56.3 | 0    | 33.2     | 5.0  | 4.8   | 0.7   | 0     |
| 62           | sim     | 1,860  | 56.4 | 0    | 30.1     | 5.9  | 7.5   | 0     | 0     |
| 63           | reports | 138    | 50.7 | 0    | 36.2     | 8.0  | 5.1   | 0     | 0     |
| 63           | sim     | 1,860  | 44.8 | 0    | 39.7     | 8.8  | 6.8   | 0     | 0     |

The sim's skeleton attacks from behind and never crits, so it has none of the reports' parries and blocks, about 3% of
swings at level 60 and under 1% above, nor the crits on targets below 60.
