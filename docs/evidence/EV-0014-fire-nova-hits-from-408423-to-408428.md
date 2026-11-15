# EV-0014: Fire Nova's hits come from rows 408423 to 408428

- Claim: Fire Nova Ranks 1 and 2 hit as 408423 and 408424, not as the rows their tooltips read (8349 and 8502).
- Sources: combat log, raw logs and Warcraft Logs reports of Fire Nova Ranks 1 and 2; report codes not recorded
- Game build: not recorded
- Date: 2026-10-07
- Pull request: #280

Rank 2's 408424 (102 base points plus 1.6 a level, spell power share 0.214, variance 0.128) explains 201 of 202 logged
hits, where the old 8502 explains 177. Players at their rank's level cap rule out 8502. Ranks 3 to 5 are not measured;
they take the rows for levels 32, 42 and 52 as ranks 1 and 2 take theirs.

| Cast   | Rank | Hit row | Row level | Tooltip row |
| ------ | ---- | ------- | --------- | ----------- |
| 408341 | 1    | 408423  | 12        | 8349        |
| 408342 | 2    | 408424  | 22        | 8502        |
| 408343 | 3    | 408426  | 32        | 8503        |
| 408344 | 4    | 408427  | 42        | 11306       |
| 408345 | 5    | 408428  | 52        | 11307       |

For reference, at level 60 Rank 5's tooltip row 11307 reads 419 + 3.4 a level over 5 levels with a 0.143 spell power
share (436 + 0.143 spell power), and the logged row 408428 reads 403 + 3.4 a level over 5 levels with 0.214 (420 + 0.214
spell power). The two agree at 225 spell power.
