package paladin

import (
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// The APL picker's submenus for the paladin kit, rotation_tab.apl.submenus keys joined by dots.
const (
	menuSpellSeals          = "spell.seals"
	menuSpellDamage         = "spell.damage"
	menuSpellHeals          = "spell.heals"
	menuSpellTank           = "spell.tank"
	menuSpellPaladinAuras   = "spell.paladin_auras"
	menuPaladin             = "paladin"
	menuPaladinSeals        = "paladin.seals"
	menuPaladinHeals        = "paladin.heals"
	menuPaladinTank         = "paladin.tank"
	menuPaladinPaladinAuras = "paladin.paladin_auras"
)

// Each castable spell's group by its ClassSpellMask, and the group its buff on the paladin goes to.
// Holy Shock is two castable spells: the damage files under Damage and the heal under Heals. Divine
// Favor and Templar's Bulwark are major cooldowns and stay under Cooldowns with their auras.
var menuGroups = []struct {
	mask        int64
	spell, aura string
}{
	{SpellMaskAllSeals | SpellMaskJudgement | SpellMaskSwiftJudgement, menuSpellSeals, menuPaladinSeals},
	{SpellMaskHolyStrike | SpellMaskConsecration | SpellMaskExorcism | SpellMaskHammerOfWrath | SpellMaskHolyWrath | SpellMaskHolyShock, menuSpellDamage, menuPaladin},
	{SpellMaskHolyLight | SpellMaskFlashOfLight | SpellMaskLayOnHands | SpellMaskHolyShockHeal | SpellMaskLightsVigil | SpellMaskOathOfLight, menuSpellHeals, menuPaladinHeals},
	{SpellMaskRighteousFury | SpellMaskHolyShield, menuSpellTank, menuPaladinTank},
	{SpellMaskDevotionAura | SpellMaskRetributionAura | SpellMaskConcentrationAura | SpellMaskFireResistanceAura | SpellMaskFrostResistanceAura | SpellMaskShadowResistanceAura, menuSpellPaladinAuras, menuPaladinPaladinAuras},
}

// Files the castable spells, every rank, and what they put on the paladin: the RelatedSelfBuff, the
// aura of an area dot (Consecration), and the paladin's own entry of an ally aura array (Light's
// Vigil, Oath of Light). The auras no spell names file themselves where they register. Each spec lists
// its own groups first.
func (paladin *Paladin) registerMenuGroups() {
	paladin.OnSpellRegistered(func(spell *core.Spell) {
		for _, group := range menuGroups {
			if !spell.Matches(group.mask) {
				continue
			}
			spell.MenuGroup = group.spell
			if spell.RelatedSelfBuff != nil {
				spell.RelatedSelfBuff.MenuGroup = group.aura
			}
			if dot := spell.AOEDot(); dot != nil {
				dot.MenuGroup = group.aura
			}
			for _, auras := range spell.RelatedAuraArrays {
				if aura := auras.Get(&paladin.Unit); aura != nil {
					aura.MenuGroup = group.aura
				}
			}
			return
		}
	})

	switch paladin.Spec {
	case proto.Spec_SpecHolyPaladin:
		paladin.MenuGroupOrder = menuGroupOrder(menuSpellHeals, menuSpellPaladinAuras, menuSpellSeals, menuSpellDamage, menuSpellTank)
	case proto.Spec_SpecProtectionPaladin:
		paladin.MenuGroupOrder = menuGroupOrder(menuSpellTank, menuSpellSeals, menuSpellDamage, menuSpellPaladinAuras, menuSpellHeals)
	default:
		paladin.MenuGroupOrder = menuGroupOrder(menuSpellSeals, menuSpellDamage, menuSpellPaladinAuras, menuSpellHeals, menuSpellTank)
	}
}

// The spell groups in the order given, then their aura groups in the same order, then the rest under
// Paladin.
func menuGroupOrder(spellGroups ...string) []string {
	order := slices.Clone(spellGroups)
	for _, path := range spellGroups {
		for _, group := range menuGroups {
			if group.spell == path && group.aura != menuPaladin {
				order = append(order, group.aura)
			}
		}
	}
	return append(order, menuPaladin)
}
