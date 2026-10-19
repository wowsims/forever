# Consumables

A consumable reaches the sim in two halves: its entry in `assets/database/db.json`, which the UI offers
and the sim looks the id up in, and an effect registered for its id, which the sim applies when the
character's `ConsumesSpec` names it. The generator writes the first from the game files, and writes the
second for every consumable a builder in `sim/common/shared` can run from its rows. Where the rows are
not enough or not right, the effect is hand-written.

- [How a consumable gets into db.json](#how-a-consumable-gets-into-dbjson)
- [Regenerating](#regenerating)
- [How the effect is registered](#how-the-effect-is-registered)
- [Exclusive groups](#exclusive-groups)
- [Stats db.json lists outside the parse](#stats-dbjson-lists-outside-the-parse)
- [Writing an effect by hand: the Natural flasks](#writing-an-effect-by-hand-the-natural-flasks)
- [Adding a consumable the rows do not support](#adding-a-consumable-the-rows-do-not-support)
- [Removing a consumable](#removing-a-consumable)
- [Correcting a number](#correcting-a-number)
- [Tests](#tests)
- [What the settings offer](#what-the-settings-offer)
- [The browser path](#the-browser-path)

## How a consumable gets into db.json

**The catalogue.** `ConsumableItemFilter` in `tools/database/tables.go` selects every item that states
an on-use spell and is one of:

- a Consumable-class item, except the permanent enhancements of subclass 6, which the enchant table
  carries;
- a Trade Goods explosive;
- an item whose on-use shares the explosives' spell category or casts a temporary enchant, whatever
  class the game files put it under: Crystal Charge 11566 is a Quest item.

An item that also teaches a spell is an Engrave rune and is left out, and so is a name containing
`Test`, `[DEP]` or `[PH]` or starting with `QA` or `UNUSED` (UNUSED Flask of Merciless Pursuit
215126). The ids in `ConsumableAllowList` and `ConsumableExclusiveGroups`
(`tools/database/overrides.go`) join the catalogue by id, whatever the filter says
(`catalogueConsumableIDs`). The catalogue's spells, and everything they trigger, are carried in the
spell store.

**Routing.** `RouteConsumable` in `tools/database/gen_consumables.go` reads the spell the consumable's
first on-use row casts. Damage, heals and absorbs are routed the way an item's on-use is, and a
missile that leaves a cloud dealing damage over time (`shared.DamageCloudOf`) routes as damage;
otherwise the first effect of the spell, or of a spell it triggers on the user, that fits a shape
decides it:

| kind                | what the spell does                                                    |
| ------------------- | ---------------------------------------------------------------------- |
| `damage`            | deals damage (Goblin Sapper Charge 10646: 13241)                       |
| `heal`              | heals the user (Major Healing Potion 13446: 17534)                     |
| `energize`          | restores a resource (Major Mana Potion 13444: 17531)                   |
| `stat aura`         | applies an aura the parser attaches (Flask of the Titans 13510: 17626) |
| `percent stat`      | multiplies a stat; none routes there in the game files                 |
| `temporary enchant` | enchants a weapon (Brilliant Wizard Oil 20749: enchant 2628)           |
| `absorb`            | shields the user (Greater Fire Protection Potion 13457: 17543)         |

A consumable that fits no shape, or fits one the rows do not fully state (an aura with no duration, an
enchant slot the sim does not read), is unsupported, with a reason. A heal, an energize or an absorb is
supported only where the potion builder runs the spell: `shared.ConsumableUnsupported` names what stops
it, such as a stun on the drinker (Dreamless Sleep Potion 12190: 15822), an `E_DUMMY` beside the heal
(the Discolored Healing Potions) or a spell that restores only while the user sits (Lollipop 20388:
24707).

**What ships.** `ShipsConsumable` puts a catalogue consumable into db.json when it routes to a shape, or
when `ConsumableAllowList` or `ConsumableOverrides` names it. `ConsumableOverrides` is merged next, field
by field onto the game files' entry, or as a new entry where the game files have none. `KeepConsumable` then
drops what `ConsumableDenyList` names and the test and QA names; the allow list wins over the deny list.
`AddConsumableIcons` puts every shipped consumable into the item icon table, where the UI looks its icon
up: the game files' icon, or for an entry the game files do not ship, the `Icon` its override states.

**The unsupported list.** `tools/database/unsupported_consumables.txt` names every catalogue consumable
that routes to no shape, as `<item id> <spell id> <name>: <reason>`, after the count routed to each
shape. It is a local baseline, not tracked: `TestEveryConsumableIsRoutedOrListed` writes it on the first
run and then names every line that moves. Once the moves are the intended ones, rewrite it:

```sh
UPDATE_UNSUPPORTED_CONSUMABLES=1 go test ./tools/database/ -run TestEveryConsumableIsRoutedOrListed -count=1
```

## Regenerating

`make db` extracts the game files, then regenerates the spell store and then db.json. With
`tools/database/wowsims.db` already in place, the last two steps are:

```sh
go run ./tools/database/gen_spelldata
go run tools/database/gen_db/*.go -outDir=./assets -gen=db
```

The order matters: gen_db routes every consumable through the store compiled into it, so a change to
`ConsumableAllowList` or to a spell override needs the store rewritten first. The same gen_db run
rewrites the registrations below.

## How the effect is registered

`sim/core/consumes.go` keeps one effect per consumable id, and applies the registered effect of every id
the `ConsumesSpec` names. There are two ways in:

- `core.NewConsumableEffect(id, apply)` registers a hand-written effect.
- `core.NewSpellDataConsumableEffect(id, apply)` registers an effect a builder in `sim/common/shared`
  made from the consumable's rows.

A hand-written effect wins over a data-driven one for the same id, whichever registers first, so the
two can live in different packages with no init order between them. Two hand-written effects, or two
data-driven ones, for one id panic. An id db.json does not carry registers nothing.

`GenerateConsumableRegistrations` (`tools/database/gen_consumable_registrations.go`) writes the
data-driven registrations into `sim/common/forever`, one file per family, and
`common.RegisterAllEffects` runs them. It picks each consumable's builder once, by its type, the
category its on-use shares and how its spell routes, in the order of the table. The explosive and
potion files keep a consumable their builder refuses, commented out with the reason. Each builder
states what it refuses beside itself (`shared.ConsumableUnsupported`, `DamageConsumableUnsupported`,
`StatConsumableUnsupported`), and the router reads the same checks.

| family                                                         | file                                 | builder                        |
| -------------------------------------------------------------- | ------------------------------------ | ------------------------------ |
| `ConsumableTypeImbue`                                          | `consumables_imbues_auto_gen.go`     | `NewSpellDataTemporaryEnchant` |
| `ConsumableTypeExplosive` that routes as `damage`              | `consumables_explosives_auto_gen.go` | `NewSpellDataExplosive`        |
| potions (category 4) and `ConsumableTypeConjured`              | `consumables_potions_auto_gen.go`    | `NewSpellDataOnUseConsumable`  |
| other consumables that route as `heal`, `energize` or `absorb` | `consumables_potions_auto_gen.go`    | `NewSpellDataOnUseConsumable`  |
| other consumables that route as `stat aura`                    | `consumable_stats_auto_gen.go`       | `NewSpellDataStatConsumable`   |

The potion builder runs one spell on the user: a heal, an absorb, a gain of mana, rage or energy, damage
to the user and the auras the parser attaches. A spell that damages enemies it deals through the
damage builders instead (`consumableDamageSpell`, which the explosives use too), with nothing beside
it on the user; the potion is a DPS cooldown and never an explosive. A cone or an area around the user
that states no range reaches as far as its radius. A row with Fixed Travel Time lands after its Speed
in seconds at any distance (`docs/spell_data/building-an-ability.md`), and `cloudDamageSpell` puts a
cloud's damage over time on every enemy. A
potion shares category 4's timer and is a major
cooldown only as one of the spec's three potion picks (`resource_pot_id`, `defensive_pot_id`,
`offensive_pot_id`), which the APL's Resource, Defensive and Offensive Potion placeholders cast; a
conjured item is one only as the spec's conjured item; any
other is one whenever the spec names it. The builder tells them apart by the consumable's type. Damage a
consumable deals its own user, a sapper's or Demonic Rune's, always lands and can crit. The stat builder
puts an aura that outlasts the item's cooldown up from the pull, and uses the item on its cooldown
otherwise.

Every builder reads an item's cooldowns the same way (`consumableCooldownsOf`): the item row's own
cooldown and the category it shares, where the item row states -1 or no category, the spell row's.
`consumableCast` puts them on the spell a builder registers; a stat consumable that is up from the
pull registers no spell and no cooldown. Which consumables stack is their
[exclusive group](#exclusive-groups).

A bonus against creature types is an aura the parser attaches like any other (see
[Spell Data](spell_data/README.md)): Potion of Demon Slaying 9224's spell 11406 grants 180 attack power
against the creature types its misc value masks (4, Demon), for the row's 300 s, on category 4's
120 s timer; db.json lists that bonus as a [mob type stat](#stats-dbjson-lists-outside-the-parse). The imbue builder takes the equip spells of the enchant that add nothing to
the enchant's own stats, which the generator names after the item: Consecrated Sharpening Stone
23122's enchant 2684 names 28893, and core puts the enchant on each hand whose weapon the masks take.
28893 grants 99 melee and 99 ranged attack power against Undead once per such hand, since each weapon
carries its own equip aura and 28893 marks no aura unique. Blessed Wizard Oil 23123's enchant 2685
names 28890, 58 spell damage against Undead (`A_MOD_FLAT_SPELL_DAMAGE_VERSUS`, mask 32). A wizard
oil's equip spell is already the enchant's spell damage, and is not named. Elemental Sharpening Stone
18262's enchant 2506 names 22755, 2% crit (`A_MOD_WEAPON_CRIT_PERCENT`) on the melee weapons its mask
42483 names; the generated entry parses it once per stoned hand with `spelldata.OnHand`, so the crit
reaches only the attacks of the hand carrying the stone.
`consumables_gift_of_arthas.go` registers Gift of Arthas 9088.

A pet takes no consumables of its own; the scrolls the game files mark `ONLY_TARGET_PLAYERS` cannot be
used on one.

## Exclusive groups

A consumable belongs to at most one exclusive group. Of the consumables the spec names, only the first
of each group applies, in the order core applies them (the potion, the flask, `elixir_ids`, the food,
the conjured item, the explosive, the imbues, `extra_ids`); everything outside a shared group stacks.
The game files carry no stacking table, so the generator resolves each consumable's group once, and
`ShippedConsumable` writes it into db.json as `group_key` and `group_name`. The sim reads only the key
(`selectedConsumableIDs` in `sim/core/consumes.go`), the settings the key and the name.

**From the rows** (`rowGroupOf` in `tools/database/consumable_groups.go`):

- an elixir's or a scroll's group is the aura type and misc value of the first aura its on-use spell
  applies, keyed `<aura>/<misc>`. Elixir of Agility 8949 and Elixir of the Mongoose 13452 both open on
  `A_MOD_STAT` Agility (29/1), so only one of them applies; Elixir of the Sages 13447 (29/3, Intellect
  then Spirit) and Elixir of the Whale 250348 (29/4, Spirit) stack, and so do Arcane Elixir 9155
  (`A_MOD_DAMAGE_DONE` every school, 13/126) and Elixir of Fire Power 6373 (13/4);
- a scroll therefore joins the elixirs of the same key: Scroll of Agility 3012 to IV 10309 (29/1) join
  the Agility elixirs, the Strength scrolls the Strength elixirs and Juju Power, Protection the Armor
  elixirs, Intellect and Spirit theirs. A scroll whose key no elixir has shares its group with its
  ranks alone: Scroll of Stamina 1180 to IV 10307 (29/2), and Scroll of Minor Evocation 213564
  (`A_PERIODIC_ENERGIZE`, 24/0), whose Mageblood elixirs open on `A_MOD_POWER_REGEN` (85/0). A scroll
  still yields to the raid buff of its stat;
- a stat consumable whose on-use category cooldown lasts as long as each of its buffs, which are up
  from the pull, is used one at a time with the others of the category (`shared.OneAtATimeCategory`),
  keyed `category <id>`. Only the Blasted Lands buffs (R.O.I.D.S. 8410, Lung Juice Cocktail 8411,
  Ground Scorpok Assay 8412, Cerebral Cortex Compound 8423, Gizzard Gum 8424) qualify: category 103's
  1 h, buffs of 1 h.

Anything else has no group from its rows: a flask stacks with every elixir, since the 3 s of the
category 79 they share is shorter than any of their buffs.

The label of a row group comes from its aura (`auraGroupName`): "Strength", "Spell Power", "Fire Spell
Power", "Armor", "Mana Regeneration", and so on.

**From the override table.** `ConsumableExclusiveGroups` in `tools/database/overrides.go` holds the
server rules the rows do not state, and a consumable it names takes that group instead of its rows'.
An entry's `Group` either is a row group (`ElixirGroup(aura, misc)`, which carries that group's key and
label) or a group of its own, with a key and a label of its own:

```go
var ConsumableExclusiveGroups = []ConsumableGroup{
	{Group: ElixirGroup(dbcenums.A_MOD_STAT, 0), Items: []int32{
		12451, // Juju Power
	}},
	{Group: ElixirGroup(dbcenums.A_MOD_ATTACK_POWER, 0), Items: []int32{
		12460, // Juju Might
	}},
	{Group: ConsumableGroupKey{Key: "Zanza", Name: "Zanza"}, Items: []int32{
		8410,  // R.O.I.D.S.
		8411,  // Lung Juice Cocktail
		8412,  // Ground Scorpok Assay
		8423,  // Cerebral Cortex Compound
		8424,  // Gizzard Gum
		20079, // Spirit of Zanza
	}},
	{Group: ConsumableGroupKey{Key: "Hit food", Name: "Hit food"}, Items: []int32{
		22237, // Dark Desire
		23327, // Fire-toasted Bun
	}},
}
```

The Jujus are Quest-class items the catalogue filter leaves out; naming them here brings them into
the catalogue as well. The Blasted Lands buffs leave their category group for Zanza. The permanent
Zanza items (23794, 23795, 23796) are not obtainable and `ConsumableDenyList` keeps them out. The two
hit foods' 2% hit is listed outside the stat parse, as [below](#stats-dbjson-lists-outside-the-parse).

**Adding a group.** Add an entry to `ConsumableExclusiveGroups`: `ElixirGroup` to put items into a row
group, or a new key and label for a group of their own, and every item id with its name. An item the
catalogue filter leaves out joins the catalogue through the entry; one whose spell routes to no shape
needs `ConsumableAllowList` as well to ship. Then regenerate (store first, then
db.json). `TestEveryExclusiveGroupOverrideShips` fails for an id that does not ship or sits in two
groups; `TestAConsumableTakesItsExclusiveGroup`
pins the groups of the consumables it names. The settings pick the new group up by its key with no
change of their own.

## Stats db.json lists outside the parse

The settings offer and order a consumable by the stats db.json lists for it. `AddUnparsedStats`
(`tools/database/consumable_unparsed_stats.go`) adds what the on-use grants that the stat parse leaves
out.

**One-channel percents.** The two hit foods apply `A_MOD_HIT_CHANCE` 2, which the stat parse does not
list and which reaches melee and ranged attacks alone, so their db.json `pseudoStats` carry 2
`PseudoStatMeleeHitPercent` and 2 `PseudoStatRangedHitPercent` (the ranged total includes the melee
share, as `stats.PercentPseudoStats` reads it). The other one-channel auras go the same way:
`A_MOD_SPELL_HIT_CHANCE` lists spell hit (Scroll of Arcane Accuracy I 211954, 1%),
`A_MOD_WEAPON_CRIT_PERCENT` melee and ranged crit, and `A_MOD_SPELL_CRIT_CHANCE` spell crit (Scroll of
Arcane Power I 211957, 1%). `A_MOD_CRIT_PCT` reaches every channel, so it lists melee, ranged and
spell crit alike: Elixir of the Mongoose 13452 and the 250337/250338/250348/250351 elixirs each carry
2, listed as 2 `PseudoStatMeleeCritPercent`, 2 `PseudoStatRangedCritPercent` and 2
`PseudoStatSpellCritPercent`. A spec's default consumable filter holds the percents its
`displayStats` or `epPseudoStats` name (`consumablePseudoStatsOf` in `ui/sim/settings/consumables.ts`),
whatever hit or crit rating it weighs, so by default the Hit food picker shows for the melee and
ranged specs and not for the casters.

**Mob type stats.** A versus aura (`spelldata.VersusStat`: `A_MOD_MELEE_ATTACK_POWER_VERSUS`,
`A_MOD_RANGED_ATTACK_POWER_VERSUS` or `A_MOD_FLAT_SPELL_DAMAGE_VERSUS`) grants its stat only against
the creature types its misc value masks. Its amount goes into `mob_type_stats` (`mobTypeStats` in
db.json, `MobTypeStats` in `proto/common.proto`) under the mob types that mask names
(`spelldata.MobTypes`): Potion of Demon Slaying 9224 lists 180 attack power under Demon, Potion of
Elemental Purging 250954 105 spell damage under Elemental. A versus aura whose mask names no mob type
is listed as a plain stat.

**Zone stats.** A Natural flask's zone bonus goes into `zone_stats` (`zoneStats`, `ZoneStats`) under
the zones it applies in, as [the Natural flasks](#writing-an-effect-by-hand-the-natural-flasks)
describe.

The settings read these lists as described under [What the settings offer](#what-the-settings-offer).
The sim reads the rows, not these lists.

## Writing an effect by hand: the Natural flasks

Flasks of Natural Accuracy 274273, Aggression 274274, Precision 274275 and Swiftness 274276 cast
1293740-1293743. Each spell's first effect is 60 Stamina; its second is an `A_DUMMY` whose amount is the
bonus the text grants in Mount Hyjal, Hyjal Summit and the Barrow Deeps, and the auras after it state 0
and name what the bonus is (hit, crit, a lower chance to be dodged or parried, haste). The parser does
not attach an `A_DUMMY`, so the rows route as unsupported and no builder registers them.

`ConsumableAllowList` names the four, so db.json ships them although they route as unsupported, and
`sim/common/forever/consumables_natural_flasks.go` registers each by hand, still reading every number
from its row:

```go
var NaturalFlasks = []int32{274273, 274274, 274275, 274276}

func init() {
	for _, itemID := range NaturalFlasks {
		core.NewConsumableEffect(itemID, func(agent core.Agent) {
			character := agent.GetCharacter()
			consumable := core.GetConsumableByID(itemID)
			row := spelldata.MustFind(consumable.OnUse().BuffId)

			config := spelldata.AuraConfig(row, spelldata.Label(consumable.Name))
			config.ActionID = core.ActionID{ItemID: itemID}
			config.BuildPhase = core.CharacterBuildPhaseConsumes
			spelldata.Permanent()(&config)
			aura := character.GetOrRegisterAura(config)

			spelldata.ParseEffects(character, aura, row, spelldata.Effects(1))
			bonus, err := NaturalFlaskZoneBonusOf(row, row.EffectN(2).Average())
			if err != nil {
				panic(err)
			}
			aura.AttachStatsBuff(bonus.Stats)
			if bonus.AttackHastePercent != 0 {
				aura.AttachMultiplyAttackSpeed(1 + bonus.AttackHastePercent/100)
			}
			if bonus.CastHastePercent != 0 {
				aura.AttachMultiplyCastSpeed(1 + bonus.CastHastePercent/100)
			}
		})
	}
}
```

The Stamina is the parser's, from the first effect. `NaturalFlaskZoneBonusOf` reads the dummy's amount
as the later auras name it, as though every fight is in those zones; `A_MOD_EXPERTISE` takes it as
`stats.ExpertisePercent`, since the text reads "5% reduced chance to be dodged or parried". A
hand-written effect for an id a builder does register takes that registration's place;
deleting the hand-written one hands the id back to the rows.

The stat parse lists only the Stamina in db.json. `AddUnparsedStats`
(`tools/database/consumable_unparsed_stats.go`) lists the same `NaturalFlaskZoneBonusOf` in each
flask's `zone_stats` (`zoneStats` in db.json) at the rating one percent takes (Accuracy 50 hit rating,
Aggression 56 crit rating, Precision 50 expertise rating at
`core.ExpertiseRatingPerExpertisePercent` (10), Swiftness 50 haste rating), under the zones of
`naturalFlaskZones`: Mount Hyjal 616 and Hyjal Summit 3606, which only the tooltip names; the Barrow
Deeps has no AreaTable row. The pickers count those stats toward a flask's EP only while the
encounter's zone is one of them. One rating feeds the physical and the spell percent alike, so a flask
that grants both lists the rating once. The sim reads the rows, not those stats, and applies the bonus
in every fight.

## Adding a consumable the rows do not support

The consumable needs a db.json entry and a hand-written effect.

**The game files have it.** Add the id to `ConsumableAllowList`, as with the Natural flasks above. Gift of
Arthas 9088 routes to no shape (its 11371 applies `A_PROC_TRIGGER_SPELL`) and ships because the list
names it:

```go
var ConsumableAllowList = []int32{
	9088, // Gift of Arthas
	// ...
}
```

An item outside the catalogue joins it through the same list. A catalogue consumable that
`ConsumableOverrides` names ships as well, with the override's fields merged onto the game files'.

**The game files do not have it.** Add the whole entry to `ConsumableOverrides`, since nothing else fills
it. Name, type, icon and every number the sim and the UI read come from there:

```go
var ConsumableOverrides = []*proto.Consumable{
	{Id: <item id>, Name: "<name>", Icon: "<icon file name>", Type: proto.ConsumableType_ConsumableTypeOther,
		ItemEffects: []*proto.ItemEffect{{BuffId: <spell id>, Effect: &proto.ItemEffect_OnUse{
			OnUse: &proto.OnUseEffect{CooldownMs: <ms>}}}}},
}
```

Then register its effect with `core.NewConsumableEffect` as above, and regenerate.

## Removing a consumable

Add the id to `ConsumableDenyList` and regenerate. The consumable leaves db.json, and a registration for
it then registers nothing. An id in `ConsumableAllowList` stays, so take it out of that list too.

```go
var ConsumableDenyList = []int32{
	10646, // Goblin Sapper Charge
}
```

## Correcting a number

A consumable's numbers are its spells' rows, so a number the game files do not state goes in
`tools/database/overrides/spell_overrides.go` like any other spell's; the fields and the rule that
retires each are in [Spell Data: Overrides](spell_data/overrides.md). A row names the spell the
consumable casts, not the item:

```go
{30164, DurationMs, <ms>, "<the reason>", "<the source>"},
```

Permanent Lung Juice Cocktail 23715's 30164 is listed unsupported as `30164 states no duration for its
aura`; a `DurationMs` override on it lifts that refusal once the store and db.json are regenerated. An
override is refused when the game files state the field itself. A number the game files state but the sim
should not take is a hand-written effect for the consumable.

## Tests

- `TestEveryConsumableIsRoutedOrListed` (`tools/database`) keeps the unsupported list above, and
  `TestConsumablesRouteByTheirSpell` pins consumables to their shapes.
- `TestEveryShippedConsumableIsRegistered` (`sim/common`, `with_db`) fails with the ids of every db.json
  consumable that has neither a hand-written nor a data-driven registration. A consumable that routes to a shape no generator writes shows up
  here.
- Each builder has unit tests beside it in `sim/common/shared` (`consumable_*_test.go`), and the Natural
  flasks have theirs in `sim/common/forever`.
- The warrior suites select consumables from db.json; a regeneration that moves them shows up in their
  .results files.
- `TestAConsumableTakesItsExclusiveGroup`, `TestTheRowsGroupEveryElixirAndByCategoryOnlyWhereTheCategoryOutlastsTheBuff`
  and `TestEveryExclusiveGroupOverrideShips` (`tools/database`) pin the groups; the stat builder's
  tests which consumables of one group apply and which stack.
- `ui/sim/settings/consumables.test.ts` pins the pickers and the relevance rules below, and
  `ConsumesPicker/utils.test.ts` the settings rows.

## What the settings offer

The settings render every consumable db.json ships from its own fields; nothing in the UI names a
consumable. `buildConsumeConfigs` (`ui/features/settings/components/ConsumesPicker/utils.ts`) puts
each in a picker by its type (`consumableSlot` in `ui/sim/settings/consumables.ts`): `ConsumableTypeConjured`
the conjured item, `ConsumableTypePotion` the potion pickers, then the flask, `ConsumableTypeFood` the
food, the explosive, the imbues and, beside each hand's imbue, `ConsumableTypeWeaponImbue` the weapon
imbue. An ungrouped consumable with a `holiday` goes into the Holiday row. A consumable with an [exclusive group](#exclusive-groups) goes into
its group's picker instead (`makeConsumableGroupInput`), which writes at most one of the group's ids:
into `elixir_ids` for an elixir, into `extra_ids` for anything else (Juju Power beside the Strength
elixirs). The pickers are labelled with `group_name` and ordered by `group_key`: a group holding an
elixir in the Elixirs row after the flask (a scroll of that group among them), one holding a scroll and
no elixir as a column of the Scrolls row, and the rest (Hit food, Zanza) after the food. A group, a column or a row that offers the player nothing hides.

A potion has three pickers, Offensive, Defensive and Resource, and each offers the
potions whose `potion_classes` hold its class. `AddPotionClasses` (`tools/database/consumable_potion_classes.go`) fills
them after `AddUnparsedStats`: a gain of mana, rage or energy the potion builder models is resource, a
heal or absorb defensive, damage to enemies offensive (`shared.ConsumableDamages`: Dragonfire
Potion, Caustic Smog Potion), the stats the potion grants, its mob type stats among them, of the class the
major cooldown manager files the buff under (`core.CooldownTypeOf`: any of `core.SurvivalStats`
defensive, only `core.ManaStats` resource, anything else offensive), and a hit or crit percent
offensive. A potion of several classes, such as Major Rejuvenation Potion or
Mighty Rage Potion, is offered in each, and one of no class fails `gen_db`. A picker keeps offering the
potion already picked, so a pick outside its class can be seen and cleared.
Everything else is picked any number at a time into the spec's `extra_ids`: in the miscellaneous row,
and an explosive with a cooldown of its own beside the shared one (Goblin Sapper Charge 10646) beside
the explosive picker.

The generator writes the type (`Consumable.ToProto` in `tools/database/dbc/consumable.go`) from the
category the first on-use row shares, which is also the consumable's `category_id`: the item row's, or
where it states none, its spell's. An imbue is typed by what it does, and an explosive is what shares
the explosives' category 24 and routes as damage (`IsExplosive` in `tools/database/gen_consumables.go`,
which db.json's type and the explosive registrations both read), so the explosive picker offers nothing
else: Bag of Marbles 1191 shares the category and deals no damage. Category 4 makes a
potion (Potion of Demon Slaying 9224 is an Elixir-subclass item and only its spell 11406 names the
category) and 1153 a conjured item (Demonic Rune 12662 is not even a Consumable-class item); food and
drink categories 11 and 59 make food of what its subclass leaves neither a potion, a flask nor an elixir;
everything else takes its item subclass.

A picker offers a consumable when all of these hold (`offeredTo` in
`ui/features/settings/model/consumables.ts`):

- the class, profession and race the game files allow it (`consumableAllowed`: `class_allowlist`,
  `required_profession` at `req_prof_rank`, `race_allowlist`, the last read from
  `ItemSparse.AllowableRace` by the generator);
- its required level reaches the filter's minimum for its type;
- it does something for the stats the consumables stat filter shows (`isConsumableRelevant`): it
  grants one of them or a percent among them, restores a resource they spend, or, granting none of
  these, has an effect of another kind (damage, an absorb, a weapon enchant). A stat it grants only
  against some mob types or in some zones counts as granted;
- for an imbue, the hand's item fits the enchant's weapon subclass and inventory type masks
  (`handTakesImbue`), so a shield coating shows on a shield.

The stat filter starts from the stats the spec weighs consumables by (`consumableStatsOf`: its
`epStats` and `consumableStats`; Stamina brings health and the resistances, Attack Power flat
physical damage) and the percents `consumablePseudoStatsOf` names, and the player can show or hide
any stat some consumable grants. A consumable already picked stays offered below the minimum level or
with its stats hidden, so it can be seen and cleared; with its stats hidden, it is marked as kept.

Each picker and each row of toggles lists its consumables by the EP their stats and percents add at
the player's EP weights, highest first (`orderedByEp`). A mob type stat counts only while the
encounter's first target is of one of its mob types, and a zone stat only while the encounter's zone
is one of its zones. A consumable that grants nothing countable there (a proc, a restore, damage, a
stat against other mob types) comes after every one that does. Ties go by required level, highest
first, then by name.

`Player.getConsumes(true)` sends the potions and conjured items relevant to the spec's own stats as
the spec's lists, whatever the filter shows (`relevantConsumableIds`). Of the ranks of one name that
end in a numeral, it lists only the highest the player could use, so a top rank the class or the
profession rules out leaves the rank below it. The UI keeps the rows and their text.

`gen_db -gen=db` writes `ui/features/settings/model/consumables_auto_gen.ts` in the same pass as db.json:
one named export per shipped consumable, carrying its item id (`MightyRagePotion = { id: 13442 }`).
Presets name consumables through it, so a preset cannot name one db.json does not ship. A name several
items share takes each item's id (`MajorHealthstone9421`).

## The browser path

`lib.wasm` is built without the database. Each request carries the consumables it needs in the player's
`database`, which `extendPlayerProtoWithMissingEffects` (`ui/sim/proto/proto_migration.ts`) fills from
the UI's db.json: every id field and every id list of the `ConsumesSpec` (the potions, the conjured
items, `elixir_ids` and `extra_ids`). A consumable travels whole, its item effects and temporary
enchant with it.

`TestRequestCarriesEveryConsumableThePresetSelects` (`sim/core`, `with_db`) runs a warrior once with every
consumable loaded and once with only those a request carries, and checks both apply the same.
