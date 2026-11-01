# Rulings

Each line is the rule the code follows, where it lives, and the pull request that settled it.

## Rolls and outcomes

- **`ATTR_EX_3_ALWAYS_HIT` on a triggered row stops only that row's own roll; the sim skips the miss only when the row
  the player casts carries the bit, so Judgement of Command keeps the miss roll of its dummy 20968.**
  `sim/core/spelldata/attributes.go`, `AlwaysHits`.
- **Whether a special can be dodged, parried or blocked follows the row's No Active Defense (`Attributes[0] 0x200000`),
  which `rowFlags` turns into `SpellFlagNoActiveDefense`; the outcome function never encodes it, so Backstab and Feint,
  whose rows lack the bit, can be dodged.** `sim/core/unit.go`, `GetTotalDodgeChanceAsDefender`, PR #712.
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
  registers `Highest()` only, apart from the class book ranks below); each per-rank DoT or HoT is labelled with
  `spelldata.RankLabel(rank, &owner.Unit)` so ranks never share one aura; category cooldowns already share one timer
  across ranks, so no hand `NewTimer()`.** `sim/core/spelldata/resolve_aura.go`, PR #246.
- **A rank that only a class book teaches registers beside the trainer's top rank, and the rotation picks the rank by
  spell id. Each simple rotation rank toggle defaults off, which casts the trainer's top rank, and the ids swap through
  `withSpellIds`, which never swaps a tag -1 id (the raid's copy of a buff); in the rotation editor the spell dropdown
  lists both ranks. The warrior and warlock presets cast the book rank: Heroic Strike 11567 and 25286 (each with its own
  queue), Battle Shout 11551 and 25289, Revenge 11601 and 25288 (one window, read by 5302), Shadow Bolt, Immolate and
  Corruption. The hunter and shaman presets cast the trainer's top rank, because the books may not be there at launch:
  Serpent Sting 13555 and 25295 (one sting to a hunter), Aspect of the Hawk 14322 and 25296, Strength of Earth Totem
  10442 and 25361, Grace of Air Totem 10627 and 25359 (one totem to a slot), and shaman has no simple rotation. The
  mage, rogue and druid presets do the same: Frostbolt 10181 and 25304, Fireball 10151 and 25306, Arcane Missiles 10212
  and 25345, Frost Ward 10177 and 28609, Backstab 11281 and 25300, Feint 11303 and 25302, Eviscerate 11300 and 31016,
  Ferocious Bite 22829 and 31018. A plain SIML name is the trainer's rank and a book rank takes a suffix, which
  `sim/core/spelldata/book_ranks.go` decides, since a book rank's game files row states it like a trained one.
  Multi-Shot has the one rank, 2643, so its Guide teaches nothing here. The raid's `battle_shout` has four states (off,
  Rank 6, Rank 7, Rank 7 with the tier 2 set; no Rank 6 with the set), and the warrior's own copies are labelled by rank.**
  `ui/sim/proto/apl_utils.ts`, `sim/warrior/heroic_strike_cleave.go`, `sim/warrior/battle_shout.go`,
  `sim/warrior/revenge.go`, `sim/core/buffs/battle_shout.go`, `sim/core/buffs/drivers.go`, `sim/hunter/serpent_sting.go`,
  `sim/hunter/aspects.go`, `sim/shaman/totems.go`, PR #736, PR #746, PR #794.
- **A raid or individual buff that a class book raises is a rank select: an int32 where 0 is off and 1 the trainer's top
  rank, listed in the manifest row's `Ranks`, from which the generators render the proto field, the apply block and the
  settings input. The totems and greater blessings take a select. A class's own copy of those still reads the row's
  `SpellID`. The four group inputs, `mark_of_the_wild`, `power_word_fortitude`, `arcane_intellect` and `shadow_protection`,
  are bools that read the single-target spell's top rank, since its amounts match the group spell's top rank (Mark of the
  Wild +16 stats and 385 armor; Power Word: Fortitude 70; Arcane Intellect 31; Shadow Protection 60). A mage in the raid
  sets `arcane_intellect`. A priest's Prayer of Fortitude and Prayer of Shadow Protection read their own rows and bid
  against the raid's input of the same stat, so the two count once. Prayer of Spirit stays a bool, since no item teaches it.**
  `tools/database/buffmanifest/buffs.go`, `sim/core/buffs/meta.go`, `sim/core/buffs/priest.go`,
  `ui/features/inputs/icon_inputs.ts`, PR #794.
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
