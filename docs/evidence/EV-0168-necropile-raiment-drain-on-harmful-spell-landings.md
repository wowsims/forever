# EV-0168: Necropile Raiment's Necropile Drain procs on about one harmful spell landing in seven

- Claim: The Necropile Raiment 3-piece bonus (1299737) casts Necropile Drain (1299735) on about 13.9% of harmful spell
  landings that are not ticks, and never on a killing blow. By spell it reads 11 to 19%, and the spread is not
  explained: it does not follow cast time (an instant Shadow Word: Death reads 16%, a 1.5 s Mind Blast 19%, a 2.5 s
  Incinerate 13%). Four priests and one warlock give it, so the rate is observed.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every fight with 3 or more pieces worn in:
  3aXTp97wjCWNxMhd, 87GwCAY4MtTg6aLB, C18pcNtwj4aRDbQk, VwC7K3aq4YzmMjQR, gBCzpA8b2NfwcWDT, mJn3ZTBcAX6MawNy,
  r3dNjhavPz8LFVWY, vLp6HPt7hg3QAb28
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-15 to 2026-11-20
- Pull request: #1146

## Method

5 players (4 priests in 7 runs, 1 warlock), 3 or 4 pieces worn per fight's CombatantInfo. Only events the wearer is
the source of count. A proc is a damage event of 1299735 (each comes with its heal). It lands on the spell's target 0
to 50 ms after the spell's damage or debuff, so a landing gets credit for a proc on the same target (id and instance)
from 2 ms before to 80 ms after it. Landings of one spell on one target within 5 ms count once. The drain lands on the
target, so killing blows are left out; they never proc it (0 of 197 Shadow Word: Death and 0 of 50 Mind Blast killing
blows).

Counted: the damage of a direct spell and the debuff of a periodic or channelled one (Shadow Word: Pain, Mind Flay,
Immolate). Devouring Plague applies 1,235 debuffs from 428 casts, so it counts per cast, with a proc on the cast's
target up to 200 ms after it.

## Rate

| Spell              |  Landings | Procs |                Chance |
| ------------------ | --------: | ----: | --------------------: |
| Mind Flay          |     1,999 |   272 |                13.61% |
| Shadow Word: Pain  |     1,858 |   222 |                11.95% |
| Mind Blast         |       831 |   155 |                18.65% |
| Devouring Plague   | 428 casts |    60 |                14.02% |
| Shadow Word: Death |       317 |    52 |                16.40% |
| Incinerate         |       136 |    18 |                13.24% |
| Immolate           |        77 |    10 |                12.99% |
| Conflagrate        |        75 |    10 |                13.33% |
| Shadowburn         |        72 |     8 |                11.11% |
| All of these       |     5,793 |   807 | 13.93% [12.97, 14.89] |

The interval is Poisson on the count, since four wearers carry it. Per wearer: 14.2%, 13.0% and 13.7% (priests) and
12.8% (the warlock); the fifth priest has 7 landings and no proc. Mind Blast (155 of 831) and Shadow Word: Pain (222 of
1,858) do not overlap, so one flat chance does not fit every spell.

These never proc it: Rain of Fire damage (0 of 421, channelled ticks), Bane of Havoc (0 of 306), Shadowguard (0 of
259), Mind Soothe (0 of 70), wand Shoot and Melee, and periodic ticks. Vampiric Embrace reads 3 of 183.
