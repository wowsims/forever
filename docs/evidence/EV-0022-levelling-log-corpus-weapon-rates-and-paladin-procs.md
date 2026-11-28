# EV-0022: Weapon rates, paladin procs and Two-Handed Weapon Specialization's Holy reach from a levelling log corpus

- Claim: the procs per minute of Ripsaw, The Hand of Antu'sul, Jang'thraze the Protector, Thrash Blade and Bigger Rock on
  Bigger Stick, and that a druid's form specials roll no chance-on-hit weapon; Redoubt rolls on every melee swing taken,
  Shieldsworn's spell block comes with every Holy Strike while its mana and Holy Shield damage need a landed strike, Seal
  of Fury's shield is a flat share in the open world, and Touch of the Grave has a 15 s cooldown; Two-Handed Weapon
  Specialization's Holy part skips Exorcism, Hammer of Wrath and Retribution Aura and reaches Consecration.
- Sources: combat log corpus of 44 logs on 1.60.1, 18 levelling logs with StatLedger and 26 from other players; logs from
  the beta realms (11 to 17 October) counted apart
- Game build: 1.60.1
- Date: 2026-11-12 for the weapon rates; not recorded for the rest
- Pull request: #982
- Later count: Bigger Rock on Bigger Stick, §Count of the 2026-11-14 to 2026-11-15 reports

The #958 log of the same date is a different pass (EV-0021).

## Weapon rates

Each rate is the procs over the landed hits of the weapon, white swings and specials alike, at the weapon's speed, as the
measured rows already in the override table count.

| Weapon                                        | Procs / hits | PPM (95%)           | Row | Players           |
| --------------------------------------------- | ------------ | ------------------- | --- | ----------------- |
| Ripsaw 9478, Wound 1292573                    | 185 / 2018   | 2.04 (1.77 to 2.34) | 2   | 1, main-hand hits |
| The Hand of Antu'sul 9639, Thunder Clap 13532 | 91 / 991     | 2.04 (1.64 to 2.44) | 2   | 1                 |
| Jang'thraze the Protector 11086, absorb 11657 | 181 / 1972   | 2.90 (2.50 to 3.30) | 2.9 | 1                 |
| Thrash Blade 17705                            | 61 / 1425    | 0.95 (0.74 to 1.21) | 1   | 1, main-hand hits |

Bigger Rock on Bigger Stick answers whether an unstated rate is procs per minute or a flat chance: it reads 11.7% a
swing at its own 3.4 speed, 8.6% at a bear's 2.5 and 2.5% at a cat's 1.0. Procs per minute fits (deviance 2.9 on 5
degrees of freedom) and a flat chance does not (49 on 5).

Long sims with the weapons equipped fire the registered procs as measured. The Hand of Antu'sul on the default
Protection paladin follows 8.0% of white swings, 8.3% of Reckoning's swings, 9.7% of Holy Strikes, 9.4% of Judgement of
Fury hits and 9.5% of Seal of Fury hits (2 PPM at 2.7 speed is 9.0%).

A druid's form specials roll no chance-on-hit weapon: 2 procs in 7736 hits, from 3 druids and 2 weapons.

| Druid                                | Maul    | Swipe (target hits) | Claw    | Lacerate |
| ------------------------------------ | ------- | ------------------- | ------- | -------- |
| One druid                            | 0 / 441 | 0 / 5401            | 0 / 248 | 0 / 309  |
| Two more druids, with another weapon | 1 / 869 | 0 / 238             | 1 / 180 | 0 / 50   |

Their white swings roll at the weapon's rate. Combat enchants on those hits are unmeasured.

## Redoubt, Shieldsworn, Seal of Fury and Touch of the Grave

| Reading                                                                              | Measured                                                                                             |
| ------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------- |
| Redoubt on avoided swings (dodged, parried, missed, fully blocked or fully absorbed) | 152 of 1394, 10.9%, about forty paladins                                                             |
| Redoubt on swings that dealt damage                                                  | 364 of 3501, 10.4%                                                                                   |
| Redoubt on melee abilities                                                           | 17 of 138                                                                                            |
| Redoubt on ranged attacks                                                            | 0 of 62                                                                                              |
| Shieldsworn's spell block on avoided Holy Strikes                                    | 46 of 46 on four paladins (15 of 15 on the two source logs), and 61 of 61 strikes made with a shield |
| Shield Specialization mana and Holy Shield damage on avoided strikes                 | 0 of 31 and 0 of 12                                                                                  |
| The same on landed strikes                                                           | 141 of 143 and 50 of 52                                                                              |
| Seal of Fury's shield in the open world                                              | 4077 of 4078 logged shields at the flat share                                                        |
| Seal of Fury's shield in a dungeon                                                   | scaled by expected creature health, see below                                                        |
| Touch of the Grave cooldown                                                          | 57 players, 707 drains, none closer than 14.99 s                                                     |

Inside a dungeon Seal of Fury's shield scales by expected creature health at a base level over the mob's level, the dungeon's
lowest mob level being the base in five dungeons: Scarlet Monastery, base 30 on two players, and Shadowfang Keep, the
Stockade, Gnomeregan and Ruins of Lordaeron on one player each, their base levels not recorded. Against mobs of the paladin's own level it stays flat, 118 of 118, three cases on two players.

The sim reproduces two of these readings. Redoubt rolls after 41 of 410 dodged, parried or missed boss swings, where the
sim before the change gave it no roll (0 of 444). Shieldsworn's spell block comes with 409 of 409 dodged or missed
strikes.

## Two-Handed Weapon Specialization's Holy part

Holy spell multipliers with the talent at 3/3 and a two-hander, once mobs with their own damage reduction are split out:

| Holy spell                 | Measured multiplier                                  |
| -------------------------- | ---------------------------------------------------- |
| Consecration               | 1.021, 1.036 and 1.049 to 1.062 for Ranks 1, 2 and 3 |
| Exorcism, Retribution Aura | 1.001 to 1.010 and 0.997 to 1.000 by rank            |
| Hammer of Wrath            | 0.991 (n = 77, one player)                           |

The talent at 3/3 gives 1.06. Holy Forgefire was measured without the bonus too.

## Count of the 2026-11-14 to 2026-11-15 reports

- Sources: 76 public reports on forever.warcraftlogs.com (zone 3000 listing, 2026-11-14 17:49 to 2026-11-15 07:00 UTC),
  every fight Bigger Rock on Bigger Stick was worn in: 17RQWNTLJAgGBhrx, 1PcrGTyatDAFx8Yk, 49dLXZ3g7NhzqABb,
  6H1QkhV2qANXvKFc, 7AmNxqPcC8LdZWtD, 7dWytj1K24G6ckJ9, 8ndg4zGq9mfVWvQk, 9rPwVnazMdWYxj8X, BLXHnFRfrhPGYa76,
  BcDaTm9KbgyXZJLC, C64x3BnYWMQZkA8T, CRgvLNQYd8y3c2m6, FXaN14xP7VhbDpQM, Fx94JmkLncTXBPQy, G7KxnMVrJw19gv2q,
  GTJw7YXampLrWcgM, GjmZfc8aLPFQVXKN, GyZmK2b8c9LMtDdw, L7VBmjgyHRCzqYa2, LQvTgHXDf3Vq1zYA, LyjHDc4nCYmBMRNX,
  PCW1JNXVx2B8nDyh, PDT2yACNzd4LMZ98, QHjqK4G3pzgRWaN6, QajbfyAcJdRCkP92, QxRZFwrkKgJPdB82, RmTM8XL4aBJY12Ak,
  RtZaGw26kvVcAjK8, TFYgK8vGyQc1MzrL, TkfYVFdKBth9C3mR, VThFn29CGbX713cy, WbfQZpaPAkL8tzdv, XvjpYan3qxF9WRgz,
  XyPY7tNcdb1fFDCZ, YAVgyZp9FNMKCh4D, YBDwzFPf9WH7pgv8, YHXAcJxWG89pMFgf, YXgt6fkMdBTKDrAn, Z4w7HKFPvcrTWd3M,
  ZH7zNkqT9j3G1Qm4, Za21hDrqWXn6Rz3M, avm98fMJj3Zc7hP1, b2YKZcFzWPtkxfpg, bcLyFrwNCaH4f9DZ, c6DWA1CV2tdmRFqL,
  cHMQhVq76BR9YFNm, cdFkXVKH1xm2JhpW, cmnDzYT96jLbRyPV, d6TMB1P784WZHVhF, dJa2G7PHZg9wt6QX, dYzgZkCNJFQAXtBP,
  fLyrVbpZAWxzQFYk, fgw1ckK8HTAvQ3pq, garVmPTCWnAh2HfG, grjxzfTdLQKbnBwZ, hRZr7VPfazMjxJym, hatJzXGTk7qjmLPf,
  hf1pF42aPHMv9kLX, j2FY7nBLCcakV9TK, j39RX7DxTqgzWHmB, k26rdYnGDNZ3b81P, m3Fjgrt124A7advV, mpRWCjLgVDxZb9Tr,
  nfLmGxNDqJpyXB6k, ngAGhLYavywCM2F3, ntaAc893bkmBd1Cg, nvXGCNxKYDHrPgjc, pb7dQrAw8CGWjMk4, qKzDTJjwvXM264Q7,
  qP2MRFvJ16NBZGgm, qyVW9rhmwbnpfc4x, tpWza62JNFrTm9hH, v6hC8nkRyQjbgJHX, yYnGv1tHqWDV8bMA, zH3TmcCQRb4Pa8tk,
  zvkWdnQb3t7yCmKZ
- Game build: 1.60.1.70938. The hotfix pushes inside the window (2026-11-14 19:52 to 2026-11-15 06:22 UTC) add 29
  items and change housing and profession spells only, none of the items or proc spells measured here; whether the
  players relogged is not recorded.
- Date: 2026-11-14 and 2026-11-15
- Pull request: #1032

The count is EV-0119 §Method, run on Bigger Rock on Bigger Stick. The rate is the pooled rate of the table, all its
rows, rounded to the nearest 0.05 step, which the owner sets to 1.5 (pooled 1.55). Its proc spell is Bigger Rock to the
Face.

| Item                        | Class   | Hits   | Procs / isolated hits | Chance | PPM [95%]         | Wearers |
| --------------------------- | ------- | ------ | --------------------- | ------ | ----------------- | ------- |
| Bigger Rock on Bigger Stick | Paladin | cast   | 350 / 3862            | 9.1%   | 1.60 [1.45, 1.77] | 28      |
| Bigger Rock on Bigger Stick | Paladin | white  | 530 / 6487            | 8.2%   | 1.44 [1.33, 1.56] | 28      |
| Bigger Rock on Bigger Stick | Shaman  | cast   | 590 / 6061            | 9.7%   | 1.72 [1.59, 1.85] | 41      |
| Bigger Rock on Bigger Stick | Shaman  | white  | 1496 / 17002          | 8.8%   | 1.55 [1.48, 1.63] | 42      |
| Bigger Rock on Bigger Stick | Warrior | cast   | 159 / 1880            | 8.5%   | 1.49 [1.29, 1.73] | 7       |
| Bigger Rock on Bigger Stick | Warrior | white  | 270 / 3474            | 7.8%   | 1.37 [1.22, 1.54] | 7       |
| Bigger Rock on Bigger Stick | all     | pooled | 3395 / 38766          | 8.76%  | 1.55 [1.50, 1.60] | –       |
