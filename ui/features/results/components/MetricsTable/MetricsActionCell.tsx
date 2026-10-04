import { useActionId } from '@sim/hooks/useActionId';
import { useActionIdWowheadDataset } from '@sim/hooks/useActionIdWowheadDataset';
import type { ActionId } from '@sim/proto/action_id';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { WowheadIcon } from '@ui-kit/WowheadIcon';

export interface MetricsActionCellProps {
	name: string;
	actionId: ActionId;
	useBuffAura?: boolean;
	expandable: boolean;
	expanded: boolean;
	onToggle: () => void;
}

export const MetricsActionCell = ({ name, actionId, useBuffAura, expandable, expanded, onToggle }: MetricsActionCellProps) => {
	const { iconUrl, href } = useActionId(actionId);
	const wowhead = useActionIdWowheadDataset(actionId, useBuffAura);

	return (
		<div data-testid="metrics-action" className="flex items-center gap-2 whitespace-normal">
			<WowheadIcon
				testId="metrics-action-icon"
				className="z-1 mr-1 h-6 w-6 align-middle"
				label={name}
				href={href || undefined}
				iconUrl={iconUrl}
				wowhead={wowhead}
			/>
			<span data-testid="metrics-action-name" className="truncate">
				{name}
			</span>
			{expandable && (
				<Button
					variant="unstyled"
					data-testid="expand-toggle"
					className="relative z-1 ml-auto border-0 bg-transparent p-0 text-inherit focus-visible:focus-ring"
					aria-expanded={expanded}
					aria-label={name}
					onClick={onToggle}>
					<Icon name={expanded ? 'caret-down' : 'caret-right'} style="base" />
				</Button>
			)}
		</div>
	);
};
