# EV-0083: Demonic Brand lands with 6 charges in Searing Pain's batch

- Claim: Demonic Brand lands with 6 charges at 3/3, though the tooltip reads 3, and in the same batch as Searing Pain's hit.
- Sources: Warcraft Logs, Fresh, report rCv9pnXYqHaQxkG6 (Stockades, solo, Demonic Brand 3/3, trait entry 130626 at rank 3)
- Game build: not recorded
- Date: not recorded
- Pull request: #414

| Searing Pain hits | Brand at the same ms | Brand 1 ms earlier | No brand event |
| ----------------- | -------------------- | ------------------ | -------------- |
| 75                | 68                   | 3                  | 4              |

- The brand's apply or refresh event is logged before Searing Pain's damage event at the same timestamp.
- The 4 hits without a brand event are each the first event of their fight (3, 8, 12, 15): the application sits just before the fight start.
- The brand lands with 6 charges (applydebuffstack 6, consumed 5 to 1), the ladder's value at 3/3.
