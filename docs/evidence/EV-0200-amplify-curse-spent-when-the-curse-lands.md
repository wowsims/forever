# EV-0200: Amplify Curse is spent when the curse lands; a miss or a full resist keeps it

- Claim: A Curse of Agony cast under Amplify Curse (18288) spends the buff when it lands. A cast that misses or is fully
  resisted leaves the buff up for the next curse. The sim applies the same rule to the other curses the talent names
  (Curse of Weakness and Curse of Exhaustion), which the logs hold no casts of under the buff.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the warlock streams on disk in the reports whose
  abilities list Amplify Curse, 74 streams after one report uploaded twice is dropped; the 67 reports and fights
  §Reports lists hold a Curse of Agony cast under the buff
- Game build: 1.60.1.70938 to 1.60.1.71037 by event time (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); the hotfix each player ran is not recorded
- Date: 2026-11-13 18:13 to 2026-11-22 23:03 UTC
- Pull request: #TBD

## Method

Each warlock's own events. A buff window runs from Amplify Curse's applybuff to its removebuff. A Curse of Agony cast
(any rank) inside a window counts. It landed when the curse's applydebuff or refreshdebuff follows within 300 ms of the
cast; it missed when the only line is a damage line with hitType 0, and it was fully resisted when that line has hitType 14. The buff counts as spent when its removebuff falls within 300 ms of the cast. No Curse of Weakness or Curse of
Exhaustion cast falls inside a window.

## Counts

| Cast                | Spent the buff | Kept it | Streams |
| ------------------- | -------------- | ------- | ------- |
| Landed              | 461            | 0       | 67      |
| Missed (hitType 0)  | 0              | 4       | 3       |
| Fully resisted (14) | 0              | 11      | 5       |

By build, landed and spent, then missed or resisted and kept: 70938 228 and 10, 70959 232 and 5, 71037 1 and 0. The 15
casts that kept the buff come from 8 reports: 3N42rgnkmyYQJLPA, 4f2ygdYCzFAvqc6K, 81LcPyjXVbm2xGa3, DjtAKCp2dN6TJvHR,
GcbndBCyf1N9m7YT, HyfjpPNRaFhQ328n, bGL8CXN6f7jwqVcP and vqQgaxHKDw6kh4PX.

## Reports

Report codes and the fights with a counted cast. DdKN3mXj6gH9yvpc, an upload of the same log as D46Xm8ncR9WbTFrL, is
left out.

1MA3qWLGa8v2hdXP 1; 1WHqx7bPa8mkGydJ 1; 21TnpDC6HBYPVWKz 2; 3N42rgnkmyYQJLPA 19; 3ytgzKdVNcxfLvAC 4; 4P3qjWzMgvwaLxKH 1;
4VwxHWJ3pgNmvQY2 1; 4f2ygdYCzFAvqc6K 10, 13, 16; 689BmaZJkvfpCVT1 1; 78wWZfmgrJYzFXvD 34; 7LrX43AZapzF18JT 1;
7NwhcAKH8mTgjPM1 1; 81LcPyjXVbm2xGa3 1; 9H72ZkdBGfhCy6P3 1, 6, 10, 20, 21; 9LK48g3hwjXy67PJ 4;
9cQz2ymYMgfL7TwK 71, 132, 137; ACTGrmZgQh38MvFf 1; BkTfV9yAZJX8P7wF 1; D46Xm8ncR9WbTFrL 71; DjtAKCp2dN6TJvHR 1;
DqGVcxR1QTKrMaP9 1, 6, 9, 11, 19, 23, 30, 34; GcbndBCyf1N9m7YT 1; Gf2q7HLBRytjmTkz 1, 3; Gx9NyLPCjwDbV3qc 1;
HYmf1vTW6hVDPytK 13; HyfjpPNRaFhQ328n 1, 3, 10, 12, 14, 16, 24; JbQyngR8NWXKardc 60; Lr12yNYCdD3a7HQW 1, 7;
MBRmC1TnXq86k2Nb 1; N6B2APcTWnzHvfbq 1, 5; N6xzv3FP7jgHtmaw 1; NBKtAxwHZPmF2Mqa 5, 9, 11, 15, 27; PAtk8J2zdTv63Cwr 2;
PmybzCa2j7NXVxng 1; VRrgQJFbdXckaZq7 1, 2; VtGAb9HNFMYhRjXq 1, 6; WTJGk6CmbMP1V8Hf 1; X4wDxLvF3YgfyKW7 10;
XPYV9J6DfrMKxGqh 1, 4; XRJrztAfDK16HYFc 1, 6; Xk697aGVLz4FWcDd 1; YpmfywHN2MTknvJD 1; ZQ2FqgD1dMH9kGcp 1;
ZR7Mbz438PfDvVNT 24; bCzt1Fq24mYcPXkB 1; bGL8CXN6f7jwqVcP 1, 4; c8mjK4TdzV2M7XLF 1, 4, 20, 25, 32; f1WqtFkXGn4pBzT6 1;
gGBmQx9nC18ZD3d2 4, 8, 9, 17, 31, 34, 37, 39, 41, 45, 47, 49, 52, 66, 67, 70, 77, 84, 85, 91, 94, 108, 113, 118, 122,
142, 189, 192, 198, 200, 228, 245, 253, 258, 273; gJXrZFBLYvfACG64 1; gXBhbVRYL2GtJymw 1, 5, 7; jr46gDtLzB8dfXnK 1;
k1taHbwL76Cnxv4m 1; kph1Df7CqFVxQ2PW 1; kqd9ZJbhaFKNLGnV 1; mTw3kz4cgqrv8JpA 1; qMPwcLCNa4f1HVzk 1, 12;
qT3Brna1YkhLCmc7 1; qZwYtNQLPG1xzkJW 1, 7; qa8wcv4jLYZpGARK 7; qkctgyRMaFxXVnmz 1, 5, 10; r6dAtgyVpcMXkL3T 1;
rY72FxaTNgBQZHXp 1, 4, 9; thpcGknfvy49gTz8 1, 6; vqQgaxHKDw6kh4PX 1, 4, 5; xkN3nXgd8RpHybzT 4; y6QcM9ZqPaYAWzX4 1.
