import type { ResolvedIcon } from '@ui-kit/Field/types';
import { WowheadAnchor, type WowheadAnchorProps } from '@ui-kit/WowheadAnchor';
import clsx from 'clsx';
import type { ReactNode } from 'react';

import { resolvedIconStyle } from '../utils/icons';

export interface IconSwatchProps extends Omit<WowheadAnchorProps, 'children' | 'icon'> {
	/** The background; `style` replaces it. The link is `href`, which the swatch never takes from the icon. */
	icon?: ResolvedIcon;
	active?: boolean;
	disabled?: boolean;
	/** Written over the bottom of the swatch. */
	text?: string;
	/** Anchors laid over the swatch. They sit beside it in one positioned box, since an anchor cannot hold another. */
	badges?: ReactNode;
}

/** An icon as a link: wowhead's tooltip script finds it through the `href`. `className` lands on the box when there are badges. */
export const IconSwatch = ({ icon, active, disabled, text, badges, className, style, ...rest }: IconSwatchProps) => {
	const anchor = (
		<WowheadAnchor
			className={clsx('ui-icon-picker-swatch', !badges && className)}
			data-active={active ? '' : undefined}
			data-disabled={disabled ? '' : undefined}
			disabled={disabled}
			style={style ?? resolvedIconStyle(icon)}
			{...rest}>
			{text !== undefined && <span className="ui-icon-swatch-text">{text}</span>}
		</WowheadAnchor>
	);

	if (!badges) return anchor;

	return (
		<div className={clsx('ui-icon-swatch-box', className)}>
			{anchor}
			{badges}
		</div>
	);
};
