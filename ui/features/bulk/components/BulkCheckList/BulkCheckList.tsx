import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { useId } from 'react';

import type { SavedRef } from '../../model/builds/saved_refs';

export interface BulkCheckRow {
	ref: SavedRef;
	name: string;
	summary: string;
	checked: boolean;
	duplicate: boolean;
	/** A row with one shows unchecked and cannot be ticked, whatever `checked` says. */
	disabledReason?: string;
}

export interface BulkCheckListProps {
	/** `presets` heads the presets and `duplicate` describes a row flagged `duplicate`; a list with no presets, or with no row flagged, leaves the string out. */
	strings: { presets?: string; saved: string; empty: string; duplicate?: string };
	/** The first row; null for a list with no current entry. */
	current: { label: string; summary: string; checked: boolean; onCheckedChange: (checked: boolean) => void } | null;
	presets: readonly BulkCheckRow[];
	saved: readonly BulkCheckRow[];
	onCheckedChange: (ref: SavedRef, checked: boolean) => void;
	testIds: { list: string; row: string; current?: string };
}

interface CheckRowProps {
	id: string;
	label: string;
	summary: string;
	checked: boolean;
	duplicate?: boolean;
	duplicateText?: string;
	disabledReason?: string;
	onChange: (checked: boolean) => void;
	testId?: string;
}

// The summary shows as the label's hover tooltip, the disabled reason in its place on a disabled row, right of the label,
// whose box stays in the tab order and shows the reason on focus too. The sr-only copies of the summary, the duplicate note and the reason
// describe the checkbox for screen readers, which the tooltip does not reach.
const CheckRow = ({ id, label, summary, checked, duplicate = false, duplicateText, disabledReason, onChange, testId }: CheckRowProps) => (
	<>
		<BooleanPicker
			id={id}
			className="wrap-anywhere"
			label={label}
			labelTooltip={disabledReason ?? summary}
			labelTooltipPlace={disabledReason ? 'right' : undefined}
			layout="inline"
			size="xs"
			value={disabledReason ? false : checked}
			disabled={!!disabledReason}
			focusableWhenDisabled
			invalid={duplicate}
			aria-describedby={[`${id}-summary`, duplicate && `${id}-duplicate`, disabledReason && `${id}-reason`].filter(Boolean).join(' ')}
			onChange={onChange}
			testId={testId}
		/>
		<span id={`${id}-summary`} className="sr-only">
			{summary}
		</span>
		{duplicate && (
			<span id={`${id}-duplicate`} className="sr-only">
				{duplicateText}
			</span>
		)}
		{disabledReason && (
			<span id={`${id}-reason`} className="sr-only">
				{disabledReason}
			</span>
		)}
	</>
);

export const BulkCheckList = ({ strings, current, presets, saved, onCheckedChange, testIds }: BulkCheckListProps) => {
	const idPrefix = useId();
	// A labelled group per heading, so two rows of one name under two headings read apart.
	const group = (key: string, heading: string, rows: readonly BulkCheckRow[]) =>
		rows.length > 0 && (
			<div key={key} role="group" aria-labelledby={`${idPrefix}-${key}`} className="grid gap-1">
				<h6 id={`${idPrefix}-${key}`} className="mb-0 text-xs text-gray-400">
					{heading}
				</h6>
				{rows.map((row, index) => (
					<CheckRow
						key={row.name}
						id={`${idPrefix}-${key}-${index}`}
						label={row.name}
						summary={row.summary}
						checked={row.checked}
						duplicate={row.duplicate}
						duplicateText={strings.duplicate}
						disabledReason={row.disabledReason}
						onChange={checked => onCheckedChange(row.ref, checked)}
						testId={testIds.row}
					/>
				))}
			</div>
		);

	return (
		<div className="grid gap-2" data-testid={testIds.list}>
			{current && (
				<CheckRow
					id={`${idPrefix}-current`}
					label={current.label}
					summary={current.summary}
					checked={current.checked}
					onChange={current.onCheckedChange}
					testId={testIds.current}
				/>
			)}
			{group('presets', strings.presets ?? '', presets)}
			{group('saved', strings.saved, saved)}
			{!presets.length && !saved.length && <p className="mb-0 text-sm text-gray-400">{strings.empty}</p>}
		</div>
	);
};
