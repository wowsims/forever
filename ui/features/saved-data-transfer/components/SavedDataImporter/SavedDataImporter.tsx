import { Importer } from '@features/import-export/components/Importer/Importer';
import { type ImporterPreview, ImportProblems } from '@features/import-export/importers/types';
import i18n from '@i18n/config';
import type { IndividualSimHost } from '@sim/sim_host';
import type { UseSavedDataResult } from '@ui-kit/hooks/useSavedData';
import { useCallback, useMemo } from 'react';

import { trackEvent } from '../../../../tracking/utils';
import { type ImportPlan, type ImportRow, tickedRows, withRowTicked } from '../../model/import_plan';
import { readSavedDataImport } from '../../model/read_import';
import { MAX_TEXT_LENGTH, type SavedDataType } from '../../model/transfer_text';
import { SavedDataTickList } from '../SavedDataTickList';

export interface SavedDataImporterProps<T> {
	type: SavedDataType;
	label: string;
	storage: Pick<UseSavedDataResult<T>, 'entries' | 'storedNames' | 'saveMany' | 'codec'>;
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

const rowNote = ({ status, name, saveAs, savedAs, ticked }: ImportRow<unknown>): string | undefined => {
	if (status === 'new') return undefined;
	if (status === 'renamed' || ticked) return i18n.t('saved_data_transfer.saves_as', { name: saveAs });
	return savedAs === name ? i18n.t('saved_data_transfer.already_saved') : i18n.t('saved_data_transfer.already_saved_as', { name: savedAs });
};

export const SavedDataImporter = <T,>({ type, label, storage: { entries, storedNames, saveMany, codec }, open, onOpenChange }: SavedDataImporterProps<T>) => {
	const preview = useMemo(
		(): ImporterPreview<ImportPlan<T>> => ({
			parse: (host: IndividualSimHost<any>, text: string) => readSavedDataImport(text, host, type, { codec, saved: entries, storedNames }),
			render: (plan, change) => (
				<>
					<SavedDataTickList
						rows={plan.rows.map(row => ({ name: row.name, note: rowNote(row), checked: row.ticked }))}
						onChange={(index, ticked) => change(withRowTicked(plan, index, ticked))}
					/>
					{plan.unreadable > 0 && (
						<p className="mb-0 text-sm" data-testid="saved-data-import-unreadable">
							{i18n.t('saved_data_transfer.unreadable', { count: plan.unreadable })}
						</p>
					)}
				</>
			),
			importLabel: plan => i18n.t('saved_data_transfer.import_button', { count: tickedRows(plan).length }),
			canImport: plan => tickedRows(plan).length > 0,
		}),
		[type, codec, entries, storedNames],
	);

	const onImport = useCallback(
		async (_host: IndividualSimHost<any>, _text: string, _fileName?: string, _signal?: AbortSignal, plan?: ImportPlan<T>) => {
			const saved = saveMany(tickedRows(plan!).map(({ saveAs, data }) => ({ name: saveAs, data })));
			if (!saved) throw new ImportProblems(i18n.t('saved_data_transfer.refused.storage_full'), []);
			trackEvent({ action: 'settings', category: 'import', label });
		},
		[label, saveMany],
	);

	return (
		<Importer
			open={open}
			onOpenChange={onOpenChange}
			title={i18n.t('saved_data_transfer.import_title', { type: i18n.t(`saved_data_transfer.types.${type}`) })}
			allowFileUpload
			holdsLink
			maxFileBytes={MAX_TEXT_LENGTH}
			preview={preview}
			onImport={onImport}>
			<p>{i18n.t('saved_data_transfer.import_instructions')}</p>
		</Importer>
	);
};
