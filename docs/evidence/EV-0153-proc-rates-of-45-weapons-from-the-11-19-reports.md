# EV-0153: Proc rates of 43 weapons from the Warcraft Logs reports up to 2026-11-19

EV-0157 replaces the Loksey's Training Stick rate, counted again on every wearer stream of the reports up to 2026-11-19.

- Claim: each weapon below procs at the rate its section gives, priced on the weapon base speed of each hit that rolls
  it. 28 rates are measured and 15 observed.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each weapon's section lists, from
  every indexed wearer of the weapons; 2,155 reports in all
- Game build: 1.60.1.70790 to 1.60.1.70959 by report date; the sim's database is 1.60.1.70959, push 114196; the hotfix
  each player ran is not recorded
- Date: 2026-11-05 to 2026-11-19
- Pull request: #1108
- Later count: Wraith Scythe, §Count of the 2026-11-06 to 2026-11-21 reports

This replaces the rates EV-0136 gives for 34 of these weapons, EV-0132's Destiny rate and EV-0146's Killmaim rate.

## Method

A proc is an event of the item's proc spell: a damage event, or an apply, refresh or stack of its aura. Events of one
proc within 50 ms count once, unless the same spell, target and kind repeat (two procs that landed together). A proc
that grants a damage shield counts once per apply or refresh of the shield. A target is keyed on its NPC id and its
instance together, so two mobs of one NPC id stay apart.

The rolling hits are the ones the class census (EV-0155) finds rolling a weapon proc, each on its own hand's weapon:

- landed white swings;
- every other landed hit on the melee table the wearer casts, on the main hand: the specials the census reads
  (Sinister Strike, Eviscerate, Hemorrhage, Kick, Revenge, Cleave, Heroic Strike, Overpower, Slam, Hamstring,
  Stormstrike, Raptor Strike, Wing Clip) and those it reads on thin samples or does not settle (Backstab, Mortal Strike,
  Bloodthirst, Gouge, Riposte, a two-hander's Whirlwind);
- every landing of a melee-table ability that logs a debuff and no damage, on the main hand (Sunder Armor, Rend,
  Rupture, Cheap Shot, Kidney Shot, Expose Armor, Disarm, Concussion Blow);
- Windfury Weapon and Windfury Totem strikes, once each;
- a hunter's Multi-Shot, Arcane Shot, Aimed Shot and Summon Hawk hits, and the landings of Serpent Sting, Concussive
  Shot and Scorpid Sting, on the ranged weapon;
- a dual wielder's Whirlwind, once per hand: each target takes both weapons' hits at one millisecond, the larger hit
  rolls the main-hand weapon and the smaller the off-hand weapon;
- the strike chain (EV-0145): the first landed melee-table strike of a weapon proc within 30 ms after a main-hand white
  swing is one more roll. A magic proc adds none, and neither does a strike with its own cooldown.

Sap, Death Mark, Shield Slam, Shield Bash, Shank, Thunder Clap, Volley, periodic ticks, damage shields, Deep Wounds,
Sweeping Strikes hits, poisons, imbue strikes, pets and spells roll nothing.

Execute and Judgement of Command roll twice (EV-0155). They are left out, and so are Mongoose Bite and Mocking Blow,
whose few hits do not settle whether they roll once or twice: their hits count as no roll, and every proc from 20 ms
before one of them to 50 ms after it is dropped.

A proc that lands on the target logs almost never on a killing blow: over the 45 weapons #1108 counted, 3.7% of the white swings
that leave the target alive are followed by a proc and 0.7% of the killing ones. A proc that buffs the wearer logs as
usual. For a proc that lands on the target, a killing blow (a hit with overkill) is no roll, and a proc from 20 ms
before one to 50 ms after it is dropped; a killing blow is no lone unit either. Some reports list one stretch of the
log under several fight ids, and their events come several times over: a run whose white swings repeat, to the
millisecond and the point, keeps each event once.

Paladins are left out of every rate, since their rolls are not settled: they read above the other classes on some
weapons and below them on others (owner ruling, 2026-11-19). Druids in Cat, Bear or Dire Bear Form are left out, since
they swing the paw and not the weapon. A run logged in two reports counts once. A stretch of activity with no proc where
several were due is dropped as the item swapped away, from the hits and from the lone units alike.

The rate over all hits is the procs over the sum of each rolling hit's weapon base speed over 60. Its interval is 2,000
bootstrap draws of whole wearers; under 5 wearers the section gives a Poisson interval as well. Hits split between the
hands are fractional, so the class rows are rounded. A class row counts every wearer of the class, those with no rolling
hit included; a section's wearers count those with a rolling hit.

Three rates over lone units check it. A lone unit is a hit with no other event of the wearer (any target, ability, tick,
debuff or heal) within the proc's window before or after it, and a proc in that window counts once.

- One-hand white: lone landed white swings of wearers with nothing in the off hand. Dual wielders' lone swings are split
  between the hands by timing and damage; about 10% get the wrong hand, so the dual-wield white rate reads low and is
  listed but never compared. For a ranged weapon, lone Auto Shots take the place of white swings.
- Specials: the specials the census reads, listed above, and Rend's landing, on the main hand; Multi-Shot, Arcane Shot,
  Aimed Shot and Serpent Sting's landing on a ranged weapon. The other melee hits, Sunder Armor and Windfury strikes are
  not in it.
- Two-roll abilities, priced as one roll: a true double roll reads about twice the white rate. This checks the roll
  table and sets no rate.

A direct proc logs within about [-10, 100] ms of its hit. A proc spell whose game files row has a missile speed lands
after its flight, and its window is the 2nd to 98th percentile of the delay after isolated hits, plus 10 ms.

The rate is the rate over all hits, rounded to 0.05, where the one-hand white rate or the specials rate, each with 50 or
more procs, agrees with it (their intervals overlap). Where they disagree, the rate is open; no weapon in this file is.
A stacking debuff refreshed at full stacks often logs
nothing, and a proc whose travel window passes 400 ms or reaches past the isolation has no usable lone unit: these take
the rate over all hits as it stands, and so does a weapon whose lone-unit rates stay under 50 procs; such a rate is
observed, unless a separate count checks it (Strike of the Hydra, Spellsever Crossbow and Bow of Searing Arrows, from
#1108's review). A rate is measured when a lone-unit rate checks it and 5 or more wearers of more than one class give
it with an interval within 10% of the read either side, and observed otherwise.

Scripts and inputs, in `forever-ppm-review-1108/final-1119-v12`: `analyze.py` for the rates, `MEASUREMENT.md` for the
method, and `targets_final.json` for each weapon's rate and call.

## Rates

Rate is the override row's rate. The one-hand white and specials columns give the read and, in parentheses, the procs; Was is the row's rate before this count.

| Weapon                     |   Item | Rate | Call     | All hits [95%]    |     Procs / hits | Wearers | One-hand white | Specials     |  Was |
| -------------------------- | -----: | ---: | -------- | ----------------- | ---------------: | ------: | -------------- | ------------ | ---: |
| Angerforge's Battle Axe    |  11816 |    2 | measured | 2.01 [1.93, 2.10] |   1,068 / 12,283 |      14 | 1.70 (298)     | 2.25 (27)    |    2 |
| Archeus                    |   2000 | 0.95 | observed | 0.94 [0.81, 1.08] |      221 / 4,268 |      27 | 0.87 (69)      | 0.87 (32)    | 0.95 |
| Barbaric Crossbow          | 272999 | 2.95 | measured | 2.96 [2.89, 3.03] |   3,794 / 38,444 |      58 | -              | 3.11 (1,002) |  3.3 |
| Blade of the Basilisk      |   8223 |  1.1 | observed | 1.09 [0.94, 1.24] |      157 / 4,564 |       5 | 1.14 (55)      | 1.00 (22)    |  1.1 |
| Bleeding Crescent          |   6738 |    1 | observed | 0.99 [0.90, 1.09] |     482 / 12,205 |      31 | 0.95 (69)      | 0.87 (50)    |    1 |
| Bow of Searing Arrows      |   2825 |  4.9 | measured | 4.88 [4.68, 5.09] |   2,876 / 13,102 |      19 | -              | -            |  5.1 |
| Burning War Axe            |   2299 | 0.95 | measured | 0.97 [0.91, 1.03] |     627 / 11,799 |      25 | 1.06 (168)     | 1.27 (25)    | 0.95 |
| Coldrage Dagger            |  10761 | 2.95 | measured | 2.95 [2.85, 3.04] |   5,110 / 69,251 |      58 | 2.62 (62)      | 2.87 (94)    | 2.85 |
| Cursed Felblade            |  14145 | 2.05 | measured | 2.07 [2.02, 2.12] |   7,913 / 99,778 |     112 | 2.06 (2,518)   | 2.06 (1,860) |    2 |
| Darkwater Talwar           |  11121 | 2.95 | measured | 2.96 [2.90, 3.03] | 12,936 / 119,119 |     138 | 3.09 (684)     | 2.92 (582)   |  2.9 |
| Dazzling Longsword         |    869 |  2.7 | measured | 2.72 [2.58, 2.85] |   4,541 / 38,572 |      41 | 2.49 (154)     | 2.83 (726)   |  2.6 |
| Destiny                    |    647 | 1.05 | observed | 1.05 [0.96, 1.12] |     867 / 13,701 |       6 | 0.89 (121)     | 1.06 (39)    |  1.2 |
| Diabolic Skiver            |   9475 | 1.75 | measured | 1.74 [1.58, 1.90] |      462 / 5,491 |      15 | 2.00 (160)     | 1.86 (54)    | 1.75 |
| Electrocutioner Leg        |   9446 | 3.95 | measured | 3.94 [3.89, 3.98] | 32,359 / 290,009 |     207 | 3.97 (7,736)   | 3.97 (4,305) |  3.9 |
| Fang of the Crystal Spider |  13218 |    1 | observed | 1.00 [0.92, 1.08] |     365 / 13,746 |      13 | -              | 0.00 (0)     | 0.95 |
| Felstriker                 |  12590 | 0.55 | measured | 0.53 [0.49, 0.56] |     864 / 57,996 |      21 | 0.53 (58)      | 0.67 (34)    |  0.6 |
| Fiery War Axe              |    870 | 0.85 | observed | 0.83 [0.74, 0.89] |     583 / 11,685 |      14 | 0.94 (130)     | 0.90 (9)     | 0.95 |
| Fist of the Damned         |  10804 |  3.1 | measured | 3.10 [3.01, 3.20] |   3,141 / 32,009 |      22 | 3.18 (764)     | 3.11 (257)   | 3.35 |
| Glutton's Cleaver          |  10772 |    2 | measured | 2.02 [1.95, 2.10] |   2,632 / 39,055 |      47 | 2.07 (279)     | 1.73 (77)    |  2.1 |
| Grimclaw                   |   1481 | 2.05 | measured | 2.04 [1.91, 2.19] |     783 / 11,492 |      15 | 2.28 (81)      | 1.63 (44)    | 2.05 |
| Hammer of Expertise        |   8708 | 1.85 | observed | 1.83 [1.72, 1.95] |      539 / 8,403 |       7 | 2.04 (149)     | 1.90 (70)    |    2 |
| Hanzo Sword                |   8190 |    1 | observed | 0.99 [0.91, 1.07] |     492 / 19,895 |      17 | 0.75 (10)      | 0.33 (1)     | 1.05 |
| Iron Knuckles              |   2942 |    2 | measured | 2.02 [1.91, 2.12] |      373 / 6,508 |       7 | 2.14 (150)     | 1.78 (33)    | 2.05 |
| Ironforge Greathammer      | 279895 |  2.9 | measured | 2.88 [2.82, 2.93] |  12,165 / 81,815 |     227 | 2.85 (2,475)   | 2.92 (624)   |  2.8 |
| Killmaim                   |  13016 | 0.95 | measured | 0.94 [0.87, 1.00] |     872 / 17,458 |      39 | 0.82 (210)     | 1.08 (34)    |  0.9 |
| Loksey's Training Stick    |   7710 |  1.9 | measured | 1.91 [1.85, 1.98] |   3,292 / 33,406 |     179 | 2.09 (1,001)   | 1.97 (175)   |    2 |
| Lord General's Sword       |  11817 |    2 | measured | 1.98 [1.90, 2.04] |   7,168 / 83,360 |      25 | 1.94 (167)     | 2.04 (1,081) | 1.95 |
| Nightblade                 |   1982 |  2.1 | observed | 2.09 [2.02, 2.11] |         86 / 705 |       2 | 2.17 (18)      | 1.73 (12)    |  2.1 |
| Pendulum of Doom           |   9425 |  0.9 | measured | 0.90 [0.83, 0.93] |     641 / 10,733 |       8 | 1.02 (158)     | 1.00 (37)    | 1.05 |
| Phantom Blade              |   7961 | 2.05 | observed | 2.05 [1.88, 2.23] |     780 / 14,264 |      16 | 2.24 (25)      | 2.50 (8)     | 2.05 |
| Plaguefang                 | 279876 |    2 | measured | 1.98 [1.87, 2.11] |   2,944 / 42,419 |      55 | 1.99 (560)     | 1.98 (324)   | 1.95 |
| Sang'thraze the Deflector  |   9379 | 2.05 | measured | 2.06 [1.97, 2.17] |   2,246 / 38,451 |      29 | 2.03 (653)     | 2.13 (212)   |  2.1 |
| Satyr's Lash               |  17752 | 1.95 | observed | 1.97 [1.80, 2.12] |   1,105 / 19,794 |      27 | -              | -            |    2 |
| Silent Fang                |  13953 | 1.45 | observed | 1.47 [1.30, 1.61] |     466 / 11,915 |       9 | 1.50 (154)     | 1.49 (64)    | 1.45 |
| Spellsever Crossbow        | 274293 | 0.95 | measured | 0.96 [0.94, 0.99] |  4,593 / 119,039 |     143 | -              | -            |    1 |
| Strike of the Hydra        |   6909 |  2.1 | measured | 2.11 [2.04, 2.18] |   4,070 / 35,041 |      95 | -              | -            | 1.95 |
| Sword of Decay             |   1727 |    1 | measured | 1.01 [0.92, 1.10] |      404 / 8,932 |       9 | 0.88 (42)      | 1.10 (104)   |  1.1 |
| Tainted Pierce             |   8225 |    2 | measured | 1.99 [1.81, 2.11] |     872 / 13,855 |      17 | 1.26 (8)       | 1.88 (69)    |    2 |
| Taran Icebreaker           |   2915 | 0.95 | observed | 0.95 [0.90, 1.01] |     694 / 11,509 |      10 | 1.10 (219)     | -            | 0.95 |
| Thornspike                 |   6681 | 3.15 | observed | 3.15 [3.08, 3.23] | 17,132 / 191,806 |     242 | -              | -            |  3.2 |
| Toofbreaker                | 286430 |    2 | measured | 2.00 [1.95, 2.07] |   4,905 / 77,305 |      49 | 2.01 (1,411)   | 2.02 (480)   | 1.95 |
| Wyvern Tailspike           |   5752 | 2.05 | observed | 2.03 [1.83, 2.27] |      445 / 7,293 |      14 | 2.25 (26)      | 2.36 (26)    |    2 |

## Angerforge's Battle Axe (item 11816, spell 1300783): 2, measured

1,068 procs on 12,283 rolling hits, 14 wearers, 14 runs, in the main hand: all hits 2.01 [1.93, 2.10]. The rate is all
hits, which the one-hand white rate agrees with. One-hand white: 1.70 [1.41, 2.00], 298 procs on 4,035 lone units, 14
wearers. Specials: 2.25 [2.24, 2.56], 27 procs on 277 lone units, 2 wearers. Two-roll abilities, priced as one roll:
0.00, 0 procs on 2 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 0 ms, 257 isolated hits.
Reports from 2026-11-10 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      13 |   965 | 11,204 | 1.99 |
| Warrior |       1 |   103 |  1,079 | 2.20 |

Reports and fights: 1GBhzTR8CZwcJfFV 20-32, 35, 37-38, 40-51, 53-63, 65-66, 68-82, 84-88, 90-99; 4rfgMCY9qjyvQhHD 1;
8CbGckB9gavMh6QW 1, 3-6; AfJ6DYbh9dmkgHZL 1-2; CYbmtxQ1THrG248Z 1-2; GWL9jZvR4rpaKftb 1-4, 6-8, 10, 12-13;
HZpdWaVtA8TM6bqX 1; Kfk36GzDwnPVbHaN 1-18; cgmhtvNdTGCn7f6L 73-84; dnDg7j9kRw6WtbML 55-61, 63-69; jKQDqkwmx46ZLBFg
17-24; jr7TAd93qLPxWzhp 1-2; zmDY4LVX7N8vh9aR 1-3; znBQHJTxXapkRLt3 1-11.

## Archeus (item 2000, spell 18091): 0.95, observed

221 procs on 4,268 rolling hits, 27 wearers, 30 runs, in the main hand: all hits 0.94 [0.81, 1.08]. The rate is all
hits, which the one-hand white rate agrees with. Observed: its interval reaches further than 10% from the read. One-hand
white: 0.87 [0.70, 1.09], 69 procs on 1,437 lone units, 26 wearers. Specials: 0.87 [0.60, 1.24], 32 procs on 670 lone
units, 23 wearers. Two-roll abilities, priced as one roll: 2.11, 8 procs on 69 lone units. Direct: window [-10, 100] ms,
median 0 ms, 98th percentile 1 ms, 108 isolated hits. Reports from 2026-11-10 to 2026-11-19.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       6 |   176 | 3,427 | 0.93 |
| Hunter  |      21 |    45 |   841 | 0.97 |

Reports and fights: 1NRQ9tBnTrq24D3k 1-11; 2qk4MYw7cr8KGFtm 1-8; 6z9hqHARQr3vwLXd 1-2; 76ZV8FQ1RHDhbq3z 1-4, 6-9;
FRxmtWLHr3wvB8Q1 1-8; FfAC3tWBxbp2HV9j 1-3; Hgp61KVdvPqYyr2N 1-11; KBZyYLg1WTqm6Chw 1-14; MmqC2hGvFY7x6Dgd 1-3;
NFD1RP4LdQrhfw3W 12-17; TDjrhyCkw3P9YApc 1-9; W6nhHydV1wpMmZ7a 4-5; XvWP1x3FkAzwLHZC 1-9; aQGPVW4TBcv8y3wD 1-2, 4, 6;
b186j7df4hKxNPkV 1-4; bPqj8zkJN3YatpwD 1-2; cvfBaDxQX7K2zFtk 1-7; gr9zjCAX3mY41aqf 8-12; kJq81KxZyTHD3A6z 1-3;
mtBc1WZf3ydHVPFv 1-6; pdqmMHchNGQjTwgk 1-3; qpXkvRdWxKzfNr2G 1; rNXvGcnZabhLJd9A 1-2; rQdnmBHXCpKgJqw4 1;
vN8WfZatBGgdCDHJ 1-4, 6-7; vRPQ237fWm1tFYDV 1-4; vTX3cKpNbY6jJVDa 1; wMRNLvP8CBrgd4x7 1-4, 6-13, 15-17, 19-26, 28-39;
z1APqCWaMYvxbkd7 1-3; z4WhK2CnLbaHgNVQ 1-6.

## Barbaric Crossbow (item 272999, spell 1291551): 2.95, measured

3,794 procs on 38,444 rolling hits, 58 wearers, 72 runs, in the ranged slot: all hits 2.96 [2.89, 3.03]. The rate is all
hits, which the specials rate and the lone Auto Shot rate agree with. Lone Auto Shots: 3.00 [2.88, 3.11], 2,070 procs on
20,711 lone units, 58 wearers. Specials: 3.11 [2.96, 3.30], 1,002 procs on 9,653 lone units, 45 wearers. Direct: window
[-10, 100] ms, median 19 ms, 98th percentile 44 ms, 2,432 isolated hits. Reports from 2026-11-07 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Hunter  |      45 | 3,786 | 38,317 | 2.96 |
| Warrior |      17 |     6 |     96 | 1.87 |
| Rogue   |       7 |     2 |     31 | 1.94 |

Reports and fights: 1AfDWYQT7Xjd24K3 13-14, 16; 1Gg92cW6XBhkwYb4 1-5; 1RrDAhp7ZvVYc9P3 1-6; 1W9N26dHfnPV4G7A 1, 3-9;
1y6GV7KLTZPrJH2w 1-3; 2GbXqD7KcWTfRnYa 1-2; 2MHXV8Y9jx43vhpk 3-12, 14-18; 2RygmFXpHCbnazVA 1-5; 2tJkzh16X38wrQfP 1-3;
3pTj2FQg6VvJz4DP 4, 6-7; 4ChPnBrVK2fJLMaX 4; 72XWYanybGdwPg41 1; 7VkrJpR8XyFtL6Hm 1-2; 7vyJCbH4kw8ZdfFV 1-10;
8Xxb9LDGHP7vcJfh 1-5; 92JNtax7RTkQ6cHv 1; CVFcbDGj6n8WMKyL 1-2; DACv8aKzbkxfGygr 1, 3-4; DZvkgfwrVLbpmWyQ 1-2;
DaNQ6WPf7GqTpKVk 1-8; Ghp9gxWFkBNacLfn 1-5; GkLDK8WpcdvPatfB 1-2; Gm7xJ3W81tHyTfXc 1-3; H9QxwyGt4z8MmqCj 1;
HwqhFkWZCBRTQVDf 5-7, 9-13; J4mG8yjgVCALMfRw 1-2; KBzML9kdHFCXVpQc 1-2; KYD67nVqZT4jxWyN 5-6; L7QHtkhacypgd3D1 1-2;
LnPRx96hKZN3CgTF 1; MfD1n7XFZrG4dcbH 1516 fights from 1 to 1743; NQyGkApg7zMc18La 1-3; NyG7zfKj2Mvk1FLJ 1-6;
P7MBDvyCRQtLkJHz 1, 3-8; QtGYMWzJZPNpmjvB 1; T6MXt73janNJgmyY 1-3; Td29Nzahj1cmGMWZ 1-2; TdXmWDR3QFJwpYfq 1;
VXcPMWHwZJftbCrm 1-9; VcDy8ZTRgQpWrqXz 1-3; WN1vpX6t7JkVDYZf 1-6; XV7Cf348hazGFg1t 1; YHh2t9agcvQjTAZq 2-3;
Z9CQ6fqYjzcB4tG2 1-3; aMQ63JtDnpmFz9WR 1-5; akDmAcjZrpgTFCQM 1-13; bLd41TDnmVYJZj2x 1-3; cAdCm1NLGaQnXrHW 1-4;
dc3Y84hWFnrA2wqp 1-3; dtLy89QkrcHx3pMf 1-2; fxbtCJFLjv6WMBcy 1-2; gBvQqcCZaW9TPLnV 1-2; gBxakRwhQDvnLXp1 1-2;
hMdFnHqgVRDTyKPY 1; hRzbpMrDY3cgQX7n 1-7; jRytczTZF6pBJgYL 1-2; kBHbV3DZjamvfJY4 1-3; nALy2WrwHC9qdMFm 1-2;
p9HgV7zdCKqbrwJ3 1-3; ph6FqnCZaRQLvxbM 1-2; pw8aYCczKgjMbvAd 1-4; r2BYfmZW1wkLDdNx 1-2; rKXWhHqLza39gkmA 1-2;
tbWaYgLTw8XQAzdC 1-3; v9fnrCd8TLaPDYH4 1-2; vqpcV7yDGCkbftmP 1; vrZD4X7Vb3pW9FNK 1, 3-4; vxkFd3MfgJX7NyAc 1, 3-8;
wmT1zqYBat3fHDLy 1-5, 8-9; xrGFH2DzcNBZq3Cv 1-3; zcL8MvJnPXxYhNyC 1.

## Blade of the Basilisk (item 8223, spell 10351): 1.1, observed

157 procs on 4,564 rolling hits, 5 wearers, 5 runs, in the main hand: all hits 1.09 [0.94, 1.24]. The rate is all hits,
which the one-hand white rate agrees with. Observed: every wearer is a warrior and its interval reaches further than 10%
from the read. One-hand white: 1.14 [1.00, 1.23], 55 procs on 1,529 lone units, 5 wearers. Specials: 1.00 [0.79, 1.25],
22 procs on 698 lone units, 5 wearers. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 0 ms, 56 isolated
hits. Reports from 2026-11-16 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       5 |   157 | 4,564 | 1.09 |

Reports and fights: 7fgdkyv1njDKa8Gc 6; 96pBnGdAKJMaCb3N 1-12; hYHBQ9ZJX3vFAnkj 7-14; qa1tKFnvZ98hwj6x 1-6;
xH8TDPQKb6mYragf 1.

## Bleeding Crescent (item 6738, spell 16403): 1, observed

482 procs on 12,205 rolling hits, 31 wearers, 32 runs, in the main hand: all hits 0.99 [0.90, 1.09]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. Observed: its interval reaches further than 10%
from the read. One-hand white: 0.95 [0.79, 1.16], 69 procs on 1,817 lone units, 10 wearers. Dual-wield white: 0.76
[0.53, 1.11], 62 procs on 2,041 lone units, 18 wearers. Specials: 0.87 [0.57, 1.13], 50 procs on 1,445 lone units, 22
wearers. Two-roll abilities, priced as one roll: 2.10, 11 procs on 131 lone units. Direct: window [-11, 100] ms, median
0 ms, 98th percentile 0 ms, 166 isolated hits. Dropping the swapped stretches (99 active minutes, 21% of hits, in 5
runs) moved the read from 0.78 to 0.99. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |      11 |   280 | 7,598 | 0.92 |
| Rogue   |       3 |   117 | 2,209 | 1.32 |
| Shaman  |       3 |    46 | 1,265 | 0.91 |
| Hunter  |      15 |    39 | 1,133 | 0.86 |

Reports and fights: 1WC3Bnw4A7RKpLDj 1-4; 1rdmQ9wptPWHfNG6 1-4, 6, 8-9, 11-23; 37gGDYxn4ZKLbMV1 1-7; 8KwxgLPtkpD7h2Ry
4-5, 8-32; 9CGYv1Jp24aWyhf8 1-2; 9mZgR7wqPxWMat6L 8-13; AwkvX3Tt1jqLhMVc 1, 3; B3vtVgmqMTHDAxNY 1-10; CLbmK261T3d9gAR4
1-6; CNWJgkfyAhjB3vRt 1-11; FbNk9QwxY7nG618P 1-19; H4NThcMa2dfR8gGr 1-2; NgAWV9MGwvbFLBC1 1; NwKVWqHfgk1jLvAG 1-5;
P8Znjv1fgBD2CAVY 6-12; QDVfcwHpjZRzL94q 1-6, 8-13; RPnVYQcG9xTJh1ar 1-4; WnvfNJK8dQG9xPMY 1-2; XgZdn3D9KR8cfLNY 1-10;
YhCcTtxfGJR82M7Q 1-11; ZAWJk6BdKhr9nC8T 1-6; ZyGV2gN1w6jc7zM9 1, 3-5, 7-18; f2zK39XdhnNqZJgR 1-2, 5; fjb2rRngTPa48WA9
1-4; h92ZHrfm3RFQtYvT 1-8; jC4HdxZGVNLz1Yvm 1-2, 4-6; jGrv6B1hNMDJPc3W 7-10; kmBptFTG68Dcq3jx 9-10; mQ4MqkD1nbaBXLdy 2,
4-6; npbfVjrQdDLhtRWY 20-22; pKn8wBZcQ62PyJH7 12-24; tD7Rm3GBjdqPzhwC 16-17.

## Bow of Searing Arrows (item 2825, spell 29638): 4.9, measured

2,876 procs on 13,102 rolling hits, 19 wearers, 20 runs, in the ranged slot: all hits 4.88 [4.68, 5.09]. No lone-unit
rate applies, since its travel reaches past the isolation: the rate is all hits. A separate count reads its lone Auto
Shots at 4.98 [4.62, 5.31], which agrees. Specials: none. Travel time: the proc lands a median 553 ms after its hit,
98th percentile 958 ms, which reaches the 1,000 ms isolation, so no window was measured (674 isolated procs). Reports
from 2026-11-13 to 2026-11-19.

| Class  | Wearers | Procs |   Hits |   PPM |
| ------ | ------: | ----: | -----: | ----: |
| Hunter |      18 | 2,875 | 13,101 |  4.88 |
| Rogue  |       1 |     1 |      1 | 22.22 |

Reports and fights: 1XGPdhkb28MAagN4 1-7, 9-10; 4CBckLvG7RaxFzYA 1-4; 6Krp8bkfx4cRjBaz 1, 3-15; 6WG74An9RrCdqkwD 3-8;
7BjkRJr2DZp1g8TL 1-4; AhLc6aGKMCdvBgDW 1-5; ApJhHfRkMDqFj3rb 4-5; G6yaZfLzMtN13hcA 1; H2GvZtVYNy4QzqLw 12-21;
KGQHdbFCZpt1YX23 18, 20-23; KtDJH283f4bk7R9P 1-4; Pf4Y2xdhVz1vq9rn 1-12; XqBgvYQZFw4H2czt 1-2; YWGNyaDcLwT1QR6J 1-8;
cJhVzAdLx3BHqpkT 1-4; fHbJKd4LnVXDW2gj 1-2; krpvb6WaLmFnV1KG 1-2; mtcXBg7ZkVxFHQnh 1-4; qHgyawJNVb7126xL 1;
tk86znZXxMCdwBKb 2-18.

## Burning War Axe (item 2299, spell 18199): 0.95, measured

627 procs on 11,799 rolling hits, 25 wearers, 25 runs, in the main hand: all hits 0.97 [0.91, 1.03]. The rate is all
hits, which the one-hand white rate agrees with. One-hand white: 1.06 [0.90, 1.25], 168 procs on 2,881 lone units, 23
wearers. Specials: 1.27 [0.89, 1.46], 25 procs on 359 lone units, 5 wearers. Two-roll abilities, priced as one roll:
2.07, 15 procs on 132 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 159 isolated hits.
Reports from 2026-11-13 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      20 |   535 | 10,030 | 0.97 |
| Warrior |       4 |    92 |  1,763 | 0.95 |
| Hunter  |       1 |     0 |      6 | 0.00 |

Reports and fights: 3cKTmVyb1ZPwFXkf 1, 3-13; 93NcnZPAhCfQ4wHz 1-4; BqjTQP3Ghbt9JyYn 1-11; FMcvKqTzn7m8Va3P 1-15;
Fgy2bt81NL9CAzaR 1-13, 15; GJHgk6MDpxfVdacv 1-23, 25-29, 32-33; KRCBMJvg72fHVwWL 1-17; LwTzGJXrN3HWqCfD 1-5;
MBaXr1H9GcPhDJLm 1-2; PZA4xq2WvQJ3amjH 1-5; Pc1RkCWp8wTNfFr2 3-4; QvKPgyHDkBMbcp7T 1-10, 12-21; RBmtAMLxz78Pj2Wh 7-10;
VQFaRpT6J2xLDbZg 1, 3-5; Wf8ghC6TtQKrkXdv 8-10, 12-15; bkMdQTwZmqNPvhrR 1-5; gFZ7BcKt9JhaY2dQ 1, 3-9; gGtkavy2n8bqADQ9
1-2; gJkrfBtbp7YM3PCd 6, 8-21; h2FAdwQfcJLHpzR3 1-16; j68nZp1gcrGChaRW 1-12; qMD3CVapwb6nAgx2 16-40, 42-43;
rMyzaJhG4k8wHPj6 1-10; ryqgQphLxMBz4TmH 1-4; t6HbjRZmkyx3pQwK 1-14.

## Coldrage Dagger (item 10761, spell 1293790): 2.95, measured

5,110 procs on 69,251 rolling hits, 58 wearers, 64 runs, in the main hand or the off hand: all hits 2.95 [2.85, 3.04].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.62 [1.95, 3.58],
62 procs on 945 lone units, 2 wearers. Dual-wield white: 2.63 [2.38, 2.85], 900 procs on 13,691 lone units, 54 wearers.
Specials: 2.87 [2.38, 3.34], 94 procs on 1,308 lone units, 18 wearers. Travel time: the proc lands a median 100 ms after
its hit, 98th percentile 257 ms, window [41, 267] ms (70 isolated hits). Reports from 2026-11-14 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      47 | 4,592 | 62,481 | 2.94 |
| Warrior |       5 |   465 |  5,970 | 3.12 |
| Hunter  |       6 |    53 |    799 | 2.65 |

Reports and fights: 1RVnhHGv2bwDW9Yr 1-2; 1jgTLvJQBMP6kxrm 1-12; 1qcv4MKB7R9bWxJZ 1-9; 1w3vfpYKqbrBNZPQ 1-2;
36h91AnYPjTkbxcM 1-12; 3PAwV6XFaB7hqDgz 1-12; 3THgCcxGJ4XYd6Kp 3; 3ZnAfaWCBgJhxbVY 8; 3bvPyGC6HmRNnVjz 1-11;
4xnDCWwrGKAFdNHv 1; 6734nzHBJ8yGVjFb 1-15; 6DJfd2yhL8ZNpkXv 1-4; 7YKFdxrDHLGcybjT 1-4; 7dZrzYqjKHh28mvW 1-2;
987q1pjnYdVQbCmT 1-6; 9QnWbJphKLFVgZ6z 3-8; ALhN4BDjHGv7P82V 1-7; ANWTMG9KQ3mXYh72 2; AfwXCyb2WRDMQmzt 1;
BPf7rVXwgjLF4yR9 1-2; FCXJL6RHZjzQY8d7 1-5; GCTbWRf9mrHqxAX6 1-3, 5; GyQ6cqtCxPXHaDFz 22; H9FXjaBWRpYVMq4r 1;
KF87fDdVjrhgbMpB 1-6; LYPNJyTX4rHv6cxk 1-11; Lr6VTwchNAxWqzvK 1-5; NZXhmPJaAzCDBb4d 1-25; NfqRryF7gnMcX2ak 1-2;
QKg4z1amMYyBCp69 1-3, 5-30; QYvzMfCm3pkRXPr1 1-10, 12-13; QyqcgFJZA6mLH7wb 1; R1jPGYZyTtBkKr8J 1-4; R9njq1AGDvmp36bM 9;
RW1BNk6Ybyh3XjLQ 1-21; Rfwtj4ZF1nTx2mJp 1; RvcrA3P9jznxhHVQ 1-2, 4-7; TADvnpKkdwgZGJ6b 1; VQ36t9NzpFjnmvka 1-4;
W3mNJr96cDpqYQLy 1-2; XgB2qrQDx34ZMjH9 1-3; YCwpT4hDKgZB2JFA 1; YXmAnhFPdarR48qw 1; ZHxr2qha3V4QBvLN 1-12;
acxJ2gRWPwtpDCZh 1; d6CYVkHJhGXjByTP 1, 3-11; dn62GTvrBYFQbWfV 7-20; fX6JWBtcRQrmx7gD 1-5; g41xhtdFCXZ8Vr9z 1-2;
gcnGw3DaJdrFCAYX 1-5; jVhAWnaHJgf3FPmy 1-10, 12-13; kXG9NmQ7rJKbj1vM 1-7; mD3vxz7paH4jKgTF 1-5; mNzF1MgYJypDfnPb 1-14;
nYMZcpQ8WhRTt97q 1-4; pKn8wBZcQ62PyJH7 8-24; pqfK2AgHZ7DnhvPF 1; rcCT7JVyW893mPQk 1-2; rtbzWXcByFwk7G8n 1-17;
tyFZPdH39kGq21An 3-25; w6J3Ykh4cWzq7j8L 1; x3bWXZa1KDCmFvtn 1; yjqVPg81hM6TCW97 1-2; zYBvPW8dqghTAaZp 15-22, 24, 26-30.

## Cursed Felblade (item 14145, spell 18381): 2.05, measured

7,913 procs on 99,778 rolling hits, 112 wearers, 127 runs, in the main hand: all hits 2.07 [2.02, 2.12]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.06 [1.97, 2.14], 2,518 procs on
31,918 lone units, 70 wearers. Dual-wield white: 1.91 [1.79, 2.03], 992 procs on 13,579 lone units, 41 wearers.
Specials: 2.06 [1.97, 2.15], 1,860 procs on 23,579 lone units, 109 wearers. Two-roll abilities, priced as one roll:
0.00, 0 procs on 4 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 3,915 isolated hits.
Reports from 2026-11-05 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      68 | 5,206 | 66,167 | 2.05 |
| Rogue   |      35 | 2,656 | 33,044 | 2.10 |
| Hunter  |       8 |    51 |    567 | 2.35 |
| Mage    |       1 |     0 |      1 | 0.00 |

Reports and fights: 12VhLkPa98DYFH4g 1-3; 1RrDAhp7ZvVYc9P3 1-6; 1cf7ba34WkwGyzQY 1-12, 14-16; 3CyQk69LzJjTn2X4 1-5;
3rDXa6gPxYNZJjFH 2; 6P4VND3w2r1LGQab 1-2; 6jfqGYcR9AWdJvFh 1-3; 6mrVNdqBkwgvGxW3 4-5; 6tANqvfKQP1Fkdzg 1-3;
7M48dqLp2yYHaA6Z 1; 7NqcJQvyKpCd3YAt 4-7; 7RyBGr4dYVFqJHDb 4; 7Yt1bzjCBg9JycdT 1-3; 7bch4MKwzf9HdT3v 1-3;
7qLWf2j9a8tvCNwH 1-7; 8NTz6vHAcj1W2ZmQ 1-7; 8Xxb9LDGHP7vcJfh 1-5; 8aymDGwFnWZAx7YR 1-3; 8tmqwrpyARjY6hnc 23-34;
9NbWjdLhagPZn1X7 1-2; 9aCJ7GH6QRFjgqb4 1-11; 9mKt61MCAyfP3GRN 1-4; ABX72zJxvr6YVpk1 4-5; APNwHKda8WZQRt1G 1-2;
AT1kyCR8NhMpcbjm 1-3; AdNRV84rQFWcxk6q 1-2; AtH8ZJm36y74V1d2 1-6; BwtcDCWfMLdTYFqN 1-7; CyJ13pMra2W6vdQH 1-2;
D8qzgc9h1YLf6krp 1-6; DaNQ6WPf7GqTpKVk 1-8; DmFcCT4QpH3AbhPR 1-2; G6M9XwZ1gQCTFbPa 1-2; GbrjBnmfgvphqcPZ 1-11;
GgTjx7XKkMvtwnQ3 1-5; Gv74gVkXcyPHaYzj 1-12; GzbK6XYDf1vQH4W9 1; HwqhFkWZCBRTQVDf 5-7, 9-13; HxVGdyMjDB6JCPKQ 1-2, 4-6;
KAvPryL1zCdtWTxV 2, 4; KYRtHcN6L7nmq3Cz 1-2; KgkThB9FPJv3MrZw 5-10; KpD3hj9wBk1bNL7d 1-5; KpTQqRdvf71n2Wm4 12-21;
LcaMApJGDWjg4Tbf 1-10; LwYXzRFCP8GvVKfA 6; MLTmgKZDkJaAcwYR 1-3; Mt8kPGravCBTjdA7 1, 3; MzwcDrG3fdFBRy6a 1-5;
N4tMyW9qJcBKhvY8 1-3; NZrPmMxy3LXtj9G8 1-3; Nkndb4wD6WgAK2Lh 3; PW7k29jyJ3C4chqQ 5-7; PtzakwgYVNR783An 1-2;
QDGMAJr2K3Lb91kX 1-7; Qmw4FNtdvqnkPcKy 1; RF61yp4dkXNCZTwh 1-2; T7HKGnYfbMLwWCJp 1-7; TMj2XLPvzgq3ZNJ9 1-5;
TbPzqABaXyWmLxD9 1-8; TdXmWDR3QFJwpYfq 1; VKxkQv9dyNAgpjPh 1-3; VXcPMWHwZJftbCrm 1-9; VYHAwTB2pP1r7mgJ 1-5;
W416FHmftD9rPRgA 1-2; Wk1G3TqNcZzPvRHD 1-8; XBQMrcVdkvqGCJF1 1-8, 10-11; XKc78ka3dWNgGRDr 1, 3; Y6p9C7nNx1KBWGdM 1-13,
15-33; YaBQZ2tbVFJMrAhk 1-4; YbxV4LgjKt2BfTCp 1, 3-4; ZrM8cydqRt1BpLfx 4; ZwHMphjVvdtxTY9C 1-4; a9fR1QjqFvnhgxW7 1-5;
aDRLCWyFkdqVc2Z6 1-2; aVyXCj1NpnRTFKxt 1-2; ad24TYKCAb8LwBxc 1-4; agA7pcDMCb9qJFKv 1; bLd41TDnmVYJZj2x 3;
bXYpRJw87M62gKar 1-4; cAdCm1NLGaQnXrHW 1-4; dwq8K9CRNBAQtLjp 1; dzTmC6F7xwLPJrq9 1-17; fTpYLw71QVR2qky8 2-13;
fkGrDPXmVWazJxpH 1-7; fxbtCJFLjv6WMBcy 1-2; gBvQqcCZaW9TPLnV 1-2; hYDQPX1rjn7wZLxz 1-3; jGkWVBDa7FPmpwn4 2-9;
jKfkzadV3BLp1cbr 3; jRytczTZF6pBJgYL 1-2; k9wzA3ZbgnX7j6x2 1-4; kB7JaQmdX8RZjcnT 1, 3; kHnFMcBXRrx9VajJ 1-3;
kcwp7a2jqCF91ZWK 1-4; kgFaYjKwPH1CAx2m 1-2; mTPgn7Kf2XB6ykpt 1-10; mzTZRY4Hg9aGdvxP 1-13; nPZbN3Ymk1LVtDQK 1-2;
nZ1gTJbNYPaLkx92 1; nfCzby7RrmYFxTXc 1-3; nkVby8fgJ1PFTBLj 1-2; pL3Fbdfc2vNKCxtJ 1-4; ph6FqnCZaRQLvxbM 1-2;
ptWXb1v93rkZR462 1-4, 6-9; q6fAZYgapznjxPK9 4; qcyGBRWh2xVZgKAD 1-2; qhFK6CrAVTH932kt 9, 11-19; rzchPRZJy4qNp3mW 1-3;
txhaZrbjXR7Kf2TD 1, 3; v2y9BTtzLhYr1XHg 4; v9fnrCd8TLaPDYH4 1-2; vBCYwTV1D9gxHPMQ 1-9; vba4w71K9dVTXZ3r 1-2;
vrZD4X7Vb3pW9FNK 1, 3-4; vxkFd3MfgJX7NyAc 1, 3-8; xCQDhmMVT4qdRWYb 1-13; xPmgn423azWr8F9L 9-11; zCjDnVmyRZPg2J7Q 1-3;
zLK3mTNayZMgR6vh 1-7; zPJ2w1BrKA3Dv968 1-2; zRd3nAXrVQqg9ZH6 1, 3-5; zbV7pwYGARa4Tqgk 1-2; zhj69Dt2X1b3fvMT 1;
zk682jCMJ4AdKfV9 1-2.

## Darkwater Talwar (item 11121, spell 16408): 2.95, measured

12,936 procs on 119,119 rolling hits, 138 wearers, 154 runs, in the main hand: all hits 2.96 [2.90, 3.03]. No lone-unit
rate applies, since its travel window passes 400 ms: the rate is all hits. One-hand white: 3.09 [2.83, 3.32], 684 procs
on 6,046 lone units, 50 wearers. Dual-wield white: 2.70 [2.43, 2.95], 549 procs on 5,550 lone units, 72 wearers.
Specials: 2.92 [2.71, 3.15], 582 procs on 5,429 lone units, 118 wearers. Two-roll abilities, priced as one roll: 4.98,
19 procs on 104 lone units. Travel time: the proc lands a median 146 ms after its hit, 98th percentile 601 ms, window
[39, 611] ms (701 isolated hits). Reports from 2026-11-07 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      66 | 7,790 | 71,391 | 2.98 |
| Rogue   |      47 | 4,838 | 45,007 | 2.93 |
| Hunter  |      24 |   308 |  2,721 | 3.09 |
| Warlock |       1 |     0 |      0 | 0.00 |

Reports and fights: 1CYTfP3pFDcaRrJm 1-8; 1h9vznDdxyjC7KWZ 1; 1vJgfm9dGnkDzwXb 1-2; 1w4hbgxXHpf3N6M9 2-3;
24KTy3bLGNQm1rdx 1-20; 24TVGCfvHpjtYzrD 1-9; 28GbNhyfwQ1DcRaP 8; 2GYNtHzfKmWpTPVZ 1-14; 2Z3YxGdfPhMK97kQ 1-4;
2pGYFgHwLDC17zZ9 1-4; 4CRDaYFWAqt1JnmV 2; 4LrMDtG6T8AVRJzQ 1-11, 14-28; 4ktfmxnqcbv61B2L 1-18; 4wJgRGx3rCKjQZmD 1-12;
6FZtkRcQHAKxNfrC 1-19; 6X1L4nB7zCfqk8vV 1-2; 6Zzwx92Wh4qFVLBp 1-8; 6prtaJW4g7ZTVQRK 1-6; 6wXA2CY3Bdcp7gDj 1-4;
6xhBGvgcM9H48tyK 1-8; 7BKLx2tGCwqQHr63 8-13; 7J9jfrb2LaHxgv6w 1-13; 7NHAcVG9nZCyBMjv 1-6; 7xY1zPCyKnk8DGQj 1, 3;
84c12nJmjaKCFr7H 1-3; 8BG41brwTQkLRjqJ 1; 8Bgm9rLTF67QvNwn 8-11; 8tmqwrpyARjY6hnc 1-10, 12-21, 23-34; 9GMN2DQLhrB64Atv
1-2; 9yNQP3FWHgcVvYDq 1-14; A1WnB2b3jvaJKyfp 1-3; AL1DxTaHGfVygm3p 1-15; ARwG2htz3VgjDd7v 1, 3-5; Ahv9342TJmG7bCzr
13-15, 17; B3vtVgmqMTHDAxNY 1-10; BmVrHWZvtj9GJTxD 1-13; C8J2PmQbz9dwcKXH 1-4; C8QWj29pvm46kMTF 1-3; D3Kbtnm2yaC4Tqwc
1-2; DK72jqp46VtnYHfZ 1; Dtph4q9QHVGPXyva 1-7; FR9jJ7b31XCkLmtQ 2, 4-7; Fw6bq2BvXnJmaRk9 1-3; GDWLqtcJgbrz1P89 1-8;
GntwzVb68vMyWjBd 1-2; GxWL9RvzyJC76d13 1-2; H1MQjnkcT6GraYbh 1-3; H3bRaDmzyjxTW2BJ 1; H3vdcCr6JX1gwYtq 1-3;
HG4YQy7Jf3xnFpLM 1, 3-9; HPxv4aJ1WgcTzMCd 1-13; HhFaG7mzPvZWpTtK 1-2; JzRmL9FYywDQgNcW 1-2; KqZaH876NnQV2CJw 1, 3-21,
23; Kz8FhGRY6kVTBjwb 5; KzctnTrXN64fmWPv 1-2; LHD7ZAYVNm2zjdXP 1-17; LHRhtADNW8GfyYqT 1; LK3jvm28bhWQadyr 1-4;
LhMNaKbHFvTt1Ax7 1; LzFXxt3Y1HyVK4nc 1-7; NDf1mPJRCaTM8WkK 2; NfLb7C1qDZ8cHxFp 19-39; NfjK7rVnPvkCB3XD 4-9;
NxmwqXVA7JkpP8Qn 1-4; PrNcBn64hjYw7KQ2 2-8; Q73HMJW9AkRPj4aF 1-10; QGLzd2PDcm763VRr 2-3; QRnGzYBNkWyCD96g 1-7;
QVg7ANTY4B8RhDJM 2-9; QgR6kbHvY1CBhXFj 1-2; R2zHVBktcw8j6rhM 1-3, 5-7; R7Fw1Kdcab3VYTx6 1, 3-9; RgjLQAhp3HTCwkca 1-15;
T2zBbgrnJdK7QNPc 1-8; T87bpgKckHzNDAfY 1, 3-6; TQ4PXCqWA3KGYzh2 1-5; V64yC1A3DZdWfGBP 7-25; VWF2YryN8j4qdK9H 1-51;
VaFNJB32Ah1WGLrn 1-11; VdwjgLC7qNzvPFpT 1-18; VqybFQPLtxDMnv2h 1-2; Vr12d8kxGHpKazBX 1, 4; XAyLYahcRwMWQfF7 1-4;
XWnJCQ1Mc38fRtwF 1-3; XqhkLtanYFGgx9jP 1-14; Y6wnvNGfpPVrR7XJ 3; YC214LNaFPwkRrfm 10-12; Z19fCFxJnyLc6dPN 1-20;
ZjgDb23AFJQkvpWL 1-3; Zvbh2XrFYqQPzDnJ 1-5, 7-20, 22-27; aKwXBJkG2bW6tVNY 1-5; aTHRgr6nQqVF7DC9 1-9, 11-20, 23-25;
azpj2hg1DVwnMCkZ 19-21; bH6GZhykAPqgQ3mt 1-4; c7gHWKFrZMnm2yb4 1; cfDYLdNC1t6kHQg7 1-5; dkAMWqBczZtbHTaV 1-2;
drvpchNFDMPXB8t2 1-10; f297FtyjGHmRVqYw 6-9; fLa4hw8d6pNFYMng 1; fXk6pdvPNF9wZCjc 4-5; g3nkvH9atfKRdVAh 1-11;
gAmJdBx7QyTcR3YK 1-17; ghCVqbxDmnJH6P9w 1-18; hfJZb8HL41KmtdMA 1-4; hvrBLpKmakJ8nG3W 1-19; j4HkVFBWdQvMR3cn 1;
jFDWcdYtPzkRhqB8 1, 3-12; kFxDzj7pGtTqCZg8 1-3; kHfNcJTYj3RtMFGL 1-3; khV3XMmJ4aRKzxBt 1-3; kp4FHxfa9CL7WArB 15-30;
mKHgTzDZXkBbxarG 1-12; mRxb9L2wNkMJTWXY 1-7; mj379AhvbKZVpWHC 1-7; mkGcRA9M37QyJfL6 34-55; n8ZQCDKLjHP9rdkR 1-2;
nKCDhXNbHpJaBcAR 1-15; nmFLdPWf8agkC7y2 1-3; pHZ86RKXPNY4nF1V 1-4; pKRWfNGZg96VBQTx 1-20; pwNAZn643XgzcbYa 1-3, 5-6,
8-10; qavTP8NJtW4hcQLH 5; qhGZDM3ygKAHtjbN 1-12; qjryBx72WJPmA4hT 22; qpXkvRdWxKzfNr2G 1; rTzVg1BxmZh6pFKC 1-11;
ryL9Dfkm6jtzFXgG 1-17; tD3pKRxTyNrqGchC 1-14, 16-18; tbT31mz6ahFMfnCW 1-2; tdxJaDzc8Qv2jFbq 1; tjN8BgxX7k4GyAfd 1-4;
tvrgBcqMzCRk1wF8 1-2; v12HBpQDLWa4XZ9t 1-4; vk2wGPDbQp8TZBR4 1; vtHaBZF2VAqJ19xY 15-16; wFn3hXfqxRHpQcbB 1-4;
wQAj94dkrc3MvXpD 1-6; wYmr1kfWPpVz7jyZ 6-13; wqCht6J1RkcKfVpy 1-4; wxdb7vaQp4CATFyj 1-7, 9; x1YCtNkmg2XLRbGr 1;
xQtGrv7cg6V8Z4BT 19-21; yX2kx6TbwHn4mcBq 1-5; yqntQdrPj7pvw8Fm 1-3; zcK9qfp32VnQWTMP 1-2; zhZcTYBNvwr48gP6 1-6;
zq6DmwtPJkNZy9KL 1-9.

## Dazzling Longsword (item 869, spell 13752): 2.7, measured

4,541 procs on 38,572 rolling hits, 41 wearers, 41 runs, in the main hand: all hits 2.72 [2.58, 2.85]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.49 [2.06, 2.99], 154 procs on
1,426 lone units, 6 wearers. Dual-wield white: 2.23 [1.98, 2.46], 582 procs on 6,029 lone units, 35 wearers. Specials:
2.83 [2.62, 3.03], 726 procs on 5,914 lone units, 40 wearers. Two-roll abilities, priced as one roll: 4.21, 27 procs on
148 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 1,030 isolated hits. Reports from
2026-11-09 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      28 | 3,280 | 26,778 | 2.83 |
| Warrior |      12 | 1,131 | 10,635 | 2.45 |
| Hunter  |       1 |   130 |  1,158 | 2.59 |

Reports and fights: 1FV3WyfXGCcPrT9b 1, 3-8; 4C28YPXqgTMQJxB7 5-9; 4qZXFGfc1wCMQKL7 1-8; 8F4MhkmNBxRrf6Zw 55-68, 71-83,
85-89, 92, 94-105, 108-123, 125-141, 143-151; 8X2Kc46faTnxVBLm 1-3; 9cajk6Jqd3NTQngf 6-8; A1FKkYpPxrX8tbNv 1-2, 4-5;
BCZgcRxrh7dA4ayX 1-4; BgKt3Qm2fJDAWLzh 1-23; CrR9c6VtXG4bzJhx 1-17; D4A3QNCrX19W7cqJ 1-10; FR9jJ7b31XCkLmtQ 5-7;
FzRmWck7KgT9yt6D 1; GzZH1WhTLQ3JaDXt 1-11; JTztbhZyaWHC6kxf 1-5; KVdX1Wn4GpzAH9MN 11-12; Lg82ZFDyQ4MvRm9p 1-2, 5-11,
13-26; M4CvXLZz3qm6pDRc 1-3, 5; MVcaxDQwWLpnK6HJ 7-9, 11-12; PKRTfgw2Q7ZBxt4r 1-26; Q7BxMTcG1ZbgX2v8 1-2;
QftYMR2haLcXZ7Nz 1; T71XnfaF3cY9M2Rg 1-13; THqFxLkCP6ayd2tZ 5, 7-9, 11; Thf9agqHJyRmtdpD 1-5; VLmAcpbQxWf1CtaR 1-17;
XCJdxAF17zqHMhLY 1-2; ZtndFBMraKvcLfw9 7-9; dmKgVGCJMQpq3t17 1-14; dzXT7JngpxRPwG32 1-17; fCrvDxhZXqJAg62m 4-10;
fkVmJwWZjMy3LBX4 1-3; gTPNZcJ9mG4dRC2X 1-4; h6wkBb9Z7jFtTrnp 1; nF7RcakdVTQD2tvH 1-3; nLk2TKDJWP6vBgam 1;
pKj3bDrf1h2qtNMF 9-15; qW3KvYDpw6ARxZbc 1-7; tkQKL9bGT6JVR3na 14-16; xM9bRGyajmBFvCqt 1; xqy2GfPj67XCwrT3 3-4.

## Destiny (item 647, spell 17152): 1.05, observed

867 procs on 13,701 rolling hits, 6 wearers, 9 runs, in the main hand: all hits 1.05 [0.96, 1.12]. The rate is all hits,
which the one-hand white rate agrees with. Observed: every wearer is a warrior. One-hand white: 0.89 [0.67, 0.98], 121
procs on 2,277 lone units, 6 wearers. Specials: 1.06 [0.64, 2.26], 39 procs on 616 lone units, 5 wearers. Two-roll
abilities, priced as one roll: 2.01, 91 procs on 753 lone units. Direct: window [-20, 100] ms, median 0 ms, 98th
percentile 0 ms, 184 isolated hits. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       6 |   867 | 13,701 | 1.05 |

Reports and fights: 1ZyhaVDdFtTYWMcC 10-12; 3jARkacpG1LJgCTN 1-2; RFtjqcPd9yKr26wh 1-4; VLz78BfZyWa3Cxr2 1-9;
fdj2MNckqQ3YDHmR 1-2; gcBkwTfAr76pHzba 1-3, 5; nX1qbatTzGZ3CAwB 1, 3-7; nhtYVjb7CAw9DvQR 1; tkPZYqTRgCyWG3jM 1-23.

## Diabolic Skiver (item 9475, spell 18206): 1.75, measured

462 procs on 5,491 rolling hits, 15 wearers, 17 runs, in the main hand: all hits 1.74 [1.58, 1.90]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.00 [1.64, 2.37], 160 procs on
1,654 lone units, 15 wearers. Specials: 1.86 [1.34, 2.19], 54 procs on 601 lone units, 14 wearers. Two-roll abilities,
priced as one roll: 1.43, 4 procs on 58 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 198
isolated hits. Left out: 29 of 46 runs in Cat or Bear Form. Reports from 2026-11-10 to 2026-11-19.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       3 |   291 | 3,403 | 1.77 |
| Hunter  |      13 |   171 | 2,088 | 1.69 |

Reports and fights: 2zXhygVYpkrMZGxT 12-15; 7gWZKLnBmty1cT3J 1; 94fVcDCNBLztXh1k 1-12; FvqNn8XAhHD6M4pY 1-10, 13;
HGn9XrYNPvwKtC1m 1-5; NR1k2zLFqKt7GXAD 6; RXDCK9PtZy3Jp2QA 1-3; VLmAcpbQxWf1CtaR 2-17; X47TDyBdAh2nQ6LK 1-10;
YHXAcJxWG89pMFgf 1-9; bhyYq9HXd4xvAwtJ 1-13, 15; ghWq2C84BJj9bQHd 2-4; jZtdhVaHckGN72L9 1-11; rDJL7nqX1VZHm9zk 1-8;
tk86znZXxMCdwBKb 18; xwrkGVNA2Z7Hq4fW 1; zTbHr3BPvKxRCMZg 9.

## Electrocutioner Leg (item 9446, spell 1292851): 3.95, measured

32,359 procs on 290,009 rolling hits, 207 wearers, 250 runs, in the main hand: all hits 3.94 [3.89, 3.98]. The rate is
all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 3.97 [3.87, 4.06], 7,736 procs
on 68,738 lone units, 119 wearers. Dual-wield white: 3.36 [3.07, 3.62], 2,439 procs on 25,594 lone units, 87 wearers.
Specials: 3.97 [3.86, 4.08], 4,305 procs on 38,306 lone units, 201 wearers. Two-roll abilities, priced as one roll:
7.88, 136 procs on 609 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 10,318 isolated
hits. Reports from 2026-11-07 to 2026-11-19.

| Class   | Wearers |  Procs |    Hits |  PPM |
| ------- | ------: | -----: | ------: | ---: |
| Warrior |     132 | 23,244 | 210,509 | 3.90 |
| Rogue   |      58 |  8,748 |  75,983 | 4.06 |
| Hunter  |      17 |    367 |   3,516 | 3.68 |

Reports and fights: 19DdXA3BwjZHFbVJ 1; 249KNqjY61fcwdMn 1-9; 27dNt8HLBCPpAmYZ 1-31; 2N4yhcrz8GW7mRnt 1-6;
2jYyAK1RrPvTC8WM 1-5; 2ywhMnJq4Y3zaWtg 1-9; 3HRvtBnAbfM9xTkK 1-3; 3HZNbTwqdtnLFg4K 1-4; 3wCRWQcyNrgKvHz9 1-9;
46bPRkzVHgafvMwA 1-12; 48jfdvMD9HQbP1tC 1; 48pmMdNLnqxWj2Fz 1-12; 4C28YPXqgTMQJxB7 1, 3-9; 4M8gQTJhZvRPWf2B 1-5;
4Ppkndm9fWvTh1Gr 1-5, 7-12; 4Wq6Kcgnz3r29fTh 1, 3-5; 4ZNwHaLtygF1k3xC 1-3; 4jCpGrhx3NvnDmZ1 8-11; 4nbDkxaLmgcwGYQv 1-9;
4ynKGp8XCQPNmJLF 1-4; 629cLJxdjGkVBDTa 1; 69atW41MxnTY2cXN 1-26, 28-37; 6Pwh4WBYyAbMGD3c 3-6; 6QvVFqNYty8bzrcf 11,
13-21; 6WkrKt3RwHGmb4Zf 1-3, 5-7, 9-15; 6aBqgmpQhVAvn94Y 1-5, 7-9; 6jVXd48tLqWgQvK3 3-10; 6z9hqHARQr3vwLXd 3-4;
7f6DBRAPMLGYzyQm 1-2; 7kVKd6JMNTCj3P4z 1-10, 12-16; 7n6rKTfk1vhzgXyF 1-2; 7tYjam2PXvfcGz61 1-14; 7vxRAKHWC83tgPDz 1,
3-7; 7x3ycAVKT9Ba4dMZ 1, 3-7; 8VQDf7JtwbTknMLz 1-3, 5; 8dT1VNMQfLgrj9vc 15-17; 8rCvtag1mFYqBHM3 1-6, 8; 94FmbDrtqgPRKyH1
1-5, 7-13, 15-36, 38-57; 97CVmB4zqyrcWpjT 1-7; 9T6Zbt2QxPXhmJpk 1, 3-27; 9bd6VQvJR2ZqGW3k 1-6; 9g7NQBqDH6FGfP8L 1-4;
9v23hzNwDpWM67Xy 1-3, 5-18, 21-22; AGjFnqDY42gfQBTx 1-2; ARqz9awMbfr6QD7Y 1-2; AhxGmJv3nYDzTW9q 1-13; AkBH7pF6XvZc3Gdw
6-8; AxCwvBkjqYVtLgnc 1-13; Azq6hMZn2RyWwagv 1-6; B6jfw4mcnPxWCKYD 2-18; B7wf2rAVZJLQgT3m 1-5; B9RcfwgQpTHL8m31 1;
BTAM7pf2VtYCnJ8Z 1-16; BydpaqJVW29G8QAc 1-9; C4TgLGnkKZcPqF8W 1-12; CBK8LRX4tZzWANqF 11-13; CDZG21zLfmMdh6FN 1-5, 7,
9-10, 12-15, 17-21; CLXHZ3c2GyN6mYBD 1-4; CcAYgrB3Q1w8KJvV 1; CfFMHPahckGpR8Wd 1-3, 5; Ct8XcxY26Mf34NjR 1-9;
DF6Vy2mcQBbrNp7n 1-6; DTrLRQbfXny12B6t 1, 3-8, 10-11; DVdNpYmZfz9MGh1X 5-17; DcyRfjPp3YzHMkra 1-18, 20-22;
DqHT31fVnh4rjaky 3-7; DzWf9A8B1MNnrm2L 2-5; F9DyXdCazKmPqWfN 3-5; FBQzxArfjdCZp6Lg 2-11, 13-18; FarzJZ1mH2ydBWxh 1-17;
GCTbWRf9mrHqxAX6 1-3, 5; GZzaJH1QVjMKd8qA 21-34; GxTn6w2dD9MmXypP 1-9; H7PBRD9CqnGxkF4t 1; HRJA3nGpVzDaXyYP 4;
HZCD4zaBgtKh63rb 1-10; HmnhWYy8BtCGv6R7 1-3; HqzVaYrZLQxAPKhF 4-10; J2XL6AyZqgpjx8QK 1-13; JBj14VkRgmcTqDFd 1-7;
JHDxh1BWZFGTfQ2n 1-4; JWbdPL3FRphQcM6z 1-10; JZRnFQDrmYqBycvG 1-7; JcanfA1v3rXN6bBW 18-22; Jcx8g4CXq1QBpzVP 1, 3-12;
JzRmL9FYywDQgNcW 1-2; K3jWRx2qQ1aXy9PA 1, 3, 6-13; K47ntNhqVQz83H9g 1-2, 4-5, 7-9; KAfmrqFHXh7TzdL1 13-20;
KFdN4BrZP9X6tbcL 1-9; KGLhazAPkWfd9rw3 1-6; KTz6Gaym8fdCxhkb 3-6, 8-13, 15, 18-23, 25-26, 28-45; KZzq7NTcBXj4Ln9r 8;
LBYKWV4xgfkTvHRt 1-6; LJ43xqm7ATnCH8MX 1-13; LKkr1pxAtT7CvcZN 2-17; LMdRqJGhV17XwWyT 4-22; LQbJn6gcWz2D3HCT 1-2;
LWtG286RrDcJZ1xw 1, 3-5, 7-11; LXNz7Ym8rvwqcHax 1-8; LayrxDGdQb712wpC 1-10; LjFr1v7pJKVcfhWB 1-11, 13-14;
LwTzGJXrN3HWqCfD 1-5; LzXRwTc3YCfFkG89 16-21; ML1xyRfnNVbqkhAX 1-7; MmTXHCN4KDfpaFwz 1, 3-12; Mr7nxYapkv693wRh 1-10;
MvgK14QZDxjWGBXp 3-14; N9qCv8pYmPQRb7a3 1-18; NKnQBL2MrwCyAHzd 1-6; NTcKDtjPwZk8QaXg 1; NXAxdr9JVKLzFDWg 1-4;
NXgv93kr2G7zyTxm 1, 3-9; NcRqzah1TQjMr3nF 19-25; NgVPn7M3yX6AB8rT 1, 3-8; Nh9wMxkpzqJTPA2f 1-8; NmwLfZHTpdqPWAMC 1-5, 7;
NnmBjCcpKrxkg9GW 1-4; Nr8n1KpjcYwk2CvH 1-5; P1QTDgLwfytJKrBh 1-8; P3JYyjrc1FWgfN6q 1-4; PNnXcA9hFBx7Wtdj 7-11;
PpLZ7FAgkvTwc1Cr 1-11; Q6hKWVt9H47vNwnb 1-8, 10-12, 14-16, 18-23; QP32XzqKRHTvZ6FN 1-4; QVg7ANTY4B8RhDJM 2-9;
QwL7NjYPXAmzakRp 1-9; RD6bnVNTgPhGYCvy 1-4; RLnWB4FrJgpMyNZj 1-5; RPMNvqn7tXwJjak9 8; Rfwtj4ZF1nTx2mJp 1;
T13Kx4nhGJW9vR8Y 1-16; TC29cvWmBMhfZJzV 1-2; TGmKaXrQPp7czjx1 1-4; TJ8fyKCp4BHtWdLm 1-2; TdP4nFMvCf2mW6rB 1-18;
TgVP6hxZKqY7MGQ4 1-5; TnbzrVaxtHcpMKLY 1-3; V3L7kJDzTGg4QqCF 12-18, 20-27; V4LaQWhtxHqXDpwT 2; VQbrqK3D1RHXWpa4 2-20,
22-27; VaFNJB32Ah1WGLrn 11; VaYkQ91Tt74yKbrc 1-2, 4-8; Vm9wxAyCHQX7q8Bk 1, 3-4; VxmDcHTK8LqgNCwQ 1-7; W2rLkBmgb864JNDh
1-6; WA9GBTPQRKvazHmy 5-9; WBZ7gMLqXFcQrzDJ 1-2, 4-20; WCQGyn1J9q4Kdzbk 1-5, 7-15; Wfw3CbM28cyT7Nn9 1-9;
WjgGnJtpyKzVY3b1 1-21; WrHQ38tb4KTnAy7h 1-3, 5-11; X3RtbZzTdYAxfKV4 1-4; XhZAV37vMFbYWRt6 1, 3, 5; XpnkZCf7j6qDcmy9 1-4;
Y2n7kAxTCch1mgf3 1, 3-17; YZyAfq6XN9k4a8md 1-5; Ydx67HwgajJKpbNr 18, 20-22; Yq9kmzBHDP3jM2Z7 5-6; YzMA1vd3BrqLWxhH 1-3;
Z182HvNhFzdqgCrj 1-5; ZD7YLw2qy4mWRhJ1 1-4; ZKzP2YvapHfqM8WC 1-3; ZQ2NaCptbjmWKJ6P 1-18; ZaPRprzfmqY4CtF6 3-5;
ZrD7NdhjWX4zG2PL 1-5; aDXfbQzV1Bwr9n2x 7-8; aKwXBJkG2bW6tVNY 4-5; aMNVKZpg8tRkhBP3 1-17; apqZQxfDMJB4LPTb 1-7;
avnKJBPL6ANFTC3V 1-2; avrcxzCymh1DNqA8 1-19; bjcHnakF34w6YLvC 1-6; btxm3BCN4jX61VhY 1-7; bvzqmdxkA4KVQaZh 3-10, 12;
c7XrRMZWYmzwgfq1 1-3; cBX7WxzLNYrmMdyQ 4-6; ctRfwLFrZ8BpaTJY 1-7; d2wFYHqnNcJpA1gt 1-10; dZYgMabwJXnD1rfK 1-4;
fWRgvP17z8yrL36h 45-55, 57-62, 64-68, 70-73; g7b3Jf29CVhKtYMD 1-2; gdMNAjD2Kyrtqh97 1-8; gn4HAC1RqMfLt8wN 1-10, 12-21;
gnMWYZa9ArC8jbNc 5-12; h3bCaBdyYJMPzvRk 9-19; hAJa2PcBmdnNbGFq 1-16; hMaTJ3ZtDzqPWHfx 1-7; hZvn9rLzYyNTaVBR 1-5;
hatAfwb9QHydP62c 1-2; hf1pF42aPHMv9kLX 26-36; jDHpNnbqf68kZxA9 1-3; jLtGKrwPDBMmF3Xc 1-3, 5-11; jPh97VNpz3Ark2YQ 1-4;
jVRTr4dWa8xfZHgv 1-7; jgByJqcWN6HFdLY1 1-7; jz3qpYLQ4PwTyHtr 1-12; kB3gRfF4hYJ2nAQH 1-5; kX94Vmn8tpJzGTD2 1-11;
kmBptFTG68Dcq3jx 9-10; kr13fAvjzZXMdNDT 1-2; m2jhcywg86TDGMZn 1, 3-6; m3dN8Wkg4LyqXZD2 1-11; m9PX3r2NY8jLQ7Dw 1;
mN7ZTVPw8WdJrhA1 1, 3-5; mR9Ch8QVfX1pJkT4 1-2; mwAFfdjYy3chWMxb 12-15; ndHRLVB9Kaz7xpqT 1-2; nqmgcYrPz9B1fLDx 1-17;
p4MR7G2atyhQwV6k 1-6; p92NRBfL3QCYrAgy 10-11; pbhgCm6PjYwZ3fRd 1-3; pbvkXazLqcHFhgRf 4-7; pkJPxXb3aYLmG8qK 1-4;
pwNAZn643XgzcbYa 9-10; pwd6RvtM2NYqmfGT 1; pyY6Kg9bPRkQfMHL 1-22; pzx4LKwq6PB8FrWy 1-10, 15-17, 19-25, 27-32, 35-36,
38-41, 43-46, 48-49, 52, 54-63, 65-70, 72-80; q6VRad2JDnpxXYAL 1-13; q7xagw1ZkdvnLQmc 1-6; q8wrpJkfFWdVg6Tc 1-3;
qRJ8DvjZfh3Wzwgx 1-3; qT76ZPycRHrL41p9 1-4; qbDZa2XcdrM1npAF 1-6; qhGZDM3ygKAHtjbN 1-12; r9FD3YQGHcqdjC7g 1;
rBy3ctF2NpKPQGWb 1, 3; rykhNAvxCwXG4m3Y 1; tAFwQbTHpJMCxZXP 13; tCj8cKdQr47MN6qb 1-14; tc1qGLjpgVrJ8WdQ 4-5, 7-15;
tyAMjpZb9Yq2wkV7 1-4, 6-10, 12-19; v8VbYrxBCyMdJhN6 1-4; vkG2m6pNJ8PWrQh1 1-4; wV9frJcGT4nzm8vK 1, 3; wYmTW817qdpJHRcK
1-2; wrVYM4xbWLNaczhP 1-3; xR6rvTQLBwZ9nkzy 1-12; xmvMVYpyRjX3n8d7 3; yFWHKnAfCQpjNk96 11-12; yqVzktK2B7Lgp9Yb 10;
z9mYPDLpcM81JCFA 1-3; zD9hTZPfwC38bgq1 1-2; ztamxJTvkp82GK6h 1-18.

## Fang of the Crystal Spider (item 13218, spell 17331): 1, observed

365 procs on 13,746 rolling hits, 13 wearers, 14 runs, in the main hand or the off hand: all hits 1.00 [0.92, 1.08].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: none. Dual-wield white: 0.90 [0.72, 1.08], 103 procs on 4,312
lone units, 13 wearers. Specials: 0.00, 0 procs on 27 lone units, 1 wearer. Travel time: the proc lands a median 97 ms
after its hit, 98th percentile 237 ms, window [22, 247] ms (96 isolated hits). Reports from 2026-11-14 to 2026-11-18.

| Class  | Wearers | Procs |   Hits |  PPM |
| ------ | ------: | ----: | -----: | ---: |
| Rogue  |      12 |   349 | 13,113 | 1.00 |
| Hunter |       1 |    16 |    632 | 0.95 |

Reports and fights: AaKzvcY1wmH9pBbq 1; D46Xm8ncR9WbTFrL 1-31; G7X1CkT8z6PrfjWM 144 fights from 695 to 871;
KNTYZJ3Xg9d8HftG 1; QXFd74zDW8JZMmKx 12-13; Tx7MPH9f8Ga6hdrF 1-20; XPYV9J6DfrMKxGqh 5-6; XgQYbwGBZcHLKRxV 4;
c8mjK4TdzV2M7XLF 1-5, 7-46; h3P9VjGMZCJWfTcm 7-10; k1taHbwL76Cnxv4m 4-10; kaH1G4ztmYpgMTJ6 1-2; xXWkB3fqAjt9NJGH 5;
y6D3jqgWt8TvpNYM 1.

## Felstriker (item 12590, spell 16551): 0.55, measured

864 procs on 57,996 rolling hits, 21 wearers, 35 runs, in the main hand or the off hand: all hits 0.53 [0.49, 0.56]. The
rate is all hits, which the one-hand white rate agrees with. One-hand white: 0.53 [0.38, 0.67], 58 procs on 3,873 lone
units, 6 wearers. Dual-wield white: 0.44 [0.40, 0.53], 222 procs on 17,783 lone units, 15 wearers. Specials: 0.67 [0.42,
0.97], 34 procs on 1,804 lone units, 11 wearers. Two-roll abilities, priced as one roll: 2.52, 1 proc on 14 lone units.
Direct: window [-13, 100] ms, median 0 ms, 98th percentile 0 ms, 212 isolated hits. Reports from 2026-11-15 to
2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       9 |   551 | 36,551 | 0.53 |
| Rogue   |      11 |   302 | 20,914 | 0.51 |
| Hunter  |       1 |    11 |    530 | 0.73 |

Reports and fights: 2jJ1L8FPHzTB4fKG 1-5, 7-12; 2wMznDRvmkpbtHxA 5-7; 2wPQmVqbdCcZW9gv 528 fights from 1 to 629;
4TxMfmFpzBJ6aN1j 5-7; 7tbWfYqJFd4vzL3p 1-2; 8Gw7m39ZXFz1y2fv 1-32; 8cqCXZHgpw9Px2hK 1-28; B6bvH1fQ4kAcyRzt 1-2;
FhG3AxYL9ZNP4RvV 1-15; KkRW7Am81LaJqHYz 1-3; M1tPx9DBZGYgCa4f 1-3; MqzhQN9FPHL3fBvT 1-2; NjWKPpHgymCT832A 1-19;
PKBd6Vz2Gkjwp9Zt 1-2; QNfaZ9MTnWpFrqb4 151 fights from 1 to 199; QkVRc329w8nLTzKp 1-3; TNW9rHBt3MD1X8Ag 1-2;
YFtvD87kznr1CThg 6-7; bP2CFzkZTLJ8q6Dn 1-8, 10-16; cgMGCdPJQ84wzaBV 12-13; dnDg7j9kRw6WtbML 61, 63-69; gYm2HAtf4bnC8QZT
6-22; hZLykw9dRnYbtvKP 1-2; jXTVfy3ApWKPYwDG 1-22; jn3dwzWPZ6r2ky1K 1-3; mpgnDQbLWXY8jJP1 1; n7hQW3NYzJwmDaG4 36;
nX1qbatTzGZ3CAwB 4-7; pRA73YvkarFVLNCZ 1-3; rdcZKVykDFH4f7mn 1-4; tkPZYqTRgCyWG3jM 22-23; vZqVmLYQa6jwP3A9 5-16, 18-21;
yPVkKpqtfJM27WFB 1-7; yfKb1MthLFDc4ANG 1-7.

## Fiery War Axe (item 870, spell 18796): 0.85, observed

583 procs on 11,685 rolling hits, 14 wearers, 15 runs, in the main hand: all hits 0.83 [0.74, 0.89]. The rate is all
hits, which the one-hand white rate agrees with. Observed: its interval reaches further than 10% from the read. One-hand
white: 0.94 [0.81, 1.12], 130 procs on 2,299 lone units, 13 wearers. Specials: 0.90 [0.71, 1.03], 9 procs on 167 lone
units, 2 wearers. Two-roll abilities, priced as one roll: 1.25, 3 procs on 40 lone units. Direct: window [-10, 100] ms,
median 0 ms, 98th percentile 1 ms, 110 isolated hits. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      11 |   518 | 10,015 | 0.86 |
| Warrior |       1 |    50 |  1,040 | 0.80 |
| Hunter  |       2 |    15 |    630 | 0.40 |

Reports and fights: 1L73Gh9FVMm2J6jg 1-4; 3A1rzFqNBxTQykwC 1-2; 4C28YPXqgTMQJxB7 1, 3-9; 4LrMDtG6T8AVRJzQ 3-11, 14-28;
7n6rKTfk1vhzgXyF 1-2; Kzn8AkXFr3yvMZHQ 1-9; WAxmzHdKLVabk6nR 1-8; Wb9tmLq2cGKFjw3v 1-6; ZTxYXrtHQkcNGjA1 1-8, 10-18;
hVxH72qr8PYB3DNm 1-11; pravyY1JCwDAxZLb 7-24; rTLVqFpfa9GJQgHh 1-2, 4-5; tvBjG6TXDZq8gdrz 4-5; vkayWxntqQJ6pfjB 6-9;
zDKZYwnyML7WFHXB 1-6, 8-17.

## Fist of the Damned (item 10804, spell 18084): 3.1, measured

3,141 procs on 32,009 rolling hits, 22 wearers, 25 runs, in the main hand or the off hand: all hits 3.10 [3.01, 3.20].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 3.18 [2.93, 3.53],
764 procs on 7,598 lone units, 13 wearers. Dual-wield white: 2.86 [2.63, 3.27], 297 procs on 3,284 lone units, 9
wearers. Specials: 3.11 [2.75, 3.51], 257 procs on 2,608 lone units, 14 wearers. Direct: window [-10, 100] ms, median 0
ms, 98th percentile 1 ms, 856 isolated hits. Reports from 2026-11-13 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      19 | 2,994 | 30,656 | 3.08 |
| Rogue   |       3 |   147 |  1,353 | 3.43 |

Reports and fights: 1GN2LMtwvY4JQHba 8-11; 498vhQAtCw2XgZJM 11, 13-15; 4vcn3WF1RxXf8VZt 6-17; 6MYqkGpyJxDVCvRg 1-3;
8xjZmDw2LKv9bckh 1-15, 17-23; 9MCqcwjzLk4JBARG 1-5, 7-14, 16-22; 9WGcCxKHQ4pkX6Rj 1-11; B1qcTGpnHCrQ64Xa 1-4;
By41zJVmM3cdPQkG 1; JMnDtLGxf4FqCV6w 1-3; KYmwz2q1tG8yAcdP 1-6; NcwFg2bZjkrnXfQV 3; RTfqrCdVvLGYzj1A 2-3;
Tx7MPH9f8Ga6hdrF 1-20; VjT3aMZghKCyzGrt 1-6; WMBFLQ43XvqTK7rG 8; Wb8w1TtxA3kLF9M4 1-8; Z4w7HKFPvcrTWd3M 5;
fNMJcqDHb9ZzPLnk 5-10; j8bQnLtWwRMBXkDz 1-4; mbZcgPavhdG7fKDW 1-20, 22-24; nmwM719fpABxvjdz 5-6; qprynNbXtRZV4DHP 1-7;
y6D3jqgWt8TvpNYM 1; yYPpk4TqAhjX3xC8 12.

## Glutton's Cleaver (item 10772, spell 18075): 2, measured

2,632 procs on 39,055 rolling hits, 47 wearers, 51 runs, in the main hand or the off hand: all hits 2.02 [1.95, 2.10].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.07 [1.86, 2.27],
279 procs on 4,039 lone units, 12 wearers. Dual-wield white: 1.69 [1.58, 1.79], 552 procs on 9,816 lone units, 34
wearers. Specials: 1.73 [1.33, 2.13], 77 procs on 1,339 lone units, 15 wearers. Two-roll abilities, priced as one roll:
5.82, 13 procs on 67 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 704 isolated hits.
Reports from 2026-11-10 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      18 | 1,630 | 24,533 | 1.99 |
| Rogue   |      17 |   706 | 10,336 | 2.05 |
| Hunter  |      10 |   229 |  3,198 | 2.15 |
| Shaman  |       2 |    67 |    989 | 2.03 |

Reports and fights: 1RVnhHGv2bwDW9Yr 1-2; 6VJxcqnYa4wdBWQv 1-9; 7NgdQzG6VcRaAJ9X 1-6; 7tCZB31vV2Jp6XWd 1-2;
8vDJ9jmHd4FA7Twt 1-2; 92rYfXaKn1CckRM4 1-2; 9VacfmDpXwCgzZWq 1-9; CR4DHc2wFgjd7yrp 11-13; CnNwpaYW83cJr6fR 1-6;
Gx3AXVk9cCaR18Df 9-10, 12-16; Jad2wbYTXQjBc1n7 1-2; JcwBmYGVgnMQH9xd 3-19; K1zQgCcqbnt3kHXT 5; KzNcw4X2JPRCgVx7 11-19;
L7JpRfWtAZ6T1hGx 3, 5; M7BnDF4VPXcWRaZ1 1-5; MkrFhTq2HPnzmCVJ 3; NrJZ7avW13QBVTGw 1-7; P8Znjv1fgBD2CAVY 2-12;
PDT2yACNzd4LMZ98 1-14; PJXwqxpANGtya937 7-11; PkWRr8jZp9ATL7N4 1-23, 25-26; W7VPQB18nHfkMaAc 6-13, 15-31;
ZDnyzq6XAxdk3K7g 1-5; b2YKZcFzWPtkxfpg 1-3, 6; bCcFJj4aQwP1rt8m 1-4; d8c6LnPYbKkDhgVw 1-5; dn62GTvrBYFQbWfV 7-20;
f8bZ6ytgj7mMGHCk 6-12; gBDMLv8J1K7nqkpd 1-25, 27-70, 72-74, 76-157; gBYCvhH6MFmbVwkp 1-3, 5-6; gaF6jK78BpnfZqxD 1-4;
gnM2vVF7zWXhm9KP 1-4; h9XqNAJk6aPntHpv 12; hbnaypTcAWmqDMt6 10-13; kpPdVKFmw72nCyM4 1-3; mQJtAfacYp7wGryx 7;
nX9kCRqh3WGa8g4d 5-6; p4zvjxb3QZJY2d9a 1-2; pjZHqhYvkLKTAQCJ 1-2, 4-5; q1ytzBDkYgd6jZMH 1, 3-9; qRMCzhyNYfmHAcZk 1-3;
qnyc1zwaB4FrvM9x 1; rzvB21j7yD9hpaVF 7-20, 22-30; t681v9FWrLq4Tkxp 1-3, 5-7; wZGf4Y6tjLDyTP92 1, 3-7; wvbng2zVTYRrc9Zd
1-2; xtv8KX4ZzfAmLFPw 3-12; yRQMz8LfGjHNBx1T 1; zwPKLTQ4W6YkyRqd 3-8.

## Grimclaw (item 1481, spell 1292158): 2.05, measured

783 procs on 11,492 rolling hits, 15 wearers, 15 runs, in the main hand or the off hand: all hits 2.04 [1.91, 2.19]. The
rate is all hits, which the one-hand white rate agrees with. One-hand white: 2.28 [2.04, 2.45], 81 procs on 1,067 lone
units, 3 wearers. Dual-wield white: 1.90 [1.47, 2.21], 161 procs on 2,539 lone units, 12 wearers. Specials: 1.63 [1.39,
1.79], 44 procs on 810 lone units, 5 wearers. Two-roll abilities, priced as one roll: 2.61, 2 procs on 23 lone units.
Travel time: the proc lands a median 163 ms after its hit, 98th percentile 249 ms, window [72, 259] ms (36 isolated
hits). Reports from 2026-11-10 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       7 |   467 | 7,279 | 1.92 |
| Rogue   |       6 |   300 | 4,018 | 2.24 |
| Hunter  |       2 |    16 |   196 | 2.45 |

Reports and fights: 1QpaxjXWFDA7vPV3 1-3; 6rgDRnVFX9a3pbGy 1-4; 7awbvWHZXYF8Kxgk 1-5; 8WgnxAakbzDZdFBX 2-6;
AwYnpMhJDxPFavjr 1-10; Fy81NHrkKD3W97qw 1-2; HJxqkQPAnmKtwNjd 1-16, 18, 20-24, 26-31; JgL3qFZA4aQcpXPC 1, 3-9, 11-13;
KjLXqdCZANvQc9BT 1-3; Lcy3V1wHPpqZQAG7 2, 4-12; PJGwRWQYyA3nzXpv 1-12, 14-19; h9pRvCYG8fzNDrcX 4-11; k6XaY3Dy4r7jfTK9
1-3; m9JVLBndtZ63vhDW 1-2; mBRL3hY1zjaKqkX4 1, 3, 5, 7-25.

## Hammer of Expertise (item 8708, spell 1317432): 1.85, observed

539 procs on 8,403 rolling hits, 7 wearers, 7 runs, in the main hand: all hits 1.83 [1.72, 1.95]. The rate is all hits,
which the one-hand white rate and the specials rate agree with. Observed: every wearer is a warrior. One-hand white:
2.04 [1.94, 2.14], 149 procs on 2,085 lone units, 5 wearers. Dual-wield white: 1.85 [1.49, 2.70], 37 procs on 571 lone
units, 2 wearers. Specials: 1.90 [1.37, 2.31], 70 procs on 1,050 lone units, 7 wearers. Two-roll abilities, priced as
one roll: 3.57, 11 procs on 88 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 186 isolated
hits. Reports from 2026-11-14 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       7 |   539 | 8,403 | 1.83 |

Reports and fights: 89yPg6z4AjZGfWRm 1-12; BQnXb3hTC4rmJ2pd 1-2; J97GvmrRXCFZfQ8a 1-2; JtvQWAmT7HCDV8h4 4-15;
ZgTqAy2xaHXL8djt 1-3; dGXqkVhwT17nHfBc 1-2; ybFmNGwXcnBQW1rz 1-4.

## Hanzo Sword (item 8190, spell 16405): 1, observed

492 procs on 19,895 rolling hits, 17 wearers, 20 runs, in the main hand or the off hand: all hits 0.99 [0.91, 1.07].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: 0.75, 10 procs on 530 lone units, 1 wearer. Dual-wield white:
0.98 [0.87, 1.12], 231 procs on 9,417 lone units, 16 wearers. Specials: 0.33, 1 proc on 123 lone units, 1 wearer.
Direct: window [-10, 100] ms, median 0 ms, 98th percentile 0 ms, 164 isolated hits. Reports from 2026-11-14 to
2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      11 |   309 | 12,384 | 1.00 |
| Warrior |       3 |   103 |  4,180 | 0.99 |
| Hunter  |       3 |    80 |  3,331 | 0.96 |

Reports and fights: 4f2ygdYCzFAvqc6K 6-8, 10, 12-14, 16-18; 6xHzpKd8wCna37Lb 1-14; 824kNhdFaALMBnvR 1-4, 7-18, 20-45;
DNc6QALhxmGbCH3Y 1-6; GHK2gB78FMVCmPfW 1-9; J2ThDq8RCPKNZdrB 7-11; KPqgbZxhn6BXwCvJ 1; MK6XJDzfLpFjnRWc 1-5;
bZaY8JyWxpmwQDjB 1; hwrmtk6aA7zJ8N1Q 1; jn3dwzWPZ6r2ky1K 1-3; kaPpXgt1bWJBr4z2 1-3; mtDjVd6a9rTgWBnC 1; pRA73YvkarFVLNCZ
1-3; qNyQvka41YhVmdHR 4; qbtZdThvBCXVcfAW 1-6; rTDcQk4gGxfmjN1q 1-7; x4TPAZXnW8kbv19j 1-6, 8; yPVkKpqtfJM27WFB 6-7;
zqpJDYHZC4Pa76bV 1-2.

## Iron Knuckles (item 2942, spell 13491): 2, measured

373 procs on 6,508 rolling hits, 7 wearers, 8 runs, in the main hand or the off hand: all hits 2.02 [1.91, 2.12]. The
rate is all hits, which the one-hand white rate agrees with. One-hand white: 2.14 [1.87, 2.26], 150 procs on 2,475 lone
units, 3 wearers. Dual-wield white: 0.95 [0.59, 1.56], 23 procs on 858 lone units, 4 wearers. Specials: 1.78 [1.39,
2.27], 33 procs on 655 lone units, 4 wearers. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 0 ms, 140
isolated hits. Reports from 2026-11-10 to 2026-11-17.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       3 |   289 | 5,002 | 2.04 |
| Rogue   |       4 |    84 | 1,506 | 1.97 |

Reports and fights: 38D6JN4nZzjcGtHg 1-4; 4QAtNfTXMyDGPqCr 1-13; AzbN8GY9DnBrH4aW 19-31; D8vykhYFL14KH7xX 1-3;
DwWM9H3rpTnGPfbt 1-14; GRTmQn3VDpwZKbCA 1-6, 8; H3vdcCr6JX1gwYtq 1-3; ZPyMjD1FkVGJH8Lb 1-4.

## Ironforge Greathammer (item 279895, spell 1322295): 2.9, measured

12,165 procs on 81,815 rolling hits, 227 wearers, 259 runs, in the main hand: all hits 2.88 [2.82, 2.93]. No lone-unit
rate applies, since its travel window passes 400 ms: the rate is all hits. One-hand white: 2.85 [2.73, 2.98], 2,475
procs on 16,796 lone units, 182 wearers. Specials: 2.92 [2.71, 3.12], 624 procs on 4,136 lone units, 47 wearers.
Two-roll abilities, priced as one roll: 4.58, 9 procs on 38 lone units. Travel time: the proc lands a median 137 ms
after its hit, 98th percentile 601 ms, window [32, 611] ms (2,202 isolated hits). Left out: 180 of 440 runs in Cat or
Bear Form. Reports from 2026-11-05 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |     178 | 8,146 | 54,859 | 2.87 |
| Warrior |      47 | 3,992 | 26,709 | 2.89 |
| Druid   |       5 |    27 |    247 | 2.12 |

Reports and fights: 17jHXDAWBRQdgqV6 1-4; 1HK7gFbwy8aJPAh6 1-3; 1fpaDyG3mFnB9KW8 1; 1h9P4rYzWDvtHRxb 1; 1pDcTvkrCBhZJRgf
1-14; 1rH7mXJAxcqkRGt2 1-4; 21j6cGMKRFLpzqrf 1-8; 2FytXTANpbh6dRx1 1-3; 2HahmvNTK89f3qxQ 5-9; 2gjR3tVq4Zdw7kzK 1-3;
2kFjvxn3MzRLQWTD 4-10; 2mW9A7wQ4bhv1YMF 1-3; 3YRkhNP8BZbtDWgx 1-8; 4279fmhA1HPGpDcX 1-2; 48QbknMZw2xt1R79 1-4;
63KPqwWGkh29Abmj 1; 6A24xBzrQkqRFpKh 1-9; 6GcDXN8kWnb4P9mh 1-3; 6HTpWDrd8LKGXwjf 1-5; 6NdR8TAqXMyLmkYg 1-3;
6Y8AjXnH4daWgKym 1-6; 7WLGVBNDXqcar6Ad 1-2; 7kYG3K21FmyMVABD 1-4; 7qTnPC4NDM6Ypc2W 1-3; 7tahbgBF1jQGRKMq 1-3;
7zK9fMmcNVxnLQA2 1-3; 7zcat1AyFnrJ4BQR 1-19; 8134cDXJahtPZB9W 1; 8PD6Lc3j2qQ91RC7 1; 8Z1TNzMHLCXyGj7k 1-2;
8c1GqjXaPHTpZthY 1-2; 8c7DafNhKtndMBzm 1-3; 8xcD6XhLdNrGq3bP 1; 91YHvCrBN7atKcAq 1-4, 6; A4dNxBP7bY2XkMnc 1-3;
ACvNpyX4DafTgcZw 4-18; AGLkhzpCJdVnPDF1 1-3; AL1DxTaHGfVygm3p 6-15; AQJdTrDPn8qyK4Zt 1-6; AfntVm7bWdMvrP1w 1;
B8AKvLZ6bQpaXktV 15-17; Bbd2CDQZyrVv7gFw 8-10; BrTM6CbayRvXgmc7 1-3; ByrG7LFC9mH1N8gD 4-6; C1fWAmVtPFBGK6ky 1-3;
CD1QAmVqHrLyY2Fc 1-8; CQ1hBPYpTF9jZyak 1-2; CXqJYaRfBjnMFxD6 1-15; CYNKBypF2g17j6tA 10-14; CdH9vgqPbRX1JKM3 1;
Cn6rjJpgNBDP7Vxa 1; DmyKc6zWJ7MwHdNR 1-4; DrCRwbFGX3dmtHVW 1-5; FHpqmBdPGYZ3JaDy 1-5; FTw7k9hdbYKVfH8C 1-5;
FVPg91ybaXmGtAYp 2; FkNW7XzKjdVxbqfJ 4-21; FqVXJryknZYWTvKH 1-2; G3KCZD9NbzVpH1Xk 1-4; GDpvwrBfZ2yYCxT1 1;
GVqw3aXB9g7tnZ8W 1-8; GaynFPTvZf4Xq2b3 1-2; GdDn73hPYt2JQmHL 1-5, 7-8, 10-12; GwtNT4JLhd8AVWkD 1-2; Gyw6BrhCDx42qcHT
1-4; HBXA4RjC91yxbNQV 1-2; HKQW3wCmr4kJXz2j 1-3; J4LX7AMRmxVh96P1 1; J8YVraRnGTbtyq6L 1-2; JPDfnQGCH9Xwcbv8 1-4;
JnWQzq1m82fjg6Vk 1-6; KGgyJXnVzvMxc28h 1-4; KHqR9VAYTzZ24wxJ 1-11; KVh6PxG3fHTDWw8A 1-3; KYDWXztmZHQNCVkF 4-5, 7;
KtzqRMDGk27NwYyp 1-3; LJ8NCMHzpym4dVqk 5-10; LPQaWZVAkzpbJGFw 1-5; LVwmRx1Mqak7vp4y 1-8; LmhQHWJDvcV1RdPj 1-4;
LthNmKqbQFaAZ2JM 1; MBR3dy4azD1L78Qq 1-3; MQjCX4kG28BxztgT 1-4; MWqZ1xaVht9TkXJA 1-6; MhfdwPABCaxn9b82 1;
NFWnKbtVw8TqCcZp 1-4; NH8YCdhDgtXG6bZK 1-4; NJhx3D4knajQTYcP 1; NLbKRCyqrYPJdmFg 1-6; NRH63zjvtwpxn1Wg 1-3;
NfK46gDLj8wrJz39 1-3; NngHmf2D46Gw7bcL 1-3; PBkXcr2JGgx4NR9Q 5-8; PJVn4jX8WA1TQzg6 1-3; PK4Z2NcFnQdpLwYg 1-4, 6-7;
PN7jfn93M6HKwpGX 1-3; Pck4Hb83WgXxrfJF 1-3; PrQ4yZdgFbGTcpWv 1-4; QFCfmhc2yt4zR168 1-10; QMH23Y1hqCxamFLj 1-8;
QmCHxRYD9tbjLfW1 1-4; QnNkg8ATL4Dbv1hf 1-4; R76N48HMyW9V3ZYc 1-5; RGw7haYADHg8yC2m 1-3; RHDbqGV3fgjMrLCX 1-9;
RHYb2XPvpDdf17LA 1-5; RKDTjJY4mNc9wbVk 1-5; TAwnY8WFB4HmQZ9X 1; TBYt3rwJqhGxD1py 1-3; TGvNBcgzCQYD7hya 1-12;
TQ4PXCqWA3KGYzh2 1-5; V1L8nj6gMpqRB4WN 1; VkPnGcCgJ17b2YvH 1-4; VxcH4X3W8da1jYLM 1-4; VyjQ1bDHvPZFdxcN 1-10;
W3wGpxR4tQFcak9K 6-9; WnjRtrhBVyYwXx3M 1-4; XR78naNGyAmZhMKB 1-5; XWgjPGZD3ryACTv8 1-3; Y3cP8mqfA7F9RjKk 1-4;
YC214LNaFPwkRrfm 1-12; YFBpbDdgxG17CXhL 1-4; YM7tXFPkLZhTgj6z 1, 3-7; YPaFmdJbpXt1W8ZN 1-4; Ynw8cNZv4GJRBC1L 1;
YqnVZLX68FRmctKQ 1-25; YxKjwhRDNyJ4z9cF 1-7; ZADc3bTCKhn9jJdz 1-3; ZPBTJjpyxndNg1Lm 1-2; Zfwc2Q1bKRpMnWGg 1-7;
ZmQq1dyC8VJRcpAM 1-4; ZwBW1aVvkDAfcN3z 1-4; aD6nv3CRjrFqXfd2 1-3; aMVzJNgDwpbG143A 1-3; adR7kpKcMDmtHThG 1-8;
agztmnCXVNRGqwvx 1-4; anKHZjtTkdVv239J 1-4; arf1LxvVJGycRFAm 1-2; b2rJTDtLjC7BxgGv 1-3, 5-6; b6zYpvJBq2CVLyfj 4;
b9xGzmYPRWNCQcXr 1-3; bDrG4Vpyc9amg1KR 1-3; bTFC3LZRKaHgq8mJ 1-13; bVNRgda481CqpmfG 1-3; brAQWYqMfywd1JPt 1-3;
c7zvhTJGCLaQPZKp 1-5; cH3w7AN4VZXTfr6W 1-3; cT941jZhxvWM2aKd 5, 7-8; cnz1wmVkXtgGJjZA 1-5; dAgmNB1rwGy8f76Z 1-2;
dM46X2j1BZGhYTmx 1-2; dQqfmPcwk4avZ6Gg 3-9; dafrWBRmTJpnPVYv 1-5; dyTRnv6MZQ7BKaGC 1-4; dzBLDHJTfc13yAWQ 12-27;
fCbHhaK7jxv9FkYq 1-2; fDycAFZWt4NkRX1x 1-3; fTHKyQa6YFVj7JWL 1; fcCgK1R6GpBz8tPX 1-4; fjac46MdLgvHprZq 1;
fpGWB7am4PzYnMyN 1-3; fzyDKTP7x8B2hRNw 1; g9xAZ7jBQmRN6TGn 1-5; gNMJVXAKF7hH6WQx 3-5; gPbhJ73prMtxXcjn 1, 3-7;
gkFbNhdxDc384Pyp 5-6; h28ZpnT3kCm6NHMb 1-3; h2CgmZyNBarVjtYP 1, 3-5; h9L8pBqnVdJ1vDrb 1-12; hBWtTfC1rk8nMQmX 1;
hM67TLJGypgwvrzP 1; hVFtnpRGby9azXg8 1-2, 4-26, 28-42; hdtzWJKAZ8cHpGRx 1-3; hfJZb8HL41KmtdMA 1-4; hjkVZTFrGbRv7cH8 1;
hwcWQn8VLN7kJxmT 1; j1CGayX43bMdF8hP 1-3; j23K74gP1LNyYz6d 1, 3-6, 8-17; j3vLB41G8pkctPMH 1-2; jQP1aWnwH8C2T7Lt 1-4;
jpF1ML4ZTkcwGQ8W 1-3; jvtkpbcQCYFz83mT 1-3; k2rGgh6TVDXzc934 1-2, 4; kY3DG8tzBr2vX1Rm 1-2; kczT3CJVBXa9Gj84 1-16;
kyLDgPjrQYdHAc1a 1-2, 4; kzpRQ2gqGJ3yKh9Y 1-4; m9A2MvGDfPFRCWj6 1; n4xHq7mpR2htyXBr 1, 3-11; nFHZX8TMVG61PRpW 146-149;
nWhPF7JrzNRLmx2v 1-5; nadqr4JWQ82ZTkv1 1-3; pG2cZKTAjfYzv8yX 1-10; pJN1cYKwfgaZ8jt4 1-5; pJQ4mtrDqcaAzwBP 1-4;
pRrGZhvDw83LJHTy 2-4, 6; pT6jGtaZFDr92K7R 1; pqBwzhZQy7d9jFnY 1-8; pqPBFkXgnjDvN9ZG 1; pwNxagqK7HBZ1b9X 1-5;
q2h6pRvdBCKMabnL 1-2; q6bKHQCLRJaWTwGy 1-2; qCNzdMA63yLWmhGk 1-3; qFbMHLVtz94cXnCy 1-4; qNBhWJka9TtLPwV2 1-4;
qh8Lgfy31XnQF4jH 1-3; qxY87bFJjG6Tz14A 1-3; qyp6QA92RLJNKxTd 1-14; qzBa8HcG7xAfPyT3 1-14; r2c7bMKxfd3AHVYT 1-4;
r7RQ8FJpnBVgAy6f 1; r8paZAQhXF24Wdw6 1-2; rC3VbJjLQ6YBp8vM 1-3; rL2DzHGy7t4wxh3N 1-3; rLmJfNtq7v6BFcxy 1-2;
rQdnmBHXCpKgJqw4 1; rQt1k2LGcwVfFad7 11-14; rcYRpQA1aZVhPz9B 1-3; t9DRvJmjap3BrcCw 1-4; tH8kRGA3bph7V9vN 1-3;
tNxJXzB3q7rYfRK8 1-5; tWV1d42zgK7H3qNM 1-4; tZK2jg7HqTGAWnRd 1-3; tbQBwaTD7R2rxpkP 1-3; tgqAx3jr6vGW8nck 1-2;
thJy8Y43TPRnHvaK 1-5; v3WBbzYNjaqQJ67c 1-5; vCD8yMjK4WL1ZnqY 1-2; vRmX3tw8dbkx1HTV 4-8; vVCHXQgNFhqMnbjp 1;
vXnHY1xFwVzDBa2k 1-4; vf3QxCLyhckPdbM2 1-5; vk4gw8GBahr6VMxb 1-2; vm7QzHZCjFGtJ6DM 1-2; w3jgGmAkt7Rfnxah 1-3;
w4JfLXqF7HDK32ZV 1-3; wH9M8XGJN1K4h3tk 1-7; wTYjbBFDZN2zkLyV 8; wfMAD38qZRybWmn6 7-9; wnJHrLF14XD6d3RT 6-8;
wtzQKj8GP4cJ2W6H 1; yQ624zGVT18w9nbd 1-10; yVATBnQjpz1MLZPh 1-2; ykcwLtjHFQT87Mv9 1-3; zfW2RHdjLqKA1yD8 1-5;
zr6Z1GFPkpwQVmX3 1-5; zyFZC9kAVDB3Mvj1 1-3.

## Killmaim (item 13016, spell 13318): 0.95, measured

872 procs on 17,458 rolling hits, 39 wearers, 43 runs, in the main hand: all hits 0.94 [0.87, 1.00]. The rate is all
hits, which the one-hand white rate agrees with. One-hand white: 0.82 [0.72, 0.92], 210 procs on 4,820 lone units, 37
wearers. Specials: 1.08 [0.82, 1.29], 34 procs on 590 lone units, 12 wearers. Two-roll abilities, priced as one roll:
0.96, 6 procs on 117 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 766 ms, 160 isolated hits.
Reports from 2026-11-13 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      28 |   607 | 11,657 | 0.98 |
| Warrior |      10 |   259 |  5,604 | 0.87 |
| Hunter  |       2 |     6 |    197 | 0.57 |

Reports and fights: 1g7hAY6CBNtW84HX 1-2; 29A3dKwPgpRNnWvc 7-11; 2aVnTxFzfckdXpBy 1-7; 3vDf6TMYBX9LK7bz 1-2;
3xgPkwfn8yQ9BqAW 1-4; 6hrcXFkgDamVB8jT 1-3; 6v4wTWQkjb3Xd7mM 6; 7hNgRwMfkQZ2Yv4b 1-3; 8GW16QY9BmFHJXLb 1;
8RFv6Xyr43dLpJ7h 1-2; DaXWGJMR2bnc9YtV 1-2; Dx2tAm7YawZjV6hG 1-4; G3VKdgR2aLAkNqWy 1-3; M4THD3Qq9dBpbKvg 1-2;
MrW6JCFZTDzdtYpH 1-4; P16dx3VLmRyaDJZQ 1-18; P7cV2HjLgxGpmZQk 1-7; PN79G4Lfv3KBrW8j 22-323, 325-361; QhYmjJ4bHZXcgfk1
1-2, 4-27; QqKcMJwHjxNPkgRT 1-3; T6MXt73janNJgmyY 1-3; Vvfz8wYbjTBFaxRq 1; XVJtpHkwZy4ancAQ 1-7; aQwNVCPHWmzDLXqk 1-2;
acHy1fNktP2m47pd 1-2; bdL8j9KGMR6r2xmk 1-3; d7DFtXL2jfNa8Bw4 1-5; fn2Dc941K63CNvmj 3-4; ftNBwvGgPcMqh8dJ 1-2;
gdKNPjVvcQt2LJnq 1-3; ghzjGKVBvJbRZnMr 1-2; grFA3BHxtCNyzR7v 1-2; hbW1zgMXyRv8JAVL 1-6; jKJPF16ZnWLkrR2d 1-2;
nT2Wz7mVQZbtBRG4 1-4; nyhMJLqBYxPpRaXf 1-2; pDRxq6wJynmhgdjL 1-9; pqr4HxC3vWJhfG7B 13-20; tMxrNWC6qPVzgwAv 7-8;
vJGT19YL26XFQpNB 1-2; zJ6C2148dTYfaXHb 2-4; zhj69Dt2X1b3fvMT 1.

## Loksey's Training Stick (item 7710, spell 1293482): 1.9, measured

3,292 procs on 33,406 rolling hits, 179 wearers, 222 runs, in the main hand: all hits 1.91 [1.85, 1.98]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.09 [1.97, 2.23], 1,001 procs on
9,253 lone units, 168 wearers. Specials: 1.97 [1.61, 2.21], 175 procs on 1,723 lone units, 93 wearers. Two-roll
abilities, priced as one roll: 4.18, 67 procs on 310 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th
percentile 1 ms, 1,078 isolated hits. Left out: 171 of 394 runs in Cat or Bear Form. Reports from 2026-11-09 to
2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      25 | 1,397 | 14,426 | 1.87 |
| Warrior |       8 |   981 |  9,950 | 1.91 |
| Hunter  |     144 |   913 |  9,024 | 1.96 |
| Warlock |       1 |     1 |      5 | 3.87 |
| Mage    |       3 |     0 |      1 | 0.00 |
| Priest  |       2 |     0 |      0 |    - |

Reports and fights: 1QXAJwhcYL6M7TDk 8-30, 32-33; 1aBhkzdDMZKGtb37 198 fights from 162 to 376; 1g2QAZPJrLTjCz3m 10,
12-14; 1hpF2AyzjRdrL78N 1, 3-12; 1xqBcj4CZQXV38Pv 1-2; 1yMxz28Pm4NZbAJf 1-12; 214ndcFyjpPM8wah 1-3, 6-9, 11-15, 19-28,
30; 27jZGhfHgp1LR8MB 1-11; 2PpAjf1qRmTv6gDz 1-17; 2WH4jLt3FADy8fnh 1-2, 5-6; 2aG3Ld8KVARCzrh1 1-7; 2vMwVxLqdB3GRZn6 1-4,
6-7; 39BFWKC2J8RTzyMx 1-3; 3HJrg1BGh8xR6Y2z 1-13; 3PXfhvmTp6NQdwRb 10-12; 3TgpnwxMD71qfLm4 1; 4ARKtCTVhrgBwadG 7-10;
4BhykDjm2p1ncLGK 10-11; 4JYa2mQgpVNMF89R 10-21, 23-30, 32-82, 84-103, 105-110; 4V2RLGtF9anbCKWD 1-23; 4wHnDaRqY9Z2zv1F
1-14, 16; 6NrFMaRQ9yAjxWC2 2, 4-12; 6WG74An9RrCdqkwD 3-8; 6wmLW1N4tBXpTVqy 1-3; 7Vywm9BAZgcMfbT3 14; 7d6p4TBxcjFyPNVr
1-4, 6; 7fgdkyv1njDKa8Gc 6; 7g9DpfNFBA3PatdZ 8-14; 7nCPmLqHd8jMfQk3 4-5, 7-10, 12-13; 7tL3AyRHfawBJF94 1-12;
8VQDf7JtwbTknMLz 6-8; 8X2Kc46faTnxVBLm 3; 8ZdnfAT79zPwmyVM 1, 3-6; 8frHF64GzAhBxDyb 13-14, 16-22; 8tgNprk9LJwW47MK 1-5;
91XtwNrLRzvMkQ62 1-3; 91ZzNbJMRtjrLW6K 1-2; 97FVRX1WLHqDmjAQ 5-12; 9TrdzatyJbQV3vmc 1-7; 9g7NQBqDH6FGfP8L 1-4;
9nxtzJ1bC42FrgWN 1-3; 9rPx2JNYLbCy1waX 11-15; A8NWyY7dhpLDt3zX 12-30, 32-38; A8tQY9p3Dzn6TLRJ 1; BTRMkxmvfrwcZNaG 3;
Bmk9chwYjd2F3zpZ 1, 3-6; BydpaqJVW29G8QAc 1-9; CMxBvHAN37YRqbGg 5-6; CPNY7nxkfXpvQgH1 1-3; CXvjmgaJ6YqKzD4k 1-2;
CpDKr6AWJ4fw7xhg 5-7; D6bxHwnpyvc2VNR7 1, 3-8, 10; DWJhjnmAM6GZdbKv 1-2; DcgdHfFZjb87qRJa 1-4, 6-9, 11-12;
FDybHvQ3pPqBracN 1-4; FQnMCW8kRqVv4mT2 8-13; FTXVdBrHJZD9aw7P 1-7, 9-23; FzGMchfD4PkdmJT1 1-4; G6XMnQyqxTdJa2mb 4-15,
17-18; GAdW4fkq7bx9BPmL 1; GBDAWjKNh6Q9XnF7 1-4; GJgyrn6t1vABDQaW 1-3; GKZJBtWAvfPL81bc 1; GrW7bVHhKwJgF4QT 7-9, 11-12;
Gv6NDzAmFjC4P3qL 1-15; H6h87LGmw3gKRc9f 1; HL9zDvk2wWTXcP3N 1-6; HQTANmjP8FB29vYW 1-2; Hp9ThMznR74PwYKb 5-8;
J1QrDzNc4ydjV8ak 1-10; JZv9XDFLC1yKxrPY 5-10; JcwBmYGVgnMQH9xd 14-19; JtrVj2cqL3znKNXh 1; KLNhgMTDAV6xY1Wj 1-3;
KapmH2bcWfPqnrY7 1-8; KmjAV91JBhgQLrfP 3-4; L6jJNDYaATph4cfK 1-3; LJMmvtPTzdCpG9k6 5-9; M4G1VJQdXbvyYLBD 1-2;
MVcaxDQwWLpnK6HJ 1, 3-9, 11-12; MgcQ6txHr9R1VjPv 1-5; MpjtZc2zqHaR9hC3 1-19; MyzjakDC9gBbpq6K 2, 4; MzQJ9fgkwjhW1Lv6
12-14; NHvp73cXPQRC8agz 7-11; NJ7vtRpCzDTfk2W8 4; NWvgFp3kfR2bHyYa 1-35; Nc2grh6ByLvJCkRY 4-6; NkTjKdbwYAtfJQBn 1-4;
NkXjqd43AFVYpzGK 5-8; NkyKCr2XmgLt1YWv 83-87; NzgdqKHnfLJWkFVM 1, 3-12; PAmg9FX8fr23xHt4 1-5; PDT2yACNzd4LMZ98 1-14;
PLGKvft3bq8WgRYN 1-2, 4; PQDRkypFY94rzKwv 1-4, 6-7; PWhMrYaD4Aw8HcCL 1, 3-7; PmR8Lpvazd4q61nA 6-12; Pn2LYHXpjZMDywgA
7-9; PqBvJ94Nhb7ax1Vz 4; Q1CHmPvpYaxM49qR 4; Q1arfj4WPBNXqmLk 13; Q2xTrDcZqjB9pA7F 9-18; Q8GxhTmH9f6ajcNr 6-10, 12-39,
41, 43-51; QDadybK3XVPRvwLh 30; QJafvkCx4yhVXBmT 1-2; QPmw1C9MNc4hqH3K 1-5; QYTW7Hza9nBjJFwP 1-3; QnjR1t9Y3WHpG6Vd 1-2;
Qq6CvRP12rAGwngY 1-2; QtB8qHZ4X2pyrDPY 1-10, 12-13; R6xvn3yVf4DGJpka 1-3; RNpVm3CcDJbxjB6k 11-21; RZhVjDGx1vLPkAag 1;
RmTM8XL4aBJY12Ak 11-16; T6F89dBtbNG3MjpZ 1-3; T9cJG8bq1HaNyLQC 1-3, 5-11; TQAKbWq7vVJHLgBR 6-18; TnLAmyRdXvz1cBhC 1-18;
TvMmZw8nc4hKjd1x 12-13; V39PRKTHt8kahfxX 1-6; V4HDKvxmYCkR1JnB 1-3; V4aAZ2PGgkHw6mvB 2-3; VF1b79fRAWhtGrwa 1-3;
VZdxq6yNBrW48ztG 1-3; Vb3ZxfXtyDWr1ndP 1; VmQK1zM4hFHBxXgJ 6-15; WYrCyJkdfMXw84Ph 13-21, 23-35, 37-39, 41-48;
Wf26RdmcnLDTXwV3 1-2; WgtPVTmY4XQ2DrGB 26-33; WqKd6nTV1GX8Qka4 5-10; Wr2mNV3McRqf1Zwg 1-6; YH3c4zvFrpLtR1M7 5-9;
YJR78AMPDNKGQWjF 12; YLHkP9qZWz2b4AJt 1-3; YZyAfq6XN9k4a8md 12-15; ZNQyKft4zAGCgcDY 1-8; ZjdqP6bg7tTR894y 1, 3-5;
a6QpkvcR1YHLBqPh 1-6; a8bQ4fNyGVnFcZAH 1-10, 12-25; aDKxMy2WYAfcnvGh 1-2; aGWKDzhkCNJFXZgt 1-5; aj46hBQX1AfRFHtm 1-15;
ajLmndyrzNJfwGCV 11-12; amTJhvjMAyCLdbXR 1-2; bCzt1Fq24mYcPXkB 1-2; bT3LyRBtax8FQnch 1; bvqJAwG8D4KycjhC 1-18;
byYP67TH4KtZqm31 11-13, 15-17; c2J4C8kzHaQxnwXW 1; cNhpPv28MHBJCwWm 1-2; crtFfPbdavN6QHnX 1-9; cwtKD6v8xmJG3jNY 1-20;
d37CTyKPbHaV1Yzc 1-15; dBPjxVMQmgZ9h2vr 10-26; dDKQgLW3MFPatynJ 1-25; dnvYkhpKZQaqM2Hw 1, 3-5, 7; f3HYhjJkZdtDgzM6 1-5;
fAXbZvgpmLFxWw4z 1-6; fd4MgwavrqWcYHZG 1-14; fjqQNyMbD1Y7R92h 16-21; fnYThW3BpxKZv7jw 1-18; ftkqY3cwR7dBZK1J 1, 3-5;
fzKwkZLcbYraJdF1 6-9, 11-16; g7Bh9F2b1kz6Cnar 1-2; g7RXrQTKBCNMxqJc 1; g8BWDmknXJpvFbdf 1-4; gAdCTkV1LBxNYFca 1-20,
22-24, 26-32; hBr7xLW6nkVdMKF2 1-2, 4, 6-7; hCzY42vqXrc1mG7w 1-4; hrFwMYgVGq41fNjQ 1-2; j2RNAQYzZK6DyT17 1;
j7V46tC2GkwNYTrz 1-2; jGL3RyT4fYMtwPrn 12-14; jGhpmnHrtyc1VD8W 1-4; jXLYwBW7bZ3nQdfJ 3-4; kH2bjrXh4BF39GmC 1-2;
kQhZqY4DwGBKLdag 1-4; mDbPCdX9AvwYJxGa 16-23; mH68YTDck3dKV4fx 1-2; mqjv1nLcRh6289yk 1-9; mqnZ7bkf3v86prGV 4-13;
n7qPAd4rN1JyHfpR 1-3, 5-9; nGctFJyB8jwX9WHL 1; nZ3NhbvpVXg649Bm 1-10; naCtKF4L3qzpGPmf 4-5; p1zYDag2NHXPyRt3 1-4;
p64RG9PVZcDg1XCk 4-6; pJ4AKDqYPvfwbTkM 5-8; pqfK2AgHZ7DnhvPF 1; pxAgrJDCXHc9tbB6 1-19; qGywkYrdHKT8NgAQ 1-2, 4-8;
qVtTPaDMHfhmG9gr 1-10; qj6gnCAM8kB2RNPy 1-2, 4; qzv6W9fkDaP1MZYH 1-6; r1xm2Tt3bjgBahZF 1-17; rD8m4JW1bVyYPKGX 1-3, 6;
rPAKqpWdFTR6atmN 1-2, 4; tBQwPjhFqYAyCv1L 1-16; tHz9M3arKp6PxA1W 1-14; tLkym2AVWdJ7bM9H 1-8; tXFAxTLmg3bBNZrv 12-13;
taFcCyAx9BD1KWXp 1-16; tkQKL9bGT6JVR3na 5-7, 9-16; twRbTJHfrDXBLj97 1-5; v9Hf4bPtGq1LVykC 1-5, 7-8; vfFGXn4mbLAqkhWR
14-17, 19-20; vnrGwA8Z2RYtPc9g 1-4; wXNczvgHAqT7L8nV 1-13; xQ2J8AbFN19pPD6X 1-15; xRTpHyY7qC1GmghM 1-3; xYTCXa9myPJr6R8M
1-13; xk367aQTJVzfrKXq 1; xnvgBCHkJf4Y19rV 1-4; xydMjQBYGHW91vTL 1, 3-6; yKvCQdzpgqY1WwbA 1-7; ythJ7TZDXk6AjHMq 9-10;
yzFXjCY3WVbgRPx9 1-2; zjgkHTcZDJNv13bF 40-85, 87, 89-98, 100-104, 106-158, 160-166, 168-174.

## Lord General's Sword (item 11817, spell 15602): 2, measured

7,168 procs on 83,360 rolling hits, 25 wearers, 39 runs, in the main hand: all hits 1.98 [1.90, 2.04]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.94 [1.39, 2.32], 167 procs on
1,991 lone units, 4 wearers. Dual-wield white: 1.52 [1.29, 1.69], 876 procs on 13,322 lone units, 22 wearers. Specials:
2.04 [1.88, 2.16], 1,081 procs on 12,234 lone units, 25 wearers. Two-roll abilities, priced as one roll: 2.87, 113 procs
on 908 lone units. Direct: window [-20, 100] ms, median 0 ms, 98th percentile 0 ms, 1,488 isolated hits. Reports from
2026-11-13 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      12 | 4,716 | 53,845 | 2.02 |
| Warrior |      11 | 2,173 | 26,303 | 1.91 |
| Hunter  |       2 |   279 |  3,212 | 2.00 |

Reports and fights: 24kgnZAXbKTW89x6 1-2; 26AYbFNQzMZCmfkh 1-16; 6c8P4BLaVAj3bpK7 9; 7TXcFNRzjwHdAybn 3-19;
7dgfrZ3bvhVqxtY8 3-14; 7tbWfYqJFd4vzL3p 3-6; ARFXzbqHDpBLJdhy 1, 3-4; CDKVcP6ghZJtv8AW 3-6, 8-16; D9frqPV8dam7B1AM
35-50, 53-55, 57, 59-61; HwyJMCLKfbY1PkmG 1-8, 10-38; JfNv1pWML9ZA7ghR 1-4; K6VvbDypXFLQJW7d 1-27, 29; MgC1BXh8fyArkYqa
1-30; MwVH1KQmak73NPd8 8-16; NK6cCM4pJyRhaAFW 1-7; NXMZV9xjvWpB7GtP 10, 12-107, 109-116, 118-122; NcxbzW4kJYQ7DGhj 1-3;
QFqKVZgc6vAJfbrW 26-28, 30-31; RVkGfcrbZ13MWqQ2 1-3; T92NqcgBXaZrKvCm 1-5; TLp1jza6PCrBAX94 1-7; Wn8wHhaxy31MLzjp 1-10;
X8z9dgD6HRTZ24Vc 23; XaF296wZvArq7Qxh 2-6, 8-19, 21; fCB4wQt3NWm6k8M9 5-7; gvWZG1cjrChKwdtz 1-2; hF7dxZ8nJDpk4wNG 1,
3-5; jn3dwzWPZ6r2ky1K 1-3; pPbM8VAdmWCTvLB3 1; pQPwFg8rf6VWhRNc 1-3; pgN9QwMLYVfrqGbv 1-4; q21PmQwWRfMLvKXc 1;
qbtZdThvBCXVcfAW 1-6; rfvtaC13gNqRx4Az 1-3; v1bPdgDtXypHncfN 1-4; yPVkKpqtfJM27WFB 6-7; yjV6m7kpfgaPzXxt 5-9;
zKgGQA6BqvcJtYZ9 1-3.

## Nightblade (item 1982, spell 18211): 2.1, observed

86 procs on 705 rolling hits, 2 wearers, 2 runs, in the main hand: all hits 2.09 [2.02, 2.11]. With 2 wearers the wearer
bootstrap means little; the Poisson interval on the 86 procs is [1.67, 2.58]. Neither the one-hand white swings nor the
specials reach 50 procs, so no lone-unit rate checks the all-hits rate. Observed: fewer than 5 wearers. One-hand white:
2.17 [1.77, 3.96], 18 procs on 142 lone units, 2 wearers. Specials: 1.73 [0.00, 1.84], 12 procs on 119 lone units, 2
wearers. Two-roll abilities, priced as one roll: 5.71, 1 proc on 3 lone units. Travel time: the proc lands a median 137
ms after its hit, 98th percentile 254 ms, window [41, 264] ms (34 isolated hits). Reports from 2026-11-14 to 2026-11-15.

| Class   | Wearers | Procs | Hits |  PPM |
| ------- | ------: | ----: | ---: | ---: |
| Hunter  |       1 |    72 |  586 | 2.11 |
| Warrior |       1 |    14 |  119 | 2.02 |

Reports and fights: nNmBkKp3J96FTGfg 1; vdwBDjmZRykqHaWL 7.

## Pendulum of Doom (item 9425, spell 10373): 0.9, measured

641 procs on 10,733 rolling hits, 8 wearers, 10 runs, in the main hand: all hits 0.90 [0.83, 0.93]. The rate is all
hits, which the one-hand white rate agrees with. One-hand white: 1.02 [0.81, 1.16], 158 procs on 2,333 lone units, 8
wearers. Specials: 1.00 [0.60, 1.10], 37 procs on 553 lone units, 3 wearers. Two-roll abilities, priced as one roll:
2.34, 12 procs on 77 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 178 isolated hits.
Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       3 |   393 | 6,633 | 0.89 |
| Shaman  |       5 |   248 | 4,100 | 0.91 |

Reports and fights: 9pMcvgxrQNfLZ7hd 1-15; AdD17hyxFtqfrQmC 1-4; Pjz8kK2XvxyQrTCB 4-16; WBfhVMYkvLHmPQjx 6;
X186Gx2gRBZChbVA 1-7; ga9Qk6DcjJmNTGVH 2-3; qw3ZzkBmQYW7pXVH 1-19; t6dr1MY8zTJBWnkg 1-5, 7; wTakYgvxcBQV9qRr 1-19;
zc1hygaCXjDAPxL6 1-2.

## Phantom Blade (item 7961, spell 9806): 2.05, observed

780 procs on 14,264 rolling hits, 16 wearers, 17 runs, in the main hand or the off hand: all hits 2.05 [1.88, 2.23].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: 2.24, 25 procs on 419 lone units, 1 wearer. Dual-wield white:
1.97 [1.75, 2.19], 284 procs on 5,413 lone units, 15 wearers. Specials: 2.50, 8 procs on 120 lone units, 1 wearer.
Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 219 isolated hits. Reports from 2026-11-15 to
2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      12 |   541 | 10,224 | 1.98 |
| Hunter  |       3 |   170 |  2,792 | 2.28 |
| Warrior |       1 |    69 |  1,248 | 2.07 |

Reports and fights: 2YfdbFMgGhr7zHwQ 3-4; 6kN78G2xfqy3vwTP 1-6; 8L9G1c3bKt6BAxJW 1-10, 12; Ap2fHzmtJLhv83DN 2-15;
D2d3hNAbwkPgL1vF 1-10; GzZH1WhTLQ3JaDXt 1-11; QAHXb9zy4jfgKPVJ 1-6; XFdcWaPL8MthYH6A 1-5; cRtHw9DWZMrKnyvQ 1-4;
dt4M3ypRnTkwaVQY 1, 3-10; hrV3QgyzvfdZKJwF 1-30; kQ2gDVr4h7NPm1pa 1; pLtJjx8DCPbHRacv 1-21; twcjmZNGMQP2FABR 1-12;
y2dFcNR1rMmKTPAj 1-25; ynGzYZ3cLrpvwj1N 1-9; yqFPH1MKv746dQhB 4-7.

## Plaguefang (item 279876, spell 1309315): 2, measured

2,944 procs on 42,419 rolling hits, 55 wearers, 63 runs, in the main hand or the off hand: all hits 1.98 [1.87, 2.11].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.99 [1.82, 2.13],
560 procs on 8,058 lone units, 14 wearers. Dual-wield white: 1.70 [1.55, 1.82], 942 procs on 15,817 lone units, 40
wearers. Specials: 1.98 [1.69, 2.23], 324 procs on 4,674 lone units, 23 wearers. Two-roll abilities, priced as one roll:
0.00, 0 procs on 1 lone unit. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 1,317 isolated hits.
Reports from 2026-11-05 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      29 | 1,348 | 20,892 | 1.84 |
| Warrior |      19 | 1,530 | 20,519 | 2.13 |
| Hunter  |       7 |    66 |  1,008 | 1.87 |

Reports and fights: 149bNrXGaKgQL38P 1, 3; 1Hgm9RqadvcjJh2w 1; 1HpaBdm9fFXDjYPy 1-4; 1XCND7YQJ86Ax32q 1, 3, 5;
1ZJVMgcYWTyRQh2w 1-3; 3tPjXKNACMHcn9b7 1-2; 4ChPnBrVK2fJLMaX 4; 4hNWVC2Kzj8RxpZL 1, 3-4; 6TQkbpNXgGZwYat4 1-3;
6tANqvfKQP1Fkdzg 1-2; 7DKRq39nf4yvFWLG 9-14; 7NbYtg6T2pjr1yL8 1-2; 8mBLvfVzCn4tWyTM 1-3; A4NykGFTwmXKj73Y 1-2;
ARBxdp9znVDMTwJ4 1-4; AZLYpdc6r7P2CMRm 1-5; AgFRCzn6X2KDdtMr 1-11, 13-18; BgCQ6Xzp3Nh2HWkP 1, 4; Bxk8b1hKMz6vt9Vd 1;
CJZhDp9FxwRMaWYy 1-2; CdXV4pYTKPwAqBmN 1-5; CfhkbPBKAZFRGp9g 2-3; FcC4vd2DkRqyMrVT 1-2; GAK2BMH8Wt3k9JFj 1-5;
HPRbL3N7p6wFhrqk 1-3; HqQTwzArBZpPN1y3 1-12; JDH7bFz23Qakdgnh 1, 3; KAvPryL1zCdtWTxV 2, 4; KWZaQjqPLJ7Ry8Gd 1-2;
KYD67nVqZT4jxWyN 1-6; L7QHtkhacypgd3D1 1-2; MzwcDrG3fdFBRy6a 3-5; PqjRDCnyaHt13QJb 1-2, 4-5, 7-8, 10-11;
Qb7TJqYkKtgyz1ah 1; RB47hLrMjdDG8Z6a 1; Rp4QCr31TVmgyqnz 1-2; TH63PWxNv9gRarCY 1-2; TnbVQWRJ7r26CqpK 1, 3;
W8YC9TA6bcXwLRMk 1-6; Z4GKmBf1HXdtkYDL 1-5; a8cL9vKhNmkQfBGd 1-3; aVyXCj1NpnRTFKxt 1-2, 4; afNcyKHdtY2qjPCp 1-4, 6;
bcATfGPv3DX1Jnxy 1-2; d7DFtXL2jfNa8Bw4 1-5; dzTmC6F7xwLPJrq9 5-17; fTpYLw71QVR2qky8 2-13; gKmcRvAdBYqrDnbt 1-3;
jKfkzadV3BLp1cbr 3; kLB1K4faPvhmGzqV 1-5; m2FqGpwzr9cQf7PK 1; n6DtHVNKp1wgymFL 1-2; nWxftVpvMycmw2FL 1-4;
q8DXKyGLJFb1CwVp 10-13; qtcFwCBmkXz7JMH8 21-22; rKXWhHqLza39gkmA 1-2; tm2FQVrcZzfB97v8 1-2; vYQqkXLm43RJZGy7 1-3;
x9vCJk6RYqFLhWGa 2-5; xXtbFrdR4qKAQVB1 7-8; yLKXxQgrPWTnmdZA 1-3; yMNq3krVBAHxG68w 1-3; yTDCtLmVK4kq7gQ8 1-5.

## Sang'thraze the Deflector (item 9379, spell 1294243): 2.05, measured

2,246 procs on 38,451 rolling hits, 29 wearers, 33 runs, in the main hand or the off hand: all hits 2.06 [1.97, 2.17].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.03 [1.89, 2.20],
653 procs on 11,343 lone units, 17 wearers. Dual-wield white: 1.69 [1.43, 1.90], 175 procs on 3,649 lone units, 12
wearers. Specials: 2.13 [1.91, 2.37], 212 procs on 3,521 lone units, 18 wearers. Direct: window [-19, 100] ms, median 0
ms, 98th percentile 0 ms, 723 isolated hits. Reports from 2026-11-10 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      18 | 1,898 | 32,465 | 2.06 |
| Rogue   |       8 |   293 |  5,011 | 2.06 |
| Hunter  |       3 |    55 |    975 | 1.99 |

Reports and fights: 1CYAWTtmcGdFqbPN 1-5; 28AhM3VGnvFcJ4rD 1-4; 4NqLGxvrZYX2Qwnf 1-4; 6bM1k938LGNghacx 1-9;
8FBqfvLyWTdKxbZh 1-38; 8KPHjdAxZYFNLtgR 1-5; 987q1pjnYdVQbCmT 6; Axv7w1D9JW8ykLhY 1-9, 11; C3zW7BjkJXgRtq8m 1-11, 13-18;
FAL4Hv1D7nWf8VNh 1, 3-8; FHmqabcjK2RAM9vZ 1-3; FVv7P3a2KqJrdG4g 1-3; JDZFAK1z7QbckdHN 1-3; KfvWRDrFXQ1TMq9P 1-8;
LXWzFKcBmvyMZaH2 1-2, 4-9; NbwyAxHKzhYXV6L2 1; VQCRJWfq3KX7M1Nh 1-2; XCJdxAF17zqHMhLY 1-2; Xj23FkTPgrHawv7J 1;
ZLpHRCgGFDBk9TPf 1-16; ZLpcV6krn1wYHJvt 1-8; abWvgzBMf2rJXhcL 1-2; bXza3gZq7nRTGVfW 1-4; brf4Pzj8BkyMmCHJ 1;
cPLzXv4nTtNaYWF3 4-20; hgmcV2xfkQFbWMN8 1-3; qVJHr3mzbkAvxgyd 1; tCprA1RmvTcJkq3D 1-4; tnyh4dxq6pNXQzLw 1-3;
w63W4VhTgCbt9knB 1, 3-10; yQnph9W1VB4z6a2J 1; z2QFnM7tAjJmqgvw 1-4; zLacx86Thr3AMtqn 1-15.

## Satyr's Lash (item 17752, spell 1292155): 1.95, observed

1,105 procs on 19,794 rolling hits, 27 wearers, 28 runs, in the main hand or the off hand: all hits 1.97 [1.80, 2.12].
No lone-unit rate applies, since its travel reaches past the isolation: the rate is all hits. Observed: no lone-unit
rate checks it. One-hand white: none. Specials: none. Travel time: the proc lands a median 133 ms after its hit, 98th
percentile 265 ms, which reaches the 300 ms isolation, so no window was measured (239 isolated procs). Reports from
2026-11-13 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      20 |   854 | 15,466 | 1.95 |
| Warrior |       2 |   203 |  3,500 | 2.05 |
| Hunter  |       5 |    48 |    828 | 2.05 |

Reports and fights: 4NqLGxvrZYX2Qwnf 5-8; 4PtKGL1zcRV8Chfn 3; 4YzXaGtg81WdqM2N 7-15; 6zJNTjMwWD31vrgy 1;
96tBVjg8RYJNMa7r 1; Bfk6ChZgAr8yWtXv 1-7; C8Qd1LXMrGT62kpY 8-9; L2c7jBZNwFPpb1tV 1-4; Lg82ZFDyQ4MvRm9p 14-26;
MKmntyr82Q4v7jDh 1-5; R1acrQhyNvwXLmBj 1-3; Rc4dCWA6N8PwpBvT 1-8, 10-11; T71XnfaF3cY9M2Rg 1-13; ThKD4jr8MfawcCPv 1-19;
V7Xna6bcWLRzZ3GN 1-6, 8-11; W1zqgCZVPAtTM27b 1-5; YamkzNdPTZFLAnGq 1-8; Yf91cBLN26jdqrxC 1-4; ZwxX9G3bRPFYfjng 1, 3-4;
cKyZr8MFNHBT7w9t 1-5; d4LFCkmQYhPRDx2y 1-6; ghWq2C84BJj9bQHd 2-4; hyCdAYx8GjgKfZHp 1-5; qm1Bvy29hGK6bMaQ 1-2;
rgKPxdF6G3NbMzmj 1-3; tCprA1RmvTcJkq3D 4; tdYLMmfgcAJCTWyr 8, 10-12; yJmtnzPNDVK42TQR 5-8.

## Silent Fang (item 13953, spell 1293654): 1.45, observed

466 procs on 11,915 rolling hits, 9 wearers, 9 runs, in the main hand: all hits 1.47 [1.30, 1.61]. The rate is all hits,
which the one-hand white rate and the specials rate agree with. Observed: its interval reaches further than 10% from the
read. One-hand white: 1.50 [1.41, 1.81], 154 procs on 3,839 lone units, 6 wearers. Dual-wield white: 0.92 [0.48, 1.19],
16 procs on 651 lone units, 3 wearers. Specials: 1.49 [1.23, 2.11], 64 procs on 1,609 lone units, 9 wearers. Two-roll
abilities, priced as one roll: 0.00, 0 procs on 1 lone unit. Direct: window [-11, 100] ms, median 0 ms, 98th percentile
1 ms, 188 isolated hits. Reports from 2026-11-16 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       7 |   397 | 9,673 | 1.54 |
| Hunter  |       2 |    69 | 2,242 | 1.15 |

Reports and fights: 2qMm9WvRpzYx3GaB 1-2; 4hHBvakRnT7p1GVK 6-7; DmjBrYd7M9VZwCHF 14-22; YZkH8CA3wQxNc1hB 7;
Ya9kKCWRwqjFpgz1 3-4; fMLcHzyZvt7QTCmg 1-9; nyAx3XNDVmYWGdrh 1-6; pNKJR39CwGnmXtbL 1; y6FZRHQnckD8YjAC 5.

## Spellsever Crossbow (item 274293, spell 1293714): 0.95, measured

4,593 procs on 119,039 rolling hits, 143 wearers, 199 runs, in the ranged slot: all hits 0.96 [0.94, 0.99]. No lone-unit
rate applies, since its travel reaches past the isolation: the rate is all hits. A separate count reads its lone Auto
Shots at 0.94 [0.90, 0.97], which agrees. Specials: none. Travel time: the proc lands a median 515 ms after its hit,
98th percentile 880 ms, which reaches the 1,000 ms isolation, so no window was measured (985 isolated procs). Reports
from 2026-11-09 to 2026-11-19.

| Class   | Wearers | Procs |    Hits |  PPM |
| ------- | ------: | ----: | ------: | ---: |
| Hunter  |     125 | 4,582 | 118,881 | 0.96 |
| Warrior |      24 |     7 |     119 | 1.47 |
| Rogue   |      27 |     4 |      39 | 2.56 |

Reports and fights: 1DVFRby49Pg2tpzG 1-19; 1KF2jNqTzc6VJ9fp 6-8; 1VCHKmNy8FgrLnAh 1-5; 1fHnGhFabmPTq6Dv 7-17;
1hpF2AyzjRdrL78N 1, 3-12; 1jgTLvJQBMP6kxrm 1-12; 1qpQ6FW8fxD9gC4G 1-7, 9-11, 13-27; 1w9jNgyZv7YKAacb 1-13;
1yMxz28Pm4NZbAJf 1-12; 214ndcFyjpPM8wah 3, 6-9, 11-15, 19-28, 30; 2WH4jLt3FADy8fnh 1-2, 5-6; 2aG3Ld8KVARCzrh1 1-7;
3LhpDbvyrtNFWdxH 1-5; 3PaV6RHdnzw7Jpk8 1; 3pCKd6jqgHyFc9kf 1-17; 3vn6yQxN8AWmpKaM 1; 4BhykDjm2p1ncLGK 10-11;
4KWV2rZDJMqNLRmT 1-3, 6-7; 4MJWTYdRLrtPjDN1 1; 4NMcqLwQ2rB9amCH 1-3; 6VJxcqnYa4wdBWQv 1-9; 6gvfVjrpZR8mcYxk 1;
6kN78G2xfqy3vwTP 1-6; 84VDXdcKZhHMPBAm 1-5; 8CzwkcFWrA91QaxG 1-12; 8R1c427qCbhPKgzx 1-3; 8WgnxAakbzDZdFBX 2-9, 11-17,
19-20, 22, 24-25, 27-28, 30-31; 8frHF64GzAhBxDyb 13-14, 16-22; 8rNkjZ2B6WdzFcKq 1-3, 5-7; 94fVcDCNBLztXh1k 1-12;
97FVRX1WLHqDmjAQ 5-12; 9DVWQNAfHXJBLy61 1-2; 9RrTvMpkamJPVw7g 4-15; 9g7NQBqDH6FGfP8L 1-4; 9zfKMnRxtJGZQXBL 1-4;
A8tQY9p3Dzn6TLRJ 1; BTRMkxmvfrwcZNaG 3; Bmk9chwYjd2F3zpZ 1, 3-6; CLXHZ3c2GyN6mYBD 1-4; CR4DHc2wFgjd7yrp 11-13;
CXvjmgaJ6YqKzD4k 1-2; CdD48cWzqZGQYmp6 1-4; D8JF7rbYQRcWxqn3 1-2, 4-8; DWJhjnmAM6GZdbKv 1-2; DtAhCGwd3cmQ4P2g 1;
DyYJRTv2XBfGcwC1 1; FDybHvQ3pPqBracN 1-4; FTXVdBrHJZD9aw7P 1-7, 9-23; FbkAhDQcP4KgpXGw 1-2; Fwcmz4r9aDnpRWAd 1-2;
GALhdpX84abJWPFm 1-4; GKZJBtWAvfPL81bc 1; GXmnxzAfNabcMB6Z 1; GZzaJH1QVjMKd8qA 23-34; GfwAMF6LZDm7TVWz 1, 3;
Gv6NDzAmFjC4P3qL 5-15; Gyw2T3LjptRW6vH9 1-6, 8-10; H9FXjaBWRpYVMq4r 1; HARvQ2hJWp8f3grD 4-9; HLrCnVfNMgYAJ1jd 1;
HN7bnjM6mrKWqDRT 1, 3-4; HnC7N2YVAdg3wtrk 1-4; Hp9ThMznR74PwYKb 5-8; J2XL6AyZqgpjx8QK 8-13; K6rZRbwYD1Wyjhpx 1-4;
K7T21rXDGxcWabtQ 1-3; KCYXdTfNgGHczbLw 1-6; KLNhgMTDAV6xY1Wj 1-3; Ky6k1d7mP2zVXWpt 1; KyYwHQjxAL8tzvdJ 1-7;
LKkr1pxAtT7CvcZN 2-17; M4CvXLZz3qm6pDRc 1-3, 5; M7BnDF4VPXcWRaZ1 1-5; MNZwjf63JzdrYgTC 1-3; MPkVah87LYAcnFdZ 1;
MgcQ6txHr9R1VjPv 1-5; MkrFhTq2HPnzmCVJ 3; MnX2ZVD8bxTFfjdG 1-7; MpcVyhTHnqCJvDFx 1-5; MtxT78RD4NkCJd1q 1-3, 5-7;
MyzjakDC9gBbpq6K 2, 4; Nc2grh6ByLvJCkRY 4-6; NzgdqKHnfLJWkFVM 1, 3-12; P8jXWFgJ7GVvwZ6m 1-4; PDT2yACNzd4LMZ98 1-14;
PFGvpDf9WZgcY6Rk 1-12; PLGKvft3bq8WgRYN 1-2, 4; PkCJZ74Lmc3HyW2d 1-8; PkWRr8jZp9ATL7N4 1-23, 25-26; Pn2LYHXpjZMDywgA 1,
3-9; Q1arfj4WPBNXqmLk 13; Q4GadwADtxbFBNW9 1-8; Q8GxhTmH9f6ajcNr 18-39, 41, 43-51; QJafvkCx4yhVXBmT 1-2;
QtVv7GfBDHLyn1bN 1; TFzR3fQkt2ap7K4Z 1-10; THqFxLkCP6ayd2tZ 8-9, 11; TQAKbWq7vVJHLgBR 6-18; TZnCMWqK9vDgHjBa 1;
Tc9DkwQFmvJr1NCt 1-5; V4HDKvxmYCkR1JnB 1-3; VZdxq6yNBrW48ztG 1-3; Vb3ZxfXtyDWr1ndP 1; VkJvpT6FhYwx4q9G 1, 3-11;
WYrCyJkdfMXw84Ph 1-21, 23-35, 37-39, 41-48; Wf26RdmcnLDTXwV3 1-2; XHDjdLb7R9P6NQm1 1-2; XYT1F4RgJLh9pyb7 1, 3-5;
XzxyRLg8QW9rqPGn 1-5; YH3c4zvFrpLtR1M7 5-9; YZyAfq6XN9k4a8md 12-15; Ya9kKCWRwqjFpgz1 3-4; Z4w7HKFPvcrTWd3M 5;
ZkVJAmC8vjaM6QTK 1-3; ZrD7NdhjWX4zG2PL 1-5; ZxWz2nQqXRcYNAbV 1-5, 7-10; a42r3p6ZxtBTRqfQ 1-6, 8-9; a6QpkvcR1YHLBqPh 1-6;
aVXmtyDj48AWwYQr 1-2; aj46hBQX1AfRFHtm 1-15; b12KD98fndF4hMq3 1-18; bCzt1Fq24mYcPXkB 1-2; bT3LyRBtax8FQnch 1;
bpV7JWX8B9xgrj3c 1-5; byYP67TH4KtZqm31 11-13, 15-17; c2J4C8kzHaQxnwXW 1; cNhpPv28MHBJCwWm 1-2; cY6xjKJN7QfdbHtZ 1;
d2qfGRQw8Pr3NaVA 1-8; dDKQgLW3MFPatynJ 1-25; dgnB4QvPHR8mhtWz 1-3; dn62GTvrBYFQbWfV 7-20; dnvYkhpKZQaqM2Hw 1, 3-5, 7;
ftTngaZr73hXKHLq 1-11; g7Bh9F2b1kz6Cnar 1-2; gk2jJvTHhW4zNDKB 10-11, 13-18; h6wkBb9Z7jFtTrnp 1; hCzY42vqXrc1mG7w 1-4;
hYHBQ9ZJX3vFAnkj 7-14; hrV3QgyzvfdZKJwF 1-30; j7V46tC2GkwNYTrz 1-2; jJWt7Dnm4qc9rzQd 1; jLRdP7ah9pzfgQ4r 169-178, 180,
183-184; k9dcpgjKfyXGwA17 1-4; kWBdFxYXLaN1vryA 1-6; mDbPCdX9AvwYJxGa 16-23; mDndY9rhZbAtWNPB 1-6; mYaWhF4n8zRLDZtw 1-7,
11-12, 16-17, 19-20, 22, 24; n1CBNaJq4WhAfw8Y 1-2; n7qPAd4rN1JyHfpR 1-3, 5-9; nhtYVjb7CAw9DvQR 1; pCjdMQw32FtPkYGN 1-5;
pKn8wBZcQ62PyJH7 12-24; qTHXf9y6ag1CMtdp 1-2; qg8YWNaCdDvV7XZM 1-8, 10; qprynNbXtRZV4DHP 1-7; qzv6W9fkDaP1MZYH 1-6;
rD8m4JW1bVyYPKGX 1-3, 6; rQ4dFakMnWXj89vz 1-6; rY4H13hP6MKZfXWR 20-23, 25, 27-31, 33-62, 64-66; rpLZ287nfKWyJGtb 4-6, 8;
rzvB21j7yD9hpaVF 1-20, 22-30; t681v9FWrLq4Tkxp 1-3, 5-7; tCprA1RmvTcJkq3D 1-4; tD7Rm3GBjdqPzhwC 1-4, 6-8, 10-14, 16-17;
tDK9h6nGz4cLTRq8 1-9, 11-16; tDhCZyVrn84xBaFR 1-7; tHz9M3arKp6PxA1W 1-14; tXFAxTLmg3bBNZrv 1-13; taFcCyAx9BD1KWXp 16;
tkQKL9bGT6JVR3na 14-16; tyFZPdH39kGq21An 3-25; v9Hf4bPtGq1LVykC 1-5, 7-8; vdNjgtCXJVM4PT9Y 1-3; vgrm7nPFN231GYRW 1,
3-15; vqCJK6rZNzHX2Ymd 1-9; w6A8bVm7vTatMCYr 1-12; wNqHMTvYmCQKXV3W 1-10; x3VKzQDvwLtCgbrj 1-6; x3p7mzNhFcarw1fk 1;
x7ybf4LjBD6GJqWn 1-9, 11-15; x89ayWwRt7DTvYCr 1-2; xRTpHyY7qC1GmghM 1-3; xYTCXa9myPJr6R8M 1-13; xnZRBaDMKtNHGQfj 1-4;
xnvgBCHkJf4Y19rV 1-4; xydMjQBYGHW91vTL 1, 3-6; y2dFcNR1rMmKTPAj 1-25; yKvCQdzpgqY1WwbA 1-7; ynL7Gb2MpkZT46j3 1-9;
ypQ9d13tMFJPNfvw 1, 3; ythJ7TZDXk6AjHMq 9-10; z4ybMrFWNGa1L2xJ 1-3; zD9hTZPfwC38bgq1 1-2; zgG84wjLNY1Pv9tp 1-2.

## Strike of the Hydra (item 6909, spell 1292657): 2.1, measured

4,070 procs on 35,041 rolling hits, 95 wearers, 110 runs, in the main hand: all hits 2.11 [2.04, 2.18]. No lone-unit
rate applies, since its travel reaches past the isolation: the rate is all hits. A separate count, leaving the Hydra
Strike events beside each proc out of the isolation, reads its lone white swings at 1.99 [1.88, 2.09], which agrees.
One-hand white: none. Specials: none. Travel time: the proc lands a median 133 ms after its hit, 98th percentile 283 ms,
which reaches the 300 ms isolation, so no window was measured (43 isolated procs). Reports from 2026-11-07 to
2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      56 | 3,736 | 31,979 | 2.12 |
| Hunter  |      40 |   334 |  3,062 | 1.98 |

Reports and fights: 12A94BJt3KCFZrY8 1-12, 14-21; 14caBjX2QdynLkq3 1-9; 1AxvQpfrXyTg3dCW 24-42; 1a3pDtCcPgBkL6wm 1-22;
1mRKL7Tb8qwjkgv3 1-2; 1yjwF8MxXqvtD4bT 1-3; 23ZyGvk7BXAVDT4F 1-14; 3BH2d16wtD9F78QY 11-18; 3tRNmGkZB2Jhnv6b 5, 7-9;
4nmJDcGNkZC2Phdw 1-6; 4y8nq6Xjg93zfrc7 1-9; 68fgHQ74dRxrwac3 1-3; 6DYqj7B2WHxPcabQ 128-134; 6X1L4nB7zCfqk8vV 1-2;
6jHPJBnLy1pF9Rkf 1; 6pjNwdr12qzBAKhW 1-2; 7GNK4Y9bkH1awZxL 1-10; 7yRD39nwhmMWYjbX 7-8; 8Zfz4JNLxhHcdDtX 2, 4-5;
8apqXKQg6mbC91L3 14-18, 20, 22-28, 30-36, 38, 40-46, 48-49, 51-71; AqxY6kh43jc2RPaT 16-17; AzMD4BWYKT6NGX2d 2-3;
B2k3Xcxz1hqQwJ6P 1-12; G4hzybWYmgrXvL83 1-2; GZzaJH1QVjMKd8qA 23-34; Gv6NDzAmFjC4P3qL 5-15; H3bRaDmzyjxTW2BJ 1;
HARvQ2hJWp8f3grD 1-9; HpftBG4KPCZrMxbF 1-3, 5-9, 11; HpyNQCMnDYFbJtP6 1-18; HtQr7kYaNq9MGgvZ 1-12; JQH2y6NWBXchGRbg 1-3;
KZYNzqfJFdW34DVr 1; KrtPndVLH4BjJcfD 1-4; L9k3ND6tw8VvcgjP 1-7, 9-18; LAm6bcJw47YBh2kF 1-2; MLhC3KjWrngq2vtZ 1-2;
MQwzXcfxb4v7hnFy 1-2; McYJrBbXAfKj82W4 2; MjfT3pQKRWcVG4gY 8-10; MvrY6aBgftZxj1QX 1-9; N178cBjPtChmbk4w 1-7;
NJLbaF8mGnk9yYAq 1; NKXBATdWYqmz78CV 1-2; NvnJVbtwK8z7THyx 1-2; PCfQFpjYdm3rD8qy 1-7; Pc3gyq8AFL9CD64m 2;
QCLPznqJMZ8vH3mf 1-12; QJ42CwNWLjFyvqfR 1-3; QhYmjJ4bHZXcgfk1 12-27; TCzRJ82rWcPZmp94 1, 3-7; TDqbRJWAN3x1FGhv 1, 3-9;
TFVvAaY6HLj2NqQp 1-3; TpxWY1MakPNb72H9 1-9; VHth6gR7xLbjQmnr 4-14; VQaYAf1Wxvm6ndHK 1-11; Vd1NZCgLMqxhb9kP 1-28, 30-39;
XF1faW9mcYtKrqMC 1-18; XVvTwBdtzAcKhPg8 1-14; XpQW1hmwHYDAJ8at 1-12; YX61J8CKqc42paNZ 1-4; Yt6CVhy34zjNHMcm 1, 3;
ZFLd4hkVbrJ68vnK 1-13; Zf21kBmtDM3AyRjF 1-13; ZkA3Vhz72n9Gy4dY 2-7; aRmjxFbtnGYAKCqk 1, 3-4; adXPnhT3kjHp9VcC 1-13, 15;
bPQCc3Nvq27HRWt6 1-3; bdYgfJk7Mh8GqZ4t 1-5; bnx7WwavBpKrQjNh 1-4; btfTZQLYD4hwvzqx 1-4, 6-9, 11; c7gHWKFrZMnm2yb4 1;
d6f294aV7vYnWmFj 1-5; dAV7JrjMkGfKNZtY 1-14; dQMKjpxJBCz1wrDW 1, 3-12; dZVp6gHA92mbDnPk 1-3; fnYThW3BpxKZv7jw 1-18;
gaDnjKvP6RkQTW2B 1-10; hG8tvQjLBJgf4WHY 1-2; hgd1qH4QYFyjz6AK 1, 3-6; jbhGzfrnJRWX978L 1-10; jrwhRBWdbZPK1Xka 1-3;
jtn8ap4XY6PzwDQL 1-11; jwDg8KGcd3QJ9kNZ 1-3; nB7qNVpyZvLPXCJY 1-20; nJtB3F7cQy6NmzGZ 1-12; nKdyat38zLA7fNJq 1-18;
njaHZVpvYFtrTLNh 1-10; nqjwyVJp17PDBYTX 1; pLFZK3MWV2TRw6t7 1-2; qfG6QWjd3wg1JVXv 1-18; rY4H13hP6MKZfXWR 20-23, 25,
27-31, 33-62, 64-66; t1rjBxRdP8DzVHhT 1-10; tVFZybKHRWp8xMcf 1, 3-10; txP9jAb7gGHhXckB 1-6; v4QYKcD7XymZq2JB 1-12;
vAGVD7aQZxNWX8h1 1-17; vnwtmMTpkR1aYP7G 1-9; vpdJ3cMhwTKNHX7B 1-2; vqBtZRrXwKQb6xHj 1, 3-7; vyzbt6kgRhTq4fdr 1-4, 6-9;
wYmr1kfWPpVz7jyZ 5-13; xnrg9m6jDMR1kaLN 25, 27-29; xwGgV9zJ3FNayZf1 1-2; yDMwzc49TJKmFfCr 1, 3; ycZLWXxQ4jk21w3q 2-28;
z4ybMrFWNGa1L2xJ 1-3; zWrR937TBH1FgvfP 1-2; ztD4pKHA6jGvfcVr 1-2.

## Sword of Decay (item 1727, spell 13528): 1, measured

404 procs on 8,932 rolling hits, 9 wearers, 9 runs, in the main hand: all hits 1.01 [0.92, 1.10]. The rate is all hits,
which the specials rate agrees with. One-hand white: 0.88 [0.84, 0.94], 42 procs on 1,059 lone units, 3 wearers.
Dual-wield white: 0.86 [0.50, 1.07], 66 procs on 1,714 lone units, 6 wearers. Specials: 1.10 [0.95, 1.34], 104 procs on
2,109 lone units, 9 wearers. Two-roll abilities, priced as one roll: 0.00, 0 procs on 8 lone units. Direct: window [-20,
100] ms, median 0 ms, 98th percentile 0 ms, 165 isolated hits. Reports from 2026-11-14 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       5 |   225 | 4,715 | 1.06 |
| Rogue   |       4 |   179 | 4,216 | 0.94 |

Reports and fights: 2mr4qBpFYGxVMAnj 1-2; 9nCtfP6kvHrzV4FL 2; LDg2vXctwnYM9jBq 1-6; MXjF1GhRmLfabNZV 1-18;
TaNXwvhfBLxdyQc9 1-2; ZHK32LtTCmqR4Pxf 1, 3; ZqghBpkcy7xtPfmH 1-9; nZ3NhbvpVXg649Bm 4-10; qg8YWNaCdDvV7XZM 1-8, 10.

## Tainted Pierce (item 8225, spell 13530): 2, measured

872 procs on 13,855 rolling hits, 17 wearers, 19 runs, in the main hand or the off hand: all hits 1.99 [1.81, 2.11]. The
rate is all hits, which the specials rate agrees with. One-hand white: 1.26, 8 procs on 200 lone units, 1 wearer.
Dual-wield white: 1.80 [1.46, 2.01], 313 procs on 5,489 lone units, 15 wearers. Specials: 1.88 [1.44, 2.28], 69 procs on
1,159 lone units, 8 wearers. Two-roll abilities, priced as one roll: 3.54, 22 procs on 196 lone units. Direct: window
[-10, 100] ms, median 0 ms, 98th percentile 1 ms, 277 isolated hits. Reports from 2026-11-14 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       9 |   430 | 6,503 | 2.09 |
| Warrior |       6 |   356 | 5,872 | 1.91 |
| Hunter  |       2 |    86 | 1,480 | 1.84 |

Reports and fights: 1zLtZcp3DRP97Hdx 1-6; 4na1LcCpP8QHTNVD 1, 3-8; 8WfLRh1qTpFxJgra 1-2; 8rNkjZ2B6WdzFcKq 1-3, 5-7;
J97GvmrRXCFZfQ8a 1-2; LGj3Kfvkz7cgZ1ya 1-5; NXgv93kr2G7zyTxm 1, 3-9; R3CDcTfm6bdNJrHx 1, 3-7; RcTKFQbGnmtJHWpa 1-2;
THqFxLkCP6ayd2tZ 8-9, 11; TnbzrVaxtHcpMKLY 1-3; VyZXRT7nrPJKBAt8 5-13; akPJ4AcBZQnb36f2 1-5; c9tKjzRyaYXQfnk8 1-3;
gn4HAC1RqMfLt8wN 16-21; pFNnChR7GzAX6aH3 1; qcrKJGPvkfhAtdjQ 1-4; xqy2GfPj67XCwrT3 1-4; zTMW1B2rRJxbm3w6 4.

## Taran Icebreaker (item 2915, spell 16415): 0.95, observed

694 procs on 11,509 rolling hits, 10 wearers, 11 runs, in the main hand: all hits 0.95 [0.90, 1.01]. The rate is all
hits, which the one-hand white rate agrees with. Observed: every wearer is a shaman. One-hand white: 1.10 [0.95, 1.27],
219 procs on 3,150 lone units, 10 wearers. Specials: none. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 2
ms, 188 isolated hits. Reports from 2026-11-13 to 2026-11-19.

| Class  | Wearers | Procs |   Hits |  PPM |
| ------ | ------: | ----: | -----: | ---: |
| Shaman |      10 |   694 | 11,509 | 0.95 |

Reports and fights: 1MA3qWLGa8v2hdXP 1-48; Bz8CAjZnNyXhgTtH 1-14; FRmWJKpCD8t7QGb2 1-3; PJHx4V21KhfCaryz 1-2;
Tx7MPH9f8Ga6hdrF 1-20; VhN7grtcvRYJxjBH 5; aAcjGPbyZXNTzR1K 1; aQAhMLzBbWJcdq6G 1-6; cLxk98bVfwpXWaqN 1-21, 23;
mg3MBw6rcNtZpjbA 1-14; nhtYVjb7CAw9DvQR 1.

## Thornspike (item 6681, spell 1293183): 3.15, observed

17,132 procs on 191,806 rolling hits, 242 wearers, 279 runs, in the main hand or the off hand: all hits 3.15 [3.08,
3.23]. No lone-unit rate applies, since its travel reaches past the isolation: the rate is all hits. Observed: no
lone-unit rate checks it. One-hand white: none. Specials: none. Travel time: the proc lands a median 525 ms after its
hit, 98th percentile 948 ms, which reaches the 1,000 ms isolation, so no window was measured (342 isolated procs). Left
out: 2 of 282 runs in Cat or Bear Form. Reports from 2026-11-07 to 2026-11-19.

| Class   | Wearers |  Procs |    Hits |  PPM |
| ------- | ------: | -----: | ------: | ---: |
| Rogue   |     169 | 12,619 | 140,616 | 3.17 |
| Warrior |      41 |  3,746 |  42,171 | 3.14 |
| Shaman  |       4 |    380 |   4,586 | 2.92 |
| Hunter  |      28 |    387 |   4,428 | 3.08 |
| Warlock |       1 |      0 |       6 | 0.00 |

Reports and fights: 12RB4fbmrp3YZjPF 1-5; 192ZN3dKFymLACY8 1; 1VKQY2Dafc7PJzpR 1-2; 1Zv6bBfgDP3RVFYa 1-2, 4-7;
1kWMb9cQ67dJhrZz 1-3; 1tCvNKHFM3c2WkmA 1-12; 1w8AbW4RBNLX6Kfy 1; 21PKRb9nk3HdYFm8 1-11; 2K4Z1jLyGWzdnD3f 1-8, 10, 12;
2KkaByn6H9mzb4Yq 1-4; 2t6mDPLkfWBxcr4q 1-18; 39BFWKC2J8RTzyMx 1-3; 3DdtwKbCf7g1XVyT 1-2; 3PXfhvmTp6NQdwRb 1, 3-12;
3ZFVrmzh4fanLM1B 1-8; 3bqWjVMPpDfdCgYz 1, 3; 3f4WnNCQJX9azxmV 1-5; 3hNwr9K71bz4xjCc 1; 3kqWrL1dwJfxMcX2 1-5;
3npdHy9Ag7rMTC6f 1-5; 4KCzQ8mFr9A1bkxG 1-8; 4L7DBHTV2jA6gkQW 3-19; 4WYavRTZm7QHxLgA 1-20; 4ZNwHaLtygF1k3xC 1-3;
4zYkPVvpb9KXLxA7 3-5, 7; 6CPXjKh3z1TDWtqM 1-10; 6tTZVHLYkbzrCXqD 1-8; 72gPMatGLpwd93jW 13; 76ZV8FQ1RHDhbq3z 1-4, 6-9;
7M1KwnDZkXvyRFB3 1-5; 7dZrzYqjKHh28mvW 1-2; 7dZt6b9Kh4F2CNAn 1-12; 7f6DBRAPMLGYzyQm 1-2; 7g9DpfNFBA3PatdZ 1-7;
7gPtBMrHwGyjWKRc 1; 7xGQWc4AbvD8CNBw 1-7; 81CznKf7taTFArLw 1-8, 10-16; 8MjcygF3zWK4d9NB 4362 fights from 1 to 4766;
8Z1rYqmCWKbLjvwz 1-2; 8ZTCpbxtFA1zfRNX 1-3; 8cyBqw6Cr3WjV9Ft 1-6; 8raZtT1Qq7WkYv2P 2-5; 9B8Wf16dHKjgaMDY 1;
9Nayz1f8BJnLcgQm 1; 9PxCyQfrdhXFmNGV 1-2; 9g7NQBqDH6FGfP8L 1-4; 9v6VdC7ymQfb4cWh 1; 9zfKMnRxtJGZQXBL 1-4;
B1xchF3f6L2Mkga4 198-205, 207-219, 221-238, 240-257, 259-269, 271-282, 284-287, 289-292, 294-306; B6jfw4mcnPxWCKYD 2-18;
BZxkqfarM97yHgLp 1-6; BdLc4AgFbrz8Tq2k 1-2; BxpTfqGZb34aJLc6 12-14; CPj7JpNFLDztgQRa 10-19; CbG9THJkRBjDmPX4 1, 3-6;
CjLKJRNhnb7TBzAZ 17-20; Ct8XcxY26Mf34NjR 1-9; CwYNtp2VQrJdbfx9 5-8; D8qzgc9h1YLf6krp 5-6; DHyYh18vLXAFTJqt 1-5;
DRv3TVCwrAPJ9zK8 1-2; FBQzxArfjdCZp6Lg 2-11, 13-18; FTrHGQmvDM2jzRn4 7-10; Fj9WzagMcZxw4mXY 1; FvdWRrT67kAPY43x 1-11;
FzCJvnHfGyMLWVwb 1-11; G6wYN314W9FfVXpr 1-2; GAgDtJmKHdBhnpwV 1-5; GCyghdxmrfYL8zcX 1; GJxyK8rTBamY7CvH 1-11;
GKmRMd7XzWtavJg8 1, 3-9; GNRbXQDvCkYpAta4 1-8, 10-12, 14-22; GRZrtByWvML6mzXA 1-11; GRdTC8kXNwVxpjKQ 1-2;
GV4wtvzD8RcfTqNC 2, 4, 6-15; GVzLqkhK4PMXjYmn 38-50, 52-70; GX7CDf4dvtLmR1Wh 1-3; GXmnxzAfNabcMB6Z 1; GYPbQpyCvVajRH4L
1-16; GYQB4c79w6TF1nWZ 1-6, 8, 10-11; Gb2W8Tga4VqvZy9c 5-11; GpdNkxVgzrtF49Xb 1-3; H7yaPANhDBKQ8Lmq 1-13, 15-17;
HTpFLJAq8CWmPvzr 1-2, 4; HWcyVPwhnNB9brYz 1-5; Hdnh1GJgD4TbQkXL 1-14; HqW4rjd6XmYgDM9F 1-2; J3tdCaXR8hAP9m2M 1-4, 6-9,
11; J9jRmkxvtXcAKfM3 1-4; JB7fzC2ZnP3Kwjxd 1, 3; JD1ypjL3xcCKfTGP 1-2, 4-6; JNZVwGWTBMYLz9an 1-13; JPB6NXkAGCQzmh38 1-5;
JZx3p6Nqdj9tCrMc 1-3, 5; JvdQP6VyY19kZDXm 1-9; K3jWRx2qQ1aXy9PA 1, 3, 6-13; KG4FcnrYWVHCfzvt 1-9; KNqTmpW6HZCnc378 1-4;
KmahZwCJcYFLj1Rq 1-7, 9-16; KmjAV91JBhgQLrfP 3-4; KvF7tTamGpn9kW34 15-24; LN78yvQwgJpcAYCG 4-6; LW4qmQykhj3p9atv 1-2;
LWx7bgKmpQ6ktYRr 5-12, 14; LaMmthYQK4HRWwXC 1-5; LaQztr3mDTwHyknB 5-8; LmDb8njHfXvAgC79 1, 3-42, 44-51, 53-63;
LxdYWR47ZFbTNBhM 1-2; M4CvXLZz3qm6pDRc 1-3, 5; MCdZTywRz3HYnvWg 1-10; MKk2adJtjyANf8LW 1-13; MW1Kw23ncjRDrxpA 1-14;
McNx4Dwk89vYJpdH 1-5; MdaKf7zpNgWHF1Jj 1-4; Mf1AyXnC6W3dpNJR 1-12, 14-16; MyZ8p6RaAFBWcJ1f 1-7; NLzDkMAj8n2TmbxV 1-6;
NWB8HZd1JvxzGjDr 1-3, 5; NZcYhqPxkrJMKz6a 1; NfLb7C1qDZ8cHxFp 19-29; NwKVWqHfgk1jLvAG 1-8, 10-13; NyZgJMVBXHn8vtWk 1-8;
PFGvpDf9WZgcY6Rk 1-12; PHqvNCxT1JVtpWFj 1-4; PN64hvHQLZRCptGg 1-5; PNnXcA9hFBx7Wtdj 7-11; PyH9Adag8VwRjKJh 1-16;
Q1arfj4WPBNXqmLk 11, 13; Q3kAgpdcGJ27VPvT 1-3, 5; Q4GadwADtxbFBNW9 4-8; QAc4wgTL7PfDphk6 1-5; QMkfqDhm3L69A7VY 1-11,
13-14, 16; QRbmZWh2yDGakt9Y 1-4, 7; QxJKAphaZt61kzvg 1; QyYtrwM6zKPXHDj3 11-15; RDPLkYamNgHKXFh2 3-7; RbLqYxNQj1MaWFdP
23-29; RdmgVNQ4XJMfnC1F 4, 6-10; RfxgwvQjA1pVy98k 1-6; T2MrDwA9hdgPXxRv 1-4; T4pxtrKv178PGNkA 9, 11-14, 16, 18-26,
29-41, 43-57, 59-120, 122-142, 144-164, 166-208; TgVP6hxZKqY7MGQ4 1-5; Twr9aYQFMWjnLPqR 1; VMP6Dv2pyTnYhaKJ 1-3;
VRM3KtXTh2W7kzJx 1-4; VYb6f27NazpmCch3 1-2; Vh3bzdRMP6ymNKq9 1-2; VhHLyw9CP76Wv3fX 1; VvmGw1CT7bZMNXp6 1-12, 14-19;
WAKwhqRtmPGT49ZQ 21-35; WFJf43apKLXHVzBh 1-4, 6-7; WqBa2AwcGHfgj8Lt 7; Wza8bdXVRhcgqGnZ 1-5; X16Yp9FzW87cNADq 1-5;
XYTRKMVFWdypLxtB 1-14; XhZAV37vMFbYWRt6 1, 3; XwqT7ZK48Df69tkh 14-22; YJzQRVL4r9DqFcpf 1-16; YcgyKZVCJvbW7Hj8 1;
ZCRLWJqmb7v3gjaN 1, 3-7; ZKLvG3Bmy9CAgjT1 1-7; ZKRmDGdV8FM1aYwt 1-13; ZPK6pnD7HzmyhvjT 1; ZPxzwLAmBpRMTNbC 1-3;
ZPyMjD1FkVGJH8Lb 1-4; avQywnDKt7RxcpfF 1-8; b12KD98fndF4hMq3 10-18; b4fNagK2jvXQJAGL 1-10; bAZJWP3hgxRav6Fm 1-7;
bdPKyBjRNDX9CHaQ 15; bpV7JWX8B9xgrj3c 1-5; bxywkpHBTzg2mtCQ 1-20; byYP67TH4KtZqm31 1-13, 15-17; cLkwmZbzFDKdPhYn 21-22;
cPHdgRMCy7ZGDqFf 1-8; cT8wJVq316kAWpn9 1-3; cVPtG1agnWhfX2xC 1-7; cXCfvnWtNrR63PVx 1-2; ctAnYvNLQK6F38PD 1-2;
d8KLZbmzWhwr9at7 1-4; dXR2BPqTYQ4aMxZj 1-5, 7; dtHTkZgMNvR3YBK4 1; f9vADt2CZ3M1ygwq 1-11; fRPcJYzWKXBCv32N 1-2;
fjpy1TBL4b3kgJhr 1-8, 10-15; fjqQNyMbD1Y7R92h 4-21; fkD9m3dBKcGAyqRY 1-6; fnMH4F2pkGC8T6Kz 1-8; fr2WyGqPVYNTXd91 1-4;
g2qW8NnDvCQ3TfVG 1-5; g8b1MkJnKtyfzvPa 1-5; gGn1vQ28YxpFrmaP 1-3; gT2mW39CV7DMwhzx 1-21; gbFJzMXnTKkxNQY7 1;
gdMNAjD2Kyrtqh97 1-8; ghbnmqx4Wz8Rj3Ck 1, 3; ghcTDmb871vtxyWM 1-5, 7-11; gn7Wva1FKAm8HhGZ 21-23; gwkN81FaqdAPQTfW 1-4,
6-8, 10-11; hXT9DqC7vpQW6ZLb 1-9; hf1pF42aPHMv9kLX 1-9, 11-20, 22, 24-36; hmtTqH16pPGzZJax 1, 3-17; hvgw9t1par7YCD3b
1-17; j7J4bKFHq6M2Wfvy 3-6; jD7whNgBrM4GxVvF 1-8; jGrv6B1hNMDJPc3W 5-10; jHXnNA128R4FmwYB 1-5; jQzWyR89avmgpd7H 1, 3-7;
jYAhJZ3BzPMHWx1q 1-5; jnxaGm43ctgQMFhf 1-8; jtdGCBNJq38zxcQ9 1, 3-9; jyxcbwYXG3pL9NnP 1-2; kHfNcJTYj3RtMFGL 1-3;
kJHLMmf3rGaj28d1 1-2; kX94Vmn8tpJzGTD2 1-11; kqXb3Tja4gZfxP7V 4-6; mDbPCdX9AvwYJxGa 7-30; mMFCR2atzKHWDcdQ 1-6;
mPMpkAgjbtR7HBXD 1-19; mc821XQyrpDjGJh9 1-3; n3hjGAFXxTaBfmRz 1-3; ndHRLVB9Kaz7xpqT 1-2; nmrNLB3j4pAPg8fQ 1-2, 4-15;
pKn8wBZcQ62PyJH7 8-24; prbPqjNgMz6FaBX7 1-7; pz7YaGwxXZgcbQPm 7-9; q9FHYznbZLDKtQTP 1-2; qLk1fVdN3rcaDxm6 1;
qMD3CVapwb6nAgx2 35-40, 42-43; qQhkt3L9yKfTFp2D 1-6, 8-13; qYRLwvVKDkTZGP86 20-43; qg8YWNaCdDvV7XZM 1-8, 10;
qyWNYFVZjLX1mtJA 1-19; r67HcKRnkYdzTQZP 1-19; rbg3wR9B7dKf8JmW 1-3; rpLZ287nfKWyJGtb 1-6, 8; t7432ga9HXvFBKzy 1-6;
tD7Rm3GBjdqPzhwC 1-4, 6-8, 10-14, 16-17; tLXnZDMCvJ34mzbF 1; tc1qGLjpgVrJ8WdQ 4-5, 7-15; th1nf93NzQjmVdrD 1-15;
tjN8BgxX7k4GyAfd 4; tyAMjpZb9Yq2wkV7 1-4, 6-10, 12-19; vCZk3N8n2dbTyMFV 1-34; vDg3LH1Zyfz49kKr 1-12; vFqn26gPR4NJBfjw
1-4; vKznWFRC7cBHTGwx 3-6, 8, 10-11; vpBwx4KbaWzRZ8V7 1-3, 5-10; vtRzBNFnW6KH3pjT 7-8; vzZPBtWcfm9HY8Ty 4-10, 12-18;
w3L1hH2ya4gCTKAq 1-13; wCGWfTxNgFrB8hDy 1-15; wHtnq3pgjxYVy9FZ 1-2; wRY9NtQp7bPZ1VDz 1-6, 8-12; waYzfQgyrcKV7H9J 1;
wfgxbHpNB2vZJYLW 1-6; wvbng2zVTYRrc9Zd 1-2; x7ybf4LjBD6GJqWn 1-9, 11-15; xDwACkbMZRyNcd1z 1-5; xJbw6RTvtHPncNzp 1-5;
xRNZwp7fWvTY29qV 1-15; xbCaBqNdAn3ZWKDw 1-3, 5; xmYzGavFHWpfbLkt 1-3; xnrg9m6jDMR1kaLN 21-23, 25, 27-29;
xrtBTqdwGDfzK72Q 1-2; xtFwVRc3gmdjLAyC 2; xw86k9LyMNF3YHjr 7-14, 16-21, 23-31; y6kYARn4dpCKavZ8 1-3, 5-6;
ycjXJnNkC8pGWH1Z 2-7; ypQ9d13tMFJPNfvw 1, 3; zDhrvnfPYdt9FXMj 1-4; zjxyZAgPDBMGndRK 1-16, 18-20, 22-24, 26;
zwPKLTQ4W6YkyRqd 3-8.

## Toofbreaker (item 286430, spell 1320269): 2, measured

4,905 procs on 77,305 rolling hits, 49 wearers, 53 runs, in the main hand or the off hand: all hits 2.00 [1.95, 2.07].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.01 [1.90, 2.12],
1,411 procs on 22,187 lone units, 33 wearers. Dual-wield white: 1.99 [1.90, 2.09], 421 procs on 6,684 lone units, 16
wearers. Specials: 2.02 [1.85, 2.19], 480 procs on 7,519 lone units, 34 wearers. Two-roll abilities, priced as one roll:
0.00, 0 procs on 1 lone unit. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 1,631 isolated hits. Left
out: 1 of 54 runs in Cat or Bear Form. Reports from 2026-11-13 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      39 | 4,411 | 69,942 | 1.99 |
| Rogue   |      10 |   494 |  7,364 | 2.12 |

Reports and fights: 1BNXynxVbH4kGt8R 1-2; 1L9BQW2AcqdP8bfG 1-2; 1cm2k9qHaRdBMpFx 21-36; 2PZCz9c1HgtrX7Th 3-5;
2taZz4vTdfXYAk3b 1-2; 2xY1D3mCwagdy7Wc 1-2, 4-6; 2yqzTJfLN8cvwaDK 1-2, 4-6; 3JbjLGXVzctAmfaD 1-4; 3cWzXTLmvDh2bC7H 1-2;
4YVTAqgG3dtmD8C9 1-23; 4vZxJdFNqTXMRkD8 1-3; 4wWtcgbPhNdMJ2LQ 1-33, 35-37; 6aN7CQwbfTW182vp 1-5, 7-13; 72JjqrWXMdtwKx68
1-3; 84nxvZk9BmHf7Ldr 1-8; 9vWqJjhLGCpdQZKc 1-11; ALpDvxc3GtT7Bnkg 1-2; Cp3ND2J1xWhbFzP7 1-26; DavdRxwG1TZ6rzM7 1-4;
FLK4Vqw3cp6XkRAT 1-15; HkNB7A6tanYcGLWj 1-7; JX4afWrdcw8RkHDt 1-3; KGbxV6BjrZ3QkdDL 2-3; KJGzHwk2Xyfdc9RL 1-10;
LH9kgYvBqyVKadh3 1, 3; LPQd7pfbjDkYZyRz 1-2, 4-5; LcvAbQBGHk6MwgWR 1-3; Mc16FCfkmzRQZ4TH 1-2; P6qAmjBhrNa8b9MR 1-12;
PXbhqkFHpt47RmDy 1-5; PqNHCAjBt2RgrJwQ 1-41; R1jPGYZyTtBkKr8J 5-12; RyrkQYbMVXBFv23J 1-2; VdzPcGgaCnKHY4jF 1, 3-13;
YP3LzZWFTMDGXdyR 1-4; ZqBnHh1CtfDcpwK7 1-6; a7F96kX1mQcbKDvw 1-2, 4-12; ahyJfHkjWGV7P28M 1-26; bA8g6NZK4h9MPHRJ 1-10;
cRtHw9DWZMrKnyvQ 1-4; gTPNZcJ9mG4dRC2X 1-4; hRzbpMrDY3cgQX7n 7; jVhAWnaHJgf3FPmy 1-10, 12-13; qTHXf9y6ag1CMtdp 1-2;
qr7Vt9aW8TpFXJAD 1-2; rTYJ6yWfNcbKXztR 1; tLGanMDfw82FJymz 1-16; tac7hXVRgkbZNMDy 3; vtyhfFz9wPGJdHqM 1-3;
w6A8bVm7vTatMCYr 1-12; zTbHr3BPvKxRCMZg 9; zgG84wjLNY1Pv9tp 1-2; ztpB7m1kbLQ3KdqR 11-14.

## Wraith Scythe (item 11920, spell 16414): 3.45, measured

- Count: §Count of the 2026-11-06 to 2026-11-21 reports

1,136 procs on 8,989 rolling hits, 8 wearers, 8 runs, in the main hand: all hits 3.45 [3.17, 3.73]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 3.09 [2.75, 3.46], 218 procs on
1,922 lone units, 5 wearers. Dual-wield white: 2.54 [1.93, 2.80], 40 procs on 430 lone units, 3 wearers. Specials: 3.81
[2.88, 4.25], 141 procs on 1,008 lone units, 7 wearers. Two-roll abilities, priced as one roll: 4.70, 5 procs on 29 lone
units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 286 isolated hits. Dropping the swapped
stretches (19 active minutes, 7% of hits, in 1 run) moved the read from 3.22 to 3.45. By report date: before 2026-11-18
3.57 (443 procs); from 2026-11-18 to push 114313 3.37 (693 procs). Reports from 2026-11-14 to 2026-11-20.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       6 |   849 | 6,899 | 3.36 |
| Rogue   |       1 |   247 | 1,760 | 3.83 |
| Hunter  |       1 |    40 |   330 | 3.30 |

Reports and fights: B6JLrnWKF9htyjv2 1; DZjKqdLXkBztYJTC 1; p1VPMyG8jgR6Ymc7 1-4; PtKXrwfj8zBHm2Gy 7-8; PtmYHjkDrqhdR9MA
1-6, 9-10; XPgHryZdABkRDhwV 3; YnVFyQq3w28mr41R 1; ZQtv6AqYmdJFWwhy 7-8.

## Wyvern Tailspike (item 5752, spell 16400): 2.05, observed

445 procs on 7,293 rolling hits, 14 wearers, 16 runs, in the main hand or the off hand: all hits 2.03 [1.83, 2.27].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: its interval reaches further than 10% from the read. One-hand white: 2.25 [0.00, 2.26], 26 procs on 385 lone
units, 2 wearers. Dual-wield white: 1.88 [1.60, 2.22], 190 procs on 3,376 lone units, 11 wearers. Specials: 2.36 [1.11,
3.13], 26 procs on 367 lone units, 7 wearers. Two-roll abilities, priced as one roll: 0.00, 0 procs on 1 lone unit.
Direct: window [-10, 100] ms, median 0 ms, 98th percentile 0 ms, 174 isolated hits. Reports from 2026-11-10 to
2026-11-17.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Rogue  |       9 |   386 | 6,288 | 2.05 |
| Shaman |       2 |    47 |   718 | 2.18 |
| Hunter |       4 |    12 |   287 | 1.39 |

Reports and fights: 2RdqgkHFNVMnKCh8 1-2; 4NWagKbPRwT81yh7 3-8; 8129XG6zfMVJrvmg 1-3; Amh4kVqaFGB3xPZ9 1-3;
JD2ag1wYV8qCxhtQ 1; JPymD8BKctgqMFLH 1; MLhC3KjWrngq2vtZ 1-2; QZyPVgTvLJnamk84 1-3; Wh7fgRNMmLP4YBnX 1; b6QHjXT2GF7A1YPr
1-8; cT71Vr9tPDKkxWv3 1-2; dGhawD4rxCAH7bR2 1-4, 6-10, 12, 14-21, 23-30, 32-37, 39, 41-51; fyb2xc14zNRdvDMY 1-2;
hxZPKTfQYzdVjvN3 6-9, 11; rqhGLfWyaAtZvjTC 3, 6; vnDM8K4RYx6kW2GH 1-2.

## Count of the 2026-11-06 to 2026-11-21 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every wearer stream on disk for its weapon; 8 reports in all
- Game build: by report date, 1.60.1.70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC and 71037 from
  2026-11-20 18:23 UTC; no report starts after hotfix push 114313 (2026-11-20 23:47 UTC). The sim's database is
  1.60.1.71037, push 114326; the hotfix each player ran is not recorded
- Date: 2026-11-14 to 2026-11-20
- Pull request: #1156

The count is §Method on the hits EV-0155 finds rolling a weapon proc, run again over every wearer stream on disk for
these weapons, with more wearers fetched on 2026-11-21. The same rolling hits, classes and forms are left out: paladins
from every rate, druids in Cat or Bear Form, and the two-roll abilities with the procs around them. The intervals,
lone-unit checks, swap drop and credit windows are this file's.

A rate is measured where a lone-unit rate of 50 or more procs (lone one-hand white swings or specials) overlaps the
all-hits interval, and 5 or more wearers of more than one class give an all-hits interval within 10% of the rate either
side. The rate is the all-hits rate rounded to the nearest 0.05. A lone-unit interval from under 5 wearers is a Poisson
interval. On Wraith Scythe every class past the first is one wearer.

Each section's read by report date is the all-hits rate over the reports that start before 2026-11-18, from then to push
114313, and after it.

Scripts and inputs, in `forever-notices-1121`: `results.json`, the counts of `forever-sweep-auto/analyze.py` (version
12, as EV-0157's, with the paladin check that sets no rate) run with `--items items.txt --census census_used.json`;
`report.py`, the call per weapon (`rep_final.json`); and `dates.py`, the report dates (`dates_final.json`).
