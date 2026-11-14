# EV-0011: Searing Totem waits after each cast

- Claim: the totem waits 0.225 s after each cast ends before it starts the next while it keeps its target, one attack every 2.425 s.
- Sources: combat log with StatLedger, a level 60 Dwarf Enhancement shaman (Un'Goro and Blackrock Spire), evening log, and the morning logs of #948 (Blackrock Spire and Dire Maul)
- Game build: 1.60.1
- Date: 2026-11-11
- Pull request: #949

The attack's cast takes 2.2 s (2.204 s in the log). The wait from a cast's end to the next cast's start reads 0.224 s on
164 samples in the evening log and 0.222 s on 286 samples in the morning logs, so a totem that keeps its target attacks
every 2.425 to 2.428 s. An earlier reading of 2.472 s (success to success, 2.472 ± 0.011 s, 256 samples) was a mean that
counted the longer gaps where the totem changes target, which a single living target never has. The first attack lands
at the end of the first cast, so a full 55 s drop holds 22 attacks in the earlier reading.
