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
  one charge per cast however many targets the cast hits (`core.OncePerCast`), and a channel's once at its end.
  Without a spender one proc pays for every cast in the buff's duration.
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
- **A spell's own proc mask decides which listeners hear it**, and the hit table decides the mask, not the school.
  The cases (a weapon proc's strike off either hand, oils) are in `references/rulings.md`.
- **A death listener (`CallbackOnKill`, `CallbackOnTargetDies`) gets a nil spell**, and a rated trigger hears no
  death; both are in `procs.md` (§ the two death bits), the order of a death in `wowsims-core-mechanics`
  `## Targets and encounter designs`.

## Item and enchant procs

Registration, every `NewSpellData*Proc` shape, the enchant slot rules, `ItemProcUnsupportedWithRate` and the two local
snapshot lists are in `procs.md` §Item and enchant procs; which builder an item effect routes to is
`wowsims-consumables-items`.

- **A rate the game files lack**: `wowsims-db-pipeline` `## Overrides`.

## Measuring a rate

- **Read the tooltip before the column.** The tooltip decides which `ProcChanceSource` shape a row is (the opening
  table of `procs.md`); 101 usually means no proc at all. A rate the game files lack
  is an override onto the row's `RPPM` first (`Field: PPM` in `spell_overrides.go`, Omen of Clarity is the case) or
  an override row on gear, and a `PPM(n)` literal with a review TODO only where no override fits.
- **An unstated rate follows the carry-over ruling** in `references/rulings.md`. Its `Source` drives the gear
  tab's notice for an unvalidated rate (`rateOrigin` in `tools/database/gen_effects.go`,
  `ui/features/gear/item_notices.tsx`): `vanilla-carryover` and `tbc-carryover` name their origin, a measured rate's
  `wcl:`, `issue #` or `combat log, ` source shows none. A gear rate kept as a literal in the sim is listed in
  `literalRates` there, or the notice never names it.
- **The class `.results` never move on a proc rate**: `wowsims-testing` `## Proving a claim at runtime`.
- **A measured rate that disagrees with the tooltip** is decided by `wowsims-spells`
  `## When a sim number disagrees with the game files`, and the outcome is a comment at the proc's code: Omen of
  Clarity's is in `sim/druid/omen_of_clarity.go`, Revelation's in `sim/common/forever/enchants.go`.

## Traps

- **A trigger named after its row never hears anything when the buff shares that name.** `ProcTrigger` names the
  trigger after the row, and `MakeProcTriggerAura` registers through `GetOrRegisterAura`, which hands back the
  `AuraConfig` buff of the same label and drops the listener. Set `trigger.Name = "<X> - Trigger"` (PR #140).
- **A periodic listener built from a row never hears a bleed.** `DecodeProcTypeMask` files
  `PROC_FLAG_DEAL_HARMFUL_PERIODIC` under `ProcMaskSpellDamage`, and a bleed tick carries a melee mask. Set
  `trigger.ProcMask = core.ProcMaskUnknown` at the call site, as Thousand Cuts does (`sim/rogue/talents_subtlety.go`).
- **A harmful-spell-only mask (0x10000) decodes to a cast trigger that ignores misses.** The decoder keeps that rule
  for items, so a spell that must land to proc moves to the hit listener at its call site, as Nature's Grace does
  (`sim/druid/talents_balance.go`, PR #382).
- **A `core.ProcTrigger` typed by hand with no chance fires on every qualifying hit.** Core turns a `ProcChance` of 0
  into 1; only `spelldata.ProcTrigger` panics on a row that states no rate.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. The commands that
re-derive each claim, the "You changed → Update" table and the path check are in `references/maintenance.md`.

A ruling, trap or workflow fact learned while using this skill is written down in the same change that uses it, where
`AGENTS.md` says: a comment at the code when one file owns the rule, this skill or its `references/rulings.md` otherwise. A memory file is a pointer
to this skill, not a second copy.
