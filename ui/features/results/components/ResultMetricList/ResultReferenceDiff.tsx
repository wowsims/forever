import i18n from '@i18n/config';
import { Pill } from '@ui-kit/Pill';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';

import type { ReferenceDiff } from '../../model/reference_diffs';

export interface ResultReferenceDiffProps {
	diff?: ReferenceDiff;
	tooltipId?: string;
	/** What the delta is measured against; "vs ref" when left out, nothing when `null`. */
	label?: string | null;
}

/** The delta against the saved reference run. Every layout renders the slot; only the sidebar, and only once a reference is set, fills it in. */
export const ResultReferenceDiff = ({ diff, tooltipId, label = i18n.t('sidebar.results.reference.vs_ref') }: ResultReferenceDiffProps) =>
	diff ? (
		<Pill tone={diff.tone ?? 'neutral'} className="flex" testId="results-reference">
			<span data-testid="results-reference-diff" data-sign={diff.tone ?? undefined} {...tooltipAnchorProps(tooltipId, diff.significance)}>
				{diff.text}
			</span>
			{label}
		</Pill>
	) : null;
