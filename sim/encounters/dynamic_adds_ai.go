package encounters

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	dynamicBossID int32 = 99999
	dynamicAddID  int32 = 99998
)

// A despawning add dies, and the raid gets the kill at this chance unless the boss's Kill share % input sets another.
const dynamicAddKillSharePercent = 20

func addDynamicAddsAI() {
	createDynamicAddsAIPreset()
}

func createDynamicAddsAIPreset() {
	bossName := "Dynamic Boss"
	addName := "Dynamic Add"

	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: "Default",

		Config: &proto.Target{
			Id:        dynamicBossID,
			Name:      bossName,
			Level:     63,
			MobType:   proto.MobType_MobTypeMechanical,
			TankIndex: 0,

			Stats: stats.Stats{
				stats.Health:      1_428_900,
				stats.Armor:       3731,
				stats.AttackPower: 320,
			}.ToProtoArray(),

			SpellSchool:   proto.SpellSchool_SpellSchoolPhysical,
			SwingSpeed:    2.0,
			MinBaseDamage: 2601,
			DamageSpread:  0.5,
			ParryHaste:    true,
			CanCrush:      true,
			TargetInputs:  dynamicAddsTargetInputs(),
		},

		AI: makeDynamicAddsAI(true),
	})

	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: "Default",

		Config: &proto.Target{
			Id:        dynamicAddID,
			Name:      addName,
			Level:     63,
			MobType:   proto.MobType_MobTypeMechanical,
			TankIndex: 1,

			Stats: stats.Stats{
				stats.Armor:       6193,
				stats.AttackPower: 320,
			}.ToProtoArray(),

			SpellSchool:   proto.SpellSchool_SpellSchoolPhysical,
			SwingSpeed:    1.5,
			MinBaseDamage: 5240,
			DamageSpread:  0.4,
			TargetInputs:  []*proto.TargetInput{},
		},

		AI: makeDynamicAddsAI(false),
	})

	core.AddPresetEncounter("Dynamic Adds", []string{
		"Default/" + bossName,
		"Default/" + addName,
	})
}

func dynamicAddsTargetInputs() []*proto.TargetInput {
	return []*proto.TargetInput{
		{
			Label:       "Add(s) respawn Time",
			Tooltip:     "Time for add(s) to respawn after previous died (in seconds)",
			InputType:   proto.InputType_Number,
			NumberValue: 10,
		},
		{
			Label:       "Add(s) lifetime",
			Tooltip:     "How long the add(s) stay alive (in seconds)",
			InputType:   proto.InputType_Number,
			NumberValue: 20,
		},
		{
			Label:       "Add(s) spawn delay",
			Tooltip:     "Initial delay before the add(s) spawn (in seconds)",
			InputType:   proto.InputType_Number,
			NumberValue: 10,
		},
		{
			Label:       "Kill share %",
			Tooltip:     "Chance that a despawning add counts as your kill (on-kill effects).",
			InputType:   proto.InputType_Number,
			NumberValue: dynamicAddKillSharePercent,
		},
		{
			Label:     "Apply debuffs to adds",
			Tooltip:   "Whether the add(s) get the debuffs from the Debuffs panel. Your own debuffs land as you cast them.",
			InputType: proto.InputType_Bool,
			BoolValue: true,
		},
		{
			Label:     "Debuff delay",
			Tooltip:   "Time after each spawn before the add(s) get the debuffs from the Debuffs panel (in seconds)",
			InputType: proto.InputType_Number,
			ShownBy:   "Apply debuffs to adds",
		},
		{
			Label:     "Debuff jitter",
			Tooltip:   "Random change to the debuff delay, up to this much earlier or later, for each add at each spawn (in seconds). The debuffs never land before the spawn.",
			InputType: proto.InputType_Number,
			ShownBy:   "Apply debuffs to adds",
		},
	}
}

func makeDynamicAddsAI(isBoss bool) core.AIFactory {
	return func() core.TargetAI {
		return &DynamicAddsAI{
			isBoss: isBoss,
		}
	}
}

type DynamicAddsAI struct {
	Target     *core.Target
	AddTargets []*core.Target // Every target after the first; the boss spawns and despawns them.

	// Static parameters associated with a given preset
	isBoss bool

	respawnTime time.Duration
	addLifetime time.Duration
	spawnDelay  time.Duration
	killShare   float64 // 0..1, the chance the raid gets the kill of each despawning add.

	applyDebuffs bool // Whether the adds get the raid's debuffs, debuffDelay ± debuffJitter after each spawn; on for a boss saved without the input.
	debuffDelay  time.Duration
	debuffJitter time.Duration

	lifeStartAt time.Duration // The start of the adds' current life, or of their next one while they are down.
}

func (ai *DynamicAddsAI) Initialize(target *core.Target, config *proto.Target) {
	ai.Target = target
	ai.Target.AutoAttacks.MHConfig().ActionID.Tag = core.TernaryInt32(ai.isBoss, dynamicBossID, dynamicAddID)

	ai.AddTargets = target.Env.Encounter.AllTargets[1:]

	if ai.isBoss {
		for _, addTarget := range ai.AddTargets {
			addTarget.Schedule = ai
		}
	}

	killSharePercent := float64(dynamicAddKillSharePercent)
	if ai.isBoss && len(config.TargetInputs) >= 3 {
		ai.addLifetime = core.DurationFromSeconds(config.TargetInputs[1].NumberValue)
		ai.respawnTime = core.DurationFromSeconds(config.TargetInputs[0].NumberValue)
		ai.spawnDelay = core.DurationFromSeconds(config.TargetInputs[2].NumberValue)
	}
	if ai.isBoss && len(config.TargetInputs) >= 4 {
		killSharePercent = config.TargetInputs[3].NumberValue
	}
	ai.killShare = core.Clamp(killSharePercent/100, 0, 1)

	ai.applyDebuffs = true
	if ai.isBoss && len(config.TargetInputs) >= 5 {
		ai.applyDebuffs = config.TargetInputs[4].BoolValue
	}
	if ai.isBoss && len(config.TargetInputs) >= 6 {
		ai.debuffDelay = core.DurationFromSeconds(max(config.TargetInputs[5].NumberValue, 0))
	}
	if ai.isBoss && len(config.TargetInputs) >= 7 {
		ai.debuffJitter = core.DurationFromSeconds(max(config.TargetInputs[6].NumberValue, 0))
	}
}

func (ai *DynamicAddsAI) NextActiveAt(_ *core.Simulation) time.Duration {
	return ai.lifeStartAt
}

func (ai *DynamicAddsAI) DiesAt(_ *core.Simulation) time.Duration {
	return ai.lifeStartAt + ai.addLifetime
}

func (ai *DynamicAddsAI) Reset(sim *core.Simulation) {
	ai.Target.AutoAttacks.RandomizeMeleeTiming(sim)

	// An encounter can hold the add without the boss, or the boss without an add: then there is nothing to spawn.
	if !ai.isBoss || len(ai.AddTargets) == 0 {
		return
	}

	for _, addTarget := range ai.AddTargets {
		addTarget.Disable(sim, true)
		addTarget.DisableRaidDebuffs(sim)
	}

	if ai.spawnDelay > 0 {
		ai.lifeStartAt = ai.spawnDelay
		pa := sim.GetConsumedPendingActionFromPool()
		pa.NextActionAt = ai.spawnDelay
		pa.Priority = core.ActionPriorityDOT
		pa.OnAction = func(sim *core.Simulation) {
			ai.spawnAdds(sim)
		}
		sim.AddPendingAction(pa)
	} else {
		ai.spawnAdds(sim)
	}
}

func (ai *DynamicAddsAI) spawnAdds(sim *core.Simulation) {
	for _, addTarget := range ai.AddTargets {
		addTarget.Enable(sim)
		sim.Encounter.ScheduleAddRaidDebuffs(sim, addTarget, ai.applyDebuffs, ai.debuffDelay, ai.debuffJitter)
	}

	if sim.Log != nil {
		sim.Log("Spawned %d adds at %s.", len(ai.AddTargets), sim.CurrentTime)
	}

	ai.lifeStartAt = sim.CurrentTime
	pa := sim.GetConsumedPendingActionFromPool()
	pa.NextActionAt = ai.lifeStartAt + ai.addLifetime
	pa.Priority = core.ActionPriorityDOT
	pa.OnAction = func(sim *core.Simulation) {
		ai.despawnAdds(sim)
	}
	sim.AddPendingAction(pa)
}

func (ai *DynamicAddsAI) despawnAdds(sim *core.Simulation) {
	for _, addTarget := range ai.AddTargets {
		addTarget.Die(sim, ai.killShare)
		addTarget.DisableRaidDebuffs(sim)
	}

	if sim.Log != nil {
		sim.Log("Despawned %d adds at %s.", len(ai.AddTargets), sim.CurrentTime)
	}

	// Only schedule next spawn if there is a respawn time
	if ai.respawnTime > 0 {
		ai.lifeStartAt = sim.CurrentTime + ai.respawnTime
		pa := sim.GetConsumedPendingActionFromPool()
		pa.NextActionAt = ai.lifeStartAt
		pa.Priority = core.ActionPriorityDOT
		pa.OnAction = func(sim *core.Simulation) {
			ai.spawnAdds(sim)
		}
		sim.AddPendingAction(pa)
	} else {
		ai.spawnAdds(sim)
	}
}

func (ai *DynamicAddsAI) ExecuteCustomRotation(sim *core.Simulation) {
	ai.Target.ExtendGCDUntil(sim, sim.CurrentTime+core.BossGCD)
}
