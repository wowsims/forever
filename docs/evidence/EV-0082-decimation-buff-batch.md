# EV-0082: Decimation's buff comes in the cast's batch

- Claim: in execute, a Shadow Bolt or Searing Pain that hits grants Decimation in the same batch as the cast, and a miss or a full resist grants nothing.
- Sources: combat logs; report codes not recorded
- Game build: not recorded
- Date: not recorded
- Pull request: #349

Searing Pain and Shadow Bolt casts that hit gain Decimation in the same millisecond as the cast line (Shadow Bolt 1 ms after, about 1.5 s before its damage). A cast that misses gains nothing. That a full resist grants nothing is an in-game observation with no count recorded.

| Cast                                   | Buff                                   |
| -------------------------------------- | -------------------------------------- |
| Hit or crit, partially resisted or not | granted at the cast, in the same batch |
| Miss                                   | none                                   |
| Full resist                            | none                                   |
