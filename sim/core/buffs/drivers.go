package buffs

import (
	"fmt"
	"math"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The buff drivers the generator calls by name. A generated apply block for a
// cooldown, proc, uptime, item-count or manual row, or for a row the manifest
// marks as driven, calls drive<Go>(unit, scope); declaring that function here
// is what makes the row compile. A buff-scope row
// hands the driver the *Character the buffs are applied to and the message its
// field lives on, a debuff-scope row the *Unit the debuffs land on plus the
// debuffs and the raid. The driver reads its own field out of the message, and
// any other field it needs: Flametongue Totem does nothing under Windfury Totem.

// Three pieces of Battlegear of Wrath are worth 30 more attack power on Battle
// Shout: item set 218's ItemSetSpell at three pieces is 23563, which adds a flat
// 30 to every effect of the Battle Shout family.
var BattleShoutT2Bonus = spelldata.MustFind(23563).EffectN(1).BaseValue()

// The raid's Battle Shout is the external caster's copy, which chains behind
// the player's own shout rather than being up from the start. Its improved state
// says that warrior shouted with the set on.
func driveBattleShout(char *core.Character, raid *proto.RaidBuffs) {
	aura := NewSpellDataBattleShoutAura(&char.Unit, false, 0)
	if raid.BattleShout == proto.TristateEffect_TristateEffectImproved {
		AddGeneratedFlatBonus(aura, stats.AttackPower, BattleShoutValue(0), BattleShoutT2Bonus)
	}
	applyFixedShoutAura(char, aura, BattleShoutCategory)
}

func applyFixedShoutAura(char *core.Character, aura *core.Aura, category string) {
	aura.ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
		auras := char.GetAurasWithTag(category)
		var playerAura *core.Aura
		for _, bsAura := range auras {
			if bsAura.ActionID.Tag == 0 {
				playerAura = bsAura
				break
			}
		}

		if playerAura == nil {
			return
		}

		playerAura.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
			pa := sim.GetConsumedPendingActionFromPool()
			pa.NextActionAt = playerAura.ExpiresAt() + char.ReactionTime
			pa.OnAction = func(sim *core.Simulation) {
				aura.Activate(sim)

				core.StartPeriodicAction(sim, core.PeriodicActionOptions{
					Period:   aura.Duration + 1,
					NumTicks: 1,
					OnAction: func(sim *core.Simulation) {
						aura.Activate(sim)
					},
				})
			}
			sim.AddPendingAction(pa)
		})
	})

	applyFixedUptimeAura(aura, 1, aura.Duration+1, -1)
}

func scheduledAura(aura *core.Aura, options core.PeriodicActionOptions) {
	aura.OnReset = func(aura *core.Aura, sim *core.Simulation) {
		aura.Duration = core.NeverExpires
		core.StartPeriodicAction(sim, options)
	}
}

func applyFixedUptimeAura(aura *core.Aura, uptime float64, tickLength time.Duration, startTime time.Duration) {
	auraDuration := aura.Duration
	ticksPerAura := float64(auraDuration) / float64(tickLength)
	chancePerTick := core.TernaryFloat64(uptime == 1, 1, 1.0-math.Pow(1-uptime, 1/ticksPerAura))

	aura.Unit.RegisterResetEffect(func(sim *core.Simulation) {
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period: tickLength,
			OnAction: func(sim *core.Simulation) {
				if sim.RandomFloat("FixedAura") < chancePerTick {
					aura.Activate(sim)
					if aura.MaxStacks > 0 {
						aura.AddStack(sim)
					}
				}
			},
		})

		// Also try once at the start.
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period:   startTime,
			NumTicks: 1,
			OnAction: func(sim *core.Simulation) {
				if sim.RandomFloat("FixedAura") < uptime {
					// Use random duration to compensate for increased chance collapsed into single tick.
					randomDur := tickLength + time.Duration(float64(auraDuration-tickLength)*sim.RandomFloat("FixedAuraDur"))

					aura.Duration = randomDur
					aura.Activate(sim)
					aura.Duration = auraDuration
				}
			},
		})
	})
}

// A druid innervates a character who is nearly out of mana, so that every other
// mana cooldown is spent first. The row's auras raise the spirit regen while it
// is up, and the metrics record what the character gains from it.
func driveInnervates(char *core.Character, individual *proto.IndividualBuffs) {
	aura := NewSpellDataInnervatesAura(&char.Unit, false, 0)
	manaMetrics := char.NewManaMetrics(aura.ActionID)

	threshold, expectedMana := 0.0, 0.0
	char.Env.RegisterPostFinalizeEffect(func() {
		threshold = innervateManaThreshold(char)
		expectedMana = InnervateMana(&char.Unit, aura)
	})

	const ticks = 10
	aura.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
		perTick := expectedMana / ticks
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period:   aura.Duration / ticks,
			NumTicks: ticks,
			OnAction: func(sim *core.Simulation) {
				manaMetrics.AddEvent(perTick, perTick)
			},
		})
	})

	newGeneratedExternalCD(char, aura, generatedExternalCD{
		NumSources: individual.Innervates,
		Cooldown:   InnervatesCooldown(),
		Type:       core.CooldownTypeMana,
		ShouldActivate: func(_ *core.Simulation, char *core.Character) bool {
			return char.CurrentMana() <= threshold
		},
	})
}

// The mana an Innervate brings the unit while the aura lasts: spell 29166 states +400% spirit regen
// on the mana bar (aura 110).
func InnervateMana(unit *core.Unit, aura *core.Aura) float64 {
	return unit.SpiritManaRegenPerSecond() *
		core.MultiplierFromPercent(innervatesSpell.Effect(dbcenums.A_MOD_POWER_REGEN_PERCENT, int32(dbcenums.POWER_MANA)).BaseValue()) *
		aura.Duration.Seconds()
}

// A mage burns mana fast enough that waiting for a flat thousand left would
// waste most of the innervate.
func innervateManaThreshold(char *core.Character) float64 {
	if char.Class == proto.Class_ClassMage {
		return char.MaxMana() * 0.4
	}
	return 1000
}

// The priest has nothing to hold Power Infusion for, so it goes out on cooldown.
func drivePowerInfusions(char *core.Character, individual *proto.IndividualBuffs) {
	newGeneratedExternalCD(char, NewSpellDataPowerInfusionsAura(&char.Unit, false, 0), generatedExternalCD{
		NumSources: individual.PowerInfusions,
		Cooldown:   PowerInfusionsCooldown(),
		Type:       core.CooldownTypeDPS,
	})
}

// A restoration shaman drops Mana Tide once the party has mana to refill, which
// is 40 seconds in, or halfway through a fight shorter than that.
func driveManaTideTotems(char *core.Character, party *proto.PartyBuffs) {
	initialDelay := time.Duration(0)
	char.Env.RegisterPostFinalizeEffect(func() {
		initialDelay = min(char.Env.BaseDuration/2, time.Second*40)
	})

	totem := NewSpellDataManaTideTotemsAura(&char.Unit, false, 0)
	newGeneratedExternalCD(char, totem, generatedExternalCD{
		NumSources: party.ManaTideTotems,
		Cooldown:   ManaTideTotemsCooldown(),
		Type:       core.CooldownTypeMana,
		ShouldActivate: func(sim *core.Simulation, _ *core.Character) bool {
			return sim.CurrentTime >= initialDelay
		},
	})
	inheritTotem(char, manaTideTotemsSpell, totem, func(pet *core.Pet) *core.Aura {
		return NewSpellDataManaTideTotemsAura(&pet.Unit, false, 0)
	})
}

// The raid aura the totem puts up, a Proc Trigger Spell aura that names 10610.
var windfuryTotemRaidAura = spelldata.MustFind(10612)

// The totem's summon, which names the totem aura in the metrics.
var windfuryTotemCast = spelldata.MustFind(10614)

// The totem's aura is the attack power a windfury proc grants. The totem hands
// the proc out as the raid aura 10612, whose row states the chance, what it
// hears and its 100 ms cooldown. The driver keeps the extra attack the proc
// lands, and the totem aura that holds the category.
func driveWindfuryTotem(char *core.Character, _ *proto.RaidBuffs) {
	inheritWeaponTotem(char, windfuryTotemRaidAura, standWindfuryTotem)
}

func standWindfuryTotem(char *core.Character) *core.Aura {
	procTrigger := NewWindfuryTotemProcTrigger(char)

	totemAura := char.GetOrRegisterAura(core.Aura{
		Label:    "Windfury Totem",
		ActionID: core.ActionID{SpellID: windfuryTotemCast.ID, Tag: -1},
	})
	holdTotemAura(totemAura, WindfuryTotemCategory, WindfuryTotemValue(0), procTrigger)
	HoldOverWindfuryTotemRivals(totemAura)
	return totemAura
}

const windfuryTotemTriggerLabel = "Windfury Totem Trigger"

// Windfury Totem does not stack with Flametongue Totem: while it stands, it
// outbids Flametongue Totem's proc, whoever stands each totem. Grace of Air
// Totem stacks with it as of 1.60.1.70425.
var windfuryTotemHoldsOver = []string{
	FlametongueTotemCategory,
}

// The Windfury Totem proc on the character's main hand, inactive until a totem
// aura holding WindfuryTotemCategory turns it on. A character under both the
// party buff and a totem of its own casting has one proc, built once.
func NewWindfuryTotemProcTrigger(char *core.Character) *core.Aura {
	if procTrigger := char.GetAura(windfuryTotemTriggerLabel); procTrigger != nil {
		return procTrigger
	}

	procAura := NewSpellDataWindfuryTotemAura(&char.Unit, false, 0)
	// The attack power is only there for a moment after a proc, so it is not
	// part of the stats the character sheet is measured with.
	procAura.BuildPhase = core.CharacterBuildPhaseNone

	// The row's own proc flags say what spends a charge: every auto attack that
	// lands, since the column's chance is the raid aura's roll. The attack
	// power stays until the charges are gone or the row's duration runs out.
	procAura.MaxStacks = int32(windfuryTotemSpell.ProcCharges)
	spender := spelldata.ProcTrigger(char, windfuryTotemSpell, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		procAura.RemoveStack(sim)
	}, spelldata.Chance(1))
	spender.Name = "Windfury Totem Charges"
	spender.TriggerImmediately = true
	procAura.AttachProcTriggerCallback(&char.Unit, spender)

	// The row hears auto attacks and melee abilities, so the hit of a proc
	// without Not a Proc does not reach it. The mask names both hands; the
	// main-hand rule the tooltip states is the server's. The proc rolls off the
	// hit's cast result and puts the attack power up with every charge; the
	// hit's damage, rolled before the proc, does not gain it. The hit is then
	// dealt and spends a charge if it is a landed auto, and the extra attack
	// after it spends the next, so a main-hand auto's extra attack is the only
	// swing buffed. A special spends none and leaves the last charge to the auto
	// after its extra attack.
	trigger := spelldata.ProcTrigger(char, windfuryTotemRaidAura, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		procAura.Activate(sim)
		procAura.SetStacks(sim, procAura.MaxStacks)
	})
	trigger.Name = windfuryTotemTriggerLabel
	trigger.MetricsActionID = core.ActionID{SpellID: windfuryTotemRaidAura.ID, Tag: -1}
	trigger.Duration = core.NeverExpires
	trigger.Callback = core.CallbackOnCastResult
	trigger.ProcMask &= core.ProcMaskMeleeMH
	// TODO: verify in game whether a landed hit that deals no damage, such as Sunder Armor, rolls the totem, and if it does, clear RequireDamageDealt.
	onInit := char.ExtraAttackProc(&trigger, windfuryTotemSpell.ID, 1)
	return char.MakeProcTriggerAura(trigger).ApplyOnInit(onInit)
}

// A standing Windfury Totem, the raid's or the shaman's own, outbids its rivals
// even while Windfury Weapon on the main hand keeps its proc down.
func HoldOverWindfuryTotemRivals(totemAura *core.Aura) {
	for _, category := range windfuryTotemHoldsOver {
		totemAura.NewExclusiveEffect(category, false, core.ExclusiveEffect{Priority: math.MaxFloat64})
	}
}

// A weapon totem's raid aura bids for the category while it is up, which keeps
// the proc trigger up while it holds it. Outbid, by a main-hand
// imbue, the shaman's own totem or a standing Windfury Totem, it stays up and takes
// the category back once the higher bid goes.
func holdTotemAura(totemAura *core.Aura, category string, priority float64, procTrigger *core.Aura) {
	totemAura.NewExclusiveEffect(category, false, core.ExclusiveEffect{
		Priority: priority,
		OnGain: func(_ *core.ExclusiveEffect, sim *core.Simulation) {
			procTrigger.Activate(sim)
		},
		OnExpire: func(_ *core.ExclusiveEffect, sim *core.Simulation) {
			procTrigger.Deactivate(sim)
		},
	})
}

// The totem's raid aura names its proc in its base points rather than its
// trigger column, so the proc is read by id.
var flametongueTotemProc = spelldata.MustFind(16389)

// The totem's hit logs in game as Flametongue Attack; 16389 only states its base points.
var flametongueTotemAttack = spelldata.MustFind(16368)

// The totem's summon, which the hit reports under, tagged 1, so the results show the totem and its hits as one row.
var flametongueTotemCast = spelldata.MustFind(16387)

// A Flametongue Totem bids its proc's base points for the category.
func FlametongueTotemPriority() float64 {
	return flametongueTotemProc.EffectN(1).BaseValue()
}

// The fire damage a main-hand hit adds for a weapon of this speed.
func FlametongueTotemDamage(weaponSpeed float64) float64 {
	return flametongueTotemProc.EffectN(1).BaseValue() * weaponSpeed / 100
}

// The totem hands its proc out as the raid aura 15036, whose row hears auto
// attacks at 100%. Every main-hand auto attack that lands adds fire damage
// priced by the weapon's speed.
//
// The two totems do not stack: under both, Windfury Totem's proc holds and
// Flametongue Totem does nothing.
func driveFlametongueTotem(char *core.Character, raid *proto.RaidBuffs) {
	if raid.WindfuryTotem {
		return
	}
	inheritWeaponTotem(char, flametongueTotemSpell, standFlametongueTotem)
}

func standFlametongueTotem(char *core.Character) *core.Aura {
	procTrigger := NewFlametongueTotemProcTrigger(char)

	totemAura := NewSpellDataFlametongueTotemAura(&char.Unit, false, 0)
	totemAura.BuildPhase = core.CharacterBuildPhaseNone
	holdTotemAura(totemAura, FlametongueTotemCategory, FlametongueTotemPriority(), procTrigger)
	return totemAura
}

const flametongueTotemTriggerLabel = "Flametongue Totem Trigger"

// The Flametongue Totem proc on the character's main hand, inactive until a
// totem aura holding FlametongueTotemCategory turns it on. A character under
// both the raid buff and a totem of its own casting has one proc, built once,
// and it does nothing while the Windfury Totem proc is up.
func NewFlametongueTotemProcTrigger(char *core.Character) *core.Aura {
	if procTrigger := char.GetAura(flametongueTotemTriggerLabel); procTrigger != nil {
		return procTrigger
	}

	config := spelldata.SpellConfig(&char.Unit, flametongueTotemAttack, spelldata.Proc())
	config.ActionID = core.ActionID{SpellID: flametongueTotemCast.ID, Tag: 1}
	config.ProcMask = core.ProcMaskSpellDamageProc
	if flametongueTotemAttack.IsAProc() {
		config.Flags |= core.SpellFlagProc
	}
	config.DamageMultiplier = 1
	config.ThreatMultiplier = 1
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		damage := FlametongueTotemDamage(char.AutoAttacks.MH().SwingSpeed)
		spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHitAndCrit)
	}
	strike := char.RegisterSpell(config)

	// The mask names both hands; the main-hand rule the tooltip states is the
	// server's.
	trigger := spelldata.ProcTrigger(char, flametongueTotemSpell, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		strike.Cast(sim, result.Target)
	})
	trigger.Name = flametongueTotemTriggerLabel
	trigger.Duration = core.NeverExpires
	trigger.TriggerImmediately = true
	trigger.ProcMask &= core.ProcMaskMeleeMH
	return char.MakeProcTriggerAura(trigger)
}

// The party's Retribution Aura is the top rank, and its damage carries the
// providing paladin's Holy spell power, which the party states alongside it.
func driveRetributionAura(char *core.Character, party *proto.PartyBuffs) {
	core.MakePermanent(NewSpellDataRetributionAuraRank(char, false, RetributionAuraMaxRank, party.RetributionAuraSpellPower))
}

// The blessing does nothing on its own: the paladin's Holy Light and Flash of
// Light read their bonus off their own rows when the target carries it.
func driveGreaterBlessingOfLight(char *core.Character, _ *proto.IndividualBuffs) {
	core.MakePermanent(NewSpellDataGreaterBlessingOfLightAura(&char.Unit, false, 0))
}

// The staff's aura is worth its amounts once per Atiesh in the party.
func driveAtieshDruid(char *core.Character, party *proto.PartyBuffs) {
	core.MakePermanent(NewSpellDataAtieshDruidAura(&char.Unit, false, 0, float64(party.AtieshDruid)))
}

func driveAtieshMage(char *core.Character, party *proto.PartyBuffs) {
	core.MakePermanent(NewSpellDataAtieshMageAura(&char.Unit, false, 0, float64(party.AtieshMage)))
}

func driveAtieshPriest(char *core.Character, party *proto.PartyBuffs) {
	core.MakePermanent(NewSpellDataAtieshPriestAura(&char.Unit, false, 0, float64(party.AtieshPriest)))
}

func driveAtieshWarlock(char *core.Character, party *proto.PartyBuffs) {
	core.MakePermanent(NewSpellDataAtieshWarlockAura(&char.Unit, false, 0, float64(party.AtieshWarlock)))
}

// Darkspear's equip spell, which states the chance on hit and the time between procs.
var darkspearTrigger = spelldata.MustFind(1314143)

const darkspearItemID = 12802

// The party's copy is up for the share of the fight the wielder keeps it up: 16939's 20 s out of each
// 60 s internal cooldown plus the 64 s a 5% chance takes to land one 3.2 s swing on average, 16%.
// The wielder is assumed to land one melee hit per swing and nothing else.
func driveDarkspear(char *core.Character, _ *proto.PartyBuffs) {
	aura := NewSpellDataDarkspearAura(&char.Unit, false, 0)
	// The crit is up for a share of the fight only, so it is not part of the stats the character
	// sheet is measured with.
	aura.BuildPhase = core.CharacterBuildPhaseNone
	item := core.GetItemByID(darkspearItemID)
	if item == nil {
		panic(fmt.Sprintf("the party's Darkspear reads the swing speed of item %d, which the sim's item database lacks; the request's database must carry it", darkspearItemID))
	}
	swing := core.DurationFromSeconds(item.SwingSpeed)
	wait := time.Duration(float64(swing) / darkspearTrigger.StatedChance())
	uptime := aura.Duration.Seconds() / (darkspearTrigger.ICD() + wait).Seconds()
	applyFixedUptimeAura(aura, uptime, swing, 1)
}

// The judgement the paladin leaves on the target heals whoever strikes it; the
// game files' trigger spell 5373 is a dummy, so how much and how often is the
// driver's. The raid's copy is the top rank the paladin's own judgement states.
func driveJudgementOfLight(target *core.Unit, _ *proto.Debuffs, _ *proto.Raid) {
	AttachJudgementOfLightHeal(core.MakePermanent(NewSpellDataJudgementOfLightAura(target, false, 0)), nil, JudgementOfLightMaxRank)
}

// Judgement of Wisdom returns mana to whoever strikes the target, on the same
// terms: 1826 is a dummy, so the amount and the chance stay with the paladin's
// top rank.
func driveJudgementOfWisdom(target *core.Unit, _ *proto.Debuffs, _ *proto.Raid) {
	AttachJudgementOfWisdomMana(core.MakePermanent(NewSpellDataJudgementOfWisdomAura(target, false, 0)), JudgementOfWisdomMaxRank)
}

// A stack of Sunder Armor is worth nothing until it is on the target, so the
// raid's copy is ramped to five over the first five global cooldowns, which is
// how long a warrior takes to stack it. The stacks build while a stronger armor
// reduction holds the category, and the sunder applies once it outbids it or
// that one ends. With a warrior in the raid the raid's copy is the player's
// own, which the warrior's Sunder Armor stacks, so the target carries one.
func driveSunderArmor(target *core.Unit, _ *proto.Debuffs, raid *proto.Raid) {
	aura := core.MakePermanent(NewSpellDataSunderArmorAura(target, raidHasClass(raid, proto.Class_ClassWarrior), 0))

	scheduledAura(aura, core.PeriodicActionOptions{
		Period:          core.GCDDefault,
		NumTicks:        5,
		TickImmediately: true,
		Priority:        core.ActionPriorityDOT,
		OnAction: func(sim *core.Simulation) {
			aura.Activate(sim)
			aura.AddStack(sim)
		},
	})
}

func raidHasClass(raid *proto.Raid, class proto.Class) bool {
	for _, party := range raid.GetParties() {
		for _, player := range party.GetPlayers() {
			if player.GetClass() == class {
				return true
			}
		}
	}
	return false
}

// Annihilator procs often enough to keep all three stacks on the target.
func driveArmorShatter(target *core.Unit, _ *proto.Debuffs, _ *proto.Raid) {
	holdMaxStacks(core.MakePermanent(NewSpellDataArmorShatterAura(target, false, 0)))
}

// The game files state one stack, so the raid's copy takes it as soon as it lands.
func driveDustCloud(target *core.Unit, _ *proto.Debuffs, _ *proto.Raid) {
	holdMaxStacks(core.MakePermanent(NewSpellDataDustCloudAura(target, false, 0)))
}

func holdMaxStacks(aura *core.Aura) {
	aura.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
		aura.SetStacks(sim, aura.MaxStacks)
	})
}

var thunderfuryCycloneSpell = spelldata.MustFind(27648)

// The resistance aura is the row's; the triggered slow is an aura of its own in the attack-speed
// category, so the strongest slow on the target wins without touching the resistance.
func driveThunderfury(target *core.Unit, _ *proto.Debuffs, _ *proto.Raid) {
	core.MakePermanent(NewSpellDataThunderfuryAura(target, false, 0))

	cyclone := core.MakePermanent(target.GetOrRegisterAura(core.Aura{
		Label:    "Thunderfury Cyclone (External)",
		ActionID: core.ActionID{SpellID: thunderfuryCycloneSpell.ID, Tag: -1},
		Duration: core.NeverExpires,
	}))
	AtkSpeedReductionEffect(cyclone, core.MultiplierFromPercent(-thunderfuryCycloneSpell.EffectN(1).BaseValue()))
}
