import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import type { ChoiceFieldConfig } from '@features/inputs/configs';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import type { Encounter } from '@sim/raid/encounter';
import { Dialog } from '@ui-kit/Dialog';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { useMemo } from 'react';

import { trackEvent } from '../../../../tracking/utils';
import { durationConfigs, executeConfigs } from '../EncounterPicker/utils/configs';
import { TargetsPicker } from '../TargetsPicker';

export interface AdvancedEncounterModalProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

export const AdvancedEncounterModal = ({ open, onOpenChange }: AdvancedEncounterModalProps) => {
	const host = useSimHost();
	const encounter = host.sim.encounter;

	const presetConfig = useMemo((): ChoiceFieldConfig<Encounter> => {
		const presets = encounter.sim.db.getAllPresetEncounters();
		return {
			id: 'aem-encounter-picker',
			label: i18n.t('settings_tab.encounter.encounter_preset.label'),
			extraClassNames: ['pr-2', 'order-first', 'w-1/3', 'max-sm:w-1/2'],
			values: [{ name: i18n.t('common.custom'), value: -1 }, ...presets.map((preset, index) => ({ name: preset.path, value: index }))],
			storeField: 'encounter:*',
			getValue: (subject: Encounter) => presets.findIndex(preset => subject.matchesPreset(preset)),
			setValue: (subject: Encounter, newValue: number) => {
				if (newValue === -1) return;
				const preset = presets[newValue];
				trackEvent({ action: 'settings', category: 'encounter', label: 'preset', value: preset.path });
				subject.applyPreset(preset);
			},
		};
	}, [encounter]);

	const duration = useMemo(() => durationConfigs(), []);
	const execute = useMemo(() => executeConfigs(), []);

	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			testId="advanced-encounter-picker-modal"
			bodyClassName="gap-2 overflow-auto"
			headerChildren={<ConnectedChoiceField modObject={encounter} config={presetConfig} />}>
			<div className="flex flex-col gap-2" data-testid="encounter-header">
				<FieldGrid>
					{duration.map(config => (
						<ConnectedNumberField key={config.id} modObject={encounter} config={config} />
					))}
				</FieldGrid>
				<FieldGrid>
					{execute.map(config => (
						<ConnectedNumberField key={config.id} modObject={encounter} config={config} />
					))}
				</FieldGrid>
			</div>
			<TargetsPicker encounter={encounter} />
		</Dialog>
	);
};
