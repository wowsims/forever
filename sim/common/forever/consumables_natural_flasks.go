package forever

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Flasks of Natural Accuracy 274273, Aggression 274274, Precision 274275 and Swiftness 274276. The
// first effect of each spell is its Stamina. The second, an A_DUMMY, is the bonus the spell's text
// grants in Mount Hyjal, Hyjal Summit and the Barrow Deeps, where the sim takes every fight to be; the
// auras after it name what the bonus is and state 0 themselves.
var NaturalFlasks = []int32{274273, 274274, 274275, 274276}

func init() {
	for _, itemID := range NaturalFlasks {
		core.NewConsumableEffect(itemID, func(agent core.Agent) {
			character := agent.GetCharacter()
			consumable := core.GetConsumableByID(itemID)
			row := spelldata.MustFind(consumable.OnUse().BuffId)

			config := spelldata.AuraConfig(row, spelldata.Label(consumable.Name))
			config.ActionID = core.ActionID{ItemID: itemID}
			config.BuildPhase = core.CharacterBuildPhaseConsumes
			spelldata.Permanent()(&config)
			aura := character.GetOrRegisterAura(config)

			spelldata.ParseEffects(character, aura, row, spelldata.Effects(1))
			bonus, err := NaturalFlaskZoneBonusOf(row, row.EffectN(2).Average())
			if err != nil {
				panic(err)
			}
			aura.AttachStatsBuff(bonus.Stats)
			if bonus.AttackHastePercent != 0 {
				aura.AttachMultiplyAttackSpeed(core.SymmetricMultiplierFromPercent(bonus.AttackHastePercent))
			}
			if bonus.CastHastePercent != 0 {
				aura.AttachMultiplyCastSpeed(core.SymmetricMultiplierFromPercent(bonus.CastHastePercent))
			}
		})
	}
}

// The zone bonus as the auras after the dummy name it, each by the dummy's percent: hit, crit and a
// lower chance to be dodged or parried as stats, and melee and ranged or casting speed as haste.
type NaturalFlaskZoneBonus struct {
	Stats              stats.Stats
	AttackHastePercent float64
	CastHastePercent   float64
}

func NaturalFlaskZoneBonusOf(row *spelldata.Spell, percent float64) (NaturalFlaskZoneBonus, error) {
	var bonus NaturalFlaskZoneBonus
	for _, e := range row.Effects[2:] {
		switch e.Aura {
		case dbcenums.A_MOD_HIT_CHANCE:
			bonus.Stats[stats.PhysicalHitPercent] = percent
		case dbcenums.A_MOD_SPELL_HIT_CHANCE:
			bonus.Stats[stats.SpellHitPercent] = percent
		case dbcenums.A_MOD_CRIT_PCT:
			bonus.Stats[stats.PhysicalCritPercent] = percent
			bonus.Stats[stats.SpellCritPercent] = percent
		case dbcenums.A_MOD_EXPERTISE:
			bonus.Stats[stats.ExpertisePercent] = percent
		case dbcenums.A_MOD_MELEE_RANGED_HASTE_2:
			bonus.AttackHastePercent = percent
		case dbcenums.A_MOD_CASTING_SPEED_NOT_STACK:
			bonus.CastHastePercent = percent
		default:
			return bonus, fmt.Errorf("%d %s: no zone bonus for %v", row.ID, row.Name, e.Aura)
		}
	}
	return bonus, nil
}
