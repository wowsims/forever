# EV-0047: Shield Block's last charge and its aura's removal, and shield spikes proc on every block

- Claim: Shield Block's aura goes a median 1 ms after its last charge, and the extra blocks after that charge land 0 to
  2 ms after it; Mithril Shield Spike and Aegis of the Scarlet Commander's Scarlet Retribution proc on nearly every
  block with no internal cooldown.
- Sources: combat logs of a protection warrior at levels 41 to 43 with a StatLedger and a ThreatLog; report codes not
  recorded
- Game build: 1.60.1.70862
- Date: 2026-11-08 to 2026-11-09
- Pull request: #919

## Shield Block's last charge

| Reading                              | Measured                                                                           |
| ------------------------------------ | ---------------------------------------------------------------------------------- |
| Extra blocks after the last charge   | 19 of 20 came 0 to 2 ms after it, the other at 24 ms (the aura's removal at 25 ms) |
| Aura removal                         | a median 1 ms after the last charge                                                |
| Swings logged before the removal     | blocked 16 of 16                                                                   |
| Swings within 5 ms after the removal | blocked 5 of 15                                                                    |

The counts of spent auras that took extra blocks, and the batch rule the sim follows, are #914's (EV-0004).

## Shield spikes

Mithril Shield Spike and Aegis of the Scarlet Commander's Scarlet Retribution dealt their damage on nearly every block.
The rows state 5 s in RecoveryTime, which is the cooldown on casting the aura, a cast an item never makes.

| Reading                  | Mithril Shield Spike | Aegis of the Scarlet Commander |
| ------------------------ | -------------------- | ------------------------------ |
| Full blocks that proc    | 1064 of 1069         | 812 of 812                     |
| Partial blocks that proc | 510 of 512           | 474 of 475                     |

Of 1658 spike gaps, 1178 were under 5 s and 67 under 10 ms. In game the spike damage is logged 90 to 210 ms after the
block.
