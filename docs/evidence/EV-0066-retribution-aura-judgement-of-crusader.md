# EV-0066: Retribution Aura takes no Judgement of the Crusader bonus

- Claim: a rank 1 Retribution Aura hit on a beast under the paladin's own Judgement of the Crusader takes no share of the
  judgement; its flat 5 takes the racial's 5% alone.
- Sources: combat logs of a Dwarf Retribution paladin at levels 1 to 19; report codes not recorded
- Game build: 1.60.1.70790
- Date: 2026-11-05
- Pull request: #872

9 of 9 rank 1 Retribution Aura (7294) hits on Rabid Thistle Bears under the paladin's own Judgement of the Crusader
(20188, rank 2, +35 Holy damage taken) dealt 5 or 6. That is the flat 5 at a Dwarf's 5% against beasts, with no share of
the judgement. The debuff was up for every hit (applied and refresh events from 20:41:40 to 20:42:33). The other 700 or
so hits were 5 on non-beasts and 5.25 on average against beasts. The sim added 0.06 x the target's flat school bonus to
every player cast shield hit, about +2.1 per hit at rank 2. The paladin had 0 spell power. Only Retribution Aura under
the judgement is verified: Thorns is Nature and never met a Holy bonus, and a Fire item shield against the mage's
Frostfire debuff is untested.
