# Rulings

Each line is the rule the code follows, where it lives, and the pull request that settled it.

- **Elixirs stack their stats across groups, but only the strongest elixir crit applies.**
  `sim/common/shared/consumable_stats.go`, `consumableAura` with `spelldata.Exclusive(elixirCritCategory, false)` (the category `"ElixirCrit"`), PR #209.
- **Holiday consumables (`Consumable.holiday`) sit in their own row above Miscellaneous, are not Well Fed and stack
  with food; a Forever food's stats are the amount Nutritious Food's `A_PERIODIC_TRIGGER_SPELL_WITH_VALUE` hands Well
  Fed.** `tools/database/gen_consumables.go`, `consumableHoliday`; `sim/common/shared/consumable_stats.go`,
  `ConsumableSpells`, PR #207.
- **A hand-written item effect only one class can use lives in `sim/<class>/item_<kind>.go` (sets, librams, idols,
  totems, trinkets, weapons); class-agnostic items live in `sim/common/forever/items_*.go`. A set the game files open
  to every class is shared, in `sim/common/forever/items_sets.go` with `agent.GetCharacter()`, and so are the dungeon
  sets; Naxxramas tier 3 armor is class-locked through `ItemSetClasses` and coded in the class folders.**
  `tools/database/overrides.go`, `ApplyItemSetClasses`, PR #316 and PR #539.
- **A survival cooldown that adds no damage or mana has `DisableAutocast` set unless the character sits in a raid
  Tanks slot, so Autocast Other Cooldowns uses it only at a cooldown timing; defensive on-use buffs count as Survival
  and MP5 or Spirit ones as Mana; explicit APL casts are unaffected.**
  `sim/core/major_cooldown.go`, PR #335.
- **ItemSparse `REDUCED_CASTER_DPS` gives weapon DPS = damage table × 2/3 one-hand or × 29/39 two-hand, with
  QualityModifier an additive offset; the shares are fitted to in-game ranges, not read from the game files.**
  `tools/database/dbc/item_damage.go`, `reducedCasterDpsShare`, PR #298.
- **`CASTER_SPELL_POWER` gives spell power ⌊20000 × budget / 10000⌋ and `CASTER_HEALING_POWER` gives healing
  ⌊37600 × budget / 10000⌋ with spell damage ⌊healing / 3⌋, budget being RandPropPoints column 0.**
  `tools/database/dbc/item.go`, `addCasterSpellPower`, PR #302.
