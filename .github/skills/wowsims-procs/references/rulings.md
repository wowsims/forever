# Rulings

Each line is the rule the code follows, where it lives, and the pull request that settled it.

- **A weapon proc whose row has melee or ranged DefenseType (direct hits and DoT-only applications) is a special hit:
  triggered by a main-hand or ranged hit it chains to weapon procs (its own weapon included), combat enchants and
  oils; triggered by an off-hand hit it is heard by none (main hand chains, off hand nothing). Magic weapon procs stay
  `ProcMaskEmpty`, equip and aura procs keep the `SpellFlagProc` gate, and Windfury never hears them.**
  `sim/common/shared/shared_utils.go`, `weaponProcStrikeMask`, PR #352.
- **No weapon proc hears an ability whose row requires a shield (SpellEquippedItems), which strikes with the shield,
  or has magic DefenseType: a weapon's chance-on-hit item effect (ItemEffect TriggerType 2) and a combat enchant both
  skip its hits. Auto attacks are exempt, since a wand's shots carry magic DefenseType and cast weapon procs. Equip
  procs (TriggerType 1) still hear these abilities. Shield Slam and Shield Bash require the shield, and Thunder Clap
  is Physical with magic DefenseType. Logs of Protection warriors: Fiery Blaze (enchant 36) 0 of 617 Shield Slams, 0 of 78 Shield Bashes
  and 0 of 436 Thunder Claps, against 14.0% of white main-hand hits; Rock on Stick 0 of 157 Shield Slams; Dragon's
  Call 1 of 50 Shield Slams (sharing its millisecond with a white hit) and 0 of 4 Shield Bashes; Dreadfrost Saber's
  equip proc 54 of 80 and 60 of 105 Shield Slams and 7 of 8 Shield Bashes. Non-damaging Sunder Armor, Concussion Blow
  and Rend applications proc Fiery Blaze and Dragon's Call; Rend ticks proc neither.** `sim/core/aura_helpers.go`,
  `canProcFrom`; PR #958.
- **No weapon proc hears a melee or ranged special whose row states no DefenseType; auto attacks stay exempt. Sweeping
  Strikes' copied hit 12723 has no SpellCategories row: Arms warrior logs show 0 chance-on-hit procs (Nightblade,
  Runic Darkblade, Sword of Dismemberment, Rock on Stick, Bigger Rock) on 567 copies where the white hit rate predicts
  about 61, while its normalized hit 26654, with melee DefenseType, rolls them at about the white hit rate (0.83 of
  it, 95% 0.37 to 1.44). Equip procs such as Hand of Justice and Dreadfrost Saber's are unmeasured on 12723 and still
  hear it.** `sim/core/aura_helpers.go`,
  `canProcFrom`; `sim/warrior/talents_arms.go`, `registerSweepingStrikes`; PR #965.
- **The hit table decides a strike's mask, not its school (Bow of Searing Arrows is a ranged special); a strike's own
  re-cast during its cast goes one batch later.** `sim/common/shared/shared_utils.go`, `weaponProcStrikeMask` and
  `procDamageHandler`, PR #352.
- **An oil's combat spell rolls on its hand's landed white swings and specials: Frost Oil's and Shadow Oil's enchant
  rows (26 and 25) state a chance and no hit kind.** `sim/common/shared/consumable_imbues.go`,
  `registerTemporaryEnchantCombatSpell`, PR #352, PR #956.
- **A rate the game files do not state takes the carry-over value (`vanilla-carryover`) with a TODO, or is written and
  commented out waiting for a procs-per-minute rate, and the gear tab shows an in-game validation notice for every
  unvalidated rate, driven by the override tables' `Source`.** `tools/database/gen_effects.go`, `rateOrigin`, PR #322.
- **A hand-written proc that runs at the procs per minute on its item's entry registers nothing while the entry states
  none, and no rate is made up for it: the item keeps its stats, and the generator lists it in the gear tab's
  missing-effect notices until a rate reaches the entry.** `sim/common/forever/items_weapons.go`,
  `newEntryRateItemEffect`; `tools/database/gen_effects.go`, `handProcStatesNoRate`; PR #905.
- **A charge buff hit by the wearer's channel is reserved and spent at that channel's end (`Unit.OnChannelEnd`), an
  interrupted channel included, and a charge gained while a channel its class flags name is running is spent at that
  channel's end.** `sim/core/spelldata/charges.go`, `AttachChargeSpender`, PR #658.
