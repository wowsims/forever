# EV-0003: Chance-on-hit weapon rates from end-game dungeon logs

- Claim: the procs per minute of 18 chance-on-hit weapons whose game files state no rate.
- Sources: Warcraft Logs, the report and fight of each row's first run in the table
- Game build: not recorded
- Date: 2026-11-12
- Pull request: not recorded
- Later count: Stoneslayer, §Count of the 2026-11-14 to 2026-11-15 reports

Each rate is the procs over the landed hits of the weapon carrying them, white swings and specials alike, as the sim's
weapon proc listens, at the weapon's speed (a druid's at the paw's). Shield Slam, Shield Bash and Shank hits are left
out, since they roll a chance-on-hit weapon about once in a thousand.

| Weapon                      | Item   | PPM | Measured                                                                                          | First report and fight              |
| --------------------------- | ------ | --- | ------------------------------------------------------------------------------------------------- | ----------------------------------- |
| Lifeforce Dirk              | 10750  | 1.5 | 161 procs on landed hits, 2 end-game dungeon runs, 2 players, off hand                            | jT7NDMWFfhKk1bdz/1 and 1 more run   |
| Bloodfist                   | 11744  | 3.2 | 91 procs on landed hits, 3 end-game dungeon runs, 2 players, off hand                             | fFMPzh2LjnRTtaAk/1 and 2 more runs  |
| Force of Magma              | 11803  | 2   | 398 procs on landed hits, 3 end-game dungeon runs, 2 players                                      | fgMRXWJKHzaLt2Gh/1 and 2 more runs  |
| Drakefang Butcher           | 12463  | 1.5 | 146 procs on landed hits, 1 end-game dungeon run, 1 player                                        | LQG6jWCnAP1gwvdt/1                  |
| Keris of Zul'Serak          | 12582  | 1.7 | 295 procs on landed hits, 6 end-game dungeon runs, 2 players, dual wielded                        | cvRTFb1qhGBKNkCD/45 and 5 more runs |
| Seeping Willow              | 12969  | 0.9 | 78 procs on landed hits, 4 end-game dungeon runs, 2 players, cat form, at the 1.0 s paw           | T79gRkJz4QVGbyYW/9 and 3 more runs  |
| Venomspitter                | 13183  | 1.7 | 157 procs on landed hits, 4 end-game dungeon runs, 3 players, dual wielded                        | 6AX1yxVMGKT7zbqL/1 and 3 more runs  |
| Demonshear                  | 13348  | 1.1 | 67 procs on landed hits, 1 end-game dungeon run, 1 player                                         | DpGg6kA8nyMQ1Hzh/1                  |
| Gravestone War Axe          | 13983  | 1.9 | 116 procs on landed hits, 3 end-game dungeon runs, 1 player                                       | NRK61HZcDdJkrwXG/5 and 2 more runs  |
| Gatorbite Axe               | 17730  | 1.2 | 188 procs on landed hits, 11 end-game dungeon runs, 2 players                                     | kW9V4CMKfmBbtrRH/3 and 10 more runs |
| Verdant Keeper's Aim        | 17753  | 3.5 | 6839 procs on landed ranged hits, 33 end-game dungeon runs, 14 players; melee hits do not roll it | DpGg6kA8nyMQ1Hzh/1 and 32 more runs |
| Princess Theradras' Scepter | 17766  | 1.8 | 307 procs on landed hits, 2 end-game dungeon runs, a shaman and a cat druid                       | x3qQt7yf9mAgbTrZ/1 and 1 more run   |
| Fist of Stone               | 17943  | 3.2 | 105 procs on 1086 landed autos, Holy Strikes and Judgements, 1 end-game dungeon run, 1 player     | vQZLCzhPNatk7r3d/1                  |
| Stoneshatter                | 18388  | 1.7 | 146 procs on landed hits, 3 end-game dungeon runs, 1 player                                       | YH7n2xZqPKzWjQrR/6 and 2 more runs  |
| Weed Whacker                | 274270 | 1.1 | 119 procs on landed hits, 1 end-game dungeon run, 1 player                                        | LQG6jWCnAP1gwvdt/1                  |
| Cremation Longbow           | 275441 | 2.3 | 292 procs on landed hits, 3 end-game dungeon runs, 1 player                                       | NgRFKH1PWzykqC9Q/12 and 2 more runs |
| Lightcrest Hammer           | 276419 | 3.6 | 148 procs on landed hits, 1 end-game dungeon run, 1 player, main hand while dual wielding         | AHaVwhrgcm9NM2TZ/19                 |

## Count of the 2026-11-14 to 2026-11-15 reports

- Sources: 10 public reports on forever.warcraftlogs.com (zone 3000 listing, 2026-11-14 17:49 to 2026-11-15 07:00 UTC),
  every fight Stoneslayer was worn in: 49zqLvTQayXdg6WN, TF6dc3NCHbKrfxJt, VjBTMHXxFvbthkw2, XpnDCMx9BtbWa1FG,
  a2ZkBR4b6HcFf7dW, c6DWA1CV2tdmRFqL, gV4Xj6c8MdzYwALt, grjxzfTdLQKbnBwZ, v8DRwt2gJBXGxcVN, vDR2nH7fKy6az1rb
- Game build: 1.60.1.70938. The hotfix pushes inside the window (2026-11-14 19:52 to 2026-11-15 06:22 UTC) add 29
  items and change housing and profession spells only, none of the items or proc spells measured here; whether the
  players relogged is not recorded.
- Date: 2026-11-14 and 2026-11-15
- Pull request: #1032

The count is EV-0119 §Method, run on Stoneslayer. The rate is the pooled rate of the table, all its rows, rounded to the
nearest 0.05 step.

| Item        | Class   | Hits   | Procs / isolated hits | Chance | PPM [95%]         | Wearers |
| ----------- | ------- | ------ | --------------------- | ------ | ----------------- | ------- |
| Stoneslayer | Paladin | cast   | 141 / 1115            | 12.6%  | 1.95 [1.66, 2.27] | 8       |
| Stoneslayer | Paladin | white  | 152 / 1325            | 11.5%  | 1.76 [1.52, 2.05] | 8       |
| Stoneslayer | Warrior | cast   | 144 / 982             | 14.7%  | 2.26 [1.94, 2.62] | 1       |
| Stoneslayer | Warrior | white  | 100 / 778             | 12.9%  | 1.98 [1.64, 2.37] | 1       |
| Stoneslayer | all     | pooled | 537 / 4200            | 12.79% | 1.97 [1.82, 2.13] | 9       |
