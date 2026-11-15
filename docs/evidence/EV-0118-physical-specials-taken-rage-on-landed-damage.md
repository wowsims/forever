# EV-0118: A physical hit taken that is not a melee auto attack grants twice the melee rule's rage on the damage that lands

- Claim: a physical direct hit taken that is not a melee auto attack grants twice the melee rule's rage, on the damage
  left after armor and the block.
- Sources: combat logs of a protection warrior, levels 19 to 39, hits from rage-using attackers of levels 19 to 44
  outside dungeons, Defensive Stance divided out
- Game build: 1.60.1.70790, 1.60.1.70862, 1.60.1.70888 and 1.60.1.70938; hotfixes as current at each login, from the
  2026-11-05 07:01 UTC hotfix (wowsims/gamedata #74) through the 2026-11-14 05:02 UTC hotfix (#294); #295 to #297
  landed during the last session without a relog. The rule reads the same in every log.
- Date: 2026-11-05 to 2026-11-14
- Pull request: #TBD

The rage of a hit is the tenths of rage on the hit's own line minus the warrior's previous line. A pair counts only when
no other attack lands within 20 ms, no absorb takes part of the hit, rage stays below the cap and nothing spends or loses
rage between the lines; hits taken under Shield Wall are left out. A melee auto attack is a swing line; every other
physical direct hit (Shoot, Throw, Shield Slam, Backstab, Arcing Smash and the like) is a non-swing hit.

Each hit is read two ways. On the base, x is the log's base amount plus 0.5, times 2 on a crit and 1.5 on a crushing
blow, times the stance's damage taken factor (Defensive Stance's 0.9), over the rage conversion value c at the
attacker's level: the damage before armor and the block. On the landed amount, x is the logged amount plus 0.5 over c:
the damage after armor, the stance and the block, with no other term. The factor is k = Σ observed / Σ 10x, scaled to
the sim's assumed armor at each attacker level and quoted at levels 19 to 38, with a 95% interval from a bootstrap over
attackers. The melee rule is 2.5 × (1 - 0.218) = 1.955, so twice it is 3.910. The warrior's armor reduction DR is armor
/ (armor + 400 + 85 × attacker level).

| Hits                                         | n     | Attackers | k on the base        | k on the landed amount |
| -------------------------------------------- | ----- | --------- | -------------------- | ---------------------- |
| Non-swing physical, unblocked                | 166   | 34        | 2.301 [2.249, 2.358] | 3.832 [3.782, 3.874]   |
| Non-swing physical, unblocked, DR 0 to 0.4   | 42    |           | 2.452 [2.361, 2.578] | 3.812 [3.784, 3.840]   |
| Non-swing physical, unblocked, DR 0.4 to 0.6 | 124   |           | 2.248 [2.184, 2.324] | 3.840 [3.768, 3.894]   |
| Non-swing physical, partly blocked           | 33    | 13        | 0.932 [0.657, 1.230] | 3.659 [3.404, 3.854]   |
| Melee swings, unblocked                      | 5,974 | 274       | 1.954 [1.947, 1.962] | 3.145 [3.106, 3.180]   |
| Melee swings, unblocked, DR 0 to 0.4         | 2,920 |           | 1.950 [1.943, 1.958] | 2.974 [2.940, 3.009]   |
| Melee swings, unblocked, DR 0.4 to 0.6       | 3,054 |           | 1.958 [1.945, 1.971] | 3.329 [3.300, 3.357]   |
| Melee swings, partly blocked                 | 1,652 | 219       | 1.965 [1.951, 1.983] | 7.408 [6.879, 7.961]   |

A non-swing hit's rage follows the damage that lands. On the landed amount its factor stays flat across the warrior's
armor, 2% under 3.910, and a partial block keeps it there; on the base it falls as the warrior's armor rises and drops
to 0.93 when part of the hit is blocked. A melee swing shows the reverse: its base reading is flat at the melee rule,
blocked or not, and its landed reading moves with armor and the block.

On the base, the unblocked non-swing hits read Shoot 2.27 (67), Throw 2.35 (38), Shield Slam 2.32 (12), Backstab 2.31
(11) and Arcing Smash 2.18 (12). The partly blocked ones were Throw 9, Shield Slam 9, Shoot 7, Arcing Smash 4, Mortal
Strike 2 and Heroic Strike 1.

No absorbed hit is in the sample, and no attacker that uses mana.
