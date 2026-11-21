# EV-0147: Weapon armor debuffs against Faerie Fire, Curse of Recklessness and Sunder Armor

- Claim: Armor Shatter (Annihilator) and Puncture Armor (Bashguuder, Rivenspike, Vibroblade) never land while Faerie Fire
  or Curse of Recklessness is on the target, and either one landing removes them in the same millisecond; they stay
  beside Sunder Armor. Corrosive Poison (Strike of the Hydra), Phantom Strike (Phantom Blade) and Wound (Princess
  Theradras' Scepter) stay beside all three. The druid's Faerie Fire removes Dazzling Longsword's Faerie Fire. A
  main-hand wielder in the sim lands about 48 hits a minute beside its swings, each rolling the weapon's proc.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), 178 reports with a wearer of one of the eight weapons;
  the reports and fights each example names
- Game build: 1.60.1.70888 to 1.60.1.70938 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-08 to 2026-11-17
- Pull request: #1092

## Method

Every debuff event on enemies whose spell lowers armor (an A_MOD_ARMOR_NO_MODS effect with a negative amount in the game
files), from every player, sorted by spell id, so a weapon's proc and the class spell of the same name (Dazzling
Longsword's Faerie Fire 13752, the druid's 9907) stay apart. Per enemy (unit and instance), each debuff is up from its
apply to its remove. Three counts:

- a proc debuff removed in the same millisecond another armor debuff is applied or refreshed on that enemy, and the reverse;
- time a proc debuff and another armor debuff are up on one enemy together;
- each wearer's landed, non-periodic melee hits, by which class armor debuff was up on the enemy at the hit, and the
  wearer's proc applications (apply, refresh, stack) by the same state.

The proc rates per hit compare the same wearers' hits under each state; the hits of both hands count, so the rate is
not the weapon's chance on hit. A "never lands" example is a stretch of hits on one enemy under Faerie Fire or Curse of
Recklessness, split at a 10 s gap, with the procs the base rate would give.

Proc landings per landed melee hit of the wearer, by which class armor debuff was already on the target:

| Weapon (debuff)                        | Nothing else         | Sunder Armor only  | Faerie Fire         | Curse of Recklessness | Wearers / reports |
| -------------------------------------- | -------------------- | ------------------ | ------------------- | --------------------- | ----------------- |
| Annihilator (Armor Shatter)            | 72/5,918 (1.22%)     | 0/65 (0.00%)       | 0/549 (0.00%)       | 0/1,577 (0.00%)       | 3 / 2             |
| Rivenspike (Puncture Armor)            | 2500/69,297 (3.61%)  | 307/8,344 (3.68%)  | 0/17,612 (0.00%)    | 0/351 (0.00%)         | 23 / 22           |
| Bashguuder (Puncture Armor)            | 1208/82,349 (1.47%)  | 261/9,260 (2.82%)  | 0/7,169 (0.00%)     | 0/740 (0.00%)         | 11 / 10           |
| Vibroblade (Puncture Armor)            | 141/16,901 (0.83%)   | 22/1,469 (1.50%)   | 0/1,076 (0.00%)     | 0/1,499 (0.00%)       | 10 / 10           |
| Dazzling Longsword (Faerie Fire 13752) | 2044/69,371 (2.95%)  | 401/6,875 (5.83%)  | 180/6,207 (2.90%)   | 229/3,443 (6.65%)     | 24 / 24           |
| Strike of the Hydra (Corrosive Poison) | 2345/100,079 (2.34%) | 443/19,505 (2.27%) | 292/12,014 (2.43%)  | 12/512 (2.34%)        | 60 / 60           |
| Phantom Blade (Phantom Strike)         | 430/37,654 (1.14%)   | 85/4,331 (1.96%)   | 97/8,598 (1.13%)    | –                     | 11 / 11           |
| Princess Theradras' Scepter (Wound)    | 4790/133,707 (3.58%) | 699/12,099 (5.78%) | 1113/31,100 (3.58%) | 7/144 (4.86%)         | 40 / 40           |

## Annihilator (Armor Shatter)

- **Removed when Curse of Recklessness lands** (12×): [aANdrQRjKP7MJnCY fight 1 02:16.866](https://forever.warcraftlogs.com/reports/aANdrQRjKP7MJnCY?fight=1&type=summary&view=events&start=142930&end=147930)
- **Never lands under Faerie Fire or Curse of Recklessness** (base 1.22% a hit): [aANdrQRjKP7MJnCY fight 1 30:18.974](https://forever.warcraftlogs.com/reports/aANdrQRjKP7MJnCY?fight=1&type=summary&view=events&start=1825038&end=1884369) 72 hits under FF, ~0.9 procs expected, 0 landed

## Rivenspike (Puncture Armor)

- **Removed when Faerie Fire lands** (52×): [v8XzFwL9pbHhBCDg fight 1 05:02.559](https://forever.warcraftlogs.com/reports/v8XzFwL9pbHhBCDg?fight=1&type=summary&view=events&start=358280&end=363280), [CqM4Vnh831tFgcYw fight 1 08:24.741](https://forever.warcraftlogs.com/reports/CqM4Vnh831tFgcYw?fight=1&type=summary&view=events&start=562500&end=567500), [T6VDdFhPxQ43pHM2 fight 1 05:55.169](https://forever.warcraftlogs.com/reports/T6VDdFhPxQ43pHM2?fight=1&type=summary&view=events&start=632975&end=637975)
- **Never lands under Faerie Fire or Curse of Recklessness** (base 3.61% a hit): [DWfMg7hzPF4d1mQ6 fight 6 33:23.912](https://forever.warcraftlogs.com/reports/DWfMg7hzPF4d1mQ6?fight=6&type=summary&view=events&start=13009879&end=13088565) 100 hits under FF, ~3.6 procs expected, 0 landed; [NXMZV9xjvWpB7GtP fight 28 27:33.761](https://forever.warcraftlogs.com/reports/NXMZV9xjvWpB7GtP?fight=28&type=summary&view=events&start=30514798&end=30566066) 83 hits under FF, ~3.0 procs expected, 0 landed; [q6379W8YDwVcZ4Rp fight 1 63:19.234](https://forever.warcraftlogs.com/reports/q6379W8YDwVcZ4Rp?fight=1&type=summary&view=events&start=3831442&end=3900288) 82 hits under CoR, ~3.0 procs expected, 0 landed
- **Shares a target with Sunder Armor** (236×): [AyTmpNtxB9jbnJ2H fight 10 14:28.291](https://forever.warcraftlogs.com/reports/AyTmpNtxB9jbnJ2H?fight=10&type=summary&view=events&start=12498278&end=12547354) (44.1 s), [hQY9xzt7Mr4nG6Ap fight 1 26:47.096](https://forever.warcraftlogs.com/reports/hQY9xzt7Mr4nG6Ap?fight=1&type=summary&view=events&start=1857304&end=1896574) (34.3 s), [v8XzFwL9pbHhBCDg fight 1 35:19.170](https://forever.warcraftlogs.com/reports/v8XzFwL9pbHhBCDg?fight=1&type=summary&view=events&start=2174891&end=2211628) (31.7 s)

## Bashguuder (Puncture Armor)

- **Removed when Faerie Fire lands** (15×): [v8Zp2QF7An4d9xMj fight 1 08:25.186](https://forever.warcraftlogs.com/reports/v8Zp2QF7An4d9xMj?fight=1&type=summary&view=events&start=585952&end=590952), [m76CKntQZWLYhTMg fight 1 16:50.877](https://forever.warcraftlogs.com/reports/m76CKntQZWLYhTMg?fight=1&type=summary&view=events&start=1089564&end=1094564), [JRQqNL6DkvWKYdPB fight 1 32:12.701](https://forever.warcraftlogs.com/reports/JRQqNL6DkvWKYdPB?fight=1&type=summary&view=events&start=1931393&end=1936393)
- **Never lands under Faerie Fire or Curse of Recklessness** (base 1.47% a hit): [FHqYCNcKntvDAjxg fight 1 06:06.780](https://forever.warcraftlogs.com/reports/FHqYCNcKntvDAjxg?fight=1&type=summary&view=events&start=364963&end=465163) 90 hits under CoR, ~1.3 procs expected, 0 landed; [v8Zp2QF7An4d9xMj fight 3 50:13.996](https://forever.warcraftlogs.com/reports/v8Zp2QF7An4d9xMj?fight=3&type=summary&view=events&start=5476732&end=5515925) 74 hits under FF, ~1.1 procs expected, 0 landed; [JRQqNL6DkvWKYdPB fight 1 57:28.422](https://forever.warcraftlogs.com/reports/JRQqNL6DkvWKYdPB?fight=1&type=summary&view=events&start=3447114&end=3487407) 56 hits under FF, ~0.8 procs expected, 0 landed
- **Shares a target with Sunder Armor** (184×): [VYBjgJx8aZtRhmpy fight 38 41:33.068](https://forever.warcraftlogs.com/reports/VYBjgJx8aZtRhmpy?fight=38&type=summary&view=events&start=84311170&end=84374900) (58.7 s), [FHqYCNcKntvDAjxg fight 34 47:21.384](https://forever.warcraftlogs.com/reports/FHqYCNcKntvDAjxg?fight=34&type=summary&view=events&start=355610828&end=355655844) (40.0 s), [6xnwM2qCN43Tv8Af fight 1 29:01.463](https://forever.warcraftlogs.com/reports/6xnwM2qCN43Tv8Af?fight=1&type=summary&view=events&start=1858364&end=1898090) (34.7 s)

## Vibroblade (Puncture Armor)

- **Removed when Curse of Recklessness lands** (3×): [MKx2v8JfHVc6kmND fight 1 07:03.445](https://forever.warcraftlogs.com/reports/MKx2v8JfHVc6kmND?fight=1&type=summary&view=events&start=498395&end=503395)
- **Never lands under Faerie Fire or Curse of Recklessness** (base 0.83% a hit): [4M8gQTJhZvRPWf2B fight 1 24:10.581](https://forever.warcraftlogs.com/reports/4M8gQTJhZvRPWf2B?fight=1&type=summary&view=events&start=1603243&end=1675126) 100 hits under FF, ~0.8 procs expected, 0 landed; [BtN6VMJRjgp1Ay9G fight 1 06:07.785](https://forever.warcraftlogs.com/reports/BtN6VMJRjgp1Ay9G?fight=1&type=summary&view=events&start=561540&end=608304) 92 hits under CoR, ~0.8 procs expected, 0 landed; [MKx2v8JfHVc6kmND fight 1 16:13.272](https://forever.warcraftlogs.com/reports/MKx2v8JfHVc6kmND?fight=1&type=summary&view=events&start=1048222&end=1096027) 53 hits under CoR, ~0.4 procs expected, 0 landed
- **Shares a target with Sunder Armor** (27×): [HpCXWPnytdkvzqg2 fight 1 46:37.884](https://forever.warcraftlogs.com/reports/HpCXWPnytdkvzqg2?fight=1&type=summary&view=events&start=2819284&end=2855834) (31.6 s), [fA1GLbc6gw9nKHM2 fight 1 06:56.039](https://forever.warcraftlogs.com/reports/fA1GLbc6gw9nKHM2?fight=1&type=summary&view=events&start=646150&end=670678) (19.5 s), [BtN6VMJRjgp1Ay9G fight 1 00:05.227](https://forever.warcraftlogs.com/reports/BtN6VMJRjgp1Ay9G?fight=1&type=summary&view=events&start=198982&end=210407) (6.4 s)

## Dazzling Longsword (Faerie Fire 13752)

- **Removed when Faerie Fire lands** (115×): [z639ZrPafCA1m7WT fight 1 00:28.503](https://forever.warcraftlogs.com/reports/z639ZrPafCA1m7WT?fight=1&type=summary&view=events&start=79714&end=84714), [gTPNZcJ9mG4dRC2X fight 1 01:16.678](https://forever.warcraftlogs.com/reports/gTPNZcJ9mG4dRC2X?fight=1&type=summary&view=events&start=409190&end=414190), [1FV3WyfXGCcPrT9b fight 1 13:27.111](https://forever.warcraftlogs.com/reports/1FV3WyfXGCcPrT9b?fight=1&type=summary&view=events&start=815706&end=820706)
- **Removed when Sunder Armor lands** (1×): [KVdX1Wn4GpzAH9MN fight 11 15:30.145](https://forever.warcraftlogs.com/reports/KVdX1Wn4GpzAH9MN?fight=11&type=summary&view=events&start=9537380&end=9542380)
- **Shares a target with Sunder Armor** (387×): [8X2Kc46faTnxVBLm fight 1 25:58.902](https://forever.warcraftlogs.com/reports/8X2Kc46faTnxVBLm?fight=1&type=summary&view=events&start=1626859&end=1675795) (43.9 s), [pKj3bDrf1h2qtNMF fight 13 31:58.146](https://forever.warcraftlogs.com/reports/pKj3bDrf1h2qtNMF?fight=13&type=summary&view=events&start=8936340&end=8984509) (43.2 s), [KVdX1Wn4GpzAH9MN fight 11 10:19.085](https://forever.warcraftlogs.com/reports/KVdX1Wn4GpzAH9MN?fight=11&type=summary&view=events&start=9226320&end=9274001) (42.7 s)
- **Shares a target with Faerie Fire** (116×): [tkQKL9bGT6JVR3na fight 14 25:11.989](https://forever.warcraftlogs.com/reports/tkQKL9bGT6JVR3na?fight=14&type=summary&view=events&start=8314880&end=8347832) (28.0 s), [pKj3bDrf1h2qtNMF fight 13 19:16.581](https://forever.warcraftlogs.com/reports/pKj3bDrf1h2qtNMF?fight=13&type=summary&view=events&start=8174775&end=8207449) (27.7 s), [QDadybK3XVPRvwLh fight 21 26:27.208](https://forever.warcraftlogs.com/reports/QDadybK3XVPRvwLh?fight=21&type=summary&view=events&start=12432648&end=12464973) (27.3 s)
- **Shares a target with Curse of Recklessness** (123×): [BCZgcRxrh7dA4ayX fight 1 77:42.306](https://forever.warcraftlogs.com/reports/BCZgcRxrh7dA4ayX?fight=1&type=summary&view=events&start=4695770&end=4735325) (34.6 s), [QDadybK3XVPRvwLh fight 27 24:32.160](https://forever.warcraftlogs.com/reports/QDadybK3XVPRvwLh?fight=27&type=summary&view=events&start=15007943&end=15040686) (27.7 s), [nLk2TKDJWP6vBgam fight 1 27:33.073](https://forever.warcraftlogs.com/reports/nLk2TKDJWP6vBgam?fight=1&type=summary&view=events&start=1768016&end=1800130) (27.1 s)

## Strike of the Hydra (Corrosive Poison)

- **Shares a target with Sunder Armor** (397×): [mFMLvknrhNyaARtB fight 1 41:36.919](https://forever.warcraftlogs.com/reports/mFMLvknrhNyaARtB?fight=1&type=summary&view=events&start=2538425&end=2580431) (37.0 s), [rRLkva7gz9ymhb4V fight 1 48:40.846](https://forever.warcraftlogs.com/reports/rRLkva7gz9ymhb4V?fight=1&type=summary&view=events&start=3161648&end=3199370) (32.7 s), [LcjBAPhTgdnRCHQG fight 1 45:23.521](https://forever.warcraftlogs.com/reports/LcjBAPhTgdnRCHQG?fight=1&type=summary&view=events&start=2847688&end=2883292) (30.6 s)
- **Shares a target with Faerie Fire** (251×): [LHRhtADNW8GfyYqT fight 1 27:12.943](https://forever.warcraftlogs.com/reports/LHRhtADNW8GfyYqT?fight=1&type=summary&view=events&start=4348090&end=4378679) (25.6 s), [j1CGayX43bMdF8hP fight 1 09:14.251](https://forever.warcraftlogs.com/reports/j1CGayX43bMdF8hP?fight=1&type=summary&view=events&start=588637&end=615003) (21.4 s), [BKnf4rNwJdcqpQ18 fight 1 18:31.733](https://forever.warcraftlogs.com/reports/BKnf4rNwJdcqpQ18?fight=1&type=summary&view=events&start=1456284&end=1479927) (18.6 s)
- **Shares a target with Curse of Recklessness** (13×): [9zabj3grQkGB2FKL fight 8 22:41.546](https://forever.warcraftlogs.com/reports/9zabj3grQkGB2FKL?fight=8&type=summary&view=events&start=7181960&end=7207895) (20.9 s), [HtQr7kYaNq9MGgvZ fight 1 22:23.982](https://forever.warcraftlogs.com/reports/HtQr7kYaNq9MGgvZ?fight=1&type=summary&view=events&start=1373177&end=1390310) (12.1 s), [6X1L4nB7zCfqk8vV fight 1 40:00.060](https://forever.warcraftlogs.com/reports/6X1L4nB7zCfqk8vV?fight=1&type=summary&view=events&start=2753978&end=2768084) (9.1 s)

## Phantom Blade (Phantom Strike)

- **Shares a target with Sunder Armor** (79×): [pLtJjx8DCPbHRacv fight 11 01:52.080](https://forever.warcraftlogs.com/reports/pLtJjx8DCPbHRacv?fight=11&type=summary&view=events&start=5239507&end=5286453) (41.9 s), [6kN78G2xfqy3vwTP fight 4 00:29.916](https://forever.warcraftlogs.com/reports/6kN78G2xfqy3vwTP?fight=4&type=summary&view=events&start=2449149&end=2493522) (39.4 s), [D2d3hNAbwkPgL1vF fight 10 17:24.847](https://forever.warcraftlogs.com/reports/D2d3hNAbwkPgL1vF?fight=10&type=summary&view=events&start=4323596&end=4348606) (20.0 s)
- **Shares a target with Faerie Fire** (92×): [J23GjB7aNAKgrLqM fight 1 14:11.614](https://forever.warcraftlogs.com/reports/J23GjB7aNAKgrLqM?fight=1&type=summary&view=events&start=1052027&end=1091840) (34.8 s), [hrV3QgyzvfdZKJwF fight 24 00:05.832](https://forever.warcraftlogs.com/reports/hrV3QgyzvfdZKJwF?fight=24&type=summary&view=events&start=7300887&end=7326515) (20.6 s), [QAHXb9zy4jfgKPVJ fight 1 15:36.111](https://forever.warcraftlogs.com/reports/QAHXb9zy4jfgKPVJ?fight=1&type=summary&view=events&start=1677681&end=1702675) (20.0 s)

## Princess Theradras' Scepter (Wound)

- **Shares a target with Sunder Armor** (713×): [qprynNbXtRZV4DHP fight 1 69:44.530](https://forever.warcraftlogs.com/reports/qprynNbXtRZV4DHP?fight=1&type=summary&view=events&start=4462371&end=4527420) (60.0 s), [cgMGCdPJQ84wzaBV fight 12 26:52.212](https://forever.warcraftlogs.com/reports/cgMGCdPJQ84wzaBV?fight=12&type=summary&view=events&start=15005262&end=15061761) (51.5 s), [kr3AzhZ9FVB1nG2D fight 4 14:39.179](https://forever.warcraftlogs.com/reports/kr3AzhZ9FVB1nG2D?fight=4&type=summary&view=events&start=2924400&end=2980432) (51.0 s)
- **Shares a target with Faerie Fire** (784×): [Hm2Yx9zLQVJqDW6v fight 13 25:05.932](https://forever.warcraftlogs.com/reports/Hm2Yx9zLQVJqDW6v?fight=13&type=summary&view=events&start=7361489&end=7428569) (62.1 s), [cgMGCdPJQ84wzaBV fight 12 26:50.462](https://forever.warcraftlogs.com/reports/cgMGCdPJQ84wzaBV?fight=12&type=summary&view=events&start=15003512&end=15061761) (53.2 s), [mkywZBCNJV2QbPjd fight 1 33:00.539](https://forever.warcraftlogs.com/reports/mkywZBCNJV2QbPjd?fight=1&type=summary&view=events&start=2597654&end=2652211) (49.6 s)
- **Shares a target with Curse of Recklessness** (10×): [9gJ1KcWkwn7dQpmP fight 1 13:01.767](https://forever.warcraftlogs.com/reports/9gJ1KcWkwn7dQpmP?fight=1&type=summary&view=events&start=824663&end=850638) (21.0 s)

## A main-hand wielder's landed hits in the sim

The raid's copy stands for one main-hand wielder. Measured with `wowsimcli` on #1092's branch, where the player's
Puncture Armor landed with no hit roll of its own (180 s, one level 63
target from behind, 5,000 iterations, the raid's armor debuffs off but Sunder Armor): the pre-BiS Fury warrior with each
weapon in the main hand and the Combat rogue with Rivenspike. The hits a minute are the debuff's procs a minute over the
chance a hit, PPM × speed / 60, and match the summed landed hits of the warrior's Annihilator and the rogue's Rivenspike
within 4%.

| Wielder                    | Procs a minute | Landed hits a minute | Swings a minute at base speed | Other hits a minute |
| -------------------------- | -------------: | -------------------: | ----------------------------: | ------------------: |
| Warrior, Annihilator (1.0) |           2.94 |                 73.5 |                          25.0 |                48.5 |
| Warrior, Bashguuder (3.2)  |           8.00 |                 83.4 |                          33.3 |                50.1 |
| Warrior, Rivenspike (1.6)  |           5.20 |                 67.3 |                          20.7 |                46.6 |
| Rogue, Rivenspike (1.6)    |           5.20 |                 67.2 |                          20.7 |                46.5 |

The other hits (specials, Windfury strikes, Whirlwind, extra attacks) come to about 48 a minute whatever the weapon's
speed, so the raid's copy rolls on its swings and 48 more hits a minute. The owner chose the sim's wielder as the model.

The logs land fewer procs a hit than PPM × speed / 60. Over stretches with no Faerie Fire, Curse of Recklessness,
Expose Armor or another player's copy of the debuff on the target, counting only the wearer's own landed hits:

| Wearer                | Procs expected | Logged | Logged / expected |
| --------------------- | -------------: | -----: | ----------------: |
| Rivenspike, main hand |            314 |    236 |              0.75 |
| Bashguuder, main hand |            422 |    140 |              0.33 |
| Bashguuder, off hand  |            196 |    140 |              0.71 |
