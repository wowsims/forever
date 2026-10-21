package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The poison items whose weapon imbue casts a rank of Deadly or Instant Poison. Each is the item's
// consumable effect, in place of the generic imbue builder, which would apply Deadly Poison's
// stacking damage over time as one.
var (
	deadlyPoisonItemIDs  = []int32{2892, 2893, 8984, 8985, 20844}
	instantPoisonItemIDs = []int32{6947, 6949, 6950, 8926, 8927, 8928}
)

func init() {
	for _, itemID := range deadlyPoisonItemIDs {
		core.NewConsumableEffect(itemID, func(agent core.Agent) {
			if rogueAgent, ok := agent.(RogueAgent); ok {
				rogueAgent.GetRogue().applyDeadlyPoison(core.GetConsumableByID(itemID))
			}
		})
	}
	for _, itemID := range instantPoisonItemIDs {
		core.NewConsumableEffect(itemID, func(agent core.Agent) {
			if rogueAgent, ok := agent.(RogueAgent); ok {
				rogueAgent.GetRogue().applyInstantPoison(core.GetConsumableByID(itemID))
			}
		})
	}
}

// Whether a hand of the main set or of the item swap set carries the consumable's imbue.
func (rogue *Rogue) carriesImbue(consumable core.Consumable) bool {
	if consumable.TemporaryEnchant == nil {
		return false
	}
	enchantID := consumable.TemporaryEnchant.EffectID
	for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand} {
		if rogue.Equipment[slot].HasTempEnchant(enchantID) || rogue.ItemSwap.GetUnequippedItemBySlot(slot).HasTempEnchant(enchantID) {
			return true
		}
	}
	return false
}

// The rank of the ladder the consumable's imbue casts, 0 where it casts none of them.
func poisonRank(ladder spelldata.Ladder, consumable core.Consumable) int32 {
	var rank int32
	if consumable.TemporaryEnchant == nil {
		return rank
	}
	for _, effect := range consumable.TemporaryEnchant.EnchantEffects {
		ladder.Each(func(n int32, s *spelldata.Spell) {
			if s.ID == effect.BuffId {
				rank = n
			}
		})
	}
	return rank
}

// The highest rank of the ladder that any imbue the rogue carries casts. Both hands and both weapon
// sets share the one poison spell, so a lower rank on another weapon procs at this one.
func (rogue *Rogue) highestPoisonRank(ladder spelldata.Ladder) *spelldata.Spell {
	var highest int32
	for _, id := range rogue.SelectedConsumableIDs() {
		consumable := core.GetConsumableByID(id)
		if rank := poisonRank(ladder, consumable); rank > highest && rogue.carriesImbue(consumable) {
			highest = rank
		}
	}
	return ladder.Rank(highest)
}

// Casts the poison off the melee hits of the weapons carrying the consumable's imbue, at their own
// speeds, following the weapons through an item swap.
func (rogue *Rogue) registerPoisonProc(consumable core.Consumable, chance float64, poison *core.Spell) {
	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:                consumable.Name,
		ActionID:            core.ActionID{ItemID: consumable.Id},
		Outcome:             core.OutcomeLanded,
		Callback:            core.CallbackOnSpellHitDealt,
		TriggerImmediately:  true,
		ProcMask:            core.ProcMaskMelee,
		IsWeaponProc:        true,
		ProcSpellClassFlags: poison.ClassFlags,
		DPM: rogue.NewDynamicLegacyProcForTempEnchant(consumable.TemporaryEnchant.EffectID, 0, func(core.ProcMask) float64 {
			return chance
		}),
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			poison.Cast(sim, result.Target)
		},
	})
}

func (rogue *Rogue) registerDeadlyPoisonSpell(deadlyPoisonRank *spelldata.Spell) {
	config := spelldata.SpellConfig(&rogue.Unit, deadlyPoisonRank, spelldata.Magic(core.ProcMaskSpellDamageProc))
	config.ClassSpellMask = RogueSpellDeadlyPoison
	tick := deadlyPoisonRank.PeriodicEffect()
	numTicks := deadlyPoisonRank.Duration().Seconds() / tick.Period().Seconds()
	apScaling := 0.0045 / numTicks // Confirmed by Zirene

	config.Dot = spelldata.DotConfig(deadlyPoisonRank, tick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		stacks := dot.GetStacks()
		baseTick := tick.Average()
		apFactor := dot.Spell.MeleeAttackPower(target) * apScaling
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, (baseTick+apFactor)*float64(stacks), deadlyPoisonRank.TickOutcome(dot))
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
		if !result.Landed() {
			return
		}

		dot := spell.Dot(target)
		if dot.IsActive() {
			dot.Refresh(sim)
			dot.AddStack(sim)
		} else {
			dot.Apply(sim)
			dot.SetStacks(sim, 1)
		}
	}

	rogue.DeadlyPoison = rogue.RegisterSpell(config)
}

func (rogue *Rogue) registerInstantPoisonSpell(instantPoisonRank *spelldata.Spell) {
	ipBaseDamage := instantPoisonRank.DamageEffect().Average()
	config := spelldata.SpellConfig(&rogue.Unit, instantPoisonRank, spelldata.Magic(core.ProcMaskSpellDamageProc))
	config.ClassSpellMask = RogueSpellInstantPoison
	apScaling := 0.005 // Confirmed by Zirene

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, ipBaseDamage+apScaling*spell.MeleeAttackPower(target), spell.OutcomeMagicHitAndCrit)
	}

	rogue.InstantPoison = rogue.RegisterSpell(config)
}

func (rogue *Rogue) applyDeadlyPoison(consumable core.Consumable) {
	if !rogue.carriesImbue(consumable) {
		return
	}
	if rogue.DeadlyPoison == nil {
		rogue.registerDeadlyPoisonSpell(rogue.highestPoisonRank(spellData.DeadlyPoison))
	}
	rogue.registerPoisonProc(consumable, 0.3, rogue.DeadlyPoison)
}

func (rogue *Rogue) applyInstantPoison(consumable core.Consumable) {
	if !rogue.carriesImbue(consumable) {
		return
	}
	if rogue.InstantPoison == nil {
		rogue.registerInstantPoisonSpell(rogue.highestPoisonRank(spellData.InstantPoison))
	}
	rogue.registerPoisonProc(consumable, 0.2, rogue.InstantPoison) //Confirmed in game 9/30/2026
}
