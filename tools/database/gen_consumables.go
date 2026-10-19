package database

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
)

// The shape a consumable's on-use spell is built in. Damage and heals are routed the way an item's
// on-use is, and have builders; the other shapes name the builder a consumable of that kind needs.
type ConsumableKind string

const (
	ConsumableDamage           ConsumableKind = "damage"
	ConsumableHeal             ConsumableKind = "heal"
	ConsumableEnergize         ConsumableKind = "energize"
	ConsumableStatAura         ConsumableKind = "stat aura"
	ConsumablePercentStat      ConsumableKind = "percent stat"
	ConsumableTemporaryEnchant ConsumableKind = "temporary enchant"
	ConsumableAbsorb           ConsumableKind = "absorb"
)

var ConsumableKinds = []ConsumableKind{
	ConsumableDamage, ConsumableHeal, ConsumableEnergize, ConsumableStatAura, ConsumablePercentStat,
	ConsumableTemporaryEnchant, ConsumableAbsorb,
}

type ConsumableRouting struct {
	ItemID  int
	SpellID int32
	// Empty when no shape fits the spell.
	Kind        ConsumableKind
	Unsupported []string
}

func (r *ConsumableRouting) Supported() bool {
	return r.Kind != "" && len(r.Unsupported) == 0
}

func (r *ConsumableRouting) Reason() string {
	return strings.Join(r.Unsupported, "; ")
}

// The consumables the catalogue takes by id, whatever ConsumableItemFilter says: those
// ConsumableAllowList names and those an exclusive group names, such as the Quest-class Jujus.
func catalogueConsumableIDs() []int32 {
	ids := slices.Clone(ConsumableAllowList)
	for _, group := range ConsumableExclusiveGroups {
		ids = append(ids, group.Items...)
	}
	return ids
}

// Whether the catalogue's consumable goes into db.json: its spell routes to a shape, or
// ConsumableAllowList, ClassBuiltImbues or ConsumableOverrides names it. KeepConsumable runs after.
func ShipsConsumable(routing *ConsumableRouting) bool {
	id := int32(routing.ItemID)
	_, classBuilt := ClassBuiltImbues[id]
	return routing.Supported() || classBuilt || slices.Contains(ConsumableAllowList, id) ||
		slices.ContainsFunc(ConsumableOverrides, func(c *proto.Consumable) bool { return c.Id == id })
}

// The catalogue's consumable as db.json carries it, typed an explosive where IsExplosive says so, with
// its exclusive group.
func ShippedConsumable(c *dbc.Consumable, routing *ConsumableRouting, instance *dbc.DBC) *proto.Consumable {
	consumable := c.ToProto()
	if IsExplosive(c, routing) {
		consumable.Type = proto.ConsumableType_ConsumableTypeExplosive
	}
	setConsumableGroup(c, consumable, routing, instance)
	consumable.Holiday = consumableHoliday(c)
	return consumable
}

// The world event the consumable belongs to: HolidayConsumables names it, or its item's RequiredHoliday.
func consumableHoliday(c *dbc.Consumable) string {
	if holiday, ok := HolidayConsumables[int32(c.Id)]; ok {
		return holiday
	}
	if c.RequiredHoliday == 0 {
		return ""
	}
	if holiday, ok := HolidayNames[c.RequiredHoliday]; ok {
		return holiday
	}
	return fmt.Sprintf("Holiday %d", c.RequiredHoliday)
}

// Whether the catalogue's consumable is an explosive, which the explosive picker offers and the
// explosives builder runs: its on-use shares the explosives' category and its spell routes as damage.
// Bag of Marbles 1191 shares the category and deals no damage.
func IsExplosive(c *dbc.Consumable, routing *ConsumableRouting) bool {
	return routing.Kind == ConsumableDamage && c.SharesExplosivesCategory()
}

// Whether a consumable in db.json survives the global filters. ConsumableAllowList keeps it,
// ConsumableDenyList drops it, and so does a test or QA name.
func KeepConsumable(consumable *proto.Consumable) bool {
	if slices.Contains(ConsumableAllowList, consumable.Id) {
		return true
	}
	if slices.Contains(ConsumableDenyList, consumable.Id) {
		return false
	}
	for _, pattern := range DenyListNameRegexes {
		if pattern.MatchString(consumable.Name) {
			return false
		}
	}

	return !strings.HasPrefix(consumable.Name, "QA") && !strings.HasPrefix(consumable.Name, "Test") && !strings.HasPrefix(consumable.Name, "TEST") && !strings.Contains(consumable.Name, "Flaskataur")
}

// Routes the spell the consumable's first on-use row casts: as damage, a heal or an absorb where the
// on-use routing takes it, as damage where its missile leaves a cloud that deals it
// (shared.DamageCloudOf), and otherwise by the first of its effects, or of the spells it triggers on
// the user, that fits a consumable shape. A heal, an absorb or a gain of a resource is supported only
// where the potion builder runs the spell.
func RouteConsumable(c *dbc.Consumable, instance *dbc.DBC) *ConsumableRouting {
	routing := &ConsumableRouting{ItemID: c.Id}
	onUses := c.OnUseEffects()
	if len(onUses) == 0 {
		routing.Unsupported = []string{"the item states no on-use spell"}
		return routing
	}
	routing.SpellID = onUses[0].BuffId

	onUse := routeOnUseSpell(c.Id, routing.SpellID, instance)
	switch {
	case onUse.Supported() && onUse.Shape == ShapeDamage:
		routing.Kind = ConsumableDamage
		routing.Unsupported = shared.DamageConsumableUnsupported(spelldata.Find(routing.SpellID))
		return routing
	case onUse.Supported() && onUse.Shape == ShapeHeal:
		routing.Kind = ConsumableHeal
		routing.Unsupported = shared.ConsumableUnsupported(spelldata.Find(routing.SpellID))
		return routing
	case onUse.Shape == ShapeAbsorb:
		routing.Kind = ConsumableAbsorb
		routing.Unsupported = append(onUse.Unsupported, shared.ConsumableUnsupported(spelldata.Find(routing.SpellID))...)
		return routing
	}
	if cloud, _ := shared.DamageCloudOf(spelldata.Find(routing.SpellID)); cloud != spelldata.Nil {
		routing.Kind = ConsumableDamage
		routing.Unsupported = shared.DamageConsumableUnsupported(spelldata.Find(routing.SpellID))
		return routing
	}

	routing.Kind, routing.Unsupported = consumableShape(c, spelldata.Find(routing.SpellID))
	if routing.Kind == "" {
		routing.Unsupported = onUse.Unsupported
		if onUse.Supported() {
			routing.Unsupported = []string{fmt.Sprintf("the on-use routes to %s, which no consumable shape builds", onUse.OnUseConstructor())}
		}
	}
	return routing
}

func consumableShape(c *dbc.Consumable, s *spelldata.Spell) (ConsumableKind, []string) {
	var seated *spelldata.Effect
	for _, e := range consumableEffects(s) {
		switch {
		case dbc.IsTemporaryEnchantEffect(e.Type):
			return ConsumableTemporaryEnchant, temporaryEnchantUnsupported(c.TemporaryEnchant)
		case !shared.LandsOnTheUser(e):
		case e.Type == dbcenums.E_ENERGIZE, e.Type == dbcenums.E_ENERGIZE_PCT,
			e.Aura == dbcenums.A_PERIODIC_ENERGIZE, e.Aura == dbcenums.A_OBS_MOD_POWER:
			return ConsumableEnergize, shared.ConsumableUnsupported(s)
		case e.Aura == dbcenums.A_SCHOOL_ABSORB:
			return ConsumableAbsorb, []string{fmt.Sprintf("%d absorbs damage, which no builder takes from the rows", e.SpellID)}
		case e.Aura == dbcenums.A_MOD_PERCENT_STAT, e.Aura == dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE:
			return ConsumablePercentStat, shared.AuraDurationUnsupported(e)
		case e.ParsedAura() && spelldata.Find(e.SpellID).WhileSeated():
			seated = cmp.Or(seated, e)
		case e.ParsedAura():
			return ConsumableStatAura, shared.StatConsumableUnsupported(s)
		}
	}
	if seated != nil {
		return ConsumableStatAura, append([]string{fmt.Sprintf("%d restores only while the user sits, out of combat (%v)",
			seated.SpellID, seated.Aura)}, shared.StatConsumableUnsupported(s)...)
	}
	return "", nil
}

// The effects of the consumable's spell and of each spell it triggers on the user.
func consumableEffects(s *spelldata.Spell) []*spelldata.Effect {
	var effects []*spelldata.Effect
	for _, spell := range shared.ConsumableSpells(s) {
		for i := range spell.Effects {
			effects = append(effects, &spell.Effects[i])
		}
	}
	return effects
}

// A temporary enchant's slots, by what the sim's enchant handling reads: flat weapon damage, a
// resistance or a stat, a combat spell that deals damage at the chance the slot states, or an equip
// spell whose auras the parser attaches or whose proc the imbue builder takes from its row. Spells
// that work only in one area do nothing in the sim.
func temporaryEnchantUnsupported(enchant dbc.TemporaryEnchant) []string {
	if enchant.ID == 0 {
		return []string{"the temporary enchant names no SpellItemEnchantment row"}
	}

	var unsupported []string
	for slot, effect := range enchant.Effects {
		spellID := int32(enchant.Args[slot])
		switch effect {
		case dbc.ITEM_ENCHANTMENT_NONE, dbc.ITEM_ENCHANTMENT_DAMAGE, dbc.ITEM_ENCHANTMENT_RESISTANCE, dbc.ITEM_ENCHANTMENT_STAT:
		case dbc.ITEM_ENCHANTMENT_COMBAT_SPELL:
			if !dealsDamageToAnEnemy(spelldata.Find(spellID)) {
				unsupported = append(unsupported, fmt.Sprintf("enchant %d casts %d, which deals no damage to an enemy", enchant.ID, spellID))
			}
			if enchant.Points[slot] <= 0 {
				unsupported = append(unsupported, fmt.Sprintf("enchant %d states no chance for its combat spell %d", enchant.ID, spellID))
			}
			if s := spelldata.Find(spellID); s.PeriodicDamageEffect() != spelldata.NilEffect && s.MaxStack > 0 {
				unsupported = append(unsupported, fmt.Sprintf("enchant %d casts %d, whose damage over time stacks to %d, which the proc builder applies as one", enchant.ID, spellID, s.MaxStack))
			}
		case dbc.ITEM_ENCHANTMENT_EQUIP_SPELL:
			unsupported = append(unsupported, equipSpellUnsupported(enchant.ID, spelldata.Find(spellID), spellID)...)
		default:
			unsupported = append(unsupported, fmt.Sprintf("enchant %d has a slot of type %d", enchant.ID, effect))
		}
	}
	if enchant.RequiredAreasID != 0 {
		unsupported = append(unsupported, fmt.Sprintf("enchant %d's spells work only in area group %d", enchant.ID, enchant.RequiredAreasID))
	}
	return unsupported
}

func dealsDamageToAnEnemy(s *spelldata.Spell) bool {
	for _, e := range []*spelldata.Effect{s.DamageEffect(), s.PeriodicDamageEffect()} {
		if e != spelldata.NilEffect && e.HitsAnEnemy() {
			return true
		}
	}
	return false
}

func equipSpellUnsupported(enchantID int, s *spelldata.Spell, spellID int32) []string {
	if s == spelldata.Nil {
		return []string{fmt.Sprintf("enchant %d's equip spell %d has no row in the store", enchantID, spellID)}
	}
	var unsupported []string
	if shared.TemporaryEnchantEquipSpellProcs(s) {
		for _, reason := range append(spelldata.EnchantAuraUnsupported(s), shared.FiredSpellUnsupported(s)...) {
			unsupported = append(unsupported, fmt.Sprintf("enchant %d's equip spell %d: %s", enchantID, spellID, reason))
		}
		return unsupported
	}
	for i := range s.Effects {
		if e := &s.Effects[i]; e.Aura != 0 && !e.ParsedAura() {
			unsupported = append(unsupported, fmt.Sprintf("enchant %d's equip spell %d applies %v, which the parser does not attach", enchantID, spellID, e.Aura))
		}
	}
	return unsupported
}

// Puts every consumable db.json ships into the item icon table, which is where the UI looks a
// consumable's icon up: the game files' icon for an item the catalogue or the item table carries, and
// the icon an override states for one the game files do not ship.
func AddConsumableIcons(db *WowDatabase, catalogue []dbc.Consumable, instance *dbc.DBC, artPaths map[int]string) {
	fileDataIDs := map[int32]int{}
	for _, consumable := range catalogue {
		fileDataIDs[int32(consumable.Id)] = consumable.IconFileDataID
	}

	for id, consumable := range db.Consumables {
		icon := consumable.Icon
		if fdid, ok := fileDataIDs[id]; ok {
			icon = GetIconName(artPaths, fdid)
		} else if item, ok := instance.Items[int(id)]; ok {
			icon = GetIconName(artPaths, item.FDID)
		}
		db.AddItemIcon(id, strings.ToLower(icon), consumable.Name)
	}
}
