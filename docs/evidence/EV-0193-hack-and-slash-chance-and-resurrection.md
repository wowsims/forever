# EV-0193: Hack and Slash's chance, and the resurrections that switch it off

- Claim: Hack and Slash (talent 13960, extra attack 16459) procs on a landed hit from a sword or axe hand at about 0.9%
  per rank, 4.5% at rank 5. About half of the resurrections by another player switch it off, and it stays off until
  something not yet identified turns it back on.
- Sources: Warcraft Logs reports of rogues from 2026-11-09 to 2026-11-25 (the per-report table below lists the reports
  the resurrection count rests on), and combat log `WoWCombatLog-112426_191754`, a level 60 rogue in three Scholomance
  runs on 2026-11-24 19:17 to 23:25 UTC-5 (2026-11-25 00:17 to 04:25 UTC)
- Game build: the Warcraft Logs reports span every build and hotfix from 2026-11-09 to 2026-11-25, with no report's
  build recorded; the combat log is 1.60.1.71091, its hotfix push not recorded
- Date: 2026-11-09 to 2026-11-25
- Pull request: #1219

## Chance

A roll is a landed white swing from a hand holding a sword or axe, or a landed direct special (Sinister Strike,
Eviscerate, Backstab, Hemorrhage, Ghostly Strike, Riposte, Gouge, Kick, Shank, Ambush, Mutilate) while the main hand
holds one. The white swings are split between the hands by their logged unmitigated amount (Otsu's threshold on its
log, kept when the split's eta is 0.85 or more). A proc is an extra attack event with ability 16459. The rank and the
gear come from the combatant info in force. A rogue counted in two reports for the same swings counts once. The
intervals are a 2,000-draw bootstrap over players. The counts include the stretches where a resurrection had switched
the talent off, so the chance while it works sits a little above them.

| Rank | Chance per roll | 95% interval | Rogues |
| ---- | --------------- | ------------ | ------ |
| 1    | 0.94%           | 0.87 to 1.00 | 165    |
| 2    | 1.83%           | 1.70 to 1.94 | 143    |
| 3    | 2.61%           | 2.32 to 2.88 | 53     |
| 4    | 3.73%           | 3.57 to 3.87 | 87     |
| 5    | 4.48%           | 4.42 to 4.55 | 881    |

Rank 5 by period (UTC), each within the others' intervals but the six rogues of the fifth:

| Period                           | Chance | 95% interval | Rogues |
| -------------------------------- | ------ | ------------ | ------ |
| before 2026-11-20                | 4.40%  | 4.30 to 4.49 | 402    |
| 2026-11-20 to hotfix push 114313 | 4.55%  | 4.34 to 4.79 | 80     |
| 114313 to 2026-11-24 01:00       | 4.53%  | 4.41 to 4.64 | 311    |
| 2026-11-24 01:00 to 20:54        | 4.51%  | 4.28 to 4.71 | 101    |
| 2026-11-24 20:54 to 11-25 01:13  | 4.03%  | 2.99 to 4.42 | 6      |
| from 2026-11-25 01:13            | 4.72%  | 4.42 to 4.98 | 75     |

Wearing Flurry Axe, Hand of Justice or Sword of Zeal does not move it: 4.45% with none of them (610 rogues), 4.57% with
Hand of Justice (60), 4.52% with Flurry Axe (17).

## Resurrections

Of 848 rank 5 reports with 300 rolls or more, 774 read 3 to 6%, 19 above 6%, 44 from 1 to 3% and 11 below 1%. In
each of the 11 the procs run at the usual rate and then stop. Minutes are from the report's start.

| Report             | Procs, at minutes | Resurrection on the rogue | After it                        |
| ------------------ | ----------------- | ------------------------- | ------------------------------- |
| `mLvFHr346ab2nT9c` | 20, 2.0 to 12.2   | Resurrection at 14.8      | none in 69 minutes              |
| `ftBFNZ3xA7WRVhcX` | 14, 4.0 to 13.6   | Resurrection at 19.7      | none until 101.6, then 5        |
| `2ZLD7Jvwfa1n3xyM` | 8, 1.0 to 11.4    | Resurrection at 13.2      | none in 63 minutes              |
| `XPYV9J6DfrMKxGqh` | 11, 64.9 to 80.6  | Ancestral Spirit at 81.5  | none in 63 minutes              |
| `pac2twnZq7Ck3NJx` | 10, 0.8 to 10.0   | Ancestral Spirit at 14.4  | none in 38 minutes              |
| `1TFWH4Yy38KhXMwx` | 6, 49.2 to 52.2   | Redemption at 54.1        | none in 30 minutes              |
| `G9TVWxhmHjdc3Ptw` | 2, at 16.0        | Resurrection at 17.3      | none until 61.4, then 6         |
| `nQ72mC6qVH4vwgT8` | none              | Redemption at 4.4         | none in the report's 91 minutes |
| `GB6fVLkvzPhxYJcZ` | none              | Redemption at 28.9        | none in the report's 63 minutes |
| `LrNcfVvw3PqBQ9CA` | none              | none in the report        | none in the report's 58 minutes |
| `G6yaZfLzMtN13hcA` | none              | none in the report        | none in the report's 32 minutes |

Three of these rogues show in other reports with the same talents and gear at the usual rate: `LrNcfVvw3PqBQ9CA`'s at
4.41% four hours before and 4.13%, 4.99% and 5.73% later; `GB6fVLkvzPhxYJcZ`'s at 5.02% and 2.67%;
`mLvFHr346ab2nT9c`'s at 4.67% six hours before.

Every resurrection event on a rank 5 rogue in 771 reports, with 100 landed white swings or more on each side before
the next one, is classed by whether the rogue procced on at least 2% of its landed white swings before and after:

| Before, after | Resurrections |
| ------------- | ------------- |
| on, on        | 106           |
| on, off       | 101           |
| off, on       | 13            |
| off, off      | 13            |

83 of the 101 log no proc at all after the resurrection, so the talent is either working or off. Every resurrection
spell switches it off at a like share: Ancestral Spirit 24 of 52, Redemption 23 of 53, Resurrection 24 of 43, Revive 23
of 48, Rebirth 7 of 11.

The combat log shows both ways. In the first run, with the talent at rank 5, the rogue logs no proc in 1,951 landed
swings until the rogue dies at 20:12 and returns from the graveyard, then 19 in 192. In the third run, at rank 1, it
logs 6 in 422 until a player resurrects the rogue at 22:48, then none in 480. Every main-hand swing within 400 ms of the
one before follows an extra attack line, so no extra attack went unlogged.

Not measured: what sets a resurrection that switches the talent off apart from one that does not, and what turns it
back on.
