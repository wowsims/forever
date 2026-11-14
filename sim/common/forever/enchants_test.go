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

// A caster with two druid spells: Wrath, which Revelation's druid buff 1323400 names (family 7 bit 0),
// and Moonfire, which it does not (bit 1). Hurricane is a three-tick channel whose ticks cast a spell
// of their own, both under Wrath's flags. Drain is a channel under Wrath's flags whose own hit lands
// before its dot starts and whose ticks hit nothing. Bolt, under Wrath's flags, rolls at the cast and
// lands a second later.
type testDruid struct {
	core.Character
	wrath, moonfire, hurricane, drain, bolt *core.Spell
	last                                    *core.SpellResult
	tickCrits                               []bool
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
	d.wrath = nuke(5176, 0)
	d.moonfire = nuke(8921, 1)

	wrathFlags := core.ClassFlags{Family: 7, Mask: [4]uint32{1}}
	d.bolt = d.RegisterSpell(core.SpellConfig{
		ActionID:           core.ActionID{SpellID: 2912},
		SpellSchool:        core.SpellSchoolArcane,
		DefenseType:        core.DefenseTypeMagic,
		ProcMask:           core.ProcMaskSpellDamage,
		ClassFlags:         wrathFlags,
		MissileMinDuration: time.Second,
		DamageMultiplier:   1,
		ThreatMultiplier:   1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.DealDamageAfterTravel(sim, spell.CalcDamage(sim, target, 100, spell.OutcomeMagicCrit))
		},
	})
	tick := d.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 16914, Tag: 1},
		SpellSchool:      core.SpellSchoolNature,
		DefenseType:      core.DefenseTypeMagic,
		ProcMask:         core.ProcMaskSpellDamage,
		ClassFlags:       wrathFlags,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeMagicHitAndCrit)
			d.tickCrits = append(d.tickCrits, result.DidCrit())
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
	d.hurricane = channel(16914, "Hurricane", false, func(sim *core.Simulation, target *core.Unit, _ *core.Dot) {
		tick.Cast(sim, target)
	})
	d.drain = channel(17401, "Drain", true, nil)
}

// A hunter with two ranged shots: Arcane Shot, which Revelation's hunter buff 1323410 names (family 9
// bit 11), and Aimed Shot, which it does not (bit 17).
type testHunter struct {
	core.Character
	arcaneShot, aimedShot *core.Spell
	last                  *core.SpellResult
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
}

// A mage with spells Revelation's mage buff 1248808 names: Fire Blast (family 3 bit 1), Frostbolt
// (bit 5), which rolls at the cast and lands a second later, Arcane Explosion (bit 12), which hits
// every target, and Arcane Missiles (bit 11), a three-tick channel, with a second channeled cast under
// its flags that starts no channel. None rolls to miss, and each keeps the crits of its landings.
type testMage struct {
	core.Character
	fireBlast, frostbolt, explosion, missiles, missile, stalled *core.Spell
	crits                                                       map[*core.Spell][]bool
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
// next of them and not by a spell outside the mask. A warrior has no Revelation buff.
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

	// A proc's handler runs one spell batch window after the hit, so each cast is followed by a step.
	castAndStep := func(spell *core.Spell) *core.SpellResult {
		spell.SkipCastAndApplyEffects(sim, target)
		sim.Step()
		return druid.last
	}

	for casts := 0; !buff.IsActive(); casts++ {
		if casts > 300 {
			t.Fatal("300 Moonfires and Revelation never procced")
		}
		castAndStep(druid.moonfire)
	}

	castAndStep(druid.moonfire)
	if !buff.IsActive() {
		t.Fatal("Moonfire, which the buff's mask does not name, spent its charge")
	}
	if result := castAndStep(druid.wrath); !result.DidCrit() {
		t.Error("Wrath with Revelation up did not crit")
	}
	if buff.IsActive() {
		t.Error("Wrath did not spend Revelation's charge")
	}

	_, warrior := revelationWearer(t, proto.Class_ClassWarrior, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{
		Options: &proto.DpsWarrior_Options{ClassOptions: &proto.WarriorOptions{}}}})
	if warrior.GetAura("Revelation") != nil {
		t.Error("a warrior has a Revelation buff")
	}
}

// A channel's ticks roll Revelation, but do not spend it tick by tick: every tick crits with the buff
// up, and the charge goes when the channel ends, run out or cut short.
func TestRevelationLastsTheWholeChannel(t *testing.T) {
	sim, character := revelationWearer(t, proto.Class_ClassDruid, &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{}})
	druid := sim.Raid.Parties[0].Players[0].(*testDruid)
	target := character.CurrentTarget
	buff := character.GetAuraByID(core.ActionID{SpellID: 1323400})
	channel := druid.hurricane.Dot(target)

	procced := false
	for casts := 0; !procced; casts++ {
		if casts > 40 {
			t.Fatal("40 channels and no tick procced Revelation")
		}
		druid.hurricane.SkipCastAndApplyEffects(sim, target)
		stepWhile(t, sim, func() bool {
			procced = procced || buff.IsActive()
			return channel.IsActive()
		})
	}
	sim.Step()
	buff.Deactivate(sim)

	channelFor := func(ticks int) {
		druid.tickCrits = nil
		druid.hurricane.SkipCastAndApplyEffects(sim, target)
		stepWhile(t, sim, func() bool { return len(druid.tickCrits) < ticks })
	}

	chargeRevelation(sim, buff)
	channelFor(2)
	if !buff.IsActive() {
		t.Fatal("two ticks into the channel, Revelation's charge is spent")
	}
	stepWhile(t, sim, channel.IsActive)
	sim.Step()
	if len(druid.tickCrits) != 3 || slices.Contains(druid.tickCrits, false) {
		t.Errorf("ticks crit %v, want all three", druid.tickCrits)
	}
	if buff.IsActive() {
		t.Error("the channel ended and Revelation's charge is still up")
	}

	chargeRevelation(sim, buff)
	channelFor(1)
	channel.Deactivate(sim)
	if buff.IsActive() {
		t.Error("the channel was cut short and Revelation's charge is still up")
	}

	chargeRevelation(sim, buff)
	druid.drain.SkipCastAndApplyEffects(sim, target)
	druid.drain.Dot(target).Deactivate(sim)
	if buff.IsActive() {
		t.Error("a channel whose own hit landed was cut short before its first tick and Revelation's charge is still up")
	}

	drain := druid.drain.Dot(target)
	druid.drain.SkipCastAndApplyEffects(sim, target)
	chargeRevelation(sim, buff)
	sim.Step()
	if !drain.IsActive() || !buff.IsActive() {
		t.Fatal("the charge came while a channel it covers was running, and it is gone before that channel ended")
	}
	stepWhile(t, sim, drain.IsActive)
	if buff.IsActive() {
		t.Error("the charge came while a channel it covers was running, the channel ended and the charge is still up")
	}
}

// A hunter's buff 1323410 also procs off the ranged shots its mask names, and not off a shot outside it.
func TestRevelationProcsOffAHuntersShots(t *testing.T) {
	sim, character := revelationWearer(t, proto.Class_ClassHunter, &proto.Player_Hunter{Hunter: &proto.Hunter{}})
	hunter := sim.Raid.Parties[0].Players[0].(*testHunter)
	target := character.CurrentTarget

	buff := character.GetAuraByID(core.ActionID{SpellID: 1323410})
	if buff == nil {
		t.Fatal("no Revelation buff 1323410 on the hunter")
	}

	castAndStep := func(spell *core.Spell) *core.SpellResult {
		spell.SkipCastAndApplyEffects(sim, target)
		sim.Step()
		return hunter.last
	}

	for range 300 {
		castAndStep(hunter.aimedShot)
	}
	if buff.IsActive() {
		t.Fatal("Aimed Shot, which the buff's mask does not name, procced Revelation")
	}

	for casts := 0; !buff.IsActive(); casts++ {
		if casts > 300 {
			t.Fatal("300 Arcane Shots and Revelation never procced")
		}
		castAndStep(hunter.arcaneShot)
	}
	if result := castAndStep(hunter.arcaneShot); !result.DidCrit() {
		t.Error("Arcane Shot with Revelation up did not crit")
	}
	if buff.IsActive() {
		t.Error("Arcane Shot did not spend Revelation's charge")
	}
}

// A mage's buff 1248808 spends its charge as the cast it reaches completes. Every target of an area
// cast crits and the next cast does not; a missile rolled at the cast keeps its crit in flight, a
// spell cast while it flies gets none, and its landing spends no charge. A channel keeps the charge
// for every tick and spends it at the channel's end, and a channeled cast that starts no channel
// spends it as it completes.
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
		t.Fatal("Arcane Missiles started and Revelation's charge is gone before the channel ends")
	}
	stepWhile(t, sim, channel.IsActive)
	if got := crits(mage.missile); len(got) != 3 || slices.Contains(got, false) {
		t.Errorf("Arcane Missiles' ticks crit %v, want all three", got)
	}
	if buff.IsActive() {
		t.Error("Arcane Missiles ended and Revelation's charge is still up")
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

// The druid's buff 1323400 spends its charge as the spell lands: a missile in flight leaves it up.
func TestRevelationDruidSpendsTheChargeAtImpact(t *testing.T) {
	sim, character := revelationWearer(t, proto.Class_ClassDruid, &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{}})
	druid := sim.Raid.Parties[0].Players[0].(*testDruid)
	buff := character.GetAuraByID(core.ActionID{SpellID: 1323400})

	chargeRevelation(sim, buff)
	cast := sim.CurrentTime
	druid.bolt.Cast(sim, character.CurrentTarget)
	stepTo(t, sim, cast+500*time.Millisecond)
	if !buff.IsActive() {
		t.Error("the druid's charge is gone while the bolt that should spend it is in flight")
	}
	stepTo(t, sim, cast+time.Second+100*time.Millisecond)
	if buff.IsActive() {
		t.Error("the bolt landed and the druid's charge is still up")
	}
}

// Enchant Weapon - Crusader 1900 rolls the procs per minute its enchant entry states for Holy Strength
// 20007 on the enchanted weapon's speed, a chance between none and certain.
func TestCrusaderProcsAtItsEntryRate(t *testing.T) {
	const crusader, holyStrength, weapon = 1900, 20007, 1613
	items := weaponSlots(weapon, 0)
	items[proto.ItemSlot_ItemSlotMainHand].Enchant = crusader
	player := &proto.Player{
		Name: "Wearer", Class: proto.Class_ClassWarrior, Race: proto.Race_RaceTauren,
		Equipment: &proto.EquipmentSpec{Items: items},
		Buffs:     &proto.IndividualBuffs{},
		Rotation:  &proto.APLRotation{},
	}
	core.WithSpec(player, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}})
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	character := sim.Raid.Parties[0].Players[0].GetCharacter()

	var ppm float64
	for _, effect := range core.GetEnchantByEffectID(crusader).EnchantEffects {
		if effect.BuffId == holyStrength {
			ppm = effect.GetProc().GetPpm()
		}
	}
	want := ppm * core.GetItemByID(weapon).SwingSpeed / 60
	proc := character.GetAura("Enchant Weapon - Crusader")
	if proc == nil || proc.Dpm == nil {
		t.Fatal("1900 registers no proc on a procs per minute rate")
	}
	if got := proc.Dpm.Chance(core.ProcMaskMeleeMHAuto, sim); want <= 0 || want >= 1 || math.Abs(got-want) > 1e-9 {
		t.Errorf("Crusader rolls %v on a main-hand hit, want the entry's procs per minute on the weapon's speed, %v", got, want)
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
