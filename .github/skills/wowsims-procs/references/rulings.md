# Rulings

Each bullet is one rule the code follows: its id, the rule, where it lives, the pull request that settled it, and the
evidence that measured it (`docs/evidence/`). An id is never reused; a replaced rule keeps its bullet with a
`Replaced by` note.

- **R-0040** **A weapon proc whose row has melee or ranged DefenseType (direct hits and DoT-only applications) is a
  special hit: triggered by a main-hand or ranged hit it chains to weapon procs (its own weapon included), combat
  enchants and oils; triggered by an off-hand hit it is heard by none (main hand chains, off hand nothing). Magic weapon
  procs stay `ProcMaskEmpty`, equip and aura procs keep the `SpellFlagProc` gate, and Windfury never hears them.**
  `sim/common/shared/shared_utils.go`, `weaponProcStrikeMask`, PR #352.
- **R-0041** **No weapon proc hears an ability whose row requires a shield (SpellEquippedItems), which strikes with
  the shield, or has magic DefenseType: a weapon's chance-on-hit item effect (ItemEffect TriggerType 2) and a combat
  enchant both skip its hits. Auto attacks are exempt, since a wand's shots carry magic DefenseType and cast weapon
  procs. Equip procs (TriggerType 1) still hear these abilities. Shield Slam and Shield Bash require the shield, and
  Thunder Clap is Physical with magic DefenseType. Non-damaging Sunder Armor, Concussion Blow and Rend applications
  proc Fiery Blaze and Dragon's Call; Rend ticks proc neither.** `sim/core/aura_helpers.go`, `canProcFrom`; PR #958; EV-0006.
- **R-0042** **No weapon proc hears a melee or ranged special whose row states no DefenseType; auto attacks stay
  exempt. Sweeping Strikes' copied hit 12723 has no SpellCategories row and procs no chance-on-hit weapon, while its
  normalized hit 26654, with melee DefenseType, rolls them. Equip procs such as Hand of Justice and Dreadfrost Saber's
  still hear 12723.** `sim/core/aura_helpers.go`, `canProcFrom`; `sim/warrior/talents_arms.go`,
  `registerSweepingStrikes`; PR #965; EV-0007.
- **R-0043** **A chance-on-hit weapon rolls once for each target an area melee strike lands on, since the hit listener
  hears each target's result. Weaponmaster's sword roll is the exception and rolls once a cast.**
  `sim/core/aura_helpers.go`, `matchesSpell`; `sim/warrior/talents_arms.go`, `core.OncePerCast`; PR #986; EV-0001.
- **R-0044** **The hit table decides a strike's mask, not its school (Bow of Searing Arrows is a ranged special); a
  strike's own re-cast during its cast goes one batch later.** `sim/common/shared/shared_utils.go`,
  `weaponProcStrikeMask` and `procDamageHandler`, PR #352.
- **R-0045** **An oil's combat spell rolls on its hand's landed white swings and specials: Frost Oil's and Shadow
  Oil's enchant rows (26 and 25) state a chance and no hit kind.** `sim/common/shared/consumable_imbues.go`,
  `registerTemporaryEnchantCombatSpell`, PR #352, PR #956.
- **R-0046** **A rate the game files do not state takes the carry-over value (`vanilla-carryover`) with a TODO, or is
  written and commented out waiting for a procs-per-minute rate, and the gear tab shows an in-game validation notice
  for every unvalidated rate, driven by the override tables' `Source`; a measured rate's `Source` is its evidence id.**
  `tools/database/gen_effects.go`, `rateOrigin`, PR #322.
- **R-0047** **A hand-written proc that runs at the procs per minute on its item's entry registers nothing while the
  entry states none, and no rate is made up for it: the item keeps its stats, and the generator lists it in the gear
  tab's missing-effect notices until a rate reaches the entry.** `sim/common/forever/items_weapons.go`,
  `newEntryRateItemEffect`; `tools/database/gen_effects.go`, `handProcStatesNoRate`; PR #905.
- **R-0048** **A charge buff hit by the wearer's channel is reserved and spent at that channel's end
  (`Unit.OnChannelEnd`), an interrupted channel included, and a charge gained while a channel its class flags name is
  running is spent at that channel's end.** `sim/core/spelldata/charges.go`, `AttachChargeSpender`, PR #658.
