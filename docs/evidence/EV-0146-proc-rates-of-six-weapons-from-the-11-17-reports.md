# EV-0146: Proc rates of Axe of the Deep Woods, Barman Shanker and Blunderflame Bone Bow

- Claim: Axe of the Deep Woods procs at 1.9, Barman Shanker at 1.95 and Blunderflame Bone Bow at 1.95 procs per minute,
  each priced on each landed hit's weapon base speed; Axe of the Deep Woods' rate is measured, and the other two are
  observed, from wearers of one class only. Barman Shanker's Rend deals twice its damage when applied from behind the
  target, and its ticks lose what a tenth of the target's armor takes off a hit.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), 66 reports in all
- Game build: 1.60.1.70938 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-17
- Pull request: #1091
- Later count: the rates of Axe of the Deep Woods and Blunderflame Bone Bow, §Count of the 2026-11-06 to 2026-11-22
  reports
- Later count: Barman Shanker's rate and its Rend ticks, §Count of the 2026-11-14 to 2026-11-25 reports

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

8,754 procs on 101,721 rolling hits, 45 wearers, 64 runs, in the main hand: all hits 1.91 [1.87, 1.94]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 1.83 [1.39, 2.36], 58 procs on 705 lone
units, 2 wearers (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.79 [1.64, 1.95], 499
procs on 6,180 lone units, 41 wearers. Specials: 1.93 [1.80, 2.05], 802 procs on 9,217 lone units, 44 wearers; Sinister
Strike gives 394 procs on 4,606 of them. Two-roll abilities, priced as one roll: 3.60, 65 procs on 401 lone units
(Execute 57 on 318 and Mongoose Bite 8 on 83), left out of the rate. Travel time: the proc lands a median 146 ms after
its hit, 98th percentile 350 ms, window [-20, 360] ms (1,052 isolated hits). Report D46Xm8ncR9WbTFrL logs the same run
as DdKN3mXj6gH9yvpc and counts once. Reports from 2026-11-14 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      26 | 5,903 | 67,208 | 1.95 |
| Warrior |      14 | 2,415 | 29,277 | 1.83 |
| Hunter  |       5 |   436 |  5,237 | 1.85 |

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

- Count: §Count of the 2026-11-14 to 2026-11-25 reports

8,668 procs on 132,617 rolling hits, 65 wearers, 100 runs, in the main hand: all hits 1.96 [1.92, 2.00]. The rate is all
hits, which the specials rate agrees with. Observed: every wearer is a rogue. One-hand white: none. Dual-wield white,
which sets no rate: 1.81 [1.72, 1.90], 2,556 procs on 42,262 lone units, 64 wearers. Specials: 1.87 [1.75, 2.02], 675
procs on 10,817 lone units, 64 wearers; Eviscerate gives 1.91 (392 procs on 6,162 of them), Sinister Strike 1.89 (237 on
3,770) and Kick 1.56 (46 on 884). Shank rolls none: 0 procs on 8,491 lone units. All hits by period: 1.97 [1.89, 2.04]
before hotfix push 114313, 1.93 [1.89, 1.97] from it to the 2026-11-23 hotfix, and 1.99 [1.92, 2.06] after. Direct:
window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 1,835 isolated hits. Dropping the swapped stretches (129 active
minutes, 3.6% of hits, in 8 runs) moved the read from 1.89 to 1.96. Six reports log a run another report holds and count
once: qTKRwYLvxgJp8Hkt as gbGpK6hx7Xq2rVdn, bV7Tcp3DBzFLrwtk as BPVxqXWLc3FJTpMh, ZyCMmR2WgKGra3xv as VFTC9hrLqm7RJAat,
9C2rP4Hkqwtv7RdQ as 4VphDGkmWdbTcB7P, VW6NT73MXFw8r4CQ as LNbjwpJhRTZzGP3C, and NRF9y6Mv3gTYfw4j as f1VTcBRNpybavLwD.
Reports from 2026-11-14 to 2026-11-25.

| Class | Wearers | Procs |    Hits |  PPM |
| ----- | ------: | ----: | ------: | ---: |
| Rogue |      65 | 8,668 | 132,617 | 1.96 |

Reports and fights: 13WH2NcCdqJD7L6X 1-2; 1Wc4yTf2MQPxwXDa 1-4; 1b7NrqACw9FkV3aQ 1-2; 2mZB3p9bwQdxz6M7 1-24;
4VphDGkmWdbTcB7P 8; 6QZLXm7VfRT8jdH3 4-20, 22; 76zPrWdBawDjnfYk 2; 7GFpkLqrZh9TBb8Y 1-4; 7J3gB2PzrMykZQvh 2-5;
9C2rP4Hkqwtv7RdQ 8; 9HRYgfq8DdPLFhtZ 40-43; 9zxHaWbYT3gqDQvF 12-18; APf76npVbQzxydrq 1-11, 13-23, 25-30, 32;
AgQXzkmVdFcDCJjR 1; BPVxqXWLc3FJTpMh 7-10, 12-14; BcXxkd4J8ACDvgQa 1-65; Bz2NMjHnY7PD6ykX 1-4; CGnwN8KYjxfL6ZrF 1;
CJ8PQgzXn14wdBh6 1, 3; CYbWPp6K1Tq7cVRF 1; DhyxXqwHQJYFnCaW 1; DmjBrYd7M9VZwCHF 22; Dr8kfAv94CKQdtXm 3; F7yPwbhrCaLYV8pN
2-5; FK2TQv9HJyhmLjBa 1-2; FzXGfVJr679TcNRB 2; G1dJbncV4KXCTg69 1-3; G921gkqV4PQzjHfp 1; GbNrpkt9fVKv74wQ 1-2, 4-33;
JPLDVkNF7Y43nyfp 1; JYxQXLkWCqGdjg2c 1-13; KDaNT2AhXZ8xymqH 2; KHDZjxmTvG32bpJc 1-2; KLPVXDZxFW2marcn 1-2;
KfgyHLbX6Zr4QTFa 1-5; KgZMjvNmcWPBXGfx 1; LNbjwpJhRTZzGP3C 8; NRF9y6Mv3gTYfw4j 69-71; Nkq1WZc8gATJRHnF 1-3;
PH47TVZMd3Bh8Jmr 2-4, 6; PJpRmw7bf8WvDxA9 1-3; PcJyYBT7WGMqpLm8 1-3, 5-6; QbagXw1hDWFLkyfB 1; Qjxg7zvYFnKGDBWZ 1-2;
QzjNR43MZh6XWDkP 1-2; RT1Nz9qP7b3LJdW6 2; T6vkgVaD8YqhyLzf 1-4; THQkMvRt3FnZ8wL7 7-8; TN2bP8zQdyxVBDgR 1;
TR2zbj9hk4FGyMxK 3; VFTC9hrLqm7RJAat 2-6; VGXLwZaQNTDg91dR 1; VHd1WvxAF7myMDaf 1-2; VW6NT73MXFw8r4CQ 8; VcKtXwqDkgpx4Gvr
9, 11-64; WXgrtaGqh73PRNZJ 1-2; WfKv2TphtRrzNA3d 7-10; XLgaA9MNDHZRcWdk 26; XVGLhJMyQgzbF4rH 10-48; YVMcL1DBkbNC9hxH
5-40; Z8zMfCBNnbxLwX1y 1; ZMp73LDNGr26hAqc 3-18, 20-34, 36, 39-40; ZyCMmR2WgKGra3xv 2-6; a9Z32rABdTN6Vmqf 1-2;
b9VpCT4QgYDqxNKH 1-4; bV7Tcp3DBzFLrwtk 7-10, 12-14; bWH41fw8CDQAch3K 6; ba9Xj6JFCyM12YTx 25, 27-28; bp8jCcfTa7g2P3xX 1;
dtqaDCLhXYb6BM1r 1-4; f1VTcBRNpybavLwD 1-3; fFNJ7HtK49hLDvCd 1-5, 7-9; fGCQVctRYK6g7LXx 1; fprTQwWyDJK7gamC 1-5;
gbGpK6hx7Xq2rVdn 1, 4-8, 11-13, 15-17, 19-21, 23-34; hFd38mrtGvzTXMWV 1-4; hzQRLkBxWCcdPvD1 1-5; j3dqRfHCL86cN4Fg 1-2;
jT9kHabVqcFvdQyR 1-12; jdrpwKPZAh6TYqy2 6-38, 40; k12z7AdtHvrNga4K 1-2; k4K3LG9jNv1YhJmb 1; m7FyYPBwvMngHbDk 6-11;
mw1VHbFfXWyPLqk7 1-7, 10, 12; nGCXxajMD47AQVFf 1-2; nZBfMADvjgVrWwz3 1; nazpht9wm4NYxGW3 3; nqNJVadK4vRh6pLW 1-4;
nqymxDJzKcQGbB7W 1-2; p6mVRBzvWMqnQZFc 2-12; pDFrB9xgKcNVRWqj 2; qA612gBzDRyHt7mP 5-37, 41; qKrbykwYfnMcBZ2V 1-2;
qTKRwYLvxgJp8Hkt 1, 4-8, 11-13, 15-17, 19-21, 23-34; tB6pZ1YcqAJ9NjRH 6, 8-35; tFDk4nKWQB67y1z3 4-8, 10;
thqTyPArmkxfMnK7 5-36, 38; vH9PnJDV81wrthRj 2-5; vfBK4pHa9TtyG67w 1-3; wWC7rvKj98knXZDt 5; wafLQr31X6VYCmNZ 1-3;
wjxdJBMkD8N3Z6c4 1-26; z6tKVbjgWGpdm3Yc 1; zp9yVrtfTNZJCv7g 5-7, 9-10; zy3r8qhGAH2BRp41 2.

Its Rend ticks before mitigation fall in two groups, 14 and 28; about a third of the logged ticks are the 14. The
doubled ticks are the Rends applied from behind the target, as 1300787's second effect states. About 14% of the procs
land on targets immune to the Rend (elementals, golems, constructs, skeletons and ghosts).

The Rend's ticks lose what a tenth of the target's armor takes off a hit, while Garrote's and Rupture's ticks on the
same targets lose nothing to armor. Each doubled tick that is not a crit and has no overkill, absorb or block is paired
with the mean share the wearer's normal white swings on the same target lost within 1.5 s of it, and the pairs are
grouped by that share to the nearest tenth. A white-swing share w means armor A with A/K = x = w/(1 - w), so a tenth of
the armor takes 0.1x/(1 + 0.1x) off; a tenth of the share would take 0.1w. Garrote's ticks in the same groups deal 4.6%
to 8% more than before mitigation, a talent's bonus, with no trend in armor.

| White-swing share lost | 0.235 | 0.331 | 0.385 | 0.490 | 0.567 | 0.697 |
| ---------------------- | ----: | ----: | ----: | ----: | ----: | ----: |
| Ticks                  |   806 | 2,905 | 7,020 |   708 |   378 |    91 |
| Tick share lost        |  2.1% |  4.2% |  5.5% |  9.3% | 11.9% | 19.8% |
| A tenth of the armor   |  3.0% |  4.7% |  5.9% |  8.8% | 11.6% | 18.7% |
| A tenth of the share   |  2.3% |  3.3% |  3.9% |  4.9% |  5.7% |  7.0% |

## Blunderflame Bone Bow (item 275439, spell 1296730): 1.95, observed

- Count: §Count of the 2026-11-06 to 2026-11-22 reports

914 procs on 9,962 rolling hits, 13 wearers, 15 runs, in the ranged slot: all hits 1.97 [1.86, 2.10]. Its lone-unit
rates of 50 or more procs read on a travel window past 400 ms, so none checks the all-hits rate. Observed: no lone-unit
rate checks it; every wearer is a hunter. Lone Auto Shots: 1.86 [1.76, 2.12], 136 procs on 1,568 lone units, 13 wearers.
Specials: 2.17 [1.73, 2.91], 107 procs on 1,059 lone units, 12 wearers; Multi-Shot gives 78 procs on 827 of them. Travel
time: the proc lands a median 639 ms after its hit, 98th percentile 1146 ms, window [-20, 1156] ms (132 isolated hits).
Dropping the swapped stretches (6.0 active minutes, 2% of hits, in 1 run) moved the read from 1.93 to 1.97. Reports from
2026-11-15 to 2026-11-22.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Hunter |      13 |   914 | 9,962 | 1.97 |

Reports and fights: 1B7XZcx2mgRaDJVn 1-44; GxgMyQYRHX7f1cFT 1-9, 12; HqYP86CdQZtkB2Gf 5-10; KdQtvTWfGpD97NMc 1;
QRCgnhck3KVq6vBT 1-2; QwpTdtmLyRcZCDW2 1-10; VK83G7MdBhLRFAfa 1-42; cL3KMWpnXyvDrqa2 1-38; ch87CPWVqaxZYfz3 1, 4;
dMxyPfarzTc3tLbk 6; jJca4GYpvxmbWR13 1-3; nqDNvyMgVh2xfQtF 1; qD3YA6bC4jhfFHnB 1-8; tfVvJR3aZFgz2CXn 2-3;
x9v27CNzhRgMkXG6 9, 11-12.

## Count of the 2026-11-06 to 2026-11-22 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every wearer stream on disk for its weapon; 3,714 reports in all for the count, 78 of them for these
  weapons
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 748 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 22 of them for these weapons.
  The sim's database is 1.60.1.71037, push 114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #1180

The count is EV-0183 §Method, run on these weapons.

## Count of the 2026-11-14 to 2026-11-25 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights the Barman Shanker section lists,
  from every wearer stream on disk for it; 105 reports
- Game build: 1.60.1.70938 to 1.60.1.71091 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC, 71091 from 2026-11-24 15:21 UTC in the US and 2026-11-25 02:19 UTC in the EU); 85 reports start after hotfix push
  114313 (2026-11-20 23:47 UTC). The sim's database is 1.60.1.71091, push 114405; the hotfix each player ran is not
  recorded
- Date: 2026-11-14 to 2026-11-25 07:55 UTC
- Pull request: #1214

The count is EV-0186 §Method, run on this weapon alone by the sweep of 2026-11-25, `runs/2026-11-25-barman` in
`forever-sweep-auto`, whose `barman_results.json` holds the count and whose `bleed_armor4.py` pairs the ticks with the
white swings.
