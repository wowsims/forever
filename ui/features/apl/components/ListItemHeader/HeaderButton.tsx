import { Button } from '@ui-kit/Button';
import { Icon, type IconName } from '@ui-kit/Icon';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';

export interface HeaderButtonProps {
	icon: IconName;
	label: string;
	className?: string;
	testId: string;
	tooltip?: string;
	tooltipId?: string;
	/** Kept in the header but not displayed, so the header's element count stays the same. */
	hidden?: boolean;
	/** Renders as `aria-pressed`; the hide eye sets it while its item is hidden. */
	pressed?: boolean;
	onClick?: () => void;
}

/** An icon button in an APL item's title: the validations badge and the hide eye. */
export const HeaderButton = ({ icon, label, className, testId, tooltip, tooltipId, hidden, pressed, onClick }: HeaderButtonProps) => (
	<Button
		iconOnly
		className={className}
		aria-label={label}
		aria-pressed={pressed}
		data-testid={testId}
		style={hidden === undefined ? undefined : { display: hidden ? 'none' : undefined }}
		onClick={onClick}
		{...tooltipAnchorProps(tooltipId, tooltip)}>
		<Icon name={icon} style="base" size="xl" />
	</Button>
);
