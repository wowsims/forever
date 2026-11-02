import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Encounter, EquipmentSpec, HealingModel, ItemSwap, Spec } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { translatePresetConfigurationCategory } from '@i18n/localization';
import { PresetConfigurationCategory } from '@sim/constants/preset_categories';
import { isEqualAPLRotation } from '@sim/proto/apl_utils';
import { Stats } from '@sim/proto/stats';
import type { IndividualSimHost } from '@sim/sim_host';

import type { PresetBuild } from './preset_utils';

// Keys the walk below must not name: `phase`/`group` are grouping metadata rather than settings,
// and the rest are re-added by their own rules so they read once rather than under two labels.
const NON_CATEGORY_BUILD_KEYS = ['name', 'encounter', 'settings', 'phase', 'group', 'epWeights', 'reforgeSettings'];

/** The categories a build touches, as the chip tooltip lists them: de-duplicated and sorted. */
export function buildCategories(build: PresetBuild): Array<string> {
	const categories: Array<string> = [];

	Object.keys(build).forEach(key => {
		if (!NON_CATEGORY_BUILD_KEYS.includes(key) && build[key as PresetConfigurationCategory]) {
			categories.push(translatePresetConfigurationCategory(key as PresetConfigurationCategory));
		}
	});

	if (build.encounter?.encounter) categories.push(translatePresetConfigurationCategory(PresetConfigurationCategory.Encounter));
	if (build.epWeights) categories.push(i18n.t('common.preset.stat_weights'));
	if (build.reforgeSettings) categories.push(i18n.t('common.preset.reforge_settings'));

	if (build.settings) {
		Object.keys(build.settings).forEach(key => {
			if (key === 'name') return;
			if (key === 'specOptions') categories.push(i18n.t('common.preset.class_spec_options'));
			else if (key === 'consumables') categories.push(i18n.t('common.preset.consumables'));
			else if (key === 'reforgeSettings') categories.push(i18n.t('common.preset.reforge_settings'));
			else if (key === 'debuffs') categories.push(i18n.t('common.preset.debuffs'));
			else if (['buffs', 'raidBuffs', 'partyBuffs'].includes(key)) categories.push(i18n.t('common.preset.buffs'));
			else categories.push(i18n.t('common.preset.other_settings'));
		});
	}

	return [...new Set(categories)].sort();
}

/**
 * Whether the live settings already match the build, scoped to the categories the picker showing
 * the chip cares about: a gear-tab chip lights up on gear alone, even when the build also carries
 * talents the player has not loaded.
 */
export function isBuildActive(
	{ gear, itemSwap, rotation, rotationType, talents, epWeights, encounter, settings }: PresetBuild,
	simUI: IndividualSimHost<Spec>,
	categories?: Array<PresetConfigurationCategory>,
): boolean {
	const checks = (category: PresetConfigurationCategory) => !categories || categories.includes(category);

	const hasGear = checks(PresetConfigurationCategory.Gear) && gear ? EquipmentSpec.equals(gear.gear, simUI.player.getGear().asSpec()) : true;
	// A build's swap set goes with its gear, and loading the build turns item swap on with it.
	const hasSwapSet =
		checks(PresetConfigurationCategory.Gear) && itemSwap
			? simUI.player.itemSwapSettings.getEnableItemSwap() && isEqualItemSwap(simUI.player.itemSwapSettings.toProto(), itemSwap.itemSwap)
			: true;
	const hasTalents =
		checks(PresetConfigurationCategory.Talents) && talents
			? SavedTalents.equals(talents.data, SavedTalents.create({ talentsString: simUI.player.getTalentsString() }))
			: true;

	let hasRotation = true;
	if (checks(PresetConfigurationCategory.Rotation)) {
		if (rotationType) {
			hasRotation = rotationType === simUI.player.getRotationType();
		} else if (rotation?.rotation.rotation) {
			hasRotation = isEqualAPLRotation(simUI.player, simUI.player.getResolvedAplRotation(), rotation.rotation.rotation);
		}
	}

	const hasEpWeights = checks(PresetConfigurationCategory.EPWeights) && epWeights ? simUI.player.getEpWeights().equals(epWeights.epWeights) : true;

	let hasEncounter = true;
	let hasHealingModel = true;
	if (checks(PresetConfigurationCategory.Encounter)) {
		hasEncounter = encounter?.encounter
			? Encounter.equals({ ...encounter.encounter, apiVersion: 0 }, { ...simUI.sim.encounter.toProto(), apiVersion: 0 })
			: true;
		hasHealingModel = encounter?.healingModel ? HealingModel.equals(encounter.healingModel, simUI.player.getHealingModel()) : true;
	}

	let hasRace = true;
	let hasProfession1 = true;
	let hasProfession2 = true;
	let hasDistanceFromTarget = true;
	let hasEnableItemSwap = true;
	let hasItemSwap = true;
	let hasSpecOptions = true;
	let hasConsumables = true;
	let hasPartyBuffs = true;
	let hasRaidBuffs = true;
	let hasBuffs = true;
	let hasDebuffs = true;
	if (checks(PresetConfigurationCategory.Settings)) {
		hasRace = settings?.race ? simUI.player.getRace() === settings.race : true;
		hasProfession1 = settings?.playerOptions?.profession1 === undefined || simUI.player.getProfession1() === settings.playerOptions.profession1;
		hasProfession2 = settings?.playerOptions?.profession2 === undefined || simUI.player.getProfession2() === settings.playerOptions.profession2;
		hasDistanceFromTarget =
			settings?.playerOptions?.distanceFromTarget === undefined || simUI.player.getDistanceFromTarget() === settings.playerOptions.distanceFromTarget;
		hasEnableItemSwap =
			settings?.playerOptions?.enableItemSwap === undefined ||
			simUI.player.itemSwapSettings.getEnableItemSwap() === settings.playerOptions.enableItemSwap;
		hasItemSwap =
			settings?.playerOptions?.itemSwap === undefined ||
			(!settings?.playerOptions?.enableItemSwap && !simUI.player.itemSwapSettings.getEnableItemSwap()) ||
			isEqualItemSwap(simUI.player.itemSwapSettings.toProto(), settings.playerOptions.itemSwap);
		hasSpecOptions =
			settings?.specOptions && Object.keys(settings.specOptions).length
				? JSON.stringify(simUI.player.getSpecOptions()) == JSON.stringify(settings.specOptions)
				: true;
		hasConsumables = settings?.consumables ? ConsumesSpec.equals(simUI.player.getConsumes(), settings.consumables) : true;
		hasPartyBuffs = settings?.partyBuffs ? PartyBuffs.equals(simUI.player.getParty()?.getBuffs(), settings.partyBuffs) : true;
		hasRaidBuffs = settings?.raidBuffs ? RaidBuffs.equals(simUI.sim.raid.getBuffs(), settings.raidBuffs) : true;
		hasBuffs = settings?.buffs ? IndividualBuffs.equals(simUI.player.getBuffs(), settings.buffs) : true;
		hasDebuffs = settings?.debuffs ? Debuffs.equals(simUI.sim.raid.getDebuffs(), settings.debuffs) : true;
	}

	return (
		hasGear &&
		hasTalents &&
		hasRotation &&
		hasEpWeights &&
		hasEncounter &&
		hasHealingModel &&
		hasRace &&
		hasProfession1 &&
		hasProfession2 &&
		hasDistanceFromTarget &&
		hasEnableItemSwap &&
		hasItemSwap &&
		hasSwapSet &&
		hasSpecOptions &&
		hasConsumables &&
		hasPartyBuffs &&
		hasRaidBuffs &&
		hasBuffs &&
		hasDebuffs
	);
}

/**
 * Compares the items and the prepull bonus stats by value: the live swap always carries a prepull stats
 * message, while a preset's JSON usually leaves it out, and neither apiVersion is part of the setting.
 */
function isEqualItemSwap(live: ItemSwap, preset: ItemSwap): boolean {
	return (
		ItemSwap.equals(ItemSwap.create({ items: live.items }), ItemSwap.create({ items: preset.items })) &&
		Stats.fromProto(live.prepullBonusStats).equals(Stats.fromProto(preset.prepullBonusStats))
	);
}
