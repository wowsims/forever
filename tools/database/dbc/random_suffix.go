package dbc

import (
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

type RandomSuffix struct {
	ID            int
	Name          string
	AllocationPct []int // AllocationPct_0-4
	EffectArgs    []int // EffectArg_0-4
	Effects       []int // Effect_0-4
}

func (raw RandomSuffix) ToProto() *proto.ItemRandomSuffix {
	suffix := &proto.ItemRandomSuffix{
		Name:  raw.Name,
		Id:    int32(raw.ID),
		Stats: stats.Stats{}.ToProtoArray(),
	}

	for i, effect := range raw.Effects {
		amount := float64(raw.AllocationPct[i])
		switch effect {
		case ITEM_ENCHANTMENT_RESISTANCE:
			stat, match := MapResistanceToStat(raw.EffectArgs[i])
			if !match {
				continue
			}

			suffix.Stats[stat] = amount
			if suffix.Name == "" {
				suffix.Name = stats.Stat(stat).StatName()
			}
		case ITEM_ENCHANTMENT_STAT:
			mapped, match := MapBonusStatIndexToStats(raw.EffectArgs[i])
			if !match {
				continue
			}
			for _, stat := range mapped {
				suffix.Stats[stat] = amount
				if suffix.Name == "" {
					suffix.Name = stats.Stat(stat).StatName()
				}
			}
		case ITEM_ENCHANTMENT_EQUIP_SPELL: //Buff
			equipStats := stats.Stats{}
			for _, spellEffect := range dbcInstance.SpellEffects[raw.EffectArgs[i]] {
				addEquipEffectStat(&equipStats, spellEffect, amount)
			}
			suffix.Stats = stats.FromProtoArray(suffix.Stats).Add(equipStats).ToProtoArray()
		}
	}
	return suffix
}
