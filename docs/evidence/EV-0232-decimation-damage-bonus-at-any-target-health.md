# EV-0232: Decimation's damage bonus counts at any target health

- Claim: with Decimation (440870) taken, Shadow Bolt's and Searing Pain's damage bonus has no health gate: it is the
  same against a target below 35% and at 40% and above, with the buff 440873 up or down. Its size at 2/2 fits the
  row's 6% but is not pinned down (§Size). Rank 1/2 is not measured.
- Sources: local combat logs with advanced logging, other players' warlocks: 31 logs in which a warlock gains 440873,
  and 142 logs with warlocks who never gain it as controls. Eight warlocks gain 440873 (§Warlocks); five of their logs
  list Decimation 2/2 in the combatant info, and two list no talents. The control warlocks are 16 for the health test,
  and one level 60 warlock without Decimation for the size reads.
- Game build: by each log's local time and UTC offset, against the time the CDN first listed each build. The logs of
  the warlocks with Decimation ran 1.60.1.70338 on the beta (from 2026-10-10 05:05 UTC), then 70862, 70888 (from
  2026-11-10 15:11 UTC in the US), 70938, 70959 (from 2026-11-18 02:04 UTC in the EU), 71037 and 71139 (from
  2026-11-25 19:14 UTC). The controls ran the beta and live builds of 2026-09-21 to 2026-11-24. Every log's header
  states COMBAT_LOG_VERSION 22 and BUILD_VERSION 1.60.1. The hotfix each player ran is not recorded.
- Date: 2026-10-10 to 2026-11-25; controls 2026-09-21 to 2026-11-24
- Pull request: #1280

## Method

Each hit's r is the log's base amount over the rank's average plus the spell's coefficient times the spell power the
log's advanced block states for the caster. The target's health is the advanced block's health before the hit. Only
creature targets of level 1 to 60 count, crits included, and hits at 35% to 40% health are left out. The logs state no
caster level: the size reads price the rows at level 60, and warlock B's at 52 and at 53.

The health test divides a warlock's mean r below 35% by its mean r at 40% and above, with a bootstrap 95% interval.
The pooled ratio is the inverse-variance mean of the warlocks' log ratios. Hits at 40% and above are split as well by
whether 440873 is up, each against the same warlock's hits below 35%.

The size reads take r of Searing Pain over the same warlock's r of a fire spell Decimation does not name (Soul Fire,
Immolate's direct hit), and r of Shadow Bolt over the same warlock's Corruption ticks, so an error in the logged spell
power moves both sides. Hits under Blood Fury (20572) or Power Infusion (10060) are left out there, and Shadow Bolts
on a target with Shadow Vulnerability (17794). For a warlock with Aftermath 5/5, Immolate's direct hit is divided by
Aftermath's 50% bonus. A level 60 warlock without Decimation gives the same reads with no bonus in them.

## Health

Ratio of r below 35% to r at 40% and above, with the hits on each side:

| Spell        | Decimation | Hits at 40% and above | Warlocks | Hits      | Below / above        |
| ------------ | ---------- | --------------------- | -------- | --------- | -------------------- |
| Searing Pain | taken      | all                   | 6        | 412 / 502 | 0.996 [0.990, 1.003] |
| Searing Pain | taken      | 440873 down           | 6        | 412 / 425 | 0.995 [0.989, 1.001] |
| Searing Pain | taken      | 440873 up             | 4        | 394 / 75  | 1.006 [0.995, 1.016] |
| Shadow Bolt  | taken      | all                   | 5        | 40 / 47   | 0.982 [0.966, 0.999] |
| Searing Pain | none       | all                   | 4        | 259 / 261 | 1.000 [0.993, 1.006] |
| Shadow Bolt  | none       | all                   | 15       | 439 / 488 | 0.999 [0.994, 1.003] |

A bonus gated at 35% would read 1.06. Shadow Bolt's dip comes from warlock A's 8 and 11 hits, 0.929
[0.890, 0.971]; warlocks B, D, E and G read 0.989 to 1.003.

## Warlocks

| Warlock | Log date             | Decimation in the log | Searing Pain hits | Searing Pain below / above |
| ------- | -------------------- | --------------------- | ----------------- | -------------------------- |
| A       | 2026-10-10, beta     | 2/2                   | 58 / 77           | 1.000 [0.985, 1.015]       |
| B       | 2026-11-10           | 2/2                   | 39 / 123          | 0.991 [0.977, 1.005]       |
| C       | 2026-11-16           | 2/2                   | 15 / 9            | 0.982 [0.959, 1.006]       |
| D       | 2026-11-19 and 11-21 | 2/2                   | 173 / 188         | 0.992 [0.980, 1.004]       |
| E       | 2026-11-25           | 2/2                   | 124 / 99          | 1.004 [0.993, 1.015]       |
| F       | 2026-11-23           | no talents listed     | 3 / 6             | 0.989                      |
| G       | 2026-11-08           | no talents listed     | -                 | -                          |

Warlocks C, D and E also take Agonizing Flames 3/3, and C and D Aftermath 5/5.

## Size

| Warlock       | Ratio                                              | Below 35%            | At 40% and above     |
| ------------- | -------------------------------------------------- | -------------------- | -------------------- |
| D             | Searing Pain / Soul Fire                           | 1.058 [1.042, 1.072] | 1.060 [1.044, 1.075] |
| E             | Searing Pain / Immolate's direct hit               | 1.061 [1.054, 1.068] | 1.059 [1.051, 1.068] |
| E             | Searing Pain / Soul Fire (7 Soul Fire hits)        | 1.057                | 1.056                |
| D             | Searing Pain / Immolate's direct hit, no Aftermath | 1.059                | 1.062                |
| C             | Searing Pain / Immolate's direct hit, no Aftermath | 1.053                | 1.073                |
| no Decimation | Searing Pain / Immolate's direct hit               | 0.988 [0.982, 0.994] | 0.988 [0.983, 0.993] |

Shadow Bolt over Corruption's tick, at any health:

| Warlock       | Read as level | Shadow Bolt / Corruption tick | Below 35% | At 40% and above |
| ------------- | ------------- | ----------------------------- | --------- | ---------------- |
| B             | 52            | 0.969 [0.956, 0.982]          | 0.963     | 0.975            |
| B             | 53            | 0.965 [0.952, 0.978]          | 0.960     | 0.971            |
| no Decimation | 60            | 0.915 [0.911, 0.920]          | 0.914     | 0.919            |

Warlock B and the warlock without Decimation take the same build apart from Decimation, so Shadow Bolt's bonus reads
5.5% to 5.9%. That size rests on one warlock with Decimation.

The Searing Pain reads have two open questions. The warlock without Decimation reads 0.988 on Searing Pain over
Immolate's direct hit, not 1.000; measured from that, warlock E's bonus is about 7.4%. Agonizing Flames 3/3 raises
Searing Pain and Immolate by 10% but not Soul Fire, so warlocks D and E, who take it, should read about 1.17 over Soul
Fire, yet they read 1.06; why is not known. Whether the bonus multiplies with Agonizing Flames or adds to it is not
settled either.
