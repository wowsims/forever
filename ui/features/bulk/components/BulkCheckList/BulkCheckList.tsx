import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { useId } from 'react';

import type { SavedRef } from '../../model/saved_refs';

export interface BulkCheckRow {
	ref: SavedRef;
	name: string;
	summary: string;
	checked: boolean;
	duplicate: boolean;
}

export interface BulkCheckListProps {
	strings: { description: string; current: string; presets: string; saved: string; empty: string; duplicate: string };
	currentSummary: string;
	presets: readonly BulkCheckRow[];
	saved: readonly BulkCheckRow[];
	onCheckedChange: (ref: SavedRef, checked: boolean) => void;
	testIds: { list: string; row: string; current: string };
}

interface CheckRowProps {
	id: string;
	label: string;
	summary: string;
	checked: boolean;
	disabled?: boolean;
	duplicate?: boolean;
	duplicateText: string;
	onChange: (checked: boolean) => void;
	testId: string;
}

// The summary shows as the label's hover tooltip; the sr-only copy describes the checkbox for screen readers, which
// the tooltip does not reach.
const CheckRow = ({ id, label, summary, checked, disabled, duplicate = false, duplicateText, onChange, testId }: CheckRowProps) => (
	<>
		<BooleanPicker
			id={id}
			label={label}
			labelTooltip={summary}
			layout="inline"
			density="compact"
			value={checked}
			disabled={disabled}
			invalid={duplicate}
			aria-describedby={duplicate ? `${id}-summary ${id}-duplicate` : `${id}-summary`}
			onChange={onChange}
			testId={testId}
		/>
		<span id={`${id}-summary`} className="sr-only">
			{summary}.
		</span>
		{duplicate && (
			<span id={`${id}-duplicate`} className="sr-only">
				{duplicateText}
			</span>
		)}
	</>
);

export const BulkCheckList = ({ strings, currentSummary, presets, saved, onCheckedChange, testIds }: BulkCheckListProps) => {
	const idPrefix = useId();
	const group = (preset: boolean, heading: string, rows: readonly BulkCheckRow[]) =>
		rows.length > 0 && (
			<div className="grid gap-1">
				<h6 className="mb-0 text-xs text-gray-400">{heading}</h6>
				{rows.map((row, index) => (
					<CheckRow
						key={row.name}
						id={`${idPrefix}-${preset ? 'preset' : 'saved'}-${index}`}
						label={row.name}
						summary={row.summary}
						checked={row.checked}
						duplicate={row.duplicate}
						duplicateText={strings.duplicate}
						onChange={checked => onCheckedChange(row.ref, checked)}
						testId={testIds.row}
					/>
				))}
			</div>
		);

	return (
		<div className="grid gap-2" data-testid={testIds.list}>
			<p className="mb-0 text-xs">{strings.description}</p>
			<CheckRow
				id={`${idPrefix}-current`}
				label={strings.current}
				summary={currentSummary}
				checked
				disabled
				duplicateText={strings.duplicate}
				onChange={() => {}}
				testId={testIds.current}
			/>
			{group(true, strings.presets, presets)}
			{group(false, strings.saved, saved)}
			{!presets.length && !saved.length && <p className="mb-0 text-sm text-gray-400">{strings.empty}</p>}
		</div>
	);
};
