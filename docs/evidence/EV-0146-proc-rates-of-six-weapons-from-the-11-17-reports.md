# EV-0146: Proc rates of three weapons from the 2026-11-17 Warcraft Logs reports

- Claim: Axe of the Deep Woods procs at 1.9, Barman Shanker at 1.95 and Blunderflame Bone Bow at 1.95 procs per minute,
  each priced on each landed hit's weapon base speed; Axe of the Deep Woods' rate is measured, and the other two are
  observed, from wearers of one class only. Barman Shanker's Rend deals twice its damage when applied from behind the
  target.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), 66 reports in all; Barman Shanker's section names the
  reports of its Rend ticks
- Game build: 1.60.1.70938 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-17
- Pull request: #1091
- Later count: the rates of Axe of the Deep Woods, Barman Shanker and Blunderflame Bone Bow, §Count of the 2026-11-06 to
  2026-11-22 reports

## Method

EV-0136's method on its reports and 1,003 more listed up to 2026-11-17: a proc is an event of the item's proc spell,
counted once when it logs several events within 50 ms; the hits are the wearer's landed, non-periodic hits of the item's
hand (white swings, melee-defense abilities, seals and Windfury Totem hits on the main hand; white swings on the off
hand; Auto Shot and ranged-defense hits on the ranged weapon); a dual wielder's swings are split between the hands by
swing timing; a run logged in two reports counts once and a stretch where the item was swapped away is dropped. The
interval is the wider of the per-hit Wilson interval and 2,000 bootstrap draws of whole wearers. A rate is measured when
5 or more wearers of more than one class give it and its interval lies within 10% of it either side, and observed
otherwise. The hits a dual wielder's split gives are fractional, so each table's rows are rounded and sum to within a
few hits of its total.

Julie's Dagger leaves out its one warrior run, since a warrior's off-hand Whirlwind rolls the dagger and is not counted
as a hit here (EV-0136). Barman Shanker has only rogue wearers and Blunderflame Bone Bow only hunters.

## Axe of the Deep Woods (item 811, spell 18104): 1.9, measured

- Count: §Count of the 2026-11-06 to 2026-11-22 reports

8,754 procs on 101,704 rolling hits, 45 wearers, 64 runs, in the main hand: all hits 1.91 [1.87, 1.94]. The rate is all
hits, which the specials rate agrees with. One-hand white: 1.79 [1.27, 2.44], 39 procs on 485 lone units, 2 wearers (a
Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.76 [1.57, 1.95], 351 procs on 4,426 lone
units, 41 wearers. Specials: 1.86 [1.72, 2.06], 422 procs on 5,031 lone units, 44 wearers; Sinister Strike gives 292
procs on 3,446 of them. Two-roll abilities, priced as one roll: 3.09, 40 procs on 288 lone units (Execute 33 on 225 and
Mongoose Bite 7 on 63), left out of the rate. Travel time: the proc lands a median 153 ms after its hit, 98th percentile
378 ms, window [22, 388] ms (57 isolated hits). Report D46Xm8ncR9WbTFrL logs the same run as DdKN3mXj6gH9yvpc and counts
once. Reports from 2026-11-14 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      26 | 5,903 | 67,196 | 1.95 |
| Warrior |      14 | 2,415 | 29,266 | 1.83 |
| Hunter  |       5 |   436 |  5,241 | 1.85 |

Reports and fights: 38kdWHmLZKVgTPqb 2; 3BWcJRwaMgH92h81 1-12; 3HpcAMkBxbJmVWzR 1-11; 3d87Zat9RLCnwDBm 1;
4ChPnBrVK2fJLMaX 1-2, 4; 8hBd9Wz3AXHwZf21 1-2; 8vwhgRM9FTLBD2rN 1-3; 9kb8p2jZ4YXmGQCc 1; B4QqHnGXwzpRfaYW 1-4;
BYXH4tk9GMW2C6Z7 1; BfYZrMW4xdLvk36F 2-5; D46Xm8ncR9WbTFrL 1-71; DAgp7nkzcLBXV3GM 1; DdKN3mXj6gH9yvpc 1-71, 73-75;
G6tcKk2rJvQnA3gP 1-3; GB6fVLkvzPhxYJcZ 1-2; GpXLwmPhVY8rxHgR 1-5; HB362hQ7pWVGLFCM 1-2; HwyJMCLKfbY1PkmG 3-8, 10-38;
LbPgnwx637h8JypD 1-2; McPvpTAN6jzKWxrt 1; NB4ag2CcKryDT91x 1-5; P4g9TMQX1axN8WwD 1-2; PJW2CGRQ8BAqVyag 1-2;
PW8YHxXM2mCaT63q 1; V8F7qLQXzpPAmcyG 1-4; VMfN97kLXxyTpCwW 1-10; WkrD6xYbcp8JZQBm 1-2; Xagb2QHKCf6rDwA9 1-5;
Xjrt9FRMWqJyKdpa 2-3; Z1d98D2CFxk3cyh6 1-3; a6t8yQpMdJRwDvhj 1; aArB7Nt1fvDdKn3z 1-2; c8mjK4TdzV2M7XLF 1-5, 7-46;
cAP3RrZDxBW8d67V 1-3; caR1K8kYbCxvzthT 1; dHBxA1mTK3jwtR7q 1-2; fcghxMTW4VXFQNHk 13-24, 26-53; gtJRYfmNcKLj9XTv 1;
hMKNVdYbX483ZvHw 1-9, 11-16, 18-19; kcy8pLfPGDgQ3dAX 1-3; nYXJNRv9ZBjGKtkV 1-4; nfC91mcXzG4Dq8rd 8; pFH4amZPxkBhyTtV
1-2, 4, 6-34; pVaLq4TbGH2QFtfM 1-7; qKtgYkmZMVjnLBfr 1; qWhzDXTB1nKxk9JH 41-46; t9rhYHPwgcv1fpZb 1; tHmvawdXAzRC72qD
1-12; tfVvJR3aZFgz2CXn 2-3; twbKV4fgDyW3XHmn 1-2; vGp9HFWy2ka1nzP6 1-3, 5-6, 8-12; vHp9QagZGFLBM7PC 1-2;
vZqVmLYQa6jwP3A9 2-16, 18-21; wHt1n2qzvCgkJ7KR 1; xX7APTmCNb3qY4Q8 1; xqnmW3BDpFb28cJ1 1-44, 46, 48-50, 52-53;
xyGf9QY4dwT8JHMp 1, 3-5; xzHZV7MFgNhWj9n4 1; yHz8gXWV2pmvM7Yr 1-9; ycMCt3n9ZqP2bhaH 1-2, 4-8; zBx9XMQPdpW36a2D 1-4;
zVAHZcQtTdbhMfLN 7-8; zaF4hYrwjyCqHKpv 1-2.

## Barman Shanker (item 12791, spell 1300787): 1.95, observed

- Count: §Count of the 2026-11-06 to 2026-11-22 reports

3,368 procs on 51,494 rolling hits, 32 wearers, 40 runs, in the main hand: all hits 1.96 [1.91, 2.01]. The rate is all
hits, which the specials rate agrees with. Observed: every wearer is a rogue. One-hand white: none. Dual-wield white,
which sets no rate: 1.50 [1.33, 1.67], 543 procs on 10,838 lone units, 32 wearers. Specials: 1.54 [1.29, 1.92], 167
procs on 3,248 lone units, 32 wearers; Eviscerate gives 99 procs on 1,808 of them. Direct: window [-11, 100] ms, median
0 ms, 98th percentile 1 ms, 415 isolated hits. Dropping the swapped stretches (52 active minutes, 3% of hits, in 4 runs)
moved the read from 1.90 to 1.96. Report VW6NT73MXFw8r4CQ logs the same run as LNbjwpJhRTZzGP3C and counts once. Reports
from 2026-11-14 to 2026-11-22.

| Class | Wearers | Procs |   Hits |  PPM |
| ----- | ------: | ----: | -----: | ---: |
| Rogue |      32 | 3,368 | 51,494 | 1.96 |

Reports and fights: 13WH2NcCdqJD7L6X 1-2; 1Wc4yTf2MQPxwXDa 1-4; 7J3gB2PzrMykZQvh 2-5; 9zxHaWbYT3gqDQvF 12-18;
APf76npVbQzxydrq 1-11, 13-23, 25-30, 32; BPVxqXWLc3FJTpMh 7-10, 12-14; Bz2NMjHnY7PD6ykX 1-4; DhyxXqwHQJYFnCaW 1;
DmjBrYd7M9VZwCHF 22; FK2TQv9HJyhmLjBa 1-2; G1dJbncV4KXCTg69 1-3; G921gkqV4PQzjHfp 1; GbNrpkt9fVKv74wQ 1-2, 4-33;
JPLDVkNF7Y43nyfp 1; KHDZjxmTvG32bpJc 1-2; LNbjwpJhRTZzGP3C 8; Nkq1WZc8gATJRHnF 1-3; PJpRmw7bf8WvDxA9 1-3;
PcJyYBT7WGMqpLm8 1-3, 5-6; QzjNR43MZh6XWDkP 1-2; T6vkgVaD8YqhyLzf 1-4; VGXLwZaQNTDg91dR 1; VHd1WvxAF7myMDaf 1-2;
VW6NT73MXFw8r4CQ 8; WXgrtaGqh73PRNZJ 1-2; WfKv2TphtRrzNA3d 7-10; b9VpCT4QgYDqxNKH 1-4; bp8jCcfTa7g2P3xX 1;
fFNJ7HtK49hLDvCd 1-5, 7-9; k4K3LG9jNv1YhJmb 1; m7FyYPBwvMngHbDk 6-11; nZBfMADvjgVrWwz3 1; nqymxDJzKcQGbB7W 1-2;
p6mVRBzvWMqnQZFc 2-12; qA612gBzDRyHt7mP 5-37, 41; thqTyPArmkxfMnK7 5-36, 38; vH9PnJDV81wrthRj 2-5; vfBK4pHa9TtyG67w 1-3;
wafLQr31X6VYCmNZ 1-3; z6tKVbjgWGpdm3Yc 1; zp9yVrtfTNZJCv7g 5-7, 9-10.

Its Rend ticks before mitigation fall in two groups, 12 to 14 and 26 to 28: 669 and 761 ticks in GbNrpkt9fVKv74wQ,
9zxHaWbYT3gqDQvF and zp9yVrtfTNZJCv7g. The doubled ticks are the Rends applied from behind the target, as 1300787's
second effect states.

## Blunderflame Bone Bow (item 275439, spell 1296730): 1.95, observed

- Count: §Count of the 2026-11-06 to 2026-11-22 reports

915 procs on 9,969 rolling hits, 13 wearers, 15 runs, in the ranged slot: all hits 1.97 [1.85, 2.10]. The proc's flight
reaches the 1,000 ms isolation, so no lone-unit rate is measured and none checks the all-hits rate. Observed: no
lone-unit rate checks it; every wearer is a hunter. Travel time: the proc lands a median 608 ms after its hit, 98th
percentile 986 ms (191 isolated procs). Dropping the swapped stretches (6.0 active minutes, 2% of hits, in 1 run) moved
the read from 1.94 to 1.97. Reports from 2026-11-15 to 2026-11-22.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Hunter |      13 |   915 | 9,969 | 1.97 |

Reports and fights: 1B7XZcx2mgRaDJVn 1-44; GxgMyQYRHX7f1cFT 1-9, 12; HqYP86CdQZtkB2Gf 5-10; KdQtvTWfGpD97NMc 1;
QRCgnhck3KVq6vBT 1-2; QwpTdtmLyRcZCDW2 1-10; VK83G7MdBhLRFAfa 1-42; cL3KMWpnXyvDrqa2 1-38; ch87CPWVqaxZYfz3 1, 4;
dMxyPfarzTc3tLbk 6; jJca4GYpvxmbWR13 1-3; nqDNvyMgVh2xfQtF 1; qD3YA6bC4jhfFHnB 1-8; tfVvJR3aZFgz2CXn 2-3;
x9v27CNzhRgMkXG6 9, 11-12.

## Count of the 2026-11-06 to 2026-11-22 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every wearer stream on disk for its weapon; 3,637 reports in all for the count, 119 of them for these
  weapons
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 735 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 43 of them for these weapons.
  The sim's database is 1.60.1.71037, push 114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #1180

The count is EV-0183 §Method, run on these weapons.
