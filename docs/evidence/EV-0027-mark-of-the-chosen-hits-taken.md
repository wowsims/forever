# EV-0027: Mark of the Chosen procs on hits taken at about the row's 2%

- Claim: Mark of the Chosen (item 17774, trigger 21969) procs only on the melee, ranged and harmful spell hits taken, at about the row's 2%; no helpful spell taken fired it.
- Sources: combat logs of a druid at levels 55 to 56, the session of EV-0036
- Game build: 1.60.1.70888
- Date: 2026-11-12
- Pull request: #987

The row's ProcTypeMask also names helpful spells taken, but every proc in the logs followed a hit taken.

| Reading                             | Measured                                                           |
| ----------------------------------- | ------------------------------------------------------------------ |
| Procs on eligible landed hits taken | 5 procs on 392 landed hits taken, 1.28%, 95% interval 0.55 to 2.95 |
| Procs off a helpful spell taken     | none                                                               |
