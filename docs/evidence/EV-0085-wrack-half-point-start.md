# EV-0085: Wrack ticks the half points of a dot from the channel's start

- Claim: while Wrack is channelled the half points of an effect's tick grid tick too, from the instant the channel starts, and a channel that lands up to 0.05 s after a half point may still tick it.
- Sources: combat logs, two samples: two logs of one Affliction warlock (36 effects with Wrack channelled on their target, 329 ticks), and two more logs of another warlock and three others (270 effects, timed to about 0.05 s); report codes not recorded
- Game build: not recorded
- Date: not recorded
- Pull request: #953

## The grid

An effect ticks on a grid set by when it landed: Corruption and Siphon Life every 3 s, Bane of Agony every 2 s. While Wrack is channelled the half points of that grid tick too. Starting or ending the channel moves no tick. An effect that runs out ends on a full tick at its expiry, and no tick is ever reduced.

## The start of a channel

In the second sample a half point ticked in 56 of 57 cases when the channel landed before it, in 6 of 10 when it landed up to 0.05 s after, and in 0 of 29 beyond that. Wrack cast one global cooldown after Siphon Life ticked its 1.5 s half point in 13 of 16 openers. In a chained opener, Siphon Life then Wrack 1.29 s later, just ahead of the 1.5 s half point, the half point ticked: ticks at 1.46, 3.03, 4.47 and 5.94 s, and the target died at 7.23 s. The first sample read a half point ticking whenever the channel landed up to 0.03 s after it; the second, larger sample's 0.05 s margin stands.

Half points near a channel's start, by where the channel landed relative to the half point:

| Landed                | Ticked   |
| --------------------- | -------- |
| 0.05 s or more before | 40 of 40 |
| under 0.05 s before   | 16 of 17 |
| under 0.05 s after    | 6 of 10  |
| 0.05 s or more after  | 0 of 29  |

A channel that starts more than half a period after the effect's last tick does not tick at once: 24 of 25 such starts waited for the next point, and the one that ticked started 0.045 s before a half point.

## The end of a channel

The end of a channel on a half point is not measured. Within 0.05 s either side it ticked about half the time, and the logs' 0.05 s steps cannot split it.

| Ended                 | Ticked   |
| --------------------- | -------- |
| 0.05 s or more before | 1 of 22  |
| under 0.05 s before   | 2 of 7   |
| under 0.05 s after    | 4 of 8   |
| 0.05 s or more after  | 20 of 20 |

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
