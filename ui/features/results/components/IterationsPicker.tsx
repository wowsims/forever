import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import i18n from '@i18n/config';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Sim } from '@sim/sim';
import { subscribeSimField } from '@sim/state/subscriptions';
import { formatToNumber } from '@sim/utils/format';
import { Button } from '@ui-kit/Button';

import { trackEvent } from '../../../tracking/utils';

const PRESETS = [Sim.DEFAULT_ITERATIONS, 2 * Sim.DEFAULT_ITERATIONS];

export interface IterationsPickerProps {
	sim: Sim;
}

const setIterations = (sim: Sim, newValue: number) => {
	trackEvent({
		action: 'settings',
		category: 'iterations',
		label: 'update',
		value: newValue,
	});
	sim.setIterations(newValue);
};

export const IterationsPicker = ({ sim }: IterationsPickerProps) => {
	const iterations = useStoreSubscribe(subscribeSimField(sim, 'iterations'), () => sim.getIterations());

	return (
		<div className="flex w-full items-end gap-2" data-testid="iterations-row">
			<ConnectedNumberField
				modObject={sim}
				testId="iterations-picker"
				inputClassName="m-0 w-full"
				size="md"
				config={{
					id: 'simui-iterations',
					label: i18n.t('sidebar.iterations'),
					extraClassNames: ['min-w-0 flex-1'],
					storeSubscribe: (sim: Sim) => subscribeSimField(sim, 'iterations'),
					getValue: (sim: Sim) => sim.getIterations(),
					setValue: setIterations,
				}}
			/>
			{PRESETS.map(preset => (
				<Button
					key={preset}
					size="md"
					variant={iterations === preset ? 'primary' : 'outline-primary'}
					aria-pressed={iterations === preset}
					data-testid="iterations-preset"
					onClick={() => setIterations(sim, preset)}>
					{formatToNumber(preset, { notation: 'compact' })}
				</Button>
			))}
		</div>
	);
};
