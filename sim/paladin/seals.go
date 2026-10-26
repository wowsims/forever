package paladin

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only one seal is up at a time.
const SealCategory = "PaladinSeal"

// One rank of one seal: the castable spell's row, the aura the cast puts up, the judgement it
// unleashes, and the per-hit effect Twist of Light's Echo can replay.
type sealConfig struct {
	rank      *spelldata.Spell
	classMask int64
	spell     *core.Spell
	aura      *core.Aura
	judgement *core.Spell

	// Twist of Light: the Echo this seal leaves when replaced (0 for the seals that leave none),
	// and what the Echo does to the target of the next melee attack (nil where the seal's effect
	// has no place in the sim).
	echoID int32
	echo   func(sim *core.Simulation, target *core.Unit)
}

// Only one Twist of Light Echo is up at a time.
const EchoCategory = "PaladinEcho"

// The Echoes Twist of Light names, one per seal.
const (
	echoOfCommandID       = 1311703
	echoOfFuryID          = 1311701
	echoOfRighteousnessID = 1311704
	echoOfJusticeID       = 1311705
)

// One Echo: its aura on the paladin, and the seal rank whose effect it holds.
type sealEcho struct {
	aura *core.Aura
	seal *sealConfig
}

// The seals' rows as the game files state them. Seal of Command is a talent and registers from
// registerTalentSpells.
var (
	SealOfRighteousnessRankMap = spellData.SealOfRighteousness
	SealOfCommandRankMap       = spellData.SealOfCommand
	SealOfLightRankMap         = spellData.SealOfLight
	SealOfWisdomRankMap        = spellData.SealOfWisdom
	SealOfTheCrusaderRankMap   = spellData.SealOfTheCrusader
	SealOfFuryRankMap          = spellData.SealOfFury
)

func (paladin *Paladin) registerSeals() {
	SealOfRighteousnessRankMap.Each(paladin.registerSealOfRighteousness)
	SealOfLightRankMap.Each(paladin.registerSealOfLight)
	SealOfWisdomRankMap.Each(paladin.registerSealOfWisdom)
	paladin.registerSealOfJustice()
	SealOfTheCrusaderRankMap.Each(paladin.registerSealOfTheCrusader)
	SealOfFuryRankMap.Each(paladin.registerSealOfFury)
}

// Every seal aura joins the seal category, so casting one seal drops the other.
func (paladin *Paladin) makeSealExclusive(aura *core.Aura) *core.Aura {
	aura.NewExclusiveEffect(SealCategory, true, core.ExclusiveEffect{})
	return aura
}

// The castable seal spell: instant, on the GCD, priced as the row says.
func (paladin *Paladin) registerSealSpell(cfg *sealConfig) *core.Spell {
	config := spelldata.SpellConfig(&paladin.Unit, cfg.rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = cfg.classMask

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		paladin.applySeal(sim, cfg)
	}

	config.RelatedSelfBuff = cfg.aura

	// Judgement presses the seal's judgement, which holds the casts.
	cfg.judgement.Flags |= core.SpellFlagPressed

	cfg.spell = paladin.RegisterSpell(config)
	return cfg.spell
}

func (paladin *Paladin) applySeal(sim *core.Simulation, cfg *sealConfig) {
	previous := paladin.currentSeal
	if previous != nil && previous != cfg && previous.echoID != cfg.echoID && previous.aura.IsActive() {
		if echo := paladin.echoes[previous.echoID]; echo != nil {
			echo.seal = previous
			echo.aura.Activate(sim)
		}
	}

	// A seal and its own Echo are never up together: putting the seal back up drops its Echo.
	if echo := paladin.echoes[cfg.echoID]; echo != nil && echo.aura.IsActive() {
		echo.seal = nil
		echo.aura.Deactivate(sim)
	}

	paladin.currentSeal = cfg
	// The seal category drops whatever seal was up; recasting the same seal refreshes it.
	cfg.aura.Activate(sim)
}

// The seal the paladin is under, if any.
func (paladin *Paladin) activeSeal() *sealConfig {
	if paladin.currentSeal != nil && paladin.currentSeal.aura.IsActive() {
		return paladin.currentSeal
	}
	return nil
}

// A seal's per-hit damage lands a batch window after the swing that fired it, the way the game
// resolves the proc after the melee hit.
func dealAfterBatch(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	afterBatch(sim, core.ActionPriorityLow, func(sim *core.Simulation) {
		spell.DealDamage(sim, result)
	})
}

// Runs the action a batch window from now.
func afterBatch(sim *core.Simulation, priority core.ActionPriority, onAction func(*core.Simulation)) {
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{
		DoAt:     sim.CurrentTime + core.SpellBatchWindow,
		Priority: priority,
		OnAction: onAction,
	}))
}

// The judgement debuffs Light, Wisdom and the Crusader put up: the paladin's melee strikes
// refresh them, and Judgement has to find the aura for the target it judges.
func (paladin *Paladin) newJudgementAuras(makeAura func(target *core.Unit) *core.Aura) core.AuraArray {
	auras := paladin.NewEnemyAuraArray(makeAura)
	paladin.JudgementAuras = append(paladin.JudgementAuras, auras)
	return auras
}

// Refreshes every judgement debuff this paladin has on the target, which the game does on every
// melee strike that lands.
func (paladin *Paladin) refreshJudgements(sim *core.Simulation, target *core.Unit) {
	for _, auras := range paladin.JudgementAuras {
		if aura := auras.Get(target); aura.IsActive() {
			aura.Refresh(sim)
		}
	}
}

var sealDuration = time.Second * 30
