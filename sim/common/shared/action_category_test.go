package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

const (
	categoryTrinketID     int32 = 999101
	categoryProcAuraID    int32 = 999102
	categoryOnUseID       int32 = 999103
	categoryOnUseAuraID   int32 = 999104
	categoryTalentAuraID  int32 = 999105
	categoryOwnAuraID     int32 = 999106
	categoryTalentSpellID int32 = 999107
	categoryPetTrinketID  int32 = 999108
	categoryPetBuffID     int32 = 999109
)

// A pet's aura that its owner's major cooldown casts is filed with the cooldowns, where the owner's
// spell is, not by the phase that registered it.
func TestPetAuraOfAnOwnerCooldown(t *testing.T) {
	withAuraItem(categoryPetTrinketID, proto.ItemType_ItemTypeTrinket)
	core.NewItemEffect(categoryPetTrinketID, func(agent core.Agent) {
		owner := agent.GetCharacter()
		buff := owner.Pets[0].RegisterAura(core.Aura{Label: "Category Pet Buff", ActionID: core.ActionID{SpellID: categoryPetBuffID}, Duration: time.Second})
		owner.AddMajorCooldown(core.MajorCooldown{Type: core.CooldownTypeDPS, Spell: owner.RegisterSpell(core.SpellConfig{
			ActionID:     core.ActionID{SpellID: categoryPetBuffID},
			Flags:        core.SpellFlagNoOnCastComplete,
			Cast:         core.CastConfig{CD: core.Cooldown{Timer: owner.NewTimer(), Duration: time.Minute}},
			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) { buff.Activate(sim) },
		})})
	})

	sim := newAuraSim(hunter("Hunter", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: categoryPetTrinketID}))
	for _, aura := range auraTesterAt(sim, 0).pet.GetMetadata().Auras {
		if aura.Id.GetSpellId() == categoryPetBuffID {
			if aura.Category != proto.ActionCategory_ActionCategoryCooldown {
				t.Errorf("the pet's buff: got %v, want %v", aura.Category, proto.ActionCategory_ActionCategoryCooldown)
			}
			return
		}
	}
	t.Fatalf("aura %d missing from the pet's metadata", categoryPetBuffID)
}

func registerCategoryTrinket() {
	core.NewItemEffect(categoryTrinketID, func(agent core.Agent) {
		character := agent.GetCharacter()

		character.RegisterAura(core.Aura{Label: "Category Proc", ActionID: core.ActionID{SpellID: categoryProcAuraID}, Duration: time.Second})

		onUseAura := character.RegisterAura(core.Aura{Label: "Category On-Use", ActionID: core.ActionID{SpellID: categoryOnUseAuraID}, Duration: time.Second})
		onUse := character.RegisterSpell(core.SpellConfig{
			ActionID:        core.ActionID{ItemID: categoryTrinketID},
			Flags:           core.SpellFlagNoOnCastComplete,
			Cast:            core.CastConfig{CD: core.Cooldown{Timer: character.NewTimer(), Duration: time.Minute}},
			ApplyEffects:    func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) { onUseAura.Activate(sim) },
			RelatedSelfBuff: onUseAura,
		})
		character.AddMajorCooldown(core.MajorCooldown{Spell: onUse, Type: core.CooldownTypeDPS})

		character.RegisterAura(core.Aura{Label: "Category Talent", ActionID: core.ActionID{SpellID: categoryTalentAuraID}, Source: core.SourceTalent})
		character.RegisterAura(core.Aura{Label: "Category Own", ActionID: core.ActionID{SpellID: categoryOwnAuraID}, Source: core.SourceNone})
		character.RegisterSpell(core.SpellConfig{ActionID: core.ActionID{SpellID: categoryTalentSpellID}, Flags: core.SpellFlagAPL, Source: core.SourceTalent})
	})
}

// Each aura and spell is filed by the phase that registered it, unless it names its own source, and a
// major cooldown's buff is filed with the cooldown whatever registered it. A spell a talent grants
// stays with the class's own spells.
func TestActionCategories(t *testing.T) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotTrinket1+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: categoryTrinketID, Name: "Category Trinket",
		Type: proto.ItemType_ItemTypeTrinket, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})
	registerCategoryTrinket()
	items[proto.ItemSlot_ItemSlotTrinket1] = &proto.ItemSpec{Id: categoryTrinketID}

	sim := newTestCasterSim(items, nil)
	metadata := sim.Raid.Parties[0].Players[0].GetCharacter().GetMetadata()

	auras := map[int32]proto.ActionCategory{}
	for _, aura := range metadata.Auras {
		auras[aura.Id.GetSpellId()] = aura.Category
	}
	spells := map[int32]proto.ActionCategory{}
	for _, spell := range metadata.Spells {
		spells[spell.Id.GetSpellId()+spell.Id.GetItemId()] = spell.Category
	}

	for _, id := range []int32{categoryProcAuraID, categoryOnUseAuraID, categoryTalentAuraID, categoryOwnAuraID} {
		if _, ok := auras[id]; !ok {
			t.Fatalf("aura %d missing from the metadata", id)
		}
	}
	for _, id := range []int32{categoryTrinketID, testBoltID, 20554, categoryTalentSpellID} {
		if _, ok := spells[id]; !ok {
			t.Fatalf("spell %d missing from the metadata", id)
		}
	}

	for _, c := range []struct {
		name string
		got  proto.ActionCategory
		want proto.ActionCategory
	}{
		{"item proc aura", auras[categoryProcAuraID], proto.ActionCategory_ActionCategoryItem},
		{"on-use buff", auras[categoryOnUseAuraID], proto.ActionCategory_ActionCategoryCooldown},
		{"on-use spell", spells[categoryTrinketID], proto.ActionCategory_ActionCategoryCooldown},
		{"aura tagged Talent", auras[categoryTalentAuraID], proto.ActionCategory_ActionCategoryTalent},
		{"aura tagged None", auras[categoryOwnAuraID], proto.ActionCategory_ActionCategoryUnknown},
		{"the caster's own spell", spells[testBoltID], proto.ActionCategory_ActionCategoryUnknown},
		{"a spell a talent grants", spells[categoryTalentSpellID], proto.ActionCategory_ActionCategoryUnknown},
		{"Berserking, a racial major cooldown", spells[20554], proto.ActionCategory_ActionCategoryCooldown},
	} {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
}
