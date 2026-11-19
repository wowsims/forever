# EV-0151: Flame Wrath procs at 1.9 per minute, and its damage takes no spell power

- Claim: Flame Wrath 11809 procs 16559 at about 1.9 procs per minute, priced on the base speed of what makes the hit:
  the weapon's 3.3 s, Cat Form's 1.0 s, and Bear and Dire Bear Form's 2.5 s. A paladin's melee abilities roll it; in
  Cat and Bear Form only the white swing does, and Maul, Swipe, Lacerate, Primal Bite, Claw, Rake and Shred roll
  nothing. No lockout follows a proc. 16559 hits every enemy near the wearer once for the row's 154 Fire damage with no
  spell power share, and crits at the wearer's spell crit chance.
- Sources: Warcraft Logs reports (forever.warcraftlogs.com, zone 3000), every fight in which the wearer's gear
  snapshot holds 11809 in the main hand: 1X37tZRzgkmd9rNP, 1a3Qn7JpLk8AH2hV, 1nTQBK4hmcaLbFZV, 1z6PXMGdQBDvrCqx,
  2wFyAGLf1CqZKnaz, 2wZ3P6y9N4f7kXVq, 3R69PBjxZvcfbnX8, 3qfDVTZ1J29zPChn, 4BFrCYZajATmLfgn, 4f2ygdYCzFAvqc6K,
  6CN9XfdkHq148yZQ, 6VDthbQzAJxqpn7d, 6jHn3xcCZYW9GkMy, 6kJNxp39PWqrghmG, 6yHgkT8LqP42NmxY, 7ChG2VXaALzcw9pW,
  7mnVXCgpyxkbtZGq, 7tbWfYqJFd4vzL3p, 8FBqfvLyWTdKxbZh, 8LaNq1JXfZzy3Dj9, 9dmW4zX2YbnK6CGp, AYwLfvqh9QyNj6zJ,
  B6bvH1fQ4kAcyRzt, BHqXftrKD6AadWh1, Bv6mxRb1fKG9ZgYd, DQb2rzfnNyL148BR, Dagn1fC64vQRwK9N, FJdrDz9w1Vxaf2cT,
  FN8krKfCb3BXWmwV, FhG3AxYL9ZNP4RvV, FtcXATwrHDd9vQ83, HMdLhWY3pNbnXPtF, JKPwHm7fnCp3L14c, JQwbjnVYyGcP4rg8,
  JrqFhHP6wxV7n9yX, KfjzWvhFNCDMLPwx, LQ9zMDqpmknBgAX8, Lt3QCJGKXZYTBFkD, LvDw6BjHCVGgzJyb, M1tPx9DBZGYgCa4f,
  NQv1MAX6mrxjC2nb, Nxtj7m6zcFX9Aq3Z, PdNq3JxntCKgavkm, PtKXrwfj8zBHm2Gy, Pxn7LYfMX4Qd6vNr, TmQyqnM6hN217LZC,
  YCbcQHW378fnMvXA, YZy7L3P6AVhWJN1M, ZHpKVnhPDGyaTrvw, ZJgFdtvGTYQVCqyb, ZwkY3VX4H2aLWpDt, aLwy1zjnrYFhvpDq,
  aWLTAv36Kx9HVPkY, bg169rGdXTzDhLnj, byvnpga9tMq3BfX6, cb9gzXjCK3BfH7qM, gYm2HAtf4bnC8QZT, hQY9xzt7Mr4nG6Ap,
  jKQDqkwmx46ZLBFg, k1gAvx7zVaPpLhYC, kvhJGjQNAWRgBwrm, mbpQYGL8jaqgdrAz, mzVGhWHB3ZyATfC4, p4fMBXZaJz8ctTP6,
  p4waHBZgjkA9mDLn, pcKhkTd9wzArGvfm, pzBvhjtCD9VRf2mP, qF9L7VZHR1QG8r2m, qbrGgBZfzLHFW12k, rPpHYfq3tNnwJRMd,
  tTRm37A4YwgM8nCG, tmHdvqxYj1z28PFy, w6BX8kh4ADbRW2yg, yfKb1MthLFDc4ANG, zmDY4LVX7N8vh9aR; 52 wearers in 78
  wearer-reports (37 druids in 62, 9 paladins in 9, 5 hunters in 5, 1 warrior in 2)
- Game build: 1.60.1.70888 (6kJNxp39PWqrghmG, DQb2rzfnNyL148BR), 1.60.1.70938 (the reports up to 2026-11-17 15:18 UTC)
  and 1.60.1.70938 or 1.60.1.70959 after it (2wFyAGLf1CqZKnaz, 6CN9XfdkHq148yZQ, 8FBqfvLyWTdKxbZh, B6bvH1fQ4kAcyRzt,
  FhG3AxYL9ZNP4RvV, KfjzWvhFNCDMLPwx, LQ9zMDqpmknBgAX8, LvDw6BjHCVGgzJyb, M1tPx9DBZGYgCa4f, Nxtj7m6zcFX9Aq3Z,
  PtKXrwfj8zBHm2Gy, ZJgFdtvGTYQVCqyb, aLwy1zjnrYFhvpDq, aWLTAv36Kx9HVPkY, jKQDqkwmx46ZLBFg, mbpQYGL8jaqgdrAz,
  mzVGhWHB3ZyATfC4, p4fMBXZaJz8ctTP6, pzBvhjtCD9VRf2mP, qbrGgBZfzLHFW12k and part of yfKb1MthLFDc4ANG), dated by when
  the CDN first listed each build against each event's time; the US and EU listed 70959 from 2026-11-17 15:19 UTC
  while CN, KR and TW still listed 70938 on 2026-11-18. The hotfix each player ran is not recorded
- Date: 2026-11-12 14:54 to 2026-11-18 14:18 UTC
- Pull request: #1104

This replaces the Flame Wrath rate in EV-0136.

## Method

Each apply or refresh of 16560 (Fire Shield) on the wearer is one proc. 16560 is also the spell of each hit Fire
Shield deals to an attacker, and those hits are not procs. 16559 logs one damage line per enemy it hits, and two
enemies of one kind share a target id. For the per-swing chance below, a 16559 damage line with no apply beside it
counts too, and events of the proc's spells within 100 ms of each other join into one; 115 of the 4,348 applies came
within 100 ms of another, two procs landing together.

The rate is read off white swings that landed (hit, crit, glance or block) while a gear snapshot showed 11809 in the
main hand, keeping only swings with no other damage line of the wearer within 100 ms. A swing has a proc when one
starts from 10 ms before its damage line to 50 ms after it; 16560 logs 0 to 3 ms before the swing. The same window
moved 700 ms later finds 2 procs on 23,825 swings. A run of swings with no proc longer than ln(1000) over the chance
is dropped as the weapon swapped away, 347 of the 23,825 swings. The interval is 2,000 bootstrap draws of whole
runs. Each form's base swing speed prices the chance: Cat Form 1.0 s, Bear and Dire Bear Form 2.5 s, otherwise the
weapon's 3.3 s. The hunters' few white swings are left out of the table and counted on their own below.

## Rate

| Wearer           | Swings | With a proc | Chance (95%)        | Swing speed |  PPM |
| ---------------- | -----: | ----------: | ------------------- | ----------: | ---: |
| Druid, Bear Form | 16,248 |       1,272 | 7.8% (7.4 to 8.2)   |       2.5 s | 1.88 |
| Druid, Cat Form  |  5,899 |         197 | 3.3% (2.9 to 3.7)   |       1.0 s | 2.00 |
| Druid, no form   |    147 |          17 | 11.6% (6.1 to 18.5) |       3.3 s | 2.10 |
| Paladin          |    934 |          94 | 10.1% (8.7 to 11.9) |       3.3 s | 1.83 |
| Warrior          |    250 |          17 | 6.8% (4.7 to 11.2)  |       3.3 s | 1.24 |
| All              | 23,478 |       1,597 |                     |             | 1.88 |

The chance per swing grows with the swing speed from Cat Form to Bear Form to the weapon, so the rate is procs per
minute. Without the wearer who swung most in Bear Form, Bear Form reads 8.0% (7.3 to 8.5), 1.91. Over every landed
melee hit of the paladins (abilities, seals and Windfury Totem strikes included, priced at 3.3 s), one proc to each
Fire Shield apply or refresh, the rate is 2.05 (1.96 to 2.17). The rate holds across builds: Bear Form 7.5% on 345
swings on 70888, 7.7% on 13,118 on 70938 and 8.0% on 2,889 after 70959 was listed.

Landed hits other than white swings, each with no other damage line of the wearer within 100 ms, by the same window:

| Wearer           | Hit                  | With a proc / hits |
| ---------------- | -------------------- | -----------------: |
| Druid, Bear Form | Swipe                |         8 / 16,153 |
| Druid, Bear Form | Maul                 |          0 / 5,084 |
| Druid, Bear Form | Primal Bite          |          1 / 4,653 |
| Druid, Bear Form | Lacerate             |          2 / 3,153 |
| Druid, Cat Form  | Rake                 |            2 / 635 |
| Druid, Cat Form  | Shred                |            0 / 538 |
| Druid, Cat Form  | Claw                 |            0 / 746 |
| Paladin          | Holy Strike          |           67 / 692 |
| Paladin          | Judgement of Command |           37 / 521 |

A paladin's Holy Strike and Judgement of Command roll near the white swing's chance; in a form, Maul, which takes the
place of a white swing, rolls nothing.

EV-0136 read 5.30 from one hunter in 6yHgkT8LqP42NmxY, counting 202 procs where the report holds 88 Fire Shield
applies: it counted Fire Shield's hits on attackers and a second hit on an enemy sharing a target id as procs of their
own. That hunter's isolated white swings read 20 procs on 170, 11.8%, 2.14.

Fire Shield applies came 100 ms to 1 s apart 50 times and 1 to 5 s apart 508 times, so the 5 s category cooldown
on 16559 locks nothing out.

## Damage

A proc hit up to 19 enemies, once each. Normal hits of the wearers other than druids, by the gear snapshot's spell
damage and Fire damage (items and enchants):

| Gear spell damage |  Hits |  Mean | 5th to 95th percentile |
| ----------------: | ----: | ----: | ---------------------- |
|                 0 |   320 | 158.2 | 138 to 177             |
|           1 to 39 |   507 | 154.7 | 136 to 173             |
|          40 to 59 | 1,052 | 153.3 | 135 to 173             |
|          60 to 79 |   789 | 152.4 | 135 to 172             |

The mean falls from 158 to 152 as gear spell damage rises from 0 to 79, where a spell power share would raise it, so
16559 takes none; the row's 154 with its variance runs 133.5 to 174.5. Druids hit about 10% harder: their Naturalist
talent (17069, 2% more damage of every school a rank) splits them, rank 0 druids averaging 156 and rank 5 druids 170.

Paladins crit 16.1% of 3,549 landed hits, against 15.8% on their own spells and 18.3% on white swings; the warrior
crit 11.2% of 456 against 46.6% on white swings.
