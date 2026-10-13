import { BulkEncounterRef } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { useId } from 'react';

import { useBulkEncounters } from '../../hooks/useBulkEncounters';
import { formatEncounterSummary, isBulkEncounterChecked, isDuplicateEncounter, setBulkEncounterChecked } from '../../model/encounters';

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

export const EncountersView = ({ currentSummary, presets, saved, onCheckedChange }: EncountersViewProps) => {
	const idPrefix = useId();
	const group = (preset: boolean, heading: string, rows: readonly EncounterRow[]) =>
		rows.length > 0 && (
			<div className="grid gap-1">
				<h6 className="mb-0 text-xs text-gray-400">{heading}</h6>
				{rows.map((row, index) => {
					const id = `${idPrefix}-${preset ? 'preset' : 'saved'}-${index}`;
					return (
						<div
							key={row.name}
							className="flex items-center gap-2 border border-transparent px-1 data-duplicate:border-danger"
							data-duplicate={row.duplicate || undefined}
							data-testid="bulk-encounter">
							<BooleanPicker
								id={id}
								className="min-w-0 flex-1"
								label={row.name}
								layout="inline"
								size="sm"
								value={row.checked}
								aria-describedby={row.duplicate ? `${id}-duplicate` : undefined}
								onChange={checked => onCheckedChange(preset, row.name, checked)}
							/>
							{row.duplicate && (
								<span id={`${id}-duplicate`} className="sr-only">
									{t('duplicate')}
								</span>
							)}
							<span className="text-xs text-gray-400 tabular-nums">{row.summary}</span>
						</div>
					);
				})}
			</div>
		);

	return (
		<div className="grid gap-2" data-testid="bulk-encounters">
			<p className="mb-0 text-sm">{t('description')}</p>
			<div className="flex items-center gap-2 border border-transparent px-1" data-testid="bulk-encounter-current">
				<BooleanPicker
					id={`${idPrefix}-current`}
					className="min-w-0 flex-1 whitespace-nowrap"
					label={t('current')}
					layout="inline"
					size="sm"
					value
					disabled
					onChange={() => {}}
				/>
				<span className="text-xs text-gray-400 tabular-nums" data-testid="bulk-encounter-current-summary">
					{currentSummary}
				</span>
			</div>
			{group(true, t('presets'), presets)}
			{group(false, t('saved'), saved)}
			{!presets.length && !saved.length && <p className="mb-0 text-sm text-gray-400">{t('empty')}</p>}
		</div>
	);
};

export const Encounters = () => {
	const player = usePlayer();
	const { current, entries, checked, picked } = useBulkEncounters();
	const pickedEncounters = picked.map(entry => entry.encounter);
	const rows = (preset: boolean): EncounterRow[] =>
		entries
			.filter(entry => entry.ref.preset === preset)
			.map(entry => {
				const at = picked.indexOf(entry);
				return {
					name: entry.name,
					summary: formatEncounterSummary(entry.encounter),
					checked: isBulkEncounterChecked(checked, entry.ref),
					duplicate: at >= 0 && isDuplicateEncounter(current, pickedEncounters, at),
				};
			});

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
