import { DistributionMetrics } from '@generated/proto/api';
import { APLRotation } from '@generated/proto/apl';
import { Class, ItemSlot, ItemType, Spec } from '@generated/proto/common';
import { UIItem } from '@generated/proto/ui';
import type { BulkEncounterResults, BulkPlayerSettingsSetup, BulkRankedResult, BulkRotationSetup, BulkTalentSetup } from '@sim/bulk/types';
import type { Player } from '@sim/player/player';
import { EquippedItem } from '@sim/proto/equipped_item';
import { Gear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import { patchBulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { createSimStore } from '@sim/state/sim_store';

// A set wearing the helm `id`, named `Helm <id>`; two sets built from one id are the same set to the gear identity key.
export const compareGearSet = (id: number) =>
	new Gear({
		[ItemSlot.ItemSlotHead]: new EquippedItem({
			item: UIItem.create({ id, name: `Helm ${id}`, type: ItemType.ItemTypeHead, scalingOptions: { 0: { ilvl: 60 } } }),
		}),
	});

const CURRENT = compareGearSet(1);

export const CURRENT_TALENT_SETUP: BulkTalentSetup = { name: '', talentsString: '0000000000', isCurrent: true };

export const CURRENT_ROTATION_SETUP: BulkRotationSetup = { name: '', isCurrent: true, rotation: APLRotation.create() };

export const CURRENT_SETTINGS_SETUP: BulkPlayerSettingsSetup = { name: '', isCurrent: true, settings: null };

const ARMS_TALENT_SETUP: BulkTalentSetup = { name: 'Arms', talentsString: '3030500130', isCurrent: false };

const result = (gear: Gear, avg: number, talents = CURRENT_TALENT_SETUP): BulkRankedResult => ({
	gear,
	dpsMetrics: DistributionMetrics.create({ avg, stdev: 10 }),
	talents,
	rotation: CURRENT_ROTATION_SETUP,
	gearSetup: null,
	settings: null,
});

// An encounter ranking the sets with the given helm ids at the given DPS, and the current gear at 1000, best first, all
// with the current talents.
const encounter = (name: string, duration: number | null, ranked: Array<[number, number]>): BulkEncounterResults => {
	const base = result(CURRENT, 1000);
	const results = [...ranked.map(([id, avg]) => result(compareGearSet(id), avg)), base].sort((a, b) => b.dpsMetrics.avg - a.dpsMetrics.avg);
	return {
		name,
		isCurrent: false,
		duration,
		durationVariation: duration === null ? null : 5,
		results: { chains: results.map(ranking => [ranking]), originalGearResults: base, referenceIsCurrent: true, currentGear: CURRENT, iterations: 100 },
	};
};

// Run order Patchwerk, Council, Ragnaros; display order Council, Patchwerk, Ragnaros (health-based last). Helm 2 ranks in
// every encounter, helm 3 in Council and Ragnaros, the rest in one each.
export const BULK_COMPARE_ENTRIES: BulkEncounterResults[] = [
	encounter('Patchwerk', 300, [
		[2, 1100],
		[4, 1090],
		[5, 1080],
	]),
	encounter('Council', 90, [
		[3, 1030],
		[2, 1020],
		[6, 990],
	]),
	encounter('Ragnaros', null, [
		[7, 12500],
		[3, 1150],
		[2, 1120],
	]),
];

// A Builds set of the gear set `gearSet` wearing the helm `id`, under the current rotation and settings.
const buildResult = (gearSet: string, id: number, avg: number, talents = CURRENT_TALENT_SETUP): BulkRankedResult => ({
	...result(compareGearSet(id), avg, talents),
	gearSetup: { gearSet, gear: compareGearSet(id), bonusStats: new Stats() },
	settings: CURRENT_SETTINGS_SETUP,
});

// An encounter as a Builds batch with the Current bundle unticked stores it: no current gear's row, the sets best first,
// the best the reference.
const bestEncounter = (name: string, duration: number | null, ranked: BulkRankedResult[]): BulkEncounterResults => {
	const results = [...ranked].sort((a, b) => b.dpsMetrics.avg - a.dpsMetrics.avg);
	return {
		name,
		isCurrent: false,
		duration,
		durationVariation: duration === null ? null : 5,
		results: {
			chains: results.map(ranking => [ranking]),
			originalGearResults: results[0],
			referenceIsCurrent: false,
			currentGear: CURRENT,
			iterations: 100,
		},
	};
};

// Three bundles over three encounters with the Current bundle unticked; display order Council, Patchwerk, Ragnaros. Fury
// is best in Patchwerk and Ragnaros, Arms in Council, and Raid ranks only in the timed two.
export const BULK_COMPARE_BEST_ENTRIES: BulkEncounterResults[] = [
	bestEncounter('Patchwerk', 300, [
		buildResult('Pre-BiS (Fury)', 2, 1100),
		buildResult('Raid', 4, 1080),
		buildResult('Pre-BiS (Arms)', 3, 1040, ARMS_TALENT_SETUP),
	]),
	bestEncounter('Council', 90, [
		buildResult('Pre-BiS (Arms)', 3, 1030, ARMS_TALENT_SETUP),
		buildResult('Pre-BiS (Fury)', 2, 1020),
		buildResult('Raid', 4, 990),
	]),
	bestEncounter('Ragnaros', null, [buildResult('Pre-BiS (Fury)', 2, 1150), buildResult('Pre-BiS (Arms)', 3, 1120, ARMS_TALENT_SETUP)]),
];

// A host whose batch results are `entries`, over a real store, for the comparison views. Its player cannot dual wield.
export const makeBulkCompareHost = (entries: BulkEncounterResults[] = BULK_COMPARE_ENTRIES): IndividualSimHost<any> => {
	const player = {
		sim: { store: createSimStore() },
		storeKey: 0,
		getSpec: () => Spec.SpecUnknown,
		getClass: () => Class.ClassUnknown,
	} as unknown as Player<any>;
	seedBulkSettings(player);
	patchBulkState(player, { gearStarted: true, gearResults: entries });
	return { player, sim: player.sim } as unknown as IndividualSimHost<any>;
};
