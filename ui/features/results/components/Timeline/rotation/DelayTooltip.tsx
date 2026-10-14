import i18n from '@i18n/config';
import type { CastLog } from '@sim/proto/combat_log';

export interface DelayTooltipProps {
	log: CastLog;
}

export const DelayTooltip = ({ log }: DelayTooltipProps) => (
	<div className="ui-timeline-tooltip">
		<span>{i18n.t('results_tab.details.timeline.delay_tooltip', { delay: log.delayText, readyAt: log.readyAtText })}</span>
	</div>
);
