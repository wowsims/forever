---
name: wowsims-procs
description: 'Use when WoWSims Forever Go code builds or fixes a proc: calling spelldata.ProcTrigger or its ProcOpts (PPM, Chance, WeaponProc), reading ProcChanceSource or a ProcChance of 100 or 101, reading what DecodeProcTypeMask made of a row, porting a proc-chance talent, taking an ICD from the row, spending a charged buff, or wiring an item, enchant, set, extra-attack or weapon-totem proc. Use it too when a trigger never fires, a bleed tick procs nothing, a hand trigger fires on every hit or on a dodge, the build panics with "states no proc chance", Windfury or an oil hears the wrong hits, a charge outlives the cast that should spend it, or a measured rate disagrees with the tooltip. Reading a row is wowsims-spell-store; the buff a proc applies is wowsims-abilities-auras; when a handler runs is wowsims-core-mechanics; item effect routing is wowsims-consumables-items; the proc audit is wowsims-db-pipeline; tests that prove a rate are wowsims-testing; the map of spell skills is wowsims-spells.'
---

# wowsims-procs

A proc is a listener the row describes: which hits it hears, how often it fires and how long it rests.
`spelldata.ProcTrigger` builds all of that from the row, and the caller adds the name, the handler and only what the
game files leave out. The rule this skill exists for: the `ProcChance` column is not a chance until
`ProcChanceSource` says so, and a listener typed by hand around it fires on the wrong hits, or on every hit. The
rulings the code follows are in `references/rulings.md`. Text here follows `wowsims-conventions`; the page below holds
the detail.

## Where things live

| Path                                        | What it is                                                                     |
| ------------------------------------------- | ------------------------------------------------------------------------------ |
| `sim/core/spelldata/resolve_proc.go`        | `ProcTrigger`, its `ProcOpt`s, `StatedChance`                                  |
| `sim/core/spelldata/charges.go`             | `AttachChargeSpender`, the listener that spends a charged buff                 |
| `sim/core/spelldata/item_proc.go`           | `ItemProcUnsupportedWithRate`: whether an item's rows say enough to build      |
| `sim/core/proc_types.go`                    | `DecodeProcTypeMask` and the `ProcHint` bits                                   |
| `sim/core/aura_helpers.go`                  | `core.ProcTrigger`, `MakeProcTriggerAura`, the roll, the ICD, `OncePerCast`    |
| `sim/core/procs.go`                         | `DynamicProcManager` and the procs-per-minute managers                         |
| `sim/core/buffs/drivers.go`                 | the Windfury Totem and Flametongue Totem procs                                 |
| `sim/common/shared/`                        | the item, enchant and set proc builders (`NewSpellData*Proc`)                  |
| `tools/database/overrides/weapon_procs.go`  | a chance-on-hit weapon's procs per minute, which the game files do not carry   |
| `tools/database/overrides/enchant_procs.go` | a permanent enchant's rate where its spell states none or the 100/101 sentinel |

| You are…                                                                               | Open                                                                                    |
| -------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| building a trigger from a row: the four chance shapes, what it fills, the rate options | `docs/spell_data/procs.md`, the opening tables                                          |
| porting a proc-chance or proc-cooldown talent, a pet's or a target-side trigger        | `docs/spell_data/procs.md`, the paragraphs on `ProcSpellClassFlags`                     |
| reading what the decoder made of a row, or a tooltip hint                              | `docs/spell_data/procs.md` §The decoder and the tooltip hints                           |
| registering an item, enchant or set proc                                               | `docs/spell_data/procs.md` §Item and enchant procs                                      |
| deciding a weapon proc's mask, an unstated rate, or a charge spent by a channel        | `references/rulings.md`                                                                 |
| the review case behind a rule below: a typed ICD, a look-alike effect, a cast listener | `.github/skills/wowsims-review/references/defects.md` §The wrong rank, the wrong effect |
| the review case behind a charged row with no spender                                   | `.github/skills/wowsims-review/references/defects.md` §What a cast spends, and when     |

## Building a trigger from a row

What `ProcTrigger` fills and the Enrage example are in `procs.md`. What the page leaves to this skill:

- **Register it one of two ways**, name set first (see Traps): `unit.MakeProcTriggerAura(trigger)` gives it an aura
  of its own, up all fight; `buff.AttachProcTriggerCallback(&unit, trigger)` hears hits only while that buff is up.
- **The ICD is the row's** (`s.ICD()`); an ICD typed by hand where the row states one is a defect.
- **A row with charges needs a spender.** `ProcCharges` becomes the buff's `MaxStacks` through `AuraConfig`, and
  nothing takes a stack off until `spelldata.AttachChargeSpender(character, buff, row, name)` is attached. It spends
  one charge per cast however many targets the cast hits (`core.OncePerCast`), and a channel's once at its end when
  the row names its spells by class flags; a row that names none is spent by the first hit it hears, mid-channel too. An
  option's `ExtraCondition` narrows the spells that spend it (Revelation turns melee and ranged attacks away); a charge
  gained during a running channel its class flags name is still kept for that channel's end. Without a spender one proc pays for every cast
  in the buff's duration.
- **An extra attack goes through `unit.ExtraAttackProc(&trigger, tagSpellID, count)`.** It sets `TriggerImmediately`
  itself, swings once the triggering hit is dealt, never hears its own swing, and returns an `OnInit` for the proc's
  aura (`.ApplyOnInit`). The item shape is `shared.NewSpellDataExtraAttackProc`.
- **When a handler runs against the cast that procced it**, and what `TriggerImmediately` changes, is
  `wowsims-core-mechanics` `## Cast order`.

## Chance and masks

The four `ProcChanceSource` shapes, the decoder, the proc-chance talents and every `ProcOpt` (the option table) are in
`procs.md`.

- **A `ProcMask` set on the returned trigger narrows the listener but not a manager already built.**
- **A cast listener on a proc that depends on the hit is a defect**: a dodged or missed cast still procs, for the
  reason in `wowsims-core-mechanics` `## Cast order`. `ProcTrigger` on the rank decodes the hit listener the row states.
- **Read the chance off the effect the tooltip names**, never a look-alike effect whose ladder happens to fit.
- **A weapon proc never hears an ability that requires a shield or has magic or no DefenseType.** Core keeps an
  item's chance-on-hit effects and combat enchants from hearing Shield Slam, Shield Bash, Thunder Clap and Sweeping
  Strikes' copied hit 12723. Equip procs still hear them, and auto attacks are exempt, so a wand's shots cast weapon
  procs. The rulings are in `references/rulings.md`.
- **A Windfury Totem hit is a main-hand swing to a row's melee swing bit** (`HearsCountedSwings`, EV-0136), though
  its mask is a main-hand special's. A `core.ProcTrigger` typed by hand on white hits does not hear it until it sets
  that field, and a handler that tells swings apart by mask checks `core.SpellFlagCountsAsSwing` too.
  `docs/spell_data/procs.md` §The decoder and the tooltip hints has the rule and its exceptions.
- **A spell's own proc mask decides which listeners hear it**, and the hit table decides the mask, not the school.
  The cases (a weapon proc's strike off either hand, oils) are in `references/rulings.md`.
- **Suppress Caster Procs and Suppress Target Procs keep a spell's hits from `ProcTrigger` listeners on the side each
  bit names** (EV-0165), and Suppress Caster Procs keeps its completed cast from the caster's cast triggers too.
  `rowFlags` passes all three suppress bits on only when an effect of the row does more than a dummy, a script or a cast
  of another spell, so a dummy the sim folds with its damage (Execute 20662) carries none of them. A dummy the sim
  registers as a spell of its own sets the bit from its row at the call site, as Judgement of Command's landing does
  (`sim/paladin/seal_of_command.go`); an aura with hand-written hit callbacks checks the flag itself.
- **A death listener (`CallbackOnKill`, `CallbackOnTargetDies`) gets a nil spell**, and a rated trigger hears no
  death; both are in `procs.md` (the two death bits), the order of a death in `wowsims-core-mechanics` `## Targets`.

## Item and enchant procs

Registration, every `NewSpellData*Proc` shape, the enchant slot rules, `ItemProcUnsupportedWithRate` and the two local
snapshot lists are in `procs.md` §Item and enchant procs; which builder an item effect routes to is
`wowsims-consumables-items`.

- **A rate the game files lack**: `wowsims-db-pipeline` `## Overrides`.
- **A row that deals, restores or procs N times as much against some creatures** names them only in its tooltip; its
  row id, effect index and the mob types or creature families go in `overrides.VersusBonuses`
  (`tools/database/overrides/spell_overrides.go`, `docs/spell_data/overrides.md`), or the spell deals 1× and the
  stats sheet's Versus bonuses group leaves the item out. A factor no effect states is that row's `Multiplier`. A hand-written effect
  sets `SpellConfig.VersusBonus` through `shared.RegisterVersus` on the row. A proc that is N times as likely against
  them (Hand of Justice, Ironfoe, Argent Defender) takes the bonus with `VersusEffectKindProcChance` and passes
  its `Against` to `NewTargetScaledChanceManager` or `NewAttackerSwingPPMManager` as the scale. The generated
  damage, energize and on-use builders read a `VersusBonuses` row on the row they build from as a factor on the damage
  or the restore, so a "times as likely" row comes with a hand-written effect.
  `go run ./tools/spelldata <id>` prints the versus and area overrides on `versus` and `area` lines.

## Measuring a rate

- **Read the tooltip before the column.** The tooltip decides which `ProcChanceSource` shape a row is (the opening
  table of `procs.md`); 101 usually means no proc at all. A rate the game files lack
  is an override onto the row's `RPPM` first (`Field: PPM` in `spell_overrides.go`, Omen of Clarity is the case) or
  an override row on gear, and a `PPM(n)` literal with a review TODO only where no override fits. An item whose code
  reads its rate from its chance-on-hit entry (`newEntryRateItemEffect`) takes it in `WeaponProcs`, never in a `Spells`
  PPM override: the entry carries only the first, while the generator's missing check counts either as a rate.
- **A measured procs per minute is rounded to the nearest 0.05** from the fit at full precision (1.7269 is 1.75), and
  the evidence file gives both. A fit printed on a finer grid (1.725) is not a rounded rate.
- **An unstated rate follows the carry-over ruling** in `references/rulings.md`. Its `Source` sets the rate's
  `ProcRateStatus` (`proto/ui.proto`, set by `rateOrigin` and `knownRates` in `tools/database/gen_effects.go`), which the gear tab's notice shows
  (`ui/features/gear/item_notices.tsx`): `vanilla-carryover` and `tbc-carryover` name their origin, `observed` is a
  rate measured on this server from too few logs to call final (the gear tab still warns), and a measured rate's
  `Source` is its `EV-` evidence id (`docs/evidence/README.md`), which shows none. A hand-kept rate's `Source` is one of
  those four, and the generator refuses any other. A gear rate kept as a literal in the sim gets no notice.
- **The gear tab states the rate the sim rolls** from the `proc_rate` `gen_db` writes on each item and enchant in
  `db.json` and the database's `set_bonus_proc_rates` (`proto/ui.proto`); the sim never reads them. For hand-written
  gear the generator reads the rate from its rows (`shippedRateInputs` and `knownRates` in
  `tools/database/gen_effects.go`). Each proc row (an item's chance-on-hit entry or equip spell with proc flags, an
  enchant's proc slot, a set bonus row `shared.IsSetBonusProc` names) takes the gear's own `WeaponProcs` or
  `EnchantProcs` rate for its spell, else the spell's rate override, else the row's procs per minute or chance. A
  piece whose rows disagree or state none shows no rate, and so does one whose tooltip prints the chance as a fraction
  of the column (`${$h/2}%`, The Lion Horn of Stormwind).
- **Hand-written code that rolls something its rows do not state takes a row** in
  `tools/database/overrides/gear_proc_rates.go`: code that scales the chance by target (Ironfoe), halves it by area,
  adds a flat chance, prices spell hits on their cast time (Highborne Research Tablet) or splits the procs (Bolt's
  Charged Eyeball). A row with no `Kind`
  leaves the gear with no rate and may leave `Source` empty; one with a `Kind` states its `Rate`, with the status its
  `Source` gives (`tooltip` is the game files'), though gear listed unvalidated or missing its effect still shows none
  and the test refuses a `Kind` on gear whose tooltip prints a fraction of the column. Its `Reason` names what the code
  rolls. Without the row the gear tab shows a number the sim does
  not roll, and no test fails; `wowsims-review` step 4 checks for it.

- **The class `.results` never move on a proc rate**: `wowsims-testing` `## Proving a claim at runtime`.
- **A measured rate that disagrees with the tooltip** is decided by `wowsims-spells`
  `## When a sim number disagrees with the game files`, and the outcome is a comment at the proc's code: Omen of
  Clarity's is in `sim/druid/omen_of_clarity.go`, Revelation's in `sim/common/forever/enchants.go`.
- **Which hits a weapon rate counts** is EV-0155's class census: each landed hit that rolls, priced on its own weapon's
  base speed. Execute and Judgement of Command roll the weapon once (EV-0155); a count of reports from before hotfix
  push 114313 (2026-11-20 23:47 UTC) leaves Execute out with its procs, and one from before hotfix push 114399
  (2026-11-25 01:15 UTC) Judgement of Command. Druids in a form are left out of every rate. Paladins count, by
  EV-0155's and EV-0156's rolls, with each run dated against the Judgement hotfix of 2026-11-13. A weapon where their
  rate reads apart from the other classes' even so leaves them out and says why (owner ruling, 2026-11-23; EV-0184; it
  takes the place of the rulings of 2026-11-19 in EV-0153 and 2026-11-22 in EV-0179). A proc that lands on the target
  logs almost never on a killing blow, so for such a proc a killing blow is no roll; and a report that lists one stretch
  of the log under several fight ids repeats its events, which count once (EV-0153).

## Traps

- **A trigger named after its row never hears anything when the buff shares that name.** `ProcTrigger` names the
  trigger after the row, and `MakeProcTriggerAura` registers through `GetOrRegisterAura`, which hands back the
  `AuraConfig` buff of the same label and drops the listener. Set `trigger.Name = "<X> - Trigger"` (PR #140). A death
  listener loses its kill or death the same way, so Rapid Killing (`sim/hunter/talents_marksmanship.go`) and Spirit
  Tap (`sim/priest/talents_shadow.go`) name their triggers.
- **A periodic listener built from a row never hears a bleed.** `DecodeProcTypeMask` files
  `PROC_FLAG_DEAL_HARMFUL_PERIODIC` under `ProcMaskSpellDamage`, and a bleed tick carries a melee mask. Set
  `trigger.ProcMask = core.ProcMaskUnknown` at the call site, as Thousand Cuts does (`sim/rogue/talents_subtlety.go`).
- **A harmful-spell-only mask (0x10000) decodes to a cast trigger that ignores misses.** The decoder keeps that rule
  for items, so a spell that must land to proc moves to the hit listener at its call site, as Nature's Grace does
  (`sim/druid/talents_balance.go`, PR #382).
- **Two triggers on one aura with the same callback keep only the last.** `AttachProcTriggerCallback`
  (`sim/core/aura_helpers.go`) assigns the aura's `OnCastComplete`, `OnSpellHitDealt` and the rest outright, so a
  second cast listener on a set bonus aura silently drops the first. Give each trigger its own callback or its own
  aura (`MakeProcTriggerAura`).
- **A `core.ProcTrigger` typed by hand with no chance fires on every qualifying hit.** Core turns a `ProcChance` of 0
  into 1; only `spelldata.ProcTrigger` panics on a row that states no rate.
- **`spelldata.Proc()` does not make a spell a proc.** It marks the spell passive with no cast or cost, and leaves
  `core.SpellFlagProc` off, so every trigger hears it. A spell the game triggers (a channel's tick, a wave, an item's
  bolt, a poison's hit) adds the flag from its row with `spelldata.Triggered()`, as Volley's waves, Hellfire Effect,
  Blizzard's tick and Consecration's pulses do. A spell built on the player's own cast row leaves it off.
- **A dot's tick is no cast, but a tick its aura casts is.** A listener that cannot proc from procs hears every dot's
  ticks, a proc's dot's included (Deadly Poison, Wound, Immolation Trap). It does not hear a tick that is the hit of
  a spell the aura casts each period: a Treat As Periodic row (`ATTR_EX_3_TREAT_AS_PERIODIC`, Consecration's pulses),
  or a periodic dummy or trigger with no periodic damage of its own (Ignite, Deep Wounds) (EV-0226).
  `spelldata.SpellConfig` sets `SpellFlagTicksAreCasts` from the row. Such a spell holds the dot, carries the proc
  flag and deals through `CalcPeriodicDamage` (`sim/paladin/consecration.go`). Whether a periodic listener that can
  proc from procs hears those ticks is unmeasured; the sim lets it. A wave without the attribute, Hurricane's or Rain
  of Fire's, is a direct hit.
- **A poison's or an imbue's hit is a proc's spell damage** (R-0051): `ProcMaskSpellDamage` with `SpellFlagProc`, so
  only a listener that can proc from procs hears it.
- **A mask that pairs the harmful-spell bit with melee or ranged bits needs damage.** A debuff or a dot's application
  that lands for nothing rolls none of the three items measured with that shape, and `DecodeProcTypeMask` applies the
  rule to every such mask (EV-0021, EV-0226). A mask of spell bits alone hears those landings.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. The commands that
re-derive each claim, the "You changed → Update" table and the path check are in `references/maintenance.md`.

A ruling, trap or workflow fact learned while using this skill is written down in the same change that uses it, where
`AGENTS.md` says: a comment at the code when one file owns the rule, otherwise this skill or a `docs/rulings/` file
its `references/rulings.md` indexes. A memory file is a pointer to this skill, not a second copy.
