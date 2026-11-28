import { EncounterPicker, SavedEncounter } from '@features/encounter';
import type { BulkEncounterRef } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { usePlayer, useSpecConfig } from '@sim/context/SimHostContext';
import { Button } from '@ui-kit/Button';
import { ButtonGroup } from '@ui-kit/ButtonGroup';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { Dialog } from '@ui-kit/Dialog';
import { Field } from '@ui-kit/Field';
import { Icon } from '@ui-kit/Icon';
import { useState } from 'react';

import { useBatchEncounter } from '../../hooks/useBatchEncounter';
import { useBulkEncounters } from '../../hooks/useBulkEncounters';
import { useBulkState } from '../../hooks/useBulkState';
import { setBatchEncounter } from '../../model/batch_encounter';
import { CURRENT_ENCOUNTER_REF, formatEncounterSummary } from '../../model/builds/encounters';
import { sameRef } from '../../model/builds/saved_refs';

export interface BatchEncounterViewProps {
	/** "Current encounter" (CURRENT_ENCOUNTER_REF) first, which the container prepends, then the spec's preset
	 * encounters and the saved ones, in Builds' Encounters column order. Rendered as given. Never empty. */
	options: ReadonlyArray<{ ref: BulkEncounterRef; name: string }>;
	/** CURRENT_ENCOUNTER_REF while the batch uses the Encounter tab. */
	picked: BulkEncounterRef;
	/** The picked encounter's formatEncounterSummary; the Encounter tab's while Current is picked. */
	summary: string;
	onChange(ref: BulkEncounterRef): void;
	onEdit(): void;
}

const t = (key: string) => i18n.t(`bulk_tab.settings.batch_encounter.${key}`);

// The summary is hint text under the select, which a field description would draw as a filled box like an input.
export const BatchEncounterView = ({ options, picked, summary, onChange, onEdit }: BatchEncounterViewProps) => (
	<div className="grid gap-1" data-testid="bulk-batch-encounter">
		<Field.Root id="bulk-batch-encounter-pick">
			<Field.Label tooltip={t('tooltip')}>{t('label')}</Field.Label>
			<ButtonGroup className="flex w-full">
				<ChoiceField
					id="bulk-batch-encounter-pick"
					className="min-w-0 flex-1"
					options={options.map(({ ref, name }) => ({ value: ref, label: name }))}
					value={picked}
					equals={sameRef}
					renderValue={({ label }) => <span className="truncate">{label}</span>}
					aria-describedby="bulk-batch-encounter-summary"
					onChange={onChange}
				/>
				<Button
					variant="field"
					className="aspect-square h-(--control-sm-height) shrink-0 justify-center px-0"
					aria-label={t('edit')}
					data-testid="bulk-batch-encounter-edit"
					onClick={onEdit}>
					<Icon name="pencil" />
				</Button>
			</ButtonGroup>
		</Field.Root>
		<p id="bulk-batch-encounter-summary" className="mb-0 text-xs text-gray-400">
			{summary}
		</p>
	</div>
);

export interface BatchEncounterDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

// The Settings tab's encounter editor and saved encounters. Both work on the Current encounter, which the single sim
// shares, so the dialog says so; saving one adds it to the picker.
export const BatchEncounterDialog = ({ open, onOpenChange }: BatchEncounterDialogProps) => {
	const config = useSpecConfig();
	return (
		<Dialog open={open} onOpenChange={onOpenChange} size="lg" title={t('edit_title')} scrollContents testId="bulk-batch-encounter-dialog">
			<p className="text-sm text-gray-400">{t('edit_note')}</p>
			<div className="grid gap-6">
				<EncounterPicker showExecuteProportion={config.encounterPicker.showExecuteProportion} />
				<SavedEncounter />
			</div>
		</Dialog>
	);
};

export const BatchEncounter = () => {
	const player = usePlayer();
	const [editing, setEditing] = useState(false);
	const { current, entries } = useBulkEncounters();
	const ref = useBulkState(slice => slice.batchEncounter);
	const encounter = useBatchEncounter();

	const options = [{ ref: CURRENT_ENCOUNTER_REF, name: t('current') }, ...entries.map(({ ref, name }) => ({ ref, name }))];

	return (
		<>
			<BatchEncounterView
				options={options}
				picked={encounter ? ref : CURRENT_ENCOUNTER_REF}
				summary={formatEncounterSummary(encounter ?? current)}
				onChange={newRef => setBatchEncounter(player, newRef)}
				onEdit={() => {
					setBatchEncounter(player, CURRENT_ENCOUNTER_REF);
					setEditing(true);
				}}
			/>
			{editing && <BatchEncounterDialog open onOpenChange={setEditing} />}
		</>
	);
};
