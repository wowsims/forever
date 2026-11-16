# EV-0124: Zum'rah's Vexing Cane's zombie rate and damage

- Claim: Ward of Zum'rah (1294255) summons the Zombie Troll of Zum'rah (1294258) at about 0.7% per hostile target a
  harmful spell hits, behind its 60 s cooldown, and each zombie deals about 150 damage, so the pet adds below 1 DPS.
- Sources: public reports on forever.warcraftlogs.com, zone 3000, with the fights whose combatant info lists the cane on
  the wearer. Druid: 2KkaByn6H9mzb4Yq fight 4; 37KAB4WjzJZYdvaT fight 1; 4AGtLBYXcqbQhFZj fights 1, 12, 16;
  4AWXbZk1HFzxQ6m3 fight 1; XFdcWaPL8MthYH6A fight 1; yfPHWFcbnaK9X8VD fight 1. Mage: 18Nf9mQz4RW2aYc6 fights 1, 3;
  4RcnAhpLZ96KfNDP fights 1, 3, 13; 6xHzpKd8wCna37Lb fight 1; 7p2wRkayKP8nNTfW fight 1; 8vJjYft73m9Chzkc fight 14;
  9Tqz6MDALRw3x2N4 fight 1; B4ak2F9hrztgQJH6 fight 1; BymPn39CrpcWbxv2 fights 1, 8; CBpMcGAV4PqFNa6R fight 1;
  CGF4WpJPjbhvKg9y fight 1; FKJPgVh9c36zDNpT fight 1; G2QDqn9gc1mRXByC fight 1; H7f28bdvhrFpwVBM fights 1, 4;
  LqYCDQ1nh9M3VFrw fight 1; NHa2RqtjxAvPFwLg fights 1, 8; NXMZV9xjvWpB7GtP fights 20, 23; PWx7LHKbCwrQTpRa fight 5;
  RHr283cN6AF4Za7D fight 1; RVMTj7BxchwyA4dJ fight 1; TM6tCLFdA83Pbczr fight 1; VyXwH1Wfq2bc3xZ6 fight 1;
  XxqdrnVN9DCZcL8T fights 1, 7; Z7qF4j3hazBTxHcw fights 8, 10; b4hzxLyCZXMHW8gP fight 6; bP6qxJTBpdHf47MA fight 1;
  c9q8DZt4pvzjLhYn fight 1; cx8pPRvN9Ta2zHg1 fight 1; dxaHvw9j2cpZL7TQ fight 1; hR6pzT1HkaLym2nd fight 1;
  n7913jW8dcfLvYP4 fights 1, 3, 5, 13; nAVx6ZgkQywNd3hc fight 5; nTXd4129Az6GPfjV fights 8, 10; nfLmGxNDqJpyXB6k fights
  8, 10, 16, 24; nrKZtkQhgbBYf7R6 fights 1, 6, 8; rjkhRfxLVmYvDB24 fight 1; tmVg9WB2hv3DT8AL fight 23; vRMqfB7XnCHkZg84
  fight 1; vZqVmLYQa6jwP3A9 fight 1. Priest: F4HJmAWrChz2qv1Q fight 1; K9Nzbf8cGrmywHV4 fight 1; KW4Ytd2VPbHLDJaQ fights
  1, 3; RkYCfyngJrLvwxDc fight 4; c6DWA1CV2tdmRFqL fights 1, 6, 14; czFAk3dtDYqgapVK fights 1, 24; m3Fjgrt124A7advV
  fight 1; xQqNBFtaHwX71jKY fight 1. Shaman: 2KkaByn6H9mzb4Yq fights 1, 4. Warlock: 1WHqx7bPa8mkGydJ fight 1;
  3ytgzKdVNcxfLvAC fight 1; 9cQz2ymYMgfL7TwK fight 132; F4HJmAWrChz2qv1Q fight 1; FMaAjHcwCQLkpV32 fight 1;
  GTJw7YXampLrWcgM fights 1, 2; KPkDVWCvjLw2BqNR fight 1; L7VBmjgyHRCzqYa2 fight 1; LpPyXn6tCJRGMV3d fights 1, 7;
  MV3AfgqjPFXGzB1y fight 1; RQHKbN1n82V7TAGc fights 1, 6; VtGAb9HNFMYhRjXq fight 1; WrBf9NpqGhTcm21x fight 1;
  ZQ2FqgD1dMH9kGcp fight 1; bRxwhZJ37MyvPtNL fight 1; f43Q2HLNTj8CxbXZ fight 1; gBDMLv8J1K7nqkpd fight 13;
  jBvJX1Ltky2hz7xm fights 1, 6; kQnFLCz12PyqVjG9 fight 4; kr3AzhZ9FVB1nG2D fight 1; mAwQn61fk4YzcL3P fight 1;
  mnMLPqv49K6RTjHA fight 4; nhYyv7NG3L982RaB fight 1; pxRAZ6j8byJwmMTF fights 1, 4; qMPwcLCNa4f1HVzk fight 1;
  tV6cx9nMN4782FX3 fight 1; thpcGknfvy49gTz8 fight 6; v7HFptdBW8KcmjnN fight 1; vqQgaxHKDw6kh4PX fights 1, 5.
- Game build: 1.60.1.70938; the hotfix state and relogs were not recorded.
- Date: 2026-11-09 to 2026-11-15
- Pull request: #1050

A wearer is a player whose combatant info lists item 18082; 82 wearers in 80 runs had a zombie. The log carries no line
for the summon, so a zombie is seen from its first swing, and a zombie that never swings is not counted: 346 zombies is
a floor. The combat log gives one gear snapshot per run, so a cane swapped in or out during a run is not seen.

The zombie only melees. Its swings come 2.00 s apart, a normal hit deals 39 to 53 before armor (median 46, 5th to 95th
percentile) and 29 after armor (median); 0.4% of hits crit, 16% glance and about 5% miss. Its first and last swings sit
a median 12.5 s apart (longest 20.2 s), and a zombie deals a median 147 damage (mean 151, most 473). Zombies of one
wearer come at least 57.7 s apart, which fits the 60 s cooldown.

Each zombie goes to the wearer's last cast in the 3 s before its first swing, split evenly over the casts in that
window; the median time from that cast to the first swing is 1.0 s. A trial is a cast of a spell that dealt damage to or
put a debuff on a hostile target, outside the cooldown, in a worn fight; wand shots and melee are left out. The first
cast once the cooldown ends procs 1.5% of the time, so the 100 in the row's chance column is not the chance.

| Hostile targets the cast hit | Casts  | Procs | Chance | Per target |
| ---------------------------- | ------ | ----- | ------ | ---------- |
| 1                            | 17,303 | 111.5 | 0.64%  | 0.64%      |
| 2                            | 1,945  | 39.0  | 2.01%  | 1.00%      |
| 3                            | 1,088  | 15.5  | 1.42%  | 0.47%      |
| 4                            | 287    | 12.3  | 4.30%  | 1.07%      |
| 5 or more                    | 52     | 1.7   | 3.21%  | 0.64%      |

The chance per cast grows with the targets hit and stays near 0.7% per target, so the roll looks per target hit. A rate
per minute fits worse: on 15,843 single-target casts the chance by time since the wearer's previous cast is 0.43%,
0.65%, 0.70%, 0.88%, 0.45% and 0.98% for 0-1 s to 5 s and more, and a pure rate per minute (0.16) trails a flat chance
by 12.5 log-likelihood units. By class the chance per cast is 1.07% for mages (182 procs), 0.51% for warlocks (55),
0.64% for druids (17), 1.35% for shamans (6) and 0.27% for priests (21).

Wearers saw 0.17 zombies per minute in which they cast, about 25 damage a minute or 0.4 DPS. A zombie every 60 s would
give 2.5 DPS.
