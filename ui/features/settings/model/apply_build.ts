import type { PresetBuild } from '@sim/presets/types';
import { Stats } from '@sim/proto/stats';
import type { IndividualSimHost } from '@sim/sim_host';
import { batch } from '@sim/state/batch';

export const applyBuild = (
	{ gear, itemSwap, rotation, rotationType, talents, epWeights, encounter, settings, reforgeSettings }: PresetBuild,
	simUI: IndividualSimHost<any>,
): void => {
	batch(() => {
		if (gear) simUI.player.setGear(simUI.sim.db.lookupEquipmentSpec(gear.gear));
		if (itemSwap) {
			simUI.player.itemSwapSettings.setItemSwapSettings(
				true,
				simUI.sim.db.lookupItemSwap(itemSwap.itemSwap),
				Stats.fromProto(itemSwap.itemSwap.prepullBonusStats),
			);
		} else {
			simUI.player.itemSwapSettings.setEnableItemSwap(false);
		}
		if (talents) {
			simUI.player.setTalentsString(talents.data.talentsString);
		}
		// A build may carry both. Builds are written by hand and routinely carry
		// `rotationType: TypeSimple` *and* a rotation, and taking the first branch there drops the
		// rotation on the floor: the Arms build would leave the Fury spec rotation in place.
		if (rotationType && !rotation?.rotation.rotation) {
			simUI.player.modifyAplRotation(aplRotation => {
				aplRotation.type = rotationType;
			});
		} else if (rotation?.rotation.rotation) {
			// Type first, as vanilla did: it lets a preset that differs from the current rotation
			// only by type short-circuit setAplRotation's equality check.
			if (rotationType) {
				simUI.player.modifyAplRotation(aplRotation => {
					aplRotation.type = rotationType;
				});
			}
			simUI.player.setAplRotation(rotation.rotation.rotation);
		}
		if (epWeights) simUI.player.setEpWeights(epWeights.epWeights);
		if (settings) {
			if (settings.race) simUI.player.setRace(settings.race);
			if (settings.consumables) simUI.player.setConsumes(settings.consumables);
			if (settings.playerOptions?.profession1) simUI.player.setProfession1(settings.playerOptions.profession1);
			if (settings.playerOptions?.profession2) simUI.player.setProfession2(settings.playerOptions.profession2);
			if (typeof settings.playerOptions?.distanceFromTarget === 'number') simUI.player.setDistanceFromTarget(settings.playerOptions.distanceFromTarget);
			// Reachable only when the preset states a timing: `makePresetSettingsHelper` omits these
			// two rather than copying a parsed 0, which a JSON that never mentioned them also yields.
			// The three tank specs' default build is exactly that, and it was wiping their reaction
			// time to 0 ms.
			if (typeof settings.playerOptions?.reactionTimeMs === 'number') simUI.player.setReactionTime(settings.playerOptions.reactionTimeMs);
			if (typeof settings.playerOptions?.channelClipDelayMs === 'number') simUI.player.setChannelClipDelay(settings.playerOptions.channelClipDelayMs);
			if (typeof settings.playerOptions?.inFrontOfTarget === 'boolean') simUI.player.setInFrontOfTarget(settings.playerOptions.inFrontOfTarget);
			if (settings.playerOptions?.enableItemSwap !== undefined && settings.playerOptions?.itemSwap) {
				simUI.player.itemSwapSettings.setItemSwapSettings(
					settings.playerOptions.enableItemSwap,
					simUI.sim.db.lookupItemSwap(settings.playerOptions.itemSwap),
					Stats.fromProto(settings.playerOptions.itemSwap.prepullBonusStats),
				);
			}
			if (settings.specOptions) {
				simUI.player.setSpecOptions({
					...simUI.player.getSpecOptions(),
					...settings.specOptions,
				});
			}
			if (settings.raidBuffs) simUI.sim.raid.setBuffs(settings.raidBuffs);
			if (settings.partyBuffs) simUI.player.getParty()!.setBuffs(settings.partyBuffs);
			if (settings.buffs) simUI.player.setBuffs(settings.buffs);
			if (settings.debuffs) simUI.sim.raid.setDebuffs(settings.debuffs);
		}
		if (encounter) {
			if (encounter.encounter) simUI.sim.encounter.fromProto(encounter.encounter);
			if (encounter.healingModel) simUI.player.setHealingModel(encounter.healingModel);
			if (encounter.tanks) simUI.sim.raid.setTanks(encounter.tanks);
			if (encounter.targetDummies !== undefined) simUI.sim.raid.setTargetDummies(encounter.targetDummies);
			if (encounter.healingSettings) simUI.sim.raid.setHealingSettings(encounter.healingSettings);
		}
		if (reforgeSettings && simUI.reforger) {
			simUI.reforger.applyPresetSettings(reforgeSettings);
		}
	});
};
