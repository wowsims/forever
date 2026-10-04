import { config } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import clsx from 'clsx';
import type { HTMLAttributes } from 'react';

import { iconDefinition } from './definitions';
import type { IconName, IconSize, IconStyle } from './types';

// ui/styles/main.css imports FontAwesome's stylesheet into the base layer, so `ui-icon` and utilities override it.
config.autoAddCss = false;

// `style` is the FontAwesome style here, not a CSS object, so it shadows the DOM prop.
export interface IconProps extends Omit<HTMLAttributes<HTMLElement>, 'style'> {
	name: IconName;
	style?: IconStyle;
	size?: IconSize;
	/** One width for every glyph (`fa-fw`), so icons in a column line up; otherwise the icon is as wide as its glyph (`fa-width-auto`). */
	fixedWidth?: boolean;
	className?: string;
	/** The icon's accessible name (`aria-label`); without one the icon is hidden from assistive tech. */
	title?: string;
	testId?: string;
}

// The rest props matter: an icon is an anchor for its own tooltip at several sites, and a dropped `data-tooltip-id` fails silently.
export const Icon = ({ name, style = 'solid', size, fixedWidth, className, title, testId, ...rest }: IconProps) => {
	const definition = iconDefinition(name, style);
	return (
		<FontAwesomeIcon
			data-testid={testId}
			{...(rest as Omit<HTMLAttributes<SVGSVGElement>, 'style'>)}
			icon={definition}
			className={clsx('ui-icon', className)}
			data-size={size}
			size={size}
			fixedWidth={fixedWidth}
			widthAuto={!fixedWidth}
			aria-label={title}
		/>
	);
};
