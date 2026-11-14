# EV-0010: Flurry charges spent by Windfury strikes, Stormstrike and spells

- Claim: a Windfury Weapon strike spends a Flurry charge and so does Windfury Totem's strike 10610; Stormstrike and spells spend none.
- Sources: combat log, two level 60 Enhancement shamans (Blackrock Spire and Dire Maul)
- Game build: 1.60.1
- Date: 2026-11-11
- Pull request: #948

A white swing spends a charge whatever its outcome (landed 222 of 229, dodged or parried 24 of 27), and so does a
Windfury Weapon strike (14 of 14 clean cases). Stormstrike spends none (0 of 98), nor do spells. At most one charge goes
per 0.5 s, the 16257 cooldown. A hit in the same instant as a crit's gain spends none, so a crit followed by Windfury
Weapon's strikes leaves three charges, as in the log.

A dodged or parried Windfury Weapon strike also spends in the sim, following the white swing rule; the logs have no clean
case of it. Windfury Totem's strike 10610 spending a charge is read from the same logs; its count is not recorded.
