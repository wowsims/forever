# EV-0229: Hellfire's burn on the caster, and Insight rolling once per cast

- Claim: Hellfire's burn on the caster (effect 2 of the channel row, 11684 at rank 3) takes the caster's Fire damage
  done modifiers and crits for the caster's spell crit damage, ×2 with Ruin and ×1.5 without. Enchant Weapon - Insight
  8216 rolls its trigger 1248758 once per cast, on a heal or a hit, however many targets the cast reaches.
- Sources: a combat log of a group (Flexi): three warlocks at level 60 (Destruction and Demonology with Ruin 5,
  Affliction without Ruin, Affliction and Destruction with Ruin 5 on Hellfire rank 2), and three Insight wearers, a
  Restoration and Balance druid, a Restoration shaman and a Holy priest below level 60
- Game build: 1.60.1.71091 until 2026-11-25 19:11 UTC and 1.60.1.71139 from 20:16 UTC, after hotfix push 114399,
  relogged
- Date: 2026-11-25 14:33 to 2026-11-26 03:51 UTC
- Pull request: not opened yet

## Hellfire's burn

Each damage line of the channel's spell id with the warlock as source and target is one burn. The base is the line's
amount before crit, resist and absorb; resisted lines are none, and absorbed lines keep their base.

| Warlock                                    | Burns | Crits                    | Base, non-crit | Crit amount ÷ non-crit amount |
| ------------------------------------------ | ----: | ------------------------ | -------------: | ----------------------------: |
| Destruction, Ruin 5                        | 1,081 | 154 (14.2% [12.3, 16.5]) |          286.5 |                          2.00 |
| Affliction, no Ruin                        | 1,175 | 173 (14.7% [12.8, 16.9]) |          218.0 |                          1.50 |
| Affliction and Destruction, Ruin 5, rank 2 |   168 | 21 (12.5% [8.3, 18.4])   |          147.6 |                          2.00 |

The Destruction warlock's base is 1.30 times the row's 220.6 at his spell power, the product of his Fire damage done
auras (Touch of Fire 1.15, Master Demonologist's imp 1.10, Soul Link 1.03). 911 of his 927 non-crit burns were partly
absorbed. The crit chance of the burn against the warlocks' spell crit was not read: the log carries no crit sheet.

## Insight

Each heal or direct hit of a wearer counts as one landing, landings of one spell within 50 ms as one. A landing rolls
when the 45 s cooldown has run out since the wearer's last gain of buff 1299796, and it gains when an apply or refresh
of 1299796 falls within 60 ms after it.

| Spell      | Landings with the cooldown ready | Gains | Rate               |
| ---------- | -------------------------------: | ----: | ------------------ |
| Chain Heal |                              128 |    43 | 33.6% [26.0, 42.1] |
| Wrath      |                              105 |    38 | 36.2%              |
| Riptide    |                              101 |    33 | 32.7%              |
| Regrowth   |                               60 |    20 | 33.3%              |

A Chain Heal reaches about 2.9 targets, where a roll for each target at 35% would gain on about 72% of casts. The
smallest gap between two gains is 45.0 s. The priest below level 60 gained on fewer of his Lesser Heals (5 of 54) and
Renews (6 of 47), and the druid on 70 of 292 Rejuvenations (24.0%); neither is explained.
