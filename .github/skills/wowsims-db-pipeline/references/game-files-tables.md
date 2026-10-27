# Reading wowsims.db

This page covers reading a table the settings do not list, what the tables in `tools/database/wowsims.db` hold and how
a script reads them.

- **Another table, once.** Copy `tools/database/generator-settings.json` with only that table listed and run
  `db2tool -s <copy> --output <other.db>`. To add one table to the working database, `ATTACH` the other and
  `CREATE TABLE … AS SELECT`.
- **Empty tables are real.** The game files ship no gem system, zero `SpellScaling` columns, no `SpellLearnSpell`, no
  procs per minute (`SpellProcsPerMinuteID` 0), no Jewelcrafting or Inscription, and `ItemSparse.AllowableClass` is -1
  on class-only items such as Atiesh.
- **Array columns hide generated columns.** `db2tool` adds a `<Column>_<n>` column per element
  (`EffectSpellClassMask_0..3`, `ImplicitTarget_0`, `Attributes_N`; `tools/db2tool/sqlite/schema.go`) that
  `pragma table_info` hides and `select *` returns. Take column names from `cursor.description` or select an explicit
  list.
- **`SkillLineAbility.RaceMasks` is `[low, high]`** words indexed by `ChrRaces.PlayableRaceBit`, so Skyborne's bits
  32 and 33 land in the high word as 1 and 2.
