# EV-0002: Windfury Totem hits after the 2026-11-14 hotfix

- Claim: after the 2026-11-14 02:00 UTC hotfix a Windfury Totem hit takes armor and the 100 ms cooldown, and
  extra attack procs and the Seals of Command, Fury, Righteousness and Wisdom proc off it; an Execute rolls the totem from
  its cast attack and again from its landed damage; Sweeping Strikes' copied hit 12723 never rolls it and its normalized hit
  26654 rolls it once.
- Sources: Warcraft Logs reports uploaded after the 2026-11-14 02:00 UTC hotfix; report codes not recorded
- Game build: not recorded
- Date: 2026-11-14
- Pull request: #1012

Only clean Windfury hits count: the parent hit has no grant of its own, and nothing else happens within 150 ms. Each
child is counted in its own measured lag window. The same window on parents with no Windfury hit expects 0.3 grants, and
parents that cannot proc the item themselves still show grants after their Windfury hit: Thrash Blade 5 of 91 against 1
of 2,993, Hand of Justice 2 of 84 against 0 of 1,867. One protection warrior's raw logs from before the hotfix read
Thrash Blade 1 of 264; the Windfury hit rolls weapon procs before the hotfix as well, from when it logs in its own
damage line (about 2026-11-13 18:10 UTC; EV-0128).

The extra attack procs below are counted per clean Windfury hit against per white swing. The white rates of Thrash Blade
and Hack and Slash mix both hands' swings and drop the swings whose granted swing was avoided, so they read low; per
main-hand hit the Windfury hit procs them at the white chance (EV-0128).

| Proc                 | Off the Windfury hit | Rate | White rate                            |
| -------------------- | -------------------- | ---- | ------------------------------------- |
| Hack and Slash       | 60 of 2,208          | 2.7% | 1.5%                                  |
| Thrash Blade         | 30 of 1,003          | 3.0% | 1.3%, or 2.5% with no off-hand weapon |
| Sword Specialization | 8 of 198             | 4.0% | 4.8%                                  |
| Hand of Justice      | 4 of 272             | 1.5% | 1.8%                                  |

Paladin seals are counted the same way:

| Seal                  | Off the Windfury hit | White swings   |
| --------------------- | -------------------- | -------------- |
| Seal of Command       | 29 of 136            | 2,055 of 7,096 |
| Seal of Fury          | 88 of 97             | 5,036 of 5,657 |
| Seal of Righteousness | 32 of 38             | 1,503 of 1,948 |
| Seal of Wisdom        | 23 of 47             | 1,173 of 2,721 |

Seal of Command also read 24 of 101 when the parent hit had not procced it. Seal of Light was not measured;
it and Seal of Wisdom hear melee abilities through their rows, the Windfury hit among them. Echo of Command off the
Totem hit read 0 of 64 (white swings 2,663 of 3,921). A Windfury hit, then Seal of Command, then a Windfury hit read 0
of 29, which the 100 ms cooldown blocks.

Execute and Sweeping Strikes, per landed or cast hit:

| Reading                             | Measured                                           | Expected                              |
| ----------------------------------- | -------------------------------------------------- | ------------------------------------- |
| Windfury hit after a landed Execute | 267 of 846 (31.6%)                                 | the cast and its damage each roll     |
| Windfury hit after a dodged Execute | 35 of 185 (18.9%), the cast logging its own attack |                                       |
| Sweeping Strikes' copied hit 12723  | 0 of 490                                           | never                                 |
| Whirlwinds with a 26654, one struck | 63 of 119                                          | 44.8 if it never rolled, 56.5 if once |

Rank 3 adds 246 attack power and ranks 1 and 2 add none, as EV-0008 reads it; the driver's comment once gave about 248.

The shares of the triggered hits that carry Not a Proc are recorded in the comments at the driver only, with no report or
count: Windfury Weapon hits 28%, Seal of Command 14%, Judgement of Command 22% and Retaliation 17%.

- The Windfury hit takes armor as a white hit does, rolls the weapon special's table of miss, dodge, parry and crit, and
  never glances.
