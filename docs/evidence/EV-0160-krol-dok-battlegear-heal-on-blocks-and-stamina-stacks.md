# EV-0160: Krol'dok Battlegear's heal on blocks and its Stamina stacks

- Claim: Krol'dok Battlegear's 4-piece Krol'dok Resolve (1324298) casts its heal 1324296 on the wearer on every
  blocked melee or ranged hit, a full block too, at most once a second; a hit not blocked, a spell, a
  tick, a hit a channel fires and Grievous Garrote never heal. The heal is 18 spread from 14 to 22 outside Strongholds
  and Cities, and a flat 44 inside them (area group 9326), 66 on a crit. The 5-piece Ogre Dominance (1324309) adds a
  stack of 1324308's Stamina on every landed melee or ranged hit taken, one that deals no damage too, at most once
  every 3 s; a miss, a dodge, a parry, a spell, a tick, a hit a channel fires and Grievous Garrote add none.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every character with a 1324296 heal or a 1324308 aura
  event in the 1,012 reports the Reports section lists; all fights.
- Game build: 1.60.1.70862, 1.60.1.70888, 1.60.1.70938 and 1.60.1.70959 by event time against the time the CDN first
  listed each build; the Game builds section has the counts. The hotfix each player ran is not recorded. The 4-piece
  bonus gained its cooldown on 2026-11-16 between 17:40 and 17:50 UTC, in the EU and US reports alike (the Cooldown
  hotfix section).
- Date: 2026-11-09 to 2026-11-20
- Pull request: #1133

## Method

A wearer is a character with a 1324296 heal or a 1324308 aura event: 777 characters, a character being a name within a
region, in 1,012 reports. A wearer in one report is a warrior 568 times, a paladin 486 times, a mage once and of a
class the report does not give 36 times. A fight logged in two reports, the same name and region with fight starts
within 2 minutes, counts once (130 dropped). The heal and the buff both prove the set is worn, so a hit counts only
after a proc in the same fight and before a later one.

The heal and the buff log at the millisecond of the hit that procced them or up to about 35 ms after it, so a proc is
credited to the latest hit taken in the 60 ms before it. A tank takes several hits inside that window, so the chances
are read on isolated hits: a hit with no other hit on the wearer within 100 ms. A white swing is ability 1 from an
NPC. Its hit types are the report's: a hit, a crit, a block, a crushing blow and a glancing blow land, and a miss, a
dodge and a parry do not. An ability's kind is its `SpellCategories.DefenseType` in the game files (1 magic, 2 melee,
3 ranged). Bladestorm 1307066, the Whirlwinds (15576, 15578, 15589, 1323828) and Grievous Garrote 1271761 are counted
apart: a channel or a periodic aura fires their hits, and they read near none for both bonuses where every other melee
ability reads near all. A fight's zone is the report's `gameZone`; it is in Strongholds and Cities when its name is an
`AreaTable` name of area group 9326 in the game files, with the reports' Sunken Temple as The Temple of Atal'Hakkar.
The game files rows are read on 1.60.1.70959 at hotfix push 114286.

## Game builds

| Build                        | Reports | Heals   | Stacks or refreshes |
| ---------------------------- | ------- | ------- | ------------------- |
| 1.60.1.70862                 | 21      | 13,152  | 1,375               |
| 1.60.1.70888                 | 68      | 45,274  | 5,340               |
| 1.60.1.70938                 | 542     | 368,288 | 78,989              |
| 1.60.1.70938 or 1.60.1.70959 | 15      | 4,253   | 2,234               |
| 1.60.1.70959                 | 372     | 149,314 | 64,459              |

The 1.60.1.70938 or 1.60.1.70959 row is the CN, KR and TW reports after 2026-11-17 15:19 UTC, when the CDN listed
1.60.1.70959 for the US and the EU first.

## The 4-piece heal

Blocked isolated hits, by the time since the wearer's last heal, after the cooldown hotfix (2026-11-16 18:00 UTC on):

| Hit                       | Since the last heal | Blocks  | Heals   | Share   | 95% interval   |
| ------------------------- | ------------------- | ------- | ------- | ------- | -------------- |
| White swing, damage       | 1 s or more         | 112,573 | 112,169 | 99.64%  | 99.60 to 99.67 |
| White swing, no damage    | 1 s or more         | 31,054  | 30,958  | 99.69%  | 99.62 to 99.75 |
| Melee ability, damage     | 1 s or more         | 5,791   | 5,771   | 99.65%  | 99.47 to 99.78 |
| Melee ability, no damage  | 1 s or more         | 458     | 454     | 99.13%  | 97.78 to 99.66 |
| Ranged ability, damage    | 1 s or more         | 1,378   | 1,371   | 99.49%  | 98.96 to 99.75 |
| Ranged ability, no damage | 1 s or more         | 308     | 308     | 100.00% | 98.77 to 100   |
| Magic ability             | 1 s or more         | 1,691   | 0       | 0.00%   | 0 to 0.23      |
| Channel hit, Garrote      | 1 s or more         | 1,014   | 0       | 0.00%   | 0 to 0.38      |
| White swing, damage       | under 1 s           | 29,241  | 504     | 1.72%   | 1.58 to 1.88   |
| White swing, no damage    | under 1 s           | 11,571  | 102     | 0.88%   | 0.73 to 1.07   |
| Melee ability             | under 1 s           | 1,847   | 22      | 1.19%   | 0.79 to 1.80   |
| Ranged ability            | under 1 s           | 411     | 4       | 0.97%   | 0.38 to 2.48   |

Isolated hits that are not blocked, 1 s or more after the last heal, heal almost never: 17 of 225,140 white hits,
none of 253 crits and 855 crushing blows, 3 of 32,548 dodges, 6 of 41,883 misses, 3 of 42,423 parries, none of 22,958
magic ability hits and 2 of 29,142 ticks. A melee ability hit not blocked reads 581 of 11,350 (5.1%) and a ranged
one 283 of 5,446 (5.2%); what heals after those was not looked at.

The heal by area, amount plus overheal:

| Area                   | Hits    | 44      | 14 to 21 | 22    | 11  | Crits  | 66 on a crit |
| ---------------------- | ------- | ------- | -------- | ----- | --- | ------ | ------------ |
| Strongholds and Cities | 453,254 | 446,766 | 237      | 6,035 | 171 | 29,681 | 29,388       |
| Elsewhere              | 92,496  | 9,057   | 74,224   | 9,215 | 0   | 4,850  | 583          |

The 44s elsewhere are mostly fights with no zone in the report (8,038 heals, 71% of them 44 or 66) and Ahn'Qiraj: The
Fallen Kingdom (4,547 heals, 87%), a zone name the game files do not hold, whose fights mix both amounts. Every zone a
report names that is out of the group heals 14 to 22, 21 to 33 on a crit, and never 44: Maraudon (81,285 heals), The
Shaper's Terrace (2,867), Wailing Caverns (2,526), Hillsbrad Foothills (212) and Ragefire Chasm (120). The nine amounts
14 to 22 take 9,192 to 9,337 of the heals elsewhere each, the row's 18 with its Variance of 0.5 rolled evenly. Inside
the group the 22s and 11s are half of 44 and of 22, which a healing reduction on the wearer would give; that was not
checked.

## Cooldown hotfix

Before 2026-11-16 17:00 UTC every block heals, inside a second of the last heal too:

| Hit                    | Since the last heal | Blocks  | Heals   | Share  |
| ---------------------- | ------------------- | ------- | ------- | ------ |
| White swing, damage    | 1 s or more         | 100,897 | 100,635 | 99.74% |
| White swing, damage    | under 1 s           | 34,243  | 34,199  | 99.87% |
| White swing, no damage | under 1 s           | 15,561  | 15,507  | 99.65% |
| Melee ability          | under 1 s           | 2,597   | 2,594   | 99.88% |
| Ranged ability         | under 1 s           | 465     | 459     | 98.71% |

The share of blocks under 1 s after a heal that heal, by 10 minutes on 2026-11-16 (UTC), with fights logged twice
kept in this table and the hourly reading below:

| Time  | EU blocks | EU heals | US blocks | US heals |
| ----- | --------- | -------- | --------- | -------- |
| 17:20 | 133       | 126      |           |          |
| 17:30 | 80        | 78       |           |          |
| 17:40 | 28        | 26       |           |          |
| 17:50 | 56        | 5        | 5         | 0        |
| 18:00 | 35        | 2        | 9         | 0        |
| 18:10 | 52        | 1        | 36        | 1        |
| 18:20 | 50        | 0        | 55        | 0        |

Every hour from 2026-11-16 18:00 to 2026-11-17 06:00 UTC reads 0 to 2.5% in every region. The spell store took the
row's 1 s cooldown from that day's hotfixes (the items update of 2026-11-16, commit f8d58d4f19); which push carried it
is not recorded, and push 114115, exported at 18:05 UTC, is consistent with the time.

## The 5-piece stacks

Gaps between two procs of a wearer in one fight start at 2,950 ms: of 150,256 gaps, 67 fall from 50 to 2,950 ms, 723
from 2,950 to 3,000 ms and 6,301 from 3,000 to 3,050 ms. 597 are under 50 ms and were not looked at.

The first landed isolated hit 3.05 s or more after a proc, with the buff below 20 stacks:

| Hit                       | Hits   | Stacks | Share   | 95% interval   |
| ------------------------- | ------ | ------ | ------- | -------------- |
| White swing, damage       | 45,046 | 44,925 | 99.73%  | 99.68 to 99.78 |
| White swing, no damage    | 7,538  | 7,519  | 99.75%  | 99.61 to 99.84 |
| Melee ability, damage     | 1,805  | 1,794  | 99.39%  | 98.91 to 99.66 |
| Melee ability, no damage  | 240    | 240    | 100.00% | 98.42 to 100   |
| Ranged ability, damage    | 982    | 980    | 99.80%  | 99.26 to 99.94 |
| Ranged ability, no damage | 158    | 158    | 100.00% | 97.63 to 100   |
| Channel hit, Garrote      | 95     | 0      | 0.00%   | 0 to 3.89      |

Isolated hits that do not land, and spells, in the same stretches: 10 stacks on 4,793 dodged white swings, 8 on 5,796
misses, 5 on 6,066 parries, 4 on 4,833 magic ability hits and 2 on 6,301 ticks.

At 20 stacks the log drops refreshes; the counts after the first one keep fights logged twice. The first hit after the
cooldown reads 91.08% (39,667 of 43,551 white swings), and the next such hit after one with no refresh logged reads
6.8% (196 of 2,892), which no chance per hit gives. 118 gaps between logged refreshes at 20 stacks run past the buff's
30 s with no removal, and 162 of the 2,097 removals at 20 stacks fall 30 s after a landed hit with no refresh logged
at it. So the 20-stack reading is the log's, and the chance is read below 20 stacks alone. The buff's size change
(`A_MOD_SCALE`) and the ogre form at 20 stacks (1324682, `A_TRANSFORM`) show nothing in a report.

## Reports

The 1,012 reports with a wearer:

13AdjNWkTKpw9DG7, 1GbgDRqrvBpdJwnP, 1L9BQW2AcqdP8bfG, 1LNfVBGYtRZ8PyvM, 1Nry3p4Ltf2ZVhwD, 1PVRtQZCwfTkmzx9,
1QGw9vM2zVCkYZBf, 1Tf47h2MvyczZbVK, 1YJdRn73FrtMwvxp, 1cm2k9qHaRdBMpFx, 1f634QRdjhPvVGtb, 1gRapATxZF9WB3rK,
1hRZ2v8PwKWt7DkX, 1hvXcDfFYnWZyGpV, 1xT7CXcnRg2ZLbFK, 1z4vRDqGdM2cVPXk, 23p9dQ6ctrFBvJAf, 24ZKBL3XhzwJcjGA,
27V6dXhtRxLW8BYa, 2DQGWp6Y8MbzhXLt, 2DrJcgXj7FP8hzVw, 2FnrpzZHh4VbLQqG, 2Gt18qjbgwPWzDTn, 2KYb3mf9hTr8ZRVL,
2LVH4fq1TJky7Dcj, 2M18XZFDCrj7QntB, 2TytDnRqwPXHa3gJ, 2Yh9jPZQKM7nz48D, 2dt8MwPW4JBLH7Nf, 2gXTmYzt3dJ4PMwc,
2mN1WDYhBgXj6nrV, 2nZAcdjGb9xNFk6P, 2tPagAHFJrhwYjvZ, 2wpZMAPcnXC9btDG, 2xKat61ANVJkZg8h, 2xY1D3mCwagdy7Wc,
2xfdGzvVtcX4LWgH, 2y16Vx9LRrpJgXhj, 36XqpAkDvTmjZrhR, 36h91AnYPjTkbxcM, 37BTPjYv6RCbqKfL, 3B4wmh7JGkQq2Ndj,
3BaM92CkJYqL1Xyt, 3CBqGay6HDMWbcd7, 3GT4tz8gxnphcbkD, 3JGPcqwy2XbZvkM4, 3JKV29qG8QNawjFv, 3MY1W4gBv9yXakD6,
3Mr8qtzkhXfgd4my, 3N42rgnkmyYQJLPA, 3XvyFNnx2BDVzMQg, 3YHN8TqBWXQxKp7r, 3cWzXTLmvDh2bC7H, 3fFatLkqGnmjQ8ZW,
3jQ7gBkGHYwRhMKZ, 3jcqW9N6XMyThx2P, 3kpnKNwfzC6h2g1b, 3r7mzpBjFKW9Y8Pf, 3tCdJpg1YTmhVcxF, 3tK2WMC1QGxhA4JH,
3y6Qm2A7a4DvhKJn, 3yjCr7kWwFnVRXTd, 3zTJ2VXvQ8GCyFbc, 47NfdyKHp6GgYhXC, 4CBckLvG7RaxFzYA, 4CVrTcXtBf2yAYH9,
4GRLjCqT9nNxQpJg, 4GgNAjPCMhqmQf7r, 4GvM7BJFwRkHqa8L, 4HhaAC2pyBDzg6kP, 4NqLGxvrZYX2Qwnf, 4PCDJtkXhry9MW13,
4QaCxqNLR7f291tz, 4RVqrpXkh7dKftAC, 4cq1YLNFBvj73tMZ, 4hBNrz7dtx9mMqby, 4nNqbfmrCT2Q9PpH, 4qZXFGfc1wCMQKL7,
4wd9fmTPrNGjthyL, 63RLWvyXFJwmbZ9k, 6Ab23vqJ9z1pRmH8, 6AvVhawr1QjRxWBz, 6CRda1YkV2m4fpbc, 6DZv2FAgWq7YrV3j,
6Df3NywAZRQkCtbx, 6HJWv3DLN7Ba8T4c, 6Krp8bkfx4cRjBaz, 6LZhJ3VkRznC9bNH, 6QvVFqNYty8bzrcf, 6T3M1HRAxyntv9DL,
6YWm7PgvHN38MbTV, 6bKwWFZfdMTGtY8A, 6c8P4BLaVAj3bpK7, 6jqwbMpRFxTmHKkX, 6q1mWrtM7zXFckvp, 6rC1PNvmxZw94VkT,
6zQp1XFcjMTLyWkf, 71fBbnXa6DCz83qt, 79KcA2wXnNBvfCLV, 79Lft2mk8QbhHXa1, 7CdcmnMjg38b94w1, 7F3hKPycMYjXVJkb,
7JGFmtxVQMncYTX6, 7LAvbJhMP3zVktKr, 7MV9Qm2HKLcXBvD8, 7NMHDaR1AF6Lkg93, 7NrkBzJa1ndg3yhC, 7RjgZhB9mAbCHqXf,
7W2YXbm9xncyQtgh, 7XK1DyrJ2TfAQ48p, 7cpTkjaFB3mz6VgM, 7cpmjq6DTBNLwFhG, 7dw18XbYPRCyMah6, 7h2m8y3bH9dWMTrL,
7pHbXz3PCGBmqhxN, 7rzAygVjwH2C69Tv, 813rPaLtcdpHy67q, 81LcPyjXVbm2xGa3, 83L69YpwgzB7Nnqd, 84m1KthLRGZr7qf2,
87yxQDJMZfwgqrF3, 89TNYtj6DkQ4qnaW, 89yPg6z4AjZGfWRm, 8ATZQ31xabwdgf2y, 8Cx9vcnzpJwk34WA, 8F6HyxVmqarpBPNt,
8G9bLVWd2YjZR7vH, 8KPHjdAxZYFNLtgR, 8NmqaQRXdzMnvDxj, 8QFr9n4zRgd3TW1C, 8RcZLHvnaMxqB2m4, 8WQtqLj6wgfGmYny,
8ZH6PVXzrwWRFgap, 8hTHxr37tDfC2NJb, 8j1V9BACX46PTLhw, 8qACFaGbm1jn3PV9, 8qKTwcCyPgd4B3xZ, 8qZ4gNDQJRj3B1kd,
8qy3fA19PTdW764B, 8rVC2z4L6dhMWGKy, 8tBp2ZK7XT4QJCgD, 8w274CkRtPMfzKbh, 8zTXyQjWpY9vmKDL, 921Y3xP8RrmgQyWB,
941apA2h6kRL8DcW, 94TYgrcdVnJFKj62, 96mDRNG32MjxHtgv, 97C6PAqzp8Qd1GYV, 97FctvGNRgBYQaA3, 987q1pjnYdVQbCmT,
9AhxYCTgbLfc4zKj, 9ChvAYZBrGqzTfyQ, 9GqdQ2Z8KhwPj64B, 9H72ZkdBGfhCy6P3, 9LWQGFk4bYRMj63K, 9XQfBmZRaNWJkH3F,
9ZTCYrFcDxQ8pk7H, 9cQz2ymYMgfL7TwK, 9jzZXJVwp6PQ148v, 9nX7rz3h2bvPJBaH, 9vGFk4MCPmhZV6bq, 9wdCHt7QKvnrXDFP,
9zMva8xTVhFkHPAt, A4PL7VF6xBaq8zNX, ALpDvxc3GtT7Bnkg, ATMmnV2tC3ZPb6G8, ATwqPbhkavQCx2fz, AbT6hNF2WLcPHDtx,
AgkbML3Q4cza6Pty, Agvct8ybQW41MGYh, Aqjav1gfrDdXZwYW, Aqn1Fpt4BHCKmrGD, Ax6CMHq9gNcVfnzp, B3jCadXYF1NtGD6L,
B4wWqM2TbRmdQy7Y, B6JLrnWKF9htyjv2, BCZgcRxrh7dA4ayX, BFPtcmbyLr71Kj63, BHLcYgTkWpKQ8vRZ, BKLwZFjpybz984PN,
BTFR8jd7kqGVJhMc, BVLajhv3mxnC16Qb, BY1RXdPLCVqTH2gc, BagZHvPMqmzC2fTr, BbW6dKVvYDa4QR81, BbaPHAw13WFkMGd6,
BjAw1m39YGJzWvRa, BjHQAdmWaRwDpZk8, Bmw1jxaLbfzRF2WP, BqdMTtCvN3fVJpRk, BxKTm8wJy7a6GXhQ, ByP4fLgNnzj1RmVr,
C1cyPbL3RQvGjq8p, C2ZAbF1aLqp6YHvX, C4MK2R3QazYG1fWx, C4TGJvzHF3aZ2NjR, C78qHJYzLKFt1maf, CABa2QHNt79hdTGp,
CF7ztDwY8pJ9KZhn, CGQ6ydwxHknfFWKR, CJxvRDMwZdGL1VTF, CNWJgkfyAhjB3vRt, CTqjD7zAGpb1ncJW, CVXQm8RpjWA7xw3z,
CWhG6LyXk3PBHdMQ, CYP3GF16XbVhx72d, CaFxj4zXAV89N3Ht, CfHvwytKJ6P4q7ag, CgLJGDqVwrcyBQRN, Ck2vbA1pfqnKDgYV,
Ck6Nf7mtnpHWyKhB, CpH8ZT3MYqh64KPj, CvatJcKfDZ6HRkzw, Cx4JdcnX8LGDBhYK, CxTmQRX2ytGcwA9r, D8QTfLZMP4ybqnkF,
D9tmTWQndR2jvAX8, DBzk8qw4yFWtb9rd, DJjfFK27RhqabXAP, DJwWd79b4B16qVCZ, DL9nAWQwcN3rvZ4K, DMLvZrTNa1xV4AX6,
DNLMXbCRQ8zyaKVf, DW9YkdvyrjtMBQnf, DYkxp3BJygTCzMh7, Da8gfF6dtZn74MXH, DkwZn1rLVgKxv8Jd, Dmt7MBWFC89AXGPq,
Dq2F6jdPpvbTXrz3, DrnNChtLJb8pmAVd, DwK31WPFbAqm6TfM, DxCakYyQqgHdXzRb, DyYJRTv2XBfGcwC1, F4HWfxYQ7atgkJ81,
F6Ajv1TQNKhx2fmL, FAKHakfGvw1nd2qx, FAz9njrRfqxwgWb3, FCXJL6RHZjzQY8d7, FGmjgRrQ9vthAWa4, FHqYCNcKntvDAjxg,
FLK4Vqw3cp6XkRAT, FMzkb2KrnWCYt3dZ, FN8krKfCb3BXWmwV, FQxGr6HqMzY8AB3b, FVLgpBbYTwR724Dj, FXaN14xP7VhbDpQM,
FbAZrLMmQz4XnDa7, FbWLNP8V1QzxXjyC, FctKwmkXnbpJRDNx, FgRz2YhmkWnBLZcC, FnaGgrZxjRNzkmCX, Fq4fzXtnj8PQHmLx,
Fr8PM2jmKX1LbkYN, FrLvVqQ8Mm4AkczY, Fx94JmkLncTXBPQy, G3DyJv2aZ9rb78c6, G4wdA869ptvWVKf7, GD4zFdVBH9rjWLXq,
GDQbpM4RN7fhJ1AW, GH4yJ3FzP7NqgbpK, GL3ZWqVb41PQfrMt, GPvYxmw6drfkAFgN, GQAVdJ6fD3NKabHq, GTY6PHCt8XcQzmrb,
GXvD1Q3mygrk4P9b, GZ7xFyhzW6RraJP4, GZcmyA6CwHnPhqzK, Gp4kmc6rWTMV23v7, GrCYtPWaMmkxhnAg, GxtapmrPYNF3TjQ2,
H2GvZtVYNy4QzqLw, H4TRJDQpBPmVWF8k, H6TJpmhDzwvKPbCW, H8bqAB6ZcG4CKfWD, HGKLB9Yvhbxdr81X, HKwpDd1XQYt8GVc7,
HL38fnFjNXGTmRCa, HLQ9Dy2tdnAWqYXR, HMdLhWY3pNbnXPtF, HPqkV1mMaGT4BcgJ, HQND4t192ka6RfTc, HY327azxVvFywT49,
HbdV6NgxDq8wLa29, HbvWQBKhRn9AX7Z3, HfpGDhx3BL8bmYgX, HjTPVfFXBxzJZgAR, HkNB7A6tanYcGLWj, HkynYAjBCvN19rmh,
HmgdjFwJNMAD1VZh, Hp9ThMznR74PwYKb, Hpywch6rA37NWXYz, HrRCt6gwkx84GzpY, HtXh2PYD1VvwK84J, Hwz8pKxhm9DAFnR2,
HxNMb7JVw4pGP6dn, HxaZA2PVvW9TrRwG, J23GjB7aNAKgrLqM, J32CHbzcQpFhW9GR, J4Cbw23F6KQPTRpq, J4MmdxjphbY7DyBt,
J81YmK2Twh9ZkNfC, JBY9aKxgTdfbPQqk, JDNPdfVbtxTmH42r, JDvaFK3wWnHmq7RY, JG2H8wqLzXDChk9p, JG8DxcCHLj6T9qXk,
JRr2jLy8vNx147fX, JT3mkMDXFYy4QbNq, JX4afWrdcw8RkHDt, JaK9D1Rwr36ptvPy, JbQyngR8NWXKardc, JctLn6NBFx2rwGjv,
JdwC6TNyqDR8b1MF, JfLC9tr3HapNzd74, Jk9bRtfDCmrhwAv7, JqYpbaFxvLTWtPKN, Jt63rzjZpBaxRVd4, JvcpwWdV3XnyRzjB,
JxvWA68g3jTkFdQV, JzrcXpaGNCARv8DV, K6rZRbwYD1Wyjhpx, K8rhAtLjcz7gdHkn, KA3qfmHynCpJVhQP, KATdwWyRtng9vNCY,
KFqrLgCGV1hAfv39, KGgkJdDcQXyqpHa8, KP3HfrRdbCxgzV1F, KP68wChr1WAkvqbT, KT7J4XL63dQxw9NA, KV1bNq8ZfXHgk96F,
KYmwz2q1tG8yAcdP, Kd6zWnZ9PFwbjMT1, KghP8zXQmF6dCBcW, Kkjqba2YAt1nyTXp, Kn6h9jbJ14ayPAMZ, KrmBQ6VF93jyaxzX,
KtcyFvNj1znaB6w4, KyTkvcGwmH32nXxd, KyYwHQjxAL8tzvdJ, Kz4VDRGJkxgBfYc3, L1JFcYKG3wXnk68A, L2c7jBZNwFPpb1tV,
L7VBmjgyHRCzqYa2, L7fywV3Nt6zR4AHj, LGp4aAnbjr7NzyDv, LN2Vk4nMtBc8vWbT, LQDB9akd6rhc7bTz, LQG6jWCnAP1gwvdt,
LRvhYmZHFn96837g, LT2mcZ4CvjKMhQkJ, LVzGkm7dyQDChAWw, LW1pRgPwCmXBqk9a, LWZt9AV1qXdxaJv8, LZ1xNTqaFR9dAkhG,
Lbf2WTQqKwCRgv8j, LbzCGj1yFptQY2Za, LcvAbQBGHk6MwgWR, LhrZ4xAt2Hbwja1K, LjA2HFg3PTpZ84ND, LtdXcGwxCvDFQKh6,
LwDfTRCWc691Maxd, LyqG8nNCBz2hM3Kg, LzDT4PHN9rX8g7FA, M2RHxFDymYWLnpPh, M2qWf9ZyVKJxnzkY, M91YXwTBmQDgzk8Z,
M9GRNjTt2BvdbQDh, MBQ7WPCxf26mKgqn, MX8PGgBF9dNcQCwt, MXZry8VfkWHhvnYc, MXdJx7rWLQn2Av9P, MZwrNTyh7t1Amnz3,
McP6DGYbrFh7zkjT, MfDHhKNZjJwm7pB2, MjzWnNZP4KX2c6kt, Mr6kyq9nvf2YJPtw, MtWFJv1bPL26KTdx, Mvhj4Bf3CV7y6dKx,
MwRP9LF4QNZ63dhY, Mxd3G1mcXFv4RfN2, My8RkJ64w1cXzqP7, N1kVp6h4WczTgMvA, N2mcXLq81vbCMJVQ, N6cHzwkKGrAjqZfM,
N7FQbLWcD3AZJkC2, N9ryTjxAmKZ2pwGC, NArZXYhmbnpF4j89, NCBrVGJTFM3a9j8q, ND1hBcGk4dCYRym7, NDzHATKqmp63kfvZ,
NPkAjF7cTYW16RKw, NR1k2zLFqKt7GXAD, NXMZV9xjvWpB7GtP, NdHfZR3vG6qjVm4t, Nk4ZJaWT2KQ8rXbd, NkyKCr2XmgLt1YWv,
NqKayk8Q9rwWvjpz, NrVY2qvGn1gWC49z, NwqvTkmaAZdBY4Dg, Nz9wVAdTvpbJ7YyD, PCW1JNXVx2B8nDyh, PDT2yACNzd4LMZ98,
PDpraK4v1GMJL6nh, PL36WKJbqatz1RkX, PNMwJT6fqcmAgWQ3, PWx7LHKbCwrQTpRa, PXbhqkFHpt47RmDy, PYj7FqDGfXtZK9M1,
PZbm4HMKTLjJA3Vh, PcKH4Cqkf1ZTpjQb, PcVWH9XhGRCvb1Y3, Pd1H3kx4ctgnvXL7, Pd3LqCXcxMg92kA6, PgFZhqm1zc3f6Y87,
PhFTQdyDxL4NWVX9, PhrCRMjvJgxX7aHm, Pjd9MWAXgCYTcn7Z, PmM8nxa1p6QkhVLC, PqK7NWmpYF1zA4fT, PqpnFHfRcTb3W9zw,
Pqv8wGpFhamdWrLK, PxZaBqXzNfgK4Vwp, PyzXR2YDdwMN6trV, Q6xgnXthqbyc8GCk, QAmXKCPV4apRd1jT, QHXdBJRj2GPpC3n1,
QJwXLaVWzx9pMqky, QMmKCT3AkR9GNhLy, QN8pLwHmn19zCPkA, QNDRXVY7zmW8B9TZ, QNwkdxbMvWmrqjAY, QYhr8G3q96dVXAgR,
QcRPFpWB6mg9YVC2, QdxK9bTDZkq8nCHz, Qf2RhFMXYDwBCzHg, QhZrG836p1mcJfqK, QjHW1hvfbNGgcLBd, Qqf8ZC1vRYhwG2cr,
QtG89p6HRX31MdcF, QvNztyTfHcxPmDY6, QwYrJ4VPjyBkRbft, QyHRbdM3YzXvDwL9, Qz9mWNhj1nBaHbPq, R7pQ93ZP8b6jvJx2,
RBb9qfdKFLM1TP7J, RJ82yLB7HcYrd39M, RKQta7wZBJbv4fCY, RW1BNk6Ybyh3XjLQ, RYXp7xNFPdyvkqMG, Rcanq4T9FhDg1JMy,
Rhx2TCj8wQN1XFkZ, RjYCG692VvtW34Ah, RmjCkTnyZQ8pg9B1, RnCZhqbjcGymHVfd, RnYmQGbdzpfLyvVD, Rry6WVb2NGD1dXMx,
Rvq3r1WmdMXa2yGz, RyrkQYbMVXBFv23J, T1jRWQ3KpxfZqLkr, T2rBvFtkWncADzh7, T6A9DCFKb2k1Mvar, TADvnpKkdwgZGJ6b,
TB8FMm9QKW2Htb6D, TDCx9BzjPGdJ8KF1, TF87cwnBHgLjVx3K, THjbK9MkY3qm8hDa, TNH4jt6mW72kKMhJ, TNW9rHBt3MD1X8Ag,
TV87BvKgRWGrPtx1, TXyBh2gwbF6QfrnH, TatW1Q7Mf2Gg6hbJ, TbB8q4RMK3LmYJNv, TbF89yn1fHYzt7Nd, Tfp1v6KhDwJntrRd,
ThKD4jr8MfawcCPv, TmtKkcp8rPgH9wLa, TqpcnkwhL4bmyNtG, TrYmWQ6vq3hLyNCk, TrjYQAgLkPz689fa, TzmA9xNqY17DrhZX,
V1wJ3t6YjPKMfDkX, V24Fy1MphnRfCvkT, V67QDxtk9jNvr4Wq, V7CnLyfRpbxXc2WZ, V7kcRxY19FAHgNyX, VDh43nQq71MjKwzX,
VGhmxKWqDCQfLn8d, VJZ4Q2KLGNqTx3PY, VJfPj3WbCzvLmQnA, VK83G7MdBhLRFAfa, VLxKYgWdn39t7Mfp, VMYQnKfh96XxR7Pt,
VQA1tZF3HKmBvDpa, VZjPQt6f9CNBWKJb, VfZY7kQLwTpPXgKa, VjrAGFkTvNy2W98B, VmakY2hNdyWxDrfp, VrPg8T3fdK2CF6n7,
VtkPTAqxCv1X8cJ3, VvGWqMdpwRFAbK1C, VvNyXpGtkjr7xYLH, VvRNkLTq867HMwn3, VwFrK3vJGLBRMqjQ, VwyW3ZP1H6RTpdmf,
VyQ3pt7Hga1XbMjL, W2ajmp7yNbK9vZPd, W3R6PLFBhJ8bqKkc, W4dkXTZH3QjR6hpK, WAVGP9BrQ71mnRK3, WAfqvJMK79R1kPYC,
WAhTxXPLaf3J2tbk, WBV7yLdw2RpX1vH6, WBhbtfCQrLTV6gGv, WCV3q8DgXfGQLBtT, WDQR3PJr2XZmp9Yb, WFBMqdAVDmJYy2Q9,
WHwZADRCP2BzTXQ8, WM1kTRnXVr7hBGKL, WMy9axzfg3jKLhYc, WR46nD1hr9PfkdB7, WRrMmXgJDdhyFcjQ, WV34XKG7MTQya1qj,
WZrKkaV4xPq8CDgR, WaHMR9NKLrwbzptn, WayYc9LQVPKmwAX1, Wb8w1TtxA3kLF9M4, WfKv2TphtRrzNA3d, Wfn9Lg6TGJPcaxK2,
WhPpmnzB6rR3gD4Z, WwKh8C1FZLcGTXfN, WyJ3AmPVpKRqnbhv, X1Jdx3bjtvpzVTDH, X6Hp4gDwv7zftxMC, XCJdxAF17zqHMhLY,
XTfyZvKL9V1wCApr, XVMgaRd2b9DtFrnQ, Xa6zFyNWmYdp3BZL, XaF296wZvArq7Qxh, XaF63gxRH2v1TVBw, XaQM4dqznRfcbNW6,
XfQ8GLDCFN6KyTvr, XgGTyamxrWvtQ62w, XgQYbwGBZcHLKRxV, XhGkm2F8rNDv4Qjq, Xhz8p6yB7bYqFaJ2, Xj23FkTPgrHawv7J,
XjxMRyCGdZNgT1vA, Xp8xn7GTrzNfDhyV, XpnDCMx9BtbWa1FG, Xvczhd3HytKV92TM, Xy3tw1gBKbmqZjn9, Xy4NrnpBH3Pbh2fL,
XznCM8awybq7xVr3, XzxyRLg8QW9rqPGn, Y1tBfF2c9yvRZWxN, Y9DcKpamAtxrqWfL, YAVgyZp9FNMKCh4D, YDWpVfZ6FnkabP27,
YJPVatA7TqpZjb6B, YJQNvFf9jdH3tMwB, YKDdbpwaHxm7cy81, YM2CzVmPnHQFjfJ4, YPmbL2Ggfn1648HR, Ya2HzxnM4kFBR8CN,
YakBLWFpjnrdhGb3, Yb7Mx8y193GPf2TK, YbwrZWhBXqAz9vyM, YcyawXvG9pqLDbTr, YgPVzX4c6K8t9xBT, Yj2Nvgp3fyhJRx1Z,
YjA14kP6FyQwLnNJ, YpjWF4A28y7VCTJB, YqLGhxAbFkDPjW29, Yt8daxRbj2NDQTVv, Yv1Vp8KLkgMfWTRB, YvBtH963jKD17fbp,
YvfTNp61Q3kc9Z4P, Z1yTBW8Jc7Dmt2Pd, Z4w7HKFPvcrTWd3M, Z9y8b6WCft3R1nzh, ZH7zNkqT9j3G1Qm4, ZJhb6FjaDK1tq2AP,
ZLpcV6krn1wYHJvt, ZN42mMgq8XH1kyC6, ZNqmbgBYpVAadyR2, ZP7ykxcTmLzag8H9, ZQJXxdPnm8fgY7Fj, ZbxJW4N6kQPDryTh,
ZdnJ42Tghazm7PvM, ZfNjknQaBvp4d6Kc, ZgG4mCFrp16PwXcq, ZmKVcyYFzfbANq7t, ZmanHqXpQJFYCKRW, ZngYVCKWb21arNtA,
Zpc21GrAjRQ3yfvg, Zpmdqhz96YNrBXCH, Zy1rP29Akw3N8Y7x, a4NDdCJPcbt8AxZp, a4XChvyGFrQAj6JY, a6F17d9mnHTBDKLg,
a7DdK41PQCgj9NAB, a7MJrGxXPzNCwDKf, aFq1fMXx8mZ6H7pQ, aL2ZJnf9pCgv6T1V, aQAhMLzBbWJcdq6G, aRFzTnCLAtqHNdB8,
aVjx6ct4kMGPQRn9, aZ1vKgphzBD9wPNn, aZnTQdtRhq9mwgYf, aZyKwYWdNFJckp4g, abLhQ43ctrMpm1Zd, acw6RHFChKXdk1QP,
ahzgLtZXDcV3GfNp, ajykLz8m72pGqxtH, apMdm4GJH32L1DVC, apNB26JmYPyhFWjH, atfzcr9h7Rbv1pPH, ayCkfvn23mjdF8NM,
b2cX8fQ9GmMYkLwy, b3X8aj2Hxw7FAdcR, b461jPtVcmwyGAqN, b4hzxLyCZXMHW8gP, b6ParKH2ZChLtzmk, bF8hVwJNDT6HyRWt,
bGVJm7RWBA4PQZvH, bHaxRQvC7VATPkX2, bKJg73VyQGqRT2hW, bTkcxyhmp4rzjBFP, bdhPYFLvzH6R8KXW, bnfaTWMHFJYk89Nr,
bpyVv8MwgHarP47t, bt3aNZQh8WdDnrw7, bt9RLg4BpGVmaM7y, btT1QcWvZdpPyr8L, bw2r3HfhqLJjmKzn, bw9x4f7y8BzFpD6H,
c4rbnMYyq9NhXAC3, c69WN31xbHBPtmhv, c6qnPXjxAvayHVK4, c9pH1QPFxA3mGtvd, cAGW6Farb9N3d2t1, cBryQR7ahfm6J8XF,
cCFkfKm1dvtZgq3a, cFGA2fy4gajBWNz6, cFRpkf1jWTaHrgL3, cHLtXwmpgQaVRxzd, cHMQhVq76BR9YFNm, cR6QhKF4dJA3qTDP,
cRtHw9DWZMrKnyvQ, cVkHBCx9Azd7pWmh, cZA4rhPNaqR72X39, cayZ7mtTxKkCvnpB, cdFkXVKH1xm2JhpW, ch4aB97GjfrM1nNP,
ct2dxzR6hTFBLwN1, cwtKD6v8xmJG3jNY, cxRCKwMzJ3d8t42L, cxgRYMkV9vZQaPWJ, czFAk3dtDYqgapVK, d8cmnvYHRjNPhMag,
d8jmg37ZWt2FC6TJ, d9Da1tbqJCyZzR8p, dFnAvhrDN3TVYgJ4, dRKLkaMAZ2rYCfhH, dXWmHA9PDYcqwg3R, dZ6CjB3hqarMzvbT,
dZCGtrLpBaT7XMvj, dZJ72G8HPnYANjqz, dbtKT6C7wD2vn8Bg, dcZKF2bPJta1mB4R, dmpB3F8wKaZ7Lqr9, dnktvpxYab6jJGXF,
dpXxrVfmk628TNHC, dzXT7JngpxRPwG32, dzYa4wnWj16cMDKA, f1DQ4aZczd3VxCkG, f3hyv2GtFW4b8wJV, f7JajNKbQYLRZHpW,
fAvjB2K4LDGmYQHR, fFJDzZ3XGA97yctj, fFd4BgHwTaWXqJkp, fRZG9Ww6rFQcA48p, fT2knhqYQ8wGNVgC, fTn3gkRLv8tXcbyA,
fWRgvP17z8yrL36h, fZbNXK4Cwax6q2Jc, fhn291rVvFtyAjbY, fjGwyNrB1npQWCH8, fjPTxNBrbkJ6v3Q1, fmcL7MQK1dFZbzwj,
fnLwg8xBhRDrCKGX, fnmrtPV18QYDTgq2, ftjDWaN8vT1XywZF, fywmzq86PRtjDJra, fz2HQxXk96ywVYv7, g1qJwva4Ynf2RFVL,
g2v7mTHfctMzhN39, g3zGbn4VD2fPycNv, g41xhtdFCXZ8Vr9z, g4Y2frPx8kBmj9Tq, gG4rMYZRD2XChQxc, gHXxvFTPC9na7fwc,
gKTQPr1X7MRFNfnD, gKdFBvXjnTfDLHzx, gRmxdjCHt6XDMBpa, gT9MP3BnwFjy7CdH, gXb8P6zWrCBmYQfF, gXr8d3nMz7avtBkZ,
gdMkYRKzyLXmnNp1, gfQkaTK4pyh3mD7X, gh1AFZdcqL4MPbQf, ghWq2C84BJj9bQHd, gj374D8h2zy19XMR, gmKvNqHFJLZXfdnP,
gqktKhpVRrDM4YxN, grvZjX6Qf3WdRMDk, gywYLPFQCJa9XfVG, h1Yc3j9w47qydCWg, h1b7cwtmxLTgZYFa, h1jcW2G7nAHYKqtF,
h76PDfjpFtVdaLgH, h76WdXRmxBFzLPT9, hCfYgdqVk7QDRZGN, hDJ6cbM4Hrz9Fwg7, hH1JxnPAtYXcQ2VZ, hMdWYxyPG3rn9VQ7,
hWJLtqnXbzwG1TMD, hWJNkrnmGvxL7qFY, hbBVRkQWKjwCH2nT, hwxjbW2XDv9BTP1Q, j1xDTH4N7ncAyJhb, j7xyJdLXHZNA6PQg,
jDa7VL9x2y3bNr8H, jH2RN9AaYWfGqBh6, jHPB1gZAkVbJrMQp, jJGnAQ7HyNVb1hzD, jJM7dvXRTLbya4Gr, jJYHChwVBRaDxNWk,
jMbQaJ8mXkqPDGrF, jNt1CqAKcwyZT3Fd, jVhAWnaHJgf3FPmy, jbhNzZaF1G4K7My8, jc2NQhKdaFwfJgx7, jc2kX6xJaVP3dB8r,
jhC1JPnvGApaZ942, jkLQ3Hhr9VxTM28p, jnLPyx8RFVX4WMN6, jrMPkv9J6nLY1WyX, jxLmMtP4DyYwAZdN, jyD2wdc6FgZfHJqh,
k26rdYnGDNZ3b81P, k28NjXqp96v1HYAc, k6PQqVgWJG249ZCb, k8mdVabt2CFphzJ7, kHZgMqWQCryAFcJt, kKv8aLDtjm3w4VyB,
kLydR7JPD8FN1aBz, kRM7acVvBqYgF9QW, kT47rwPxW6a2cDCV, kVY4TxntBWArNjyK, kYBMpfAZHabtw9z3, kdbvhDfNTcjZ2nrC,
kgQpmaCcAhZYb39L, kjGcWvaVNPYwM6yH, kjdrmaFLb2G8CAyx, km2ZtqdTFhDK4zYN, krjTFYRwZgcXxzGB, ktM3G9KT6yFCQ7Vx,
kvQ9DtmnTGpZzgJ6, kw8a9C23KrDRj6XP, kz9YWBCb4vr6x7jc, m1FkByqLYRhX4JaA, m1w2JQy4LDMVf6kB, m3Fjgrt124A7advV,
m8RLWaY7bh9jHKwD, m8gPaLCQycNvRkw3, mApW49VaFD2TKyg1, mC3qj4FNn7PtMyVv, mCNqy8LgJb6kXQWv, mDMrkYnQfd9CjhKw,
mJCw2dKNnQh7Tvrj, mM94FDqRdAfGpbBJ, mMCYac1DZLA9hVqN, mNBAdW4ajKHzhMYZ, mRJLzb4c8pKwG1Yx, mRzw46vjLTHnXQPx,
mVNgCkjB3ZxdGRrT, mYcDW3awJ2NnXBzv, majCBMFN4vpYJrnR, mbdtyQcGXALMnN86, mbgjLf9CYp47RVrc, mbwzgt3VpLXTj4MZ,
mhZ8WvgabzLwD9FK, mp3A6kb7j9rRhcZt, mqn3PdpaZkQMW7j9, mqzgv4rPVxA3JHGf, mrC9TwxYgb8H4F1d, mxbK4d1k96GzwVRh,
myBT4jMY7NJPnxfw, n7Kfm6gVjW9zdQrt, nAVx6ZgkQywNd3hc, nCDZQNWKX3twHhyY, nGctFJyB8jwX9WHL, nJqcvxFVj83NTrKR,
nL1T7ab9CXHAc8gd, nRMrVQwGxLp1ZKDk, nYch4R1Htgq2TMJa, nhYyv7NG3L982RaB, ntBYv9gkdcZR7HKL, ntmraghfPTkdbqF6,
nwh3zdqc4CjXa6MG, p34dnvkg7yDKL6cr, p4kGNXhxYPVcyH1M, pByatRLwdAmZWX86, pCrzdDcxXPF3BYgW, pFkqWf7xBAtZvTby,
pGQn8NhC6y1gdt4T, pGbh4FPcRqNJnYkC, pKJj1NyhWd4L3qMD, pLtJjx8DCPbHRacv, pPaVZhw4B6tMkNJv, pRaJG3cxCWYdKb29,
pYJ4BVZFr82LWcMy, paQw27rRXzhjCkf3, pawN2JjhCM3xKbFk, pcjBAHFzfv4ba79J, pcxf4HNtgVKq3F7Z, pdLvYMqBAt9Dy73b,
pdNRGK6j9rkbyaV4, pqfK2AgHZ7DnhvPF, prQ8j6B4ZRdyM3LH, pyhWaACkcGrFK61w, q3MhapfdVyrY9cm4, q68xzRHYhCpFLjfJ,
q7LayvYVQKwRMptf, qA3vDhTjrQVGdzMb, qCZwWn2pJ3zVGKkj, qCtD2yf7PpWYbcL1, qDp9Zn7HgR2axLVB, qLNmW7nftaD9M8pr,
qMPwcLCNa4f1HVzk, qMhB6GPgxwf1tpJk, qMr8PZfBpHWyL6tm, qR2rnzfgPakD879W, qRpAgN7x3ncLvJbW, qTHXf9y6ag1CMtdp,
qVyHkJKZ4jQdpAcX, qYPRLWv6g2D7FTxB, qbrGgBZfzLHFW12k, qgfYWFzjGtmJTpAX, qm1Bvy29hGK6bMaQ, qnyc1zwaB4FrvM9x,
qv8daLmJcwDp6MrT, qxCcfXymntY6awZk, qzZb6wcxPXnHAyV3, r1pJHqkXPwZbzhMc, r3WAfMaKHPqp9T8k, r4fBkAMgmdxvqYCD,
r7cKqA9xRHbwVPX1, rAJbdCG47n2hgwVB, rAfyZTNX92CJtqPh, rF1y7X8NKxcVYpRm, rF7XRn8KJxVj1fZL, rHM6TPA7NVckxQmh,
rJY7McXdyNvAPTaz, rKDxZb3MtTYWpAN7, rM7Qmk8w6NKhBZLg, rMPpXzxRN1L7mQbT, rPv42N9YkMmawXGD, rRpTGc98tnkZ7xHg,
rTLVqFpfa9GJQgHh, rY72FxaTNgBQZHXp, rcKYRF3zDbJna4M7, rdxLX8Z4aj6cHNvp, rfBaQ1wx82XDh7m9, rhxZQvXwm8Dz9T2A,
rqjzTHacGgmF27Xx, rtF9Qk4R1hLvTAfH, rtpQk3BXbDnNGCj6, ryMXYWxhD2FQC7Pv, t2yjAqMgTvGZHJLC, t6YVTRBMfZGhn9Nk,
t7mhQd8PjgfLWyJn, t9fw1D67gc8MqrjZ, tA1Qd6nqT789XBHP, tBH7kDj1249RnYMh, tXTpMhL8BjHG4ZqA, tZWfnFBNC8rjALad,
tdYLMmfgcAJCTWyr, tdk17mp6hAb2DZFg, tfAVyDRHYrz1QZdW, tgd6xBMhfQvDq4K2, thvrYzLKGB97kQZC, tk86znZXxMCdwBKb,
tpWza62JNFrTm9hH, tpr7PfzJ2FAnMLaQ, tqTHmcRfPW3NnZxr, twqWdhPm1Zxj3TKB, txFPzrDpvwndHg7k, tyFZPdH39kGq21An,
tzRdk7fHGKpx61vJ, tzr3fNYbwyXcK1PQ, v16yHBWaJZTFfhGn, v1TyzhDYNPXgHqmn, v1dGWTxHFVmk2yhj, v8DRwt2gJBXGxcVN,
v8J2CFhH7xkpbRnr, v8thQxgGN7ZFLXCW, v9pgRdjVcLA7mYt6, vGtQdg49pCh7zx2H, vLR2FkctCqPGKx9p, vMdAbxpnTcFYfXZ9,
vRwhV86Qd1Ytrpk7, vVjtBGNM8LAd7Yh1, vYrVnXqA8CDkGQ6N, vfcxgRqJMHkPndwV, vfdVpNJ974QG82KT, vk9HJZwrfhmCayMY,
vn2Z3JHhFGjVKm1C, vnKj7bWLgwJqm3dP, vnj8MpcGHydhY9Nt, vpTQqDkgNYLnjtZr, vqZx76MDQLbwpg3m, vwYj6cK8Xt24QLaN,
vzARLnH8MkYZGTCq, w2AV6grXdkKN7c1q, w71ZM2adFbzqK8hy, w7rBfF932pzqYgmd, w9AVh31an2CZMgXk, w9kyVqQNDvrC36Tg,
wCWnYxfbPXDQ9zmt, wKNnB9JDtv3RPVxW, wLcx8mpnGJrj7Xha, wNH9k6Fyg2pQG73d, wNqHMTvYmCQKXV3W, wVgqL9tv6PQCjkfM,
wYcynQA134JNBL8D, wYq2dWPNZkvG3hCD, waFnCzt2chBHVGQW, wdJqg9Z6nRaYC8pT, wfD2zmjVYB61p9MT, wmCW7JvhcP6taxjT,
wmyLtgPWKBHjfX69, wpcC7xvm3AXnrjbT, wpk6dDzAPLGJmR3C, wvDFQB4cjG1dLyTh, wxPA81bJ3Qz6fXWG, x7Fjy3CbrKw4tNnR,
x984bRXmVKnvdtjZ, x9byqam6WYcKFz2v, xDbJdz2KatXqfNpM, xFVhPAdXD3NYTGCk, xHwJaMnyq31hRACF, xMBa9cP1J6r2w8vn,
xMmX4ZFRbrzAChwV, xNGpFmdAnXt4g7qk, xQ2J8AbFN19pPD6X, xRK37aCvtVfj26kL, xRKkNnpG3vcTQFfA, xT7XM9qKZaJ3mpFf,
xWpq74Z8Nf63TK1a, xZNMrhpJ87wd49qQ, xdvVqWCwkcp4KNbG, xg1L67dtJKmQAqNv, xnf4K91LT7A2mNaw, xqQNadDY9JRwG1hL,
y2G76gPNrZtV9fkC, y2brwALVgPDZQJaf, y4tAZXkGTN36hmKW, y6R47pr9XC3Fdzkx, y6TdDjfqm9HwKbF3, y6ajCGVKfpJqT2LX,
y91jrfQgt3hJWvzX, yApNQaB4xhrWKGvF, yCphrBFqdjgb8Nm9, yHvTtXp4R1h9q3Wr, yLPAnRjh16BpmQHf, yMJcVHp2LQtkaRY8,
yWaH1hR43wd8vLpn, ycLkF19mTpJYjtQg, yd8vFWNnTtH6xL9K, yh9wP43x67LmWAqT, yhVbHg7vzDkKfnRY, ymkqDYNTxPjCgwnb,
yp9RdbJW2CkZcP1N, yprW782kT4GQPKgX, yq7RkH6AjBZ3tPWN, yrN2TKtxzjRcpbfh, ywCbWqZGaxX9Hr14, ywjgtx6H9fvrcC87,
z2QFnM7tAjJmqgvw, zDxgMqR83J2dXkWF, zH3TmcCQRb4Pa8tk, zKYrV6cRm3N4bk7L, zLacx86Thr3AMtqn, zLk3npv9hyCQxXZ1,
zM2hwbVmgyJHRPFt, zNmFC3BVvQLy4nX2, zRm2jyKBq67DdAXH, zThfNY7CVKqr1MFd, zXFrRAP2jamNZ7WY, zXkv9TNWZJayPtQA,
zZtqgX9JvMnT3fKB, zb3FKqwkCaRmvAh8, zc1hygaCXjDAPxL6, zctYVp8xTBkFy4hq, zhKHmYj2y6fvNFVW, zhNBgjHxQX7anZTf,
zk9DnPBZqyXArLca, zmgHQXYjpcnC8MrK, zqg6AMnWhRNDKPBf, ztpB7m1kbLQ3KdqR
