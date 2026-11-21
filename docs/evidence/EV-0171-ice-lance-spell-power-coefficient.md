# EV-0171: Ice Lance's spell power coefficient

- Claim: Ice Lance Rank 6 (1240047) deals its game files damage (145 + 0.9 per level past 56, 136 to 161 at level 60)
  plus 0.11 of spell power. Frozen targets take 4 times the damage, and Ice Barrier (+5%), Frigid Power (+15%) and Ice
  Shards' crit damage (×2) apply to it as their rows state.
- Sources: combat logs of one level 60 Frost mage with StatLedger; four sessions geared, 2026-11-19 22:33 to 2026-11-21,
  and one at 0 spell power, 2026-11-21 08:40 to 08:45
- Game build: 1.60.1.70959 (geared) and 1.60.1.71037 (0 spell power), each on the latest hotfix at the time, relogged
- Date: 2026-11-19 to 2026-11-21
- Pull request: not recorded

## Method

Every Ice Lance cast by the mage is paired with its damage event on the same target. Spell power is the caster's on the
cast line. The damage is `baseAmount`, the amount before the crit and before any partial resist. Fingers of Frost counts as frozen when one of its charges is removed within 50 ms of the cast; a
Frost Nova, Freeze or Frostbite on the target counts too. Ice Barrier and Frigid Power are read from the mage's auras at
the cast. Each hit is divided by the frozen, Ice Barrier and Frigid Power multipliers; hits on targets above level 60
are left out, since those take a level factor.

## Counts

| Set                     | Hits | Spell power | Per hit, before modifiers | Range     |
| ----------------------- | ---: | ----------: | ------------------------- | --------- |
| 0 spell power, no buffs |   28 |           0 | 147.1 ± 1.4               | 136 – 159 |
| geared, no Emblem buff  |  255 |   203 – 234 | 173.3 (at 228.5 mean)     | 161 – 186 |

The 0 spell power set fits the game files band (148.6 mean). The geared set is that band moved up by about 25:
(173.3 − 147.1) / 228.5 = 0.115 ± 0.007, and (173.3 − 148.6) / 228.5 = 0.108 against the game files mean. 1/7 sits
about four standard errors above both.

The modifiers hold within 1% across the geared set: 36 unfrozen hits with neither buff read 171.9, 696.6 frozen (×4.05), 798.6 frozen
with Frigid Power and 836.6 frozen with both. A crit deals 2.00 times `baseAmount` (134 crits, 2.00 to 2.01). Frost Nova
without Fingers of Frost reads ×4 on its three hits.

## Open

Draconic Infused Emblem's +35 spell damage (1318930) raises the cast line's spell power by 31.6 on 53 hits, and Ice
Lance gains +0.2 ± 1.0 from it, where 0.11 gives +3.5. Frostbolt Rank 10 gains +25.7 from the same buff on 286 hits,
which its 0.814 predicts (27.3). Why the buff misses Ice Lance is not known; the sim applies it.
