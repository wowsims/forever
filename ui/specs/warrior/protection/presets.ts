import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, HealingModel, Profession, Race } from '@generated/proto/common';
import { ProtectionWarrior_Options as ProtectionWarriorOptions, WarriorStance } from '@generated/proto/warrior';
import { OtherDefaults as SimUIOtherDefaults } from '@sim/spec_config';

import * as WarriorPresets from '../shared/presets';
import GenericApl from './apls/default.apl.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('Generic', GenericApl);

export const DefaultOptions = ProtectionWarriorOptions.create({
	classOptions: {
		queueDelay: 250,
		startingRage: 100,
		useBattleShout: true,
		defaultStance: WarriorStance.WarriorStanceDefensive,
		hasBsT2: true,
		stanceSnapshot: true,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	...WarriorPresets.DefaultConsumables,
	conjuredId: Consumables.MajorHealthstone9421.id,
	foodId: Consumables.DirgesKickinChimaerokChops.id,
	flaskId: undefined,
	elixirIds: [Consumables.ElixirOfTheMongoose.id, Consumables.GiftOfArthas.id],
	potId: Consumables.GreaterStoneshieldPotion.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.ScrollOfStrengthIV.id, Consumables.ScrollOfProtectionIV.id],
});

export const OtherDefaults: Partial<SimUIOtherDefaults> = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceOrc,
	distanceFromTarget: 0,
	healingModel: HealingModel.create({
		hps: 2200,
		cadenceSeconds: 0.4,
		cadenceVariation: 1.2,
		absorbFrac: 0.02,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
};
