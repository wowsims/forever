import { useSavedRotation } from '@features/apl/hooks/useSavedRotation';
import { useSavedGear } from '@features/gear/hooks/useSavedGear';
import { useCurrentSettings } from '@features/settings/hooks/useCurrentSettings';
import { useSavedSettings } from '@features/settings/hooks/useSavedSettings';
import { useSavedTalents } from '@features/talents/hooks/useSavedTalents';
import { APLRotation } from '@generated/proto/apl';
import type { BulkBuildBundle } from '@generated/proto/ui';
import { usePlayer, useSimHost, useSpecPresets } from '@sim/context/SimHostContext';
import { useAplRotation } from '@sim/hooks/useAplRotation';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useSimReady } from '@sim/hooks/useSimReady';
import { useMemo } from 'react';

import type { BulkBundleParts, CurrentBundle } from '../model/builds/bundles';
import { bulkGearSetEntries, type BulkGearSetEntry } from '../model/builds/gear_sets';
import { bulkRotationEntries } from '../model/builds/rotations';
import { bulkSettingsEntries } from '../model/builds/settings_entries';
import { bulkTalentEntries } from '../model/builds/talents';
import { useBulkState } from './useBulkState';

const NO_GEAR_SETS: BulkGearSetEntry[] = [];

// The same sources the batch resolves when it starts: the spec's presets, the user's saved gear sets, talents, rotations
// and settings, and the live gear, talents, rotation and settings the Current bundle runs. The gear sets are looked up
// in the database, so there are none until the sim is ready.
export const useBulkBundles = (): {
	ready: boolean;
	current: CurrentBundle;
	parts: BulkBundleParts;
	bundles: readonly BulkBuildBundle[];
	currentChecked: boolean;
} => {
	const host = useSimHost();
	const player = usePlayer();
	const ready = useSimReady();
	const presets = useSpecPresets();
	const { entries: savedGear } = useSavedGear();
	const { entries: savedTalents } = useSavedTalents();
	const { entries: savedRotations } = useSavedRotation();
	const { entries: savedSettings } = useSavedSettings();
	const bundles = useBulkState(slice => slice.bundles);
	const currentChecked = useBulkState(slice => slice.currentBundleChecked);
	const gear = usePlayerStore('gear');
	const bonusStats = usePlayerStore('bonusStats');
	const talentsString = usePlayerStore('talentsString');
	const rotation = useAplRotation(live => APLRotation.clone(live));
	const settings = useCurrentSettings();

	const current = useMemo(
		(): CurrentBundle => ({ gear, bonusStats, talentsString, rotation, settings }),
		[gear, bonusStats, talentsString, rotation, settings],
	);
	const gearSets = useMemo(() => (ready ? bulkGearSetEntries(presets.gear, savedGear, player) : NO_GEAR_SETS), [ready, presets, savedGear, player]);
	const talents = useMemo(() => bulkTalentEntries(presets.talents, savedTalents), [presets, savedTalents]);
	const rotations = useMemo(() => bulkRotationEntries(presets.rotations, savedRotations), [presets, savedRotations]);
	const settingsEntries = useMemo(() => bulkSettingsEntries(presets, savedSettings, host), [presets, savedSettings, host]);
	const parts = useMemo(
		(): BulkBundleParts => ({ gearSets, talents, rotations, settings: settingsEntries }),
		[gearSets, talents, rotations, settingsEntries],
	);

	return { ready, current, parts, bundles, currentChecked };
};
