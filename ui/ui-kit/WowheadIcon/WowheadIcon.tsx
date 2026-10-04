import { WowheadAnchor, type WowheadAnchorProps } from '@ui-kit/WowheadAnchor';
import clsx from 'clsx';

export interface WowheadIconProps extends WowheadAnchorProps {
	/** A `div` takes the look, the label and the dataset, and none of the anchor's props. */
	as?: 'a' | 'div';
	iconUrl?: string;
	/** The `data-wowhead` tooltip attributes, spread onto the element. */
	wowhead?: Record<string, string>;
	label?: string;
}

export const WowheadIcon = ({ as = 'a', iconUrl, wowhead, label, className, testId, children, ...rest }: WowheadIconProps) => {
	const look = {
		className: clsx('relative inline-block bg-cover bg-center bg-no-repeat', className),
		'aria-label': label,
		style: iconUrl ? { backgroundImage: `url('${iconUrl}')` } : undefined,
		...wowhead,
	};

	if (as === 'div')
		return (
			<div data-testid={testId} {...look}>
				{children}
			</div>
		);

	return (
		<WowheadAnchor testId={testId} {...look} {...rest}>
			{children}
		</WowheadAnchor>
	);
};
