import type { ItemSlot } from '@generated/proto/common';
import { translateSlotName } from '@i18n/localization';
import type { BulkEncounterResults } from '@sim/bulk/types';
import { ActionId } from '@sim/proto/action_id';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { type CsvValue, toCsv } from '@sim/utils/csv';
import { stDevToConf95 } from '@sim/utils/math';

import { encounterLabel } from './encounters';
import { formatBulkItemLabel } from './format';

export const BULK_RESULTS_CSV_FILE_NAME = 'wowsims-batch-results.csv';

const COLUMNS = ['encounter', 'duration', 'variation', 'rank', 'current_gear', 'dps', 'dps_margin_95', 'diff', 'diff_pct'] as const;

const slotHeaders = (slot: ItemSlot): string[] => {
	const name = translateSlotName(slot);
	return [name, `${name} wowhead`, `${name} item id`, `${name} enchant id`];
};

// Per slot: "Item Suffix (Enchant)" as the result rows name it, the item's wowhead link as the gear tab builds it (suffix
// as its bonus), the item id and the enchant id the gear spec stores. An empty slot leaves all four cells empty.
const slotCells = (equipped: EquippedItem | null): CsvValue[] => {
	if (!equipped) return ['', '', '', ''];
	return [formatBulkItemLabel(equipped), ActionId.makeItemUrl(equipped.id, equipped.randomSuffix?.id), equipped.id, equipped.enchant?.effectId ?? ''];
};

// One row per ranked set of every finished encounter, in the order the Results pane lists them, with the full gear set.
export const bulkResultsCsv = (entries: readonly BulkEncounterResults[]): string => {
	const slots: ItemSlot[] = entries[0]?.results.originalGearResults.gear.getItemSlots() ?? [];
	const rows = entries.flatMap(entry => {
		const { chains, originalGearResults: base, iterations } = entry.results;
		const encounter = encounterLabel(entry);
		return chains.flat().map((result, index): CsvValue[] => {
			const avg = result.dpsMetrics.avg;
			const diff = avg - base.dpsMetrics.avg;
			// The percentage is taken against the smaller of the two, as the result rows show it.
			const diffPct = (diff / (Math.min(avg, base.dpsMetrics.avg) || 1)) * 100;
			return [
				encounter,
				entry.duration ?? '',
				entry.durationVariation ?? '',
				index + 1,
				result.gear.equals(base.gear) ? 'yes' : '',
				avg,
				stDevToConf95(result.dpsMetrics.stdev, iterations),
				diff,
				diffPct,
				...slots.flatMap(slot => slotCells(result.gear.getEquippedItem(slot))),
			];
		});
	});
	return toCsv([[...COLUMNS, ...slots.flatMap(slotHeaders)], ...rows]);
};
