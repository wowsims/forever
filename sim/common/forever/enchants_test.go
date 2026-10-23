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
)

const revelationEnchant = 8217

func init() {
	core.RegisterAgentFactory(
		proto.Player_BalanceDruid{},
		proto.Spec_SpecBalanceDruid,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			return &testDruid{Character: *character}
		},
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_BalanceDruid)
		},
	)
	core.RegisterAgentFactory(
		proto.Player_Hunter{},
		proto.Spec_SpecHunter,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			return &testHunter{Character: *character}
		},
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_Hunter)
		},
	)
}

// A caster with two druid spells: Wrath, which Revelation's druid buff 1323400 names (family 7 bit 0),
// and Moonfire, which it does not (bit 1). Hurricane is a three-tick channel whose ticks cast a spell
// of their own, both under Wrath's flags. Drain is a channel under Wrath's flags whose own hit lands
// before its dot starts and whose ticks hit nothing.
type testDruid struct {
	core.Character
	wrath, moonfire, hurricane, drain *core.Spell
	last                              *core.SpellResult
	tickCrits                         []bool
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

func revelationWearer(t *testing.T, class proto.Class, spec any) (*core.Simulation, *core.Character) {
	t.Helper()
	items := weaponSlots(873, 0)
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
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
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

	stepWhile := func(cond func() bool) {
		for steps := 0; cond(); steps++ {
			if steps > 1000 {
				t.Fatal("the sim did not get there in 1000 steps")
			}
			sim.Step()
		}
	}
	procced := false
	for casts := 0; !procced; casts++ {
		if casts > 40 {
			t.Fatal("40 channels and no tick procced Revelation")
		}
		druid.hurricane.SkipCastAndApplyEffects(sim, target)
		stepWhile(func() bool {
			procced = procced || buff.IsActive()
			return channel.IsActive()
		})
	}
	sim.Step()
	buff.Deactivate(sim)

	charge := func() {
		buff.Activate(sim)
		buff.SetStacks(sim, buff.MaxStacks)
	}
	channelFor := func(ticks int) {
		druid.tickCrits = nil
		druid.hurricane.SkipCastAndApplyEffects(sim, target)
		stepWhile(func() bool { return len(druid.tickCrits) < ticks })
	}

	charge()
	channelFor(2)
	if !buff.IsActive() {
		t.Fatal("two ticks into the channel, Revelation's charge is spent")
	}
	stepWhile(channel.IsActive)
	sim.Step()
	if len(druid.tickCrits) != 3 || slices.Contains(druid.tickCrits, false) {
		t.Errorf("ticks crit %v, want all three", druid.tickCrits)
	}
	if buff.IsActive() {
		t.Error("the channel ended and Revelation's charge is still up")
	}

	charge()
	channelFor(1)
	channel.Deactivate(sim)
	if buff.IsActive() {
		t.Error("the channel was cut short and Revelation's charge is still up")
	}

	charge()
	druid.drain.SkipCastAndApplyEffects(sim, target)
	druid.drain.Dot(target).Deactivate(sim)
	if buff.IsActive() {
		t.Error("a channel whose own hit landed was cut short before its first tick and Revelation's charge is still up")
	}

	drain := druid.drain.Dot(target)
	druid.drain.SkipCastAndApplyEffects(sim, target)
	charge()
	sim.Step()
	if !drain.IsActive() || !buff.IsActive() {
		t.Fatal("the charge came while a channel it covers was running, and it is gone before that channel ended")
	}
	stepWhile(drain.IsActive)
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

// Enchant Weapon - Crusader 1900 rolls the procs per minute its enchant entry states for Holy Strength
// 20007 on the enchanted weapon's speed, a chance between none and certain.
func TestCrusaderProcsAtItsEntryRate(t *testing.T) {
	const crusader, holyStrength, weapon = 1900, 20007, 873
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
