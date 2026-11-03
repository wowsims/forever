import type { ResolvedIcon } from '@ui-kit/Field/types';
import { IconSwatch } from '@ui-kit/IconSwatch';
import clsx from 'clsx';
import type { MouseEventHandler } from 'react';

export interface ImprovedAnchorEvents {
	onActivate: () => void;
	onContextMenu: MouseEventHandler;
	onMouseDown: MouseEventHandler;
}

export interface ImprovedAnchorProps {
	icon: ResolvedIcon;
	testId: string;
	active: boolean;
	disabled?: boolean;
	/** Which bottom corner of the icon the badge sits in: the two improved badges share the box and must not stack. */
	corner: 'left' | 'right';
	/** Left out, the badge is display only: it links nowhere and lets the pointer through to the icon under it. */
	events?: ImprovedAnchorEvents;
}

export const ImprovedAnchor = ({ icon, testId, active, disabled, corner, events }: ImprovedAnchorProps) => (
	<IconSwatch
		icon={icon}
		href={events && icon.href}
		active={active}
		disabled={disabled}
		testId={testId}
		className={clsx(
			'absolute bottom-0 size-4 min-w-4',
			corner === 'left' ? 'left-px' : 'right-0',
			!active && 'border-gray-600 grayscale',
			!events && 'pointer-events-none',
		)}
		{...events}
	/>
);
