import type { EncounterPlan } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Notice } from '@ui-kit/Notice';

import type { PlanEventRef } from '../../model/lanes';
import { fractionToSeconds } from '../../model/time';
import type { PlanIssue } from '../../model/validate';
import { plannerText } from '../EncounterPlanner/utils/configs';
import { eventName, formatSeconds } from '../EncounterPlanner/utils/tips';

export interface PlannerIssuesProps {
	plan: EncounterPlan;
	issues: ReadonlyArray<PlanIssue>;
	/** The duration an issue's span is told in. */
	baseSeconds: number;
	onSelect: (ref: PlanEventRef) => void;
}

export const PlannerIssues = ({ plan, issues, baseSeconds, onSelect }: PlannerIssuesProps) => {
	const title = plannerText('issues.title');
	const seconds = (fraction: number) => formatSeconds(fractionToSeconds(fraction, baseSeconds));
	return (
		<section aria-label={title} className="flex flex-col gap-2">
			<span className="text-sm font-bold text-white">{title}</span>
			{issues.length === 0 ? (
				<span className="text-sm text-foreground-disabled">{plannerText('issues.empty')}</span>
			) : (
				<ul className="m-0 flex list-none flex-col gap-1 p-0">
					{issues.map(issue => (
						<li key={`${issue.ref.kind}:${issue.ref.index}:${issue.key}:${issue.span?.start ?? ''}`}>
							<Notice level={issue.severity}>
								<Button variant="link" size="none" className="text-[length:inherit]" onClick={() => onSelect(issue.ref)}>
									{eventName(plan, issue.ref)}
								</Button>
								{': '}
								{issue.span ? i18n.t(issue.key, { from: seconds(issue.span.start), to: seconds(issue.span.end) }) : i18n.t(issue.key)}
							</Notice>
						</li>
					))}
				</ul>
			)}
		</section>
	);
};
