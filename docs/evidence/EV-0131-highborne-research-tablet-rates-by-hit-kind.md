# EV-0131: Highborne Research Tablet rates on melee, ranged and spell hits

- Claim: Highborne Research Tablet (260205) procs 2.95 times a minute, a weapon hit priced on its own weapon speed, Auto
  Shot included, and a spell hit on its cast time, 1.5 s at least; the rate is measured. Periodic ticks, damage shields
  and Thunder Clap never proc it.
- Sources: Warcraft Logs, every player wearing the tablet in 55 reports: 1GBhzTR8CZwcJfFV, gBDMLv8J1K7nqkpd,
  FHqYCNcKntvDAjxg, 2HMyR6kfbWKzBhtv, JK1z96cCrj28tpBV, mhbXcFqf9xTaA1NP, JFqRThp6Z4HrnBwf, CF7ztDwY8pJ9KZhn,
  KqZ6xWAfDLGrVjYJ, A7ahGYbzQ8M1ZtF6, apNB26JmYPyhFWjH, rPpHYfq3tNnwJRMd, 8jzXTaqr732V9DxJ, byvnpga9tMq3BfX6,
  NXMZV9xjvWpB7GtP, 1X37tZRzgkmd9rNP, dxZCXJvPLKAB3cnQ, vyWAYhM3CFkRf12B, Zh9LTdGmgbW1BXw3, 1nTQBK4hmcaLbFZV,
  rphdM14FbnTBaj3V, 1ZyhaVDdFtTYWMcC, TkdjxNKfy9wWDhQp, AvNwTdyBbDncFf2h, 7tLY1pVK9wRWPABX, 9zxHaWbYT3gqDQvF,
  apcNj7QY8tPhLw2B, VkLX4CZnc68R2Gb1, wJtPdpxFGQCM9f3c, Q7V24RA9HTg6F38p, 369Yv2NcqrygTpwX, KG62VXF3a97CxtAg,
  gtNvbhF9aZCXyzVk, zmTNyXkfqtpgcM1H, XvjpYan3qxF9WRgz, 7rGZbTJtpAqnBh2Y, xbJBZj14NWCXrhAm, MWrPDAVFCBXzh8Z6,
  VLz78BfZyWa3Cxr2, Dbfd6avYpzwTCx3N, jTX2kZ1K7gzt3dcL, nyMqbY64tL7HhKQJ, 8GNt2hC7v1rWRXyc, cD6JqWvBhk19xL3r,
  BRxGa38JnVhPKw4A, jXTVfy3ApWKPYwDG, jrTKDZcxNBvtdAP6, H2T7NDqBgm6LxGYK, LyjHDc4nCYmBMRNX, QRCgnhck3KVq6vBT,
  4Tc3yBNVYn7wW8RJ, daT4FcQgCYw2WBq6, RY1yA2PFCBwb4MZa, 6kN78G2xfqy3vwTP, MypmYhB3RX9FJ7Kx; all fights
- Game build: 1.60.1.70888 for the 7 reports from 2026-11-10 to 2026-11-12 (the 2026-11-10 one possibly 1.60.1.70862)
  and 1.60.1.70938 for the 48 from 2026-11-13 to 2026-11-15, across the hotfixes from 113890 to 114102; the hotfix
  each player ran is not recorded
- Date: 2026-11-10 to 2026-11-15
- Pull request: #1051, written up in #1053
- Later count: the rate, §Count of the 2026-11-06 to 2026-11-22 reports

## Method

The reports hold 76 wearer streams, 73 runs once a run logged in two reports is counted once, and 14,699 Ancient
Power (1265634) damage events. A run counts from the first fight its gear shows the tablet in. A stretch of more than
5 minutes of combat with no proc is left out as the tablet swapped away (63 of 2,851 combat minutes; one run stops
proccing after 10 minutes with no gear change logged).

Ancient Power is a missile, and the log has no event for its launch: its damage lands about 100 ms after a melee hit
and 300 to 800 ms after an Auto Shot lands. So a hit counts only when it is isolated, with no other damage event of the
wearer and no Ancient Power within 400 ms of it (melee) or 1,200 ms (ranged and spells), and it procced when Ancient
Power lands inside that window after it. Which hit a proc belongs to then has one answer, and a hit's timing does not
depend on its roll. Dual wielders' white swings are split by hand on their damage before crits. Druid melee is taken at
the form's speed, 1.0 s in Cat Form and 2.5 s in Bear Form, told apart by the white swing interval.

Only the wearer's own cast hits on their first target count as spells. Left out are an area spell's pulses (Volley,
Blizzard, Consecration, Rain of Fire, Arcane Explosion, Chain Lightning and the like), a second target of the same hit
within 10 ms, a hit with no cast event of its own, and Windfury, Seal and poison hits. An area spell's chance per target
is unmeasured.

Each rate carries two 95% intervals: one from the hits (Wilson for a chance, profile likelihood for a rate per minute)
and one from 2,000 draws of whole wearers, which also carries how much wearers differ. The margin is the wider half.
With no procs the wearer draws are all zero, so the bound from the hits is the one that holds.

## Rates

| Hits                                          | Procs / isolated hits | Wearers | Rate     | 95%, hits    | 95%, wearers | Margin    |
| --------------------------------------------- | --------------------- | ------- | -------- | ------------ | ------------ | --------- |
| Main hand, white swings and specials          | 1782 / 14061          | 45      | 2.71 PPM | 2.59 to 2.82 | 2.58 to 2.82 | ±0.13     |
| Main-hand white swings                        | 1135 / 8766           | 45      | 2.77 PPM | 2.62 to 2.92 | 2.63 to 2.91 | ±0.15     |
| Melee specials                                | 647 / 5295            | 40      | 2.60 PPM | 2.42 to 2.80 | 2.32 to 2.86 | ±0.28     |
| Off-hand white swings                         | 238 / 2538            | 12      | 2.94 PPM | 2.60 to 3.31 | 2.45 to 3.44 | ±0.50     |
| Druid forms, on the form's speed              | 206 / 3471            | 6       | 2.83 PPM | 2.47 to 3.22 | 2.45 to 3.53 | ±0.70     |
| Auto Shot, on the ranged weapon's speed       | 392 / 3265            | 14      | 2.51 PPM | 2.28 to 2.75 | 2.34 to 2.73 | ±0.24     |
| Ranged specials, on the ranged weapon's speed | 39 / 412              | 20      | 2.00 PPM | 1.46 to 2.65 | 1.35 to 3.45 | ±1.46     |
| Spells, direct and landed, as a flat chance   | 138 / 2023            | 16      | 6.82%    | 5.80 to 8.00 | 5.92 to 8.03 | ±1.21 pts |
| Wand Shoot, as a flat chance                  | 38 / 426              | 7       | 8.92%    | 6.57 to 12.0 | 7.09 to 11.1 | ±3.09 pts |
| Periodic ticks                                | 0 / 1307              | 36      | 0%       | 0 to 0.29    |              | +0.29 pts |
| Damage shields (Thorns, auras, shield spikes) | 0 / 379               | 29      | 0%       | 0 to 1.00    |              | +1.00 pts |

On the main hand the chance per hit rises with the weapon's speed, from 5.3% at 1.0 s to 15.3% at 3.5 s, and a rate
per minute fits the hits far better than one flat chance (log-likelihood 91 higher for white swings, 35 for
specials). Spell hits show no such rise: 5.5% at a 1.5 s main hand, 8.4% at 3.0 s, 6.1% at 3.5 s, and a rate per
minute on the main hand's speed fits them worse than a flat chance (1.51 PPM, log-likelihood 3.4 lower). Wand Shoot
spans too few speeds to tell the two apart.

## Auto Shot and haste

Hunters shot faster than their bow's speed, three in four Auto Shots 10 to 50% faster. Over 2,415 Auto Shots fired in an unbroken string (each 0.5 to
1.15 times the bow's speed after the one before), a rate on the bow's speed (2.54 PPM, 95% 2.27 to 2.82), a rate on the
hasted interval (3.06 PPM, 95% 2.74 to 3.40) and a flat 12.1% fit equally well, within 0.2 of each other in
log-likelihood. Most wearers shoot a 2.8 s bow, so the speeds span too little to tell them apart.

| Bow speed over interval | Shots | Procs | Chance | 95%          |
| ----------------------- | ----- | ----- | ------ | ------------ |
| below 1.10              | 405   | 61    | 15.06% | 11.9 to 18.9 |
| 1.10 to 1.25            | 1023  | 125   | 12.22% | 10.4 to 14.4 |
| 1.25 to 1.50            | 792   | 81    | 10.23% | 8.3 to 12.5  |
| 1.50 and above          | 195   | 26    | 13.33% | 9.3 to 18.8  |

## Damage

Ancient Power crits at the wearer's own rate, 12.4% over all procs and from 5.5% for warriors to 16.2% for druids, and
2.7% of procs miss or are resisted. The median hit per class runs from 70 to 81 against the row's 69 to 81, and a
mage's (77) sits inside the row's range as a warrior's (70) does, so spell power adds nothing.

## Rate

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: equip proc 1318159
- Rate: 2.95 PPM, measured

24,735 procs on 214,662 rolling hits, 87 wearers, on weapon hits: all hits 2.94 [2.90, 2.98]. The rate is all hits,
which the one-hand white rate agrees with. One-hand white: 2.88 [2.77, 2.97], 1,532 procs on 10,798 lone units, 46
wearers. Dual-wield white, which sets no rate: 2.86 [2.75, 3.00], 2,521 procs on 26,614 lone units, 57 wearers. Lone
Auto Shots: 2.75 [2.55, 2.98], 664 procs on 5,206 lone units, 68 wearers. Specials: 2.89 [2.77, 3.03], 1,859 procs on
14,560 lone units, 88 wearers; Sinister Strike gives 702 procs on 5,521 of them. Two-roll abilities, priced as one roll:
2.68, 37 procs on 296 lone units (Execute 13 on 100 and Mongoose Bite 24 on 196), left out of the rate. Travel time: the
proc lands a median 379 ms after its hit, 98th percentile 1132 ms, window [-20, 1142] ms (1,455 isolated hits). Dropping
the swapped stretches (167 active minutes, 3% of hits, in 9 runs) moved the read from 2.85 to 2.94. The paladin check,
which sets no rate: lone white swings 3.00, 142 procs on 941 lone units, 13 wearers. Left out: 11 of 187 runs in Cat or
Bear Form. 6 runs logged in two reports count once. The one-hand white rate reads on the melee travel window, [-20, 276]
ms, and checks the all-hits rate; the lone Auto Shots and specials read on the ranged one, [-20, 1208] ms, past 400 ms.
The owner takes the rate as measured (2026-11-23). Casters' spell hits, each priced on its cast time and 1.5 s at least,
read 2.84 [2.47, 3.24] on 445 procs over 5,958 lone units of 24 wearers, and fit better than a flat chance per spell
hit, 7.47% [6.52, 8.63]. Of the uncast pulses Flamestrike's rolls it (73 procs on 896 lone units, 57 predicted) and
Ignite's and Blizzard's do not. Thunder Clap never rolls it: the tablet investigation on the same sweep's v16 run
(`runs/2026-11-22-2335-v16`) found 0 procs on 1,370 lone Thunder Clap hits, and 0 attributed in its per-kind fit across
3,929 Thunder Clap hits. A warrior's Thunder Clap always lands batched with its debuff, so the lone count is a pass over
those batches. Debuff-only landings roll nothing: 0 procs on 728 lone melee landings and 1,298 lone ranged ones, where
the white chance predicts 71 and 175. Per-class reads are left out, since casters' runs mix spell procs into the
weapon-hit count. Reports from 2026-11-10 to 2026-11-22.

Reports and fights: 12MCBnGNPTK9pwZD 1-16; 1GBhzTR8CZwcJfFV 89 fights from 1 to 99; 1JLq46yg7wYNhan9 1-28;
1Wc4yTf2MQPxwXDa 1-4; 1X37tZRzgkmd9rNP 1; 2DC7xQNm9wHJnVF4 6-7; 2HMyR6kfbWKzBhtv 1; 2mXywBCLNJHRW1xc 47-51;
369Yv2NcqrygTpwX 1-2, 5; 3CBqGay6HDMWbcd7 1; 3qfDVTZ1J29zPChn 1-23; 4LzF3wqX2ZA7hB1f 1-7; 4Tc3yBNVYn7wW8RJ 31;
4f2ygdYCzFAvqc6K 6-8, 10, 12-14, 16-18; 4zyTBfqQpJ6MrCjv 1-2; 6FVmWYfq7DG2pzwJ 1-7, 9-10, 12; 6jcxdQpGHvgTmVC2 1-2;
6kN78G2xfqy3vwTP 1-6; 6kZMyGqj2PhpLdT4 1-3; 7rGZbTJtpAqnBh2Y 1-11; 7tLY1pVK9wRWPABX 1-21; 82WLwYNzvjB3bXM6 1-2;
8FBqfvLyWTdKxbZh 1-38; 8GNt2hC7v1rWRXyc 1-10; 8Gw7m39ZXFz1y2fv 1-32; 8jzXTaqr732V9DxJ 3-7; 96tBVjg8RYJNMa7r 1;
9BNc8nbtQ2ZyWjkJ 2-3; 9H72ZkdBGfhCy6P3 1-21; 9JT1hKA7GmpCjNad 1; 9QDMdpt1bmcAaR4q 11-44, 46-49; 9kvqL462VaHDpMzR 3;
9zxHaWbYT3gqDQvF 8-18; A7ahGYbzQ8M1ZtF6 14-22, 24-30; AT2jv6mC1wL4KFkJ 2-3, 5-6; ATMmnV2tC3ZPb6G8 1, 3; AvNwTdyBbDncFf2h
1-19; BDMq1zmbXQW9K6ha 1; BRxGa38JnVhPKw4A 1-6; Bhm1dHcrJtR38MDL 1-2; CDKVcP6ghZJtv8AW 3-6, 8-16; CN68qfBaY3MLXkGP 1-2;
CqPk8KgFWYTMnw3D 1-4; Dbfd6avYpzwTCx3N 5; Ddx2RWnArCKf6mqP 46; DhyxXqwHQJYFnCaW 1; DmjBrYd7M9VZwCHF 14-22;
FHqYCNcKntvDAjxg 12-14; G29y3QBWV7XDR6mj 1-11; G921gkqV4PQzjHfp 1; GfH7XvV3rDKzT9Ly 1-5; Grm31RaDY9H6JLz2 1;
H2T7NDqBgm6LxGYK 1-2; HB7yCLD3wvVx4fj9 1-5, 7-15; KG62VXF3a97CxtAg 5, 7-9; KVLgwAr9bB1dJyNR 1-2; KdQtvTWfGpD97NMc 1;
KqZ6xWAfDLGrVjYJ 1-3; L6XGbAdMqBc19KQ7 1-9; L7tdvhcYjgwHTrMz 1; LFNZrg2Cpty63BnK 1-2; LyjHDc4nCYmBMRNX 1-2;
MWrPDAVFCBXzh8Z6 1-9; MypmYhB3RX9FJ7Kx 1-3; NB4ag2CcKryDT91x 1-5; NXMZV9xjvWpB7GtP 1, 3, 5-10, 12-107, 109-116, 118-122;
NbwyAxHKzhYXV6L2 1; NcxbzW4kJYQ7DGhj 1-3; NdTXJzkWHLZrVQYm 1-3; PDFVdp172Kwz8aqj 1, 3, 5-10; PThANm6GzytHYDrd 1;
Q7V24RA9HTg6F38p 1, 3-14; QAqvNHwD3Yt4VZf1 1-3; QDaBR8pV4dmwqnr9 1-17; QMLY8q2XNVry7bpj 42; QRCgnhck3KVq6vBT 1-2;
QZ3hRyAMbDrj9cHz 7-9; QbhWxjGRkF9Lmzq4 1-4; QjwpA7DPVfCZKkMy 1-42; QqKMnRmb7DrFXC36 1, 3-6, 8-9, 11-12, 14, 17-19;
QwmphLYJXT7GazMd 9-40; QzjNR43MZh6XWDkP 1-2; R792cjGCzgx4NaTM 1-2; RFtjqcPd9yKr26wh 1-4; RY1yA2PFCBwb4MZa 2-4;
T92NqcgBXaZrKvCm 1-5; TFARWN4Pc1JzGwZx 1-2; Ta4fLJj32kWzmXgw 1-29; VLz78BfZyWa3Cxr2 1-9; VTGmWXnbAK9Q1jHp 2-10;
VkLX4CZnc68R2Gb1 1-13; VtjvpP8bwcgCYTXF 3-18; WGbpC3Qz84JPa7jN 1; WqfyDhVHxT93Q7Pt 9-30; X3v8V14wrnzdGyFA 1;
XWbYtZ6mTHND2G3P 1-3; XrpxL4B8mPTWQkFK 1-3; XvjpYan3qxF9WRgz 1-9; YwJVCfGj3hnFpRy1 1-2; ZaPqVJn1bQ894pxT 1-3, 5-17,
19-52; Zh9LTdGmgbW1BXw3 3-4; ZqtpzHLbAQKWMcwN 1-3; a1ZGpngJX3DBVPNM 3-4; apNB26JmYPyhFWjH 1-37, 39-89; apcNj7QY8tPhLw2B
13, 15; bGTptRMWJyg4N7HQ 1-2; bGq3mN7JgCDzcyMZ 1, 3-5; byvnpga9tMq3BfX6 1; c2pg17TAXmaPC4WV 1, 3; cD6JqWvBhk19xL3r 1-2;
cbTt1zxajFfNVPDA 1-4; ch87CPWVqaxZYfz3 1, 4; daT4FcQgCYw2WBq6 1-2; dnDg7j9kRw6WtbML 55-61, 63-69; dxZCXJvPLKAB3cnQ 1-2;
fMPL97aCA1ngvKQk 16-20; fjPTxNBrbkJ6v3Q1 14-15; gBDMLv8J1K7nqkpd 3-25, 27-70, 72-74, 76-157; gFnZNhfJvGmXVBcx 1;
gw8fpa1YVkv4WP32 1, 3; hgmcV2xfkQFbWMN8 1-3; hnZqDRzfyXtvPkw1 1; jTX2kZ1K7gzt3dcL 1-7; jTmX9Dg4FCNWyd7q 1-6;
jXTVfy3ApWKPYwDG 1-22; jn3dwzWPZ6r2ky1K 1-3; jrTKDZcxNBvtdAP6 1-12, 14-20, 22; k4K3LG9jNv1YhJmb 1; kNDzgqdc4nQB9vpA 1;
kaPpXgt1bWJBr4z2 1-3; mhbXcFqf9xTaA1NP 1-8; mzYrQXBx6Gy9b8KR 1; nJDNV6Md3bqpXZT9 1-3; nJZvaYCj2brRc7Fw 1-24;
nX1qbatTzGZ3CAwB 1, 3-7; nZBfMADvjgVrWwz3 1; nhtYVjb7CAw9DvQR 1; nyAx3XNDVmYWGdrh 1-6; pQxYvP8khVb7ZMwB 1-3;
pgN9QwMLYVfrqGbv 1-4; ptTk1cMFjxaAhnPK 1-10; qbtZdThvBCXVcfAW 1-6; rPpHYfq3tNnwJRMd 5-19; rYAcza4N7dntgJKW 40;
rfvtaC13gNqRx4Az 1-3; rphdM14FbnTBaj3V 1-29; ry41NG3KdBnRzCAw 1-5; tNjpX1d9L4A8CJBw 2, 4; tfVvJR3aZFgz2CXn 2-3;
tmHdvqxYj1z28PFy 1-9; v1bPdgDtXypHncfN 1-4; vGp9HFWy2ka1nzP6 1-3, 5-6, 8-12; vfBK4pHa9TtyG67w 1-3; vyWAYhM3CFkRf12B 1-8;
wDkZ829XMYLnmKFc 1-10; xWDkmwXKrRcdM7Jv 1; xbJBZj14NWCXrhAm 1-4; xy2Va7jMb9P1vDcX 16-22, 24-30, 32-37, 39-42;
y2XQ6am8TtH1JdWN 1-3; yFGXLWVAKxw61gDP 1-4; yPVkKpqtfJM27WFB 6-7; zKgGQA6BqvcJtYZ9 1-3; zM4Dd8QKwBATnJR9 1-33;
zVAHZcQtTdbhMfLN 7-8; zmTNyXkfqtpgcM1H 1-15.

## Count of the 2026-11-06 to 2026-11-22 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights §Rate lists, from every wearer
  stream on disk for the tablet; 3,714 reports in all for the count, 165 of them for the tablet
- Game build: 1.60.1.70862 to 1.60.1.71037 by report date (70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11
  02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 748 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 38 of them for the tablet. The
  sim's database is 1.60.1.71037, push 114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #1180

The count is EV-0183 §Method, run on the tablet.
