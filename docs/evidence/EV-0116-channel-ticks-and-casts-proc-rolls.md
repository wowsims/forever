# EV-0116: What channel ticks, waves and casts roll: Arcane Concentration and Judgement of Wisdom

- Claim: Blizzard ticks never roll Arcane Concentration, and Hellfire Effect and Rain of Fire waves never return
  Judgement of Wisdom mana. A Blizzard cast rolls Arcane Concentration once however many enemies stand in the area, a
  Rain of Fire cast returns Judgement of Wisdom mana once for each debuffed enemy in its area, and a Hellfire cast
  returns none.
- Sources: Forever dungeon reports on Warcraft Logs, 40 with Blizzard and 45 with Hellfire or Rain of Fire, listed
  below; fights where the mage or warlock cast the spell
- Game build: 1.60.1.70862, 1.60.1.70888 and 1.60.1.70938, by report date (counts per section below); hotfix state not
  recorded. No hotfix changed the Blizzard tick rows, Arcane Concentration 11213, Clearcasting 12536, Hellfire Effect
  or the Judgement of Wisdom rows on these builds.
- Date: 2026-11-07 to 2026-11-14
- Pull request: #1021

Reports come from a search of 3,226 zone 3000 reports for the spells' ids. A hit counts when it lands; a gain or a
payment is tied to a hit or a cast by its timestamp, with the windows below.

## Arcane Concentration and Blizzard

Reports (40; builds 70862: 1, 70888: 2, 70938: 37): 176QTFWR8PtdfGZB, 17AwLhaHjGqWtgrk, 19KWay63j2HvGdgQ,
1BNXynxVbH4kGt8R, 1YJdRn73FrtMwvxp, 1ZJVMgcYWTyRQh2w, 1ZyhaVDdFtTYWMcC, 1hRZ2v8PwKWt7DkX, 1pDcTvkrCBhZJRgf,
21YWbyCHgRZ4xKFz, 2NtYqxGZK1nDTCBh, 2RdqgkHFNVMnKCh8, 2xhVyXGW7gaDjA4J, 31XVLHBdfaFMn7qT, 3ZkcGqzWn28BQD4r,
3cWGP7L8k9tMDJAQ, 3g24PkCNcApvHnh7, 3grN8Rct9Xzknm1H, 3tvhxT7fRQNdL1wK, 436rTjRbcgA27Z91, 49kgZ36jfGADhRwc,
4BFrCYZajATmLfgn, 4LrMDtG6T8AVRJzQ, 4xD1ztgPHnjbfAJ3, 4xQYfaGrkwBAnzWV, 4yR3BPdmJQkr2FCH, 64aPK71NMR9LFB3Y,
6KnwtMdF4gkxGy1N, 6NW8rCvj4bctK7PZ, 6TcNHMBR8afV3jq7, 6Yb2VmchGXjZKv9w, 6Zzwx92Wh4qFVLBp, 6jqwbMpRFxTmHKkX,
72rtnzHfdYM4DLv8, 78Bfj6rPXNndxQDg, 7XK1DyrJ2TfAQ48p, 7ZxcPNTjqvfzQHgX, 7cpTkjaFB3mz6VgM, 7kxNnWGw2FzfLJtR,
7qgPz34dAnxGT9MN. 48 mages gained Clearcasting; 35 of them cast Blizzard.

A gain counts for a tick when the only own hit in the 300 ms before it is a tick; hits within 1 s of an earlier gain
are left out. The same mages' direct spells gain on 1,209 of 14,856 hits (8.1%): Arcane Explosion 7.8%, Arcane Blast
8.9%, Frostbolt 8.1%, Flamestrike 7.1%, Fire Blast 7.2%, Cone of Cold 9.3%.

| What lands     | Rolls outside the cooldown              | Gains | If it rolled like a direct hit |
| -------------- | --------------------------------------- | ----: | -----------------------------: |
| Blizzard ticks | 5,258 tick batches (16,865 target hits) |     2 |    383 a batch, 1,250 a target |
| Blizzard casts | 769                                     |   106 |                             65 |

The 2 tick gains have no other own event near them; 13 more share their window with a tick and a mage spell, and
counting all 15 still gives 15 against 383. Windows of 150, 300 and 500 ms give the same 2.

A Blizzard cast rolls once, not once for each enemy: by the enemies its first tick hits, 1 gives 21 of 183 (11.5%),
2 gives 39 of 266, 3 gives 33 of 238, 4 gives 9 of 61, where a roll for each enemy at 8.1% gives 8, 16, 22 and 29%.
Gains come in two windows, 0 to 2 ms after the cast (72 casts, 9.4%) and 5 to 40 ms after it (41, 5.3%). 7 casts
gained twice, 13 to 19 ms apart, and 1 of 14 casts within 1 s of an earlier gain still gained. One roll at each mage's
own direct rate predicts 65.1 gains and two rolls 123.9, against 106.

## Judgement of Wisdom, Hellfire and Rain of Fire

Reports (45; builds 70862: 7, 70888: 3, 70938: 35): 1A6dkC8FLqvxrfQz, 1KF2jNqTzc6VJ9fp, 1NBKq2mDWghwR69z,
1QLd8NAZqKhDGRzy, 1aBhkzdDMZKGtb37, 1b7PdYR4HhDXBGT6, 1kKGT9BfgJjbFdNH, 1nTQBK4hmcaLbFZV, 27jZGhfHgp1LR8MB,
2HMyR6kfbWKzBhtv, 2bdPF6mYxvJgZXK7, 2zjykrAbfwG1LZBM, 34QpBng2Hw1t97NP, 3B4wmh7JGkQq2Ndj, 3T8GrWJRdBkvy4gz,
3bqWjVMPpDfdCgYz, 3grN8Rct9Xzknm1H, 3yjCr7kWwFnVRXTd, 3ytgzKdVNcxfLvAC, 49qnrcjm3ZGAPYgW, 4BAyv1YW2hdpq3fc,
6B8DyLqvZR1VFX2W, 6WgpmHQrv9kn1ZXJ, 6cB7yxbjPAYaRn9W, 6f2mtzphbVCgKYJ3, 7cpTkjaFB3mz6VgM, 7pALnmqJNKbk4cZ8,
7tLY1pVK9wRWPABX, 813rPaLtcdpHy67q, 8LaNq1JXfZzy3Dj9, 8NMWGB1n6zdLtPrY, 8YLV79MjFGtR2QWc, 8f1N9WC6ky3jZcaK,
8gXL7ZMhQPmynT9w, 9YgjFwnydrJN4b1Z, 9zBrCg3GdAQ14va2, 9zYTDgb68c24vJNm, 9zxHaWbYT3gqDQvF, ACJ4XDNLKgkZdtjc,
AH19FhrR27gZdzmV, AHaVwhrgcm9NM2TZ, AQdfnNFX82TtVagp, AZc2TaQNgb8VGX4C, Avwb6ZLJHQKyfhM9, B7hYPwdmMb49fa2A.

The mana logs as a resource change with the warlock as source and target, so the log does not name the enemy that
paid. A hit counts on a target with a Judgement of Wisdom debuff up; a clean hit has no other own hit or cast from
30 ms before it to 30 ms after.

| Clean landed hits                                                          |   Hits |          Mana within 30 ms |
| -------------------------------------------------------------------------- | -----: | -------------------------: |
| Direct spells (Shadow Bolt, Searing Pain, Incinerate, Immolate, Soul Fire) |    706 | 315 (45%; 49.5% at 300 ms) |
| Wand                                                                       |    224 |                        47% |
| Dot ticks                                                                  | 10,817 |                       0.4% |
| Hellfire Effect                                                            |    502 |                          0 |
| Rain of Fire waves                                                         |    588 |                          0 |

Rolling at the direct rate, Hellfire Effect would give 173 to 226 payments and the waves 214 to 265. No mana event has
a channel hit as its only candidate, and none of the 1,406 landed channel hits, clean or not, is followed by mana
within 30 ms.

The Rain of Fire cast pays 0 to 22 ms after the cast line (16 of 110 at 0 to 1 ms), once for each debuffed enemy its
first wave hits:

| Debuffed enemies in the first wave | Casts paid | Payments |
| ---------------------------------- | ---------: | -------: |
| 0                                  |   4 of 118 |        4 |
| 1                                  |  66 of 139 |       67 |
| 2                                  |   21 of 29 |       31 |
| 3                                  |     7 of 8 |       16 |

With no Judgement of Wisdom debuff up it pays 0 of 1,275 times. Hellfire casts pay 0 of 121 with a debuff up and 0
of 756 without; every effect of the Hellfire cast row targets the caster. Touch of the Grave, as a control, drains on
25.5% and 21.0% of the same Undead casters' direct hits and on none of their casts (Blizzard 0 of 324, Rain of Fire
0 of 171, Hellfire 0 of 100), so the casts land as hits that deal no damage.

The Blizzard reports hold no Judgement of Wisdom mana for mages, so the Blizzard cast's payment per enemy follows the
row, which matches Rain of Fire's.
