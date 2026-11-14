# EV-0017: Seal of Light 15 a minute and Judgement of Light 50%, level 30

- Claim: Seal of Light Rank 1 procs 15 times a minute on the weapon's speed and Judgement of Light Rank 1 heals on a
  flat 50% of landed hits.
- Sources: combat logs of a level 30 paladin on two days, with weapons of 1.6 s and 3.7 s
- Game build: not recorded
- Date: 2026-10-14 and 2026-10-15
- Pull request: #449

Each heal was matched to the paladin's melee swing within 0.3 s, inside the seal's aura window or against a target
carrying the judgement, and the swings were binned by weapon speed from the swing interval.

| Proc               | 1.6 s weapon                 | 3.7 s weapon               | Reading                                                                        |
| ------------------ | ---------------------------- | -------------------------- | ------------------------------------------------------------------------------ |
| Seal of Light      | 463 of 1,101 = 42.1% (± 2.9) | 509 of 554 = 91.9% (± 2.3) | 15 a minute predicts 40% and 92.5%; 14 (86% at 3.7 s) and 16 (99%) fit neither |
| Judgement of Light | 534 of 1,090 = 49.0% (± 3.0) | 256 of 537 = 47.7% (± 4.2) | speed independent, so a flat chance; 825 of 1,693 = 48.7% (± 2.4) overall      |

The judgement's counts use the 40 s debuff as its window. Neither fired on a dodge, a parry or a miss (0 of 330), crits
procced at the same rate as hits, and consecutive Seal of Light heals land 1.6 s apart with the fast weapon, so there is
no internal cooldown. 16 procs a minute fits the 1.6 s log; only the 3.7 s log rules it out.
