# Rulings

Each line is the rule the code follows, where it lives, and the pull request that settled it.

- **A weapon proc whose row has melee or ranged DefenseType (direct hits and DoT-only applications) is a special hit:
  triggered by a main-hand or ranged hit it chains to weapon procs (its own weapon included), combat enchants and
  oils; triggered by an off-hand hit it is heard by none (main hand chains, off hand nothing). Magic weapon procs stay
  `ProcMaskEmpty`, equip and aura procs keep the `SpellFlagProc` gate, and Windfury never hears them.**
  `sim/common/shared/shared_utils.go`, `weaponProcStrikeMask`, PR #352.
- **The hit table decides a strike's mask, not its school (Bow of Searing Arrows is a ranged special); Shadow Oil
  hears its hand's melee hits and Frost Oil white hits only; a strike's own re-cast during its cast goes one batch
  later.** `sim/common/shared/shared_utils.go`, `procDamageHandler`, PR #352.
- **A rate the game files do not state takes the carry-over value (`vanilla-carryover`) with a TODO, or is written and
  commented out waiting for a procs-per-minute rate, and the gear tab shows an in-game validation notice for every
  unvalidated rate, driven by the override tables' `Source`.** `tools/database/gen_effects.go`, `rateOrigin`, PR #322.
- **A charge buff hit by the wearer's channel is reserved and spent at that channel's end (`Unit.OnChannelEnd`), an
  interrupted channel included, and a charge gained while a channel its class flags name is running is spent at that
  channel's end.** `sim/core/spelldata/charges.go`, `AttachChargeSpender`, PR #658.
