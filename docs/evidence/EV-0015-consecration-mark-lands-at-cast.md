# EV-0015: Consecration's hidden mark lands at the cast and outlasts the last tick

- Claim: Consecration's hidden mark 1280353 lands at the cast and drops 0.6 s after the ground's last tick; while it is
  up every Holy hit takes Consecrated Ground's flat 10% after the crit.
- Sources: combat log, a level 30 Holy paladin with Consecrated Ground 2/2 and Consecration Rank 2
- Game build: not recorded
- Date: 2026-10-17
- Pull request: #516

The game puts a hidden debuff, spell 1280353 "Consecration", on the target about 2 ms after the cast and removes it
0.6 s after the ground's last tick, or when the target dies. The mark's row has two effects, the bonus tick's dummy with
its 0.042 coefficient and a Holy damage from caster mod with 0 points that the server sets from the talent, and a
600 ms duration row, which is the tail past the last tick.

Each Holy hit's damage multiplier with the mark up, hits without and with it:

| Spell                       | Hits without / with | Multiplier without | Multiplier with |
| --------------------------- | ------------------- | ------------------ | --------------- |
| Judgement of Righteousness  | 34 / 14             | 1.005              | 1.105           |
| Exorcism                    | not recorded        | 1.003              | 1.104           |
| Holy Strike                 | not recorded        | 1.007              | 1.112           |
| Seal of Righteousness procs | not recorded        | 1.016              | 1.114           |

Consecration's own ticks take the bonus from tick 1.
