# EV-0120: Flurry Axe's rate from off-hand white swings, matched by log order

- Claim: Flurry Axe 871 grants its extra attack (18797) at 1.65 procs per minute.
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

A wearer is a player whose combatant info lists the item in the worn fights; runs that two players logged count once.
The extraattacks event of 18797 is logged in the same millisecond as the hit that rolled it, before that hit's damage
event, and the granted swing follows. So each proc goes to the first landed hit logged after it within 30 ms, or
failing that the last landed hit before it within 50 ms: of 2,802 procs, 2,443 match a hit 0 to 9 ms later, 310 one 10
to 30 ms later, 22 one before the proc, and 27 no hit. An isolated-hit count (EV-0119) is biased low for an extra
attack, since the granted swing always sits next to the hit that rolled it.

The rate comes from the wearers with the axe in the off hand, whose off-hand white swings are told from the main hand's
by damage: a two-group split per run on the logarithm of the damage before armor, which the combat log gives before the
critical strike bonus too. The split matches the weapons' speeds: one warrior run counts 1.87 off-hand swings per
main-hand swing, where a 2.8 s main hand against the 1.5 s axe gives at most 1.87. Main-hand white swings and specials
of these wearers give 3 and 2 procs in about 22,000 and 20,000. PPM = chance × 60 / 1.5 s.

| Wearer class | Procs / landed off-hand white swings | Procs per minute |
| ------------ | -----------------------------------: | ---------------: |
| Warrior      |                          361 / 9,061 |             1.59 |
| Warrior      |                           65 / 1,571 |             1.65 |
| Rogue        |                         632 / 14,595 |             1.73 |
| Rogue        |                             30 / 751 |             1.60 |
| Hunter       |                           99 / 2,440 |             1.62 |
| All          |               1,187 / 28,418 (4.18%) |             1.67 |

The pooled rate, 1.67 (count interval 1.58 to 1.77, wearer bootstrap 1.60 to 1.72), rounds to 1.65 at the nearest
0.05 step.

## Other hits

| Hit                                                         | Count                                                                                                                                                  |
| ----------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Off-hand Whirlwind hit of a Warrior wearing it off hand     | 196 / 6,707 per target hit (2.9%)                                                                                                                      |
| Paladin white swing, no Seal of Fury up                     | matched by log order: 51 / 1,447 landed swings (3.5%, 1.41 per minute, interval 1.08 to 1.84)                                                          |
| Paladin natural swing with Seal of Fury up, against without | any proc within 30 ms of a swing with no other swing in the 300 ms before it: 10.9% of 10,257 swings, against 6.3% of 1,165 under another seal or none |
| Holy Strike                                                 | 96 / 2,503 (3.8%)                                                                                                                                      |
| Windfury Totem strike after a paladin wearer's swing        | 1 / 244, all from one paladin with Seal of Fury up, which settles neither way whether the strike rolls the axe                                         |

A Seal of Fury hit is logged in the white swing's millisecond and its damage rows carry no Not a Proc flag; with the
seal up a swing procs the axe more often, so the seal's hit rolls it too. These counts settle no rule the sim follows.
