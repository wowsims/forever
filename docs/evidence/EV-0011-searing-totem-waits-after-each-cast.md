# EV-0011: Searing Totem waits after each cast

- Claim: the totem waits 0.225 s after each cast ends before it starts the next while it keeps its target, one attack every 2.425 s.
- Sources: combat log with StatLedger, a level 60 Dwarf Enhancement shaman (Un'Goro and Blackrock Spire), evening log, and the morning logs of #948 (Blackrock Spire and Dire Maul) as this pull request recounts them
- Game build: 1.60.1
- Date: 2026-11-11
- Pull request: #949

The attack's cast takes 2.2 s. The wait from a cast's end to the next cast's start reads 0.224 s on 164 samples in the
evening log and 0.222 s on 286 samples in the morning logs, so a totem that keeps its target attacks every 2.425 to 2.428
s. The 2.472 s of #948 (EV-0010) was a mean that counted the longer gaps where the totem changes target, which a single
living target never has. The first attack lands at the end of the first cast.
