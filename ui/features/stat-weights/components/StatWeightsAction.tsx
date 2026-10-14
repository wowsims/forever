import i18n from '@i18n/config';
import { useSim } from '@sim/context/SimHostContext';
import { useSimStatus } from '@sim/hooks/useSimStatus';
import { SidebarIconAction } from '@ui-kit/SidebarIconAction';

import { trackPageView } from '../../../tracking/utils';
import { useOpenEpWeights } from '../hooks/useEpWeightsDialog';

export const StatWeightsAction = () => {
	const openEpWeights = useOpenEpWeights();
	// Weights are computed against the item database, so the button stays disabled until init settles,
	// and for good if it fails.
	const { status } = useSimStatus(useSim());

	return (
		<SidebarIconAction
			icon="scale-balanced"
			label={i18n.t('sidebar.buttons.stat_weights.title')}
			testId="ep-weights-action"
			disabled={status !== 'ready'}
			onClick={() => {
				trackPageView('Stat Weights', '/stat-weights');
				openEpWeights();
			}}
		/>
	);
};
