import i18n from '@i18n/config';
import { useSim } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import type { CastLog } from '@sim/proto/combat_log';

import { DamageResult } from '../../DamageResult';

export interface CastTooltipProps {
	log: CastLog;
	/** The casts drawn under this one's icon, ending with `log`; their damage is listed together. */
	stack?: ReadonlyArray<CastLog>;
}

export const CastTooltip = ({ log, stack = [log] }: CastTooltipProps) => {
	const { threat: showThreatMetrics } = useDisplayMetrics(useSim());
	const travelTime = log.travelTime == 0 ? '' : ` ${i18n.t('results_tab.details.timeline.cast_tooltip.travel_time', { time: log.travelTime.toFixed(2) })}`;
	// Two casts at the same timestamp can hand all of their damage to the later one, so the stack's
	// damage is one list rather than one per cast.
	const damageDealtLogs = stack.flatMap(cast => cast.damageDealtLogs);
	const totalDamage = damageDealtLogs.reduce((total, ddl) => total + ddl.amount, 0);
	const effectiveTime = stack.reduce((total, cast) => total + cast.effectiveTime, 0);
	const first = stack[0];
	const timing = log.castCancelledLog
		? i18n.t('results_tab.details.timeline.cast_tooltip.cancelled', { time: log.cancelTime.toFixed(2) })
		: log.castTime > 0
			? i18n.t('results_tab.details.timeline.cast_tooltip.cast_and_gcd_time', {
					castTime: log.castTime.toFixed(2),
					gcdTime: log.effectiveTime.toFixed(2),
				})
			: i18n.t('results_tab.details.timeline.cast_tooltip.gcd_time', { gcdTime: log.effectiveTime.toFixed(2) });

	return (
		<div className="ui-timeline-tooltip">
			<span>
				{i18n.t('results_tab.details.timeline.cast_tooltip.window', {
					spell: log.actionId!.name,
					start: log.timestamp.toFixed(2),
					end: (log.castCancelledLog ? log.castCancelledLog.timestamp : log.timestamp + log.castTime).toFixed(2),
				})}
				{` ${timing}`}
				{travelTime.length > 0 && travelTime}
			</span>
			{stack.length > 1 && (
				<span>
					{first.timestamp === log.timestamp
						? i18n.t('results_tab.details.timeline.cast_tooltip.stack_at', { casts: stack.length, time: log.timestamp.toFixed(2) })
						: i18n.t('results_tab.details.timeline.cast_tooltip.stack_range', {
								casts: stack.length,
								start: first.timestamp.toFixed(2),
								end: log.timestamp.toFixed(2),
							})}
				</span>
			)}
			{totalDamage > 0 && (
				<span>
					{i18n.t('results_tab.details.timeline.cast_tooltip.total', {
						total: totalDamage.toFixed(2),
						dpet: (totalDamage / (effectiveTime || 1)).toFixed(2),
					})}
				</span>
			)}
			{damageDealtLogs.length > 0 && (
				<ul className="ui-timeline-tooltip-hits">
					{damageDealtLogs.map((ddl, index) => (
						<li key={index}>
							<span>
								{ddl.timestamp.toFixed(2)}s - <DamageResult log={ddl} />
							</span>
							{showThreatMetrics && !ddl.source?.isTarget && (
								<span>
									{' '}
									({ddl.threat.toFixed(1)} {i18n.t('results_tab.details.timeline.tooltips.threat')})
								</span>
							)}
						</li>
					))}
				</ul>
			)}
		</div>
	);
};
