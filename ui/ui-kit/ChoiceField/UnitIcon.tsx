import { Icon, type IconName } from '@ui-kit/Icon';

export interface UnitIconProps {
	icon?: IconName;
	iconUrl?: string;
}

/** The two shapes an option icon takes, a kit glyph or an image url, as one element. */
export const UnitIcon = ({ icon, iconUrl }: UnitIconProps) =>
	icon ? (
		<Icon name={icon} style="base" className="ui-unit-icon" data-testid="unit-picker-item-icon" />
	) : (
		<img className="ui-unit-icon" data-testid="unit-picker-item-icon" src={iconUrl} alt="" />
	);
