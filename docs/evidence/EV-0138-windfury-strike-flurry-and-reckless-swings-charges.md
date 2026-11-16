# EV-0138: Windfury Totem's strike, Flurry and Reckless Swings charges after push 114115

- Claim: a Heroic Strike or Cleave spends one warrior Flurry charge and one Reckless Swings charge per cast, landed or
  not and on one target or two; Whirlwind, Bloodthirst and Hamstring spend no Reckless Swings charge; Windfury Totem's
  strike spends no Flurry charge, starts or refreshes Flurry on a crit, spends Reckless Swings and Gore Drinker
  charges, rolls Unbridled Wrath at the white swing's rate, and grants the rage of a main-hand white hit.
- Sources: combat log, a level 60 dual-wield Fury warrior (Flurry, Unbridled Wrath, Reckless Swings, Gore Drinker) and a
  level 60 Protection warrior with none of those talents, an Upper Blackrock Spire group with Windfury Totem rank 3
  (strike 10610), 2026-11-16 21:30 to 21:56 UTC
- Game build: 1.60.1.70938, hotfixes up to push 114115, relogged
- Date: 2026-11-16
- Pull request: #1069

## Method

Unbridled Wrath's energize, Reckless Swings' stack removal and energize, and Gore Drinker's stack removal log 16 to 25
ms after the hit that causes them; each is credited to the warrior's latest own hit within the 60 ms before it. A
Flurry charge a Heroic Strike or Cleave spends logs in the cast's ms, before its damage lines; one a white swing spends
logs 0 to 47 ms after it (median 17). A cast counts as clean when no white swing, strike or Flurry refresh falls in
its window. Rage is read from the warrior's own resource fields on the lines just before and after a strike, with no
other swing, ability, energize or damage taken between them. All counts are the Fury warrior's, except the strike rage,
which counts both.

## Flurry

Clean casts with Flurry up, by the charges each spent:

| Cast                     | Casts | Charges spent |
| ------------------------ | ----: | ------------- |
| Cleave, 2 targets landed |    22 | 1 each        |
| Cleave, 1 of 2 landed    |     3 | 1 each        |
| Cleave, 1 target         |     2 | 1 each        |
| Heroic Strike, landed    |     8 | 1 each        |
| Heroic Strike, dodged    |     1 | 1             |

One more Cleave seems to spend 2; a white swing 26 ms before it spent the second.

Of 74 strikes while Flurry was up, 68 have no charge removal after them and 4 have one a white swing, Heroic Strike or
Cleave explains. The other 2, and one more like them, are dodged or parried strikes that share their ms with a Flurry
refresh and removal pair. That pair occurs 50 times in the log, each 16 to 60 ms after a crit (Whirlwind 24, white 16,
Cleave 4, Bloodthirst 2, other 4), and a white swing or Whirlwind sits 16 to 17 ms before each of the three strikes.

Every strike crit applied or refreshed Flurry, 22 of 22: with Flurry down 8 applied it (7 clean), with Flurry up 14
refreshed it (5 clean, 9 with another crit nearby). The 79 strikes that did not crit gave none.

Not settled: the same refresh and removal pair follows Whirlwind crits (24 cases), so Whirlwind may spend a Flurry
charge on a crit.

A Heroic Strike or Cleave that crits spends its charge in the cast's ms, before its damage lines, and the crit then
sets Flurry to 3 charges, so the cast ends with 3. Clean casts, with no other own hit within 70 ms either side:

| Cast                | Flurry before | Flurry lines after the cast | Casts |
| ------------------- | ------------: | --------------------------- | ----: |
| Cleave, crit        |             0 | applied                     |     2 |
| Cleave, crit        |             1 | removed, applied            |     1 |
| Cleave, crit        |             2 | removed to 1, set to 3      |     5 |
| Cleave, crit        |             3 | removed to 2, set to 3      |     2 |
| Heroic Strike, crit |             2 | removed to 1, set to 3      |     1 |

## Reckless Swings

The log holds 141 Reckless Swings energizes, each with a stack removal (143 removals).

| Cast while it was up | Casts | Stacks spent per cast   |
| -------------------- | ----: | ----------------------- |
| Cleave               |    19 | 1, on 1 or 2 targets    |
| Heroic Strike        |     3 | 1                       |
| Whirlwind            |    12 | 0, on 1, 3 or 4 targets |
| Bloodthirst          |    11 | 0                       |
| Hamstring            |     6 | 0                       |

Landed white swings spent a stack 106 of 106 times while it was up, missed ones 0 of 22. Strikes spent 14, credited to
13 strikes, so one is credited to the wrong hit. Whirlwind's 78 hit lines while it was up spent none.

## Unbridled Wrath and Gore Drinker

| Landed hit    | Unbridled Wrath | Gore Drinker stack spent, while up |
| ------------- | --------------- | ---------------------------------- |
| White swing   | 276/469 = 58.8% | 150/193 = 78%                      |
| Strike        | 61/103 = 59.2%  | 9/17 = 53%                         |
| Heroic Strike | 14/20 = 70.0%   | 6/10 = 60%                         |
| Cleave        | 111/165 = 67.3% | 38/58 = 66%                        |
| Whirlwind     | 2/425 = 0.5%    | 26/161 = 16%                       |

## Strike rage

Fifteen strikes have a clean rage reading, and all fifteen gained rage, each the amount the same warrior's main-hand
white hits give:

| Warrior    | Strike          | Strike crit      |
| ---------- | --------------- | ---------------- |
| Protection | 9.3 to 9.4 (10) | none clean       |
| Fury       | 9.6 to 9.7 (3)  | 19.3 to 19.4 (2) |

The logs EV-0136 counted, up to 2026-11-16, show the strike granting no rage.
