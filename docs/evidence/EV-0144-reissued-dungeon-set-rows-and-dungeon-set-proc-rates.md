# EV-0144: Reissued dungeon set 2 pieces take their own set's bonuses, and the rates of six dungeon set procs

- Claim: a reissued dungeon set 2 piece (226772 to 227023) counts toward its reissued set (1667 to 1681, 1778) and
  fires that row's bonuses, never the Classic set's (511 to 519); and the dungeon set procs whose rows state 100% and
  no rate fire at procs per minute, a white swing priced on its weapon's base speed and a spell cast on its cast time:
  Warrior's Resolve at 1, Rogue Armor Energize at 0.8, Crusader's Wrath at about 0.9, The Furious Storm and Nature's
  Bounty's mana at about 1. Nature's Bounty restores its energy off a landed white swing at about 0.9 a minute on the
  form's paw speed, and its rage off a landed melee swing taken at a chance of about 3.3% to 4.0% that does not rise
  with the attacker's interval, 1.21 a minute on that interval; a cast that costs only rage or energy and deals no
  damage rolls nothing. The energy it restores in a bear form comes back with Cat Form. Wild Heart fires at its row's
  5% per landed hit taken, at most once a minute. The Feralheart reissue's pieces count toward its bonuses whichever
  of its four groups they come from.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each table names
- Game build: 1.60.1.70938 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-11 to 2026-11-17
- Pull request: #1086
- Later count: which hits roll Nature's Bounty's energy and rage, Wild Heart, §Count of the 2026-11-21 combat logs
- Later count: the Feralheart reissue's groups, §Count of the 2026-11-22 combat log
- Later count: Nature's Bounty's energy in a bear form, §Count of the 2026-11-23 combat log
- Later count: Nature's Bounty's energy, rage and mana rates, §Count of the 2026-11-13 to 2026-11-25 reports

## Method

A wearer is a player whose gear snapshot in the log holds enough pieces of a set for the proc: 5 of a dungeon set 1,
4 of a reissued dungeon set 2. A fight is cut at each gear snapshot, and a stretch counts while the latest snapshot
holds enough pieces. A run two reports hold, logged or uploaded twice, counts once: where two reports show one player
over the same stretch of time, the report with more events stays. A proc is an event of the proc's spell on the wearer
(an aura applied or refreshed, a heal, a resource gain), counted once when one proc logs several events within 50 ms.
The wearer's own damage, heal and cast events and the hits the wearer takes are grouped where less than 40 ms apart,
and a proc is credited to the nearest group, from 300 ms before it to 60 ms after it.

A swing is a landed white swing or a landed Windfury Totem hit in a group with no spell cast, priced on the base
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
| Warrior's Resolve (Valor, Heroism)             | 450587 |      16 |         720 | 21,667 | 3.32% (3.09–3.57) | 0.94 [0.87, 1.01] |          1 | 4,204 | 0.02%             |                   |
| Rogue Armor Energize (Shadowcraft, Darkmantle) | 27787  |      26 |       1,900 | 70,840 | 2.68% (2.57–2.80) | 0.83 [0.80, 0.86] |          0 | 1,012 | 0%                |                   |

The rows are procs per minute and no flat chance. On swings the chance rises with the weapon's speed while the rate
holds: Crusader's Wrath rolls 2.79% on weapons under 2.2 s (0.92 a minute), 2.87% from 2.2 to 2.9 s (0.73), 4.75% from
2.9 to 3.4 s (0.89) and 4.61% above (0.80); Rogue Armor Energize 0.84 a minute under 2.2 s and 0.81 from 2.2 to 2.9 s.
On casts the chance rises with the cast time: The Furious Storm procs on 2.39% of instants (50 of 2,089) and 4.35% of
casts from 2 to 2.75 s (13 of 299). An instant at 1 a minute on 1.5 s is 2.5%. Crusader's Wrath shows no such rise
on its few hardcasts: 2.75% of instants (183 of 6,655) and 1.31% of casts under 2 s (6 of 458, 95% 0.60 to 2.83),
nearly all Flash of Light.

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
- Sim rate: 1 PPM, observed, on spell casts priced on their cast time, on landed white swings priced on the form's
  speed and on landed melee swings taken priced on the attacker's interval
- Count: §Count of the 2026-11-13 to 2026-11-25 reports, for the energy, the rage and the mana; §Count of the
  2026-11-21 combat logs, for which hits roll the energy and the rage and how the energy ticks; §Count of the
  2026-11-23 combat log, for the energy restored in a bear form

A cast that deals no damage and costs nothing or only rage or energy rolls nothing, an owner ruling of 2026-11-25 that
the count of the 2026-11-13 to 2026-11-25 reports backs. In a bear form no 27782 follows Demoralizing Roar (0 of 3,840
casts), Growl (0 of 4,615) or Challenging Roar (0 of 173), and none follows Barkskin (0 of 1,451). Faerie Fire cast in a
bear form or Cat Form rolls like any spell: 27782 follows 184 of 7,545 casts in a bear form (2.44%) and 45 of 1,766 in
Cat Form (2.55%), where an instant at 1 a minute on 1.5 s is 2.5%.

## Wild Heart (sets 185 and 1667)

- Proc: 1302708, logged as the buff 1302707; Wildheart Raiment's 4-piece and the Feralheart Raiment reissue's
  5-piece
- Sim rate: the row's 5% per landed hit taken, with the row's 60 s cooldown
- Count: §Count of the 2026-11-21 combat logs

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

EV-0148 reads this rate again from a sword-and-shield warrior's combat log.

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

## Count of the 2026-11-21 combat logs

- Sources: combat logs. `WoWCombatLog-112126_112421`, Blackrock Spire: a level 62 druid in Cat Form in six Feralheart
  melee pieces and a level 60 to 61 druid in Dire Bear Form in six Feralheart tank pieces. `WoWCombatLog-112126_022007`,
  Felwood: a level 61 druid mostly in Cat Form and a level 58 druid in caster form, whose pieces the log does not snapshot
- Game build: 1.60.1.71037, on the US and EU realms from 2026-11-20 18:23 UTC; the hotfix each log ran on and whether
  the players relogged are not recorded
- Date: 2026-11-21
- Pull request: #1171

A proc is credited to the druid's own event or the hit the druid took that comes last before it, up to 150 ms before;
the Felwood log's procs trail their swing by up to 121 ms. The form is the druid's power type on its own swing or on
the swing it takes: energy is Cat Form, rage a bear form. A swing is a landed white swing. A hit taken is landed when
it deals damage; a dodge, a parry, a miss and a swing absorbed whole are avoided. A percentage's interval is a Wilson
interval per roll. Both Blackrock Spire druids wear six pieces of one group, and no log holds a druid mixing groups.

Nature's Bounty restores its energy off landed white swings alone. None of the 1,015 landed abilities in Cat Form and
4,461 in a bear form is the last event before a proc, and a swing in caster form procs it as well (1 of 24), with no
tick logged.

A Cat Form proc ticks 4 energy a second, five times (33 of 35 procs; two stop after one and two ticks), and a proc
during the ticks starts five more. The bear in tank pieces logs no tick after any of its 25 procs, while the Felwood
druid's one proc in Dire Bear Form logs five ticks of 4 energy.

A proc of the rage restores 10 rage, in a bear form, Cat Form and caster form alike. A melee ability taken rolls
nothing: in a bear form 1 of 197 landed physical abilities taken (Strike, Cleave, Mortal Strike and the like) and none
of 270 other spells taken is the last hit before a proc.

Wild Heart follows 4.52% of the landed melee swings taken off its cooldown (79 of 1,748, 95% 3.64 to 5.60), 2.53% of
the landed spells and melee abilities taken (9 of 356, 1.34 to 4.73) and none of 914 avoided swings: 4.18% of all
landed hits taken (88 of 2,104, 3.41 to 5.12), across the four druids. One proc follows no hit within 150 ms. No two
procs on a druid are less than 60.0 s apart; the shortest gaps are 60.0, 60.3, 60.7, 61.2 and 64.7 s.

## Count of the 2026-11-22 combat log

- Sources: combat log `WoWCombatLog-112226_153709`, Winterspring: the Felwood log's druid in Cat Form, whose level field
  reads 49 here, in two Feralheart tank pieces and two Feralheart melee pieces by its player's word; the log holds no
  gear snapshot, and which pieces is not recorded
- Game build: 1.60.1.71037, the US and EU build from 2026-11-20 18:23 UTC to the log; the hotfix the log ran on and
  whether the player relogged are not recorded
- Date: 2026-11-22
- Pull request: #1171

Four Feralheart pieces, two from each of two groups, fire the 4-piece: a landed white swing in Cat Form is followed in
the same millisecond by 27784, which ticks 4 energy five times. No group holds four pieces, so pieces of different
groups count toward the set together. Wildheart Raiment's 5-piece, the other set that procs 27784, needs five pieces,
more than the four armor slots left.

## Count of the 2026-11-23 combat log

- Sources: combat log `WoWCombatLog-112326_164634`, Winterspring: the Felwood log's druid, level field 61, with Furor at
  5/5 by its player's word, testing whether energy restored in a bear form reaches Cat Form; which set pieces is not
  recorded
- Game build: 1.60.1.71037; a hotfix went out on 2026-11-23, and the one the log ran on and whether the player relogged
  are not recorded
- Date: 2026-11-23
- Pull request: #1190

The druid's last energy reading in Cat Form is 16, 0.70 s before it leaves. At the 9 to 10 a second its readings in
Cat Form climb by, it leaves with about 22, the energy its player read. It spends 0.31 s in caster form and 324 s in
Dire Bear Form, where two procs of 27784 tick 4 energy ten times, none of it over the bar. It shifts straight back into
Cat Form and reads 69 energy 2.0 s later, 85 another 1.98 s on, then 100; its player read about 62 on the way back.

Without the ticks, Furor's rule gives about 25 on the way back, 22 and 3 for the time in caster form, which would need
44 in the 2 s before the 69. With all of them it gives about 65, which would leave 4 for those 2 s. The readings climb
16 in the 1.98 s after the 69, and 9 to 10 a second before the druid left Cat Form; read back at those rates, it came
back with 49 to 53. Either way it came back with 24 to 40 more than the rule gives without the ticks. Both bear forms'
rows stop energy regeneration (effect 5, -100% energy regeneration) and leave the bar as it is, so nothing but the
ticks added energy in the form. The sim adds every tick.

## Count of the 2026-11-13 to 2026-11-25 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every fight whose gear snapshot holds the set in these
  404 reports: 13PjqwGHQ9KNdVMF, 16dpWnKbmLh4Gca2, 173V9qkfcr2vb6wB, 1Fc86n3dpAK9yVL4, 1Tf47h2MvyczZbVK,
  1ZfhArNnJX3gtPcB, 1bhmLzqVX6A32nRx, 1p9ZzLCPw7dV3rJk, 29ZWyVxcrX1ha6Av, 2RPhk6zxZyLjgn4B, 2jBLczpJmTPAhG6g,
  2q1MandA4HXxL7Ym, 2wPQmVqbdCcZW9gv, 2x7ZjvdbRwFMWKTP, 2zJ3GkaPLDnrKBgQ, 38brBWmH9RLxYGva, 3CjY8yV9pQnrcRFm,
  3GrqwxDX9BMnTR7H, 3Pm4ZHqnMjLNwBvt, 3TdF61x7gMQ28pKX, 3WQwhXcFVtdHjrkZ, 3dpPvAbwfxzGHhnD, 3mavCA2tLRc9qwMy,
  3x9pNAJ8PQ6FkHnK, 49AgNFDbqv87x1yr, 4QDbGNfq7Tt6Wv2a, 4QXfD8qGpN9VYbhW, 4YhFbCdznkj9f6yX, 4aPFr7JvYn8d6RNB,
  4bLKQ2GDHZmh7z9P, 4jXDTv8KyqtNW2Jx, 4rTfkGxMwKWPAXng, 4wT3Fhqxpfky8MKz, 4z2Yy39WAVFDtjN1, 61QxWrdDtkf2VJZm,
  6FZNWyXGzDxHqv2a, 6H1ydbrf2nCY8QWN, 6KPChRcXAyjVfWB4, 6QZLXm7VfRT8jdH3, 6WwaxM19bkXqJTpc, 6dLVNKn81FrzQmyD,
  6nMbzLFKRQhy4P1B, 6vAwndj1FxBcDKq8, 6xdrXcamTAgJy3NW, 6ymdcZvW4YfgTLRt, 78wWZfmgrJYzFXvD, 7BTfKw2HpN1XYWvm,
  7KQd1Lnpqmzjfwv3, 7LrX43AZapzF18JT, 7QkP49vcqVMbHp82, 7Tjny9D2bBpC3v8r, 7VgLNkjPma4ZMAdT, 7faQXRhyZr6V32Fk,
  7mnVXCgpyxkbtZGq, 7nyCwGzLBKqR89xV, 7xD3CJ4kyjBwdngz, 81nxJWvgActT46Km, 81rnYqNT3djhJFRy, 8FWjnvVz6Dx4kYqy,
  8WNhJnqLfwv3HZRY, 8fNXPAmnpr4QcMa1, 8vnBNTdtAFJWLmcr, 91t6GAqXZxcyRVgT, 98zV1QdHZYxAPJrF, 9C2qkgtRP8AWTJav,
  9fWYjH63xNvRb2tk, 9jW8aAGpJmf3KqFr, 9w4ZyL23P1nhpca8, A6qcQ9jtTNXKbvYh, A86pf24tnjvyBLqF, AFgTcYmLZDkHfqwR,
  AT2jv6mC1wL4KFkJ, AWDftL9aJZmMBxHp, AdncDbNmRQqKfFyP, B7xXw4CjR6mDTp39, BCnafpPbAGY3cRNL, BLYHj7vZkNaQzFxG,
  BPVxqXWLc3FJTpMh, BRpZwxyfrA3N62vP, BXLbtpMR6j7fYyKq, BZbHXpPhzy4fLrD3, C4ZTK6azvLD9pAjP, C87y34vf6PMtjVgm,
  CAPt7qmH61WnKTMx, CB8T1LhbRDgYnzJV, CKwpnf78NYPcRjaV, CN1RGx46MPbzr2v9, ChNdrnFtXcqfvwBJ, CxTmQRX2ytGcwA9r,
  D7hPTJ182zvmVxjH, DALjdvgcWTYBF283, DCWdZLkm2cPJbfz9, Dd7gavzxLtNj3CpQ, DzNPfdcnZCKjvT21, F24npWZRxtv8hD76,
  F4JAb8gjwrQDCP2T, F8PLBprjC1xdztQG, Fh9AL83KjJPfRd7b, FmPYdK4qkyztLhvR, Fpqw1tPkWfbJVrjy, G7X1CkT8z6PrfjWM,
  GnM4PzfXLwCDgWTh, GtnhTdayMHCXrA1f, Gx9NyLPCjwDbV3qc, GxDzk8r3aFtdwLTc, H4WY39qjftXdBLVC, H4k1pd8F2b9GmvKM,
  HPWbDfXYdGhK8k3T, HZrpa6fchzV2Q43L, HwGDCPy8Kp1TanZq, J23kMhZrVA18pXBQ, J7v2ZfpQr3HqK1zN, JG89gCcma1Lbxrfn,
  JL1Mph8yakgv4QHX, JbQyngR8NWXKardc, Jj3KHZW7wdFTyLgt, JmxhYRVQvN3b7LTw, JnVDpj3rhNL7xWy2, JtWmqfKxnrpV6YXh,
  K3xmjRCnMpV2c4qr, KNYqn8bcp42rgfJD, KYRHMqa28bVmk3QC, Knrd8xkhRQBpqaPg, KqjwWn3L4VGxrcgT, Kv2JnCDMjhAxzNqg,
  L49qmxFXhVWbt18z, LD6PJnd87qRCTYjN, LNvtnDxd1H6XfVGk, LYw76XG1CFHQNdJz, LhwmgkKDv8bCj6G1, LjYQFGryP4kdqmp3,
  LmFvyJYgtpXArVbk, LqRa728dk9CfhvtP, LrxPVytYBd23XjcK, LwTNrdJ6zjnmCKZg, M4KkWa3jmqGthFQT, MAWRFbLpkha1cynK,
  MQH7fmgdJANkqwxr, MRJHKBWt2TCp17PG, MazXZJYPrAkG7vWm, MhZHzR6WmCtXpqJL, Mtpn21bD4ArgcxjL, N8wyX9FR3TnPBpx2,
  NAnmwq6MHRdFQ1t9, NB2hDQc4p81tPajb, NC2m46DxFhBz1vwd, NFXmgQGxDV8KZr4d, NHzjDpZvLRyPt9hr, NJnx8ktqvg6aZY9R,
  NLVFqvYkxgX9f72d, NRF9y6Mv3gTYfw4j, NVg1CmYpMX7dvy4x, NWR9Y381bdGA6Qjr, NnMA6GgHfRwLJ98Q, PC1M9hHgNxzn7VjF,
  PH6KRGkfYDm91r82, PLD3K2BfrkYxhRWQ, PQfAMVDhrB9Wvxgt, PR8dn7gT932ZbK6H, PVHbqvaXgcWGF6hf, PafbczYWBMjgHCG7,
  PcJyYBT7WGMqpLm8, PpTRzwqQxWC6vXd1, PvBX13kRyjZ7ngdh, QFB3RZGJqw6mC7NP, QHPmM21WVKjvczf6, QJPAtmBc46LqxK8n,
  QL2MaN1g3Htj6ZWF, R16PDpqXMHfF4c28, RATPcCrag87jynDN, RBY9pQPCMTtdhfX8, RJakAQgPXWvfFMnD, RW91FgNq7MBXJjGt,
  RWBahpJkAgVPGNFx, RaTGbvyP3KxH8YwN, RdqHgVKWNmTCrM8Q, RhpdPY6bLq8N3x7k, RwMv31La6FBZ9HDP, T6vkgVaD8YqhyLzf,
  TCLbZW1g4zrPKHyh, TCVybQX6LBa3j9J2, TQYALMf39F2jWh6a, TX8J9PpR2tgDVLCv, TYRPpgvztLCkNx7K, TZpHz98DxwKyMhca,
  Tg1Hcb7z2qAXWkCx, Tm42j9AJcV6hrbC8, Tx7MPH9f8Ga6hdrF, TyGawFf7MhLnB3Dv, V4FWBAJKbHZN1cG9, VKfYJ82cMTktxQzL,
  VvRNkLTq867HMwn3, VzpjcgvbnLCaTZG2, W3dLGjr2q8Pap4Jg, W6qw7QMy4FKTPtXC, WLCqdPQD9AjhFRnJ, WLGy4qAtFxmcrfw1,
  WLw7gXz9A2fBdD4y, WfZ9MhngNAmqJrkG, WmAYZ9MVRPqncTBJ, WwAp7qrmHYNDPzfg, X2BHzGg93h1vbdM7, X4vmdgTRMytH9r36,
  X7BT9Nph1tkdW4y6, X9grzJjpW4dhyCNK, XFRwAyMGbYNBjvJT, XKJkTp93czjy1mYM, XLgaA9MNDHZRcWdk, XQmMHk8y1JxC7Kq3,
  XmxVQnPa1ZNprKqA, XqCydHNfMbG8PxW3, Xz2TRxJf4WHabFQw, XzPyHxjZDvgMR9rA, Y7LzFwD98HmQNkBP, YL2NG4kJ89nQxDKc,
  YWgrM2yL8ZbR6Dc4, YXbnwGzahjNCZygW, Yb83B2kvHA9gRLGn, Yc7B2gVfwztTm9CM, YfW8dGc74x63Mrjv, Yvtw3nWdD7Tz2pxh,
  Z2qcmvGaQWDfHwNp, Z7HXT3RzjCxcqwY2, ZDC1zWJM4YFTNckv, ZGJAML2f9FmDtBRn, ZHhbDAfgNPGw7jKn, ZPJc9zfAjY8kLXqC,
  ZQaAP2GMTY3t9wjX, ZV9CLrWdTcnFRXaQ, ZXwktLq2Vd9pCHKT, ZdrpTz3bCn7FxLv9, Zk39FczKDRAma1P8, Zr83CpF2XTaRtJ7K,
  ZtWQwvfqBYA1k2Dh, Zvz9BQn4AfbGqwRJ, a6fyx7DFnqbCXrkp, aCPtJ98rdW1jvNbm, adXDZKvrNfC91VH6, agVLrW4R2jcwJhmZ,
  ayWm6ZYQx4HfbnNA, bRXTDdAWxkn4hYzG, bWH41fw8CDQAch3K, ba9Xj6JFCyM12YTx, bdmJVrhK126gwfBa, bgycvf8pzGCqNBDm,
  bhwyFTXrA6CaLQzR, bmpBgLZjywNx47PJ, bvZ8GYp2K1fN3DdL, byDnJdBYpGf9jrmQ, byY2HrkRPfvTgWxL, bymT2gMcdD17t4Pw,
  c48NByWfP2dMqmw3, cAvRpFK8Ym7zMbGq, cAy84TxZktvhKLPB, cBGdRD4xLq3fW9FZ, cKDAmhbfNLwB8nCZ, cR4bJMGaVhC38B7X,
  cav7VZrmPbNW3y29, ch87CPWVqaxZYfz3, cp4927aKxtMZJBA3, dPJwm1Bj7rQTxg6p, dbrxgnVXcAP3LFa2, dgxVz1N6JjLk87Hm,
  dmb6aGFDt19zyrXh, dr1BMh6pNxqbFJLf, dwfTvHCVXDy2t74W, f1VTcBRNpybavLwD, fMPL97aCA1ngvKQk, fQaw3NjZ8p6LJWmX,
  fTBkLPW4MY2zN8AJ, fwLdpTX23mjq1DZr, fxqc8bZrgyBDdCWQ, g1jbdaw47PcBWfCp, g8aT4fbm3tVpBCwA, gDjHvdpcKtmTy62h,
  gGBmQx9nC18ZD3d2, gJFfZmCxpnAhkDV9, gNHA97Yt4JqMvm6L, gRrPYMv7waz4jQKb, gXFLdJDqAkZhybzp, gbZhtNpCMT4zwK9m,
  gc2BTyJdxXAR7h4D, gv9V6Ty1Qa3MBZW7, hR7dDX6bPtwrWjf8, hWvTxLktRGzfjqHJ, hmCJ8BNfPkQMdqT7, htxvFBCdVDPrm2pK,
  hxp4F8HyNrQnBG6g, j38kvadrDXTmMNfK, j4gkwbLDmXyMNBPT, jD8B9Qaf7TNtH1bm, jG4ctAQMv9TbrpkY, jKQDqkwmx46ZLBFg,
  jhYXCmK4t9kQpwrH, jxFL3qMGJpZgkdPy, k4tc6gJbWzvLZpNh, kDJ4fpxHjYCbTvnK, kG98jRqpFfBryv3c, kX7FtxzjfYnw69TB,
  kYmZf7TpVjXWrwK6, kcD2LzVGrtZ1xyB6, kjDVcCYFnRdrqTHp, kqd9ZJbhaFKNLGnV, m7FyYPBwvMngHbDk, m7nh21pN9VTPtKdM,
  mAMPfXWB4qQYgv1p, mP7DvAt8gTNYGh96, mQc2yDWVTvXzbj7q, mVtZ6yW8bBvM4T3L, mXYTLpBg7Jt4vH3A, mXv9TLrc73JbCjKy,
  mdwkArzCj27X41DZ, mnVfMcAZkhW7B14w, mw1VHbFfXWyPLqk7, n6zXvjmHDJML8RfF, n7dy3G4VmhZrzF6R, nazpht9wm4NYxGW3,
  nbPCLHj1yRg6mcdz, nqDL2hKfQB4yvjgJ, nrpTgLQdcNKAqk1R, nvYtmN2QTDRgJWXa, p194waK6G3BCxjk7, p1bC8HKyQhV7vWzc,
  p8d9Hrc2L6ZmWn7t, pDFrB9xgKcNVRWqj, pFKdY162wTQaNyqn, pKJj1NyhWd4L3qMD, pLvb3VWqHTnXPCG4, pT4jnLAvtM91QBb6,
  pTPKWhJHcByDXYwj, pVaLq4TbGH2QFtfM, pYwZRftkCBqvXdxF, phDrLBTt2Z4RWm9a, pqyLrZ6Xa1wdQG2V, pzMwa8AY9ZymBbG4,
  q298k6dtBK3VhZXC, qLMkzwp97H8rAJWQ, qWb7x6nhVJNtkwa3, qZPznJCkXrNmjcv7, qh3Y6xGcjgKTQF2n, qrJRyaw31zhH8gPY,
  qwZX7YgNjLdQmTA8, rGjZQfcNHwXLR4m6, rHxPXVAJYRFKnWDZ, rNmcdj9HkhAfv1tD, rhGw29P4WApnNTqF, rmKx9LgCkPfWvp7h,
  rqpT3mxb2LAwP1dt, t7mDkLy2G6fVhXHJ, t9PLr4H8gdkXncRZ, tPbh63Rc4J8xdv1Q, tPnNz4ajZhcvAVpf, tQfhj2JbzZkYXFvx,
  thDJ8pbqRLrW6QMw, twN4W6cRgDyMA2vT, v3wnGDTZQ1FW7gcA, v6PApQm2XfYG7tHN, vBMpqfA2Jgx9Cckb, vDdAGQta1M6NHmCK,
  vJmHxwgrWjR2Kfnk, vYjL3q712CyTGkQh, vc718aZH43GMpKrP, vhHkwZ1cQJYnfLDt, vqNtgpTXwQfF6JWb, vwxP1TaLzG8A4m7h,
  vyna9XkWc2GpD1rV, wAbTZ4cMa8t3NXH2, wCJabgqLfmd8v1tK, wr3M9DfXFQZTqj72, x64Wd3Z8A2TKCDXL, xFhXRtJ4KDVHqvmW,
  xGX3ZTV1jpwHyaP4, xLVtMAzT2NmFy6gG, xNqdkFAXbJwyHhcp, xR6ZGyakX8czKpbQ, xYvZt3LdKCwz8H7T, xZtGRMyAX2FDP1qY,
  xb49gaHAVryMZh2t, xd64GKNcb7VCqarf, xdK1Pac8CynkFDYB, xqL7rjDwAhJYzVfQ, xt7BwRzWVvbKJTXp, y46bmVXZzPrvwG81,
  y72Cf4pTxQ1WFbaV, yCvdhZ2q9xLY3VaP, yHtZF9jrbgXJnRVP, yfdCN1t8FmzPxYMD, yfwhbHv8JBVXdGCF, ygMD7q4p8aWk1nG3,
  ymnQpDv9xMfPqK8k, yq7Aw3B2QkFzgNja, z2L8JXqBRwfvTbPN, z3VLyCTfvwRrdKbN, z9dafy3kFWZK4Gb8, zB1bQy6c3vFW4wdq,
  zGvXTP7gQ8JjrNmY, zRtFPvfgr8qDjkAc, zWPFXGdt4hVkq3DB, zXQyax7t1CZ8Tq9Y, zcdXFhYfnRjLB2TA, znGLPcrvxA8KbXF4,
  zpvfR1gAaKmbWTMy, zxtbh1BmNVC8W2FH and zy3r8qhGAH2BRp41. The report index lists 431 reports with a wearer; of the
  other 27, 15 were not fetched and 12 hold no fight the count kept
- Game build: 1.60.1.70938 to 1.60.1.71091 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-25
- Pull request: #1227

A wearer is a player whose gear snapshot in the report index holds 4 or more pieces of the Feralheart reissue (226772
to 226803) or 5 or more of Wildheart Raiment, and a fight counts when that snapshot holds the set. A run two reports
hold counts once: a fight of a player whose name matches a fight already counted and that starts within 2 minutes of it
is left out. An event a report logs more than once counts once. The form comes from the druid's own applications and
removals of Cat Form (768), Bear Form (5487), Dire Bear Form (9634) and Moonkin Form (24858). Travel Form (783),
Aquatic Form (1066) and Tree Form (775) are read the same way, and the swings taken and casts in them count in the All
rows alone.

An energy proc is an application or refresh of 27784 with a landed white swing of the druid's from 30 ms before it to
20 ms after; each landed white swing is priced on the form's paw, 1 s in Cat Form and 2.5 s in a bear form. A rage proc
is a resource change of 27783 from 30 ms before a landed melee swing taken (ability 1) to 20 ms after it. Each landed
swing taken is priced on its attacker instance's median gap between its swings on the wearer, over the gaps from 500 ms
to 6 s, clipped to 1 to 4 s, and an attacker with fewer than 3 such gaps is left out. The median gap of an attacker
that is slowed or dual-wields is off its weapon's speed. A mana proc is a resource change of 27782 from 60 ms before to
300 ms after a cast of a spell with the magic defense type, leaving out form shifts and the melee swing's own cast
event; the cast is priced on its time from `begincast` to `cast` and at least 1.5 s. A rate is the procs over the
summed prices over 60. An interval, on a rate or a percentage, is 2,000 bootstrap draws of whole wearer-reports. The
scripts are in `requests/forever-feralheart-1125` in the log parser's folder.

The energy reads about 0.9 a minute in both forms.

| Form        | Procs | Landed white swings | Per swing | PPM on the paw    | Wearer-reports | Players |
| ----------- | ----: | ------------------: | --------: | ----------------- | -------------: | ------: |
| Cat Form    |   468 |              31,912 |     1.47% | 0.88 [0.77, 0.98] |            107 |      81 |
| A bear form | 2,181 |              58,381 |     3.74% | 0.90 [0.86, 0.94] |            211 |     138 |

The rage's chance per landed swing taken stays near 3.3% to 4.0% across attacker intervals from 1 s to 4 s and does
not rise with them, so 1 a minute on the attacker's interval does not fit: on that interval the rage reads 1.21 a
minute, 1.68 on the fastest attackers and 0.52 on the slowest. Priced on a flat 2 s per swing taken, it reads 1.14
[1.10, 1.18].

| Form         | Procs | Landed swings taken | Per swing          | PPM on the attacker's interval | Wearer-reports | Players |
| ------------ | ----: | ------------------: | ------------------ | ------------------------------ | -------------: | ------: |
| A bear form  | 5,900 |             155,056 | 3.81% [3.66, 3.95] | 1.21 [1.16, 1.26]              |            199 |     131 |
| Cat Form     |    49 |               1,725 | 2.84% [1.79, 3.77] | 0.94 [0.59, 1.26]              |             75 |      56 |
| Caster form  |   136 |               3,649 | 3.73% [2.58, 5.20] | 1.17 [0.81, 1.64]              |            189 |     141 |
| Moonkin Form |    47 |               1,164 | 4.04% [2.80, 5.57] |                                |                |         |
| All          | 6,135 |             161,632 | 3.80%              | 1.21 [1.16, 1.26]              |            315 |     222 |

"1 a minute on it" is the chance 1 a minute gives on each swing's attacker interval, averaged over the bucket's swings.

| Attacker interval | Procs | Swings taken | Per swing | 1 a minute on it | PPM on it         |
| ----------------- | ----: | -----------: | --------: | ---------------: | ----------------- |
| 1 to 1.5 s        |   999 |       30,036 |     3.33% |            1.98% | 1.68 [1.56, 1.82] |
| 1.5 to 2 s        | 1,150 |       29,799 |     3.86% |            2.85% | 1.36 [1.26, 1.46] |
| 2 to 2.5 s        | 3,718 |       94,880 |     3.92% |            3.48% | 1.13 [1.07, 1.19] |
| 2.5 to 3 s        |   196 |        4,847 |     4.04% |            4.48% | 0.90 [0.77, 1.04] |
| 3 to 3.5 s        |    43 |        1,200 |     3.58% |            5.27% | 0.68 [0.50, 0.87] |
| 3.5 to 4 s        |    29 |          870 |     3.33% |            6.46% | 0.52 [0.34, 0.70] |

The mana reads about 1 a minute in caster form and Moonkin Form. The casts in a bear form include Growl, Demoralizing
Roar and Barkskin, which roll nothing (§Nature's Bounty (sets 185 and 1667)), so the bear forms' rate reads low and
pulls All under the caster forms'.

| Form         | Procs |   Casts | Per cast | PPM on the cast time | Wearer-reports | Players |
| ------------ | ----: | ------: | -------: | -------------------- | -------------: | ------: |
| Caster form  | 2,642 |  88,089 |    3.00% | 1.09 [1.05, 1.13]    |            410 |     287 |
| Moonkin Form |   488 |  16,157 |    3.02% | 1.03 [0.97, 1.08]    |             34 |      31 |
| Cat Form     |   120 |   5,135 |    2.34% | 0.90 [0.75, 1.07]    |            115 |      90 |
| A bear form  |   193 |  19,092 |    1.01% | 0.40 [0.35, 0.46]    |            199 |     128 |
| All          | 3,469 | 129,722 |    2.67% | 0.98 [0.94, 1.02]    |            428 |     297 |

Split at 1.60.1.71037 (2026-11-20 18:23 UTC), the reports before it read 0.97 [0.65, 1.02] for the energy in Cat Form,
1.00 [0.90, 1.09] in a bear form, 1.18 [1.05, 1.31] for the rage and 0.99 [0.85, 1.11] for the mana, and the reports
from it 0.87 [0.75, 0.99], 0.88 [0.83, 0.92], 1.21 [1.16, 1.26] and 0.98 [0.94, 1.02].
