# EV-0116: What channel ticks, waves and casts roll: Arcane Concentration and Judgement of Wisdom

- Claim: Blizzard ticks never roll Arcane Concentration, and Hellfire Effect and Rain of Fire waves never return
  Judgement of Wisdom mana. A Blizzard cast gains Arcane Concentration in two windows, at the cast at the talent's
  chance and 5 to 40 ms later at about 6%, neither once for each enemy in the area. A Rain of Fire cast returns
  Judgement of Wisdom mana once for each debuffed enemy in its area, and a Hellfire cast returns none.
- Sources: Forever dungeon reports on Warcraft Logs, 45 with Hellfire or Rain of Fire, listed below; fights where the
  warlock cast the spell
- Game build: 1.60.1.70862, 1.60.1.70888 and 1.60.1.70938, by report date (counts per section below); hotfix state not
  recorded. No hotfix changed Hellfire Effect or the Judgement of Wisdom rows on these builds.
- Date: 2026-11-07 to 2026-11-14
- Pull request: #1021
- Later count: Arcane Concentration on Blizzard ticks and casts, §Count of the 2026-11-07 to 2026-11-26 logs

Reports come from a search of 3,226 zone 3000 reports for the spells' ids. A hit counts when it lands; a gain or a
payment is tied to a hit or a cast by its timestamp, with the windows below.

## Arcane Concentration and Blizzard

- Count: §Count of the 2026-11-07 to 2026-11-26 logs

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

## Count of the 2026-11-07 to 2026-11-26 logs

- Sources: 40 Forever dungeon reports on Warcraft Logs, listed below, fights where the mage cast Blizzard; a combat log
  of a level 60 Frost mage with Arcane Concentration 5 in a group
- Game build: the reports 1.60.1.70862, 1.60.1.70888 and 1.60.1.70938, by report date, hotfix state not recorded; no
  hotfix changed the Blizzard tick rows, Arcane Concentration 11213 or Clearcasting 12536 on those builds. The combat log
  1.60.1.71091 until 2026-11-25 19:11 UTC and 1.60.1.71139 from 20:16 UTC, after hotfix push 114399, relogged
- Date: the reports 2026-11-07 to 2026-11-14; the combat log 2026-11-25 14:33 to 2026-11-26 03:51 UTC
- Pull request: #1021 (the reports); #1280 (the combat log)

Reports (40; builds 70862: 1, 70888: 2, 70938: 37): 176QTFWR8PtdfGZB, 17AwLhaHjGqWtgrk, 19KWay63j2HvGdgQ,
1BNXynxVbH4kGt8R, 1YJdRn73FrtMwvxp, 1ZJVMgcYWTyRQh2w, 1ZyhaVDdFtTYWMcC, 1hRZ2v8PwKWt7DkX, 1pDcTvkrCBhZJRgf,
21YWbyCHgRZ4xKFz, 2NtYqxGZK1nDTCBh, 2RdqgkHFNVMnKCh8, 2xhVyXGW7gaDjA4J, 31XVLHBdfaFMn7qT, 3ZkcGqzWn28BQD4r,
3cWGP7L8k9tMDJAQ, 3g24PkCNcApvHnh7, 3grN8Rct9Xzknm1H, 3tvhxT7fRQNdL1wK, 436rTjRbcgA27Z91, 49kgZ36jfGADhRwc,
4BFrCYZajATmLfgn, 4LrMDtG6T8AVRJzQ, 4xD1ztgPHnjbfAJ3, 4xQYfaGrkwBAnzWV, 4yR3BPdmJQkr2FCH, 64aPK71NMR9LFB3Y,
6KnwtMdF4gkxGy1N, 6NW8rCvj4bctK7PZ, 6TcNHMBR8afV3jq7, 6Yb2VmchGXjZKv9w, 6Zzwx92Wh4qFVLBp, 6jqwbMpRFxTmHKkX,
72rtnzHfdYM4DLv8, 78Bfj6rPXNndxQDg, 7XK1DyrJ2TfAQ48p, 7ZxcPNTjqvfzQHgX, 7cpTkjaFB3mz6VgM, 7kxNnWGw2FzfLJtR,
7qgPz34dAnxGT9MN. 48 mages gained Clearcasting; 35 of them cast Blizzard.

In the reports a gain counts for a tick when the only own hit in the 300 ms before it is a tick; hits within 1 s of
an earlier gain are left out. The same mages' direct spells gain on 1,209 of 14,856 hits (8.1%): Arcane Explosion
7.8%, Arcane Blast 8.9%, Frostbolt 8.1%, Flamestrike 7.1%, Fire Blast 7.2%, Cone of Cold 9.3%. In the combat log a
gain is an apply or refresh of Clearcasting 12536 on the mage, merged per millisecond; a cast or a tick batch more
than 0.2 s from a cast gains when one falls within 60 ms after it, and no Blizzard cast had a gain in the 1 s before
it. The mage's other direct spells, each landing within 50 ms counted once, gain on 102 of 1,177 with the cooldown
ready (8.7%): Frostbolt 62 of 675, Ice Lance 16 of 226, Arcane Explosion 14 of 141, Cone of Cold 6 of 65.

| What lands                                 | Reports                         | Combat log        | Both                 |
| ------------------------------------------ | ------------------------------- | ----------------- | -------------------- |
| Blizzard tick batches, gains               | 2 of 5,258 (16,865 target hits) | 0 of 2,962        | 2 of 8,220           |
| Blizzard casts that gained                 | 106 of 769 (13.8%)              | 68 of 437 (15.6%) | 174 of 1,206 (14.4%) |
| Gained 0 to 2 ms after the cast            | 72 (9.4%)                       | 44 (10.1%)        | 116 (9.6%)           |
| Gained 5 to 40 ms after it                 | 41 (5.3%)                       | 30 (6.9%)         | 71 (5.9%)            |
| Gained in both windows                     | 7                               | 6                 | 13                   |
| Casts gaining, one roll at the direct rate | 65.1                            | 37.9              | 103.0                |
| Casts gaining, two rolls at it             | 123.9                           | 72.5              | 196.4                |

Both together, 14.4% [12.6, 16.5] of casts gain, 9.6% [8.1, 11.4] at the cast and 5.9% [4.7, 7.4] in the later
window. If the tick batches rolled like a direct hit they would gain 383 times in the reports' batches (1,250 for each
target). The 2 tick gains have no other own event near them; 13 more share their window with a tick and a mage spell,
and counting all 15 still gives 15 against 383. Windows of 150, 300 and 500 ms give the same 2. In the reports no gain
fell 3 or 4 ms after a cast and 1 of 14 casts within 1 s of an earlier gain still gained; the two gains of a cast came
13 to 19 ms apart. In the combat log every gain fell in one of the two windows, and the later window is where
Draconic Infused Emblem 22268's gains from the same casts fall.

Neither window rolls once for each enemy. By the enemies the first tick hits, casts gain in the reports 21 of 183 at 1
enemy (11.5%), 39 of 266 at 2, 33 of 238 at 3 and 9 of 61 at 4, where a roll for each enemy at 8.1% gives 8, 16, 22
and 29%; in the combat log 12 of 71 at 3 (16.9%), 10 of 88 at 4 (11.4%) and 40 of 237 at 5 or more (16.9%), where a
roll for each enemy at 8.7% gives 24, 31 and 37% or more. What rolls in the later window, and at what chance, is not
settled.
