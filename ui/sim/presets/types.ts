import { Player as PlayerProto, ReforgeSettings } from '@generated/proto/api';
import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Encounter as EncounterProto, EquipmentSpec, Faction, HealingModel, ItemSwap, Race, UnitReference } from '@generated/proto/common';
import type { HealingSimSettings } from '@generated/proto/ui';
import { SavedRotation, SavedTalents } from '@generated/proto/ui';

import type { Phase } from '../constants/other';
import type { Player } from '../player/player';
import type { SpecOptions } from '../proto/spec_types';
import type { Stats } from '../proto/stats';

export interface PresetBase {
	name: string;
	tooltip?: string;
	enableWhen?: (obj: Player<any>) => boolean;
	onLoad?: (player: Player<any>) => void;
	// Grouping read by ui/app's preset group picker: which content phase a
	// preset belongs to, and an optional named section within that phase.
	phase?: Phase;
	group?: string;
}

export interface PresetOptionsBase extends Pick<PresetBase, 'onLoad'> {
	customCondition?: (player: Player<any>) => boolean;
	phase?: Phase;
	group?: string;
}

export interface PresetGear extends PresetBase {
	gear: EquipmentSpec;
	// Shown only on the dev server and the pull request previews, see `isPreviewOrDevMode`.
	devOnly?: boolean;
}
export interface PresetGearOptions extends PresetOptionsBase, Pick<PresetBase, 'tooltip'>, Pick<PresetGear, 'devOnly'> {
	faction?: Faction;
}

export interface PresetTalents {
	name: string;
	data: SavedTalents;
	enableWhen?: (obj: Player<any>) => boolean;
}

export interface PresetTalentsOptions {
	customCondition?: (player: Player<any>) => boolean;
}

export interface PresetRotation extends PresetBase {
	rotation: SavedRotation;
}
export interface PresetRotationOptions extends Pick<PresetOptionsBase, 'onLoad'> {
	talents?: number[];
}

export interface PresetEpWeights extends PresetBase {
	epWeights: Stats;
}
export interface PresetEpWeightsOptions extends PresetOptionsBase {}

export interface PresetItemSwap extends PresetBase {
	itemSwap: ItemSwap;
}

export interface PresetEncounter extends PresetBase {
	encounter?: EncounterProto;
	healingModel?: HealingModel;
	tanks?: UnitReference[];
	targetDummies?: number;
	healingSettings?: HealingSimSettings;
}
export interface PresetEncounterOptions extends PresetOptionsBase {}

type PresetPlayerOptions = Partial<
	Pick<
		PlayerProto,
		'reactionTimeMs' | 'channelClipDelayMs' | 'inFrontOfTarget' | 'distanceFromTarget' | 'profession1' | 'profession2' | 'enableItemSwap' | 'itemSwap'
	>
>;

export interface PresetSettings extends PresetBase {
	race?: Race;
	raidBuffs?: RaidBuffs;
	partyBuffs?: PartyBuffs;
	buffs?: IndividualBuffs;
	debuffs?: Debuffs;
	consumables?: ConsumesSpec;
	specOptions?: Partial<SpecOptions<any>>;
	playerOptions?: PresetPlayerOptions;
}

export interface PresetBuild {
	name: string;
	phase?: Phase;
	group?: string;
	gear?: PresetGear;
	itemSwap?: PresetItemSwap;
	talents?: PresetTalents;
	rotation?: PresetRotation;
	rotationType?: APLRotationType;
	epWeights?: PresetEpWeights;
	encounter?: PresetEncounter;
	settings?: PresetSettings;
	reforgeSettings?: ReforgeSettings;
}

export interface PresetBuildOptions extends Omit<PresetBuild, 'name'> {}
