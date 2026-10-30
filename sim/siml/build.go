package siml

import "github.com/wowsims/forever/sim/core/proto"

// Loads the settings of a preset build b over s, a spec's defaults, as the site applies a build (applyBuild in
// ui/features/settings/model/apply_build.ts). Each message b holds replaces s's own, except an Auto rotation, which
// keeps s's, and a build with no item swap set turns item swap off. An encounter brings the tanks and the target
// dummies with it, and the healing model when b sets one. When b sets the spec's options it also sets the distance and whether the player
// stands in front of the target, and the race, professions and timings it does not leave at 0.
func LoadBuild(s, b *proto.IndividualSimSettings) {
	p, bp := s.Player, b.GetPlayer()
	if bp.Rotation != nil && bp.Rotation.Type != proto.APLRotation_TypeAuto {
		p.Rotation = bp.Rotation
	}
	if bp.Equipment != nil {
		p.Equipment = bp.Equipment
	}
	p.EnableItemSwap = bp.ItemSwap != nil && bp.EnableItemSwap
	if bp.ItemSwap != nil {
		p.ItemSwap = bp.ItemSwap
	}
	if bp.TalentsString != "" {
		p.TalentsString = bp.TalentsString
	}
	if bp.Consumables != nil {
		p.Consumables = bp.Consumables
	}
	if bp.Buffs != nil {
		p.Buffs = bp.Buffs
	}
	if b.PartyBuffs != nil {
		s.PartyBuffs = b.PartyBuffs
	}
	if b.RaidBuffs != nil {
		s.RaidBuffs = b.RaidBuffs
	}
	if b.Debuffs != nil {
		s.Debuffs = b.Debuffs
	}
	if b.Encounter != nil {
		s.Encounter, s.Tanks, s.TargetDummies = b.Encounter, b.Tanks, b.TargetDummies
		if bp.HealingModel != nil {
			p.HealingModel = bp.HealingModel
		}
	}

	m := bp.ProtoReflect()
	fd := m.WhichOneof(m.Descriptor().Oneofs().ByName("spec"))
	if fd == nil || !m.Get(fd).Message().Has(fd.Message().Fields().ByName("options")) {
		return
	}
	p.Spec = bp.Spec
	p.DistanceFromTarget, p.InFrontOfTarget = bp.DistanceFromTarget, bp.InFrontOfTarget
	if bp.Race != proto.Race_RaceUnknown {
		p.Race = bp.Race
	}
	if bp.Profession1 != proto.Profession_ProfessionUnknown {
		p.Profession1 = bp.Profession1
	}
	if bp.Profession2 != proto.Profession_ProfessionUnknown {
		p.Profession2 = bp.Profession2
	}
	if bp.ReactionTimeMs != 0 {
		p.ReactionTimeMs = bp.ReactionTimeMs
	}
	if bp.ChannelClipDelayMs != 0 {
		p.ChannelClipDelayMs = bp.ChannelClipDelayMs
	}
}
