# EV-0165: Warblade of Caer Darrow's on-hit crits never refresh Vengeance

- Claim: the hits of a spell whose row states Suppress Caster Procs reach none of the caster's procs: Warblade of Caer
  Darrow's 1298499 crits, and the wearer's Vengeance never hears them.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the wearer streams of Retribution paladins wearing
  Warblade of Caer Darrow; the reports and fights below
- Game build: 1.60.1.70938 to 1.60.1.70959 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-16 to 2026-11-20
- Pull request: #1144

## Method

The wearers are every paladin with Warblade of Caer Darrow, Sword of Dismemberment or Blade of Hanna in a hand whose
stream is on disk and holds Vengeance (buff 20050, which the reports name Conviction) applies or refreshes. Only
Warblade of Caer Darrow's wearers have streams with Vengeance, 13 of them in 11 reports.

Each of the wearer's crits (hit type 2 or 17, a report's repeated events counted once, periodic ticks left out) that has
no other crit of the wearer within 150 ms on either side is checked for a Vengeance apply, refresh or stack from 5 ms
before to 60 ms after it. Every other crit the same wearers dealt is counted the same way as the control. A refresh at
full stacks does not always log, so the controls read under 100%; a spell that refreshes Vengeance reads near them.

## Counts

| Crit of                                          | Isolated crits | Vengeance after | Share |
| ------------------------------------------------ | -------------: | --------------: | ----: |
| Warblade of Caer Darrow 1298499                  |          3,045 |               0 |  0.0% |
| White swings                                     |          2,224 |             791 | 35.6% |
| Seal of Righteousness                            |            826 |             346 | 41.9% |
| Holy Strike                                      |            712 |             260 | 36.5% |
| Judgement of Command                             |            515 |             270 | 52.4% |
| Seals of Command                                 |            412 |             220 | 53.4% |
| Windfury Totem hit                               |            288 |             126 | 43.8% |
| Hammer of Wrath                                  |            203 |              91 | 44.8% |
| Frostbolt (a proc with no Suppress Caster Procs) |             20 |              10 | 50.0% |

1298499 states Suppress Caster Procs and Suppress Target Procs (Attributes[3] 0x10000 and 0x20000). Against the
controls' 36 to 53%, about 1,100 of its 3,045 crits would show a Vengeance event if it reached the paladin's procs.

## Reports and fights

CgXG9bvFH6jJVZKk 1-2, 12-13, 15; DfTMrcFAzq9ayQKt 3; JQwbjnVYyGcP4rg8 1, 3-4, 6, 8-9, 11-13, 15-17; PayThRGFLKmbNcnp 7,
9-11; Q3XYxAvj7yDzp4c8 1, 10-11, 117, 123, 156-157, 159-161, 163-167, 169-173, 175-178, 180-184, 186-189, 191, 194-195,
198-200, 202-203, 205-206, 208-209, 211-215, 217; dnDg7j9kRw6WtbML 52, 55, 58, 61, 64, 66-68; fxLMQjdb82cCHvmr 1, 12-13,
15-18, 20-28, 30-31; hZwALnTyCM9fzVrX 1, 3-8, 10-18; khfKmZYaDxgBn24t 1; rQpg6YfW7cLRNAVM 1, 3-5; watv9ZMA2q3jGXgH 1,
3-13, 15-16, 18-24, 26-37.
