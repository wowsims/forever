import { HealingSimSettings, HealingSimSettings_Assignment as HealingAssignment } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import type { Sim } from '@sim/sim';
import { subscribeAll, subscribePlayerChange, subscribeRaidField } from '@sim/state/subscriptions';

export const healingGroupSize = ({ groupSize }: HealingSimSettings): number => (groupSize === 5 || groupSize === 20 ? groupSize : 10);

export const healingTargetDummies = (settings: HealingSimSettings): number => {
	const groupSize = healingGroupSize(settings);
	if (groupSize === 5) return 4;
	if (settings.assignment === HealingAssignment.TankHealing) return groupSize === 20 ? 2 : 1;
	return groupSize - 1;
};

/** Healing assignments determine the count; other specs can use dummies only while their talent or item allows them. */
export const watchTargetDummies = (player: Player<any>, sim: Sim): void => {
	if (!player.canEnableTargetDummies()) return;

	const apply = () => {
		if (player.getPlayerSpec().isHealingSpec) sim.raid.setTargetDummies(healingTargetDummies(sim.raid.getHealingSettings()));
		else if (!player.shouldEnableTargetDummies() && sim.raid.getTargetDummies() !== 0) sim.raid.setTargetDummies(0);
	};

	// `targetDummies` rides in the shared link and the saved settings, so a count can arrive from a
	// spec that could use it into one that cannot. Armed after `loadSettings`, which applies the
	// whole envelope in one batch, so the rule sees the loaded state.
	apply();
	subscribeAll([subscribePlayerChange(player), subscribeRaidField(sim.raid, 'healingSettings')])(apply);
};
