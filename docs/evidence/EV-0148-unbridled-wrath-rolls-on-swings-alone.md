# EV-0148: Unbridled Wrath rolls on white swings, Windfury Totem's strike, Heroic Strike and Cleave alone

- Claim: Unbridled Wrath 12322 rolls on landed white swings, Windfury Totem's strike, Heroic Strike and each target a
  Cleave lands on, at the 60% its rank 5 row states; Shield Slam, Revenge, Thunder Clap, Retaliation, Victory Rush and
  Shield Bash roll none.
- Sources: combat log, a level 60 Protection warrior with a one-handed sword and a shield (Unbridled Wrath 5/5, Master
  of Defense 2/2, Shield Specialization 5/5, Enrage 5/5, Reckless Swings, 6 pieces of Battlegear of Valor), in groups
  with Windfury Totem at times, in Blackrock Depths, Scholomance, Lower and Upper Blackrock Spire and Stratholme
- Game build: 1.60.1 by the log header; the full build and the hotfix push are not recorded
- Date: 2026-11-17 22:09 to 2026-11-18 07:54 UTC
- Pull request: pending

## Method

Only the stretches inside the five instances count. A hit is one of the warrior's own damage or miss lines: a white
swing, a Windfury Totem strike (10610), or a Heroic Strike, Cleave, Shield Slam, Revenge, Thunder Clap, Retaliation,
Victory Rush or Shield Bash. Landed means a hit, a crit or a glancing blow. A proc is an energize of 12964, and a hit
has one when it logs from 5 ms before the hit to 40 ms after it. A hit is isolated when the warrior logs no other hit
within 100 ms of it. A Cleave's lines on its targets log within 3 ms of each other, nearly all in the same ms, and count
as one cast. The same window shifted 0.3 to 0.9 s later finds a proc after 0.7% of hits, the rate of chance
coincidence.

## Unbridled Wrath

Isolated landed hits:

| Hit                     | With a proc / hits | Share |
| ----------------------- | ------------------ | ----: |
| White swing             | 1,241 / 2,109      | 58.8% |
| Heroic Strike           | 248 / 402          | 61.7% |
| Windfury Totem's strike | 16 / 27            | 59.3% |
| Shield Slam             | 0 / 977            |     0 |
| Revenge                 | 0 / 601            |     0 |
| Thunder Clap            | 0 / 2,238          |     0 |
| Retaliation             | 0 / 49             |     0 |
| Victory Rush            | 0 / 24             |     0 |
| Shield Bash             | 0 / 44             |     0 |

Isolated Cleave casts, by the targets they landed on:

| Targets landed | Casts | Procs | Per cast |
| -------------: | ----: | ----: | -------: |
|              1 |   146 |    87 |     0.60 |
|              2 |   569 |   678 |     1.19 |

Every landed hit, isolated or not:

| Hit           | With a proc / hits | Share |
| ------------- | ------------------ | ----: |
| White swing   | 1,506 / 2,572      | 58.6% |
| Heroic Strike | 303 / 493          | 61.5% |
| Shield Slam   | 28 / 1,119         |  2.5% |
| Revenge       | 16 / 737           |  2.2% |
| Thunder Clap  | 51 / 2,777         |  1.8% |

Of the 95 Shield Slam, Revenge and Thunder Clap hits with a proc, 75 share it with a landed white swing, strike, Heroic
Strike or Cleave whose window holds it too. The other 20 come to 0.43% of those abilities' hits, under the 0.7% the
shifted window finds.

The instances hold 3,123 energizes of 12964 against 5,252 landed hits of the four kinds that roll it: 2,572 white
swings, 513 strikes, 493 Heroic Strikes and 1,674 Cleave target hits, 59.5%. No missed, dodged or parried white swing
has one (0 of 248), and no missed, dodged or parried Heroic Strike (0 of 67). Each of the 77 avoided Cleave lines and 32
avoided strikes that show one shares it with a landed hit of those kinds: the other target's line, or the hit before the
strike.

EV-0138 reads the same on a dual-wielding Fury warrior: Whirlwind 2 of 425, Heroic Strike 14 of 20, Cleave 111 of 165
hit lines. EV-0136 finds a proc after 0.26% of melee abilities alone.

## Master of Defense, Shield Specialization and Shield Spike

The same log's procs on hits taken, counted in the instances:

| Proc                                                | Events | Hits taken                                                                      |
| --------------------------------------------------- | -----: | ------------------------------------------------------------------------------- |
| Master of Defense 2/2, rage                         |  1,518 | 1,521 dodges and parries of swings and physical abilities                       |
| Shield Specialization 5/5, rage                     |  2,145 | 2,108 blocks of swings and abilities, full and partial; ranged attacks left out |
| Shield Spike, Girdle of Extracted Answers (1308777) |  2,086 | 2,100 blocks of swings and physical abilities, full and partial                 |

Each fires on about every hit of its kind. Shield Spike's hits come as little as 0 s apart, 487 gaps under 1 s, though
its row states a 5 s cooldown.
