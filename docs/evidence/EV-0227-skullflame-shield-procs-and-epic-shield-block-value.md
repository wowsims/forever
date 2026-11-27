# EV-0227: Skullflame Shield's two procs, and the block value of epic shields

- Claim: Skullflame Shield's Drain Life (equip 18815, triggers 18817) and Flamestrike (equip 18816, triggers 18818) roll
  only on landed melee hits taken, white and special, partial blocks included. Drain Life fires only on an attacker in
  front of the wearer, at the row's 3%. Flamestrike ignores facing, rolls near the row's 2%, deals the row's 90 to 140
  Fire with no spell power, crits at the wearer's spell crit for ×1.5 and reaches every enemy around the wearer, each
  hit scaled by the 20-target cap. A shield's base block value is Classic Era's ShieldBlockRegular for its quality and
  item level; Skullflame Shield's is 40.
- Sources: combat log, one Protection paladin at level 60 wearing the shield in all 85 COMBATANT_INFO snapshots, group
  dungeons (Stratholme, Sunken Temple, Blackmaw Hold, The Shaper's Terrace, Dire Maul, Scholomance, Blackrock Depths,
  Lower and Upper Blackrock Spire, Alcaz Prison). The Forever item tooltips of the 165 shields in `db.json`.
- Game build: 1.60.1.71091 until the session's 20:11 break, 1.60.1.71139 from the 21:16 restart; after hotfix push
  114399 (2026-11-25 01:15 UTC), relogged
- Date: 2026-11-25 15:33 to 2026-11-26 04:51 CET; tooltips read 2026-11-27
- Pull request: #1280

One wearer, so the rates are observed, not measured. Hits taken are the events of creatures on the wearer: 17,658 in
all, 9,968 of them landed melee (hit, partial block, crushing blow, crit). Drain Life is credited to the attacker it
hits; Flamestrike's target events group into one proc per 50 ms. Attacker angle comes from the advanced log positions
and the wearer's facing. Both procs land 0 to 54 ms after their hit, in the same spell batch.

## What rolls

Drain Life on solo attacker events (no other event of that attacker within 100 ms):

| Hit taken                          | Drain Life        |
| ---------------------------------- | ----------------- |
| White hit                          | 113 / 5,673       |
| White partial block                | 100 / 3,277       |
| Special hit, special partial block | 12 / 352, 6 / 183 |
| Miss, dodge, parry                 | 0 / 3,474         |
| Full block                         | 0 / 693           |
| Fully absorbed                     | 1 / 925           |
| Immune                             | 0 / 424           |
| Magic spell taken (any outcome)    | 0 / 1,484         |
| Ranged spell taken                 | 0 / 159           |

Flamestrike on lone hits (nothing taken within 150 ms): 76 on 4,969 landed melee hits, and none on any avoided, fully
blocked, absorbed, immune, spell or ranged event.

## Drain Life and facing

| Attacker angle from the wearer's facing | Drain Life           |
| --------------------------------------- | -------------------- |
| 0 to 30 degrees                         | 144 / 4,910 (2.93 %) |
| 30 to 60                                | 70 / 2,207 (3.17 %)  |
| 60 to 90                                | 22 / 1,091 (2.02 %)  |
| 90 to 120                               | 3 / 748 (0.40 %)     |
| 120 to 150                              | 0 / 464              |
| 150 to 180                              | 0 / 545              |

In front: 236 / 8,208 = 2.88 % [2.53, 3.26]. Behind: 3 / 1,757, at the edge of the arc where blocks and parries still
show. The failures log nothing. Flamestrike on lone hits reads 1.44 % from the front and 2.12 % from behind. The sim has
no attacker facing, so Drain Life follows the wearer's In Front of Target setting (owner ruling, 2026-11-27).

## Flamestrike rate and payload

- All landed melee hits: 170 / 9,968 = 1.71 % [1.47, 1.98]; lone landed hits 76 / 4,969 = 1.53 % [1.22, 1.91]. About
  two standard errors under the row's 2 % on one wearer; the sim keeps the row's 2 %.
- No proc came within 1.0 s of the one before: 0 on 261 landed hits there, about 4.5 expected. Two procs in different
  batches were never closer than 1.49 s, and one batch gave two at once. This is one reading after scanning several
  time bins, not a proven lockout, so the sim adds no internal cooldown. 18818's row states a 1,000 ms cast time,
  though its damage lands within 54 ms of the hit.
- Damage: non-crit unmitigated 90 to 143, mean 112.6, the row's 90 to 140, with the wearer's spell power from 165 to 348. Crits 169 / 957 (17.7 %) at ×1.5, beside the wearer's other spells at 15 to 16 %. 5 misses and 4 immunes of 966,
  89 partial resists.
- Targets per proc: mean 5.7. One proc hit 39 enemies within 5 ms for 46 to 71 each, the row's roll times 20 / 39; the
  procs up to 20 targets dealt the full roll (mean 114.6 at 20).
- Drain Life: 270 per hit, 405 on a crit (37 / 236, 15.7 %), 2 misses and 1 absorb of 239. The heal follows the damage
  dealt, lowered by the wearer's healing taken.

## Block value

The wearer's partial blocks read 40 × 1.3 (Shield Specialization) plus the Strength part across 76 snapshot windows:
2,969 of 3,718 partial blocks sit within 1 of that, and 53 × 1.3 reads 14 to 20 above every window. A second tank's
rare shields match `db.json` (Intricately Runed Shield 62, Draconian Deflector 40).

Forever's game files ship no `gametables/shieldblockregular.txt` (FDID 2200979 is not in the 1.60.1.71139 root). The
table the generator embedded was the TBC one, which the 2.5.6 installs still ship. Classic Era 1.15.9.70576's table
(CDN) has the same values in every column up to Rare and lower Epic, Legendary and Artifact columns. The base block of
all 165 shields' Forever tooltips equals the Classic Era table at the shield's quality and item level; the TBC table
was 5 to 18 higher on all 16 epic shields and matched the rest.
