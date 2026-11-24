//go:build with_db

package forever_test

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

const revelationEnchant = 8217

func init() {
	core.RegisterAgentFactory(
		proto.Player_BalanceDruid{},
		proto.Spec_SpecBalanceDruid,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			return &testDruid{Character: *character}
		},
		func(player *proto.Player, spec any) {
			player.Spec = spec.(*proto.Player_BalanceDruid)
		},
	)
	core.RegisterAgentFactory(
		proto.Player_Hunter{},
		proto.Spec_SpecHunter,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			return &testHunter{Character: *character}
		},
		func(player *proto.Player, spec any) {
			player.Spec = spec.(*proto.Player_Hunter)
		},
	)
	core.RegisterAgentFactory(
		proto.Player_Mage{},
		proto.Spec_SpecMage,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			return &testMage{Character: *character}
		},
		func(player *proto.Player, spec any) {
			player.Spec = spec.(*proto.Player_Mage)
		},
	)
}

// A caster with three druid spells at the wearer's level: Wrath, which Revelation's druid buff 1323400
// names (family 7 bit 0), Moonfire, which it does not name and reaches all the same (bit 1), and one it
// does not reach (bit 4). Hurricane is a three-tick channel whose ticks cast a spell of their own on
// every enemy, a proc that fires no cast listener, both under Wrath's flags. Drain is a channel under
// Wrath's flags whose own hit lands before its dot starts and whose ticks are its own periodic damage,
// which rolls to crit. Bolt, under Wrath's flags, rolls at the cast and lands a second later.
type testDruid struct {
	core.Character
	wrath, moonfire, outside, hurricane, drain, bolt *core.Spell
	last                                             *core.SpellResult
	tickCrits, drainCrits                            []bool
}

func (d *testDruid) GetCharacter() *core.Character       { return &d.Character }
func (d *testDruid) ApplyTalents()                       {}
func (d *testDruid) Reset(_ *core.Simulation)            {}
func (d *testDruid) OnGCDReady(_ *core.Simulation)       {}
func (d *testDruid) OnEncounterStart(_ *core.Simulation) {}

func (d *testDruid) Initialize() {
	nuke := func(id int32, bit uint) *core.Spell {
		return d.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{SpellID: id},
			SpellSchool:      core.SpellSchoolNature,
			DefenseType:      core.DefenseTypeMagic,
			ProcMask:         core.ProcMaskSpellDamage,
			ClassFlags:       core.ClassFlags{Family: 7, Mask: [4]uint32{1 << bit}},
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				d.last = spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeMagicHitAndCrit)
			},
		})
	}
	d.wrath = nuke(9912, 0)
	d.moonfire = nuke(9835, 1)
	d.outside = nuke(24977, 4)

	wrathFlags := core.ClassFlags{Family: 7, Mask: [4]uint32{1}}
	d.bolt = d.RegisterSpell(core.SpellConfig{
		ActionID:           core.ActionID{SpellID: 25298},
		SpellSchool:        core.SpellSchoolArcane,
		DefenseType:        core.DefenseTypeMagic,
		ProcMask:           core.ProcMaskSpellDamage,
		ClassFlags:         wrathFlags,
		MissileMinDuration: time.Second,
		DamageMultiplier:   1,
		ThreatMultiplier:   1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			d.last = spell.CalcDamage(sim, target, 100, spell.OutcomeMagicCrit)
			spell.DealDamageAfterTravel(sim, d.last)
		},
	})
	tick := d.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 17402, Tag: 1},
		SpellSchool:      core.SpellSchoolNature,
		DefenseType:      core.DefenseTypeMagic,
		ProcMask:         core.ProcMaskSpellDamage,
		Flags:            core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,
		ClassFlags:       wrathFlags,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			for _, target := range sim.Encounter.ActiveTargetUnits {
				result := spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeMagicCrit)
				d.tickCrits = append(d.tickCrits, result.DidCrit())
			}
		},
	})
	channel := func(id int32, label string, ownHit bool, onTick core.OnTick) *core.Spell {
		return d.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: id},
			SpellSchool: core.SpellSchoolNature,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellDamage,
			Flags:       core.SpellFlagChanneled,
			ClassFlags:  wrathFlags,
			Dot: core.DotConfig{
				Aura:          core.Aura{Label: label},
				NumberOfTicks: 3,
				TickLength:    time.Second,
				OnTick:        onTick,
			},
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				if ownHit {
					spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
				}
				spell.Dot(target).Apply(sim)
			},
		})
	}
	d.hurricane = channel(17402, "Hurricane", false, func(sim *core.Simulation, target *core.Unit, _ *core.Dot) {
		tick.Cast(sim, target)
	})
	d.drain = channel(17401, "Drain", true, func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		result := dot.Spell.CalcAndDealPeriodicDamage(sim, target, 100, dot.Spell.OutcomeTickMagicCrit)
		d.drainCrits = append(d.drainCrits, result.DidCrit())
	})
}

// A hunter with two ranged shots: Arcane Shot, which Revelation's hunter buff 1323410 names (family 9
// bit 11), and Aimed Shot, which it does not (bit 17). Volley's wave is a ranged attack with a spell
// proc mask, so only its defense type keeps it from rolling.
type testHunter struct {
	core.Character
	arcaneShot, aimedShot, wave *core.Spell
	last                        *core.SpellResult
}

func (h *testHunter) GetCharacter() *core.Character       { return &h.Character }
func (h *testHunter) ApplyTalents()                       {}
func (h *testHunter) Reset(_ *core.Simulation)            {}
func (h *testHunter) OnGCDReady(_ *core.Simulation)       {}
func (h *testHunter) OnEncounterStart(_ *core.Simulation) {}

func (h *testHunter) Initialize() {
	shot := func(id int32, bit uint) *core.Spell {
		return h.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{SpellID: id},
			SpellSchool:      core.SpellSchoolArcane,
			DefenseType:      core.DefenseTypeRanged,
			ProcMask:         core.ProcMaskRangedSpecial,
			ClassFlags:       core.ClassFlags{Family: 9, Mask: [4]uint32{1 << bit}},
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				h.last = spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeRangedHitAndCrit)
			},
		})
	}
	h.arcaneShot = shot(14287, 11)
	h.aimedShot = shot(20904, 17)
	h.wave = h.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 14295},
		SpellSchool:      core.SpellSchoolArcane,
		DefenseType:      core.DefenseTypeRanged,
		ProcMask:         core.ProcMaskSpellDamage,
		ClassFlags:       core.ClassFlags{Family: 9, Mask: [4]uint32{1 << 13}},
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			h.last = spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeMagicHit)
		},
	})
}

// A mage with spells Revelation's mage buff 1248808 names: Fire Blast (family 3 bit 1), Frostbolt
// (bit 5), which rolls at the cast and lands a second later, with its first rank beside it, Arcane
// Explosion (bit 12), which hits every target, and Arcane Missiles (bit 11), a three-tick channel, with
// a second channeled cast under its flags that starts no channel. Fireball (bit 0) is not named. None
// rolls to miss, and each keeps the crits of its landings.
type testMage struct {
	core.Character
	fireBlast, frostbolt, firstRank, fireball, explosion, missiles, missile, stalled *core.Spell
	crits                                                                            map[*core.Spell][]bool
}

func (m *testMage) GetCharacter() *core.Character       { return &m.Character }
func (m *testMage) ApplyTalents()                       {}
func (m *testMage) Reset(_ *core.Simulation)            {}
func (m *testMage) OnGCDReady(_ *core.Simulation)       {}
func (m *testMage) OnEncounterStart(_ *core.Simulation) {}

func (m *testMage) Initialize() {
	m.crits = map[*core.Spell][]bool{}
	land := func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		spell.DealDamage(sim, result)
		m.crits[spell] = append(m.crits[spell], result.DidCrit())
	}
	hit := func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		land(sim, spell, spell.CalcDamage(sim, target, 100, spell.OutcomeMagicCrit))
	}
	spell := func(id core.ActionID, bit uint, config core.SpellConfig) *core.Spell {
		config.ActionID = id
		config.SpellSchool = core.SpellSchoolFire
		config.DefenseType = core.DefenseTypeMagic
		config.ProcMask = core.ProcMaskSpellDamage
		config.ClassFlags = core.ClassFlags{Family: 3, Mask: [4]uint32{1 << bit}}
		config.DamageMultiplier = 1
		config.ThreatMultiplier = 1
		return m.RegisterSpell(config)
	}
	m.fireBlast = spell(core.ActionID{SpellID: 10199}, 1, core.SpellConfig{ApplyEffects: hit})
	m.firstRank = spell(core.ActionID{SpellID: 116}, 5, core.SpellConfig{ApplyEffects: hit})
	m.fireball = spell(core.ActionID{SpellID: 25306}, 0, core.SpellConfig{ApplyEffects: hit})
	m.frostbolt = spell(core.ActionID{SpellID: 25304}, 5, core.SpellConfig{
		MissileMinDuration: time.Second,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, 100, spell.OutcomeMagicCrit)
			spell.DealCastResult(sim, result)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) { land(sim, spell, result) })
		},
	})
	m.explosion = spell(core.ActionID{SpellID: 10202}, 12, core.SpellConfig{
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			for _, result := range spell.CalcAndDealAoeDamage(sim, 100, spell.OutcomeMagicCrit) {
				m.crits[spell] = append(m.crits[spell], result.DidCrit())
			}
		},
	})
	m.missile = spell(core.ActionID{SpellID: 25345, Tag: 1}, 11, core.SpellConfig{ApplyEffects: hit})
	m.stalled = spell(core.ActionID{SpellID: 25345, Tag: 2}, 11, core.SpellConfig{Flags: core.SpellFlagChanneled})
	m.missiles = spell(core.ActionID{SpellID: 25345}, 11, core.SpellConfig{
		Flags: core.SpellFlagChanneled,
		Dot: core.DotConfig{
			Aura:          core.Aura{Label: "Arcane Missiles"},
			NumberOfTicks: 3,
			TickLength:    time.Second,
			OnTick: func(sim *core.Simulation, target *core.Unit, _ *core.Dot) {
				m.missile.Cast(sim, target)
			},
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Dot(target).Apply(sim)
		},
	})
}

func revelationWearer(t *testing.T, class proto.Class, spec any) (*core.Simulation, *core.Character) {
	return revelationWearerFacing(t, class, spec, 1)
}

func revelationWearerFacing(t *testing.T, class proto.Class, spec any, enemies int) (*core.Simulation, *core.Character) {
	t.Helper()
	targets := make([]*proto.Target, enemies)
	for i := range targets {
		targets[i] = &proto.Target{Name: "target", Level: 60}
	}
	items := weaponSlots(1613, 0)
	items[proto.ItemSlot_ItemSlotMainHand].Enchant = revelationEnchant
	player := &proto.Player{
		Name: "Wearer", Class: class, Race: proto.Race_RaceTauren,
		Equipment: &proto.EquipmentSpec{Items: items},
		Buffs:     &proto.IndividualBuffs{},
		Rotation:  &proto.APLRotation{},
	}
	core.WithSpec(player, spec)

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{player},
			Buffs:   &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: targets, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

func chargeRevelation(sim *core.Simulation, buff *core.Aura) {
	buff.Activate(sim)
	buff.SetStacks(sim, buff.MaxStacks)
}

// Enchant Weapon - Revelation 8217 gives a druid its class's buff 1323400 on a direct spell that lands
// without a crit: a charge of crit for the row's duration on the spells its mask names, spent by the
// next of them and not by a spell the buff does not reach. Moonfire, which the mask does not name, crits
// on its direct hit and spends the charge too (EV-0149). A warrior has no Revelation buff.
func TestRevelationGrantsTheWearersClassBuff(t *testing.T) {
	sim, character := revelationWearer(t, proto.Class_ClassDruid, &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{}})
	druid := sim.Raid.Parties[0].Players[0].(*testDruid)
	target := character.CurrentTarget

	buff := character.GetAuraByID(core.ActionID{SpellID: 1323400})
	if buff == nil {
		t.Fatal("no Revelation buff 1323400 on the druid")
	}
	row := spelldata.MustFind(1323400)
	if buff.Duration != row.Duration() || buff.MaxStacks != int32(row.ProcCharges) {
		t.Errorf("buff lasts %v with %d charges, want 1323400's %v and %d", buff.Duration, buff.MaxStacks, row.Duration(), row.ProcCharges)
	}

	// A proc's handler runs one spell batch window after the hit, so each cast is followed by a wait past
	// it, which also puts each cast at a moment of its own.
	castAndStep := func(spell *core.Spell) *core.SpellResult {
		spell.Cast(sim, target)
		stepTo(t, sim, sim.CurrentTime+2*core.SpellBatchWindow)
		return druid.last
	}

	for casts := 0; !buff.IsActive(); casts++ {
		if casts > 300 {
			t.Fatal("300 casts of a direct spell and Revelation never procced")
		}
		castAndStep(druid.outside)
	}

	castAndStep(druid.outside)
	if !buff.IsActive() {
		t.Fatal("a spell the buff does not reach spent its charge")
	}
	if result := castAndStep(druid.wrath); !result.DidCrit() {
		t.Error("Wrath with Revelation up did not crit")
	}
	if buff.IsActive() {
		t.Error("Wrath did not spend Revelation's charge")
	}

	moonfireCrit := druid.moonfire.SpellCritChance(target)
	chargeRevelation(sim, buff)
	if got := druid.moonfire.SpellCritChance(target); got != moonfireCrit {
		t.Errorf("Moonfire's crit chance is %v with Revelation up and %v without, and its ticks roll that chance", got, moonfireCrit)
	}
	result := castAndStep(druid.moonfire)
	for tries := 0; !result.Landed() && tries < 20; tries++ {
		chargeRevelation(sim, buff)
		result = castAndStep(druid.moonfire)
	}
	if !result.DidCrit() {
		t.Errorf("Moonfire's direct hit with Revelation up did not crit: outcome %v", result.Outcome)
	}
	if buff.IsActive() {
		t.Error("Moonfire did not spend Revelation's charge")
	}
	if got := druid.moonfire.SpellCritChance(target); got != moonfireCrit {
		t.Errorf("Moonfire's crit chance is %v after the cast that spent the charge, want %v", got, moonfireCrit)
	}

	_, warrior := revelationWearer(t, proto.Class_ClassWarrior, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{
		Options: &proto.DpsWarrior_Options{ClassOptions: &proto.WarriorOptions{}}}})
	if warrior.GetAura("Revelation") != nil {
		t.Error("a warrior has a Revelation buff")
	}
}

// A channel whose ticks cast a spell of their own spends the charge on its first tick: the channel's
// cast leaves it, every enemy that tick strikes crits, and the ticks after it crit at their own chance.
// The ticks roll Revelation too (EV-0149).
func TestRevelationChannelSpendsTheChargeOnItsFirstTick(t *testing.T) {
	sim, character := revelationWearerFacing(t, proto.Class_ClassDruid, &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{}}, 3)
	druid := sim.Raid.Parties[0].Players[0].(*testDruid)
	target := character.CurrentTarget
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -character.GetStat(stats.SpellCritPercent)})
	buff := character.GetAuraByID(core.ActionID{SpellID: 1323400})
	channel := druid.hurricane.Dot(target)

	procced := false
	for casts := 0; !procced; casts++ {
		if casts > 40 {
			t.Fatal("40 channels and no tick procced Revelation")
		}
		druid.hurricane.Cast(sim, target)
		stepWhile(t, sim, func() bool {
			procced = procced || buff.IsActive()
			return channel.IsActive()
		})
	}
	stepTo(t, sim, sim.CurrentTime+time.Second)
	buff.Deactivate(sim)

	druid.tickCrits = nil
	chargeRevelation(sim, buff)
	druid.hurricane.Cast(sim, target)
	if !channel.IsActive() || !buff.IsActive() {
		t.Fatal("Hurricane started and Revelation's charge is gone before its first tick")
	}
	stepWhile(t, sim, func() bool { return len(druid.tickCrits) < 3 })
	if first := druid.tickCrits; len(first) != 3 || slices.Contains(first, false) {
		t.Errorf("the first tick crit %v, want every one of the 3 targets", first)
	}
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if buff.IsActive() {
		t.Fatal("the first tick landed and Revelation's charge is still up")
	}
	// A later tick that lands without a crit may proc a new charge, which is taken away before the next.
	stepWhile(t, sim, func() bool {
		buff.Deactivate(sim)
		return channel.IsActive()
	})
	if later := druid.tickCrits[3:]; len(later) != 6 || slices.Contains(later, true) {
		t.Errorf("the ticks after the first crit %v, want 6 hits and none crit", later)
	}
}

// A channel whose ticks are its own periodic damage spends nothing: every tick crits with the charge up,
// the charge outlasts the channel, run out or cut short, and the next cast the mask names crits and
// spends it. With the charge down the ticks crit at their own chance (EV-0149).
func TestRevelationLastsPastAChannelOfPeriodicTicks(t *testing.T) {
	sim, character := revelationWearer(t, proto.Class_ClassDruid, &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{}})
	druid := sim.Raid.Parties[0].Players[0].(*testDruid)
	target := character.CurrentTarget
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -character.GetStat(stats.SpellCritPercent)})
	buff := character.GetAuraByID(core.ActionID{SpellID: 1323400})
	drain := druid.drain.Dot(target)

	druid.drainCrits = nil
	druid.drain.Cast(sim, target)
	stepWhile(t, sim, drain.IsActive)
	if got := druid.drainCrits; len(got) != 3 || slices.Contains(got, true) || buff.IsActive() {
		t.Fatalf("with no crit chance and Revelation down the ticks crit %v, want 3 and none crit, and nothing procced", got)
	}

	druid.drainCrits = nil
	chargeRevelation(sim, buff)
	druid.drain.Cast(sim, target)
	if !drain.IsActive() || !buff.IsActive() {
		t.Fatal("Drain started and Revelation's charge is gone")
	}
	stepWhile(t, sim, drain.IsActive)
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if got := druid.drainCrits; len(got) != 3 || slices.Contains(got, false) {
		t.Errorf("the ticks with Revelation up crit %v, want all three", got)
	}
	if !buff.IsActive() {
		t.Fatal("Drain ran out and Revelation's charge is gone")
	}

	druid.drain.Cast(sim, target)
	drain.Deactivate(sim)
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if !buff.IsActive() {
		t.Fatal("Drain was cut short and Revelation's charge is gone")
	}

	druid.wrath.Cast(sim, target)
	for tries := 0; !druid.last.Landed() && tries < 20; tries++ {
		chargeRevelation(sim, buff)
		druid.wrath.Cast(sim, target)
	}
	if !druid.last.DidCrit() {
		t.Error("Wrath cast after the channel with Revelation up did not crit")
	}
	if buff.IsActive() {
		t.Error("Wrath did not spend the charge Drain left")
	}
}

// A ranged attack rolls nothing, so no shot procs a hunter's buff 1323410, not even Arcane Shot, which
// its mask names, nor a wave with a spell's proc mask. With the buff up, a shot the mask names crits and
// leaves the charge, as a paladin's judgement does (EV-0149).
func TestRevelationHearsNoRangedAttack(t *testing.T) {
	sim, character := revelationWearer(t, proto.Class_ClassHunter, &proto.Player_Hunter{Hunter: &proto.Hunter{}})
	hunter := sim.Raid.Parties[0].Players[0].(*testHunter)
	target := character.CurrentTarget

	buff := character.GetAuraByID(core.ActionID{SpellID: 1323410})
	if buff == nil {
		t.Fatal("no Revelation buff 1323410 on the hunter")
	}

	castAndStep := func(spell *core.Spell) *core.SpellResult {
		spell.Cast(sim, target)
		stepTo(t, sim, sim.CurrentTime+2*core.SpellBatchWindow)
		return hunter.last
	}

	for range 300 {
		castAndStep(hunter.aimedShot)
		castAndStep(hunter.arcaneShot)
		castAndStep(hunter.wave)
	}
	if buff.IsActive() {
		t.Fatal("a ranged attack procced Revelation")
	}

	chargeRevelation(sim, buff)
	for range 5 {
		if result := castAndStep(hunter.arcaneShot); !result.DidCrit() {
			t.Fatal("Arcane Shot with Revelation up did not crit")
		}
	}
	if !buff.IsActive() {
		t.Error("Arcane Shot, a ranged attack, spent Revelation's charge")
	}
}

// An item's magic proc rolls Revelation: Crest of Retribution 13375's 17496 strikes back at a block, and
// its landings proc the paladin's buff 1323419 (EV-0149).
func TestRevelationRollsOnAnItemsMagicProc(t *testing.T) {
	items := weaponSlots(2244, 13375)
	items[proto.ItemSlot_ItemSlotMainHand].Enchant = revelationEnchant
	sim, character := itemWearerWithEquipment(t, proto.Class_ClassPaladin, &proto.EquipmentSpec{Items: items})
	enemy := sim.Encounter.ActiveTargetUnits[0]
	buff := character.GetAuraByID(core.ActionID{SpellID: 1323419})
	crest := character.GetSpell(core.ActionID{SpellID: 17496})
	if buff == nil || crest == nil {
		t.Fatal("the paladin's Revelation buff or Crest of Retribution's strike is not registered")
	}

	procs := revelationProcsOnBlocks(t, sim, character, buff, 400, core.SpellBatchWindow)
	if landed := crest.SpellMetrics[enemy.UnitIndex].Hits; landed < 300 || procs == 0 {
		t.Errorf("400 blocks landed %d Crest of Retribution strikes and procced Revelation %d times, want most of them to land and some to proc", landed, procs)
	}
}

// An item's heal rolls Revelation too: Krol'dok Battlegear's 4-piece casts Krol'dok Resolve 1324296 on the
// wearer at a block, and its heals proc the paladin's buff 1323419 (EV-0149).
func TestRevelationRollsOnAnItemsHeal(t *testing.T) {
	equipment := setPieces(kroldokPieces(4))
	equipment.Items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 2244, Enchant: revelationEnchant}
	sim, character := itemWearerWithEquipment(t, proto.Class_ClassPaladin, equipment)
	buff := character.GetAuraByID(core.ActionID{SpellID: 1323419})
	heal := character.GetSpell(core.ActionID{SpellID: 1324296})
	if buff == nil || heal == nil {
		t.Fatal("the paladin's Revelation buff or Krol'dok Resolve's heal is not registered")
	}

	procs := revelationProcsOnBlocks(t, sim, character, buff, 150, spelldata.MustFind(1324298).ICD())
	if heals := heal.SpellMetrics[character.UnitIndex].Casts; heals != 150 || procs == 0 {
		t.Errorf("150 blocks a cooldown apart healed %d times and procced Revelation %d times, want a heal each and some procs", heals, procs)
	}
}

// The blocks out of n that leave the buff up, each a blocked melee swing with the buff down and a wait after.
func revelationProcsOnBlocks(t *testing.T, sim *core.Simulation, character *core.Character, buff *core.Aura, n int, wait time.Duration) int {
	t.Helper()
	enemy := sim.Encounter.ActiveTargetUnits[0]
	procs := 0
	for range n {
		buff.Deactivate(sim)
		takeOutcome(t, sim, character, meleeSwing(enemy), core.OutcomeHit|core.OutcomeBlock, 50)
		stepTo(t, sim, sim.CurrentTime+wait)
		if buff.IsActive() {
			procs++
		}
	}
	return procs
}

// A mage's buff 1248808 spends its charge as the cast it reaches completes. Every target of an area
// cast crits and the next cast does not; a missile rolled at the cast keeps its crit in flight, a
// spell cast while it flies gets none, and its landing spends no charge. A channel's first missile
// crits and spends the charge, and a channeled cast that starts no channel spends it as it completes
// (EV-0149).
func TestRevelationMageSpendsTheChargeAtTheCast(t *testing.T) {
	sim, character := revelationWearerFacing(t, proto.Class_ClassMage, &proto.Player_Mage{Mage: &proto.Mage{}}, 3)
	mage := sim.Raid.Parties[0].Players[0].(*testMage)
	target := character.CurrentTarget
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -character.GetStat(stats.SpellCritPercent)})

	buff := character.GetAuraByID(core.ActionID{SpellID: 1248808})
	if buff == nil {
		t.Fatal("no Revelation buff 1248808 on the mage")
	}
	crits := func(spell *core.Spell) []bool {
		got := mage.crits[spell]
		mage.crits[spell] = nil
		return got
	}
	// Lets the procs of the last landings run and takes away any charge they gave.
	settle := func() {
		stepTo(t, sim, sim.CurrentTime+time.Second)
		buff.Deactivate(sim)
	}

	chargeRevelation(sim, buff)
	mage.explosion.Cast(sim, target)
	if got := crits(mage.explosion); len(got) != 3 || slices.Contains(got, false) {
		t.Errorf("Arcane Explosion with Revelation up crit %v, want every one of the 3 targets", got)
	}
	if buff.IsActive() {
		t.Error("Arcane Explosion completed and Revelation's charge is still up")
	}
	mage.explosion.Cast(sim, target)
	if got := crits(mage.explosion); len(got) != 3 || slices.Contains(got, true) {
		t.Errorf("a second Arcane Explosion in the same moment crit %v, want 3 targets hit and none crit", got)
	}
	settle()

	chargeRevelation(sim, buff)
	mage.frostbolt.Cast(sim, target)
	if buff.IsActive() {
		t.Error("Frostbolt completed and Revelation's charge is still up while it flies")
	}
	mage.fireBlast.Cast(sim, target)
	if got := crits(mage.fireBlast); len(got) != 1 || got[0] {
		t.Errorf("Fire Blast cast while Frostbolt flies crit %v, want one hit without Revelation's crit", got)
	}
	chargeRevelation(sim, buff)
	stepTo(t, sim, sim.CurrentTime+time.Second+100*time.Millisecond)
	if got := crits(mage.frostbolt); len(got) != 1 || !got[0] {
		t.Errorf("Frostbolt landed with crits %v, want the crit it rolled at the cast", got)
	}
	if !buff.IsActive() {
		t.Error("Frostbolt's landing spent a charge gained while it flew")
	}
	settle()

	chargeRevelation(sim, buff)
	mage.missiles.Cast(sim, target)
	channel := mage.missiles.Dot(target)
	if !channel.IsActive() || !buff.IsActive() {
		t.Fatal("Arcane Missiles started and Revelation's charge is gone before its first missile")
	}
	stepWhile(t, sim, func() bool { return len(mage.crits[mage.missile]) == 0 })
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if buff.IsActive() {
		t.Error("Arcane Missiles' first missile landed and Revelation's charge is still up")
	}
	// A later missile that lands without a crit may proc a new charge, which is taken away before the next.
	stepWhile(t, sim, func() bool {
		buff.Deactivate(sim)
		return channel.IsActive()
	})
	if got := crits(mage.missile); len(got) != 3 || !got[0] || got[1] || got[2] {
		t.Errorf("Arcane Missiles' ticks crit %v, want the first only", got)
	}
	settle()

	chargeRevelation(sim, buff)
	mage.stalled.Cast(sim, target)
	if character.IsChanneling() {
		t.Fatal("the cast that starts no channel left a channel running")
	}
	if buff.IsActive() {
		t.Error("a channeled cast completed without starting its channel and Revelation's charge is still up")
	}
}

// A hit event rolls Revelation once however many enemies it strikes: an Arcane Explosion that lands on 5
// enemies without a crit procs it about as often as the enchant's chance for one roll (EV-0149).
func TestRevelationRollsOncePerHitEvent(t *testing.T) {
	const casts, enemies = 600, 5
	sim, character := revelationWearerFacing(t, proto.Class_ClassMage, &proto.Player_Mage{Mage: &proto.Mage{}}, enemies)
	mage := sim.Raid.Parties[0].Players[0].(*testMage)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -character.GetStat(stats.SpellCritPercent)})
	buff := character.GetAuraByID(core.ActionID{SpellID: 1248808})

	chance := 0.0
	for _, effect := range core.GetEnchantByEffectID(revelationEnchant).EnchantEffects {
		if effect.BuffId == 1248806 {
			chance = effect.GetProc().GetProcChance()
		}
	}
	if chance <= 0 || chance >= 0.5 {
		t.Fatalf("the enchant's entry states a chance of %v for 1248806, want one between none and a half", chance)
	}

	procs := 0
	for range casts {
		buff.Deactivate(sim)
		mage.explosion.Cast(sim, character.CurrentTarget)
		stepTo(t, sim, sim.CurrentTime+2*core.SpellBatchWindow)
		if buff.IsActive() {
			procs++
		}
	}
	if hits := len(mage.crits[mage.explosion]); hits != casts*enemies || slices.Contains(mage.crits[mage.explosion], true) {
		t.Fatalf("%d casts landed %d hits with crits among them, want %d and none crit", casts, hits, casts*enemies)
	}
	if want := casts * chance; float64(procs) < want/2 || float64(procs) > 2*want {
		t.Errorf("%d casts on %d enemies procced Revelation %d times, want about %.0f, one roll a cast", casts, enemies, procs, want)
	}
}

// The first result of a hit event is its one chance: an Arcane Explosion whose first target crits never
// procs Revelation, whatever its other targets do, and one whose first target does not crit can (EV-0149).
func TestRevelationRollsNothingWhenAnEventsFirstResultCrits(t *testing.T) {
	const casts, enemies = 600, 5
	sim, character := revelationWearerFacing(t, proto.Class_ClassMage, &proto.Player_Mage{Mage: &proto.Mage{}}, enemies)
	mage := sim.Raid.Parties[0].Players[0].(*testMage)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: 50 - character.GetStat(stats.SpellCritPercent)})
	buff := character.GetAuraByID(core.ActionID{SpellID: 1248808})

	critFirst, mixed := 0, 0
	procsAfter := map[bool]int{}
	for range casts {
		buff.Deactivate(sim)
		mage.crits[mage.explosion] = nil
		mage.explosion.Cast(sim, character.CurrentTarget)
		stepTo(t, sim, sim.CurrentTime+2*core.SpellBatchWindow)
		got := mage.crits[mage.explosion]
		if got[0] {
			critFirst++
			if slices.Contains(got, false) {
				mixed++
			}
		}
		if buff.IsActive() {
			procsAfter[got[0]]++
		}
	}
	if critFirst < casts/4 || mixed < casts/4 {
		t.Fatalf("%d of %d casts crit their first target and %d of those left another uncrit, want a fair share of each", critFirst, casts, mixed)
	}
	if procsAfter[true] != 0 {
		t.Errorf("%d casts whose first target crit procced Revelation, want none", procsAfter[true])
	}
	if procsAfter[false] == 0 {
		t.Errorf("none of the %d casts whose first target did not crit procced Revelation", casts-critFirst)
	}
}

// The charge goes as the cast completes on a class other than the mage too: a druid's bolt spends it as
// it is cast and keeps the crit it rolled while it flies (EV-0149).
func TestRevelationDruidSpendsTheChargeAtTheCast(t *testing.T) {
	sim, character := revelationWearer(t, proto.Class_ClassDruid, &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{}})
	druid := sim.Raid.Parties[0].Players[0].(*testDruid)
	buff := character.GetAuraByID(core.ActionID{SpellID: 1323400})

	chargeRevelation(sim, buff)
	druid.bolt.Cast(sim, character.CurrentTarget)
	if buff.IsActive() {
		t.Error("the bolt was cast and the druid's charge is still up while it flies")
	}
	if !druid.last.DidCrit() {
		t.Error("the bolt cast with Revelation up did not crit")
	}
}

// A spell far under the wearer's level rolls nothing: Frostbolt's first rank never procs Revelation at
// 60, and a Fire Blast rank near the wearer's level does (EV-0149).
func TestRevelationRollsNothingOffASpellFarUnderTheWearer(t *testing.T) {
	sim, character := revelationWearer(t, proto.Class_ClassMage, &proto.Player_Mage{Mage: &proto.Mage{}})
	if low, high := spelldata.MustFind(116).SpellLevel, spelldata.MustFind(10199).SpellLevel; int32(low)+32 > character.Level || int32(high)+12 < character.Level {
		t.Fatalf("the two ranks no longer sit at the ramp's ends: spell levels %d and %d", low, high)
	}
	mage := sim.Raid.Parties[0].Players[0].(*testMage)
	target := character.CurrentTarget
	buff := character.GetAuraByID(core.ActionID{SpellID: 1248808})

	for range 400 {
		mage.firstRank.Cast(sim, target)
		stepTo(t, sim, sim.CurrentTime+2*core.SpellBatchWindow)
	}
	if buff.IsActive() {
		t.Fatal("Frostbolt's first rank procced Revelation at level 60")
	}
	for casts := 0; !buff.IsActive(); casts++ {
		if casts > 300 {
			t.Fatal("300 Fire Blasts and Revelation never procced")
		}
		mage.fireBlast.Cast(sim, target)
		stepTo(t, sim, sim.CurrentTime+2*core.SpellBatchWindow)
	}
}

// Fireball has a periodic part and the mage's buff 1248808 does not name it. Its cast spends the charge
// all the same and its direct hit crits, while the spell's own crit chance, which a tick on the same
// spell would roll, stays where it was (EV-0149).
func TestRevelationCritsTheDirectHitOfASpellItsRowDoesNotName(t *testing.T) {
	sim, character := revelationWearer(t, proto.Class_ClassMage, &proto.Player_Mage{Mage: &proto.Mage{}})
	mage := sim.Raid.Parties[0].Players[0].(*testMage)
	target := character.CurrentTarget
	buff := character.GetAuraByID(core.ActionID{SpellID: 1248808})

	baseCrit := mage.fireball.SpellCritChance(target)
	chargeRevelation(sim, buff)
	if got := mage.fireball.SpellCritChance(target); got != baseCrit {
		t.Errorf("Fireball's crit chance is %v with Revelation up and %v without", got, baseCrit)
	}
	mage.fireball.Cast(sim, target)
	if got := mage.crits[mage.fireball]; len(got) != 1 || !got[0] {
		t.Errorf("Fireball cast with Revelation up crit %v, want its one direct hit", got)
	}
	if buff.IsActive() {
		t.Error("Fireball was cast and Revelation's charge is still up")
	}
	if got := mage.fireball.SpellCritChance(target); got != baseCrit {
		t.Errorf("Fireball's crit chance is %v after the cast that spent the charge, want %v", got, baseCrit)
	}
}

// The procs per minute the enchant's entry states for the spell, which is the rate its proc registers on.
func enchantEntryPPM(t *testing.T, enchantID int32, spellID int32) float64 {
	t.Helper()
	for _, effect := range core.GetEnchantByEffectID(enchantID).EnchantEffects {
		if effect.BuffId == spellID && effect.GetProc().GetPpm() > 0 {
			return effect.GetProc().GetPpm()
		}
	}
	t.Fatalf("%d's entry states no procs per minute for %d", enchantID, spellID)
	return 0
}

// Checks a hit under the mask rolls the procs per minute the enchant's entry states for the spell on
// the weapon's speed, a chance between none and certain.
func expectEnchantRate(t *testing.T, sim *core.Simulation, proc *core.Aura, mask core.ProcMask, enchant, spell, weapon int32) {
	t.Helper()
	want := enchantEntryPPM(t, enchant, spell) * core.GetItemByID(weapon).SwingSpeed / 60
	if proc == nil || proc.Dpm == nil {
		t.Fatalf("%d registers no proc on a procs per minute rate", enchant)
	}
	if got := proc.Dpm.Chance(mask, sim); want <= 0 || want >= 1 || math.Abs(got-want) > 1e-9 {
		t.Errorf("%d rolls %v on a hit of weapon %d, want the entry's procs per minute on the weapon's speed, %v", enchant, got, weapon, want)
	}
}

// Enchant Weapon - Crusader 1900 rolls the procs per minute its enchant entry states for Holy Strength
// 20007 on the enchanted weapon's speed.
func TestCrusaderProcsAtItsEntryRate(t *testing.T) {
	const crusader, holyStrength, weapon = 1900, 20007, 1613
	items := weaponSlots(weapon, 0)
	items[proto.ItemSlot_ItemSlotMainHand].Enchant = crusader
	sim, character := itemWearerWithEquipment(t, proto.Class_ClassWarrior, &proto.EquipmentSpec{Items: items})
	expectEnchantRate(t, sim, character.GetAura("Enchant Weapon - Crusader"), core.ProcMaskMeleeMHAuto, crusader, holyStrength, weapon)
}

// Enchant Weapon - Fiery Weapon 803 rolls the procs per minute its enchant entry states for its strike
// 13897 on the enchanted weapon's speed.
func TestFieryWeaponProcsAtItsEntryRate(t *testing.T) {
	const fieryWeapon, strike, weapon = 803, 13897, 1613
	items := weaponSlots(weapon, 0)
	items[proto.ItemSlot_ItemSlotMainHand].Enchant = fieryWeapon
	sim, character := itemWearerWithEquipment(t, proto.Class_ClassWarrior, &proto.EquipmentSpec{Items: items})
	expectEnchantRate(t, sim, character.GetAura("Enchant Weapon - Fiery Weapon"), core.ProcMaskMeleeMHAuto, fieryWeapon, strike, weapon)
}

// Fiery Weapon on the off hand alone rolls on that weapon's hits at that weapon's speed, and a hit of
// the main hand without the enchant rolls nothing.
func TestFieryWeaponRollsOnTheHandThatCarriesIt(t *testing.T) {
	const fieryWeapon, strike, mainHand, offHand = 803, 13897, 2244, 871
	if core.GetItemByID(mainHand).SwingSpeed == core.GetItemByID(offHand).SwingSpeed {
		t.Fatalf("%d and %d swing at one speed, so the test cannot tell the hands apart", mainHand, offHand)
	}
	items := weaponSlots(mainHand, offHand)
	items[proto.ItemSlot_ItemSlotOffHand].Enchant = fieryWeapon
	sim, character := itemWearerWithEquipment(t, proto.Class_ClassWarrior, &proto.EquipmentSpec{Items: items})
	proc := character.GetAura("Enchant Weapon - Fiery Weapon")
	expectEnchantRate(t, sim, proc, core.ProcMaskMeleeOHAuto, fieryWeapon, strike, offHand)
	if got := proc.Dpm.Chance(core.ProcMaskMeleeMHAuto, sim); got != 0 {
		t.Errorf("a main-hand hit rolls %v with the enchant on the off hand alone, want 0", got)
	}
}

// Enchant Chest - Retricutioner 7223 keeps 435901's damage shield up while the enchant is worn, read
// from the generated registrations: the row's amount lands whole on the attacker of a melee hit taken,
// through the target's armor and with no roll to miss, once per the row's 5 s, and nothing on a ranged
// shot or a spell. The row's proc flags name the melee hits and its cooldown the lockout.
func TestRetricutionerKeepsItsDamageShieldUp(t *testing.T) {
	sim, character := itemSpecWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotChest, &proto.ItemSpec{Id: 8245, Enchant: 7223})
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -character.GetStat(stats.SpellCritPercent)})
	enemy := sim.Encounter.ActiveTargetUnits[0]
	enemy.AddStatDynamic(sim, stats.Armor, 5000)
	row := spelldata.MustFind(435901)

	shield, shieldDamage := damageShield(t, character, 435901)
	takeHit(t, sim, character, meleeSwing(enemy))
	if got := shield.SpellMetrics[enemy.UnitIndex]; got.Casts != 1 || got.Misses != 0 || got.TotalDamage != shieldDamage {
		t.Errorf("a melee hit taken cast the shield %d times, %d missed, for %v; want once, landing the row's %v through the armor", got.Casts, got.Misses, got.TotalDamage, shieldDamage)
	}
	takeHit(t, sim, character, meleeSwing(enemy))
	if got := shield.SpellMetrics[enemy.UnitIndex].Casts; got != 1 {
		t.Errorf("a second melee hit taken inside the row's %v left the shield at %d casts, want 1", row.Cooldown(), got)
	}
	stepTo(t, sim, sim.CurrentTime+row.Cooldown())
	takeHit(t, sim, character, meleeSwing(enemy))
	if got := shield.SpellMetrics[enemy.UnitIndex].Casts; got != 2 {
		t.Errorf("a melee hit taken after the lockout left the shield at %d casts, want 2", got)
	}
	takeHit(t, sim, character, rangedShot(enemy))
	takeHit(t, sim, character, fireSpell(enemy))
	if got := shield.SpellMetrics[enemy.UnitIndex].Casts; got != 2 {
		t.Errorf("after a ranged shot and a spell hit taken the shield cast %d times, want 2 still", got)
	}
}

// Enchant Weapon - Lesser Beastslayer 853 keeps 13650 up while worn: its damage done against Beasts
// adds to weapon hits and spells alike against a Beast, and to nothing against other targets.
func TestLesserBeastslayerAddsDamageAgainstBeasts(t *testing.T) {
	const beastslayer, weapon = 853, 1613
	amount := spelldata.MustFind(13650).EffectN(1).Average()
	for _, c := range []struct {
		against proto.MobType
		want    float64
	}{
		{proto.MobType_MobTypeBeast, amount},
		{proto.MobType_MobTypeDemon, 0},
	} {
		items := weaponSlots(weapon, 0)
		items[proto.ItemSlot_ItemSlotMainHand].Enchant = beastslayer
		sim, character := itemWearerFacing(t, proto.Class_ClassWarrior, &proto.EquipmentSpec{Items: items}, c.against)
		if got, want := bonusAgainstTarget(character), (stats.Stats{stats.PhysicalDamage: c.want, stats.SpellDamage: c.want}); got != want || amount <= 0 {
			t.Errorf("the bonus against %v is %v, want %v", c.against, got, want)
		}
		attackTable := character.AttackTables[sim.Encounter.ActiveTargetUnits[0].UnitIndex]
		if got, want := character.AutoAttacks.MHAuto().BonusDamage(attackTable), character.GetStat(stats.PhysicalDamage)+c.want; got != want {
			t.Errorf("a main-hand swing at a target of %v deals %v bonus damage, want %v", c.against, got, want)
		}
	}
}
