package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerInnervateCD() {
	rank := spellData.Innervate.Highest()
	target := druid.GetUnit(druid.SelfBuffs.InnervateTarget)
	if target == nil {
		target = &druid.Unit
	}
	var targetChar *core.Character
	if player := druid.Env.Raid.GetPlayerFromUnit(target); player != nil {
		targetChar = player.GetCharacter()
	}
	aura := buffs.NewSpellDataInnervatesAura(target, true, 0)
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL | core.SpellFlagHelpful
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return !target.HasActiveAuraWithTag(buffs.InnervatesCategory)
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}
	druid.Innervate = druid.RegisterSpell(config)
	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: druid.Innervate.Spell,
		Type:  core.CooldownTypeMana,
		// As early as the mana fits, so that the cooldown comes back as often as the fight allows: once
		// half the gain fits under the target's bar while another Innervate would still fit in the fight
		// after this one. The last one waits for three quarters, for the target's other mana cooldowns,
		// whose cooldowns are shorter, and, while a mana potion or rune will be ready again before the
		// end, for the bar to drop below 30%, since that consumable covers part of the need. None while
		// the bar would last to the end at the rate it has emptied since the pull, as in a short fight,
		// where the cast would only cost a GCD. The shares, the waits and the 30% were measured on
		// Balance against casting below 10% mana.
		ShouldActivate: func(sim *core.Simulation, _ *core.Character) bool {
			if !target.HasManaBar() || sim.CurrentTime <= 0 {
				return false
			}
			missing := target.MaxMana() - target.CurrentMana()
			anotherFits := sim.GetRemainingDuration() > druid.Innervate.CD.Duration+aura.Duration
			share := 0.75
			if anotherFits {
				share = 0.5
			}
			if gain := buffs.InnervateMana(target, aura); gain <= 0 || missing < share*gain {
				return false
			}
			if !anotherFits && targetChar != nil {
				end := sim.CurrentTime + sim.GetRemainingDuration()
				for _, mcd := range targetChar.GetMajorCooldowns() {
					if mcd.Spell == druid.Innervate.Spell || !mcd.Type.Matches(core.CooldownTypeMana) {
						continue
					}
					if mcd.IsReady(sim) {
						return false
					}
					if isConsumable(mcd) && mcd.ReadyAt() < end && target.CurrentMana() >= 0.3*target.MaxMana() {
						return false
					}
				}
			}
			// The net drain since the pull, which the Time to OOM metric projects the same way.
			drain := target.Metrics.NetManaSpent() / sim.CurrentTime.Seconds()
			return target.CurrentMana() < drain*sim.GetRemainingDuration().Seconds()
		},
	})
}

// Whether the cooldown uses a consumable, such as a potion or a rune, rather than an equipped item or
// a spell.
func isConsumable(mcd *core.MajorCooldown) bool {
	id := mcd.Spell.ActionID.ItemID
	return id != 0 && core.GetConsumableByID(id).Id != 0
}
