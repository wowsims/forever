# EV-0144: Reissued dungeon set 2 pieces take their own set's bonuses, and the rates of five dungeon set procs

- Claim: a reissued dungeon set 2 piece (226772 to 227023) counts toward its reissued set (1667 to 1681, 1778) and
  fires that row's bonuses, never the Classic set's (511 to 519); and the dungeon set procs whose rows state 100% and
  no rate fire at procs per minute, a white swing priced on its weapon's base speed and a spell cast on its cast time:
  Warrior's Resolve at 1, Rogue Armor Energize at 0.8, Crusader's Wrath at about 0.9, The Furious Storm and Nature's
  Bounty's mana at about 1.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each table names
- Game build: 1.60.1.70938 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-11 to 2026-11-17
- Pull request: #1086

## Method

A wearer is a player whose gear snapshot in the log holds enough pieces of a set for the proc: 5 of a dungeon set 1,
4 of a reissued dungeon set 2. A fight is cut at each gear snapshot, and a stretch counts while the latest snapshot
holds enough pieces. A run two reports hold, logged or uploaded twice, counts once: where two reports show one player
over the same stretch of time, the report with more events stays. A proc is an event of the proc's spell on the wearer
(an aura applied or refreshed, a heal, a resource gain), counted once when one proc logs several events within 50 ms.
The wearer's own damage, heal and cast events and the hits the wearer takes are grouped where less than 40 ms apart,
and a proc is credited to the nearest group, from 300 ms before it to 60 ms after it.

A swing is a landed white swing or a landed Windfury Totem strike in a group with no spell cast, priced on the base
speed of the main-hand weapon in the snapshot, or on the harmonic mean of both hands' speeds for a dual wielder, whose
swings are not split by hand. A druid swings a form's paw whatever the weapon, so the one druid in Shadowcraft Armor is
left out of the swing rates. A cast is a group holding exactly one cast of a spell with the magic defense type and no
swing, priced on the time between its `begincast` and `cast` events and at least 1.5 s. The rate is the procs over the
sum of those times over 60. An interval on a rate is 2,000 bootstrap draws of whole wearer-reports, and one on a
percentage is a Wilson interval per roll. No stretch is dropped as swapped away; dropping the gaps a wearer at 0.8 a
minute would see less than once in a thousand moves no rate by more than 0.02, Darkmantle's 0.74 to 0.81 excepted.

The scripts are `extract.py`, `rates.py`, `stretch.py`, `hastesplit.py`, `set_rows.py` and `evtables.py`.

## Which set a reissued piece counts toward

Every player in the indexed reports who wears a dungeon set 2 piece wears a reissued one: 586 wearers of the ids
226772 to 227023 across the 20,191 reports with gear snapshots, none of the Classic ids 21994 to 22113 and none of the
72 pieces the reissued `ItemSet` rows list (226815 to 227040). The wearers of 4 or more pieces fire the reissued row's
bonuses at its piece counts:

| Set (reissued id)                | Report           | Fights                                  | Date  | Pieces | Events of the set proc spells                       |
| -------------------------------- | ---------------- | --------------------------------------- | ----- | ------ | --------------------------------------------------- |
| Battlegear of Heroism (1778)     | vYpj9gc1tk6zGFRX | 1, 4-8, 10                              | 11-14 | 4      | 89 of 450589                                        |
| Battlegear of Heroism (1778)     | 4Tc3yBNVYn7wW8RJ | 1, 8, 12, 14-16, 19-20, 22, 25, 28, 31  | 11-15 | 4      | 160 of 450589                                       |
| Battlegear of Heroism (1778)     | VtjvpP8bwcgCYTXF | 1                                       | 11-16 | 4      | 36 of 450589                                        |
| Battlegear of Heroism (1778)     | aDHTPGq6htJFx8Nw | 1                                       | 11-16 | 4      | 29 of 450589                                        |
| Battlegear of Heroism (1778)     | Wy6TdamAk4ZxJ8MC | 1, 5-6, 8, 10, 22-23, 25-30             | 11-17 | 6      | 143 of 450589                                       |
| Beastmaster Armor (1669)         | xRK37aCvtVfj26kL | 111, 130, 137, 143                      | 11-16 | 4      | 118 of 450554                                       |
| Beastmaster Armor (1669)         | mCYWgaFVAdrby9NX | 1, 5, 7                                 | 11-16 | 4      | 7 of 450554                                         |
| Beastmaster Armor (1669)         | F2vN8aVjBz1PJ4Q3 | 1                                       | 11-16 | 4      | 8 of 450554                                         |
| Darkmantle Armor (1676)          | gBDMLv8J1K7nqkpd | 159, 163                                | 11-11 | 6      | 44 of 27788                                         |
| Darkmantle Armor (1676)          | H2T7NDqBgm6LxGYK | 1                                       | 11-15 | 6      | 38 of 27788                                         |
| Darkmantle Armor (1676)          | 7fz9Kk8rY6jyQAbL | 3                                       | 11-16 | 6      | 11 of 27788                                         |
| Feralheart Raiment (1667)        | tQfhj2JbzZkYXFvx | 19, 22-23, 25, 28, 41, 44-45, 47-48, 52 | 11-13 | 6      | 3 of 27782, 98 of 27783, 34 of 27784, 51 of 1302707 |
| Sorcerer's Regalia (1671)        | JBxW16RqGXMPYH3d | 10, 12                                  | 11-17 | 6      | 3 of 27868, 39 of 450554                            |
| Soulforge Armor (1673)           | zmDY4LVX7N8vh9aR | 1                                       | 11-14 | 6      | 25 of 27499, 6 of 1302715                           |
| Soulforge Armor (1673)           | Bg7J4jxckP6pdvrZ | 1, 5                                    | 11-16 | 6      | 36 of 27499, 2 of 1302715                           |
| The Five Thunders (1679)         | 63pnYPtNJk1X7HFB | 1                                       | 11-16 | 5      | 25 of 27775, 2 of 1302717                           |
| The Five Thunders (1679)         | zM4Dd8QKwBATnJR9 | 1, 3, 5, 9, 14, 18, 29, 31, 33          | 11-17 | 6      | 83 of 27775, 3 of 1302717                           |
| Vestments of the Virtuous (1675) | 7tbWfYqJFd4vzL3p | 1, 3                                    | 11-17 | 4, 6   | 34 of 450554                                        |

450589 is Warrior's Resolve as rows 189 and 1778 fire it; the Classic row 511 fires 27418, which no report holds.
450554 is Sudden Insight, the 4-piece of rows 1669, 1671 and 1675; the Classic Beastmaster row 515 fires 27786 at 4
pieces, which no report holds either. 1302715 Rebuked, 1302717 Electrocute and 1302707 Wild Heart are the 5-piece
bonuses of rows 1673, 1679 and 1667, and no Classic row has them. Rogue Armor Energize 27788 is the 4-piece of both
Darkmantle rows, so its three wearers do not tell the rows apart, and no Deathmist wearer has 4 pieces.

## Rates

| Proc                                           | Row    | Wearers | Swing procs | Swings | Per swing         | Swing PPM         | Cast procs | Casts | Per cast          | Cast PPM          |
| ---------------------------------------------- | ------ | ------: | ----------: | -----: | ----------------- | ----------------- | ---------: | ----: | ----------------- | ----------------- |
| Crusader's Wrath (Lightforge, Soulforge)       | 450625 |      12 |         334 |  9,299 | 3.59% (3.23–3.99) | 0.85 [0.76, 0.95] |        193 | 7,217 | 2.67% (2.33–3.07) | 1.06 [0.77, 1.22] |
| The Furious Storm (Elements, Thunders)         | 450626 |       4 |          55 |  1,398 | 3.93% (3.03–5.09) | 0.66 [0.59, 1.29] |         70 | 2,654 | 2.64% (2.09–3.32) | 0.96 [0.78, 1.01] |
| Nature's Bounty, mana (Wildheart, Feralheart)  | 450608 |       7 |           0 |    622 | 0% (0–0.61)       |                   |         64 | 2,827 | 2.26% (1.78–2.88) | 0.81 [0.56, 1.04] |
| Warrior's Resolve (Valor, Heroism)             | 450587 |      16 |         720 | 21,667 | 3.32% (3.09–3.57) | 0.94 [0.87, 1.01] |          1 | 4,204 | 0.02%             |                   |
| Rogue Armor Energize (Shadowcraft, Darkmantle) | 27787  |      26 |       1,900 | 70,840 | 2.68% (2.57–2.80) | 0.83 [0.80, 0.86] |          0 | 1,012 | 0%                |                   |

The rows are procs per minute and no flat chance. On swings the chance rises with the weapon's speed while the rate
holds: Crusader's Wrath rolls 2.79% on weapons under 2.2 s (0.92 a minute), 2.87% from 2.2 to 2.9 s (0.73), 4.75% from
2.9 to 3.4 s (0.89) and 4.61% above (0.80); Rogue Armor Energize 0.84 a minute under 2.2 s and 0.81 from 2.2 to 2.9 s.
On casts the chance rises with the cast time: Nature's Bounty restores mana on 1.85% of instants (36 of 1,943), 2.49%
of casts under 2 s (13 of 523), 3.75% of casts from 2 to 2.75 s (9 of 240) and 4.96% of longer ones (6 of 121); The
Furious Storm on 2.39% of instants (50 of 2,089) and 4.35% of casts from 2 to 2.75 s (13 of 299). An instant at 1 a
minute on 1.5 s is 2.5%. Crusader's Wrath shows no such rise on its few hardcasts: 2.75% of instants (183 of 6,655)
and 1.31% of casts under 2 s (6 of 458, 95% 0.60 to 2.83), nearly all Flash of Light.

A cast that costs mana rolls the spell rows whether or not it deals damage. On the Lightforge wearers a seal cast is
followed by Crusader's Wrath 2.0% to 3.3% of the time (Seal of Command 31 of 941, Seal of Righteousness 25 of 839,
Seal of Fury 13 of 437), Holy Shield 4.42% (18 of 407) and Cleanse 3.91% (7 of 179), beside Consecration's 3.02%
(50 of 1,658). One Judgement rolls once, off the global cooldown: 2.36% (66 of 2,800), an instant's chance. The
Furious Storm follows 2.80% of Magma Totems (4 of 143) and 2 of 57 Windfury Totems. A cast that costs nothing rolls
nothing, on the global cooldown or off it: Swift Judgement 0 of 65, Sanctified Orb 0 of 35, Water Shield 0 of 75.
Melee and ranged abilities roll neither row: none of 2,796 Holy Strikes and 561 Hammers of Wrath is followed by
Crusader's Wrath, and none of 444 Stormstrikes by The Furious Storm. Hits taken and avoided swings roll nothing
(0 of 3,010 and 0 of 811 for Crusader's Wrath).

The rows that state a chance draw the same line. On the Magister's Regalia wearers Sudden Insight follows 6.99% of
Counterspells (10 of 143), which cost mana off the global cooldown, and none of 160 Arcane Powers, 146 Presences of
Mind, 141 casts of Eureka! or 91 Evocations, which cost nothing; Evocation is on the global cooldown.

Haste leaves the chance alone, so the rate is priced on the weapon's base speed. Warrior's Resolve reads 0.93 a
minute [0.84, 1.07] on swings under Flurry (331 of 10,777) and 0.95 [0.83, 1.05] without it (389 of 10,890); Rogue
Armor Energize 0.84 [0.80, 0.88] under Slice and Dice (1,504 of 56,399) and 0.81 [0.73, 0.90] without it (396 of
14,441).

Crusader's Wrath's swings read below its casts, 0.85 against 1.06. The gap, 0.20, has a 95% interval of -0.11 to 0.37
over whole wearer-reports, and five of the twelve read the other way, so the logs do not show two rates. One rate
fitted to the swings and the casts together is 0.92 [0.79, 0.99].

The sets whose rows state a chance fire at it. Sudden Insight follows 5.00% of Auto Shots fired (610 of 12,208,
95% 4.62 to 5.40) and 4.75% of landed white swings (245 of 5,163) on the Beaststalker and Beastmaster wearers, whose
row 450577 states 5%, and Vestments of the Devout's 450576 restores mana on 4.91% of Renews (188 of 3,832) and 4.73%
of Flash Heals (47 of 994).

## Crusader's Wrath (sets 188 and 1673)

- Proc: 450625, logged as the buff 27499; Lightforge Armor's 5-piece and the Soulforge Armor reissue's 4-piece
- Sim rate: 0.9 PPM, observed: one rate fitted to the swings and the casts reads 0.92 [0.79, 0.99], an interval
  wider than 10% of it below

| Report           | Fights                                             | Date  | Class   | Set              | Swing procs | Swings | Swing PPM | Speeds        | Cast procs | Casts | Cast PPM |
| ---------------- | -------------------------------------------------- | ----- | ------- | ---------------- | ----------: | -----: | --------: | ------------- | ---------: | ----: | -------: |
| NXMZV9xjvWpB7GtP | 10, 13                                             | 11-13 | Paladin | Lightforge Armor |          25 |    415 |      1.07 | 3.3, 3.5, 3.6 |         14 |   585 |     0.96 |
| NXMZV9xjvWpB7GtP | 107                                                | 11-13 | Paladin | Lightforge Armor |          17 |    478 |      1.07 | 2             |          7 |   235 |     1.19 |
| NXMZV9xjvWpB7GtP | 122                                                | 11-13 | Paladin | Lightforge Armor |          17 |    647 |      0.69 | 2.3           |         12 |   372 |     1.27 |
| AYwLfvqh9QyNj6zJ | 1                                                  | 11-15 | Paladin | Lightforge Armor |           8 |    114 |      1.31 | 3, 3.3        |          2 |   133 |     0.52 |
| zp9yVrtfTNZJCv7g | 9                                                  | 11-15 | Paladin | Lightforge Armor |           8 |    248 |      0.57 | 3.4           |         12 |   245 |     1.96 |
| LR9WfMTdbr3ZNBFH | 1                                                  | 11-15 | Paladin | Lightforge Armor |          14 |    294 |      0.87 | 3.3           |          3 |   298 |     0.40 |
| dnDg7j9kRw6WtbML | 5, 10, 26, 35-36, 48-49, 52, 55, 58, 61, 64, 67-68 | 11-16 | Paladin | Lightforge Armor |         150 |   4454 |      0.86 | 1.8, 3.3      |         92 |  2967 |     1.23 |
| DNc6QALhxmGbCH3Y | 1                                                  | 11-16 | Paladin | Lightforge Armor |           0 |      6 |      0.00 | 2.7           |          7 |   530 |     0.49 |
| 79QPzLcgZCKbJBkH | 1                                                  | 11-16 | Paladin | Lightforge Armor |          14 |    227 |      1.23 | 3             |         11 |   200 |     2.20 |
| zM4Dd8QKwBATnJR9 | 1, 3                                               | 11-17 | Paladin | Lightforge Armor |          15 |    647 |      0.60 | 2.3           |          4 |   344 |     0.46 |
| JBxW16RqGXMPYH3d | 6, 10, 12                                          | 11-17 | Paladin | Lightforge Armor |          32 |    767 |      0.83 | 3             |         19 |   739 |     1.03 |
| zmDY4LVX7N8vh9aR | 1                                                  | 11-14 | Paladin | Soulforge Armor  |          12 |    496 |      0.60 | 2.4           |          4 |   224 |     0.71 |
| Bg7J4jxckP6pdvrZ | 1, 5                                               | 11-16 | Paladin | Soulforge Armor  |          22 |    506 |      1.09 | 2.4           |          6 |   345 |     0.69 |

## The Furious Storm (sets 187 and 1679)

- Proc: 450626, logged as the buff 27775; The Elements' 5-piece and The Five Thunders reissue's 4-piece
- Sim rate: 1 PPM, observed; four wearers

| Report           | Fights                         | Date  | Class  | Set               | Swing procs | Swings | Swing PPM | Speeds | Cast procs | Casts | Cast PPM |
| ---------------- | ------------------------------ | ----- | ------ | ----------------- | ----------: | -----: | --------: | ------ | ---------: | ----: | -------: |
| 8zB1PhRQy64jHkXY | 1                              | 11-15 | Shaman | The Elements      |          10 |    201 |      0.83 | 3.6    |          5 |   219 |     0.88 |
| 1bgXr3HRqG2cTnjQ | 1, 4                           | 11-17 | Shaman | The Elements      |           3 |     35 |      1.77 | 2.9    |          9 |   389 |     0.77 |
| 63pnYPtNJk1X7HFB | 1                              | 11-16 | Shaman | The Five Thunders |           4 |     72 |      1.08 | 3.1    |         18 |   544 |     1.02 |
| zM4Dd8QKwBATnJR9 | 1, 3, 5, 9, 14, 18, 29, 31, 33 | 11-17 | Shaman | The Five Thunders |          38 |   1090 |      0.58 | 3.6    |         38 |  1502 |     1.00 |

## Nature's Bounty (sets 185 and 1667)

- Proc: 450608; Wildheart Raiment's 5-piece and the Feralheart Raiment reissue's 4-piece. A spell cast restores
  27782's mana, a landed melee swing 27784's energy over 5 s, and a melee hit taken 27783's rage
- Sim rate: 1 PPM on spell casts, observed; the swing and the hit taken are not modelled

The one Feralheart wearer, in Dire Bear Form through report tQfhj2JbzZkYXFvx, gained 27783's rage on 28 of 882 melee
hits taken with no event of the wearer's own within 40 ms (3.17%, 95% 2.21 to 4.55) and 27784's energy on 24 of 512
landed white swings. Four more swings of the Wildheart wearers procced the energy.

| Report           | Fights                                  | Date  | Class | Set                | Swing procs | Swings | Swing PPM | Speeds | Cast procs | Casts | Cast PPM |
| ---------------- | --------------------------------------- | ----- | ----- | ------------------ | ----------: | -----: | --------: | ------ | ---------: | ----: | -------: |
| tQfhj2JbzZkYXFvx | 19, 22-23, 25, 28, 41, 44-45, 47-48, 52 | 11-13 | Druid | Feralheart Raiment |           0 |    512 |      0.00 | 3.2    |          3 |   339 |     0.35 |
| LahKC3yT6JQRwtMn | 1, 4-5, 8-9, 12, 15                     | 11-14 | Druid | Wildheart Raiment  |           0 |     20 |      0.00 | 1.5    |          5 |   516 |     0.36 |
| 7mnVXCgpyxkbtZGq | 3, 10-11, 17                            | 11-15 | Druid | Wildheart Raiment  |           0 |     83 |      0.00 | 1.5    |         22 |   806 |     0.93 |
| X6Hp4gDwv7zftxMC | 2                                       | 11-15 | Druid | Wildheart Raiment  |           0 |      0 |      0.00 |        |          5 |   181 |     1.00 |
| zpvfR1gAaKmbWTMy | 1                                       | 11-15 | Druid | Wildheart Raiment  |           0 |      2 |      0.00 | 2.8    |         10 |   341 |     1.10 |
| T6VDdFhPxQ43pHM2 | 1                                       | 11-16 | Druid | Wildheart Raiment  |           0 |      0 |      0.00 |        |          5 |   220 |     0.78 |
| CxTmQRX2ytGcwA9r | 2                                       | 11-16 | Druid | Wildheart Raiment  |           0 |      5 |      0.00 | 2.1    |         14 |   424 |     1.18 |

## Warrior's Resolve on Battlegear of Heroism (set 1778)

- Proc: 450587, logged as 450589, the 4-piece of the reissue; EV-0136 measured it on Battlegear of Valor
- Sim rate: 1 PPM; measured 0.94 PPM [0.87, 1.01], 16 wearers, 720 procs on 21,667 swings, of which one Heroism
  wearer 0.94 PPM [0.87, 1.02], 399 procs on 13,019 swings

| Report           | Fights                                 | Date  | Class   | Set                   | Swing procs | Swings | Swing PPM | Speeds                 | Cast procs | Casts | Cast PPM |
| ---------------- | -------------------------------------- | ----- | ------- | --------------------- | ----------: | -----: | --------: | ---------------------- | ---------: | ----: | -------: |
| vYpj9gc1tk6zGFRX | 1, 4-8, 10                             | 11-14 | Warrior | Battlegear of Heroism |          82 |   2332 |      1.08 | 1.95 (both hands)      |          0 |    73 |     0.00 |
| 4Tc3yBNVYn7wW8RJ | 1, 8, 12, 14-16, 19-20, 22, 25, 28, 31 | 11-15 | Warrior | Battlegear of Heroism |         132 |   4468 |      0.91 | 1.95 (both hands)      |          0 |   170 |     0.00 |
| VtjvpP8bwcgCYTXF | 1                                      | 11-16 | Warrior | Battlegear of Heroism |          33 |   1014 |      1.00 | 1.95 (both hands)      |          0 |   201 |     0.00 |
| aDHTPGq6htJFx8Nw | 1                                      | 11-16 | Warrior | Battlegear of Heroism |          21 |    911 |      0.71 | 1.95 (both hands)      |          1 |    68 |     0.59 |
| Wy6TdamAk4ZxJ8MC | 1, 5-6, 8, 10, 22-23, 25-30            | 11-17 | Warrior | Battlegear of Heroism |         131 |   4294 |      0.94 | 1.95 (both hands)      |          0 |   225 |     0.00 |
| D463FZBX1C9maJ7b | 7                                      | 11-14 | Warrior | Battlegear of Valor   |           8 |    400 |      0.55 | 2.2                    |          0 |   175 |     0.00 |
| 2wMznDRvmkpbtHxA | 7                                      | 11-15 | Warrior | Battlegear of Valor   |          19 |    481 |      1.32 | 1.8                    |          0 |   180 |     0.00 |
| DGFd8KtQk19hXabN | 1                                      | 11-15 | Warrior | Battlegear of Valor   |          35 |    702 |      1.07 | 2.8                    |          0 |   357 |     0.00 |
| 7FBaHRhMVKjC1QbT | 1                                      | 11-15 | Warrior | Battlegear of Valor   |          21 |    437 |      1.31 | 2.2                    |          0 |   243 |     0.00 |
| dBNfgVDLMmnc8h4v | 10                                     | 11-15 | Warrior | Battlegear of Valor   |          11 |    535 |      0.59 | 2.1                    |          0 |   168 |     0.00 |
| FN8krKfCb3BXWmwV | 1, 4, 39                               | 11-15 | Warrior | Battlegear of Valor   |          16 |    800 |      0.54 | 2.2, 2.7               |          0 |   450 |     0.00 |
| rnfRqHt9Vk8g1JB7 | 1                                      | 11-16 | Warrior | Battlegear of Valor   |          16 |    312 |      1.10 | 2.8                    |          0 |   180 |     0.00 |
| 7fz9Kk8rY6jyQAbL | 1                                      | 11-16 | Warrior | Battlegear of Valor   |          24 |    720 |      0.92 | 2.16, 2.7 (both hands) |          0 |    30 |     0.00 |
| AfJ6DYbh9dmkgHZL | 1                                      | 11-16 | Warrior | Battlegear of Valor   |          44 |   1246 |      0.97 | 2.19 (both hands)      |          0 |    29 |     0.00 |
| rfg8k3Nw1P7TAHx6 | 1                                      | 11-16 | Warrior | Battlegear of Valor   |          24 |    558 |      0.96 | 2.7                    |          0 |   416 |     0.00 |
| wCHpdtBrqyFQJfLz | 1, 3                                   | 11-16 | Warrior | Battlegear of Valor   |          35 |    753 |      1.00 | 2.8                    |          0 |   473 |     0.00 |
| aDHTPGq6htJFx8Nw | 1                                      | 11-16 | Warrior | Battlegear of Valor   |           8 |    288 |      0.64 | 2.6                    |          0 |   178 |     0.00 |
| WTJGk6CmbMP1V8Hf | 1                                      | 11-17 | Warrior | Battlegear of Valor   |           6 |    376 |      0.53 | 1.8                    |          0 |   176 |     0.00 |
| KnYw8BkyhFdtHpVr | 1                                      | 11-17 | Warrior | Battlegear of Valor   |          28 |    577 |      1.21 | 2.4                    |          0 |   257 |     0.00 |
| 1GWa2FTw7BDnhA9v | 1                                      | 11-17 | Warrior | Battlegear of Valor   |          26 |    463 |      1.25 | 2.7                    |          0 |   155 |     0.00 |

## Rogue Armor Energize (sets 184 and 1676)

- Proc: 27787, logged as 27788; Shadowcraft Armor's 5-piece and the Darkmantle Armor reissue's 4-piece
- Sim rate: 0.8 PPM; measured 0.83 PPM [0.80, 0.86], 26 wearers, 1,900 procs on 70,840 swings

Every rogue here dual-wields, and the swings are priced on both hands' speeds together, with no split by hand. The
same pricing reads Warrior's Resolve at 0.94.

| Report           | Fights                                                         | Date  | Class | Set               | Swing procs | Swings | Swing PPM | Speeds                        | Cast procs | Casts | Cast PPM |
| ---------------- | -------------------------------------------------------------- | ----- | ----- | ----------------- | ----------: | -----: | --------: | ----------------------------- | ---------: | ----: | -------: |
| gBDMLv8J1K7nqkpd | 159, 163                                                       | 11-11 | Rogue | Darkmantle Armor  |          42 |   2381 |      0.72 | 1.47 (both hands)             |          0 |    51 |     0.00 |
| H2T7NDqBgm6LxGYK | 1                                                              | 11-15 | Rogue | Darkmantle Armor  |          34 |   1658 |      0.84 | 1.47 (both hands)             |          0 |    24 |     0.00 |
| 7fz9Kk8rY6jyQAbL | 3                                                              | 11-16 | Rogue | Darkmantle Armor  |           9 |    475 |      0.57 | 1.98 (both hands)             |          0 |    11 |     0.00 |
| NXMZV9xjvWpB7GtP | 10                                                             | 11-13 | Rogue | Shadowcraft Armor |          21 |   1270 |      0.67 | 1.47 (both hands)             |          0 |     3 |     0.00 |
| bg169rGdXTzDhLnj | 1, 3                                                           | 11-13 | Rogue | Shadowcraft Armor |          52 |   1324 |      0.95 | 2.48 (both hands)             |          0 |    22 |     0.00 |
| 1ZyhaVDdFtTYWMcC | 1, 4, 10                                                       | 11-13 | Rogue | Shadowcraft Armor |          48 |   1673 |      0.84 | 1.84, 2.16 (both hands)       |          0 |    45 |     0.00 |
| 123jYCvkVGrHfXWt | 1-2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 23, 27, 29               | 11-14 | Rogue | Shadowcraft Armor |         152 |   5295 |      0.85 | 2.02 (both hands)             |          0 |    80 |     0.00 |
| z9XbfMRh21Fg7QjN | 1                                                              | 11-14 | Rogue | Shadowcraft Armor |          59 |   1808 |      0.84 | 2.34 (both hands)             |          0 |    17 |     0.00 |
| ZfbY6RgWjx9hmpGz | 1, 4, 9, 12-13                                                 | 11-14 | Rogue | Shadowcraft Armor |         147 |   4881 |      0.77 | 2.34 (both hands)             |          0 |    28 |     0.00 |
| pNW7aGKnzHP1MftT | 3                                                              | 11-14 | Rogue | Shadowcraft Armor |           4 |    314 |      0.31 | 2.46 (both hands)             |          0 |     1 |     0.00 |
| 1z4vRDqGdM2cVPXk | 1, 7-8                                                         | 11-14 | Rogue | Shadowcraft Armor |          45 |   1634 |      0.76 | 2.09, 2.34 (both hands)       |          0 |    17 |     0.00 |
| CfHvwytKJ6P4q7ag | 1, 3, 7, 17, 21                                                | 11-14 | Rogue | Shadowcraft Armor |         103 |   3929 |      0.78 | 2.02 (both hands)             |          0 |    41 |     0.00 |
| Z3nAqFVzTrmcHQWN | 1                                                              | 11-14 | Rogue | Shadowcraft Armor |           5 |    203 |      1.06 | 1.4 (both hands)              |          0 |     2 |     0.00 |
| wPtHK4F6BYd8GVNq | 1                                                              | 11-15 | Rogue | Shadowcraft Armor |          17 |    676 |      0.69 | 2.19 (both hands)             |          0 |     0 |     0.00 |
| 2Z4G3r9pfqJtCM6W | 7                                                              | 11-15 | Rogue | Shadowcraft Armor |          17 |    793 |      0.59 | 2.19 (both hands)             |          0 |     4 |     0.00 |
| VjBTMHXxFvbthkw2 | 1                                                              | 11-15 | Rogue | Shadowcraft Armor |          25 |   1174 |      0.81 | 1.58 (both hands)             |          0 |     0 |     0.00 |
| 2wMznDRvmkpbtHxA | 1, 5, 7                                                        | 11-15 | Rogue | Shadowcraft Armor |          64 |   2265 |      0.83 | 2.02, 2.09 (both hands)       |          0 |    29 |     0.00 |
| XLvj723CgZTJGy4h | 1                                                              | 11-15 | Rogue | Shadowcraft Armor |          29 |   1298 |      0.91 | 1.47 (both hands)             |          0 |     2 |     0.00 |
| 6YnFw7gxv4kJAKaC | 2, 6-7, 9, 12                                                  | 11-15 | Rogue | Shadowcraft Armor |          52 |   1916 |      0.81 | 2.02 (both hands)             |          0 |    26 |     0.00 |
| La43JRmZxb2fnWTM | 1                                                              | 11-15 | Rogue | Shadowcraft Armor |          50 |   1215 |      1.14 | 2.16 (both hands)             |          0 |    11 |     0.00 |
| dBNfgVDLMmnc8h4v | 10                                                             | 11-15 | Rogue | Shadowcraft Armor |          32 |    968 |      0.95 | 2.02, 2.16 (both hands)       |          0 |    15 |     0.00 |
| VwC7K3aq4YzmMjQR | 19                                                             | 11-15 | Rogue | Shadowcraft Armor |           3 |    179 |      0.50 | 2.02 (both hands)             |          0 |     2 |     0.00 |
| Daj7G8xFX3bA2CmW | 1, 5, 7, 11-13, 16, 22                                         | 11-15 | Rogue | Shadowcraft Armor |          77 |   2921 |      0.82 | 1.88, 1.95 (both hands)       |          0 |    26 |     0.00 |
| zp9yVrtfTNZJCv7g | 5                                                              | 11-15 | Rogue | Shadowcraft Armor |          39 |   1816 |      0.82 | 1.58 (both hands)             |          0 |     2 |     0.00 |
| APgTZrbzNvykx1JG | 1, 4, 7-8, 15-16, 18, 21, 26, 28, 34                           | 11-16 | Rogue | Shadowcraft Armor |         138 |   4747 |      0.86 | 2.02 (both hands)             |          0 |    36 |     0.00 |
| 7fz9Kk8rY6jyQAbL | 3                                                              | 11-16 | Rogue | Shadowcraft Armor |          19 |    569 |      0.91 | 2.19 (both hands)             |          0 |     1 |     0.00 |
| 7cpmjq6DTBNLwFhG | 5, 8, 48, 50, 63, 73, 76, 78, 82, 86, 89, 92, 94, 97, 100, 102 | 11-16 | Rogue | Shadowcraft Armor |          83 |   2460 |      0.96 | 2.06, 2.19, 2.46 (both hands) |          0 |    28 |     0.00 |
| rRnVJhQFa9zf7NqK | 1-2                                                            | 11-16 | Rogue | Shadowcraft Armor |          76 |   2490 |      0.85 | 2.16 (both hands)             |          0 |    25 |     0.00 |
| Fz2wJmfGKPnvTy9h | 3                                                              | 11-16 | Rogue | Shadowcraft Armor |           2 |     70 |      0.82 | 2.09 (both hands)             |          0 |     0 |     0.00 |
| dnDg7j9kRw6WtbML | 58, 61                                                         | 11-16 | Rogue | Shadowcraft Armor |          33 |   1035 |      1.08 | 1.75, 2.16 (both hands)       |          0 |    23 |     0.00 |
| ptvYCVPymJXH21kR | 9, 11, 14                                                      | 11-16 | Rogue | Shadowcraft Armor |          99 |   4338 |      0.87 | 1.54, 1.92 (both hands)       |          0 |   101 |     0.00 |
| R9fbLyJAW1tDQnFT | 2                                                              | 11-16 | Rogue | Shadowcraft Armor |          30 |   1158 |      0.77 | 2.02 (both hands)             |          0 |    19 |     0.00 |
| h8vCa2P43nbH7JwX | 1                                                              | 11-17 | Rogue | Shadowcraft Armor |          35 |   1080 |      0.83 | 2.35 (both hands)             |          0 |    35 |     0.00 |
| PayThRGFLKmbNcnp | 7, 11                                                          | 11-17 | Rogue | Shadowcraft Armor |          40 |   1508 |      0.80 | 1.77, 2.06 (both hands)       |          0 |     3 |     0.00 |
| NTJzVdMZBQDaqG83 | 2                                                              | 11-17 | Rogue | Shadowcraft Armor |          16 |    782 |      0.60 | 2.06 (both hands)             |          0 |     0 |     0.00 |
| 1bgXr3HRqG2cTnjQ | 29                                                             | 11-17 | Rogue | Shadowcraft Armor |          20 |    713 |      0.77 | 2.19 (both hands)             |          0 |     1 |     0.00 |
| GByQ7g9ZPkqxFJXh | 1, 7, 9, 11, 15-16, 18                                         | 11-17 | Rogue | Shadowcraft Armor |         183 |   7824 |      0.91 | 1.54 (both hands)             |          0 |   281 |     0.00 |
