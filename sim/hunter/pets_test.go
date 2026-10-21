package hunter

import (
	"math"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

// The tallstrider's Dust Cloud bids in the slot the raid's Dust Cloud, Faerie Fire and Curse of
// Recklessness share: it lands beside Expose Armor and both apply, and beside the raid's own Dust
// Cloud it lands too, with the target still losing one -505.
func TestPetDustCloudSharesTheRaidsDustCloudSlot(t *testing.T) {
	for _, c := range []struct {
		name    string
		debuffs *proto.Debuffs
		other   string
		armor   float64
	}{
		{"beside Expose Armor", &proto.Debuffs{ExposeArmor: true}, "Expose Armor (External)", -505},
		{"against the raid's Dust Cloud", &proto.Debuffs{DustCloud: true}, "Dust Cloud (External)", 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, pet, target := newPetSim(proto.HunterOptions_Tallstrider, c.debuffs)

			id := core.ActionID{SpellID: spellData.DustCloudTriggered.Highest().ID}
			spell, aura := pet.GetSpell(id), target.GetAuraByID(id)
			if spell == nil || aura == nil {
				t.Fatalf("spell %v, aura %v; the tallstrider has no Dust Cloud", spell, aura)
			}
			if !spell.ExtraCastCondition(sim, target) {
				t.Error("the pet cannot cast Dust Cloud, want it castable over the other debuff")
			}

			other := target.GetAura(c.other)
			before := target.GetStat(stats.Armor)
			spell.SkipCastAndApplyEffects(sim, target)
			if !aura.IsActive() || !other.IsActive() {
				t.Errorf("the pet's Dust Cloud active %v, %s active %v; want both on the target",
					aura.IsActive(), c.other, other.IsActive())
			}
			if got := target.GetStat(stats.Armor) - before; got != c.armor {
				t.Errorf("the pet's Dust Cloud moved armor by %v, want %v", got, c.armor)
			}

			other.Deactivate(sim)
			withPet := target.GetStat(stats.Armor)
			aura.Deactivate(sim)
			if got := withPet - target.GetStat(stats.Armor); got != -505 {
				t.Errorf("with %s off, the pet's Dust Cloud is worth %v, want -505", c.other, got)
			}
		})
	}
}

// The carrion bird's Demoralizing Screech bids in the slot the raid's Demoralizing Shout holds: both
// are on the target and it loses one -205, and with the shout off the screech's -205 applies.
func TestPetDemoralizingScreechSharesTheRaidsDemoralizingSlot(t *testing.T) {
	sim, pet, target := newPetSim(proto.HunterOptions_CarrionBird, &proto.Debuffs{DemoralizingShout: true})

	id := core.ActionID{SpellID: spellData.DemoralizingScreechTriggered.Highest().ID}
	spell, aura := pet.GetSpell(id), target.GetAuraByID(id)
	shout := target.GetAura("Demoralizing Shout (External)")
	if spell == nil || aura == nil || shout == nil {
		t.Fatalf("spell %v, aura %v, the raid's shout %v; want all three", spell, aura, shout)
	}

	before := target.GetStat(stats.AttackPower)
	spell.SkipCastAndApplyEffects(sim, target)
	if !aura.IsActive() || !shout.IsActive() {
		t.Errorf("the screech active %v, the shout active %v; want both on the target", aura.IsActive(), shout.IsActive())
	}
	if got := target.GetStat(stats.AttackPower) - before; got != 0 {
		t.Errorf("the screech moved attack power by %v beside the shout, want 0", got)
	}

	shout.Deactivate(sim)
	if got := target.GetStat(stats.AttackPower) - before; got != 0 {
		t.Errorf("with the shout off, attack power moved by %v, want 0: the screech's -205 takes the shout's place", got)
	}
}

// A hunter with no gear and a pet of the family, up for the whole fight, against one target carrying
// the raid's debuffs.
func newPetSim(petType proto.HunterOptions_PetType, debuffs *proto.Debuffs) (*core.Simulation, *HunterPet, *core.Unit) {
	return newPetSimWithTalents(DefaultBMTalents, petType, debuffs)
}

func newPetSimWithTalents(talents string, petType proto.HunterOptions_PetType, debuffs *proto.Debuffs) (*core.Simulation, *HunterPet, *core.Unit) {
	player := &proto.Player{
		Name:          "Pet",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassHunter,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{
			ClassOptions: &proto.HunterOptions{PetType: petType, PetUptime: 1},
		}}},
		Rotation: &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, debuffs),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*Hunter).Pet, sim.Encounter.AllTargetUnits[0]
}

// The generated family table and the options enum name the same families: a family the game files
// add needs its enum value, and an enum value the game files no longer describe has to go.
func TestPetFamiliesMatchTheOptions(t *testing.T) {
	for name := range petFamilies {
		if _, ok := proto.HunterOptions_PetType_value[name]; !ok {
			t.Errorf("pet family %q has no HunterOptions.PetType value", name)
		}
	}
	for name, value := range proto.HunterOptions_PetType_value {
		if value == 0 {
			continue
		}
		if _, ok := petFamilies[name]; !ok {
			t.Errorf("HunterOptions.PetType %s names no generated pet family", name)
		}
	}
	for name, value := range proto.HunterOptions_PetAttackSpeed_value {
		if value == 0 {
			continue
		}
		if _, ok := petAttackSpeeds[proto.HunterOptions_PetAttackSpeed(value)]; !ok {
			t.Errorf("HunterOptions.PetAttackSpeed %s names no passive", name)
		}
	}
}

// The generated arrow table and the options enum name the same arrows, and every value is its
// arrow's item id, which is what the UI lists them by.
func TestArrowsMatchTheOptions(t *testing.T) {
	for name, arrow := range arrows {
		if value, ok := proto.HunterOptions_Ammo_value[name]; !ok {
			t.Errorf("arrow %q has no HunterOptions.Ammo value", name)
		} else if value != arrow.ItemID {
			t.Errorf("HunterOptions.Ammo %s is %d, not the arrow's item id %d", name, value, arrow.ItemID)
		}
	}
	for name, value := range proto.HunterOptions_Ammo_value {
		if value == 0 {
			continue
		}
		if _, ok := arrows[name]; !ok {
			t.Errorf("HunterOptions.Ammo %s names no generated arrow", name)
		}
	}
}

// The same for the quivers.
func TestQuiversMatchTheOptions(t *testing.T) {
	for name, quiver := range quivers {
		if value, ok := proto.HunterOptions_QuiverBonus_value[name]; !ok {
			t.Errorf("quiver %q has no HunterOptions.QuiverBonus value", name)
		} else if value != quiver.ItemID {
			t.Errorf("HunterOptions.QuiverBonus %s is %d, not the quiver's item id %d", name, value, quiver.ItemID)
		}
	}
	for name, value := range proto.HunterOptions_QuiverBonus_value {
		if value == 0 {
			continue
		}
		if _, ok := quivers[name]; !ok {
			t.Errorf("HunterOptions.QuiverBonus %s names no generated quiver", name)
		}
	}
}

// With a pet out, Lone Wolf divides the pet's damage by 1.2 and gives the hunter nothing; without
// one, the hunter has the row's 20%.
func TestLoneWolfCutsThePetsDamage(t *testing.T) {
	withoutLoneWolf := strings.Replace(DefaultMMTalents, "25115", "25015", 1)
	sim, pet, _ := newPetSimWithTalents(withoutLoneWolf, proto.HunterOptions_Cat, nil)
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	if hunter.Talents.LoneWolf {
		t.Fatalf("talents %s take Lone Wolf, want the Marksmanship set without it", withoutLoneWolf)
	}
	petBase := pet.PseudoStats.DamageDealtMultiplier
	hunterBase := hunter.PseudoStats.DamageDealtMultiplier

	sim, pet, _ = newPetSimWithTalents(DefaultMMTalents, proto.HunterOptions_Cat, nil)
	hunter = sim.Raid.Parties[0].Players[0].(*Hunter)
	if !hunter.Talents.LoneWolf {
		t.Fatalf("talents %s leave out Lone Wolf", DefaultMMTalents)
	}
	if got := pet.PseudoStats.DamageDealtMultiplier / petBase; math.Abs(got-1/1.2) > 1e-9 {
		t.Errorf("Lone Wolf multiplies the pet's damage by %v, want 1/1.2", got)
	}
	if got := hunter.PseudoStats.DamageDealtMultiplier / hunterBase; got != 1 {
		t.Errorf("Lone Wolf with a pet out multiplies the hunter's damage by %v, want 1", got)
	}

	sim, _, _ = newPetSimWithTalents(DefaultMMTalents, proto.HunterOptions_PetNone, nil)
	hunter = sim.Raid.Parties[0].Players[0].(*Hunter)
	if got := hunter.PseudoStats.DamageDealtMultiplier / hunterBase; math.Abs(got-1.2) > 1e-9 {
		t.Errorf("Lone Wolf without a pet multiplies the hunter's damage by %v, want 1.2", got)
	}
}
