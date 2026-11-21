# EV-0009: Maelstrom Weapon does not proc off Windfury Totem hits

- Claim: a Windfury Totem hit adds no Maelstrom Weapon stack, while white swings still do.
- Sources: combat log, Enhancement shamans with Maelstrom Weapon 5/5 (rank 1 and 2 totem, then rank 3 totem), and a census of Warcraft Logs reports uploaded after the 2026-11-14 hotfix; report codes not recorded
- Game build: after push 114060
- Date: 2026-11-13 to 2026-11-14
- Pull request: #1010

Landed Windfury Totem hits that gained a stack, against the white swing rate per landed hit. Gains per episode with and
without a Windfury Totem hit match (white swing plus Windfury Totem hit 0.48 against white swing alone 0.47), and no
gain logs after a Windfury Totem hit. The census of reports after the hotfix shows no gain after a Windfury Totem hit
either. The white rates fit the sim's 10 procs per minute.

| Log                                                                    | Windfury Totem hit | White swing                       |
| ---------------------------------------------------------------------- | ------------------ | --------------------------------- |
| Rank 1 and 2 totem, 2026-11-13 and 2026-11-14, 3.5 s then 3.4 s weapon | 0 of 140 landed    | 0.590 [0.556, 0.624] a landed hit |
| Rank 3 totem, 2026-11-13, 3.4 s weapon                                 | 0 of 121 landed    | 0.598 [0.560, 0.636]              |

The two logs together give 0 of 261 landed Windfury Totem hits.
