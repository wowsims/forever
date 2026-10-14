import i18n from '@i18n/config';
import { useSim, useSimHost } from '@sim/context/SimHostContext';
import { useSimStatus } from '@sim/hooks/useSimStatus';
import { useUiStore } from '@sim/hooks/useUiStore';
import type { StatSweepSettings } from '@sim/settings/stat_sweep_settings';
import { SidebarIconAction } from '@ui-kit/SidebarIconAction';
import { useState } from 'react';

import { trackPageView } from '../../../../tracking/utils';
import { StatSweepDialog } from '../StatSweepDialog';

export interface StatSweepActionProps {
	settings: StatSweepSettings;
}

export const StatSweepAction = ({ settings }: StatSweepActionProps) => {
	const host = useSimHost();
	const shown = useUiStore('showPowerTools');
	const { status } = useSimStatus(useSim());
	const [open, setOpen] = useState(false);

	if (!shown || host.simDisabled) return null;

	return (
		<>
			<SidebarIconAction
				icon="stairs"
				label={i18n.t('sidebar.buttons.stat_sweep.title')}
				testId="stat-sweep-action"
				disabled={status !== 'ready'}
				onClick={() => {
					trackPageView('Stat Sweep', '/stat-sweep');
					setOpen(true);
				}}
			/>
			<StatSweepDialog open={open} onOpenChange={setOpen} settings={settings} />
		</>
	);
};
