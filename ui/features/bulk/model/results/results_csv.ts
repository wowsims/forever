import type { ItemSlot } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateSlotName } from '@i18n/localization';
import type { BulkEncounterResults } from '@sim/bulk/types';
import { ActionId } from '@sim/proto/action_id';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { type CsvValue, toCsv } from '@sim/utils/csv';
import { stDevToConf95 } from '@sim/utils/math';

import { encounterLabel } from '../builds/encounters';
import { gearSetupLabel } from '../builds/gear_setups';
import { rotationLabel } from '../builds/rotations';
import { settingsLabel } from '../builds/settings_entries';
import { talentsLabel } from '../builds/talents';
import { changedEnchantLabels } from './compare';
import { formatBulkItemLabel } from './format';

export const BULK_RESULTS_CSV_FILE_NAME = 'wowsims-batch-results.csv';

const COLUMNS = [
	'encounter',
	'duration',
	'variation',
	'rank',
	'current_gear',
	'dps',
	'dps_margin_95',
	'diff',
	'diff_pct',
	'talents',
	'talents_string',
	'gear_set',
	'changed_enchants',
	'rotation',
	'settings',
] as const;

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
		const { chains, originalGearResults: base, currentGear, iterations } = entry.results;
		const encounter = encounterLabel(entry);
		return chains.flat().map((result, index): CsvValue[] => {
			const avg = result.dpsMetrics.avg;
			const diff = avg - base.dpsMetrics.avg;
			// In percent of the reference's DPS, as the result rows show it.
			const diffPct = (diff / (base.dpsMetrics.avg || 1)) * 100;
			return [
				encounter,
				entry.duration ?? '',
				entry.durationVariation ?? '',
				index + 1, // The current gear's row takes a rank like any other.
				result.gear.equals(currentGear) ? 'yes' : '',
				avg,
				stDevToConf95(result.dpsMetrics.stdev, iterations),
				diff,
				diffPct,
				talentsLabel(result.talents),
				result.talents.talentsString,
				// The Builds reference row wears the current gear, which no gear setup stands for.
				result.gearSetup ? gearSetupLabel(result.gearSetup) : result.settings ? i18n.t('bulk_tab.results.current_gear') : '',
				changedEnchantLabels(result).join('; '),
				rotationLabel(result.rotation),
				result.settings ? settingsLabel(result.settings) : '',
				...slots.flatMap(slot => slotCells(result.gear.getEquippedItem(slot))),
			];
		});
	});
	return toCsv([[...COLUMNS, ...slots.flatMap(slotHeaders)], ...rows]);
};
