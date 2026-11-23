import type { ResolvedIcon } from '@ui-kit/Field/types';
import { Icon } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { WowheadAnchor, type WowheadAnchorProps } from '@ui-kit/WowheadAnchor';
import clsx from 'clsx';
import { type ReactNode, useId } from 'react';

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
	/**
	 * Set for a selection the caller keeps on offer only because it is selected: dims the swatch, with its badges, and
	 * a corner badge beside it shows this on hover, so the swatch keeps its own link tooltip.
	 */
	keptTooltip?: string;
	/** Set for a selection another selection makes add nothing: dims the swatch the same way, with a badge of its own in the other corner. */
	overwrittenTooltip?: string;
	/** With `overwrittenTooltip`: the selection still adds something, so the badge shows and the swatch stays lit. */
	overwrittenPartly?: boolean;
}

/** An icon as a link: wowhead's tooltip script finds it through the `href`. `className` lands on the box when there are badges or a corner badge. */
export const IconSwatch = ({
	icon,
	active,
	disabled,
	text,
	badges,
	keptTooltip,
	overwrittenTooltip,
	overwrittenPartly,
	className,
	style,
	...rest
}: IconSwatchProps) => {
	const tooltipId = useId();
	const noticed = !!keptTooltip || !!overwrittenTooltip;
	const boxed = !!badges || noticed;
	const anchor = (
		<WowheadAnchor
			className={clsx('ui-icon-picker-swatch', !boxed && className)}
			data-active={active ? '' : undefined}
			data-disabled={disabled ? '' : undefined}
			disabled={disabled}
			style={style ?? resolvedIconStyle(icon)}
			{...rest}>
			{text !== undefined && <span className="ui-icon-swatch-text">{text}</span>}
		</WowheadAnchor>
	);

	if (!boxed) return anchor;

	const box = (
		<div
			className={clsx('ui-icon-swatch-box', className)}
			data-kept={keptTooltip ? '' : undefined}
			data-overwritten={overwrittenTooltip && !overwrittenPartly ? '' : undefined}>
			{anchor}
			{badges}
			{overwrittenTooltip && (
				<Icon
					name="circle-exclamation"
					className="ui-icon-swatch-badge ui-icon-swatch-badge-overwritten"
					title={overwrittenTooltip}
					testId="icon-swatch-overwritten"
					{...tooltipAnchorProps(tooltipId, overwrittenTooltip)}
				/>
			)}
			{keptTooltip && (
				<Icon
					name="eye-slash"
					className="ui-icon-swatch-badge ui-icon-swatch-badge-kept"
					title={keptTooltip}
					testId="icon-swatch-kept"
					{...tooltipAnchorProps(tooltipId, keptTooltip)}
				/>
			)}
		</div>
	);

	if (!noticed) return box;

	return (
		<>
			{box}
			{/* A badge's tooltip keeps its line breaks. */}
			<Tooltip id={tooltipId} className="whitespace-pre-line" />
		</>
	);
};
