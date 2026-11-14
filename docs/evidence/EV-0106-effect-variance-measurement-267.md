# EV-0106: EffectVariance is the full roll width

- Claim: EffectVariance is the full width of a roll: spell-school hits and heals roll the whole amount, power included, by ±V/2, and Revenge Rank 1, a physical hit, fits a roll of the base by ±V before attack power.
- Sources: Forever combat logs and Warcraft Logs heal groups; report codes not recorded
- Game build: not recorded
- Date: not recorded
- Pull request: #267

Each row compares the rule against the logged hit or heal width; "whole amount ± V/2" is the rule the pull request adopts.

| Spell                                            | Data                           | Whole amount ± V/2 | Base ± V/2 | Base ± V  |
| ------------------------------------------------ | ------------------------------ | ------------------ | ---------- | --------- |
| Wrath rank 5 (45, V 0.13), 40 spell power        | 1,793 non-crit hits: 63–72     | 63.4–72.3          | 64.9–70.8  | 62.0–73.7 |
| Flash of Light rank 1, one paladin, 4 groups     | 1,371 heals, 13–14 points wide | 13.2–13.3          | 10.8–10.9  | 21.5–21.8 |
| Revenge rank 1 (22, V 0.154), fixed attack power | 186 hits, 5–6 wide             | about 9.5          | about 3.4  | 6.1       |

Wrath and Flash of Light fit the whole amount ± V/2; Revenge, a physical hit, fits the base ± V. Moonfire's direct hit and the other heals point the same way. Over every Warcraft Logs heal group the log-likelihoods are whole −316, base-only half −429 and base-only full −469.

The heals rest on the Flash of Light groups and a few druid heals, and Holy Strike is unmeasured. The physical roll of
Revenge Rank 1 is read again by #906 (EV-0045), which rejects the base ± V and fits the base by half the Variance.
