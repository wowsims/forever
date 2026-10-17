import type { BulkEncounterRef } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { useId, useMemo } from 'react';

import { useBulkEncounters } from '../../hooks/useBulkEncounters';
import { bulkEncounterList, formatEncounterSummary, setBulkEncounterChecked } from '../../model/encounters';

export interface EncounterRow {
	ref: BulkEncounterRef;
	name: string;
	summary: string;
	checked: boolean;
	duplicate: boolean;
}

export interface EncountersViewProps {
	currentSummary: string;
	presets: readonly EncounterRow[];
	saved: readonly EncounterRow[];
	onCheckedChange: (ref: BulkEncounterRef, checked: boolean) => void;
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
}

// The summary shows as the label's hover tooltip; the sr-only copy describes the checkbox for screen readers, which
// the tooltip does not reach.
const EncounterCheckRow = ({ id, label, summary, checked, disabled, duplicate = false, onChange, testId }: EncounterCheckRowProps) => (
	<>
		<BooleanPicker
			id={id}
			label={label}
			labelTooltip={summary}
			layout="inline"
			size="sm"
			value={checked}
			disabled={disabled}
			invalid={duplicate}
			aria-describedby={duplicate ? `${id}-summary ${id}-duplicate` : `${id}-summary`}
			onChange={onChange}
			testId={testId}
		/>
		<span id={`${id}-summary`} className="sr-only">
			{summary}
		</span>
		{duplicate && (
			<span id={`${id}-duplicate`} className="sr-only">
				{t('duplicate')}
			</span>
		)}
	</>
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
						onChange={checked => onCheckedChange(row.ref, checked)}
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
	const rows = useMemo(() => {
		const list = bulkEncounterList(current, entries, checked).map((item): EncounterRow => ({
			ref: item.ref,
			name: item.name,
			summary: formatEncounterSummary(item.encounter),
			checked: item.checked,
			duplicate: item.duplicate,
		}));
		return { presets: list.filter(row => row.ref.preset), saved: list.filter(row => !row.ref.preset) };
	}, [current, entries, checked]);

	return (
		<EncountersView
			currentSummary={formatEncounterSummary(current)}
			presets={rows.presets}
			saved={rows.saved}
			onCheckedChange={(ref, value) =>
				setBulkEncounterChecked(
					player,
					ref,
					value,
					entries.map(entry => entry.ref),
				)
			}
		/>
	);
};
