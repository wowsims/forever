# EV-0216: Whirlwind's off-hand strike brings off-hand Sweeping Strikes strikes

- Claim: With Sweeping Strikes up, a dual-wield Whirlwind brings off-hand strikes (462890) beside the main hand's one
  normalized strike (26654), each spending a charge. One 26654 plus one 462890 for each off-hand strike, landed or not,
  capped by the charges up, fits most casts; how many a cast brings is not settled.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the streams of 6 dual-wield warriors in the 6 reports
  under Reports, all their fights
- Game build: 1.60.1.70938 for 48 of the casts, 1.60.1.70959 for 4 and 1.60.1.71037 for 7, by report date; the
  hotfix each player ran is not recorded
- Date: 2026-11-15 to 2026-11-22
- Pull request: not yet opened

## Method

Each Whirlwind cast (1680) with the warrior's own events in the 2 ms on either side: the main-hand strikes (1680), the
off-hand strikes (462891), the Sweeping Strikes strikes 26654 and 462890, and the charges of Sweeping Strikes (12292)
before the cast and spent in it, from the buff's stack events. The casts counted have charges up and both a main-hand
and an off-hand strike: 59 casts. Each model predicts the strikes before the charges cap them, and the cap is the
charges up before the cast; 11 casts end with charges left, so the cap does not decide them.

## Numbers

| Model                                                                             | Casts it fits, of 59 | Of the 11 with charges left |
| --------------------------------------------------------------------------------- | -------------------- | --------------------------- |
| One 26654, plus one 462890 per off-hand strike, landed or not                     | 55                   | 7                           |
| One 26654, plus one 462890 per landed off-hand strike                             | 52                   | 6                           |
| One 26654 if a main-hand strike lands, plus one 462890 per landed off-hand strike | 51                   | 5                           |
| One strike for each hand that lands a strike                                      | 25                   | 2                           |
| One 26654 if a main-hand strike lands, and no 462890 (the two-handed reading)     | 9                    | 1                           |

The 59 casts bring 57 of 26654 and 93 of 462890 against 159 landed off-hand strikes, and in every one of them the
charges spent equal the strikes brought. The four casts the first model misses all end with charges left and bring one
or two strikes fewer than it predicts. By report build it fits 44 of 48 casts on 70938, 4 of 4 on 70959 and 7 of 7 on 71037. None of the 2,325 Whirlwinds cast with no charges up by the 21 warrior streams
EV-0213 counts brings either strike, and no cast without an off-hand strike brings a 462890.

462890 deals off-hand damage. In the fights whose gear lists a weapon in each hand, its 96 strikes average 69.8 against
189.4 for 77 of 26654 and 125.9 for 350 copied hits (12723). Its 93 strikes above log 76 hits and 17 crits, no
avoidance. Over those fights, of 5 of the 6 warriors (one has no fight with two weapons listed), 462890 deals 0.75% of
their damage, 26654 1.64% and 12723 4.96%.

## Reports

782JNDc1pgGRZxVC, 8Zp12yQdxz67MaDL, gDtFwpGPLM2TyZKB, mPMpkAgjbtR7HBXD, y6FZRHQnckD8YjAC, ZFXTPc8fz3RHjKM9.
