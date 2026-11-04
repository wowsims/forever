import { Debuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { Encounter as EncounterProto } from '@generated/proto/common';
import { IndividualSimSettings } from '@generated/proto/ui';

import { CURRENT_API_VERSION } from '../constants/other';
import { SimSettingCategories } from '../constants/sim_settings';
import type { Player } from '../player/player';
import { Stats } from '../proto/stats';
import type { ReforgeSettings } from '../settings/reforge_settings';
import type { Sim } from '../sim';
import { getEnumValues } from '../utils/collections';
import { batch } from './batch';
// The state surface the envelope serializes besides Player/Sim: the reforge
// settings model and the EP reference-stat selections owned by the sim UI.
export interface IndividualSimSerializationContext {
	player: Player<any>;
	sim: Sim;
	reforgeSettings?: ReforgeSettings;
	// Fallback EP weights applied when a loaded proto carries none.
	defaultEpWeights: Stats;
}

export function individualSimSettingsToProto(ctx: IndividualSimSerializationContext, exportCategories?: Array<SimSettingCategories>): IndividualSimSettings {
	const exportCategory = (cat: SimSettingCategories) => !exportCategories || exportCategories.length == 0 || exportCategories.includes(cat);

	const proto = IndividualSimSettings.create({
		player: ctx.player.toProto(true, false, exportCategories),
		apiVersion: CURRENT_API_VERSION,
	});

	if (exportCategory(SimSettingCategories.Miscellaneous)) {
		IndividualSimSettings.mergePartial(proto, {
			tanks: ctx.sim.raid.getTanks(),
		});
	}
	if (exportCategory(SimSettingCategories.Encounter)) {
		IndividualSimSettings.mergePartial(proto, {
			encounter: ctx.sim.encounter.toProto(),
		});
	}
	if (exportCategory(SimSettingCategories.External)) {
		IndividualSimSettings.mergePartial(proto, {
			partyBuffs: ctx.player.getParty()?.getBuffs() || PartyBuffs.create(),
			raidBuffs: ctx.sim.raid.getBuffs(),
			debuffs: ctx.sim.raid.getDebuffs(),
			targetDummies: ctx.sim.raid.getTargetDummies(),
			healingSettings: ctx.sim.raid.getHealingSettings(),
		});
	}
	if (exportCategory(SimSettingCategories.UISettings)) {
		IndividualSimSettings.mergePartial(proto, {
			settings: ctx.sim.toProto(),
			epWeightsStats: ctx.player.getEpWeights().toProto(),
			epRatios: ctx.player.getEpRatios(),
			dpsRefStat: ctx.player.getRefStat('dpsRefStat'),
			healRefStat: ctx.player.getRefStat('healRefStat'),
			tankRefStat: ctx.player.getRefStat('tankRefStat'),
			reforgeSettings: ctx.reforgeSettings?.toProto(),
		});
	}

	return proto;
}

// The categories an export carries, read from one field each that the export always writes. An empty
// talent string reads as no talents exported.
export function individualSimSettingsCategories(settings: IndividualSimSettings): Array<SimSettingCategories> {
	const player = settings.player;
	const present: Record<SimSettingCategories, boolean> = {
		[SimSettingCategories.Gear]: !!player?.equipment,
		[SimSettingCategories.Talents]: !!player?.talentsString,
		[SimSettingCategories.Rotation]: !!player?.rotation,
		[SimSettingCategories.Consumes]: !!player?.consumables,
		[SimSettingCategories.Miscellaneous]: !!player?.healingModel,
		[SimSettingCategories.External]: !!player?.buffs || settings.targetDummies > 0,
		[SimSettingCategories.Encounter]: !!settings.encounter,
		[SimSettingCategories.UISettings]: !!settings.settings,
	};
	return (getEnumValues(SimSettingCategories) as Array<SimSettingCategories>).filter(category => present[category]);
}

export function applyIndividualSimSettings(
	ctx: IndividualSimSerializationContext,
	settings: IndividualSimSettings,
	includeCategories?: Array<SimSettingCategories>,
) {
	const loadCategory = (cat: SimSettingCategories) => !includeCategories || includeCategories.length == 0 || includeCategories.includes(cat);

	const tankSpec = ctx.player.getPlayerSpec().isTankSpec;
	const healingSpec = ctx.player.getPlayerSpec().isHealingSpec;

	batch(() => {
		if (!settings.player) {
			return;
		}

		ctx.player.fromProto(settings.player, includeCategories);

		if (loadCategory(SimSettingCategories.Miscellaneous)) {
			ctx.sim.raid.setTanks(settings.tanks || []);
		}
		if (loadCategory(SimSettingCategories.External)) {
			ctx.sim.raid.setBuffs(settings.raidBuffs || RaidBuffs.create());
			ctx.sim.raid.setDebuffs(settings.debuffs || Debuffs.create());
			const party = ctx.player.getParty();
			if (party) {
				party.setBuffs(settings.partyBuffs || PartyBuffs.create());
			}
			ctx.sim.raid.setTargetDummies(settings.targetDummies);
			if (settings.healingSettings) ctx.sim.raid.setHealingSettings(settings.healingSettings);
		}
		if (loadCategory(SimSettingCategories.Encounter)) {
			ctx.sim.encounter.fromProto(settings.encounter || EncounterProto.create());
		}
		if (loadCategory(SimSettingCategories.UISettings)) {
			if (settings.epWeightsStats) {
				ctx.player.setEpWeights(Stats.fromProto(settings.epWeightsStats));
			} else {
				ctx.player.setEpWeights(ctx.defaultEpWeights);
			}

			const defaultRatios = ctx.player.getDefaultEpRatios(tankSpec, healingSpec);
			if (settings.epRatios) {
				const missingRatios = new Array<number>(defaultRatios.length - settings.epRatios.length).fill(0);
				ctx.player.setEpRatios(settings.epRatios.concat(missingRatios));
			} else {
				ctx.player.setEpRatios(defaultRatios);
			}

			if (settings.reforgeSettings && ctx.reforgeSettings) {
				ctx.reforgeSettings.fromProto(settings.reforgeSettings);
			}

			for (const kind of ['dpsRefStat', 'healRefStat', 'tankRefStat'] as const) {
				if (settings[kind]) ctx.player.setRefStat(kind, settings[kind]);
			}

			if (settings.settings) {
				ctx.sim.fromProto(settings.settings);
			} else {
				ctx.sim.applyDefaults(tankSpec, healingSpec);
			}
		}
	});
}
