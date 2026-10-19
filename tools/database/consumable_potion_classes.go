package database

import (
	"fmt"
	"slices"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Lists the classes of each potion, which the settings offer it under and the APL potion placeholders
// cast it by. A gain of mana, rage or energy the potion builder models is a resource potion, a heal or
// absorb a defensive one, damage to enemies an offensive one, and the stats it grants of the class the
// major cooldown manager files the buff under: any armor, avoidance, health or resistance defensive,
// mana regen alone resource, anything else offensive. A potion with no class fails. Runs after
// AddUnparsedStats, so the stats against creature types, listed as mob type stats, and the hit and
// crit percents count.
func AddPotionClasses(db *WowDatabase) error {
	for id, consumable := range db.Consumables {
		consumable.PotionClasses = nil
		if consumable.Type != proto.ConsumableType_ConsumableTypePotion || len(consumable.ItemEffects) == 0 {
			continue
		}

		var classes []proto.PotionClass
		add := func(class proto.PotionClass) {
			if !slices.Contains(classes, class) {
				classes = append(classes, class)
			}
		}

		row := spelldata.Find(consumable.ItemEffects[0].BuffId)
		power, survival := shared.ConsumableRestores(row)
		if power {
			add(proto.PotionClass_PotionClassResource)
		}
		if survival {
			add(proto.PotionClass_PotionClassDefensive)
		}
		if shared.ConsumableDamages(row) {
			add(proto.PotionClass_PotionClassOffensive)
		}

		granted := stats.FromProtoArray(consumable.Stats)
		for _, versus := range consumable.MobTypeStats {
			granted = granted.Add(stats.FromProtoMap(versus.Stats))
		}
		if buffed := granted.GetBuffedStatTypes(); len(buffed) > 0 {
			switch core.CooldownTypeOf(buffed) {
			case core.CooldownTypeSurvival:
				add(proto.PotionClass_PotionClassDefensive)
			case core.CooldownTypeMana:
				add(proto.PotionClass_PotionClassResource)
			default:
				add(proto.PotionClass_PotionClassOffensive)
			}
		}
		if slices.ContainsFunc(consumable.PseudoStats, func(v float64) bool { return v != 0 }) {
			add(proto.PotionClass_PotionClassOffensive)
		}

		if len(classes) == 0 {
			return fmt.Errorf("potion %d (%s): its on-use spell %d gives it no class", id, consumable.Name, consumable.ItemEffects[0].BuffId)
		}
		slices.Sort(classes)
		consumable.PotionClasses = classes
	}
	return nil
}
