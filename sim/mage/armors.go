package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The Chilled procs on Frost and Ice Armor slow the attacker, which does nothing in the sim. One
// armor is up at a time.
func (mage *Mage) registerArmorSpells() {
	// Frost Warding raises Frost and Ice Armor's armor and resistance.
	warding := 1 + spellData.FrostWarding.EffectAt(1).FractionAt(mage.Talents.FrostWarding)
	// Arcane Shielding's effect 2 raises Mage Armor's resistances (effect 1), so that effect is
	// scaled here; the casting regen (effect 2) is parsed.
	shielding := 1 + spellData.ArcaneShielding.EffectAt(2).FractionAt(mage.Talents.ArcaneShielding)

	// Only the highest rank is registered: the mage casts the rank configured here, not every rank
	// spellData.FrostArmor lists.
	frostArmorRank := spellData.FrostArmor.Highest()
	mage.registerArmor(frostArmorRank, MageSpellFrostArmor, stats.Stats{
		stats.Armor: warding * frostArmorRank.EffectN(1).BaseValue(),
	})

	// Only the highest rank is registered: the mage casts the rank configured here, not every rank
	// spellData.IceArmor lists.
	iceArmorRank := spellData.IceArmor.Highest()
	iceArmor := mage.registerArmor(iceArmorRank, MageSpellIceArmor, stats.Stats{
		stats.Armor:           warding * iceArmorRank.EffectN(1).BaseValue(),
		stats.FrostResistance: warding * iceArmorRank.EffectN(3).BaseValue(),
	})

	// Only the highest rank is registered: the mage casts the rank configured here, not every rank
	// spellData.MageArmor lists.
	mageArmorRank := spellData.MageArmor.Highest()
	resist := shielding * mageArmorRank.EffectN(1).BaseValue()
	mageArmor := mage.registerArmor(mageArmorRank, MageSpellMageArmor, stats.Stats{
		stats.ArcaneResistance: resist,
		stats.FireResistance:   resist,
		stats.FrostResistance:  resist,
		stats.NatureResistance: resist,
		stats.ShadowResistance: resist,
	})
	spelldata.ParseEffects(&mage.Character, mageArmor, mageArmorRank, spelldata.Effects(2))

	if mage.Options == nil {
		return
	}

	defaultArmor := map[proto.MageArmor]*core.Aura{
		proto.MageArmor_MageArmorFrostArmor: iceArmor,
		proto.MageArmor_MageArmorMageArmor:  mageArmor,
	}[mage.Options.DefaultMageArmor]

	if defaultArmor != nil {
		defaultArmor.BuildPhase = core.CharacterBuildPhaseBuffs
		defaultArmor.ApplyOnReset(func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		})
	}
}

func (mage *Mage) registerArmor(rank *spelldata.Spell, classMask int64, bonus stats.Stats) *core.Aura {
	aura := mage.RegisterAura(spelldata.AuraConfig(rank, spelldata.RankLabel(rank, &mage.Unit))).AttachStatsBuff(bonus)
	aura.NewExclusiveEffect("Mage Armor", true, core.ExclusiveEffect{})

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagHelpful))
	config.ClassSpellMask = classMask
	config.RelatedSelfBuff = aura

	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return !aura.IsActive()
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}

	mage.RegisterSpell(config)
	return aura
}
