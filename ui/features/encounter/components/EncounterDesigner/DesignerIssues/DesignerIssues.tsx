import type { EncounterDesign } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Notice } from '@ui-kit/Notice';

import type { DesignEventRef } from '../../../model/design/lanes';
import type { DesignIssue } from '../../../model/design/validate';
import { designerText, eventName } from '../utils/configs';

export interface DesignerIssuesProps {
	design: EncounterDesign;
	issues: ReadonlyArray<DesignIssue>;
	onSelect: (ref: DesignEventRef) => void;
}

export const DesignerIssues = ({ design, issues, onSelect }: DesignerIssuesProps) => {
	const title = designerText('issues.title');
	return (
		<section aria-label={title} className="flex flex-col gap-2">
			<span className="text-sm font-bold text-white">{title}</span>
			{issues.length === 0 ? (
				<span className="text-sm text-foreground-disabled">{designerText('issues.empty')}</span>
			) : (
				<ul className="m-0 flex list-none flex-col gap-1 p-0">
					{issues.map(issue => (
						<li key={`${issue.ref.kind}:${issue.ref.index}:${issue.key}`}>
							<Notice level={issue.severity}>
								<Button variant="link" size="none" onClick={() => onSelect(issue.ref)}>
									{eventName(design, issue.ref)}
								</Button>{' '}
								{i18n.t(issue.key)}
							</Notice>
						</li>
					))}
				</ul>
			)}
		</section>
	);
};
