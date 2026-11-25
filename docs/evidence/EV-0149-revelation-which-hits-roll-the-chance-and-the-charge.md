# EV-0149: Revelation on every class: which hits roll it, the chance, and what spends the charge

- Claim: Revelation (enchant 8217) rolls on landed non-crit direct hits of a spell with the magic defense type, the
  wearer's own or an item's, at a chance scaled by a level ramp, clamp((SpellLevel + 32 − level) / 20, 0, 1); periodic
  ticks, damage shields, leech heals and melee or ranged attacks (the paladin's judgements and Hammer of Wrath, the
  hunter's Arcane Shot and Volley) roll nothing. A hit event rolls once however many targets it strikes, and its first
  result is that one roll: when the first result crits, the event rolls nothing. Until 2026-11-24 20:54 UTC each
  target rolled. At level 60 on spells at the full ramp the chance reads 8.94% for mages, 8.84% for warlocks, 8.65% for
  priests, 7.90% for druids, 7.30% for shamans and 7.02% for paladins, and it falls as the wearer's crit from gear and
  intellect rises. The charge goes as a spell the buff affects is cast, on every class, and the cast's hit crits even
  when it lands after the buff is gone. A channel whose ticks cast a spell of their own (Blizzard, Hellfire, Rain of
  Fire, Hurricane) loses it to its first tick, which crits on every target it strikes, and its later ticks crit at
  their own chance. Arcane Missiles rolls each missile's crit at its launch: the first missile launched after the gain
  crits and spends the charge as it lands, and a missile in flight at the gain keeps its own roll and spends the charge
  only when that roll crits. Penance's heals crit with the buff up and leave the charge. Since 2026-11-25 01:13 UTC a
  channel whose ticks are its own periodic damage (Mind Flay, Wrack, Drain Soul, Drain Life) loses it to its first
  tick too, which crits, and its later ticks crit at their own chance; from 2026-11-24 20:54 UTC to that switch such a
  channel crit on every tick with the buff up and left the charge to the next spell the buff affects. Until 20:54 a
  channel kept the buff to its end and every hit of it crit. Another spell's periodic ticks never spend it. Lightning
  Overload's bolt rolls apart from the bolt it copies. The paladin's judgements and Hammer of Wrath crit and leave the
  charge. The buff reaches spells its row does not name: Fireball, Pyroblast, Frostfire Bolt, Immolate, Flame Shock,
  Moonfire and Regrowth spend it and crit on their direct hit, and their ticks do not crit.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every character with enchant 8217 on the main hand in
  the 1,356 reports the Reports section lists; all fights.
- Game build: 1.60.1.70862, 1.60.1.70888 and 1.60.1.70938 by event time against the time the CDN first listed each
  build, and 1.60.1.70938 or 1.60.1.70959 for the events after 2026-11-17 15:19 UTC; the Game builds section has the
  procs on each. The hotfix each player ran is not recorded.
- Date: 2026-11-07 to 2026-11-18
- Pull request: #1100
- Later count: what spends the charge in a channel up to 2026-11-24 20:54 UTC, the switch at that time, the direct
  spells after it, §Count of the 2026-11-24 reports
- Later count: the roll on a hit of several targets and which result decides it, a channel's first tick, Arcane
  Missiles' charge, the channels of periodic ticks and Penance, Lightning Overload's roll, periodic ticks, §Count of the
  2026-11-24 to 2026-11-25 reports

## Method

| Class   | Characters | Worn stretches | Reports |  Procs |
| ------- | ---------: | -------------: | ------: | -----: |
| Mage    |        685 |            885 |     771 | 94,176 |
| Warlock |        224 |            336 |     324 | 20,733 |
| Shaman  |        114 |            174 |     166 |  7,240 |
| Priest  |         87 |            109 |     107 |  2,118 |
| Druid   |         63 |             90 |      86 |  6,188 |
| Paladin |         71 |             97 |      92 |  2,261 |
| Hunter  |         18 |             30 |      29 |     10 |
| Warrior |          3 |              3 |       3 |      0 |
| Rogue   |          5 |              8 |       8 |      0 |

A character is a name within a region. A stretch is the wearer's events between two of its combatant info snapshots
whose main hand carries the enchant; a fight logged in two reports counts once. A proc is an application or a refresh
of the class's Revelation buff on the wearer: mage 1248808, priest 1323377, warlock 1323392, druid 1323400, hunter
1323410, shaman 1323418, paladin 1323419. A roll is a landed hit of the wearer that does not crit, a block or a
partial resist included. The buff logs 5 to 50 ms after its hit for the casters.

Three counts are used, and each number says which:

- Counted alone: clusters of the wearer's hits, chained while 60 ms apart or less, that hold one ability id only, with
  the procs from the first hit to 50 ms after the last.
- Counted apart: a hit with no other rolling spell of the wearer from 400 ms before it to 100 ms after, and a proc
  within 100 ms after it.
- Fitted share: one chance per ability id fitted over every wearer of the class at once, each proc shared among the
  wearer's rolls in the 400 ms before it in proportion to chance times lag share. Its proc counts are fractions.

The wearer's level is not in a report. It is estimated as the highest base level among the ranked class spells the
wearer casts, a lower bound; an estimate of 60 is exact, and the level and crit sections use those stretches. A
snapshot carries the wearer's intellect and a crit value in steps of 7 that is the same for melee and spell. The crit
value is read as crit rating at 14 a percent and intellect as 0.0168% crit a point, the sim's figures at 60; neither
was checked against the wearer's items, and talents, base crit and buffs are not in a snapshot. Buff state is capped at
15 s after the last proc, since some removals are missing from the logs. A cast is the log's cast event, which a spell
with a cast time logs as the cast completes. The game files rows are read on 1.60.1.70938 at hotfix push 114102; the
rows of the proc aura 1248806 and of the seven buffs read the same on 1.60.1.70959.

## Game builds

The CDN first listed 1.60.1.70862 for product wowf on 2026-11-07 04:43 UTC, 70888 on 2026-11-10 15:10, 70938 on
2026-11-13 01:36 and 70959 on 2026-11-17 15:19 (wago.tools' build history). On 2026-11-18 the US and EU regions list
70959 and the CN, KR and TW regions 70938. When each region's realms took a build is not recorded, so the events after
70959 was listed ran on either build. The stretches come from the US, EU, CN, KR and TW regions. Procs by the build
window of their timestamp:

| Class   | 70862 |  70888 |  70938 | 70938 or 70959 |
| ------- | ----: | -----: | -----: | -------------: |
| Mage    | 3,461 | 11,505 | 70,365 |          8,845 |
| Warlock |   651 |  1,739 | 14,847 |          3,496 |
| Shaman  |   292 |    502 |  5,456 |            990 |
| Priest  |   308 |    173 |  1,357 |            280 |
| Druid   |     0 |      4 |  4,588 |          1,596 |
| Paladin |    12 |    101 |  1,853 |            295 |
| Hunter  |     0 |      0 |     10 |              0 |

## Which hits roll it

The spells with the most fitted procs in each class, ticks left out, every rank pooled:

| Class   | Spell                        |   Rolls | Fitted procs | Chance | Named by the buff |
| ------- | ---------------------------- | ------: | -----------: | -----: | ----------------- |
| Mage    | Blizzard, damage             | 378,273 |     34,940.6 |  9.24% | yes               |
| Mage    | Arcane Explosion, damage     | 208,604 |     18,204.8 |  8.73% | yes               |
| Mage    | Flamestrike, damage          | 100,316 |      9,055.3 |  9.03% | some ids          |
| Mage    | Arcane Missiles, damage      |  88,980 |      7,977.7 |  8.97% | yes               |
| Mage    | Frostbolt, damage            |  72,083 |      6,223.3 |  8.63% | yes               |
| Mage    | Cone of Cold, damage         |  49,260 |      4,673.2 |  9.49% | yes               |
| Mage    | Arcane Blast, damage         |  34,277 |      2,978.2 |  8.69% | yes               |
| Mage    | Frost Nova, damage           |  40,180 |      2,612.1 |  6.50% | yes               |
| Mage    | Fire Blast, damage           |  24,002 |      2,133.9 |  8.89% | yes               |
| Warlock | Hellfire Effect, damage      |  85,238 |      7,959.3 |  9.34% | yes               |
| Warlock | Rain of Fire, damage         |  67,222 |      6,450.9 |  9.60% | yes               |
| Warlock | Searing Pain, damage         |  18,343 |      1,721.2 |  9.38% | yes               |
| Warlock | Shadow Bolt, damage          |  17,426 |      1,582.5 |  9.08% | yes               |
| Warlock | Hellfire, damage             |  16,118 |      1,298.9 |  8.06% | yes               |
| Warlock | Immolate, damage             |   7,618 |        710.0 |  9.32% | no                |
| Warlock | Incinerate, damage           |   3,586 |        326.8 |  9.11% | yes               |
| Warlock | Shadowburn, damage           |   2,280 |        198.9 |  8.73% | yes               |
| Warlock | Conflagrate, damage          |   1,912 |        155.4 |  8.13% | yes               |
| Shaman  | Fire Nova, damage            |  32,900 |      2,510.1 |  7.63% | yes               |
| Shaman  | Lightning Bolt, damage       |  14,721 |      1,209.3 |  8.21% | yes               |
| Shaman  | Earth Shock, damage          |  16,280 |      1,158.2 |  7.11% | yes               |
| Shaman  | Chain Lightning, damage      |   6,906 |        589.7 |  8.54% | yes               |
| Shaman  | Flame Shock, damage          |   6,183 |        518.3 |  8.38% | no                |
| Shaman  | Chain Heal, heal             |   6,984 |        451.0 |  6.46% | yes               |
| Shaman  | Lightning Shield, damage     |   4,608 |        351.9 |  7.64% | no                |
| Shaman  | Lesser Healing Wave, heal    |   2,338 |        146.9 |  6.28% | yes               |
| Shaman  | Healing Wave, heal           |   2,567 |        137.8 |  5.37% | yes               |
| Priest  | Mind Blast, damage           |   7,075 |        607.8 |  8.59% | yes               |
| Priest  | Shadow Word: Death, damage   |   3,619 |        325.3 |  8.99% | yes               |
| Priest  | Penance, heal                |   3,653 |        315.0 |  8.62% | yes               |
| Priest  | Prayer of Healing, heal      |   1,319 |         86.2 |  6.54% | yes               |
| Priest  | Heal, heal                   |   1,288 |         84.0 |  6.52% | yes               |
| Priest  | Flash Heal, heal             |     935 |         59.5 |  6.37% | yes               |
| Priest  | Penance, damage              |     575 |         57.4 |  9.99% | yes               |
| Priest  | Holy Nova, damage            |     451 |         32.0 |  7.10% | no                |
| Druid   | Hurricane, damage            |  51,204 |      4,209.0 |  8.22% | yes               |
| Druid   | Wrath, damage                |  11,739 |        915.7 |  7.80% | yes               |
| Druid   | Starfire, damage             |   6,194 |        477.9 |  7.71% | yes               |
| Druid   | Moonfire, damage             |   5,215 |        414.9 |  7.96% | no                |
| Druid   | Regrowth, heal               |   1,672 |        112.0 |  6.70% | no                |
| Druid   | Healing Touch, heal          |     571 |         32.5 |  5.69% | yes               |
| Paladin | Crest of Retribution, damage |   7,751 |        619.7 |  8.00% | item or racial    |
| Paladin | Flash of Light, heal         |   8,267 |        566.0 |  6.85% | yes               |
| Paladin | Exorcism, damage             |   1,737 |        132.2 |  7.61% | no                |
| Paladin | Light's Vigil, heal          |   1,588 |        109.5 |  6.90% | yes               |
| Paladin | Holy Wrath, damage           |   1,025 |        105.5 | 10.30% | yes               |
| Paladin | Krol'dok Resolve, heal       |   2,247 |         96.8 |  4.31% | item or racial    |
| Paladin | Holy Shock, heal             |   1,307 |         91.5 |  7.00% | yes               |
| Paladin | Holy Light, heal             |   1,237 |         70.5 |  5.70% | yes               |
| Paladin | Seal of Light, heal          |     494 |         42.7 |  8.65% | no                |

Frost Nova and the heals read low because their lower ranks are in the pool; the level ramp below accounts for them.
Krol'dok Resolve heals at a block, beside the wearer's damage shields, and the fit hands part of its procs to those;
counted apart it reads 155 of 1,557, 10.0%.

Every spell that rolls has `SpellCategories.DefenseType` 1, the magic defense type, and a direct damage or heal effect,
with one exception below. What rolls nothing, with rolls and fitted procs:

| Kind                                         | Examples                                                                                                                                                          |
| -------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Periodic ticks                               | Curse of Agony 8.1 of 297,099; Corruption 1.2 of 172,664; Shadow Word: Pain 19.3 of 62,290                                                                        |
| Attributes[3] 0x2000000, treat as periodic   | Consecration's ticks 0 of 252,630; Ignite 11.1 of 20,615                                                                                                          |
| Defense type 2 or 3, melee or ranged         | Stormstrike 0 of 22,378; Seal of Fury's hit 10.7 of 39,824; Volley 0 of 28,587; Auto Shot 0 of 12,322                                                             |
| Damage shields                               | Retribution Aura 137.5 of 127,751; Thorns 29.3 of 44,012; Fire Shield 0 of 11,254                                                                                 |
| Heals of a periodic leech                    | Siphon Life 0 of 90,118; Devouring Plague 0 of 27,990; Drain Life 3.0 of 15,441                                                                                   |
| Spell level 1 or 0, by the level ramp        | Touch of the Grave 40.4 of 83,598; Demonic Energies 12.0 of 149,523; Flametongue Attack 22.2 of 16,676; Force Imbued 0 of 9,624; Cursed Murloc Eye 89.2 of 58,753 |
| No SpellLevels row                           | Ancient Power 0 of 2,267; Grounded Electroshock 13.7 of 1,133; Moonshadow 0 of 731; Iron Grenade 0 of 260; Frenzied Regeneration 0 of 189                         |
| Wand shots                                   | Shoot 5.5 of 67,663, and 0.4 of 3,986 at level estimates of 26 to 31, where the ramp leaves a share                                                               |
| Attributes[3] 0x10000, suppress caster procs | Thorncurse 0 of 8,937                                                                                                                                             |

- The paladin's judgements and Hammer of Wrath, counted apart on each build window: Judgement of Righteousness 0 of
  1,197, Judgement of Fury 2 of 6,391, Judgement of Command 2 of 2,053 and Hammer of Wrath 0 of 753. The hits of
  Righteousness, Fury and Hammer of Wrath have the ranged defense type and Command's the melee one.
- The hunter's Arcane Shot rolls 0 of 3,189 and Volley 0 of 28,587; the buff's class flags name both and both have the
  ranged defense type. The fit places the hunters' 10 procs on Flame Lash 27655, an item's spell with the magic defense
  type. So nothing a hunter casts rolls it.
- Seal of Light's heal has no SpellCategories row and rolls: 33 of 384 counted apart. Judgement of Light's heal has
  none either and rolls nothing: 1 of 1,404 counted apart on the paladins. On every class, with no hit or heal that
  could roll in the same window, it reads 0 of 896 at rank 4 (15 wearers of five classes) and 0 of 243 at rank 1 (9
  wearers of three). Its row differs in `SpellMisc.Attributes[0]`, which carries 0x10 and 0x20000000 on top of the
  0x40000 both have. The other spells with 0x10 that can be counted apart roll nothing (Shoot 1 of 57,176, Moonshadow
  0 of 585, Frenzied Regeneration 0 of 147), and each has another reason to: a wand shot, or no SpellLevels row. Holy
  Binding carries it too, and none of its 217 heals counts apart: each lands beside a Flash Heal, and the fit hands it
  11.7 of them. What makes one heal roll and the other not is not derived.
- Lightning Shield's hits roll (a fitted 351.9 of 4,608, 7.64%) though they answer a strike, as the damage shields do.
  The Holy Shield hit a paladin's block sets off rolls nothing (0 of 9,708 and 1 of 1,725 counted apart on the first
  82 paladin reports).
- Explosive Trap's direct hit rolls 0 of 3,376. The trap casts it, and the hunter does not.
- 1,474 of the mages' 94,176 procs log with no event of the wearer in the 400 ms before them and are not placed.
- No warrior (3 on the main hand, 2 on the off hand) and no rogue (5 and 1) gains a Revelation buff of any class.
  Those are too few wearers to settle it. One hunter wears the enchant on the off hand and logs no proc.

## The level ramp

A spell far under its caster's level rolls at a reduced chance, down to nothing. At a level estimate of 60, counted
alone, against the class's own chance on spells of spell level 52 and over:

| SpellLevel | Procs of rolls | Share of the class's chance | Ramp (SpellLevel − 28) / 20 |
| ---------: | -------------- | --------------------------: | --------------------------: |
|    1 to 26 | 0 of 13,123    |                        0.00 |                        0.00 |
|         32 | 3 of 186       |                        0.20 |                        0.20 |
|         34 | 6 of 262       |                        0.26 |                        0.30 |
|         38 | 4 of 245       |                        0.19 |                        0.50 |
|         40 | 81 of 1,589    |                        0.65 |                        0.60 |
|         42 | 24 of 436      |                        0.77 |                        0.70 |
|         46 | 158 of 2,071   |                        0.86 |                        0.90 |
|         48 | 103 of 1,051   |                        1.10 |                        1.00 |
|         50 | 74 of 763      |                        1.12 |                        1.00 |

A fit of the ramp's two ends over the fitted shares at level estimate 60 puts the chance at nothing from 31.5 levels
under the caster (31.0 to 32.0) and at full from 12.0 levels under (10.5 to 13.5): clamp((SpellLevel + 32 − level) / 20,
0, 1). No cut at all fits 1,340 log likelihood worse. Under 60 the same ramp shows against the level estimate: the
fitted chance reads 9.1% at a gap under 8 levels, 6.2% at 16 to 19, 4.5% at 20 to 23, 2.9% at 24 to 27, 0.8% at 28 to
31 and nothing from 32.

The ramp reads `SpellLevels.SpellLevel`, as the row's MaxLevel fits worse: Frost Nova and Chain Heal at spell level 40
and MaxLevel 45 keep 0.59 and 0.53 of the chance, where the ramp (22 + MaxLevel − level) / 20 gives 0.35. It
covers item and racial spells too. Waveslicer (spell level 45) reads a fitted 6.4% at 60. Spells of spell level 1 roll
nothing from a level estimate of 32 and a share under it: Flametongue Attack reads a fitted 5.0% at estimates of 20 to
25 (8.9 of 176) and 1.1% at 26 to 31 (13.0 of 1,202). The three spells of spell level 0 in the logs (Cursed Murloc
Eye, Soul Link, Smokey's Lighter) roll nothing, and none of them is a class spell. A spell with no SpellLevels row
rolls nothing either, which is what a spell level of 0 gives from level 32. The direct spells without one, damage
shields left out, read a fitted 13.7 of 18,038 at level estimates of 32 and over, all 13.7 on Grounded Electroshock,
and 0 of 3,239 counted apart: Ancient Power 0 of 1,214, Blood Siphon 0 of 595, Moonshadow 0 of 585, Grounded
Electroshock 0 of 472, Frenzied Regeneration 0 of 147, Satchel of Potions 0 of 104 and Libram of Divinity 0 of 104.
Under an estimate of 32 they log 8 rolls, too few to show a share.

## The chance and crit

At a level estimate of 60, counted alone on ranked class spells of spell level 52 and over:

| Class   | Procs of rolls    | Chance |
| ------- | ----------------- | -----: |
| Mage    | 27,896 of 312,128 |  8.94% |
| Warlock | 3,385 of 38,274   |  8.84% |
| Priest  | 441 of 5,101      |  8.65% |
| Druid   | 3,328 of 42,144   |  7.90% |
| Shaman  | 2,972 of 40,695   |  7.30% |
| Paladin | 426 of 6,066      |  7.02% |

- Count: §Count of the 2026-11-24 reports

On these reports each target rolled, and the number of targets a cast hits does not move the chance a roll: 8.42% on
clusters of one roll, 8.79% on 2 to 3, 8.78% on 4 to 6, 8.83% on 7 to 10 and 8.39% on 11 and over. Nor does the build
window: the mages at 60 read 8.59% on 70888 (6 characters), 8.99% on 70938 and 8.98% after 70959 was listed. Since
2026-11-24 20:54 UTC a hit event rolls once.

The tooltip says the chance falls as the wearer's crit chance rises. Against crit from gear and intellect, on the
stretches at a level estimate of 60 and the fitted shares of ranked class spells within 8 levels of the wearer:

| Class   | Characters |   Rolls | Chance at the class's mean crit | Slope, points per crit percent (95% interval) |
| ------- | ---------: | ------: | ------------------------------- | --------------------------------------------- |
| Mage    |        211 | 343,787 | 8.97% at 6.9%                   | −0.347 (−0.431 to −0.265)                     |
| Warlock |         94 |  96,256 | 9.18% at 5.3%                   | −0.096 (−0.305 to +0.083)                     |
| Shaman  |         37 |  46,323 | 7.32% at 8.0%                   | −0.217 (−0.350 to −0.058)                     |
| Priest  |         37 |   8,617 | 8.40% at 5.4%                   | −0.289 (−0.818 to +0.445)                     |
| Druid   |         29 |  45,152 | 7.89% at 5.9%                   | +0.031 (−0.204 to +0.301)                     |
| Paladin |         19 |   8,086 | 7.33% at 4.6%                   | −0.611 (−0.955 to −0.125)                     |

The slope is a weighted least squares fit over stretches, with an interval from a bootstrap over characters. One slope
for every class with an intercept each reads −0.265 (−0.334 to −0.198), and −0.185 (−0.282 to −0.083) without the
mages. For the mages:

- By crit rating on gear: 9.48% at 0, 9.32% at 7, 9.16% at 14, 8.99% at 21 and 8.35% at 28. By intellect: 9.54% at 250
  to 299, 8.99% at 300 to 349 and 8.58% at 350 to 399.
- Rating and intellect fitted together read −0.310 (−0.438 to −0.188) and −0.397 (−0.559 to −0.251) points per crit
  percent, so each counts on its own and by about the crit it gives.
- The slope holds inside each of the mage's larger spells (Blizzard −0.28, Arcane Explosion −0.22, Flamestrike −0.24),
  inside each dungeon (−0.344, −0.422 to −0.268) and inside the 70938 window alone (−0.373, −0.474 to −0.279).

Moonkin Aura 24907 gives its party 3% crit. On wearers that are not druids, at equal crit from gear and intellect, the
chance with the aura up reads 0.22 points under the chance with it down for mages (−0.49 to +0.06), where three times
the gear slope is −1.02. The two measurements disagree, and the logs do not say which crit the game counts. The
spell's own crit against its target does not lower it: Blizzard crits 65% of its hits and Hellfire 60%, with the buff
keeping them critting, and they read 9.24% and 9.34%, as high as their class's other spells.

The classes' chances differ at the same crit from gear and intellect, in the order their base crit and crit talents
would give, but those are not in the logs. No formula for the fall is derived here.

## The charge

The buff lasts 15 s: 5,363 buffs end 14.9 to 15.2 s after their last proc, 109,153 earlier and 23 later. Each of the
seven buffs adds 100% crit chance to the spells its class flags name (`SpellEffect.EffectSpellClassMask`) and has one
charge.

### It goes at the cast

Casts with the buff up just before them, the buff removed within 2 ms of the cast, and the crit of the cast's first
landed hit:

| Class   | Spell              |  Casts | Removed at the cast | First hit crits  | Hit after the cast, median |
| ------- | ------------------ | -----: | ------------------: | ---------------- | -------------------------: |
| Mage    | Arcane Explosion   | 17,682 |              17,659 | 17,350 of 17,432 |                       0 ms |
| Mage    | Frostbolt          |  6,263 |               6,256 | 5,820 of 5,820   |                     517 ms |
| Mage    | Arcane Blast       |  3,082 |               3,078 | 3,026 of 3,041   |                       0 ms |
| Warlock | Shadow Bolt        |  1,462 |               1,460 | 1,336 of 1,401   |                     674 ms |
| Warlock | Searing Pain       |  1,118 |               1,113 | 1,076 of 1,097   |                       1 ms |
| Warlock | Incinerate         |    325 |                 324 | 313 of 315       |                     741 ms |
| Shaman  | Earth Shock        |  1,935 |               1,933 | 1,865 of 1,872   |                       0 ms |
| Shaman  | Lightning Bolt     |  1,401 |               1,400 | 1,347 of 1,355   |                     309 ms |
| Shaman  | Chain Heal         |    316 |                 316 | 316 of 316       |                       0 ms |
| Priest  | Shadow Word: Death |    144 |                 144 | 137 of 139       |                       1 ms |
| Priest  | Mind Blast         |    129 |                 128 | 124 of 124       |                       0 ms |
| Priest  | Heal               |     78 |                  78 | 78 of 78         |                       0 ms |
| Druid   | Wrath              |    777 |                 777 | 704 of 720       |                     629 ms |
| Druid   | Starfire           |    583 |                 583 | 570 of 573       |                       0 ms |
| Paladin | Flash of Light     |    575 |                 574 | 575 of 575       |                       0 ms |
| Paladin | Holy Light         |    110 |                 109 | 110 of 110       |                       0 ms |

The first hit is the spell's next landed hit within 4 s of the cast, so a cast that misses is paired with a later
cast's hit and reads as no crit; Frostbolt's row pairs within 1.2 s and has none. A spell with travel time loses the
buff at the cast and still crits when it lands: Frostbolt, Shadow Bolt, Incinerate, Lightning Bolt and Wrath.

Some instant casts lose the buff at the cast and do not crit: Fire Blast 391 of 2,373 landed hits, Ice Lance 495 of
1,613 and Cone of Cold 226 of 2,533. Another cast of the wearer sits within 20 ms of 249 of those Fire Blasts, 234 of
the Ice Lances and 53 of the Cones of Cold, against 150 of the 1,982 Fire Blasts that crit, so the one charge went to
the other cast. The rest are not explained.

### A channel kept it to its end until 2026-11-24

- Count: §Count of the 2026-11-24 reports

On these reports a channelled spell the buff names does not spend the charge at its cast. Every hit of the channel
crits while the buff is up, and the buff goes when the channel ends. Since 2026-11-24 20:54 UTC a channel whose ticks
cast a spell of their own loses it to its first tick. A channel whose ticks are its own periodic damage kept it past
its end from that time and loses it to its first tick since 2026-11-25 01:13 UTC. Casts with the buff up and gained
from another spell:

| Spell            | Casts | Hits with the buff up | Hits after it is gone |
| ---------------- | ----: | --------------------- | --------------------- |
| Blizzard, mage   | 2,902 | 77,011 of 77,012 crit | 128 of 1,384 crit     |
| Hurricane, druid |   117 | 3,271 of 3,271 crit   | 11 of 98 crit         |

For Blizzard the removal comes under 1 s after the channel's last hit with no cast of the wearer at that instant in
2,037 of the 2,902 casts, and within 50 ms of the last hit in 156 more; it sits at the wearer's next cast, away from
the last hit, in 269. A buff gained from a tick of the channel behaves the same: it is removed as the next cast of the
channel begins, which is when the running one ends (1,946 Blizzard casts, 99% of them straight after a Blizzard cast).
Rain of Fire, Hellfire and Penance show the same split between casts that keep the buff and casts that begin as it is
removed. Arcane Missiles' first hit crits on 20,098 of 20,172 casts; its later hits and its removal are not measured.

The periodic ticks of a spell the buff names crit while the buff is up, as a channel's hits do:

| Spell               | Ticks with the buff up |  Crits | Crit with the buff down |
| ------------------- | ---------------------: | -----: | ----------------------: |
| Hellfire, warlock   |                 38,830 | 38,830 |                    8.6% |
| Mind Flay, priest   |                  1,803 |  1,792 |                    7.4% |
| Wrack, warlock      |                  1,104 |  1,104 |                   12.2% |
| Drain Life, warlock |                    348 |    348 |                    9.9% |

### Spells the row does not name

These spend the charge at the cast and crit, on every build window, though the buff's class flags do not name them on
1.60.1.70938 at push 114102 or on 1.60.1.70959. Each has a direct hit and a periodic part:

| Class   | Spell          | Casts with the buff up | Removed at the cast | Crits of landed hits |
| ------- | -------------- | ---------------------: | ------------------: | -------------------- |
| Mage    | Fireball       |                    373 |                 358 | 338 of 340           |
| Mage    | Pyroblast      |                    356 |                 345 | 327 of 327           |
| Mage    | Frostfire Bolt |                    225 |                 190 | 191 of 203           |
| Warlock | Immolate       |                    314 |                 313 | 297 of 297           |
| Shaman  | Flame Shock    |                    411 |                 393 | 394 of 404           |
| Druid   | Moonfire       |                    280 |                 268 | 217 of 275           |
| Druid   | Regrowth       |                     43 |                  40 | 42 of 42             |

Their ticks do not crit, neither while the buff is up nor after a cast that spent it:

| Spell          | Ticks with the buff up, crits | Ticks with it down, crits | Ticks after a cast with the buff up, crits |
| -------------- | ----------------------------- | ------------------------- | ------------------------------------------ |
| Fireball       | 72 of 583, 12.3%              | 677 of 5,755, 11.8%       | 75 of 669, 11.2%                           |
| Pyroblast      | 78 of 655, 11.9%              | 760 of 6,213, 12.2%       | 56 of 519, 10.8%                           |
| Frostfire Bolt | 23 of 268, 8.6%               | 224 of 2,320, 9.7%        | not counted                                |
| Immolate       | 246 of 2,885, 8.5%            | 2,217 of 24,803, 8.9%     | 68 of 745, 9.1%                            |
| Flame Shock    | 356 of 2,285, 15.6%           | 3,335 of 23,329, 14.3%    | 132 of 914, 14.4%                          |
| Moonfire       | 669 of 2,825, 23.7%           | 4,413 of 18,527, 23.8%    | 160 of 630, 25.4%                          |
| Regrowth       | 531 of 1,415, 37.5%           | 8,677 of 20,942, 41.4%    | 21 of 120, 17.5%                           |

Flamestrike's cast ids are named by the buff's fourth mask word and spend the charge at 755 of 756 casts; its ground
tick ids are not named, and EV-0028 has the ticks. A spell with only a periodic effect that the buff does not name does
not spend it: Insect Swarm loses the buff at 36 of 203 casts, Shadow Word: Pain at 12 of 269 and Devouring Plague at 7
of 90, which is what other casts nearby take. Their ticks crit as often with the buff up as down (Curse of Agony 10.4%
against 11.3%, Corruption 10.9% against 11.4%).

### Named spells that crit and leave it

The paladin's judgements and Hammer of Wrath are named by the buff, have the ranged or melee defense type, and crit on
every hit while the buff is up without spending it:

| Spell                      | Hits with the buff up | Crits | Buff removed within 60 ms |
| -------------------------- | --------------------: | ----: | ------------------------: |
| Judgement of Fury          |                   805 |   805 |                         2 |
| Judgement of Command       |                    45 |    45 |                         0 |
| Judgement of Righteousness |                    40 |    40 |                         0 |
| Hammer of Wrath            |                    28 |    28 |                         0 |

So one proc crits each of them for up to 15 s, until a spending cast or the 15 s end it. The priest's Holy Nova does
the same: 66 casts with the buff up, 66 crit and 2 lose the buff at the cast.

### Spells the buff does not affect

The paladin's Exorcism (171 hits with the buff up), Lay on Hands (7), Seal of Light's heal (117) and Judgement of
Light's heal (435) are not named, do not crit and do not spend the charge. Exorcism never crits at all (EV-0188). The
rows of the two seal heals carry `SpellMisc.Attributes[2]` 0x20000000, cannot crit.

### The paladin's Holy Shock and Light's Vigil

On the first 82 paladin reports, by cast with the buff up: Holy Shock's cast ids have defense type 0 and lose the buff
0 or 1 ms after the cast in all 85 casts. Its damage crits 24 of 24. Its heal crits 14 of 61, under its 33.0% with the
buff down, and logs after the removal in 57 of them. Light's Vigil loses the buff at the cast in all 21 casts, 15 of
them with no heal within 300 ms, and its heals, which come with later Holy Shock casts, crit 4 of 18 with the buff up.
A Flash of Light or Holy Light heal that logs after the removal still crits (121 of 121).

## Other readings

- A crit never rolls it: 33 procs sit in the 344,651 clusters that hold crits only, with 1,087,639 crit hits.
- There is no cooldown: 12,838 procs log in the same millisecond as the one before and 1,210 more within 100 ms.
- A proc can come with the buff up, as a refresh: 361 of the paladins' 1,437 procs at level estimate 60.
- A heal that only overheals rolls as any other: counted alone, 25 of 434 (5.8%) against 1,412 of 21,387 (6.6%) over
  the paladin's, shaman's and priest's main heals.

## Reports

BjqYNRDzxyFpnhk9, M7yHQXPwfDgatmhc, mdXbpRCtVFBZza4f, 6dDvV4KTFRJgZkfx, aQvzkM1yV9fm6LT7, WAv29rjahg4Nb7QK,
FPyXRmZ9ATB8YnGd, P2f34YNVadyXrgCF, 8f1Z4b7AztyxRFH9, QtC9dpxw2NfHPjvK, tDH2NhVbMLYcdxnQ, NMpCXH7L6Vy8cftF,
py6rTJYqjQXF1tNL, qWxtAfwcn9hKQXmG, dXgNAc96DtkWxYaL, N4WQxL1wPYnqCZ97, YKb8xArdm1LwcqQt, 7ZMpvfcF6gBtmVQG,
LCPFz3k2tDMvH9Zc, Q7DHfpTJLR83cKqa, Mg71QDha8cn4ATLB, HDRjxFYTXtBwfCcn, 3FBTxzLbhj27ZJXc, 9hGqnaW3v7ABH2Nm,
WzjcxdwJLPMRFgfV, N4dnAJM8jhvRwkQf, ChcKRrGJ832Wbj1Z, Wd7KTFxR1QNz2cr6, PfDk9GQLB1n6RXT3, vfcxgRqJMHkPndwV,
aMf6FV9XA4dgRBbW, j2791tbGmDdcAaLr, 7YH9phK3kavyQRNr, BCZgcRxrh7dA4ayX, j6q4H8G7tTMYKzCg, Z3PVk7LG1afKQjm8,
GNmLzbHkxAZJVFYt, yTDCtLmVK4kq7gQ8, 6WgpmHQrv9kn1ZXJ, JCAj374pdrnqmhaM, tg8HxwFGDd4qfyRn, Mv2qPwTBba1xWKRH,
nAj8K93Df1XV7aFQ, WAfqvJMK79R1kPYC, xJG3HkjWaLvKyF4C, 4ZWgkRbDAxz7c1aY, Gp4kmc6rWTMV23v7, RdKfrZn1VvTtGYwL,
gnMWYZa9ArC8jbNc, kfzAGMFxWgZB3Ld6, a3xGLZnNKTjQMFPY, PHqvNCxT1JVtpWFj, NRpydtnGrPhTjw4m, 9vGFk4MCPmhZV6bq,
cdj9D8q42JAvtxpH, TcYJdbaLwWR9m7GN, bqkXHJ6K3NV8WM1v, FdfXnMQwZ4xbCW37, 9MBNmDVh41aLzwfy, JfdzAYrMLnt6HGPm,
AY1qmxbnN27ZydDX, V7kcRxY19FAHgNyX, z3y1rQYZqpb7t8hL, WFD3bkgRhfMCz4Hx, KPANBtWRHTaLpzQM, 7TVZ84hJaAYxfzPd,
K4hMybcpkX1fCYAJ, gBDMLv8J1K7nqkpd, FHqYCNcKntvDAjxg, nmyDWqNwM9CVaT73, WjcBvb2MFRpn8wAm, P8mKLRWg734kFVqY,
t2XaBfMqwLbprCkJ, 8hTHxr37tDfC2NJb, NHZr1p7m98D4JXCv, CYy2LWpMtNJnjKgf, XMNWxtyvBFGPRmh7, 7D4gRyVcBb8n9Nfr,
9qwtJv18YgLmKGQj, kqTymHZfWPKv7Xpw, J13Z8p9crxhCVaPm, VznpBGaAxRyM9WbL, N46dwzAftLn2ZM9V, jT7NDMWFfhKk1bdz,
8qY7wp2A1jbhmTfg, XmzAPvt63xW2Y9M7, aYtRFrf3kGAwxq1c, d1tWxaPmNMkDfyJh, cvRTFb1qhGBKNkCD, gcVJYr4MGnKFdqmR,
JDPLbq7M6KZTrzCx, y7f9mtXnxRW2rcHv, DGRZxFkKHJwv1f7b, RdajQmzBx1HY4rCk, RhJCKYHvmVyBPNcn, 48K9Pz7f1k2GhLHC,
WCGvyfprwgdM4az2, RDWzgm8YB9fxCkyc, 13HVjRGCXBfyZpAc, 97Y4tQ8NMvPXGTpq, 8NbMjd4mkTz2KtnL, ZPvHADGq4xfNtTMJ,
NRK61HZcDdJkrwXG, C9mZfGPKnDrpJQXc, NcD3VATvRtbxfyPr, hARdfj2bHw1rnLcP, H4dmc8KwvkyQDRG9, bLPDtGgqT7vWCNcY,
J9DYB8rzxP2Lcq6b, dwFXgDzpPvK3tTY6, 1Kh4yrHfamXj7J6g, NmgMW7qdGQDBz4bh, mZw3WaVHvjzcdg2N, Kchp1BXjdDRQCVM8,
4DThFqHwQY1fkd79, RzbxPfGABaMCktmy, xZbcJMRwptHjq3VC, 1fXaxy3G4ACP6RcH, 3z9LMdfQApXGtj7a, PYkKRvb8qnFDX7dh,
HYPJBWmnTft3pjhK, bAfH71NwFmdx8QB9, b3AG8zvWTrBxV4HY, v6ALnJxGb2cy7fN3, 3DLckMNXhxpGATqK, PQdJmhCLGYRvT9Kk,
WgKHB2M6jdqhDwV1, n4TFAmyDPCpfJcVM, XnqKfgCYhxaz62w8, 27jZGhfHgp1LR8MB, v1AK2J4ChjHWVpRa, dgrXCT3xvtqJG824,
h6qwHXjvcAJ89dBF, d4HWPJyXzLRCgNKh, nLXBG72rh3KRWCwt, DPNzRZ9Md3CpTWnm, QM9ZCxFTYGpA7WgD, Tnxar6dY8ZVNjyFg,
WPvXcNrBRqfFYpwx, BQHh6gM9GCAZ431x, L1HhkPyDZR863w9N, RgChT9a1P6nv4Wjr, ThPAQmqyBVdZ3CXD, mXJzbvkghxGNQcR1,
pkdQ1vRh2ycfNYLG, 4JN3VyF8qPpwtL2Y, cjM9kVNAhwmn1KDf, YJPVatA7TqpZjb6B, 4cTqDZz1wd7XnaFp, wR8mhZn9kNtYH7Va,
1XJNPHnv34Gqhfpt, Pt6FynYJQGdhgjbf, aMFWQbvpryT1JBVY, 8vZWVpdfD9AQcRHP, RPaxXFdLc3zJpwT6, kH1dt4TpQjfYmaAv,
hqp7dj6fNmTDHaL4, bTAkcPhFLCmyWpYa, m17XBW2RqjPchKyT, BKGvfjpZ9bk4aP6L, Nq3C7bVGnvmyh2Pw, GwVHZMQR8j9BcJK7,
2Yh9jPZQKM7nz48D, fFMPzh2LjnRTtaAk, jWnC8Z7xDvpRcw46, G4H1Qrx8gRWtPVAJ, nCVTx9K2XtJ1GFHf, X4GmTFMZLky9YaRH,
D7FX6RWN8YkMtmgw, Kb4h3xFC7cnVZ8rf, YvBtH963jKD17fbp, Cm2fRBkD6Td3YJvt, qztgwTZbCRQGHV4F, hnvL8BP7cVJz3Y1A,
kwth2jHXJ6c9TaFN, 76K8gP2rbGfBtV3d, DQb2rzfnNyL148BR, YKvr2t6Pm4FXBMZy, v3wnr6YCzFJPdRqa, pFhPTV74mg9QxN8t,
JK1z96cCrj28tpBV, rAQLBaH1k7PF3vyJ, dpXxrVfmk628TNHC, zYxCPXRW4DcNGd6r, JvcpwWdV3XnyRzjB, 7hLzaxV4n3fKcWbJ,
nWGX6cj3kNtg4dAh, QrA9zMNJLvYKaybP, mjqgn2BXQbKLHzw6, xznTPKZhMd3Bk4pC, NGxrPCzY4BXAHyWh, CHM9rnRfVTgF8m6j,
TnrZd63NcXmFPGW2, kfPjNt147ygchmJp, 6hkxHfdyzFTvqMAn, T9bLgzrnpvAqPV8R, gPCcxZqBAby7G3r9, kJmfA12K6qBgQdzX,
QwTA46182VJtGNar, M74kAxJTdpNwavrq, TXyqWfghZFwdmA4p, 7jmgFw2CJ9K6YPRt, B24cafZAYkdqbCVP, v6yYFTnmx9ZXPjHK,
vGFaKgLT9mkn1Jdy, avZQz7NCgrP6qbXm, qx17r2MN9BzdAL4H, Arm1GFgKV2R7knMp, XDLdpNmG7BcaH1jf, bDC9cwqAYBkytfGF,
1VmBbcZvXLRYWMyx, XPJdBCjfpH6Nnryq, Qg6qwJy8hcGLY7PF, Aqf42MVwTzaC1pDH, VXB4cGh1MrHD9y2k, VwyW3ZP1H6RTpdmf,
3jMCdmAGrHDTLKwh, BagZHvPMqmzC2fTr, 1RBQVkNMf2gj9PHz, vAd3fpM7rKa46kTw, Hk9q3tMhCYcLKjAm, f8ydL6xMqHkBVTpJ,
KzHcDqw8tYnMGmF2, Z8NxMfyv6dTpqmP1, qd8y1pnFaQKM3zkr, 4y3W78DNaczXYTjF, Q2qrkZR6NnbzLmaw, pcxf4HNtgVKq3F7Z,
3aG6Pf4pzDtJL9dT, 7ABZ4Nm8xyVQT3Y9, 8rZAgdVHGChc7mWF, f36RaB12wyDcTCGY, FGmjgRrQ9vthAWa4, 1A3gZQMq9LVJ4dtN,
1hRZ2v8PwKWt7DkX, xwPHq9NVcMKQ7WrL, KQxhdfmW4VvJpR2C, gN2Fyzbq1VcprRZn, FRPQqv1hYckNTmX7, LkHdMGv19yKfcJjV,
Mg2ay1xDjkYpBqRX, k6VzAfn8MhFWRyxt, AyTmpNtxB9jbnJ2H, YqckzM14fbBKjgLy, XhVzD6vMargZLRqN, bG6cFAwXx1ZBzv4H,
c1VgMkF9ADLz8pT7, 9H6wdNRnY4kDBWTj, nQxb4MVAvyRjgzaD, aDzbqXNJgCVBAy8c, YF8ANJMRzCBvxpaj, jRxQ32n6WNtGb4ZP,
WHTdJgckPGAZNKmn, 3grN8Rct9Xzknm1H, CJP1Q8fjwctVbZMk, w9VWRQKpJdY2LZH7, LkJxahMY9HdXVFpP, 3FrDjnQz7Bhcb4Gt,
wAtGgFBMX9NP7a6K, Y9hCFm3pzNawtR4c, MFkW1TN8B9HcV3ra, Hwq97xLh6gv8ZPrY, JgABh2fnDwca6Vyb, A7ahGYbzQ8M1ZtF6,
vRxZJDw9XNMjthGz, xAh4LpwPJv6VXTWM, 8FkLHnc1ZBgJd43v, amVZMCTFg3v19NqH, MA3Gd6NrxRgVpBKj, KYA3cqNhkBfVCRw6,
Y83XcaZLd72qgBTt, VT81WfMQFg4z6Cyj, vLXN64YjZCrDWfVF, D73VgQNkpc1CZHrn, aHV98R4YzN6Qw3GF, z4WwnT3jGKx1rL7M,
JmTh2g1fbrqZaRwH, 3YTNbcGaRrZ7wtPD, DPAT4jfWF7NzyCGV, hW3bPBvwFK6LMptc, 8xP6yC7rLVHtdZfb, TLCvqN1QKXFM2jyH,
MpZ9JDK7hmdLcQ2y, XCK7MQjFV3fAkBnt, f4JVzQraN6tMnBdL, FzLwcpRAYGXrQqj6, rfR8bFWCq6dJ9TyP, txD1Vbm2ZNyXcjPH,
ymcTt8vprBnJFC9H, gZyFXbh3tWKQ7RrA, QTgJcaxWk91ZmG3z, THkQ2FB67cY3pdKX, rT2K3ZYPCaLFJy91, wgytGazprHPFKmJh,
rPGVHTDmaK1kZ2dC, q2chJNK3dnB1kVwt, f8DWjM36LvFwmZ4t, rPpHYfq3tNnwJRMd, nAVx6ZgkQywNd3hc, 3vD7WzBAh6da8YFM,
byvnpga9tMq3BfX6, btHDxFPjAgCyzLc7, NXMZV9xjvWpB7GtP, NM3dZJwRGKhknfyj, q8LdmM64QwVWB1Ty, 7XK1DyrJ2TfAQ48p,
fkRMFgKNQ2zWXZxP, wJZ9HYF4QqGmDNAy, xaDZ8hAXfd2vVGpL, TV62dgNhz7DvjKXY, Rj4nGwXNAKybxfJ3, 6TcNHMBR8afV3jq7,
bBW7XTh6Q2GqjCDM, 176QTFWR8PtdfGZB, qzDjCk2XQ3tnM6ZT, xmfJ61A2R8TNvQH4, pwzCJQZXDcLT39g4, 1TZvYG2tfR9wVPNH,
7dAfnDaXT2yCgvJt, RzGqJPv8dmr9w7LA, dnXbDqMcL3pT1mr6, ghJGFHpQ4d1bfWvk, qTBCMLmWzVvdGRaJ, VYBjgJx8aZtRhmpy,
LN78yvQwgJpcAYCG, CBpMcGAV4PqFNa6R, zMJvHPcN4GagXkbD, tJTjwYNRHzPrW6Q4, JW9XFvbqzNy1Twfx, c16KTAYxn3DazGMg,
mkcdYM6FVCyADQPw, RfT7K4v2jhY9FAzD, b9gAB3nH6qVwpMRC, 1ZyhaVDdFtTYWMcC, 7dHWkLBFXx83mhZf, PxHCpLJfwA7aNm6G,
vDwzQTXqM9rxH4nV, 3LwjMABn6RCd1Kz8, KFTamfryP7pXMcq2, gwTqKvMWALNRZHFC, 9VQYzFhgGrxvd6KT, TkdjxNKfy9wWDhQp,
82rgw7pCqXVbJHMY, KDnJkWbL9aCHXc28, zmTnKw2xP36gVvqY, 123jYCvkVGrHfXWt, JCpQhr8ZnHXkc97L, dxaHvw9j2cpZL7TQ,
ra3pYvDdLKwFGB14, GtqhgXLQmAVC3FPH, NkMFKRYT7pgCJahA, rAFM4QTbC8jHypZL, kr3AzhZ9FVB1nG2D, RTfqrCdVvLGYzj1A,
B82D1xdY9pJWLMPy, 9T6Zbt2QxPXhmJpk, nazyAb3mjf7H1PZ6, 8YLV79MjFGtR2QWc, 8NMWGB1n6zdLtPrY, M2qWf9ZyVKJxnzkY,
AvNwTdyBbDncFf2h, Hp9ThMznR74PwYKb, 7tLY1pVK9wRWPABX, TDjrhyCkw3P9YApc, 4nkW7JDxV12waMfr, 9F6KjAwQNzdGgr2m,
bnAT8zHVrMPctfDB, JNL9n3P8R6kMr7FG, XWxyBn1jPd78wALt, HG4YQy7Jf3xnFpLM, PC9Fd2RGr6azgnBb, bdYgfJk7Mh8GqZ4t,
xQHTp3ywBbRkjd1r, 9nJGwhXTd8pvraxb, 9zxHaWbYT3gqDQvF, fVg1x2hdaRGKQXyr, z6F71CcD3BvNnwJd, FVqQab31p7KXAJzt,
vwCXzhMQ2drgpRxD, 9Q2zVMKNrhjtPqaY, z9XbfMRh21Fg7QjN, x3njCmpkVYf7WDgh, tgXLJCypHzh8cZd6, JqYpbaFxvLTWtPKN,
3npHxB84DT2zRtmN, r38LkvhBFVH4QqfD, gJtby1TP3nmA2DFp, JCBbc3rj7AdnyRVN, Yv1Vp8KLkgMfWTRB, 81g4MpGPLjJ6x3DV,
4DmjXGdbBQTKP6zr, T8NA3FWvRdhBYyXL, BXhNKykZL7MDat8c, fkGPRpqa7rYzF3AT, ghcTDmb871vtxyWM, Gwj2k3RF6bYpzmKJ,
Avwb6ZLJHQKyfhM9, mJDNgqYVGwb6A8tF, cfjxbzJ6Q9BvtmHK, aFMwqH32y4DtGLTN, McbwGLg8qmNYnhja, 6xHzpKd8wCna37Lb,
JxvWA68g3jTkFdQV, F6ykTXvbMH7mw9PD, nawNKjpbMAqPfDYB, QtHnFxTVRCLPWqgj, 9h3cXRZ7LfM6gPwb, CVfpgcNPrMqXh68b,
Qx6AnCftgKqkN8Lh, 9H2chVMdn1ZtxXpA, Ht6wjv31akZDbgWy, aFq1fMXx8mZ6H7pQ, qftcyJ38hHnKjB47, waFnCzt2chBHVGQW,
yzrMtYV9n6GPfjKa, rTZYqFdvcCPyxWAR, 369Yv2NcqrygTpwX, JQgB3aWcjTDV9nkM, 8hWtw7cnRDPvdzXN, k263KVnBJGFaX7Np,
HxNMb7JVw4pGP6dn, 1Dn6dBtkJXpwGfb8, AdMy8PfxBQKLbnj7, PmfbLqYjZNhrptk3, QtN9Y8vyJa1p4dR6, pdkwZJNqWKCarQx9,
JZmpq6C1PvFh3At4, 67pNx9j3z4XY8Cr1, 6DMFNxznq7jZ4AY9, vdGWF6HVCJ8cgrnL, TKrhPvCj3VRF7NQf, pNW7aGKnzHP1MftT,
ptGqrPfvyBgdmw23, Jk9YT1ZcajwtRfzD, zCyk3LFp4XZwaQBd, yxTNqfGCVcB4RAvP, Z7qF4j3hazBTxHcw, KG62VXF3a97CxtAg,
nmwM719fpABxvjdz, gtNvbhF9aZCXyzVk, HcWVM3NZvGPJ6B49, nNmBkKp3J96FTGfg, NQv1MAX6mrxjC2nb, K3FqtCdPbLYxBzM8,
mbdtyQcGXALMnN86, fwvD2NRWCgtdX1Zj, vfpxJW63nYAHktR1, D463FZBX1C9maJ7b, LbzCGj1yFptQY2Za, PtmYHjkDrqhdR9MA,
J38xbrFytgM9p1LT, wWXLMGQDp1Nm9zrB, M9CxZVg3XBrm6zjk, BpkH47Z2zcYRKnNy, MT3fCDnYLRKXBJNA, rdxLX8Z4aj6cHNvp,
7rGZbTJtpAqnBh2Y, wXV8FMxymK97h2A3, YzMA1vd3BrqLWxhH, Za21hDrqWXn6Rz3M, nrKZtkQhgbBYf7R6, nNapCAWTtqfHZV2m,
z6LMfrpbwhyVCAJK, xbJBZj14NWCXrhAm, 6yHgkT8LqP42NmxY, 4RcnAhpLZ96KfNDP, 6G3gb8DwtJFXm4f1, wWKfXtZn71LNQY3j,
xqnmW3BDpFb28cJ1, 27xDRN6ApzPKQrBw, cHLtXwmpgQaVRxzd, hqvCxV724b3N9DyT, 24KTy3bLGNQm1rdx, R6Bbf3mNT2nAXDyG,
VjypzwHKQ7Y3Jba1, KZ1DBj6kfmt4xLpX, H1CMQLTxXbKyBAWw, TkfYVFdKBth9C3mR, MWrPDAVFCBXzh8Z6, 2F9bfTDrYXzwP6xN,
XG8yAaRFZnYbJCth, 61zfMJDbX8PcZmFG, tkQKL9bGT6JVR3na, Tkc78DB1NqVfC2jG, jvfw4pDgYh62FGxP, 1w3vfpYKqbrBNZPQ,
M7QKVvDpNn9q4GAW, 7fPQd1z38KbxvA4T, 1z4vRDqGdM2cVPXk, C2qtXKpD91F76bBd, tycx8Lp3JWaqgBCf, 2VAWrb67G8KMzgj9,
zKZrVxAHgRTd7fn9, 149hYJDgVqZRnxjv, gmXwaKBfrC7V3jFL, Jt6431MZTKmqQLYF, 9wn6TDHJ7jPpWhF1, rwxL1aVmWyBYA89g,
QGLzd2PDcm763VRr, bcQZaKq8rkMgn467, k28NjXqp96v1HYAc, pHRJTL6NwQjfZ3Wb, H6MV3chRTQXAbgZv, 9TrdzatyJbQV3vmc,
j38paTrnGPbyZzV6, Gky89NAVhKT4zxfB, JVryqmadLFxw1Np2, MqkxfZQWLYJwjgth, vTrR93LjF4kW8Hg7, QHXdBJRj2GPpC3n1,
64CYxjcPzMtd7A9m, CfHvwytKJ6P4q7ag, PCW1JNXVx2B8nDyh, FXBD4HRYbGgzT213, bBpwXG2Jg9HrCPKN, JKtLp9FHxZVhgG1k,
qMQdaN1g43vXpLzk, NGjgKWaxdYn91Bkv, HmF9nyBdrvAL1wKt, zrALwGH82WM6Jcbx, qCzMwt7fy63B1n2k, Pn4rLcAkGagpWDm3,
kaH1G4ztmYpgMTJ6, nyMqbY64tL7HhKQJ, 8ndg4zGq9mfVWvQk, x7ybf4LjBD6GJqWn, 7jpwkL6dfCgFB4Gx, 8GNt2hC7v1rWRXyc,
tZmyBzgxL4C9Kwca, MDvLV1BRcJTaPjFY, TbLC8ZYxRQ9cajzh, ryqQx6DaXNLVfWGt, hpMj86cV3yanwfqQ, 6zWXCmK8F2DgNwft,
TLa4JwDFtzZ78gqf, xjVvrwFDAP4Kp3La, cD6JqWvBhk19xL3r, BRxGa38JnVhPKw4A, gnZqNkAG9JDT2pQW, vqQgaxHKDw6kh4PX,
NYGWBHyC3nDJXpdj, yhVtPGgQpYNrHRbm, jXTVfy3ApWKPYwDG, KpQ8cPBDAyrG7nzd, 9hywC8jMTLV36Pxd, dpVr9znc2yCBHRvX,
HgDdVqRj1YnBPw39, hH4YRGVFz3KnN6bw, d8cAKDNpQLRM9Gva, mb8FgjhwVTLZ4q1d, DWfMg7hzPF4d1mQ6, 3WyYMqr8ZxQRN1vD,
mZdjRTQLyb1hnCY3, K9Nzbf8cGrmywHV4, h3VmpYZ4Gt6xNJC7, WDQR3PJr2XZmp9Yb, QAgrxLyh1CVwcf6k, Q9YA8L6BxM43zm7t,
dTWHj1JVY3mrfXhz, vZqVmLYQa6jwP3A9, 2Z4G3r9pfqJtCM6W, WLKn4vH3VrxfF1Pg, Qagcytmx7nMZTwXF, cWnpRJYDmg21bBZV,
FbDMncd3mB92HYxp, H2T7NDqBgm6LxGYK, bHaxRQvC7VATPkX2, YBDwzFPf9WH7pgv8, bzwY28XH6Vfc4Zpg, C9rN1WjYgdJX4QPZ,
pz4jnqADgcrTFMBW, pqmQRZAcyMa1L98X, fFd4BgHwTaWXqJkp, TM6tCLFdA83Pbczr, PAFTBzCh3pcD72Wx, Mxk96r3XqTv1bLz4,
8At6HgB3hLk2d1ZP, QczFgrTbG4CPpB3n, xQ6W8rHmCn7AcTwz, DNfvpnLRcPJyMKQA, t4zDgB6KpwdRQ2fF, 2LVH4fq1TJky7Dcj,
1DVFRby49Pg2tpzG, 7mnVXCgpyxkbtZGq, pqDYWFBywARJa24t, YHXAcJxWG89pMFgf, c4zqrjN6LgZpAQaw, 2wMznDRvmkpbtHxA,
pByatRLwdAmZWX86, pjRYqGgB2rdvawNn, MCgAavKbRFtLpzBJ, fbRDK6JFGjTaZQmp, BnzXDjYrJagdT3N8, hZGvF2bkQ37gK4wL,
AzW2HyM8xZKjQ14h, 2c19k4Q7nGYHb3Nm, 7yLVTagpRbjYMhnz, xTdW6NpRYKrM1Lyb, ty79RgvAwWdLH14P, byh1C6aYnq7DGwpK,
8YHfcN1BTDCVvGgM, DGFd8KtQk19hXabN, mM94FDqRdAfGpbBJ, AdfMXDLW29k6pGy1, kbPtGMLcJ2Brp9vN, a1GLDT9Z4fvAwX2N,
rJb3LkQVwa2CGNDW, XLvj723CgZTJGy4h, MR8g1PhjbDN3tpkv, vp8tKmzLbdrhq21V, CR7j86cYaNgfDzxL, QADVqHgcjLyvBWP6,
fPbAGZk8qwyJ1mpt, TA3pLxKFXQ9Ckm41, M91YXwTBmQDgzk8Z, Yj2Nvgp3fyhJRx1Z, D39jWq2CaRgbh7Jv, wvDFQB4cjG1dLyTh,
d6abnX7kCPy2rT3W, kJxnvX7wMALVGzFa, XmpdVHMF9W3aY8BZ, 3xLFTfmBJ1ZnYwpt, ZRHAwVNzx7MkyTtK, 7FBaHRhMVKjC1QbT,
8wbVrvW76Ph94q1Z, RNg2fkbr3VcYpx7q, hmDYk8MKdpqx1FzJ, HL38fnFjNXGTmRCa, kdbvhDfNTcjZ2nrC, 6YnFw7gxv4kJAKaC,
VJZ4Q2KLGNqTx3PY, 38bwApgqTf2rh9Pc, ntBYv9gkdcZR7HKL, XpCgBcn9t3GrDxw1, 9DMJ2A7341nTZwXb, 7wQGzqLH6p8hZTVW,
La43JRmZxb2fnWTM, qNGcYfwtZHxWD2pa, 8RcZLHvnaMxqB2m4, 6c2NQwpvf8RDCbkh, 26CZDjqdH93v4bfA, xhpkKbTwa7ZBv3f4,
TmQyqnM6hN217LZC, 4VaPYyXCZBMJtA3x, 4GvM7BJFwRkHqa8L, nJFByMAYgxH93pzb, Vdv8p3ZnTB6fGMQN, nrZ31zYa24hV8wpC,
fmcL7MQK1dFZbzwj, PRLYtk7bhHKAGVpw, h2KnX6VALRZ8xyrk, ZQJXxdPnm8fgY7Fj, Jf9LwyAvhnp2KrVb, QjHW1hvfbNGgcLBd,
7ChZtckBQRfp3KAD, 3qR2PmyNvQk8J4Bc, cwfFWmPT4dND3pL8, j6XFRzHxTC1kvVpy, BQP4H1V9GFAtmxKT, 1hFHVpb7MkJQx8dg,
TtqrRChN7nMvYQm3, KAHZN7fJ2rPVvRLb, KN3PLkDpaQm94M1A, TxKGHQ1tAcqP9NYp, jMNVkXY9nLT8Ry3C, cb9gzXjCK3BfH7qM,
kxjX18BwtZymgzMq, v8L6NkQTzMV2mWq3, ky2zDj3JCp48M6Fd, KANdjmV49Pa1fQZp, RLp3KZHty9cNkv8T, Wv4QYA6C7antxKXr,
WwvKVPA3ZQ42Df18, yhW9cLwxnHNDP2jF, QnDdj3rkpTfBcmGz, C4hdg9Y2tWQXGAaH, h7QgLKtRjF9JckXr, wLcx8mpnGJrj7Xha,
bKrnvdTfCcm7gQDA, gDtFwpGPLM2TyZKB, VTMrXFhcLqwGyf3j, zXm6J8HcCrvBWk3M, zgqYjRFPMTV2LA31, VjrAGFkTvNy2W98B,
P3knGgVw2dyDhQa6, kYX6FqpHthxnKcTM, nTYrM2FLPbcjpJN1, Qb1yWxa8vHCGL7Yr, B9gTmLcFW3R7GAHt, vGtQdg49pCh7zx2H,
w3g8jNXJPhxVrCZz, KfpmMXHQDTcqC7j6, PFY6KBxymcbDqjT7, dvkzFtPH3WR9C4Km, tFm4dXCGDwznQ61a, RB1cnAmk2gfYzFTd,
HKY8mdFptfNkDgwb, xw86k9LyMNF3YHjr, Bnj9K4XaMgV2N7Jq, m8RLWaY7bh9jHKwD, MfvNa4CQbRdJVAGW, bYaT9KXtGwqvV6d4,
XQkavwVbDrLKmhJ9, 3TDwbg41YhvNpPAq, cgMGCdPJQ84wzaBV, mT2F3XZgcnyf7wzA, xvV7DRwgKY31bA9j, rZK2GJjt8W1zw7Lc,
TcLMDwAGYvQk67tC, 8fAbWVqGgZQNaztM, ByRxHQmnzfhbvtcV, qcdyZAr2ganDtxNW, GAw9TxLcJYmqMZbn, h9FNKMfq24Lp6JZv,
VkJBAaKPFGwvq8CM, 2b3Tf7QGKHA1WMxD, 8yPRmZdrq4QhcGn1, qVkdG2t67FhNgw3A, QDadybK3XVPRvwLh, BpyYDt1XxNm4qLCQ,
P8zaXgfTG1NxbH32, 8bWqTVN2QCvK7HPR, Ag69GWdLRDhKCrcT, 7qDZP6zcXy9G2Qvt, mdXkc6xwP8Y9pR47, FN8krKfCb3BXWmwV,
2VpbrfHatzjZFkGP, K1TYnkvJmFfz2xN8, KRjh4aMPQDGJCT36, q3KVJz9bZ6HmBtTg, RmqHtxk2T8PLZ4QM, dwjtaQyAR98XB1nF,
nX1qbatTzGZ3CAwB, CMGKZHRVPwzvFmqy, jXLYwBW7bZ3nQdfJ, rQcmf2q3M1JNp7xz, LKqMfFHzj6ZVvmnN, QXFd74zDW8JZMmKx,
3mwdt9jJF6Tyb1q2, P4bXMwRjD1zyhZdJ, GbvNW3tAP68ZpmcY, Cv9xWtarK3zRLqg7, yHRFJ8kn9vQA4YLr, zM8YtyTd3hpBqNwA,
GRnMHmK69BXfWVTj, 2k1cd6J7KVmN8XqQ, YwhQtdMbNARyqL6B, wRY9NtQp7bPZ1VDz, mbZcgPavhdG7fKDW, wCvhA1WtFnVLB3mX,
9Fz4MbxmnpV1QKaD, tFRwNQZPc6WkYqaC, ry2RGWFxPQvZ1YkX, NPf7nTFCmDq4aG6Z, Zh3vXkjgrwYcNfFQ, 3PJWB1FjCh8kbnH6,
ZDWyQGqFBvagT7kA, vTzn9cX2Nr3HCGkK, r3pTVR7baxnL24Jq, byYP67TH4KtZqm31, TP91AFYK68WDgNVh, 2PZCz9c1HgtrX7Th,
2DC7xQNm9wHJnVF4, 4f2ygdYCzFAvqc6K, KbGTn4273tpaJ1Rd, vwYj6cK8Xt24QLaN, 3Bq6fngvQCh1wyXM, MHvXc1w9jCdATF76,
YCbcQHW378fnMvXA, xvAkYp4FZKh1QC23, zp9yVrtfTNZJCv7g, cABaY1tQxPwHDhMv, FbWLNP8V1QzxXjyC, 8zB1PhRQy64jHkXY,
KDJYHPcVThLfyv67, v1K4RDWbxfcpTVPa, MXwtGVx4Q6p1Kr8P, Krh9MXz1nLtTwaVP, bNYL87fxAQm2hgFW, y7ZnpkdF4hQWH36a,
gRmxdjCHt6XDMBpa, WK4q6NjkrDpZwvxT, X8PH2CG1hMkcfxYJ, qnyPVY87HADk64WX, cxbpCDZRrYkhd9Nv, dpWjv2FnZrRPg87A,
BX8n2xTycaKr4CGL, Ln4CrgN3MZbcGKJ2, rnfRqHt9Vk8g1JB7, nkrMNqLAJczYx269, xdvQ6hbRDmTVaH78, APgTZrbzNvykx1JG,
8tBp2ZK7XT4QJCgD, n24x1KqBamh8gHQj, fZbNXK4Cwax6q2Jc, 6twjMX9bCLcDv1qg, h43GgypRxfbcznLD, JavZtXcdp9WgLQ8x,
tVWcPZ1YDMNFzdXr, fJ3Xy1ktmcKb9F7x, 1rfVNktxLJvdXwMh, gmKvNqHFJLZXfdnP, 2HkdGx7W9684CzrL, acw6RHFChKXdk1QP,
MK76acZnpDCLNAHm, aqB6PkjN9YZmKJAp, NrpM7DJQBKWqbzVF, v23ytJHDz6QkbXnN, GYQB4c79w6TF1nWZ, hc9VBMm83TdrRfWK,
RkD1F2YMK36hGXBN, ATXWQCH9kP7v6rFh, aHKFrZvJgBypW24m, MPgTz8CHZKDjwcBq, gGLmCaHknFdjTtJ9, qGywkYrdHKT8NgAQ,
Q2awcL8VGyB6tKZX, btAc341MaWNQH8RL, CDKVcP6ghZJtv8AW, 3MY1W4gBv9yXakD6, rKMAmp63nLgqyzQj, hQGY6tKgVDBfHbzZ,
yDWdhcfxXp14TRn8, 7fz9Kk8rY6jyQAbL, 7JGFmtxVQMncYTX6, aLtQCxqmgAzhcwdB, mjYGdKhwCQ4D6ATf, TndajWC3c76QymF4,
34AQbC8vq9GTzNnc, 3r7mzpBjFKW9Y8Pf, M4aF6w9Pz7JmrXgW, 7cpmjq6DTBNLwFhG, qVvZxP3gNp12KWyb, kDfrNJXH69h7xnWm,
HkfzanMQcVptAB19, 1xzPrQWDC4pv9HhJ, HqwWnYGKdhpCgZVR, 2mL3ZCv9WKTQ18by, kjGcWvaVNPYwM6yH, 9MwvzA3QN7CtHcyg,
zFXxNAn69HBGRaqg, CqM4Vnh831tFgcYw, ajykLz8m72pGqxtH, 8RWqvN4tjFHXz1DY, xRK37aCvtVfj26kL, PYQf86cTrZ1GjdKx,
QD8ypCXWJcv9Y2B3, BfVjHYvWdKkP8Jz7, J6rB1Y8dvjcqTFbC, cCGFq9PDHpKjtMZV, NLzDkMAj8n2TmbxV, XhGkm2F8rNDv4Qjq,
qW4HRXA31pNm6jZg, hH7f9CjFkQTRJY8r, 1z6PXMGdQBDvrCqx, ZvNFrcnf4qK9tTxV, 16Mk3bhNGfVyRx29, D8JF7rbYQRcWxqn3,
rRnVJhQFa9zf7NqK, MB1yTdQ9xmRXPaK8, zqJTVRdZ4jyCXLFM, pgN9QwMLYVfrqGbv, 9zabj3grQkGB2FKL, mRxb9L2wNkMJTWXY,
ARFXzbqHDpBLJdhy, A2YGVnmxNFPqk7Ka, vny4RK9TkMzwtAQV, 6DZv2FAgWq7YrV3j, Fz2wJmfGKPnvTy9h, cQxMnGKHRyZaNV39,
Z9DMkXc2KqNAxVtr, zX7thnkw1pR6WFyJ, 2Cg8R4ZMz7HV3rfN, TbB8q4RMK3LmYJNv, jrPWFdYfQNz3xG6C, vgrm7nPFN231GYRW,
Pn7j3yhC1vFDKcNM, aANyFDgb69hxqY2w, 8aNhq4DxJjLt9k3c, 7Q6WxzYRwjNqXvgZ, nbLKdYfXHDam79R6, PmR8Lpvazd4q61nA,
7pmhcYx1JNCKvWjn, zWQLAnHTjq4RJCby, Lg8H1XvZtD3knajr, v9CxPFNqVjnbkZM8, dnDg7j9kRw6WtbML, DNc6QALhxmGbCH3Y,
dfD1PZmkLayrtXB8, 7mrf9G3yTVJZxWgD, WYFZHxvL7zg6ydnQ, YFtvD87kznr1CThg, LFNZrg2Cpty63BnK, dc2yWw3v6j4AGCg9,
PDFVdp172Kwz8aqj, kyAp3j7XZ8fqChd6, qnyc1zwaB4FrvM9x, XZgVP8K4kwALRbpr, NCHVzbKtjGYpaJZW, 2ZLD7Jvwfa1n3xyM,
fNxhtGQYCBnPbzcp, RpFQn1ftYCAgMW47, mCqj9fpbthFNGgkc, dCFnA2kaP9MVYNhc, NcxbzW4kJYQ7DGhj, gMKLp31GxfDwdJyc,
qv8daLmJcwDp6MrT, VNpDMaKXy2LHAPZh, gpXqDYBNbxKc1y7W, LKzDgyYcfakX38xP, qVnt8TGrHXY3B2Q6, MjYTvbXQ7AcgR9P4,
RZLrK4yTkzN6YjBb, XBHcxn94waThJg8W, P1fVdygv4GNzZmBL, HKvLrV4f1wgWJ3h6, wCHpdtBrqyFQJfLz, XyvpL8HaKfgYjm47,
fApnWvCdJ3Q9Rjaw, aqZnYb67BPjLD1mg, Gkb3rKHF6V9gz4pR, Bg7J4jxckP6pdvrZ, Jcqz8grLfdPF9NDx, Zxvmb47dcjWCQAF6,
mhZ8WvgabzLwD9FK, Y2DyfTtcBZV4GPgh, TG7ZMLJB1cYAhq4k, 71aKdZW3xFHLCT8P, nR1WGzY8pJKDxhQj, H1q3fJCypdD4cmFk,
ptvYCVPymJXH21kR, mCYWgaFVAdrby9NX, PXzcpwyk8tGhY7Q2, BJvAYckHpKt2LQ7F, BZHxG9m1ky6LhXRF, D2d3hNAbwkPgL1vF,
xGTrF2nMRZAqVLvh, yWFhcNBab8wYMQtG, F2vN8aVjBz1PJ4Q3, XcG96Wb4aLNpkTJ3, hMqNmFQ4t6ZkGaVn, RnNKLbX8hCqaxYgd,
7VhgwfAFkBHrbLNC, RjvDtZMPhLm7g6XQ, Qq6CvRP12rAGwngY, Ja9jwXnRbkxLMy3g, P6GLNM3n27dtcYJD, cnKX3wymHVjdW2fq,
ApJhHfRkMDqFj3rb, aGKC2rxPT7fvFdDB, JVX9x47RhYkTN3vK, mJHPBcAZ8Lz9vVjg, aYBjphvLn8xMK2TN, mvWFndH1Tw4b2VKa,
Gk9fF3BXTzMCWPvK, t3PnpMQx9bVqDjYW, 86hYk1JbGHVpxPNv, m6htAf2g7FvD3WVj, wr3tCAx6cF7YkmKG, 2B3j1wN4FcDdhZQA,
3kwcDX7JjVPdT6Mf, 3aW948VYvBfhLbRz, gBK1ZQAP9aDTMhJ7, zhNBgjHxQX7anZTf, hLxABfWVavkM8HGT, Lb2rVF3gtJXMhQzK,
MwRcK6HVh7J9Ff1Y, bKJcjWDRzvHpVmwa, vP3zD4TryW2FnXQK, kYZpad4rH9h6LGqP, KCYXdTfNgGHczbLw, tkPZYqTRgCyWG3jM,
bDjF6xfrQ8Zz9nL7, N9PGBMt7faY6W2J8, 9v23hzNwDpWM67Xy, W3DMV9g1FxbkAjaP, c8mjK4TdzV2M7XLF, g1qJwva4Ynf2RFVL,
Q62Pv4drJCRpwmBx, zXFrRAP2jamNZ7WY, TYqtvFPVya2N3DGB, FctKwmkXnbpJRDNx, DGkKpLa9VACF7Mbj, zLk3npv9hyCQxXZ1,
G8j6abxLKCR4hpgZ, BMW9qhnLyTVjDpkw, h3P9VjGMZCJWfTcm, ZDhHA6c3n49qMT72, wyr4TzWHCfJhMjvN, zCPWhR3QGcnf6Xkg,
q34T1hKrtaDmQWdH, axbF8PtgpH47v9wf, JF9p148AmDqdgtyb, G1LzWdAk9qDbhJZP, H19jnWMgmZRp2cwr, nj81cxprGKNXBhVQ,
v8thQxgGN7ZFLXCW, fdq6Y4wtvZjyX7hQ, tHzN3kKbaPpVW8TY, 6gnAywbVMGq7vRxZ, j7xyJdLXHZNA6PQg, NW7A8vwpDybJkKX6,
qzDvKLQFr1wf2JxX, a6GB8AKFtcMLTrhP, tNnGBcZfF98hwgX1, q6VRad2JDnpxXYAL, 1ZvM4FVfYhyPkTp2, D3MdKLJbgRmz246G,
1vyTrGKzwQdMRmtb, 4CFdJLNyq19Wxp8R, aDHTPGq6htJFx8Nw, 3Nwp2MvmkjHhA1JT, ky84M92RxVhwpfFB, x9byqam6WYcKFz2v,
8cj6fJ9kYvGmqx3W, ZHpKVnhPDGyaTrvw, nfrq1AyF4HpKxaBY, GKZJBtWAvfPL81bc, 63pnYPtNJk1X7HFB, LKDX29B3FgmPkvTG,
4CtGdr2WLmy8cKTA, FaBKy41kwpxCRGfr, gA94CztvFpWYyJ26, 8h7RnkbYaXT1qyDj, x9v27CNzhRgMkXG6, 8xjZmDw2LKv9bckh,
CFRzVcmKkYLpbdnh, KVHMtTfGqRJNc4kn, d3Tn9a7Qpt6W8VKD, A6YxK4QwLtfGaDdh, R4rLFGNfpg7wxHYC, 6ptjCbz4gVwvcDq8,
ZxtpKYwrh6C9MB4N, Pxn7LYfMX4Qd6vNr, bkDMVFPT6a91AXZq, Anwb28qh4CpmK7Mt, D7r9jqfdHyBYNGtR, KgzYF2b1W6hN7AmD,
YDKqdQLcp3WmRHf9, 6jBzR7Lap42nJCbW, L6h2Q3kMftCpjgVn, MycfAxt23K6Z1a9w, mBfLwRdMJX37HNz4, ykz7Bdp9avqbAjRV,
xhtLrB9MjHqWvVY7, h29zDjKXGxvJfpkH, R9fbLyJAW1tDQnFT, XvhwZD31Vz46JWG8, khdjfcVnRtM3JxqL, bJkCVzwY9dNAn2pD,
3N42rgnkmyYQJLPA, pw9ABtND6gGRmYhf, T2rBvFtkWncADzh7, hnF9yrHdX43YwQap, bx9rg8PMj4ACT6Dp, KP3HfrRdbCxgzV1F,
DdRz8jwpVr396WyF, La3XNmFjMcJkn9GY, ZcFKV7tLHbDR4qJB, VYTnz7v9mbtgf3pF, 9GJNhYFB1QnKqzty, h8vCa2P43nbH7JwX,
1VpHn3kYq6B8JMWw, R2kbrZ4NWwAqLMtT, bWP43DJk2dywtfA6, zM4Dd8QKwBATnJR9, fDtYaBcQMK2dgn7N, XG2JqKDjdacwNkQT,
JBxW16RqGXMPYH3d, 6P7TZL4ybHqGNkpf, dv3cF6ygfawN2BL9, 2qbAr7t6jJpWD3xM, wt12gcL9XQzaDpTR, 3H1fvFaqxRL79PY8,
P3QtFzTLVp1aWBbA, T16aGtPDyxdhb3Lr, mA1ZpGwxPFjLQfWt, Z8RB7wa9YxfLH4nF, 68y1GR9CaVNjgt2A, w3FCT7kNn8avMcRr,
24ZKBL3XhzwJcjGA, pRFBYvC8fVr9mcHa, T4XDxPmV6fKQLp8Y, wXYW7VL6hQnqxk3C, CRBc6yTv1GrnQk9z, 27V6dXhtRxLW8BYa,
D8atXM1Hz376Zwyp, 9zKh6JfAm8pVbdLj, VZfL7tK8q2Pj9Wwv, YBCj2JhfrxNL8VpX, 3tK2WMC1QGxhA4JH, WLb1qyhfkdgFtVDz,
DJfFBc3Rm7Xqnvxr, Xahqp2HFvDMWgjAY, v16yHBWaJZTFfhGn, TatW1Q7Mf2Gg6hbJ, tjcwR486HaA7Zr1h, VL6DAR7kgJdHvfFY,
pMkPWzaC7wf9XZNt, CF1GJR8xjAzb3Zdc, PayThRGFLKmbNcnp, d24NB9HwQz7ybVTc, XBwftbj8Y3zGq7mT, Gx3AXVk9cCaR18Df,
LHq6X4d98xb3QvAn, caPRGzDMHx8Qr93v, 9GhRNB7y18HcVdYm, rqcFQ4m2yawKCNhx, 8TdKckGmy9rAxt74, FWPkHAtQCLGfqKpx,
mbwzgt3VpLXTj4MZ, Vnb3xPaXM2ZK6tWd, tXNHjFKnVfMyGRLb, RD6bnVNTgPhGYCvy, qH7ncQZ4BNMvmhpb, WTJGk6CmbMP1V8Hf,
q6379W8YDwVcZ4Rp, KnYw8BkyhFdtHpVr, VdzNW1LkjM2m9r6H, zCj2h9XKkHwn6qVM, gZNFvX2Bxb49qTRz, thb9WXqjxfpdMnwz,
bGVJm7RWBA4PQZvH, Rmn79CVQwN6qYxJk, 1mHkTJgX8G4NDMPx, NB4ag2CcKryDT91x, RKczQVCnmjXJyA28, KvQYAx8bW9FpVXdB,
H3DB9yJKA4hqw6Lg, mzB4Wj3q6Ktv12nG, TNW9rHBt3MD1X8Ag, fRZG9Ww6rFQcA48p, KZYm38HcJx9wP6LQ, 26AYbFNQzMZCmfkh,
pCDkfLHbYxNyvntP, NTJzVdMZBQDaqG83, pNxkRa8jzG1yDbh7, h4q8kfT9GVmMczbB, gDLz1ZAwkX2YPNK6, 1bgXr3HRqG2cTnjQ,
j1xDTH4N7ncAyJhb, GByQ7g9ZPkqxFJXh, mvkNpHDAwX3qfKYZ, c69WN31xbHBPtmhv, v1TyzhDYNPXgHqmn, ymzkqcaG7HfjCY4F,
7kMdW3TNV16KHFgh, McP6DGYbrFh7zkjT, N2f1naYGDykXzVpq, madL1Art92nRQFTC, ZKphvLAPfVxNCX6z, tyFZPdH39kGq21An,
QCRFhVqmHWMK4Gv7, TqDxtfJKXPCLk3zG, 7BraqHpPmfyRjwc6, 4pynYMkBdVgRDvKH, RP9bjvBAqThyndCc, D46Xm8ncR9WbTFrL,
6c8P4BLaVAj3bpK7, kXAF71Wfy6Tx2hZd, ht9mXk7AM24JdGRH, 39NYyDpaM6LFX7nv, dy3Rn9FpG2m4rZhk, 2WkFQGcDVYwpTMJ9,
1Y4NWLka9pAJbB6G, ZhbCF1LvkQ9gBAMy, vLp6HPt7hg3QAb28, XCJdxAF17zqHMhLY, NwqvTkmaAZdBY4Dg, h1Yc3j9w47qydCWg,
4VMAzfwbmy7CrgFY, ZjTwPR1D2nVyC7xF, nrPC1XpNxQVy23TF, 3JGPcqwy2XbZvkM4, ZQtv6AqYmdJFWwhy, v1bPdgDtXypHncfN,
DBzk8qw4yFWtb9rd, p6mVRBzvWMqnQZFc, AhVvc7rK9dRFQMD2, DmjBrYd7M9VZwCHF, QYvzMfCm3pkRXPr1, QFqCayjpTLM7D3wH,
tkvPjpQGY9Cf2zhT, JQwbjnVYyGcP4rg8, hMv4dJgY3pnyVXTk, 3pBbkwHZcVYJdavj, mLhwHyCBfVx2vgZc, MmqaKyjg38TtVwWB,
4vZxJdFNqTXMRkD8, FphkRr6WvjMmyZAX, v8XzFwL9pbHhBCDg, p4waHBZgjkA9mDLn, 6q1mWrtM7zXFckvp, Wm1r28PDVXtynw3A,
2Lh6QkybmzcYM8Jv, pFH4amZPxkBhyTtV, fqYWmr3cC98AwXaJ, g4KqJfjzTN8AFQCr, BZVyF1nmcbpk6Nh7, R78CFp1D24jAatwZ,
qmfjQ63TN81AnGkB, TV9r768ADt4qm31B, NtxKh7PgTAaZnHJd, XD23HFnxbzvPjACV, TnLAmyRdXvz1cBhC, LyqG8nNCBz2hM3Kg,
fdj2MNckqQ3YDHmR, wdYZbrDagm7cjpK3, dz6baRT1jkPGCNvM, cAGW6Farb9N3d2t1, 7g3jCAv8QdDF2zmL, Hk1mCnrfZ98hJq3w,
TQtaGPBgpC1KZy7c, 1QXAJwhcYL6M7TDk, NK6cCM4pJyRhaAFW, 6X7PdJZTFjmR4Anc, bmyL3XzZ8T2p96Nh, Q2xTrDcZqjB9pA7F,
H2Zg1jRm93PpTqK4, fnLwg8xBhRDrCKGX, RAcx3JtdynDbfgwh, ab8VMcAFtN6T42jK, v19wAVhZBPacHQFL, MPfbhA7yLxpkdtZQ,
8ZdnfAT79zPwmyVM, C7DKyWd1AhNvF9tB, NyYBCtKgGQr8Mnpf, WFHghVcLRtPpw4MZ, ACJhv8TgVcbpRM3f, PcXMfHWbVzaqKmN2,
p4fMBXZaJz8ctTP6, KcLtDQRdjbrxyqXJ, 16GhDNRmVn8KWjBT, aANdrQRjKP7MJnCY, 4ChPnBrVK2fJLMaX, xdFjKnGbC2caDhPJ,
9C31KX8RHfk6hpFM, zDKZYwnyML7WFHXB, 4hHBvakRnT7p1GVK, tmd2HZRwKyMFPQqz, JKavrBGqLD6hVcHZ, ywCbWqZGaxX9Hr14,
mB8CGDhM9Knzqyda, 7kQzFZAn6vapjy2B, F4kZ8RhXvzx1CQDG, WfKv2TphtRrzNA3d, ZJgFdtvGTYQVCqyb, 2npcVzdGrXR6PZv9,
1Yh4QWJARc6Xyxrk, ndZyRFxW9vqhbNfY, 7KZzkbdXjBHGWQJa, z78a2dMLTcfNB4nj, HNtn34KXCMyDrQJj, 3TfqmyvRMQLnH1J6,
z13WA2jMnGqJwRZv, m3ctFRXJaZnrKzkC, m76CKntQZWLYhTMg, jKQDqkwmx46ZLBFg, mZLwTbkM897RJXjn, hZwALnTyCM9fzVrX,
bt9RLg4BpGVmaM7y, wj7k1zy3QLmBNY2t, 3Z1AhpmLF48MK6gy, Y8bALnZFcgK6R1X3, Ya9kKCWRwqjFpgz1, LDz4wTgxcnrbKZ7q,
FHnX82dpGtm6ANYh, h3QfbKVcRLX9axWz, B4QqHnGXwzpRfaYW, vkayWxntqQJ6pfjB, tXTpMhL8BjHG4ZqA, XpyJdxgfcaGQvHjY,
y89cB4pVMf7xJk1g, YyVjhcz81AHad3X4, JfjCVkWmRh93B6d7, GFBCh3VZH7TmAkL4, bkm7pBz4N9JtC23L, XfrWGm96NY1BTMng,
hMdxA2bRGqr8ncvV, XFTxJmGBzQNaCbfg, F6DBahfCjv83TRQA, aWLTAv36Kx9HVPkY, 1Wc4yTf2MQPxwXDa, hVJ6Km2AGkLPdt8B,
ma7tcfxDzg2PRGjk, xtBTbfYpCvFaRmNQ, KN3Q18TayHFAmz6r, zaF4hYrwjyCqHKpv, LQ9zMDqpmknBgAX8, 1hWagLrH3BJYnw4t,
XH1qnQY4mjTALgcr, jfHwWDKtgdFxYznh, pdNRGK6j9rkbyaV4, Wf26RdmcnLDTXwV3, mbpQYGL8jaqgdrAz, CXwPzqjmtcM2hnNZ,
vVdmQNKbtTZfMXgq, dyNbRtmWjgT8xfMB, raWPAQ19tf2NVb47, f6mp4yJZwxAR7NLz, mtYDAGM9p3bqV8Wx, JfwFPZYkBmCWn163,
2wFyAGLf1CqZKnaz, Rr1GThF93wHadWcx, aLwy1zjnrYFhvpDq, WLj6wct4CHavz3NP, c8faRh2FPHkCt46b, 1tC7rqQaBfRxbG9g,
68rzY3ca1GwZQfHq, LN2Vk4nMtBc8vWbT, P8YxXZwmqfzyCMtd, xaKvmGH8b4Pkz7V9, 4Tb63aXJykLnKWgN, JX4afWrdcw8RkHDt,
cKG7yDArxfqhJPp6, Y3M8pwarfXvKhbxL, HYmf1vTW6hVDPytK, JZDFdhkvn3NP6xLm, xXGCDmpRakYWJBFy, QXWcwaGZvCKjJFM4,
BManr8GRymQD7v3d, 9ZTCYrFcDxQ8pk7H, YGRqJZz92X4r1Vhw, hKk1yjGvx9DVnTmd, yYPpk4TqAhjX3xC8, CxYkTP1wrRhd6yGN,
bahZczXqfLk3pQR2, 4vcn3WF1RxXf8VZt, a42r3p6ZxtBTRqfQ, hZLykw9dRnYbtvKP, 9R8rhcAwz6V17fpF, NqfgxbL4yt8pHRPk,
26j47DAXZCgwdxKB, wCrx13n6DFLHgfmQ, 4qx98J2pmryaTb7R, AKQPn8Ha7wG1XJVv, yKRFZQN9CXDw1rYG, 8LFvmxKGD4Pgpfcd,
4bVQqW6vk7fHzyrw, nYHgB2LKbtFDWVGw, L2c7jBZNwFPpb1tV, jgxBZTv69zf4aMhR, 9xpQ267J41GrDaMf, 1GN2LMtwvY4JQHba,
p64RG9PVZcDg1XCk, aF69WRYXAT1qNZjM, 4JYa2mQgpVNMF89R, wkNFMH1nAhZ82Rym, tG4CqkJQvyPT2A1Y, t2yjAqMgTvGZHJLC,
NqKayk8Q9rwWvjpz, 7Kz3LtgfDmbBRGXP, Wza8bdXVRhcgqGnZ, k3QGKZ2jr1ptgRPx, Y8L1VRGqHjKJpbCM, 2PfTacjwkD9xJtzd

## Count of the 2026-11-24 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), wearers of enchant 8217 with a Revelation buff in the
  56 reports listed below, read in two runs: every such wearer in the first 39, and in the last 17 only the wearers who
  cast a channel; and a local combat log of a warlock and a mage, levels not recorded, taken 2026-11-24 21:09 local
  time (WoWCombatLog-112426_210934).
- Game build: 1.60.1.71037 on EU realms and 1.60.1.71091 on US realms, hotfix push 114378 cache. The change is server side and shows in the reports between
  20:53:45 and 20:54:06 UTC on 2026-11-24. No game files row changed: the rows of the proc aura 1248806 and of the seven
  buffs were pushed again at 114284, byte-identical.
- Date: 2026-11-24
- Pull request: #1208

The first run read the first 39 reports listed, with events to 21:34 UTC. Before is the events from 20:00 to 20:53 UTC,
52 wearers in 35 reports; after is the events from 20:55 UTC, 47 wearers in 31 reports. The second run, at 22:09 UTC,
fetched the events from 20:00 to about 22:00 UTC of 38 wearers who cast a channel, in 33 reports. The last 17 reports
listed are among those 33 and new to this count. It read the first run's other wearers as they were. The procs, the
buffs and the rolls are counted as in §Method. A buff episode runs from an application to its removal, capped at 15.2 s after
the last proc. A group is a spell's landings within 20 ms of the first, so a channel tick or an area cast on several
targets is one group, and its first result is its first landing in the log.

Per 5 minutes, buffs that a channel's first tick takes within 300 ms: 0 or 1 in every window from 20:00 to 20:45, of
15 to 52 each; 5 of 49 at 20:50; and all of them in every window from 20:55 to 21:30, 15 to 34 each.

### A channel's first tick around the switch

Before, the buff went within 60 ms of the first tick in 1 of 338 Blizzard episodes, 0 of 36 Hellfire, 0 of 32 Rain of
Fire and 0 of 9 Hurricane, and later in 325, 33, 30 and 9 of them; the ticks after the first with the buff up crit on
every target: Blizzard 6,024 of 6,024, Hellfire 2,101 of 2,101, Rain of Fire 341 of 341, Hurricane 231 of 231. After,
a later tick that comes with a new proc's buff up crits on every target: Blizzard 114 of 114, Hellfire 20 of 20,
Hurricane 4 of 4.

In the local log, the warlock's Hellfire first tick crits and the buff goes 103 ms later, with the next tick not
critting; the mage's Blizzard first tick crits and the buff goes 103 ms later, with the three ticks after it not
critting.

### Arcane Missiles before the switch

Before, of 90 episodes with a missile landing inside, 4 lost the buff at a missile's landing and 56 at a missile's
cast event with no landing.

### Direct spells

The charge still goes at the cast and the hit keeps its crit in flight. After, by cast with the buff up:

| Spell          | Casts | Removed at the cast | First hit crits |
| -------------- | ----: | ------------------: | --------------: |
| Frostbolt      |    55 |                  54 |              54 |
| Lightning Bolt |    25 |                  25 |              25 |
| Shadow Bolt    |     8 |                   7 |               7 |
| Pyroblast      |     4 |                   4 |               4 |

An instant area cast crits on every target, as a cast that spends the charge at its completion does: Arcane Explosion
240 of 247 targets in 67 groups of 2 or more, Chain Lightning 20 of 20 in 8.

### Reports

1ZfhArNnJX3gtPcB, 2ztdfwHnv4PgYAxh, 6dwcWm1rYyhgnTxf, bqafJg6RZAL4P2hH, D3yAqmjMcztNdaRW, Dtc8bkGdZ14v7fFa,
DwKyLCcW9xnX1APV, GtXgCKbJQLxda613, hzQRLkBxWCcdPvD1, jG4ctAQMv9TbrpkY, jPTNyJWtxLzQ8MFZ, JXtPHx4p12WyQqjm,
KJnv23ZtWfXraV81, LwTNrdJ6zjnmCKZg, man3DYTNht7JxFGk, pqrmA6VNzPaRHdCT, rMydHq3vx9wPjmR1, rvWB6DTH4dPRJVCF,
tjYN3m2rzFBp1dTb, w63ZvnMVbhQJ1xKC, WXDVmBJ4afK2Lnrh, xZtGRMyAX2FDP1qY, ZKFfkr1J4XaGP38Q, zRFDJLjX42NPpvVn,
ZzL2CYkARtBjg16v, cHLmxkQT1ZCXbKwy, rNMBcQ4awpV7gz6R, XkTcqKrpMC4L3N1x, abBzgq4AFTM9V3kr, BrMn7X4LPwc2DgN6,
dBDT9XHZPvwJ3zMR, gnThGBckVLrFDp8A, Hy9FbkDXzLRmZBAc, ktXHy6ZJL4pMwBqA, R4y92czhFrkmtfYA, t9PLr4H8gdkXncRZ,
6tGyrxPaNM4VQBvq, aFkCTQyW6HxYtN7Z, R3JMNnkVWCDwxyda, 1ZKkyGCWfQaXMvw8, 4JP8jzWgyApchHmb, ARF7PMY3Hp2Df4zK,
D8Jm7wRdZVnB41z6, JFzM8qt3pDGvxmwa, M7rbCABmXtVTZH3v, nJhyt3PkFzB9Gx7K, Q9BKhWPpadRNXrf7, RckBpyKaVChNjD8W,
t8gdhDXHfaRGpWVZ, Tb7Pqt16wy2G38kj, wfPqRDzW1ayK6HM4, XzhM6rRYgv1DxGtq, ybrZWq943vHfhXTz, YXbnwGzahjNCZygW,
yZ3xq9jPwTDNrVXM, Z1hd6JzHVxmPqbtX

## Count of the 2026-11-24 to 2026-11-25 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the streams of wearers of enchant 8217 with a
  Revelation buff from 477 reports: the 56 §Count of the 2026-11-24 reports lists and the 421 listed below. 502 of the
  streams, from 295 of the reports, were fetched from 2026-11-24 23:30 UTC to about 2026-11-25 06:50 UTC, and the rest
  on 2026-11-24; a wearer's run that two reports log is read once.
- Game build: 1.60.1.71037 on EU realms and 1.60.1.71091 on US realms. A second server change shows in the reports between 01:12:49 and 01:12:55 UTC on 2026-11-25. Hotfix
  push 114389 adds the periodic harmful proc flag (PROC_FLAG_DEAL_HARMFUL_PERIODIC, 0x40000) to the priest's and the
  warlock's buff rows, 1323377 and 1323392; the other five buffs and the proc aura 1248806 read as before. The hotfix
  each player ran is not recorded.
- Date: 2026-11-24 20:00 UTC to 2026-11-25 06:50 UTC
- Pull request: #1208

One run at 07:58 UTC on 2026-11-25 read every stream. Before is the events from 20:00 to 20:53 UTC on 2026-11-24,
between is from 20:55 UTC to 01:12:49 UTC on 2026-11-25, and after is from 01:31 UTC on; §A channel's first tick spends
it and §Arcane Missiles end between at 01:01:10 UTC. The procs, the buffs, the rolls, the episodes and the groups are
counted as in §Count of the 2026-11-24 reports, with three changes. 45 more ability ids are read under their spell
(8289 as Drain Soul, 1278965 as Hurricane, 8418 as Arcane Missiles, 1282383 as Rain of Fire and others), where the
earlier runs counted each as a spell of its own. A proc is credited to the wearer's latest landing group 0 to 50 ms
before it, or within 100 ms for §The first result decides, and a heal the wearer gets from its own spell's damage
(Vampiric Embrace's) is no group. The first landing after a gain skips another spell's periodic tick, and a buff that
ran out at 15 s counts as ended.

### The switch on 2026-11-25

Buff episodes whose first landing after the gain is a tick of Mind Flay, Wrack, Drain Soul, Drain Life or Starshards,
per 15 minutes: every window from 20:45 UTC to 00:45 UTC has none lost within 60 ms of that tick and 3 to 30 kept
past it, the window of 01:00 has 4 lost and 8 kept, and every window from 01:15 to 06:30 has 1 to 27 lost and none
kept. The first tick that takes the buff logs at 01:12:49 (6RqzaWpYKx7XgjNr, Mind Flay) and the last that leaves it at
01:12:55 (CVKkd78thnxJPFgW, Wrack). Seven reports show both sides, and each one's last kept and first lost tick bracket
01:12:55. The last kept and first lost ticks on PvP realms (01:12:55 and 01:12:49, 60 reports) and on Normal realms
(01:05:33 and 01:14:03, 61 reports) agree, with no sign of a roll-out by instance or by realm.

### One roll a hit event

Groups with the buff down and the share with a proc credited to them, by the number of targets:

| Spell                    | 1 target            | 2                   | 3                   | 4                   | 5 and more          |
| ------------------------ | ------------------- | ------------------- | ------------------- | ------------------- | ------------------- |
| Blizzard, after          | 160 of 1,815, 0.088 | 441 of 5,374, 0.082 | 774 of 9,950, 0.078 | 670 of 8,018, 0.084 | 746 of 8,974, 0.083 |
| Blizzard, before         | 10 of 89, 0.112     | 35 of 306, 0.114    | 78 of 329, 0.237    | 91 of 314, 0.290    | 129 of 302, 0.427   |
| Hellfire, after          | 83 of 1,064, 0.078  | 197 of 2,537, 0.078 | 358 of 4,479, 0.080 | 345 of 4,604, 0.075 | 309 of 4,249, 0.073 |
| Hellfire, before         | 6 of 59, 0.102      | 15 of 134, 0.112    | 56 of 217, 0.258    | 93 of 349, 0.266    | 124 of 290, 0.428   |
| Arcane Explosion, after  | 51 of 716, 0.071    | 124 of 1,772, 0.070 | 221 of 3,265, 0.068 | 233 of 3,130, 0.074 | 232 of 3,269, 0.071 |
| Arcane Explosion, before | 5 of 49, 0.102      | 19 of 132, 0.144    | 54 of 243, 0.222    | 52 of 228, 0.228    | 89 of 213, 0.418    |
| Rain of Fire, after      | 16 of 222, 0.072    | 45 of 522, 0.086    | 85 of 940, 0.090    | 64 of 899, 0.071    | 120 of 1,476, 0.081 |
| Rain of Fire, before     | 2 of 27, 0.074      | 8 of 52, 0.154      | 28 of 159, 0.176    | 43 of 168, 0.256    | 80 of 228, 0.351    |
| Hurricane, after         | 50 of 752, 0.066    | 95 of 1,392, 0.068  | 110 of 1,954, 0.056 | 133 of 1,892, 0.070 | 155 of 2,330, 0.067 |
| Flamestrike, after       | 32 of 439, 0.073    | 72 of 978, 0.074    | 126 of 1,716, 0.073 | 100 of 1,450, 0.069 | 85 of 1,399, 0.061  |

Before, the share grows with the targets as one roll a target gives; after, it stays at one roll's chance. Holy Nova
reads 43 of 265 (0.162) at 5 targets and more from 5 wearers; its damage and its heals may fall in one group, and this
is not checked.

### The first result decides

Groups with the buff down of Blizzard, Hellfire, Arcane Explosion, Rain of Fire and Chain Lightning together, by which
of their results crit, and the share with a proc within 100 ms. A mixed group has crits and non-crits among its
targets:

| Group                        | After                  | Before              |
| ---------------------------- | ---------------------- | ------------------- |
| 1 target, no crit            | 320 of 3,571, 0.090    | 28 of 220, 0.127    |
| 1 target, crit               | 0 of 428               | 0 of 36             |
| 2 or more, none crit         | 3,626 of 41,522, 0.087 | 678 of 2,396, 0.283 |
| Mixed, first result not crit | 1,469 of 16,284, 0.090 | 273 of 1,027, 0.266 |
| Mixed, first result crit     | 4 of 7,172, 0.001      | 90 of 472, 0.191    |
| 2 or more, all crit          | 2 of 291               | 0 of 20             |

After, the mixed groups whose first result crits proc by spell: Blizzard 1 of 2,964, Hellfire 1 of 1,508, Arcane
Explosion 1 of 2,033, Rain of Fire 0 of 357, Chain Lightning 1 of 310. Of the 7,172, 9 have a proc within 100 ms: 4
Chain Lightning casts whose overload lands 16 to 32 ms after, 3 with no other landing and 2 with another spell landing
beside the proc. Before, the targets that did not crit rolled in those groups.

### A channel's first tick spends it

Buff episodes whose first landing after the gain is a tick of a channel whose ticks cast a spell of their own, after
and between:

| Spell        | Episodes after | Taken within 60 ms of the first tick | Ended before it | Episodes between | Taken within 60 ms | Ended before it |
| ------------ | -------------: | -----------------------------------: | --------------: | ---------------: | -----------------: | --------------: |
| Blizzard     |          2,695 |                                2,613 |              76 |            1,934 |              1,723 |             206 |
| Hellfire     |            843 |                                  817 |              17 |              765 |                743 |              21 |
| Hurricane    |            507 |                                  492 |              12 |              689 |                659 |              29 |
| Rain of Fire |            285 |                                  261 |              23 |              304 |                282 |              21 |

The others outlive the first tick (Hellfire 3 after, Blizzard 1 between), go 60 to 300 ms after it (5 after, 2
between) or have no removal in the data (11 after, 5 between). The tick that takes it crits on every target it
strikes, after: Blizzard 11,359 of 11,362 results, Hellfire 4,033 of 4,033, Hurricane 1,907 of 1,907, Rain of Fire
1,055 of 1,055; between: 6,752 of 6,752, 2,863 of 2,863, 2,633 of 2,633 and 1,184 of 1,184. The ticks after it with
the buff down crit at the spell's own chance, after:

| Spell        | Later ticks, buff down, crits | The spell's crit with the buff down |
| ------------ | ----------------------------- | ----------------------------------- |
| Blizzard     | 2,822 of 31,473, 9.0%         | 13,423 of 146,021, 9.2%             |
| Hellfire     | 1,399 of 13,766, 10.2%        | 7,870 of 77,726, 10.1%              |
| Hurricane    | 926 of 6,124, 15.1%           | 4,912 of 33,180, 14.8%              |
| Rain of Fire | 96 of 925, 10.4%              | 1,747 of 17,562, 9.9%               |

### Arcane Missiles

A missile logs a cast event as it launches and its damage 0.1 to 1.5 s later. Missiles pair with their launches
channel by channel (39,049 of 42,762 damage events). After, 1,235 episodes have a missile as the first landing after
the gain: 930 gained from a missile of the running channel, of which 404 lose the buff 0 to 60 ms after that missile
lands, 401 of them on a crit; and 305 gained from another spell, of which 249 lose it there, all on a crit. By when the
next missile launched against the gain:

| Gained from              | Next missile launched after the gain | Launched before it or in its millisecond |
| ------------------------ | ------------------------------------ | ---------------------------------------- |
| A missile of the channel | 305 of 306 spend it, 87 wearers      | 70 of 543 spend it, 100 wearers          |
| Another spell            | 192 of 192 spend it, 82 wearers      | 7 of 57 spend it, 38 wearers             |

A missile launched after the gain crits and spends the charge as it lands. One already in flight at the gain keeps its
own roll and spends the charge only when it crits: those missiles crit 66 of 543 times, and 70 spend it. The launch splits them at the gain's millisecond: launched 0 to 20 ms after the gain, 17 of 17 spend it; 0 to
20 ms before, 2 of 21. The in-flight missiles crit less than their wearers' own chance with the buff down suggests,
66 of 539 against 96.5 expected (96 wearers), where between they read 97 of 431 against 86.1; the cause is not known.
Missiles of the channel that land more than 100 ms after the one that took the charge and before the wearer's next cast
event, so launched before it landed, crit 26 of 41 times with the buff down, against 3,365 of 18,407 (18.3%) for the
spell.

### A channel of periodic ticks spends it on its first tick

After, buff episodes whose first landing after the gain is a tick of a channel whose ticks are its own periodic
damage:

| Channel    | Wearers | Episodes | Taken within 60 ms of the first tick                  | Ended before it | First tick crits | Later ticks, buff down, crits | The spell's crit with the buff down |
| ---------- | ------: | -------: | ----------------------------------------------------- | --------------: | ---------------- | ----------------------------- | ----------------------------------- |
| Mind Flay  |      27 |      183 | 183                                                   |               0 | 183 of 183       | 20 of 259, 7.7%               | 1,576 of 17,394, 9.1%               |
| Wrack      |      35 |      110 | 95                                                    |              15 | 97 of 97 results | 46 of 269, 17.1%              | 2,951 of 21,871, 13.5%              |
| Drain Soul |       5 |        5 | 4, the 5th 207 ms after a tick that killed the target |               0 | 4 of 4           | 0 of 3                        | 52 of 420, 12.4%                    |
| Drain Life |       9 |       13 | 12                                                    |               1 | 21 of 21 results | 3 of 26                       | 131 of 1,197, 10.9%                 |

The tick that takes it is always the first of a channel cast after the gain, and the buff goes 0 to 34 ms after it,
11 ms at the median. No episode gains the buff while such a channel runs, as its ticks roll nothing, so a charge
gained mid-channel is not counted. Starshards has no
episode after: 6 casts by 3 wearers, none with the buff up.

Between, none of these episodes loses the buff within 60 ms of the first tick: Mind Flay 0 of 210 on 30 wearers, whose
first tick crits on every target in 207 of 208, Wrack 0 of 67 on 34 wearers, 27 of them ended before the tick, Drain
Soul 0 of 5, Drain Life 0 of 8 and Starshards 0 of 2. Before, the buff went within 60 ms of the channel's third tick in
19 of 22 Mind Flay episodes on 4 priests, a tick that crit on every target, and of the sixth tick in 3 of 7 Wrack
episodes on 5 warlocks.

Another spell's periodic tick does not take it: of 2,323 episodes whose first landing after the gain is such a tick
(Curse of Agony 541, Corruption 260, Shadow Word: Pain 207, Moonfire 199, Immolate 170, Flame Shock 132, Insect Swarm
122 and others), 70 lose the buff within 60 ms after it, and 9 of those with no cast or direct landing of the wearer
at the removal.

### Penance

After, 26 buff episodes on 4 wearers have a Penance heal as the first landing after the gain: 23 keep the buff past it,
2 end before it and 1 loses it within 60 ms after it, the cause not checked. Its heals with the buff up crit 58 of 58,
and no Penance heal takes the buff: 13 run out at 15 s and 12 go at another spell. Between, 5 episodes keep it past
the first heal. A Penance heal logs as a heal, not as a periodic tick, and rolls the chance: 30 of 507 groups with the
buff down procced, on 5 wearers. Penance's damage is thin: in 2 episodes on 1 wearer whose first Penance landing after
the gain is a damage bolt, the buff goes within 60 ms of that bolt, which crits.

### Lightning Overload

An overload logs under its own spell id (408477 for the highest Lightning Bolt rank, 408481 and 408484 for Chain
Lightning). A bolt's overload lands one more flight time after its bolt, 14 ms from that at the median (1,850
paired), and a Chain Lightning overload 16 to 51 ms after the chain. A cast with another spell landing within 50 ms of
its landing is left out. After, with the buff down at the landing, the procs per landing:

| Landing                         | Procs               |
| ------------------------------- | ------------------- |
| Bolt crit, its overload not     | 16 of 245, 0.065    |
| Bolt not crit, its overload not | 34 of 612, 0.056    |
| Bolt not crit, with no overload | 204 of 3,258, 0.063 |
| Overload crit                   | 0 of 212            |
| Bolt crit, with no overload     | 0 of 1,029          |

A bolt that crits rolls nothing, so the 16 procs of a critting bolt come from its overload's own roll, on 8 wearers,
7 to 35 ms after the overload lands, one of them with another landing in the 60 ms before. Chain Lightning's casts
are fewer and do not settle it: a chain that crits with an overload that does not procs 1 of 44, where a roll of its
own at about 0.07 gives 3; a chain that does not crit with an overload that does 27 of 61, where the chain's roll gives
the buff before the overload lands and the overload then crits; both not critting 21 of 152; and a chain with no
overload 67 of 938.

### Ticks and other readings

Periodic ticks roll nothing: after, 78 of 223,281 groups of periodic ticks with the buff down have a proc credited to
them, and a tick that lands up to 50 ms after another spell takes that spell's proc; Mind Flay 0 of 15,812, Wrack 0 of 16,198, Drain Life 0 of 992,
Drain Soul 0 of 348, Starshards 0 of 24, Tranquility 0 of 32. A channel's tick that casts a spell of its own rolls at
about a direct spell's chance, on groups of 2 targets or more with the buff down and no crit: Blizzard 1,978 of 22,135
(0.089), Hellfire 884 of 10,513 (0.084), Hurricane 312 of 4,058 (0.077), against Arcane Explosion 457 of 5,376
(0.085).

### Reports

1CqLGARDZkPgbvfT, 1FydH2zLq7PhwBQX, 1Fz28TJwRaNkbx74, 1ZvtLH47RNMm3Pfp, 1cGX3fp7CamP62Wx, 1dXDVmCn82Ky69rZ,
1rzF8qV3jYwLAXZk, 1yxBqh9pwYv6Z7kH, 27LWXV3PAQabzw6C, 28zZgXVaFPvmHTM9, 2FAZ8pGqL9JvfTRg, 2WfkT9ZRCFYvdyQN,
2ZK3wq1d7MWmvbhC, 2fNJjyTc3C4ABxVZ, 2hcGL8FdDgK9xbrn, 2zJ3GkaPLDnrKBgQ, 32T1KMZz89kwdqFc, 3F4716PHR2fNrkcq,
3GHvV4hmb96yWB7a, 3LXKpR8kYbWactCm, 3LntpD2XGQfb4jV9, 42XtM7Jwdk9gnvqW, 46vZ2R1qTkyj9dHN, 4A8RTXLg32nDkwhV,
4MmPTkfjFHbYZKny, 4QDbGNfq7Tt6Wv2a, 4cAnPtXrbLxavk93, 4ky7qHGcVFP3QTaL, 69JqBW7kRr4FTyCP, 6C8bWZy3DQwKzFM2,
6CkrBFvzWc72Pgnd, 6FZNWyXGzDxHqv2a, 6PN8t7yDdwrVncJb, 6RqzaWpYKx7XgjNr, 6WnAcHKD71rm8FNg, 6YGCTMrthm1djgVz,
6YZpvDf7xPtaHGdC, 6hGK4Dkv8CyMmfLw, 6j7Ntg1VcpPZzrFf, 6pyjXZhtkWJLmFTM, 6wFAMdmGYJpDj2HZ, 7XRjxY6wkPbpBzrv,
7dBrzHDjR38Mg4yA, 7dhc14nJpBR2YjyX, 7gCbcw9nrFGZJW83, 7h86LGqYKzdj9QPD, 7nyCwGzLBKqR89xV, 7rbPgnDH8J4mTjwV,
7wNAZfFxb1rWmX6L, 81DqGMFYnZyRaVtH, 8CnFhxzYJGK2RbMN, 8FWjnvVz6Dx4kYqy, 8H7Xx6J3brhPDjat, 8KQqLYkfhXBrRjyv,
8PWpYzaDMxwHN2QB, 8dDFP2W76vHJV9QL, 8fKVGRP6JabCTg4z, 93bZ2gByRhHQnLTN, 9ZxdF21LmnDa4wXQ, 9gc7TxLjZhYJVmBA,
9tgbwD36PjfrncqR, 9w4ZyL23P1nhpca8, A6pmrCYMfGjWK12z, A9jRKpTqfvJynhxt, AFTnz1NcdZjg3qMf, AJCFnYvZ4k8gfBwd,
AK2QvbhTxdgJLC9n, AMR4zCvWQYpH2LFB, ARwGx6VgavfWDQ4L, AWDftL9aJZmMBxHp, AgQXzkmVdFcDCJjR, Agk16w2fKZrW9vqt,
AjPcMbG32rNLgHZX, AjWBvdLHzwRMT1kQ, AtT6VF3hPC2ZvQz7, AwLpv269JFtdQhcf, AzC64VwWcap9mNkZ, B16hLXQkryWKFRcb,
B1aX2vjZQpdYWLgH, B4pz8Pnw9xt3yMgJ, B8YhdtQbP7pGLnqM, B9z43TwnWCjHFDgL, BHT9zJALK4CqP7rv, BLFWVRbKH3ZMTcYy,
BRqbW8d4r2TDgJtj, Bc4xdN9pVZ2y7WHG, BcXxkd4J8ACDvgQa, ByKpfCMF79xd2b6N, C1gLQGwKhynD7vkj, CAPt7qmH61WnKTMx,
CQA3rBFw9g21HLct, CRBZjnQmAVxbP7c4, CTRf8tF4YbwgGkzx, CVKkd78thnxJPFgW, CWXBLFfMDAcxN8Jh, CgdpWP2c1XtVJyma,
CjMJ3LB4kvwWpFAf, CvGk7b3QgfDP92Hn, CxdWQMGavTmFHLcf, D8Apm2FQwW4tv1x3, DB3yYHG8rcQNbpjf, DG3nN14cK9ZBJg6H,
DM6Y8kJHx7djwBCZ, DRyzwVQch9aGtJFZ, DTd7XJrKnjqV6HmZ, DkgbqPc9LhRzvQtT, DwHL1AWJTaRfpxzk, F24npWZRxtv8hD76,
F3YnkGDzhZXVRrBK, F8PLBprjC1xdztQG, FNxMyCpTLDgc6qv1, FRWwdT6ajk4J2tVf, FkPczLX6dtYpJ8j1, FnJqWC7tP3HMdjz9,
Fr2YcVMyWAgQv3XH, Fy47gNjCAGLdM86k, GLrQ4hatzpbYc2CW, GT1hZgpaQfbBCWvt, GXNt7bdaC4nFADHr, GcLFMVzHmqhYB1Q4,
Ghzvw7mRtXB8HkKM, GnZdPBVXFW6kwp78, Gph6gaZYMz92RbcJ, GwKmMFyX3tT4LjHD, Gyq7vrzaNp86Df2J, H3TqYxR1k4phg2yw,
HFdPMCzAGqphb7jV, HKbAp9nmrwkC8Y6f, HNDXaqgxpjQznYL9, HxaCdGfh98cPqWFr, JCKxQ1PgBdv3T4z9, JGARfPNCHQ4t9jXW,
JNTYBgF8W9tp2afq, JRnh1y3aWrdx6TMq, JTzg7HVqtNhQD24Z, JYxQXLkWCqGdjg2c, JgTztkD9c3CFYBAQ, JmxhYRVQvN3b7LTw,
K2hfvqXYdxJ73DPZ, K3fTgQYRjdJAqacz, KAgb1CFjaxt8VwXf, KDAXQbt3cFJwpL9Y, KZC69BgP8Gjbxn7r, KZNdpF4HwrXVbMhf,
Kdw6P2fQYBnNcFaq, KfgyHLbX6Zr4QTFa, KqjwWn3L4VGxrcgT, KvZ7f3GLJQ9MVmPB, KwjknzZ8PNcCXvVF, KyRzjkdcPghabTFJ,
LCqYwgcpM4N9vJtb, LDjcQByk8Z1WqGbK, LFJ4WwtvaHbBjQhq, LQGDyagcxNkrqpVC, LYAyQpRDbCgXF9xj, LgcdyYf4hz13p8Ba,
LmWCg9VM2zwtAbxT, LrNwTH789mfYqVgA, Lrpfd2D6NCA4a83B, LtmGvDKfBX4QJ6Ac, MCxV7YzAHgK3tZQ1, MYwn3QcLGk2dv6WT,
MbmKz9aTvV2yX86B, MmPtc9xH3D7gXChz, MrHngb6GtXvBydaJ, Mtpn21bD4ArgcxjL, MxGmyTrH82zNPw6v, MyVQdCAKLPBFrmfY,
MyZGhvYbCJVtdj7H, NDAryTb6PpkHKM9F, NGzfyVaXFZc3xtwH, NHy97zmbJ3FQWMAr, NLgGJhYQZyPWvdFz, NLzRZTDa2CHnk9X7,
Ngcb48qvWpQRMtfh, NjA7nRMvt2y3BqdQ, NkhmzpbxcP7nC3t4, NnMA6GgHfRwLJ98Q, P1xTA3jMXHybdLNp, PBJW1T8KhpnaZFRw,
PRjhkcVtAGxMbNq7, PyjdHqGZ319xFvMf, Pz1mrwty3JaAdT9n, QBN7qpjxz8fg1ndK, QCfgHym3rT1dnXvp, QFfw4GAP8qtHgRjc,
QJdHMt2h7Bx6pk3F, QNKD3HnVZFApMajY, QTLdrxZXFmPJYv9W, Qjxg7zvYFnKGDBWZ, Qk3DbRw6cpVmZLKP, Qn4dj1wYXBqD3m7L,
Qqyd3T9HFAfK4vjM, Qr2hVGLaFDydjxNC, QvWPdyTnGJrg1LRD, R4bh7219AjnrNv6G, RB1fCw2AbhVnGgHY, RCzf4xwT91qbMVn6,
RNMFW6PJgty9zX1H, RVtmpwqLCMrJKajH, RW91FgNq7MBXJjGt, RdqHgVKWNmTCrM8Q, RyFPnfw16q8Kmtxk, T9Q7zcVmL6DAtxkw,
TB3XRpkHLaMhFrAW, TCVybQX6LBa3j9J2, TCWgZG1HRthpncPY, TCwdgVZb7J9kaQny, TJcwZN8d2rqknGCg, TRDHw9q17a3dPBmg,
TY8hRktpNaCnDXM2, TaCQNVZFp4MX67Yw, Tc732j1RgGp9xPtW, TgQJ7t3wHrzPqM4C, ThdntKDCMJc1wPkV, Tk17qAa9QyWMwRPJ,
Tn7YDLcxwgpHqP84, Tqp2f38JaArcjDy6, TyGawFf7MhLnB3Dv, V16TNpMCdKaHqZ2m, V1RFaDpWtq6H793m, V4vAnTpmG3jLMQkq,
VF4Xdf1KwWjvkmRa, VGqvRAaDPXFkwbzm, VKX1hJ6QZPF2jmbH, VaTW6Kprkfwg9DRX, VpR1MyB7gNHwn9t6, WLw7gXz9A2fBdD4y,
WMT8mt7CrFxcXaGw, WN8J2FCYbKq9jmL3, WTLqhMGyRAQcBwf8, Wac1vKTnzyQXgBCj, WcVyhzZHXntRwdNT, WpL8cvaMrbgFRYyn,
WvkHYMPLg917jKwV, Ww7g2kXh3ZMYHLcd, XCR9xc7JmtHy81p6, XF3aTZxCrpBjPhR9, XwPJHVmAjQh3YTF2, Xz2TRxJf4WHabFQw,
Y2nqwpmRTPL8x37z, Y36rGfA2c7d1RWzH, Y6Lhv1bJw2FkNmn3, YmDQRkyM1AbLrTfj, Yq67b1KDRtWngmQ9, Yv79mPLNKz3FtwWT,
Yyrb2WHMXnjg4qhk, ZCjrmW6pqQHz4KTB, ZHhbDAfgNPGw7jKn, ZKVYvq3yDgrLC1Bb, Zk39FczKDRAma1P8, a4Z37MJjgwtFCd2p,
aQJbGwRPqnHF2k3C, b1kBgPqM73QcRF6C, b8yfFZkDXMAgj1QT, bDMNL9yHZqmkBzRg, bDyYhZqdrj6GvCTP, bNfrjh3dkPDBc6QC,
bR36KWrN1AqhyvpV, bVLmacnZvd4Wr3jT, bWH41fw8CDQAch3K, bYQ4hR6ANxtLHJVF, bZyv764FwND9rxnf, bkKYqzFPa82pVZgn,
bmpBgLZjywNx47PJ, bnZ2PATGBd6LJRgF, c2FZYnfzkpQ9TG4M, cHZfJLMXvgwby7Wm, cTqwJd7hWF268ZaY, cVCh2Xt8yWDFKf1x,
cawpdgAx8tfYmGRC, cd8rLAh3Vv6zTqYF, ckFbpzH4TQtJngZm, cqwaR6bk2MPfNAWK, dAjVx8cpfZaWD6F2, dFL2bVYKqmgxHXaQ,
dJWbNhCMmz7nZcrD, dKnYzPTqv8JCM21t, dLWzb32mAyh1TV6H, dXRxaK8ZT2Mm9rDC, dkNYG8gAxrHKZV6h, dyqMvNQT2X1nFgR6,
f3hvJkNWbpHPngQz, fTBkLPW4MY2zN8AJ, fWJCNTz48aYgtpMn, fpgRbDCBT7vW1r4w, ftZxDMRNjkHWdG9J, ftmT9VHRhPApYzXd,
fzDhrGXtF1TapcCj, g4xCKbBzFNA8GDrT, g8aT4fbm3tVpBCwA, gCH7z2k6DnpjF4df, gDFPnNdGWavmjtb2, gJRNmr9XkFBL1dK4,
gMXbHq74Dtv9njcJ, gRJBtXAQ6vaDP3kT, gZM2vjAfpdt3YwKk, gZfAHd47XvKpW89h, gbPJcNvHG6ydWazX, gfRMznjvGqKPXwbA,
gpqyFGB3MXWcdjQK, gxFC8pWLVyKQnaXh, gxy3B74pZD9GYHbX, h7cbfBAzpTVQG9y1, h9txYKraHwfn26MB, hG97mXtdPQrzTjKD,
hMZWYNwyLcDkXJVj, hfc6HbNWJ4FxV9zd, hkHDY7bf1ctRBp9L, hknaKZb92zLWM673, hrxHCqb68LZ7mMaY, hwgAa8M2P7RWdrcj,
hxW4RgXkVK9NQYPA, j4gkwbLDmXyMNBPT, jBa6QVbPcwDR2v1F, jMm29YB1HRP3VQT4, jRV4tPAG2Bp71Cd9, jXNZ9MbTdWFmHG3D,
jq8WhaLyYD74gcXB, k1NhZqby2YfHP8zg, k2NcVdrLyzmwP7hB, k7KraJTRmyGDBgWX, kALYFBZ8btdqx432, kAfdFYXcHyjmgvta,
kL1g29NH7MhXvBRQ, kLMPAKXnJvC7Q1zj, kT1PNbfgAKvY84rz, kcrGd2whRZx6YbTF, m13HVkxp46WCarTd, m2zTNDx8Bhc7pyr3,
m3GbaCBXw1dj6rQL, m4fkcgFvj9zaAWBp, m7nh21pN9VTPtKdM, mD46wk93VgWpPxrB, mNDwAgTbxnX8MGLF, mNQpyJGVzdhrRgCX,
mQxY8pRvq6GZykDj, mdwkArzCj27X41DZ, mkP4ZBTpfjAhHqVR, mpx92zHCYLnTKfAv, n4vQdtMkRTc1zxXV, nLp4gAfrQPXRx68k,
nLxh7GVZpbf8APQt, nM8ZvpbrDqNzFACd, nNaMPcDkxp8tFBzG, nTz1vgfdJYW8XpKF, nYtvxcMjAaDLpzf9, nbPCLHj1yRg6mcdz,
nf7Vka8WJQdLP9yq, nfzKjkVt6vwLcM2A, nqNJVadK4vRh6pLW, pFv7kyr3RLNAQHCX, pPTzQtF2dXRqraAh, pQRy8fBzdNXnMKW6,
pTFt2yWhkJrDGRmK, pYzA34RF9GZNqHmd, pc8r6QFxBzdCaKtZ, pqyLrZ6Xa1wdQG2V, pwVcnYPzD7XjAm4t, pwkhWHQqRm8CGX6T,
py2QfBhmnZvwA38j, q28MR6X4dpf93QLb, qZHchLpJYMRQrG1A, qnAWK2hGrpBfwRjc, qp6P4GLHwa8mdVjM, qtjGQKhR8axwY1pn,
r2ZR8QXqABm9aTLg, r7wKNFQyWLvka3Z1, rQ8FL4MKkbm6a39X, rXNYR7VkQd4xW1FC, rXVcgLw9B27JACHv, rgjPnt9fdb2LV43C,
rpLqJyR8932nNM6g, tAjZJdWRzknch9y2, tHXpFG43A7Mbfd1C, tRQVM1c2DCPd8qZY, tdw9jmPgrpzXWZB6, thGK6BjaVJvfnWNP,
thTwHnVYF4J2WN7P, tv7qQjA1yY4WX82G, txRbXZwMJm4BW7q6, v1qh7nCdNw8QH4Gp, v89dcCW264jZfAXV, vAF3DMCYNjLm9wrn,
vLa1NzdYWTJbjc7V, vNJaB1t9d2hTPxAC, vNPnZjryCdJqGQpt, vNYrW1QH942ChfpZ, vXprm74b1dYtcJAC, vmWbK6QA4VgnhcFy,
vyg31APTh4RfpVBN, wYAFrKfHRgdTCWyB, wa3CrXD2QdZ4NFcY, waN34KDtpkBgcX1h, wkDVn8bA9QgrFZCt, wknr9zFjQ68DRcv1,
wrPjGzWQnHJX7A8K, wtGDFX91YWm8Q6KA, x1WVhMmGHqrXgJK4, xJKy8gwqzmvFd3h2, y4rjQqFwHBzbdRZk, y72Cf4pTxQ1WFbaV,
yAzcZkPHYrpQ9fGh, yLFbzXntM8dvQ3YC, yQk7NDW4whnA3Rq9, yYrPQWLDJk7AvT46, ydPDgaLYZGzw7R3V, yhjnabTfN9RMz8Pk,
ymJVMGnA8fcTvXBC, ytkHZ2MNrvpx1YTK, zK2GbRZf6jYcLvmy, zXWcxDG6gMLq4vCj, zY9gApRPwVfj1MT3, zYPAwt4gc1GRDWVk,
zd39tLrF4RDKZ21A, zd4rcDQ7f2j1bJ3K, znBtTYGHyCgfWcQk, ztnPqVJdCk9YwZBN, zvDJFrNWYBXMT18Z, zw1GNqDB83a2FtpT,
zyF7WfbwGVCg68cH
