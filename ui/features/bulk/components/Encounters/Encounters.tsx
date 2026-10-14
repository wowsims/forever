import { BulkEncounterRef } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { useId } from 'react';

import { useBulkEncounters } from '../../hooks/useBulkEncounters';
import { bulkEncounterList, formatEncounterSummary, setBulkEncounterChecked } from '../../model/encounters';

export interface EncounterRow {
	name: string;
	summary: string;
	checked: boolean;
	duplicate: boolean;
}

export interface EncountersViewProps {
	currentSummary: string;
	presets: readonly EncounterRow[];
	saved: readonly EncounterRow[];
	onCheckedChange: (preset: boolean, name: string, checked: boolean) => void;
}

const t = (key: string) => i18n.t(`bulk_tab.settings.encounters.${key}`);

interface EncounterCheckRowProps {
	id: string;
	label: string;
	summary: string;
	checked: boolean;
	disabled?: boolean;
	duplicate?: boolean;
	onChange: (checked: boolean) => void;
	testId: string;
	summaryTestId?: string;
}

// The label column shrinks and wraps, so a long name never runs into the summary.
const EncounterCheckRow = ({ id, label, summary, checked, disabled, duplicate = false, onChange, testId, summaryTestId }: EncounterCheckRowProps) => (
	<div className="grid grid-cols-[1fr_auto] items-center gap-2" data-testid={testId}>
		<BooleanPicker
			id={id}
			className="min-w-0"
			label={label}
			layout="inline"
			size="sm"
			value={checked}
			disabled={disabled}
			invalid={duplicate}
			aria-describedby={duplicate ? `${id}-duplicate` : undefined}
			onChange={onChange}
		/>
		{duplicate && (
			<span id={`${id}-duplicate`} className="sr-only">
				{t('duplicate')}
			</span>
		)}
		<span className="text-xs whitespace-nowrap text-gray-400 tabular-nums" data-testid={summaryTestId}>
			{summary}
		</span>
	</div>
);

export const EncountersView = ({ currentSummary, presets, saved, onCheckedChange }: EncountersViewProps) => {
	const idPrefix = useId();
	const group = (preset: boolean, heading: string, rows: readonly EncounterRow[]) =>
		rows.length > 0 && (
			<div className="grid gap-1">
				<h6 className="mb-0 text-xs text-gray-400">{heading}</h6>
				{rows.map((row, index) => (
					<EncounterCheckRow
						key={row.name}
						id={`${idPrefix}-${preset ? 'preset' : 'saved'}-${index}`}
						label={row.name}
						summary={row.summary}
						checked={row.checked}
						duplicate={row.duplicate}
						onChange={checked => onCheckedChange(preset, row.name, checked)}
						testId="bulk-encounter"
					/>
				))}
			</div>
		);

	return (
		<div className="grid gap-2" data-testid="bulk-encounters">
			<p className="mb-0 text-xs">{t('description')}</p>
			<EncounterCheckRow
				id={`${idPrefix}-current`}
				label={t('current')}
				summary={currentSummary}
				checked
				disabled
				onChange={() => {}}
				testId="bulk-encounter-current"
				summaryTestId="bulk-encounter-current-summary"
			/>
			{group(true, t('presets'), presets)}
			{group(false, t('saved'), saved)}
			{!presets.length && !saved.length && <p className="mb-0 text-sm text-gray-400">{t('empty')}</p>}
		</div>
	);
};

export const Encounters = () => {
	const player = usePlayer();
	const { current, entries, checked } = useBulkEncounters();
	const list = bulkEncounterList(current, entries, checked);
	const rows = (preset: boolean): EncounterRow[] =>
		list
			.filter(item => item.ref.preset === preset)
			.map(item => ({ name: item.name, summary: formatEncounterSummary(item.encounter), checked: item.checked, duplicate: item.duplicate }));

	return (
		<EncountersView
			currentSummary={formatEncounterSummary(current)}
			presets={rows(true)}
			saved={rows(false)}
			onCheckedChange={(preset, name, value) =>
				setBulkEncounterChecked(
					player,
					BulkEncounterRef.create({ name, preset }),
					value,
					entries.map(entry => entry.ref),
				)
			}
		/>
	);
};
