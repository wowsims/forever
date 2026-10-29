import { Player } from '@generated/proto/api';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, HealingModel, ItemSwap } from '@generated/proto/common';
import { SavedSettings } from '@generated/proto/ui';
import { isSameSpecArm, withSpec } from '@sim/proto/spec_functions';
import { Stats } from '@sim/proto/stats';
import type { IndividualSimHost } from '@sim/sim_host';
import { batch } from '@sim/state/batch';

const liveSpec = (host: IndividualSimHost<any>): SavedSettings['spec'] => withSpec(host.player.getSpec(), Player.create(), host.player.getSpecOptions()).spec;

export const readSavedSettings = (host: IndividualSimHost<any>): SavedSettings =>
	SavedSettings.create({
		raidBuffs: host.sim.raid.getBuffs(),
		partyBuffs: host.player.getParty()?.getBuffs() || PartyBuffs.create(),
		playerBuffs: host.player.getBuffs(),
		debuffs: host.sim.raid.getDebuffs(),
		consumables: host.player.getConsumes(),
		race: host.player.getRace(),
		professions: host.player.getProfessions(),
		enableItemSwap: host.player.itemSwapSettings.getEnableItemSwap(),
		itemSwap: host.player.itemSwapSettings.toProto(),
		reactionTimeMs: host.player.getReactionTime(),
		channelClipDelayMs: host.player.getChannelClipDelay(),
		inFrontOfTarget: host.player.getInFrontOfTarget(),
		distanceFromTarget: host.player.getDistanceFromTarget(),
		healingModel: host.player.getHealingModel(),
		spec: liveSpec(host),
	});

export const applySavedSettings = (host: IndividualSimHost<any>, settings: SavedSettings): void => {
	batch(() => {
		host.sim.raid.setBuffs(settings.raidBuffs || RaidBuffs.create());
		host.sim.raid.setDebuffs(settings.debuffs || Debuffs.create());
		host.player.getParty()?.setBuffs(settings.partyBuffs || PartyBuffs.create());
		host.player.setBuffs(settings.playerBuffs || IndividualBuffs.create());
		host.player.itemSwapSettings.setItemSwapSettings(
			settings.enableItemSwap,
			host.sim.db.lookupItemSwap(settings.itemSwap || ItemSwap.create()),
			Stats.fromProto(settings.itemSwap?.prepullBonusStats),
		);
		host.player.setConsumes(settings.consumables || ConsumesSpec.create());
		host.player.setRace(settings.race);
		host.player.setProfessions(settings.professions);
		host.player.setReactionTime(settings.reactionTimeMs);
		host.player.setChannelClipDelay(settings.channelClipDelayMs);
		host.player.setInFrontOfTarget(settings.inFrontOfTarget);
		host.player.setDistanceFromTarget(settings.distanceFromTarget);
		host.player.setHealingModel(settings.healingModel || HealingModel.create());
		if (isSameSpecArm(settings.spec, liveSpec(host)))
			host.player.setSpecOptions(host.player.specTypeFunctions.optionsFromPlayer(Player.create({ spec: settings.spec })));
	});
};
