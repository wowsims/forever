# Buffs and debuffs

Every raid, party, individual and enemy-debuff proto field has one row in
`tools/database/buffmanifest`. The manifest owns the field numbers and names the spell each row reads;
the store owns the values. From the same captured inputs, the `gen_spelldata` pass that renders the
store also renders:

- `sim/core/buffs/buffs_auto_gen.go` and `sim/core/buffs/debuffs_auto_gen.go`, a constructor per row;
- `ui/features/settings/model/buffs_debuffs_auto_gen.ts`, the settings inputs. An input's icon and
  tooltip use `TooltipID` where the row names one, else `CastID`, else `SpellID` (the choice never
  changes the numbers a row reads);
- `proto/buffs.proto`, the four messages.

They go through the same staging type-check and the same `-check` as the store, and regenerate without
the game files database.

## How a buff reads the store

The generated constructors live in `sim/core/buffs`, above core, because the store imports core. Each
supported row holds a package-level `*Meta` and reads every number off it at runtime. The read goes
through `spelldata.ParseEffects`, the same call a class's own aura reads its stats through, so a generated
buff carries no effect routing of its own:

```go
var battleShoutSpell = spelldata.MustFind(25289)
var battleShoutMeta = &Meta{
	Label:      "Battle Shout",
	Spell:      battleShoutSpell,
	Category:   BattleShoutCategory,
	SingleAura: true,
}

func BattleShoutValue(talentPoints int32) float64 {
	return battleShoutMeta.Value(talentPoints)
}
func BattleShoutDuration(talentPoints int32) time.Duration {
	return battleShoutMeta.Duration(talentPoints)
}
func NewSpellDataBattleShoutAura(unit *core.Unit, isPlayer bool, talentPoints int32) *core.Aura {
	return newBuff(unit, battleShoutMeta, isPlayer, talentPoints)
}
```

`Meta.Options(talentPoints)` states the `spelldata.ParseOpt`s every row needs:
`spelldata.RaidBuffOptions` (`Level(core.CharacterLevel)`, `BuffAuras`, `SkipAuras` where the
generator found any, and `FullComboPoints` for a finisher, which the generator also parses with) and
`ScaledBy` the improving talent. `ScaledBy` is left out where the talent scales the duration instead,
since `Duration` reads that separately. A row with a `TalentDivisor` takes
`ScaledByRanks(talentPoints, TalentDivisor)` instead: every amount the parse attaches is worth that
fraction of the spell's, rounded to the nearest whole amount the way the tooltip shows it, and 0
points is the full amount.

`newBuff` adds `SchoolResistances` and, where `Category` is set, `Exclusive(Category, true)` for a
`SingleAura` row or `ExclusivePerStat(Category)` for any other. A generated buff and a hand-written
scroll of the same stat or resistance therefore bid against each other under the categories core's own
exclusive stat buffs use. `newDebuff` adds only `Exclusive(Category, SingleAura)`, since a debuff never
carries a resistance of its own. Without `SingleAura`, each of its stats and pseudo-stats bids alone
under `Category` + its key, priced per stack. A stronger debuff thus holds only the effects of its own
kind, and Hex of Weakness keeps its healing taken under Demoralizing Shout's attack power. When an item
proc puts a row that a raid debuff reads on the target, the aura is that debuff's player copy
(`NewSpellDataDebuffAuraBySpell`). `newItemCountBuff` adds `Count`, so a party with several of the
same item is worth that many copies of the amount.

`Value` is a damage shield's own effect where the spell states one, and otherwise the first thing
`spelldata.DryRun` attaches for those options. `Duration` reads the spell, or `Cast` where the spell
states no duration of its own. `Cooldown` reads `Cast` where the `Meta` carries one (an external
cooldown whose `CastID` is not its `SpellID`), else the spell itself. Both go through the helpers in
`sim/core/buffs/amounts.go`, and a talent that scales the duration truncates it the way
`talentScaled` truncates an amount everywhere else.

A row whose aura is a damage shield skips the parse outright: `newDamageShield` calls
`NewDamageShield` with the spell's school and `Value`.

`sim/core/spelldata/buff.go` holds a second table, read only when a parse states `BuffAuras()`:
`A_PERIODIC_ENERGIZE` on a mana row becomes MP5 there, and `A_MOD_ATTACK_POWER_PCT` a percentage on
attack power. Neither sits in the parse's shared table. Warrior rows carry them for a value the class
wires itself (Bloodrage's rage tick is an `A_PERIODIC_ENERGIZE`, Berserker Stance Passive an
`A_MOD_ATTACK_POWER_PCT` of 0), and reading either off the shared table would change what those rows
mean there.

The generator asks the same parse about every manifest row. Given the row's options and no character,
`spelldata.DryRun` reports what the parse attaches, and `SkippedNotes` on that result lists what it
leaves out, so the sim builds a row the generator writes the same way. An effect the parse attaches at
0 beside the buff (the healing-taken row every paladin aura states) goes into the row's `SkipAuras`,
so the built aura leaves it out. A row whose meaning a driver decides (`KindExternalCD`, `KindProc`,
`KindManual`) is written whatever the parse attaches. A damage shield needs its shield effect, and
every other row needs an amount. Where the parse attaches nothing, the row renders as a commented
shell naming the reason. A row the generator does write states what it could not read as a
`// Left out:` note above it in the generated file, one per aura effect the parse has no row for.

Core applies the buffs through `core.BuffHooks`, which `sim/core/buffs` registers from its `init`:
`ApplyBuffs`, `ApplyDebuffs` and the Gift of Arthas elixir's debuff. A sim that
never imported the package panics naming the import. `sim/common` imports it, which links it into every
sim and every class test. Core's own tests reach it through the generated-buff tests, which sit in
`package core_test` beside `sim/core/export_test.go`.

## Adding a buff

1. Append the field's row to its message's slice in `tools/database/buffmanifest/buffs.go`. State the
   `Field` and the `SpellID` its numbers are read from; a `CastID` where the player learns a spell
   other than the aura, as for a totem; a `Talent` or an `ImpAction` where something improves it; a
   `Category` where other sources of the same effect bid against it; and the `Stats` it matters to. A
   row that is not simply its aura states a `Kind`.
2. `go run ./tools/gen_buffs_proto` and `make proto`, so the compiled protos carry the field. The pass
   below type-checks against them and writes nothing while the field is missing.
3. `go run ./tools/database/gen_spelldata`. The spell becomes a root of the store, and the constructor,
   the apply block and the settings input are written. Read the `buffs:` lines it prints.
4. Where the kind cannot express the row outright, the row states `Driver: true` and
   `sim/core/buffs/drivers.go` declares `drive<Go>`.
5. With a database, `TestManifestAnchorsMatchTheGameFiles` checks that the pin is the top rank the owner's skill
   lines grant.

## The manifest row

`BuffSpec` in `tools/database/buffmanifest/manifest.go` holds what the game files cannot state about a
field. `buffs.go` holds the rows in four slices, `Raid`, `Party`, `Individual` and `Debuffs`, one per
proto message and in the order the settings tab lists them. A row's scope is the slice it sits in, and
its proto number is its position there. Each message's numbers are therefore dense from 1, and a row
placed anywhere but the end renumbers the rows after it. `buffmanifest.All()` pairs every row with both as a
`Row`.

| Field            | What it is                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Field`          | the proto field, snake_case                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `SpellID`        | the spell the numbers are read from, and a root of the store: the top rank of the castable family, or the aura that family's cast applies when the cast is a summon or a dummy                                                                                                                                                                                                                                                                                                                                                           |
| `CastID`         | the cast the player learns, where it is not `SpellID`. It names the buff and times an external cooldown: Mana Tide's aura 1325257 carries the mana, the cast 1325256 the 13 seconds and the 5 minutes                                                                                                                                                                                                                                                                                                                                    |
| `TooltipID`      | the settings input's icon and tooltip spell, for a row whose input shows a spell that is neither its aura nor its `CastID` (Moonkin Aura's row points at Moonkin Form, the Judgements' rows at their Seals, Vindication's at the talent). `0` defers to `CastID`, then `SpellID`; it never changes the numbers the row reads                                                                                                                                                                                                             |
| `Talent`         | the spell of the trait node that prices the improved state. Its one spell modifier whose class mask reaches `SpellID` is the improvement (two fail the pass), and a `SPELLMOD_DURATION` modifier scales the duration, any other the value. The generated `Meta.TalentEffect` names that effect by its `EffectN` position. Where no modifier reaches it, a tooltip term `${$m1/-N*$<SpellID>m1}` prices the buff at rank/N of its amount (Vindication), which the generated `Meta.TalentDivisor` states as N; with neither the pass fails |
| `Category`       | the exclusive-effect category the aura bids in, `""` for none                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `SharedCategory` | a second category the aura joins without an effect of its own, which is how the paladin auras exclude each other across schools. Applied to the player's copy only, and declared once in the generated file as `<Name>Category`                                                                                                                                                                                                                                                                                                          |
| `SingleAura`     | the category holds one aura at a time, so the loser is deactivated rather than outbid                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `Driver`         | the apply block hands the row to `drive<Go>` instead of activating the aura outright                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `Stats`          | the stat tags the settings' buff filter matches against; the filter starts from a spec's `displayStats`, `epStats` and included stats                                                                                                                                                                                                                                                                                                                                                                                                    |
| `ImpAction`      | the improved state's source when it is not a talent (an item, or the spell an item set grants at a piece threshold), and the icon that state shows                                                                                                                                                                                                                                                                                                                                                                                       |
| `Label`          | a UI label override                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `Kind`           | what the row is when it is not simply the aura its spell states, below                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `Proto`          | an override of the derived proto type, for a flag that is a number: `retribution_aura_spell_power` is `ProtoDouble`                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `Owner`          | the providing class, only for a spell no class family files: the Atiesh rows, pet abilities and talent-granted debuffs. A row whose spell a family files and that names an owner fails the pass                                                                                                                                                                                                                                                                                                                                          |
| `Reason`         | what a `KindFlag` row is, which the shell it renders as carries. Required on a flag and refused on any other row                                                                                                                                                                                                                                                                                                                                                                                                                         |

The last four are exceptions; most rows state none of them. The rest of a row is derived:

- **Go stem.** `GoStem()` is `GoField()`, protoc-gen-go's name for the field, without underscores:
  `battle_shout` gives `NewSpellDataBattleShoutAura`, `BattleShoutValue`, `BattleShoutDuration`,
  `BattleShoutCategory`, `battleShoutSpell` and `battleShoutMeta`. A stem ending in Aura takes no
  second one: `devotion_aura` gives `NewSpellDataDevotionAura`.
- **Proto type.** `ProtoType()` is `ProtoTristate` where a `Talent` or an `ImpAction` prices an improved
  state, `ProtoInt32` for `KindExternalCD` and `KindItemCount`, and `ProtoBool` otherwise. The pass
  checks it against the compiled message and renders a row as a shell while protoc has not yet retyped
  its field.
- **Name and AuraName.** `Name` is the game files' `SpellName` of `CastID`, or of `SpellID` where no cast
  is pinned. `AuraName` is the `SpellName` of `SpellID`, only where a cast is pinned. `Name` is the label
  unless `Label` overrides it; a spell no class family files has neither and is labelled by its own name. `TestManifestAnchorsMatchTheGameFiles` resolves the
  pins from both.
- **Owner.** The providing class: the one whose `core.ClassSpellFamilies` entry is the spell's family,
  or, for a spell no class family files (the Atiesh rows, pet abilities, talent-granted debuffs), the
  row's own `Owner`. It narrows the rank
  lookup, and marks the row "(External)" on that class's settings tab.
- **SkipAuras.** Every effect the parse attaches at 0, as above: the six paladin auras skip
  `A_MOD_HEALING_PCT`.
- **Damage shield.** A plain row whose spell applies `A_DAMAGE_SHIELD` is `KindDamageShield`.
- **Talent ranks.** The trait node's ranks are the points the apply block hands the constructor for an
  improved state.
- **Totem.** A row whose cast summons into one of the four totem slots is a totem: an `E_SUMMON`
  effect whose second misc value is SummonProperties 63, 81, 82 or 83 (fire, earth, water, air). The
  game files database carries no SummonProperties table, so the four ids are written in
  `tools/database/gen_buffs.go`. See [Totems on pets](#totems-on-pets).

## The kinds

`KindPlain`, the zero value, is a row whose aura the parse builds outright, and `KindDamageShield` is
read off the spell, as above. The rest are stated: `KindExternalCD` is a count of other players' casts,
which a driver schedules on the cast's cooldown; `KindProc` needs a driver for the trigger;
`KindItemCount` takes a count and applies its amounts once per item; `KindManual` is a row the sim
models by hand through its driver; and `KindFlag` is a sim input rather than a buff (a toggle, or a
number such as `retribution_aura_spell_power`), which names no spell and renders as a commented shell
carrying its `Reason`. A row renders as a shell too when its spell states no aura effect the parse
attaches, when an external cooldown's cast states no cooldown or its aura no duration, or when the
compiled field has another type. No `Proto` value is an enum. A field that wants one would add its own
value, and a name for it, in both emitters.

## The apply order

`buffmanifest.Scopes` is Party, Raid, Individual, Debuffs: the rows resolve, render and apply in that
order, each scope in slice order. The apply order is the order the auras register in, and the results
follow it. The party's Retribution Aura and the individual Thorns are both damage shields. A hit taken
reaches them in the order they activate, which is the order they registered, so the protection suites'
results hold with Party first.
`proto/buffs.proto` keeps its own message order: Raid, Party, Individual, Debuffs.

## Warnings

The pass prints a `buffs:` line for what it resolves but doubts, and writes the row regardless:

- **An unpriced talent.** A passive trait node whose flat or percent spell modifier raises an effect
  value of the row's spell, on a row that names no `Talent`. It prices an improved state the row does
  not offer.
- **A talent with nothing to scale.** A `Talent` that scales the value of a row the parse attaches no
  amount of. The row keeps no ranks, so `TestResolvedBuffInvariants` rejects it as a tristate nothing
  prices.
- **A scope mismatch.** The game files state an area or a target the row's slice does not name.
  `TestScopeMatchesTheGameFilesTargeting` holds the exemptions.

## What stays hand-written

Judgement of the Crusader states its effect in a shape no manifest row can carry (a holy school
alone), so its row is a shell, and `applyDebuffs` in `sim/core/buffs/debuffs.go` applies the
hand-written aura after the generated ones. The paladin's own aura and judgement ranks
(`PaladinAuraRank`, `JudgementRank`) live in `sim/core/buffs/paladin.go` beside the generated
categories they join. Each is built at the rank the paladin cast: `NewSpellDataDevotionAuraRank`
beside the generated top rank `NewSpellDataDevotionAura`, and likewise for Concentration, Retribution and
the three resistance auras; `NewSpellDataJudgementOfLightAuraRank`,
`NewSpellDataJudgementOfWisdomAuraRank` and `NewSpellDataJudgementOfTheCrusaderAuraRank` beside the
generated `NewSpellDataJudgementOfLightAura`, `NewSpellDataJudgementOfWisdomAura` and the commented
shell `NewSpellDataJudgementOfTheCrusaderAura`.

## Drivers

A row the generator cannot express outright either states `Driver: true` or is a kind that always
needs one, and the apply block calls `drive<Go>` instead. The contract is in
`sim/core/buffs/drivers.go`: a buff row's driver takes the `*Character` and the whole scope message,
and a debuff row's takes the `*Unit`, the debuffs and the raid. Because the driver gets the whole
message, Flametongue Totem can read `raid.WindfuryTotem`. The driver builds the generated aura with its
`NewSpellData...` constructor and adds what the game files do not state: a proc trigger, a cooldown, a
delay, the metrics of a regen. The row compiles once the function is declared.

## Totems on pets

A true pet stands under every totem its owner does whose aura a pet may carry, with the same
exclusivity; a guardian (`PetConfig.IsGuardian`) stands under none. The apply block of a totem row that no driver handles calls
`standTotem` in `sim/core/buffs/totems.go`, which makes the owner's copy permanent and builds the same
constructor on each true pet's unit, so the pet's copy bids in the same categories on the pet. The
three driven totems pass their pet copies to `inheritTotem` themselves: Mana Tide's copy of the
external cooldown's aura, and the weapon totems' whole chain (totem aura, category bid and proc) on the
pet's character, for a pet that melees. Flametongue Totem does nothing under Windfury Totem; Grace of
Air stacks with it as of 1.60.1.70425.

The exception is a totem whose aura spell sets the only on player bit of Attributes[3]
(`ATTR_EX_3_ONLY_ON_PLAYER`, read by `Spell.OnlyOnPlayers`): a target that is not a player is invalid,
so `inheritTotem` gives no pet a copy. Every caller passes the spell the check reads: the row's own
spell for Mana Tide and every generated totem, and the raid aura (10612, 15036) for a
weapon totem. Strength of Earth, Grace of Air, Windfury Totem and Flametongue Totem set the bit, so a
pet carries none of them.

The pet's copy never expires on its own and is never in a build phase. It is up while the owner's
copy is up and the pet is summoned: the owner's copy turns it on and off, and the pet's enable and
disable effects (`Pet.RegisterEnableEffect`, `Pet.RegisterDisableEffect`) put it back when a pet is
summoned under a standing totem and take it off when the pet is dismissed. Mana Tide reaches the pet
this way without the pet running a cooldown of its own.

## Regenerating

`go run ./tools/database/gen_spelldata` writes the buff files with everything else, and `-check` names
them when they are stale. `tools/database` imports `sim/core`, so the generator cannot run while the
tree it generates into does not compile. A change that breaks the build (retiring a proto field, or
deleting a driver the generated file still calls) needs the generated file cut by hand first: delete
the lines that name the field or symbol that is going away, run `go build ./...`, then run the
generator, which writes the whole file back. `tools/gen_buffs_proto` renders `proto/buffs.proto` alone, with nothing but
the manifest, for the one moment the sim cannot build: before protoc has seen a new field.

## The guard tests

`go test ./tools/database/...` needs no game files database. It is the "Test the buff generator" step
of `.github/workflows/run_tests.yml`, the gate list run locally before each commit; only
`TestManifestAnchorsMatchTheGameFiles` and `TestProcShapeOfNamedSpells` skip without a database.

| Test                                                                       | What it holds                                                                                                                                        |
| -------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `TestUniqueScopeField`, `TestUniqueGoStem`                                 | no two rows collide                                                                                                                                  |
| `TestFlagRowsHaveAReason`, `TestResolvableRowsNameASpell`, `TestProtoType` | the schema rules above                                                                                                                               |
| `TestBuffsMatchProto`                                                      | the committed `proto/buffs.proto` declares every row at its number and type, and nothing else                                                        |
| `TestFieldNaming`, `TestFieldNamesRoundTrip`                               | `GoField()` and `TSField()` reproduce protoc's and protobuf-ts's camel case                                                                          |
| `TestRenderProtoNextIndex`, `TestRenderProtoHeader`                        | the next free number above each message, and the header protoc reads                                                                                 |
| `TestBuffFilesRegenerateFromTheCommittedInputs`                            | the two Go files, the settings inputs and `proto/buffs.proto` are what the committed store inputs render                                             |
| `TestRenderedBuffFilesMatchTheFixtures`, `TestRenderedBuffFilesCompile`    | synthetic rows of every shape render to the committed fixtures, and those fixtures compile against the real `sim/core/buffs` through a build overlay |
| `TestRenderBuffsDebuffsTS*`                                                | the settings inputs each proto type and kind renders                                                                                                 |
| `TestResolvedBuffInvariants`                                               | a tristate row is priced by a talent or an `ImpAction`, talent ranks match the trait node, pinned categories hold, amounts match the spell's row     |
| `TestScopeMatchesTheGameFilesTargeting`                                    | a row whose spell states an area aura, or an aura aimed over an area, sits in the scope that targeting names                                         |
| `TestManifestAnchorsMatchTheGameFiles`                                     | with a database, each pinned spell is still the top rank, aura or cast the game files' skill lines grant                                             |

`sim/core/buffs_generated_test.go` and `sim/core/debuffs_generated_test.go` hold the generated
constructors' behaviour: categories, stacks and drivers. Rewrite the synthetic
fixtures with `UPDATE_BUFF_FIXTURES=1 go test ./tools/database/`.

## Pitfalls

**A negative amount that scales by level floors away from zero.** `Average` floors an aura's folded
amount, so Demoralizing Shout's -196 and -1.4 a level read -205 at 60, and Demoralizing Roar's -193 does too.
A game build that truncated the per-level part toward zero would state -204. The sim takes the store's
reading until the game says otherwise.

**A debuff prices at the character's level, not the target's.** `Meta.Options` states `Level`
unconditionally, so `newDebuff` reads the row at `core.CharacterLevel` (60) whatever level the target
carries. Demoralizing Shout's -196 and -1.4 a level reads -205 against a level 60 target and still -205
against a level 63 raid boss, where pricing at the boss's own level would give -209. A raid's debuff is
cast by a player of the sim's level, and the target has no character for the parse to read a level off
in the first place.

**`A_PERIODIC_ENERGIZE` and `A_MOD_ATTACK_POWER_PCT` read only for a buff.** Both sit in the table
`BuffAuras()` adds rather than the shared one. Bloodrage reads its rage tick by hand
(`sim/warrior/bloodrage.go`). Berserker Stance Passive's own `A_MOD_ATTACK_POWER_PCT` of 0 is already
parsed onto the stance aura through the shared table (`sim/warrior/stances.go`), which skips and
reports it because the shared table names no row for it. A shared-table entry for either aura would
attach a stat change there instead and change what the class's own row means.

**Dropping a row renumbers the ones after it.** Each scope's numbers are dense from 1, which
`TestScopeNumbersAreDense` holds, so a field that goes away shifts every later number down by one.
Nothing is live, so no saved payload rides on the old numbers.

**A proto change here breaks whatever a browser holds.** Nothing migrates a saved payload: a field the
manifest retypes makes `fromJson` throw on the enum name an older save spells out. Every `fromJson` of
a settings envelope passes `ignoreUnknownFields`, so a field that only goes away costs nothing.

**Drums are not a Forever consumable.** The game files describe no row for the TBC drums (35476, 35475
and 35478: Battle, War and Restoration) nor for their Greater variants, so there is nothing to model and
no manifest row to hang them on. The only "Drums of War" in the game files is 1259907, fifteen seconds
of party movement speed and no stat buff at all.

**A talent curve only scales the row's first stat.** A row whose talent improves a second amount would
need the generator extended; no manifest row has such a talent.

**A set bonus is not a talent, but it can be a tristate.** The resolver reads `SkillLineAbility`, the
trait trees and the spell effects. It does not read `ItemSetSpell`, so a set that modifies a buff
cannot be a manifest `Talent`. It can be the row's `ImpAction`: the improved state a tristate row needs
when no trait node prices one, and the icon that state shows. Battlegear of Wrath is the case: item set
218's `ItemSetSpell` at three pieces (2336) is 23563, an `A_ADD_FLAT_MODIFIER` of 30 against every
effect of the Battle Shout family, which raises the shout from 139 attack power to 169. The two halves
of that are modelled separately. The warrior's own cast reads the `has_bs_t2` class option (the user's
word that this warrior wears the set, not the equipped gear). The raid's copy is `battle_shout`'s
improved state, which `driveBattleShout` reads because the resolver has no curve to give it. Both call `AddGeneratedFlatBonus`, which raises what the aura
applies and what it bids for its category together, so the stronger of the two copies is the one the
character sheet shows. It is told what the buff is worth without the bonus, because the aura belongs
to the unit rather than to whoever raised it: two warriors in a party wearing the same set ask for
the same total and the second call does nothing. The 30 itself lives in
`buffs.BattleShoutT2Bonus`, with the set and the spell it came from written next to it.

**A party or raid flag means an external caster provides the buff.** The generated apply block builds
the row's `isPlayer=false` copy whenever the proto field is set, so a class port that registers its own
`isPlayer=true` copy has two copies on the character. Either the port stops setting the flag in
`AddPartyBuffs`/`AddRaidBuffs`, or the row states a `Category` with `SingleAura` and the two copies bid
against each other, so the character sheet shows the buff once. The higher bid deactivates the other
copy. On a tie the incumbent keeps the category when its remaining duration is the longer one, so a
druid casting its own Thorns is turned away while the settings' permanent copy is up. Both deal the same
22, so the character strikes back for the same amount either way. Battle Shout is the worked example.
Neither copy is permanent, and the two are worth the same unless one side wears the tier 2 set, so
`TestPlayerBattleShoutTakesTheCategoryOnATie` holds the player's own copy to the tie and
`TestTheStrongerBattleShoutTakesTheCategory` holds the stronger one to the rest. The rows this decides
are `thorns`, `leader_of_the_pack`, `moonkin_aura` and `trueshot_aura`: `sim/druid/druid.go`,
`sim/druid/feralcat` and `sim/druid/feralbear` raise the raid's Leader of the Pack or Moonkin Aura
from a talent, `sim/hunter/hunter.go` raises Trueshot Aura, and the druid's own Thorns waits on the
druid port. `thorns`, `battle_shout`, `leader_of_the_pack` and `moonkin_aura` carry a category, so a
druid registering its own copy of any of them has nothing to decide: the copy joins the same category
and the two bid. Leader of the Pack and Moonkin Aura each have their own category: one raises melee and
ranged crit, the other spell crit, so a raid that ticks both gets both. Only `trueshot_aura` is left
without a category, so the hunter port has to choose. The priest's three prayers (`registerPrayer` in
`sim/priest/priest.go`) bid without `SingleAura`: `prayer_of_fortitude` and `prayer_of_spirit` each
stat on its own in `StatBuff`, with Arcane Brilliance and the scrolls, and `prayer_of_shadow_protection`
in the shadow resistance category every buff joins. A priest's own prayer and the raid's count once,
while Gift of the Wild and Blood Pact still add their Stamina on top.
