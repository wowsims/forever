# EV-0194: How much armor periodic physical damage meets, by the row's attributes

- Claim: The ticks of a spell whose row sets Attributes[6] 0x200000 ignore armor (Garrote, Rupture, Rip, Rake, Pounce
  Bleed and the warrior's Rend); the ticks of a spell whose row sets Attributes[0] 0x40000 without it meet a tenth of the
  target's armor (the weapons' Rends, Barman Shanker's among them); the ticks of a spell whose row sets neither meet the
  whole armor (Wound, Bigger Rock to the Face, Rock to the Face). The bleed mechanic does not decide it: the Rocks are
  bleeds and meet the whole armor. No name of either bit says this, so the rule is a correlation across these spells.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the player streams on disk: for the weapons, every
  stream of a player wielding one, from 2,265 reports; for the class spells, 2,500 of the 13,241 report stream files,
  drawn at random (the sorted file list shuffled with Python's `random.seed(1)`), of which 999 reports hold ticks of
  the class spells. The reports are too many to list here
- Game build: 1.60.1.70862 to 1.60.1.71091 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-07 to 2026-11-25 03:31 UTC
- Pull request: #1214

## Method

Each periodic tick of the spell by the player that is not a crit and has no overkill, absorb or block is paired with the
mean share the same player's normal white swings on the same target lost within 1.5 s of it, and the pairs are grouped
by that share w to the nearest tenth. A tick's share lost is 1 minus its amount over its amount before mitigation. A
white-swing share w means armor A with A/K = x = w/(1 - w), so a tenth of the armor takes 0.1x/(1 + 0.1x) off and the
whole armor takes w. Cells with fewer than 20 pairs are left out, and the tables stop at 0.6. In the groups from 0.7
up, which hold 1 to 161 pairs a weapon, the ticks of Barman Shanker, Kang the Decapitator, Hameya's Slayer, Gargoyle
Shredder Talons and Drakefang Butcher still lose about what a tenth of the armor takes. Those of Blood Talon and Killmaim,
Bleeding Crescent and Silithid Ripper, Bloodrazor, Glutton's Cleaver, Flesh Piercer, Bloodpike and Gatorbite Axe lose
42% to 83%, which no share of the armor fits. Bloodspiller and Gutwrencher have no pairs there.

A tick's amount before mitigation already holds the player's damage bonuses, so a share lost below zero is a bonus no
armor took back: Garrote's and Rupture's talents. No cell of the class spells' table reaches what a tenth of the armor
takes.

## Ticks that ignore armor

Share each spell's ticks lose, by white-swing share. The two rows below are the warrior's Rend pairs' own; the other
spells' pairs differ from them by 2 points at most.

| Spell            |  Pairs | w 0.1 | w 0.2 | w 0.3 | w 0.4 | w 0.5 | w 0.6 |
| ---------------- | -----: | ----: | ----: | ----: | ----: | ----: | ----: |
| Rend (warrior's) | 27,702 | -0.8% | -2.0% | -1.0% | -0.1% |  3.7% |  2.6% |
| Rake             | 17,743 | -0.6% | -0.4% |  0.0% |  0.2% |  5.6% |  7.6% |
| Garrote          | 13,918 | -2.6% | -2.9% | -3.7% | -3.6% |  0.0% |  1.4% |
| Rupture          |  9,135 | -8.8% | -7.0% | -8.7% | -6.8% | -5.7% |  4.9% |
| Rip              |  6,891 | -1.4% | -1.1% | -0.2% |  0.1% |  2.3% |       |
| Pounce Bleed     |    318 |       | -0.1% | -0.2% | -0.4% |       |       |
| A tenth of armor |        |  1.3% |  2.8% |  4.1% |  5.9% |  8.4% | 11.8% |
| The whole armor  |        | 11.2% | 22.1% | 29.9% | 38.4% | 47.7% | 57.2% |

## Ticks that meet a tenth of the armor

Each cell is the share the ticks lose, then what a tenth of the armor takes, in percent. Drakefang Butcher's Rend 14118
gives 10 pairs, which agree with a tenth.

| Spell (item)                                    |  Pairs |     w 0.1 |     w 0.2 |     w 0.3 |     w 0.4 |     w 0.5 |       w 0.6 |
| ----------------------------------------------- | -----: | --------: | --------: | --------: | --------: | --------: | ----------: |
| Rend 13318 (Blood Talon, Killmaim)              |  1,187 |           | 1.8 / 2.8 | 2.8 / 3.7 | 7.7 / 6.3 |           |             |
| Rend 16403 (Bleeding Crescent, Silithid Ripper) |    224 |           | 1.7 / 2.7 | 3.1 / 4.0 | 7.2 / 6.2 |           |             |
| Rend 16406 (Hameya's Slayer)                    |    506 |           | 1.9 / 2.9 | 4.9 / 4.8 | 5.1 / 5.9 |           |             |
| Rend 17153 (Kang the Decapitator)               |  1,981 |           | 2.7 / 2.9 | 4.3 / 4.7 | 5.4 / 5.7 | 8.9 / 8.2 |             |
| Rend 17504 (Bloodrazor)                         |  4,650 |           | 1.9 / 2.9 | 3.9 / 4.8 | 5.1 / 5.8 | 9.0 / 8.7 | 16.0 / 11.5 |
| Rend 18075 (Glutton's Cleaver)                  | 10,561 | 0.1 / 1.1 | 1.0 / 2.5 | 2.7 / 4.2 | 4.7 / 5.9 | 8.5 / 8.0 |             |
| Rend 18078 (Flesh Piercer)                      |    256 |           | 1.5 / 2.8 | 4.4 / 3.7 | 4.8 / 6.2 |           |             |
| Rend 18200 (Bloodspiller)                       |  1,622 | 1.0 / 1.2 | 2.0 / 2.7 | 3.3 / 3.8 | 9.9 / 6.2 |           |             |
| Rend 18202 (Bloodpike)                          |  1,084 |           | 1.8 / 2.7 | 4.1 / 3.7 | 5.8 / 5.9 |           |             |
| Rend 21949 (Gatorbite Axe)                      | 12,940 | 1.0 / 1.5 | 2.3 / 3.0 | 4.3 / 4.7 | 5.1 / 5.9 | 8.1 / 8.3 | 12.8 / 11.3 |
| Rend 1294426 (Gutwrencher)                      |    104 |           |           |           | 2.6 / 5.9 |           |             |
| Rend 1300787 (Barman Shanker)                   | 21,494 | 1.0 / 1.4 | 2.3 / 2.9 | 4.1 / 4.7 | 5.4 / 5.9 | 9.8 / 8.7 | 11.7 / 11.5 |
| Rend 1320824 (Gargoyle Shredder Talons)         |    444 |           | 2.6 / 3.0 | 3.1 / 4.7 | 4.5 / 5.9 | 9.0 / 9.2 |  9.0 / 11.2 |

## Ticks that meet the whole armor

Each cell is the share the ticks lose, then the white-swing share, which is what the whole armor takes, in percent. The
white-swing share is worked back from the tenth the count printed to three places, so it holds to half a point.

| Spell (item)                                                  |   Pairs |       w 0.1 |       w 0.2 |       w 0.3 |       w 0.4 |       w 0.5 |       w 0.6 |
| ------------------------------------------------------------- | ------: | ----------: | ----------: | ----------: | ----------: | ----------: | ----------: |
| Wound 17407 (The Blackrock Slicer)                            |  18,325 | 11.6 / 12.4 | 21.9 / 24.2 | 32.5 / 33.5 | 38.0 / 39.0 | 47.8 / 49.1 | 55.7 / 56.3 |
| Bigger Rock to the Face 1306816 (Bigger Rock on Bigger Stick) | 114,576 | 11.0 / 14.0 | 21.7 / 21.7 | 31.5 / 32.0 | 37.0 / 37.7 | 47.6 / 47.8 | 58.4 / 57.2 |
| Rock to the Face 1309967 (Rock on Stick)                      |  51,746 |  7.0 / 13.2 | 17.5 / 21.1 | 28.2 / 32.0 | 32.4 / 37.2 | 46.1 / 47.8 | 62.2 / 57.7 |

## Attributes

| Spells                                    | Attributes[0] 0x40000 | Attributes[6] 0x200000 | Mechanic                 | Armor     |
| ----------------------------------------- | :-------------------: | :--------------------: | ------------------------ | --------- |
| Garrote, Rupture, Rip, Rend (warrior's)   |          set          |          set           | bleed                    | none      |
| Pounce Bleed                              |                       |          set           | bleed                    | none      |
| Rake                                      |          set          |          set           | bleed, on the tick alone | none      |
| The weapons' Rends above                  |          set          |                        | bleed                    | a tenth   |
| Wound                                     |                       |                        | none                     | the whole |
| Bigger Rock to the Face, Rock to the Face |                       |                        | bleed                    | the whole |
