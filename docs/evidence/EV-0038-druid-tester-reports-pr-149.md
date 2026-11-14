# EV-0038: Swipe's attack power share, Rake's hit and armor, and Rend and Tear, from tester reports on the feral druid pull request

- Claim: Swipe takes 3% of attack power on each hit, Rake's initial hit is not reduced by armor, and Rend and Tear raises melee special attacks and bleed ticks against a bleeding target but not white swings.
- Sources: tester reports and tables quoted on the pull request; its body holds a short overrides table only. Two testers' Rake and Swipe attack power scaling tests, and a Rip attack power scaling test sheet
- Game build: not recorded
- Date: 2026-10-01 for the Rake and Swipe tests; 2026-09-26 for the Rip sheet
- Pull request: #149

What the pull request body records for Rip and Rake, and the Swipe Rank 5 override's Source (the body lists Swipe's share as an open question):

| Row                                                | Value                          | Source                                                |
| -------------------------------------------------- | ------------------------------ | ----------------------------------------------------- |
| Rip 9896, attack power per combo point per tick    | 0.01 (APCoefPeriodic)          | Rip attack power scaling test sheet, 2026-09-26       |
| Rake 1822, 1823, 1824, 9904, attack power per tick | 0.055 (APCoefPeriodic)         | two testers, 2026-10-01                               |
| Swipe Rank 5 (9908)                                | 3% of attack power on each hit | Rake and Swipe attack power scaling tests, 2026-10-01 |

A tester's table showed Rake and Rip ticks, the Rake hit and Claw rising by the talent's percent with a bleed on the target, and the Rake hit staying put without one. A tester's report stated that Rake's initial hit is not reduced by armor, although only the periodic effect states the bleed mechanic. No counts are recorded in the body.
