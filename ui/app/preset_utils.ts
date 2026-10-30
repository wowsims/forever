import { Player as PlayerProto } from '@generated/proto/api';
import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import {
	ConsumesSpec,
	Cooldowns,
	Encounter as EncounterProto,
	EquipmentSpec,
	HealingModel,
	ItemSwap,
	PseudoStat,
	Spec,
	Stat,
	UnitReference,
} from '@generated/proto/common';
import { HealingSimSettings, IndividualSimSettings, SavedRotation, SavedTalents } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { Player } from '@sim/player/player';
import type {
	PresetBuild,
	PresetBuildOptions,
	PresetEncounter,
	PresetEncounterOptions,
	PresetEpWeights,
	PresetEpWeightsOptions,
	PresetGear,
	PresetGearOptions,
	PresetItemSwap,
	PresetOptionsBase,
	PresetRotation,
	PresetRotationOptions,
	PresetSettings,
	PresetTalents,
	PresetTalentsOptions,
} from '@sim/presets/types';
import { specTypeFunctions } from '@sim/proto/spec_functions';
import type { SpecRotation, SpecTypeFunctions } from '@sim/proto/spec_types';
import { Stats } from '@sim/proto/stats';
import { getPlayerSpecFromPlayer } from '@sim/proto/utils';
import type { IndividualSimUIConfig } from '@sim/spec_config';

export type {
	PresetBase,
	PresetBuild,
	PresetBuildOptions,
	PresetEncounter,
	PresetEncounterOptions,
	PresetEpWeights,
	PresetEpWeightsOptions,
	PresetGear,
	PresetGearOptions,
	PresetItemSwap,
	PresetOptionsBase,
	PresetRotation,
	PresetRotationOptions,
	PresetSettings,
	PresetTalents,
	PresetTalentsOptions,
} from '@sim/presets/types';

export const makePresetGear = (name: string, gearJson: any, options?: PresetGearOptions): PresetGear => {
	const gear = EquipmentSpec.fromJson(gearJson);
	return makePresetGearHelper(name, gear, options || {});
};

const makePresetGearHelper = (name: string, gear: EquipmentSpec, options: PresetGearOptions): PresetGear => {
	const conditions: Array<(player: Player<any>) => boolean> = [];

	if (options.faction !== undefined) {
		conditions.push((player: Player<any>) => player.getFaction() == options.faction);
	}
	if (options.customCondition !== undefined) {
		conditions.push(options.customCondition);
	}

	return {
		name,
		tooltip: options.tooltip || i18n.t('sim.basic_bis_disclaimer'),
		gear,
		enableWhen: !!conditions.length ? (player: Player<any>) => conditions.every(cond => cond(player)) : undefined,
		onLoad: options?.onLoad,
		// The gear tab groups its preset chips by phase and named group.
		phase: options.phase,
		group: options.group,
		devOnly: options.devOnly,
	};
};

export const makePresetTalents = (name: string, data: SavedTalents, options?: PresetTalentsOptions): PresetTalents => {
	const conditions: Array<(player: Player<any>) => boolean> = [];
	if (options && options.customCondition) {
		conditions.push(options.customCondition);
	}

	return {
		name,
		data,
		enableWhen: conditions.length > 0 ? (player: Player<any>) => conditions.every(cond => cond(player)) : undefined,
	};
};

export const makePresetEpWeights = (name: string, epWeights: Stats, options?: PresetEpWeightsOptions): PresetEpWeights => {
	return makePresetEpWeightHelper(name, epWeights, options || {});
};

const makePresetEpWeightHelper = (name: string, epWeights: Stats, options?: PresetEpWeightsOptions): PresetEpWeights => {
	const conditions: Array<(player: Player<any>) => boolean> = [];
	if (options?.customCondition !== undefined) {
		conditions.push(options.customCondition);
	}

	return {
		name,
		epWeights,
		enableWhen: !!conditions.length ? (player: Player<any>) => conditions.every(cond => cond(player)) : undefined,
		onLoad: options?.onLoad,
	};
};

// JSON shape for an EP preset read from a file. Enum keys are stored as names (e.g. "StatCritRating") so
// they stay stable across proto regenerations. No spec reads one: spec EP weights are TS literals in presets.ts.
export type PresetEpWeightsJson = {
	name: string;
	stats?: Partial<Record<keyof typeof Stat, number>>;
	pseudoStats?: Partial<Record<keyof typeof PseudoStat, number>>;
};

export const makePresetEpWeightsFromJSON = (json: PresetEpWeightsJson, options?: PresetEpWeightsOptions): PresetEpWeights => {
	// A key that is not an enum member resolves to `undefined` and would otherwise silently drop
	// that weight — TypeScript cannot check an imported .json's key strings against the enum.
	const statsMap: Partial<Record<Stat, number>> = {};
	Object.entries(json.stats ?? {}).forEach(([key, value]) => {
		const stat: Stat | undefined = Stat[key as keyof typeof Stat];
		if (stat === undefined) throw new Error(`Unknown stat '${key}' in EP preset '${json.name}'.`);
		statsMap[stat] = value;
	});

	const pseudoStatsMap: Partial<Record<PseudoStat, number>> = {};
	Object.entries(json.pseudoStats ?? {}).forEach(([key, value]) => {
		const pseudoStat: PseudoStat | undefined = PseudoStat[key as keyof typeof PseudoStat];
		if (pseudoStat === undefined) throw new Error(`Unknown pseudo stat '${key}' in EP preset '${json.name}'.`);
		pseudoStatsMap[pseudoStat] = value;
	});

	return makePresetEpWeights(json.name, Stats.fromMap(statsMap, pseudoStatsMap), options);
};

export const makePresetAPLRotation = (name: string, rotationJson: any, options?: PresetRotationOptions): PresetRotation => {
	const rotation = SavedRotation.create({
		rotation: APLRotation.fromJson(rotationJson),
	});

	return makePresetRotationHelper(name, rotation, options);
};

export const makePresetSimpleRotation = <SpecType extends Spec>(
	name: string,
	spec: SpecType,
	simpleRotation: SpecRotation<SpecType>,
	options?: PresetRotationOptions,
): PresetRotation => {
	const isTankSpec = spec == Spec.SpecFeralBearDruid || spec == Spec.SpecProtectionPaladin || spec == Spec.SpecProtectionWarrior;
	const rotation = SavedRotation.create({
		rotation: {
			type: APLRotationType.TypeSimple,
			simple: {
				specRotationJson: JSON.stringify(specTypeFunctions[spec].rotationToJson(simpleRotation)),
				cooldowns: Cooldowns.create({
					hpPercentForDefensives: isTankSpec ? 0.4 : 0,
				}),
			},
		},
	});

	return makePresetRotationHelper(name, rotation, options);
};

const makePresetRotationHelper = (name: string, rotation: SavedRotation, options?: PresetRotationOptions): PresetRotation => {
	const conditions: Array<(player: Player<any>) => boolean> = [];
	if (options?.talents != undefined) {
		conditions.push((player: Player<any>) => (options.talents || []).join('') === player.getTalentTreePoints().join(''));
	}
	return {
		name,
		rotation,
		enableWhen: !!conditions.length ? (player: Player<any>) => conditions.every(cond => cond(player)) : undefined,
		onLoad: options?.onLoad,
	};
};

export const makePresetEncounter = (
	name: string,
	encounter?: EncounterProto,
	healingModel?: HealingModel,
	tanks?: UnitReference[],
	targetDummies?: number,
	healingSettings?: HealingSimSettings,
	options?: PresetEncounterOptions,
): PresetEncounter => {
	return {
		name,
		encounter,
		targetDummies,
		healingSettings,
		tanks,
		healingModel,
		...options,
	};
};

export const makePresetItemSwapGear = (name: string, itemSwapJson: any): PresetItemSwap => {
	const itemSwap = ItemSwap.fromJson(itemSwapJson);
	return makePresetItemSwapGearHelper(name, itemSwap);
};

export const makePresetItemSwapGearHelper = (name: string, itemSwap: ItemSwap): PresetItemSwap => {
	return {
		name,
		itemSwap,
	};
};

// Whether the settings set the spec's options. A SIML `settings:` section always does (sim/siml/SPEC.md §5.1), so
// without them a file states none of the player's settings, as a build of only gear, talents and a rotation.
const hasSpecOptions = ({ spec }: PlayerProto) =>
	spec.oneofKind !== undefined && (spec as unknown as Record<string, { options?: unknown }>)[spec.oneofKind].options !== undefined;

const makePresetSettingsHelper = (name: string, spec: Spec, simSettings: IndividualSimSettings): PresetSettings => {
	const settings: PresetSettings = { name };

	if (simSettings.player) {
		if (simSettings.player.buffs) {
			settings.buffs = simSettings.player.buffs;
		}

		if (simSettings.player.consumables) {
			settings.consumables = simSettings.player.consumables;
		}
	}

	if (simSettings.player && hasSpecOptions(simSettings.player)) {
		if (simSettings.player.race) {
			settings.race = simSettings.player.race;
		}

		settings.specOptions = specTypeFunctions[spec].optionsFromPlayer(simSettings.player);

		settings.playerOptions = {
			inFrontOfTarget: simSettings.player.inFrontOfTarget,
			distanceFromTarget: simSettings.player.distanceFromTarget,
			// Settings with no `item_swap:` section turn swapping off.
			enableItemSwap: simSettings.player.enableItemSwap,
		};

		// proto3 cannot tell an absent scalar from a zero one, so a JSON that never mentions these
		// timings parses to 0 for both. Leaving the field out here is what lets `applyBuild` keep a
		// `typeof === 'number'` guard and so still honour a hand-written preset that means 0.
		if (simSettings.player.reactionTimeMs) {
			settings.playerOptions.reactionTimeMs = simSettings.player.reactionTimeMs;
		}

		if (simSettings.player.channelClipDelayMs) {
			settings.playerOptions.channelClipDelayMs = simSettings.player.channelClipDelayMs;
		}

		if (!!simSettings.player.profession1) {
			settings.playerOptions.profession1 = simSettings.player.profession1;
		}

		if (!!simSettings.player.profession2) {
			settings.playerOptions.profession2 = simSettings.player.profession2;
		}
	}

	// An item swap set is gear, not a setting (sim/siml/SPEC.md §5.1), so a file states it with or without them.
	if (simSettings.player?.itemSwap) {
		settings.playerOptions = { ...settings.playerOptions, enableItemSwap: simSettings.player.enableItemSwap, itemSwap: simSettings.player.itemSwap };
	}

	if (simSettings.raidBuffs) {
		settings.raidBuffs = simSettings.raidBuffs;
	}

	if (simSettings.partyBuffs) {
		settings.partyBuffs = simSettings.partyBuffs;
	}

	if (simSettings.debuffs) {
		settings.debuffs = simSettings.debuffs;
	}

	return settings;
};

export const makePresetBuild = (name: string, options: PresetBuildOptions): PresetBuild => {
	return { name, ...options };
};

export const makePresetBuildFromJSON = (
	name: string,
	spec: Spec,
	json: any,
	{ settings: customSimSettings, ...customBuildOptions }: PresetBuildOptions = {},
	options?: PresetOptionsBase,
): PresetBuild => {
	const simSettings = IndividualSimSettings.fromJson(json, { ignoreUnknownFields: true });
	const buildConfig: PresetBuildOptions = {};

	if (simSettings.player) {
		if (simSettings.player.equipment) {
			buildConfig.gear = makePresetGear(name, simSettings.player.equipment, options);
		}

		if (simSettings.player?.talentsString) {
			buildConfig.talents = makePresetTalents(name, SavedTalents.create({ talentsString: simSettings.player?.talentsString }), options);
		}

		if (simSettings.player?.rotation?.type === APLRotationType.TypeAuto) {
			buildConfig.rotationType = APLRotationType.TypeAuto;
		} else if (simSettings.player?.rotation) {
			buildConfig.rotation = makePresetRotationHelper(name, SavedRotation.create({ rotation: simSettings.player.rotation }), options);
		}
	}

	if (simSettings.encounter) {
		buildConfig.encounter = makePresetEncounter(
			name,
			simSettings.encounter,
			simSettings.player?.healingModel,
			simSettings.tanks,
			simSettings.targetDummies,
			simSettings.healingSettings,
			options,
		);
	}

	const settings = makePresetSettingsHelper(name, spec, simSettings);
	if (Object.keys(settings).length > 1 || customSimSettings) {
		buildConfig.settings = { ...settings, ...customSimSettings };
	}

	if (simSettings.epWeightsStats) {
		buildConfig.epWeights = makePresetEpWeightHelper(name, Stats.fromProto(simSettings.epWeightsStats), options);
	}

	if (simSettings.reforgeSettings) {
		buildConfig.reforgeSettings = simSettings.reforgeSettings;
	}

	return makePresetBuild(name, { ...buildConfig, ...customBuildOptions });
};

type SpecSettingsDefaults<SpecType extends Spec> = Pick<
	IndividualSimUIConfig<SpecType>['defaults'],
	| 'gear'
	| 'consumables'
	| 'talents'
	| 'specOptions'
	| 'raidBuffs'
	| 'partyBuffs'
	| 'individualBuffs'
	| 'debuffs'
	| 'rotationType'
	| 'aplRotation'
	| 'simpleRotation'
	| 'other'
	| 'encounter'
	| 'itemSwap'
	| 'settings'
> &
	Partial<Pick<IndividualSimUIConfig<SpecType>['defaults'], 'epWeights'>>;

// The proto-backed part of a spec's `defaults`, from the IndividualSimSettings JSON of its defaults.siml. A message the
// file does not hold is empty, and `other` holds 0 for what the file does not set, which applying the defaults reads
// as unset. `settings` keeps the whole file, so applying the defaults also reads what has no field of its own here.
export const specDefaultsFromSettings = <SpecType extends Spec>(spec: SpecType, json: any): SpecSettingsDefaults<SpecType> => {
	const settings = IndividualSimSettings.fromJson(json);
	const player = settings.player;
	if (!player?.spec.oneofKind || getPlayerSpecFromPlayer(player).specID !== spec) {
		throw new Error(`The defaults are not ${Spec[spec]} settings.`);
	}
	const functions = specTypeFunctions[spec] as SpecTypeFunctions<SpecType>;
	const rotation = player.rotation;

	return {
		gear: player.equipment ?? EquipmentSpec.create(),
		consumables: player.consumables ?? ConsumesSpec.create(),
		talents: SavedTalents.create({ talentsString: player.talentsString }),
		specOptions: functions.optionsFromPlayer(player),
		raidBuffs: settings.raidBuffs ?? RaidBuffs.create(),
		partyBuffs: settings.partyBuffs ?? PartyBuffs.create(),
		individualBuffs: player.buffs ?? IndividualBuffs.create(),
		debuffs: settings.debuffs ?? Debuffs.create(),
		...(rotation && { rotationType: rotation.type }),
		...(rotation?.type === APLRotationType.TypeAPL && { aplRotation: rotation }),
		...(rotation?.type === APLRotationType.TypeSimple && {
			simpleRotation: functions.rotationFromJson(JSON.parse(rotation.simple?.specRotationJson || '{}')),
		}),
		...(settings.encounter && {
			encounter: makePresetEncounter('Default', settings.encounter, undefined, undefined, settings.targetDummies),
		}),
		...(player.enableItemSwap && player.itemSwap && { itemSwap: player.itemSwap }),
		...(settings.epWeightsStats && { epWeights: Stats.fromProto(settings.epWeightsStats) }),
		settings,
		other: {
			race: player.race,
			profession1: player.profession1,
			profession2: player.profession2,
			distanceFromTarget: player.distanceFromTarget,
			reactionTime: player.reactionTimeMs,
			channelClipDelay: player.channelClipDelayMs,
			healingModel: player.healingModel,
			iterationCount: settings.settings?.iterations,
			inFrontOfTarget: player.inFrontOfTarget,
			tanks: settings.tanks,
		},
	};
};
