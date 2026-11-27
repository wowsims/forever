# EV-0222: Venom goes up before its Shank's damage

- Claim: Venom (1310703) goes up before the damage of the Shank that applies it, so the poison that Shank's hit procs
  already deals Venom's 30% more damage.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the streams of the 474 rogues in 521 reports that show
  Venom, too many to list
- Game build: 1.60.1.70938 to 1.60.1.71139 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20
  18:23, 71091 from 2026-11-24 15:21, 71139 from 2026-11-25 19:14); the hotfix each player ran is not recorded, and the
  reports from after the 2026-11-23 hotfix read the same
- Date: 2026-11-13 to 2026-11-26
- Pull request: #1269

Each Shank's damage event is matched with the Venom apply or refresh logged within 50 ms of it. A poison hit is an
Instant Poison damage event with a normal hit (no crit, no partial resist) on the Shank's target, 0 to 50 ms after the
Shank's damage. A Shank is fresh when no Venom was up before it. Each poison hit is taken as a share of the median normal
hit of the same player and poison rank with Venom down.

| Venom's log time against its Shank's damage | Shanks |
| ------------------------------------------- | -----: |
| Before it                                   | 33,743 |
| The same millisecond                        | 36,359 |
| After it                                    |      1 |

| Instant Poison hits    |    Hits | Median share | 0.8 to 1.12 | 1.12 to 1.5 |
| ---------------------- | ------: | -----------: | ----------: | ----------: |
| From a fresh Shank     |   6,589 |         1.24 |         896 |       5,554 |
| Other hits, Venom up   | 128,800 |         1.29 |       5,257 |     118,480 |
| Other hits, Venom down |  69,129 |         0.99 |      64,138 |       3,625 |

The poison a fresh Shank procs reads like a poison under Venom, not like one with Venom down. The reports from after
the 2026-11-23 hotfix (125 reports, 111 rogues) give the same: 1.24 on 2,491 fresh-Shank poison hits, 1.28 under Venom
and 0.98 with Venom down. About one fresh-Shank poison hit in seven reads in the Venom-down band, against one in 24 under
Venom; this count does not explain the difference.

Scripts and output, in `forever-venom-1127`: `venom.py` (`venom_out.txt`, and `venom_out_posthotfix.txt` with
`SINCE=1795482000000`) and `venom_order.py`.
