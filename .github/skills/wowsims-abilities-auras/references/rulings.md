# Rulings

Each line is the rule the code follows, where it lives, and the pull request that settled it.

## Rolls and outcomes

- **`ATTR_EX_3_ALWAYS_HIT` on a triggered row stops only that row's own roll; the sim skips the miss only when the row
  the player casts carries the bit, so Judgement of Command keeps the miss roll of its dummy 20968.**
  `sim/core/spelldata/attributes.go`, `AlwaysHits`.
- **A spell whose row has DefenseType magic and no Always Hit rolls `OutcomeMagicHit` even when it deals no damage
  (Taunt, Challenging Shout, Intimidating Shout, Growl, Challenging Roar); taunts carry `ProcMaskEmpty`.**
  `sim/warrior/taunt.go`, PR #649.
- **Intercept's stun is its own spell from the stun row, rolling `OutcomeMagicHitAndCrit` with an empty proc mask;
  targets at character level + 3 are immune to its stun and damage, and nothing is dealt at fight-end teardown.**
  `sim/warrior/intercept.go`, PR #652.
- **Consecration ticks every second from a tick one second after the cast, each tick rolling its own hit, and Serpent
  Sting is a full ranged ability that reaches Judgement of Wisdom and ranged procs.** `sim/paladin/consecration.go`,
  `sim/hunter/serpent_sting.go`, PR #452.
- **DoTs do not snapshot: every tick deals damage on the stats and modifiers of the moment it lands
  (`CalcAndDealPeriodicDamage` in `OnTick`, no `OnSnapshot`), Deep Wounds' rollover pool being the one amount fixed at
  the crit.** `sim/core/spelldata/resolve_aura.go`, `DotConfig`.
- **Deep Wounds bleeds the talent share of average weapon damage without attack power; a new crit rolls the owed damage
  over and extends the timer without resetting it, and the ticks never crit.** `sim/warrior/talents_arms.go`, PR #231.

## Reading the row

- **A talent's trait curve applies with the sign the game files state, even where base points and tooltip say the
  opposite; no sign override exists.** `sim/core/spelldata/ladder.go`, PR #399.
- **Seal of Righteousness keeps the tooltip description-variable coefficients 0.12 and 0.0765 over the rows' 0.**
  `sim/paladin/seal_of_righteousness.go`, PR #469.
- **Shield Bash and Shield Slam register as `ProcMaskMeleeOHSpecial`, which keeps Windfury off them; the game files give
  no hand, so this is a ruling, not data.** `sim/warrior/shield_bash.go`.

## Registering

- **Casters register every rank only when the implementation requires it, a question asked during integration (melee
  registers `Highest()` only); each per-rank DoT or HoT is labelled with `spelldata.RankLabel(rank, &owner.Unit)` so
  ranks never share one aura; category cooldowns already share one timer across ranks, so no hand `NewTimer()`.**
  `sim/core/spelldata/resolve_aura.go`, PR #246.
- **Devouring Plague is learnable by every race, and Inner Focus restarts its cooldown when the buff is spent
  (`OnExpire` calls `CD.Use`) though the row has no cooldown-on-event bit.** `sim/priest/devouring_plague.go`,
  `sim/priest/talents_discipline.go`, PR #246.
- **A school hit talent (SPELLMOD_RESIST_MISS_CHANCE on class flags) shows on the sheet by opting in with
  `spelldata.ShowAsSchoolHit(schools)` and its tooltip's schools, never `SchoolBonusHitChance`; all-spell hit talents
  (aura 55 misc 0) already show and must not opt in.** `sim/core/spelldata/parse_effects.go`, PR #285.
- **Weaponmaster covers a mace or staff in either hand without doubling, Anger Management comes from its aura, crit
  rage ×2 follows the row, a wand keeps `SpellFlagNoPartialResist`, and Flamestrike's burn stays a dot.**
  `sim/warrior/talents_arms.go`, `sim/mage/flamestrike.go`, PR #494.

## Forms

- **A shift into Bear or Dire Bear Form empties the rage bar in the shift spell's `ApplyEffects`, Furor's 10 rage comes
  after, and the starting form at the pull keeps `StartingRage`.** `sim/druid/forms.go`, PR #149.
- **A druid's change of form keeps the current health percentage both ways and a Dire Bear guardian starts the pull at
  100%; any new form or max-health form effect sits inside `keepHealthPercentAcrossForms`.** `sim/druid/forms.go`,
  PR #149.
