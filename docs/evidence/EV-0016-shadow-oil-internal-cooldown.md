# EV-0016: Shadow Oil's proc has a 10 second internal cooldown

- Claim: Shadow Oil's proc 350026, cast off the enchanted weapon's hits at 15%, has a 10 s internal cooldown although the game files state no ProcCategoryRecovery on it.
- Sources: Warcraft Logs Y4mTwRvPtqAn36CB, fights 1 to 9
- Game build: not recorded
- Date: 2026-10-07
- Pull request: #268

| Window since the last proc | Landed hits | Procs |
| -------------------------- | ----------- | ----- |
| Within 10 s                | 425         | 0     |
| After 10 s                 | 633         | 80    |

Procs sit at least 10.26 s apart, and no hit from 9.98 s on the clock procs.
