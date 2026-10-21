import { Class } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translatePlayerClass, translatePlayerSpec, translateStatus } from '@i18n/localization';
import { LaunchStatus } from '@sim/constants/other';
import { PlayerClasses } from '@sim/player/classes/index';
import type { PlayerClass } from '@sim/player/player_class';
import type { PlayerSpec } from '@sim/player/player_spec';
import { Button } from '@ui-kit/Button';
import { SimLinkContent } from '@ui-kit/SimLinkContent';
import clsx from 'clsx';
import { type CSSProperties, type ReactNode, useState } from 'react';

import { classLaunchStatus, SPEC_THEME_CLASS } from './landing_classes';

export interface LandingClassCardProps {
	playerClass: PlayerClass<Class>;
	// Position in the grid, used to stagger the entrance animation.
	index: number;
}

// A launched sim shows no status; every other status is shown under the class name.
const statusBadge = (status: LaunchStatus): string | undefined => (status === LaunchStatus.Launched ? undefined : translateStatus(status));

interface CardContentProps {
	coverClassName: string;
	iconPath: string;
	title: string;
	status?: string;
	children?: ReactNode;
}

const CardContent = ({ coverClassName, iconPath, title, status, children }: CardContentProps) => (
	<>
		<span className={clsx('ui-landing-class-card-cover', coverClassName)} aria-hidden="true" />
		<span className="ui-landing-class-card-body">
			<span className="ui-landing-class-card-title">
				<SimLinkContent iconPath={iconPath} title={title} status={status} />
			</span>
			{children}
		</span>
		<span className="ui-landing-class-card-sheen" aria-hidden="true" />
	</>
);

export const LandingClassCard = ({ playerClass, index }: LandingClassCardProps) => {
	const className = translatePlayerClass(playerClass);
	const scheme = PlayerClasses.getCssScheme(playerClass);
	const specs = Object.values(playerClass.specs) as Array<PlayerSpec<any>>;
	// Picked when the card mounts, so a re-render keeps the card's background.
	const [coverSpec] = useState(() => specs[Math.floor(Math.random() * specs.length)]);
	// Priest's class-level icon (`getIcon`) is a spell icon (spell_shadow_shadowwordpain), not the
	// generic class_*.jpg every other class returns, so the landing page hardcodes class_priest.jpg here.
	const classIconPath =
		playerClass.classID === Class.ClassPriest ? 'https://wow.zamimg.com/images/wow/icons/large/class_priest.jpg' : playerClass.getIcon('large');

	// A class with one sim for all its specs gets a single "All Specializations" link, under the class icon.
	const links =
		specs.length === 1
			? [{ href: specs[0].simLink, iconPath: classIconPath, label: i18n.t('landing.home.allSpecs') }]
			: specs.map(spec => ({ href: spec.simLink, iconPath: spec.getIcon('large'), label: translatePlayerSpec(spec) }));

	return (
		<div className="ui-landing-class-card" data-class={scheme} data-testid="class-card" style={{ '--ui-landing-card-index': index } as CSSProperties}>
			<CardContent
				coverClassName={SPEC_THEME_CLASS[coverSpec.specID as keyof typeof SPEC_THEME_CLASS]}
				iconPath={classIconPath}
				title={className}
				status={statusBadge(classLaunchStatus(playerClass))}>
				<span className="flex flex-wrap gap-2">
					{links.map(link => (
						<Button key={link.href} href={link.href} variant="unstyled" className="ui-landing-spec-pill" data-testid="spec-link">
							<img className="size-6 rounded-full" src={link.iconPath} alt="" />
							{link.label}
						</Button>
					))}
				</span>
			</CardContent>
		</div>
	);
};
