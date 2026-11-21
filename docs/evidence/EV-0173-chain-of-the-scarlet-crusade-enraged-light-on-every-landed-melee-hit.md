# EV-0173: Chain of the Scarlet Crusade's Enraged Light on every landed melee hit

- Claim: while the 4-piece bonus's Enraging Light (1293673) is up, every landed melee swing or melee ability of the wearer
  that does not kill its target casts Enraged Light (1293675) at that target, and the Enraged Light neither misses nor
  is resisted. The triggered strikes of weapon procs cast none. The damage logs at 17 to 20 and, on part of the hits, at
  53 to 54, three times as much, as the tooltip's Undead multiplier gives; the count does not read the target's
  creature type.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every streamed run of a wearer with 4 or more pieces: 14
  runs of 12 players (11 warriors and a hunter) in 2PZCz9c1HgtrX7Th, 4YVTAqgG3dtmD8C9, 7NV1DbqPW2vntyRL,
  CJFAZxNLH4BDV9Ka, RFtjqcPd9yKr26wh, TqpYcJC4varfmFkh, ZxWz2nQqXRcYNAbV, c2J4C8kzHaQxnwXW, dVWHbL12mF3gyZ6z,
  gTPNZcJ9mG4dRC2X, mAwQn61fk4YzcL3P, qw3ZzkBmQYW7pXVH, v2aNxQ7mGTdkJryV and xqy2GfPj67XCwrT3. Nine paladin wearers, a
  check only, in 4D3gfhkxKHVT6Pjw, CnVwmNhJgGRXKtq2, DyC3KYB7wAZHz4f6, Hdnh1GJgD4TbQkXL, Ljr8XQmFWJ9n6q4K,
  cLkwmZbzFDKdPhYn, cfqXhmbrD1KC2n84, g4Y2frPx8kBmj9Tq and tWx3Cm4MNaHwvRfq.
- Game build: 1.60.1.70938 to 1.60.1.70959 by report date (70959 from 2026-11-17 15:19 UTC); the hotfix each player ran
  is not recorded
- Date: 2026-11-14 to 2026-11-20
- Pull request: #TBD

## Method

Gear is read per fight from CombatantInfo. A window runs from an applybuff of 1293673 to its removebuff; an apply with no
remove closes 10 s later. A hit is a damage event of the wearer that is no periodic tick, on a hostile target, inside a
window, from a white swing or a spell whose DefenseType in the game files is melee; Thorns, Retribution Aura, Lightning
Shield and Fire Shield are left out. An Enraged Light is a damage event of 1293675, credited to a hit on the same target
(NPC id and instance) from 5 ms before to 100 ms after it. Killing blows and hits within 15 ms of a window's end are
counted apart, and so are the strikes weapon procs trigger (Toofbreak, Wound, Fatal Wound).

Scripts and inputs, in `forever-notices-1121`: `scarlet_final.py`, whose output is `scarlet_final.md`.

## Counts

| Wearers                | Runs | Landed non-killing hits | With an Enraged Light | Share |
| ---------------------- | ---: | ----------------------: | --------------------: | ----: |
| Warriors and a hunter  |   14 |                   6,634 |                 6,577 | 99.1% |
| Paladins, a check only |    9 |                   1,725 |                 1,072 | 62.1% |

Of the 57 landed hits with no Enraged Light, 14 dealt no damage (fully absorbed or blocked). Of the 6,591 Enraged
Lights, 6,577 pair with a landed hit, 14 with none, and none lands outside a window; 998 hits in the windows did not land
(miss, dodge, parry), so at most 14 of them could have cast one. The Enraged Lights log 6,410 hits and 181 crits, no
miss and no resist, a median 31 ms after their hit (2nd percentile 16 ms, 98th 49 ms, longest 100 ms). The 140 weapon
proc strikes inside the windows (Toofbreak 120, Fatal Wound 20) cast none. The 1,043 windows last a median 10.0 s, and
two applies stand at least 15.0 s apart (median 28.5 s). The amounts most logged are 18 (3,247 times), 54 (1,179), 20
(755), 17 (503), 53 (244) and 19 (201).

By what landed, for the warriors and the hunter:

| Hit                  | Landed in a window | With an Enraged Light | Share |
| -------------------- | -----------------: | --------------------: | ----: |
| White swing          |              3,465 |                 3,440 |  0.99 |
| Shield Slam          |                751 |                   748 |  1.00 |
| Revenge              |                750 |                   747 |  1.00 |
| Cleave               |                502 |                   500 |  1.00 |
| Windfury Totem       |                373 |                   365 |  0.98 |
| Whirlwind            |                301 |                   297 |  0.99 |
| Heroic Strike        |                167 |                   165 |  0.99 |
| Shield Bash          |                 90 |                    88 |  0.98 |
| Retaliation          |                 70 |                    66 |  0.94 |
| Bloodthirst          |                 59 |                    57 |  0.97 |
| Victory Rush         |                 37 |                    37 |  1.00 |
| Execute              |                 32 |                    30 |  0.94 |
| Slam                 |                  9 |                     9 |  1.00 |
| Mortal Strike        |                  8 |                     8 |  1.00 |
| Overpower            |                  7 |                     7 |  1.00 |
| Sweeping Strikes     |                  6 |                     6 |  1.00 |
| Pummel               |                  5 |                     5 |  1.00 |
| Hamstring            |                  1 |                     1 |  1.00 |
| Mocking Blow         |                  1 |                     1 |  1.00 |
| Toofbreak (a proc)   |                120 |                     0 |  0.00 |
| Fatal Wound (a proc) |                 20 |                     0 |  0.00 |

## Paladins

The paladins read 62% because their seal and Windfury Totem strikes cast it less often: white swings 0.93 (696 of 749),
Holy Strike 0.97 (144 of 148), Judgement of Command 1.00 (35 of 35), Seals of Command 0.84 (48 of 57), Windfury Totem
0.42 (45 of 106), Seal of Fury 0.17 (73 of 440) and Seal of Righteousness 0.16 (31 of 190). Why those strikes cast it
less often is a separate question this count does not settle, and it sets nothing here.
