import { useActionId } from '@sim/hooks/useActionId';
import { useActionIdWowheadDataset } from '@sim/hooks/useActionIdWowheadDataset';
import type { ActionId } from '@sim/proto/action_id';
import { WowheadIcon } from '@ui-kit/WowheadIcon';

export interface RotationRowIconProps {
	actionId: ActionId;
	/** Absent for a section header, whose icon links out but carries no Wowhead tooltip. */
	tooltip?: 'spell' | 'buffAura';
}

export const RotationRowIcon = ({ actionId, tooltip }: RotationRowIconProps) => {
	const { iconUrl, href } = useActionId(actionId);
	const wowhead = useActionIdWowheadDataset(tooltip ? actionId : null, tooltip === 'buffAura');

	return <WowheadIcon testId="rotation-row-icon" className="ui-timeline-row-icon" href={href || undefined} iconUrl={iconUrl} wowhead={wowhead} />;
};
