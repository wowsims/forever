import { Class } from '@generated/proto/common';
import { translatePlayerClass, translatePlayerSpec, translateStatus } from '@i18n/localization';
import { PlayerClasses } from '@sim/player/classes/index';
import type { PlayerClass } from '@sim/player/player_class';
import type { PlayerSpec } from '@sim/player/player_spec';
import { textClassNameForClass, textClassNameForSpec } from '@sim/proto/utils';
import { Button } from '@ui-kit/Button';
import { Menu, MenuLinkItem } from '@ui-kit/Menu';
import { SimLinkContent } from '@ui-kit/SimLinkContent';
import { CLASS_BORDER } from '@ui-kit/utils/colors';
import clsx from 'clsx';
import { useState } from 'react';
import { useMedia } from 'react-use';

import { classLaunchStatus } from './landing_classes';

// Bootstrap's `lg` breakpoint: below it there is no room beside a row, so the menu opens under it instead.
const STACKED = '(max-width: 991.98px)';

export interface LandingClassMenuProps {
	playerClass: PlayerClass<Class>;
}

export const LandingClassMenu = ({ playerClass }: LandingClassMenuProps) => {
	const [container, setContainer] = useState<HTMLDivElement | null>(null);
	const stacked = useMedia(STACKED, false);
	const className = translatePlayerClass(playerClass);
	const specs = Object.values(playerClass.specs) as Array<PlayerSpec<any>>;
	// Priest's class-level icon (`getIcon`) is a spell icon (spell_shadow_shadowwordpain), not the
	// generic class_*.jpg every other class returns, so the landing page hardcodes class_priest.jpg here.
	const classIconPath =
		playerClass.classID === Class.ClassPriest ? 'https://wow.zamimg.com/images/wow/icons/large/class_priest.jpg' : playerClass.getIcon('large');

	// A class with a single spec links straight to it rather than opening a one-entry menu.
	if (specs.length === 1) {
		const [spec] = specs;
		return (
			<div className="ui-landing-sim-link-dropdown" data-testid="sim-link-dropdown">
				<Button
					href={spec.simLink}
					variant="unstyled"
					className={clsx('ui-landing-sim-link-cell', textClassNameForClass(playerClass))}
					data-testid="sim-link">
					<SimLinkContent
						iconPath={classIconPath}
						iconClassName={CLASS_BORDER[PlayerClasses.getCssScheme(playerClass)]}
						title={className}
						status={translateStatus(spec.launch.status)}
					/>
				</Button>
			</div>
		);
	}

	return (
		<div className="ui-landing-sim-link-dropdown" ref={setContainer} data-testid="sim-link-dropdown">
			<Menu
				surface="none"
				trigger={
					<SimLinkContent
						iconPath={classIconPath}
						iconClassName={CLASS_BORDER[PlayerClasses.getCssScheme(playerClass)]}
						title={className}
						status={translateStatus(classLaunchStatus(playerClass))}
					/>
				}
				triggerProps={{
					openOnHover: true,
					delay: 0,
					className: clsx('ui-landing-sim-link-cell', textClassNameForClass(playerClass)),
					'data-testid': 'sim-link',
				}}
				side={stacked ? 'bottom' : 'right'}
				align="start"
				sideOffset={0}
				collisionPadding={4}
				container={container}
				// `keepMounted`: the 18 spec links are the landing page's only content for a crawler, and an unmounted popup has none of them in the document.
				keepMounted
				className="ui-landing-sim-link-popup">
				{specs.map(spec => (
					<MenuLinkItem
						key={spec.simLink}
						layout="block"
						bare
						href={spec.simLink}
						className={clsx('ui-landing-sim-link-cell', textClassNameForSpec(spec))}
						testId="sim-link">
						<SimLinkContent
							iconPath={spec.getIcon('large')}
							label={className}
							title={translatePlayerSpec(spec)}
							status={translateStatus(spec.launch.status)}
						/>
					</MenuLinkItem>
				))}
			</Menu>
		</div>
	);
};
