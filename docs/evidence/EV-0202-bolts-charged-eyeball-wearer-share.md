# EV-0202: About a quarter of Bolt's Charged Eyeball's shocks hit the wearer

- Claim: Of the shocks Bolt's Charged Eyeball (item 280032) casts (1315996), about one in four hits the wearer and the
  rest hit the target; each proc casts one shock, never one on each.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the 35 wearer streams on disk, counted in the fights
  whose gear lists the item; 34 of them, from 33 reports, hold a shock, in the reports and fights §Reports lists
- Game build: 1.60.1.70862 to 1.60.1.71037 by event time (70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11
  02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); the hotfix each player ran is not recorded
- Date: 2026-11-10 13:47 to 2026-11-24 11:08 UTC
- Pull request: #TBD

## Method

Every damage line of 1315996 in the wearer's stream, in the fights whose gear lists the item. Every one has the wearer
as its source. A line counts for the wearer when the wearer is also its target, and for the target otherwise (5,019
lines on NPCs, 15 on pets and players). The target's lines include its 6 misses and 56 immune lines. No millisecond
holds a shock on the wearer and one on another unit together. Builds are bucketed by event time against the US dates.

## Counts

| Build | On the wearer | On the target | Wearer share |
| ----- | ------------- | ------------- | ------------ |
| 70862 | 19            | 60            | 24.1%        |
| 70938 | 465           | 1,420         | 24.7%        |
| 70959 | 650           | 1,711         | 27.5%        |
| 71037 | 663           | 1,843         | 26.5%        |
| All   | 1,797         | 5,034         | 26.3%        |

No shock falls on 70888, between its first listing and 70938's. The pooled share's binomial 95% range is 25.3% to 27.4%.
The median of the 34 wearers' shares is 25.9%; 32 of them, each with 72 shocks or more, run from 20.8% to 33.1%, and the
other two read 9.1% (44 shocks) and 42.9% (7). The largest wearer, 2wPQmVqbdCcZW9gv, holds 1,118 shocks at 31.1%;
without it the share is 25.4% (1,449 of 5,713). By class: rogue 26.8% (14 wearers, 4,011 shocks), shaman 26.7% (9,
1,189), warrior 26.0% (3, 550), mage 25.8% (5, 466) and hunter 23.1% (3, 615).

## Reports

Report codes and the fights with a shock. dnDg7j9kRw6WtbML holds two wearers.

1b7PdYR4HhDXBGT6 13; 1wyWtjCrGpQBqgJN 1; 2wPQmVqbdCcZW9gv 1, 3, 5, 6, 7, 8, 11, 26, 27, 30, 41, 52, 63, 75, 87, 99, 111,
113, 132, 134; 7fz9Kk8rY6jyQAbL 1; 9dbAz4PtMKWRhwkq 1; 9zxHaWbYT3gqDQvF 17; AwjKkm6rcg7bDYCW 1, 2; DJfFBc3Rm7Xqnvxr 8;
DdKN3mXj6gH9yvpc 73; HwGDCPy8Kp1TanZq 2; K6VvbDypXFLQJW7d 21; Mtqz1p49BfvnGyAd 8, 12, 16, 21, 29, 32, 34, 42, 45;
MypmYhB3RX9FJ7Kx 1, 3; QMLY8q2XNVry7bpj 42; QZ3hRyAMbDrj9cHz 12, 14; RY1yA2PFCBwb4MZa 2, 4; VHWwATjhrxtnzCyL 1;
WQDN17aAkMT2tf4q 45; WcCH7rnkQdaLZJ6j 1, 3; Wn8wHhaxy31MLzjp 1; bp8jCcfTa7g2P3xX 1; dnDg7j9kRw6WtbML 55, 58, 61, 64, 67,
68; fMLcHzyZvt7QTCmg 1; hwxz4y2NbJfpCtaj 1; kYdgq87bCmXyMKAV 1; mpgnDQbLWXY8jJP1 1; n2HkhcYFrLK1tfZj 1;
nYXJNRv9ZBjGKtkV 1; qFnZWz3fpbdBVQrg 1; qNyQvka41YhVmdHR 1; rJmGPzj4QNTq1hg8 1; tgQ4FTkDX6NRyhmG 1, 5;
wrzB6GtjPaqCKJZW 1.
