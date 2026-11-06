import { Importer } from '@features/import-export/components/Importer/Importer';
import { type ImporterPreview, ImportProblems } from '@features/import-export/importers/types';
import i18n from '@i18n/config';
import type { UseSavedDataResult } from '@ui-kit/hooks/useSavedData';
import { useMemo, useRef } from 'react';

import { trackEvent } from '../../../../tracking/utils';
import { type ImportPlan, type ImportRow, tickedRows, withTicksKept } from '../../model/import_plan';
import { readSavedDataImport } from '../../model/read_import';
import { MAX_TEXT_LENGTH, type SavedDataType } from '../../model/transfer_text';
import { SavedDataTickList } from '../SavedDataTickList';

export interface SavedDataImporterProps<T> {
	type: SavedDataType;
	label: string;
	storage: Pick<UseSavedDataResult<T>, 'entries' | 'storedNames' | 'saveMany' | 'codec'>;
	presets: ReadonlyArray<{ name: string }>;
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

const rowNote = ({ name, saveAs, savedAs, ticked }: ImportRow<unknown>): string | undefined => {
	if (savedAs === undefined && saveAs === name) return undefined;
	if (savedAs === undefined || ticked) return i18n.t('saved_data_transfer.saves_as', { name: saveAs });
	return savedAs === name ? i18n.t('saved_data_transfer.already_saved') : i18n.t('saved_data_transfer.already_saved_as', { name: savedAs });
};

const noPicks = () => ({ text: '', ticks: new Map<string, boolean>() });

export const SavedDataImporter = <T,>({
	type,
	label,
	storage: { entries, storedNames, saveMany, codec },
	presets,
	open,
	onOpenChange,
}: SavedDataImporterProps<T>) => {
	// What the player ticked in the text on show, since a write to the slot from another tab plans that text again.
	const picks = useRef(noPicks());

	const preview = useMemo(
		(): ImporterPreview<ImportPlan<T>> => ({
			parse: async (host, text, signal) => {
				if (picks.current.text !== text) picks.current = { text, ticks: new Map() };
				const { ticks } = picks.current;
				const plan = await readSavedDataImport(
					text,
					host,
					type,
					{ codec, saved: entries, storedNames, presetNames: presets.map(({ name }) => name) },
					signal,
				);
				return withTicksKept(plan, ticks);
			},
			render: (plan, change) => (
				<>
					<SavedDataTickList
						rows={plan.rows.map(row => ({ name: row.name, note: rowNote(row), checked: row.ticked }))}
						onChange={(index, ticked) => {
							const { ticks } = picks.current;
							ticks.set(plan.rows[index].name, ticked);
							change(withTicksKept(plan, ticks));
						}}
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
			commit: async plan => {
				const saved = saveMany(tickedRows(plan).map(({ saveAs, data }) => ({ name: saveAs, data })));
				if (!saved) throw new ImportProblems(i18n.t('saved_data_transfer.refused.storage_full'), []);
				trackEvent({ action: 'settings', category: 'import', label });
			},
		}),
		[type, label, codec, entries, storedNames, saveMany, presets],
	);

	return (
		<Importer
			open={open}
			onOpenChange={next => {
				if (!next) picks.current = noPicks();
				onOpenChange(next);
			}}
			title={i18n.t('saved_data_transfer.import_title', { type: i18n.t(`saved_data_transfer.types.${type}`) })}
			allowFileUpload
			holdsLink
			maxFileBytes={MAX_TEXT_LENGTH}
			preview={preview}>
			<p>{i18n.t('saved_data_transfer.import_instructions')}</p>
		</Importer>
	);
};
