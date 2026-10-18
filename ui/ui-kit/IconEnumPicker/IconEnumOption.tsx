import { Menu } from '@base-ui/react/menu';
import { IconSwatch } from '@ui-kit/IconSwatch';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';

import type { IconEnumOption } from './types';
import { iconStyleOf } from './utils';

export interface IconEnumOptionItemProps<T> {
	option: IconEnumOption<T>;
	tooltipId?: string;
	onSelect: () => void;
}

export const IconEnumOptionItem = <T,>({ option, tooltipId, onSelect }: IconEnumOptionItemProps<T>) => {
	if (option.hidden) return null;

	return (
		<li role="none">
			<Menu.LinkItem
				closeOnClick
				render={
					<IconSwatch
						href={option.icon?.href}
						style={iconStyleOf(option)}
						text={option.text}
						className="filter-[opacity(0.7)] transition-none hover:filter-none"
						testId="icon-enum-option"
						onActivate={onSelect}
						{...option.wowhead}
					/>
				}
				{...tooltipAnchorProps(option.tooltip ? tooltipId : undefined, option.tooltip)}
			/>
		</li>
	);
};
