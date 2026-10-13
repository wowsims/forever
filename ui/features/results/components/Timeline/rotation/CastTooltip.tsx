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
	const travelTime = log.travelTime == 0 ? '' : ` + ${log.travelTime.toFixed(2)}s travel time`;
	// Two casts at the same timestamp can hand all of their damage to the later one, so the stack's
	// damage is one list rather than one per cast.
	const damageDealtLogs = stack.flatMap(cast => cast.damageDealtLogs);
	const totalDamage = damageDealtLogs.reduce((total, ddl) => total + ddl.amount, 0);
	const effectiveTime = stack.reduce((total, cast) => total + cast.effectiveTime, 0);
	const first = stack[0];

	return (
		<div className="ui-timeline-tooltip">
			<span>
				{log.actionId!.name} from {log.timestamp.toFixed(2)}s to{' '}
				{(log.castCancelledLog ? log.castCancelledLog.timestamp : log.timestamp + log.castTime).toFixed(2)}s
				{log.castCancelledLog
					? ` (Cancelled after ${log.cancelTime.toFixed(2)}s)`
					: ` (${log.castTime > 0 ? `${log.castTime.toFixed(2)}s, ` : ''}${log.effectiveTime.toFixed(2)}s GCD Time)`}
				{travelTime.length > 0 && travelTime}
			</span>
			{stack.length > 1 && (
				<span>
					{stack.length} casts {first.timestamp === log.timestamp ? 'at' : `from ${first.timestamp.toFixed(2)}s to`} {log.timestamp.toFixed(2)}s
				</span>
			)}
			{totalDamage > 0 && (
				<span>
					Total: {totalDamage.toFixed(2)} ({(totalDamage / (effectiveTime || 1)).toFixed(2)} DPET)
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
