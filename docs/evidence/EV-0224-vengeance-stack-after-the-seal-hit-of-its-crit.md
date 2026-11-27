# EV-0224: A crit's Vengeance stack comes after the seal hit of the same swing

- Claim: the Vengeance stack (talent 20049, buff 20050, logged as Conviction) that a white swing's crit adds lands after
  the hits that swing procs. The swing's Seal of Righteousness and Seal of Fury hits are priced on the stacks from
  before the crit, and so is its Windfury Totem hit.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every paladin stream on disk in the sweep's stream cache
  that holds a Vengeance event: 1,246 streams in 1,224 reports, 600 streams in 586 reports before hotfix push 114313
  and 646 streams in 638 reports after it. The report codes are in the scripts' inputs and are not listed here.
- Game build: 1.60.1.70862 to 1.60.1.71139 by report date (70862 from 2026-11-07 04:43 UTC, 70888 from 2026-11-10 15:11
  UTC in the US, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23 UTC,
  71091 from 2026-11-24 15:21 UTC in the US, 71139 from 2026-11-25 19:14 UTC). By build, 6, 6, 225, 327, 479, 108 and
  73 reports. The reading holds on both sides of hotfix push 114313 (2026-11-20 23:47 UTC). The hotfix each player ran
  is not recorded.
- Date: 2026-11-07 to 2026-11-27
- Pull request: #1270

## Method

The probe reads each paladin's own damage and aura events. Vengeance's stacks come from the paladin's own Vengeance
apply, stack, refresh and remove events, with no expiry timer: a crit at five stacks logs no refresh, so the state
follows the remove events alone.

Each white swing crit (ability 1, hit type 2) is paired with the nearest Vengeance apply, stack or refresh within 50 ms
of it. That event names the crit's change, k−1→k. Of the paired events, 52,173 log after the crit, 1,229 in the same
millisecond after it, 171 in the same millisecond before it and 104 before it. 44,442 crits have no Vengeance event
within 50 ms. The Seal of Righteousness hits of those swings are dealt at five stacks in 13,965 of 14,092 cases.

A seal hit of the swing is a normal hit (hit type 1) of that seal on the swing's target 0 to 50 ms after the crit. Its
amount is divided by the same player's median for that spell id with Vengeance down. Each "other" hit is a normal hit
of the same seal more than 60 ms from any white crit, grouped by the stacks it was dealt under, and gives what a hit at
those stacks reads. A hit priced before the crit reads like the other hits at k−1, and one priced after it like those
at k. Windfury Totem and Seal of Command hits are read the same way.

Scripts and outputs: `requests/forever-aura-order-1127` in the log parser's folder: `extract.py` builds the stream
cache, `probe.py` runs the probes (`conviction_noexp_sor`, `conviction_noexp_wft`, `conviction_noexp_soc` and
`conviction_sof`), and `probe_out.txt`, `probe_split_out.txt` and `conviction_out.txt` hold the output.

## Seal of Righteousness

Median hit over the player's median with Vengeance down; n is the number of hits.

| Crit's change | Hits of the swing | Other hits at k−1 | Other hits at k |
| ------------- | ----------------: | ----------------: | --------------: |
| 0→1           |     2,254 · 1.000 |    15,354 · 1.000 |  17,448 · 1.030 |
| 1→2           |     2,734 · 1.029 |    17,448 · 1.030 |  16,967 · 1.060 |
| 2→3           |     2,682 · 1.059 |    16,967 · 1.060 |  14,514 · 1.090 |
| 3→4           |     2,269 · 1.091 |    14,514 · 1.090 |  12,106 · 1.119 |
| 4→5           |     1,872 · 1.120 |    12,106 · 1.119 |  68,961 · 1.150 |

Split at hotfix push 114313, the medians of the swing's hits and of the other hits at k−1 and at k:

| Crit's change | Before the push       | After the push        |
| ------------- | --------------------- | --------------------- |
| 0→1           | 1.000 (1.000 / 1.029) | 1.000 (1.000 / 1.031) |
| 1→2           | 1.028 (1.029 / 1.059) | 1.031 (1.031 / 1.061) |
| 2→3           | 1.061 (1.059 / 1.090) | 1.058 (1.061 / 1.090) |
| 3→4           | 1.091 (1.090 / 1.118) | 1.092 (1.090 / 1.120) |
| 4→5           | 1.117 (1.118 / 1.149) | 1.121 (1.120 / 1.150) |

Before the push there are 934, 1,178, 1,147, 1,024 and 817 hits of the swing; after it, 1,320, 1,556, 1,535, 1,245 and
1,055.

## Windfury Totem hit

| Crit's change | Hits of the swing | Other hits at k−1 | Other hits at k |
| ------------- | ----------------: | ----------------: | --------------: |
| 0→1           |       622 · 1.014 |     8,274 · 1.000 |   5,938 · 1.033 |
| 1→2           |       496 · 1.039 |     5,938 · 1.033 |   4,691 · 1.062 |
| 2→3           |       409 · 1.053 |     4,691 · 1.062 |   3,835 · 1.088 |
| 3→4           |       330 · 1.083 |     3,835 · 1.088 |   3,221 · 1.121 |
| 4→5           |       281 · 1.134 |     3,221 · 1.121 |  16,142 · 1.162 |

Every change reads nearer to k−1 than to k.

## Seal of Fury

Seal of Fury was read only by the probe that ends Vengeance 30 s after its last logged event. That timer drops five
stacks that log no refresh to none, which lifts the median with Vengeance down that every ratio is divided by. The
ratios therefore compare only with each other. In the same run, Seal of Righteousness reads 0.882, 0.903, 0.929, 0.952
and 0.977 for the swing's hits against 0.908, 0.932, 0.957, 0.982 and 1.007 for other hits at one to five stacks.

| Crit's change | Hits of the swing | Other hits at k−1 | Other hits at k |
| ------------- | ----------------: | ----------------: | --------------: |
| 0→1           |       369 · 0.905 |    30,986 · 1.000 |   2,648 · 0.931 |
| 1→2           |       321 · 0.935 |     2,648 · 0.931 |   2,061 · 0.956 |
| 2→3           |       300 · 0.938 |     2,061 · 0.956 |   1,926 · 0.980 |
| 3→4           |       247 · 0.970 |     1,926 · 0.980 |   1,664 · 1.000 |
| 4→5           |       239 · 1.000 |     1,664 · 1.000 |   5,259 · 1.000 |

From 1→2 to 3→4, the swing's hits read nearer to the other hits at k−1 than to those at k, as Seal of Righteousness's do
in the same run. At 4→5 the two columns tie. The 0→1 row says nothing in this run, since its k−1 column is the lifted
median itself.

## Seal of Command

| Crit's change | Hits of the swing | Other hits at k−1 | Other hits at k |
| ------------- | ----------------: | ----------------: | --------------: |
| 0→1           |     3,396 · 1.006 |    21,195 · 1.000 |  18,390 · 1.028 |
| 1→2           |     4,115 · 1.031 |    18,390 · 1.028 |  13,638 · 1.053 |
| 2→3           |     2,664 · 1.044 |    13,638 · 1.053 |  10,891 · 1.084 |
| 3→4           |     1,792 · 1.086 |    10,891 · 1.084 |   8,580 · 1.109 |
| 4→5           |     1,507 · 1.136 |     8,580 · 1.109 |  42,249 · 1.143 |

Seal of Command does not settle the order. Its 2→3 hits read below both columns, and its 4→5 hits read nearer to k.
