# EV-0103: Shield spikes proc on every block, and a death rolls the row's plain chance

- Claim: block-triggered damage procs such as shield spikes fire on every block with no internal cooldown, and a death rolls the row's plain chance.
- Sources: combat logs of the block-triggered procs, report codes not recorded; for the death roll, nothing recorded
- Game build: not recorded
- Date: not recorded
- Pull request: #919 (shield spikes), #826 (death roll)

The shield spike reading is from #919. Mithril Shield Spike and the Scarlet Commander's Scarlet Retribution dealt their damage on nearly every block. The rows state 5 s in RecoveryTime, which is the cooldown on casting the aura, a cast an item never makes.

| Reading                  | Mithril Shield Spike | Aegis of the Scarlet Commander |
| ------------------------ | -------------------- | ------------------------------ |
| Full blocks that proc    | 1064 of 1069         | 812 of 812                     |
| Partial blocks that proc | 510 of 512           | 474 of 475                     |

Of 1658 spike gaps, 1178 were under 5 s and 67 under 10 ms.

The death roll is from #826, which states only that a rated trigger drops the death callbacks because a death rolls the row's plain chance, which a rated row leaves at 0. That pull request records no count.
