import { DistributionMetrics } from '@generated/proto/api';
import { Class, ItemSlot, ItemType, Spec } from '@generated/proto/common';
import { UIItem } from '@generated/proto/ui';
import type { BulkEncounterResults, BulkRankedResult, BulkTalentSetup } from '@sim/bulk/types';
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

export const CURRENT_TALENT_SETUP: BulkTalentSetup = { name: '', talentsString: '0000000000', isCurrent: true };

const ARMS_TALENT_SETUP: BulkTalentSetup = { name: 'Arms', talentsString: '3030500130', isCurrent: false };

const result = (gear: Gear, avg: number, talents = CURRENT_TALENT_SETUP): BulkRankedResult => ({
	gear,
	dpsMetrics: DistributionMetrics.create({ avg, stdev: 10 }),
	talents,
});

// An encounter ranking the sets with the given helm ids at the given DPS, with the current talents unless a set names
// others, and the current gear with the current talents at 1000, best first.
const encounter = (name: string, duration: number | null, ranked: Array<[number, number, BulkTalentSetup?]>, talentSetups = 1): BulkEncounterResults => {
	const base = result(CURRENT, 1000);
	const results = [...ranked.map(([id, avg, talents]) => result(compareGearSet(id), avg, talents)), base].sort((a, b) => b.dpsMetrics.avg - a.dpsMetrics.avg);
	return {
		name,
		isCurrent: false,
		duration,
		durationVariation: duration === null ? null : 5,
		results: { chains: results.map(ranking => [ranking]), originalGearResults: base, iterations: 100, talentSetups },
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

// Two encounters that each ran the current talents and Arms; display order Council, Patchwerk. Helm 2 with Arms ranks in
// both, the current gear (helm 1) with Arms only in Council, helm 3 with the current talents in both.
export const BULK_COMPARE_TALENT_ENTRIES: BulkEncounterResults[] = [
	encounter(
		'Patchwerk',
		300,
		[
			[3, 1100],
			[2, 1080, ARMS_TALENT_SETUP],
		],
		2,
	),
	encounter(
		'Council',
		90,
		[
			[2, 1060, ARMS_TALENT_SETUP],
			[1, 1030, ARMS_TALENT_SETUP],
			[3, 1020],
		],
		2,
	),
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
	patchBulkState(player, { started: true, results: entries });
	return { player, sim: player.sim } as unknown as IndividualSimHost<any>;
};
