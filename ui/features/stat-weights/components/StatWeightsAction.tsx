import i18n from '@i18n/config';
import { useSim } from '@sim/context/SimHostContext';
import { useSimStatus } from '@sim/hooks/useSimStatus';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { SidebarDisabledContext } from '@ui-kit/SidebarActionButton';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { use, useId } from 'react';

import { trackPageView } from '../../../tracking/utils';
import { useOpenEpWeights } from '../hooks/useEpWeightsDialog';

export const StatWeightsAction = () => {
	const openEpWeights = useOpenEpWeights();
	const sidebarDisabled = use(SidebarDisabledContext);
	const tooltipId = useId();
	const label = i18n.t('sidebar.buttons.stat_weights.title');
	// Weights are computed against the item database, so the button stays disabled until init settles,
	// and for good if it fails.
	const { status } = useSimStatus(useSim());

	return (
		<>
			<Button
				variant="outline-primary"
				aria-label={label}
				data-testid="ep-weights-action"
				data-sidebar-action=""
				disabled={status !== 'ready' || sidebarDisabled}
				onClick={() => {
					trackPageView('Stat Weights', '/stat-weights');
					openEpWeights();
				}}
				{...tooltipAnchorProps(tooltipId)}>
				<Icon name="scale-balanced" />
			</Button>
			<Tooltip id={tooltipId} content={label} place="bottom" />
		</>
	);
};
