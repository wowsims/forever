import { DistributionMetrics } from '@generated/proto/api';
import { ItemSlot, ItemType } from '@generated/proto/common';
import { UIItem } from '@generated/proto/ui';
import type { BulkEncounterResults, TopGearResult } from '@sim/bulk/types';
import type { Player } from '@sim/player/player';
import { EquippedItem } from '@sim/proto/equipped_item';
import { Gear } from '@sim/proto/gear';
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

const result = (gear: Gear, avg: number): TopGearResult => ({ gear, dpsMetrics: DistributionMetrics.create({ avg, stdev: 10 }) });

// An encounter ranking the sets with the given helm ids at the given DPS, and the current gear at 1000, best first.
const encounter = (name: string, duration: number | null, ranked: Array<[number, number]>): BulkEncounterResults => {
	const base = result(CURRENT, 1000);
	const results = [...ranked.map(([id, avg]) => result(compareGearSet(id), avg)), base].sort((a, b) => b.dpsMetrics.avg - a.dpsMetrics.avg);
	return {
		name,
		isCurrent: false,
		durationLabel: '',
		duration,
		durationVariation: duration === null ? null : 5,
		results: { chains: results.map(ranking => [ranking]), originalGearResults: base, iterations: 100 },
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

// A host whose batch results are `entries`, over a real store, for the comparison views.
export const makeBulkCompareHost = (entries: BulkEncounterResults[] = BULK_COMPARE_ENTRIES): IndividualSimHost<any> => {
	const player = { sim: { store: createSimStore() }, storeKey: 0 } as unknown as Player<any>;
	seedBulkSettings(player);
	patchBulkState(player, { started: true, results: { entries } });
	return { player, sim: player.sim } as unknown as IndividualSimHost<any>;
};
