# EV-0170: Defias Enforcer's Garb's Defias Juggernaut is about 1 PPM on either hand

- Claim: The Defias Enforcer's Garb 5-piece bonus (1324684) gives Defias Juggernaut (1324683) at about 1.0 PPM per
  landed melee hit, priced on the speed of the weapon that hit (the paw in Cat or Bear Form). Off-hand white swings
  proc it as well as the main hand, though the row names the main hand alone. A PPM fits better than a flat chance per
  hit. The row's base of 25 is not the chance. Every lone-unit count reads 0.93 to 1.12 and the sim takes the pooled
  1.0; the all-hits count reads 1.13 and does not overlap the specials, so the rate is open between those numbers.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every fight with all 5 pieces worn in the 73 reports
  under Reports (rogue and druid wearers). 4ntRLypA6zd8YBQm, Wd2TpPj8GnCyM4bD, aAMfcnCKRpT49J27 and jn3dwzWPZ6r2ky1K
  log runs that those reports hold too, and count once. The hunter wearers' reports are a check only.
- Game build: 1.60.1.70888 to 1.60.1.71037 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-11 to 2026-11-20
- Pull request: #1146

## Method

44 rogues (all dual wielding), 7 druids (3 with Cat Form stretches) and 4 hunters (6 runs), with all 5 pieces worn per fight's
CombatantInfo. Only events the wearer is the source of count (pets left out). A run logged in several reports counts
once: two runs of one player whose damage events share over 30% of their seconds are one run. A proc is an applybuff or refreshbuff of 1324683. The buff holds one stack and logs a refresh when it procs again while up (5,460 refreshes against 5,104
applies). It logs 0 to 70 ms after the hit that rolled it, mostly at 17, 33 or 50 ms, so a hit gets credit for a proc 20
ms before to 120 ms after it. Druid swings are priced on the paw: Cat Form 1.0 s, Bear Form 2.5 s.

- All hits (rogues and druids in Cat Form): landed white swings of both hands, each priced on its own hand's speed;
  melee specials priced on the main hand (Swipe (Bear), Cleave and Maul left out); and the landing of a melee debuff
  with no damage (Kidney Shot, Garrote and the like), priced on the main hand. Procs from 120 ms before to 20 ms after
  a trigger outside this table (Distract, Slice and Dice, traps, Demoralizing Roar, Gouge, Blind, Expose Armor) are
  dropped: 197 procs. Hands are split by Otsu on log damage (crits halved). Where the split fails (close speeds or few
  swings), the swing is priced on the harmonic mean of the two speeds.
- Lone units: a hit with no other event of the wearer within 200 ms. White swings count only where the damage split
  and a swing-timing split name the same hand.
- Shape: on all lone units, a PPM model (chance = PPM × speed / 60) is set against one flat chance per hit by likelihood,
  with a bootstrap over wearers.
- Swap check: between two procs, the rolls due at 1.05 PPM were summed. 33 of 8,964 such stretches of rogues had 5
  or more due and no proc, fewer than chance alone gives (about 60, e^-5 of them), so no stretch is dropped as a swap.

## Rate

| Count                              |   Procs / rolls | PPM [95%, bootstrap over wearers] | Chance per roll | Wearers |
| ---------------------------------- | --------------: | --------------------------------- | --------------- | ------: |
| All hits (rogues, Cat Form druids) | 8,979 / 212,580 | 1.128 [1.110, 1.161]              | 4.22%           |      47 |
| Lone specials, all classes         |    730 / 16,601 | 1.027 [0.950, 1.072]              | 4.40%           |      51 |
| Lone specials, rogues              |    685 / 15,115 | 1.037 [0.965, 1.085]              | 4.53%           |      44 |
| Lone white, rogue off hand         |    380 / 11,758 | 1.054 [0.968, 1.254]              | 3.23%           |      36 |
| Lone white, rogue main hand        |     205 / 5,151 | 0.930 [0.751, 1.032]              | 3.98%           |      36 |
| Lone white, Cat Form (1.0 s)       |      57 / 3,052 | 1.121 [0.538, 1.825]              | 1.87%           |       3 |
| Lone white, Bear Form (2.5 s)      |      48 / 1,112 | 1.04                              | 4.32%           |       7 |
| Every lone unit pooled             |  1,490 / 40,827 | 1.000 [0.926, 1.064]              | 3.65%           |      51 |
| Hunters, lone white and specials   |     154 / 5,167 | 0.63 to 1.10 by kind              | 3.0%            |       4 |

The PPM model fits the lone units better than a flat chance: the log-likelihood is 55.3 higher (bootstrap 95% from 14.3
to 101.2). The chance per swing follows the speed: Cat Form's 1.0 s paw reads 1.87% and Bear Form's 2.5 s paw 4.32%;
rogue specials read 1.9% on a 1.7 s main hand (5 of 265) and 4.4% on 2.7 s (330 of 7,432).

The off hand procs it. Lone off-hand white swings on 1.3 to 1.5 s off-hand weapons, where the damage split is
cleanest, read 2.84% (55 of 1,935, 7 wearers), 1.2 PPM. All lone off-hand swings read 1.05 PPM, the main hand 0.93.
A check with no hand labels: rogues' lone white swings proc 1,364 of 37,316 times (3.66%). With the main hand taking
42% of the swings (2.57 s against 1.84 s), a main-hand-only proc would need 8.8% on main-hand swings, 2.05 PPM, twice
what main-hand specials read (1.04 PPM). Both hands at 1.04 PPM give 3.71%.

The all-hits count reads 10% above the lone units. Procs land beside rolls the table leaves out or may misprice:
poison applications, Blade Flurry's copied hits, Windfury Totem hits, and harmful casts with no damage. Which of
these roll is not settled.
On rogues' lone units Distract procs it 111 of 1,145 times (a debuff on every mob in its area), Slice and Dice 71 of
3,157 and Kick 24 of 366. Shank (10 of 6,308), Sap (0 of 181), misses (5 of 17,306), ticks (0 of 6,290) and spell hits
such as Instant Poison (0 of 1,616) do not proc it.

## Reports

1GbLMCdmYgtXn2pB, 1HQ8vBNRDJjLyxCd, 2mL3ZCv9WKTQ18by, 3YHygp8nqx7DRkbT, 4VtCK6Qqr1MD2hpc, 6Bf9dRJMNXyFtD8n,
73xcRhZpPmNvC6fw, 74PAn1mMkWwYGcDd, 79QPzLcgZCKbJBkH, 7fz9Kk8rY6jyQAbL, 8qKTwcCyPgd4B3xZ, 9tjaL4MdpWx6rmhT,
ATMmnV2tC3ZPb6G8, BDFj9VNrfXhxwQT3, CFJKwbPfgA7Hha43, HDmZ4zdFK6xgPCT8, JHMxQaYWm8q7k9GZ, JavZtXcdp9WgLQ8x,
Kphk63jtr1wb9NyL, L4znYqVQRx3j6phB, LFk4QfzPtr9ZbBh1, LHTZ1cQVBxgAFY3R, LWZt9AV1qXdxaJv8, MAdN3c4XZVtH1bvF,
NarJRCMQth3qkXV1, NcxbzW4kJYQ7DGhj, QFqKVZgc6vAJfbrW, QGqj8mLPkN71DKzy, W3nLgjDvbPz46BdR, WP97jBRk24ynfYAJ,
YTn9F1x6jLqN2fdQ, Z3qdRnrYyXtkhLPM, aJgXTdrAbyLFpfm1, aZKTvVcCHLwhgXzM, b26xVgATjNwfDpYz, b3aG821WYQj9RVhK,
bBZXmynRjLkzwW6Q, bGjk9YQT8DgdKFPf, bZaY8JyWxpmwQDjB, cXf64Z1GR72MYnbF, dK9TAcBfjVmD2n1J, fM8Qchgt9dP4aNWL,
gBDMLv8J1K7nqkpd, gXhQJLaTxK94WfHr, h43GgypRxfbcznLD, h4q8kfT9GVmMczbB, h9nNL4WmZfjHzvB6, hamTvL4wFpx6nYNj,
hwrmtk6aA7zJ8N1Q, jTX2kZ1K7gzt3dcL, jrPWFdYfQNz3xG6C, kxt1vq6DGM7YPA8n, myBT4jMY7NJPnxfw, n4zkZb93PKQWx8aC,
nNmBkKp3J96FTGfg, pX4cj6KQv7JbWAC2, pc7drLMKbFJ1nhGv, pgN9QwMLYVfrqGbv, ptvYCVPymJXH21kR, q9K4Zv2mk3dnGaAP,
qL9HjmZWxGKwkMr2, rmKx9LgCkPfWvp7h, tJnq3MAB7ChKmrfF, tz7p9c2gw1NPJMkn, v1bPdgDtXypHncfN, vLGwZxXBaVcpMT7H,
x6MTjGBYanhpJf1k, xFhXRtJ4KDVHqvmW, xdFjKnGbC2caDhPJ, yLRATGz8NbQJdKP4, yPVkKpqtfJM27WFB, zhNBgjHxQX7anZTf,
zxLCmJbZA2pQhyHR
