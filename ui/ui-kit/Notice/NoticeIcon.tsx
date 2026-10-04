import { Button } from '@ui-kit/Button';
import { Icon, type IconSize } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { type ReactNode, useId } from 'react';

import { NOTICE_COLOR, NOTICE_ICON_NAME, type NoticeLevel } from './noticeLevel';

export interface NoticeIconProps {
	level: NoticeLevel;
	/** The button's accessible name. */
	label: string;
	/** The tooltip's content. */
	children: ReactNode;
	size?: IconSize;
	/** Rendered but `display: none`, so a header's element count stays stable. */
	hidden?: boolean;
	className?: string;
	testId?: string;
}

export const NoticeIcon = ({ level, label, children, size = 'xl', hidden, className, testId }: NoticeIconProps) => {
	const tooltipId = useId();
	return (
		<>
			<Button
				iconOnly
				aria-label={label}
				className={className}
				data-level={level}
				data-testid={testId}
				style={hidden === undefined ? undefined : { display: hidden ? 'none' : undefined }}
				{...tooltipAnchorProps(tooltipId)}>
				<Icon name={NOTICE_ICON_NAME[level]} size={size} fixedWidth className={NOTICE_COLOR[level]} />
			</Button>
			<Tooltip id={tooltipId} content={children} clickable />
		</>
	);
};
