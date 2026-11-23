# EV-0120: Flurry Axe's rate, and its procs matched to their hits by log order

- Claim: Flurry Axe 871 grants its extra attack (18797) at 1.75 procs per minute, priced on the axe's speed for each hit
  that rolls it; the rate is observed. Each proc logs in the same millisecond as the hit that rolled it, before that
  hit's damage.
- Sources: public reports on forever.warcraftlogs.com (zone 3000), the fights each wearer had the item on:
    - off hand, Warrior: 4Tc3yBNVYn7wW8RJ/1-31, FtcXATwrHDd9vQ83/3-4, GW1vdX2rKMVPaL78/1-7, VkLX4CZnc68R2Gb1/1-17,
      VtcPT63wQ7WaACvj/9-12, XWfw2PzpNcBhbayY/5-9
    - off hand, Rogue: 7XK1DyrJ2TfAQ48p/2-3, A1yxMTmKatkf3HRJ/1-44, FNyng7tvQ1T2bra8/49-71, RXzv1k3V4PgaWhqj/3-9,
      T79gRkJz4QVGbyYW/6-14, VTfg4qtPDFz6wJvX/1-13, XWfw2PzpNcBhbayY/5-9
    - off hand, Hunter: 4nkW7JDxV12waMfr/38-74, nDNT9zhcQ8trJbwf/38-44, xqnmW3BDpFb28cJ1/1-53
    - main hand, Paladin: Mg2ay1xDjkYpBqRX/1-47, QczFgrTbG4CPpB3n/17-29, VG18w7rW3KRQa9B4/4-10,
      j6mDGcMhb7kFCpgn/1-6, kpy279HLf1VWgPdx/1-6, mRnQ6xjzVDvw2PMA/1-2
- Game build: 1.60.1.70888 and 1.60.1.70938, by report date (6 reports before 2026-11-13, 15 after); hotfix state and
  relogs not recorded. No hotfix changed 18797, the item's on-hit entry or its 1.5 s speed on these builds.
- Date: 2026-11-12 to 2026-11-15
- Pull request: #1039
- Later count: the rate, §Count of the 2026-11-06 to 2026-11-22 reports

A wearer is a player whose combatant info lists the item in the worn fights; runs that two players logged count once.
The extraattacks event of 18797 is logged in the same millisecond as the hit that rolled it, before that hit's damage
event, and the granted swing follows. So each proc goes to the first landed hit logged after it within 30 ms, or
failing that the last landed hit before it within 50 ms: of 2,802 procs, 2,443 match a hit 0 to 9 ms later, 310 one 10
to 30 ms later, 22 one before the proc, and 27 no hit. An isolated-hit count (EV-0119) is biased low for an extra
attack, since the granted swing always sits next to the hit that rolled it.

These reports' wearers with the axe in the off hand have their off-hand white swings told from the main hand's by
damage: a two-group split per run on the logarithm of the damage before armor, which the combat log gives before the
critical strike bonus too. The split matches the weapons' speeds: one warrior run counts 1.87 off-hand swings per
main-hand swing, where a 2.8 s main hand against the 1.5 s axe gives at most 1.87. Main-hand white swings and specials
of these wearers give 3 and 2 procs in about 22,000 and 20,000.

## Other hits

| Hit                                                         | Count                                                                                                                                                  |
| ----------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Off-hand Whirlwind hit of a Warrior wearing it off hand     | 196 / 6,707 per target hit (2.9%)                                                                                                                      |
| Paladin white swing, no Seal of Fury up                     | matched by log order: 51 / 1,447 landed swings (3.5%, 1.41 per minute, interval 1.08 to 1.84)                                                          |
| Paladin natural swing with Seal of Fury up, against without | any proc within 30 ms of a swing with no other swing in the 300 ms before it: 10.9% of 10,257 swings, against 6.3% of 1,165 under another seal or none |
| Holy Strike                                                 | 96 / 2,503 (3.8%)                                                                                                                                      |
| Windfury Totem hit after a paladin wearer's swing           | 1 / 244, all from one paladin with Seal of Fury up, which settles neither way whether the Windfury hit rolls the axe                                   |

A Seal of Fury hit is logged in the white swing's millisecond and its damage rows carry no Not a Proc flag; with the
seal up a swing procs the axe more often, so the seal's hit rolls it too. These counts settle no rule the sim follows.

## Rate

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 18797
- Rate: 1.75 PPM, observed

2,047 procs on 47,060 rolling hits, 24 wearers, 35 runs, in the off hand: all hits 1.74 [1.67, 1.81]. The count finds no
isolated hit with a proc after it, so it measures no proc window and no lone-unit rate, and none checks the all-hits
rate. Observed: no lone-unit rate checks it. Report PThANm6GzytHYDrd logs the same run as L7tdvhcYjgwHTrMz and counts
once. Reports from 2026-11-12 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      16 | 1,351 | 32,111 | 1.68 |
| Warrior |       7 |   608 | 13,126 | 1.85 |
| Hunter  |       1 |    88 |  1,823 | 1.93 |

Reports and fights: 4ChPnBrVK2fJLMaX 1-2, 4; 4azX71FvCjQgmJkc 1-2; 7faQXRhyZr6V32Fk 5-8, 10-13, 15-22, 24-28, 30-35;
7mFKVrW6QLHy3fgk 1; 8qKTwcCyPgd4B3xZ 162-179; A1yxMTmKatkf3HRJ 1-44; ARMrB9JYb2D1TLZC 7-15; AXvy9JkawFZRgdNx 1, 3;
C2dpqPGHK4ryZkRa 1-4; Daj7G8xFX3bA2CmW 1-23; FmdgnGJ3YDPWr8RL 1; FtcXATwrHDd9vQ83 3-4; L7tdvhcYjgwHTrMz 1;
M3azYkjNDRA96rxQ 4-58; PThANm6GzytHYDrd 1; QnJkA2wNM6hYygL1 52 fights from 1 to 61; RwMv31La6FBZ9HDP 1, 3-24;
T9FpP7vKQaL6nNj4 1-6; VtjvpP8bwcgCYTXF 1, 3-18; YwJVCfGj3hnFpRy1 1-2; cAP3RrZDxBW8d67V 1-3; gQnavNZ3Wk4DPJGF 1-2, 4-9,
11-26; gYm2HAtf4bnC8QZT 1-5; hQY9xzt7Mr4nG6Ap 1-2; hWYGMDpr1Bbc6Zzj 1-85; jJWt7Dnm4qc9rzQd 1; m69XVCZpd7Dn8AyH 4-8;
mAW1KFR9YP8r3bNx 1-2; pFH4amZPxkBhyTtV 1-2, 4, 6-34; tqxpQJPZDHkLA2Ka 3-5; vGp9HFWy2ka1nzP6 1-3, 5-6, 8-12;
wtR76PQ8nxpANYkH 1-2; xkN3nXgd8RpHybzT 1-4; xqnmW3BDpFb28cJ1 1-44, 46, 48-50, 52-53; xyGf9QY4dwT8JHMp 1, 3-5;
xzHZV7MFgNhWj9n4 1.

## Count of the 2026-11-06 to 2026-11-22 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights §Rate lists, from every wearer
  stream on disk for the axe; 3,714 reports in all for the count, 36 of them for the axe
- Game build: 1.60.1.70888 to 1.60.1.71037 by report date (70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19
  UTC, 71037 from 2026-11-20 18:23 UTC); 748 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC),
  14 of them for the axe. The sim's database is 1.60.1.71037, push 114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #1180

The count is EV-0183 §Method, run on the axe.
