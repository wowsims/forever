package shared

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var shootRow = spelldata.MustFind(5019)

// Shoot is a channel during which the wand fires one shot per wand speed, until something takes a shot's
// slot. A shot rolls the wand's damage in its school: no attack power, no spell power. It misses on the
// ranged weapon table, which wand skill sets and physical and ranged hit lower, so spell hit leaves it
// alone. It has no dodge, parry or block, and crits on the caster's spell crit, for 1.5 times like a spell.
func RegisterShoot(character *core.Character) {
	weapon := character.WeaponFromRanged()
	wandSpeed := func() time.Duration {
		if wand := character.GetRangedWeapon(); wand != nil {
			return core.DurationFromSeconds(max(wand.SwingSpeed, 1))
		}
		return 0
	}

	// Each shot is a cast of its own, which carries the shots' damage, hits and crits; Shoot's casts
	// count the channels.
	shotConfig := spelldata.SpellConfig(&character.Unit, shootRow, spelldata.Magic(core.ProcMaskRangedAuto), spelldata.Proc(), spelldata.Tag(1))
	shotConfig.Flags |= core.SpellFlagNoPartialResist
	shotConfig.SpellSchool = weapon.GetSpellSchool()
	shotConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := weapon.CalculateWeaponDamage(sim, 0)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndMagicCrit)
		spell.DealDamageAfterTravel(sim, result)
	}
	shot := character.RegisterSpell(shotConfig)

	config := spelldata.SpellConfig(&character.Unit, shootRow, spelldata.Magic(core.ProcMaskRangedAuto))
	config.Flags |= core.SpellFlagChanneled | core.SpellFlagAPL | core.SpellFlagNoPartialResist
	config.SpellSchool = weapon.GetSpellSchool()
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return !character.Moving
	}
	config.Dot = core.WandChannel()
	config.ApplyEffects = func(_ *core.Simulation, _ *core.Unit, spell *core.Spell) {
		weapon = character.WeaponFromRanged()
		spell.SetSpellSchool(weapon.GetSpellSchool())
		shot.SetSpellSchool(weapon.GetSpellSchool())
	}

	shoot := character.RegisterSpell(config)
	character.EnableWand(shoot, wandSpeed, func(sim *core.Simulation, target *core.Unit) {
		shot.Cast(sim, target)
	})
}
