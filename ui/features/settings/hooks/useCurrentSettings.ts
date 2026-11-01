import { readSavedSettings } from '@features/settings/model/saved_settings';
import { useSimHost } from '@sim/context/SimHostContext';
import { useReadyStoreSubscribe } from '@sim/hooks/useReadyStoreSubscribe';
import { useSimReady } from '@sim/hooks/useSimReady';
import { subscribeAll, subscribePartyBuffs, subscribePlayerField, subscribeRaidField } from '@sim/state/subscriptions';

// The live settings as the Saved Settings panel saves them; null until the sim is ready.
export const useCurrentSettings = () => {
	const host = useSimHost();
	const ready = useSimReady();

	return useReadyStoreSubscribe(
		subscribeAll([
			subscribeRaidField(host.sim.raid, 'buffs'),
			subscribeRaidField(host.sim.raid, 'debuffs'),
			subscribePartyBuffs(host.player.getParty()!),
			subscribePlayerField(host.player, 'buffs'),
			subscribePlayerField(host.player, 'consumables'),
			subscribePlayerField(host.player, 'race'),
			subscribePlayerField(host.player, 'profession1'),
			subscribePlayerField(host.player, 'profession2'),
			subscribePlayerField(host.player, 'itemSwap'),
			subscribePlayerField(host.player, 'reactionTime'),
			subscribePlayerField(host.player, 'channelClipDelay'),
			subscribePlayerField(host.player, 'inFrontOfTarget'),
			subscribePlayerField(host.player, 'distanceFromTarget'),
			subscribePlayerField(host.player, 'healingModel'),
			subscribePlayerField(host.player, 'specOptions'),
		]),
		// Reaches `player.getConsumes()`, so it dereferences a null `sim.db` before the sim is ready.
		() => readSavedSettings(host),
		ready,
	);
};
