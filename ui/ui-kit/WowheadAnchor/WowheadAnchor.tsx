import { externalRel } from '@sim/utils/links';
import type { KeyboardEvent, MouseEvent } from 'react';

import { wowheadAnchorProps } from '../utils/wowhead';
import type { WowheadAnchorProps } from './types';

/**
 * The anchor behind every wowhead link, swatch and icon: wowhead's tooltip script finds it through the `href` and the dataset.
 * An anchor that acts takes a native button's keys, Enter on keydown and Space on keyup, so a Base UI trigger rendered onto it sees each key as
 * already handled. With an href, Enter stays the link's own, whose click acts.
 */
export const WowheadAnchor = ({
	href,
	icon = true,
	navigate = true,
	onActivate,
	disabled,
	onClick,
	onKeyDown,
	onKeyUp,
	role,
	tabIndex,
	rel,
	testId,
	ref,
	children,
	...rest
}: WowheadAnchorProps) => {
	const button = !!onActivate && !href;
	const cancels = !navigate || !!onActivate;

	return (
		<a
			ref={ref}
			data-testid={testId}
			{...wowheadAnchorProps({ icon: icon && !!href, acts: !!onActivate })}
			{...rest}
			href={href || undefined}
			rel={externalRel(href || undefined, rel)}
			role={role ?? (button ? 'button' : undefined)}
			aria-disabled={disabled ? 'true' : undefined}
			tabIndex={tabIndex ?? (disabled && !!onActivate ? -1 : button ? 0 : undefined)}
			onClick={
				cancels
					? (event: MouseEvent<HTMLAnchorElement>) => {
							event.preventDefault();
							if (!disabled) onActivate?.(event);
							onClick?.(event);
						}
					: onClick
			}
			onKeyDown={
				onActivate
					? (event: KeyboardEvent<HTMLAnchorElement>) => {
							if (event.key === ' ' || (button && event.key === 'Enter')) event.preventDefault();
							if (button && event.key === 'Enter') event.currentTarget.click();
							onKeyDown?.(event);
						}
					: onKeyDown
			}
			onKeyUp={
				onActivate
					? (event: KeyboardEvent<HTMLAnchorElement>) => {
							if (event.key === ' ') {
								event.preventDefault();
								event.currentTarget.click();
							}
							onKeyUp?.(event);
						}
					: onKeyUp
			}>
			{children}
		</a>
	);
};
