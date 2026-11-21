# EV-0119: Weapon proc rates from isolated hits in end-game and levelling dungeon logs

- Claim: Stinging Viper rolls 1.75 procs per minute, Toxic Revenger 1.8, Meteor Shard 1.55, Bite of Serra'kis 1.45,
  White Obsidian Wand 3.75 (on wand shots), Ravager 1, The Black Knight 0.6 and Heartseeking Crossbow 2.7. A melee-type
  hit cast directly or flagged
  Not a Proc rolls an on-hit weapon proc, Windfury Totem
  strikes included; a ranged-type hit never does, and a Judgement of Command rolls it twice.
  Ironfoe's chance is 3% against other targets and 6% against Orcs.
- Sources: 709 public reports on forever.warcraftlogs.com (zone 3000 listing, 2026-11-14 17:49 to 2026-11-15 07:00
  UTC) that list one of the procs; the codes behind each rate are listed under the rates table.
- Game build: 1.60.1.70938. The hotfix pushes inside the window (2026-11-14 19:52 to 2026-11-15 06:22 UTC) add 29
  items and change housing and profession spells only, none of the items or proc spells measured here; whether the
  players relogged is not recorded.
- Date: 2026-11-14 and 2026-11-15
- Pull request: #1032

## Method

A wearer is a player whose combatant info lists the item in the worn fights; runs that two players logged count once.
A proc is one event of the proc spell's main event (its damage, its aura application or refresh, or its extra
attack); same-target events of one millisecond from two hits count as two, an area proc's other targets as one.

Each rate comes from isolated hits: a landed hit with no other landed hit of the wearer within the proc's logging
window (50 to 400 ms, from the gap between a proc and the hit before it), so each proc has one hit that can have
rolled it. Only wearers with the item in the main hand (or the ranged slot) and no off-hand weapon count, by class:
dual wielders log both hands' swings under one ability, and Druid forms swing the paw (feral wearers are left out).
Casts that land without a damage event (Sunder Armor, Rend) count as hits; Shield Slam, Shield Bash, Shank and
Thunder Clap roll no weapon proc and are left out. PPM = chance × 60 / the weapon's speed. Haste leaves the per-hit
chance alone, since the rate is priced on the weapon's base speed.

Each rate is the pooled rate of its table, all its rows, rounded to the nearest 0.05 step. Ravager's is an owner
ruling: it stays at 1 (pooled 0.94 [0.89, 1.00], its cast hits alone near 1.02).

## Which hits roll

| Hit                                            | Not a Proc | Defense type | Rolls      | Count                                                                                                                                                                                        |
| ---------------------------------------------- | ---------- | ------------ | ---------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Windfury Totem strike (8516, 10608, 10610)     | yes        | melee        | yes        | swings followed by a strike, which logs about 33 ms after its hit: procs at a white swing's chance                                                                                           |
| Judgement of Command (20467, 20963–20966)      | yes        | melee        | yes, twice | Bigger Rock: 17.3% at least one proc, 0.95% two (22 / 2,314) against two rolls at 9.1%; Stoneslayer: 23.5%, 1.76% (8 / 455) against two rolls at 12.5%, the chance per roll that gives 23.5% |
| Judgement of Fury, Judgement of Righteousness  | yes        | ranged       | no         | 0 / 1,242; 0 / 736                                                                                                                                                                           |
| any ranged-type hit of a melee weapon's wearer | –          | ranged       | no         | 0 in every item                                                                                                                                                                              |
| Druid Cat Form specials                        | –          | melee        | no         | Bigger Rock 0 / 16,655                                                                                                                                                                       |

## Rates

| Item                  | Class   | Hits   | Procs / isolated hits | Chance | PPM [95%]         | Wearers |
| --------------------- | ------- | ------ | --------------------- | ------ | ----------------- | ------- |
| Stinging Viper        | Paladin | cast   | 78 / 1073             | 7.3%   | 1.56 [1.26, 1.92] | 10      |
| Stinging Viper        | Paladin | white  | 71 / 974              | 7.3%   | 1.56 [1.25, 1.95] | 11      |
| Stinging Viper        | Shaman  | cast   | 3 / 39                | 7.7%   | 1.65 [0.57, 4.35] | 1       |
| Stinging Viper        | Shaman  | white  | 53 / 719              | 7.4%   | 1.58 [1.22, 2.04] | 3       |
| Stinging Viper        | Warrior | cast   | 483 / 5934            | 8.1%   | 1.74 [1.60, 1.90] | 16      |
| Stinging Viper        | Warrior | white  | 640 / 7556            | 8.5%   | 1.82 [1.68, 1.95] | 16      |
| Stinging Viper        | all     | pooled | 1328 / 16295          | 8.15%  | 1.75 [1.66, 1.84] | –       |
| Toxic Revenger        | Warrior | cast   | 214 / 3842            | 5.6%   | 1.76 [1.54, 2.00] | 7       |
| Toxic Revenger        | Warrior | white  | 387 / 6765            | 5.7%   | 1.81 [1.64, 1.99] | 7       |
| Toxic Revenger        | all     | pooled | 601 / 10607           | 5.67%  | 1.79 [1.66, 1.93] | 7       |
| Meteor Shard          | Warrior | cast   | 73 / 1408             | 5.2%   | 1.73 [1.38, 2.16] | 4       |
| Meteor Shard          | Warrior | white  | 119 / 2744            | 4.3%   | 1.45 [1.21, 1.72] | 4       |
| Meteor Shard          | all     | pooled | 192 / 4152            | 4.62%  | 1.54 [1.34, 1.77] | 4       |
| Bite of Serra'kis     | Warrior | cast   | 28 / 869              | 3.2%   | 1.49 [1.03, 2.13] | 4       |
| Bite of Serra'kis     | Warrior | white  | 71 / 2232             | 3.2%   | 1.47 [1.17, 1.84] | 4       |
| Bite of Serra'kis     | all     | pooled | 99 / 3101             | 3.19%  | 1.47 [1.21, 1.79] | 4       |
| White Obsidian Wand   | Mage    | rng    | 63 / 531              | 11.9%  | 4.19 [3.31, 5.26] | 11      |
| White Obsidian Wand   | Priest  | rng    | 516 / 4768            | 10.8%  | 3.82 [3.52, 4.14] | 18      |
| White Obsidian Wand   | Warlock | rng    | 89 / 1030             | 8.6%   | 3.05 [2.50, 3.71] | 8       |
| White Obsidian Wand   | all     | pooled | 668 / 6329            | 10.55% | 3.73 [3.47, 4.00] | 37      |
| Ravager               | Hunter  | cast   | 27 / 375              | 7.2%   | 1.23 [0.86, 1.76] | 2       |
| Ravager               | Hunter  | white  | 10 / 242              | 4.1%   | 0.71 [0.39, 1.28] | 2       |
| Ravager               | Paladin | cast   | 185 / 3203            | 5.8%   | 0.99 [0.86, 1.14] | 44      |
| Ravager               | Paladin | white  | 356 / 6589            | 5.4%   | 0.93 [0.84, 1.02] | 44      |
| Ravager               | Shaman  | cast   | 71 / 1315             | 5.4%   | 0.93 [0.74, 1.16] | 21      |
| Ravager               | Shaman  | white  | 197 / 3880            | 5.1%   | 0.87 [0.76, 1.00] | 21      |
| Ravager               | Warrior | cast   | 101 / 1563            | 6.5%   | 1.11 [0.92, 1.34] | 10      |
| Ravager               | Warrior | white  | 140 / 2646            | 5.3%   | 0.91 [0.77, 1.06] | 10      |
| Ravager               | all     | pooled | 1087 / 19813          | 5.49%  | 0.94 [0.89, 1.00] | 77      |
| The Black Knight      | Paladin | cast   | 4 / 186               | 2.2%   | 0.68 [0.27, 1.70] | 2       |
| The Black Knight      | Paladin | white  | 1 / 152               | 0.7%   | 0.21 [0.04, 1.15] | 2       |
| The Black Knight      | Warrior | cast   | 12 / 665              | 1.8%   | 0.57 [0.33, 0.99] | 2       |
| The Black Knight      | Warrior | white  | 30 / 1487             | 2.0%   | 0.64 [0.45, 0.90] | 2       |
| The Black Knight      | all     | pooled | 47 / 2490             | 1.89%  | 0.60 [0.45, 0.79] | 4       |
| Heartseeking Crossbow | Hunter  | rng    | 105 / 747             | 14.1%  | 2.72 [2.27, 3.24] | 4       |
| Heartseeking Crossbow | all     | pooled | 105 / 747             | 14.1%  | 2.72 [2.27, 3.24] | 4       |

Reports, every fight the item was worn in:

- Stinging Viper (30): 3hNwr9K71bz4xjCc, 3k7rhyjFVNK4B1x8, 6Bm38KktgLMYT4xp, 6FDgBtVqvNPYwaWX, 6x2BFbkYgTPWN8XK, 7bch4MKwzf9HdT3v, Ag1GaNKfMLVRJnjk, Cqkfvy2Lrdtc47zJ, KAgmwHzyvtD8TXVc, MB1kQD3t2nHWC97A, MPx39BYZg1kW4qvj, N178cBjPtChmbk4w, PpLZ7FAgkvTwc1Cr, QGLzd2PDcm763VRr, RjPtJkznL4grBGqv, W6GLfFa4NCbXqBJh, XAD1zkMGQFgtjPTy, XWpMCYA1q97ckaGV, Y3PfRbtk7FC6HJ4y, Y6bcWnJv9kFPyXmG, ZHK32LtTCmqR4Pxf, aXrMN8hPypRZB6Tg, bAzcMXt1HnDNQG8y, fFYzDvMwta1LBA4X, fMG93WJngbRQ8vNy, m8CHh2RAXrGtMQK9, mCnfLqgbdxhJZ8aH, pbcRVhYPg36K4XLv, tDqCBf2zgnXWpmKJ, yDbnf2Y6MQWm7KNp
- Toxic Revenger (7): G9PLgJ8rWcCnkmKF, KXCzdLrATV4Gfqtw, Rxd8h1LgQp2FTB6j, T87bpgKckHzNDAfY, VJwzgFmrtN8dRQGf, dPgwL6an3HtBkWGc, vMn4tpwVxA3Q71dL
- Meteor Shard (5): 8PDGYhyA1rkCMv7V, V39PRKTHt8kahfxX, Ymz4XLKyaWqV81vw, nkdtrxBfh4QRLz2g, xb3QNfWvDdy7hTkp
- Bite of Serra'kis (4): 94Y7DJ2wCPRxXBqM, kTVX6NG72tgAM4Ch, mxLYatfbnP71kzqR, vTMxmyJYbZ6dVkf4
- White Obsidian Wand (40): 17RQWNTLJAgGBhrx, 3WyYMqr8ZxQRN1vD, 3zTJ2VXvQ8GCyFbc, 86rq1jAvcTBZ9xGV, AafdTnj4h7CY6xcp, BqdMTtCvN3fVJpRk, C2qtXKpD91F76bBd, J97GvmrRXCFZfQ8a, JG9PjfzvTYADhqx1, JKtLp9FHxZVhgG1k, JaK9D1Rwr36ptvPy, KG62VXF3a97CxtAg, KrWzgwGRN7Y9P4kx, L7VBmjgyHRCzqYa2, LxdYWR47ZFbTNBhM, PAFTBzCh3pcD72Wx, PCW1JNXVx2B8nDyh, PWx7LHKbCwrQTpRa, V7CnLyfRpbxXc2WZ, Wr7FcVnXCHdA2w34, XvjpYan3qxF9WRgz, Y4qpRvTXwaNZx21f, YAVgyZp9FNMKCh4D, Za21hDrqWXn6Rz3M, bmJTGwL1djRDn9v3, dLXJ21r9jzAmtgf6, fLyrVbpZAWxzQFYk, gL9RXr7YqywJhN1T, gwdtc8fFVQaKGjq1, hV2pjPWqr1MbYBdC, jrTKDZcxNBvtdAP6, jw4MF32ZkzJmDnad, nDafN2KYqRyXbkjG, nfLmGxNDqJpyXB6k, nmwM719fpABxvjdz, pqmQRZAcyMa1L98X, prQ8j6B4ZRdyM3LH, pxRAZ6j8byJwmMTF, rpXY2QZPH7aKbnyk, wpQDbBkdCYzjF7AR
- Ravager (73): 1rdmQ9wptPWHfNG6, 2RVJaMQTN7LrCB3m, 2VAWrb67G8KMzgj9, 4NYV7znpwqtKM3mc, 6xz4k73CtjgrZATF, 8Hf42pFZjN9VBrXL, ALXpC8JYPQkMKn37, B4ak2F9hrztgQJH6, BqdMTtCvN3fVJpRk, CRgvLNQYd8y3c2m6, ChQpLdxH2Mfj3cVk, Dk63pRXmjnJ7bANt, DwjkMx9mTZ24gaFh, FMaAjHcwCQLkpV32, H1CMQLTxXbKyBAWw, J9B3mTqZ6KPNjAFY, JHrkTPq23LpAcGxY, K1RGmwayXJ6Lxvb4, K1ybYhmj47pWfG3D, KFLjdptc3gqCGByx, KXCzdLrATV4Gfqtw, LNkzZDmPG9CKw8vV, MZbXt61x7yK9arDc, MyRFjDzr1Kn4b2d9, N3Y1mFkJ9Dwr74bA, QHjqK4G3pzgRWaN6, QVpNH3WvmhYbaFwy, RNCXFn4H2hgqAkp6, TJ9hP23LykXNDfn8, V3L7kJDzTGg4QqCF, VRcx4A9K8PZFyzHf, Wb79QjX8wN4GkgrV, WpgMVC4w2LGy9HQT, XpnDCMx9BtbWa1FG, Y7ZntCL98FcmXRNd, YXgt6fkMdBTKDrAn, Z64QwbKqnJCT8NRz, agdqG2bC714NjD3r, amTgyKJbBwWcxZ69, avhktV7dAyZPzDnb, cMP4Jb8QznHwN6Th, d3bDt8AVT1caMYWn, d8cAKDNpQLRM9Gva, dLXJ21r9jzAmtgf6, dpVr9znc2yCBHRvX, fWgMdvDnKGRxAt4T, gapAZfyCTYnmQkKF, gn7Wva1FKAm8HhGZ, hf1pF42aPHMv9kLX, jVqTWyNz1amfc7ng, jzmMkJpRG2X1HNxr, kDjcaHqQvWpXryKn, kK1rBNaf9Yzq7Cbm, kQZqRBDMaNbX2jdV, kYrgAKnb6GNpvcTh, n9bvwD3yFhLpH2ra, nNapCAWTtqfHZV2m, pK7H2P8mqjyJ34QD, pgPZv2kcYVyA87fW, qMQdaN1g43vXpLzk, qTctabkYwxQKHLhW, rXKd6hBGZkvYFyTc, tYamjwLT8B4k12Fh, vrX8N4PRwHFqgJYM, w2PLVMzXTNvB1pAW, wp7Wx3jJVmtyn9qY, wzpyhaFkq34CGBZc, x268nX3VfMqAvwtN, xCMngy9JrNtkGFVp, xF4yWdk2RnvzBChL, yG2MYFkJjKvmr1fg, ydbJ42ZkTaHnN3Br, zM2hwbVmgyJHRPFt
- The Black Knight (4): 6cQDyzvCrgRdMp8Y, 743vgn8zKVJmPkrC, GrcTMNyY8p2f6xV9, PTcFdkjJ3rBCWNq6
- Heartseeking Crossbow (5): XtFzCQ93xfPk1BqD, daT4FcQgCYw2WBq6, mxbK4d1k96GzwVRh, vZqVmLYQa6jwP3A9, yd8vFWNnTtH6xL9K

## Ironfoe

Ironfoe's extra attacks log as 15494 (Fury of Forgewright), one `extraattacks` event of 2 per proc. Two
protection warriors wore it in the main hand with a shield and no Hand of Justice, in Blackrock Spire, Dire Maul and
Blackmaw Hold (BDFj9VNrfXhxwQT3, bzwY28XH6Vfc4Zpg; every fight). Hits are white swings, cast melee abilities and
Windfury Totem strikes; Shield Slam, Shield Bash, Thunder Clap and the swings within 600 ms after a proc are left out.
Hits within 50 ms of each other form a group; a group of k hits on targets of one race has a 1 - (1 - p)^k chance to
proc, and p is fitted per race by maximum likelihood (95% interval from the likelihood ratio). Targets are Orcs by
name: Scarshield, Blackhand, Firebrand and Bloodaxe (not their worgs), Quartermaster Zigris, Overlord Wyrmthalak and
Warchief Rend Blackhand. Groups that mix races (25 of 157 procs) are left out.

| Targets  | Hits  | Procs | Chance [95%]       | Log-likelihood below the best fit at 3% / 6% |
| -------- | ----- | ----- | ------------------ | -------------------------------------------- |
| Not Orcs | 1,846 | 61    | 3.34% [2.58, 4.22] | 0.34 / 13.57                                 |
| Orcs     | 1,385 | 71    | 5.22% [4.12, 6.48] | 9.46 / 0.77                                  |
