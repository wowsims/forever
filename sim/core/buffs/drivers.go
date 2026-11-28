package buffs

import (
	"fmt"
	"math"
	"slices"
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

// Battlegear of Wrath: item set 218's ItemSetSpell at three pieces is 23563, which adds a flat
// 30 to every effect of the Battle Shout family.
var BattleShoutT2Bonus = spelldata.MustFind(23563).EffectN(1).BaseValue()

// The raid's Battle Shout is the external caster's copy, which chains behind
// the player's own shout. Its state is the shout that warrior casts: 1 the
// trainer's Rank 6, 2 Rank 7 from Manual of Battle Shout VII, 3 Rank 7 with the set on.
func driveBattleShout(char *core.Character, raid *proto.RaidBuffs) {
	rank := core.Ternary(raid.BattleShout == 1, battleShoutSpell, battleShoutBookRank)
	aura := NewSpellDataBattleShoutRankAura(&char.Unit, false, rank)
	if raid.BattleShout >= 3 {
		AddGeneratedFlatBonus(aura, stats.AttackPower, BattleShoutRankValue(rank), BattleShoutT2Bonus)
	}
	applyFixedShoutAura(char, aura, BattleShoutCategory)
}

// The external copy chains behind every copy of the player's own (tag 0), since a warrior registers
// one per rank and the rotation picks which it casts.
func applyFixedShoutAura(char *core.Character, aura *core.Aura, category string) {
	aura.ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
		for _, playerAura := range char.GetAurasWithTag(category) {
			if playerAura.ActionID.Tag != 0 {
				continue
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
		}
	})

	applyFixedUptimeAura(aura, 1, aura.Duration+1, -1)
}

// A landing of the raid's debuffs (Target.ApplyRaidDebuffsAt) runs the OnReset again, and the schedule it starts
// replaces the one still running. A tick on a target with the raid's debuffs disabled does nothing.
func scheduledAura(aura *core.Aura, options core.PeriodicActionOptions) {
	var schedule *core.PendingAction
	onAction := options.OnAction
	options.OnAction = func(sim *core.Simulation) {
		if !aura.Unit.RaidDebuffsDisabled() {
			onAction(sim)
		}
	}
	aura.OnReset = func(aura *core.Aura, sim *core.Simulation) {
		if schedule != nil {
			schedule.Cancel(sim)
		}
		schedule = core.StartPeriodicAction(sim, options)
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
// mana cooldown is used first.
func driveInnervates(char *core.Character, individual *proto.IndividualBuffs) {
	aura := NewSpellDataInnervatesAura(&char.Unit, false, 0)

	threshold := 0.0
	char.Env.RegisterPostFinalizeEffect(func() {
		threshold = innervateManaThreshold(char)
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

// The Windfury Totem hit the raid aura procs, the main hand's weapon damage.
var windfuryTotemStrike = spelldata.MustFind(10610)

// For the procs that hear Windfury Totem hits apart from other main-hand specials, as a paladin's seals do.
func IsWindfuryTotemStrike(spell *core.Spell) bool {
	return spell.SpellID == windfuryTotemStrike.ID
}

// The totem's summon, which names the totem aura in the metrics.
var windfuryTotemCast = spelldata.MustFind(10614)

// The driver keeps the Windfury Totem hit the proc lands, and the totem aura that holds the category.
func driveWindfuryTotem(char *core.Character, _ *proto.RaidBuffs) {
	inheritWeaponTotem(char, windfuryTotemSpell, standWindfuryTotem)
}

func standWindfuryTotem(char *core.Character) *core.Aura {
	procTrigger := NewWindfuryTotemProcTrigger(char)

	totemAura := char.GetOrRegisterAura(core.Aura{
		Label:    "Windfury Totem",
		ActionID: core.ActionID{SpellID: windfuryTotemCast.ID, Tag: -1},
	})
	holdTotemAura(totemAura, WindfuryTotemCategory, WindfuryTotemPriority(), procTrigger)
	HoldOverWindfuryTotemRivals(totemAura)
	return totemAura
}

// A Windfury Totem bids the attack power its hits add, the raid aura's
// dummy effect, for the category.
func WindfuryTotemPriority() float64 {
	return windfuryTotemSpell.Effect(dbcenums.A_DUMMY, 0).BaseValue()
}

const windfuryTotemTriggerLabel = "Windfury Totem Trigger"

// Windfury Totem does not stack with Flametongue Totem: while it stands, it
// outbids Flametongue Totem's proc, whoever stands each totem. Grace of Air
// Totem stacks with it (EV-0102).
var windfuryTotemHoldsOver = []string{
	FlametongueTotemCategory,
}

// The Windfury Totem hit deals the main hand's weapon damage, after a Shank too (EV-0230). Rank 3 adds
// attack power and ranks 1 and 2 add none (EV-0002). The sim stands rank 3 only. Swing counting
// in EV-0136. Swing rage in EV-0138.
func newWindfuryTotemStrike(char *core.Character) *core.Spell {
	attackPower := WindfuryTotemPriority()
	config := spelldata.SpellConfig(&char.Unit, windfuryTotemStrike,
		spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Proc(),
		spelldata.Flags(core.SpellFlagCountsAsSwing), spelldata.Triggered())
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		weaponDamage := char.MHWeaponDamage(sim, spell.MeleeAttackPower(target)+attackPower)
		baseDamage := windfuryTotemStrike.WeaponStrike(sim, spell, weaponDamage, 0)
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
	}
	return char.RegisterSpell(config)
}

// The Windfury Totem proc, inactive until a totem
// aura holding WindfuryTotemCategory turns it on. A character under both the
// party buff and a totem of its own casting has one proc, built once.
func NewWindfuryTotemProcTrigger(char *core.Character) *core.Aura {
	if procTrigger := char.GetAura(windfuryTotemTriggerLabel); procTrigger != nil {
		return procTrigger
	}
	strike := newWindfuryTotemStrike(char)

	// Off-hand hits in EV-0155. The row's mask names both hands; off-hand swings never roll the totem
	// (EV-0008). The sim holds Whirlwind's off-hand strike to the same rule (no log measures it).
	// Ranged attacks and spells never roll it, and neither does the hit of a proc without Not a Proc,
	// since the row lacks Can Proc From Procs. An auto attack, another proc's extra swing included,
	// always rolls it.
	// The proc rolls off the hit's cast result, and the Windfury Totem hit follows once that hit has
	// been dealt, in the same instant. It requires a main hand, but it lands inside the 100 ms cooldown
	// its proc has just started, so it never rolls.
	trigger := spelldata.ProcTrigger(char, windfuryTotemSpell, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		target := result.Target
		sim.RunAfterCurrentEvent(func(sim *core.Simulation) {
			if target.IsEnabled() {
				strike.Cast(sim, target)
			}
		})
	})
	trigger.Name = windfuryTotemTriggerLabel
	trigger.MetricsActionID = core.ActionID{SpellID: windfuryTotemSpell.ID, Tag: -1}
	trigger.Duration = core.NeverExpires
	trigger.Callback = core.CallbackOnCastResult
	trigger.TriggerImmediately = true
	// Triggered hits with Not a Proc roll the totem as abilities do (EV-0002).
	// A hit with no defense type is no melee attack and never rolls it.
	// Off-hand specials in EV-0155.
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spell.DefenseType != core.DefenseTypeNone &&
			(spell.IsMH() || spell.Flags.Matches(core.SpellFlagAPL))
	}
	trigger.ProcMask &= core.ProcMaskMeleeMH | core.ProcMaskMeleeOHSpecial
	trigger.RequireDamageDealt = false
	return char.MakeProcTriggerAura(trigger)
}

// A standing Windfury Totem, the raid's or the shaman's own, outbids its rivals.
func HoldOverWindfuryTotemRivals(totemAura *core.Aura) {
	for _, category := range windfuryTotemHoldsOver {
		totemAura.NewExclusiveEffect(category, false, core.ExclusiveEffect{Priority: math.MaxFloat64})
	}
}

// A weapon totem's raid aura bids for the category while it is up, which keeps
// the proc trigger up while it holds it. Outbid, by the shaman's own totem or a
// standing Windfury Totem, it stays up and takes the category back once the higher
// bid goes.
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

// The totem's hit logs in game as Flametongue Attack (EV-0013); 16389 only states its base points.
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

// Every main-hand auto attack that lands adds fire damage priced by the
// weapon's speed. The two totems do not stack: under both, Windfury Totem's proc holds and
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

	config := spelldata.SpellConfig(&char.Unit, flametongueTotemAttack, spelldata.Proc(), spelldata.Triggered())
	config.ActionID = core.ActionID{SpellID: flametongueTotemCast.ID, Tag: 1}
	config.ProcMask = core.ProcMaskSpellDamage
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

// The individual Thorns' damage carries the providing druid's Nature spell
// power, which the buffs state alongside it.
func driveThorns(char *core.Character, individual *proto.IndividualBuffs) {
	core.MakePermanent(NewSpellDataThornsAuraWithSpellPower(&char.Unit, false, individual.ThornsSpellPower))
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
// 60 s ICD plus the 64 s a 5% chance takes to land one 3.2 s swing on average, 16%.
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

const annihilatorItemID = 12798

// The weapon each state of the raid's Puncture Armor names, from 1, and the debuff its proc lands:
// Bashguuder and Rivenspike land the row's 17315, Vibroblade its own 1293005.
var punctureArmorWeapons = []struct {
	itemID int32
	meta   *Meta
}{
	{13204, punctureArmorMeta},
	{13286, punctureArmorMeta},
	{9485, punctureArmorMeta.atState(spelldata.MustFind(1293005), punctureArmorMeta.Label)},
}

// With the player wielding Annihilator, the raid's copy is the player's own, which the player's procs
// stack and refresh, so the target carries one stack count as with Sunder Armor.
func driveArmorShatter(target *core.Unit, _ *proto.Debuffs, raid *proto.Raid) {
	rollWielderHits(NewSpellDataArmorShatterAura(target, playerWields(raid, annihilatorItemID), 0), annihilatorItemID)
}

// Bashguuder and Rivenspike both land 17315, so with the player wielding either, the raid's copy of
// either is the player's own, which the player's procs stack and refresh. Vibroblade's copy
// stays the raid's: its 1293005 shares the label, so a player copy of it would take the 17315 procs
// on a row they do not land.
func drivePunctureArmor(target *core.Unit, debuffs *proto.Debuffs, raid *proto.Raid) {
	weapon := punctureArmorWeapons[min(int(debuffs.PunctureArmor), len(punctureArmorWeapons))-1]
	isPlayer := weapon.meta == punctureArmorMeta && playerWields(raid, 13204, 13286)
	rollWielderHits(newDebuff(target, weapon.meta, isPlayer, 0), weapon.itemID)
}

// Whether the simmed player carries one of the items, equipped or in an enabled item swap, where its
// procs are registered as well. A raid with no player wields nothing.
func playerWields(raid *proto.Raid, itemIDs ...int32) bool {
	var player *proto.Player
	if parties := raid.GetParties(); len(parties) > 0 && len(parties[0].GetPlayers()) > 0 {
		player = parties[0].GetPlayers()[0]
	}
	items := player.GetEquipment().GetItems()
	if player.GetEnableItemSwap() {
		items = slices.Concat(items, player.GetItemSwap().GetItems())
	}
	return slices.ContainsFunc(items, func(item *proto.ItemSpec) bool { return slices.Contains(itemIDs, item.GetId()) })
}

// Wielder hits in EV-0147.
const wielderOtherHitsPerMinute = 48

// The share of its proc's strikes a wielder's target dodges, whatever the target's level. The wielder
// has the hit its strikes need to never miss (EV-0152).
const wielderStrikeDodgeChance = 0.065

// A weapon's proc debuff from one main-hand wielder. Every target the raid's debuffs land on has a
// wielder of its own, and the player wielding the weapon stacks the same aura. The wielder lands its
// swings at the item's base speed and wielderOtherHitsPerMinute more hits, each rolling the item's
// procs per minute; a proc refreshes the aura and adds a stack up to its max. Proc strikes in R-0040.
func rollWielderHits(aura *core.Aura, itemID int32) {
	item := core.GetItemByID(itemID)
	if item == nil {
		panic(fmt.Sprintf("the raid's %s reads the swing speed and procs per minute of item %d, which the sim's item database lacks; the request's database must carry it", aura.Label, itemID))
	}
	chance := itemProcPPM(item, aura.ActionID.SpellID) * item.SwingSpeed / 60
	swingsPerMinute := 60 / item.SwingSpeed
	hitsPerMinute := swingsPerMinute + wielderOtherHitsPerMinute
	swingShare := swingsPerMinute / hitsPerMinute
	strikes := spelldata.MustFind(aura.ActionID.SpellID).DefenseTypeCore() == core.DefenseTypeMelee

	lands := func(sim *core.Simulation) bool {
		if sim.RandomFloat(aura.Label) >= chance {
			return false
		}
		if strikes && sim.RandomFloat(aura.Label) < wielderStrikeDodgeChance {
			return false
		}
		aura.Activate(sim)
		if aura.MaxStacks > 0 {
			aura.AddStack(sim)
		}
		return true
	}

	scheduledAura(aura, core.PeriodicActionOptions{
		Period:          core.DurationFromSeconds(60 / hitsPerMinute),
		TickImmediately: true,
		Priority:        core.ActionPriorityDOT,
		OnAction: func(sim *core.Simulation) {
			if lands(sim) && strikes && sim.RandomFloat(aura.Label) < swingShare {
				lands(sim)
			}
		},
	})
}

// The procs per minute the item's chance-on-hit entry for the spell states.
func itemProcPPM(item *core.Item, spellID int32) float64 {
	for _, effect := range item.ItemEffects {
		if effect.BuffId == spellID && effect.GetProc().GetPpm() > 0 {
			return effect.GetProc().GetPpm()
		}
	}
	panic(fmt.Sprintf("item %d states no procs per minute for spell %d", item.ID, spellID))
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
