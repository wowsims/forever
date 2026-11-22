# EV-0010: Flurry charges spent by Windfury hits, Stormstrike and spells, and Searing Totem's first attack

- Claim: a white swing spends a Flurry charge whatever its outcome, and a Windfury Weapon hit spends one, at most one
  per 0.5 s; a Windfury Weapon hit spends one in the instant a white or Stormstrike crit gained the charges too, but
  none in the instant another Windfury Weapon hit's crit gained them; a Windfury Totem hit (10610) spends none, and its
  crit applies or refreshes Flurry; Stormstrike and spells spend none; Searing Totem's first attack lands at the end of
  its first cast.
- Sources: combat log, two level 60 Enhancement shamans (Blackrock Spire and Dire Maul); the Windfury hits name a count
  section and list their own
- Game build: 1.60.1; the Windfury hits have their own
- Date: 2026-11-11; the Windfury hits have their own
- Pull request: #948; the Windfury hits have their own
- Later count: the Windfury hits, §Count of the 2026-11-13 to 2026-11-21 logs

## White swings, Stormstrike and spells

A white swing spends a charge whatever its outcome (landed 222 of 229, dodged or parried 24 of 27). Stormstrike spends
none (0 of 98), nor do spells. At most one charge goes per 0.5 s, the 16257 cooldown.

## Windfury hits

- Count: §Count of the 2026-11-13 to 2026-11-21 logs

A Windfury Weapon hit spends a charge whenever the 0.5 s cooldown allows, also in the instant of the white or
Stormstrike crit that gained the charges; the crit spends none. In the instant a Windfury Weapon hit's own crit gained
or refreshed Flurry, the next Windfury Weapon hit spends none. A Windfury Totem hit spends none, landed or avoided, and
its crit applies Flurry from no charges or refreshes it to three.

A dodged or parried Windfury Weapon hit also spends in the sim, following the white swing rule; the logs have no clean
case of it.

The 2026-11-11 logs and the count section's 2026-11-15 log read a Windfury Totem hit spending a charge (after a
Stormstrike that did not crit, 5 of 5 on 2026-11-15), and its 2026-11-13 log counts Windfury Totem hits among the
spenders. All of them ran before push 114115 (2026-11-16); the logs after it read none.

## Searing Totem's first attack

Searing Totem's first attack lands at the end of its 2.2 s cast; the attack's cast takes 2.204 s in the log, and
from success to success the attacks read 2.472 ± 0.011 s (n 256), so a full 55 s drop holds 22 attacks. #949 recounts
the wait between casts on a totem that keeps its target (EV-0011).

## Count of the 2026-11-13 to 2026-11-21 logs

- Sources: combat logs of one level 60 Dwarf Enhancement shaman with Flurry 2/5, Windfury Weapon and Windfury Totem
  (hit 10610): Lower and Upper Blackrock Spire on 2026-11-19, Alterac Valley on 2026-11-21. The earlier readings: a
  level 60 Dwarf Enhancement shaman with Flurry 3/5 in Stratholme on 2026-11-13, and a level 44 to 49 Enhancement
  shaman with Flurry 5/5 on the night of 2026-11-13 to 2026-11-14 and in Maraudon on 2026-11-15
- Game build: 1.60.1.70959 on 2026-11-19 and 1.60.1.71037 on 2026-11-21, hotfixes up to push 114122; relog not known.
  The earlier readings: 1.60.1 on 2026-11-13 and 2026-11-14, full build not recorded; 1.60.1.70938 on 2026-11-15
- Date: 2026-11-13 to 2026-11-21
- Pull request: not opened yet

A bundle is a white swing or a Stormstrike with the Windfury Weapon and Windfury Totem hits that follow it within 120
ms, a second Windfury Weapon hit joining up to 300 ms after the lead. It owns the Flurry lines from 5 ms before its first
attack to 150 ms after its last, up to the next bundle. The charges before a bundle and the time of the last spend are
read from the log's own Flurry lines; the cooldown counts as ready when 0.5 s have passed since that spend. The Alterac
Valley log is a battleground, so its counts are small.

Windfury Totem hits:

| Case                                                                                                        | 2026-11-19 | 2026-11-21 |
| ----------------------------------------------------------------------------------------------------------- | ---------- | ---------- |
| Stormstrike, no crit, then Windfury Totem hits only, charges up and the cooldown ready: a charge spent      | 0 of 23    | 0 of 3     |
| The same with Windfury Weapon hits only: a charge spent                                                     | 18 of 18   | 1 of 2     |
| The same with no hit after the Stormstrike: a charge spent                                                  | 7 of 170   | 1 of 25    |
| Dodged or parried Windfury Totem hit after a Stormstrike, charges up and the cooldown ready: a charge spent | 0 of 3     | none       |
| White or Stormstrike crit, then Windfury Totem hits only, cooldown ready: a charge spent                    | 0 of 58    | 0 of 4     |
| Windfury Totem hit crit with Flurry down: Flurry applied                                                    | 33 of 33   | 4 of 4     |
| Windfury Totem hit crit with Flurry up: Flurry refreshed                                                    | 28 of 28   | 3 of 3     |

27 Windfury Totem hit crits of 2026-11-19 with another crit within 60 ms are left out of the last two rows.

Bundles with Windfury Weapon hits, by their lead and the charges before it, with the cooldown ready at the lead. After
a white swing that does not crit, the one charge is the swing's and the Windfury Weapon hits fall inside the cooldown it
starts; after a Stormstrike or a crit, it is the first Windfury Weapon hit's.

| Lead                                     | One charge spent, 2026-11-19 | One charge spent, 2026-11-21 |
| ---------------------------------------- | ---------------------------- | ---------------------------- |
| White swing, no crit, charges up         | 82 of 82                     | 9 of 10                      |
| Stormstrike, no crit, charges up         | 24 of 24                     | 1 of 2                       |
| White crit, charges up                   | 17 of 17                     | 1 of 1                       |
| White crit, from no charges              | 15 of 16                     | none                         |
| Stormstrike crit, charges up             | 5 of 5                       | 1 of 2                       |
| Stormstrike crit, from no charges        | 4 of 4                       | 0 of 1                       |
| Stormstrike, no crit, cooldown not ready | none spent, 1 of 1           | none spent, 1 of 1           |

A white or Stormstrike crit with no hit after it ends at three charges 378 of 385 times on 2026-11-19 and 41 of 43 on
2026-11-21. A Windfury Weapon hit that follows, in the same instant, another Windfury Weapon hit whose crit set three
charges spends none with the cooldown ready: three charges after 47 of 47 times on 2026-11-19 (25 of them where that
crit applied Flurry from no charges) and 6 of 6 on 2026-11-21 (4). Where charges were up before the critting hit, its
own spend logs before its refresh. The shortest gap between two spends is 0.615 s on 2026-11-19 and 1.157 s on 2026-11-21.

Earlier readings agree. A level 60 Dwarf Enhancement shaman with Flurry 3/5 and Windfury Totem rank 3 in a Stratholme
combat log of 2026-11-13 spent a charge after a crit that set three charges 66 of 66 times when Windfury Weapon or
Windfury Totem hits followed it, and 3 of 150 times when none did. A level 48 to 49 Enhancement shaman with Flurry 5/5 in
a Maraudon combat log of 2026-11-15 (1.60.1.70938) spent one after a crit from no charges 10 of 10 times when Windfury
Weapon hits alone followed it. The same shaman's log of the night of 2026-11-13 to 2026-11-14, at levels 44 to 45, read
1 of 13; it ran before the 2026-11-14 02:00 UTC hotfix.
