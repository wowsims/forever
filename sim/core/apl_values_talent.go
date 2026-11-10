package core

import (
	"fmt"

	"github.com/wowsims/forever/sim/core/proto"
)

// A rank counts as known when the player holds it or a higher one, so a condition on rank 3 still
// holds for a build with 5 points.
func (rot *APLRotation) newValueTalentIsKnown(config *proto.APLValueTalentIsKnown, uuid *proto.UUID) APLValue {
	if config.TalentId == nil {
		return nil
	}

	spellID := config.TalentId.GetSpellId()
	rank := max(config.TalentId.Rank, 1)

	known := rot.talentRank(spellID) >= rank
	if !known {
		rot.ValidationMessage(proto.LogLevel_Warning, "%s does not know talent %s (Rank %d)", rot.unit.Label, ActionID{SpellID: spellID}, rank)
	}

	return rot.newValueConst(&proto.APLValueConst{
		Val: Ternary(known, "true", "false"),
	}, uuid)
}

type APLValueTalentRank struct {
	DefaultAPLValueImpl
	talentID ActionID
	rank     int32
}

func (rot *APLRotation) newValueTalentRank(config *proto.APLValueTalentRank, _ *proto.UUID) APLValue {
	if config.TalentId == nil {
		return nil
	}

	talentID := ActionID{SpellID: config.TalentId.GetSpellId()}
	rank := rot.talentRank(talentID.SpellID)
	if rank == 0 {
		rot.ValidationMessage(proto.LogLevel_Warning, "%s does not know talent %s", rot.unit.Label, talentID)
	}

	return &APLValueTalentRank{
		talentID: talentID,
		rank:     rank,
	}
}
func (value *APLValueTalentRank) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueTalentRank) GetInt(_ *Simulation) int32 {
	return value.rank
}
func (value *APLValueTalentRank) GetFloat(_ *Simulation) float64 {
	return float64(value.rank)
}
func (value *APLValueTalentRank) String() string {
	return fmt.Sprintf("Talent Rank(%s)", value.talentID)
}

// The rank the rotation's player holds in the talent, 0 for a unit that is not a player.
func (rot *APLRotation) talentRank(spellID int32) int32 {
	if agent := rot.unit.Env.Raid.GetPlayerFromUnit(rot.unit); agent != nil {
		return agent.GetCharacter().TalentRank(spellID)
	}
	return 0
}
