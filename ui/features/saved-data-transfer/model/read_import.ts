import { ImportProblems } from '@features/import-export/importers/types';
import { Spec } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { IndividualSimHost } from '@sim/sim_host';

import { type ImportPlan, type ImportPlanSources, planImport } from './import_plan';
import { SAVED_DATA_TYPE_RULES, specFromName, specFullName } from './saved_data_types';
import { decodeSavedData, type SavedDataExport, SavedDataTextError, type SavedDataType } from './transfer_text';

const MAX_IMPORT_ENTRIES = 500;

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`saved_data_transfer.${key}`, options);

const refusal = (key: string, options?: Record<string, unknown>) => new ImportProblems(t(`refused.${key}`, options), []);

const decodeOrRefuse = async (text: string): Promise<SavedDataExport> => {
	try {
		return await decodeSavedData(text);
	} catch (error) {
		throw error instanceof SavedDataTextError ? refusal(error.problem) : error;
	}
};

export const readSavedDataImport = async <T>(
	text: string,
	host: IndividualSimHost<any>,
	type: SavedDataType,
	sources: Omit<ImportPlanSources<T>, 'check'>,
	signal: AbortSignal,
): Promise<ImportPlan<T>> => {
	const data = await decodeOrRefuse(text);
	if (data.type !== type) throw refusal('wrong_type', { found: t(`types.${data.type}`), wanted: t(`types.${type}`) });

	const { specLocked, check } = SAVED_DATA_TYPE_RULES[type];
	if (specLocked && data.spec !== Spec[host.player.getSpec()]) {
		const spec = specFromName(data.spec);
		throw spec === undefined ? refusal('unknown_spec') : refusal('other_spec', { spec: specFullName(spec) });
	}
	if (Object.keys(data.entries).length > MAX_IMPORT_ENTRIES) throw refusal('too_large');

	signal.throwIfAborted();
	const plan = planImport<T>(data.entries, {
		...sources,
		check: check && (entry => check(entry, host)),
	});
	if (!plan.rows.length) throw refusal(plan.unreadable ? 'none_readable' : 'empty');
	return plan;
};
