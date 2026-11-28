# EV-0085: Wrack ticks the half points of a dot from the channel's start

- Claim: while Wrack is channelled the half points of an effect's tick grid tick too: a half point ticks when the channel is on the target at its instant, from the channel's start up to its end, and a 6 s channel lined up on the grid ticks one of the two half points at its edges.
- Sources: combat logs, two samples: two logs of one Affliction warlock (36 effects with Wrack channelled on their target, 329 ticks), and two more logs of another warlock and three others (270 effects, timed to about 0.05 s); report codes not recorded
- Game build: not recorded
- Date: not recorded
- Pull request: #953
- Later count: the start of a channel, the end of a channel and a channel lined up on the grid, §Count of the 2026-11-13 to 2026-11-27 reports

## The grid

An effect ticks on a grid set by when it landed: Corruption and Siphon Life every 3 s, Bane of Agony every 2 s. While Wrack is channelled the half points of that grid tick too. Starting or ending the channel moves no tick. An effect that runs out ends on a full tick at its expiry, and no tick is ever reduced.

## The start of a channel

- Count: §Count of the 2026-11-13 to 2026-11-27 reports

Wrack cast one global cooldown after Siphon Life ticked its 1.5 s half point in 13 of 16 openers. In a chained opener, Siphon Life then Wrack 1.29 s later, just ahead of the 1.5 s half point, the half point ticked: ticks at 1.46, 3.03, 4.47 and 5.94 s, and the target died at 7.23 s.

A channel that starts more than half a period after the effect's last tick does not tick at once: 24 of 25 such starts waited for the next point, and the one that ticked started 0.045 s before a half point.

## The end of a channel

- Count: §Count of the 2026-11-13 to 2026-11-27 reports

## A channel lined up on the grid

- Count: §Count of the 2026-11-13 to 2026-11-27 reports

## Four candidate rules

Fit of four candidate rules over the first sample's 36 effects, and over the second sample's 270:

| Rule                                            | First sample matched, missing, extra | Second sample matched, missing, extra |
| ----------------------------------------------- | ------------------------------------ | ------------------------------------- |
| Grid from the effect's application              | 312, 4, 3                            | 1,689, 9, 10                          |
| Time left to the next tick scaled by the period | 221, 99, 94                          | 1,020, 699, 679                       |
| Previous tick plus the new period               | 248, 76, 67                          | 1,353, 375, 346                       |
| Pending tick stands, new period after it        | 250, 39, 65                          | 1,490, 130, 209                       |

The grid's 7 misses in the first sample are a death at a tick, an expiry logged 0.03 s early, and five channel starts or ends within the logs' 0.2 s jitter of a half point. Of its 19 misses in the second sample, 15 are half points within 0.05 s of a channel's start or end and 3 sit in a death's batch; one is a real exception, a Siphon Life that lost a normal tick once, between two channels.

## Expiry and ramp

All four effects that ran out after a channel ended on a full tick: 35 ticks in the logs. Three sped-up Corruptions ran out in the second sample, each on a full tick at its expiry, and no reduced tick appears among 919 sped-up Corruption and Siphon Life ticks. Bane of Agony's ramp follows the time since it landed, with a tick on a third's end taking the lower third: 160 of 160 ticks in the first sample and 1,938 of 1,938 in the second, where the tick number misses 174. Soul Siphon adds 12% for each of the three effects, read at each tick, and Wrack speeds up only its caster's effects.

## Logged cases

A logged Corruption with Wrack channelled from 3.20 s to 5.94 s after it landed ticked at 3.2, 4.7, 6.0, 9.2, 12.1, 15.2 and 18.2 s. A Corruption with Wrack from 6.21 s to 8.71 s ticked its 7.5 s half point at 7.43 s and ended on a full tick at 18 s with 7 ticks of 78, where it has 6 without Wrack. A Corruption with Wrack from 12.44 s to 16.41 s ticked at 2.98, 6.00, 8.98, 11.98, 13.47, 14.90 and 17.91 s: the half point at 16.5 s did not tick, the channel having ended 0.09 s before it.

## Open

Three things are not modelled and not settled: 20 channels ended near 4.5 s with a full Wrack tick logged at their removal, with no cast or death at that moment; an effect applied during a channel; and a sped-up Bane of Agony or Siphon Life running out.

## Count of the 2026-11-13 to 2026-11-27 reports

- Sources: Warcraft Logs, 387 Forever reports with 391 warlocks who channelled Wrack, every warlock stream fetched for the proc-rate sweeps; report codes in the count's `records.json`, kept outside the repository with its scripts
- Game build: 1.60.1.70938 to 1.60.1.71139 and the hotfixes between, pooled
- Date: 2026-11-13 17:05 to 2026-11-27 04:07 UTC
- Pull request: not recorded

Each Corruption, Siphon Life and Bane of Agony instance with no refresh is put on the grid of its application as logged (its applydebuff), and each half point within 0.3 s of a Wrack channel's start or end on the same target is checked for a tick within 0.15 s of it that is not one of the effect's own ticks. Timestamps are in milliseconds. The effects' own ticks log a median 4 ms after their grid point (5th to 95th percentile −10 to +17 ms, 183,187 ticks), so a point's instant is known to about that.

Half points near an edge with no other channel edge near the point, by where the point lies relative to the edge:

| Half point              | Ticked, near a start  | Ticked, near an end  |
| ----------------------- | --------------------- | -------------------- |
| 0.16 s or more before   | 0 of 3,372            | 3,442 of 3,475       |
| 0.04 to 0.02 s before   | 30 of 541             | 349 of 355           |
| under 0.02 s before     | 517 of 1,771          | 511 of 700           |
| under 0.02 s after      | 2,184 of 2,663        | 112 of 798           |
| 0.02 to 0.04 s after    | 4,800 of 4,816        | 2 of 935             |
| 0.16 s or more after    | 5,077 of 5,101        | 0 of 2,440           |

A half point ticks when the channel is on the target at its instant, at the start and at the end alike: either way the share passes one half within 5 ms before the logged edge, and within 0.02 s the share follows the spread of the effects' own ticks. No tick logged after the channel's removal, and 413 logged in the same millisecond as it; 569 logged in the same millisecond as the channel's start. A channel that lands a few hundredths of a second after a half point ticks it only where the point fell due in that same instant.

Of 309 full channels (5.99 to 6.01 s) that started within 0.01 s of a Corruption or Siphon Life half point, with both edge points away from any other channel's edge and the effect up past the channel's end, 285 ticked exactly one of their two edge points, 22 both and 2 neither: 1.07 edge points on average, so 2.07 half points a channel. Bane of Agony's 108 such channels ticked one edge point in 104, both in 4.
