import { Button } from '@ui-kit/Button';
import { Icon, type IconName } from '@ui-kit/Icon';
import { SidebarDisabledContext } from '@ui-kit/SidebarActionButton';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { use, useId } from 'react';

export interface SidebarIconActionProps {
	icon: IconName;
	/** The button's accessible name and its tooltip. */
	label: string;
	testId: string;
	onClick: () => void;
	disabled?: boolean;
}

/** A square sidebar button that shows only its icon and names itself in a tooltip. */
export const SidebarIconAction = ({ icon, label, testId, onClick, disabled }: SidebarIconActionProps) => {
	const sidebarDisabled = use(SidebarDisabledContext);
	const tooltipId = useId();
	return (
		<>
			<Button
				variant="outline-primary"
				// Square: the same padding on every side around a box as wide as the 1.5 line height.
				className="shrink-0 px-(--control-padding-y)"
				aria-label={label}
				data-testid={testId}
				data-sidebar-action=""
				disabled={disabled || sidebarDisabled}
				onClick={onClick}
				{...tooltipAnchorProps(tooltipId)}>
				<span className="inline-flex w-[1.5em] justify-center">
					<Icon name={icon} />
				</span>
			</Button>
			<Tooltip id={tooltipId} content={label} place="bottom" />
		</>
	);
};
