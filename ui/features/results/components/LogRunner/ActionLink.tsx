import { useActionId } from '@sim/hooks/useActionId';
import { useActionIdWowheadDataset } from '@sim/hooks/useActionIdWowheadDataset';
import type { ActionId } from '@sim/proto/action_id';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';

export interface ActionLinkProps {
	actionId: ActionId;
	/** Resolves the tooltip against the buff aura rather than the spell — an aura line names the aura. */
	isAura?: boolean;
}

/**
 * The Wowhead tooltip this anchors is appended to `<body>` by Wowhead's own script, not rendered
 * here, so it is not a `position: fixed` descendant of the virtual list's transformed row.
 */
export const ActionLink = ({ actionId, isAura }: ActionLinkProps) => {
	const { iconUrl, name, href } = useActionId(actionId);
	const wowheadProps = useActionIdWowheadDataset(actionId, isAura);

	return (
		<WowheadAnchor testId="log-action" className="text-brand" target="_blank" href={href || undefined} {...wowheadProps}>
			<span>
				<span
					data-testid="log-action-icon"
					className="inline-block size-4 bg-cover align-middle"
					style={iconUrl ? { backgroundImage: `url('${iconUrl}')` } : undefined}
				/>{' '}
				{name}
			</span>
		</WowheadAnchor>
	);
};
