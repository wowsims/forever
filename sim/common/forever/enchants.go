package forever

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Revelation's buff per class. No warrior or rogue gains one in the logs (EV-0149).
// TODO: Verify in game that warriors and rogues get no buff; EV-0149 has too few wearers to settle it.
var revelationBuffs = map[proto.Class]int32{
	proto.Class_ClassMage:    1248808,
	proto.Class_ClassPriest:  1323377,
	proto.Class_ClassWarlock: 1323392,
	proto.Class_ClassDruid:   1323400,
	proto.Class_ClassHunter:  1323410,
	proto.Class_ClassShaman:  1323418,
	proto.Class_ClassPaladin: 1323419,
}

// Spells the class's buff row does not name whose cast spends the charge all the same and crits on
// its direct hit. Each has a periodic part, which the buff leaves alone (EV-0149).
var revelationDirectHits = map[proto.Class]core.ClassFlags{
	proto.Class_ClassMage:    {Family: 3, Mask: [4]uint32{0x400001, 0x1000}}, // Fireball, Pyroblast, Frostfire Bolt
	proto.Class_ClassWarlock: {Family: 5, Mask: [4]uint32{0x4}},              // Immolate
	proto.Class_ClassDruid:   {Family: 7, Mask: [4]uint32{0x42}},             // Moonfire, Regrowth
	proto.Class_ClassShaman:  {Family: 11, Mask: [4]uint32{0x10000000}},      // Flame Shock
}

// Heals that roll nothing for Revelation: Judgement of Light's at each rank, which whoever strikes the
// judged target casts, where Seal of Light's heal rolls and the rows show no cause (EV-0149), and
// Vampiric Embrace's party heal 15290.
var revelationSkippedHeals = []int32{20267, 20341, 20342, 20343, 15290}

// Penance's heal bolt, which its channel casts each tick: it crits with the charge up and leaves it
// (EV-0149).
var revelationPenanceHeal = core.ClassFlags{Family: 6, Mask: [4]uint32{0, 0x10000}}

// The share of Revelation's chance a spell keeps for its level: all of it down to 12 levels under
// the wearer, none from 32 levels under, and a straight line between (EV-0149).
func revelationLevelShare(wearerLevel int32, spellLevel int32) float64 {
	return max(0, min(1, float64(spellLevel+32-wearerLevel)/20))
}

// The buff's row with the given class flags added to the spells its crit modifier names, which are
// the spells its charge is spent by.
func revelationSpentBy(row *spelldata.Spell, flags core.ClassFlags) *spelldata.Spell {
	widened := *row
	widened.Effects = slices.Clone(row.Effects)
	for i := range widened.Effects {
		if effect := &widened.Effects[i]; effect.Aura == dbcenums.A_ADD_FLAT_MODIFIER && !effect.ClassFlags.IsZero() {
			for word := range effect.ClassFlags.Mask {
				effect.ClassFlags.Mask[word] |= flags.Mask[word]
			}
		}
	}
	return &widened
}

func init() {

	// Enchant Weapon - Fiery Weapon
	// EffectID: 803, Proc SpellID: 13897
	// Permanently enchant a melee weapon to often strike for 40 additional fire damage.
	// https://www.wowhead.com/forever/spell=13898
	// The game files state no rate; the enchant's effect entry carries its procs per minute (EV-0130).
	shared.NewSpellDataDamageProc(shared.SpellDataProc{
		Name:           "Enchant Weapon - Fiery Weapon",
		EnchantID:      803,
		TriggerSpellID: 13897,
		BuffSpellID:    13897,
		IsWeaponProc:   true,
	}, nil)

	// Felsteel Shield Spike
	// EffectID: 2714, Proc SpellID: 29455
	// Permanently attaches a felsteel spike to your shield that deals 26 to 38 damage to attackers whose melee attacks you block.
	// https://www.wowhead.com/forever/spell=29455
	shared.NewProcDamageEffect(shared.ProcDamageEffect{
		EnchantID: 2714,
		SpellID:   29455,
		School:    core.SpellSchoolPhysical,
		MinDmg:    26,
		MaxDmg:    38,
		IsMelee:   true,
		Flags:     core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagNoOnDamageDealt,
		Trigger: core.ProcTrigger{
			Name:       "Felsteel Shield Spike",
			Callback:   core.CallbackOnSpellHitTaken,
			ProcMask:   core.ProcMaskMelee,
			Outcome:    core.OutcomeBlock,
			ProcChance: 1,
		},
	})

	// Mongoose
	// EffectID: 2673, Proc SpellID: 28093
	// PPM: 1, ICD: 0
	// Permanently enchant a Melee Weapon to occasionally increase Agility by 120 and attack speed slightly (2%).
	core.NewEnchantEffect(2673, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 15

		createMongooseAuras := func(slot proto.ItemSlot) *core.StatBuffAura {
			labelSuffix := core.Ternary(slot == proto.ItemSlot_ItemSlotMainHand, " (MH)", " (OH)")
			aura := character.NewTemporaryStatsAuraWrapped(
				"Lightning Speed"+labelSuffix,
				core.ActionID{SpellID: 28093}.WithHand(slot),
				stats.Stats{stats.Agility: 120},
				duration,
				func(aura *core.Aura) {
					aura.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
						character.MultiplyAttackSpeed(sim, 1.02)
					})
					aura.ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
						character.MultiplyAttackSpeed(sim, 1/1.02)
					})
				},
			)
			character.AddStatProcBuff(2673, aura, true, []proto.ItemSlot{slot})
			character.ItemSwap.RegisterWeaponEnchantBuff(aura.Aura, 2673)
			return aura
		}

		mhAuras := createMongooseAuras(proto.ItemSlot_ItemSlotMainHand)
		ohAuras := createMongooseAuras(proto.ItemSlot_ItemSlotOffHand)

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:         "Enchant Weapon - Mongoose",
			Callback:     core.CallbackOnSpellHitDealt,
			ActionID:     core.ActionID{SpellID: 28093},
			IsWeaponProc: true,
			DPM:          character.NewDynamicLegacyProcForEnchant(2673, 1.0, 0),
			Outcome:      core.OutcomeLanded,
			Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				core.Ternary(spell.IsOH(), ohAuras, mhAuras).Activate(sim)
			},
		})
	})

	// Executioner
	// EffectID: 3225, Proc SpellID: 42976
	// PPM: ?, ICD: 0
	// Permanently enchant a Melee Weapon to occasionally ignore 840 of your enemy's armor.  Requires a level 60 or higher item.
	core.NewEnchantEffect(3225, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 15

		aura := character.NewTemporaryStatsAura(
			"Executioner",
			core.ActionID{SpellID: 42976},
			stats.Stats{stats.ArmorPenetration: 840},
			duration,
		)
		character.AddStatProcBuff(3225, aura, true, core.AllMeleeWeaponSlots())
		character.ItemSwap.RegisterWeaponEnchantBuff(aura.Aura, 3225)

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:         "Enchant Weapon - Executioner",
			Callback:     core.CallbackOnSpellHitDealt,
			ActionID:     core.ActionID{SpellID: 28093},
			IsWeaponProc: true,
			DPM:          character.NewDynamicLegacyProcForEnchant(3225, 1.0, 0),
			Outcome:      core.OutcomeLanded,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
			},
		})
	})

	// Deathfrost
	// EffectID: 3273, Proc SpellID: 46579, Damage SpellID: 46579, Debuff SpellID: 46629
	// Proc Chance: 50%, ICD: 25s
	// Permanently enchant a weapon so your damaging spells and melee weapon hits occasionally inflict an additional 150 Frost damage
	// and reduce the target's melee, ranged, and casting speed by 15% for 8 sec.  Requires a level 60 or higher item.
	core.NewEnchantEffect(3273, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 8
		effect := 0.85

		debuffArray := character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return target.GetOrRegisterAura(core.Aura{
				Label:    "Deathfrost",
				Duration: duration,
				ActionID: core.ActionID{SpellID: 46629},
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					aura.Unit.MultiplyAttackSpeed(sim, effect)
					aura.Unit.MultiplyRangedSpeed(sim, effect)
					aura.Unit.MultiplyCastSpeed(sim, effect)
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					aura.Unit.MultiplyAttackSpeed(sim, 1/effect)
					aura.Unit.MultiplyRangedSpeed(sim, 1/effect)
					aura.Unit.MultiplyCastSpeed(sim, 1/effect)
				},
			})
		})

		dfSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 46579},
			SpellSchool: core.SpellSchoolFrost,
			DefenseType: core.DefenseTypeMagic,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagProc,
			ProcMask:    core.ProcMaskSpellDamage,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, 150, spell.OutcomeMagicCrit)
			},
		})

		// Enchant 3273 carries two effects: a combat spell (Effect 1) that rolls on melee hits, and an
		// Equip aura 46662 (Effect 3) that rolls on spell damage, direct and periodic (ProcTypeMask
		// 2424832), without Can Proc From Procs. Both cast 46579. The 25 s lockout is 46662's proc
		// recovery; the combat half has none in the DBC, so the two halves share one here rather than
		// doubling the rate the old single trigger had.
		//
		// The roll and the lockout are taken in ExtraCondition, which runs synchronously in the
		// callback. The handler runs a batch window later, and hits landing on the same timestamp
		// (both hands, extra attacks, multiple targets) would all pass a lockout consumed there.
		lockout := core.Cooldown{
			Timer:    character.NewTimer(),
			Duration: time.Second * 25,
		}
		meleeDPM := character.NewDynamicLegacyProcForEnchant(3273, 0, 0.5)
		handler := func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			debuffArray.Get(result.Target).Activate(sim)
			dfSpell.Cast(sim, result.Target)
		}

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:         "Enchant Weapon - Deathfrost (Melee)",
			Callback:     core.CallbackOnSpellHitDealt,
			ActionID:     core.ActionID{SpellID: 46579},
			IsWeaponProc: true,
			Outcome:      core.OutcomeLanded,
			ExtraCondition: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
				if !lockout.IsReady(sim) || !meleeDPM.Proc(sim, spell.ProcMask, "Deathfrost (Melee)") {
					return false
				}
				lockout.Use(sim)
				return true
			},
			Handler: handler,
		})

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "Enchant Weapon - Deathfrost (Spell)",
			Callback: core.CallbackOnSpellHitDealt | core.CallbackOnPeriodicDamageDealt,
			ActionID: core.ActionID{SpellID: 46662},
			ProcMask: core.ProcMaskSpellDamage,
			Outcome:  core.OutcomeLanded,

			RequireDamageComponent: true,

			ExtraCondition: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
				if !lockout.IsReady(sim) || !sim.Proc(0.5, "Deathfrost (Spell)") {
					return false
				}
				lockout.Use(sim)
				return true
			},
			Handler: handler,
		})
	})

	// Enchant Weapon - Revelation
	// EffectID: 8217, Proc SpellID: 1248806
	// The proc spell names no buff: each class with spells to crit has a Revelation of its own, and the
	// wearer gets its class's, with that row's class flags and charge. The game files state no rate,
	// only that a direct spell or heal landing without a crit rolls it, and that the chance falls as
	// the wearer's crit chance rises; the enchant's effect entry carries the chance. Only a hit that
	// deals damage or heals rolls it: a landing for nothing, such as a channel's start or an aura's
	// application, does not. The spell a channel casts each tick (Blizzard's) rolls it too.
	core.NewEnchantEffect(8217, func(agent core.Agent) {
		character := agent.GetCharacter()
		buffID, ok := revelationBuffs[character.Class]
		if !ok {
			return
		}

		row := spelldata.MustFind(buffID)
		buff := character.RegisterAura(spelldata.AuraConfig(row))
		spelldata.ParseEffects(character, buff, row)
		character.ItemSwap.RegisterWeaponEnchantBuff(buff, 8217)

		// The charge goes as a cast the buff reaches completes, on every class, and a spell still in
		// flight keeps the crit it rolled. A melee or ranged attack the mask names, a paladin's
		// judgement or Hammer of Wrath, crits and leaves the charge (EV-0028, EV-0149). A channel's cast
		// spends nothing; a tick of the channel takes the charge, as the listeners below set out.
		// TODO: Holy Shock's and Light's Vigil's heals crit here on every cast with the buff up, and
		// in the logs mostly do not, landing after the charge is gone (EV-0149).
		// TODO: A charge gained while Volley or Penance's heal runs, whose ticks take no charge, is still
		// spent as the channel ends (R-0048). EV-0149 records no such charge.
		isAttack := func(spell *core.Spell) bool {
			return spell.DefenseType == core.DefenseTypeMelee || spell.DefenseType == core.DefenseTypeRanged
		}
		spends := row
		direct, critsDirectHits := revelationDirectHits[character.Class]
		if critsDirectHits {
			spends = revelationSpentBy(row, direct)
		}
		spender := spelldata.AttachChargeSpender(character, buff, spends, "Revelation Charge",
			func(_ *core.Character, trigger *core.ProcTrigger) {
				trigger.ProcMask = core.ProcMaskUnknown
				trigger.Callback = core.CallbackOnCastComplete
				trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
					return !isAttack(spell) && !(spell.Flags.Matches(core.SpellFlagChanneled) && character.IsChanneling())
				}
			})

		spendCharge := func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			if buff.IsActive() {
				buff.RemoveStack(sim)
			}
		}

		// A channel the mask names whose ticks cast a spell of their own (Blizzard, Hurricane, Hellfire,
		// Rain of Fire) spends the charge on the first tick that crits from it. Every target that tick
		// strikes crits, and the charge goes a spell batch window later, so the ticks after it crit at
		// their own chance. Arcane Missiles' missiles roll their crit as they launch: the first launched
		// after the charge comes crits, one already in flight keeps its own roll, and the first to land
		// with a crit spends the charge. Penance's heal bolts leave it (EV-0149).
		firstTickHit, _ := core.OncePerCast()
		buff.AttachProcTriggerCallback(&character.Unit, core.ProcTrigger{
			Name:               "Revelation Charge - Channel Tick",
			Callback:           core.CallbackOnSpellHitDealt | core.CallbackOnHealDealt,
			ClassFlags:         spender.ClassFlags,
			CanProcFromProcs:   true,
			Outcome:            core.OutcomeCrit,
			RequireDamageDealt: true,
			ExtraCondition: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
				channel := character.ChanneledDot
				return channel != nil && channel.Spell != spell && channel.Spell.MatchesFlags(spender.ClassFlags) &&
					!spell.Flags.Matches(core.SpellFlagChanneled) && !isAttack(spell) &&
					!spell.MatchesFlags(revelationPenanceHeal) && firstTickHit(sim, spell)
			},
			Handler: spendCharge,
		})

		// A channel the mask names whose ticks are its own periodic damage (Mind Flay, Wrack, Drain Soul,
		// Drain Life) spends the charge on its first tick, which crits from it, and the ticks after it
		// crit at their own chance. Starshards is such a channel and is not counted. Another spell's
		// periodic ticks leave the charge (EV-0149).
		buff.AttachProcTriggerCallback(&character.Unit, core.ProcTrigger{
			Name:       "Revelation Charge - Periodic Channel Tick",
			Callback:   core.CallbackOnPeriodicDamageDealt | core.CallbackOnPeriodicHealDealt,
			ClassFlags: spender.ClassFlags,
			Outcome:    core.OutcomeCrit,
			ExtraCondition: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
				channel := character.ChanneledDot
				return channel != nil && channel.Spell == spell
			},
			Handler: spendCharge,
		})

		// The direct hit of a spell in revelationDirectHits crits for the one cast that spends the
		// charge, and its ticks never do.
		if critsDirectHits {
			crit := row.Effect(dbcenums.A_ADD_FLAT_MODIFIER, int32(dbcenums.SPELLMOD_CRITICAL_CHANCE)).BaseValue()
			var critting *core.Spell
			stopCritting := func() {
				if critting != nil {
					critting.BonusCritPercent -= crit
					critting = nil
				}
			}
			buff.OnApplyEffects = func(_ *core.Aura, _ *core.Simulation, _ *core.Unit, spell *core.Spell) {
				if spell.Unit != &character.Unit || !spell.MatchesFlags(direct) || spell.Flags.Matches(core.SpellFlagProc) {
					return
				}
				stopCritting()
				if buff.GetStacks() > 0 {
					critting = spell
					spell.BonusCritPercent += crit
				}
			}
			buff.ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) { stopCritting() })
		}

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(1248806), func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			buff.Activate(sim)
			buff.SetStacks(sim, buff.MaxStacks)
		}, spelldata.Chance(enchantProcChance(8217, 1248806))) // TODO: Model how the chance falls with the wearer's crit from gear and intellect (EV-0149).
		trigger.Name = "Enchant Weapon - Revelation"
		trigger.Outcome = core.OutcomeLanded
		trigger.RequireDamageDealt = true
		// The row's listener hears direct spells and heals. A melee or ranged attack rolls nothing, even
		// one the buff's class flags name, and neither do Judgement of Light's and Vampiric Embrace's
		// heals. A spell far under the wearer's level rolls a share of the chance, and a spell with no
		// SpellLevels row rolls nothing (EV-0149). A spell the store has no row for, such as an item's
		// spell registered under its item id, reads level 0 and keeps none.
		// TODO: Read an item spell's own row, so one near the wearer's level rolls.
		// TODO: A hunter's trap rolls nothing in the logs and rolls here (EV-0149).
		// One hit event rolls once, however many targets it strikes: the first result of a spell at a
		// moment that lands for damage or heals is the event's one chance. When that result crits, the
		// event rolls nothing, whatever its other targets do. Lightning Overload's bolt rolls apart from
		// the bolt it copies, which shares its spell id, even when both land at one moment (EV-0149).
		shares := map[*core.Spell]float64{}
		var lastSpell *core.Spell
		var lastAt time.Duration
		trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
			if isAttack(spell) {
				return false
			}
			first := spell != lastSpell || sim.CurrentTime != lastAt
			lastSpell, lastAt = spell, sim.CurrentTime
			// A landing while the buff is up rolls too and refreshes it.
			if !first || result.DidCrit() {
				return false
			}
			share, ok := shares[spell]
			if !ok {
				if row := spelldata.Find(spell.SpellID); !row.NoSpellLevels && !slices.Contains(revelationSkippedHeals, spell.SpellID) {
					share = revelationLevelShare(character.Level, int32(row.SpellLevel))
				}
				shares[spell] = share
			}
			return share >= 1 || (share > 0 && sim.Proc(share, "Revelation Level Share"))
		}
		character.ItemSwap.RegisterEnchantProc(8217, character.MakeProcTriggerAura(trigger))
	})

	// Scopes
	core.NewEnchantEffect(2722, func(agent core.Agent) {
		character := agent.GetCharacter()
		ranged := character.AutoAttacks.Ranged()
		ranged.BaseDamageMin += 10
		ranged.BaseDamageMax += 10
	})

	core.NewEnchantEffect(2723, func(agent core.Agent) {
		character := agent.GetCharacter()
		ranged := character.AutoAttacks.Ranged()
		ranged.BaseDamageMin += 12
		ranged.BaseDamageMax += 12
	})

	movementSpeedEnchants := []int32{
		911,  // Enchant Boots - Minor Speed
		2939, // Enchant Boots - Cat's Swiftness
		2940, // Enchant Boots - Boar's Speed
	}

	for _, enchantID := range movementSpeedEnchants {
		core.NewEnchantEffect(enchantID, func(agent core.Agent) {
			character := agent.GetCharacter()
			// The game files state the run speed as A_MOD_INCREASE_SPEED, as Strider Kick and Dash do,
			// and only the highest of those counts, so it bids in the active category with them.
			aura := core.MakePermanent(character.GetOrRegisterAura(core.Aura{
				Label:    "Minor Run Speed",
				ActionID: core.ActionID{SpellID: 13889},
			}))
			aura.NewActiveMovementSpeedEffect(spelldata.MustFind(13889).EffectN(1).Percent())

			character.ItemSwap.RegisterEnchantProc(enchantID, aura)
		})
	}
}

// The chance per hit the enchant's effect entry carries for the spell.
func enchantProcChance(effectID int32, spellID int32) float64 {
	for _, effect := range core.GetEnchantByEffectID(effectID).EnchantEffects {
		if effect.BuffId == spellID && effect.GetProc().GetProcChance() > 0 {
			return effect.GetProc().GetProcChance()
		}
	}
	panic(fmt.Sprintf("enchant %d states no proc chance for spell %d", effectID, spellID))
}
