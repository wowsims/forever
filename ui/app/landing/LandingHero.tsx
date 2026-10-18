import i18n from '@i18n/config';
import { Trans } from 'react-i18next';

import { LandingActions } from './LandingActions';

export const LandingHero = () => (
	<section className="mb-12 flex max-w-landing-hero flex-col">
		<h1 className="ui-landing-hero-title" data-testid="landing-heading">
			<Trans i18n={i18n} i18nKey="landing.home.heading" components={{ expansion: <span className="text-expansion" /> }} />
		</h1>
		<p id="description" className="ui-landing-hero-blurb">
			<Trans
				i18n={i18n}
				i18nKey="landing.home.welcomeDescription"
				components={{
					brand: <strong className="text-brand" />,
					bold: <strong />,
					expansion: <strong className="text-expansion" />,
				}}
			/>
		</p>
		<LandingActions />
	</section>
);
