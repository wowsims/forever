import i18n from '@i18n/config';
import { formatToNumber } from '@sim/utils/format';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Spinner } from '@ui-kit/Spinner';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { TooltipButton } from '@ui-kit/TooltipButton';
import clsx from 'clsx';
import { useId } from 'react';

export interface TotalIterationsProps {
	/** A spinner replaces the number while the count is not final. */
	pending: boolean;
	iterations: number;
	overLimit: boolean;
	/** What the warning icon says when `overLimit`. */
	warningText: string;
}

// The Minimum iterations row of a run card's count list: a label with its help, and the number with the limit warning. In a
// narrow card the label wraps, and the help, the number and the warning stay on its first line.
export const TotalIterations = ({ pending, iterations, overLimit, warningText }: TotalIterationsProps) => {
	const tooltipId = useId();
	return (
		<>
			<dt className="flex items-start gap-1 self-start font-normal">
				<span>{i18n.t('bulk_tab.settings.total_iterations')}</span>
				<span className="flex h-lh shrink-0 items-center">
					<TooltipButton
						tooltip={i18n.t('bulk_tab.settings.total_iterations_tooltip')}
						label={i18n.t('common.help')}
						testId="bulk-total-iterations-help"
					/>
				</span>
			</dt>
			<dd
				className={clsx('mb-0 flex items-center justify-end gap-2 self-start font-bold tabular-nums', overLimit && 'text-danger')}
				data-testid="bulk-total-iterations"
				data-over-limit={overLimit ? '' : undefined}>
				{overLimit && (
					<>
						<Button
							variant="unstyled"
							aria-label={warningText}
							className="text-link-warning"
							data-testid="warning"
							{...tooltipAnchorProps(tooltipId)}>
							<Icon name="triangle-exclamation" />
						</Button>
						<Tooltip id={tooltipId} place="left" content={warningText} />
					</>
				)}
				{pending ? <Spinner size="sm" /> : formatToNumber(iterations)}
			</dd>
		</>
	);
};
