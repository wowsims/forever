package mage

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// One of each conjured gem, best first, one use each per fight. Each is shown as its item. The
// conjure spells' E_CREATE_ITEM states item 0 (Conjure Mana Ruby, 10054), so the item ids are stated
// here; ItemXItemEffect links each item to its Replenish Mana spell.
var manaGems = []struct {
	itemID int32
	name   string
	rank   *spelldata.Spell
}{
	{8008, "Mana Ruby", spellData.ConjureManaRubyTriggered.Highest()},
	{8007, "Mana Citrine", spellData.ConjureManaCitrineTriggered.Highest()},
	{5513, "Mana Jade", spellData.ConjureManaJadeTriggered.Highest()},
	{5514, "Mana Agate", spellData.ConjureManaAgateTriggered.Highest()},
}

// The cooldown the gems share: their items' on-use category, which healthstones and the other
// conjured consumables share too, not the Replenish Mana rows' 60 sec category 30. Every gem's item
// states the same category, so the first gem the consumables catalogue carries sets it for all four;
// Mana Agate is not in the catalogue.
func (mage *Mage) manaGemCooldown() core.Cooldown {
	for _, gem := range manaGems {
		if onUses := core.GetConsumableByID(gem.itemID).OnUseEffects(); len(onUses) > 0 {
			onUse := onUses[0].GetOnUse()
			return core.Cooldown{
				Timer:    mage.GetOrInitSpellCategoryTimer(onUse.CategoryId),
				Duration: core.DurationFromMillis(onUse.CategoryCooldownMs),
			}
		}
	}
	// A browser request carries only the consumables the player picked, so it may hold no gem. Every
	// gem's ItemEffect row states category 1153 for 120 000 ms.
	return core.Cooldown{Timer: mage.GetCombatConsumableCD(), Duration: 2 * time.Minute}
}

func (mage *Mage) registerManaGems() {
	cooldown := mage.manaGemCooldown()
	// The charges of the gems already registered, which are the better ones.
	var betterCharges []*core.Aura
	for _, gem := range manaGems {
		gemRank := gem.rank
		actionID := core.ActionID{ItemID: gem.itemID}
		manaMetrics := mage.NewManaMetrics(actionID)
		manaEffect := gemRank.EnergizeEffect()

		charge := core.MakePermanent(mage.GetOrRegisterAura(core.Aura{
			Label:     gem.name,
			ActionID:  actionID,
			MaxStacks: 1,
		})).ApplyOnReset(func(aura *core.Aura, sim *core.Simulation) {
			aura.SetStacks(sim, 1)
		})

		config := spelldata.SpellConfig(&mage.Unit, gemRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagHelpful))
		config.ActionID = actionID
		config.Rank = 0
		config.ClassSpellMask = MageSpellManaGem
		config.Cast.DefaultCast.NonEmpty = true
		config.Cast.CD = cooldown

		config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
			return charge.GetStacks() > 0
		}

		config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			charge.RemoveStack(sim)
			mage.AddMana(sim, manaEffect.Roll(sim), manaMetrics)
		}

		// Hold this gem until every better one has been used.
		higherGems := betterCharges
		betterCharges = append(betterCharges, charge)

		mage.AddMajorCooldown(core.MajorCooldown{
			Spell: mage.RegisterSpell(config),
			Type:  core.CooldownTypeMana,
			ShouldActivate: func(_ *core.Simulation, character *core.Character) bool {
				for _, higher := range higherGems {
					if higher.GetStacks() > 0 {
						return false
					}
				}
				return character.CurrentMana()+manaEffect.Max() <= character.MaxMana()
			},
		})
	}
}
