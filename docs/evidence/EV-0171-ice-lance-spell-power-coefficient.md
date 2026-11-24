# EV-0171: Ice Lance's spell power coefficient

- Claim: Ice Lance Rank 6 (1240047) deals its game files damage (145 + 0.9 per level past 56, 136 to 161 at level 60)
  plus 0.115 of spell power, Draconic Infused Emblem's spell damage included. Frozen targets take 4 times the damage,
  and Ice Barrier (+5%), Frigid Power (+15%) and Ice Shards' crit damage (×2) apply to it as their rows state.
- Sources: combat logs of one level 60 Frost mage with StatLedger, the player's local time; geared casts from three log
  files, 2026-11-20 01:54 to 2026-11-21 07:36, casts at 0 spell power, 2026-11-21 08:40 to 08:45, and geared casts from
  three more log files (`WoWCombatLog-112226_225034`, `112326_045021`, `112326_195011`), 2026-11-22 23:26 to
  2026-11-24 08:07, with the same talents as the first set
- Game build: 1.60.1.70959 (first geared set; StatLedger's logins read it up to 2026-11-21 06:18, on the latest hotfix at
  the time, relogged) and 1.60.1.71037 (0 spell power, from the 08:35 login on, on the latest hotfix at the time,
  relogged; second geared set, on a hotfix older than the latest, which one not known)
- Date: 2026-11-20 to 2026-11-24
- Pull requests: #1150, #1199

## Method

Every Ice Lance cast by the mage is paired with its damage event on the same target. Spell power is the caster's on the
cast line. The damage is `baseAmount`, the amount before the crit and before any partial resist. Fingers of Frost counts
as frozen when one of its charges is removed within 50 ms of the cast; a Frost Nova, Freeze or Frostbite on the target
counts too. Ice Barrier, Frigid Power and Draconic Infused Emblem are read from the mage's auras at the cast. Each hit
is divided by the frozen, Ice Barrier and Frigid Power multipliers; hits on targets above level 60 are left out, since
those take a level factor.

## Counts

| Set                           | Hits | Spell power | Per hit, before modifiers | Range     |
| ----------------------------- | ---: | ----------: | ------------------------- | --------- |
| 0 spell power, no buffs       |   28 |           0 | 147.1 ± 1.4               | 136 – 159 |
| first geared, no Emblem buff  |  255 |   203 – 234 | 173.3 ± 0.5 (at 228.5)    | 161 – 186 |
| second geared, no Emblem buff |  185 |   218 – 257 | 176.4 ± 0.6 (at 242.4)    | 161 – 193 |
| second geared, Emblem buff    |   70 |   253 – 292 | 179.4 ± 0.9 (at 278.9)    | 166 – 194 |

The 0 spell power set fits the game files band (148.6 mean). Against it, the first geared set reads
(173.3 − 147.1) / 228.5 = 0.115 and the second (176.4 − 147.1) / 242.4 = 0.121; against the game files mean they read
0.108 and 0.115. Each reading is good to about ±0.006, most of it from the 28 hits at 0 spell power, which all four
share. The sim takes 0.115, the middle of the four. 1/7 sits at least four standard errors above every reading.

The modifiers hold within 1.5% across the first geared set: 36 unfrozen hits with neither buff read 171.9, 696.6 frozen
(×4.05), 798.6 frozen with Frigid Power and 836.6 frozen with both. A crit deals 2.00 times `baseAmount` (134 crits,
2.00 to 2.01). Frost Nova without Fingers of Frost reads ×4 on its three hits.

## Draconic Infused Emblem

The Emblem's buff (1318930) adds 35 spell damage while it lasts. Its gain is read from casts in the same gear, buff up
against buff down, weighted by each pair's error. In the second geared set Ice Lance gains +3.0 ± 0.95 from it over five
gear levels, where 0.115 gives +4.0. Frostbolt Rank 10 in the same logs gains +29.8 ± 1.5 on 356 hits without Ice Barrier
or Frigid Power (122 with the buff), where its 0.814 gives +28.5. The first geared set reads +1.9 ± 1.1 for Ice Lance
over three gear levels, within two standard errors of both. The buff counts toward Ice Lance like any other spell damage.
