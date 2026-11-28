import * as PresetUtils from '@app/preset_utils';
import { Spec } from '@generated/proto/common';

import DefaultApl from './apls/default.apl.siml';
import DefaultsJson from './defaults.siml';
import DefaultGear from './gear_sets/default.gear.siml';

// The spec's defaults, from defaults.siml.
export const Defaults = PresetUtils.specDefaultsFromSettings(Spec.SpecProtectionPaladin, DefaultsJson);

export const APL_PRESET = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

export const GEAR_DEFAULT = PresetUtils.makePresetGear('Default', DefaultGear, {
	devOnly: true,
	tooltip: 'Level 60 blues from dungeons, quests, reputations and crafting. A placeholder until the Phase 1 gear is known.',
});

export const DefaultTalents = PresetUtils.makePresetTalents('Default', Defaults.talents);
