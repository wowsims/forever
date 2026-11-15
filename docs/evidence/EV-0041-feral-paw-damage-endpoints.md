# EV-0041: Base paw damage endpoints of Cat and Bear Form

- Claim: the base paw endpoints before attack power and flat weapon damage, as the in-game measurements recovered them.
- Sources: in-game measurements; nothing else recorded
- Game build: not recorded
- Date: not recorded
- Pull request: #623; its body holds no table

The pull request computes paw damage from the weapon's table DPS and each form's variance (0.4), in float32 milliseconds, and keeps the existing measured endpoint and flat-damage tests. The endpoints themselves are the test table in `sim/druid/imbue_test.go`.
