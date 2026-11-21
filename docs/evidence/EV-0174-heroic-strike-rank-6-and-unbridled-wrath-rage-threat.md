# EV-0174: Heroic Strike Rank 6's flat threat and Unbridled Wrath's rage threat

- Claim: Heroic Strike Rank 6 (11565) adds 100 threat over its damage, times the stance, and Unbridled Wrath's rage
  (12964) makes 5 threat a rage with neither the stance's multiplier nor Defiance's.
- Sources: combat logs, ThreatLog and StatLedger of one protection warrior at levels 42 to 47, Defiance 3 throughout,
  Unbridled Wrath 1 from level 46 and 2 from level 47
- Game build: 1.60.1.70938 on 2026-11-17, 1.60.1.70959 from the 2026-11-18 09:07 login; hotfixes 114140, 114196,
  114206, 114230, 114281 and 114286 each taken at a login once pushed (their push times are known only as upper
  bounds); no reading moves between them
- Date: 2026-11-17 to 2026-11-20
- Pull request: #1157

The method is EV-0043's, run per log, with each login that can carry a new hotfix starting its own stretch. The solo
fit takes 490 fights and 4,725 windows, the party fit 624 fights and 5,394 windows. Seven stretches read every rage
source and heal at a half (one at a third) while damage and flat terms stay whole; six of them open after a Cozy
Sleeping Bag is put down (Unroll Bag 429961) and last until the next login, and the fit splits those sources by the
units counted there.

| Reading                                                      | Measured                                                                                                                                                                                 |
| ------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Heroic Strike Rank 6 (11565), threat over damage             | +100.06, median of n = 50 clean one-mob windows, +100.00 (n = 39) party; 99.94 over every one-mob window (n = 404); 100.00 inside the split stretches, 100.03 outside them               |
| Unbridled Wrath rage, 1 a proc, Defensive Stance, Defiance 3 | 5 ± 0.5 on 8 of 8 one-mob windows whose only rage is this one, 4 of 6 party; 10.8 (5 x 1.8 x 1.2) on none of the 14; 5.04 [4.92, 5.24] a rage (n = 19), party 4.96 [4.66, 5.16] (n = 22) |

Unbridled Wrath was taken on the last day only, and every reading is in Defensive Stance; Battle and Berserker Stance
were not measured. Bloodrage's ticks in the same unsplit stretches read the stance and Defiance's 10.8 a rage, so the 5
is not a split reading.
