//go:build with_db

// Only include this file in the build when we specify the 'with_db' tag.
// Without the tag, the database will start out completely empty.
package core

import (
	"github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	db := database.Load()
	WITH_DB = true

	simDB := &proto.SimDatabase{
		Items:                    make([]*proto.SimItem, len(db.Items)),
		Enchants:                 make([]*proto.SimEnchant, len(db.Enchants)),
		ItemEffectRandPropPoints: make([]*proto.ItemEffectRandPropPoints, len(db.ItemEffectRandPropPoints)),
		RandomSuffixes:           make([]*proto.ItemRandomSuffix, len(db.RandomSuffixes)),
		Consumables:              make([]*proto.Consumable, len(db.Consumables)),
		SpellEffects:             make([]*proto.SpellEffect, len(db.SpellEffects)),
	}

	for i, item := range db.Items {
		simDB.Items[i] = &proto.SimItem{
			Id:               item.Id,
			Name:             item.Name,
			Type:             item.Type,
			ArmorType:        item.ArmorType,
			WeaponType:       item.WeaponType,
			HandType:         item.HandType,
			RangedWeaponType: item.RangedWeaponType,
			PseudoStats:      item.PseudoStats,
			WeaponSpeed:      item.WeaponSpeed,
			QualityModifier:  item.QualityModifier,
			SpellSchool:      item.SpellSchool,
			Unique:           item.Unique,
			LimitCategory:    item.LimitCategory,
			SetName:          item.SetName,
			SetId:            item.SetId,
			ClassAllowlist:   item.ClassAllowlist,
			ScalingOptions:   item.ScalingOptions,
		}
	}

	for i, suffix := range db.RandomSuffixes {
		simDB.RandomSuffixes[i] = &proto.ItemRandomSuffix{
			Id:          suffix.Id,
			Name:        suffix.Name,
			Stats:       suffix.Stats,
			PseudoStats: suffix.PseudoStats,
		}
	}

	for i, enchant := range db.Enchants {
		simDB.Enchants[i] = &proto.SimEnchant{
			EffectId:     enchant.EffectId,
			Stats:        enchant.Stats,
			PseudoStats:  enchant.PseudoStats,
			WeaponDamage: enchant.WeaponDamage,
			Name:         enchant.Name,
			Type:         enchant.Type,
			EnchantType:  enchant.EnchantType,
			ExtraTypes:   enchant.ExtraTypes,
			ProcKind:     enchant.ProcKind,
		}
	}

	for i, itemEffectRpp := range db.ItemEffectRandPropPoints {
		simDB.ItemEffectRandPropPoints[i] = &proto.ItemEffectRandPropPoints{
			Ilvl:           itemEffectRpp.Ilvl,
			RandPropPoints: itemEffectRpp.RandPropPoints,
		}
	}

	for i, consumable := range db.Consumables {
		simDB.Consumables[i] = consumable
	}

	for i, effect := range db.SpellEffects {
		simDB.SpellEffects[i] = effect
	}

	addToDatabase(simDB)
}
