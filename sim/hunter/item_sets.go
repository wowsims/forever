package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The rows ItemSetSpell names for the set's thresholds. None of them is a hunter class spell, so the
// generated table carries no reference to them.
var (
	cryptstalkerRapidFire      = spelldata.MustFind(28755) // 2pc: Rapid Fire lasts 4 s longer
	cryptstalkerAdrenalineProc = spelldata.MustFind(28752) // 6pc: Adrenaline Rush, the proc
	cryptstalkerAdrenalineMana = spelldata.MustFind(28753) // 6pc: Adrenaline Rush, the mana it grants
	cryptstalkerShotCost       = spelldata.MustFind(28751) // 8pc: Multi-Shot and Aimed Shot cost 20 less
)

var ItemSetCryptstalkerArmor = core.NewItemSet(core.ItemSet{
	Name: "Cryptstalker Armor",
	ID:   530,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			hunter := agent.(HunterAgent).GetHunter()
			spelldata.ParseEffects(&hunter.Character, setBonusAura, cryptstalkerRapidFire)
		},
		// (4) Set: While your pet is active, increases Attack Power by 50 for both you and your pet.
		// The pet is not modelled.
		6: func(agent core.Agent, setBonusAura *core.Aura) {
			hunter := agent.(HunterAgent).GetHunter()
			manaMetrics := hunter.NewManaMetrics(core.ActionID{SpellID: cryptstalkerAdrenalineMana.ID})
			mana := cryptstalkerAdrenalineMana.EnergizeEffect().Average()

			// 28752 states "always" on ranged hits and its proc effect names Multi-Shot alone; the
			// tooltip says your ranged critical hits, and the tooltip wins: the crit is the outcome
			// and the mask comes off.
			trigger := spelldata.ProcTrigger(&hunter.Character, cryptstalkerAdrenalineProc,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					hunter.AddMana(sim, mana, manaMetrics)
				})
			trigger.Name = "Cryptstalker Armor - 6PC"
			trigger.Outcome = core.OutcomeCrit
			trigger.ClassFlags = core.ClassFlags{}

			setBonusAura.AttachProcTrigger(trigger)
		},
		8: func(agent core.Agent, setBonusAura *core.Aura) {
			hunter := agent.(HunterAgent).GetHunter()
			spelldata.ParseEffects(&hunter.Character, setBonusAura, cryptstalkerShotCost)
		},
	},
})

var pvpGloveItemIDs = []int32{23279, 22862, 16463, 16571}

// The gloves' equip spell: Multi-Shot deals 4% more.
var pvpGloveMultiShot = spelldata.MustFind(28539)

func init() {
	for _, itemID := range pvpGloveItemIDs {
		core.NewItemEffect(itemID, func(_ core.Agent) {})
	}
}

func (hunter *Hunter) addPvpGloves() {
	hunter.RegisterPvPGloveMod(
		pvpGloveItemIDs,
		core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Flat,
			ClassFlags: multiShotRank.ClassFlags,
			FloatValue: pvpGloveMultiShot.EffectN(1).Percent(),
		})
}
