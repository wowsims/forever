# Item and enchant data traps

What the game files' item, enchant, suffix and set rows say, and where they mislead.

- **An enchant effect id can carry several rows of different type** (1897 is both 2H Lesser Impact and Superior
  Striking). `core.GetEnchantByEffectID` answers the first, so a fit or slot rule uses `core.FittingEnchantRow` or
  `core.GetEnchantRowsByEffectID` (`sim/core/database.go`), and the request database ships every row.
- **An enchant's name can point at a dead slot.** Read the description of the `E_ENCHANT_ITEM` spell that grants it
  instead, treat the grant effect's base points as a display magnitude, never a proc chance, and count defense as 1
  rating per skill point. The table is `SpellItemEnchantment` in `tools/database/wowsims.db`.
- **An Item row with no ItemSparse row is an item the game build does not ship** (Ironfoe, Felstriker), so it needs
  no sim work. Resolve names through `ItemSparse.Display_lang` and ignore `Monster -` rows.
- **Random enchantments are bonus lists**: `ItemXBonusTree` to `ItemBonusTreeNode` context 0 to `ItemBonus` types 2,
  5 and 23 (`tools/database/tables.go`). They are keyed by bonus list id, since one name can carry different
  allocations, and each stat is rounded to nearest from allocation × points / 10000 (`sim/core/database.go`,
  `ui/sim/proto/equipped_item.ts`).
- **An `S03 - …` ItemSetSpell row still ships when it sits in Forever's own threshold layout of the set.** Compare
  the layout with the set's Season of Discovery copy in the game files, not the name.
