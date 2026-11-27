import { Tabs } from '@base-ui/react/tabs';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import type { BulkMode } from '@sim/state/sim_store';
import { TabNav, type TabNavProps } from '@ui-kit/TabNav';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { cloneElement, type ReactElement, useId } from 'react';

import { useBulkMode } from '../../hooks/useBulkMode';
import { bulkMenuModes, setBulkMode } from '../../model/mode';

const tabId = (mode: BulkMode) => `bulk-mode-${mode}`;

export interface BulkModeTabsViewProps {
	modes: readonly BulkMode[];
	current: BulkMode;
	onPick: (mode: BulkMode) => void;
	placement?: TabNavProps['placement'];
	className?: string;
}

// The Upgrades side's tab strip: one tab per batch mode, in the place the single sim's tabs take. It has its own
// `Tabs.Root` with no panels, since every mode shares the one batch pane.
export const BulkModeTabsView = ({ modes, current, onPick, placement = 'top', className }: BulkModeTabsViewProps) => {
	const tooltipId = useId();
	return (
		<Tabs.Root className="contents" value={tabId(current)} onValueChange={next => onPick(modes.find(mode => tabId(mode) === next) ?? current)}>
			<TabNav
				variant="sim"
				placement={placement}
				className={className}
				testId="bulk-mode-tabs"
				tabs={modes.map(mode => ({
					id: tabId(mode),
					label: i18n.t(`bulk_tab.modes.${mode}.label`),
					renderTab: (tab: ReactElement) => cloneElement(tab, tooltipAnchorProps(tooltipId, i18n.t(`bulk_tab.modes.${mode}.description`))),
				}))}
			/>
			<Tooltip id={tooltipId} place={placement === 'bottom' ? 'top' : 'bottom'} />
		</Tabs.Root>
	);
};

export const BulkModeTabs = ({ placement, className }: Pick<BulkModeTabsViewProps, 'placement' | 'className'>) => {
	const player = usePlayer();
	const { powerTools, runMode } = useBulkMode();
	return (
		<BulkModeTabsView
			modes={bulkMenuModes(powerTools)}
			current={runMode}
			onPick={mode => setBulkMode(player, mode)}
			placement={placement}
			className={className}
		/>
	);
};
