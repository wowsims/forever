package buffs

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func init() {
	core.RegisterEurekaMods(eurekaMods)
}

func eurekaMods(spellID int32) core.EurekaMods {
	row := spelldata.MustFind(spellID)
	cost := row.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_COST))
	damage := row.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_DAMAGE))
	channel := row.FindEffect(dbcenums.E_APPLY_AURA, dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_DOT))
	return core.EurekaMods{
		CostFlags:    cost.ClassFlags,
		DamageFlags:  damage.ClassFlags,
		ChannelFlags: channel.ClassFlags,
		Cost:         cost.Percent(),
		Damage:       damage.Percent(),
		Channel:      channel.Percent(),
		Duration:     row.Duration(),
		Cooldown:     row.Cooldown(),
	}
}
