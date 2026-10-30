import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import { useSimHost } from '@sim/context/SimHostContext';
import { Dialog } from '@ui-kit/Dialog';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { useMemo } from 'react';

import { useEncounterChoice } from '../../hooks/useEncounterChoice';
import { durationConfigs, executeConfigs } from '../EncounterPicker/utils/configs';
import { TargetsPicker } from '../TargetsPicker';

export interface AdvancedEncounterModalProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

const CHOICE_FIELD = { id: 'aem-encounter-picker', extraClassNames: ['pr-2', 'order-first', 'w-1/3', 'max-sm:w-1/2'] };

export const AdvancedEncounterModal = ({ open, onOpenChange }: AdvancedEncounterModalProps) => {
	const host = useSimHost();
	const encounter = host.sim.encounter;
	const choice = useEncounterChoice(encounter, CHOICE_FIELD);

	const duration = useMemo(() => durationConfigs(), []);
	const execute = useMemo(() => executeConfigs(), []);

	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			testId="advanced-encounter-picker-modal"
			bodyClassName="gap-2 overflow-auto"
			headerChildren={<ConnectedChoiceField {...choice} />}>
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
