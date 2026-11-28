import type { Player } from '@sim/player/player';
import type { Sim } from '@sim/sim';
import { subscribePlayerChange } from '@sim/state/subscriptions';

/** The target-dummy count is the raid's, but whether it may be non-zero is the player's business: a talent or an item swap can take the ability away. */
export const watchTargetDummies = (player: Player<any>, sim: Sim): void => {
	if (!player.canEnableTargetDummies()) return;

	const apply = () => {
		if (!player.shouldEnableTargetDummies() && sim.raid.getTargetDummies() !== 0) sim.raid.setTargetDummies(0);
	};

	// `targetDummies` rides in the shared link and the saved settings, so a count can arrive from a
	// spec that could use it into one that cannot. Armed after `loadSettings`, so the rule sees the
	// loaded state.
	apply();
	subscribePlayerChange(player)(apply);
};
