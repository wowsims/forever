import type { BulkEncounterRef } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { ChoiceField } from '@ui-kit/ChoiceField';

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
}

const t = (key: string) => i18n.t(`bulk_tab.settings.batch_encounter.${key}`);

export const BatchEncounterView = ({ options, picked, summary, onChange }: BatchEncounterViewProps) => (
	<div className="grid gap-2" data-testid="bulk-batch-encounter">
		<ChoiceField
			id="bulk-batch-encounter-pick"
			label={t('label')}
			labelTooltip={t('tooltip')}
			options={options.map(({ ref, name }) => ({ value: ref, label: name }))}
			value={picked}
			equals={sameRef}
			description={summary}
			onChange={onChange}
		/>
	</div>
);

export const BatchEncounter = () => {
	const player = usePlayer();
	const { current, entries } = useBulkEncounters();
	const ref = useBulkState(slice => slice.batchEncounter);
	const encounter = useBatchEncounter();

	const options = [{ ref: CURRENT_ENCOUNTER_REF, name: t('current') }, ...entries.map(({ ref, name }) => ({ ref, name }))];

	return (
		<BatchEncounterView
			options={options}
			picked={encounter ? ref : CURRENT_ENCOUNTER_REF}
			summary={formatEncounterSummary(encounter ?? current)}
			onChange={newRef => setBatchEncounter(player, newRef)}
		/>
	);
};
