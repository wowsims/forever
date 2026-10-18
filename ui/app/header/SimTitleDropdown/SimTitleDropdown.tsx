import type { Class } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translatePlayerClass, translatePlayerSpec, translateStatus } from '@i18n/localization';
import { PlayerClasses } from '@sim/player/classes/index';
import type { PlayerClass } from '@sim/player/player_class';
import type { PlayerSpec } from '@sim/player/player_spec';
import { PlayerSpecs } from '@sim/player/specs/index';
import { textClassNameForClass, textClassNameForSpec } from '@sim/proto/utils';
import { Button } from '@ui-kit/Button';
import { Menu, MenuLinkItem } from '@ui-kit/Menu';
import { SimLinkContent } from '@ui-kit/SimLinkContent';
import clsx from 'clsx';

export interface SimTitleDropdownProps {
	currentSpec: PlayerSpec<any>;
}

const SIM_LINK_CLASSES =
	'block w-full border-0 bg-background text-left no-underline hover:bg-surface-hover focus:bg-surface-hover data-highlighted:bg-surface-hover data-popup-open:bg-surface-hover';

const launchLabel = (launch: { phase: number; status: number }) =>
	i18n.t('sidebar.header.phase', { phase: i18n.t(`common.phases.${launch.phase}`), status: translateStatus(launch.status) });

const ClassSubmenu = ({ playerClass }: { playerClass: PlayerClass<Class> }) => {
	const specs = Object.values(playerClass.specs);
	// A class with one sim for all its specs names it "All Specializations", as the landing page does.
	const specTitle = (spec: PlayerSpec<any>) => (specs.length === 1 ? i18n.t('landing.home.allSpecs') : translatePlayerSpec(spec));
	return (
		<Menu
			submenu
			surface="plain"
			trigger={<SimLinkContent iconPath={playerClass.getIcon('large')} title={translatePlayerClass(playerClass)} />}
			triggerRender={<Button variant="unstyled" />}
			triggerProps={{ className: clsx(SIM_LINK_CLASSES, textClassNameForClass(playerClass)), 'data-testid': 'sim-link' }}
			side="right"
			align="start"
			sideOffset={0}
			className="ui-sim-title-popup w-auto min-w-popup-min-w"
			positionerProps={{ 'data-testid': 'sim-title-positioner' }}
			popupProps={{ 'data-testid': 'sim-title-popup' }}>
			{specs.map(spec => (
				<MenuLinkItem
					key={spec.simLink}
					layout="block"
					bare
					href={new URL(spec.simLink, window.location.href).toString()}
					className={clsx(SIM_LINK_CLASSES, textClassNameForSpec(spec))}
					testId="sim-link">
					<SimLinkContent
						iconPath={spec.getIcon('large')}
						label={translatePlayerClass(PlayerSpecs.getPlayerClass(spec))}
						title={specTitle(spec)}
						status={launchLabel(spec.launch)}
					/>
				</MenuLinkItem>
			))}
		</Menu>
	);
};

export const SimTitleDropdown = ({ currentSpec }: SimTitleDropdownProps) => (
	<div className="z-sim-title h-sim-header-plus border-b border-b-border pb-px max-lg:sticky max-lg:-top-px" data-testid="sim-title-dropdown-root">
		<div className="relative h-full" data-testid="sim-link-dropdown">
			<Menu
				surface="plain"
				anchorWidth
				trigger={
					<SimLinkContent
						iconPath={currentSpec.getIcon('large')}
						label={i18n.t('sidebar.header.title')}
						labelClassName="text-white"
						title={PlayerSpecs.getFullSpecName(currentSpec)}
						status={launchLabel(currentSpec.launch)}
					/>
				}
				triggerProps={{
					className: clsx(
						'ui-sim-title-trigger flex h-full w-full bg-background hover:bg-surface-hover focus:bg-surface-hover data-popup-open:bg-surface-hover',
						textClassNameForSpec(currentSpec),
					),
					'data-testid': 'sim-link',
				}}
				align="start"
				sideOffset={0}
				className="ui-sim-title-popup"
				positionerClassName="ui-sim-title-positioner"
				positionerProps={{ 'data-testid': 'sim-title-positioner' }}
				popupProps={{ 'data-testid': 'sim-title-popup' }}>
				{PlayerClasses.naturalOrder.map(playerClass => (
					<ClassSubmenu key={playerClass.friendlyName} playerClass={playerClass} />
				))}
			</Menu>
		</div>
	</div>
);
