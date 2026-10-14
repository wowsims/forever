import i18n from '@i18n/config';
import type { CastLog } from '@sim/proto/combat_log';

export interface GcdSegmentTooltipProps {
	log: CastLog;
}

export const GcdSegmentTooltip = ({ log }: GcdSegmentTooltipProps) => (
	<div className="ui-timeline-tooltip">
		<span>
			{i18n.t('results_tab.details.timeline.gcd_tooltip', {
				spell: log.actionId!.name,
				gcd: log.gcd.toFixed(2),
				start: log.timestamp.toFixed(2),
				end: (log.timestamp + log.gcd).toFixed(2),
			})}
		</span>
	</div>
);
