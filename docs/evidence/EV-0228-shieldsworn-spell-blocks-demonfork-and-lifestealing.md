# EV-0228: Shieldsworn's spell block, and the rates of Demonfork and Lifestealing

- Claim: while Shieldsworn's spell block 1328551 is up, a Protection paladin blocks the magic spells he takes at his
  melee block chance, and the block takes the block value off; a blocked spell deals Holy Shield's damage and gives
  Reckoning and Shield Specialization's mana. Demonfork 12621 casts its drain 16603 at 3.25 procs per minute and
  Enchant Weapon - Lifestealing 1898 casts Life Steal 20004 at 4 procs per minute, each priced on the weapon's base
  speed with Seal of Fury's hits rolling them.
- Sources: one combat log of a group, logged with advanced combat logging: a level 60 Protection paladin with
  Shieldsworn, Holy Shield, Redoubt 5/5, Reckoning 5/5 and Shield Specialization 2/2, in Seal of Fury, wearing Demonfork
  (2.8 s) with Lifestealing in the main hand and a shield
- Game build: 1.60.1.71091 up to the log's gap at 2026-11-25 19:11 UTC, 1.60.1.71139 from its restart at 20:16 UTC,
  after hotfix push 114399, relogged
- Date: 2026-11-25 14:33 to 2026-11-26 03:51 UTC
- Pull request: #1280

One log and one wearer of each item. No count of public reports (Warcraft Logs) has checked these yet.

## Shieldsworn's spell block

A magic spell is a creature spell taken whose row has the magic defense type. It is blocked when it logs as a full
block or as a hit with a blocked amount. Each is sorted by whether 1328551, Holy Shield 1239570 and Redoubt's block
buff were up at it, and set against the paladin's melee swings taken in the same states.

| Spell block | Holy Shield | Redoubt | Spells blocked                | Melee swings blocked |
| ----------- | ----------- | ------- | ----------------------------- | -------------------- |
| up          | down        | down    | 12 / 179, 6.7% [3.9, 11.4]    | 7.8%                 |
| up          | down        | up      | 105 / 335, 31.3% [26.6, 36.5] | 36.4%                |
| up          | up          | down    | 75 / 222, 33.8% [27.9, 40.2]  | 38.0%                |
| up          | up          | up      | 94 / 165, 57.0% [49.3, 64.3]  | 61.4%                |
| down        | any         | any     | 3 / 274                       |                      |

A blocked spell's blocked amount reads 55 to 77 (median 75) on the 241 partly blocked spells, the paladin's block value
at the time, and 48 more are blocked in full. An earlier Scholomance log of another Shieldsworn paladin (2026-11-24,
build and hotfix not recorded) read the same: 18.2% [13.0, 24.8] of spells blocked with the spell block up and Holy
Shield down, 46.0% [39.2, 52.9] with both up, and none of 94 without the spell block.

What a blocked spell gives:

| After a blocked spell                                 | Spells                        | After a blocked melee hit |
| ----------------------------------------------------- | ----------------------------- | ------------------------- |
| Holy Shield's damage within 60 ms, Holy Shield up     | 130 / 171, 76.0% [69.1, 81.8] | 97.8%                     |
| Reckoning, out of its 1.5 s cooldown                  | 58 / 199, 29.2% [23.3, 35.8]  | 39.9%                     |
| Shield Specialization's mana, out of its 3 s cooldown | 46 / 60, 76.7%                | 99.8%                     |
| Thorium Shield Spike                                  | 2 / 241                       | 4,380 / 4,434             |

Not settled: a blocked spell gives each of these less often than a blocked melee hit does. Holy Shield's damage follows
130 of 147 partly blocked spells and none of 24 blocked in full. The sim gives Holy Shield, Reckoning and Shield
Specialization every blocked spell, and the shield spikes none.

## Demonfork and Lifestealing

Each rate is the procs over Σ(2.8 s ÷ 60) of the rolling hits: the paladin's landed main-hand white swings, Holy
Strikes, Windfury Totem hits and Seal of Fury's hits, each one roll, as the sim rolls a chance-on-hit weapon and a
combat enchant on Seal of Fury's hit. A proc is a cast of the drain's aura (applied or refreshed) for Demonfork and a
Life Steal damage event for Lifestealing. Killing blows are kept, so these target-side procs read a few percent low; no
lone-unit rate was counted. With Seal of Fury's hits left out of the rolls the rates read 5.47 [5.13, 5.84] and 6.71
[6.32, 7.11].

| Item                     | Procs | Rolls (white, Seal of Fury, Holy Strike, Windfury) | PPM [95%]         | Rate |
| ------------------------ | ----: | -------------------------------------------------- | ----------------- | ---: |
| Demonfork 16603          |   934 | 6,112 (2,706, 2,456, 787, 163)                     | 3.27 [3.07, 3.49] | 3.25 |
| Lifestealing 1898, 20004 | 1,144 | 6,112 (2,706, 2,456, 787, 163)                     | 4.01 [3.78, 4.25] |    4 |

Both rates come from one paladin, so both are observed.
