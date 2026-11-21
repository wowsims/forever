# EV-0010: Flurry charges spent by Windfury hits, Stormstrike and spells, and Searing Totem's first attack

- Claim: a Windfury Weapon hit spends a Flurry charge and so does a Windfury Totem hit (10610); Stormstrike and spells spend none; Searing Totem's first attack lands at the end of its first cast.
- Sources: combat log, two level 60 Enhancement shamans (Blackrock Spire and Dire Maul)
- Game build: 1.60.1
- Date: 2026-11-11
- Pull request: #948

A white swing spends a charge whatever its outcome (landed 222 of 229, dodged or parried 24 of 27), and so does a
Windfury Weapon hit (14 of 14 clean cases). Stormstrike spends none (0 of 98), nor do spells. At most one charge goes
per 0.5 s, the 16257 cooldown. A hit in the same instant as a crit's gain spends none, so a crit followed by Windfury
Weapon hits leaves three charges, as in the log.

A dodged or parried Windfury Weapon hit also spends in the sim, following the white swing rule; the logs have no clean
case of it. A Windfury Totem hit (10610) spending a charge is read from the same logs; its count is not recorded.

Searing Totem's first attack lands at the end of its 2.2 s cast; the attack's cast takes 2.204 s in the log, and
from success to success the attacks read 2.472 ± 0.011 s (n 256), so a full 55 s drop holds 22 attacks. #949 recounts
the wait between casts on a totem that keeps its target (EV-0011).
