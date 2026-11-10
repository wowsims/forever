import i18n from '@i18n/config';
import { WOWSIMS_URL } from '@sim/constants/other';
import { Button } from '@ui-kit/Button';

/** The WoWSims logo and name, linking to wowsims.com in the same tab. */
export const LandingBrand = () => (
	<Button href={WOWSIMS_URL} variant="unstyled" className="ui-landing-brand" data-testid="landing-brand">
		<img className="ui-landing-brand-logo" src="/forever/assets/img/WoW-Simulator-Icon.png" alt="" />
		<span className="flex flex-col leading-none">
			<span className="text-xl font-bold text-brand" data-testid="wowsims-title">
				{i18n.t('landing.header.wowsims')}
			</span>
			<small className="mt-[0.2rem] text-[0.7rem] font-normal text-white/65">{i18n.t('landing.header.subtitle')}</small>
		</span>
	</Button>
);
