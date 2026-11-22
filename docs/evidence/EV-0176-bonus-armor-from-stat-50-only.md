# EV-0176: Bonus armor comes from stat 50, and an armor item's QualityModifier gives none

- Claim: an item's bonus armor in game is its stat 50 (`ITEM_MOD_EXTRA_ARMOR`) amount; an armor item whose only bonus
  armor is its `QualityModifier` gives none, and one with both gives the stat 50 amount once.
- Sources: 37 StatLedger files from 8 characters
- Game build: not recorded
- Date: not recorded
- Pull request: #PR

Each of the 16 items with bonus armor that the ledgers show worn was compared with the armor on the character sheet,
through gear swaps and through lines where nothing else changed. For the two items whose only bonus armor is a
`QualityModifier`, the sheet reads the whole amount below a model that grants it and never matches that model:
Arcanoweave Cloak on 40 lines at −70, Turtle Scale Bracers on 826 lines at −100. The sheet gives the other 14 their stat
50 amount, including the two that also carry a `QualityModifier`.

| Item                            | Slot     | QualityModifier | Stat 50 | In game |
| ------------------------------- | -------- | --------------- | ------- | ------- |
| Arcanoweave Cloak (272411)      | Back     | 70              |         | 0       |
| Turtle Scale Bracers (8198)     | Wrist    | 100             |         | 0       |
| Stonefist Girdle (6742)         | Waist    | 100             | 100     | 100     |
| Shield of Thorsen (13079)       | Shield   | 50              | 50      | 50      |
| Tusken Helm (6686)              | Head     |                 | 120     | 120     |
| Dual Reinforced Leggings (9625) | Legs     |                 | 120     | 120     |
| Archaedic Stone (11118)         | Trinket  |                 | 50      | 50      |
| Heavehammer (271766)            | Two-hand |                 | 60      | 60      |
| Refractory Scaleguards (273048) | Legs     |                 | 100     | 100     |
| Quilrager Throwing Axe (274160) | Thrown   |                 | 40      | 40      |
| Crocskin Waders (274962)        | Legs     |                 | 47      | 47      |
| Enchanted Sandals (279843)      | Feet     |                 | 50      | 50      |
| Golem War Cloak (281319)        | Back     |                 | 72      | 72      |
| Nature's Cull (281588)          | Staff    |                 | 115     | 115     |
| Mantle of Woe (7750)            | Shoulder |                 | 33      | 33      |
| Crystal Thread Cloak (273654)   | Back     |                 | 40      | 40      |

No item whose two amounts differ (Force Imbued Gauntlets, 18383: 140 and 80) was worn.
