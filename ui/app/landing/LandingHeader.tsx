import i18n from '@i18n/config';
import { ADDON_URL, DISCORD_URL, GITHUB_ORG_URL, PATREON_URL } from '@sim/constants/other';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import clsx from 'clsx';
import { useState } from 'react';

import { LandingBrand } from './LandingBrand';
import { LandingLanguageMenu } from './LandingLanguageMenu';

const COLLAPSE_ID = 'homepageHeaderCollapse';

export const LandingHeader = () => {
	const [open, setOpen] = useState(false);
	const toggleLabel = i18n.t('landing.navigation.toggle');

	const toggler = (icon: 'bars' | 'xmark') => (
		<Button
			variant="unstyled"
			className="border-0 px-2 py-1 text-lg leading-none text-white max-md:absolute max-md:top-0 max-md:right-0 md:hidden"
			aria-controls={COLLAPSE_ID}
			aria-expanded={open}
			aria-label={toggleLabel}
			data-testid="navbar-toggler"
			onClick={() => setOpen(current => !current)}>
			<Icon name={icon} size="2x" />
		</Button>
	);

	return (
		<header className="sticky top-0 z-header border-b border-solid border-white/8 bg-background/90">
			<div className="ui-landing-container flex py-4">
				<nav className="relative flex w-full flex-wrap items-center justify-between md:justify-start">
					<div className="order-0 flex max-md:w-full max-md:items-center max-md:justify-between">
						<LandingBrand />
						{toggler('bars')}
					</div>
					<div
						id={COLLAPSE_ID}
						className={clsx(
							'order-2 grow basis-full items-center justify-end max-md:fixed max-md:inset-0 max-md:z-dropdown max-md:bg-black/90 max-md:p-4 md:order-1 md:flex md:basis-auto',
							!open && 'hidden',
						)}>
						<div className="flex flex-col max-md:relative max-md:items-start md:flex-row md:items-center md:gap-6" data-testid="navbar-nav">
							{toggler('xmark')}
							<Button
								href={DISCORD_URL}
								target="_blank"
								variant="unstyled"
								title={i18n.t('landing.actions.discord')}
								className="ui-landing-nav-link">
								<Icon name="discord" style="brands" size="lg" className="mr-2" />
								<strong>{i18n.t('landing.header.discord')}</strong>
							</Button>
							<Button
								href={GITHUB_ORG_URL}
								target="_blank"
								variant="unstyled"
								title={i18n.t('landing.actions.github')}
								className="ui-landing-nav-link">
								<Icon name="github" style="brands" size="lg" className="mr-2" />
								<strong>{i18n.t('landing.header.github')}</strong>
							</Button>
							<Button href={ADDON_URL} target="_blank" variant="unstyled" title={i18n.t('landing.header.addon')} className="ui-landing-nav-badge">
								<img className="h-landing-badge" src="/forever/assets/img/curseforge_install.png" alt={i18n.t('landing.header.addon')} />
							</Button>
							<Button href={PATREON_URL} target="_blank" variant="unstyled" className="ui-landing-nav-badge">
								<img className="h-landing-badge" src="/forever/assets/img/patreon_become_a_patron.png" alt={i18n.t('landing.header.patreon')} />
							</Button>
							<LandingLanguageMenu triggerClassName="px-0 max-md:py-2" />
						</div>
					</div>
				</nav>
			</div>
		</header>
	);
};
