import { useActionId } from '@sim/hooks/useActionId';
import { useActionIdWowheadDataset } from '@sim/hooks/useActionIdWowheadDataset';
import type { ActionId } from '@sim/proto/action_id';
import { WowheadIcon } from '@ui-kit/WowheadIcon';

export interface ActionIdIconProps {
	actionId: ActionId;
	/** Aura sets show the buff's own tooltip rather than the spell that applies it. */
	useBuffAura: boolean;
}

/** The icon beside one spell or aura in an action-id menu. */
export const ActionIdIcon = ({ actionId, useBuffAura }: ActionIdIconProps) => {
	const { iconUrl, href } = useActionId(actionId);
	const wowheadProps = useActionIdWowheadDataset(actionId, useBuffAura);

	return <WowheadIcon className="mr-1 size-4" href={href || undefined} iconUrl={iconUrl || undefined} {...wowheadProps} />;
};
